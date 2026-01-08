package recommendations

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// SeedDeckBuilder provides intelligent deck building suggestions based on a seed card.
type SeedDeckBuilder struct {
	setCardRepo    repository.SetCardRepository
	collectionRepo repository.CollectionRepository
	standardRepo   repository.StandardRepository
	cardService    *cards.Service
}

// NewSeedDeckBuilder creates a new seed deck builder.
func NewSeedDeckBuilder(
	setCardRepo repository.SetCardRepository,
	collectionRepo repository.CollectionRepository,
	standardRepo repository.StandardRepository,
	cardService *cards.Service,
) *SeedDeckBuilder {
	return &SeedDeckBuilder{
		setCardRepo:    setCardRepo,
		collectionRepo: collectionRepo,
		standardRepo:   standardRepo,
		cardService:    cardService,
	}
}

// SeedDeckBuilderRequest represents a request to build around a seed card.
type SeedDeckBuilderRequest struct {
	SeedCardID     int      `json:"seedCardID"`
	MaxResults     int      `json:"maxResults"`     // Default: 40
	BudgetMode     bool     `json:"budgetMode"`     // Only collection cards
	SetRestriction string   `json:"setRestriction"` // "single", "multiple", "all"
	AllowedSets    []string `json:"allowedSets"`    // Specific set codes if "multiple"
}

// SeedDeckBuilderResponse contains suggested cards with ownership info.
type SeedDeckBuilderResponse struct {
	SeedCard        *CardWithOwnership   `json:"seedCard"`
	Suggestions     []*CardWithOwnership `json:"suggestions"`
	LandSuggestions []*SuggestedLand     `json:"lands"`
	Analysis        *SeedDeckAnalysis    `json:"analysis"`
}

// ScoreBreakdown provides detailed scoring factors for a card suggestion.
type ScoreBreakdown struct {
	ColorFit float64 `json:"colorFit"` // 0.0-1.0, weight: 25%
	CurveFit float64 `json:"curveFit"` // 0.0-1.0, weight: 20%
	Synergy  float64 `json:"synergy"`  // 0.0-1.0, weight: 30%
	Quality  float64 `json:"quality"`  // 0.0-1.0, weight: 15%
	Overall  float64 `json:"overall"`  // Final weighted score
}

// SynergyDetail describes a specific synergy between a card and the deck.
type SynergyDetail struct {
	Type        string `json:"type"`        // "keyword", "theme", "creature_type"
	Name        string `json:"name"`        // e.g., "flying", "tokens", "Elf"
	Description string `json:"description"` // e.g., "Matches 3 other flying creatures"
}

// CardWithOwnership extends card info with ownership data.
type CardWithOwnership struct {
	CardID            int             `json:"cardID"`
	Name              string          `json:"name"`
	ManaCost          string          `json:"manaCost,omitempty"`
	CMC               int             `json:"cmc"`
	Colors            []string        `json:"colors"`
	TypeLine          string          `json:"typeLine"`
	Rarity            string          `json:"rarity,omitempty"`
	ImageURI          string          `json:"imageURI,omitempty"`
	Score             float64         `json:"score"`
	Reasoning         string          `json:"reasoning"`
	InCollection      bool            `json:"inCollection"`
	OwnedCount        int             `json:"ownedCount"`
	NeededCount       int             `json:"neededCount"`
	CurrentCopies     int             `json:"currentCopies"`     // Copies currently in deck
	RecommendedCopies int             `json:"recommendedCopies"` // Recommended total copies (1-4)
	ScoreBreakdown    *ScoreBreakdown `json:"scoreBreakdown,omitempty"`
	SynergyDetails    []SynergyDetail `json:"synergyDetails,omitempty"`
}

// SeedDeckAnalysis provides analysis of the seed card and suggestions.
type SeedDeckAnalysis struct {
	ColorIdentity       []string       `json:"colorIdentity"`
	Keywords            []string       `json:"keywords"`
	Themes              []string       `json:"themes"`
	IdealCurve          map[int]int    `json:"idealCurve"`
	SuggestedLandCount  int            `json:"suggestedLandCount"`
	TotalCards          int            `json:"totalCards"`
	InCollectionCount   int            `json:"inCollectionCount"`
	MissingCount        int            `json:"missingCount"`
	MissingWildcardCost map[string]int `json:"missingWildcardCost"` // rarity -> count
}

// SeedCardAnalysis contains analyzed seed card data.
type SeedCardAnalysis struct {
	Card          *cards.Card
	Colors        []string
	Keywords      []KeywordInfo
	Themes        []string
	CardTypes     []string
	CMC           int
	IsCreature    bool
	CreatureTypes []string
}

// IterativeBuildAroundRequest represents a request for iterative deck building suggestions.
type IterativeBuildAroundRequest struct {
	SeedCardID     int      `json:"seedCardID"`     // Optional - original seed card (ignored if DeckCardIDs provided)
	DeckCardIDs    []int    `json:"deckCardIDs"`    // All cards currently in deck (required)
	MaxResults     int      `json:"maxResults"`     // Default: 15
	BudgetMode     bool     `json:"budgetMode"`     // Only collection cards
	SetRestriction string   `json:"setRestriction"` // "single", "multiple", "all"
	AllowedSets    []string `json:"allowedSets"`    // Specific set codes if "multiple"
}

// IterativeBuildAroundResponse contains suggestions for the next cards to add.
type IterativeBuildAroundResponse struct {
	Suggestions     []*CardWithOwnership `json:"suggestions"`
	DeckAnalysis    *LiveDeckAnalysis    `json:"deckAnalysis"`
	SlotsRemaining  int                  `json:"slotsRemaining"`
	LandSuggestions []*SuggestedLand     `json:"landSuggestions"`
}

// LiveDeckAnalysis provides real-time analysis of the deck being built.
type LiveDeckAnalysis struct {
	ColorIdentity        []string    `json:"colorIdentity"`
	Keywords             []string    `json:"keywords"`
	Themes               []string    `json:"themes"`
	CurrentCurve         map[int]int `json:"currentCurve"`
	RecommendedLandCount int         `json:"recommendedLandCount"`
	TotalCards           int         `json:"totalCards"`
	InCollectionCount    int         `json:"inCollectionCount"`
}

// CollectiveDeckAnalysis aggregates analysis across all cards in the deck.
type CollectiveDeckAnalysis struct {
	Colors        map[string]int
	Keywords      []KeywordInfo
	Themes        map[string]int
	CreatureTypes map[string]int
	ManaCurve     map[int]int
	TotalCards    int
}

// BuildAroundSeed generates deck suggestions based on a seed card.
func (s *SeedDeckBuilder) BuildAroundSeed(ctx context.Context, req *SeedDeckBuilderRequest) (*SeedDeckBuilderResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.SeedCardID <= 0 {
		return nil, fmt.Errorf("seed card ID is required")
	}

	// Apply defaults
	if req.MaxResults <= 0 {
		req.MaxResults = 40
	}
	if req.SetRestriction == "" {
		req.SetRestriction = "all"
	}

	// Get and analyze seed card
	seedAnalysis, err := s.analyzeSeedCard(ctx, req.SeedCardID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze seed card: %w", err)
	}

	// Get candidate cards
	candidates, err := s.getCandidates(ctx, req, seedAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// Score candidates
	scoredCards := s.scoreAndRankCandidates(candidates, seedAnalysis)

	// Get collection ownership
	collection, err := s.getCollectionMap(ctx)
	if err != nil {
		// Non-fatal - continue without ownership
		collection = make(map[int]int)
	}

	// Apply budget mode filter if enabled
	if req.BudgetMode {
		scoredCards = s.filterToCollection(scoredCards, collection)
	}

	// Limit results
	if len(scoredCards) > req.MaxResults {
		scoredCards = scoredCards[:req.MaxResults]
	}

	// Enrich with ownership
	suggestions := s.enrichWithOwnership(scoredCards, collection)

	// Generate land suggestions
	landSuggestions := s.suggestLands(seedAnalysis, suggestions)

	// Build analysis
	analysis := s.buildAnalysis(seedAnalysis, suggestions, landSuggestions)

	// Build seed card response
	seedCardWithOwnership := s.buildSeedCardResponse(seedAnalysis, collection)

	return &SeedDeckBuilderResponse{
		SeedCard:        seedCardWithOwnership,
		Suggestions:     suggestions,
		LandSuggestions: landSuggestions,
		Analysis:        analysis,
	}, nil
}

// analyzeSeedCard extracts key information from the seed card.
func (s *SeedDeckBuilder) analyzeSeedCard(ctx context.Context, cardID int) (*SeedCardAnalysis, error) {
	// Get card from card service
	card, err := s.cardService.GetCard(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seed card: %w", err)
	}
	if card == nil {
		return nil, fmt.Errorf("seed card not found: %d", cardID)
	}

	analysis := &SeedCardAnalysis{
		Card:   card,
		Colors: card.Colors,
		CMC:    int(card.CMC),
	}

	// Extract keywords and themes from oracle text
	if card.OracleText != nil && *card.OracleText != "" {
		keywords := ExtractKeywordsWithInfo(*card.OracleText)
		analysis.Keywords = keywords

		// Extract theme names
		themes := make([]string, 0)
		seenThemes := make(map[string]bool)
		for _, kw := range keywords {
			if kw.Category == CategoryTheme && !seenThemes[kw.Keyword] {
				themes = append(themes, kw.Keyword)
				seenThemes[kw.Keyword] = true
			}
		}
		analysis.Themes = themes
	}

	// Extract card types from type line
	analysis.CardTypes = extractTypesFromTypeLine(card.TypeLine)

	// Check if creature and extract creature types
	analysis.IsCreature = containsTypeInTypeLine(card.TypeLine, "Creature")
	if analysis.IsCreature {
		creatureTypes := extractCreatureTypesFromLine(card.TypeLine)
		for ct := range creatureTypes {
			analysis.CreatureTypes = append(analysis.CreatureTypes, ct)
		}
	}

	return analysis, nil
}

// getCandidates retrieves candidate cards based on request filters.
func (s *SeedDeckBuilder) getCandidates(ctx context.Context, req *SeedDeckBuilderRequest, seedAnalysis *SeedCardAnalysis) ([]*cards.Card, error) {
	var candidates []*cards.Card

	// Get Standard-legal cards
	switch req.SetRestriction {
	case "single":
		// Use seed card's set
		if seedAnalysis.Card != nil {
			setCards, err := s.getCardsFromSet(ctx, seedAnalysis.Card.SetCode)
			if err != nil {
				return nil, err
			}
			candidates = setCards
		}
	case "multiple":
		// Get cards from specified sets
		for _, setCode := range req.AllowedSets {
			setCards, err := s.getCardsFromSet(ctx, setCode)
			if err != nil {
				continue // Skip sets that fail
			}
			candidates = append(candidates, setCards...)
		}
	default: // "all"
		// Get all Standard-legal cards
		standardSets, err := s.standardRepo.GetStandardSets(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get standard sets: %w", err)
		}
		for _, set := range standardSets {
			setCards, err := s.getCardsFromSet(ctx, set.Code)
			if err != nil {
				continue // Skip sets that fail
			}
			candidates = append(candidates, setCards...)
		}
	}

	// Filter out the seed card itself
	filtered := make([]*cards.Card, 0, len(candidates))
	for _, card := range candidates {
		if card.ArenaID != req.SeedCardID {
			filtered = append(filtered, card)
		}
	}

	return filtered, nil
}

// getCardsFromSet retrieves cards from a set and converts them to cards.Card.
func (s *SeedDeckBuilder) getCardsFromSet(ctx context.Context, setCode string) ([]*cards.Card, error) {
	setCards, err := s.setCardRepo.GetCardsBySet(ctx, setCode)
	if err != nil {
		return nil, err
	}

	result := make([]*cards.Card, 0, len(setCards))
	for _, sc := range setCards {
		card := convertSetCardToCardsCard(sc)
		if card != nil {
			result = append(result, card)
		}
	}

	return result, nil
}

// scoreAndRankCandidates scores all candidates against the seed card.
func (s *SeedDeckBuilder) scoreAndRankCandidates(candidates []*cards.Card, seedAnalysis *SeedCardAnalysis) []*scoredCard {
	scored := make([]*scoredCard, 0, len(candidates))

	for _, card := range candidates {
		score, reasoning, breakdown, synergyDetails := s.scoreCardForSeed(card, seedAnalysis)

		// Skip cards with very low scores
		if score < 0.3 {
			continue
		}

		scored = append(scored, &scoredCard{
			card:           card,
			score:          score,
			reasoning:      reasoning,
			scoreBreakdown: breakdown,
			synergyDetails: synergyDetails,
		})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored
}

// scoreCardForSeed calculates how well a card fits with the seed card.
// Returns the overall score, reasoning text, score breakdown, and synergy details.
func (s *SeedDeckBuilder) scoreCardForSeed(card *cards.Card, seedAnalysis *SeedCardAnalysis) (float64, string, *ScoreBreakdown, []SynergyDetail) {
	reasons := make([]string, 0)
	synergyDetails := make([]SynergyDetail, 0)

	// Factor 1: Color Compatibility (25%)
	colorScore := s.scoreColorCompatibility(card, seedAnalysis)
	if colorScore >= 0.8 {
		reasons = append(reasons, "matches your colors")
	}

	// Factor 2: Mana Curve (20%)
	curveScore := s.scoreManaCurveFit(card)
	if curveScore >= 0.7 {
		reasons = append(reasons, fmt.Sprintf("good curve fit at %d CMC", int(card.CMC)))
	}

	// Factor 3: Synergy with Seed (30%) - now captures synergy details
	synergyScore, cardSynergyDetails := s.scoreSynergyWithSeedDetailed(card, seedAnalysis)
	synergyDetails = append(synergyDetails, cardSynergyDetails...)
	if synergyScore >= 0.7 {
		reasons = append(reasons, "synergizes with your strategy")
	}

	// Factor 4: Card Quality (15%)
	qualityScore := s.scoreCardQuality(card)
	if qualityScore >= 0.7 {
		reasons = append(reasons, "high-quality card")
	}

	// Factor 5: Standard Legality (5%) - should be 1.0 for all candidates
	legalityScore := 1.0

	// Factor 6: Playability (5%)
	playabilityScore := 0.8 // Default for Standard

	// Calculate weighted score
	score := (colorScore * 0.25) +
		(curveScore * 0.20) +
		(synergyScore * 0.30) +
		(qualityScore * 0.15) +
		(legalityScore * 0.05) +
		(playabilityScore * 0.05)

	// Build score breakdown
	breakdown := &ScoreBreakdown{
		ColorFit: colorScore,
		CurveFit: curveScore,
		Synergy:  synergyScore,
		Quality:  qualityScore,
		Overall:  score,
	}

	// Build reasoning string
	reasoning := "This card "
	if len(reasons) == 0 {
		reasoning = "This card could work in your deck."
	} else {
		for i, r := range reasons {
			if i == 0 {
				reasoning += r
			} else if i == len(reasons)-1 {
				reasoning += ", and " + r
			} else {
				reasoning += ", " + r
			}
		}
		reasoning += "."
	}

	return score, reasoning, breakdown, synergyDetails
}

// scoreColorCompatibility scores how well a card's colors match the seed.
func (s *SeedDeckBuilder) scoreColorCompatibility(card *cards.Card, seedAnalysis *SeedCardAnalysis) float64 {
	if len(card.Colors) == 0 {
		// Colorless cards fit any deck
		return 1.0
	}

	if len(seedAnalysis.Colors) == 0 {
		// Seed is colorless - any color works
		return 0.8
	}

	// Check how many card colors match seed colors
	matchingColors := 0
	for _, cardColor := range card.Colors {
		for _, seedColor := range seedAnalysis.Colors {
			if cardColor == seedColor {
				matchingColors++
				break
			}
		}
	}

	if matchingColors == 0 {
		// No color overlap
		return 0.0
	}

	if matchingColors == len(card.Colors) {
		// All card colors are in seed's colors
		return 1.0
	}

	// Partial match
	return float64(matchingColors) / float64(len(card.Colors)) * 0.7
}

// scoreManaCurveFit scores how well a card fits the ideal mana curve.
func (s *SeedDeckBuilder) scoreManaCurveFit(card *cards.Card) float64 {
	if containsTypeInTypeLine(card.TypeLine, "Land") {
		return 0.5 // Neutral for lands
	}

	cmc := int(card.CMC)

	// Ideal distribution for Standard constructed
	// More 2-3 drops, fewer high CMC cards
	idealWeights := map[int]float64{
		0: 0.6, // CMC 0 cards are situational
		1: 0.8,
		2: 1.0, // Sweet spot
		3: 1.0, // Sweet spot
		4: 0.8,
		5: 0.6,
		6: 0.4,
	}

	weight, ok := idealWeights[cmc]
	if !ok {
		if cmc > 6 {
			weight = 0.3 // High CMC cards are risky
		} else {
			weight = 0.5
		}
	}

	return weight
}

// scoreSynergyWithSeed scores synergy between a card and the seed.
func (s *SeedDeckBuilder) scoreSynergyWithSeed(card *cards.Card, seedAnalysis *SeedCardAnalysis) float64 {
	score, _ := s.scoreSynergyWithSeedDetailed(card, seedAnalysis)
	return score
}

// scoreSynergyWithSeedDetailed scores synergy and returns detailed synergy information.
func (s *SeedDeckBuilder) scoreSynergyWithSeedDetailed(card *cards.Card, seedAnalysis *SeedCardAnalysis) (float64, []SynergyDetail) {
	synergy := 0.0
	synergyCount := 0
	details := make([]SynergyDetail, 0)

	// Extract card keywords
	var cardKeywords []KeywordInfo
	if card.OracleText != nil && *card.OracleText != "" {
		cardKeywords = ExtractKeywordsWithInfo(*card.OracleText)
	}

	// Keyword synergy
	if len(cardKeywords) > 0 && len(seedAnalysis.Keywords) > 0 {
		keywordSynergy, matchedKeywords := CalculateKeywordSynergyDetailed(seedAnalysis.Keywords, cardKeywords)
		if keywordSynergy > 0 {
			synergy += keywordSynergy
			synergyCount++
			for _, kw := range matchedKeywords {
				details = append(details, SynergyDetail{
					Type:        "keyword",
					Name:        kw,
					Description: fmt.Sprintf("Shares %s with seed card", kw),
				})
			}
		}
	}

	// Creature type synergy (tribal)
	if containsTypeInTypeLine(card.TypeLine, "Creature") && seedAnalysis.IsCreature {
		cardCreatureTypes := extractCreatureTypesFromLine(card.TypeLine)
		for cardType := range cardCreatureTypes {
			for _, seedType := range seedAnalysis.CreatureTypes {
				if cardType == seedType {
					synergy += 0.8 // Strong tribal synergy
					synergyCount++
					details = append(details, SynergyDetail{
						Type:        "creature_type",
						Name:        cardType,
						Description: fmt.Sprintf("%s tribal synergy", cardType),
					})
					break
				}
			}
		}
	}

	// Theme synergy (e.g., both care about tokens, graveyard, etc.)
	cardThemes := make(map[string]bool)
	for _, kw := range cardKeywords {
		if kw.Category == CategoryTheme {
			cardThemes[kw.Keyword] = true
		}
	}
	for _, seedTheme := range seedAnalysis.Themes {
		if cardThemes[seedTheme] {
			synergy += 0.7
			synergyCount++
			details = append(details, SynergyDetail{
				Type:        "theme",
				Name:        seedTheme,
				Description: fmt.Sprintf("Supports %s theme", seedTheme),
			})
		}
	}

	if synergyCount == 0 {
		return 0.5, details // Neutral score
	}

	avgSynergy := synergy / float64(synergyCount)
	if avgSynergy > 1.0 {
		avgSynergy = 1.0
	}

	return avgSynergy, details
}

// scoreCardQuality scores intrinsic card quality based on rarity.
func (s *SeedDeckBuilder) scoreCardQuality(card *cards.Card) float64 {
	rarityScores := map[string]float64{
		"mythic":   0.85,
		"rare":     0.75,
		"uncommon": 0.60,
		"common":   0.50,
	}

	if score, ok := rarityScores[strings.ToLower(card.Rarity)]; ok {
		return score
	}

	return 0.5
}

// getCollectionMap retrieves the user's collection as a map.
func (s *SeedDeckBuilder) getCollectionMap(ctx context.Context) (map[int]int, error) {
	if s.collectionRepo == nil {
		return make(map[int]int), nil
	}

	return s.collectionRepo.GetAll(ctx)
}

// filterToCollection filters scored cards to only those in the collection.
func (s *SeedDeckBuilder) filterToCollection(scored []*scoredCard, collection map[int]int) []*scoredCard {
	filtered := make([]*scoredCard, 0)
	for _, sc := range scored {
		if collection[sc.card.ArenaID] > 0 {
			filtered = append(filtered, sc)
		}
	}
	return filtered
}

// enrichWithOwnership adds ownership data to scored cards.
func (s *SeedDeckBuilder) enrichWithOwnership(scored []*scoredCard, collection map[int]int) []*CardWithOwnership {
	result := make([]*CardWithOwnership, 0, len(scored))

	for _, sc := range scored {
		owned := collection[sc.card.ArenaID]
		needed := 4 - owned
		if needed < 0 {
			needed = 0
		}

		manaCost := ""
		if sc.card.ManaCost != nil {
			manaCost = *sc.card.ManaCost
		}

		imageURI := ""
		if sc.card.ImageURI != nil {
			imageURI = *sc.card.ImageURI
		}

		card := &CardWithOwnership{
			CardID:         sc.card.ArenaID,
			Name:           sc.card.Name,
			ManaCost:       manaCost,
			CMC:            int(sc.card.CMC),
			Colors:         sc.card.Colors,
			TypeLine:       sc.card.TypeLine,
			Rarity:         sc.card.Rarity,
			ImageURI:       imageURI,
			Score:          sc.score,
			Reasoning:      sc.reasoning,
			InCollection:   owned > 0,
			OwnedCount:     owned,
			NeededCount:    needed,
			ScoreBreakdown: sc.scoreBreakdown,
			SynergyDetails: sc.synergyDetails,
		}

		result = append(result, card)
	}

	return result
}

// suggestLands generates land suggestions based on color distribution.
func (s *SeedDeckBuilder) suggestLands(seedAnalysis *SeedCardAnalysis, suggestions []*CardWithOwnership) []*SuggestedLand {
	// Count colors across seed + suggestions
	colorCounts := make(map[string]int)

	// Add seed colors
	for _, c := range seedAnalysis.Colors {
		colorCounts[c] += 4 // Weight seed card heavily
	}

	// Add suggestion colors (weighted by how many we'd include)
	for i, card := range suggestions {
		weight := 1
		if i < 20 {
			weight = 2 // Top suggestions weighted more
		}
		for _, c := range card.Colors {
			colorCounts[c] += weight
		}
	}

	// Calculate land distribution (24 lands total for 60-card deck)
	totalLands := 24
	totalColorWeight := 0
	for _, count := range colorCounts {
		totalColorWeight += count
	}

	lands := make([]*SuggestedLand, 0)
	if totalColorWeight == 0 {
		// Colorless deck - suggest Wastes or utility lands
		return lands
	}

	for color, count := range colorCounts {
		land, ok := basicLandsByColor[color]
		if !ok {
			continue
		}

		proportion := float64(count) / float64(totalColorWeight)
		quantity := int(proportion*float64(totalLands) + 0.5)
		if quantity < 1 && count > 0 {
			quantity = 1 // At least 1 land of each color
		}

		lands = append(lands, &SuggestedLand{
			CardID:   land.ArenaID,
			Name:     land.Name,
			Quantity: quantity,
			Color:    color,
		})
	}

	return lands
}

// buildAnalysis generates the analysis summary.
func (s *SeedDeckBuilder) buildAnalysis(
	seedAnalysis *SeedCardAnalysis,
	suggestions []*CardWithOwnership,
	lands []*SuggestedLand,
) *SeedDeckAnalysis {
	// Count ownership stats
	inCollection := 0
	missing := 0
	wildcardCost := make(map[string]int)

	for _, card := range suggestions {
		if card.InCollection {
			inCollection++
		} else {
			missing++
			wildcardCost[strings.ToLower(card.Rarity)]++
		}
	}

	// Calculate total lands
	totalLands := 0
	for _, land := range lands {
		totalLands += land.Quantity
	}

	// Extract keyword names
	keywordNames := make([]string, 0)
	for _, kw := range seedAnalysis.Keywords {
		keywordNames = append(keywordNames, kw.Keyword)
	}

	return &SeedDeckAnalysis{
		ColorIdentity:       seedAnalysis.Colors,
		Keywords:            keywordNames,
		Themes:              seedAnalysis.Themes,
		IdealCurve:          map[int]int{1: 4, 2: 8, 3: 8, 4: 6, 5: 4, 6: 2}, // Standard curve
		SuggestedLandCount:  totalLands,
		TotalCards:          len(suggestions) + totalLands + 4, // +4 for seed card copies
		InCollectionCount:   inCollection,
		MissingCount:        missing,
		MissingWildcardCost: wildcardCost,
	}
}

// buildSeedCardResponse creates the seed card response with ownership.
func (s *SeedDeckBuilder) buildSeedCardResponse(seedAnalysis *SeedCardAnalysis, collection map[int]int) *CardWithOwnership {
	owned := collection[seedAnalysis.Card.ArenaID]
	needed := 4 - owned
	if needed < 0 {
		needed = 0
	}

	manaCost := ""
	if seedAnalysis.Card.ManaCost != nil {
		manaCost = *seedAnalysis.Card.ManaCost
	}

	imageURI := ""
	if seedAnalysis.Card.ImageURI != nil {
		imageURI = *seedAnalysis.Card.ImageURI
	}

	return &CardWithOwnership{
		CardID:       seedAnalysis.Card.ArenaID,
		Name:         seedAnalysis.Card.Name,
		ManaCost:     manaCost,
		CMC:          int(seedAnalysis.Card.CMC),
		Colors:       seedAnalysis.Card.Colors,
		TypeLine:     seedAnalysis.Card.TypeLine,
		Rarity:       seedAnalysis.Card.Rarity,
		ImageURI:     imageURI,
		Score:        1.0, // Seed card has max score
		Reasoning:    "This is your build-around card.",
		InCollection: owned > 0,
		OwnedCount:   owned,
		NeededCount:  needed,
	}
}

// SuggestNextCards generates suggestions based on the current deck composition.
// This analyzes ALL cards in the deck collectively to find commonalities and
// suggests cards that complement the deck's colors, themes, and mana curve.
func (s *SeedDeckBuilder) SuggestNextCards(ctx context.Context, req *IterativeBuildAroundRequest) (*IterativeBuildAroundResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if len(req.DeckCardIDs) == 0 {
		return nil, fmt.Errorf("deck must have at least one card")
	}

	// Apply defaults
	if req.MaxResults <= 0 {
		req.MaxResults = 15
	}
	if req.SetRestriction == "" {
		req.SetRestriction = "all"
	}

	// Analyze current deck collectively - this determines colors, themes, keywords, etc.
	deckAnalysis, err := s.analyzeDeckCards(ctx, req.DeckCardIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze deck: %w", err)
	}

	// Get candidate cards from all standard-legal sets
	candidates, err := s.getCandidatesFromDeckAnalysis(ctx, req, deckAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// Count copies of each card in deck (instead of excluding them)
	deckCardCounts := make(map[int]int)
	for _, cardID := range req.DeckCardIDs {
		deckCardCounts[cardID]++
	}

	// Only exclude cards that already have 4 copies (can't add more)
	excludeSet := make(map[int]bool)
	for cardID, count := range deckCardCounts {
		if count >= 4 {
			excludeSet[cardID] = true
		}
	}

	// Score candidates against the collective deck analysis
	scoredCards := s.scoreAndRankForDeck(candidates, deckAnalysis, excludeSet)

	// Get collection ownership
	collection, err := s.getCollectionMap(ctx)
	if err != nil {
		collection = make(map[int]int)
	}

	// Apply budget mode filter if enabled
	if req.BudgetMode {
		scoredCards = s.filterToCollection(scoredCards, collection)
	}

	// Limit results
	if len(scoredCards) > req.MaxResults {
		scoredCards = scoredCards[:req.MaxResults]
	}

	// Enrich with ownership
	suggestions := s.enrichWithOwnership(scoredCards, collection)

	// Add current deck counts and recommended copies to each card
	for _, card := range suggestions {
		card.CurrentCopies = deckCardCounts[card.CardID]
		recommended := s.calculateRecommendedCopies(card)
		if recommended < 1 {
			recommended = 4 // Default to 4 if calculation fails
		}
		card.RecommendedCopies = recommended
	}

	// Calculate slots remaining (60-card deck standard)
	slotsRemaining := 60 - len(req.DeckCardIDs)
	if slotsRemaining < 0 {
		slotsRemaining = 0
	}

	// Generate land suggestions based on current deck
	landSuggestions := s.suggestLandsForDeck(deckAnalysis)

	// Build live deck analysis
	liveAnalysis := s.buildLiveDeckAnalysis(deckAnalysis, collection, req.DeckCardIDs)

	return &IterativeBuildAroundResponse{
		Suggestions:     suggestions,
		DeckAnalysis:    liveAnalysis,
		SlotsRemaining:  slotsRemaining,
		LandSuggestions: landSuggestions,
	}, nil
}

// analyzeDeckCards analyzes multiple cards collectively for the deck.
func (s *SeedDeckBuilder) analyzeDeckCards(ctx context.Context, cardIDs []int) (*CollectiveDeckAnalysis, error) {
	analysis := &CollectiveDeckAnalysis{
		Colors:        make(map[string]int),
		Keywords:      make([]KeywordInfo, 0),
		Themes:        make(map[string]int),
		CreatureTypes: make(map[string]int),
		ManaCurve:     make(map[int]int),
		TotalCards:    len(cardIDs),
	}

	for _, cardID := range cardIDs {
		card, err := s.cardService.GetCard(cardID)
		if err != nil || card == nil {
			continue
		}

		// Aggregate colors
		for _, color := range card.Colors {
			analysis.Colors[color]++
		}

		// Track mana curve
		analysis.ManaCurve[int(card.CMC)]++

		// Extract and aggregate keywords
		if card.OracleText != nil && *card.OracleText != "" {
			keywords := ExtractKeywordsWithInfo(*card.OracleText)
			analysis.Keywords = append(analysis.Keywords, keywords...)

			// Aggregate themes
			for _, kw := range keywords {
				if kw.Category == CategoryTheme {
					analysis.Themes[kw.Keyword]++
				}
			}
		}

		// Extract creature types
		if containsTypeInTypeLine(card.TypeLine, "Creature") {
			creatureTypes := extractCreatureTypesFromLine(card.TypeLine)
			for ct := range creatureTypes {
				analysis.CreatureTypes[ct]++
			}
		}
	}

	return analysis, nil
}

// getCandidatesFromDeckAnalysis retrieves candidate cards based on the deck's collective analysis.
// This fetches cards from standard-legal sets that match the deck's established colors.
func (s *SeedDeckBuilder) getCandidatesFromDeckAnalysis(ctx context.Context, req *IterativeBuildAroundRequest, deckAnalysis *CollectiveDeckAnalysis) ([]*cards.Card, error) {
	var candidates []*cards.Card

	// Get standard-legal sets
	standardSets, err := s.standardRepo.GetStandardSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get standard sets: %w", err)
	}

	// If no standard sets defined, fallback to all cached sets
	if len(standardSets) == 0 {
		cachedSets, err := s.setCardRepo.GetCachedSets(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get cached sets: %w", err)
		}
		for _, setCode := range cachedSets {
			setCards, err := s.getCardsFromSet(ctx, setCode)
			if err != nil {
				continue // Skip sets that fail
			}
			candidates = append(candidates, setCards...)
		}
	} else {
		// Get cards from all standard sets
		for _, set := range standardSets {
			setCards, err := s.getCardsFromSet(ctx, set.Code)
			if err != nil {
				continue // Skip sets that fail
			}
			candidates = append(candidates, setCards...)
		}
	}

	// Filter out cards that don't match the deck's colors (colorless always allowed)
	deckColors := deckAnalysis.Colors
	if len(deckColors) > 0 {
		filtered := make([]*cards.Card, 0, len(candidates))
		for _, card := range candidates {
			// Colorless cards always fit
			if len(card.Colors) == 0 {
				filtered = append(filtered, card)
				continue
			}

			// Check if card has at least one color that matches deck
			hasMatchingColor := false
			for _, cardColor := range card.Colors {
				if deckColors[cardColor] > 0 {
					hasMatchingColor = true
					break
				}
			}

			if hasMatchingColor {
				filtered = append(filtered, card)
			}
		}
		candidates = filtered
	}

	// Filter out deck cards
	excludeSet := make(map[int]bool)
	for _, cardID := range req.DeckCardIDs {
		excludeSet[cardID] = true
	}

	finalCandidates := make([]*cards.Card, 0, len(candidates))
	for _, card := range candidates {
		if !excludeSet[card.ArenaID] {
			finalCandidates = append(finalCandidates, card)
		}
	}

	return finalCandidates, nil
}

// scoreAndRankForDeck scores candidates against the collective deck analysis.
func (s *SeedDeckBuilder) scoreAndRankForDeck(candidates []*cards.Card, deckAnalysis *CollectiveDeckAnalysis, excludeSet map[int]bool) []*scoredCard {
	scored := make([]*scoredCard, 0, len(candidates))

	for _, card := range candidates {
		// Skip excluded cards
		if excludeSet[card.ArenaID] {
			continue
		}

		score, reasoning, breakdown, synergyDetails := s.scoreCardForDeck(card, deckAnalysis)

		// Skip cards with very low scores
		if score < 0.3 {
			continue
		}

		scored = append(scored, &scoredCard{
			card:           card,
			score:          score,
			reasoning:      reasoning,
			scoreBreakdown: breakdown,
			synergyDetails: synergyDetails,
		})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored
}

// scoreCardForDeck calculates how well a card fits with the current deck.
// Returns the overall score, reasoning text, score breakdown, and synergy details.
func (s *SeedDeckBuilder) scoreCardForDeck(card *cards.Card, deckAnalysis *CollectiveDeckAnalysis) (float64, string, *ScoreBreakdown, []SynergyDetail) {
	reasons := make([]string, 0)
	synergyDetails := make([]SynergyDetail, 0)

	// Factor 1: Color Compatibility (25%)
	colorScore := s.scoreColorForDeck(card, deckAnalysis)
	if colorScore >= 0.8 {
		reasons = append(reasons, "matches deck colors")
	}

	// Factor 2: Mana Curve Gap Filling (20%)
	curveScore := s.scoreCurveGapFilling(card, deckAnalysis)
	if curveScore >= 0.7 {
		reasons = append(reasons, fmt.Sprintf("fills curve gap at %d CMC", int(card.CMC)))
	}

	// Factor 3: Synergy with Deck (30%) - now captures synergy details
	synergyScore, cardSynergyDetails := s.scoreSynergyWithDeckDetailed(card, deckAnalysis)
	synergyDetails = append(synergyDetails, cardSynergyDetails...)
	if synergyScore >= 0.7 {
		reasons = append(reasons, "synergizes with deck strategy")
	}

	// Factor 4: Card Quality (15%)
	qualityScore := s.scoreCardQuality(card)
	if qualityScore >= 0.7 {
		reasons = append(reasons, "high-quality card")
	}

	// Factor 5: Standard Legality (5%)
	legalityScore := 1.0

	// Factor 6: Playability (5%)
	playabilityScore := 0.8

	// Calculate weighted score
	score := (colorScore * 0.25) +
		(curveScore * 0.20) +
		(synergyScore * 0.30) +
		(qualityScore * 0.15) +
		(legalityScore * 0.05) +
		(playabilityScore * 0.05)

	// Build score breakdown
	breakdown := &ScoreBreakdown{
		ColorFit: colorScore,
		CurveFit: curveScore,
		Synergy:  synergyScore,
		Quality:  qualityScore,
		Overall:  score,
	}

	// Build reasoning string
	reasoning := "This card "
	if len(reasons) == 0 {
		reasoning = "This card could work in your deck."
	} else {
		for i, r := range reasons {
			if i == 0 {
				reasoning += r
			} else if i == len(reasons)-1 {
				reasoning += ", and " + r
			} else {
				reasoning += ", " + r
			}
		}
		reasoning += "."
	}

	return score, reasoning, breakdown, synergyDetails
}

// scoreColorForDeck scores color compatibility against the deck's established colors.
func (s *SeedDeckBuilder) scoreColorForDeck(card *cards.Card, deckAnalysis *CollectiveDeckAnalysis) float64 {
	if len(card.Colors) == 0 {
		return 1.0 // Colorless fits any deck
	}

	if len(deckAnalysis.Colors) == 0 {
		return 0.8 // Deck is colorless, any color is fine
	}

	// Check how many card colors match deck colors
	matchingColors := 0
	for _, cardColor := range card.Colors {
		if deckAnalysis.Colors[cardColor] > 0 {
			matchingColors++
		}
	}

	if matchingColors == 0 {
		return 0.0 // No color overlap
	}

	if matchingColors == len(card.Colors) {
		return 1.0 // All card colors are in deck
	}

	return float64(matchingColors) / float64(len(card.Colors)) * 0.7
}

// scoreCurveGapFilling scores how well a card fills gaps in the current mana curve.
func (s *SeedDeckBuilder) scoreCurveGapFilling(card *cards.Card, deckAnalysis *CollectiveDeckAnalysis) float64 {
	if containsTypeInTypeLine(card.TypeLine, "Land") {
		return 0.5 // Neutral for lands
	}

	cmc := int(card.CMC)

	// Ideal curve targets for a 36-spell deck
	idealCurve := map[int]int{
		0: 2,
		1: 4,
		2: 8,
		3: 8,
		4: 6,
		5: 4,
		6: 2,
	}

	currentCount := deckAnalysis.ManaCurve[cmc]
	idealCount := idealCurve[cmc]
	if cmc > 6 {
		idealCount = 1 // Few high-CMC cards
	}

	if idealCount == 0 {
		return 0.3 // CMC we don't want many of
	}

	// Calculate how much we need this CMC
	need := idealCount - currentCount
	if need <= 0 {
		return 0.4 // Already have enough at this CMC
	}

	// More need = higher score
	needRatio := float64(need) / float64(idealCount)
	if needRatio > 1.0 {
		needRatio = 1.0
	}

	return 0.5 + (needRatio * 0.5)
}

// scoreSynergyWithDeckDetailed scores synergy and returns detailed synergy information.
func (s *SeedDeckBuilder) scoreSynergyWithDeckDetailed(card *cards.Card, deckAnalysis *CollectiveDeckAnalysis) (float64, []SynergyDetail) {
	synergy := 0.0
	synergyCount := 0
	details := make([]SynergyDetail, 0)

	// Extract card keywords
	var cardKeywords []KeywordInfo
	if card.OracleText != nil && *card.OracleText != "" {
		cardKeywords = ExtractKeywordsWithInfo(*card.OracleText)
	}

	// Keyword synergy with deck
	if len(cardKeywords) > 0 && len(deckAnalysis.Keywords) > 0 {
		keywordSynergy, matchedKeywords := CalculateKeywordSynergyDetailed(deckAnalysis.Keywords, cardKeywords)
		if keywordSynergy > 0 {
			synergy += keywordSynergy
			synergyCount++
			for _, kw := range matchedKeywords {
				details = append(details, SynergyDetail{
					Type:        "keyword",
					Name:        kw,
					Description: fmt.Sprintf("Shares %s with deck cards", kw),
				})
			}
		}
	}

	// Creature type synergy (tribal) - count ALL matching types for multi-type creatures
	if containsTypeInTypeLine(card.TypeLine, "Creature") && len(deckAnalysis.CreatureTypes) > 0 {
		cardCreatureTypes := extractCreatureTypesFromLine(card.TypeLine)
		for cardType := range cardCreatureTypes {
			if count := deckAnalysis.CreatureTypes[cardType]; count > 0 {
				synergy += 0.8 // Strong tribal synergy
				synergyCount++
				details = append(details, SynergyDetail{
					Type:        "creature_type",
					Name:        cardType,
					Description: fmt.Sprintf("%s tribal - matches %d cards in deck", cardType, count),
				})
				// No break - continue checking all creature types for multi-type synergy
			}
		}
	}

	// Theme synergy
	for _, kw := range cardKeywords {
		if kw.Category == CategoryTheme {
			if count := deckAnalysis.Themes[kw.Keyword]; count > 0 {
				synergy += 0.7
				synergyCount++
				details = append(details, SynergyDetail{
					Type:        "theme",
					Name:        kw.Keyword,
					Description: fmt.Sprintf("Supports %s theme (%d cards)", kw.Keyword, count),
				})
			}
		}
	}

	if synergyCount == 0 {
		return 0.5, details // Neutral score
	}

	avgSynergy := synergy / float64(synergyCount)
	if avgSynergy > 1.0 {
		avgSynergy = 1.0
	}

	return avgSynergy, details
}

// suggestLandsForDeck generates land suggestions based on current deck colors and size.
func (s *SeedDeckBuilder) suggestLandsForDeck(deckAnalysis *CollectiveDeckAnalysis) []*SuggestedLand {
	// Calculate recommended land count based on average CMC
	recommendedLands := 24
	avgCMC := 0.0
	totalNonLand := 0
	for cmc, count := range deckAnalysis.ManaCurve {
		avgCMC += float64(cmc) * float64(count)
		totalNonLand += count
	}
	if totalNonLand > 0 {
		avgCMC /= float64(totalNonLand)
		if avgCMC < 2.5 {
			recommendedLands = 22
		} else if avgCMC > 3.5 {
			recommendedLands = 26
		}
	}

	// Calculate how many lands are needed to reach 60 cards
	// Current non-land cards = deckAnalysis.TotalCards
	// Lands needed = 60 - currentCards, but capped at recommendedLands
	currentCards := deckAnalysis.TotalCards
	landsToAdd := 60 - currentCards
	if landsToAdd > recommendedLands {
		landsToAdd = recommendedLands
	}
	if landsToAdd < 0 {
		landsToAdd = 0
	}

	totalColorWeight := 0
	for _, count := range deckAnalysis.Colors {
		totalColorWeight += count
	}

	lands := make([]*SuggestedLand, 0)
	if totalColorWeight == 0 || landsToAdd == 0 {
		return lands
	}

	for color, count := range deckAnalysis.Colors {
		land, ok := basicLandsByColor[color]
		if !ok {
			continue
		}

		proportion := float64(count) / float64(totalColorWeight)
		quantity := int(proportion*float64(landsToAdd) + 0.5)
		if quantity < 1 && count > 0 {
			quantity = 1
		}

		lands = append(lands, &SuggestedLand{
			CardID:   land.ArenaID,
			Name:     land.Name,
			Quantity: quantity,
			Color:    color,
		})
	}

	return lands
}

// buildLiveDeckAnalysis builds real-time analysis for the response.
func (s *SeedDeckBuilder) buildLiveDeckAnalysis(deckAnalysis *CollectiveDeckAnalysis, collection map[int]int, cardIDs []int) *LiveDeckAnalysis {
	// Extract top colors
	colors := make([]string, 0)
	for color := range deckAnalysis.Colors {
		colors = append(colors, color)
	}

	// Extract top keywords
	keywordCounts := make(map[string]int)
	for _, kw := range deckAnalysis.Keywords {
		keywordCounts[kw.Keyword]++
	}
	keywords := make([]string, 0)
	for kw, count := range keywordCounts {
		if count >= 2 { // Only include keywords that appear multiple times
			keywords = append(keywords, kw)
		}
	}

	// Extract top themes
	themes := make([]string, 0)
	for theme, count := range deckAnalysis.Themes {
		if count >= 2 {
			themes = append(themes, theme)
		}
	}

	// Calculate recommended land count based on curve
	recommendedLands := 24
	avgCMC := 0.0
	totalNonLand := 0
	for cmc, count := range deckAnalysis.ManaCurve {
		avgCMC += float64(cmc) * float64(count)
		totalNonLand += count
	}
	if totalNonLand > 0 {
		avgCMC /= float64(totalNonLand)
		// Adjust lands based on average CMC
		if avgCMC < 2.5 {
			recommendedLands = 22
		} else if avgCMC > 3.5 {
			recommendedLands = 26
		}
	}

	// Calculate how many deck cards are in the collection
	inCollectionCount := 0
	for _, cardID := range cardIDs {
		if collection[cardID] > 0 {
			inCollectionCount++
		}
	}

	return &LiveDeckAnalysis{
		ColorIdentity:        colors,
		Keywords:             keywords,
		Themes:               themes,
		CurrentCurve:         deckAnalysis.ManaCurve,
		RecommendedLandCount: recommendedLands,
		TotalCards:           deckAnalysis.TotalCards,
		InCollectionCount:    inCollectionCount,
	}
}

// calculateRecommendedCopies determines optimal copy count (1-4) for a card.
// Considers: legendary status, mana cost, card type, synergy score.
func (s *SeedDeckBuilder) calculateRecommendedCopies(card *CardWithOwnership) int {
	typeLine := strings.ToLower(card.TypeLine)

	// Legendary cards: usually 2-3 copies (can't have multiples in play)
	if strings.Contains(typeLine, "legendary") {
		if card.CMC >= 5 {
			return 2 // Expensive legendaries: 2 copies
		}
		return 3 // Cheaper legendaries: 3 copies for consistency
	}

	// Planeswalkers: usually 2-3 copies
	if strings.Contains(typeLine, "planeswalker") {
		if card.CMC >= 5 {
			return 2
		}
		return 3
	}

	// High-synergy cards (score > 0.8): want full playsets
	if card.Score > 0.8 {
		return 4
	}

	// Expensive cards (5+ CMC): usually 2-3 copies
	if card.CMC >= 5 {
		return 2
	}
	if card.CMC == 4 {
		return 3
	}

	// Instants and sorceries: usually 3-4 copies for consistency
	if strings.Contains(typeLine, "instant") || strings.Contains(typeLine, "sorcery") {
		if card.Score > 0.6 {
			return 4
		}
		return 3
	}

	// Creatures and other permanents: 3-4 copies based on score
	if card.Score > 0.6 {
		return 4
	}
	return 3
}

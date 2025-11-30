package recommendations

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// ColorCombination represents a mono or two-color combination.
type ColorCombination struct {
	Colors []string `json:"colors"` // e.g., ["W", "U"] or ["R"] for mono
	Name   string   `json:"name"`   // e.g., "Azorius" or "Mono-Red"
}

// SuggestedDeck represents a complete deck suggestion.
type SuggestedDeck struct {
	ColorCombo ColorCombination        `json:"colorCombo"`
	Spells     []*SuggestedCard        `json:"spells"` // 23 non-land cards
	Lands      []*SuggestedLand        `json:"lands"`  // 17 basic lands
	TotalCards int                     `json:"totalCards"`
	Score      float64                 `json:"score"`     // 0.0-1.0 overall quality
	Viability  string                  `json:"viability"` // "strong", "viable", "weak"
	Analysis   *DeckSuggestionAnalysis `json:"analysis"`
}

// SuggestedCard represents a card in the suggested deck.
type SuggestedCard struct {
	CardID    int      `json:"cardID"`
	Name      string   `json:"name"`
	TypeLine  string   `json:"typeLine"`
	ManaCost  string   `json:"manaCost,omitempty"`
	ImageURI  string   `json:"imageURI,omitempty"`
	CMC       int      `json:"cmc"`
	Colors    []string `json:"colors"`
	Rarity    string   `json:"rarity,omitempty"`
	Score     float64  `json:"score"`
	Reasoning string   `json:"reasoning"`
}

// SuggestedLand represents basic lands in the suggestion.
type SuggestedLand struct {
	CardID   int    `json:"cardID"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Color    string `json:"color"`
}

// DeckSuggestionAnalysis provides deck composition details.
type DeckSuggestionAnalysis struct {
	CreatureCount     int            `json:"creatureCount"`
	SpellCount        int            `json:"spellCount"` // Non-creature, non-land
	AverageCMC        float64        `json:"averageCMC"`
	ManaCurve         map[int]int    `json:"manaCurve"`
	ColorDistribution map[string]int `json:"colorDistribution"`
	TopCards          []string       `json:"topCards"`  // Names of best cards
	Synergies         []string       `json:"synergies"` // Detected synergies
	PlayableCount     int            `json:"playableCount"`
}

// SuggestDecksResponse contains all viable deck suggestions.
type SuggestDecksResponse struct {
	Suggestions  []*SuggestedDeck  `json:"suggestions"`
	TotalCombos  int               `json:"totalCombos"`  // Always 15 (5 mono + 10 two-color)
	ViableCombos int               `json:"viableCombos"` // How many are playable
	BestCombo    *ColorCombination `json:"bestCombo,omitempty"`
	Error        string            `json:"error,omitempty"`
}

// DeckSuggester generates deck suggestions from a draft pool.
type DeckSuggester struct {
	engine      *RuleBasedEngine
	cardService *cards.Service
	setCardRepo repository.SetCardRepository
	ratingsRepo repository.DraftRatingsRepository
}

// NewDeckSuggester creates a new DeckSuggester instance.
func NewDeckSuggester(
	engine *RuleBasedEngine,
	cardService *cards.Service,
	setCardRepo repository.SetCardRepository,
	ratingsRepo repository.DraftRatingsRepository,
) *DeckSuggester {
	return &DeckSuggester{
		engine:      engine,
		cardService: cardService,
		setCardRepo: setCardRepo,
		ratingsRepo: ratingsRepo,
	}
}

// All color combinations to evaluate (5 mono + 10 two-color = 15 total).
var allColorCombinations = []ColorCombination{
	// Mono-color
	{Colors: []string{"W"}, Name: "Mono-White"},
	{Colors: []string{"U"}, Name: "Mono-Blue"},
	{Colors: []string{"B"}, Name: "Mono-Black"},
	{Colors: []string{"R"}, Name: "Mono-Red"},
	{Colors: []string{"G"}, Name: "Mono-Green"},
	// Allied pairs
	{Colors: []string{"W", "U"}, Name: "Azorius"},
	{Colors: []string{"U", "B"}, Name: "Dimir"},
	{Colors: []string{"B", "R"}, Name: "Rakdos"},
	{Colors: []string{"R", "G"}, Name: "Gruul"},
	{Colors: []string{"G", "W"}, Name: "Selesnya"},
	// Enemy pairs
	{Colors: []string{"W", "B"}, Name: "Orzhov"},
	{Colors: []string{"U", "R"}, Name: "Izzet"},
	{Colors: []string{"B", "G"}, Name: "Golgari"},
	{Colors: []string{"R", "W"}, Name: "Boros"},
	{Colors: []string{"G", "U"}, Name: "Simic"},
}

// Basic land Arena IDs.
var basicLandsByColor = map[string]struct {
	ArenaID int
	Name    string
}{
	"W": {ArenaID: 81716, Name: "Plains"},
	"U": {ArenaID: 81717, Name: "Island"},
	"B": {ArenaID: 81718, Name: "Swamp"},
	"R": {ArenaID: 81719, Name: "Mountain"},
	"G": {ArenaID: 81720, Name: "Forest"},
}

// scoredCard holds a card with its calculated score.
type scoredCard struct {
	card      *cards.Card
	score     float64
	reasoning string
}

// SuggestDecks generates all viable deck suggestions for a draft pool.
func (s *DeckSuggester) SuggestDecks(
	ctx context.Context,
	draftPool []int,
	setCode string,
	draftFormat string,
) (*SuggestDecksResponse, error) {
	if len(draftPool) == 0 {
		return &SuggestDecksResponse{
			Error: "No cards in draft pool",
		}, nil
	}

	// Load all cards from the draft pool
	poolCards := make([]*cards.Card, 0, len(draftPool))
	for _, arenaID := range draftPool {
		card := s.getCard(ctx, arenaID)
		if card != nil {
			poolCards = append(poolCards, card)
		}
	}

	if len(poolCards) == 0 {
		return &SuggestDecksResponse{
			Error: "Could not load any cards from draft pool",
		}, nil
	}

	log.Printf("SuggestDecks: Loaded %d cards from pool", len(poolCards))

	// Evaluate each color combination
	suggestions := make([]*SuggestedDeck, 0)
	for _, combo := range allColorCombinations {
		suggestion := s.evaluateColorCombination(ctx, combo, poolCards, setCode, draftFormat)
		if suggestion != nil {
			suggestions = append(suggestions, suggestion)
		}
	}

	// Sort suggestions by score (descending)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	response := &SuggestDecksResponse{
		Suggestions:  suggestions,
		TotalCombos:  len(allColorCombinations),
		ViableCombos: len(suggestions),
	}

	if len(suggestions) > 0 {
		response.BestCombo = &suggestions[0].ColorCombo
	}

	log.Printf("SuggestDecks: Found %d viable color combinations", len(suggestions))
	return response, nil
}

// evaluateColorCombination evaluates a single color combination and returns a deck suggestion if viable.
func (s *DeckSuggester) evaluateColorCombination(
	ctx context.Context,
	combo ColorCombination,
	poolCards []*cards.Card,
	setCode string,
	draftFormat string,
) *SuggestedDeck {
	// Filter cards that fit this color combination
	candidates := s.filterByColorFit(poolCards, combo)

	// Check viability
	if !s.isViable(candidates) {
		return nil
	}

	// Score all candidates
	scoredCards := make([]*scoredCard, 0, len(candidates))
	for _, card := range candidates {
		score, reasoning := s.scoreCardForDeck(ctx, card, candidates, setCode, draftFormat)
		scoredCards = append(scoredCards, &scoredCard{
			card:      card,
			score:     score,
			reasoning: reasoning,
		})
	}

	// Select best 23 cards with curve constraints
	selectedCards := s.selectBestCards(scoredCards, 23)

	// Distribute 17 lands
	lands := s.distributeLands(selectedCards, combo)

	// Build the suggested deck
	spells := make([]*SuggestedCard, len(selectedCards))
	for i, sc := range selectedCards {
		spells[i] = s.toSuggestedCard(sc)
	}

	// Calculate deck analysis
	analysis := s.analyzeDeckSuggestion(selectedCards, candidates)

	// Calculate overall deck score
	deckScore := s.calculateDeckScore(selectedCards, analysis)

	// Determine viability status
	viability := s.determineViability(deckScore, analysis)

	return &SuggestedDeck{
		ColorCombo: combo,
		Spells:     spells,
		Lands:      lands,
		TotalCards: len(spells) + s.countLands(lands),
		Score:      deckScore,
		Viability:  viability,
		Analysis:   analysis,
	}
}

// filterByColorFit returns cards that fit the given color combination.
// A card fits if all its colors are within the combination (or it's colorless).
func (s *DeckSuggester) filterByColorFit(poolCards []*cards.Card, combo ColorCombination) []*cards.Card {
	result := make([]*cards.Card, 0)
	comboColors := make(map[string]bool)
	for _, c := range combo.Colors {
		comboColors[c] = true
	}

	for _, card := range poolCards {
		// Skip lands - we'll add basic lands separately
		if containsTypeInTypeLine(card.TypeLine, "Land") {
			continue
		}

		// Colorless cards always fit
		if len(card.Colors) == 0 {
			result = append(result, card)
			continue
		}

		// Check if all card colors are in the combination
		fits := true
		for _, cardColor := range card.Colors {
			if !comboColors[cardColor] {
				fits = false
				break
			}
		}

		if fits {
			result = append(result, card)
		}
	}

	return result
}

// isViable checks if a color combination has enough playables.
func (s *DeckSuggester) isViable(candidates []*cards.Card) bool {
	// Need at least 18 playable spells
	if len(candidates) < 18 {
		return false
	}

	// Need at least 10 creatures
	creatureCount := 0
	for _, card := range candidates {
		if containsTypeInTypeLine(card.TypeLine, "Creature") {
			creatureCount++
		}
	}

	return creatureCount >= 10
}

// scoreCardForDeck scores a card specifically for deck building.
// Uses modified weights: Quality (40%), ManaCurve (30%), Synergy (20%), ColorFit (10%).
func (s *DeckSuggester) scoreCardForDeck(
	ctx context.Context,
	card *cards.Card,
	poolCards []*cards.Card,
	setCode string,
	draftFormat string,
) (float64, string) {
	reasons := make([]string, 0)

	// Factor 1: Quality from 17Lands (40% weight)
	qualityScore := s.scoreQuality(ctx, card, setCode, draftFormat)
	if qualityScore >= 0.7 {
		reasons = append(reasons, "high-quality card")
	}

	// Factor 2: Mana curve fit (30% weight)
	curveScore := s.scoreCurve(card, poolCards)
	if curveScore >= 0.7 {
		reasons = append(reasons, fmt.Sprintf("good %d-drop", int(card.CMC)))
	}

	// Factor 3: Synergy with pool (20% weight)
	synergyScore := s.scoreSynergyWithPool(card, poolCards)
	if synergyScore >= 0.6 {
		reasons = append(reasons, "synergizes with pool")
	}

	// Factor 4: Color fit bonus (10% weight)
	// Mono-color cards score higher than multi-color to avoid mana issues
	colorFitScore := 1.0
	if len(card.Colors) > 1 {
		colorFitScore = 0.85
	}

	// Calculate weighted score
	score := (qualityScore * 0.40) +
		(curveScore * 0.30) +
		(synergyScore * 0.20) +
		(colorFitScore * 0.10)

	reasoning := "Standard playable"
	if len(reasons) > 0 {
		reasoning = strings.Join(reasons, ", ")
	}

	return score, reasoning
}

// scoreQuality returns the quality score based on 17Lands data.
func (s *DeckSuggester) scoreQuality(ctx context.Context, card *cards.Card, setCode, draftFormat string) float64 {
	if s.ratingsRepo == nil || setCode == "" || draftFormat == "" {
		return s.fallbackQualityScore(card)
	}

	arenaIDStr := fmt.Sprintf("%d", card.ArenaID)
	rating, err := s.ratingsRepo.GetCardRatingByArenaID(ctx, setCode, draftFormat, arenaIDStr)
	if err != nil || rating == nil {
		return s.fallbackQualityScore(card)
	}

	// Normalize GIHWR: 45% -> 0.0, 55% -> 0.5, 65%+ -> 1.0
	gihScore := (rating.GIHWR - 0.45) / 0.20
	if gihScore < 0 {
		gihScore = 0
	} else if gihScore > 1 {
		gihScore = 1
	}

	return gihScore
}

// fallbackQualityScore returns quality based on rarity when ratings unavailable.
func (s *DeckSuggester) fallbackQualityScore(card *cards.Card) float64 {
	switch strings.ToLower(card.Rarity) {
	case "mythic":
		return 0.85
	case "rare":
		return 0.75
	case "uncommon":
		return 0.60
	default:
		return 0.50
	}
}

// scoreCurve scores how well a card fits the mana curve.
func (s *DeckSuggester) scoreCurve(card *cards.Card, poolCards []*cards.Card) float64 {
	cmc := int(card.CMC)

	// Count cards at each CMC in the pool (non-land)
	cmcCounts := make(map[int]int)
	for _, c := range poolCards {
		if !containsTypeInTypeLine(c.TypeLine, "Land") {
			cmcCounts[int(c.CMC)]++
		}
	}

	// Ideal curve for limited (23 non-land cards)
	idealCounts := map[int]int{
		1: 2,
		2: 5,
		3: 5,
		4: 4,
		5: 3,
		6: 2,
		7: 2,
	}

	ideal := idealCounts[cmc]
	if cmc > 7 {
		ideal = 1 // Very high CMC cards are rarely needed
	}

	current := cmcCounts[cmc]

	// Score higher for CMC slots that are underfilled
	if current < ideal {
		gap := ideal - current
		return 0.7 + (float64(gap) * 0.1)
	} else if current == ideal {
		return 0.6
	} else {
		// Over-filled slot
		return 0.4
	}
}

// scoreSynergyWithPool calculates synergy with other cards in the pool.
func (s *DeckSuggester) scoreSynergyWithPool(card *cards.Card, poolCards []*cards.Card) float64 {
	if card.OracleText == nil {
		return 0.5
	}

	cardKeywords := extractKeywordsFromText(*card.OracleText)
	if len(cardKeywords) == 0 {
		return 0.5
	}

	// Count how many other cards share keywords
	synergyCount := 0
	for _, other := range poolCards {
		if other.ArenaID == card.ArenaID || other.OracleText == nil {
			continue
		}
		otherKeywords := extractKeywordsFromText(*other.OracleText)
		for kw := range cardKeywords {
			if otherKeywords[kw] {
				synergyCount++
				break
			}
		}
	}

	// Normalize: 5+ synergistic cards = max score
	score := float64(synergyCount) / 5.0
	if score > 1.0 {
		score = 1.0
	}

	return score*0.5 + 0.5 // Range: 0.5 to 1.0
}

// selectBestCards selects the best cards with curve constraints using greedy selection.
func (s *DeckSuggester) selectBestCards(scoredCards []*scoredCard, targetCount int) []*scoredCard {
	// Sort by score descending
	sort.Slice(scoredCards, func(i, j int) bool {
		return scoredCards[i].score > scoredCards[j].score
	})

	// Curve constraints for 23 non-land cards
	curveSlots := map[int]int{
		0: 0, 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0,
	}
	idealCurve := map[int]int{
		0: 0, // No 0-cost spells typically
		1: 2,
		2: 5,
		3: 5,
		4: 4,
		5: 3,
		6: 2,
		7: 2, // 7+ combined
	}

	selected := make([]*scoredCard, 0, targetCount)
	usedCards := make(map[int]bool)

	// First pass: fill curve gaps with best available
	for _, sc := range scoredCards {
		if len(selected) >= targetCount {
			break
		}
		if usedCards[sc.card.ArenaID] {
			continue
		}

		cmc := int(sc.card.CMC)
		if cmc > 7 {
			cmc = 7 // Cap at 7 for curve tracking
		}

		if curveSlots[cmc] < idealCurve[cmc] {
			selected = append(selected, sc)
			usedCards[sc.card.ArenaID] = true
			curveSlots[cmc]++
		}
	}

	// Second pass: fill remaining with highest scores
	for _, sc := range scoredCards {
		if len(selected) >= targetCount {
			break
		}
		if usedCards[sc.card.ArenaID] {
			continue
		}

		selected = append(selected, sc)
		usedCards[sc.card.ArenaID] = true
	}

	return selected
}

// distributeLands calculates the land distribution for a deck.
func (s *DeckSuggester) distributeLands(selectedCards []*scoredCard, combo ColorCombination) []*SuggestedLand {
	// Count mana pips in selected cards
	pipCounts := make(map[string]int)
	for _, color := range combo.Colors {
		pipCounts[color] = 0
	}

	for _, sc := range selectedCards {
		if sc.card.ManaCost == nil {
			continue
		}
		manaCost := *sc.card.ManaCost
		for _, color := range combo.Colors {
			// Count occurrences of each color symbol in mana cost
			pipCounts[color] += strings.Count(manaCost, "{"+color+"}")
		}
	}

	// Calculate total pips
	totalPips := 0
	for _, count := range pipCounts {
		totalPips += count
	}

	lands := make([]*SuggestedLand, 0)
	const totalLands = 17

	if totalPips == 0 {
		// Equal distribution for mono-color or colorless-heavy decks
		if len(combo.Colors) == 1 {
			color := combo.Colors[0]
			land := basicLandsByColor[color]
			lands = append(lands, &SuggestedLand{
				CardID:   land.ArenaID,
				Name:     land.Name,
				Quantity: totalLands,
				Color:    color,
			})
		} else {
			// Split evenly for two colors
			firstLand := basicLandsByColor[combo.Colors[0]]
			secondLand := basicLandsByColor[combo.Colors[1]]
			lands = append(lands, &SuggestedLand{
				CardID:   firstLand.ArenaID,
				Name:     firstLand.Name,
				Quantity: 9,
				Color:    combo.Colors[0],
			})
			lands = append(lands, &SuggestedLand{
				CardID:   secondLand.ArenaID,
				Name:     secondLand.Name,
				Quantity: 8,
				Color:    combo.Colors[1],
			})
		}
		return lands
	}

	// Proportional distribution
	landCounts := make(map[string]int)
	allocated := 0

	for color, pips := range pipCounts {
		proportion := float64(pips) / float64(totalPips)
		count := int(float64(totalLands) * proportion)
		landCounts[color] = count
		allocated += count
	}

	// Distribute remaining lands to highest pip color
	remaining := totalLands - allocated
	if remaining > 0 {
		maxPips := 0
		maxColor := combo.Colors[0]
		for color, pips := range pipCounts {
			if pips > maxPips {
				maxPips = pips
				maxColor = color
			}
		}
		landCounts[maxColor] += remaining
	}

	// Build land list
	for color, count := range landCounts {
		if count > 0 {
			land := basicLandsByColor[color]
			lands = append(lands, &SuggestedLand{
				CardID:   land.ArenaID,
				Name:     land.Name,
				Quantity: count,
				Color:    color,
			})
		}
	}

	return lands
}

// toSuggestedCard converts a scoredCard to SuggestedCard.
func (s *DeckSuggester) toSuggestedCard(sc *scoredCard) *SuggestedCard {
	manaCost := ""
	if sc.card.ManaCost != nil {
		manaCost = *sc.card.ManaCost
	}
	imageURI := ""
	if sc.card.ImageURI != nil {
		imageURI = *sc.card.ImageURI
	}

	return &SuggestedCard{
		CardID:    sc.card.ArenaID,
		Name:      sc.card.Name,
		TypeLine:  sc.card.TypeLine,
		ManaCost:  manaCost,
		ImageURI:  imageURI,
		CMC:       int(sc.card.CMC),
		Colors:    sc.card.Colors,
		Rarity:    sc.card.Rarity,
		Score:     sc.score,
		Reasoning: sc.reasoning,
	}
}

// analyzeDeckSuggestion generates analysis for the suggested deck.
func (s *DeckSuggester) analyzeDeckSuggestion(selectedCards []*scoredCard, allCandidates []*cards.Card) *DeckSuggestionAnalysis {
	analysis := &DeckSuggestionAnalysis{
		ManaCurve:         make(map[int]int),
		ColorDistribution: make(map[string]int),
		TopCards:          make([]string, 0),
		Synergies:         make([]string, 0),
		PlayableCount:     len(allCandidates),
	}

	totalCMC := 0.0
	for _, sc := range selectedCards {
		card := sc.card

		// Count creatures vs spells
		if containsTypeInTypeLine(card.TypeLine, "Creature") {
			analysis.CreatureCount++
		} else {
			analysis.SpellCount++
		}

		// Mana curve
		cmc := int(card.CMC)
		analysis.ManaCurve[cmc]++
		totalCMC += card.CMC

		// Color distribution
		for _, color := range card.Colors {
			analysis.ColorDistribution[color]++
		}
	}

	if len(selectedCards) > 0 {
		analysis.AverageCMC = totalCMC / float64(len(selectedCards))
	}

	// Top 3 cards by score
	for i := 0; i < 3 && i < len(selectedCards); i++ {
		analysis.TopCards = append(analysis.TopCards, selectedCards[i].card.Name)
	}

	// Detect synergies
	synergies := s.detectSynergies(selectedCards)
	analysis.Synergies = synergies

	return analysis
}

// detectSynergies identifies synergy themes in the deck.
func (s *DeckSuggester) detectSynergies(selectedCards []*scoredCard) []string {
	synergies := make([]string, 0)
	keywordCounts := make(map[string]int)

	for _, sc := range selectedCards {
		if sc.card.OracleText == nil {
			continue
		}
		keywords := extractKeywordsFromText(*sc.card.OracleText)
		for kw := range keywords {
			keywordCounts[kw]++
		}
	}

	// Report themes with 3+ cards
	themeNames := map[string]string{
		"flying":         "Flying",
		"tokens":         "Tokens",
		"+1/+1 counters": "+1/+1 Counters",
		"graveyard":      "Graveyard",
		"sacrifice":      "Sacrifice",
		"card draw":      "Card Draw",
		"lifegain":       "Lifegain",
		"ETB":            "Enter the Battlefield",
		"cast triggers":  "Spells Matter",
		"death triggers": "Death Triggers",
	}

	for keyword, count := range keywordCounts {
		if count >= 3 {
			name := keyword
			if friendlyName, ok := themeNames[keyword]; ok {
				name = friendlyName
			}
			synergies = append(synergies, fmt.Sprintf("%s (%d cards)", name, count))
		}
	}

	return synergies
}

// calculateDeckScore calculates the overall deck quality score.
func (s *DeckSuggester) calculateDeckScore(selectedCards []*scoredCard, analysis *DeckSuggestionAnalysis) float64 {
	if len(selectedCards) == 0 {
		return 0.0
	}

	// Average card quality (60% weight)
	totalScore := 0.0
	for _, sc := range selectedCards {
		totalScore += sc.score
	}
	avgCardScore := totalScore / float64(len(selectedCards))

	// Creature ratio score (20% weight)
	// Ideal: 14-17 creatures in a 23-spell deck
	creatureRatio := float64(analysis.CreatureCount) / 23.0
	creatureScore := 0.0
	if creatureRatio >= 0.6 && creatureRatio <= 0.75 {
		creatureScore = 1.0
	} else if creatureRatio >= 0.5 && creatureRatio <= 0.8 {
		creatureScore = 0.7
	} else {
		creatureScore = 0.4
	}

	// Curve score (20% weight)
	// Check if we have cards at each important CMC
	curveScore := 0.0
	has2Drop := analysis.ManaCurve[2] >= 3
	has3Drop := analysis.ManaCurve[3] >= 3
	has4Drop := analysis.ManaCurve[4] >= 2
	if has2Drop && has3Drop && has4Drop {
		curveScore = 1.0
	} else if has2Drop && has3Drop {
		curveScore = 0.7
	} else if has2Drop || has3Drop {
		curveScore = 0.5
	} else {
		curveScore = 0.3
	}

	// Weighted total
	deckScore := (avgCardScore * 0.60) + (creatureScore * 0.20) + (curveScore * 0.20)

	return deckScore
}

// determineViability returns the viability status based on score and analysis.
func (s *DeckSuggester) determineViability(score float64, analysis *DeckSuggestionAnalysis) string {
	if score >= 0.7 && analysis.CreatureCount >= 13 && analysis.PlayableCount >= 25 {
		return "strong"
	} else if score >= 0.5 && analysis.CreatureCount >= 10 {
		return "viable"
	}
	return "weak"
}

// countLands counts total lands from the land distribution.
func (s *DeckSuggester) countLands(lands []*SuggestedLand) int {
	total := 0
	for _, land := range lands {
		total += land.Quantity
	}
	return total
}

// getCard retrieves a card by Arena ID from the SetCardRepo or CardService.
func (s *DeckSuggester) getCard(ctx context.Context, arenaID int) *cards.Card {
	arenaIDStr := fmt.Sprintf("%d", arenaID)

	// Try SetCardRepo first (faster)
	if s.setCardRepo != nil {
		setCard, err := s.setCardRepo.GetCardByArenaID(ctx, arenaIDStr)
		if err == nil && setCard != nil {
			return convertSetCardToCardsCard(setCard)
		}
	}

	// Fallback to CardService
	if s.cardService != nil {
		card, err := s.cardService.GetCard(arenaID)
		if err == nil && card != nil {
			return card
		}
	}

	return nil
}

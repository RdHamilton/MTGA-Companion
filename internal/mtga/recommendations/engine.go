package recommendations

import (
	"context"
	"fmt"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// RecommendationEngine provides intelligent card recommendations for deck building.
type RecommendationEngine interface {
	// GetRecommendations returns recommended cards for a deck based on filters.
	GetRecommendations(ctx context.Context, deck *DeckContext, filters *Filters) ([]*CardRecommendation, error)

	// ExplainRecommendation explains why a specific card is recommended for a deck.
	ExplainRecommendation(ctx context.Context, cardID int, deck *DeckContext) (string, error)

	// RecordAcceptance records whether a recommendation was accepted/rejected (for learning).
	RecordAcceptance(ctx context.Context, deckID string, cardID int, accepted bool) error
}

// DeckContext contains all information about a deck needed for recommendations.
type DeckContext struct {
	Deck         *models.Deck
	Cards        []*models.DeckCard
	CardMetadata map[int]*cards.Card
	DraftCardIDs []int  // Available cards for draft decks (nil for constructed)
	Format       string // "Limited", "Standard", "Historic", etc.
	SetCode      string // Set code for draft ratings (e.g., "BLB")
	DraftFormat  string // Draft format for ratings (e.g., "PremierDraft")
}

// Filters control which recommendations are returned.
type Filters struct {
	MaxResults    int      // Maximum number of recommendations (default: 10)
	MinScore      float64  // Minimum score threshold (0.0-1.0, default: 0.3)
	Colors        []string // Filter by colors (nil = any color)
	CardTypes     []string // Filter by card types (nil = any type)
	CMCRange      *CMCRange
	IncludeLands  bool  // Include land recommendations
	OnlyDraftPool bool  // Only recommend cards from draft pool
	DraftPool     []int // Available cards in draft pool (only used when OnlyDraftPool is true)
}

// CMCRange filters by converted mana cost.
type CMCRange struct {
	Min int
	Max int
}

// CardRecommendation represents a recommended card with scoring and explanation.
type CardRecommendation struct {
	Card       *cards.Card
	Score      float64 // 0.0-1.0 overall recommendation score
	Reasoning  string  // Human-readable explanation
	Source     string  // "color-fit", "mana-curve", "synergy", "17lands", "quality"
	Confidence float64 // How confident the system is (0.0-1.0)
	Factors    *ScoreFactors
}

// ScoreFactors breaks down the recommendation score components.
type ScoreFactors struct {
	ColorFit  float64 // How well colors match deck (0.0-1.0)
	ManaCurve float64 // How well CMC fills curve gaps (0.0-1.0)
	Synergy   float64 // Synergy with existing cards (0.0-1.0)
	Quality   float64 // Intrinsic card quality/rating (0.0-1.0)
	Playable  float64 // Playability in format/context (0.0-1.0)
}

// RuleBasedEngine implements a rule-based recommendation system.
type RuleBasedEngine struct {
	cardService    *cards.Service
	setCardRepo    repository.SetCardRepository // For faster lookups from local DB
	ratingsRepo    repository.DraftRatingsRepository
	cfbRatingsRepo repository.CFBRatingsRepository // ChannelFireball ratings (optional)
}

// NewRuleBasedEngine creates a new rule-based recommendation engine.
func NewRuleBasedEngine(cardService *cards.Service, ratingsRepo repository.DraftRatingsRepository) *RuleBasedEngine {
	return &RuleBasedEngine{
		cardService: cardService,
		ratingsRepo: ratingsRepo,
	}
}

// NewRuleBasedEngineWithSetRepo creates a new rule-based recommendation engine with SetCardRepo support.
func NewRuleBasedEngineWithSetRepo(cardService *cards.Service, setCardRepo repository.SetCardRepository, ratingsRepo repository.DraftRatingsRepository) *RuleBasedEngine {
	return &RuleBasedEngine{
		cardService: cardService,
		setCardRepo: setCardRepo,
		ratingsRepo: ratingsRepo,
	}
}

// NewRuleBasedEngineWithCFB creates a new rule-based recommendation engine with CFB ratings support.
func NewRuleBasedEngineWithCFB(cardService *cards.Service, setCardRepo repository.SetCardRepository, ratingsRepo repository.DraftRatingsRepository, cfbRatingsRepo repository.CFBRatingsRepository) *RuleBasedEngine {
	return &RuleBasedEngine{
		cardService:    cardService,
		setCardRepo:    setCardRepo,
		ratingsRepo:    ratingsRepo,
		cfbRatingsRepo: cfbRatingsRepo,
	}
}

// SetCFBRatingsRepo sets the CFB ratings repository for the engine.
func (e *RuleBasedEngine) SetCFBRatingsRepo(cfbRatingsRepo repository.CFBRatingsRepository) {
	e.cfbRatingsRepo = cfbRatingsRepo
}

// GetRecommendations returns recommended cards based on deck analysis.
func (e *RuleBasedEngine) GetRecommendations(ctx context.Context, deck *DeckContext, filters *Filters) ([]*CardRecommendation, error) {
	if e == nil {
		return nil, fmt.Errorf("engine is nil")
	}
	if e.cardService == nil {
		return nil, fmt.Errorf("card service is not initialized")
	}
	if deck == nil {
		return nil, fmt.Errorf("deck context is nil")
	}

	// Apply default filters
	if filters == nil {
		filters = &Filters{
			MaxResults:   10,
			MinScore:     0.3,
			IncludeLands: true,
		}
	}

	// Analyze deck composition
	analysis := analyzeDeck(deck)

	// Get candidate cards
	var candidates []*cards.Card
	if filters.OnlyDraftPool && len(filters.DraftPool) > 0 {
		// For draft decks, only consider draft pool
		// Try SetCardRepo first (faster, has cards from log parsing)
		if e.setCardRepo != nil {
			candidates = make([]*cards.Card, 0, len(filters.DraftPool))
			for _, arenaID := range filters.DraftPool {
				setCard, err := e.setCardRepo.GetCardByArenaID(ctx, fmt.Sprintf("%d", arenaID))
				if err == nil && setCard != nil {
					// Convert setCard to cards.Card
					card := convertSetCardToCardsCard(setCard)
					candidates = append(candidates, card)
				}
			}
		} else {
			// Fallback to CardService (Scryfall API)
			cardMap, err := e.cardService.GetCards(filters.DraftPool)
			if err != nil {
				return nil, fmt.Errorf("failed to get draft cards: %w", err)
			}
			candidates = make([]*cards.Card, 0, len(cardMap))
			for _, card := range cardMap {
				candidates = append(candidates, card)
			}
		}
	} else {
		// For constructed, we'd query all available cards
		// For now, return empty since we need card database integration
		return []*CardRecommendation{}, nil
	}

	// Score each candidate
	recommendations := make([]*CardRecommendation, 0)
	for _, card := range candidates {
		// Skip cards already in deck
		if isCardInDeck(card.ArenaID, deck.Cards) {
			continue
		}

		// Apply filters
		if !matchesFilters(card, filters) {
			continue
		}

		// Calculate recommendation score
		rec := e.scoreCard(ctx, card, deck, analysis)

		// Apply score threshold
		if rec.Score >= filters.MinScore {
			recommendations = append(recommendations, rec)
		}
	}

	// Sort by score (descending)
	sortRecommendations(recommendations)

	// Limit results
	if len(recommendations) > filters.MaxResults {
		recommendations = recommendations[:filters.MaxResults]
	}

	return recommendations, nil
}

// ExplainRecommendation generates a detailed explanation for a card recommendation.
func (e *RuleBasedEngine) ExplainRecommendation(ctx context.Context, cardID int, deck *DeckContext) (string, error) {
	card, err := e.cardService.GetCard(cardID)
	if err != nil {
		return "", fmt.Errorf("failed to get card: %w", err)
	}

	analysis := analyzeDeck(deck)
	rec := e.scoreCard(ctx, card, deck, analysis)

	return rec.Reasoning, nil
}

// RecordAcceptance records recommendation acceptance (for future ML training).
func (e *RuleBasedEngine) RecordAcceptance(ctx context.Context, deckID string, cardID int, accepted bool) error {
	// For Phase 1A, this is a no-op
	// In Phase 1B, we'll store this in database for analysis
	// In Phase 1C, we'll use this for ML training
	return nil
}

// scoreCard calculates recommendation score for a card.
func (e *RuleBasedEngine) scoreCard(ctx context.Context, card *cards.Card, deck *DeckContext, analysis *DeckAnalysis) *CardRecommendation {
	factors := &ScoreFactors{}

	// Factor 1: Color fit (30% weight)
	factors.ColorFit = scoreColorFit(card, analysis)

	// Factor 2: Mana curve (25% weight)
	factors.ManaCurve = scoreManaCurve(card, analysis)

	// Factor 3: Card quality from ratings (25% weight)
	factors.Quality = e.scoreCardQuality(ctx, card, deck)

	// Factor 4: Synergy (15% weight)
	factors.Synergy = scoreSynergy(card, deck, analysis)

	// Factor 5: Playability (5% weight)
	factors.Playable = scorePlayability(card, deck)

	// Calculate weighted overall score
	score := (factors.ColorFit * 0.30) +
		(factors.ManaCurve * 0.25) +
		(factors.Quality * 0.25) +
		(factors.Synergy * 0.15) +
		(factors.Playable * 0.05)

	// Generate explanation
	reasoning := generateExplanation(card, factors, analysis)

	// Determine primary source
	source := determinePrimarySource(factors)

	// Confidence based on how many factors agree
	confidence := calculateConfidence(factors)

	return &CardRecommendation{
		Card:       card,
		Score:      score,
		Reasoning:  reasoning,
		Source:     source,
		Confidence: confidence,
		Factors:    factors,
	}
}

// DeckAnalysis contains analyzed deck composition data.
type DeckAnalysis struct {
	Colors        map[string]int // Color distribution
	ManaCurve     map[int]int    // CMC distribution
	CardTypes     map[string]int // Card type distribution
	TotalCards    int
	TotalNonLands int
	AverageCMC    float64
	ColorIdentity []string       // Deck's color identity
	PrimaryColors []string       // Most common colors
	Keywords      map[string]int // Keyword counts (Flying, Trample, etc.)
	CreatureTypes map[string]int // Creature type distribution
}

// analyzeDeck performs comprehensive deck analysis for recommendations.
func analyzeDeck(deck *DeckContext) *DeckAnalysis {
	analysis := &DeckAnalysis{
		Colors:        make(map[string]int),
		ManaCurve:     make(map[int]int),
		CardTypes:     make(map[string]int),
		Keywords:      make(map[string]int),
		CreatureTypes: make(map[string]int),
	}

	totalCMC := 0.0
	nonLandCount := 0

	for _, deckCard := range deck.Cards {
		if deckCard.Board != "main" {
			continue // Only analyze mainboard
		}

		card, ok := deck.CardMetadata[deckCard.CardID]
		if !ok {
			continue
		}

		analysis.TotalCards += deckCard.Quantity

		// Extract and track card types from TypeLine
		cardTypes := extractTypesFromTypeLine(card.TypeLine)
		for _, cardType := range cardTypes {
			analysis.CardTypes[cardType] += deckCard.Quantity
		}

		// Skip lands for CMC/color analysis
		isLand := containsTypeInTypeLine(card.TypeLine, "Land")

		if !isLand {
			// Track colors
			for _, color := range card.Colors {
				analysis.Colors[color] += deckCard.Quantity
			}

			// Track mana curve
			cmc := int(card.CMC)
			analysis.ManaCurve[cmc] += deckCard.Quantity

			totalCMC += card.CMC * float64(deckCard.Quantity)
			nonLandCount += deckCard.Quantity
		}

		// Extract keywords using sophisticated extraction
		if card.OracleText != nil {
			extractKeywords(*card.OracleText, analysis.Keywords, deckCard.Quantity)
		}

		// Extract creature types
		if containsTypeInTypeLine(card.TypeLine, "Creature") {
			extractCreatureTypes(card.TypeLine, analysis.CreatureTypes, deckCard.Quantity)
		}
	}

	analysis.TotalNonLands = nonLandCount
	if nonLandCount > 0 {
		analysis.AverageCMC = totalCMC / float64(nonLandCount)
	}

	// Determine color identity and primary colors
	analysis.ColorIdentity = getColorIdentity(analysis.Colors)
	analysis.PrimaryColors = getPrimaryColors(analysis.Colors, 2) // Top 2 colors

	return analysis
}

// scoreColorFit calculates how well a card's colors match the deck's color identity.
func scoreColorFit(card *cards.Card, analysis *DeckAnalysis) float64 {
	if len(card.Colors) == 0 {
		// Colorless cards are always playable
		return 1.0
	}

	// Check if all card colors are in deck's color identity
	matchingColors := 0
	for _, cardColor := range card.Colors {
		for _, deckColor := range analysis.ColorIdentity {
			if cardColor == deckColor {
				matchingColors++
				break
			}
		}
	}

	if matchingColors == 0 {
		// Card colors don't match deck at all
		return 0.0
	}

	if matchingColors == len(card.Colors) {
		// All card colors are in deck
		// Bonus if colors match primary colors
		primaryMatch := 0
		for _, cardColor := range card.Colors {
			for _, primary := range analysis.PrimaryColors {
				if cardColor == primary {
					primaryMatch++
					break
				}
			}
		}
		if primaryMatch == len(card.Colors) {
			return 1.0 // Perfect fit with primary colors
		}
		return 0.85 // Good fit, but includes splash colors
	}

	// Partial color match - card has some colors outside deck identity
	return float64(matchingColors) / float64(len(card.Colors)) * 0.5
}

// scoreManaCurve calculates how well a card fits the deck's mana curve needs.
func scoreManaCurve(card *cards.Card, analysis *DeckAnalysis) float64 {
	// Skip lands
	if containsTypeInTypeLine(card.TypeLine, "Land") {
		return 0.5 // Neutral score for lands
	}

	cmc := int(card.CMC)

	// Get current count at this CMC
	currentCount := analysis.ManaCurve[cmc]

	// Ideal distribution for Limited (rough approximation)
	// CMC 1: 1-2, CMC 2: 4-6, CMC 3: 4-5, CMC 4: 3-4, CMC 5: 2-3, CMC 6+: 1-2
	idealCounts := map[int]int{
		1: 2,
		2: 5,
		3: 5,
		4: 4,
		5: 3,
		6: 2,
	}

	ideal := idealCounts[cmc]
	if cmc > 6 {
		ideal = 2
	}

	// If we're under the ideal, this card is more valuable
	if currentCount < ideal {
		gap := ideal - currentCount
		score := 0.7 + (float64(gap) * 0.1) // Higher score for bigger gaps
		if score > 1.0 {
			score = 1.0 // Cap at 1.0
		}
		return score
	}

	// If we're at ideal, still decent
	if currentCount == ideal {
		return 0.6
	}

	// If we're over ideal, less valuable
	excess := currentCount - ideal
	score := 0.5 - (float64(excess) * 0.1)
	if score < 0.1 {
		score = 0.1 // Minimum score
	}
	return score
}

// scoreCardQuality calculates intrinsic card quality based on ratings.
// Uses 17Lands data as primary source and CFB ratings as secondary/supplementary.
func (e *RuleBasedEngine) scoreCardQuality(ctx context.Context, card *cards.Card, deck *DeckContext) float64 {
	// Track which rating sources we have
	seventeenLandsScore := -1.0
	cfbScore := -1.0

	// Try to get 17Lands score
	if e.ratingsRepo != nil && deck.SetCode != "" && deck.DraftFormat != "" {
		arenaIDStr := fmt.Sprintf("%d", card.ArenaID)
		rating, err := e.ratingsRepo.GetCardRatingByArenaID(ctx, deck.SetCode, deck.DraftFormat, arenaIDStr)
		if err == nil && rating != nil {
			seventeenLandsScore = e.calculate17LandsScore(rating)
		}
	}

	// Try to get CFB score
	if e.cfbRatingsRepo != nil {
		cfbRating, err := e.cfbRatingsRepo.GetRatingByArenaID(ctx, card.ArenaID)
		if err == nil && cfbRating != nil {
			cfbScore = cfbRating.LimitedScore
		}
	}

	// Blend scores based on availability
	// If both available: 70% 17Lands + 30% CFB (17Lands is data-driven, CFB is expert opinion)
	// If only 17Lands: 100% 17Lands
	// If only CFB: 100% CFB
	// If neither: fallback to rarity

	if seventeenLandsScore >= 0 && cfbScore >= 0 {
		return (seventeenLandsScore * 0.70) + (cfbScore * 0.30)
	} else if seventeenLandsScore >= 0 {
		return seventeenLandsScore
	} else if cfbScore >= 0 {
		return cfbScore
	}

	return e.fallbackQualityScore(card)
}

// calculate17LandsScore calculates quality score from 17Lands metrics.
func (e *RuleBasedEngine) calculate17LandsScore(rating *seventeenlands.CardRating) float64 {
	// Weight: 50% GIHWR, 30% OHWR, 10% ATA, 10% ALSA

	// Normalize GIHWR and OHWR (they're percentages, typically 45-60% range)
	// Map 45% → 0.0, 55% → 0.5, 65%+ → 1.0
	gihScore := (rating.GIHWR - 0.45) / 0.20 // 0.45-0.65 → 0.0-1.0
	if gihScore < 0 {
		gihScore = 0
	} else if gihScore > 1 {
		gihScore = 1
	}

	ohScore := (rating.OHWR - 0.45) / 0.20
	if ohScore < 0 {
		ohScore = 0
	} else if ohScore > 1 {
		ohScore = 1
	}

	// ATA (Average Taken At): Lower is better. Typical range 1-14
	// Map 1 → 1.0, 5 → 0.7, 10 → 0.3, 14 → 0.0
	ataScore := 1.0 - ((rating.ATA - 1.0) / 13.0)
	if ataScore < 0 {
		ataScore = 0
	} else if ataScore > 1 {
		ataScore = 1
	}

	// ALSA (Average Last Seen At): Higher means it wheels more (less valuable)
	// Map 1 → 1.0, 5 → 0.7, 10 → 0.3, 14 → 0.0
	alsaScore := 1.0 - ((rating.ALSA - 1.0) / 13.0)
	if alsaScore < 0 {
		alsaScore = 0
	} else if alsaScore > 1 {
		alsaScore = 1
	}

	// Weighted combination
	return (gihScore * 0.50) + (ohScore * 0.30) + (ataScore * 0.10) + (alsaScore * 0.10)
}

// fallbackQualityScore provides quality score based on rarity when ratings unavailable.
func (e *RuleBasedEngine) fallbackQualityScore(card *cards.Card) float64 {
	rarityScores := map[string]float64{
		"mythic":   0.85,
		"rare":     0.75,
		"uncommon": 0.60,
		"common":   0.50,
	}

	if score, ok := rarityScores[strings.ToLower(card.Rarity)]; ok {
		return score
	}

	return 0.5 // Default neutral score
}

// scoreSynergy calculates synergy with existing deck cards.
func scoreSynergy(card *cards.Card, deck *DeckContext, analysis *DeckAnalysis) float64 {
	synergy := 0.0
	synergyCount := 0

	// Keyword synergy
	if card.OracleText != nil {
		cardKeywords := extractKeywordsFromText(*card.OracleText)

		// Check if deck has similar keywords
		for keyword := range cardKeywords {
			if count, ok := analysis.Keywords[keyword]; ok && count > 0 {
				synergy += 0.2
				synergyCount++
			}
		}
	}

	// Creature type synergy (tribal)
	if containsTypeInTypeLine(card.TypeLine, "Creature") {
		cardTypes := extractCreatureTypesFromLine(card.TypeLine)
		for creatureType := range cardTypes {
			if count, ok := analysis.CreatureTypes[creatureType]; ok && count >= 3 {
				// Tribal synergy scales with how many of this type we have
				// 3-4 creatures: 0.4 bonus
				// 5-7 creatures: 0.6 bonus
				// 8+ creatures: 0.8 bonus (strong tribal theme)
				var tribalBonus float64
				if count >= 8 {
					tribalBonus = 0.8
				} else if count >= 5 {
					tribalBonus = 0.6
				} else {
					tribalBonus = 0.4
				}
				synergy += tribalBonus
				synergyCount++
			}
		}
	}

	// If we found synergies, average them
	if synergyCount > 0 {
		synergy = synergy / float64(synergyCount)
		if synergy > 1.0 {
			synergy = 1.0
		}
		return synergy
	}

	// No specific synergy found
	return 0.5 // Neutral score
}

// scorePlayability calculates format-specific playability.
func scorePlayability(card *cards.Card, deck *DeckContext) float64 {
	// For Phase 1A, basic playability check

	// Check if card is legal in format
	// For Limited, all cards in draft pool are legal
	if deck.Format == "Limited" {
		// In draft, playability is high for cards in the pool
		if deck.DraftCardIDs != nil {
			for _, id := range deck.DraftCardIDs {
				if id == card.ArenaID {
					return 0.9
				}
			}
			return 0.1 // Not in draft pool
		}
		return 0.8 // Default for Limited
	}

	// For Constructed formats, assume cards are playable
	// Phase 1B will add format legality checking
	return 0.8
}

// generateExplanation creates a human-readable explanation for a recommendation.
func generateExplanation(card *cards.Card, factors *ScoreFactors, analysis *DeckAnalysis) string {
	reasons := make([]string, 0)

	// Color fit reasoning
	if factors.ColorFit >= 0.85 {
		reasons = append(reasons, "matches your deck's colors perfectly")
	} else if factors.ColorFit >= 0.7 {
		reasons = append(reasons, "fits your color identity")
	} else if factors.ColorFit < 0.3 {
		reasons = append(reasons, "color requirements may be difficult")
	}

	// Mana curve reasoning
	if factors.ManaCurve >= 0.7 {
		reasons = append(reasons, fmt.Sprintf("fills a gap in your mana curve at %d CMC", int(card.CMC)))
	} else if factors.ManaCurve <= 0.3 {
		reasons = append(reasons, fmt.Sprintf("your deck already has many %d-drops", int(card.CMC)))
	}

	// Quality reasoning
	if factors.Quality >= 0.8 {
		reasons = append(reasons, "is a high-quality card")
	} else if factors.Quality >= 0.7 {
		reasons = append(reasons, "has strong ratings")
	}

	// Synergy reasoning
	if factors.Synergy >= 0.8 {
		reasons = append(reasons, "has excellent synergy with your deck's strategy")
	} else if factors.Synergy >= 0.7 {
		reasons = append(reasons, "has strong synergy with your existing cards")
	} else if factors.Synergy >= 0.6 {
		reasons = append(reasons, "synergizes well with your deck")
	}

	// Construct final explanation
	if len(reasons) == 0 {
		return "This card could work in your deck."
	}

	explanation := "This card "
	for i, reason := range reasons {
		if i == 0 {
			explanation += reason
		} else if i == len(reasons)-1 {
			explanation += ", and " + reason
		} else {
			explanation += ", " + reason
		}
	}
	explanation += "."

	return explanation
}

// determinePrimarySource identifies the primary factor driving the recommendation.
func determinePrimarySource(factors *ScoreFactors) string {
	maxScore := 0.0
	source := "quality"

	if factors.ColorFit > maxScore {
		maxScore = factors.ColorFit
		source = "color-fit"
	}
	if factors.ManaCurve > maxScore {
		maxScore = factors.ManaCurve
		source = "mana-curve"
	}
	if factors.Quality > maxScore {
		maxScore = factors.Quality
		source = "quality"
	}
	if factors.Synergy > maxScore {
		maxScore = factors.Synergy
		source = "synergy"
	}
	if factors.Playable > maxScore {
		source = "playability"
	}

	return source
}

// calculateConfidence calculates confidence based on factor agreement.
func calculateConfidence(factors *ScoreFactors) float64 {
	// Count how many factors are "positive" (> 0.6)
	positiveCount := 0
	totalFactors := 5

	if factors.ColorFit > 0.6 {
		positiveCount++
	}
	if factors.ManaCurve > 0.6 {
		positiveCount++
	}
	if factors.Quality > 0.6 {
		positiveCount++
	}
	if factors.Synergy > 0.6 {
		positiveCount++
	}
	if factors.Playable > 0.6 {
		positiveCount++
	}

	// Calculate confidence as percentage of agreeing factors
	confidence := float64(positiveCount) / float64(totalFactors)

	// Boost confidence if multiple high scores
	highCount := 0
	if factors.ColorFit > 0.8 {
		highCount++
	}
	if factors.ManaCurve > 0.8 {
		highCount++
	}
	if factors.Quality > 0.8 {
		highCount++
	}
	if factors.Synergy > 0.8 {
		highCount++
	}
	if factors.Playable > 0.8 {
		highCount++
	}

	if highCount >= 2 {
		confidence += 0.1
	}
	if highCount >= 3 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// isCardInDeck checks if a card is already in the deck.
func isCardInDeck(cardID int, cards []*models.DeckCard) bool {
	for _, deckCard := range cards {
		if deckCard.CardID == cardID {
			return true
		}
	}
	return false
}

// matchesFilters checks if a card matches the provided filters.
func matchesFilters(card *cards.Card, filters *Filters) bool {
	// Color filter
	if len(filters.Colors) > 0 {
		hasMatchingColor := false
		for _, filterColor := range filters.Colors {
			for _, cardColor := range card.Colors {
				if cardColor == filterColor {
					hasMatchingColor = true
					break
				}
			}
			if hasMatchingColor {
				break
			}
		}
		if !hasMatchingColor && len(card.Colors) > 0 {
			return false // Card doesn't match color filter
		}
	}

	// Card type filter
	if len(filters.CardTypes) > 0 {
		hasMatchingType := false
		for _, filterType := range filters.CardTypes {
			if containsTypeInTypeLine(card.TypeLine, filterType) {
				hasMatchingType = true
				break
			}
		}
		if !hasMatchingType {
			return false
		}
	}

	// CMC range filter
	if filters.CMCRange != nil {
		cmc := int(card.CMC)
		if cmc < filters.CMCRange.Min || cmc > filters.CMCRange.Max {
			return false
		}
	}

	// Land filter
	if !filters.IncludeLands && containsTypeInTypeLine(card.TypeLine, "Land") {
		return false
	}

	return true
}

// sortRecommendations sorts recommendations by score (descending).
func sortRecommendations(recommendations []*CardRecommendation) {
	// Simple bubble sort for now (fine for small lists)
	// Can optimize later if needed
	n := len(recommendations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if recommendations[j].Score < recommendations[j+1].Score {
				recommendations[j], recommendations[j+1] = recommendations[j+1], recommendations[j]
			}
		}
	}
}

// getColorIdentity determines the deck's color identity from color distribution.
func getColorIdentity(colors map[string]int) []string {
	identity := make([]string, 0)

	// A color is part of identity if it has any cards
	for color, count := range colors {
		if count > 0 {
			identity = append(identity, color)
		}
	}

	return identity
}

// getPrimaryColors identifies the most common colors in the deck.
func getPrimaryColors(colors map[string]int, limit int) []string {
	type colorCount struct {
		color string
		count int
	}

	// Convert map to slice for sorting
	counts := make([]colorCount, 0, len(colors))
	for color, count := range colors {
		if count > 0 {
			counts = append(counts, colorCount{color, count})
		}
	}

	// Sort by count (descending)
	for i := 0; i < len(counts)-1; i++ {
		for j := 0; j < len(counts)-i-1; j++ {
			if counts[j].count < counts[j+1].count {
				counts[j], counts[j+1] = counts[j+1], counts[j]
			}
		}
	}

	// Take top N colors
	primaryColors := make([]string, 0, limit)
	for i := 0; i < limit && i < len(counts); i++ {
		primaryColors = append(primaryColors, counts[i].color)
	}

	return primaryColors
}

// extractKeywords extracts keywords from card text and adds to the map.
func extractKeywords(text string, keywords map[string]int, quantity int) {
	keywordList := extractKeywordsFromText(text)
	for keyword := range keywordList {
		keywords[keyword] += quantity
	}
}

// KeywordCategory represents the type of keyword extracted.
type KeywordCategory string

const (
	CategoryCombat     KeywordCategory = "combat"
	CategoryAbility    KeywordCategory = "ability"
	CategoryMechanic   KeywordCategory = "mechanic"
	CategoryTheme      KeywordCategory = "theme"
	CategoryTrigger    KeywordCategory = "trigger"
	CategoryActivated  KeywordCategory = "activated"
	CategoryProtection KeywordCategory = "protection"
)

// KeywordInfo contains information about an extracted keyword.
type KeywordInfo struct {
	Keyword  string
	Category KeywordCategory
	Weight   float64 // How important this keyword is for synergy (0.0-1.0)
}

// keywordDictionary maps keywords/patterns to their category and weight.
var keywordDictionary = map[string]KeywordInfo{
	// Combat keywords (high synergy weight)
	"flying":        {Keyword: "flying", Category: CategoryCombat, Weight: 0.8},
	"first strike":  {Keyword: "first strike", Category: CategoryCombat, Weight: 0.7},
	"double strike": {Keyword: "double strike", Category: CategoryCombat, Weight: 0.9},
	"deathtouch":    {Keyword: "deathtouch", Category: CategoryCombat, Weight: 0.8},
	"haste":         {Keyword: "haste", Category: CategoryCombat, Weight: 0.6},
	"lifelink":      {Keyword: "lifelink", Category: CategoryCombat, Weight: 0.7},
	"menace":        {Keyword: "menace", Category: CategoryCombat, Weight: 0.7},
	"reach":         {Keyword: "reach", Category: CategoryCombat, Weight: 0.5},
	"trample":       {Keyword: "trample", Category: CategoryCombat, Weight: 0.7},
	"vigilance":     {Keyword: "vigilance", Category: CategoryCombat, Weight: 0.6},

	// Protection/evasion keywords
	"hexproof":       {Keyword: "hexproof", Category: CategoryProtection, Weight: 0.8},
	"indestructible": {Keyword: "indestructible", Category: CategoryProtection, Weight: 0.9},
	"ward":           {Keyword: "ward", Category: CategoryProtection, Weight: 0.7},
	"shroud":         {Keyword: "shroud", Category: CategoryProtection, Weight: 0.7},
	"protection":     {Keyword: "protection", Category: CategoryProtection, Weight: 0.7},

	// Static ability keywords
	"flash":    {Keyword: "flash", Category: CategoryAbility, Weight: 0.6},
	"defender": {Keyword: "defender", Category: CategoryAbility, Weight: 0.3},
	"prowess":  {Keyword: "prowess", Category: CategoryAbility, Weight: 0.8},
	"convoke":  {Keyword: "convoke", Category: CategoryAbility, Weight: 0.7},

	// Set mechanics
	"flashback": {Keyword: "flashback", Category: CategoryMechanic, Weight: 0.8},
	"kicker":    {Keyword: "kicker", Category: CategoryMechanic, Weight: 0.6},
	"adventure": {Keyword: "adventure", Category: CategoryMechanic, Weight: 0.7},
	"transform": {Keyword: "transform", Category: CategoryMechanic, Weight: 0.6},
	"disturb":   {Keyword: "disturb", Category: CategoryMechanic, Weight: 0.7},
	"exploit":   {Keyword: "exploit", Category: CategoryMechanic, Weight: 0.8},
	"escape":    {Keyword: "escape", Category: CategoryMechanic, Weight: 0.8},
	"madness":   {Keyword: "madness", Category: CategoryMechanic, Weight: 0.7},
	"cycling":   {Keyword: "cycling", Category: CategoryMechanic, Weight: 0.6},
	"cascade":   {Keyword: "cascade", Category: CategoryMechanic, Weight: 0.9},
	"mutate":    {Keyword: "mutate", Category: CategoryMechanic, Weight: 0.8},
	"foretell":  {Keyword: "foretell", Category: CategoryMechanic, Weight: 0.7},
	"learn":     {Keyword: "learn", Category: CategoryMechanic, Weight: 0.6},
	"ninjutsu":  {Keyword: "ninjutsu", Category: CategoryMechanic, Weight: 0.8},
	"channel":   {Keyword: "channel", Category: CategoryMechanic, Weight: 0.6},
	"bargain":   {Keyword: "bargain", Category: CategoryMechanic, Weight: 0.7},
	"craft":     {Keyword: "craft", Category: CategoryMechanic, Weight: 0.7},
	"descend":   {Keyword: "descend", Category: CategoryMechanic, Weight: 0.7},
	"threshold": {Keyword: "threshold", Category: CategoryMechanic, Weight: 0.7},
	"delirium":  {Keyword: "delirium", Category: CategoryMechanic, Weight: 0.7},
	"affinity":  {Keyword: "affinity", Category: CategoryMechanic, Weight: 0.8},
	"modular":   {Keyword: "modular", Category: CategoryMechanic, Weight: 0.7},
	"equip":     {Keyword: "equip", Category: CategoryMechanic, Weight: 0.6},
	"crew":      {Keyword: "crew", Category: CategoryMechanic, Weight: 0.7},
	"offspring": {Keyword: "offspring", Category: CategoryMechanic, Weight: 0.7},
	"valiant":   {Keyword: "valiant", Category: CategoryMechanic, Weight: 0.7},
}

// themePatterns are regex-like patterns for identifying card themes.
var themePatterns = []struct {
	pattern  string
	keyword  string
	category KeywordCategory
	weight   float64
}{
	// Token generation
	{"create a token", "tokens", CategoryTheme, 0.9},
	{"create a.*token", "tokens", CategoryTheme, 0.9},
	{"creates a token", "tokens", CategoryTheme, 0.9},
	{"create two", "tokens", CategoryTheme, 0.9},
	{"create three", "tokens", CategoryTheme, 0.9},

	// Counter themes
	{"+1/+1 counter", "+1/+1 counters", CategoryTheme, 0.9},
	{"put a counter", "counters", CategoryTheme, 0.7},
	{"put counters", "counters", CategoryTheme, 0.7},
	{"-1/-1 counter", "-1/-1 counters", CategoryTheme, 0.8},
	{"loyalty counter", "planeswalkers", CategoryTheme, 0.7},
	{"charge counter", "artifacts", CategoryTheme, 0.6},

	// Graveyard themes
	{"from your graveyard", "graveyard", CategoryTheme, 0.9},
	{"from a graveyard", "graveyard", CategoryTheme, 0.9},
	{"in your graveyard", "graveyard", CategoryTheme, 0.8},
	{"return.*from.*graveyard", "graveyard", CategoryTheme, 0.9},
	{"mill", "mill", CategoryTheme, 0.8},
	{"self-mill", "mill", CategoryTheme, 0.9},

	// Sacrifice themes (aristocrats)
	{"sacrifice a", "sacrifice", CategoryTheme, 0.9},
	{"sacrifice another", "sacrifice", CategoryTheme, 0.9},
	{"sacrificed", "sacrifice", CategoryTheme, 0.8},
	{"when.*dies", "death triggers", CategoryTheme, 0.8},
	{"whenever.*dies", "death triggers", CategoryTheme, 0.8},
	{"whenever another creature you control dies", "death payoff", CategoryTheme, 0.95},
	{"whenever a creature you control dies", "death payoff", CategoryTheme, 0.9},
	{"whenever a nontoken creature", "death triggers", CategoryTheme, 0.85},
	{"sacrifice a token", "token sacrifice", CategoryTheme, 0.9},
	{"sacrifices a token", "token sacrifice", CategoryTheme, 0.9},
	{"when you sacrifice", "sacrifice payoff", CategoryTheme, 0.9},
	{"whenever you sacrifice", "sacrifice payoff", CategoryTheme, 0.9},
	{"whenever you sacrifice a", "sacrifice payoff", CategoryTheme, 0.9},
	{"as an additional cost", "sacrifice", CategoryTheme, 0.7},
	{"blood token", "blood tokens", CategoryTheme, 0.8},
	{"food token", "food tokens", CategoryTheme, 0.8},
	{"treasure token", "treasure tokens", CategoryTheme, 0.8},
	{"clue token", "clue tokens", CategoryTheme, 0.8},
	{"map token", "map tokens", CategoryTheme, 0.8},

	// Draw/card advantage
	{"draw a card", "card draw", CategoryTheme, 0.7},
	{"draw cards", "card draw", CategoryTheme, 0.8},
	{"draw two", "card draw", CategoryTheme, 0.8},
	{"draw three", "card draw", CategoryTheme, 0.9},
	{"scry", "scry", CategoryTheme, 0.6},
	{"surveil", "surveil", CategoryTheme, 0.7},

	// Second card drawn triggers (e.g., Jolrael, Teferi's Ageless Insight)
	{"whenever you draw your second card", "second draw", CategoryTheme, 0.9},
	{"draw your second card", "second draw", CategoryTheme, 0.9},
	{"draw an additional card", "card advantage", CategoryTheme, 0.8},
	{"you may draw an additional card", "card advantage", CategoryTheme, 0.8},
	{"draw two additional cards", "card advantage", CategoryTheme, 0.9},

	// Life themes
	{"gain life", "lifegain", CategoryTheme, 0.7},
	{"gains life", "lifegain", CategoryTheme, 0.7},
	{"whenever you gain life", "lifegain payoff", CategoryTheme, 0.9},
	{"lose life", "drain", CategoryTheme, 0.7},
	{"pay life", "pay life", CategoryTheme, 0.6},

	// Combat triggers
	{"deals combat damage", "combat damage", CategoryTheme, 0.8},
	{"whenever.*attacks", "attack triggers", CategoryTheme, 0.8},
	{"whenever.*blocks", "block triggers", CategoryTheme, 0.7},
	{"can't be blocked", "evasion", CategoryTheme, 0.8},

	// Enters triggers
	{"when.*enters", "enters triggers", CategoryTrigger, 0.8},
	{"when.*enters the battlefield", "ETB", CategoryTrigger, 0.9},
	{"whenever.*enters", "enters triggers", CategoryTrigger, 0.8},

	// Cast triggers
	{"whenever you cast", "cast triggers", CategoryTrigger, 0.8},
	{"when you cast", "cast triggers", CategoryTrigger, 0.8},

	// Upkeep/beginning triggers
	{"at the beginning of", "upkeep triggers", CategoryTrigger, 0.7},
	{"at the beginning of your upkeep", "upkeep", CategoryTrigger, 0.7},
	{"at the beginning of your end step", "end step", CategoryTrigger, 0.7},

	// Activated abilities
	{"{t}:", "tap abilities", CategoryActivated, 0.7},
	{"{t},", "tap abilities", CategoryActivated, 0.7},
	{"tap:", "tap abilities", CategoryActivated, 0.7},

	// Spell themes (prowess/magecraft)
	{"instant or sorcery", "spells matter", CategoryTheme, 0.8},
	{"noncreature spell", "spells matter", CategoryTheme, 0.8},
	{"whenever you cast a noncreature spell", "spells matter", CategoryTheme, 0.9},
	{"whenever you cast an instant or sorcery", "spells matter", CategoryTheme, 0.9},
	{"whenever you cast your first instant or sorcery", "spells matter", CategoryTheme, 0.85},
	{"magecraft", "spells matter", CategoryTheme, 0.9},
	{"copy target instant or sorcery", "spell copy", CategoryTheme, 0.8},
	{"copy that spell", "spell copy", CategoryTheme, 0.8},

	// Artifact/enchantment themes
	{"artifact you control", "artifacts matter", CategoryTheme, 0.8},
	{"enchantment you control", "enchantments matter", CategoryTheme, 0.8},
	{"aura", "auras", CategoryTheme, 0.7},
	{"equipment", "equipment", CategoryTheme, 0.7},
	{"whenever an artifact enters", "artifacts ETB", CategoryTheme, 0.85},
	{"whenever an artifact enters the battlefield", "artifacts ETB", CategoryTheme, 0.9},
	{"whenever an artifact leaves", "artifacts matter", CategoryTheme, 0.8},
	{"whenever an artifact or creature you control", "artifacts matter", CategoryTheme, 0.8},
	{"whenever an enchantment enters", "enchantments ETB", CategoryTheme, 0.85},
	{"whenever an enchantment enters the battlefield", "enchantments ETB", CategoryTheme, 0.9},

	// Power/toughness manipulation
	{"gets +", "pump", CategoryTheme, 0.6},
	{"get +", "pump", CategoryTheme, 0.6},
	{"target creature gets", "pump", CategoryTheme, 0.6},

	// Removal
	{"destroy target", "removal", CategoryTheme, 0.8},
	{"exile target", "removal", CategoryTheme, 0.8},
	{"deals.*damage to", "damage", CategoryTheme, 0.7},
	{"fight", "fight", CategoryTheme, 0.7},
	{"bite", "fight", CategoryTheme, 0.7},

	// Discard themes
	{"discard a card", "discard", CategoryTheme, 0.7},
	{"discards a card", "discard", CategoryTheme, 0.7},
	{"whenever you discard", "discard payoff", CategoryTheme, 0.85},
	{"whenever a player discards", "discard payoff", CategoryTheme, 0.8},
	{"discard your hand", "discard", CategoryTheme, 0.8},
	{"discarded this turn", "discard payoff", CategoryTheme, 0.8},

	// Blink/flicker
	{"exile.*then return", "blink", CategoryTheme, 0.9},
	{"exile target creature you control, then return", "blink", CategoryTheme, 0.9},
	{"exile it, then return", "blink", CategoryTheme, 0.85},
	{"exile up to one", "blink", CategoryTheme, 0.8},

	// Creature-specific themes
	{"whenever a creature you control attacks", "go wide", CategoryTheme, 0.85},
	{"whenever one or more creatures you control", "go wide", CategoryTheme, 0.8},
	{"creatures you control get", "anthem", CategoryTheme, 0.9},
	{"other creatures you control get", "anthem", CategoryTheme, 0.9},
	{"creatures you control have", "anthem", CategoryTheme, 0.85},

	// Landfall
	{"whenever a land enters the battlefield under your control", "landfall", CategoryTheme, 0.9},
	{"landfall", "landfall", CategoryTheme, 0.9},

	// Mana themes
	{"add one mana", "mana ramp", CategoryTheme, 0.6},
	{"add two mana", "mana ramp", CategoryTheme, 0.7},
	{"search your library for a.*land", "land ramp", CategoryTheme, 0.8},
	{"search your library for a basic land", "land ramp", CategoryTheme, 0.8},

	// Historic (legendary/artifact/saga)
	{"historic", "historic", CategoryTheme, 0.7},
	{"legendary", "legends matter", CategoryTheme, 0.6},
	{"whenever you cast a legendary", "legends matter", CategoryTheme, 0.85},

	// Type-specific triggers
	{"whenever another.*enters", "type ETB", CategoryTheme, 0.75},
	{"for each creature you control", "go wide", CategoryTheme, 0.85},
	{"equal to the number of creatures you control", "go wide", CategoryTheme, 0.85},

	// Protection/control
	{"can't be countered", "uncounterable", CategoryTheme, 0.7},
	{"this spell can't be countered", "uncounterable", CategoryTheme, 0.7},
	{"counter target spell", "control", CategoryTheme, 0.8},
	{"return target", "bounce", CategoryTheme, 0.7},

	// Energy
	{"energy counter", "energy", CategoryTheme, 0.9},
	{"{e}", "energy", CategoryTheme, 0.9},
}

// extractKeywordsFromText extracts keywords from card text using sophisticated pattern matching.
func extractKeywordsFromText(text string) map[string]bool {
	keywords := make(map[string]bool)
	lowerText := strings.ToLower(text)

	// Extract keywords from dictionary
	for key, info := range keywordDictionary {
		if strings.Contains(lowerText, key) {
			keywords[info.Keyword] = true
		}
	}

	// Extract theme patterns
	for _, pattern := range themePatterns {
		if containsPattern(lowerText, pattern.pattern) {
			keywords[pattern.keyword] = true
		}
	}

	return keywords
}

// ExtractKeywordsWithInfo extracts keywords with full information including category and weight.
// This is useful for more sophisticated synergy calculations.
func ExtractKeywordsWithInfo(text string) []KeywordInfo {
	var result []KeywordInfo
	seen := make(map[string]bool)
	lowerText := strings.ToLower(text)

	// Extract keywords from dictionary
	for key, info := range keywordDictionary {
		if strings.Contains(lowerText, key) && !seen[info.Keyword] {
			result = append(result, info)
			seen[info.Keyword] = true
		}
	}

	// Extract theme patterns
	for _, pattern := range themePatterns {
		if containsPattern(lowerText, pattern.pattern) && !seen[pattern.keyword] {
			result = append(result, KeywordInfo{
				Keyword:  pattern.keyword,
				Category: pattern.category,
				Weight:   pattern.weight,
			})
			seen[pattern.keyword] = true
		}
	}

	return result
}

// GetKeywordWeight returns the synergy weight for a keyword.
// Higher weight means the keyword is more important for synergy calculations.
func GetKeywordWeight(keyword string) float64 {
	lowerKeyword := strings.ToLower(keyword)

	// Check dictionary first
	if info, ok := keywordDictionary[lowerKeyword]; ok {
		return info.Weight
	}

	// Check theme patterns
	for _, pattern := range themePatterns {
		if pattern.keyword == keyword {
			return pattern.weight
		}
	}

	// Default weight for unknown keywords
	return 0.5
}

// GetKeywordCategory returns the category for a keyword.
func GetKeywordCategory(keyword string) KeywordCategory {
	lowerKeyword := strings.ToLower(keyword)

	// Check dictionary first
	if info, ok := keywordDictionary[lowerKeyword]; ok {
		return info.Category
	}

	// Check theme patterns
	for _, pattern := range themePatterns {
		if pattern.keyword == keyword {
			return pattern.category
		}
	}

	// Default category
	return CategoryAbility
}

// containsPattern performs simple pattern matching.
// Supports basic patterns like "create a.*token" where .* matches any characters.
func containsPattern(text, pattern string) bool {
	// Handle simple patterns without wildcards
	if !strings.Contains(pattern, ".*") {
		return strings.Contains(text, pattern)
	}

	// Handle .* wildcard (matches any characters)
	parts := strings.Split(pattern, ".*")
	if len(parts) != 2 {
		// For patterns with multiple wildcards, fall back to simple contains
		return strings.Contains(text, parts[0])
	}

	// Find first part, then check if second part exists after it
	idx := strings.Index(text, parts[0])
	if idx == -1 {
		return false
	}

	// Check if second part exists after the first part
	remaining := text[idx+len(parts[0]):]
	return strings.Contains(remaining, parts[1])
}

// CalculateKeywordSynergy calculates synergy score between two cards based on shared keywords.
// Returns a score from 0.0 to 1.0.
func CalculateKeywordSynergy(card1Keywords, card2Keywords []KeywordInfo) float64 {
	score, _ := CalculateKeywordSynergyDetailed(card1Keywords, card2Keywords)
	return score
}

// CalculateKeywordSynergyDetailed calculates synergy score and returns matched keywords.
// Returns a score from 0.0 to 1.0 and a list of matched keyword names.
func CalculateKeywordSynergyDetailed(card1Keywords, card2Keywords []KeywordInfo) (float64, []string) {
	matchedKeywords := make([]string, 0)
	seenKeywords := make(map[string]bool)

	if len(card1Keywords) == 0 || len(card2Keywords) == 0 {
		return 0.0, matchedKeywords
	}

	// Build lookup for card1 keywords by category
	card1ByCategory := make(map[KeywordCategory][]KeywordInfo)
	for _, kw := range card1Keywords {
		card1ByCategory[kw.Category] = append(card1ByCategory[kw.Category], kw)
	}

	totalSynergy := 0.0
	matchCount := 0

	for _, kw2 := range card2Keywords {
		// Check for exact keyword match
		for _, kw1 := range card1Keywords {
			if kw1.Keyword == kw2.Keyword {
				// Exact match - use average of weights
				totalSynergy += (kw1.Weight + kw2.Weight) / 2
				matchCount++
				// Track matched keyword (only add once)
				if !seenKeywords[kw1.Keyword] {
					matchedKeywords = append(matchedKeywords, kw1.Keyword)
					seenKeywords[kw1.Keyword] = true
				}
			}
		}

		// Check for same category match (weaker synergy)
		if categoryKws, ok := card1ByCategory[kw2.Category]; ok {
			for _, kw1 := range categoryKws {
				if kw1.Keyword != kw2.Keyword {
					// Same category but different keyword - partial synergy
					totalSynergy += (kw1.Weight + kw2.Weight) / 4
					matchCount++
				}
			}
		}
	}

	if matchCount == 0 {
		return 0.0, matchedKeywords
	}

	// Normalize by the number of possible matches
	maxMatches := len(card1Keywords) * len(card2Keywords)
	normalized := totalSynergy / float64(maxMatches)

	// Clamp to 0.0-1.0
	if normalized > 1.0 {
		return 1.0, matchedKeywords
	}
	return normalized, matchedKeywords
}

// extractCreatureTypes extracts creature types from type line and adds to the map.
func extractCreatureTypes(typeLine string, types map[string]int, quantity int) {
	creatureTypes := extractCreatureTypesFromLine(typeLine)
	for creatureType := range creatureTypes {
		types[creatureType] += quantity
	}
}

// extractCreatureTypesFromLine extracts creature types from a type line.
func extractCreatureTypesFromLine(typeLine string) map[string]bool {
	types := make(map[string]bool)

	// Type line format: "Creature — Human Warrior" or "Legendary Creature — Elf Wizard"
	parts := strings.Split(typeLine, "—")
	if len(parts) < 2 {
		parts = strings.Split(typeLine, "-") // Try single dash
	}

	if len(parts) >= 2 {
		// Second part contains creature types
		typesPart := strings.TrimSpace(parts[1])
		individualTypes := strings.Fields(typesPart)

		for _, t := range individualTypes {
			types[t] = true
		}
	}

	return types
}

// containsTypeInTypeLine checks if a card's TypeLine contains a specific type.
func containsTypeInTypeLine(typeLine, targetType string) bool {
	return strings.Contains(strings.ToLower(typeLine), strings.ToLower(targetType))
}

// extractTypesFromTypeLine extracts the main card types from a type line.
// For example: "Legendary Creature — Human Warrior" -> ["Legendary", "Creature"]
func extractTypesFromTypeLine(typeLine string) []string {
	if typeLine == "" {
		return []string{}
	}

	// Split on — to get just the supertypes/types part
	parts := strings.Split(typeLine, "—")
	if len(parts) == 0 {
		parts = strings.Split(typeLine, "-") // Try single dash
	}

	// Take the first part (before —) which contains supertypes and types
	typePart := strings.TrimSpace(parts[0])

	// Common card types
	mainTypes := []string{"Creature", "Artifact", "Enchantment", "Instant", "Sorcery", "Land", "Planeswalker"}
	supertypes := []string{"Legendary", "Basic", "Snow", "World"}

	types := []string{}
	lowerTypePart := strings.ToLower(typePart)

	// Check for main types
	for _, t := range mainTypes {
		if strings.Contains(lowerTypePart, strings.ToLower(t)) {
			types = append(types, t)
		}
	}

	// Check for supertypes
	for _, t := range supertypes {
		if strings.Contains(lowerTypePart, strings.ToLower(t)) {
			types = append(types, t)
		}
	}

	return types
}

// convertSetCardToCardsCard converts a models.SetCard to a cards.Card.
// This allows us to use SetCardRepo data in the recommendation engine.
func convertSetCardToCardsCard(setCard *models.SetCard) *cards.Card {
	if setCard == nil {
		return nil
	}

	// Parse ArenaID from string to int
	arenaID := 0
	_, _ = fmt.Sscanf(setCard.ArenaID, "%d", &arenaID)

	// Build TypeLine from Types array
	typeLine := ""
	if len(setCard.Types) > 0 {
		typeLine = setCard.Types[0]
		for i := 1; i < len(setCard.Types); i++ {
			typeLine += " " + setCard.Types[i]
		}
	}

	card := &cards.Card{
		ArenaID:    arenaID,
		ScryfallID: setCard.ScryfallID,
		Name:       setCard.Name,
		TypeLine:   typeLine,
		SetCode:    setCard.SetCode,
		CMC:        float64(setCard.CMC),
		Colors:     setCard.Colors,
		Rarity:     setCard.Rarity,
	}

	// Convert string fields to *string where needed
	if setCard.ManaCost != "" {
		card.ManaCost = &setCard.ManaCost
	}
	if setCard.Power != "" {
		card.Power = &setCard.Power
	}
	if setCard.Toughness != "" {
		card.Toughness = &setCard.Toughness
	}
	if setCard.Text != "" {
		card.OracleText = &setCard.Text
	}
	if setCard.ImageURL != "" {
		card.ImageURI = &setCard.ImageURL
	}

	return card
}

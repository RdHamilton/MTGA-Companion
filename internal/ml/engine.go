package ml

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/recommendations"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Engine wraps the ML model and provides the RecommendationEngine interface.
// It combines ML-based scoring with the existing rule-based engine as a fallback.
type Engine struct {
	model        *Model
	ruleEngine   recommendations.RecommendationEngine
	cardService  *cards.Service
	setCardRepo  repository.SetCardRepository
	feedbackRepo repository.RecommendationFeedbackRepository
	config       *EngineConfig
	mu           sync.RWMutex
}

// EngineConfig configures the ML engine behavior.
type EngineConfig struct {
	// EnableML enables ML-based recommendations.
	EnableML bool

	// FallbackToRules falls back to rule-based engine when ML fails.
	FallbackToRules bool

	// MLWeight is the weight for ML scores when blending with rules (0.0-1.0).
	MLWeight float64

	// MinConfidence is the minimum confidence to use ML predictions.
	MinConfidence float64

	// MaxRecommendations is the maximum number of recommendations to return.
	MaxRecommendations int
}

// DefaultEngineConfig returns sensible defaults.
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		EnableML:           true,
		FallbackToRules:    true,
		MLWeight:           0.6,
		MinConfidence:      0.3,
		MaxRecommendations: 10,
	}
}

// NewEngine creates a new ML recommendation engine.
func NewEngine(
	model *Model,
	ruleEngine recommendations.RecommendationEngine,
	cardService *cards.Service,
	setCardRepo repository.SetCardRepository,
	feedbackRepo repository.RecommendationFeedbackRepository,
	config *EngineConfig,
) *Engine {
	if config == nil {
		config = DefaultEngineConfig()
	}

	return &Engine{
		model:        model,
		ruleEngine:   ruleEngine,
		cardService:  cardService,
		setCardRepo:  setCardRepo,
		feedbackRepo: feedbackRepo,
		config:       config,
	}
}

// GetRecommendations returns ML-enhanced recommendations for a deck.
func (e *Engine) GetRecommendations(ctx context.Context, deck *recommendations.DeckContext, filters *recommendations.Filters) ([]*recommendations.CardRecommendation, error) {
	e.mu.RLock()
	config := e.config
	e.mu.RUnlock()

	// Check if ML is enabled and model is ready
	mlAvailable := config.EnableML && e.model != nil
	if mlAvailable {
		info := e.model.GetModelInfo()
		mlAvailable = info.IsReady
	}

	// Get rule-based recommendations
	var ruleRecs []*recommendations.CardRecommendation
	var ruleErr error

	if e.ruleEngine != nil {
		ruleRecs, ruleErr = e.ruleEngine.GetRecommendations(ctx, deck, filters)
	}

	if !mlAvailable {
		// ML not available, return rule-based results
		if ruleErr != nil {
			return nil, fmt.Errorf("rule-based recommendations failed: %w", ruleErr)
		}
		return ruleRecs, nil
	}

	// Build ML deck context
	mlDeckCtx := e.buildMLDeckContext(deck)

	// Get candidate cards (from rule-based or filters)
	candidates := e.getCandidates(deck, filters, ruleRecs)
	if len(candidates) == 0 {
		return ruleRecs, nil
	}

	// Get ML scores for candidates
	accountID := e.getAccountID(deck)
	mlScores, err := e.model.ScoreCards(ctx, candidates, mlDeckCtx, accountID)
	if err != nil {
		if config.FallbackToRules {
			return ruleRecs, nil
		}
		return nil, fmt.Errorf("ML scoring failed: %w", err)
	}

	// Blend ML and rule-based scores
	blendedRecs := e.blendRecommendations(mlScores, ruleRecs, config.MLWeight)

	// Sort by blended score
	sort.Slice(blendedRecs, func(i, j int) bool {
		return blendedRecs[i].Score > blendedRecs[j].Score
	})

	// Apply filters and limits
	maxResults := config.MaxRecommendations
	if filters != nil && filters.MaxResults > 0 {
		maxResults = filters.MaxResults
	}
	if len(blendedRecs) > maxResults {
		blendedRecs = blendedRecs[:maxResults]
	}

	return blendedRecs, nil
}

// ExplainRecommendation provides an explanation for why a card is recommended.
func (e *Engine) ExplainRecommendation(ctx context.Context, cardID int, deck *recommendations.DeckContext) (string, error) {
	// Get ML explanation if available
	mlExplanation := ""
	if e.model != nil {
		mlDeckCtx := e.buildMLDeckContext(deck)
		accountID := e.getAccountID(deck)

		scores, err := e.model.ScoreCards(ctx, []int{cardID}, mlDeckCtx, accountID)
		if err == nil && len(scores) > 0 {
			score := scores[0]
			if len(score.Factors) > 0 {
				mlExplanation = fmt.Sprintf("ML suggests this card %s. ", score.Factors[0])
				for i := 1; i < len(score.Factors); i++ {
					mlExplanation += fmt.Sprintf("Also, it %s. ", score.Factors[i])
				}
			}
		}
	}

	// Get rule-based explanation
	ruleExplanation := ""
	if e.ruleEngine != nil {
		var err error
		ruleExplanation, err = e.ruleEngine.ExplainRecommendation(ctx, cardID, deck)
		if err != nil {
			ruleExplanation = ""
		}
	}

	// Combine explanations
	if mlExplanation != "" && ruleExplanation != "" {
		return mlExplanation + ruleExplanation, nil
	}
	if mlExplanation != "" {
		return mlExplanation, nil
	}
	if ruleExplanation != "" {
		return ruleExplanation, nil
	}

	return "This card could be a good addition to your deck.", nil
}

// RecordAcceptance records recommendation acceptance for learning.
func (e *Engine) RecordAcceptance(ctx context.Context, deckID string, cardID int, accepted bool) error {
	// Record in feedback repository
	if e.feedbackRepo != nil {
		action := "rejected"
		if accepted {
			action = "accepted"
		}

		feedback := &models.RecommendationFeedback{
			RecommendationType: "deck_card",
			RecommendedCardID:  &cardID,
			Action:             action,
			ContextData:        fmt.Sprintf(`{"deck_id": "%s"}`, deckID),
		}

		if err := e.feedbackRepo.Create(ctx, feedback); err != nil {
			return fmt.Errorf("failed to record feedback: %w", err)
		}

		// Update ML model incrementally
		if e.model != nil {
			if err := e.model.UpdateFromFeedback(ctx, feedback); err != nil {
				// Log but don't fail - incremental learning is optional
				_ = err
			}
		}
	}

	// Also record in rule engine if it supports it
	if e.ruleEngine != nil {
		_ = e.ruleEngine.RecordAcceptance(ctx, deckID, cardID, accepted)
	}

	return nil
}

// buildMLDeckContext converts recommendations.DeckContext to ml.DeckContext.
func (e *Engine) buildMLDeckContext(deck *recommendations.DeckContext) *DeckContext {
	mlCtx := &DeckContext{
		CMCDistribution:  make(map[int]int),
		TypeDistribution: make(map[string]int),
		Keywords:         make(map[string]int),
		CreatureTypes:    make(map[string]int),
		Format:           deck.Format,
		SetCode:          deck.SetCode,
	}

	if deck.Deck != nil {
		mlCtx.DeckID = deck.Deck.ID
	}

	// Build card list and analyze
	cardIDs := make([]int, 0, len(deck.Cards))
	for _, deckCard := range deck.Cards {
		if deckCard.Board != "main" {
			continue
		}
		cardIDs = append(cardIDs, deckCard.CardID)

		// Get card metadata for analysis
		if cardMeta, ok := deck.CardMetadata[deckCard.CardID]; ok {
			// Track CMC distribution
			cmc := int(cardMeta.CMC)
			mlCtx.CMCDistribution[cmc] += deckCard.Quantity

			// Track colors
			for _, color := range cardMeta.Colors {
				found := false
				for _, c := range mlCtx.ColorIdentity {
					if c == color {
						found = true
						break
					}
				}
				if !found {
					mlCtx.ColorIdentity = append(mlCtx.ColorIdentity, color)
				}
			}

			// Track types
			for _, cardType := range getCardTypes(cardMeta.TypeLine) {
				mlCtx.TypeDistribution[cardType] += deckCard.Quantity
			}

			// Track keywords
			if cardMeta.OracleText != nil {
				keywords := extractKeywordsFromText(*cardMeta.OracleText)
				for kw := range keywords {
					mlCtx.Keywords[kw] += deckCard.Quantity
				}
			}

			// Track creature types
			if isCreature(cardMeta.TypeLine) {
				for ct := range extractCreatureTypes(cardMeta.TypeLine) {
					mlCtx.CreatureTypes[ct] += deckCard.Quantity
				}
			}
		}
	}

	mlCtx.Cards = cardIDs

	return mlCtx
}

// getCandidates extracts candidate card IDs from various sources.
func (e *Engine) getCandidates(deck *recommendations.DeckContext, filters *recommendations.Filters, ruleRecs []*recommendations.CardRecommendation) []int {
	candidateSet := make(map[int]bool)

	// Add cards from rule-based recommendations
	for _, rec := range ruleRecs {
		if rec.Card != nil {
			candidateSet[rec.Card.ArenaID] = true
		}
	}

	// Add cards from draft pool if available
	if filters != nil && filters.OnlyDraftPool && len(filters.DraftPool) > 0 {
		for _, cardID := range filters.DraftPool {
			candidateSet[cardID] = true
		}
	}

	// Convert to slice
	candidates := make([]int, 0, len(candidateSet))
	for cardID := range candidateSet {
		candidates = append(candidates, cardID)
	}

	return candidates
}

// getAccountID extracts account ID from deck context.
func (e *Engine) getAccountID(deck *recommendations.DeckContext) int {
	if deck.Deck != nil {
		return deck.Deck.AccountID
	}
	return 0
}

// blendRecommendations combines ML scores with rule-based recommendations.
func (e *Engine) blendRecommendations(mlScores []*CardScore, ruleRecs []*recommendations.CardRecommendation, mlWeight float64) []*recommendations.CardRecommendation {
	// Build lookup for ML scores
	mlScoreMap := make(map[int]*CardScore)
	for _, score := range mlScores {
		mlScoreMap[score.CardID] = score
	}

	// Build lookup for rule recommendations
	ruleRecMap := make(map[int]*recommendations.CardRecommendation)
	for _, rec := range ruleRecs {
		if rec.Card != nil {
			ruleRecMap[rec.Card.ArenaID] = rec
		}
	}

	// Blend all cards
	result := make([]*recommendations.CardRecommendation, 0)
	processed := make(map[int]bool)

	// Process cards with both ML and rule scores
	for cardID, mlScore := range mlScoreMap {
		processed[cardID] = true

		rec := &recommendations.CardRecommendation{
			Confidence: mlScore.Confidence,
		}

		// Get rule-based recommendation if exists
		if ruleRec, exists := ruleRecMap[cardID]; exists {
			rec.Card = ruleRec.Card
			rec.Factors = ruleRec.Factors
			rec.Reasoning = ruleRec.Reasoning

			// Blend scores
			ruleWeight := 1.0 - mlWeight
			rec.Score = (mlScore.Score * mlWeight) + (ruleRec.Score * ruleWeight)

			// Blend confidence
			rec.Confidence = (mlScore.Confidence * mlWeight) + (ruleRec.Confidence * ruleWeight)

			// Update source based on dominant score
			if mlScore.Score > ruleRec.Score {
				rec.Source = "ml-enhanced"
			} else {
				rec.Source = ruleRec.Source
			}
		} else {
			// Only ML score available - we need to get the card info
			rec.Score = mlScore.Score
			rec.Source = "ml"
			rec.Reasoning = buildMLReasoning(mlScore)

			// Try to get card info
			if e.cardService != nil {
				if card, err := e.cardService.GetCard(cardID); err == nil {
					rec.Card = card
				}
			}
		}

		if rec.Card != nil {
			result = append(result, rec)
		}
	}

	// Add rule-based recommendations not in ML scores
	for cardID, ruleRec := range ruleRecMap {
		if !processed[cardID] {
			result = append(result, ruleRec)
		}
	}

	return result
}

// buildMLReasoning creates a reasoning string from ML score factors.
func buildMLReasoning(score *CardScore) string {
	if len(score.Factors) == 0 {
		return "ML model recommends this card based on learned patterns."
	}

	reasoning := "This card is recommended because it "
	for i, factor := range score.Factors {
		if i == 0 {
			reasoning += factor
		} else if i == len(score.Factors)-1 {
			reasoning += ", and " + factor
		} else {
			reasoning += ", " + factor
		}
	}
	reasoning += "."

	return reasoning
}

// SetConfig updates the engine configuration.
func (e *Engine) SetConfig(config *EngineConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = config
}

// GetConfig returns the current configuration.
func (e *Engine) GetConfig() *EngineConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// GetModelInfo returns information about the underlying ML model.
func (e *Engine) GetModelInfo() *ModelInfo {
	if e.model == nil {
		return &ModelInfo{
			Version: "none",
			IsReady: false,
		}
	}
	return e.model.GetModelInfo()
}

// TrainModel triggers model training.
func (e *Engine) TrainModel(ctx context.Context) error {
	if e.model == nil {
		return fmt.Errorf("ML model not initialized")
	}
	return e.model.Train(ctx)
}

// Helper functions

func getCardTypes(typeLine string) []string {
	mainTypes := []string{"Creature", "Instant", "Sorcery", "Enchantment", "Artifact", "Land", "Planeswalker"}
	found := make([]string, 0)

	for _, t := range mainTypes {
		if containsIgnoreCase(typeLine, t) {
			found = append(found, t)
		}
	}

	return found
}

func isCreature(typeLine string) bool {
	return containsIgnoreCase(typeLine, "Creature")
}

func containsIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldASCII(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func extractKeywordsFromText(text string) map[string]bool {
	keywords := make(map[string]bool)
	commonKeywords := []string{
		"Flying", "First strike", "Double strike", "Deathtouch", "Haste",
		"Hexproof", "Indestructible", "Lifelink", "Menace", "Reach",
		"Trample", "Vigilance", "Ward", "Flash", "Defender",
	}

	for _, kw := range commonKeywords {
		if containsIgnoreCase(text, kw) {
			keywords[kw] = true
		}
	}

	return keywords
}

func extractCreatureTypes(typeLine string) map[string]bool {
	types := make(map[string]bool)

	// Find everything after the em-dash or hyphen
	dashIdx := -1
	for i := 0; i < len(typeLine); i++ {
		if typeLine[i] == '-' || (i < len(typeLine)-2 && typeLine[i:i+3] == "—") {
			dashIdx = i
			break
		}
	}

	if dashIdx < 0 {
		return types
	}

	// Skip the dash
	subtype := typeLine[dashIdx+1:]
	if len(subtype) > 0 && subtype[0] == '-' {
		subtype = subtype[1:]
	}

	// Split by spaces
	current := ""
	for i := 0; i <= len(subtype); i++ {
		if i == len(subtype) || subtype[i] == ' ' {
			if current != "" {
				types[current] = true
			}
			current = ""
		} else {
			current += string(subtype[i])
		}
	}

	return types
}

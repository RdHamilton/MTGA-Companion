package ml

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/meta"
)

// MetaWeightingConfig configures the meta weighting system.
type MetaWeightingConfig struct {
	// Tier weights (how much each tier contributes to scores)
	Tier1Weight float64
	Tier2Weight float64
	Tier3Weight float64
	Tier4Weight float64

	// TournamentBonus adds extra weight for tournament-proven cards
	TournamentBonus float64

	// MetaShareMultiplier scales the meta share contribution
	MetaShareMultiplier float64

	// TrendBoost adds extra weight for trending archetypes
	TrendBoostUp     float64
	TrendBoostDown   float64
	TrendBoostStable float64

	// CacheTTL is how long to cache meta scores
	CacheTTL time.Duration
}

// DefaultMetaWeightingConfig returns sensible defaults.
func DefaultMetaWeightingConfig() *MetaWeightingConfig {
	return &MetaWeightingConfig{
		Tier1Weight:         1.0,
		Tier2Weight:         0.75,
		Tier3Weight:         0.5,
		Tier4Weight:         0.25,
		TournamentBonus:     0.1,
		MetaShareMultiplier: 1.0,
		TrendBoostUp:        0.1,
		TrendBoostDown:      -0.05,
		TrendBoostStable:    0.0,
		CacheTTL:            30 * time.Minute,
	}
}

// MetaWeighter calculates meta-based scores for cards and archetypes.
type MetaWeighter struct {
	metaService *meta.Service
	config      *MetaWeightingConfig

	// Cached archetype scores per format
	archetypeScores map[string]map[string]*ArchetypeMetaScore
	cardScores      map[string]map[int]*CardMetaScore
	cacheTimes      map[string]time.Time
	mu              sync.RWMutex
}

// ArchetypeMetaScore represents meta-based scores for an archetype.
type ArchetypeMetaScore struct {
	ArchetypeName   string
	MetaShare       float64
	TournamentScore float64
	TierScore       float64
	TrendScore      float64
	OverallScore    float64 // Combined score (0.0-1.0)
	Confidence      float64 // How reliable is this score
	Colors          []string
	LastUpdated     time.Time
}

// CardMetaScore represents meta-based scores for a card.
type CardMetaScore struct {
	CardID           int
	Score            float64 // Overall meta score (0.0-1.0)
	ArchetypeMatches []string
	TopArchetype     string
	Confidence       float64
	Factors          []string
	LastUpdated      time.Time
}

// NewMetaWeighter creates a new meta weighter.
func NewMetaWeighter(metaService *meta.Service, config *MetaWeightingConfig) *MetaWeighter {
	if config == nil {
		config = DefaultMetaWeightingConfig()
	}

	return &MetaWeighter{
		metaService:     metaService,
		config:          config,
		archetypeScores: make(map[string]map[string]*ArchetypeMetaScore),
		cardScores:      make(map[string]map[int]*CardMetaScore),
		cacheTimes:      make(map[string]time.Time),
	}
}

// GetArchetypeScore returns the meta score for an archetype.
func (w *MetaWeighter) GetArchetypeScore(ctx context.Context, format, archetype string) (*ArchetypeMetaScore, error) {
	// Check cache first
	if score := w.getArchetypeFromCache(format, archetype); score != nil {
		return score, nil
	}

	// Fetch and calculate
	if err := w.refreshFormatCache(ctx, format); err != nil {
		return nil, err
	}

	return w.getArchetypeFromCache(format, archetype), nil
}

// GetCardMetaScore returns the meta score for a card based on its archetype fit.
func (w *MetaWeighter) GetCardMetaScore(ctx context.Context, format string, cardID int, cardColors []string, cardArchetype string) (*CardMetaScore, error) {
	// Check cache first
	if score := w.getCardFromCache(format, cardID); score != nil {
		return score, nil
	}

	// Refresh if needed
	if err := w.refreshFormatCache(ctx, format); err != nil {
		return nil, err
	}

	// Calculate card meta score
	return w.calculateCardMetaScore(format, cardID, cardColors, cardArchetype), nil
}

// GetTopArchetypes returns the top N archetypes by meta score.
func (w *MetaWeighter) GetTopArchetypes(ctx context.Context, format string, limit int) ([]*ArchetypeMetaScore, error) {
	if err := w.refreshFormatCache(ctx, format); err != nil {
		return nil, err
	}

	w.mu.RLock()
	formatCache := w.archetypeScores[strings.ToLower(format)]
	w.mu.RUnlock()

	if formatCache == nil {
		return []*ArchetypeMetaScore{}, nil
	}

	// Convert map to slice
	scores := make([]*ArchetypeMetaScore, 0, len(formatCache))
	for _, score := range formatCache {
		scores = append(scores, score)
	}

	// Sort by overall score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].OverallScore > scores[j].OverallScore
	})

	if limit <= 0 || limit > len(scores) {
		return scores, nil
	}

	return scores[:limit], nil
}

// GetMetaScoreForDeck calculates a meta score for a deck based on its composition.
func (w *MetaWeighter) GetMetaScoreForDeck(ctx context.Context, format string, colors []string, archetype string) (float64, error) {
	// Try to find matching archetype
	archetypeScore, err := w.GetArchetypeScore(ctx, format, archetype)
	if err == nil && archetypeScore != nil {
		return archetypeScore.OverallScore, nil
	}

	// Fall back to color-based matching
	w.mu.RLock()
	formatCache := w.archetypeScores[strings.ToLower(format)]
	w.mu.RUnlock()

	if formatCache == nil {
		return 0.5, nil // Neutral if no data
	}

	// Find archetypes with matching colors
	var bestScore float64
	for _, score := range formatCache {
		if w.colorsMatch(score.Colors, colors) {
			if score.OverallScore > bestScore {
				bestScore = score.OverallScore
			}
		}
	}

	if bestScore == 0 {
		return 0.5, nil // Neutral if no match
	}

	return bestScore, nil
}

// refreshFormatCache refreshes the cache for a format if needed.
func (w *MetaWeighter) refreshFormatCache(ctx context.Context, format string) error {
	formatLower := strings.ToLower(format)

	w.mu.RLock()
	cacheTime, exists := w.cacheTimes[formatLower]
	w.mu.RUnlock()

	if exists && time.Since(cacheTime) < w.config.CacheTTL {
		return nil // Cache is still valid
	}

	// Fetch fresh data
	aggregated, err := w.metaService.GetAggregatedMeta(ctx, format)
	if err != nil {
		return err
	}

	// Calculate scores for all archetypes
	scores := make(map[string]*ArchetypeMetaScore)
	for _, arch := range aggregated.Archetypes {
		score := w.calculateArchetypeScore(arch)
		scores[strings.ToLower(arch.Name)] = score
	}

	// Update cache
	w.mu.Lock()
	w.archetypeScores[formatLower] = scores
	w.cardScores[formatLower] = make(map[int]*CardMetaScore) // Clear card cache
	w.cacheTimes[formatLower] = time.Now()
	w.mu.Unlock()

	return nil
}

// calculateArchetypeScore calculates meta scores for an archetype.
func (w *MetaWeighter) calculateArchetypeScore(arch *meta.AggregatedArchetype) *ArchetypeMetaScore {
	score := &ArchetypeMetaScore{
		ArchetypeName: arch.Name,
		MetaShare:     arch.MetaShare,
		Colors:        arch.Colors,
		LastUpdated:   time.Now(),
	}

	// Calculate tier score (0.0-1.0)
	switch arch.Tier {
	case 1:
		score.TierScore = w.config.Tier1Weight
	case 2:
		score.TierScore = w.config.Tier2Weight
	case 3:
		score.TierScore = w.config.Tier3Weight
	default:
		score.TierScore = w.config.Tier4Weight
	}

	// Calculate tournament score (0.0-1.0)
	// Normalize based on typical top 8 counts (20+ is excellent)
	if arch.TournamentTop8s > 0 {
		score.TournamentScore = min(float64(arch.TournamentTop8s)/20.0, 1.0)
		// Bonus for tournament wins
		if arch.TournamentWins > 0 {
			score.TournamentScore = min(score.TournamentScore+float64(arch.TournamentWins)*0.05, 1.0)
		}
	}

	// Calculate trend score
	switch arch.TrendDirection {
	case "up":
		score.TrendScore = w.config.TrendBoostUp
	case "down":
		score.TrendScore = w.config.TrendBoostDown
	default:
		score.TrendScore = w.config.TrendBoostStable
	}

	// Calculate overall score
	// Combine: 40% tier, 30% tournament, 20% meta share, 10% trend
	metaShareNormalized := min(arch.MetaShare/20.0, 1.0) * w.config.MetaShareMultiplier

	score.OverallScore = (score.TierScore * 0.4) +
		(score.TournamentScore * 0.3) +
		(metaShareNormalized * 0.2) +
		((0.5 + score.TrendScore) * 0.1)

	// Clamp to 0.0-1.0
	score.OverallScore = max(0.0, min(1.0, score.OverallScore))

	// Calculate confidence based on data availability
	score.Confidence = arch.ConfidenceScore

	return score
}

// calculateCardMetaScore calculates meta score for a card.
func (w *MetaWeighter) calculateCardMetaScore(format string, cardID int, colors []string, archetype string) *CardMetaScore {
	formatLower := strings.ToLower(format)

	w.mu.RLock()
	formatCache := w.archetypeScores[formatLower]
	w.mu.RUnlock()

	score := &CardMetaScore{
		CardID:           cardID,
		Score:            0.5, // Neutral default
		ArchetypeMatches: make([]string, 0),
		Factors:          make([]string, 0),
		LastUpdated:      time.Now(),
	}

	if formatCache == nil {
		score.Factors = append(score.Factors, "No meta data available")
		return score
	}

	// Find matching archetypes
	var bestMatch *ArchetypeMetaScore
	var matches []*ArchetypeMetaScore

	for _, archScore := range formatCache {
		// Check for archetype name match
		if archetype != "" && strings.Contains(strings.ToLower(archScore.ArchetypeName), strings.ToLower(archetype)) {
			matches = append(matches, archScore)
			if bestMatch == nil || archScore.OverallScore > bestMatch.OverallScore {
				bestMatch = archScore
			}
		}

		// Check for color match
		if w.colorsMatch(archScore.Colors, colors) {
			matches = append(matches, archScore)
			if bestMatch == nil || archScore.OverallScore > bestMatch.OverallScore {
				bestMatch = archScore
			}
		}
	}

	// Deduplicate matches
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m.ArchetypeName] {
			score.ArchetypeMatches = append(score.ArchetypeMatches, m.ArchetypeName)
			seen[m.ArchetypeName] = true
		}
	}

	if bestMatch != nil {
		score.TopArchetype = bestMatch.ArchetypeName
		score.Score = bestMatch.OverallScore
		score.Confidence = bestMatch.Confidence

		// Add explanatory factors
		if bestMatch.TierScore >= 0.75 {
			score.Factors = append(score.Factors, "Fits tier 1-2 archetype")
		}
		if bestMatch.TournamentScore > 0.5 {
			score.Factors = append(score.Factors, "Tournament proven archetype")
		}
		if bestMatch.TrendScore > 0 {
			score.Factors = append(score.Factors, "Archetype is trending up")
		} else if bestMatch.TrendScore < 0 {
			score.Factors = append(score.Factors, "Archetype is trending down")
		}
	} else {
		score.Factors = append(score.Factors, "No strong archetype match")
	}

	// Cache the result
	w.mu.Lock()
	if w.cardScores[formatLower] == nil {
		w.cardScores[formatLower] = make(map[int]*CardMetaScore)
	}
	w.cardScores[formatLower][cardID] = score
	w.mu.Unlock()

	return score
}

// getArchetypeFromCache retrieves archetype score from cache.
func (w *MetaWeighter) getArchetypeFromCache(format, archetype string) *ArchetypeMetaScore {
	w.mu.RLock()
	defer w.mu.RUnlock()

	formatCache := w.archetypeScores[strings.ToLower(format)]
	if formatCache == nil {
		return nil
	}

	// Try exact match
	if score, ok := formatCache[strings.ToLower(archetype)]; ok {
		return score
	}

	// Try partial match
	archetypeLower := strings.ToLower(archetype)
	for name, score := range formatCache {
		if strings.Contains(name, archetypeLower) {
			return score
		}
	}

	return nil
}

// getCardFromCache retrieves card score from cache.
func (w *MetaWeighter) getCardFromCache(format string, cardID int) *CardMetaScore {
	w.mu.RLock()
	defer w.mu.RUnlock()

	formatCache := w.cardScores[strings.ToLower(format)]
	if formatCache == nil {
		return nil
	}

	return formatCache[cardID]
}

// colorsMatch checks if colors match (subset matching).
func (w *MetaWeighter) colorsMatch(archetypeColors, cardColors []string) bool {
	if len(cardColors) == 0 {
		return true // Colorless matches everything
	}

	// Check if card colors are a subset of archetype colors
	for _, cardColor := range cardColors {
		found := false
		for _, archColor := range archetypeColors {
			if cardColor == archColor {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// ClearCache clears all cached scores.
func (w *MetaWeighter) ClearCache() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.archetypeScores = make(map[string]map[string]*ArchetypeMetaScore)
	w.cardScores = make(map[string]map[int]*CardMetaScore)
	w.cacheTimes = make(map[string]time.Time)
}

// GetConfig returns the current configuration.
func (w *MetaWeighter) GetConfig() *MetaWeightingConfig {
	return w.config
}

// UpdateConfig updates the weighting configuration.
func (w *MetaWeighter) UpdateConfig(config *MetaWeightingConfig) {
	w.config = config
	w.ClearCache() // Clear cache to recalculate with new weights
}

// MetaWeightedScorer wraps the Model to add meta weighting.
type MetaWeightedScorer struct {
	model        *Model
	metaWeighter *MetaWeighter
}

// NewMetaWeightedScorer creates a scorer that combines ML model with meta weighting.
func NewMetaWeightedScorer(model *Model, metaWeighter *MetaWeighter) *MetaWeightedScorer {
	return &MetaWeightedScorer{
		model:        model,
		metaWeighter: metaWeighter,
	}
}

// ScoreCardsWithMeta scores cards using both ML model and meta weighting.
// It wraps the Model's ScoreCards and enhances the results with meta data.
func (s *MetaWeightedScorer) ScoreCardsWithMeta(ctx context.Context, format string, candidates []int, deck *DeckContext, accountID int) ([]*CardScore, error) {
	if s.model == nil {
		return s.scoreCardsMetaOnly(ctx, format, candidates, deck)
	}

	// Get base ML scores
	mlScores, err := s.model.ScoreCards(ctx, candidates, deck, accountID)
	if err != nil {
		return nil, err
	}

	// Enhance with meta scores
	for _, score := range mlScores {
		metaScore, metaErr := s.metaWeighter.GetCardMetaScore(ctx, format, score.CardID, deck.ColorIdentity, deck.Archetype)
		if metaErr == nil && metaScore != nil {
			// Update the meta component
			score.MetaScore = metaScore.Score

			// Add meta factors to explanation
			score.Factors = append(score.Factors, metaScore.Factors...)

			// Recalculate overall score with meta weight
			s.recalculateScore(score)

			// Blend confidence
			score.Confidence = (score.Confidence + metaScore.Confidence) / 2
		}
	}

	// Re-sort by updated scores
	sort.Slice(mlScores, func(i, j int) bool {
		return mlScores[i].Score > mlScores[j].Score
	})

	return mlScores, nil
}

// scoreCardsMetaOnly scores cards using only meta data when no ML model is available.
func (s *MetaWeightedScorer) scoreCardsMetaOnly(ctx context.Context, format string, candidates []int, deck *DeckContext) ([]*CardScore, error) {
	scores := make([]*CardScore, 0, len(candidates))

	for _, cardID := range candidates {
		metaScore, err := s.metaWeighter.GetCardMetaScore(ctx, format, cardID, deck.ColorIdentity, deck.Archetype)
		if err != nil {
			scores = append(scores, &CardScore{
				CardID:     cardID,
				Score:      0.5,
				MetaScore:  0.5,
				Confidence: 0.0,
				Factors:    []string{"No meta data available"},
			})
			continue
		}

		scores = append(scores, &CardScore{
			CardID:     cardID,
			Score:      metaScore.Score,
			MetaScore:  metaScore.Score,
			Confidence: metaScore.Confidence,
			Factors:    metaScore.Factors,
		})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores, nil
}

// recalculateScore recalculates the overall score including meta weight.
func (s *MetaWeightedScorer) recalculateScore(score *CardScore) {
	if s.model == nil || s.model.config == nil {
		return
	}

	config := s.model.config
	metaWeight := config.MetaWeight
	collaborativeWeight := config.CollaborativeWeight * (1.0 - metaWeight)
	contentWeight := (1.0 - config.CollaborativeWeight) * (1.0 - metaWeight)

	// Calculate weighted base score
	baseScore := (score.CollaborativeScore * collaborativeWeight) +
		(score.ContentScore * contentWeight) +
		(score.MetaScore * metaWeight)

	// Apply personal adjustment
	personalAdjustment := (score.PersonalScore - 0.5) * config.PersonalWeight
	score.Score = max(0.0, min(1.0, baseScore+personalAdjustment))
}

// GetMetaWeighter returns the underlying meta weighter.
func (s *MetaWeightedScorer) GetMetaWeighter() *MetaWeighter {
	return s.metaWeighter
}

// GetModel returns the underlying ML model.
func (s *MetaWeightedScorer) GetModel() *Model {
	return s.model
}

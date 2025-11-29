package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// ModelType identifies the type of ML model being used.
type ModelType string

const (
	// ModelTypeHybrid combines collaborative filtering with content-based approaches.
	ModelTypeHybrid ModelType = "hybrid"
	// ModelTypeCollaborative uses player/deck similarity patterns.
	ModelTypeCollaborative ModelType = "collaborative"
	// ModelTypeContentBased uses card feature similarity.
	ModelTypeContentBased ModelType = "content"
)

// ModelConfig holds configuration for the ML model.
type ModelConfig struct {
	// MinTrainingSamples is the minimum number of feedback samples needed before training.
	MinTrainingSamples int

	// CollaborativeWeight is the weight for collaborative filtering (0.0-1.0).
	// The remaining weight (1.0 - CollaborativeWeight) goes to content-based.
	CollaborativeWeight float64

	// PersonalWeight is the weight for personal preference learning (0.0-1.0).
	// This weight is applied on top of the model scores.
	PersonalWeight float64

	// MetaWeight is the weight for meta-game influence (0.0-1.0).
	MetaWeight float64

	// LearningRate controls how fast the model adapts to new data.
	LearningRate float64

	// DecayFactor controls how quickly old data loses influence (0.0-1.0).
	// Higher values mean slower decay (old data stays relevant longer).
	DecayFactor float64

	// SimilarityThreshold is the minimum similarity score for collaborative filtering.
	SimilarityThreshold float64
}

// DefaultModelConfig returns a ModelConfig with sensible defaults.
func DefaultModelConfig() *ModelConfig {
	return &ModelConfig{
		MinTrainingSamples:  50,
		CollaborativeWeight: 0.4,
		PersonalWeight:      0.3,
		MetaWeight:          0.1,
		LearningRate:        0.01,
		DecayFactor:         0.95,
		SimilarityThreshold: 0.5,
	}
}

// CardFeatures represents the feature vector for a card.
type CardFeatures struct {
	CardID        int
	ArenaID       string
	Name          string
	CMC           float64
	Colors        []string
	Types         []string
	Keywords      []string
	CreatureTypes []string
	Rarity        string
	SetCode       string

	// Derived features
	ColorCount     int
	IsCreature     bool
	IsInstant      bool
	IsSorcery      bool
	IsEnchantment  bool
	IsArtifact     bool
	IsLand         bool
	IsPlaneswalker bool
}

// CardScore represents an ML-generated score for a card.
type CardScore struct {
	CardID             int
	Score              float64  // Overall ML score (0.0-1.0)
	CollaborativeScore float64  // Collaborative filtering component
	ContentScore       float64  // Content-based component
	PersonalScore      float64  // Personal preference component
	MetaScore          float64  // Meta-game component
	Confidence         float64  // Model confidence in this score
	Factors            []string // Contributing factors for explanation
}

// DeckEmbedding represents a deck as a vector in feature space.
type DeckEmbedding struct {
	DeckID       string
	ColorProfile [5]float64         // WUBRG color weights
	CMCProfile   [8]float64         // CMC 0-7+ distribution
	TypeProfile  [7]float64         // Creature, Instant, Sorcery, Enchantment, Artifact, Land, Planeswalker
	KeywordFreq  map[string]float64 // Keyword frequency
	Archetype    string
	WinRate      float64
	MatchCount   int
}

// Model represents the ML recommendation model.
type Model struct {
	config *ModelConfig

	// Repositories
	feedbackRepo    repository.RecommendationFeedbackRepository
	performanceRepo repository.DeckPerformanceRepository

	// Card embeddings (card ID -> features)
	cardFeatures map[int]*CardFeatures
	cardMu       sync.RWMutex

	// Deck embeddings for collaborative filtering
	deckEmbeddings map[string]*DeckEmbedding

	// Card-card co-occurrence matrix (for collaborative filtering)
	// Maps card pair (lower ID, higher ID) -> co-occurrence score
	cardCooccurrence map[cardPair]float64
	cooccurrenceMu   sync.RWMutex

	// Card acceptance rates (for learning from feedback)
	cardAcceptanceRates map[int]*acceptanceStats
	acceptanceMu        sync.RWMutex

	// Archetype-card affinities (archetype -> card -> affinity score)
	archetypeAffinities map[string]map[int]float64
	affinityMu          sync.RWMutex

	// Personal preferences per account
	personalPreferences map[int]*PersonalPreferences
	personalMu          sync.RWMutex

	// Model metadata
	version         string
	lastTrainedAt   time.Time
	trainingSamples int
	mu              sync.RWMutex
}

// cardPair represents a pair of card IDs for co-occurrence tracking.
type cardPair struct {
	lowerID  int
	higherID int
}

// acceptanceStats tracks acceptance statistics for a card.
type acceptanceStats struct {
	Accepted       int
	Rejected       int
	WinsOnAccept   int
	LossesOnAccept int
	LastUpdated    time.Time
}

// PersonalPreferences stores learned preferences for a user.
type PersonalPreferences struct {
	AccountID       int
	PreferredColors map[string]float64 // Color preferences (W, U, B, R, G)
	PreferredTypes  map[string]float64 // Card type preferences
	CMCPreference   float64            // Preferred average CMC
	StyleProfile    map[string]float64 // Aggro, Control, Midrange, Tempo weights
	ArchetypePrefs  map[string]float64 // Archetype preferences
	LastUpdated     time.Time
}

// NewModel creates a new ML recommendation model.
func NewModel(
	feedbackRepo repository.RecommendationFeedbackRepository,
	performanceRepo repository.DeckPerformanceRepository,
	config *ModelConfig,
) *Model {
	if config == nil {
		config = DefaultModelConfig()
	}

	return &Model{
		config:              config,
		feedbackRepo:        feedbackRepo,
		performanceRepo:     performanceRepo,
		cardFeatures:        make(map[int]*CardFeatures),
		deckEmbeddings:      make(map[string]*DeckEmbedding),
		cardCooccurrence:    make(map[cardPair]float64),
		cardAcceptanceRates: make(map[int]*acceptanceStats),
		archetypeAffinities: make(map[string]map[int]float64),
		personalPreferences: make(map[int]*PersonalPreferences),
		version:             "1.0.0",
	}
}

// ScoreCards scores a list of candidate cards for recommendation.
func (m *Model) ScoreCards(ctx context.Context, candidates []int, deck *DeckContext, accountID int) ([]*CardScore, error) {
	if m == nil {
		return nil, fmt.Errorf("model is nil")
	}

	scores := make([]*CardScore, 0, len(candidates))

	// Build deck embedding for similarity comparisons
	deckEmbed := m.buildDeckEmbedding(deck)

	// Get personal preferences if available
	personal := m.getPersonalPreferences(accountID)

	for _, cardID := range candidates {
		score := m.scoreCard(ctx, cardID, deck, deckEmbed, personal)
		scores = append(scores, score)
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores, nil
}

// scoreCard calculates the ML score for a single card.
func (m *Model) scoreCard(ctx context.Context, cardID int, deck *DeckContext, deckEmbed *DeckEmbedding, personal *PersonalPreferences) *CardScore {
	score := &CardScore{
		CardID:  cardID,
		Factors: make([]string, 0),
	}

	// 1. Collaborative filtering score
	score.CollaborativeScore = m.collaborativeScore(cardID, deck.Cards, deckEmbed)

	// 2. Content-based score
	score.ContentScore = m.contentBasedScore(cardID, deck)

	// 3. Personal preference score
	if personal != nil {
		score.PersonalScore = m.personalScore(cardID, personal)
	} else {
		score.PersonalScore = 0.5 // Neutral if no personal data
	}

	// 4. Meta score (placeholder - will be implemented with meta integration)
	score.MetaScore = 0.5 // Neutral until meta data is available

	// Combine scores using configured weights
	collabWeight := m.config.CollaborativeWeight
	contentWeight := 1.0 - collabWeight

	baseScore := (score.CollaborativeScore * collabWeight) + (score.ContentScore * contentWeight)

	// Apply personal and meta adjustments
	personalAdjustment := (score.PersonalScore - 0.5) * m.config.PersonalWeight
	metaAdjustment := (score.MetaScore - 0.5) * m.config.MetaWeight

	score.Score = baseScore + personalAdjustment + metaAdjustment

	// Clamp to 0.0-1.0
	if score.Score < 0 {
		score.Score = 0
	} else if score.Score > 1 {
		score.Score = 1
	}

	// Calculate confidence based on data availability
	score.Confidence = m.calculateConfidence(cardID, deckEmbed.Archetype)

	// Build explanation factors
	if score.CollaborativeScore > 0.7 {
		score.Factors = append(score.Factors, "frequently picked with similar cards")
	}
	if score.ContentScore > 0.7 {
		score.Factors = append(score.Factors, "good fit for deck archetype")
	}
	if score.PersonalScore > 0.7 && personal != nil {
		score.Factors = append(score.Factors, "matches your play style")
	}

	return score
}

// collaborativeScore calculates similarity-based score using card co-occurrence.
func (m *Model) collaborativeScore(cardID int, deckCards []int, deckEmbed *DeckEmbedding) float64 {
	m.cooccurrenceMu.RLock()
	defer m.cooccurrenceMu.RUnlock()

	if len(m.cardCooccurrence) == 0 {
		return 0.5 // Neutral if no co-occurrence data
	}

	totalScore := 0.0
	matchCount := 0

	for _, existingCard := range deckCards {
		pair := makeCardPair(cardID, existingCard)
		if cooc, exists := m.cardCooccurrence[pair]; exists {
			totalScore += cooc
			matchCount++
		}
	}

	if matchCount == 0 {
		return 0.5 // Neutral if no matches
	}

	// Average co-occurrence score
	avgScore := totalScore / float64(matchCount)

	// Also consider archetype affinity
	m.affinityMu.RLock()
	archetypeAffinity := 0.0
	if affinities, exists := m.archetypeAffinities[deckEmbed.Archetype]; exists {
		if aff, hasCard := affinities[cardID]; hasCard {
			archetypeAffinity = aff
		}
	}
	m.affinityMu.RUnlock()

	// Combine co-occurrence and archetype affinity (70/30 split)
	return (avgScore * 0.7) + (archetypeAffinity * 0.3)
}

// contentBasedScore calculates score based on card feature similarity.
func (m *Model) contentBasedScore(cardID int, deck *DeckContext) float64 {
	m.cardMu.RLock()
	features, exists := m.cardFeatures[cardID]
	m.cardMu.RUnlock()

	if !exists {
		return 0.5 // Neutral if no feature data
	}

	score := 0.0
	factors := 0

	// Color fit
	colorScore := m.scoreColorFit(features.Colors, deck.ColorIdentity)
	score += colorScore
	factors++

	// CMC fit (prefer filling curve gaps)
	cmcScore := m.scoreCMCFit(features.CMC, deck.CMCDistribution)
	score += cmcScore
	factors++

	// Type balance
	typeScore := m.scoreTypeFit(features, deck.TypeDistribution)
	score += typeScore
	factors++

	// Synergy potential (keyword/creature type overlap)
	synergyScore := m.scoreSynergyPotential(features, deck.Keywords, deck.CreatureTypes)
	score += synergyScore
	factors++

	if factors == 0 {
		return 0.5
	}

	return score / float64(factors)
}

// personalScore calculates score based on personal preferences.
func (m *Model) personalScore(cardID int, prefs *PersonalPreferences) float64 {
	m.cardMu.RLock()
	features, exists := m.cardFeatures[cardID]
	m.cardMu.RUnlock()

	if !exists || prefs == nil {
		return 0.5
	}

	score := 0.0
	factors := 0

	// Color preference match
	for _, color := range features.Colors {
		if pref, ok := prefs.PreferredColors[color]; ok {
			score += pref
			factors++
		}
	}

	// Type preference match
	for _, cardType := range features.Types {
		if pref, ok := prefs.PreferredTypes[cardType]; ok {
			score += pref
			factors++
		}
	}

	if factors == 0 {
		return 0.5
	}

	return score / float64(factors)
}

// calculateConfidence estimates model confidence for this prediction.
func (m *Model) calculateConfidence(cardID int, archetype string) float64 {
	confidence := 0.3 // Base confidence

	// More data = higher confidence
	m.acceptanceMu.RLock()
	if stats, exists := m.cardAcceptanceRates[cardID]; exists {
		total := stats.Accepted + stats.Rejected
		if total > 10 {
			confidence += 0.2
		} else if total > 5 {
			confidence += 0.1
		}
	}
	m.acceptanceMu.RUnlock()

	// Archetype data available = higher confidence
	m.affinityMu.RLock()
	if _, exists := m.archetypeAffinities[archetype]; exists {
		confidence += 0.2
	}
	m.affinityMu.RUnlock()

	// More training samples = higher confidence
	m.mu.RLock()
	if m.trainingSamples > 100 {
		confidence += 0.2
	} else if m.trainingSamples > 50 {
		confidence += 0.1
	}
	m.mu.RUnlock()

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// scoreColorFit calculates how well card colors match deck identity.
func (m *Model) scoreColorFit(cardColors, deckColors []string) float64 {
	if len(cardColors) == 0 {
		return 1.0 // Colorless always fits
	}

	deckColorSet := make(map[string]bool)
	for _, c := range deckColors {
		deckColorSet[c] = true
	}

	matchCount := 0
	for _, color := range cardColors {
		if deckColorSet[color] {
			matchCount++
		}
	}

	if matchCount == len(cardColors) {
		return 1.0 // Perfect match
	}
	if matchCount == 0 {
		return 0.0 // No match
	}

	return float64(matchCount) / float64(len(cardColors))
}

// scoreCMCFit calculates how well a card fills curve gaps.
func (m *Model) scoreCMCFit(cardCMC float64, cmcDist map[int]int) float64 {
	cmc := int(cardCMC)

	// Ideal distribution for Limited
	ideal := map[int]int{1: 2, 2: 5, 3: 5, 4: 4, 5: 3, 6: 2}
	if cmc > 6 {
		cmc = 6
	}

	current := cmcDist[cmc]
	target := ideal[cmc]

	if current < target {
		// Under ideal - good
		gap := float64(target-current) / float64(target)
		return 0.7 + (gap * 0.3)
	}

	if current == target {
		return 0.6
	}

	// Over ideal - less good
	excess := float64(current-target) / float64(target)
	score := 0.5 - (excess * 0.3)
	if score < 0.1 {
		score = 0.1
	}
	return score
}

// scoreTypeFit evaluates card type balance.
func (m *Model) scoreTypeFit(features *CardFeatures, typeDist map[string]int) float64 {
	// Prefer maintaining creature-heavy balance for Limited
	creatureCount := typeDist["Creature"]
	totalCount := 0
	for _, count := range typeDist {
		totalCount += count
	}

	if totalCount == 0 {
		return 0.7 // New deck, most cards are good
	}

	creatureRatio := float64(creatureCount) / float64(totalCount)

	if features.IsCreature {
		// If low on creatures, creatures score higher
		if creatureRatio < 0.5 {
			return 0.8
		}
		return 0.6
	}

	// Non-creature spells
	if creatureRatio > 0.6 {
		return 0.7 // Could use more spells
	}
	return 0.5
}

// scoreSynergyPotential evaluates synergy with deck keywords/tribes.
func (m *Model) scoreSynergyPotential(features *CardFeatures, deckKeywords map[string]int, deckTribes map[string]int) float64 {
	synergy := 0.0

	// Keyword overlap
	for _, kw := range features.Keywords {
		if count, exists := deckKeywords[kw]; exists && count > 0 {
			synergy += 0.2
		}
	}

	// Tribal synergy
	for _, ct := range features.CreatureTypes {
		if count, exists := deckTribes[ct]; exists && count >= 3 {
			synergy += 0.3 // Strong tribal bonus
		}
	}

	// Cap and normalize
	if synergy > 1.0 {
		synergy = 1.0
	}

	// If no synergy found, return neutral
	if synergy == 0 {
		return 0.5
	}

	return synergy
}

// buildDeckEmbedding creates a feature vector representation of a deck.
func (m *Model) buildDeckEmbedding(deck *DeckContext) *DeckEmbedding {
	embed := &DeckEmbedding{
		DeckID:      deck.DeckID,
		KeywordFreq: make(map[string]float64),
		Archetype:   deck.Archetype,
	}

	// Build color profile
	colorIndex := map[string]int{"W": 0, "U": 1, "B": 2, "R": 3, "G": 4}
	for _, color := range deck.ColorIdentity {
		if idx, exists := colorIndex[color]; exists {
			embed.ColorProfile[idx] = 1.0
		}
	}

	// Build CMC profile from distribution
	for cmc, count := range deck.CMCDistribution {
		if cmc >= 0 && cmc < 8 {
			embed.CMCProfile[cmc] = float64(count)
		} else if cmc >= 8 {
			embed.CMCProfile[7] += float64(count)
		}
	}

	// Normalize CMC profile
	total := 0.0
	for _, v := range embed.CMCProfile {
		total += v
	}
	if total > 0 {
		for i := range embed.CMCProfile {
			embed.CMCProfile[i] /= total
		}
	}

	// Build type profile
	typeIndex := map[string]int{
		"Creature": 0, "Instant": 1, "Sorcery": 2,
		"Enchantment": 3, "Artifact": 4, "Land": 5, "Planeswalker": 6,
	}
	for cardType, count := range deck.TypeDistribution {
		if idx, exists := typeIndex[cardType]; exists {
			embed.TypeProfile[idx] = float64(count)
		}
	}

	// Normalize type profile
	total = 0.0
	for _, v := range embed.TypeProfile {
		total += v
	}
	if total > 0 {
		for i := range embed.TypeProfile {
			embed.TypeProfile[i] /= total
		}
	}

	// Build keyword frequency
	for kw, count := range deck.Keywords {
		embed.KeywordFreq[kw] = float64(count)
	}

	return embed
}

// getPersonalPreferences retrieves personal preferences for an account.
func (m *Model) getPersonalPreferences(accountID int) *PersonalPreferences {
	m.personalMu.RLock()
	defer m.personalMu.RUnlock()

	if prefs, exists := m.personalPreferences[accountID]; exists {
		return prefs
	}
	return nil
}

// makeCardPair creates a normalized card pair key.
func makeCardPair(id1, id2 int) cardPair {
	if id1 < id2 {
		return cardPair{lowerID: id1, higherID: id2}
	}
	return cardPair{lowerID: id2, higherID: id1}
}

// Train trains the model on available feedback and performance data.
func (m *Model) Train(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load feedback data
	feedbackData, err := m.feedbackRepo.GetForMLTraining(ctx, 10000)
	if err != nil {
		return fmt.Errorf("failed to load training data: %w", err)
	}

	if len(feedbackData) < m.config.MinTrainingSamples {
		return fmt.Errorf("insufficient training data: %d samples (need %d)", len(feedbackData), m.config.MinTrainingSamples)
	}

	// Train card acceptance model
	if err := m.trainAcceptanceModel(feedbackData); err != nil {
		return fmt.Errorf("failed to train acceptance model: %w", err)
	}

	// Train co-occurrence model
	if err := m.trainCooccurrenceModel(ctx); err != nil {
		return fmt.Errorf("failed to train co-occurrence model: %w", err)
	}

	// Train archetype affinity model
	if err := m.trainArchetypeModel(ctx); err != nil {
		return fmt.Errorf("failed to train archetype model: %w", err)
	}

	m.trainingSamples = len(feedbackData)
	m.lastTrainedAt = time.Now()

	return nil
}

// trainAcceptanceModel learns from user acceptance/rejection patterns.
func (m *Model) trainAcceptanceModel(feedback []*models.RecommendationFeedback) error {
	m.acceptanceMu.Lock()
	defer m.acceptanceMu.Unlock()

	for _, fb := range feedback {
		if fb.RecommendedCardID == nil {
			continue
		}

		cardID := *fb.RecommendedCardID
		stats, exists := m.cardAcceptanceRates[cardID]
		if !exists {
			stats = &acceptanceStats{}
			m.cardAcceptanceRates[cardID] = stats
		}

		switch fb.Action {
		case "accepted":
			stats.Accepted++
			if fb.OutcomeResult != nil {
				if *fb.OutcomeResult == "win" {
					stats.WinsOnAccept++
				} else {
					stats.LossesOnAccept++
				}
			}
		case "rejected":
			stats.Rejected++
		}

		stats.LastUpdated = time.Now()
	}

	return nil
}

// trainCooccurrenceModel learns card co-occurrence patterns from successful decks.
func (m *Model) trainCooccurrenceModel(ctx context.Context) error {
	// Load performance history for winning decks
	// Use accountID 0 to get all accounts
	history, err := m.performanceRepo.GetPerformanceByDateRange(ctx, 0, time.Now().AddDate(-1, 0, 0), time.Now())
	if err != nil {
		return fmt.Errorf("failed to load performance history: %w", err)
	}

	m.cooccurrenceMu.Lock()
	defer m.cooccurrenceMu.Unlock()

	// Build co-occurrence from winning deck compositions
	// This is a simplified version - a full implementation would parse deck card lists
	for _, perf := range history {
		if perf.Result != "win" {
			continue
		}

		// In a full implementation, we would:
		// 1. Load the deck's card list
		// 2. For each pair of cards, increment their co-occurrence score
		// 3. Weight by win rate and archetype relevance

		// For now, we track archetype-level co-occurrence through the archetype model
		_ = perf
	}

	return nil
}

// trainArchetypeModel learns archetype-card affinity patterns.
func (m *Model) trainArchetypeModel(ctx context.Context) error {
	// Load all archetype card weights from the repository
	archetypes, err := m.performanceRepo.ListArchetypes(ctx, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to load archetypes: %w", err)
	}

	m.affinityMu.Lock()
	defer m.affinityMu.Unlock()

	for _, arch := range archetypes {
		weights, err := m.performanceRepo.GetCardWeights(ctx, arch.ID)
		if err != nil {
			continue
		}

		affinities := make(map[int]float64)
		for _, w := range weights {
			// Normalize weight to 0.0-1.0 (weights are 0.0-10.0)
			affinities[w.CardID] = w.Weight / 10.0
		}

		m.archetypeAffinities[arch.Name] = affinities
	}

	return nil
}

// UpdateFromFeedback performs incremental learning from new feedback.
func (m *Model) UpdateFromFeedback(ctx context.Context, feedback *models.RecommendationFeedback) error {
	if feedback == nil || feedback.RecommendedCardID == nil {
		return nil
	}

	cardID := *feedback.RecommendedCardID

	m.acceptanceMu.Lock()
	stats, exists := m.cardAcceptanceRates[cardID]
	if !exists {
		stats = &acceptanceStats{}
		m.cardAcceptanceRates[cardID] = stats
	}

	// Apply learning rate for incremental update
	switch feedback.Action {
	case "accepted":
		stats.Accepted++
		if feedback.OutcomeResult != nil {
			if *feedback.OutcomeResult == "win" {
				stats.WinsOnAccept++
			} else {
				stats.LossesOnAccept++
			}
		}
	case "rejected":
		stats.Rejected++
	}

	stats.LastUpdated = time.Now()
	m.acceptanceMu.Unlock()

	return nil
}

// UpdatePersonalPreferences updates personal preferences for a user.
func (m *Model) UpdatePersonalPreferences(ctx context.Context, accountID int, deck *DeckContext, outcome string) error {
	m.personalMu.Lock()
	defer m.personalMu.Unlock()

	prefs, exists := m.personalPreferences[accountID]
	if !exists {
		prefs = &PersonalPreferences{
			AccountID:       accountID,
			PreferredColors: make(map[string]float64),
			PreferredTypes:  make(map[string]float64),
			StyleProfile:    make(map[string]float64),
			ArchetypePrefs:  make(map[string]float64),
		}
		m.personalPreferences[accountID] = prefs
	}

	// Only learn from wins to avoid reinforcing losing patterns
	if outcome != "win" {
		return nil
	}

	learningRate := m.config.LearningRate

	// Update color preferences
	for _, color := range deck.ColorIdentity {
		current := prefs.PreferredColors[color]
		prefs.PreferredColors[color] = current + (1.0-current)*learningRate
	}

	// Update type preferences
	for cardType, count := range deck.TypeDistribution {
		if count > 0 {
			current := prefs.PreferredTypes[cardType]
			prefs.PreferredTypes[cardType] = current + (1.0-current)*learningRate
		}
	}

	// Update archetype preference
	if deck.Archetype != "" {
		current := prefs.ArchetypePrefs[deck.Archetype]
		prefs.ArchetypePrefs[deck.Archetype] = current + (1.0-current)*learningRate
	}

	prefs.LastUpdated = time.Now()

	return nil
}

// RegisterCardFeatures registers feature data for a card.
func (m *Model) RegisterCardFeatures(cardID int, features *CardFeatures) {
	m.cardMu.Lock()
	defer m.cardMu.Unlock()
	m.cardFeatures[cardID] = features
}

// GetModelInfo returns information about the model state.
func (m *Model) GetModelInfo() *ModelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.cardMu.RLock()
	cardCount := len(m.cardFeatures)
	m.cardMu.RUnlock()

	m.cooccurrenceMu.RLock()
	coocCount := len(m.cardCooccurrence)
	m.cooccurrenceMu.RUnlock()

	m.affinityMu.RLock()
	archCount := len(m.archetypeAffinities)
	m.affinityMu.RUnlock()

	m.personalMu.RLock()
	personalCount := len(m.personalPreferences)
	m.personalMu.RUnlock()

	return &ModelInfo{
		Version:           m.version,
		LastTrainedAt:     m.lastTrainedAt,
		TrainingSamples:   m.trainingSamples,
		CardFeaturesCount: cardCount,
		CooccurrenceCount: coocCount,
		ArchetypeCount:    archCount,
		PersonalPrefCount: personalCount,
		IsReady:           m.trainingSamples >= m.config.MinTrainingSamples,
	}
}

// ModelInfo provides information about the model state.
type ModelInfo struct {
	Version           string
	LastTrainedAt     time.Time
	TrainingSamples   int
	CardFeaturesCount int
	CooccurrenceCount int
	ArchetypeCount    int
	PersonalPrefCount int
	IsReady           bool
}

// Serialize serializes the model to JSON for persistence.
func (m *Model) Serialize() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := struct {
		Version         string                       `json:"version"`
		LastTrainedAt   time.Time                    `json:"last_trained_at"`
		TrainingSamples int                          `json:"training_samples"`
		AcceptanceRates map[int]*acceptanceStats     `json:"acceptance_rates"`
		ArchetypeAffin  map[string]map[int]float64   `json:"archetype_affinities"`
		PersonalPrefs   map[int]*PersonalPreferences `json:"personal_preferences"`
	}{
		Version:         m.version,
		LastTrainedAt:   m.lastTrainedAt,
		TrainingSamples: m.trainingSamples,
	}

	m.acceptanceMu.RLock()
	data.AcceptanceRates = m.cardAcceptanceRates
	m.acceptanceMu.RUnlock()

	m.affinityMu.RLock()
	data.ArchetypeAffin = m.archetypeAffinities
	m.affinityMu.RUnlock()

	m.personalMu.RLock()
	data.PersonalPrefs = m.personalPreferences
	m.personalMu.RUnlock()

	return json.Marshal(data)
}

// Deserialize loads model state from JSON.
func (m *Model) Deserialize(data []byte) error {
	var loaded struct {
		Version         string                       `json:"version"`
		LastTrainedAt   time.Time                    `json:"last_trained_at"`
		TrainingSamples int                          `json:"training_samples"`
		AcceptanceRates map[int]*acceptanceStats     `json:"acceptance_rates"`
		ArchetypeAffin  map[string]map[int]float64   `json:"archetype_affinities"`
		PersonalPrefs   map[int]*PersonalPreferences `json:"personal_preferences"`
	}

	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to deserialize model: %w", err)
	}

	m.mu.Lock()
	m.version = loaded.Version
	m.lastTrainedAt = loaded.LastTrainedAt
	m.trainingSamples = loaded.TrainingSamples
	m.mu.Unlock()

	m.acceptanceMu.Lock()
	if loaded.AcceptanceRates != nil {
		m.cardAcceptanceRates = loaded.AcceptanceRates
	}
	m.acceptanceMu.Unlock()

	m.affinityMu.Lock()
	if loaded.ArchetypeAffin != nil {
		m.archetypeAffinities = loaded.ArchetypeAffin
	}
	m.affinityMu.Unlock()

	m.personalMu.Lock()
	if loaded.PersonalPrefs != nil {
		m.personalPreferences = loaded.PersonalPrefs
	}
	m.personalMu.Unlock()

	return nil
}

// DeckContext provides deck information for ML scoring.
type DeckContext struct {
	DeckID           string
	Cards            []int // Card IDs in deck
	ColorIdentity    []string
	CMCDistribution  map[int]int
	TypeDistribution map[string]int
	Keywords         map[string]int
	CreatureTypes    map[string]int
	Archetype        string
	Format           string
	SetCode          string
}

// CosineSimilarity calculates cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

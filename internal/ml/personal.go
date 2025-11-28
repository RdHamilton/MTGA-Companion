package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// PersonalLearner handles user-specific preference learning.
type PersonalLearner struct {
	performanceRepo repository.DeckPerformanceRepository
	feedbackRepo    repository.RecommendationFeedbackRepository

	// Personal profiles per account
	profiles   map[int]*PersonalProfile
	profilesMu sync.RWMutex

	config *PersonalLearnerConfig
}

// PersonalLearnerConfig configures personal learning behavior.
type PersonalLearnerConfig struct {
	// MinMatchesForPersonalization is the minimum matches needed before personalization kicks in.
	MinMatchesForPersonalization int

	// LearningRate controls how fast preferences adapt (0.0-1.0).
	LearningRate float64

	// DecayRate controls how quickly old data loses influence (0.0-1.0).
	// Higher values mean slower decay.
	DecayRate float64

	// MaxHistorySize is the maximum number of matches to consider.
	MaxHistorySize int

	// StyleDetectionEnabled enables automatic play style detection.
	StyleDetectionEnabled bool

	// ColorBiasDetectionEnabled enables color preference detection.
	ColorBiasDetectionEnabled bool
}

// DefaultPersonalLearnerConfig returns sensible defaults.
func DefaultPersonalLearnerConfig() *PersonalLearnerConfig {
	return &PersonalLearnerConfig{
		MinMatchesForPersonalization: 10,
		LearningRate:                 0.1,
		DecayRate:                    0.95,
		MaxHistorySize:               500,
		StyleDetectionEnabled:        true,
		ColorBiasDetectionEnabled:    true,
	}
}

// PersonalProfile represents a user's learned preferences.
type PersonalProfile struct {
	AccountID int `json:"account_id"`

	// Color preferences (W, U, B, R, G -> 0.0-1.0)
	ColorPreferences map[string]float64 `json:"color_preferences"`

	// Type preferences (Creature, Instant, etc. -> 0.0-1.0)
	TypePreferences map[string]float64 `json:"type_preferences"`

	// CMC preferences (average CMC they tend to win with)
	PreferredCMC   float64 `json:"preferred_cmc"`
	CMCVariance    float64 `json:"cmc_variance"`
	PreferAggro    bool    `json:"prefer_aggro"`    // CMC < 2.5
	PreferControl  bool    `json:"prefer_control"`  // CMC > 3.5
	PreferMidrange bool    `json:"prefer_midrange"` // CMC 2.5-3.5

	// Play style profile
	Style *PlayStyleProfile `json:"style"`

	// Archetype preferences (archetype name -> preference score)
	ArchetypePreferences map[string]float64 `json:"archetype_preferences"`

	// Card preferences (card ID -> preference score based on acceptance)
	CardPreferences map[int]float64 `json:"card_preferences"`

	// Performance statistics
	TotalMatches   int       `json:"total_matches"`
	TotalWins      int       `json:"total_wins"`
	WinRate        float64   `json:"win_rate"`
	LastMatchDate  time.Time `json:"last_match_date"`
	LastUpdateDate time.Time `json:"last_update_date"`

	// Confidence in the profile (based on data quantity)
	Confidence float64 `json:"confidence"`
}

// PlayStyleProfile represents detected play style preferences.
type PlayStyleProfile struct {
	// Primary style weights (should sum to ~1.0)
	Aggro    float64 `json:"aggro"`
	Control  float64 `json:"control"`
	Midrange float64 `json:"midrange"`
	Tempo    float64 `json:"tempo"`
	Combo    float64 `json:"combo"`

	// Derived characteristics
	PreferCreatureHeavy bool    `json:"prefer_creature_heavy"` // >60% creatures
	PreferSpellHeavy    bool    `json:"prefer_spell_heavy"`    // >40% non-creature spells
	PreferInteraction   bool    `json:"prefer_interaction"`    // High removal/counter count
	AvgCardAdvantage    float64 `json:"avg_card_advantage"`    // Cards that draw cards

	// Strategy preferences
	PreferGoWide bool `json:"prefer_go_wide"` // Token/swarm strategies
	PreferGoTall bool `json:"prefer_go_tall"` // Few big threats
}

// NewPersonalLearner creates a new personal learner.
func NewPersonalLearner(
	performanceRepo repository.DeckPerformanceRepository,
	feedbackRepo repository.RecommendationFeedbackRepository,
	config *PersonalLearnerConfig,
) *PersonalLearner {
	if config == nil {
		config = DefaultPersonalLearnerConfig()
	}

	return &PersonalLearner{
		performanceRepo: performanceRepo,
		feedbackRepo:    feedbackRepo,
		profiles:        make(map[int]*PersonalProfile),
		config:          config,
	}
}

// GetProfile retrieves or creates a personal profile for an account.
func (l *PersonalLearner) GetProfile(ctx context.Context, accountID int) (*PersonalProfile, error) {
	l.profilesMu.RLock()
	profile, exists := l.profiles[accountID]
	l.profilesMu.RUnlock()

	if exists {
		return profile, nil
	}

	// Create new profile
	profile = l.newProfile(accountID)

	// Try to load from performance history
	if err := l.loadProfileFromHistory(ctx, profile); err != nil {
		// Log but continue with empty profile
		_ = err
	}

	l.profilesMu.Lock()
	l.profiles[accountID] = profile
	l.profilesMu.Unlock()

	return profile, nil
}

// newProfile creates a new empty profile.
func (l *PersonalLearner) newProfile(accountID int) *PersonalProfile {
	return &PersonalProfile{
		AccountID:            accountID,
		ColorPreferences:     make(map[string]float64),
		TypePreferences:      make(map[string]float64),
		ArchetypePreferences: make(map[string]float64),
		CardPreferences:      make(map[int]float64),
		Style:                &PlayStyleProfile{},
		PreferredCMC:         3.0,
	}
}

// loadProfileFromHistory loads profile data from performance history.
func (l *PersonalLearner) loadProfileFromHistory(ctx context.Context, profile *PersonalProfile) error {
	if l.performanceRepo == nil {
		return nil
	}

	history, err := l.performanceRepo.GetHistoryByAccount(ctx, profile.AccountID, l.config.MaxHistorySize)
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(history) == 0 {
		return nil
	}

	// Analyze history to build profile
	l.analyzeHistory(profile, history)

	return nil
}

// analyzeHistory analyzes match history to build preferences.
func (l *PersonalLearner) analyzeHistory(profile *PersonalProfile, history []*models.DeckPerformanceHistory) {
	colorCounts := make(map[string]int)
	colorWins := make(map[string]int)
	archetypeCounts := make(map[string]int)
	archetypeWins := make(map[string]int)

	totalWins := 0
	totalMatches := len(history)

	for _, match := range history {
		isWin := match.Result == "win"
		if isWin {
			totalWins++
		}

		// Track colors
		for _, color := range match.ColorIdentity {
			colorStr := string(color)
			colorCounts[colorStr]++
			if isWin {
				colorWins[colorStr]++
			}
		}

		// Track archetypes
		if match.Archetype != nil && *match.Archetype != "" {
			arch := *match.Archetype
			archetypeCounts[arch]++
			if isWin {
				archetypeWins[arch]++
			}
		}

		// Update last match date
		if match.MatchTimestamp.After(profile.LastMatchDate) {
			profile.LastMatchDate = match.MatchTimestamp
		}
	}

	// Calculate color preferences (normalized win rate)
	for color, count := range colorCounts {
		if count >= 5 { // Minimum sample size
			winRate := 0.0
			if count > 0 {
				winRate = float64(colorWins[color]) / float64(count)
			}
			// Blend with count frequency for preference
			frequency := float64(count) / float64(totalMatches)
			profile.ColorPreferences[color] = (winRate * 0.7) + (frequency * 0.3)
		}
	}

	// Calculate archetype preferences
	for arch, count := range archetypeCounts {
		if count >= 3 {
			winRate := 0.0
			if count > 0 {
				winRate = float64(archetypeWins[arch]) / float64(count)
			}
			profile.ArchetypePreferences[arch] = winRate
		}
	}

	// Update stats
	profile.TotalMatches = totalMatches
	profile.TotalWins = totalWins
	if totalMatches > 0 {
		profile.WinRate = float64(totalWins) / float64(totalMatches)
	}

	// Calculate confidence
	profile.Confidence = l.calculateConfidence(totalMatches)

	profile.LastUpdateDate = time.Now()
}

// calculateConfidence calculates confidence based on data quantity.
func (l *PersonalLearner) calculateConfidence(matchCount int) float64 {
	if matchCount < l.config.MinMatchesForPersonalization {
		return 0
	}

	// Logarithmic growth: 10 matches = 0.3, 50 matches = 0.6, 200 matches = 0.9
	confidence := math.Log10(float64(matchCount)) / math.Log10(200.0)
	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

// LearnFromMatch updates preferences based on a match result.
func (l *PersonalLearner) LearnFromMatch(ctx context.Context, accountID int, match *MatchLearningData) error {
	profile, err := l.GetProfile(ctx, accountID)
	if err != nil {
		return err
	}

	l.profilesMu.Lock()
	defer l.profilesMu.Unlock()

	lr := l.config.LearningRate
	isWin := match.Result == "win"

	// Update color preferences
	for _, color := range match.DeckColors {
		current := profile.ColorPreferences[color]
		target := 0.0
		if isWin {
			target = 1.0
		}
		profile.ColorPreferences[color] = current + lr*(target-current)
	}

	// Update type preferences
	for cardType, count := range match.TypeDistribution {
		if count > 0 {
			current := profile.TypePreferences[cardType]
			target := 0.0
			if isWin {
				target = float64(count) / float64(match.TotalCards)
			}
			profile.TypePreferences[cardType] = current + lr*(target-current)
		}
	}

	// Update archetype preference
	if match.Archetype != "" {
		current := profile.ArchetypePreferences[match.Archetype]
		target := 0.0
		if isWin {
			target = 1.0
		}
		profile.ArchetypePreferences[match.Archetype] = current + lr*(target-current)
	}

	// Update CMC preference (only from wins)
	if isWin && match.AverageCMC > 0 {
		profile.PreferredCMC = profile.PreferredCMC + lr*(match.AverageCMC-profile.PreferredCMC)

		// Update style based on CMC
		if match.AverageCMC < 2.5 {
			profile.PreferAggro = true
		} else if match.AverageCMC > 3.5 {
			profile.PreferControl = true
		} else {
			profile.PreferMidrange = true
		}
	}

	// Update statistics
	profile.TotalMatches++
	if isWin {
		profile.TotalWins++
	}
	profile.WinRate = float64(profile.TotalWins) / float64(profile.TotalMatches)
	profile.LastMatchDate = time.Now()
	profile.LastUpdateDate = time.Now()
	profile.Confidence = l.calculateConfidence(profile.TotalMatches)

	// Detect play style
	if l.config.StyleDetectionEnabled {
		l.updatePlayStyle(profile, match, isWin)
	}

	return nil
}

// updatePlayStyle updates the play style profile based on match data.
func (l *PersonalLearner) updatePlayStyle(profile *PersonalProfile, match *MatchLearningData, isWin bool) {
	if !isWin {
		return // Only learn from wins
	}

	lr := l.config.LearningRate
	style := profile.Style

	// Determine style from deck composition
	creatureRatio := 0.0
	if match.TotalCards > 0 {
		creatureRatio = float64(match.TypeDistribution["Creature"]) / float64(match.TotalCards)
	}

	// Update creature preference
	style.PreferCreatureHeavy = creatureRatio > 0.6
	style.PreferSpellHeavy = creatureRatio < 0.4

	// Infer style from CMC and creature ratio
	if match.AverageCMC < 2.5 && creatureRatio > 0.5 {
		style.Aggro = style.Aggro + lr*(1.0-style.Aggro)
	} else if match.AverageCMC > 3.5 && creatureRatio < 0.5 {
		style.Control = style.Control + lr*(1.0-style.Control)
	} else if creatureRatio > 0.4 && creatureRatio < 0.6 {
		style.Midrange = style.Midrange + lr*(1.0-style.Midrange)
	}

	// Normalize style weights
	total := style.Aggro + style.Control + style.Midrange + style.Tempo + style.Combo
	if total > 0 {
		style.Aggro /= total
		style.Control /= total
		style.Midrange /= total
		style.Tempo /= total
		style.Combo /= total
	}
}

// LearnFromFeedback updates preferences based on recommendation feedback.
func (l *PersonalLearner) LearnFromFeedback(ctx context.Context, accountID int, feedback *models.RecommendationFeedback) error {
	if feedback.RecommendedCardID == nil {
		return nil
	}

	profile, err := l.GetProfile(ctx, accountID)
	if err != nil {
		return err
	}

	l.profilesMu.Lock()
	defer l.profilesMu.Unlock()

	cardID := *feedback.RecommendedCardID
	lr := l.config.LearningRate

	current := profile.CardPreferences[cardID]

	switch feedback.Action {
	case "accepted":
		// Card was accepted - increase preference
		profile.CardPreferences[cardID] = current + lr*(1.0-current)

		// If outcome is known, adjust further
		if feedback.OutcomeResult != nil {
			if *feedback.OutcomeResult == "win" {
				profile.CardPreferences[cardID] += lr * 0.2
			}
		}

	case "rejected":
		// Card was rejected - decrease preference
		profile.CardPreferences[cardID] = current - lr*current

	case "alternate":
		// User picked something else - slightly decrease
		profile.CardPreferences[cardID] = current - lr*current*0.5
	}

	// Clamp to 0-1
	if profile.CardPreferences[cardID] < 0 {
		profile.CardPreferences[cardID] = 0
	}
	if profile.CardPreferences[cardID] > 1 {
		profile.CardPreferences[cardID] = 1
	}

	profile.LastUpdateDate = time.Now()

	return nil
}

// GetPersonalScore calculates a personal preference score for a card.
func (l *PersonalLearner) GetPersonalScore(ctx context.Context, accountID int, card *CardFeatures) (float64, error) {
	profile, err := l.GetProfile(ctx, accountID)
	if err != nil {
		return 0.5, err
	}

	if profile.Confidence < 0.1 {
		return 0.5, nil // Not enough data
	}

	score := 0.0
	factors := 0

	// Color preference
	if len(card.Colors) > 0 {
		colorScore := 0.0
		for _, color := range card.Colors {
			if pref, ok := profile.ColorPreferences[color]; ok {
				colorScore += pref
			}
		}
		colorScore /= float64(len(card.Colors))
		score += colorScore
		factors++
	}

	// Type preference
	for _, cardType := range card.Types {
		if pref, ok := profile.TypePreferences[cardType]; ok {
			score += pref
			factors++
		}
	}

	// Direct card preference
	if pref, ok := profile.CardPreferences[card.CardID]; ok {
		score += pref * 2 // Weight direct preferences higher
		factors += 2
	}

	// CMC preference
	cmcDiff := math.Abs(card.CMC - profile.PreferredCMC)
	cmcScore := 1.0 - (cmcDiff / 5.0)
	if cmcScore < 0 {
		cmcScore = 0
	}
	score += cmcScore
	factors++

	if factors == 0 {
		return 0.5, nil
	}

	// Normalize and apply confidence
	normalizedScore := score / float64(factors)
	adjustedScore := 0.5 + (normalizedScore-0.5)*profile.Confidence

	return adjustedScore, nil
}

// ResetProfile resets a user's personal learning data.
func (l *PersonalLearner) ResetProfile(ctx context.Context, accountID int) error {
	l.profilesMu.Lock()
	defer l.profilesMu.Unlock()

	delete(l.profiles, accountID)
	return nil
}

// IsPersonalizationReady checks if personalization is available for a user.
func (l *PersonalLearner) IsPersonalizationReady(ctx context.Context, accountID int) (bool, float64) {
	profile, err := l.GetProfile(ctx, accountID)
	if err != nil {
		return false, 0
	}

	return profile.TotalMatches >= l.config.MinMatchesForPersonalization, profile.Confidence
}

// GetProfileStats returns statistics about a user's profile.
func (l *PersonalLearner) GetProfileStats(ctx context.Context, accountID int) (*PersonalProfileStats, error) {
	profile, err := l.GetProfile(ctx, accountID)
	if err != nil {
		return nil, err
	}

	stats := &PersonalProfileStats{
		AccountID:     accountID,
		TotalMatches:  profile.TotalMatches,
		TotalWins:     profile.TotalWins,
		WinRate:       profile.WinRate,
		Confidence:    profile.Confidence,
		LastMatchDate: profile.LastMatchDate,
		IsReady:       profile.TotalMatches >= l.config.MinMatchesForPersonalization,
	}

	// Find preferred colors
	for color, pref := range profile.ColorPreferences {
		if pref > 0.6 {
			stats.PreferredColors = append(stats.PreferredColors, color)
		}
	}

	// Find preferred archetypes
	for arch, pref := range profile.ArchetypePreferences {
		if pref > 0.55 {
			stats.PreferredArchetypes = append(stats.PreferredArchetypes, arch)
		}
	}

	// Determine primary style
	if profile.Style != nil {
		maxStyle := 0.0
		if profile.Style.Aggro > maxStyle {
			maxStyle = profile.Style.Aggro
			stats.PrimaryStyle = "Aggro"
		}
		if profile.Style.Control > maxStyle {
			maxStyle = profile.Style.Control
			stats.PrimaryStyle = "Control"
		}
		if profile.Style.Midrange > maxStyle {
			maxStyle = profile.Style.Midrange
			stats.PrimaryStyle = "Midrange"
		}
		if profile.Style.Tempo > maxStyle {
			maxStyle = profile.Style.Tempo
			stats.PrimaryStyle = "Tempo"
		}
		if profile.Style.Combo > maxStyle {
			stats.PrimaryStyle = "Combo"
		}
	}

	return stats, nil
}

// PersonalProfileStats provides statistics about a user's profile.
type PersonalProfileStats struct {
	AccountID           int       `json:"account_id"`
	TotalMatches        int       `json:"total_matches"`
	TotalWins           int       `json:"total_wins"`
	WinRate             float64   `json:"win_rate"`
	Confidence          float64   `json:"confidence"`
	LastMatchDate       time.Time `json:"last_match_date"`
	IsReady             bool      `json:"is_ready"`
	PreferredColors     []string  `json:"preferred_colors"`
	PreferredArchetypes []string  `json:"preferred_archetypes"`
	PrimaryStyle        string    `json:"primary_style"`
}

// MatchLearningData contains data from a match for learning.
type MatchLearningData struct {
	DeckID           string
	DeckColors       []string
	TypeDistribution map[string]int
	TotalCards       int
	AverageCMC       float64
	Archetype        string
	Result           string // "win" or "loss"
}

// Serialize serializes the personal learner state.
func (l *PersonalLearner) Serialize() ([]byte, error) {
	l.profilesMu.RLock()
	defer l.profilesMu.RUnlock()

	return json.Marshal(l.profiles)
}

// Deserialize loads personal learner state from JSON.
func (l *PersonalLearner) Deserialize(data []byte) error {
	var profiles map[int]*PersonalProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return fmt.Errorf("failed to deserialize profiles: %w", err)
	}

	l.profilesMu.Lock()
	l.profiles = profiles
	l.profilesMu.Unlock()

	return nil
}

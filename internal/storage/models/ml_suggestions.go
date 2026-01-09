package models

import (
	"encoding/json"
	"time"
)

// CardCombinationStats tracks how card pairs perform together.
type CardCombinationStats struct {
	ID      int64  `json:"id" db:"id"`
	CardID1 int    `json:"cardId1" db:"card_id_1"`
	CardID2 int    `json:"cardId2" db:"card_id_2"`
	DeckID  string `json:"deckId,omitempty" db:"deck_id"`
	Format  string `json:"format" db:"format"`

	// Co-occurrence metrics
	GamesTogether  int `json:"gamesTogether" db:"games_together"`
	GamesCard1Only int `json:"gamesCard1Only" db:"games_card1_only"`
	GamesCard2Only int `json:"gamesCard2Only" db:"games_card2_only"`

	// Win rate metrics
	WinsTogether  int `json:"winsTogether" db:"wins_together"`
	WinsCard1Only int `json:"winsCard1Only" db:"wins_card1_only"`
	WinsCard2Only int `json:"winsCard2Only" db:"wins_card2_only"`

	// Derived scores
	SynergyScore    float64 `json:"synergyScore" db:"synergy_score"`
	ConfidenceScore float64 `json:"confidenceScore" db:"confidence_score"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// WinRateTogether returns the win rate when both cards are present.
func (c *CardCombinationStats) WinRateTogether() float64 {
	if c.GamesTogether == 0 {
		return 0
	}
	return float64(c.WinsTogether) / float64(c.GamesTogether)
}

// WinRateCard1Only returns the win rate when only card 1 is present.
func (c *CardCombinationStats) WinRateCard1Only() float64 {
	if c.GamesCard1Only == 0 {
		return 0
	}
	return float64(c.WinsCard1Only) / float64(c.GamesCard1Only)
}

// WinRateCard2Only returns the win rate when only card 2 is present.
func (c *CardCombinationStats) WinRateCard2Only() float64 {
	if c.GamesCard2Only == 0 {
		return 0
	}
	return float64(c.WinsCard2Only) / float64(c.GamesCard2Only)
}

// MLSuggestion represents an ML-generated suggestion for deck improvement.
type MLSuggestion struct {
	ID             int64  `json:"id" db:"id"`
	DeckID         string `json:"deckId" db:"deck_id"`
	SuggestionType string `json:"suggestionType" db:"suggestion_type"` // add, remove, swap

	// Card references
	CardID          int    `json:"cardId,omitempty" db:"card_id"`
	CardName        string `json:"cardName,omitempty" db:"card_name"`
	SwapForCardID   int    `json:"swapForCardId,omitempty" db:"swap_for_card_id"`
	SwapForCardName string `json:"swapForCardName,omitempty" db:"swap_for_card_name"`

	// Scoring
	Confidence            float64 `json:"confidence" db:"confidence"`
	ExpectedWinRateChange float64 `json:"expectedWinRateChange" db:"expected_win_rate_change"`

	// Explanation
	Title       string `json:"title" db:"title"`
	Description string `json:"description,omitempty" db:"description"`
	Reasoning   string `json:"reasoning,omitempty" db:"reasoning"` // JSON array
	Evidence    string `json:"evidence,omitempty" db:"evidence"`   // JSON object

	// Status
	IsDismissed          bool     `json:"isDismissed" db:"is_dismissed"`
	WasApplied           bool     `json:"wasApplied" db:"was_applied"`
	OutcomeWinRateChange *float64 `json:"outcomeWinRateChange,omitempty" db:"outcome_win_rate_change"`

	// Timestamps
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	AppliedAt         *time.Time `json:"appliedAt,omitempty" db:"applied_at"`
	OutcomeRecordedAt *time.Time `json:"outcomeRecordedAt,omitempty" db:"outcome_recorded_at"`
}

// MLSuggestionReason represents a single reason for a suggestion.
type MLSuggestionReason struct {
	Type        string  `json:"type"` // synergy, performance, curve, meta
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`     // -1.0 to 1.0
	Confidence  float64 `json:"confidence"` // 0.0 to 1.0
}

// GetReasons parses the Reasoning JSON into structured reasons.
func (s *MLSuggestion) GetReasons() ([]MLSuggestionReason, error) {
	if s.Reasoning == "" {
		return nil, nil
	}
	var reasons []MLSuggestionReason
	if err := json.Unmarshal([]byte(s.Reasoning), &reasons); err != nil {
		return nil, err
	}
	return reasons, nil
}

// SetReasons serializes reasons to JSON.
func (s *MLSuggestion) SetReasons(reasons []MLSuggestionReason) error {
	data, err := json.Marshal(reasons)
	if err != nil {
		return err
	}
	s.Reasoning = string(data)
	return nil
}

// CardAffinity represents pre-computed synergy between two cards.
type CardAffinity struct {
	ID      int64  `json:"id" db:"id"`
	CardID1 int    `json:"cardId1" db:"card_id_1"`
	CardID2 int    `json:"cardId2" db:"card_id_2"`
	Format  string `json:"format" db:"format"`

	// Affinity metrics
	AffinityScore float64 `json:"affinityScore" db:"affinity_score"`
	SampleSize    int     `json:"sampleSize" db:"sample_size"`
	Confidence    float64 `json:"confidence" db:"confidence"`

	// Source
	Source string `json:"source" db:"source"` // historical, keyword, tribal, external

	// Timestamps
	ComputedAt time.Time `json:"computedAt" db:"computed_at"`
}

// UserPlayPatterns stores aggregated play style statistics for personalization.
type UserPlayPatterns struct {
	ID        int64  `json:"id" db:"id"`
	AccountID string `json:"accountId" db:"account_id"`

	// Archetype preferences
	PreferredArchetype string  `json:"preferredArchetype,omitempty" db:"preferred_archetype"`
	AggroAffinity      float64 `json:"aggroAffinity" db:"aggro_affinity"`
	MidrangeAffinity   float64 `json:"midrangeAffinity" db:"midrange_affinity"`
	ControlAffinity    float64 `json:"controlAffinity" db:"control_affinity"`
	ComboAffinity      float64 `json:"comboAffinity" db:"combo_affinity"`

	// Color preferences (JSON)
	ColorPreferences string `json:"colorPreferences,omitempty" db:"color_preferences"`

	// Play style metrics
	AvgGameLength    float64 `json:"avgGameLength" db:"avg_game_length"`
	AggressionScore  float64 `json:"aggressionScore" db:"aggression_score"`
	InteractionScore float64 `json:"interactionScore" db:"interaction_score"`

	// Sample sizes
	TotalMatches int `json:"totalMatches" db:"total_matches"`
	TotalDecks   int `json:"totalDecks" db:"total_decks"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// GetColorPreferences parses the color preferences JSON.
func (p *UserPlayPatterns) GetColorPreferences() (map[string]float64, error) {
	if p.ColorPreferences == "" {
		return make(map[string]float64), nil
	}
	var prefs map[string]float64
	if err := json.Unmarshal([]byte(p.ColorPreferences), &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}

// SetColorPreferences serializes color preferences to JSON.
func (p *UserPlayPatterns) SetColorPreferences(prefs map[string]float64) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	p.ColorPreferences = string(data)
	return nil
}

// MLModelMetadata tracks model versions and performance.
type MLModelMetadata struct {
	ID           int64  `json:"id" db:"id"`
	ModelName    string `json:"modelName" db:"model_name"`
	ModelVersion string `json:"modelVersion" db:"model_version"`

	// Training info
	TrainingSamples int        `json:"trainingSamples" db:"training_samples"`
	TrainingDate    *time.Time `json:"trainingDate,omitempty" db:"training_date"`

	// Performance metrics
	Accuracy       *float64 `json:"accuracy,omitempty" db:"accuracy"`
	PrecisionScore *float64 `json:"precisionScore,omitempty" db:"precision_score"`
	Recall         *float64 `json:"recall,omitempty" db:"recall"`
	F1Score        *float64 `json:"f1Score,omitempty" db:"f1_score"`

	// Model state
	IsActive  bool   `json:"isActive" db:"is_active"`
	ModelData []byte `json:"-" db:"model_data"` // Serialized model, not exposed in JSON

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// CardIndividualStats tracks how individual cards perform across all decks.
// This is used to calculate separate win rates for synergy scoring.
type CardIndividualStats struct {
	CardID     int       `json:"cardId" db:"card_id"`
	Format     string    `json:"format" db:"format"`
	TotalGames int       `json:"totalGames" db:"total_games"`
	Wins       int       `json:"wins" db:"wins"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// WinRate returns the win rate for this card.
func (c *CardIndividualStats) WinRate() float64 {
	if c.TotalGames == 0 {
		return 0
	}
	return float64(c.Wins) / float64(c.TotalGames)
}

// MLSuggestionType constants for suggestion types.
const (
	MLSuggestionTypeAdd    = "add"
	MLSuggestionTypeRemove = "remove"
	MLSuggestionTypeSwap   = "swap"
)

// AffinitySource constants for affinity sources.
const (
	AffinitySourceHistorical = "historical"
	AffinitySourceKeyword    = "keyword"
	AffinitySourceTribal     = "tribal"
	AffinitySourceExternal   = "external"
)

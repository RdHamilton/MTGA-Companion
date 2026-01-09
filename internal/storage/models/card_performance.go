package models

// CardPerformance represents performance metrics for a single card within a deck.
type CardPerformance struct {
	CardID   int    `json:"card_id"`
	CardName string `json:"card_name"`
	Quantity int    `json:"quantity"` // Number of copies in the deck

	// Core metrics
	GamesWithCard int `json:"games_with_card"` // Games where this card was in the deck
	GamesDrawn    int `json:"games_drawn"`     // Games where card was drawn
	GamesPlayed   int `json:"games_played"`    // Games where card was actually cast

	// Win rates
	WinRateWhenDrawn  float64 `json:"win_rate_when_drawn"`  // Win % in games where card was drawn
	WinRateWhenPlayed float64 `json:"win_rate_when_played"` // Win % in games where card was cast
	DeckWinRate       float64 `json:"deck_win_rate"`        // Overall deck win rate for comparison

	// Derived metrics
	PlayRate         float64 `json:"play_rate"`         // % of drawn games where card was played
	WinContribution  float64 `json:"win_contribution"`  // How much this card contributes to wins vs deck average
	ImpactScore      float64 `json:"impact_score"`      // Combined metric of card impact (-1 to +1)
	ConfidenceLevel  string  `json:"confidence_level"`  // "high", "medium", "low" based on sample size
	SampleSize       int     `json:"sample_size"`       // Number of games for this analysis
	PerformanceGrade string  `json:"performance_grade"` // "excellent", "good", "average", "poor", "bad"

	// Turn distribution
	AvgTurnPlayed  float64     `json:"avg_turn_played"`  // Average turn when card is played
	TurnPlayedDist map[int]int `json:"turn_played_dist"` // Distribution of turns when played

	// Mulligan stats
	MulliganedAway int     `json:"mulliganed_away"` // Times card was seen in mulligan hands
	MulliganRate   float64 `json:"mulligan_rate"`   // Rate of mulliganing hands with this card
}

// CardRecommendation represents a suggestion to add, remove, or swap a card.
type CardRecommendation struct {
	Type           string  `json:"type"` // "add", "remove", "swap"
	CardID         int     `json:"card_id"`
	CardName       string  `json:"card_name"`
	Reason         string  `json:"reason"`
	ImpactEstimate float64 `json:"impact_estimate"` // Estimated win rate change
	Confidence     string  `json:"confidence"`      // "high", "medium", "low"
	Priority       int     `json:"priority"`        // 1 = highest priority

	// For swap recommendations
	SwapForCardID   *int    `json:"swap_for_card_id,omitempty"`
	SwapForCardName *string `json:"swap_for_card_name,omitempty"`

	// Evidence
	BasedOnGames int `json:"based_on_games"` // Number of games this recommendation is based on
}

// DeckPerformanceAnalysis contains complete performance analysis for a deck.
type DeckPerformanceAnalysis struct {
	DeckID          string             `json:"deck_id"`
	DeckName        string             `json:"deck_name"`
	TotalMatches    int                `json:"total_matches"`
	TotalGames      int                `json:"total_games"`
	OverallWinRate  float64            `json:"overall_win_rate"`
	CardPerformance []*CardPerformance `json:"card_performance"`

	// Summary
	BestPerformers  []string `json:"best_performers"`  // Card names
	WorstPerformers []string `json:"worst_performers"` // Card names
	AnalysisDate    string   `json:"analysis_date"`
}

// CardPerformanceFilter provides filtering options for performance queries.
type CardPerformanceFilter struct {
	DeckID       string `json:"deck_id"`
	MinGames     int    `json:"min_games"`       // Minimum games for inclusion
	IncludeLands bool   `json:"include_lands"`   // Include basic lands in analysis
	DateRange    *int   `json:"date_range_days"` // Only analyze recent games (nil = all)
}

// RecommendationsRequest defines parameters for getting card recommendations.
type RecommendationsRequest struct {
	DeckID       string `json:"deck_id"`
	Format       string `json:"format,omitempty"`
	MaxResults   int    `json:"max_results"`
	IncludeSwaps bool   `json:"include_swaps"`
}

// RecommendationsResponse contains add/remove/swap recommendations for a deck.
type RecommendationsResponse struct {
	DeckID                string                `json:"deck_id"`
	DeckName              string                `json:"deck_name"`
	CurrentWinRate        float64               `json:"current_win_rate"`
	AddRecommendations    []*CardRecommendation `json:"add_recommendations"`
	RemoveRecommendations []*CardRecommendation `json:"remove_recommendations"`
	SwapRecommendations   []*CardRecommendation `json:"swap_recommendations"`
	ProjectedWinRate      float64               `json:"projected_win_rate"` // If all recommendations applied
}

// CardPlayEvent represents a card being played in a specific game.
type CardPlayEvent struct {
	MatchID     string `json:"match_id"`
	GameID      int    `json:"game_id"`
	CardID      int    `json:"card_id"`
	CardName    string `json:"card_name"`
	TurnNumber  int    `json:"turn_number"`
	Phase       string `json:"phase"`
	MatchResult string `json:"match_result"` // "win" or "loss"
}

// CardDrawEvent represents a card being drawn in a specific game.
type CardDrawEvent struct {
	MatchID     string `json:"match_id"`
	GameID      int    `json:"game_id"`
	CardID      int    `json:"card_id"`
	CardName    string `json:"card_name"`
	TurnDrawn   int    `json:"turn_drawn"`
	WasPlayed   bool   `json:"was_played"`
	MatchResult string `json:"match_result"`
}

// Performance grade thresholds
const (
	PerformanceGradeExcellent = "excellent" // > +10% vs deck average
	PerformanceGradeGood      = "good"      // +5% to +10%
	PerformanceGradeAverage   = "average"   // -5% to +5%
	PerformanceGradePoor      = "poor"      // -10% to -5%
	PerformanceGradeBad       = "bad"       // < -10%
)

// Confidence level thresholds
const (
	ConfidenceHigh   = "high"   // >= 30 games
	ConfidenceMedium = "medium" // 10-29 games
	ConfidenceLow    = "low"    // < 10 games
)

// MinGamesForAnalysis is the minimum number of games required for meaningful analysis.
const MinGamesForAnalysis = 5

// MinGamesForHighConfidence is the threshold for high confidence recommendations.
const MinGamesForHighConfidence = 30

// MinGamesForMediumConfidence is the threshold for medium confidence.
const MinGamesForMediumConfidence = 10

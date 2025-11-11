package seventeenlands

import "time"

// CardRating represents card performance statistics from 17Lands.
type CardRating struct {
	// Basic card information
	Name   string `json:"name"`
	Color  string `json:"color"`
	Rarity string `json:"rarity"`

	// MTGA identifiers
	MTGAID int `json:"mtga_id,omitempty"` // May not always be present

	// Win rate metrics
	GIHWR float64 `json:"ever_drawn_win_rate"`      // Games in Hand Win Rate
	OHWR  float64 `json:"opening_hand_win_rate"`    // Opening Hand Win Rate
	GPWR  float64 `json:"ever_drawn_game_win_rate"` // Game Present Win Rate
	GDWR  float64 `json:"drawn_win_rate"`           // Game Drawn Win Rate
	IHDWR float64 `json:"in_hand_win_rate"`         // In Hand Drawn Win Rate

	// Improvement metrics
	GIHWRDelta float64 `json:"ever_drawn_improvement_win_rate"`   // GIH Win Rate Delta
	OHWRDelta  float64 `json:"opening_hand_improvement_win_rate"` // OH Win Rate Delta
	GDWRDelta  float64 `json:"drawn_improvement_win_rate"`        // GD Win Rate Delta
	IHDWRDelta float64 `json:"in_hand_improvement_win_rate"`      // IHD Win Rate Delta

	// Pick metrics
	ALSA     float64 `json:"avg_seen"`            // Average Last Seen At
	ATA      float64 `json:"avg_pick"`            // Average Taken At
	PickRate float64 `json:"pick_rate,omitempty"` // Pick rate percentage

	// Game count metrics
	GIH int `json:"# ever_drawn"`    // Games in Hand count
	OH  int `json:"# opening_hand"`  // Opening Hand count
	GP  int `json:"# games"`         // Game Present count
	GD  int `json:"# drawn"`         // Game Drawn count
	IHD int `json:"# in_hand_drawn"` // In Hand Drawn count

	// Deck metrics
	GamesPlayed int `json:"# games_played,omitempty"` // Games played with this card
	NumberDecks int `json:"# decks,omitempty"`        // Number of decks with this card
}

// ColorRating represents color combination performance statistics from 17Lands.
type ColorRating struct {
	// Color combination (e.g., "W", "UB", "WUG")
	ColorName   string   `json:"color_name"`
	Colors      []string `json:"colors,omitempty"` // Parsed colors
	IsSplash    bool     `json:"is_splash,omitempty"`
	SplashColor string   `json:"splash_color,omitempty"`

	// Win rate metrics
	WinRate      float64 `json:"win_rate"`
	MatchWinRate float64 `json:"match_win_rate,omitempty"`
	GameWinRate  float64 `json:"game_win_rate,omitempty"`

	// Game counts
	GamesPlayed   int `json:"# games"`
	MatchesPlayed int `json:"# matches,omitempty"`
	Wins          int `json:"# wins,omitempty"`
	Losses        int `json:"# losses,omitempty"`

	// Deck metrics
	NumberDecks  int     `json:"# decks,omitempty"`
	AvgMainboard float64 `json:"avg_mainboard,omitempty"`
	AvgSideboard float64 `json:"avg_sideboard,omitempty"`
}

// CardRatingsResponse represents the response from the card ratings API.
type CardRatingsResponse []CardRating

// ColorRatingsResponse represents the response from the color ratings API.
type ColorRatingsResponse []ColorRating

// QueryParams holds parameters for 17Lands API queries.
type QueryParams struct {
	// Required parameters
	Expansion string // Set code (e.g., "BLB", "MKM")
	Format    string // Format (e.g., "PremierDraft", "QuickDraft", "TradDraft")

	// Optional parameters
	StartDate string   // YYYY-MM-DD format
	EndDate   string   // YYYY-MM-DD format
	Colors    []string // Color filter (e.g., ["W", "U"])

	// Event type (for color ratings)
	EventType string // e.g., "PremierDraft"

	// Additional options
	CombineSplash bool // Combine splash colors (for color ratings)
}

// ClientStats tracks 17Lands API client statistics.
type ClientStats struct {
	TotalRequests     int
	FailedRequests    int
	CachedResponses   int
	AverageLatency    time.Duration
	LastRequestTime   time.Time
	LastSuccessTime   time.Time
	LastFailureTime   time.Time
	ConsecutiveErrors int
}

// Error types for 17Lands API
const (
	ErrRateLimited   = "rate_limited"
	ErrUnavailable   = "unavailable"
	ErrInvalidParams = "invalid_params"
	ErrParseError    = "parse_error"
)

// APIError represents an error from the 17Lands API.
type APIError struct {
	Type       string
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

package seventeenlands

import "time"

// CardRating represents card performance statistics from 17Lands.
type CardRating struct {
	// Basic card information
	Name   string `json:"name"`
	Color  string `json:"color"`
	Rarity string `json:"rarity"`

	// Image URLs (contain Scryfall ID in path)
	URL     string `json:"url"`      // Front face image URL from Scryfall
	URLBack string `json:"url_back"` // Back face image URL (for DFCs)

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
	ErrRateLimited      = "rate_limited"
	ErrUnavailable      = "unavailable"
	ErrInvalidParams    = "invalid_params"
	ErrParseError       = "parse_error"
	ErrStatsUnavailable = "stats_unavailable" // No data available (API failed, no cache)
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

// SetFile represents a complete set file with card metadata and ratings from multiple sources.
type SetFile struct {
	Meta         SetMeta                    `json:"meta"`
	CardRatings  map[string]*CardRatingData `json:"card_ratings"`  // Keyed by Arena ID (as string)
	ColorRatings map[string]float64         `json:"color_ratings"` // Keyed by color combination (e.g., "W", "UB", "WUG")
}

// SetMeta contains metadata about a set file.
type SetMeta struct {
	Version               int       `json:"version"`
	SetCode               string    `json:"set_code"`
	DraftFormat           string    `json:"draft_format"`
	StartDate             string    `json:"start_date"`            // YYYY-MM-DD
	EndDate               string    `json:"end_date"`              // YYYY-MM-DD
	CollectionDate        time.Time `json:"collection_date"`       // When data was collected
	SeventeenLandsEnabled bool      `json:"17lands_enabled"`       // Whether 17Lands data is included
	ScryfallEnabled       bool      `json:"scryfall_enabled"`      // Whether Scryfall data is included
	ArenaFilesEnabled     bool      `json:"arena_files_enabled"`   // Whether Arena local files were used
	TotalCards            int       `json:"total_cards,omitempty"` // Total cards in set
	TotalGames            int       `json:"total_games,omitempty"` // Total games analyzed (for 17Lands)
}

// CardRatingData represents a card with combined metadata and ratings from multiple sources.
type CardRatingData struct {
	// Basic card information (from Scryfall or Arena files)
	Name       string   `json:"name"`
	ManaCost   string   `json:"mana_cost"`
	CMC        float64  `json:"cmc"`
	Types      []string `json:"types"`  // e.g., ["Creature", "Elf", "Warrior"]
	Colors     []string `json:"colors"` // e.g., ["G"]
	Rarity     string   `json:"rarity"` // common, uncommon, rare, mythic
	Images     []string `json:"images"` // URLs to card images
	OracleID   string   `json:"oracle_id,omitempty"`
	ScryfallID string   `json:"scryfall_id,omitempty"`
	ArenaID    int      `json:"arena_id,omitempty"`

	// 17Lands ratings by deck color combination
	// Keyed by color combination: "ALL", "W", "U", "B", "R", "G", "WU", "UB", etc.
	DeckColors map[string]*DeckColorRatings `json:"deck_colors,omitempty"`
}

// DeckColorRatings represents card performance metrics for a specific deck color combination.
type DeckColorRatings struct {
	// Win rate metrics
	GIHWR float64 `json:"GIHWR"`          // Games In Hand Win Rate (%)
	OHWR  float64 `json:"OHWR"`           // Opening Hand Win Rate (%)
	GPWR  float64 `json:"GPWR"`           // Game Present Win Rate (%)
	GDWR  float64 `json:"GDWR,omitempty"` // Game Drawn Win Rate (%)

	// Game count metrics
	GIH int `json:"GIH"`          // Games In Hand count
	OH  int `json:"OH,omitempty"` // Opening Hand count
	GP  int `json:"GP,omitempty"` // Game Present count
	GD  int `json:"GD,omitempty"` // Game Drawn count

	// Pick metrics
	ALSA float64 `json:"ALSA"` // Average Last Seen At (pick number)
	ATA  float64 `json:"ATA"`  // Average Taken At (pick number)
	IWD  float64 `json:"IWD"`  // Improvement When Drawn (percentage points)

	// Additional metrics
	PickRate    float64 `json:"pick_rate,omitempty"`    // Pick rate percentage
	GamesPlayed int     `json:"games_played,omitempty"` // Games played with this card
	NumberDecks int     `json:"number_decks,omitempty"` // Number of decks with this card
}

// DownloadProgress tracks the progress of downloading set file data.
type DownloadProgress struct {
	CurrentStep int     `json:"current_step"` // Current step number (1-based)
	TotalSteps  int     `json:"total_steps"`  // Total number of steps
	StepName    string  `json:"step_name"`    // Name of current step
	Message     string  `json:"message"`      // Detailed progress message
	Percent     float64 `json:"percent"`      // Overall completion percentage (0-100)

	// Per-step progress (for steps with sub-tasks like fetching multiple color combinations)
	SubTask        string  `json:"sub_task,omitempty"`         // Current sub-task (e.g., "W", "UB")
	SubTaskCurrent int     `json:"sub_task_current,omitempty"` // Current sub-task number
	SubTaskTotal   int     `json:"sub_task_total,omitempty"`   // Total sub-tasks
	SubTaskPercent float64 `json:"sub_task_percent,omitempty"` // Sub-task completion percentage

	// Timing
	StartTime     time.Time      `json:"start_time,omitempty"`
	EstimatedTime *time.Duration `json:"estimated_time,omitempty"` // Estimated time remaining

	// Errors
	Errors      []string `json:"errors,omitempty"`       // Non-fatal errors
	FailedTasks []string `json:"failed_tasks,omitempty"` // Failed sub-tasks (e.g., color combinations that failed)

	// Status
	Complete bool `json:"complete"` // Whether download is complete
	Failed   bool `json:"failed"`   // Whether download failed
}

// SetFileInfo represents metadata about a set file without loading the full file.
type SetFileInfo struct {
	FilePath       string    `json:"file_path"`
	SetCode        string    `json:"set_code"`
	DraftFormat    string    `json:"draft_format"`
	StartDate      string    `json:"start_date"`
	EndDate        string    `json:"end_date"`
	CollectionDate time.Time `json:"collection_date"`
	TotalCards     int       `json:"total_cards"`
	FileSize       int64     `json:"file_size"` // File size in bytes
	Enabled        bool      `json:"enabled"`   // Whether this set file is enabled for use
	Active         bool      `json:"active"`    // Whether this is the currently active set
}

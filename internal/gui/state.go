package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AppState represents the application's persistent state.
type AppState struct {
	mu sync.RWMutex

	// Filter states
	GlobalDateRange   DateRangeState    `json:"global_date_range"`
	MatchHistoryState MatchHistoryState `json:"match_history"`
	ChartsState       ChartsState       `json:"charts"`

	// Window state
	WindowSize WindowSize `json:"window_size"`

	// Last updated timestamp
	LastUpdated time.Time `json:"last_updated"`
}

// DateRangeState stores date range filter state.
type DateRangeState struct {
	StartDate string `json:"start_date"` // YYYY-MM-DD format
	EndDate   string `json:"end_date"`   // YYYY-MM-DD format
}

// MatchHistoryState stores match history view state.
type MatchHistoryState struct {
	SearchText   string `json:"search_text"`
	FormatFilter string `json:"format_filter"`
	ResultFilter string `json:"result_filter"`
	CurrentPage  int    `json:"current_page"`
}

// ChartsState stores charts view state.
type ChartsState struct {
	SelectedChartType string         `json:"selected_chart_type"`
	DateRange         DateRangeState `json:"date_range"`
	FormatFilter      string         `json:"format_filter"`
}

// WindowSize stores window dimensions.
type WindowSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NewAppState creates a new application state with defaults.
func NewAppState() *AppState {
	return &AppState{
		WindowSize: WindowSize{
			Width:  800,
			Height: 600,
		},
		LastUpdated: time.Now(),
	}
}

// statePath returns the path to the state file.
func statePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	stateDir := filepath.Join(homeDir, ".mtga-companion")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(stateDir, "state.json"), nil
}

// LoadState loads the application state from disk.
func LoadState() (*AppState, error) {
	path, err := statePath()
	if err != nil {
		return NewAppState(), nil
	}

	// If file doesn't exist, return new state
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewAppState(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return NewAppState(), nil
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return NewAppState(), nil
	}

	return &state, nil
}

// Save saves the application state to disk.
func (s *AppState) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastUpdated = time.Now()

	path, err := statePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// UpdateGlobalDateRange updates the global date range filter.
func (s *AppState) UpdateGlobalDateRange(startDate, endDate string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GlobalDateRange.StartDate = startDate
	s.GlobalDateRange.EndDate = endDate
}

// UpdateMatchHistoryState updates match history view state.
func (s *AppState) UpdateMatchHistoryState(search, format, result string, page int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MatchHistoryState.SearchText = search
	s.MatchHistoryState.FormatFilter = format
	s.MatchHistoryState.ResultFilter = result
	s.MatchHistoryState.CurrentPage = page
}

// UpdateChartsState updates charts view state.
func (s *AppState) UpdateChartsState(chartType, format string, dateRange DateRangeState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ChartsState.SelectedChartType = chartType
	s.ChartsState.FormatFilter = format
	s.ChartsState.DateRange = dateRange
}

// UpdateWindowSize updates window dimensions.
func (s *AppState) UpdateWindowSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.WindowSize.Width = width
	s.WindowSize.Height = height
}

// GetGlobalDateRange returns the global date range filter (thread-safe).
func (s *AppState) GetGlobalDateRange() DateRangeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.GlobalDateRange
}

// GetMatchHistoryState returns match history state (thread-safe).
func (s *AppState) GetMatchHistoryState() MatchHistoryState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MatchHistoryState
}

// GetChartsState returns charts state (thread-safe).
func (s *AppState) GetChartsState() ChartsState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ChartsState
}

// GetWindowSize returns window size (thread-safe).
func (s *AppState) GetWindowSize() WindowSize {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.WindowSize
}

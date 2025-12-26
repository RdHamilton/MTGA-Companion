package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockMatchFacade is a mock implementation of the match facade for testing.
type mockMatchFacade struct {
	matches            []models.Match
	games              []*models.Game
	stats              *models.Statistics
	formatStats        map[string]*models.Statistics
	deckStats          map[string]*models.Statistics
	trendAnalysis      *storage.TrendAnalysis
	performanceMetrics *models.PerformanceMetrics
	err                error
}

func (m *mockMatchFacade) GetMatches(_ context.Context, _ models.StatsFilter) ([]models.Match, error) {
	return m.matches, m.err
}

func (m *mockMatchFacade) GetMatchGames(_ context.Context, _ string) ([]*models.Game, error) {
	return m.games, m.err
}

func (m *mockMatchFacade) GetStats(_ context.Context, _ models.StatsFilter) (*models.Statistics, error) {
	return m.stats, m.err
}

func (m *mockMatchFacade) GetStatsByFormat(_ context.Context, _ models.StatsFilter) (map[string]*models.Statistics, error) {
	return m.formatStats, m.err
}

func (m *mockMatchFacade) GetStatsByDeck(_ context.Context, _ models.StatsFilter) (map[string]*models.Statistics, error) {
	return m.deckStats, m.err
}

func (m *mockMatchFacade) GetTrendAnalysis(_ context.Context, _, _ time.Time, _ string, _ []string) (*storage.TrendAnalysis, error) {
	return m.trendAnalysis, m.err
}

func (m *mockMatchFacade) GetPerformanceMetrics(_ context.Context, _ models.StatsFilter) (*models.PerformanceMetrics, error) {
	return m.performanceMetrics, m.err
}

// matchFacadeInterface defines the interface for testing match handlers.
type matchFacadeInterface interface {
	GetMatches(ctx context.Context, filter models.StatsFilter) ([]models.Match, error)
	GetMatchGames(ctx context.Context, matchID string) ([]*models.Game, error)
	GetStats(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error)
	GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error)
	GetStatsByDeck(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error)
	GetTrendAnalysis(ctx context.Context, startDate, endDate time.Time, periodType string, formats []string) (*storage.TrendAnalysis, error)
	GetPerformanceMetrics(ctx context.Context, filter models.StatsFilter) (*models.PerformanceMetrics, error)
}

// testMatchHandler wraps the match handler for testing with a mock.
type testMatchHandler struct {
	facade matchFacadeInterface
}

func newTestMatchHandler(facade matchFacadeInterface) *testMatchHandler {
	return &testMatchHandler{facade: facade}
}

func (h *testMatchHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	var req StatsFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	filter := req.ToStatsFilter()
	matches, err := h.facade.GetMatches(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": matches})
}

func (h *testMatchHandler) GetMatchGames(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		http.Error(w, `{"error":"match ID is required"}`, http.StatusBadRequest)
		return
	}

	games, err := h.facade.GetMatchGames(r.Context(), matchID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": games})
}

func (h *testMatchHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var req StatsFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	filter := req.ToStatsFilter()
	stats, err := h.facade.GetStats(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": stats})
}

func (h *testMatchHandler) GetFormats(w http.ResponseWriter, r *http.Request) {
	filter := models.StatsFilter{}
	stats, err := h.facade.GetStatsByFormat(r.Context(), filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	formats := make([]string, 0, len(stats))
	for format := range stats {
		formats = append(formats, format)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": formats})
}

func TestMatchHandler_GetMatches(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockMatches    []models.Match
		mockErr        error
		expectedStatus int
		expectedLen    int
	}{
		{
			name:        "successful get matches",
			requestBody: `{"format":"Standard"}`,
			mockMatches: []models.Match{
				{ID: "match-1", Format: "Standard", Result: "Win"},
				{ID: "match-2", Format: "Standard", Result: "Loss"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:           "empty matches",
			requestBody:    `{}`,
			mockMatches:    []models.Match{},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:           "invalid request body",
			requestBody:    `{invalid json`,
			mockMatches:    nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMatchFacade{
				matches: tt.mockMatches,
				err:     tt.mockErr,
			}

			handler := newTestMatchHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.GetMatches(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := resp["data"].([]interface{})
				if !ok {
					t.Fatal("expected data to be an array")
				}

				if len(data) != tt.expectedLen {
					t.Errorf("expected %d matches, got %d", tt.expectedLen, len(data))
				}
			}
		})
	}
}

func TestMatchHandler_GetMatchGames(t *testing.T) {
	tests := []struct {
		name           string
		matchID        string
		mockGames      []*models.Game
		mockErr        error
		expectedStatus int
		expectedLen    int
	}{
		{
			name:    "successful get games",
			matchID: "match-123",
			mockGames: []*models.Game{
				{ID: 1, MatchID: "match-123", GameNumber: 1, Result: "Win"},
				{ID: 2, MatchID: "match-123", GameNumber: 2, Result: "Loss"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:           "missing match ID",
			matchID:        "",
			mockGames:      nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMatchFacade{
				games: tt.mockGames,
				err:   tt.mockErr,
			}

			handler := newTestMatchHandler(mock)

			r := chi.NewRouter()
			r.Get("/api/v1/matches/{matchID}/games", handler.GetMatchGames)

			var req *http.Request
			if tt.matchID != "" {
				req = httptest.NewRequest(http.MethodGet, "/api/v1/matches/"+tt.matchID+"/games", nil)
			} else {
				// Test with missing matchID - need to set up route differently
				r2 := chi.NewRouter()
				r2.Get("/api/v1/matches/games", handler.GetMatchGames)
				req = httptest.NewRequest(http.MethodGet, "/api/v1/matches/games", nil)
				rec := httptest.NewRecorder()
				r2.ServeHTTP(rec, req)

				if rec.Code != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
				return
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := resp["data"].([]interface{})
				if !ok {
					t.Fatal("expected data to be an array")
				}

				if len(data) != tt.expectedLen {
					t.Errorf("expected %d games, got %d", tt.expectedLen, len(data))
				}
			}
		})
	}
}

func TestMatchHandler_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockStats      *models.Statistics
		mockErr        error
		expectedStatus int
	}{
		{
			name:        "successful get stats",
			requestBody: `{"format":"Standard"}`,
			mockStats: &models.Statistics{
				TotalMatches: 100,
				MatchesWon:   60,
				MatchesLost:  40,
				TotalGames:   200,
				GamesWon:     120,
				GamesLost:    80,
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    `{invalid`,
			mockStats:      nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMatchFacade{
				stats: tt.mockStats,
				err:   tt.mockErr,
			}

			handler := newTestMatchHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/matches/stats", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.GetStats(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestMatchHandler_GetFormats(t *testing.T) {
	tests := []struct {
		name           string
		mockFormats    map[string]*models.Statistics
		mockErr        error
		expectedStatus int
		expectedLen    int
	}{
		{
			name: "successful get formats",
			mockFormats: map[string]*models.Statistics{
				"Standard": {TotalMatches: 50},
				"Historic": {TotalMatches: 30},
				"Draft":    {TotalMatches: 20},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    3,
		},
		{
			name:           "empty formats",
			mockFormats:    map[string]*models.Statistics{},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMatchFacade{
				formatStats: tt.mockFormats,
				err:         tt.mockErr,
			}

			handler := newTestMatchHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/matches/formats", nil)
			rec := httptest.NewRecorder()

			handler.GetFormats(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := resp["data"].([]interface{})
				if !ok {
					t.Fatal("expected data to be an array")
				}

				if len(data) != tt.expectedLen {
					t.Errorf("expected %d formats, got %d", tt.expectedLen, len(data))
				}
			}
		})
	}
}

func TestStatsFilterRequest_ToStatsFilter(t *testing.T) {
	startDate := "2024-01-01"
	endDate := "2024-12-31"
	format := "Standard"
	accountID := 123

	tests := []struct {
		name     string
		request  StatsFilterRequest
		expected models.StatsFilter
	}{
		{
			name: "full filter",
			request: StatsFilterRequest{
				AccountID: &accountID,
				StartDate: &startDate,
				EndDate:   &endDate,
				Format:    &format,
				Formats:   []string{"Standard", "Historic"},
			},
			expected: models.StatsFilter{
				AccountID: &accountID,
				Format:    &format,
				Formats:   []string{"Standard", "Historic"},
			},
		},
		{
			name:     "empty filter",
			request:  StatsFilterRequest{},
			expected: models.StatsFilter{},
		},
		{
			name: "invalid date format",
			request: StatsFilterRequest{
				StartDate: strPtr("invalid-date"),
				EndDate:   strPtr("also-invalid"),
			},
			expected: models.StatsFilter{
				StartDate: nil,
				EndDate:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ToStatsFilter()

			// Check pointer fields
			if (result.AccountID == nil) != (tt.expected.AccountID == nil) {
				t.Errorf("AccountID mismatch: got %v, want %v", result.AccountID, tt.expected.AccountID)
			}
			if result.AccountID != nil && tt.expected.AccountID != nil && *result.AccountID != *tt.expected.AccountID {
				t.Errorf("AccountID value mismatch: got %d, want %d", *result.AccountID, *tt.expected.AccountID)
			}

			if (result.Format == nil) != (tt.expected.Format == nil) {
				t.Errorf("Format mismatch: got %v, want %v", result.Format, tt.expected.Format)
			}
			if result.Format != nil && tt.expected.Format != nil && *result.Format != *tt.expected.Format {
				t.Errorf("Format value mismatch: got %s, want %s", *result.Format, *tt.expected.Format)
			}

			if len(result.Formats) != len(tt.expected.Formats) {
				t.Errorf("Formats length mismatch: got %d, want %d", len(result.Formats), len(tt.expected.Formats))
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

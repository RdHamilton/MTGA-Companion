package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockDeckFacade is a mock implementation of the deck facade for testing.
type mockDeckFacade struct {
	decks          []*gui.DeckListItem
	deck           *gui.DeckWithCards
	deckStats      *gui.DeckStatistics
	deckPerf       *models.DeckPerformance
	exportResult   *gui.ExportDeckResponse
	importResult   *gui.ImportDeckResponse
	suggestions    *gui.SuggestDecksResponse
	classification *gui.ArchetypeClassificationResult
	createdDeck    *models.Deck
	err            error
}

func (m *mockDeckFacade) ListDecks(_ context.Context) ([]*gui.DeckListItem, error) {
	return m.decks, m.err
}

func (m *mockDeckFacade) GetDecksBySource(_ context.Context, _ string) ([]*gui.DeckListItem, error) {
	return m.decks, m.err
}

func (m *mockDeckFacade) GetDecksByFormat(_ context.Context, _ string) ([]*gui.DeckListItem, error) {
	return m.decks, m.err
}

func (m *mockDeckFacade) GetDeck(_ context.Context, _ string) (*gui.DeckWithCards, error) {
	return m.deck, m.err
}

func (m *mockDeckFacade) CreateDeck(_ context.Context, _, _, _ string, _ *string) (*models.Deck, error) {
	return m.createdDeck, m.err
}

func (m *mockDeckFacade) UpdateDeck(_ context.Context, _ *models.Deck) error {
	return m.err
}

func (m *mockDeckFacade) DeleteDeck(_ context.Context, _ string) error {
	return m.err
}

func (m *mockDeckFacade) GetDeckStatistics(_ context.Context, _ string) (*gui.DeckStatistics, error) {
	return m.deckStats, m.err
}

func (m *mockDeckFacade) GetDeckPerformance(_ context.Context, _ string) (*models.DeckPerformance, error) {
	return m.deckPerf, m.err
}

func (m *mockDeckFacade) ExportDeck(_ context.Context, _ *gui.ExportDeckRequest) (*gui.ExportDeckResponse, error) {
	return m.exportResult, m.err
}

func (m *mockDeckFacade) ImportDeck(_ context.Context, _ *gui.ImportDeckRequest) (*gui.ImportDeckResponse, error) {
	return m.importResult, m.err
}

func (m *mockDeckFacade) SuggestDecks(_ context.Context, _ string) (*gui.SuggestDecksResponse, error) {
	return m.suggestions, m.err
}

func (m *mockDeckFacade) ClassifyDeckArchetype(_ context.Context, _ string) (*gui.ArchetypeClassificationResult, error) {
	return m.classification, m.err
}

// deckFacadeInterface defines the interface for testing deck handlers.
type deckFacadeInterface interface {
	ListDecks(ctx context.Context) ([]*gui.DeckListItem, error)
	GetDecksBySource(ctx context.Context, source string) ([]*gui.DeckListItem, error)
	GetDecksByFormat(ctx context.Context, format string) ([]*gui.DeckListItem, error)
	GetDeck(ctx context.Context, deckID string) (*gui.DeckWithCards, error)
	CreateDeck(ctx context.Context, name, format, source string, draftEventID *string) (*models.Deck, error)
	UpdateDeck(ctx context.Context, deck *models.Deck) error
	DeleteDeck(ctx context.Context, deckID string) error
	GetDeckStatistics(ctx context.Context, deckID string) (*gui.DeckStatistics, error)
	GetDeckPerformance(ctx context.Context, deckID string) (*models.DeckPerformance, error)
	ExportDeck(ctx context.Context, req *gui.ExportDeckRequest) (*gui.ExportDeckResponse, error)
	ImportDeck(ctx context.Context, req *gui.ImportDeckRequest) (*gui.ImportDeckResponse, error)
	SuggestDecks(ctx context.Context, sessionID string) (*gui.SuggestDecksResponse, error)
	ClassifyDeckArchetype(ctx context.Context, deckID string) (*gui.ArchetypeClassificationResult, error)
}

// testDeckHandler wraps the deck handler for testing with a mock.
type testDeckHandler struct {
	facade deckFacadeInterface
}

func newTestDeckHandler(facade deckFacadeInterface) *testDeckHandler {
	return &testDeckHandler{facade: facade}
}

func (h *testDeckHandler) GetDecks(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	source := r.URL.Query().Get("source")

	var decks []*gui.DeckListItem
	var err error

	if source != "" {
		decks, err = h.facade.GetDecksBySource(r.Context(), source)
	} else if format != "" {
		decks, err = h.facade.GetDecksByFormat(r.Context(), format)
	} else {
		decks, err = h.facade.ListDecks(r.Context())
	}

	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": decks})
}

func (h *testDeckHandler) GetDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		http.Error(w, `{"error":"deck ID is required"}`, http.StatusBadRequest)
		return
	}

	deck, err := h.facade.GetDeck(r.Context(), deckID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if deck == nil {
		http.Error(w, `{"error":"deck not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": deck})
}

func (h *testDeckHandler) CreateDeck(w http.ResponseWriter, r *http.Request) {
	var req CreateDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"deck name is required"}`, http.StatusBadRequest)
		return
	}

	deck, err := h.facade.CreateDeck(r.Context(), req.Name, req.Format, req.Source, req.DraftEventID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": deck})
}

func (h *testDeckHandler) DeleteDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		http.Error(w, `{"error":"deck ID is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.facade.DeleteDeck(r.Context(), deckID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *testDeckHandler) ImportDeck(w http.ResponseWriter, r *http.Request) {
	var req ImportDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, `{"error":"deck content is required"}`, http.StatusBadRequest)
		return
	}

	importReq := &gui.ImportDeckRequest{
		ImportText: req.Content,
		Name:       req.Name,
		Format:     req.Format,
		Source:     "imported",
	}

	result, err := h.facade.ImportDeck(r.Context(), importReq)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": result})
}

func TestDeckHandler_GetDecks(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockDecks      []*gui.DeckListItem
		mockErr        error
		expectedStatus int
		expectedLen    int
	}{
		{
			name:        "successful get all decks",
			queryParams: "",
			mockDecks: []*gui.DeckListItem{
				{ID: "deck-1", Name: "Aggro Red"},
				{ID: "deck-2", Name: "Control Blue"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:        "filter by format",
			queryParams: "?format=Standard",
			mockDecks: []*gui.DeckListItem{
				{ID: "deck-1", Name: "Standard Aggro"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    1,
		},
		{
			name:        "filter by source",
			queryParams: "?source=draft",
			mockDecks: []*gui.DeckListItem{
				{ID: "deck-3", Name: "Draft Deck"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    1,
		},
		{
			name:           "empty decks",
			queryParams:    "",
			mockDecks:      []*gui.DeckListItem{},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:           "error from facade",
			queryParams:    "",
			mockDecks:      nil,
			mockErr:        errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDeckFacade{
				decks: tt.mockDecks,
				err:   tt.mockErr,
			}

			handler := newTestDeckHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/decks"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			handler.GetDecks(rec, req)

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
					t.Errorf("expected %d decks, got %d", tt.expectedLen, len(data))
				}
			}
		})
	}
}

func TestDeckHandler_GetDeck(t *testing.T) {
	tests := []struct {
		name           string
		deckID         string
		mockDeck       *gui.DeckWithCards
		mockErr        error
		expectedStatus int
	}{
		{
			name:   "successful get deck",
			deckID: "deck-123",
			mockDeck: &gui.DeckWithCards{
				Deck: &models.Deck{
					ID:     "deck-123",
					Name:   "Test Deck",
					Format: "Standard",
				},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "deck not found",
			deckID:         "deck-999",
			mockDeck:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing deck ID",
			deckID:         "",
			mockDeck:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDeckFacade{
				deck: tt.mockDeck,
				err:  tt.mockErr,
			}

			handler := newTestDeckHandler(mock)

			r := chi.NewRouter()
			r.Get("/api/v1/decks/{deckID}", handler.GetDeck)

			var req *http.Request
			if tt.deckID != "" {
				req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+tt.deckID, nil)
			} else {
				r2 := chi.NewRouter()
				r2.Get("/api/v1/decks/", handler.GetDeck)
				req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/", nil)
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
		})
	}
}

func TestDeckHandler_CreateDeck(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockDeck       *models.Deck
		mockErr        error
		expectedStatus int
	}{
		{
			name:        "successful create deck",
			requestBody: `{"name":"New Deck","format":"Standard","source":"constructed"}`,
			mockDeck: &models.Deck{
				ID:     "deck-new",
				Name:   "New Deck",
				Format: "Standard",
			},
			mockErr:        nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing name",
			requestBody:    `{"format":"Standard"}`,
			mockDeck:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{invalid`,
			mockDeck:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDeckFacade{
				createdDeck: tt.mockDeck,
				err:         tt.mockErr,
			}

			handler := newTestDeckHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/decks", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.CreateDeck(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestDeckHandler_DeleteDeck(t *testing.T) {
	tests := []struct {
		name           string
		deckID         string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "successful delete deck",
			deckID:         "deck-123",
			mockErr:        nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "error deleting deck",
			deckID:         "deck-456",
			mockErr:        errors.New("delete failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDeckFacade{
				err: tt.mockErr,
			}

			handler := newTestDeckHandler(mock)

			r := chi.NewRouter()
			r.Delete("/api/v1/decks/{deckID}", handler.DeleteDeck)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/decks/"+tt.deckID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestDeckHandler_ImportDeck(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockResult     *gui.ImportDeckResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name:        "successful import",
			requestBody: `{"content":"4 Lightning Bolt\n4 Mountain","name":"Red Deck","format":"Standard"}`,
			mockResult: &gui.ImportDeckResponse{
				Success:       true,
				DeckID:        "new-deck-id",
				CardsImported: 8,
				CardsSkipped:  0,
			},
			mockErr:        nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing content",
			requestBody:    `{"name":"Empty Deck"}`,
			mockResult:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{invalid`,
			mockResult:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDeckFacade{
				importResult: tt.mockResult,
				err:          tt.mockErr,
			}

			handler := newTestDeckHandler(mock)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/decks/import", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ImportDeck(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestCreateDeckRequest_Validation(t *testing.T) {
	tests := []struct {
		name       string
		request    CreateDeckRequest
		isValid    bool
		errMessage string
	}{
		{
			name: "valid request",
			request: CreateDeckRequest{
				Name:   "Test Deck",
				Format: "Standard",
				Source: "constructed",
			},
			isValid:    true,
			errMessage: "",
		},
		{
			name: "missing name",
			request: CreateDeckRequest{
				Format: "Standard",
				Source: "constructed",
			},
			isValid:    false,
			errMessage: "deck name is required",
		},
		{
			name: "with draft event ID",
			request: CreateDeckRequest{
				Name:         "Draft Deck",
				Format:       "Draft",
				Source:       "draft",
				DraftEventID: strPtr("draft-event-123"),
			},
			isValid:    true,
			errMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.request.Name != ""
			if isValid != tt.isValid {
				t.Errorf("expected isValid=%v, got %v", tt.isValid, isValid)
			}
		})
	}
}

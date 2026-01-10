package handlers

import (
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

// mockCollectionFacade is a mock implementation of the collection facade for testing.
type mockCollectionFacade struct {
	collection      *gui.CollectionResponse
	collectionStats *gui.CollectionStats
	setCompletion   []*models.SetCompletion
	recentChanges   []*gui.CollectionChangeEntry
	missingCards    *gui.MissingCardsForSetResponse
	collectionValue *gui.CollectionValue
	deckValue       *gui.DeckValue
	err             error
}

func (m *mockCollectionFacade) GetCollection(_ context.Context, _ *gui.CollectionFilter) (*gui.CollectionResponse, error) {
	return m.collection, m.err
}

func (m *mockCollectionFacade) GetCollectionStats(_ context.Context) (*gui.CollectionStats, error) {
	return m.collectionStats, m.err
}

func (m *mockCollectionFacade) GetSetCompletion(_ context.Context) ([]*models.SetCompletion, error) {
	return m.setCompletion, m.err
}

func (m *mockCollectionFacade) GetRecentChanges(_ context.Context, _ int) ([]*gui.CollectionChangeEntry, error) {
	return m.recentChanges, m.err
}

func (m *mockCollectionFacade) GetMissingCardsForSet(_ context.Context, _ string) (*gui.MissingCardsForSetResponse, error) {
	return m.missingCards, m.err
}

func (m *mockCollectionFacade) GetMissingCardsForDeck(_ context.Context, _ string) (*gui.MissingCardsForDeckResponse, error) {
	return nil, m.err
}

func (m *mockCollectionFacade) GetCollectionValue(_ context.Context) (*gui.CollectionValue, error) {
	return m.collectionValue, m.err
}

func (m *mockCollectionFacade) GetDeckValue(_ context.Context, _ string) (*gui.DeckValue, error) {
	return m.deckValue, m.err
}

// collectionFacadeInterface defines the interface for testing collection handlers.
type collectionFacadeInterface interface {
	GetCollection(ctx context.Context, filter *gui.CollectionFilter) (*gui.CollectionResponse, error)
	GetCollectionStats(ctx context.Context) (*gui.CollectionStats, error)
	GetSetCompletion(ctx context.Context) ([]*models.SetCompletion, error)
	GetRecentChanges(ctx context.Context, limit int) ([]*gui.CollectionChangeEntry, error)
	GetMissingCardsForSet(ctx context.Context, setCode string) (*gui.MissingCardsForSetResponse, error)
	GetMissingCardsForDeck(ctx context.Context, deckID string) (*gui.MissingCardsForDeckResponse, error)
	GetCollectionValue(ctx context.Context) (*gui.CollectionValue, error)
	GetDeckValue(ctx context.Context, deckID string) (*gui.DeckValue, error)
}

// testCollectionHandler wraps the collection handler for testing with a mock.
type testCollectionHandler struct {
	facade collectionFacadeInterface
}

func newTestCollectionHandler(facade collectionFacadeInterface) *testCollectionHandler {
	return &testCollectionHandler{facade: facade}
}

func (h *testCollectionHandler) GetCollectionValue(w http.ResponseWriter, r *http.Request) {
	value, err := h.facade.GetCollectionValue(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": value})
}

func (h *testCollectionHandler) GetDeckValue(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		http.Error(w, `{"error":"deck ID is required"}`, http.StatusBadRequest)
		return
	}

	value, err := h.facade.GetDeckValue(r.Context(), deckID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": value})
}

func TestCollectionHandler_GetCollectionValue(t *testing.T) {
	tests := []struct {
		name            string
		mockValue       *gui.CollectionValue
		mockErr         error
		expectedStatus  int
		expectValueData bool
	}{
		{
			name: "successful get collection value",
			mockValue: &gui.CollectionValue{
				TotalValueUSD:        1234.56,
				TotalValueEUR:        1100.00,
				UniqueCardsWithPrice: 500,
				CardCount:            1500,
				ValueByRarity: map[string]float64{
					"mythic": 500.00,
					"rare":   600.00,
					"common": 134.56,
				},
				TopCards: []*gui.CardValue{
					{CardID: 1, Name: "Black Lotus", SetCode: "LEA", Rarity: "rare", Quantity: 1, PriceUSD: 50000.00, TotalUSD: 50000.00},
				},
			},
			mockErr:         nil,
			expectedStatus:  http.StatusOK,
			expectValueData: true,
		},
		{
			name:            "empty collection value",
			mockValue:       &gui.CollectionValue{},
			mockErr:         nil,
			expectedStatus:  http.StatusOK,
			expectValueData: true,
		},
		{
			name:            "error from facade",
			mockValue:       nil,
			mockErr:         errors.New("database error"),
			expectedStatus:  http.StatusInternalServerError,
			expectValueData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCollectionFacade{
				collectionValue: tt.mockValue,
				err:             tt.mockErr,
			}

			handler := newTestCollectionHandler(mock)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/value", nil)
			rec := httptest.NewRecorder()

			handler.GetCollectionValue(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectValueData {
				var resp map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("expected data to be an object")
				}

				if tt.mockValue != nil && tt.mockValue.TotalValueUSD > 0 {
					if data["totalValueUsd"].(float64) != tt.mockValue.TotalValueUSD {
						t.Errorf("expected totalValueUsd %f, got %f", tt.mockValue.TotalValueUSD, data["totalValueUsd"].(float64))
					}
				}
			}
		})
	}
}

func TestCollectionHandler_GetDeckValue(t *testing.T) {
	tests := []struct {
		name            string
		deckID          string
		mockValue       *gui.DeckValue
		mockErr         error
		expectedStatus  int
		expectValueData bool
	}{
		{
			name:   "successful get deck value",
			deckID: "deck-123",
			mockValue: &gui.DeckValue{
				DeckID:         "deck-123",
				DeckName:       "Test Deck",
				TotalValueUSD:  250.50,
				TotalValueEUR:  225.00,
				CardCount:      60,
				CardsWithPrice: 55,
				TopCards: []*gui.CardValue{
					{CardID: 1, Name: "Mox Diamond", SetCode: "STH", Rarity: "rare", Quantity: 1, PriceUSD: 100.00, TotalUSD: 100.00},
				},
			},
			mockErr:         nil,
			expectedStatus:  http.StatusOK,
			expectValueData: true,
		},
		{
			name:            "missing deck ID",
			deckID:          "",
			mockValue:       nil,
			mockErr:         nil,
			expectedStatus:  http.StatusBadRequest,
			expectValueData: false,
		},
		{
			name:            "error from facade",
			deckID:          "deck-456",
			mockValue:       nil,
			mockErr:         errors.New("deck not found"),
			expectedStatus:  http.StatusInternalServerError,
			expectValueData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCollectionFacade{
				deckValue: tt.mockValue,
				err:       tt.mockErr,
			}

			handler := newTestCollectionHandler(mock)

			r := chi.NewRouter()
			r.Get("/api/v1/decks/{deckID}/value", handler.GetDeckValue)

			var req *http.Request
			if tt.deckID != "" {
				req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/"+tt.deckID+"/value", nil)
			} else {
				// Test missing deck ID by hitting endpoint without parameter
				r2 := chi.NewRouter()
				r2.Get("/api/v1/decks/value", handler.GetDeckValue)
				req = httptest.NewRequest(http.MethodGet, "/api/v1/decks/value", nil)
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

			if tt.expectValueData {
				var resp map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				data, ok := resp["data"].(map[string]interface{})
				if !ok {
					t.Fatal("expected data to be an object")
				}

				if tt.mockValue != nil {
					if data["deckId"].(string) != tt.mockValue.DeckID {
						t.Errorf("expected deckId %s, got %s", tt.mockValue.DeckID, data["deckId"].(string))
					}
					if data["totalValueUsd"].(float64) != tt.mockValue.TotalValueUSD {
						t.Errorf("expected totalValueUsd %f, got %f", tt.mockValue.TotalValueUSD, data["totalValueUsd"].(float64))
					}
				}
			}
		})
	}
}

func TestCollectionHandler_GetCollectionValueTypes(t *testing.T) {
	// Test that all response fields are correctly serialized
	lastUpdated := int64(1704067200)
	mock := &mockCollectionFacade{
		collectionValue: &gui.CollectionValue{
			TotalValueUSD:        999.99,
			TotalValueEUR:        888.88,
			UniqueCardsWithPrice: 100,
			CardCount:            200,
			ValueByRarity: map[string]float64{
				"mythic":   300.00,
				"rare":     400.00,
				"uncommon": 200.00,
				"common":   99.99,
			},
			TopCards: []*gui.CardValue{
				{
					CardID:   12345,
					Name:     "Test Card",
					SetCode:  "TST",
					Rarity:   "mythic",
					Quantity: 4,
					PriceUSD: 25.00,
					TotalUSD: 100.00,
				},
			},
			LastUpdated: &lastUpdated,
		},
		err: nil,
	}

	handler := newTestCollectionHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/collection/value", nil)
	rec := httptest.NewRecorder()

	handler.GetCollectionValue(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data := resp["data"].(map[string]interface{})

	// Verify nested structures
	valueByRarity, ok := data["valueByRarity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected valueByRarity to be a map")
	}
	if valueByRarity["mythic"].(float64) != 300.00 {
		t.Errorf("expected mythic value 300.00, got %f", valueByRarity["mythic"].(float64))
	}

	topCards, ok := data["topCards"].([]interface{})
	if !ok {
		t.Fatal("expected topCards to be an array")
	}
	if len(topCards) != 1 {
		t.Errorf("expected 1 top card, got %d", len(topCards))
	}

	topCard := topCards[0].(map[string]interface{})
	if topCard["name"].(string) != "Test Card" {
		t.Errorf("expected top card name 'Test Card', got %s", topCard["name"].(string))
	}
}

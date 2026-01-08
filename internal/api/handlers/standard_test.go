package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockStandardStorage implements storage methods needed by StandardHandler for testing.
type mockStandardStorage struct {
	standardSets          []*models.StandardSet
	upcomingRotation      *models.UpcomingRotation
	rotationAffectedDecks []*models.RotationAffectedDeck
	config                *models.StandardConfig
	cardLegality          *models.CardLegality
	cardsLegality         map[string]*models.CardLegality
	deck                  *models.Deck
	deckCards             []*models.DeckCard
	cardNames             map[string]string
	err                   error
}

// mockStandardRepo implements StandardRepository interface.
type mockStandardRepo struct {
	storage *mockStandardStorage
}

func (m *mockStandardRepo) GetStandardSets(_ context.Context) ([]*models.StandardSet, error) {
	return m.storage.standardSets, m.storage.err
}

func (m *mockStandardRepo) GetUpcomingRotation(_ context.Context) (*models.UpcomingRotation, error) {
	return m.storage.upcomingRotation, m.storage.err
}

func (m *mockStandardRepo) GetRotationAffectedDecks(_ context.Context) ([]*models.RotationAffectedDeck, error) {
	return m.storage.rotationAffectedDecks, m.storage.err
}

func (m *mockStandardRepo) GetConfig(_ context.Context) (*models.StandardConfig, error) {
	return m.storage.config, m.storage.err
}

func (m *mockStandardRepo) UpdateConfig(_ context.Context, _ *models.StandardConfig) error {
	return m.storage.err
}

func (m *mockStandardRepo) GetCardLegality(_ context.Context, _ string) (*models.CardLegality, error) {
	return m.storage.cardLegality, m.storage.err
}

func (m *mockStandardRepo) GetCardsLegality(_ context.Context, _ []string) (map[string]*models.CardLegality, error) {
	return m.storage.cardsLegality, m.storage.err
}

func (m *mockStandardRepo) UpdateCardLegality(_ context.Context, _ string, _ *models.CardLegality) error {
	return m.storage.err
}

func (m *mockStandardRepo) UpdateSetStandardStatus(_ context.Context, _ string, _ bool, _ *string) error {
	return m.storage.err
}

// mockDeckRepo implements DeckRepository interface methods needed for testing.
type mockDeckRepo struct {
	storage *mockStandardStorage
}

func (m *mockDeckRepo) GetByID(_ context.Context, _ string) (*models.Deck, error) {
	return m.storage.deck, m.storage.err
}

func (m *mockDeckRepo) GetCards(_ context.Context, _ string) ([]*models.DeckCard, error) {
	return m.storage.deckCards, m.storage.err
}

// testStandardHandler wraps handler for testing with mocks.
type testStandardHandler struct {
	mockStorage *mockStandardStorage
}

func newTestStandardHandler() *testStandardHandler {
	return &testStandardHandler{
		mockStorage: &mockStandardStorage{},
	}
}

// GetStandardSets handler for testing.
func (h *testStandardHandler) GetStandardSets(w http.ResponseWriter, r *http.Request) {
	repo := &mockStandardRepo{storage: h.mockStorage}
	sets, err := repo.GetStandardSets(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": sets})
}

// GetUpcomingRotation handler for testing.
func (h *testStandardHandler) GetUpcomingRotation(w http.ResponseWriter, r *http.Request) {
	repo := &mockStandardRepo{storage: h.mockStorage}
	rotation, err := repo.GetUpcomingRotation(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": rotation})
}

// GetRotationAffectedDecks handler for testing.
func (h *testStandardHandler) GetRotationAffectedDecks(w http.ResponseWriter, r *http.Request) {
	repo := &mockStandardRepo{storage: h.mockStorage}
	decks, err := repo.GetRotationAffectedDecks(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": decks})
}

// GetStandardConfig handler for testing.
func (h *testStandardHandler) GetStandardConfig(w http.ResponseWriter, r *http.Request) {
	repo := &mockStandardRepo{storage: h.mockStorage}
	config, err := repo.GetConfig(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": config})
}

// GetCardLegality handler for testing.
func (h *testStandardHandler) GetCardLegality(w http.ResponseWriter, r *http.Request) {
	arenaID := chi.URLParam(r, "arenaID")
	if arenaID == "" {
		http.Error(w, `{"error":"arena ID is required"}`, http.StatusBadRequest)
		return
	}

	repo := &mockStandardRepo{storage: h.mockStorage}
	legality, err := repo.GetCardLegality(r.Context(), arenaID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if legality == nil {
		http.Error(w, `{"error":"card legality not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": legality})
}

func TestStandardHandler_GetStandardSets(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.standardSets = []*models.StandardSet{
		{Code: "DSK", Name: "Duskmourn: House of Horror", IsStandardLegal: true},
		{Code: "BLB", Name: "Bloomburrow", IsStandardLegal: true},
		{Code: "OTJ", Name: "Outlaws of Thunder Junction", IsStandardLegal: true},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/sets", nil)
	rr := httptest.NewRecorder()

	handler.GetStandardSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) != 3 {
		t.Errorf("Expected 3 sets, got %d", len(data))
	}
}

func TestStandardHandler_GetStandardSets_Empty(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.standardSets = []*models.StandardSet{}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/sets", nil)
	rr := httptest.NewRecorder()

	handler.GetStandardSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestStandardHandler_GetUpcomingRotation(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.upcomingRotation = &models.UpcomingRotation{
		NextRotationDate:  "2025-09-01",
		DaysUntilRotation: 237,
		RotatingSets: []models.StandardSet{
			{Code: "WOE", Name: "Wilds of Eldraine", IsStandardLegal: true},
			{Code: "LCI", Name: "Lost Caverns of Ixalan", IsStandardLegal: true},
		},
		RotatingCardCount: 150,
		AffectedDecks:     5,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/rotation", nil)
	rr := httptest.NewRecorder()

	handler.GetUpcomingRotation(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be an object")
	}

	if data["nextRotationDate"] != "2025-09-01" {
		t.Errorf("Expected nextRotationDate '2025-09-01', got %v", data["nextRotationDate"])
	}
}

func TestStandardHandler_GetRotationAffectedDecks(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.rotationAffectedDecks = []*models.RotationAffectedDeck{
		{
			DeckID:            "deck-1",
			DeckName:          "Mono Red Aggro",
			Format:            "Standard",
			RotatingCardCount: 8,
			TotalCards:        60,
			PercentAffected:   13.3,
			RotatingCards:     []models.RotatingCard{},
		},
		{
			DeckID:            "deck-2",
			DeckName:          "Azorius Control",
			Format:            "Standard",
			RotatingCardCount: 12,
			TotalCards:        60,
			PercentAffected:   20.0,
			RotatingCards:     []models.RotatingCard{},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/rotation/affected-decks", nil)
	rr := httptest.NewRecorder()

	handler.GetRotationAffectedDecks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 affected decks, got %d", len(data))
	}
}

func TestStandardHandler_GetStandardConfig(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.config = &models.StandardConfig{
		ID:               1,
		NextRotationDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		RotationEnabled:  true,
		UpdatedAt:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/config", nil)
	rr := httptest.NewRecorder()

	handler.GetStandardConfig(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be an object")
	}

	if data["rotationEnabled"] != true {
		t.Errorf("Expected rotationEnabled true, got %v", data["rotationEnabled"])
	}
}

func TestStandardHandler_GetCardLegality(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.cardLegality = &models.CardLegality{
		Standard:  "not_legal",
		Historic:  "legal",
		Explorer:  "legal",
		Pioneer:   "legal",
		Modern:    "legal",
		Alchemy:   "not_legal",
		Brawl:     "legal",
		Commander: "legal",
	}

	r := chi.NewRouter()
	r.Get("/api/v1/standard/cards/{arenaID}/legality", handler.GetCardLegality)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/cards/12345/legality", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be an object")
	}

	if data["standard"] != "not_legal" {
		t.Errorf("Expected standard 'not_legal', got %v", data["standard"])
	}
}

func TestStandardHandler_GetCardLegality_NotFound(t *testing.T) {
	handler := newTestStandardHandler()
	handler.mockStorage.cardLegality = nil

	r := chi.NewRouter()
	r.Get("/api/v1/standard/cards/{arenaID}/legality", handler.GetCardLegality)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/cards/99999/legality", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestStandardHandler_GetCardLegality_MissingArenaID(t *testing.T) {
	handler := newTestStandardHandler()

	// Test without chi router to simulate missing param
	req := httptest.NewRequest(http.MethodGet, "/api/v1/standard/cards//legality", nil)
	rr := httptest.NewRecorder()

	handler.GetCardLegality(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestValidateDeckForStandard_LegalDeck(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 4, Board: "main"},
		{CardID: 2, Quantity: 4, Board: "main"},
		{CardID: 3, Quantity: 4, Board: "main"},
		{CardID: 4, Quantity: 4, Board: "main"},
		{CardID: 5, Quantity: 20, Board: "main"}, // Basic lands
		{CardID: 6, Quantity: 24, Board: "main"}, // More lands
	}

	legalities := map[string]*models.CardLegality{
		"1": {Standard: "legal"},
		"2": {Standard: "legal"},
		"3": {Standard: "legal"},
		"4": {Standard: "legal"},
		"5": {Standard: "legal"},
		"6": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Card A",
		"2": "Card B",
		"3": "Card C",
		"4": "Card D",
		"5": "Plains",
		"6": "Island",
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	if !result.IsLegal {
		t.Errorf("Expected deck to be legal, but got errors: %v", result.Errors)
	}
}

func TestValidateDeckForStandard_BannedCard(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 4, Board: "main"},
		{CardID: 2, Quantity: 56, Board: "main"},
	}

	legalities := map[string]*models.CardLegality{
		"1": {Standard: "banned"},
		"2": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Banned Card",
		"2": "Plains",
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	if result.IsLegal {
		t.Error("Expected deck to be illegal due to banned card")
	}

	foundBannedError := false
	for _, err := range result.Errors {
		if err.Reason == "banned" {
			foundBannedError = true
			break
		}
	}

	if !foundBannedError {
		t.Error("Expected banned card error")
	}
}

func TestValidateDeckForStandard_TooManyCopies(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 5, Board: "main"}, // 5 copies - illegal
		{CardID: 2, Quantity: 55, Board: "main"},
	}

	legalities := map[string]*models.CardLegality{
		"1": {Standard: "legal"},
		"2": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Non-Basic Card",
		"2": "Plains",
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	if result.IsLegal {
		t.Error("Expected deck to be illegal due to 5 copies")
	}

	foundCopiesError := false
	for _, err := range result.Errors {
		if err.Reason == "too_many_copies" {
			foundCopiesError = true
			break
		}
	}

	if !foundCopiesError {
		t.Error("Expected too_many_copies error")
	}
}

func TestValidateDeckForStandard_BasicLandsExempt(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 24, Board: "main"}, // 24 Plains - legal (basic land exempt)
		{CardID: 2, Quantity: 4, Board: "main"},
		{CardID: 3, Quantity: 32, Board: "main"},
	}

	legalities := map[string]*models.CardLegality{
		"1": {Standard: "legal"},
		"2": {Standard: "legal"},
		"3": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Plains", // Basic land - exempt from 4-copy rule
		"2": "Lightning Bolt",
		"3": "Island", // Basic land
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	if !result.IsLegal {
		t.Errorf("Expected deck to be legal (basic lands exempt), but got errors: %v", result.Errors)
	}
}

func TestValidateDeckForStandard_DeckTooSmall(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 30, Board: "main"}, // Only 30 cards
	}

	legalities := map[string]*models.CardLegality{
		"1": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Some Card",
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	if result.IsLegal {
		t.Error("Expected deck to be illegal due to size < 60")
	}

	foundSizeError := false
	for _, err := range result.Errors {
		if err.Reason == "deck_size" {
			foundSizeError = true
			break
		}
	}

	if !foundSizeError {
		t.Error("Expected deck_size error")
	}
}

func TestValidateDeckForStandard_NotLegalCard(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 4, Board: "main"},
		{CardID: 2, Quantity: 56, Board: "main"},
	}

	legalities := map[string]*models.CardLegality{
		"1": {Standard: "not_legal"},
		"2": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Old Card",
		"2": "Plains",
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	if result.IsLegal {
		t.Error("Expected deck to be illegal due to not_legal card")
	}

	foundNotLegalError := false
	for _, err := range result.Errors {
		if err.Reason == "not_legal" {
			foundNotLegalError = true
			break
		}
	}

	if !foundNotLegalError {
		t.Error("Expected not_legal error")
	}
}

func TestValidateDeckForStandard_UnknownLegality(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 1, Quantity: 4, Board: "main"},
		{CardID: 2, Quantity: 56, Board: "main"},
	}

	legalities := map[string]*models.CardLegality{
		// Card 1 not in legalities map
		"2": {Standard: "legal"},
	}

	cardNames := map[string]string{
		"1": "Unknown Card",
		"2": "Plains",
	}

	result := validateDeckForStandard(cards, legalities, cardNames)

	// Should have warning, not error
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for unknown legality")
	}

	foundUnknownWarning := false
	for _, warn := range result.Warnings {
		if warn.Type == "unknown_legality" {
			foundUnknownWarning = true
			break
		}
	}

	if !foundUnknownWarning {
		t.Error("Expected unknown_legality warning")
	}
}

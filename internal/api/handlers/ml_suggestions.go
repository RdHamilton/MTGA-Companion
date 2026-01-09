package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/analysis"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// MLSuggestionsHandler handles ML-powered suggestion API requests.
type MLSuggestionsHandler struct {
	mlRepo   *repository.MLSuggestionRepository
	mlEngine *analysis.MLEngine
}

// NewMLSuggestionsHandler creates a new MLSuggestionsHandler.
func NewMLSuggestionsHandler(
	mlRepo *repository.MLSuggestionRepository,
	mlEngine *analysis.MLEngine,
) *MLSuggestionsHandler {
	return &MLSuggestionsHandler{
		mlRepo:   mlRepo,
		mlEngine: mlEngine,
	}
}

// GetMLSuggestions returns ML-powered suggestions for a deck.
// GET /decks/{deckID}/ml-suggestions
func (h *MLSuggestionsHandler) GetMLSuggestions(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	activeOnly := r.URL.Query().Get("active") != "false"

	var suggestions interface{}
	var err error

	if activeOnly {
		suggestions, err = h.mlRepo.GetActiveSuggestions(r.Context(), deckID)
	} else {
		suggestions, err = h.mlRepo.GetSuggestionsByDeck(r.Context(), deckID)
	}

	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, suggestions)
}

// GenerateMLSuggestions generates new ML-powered suggestions for a deck.
// POST /decks/{deckID}/ml-suggestions/generate
func (h *MLSuggestionsHandler) GenerateMLSuggestions(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	suggestions, err := h.mlEngine.GenerateMLSuggestions(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if suggestions == nil {
		suggestions = []*analysis.MLSuggestionResult{}
	}

	response.Success(w, suggestions)
}

// DismissMLSuggestion marks an ML suggestion as dismissed.
// PUT /ml-suggestions/{suggestionID}/dismiss
func (h *MLSuggestionsHandler) DismissMLSuggestion(w http.ResponseWriter, r *http.Request) {
	suggestionIDStr := chi.URLParam(r, "suggestionID")
	if suggestionIDStr == "" {
		response.BadRequest(w, errors.New("suggestion ID is required"))
		return
	}

	suggestionID, err := strconv.ParseInt(suggestionIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, errors.New("invalid suggestion ID"))
		return
	}

	if err := h.mlRepo.DismissSuggestion(r.Context(), suggestionID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// ApplyMLSuggestion marks an ML suggestion as applied.
// PUT /ml-suggestions/{suggestionID}/apply
func (h *MLSuggestionsHandler) ApplyMLSuggestion(w http.ResponseWriter, r *http.Request) {
	suggestionIDStr := chi.URLParam(r, "suggestionID")
	if suggestionIDStr == "" {
		response.BadRequest(w, errors.New("suggestion ID is required"))
		return
	}

	suggestionID, err := strconv.ParseInt(suggestionIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, errors.New("invalid suggestion ID"))
		return
	}

	if err := h.mlRepo.ApplySuggestion(r.Context(), suggestionID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// GetSynergyReport returns a synergy analysis report for a deck.
// GET /decks/{deckID}/synergy-report
func (h *MLSuggestionsHandler) GetSynergyReport(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	report, err := h.mlEngine.GetSynergyReport(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, report)
}

// GetTopSynergies returns top synergistic cards for a given card.
// GET /cards/{cardID}/synergies
func (h *MLSuggestionsHandler) GetTopSynergies(w http.ResponseWriter, r *http.Request) {
	cardIDStr := chi.URLParam(r, "cardID")
	if cardIDStr == "" {
		response.BadRequest(w, errors.New("card ID is required"))
		return
	}

	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil {
		response.BadRequest(w, errors.New("invalid card ID"))
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "Standard"
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	synergies, err := h.mlRepo.GetTopSynergiesForCard(r.Context(), cardID, format, limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, synergies)
}

// ProcessMatchHistory triggers processing of match history to build synergy data.
// POST /ml/process-history
func (h *MLSuggestionsHandler) ProcessMatchHistory(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	lookbackDays := 90

	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 && parsed <= 365 {
			lookbackDays = parsed
		}
	}

	if err := h.mlEngine.ProcessMatchHistory(r.Context(), format, lookbackDays); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":  "success",
		"message": "Match history processed successfully",
	})
}

// GetUserPlayPatterns returns the user's play pattern profile.
// GET /ml/play-patterns
func (h *MLSuggestionsHandler) GetUserPlayPatterns(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		accountID = "default" // Use default if not specified
	}

	patterns, err := h.mlRepo.GetUserPlayPatterns(r.Context(), accountID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if patterns == nil {
		response.NotFound(w, errors.New("no play patterns found"))
		return
	}

	response.Success(w, patterns)
}

// UpdateUserPlayPatterns triggers an update of the user's play pattern profile.
// POST /ml/play-patterns/update
func (h *MLSuggestionsHandler) UpdateUserPlayPatterns(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		accountID = "default"
	}

	if err := h.mlEngine.UpdateUserPlayPatterns(r.Context(), accountID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":  "success",
		"message": "Play patterns updated successfully",
	})
}

// GetCombinationStats returns synergy statistics for a card pair.
// GET /ml/combinations
func (h *MLSuggestionsHandler) GetCombinationStats(w http.ResponseWriter, r *http.Request) {
	card1Str := r.URL.Query().Get("card1")
	card2Str := r.URL.Query().Get("card2")
	format := r.URL.Query().Get("format")

	if card1Str == "" || card2Str == "" {
		response.BadRequest(w, errors.New("card1 and card2 are required"))
		return
	}

	card1, err := strconv.Atoi(card1Str)
	if err != nil {
		response.BadRequest(w, errors.New("invalid card1 ID"))
		return
	}

	card2, err := strconv.Atoi(card2Str)
	if err != nil {
		response.BadRequest(w, errors.New("invalid card2 ID"))
		return
	}

	if format == "" {
		format = "Standard"
	}

	stats, err := h.mlRepo.GetCombinationStats(r.Context(), card1, card2, format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if stats == nil {
		response.NotFound(w, errors.New("no combination stats found"))
		return
	}

	response.Success(w, stats)
}

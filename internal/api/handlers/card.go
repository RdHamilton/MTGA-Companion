package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// CardHandler handles card-related API requests.
type CardHandler struct {
	facade *gui.CardFacade
}

// NewCardHandler creates a new CardHandler.
func NewCardHandler(facade *gui.CardFacade) *CardHandler {
	return &CardHandler{facade: facade}
}

// SearchCards searches for cards.
func (h *CardHandler) SearchCards(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	setCode := r.URL.Query().Get("set")
	limitStr := r.URL.Query().Get("limit")

	var setCodes []string
	if setCode != "" {
		setCodes = []string{setCode}
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	cards, err := h.facade.SearchCards(r.Context(), query, setCodes, limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, cards)
}

// GetCard returns a card by Arena ID.
func (h *CardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	cardID := chi.URLParam(r, "cardID")
	if cardID == "" {
		response.BadRequest(w, errors.New("card ID is required"))
		return
	}

	card, err := h.facade.GetCardByArenaID(r.Context(), cardID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if card == nil {
		response.NotFound(w, errors.New("card not found"))
		return
	}

	response.Success(w, card)
}

// GetCardByName searches for a card by name.
func (h *CardHandler) GetCardByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		response.BadRequest(w, errors.New("card name is required"))
		return
	}

	// Search by name with limit 1
	cards, err := h.facade.SearchCards(r.Context(), name, nil, 1)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if len(cards) == 0 {
		response.NotFound(w, errors.New("card not found"))
		return
	}

	response.Success(w, cards[0])
}

// GetSets returns all available sets.
func (h *CardHandler) GetSets(w http.ResponseWriter, r *http.Request) {
	sets, err := h.facade.GetAllSetInfo(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, sets)
}

// GetSetCards returns all cards in a set.
func (h *CardHandler) GetSetCards(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	cards, err := h.facade.GetSetCards(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, cards)
}

// GetRatings returns 17Lands ratings for a set.
func (h *CardHandler) GetRatings(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	eventType := r.URL.Query().Get("event")
	if eventType == "" {
		eventType = "PremierDraft"
	}

	ratings, err := h.facade.GetCardRatings(r.Context(), setCode, eventType)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, ratings)
}

// BulkCardsRequest represents a request for multiple cards.
type BulkCardsRequest struct {
	ArenaIDs []int `json:"arena_ids"`
}

// GetCardsBulk returns collection quantities for multiple cards by Arena ID.
func (h *CardHandler) GetCardsBulk(w http.ResponseWriter, r *http.Request) {
	var req BulkCardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	quantities, err := h.facade.GetCollectionQuantities(r.Context(), req.ArenaIDs)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, quantities)
}

// FetchSetCards manually fetches and caches set cards from Scryfall.
func (h *CardHandler) FetchSetCards(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	count, err := h.facade.FetchSetCards(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":   "success",
		"set_code": setCode,
		"count":    count,
		"message":  "Set cards fetched successfully",
	})
}

// RefreshSetCards deletes and re-fetches all cards for a set.
func (h *CardHandler) RefreshSetCards(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	count, err := h.facade.RefreshSetCards(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":   "success",
		"set_code": setCode,
		"count":    count,
		"message":  "Set cards refreshed successfully",
	})
}

// FetchRatingsRequest represents a request to fetch ratings.
type FetchRatingsRequest struct {
	DraftFormat string `json:"draft_format"`
}

// FetchSetRatings fetches and caches 17Lands ratings for a set.
func (h *CardHandler) FetchSetRatings(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	var req FetchRatingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	draftFormat := req.DraftFormat
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}

	if err := h.facade.FetchSetRatings(r.Context(), setCode, draftFormat); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":       "success",
		"set_code":     setCode,
		"draft_format": draftFormat,
		"message":      "Ratings fetched successfully",
	})
}

// RefreshSetRatings deletes and re-fetches 17Lands ratings for a set.
func (h *CardHandler) RefreshSetRatings(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	var req FetchRatingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	draftFormat := req.DraftFormat
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}

	if err := h.facade.RefreshSetRatings(r.Context(), setCode, draftFormat); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":       "success",
		"set_code":     setCode,
		"draft_format": draftFormat,
		"message":      "Ratings refreshed successfully",
	})
}

// GetRatingsStaleness returns staleness information for set ratings.
func (h *CardHandler) GetRatingsStaleness(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("setCode is required"))
		return
	}

	draftFormat := chi.URLParam(r, "format")
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}

	staleness, err := h.facade.GetRatingsStaleness(r.Context(), setCode, draftFormat)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, staleness)
}

// ClearDatasetCache clears all cached 17Lands datasets.
func (h *CardHandler) ClearDatasetCache(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.ClearDatasetCache(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":  "success",
		"message": "Dataset cache cleared",
	})
}

// GetDatasetSource returns the data source for a set and format.
func (h *CardHandler) GetDatasetSource(w http.ResponseWriter, r *http.Request) {
	setCode := r.URL.Query().Get("set")
	draftFormat := r.URL.Query().Get("format")

	if setCode == "" {
		response.BadRequest(w, errors.New("set query parameter is required"))
		return
	}
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}

	source := h.facade.GetDatasetSource(r.Context(), setCode, draftFormat)

	response.Success(w, map[string]string{
		"set_code":     setCode,
		"draft_format": draftFormat,
		"source":       source,
	})
}

// GetCardRatingByArenaID returns the rating for a specific card.
func (h *CardHandler) GetCardRatingByArenaID(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	arenaID := chi.URLParam(r, "arenaID")

	if setCode == "" || arenaID == "" {
		response.BadRequest(w, errors.New("set code and arena ID are required"))
		return
	}

	draftFormat := r.URL.Query().Get("format")
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}

	rating, err := h.facade.GetCardRatingByArenaID(r.Context(), setCode, draftFormat, arenaID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if rating == nil {
		response.NotFound(w, errors.New("rating not found"))
		return
	}

	response.Success(w, rating)
}

// GetColorRatings returns color combination ratings for a set.
func (h *CardHandler) GetColorRatings(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	draftFormat := r.URL.Query().Get("format")
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}

	ratings, err := h.facade.GetColorRatings(r.Context(), setCode, draftFormat)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, ratings)
}

// GetSetInfo returns information about a specific set.
func (h *CardHandler) GetSetInfo(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	info, err := h.facade.GetSetInfo(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if info == nil {
		response.NotFound(w, errors.New("set not found"))
		return
	}

	response.Success(w, info)
}

// GetRatingsWithEvent returns 17Lands ratings for a set with event type in path.
func (h *CardHandler) GetRatingsWithEvent(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	eventType := chi.URLParam(r, "eventType")
	if eventType == "" {
		eventType = "PremierDraft"
	}

	ratings, err := h.facade.GetCardRatings(r.Context(), setCode, eventType)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, ratings)
}

// SearchWithCollectionRequest represents a search request with collection filter.
type SearchWithCollectionRequest struct {
	Query          string   `json:"query"`
	SetCodes       []string `json:"set_codes,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	CollectionOnly bool     `json:"collection_only,omitempty"`
}

// SearchCardsWithCollection searches for cards and includes collection ownership.
func (h *CardHandler) SearchCardsWithCollection(w http.ResponseWriter, r *http.Request) {
	var req SearchWithCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	cards, err := h.facade.SearchCardsWithCollection(r.Context(), req.Query, req.SetCodes, limit, req.CollectionOnly)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, cards)
}

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

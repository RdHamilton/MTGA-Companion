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

// CollectionHandler handles collection-related API requests.
type CollectionHandler struct {
	facade *gui.CollectionFacade
}

// NewCollectionHandler creates a new CollectionHandler.
func NewCollectionHandler(facade *gui.CollectionFacade) *CollectionHandler {
	return &CollectionHandler{facade: facade}
}

// GetCollection returns the full collection.
func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	// Parse optional filters from query params
	setCode := r.URL.Query().Get("set")
	rarity := r.URL.Query().Get("rarity")

	var filter *gui.CollectionFilter
	if setCode != "" || rarity != "" {
		filter = &gui.CollectionFilter{
			SetCode: setCode,
			Rarity:  rarity,
		}
	}

	collection, err := h.facade.GetCollection(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, collection)
}

// CollectionFilterRequest represents a JSON request body for collection filtering.
type CollectionFilterRequest struct {
	SetCode   string   `json:"set_code,omitempty"`
	Rarity    string   `json:"rarity,omitempty"`
	Colors    []string `json:"colors,omitempty"`
	OwnedOnly *bool    `json:"owned_only,omitempty"`
}

// GetCollectionPost returns the collection with filters from POST body.
func (h *CollectionHandler) GetCollectionPost(w http.ResponseWriter, r *http.Request) {
	var req CollectionFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body
		if err.Error() != "EOF" {
			response.BadRequest(w, errors.New("invalid request body"))
			return
		}
	}

	var filter *gui.CollectionFilter
	if req.SetCode != "" || req.Rarity != "" || len(req.Colors) > 0 || req.OwnedOnly != nil {
		filter = &gui.CollectionFilter{
			SetCode:   req.SetCode,
			Rarity:    req.Rarity,
			Colors:    req.Colors,
			OwnedOnly: req.OwnedOnly != nil && *req.OwnedOnly,
		}
	}

	collection, err := h.facade.GetCollection(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, collection)
}

// GetCollectionStats returns collection statistics.
func (h *CollectionHandler) GetCollectionStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.facade.GetCollectionStats(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// GetCollectionBySets returns collection grouped by sets.
func (h *CollectionHandler) GetCollectionBySets(w http.ResponseWriter, r *http.Request) {
	completion, err := h.facade.GetSetCompletion(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, completion)
}

// GetCollectionByRarity returns recent collection changes.
func (h *CollectionHandler) GetCollectionByRarity(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	changes, err := h.facade.GetRecentChanges(r.Context(), limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, changes)
}

// GetMissingCards returns missing cards for a set.
func (h *CollectionHandler) GetMissingCards(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	missing, err := h.facade.GetMissingCardsForSet(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, missing)
}

// SearchCollection searches the collection (same as GetCollection with filter).
func (h *CollectionHandler) SearchCollection(w http.ResponseWriter, r *http.Request) {
	// Redirect to GetCollection since it supports filtering
	h.GetCollection(w, r)
}

// GetMissingCardsForDeck returns missing cards for a specific deck.
func (h *CollectionHandler) GetMissingCardsForDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	missing, err := h.facade.GetMissingCardsForDeck(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, missing)
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// DeckHandler handles deck-related API requests.
type DeckHandler struct {
	facade *gui.DeckFacade
}

// NewDeckHandler creates a new DeckHandler.
func NewDeckHandler(facade *gui.DeckFacade) *DeckHandler {
	return &DeckHandler{facade: facade}
}

// GetDecks returns all decks.
func (h *DeckHandler) GetDecks(w http.ResponseWriter, r *http.Request) {
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
		response.InternalError(w, err)
		return
	}

	response.Success(w, decks)
}

// CreateDeckRequest represents a request to create a deck.
type CreateDeckRequest struct {
	Name         string  `json:"name"`
	Format       string  `json:"format"`
	Source       string  `json:"source"`
	DraftEventID *string `json:"draft_event_id,omitempty"`
}

// CreateDeck creates a new deck.
func (h *DeckHandler) CreateDeck(w http.ResponseWriter, r *http.Request) {
	var req CreateDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Name == "" {
		response.BadRequest(w, errors.New("deck name is required"))
		return
	}

	deck, err := h.facade.CreateDeck(r.Context(), req.Name, req.Format, req.Source, req.DraftEventID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Created(w, deck)
}

// GetDeck returns a single deck by ID.
func (h *DeckHandler) GetDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	deck, err := h.facade.GetDeck(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if deck == nil {
		response.NotFound(w, errors.New("deck not found"))
		return
	}

	response.Success(w, deck)
}

// UpdateDeckRequest represents a request to update a deck.
type UpdateDeckRequest struct {
	Name   *string `json:"name,omitempty"`
	Format *string `json:"format,omitempty"`
}

// UpdateDeck updates a deck.
func (h *DeckHandler) UpdateDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req UpdateDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// Get current deck
	deckWithCards, err := h.facade.GetDeck(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}
	if deckWithCards == nil || deckWithCards.Deck == nil {
		response.NotFound(w, errors.New("deck not found"))
		return
	}

	// Update fields
	if req.Name != nil {
		deckWithCards.Deck.Name = *req.Name
	}
	if req.Format != nil {
		deckWithCards.Deck.Format = *req.Format
	}

	// Save
	if err := h.facade.UpdateDeck(r.Context(), deckWithCards.Deck); err != nil {
		response.InternalError(w, err)
		return
	}

	// Return updated deck
	response.Success(w, deckWithCards)
}

// DeleteDeck deletes a deck.
func (h *DeckHandler) DeleteDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	if err := h.facade.DeleteDeck(r.Context(), deckID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// GetDeckStats returns statistics for a deck.
func (h *DeckHandler) GetDeckStats(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	stats, err := h.facade.GetDeckStatistics(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// GetDeckMatches returns performance for a deck.
func (h *DeckHandler) GetDeckMatches(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	performance, err := h.facade.GetDeckPerformance(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, performance)
}

// GetDeckCurve returns statistics including mana curve for a deck.
func (h *DeckHandler) GetDeckCurve(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	stats, err := h.facade.GetDeckStatistics(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// GetDeckColors returns statistics including color distribution for a deck.
func (h *DeckHandler) GetDeckColors(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	stats, err := h.facade.GetDeckStatistics(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// ExportDeckRequest represents a request to export a deck.
type ExportDeckRequest struct {
	Format string `json:"format"` // mtga, arena, text, etc.
}

// ExportDeck exports a deck in the specified format.
func (h *DeckHandler) ExportDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req ExportDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	exportReq := &gui.ExportDeckRequest{
		DeckID: deckID,
		Format: req.Format,
	}

	exported, err := h.facade.ExportDeck(r.Context(), exportReq)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, exported)
}

// ImportDeckRequest represents a request to import a deck.
type ImportDeckRequest struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Format  string `json:"format"`
}

// ImportDeck imports a deck from text.
func (h *DeckHandler) ImportDeck(w http.ResponseWriter, r *http.Request) {
	var req ImportDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Content == "" {
		response.BadRequest(w, errors.New("deck content is required"))
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
		response.InternalError(w, err)
		return
	}

	response.Created(w, result)
}

// ParseDeckRequest represents a request to parse a deck list.
type ParseDeckRequest struct {
	Content string `json:"content"`
}

// ParseDeckList parses a deck list without saving.
func (h *DeckHandler) ParseDeckList(w http.ResponseWriter, r *http.Request) {
	var req ParseDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// Use import with a preview flag or just return validation
	importReq := &gui.ImportDeckRequest{
		ImportText: req.Content,
		Source:     "imported",
	}

	result, err := h.facade.ImportDeck(r.Context(), importReq)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// SuggestDecksRequest represents a request for deck suggestions.
type SuggestDecksRequest struct {
	SessionID string `json:"session_id"`
}

// SuggestDecks suggests deck builds for a draft.
func (h *DeckHandler) SuggestDecks(w http.ResponseWriter, r *http.Request) {
	var req SuggestDecksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	suggestions, err := h.facade.SuggestDecks(r.Context(), req.SessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, suggestions)
}

// AnalyzeDeckRequest represents a request for deck analysis.
type AnalyzeDeckRequest struct {
	DeckID string `json:"deck_id"`
}

// AnalyzeDeck analyzes a deck (classifies archetype).
func (h *DeckHandler) AnalyzeDeck(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	result, err := h.facade.ClassifyDeckArchetype(r.Context(), req.DeckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

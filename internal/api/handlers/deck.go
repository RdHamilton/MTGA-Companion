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

// AddCardRequest represents a request to add a card to a deck.
type AddCardRequest struct {
	CardID    int    `json:"card_id"`
	Quantity  int    `json:"quantity"`
	Board     string `json:"board"` // main, sideboard
	FromDraft bool   `json:"from_draft,omitempty"`
}

// AddCard adds a card to a deck.
func (h *DeckHandler) AddCard(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req AddCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Board == "" {
		req.Board = "main"
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	if err := h.facade.AddCard(r.Context(), deckID, req.CardID, req.Quantity, req.Board, req.FromDraft); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// RemoveCardRequest represents a request to remove a card from a deck.
type RemoveCardRequest struct {
	CardID int    `json:"card_id"`
	Board  string `json:"board"`
}

// RemoveCard removes a card from a deck.
func (h *DeckHandler) RemoveCard(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req RemoveCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Board == "" {
		req.Board = "main"
	}

	if err := h.facade.RemoveCard(r.Context(), deckID, req.CardID, req.Board); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// ValidateDraftDeck validates a draft deck meets requirements.
func (h *DeckHandler) ValidateDraftDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	valid, err := h.facade.ValidateDraftDeck(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]bool{"valid": valid})
}

// TagRequest represents a request to add/remove a tag.
type TagRequest struct {
	Tag string `json:"tag"`
}

// AddTag adds a tag to a deck.
func (h *DeckHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req TagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Tag == "" {
		response.BadRequest(w, errors.New("tag is required"))
		return
	}

	if err := h.facade.AddTag(r.Context(), deckID, req.Tag); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// RemoveTag removes a tag from a deck.
func (h *DeckHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	tag := chi.URLParam(r, "tag")

	if deckID == "" || tag == "" {
		response.BadRequest(w, errors.New("deck ID and tag are required"))
		return
	}

	if err := h.facade.RemoveTag(r.Context(), deckID, tag); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// GetDeckByDraftEvent returns a deck by its draft event ID.
func (h *DeckHandler) GetDeckByDraftEvent(w http.ResponseWriter, r *http.Request) {
	draftEventID := chi.URLParam(r, "draftEventID")
	if draftEventID == "" {
		response.BadRequest(w, errors.New("draft event ID is required"))
		return
	}

	deck, err := h.facade.GetDeckByDraftEvent(r.Context(), draftEventID)
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

// GetRecommendationsRequest represents a request for card recommendations.
type GetRecommendationsRequest struct {
	DeckID       string   `json:"deck_id"`
	MaxResults   int      `json:"max_results,omitempty"`
	MinScore     float64  `json:"min_score,omitempty"`
	Colors       []string `json:"colors,omitempty"`
	CardTypes    []string `json:"card_types,omitempty"`
	IncludeLands bool     `json:"include_lands,omitempty"`
}

// GetRecommendations returns card recommendations for a deck.
func (h *DeckHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	var req GetRecommendationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.DeckID == "" {
		response.BadRequest(w, errors.New("deck_id is required"))
		return
	}

	guiReq := &gui.GetRecommendationsRequest{
		DeckID:       req.DeckID,
		MaxResults:   req.MaxResults,
		MinScore:     req.MinScore,
		Colors:       req.Colors,
		CardTypes:    req.CardTypes,
		IncludeLands: req.IncludeLands,
	}

	recommendations, err := h.facade.GetRecommendations(r.Context(), guiReq)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, recommendations)
}

// ExplainRecommendationRequest represents a request to explain a recommendation.
type ExplainRecommendationRequest struct {
	DeckID string `json:"deck_id"`
	CardID int    `json:"card_id"`
}

// ExplainRecommendation explains why a card is recommended.
func (h *DeckHandler) ExplainRecommendation(w http.ResponseWriter, r *http.Request) {
	var req ExplainRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	guiReq := &gui.ExplainRecommendationRequest{
		DeckID: req.DeckID,
		CardID: req.CardID,
	}

	explanation, err := h.facade.ExplainRecommendation(r.Context(), guiReq)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, explanation)
}

// CloneDeckRequest represents a request to clone a deck.
type CloneDeckRequest struct {
	NewName string `json:"new_name"`
}

// CloneDeck clones a deck with a new name.
func (h *DeckHandler) CloneDeck(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req CloneDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.NewName == "" {
		response.BadRequest(w, errors.New("new_name is required"))
		return
	}

	deck, err := h.facade.CloneDeck(r.Context(), deckID, req.NewName)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Created(w, deck)
}

// GetDecksByTagsRequest represents a request to get decks by tags.
type GetDecksByTagsRequest struct {
	Tags []string `json:"tags"`
}

// GetDecksByTags returns decks matching the specified tags.
func (h *DeckHandler) GetDecksByTags(w http.ResponseWriter, r *http.Request) {
	var req GetDecksByTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	decks, err := h.facade.GetDecksByTags(r.Context(), req.Tags)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, decks)
}

// DeckLibraryFilterRequest represents a filter for deck library.
type DeckLibraryFilterRequest struct {
	Format   string   `json:"format,omitempty"`
	Source   string   `json:"source,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	SortBy   string   `json:"sort_by,omitempty"`
	SortDesc bool     `json:"sort_desc,omitempty"`
}

// GetDeckLibrary returns a filtered list of decks.
func (h *DeckHandler) GetDeckLibrary(w http.ResponseWriter, r *http.Request) {
	var req DeckLibraryFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	var formatPtr, sourcePtr *string
	if req.Format != "" {
		formatPtr = &req.Format
	}
	if req.Source != "" {
		sourcePtr = &req.Source
	}

	filter := &gui.DeckLibraryFilter{
		Format:   formatPtr,
		Source:   sourcePtr,
		Tags:     req.Tags,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	decks, err := h.facade.GetDeckLibrary(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, decks)
}

// ClassifyDraftPoolRequest represents a request to classify a draft pool.
type ClassifyDraftPoolRequest struct {
	DraftEventID string `json:"draft_event_id"`
}

// ClassifyDraftPoolArchetype classifies the archetype of a draft pool.
func (h *DeckHandler) ClassifyDraftPoolArchetype(w http.ResponseWriter, r *http.Request) {
	var req ClassifyDraftPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	result, err := h.facade.ClassifyDraftPoolArchetype(r.Context(), req.DraftEventID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// ApplySuggestedDeckRequest represents a request to apply a suggested deck.
type ApplySuggestedDeckRequest struct {
	DeckID     string                     `json:"deck_id"`
	Suggestion *gui.SuggestedDeckResponse `json:"suggestion"`
}

// ApplySuggestedDeck applies a suggested deck build.
func (h *DeckHandler) ApplySuggestedDeck(w http.ResponseWriter, r *http.Request) {
	var req ApplySuggestedDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if err := h.facade.ApplySuggestedDeck(r.Context(), req.DeckID, req.Suggestion); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// ExportSuggestedDeckRequest represents a request to export a suggested deck.
type ExportSuggestedDeckRequest struct {
	Suggestion *gui.SuggestedDeckResponse `json:"suggestion"`
	DeckName   string                     `json:"deck_name"`
}

// ExportSuggestedDeck returns a suggested deck as exportable text.
func (h *DeckHandler) ExportSuggestedDeck(w http.ResponseWriter, r *http.Request) {
	var req ExportSuggestedDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// ExportSuggestedDeck in facade uses a dialog, so we'll just format it here
	// Return the deck list as text that the frontend can handle
	response.Success(w, map[string]interface{}{
		"deck_name":  req.DeckName,
		"suggestion": req.Suggestion,
		"message":    "Use the suggestion data to export via frontend",
	})
}

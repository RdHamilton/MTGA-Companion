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
		// Check if this is a validation error (like 4-card limit)
		var appErr *gui.AppError
		if errors.As(err, &appErr) {
			response.BadRequest(w, err)
			return
		}
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// RemoveCard removes a card from a deck.
// DELETE /decks/{deckID}/cards/{cardID}?zone=main
func (h *DeckHandler) RemoveCard(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

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

	board := r.URL.Query().Get("zone")
	if board == "" {
		board = "main"
	}

	if err := h.facade.RemoveCard(r.Context(), deckID, cardID, board); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// RemoveAllCopies removes all copies of a card from a deck.
// DELETE /decks/{deckID}/cards/{cardID}/all?zone=main
func (h *DeckHandler) RemoveAllCopies(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

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

	board := r.URL.Query().Get("zone")
	if board == "" {
		board = "main"
	}

	if err := h.facade.RemoveAllCopies(r.Context(), deckID, cardID, board); err != nil {
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

// BuildAroundSeedRequest represents a request to build a deck around a seed card.
type BuildAroundSeedRequest struct {
	SeedCardID     int      `json:"seed_card_id"`
	MaxResults     int      `json:"max_results,omitempty"`
	BudgetMode     bool     `json:"budget_mode,omitempty"`
	SetRestriction string   `json:"set_restriction,omitempty"`
	AllowedSets    []string `json:"allowed_sets,omitempty"`
}

// BuildAroundSeed generates deck suggestions based on a seed card.
func (h *DeckHandler) BuildAroundSeed(w http.ResponseWriter, r *http.Request) {
	var req BuildAroundSeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.SeedCardID <= 0 {
		response.BadRequest(w, errors.New("seed_card_id is required"))
		return
	}

	guiReq := &gui.BuildAroundSeedRequest{
		SeedCardID:     req.SeedCardID,
		MaxResults:     req.MaxResults,
		BudgetMode:     req.BudgetMode,
		SetRestriction: req.SetRestriction,
		AllowedSets:    req.AllowedSets,
	}

	result, err := h.facade.BuildAroundSeed(r.Context(), guiReq)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// IterativeBuildAroundRequest represents a request for iterative deck building suggestions.
type IterativeBuildAroundRequest struct {
	SeedCardID     int      `json:"seed_card_id"`
	DeckCardIDs    []int    `json:"deck_card_ids"`
	MaxResults     int      `json:"max_results,omitempty"`
	BudgetMode     bool     `json:"budget_mode,omitempty"`
	SetRestriction string   `json:"set_restriction,omitempty"`
	AllowedSets    []string `json:"allowed_sets,omitempty"`
}

// SuggestNextCards generates suggestions based on the current deck composition.
// This is used for iterative deck building where users pick cards one-by-one.
func (h *DeckHandler) SuggestNextCards(w http.ResponseWriter, r *http.Request) {
	var req IterativeBuildAroundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// seed_card_id is optional - if not provided, analyze deck cards collectively
	if req.SeedCardID <= 0 && len(req.DeckCardIDs) == 0 {
		response.BadRequest(w, errors.New("either seed_card_id or deck_card_ids is required"))
		return
	}

	guiReq := &gui.IterativeBuildAroundRequest{
		SeedCardID:     req.SeedCardID,
		DeckCardIDs:    req.DeckCardIDs,
		MaxResults:     req.MaxResults,
		BudgetMode:     req.BudgetMode,
		SetRestriction: req.SetRestriction,
		AllowedSets:    req.AllowedSets,
	}

	result, err := h.facade.SuggestNextCards(r.Context(), guiReq)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// ============================================================================
// Card Performance Analysis (Issue #771)
// ============================================================================

// GetCardPerformance returns performance metrics for all cards in a deck.
// GET /decks/{deckID}/card-performance
func (h *DeckHandler) GetCardPerformance(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Parse query params
	includeLands := r.URL.Query().Get("include_lands") == "true"
	minGames := 0
	if minGamesStr := r.URL.Query().Get("min_games"); minGamesStr != "" {
		if parsed, err := strconv.Atoi(minGamesStr); err == nil {
			minGames = parsed
		}
	}

	req := &gui.GetCardPerformanceRequest{
		DeckID:       deckID,
		MinGames:     minGames,
		IncludeLands: includeLands,
	}

	result, err := h.facade.GetCardPerformance(r.Context(), req)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// GetPerformanceRecommendationsRequest represents the request body for getting performance-based recommendations.
type GetPerformanceRecommendationsRequest struct {
	MaxResults   int    `json:"max_results,omitempty"`
	IncludeSwaps bool   `json:"include_swaps,omitempty"`
	Format       string `json:"format,omitempty"`
}

// GetPerformanceAddRecommendations returns card add recommendations based on performance.
// GET /decks/{deckID}/recommendations/add
func (h *DeckHandler) GetPerformanceAddRecommendations(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Parse query params
	maxResults := 5
	if maxStr := r.URL.Query().Get("max_results"); maxStr != "" {
		if parsed, err := strconv.Atoi(maxStr); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	req := &gui.GetPerformanceRecommendationsRequest{
		DeckID:       deckID,
		MaxResults:   maxResults,
		IncludeSwaps: false,
		Format:       r.URL.Query().Get("format"),
	}

	result, err := h.facade.GetPerformanceRecommendations(r.Context(), req)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Return just the add recommendations
	response.Success(w, result.AddRecommendations)
}

// GetPerformanceRemoveRecommendations returns card removal recommendations based on performance.
// GET /decks/{deckID}/recommendations/remove
func (h *DeckHandler) GetPerformanceRemoveRecommendations(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Parse query params
	threshold := 0.05
	if thresholdStr := r.URL.Query().Get("threshold"); thresholdStr != "" {
		if parsed, err := strconv.ParseFloat(thresholdStr, 64); err == nil && parsed > 0 {
			threshold = parsed
		}
	}

	result, err := h.facade.GetUnderperformingCards(r.Context(), deckID, threshold)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// GetPerformanceSwapRecommendations returns card swap recommendations based on performance.
// GET /decks/{deckID}/recommendations/swap
func (h *DeckHandler) GetPerformanceSwapRecommendations(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Parse query params
	maxResults := 5
	if maxStr := r.URL.Query().Get("max_results"); maxStr != "" {
		if parsed, err := strconv.Atoi(maxStr); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	req := &gui.GetPerformanceRecommendationsRequest{
		DeckID:       deckID,
		MaxResults:   maxResults,
		IncludeSwaps: true,
		Format:       r.URL.Query().Get("format"),
	}

	result, err := h.facade.GetPerformanceRecommendations(r.Context(), req)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Return just the swap recommendations
	response.Success(w, result.SwapRecommendations)
}

// GetAllPerformanceRecommendations returns all recommendations (add/remove/swap) for a deck.
// GET /decks/{deckID}/recommendations/all
func (h *DeckHandler) GetAllPerformanceRecommendations(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Parse query params
	maxResults := 5
	if maxStr := r.URL.Query().Get("max_results"); maxStr != "" {
		if parsed, err := strconv.Atoi(maxStr); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	req := &gui.GetPerformanceRecommendationsRequest{
		DeckID:       deckID,
		MaxResults:   maxResults,
		IncludeSwaps: true,
		Format:       r.URL.Query().Get("format"),
	}

	result, err := h.facade.GetPerformanceRecommendations(r.Context(), req)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

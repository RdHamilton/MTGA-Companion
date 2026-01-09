package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/analysis"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// NotesHandler handles notes and suggestions API requests.
type NotesHandler struct {
	notesRepo     repository.NotesRepository
	suggRepo      repository.SuggestionRepository
	suggGenerator *analysis.SuggestionGenerator
}

// NewNotesHandler creates a new NotesHandler.
func NewNotesHandler(
	notesRepo repository.NotesRepository,
	suggRepo repository.SuggestionRepository,
	suggGenerator *analysis.SuggestionGenerator,
) *NotesHandler {
	return &NotesHandler{
		notesRepo:     notesRepo,
		suggRepo:      suggRepo,
		suggGenerator: suggGenerator,
	}
}

// CreateDeckNoteRequest represents a request to create a deck note.
type CreateDeckNoteRequest struct {
	Content  string `json:"content"`
	Category string `json:"category,omitempty"`
}

// UpdateDeckNoteRequest represents a request to update a deck note.
type UpdateDeckNoteRequest struct {
	Content  string `json:"content"`
	Category string `json:"category,omitempty"`
}

// UpdateMatchNotesRequest represents a request to update match notes.
type UpdateMatchNotesRequest struct {
	Notes  string `json:"notes"`
	Rating int    `json:"rating"`
}

// GetDeckNotes returns all notes for a deck.
func (h *NotesHandler) GetDeckNotes(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	category := r.URL.Query().Get("category")

	var notes []*models.DeckNote
	var err error

	if category != "" {
		notes, err = h.notesRepo.GetDeckNotesByCategory(r.Context(), deckID, category)
	} else {
		notes, err = h.notesRepo.GetDeckNotes(r.Context(), deckID)
	}

	if err != nil {
		response.InternalError(w, err)
		return
	}

	if notes == nil {
		notes = []*models.DeckNote{}
	}

	response.Success(w, notes)
}

// CreateDeckNote creates a new note for a deck.
func (h *NotesHandler) CreateDeckNote(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	var req CreateDeckNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Content == "" {
		response.BadRequest(w, errors.New("note content is required"))
		return
	}

	// Default category to "general" if not provided
	category := req.Category
	if category == "" {
		category = models.NoteCategoryGeneral
	}

	note := &models.DeckNote{
		DeckID:   deckID,
		Content:  req.Content,
		Category: category,
	}

	if err := h.notesRepo.CreateDeckNote(r.Context(), note); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Created(w, note)
}

// GetDeckNote returns a single deck note by ID.
func (h *NotesHandler) GetDeckNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := chi.URLParam(r, "noteID")
	if noteIDStr == "" {
		response.BadRequest(w, errors.New("note ID is required"))
		return
	}

	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, errors.New("invalid note ID"))
		return
	}

	note, err := h.notesRepo.GetDeckNoteByID(r.Context(), noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, errors.New("note not found"))
			return
		}
		response.InternalError(w, err)
		return
	}

	if note == nil {
		response.NotFound(w, errors.New("note not found"))
		return
	}

	response.Success(w, note)
}

// UpdateDeckNote updates an existing deck note.
func (h *NotesHandler) UpdateDeckNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := chi.URLParam(r, "noteID")
	if noteIDStr == "" {
		response.BadRequest(w, errors.New("note ID is required"))
		return
	}

	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, errors.New("invalid note ID"))
		return
	}

	var req UpdateDeckNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Content == "" {
		response.BadRequest(w, errors.New("note content is required"))
		return
	}

	// Get existing note to preserve deck ID
	existingNote, err := h.notesRepo.GetDeckNoteByID(r.Context(), noteID)
	if err != nil {
		response.InternalError(w, err)
		return
	}
	if existingNote == nil {
		response.NotFound(w, errors.New("note not found"))
		return
	}

	// Update note
	existingNote.Content = req.Content
	if req.Category != "" {
		existingNote.Category = req.Category
	}

	if err := h.notesRepo.UpdateDeckNote(r.Context(), existingNote); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, existingNote)
}

// DeleteDeckNote deletes a deck note.
func (h *NotesHandler) DeleteDeckNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := chi.URLParam(r, "noteID")
	if noteIDStr == "" {
		response.BadRequest(w, errors.New("note ID is required"))
		return
	}

	noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, errors.New("invalid note ID"))
		return
	}

	if err := h.notesRepo.DeleteDeckNote(r.Context(), noteID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

// GetMatchNotes returns the notes and rating for a match.
func (h *NotesHandler) GetMatchNotes(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	notes, err := h.notesRepo.GetMatchNotes(r.Context(), matchID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, notes)
}

// UpdateMatchNotes updates the notes and rating for a match.
func (h *NotesHandler) UpdateMatchNotes(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	var req UpdateMatchNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// Validate rating is 0-5
	if req.Rating < 0 || req.Rating > 5 {
		response.BadRequest(w, errors.New("rating must be between 0 and 5"))
		return
	}

	if err := h.notesRepo.UpdateMatchNotes(r.Context(), matchID, req.Notes, req.Rating); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, &models.MatchNotes{
		MatchID: matchID,
		Notes:   req.Notes,
		Rating:  req.Rating,
	})
}

// GetDeckSuggestions returns improvement suggestions for a deck.
func (h *NotesHandler) GetDeckSuggestions(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	activeOnly := r.URL.Query().Get("active") != "false"

	suggestions, err := h.suggGenerator.GetDeckSuggestions(r.Context(), deckID, activeOnly)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if suggestions == nil {
		suggestions = []*models.ImprovementSuggestion{}
	}

	response.Success(w, suggestions)
}

// GenerateSuggestions generates new improvement suggestions for a deck.
func (h *NotesHandler) GenerateSuggestions(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Get optional minGames parameter (default: 5)
	minGames := 5
	if minGamesStr := r.URL.Query().Get("min_games"); minGamesStr != "" {
		if parsed, err := strconv.Atoi(minGamesStr); err == nil && parsed > 0 {
			minGames = parsed
		}
	}

	suggestions, err := h.suggGenerator.GenerateSuggestions(r.Context(), deckID, minGames)
	if err != nil {
		// Check for insufficient games error
		if errors.Is(err, analysis.ErrInsufficientGames) {
			response.BadRequest(w, err)
			return
		}
		response.InternalError(w, err)
		return
	}

	response.Success(w, suggestions)
}

// DismissSuggestion marks a suggestion as dismissed.
func (h *NotesHandler) DismissSuggestion(w http.ResponseWriter, r *http.Request) {
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

	if err := h.suggRepo.DismissSuggestion(r.Context(), suggestionID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.NoContent(w)
}

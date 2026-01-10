package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// CFBHandler handles ChannelFireball ratings API requests.
type CFBHandler struct {
	facade *gui.CardFacade
}

// NewCFBHandler creates a new CFB ratings handler.
func NewCFBHandler(facade *gui.CardFacade) *CFBHandler {
	return &CFBHandler{facade: facade}
}

// GetCFBRatings returns all CFB ratings for a set.
func (h *CFBHandler) GetCFBRatings(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	ratings, err := h.facade.GetCFBRatings(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, ratings)
}

// GetCFBRatingByCard returns a CFB rating for a specific card.
func (h *CFBHandler) GetCFBRatingByCard(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	cardName := chi.URLParam(r, "cardName")

	if setCode == "" || cardName == "" {
		response.BadRequest(w, errors.New("set code and card name are required"))
		return
	}

	rating, err := h.facade.GetCFBRatingByCardName(r.Context(), cardName, setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if rating == nil {
		response.NotFound(w, errors.New("CFB rating not found"))
		return
	}

	response.Success(w, rating)
}

// CFBImportRequest represents a request to import CFB ratings.
type CFBImportRequest struct {
	Ratings []CFBRatingData `json:"ratings"`
}

// CFBRatingData represents a single CFB rating to import.
type CFBRatingData struct {
	CardName          string `json:"card_name"`
	SetCode           string `json:"set_code"`
	LimitedRating     string `json:"limited_rating"`
	ConstructedRating string `json:"constructed_rating,omitempty"`
	ArchetypeFit      string `json:"archetype_fit,omitempty"`
	Commentary        string `json:"commentary,omitempty"`
	SourceURL         string `json:"source_url,omitempty"`
	Author            string `json:"author,omitempty"`
}

// ImportCFBRatings imports CFB ratings from the request body.
func (h *CFBHandler) ImportCFBRatings(w http.ResponseWriter, r *http.Request) {
	var req CFBImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if len(req.Ratings) == 0 {
		response.BadRequest(w, errors.New("no ratings provided"))
		return
	}

	imported, err := h.facade.ImportCFBRatings(r.Context(), req.Ratings)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":   "success",
		"imported": imported,
		"message":  "CFB ratings imported successfully",
	})
}

// LinkCFBArenaIDs links CFB ratings to Arena IDs based on card name matching.
func (h *CFBHandler) LinkCFBArenaIDs(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	linked, err := h.facade.LinkCFBArenaIDs(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":   "success",
		"set_code": setCode,
		"linked":   linked,
		"message":  "CFB ratings linked to Arena IDs",
	})
}

// DeleteCFBRatings deletes all CFB ratings for a set.
func (h *CFBHandler) DeleteCFBRatings(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	if err := h.facade.DeleteCFBRatings(r.Context(), setCode); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status":   "success",
		"set_code": setCode,
		"message":  "CFB ratings deleted",
	})
}

// GetCFBRatingsCount returns the count of CFB ratings for a set.
func (h *CFBHandler) GetCFBRatingsCount(w http.ResponseWriter, r *http.Request) {
	setCode := chi.URLParam(r, "setCode")
	if setCode == "" {
		response.BadRequest(w, errors.New("set code is required"))
		return
	}

	count, err := h.facade.GetCFBRatingsCount(r.Context(), setCode)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"set_code": setCode,
		"count":    count,
	})
}

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// GamePlayHandler handles game play related API requests.
type GamePlayHandler struct {
	storage *storage.Service
}

// NewGamePlayHandler creates a new GamePlayHandler.
func NewGamePlayHandler(storage *storage.Service) *GamePlayHandler {
	return &GamePlayHandler{storage: storage}
}

// GetMatchPlays returns all plays for a specific match.
func (h *GamePlayHandler) GetMatchPlays(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	plays, err := h.storage.GamePlayRepo().GetPlaysByMatch(r.Context(), matchID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if plays == nil {
		plays = []*models.GamePlay{}
	}

	response.Success(w, plays)
}

// GetMatchTimeline returns plays organized by turn for a specific match.
func (h *GamePlayHandler) GetMatchTimeline(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	timeline, err := h.storage.GamePlayRepo().GetPlayTimeline(r.Context(), matchID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if timeline == nil {
		timeline = []*models.PlayTimelineEntry{}
	}

	response.Success(w, timeline)
}

// GetMatchOpponentCards returns cards observed from the opponent during a match.
func (h *GamePlayHandler) GetMatchOpponentCards(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	cards, err := h.storage.GamePlayRepo().GetOpponentCardsByMatch(r.Context(), matchID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if cards == nil {
		cards = []*models.OpponentCardObserved{}
	}

	response.Success(w, cards)
}

// GetMatchSnapshots returns game state snapshots for a specific match.
func (h *GamePlayHandler) GetMatchSnapshots(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	// Get gameID from query param if provided, otherwise get all for match
	gameIDStr := r.URL.Query().Get("gameID")
	var gameID int
	if gameIDStr != "" {
		var err error
		gameID, err = strconv.Atoi(gameIDStr)
		if err != nil {
			response.BadRequest(w, errors.New("invalid game ID"))
			return
		}
	}

	var snapshots []*models.GameStateSnapshot
	var err error

	if gameID > 0 {
		snapshots, err = h.storage.GamePlayRepo().GetSnapshotsByGame(r.Context(), gameID)
	} else {
		snapshots, err = h.storage.GamePlayRepo().GetSnapshotsByMatch(r.Context(), matchID)
	}

	if err != nil {
		response.InternalError(w, err)
		return
	}

	if snapshots == nil {
		snapshots = []*models.GameStateSnapshot{}
	}

	response.Success(w, snapshots)
}

// GetMatchPlaySummary returns a summary of plays for a match.
func (h *GamePlayHandler) GetMatchPlaySummary(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	summary, err := h.storage.GamePlayRepo().GetPlaySummary(r.Context(), matchID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if summary == nil {
		summary = &models.GamePlaySummary{
			MatchID:           matchID,
			TotalPlays:        0,
			CardPlays:         0,
			LandDrops:         0,
			Attacks:           0,
			Blocks:            0,
			TotalTurns:        0,
			OpponentCardsSeen: 0,
		}
	}

	response.Success(w, summary)
}

// GetPlaysByGame returns all plays for a specific game within a match.
func (h *GamePlayHandler) GetPlaysByGame(w http.ResponseWriter, r *http.Request) {
	gameIDStr := chi.URLParam(r, "gameID")
	if gameIDStr == "" {
		response.BadRequest(w, errors.New("game ID is required"))
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		response.BadRequest(w, errors.New("invalid game ID"))
		return
	}

	plays, err := h.storage.GamePlayRepo().GetPlaysByGame(r.Context(), gameID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if plays == nil {
		plays = []*models.GamePlay{}
	}

	response.Success(w, plays)
}

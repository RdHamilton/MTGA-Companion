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

// checkStorage validates the storage service is available.
func (h *GamePlayHandler) checkStorage(w http.ResponseWriter) bool {
	if h.storage == nil {
		response.ServiceUnavailable(w, errors.New("storage service unavailable"))
		return false
	}
	return true
}

// GetMatchPlays returns all plays for a specific match.
func (h *GamePlayHandler) GetMatchPlays(w http.ResponseWriter, r *http.Request) {
	if !h.checkStorage(w) {
		return
	}

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

// TimelineEntryResponse is the API response format for timeline entries.
// It splits plays by player type for easier frontend consumption.
type TimelineEntryResponse struct {
	Turn          int                       `json:"turn"`
	ActivePlayer  string                    `json:"active_player"`
	PlayerPlays   []*models.GamePlay        `json:"player_plays"`
	OpponentPlays []*models.GamePlay        `json:"opponent_plays"`
	Snapshot      *models.GameStateSnapshot `json:"snapshot,omitempty"`
}

// GetMatchTimeline returns plays organized by turn for a specific match.
func (h *GamePlayHandler) GetMatchTimeline(w http.ResponseWriter, r *http.Request) {
	if !h.checkStorage(w) {
		return
	}

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

	// Transform backend model to frontend-expected format
	result := make([]*TimelineEntryResponse, 0, len(timeline))
	for _, entry := range timeline {
		resp := &TimelineEntryResponse{
			Turn:          entry.Turn,
			ActivePlayer:  entry.Phase, // Use phase as active player indicator
			PlayerPlays:   make([]*models.GamePlay, 0),
			OpponentPlays: make([]*models.GamePlay, 0),
			Snapshot:      entry.Snapshot,
		}

		// Split plays by player type
		for _, play := range entry.Plays {
			if play.PlayerType == models.PlayerTypePlayer {
				resp.PlayerPlays = append(resp.PlayerPlays, play)
			} else {
				resp.OpponentPlays = append(resp.OpponentPlays, play)
			}
		}

		// Determine active player from snapshot if available
		if entry.Snapshot != nil && entry.Snapshot.ActivePlayer != "" {
			resp.ActivePlayer = entry.Snapshot.ActivePlayer
		}

		result = append(result, resp)
	}

	response.Success(w, result)
}

// GetMatchOpponentCards returns cards observed from the opponent during a match.
func (h *GamePlayHandler) GetMatchOpponentCards(w http.ResponseWriter, r *http.Request) {
	if !h.checkStorage(w) {
		return
	}

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
	if !h.checkStorage(w) {
		return
	}

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
	if !h.checkStorage(w) {
		return
	}

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
	if !h.checkStorage(w) {
		return
	}

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

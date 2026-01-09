package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/analysis"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// OpponentHandler handles opponent analysis API requests.
type OpponentHandler struct {
	analyzer     *analysis.OpponentAnalyzer
	opponentRepo repository.OpponentRepository
	accountID    func() int
}

// NewOpponentHandler creates a new opponent handler.
func NewOpponentHandler(analyzer *analysis.OpponentAnalyzer, opponentRepo repository.OpponentRepository, accountIDFunc func() int) *OpponentHandler {
	return &OpponentHandler{
		analyzer:     analyzer,
		opponentRepo: opponentRepo,
		accountID:    accountIDFunc,
	}
}

// GetOpponentAnalysis retrieves opponent analysis for a match.
// GET /matches/{matchID}/opponent-analysis
func (h *OpponentHandler) GetOpponentAnalysis(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.Error(w, http.StatusBadRequest, errors.New("match ID required"))
		return
	}

	opponentAnalysis, err := h.analyzer.AnalyzeOpponent(r.Context(), matchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, opponentAnalysis)
}

// ListOpponentDecks lists reconstructed opponent decks.
// GET /opponents/decks
func (h *OpponentHandler) ListOpponentDecks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	archetype := r.URL.Query().Get("archetype")
	format := r.URL.Query().Get("format")
	minConfidenceStr := r.URL.Query().Get("min_confidence")
	limitStr := r.URL.Query().Get("limit")

	filter := &repository.OpponentProfileFilter{
		Limit: 50,
	}

	if archetype != "" {
		filter.Archetype = &archetype
	}
	if format != "" {
		filter.Format = &format
	}
	if minConfidenceStr != "" {
		if minConf, err := strconv.ParseFloat(minConfidenceStr, 64); err == nil {
			filter.MinConfidence = &minConf
		}
	}
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	profiles, err := h.opponentRepo.ListProfiles(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"total":    len(profiles),
	})
}

// GetMatchupStats retrieves matchup statistics.
// GET /analytics/matchups
func (h *OpponentHandler) GetMatchupStats(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	var formatPtr *string
	if format != "" {
		formatPtr = &format
	}

	stats, err := h.analyzer.GetMatchupSummary(r.Context(), h.accountID(), formatPtr)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"matchups": stats,
		"total":    len(stats),
	})
}

// GetOpponentHistory retrieves opponent history summary.
// GET /analytics/opponent-history
func (h *OpponentHandler) GetOpponentHistory(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	var formatPtr *string
	if format != "" {
		formatPtr = &format
	}

	summary, err := h.analyzer.GetOpponentHistory(r.Context(), h.accountID(), formatPtr)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, summary)
}

// GetExpectedCards retrieves expected cards for an archetype.
// GET /archetypes/{name}/expected-cards
func (h *OpponentHandler) GetExpectedCards(w http.ResponseWriter, r *http.Request) {
	archetypeName := chi.URLParam(r, "name")
	format := r.URL.Query().Get("format")

	if archetypeName == "" {
		response.Error(w, http.StatusBadRequest, errors.New("archetype name required"))
		return
	}

	if format == "" {
		format = "Standard" // Default format
	}

	cards, err := h.opponentRepo.GetExpectedCards(r.Context(), archetypeName, format)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"archetype":     archetypeName,
		"format":        format,
		"expectedCards": cards,
		"total":         len(cards),
	})
}

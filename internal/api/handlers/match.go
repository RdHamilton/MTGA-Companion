package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MatchHandler handles match-related API requests.
type MatchHandler struct {
	facade *gui.MatchFacade
}

// NewMatchHandler creates a new MatchHandler.
func NewMatchHandler(facade *gui.MatchFacade) *MatchHandler {
	return &MatchHandler{facade: facade}
}

// StatsFilterRequest represents the JSON request body for filtering.
type StatsFilterRequest struct {
	AccountID    *int     `json:"account_id,omitempty"`
	StartDate    *string  `json:"start_date,omitempty"`
	EndDate      *string  `json:"end_date,omitempty"`
	Format       *string  `json:"format,omitempty"`
	Formats      []string `json:"formats,omitempty"`
	DeckFormat   *string  `json:"deck_format,omitempty"`
	DeckID       *string  `json:"deck_id,omitempty"`
	EventName    *string  `json:"event_name,omitempty"`
	EventNames   []string `json:"event_names,omitempty"`
	OpponentName *string  `json:"opponent_name,omitempty"`
	OpponentID   *string  `json:"opponent_id,omitempty"`
	Result       *string  `json:"result,omitempty"`
	RankClass    *string  `json:"rank_class,omitempty"`
	RankMinClass *string  `json:"rank_min_class,omitempty"`
	RankMaxClass *string  `json:"rank_max_class,omitempty"`
	ResultReason *string  `json:"result_reason,omitempty"`
}

// ToStatsFilter converts the request to a StatsFilter model.
func (r *StatsFilterRequest) ToStatsFilter() models.StatsFilter {
	filter := models.StatsFilter{
		AccountID:    r.AccountID,
		Format:       r.Format,
		Formats:      r.Formats,
		DeckFormat:   r.DeckFormat,
		DeckID:       r.DeckID,
		EventName:    r.EventName,
		EventNames:   r.EventNames,
		OpponentName: r.OpponentName,
		OpponentID:   r.OpponentID,
		Result:       r.Result,
		RankClass:    r.RankClass,
		RankMinClass: r.RankMinClass,
		RankMaxClass: r.RankMaxClass,
		ResultReason: r.ResultReason,
	}

	if r.StartDate != nil {
		if t, err := time.Parse("2006-01-02", *r.StartDate); err == nil {
			filter.StartDate = &t
		}
	}
	if r.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *r.EndDate); err == nil {
			filter.EndDate = &t
		}
	}

	return filter
}

// GetMatches returns matches based on the provided filter.
func (h *MatchHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	var req StatsFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	filter := req.ToStatsFilter()
	matches, err := h.facade.GetMatches(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, matches)
}

// GetMatch returns a single match by ID.
func (h *MatchHandler) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	// Create filter with just the match we want
	filter := models.StatsFilter{
		DeckID: &matchID, // Note: We may need a dedicated GetMatchByID method
	}

	matches, err := h.facade.GetMatches(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if len(matches) == 0 {
		response.NotFound(w, errors.New("match not found"))
		return
	}

	response.Success(w, matches[0])
}

// GetMatchGames returns games for a specific match.
func (h *MatchHandler) GetMatchGames(w http.ResponseWriter, r *http.Request) {
	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		response.BadRequest(w, errors.New("match ID is required"))
		return
	}

	games, err := h.facade.GetMatchGames(r.Context(), matchID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, games)
}

// GetStats returns statistics based on the provided filter.
func (h *MatchHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var req StatsFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	filter := req.ToStatsFilter()
	stats, err := h.facade.GetStats(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// TrendAnalysisRequest represents a request for trend analysis.
type TrendAnalysisRequest struct {
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	PeriodType string   `json:"period_type"`
	Formats    []string `json:"formats,omitempty"`
}

// GetTrendAnalysis returns trend analysis for the specified period.
func (h *MatchHandler) GetTrendAnalysis(w http.ResponseWriter, r *http.Request) {
	var req TrendAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(w, errors.New("invalid start_date format, expected YYYY-MM-DD"))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.BadRequest(w, errors.New("invalid end_date format, expected YYYY-MM-DD"))
		return
	}

	trends, err := h.facade.GetTrendAnalysis(r.Context(), startDate, endDate, req.PeriodType, req.Formats)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, trends)
}

// GetFormats returns all available match formats.
func (h *MatchHandler) GetFormats(w http.ResponseWriter, r *http.Request) {
	// Get all matches and extract unique formats
	filter := models.StatsFilter{}
	stats, err := h.facade.GetStatsByFormat(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	formats := make([]string, 0, len(stats))
	for format := range stats {
		formats = append(formats, format)
	}

	response.Success(w, formats)
}

// GetArchetypes returns all available archetypes.
func (h *MatchHandler) GetArchetypes(w http.ResponseWriter, r *http.Request) {
	// This would require a dedicated method in the facade
	// For now, return empty list
	response.Success(w, []string{})
}

// FormatDistributionRequest represents a request for format distribution.
type FormatDistributionRequest struct {
	StatsFilterRequest
}

// GetFormatDistribution returns match distribution by format.
func (h *MatchHandler) GetFormatDistribution(w http.ResponseWriter, r *http.Request) {
	var req FormatDistributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	filter := req.ToStatsFilter()
	stats, err := h.facade.GetStatsByFormat(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// GetWinRateOverTime returns win rate trends over time.
func (h *MatchHandler) GetWinRateOverTime(w http.ResponseWriter, r *http.Request) {
	var req TrendAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(w, errors.New("invalid start_date format"))
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.BadRequest(w, errors.New("invalid end_date format"))
		return
	}

	trends, err := h.facade.GetTrendAnalysis(r.Context(), startDate, endDate, req.PeriodType, req.Formats)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, trends)
}

// GetPerformanceByHour returns performance metrics grouped by hour.
func (h *MatchHandler) GetPerformanceByHour(w http.ResponseWriter, r *http.Request) {
	var req StatsFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	filter := req.ToStatsFilter()
	metrics, err := h.facade.GetPerformanceMetrics(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, metrics)
}

// GetMatchupMatrix returns win rates against different deck types.
func (h *MatchHandler) GetMatchupMatrix(w http.ResponseWriter, r *http.Request) {
	var req StatsFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	filter := req.ToStatsFilter()
	stats, err := h.facade.GetStatsByDeck(r.Context(), filter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

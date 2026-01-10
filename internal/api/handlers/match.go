package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
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

	// Return empty array instead of nil
	if matches == nil {
		matches = []*models.Match{}
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

// GetRankProgression returns rank progression data for a specific format.
func (h *MatchHandler) GetRankProgression(w http.ResponseWriter, r *http.Request) {
	format := chi.URLParam(r, "format")
	if format == "" {
		response.BadRequest(w, errors.New("format is required"))
		return
	}

	progression, err := h.facade.GetRankProgression(r.Context(), format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if progression == nil {
		// Return empty progression structure
		response.Success(w, map[string]interface{}{
			"format":       format,
			"current_rank": nil,
			"peak_rank":    nil,
			"season_start": nil,
			"matches_won":  0,
			"matches_lost": 0,
			"win_rate":     0.0,
			"rank_changes": []interface{}{},
		})
		return
	}

	response.Success(w, progression)
}

// GetRankProgressionTimeline returns a timeline of rank progression.
func (h *MatchHandler) GetRankProgressionTimeline(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "constructed"
	}

	// Parse date parameters
	var startDate, endDate *time.Time
	if startStr := r.URL.Query().Get("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := r.URL.Query().Get("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &t
		}
	}

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		periodStr = "daily"
	}

	// Convert period string to TimelinePeriod
	var period storage.TimelinePeriod
	switch periodStr {
	case "all":
		period = storage.PeriodAll
	case "daily":
		period = storage.PeriodDaily
	case "weekly":
		period = storage.PeriodWeekly
	case "monthly":
		period = storage.PeriodMonthly
	default:
		period = storage.PeriodDaily
	}

	timeline, err := h.facade.GetRankProgressionTimeline(r.Context(), format, startDate, endDate, period)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, timeline)
}

// ComparisonGroupRequest represents a group in a comparison request.
type ComparisonGroupRequest struct {
	Label  string             `json:"label"`
	Filter StatsFilterRequest `json:"filter"`
}

// CompareMatchesRequest represents a request to compare multiple groups.
type CompareMatchesRequest struct {
	Groups []ComparisonGroupRequest `json:"groups"`
}

// CompareFormatsRequest represents a request to compare formats.
type CompareFormatsRequest struct {
	Formats    []string           `json:"formats"`
	BaseFilter StatsFilterRequest `json:"base_filter,omitempty"`
}

// CompareDecksRequest represents a request to compare decks.
type CompareDecksRequest struct {
	DeckIDs    []string           `json:"deck_ids"`
	BaseFilter StatsFilterRequest `json:"base_filter,omitempty"`
}

// TimePeriodRequest represents a time period for comparison.
type TimePeriodRequest struct {
	Label     string `json:"label"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// CompareTimePeriodsRequest represents a request to compare time periods.
type CompareTimePeriodsRequest struct {
	Periods    []TimePeriodRequest `json:"periods"`
	BaseFilter StatsFilterRequest  `json:"base_filter,omitempty"`
}

// CompareMatches compares multiple groups of matches.
func (h *MatchHandler) CompareMatches(w http.ResponseWriter, r *http.Request) {
	var req CompareMatchesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if len(req.Groups) < 2 {
		response.BadRequest(w, errors.New("need at least 2 groups to compare"))
		return
	}

	groups := make([]storage.ComparisonGroup, 0, len(req.Groups))
	for _, g := range req.Groups {
		if g.Label == "" {
			response.BadRequest(w, errors.New("group label is required"))
			return
		}
		// Validate date strings if provided
		if g.Filter.StartDate != nil && *g.Filter.StartDate != "" {
			if _, err := time.Parse("2006-01-02", *g.Filter.StartDate); err != nil {
				response.BadRequest(w, errors.New("invalid start_date format, expected YYYY-MM-DD"))
				return
			}
		}
		if g.Filter.EndDate != nil && *g.Filter.EndDate != "" {
			if _, err := time.Parse("2006-01-02", *g.Filter.EndDate); err != nil {
				response.BadRequest(w, errors.New("invalid end_date format, expected YYYY-MM-DD"))
				return
			}
		}
		groups = append(groups, storage.ComparisonGroup{
			Label:  g.Label,
			Filter: g.Filter.ToStatsFilter(),
		})
	}

	result, err := h.facade.CompareMatches(r.Context(), groups)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// CompareFormats compares performance across different formats.
func (h *MatchHandler) CompareFormats(w http.ResponseWriter, r *http.Request) {
	var req CompareFormatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// Filter out empty strings
	validFormats := make([]string, 0, len(req.Formats))
	for _, f := range req.Formats {
		if f != "" {
			validFormats = append(validFormats, f)
		}
	}

	if len(validFormats) < 2 {
		response.BadRequest(w, errors.New("need at least 2 non-empty formats to compare"))
		return
	}

	baseFilter := req.BaseFilter.ToStatsFilter()
	result, err := h.facade.CompareFormats(r.Context(), validFormats, baseFilter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// CompareDecks compares performance across different decks.
func (h *MatchHandler) CompareDecks(w http.ResponseWriter, r *http.Request) {
	var req CompareDecksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	// Filter out empty strings
	validDeckIDs := make([]string, 0, len(req.DeckIDs))
	for _, id := range req.DeckIDs {
		if id != "" {
			validDeckIDs = append(validDeckIDs, id)
		}
	}

	if len(validDeckIDs) < 2 {
		response.BadRequest(w, errors.New("need at least 2 non-empty deck IDs to compare"))
		return
	}

	baseFilter := req.BaseFilter.ToStatsFilter()
	result, err := h.facade.CompareDecks(r.Context(), validDeckIDs, baseFilter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// CompareTimePeriods compares performance across different time periods.
func (h *MatchHandler) CompareTimePeriods(w http.ResponseWriter, r *http.Request) {
	var req CompareTimePeriodsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if len(req.Periods) < 2 {
		response.BadRequest(w, errors.New("need at least 2 time periods to compare"))
		return
	}

	periods := make([]storage.TimePeriod, 0, len(req.Periods))
	for _, p := range req.Periods {
		// Skip periods with empty labels or dates
		if p.Label == "" || p.StartDate == "" || p.EndDate == "" {
			continue
		}
		start, err := time.Parse("2006-01-02", p.StartDate)
		if err != nil {
			response.BadRequest(w, errors.New("invalid start_date format, expected YYYY-MM-DD"))
			return
		}
		end, err := time.Parse("2006-01-02", p.EndDate)
		if err != nil {
			response.BadRequest(w, errors.New("invalid end_date format, expected YYYY-MM-DD"))
			return
		}
		if start.After(end) {
			response.BadRequest(w, errors.New("start_date must be on/before end_date"))
			return
		}
		periods = append(periods, storage.TimePeriod{
			Label: p.Label,
			Start: start,
			End:   end,
		})
	}

	if len(periods) < 2 {
		response.BadRequest(w, errors.New("need at least 2 valid time periods to compare"))
		return
	}

	baseFilter := req.BaseFilter.ToStatsFilter()
	result, err := h.facade.CompareTimePeriods(r.Context(), periods, baseFilter)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftHandler handles draft-related API requests.
type DraftHandler struct {
	facade *gui.DraftFacade
}

// NewDraftHandler creates a new DraftHandler.
func NewDraftHandler(facade *gui.DraftFacade) *DraftHandler {
	return &DraftHandler{facade: facade}
}

// DraftFilterRequest represents the JSON request body for draft filtering.
type DraftFilterRequest struct {
	SetCode   *string `json:"set_code,omitempty"`
	DraftType *string `json:"draft_type,omitempty"`
	Status    *string `json:"status,omitempty"`
	Limit     int     `json:"limit,omitempty"`
}

// GetDraftSessions returns draft sessions based on filters.
func (h *DraftHandler) GetDraftSessions(w http.ResponseWriter, r *http.Request) {
	var req DraftFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	// Filter by status if provided
	if req.Status != nil {
		switch *req.Status {
		case "active":
			activeSessions, err := h.facade.GetActiveDraftSessions(r.Context())
			if err != nil {
				response.InternalError(w, err)
				return
			}
			if activeSessions == nil {
				activeSessions = []*models.DraftSession{}
			}
			response.Success(w, activeSessions)
			return
		case "completed":
			completedSessions, err := h.facade.GetCompletedDraftSessions(r.Context(), limit)
			if err != nil {
				response.InternalError(w, err)
				return
			}
			if completedSessions == nil {
				completedSessions = []*models.DraftSession{}
			}
			response.Success(w, completedSessions)
			return
		}
	}

	// Get both active and completed sessions when no status filter
	activeSessions, err := h.facade.GetActiveDraftSessions(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	completedSessions, err := h.facade.GetCompletedDraftSessions(r.Context(), limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Combine results
	allSessions := append(activeSessions, completedSessions...)
	if allSessions == nil {
		allSessions = []*models.DraftSession{}
	}

	response.Success(w, allSessions)
}

// GetDraftSession returns a single draft session by ID.
func (h *DraftHandler) GetDraftSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	session, err := h.facade.GetDraftSession(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if session == nil {
		response.NotFound(w, errors.New("draft session not found"))
		return
	}

	response.Success(w, session)
}

// GetDraftPicks returns picks for a draft session.
func (h *DraftHandler) GetDraftPicks(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	picks, err := h.facade.GetDraftPicks(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, picks)
}

// GetDraftPool returns the deck metrics for a draft session.
func (h *DraftHandler) GetDraftPool(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	metrics, err := h.facade.GetDraftDeckMetrics(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, metrics)
}

// GetDraftAnalysis returns the grade for a draft session.
func (h *DraftHandler) GetDraftAnalysis(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	grade, err := h.facade.GetDraftGrade(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, grade)
}

// GetDraftCurve returns the deck metrics for a draft session.
func (h *DraftHandler) GetDraftCurve(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	metrics, err := h.facade.GetDraftDeckMetrics(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, metrics)
}

// GetDraftColors returns the deck metrics for a draft session.
func (h *DraftHandler) GetDraftColors(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	metrics, err := h.facade.GetDraftDeckMetrics(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, metrics)
}

// GetDraftDeckMetrics returns comprehensive deck metrics for a draft session.
func (h *DraftHandler) GetDraftDeckMetrics(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	metrics, err := h.facade.GetDraftDeckMetrics(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, metrics)
}

// DraftStatsRequest represents a request for draft statistics.
type DraftStatsRequest struct {
	SetCode   *string `json:"set_code,omitempty"`
	DraftType *string `json:"draft_type,omitempty"`
}

// GetDraftStats returns draft performance metrics.
func (h *DraftHandler) GetDraftStats(w http.ResponseWriter, r *http.Request) {
	stats := h.facade.GetDraftPerformanceMetrics(r.Context())
	response.Success(w, stats)
}

// GetDraftFormats returns available draft sets (from completed sessions).
func (h *DraftHandler) GetDraftFormats(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.facade.GetCompletedDraftSessions(r.Context(), 100)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Extract unique set codes
	formatSet := make(map[string]bool)
	for _, s := range sessions {
		if s.SetCode != "" {
			formatSet[s.SetCode] = true
		}
	}

	formats := make([]string, 0, len(formatSet))
	for f := range formatSet {
		formats = append(formats, f)
	}

	response.Success(w, formats)
}

// GetRecentDrafts returns recent draft sessions.
func (h *DraftHandler) GetRecentDrafts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	sessions, err := h.facade.GetCompletedDraftSessions(r.Context(), limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, sessions)
}

// GradePickRequest represents a request to grade a draft pick.
type GradePickRequest struct {
	SessionID  string `json:"session_id"`
	PackNumber int    `json:"pack_number"`
	PickNumber int    `json:"pick_number"`
}

// GradePick grades a draft pick using pick alternatives.
func (h *DraftHandler) GradePick(w http.ResponseWriter, r *http.Request) {
	var req GradePickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	quality, err := h.facade.GetPickAlternatives(r.Context(), req.SessionID, req.PackNumber, req.PickNumber)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, quality)
}

// DraftInsightsRequest represents a request for draft insights.
type DraftInsightsRequest struct {
	SetCode     string `json:"set_code"`
	DraftFormat string `json:"draft_format"`
}

// GetDraftInsights returns format insights for a set.
func (h *DraftHandler) GetDraftInsights(w http.ResponseWriter, r *http.Request) {
	var req DraftInsightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	insights, err := h.facade.GetFormatInsights(r.Context(), req.SetCode, req.DraftFormat)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, insights)
}

// WinProbabilityRequest represents a request for win probability prediction.
type WinProbabilityRequest struct {
	SessionID string `json:"session_id"`
}

// PredictWinProbability predicts win probability for a draft.
func (h *DraftHandler) PredictWinProbability(w http.ResponseWriter, r *http.Request) {
	var req WinProbabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	prediction, err := h.facade.GetDraftWinRatePrediction(r.Context(), req.SessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, prediction)
}

// GetDraftPacks returns pack data for a draft session.
func (h *DraftHandler) GetDraftPacks(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	packs, err := h.facade.GetDraftPacks(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, packs)
}

// MissingCardsRequest represents a request for missing cards analysis.
type MissingCardsRequest struct {
	PackNum int `json:"pack_num"`
	PickNum int `json:"pick_num"`
}

// GetMissingCards returns missing cards analysis for a pick.
func (h *DraftHandler) GetMissingCards(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	var req MissingCardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	analysis, err := h.facade.GetMissingCards(r.Context(), sessionID, req.PackNum, req.PickNum)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, analysis)
}

// AnalyzePickQuality triggers pick quality analysis for a session.
func (h *DraftHandler) AnalyzePickQuality(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	if err := h.facade.AnalyzeSessionPickQuality(r.Context(), sessionID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// CalculateGrade calculates draft grade for a session.
func (h *DraftHandler) CalculateGrade(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	grade, err := h.facade.CalculateDraftGrade(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, grade)
}

// CalculatePrediction calculates win rate prediction for a session.
func (h *DraftHandler) CalculatePrediction(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	prediction, err := h.facade.PredictDraftWinRate(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, prediction)
}

// RepairSession repairs a draft session.
func (h *DraftHandler) RepairSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	if err := h.facade.RepairDraftSession(r.Context(), sessionID); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// ArchetypeCardsRequest represents a request for archetype cards.
type ArchetypeCardsRequest struct {
	SetCode     string `json:"set_code"`
	DraftFormat string `json:"draft_format"`
	Colors      string `json:"colors"`
}

// GetArchetypeCards returns cards for a color archetype.
func (h *DraftHandler) GetArchetypeCards(w http.ResponseWriter, r *http.Request) {
	var req ArchetypeCardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	cards, err := h.facade.GetArchetypeCards(r.Context(), req.SetCode, req.DraftFormat, req.Colors)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, cards)
}

// GetCurrentPack returns current pack with recommendation for a session.
func (h *DraftHandler) GetCurrentPack(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		response.BadRequest(w, errors.New("session ID is required"))
		return
	}

	pack, err := h.facade.GetCurrentPackWithRecommendation(r.Context(), sessionID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, pack)
}

// ResetStats resets draft performance metrics.
func (h *DraftHandler) ResetStats(w http.ResponseWriter, r *http.Request) {
	h.facade.ResetDraftPerformanceMetrics(r.Context())
	response.Success(w, map[string]string{"status": "success"})
}

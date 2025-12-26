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

	// Get both active and completed sessions
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

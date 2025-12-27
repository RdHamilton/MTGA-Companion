package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// FeedbackHandler handles feedback-related API requests.
type FeedbackHandler struct {
	facade *gui.FeedbackFacade
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(facade *gui.FeedbackFacade) *FeedbackHandler {
	return &FeedbackHandler{facade: facade}
}

// SubmitFeedback submits general feedback (legacy endpoint).
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	// Redirect to RecordRecommendation for actual feedback recording
	h.RecordRecommendation(w, r)
}

// SubmitBugReport submits a bug report (legacy endpoint).
func (h *FeedbackHandler) SubmitBugReport(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{
		"status":  "acknowledged",
		"message": "Bug reports should be submitted via GitHub issues",
	})
}

// SubmitFeatureRequest submits a feature request (legacy endpoint).
func (h *FeedbackHandler) SubmitFeatureRequest(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{
		"status":  "acknowledged",
		"message": "Feature requests should be submitted via GitHub issues",
	})
}

// RecordRecommendation records a new recommendation event.
func (h *FeedbackHandler) RecordRecommendation(w http.ResponseWriter, r *http.Request) {
	var req gui.RecordRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	result, err := h.facade.RecordRecommendation(r.Context(), &req)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

// RecordAction records the user's action on a recommendation.
func (h *FeedbackHandler) RecordAction(w http.ResponseWriter, r *http.Request) {
	var req gui.RecordActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.RecommendationID == "" {
		response.BadRequest(w, errors.New("recommendation_id is required"))
		return
	}

	if req.Action == "" {
		response.BadRequest(w, errors.New("action is required"))
		return
	}

	if err := h.facade.RecordAction(r.Context(), &req); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// RecordOutcome records the match outcome for a recommendation.
func (h *FeedbackHandler) RecordOutcome(w http.ResponseWriter, r *http.Request) {
	var req gui.RecordOutcomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.RecommendationID == "" {
		response.BadRequest(w, errors.New("recommendation_id is required"))
		return
	}

	if req.MatchID == "" {
		response.BadRequest(w, errors.New("match_id is required"))
		return
	}

	if req.Result == "" {
		response.BadRequest(w, errors.New("result is required"))
		return
	}

	if err := h.facade.RecordOutcome(r.Context(), &req); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// GetRecommendationStatsRequest represents a request for recommendation stats.
type GetRecommendationStatsRequest struct {
	Type *string `json:"type,omitempty"`
}

// GetRecommendationStats returns aggregated recommendation statistics.
func (h *FeedbackHandler) GetRecommendationStats(w http.ResponseWriter, r *http.Request) {
	var recType *string
	typeParam := r.URL.Query().Get("type")
	if typeParam != "" {
		recType = &typeParam
	}

	stats, err := h.facade.GetRecommendationStats(r.Context(), recType)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, stats)
}

// GetDashboardMetrics returns comprehensive feedback metrics for the dashboard.
func (h *FeedbackHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.facade.GetDashboardMetrics(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, metrics)
}

// ExportMLTrainingData exports feedback data for ML training.
func (h *FeedbackHandler) ExportMLTrainingData(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 1000
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	data, err := h.facade.ExportMLTrainingData(r.Context(), limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, data)
}

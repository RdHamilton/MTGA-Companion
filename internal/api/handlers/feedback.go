package handlers

import (
	"net/http"

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

// SubmitFeedback submits general feedback (placeholder).
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Feedback submission requires facade implementation"})
}

// SubmitBugReport submits a bug report (placeholder).
func (h *FeedbackHandler) SubmitBugReport(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Bug report submission requires facade implementation"})
}

// SubmitFeatureRequest submits a feature request (placeholder).
func (h *FeedbackHandler) SubmitFeatureRequest(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Feature request submission requires facade implementation"})
}

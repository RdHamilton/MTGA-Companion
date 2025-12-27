package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// QuestHandler handles quest-related API requests.
type QuestHandler struct {
	facade *gui.MatchFacade
}

// NewQuestHandler creates a new QuestHandler.
func NewQuestHandler(facade *gui.MatchFacade) *QuestHandler {
	return &QuestHandler{facade: facade}
}

// GetActiveQuests returns active quests.
func (h *QuestHandler) GetActiveQuests(w http.ResponseWriter, r *http.Request) {
	quests, err := h.facade.GetActiveQuests(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Return empty array instead of nil
	if quests == nil {
		quests = []*models.Quest{}
	}

	response.Success(w, quests)
}

// GetQuestHistory returns quest history.
func (h *QuestHandler) GetQuestHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	limitStr := r.URL.Query().Get("limit")

	limit := 50 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Default dates if not provided
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02") // 1 month ago
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	quests, err := h.facade.GetQuestHistory(r.Context(), startDate, endDate, limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Return empty array instead of nil
	if quests == nil {
		quests = []*models.Quest{}
	}

	response.Success(w, quests)
}

// GetDailyWins returns daily wins progress.
func (h *QuestHandler) GetDailyWins(w http.ResponseWriter, r *http.Request) {
	account, err := h.facade.GetCurrentAccount(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"dailyWins": account.DailyWins,
		"goal":      15,
	})
}

// GetWeeklyWins returns weekly wins progress.
func (h *QuestHandler) GetWeeklyWins(w http.ResponseWriter, r *http.Request) {
	account, err := h.facade.GetCurrentAccount(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"weeklyWins": account.WeeklyWins,
		"goal":       15,
	})
}

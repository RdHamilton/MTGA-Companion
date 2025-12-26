package handlers

import (
	"net/http"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// QuestHandler handles quest-related API requests.
type QuestHandler struct {
	facade *gui.SystemFacade
}

// NewQuestHandler creates a new QuestHandler.
func NewQuestHandler(facade *gui.SystemFacade) *QuestHandler {
	return &QuestHandler{facade: facade}
}

// GetActiveQuests returns active quests (placeholder).
func (h *QuestHandler) GetActiveQuests(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Active quests requires facade method"})
}

// GetQuestHistory returns quest history (placeholder).
func (h *QuestHandler) GetQuestHistory(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Quest history requires facade method"})
}

// GetDailyWins returns daily wins progress (placeholder).
func (h *QuestHandler) GetDailyWins(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Daily wins requires facade method"})
}

// GetWeeklyWins returns weekly wins progress (placeholder).
func (h *QuestHandler) GetWeeklyWins(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Weekly wins requires facade method"})
}

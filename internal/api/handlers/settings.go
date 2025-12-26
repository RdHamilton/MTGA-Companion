package handlers

import (
	"net/http"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// SettingsHandler handles settings-related API requests.
type SettingsHandler struct {
	facade *gui.SettingsFacade
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(facade *gui.SettingsFacade) *SettingsHandler {
	return &SettingsHandler{facade: facade}
}

// GetSettings returns all settings (placeholder).
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Get settings requires facade method"})
}

// UpdateSettings updates multiple settings (placeholder).
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Update settings requires facade method"})
}

// GetSetting returns a single setting (placeholder).
func (h *SettingsHandler) GetSetting(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Get setting requires facade method"})
}

// UpdateSetting updates a single setting (placeholder).
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Update setting requires facade method"})
}

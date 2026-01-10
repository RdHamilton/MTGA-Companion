package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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

// GetSettings returns all settings.
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.facade.GetAllSettings(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to get settings: %w", err))
		return
	}
	response.Success(w, settings)
}

// UpdateSettings updates multiple settings.
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings gui.AppSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if err := h.facade.SaveAllSettings(r.Context(), &settings); err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to save settings: %w", err))
		return
	}

	response.Success(w, settings)
}

// GetSetting returns a single setting by key.
func (h *SettingsHandler) GetSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.Error(w, http.StatusBadRequest, errors.New("setting key is required"))
		return
	}

	value, err := h.facade.GetSetting(r.Context(), key)
	if err != nil {
		response.Error(w, http.StatusNotFound, fmt.Errorf("setting not found: %s", key))
		return
	}

	response.Success(w, map[string]interface{}{"key": key, "value": value})
}

// UpdateSetting updates a single setting by key.
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.Error(w, http.StatusBadRequest, errors.New("setting key is required"))
		return
	}

	var body struct {
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if err := h.facade.SetSetting(r.Context(), key, body.Value); err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to save setting: %w", err))
		return
	}

	response.Success(w, map[string]interface{}{"key": key, "value": body.Value})
}

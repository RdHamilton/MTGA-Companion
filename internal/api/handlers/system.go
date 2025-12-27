package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// SystemHandler handles system-related API requests.
type SystemHandler struct {
	facade *gui.SystemFacade
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(facade *gui.SystemFacade) *SystemHandler {
	return &SystemHandler{facade: facade}
}

// GetStatus returns the system status.
func (h *SystemHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.facade.GetConnectionStatus()
	response.Success(w, status)
}

// GetDaemonStatus returns the daemon connection status.
func (h *SystemHandler) GetDaemonStatus(w http.ResponseWriter, r *http.Request) {
	status := h.facade.GetConnectionStatus()
	response.Success(w, status)
}

// ConnectDaemon attempts to connect to the daemon.
func (h *SystemHandler) ConnectDaemon(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.ReconnectToDaemon(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "connected"})
}

// DisconnectDaemon disconnects from the daemon.
func (h *SystemHandler) DisconnectDaemon(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.SwitchToStandaloneMode(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "disconnected"})
}

// GetVersion returns the application version.
func (h *SystemHandler) GetVersion(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{
		"version": "1.4.0",
		"service": "mtga-companion-api",
	})
}

// GetCurrentAccount returns the current account information.
func (h *SystemHandler) GetCurrentAccount(w http.ResponseWriter, r *http.Request) {
	account, err := h.facade.GetCurrentAccount(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, account)
}

// GetDatabasePath returns the current database path (placeholder).
func (h *SystemHandler) GetDatabasePath(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Database path requires facade method"})
}

// SetDatabasePath sets the database path (placeholder).
func (h *SystemHandler) SetDatabasePath(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Set database path requires facade method"})
}

// SetDaemonPortRequest represents a request to set daemon port.
type SetDaemonPortRequest struct {
	Port int `json:"port"`
}

// SetDaemonPort sets the daemon port.
func (h *SystemHandler) SetDaemonPort(w http.ResponseWriter, r *http.Request) {
	var req SetDaemonPortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		response.BadRequest(w, errors.New("invalid port number"))
		return
	}

	if err := h.facade.SetDaemonPort(req.Port); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]interface{}{
		"status": "success",
		"port":   req.Port,
	})
}

// SwitchToDaemonMode switches to daemon mode.
func (h *SystemHandler) SwitchToDaemonMode(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.SwitchToDaemonMode(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success", "mode": "daemon"})
}

// SwitchToStandaloneMode switches to standalone mode.
func (h *SystemHandler) SwitchToStandaloneMode(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.SwitchToStandaloneMode(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success", "mode": "standalone"})
}

// TriggerReplayRequest represents a request to trigger log replay.
type TriggerReplayRequest struct {
	ClearData bool `json:"clear_data,omitempty"`
}

// TriggerReplay triggers log replay.
func (h *SystemHandler) TriggerReplay(w http.ResponseWriter, r *http.Request) {
	var req TriggerReplayRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, errors.New("invalid request body"))
			return
		}
	}

	if err := h.facade.TriggerReplayLogs(r.Context(), req.ClearData); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "success"})
}

// PauseReplay pauses the current replay.
func (h *SystemHandler) PauseReplay(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.PauseReplay(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "paused"})
}

// ResumeReplay resumes the current replay.
func (h *SystemHandler) ResumeReplay(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.ResumeReplay(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "resumed"})
}

// StopReplay stops the current replay.
func (h *SystemHandler) StopReplay(w http.ResponseWriter, r *http.Request) {
	if err := h.facade.StopReplay(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{"status": "stopped"})
}

// GetReplayStatus returns the current replay status.
func (h *SystemHandler) GetReplayStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.facade.GetReplayStatus(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, status)
}

// GetReplayProgress returns the current replay progress.
func (h *SystemHandler) GetReplayProgress(w http.ResponseWriter, r *http.Request) {
	progress, err := h.facade.GetLogReplayProgress(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, progress)
}

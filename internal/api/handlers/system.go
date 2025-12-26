package handlers

import (
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

// GetDatabasePath returns the current database path (placeholder).
func (h *SystemHandler) GetDatabasePath(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Database path requires facade method"})
}

// SetDatabasePath sets the database path (placeholder).
func (h *SystemHandler) SetDatabasePath(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Set database path requires facade method"})
}

package handlers

import (
	"net/http"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// ExportHandler handles export-related API requests.
type ExportHandler struct {
	facade *gui.ExportFacade
}

// NewExportHandler creates a new ExportHandler.
func NewExportHandler(facade *gui.ExportFacade) *ExportHandler {
	return &ExportHandler{facade: facade}
}

// ExportMatches exports matches (placeholder - needs facade implementation).
func (h *ExportHandler) ExportMatches(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Export matches endpoint requires facade implementation"})
}

// ExportDrafts exports drafts (placeholder - needs facade implementation).
func (h *ExportHandler) ExportDrafts(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Export drafts endpoint requires facade implementation"})
}

// ExportCollection exports collection (placeholder - needs facade implementation).
func (h *ExportHandler) ExportCollection(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Export collection endpoint requires facade implementation"})
}

// ExportDeck exports a deck (placeholder - needs facade implementation).
func (h *ExportHandler) ExportDeck(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Export deck endpoint requires facade implementation"})
}

// GetExportFormats returns available export formats.
func (h *ExportHandler) GetExportFormats(w http.ResponseWriter, _ *http.Request) {
	formats := []map[string]string{
		{"id": "csv", "name": "CSV", "description": "Comma-separated values"},
		{"id": "json", "name": "JSON", "description": "JavaScript Object Notation"},
		{"id": "mtga", "name": "MTGA", "description": "MTG Arena format"},
		{"id": "arena", "name": "Arena", "description": "Arena export format"},
		{"id": "text", "name": "Text", "description": "Plain text format"},
	}

	response.Success(w, formats)
}

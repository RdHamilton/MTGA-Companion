package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

// ExportMatchesRequest represents a request to export matches.
type ExportMatchesRequest struct {
	Format string `json:"format"` // "json" or "csv"
}

// ExportMatches exports matches in JSON format.
func (h *ExportHandler) ExportMatches(w http.ResponseWriter, r *http.Request) {
	// Get export data from facade (nil filter = all matches)
	data, err := h.facade.GetMatchesExportData(r.Context(), nil)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, data)
}

// ExportDraftsRequest represents a request to export drafts.
type ExportDraftsRequest struct {
	Limit int `json:"limit,omitempty"`
}

// ExportDrafts exports drafts in JSON format.
func (h *ExportHandler) ExportDrafts(w http.ResponseWriter, r *http.Request) {
	var req ExportDraftsRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, errors.New("invalid request body"))
			return
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}

	// Get export data from facade
	data, err := h.facade.GetDraftsExportData(r.Context(), limit)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, data)
}

// ExportCollection exports collection in JSON format.
func (h *ExportHandler) ExportCollection(w http.ResponseWriter, r *http.Request) {
	// Get export data from facade
	data, err := h.facade.GetCollectionExportData(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, data)
}

// DeckExportRequest represents a request to export a deck.
type DeckExportRequest struct {
	DeckID string `json:"deck_id"`
	Format string `json:"format"` // "json", "mtga", "arena", "text"
}

// ExportDeck exports a deck in the requested format.
func (h *ExportHandler) ExportDeck(w http.ResponseWriter, r *http.Request) {
	var req DeckExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.DeckID == "" {
		response.BadRequest(w, errors.New("deck_id is required"))
		return
	}

	if req.Format == "" {
		req.Format = "mtga"
	}

	// Get exported deck from facade
	exportedDeck, err := h.facade.GetDeckExportData(r.Context(), req.DeckID, req.Format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, exportedDeck)
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

// ImportMatchesRequest represents a request to import matches.
type ImportMatchesRequest struct {
	Matches []interface{} `json:"matches"`
}

// ImportMatches imports matches from JSON data.
func (h *ExportHandler) ImportMatches(w http.ResponseWriter, r *http.Request) {
	// For now, just acknowledge the request
	// Full import would require typed match data
	response.Success(w, map[string]string{
		"status":  "acknowledged",
		"message": "Match import requires structured data - use log file import for historical data",
	})
}

// ClearDataRequest represents a request to clear all data.
type ClearDataRequest struct {
	Confirmed bool `json:"confirmed"`
}

// ClearAllData clears all data with confirmation.
func (h *ExportHandler) ClearAllData(w http.ResponseWriter, r *http.Request) {
	var req ClearDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if !req.Confirmed {
		response.BadRequest(w, errors.New("confirmation required: set confirmed=true to proceed"))
		return
	}

	if err := h.facade.ClearAllDataWithoutDialog(r.Context()); err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, map[string]string{
		"status":  "success",
		"message": "All data cleared successfully",
	})
}

// ImportLogFileRequest represents a request to import a log file.
type ImportLogFileRequest struct {
	Content  string `json:"content"`  // Base64 encoded log file content
	FileName string `json:"fileName"` // Original file name
}

// ImportLogFile imports an MTGA log file.
func (h *ExportHandler) ImportLogFile(w http.ResponseWriter, r *http.Request) {
	var req ImportLogFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	if req.Content == "" {
		response.BadRequest(w, errors.New("content is required"))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if _, err := strconv.Atoi(limitStr); err != nil {
			response.BadRequest(w, errors.New("invalid limit parameter"))
			return
		}
	}

	result, err := h.facade.ImportLogFileData(r.Context(), req.Content, req.FileName)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, result)
}

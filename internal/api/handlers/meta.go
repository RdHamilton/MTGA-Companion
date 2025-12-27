package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// MetaHandler handles meta-related API requests.
type MetaHandler struct {
	facade *gui.MetaFacade
}

// NewMetaHandler creates a new MetaHandler.
func NewMetaHandler(facade *gui.MetaFacade) *MetaHandler {
	return &MetaHandler{facade: facade}
}

// GetMetaArchetypes returns meta archetypes (alias for GetMetaDashboard).
func (h *MetaHandler) GetMetaArchetypes(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "standard"
	}

	dashboard, err := h.facade.GetMetaDashboard(r.Context(), format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, dashboard.Archetypes)
}

// GetDeckAnalysis returns deck analysis from meta dashboard.
func (h *MetaHandler) GetDeckAnalysis(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "standard"
	}

	dashboard, err := h.facade.GetMetaDashboard(r.Context(), format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, dashboard)
}

// IdentifyArchetypeRequest represents a request to identify an archetype.
type IdentifyArchetypeRequest struct {
	Colors   []string `json:"colors"`
	CardIDs  []int    `json:"card_ids,omitempty"`
	DeckName string   `json:"deck_name,omitempty"`
	Format   string   `json:"format,omitempty"`
}

// IdentifyArchetype identifies the archetype of a deck.
func (h *MetaHandler) IdentifyArchetype(w http.ResponseWriter, r *http.Request) {
	var req IdentifyArchetypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, errors.New("invalid request body"))
		return
	}

	format := req.Format
	if format == "" {
		format = "standard"
	}

	// Get meta data and try to match archetype based on colors
	dashboard, err := h.facade.GetMetaDashboard(r.Context(), format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Simple matching based on colors
	var matchedArchetypes []*gui.ArchetypeInfo
	for _, arch := range dashboard.Archetypes {
		if colorsMatch(arch.Colors, req.Colors) {
			matchedArchetypes = append(matchedArchetypes, arch)
		}
	}

	if len(matchedArchetypes) == 0 {
		response.Success(w, map[string]interface{}{
			"identified":  false,
			"message":     "No matching archetype found",
			"suggestions": dashboard.Archetypes[:min(5, len(dashboard.Archetypes))],
		})
		return
	}

	response.Success(w, map[string]interface{}{
		"identified": true,
		"archetype":  matchedArchetypes[0],
		"matches":    matchedArchetypes,
	})
}

// GetMetaDashboard returns the full meta dashboard.
func (h *MetaHandler) GetMetaDashboard(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "standard"
	}

	dashboard, err := h.facade.GetMetaDashboard(r.Context(), format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, dashboard)
}

// RefreshMetaData forces a refresh of meta data.
func (h *MetaHandler) RefreshMetaData(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "standard"
	}

	dashboard, err := h.facade.RefreshMetaData(r.Context(), format)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, dashboard)
}

// GetSupportedFormats returns the list of supported formats.
func (h *MetaHandler) GetSupportedFormats(w http.ResponseWriter, _ *http.Request) {
	formats := h.facade.GetSupportedFormats()
	response.Success(w, formats)
}

// GetTierArchetypes returns archetypes for a specific tier.
func (h *MetaHandler) GetTierArchetypes(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "standard"
	}

	tierStr := r.URL.Query().Get("tier")
	tier := 1
	if tierStr != "" {
		if t, err := strconv.Atoi(tierStr); err == nil && t > 0 {
			tier = t
		}
	}

	archetypes, err := h.facade.GetTierArchetypes(r.Context(), format, tier)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, archetypes)
}

// Helper function to check if colors match
func colorsMatch(archColors, deckColors []string) bool {
	if len(archColors) != len(deckColors) {
		return false
	}

	archSet := make(map[string]bool)
	for _, c := range archColors {
		archSet[c] = true
	}

	for _, c := range deckColors {
		if !archSet[c] {
			return false
		}
	}

	return true
}

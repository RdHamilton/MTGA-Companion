package handlers

import (
	"net/http"

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

// GetMetaArchetypes returns meta archetypes (placeholder).
func (h *MetaHandler) GetMetaArchetypes(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Meta archetypes requires facade implementation"})
}

// GetDeckAnalysis returns deck analysis (placeholder).
func (h *MetaHandler) GetDeckAnalysis(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Deck analysis requires facade implementation"})
}

// IdentifyArchetype identifies the archetype of a deck (placeholder).
func (h *MetaHandler) IdentifyArchetype(w http.ResponseWriter, _ *http.Request) {
	response.Success(w, map[string]string{"status": "not_implemented", "message": "Archetype identification requires facade implementation"})
}

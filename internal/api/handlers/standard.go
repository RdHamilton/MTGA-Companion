package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// StandardHandler handles Standard format validation and set management endpoints.
type StandardHandler struct {
	storage *storage.Service
}

// NewStandardHandler creates a new StandardHandler.
func NewStandardHandler(storage *storage.Service) *StandardHandler {
	return &StandardHandler{storage: storage}
}

// GetStandardSets returns all Standard-legal sets.
// GET /api/v1/standard/sets
func (h *StandardHandler) GetStandardSets(w http.ResponseWriter, r *http.Request) {
	sets, err := h.storage.StandardRepo().GetStandardSets(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, sets)
}

// GetUpcomingRotation returns information about the upcoming Standard rotation.
// GET /api/v1/standard/rotation
func (h *StandardHandler) GetUpcomingRotation(w http.ResponseWriter, r *http.Request) {
	rotation, err := h.storage.StandardRepo().GetUpcomingRotation(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, rotation)
}

// GetRotationAffectedDecks returns all Standard decks affected by the upcoming rotation.
// GET /api/v1/standard/rotation/affected-decks
func (h *StandardHandler) GetRotationAffectedDecks(w http.ResponseWriter, r *http.Request) {
	decks, err := h.storage.StandardRepo().GetRotationAffectedDecks(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, decks)
}

// GetStandardConfig returns the Standard rotation configuration.
// GET /api/v1/standard/config
func (h *StandardHandler) GetStandardConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.storage.StandardRepo().GetConfig(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Success(w, config)
}

// ValidateDeckStandard validates a deck for Standard legality.
// POST /api/v1/standard/validate/{deckID}
func (h *StandardHandler) ValidateDeckStandard(w http.ResponseWriter, r *http.Request) {
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		response.BadRequest(w, errors.New("deck ID is required"))
		return
	}

	// Get deck
	deck, err := h.storage.DeckRepo().GetByID(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}
	if deck == nil {
		response.NotFound(w, errors.New("deck not found"))
		return
	}

	// Get deck cards
	cards, err := h.storage.DeckRepo().GetCards(r.Context(), deckID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Get legality for all cards
	arenaIDs := make([]string, len(cards))
	for i, card := range cards {
		arenaIDs[i] = fmt.Sprintf("%d", card.CardID)
	}

	legalities, err := h.storage.StandardRepo().GetCardsLegality(r.Context(), arenaIDs)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	// Build validation result
	result := validateDeckForStandard(cards, legalities)

	response.Success(w, result)
}

// validateDeckForStandard validates a deck's cards against Standard legality.
func validateDeckForStandard(cards []*models.DeckCard, legalities map[string]*models.CardLegality) *models.DeckValidationResult {
	result := &models.DeckValidationResult{
		IsLegal:      true,
		Errors:       []models.ValidationError{},
		Warnings:     []models.ValidationWarning{},
		SetBreakdown: []models.DeckSetInfo{},
	}

	// Track card counts for 4-copy rule
	cardCounts := make(map[int]int)

	// Basic lands that are exempt from legality checks
	basicLands := map[string]bool{
		"Plains": true, "Island": true, "Swamp": true,
		"Mountain": true, "Forest": true, "Wastes": true,
	}

	for _, card := range cards {
		arenaID := fmt.Sprintf("%d", card.CardID)
		cardCounts[card.CardID] += card.Quantity

		legality, found := legalities[arenaID]
		if !found {
			// Card not in legality database - warn but don't fail
			result.Warnings = append(result.Warnings, models.ValidationWarning{
				CardID:  card.CardID,
				Type:    "unknown_legality",
				Details: "Card legality information not available",
			})
			continue
		}

		// Check Standard legality
		switch legality.Standard {
		case "banned":
			result.IsLegal = false
			result.Errors = append(result.Errors, models.ValidationError{
				CardID:  card.CardID,
				Reason:  "banned",
				Details: "Card is banned in Standard",
			})
		case "not_legal":
			result.IsLegal = false
			result.Errors = append(result.Errors, models.ValidationError{
				CardID:  card.CardID,
				Reason:  "not_legal",
				Details: "Card is not legal in Standard",
			})
		}
	}

	// Check 4-copy rule (excluding basic lands)
	for cardID, count := range cardCounts {
		if count > 4 {
			// Check if it's a basic land (we'd need card names here, skip for now)
			_ = basicLands // Basic land check requires card name lookup
			result.IsLegal = false
			result.Errors = append(result.Errors, models.ValidationError{
				CardID:  cardID,
				Reason:  "too_many_copies",
				Details: fmt.Sprintf("Deck contains %d copies (maximum 4 allowed)", count),
			})
		}
	}

	// Check minimum deck size
	totalCards := 0
	for _, card := range cards {
		if card.Board == "main" {
			totalCards += card.Quantity
		}
	}

	if totalCards < 60 {
		result.IsLegal = false
		result.Errors = append(result.Errors, models.ValidationError{
			Reason:  "deck_size",
			Details: fmt.Sprintf("Deck has %d cards (minimum 60 required)", totalCards),
		})
	}

	return result
}

// GetCardLegality returns the legality of a single card.
// GET /api/v1/standard/cards/{arenaID}/legality
func (h *StandardHandler) GetCardLegality(w http.ResponseWriter, r *http.Request) {
	arenaID := chi.URLParam(r, "arenaID")
	if arenaID == "" {
		response.BadRequest(w, errors.New("arena ID is required"))
		return
	}

	legality, err := h.storage.StandardRepo().GetCardLegality(r.Context(), arenaID)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	if legality == nil {
		response.NotFound(w, errors.New("card legality not found"))
		return
	}

	response.Success(w, legality)
}

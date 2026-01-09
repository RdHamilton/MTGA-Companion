package gui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ramonehamilton/MTGA-Companion/internal/archetype"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckexport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/recommendations"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// DeckFacade handles all deck builder operations.
type DeckFacade struct {
	services *Services
}

// NewDeckFacade creates a new DeckFacade with the given services.
func NewDeckFacade(services *Services) *DeckFacade {
	return &DeckFacade{
		services: services,
	}
}

// DeckWithCards represents a deck with its associated cards.
type DeckWithCards struct {
	Deck  *models.Deck       `json:"deck"`
	Cards []*models.DeckCard `json:"cards"`
	Tags  []*models.DeckTag  `json:"tags,omitempty"`
}

// DeckListItem represents a summary of a deck for list views.
type DeckListItem struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Format           string     `json:"format"`
	Source           string     `json:"source"`
	ColorIdentity    *string    `json:"colorIdentity"`
	PrimaryArchetype *string    `json:"primaryArchetype,omitempty"` // Detected archetype (e.g., "Aggro", "Control")
	CardCount        int        `json:"cardCount"`
	MatchesPlayed    int        `json:"matchesPlayed"`
	MatchWinRate     float64    `json:"matchWinRate"`
	ModifiedAt       time.Time  `json:"modifiedAt"`
	LastPlayed       *time.Time `json:"lastPlayed,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	CurrentStreak    int        `json:"currentStreak"`   // Positive for wins, negative for losses
	AverageDuration  *float64   `json:"averageDuration"` // Average match duration in seconds
}

// normalizeDeckSource validates and normalizes deck source values.
// It accepts 'manual' as alias for 'constructed', and 'import' as alias for 'imported'.
func normalizeDeckSource(source string) (string, error) {
	switch source {
	case "manual":
		return "constructed", nil
	case "import":
		return "imported", nil
	case "draft", "constructed", "imported":
		return source, nil
	default:
		return "", fmt.Errorf("invalid deck source: %s", source)
	}
}

// CreateDeck creates a new deck.
func (d *DeckFacade) CreateDeck(ctx context.Context, name, format, source string, draftEventID *string) (*models.Deck, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get current account ID
	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	now := time.Now()
	// Validate and normalize source
	normalizedSource, err := normalizeDeckSource(source)
	if err != nil {
		return nil, &AppError{Message: err.Error()}
	}

	deck := &models.Deck{
		ID:            uuid.New().String(),
		AccountID:     accountID,
		Name:          name,
		Format:        format,
		Source:        normalizedSource,
		DraftEventID:  draftEventID,
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	// If source is draft, draft_event_id is required
	if normalizedSource == "draft" && draftEventID == nil {
		return nil, &AppError{Message: "Draft event ID required for draft decks"}
	}

	err = storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().Create(ctx, deck)
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to create deck: %v", err)}
	}

	log.Printf("Created deck %s (%s) with source: %s", deck.Name, deck.ID, deck.Source)
	return deck, nil
}

// GetDeck retrieves a deck by ID with its cards and tags.
func (d *DeckFacade) GetDeck(ctx context.Context, deckID string) (*DeckWithCards, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, deckID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck: %v", err)}
	}
	if deck == nil {
		return nil, &AppError{Message: "Deck not found"}
	}

	// Get cards
	var cards []*models.DeckCard
	err = storage.RetryOnBusy(func() error {
		var err error
		cards, err = d.services.Storage.DeckRepo().GetCards(ctx, deckID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck cards: %v", err)}
	}

	// Get tags
	var tags []*models.DeckTag
	err = storage.RetryOnBusy(func() error {
		var err error
		tags, err = d.services.Storage.DeckRepo().GetTags(ctx, deckID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck tags: %v", err)}
	}

	return &DeckWithCards{
		Deck:  deck,
		Cards: cards,
		Tags:  tags,
	}, nil
}

// ListDecks retrieves all decks for the current account.
func (d *DeckFacade) ListDecks(ctx context.Context) ([]*DeckListItem, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	var decks []*models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		decks, err = d.services.Storage.DeckRepo().List(ctx, accountID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to list decks: %v", err)}
	}

	// Convert to list items with card counts and tags
	items := make([]*DeckListItem, 0, len(decks))
	for _, deck := range decks {
		// Get card count
		var cards []*models.DeckCard
		err = storage.RetryOnBusy(func() error {
			var err error
			cards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
			return err
		})
		if err != nil {
			log.Printf("Warning: Failed to get cards for deck %s: %v", deck.ID, err)
		}

		// Count total cards (quantity)
		cardCount := 0
		for _, card := range cards {
			cardCount += card.Quantity
		}

		// Get tags
		var tags []*models.DeckTag
		err = storage.RetryOnBusy(func() error {
			var err error
			tags, err = d.services.Storage.DeckRepo().GetTags(ctx, deck.ID)
			return err
		})
		if err != nil {
			log.Printf("Warning: Failed to get tags for deck %s: %v", deck.ID, err)
		}

		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Tag
		}

		// Calculate win rate
		var winRate float64
		if deck.MatchesPlayed > 0 {
			winRate = float64(deck.MatchesWon) / float64(deck.MatchesPlayed)
		}

		// Get performance data for streak and duration
		var currentStreak int
		var avgDuration *float64
		perf, perfErr := d.services.Storage.DeckRepo().GetPerformance(ctx, deck.ID)
		if perfErr == nil && perf != nil {
			currentStreak = perf.CurrentWinStreak
			avgDuration = perf.AverageDuration
		}

		items = append(items, &DeckListItem{
			ID:              deck.ID,
			Name:            deck.Name,
			Format:          deck.Format,
			Source:          deck.Source,
			ColorIdentity:   deck.ColorIdentity,
			CardCount:       cardCount,
			MatchesPlayed:   deck.MatchesPlayed,
			MatchWinRate:    winRate,
			ModifiedAt:      deck.ModifiedAt,
			LastPlayed:      deck.LastPlayed,
			Tags:            tagNames,
			CurrentStreak:   currentStreak,
			AverageDuration: avgDuration,
		})
	}

	return items, nil
}

// GetDecksBySource retrieves decks filtered by source (draft/constructed/imported).
func (d *DeckFacade) GetDecksBySource(ctx context.Context, source string) ([]*DeckListItem, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	var decks []*models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		decks, err = d.services.Storage.DeckRepo().GetBySource(ctx, accountID, source)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get decks by source: %v", err)}
	}

	// Convert to list items (same as ListDecks)
	items := make([]*DeckListItem, 0, len(decks))
	for _, deck := range decks {
		var cards []*models.DeckCard
		err = storage.RetryOnBusy(func() error {
			var err error
			cards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
			return err
		})
		if err != nil {
			log.Printf("Warning: Failed to get cards for deck %s: %v", deck.ID, err)
			// Continue processing other decks
		}

		cardCount := 0
		for _, card := range cards {
			cardCount += card.Quantity
		}

		var winRate float64
		if deck.MatchesPlayed > 0 {
			winRate = float64(deck.MatchesWon) / float64(deck.MatchesPlayed)
		}

		// Get performance data for streak and duration
		var currentStreak int
		var avgDuration *float64
		perf, perfErr := d.services.Storage.DeckRepo().GetPerformance(ctx, deck.ID)
		if perfErr == nil && perf != nil {
			currentStreak = perf.CurrentWinStreak
			avgDuration = perf.AverageDuration
		}

		items = append(items, &DeckListItem{
			ID:              deck.ID,
			Name:            deck.Name,
			Format:          deck.Format,
			Source:          deck.Source,
			ColorIdentity:   deck.ColorIdentity,
			CardCount:       cardCount,
			MatchesPlayed:   deck.MatchesPlayed,
			MatchWinRate:    winRate,
			ModifiedAt:      deck.ModifiedAt,
			LastPlayed:      deck.LastPlayed,
			CurrentStreak:   currentStreak,
			AverageDuration: avgDuration,
		})
	}

	return items, nil
}

// UpdateDeck updates an existing deck's metadata.
func (d *DeckFacade) UpdateDeck(ctx context.Context, deck *models.Deck) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Update modified timestamp
	deck.ModifiedAt = time.Now()

	err := storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().Update(ctx, deck)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to update deck: %v", err)}
	}

	log.Printf("Updated deck %s (%s)", deck.Name, deck.ID)
	return nil
}

// DeleteDeck deletes a deck and all its cards.
func (d *DeckFacade) DeleteDeck(ctx context.Context, deckID string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	err := storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().Delete(ctx, deckID)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to delete deck: %v", err)}
	}

	log.Printf("Deleted deck %s", deckID)
	return nil
}

// AddCard adds a card to a deck.
func (d *DeckFacade) AddCard(ctx context.Context, deckID string, cardID, quantity int, board string, fromDraft bool) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Validate board
	if board != "main" && board != "sideboard" {
		return &AppError{Message: fmt.Sprintf("Invalid board: %s (must be 'main' or 'sideboard')", board)}
	}

	// Get the deck to check if it's a draft deck
	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, deckID)
		return err
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get deck: %v", err)}
	}
	if deck == nil {
		return &AppError{Message: "Deck not found"}
	}

	// Check if this is a basic land (basic lands have no quantity limits)
	basicLandIDs := map[int]bool{
		81716: true, // Plains
		81717: true, // Island
		81718: true, // Swamp
		81719: true, // Mountain
		81720: true, // Forest
	}

	// If this is a draft deck, validate that the card is from the draft
	// Exception: Basic lands are always allowed (they have unlimited availability)
	if deck.Source == "draft" && deck.DraftEventID != nil {
		if !basicLandIDs[cardID] {
			// Not a basic land, so validate it's in the draft pool
			var draftCards []int
			err = storage.RetryOnBusy(func() error {
				var err error
				draftCards, err = d.services.Storage.DeckRepo().GetDraftCards(ctx, *deck.DraftEventID)
				return err
			})
			if err != nil {
				return &AppError{Message: fmt.Sprintf("Failed to get draft cards: %v", err)}
			}

			// Check if card is in draft
			cardInDraft := false
			for _, draftCardID := range draftCards {
				if draftCardID == cardID {
					cardInDraft = true
					break
				}
			}

			if !cardInDraft {
				return &AppError{Message: "Card not in draft pool - draft decks can only contain cards from the associated draft"}
			}
		}
	}

	// Enforce 4-card limit (unless basic land)
	if !basicLandIDs[cardID] {
		var deckCards []*models.DeckCard
		err = storage.RetryOnBusy(func() error {
			var err error
			deckCards, err = d.services.Storage.DeckRepo().GetCards(ctx, deckID)
			return err
		})
		if err != nil {
			return &AppError{Message: fmt.Sprintf("Failed to get deck cards: %v", err)}
		}

		// Find current quantity of this card in the deck
		currentQty := 0
		for _, dc := range deckCards {
			if dc.CardID == cardID {
				currentQty += dc.Quantity
			}
		}

		// Check if adding would exceed 4-card limit
		if currentQty+quantity > 4 {
			maxCanAdd := 4 - currentQty
			if maxCanAdd <= 0 {
				return &AppError{Message: fmt.Sprintf("Cannot add more copies - card already at 4-copy limit (current: %d)", currentQty)}
			}
			return &AppError{Message: fmt.Sprintf("Cannot add %d copies - would exceed 4-copy limit (current: %d, max can add: %d)", quantity, currentQty, maxCanAdd)}
		}
	}

	card := &models.DeckCard{
		DeckID:        deckID,
		CardID:        cardID,
		Quantity:      quantity,
		Board:         board,
		FromDraftPick: fromDraft,
	}

	err = storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().AddCard(ctx, card)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to add card to deck: %v", err)}
	}

	// Update deck modified timestamp
	deck.ModifiedAt = time.Now()
	err = storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().Update(ctx, deck)
	})
	if err != nil {
		log.Printf("Warning: Failed to update deck modified timestamp: %v", err)
	}

	log.Printf("Added card %d (x%d) to deck %s", cardID, quantity, deckID)
	return nil
}

// RemoveCard removes a card from a deck.
func (d *DeckFacade) RemoveCard(ctx context.Context, deckID string, cardID int, board string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	err := storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().RemoveCard(ctx, deckID, cardID, board)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to remove card from deck: %v", err)}
	}

	// Update deck modified timestamp
	var deck *models.Deck
	err = storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, deckID)
		return err
	})
	if err == nil && deck != nil {
		deck.ModifiedAt = time.Now()
		_ = storage.RetryOnBusy(func() error {
			return d.services.Storage.DeckRepo().Update(ctx, deck)
		})
	}

	log.Printf("Removed 1 copy of card %d from deck %s", cardID, deckID)
	return nil
}

// RemoveAllCopies removes all copies of a card from a deck.
func (d *DeckFacade) RemoveAllCopies(ctx context.Context, deckID string, cardID int, board string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	err := storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().RemoveAllCopies(ctx, deckID, cardID, board)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to remove all copies from deck: %v", err)}
	}

	// Update deck modified timestamp
	var deck *models.Deck
	err = storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, deckID)
		return err
	})
	if err == nil && deck != nil {
		deck.ModifiedAt = time.Now()
		_ = storage.RetryOnBusy(func() error {
			return d.services.Storage.DeckRepo().Update(ctx, deck)
		})
	}

	log.Printf("Removed all copies of card %d from deck %s", cardID, deckID)
	return nil
}

// ValidateDraftDeck validates that all cards in a draft deck are from the associated draft.
func (d *DeckFacade) ValidateDraftDeck(ctx context.Context, deckID string) (bool, error) {
	if d.services.Storage == nil {
		return false, &AppError{Message: "Database not initialized"}
	}

	var isValid bool
	err := storage.RetryOnBusy(func() error {
		var err error
		isValid, err = d.services.Storage.DeckRepo().ValidateDraftDeck(ctx, deckID)
		return err
	})
	if err != nil {
		return false, &AppError{Message: fmt.Sprintf("Failed to validate deck: %v", err)}
	}

	return isValid, nil
}

// GetDeckPerformance retrieves performance metrics for a deck.
func (d *DeckFacade) GetDeckPerformance(ctx context.Context, deckID string) (*models.DeckPerformance, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var perf *models.DeckPerformance
	err := storage.RetryOnBusy(func() error {
		var err error
		perf, err = d.services.Storage.DeckRepo().GetPerformance(ctx, deckID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck performance: %v", err)}
	}

	return perf, nil
}

// AddTag adds a tag to a deck for categorization.
func (d *DeckFacade) AddTag(ctx context.Context, deckID, tag string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	deckTag := &models.DeckTag{
		DeckID:    deckID,
		Tag:       tag,
		CreatedAt: time.Now(),
	}

	err := storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().AddTag(ctx, deckTag)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to add tag to deck: %v", err)}
	}

	log.Printf("Added tag '%s' to deck %s", tag, deckID)
	return nil
}

// RemoveTag removes a tag from a deck.
func (d *DeckFacade) RemoveTag(ctx context.Context, deckID, tag string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	err := storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().RemoveTag(ctx, deckID, tag)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to remove tag from deck: %v", err)}
	}

	log.Printf("Removed tag '%s' from deck %s", tag, deckID)
	return nil
}

// GetDeckByDraftEvent retrieves the deck associated with a draft event.
func (d *DeckFacade) GetDeckByDraftEvent(ctx context.Context, draftEventID string) (*DeckWithCards, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByDraftEvent(ctx, draftEventID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck by draft event: %v", err)}
	}
	if deck == nil {
		return nil, nil // No deck for this draft yet
	}

	// Get cards and tags
	var cards []*models.DeckCard
	var tags []*models.DeckTag
	err = storage.RetryOnBusy(func() error {
		var err error
		cards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck cards: %v", err)}
	}

	err = storage.RetryOnBusy(func() error {
		var err error
		tags, err = d.services.Storage.DeckRepo().GetTags(ctx, deck.ID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck tags: %v", err)}
	}

	return &DeckWithCards{
		Deck:  deck,
		Cards: cards,
		Tags:  tags,
	}, nil
}

// ImportDeckRequest represents a request to import a deck from text.
type ImportDeckRequest struct {
	Name         string  `json:"name"`
	Format       string  `json:"format"`
	ImportText   string  `json:"importText"`
	Source       string  `json:"source"`       // "constructed" or "imported"
	DraftEventID *string `json:"draftEventID"` // Required if source is "draft"
}

// ImportDeckResponse contains the result of a deck import operation.
type ImportDeckResponse struct {
	Success       bool     `json:"success"`
	DeckID        string   `json:"deckID,omitempty"`
	Errors        []string `json:"errors,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
	CardsImported int      `json:"cardsImported"`
	CardsSkipped  int      `json:"cardsSkipped"`
}

// ImportDeck imports a deck from text (Arena format or plain text).
func (d *DeckFacade) ImportDeck(ctx context.Context, req *ImportDeckRequest) (*ImportDeckResponse, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if d.services.CardService == nil {
		return nil, &AppError{Message: "Card service not initialized"}
	}

	// Get current account ID
	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	response := &ImportDeckResponse{
		Success:       false,
		Errors:        make([]string, 0),
		Warnings:      make([]string, 0),
		CardsImported: 0,
		CardsSkipped:  0,
	}

	// Validate request
	if req.Name == "" {
		response.Errors = append(response.Errors, "Deck name is required")
		return response, nil
	}

	if req.ImportText == "" {
		response.Errors = append(response.Errors, "Import text is required")
		return response, nil
	}

	if req.Source != "constructed" && req.Source != "imported" && req.Source != "draft" {
		response.Errors = append(response.Errors, fmt.Sprintf("Invalid source: %s (must be 'constructed', 'imported', or 'draft')", req.Source))
		return response, nil
	}

	if req.Source == "draft" && req.DraftEventID == nil {
		response.Errors = append(response.Errors, "Draft event ID is required for draft imports")
		return response, nil
	}

	// Parse the import text
	parser := d.services.DeckImportParser
	if parser == nil {
		return nil, &AppError{Message: "Deck import parser not initialized"}
	}

	parseResult, err := parser.Parse(req.ImportText)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Failed to parse import: %v", err))
		return response, nil
	}

	// Add parse errors and warnings to response
	response.Errors = append(response.Errors, parseResult.Deck.Errors...)
	response.Warnings = append(response.Warnings, parseResult.Deck.Warnings...)
	response.Warnings = append(response.Warnings, parseResult.Warnings...)

	if !parseResult.Deck.ParsedOK {
		return response, nil
	}

	// If this is a draft import, validate against draft pool
	if req.Source == "draft" && req.DraftEventID != nil {
		var draftCardIDs []int
		err = storage.RetryOnBusy(func() error {
			var err error
			draftCardIDs, err = d.services.Storage.DeckRepo().GetDraftCards(ctx, *req.DraftEventID)
			return err
		})
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Failed to get draft cards: %v", err))
			return response, nil
		}

		draftErrors := parser.ValidateDraftImport(parseResult, draftCardIDs)
		for _, draftErr := range draftErrors {
			response.Errors = append(response.Errors, draftErr.Error())
		}

		if len(draftErrors) > 0 {
			return response, nil
		}
	}

	// Create the deck
	deck, err := d.CreateDeck(ctx, req.Name, req.Format, req.Source, req.DraftEventID)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Failed to create deck: %v", err))
		return response, nil
	}

	response.DeckID = deck.ID

	// Add mainboard cards
	for _, parsedCard := range parseResult.Deck.Mainboard {
		cardID, ok := parseResult.CardIDs[parsedCard.Name]
		if !ok {
			response.CardsSkipped++
			response.Warnings = append(response.Warnings, fmt.Sprintf("Skipping '%s': card not found in database", parsedCard.Name))
			continue
		}

		// Check if this is a draft card
		fromDraft := req.Source == "draft"

		err = d.AddCard(ctx, deck.ID, cardID, parsedCard.Quantity, "main", fromDraft)
		if err != nil {
			response.CardsSkipped++
			response.Warnings = append(response.Warnings, fmt.Sprintf("Failed to add '%s': %v", parsedCard.Name, err))
			continue
		}

		response.CardsImported++
	}

	// Add sideboard cards
	for _, parsedCard := range parseResult.Deck.Sideboard {
		cardID, ok := parseResult.CardIDs[parsedCard.Name]
		if !ok {
			response.CardsSkipped++
			response.Warnings = append(response.Warnings, fmt.Sprintf("Skipping '%s': card not found in database", parsedCard.Name))
			continue
		}

		// Check if this is a draft card
		fromDraft := req.Source == "draft"

		err = d.AddCard(ctx, deck.ID, cardID, parsedCard.Quantity, "sideboard", fromDraft)
		if err != nil {
			response.CardsSkipped++
			response.Warnings = append(response.Warnings, fmt.Sprintf("Failed to add '%s': %v", parsedCard.Name, err))
			continue
		}

		response.CardsImported++
	}

	response.Success = response.CardsImported > 0 && len(response.Errors) == 0
	log.Printf("Imported deck '%s' (%s): %d cards imported, %d skipped", req.Name, deck.ID, response.CardsImported, response.CardsSkipped)

	return response, nil
}

// GetRecommendationsRequest represents a request for card recommendations.
type GetRecommendationsRequest struct {
	DeckID        string   `json:"deckID"`
	MaxResults    int      `json:"maxResults,omitempty"`    // Default: 10
	MinScore      float64  `json:"minScore,omitempty"`      // Default: 0.3
	Colors        []string `json:"colors,omitempty"`        // Filter by colors
	CardTypes     []string `json:"cardTypes,omitempty"`     // Filter by card types
	CMCMin        *int     `json:"cmcMin,omitempty"`        // Min CMC
	CMCMax        *int     `json:"cmcMax,omitempty"`        // Max CMC
	IncludeLands  bool     `json:"includeLands"`            // Include land recommendations
	OnlyDraftPool bool     `json:"onlyDraftPool,omitempty"` // Only draft pool cards (for draft decks)
}

// GetRecommendationsResponse represents the response with card recommendations.
type GetRecommendationsResponse struct {
	Recommendations []*CardRecommendation `json:"recommendations"`
	Error           string                `json:"error,omitempty"`
}

// CardRecommendation represents a single card recommendation for the frontend.
type CardRecommendation struct {
	CardID     int           `json:"cardID"`
	Name       string        `json:"name"`
	TypeLine   string        `json:"typeLine"`
	ManaCost   string        `json:"manaCost,omitempty"`
	ImageURI   string        `json:"imageURI,omitempty"`
	Score      float64       `json:"score"`
	Reasoning  string        `json:"reasoning"`
	Source     string        `json:"source"`
	Confidence float64       `json:"confidence"`
	Factors    *ScoreFactors `json:"factors"`
}

// ScoreFactors breaks down the recommendation score components.
type ScoreFactors struct {
	ColorFit  float64 `json:"colorFit"`
	ManaCurve float64 `json:"manaCurve"`
	Synergy   float64 `json:"synergy"`
	Quality   float64 `json:"quality"`
	Playable  float64 `json:"playable"`
}

// GetRecommendations returns card recommendations for a deck.
func (d *DeckFacade) GetRecommendations(ctx context.Context, req *GetRecommendationsRequest) (*GetRecommendationsResponse, error) {
	if req.DeckID == "" {
		return nil, fmt.Errorf("deck ID is required")
	}

	// Get deck from database
	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, req.DeckID)
		return err
	})
	if err != nil {
		return &GetRecommendationsResponse{
			Error: fmt.Sprintf("Failed to get deck: %v", err),
		}, nil
	}

	// Get deck cards
	var deckCards []*models.DeckCard
	err = storage.RetryOnBusy(func() error {
		var err error
		deckCards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
		return err
	})
	if err != nil {
		return &GetRecommendationsResponse{
			Error: fmt.Sprintf("Failed to get deck cards: %v", err),
		}, nil
	}

	// Get card metadata for all cards in deck
	cardMetadata := make(map[int]*cards.Card)
	for _, deckCard := range deckCards {
		if _, exists := cardMetadata[deckCard.CardID]; !exists {
			// Try SetCardRepo first (faster, has cards from log parsing and datasets)
			setCard, err := d.services.Storage.SetCardRepo().GetCardByArenaID(ctx, fmt.Sprintf("%d", deckCard.CardID))
			if err == nil && setCard != nil {
				// Convert models.SetCard to cards.Card
				card := convertSetCardToCard(setCard)
				cardMetadata[deckCard.CardID] = card
				continue
			}

			// Fallback to CardService (Scryfall API)
			card, err := d.services.CardService.GetCard(deckCard.CardID)
			if err != nil {
				log.Printf("Warning: Failed to get card %d from both SetCardRepo and CardService: %v", deckCard.CardID, err)
				continue
			}
			cardMetadata[deckCard.CardID] = card
		}
	}

	// Build deck context for recommendations
	deckContext := &recommendations.DeckContext{
		Deck:         deck,
		Cards:        deckCards,
		CardMetadata: cardMetadata,
		Format:       deck.Format,
	}

	// For draft decks, get the draft pool and set/format info for ratings
	var draftPool []int
	if deck.DraftEventID != nil && req.OnlyDraftPool {
		// Get the draft session for set code and format
		session, err := d.services.Storage.DraftRepo().GetSession(ctx, *deck.DraftEventID)
		if err != nil {
			log.Printf("Warning: Failed to get draft session for deck %s: %v", deck.ID, err)
		} else {
			// Extract set code from EventName (e.g., "QuickDraft_BLB_20250101" -> "BLB")
			eventParts := strings.Split(session.EventName, "_")
			if len(eventParts) >= 2 {
				deckContext.SetCode = eventParts[1]
				// Determine draft format from event name
				if strings.HasPrefix(session.EventName, "PremierDraft") {
					deckContext.DraftFormat = "PremierDraft"
				} else if strings.HasPrefix(session.EventName, "QuickDraft") {
					deckContext.DraftFormat = "QuickDraft"
				} else {
					deckContext.DraftFormat = "PremierDraft" // Default
				}
				log.Printf("Info: Using set=%s, format=%s for ratings", deckContext.SetCode, deckContext.DraftFormat)
			}
		}

		// Get all cards from the draft session
		draftPool, err = d.services.Storage.DeckRepo().GetDraftCards(ctx, *deck.DraftEventID)
		if err != nil {
			log.Printf("Warning: Failed to get draft pool for deck %s: %v", deck.ID, err)
		} else {
			log.Printf("Info: Loaded draft pool for deck %s: %d cards", deck.ID, len(draftPool))
		}
	}

	// Build filters from request
	filters := &recommendations.Filters{
		MaxResults:    req.MaxResults,
		MinScore:      req.MinScore,
		Colors:        req.Colors,
		CardTypes:     req.CardTypes,
		IncludeLands:  req.IncludeLands,
		OnlyDraftPool: req.OnlyDraftPool,
		DraftPool:     draftPool,
	}

	// Set defaults
	if filters.MaxResults == 0 {
		filters.MaxResults = 10
	}
	if filters.MinScore == 0 {
		filters.MinScore = 0.3
	}

	// Set CMC range if provided
	if req.CMCMin != nil || req.CMCMax != nil {
		filters.CMCRange = &recommendations.CMCRange{}
		if req.CMCMin != nil {
			filters.CMCRange.Min = *req.CMCMin
		}
		if req.CMCMax != nil {
			filters.CMCRange.Max = *req.CMCMax
		}
	}

	// Get recommendations from engine
	engine := d.services.RecommendationEngine
	if engine == nil {
		return &GetRecommendationsResponse{
			Error: "Recommendation engine not available",
		}, nil
	}

	recs, err := engine.GetRecommendations(ctx, deckContext, filters)
	if err != nil {
		return &GetRecommendationsResponse{
			Error: fmt.Sprintf("Failed to get recommendations: %v", err),
		}, nil
	}

	// Convert to response format
	responseRecs := make([]*CardRecommendation, 0, len(recs))
	for _, rec := range recs {
		manaCost := ""
		if rec.Card.ManaCost != nil {
			manaCost = *rec.Card.ManaCost
		}
		imageURI := ""
		if rec.Card.ImageURI != nil {
			imageURI = *rec.Card.ImageURI
		}

		responseRecs = append(responseRecs, &CardRecommendation{
			CardID:     rec.Card.ArenaID,
			Name:       rec.Card.Name,
			TypeLine:   rec.Card.TypeLine,
			ManaCost:   manaCost,
			ImageURI:   imageURI,
			Score:      rec.Score,
			Reasoning:  rec.Reasoning,
			Source:     rec.Source,
			Confidence: rec.Confidence,
			Factors: &ScoreFactors{
				ColorFit:  rec.Factors.ColorFit,
				ManaCurve: rec.Factors.ManaCurve,
				Synergy:   rec.Factors.Synergy,
				Quality:   rec.Factors.Quality,
				Playable:  rec.Factors.Playable,
			},
		})
	}

	return &GetRecommendationsResponse{
		Recommendations: responseRecs,
	}, nil
}

// ExplainRecommendationRequest represents a request to explain a recommendation.
type ExplainRecommendationRequest struct {
	DeckID string `json:"deckID"`
	CardID int    `json:"cardID"`
}

// ExplainRecommendationResponse represents the response with the explanation.
type ExplainRecommendationResponse struct {
	Explanation string `json:"explanation"`
	Error       string `json:"error,omitempty"`
}

// ExplainRecommendation explains why a card is recommended for a deck.
func (d *DeckFacade) ExplainRecommendation(ctx context.Context, req *ExplainRecommendationRequest) (*ExplainRecommendationResponse, error) {
	if req.DeckID == "" || req.CardID == 0 {
		return nil, fmt.Errorf("deck ID and card ID are required")
	}

	// Get deck from database
	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, req.DeckID)
		return err
	})
	if err != nil {
		return &ExplainRecommendationResponse{
			Error: fmt.Sprintf("Failed to get deck: %v", err),
		}, nil
	}

	// Get deck cards
	var deckCards []*models.DeckCard
	err = storage.RetryOnBusy(func() error {
		var err error
		deckCards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
		return err
	})
	if err != nil {
		return &ExplainRecommendationResponse{
			Error: fmt.Sprintf("Failed to get deck cards: %v", err),
		}, nil
	}

	// Get card metadata for all cards in deck
	cardMetadata := make(map[int]*cards.Card)
	for _, deckCard := range deckCards {
		if _, exists := cardMetadata[deckCard.CardID]; !exists {
			card, err := d.services.CardService.GetCard(deckCard.CardID)
			if err != nil {
				log.Printf("Warning: Failed to get card %d: %v", deckCard.CardID, err)
				continue
			}
			cardMetadata[deckCard.CardID] = card
		}
	}

	// Build deck context
	deckContext := &recommendations.DeckContext{
		Deck:         deck,
		Cards:        deckCards,
		CardMetadata: cardMetadata,
		Format:       deck.Format,
	}

	// Get explanation from engine
	engine := d.services.RecommendationEngine
	if engine == nil {
		return &ExplainRecommendationResponse{
			Error: "Recommendation engine not available",
		}, nil
	}

	explanation, err := engine.ExplainRecommendation(ctx, req.CardID, deckContext)
	if err != nil {
		return &ExplainRecommendationResponse{
			Error: fmt.Sprintf("Failed to explain recommendation: %v", err),
		}, nil
	}

	return &ExplainRecommendationResponse{
		Explanation: explanation,
	}, nil
}

// ExportDeckRequest represents a request to export a deck.
type ExportDeckRequest struct {
	DeckID         string `json:"deckID"`
	Format         string `json:"format"`         // "arena", "plaintext", "mtgo", "mtggoldfish"
	IncludeHeaders bool   `json:"includeHeaders"` // Include section headers
	IncludeStats   bool   `json:"includeStats"`   // Include deck statistics as comments
}

// ExportDeckResponse represents the exported deck data.
type ExportDeckResponse struct {
	Content  string `json:"content"`         // The exported deck text
	Filename string `json:"filename"`        // Suggested filename
	Format   string `json:"format"`          // The format used
	Error    string `json:"error,omitempty"` // Error message if failed
}

// CloneDeck creates a copy of an existing deck.
func (d *DeckFacade) CloneDeck(ctx context.Context, deckID, newName string) (*models.Deck, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if deckID == "" {
		return nil, &AppError{Message: "Deck ID is required"}
	}

	if newName == "" {
		return nil, &AppError{Message: "New deck name is required"}
	}

	var clonedDeck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		clonedDeck, err = d.services.Storage.DeckRepo().Clone(ctx, deckID, newName)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to clone deck: %v", err)}
	}

	log.Printf("Cloned deck %s to %s (%s)", deckID, newName, clonedDeck.ID)
	return clonedDeck, nil
}

// GetDecksByFormat retrieves decks filtered by format (Standard, Historic, etc.).
func (d *DeckFacade) GetDecksByFormat(ctx context.Context, format string) ([]*DeckListItem, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	var decks []*models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		decks, err = d.services.Storage.DeckRepo().GetByFormat(ctx, accountID, format)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get decks by format: %v", err)}
	}

	return d.convertToDeckListItems(ctx, decks)
}

// GetDecksByTags retrieves decks that have ALL specified tags.
func (d *DeckFacade) GetDecksByTags(ctx context.Context, tags []string) ([]*DeckListItem, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	var decks []*models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		decks, err = d.services.Storage.DeckRepo().GetByTags(ctx, accountID, tags)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get decks by tags: %v", err)}
	}

	return d.convertToDeckListItems(ctx, decks)
}

// DeckLibraryFilter represents filter options for the deck library.
type DeckLibraryFilter struct {
	Format   *string  `json:"format,omitempty"`   // Filter by format
	Source   *string  `json:"source,omitempty"`   // Filter by source
	Tags     []string `json:"tags,omitempty"`     // Filter by tags (must have ALL)
	SortBy   string   `json:"sortBy,omitempty"`   // Sort field: "modified", "created", "name", "performance"
	SortDesc bool     `json:"sortDesc,omitempty"` // Sort descending
}

// GetDeckLibrary retrieves all decks with advanced filtering and sorting.
func (d *DeckFacade) GetDeckLibrary(ctx context.Context, filter *DeckLibraryFilter) ([]*DeckListItem, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	accountID := d.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	// Convert to repository filter
	repoFilter := &repository.DeckFilter{
		AccountID: accountID,
		SortDesc:  true, // Default to descending
	}

	if filter != nil {
		repoFilter.Format = filter.Format
		repoFilter.Source = filter.Source
		repoFilter.Tags = filter.Tags
		repoFilter.SortBy = filter.SortBy
		repoFilter.SortDesc = filter.SortDesc
	}

	var decks []*models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		decks, err = d.services.Storage.DeckRepo().GetByFilters(ctx, repoFilter)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck library: %v", err)}
	}

	return d.convertToDeckListItems(ctx, decks)
}

// convertToDeckListItems is a helper to convert decks to DeckListItem format.
func (d *DeckFacade) convertToDeckListItems(ctx context.Context, decks []*models.Deck) ([]*DeckListItem, error) {
	items := make([]*DeckListItem, 0, len(decks))

	// Create archetype classifier for deck classification
	var classifier *archetype.Classifier
	if d.services.CardService != nil && d.services.Storage != nil {
		classifier = archetype.NewClassifier(
			d.services.CardService,
			d.services.Storage.DeckRepo(),
			d.services.Storage.DeckPerformanceRepo(),
		)
	}

	for _, deck := range decks {
		// Get card count
		var cards []*models.DeckCard
		err := storage.RetryOnBusy(func() error {
			var err error
			cards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
			return err
		})
		if err != nil {
			log.Printf("Warning: Failed to get cards for deck %s: %v", deck.ID, err)
		}

		// Count total cards (quantity)
		cardCount := 0
		for _, card := range cards {
			cardCount += card.Quantity
		}

		// Get tags
		var tags []*models.DeckTag
		err = storage.RetryOnBusy(func() error {
			var err error
			tags, err = d.services.Storage.DeckRepo().GetTags(ctx, deck.ID)
			return err
		})
		if err != nil {
			log.Printf("Warning: Failed to get tags for deck %s: %v", deck.ID, err)
		}

		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Tag
		}

		// Calculate win rate
		var winRate float64
		if deck.MatchesPlayed > 0 {
			winRate = float64(deck.MatchesWon) / float64(deck.MatchesPlayed)
		}

		// Classify deck archetype if classifier is available and deck has cards
		var primaryArchetype *string
		if classifier != nil && len(cards) > 0 {
			result, classErr := classifier.ClassifyDeck(ctx, deck.ID)
			if classErr == nil && result != nil && result.PrimaryArchetype != "" {
				primaryArchetype = &result.PrimaryArchetype
			}
		}

		items = append(items, &DeckListItem{
			ID:               deck.ID,
			Name:             deck.Name,
			Format:           deck.Format,
			Source:           deck.Source,
			ColorIdentity:    deck.ColorIdentity,
			PrimaryArchetype: primaryArchetype,
			CardCount:        cardCount,
			MatchesPlayed:    deck.MatchesPlayed,
			MatchWinRate:     winRate,
			ModifiedAt:       deck.ModifiedAt,
			LastPlayed:       deck.LastPlayed,
			Tags:             tagNames,
		})
	}

	return items, nil
}

// DeckStatistics represents comprehensive deck statistics and analysis.
type DeckStatistics struct {
	// Basic counts
	TotalCards     int     `json:"totalCards"`
	TotalMainboard int     `json:"totalMainboard"`
	TotalSideboard int     `json:"totalSideboard"`
	AverageCMC     float64 `json:"averageCMC"`

	// Mana curve (CMC -> count)
	ManaCurve map[int]int `json:"manaCurve"`
	MaxCMC    int         `json:"maxCMC"`

	// Color distribution
	Colors ColorStats `json:"colors"`

	// Type breakdown
	Types TypeStats `json:"types"`

	// Land analysis
	Lands LandStats `json:"lands"`

	// Creature statistics
	Creatures CreatureStats `json:"creatures"`

	// Format legality
	Legality FormatLegality `json:"legality"`
}

// ColorStats represents color distribution in the deck.
type ColorStats struct {
	White      int `json:"white"`
	Blue       int `json:"blue"`
	Black      int `json:"black"`
	Red        int `json:"red"`
	Green      int `json:"green"`
	Colorless  int `json:"colorless"`
	Multicolor int `json:"multicolor"`
}

// TypeStats represents card type breakdown.
type TypeStats struct {
	Creatures     int `json:"creatures"`
	Instants      int `json:"instants"`
	Sorceries     int `json:"sorceries"`
	Enchantments  int `json:"enchantments"`
	Artifacts     int `json:"artifacts"`
	Planeswalkers int `json:"planeswalkers"`
	Lands         int `json:"lands"`
	Other         int `json:"other"`
}

// LandStats represents land analysis and recommendations.
type LandStats struct {
	Total         int     `json:"total"`
	Basic         int     `json:"basic"`
	NonBasic      int     `json:"nonBasic"`
	Ratio         float64 `json:"ratio"`         // Percentage of deck
	Recommended   int     `json:"recommended"`   // Recommended land count
	Status        string  `json:"status"`        // "optimal", "too_few", "too_many"
	StatusMessage string  `json:"statusMessage"` // Human-readable message
}

// CreatureStats represents creature-specific statistics.
type CreatureStats struct {
	Total            int     `json:"total"`
	AveragePower     float64 `json:"averagePower"`
	AverageToughness float64 `json:"averageToughness"`
	TotalPower       int     `json:"totalPower"`
	TotalToughness   int     `json:"totalToughness"`
}

// FormatLegality represents deck legality in various formats.
type FormatLegality struct {
	Standard  LegalityStatus `json:"standard"`
	Historic  LegalityStatus `json:"historic"`
	Explorer  LegalityStatus `json:"explorer"`
	Alchemy   LegalityStatus `json:"alchemy"`
	Brawl     LegalityStatus `json:"brawl"`
	Commander LegalityStatus `json:"commander"`
}

// LegalityStatus represents legality status for a format.
type LegalityStatus struct {
	Legal   bool     `json:"legal"`
	Reasons []string `json:"reasons,omitempty"` // Why it's not legal
}

// GetDeckStatistics calculates comprehensive deck statistics.
func (d *DeckFacade) GetDeckStatistics(ctx context.Context, deckID string) (*DeckStatistics, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get deck and cards
	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, deckID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck: %v", err)}
	}
	if deck == nil {
		return nil, &AppError{Message: "Deck not found"}
	}

	var deckCards []*models.DeckCard
	err = storage.RetryOnBusy(func() error {
		var err error
		deckCards, err = d.services.Storage.DeckRepo().GetCards(ctx, deckID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck cards: %v", err)}
	}

	stats := &DeckStatistics{
		ManaCurve: make(map[int]int),
	}

	// Separate mainboard and sideboard
	var mainboard, sideboard []*models.DeckCard
	for _, card := range deckCards {
		switch card.Board {
		case "main":
			mainboard = append(mainboard, card)
		case "sideboard":
			sideboard = append(sideboard, card)
		}
	}

	// Calculate statistics from mainboard
	stats = d.calculateDeckStats(ctx, mainboard, stats)

	// Set totals
	stats.TotalMainboard = stats.TotalCards
	for _, card := range sideboard {
		stats.TotalSideboard += card.Quantity
	}
	stats.TotalCards = stats.TotalMainboard + stats.TotalSideboard

	// Calculate land recommendations
	d.calculateLandRecommendations(stats, deck.Format)

	// Check format legality
	stats.Legality = d.checkFormatLegality(ctx, mainboard, deck.Format)

	return stats, nil
}

// calculateDeckStats performs the core statistical calculations.
func (d *DeckFacade) calculateDeckStats(ctx context.Context, deckCards []*models.DeckCard, stats *DeckStatistics) *DeckStatistics {
	totalCMC := 0.0
	nonLandCount := 0
	totalCreaturePower := 0
	totalCreatureToughness := 0
	creatureCountForAvg := 0

	basicLands := map[string]bool{
		"Plains": true, "Island": true, "Swamp": true, "Mountain": true, "Forest": true, "Wastes": true,
	}

	// Basic land IDs (for when metadata is unavailable)
	basicLandIDs := map[int]string{
		81716: "Plains",
		81717: "Island",
		81718: "Swamp",
		81719: "Mountain",
		81720: "Forest",
	}

	for _, deckCard := range deckCards {
		quantity := deckCard.Quantity

		// Check if this is a basic land by ID (handle even without metadata)
		if _, isBasicLand := basicLandIDs[deckCard.CardID]; isBasicLand {
			stats.TotalCards += quantity
			stats.Lands.Total += quantity
			stats.Lands.Basic += quantity
			stats.Types.Lands += quantity
			stats.ManaCurve[0] += quantity // Basic lands have CMC 0
			continue
		}

		// Get card metadata for non-basic-land cards
		// Try SetCardRepo first (faster, has cards from log parsing and datasets)
		setCard, setErr := d.services.Storage.SetCardRepo().GetCardByArenaID(ctx, fmt.Sprintf("%d", deckCard.CardID))
		var card *cards.Card
		if setErr == nil && setCard != nil {
			card = convertSetCardToCard(setCard)
		} else {
			// Fallback to CardService (Scryfall API)
			var err error
			card, err = d.services.CardService.GetCard(deckCard.CardID)
			if err != nil || card == nil {
				log.Printf("Warning: Failed to get card metadata for card ID %d: %v", deckCard.CardID, err)
				continue
			}
		}

		stats.TotalCards += quantity

		// Mana curve
		cmc := int(card.CMC)
		stats.ManaCurve[cmc] += quantity
		if cmc > stats.MaxCMC {
			stats.MaxCMC = cmc
		}

		// Color distribution
		d.analyzeCardColors(card, quantity, &stats.Colors)

		// Type breakdown
		isLand := d.analyzeCardTypes(card, quantity, &stats.Types)

		// Land analysis
		if isLand {
			stats.Lands.Total += quantity
			if basicLands[card.Name] {
				stats.Lands.Basic += quantity
			} else {
				stats.Lands.NonBasic += quantity
			}
		} else {
			// Calculate average CMC for non-lands
			totalCMC += card.CMC * float64(quantity)
			nonLandCount += quantity
		}

		// Creature statistics
		if strings.Contains(strings.ToLower(card.TypeLine), "creature") {
			// Parse power and toughness
			if card.Power != nil && card.Toughness != nil {
				power := d.parsePowerToughness(*card.Power)
				toughness := d.parsePowerToughness(*card.Toughness)

				totalCreaturePower += power * quantity
				totalCreatureToughness += toughness * quantity
				creatureCountForAvg += quantity
			}
		}
	}

	// Calculate averages
	if nonLandCount > 0 {
		stats.AverageCMC = totalCMC / float64(nonLandCount)
	}

	if stats.TotalCards > 0 {
		stats.Lands.Ratio = float64(stats.Lands.Total) / float64(stats.TotalCards) * 100
	}

	// Creature stats
	stats.Creatures.Total = stats.Types.Creatures
	stats.Creatures.TotalPower = totalCreaturePower
	stats.Creatures.TotalToughness = totalCreatureToughness
	if creatureCountForAvg > 0 {
		stats.Creatures.AveragePower = float64(totalCreaturePower) / float64(creatureCountForAvg)
		stats.Creatures.AverageToughness = float64(totalCreatureToughness) / float64(creatureCountForAvg)
	}

	return stats
}

// analyzeCardColors updates color statistics.
func (d *DeckFacade) analyzeCardColors(card *cards.Card, quantity int, colors *ColorStats) {
	if len(card.Colors) == 0 {
		colors.Colorless += quantity
		return
	}

	if len(card.Colors) > 1 {
		colors.Multicolor += quantity
		return
	}

	// Single color
	switch card.Colors[0] {
	case "W":
		colors.White += quantity
	case "U":
		colors.Blue += quantity
	case "B":
		colors.Black += quantity
	case "R":
		colors.Red += quantity
	case "G":
		colors.Green += quantity
	}
}

// analyzeCardTypes updates type statistics and returns true if it's a land.
func (d *DeckFacade) analyzeCardTypes(card *cards.Card, quantity int, types *TypeStats) bool {
	typeLine := strings.ToLower(card.TypeLine)

	if strings.Contains(typeLine, "land") {
		types.Lands += quantity
		return true
	} else if strings.Contains(typeLine, "creature") {
		types.Creatures += quantity
	} else if strings.Contains(typeLine, "planeswalker") {
		types.Planeswalkers += quantity
	} else if strings.Contains(typeLine, "instant") {
		types.Instants += quantity
	} else if strings.Contains(typeLine, "sorcery") {
		types.Sorceries += quantity
	} else if strings.Contains(typeLine, "enchantment") {
		types.Enchantments += quantity
	} else if strings.Contains(typeLine, "artifact") {
		types.Artifacts += quantity
	} else {
		types.Other += quantity
	}

	return false
}

// parsePowerToughness parses power/toughness strings (* becomes 0).
func (d *DeckFacade) parsePowerToughness(value string) int {
	if value == "*" || value == "" {
		return 0
	}

	var result int
	_, _ = fmt.Sscanf(value, "%d", &result)
	return result
}

// calculateLandRecommendations provides land count recommendations.
func (d *DeckFacade) calculateLandRecommendations(stats *DeckStatistics, format string) {
	deckSize := stats.TotalMainboard
	avgCMC := stats.AverageCMC

	// Standard deck sizes and recommendations
	var recommendedLands int

	if deckSize >= 99 {
		// Commander/Brawl (100 cards)
		recommendedLands = 37 + int((avgCMC-2.5)*2)
	} else if deckSize >= 60 {
		// Standard 60-card deck
		// Base: 24 lands for avg CMC ~2.5
		// Adjust based on curve
		recommendedLands = 24 + int((avgCMC-2.5)*2)
	} else {
		// Limited (40 cards)
		recommendedLands = 17 + int((avgCMC-2.5)*1.5)
	}

	// Clamp to reasonable ranges
	if deckSize >= 99 {
		if recommendedLands < 33 {
			recommendedLands = 33
		} else if recommendedLands > 42 {
			recommendedLands = 42
		}
	} else if deckSize >= 60 {
		if recommendedLands < 20 {
			recommendedLands = 20
		} else if recommendedLands > 28 {
			recommendedLands = 28
		}
	} else {
		if recommendedLands < 15 {
			recommendedLands = 15
		} else if recommendedLands > 19 {
			recommendedLands = 19
		}
	}

	stats.Lands.Recommended = recommendedLands

	difference := stats.Lands.Total - recommendedLands

	if difference >= -1 && difference <= 1 {
		stats.Lands.Status = "optimal"
		stats.Lands.StatusMessage = "Land count is optimal for your deck"
	} else if difference < -1 {
		stats.Lands.Status = "too_few"
		missing := -difference
		stats.Lands.StatusMessage = fmt.Sprintf("Consider adding %d more land%s (currently %d, recommended %d)",
			missing, pluralize(missing), stats.Lands.Total, recommendedLands)
	} else {
		stats.Lands.Status = "too_many"
		extra := difference
		stats.Lands.StatusMessage = fmt.Sprintf("Consider removing %d land%s (currently %d, recommended %d)",
			extra, pluralize(extra), stats.Lands.Total, recommendedLands)
	}
}

// pluralize returns "s" if count != 1.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// checkFormatLegality checks deck legality in various formats.
func (d *DeckFacade) checkFormatLegality(ctx context.Context, deckCards []*models.DeckCard, deckFormat string) FormatLegality {
	legality := FormatLegality{
		Standard:  LegalityStatus{Legal: true},
		Historic:  LegalityStatus{Legal: true},
		Explorer:  LegalityStatus{Legal: true},
		Alchemy:   LegalityStatus{Legal: true},
		Brawl:     LegalityStatus{Legal: true},
		Commander: LegalityStatus{Legal: true},
	}

	// Count total mainboard cards
	totalCards := 0
	cardCounts := make(map[int]int)

	for _, deckCard := range deckCards {
		totalCards += deckCard.Quantity
		cardCounts[deckCard.CardID] += deckCard.Quantity
	}

	// Check minimum deck size
	if totalCards < 60 && deckFormat != "Brawl" && deckFormat != "Limited" {
		reason := fmt.Sprintf("Deck has only %d cards (minimum 60 for constructed)", totalCards)
		legality.Standard.Legal = false
		legality.Standard.Reasons = append(legality.Standard.Reasons, reason)
		legality.Historic.Legal = false
		legality.Historic.Reasons = append(legality.Historic.Reasons, reason)
		legality.Explorer.Legal = false
		legality.Explorer.Reasons = append(legality.Explorer.Reasons, reason)
		legality.Alchemy.Legal = false
		legality.Alchemy.Reasons = append(legality.Alchemy.Reasons, reason)
	}

	// Check for duplicates (max 4 copies, except basic lands)
	basicLands := map[string]bool{
		"Plains": true, "Island": true, "Swamp": true, "Mountain": true, "Forest": true, "Wastes": true,
	}

	for cardID, count := range cardCounts {
		if count > 4 {
			card, err := d.services.CardService.GetCard(cardID)
			if err == nil && card != nil && !basicLands[card.Name] {
				reason := fmt.Sprintf("Card '%s' has %d copies (maximum 4)", card.Name, count)
				legality.Standard.Legal = false
				legality.Standard.Reasons = append(legality.Standard.Reasons, reason)
				legality.Historic.Legal = false
				legality.Historic.Reasons = append(legality.Historic.Reasons, reason)
				legality.Explorer.Legal = false
				legality.Explorer.Reasons = append(legality.Explorer.Reasons, reason)
				legality.Alchemy.Legal = false
				legality.Alchemy.Reasons = append(legality.Alchemy.Reasons, reason)
			}
		}
	}

	// Commander/Brawl specific checks
	if deckFormat == "Brawl" || deckFormat == "Commander" {
		if deckFormat == "Brawl" && totalCards != 60 {
			legality.Brawl.Legal = false
			legality.Brawl.Reasons = append(legality.Brawl.Reasons,
				fmt.Sprintf("Brawl decks must have exactly 60 cards (currently %d)", totalCards))
		}
		if deckFormat == "Commander" && totalCards != 99 {
			legality.Commander.Legal = false
			legality.Commander.Reasons = append(legality.Commander.Reasons,
				fmt.Sprintf("Commander decks must have exactly 99 cards plus commander (currently %d)", totalCards))
		}

		// Check singleton (max 1 copy except basic lands)
		for cardID, count := range cardCounts {
			if count > 1 {
				card, err := d.services.CardService.GetCard(cardID)
				if err == nil && card != nil && !basicLands[card.Name] {
					reason := fmt.Sprintf("Card '%s' has %d copies (singleton format allows only 1)", card.Name, count)
					if deckFormat == "Brawl" {
						legality.Brawl.Legal = false
						legality.Brawl.Reasons = append(legality.Brawl.Reasons, reason)
					}
					if deckFormat == "Commander" {
						legality.Commander.Legal = false
						legality.Commander.Reasons = append(legality.Commander.Reasons, reason)
					}
				}
			}
		}
	}

	return legality
}

// ExportDeck exports a deck to the requested format.
func (d *DeckFacade) ExportDeck(ctx context.Context, req *ExportDeckRequest) (*ExportDeckResponse, error) {
	if req.DeckID == "" {
		return nil, fmt.Errorf("deck ID is required")
	}
	if req.Format == "" {
		req.Format = "arena" // Default to Arena format
	}

	// Get deck from database
	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, req.DeckID)
		return err
	})
	if err != nil {
		return &ExportDeckResponse{
			Error: fmt.Sprintf("Failed to get deck: %v", err),
		}, nil
	}

	// Get deck cards
	var deckCards []*models.DeckCard
	err = storage.RetryOnBusy(func() error {
		var err error
		deckCards, err = d.services.Storage.DeckRepo().GetCards(ctx, deck.ID)
		return err
	})
	if err != nil {
		return &ExportDeckResponse{
			Error: fmt.Sprintf("Failed to get deck cards: %v", err),
		}, nil
	}

	// Convert format string to ExportFormat
	var exportFormat deckexport.ExportFormat
	switch req.Format {
	case "arena":
		exportFormat = deckexport.FormatArena
	case "plaintext":
		exportFormat = deckexport.FormatPlainText
	case "mtgo":
		exportFormat = deckexport.FormatMTGO
	case "mtggoldfish":
		exportFormat = deckexport.FormatMTGGoldfish
	case "moxfield":
		exportFormat = deckexport.FormatMoxfield
	case "archidekt":
		exportFormat = deckexport.FormatArchidekt
	default:
		return &ExportDeckResponse{
			Error: fmt.Sprintf("Unsupported export format: %s", req.Format),
		}, nil
	}

	// Export the deck
	exporter := d.services.DeckExporter
	if exporter == nil {
		return &ExportDeckResponse{
			Error: "Deck exporter not available",
		}, nil
	}

	options := &deckexport.ExportOptions{
		Format:         exportFormat,
		IncludeHeaders: req.IncludeHeaders,
		IncludeStats:   req.IncludeStats,
	}

	result, err := exporter.Export(deck, deckCards, options)
	if err != nil {
		return &ExportDeckResponse{
			Error: fmt.Sprintf("Failed to export deck: %v", err),
		}, nil
	}

	return &ExportDeckResponse{
		Content:  result.Content,
		Filename: result.Filename,
		Format:   string(result.Format),
	}, nil
}

// ExportDeckToFile exports a deck to the specified file path.
// If filePath is empty, returns an error. Use ExportDeck to get the content
// and let the frontend handle file saving via browser download.
func (d *DeckFacade) ExportDeckToFile(ctx context.Context, deckID string, filePath string) error {
	log.Printf("ExportDeckToFile called with deckID: %s, filePath: %s", deckID, filePath)

	if deckID == "" {
		return fmt.Errorf("deck ID is required")
	}
	if filePath == "" {
		return fmt.Errorf("file path is required")
	}

	// Export deck content
	req := &ExportDeckRequest{
		DeckID:         deckID,
		Format:         "arena",
		IncludeHeaders: true,
		IncludeStats:   false,
	}

	log.Printf("Calling ExportDeck...")
	response, err := d.ExportDeck(ctx, req)
	if err != nil {
		log.Printf("ExportDeck returned error: %v", err)
		return fmt.Errorf("failed to export deck: %v", err)
	}

	if response.Error != "" {
		log.Printf("Response has error: %s", response.Error)
		return fmt.Errorf("%s", response.Error)
	}
	log.Printf("Export content length: %d bytes", len(response.Content))

	// Save the file
	err = os.WriteFile(filePath, []byte(response.Content), 0o644)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	log.Printf("Successfully exported deck to %s", filepath.Base(filePath))
	return nil
}

// convertSetCardToCard converts a models.SetCard to a cards.Card.
// This allows us to use SetCardRepo data in the recommendation engine.
func convertSetCardToCard(setCard *models.SetCard) *cards.Card {
	if setCard == nil {
		return nil
	}

	// Parse ArenaID from string to int
	arenaID := 0
	_, _ = fmt.Sscanf(setCard.ArenaID, "%d", &arenaID)

	// Build TypeLine from Types array
	typeLine := ""
	if len(setCard.Types) > 0 {
		typeLine = setCard.Types[0]
		for i := 1; i < len(setCard.Types); i++ {
			typeLine += " " + setCard.Types[i]
		}
	}

	card := &cards.Card{
		ArenaID:    arenaID,
		ScryfallID: setCard.ScryfallID,
		Name:       setCard.Name,
		TypeLine:   typeLine,
		SetCode:    setCard.SetCode,
		CMC:        float64(setCard.CMC),
		Colors:     setCard.Colors,
		Rarity:     setCard.Rarity,
	}

	// Convert string fields to *string where needed
	if setCard.ManaCost != "" {
		card.ManaCost = &setCard.ManaCost
	}
	if setCard.Power != "" {
		card.Power = &setCard.Power
	}
	if setCard.Toughness != "" {
		card.Toughness = &setCard.Toughness
	}
	if setCard.Text != "" {
		card.OracleText = &setCard.Text
	}
	if setCard.ImageURL != "" {
		card.ImageURI = &setCard.ImageURL
	}

	return card
}

// ArchetypeClassificationResult represents the result of classifying a deck archetype.
type ArchetypeClassificationResult struct {
	PrimaryArchetype   string                   `json:"primaryArchetype"`
	SecondaryArchetype *string                  `json:"secondaryArchetype,omitempty"`
	Confidence         float64                  `json:"confidence"`        // 0.0-1.0
	ConfidencePercent  int                      `json:"confidencePercent"` // 0-100
	ColorIdentity      string                   `json:"colorIdentity"`
	DominantColors     []string                 `json:"dominantColors"`
	ColorPair          *ColorPairInfo           `json:"colorPair,omitempty"`
	SignatureCards     []int                    `json:"signatureCards"`
	Indicators         []ArchetypeIndicatorInfo `json:"indicators"`
	TotalCards         int                      `json:"totalCards"`
	Analysis           *DeckArchetypeAnalysis   `json:"analysis"`
}

// ColorPairInfo represents a detected color pair.
type ColorPairInfo struct {
	Colors string `json:"colors"` // e.g., "WU"
	Name   string `json:"name"`   // e.g., "Azorius"
}

// ArchetypeIndicatorInfo represents a card that indicates a specific archetype.
type ArchetypeIndicatorInfo struct {
	CardID   int     `json:"cardID"`
	CardName string  `json:"cardName"`
	Weight   float64 `json:"weight"`
	Reason   string  `json:"reason"`
}

// DeckArchetypeAnalysis provides detailed breakdown of deck composition.
type DeckArchetypeAnalysis struct {
	ColorCounts       map[string]int `json:"colorCounts"`
	ColorlessCount    int            `json:"colorlessCount"`
	GoldCount         int            `json:"goldCount"`
	CreatureCount     int            `json:"creatureCount"`
	InstantCount      int            `json:"instantCount"`
	SorceryCount      int            `json:"sorceryCount"`
	ArtifactCount     int            `json:"artifactCount"`
	EnchantmentCount  int            `json:"enchantmentCount"`
	PlaneswalkerCount int            `json:"planeswalkerCount"`
	LandCount         int            `json:"landCount"`
	ManaCurve         map[int]int    `json:"manaCurve"`
	AvgCMC            float64        `json:"avgCMC"`
	RareCounts        map[string]int `json:"rareCounts"`
}

// ClassifyDeckArchetype classifies a deck into its archetype.
func (d *DeckFacade) ClassifyDeckArchetype(ctx context.Context, deckID string) (*ArchetypeClassificationResult, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if d.services.CardService == nil {
		return nil, &AppError{Message: "Card service not initialized"}
	}

	// Create classifier
	classifier := archetype.NewClassifier(
		d.services.CardService,
		d.services.Storage.DeckRepo(),
		d.services.Storage.DeckPerformanceRepo(),
	)

	// Classify the deck
	result, err := classifier.ClassifyDeck(ctx, deckID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to classify deck: %v", err)}
	}

	// Convert to response format
	response := &ArchetypeClassificationResult{
		PrimaryArchetype:   result.PrimaryArchetype,
		SecondaryArchetype: result.SecondaryArchetype,
		Confidence:         result.Confidence,
		ConfidencePercent:  int(result.Confidence * 100),
		ColorIdentity:      result.ColorIdentity,
		DominantColors:     result.DominantColors,
		SignatureCards:     result.SignatureCards,
		TotalCards:         result.TotalCards,
	}

	// Convert color pair
	if result.ColorPair != nil {
		response.ColorPair = &ColorPairInfo{
			Colors: result.ColorPair.Colors,
			Name:   result.ColorPair.Name,
		}
	}

	// Convert indicators
	response.Indicators = make([]ArchetypeIndicatorInfo, len(result.ArchetypeIndicators))
	for i, ind := range result.ArchetypeIndicators {
		response.Indicators[i] = ArchetypeIndicatorInfo{
			CardID:   ind.CardID,
			CardName: ind.CardName,
			Weight:   ind.Weight,
			Reason:   ind.Reason,
		}
	}

	// Convert analysis
	if result.Analysis != nil {
		response.Analysis = &DeckArchetypeAnalysis{
			ColorCounts:       result.Analysis.ColorCounts,
			ColorlessCount:    result.Analysis.ColorlessCount,
			GoldCount:         result.Analysis.GoldCount,
			CreatureCount:     result.Analysis.CreatureCount,
			InstantCount:      result.Analysis.InstantCount,
			SorceryCount:      result.Analysis.SorceryCount,
			ArtifactCount:     result.Analysis.ArtifactCount,
			EnchantmentCount:  result.Analysis.EnchantmentCount,
			PlaneswalkerCount: result.Analysis.PlaneswalkerCount,
			LandCount:         result.Analysis.LandCount,
			ManaCurve:         result.Analysis.ManaCurve,
			AvgCMC:            result.Analysis.AvgCMC,
			RareCounts:        result.Analysis.RareCounts,
		}
	}

	log.Printf("Classified deck %s as %s (%.0f%% confidence)", deckID, result.PrimaryArchetype, result.Confidence*100)
	return response, nil
}

// ClassifyDraftPoolArchetype classifies a draft pool based on picked cards.
func (d *DeckFacade) ClassifyDraftPoolArchetype(ctx context.Context, draftEventID string) (*ArchetypeClassificationResult, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if d.services.CardService == nil {
		return nil, &AppError{Message: "Card service not initialized"}
	}

	// Get draft cards
	var cardIDs []int
	err := storage.RetryOnBusy(func() error {
		var err error
		cardIDs, err = d.services.Storage.DeckRepo().GetDraftCards(ctx, draftEventID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft cards: %v", err)}
	}

	if len(cardIDs) == 0 {
		return nil, &AppError{Message: "No cards in draft pool"}
	}

	// Create classifier
	classifier := archetype.NewClassifier(
		d.services.CardService,
		d.services.Storage.DeckRepo(),
		d.services.Storage.DeckPerformanceRepo(),
	)

	// Classify the draft pool
	result, err := classifier.ClassifyDraftPool(cardIDs)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to classify draft pool: %v", err)}
	}

	// Convert to response format (same as ClassifyDeckArchetype)
	response := &ArchetypeClassificationResult{
		PrimaryArchetype:   result.PrimaryArchetype,
		SecondaryArchetype: result.SecondaryArchetype,
		Confidence:         result.Confidence,
		ConfidencePercent:  int(result.Confidence * 100),
		ColorIdentity:      result.ColorIdentity,
		DominantColors:     result.DominantColors,
		SignatureCards:     result.SignatureCards,
		TotalCards:         result.TotalCards,
	}

	if result.ColorPair != nil {
		response.ColorPair = &ColorPairInfo{
			Colors: result.ColorPair.Colors,
			Name:   result.ColorPair.Name,
		}
	}

	response.Indicators = make([]ArchetypeIndicatorInfo, len(result.ArchetypeIndicators))
	for i, ind := range result.ArchetypeIndicators {
		response.Indicators[i] = ArchetypeIndicatorInfo{
			CardID:   ind.CardID,
			CardName: ind.CardName,
			Weight:   ind.Weight,
			Reason:   ind.Reason,
		}
	}

	if result.Analysis != nil {
		response.Analysis = &DeckArchetypeAnalysis{
			ColorCounts:       result.Analysis.ColorCounts,
			ColorlessCount:    result.Analysis.ColorlessCount,
			GoldCount:         result.Analysis.GoldCount,
			CreatureCount:     result.Analysis.CreatureCount,
			InstantCount:      result.Analysis.InstantCount,
			SorceryCount:      result.Analysis.SorceryCount,
			ArtifactCount:     result.Analysis.ArtifactCount,
			EnchantmentCount:  result.Analysis.EnchantmentCount,
			PlaneswalkerCount: result.Analysis.PlaneswalkerCount,
			LandCount:         result.Analysis.LandCount,
			ManaCurve:         result.Analysis.ManaCurve,
			AvgCMC:            result.Analysis.AvgCMC,
			RareCounts:        result.Analysis.RareCounts,
		}
	}

	log.Printf("Classified draft pool %s as %s (%.0f%% confidence)", draftEventID, result.PrimaryArchetype, result.Confidence*100)
	return response, nil
}

// SuggestDecksResponse wraps the recommendations response for the frontend.
type SuggestDecksResponse struct {
	Suggestions  []*SuggestedDeckResponse  `json:"suggestions"`
	TotalCombos  int                       `json:"totalCombos"`
	ViableCombos int                       `json:"viableCombos"`
	BestCombo    *ColorCombinationResponse `json:"bestCombo,omitempty"`
	Error        string                    `json:"error,omitempty"`
}

// ColorCombinationResponse represents a color combination for the frontend.
type ColorCombinationResponse struct {
	Colors []string `json:"colors"`
	Name   string   `json:"name"`
}

// SuggestedDeckResponse represents a suggested deck for the frontend.
type SuggestedDeckResponse struct {
	ColorCombo ColorCombinationResponse        `json:"colorCombo"`
	Spells     []*SuggestedCardResponse        `json:"spells"`
	Lands      []*SuggestedLandResponse        `json:"lands"`
	TotalCards int                             `json:"totalCards"`
	Score      float64                         `json:"score"`
	Viability  string                          `json:"viability"`
	Analysis   *DeckSuggestionAnalysisResponse `json:"analysis"`
}

// SuggestedCardResponse represents a suggested card for the frontend.
type SuggestedCardResponse struct {
	CardID    int      `json:"cardID"`
	Name      string   `json:"name"`
	TypeLine  string   `json:"typeLine"`
	ManaCost  string   `json:"manaCost,omitempty"`
	ImageURI  string   `json:"imageURI,omitempty"`
	CMC       int      `json:"cmc"`
	Colors    []string `json:"colors"`
	Rarity    string   `json:"rarity,omitempty"`
	Score     float64  `json:"score"`
	Reasoning string   `json:"reasoning"`
}

// SuggestedLandResponse represents a suggested land for the frontend.
type SuggestedLandResponse struct {
	CardID   int    `json:"cardID"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Color    string `json:"color"`
}

// DeckSuggestionAnalysisResponse provides deck composition details for the frontend.
type DeckSuggestionAnalysisResponse struct {
	CreatureCount     int            `json:"creatureCount"`
	SpellCount        int            `json:"spellCount"`
	AverageCMC        float64        `json:"averageCMC"`
	ManaCurve         map[int]int    `json:"manaCurve"`
	ColorDistribution map[string]int `json:"colorDistribution"`
	TopCards          []string       `json:"topCards"`
	Synergies         []string       `json:"synergies"`
	PlayableCount     int            `json:"playableCount"`
}

// SuggestDecks generates all viable deck suggestions for a draft pool.
func (d *DeckFacade) SuggestDecks(ctx context.Context, draftEventID string) (*SuggestDecksResponse, error) {
	log.Printf("[SuggestDecks] Called with draftEventID=%s", draftEventID)

	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get the draft session for set code and format
	session, err := d.services.Storage.DraftRepo().GetSession(ctx, draftEventID)
	if err != nil {
		return &SuggestDecksResponse{
			Error: fmt.Sprintf("Failed to get draft session: %v", err),
		}, nil
	}
	if session == nil {
		return &SuggestDecksResponse{
			Error: "Draft session not found",
		}, nil
	}

	// Extract set code from EventName (e.g., "QuickDraft_BLB_20250101" -> "BLB")
	var setCode, draftFormat string
	eventParts := strings.Split(session.EventName, "_")
	if len(eventParts) >= 2 {
		setCode = eventParts[1]
		if strings.HasPrefix(session.EventName, "PremierDraft") {
			draftFormat = "PremierDraft"
		} else if strings.HasPrefix(session.EventName, "QuickDraft") {
			draftFormat = "QuickDraft"
		} else {
			draftFormat = "PremierDraft"
		}
	}

	// Get all cards from the draft session
	draftPool, err := d.services.Storage.DeckRepo().GetDraftCards(ctx, draftEventID)
	if err != nil {
		return &SuggestDecksResponse{
			Error: fmt.Sprintf("Failed to get draft pool: %v", err),
		}, nil
	}

	if len(draftPool) == 0 {
		return &SuggestDecksResponse{
			Error: "No cards in draft pool",
		}, nil
	}

	log.Printf("SuggestDecks: Draft pool has %d cards, set=%s, format=%s", len(draftPool), setCode, draftFormat)

	// Create deck suggester
	suggester := recommendations.NewDeckSuggester(
		d.services.RecommendationEngine.(*recommendations.RuleBasedEngine),
		d.services.CardService,
		d.services.Storage.SetCardRepo(),
		d.services.Storage.DraftRatingsRepo(),
	)

	// Get suggestions
	result, err := suggester.SuggestDecks(ctx, draftPool, setCode, draftFormat)
	if err != nil {
		return &SuggestDecksResponse{
			Error: fmt.Sprintf("Failed to generate suggestions: %v", err),
		}, nil
	}

	// Check if the suggester returned an error
	if result.Error != "" {
		return &SuggestDecksResponse{
			Error: result.Error,
		}, nil
	}

	// Convert to response format
	response := &SuggestDecksResponse{
		Suggestions:  make([]*SuggestedDeckResponse, len(result.Suggestions)),
		TotalCombos:  result.TotalCombos,
		ViableCombos: result.ViableCombos,
	}

	if result.BestCombo != nil {
		response.BestCombo = &ColorCombinationResponse{
			Colors: result.BestCombo.Colors,
			Name:   result.BestCombo.Name,
		}
	}

	for i, suggestion := range result.Suggestions {
		response.Suggestions[i] = convertSuggestedDeck(suggestion)
	}

	log.Printf("SuggestDecks: Found %d viable color combinations", len(response.Suggestions))
	return response, nil
}

// convertSuggestedDeck converts internal suggestion to response format.
func convertSuggestedDeck(s *recommendations.SuggestedDeck) *SuggestedDeckResponse {
	spells := make([]*SuggestedCardResponse, len(s.Spells))
	for i, card := range s.Spells {
		spells[i] = &SuggestedCardResponse{
			CardID:    card.CardID,
			Name:      card.Name,
			TypeLine:  card.TypeLine,
			ManaCost:  card.ManaCost,
			ImageURI:  card.ImageURI,
			CMC:       card.CMC,
			Colors:    card.Colors,
			Rarity:    card.Rarity,
			Score:     card.Score,
			Reasoning: card.Reasoning,
		}
	}

	lands := make([]*SuggestedLandResponse, len(s.Lands))
	for i, land := range s.Lands {
		lands[i] = &SuggestedLandResponse{
			CardID:   land.CardID,
			Name:     land.Name,
			Quantity: land.Quantity,
			Color:    land.Color,
		}
	}

	var analysis *DeckSuggestionAnalysisResponse
	if s.Analysis != nil {
		analysis = &DeckSuggestionAnalysisResponse{
			CreatureCount:     s.Analysis.CreatureCount,
			SpellCount:        s.Analysis.SpellCount,
			AverageCMC:        s.Analysis.AverageCMC,
			ManaCurve:         s.Analysis.ManaCurve,
			ColorDistribution: s.Analysis.ColorDistribution,
			TopCards:          s.Analysis.TopCards,
			Synergies:         s.Analysis.Synergies,
			PlayableCount:     s.Analysis.PlayableCount,
		}
	}

	return &SuggestedDeckResponse{
		ColorCombo: ColorCombinationResponse{
			Colors: s.ColorCombo.Colors,
			Name:   s.ColorCombo.Name,
		},
		Spells:     spells,
		Lands:      lands,
		TotalCards: s.TotalCards,
		Score:      s.Score,
		Viability:  s.Viability,
		Analysis:   analysis,
	}
}

// ApplySuggestedDeck replaces the current deck with a suggested deck.
func (d *DeckFacade) ApplySuggestedDeck(ctx context.Context, deckID string, suggestion *SuggestedDeckResponse) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Get the deck
	var deck *models.Deck
	err := storage.RetryOnBusy(func() error {
		var err error
		deck, err = d.services.Storage.DeckRepo().GetByID(ctx, deckID)
		return err
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get deck: %v", err)}
	}
	if deck == nil {
		return &AppError{Message: "Deck not found"}
	}

	// Get existing cards to preserve sideboard
	var existingCards []*models.DeckCard
	err = storage.RetryOnBusy(func() error {
		var err error
		existingCards, err = d.services.Storage.DeckRepo().GetCards(ctx, deckID)
		return err
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get existing cards: %v", err)}
	}

	// Separate sideboard cards to preserve them
	sideboardCards := make([]*models.DeckCard, 0)
	for _, card := range existingCards {
		if card.Board == "sideboard" {
			sideboardCards = append(sideboardCards, card)
		}
	}

	// Clear all cards
	err = storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().ClearCards(ctx, deckID)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to clear deck: %v", err)}
	}

	// Add suggested spells
	for _, card := range suggestion.Spells {
		deckCard := &models.DeckCard{
			DeckID:        deckID,
			CardID:        card.CardID,
			Quantity:      1,
			Board:         "main",
			FromDraftPick: true,
		}
		err = storage.RetryOnBusy(func() error {
			return d.services.Storage.DeckRepo().AddCard(ctx, deckCard)
		})
		if err != nil {
			log.Printf("Warning: Failed to add card %d to deck: %v", card.CardID, err)
		}
	}

	// Add suggested lands
	for _, land := range suggestion.Lands {
		deckCard := &models.DeckCard{
			DeckID:        deckID,
			CardID:        land.CardID,
			Quantity:      land.Quantity,
			Board:         "main",
			FromDraftPick: false,
		}
		err = storage.RetryOnBusy(func() error {
			return d.services.Storage.DeckRepo().AddCard(ctx, deckCard)
		})
		if err != nil {
			log.Printf("Warning: Failed to add land %s to deck: %v", land.Name, err)
		}
	}

	// Restore sideboard cards
	for _, card := range sideboardCards {
		err = storage.RetryOnBusy(func() error {
			return d.services.Storage.DeckRepo().AddCard(ctx, card)
		})
		if err != nil {
			log.Printf("Warning: Failed to restore sideboard card %d: %v", card.CardID, err)
		}
	}

	// Update deck modified timestamp
	deck.ModifiedAt = time.Now()
	err = storage.RetryOnBusy(func() error {
		return d.services.Storage.DeckRepo().Update(ctx, deck)
	})
	if err != nil {
		log.Printf("Warning: Failed to update deck modified timestamp: %v", err)
	}

	log.Printf("Applied suggested %s deck to %s", suggestion.ColorCombo.Name, deckID)
	return nil
}

// GetSuggestedDeckExportContent returns the export content for a suggested deck.
// This allows the frontend to handle file saving via browser download.
func (d *DeckFacade) GetSuggestedDeckExportContent(_ context.Context, suggestion *SuggestedDeckResponse, deckName string) (string, error) {
	// Build export content in Arena format
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Deck: %s (%s)\n\n", deckName, suggestion.ColorCombo.Name))

	// Add spells grouped by type
	creatures := make([]*SuggestedCardResponse, 0)
	nonCreatures := make([]*SuggestedCardResponse, 0)

	for _, card := range suggestion.Spells {
		if strings.Contains(strings.ToLower(card.TypeLine), "creature") {
			creatures = append(creatures, card)
		} else {
			nonCreatures = append(nonCreatures, card)
		}
	}

	// Write creatures
	if len(creatures) > 0 {
		content.WriteString("// Creatures\n")
		for _, card := range creatures {
			content.WriteString(fmt.Sprintf("1 %s\n", card.Name))
		}
		content.WriteString("\n")
	}

	// Write non-creatures
	if len(nonCreatures) > 0 {
		content.WriteString("// Spells\n")
		for _, card := range nonCreatures {
			content.WriteString(fmt.Sprintf("1 %s\n", card.Name))
		}
		content.WriteString("\n")
	}

	// Write lands
	content.WriteString("// Lands\n")
	for _, land := range suggestion.Lands {
		content.WriteString(fmt.Sprintf("%d %s\n", land.Quantity, land.Name))
	}

	return content.String(), nil
}

// ExportSuggestedDeckToFile exports a suggested deck to the specified file path.
func (d *DeckFacade) ExportSuggestedDeckToFile(ctx context.Context, suggestion *SuggestedDeckResponse, deckName string, filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is required")
	}

	content, err := d.GetSuggestedDeckExportContent(ctx, suggestion, deckName)
	if err != nil {
		return err
	}

	// Save the file
	err = os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	log.Printf("Exported suggested deck '%s' to %s", deckName, filepath.Base(filePath))
	return nil
}

// BuildAroundSeedRequest represents a request to build a deck around a seed card.
type BuildAroundSeedRequest struct {
	SeedCardID     int      `json:"seedCardID"`
	MaxResults     int      `json:"maxResults,omitempty"`
	BudgetMode     bool     `json:"budgetMode,omitempty"`
	SetRestriction string   `json:"setRestriction,omitempty"`
	AllowedSets    []string `json:"allowedSets,omitempty"`
}

// BuildAroundSeedResponse contains the deck building suggestions.
type BuildAroundSeedResponse struct {
	SeedCard        *CardWithOwnershipResponse   `json:"seedCard"`
	Suggestions     []*CardWithOwnershipResponse `json:"suggestions"`
	LandSuggestions []*SuggestedLandResponse     `json:"lands"`
	Analysis        *SeedDeckAnalysisResponse    `json:"analysis"`
}

// CardWithOwnershipResponse represents a card with ownership info for API response.
type CardWithOwnershipResponse struct {
	CardID       int      `json:"cardID"`
	Name         string   `json:"name"`
	ManaCost     string   `json:"manaCost,omitempty"`
	CMC          int      `json:"cmc"`
	Colors       []string `json:"colors"`
	TypeLine     string   `json:"typeLine"`
	Rarity       string   `json:"rarity,omitempty"`
	ImageURI     string   `json:"imageURI,omitempty"`
	Score        float64  `json:"score"`
	Reasoning    string   `json:"reasoning"`
	InCollection bool     `json:"inCollection"`
	OwnedCount   int      `json:"ownedCount"`
	NeededCount  int      `json:"neededCount"`
}

// SeedDeckAnalysisResponse provides analysis of the seed card and suggestions.
type SeedDeckAnalysisResponse struct {
	ColorIdentity       []string       `json:"colorIdentity"`
	Keywords            []string       `json:"keywords"`
	Themes              []string       `json:"themes"`
	IdealCurve          map[int]int    `json:"idealCurve"`
	SuggestedLandCount  int            `json:"suggestedLandCount"`
	TotalCards          int            `json:"totalCards"`
	InCollectionCount   int            `json:"inCollectionCount"`
	MissingCount        int            `json:"missingCount"`
	MissingWildcardCost map[string]int `json:"missingWildcardCost"`
}

// BuildAroundSeed generates deck suggestions based on a seed card.
func (d *DeckFacade) BuildAroundSeed(ctx context.Context, req *BuildAroundSeedRequest) (*BuildAroundSeedResponse, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}
	if d.services.CardService == nil {
		return nil, &AppError{Message: "Card service not initialized"}
	}

	// Create seed deck builder inline (like SuggestDecks does with DeckSuggester)
	builder := recommendations.NewSeedDeckBuilder(
		d.services.Storage.SetCardRepo(),
		d.services.Storage.CollectionRepo(),
		d.services.Storage.StandardRepo(),
		d.services.CardService,
	)

	// Build the request
	builderReq := &recommendations.SeedDeckBuilderRequest{
		SeedCardID:     req.SeedCardID,
		MaxResults:     req.MaxResults,
		BudgetMode:     req.BudgetMode,
		SetRestriction: req.SetRestriction,
		AllowedSets:    req.AllowedSets,
	}

	// Get suggestions
	result, err := builder.BuildAroundSeed(ctx, builderReq)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to build deck: %v", err)}
	}

	// Convert to response type
	return convertBuildAroundResponse(result), nil
}

// convertBuildAroundResponse converts the internal response to the API response type.
func convertBuildAroundResponse(r *recommendations.SeedDeckBuilderResponse) *BuildAroundSeedResponse {
	if r == nil {
		return nil
	}

	response := &BuildAroundSeedResponse{}

	// Convert seed card
	if r.SeedCard != nil {
		response.SeedCard = convertCardWithOwnership(r.SeedCard)
	}

	// Convert suggestions
	response.Suggestions = make([]*CardWithOwnershipResponse, 0, len(r.Suggestions))
	for _, card := range r.Suggestions {
		response.Suggestions = append(response.Suggestions, convertCardWithOwnership(card))
	}

	// Convert land suggestions
	response.LandSuggestions = make([]*SuggestedLandResponse, 0, len(r.LandSuggestions))
	for _, land := range r.LandSuggestions {
		response.LandSuggestions = append(response.LandSuggestions, &SuggestedLandResponse{
			CardID:   land.CardID,
			Name:     land.Name,
			Quantity: land.Quantity,
			Color:    land.Color,
		})
	}

	// Convert analysis
	if r.Analysis != nil {
		response.Analysis = &SeedDeckAnalysisResponse{
			ColorIdentity:       r.Analysis.ColorIdentity,
			Keywords:            r.Analysis.Keywords,
			Themes:              r.Analysis.Themes,
			IdealCurve:          r.Analysis.IdealCurve,
			SuggestedLandCount:  r.Analysis.SuggestedLandCount,
			TotalCards:          r.Analysis.TotalCards,
			InCollectionCount:   r.Analysis.InCollectionCount,
			MissingCount:        r.Analysis.MissingCount,
			MissingWildcardCost: r.Analysis.MissingWildcardCost,
		}
	}

	return response
}

// convertCardWithOwnership converts a card with ownership to the response type.
func convertCardWithOwnership(c *recommendations.CardWithOwnership) *CardWithOwnershipResponse {
	if c == nil {
		return nil
	}
	return &CardWithOwnershipResponse{
		CardID:       c.CardID,
		Name:         c.Name,
		ManaCost:     c.ManaCost,
		CMC:          c.CMC,
		Colors:       c.Colors,
		TypeLine:     c.TypeLine,
		Rarity:       c.Rarity,
		ImageURI:     c.ImageURI,
		Score:        c.Score,
		Reasoning:    c.Reasoning,
		InCollection: c.InCollection,
		OwnedCount:   c.OwnedCount,
		NeededCount:  c.NeededCount,
	}
}

// IterativeBuildAroundRequest represents a request for iterative deck building suggestions.
type IterativeBuildAroundRequest struct {
	SeedCardID     int      `json:"seedCardID"`
	DeckCardIDs    []int    `json:"deckCardIDs"`
	MaxResults     int      `json:"maxResults,omitempty"`
	BudgetMode     bool     `json:"budgetMode,omitempty"`
	SetRestriction string   `json:"setRestriction,omitempty"`
	AllowedSets    []string `json:"allowedSets,omitempty"`
}

// IterativeBuildAroundResponse contains suggestions for iterative deck building.
type IterativeBuildAroundResponse struct {
	Suggestions     []*CardWithOwnershipResponse `json:"suggestions"`
	DeckAnalysis    *LiveDeckAnalysisResponse    `json:"deckAnalysis"`
	SlotsRemaining  int                          `json:"slotsRemaining"`
	LandSuggestions []*SuggestedLandResponse     `json:"landSuggestions"`
}

// LiveDeckAnalysisResponse provides real-time analysis of the deck being built.
type LiveDeckAnalysisResponse struct {
	ColorIdentity        []string    `json:"colorIdentity"`
	Keywords             []string    `json:"keywords"`
	Themes               []string    `json:"themes"`
	CurrentCurve         map[int]int `json:"currentCurve"`
	RecommendedLandCount int         `json:"recommendedLandCount"`
	TotalCards           int         `json:"totalCards"`
	InCollectionCount    int         `json:"inCollectionCount"`
}

// SuggestNextCards generates suggestions based on the current deck composition.
// This is used for iterative deck building where users pick cards one-by-one.
func (d *DeckFacade) SuggestNextCards(ctx context.Context, req *IterativeBuildAroundRequest) (*IterativeBuildAroundResponse, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}
	if d.services.CardService == nil {
		return nil, &AppError{Message: "Card service not initialized"}
	}

	// Create seed deck builder
	builder := recommendations.NewSeedDeckBuilder(
		d.services.Storage.SetCardRepo(),
		d.services.Storage.CollectionRepo(),
		d.services.Storage.StandardRepo(),
		d.services.CardService,
	)

	// Convert request
	builderReq := &recommendations.IterativeBuildAroundRequest{
		SeedCardID:     req.SeedCardID,
		DeckCardIDs:    req.DeckCardIDs,
		MaxResults:     req.MaxResults,
		BudgetMode:     req.BudgetMode,
		SetRestriction: req.SetRestriction,
		AllowedSets:    req.AllowedSets,
	}

	result, err := builder.SuggestNextCards(ctx, builderReq)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to suggest cards: %v", err)}
	}

	// Convert to response type
	return convertIterativeResponse(result), nil
}

// convertIterativeResponse converts the internal response to the API response type.
func convertIterativeResponse(r *recommendations.IterativeBuildAroundResponse) *IterativeBuildAroundResponse {
	if r == nil {
		return nil
	}

	response := &IterativeBuildAroundResponse{
		SlotsRemaining: r.SlotsRemaining,
	}

	// Convert suggestions
	response.Suggestions = make([]*CardWithOwnershipResponse, 0, len(r.Suggestions))
	for _, card := range r.Suggestions {
		response.Suggestions = append(response.Suggestions, convertCardWithOwnership(card))
	}

	// Convert land suggestions
	response.LandSuggestions = make([]*SuggestedLandResponse, 0, len(r.LandSuggestions))
	for _, land := range r.LandSuggestions {
		response.LandSuggestions = append(response.LandSuggestions, &SuggestedLandResponse{
			CardID:   land.CardID,
			Name:     land.Name,
			Quantity: land.Quantity,
			Color:    land.Color,
		})
	}

	// Convert deck analysis
	if r.DeckAnalysis != nil {
		response.DeckAnalysis = &LiveDeckAnalysisResponse{
			ColorIdentity:        r.DeckAnalysis.ColorIdentity,
			Keywords:             r.DeckAnalysis.Keywords,
			Themes:               r.DeckAnalysis.Themes,
			CurrentCurve:         r.DeckAnalysis.CurrentCurve,
			RecommendedLandCount: r.DeckAnalysis.RecommendedLandCount,
			TotalCards:           r.DeckAnalysis.TotalCards,
			InCollectionCount:    r.DeckAnalysis.InCollectionCount,
		}
	}

	return response
}

// GenerateCompleteDeckRequest represents a request to generate a complete 60-card deck.
type GenerateCompleteDeckRequest struct {
	SeedCardID     int      `json:"seedCardID"`
	Archetype      string   `json:"archetype"`                // "aggro", "midrange", "control"
	BudgetMode     bool     `json:"budgetMode,omitempty"`     // Only collection cards
	SetRestriction string   `json:"setRestriction,omitempty"` // "single", "multiple", "all"
	AllowedSets    []string `json:"allowedSets,omitempty"`    // Specific set codes if "multiple"
}

// GenerateCompleteDeckResponse contains a complete 60-card deck with strategy.
type GenerateCompleteDeckResponse struct {
	SeedCard *CardWithOwnershipResponse     `json:"seedCard"`
	Spells   []*CardWithQuantityResponse    `json:"spells"` // Non-land cards with quantities
	Lands    []*LandWithQuantityResponse    `json:"lands"`  // Lands with quantities
	Strategy *DeckStrategyResponse          `json:"strategy"`
	Analysis *GeneratedDeckAnalysisResponse `json:"analysis"`
}

// CardWithQuantityResponse represents a card with how many copies to include.
type CardWithQuantityResponse struct {
	CardID         int                     `json:"cardID"`
	Name           string                  `json:"name"`
	ManaCost       string                  `json:"manaCost,omitempty"`
	CMC            int                     `json:"cmc"`
	Colors         []string                `json:"colors"`
	TypeLine       string                  `json:"typeLine"`
	Rarity         string                  `json:"rarity,omitempty"`
	ImageURI       string                  `json:"imageURI,omitempty"`
	Score          float64                 `json:"score"`
	Reasoning      string                  `json:"reasoning"`
	Quantity       int                     `json:"quantity"`
	InCollection   bool                    `json:"inCollection"`
	OwnedCount     int                     `json:"ownedCount"`
	NeededCount    int                     `json:"neededCount"`
	ScoreBreakdown *ScoreBreakdownResponse `json:"scoreBreakdown,omitempty"`
	SynergyDetails []SynergyDetailResponse `json:"synergyDetails,omitempty"`
}

// ScoreBreakdownResponse provides detailed scoring factors.
type ScoreBreakdownResponse struct {
	ColorFit float64 `json:"colorFit"`
	CurveFit float64 `json:"curveFit"`
	Synergy  float64 `json:"synergy"`
	Quality  float64 `json:"quality"`
	Overall  float64 `json:"overall"`
}

// SynergyDetailResponse describes a specific synergy.
type SynergyDetailResponse struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LandWithQuantityResponse represents a land with quantity and type information.
type LandWithQuantityResponse struct {
	CardID       int      `json:"cardID"`
	Name         string   `json:"name"`
	Quantity     int      `json:"quantity"`
	Colors       []string `json:"colors"`
	IsBasic      bool     `json:"isBasic"`
	EntersTapped bool     `json:"entersTapped"`
}

// DeckStrategyResponse provides human-readable deck strategy information.
type DeckStrategyResponse struct {
	Summary    string   `json:"summary"`
	GamePlan   string   `json:"gamePlan"`
	KeyCards   []string `json:"keyCards"`
	Mulligan   string   `json:"mulligan"`
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
}

// GeneratedDeckAnalysisResponse provides detailed analysis of the generated deck.
type GeneratedDeckAnalysisResponse struct {
	TotalCards          int            `json:"totalCards"`
	SpellCount          int            `json:"spellCount"`
	LandCount           int            `json:"landCount"`
	CreatureCount       int            `json:"creatureCount"`
	NonCreatureCount    int            `json:"nonCreatureCount"`
	AverageCMC          float64        `json:"averageCMC"`
	ManaCurve           map[int]int    `json:"manaCurve"`
	ColorDistribution   map[string]int `json:"colorDistribution"`
	InCollectionCount   int            `json:"inCollectionCount"`
	MissingCount        int            `json:"missingCount"`
	MissingWildcardCost map[string]int `json:"missingWildcardCost"`
	ArchetypeMatch      float64        `json:"archetypeMatch"`
}

// ArchetypeProfileResponse represents an archetype profile for the frontend.
type ArchetypeProfileResponse struct {
	Name          string      `json:"name"`
	LandCount     int         `json:"landCount"`
	CurveTargets  map[int]int `json:"curveTargets"`
	CreatureRatio float64     `json:"creatureRatio"`
	RemovalCount  int         `json:"removalCount"`
	CardAdvantage int         `json:"cardAdvantage"`
	Description   string      `json:"description"`
}

// GenerateCompleteDeck generates a complete 60-card deck from a seed card and archetype.
func (d *DeckFacade) GenerateCompleteDeck(ctx context.Context, req *GenerateCompleteDeckRequest) (*GenerateCompleteDeckResponse, error) {
	if req == nil {
		return nil, &AppError{Message: "Request is required"}
	}
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}
	if d.services.CardService == nil {
		return nil, &AppError{Message: "Card service not initialized"}
	}

	// Create seed deck builder
	builder := recommendations.NewSeedDeckBuilder(
		d.services.Storage.SetCardRepo(),
		d.services.Storage.CollectionRepo(),
		d.services.Storage.StandardRepo(),
		d.services.CardService,
	)

	// Convert request
	builderReq := &recommendations.GenerateCompleteDeckRequest{
		SeedCardID:     req.SeedCardID,
		Archetype:      req.Archetype,
		BudgetMode:     req.BudgetMode,
		SetRestriction: req.SetRestriction,
		AllowedSets:    req.AllowedSets,
	}

	result, err := builder.GenerateCompleteDeck(ctx, builderReq)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to generate deck: %v", err)}
	}

	// Convert to response type
	return convertGenerateCompleteDeckResponse(result), nil
}

// GetArchetypeProfiles returns all available archetype profiles.
func (d *DeckFacade) GetArchetypeProfiles() []*ArchetypeProfileResponse {
	profiles := recommendations.GetAllArchetypeProfiles()
	result := make([]*ArchetypeProfileResponse, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, &ArchetypeProfileResponse{
			Name:          p.Name,
			LandCount:     p.LandCount,
			CurveTargets:  p.CurveTargets,
			CreatureRatio: p.CreatureRatio,
			RemovalCount:  p.RemovalCount,
			CardAdvantage: p.CardAdvantage,
			Description:   p.Description,
		})
	}
	return result
}

// convertGenerateCompleteDeckResponse converts the internal response to the API response type.
func convertGenerateCompleteDeckResponse(r *recommendations.GenerateCompleteDeckResponse) *GenerateCompleteDeckResponse {
	if r == nil {
		return nil
	}

	response := &GenerateCompleteDeckResponse{}

	// Convert seed card
	if r.SeedCard != nil {
		response.SeedCard = convertCardWithOwnership(r.SeedCard)
	}

	// Convert spells with quantity
	response.Spells = make([]*CardWithQuantityResponse, 0, len(r.Spells))
	for _, spell := range r.Spells {
		response.Spells = append(response.Spells, convertCardWithQuantity(spell))
	}

	// Convert lands
	response.Lands = make([]*LandWithQuantityResponse, 0, len(r.Lands))
	for _, land := range r.Lands {
		response.Lands = append(response.Lands, &LandWithQuantityResponse{
			CardID:       land.CardID,
			Name:         land.Name,
			Quantity:     land.Quantity,
			Colors:       land.Colors,
			IsBasic:      land.IsBasic,
			EntersTapped: land.EntersTapped,
		})
	}

	// Convert strategy
	if r.Strategy != nil {
		response.Strategy = &DeckStrategyResponse{
			Summary:    r.Strategy.Summary,
			GamePlan:   r.Strategy.GamePlan,
			KeyCards:   r.Strategy.KeyCards,
			Mulligan:   r.Strategy.Mulligan,
			Strengths:  r.Strategy.Strengths,
			Weaknesses: r.Strategy.Weaknesses,
		}
	}

	// Convert analysis
	if r.Analysis != nil {
		response.Analysis = &GeneratedDeckAnalysisResponse{
			TotalCards:          r.Analysis.TotalCards,
			SpellCount:          r.Analysis.SpellCount,
			LandCount:           r.Analysis.LandCount,
			CreatureCount:       r.Analysis.CreatureCount,
			NonCreatureCount:    r.Analysis.NonCreatureCount,
			AverageCMC:          r.Analysis.AverageCMC,
			ManaCurve:           r.Analysis.ManaCurve,
			ColorDistribution:   r.Analysis.ColorDistribution,
			InCollectionCount:   r.Analysis.InCollectionCount,
			MissingCount:        r.Analysis.MissingCount,
			MissingWildcardCost: r.Analysis.MissingWildcardCost,
			ArchetypeMatch:      r.Analysis.ArchetypeMatch,
		}
	}

	return response
}

// convertCardWithQuantity converts a card with quantity to the response type.
func convertCardWithQuantity(c *recommendations.CardWithQuantity) *CardWithQuantityResponse {
	if c == nil {
		return nil
	}

	response := &CardWithQuantityResponse{
		CardID:       c.CardID,
		Name:         c.Name,
		ManaCost:     c.ManaCost,
		CMC:          c.CMC,
		Colors:       c.Colors,
		TypeLine:     c.TypeLine,
		Rarity:       c.Rarity,
		ImageURI:     c.ImageURI,
		Score:        c.Score,
		Reasoning:    c.Reasoning,
		Quantity:     c.Quantity,
		InCollection: c.InCollection,
		OwnedCount:   c.OwnedCount,
		NeededCount:  c.NeededCount,
	}

	// Convert score breakdown
	if c.ScoreBreakdown != nil {
		response.ScoreBreakdown = &ScoreBreakdownResponse{
			ColorFit: c.ScoreBreakdown.ColorFit,
			CurveFit: c.ScoreBreakdown.CurveFit,
			Synergy:  c.ScoreBreakdown.Synergy,
			Quality:  c.ScoreBreakdown.Quality,
			Overall:  c.ScoreBreakdown.Overall,
		}
	}

	// Convert synergy details
	if len(c.SynergyDetails) > 0 {
		response.SynergyDetails = make([]SynergyDetailResponse, 0, len(c.SynergyDetails))
		for _, sd := range c.SynergyDetails {
			response.SynergyDetails = append(response.SynergyDetails, SynergyDetailResponse{
				Type:        sd.Type,
				Name:        sd.Name,
				Description: sd.Description,
			})
		}
	}

	return response
}

// ============================================================================
// Card Performance Analysis (Issue #771)
// ============================================================================

// CardPerformanceResponse represents a single card's performance metrics for the frontend.
type CardPerformanceResponse struct {
	CardID            int         `json:"cardId"`
	CardName          string      `json:"cardName"`
	Quantity          int         `json:"quantity"`
	GamesWithCard     int         `json:"gamesWithCard"`
	GamesDrawn        int         `json:"gamesDrawn"`
	GamesPlayed       int         `json:"gamesPlayed"`
	WinRateWhenDrawn  float64     `json:"winRateWhenDrawn"`
	WinRateWhenPlayed float64     `json:"winRateWhenPlayed"`
	DeckWinRate       float64     `json:"deckWinRate"`
	PlayRate          float64     `json:"playRate"`
	WinContribution   float64     `json:"winContribution"`
	ImpactScore       float64     `json:"impactScore"`
	ConfidenceLevel   string      `json:"confidenceLevel"`
	SampleSize        int         `json:"sampleSize"`
	PerformanceGrade  string      `json:"performanceGrade"`
	AvgTurnPlayed     float64     `json:"avgTurnPlayed"`
	TurnPlayedDist    map[int]int `json:"turnPlayedDist,omitempty"`
	MulliganedAway    int         `json:"mulliganedAway"`
	MulliganRate      float64     `json:"mulliganRate"`
}

// DeckPerformanceAnalysisResponse represents the full deck performance analysis.
type DeckPerformanceAnalysisResponse struct {
	DeckID          string                     `json:"deckId"`
	DeckName        string                     `json:"deckName"`
	TotalMatches    int                        `json:"totalMatches"`
	TotalGames      int                        `json:"totalGames"`
	OverallWinRate  float64                    `json:"overallWinRate"`
	CardPerformance []*CardPerformanceResponse `json:"cardPerformance"`
	BestPerformers  []string                   `json:"bestPerformers"`
	WorstPerformers []string                   `json:"worstPerformers"`
	AnalysisDate    string                     `json:"analysisDate"`
}

// CardRecommendationResponse represents a suggestion to add, remove, or swap a card.
type CardRecommendationResponse struct {
	Type            string  `json:"type"`
	CardID          int     `json:"cardId"`
	CardName        string  `json:"cardName"`
	Reason          string  `json:"reason"`
	ImpactEstimate  float64 `json:"impactEstimate"`
	Confidence      string  `json:"confidence"`
	Priority        int     `json:"priority"`
	SwapForCardID   *int    `json:"swapForCardId,omitempty"`
	SwapForCardName *string `json:"swapForCardName,omitempty"`
	BasedOnGames    int     `json:"basedOnGames"`
}

// DeckRecommendationsResponse contains add/remove/swap recommendations for a deck.
type DeckRecommendationsResponse struct {
	DeckID                string                        `json:"deckId"`
	DeckName              string                        `json:"deckName"`
	CurrentWinRate        float64                       `json:"currentWinRate"`
	AddRecommendations    []*CardRecommendationResponse `json:"addRecommendations"`
	RemoveRecommendations []*CardRecommendationResponse `json:"removeRecommendations"`
	SwapRecommendations   []*CardRecommendationResponse `json:"swapRecommendations"`
	ProjectedWinRate      float64                       `json:"projectedWinRate"`
}

// GetCardPerformanceRequest represents a request to get card performance metrics.
type GetCardPerformanceRequest struct {
	DeckID       string `json:"deckId"`
	MinGames     int    `json:"minGames,omitempty"`
	IncludeLands bool   `json:"includeLands,omitempty"`
}

// GetCardPerformance returns performance metrics for all cards in a deck.
func (d *DeckFacade) GetCardPerformance(ctx context.Context, req *GetCardPerformanceRequest) (*DeckPerformanceAnalysisResponse, error) {
	if req.DeckID == "" {
		return nil, &AppError{Message: "deck_id is required"}
	}

	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	repo := d.services.Storage.CardPerformanceAnalysisRepo()
	if repo == nil {
		return nil, &AppError{Message: "Card performance repository not available"}
	}

	filter := models.CardPerformanceFilter{
		DeckID:       req.DeckID,
		MinGames:     req.MinGames,
		IncludeLands: req.IncludeLands,
	}
	analysis, err := repo.GetDeckPerformanceAnalysis(ctx, filter)
	if err != nil {
		return nil, &AppError{
			Message: fmt.Sprintf("Failed to get card performance: %v", err),
			Err:     err,
		}
	}

	if analysis == nil {
		return nil, &AppError{
			Message: "Deck not found or not enough data",
			Err:     repository.ErrNotEnoughData,
		}
	}

	// Convert to response format
	response := &DeckPerformanceAnalysisResponse{
		DeckID:          analysis.DeckID,
		DeckName:        analysis.DeckName,
		TotalMatches:    analysis.TotalMatches,
		TotalGames:      analysis.TotalGames,
		OverallWinRate:  analysis.OverallWinRate,
		BestPerformers:  analysis.BestPerformers,
		WorstPerformers: analysis.WorstPerformers,
		AnalysisDate:    time.Now().UTC().Format(time.RFC3339),
	}

	// Convert card performance
	response.CardPerformance = make([]*CardPerformanceResponse, 0, len(analysis.CardPerformance))
	for _, perf := range analysis.CardPerformance {
		cardPerf := &CardPerformanceResponse{
			CardID:            perf.CardID,
			CardName:          perf.CardName,
			Quantity:          perf.Quantity,
			GamesWithCard:     perf.GamesWithCard,
			GamesDrawn:        perf.GamesDrawn,
			GamesPlayed:       perf.GamesPlayed,
			WinRateWhenDrawn:  perf.WinRateWhenDrawn,
			WinRateWhenPlayed: perf.WinRateWhenPlayed,
			DeckWinRate:       perf.DeckWinRate,
			PlayRate:          perf.PlayRate,
			WinContribution:   perf.WinContribution,
			ImpactScore:       perf.ImpactScore,
			ConfidenceLevel:   perf.ConfidenceLevel,
			SampleSize:        perf.SampleSize,
			PerformanceGrade:  perf.PerformanceGrade,
			AvgTurnPlayed:     perf.AvgTurnPlayed,
			TurnPlayedDist:    perf.TurnPlayedDist,
			MulliganedAway:    perf.MulliganedAway,
			MulliganRate:      perf.MulliganRate,
		}
		response.CardPerformance = append(response.CardPerformance, cardPerf)
	}

	return response, nil
}

// GetPerformanceRecommendationsRequest represents a request for performance-based recommendations.
type GetPerformanceRecommendationsRequest struct {
	DeckID       string `json:"deckId"`
	MaxResults   int    `json:"maxResults,omitempty"`
	IncludeSwaps bool   `json:"includeSwaps,omitempty"`
	Format       string `json:"format,omitempty"`
}

// GetPerformanceRecommendations returns add/remove/swap recommendations based on card performance.
func (d *DeckFacade) GetPerformanceRecommendations(ctx context.Context, req *GetPerformanceRecommendationsRequest) (*DeckRecommendationsResponse, error) {
	if req.DeckID == "" {
		return nil, &AppError{Message: "deck_id is required"}
	}

	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	repo := d.services.Storage.CardPerformanceAnalysisRepo()
	if repo == nil {
		return nil, &AppError{Message: "Card performance repository not available"}
	}

	// Get the repo's recommendations method via type assertion
	repoWithRecs, ok := repo.(interface {
		GetCardRecommendations(ctx context.Context, req models.RecommendationsRequest) (*models.RecommendationsResponse, error)
	})
	if !ok {
		return nil, &AppError{Message: "Card recommendations not supported"}
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}

	recsReq := models.RecommendationsRequest{
		DeckID:       req.DeckID,
		Format:       req.Format,
		MaxResults:   maxResults,
		IncludeSwaps: req.IncludeSwaps,
	}

	recs, err := repoWithRecs.GetCardRecommendations(ctx, recsReq)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get recommendations: %v", err)}
	}

	if recs == nil {
		return nil, &AppError{Message: "Could not generate recommendations - not enough data"}
	}

	// Convert to response format
	response := &DeckRecommendationsResponse{
		DeckID:           recs.DeckID,
		DeckName:         recs.DeckName,
		CurrentWinRate:   recs.CurrentWinRate,
		ProjectedWinRate: recs.ProjectedWinRate,
	}

	// Convert add recommendations
	response.AddRecommendations = make([]*CardRecommendationResponse, 0, len(recs.AddRecommendations))
	for _, rec := range recs.AddRecommendations {
		response.AddRecommendations = append(response.AddRecommendations, convertRecommendation(rec))
	}

	// Convert remove recommendations
	response.RemoveRecommendations = make([]*CardRecommendationResponse, 0, len(recs.RemoveRecommendations))
	for _, rec := range recs.RemoveRecommendations {
		response.RemoveRecommendations = append(response.RemoveRecommendations, convertRecommendation(rec))
	}

	// Convert swap recommendations
	response.SwapRecommendations = make([]*CardRecommendationResponse, 0, len(recs.SwapRecommendations))
	for _, rec := range recs.SwapRecommendations {
		response.SwapRecommendations = append(response.SwapRecommendations, convertRecommendation(rec))
	}

	return response, nil
}

// convertRecommendation converts a model recommendation to a response recommendation.
func convertRecommendation(rec *models.CardRecommendation) *CardRecommendationResponse {
	return &CardRecommendationResponse{
		Type:            rec.Type,
		CardID:          rec.CardID,
		CardName:        rec.CardName,
		Reason:          rec.Reason,
		ImpactEstimate:  rec.ImpactEstimate,
		Confidence:      rec.Confidence,
		Priority:        rec.Priority,
		SwapForCardID:   rec.SwapForCardID,
		SwapForCardName: rec.SwapForCardName,
		BasedOnGames:    rec.BasedOnGames,
	}
}

// GetUnderperformingCards returns cards that hurt deck performance.
func (d *DeckFacade) GetUnderperformingCards(ctx context.Context, deckID string, threshold float64) ([]*CardPerformanceResponse, error) {
	if deckID == "" {
		return nil, &AppError{Message: "deck_id is required"}
	}

	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	repo := d.services.Storage.CardPerformanceAnalysisRepo()
	if repo == nil {
		return nil, &AppError{Message: "Card performance repository not available"}
	}

	if threshold <= 0 {
		threshold = 0.05 // Default 5% below deck average
	}

	cards, err := repo.GetUnderperformingCards(ctx, deckID, threshold)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get underperforming cards: %v", err)}
	}

	// Convert to response format
	response := make([]*CardPerformanceResponse, 0, len(cards))
	for _, perf := range cards {
		cardPerf := &CardPerformanceResponse{
			CardID:            perf.CardID,
			CardName:          perf.CardName,
			Quantity:          perf.Quantity,
			GamesWithCard:     perf.GamesWithCard,
			GamesDrawn:        perf.GamesDrawn,
			GamesPlayed:       perf.GamesPlayed,
			WinRateWhenDrawn:  perf.WinRateWhenDrawn,
			WinRateWhenPlayed: perf.WinRateWhenPlayed,
			DeckWinRate:       perf.DeckWinRate,
			PlayRate:          perf.PlayRate,
			WinContribution:   perf.WinContribution,
			ImpactScore:       perf.ImpactScore,
			ConfidenceLevel:   perf.ConfidenceLevel,
			SampleSize:        perf.SampleSize,
			PerformanceGrade:  perf.PerformanceGrade,
			AvgTurnPlayed:     perf.AvgTurnPlayed,
			MulliganedAway:    perf.MulliganedAway,
			MulliganRate:      perf.MulliganRate,
		}
		response = append(response, cardPerf)
	}

	return response, nil
}

// GetOverperformingCards returns cards with high win impact.
func (d *DeckFacade) GetOverperformingCards(ctx context.Context, deckID string, threshold float64) ([]*CardPerformanceResponse, error) {
	if deckID == "" {
		return nil, &AppError{Message: "deck_id is required"}
	}

	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	repo := d.services.Storage.CardPerformanceAnalysisRepo()
	if repo == nil {
		return nil, &AppError{Message: "Card performance repository not available"}
	}

	if threshold <= 0 {
		threshold = 0.05 // Default 5% above deck average
	}

	cards, err := repo.GetOverperformingCards(ctx, deckID, threshold)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get overperforming cards: %v", err)}
	}

	// Convert to response format
	response := make([]*CardPerformanceResponse, 0, len(cards))
	for _, perf := range cards {
		cardPerf := &CardPerformanceResponse{
			CardID:            perf.CardID,
			CardName:          perf.CardName,
			Quantity:          perf.Quantity,
			GamesWithCard:     perf.GamesWithCard,
			GamesDrawn:        perf.GamesDrawn,
			GamesPlayed:       perf.GamesPlayed,
			WinRateWhenDrawn:  perf.WinRateWhenDrawn,
			WinRateWhenPlayed: perf.WinRateWhenPlayed,
			DeckWinRate:       perf.DeckWinRate,
			PlayRate:          perf.PlayRate,
			WinContribution:   perf.WinContribution,
			ImpactScore:       perf.ImpactScore,
			ConfidenceLevel:   perf.ConfidenceLevel,
			SampleSize:        perf.SampleSize,
			PerformanceGrade:  perf.PerformanceGrade,
			AvgTurnPlayed:     perf.AvgTurnPlayed,
			MulliganedAway:    perf.MulliganedAway,
			MulliganRate:      perf.MulliganRate,
		}
		response = append(response, cardPerf)
	}

	return response, nil
}

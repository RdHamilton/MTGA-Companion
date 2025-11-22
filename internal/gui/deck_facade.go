package gui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
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
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Format        string     `json:"format"`
	Source        string     `json:"source"`
	ColorIdentity *string    `json:"colorIdentity"`
	CardCount     int        `json:"cardCount"`
	MatchesPlayed int        `json:"matchesPlayed"`
	MatchWinRate  float64    `json:"matchWinRate"`
	ModifiedAt    time.Time  `json:"modifiedAt"`
	LastPlayed    *time.Time `json:"lastPlayed,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
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
	deck := &models.Deck{
		ID:            uuid.New().String(),
		AccountID:     accountID,
		Name:          name,
		Format:        format,
		Source:        source,
		DraftEventID:  draftEventID,
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	// Validate source
	if source != "draft" && source != "constructed" && source != "imported" {
		return nil, &AppError{Message: fmt.Sprintf("Invalid deck source: %s", source)}
	}

	// If source is draft, draft_event_id is required
	if source == "draft" && draftEventID == nil {
		return nil, &AppError{Message: "Draft event ID required for draft decks"}
	}

	err := storage.RetryOnBusy(func() error {
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

		items = append(items, &DeckListItem{
			ID:            deck.ID,
			Name:          deck.Name,
			Format:        deck.Format,
			Source:        deck.Source,
			ColorIdentity: deck.ColorIdentity,
			CardCount:     cardCount,
			MatchesPlayed: deck.MatchesPlayed,
			MatchWinRate:  winRate,
			ModifiedAt:    deck.ModifiedAt,
			LastPlayed:    deck.LastPlayed,
			Tags:          tagNames,
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

		items = append(items, &DeckListItem{
			ID:            deck.ID,
			Name:          deck.Name,
			Format:        deck.Format,
			Source:        deck.Source,
			ColorIdentity: deck.ColorIdentity,
			CardCount:     cardCount,
			MatchesPlayed: deck.MatchesPlayed,
			MatchWinRate:  winRate,
			ModifiedAt:    deck.ModifiedAt,
			LastPlayed:    deck.LastPlayed,
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

	// If this is a draft deck, validate that the card is from the draft
	if deck.Source == "draft" && deck.DraftEventID != nil {
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

	log.Printf("Removed card %d from deck %s", cardID, deckID)
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

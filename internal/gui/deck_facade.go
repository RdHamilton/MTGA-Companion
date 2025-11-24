package gui

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
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
	// Exception: Basic lands are always allowed (they have unlimited availability)
	if deck.Source == "draft" && deck.DraftEventID != nil {
		// Check if this is a basic land (basic lands are always allowed)
		basicLandIDs := map[int]bool{
			81716: true, // Plains
			81717: true, // Island
			81718: true, // Swamp
			81719: true, // Mountain
			81720: true, // Forest
		}

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

package gui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckexport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/recommendations"
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
			card, err := d.services.CardService.GetCard(deckCard.CardID)
			if err != nil {
				log.Printf("Warning: Failed to get card %d: %v", deckCard.CardID, err)
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

	// For draft decks, get the draft pool
	// TODO: Phase 1B - Integrate with draft session data to get available cards
	if deck.DraftEventID != nil {
		// For now, draft pool recommendations will be limited to cards already in the deck
		log.Printf("Info: Draft pool recommendations not yet fully integrated for deck %s", deck.ID)
	}

	// Build filters from request
	filters := &recommendations.Filters{
		MaxResults:    req.MaxResults,
		MinScore:      req.MinScore,
		Colors:        req.Colors,
		CardTypes:     req.CardTypes,
		IncludeLands:  req.IncludeLands,
		OnlyDraftPool: req.OnlyDraftPool,
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

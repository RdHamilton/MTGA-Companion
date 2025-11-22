package viewer

import (
	"context"
	"database/sql"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// DeckViewer provides deck viewing with card metadata.
type DeckViewer struct {
	deckRepo    repository.DeckRepository
	cardService *cards.Service
}

// NewDeckViewer creates a new deck viewer.
func NewDeckViewer(db *sql.DB, cardService *cards.Service) *DeckViewer {
	return &DeckViewer{
		deckRepo:    repository.NewDeckRepository(db),
		cardService: cardService,
	}
}

// GetDeck retrieves a deck with all card metadata.
func (dv *DeckViewer) GetDeck(ctx context.Context, deckID string) (*models.DeckView, error) {
	// Get deck from repository
	deck, err := dv.deckRepo.GetByID(ctx, deckID)
	if err != nil || deck == nil {
		return nil, err
	}

	// Get deck cards
	deckCards, err := dv.deckRepo.GetCards(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Extract card IDs
	cardIDs := make([]int, 0, len(deckCards))
	for _, card := range deckCards {
		cardIDs = append(cardIDs, card.CardID)
	}

	// Fetch metadata for all cards
	cardMetadata, err := dv.cardService.GetCards(cardIDs)
	if err != nil {
		cardMetadata = make(map[int]*cards.Card)
	}

	// Build deck view
	view := &models.DeckView{
		Deck:           deck,
		MainboardCards: make([]*models.DeckCardView, 0),
		SideboardCards: make([]*models.DeckCardView, 0),
		TotalMainboard: 0,
		TotalSideboard: 0,
		ColorIdentity:  make([]string, 0),
		ManaCurve:      make(map[int]int),
	}

	// Process deck cards
	colorSet := make(map[string]bool)
	for _, deckCard := range deckCards {
		cardView := &models.DeckCardView{
			ID:       deckCard.ID,
			DeckID:   deckCard.DeckID,
			CardID:   deckCard.CardID,
			Quantity: deckCard.Quantity,
			Board:    deckCard.Board,
			Metadata: cardMetadata[deckCard.CardID],
		}

		switch deckCard.Board {
		case "main":
			view.MainboardCards = append(view.MainboardCards, cardView)
			view.TotalMainboard += deckCard.Quantity

			// Calculate mana curve
			if cardView.Metadata != nil {
				cmc := int(cardView.Metadata.CMC)
				view.ManaCurve[cmc] += deckCard.Quantity

				// Track color identity
				for _, color := range cardView.Metadata.ColorIdentity {
					colorSet[color] = true
				}
			}
		case "sideboard":
			view.SideboardCards = append(view.SideboardCards, cardView)
			view.TotalSideboard += deckCard.Quantity
		}
	}

	// Set color identity
	for color := range colorSet {
		view.ColorIdentity = append(view.ColorIdentity, color)
	}

	return view, nil
}

// ListDecks retrieves all decks with basic information for an account.
func (dv *DeckViewer) ListDecks(ctx context.Context, accountID int) ([]*models.Deck, error) {
	return dv.deckRepo.List(ctx, accountID)
}

// GetDecksByFormat retrieves all decks for a specific format and account.
func (dv *DeckViewer) GetDecksByFormat(ctx context.Context, accountID int, format string) ([]*models.Deck, error) {
	return dv.deckRepo.GetByFormat(ctx, accountID, format)
}

// GetDeckCards retrieves just the cards in a deck with metadata.
func (dv *DeckViewer) GetDeckCards(ctx context.Context, deckID string) ([]*models.DeckCardView, error) {
	// Get deck cards
	deckCards, err := dv.deckRepo.GetCards(ctx, deckID)
	if err != nil {
		return nil, err
	}

	// Extract card IDs
	cardIDs := make([]int, 0, len(deckCards))
	for _, card := range deckCards {
		cardIDs = append(cardIDs, card.CardID)
	}

	// Fetch metadata
	cardMetadata, err := dv.cardService.GetCards(cardIDs)
	if err != nil {
		cardMetadata = make(map[int]*cards.Card)
	}

	// Build card views
	views := make([]*models.DeckCardView, 0, len(deckCards))
	for _, deckCard := range deckCards {
		view := &models.DeckCardView{
			ID:       deckCard.ID,
			DeckID:   deckCard.DeckID,
			CardID:   deckCard.CardID,
			Quantity: deckCard.Quantity,
			Board:    deckCard.Board,
			Metadata: cardMetadata[deckCard.CardID],
		}
		views = append(views, view)
	}

	return views, nil
}

// GetMainboard retrieves only the mainboard cards with metadata.
func (dv *DeckViewer) GetMainboard(ctx context.Context, deckID string) ([]*models.DeckCardView, error) {
	cards, err := dv.GetDeckCards(ctx, deckID)
	if err != nil {
		return nil, err
	}

	mainboard := make([]*models.DeckCardView, 0)
	for _, card := range cards {
		if card.Board == "main" {
			mainboard = append(mainboard, card)
		}
	}

	return mainboard, nil
}

// GetSideboard retrieves only the sideboard cards with metadata.
func (dv *DeckViewer) GetSideboard(ctx context.Context, deckID string) ([]*models.DeckCardView, error) {
	cards, err := dv.GetDeckCards(ctx, deckID)
	if err != nil {
		return nil, err
	}

	sideboard := make([]*models.DeckCardView, 0)
	for _, card := range cards {
		if card.Board == "sideboard" {
			sideboard = append(sideboard, card)
		}
	}

	return sideboard, nil
}

// AnalyzeDeck provides detailed deck analysis.
type DeckAnalysis struct {
	TotalCards      int
	Creatures       int
	Instants        int
	Sorceries       int
	Artifacts       int
	Enchantments    int
	Planeswalkers   int
	Lands           int
	AverageCMC      float64
	ColorBreakdown  map[string]int
	RarityBreakdown map[string]int
}

// AnalyzeDeck analyzes a deck's composition.
func (dv *DeckViewer) AnalyzeDeck(ctx context.Context, deckID string) (*DeckAnalysis, error) {
	deckView, err := dv.GetDeck(ctx, deckID)
	if err != nil {
		return nil, err
	}

	analysis := &DeckAnalysis{
		ColorBreakdown:  make(map[string]int),
		RarityBreakdown: make(map[string]int),
	}

	var totalCMC float64
	var nonLandCount int

	for _, cardView := range deckView.MainboardCards {
		if cardView.Metadata == nil {
			continue
		}

		quantity := cardView.Quantity
		analysis.TotalCards += quantity

		// Type breakdown
		typeLine := cardView.Metadata.TypeLine
		if contains(typeLine, "Creature") {
			analysis.Creatures += quantity
		}
		if contains(typeLine, "Instant") {
			analysis.Instants += quantity
		}
		if contains(typeLine, "Sorcery") {
			analysis.Sorceries += quantity
		}
		if contains(typeLine, "Artifact") {
			analysis.Artifacts += quantity
		}
		if contains(typeLine, "Enchantment") {
			analysis.Enchantments += quantity
		}
		if contains(typeLine, "Planeswalker") {
			analysis.Planeswalkers += quantity
		}
		if contains(typeLine, "Land") {
			analysis.Lands += quantity
		} else {
			// Calculate average CMC for non-lands
			totalCMC += cardView.Metadata.CMC * float64(quantity)
			nonLandCount += quantity
		}

		// Color breakdown
		for _, color := range cardView.Metadata.Colors {
			analysis.ColorBreakdown[color] += quantity
		}

		// Rarity breakdown
		analysis.RarityBreakdown[cardView.Metadata.Rarity] += quantity
	}

	// Calculate average CMC
	if nonLandCount > 0 {
		analysis.AverageCMC = totalCMC / float64(nonLandCount)
	}

	return analysis, nil
}

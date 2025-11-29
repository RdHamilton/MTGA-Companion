package gui

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CollectionFacade handles all collection-related operations for the GUI.
type CollectionFacade struct {
	services *Services
}

// NewCollectionFacade creates a new CollectionFacade with the given services.
func NewCollectionFacade(services *Services) *CollectionFacade {
	return &CollectionFacade{
		services: services,
	}
}

// CollectionCard represents a card in the collection with metadata for the frontend.
type CollectionCard struct {
	CardID        int      `json:"cardId"`
	ArenaID       int      `json:"arenaId"`
	Quantity      int      `json:"quantity"`
	Name          string   `json:"name"`
	SetCode       string   `json:"setCode"`
	SetName       string   `json:"setName"`
	Rarity        string   `json:"rarity"`
	ManaCost      string   `json:"manaCost"`
	CMC           float64  `json:"cmc"`
	TypeLine      string   `json:"typeLine"`
	Colors        []string `json:"colors"`
	ColorIdentity []string `json:"colorIdentity"`
	ImageURI      string   `json:"imageUri"`
	Power         string   `json:"power,omitempty"`
	Toughness     string   `json:"toughness,omitempty"`
}

// CollectionFilter specifies filter criteria for collection queries.
type CollectionFilter struct {
	SearchTerm string   `json:"searchTerm,omitempty"` // Filter by name (case-insensitive)
	SetCode    string   `json:"setCode,omitempty"`    // Filter by set code
	Rarity     string   `json:"rarity,omitempty"`     // Filter by rarity (common, uncommon, rare, mythic)
	Colors     []string `json:"colors,omitempty"`     // Filter by colors (W, U, B, R, G)
	CardType   string   `json:"cardType,omitempty"`   // Filter by type line (creature, instant, etc.)
	OwnedOnly  bool     `json:"ownedOnly"`            // Only show cards with quantity > 0
	SortBy     string   `json:"sortBy,omitempty"`     // Sort field: name, quantity, rarity, cmc, setCode
	SortDesc   bool     `json:"sortDesc"`             // Sort descending
	Limit      int      `json:"limit,omitempty"`      // Maximum results
	Offset     int      `json:"offset,omitempty"`     // Offset for pagination
}

// CollectionResponse contains collection data with pagination info.
type CollectionResponse struct {
	Cards       []*CollectionCard `json:"cards"`
	TotalCount  int               `json:"totalCount"`
	FilterCount int               `json:"filterCount"` // Count after filters but before pagination
}

// CollectionStats provides summary statistics about the collection.
type CollectionStats struct {
	TotalUniqueCards int `json:"totalUniqueCards"` // Number of unique cards owned
	TotalCards       int `json:"totalCards"`       // Total cards including multiples
	CommonCount      int `json:"commonCount"`
	UncommonCount    int `json:"uncommonCount"`
	RareCount        int `json:"rareCount"`
	MythicCount      int `json:"mythicCount"`
}

// GetCollection returns the player's collection with optional filtering.
func (c *CollectionFacade) GetCollection(ctx context.Context, filter *CollectionFilter) (*CollectionResponse, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get raw collection data (cardID -> quantity)
	var collection map[int]int
	err := storage.RetryOnBusy(func() error {
		var err error
		collection, err = c.services.Storage.GetCollection(ctx)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get collection: %v", err)}
	}

	// Track which cards we've already added
	addedCards := make(map[int]bool)
	collectionCards := make([]*CollectionCard, 0)

	// If filtering by set and not owned only, get all cards from downloaded set(s)
	if filter != nil && !filter.OwnedOnly {
		var setCards []*models.SetCard

		if filter.SetCode != "" {
			// Get cards from specific set
			setCards, err = c.services.Storage.SetCardRepo().GetCardsBySet(ctx, filter.SetCode)
			if err != nil {
				return nil, &AppError{Message: fmt.Sprintf("Failed to get set cards: %v", err)}
			}
		} else {
			// Get all cards from all downloaded sets
			setCodes, err := c.services.Storage.SetCardRepo().GetCachedSets(ctx)
			if err != nil {
				return nil, &AppError{Message: fmt.Sprintf("Failed to get sets: %v", err)}
			}
			for _, setCode := range setCodes {
				cards, err := c.services.Storage.SetCardRepo().GetCardsBySet(ctx, setCode)
				if err == nil {
					setCards = append(setCards, cards...)
				}
			}
		}

		// Build cards from set data with collection quantities
		for _, meta := range setCards {
			arenaID, _ := strconv.Atoi(meta.ArenaID)
			if arenaID == 0 {
				continue // Skip cards without arena IDs
			}

			quantity := collection[arenaID] // Will be 0 if not owned

			card := &CollectionCard{
				CardID:    arenaID,
				ArenaID:   arenaID,
				Quantity:  quantity,
				Name:      meta.Name,
				SetCode:   meta.SetCode,
				Rarity:    meta.Rarity,
				ManaCost:  meta.ManaCost,
				CMC:       float64(meta.CMC),
				TypeLine:  strings.Join(meta.Types, " "),
				Colors:    meta.Colors,
				ImageURI:  meta.ImageURL,
				Power:     meta.Power,
				Toughness: meta.Toughness,
			}

			if card.ImageURI == "" {
				card.ImageURI = "https://cards.scryfall.io/back.png"
			}

			collectionCards = append(collectionCards, card)
			addedCards[arenaID] = true
		}
	}

	// Add owned cards that aren't already in the list (either all owned cards if ownedOnly,
	// or just cards from sets we don't have downloaded)
	for cardID, quantity := range collection {
		if addedCards[cardID] {
			continue
		}

		card := &CollectionCard{
			CardID:   cardID,
			ArenaID:  cardID,
			Quantity: quantity,
		}

		// Try to get metadata from SetCardRepo
		arenaID := fmt.Sprintf("%d", cardID)
		meta, err := c.services.Storage.SetCardRepo().GetCardByArenaID(ctx, arenaID)
		if err == nil && meta != nil {
			card.Name = meta.Name
			card.SetCode = meta.SetCode
			card.Rarity = meta.Rarity
			card.ManaCost = meta.ManaCost
			card.CMC = float64(meta.CMC)
			card.TypeLine = strings.Join(meta.Types, " ")
			card.Colors = meta.Colors
			card.ImageURI = meta.ImageURL
			if meta.Power != "" {
				card.Power = meta.Power
			}
			if meta.Toughness != "" {
				card.Toughness = meta.Toughness
			}
		}

		if card.ImageURI == "" {
			card.ImageURI = "https://cards.scryfall.io/back.png"
		}

		collectionCards = append(collectionCards, card)
	}

	totalCount := len(collectionCards)

	// Apply filters
	if filter != nil {
		collectionCards = applyCollectionFilters(collectionCards, filter)
	}

	filterCount := len(collectionCards)

	// Apply sorting
	if filter != nil && filter.SortBy != "" {
		sortCollectionCards(collectionCards, filter.SortBy, filter.SortDesc)
	} else {
		// Default sort by name
		sortCollectionCards(collectionCards, "name", false)
	}

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(collectionCards) {
			collectionCards = collectionCards[filter.Offset:]
		} else if filter.Offset >= len(collectionCards) {
			collectionCards = []*CollectionCard{}
		}

		if filter.Limit > 0 && filter.Limit < len(collectionCards) {
			collectionCards = collectionCards[:filter.Limit]
		}
	}

	return &CollectionResponse{
		Cards:       collectionCards,
		TotalCount:  totalCount,
		FilterCount: filterCount,
	}, nil
}

// GetCollectionStats returns summary statistics about the collection.
func (c *CollectionFacade) GetCollectionStats(ctx context.Context) (*CollectionStats, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get collection with full data
	response, err := c.GetCollection(ctx, nil)
	if err != nil {
		return nil, err
	}

	stats := &CollectionStats{}
	for _, card := range response.Cards {
		if card.Quantity > 0 {
			stats.TotalUniqueCards++
			stats.TotalCards += card.Quantity

			switch strings.ToLower(card.Rarity) {
			case "common":
				stats.CommonCount += card.Quantity
			case "uncommon":
				stats.UncommonCount += card.Quantity
			case "rare":
				stats.RareCount += card.Quantity
			case "mythic":
				stats.MythicCount += card.Quantity
			}
		}
	}

	return stats, nil
}

// GetSetCompletion returns set completion statistics.
func (c *CollectionFacade) GetSetCompletion(ctx context.Context) ([]*models.SetCompletion, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var completion []*models.SetCompletion
	err := storage.RetryOnBusy(func() error {
		var err error
		completion, err = c.services.Storage.GetSetCompletion(ctx)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get set completion: %v", err)}
	}

	return completion, nil
}

// GetRecentChanges returns recent collection changes.
func (c *CollectionFacade) GetRecentChanges(ctx context.Context, limit int) ([]*CollectionChangeEntry, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if limit <= 0 {
		limit = 50
	}

	var changes []*storage.CollectionHistory
	err := storage.RetryOnBusy(func() error {
		var err error
		changes, err = c.services.Storage.GetRecentCollectionChanges(ctx, limit)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get collection changes: %v", err)}
	}

	// Convert to frontend format with card names
	result := make([]*CollectionChangeEntry, 0, len(changes))
	for _, change := range changes {
		entry := &CollectionChangeEntry{
			CardID:        change.CardID,
			QuantityDelta: change.QuantityDelta,
			QuantityAfter: change.QuantityAfter,
			Timestamp:     change.Timestamp.Unix(),
		}
		if change.Source != nil {
			entry.Source = *change.Source
		}

		// Try to get card name
		arenaID := fmt.Sprintf("%d", change.CardID)
		card, err := c.services.Storage.SetCardRepo().GetCardByArenaID(ctx, arenaID)
		if err == nil && card != nil {
			entry.CardName = card.Name
			entry.SetCode = card.SetCode
			entry.Rarity = card.Rarity
		}

		result = append(result, entry)
	}

	return result, nil
}

// CollectionChangeEntry represents a single collection change for the frontend.
type CollectionChangeEntry struct {
	CardID        int    `json:"cardId"`
	CardName      string `json:"cardName,omitempty"`
	SetCode       string `json:"setCode,omitempty"`
	Rarity        string `json:"rarity,omitempty"`
	QuantityDelta int    `json:"quantityDelta"`
	QuantityAfter int    `json:"quantityAfter"`
	Timestamp     int64  `json:"timestamp"` // Unix timestamp
	Source        string `json:"source,omitempty"`
}

// applyCollectionFilters applies filter criteria to a collection of cards.
func applyCollectionFilters(cards []*CollectionCard, filter *CollectionFilter) []*CollectionCard {
	if filter == nil {
		return cards
	}

	result := make([]*CollectionCard, 0, len(cards))

	for _, card := range cards {
		// OwnedOnly filter
		if filter.OwnedOnly && card.Quantity <= 0 {
			continue
		}

		// Search term filter (case-insensitive name match)
		if filter.SearchTerm != "" {
			if !strings.Contains(strings.ToLower(card.Name), strings.ToLower(filter.SearchTerm)) {
				continue
			}
		}

		// Set code filter
		if filter.SetCode != "" && !strings.EqualFold(card.SetCode, filter.SetCode) {
			continue
		}

		// Rarity filter
		if filter.Rarity != "" && !strings.EqualFold(card.Rarity, filter.Rarity) {
			continue
		}

		// Color filter (card must have at least one of the specified colors)
		if len(filter.Colors) > 0 {
			hasColor := false
			for _, filterColor := range filter.Colors {
				for _, cardColor := range card.Colors {
					if strings.EqualFold(cardColor, filterColor) {
						hasColor = true
						break
					}
				}
				if hasColor {
					break
				}
			}
			if !hasColor {
				continue
			}
		}

		// Card type filter (case-insensitive type line match)
		if filter.CardType != "" {
			if !strings.Contains(strings.ToLower(card.TypeLine), strings.ToLower(filter.CardType)) {
				continue
			}
		}

		result = append(result, card)
	}

	return result
}

// sortCollectionCards sorts cards by the specified field.
func sortCollectionCards(cards []*CollectionCard, sortBy string, desc bool) {
	sort.Slice(cards, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = strings.ToLower(cards[i].Name) < strings.ToLower(cards[j].Name)
		case "quantity":
			less = cards[i].Quantity < cards[j].Quantity
		case "rarity":
			less = rarityOrder(cards[i].Rarity) < rarityOrder(cards[j].Rarity)
		case "cmc":
			less = cards[i].CMC < cards[j].CMC
		case "setCode":
			less = strings.ToLower(cards[i].SetCode) < strings.ToLower(cards[j].SetCode)
		default:
			less = strings.ToLower(cards[i].Name) < strings.ToLower(cards[j].Name)
		}

		if desc {
			return !less
		}
		return less
	})
}

// rarityOrder returns a numeric value for sorting by rarity.
func rarityOrder(rarity string) int {
	switch strings.ToLower(rarity) {
	case "common":
		return 1
	case "uncommon":
		return 2
	case "rare":
		return 3
	case "mythic":
		return 4
	default:
		return 0
	}
}

// MissingCard represents a card that is missing from the player's collection.
type MissingCard struct {
	CardID         int      `json:"cardId"`
	ArenaID        int      `json:"arenaId"`
	Name           string   `json:"name"`
	SetCode        string   `json:"setCode"`
	SetName        string   `json:"setName"`
	Rarity         string   `json:"rarity"`
	ManaCost       string   `json:"manaCost"`
	CMC            float64  `json:"cmc"`
	TypeLine       string   `json:"typeLine"`
	Colors         []string `json:"colors"`
	ImageURI       string   `json:"imageUri"`
	QuantityNeeded int      `json:"quantityNeeded"` // How many copies are needed
	QuantityOwned  int      `json:"quantityOwned"`  // How many copies owned
}

// MissingCardsForDeckResponse contains missing cards analysis for a deck.
type MissingCardsForDeckResponse struct {
	DeckID          string         `json:"deckId"`
	DeckName        string         `json:"deckName"`
	TotalMissing    int            `json:"totalMissing"`    // Total number of missing card copies
	UniqueMissing   int            `json:"uniqueMissing"`   // Number of unique missing cards
	MissingCards    []*MissingCard `json:"missingCards"`    // List of missing cards, sorted by rarity (mythic first)
	WildcardsNeeded *WildcardCost  `json:"wildcardsNeeded"` // Wildcard cost breakdown
	IsComplete      bool           `json:"isComplete"`      // True if deck can be built with current collection
}

// MissingCardsForSetResponse contains missing cards analysis for a set.
type MissingCardsForSetResponse struct {
	SetCode         string         `json:"setCode"`
	SetName         string         `json:"setName"`
	TotalMissing    int            `json:"totalMissing"`    // Total missing card copies
	UniqueMissing   int            `json:"uniqueMissing"`   // Unique missing cards
	MissingCards    []*MissingCard `json:"missingCards"`    // List of missing cards
	WildcardsNeeded *WildcardCost  `json:"wildcardsNeeded"` // Wildcard cost breakdown
	CompletionPct   float64        `json:"completionPct"`   // Set completion percentage
}

// WildcardCost represents the wildcard cost to complete a deck or set.
type WildcardCost struct {
	Common   int `json:"common"`
	Uncommon int `json:"uncommon"`
	Rare     int `json:"rare"`
	Mythic   int `json:"mythic"`
	Total    int `json:"total"`
}

// GetMissingCardsForDeck returns missing cards analysis for a specific deck.
func (c *CollectionFacade) GetMissingCardsForDeck(ctx context.Context, deckID string) (*MissingCardsForDeckResponse, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get the deck
	deck, err := c.services.Storage.DeckRepo().GetByID(ctx, deckID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck: %v", err)}
	}
	if deck == nil {
		return nil, &AppError{Message: "Deck not found"}
	}

	// Get deck cards
	deckCards, err := c.services.Storage.DeckRepo().GetCards(ctx, deckID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get deck cards: %v", err)}
	}

	// Get player's collection
	var collection map[int]int
	err = storage.RetryOnBusy(func() error {
		var err error
		collection, err = c.services.Storage.GetCollection(ctx)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get collection: %v", err)}
	}

	// Calculate missing cards
	missingCards := make([]*MissingCard, 0)
	wildcardCost := &WildcardCost{}
	totalMissing := 0

	for _, deckCard := range deckCards {
		owned := collection[deckCard.CardID]
		needed := deckCard.Quantity - owned

		if needed > 0 {
			totalMissing += needed

			// Get card metadata
			arenaID := fmt.Sprintf("%d", deckCard.CardID)
			cardMeta, _ := c.services.Storage.SetCardRepo().GetCardByArenaID(ctx, arenaID)

			missing := &MissingCard{
				CardID:         deckCard.CardID,
				ArenaID:        deckCard.CardID,
				QuantityNeeded: needed,
				QuantityOwned:  owned,
			}

			if cardMeta != nil {
				missing.Name = cardMeta.Name
				missing.SetCode = cardMeta.SetCode
				missing.Rarity = cardMeta.Rarity
				missing.ManaCost = cardMeta.ManaCost
				missing.CMC = float64(cardMeta.CMC)
				missing.TypeLine = strings.Join(cardMeta.Types, " ")
				missing.Colors = cardMeta.Colors
				missing.ImageURI = cardMeta.ImageURL

				// Add to wildcard cost
				switch strings.ToLower(cardMeta.Rarity) {
				case "common":
					wildcardCost.Common += needed
				case "uncommon":
					wildcardCost.Uncommon += needed
				case "rare":
					wildcardCost.Rare += needed
				case "mythic":
					wildcardCost.Mythic += needed
				}
			}

			missingCards = append(missingCards, missing)
		}
	}

	// Sort missing cards by rarity (mythic first, then rare, uncommon, common)
	sort.Slice(missingCards, func(i, j int) bool {
		return rarityOrder(missingCards[i].Rarity) > rarityOrder(missingCards[j].Rarity)
	})

	wildcardCost.Total = wildcardCost.Common + wildcardCost.Uncommon + wildcardCost.Rare + wildcardCost.Mythic

	return &MissingCardsForDeckResponse{
		DeckID:          deckID,
		DeckName:        deck.Name,
		TotalMissing:    totalMissing,
		UniqueMissing:   len(missingCards),
		MissingCards:    missingCards,
		WildcardsNeeded: wildcardCost,
		IsComplete:      totalMissing == 0,
	}, nil
}

// GetMissingCardsForSet returns missing cards analysis for a specific set.
func (c *CollectionFacade) GetMissingCardsForSet(ctx context.Context, setCode string) (*MissingCardsForSetResponse, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get all cards in the set
	setCards, err := c.services.Storage.SetCardRepo().GetCardsBySet(ctx, setCode)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get set cards: %v", err)}
	}
	if len(setCards) == 0 {
		return nil, &AppError{Message: fmt.Sprintf("No cards found for set: %s", setCode)}
	}

	// Get player's collection
	var collection map[int]int
	err = storage.RetryOnBusy(func() error {
		var err error
		collection, err = c.services.Storage.GetCollection(ctx)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get collection: %v", err)}
	}

	// Calculate missing cards (max 4 copies of each)
	missingCards := make([]*MissingCard, 0)
	wildcardCost := &WildcardCost{}
	totalMissing := 0
	totalPossible := 0
	totalOwned := 0
	setName := ""

	for _, card := range setCards {
		// Get ArenaID as int for collection lookup
		arenaID := 0
		if card.ArenaID != "" {
			_, _ = fmt.Sscanf(card.ArenaID, "%d", &arenaID)
		}

		// Skip basic lands for completion calculation (unlimited copies allowed)
		isBasicLand := strings.Contains(strings.ToLower(card.Rarity), "basic") ||
			(strings.Contains(strings.Join(card.Types, " "), "Basic") && strings.Contains(strings.Join(card.Types, " "), "Land"))

		maxCopies := 4
		if isBasicLand {
			continue // Skip basic lands
		}

		owned := collection[arenaID]
		if owned > maxCopies {
			owned = maxCopies
		}

		totalPossible += maxCopies
		totalOwned += owned

		needed := maxCopies - owned
		if needed > 0 {
			totalMissing += needed

			missing := &MissingCard{
				CardID:         arenaID,
				ArenaID:        arenaID,
				Name:           card.Name,
				SetCode:        card.SetCode,
				Rarity:         card.Rarity,
				ManaCost:       card.ManaCost,
				CMC:            float64(card.CMC),
				TypeLine:       strings.Join(card.Types, " "),
				Colors:         card.Colors,
				ImageURI:       card.ImageURL,
				QuantityNeeded: needed,
				QuantityOwned:  owned,
			}

			missingCards = append(missingCards, missing)

			// Add to wildcard cost
			switch strings.ToLower(card.Rarity) {
			case "common":
				wildcardCost.Common += needed
			case "uncommon":
				wildcardCost.Uncommon += needed
			case "rare":
				wildcardCost.Rare += needed
			case "mythic":
				wildcardCost.Mythic += needed
			}
		}

		// Capture set name
		if setName == "" {
			setName = card.SetCode // Use set code as fallback
		}
	}

	// Sort missing cards by rarity (mythic first)
	sort.Slice(missingCards, func(i, j int) bool {
		return rarityOrder(missingCards[i].Rarity) > rarityOrder(missingCards[j].Rarity)
	})

	wildcardCost.Total = wildcardCost.Common + wildcardCost.Uncommon + wildcardCost.Rare + wildcardCost.Mythic

	// Calculate completion percentage
	completionPct := 0.0
	if totalPossible > 0 {
		completionPct = float64(totalOwned) / float64(totalPossible) * 100
	}

	return &MissingCardsForSetResponse{
		SetCode:         setCode,
		SetName:         setName,
		TotalMissing:    totalMissing,
		UniqueMissing:   len(missingCards),
		MissingCards:    missingCards,
		WildcardsNeeded: wildcardCost,
		CompletionPct:   completionPct,
	}, nil
}

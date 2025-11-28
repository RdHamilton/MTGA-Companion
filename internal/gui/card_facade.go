package gui

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CardFacade handles all card data operations including set cards and ratings.
type CardFacade struct {
	services *Services
}

// NewCardFacade creates a new CardFacade with the given services.
func NewCardFacade(services *Services) *CardFacade {
	return &CardFacade{
		services: services,
	}
}

// CardRatingWithTier extends CardRating with tier and colors information.
type CardRatingWithTier struct {
	seventeenlands.CardRating
	Tier   string   `json:"tier"`   // S, A, B, C, D, or F
	Colors []string `json:"colors"` // All colors in mana cost (e.g., ["W", "U"])
}

// SetInfo contains information about a Magic set including the icon URL.
// Used by the frontend to display set symbols.
type SetInfo struct {
	Code       string `json:"code"`       // Set code (e.g., "DSK", "BLB")
	Name       string `json:"name"`       // Full set name (e.g., "Duskmourn: House of Horror")
	IconSVGURI string `json:"iconSvgUri"` // URL to the set symbol SVG
	SetType    string `json:"setType"`    // Type of set (e.g., "expansion", "core")
	ReleasedAt string `json:"releasedAt"` // Release date
	CardCount  int    `json:"cardCount"`  // Number of cards in set
}

// GetSetCards returns all cards for a set, fetching from Scryfall if not cached.
func (c *CardFacade) GetSetCards(ctx context.Context, setCode string) ([]*models.SetCard, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Check if set is already cached (with retry)
	var isCached bool
	err := storage.RetryOnBusy(func() error {
		var err error
		isCached, err = c.services.Storage.SetCardRepo().IsSetCached(ctx, setCode)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to check set cache: %v", err)}
	}

	// If not cached, fetch from Scryfall
	if !isCached {
		log.Printf("Set %s not cached, fetching from Scryfall...", setCode)
		count, err := c.services.SetFetcher.FetchAndCacheSet(ctx, setCode)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Failed to fetch set cards from Scryfall: %v", err)}
		}
		log.Printf("Fetched and cached %d cards for set %s", count, setCode)
	}

	var cards []*models.SetCard
	err = storage.RetryOnBusy(func() error {
		var err error
		cards, err = c.services.Storage.SetCardRepo().GetCardsBySet(ctx, setCode)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get set cards: %v", err)}
	}

	return cards, nil
}

// FetchSetCards manually fetches and caches set cards from Scryfall.
// Returns the number of cards fetched and cached.
func (c *CardFacade) FetchSetCards(ctx context.Context, setCode string) (int, error) {
	if c.services.Storage == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	log.Printf("Manually fetching set %s from Scryfall...", setCode)
	count, err := c.services.SetFetcher.FetchAndCacheSet(ctx, setCode)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to fetch set cards: %v", err)}
	}

	log.Printf("Successfully fetched and cached %d cards for set %s", count, setCode)
	return count, nil
}

// RefreshSetCards deletes and re-fetches all cards for a set.
func (c *CardFacade) RefreshSetCards(ctx context.Context, setCode string) (int, error) {
	if c.services.Storage == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	log.Printf("Refreshing set %s from Scryfall...", setCode)
	count, err := c.services.SetFetcher.RefreshSet(ctx, setCode)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to refresh set cards: %v", err)}
	}

	log.Printf("Successfully refreshed %d cards for set %s", count, setCode)
	return count, nil
}

// FetchSetRatings fetches and caches 17Lands card ratings for a set and draft format.
func (c *CardFacade) FetchSetRatings(ctx context.Context, setCode string, draftFormat string) error {
	if c.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	if c.services.RatingsFetcher == nil {
		return &AppError{Message: "Ratings fetcher not initialized"}
	}

	log.Printf("Fetching 17Lands ratings for set %s, format %s...", setCode, draftFormat)
	err := c.services.RatingsFetcher.FetchAndCacheRatings(ctx, setCode, draftFormat)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to fetch ratings: %v", err)}
	}

	log.Printf("Successfully fetched and cached ratings for set %s, format %s", setCode, draftFormat)
	return nil
}

// RefreshSetRatings deletes and re-fetches 17Lands ratings for a set and draft format.
func (c *CardFacade) RefreshSetRatings(ctx context.Context, setCode string, draftFormat string) error {
	if c.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	if c.services.RatingsFetcher == nil {
		return &AppError{Message: "Ratings fetcher not initialized"}
	}

	log.Printf("Refreshing 17Lands ratings for set %s, format %s...", setCode, draftFormat)
	err := c.services.RatingsFetcher.RefreshRatings(ctx, setCode, draftFormat)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to refresh ratings: %v", err)}
	}

	log.Printf("Successfully refreshed ratings for set %s, format %s", setCode, draftFormat)
	return nil
}

// ClearDatasetCache clears all cached 17Lands datasets to free up disk space.
// This removes the locally cached CSV files but keeps the ratings in the database.
func (c *CardFacade) ClearDatasetCache(ctx context.Context) error {
	if c.services.DatasetService == nil {
		// No dataset service means legacy API mode - nothing to clear
		log.Println("No dataset cache to clear (using legacy API mode)")
		return nil
	}

	log.Println("Clearing 17Lands dataset cache...")
	err := c.services.DatasetService.ClearCache()
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to clear dataset cache: %v", err)}
	}

	log.Println("Successfully cleared dataset cache")
	return nil
}

// GetDatasetSource returns the data source for a given set and format ("s3" or "web_api").
// Returns "unknown" if dataset service is not available.
func (c *CardFacade) GetDatasetSource(ctx context.Context, setCode string, draftFormat string) string {
	if c.services.DatasetService == nil {
		return "legacy_api"
	}

	source := c.services.DatasetService.GetDataSource(ctx, setCode, draftFormat)
	return source
}

// GetCardByArenaID returns a card by its Arena ID.
func (c *CardFacade) GetCardByArenaID(ctx context.Context, arenaID string) (*models.SetCard, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	log.Printf("[GetCardByArenaID] Looking up card with ArenaID: %s", arenaID)

	card, err := c.services.Storage.SetCardRepo().GetCardByArenaID(ctx, arenaID)
	if err != nil {
		log.Printf("[GetCardByArenaID] Error looking up card %s: %v", arenaID, err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card: %v", err)}
	}

	if card == nil {
		log.Printf("[GetCardByArenaID] Card %s not found in database", arenaID)
		return nil, nil
	}

	log.Printf("[GetCardByArenaID] Found card %s: Name=%s", arenaID, card.Name)
	return card, nil
}

// GetCardRatings returns all card ratings for a set and draft format with tier information.
func (c *CardFacade) GetCardRatings(ctx context.Context, setCode string, draftFormat string) ([]CardRatingWithTier, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get ratings from repository
	ratings, _, err := c.services.Storage.DraftRatingsRepo().GetCardRatings(ctx, setCode, draftFormat)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card ratings: %v", err)}
	}

	// Build a map of arena ID to colors by fetching set cards
	arenaIDToColors := make(map[string][]string)
	for _, rating := range ratings {
		if rating.MTGAID != 0 {
			arenaID := fmt.Sprintf("%d", rating.MTGAID)
			card, err := c.services.Storage.SetCardRepo().GetCardByArenaID(ctx, arenaID)
			if err == nil && card != nil && len(card.Colors) > 0 {
				arenaIDToColors[arenaID] = card.Colors
			}
		}
	}

	// Add tier and colors to each rating
	result := make([]CardRatingWithTier, len(ratings))
	for i, rating := range ratings {
		arenaID := fmt.Sprintf("%d", rating.MTGAID)
		colors := arenaIDToColors[arenaID]
		// If no colors found in set_cards, fall back to single color from rating
		if len(colors) == 0 && rating.Color != "" && rating.Color != "C" {
			colors = []string{rating.Color}
		}
		result[i] = CardRatingWithTier{
			CardRating: rating,
			Tier:       calculateTier(rating.GIHWR),
			Colors:     colors,
		}
	}

	return result, nil
}

// GetCardRatingByArenaID returns the 17Lands rating for a specific card.
func (c *CardFacade) GetCardRatingByArenaID(ctx context.Context, setCode string, draftFormat string, arenaID string) (*CardRatingWithTier, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	rating, err := c.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, setCode, draftFormat, arenaID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card rating: %v", err)}
	}

	if rating == nil {
		return nil, nil
	}

	return &CardRatingWithTier{
		CardRating: *rating,
		Tier:       calculateTier(rating.GIHWR),
	}, nil
}

// GetColorRatings returns 17Lands color combination ratings for a set and draft format.
func (c *CardFacade) GetColorRatings(ctx context.Context, setCode string, draftFormat string) ([]seventeenlands.ColorRating, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	ratings, _, err := c.services.Storage.DraftRatingsRepo().GetColorRatings(ctx, setCode, draftFormat)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get color ratings: %v", err)}
	}

	return ratings, nil
}

// calculateTier determines the tier (S, A, B, C, D, F) based on GIHWR percentage.
// Tier thresholds:
// - S Tier (Bombs): GIHWR ≥ 60% - Format-defining cards
// - A Tier: 57-59% - Excellent cards, high picks
// - B Tier: 54-56% - Good playables
// - C Tier: 51-53% - Filler/role players
// - D Tier: 48-50% - Below average
// - F Tier: < 48% - Avoid/sideboard
func calculateTier(gihwr float64) string {
	if gihwr >= 60 {
		return "S"
	}
	if gihwr >= 57 {
		return "A"
	}
	if gihwr >= 54 {
		return "B"
	}
	if gihwr >= 51 {
		return "C"
	}
	if gihwr >= 48 {
		return "D"
	}
	return "F"
}

// GetSetInfo returns information about a specific set including its icon URL.
func (c *CardFacade) GetSetInfo(ctx context.Context, setCode string) (*SetInfo, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	set, err := c.services.Storage.GetSet(ctx, setCode)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get set info: %v", err)}
	}

	if set == nil {
		return nil, nil
	}

	return &SetInfo{
		Code:       set.Code,
		Name:       set.Name,
		IconSVGURI: set.IconSVGURI,
		SetType:    set.SetType,
		ReleasedAt: set.ReleasedAt,
		CardCount:  set.CardCount,
	}, nil
}

// SearchCards searches for cards by name across all cached sets.
// If setCodes is empty or nil, searches all sets.
// Returns up to limit results (default 50, max 200).
func (c *CardFacade) SearchCards(ctx context.Context, query string, setCodes []string, limit int) ([]*models.SetCard, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if query == "" {
		return []*models.SetCard{}, nil
	}

	var cards []*models.SetCard
	err := storage.RetryOnBusy(func() error {
		var err error
		cards, err = c.services.Storage.SetCardRepo().SearchCards(ctx, query, setCodes, limit)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to search cards: %v", err)}
	}

	return cards, nil
}

// CardWithOwned represents a SetCard with collection ownership information.
type CardWithOwned struct {
	*models.SetCard
	OwnedQuantity int `json:"ownedQuantity"`
}

// SearchCardsWithCollection searches for cards and includes collection ownership information.
// If collectionOnly is true, only returns cards that are in the collection.
func (c *CardFacade) SearchCardsWithCollection(ctx context.Context, query string, setCodes []string, limit int, collectionOnly bool) ([]*CardWithOwned, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if query == "" {
		return []*CardWithOwned{}, nil
	}

	// First, search for cards
	var cards []*models.SetCard
	err := storage.RetryOnBusy(func() error {
		var err error
		cards, err = c.services.Storage.SetCardRepo().SearchCards(ctx, query, setCodes, limit)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to search cards: %v", err)}
	}

	if len(cards) == 0 {
		return []*CardWithOwned{}, nil
	}

	// Extract arena IDs to look up in collection
	cardIDs := make([]int, 0, len(cards))
	for _, card := range cards {
		// Parse ArenaID as int
		var arenaID int
		if _, err := fmt.Sscanf(card.ArenaID, "%d", &arenaID); err == nil && arenaID > 0 {
			cardIDs = append(cardIDs, arenaID)
		}
	}

	// Get collection quantities
	var collectionMap map[int]int
	err = storage.RetryOnBusy(func() error {
		var err error
		collectionMap, err = c.services.Storage.CollectionRepo().GetCards(ctx, cardIDs)
		return err
	})
	if err != nil {
		// Log but don't fail - collection data is optional
		log.Printf("Warning: Failed to get collection quantities: %v", err)
		collectionMap = make(map[int]int)
	}

	// Build results with ownership info
	results := make([]*CardWithOwned, 0, len(cards))
	for _, card := range cards {
		var arenaID int
		if _, err := fmt.Sscanf(card.ArenaID, "%d", &arenaID); err != nil {
			continue
		}

		owned := collectionMap[arenaID]

		// If collectionOnly, skip cards not in collection
		if collectionOnly && owned == 0 {
			continue
		}

		results = append(results, &CardWithOwned{
			SetCard:       card,
			OwnedQuantity: owned,
		})
	}

	return results, nil
}

// GetCollectionQuantities returns the collection quantities for a list of arena IDs.
func (c *CardFacade) GetCollectionQuantities(ctx context.Context, arenaIDs []int) (map[int]int, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if len(arenaIDs) == 0 {
		return make(map[int]int), nil
	}

	var collectionMap map[int]int
	err := storage.RetryOnBusy(func() error {
		var err error
		collectionMap, err = c.services.Storage.CollectionRepo().GetCards(ctx, arenaIDs)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get collection quantities: %v", err)}
	}

	return collectionMap, nil
}

// GetAllSetInfo returns information about all known sets.
// Falls back to set_cards table if sets table is empty.
func (c *CardFacade) GetAllSetInfo(ctx context.Context) ([]*SetInfo, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// First try the sets table (populated from Scryfall)
	sets, err := c.services.Storage.GetAllSets(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get all sets: %v", err)}
	}

	// If we have sets from the main table, use them
	if len(sets) > 0 {
		result := make([]*SetInfo, len(sets))
		for i, set := range sets {
			result[i] = &SetInfo{
				Code:       set.Code,
				Name:       set.Name,
				IconSVGURI: set.IconSVGURI,
				SetType:    set.SetType,
				ReleasedAt: set.ReleasedAt,
				CardCount:  set.CardCount,
			}
		}
		return result, nil
	}

	// Fallback: get unique set codes from set_cards table
	cachedSets, err := c.services.Storage.SetCardRepo().GetCachedSets(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get cached sets: %v", err)}
	}

	// Build set info from cached set codes (name defaults to uppercase code)
	result := make([]*SetInfo, len(cachedSets))
	for i, code := range cachedSets {
		result[i] = &SetInfo{
			Code: code,
			Name: strings.ToUpper(code), // Use uppercase code as name fallback
		}
	}

	return result, nil
}

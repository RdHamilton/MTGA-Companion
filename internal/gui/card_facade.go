package gui

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

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
	Code       string  `json:"code"`                 // Set code (e.g., "DSK", "BLB")
	Name       string  `json:"name"`                 // Full set name (e.g., "Duskmourn: House of Horror")
	IconSVGURI *string `json:"iconSvgUri,omitempty"` // URL to the set symbol SVG (may be null)
	SetType    *string `json:"setType,omitempty"`    // Type of set (may be null)
	ReleasedAt *string `json:"releasedAt,omitempty"` // Release date (may be null for unreleased sets)
	CardCount  *int    `json:"cardCount,omitempty"`  // Number of cards in set (may be null)
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
// If the card is not in the local database, it will attempt to fetch from Scryfall.
func (c *CardFacade) GetCardByArenaID(ctx context.Context, arenaID string) (*models.SetCard, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	log.Printf("[GetCardByArenaID] Looking up card with ArenaID: %s", arenaID)

	// First, try local database
	card, err := c.services.Storage.SetCardRepo().GetCardByArenaID(ctx, arenaID)
	if err != nil {
		log.Printf("[GetCardByArenaID] Error looking up card %s: %v", arenaID, err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card: %v", err)}
	}

	if card != nil {
		log.Printf("[GetCardByArenaID] Found card %s in database: Name=%s", arenaID, card.Name)
		return card, nil
	}

	// Card not in database - try fetching from Scryfall
	log.Printf("[GetCardByArenaID] Card %s not found in database, attempting Scryfall fetch...", arenaID)

	if c.services.SetFetcher != nil {
		// Convert arenaID string to int for FetchCardByArenaID
		var arenaIDInt int
		if _, err := fmt.Sscanf(arenaID, "%d", &arenaIDInt); err != nil {
			log.Printf("[GetCardByArenaID] Invalid arenaID format %s: %v", arenaID, err)
			return nil, nil
		}

		// Use FetchCardByArenaID which fetches from Scryfall and handles basic land fallbacks
		fetchedCard, fetchErr := c.services.SetFetcher.FetchCardByArenaID(ctx, arenaIDInt)
		if fetchErr != nil {
			log.Printf("[GetCardByArenaID] Failed to fetch card %s from Scryfall: %v", arenaID, fetchErr)
			// Return nil without error - card simply doesn't exist
			return nil, nil
		}
		if fetchedCard != nil {
			log.Printf("[GetCardByArenaID] Successfully fetched card %s from Scryfall: Name=%s", arenaID, fetchedCard.Name)
			return fetchedCard, nil
		}
	}

	log.Printf("[GetCardByArenaID] Card %s not found in database or Scryfall", arenaID)
	return nil, nil
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

// RatingsStaleness contains information about when ratings were last updated.
type RatingsStaleness struct {
	CachedAt  time.Time `json:"cachedAt"`
	IsStale   bool      `json:"isStale"`
	CardCount int       `json:"cardCount"`
}

// StaleThreshold is the duration after which ratings are considered stale (2 weeks).
const StaleThreshold = 14 * 24 * time.Hour

// GetRatingsStaleness returns staleness information for card ratings.
func (c *CardFacade) GetRatingsStaleness(ctx context.Context, setCode string, draftFormat string) (*RatingsStaleness, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	ratings, cachedAt, err := c.services.Storage.DraftRatingsRepo().GetCardRatings(ctx, setCode, draftFormat)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get ratings staleness: %v", err)}
	}

	// No ratings means no cached data
	if len(ratings) == 0 {
		return &RatingsStaleness{
			CachedAt:  time.Time{},
			IsStale:   true,
			CardCount: 0,
		}, nil
	}

	isStale := time.Since(cachedAt) > StaleThreshold

	return &RatingsStaleness{
		CachedAt:  cachedAt,
		IsStale:   isStale,
		CardCount: len(ratings),
	}, nil
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

	// Ensure specified sets are fetched/refreshed before searching
	// This handles Arena-exclusive sets like TLA that need special fetching
	if len(setCodes) > 0 && c.services.SetFetcher != nil {
		for _, setCode := range setCodes {
			// FetchAndCacheSet will check cache completeness and refresh if needed
			_, err := c.services.SetFetcher.FetchAndCacheSet(ctx, setCode)
			if err != nil {
				log.Printf("[SearchCards] Warning: Failed to ensure set %s is cached: %v", setCode, err)
				// Continue anyway - partial data is better than none
			}
		}
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
	log.Printf("[SearchCardsWithCollection] Called with query=%q, setCodes=%v, limit=%d, collectionOnly=%t", query, setCodes, limit, collectionOnly)

	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if query == "" {
		return []*CardWithOwned{}, nil
	}

	// Ensure specified sets are fetched/refreshed before searching
	// This handles Arena-exclusive sets like TLA that need special fetching
	log.Printf("[SearchCardsWithCollection] SetFetcher is nil: %t", c.services.SetFetcher == nil)
	if len(setCodes) > 0 && c.services.SetFetcher != nil {
		for _, setCode := range setCodes {
			// FetchAndCacheSet will check cache completeness and refresh if needed
			_, err := c.services.SetFetcher.FetchAndCacheSet(ctx, setCode)
			if err != nil {
				log.Printf("[SearchCardsWithCollection] Warning: Failed to ensure set %s is cached: %v", setCode, err)
				// Continue anyway - partial data is better than none
			}
		}
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
// Merges sets from: Scryfall sets table, cached set_cards, and 17Lands ratings.
func (c *CardFacade) GetAllSetInfo(ctx context.Context) ([]*SetInfo, error) {
	if c.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Collect all set codes from multiple sources
	setInfoMap := make(map[string]*SetInfo)

	// Source 1: Sets table (populated from Scryfall - has full metadata)
	sets, err := c.services.Storage.GetAllSets(ctx)
	if err != nil {
		log.Printf("[GetAllSetInfo] Warning: Failed to get sets from sets table: %v", err)
	} else {
		for _, set := range sets {
			setInfoMap[set.Code] = &SetInfo{
				Code:       set.Code,
				Name:       set.Name,
				IconSVGURI: set.IconSVGURI,
				SetType:    set.SetType,
				ReleasedAt: set.ReleasedAt,
				CardCount:  set.CardCount,
			}
		}
	}

	// Source 2: Cached set_cards (may have sets not in main table)
	cachedSets, err := c.services.Storage.SetCardRepo().GetCachedSets(ctx)
	if err != nil {
		log.Printf("[GetAllSetInfo] Warning: Failed to get cached sets: %v", err)
	} else {
		for _, code := range cachedSets {
			if _, exists := setInfoMap[code]; !exists {
				setInfoMap[code] = &SetInfo{
					Code: code,
					Name: strings.ToUpper(code),
				}
			}
		}
	}

	// Source 3: Sets with 17Lands ratings (for Arena-exclusive sets like TLA)
	ratingsRepo := c.services.Storage.DraftRatingsRepo()
	if ratingsRepo != nil {
		ratingSets, err := ratingsRepo.GetSetsWithRatings(ctx)
		if err != nil {
			log.Printf("[GetAllSetInfo] Warning: Failed to get sets with ratings: %v", err)
		} else {
			for _, code := range ratingSets {
				if _, exists := setInfoMap[code]; !exists {
					setInfoMap[code] = &SetInfo{
						Code: code,
						Name: strings.ToUpper(code),
					}
				}
			}
		}
	}

	// Convert map to slice
	result := make([]*SetInfo, 0, len(setInfoMap))
	for _, info := range setInfoMap {
		result = append(result, info)
	}

	// Sort by code for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result, nil
}

// GetCFBRatings returns all CFB ratings for a set.
func (c *CardFacade) GetCFBRatings(ctx context.Context, setCode string) ([]*models.CFBRating, error) {
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()
	return cfbRepo.GetRatingsForSet(ctx, strings.ToUpper(setCode))
}

// GetCFBRatingByCardName returns a CFB rating by card name and set code.
func (c *CardFacade) GetCFBRatingByCardName(ctx context.Context, cardName, setCode string) (*models.CFBRating, error) {
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()
	return cfbRepo.GetRating(ctx, cardName, strings.ToUpper(setCode))
}

// GetCFBRatingByArenaID returns a CFB rating by Arena ID.
func (c *CardFacade) GetCFBRatingByArenaID(ctx context.Context, arenaID int) (*models.CFBRating, error) {
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()
	return cfbRepo.GetRatingByArenaID(ctx, arenaID)
}

// ImportCFBRatings imports CFB ratings from structured data.
func (c *CardFacade) ImportCFBRatings(ctx context.Context, ratings interface{}) (int, error) {
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()

	// Type assert to get the slice of ratings
	cfbRatings, ok := ratings.([]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid ratings data type")
	}

	imported := 0
	for _, r := range cfbRatings {
		ratingMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		cardName, _ := ratingMap["card_name"].(string)
		setCode, _ := ratingMap["set_code"].(string)
		limitedRating, _ := ratingMap["limited_rating"].(string)

		if cardName == "" || setCode == "" || limitedRating == "" {
			continue
		}

		rating := &models.CFBRating{
			CardName:      cardName,
			SetCode:       strings.ToUpper(setCode),
			LimitedRating: limitedRating,
			LimitedScore:  models.LimitedGradeToScore(limitedRating),
		}

		// Optional fields
		if constructedRating, ok := ratingMap["constructed_rating"].(string); ok {
			rating.ConstructedRating = constructedRating
			rating.ConstructedScore = models.ConstructedRatingToScore(constructedRating)
		}
		if archetypeFit, ok := ratingMap["archetype_fit"].(string); ok {
			rating.ArchetypeFit = archetypeFit
		}
		if commentary, ok := ratingMap["commentary"].(string); ok {
			rating.Commentary = commentary
		}
		if sourceURL, ok := ratingMap["source_url"].(string); ok {
			rating.SourceURL = sourceURL
		}
		if author, ok := ratingMap["author"].(string); ok {
			rating.Author = author
		}

		if err := cfbRepo.UpsertRating(ctx, rating); err != nil {
			log.Printf("[CardFacade] Warning: failed to import CFB rating for %s: %v", cardName, err)
			continue
		}
		imported++
	}

	return imported, nil
}

// LinkCFBArenaIDs links CFB ratings to Arena IDs for a set.
func (c *CardFacade) LinkCFBArenaIDs(ctx context.Context, setCode string) (int, error) {
	setCode = strings.ToUpper(setCode)
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()
	setCardRepo := c.services.Storage.SetCardRepo()

	// Get all set cards for this set
	cards, err := setCardRepo.GetCardsBySet(ctx, setCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get set cards: %w", err)
	}

	if len(cards) == 0 {
		return 0, nil
	}

	// Build name to Arena ID map
	cardNameToArenaID := make(map[string]int)
	for _, card := range cards {
		normalizedName := strings.TrimSpace(strings.ToLower(card.Name))
		var arenaID int
		if _, err := fmt.Sscanf(card.ArenaID, "%d", &arenaID); err == nil {
			cardNameToArenaID[normalizedName] = arenaID
			cardNameToArenaID[strings.TrimSpace(card.Name)] = arenaID
		}
	}

	// Get CFB ratings for this set
	ratings, err := cfbRepo.GetRatingsForSet(ctx, setCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get CFB ratings: %w", err)
	}

	linked := 0
	for _, rating := range ratings {
		if rating.ArenaID != nil {
			continue
		}

		normalizedName := strings.TrimSpace(strings.ToLower(rating.CardName))
		if arenaID, found := cardNameToArenaID[normalizedName]; found {
			rating.ArenaID = &arenaID
			if err := cfbRepo.UpsertRating(ctx, rating); err != nil {
				continue
			}
			linked++
		} else if arenaID, found := cardNameToArenaID[rating.CardName]; found {
			rating.ArenaID = &arenaID
			if err := cfbRepo.UpsertRating(ctx, rating); err != nil {
				continue
			}
			linked++
		}
	}

	return linked, nil
}

// DeleteCFBRatings deletes all CFB ratings for a set.
func (c *CardFacade) DeleteCFBRatings(ctx context.Context, setCode string) error {
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()
	return cfbRepo.DeleteRatingsForSet(ctx, strings.ToUpper(setCode))
}

// GetCFBRatingsCount returns the count of CFB ratings for a set.
func (c *CardFacade) GetCFBRatingsCount(ctx context.Context, setCode string) (int, error) {
	cfbRepo := c.services.Storage.NewCFBRatingsRepo()
	return cfbRepo.GetRatingsCount(ctx, strings.ToUpper(setCode))
}

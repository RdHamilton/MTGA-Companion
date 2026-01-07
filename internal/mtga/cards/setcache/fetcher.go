package setcache

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// scryfallIDRegex matches the UUID in Scryfall image URLs
// Example URL: https://cards.scryfall.io/large/front/f/a/fa940e68-010e-4b68-be8a-555d7068f7b4.jpg
var scryfallIDRegex = regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

// ExtractScryfallIDFromURL extracts the Scryfall card ID (UUID) from a 17Lands image URL.
// 17Lands uses Scryfall image URLs which contain the card's UUID.
// Example: https://cards.scryfall.io/large/front/f/a/fa940e68-010e-4b68-be8a-555d7068f7b4.jpg
// Returns empty string if no UUID is found.
func ExtractScryfallIDFromURL(url string) string {
	if url == "" {
		return ""
	}
	match := scryfallIDRegex.FindStringSubmatch(url)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// MTGASetToScryfall maps MTGA set codes to Scryfall set codes.
var MTGASetToScryfall = map[string]string{
	"TLA": "tla", // Avatar: The Last Airbender
	"TDM": "tdm", // Tarkir Dragonstorm
	"DSK": "dsk", // Duskmourn: House of Horror
	"BLB": "blb", // Bloomburrow
	"OTJ": "otj", // Outlaws of Thunder Junction
	"MKM": "mkm", // Murders at Karlov Manor
	"LCI": "lci", // The Lost Caverns of Ixalan
	"WOE": "woe", // Wilds of Eldraine
	"LTR": "ltr", // The Lord of the Rings: Tales of Middle-earth
	"MOM": "mom", // March of the Machine
	"ONE": "one", // Phyrexia: All Will Be One
	"BRO": "bro", // The Brothers' War
	"DMU": "dmu", // Dominaria United
	"SNC": "snc", // Streets of New Capenna
	"NEO": "neo", // Kamigawa: Neon Dynasty
	"VOW": "vow", // Innistrad: Crimson Vow
	"MID": "mid", // Innistrad: Midnight Hunt
	"AFR": "afr", // Adventures in the Forgotten Realms
}

// ArenaExclusiveBasicLands maps Arena IDs to basic land names for sets that
// don't have Arena IDs in Scryfall (like TLA - Avatar: The Last Airbender).
// Format: arenaID -> {setCode, cardName}
var ArenaExclusiveBasicLands = map[int]struct {
	SetCode  string
	CardName string
}{
	// TLA (Avatar: The Last Airbender) basic lands
	97563: {"TLA", "Plains"},
	97564: {"TLA", "Island"},
	97565: {"TLA", "Swamp"},
	97566: {"TLA", "Mountain"},
	97567: {"TLA", "Forest"},
}

// Fetcher handles fetching and caching set cards from Scryfall.
type Fetcher struct {
	scryfallClient *scryfall.Client
	setCardRepo    repository.SetCardRepository
	ratingsRepo    repository.DraftRatingsRepository
}

// NewFetcher creates a new set card fetcher.
func NewFetcher(scryfallClient *scryfall.Client, setCardRepo repository.SetCardRepository, ratingsRepo repository.DraftRatingsRepository) *Fetcher {
	return &Fetcher{
		scryfallClient: scryfallClient,
		setCardRepo:    setCardRepo,
		ratingsRepo:    ratingsRepo,
	}
}

// FetchAndCacheSet fetches all cards for a set from Scryfall and caches them.
// Returns the number of cards cached.
func (f *Fetcher) FetchAndCacheSet(ctx context.Context, mtgaSetCode string) (int, error) {
	log.Printf("[FetchAndCacheSet] Starting fetch for %s", mtgaSetCode)

	// Map MTGA set code to Scryfall set code
	scryfallSetCode, ok := MTGASetToScryfall[mtgaSetCode]
	if !ok {
		scryfallSetCode = strings.ToLower(mtgaSetCode)
	}
	log.Printf("[FetchAndCacheSet] Mapped %s -> Scryfall code: %s", mtgaSetCode, scryfallSetCode)

	// Check if set is already cached
	isCached, err := f.setCardRepo.IsSetCached(ctx, mtgaSetCode)
	if err != nil {
		return 0, fmt.Errorf("check if set is cached: %w", err)
	}
	log.Printf("[FetchAndCacheSet] IsSetCached returned: %v", isCached)

	if isCached {
		// Check if cached count is complete by comparing to Scryfall's Arena card count
		needsRefresh, err := f.checkCacheCompleteness(ctx, mtgaSetCode, scryfallSetCode)
		if err != nil {
			log.Printf("[FetchAndCacheSet] Failed to check cache completeness: %v", err)
			// On error, assume cache is valid to avoid unnecessary API calls
		} else if needsRefresh {
			log.Printf("[FetchAndCacheSet] Cache is incomplete, refreshing %s", mtgaSetCode)
			return f.RefreshSet(ctx, mtgaSetCode)
		}

		log.Printf("[FetchAndCacheSet] Set %s is already cached and complete, skipping fetch", mtgaSetCode)
		return 0, nil // Already cached and complete
	}

	// Search for all cards in the set (with pagination)
	query := fmt.Sprintf("set:%s", scryfallSetCode)
	allCards := []*models.SetCard{}
	fetchedAt := time.Now()
	pageNum := 1

	log.Printf("[FetchAndCacheSet] Searching Scryfall with query: %s", query)
	searchResult, err := f.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		log.Printf("[FetchAndCacheSet] Scryfall search failed: %v", err)
		return 0, fmt.Errorf("search cards for set %s: %w", scryfallSetCode, err)
	}

	log.Printf("[FetchAndCacheSet] First page: found %d cards, hasMore=%v", len(searchResult.Data), searchResult.HasMore)

	// Process first page
	cardsWithArenaID := 0
	cardsWithoutArenaID := 0
	for _, scryfallCard := range searchResult.Data {
		// Skip cards without Arena IDs (not in MTGA)
		if scryfallCard.ArenaID == nil {
			cardsWithoutArenaID++
			continue
		}

		cardsWithArenaID++
		card := convertScryfallCard(&scryfallCard, mtgaSetCode, fetchedAt)
		allCards = append(allCards, card)
	}
	log.Printf("[FetchAndCacheSet] Page 1: %d cards with ArenaID, %d without", cardsWithArenaID, cardsWithoutArenaID)

	// Handle pagination if there are more results
	for searchResult.HasMore && searchResult.NextPage != "" {
		pageNum++

		// Fetch next page using the NextPage URL
		var nextResult scryfall.SearchResult
		if err := f.scryfallClient.DoRequestRaw(ctx, searchResult.NextPage, &nextResult); err != nil {
			return 0, fmt.Errorf("fetch page %d for set %s: %w", pageNum, scryfallSetCode, err)
		}

		// Process this page
		for _, scryfallCard := range nextResult.Data {
			// Skip cards without Arena IDs (not in MTGA)
			if scryfallCard.ArenaID == nil {
				continue
			}

			card := convertScryfallCard(&scryfallCard, mtgaSetCode, fetchedAt)
			allCards = append(allCards, card)
		}

		searchResult = &nextResult
	}

	// Save all cards to database
	log.Printf("[FetchAndCacheSet] Total cards to save: %d", len(allCards))
	if len(allCards) > 0 {
		if err := f.setCardRepo.SaveCards(ctx, allCards); err != nil {
			log.Printf("[FetchAndCacheSet] Failed to save cards: %v", err)
			return 0, fmt.Errorf("save cards to database: %w", err)
		}
		log.Printf("[FetchAndCacheSet] Successfully saved %d cards for %s", len(allCards), mtgaSetCode)
		return len(allCards), nil
	}

	// If no cards found with Arena IDs, this might be an Arena-exclusive set
	// Try using 17Lands ratings data which includes Arena IDs
	log.Printf("[FetchAndCacheSet] No cards found with Arena IDs - checking 17Lands for Arena-exclusive set")
	if f.ratingsRepo != nil {
		return f.fetchArenaExclusiveSet(ctx, mtgaSetCode, scryfallSetCode, fetchedAt)
	}

	log.Printf("[FetchAndCacheSet] WARNING: No cards found for %s and no ratings repository available", mtgaSetCode)
	return 0, nil
}

// convertScryfallCard converts a Scryfall card to a SetCard model.
func convertScryfallCard(scryfallCard *scryfall.Card, setCode string, fetchedAt time.Time) *models.SetCard {
	// Parse type line into types
	types := parseTypeLine(scryfallCard.TypeLine)

	// Get image URLs - check top-level first, then card_faces for DFCs
	imageURL := ""
	imageURLSmall := ""
	imageURLArt := ""
	if scryfallCard.ImageURIs != nil {
		imageURL = scryfallCard.ImageURIs.Normal
		imageURLSmall = scryfallCard.ImageURIs.Small
		imageURLArt = scryfallCard.ImageURIs.ArtCrop
	} else if len(scryfallCard.CardFaces) > 0 && scryfallCard.CardFaces[0].ImageURIs != nil {
		// For double-faced cards, use the front face image
		imageURL = scryfallCard.CardFaces[0].ImageURIs.Normal
		imageURLSmall = scryfallCard.CardFaces[0].ImageURIs.Small
		imageURLArt = scryfallCard.CardFaces[0].ImageURIs.ArtCrop
	}

	// Handle Arena ID (may be nil for cards not yet in MTGA)
	arenaID := ""
	if scryfallCard.ArenaID != nil {
		arenaID = fmt.Sprintf("%d", *scryfallCard.ArenaID)
	}

	return &models.SetCard{
		SetCode:       setCode,
		ArenaID:       arenaID,
		ScryfallID:    scryfallCard.ID,
		Name:          scryfallCard.Name,
		ManaCost:      scryfallCard.ManaCost,
		CMC:           int(scryfallCard.CMC),
		Types:         types,
		Colors:        scryfallCard.Colors,
		Rarity:        scryfallCard.Rarity,
		Text:          scryfallCard.OracleText,
		Power:         scryfallCard.Power,
		Toughness:     scryfallCard.Toughness,
		ImageURL:      imageURL,
		ImageURLSmall: imageURLSmall,
		ImageURLArt:   imageURLArt,
		FetchedAt:     fetchedAt,
	}
}

// GetCachedSet retrieves all cached cards for a set.
func (f *Fetcher) GetCachedSet(ctx context.Context, setCode string) ([]*models.SetCard, error) {
	return f.setCardRepo.GetCardsBySet(ctx, setCode)
}

// GetCardByArenaID retrieves a cached card by its Arena ID.
func (f *Fetcher) GetCardByArenaID(ctx context.Context, arenaID string) (*models.SetCard, error) {
	return f.setCardRepo.GetCardByArenaID(ctx, arenaID)
}

// RefreshSet deletes and re-fetches all cards for a set.
func (f *Fetcher) RefreshSet(ctx context.Context, setCode string) (int, error) {
	log.Printf("[RefreshSet] Deleting existing cache for %s", setCode)
	// Delete existing cache
	if err := f.setCardRepo.DeleteSet(ctx, setCode); err != nil {
		log.Printf("[RefreshSet] Failed to delete cache: %v", err)
		return 0, fmt.Errorf("delete existing cache: %w", err)
	}
	log.Printf("[RefreshSet] Successfully deleted cache for %s", setCode)

	// Fetch and cache again
	log.Printf("[RefreshSet] Fetching fresh data for %s", setCode)
	return f.FetchAndCacheSet(ctx, setCode)
}

// FetchCardByName fetches a single card from Scryfall by exact name and set code.
// Returns nil if the card is not found.
// Checks cache first to avoid unnecessary API calls.
func (f *Fetcher) FetchCardByName(ctx context.Context, setCode, cardName, arenaID string) (*models.SetCard, error) {
	// Check if card is already cached by Arena ID
	cachedCard, err := f.setCardRepo.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		return nil, fmt.Errorf("check cache: %w", err)
	}
	if cachedCard != nil {
		return cachedCard, nil // Already cached
	}

	// Map MTGA set code to Scryfall set code
	scryfallSetCode, ok := MTGASetToScryfall[setCode]
	if !ok {
		scryfallSetCode = strings.ToLower(setCode)
	}

	// Search Scryfall for this specific card (!"name" means exact match)
	query := fmt.Sprintf(`!"%s" set:%s`, cardName, scryfallSetCode)
	log.Printf("[FetchCardByName] Searching Scryfall with query: %s", query)
	searchResult, err := f.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		log.Printf("[FetchCardByName] Scryfall API error for '%s': %v", cardName, err)
		return nil, fmt.Errorf("scryfall search failed: %w", err)
	}
	if len(searchResult.Data) == 0 {
		log.Printf("[FetchCardByName] No results from Scryfall for '%s' (query: %s)", cardName, query)
		return nil, nil // Card not found
	}
	log.Printf("[FetchCardByName] Found %d result(s) for '%s'", len(searchResult.Data), cardName)

	// Take the first result
	scryfallCard := searchResult.Data[0]

	// Convert and use our Arena ID
	card := convertScryfallCard(&scryfallCard, setCode, time.Now())
	card.ArenaID = arenaID

	// Save to database
	if err := f.setCardRepo.SaveCard(ctx, card); err != nil {
		return nil, fmt.Errorf("save card: %w", err)
	}

	return card, nil
}

// FetchCardByArenaID fetches a single card from Scryfall by Arena ID and caches it.
// Returns the cached card if it already exists, otherwise fetches from Scryfall.
// For Arena-exclusive sets (like TLA), falls back to basic land mapping if needed.
func (f *Fetcher) FetchCardByArenaID(ctx context.Context, arenaID int) (*models.SetCard, error) {
	arenaIDStr := fmt.Sprintf("%d", arenaID)

	// Check if card is already cached
	cachedCard, err := f.setCardRepo.GetCardByArenaID(ctx, arenaIDStr)
	if err != nil {
		return nil, fmt.Errorf("check cache: %w", err)
	}
	if cachedCard != nil {
		return cachedCard, nil // Already cached
	}

	// Fetch from Scryfall
	log.Printf("[FetchCardByArenaID] Fetching card %d from Scryfall", arenaID)
	scryfallCard, err := f.scryfallClient.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		log.Printf("[FetchCardByArenaID] Scryfall API error for ArenaID %d: %v", arenaID, err)

		// Check if this is a known basic land from an Arena-exclusive set
		if basicLand, ok := ArenaExclusiveBasicLands[arenaID]; ok {
			log.Printf("[FetchCardByArenaID] Found basic land mapping for %d: %s (%s)", arenaID, basicLand.CardName, basicLand.SetCode)
			return f.fetchBasicLandByName(ctx, arenaID, basicLand.SetCode, basicLand.CardName)
		}

		return nil, fmt.Errorf("scryfall fetch failed: %w", err)
	}

	// Determine set code - use uppercase MTGA convention
	setCode := strings.ToUpper(scryfallCard.SetCode)

	// Convert Scryfall card to SetCard
	card := convertScryfallCard(scryfallCard, setCode, time.Now())

	// Save to database
	if err := f.setCardRepo.SaveCard(ctx, card); err != nil {
		log.Printf("[FetchCardByArenaID] Failed to save card %d: %v", arenaID, err)
		return nil, fmt.Errorf("save card: %w", err)
	}

	log.Printf("[FetchCardByArenaID] Cached card %d: %s (%s)", arenaID, card.Name, card.SetCode)
	return card, nil
}

// fetchBasicLandByName fetches a basic land from Scryfall by name and set code.
// Used as fallback for Arena-exclusive sets that don't have Arena IDs in Scryfall.
func (f *Fetcher) fetchBasicLandByName(ctx context.Context, arenaID int, setCode, cardName string) (*models.SetCard, error) {
	// Map to Scryfall set code
	scryfallSetCode, ok := MTGASetToScryfall[setCode]
	if !ok {
		scryfallSetCode = strings.ToLower(setCode)
	}

	// Search for the basic land by name and set
	query := fmt.Sprintf(`!"%s" set:%s type:basic`, cardName, scryfallSetCode)
	log.Printf("[fetchBasicLandByName] Searching Scryfall: %s", query)

	searchResult, err := f.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		log.Printf("[fetchBasicLandByName] Scryfall search failed: %v", err)
		return nil, fmt.Errorf("scryfall search failed: %w", err)
	}

	if len(searchResult.Data) == 0 {
		log.Printf("[fetchBasicLandByName] No results for %s in %s", cardName, setCode)
		return nil, fmt.Errorf("basic land %s not found in set %s", cardName, setCode)
	}

	// Take the first result and convert
	scryfallCard := &searchResult.Data[0]
	card := convertScryfallCard(scryfallCard, setCode, time.Now())

	// Override Arena ID with our known mapping
	card.ArenaID = fmt.Sprintf("%d", arenaID)

	// Save to database
	if err := f.setCardRepo.SaveCard(ctx, card); err != nil {
		log.Printf("[fetchBasicLandByName] Failed to save card: %v", err)
		return nil, fmt.Errorf("save card: %w", err)
	}

	log.Printf("[fetchBasicLandByName] Cached basic land %d: %s (%s)", arenaID, card.Name, card.SetCode)
	return card, nil
}

// parseTypeLine parses a type line into individual types.
// Example: "Creature — Elf Warrior" -> ["Creature", "Elf", "Warrior"]
func parseTypeLine(typeLine string) []string {
	// Split by " — " (em dash) to separate card types from subtypes
	parts := strings.Split(typeLine, " — ")

	types := []string{}

	// First part contains main types (e.g., "Legendary Creature")
	if len(parts) > 0 {
		mainTypes := strings.Fields(parts[0])
		types = append(types, mainTypes...)
	}

	// Second part contains subtypes (e.g., "Elf Warrior")
	if len(parts) > 1 {
		subtypes := strings.Fields(parts[1])
		types = append(types, subtypes...)
	}

	return types
}

// fetchArenaExclusiveSet handles fetching cards for Arena-exclusive sets (like TLA)
// that don't have Arena IDs in Scryfall. Uses 17Lands ratings data for Arena IDs
// and fetches card details from Scryfall using batch API.
//
// Strategy:
// 1. Extract Scryfall IDs from 17Lands image URLs (most reliable)
// 2. Fall back to name-based lookup for cards without URLs
func (f *Fetcher) fetchArenaExclusiveSet(ctx context.Context, mtgaSetCode, scryfallSetCode string, fetchedAt time.Time) (int, error) {
	log.Printf("[fetchArenaExclusiveSet] Attempting to fetch Arena-exclusive set %s using 17Lands data", mtgaSetCode)

	// Try both draft formats and use the one with more ratings
	premierRatings, _, err := f.ratingsRepo.GetCardRatings(ctx, mtgaSetCode, "PremierDraft")
	if err != nil {
		log.Printf("[fetchArenaExclusiveSet] Failed to get PremierDraft ratings: %v", err)
		premierRatings = nil
	}

	quickRatings, _, err := f.ratingsRepo.GetCardRatings(ctx, mtgaSetCode, "QuickDraft")
	if err != nil {
		log.Printf("[fetchArenaExclusiveSet] Failed to get QuickDraft ratings: %v", err)
		quickRatings = nil
	}

	// Use whichever format has more ratings
	var ratings []seventeenlands.CardRating
	var formatUsed string
	if len(premierRatings) >= len(quickRatings) && len(premierRatings) > 0 {
		ratings = premierRatings
		formatUsed = "PremierDraft"
	} else if len(quickRatings) > 0 {
		ratings = quickRatings
		formatUsed = "QuickDraft"
	}

	if len(ratings) == 0 {
		log.Printf("[fetchArenaExclusiveSet] No 17Lands ratings found for %s in any format", mtgaSetCode)
		return 0, nil
	}

	log.Printf("[fetchArenaExclusiveSet] Using %s with %d ratings (PremierDraft: %d, QuickDraft: %d)",
		formatUsed, len(ratings), len(premierRatings), len(quickRatings))

	// Separate ratings into those with Scryfall IDs (from URLs) and those without
	scryfallIDToArenaID := make(map[string]string)
	nameToArenaID := make(map[string]string)
	scryfallIDs := make([]string, 0)
	cardNamesWithoutURL := make([]string, 0)

	for _, rating := range ratings {
		arenaID := fmt.Sprintf("%d", rating.MTGAID)

		// Try to extract Scryfall ID from URL (most reliable method)
		scryfallID := ExtractScryfallIDFromURL(rating.URL)
		if scryfallID != "" {
			scryfallIDToArenaID[scryfallID] = arenaID
			scryfallIDs = append(scryfallIDs, scryfallID)
		} else {
			// Fall back to name-based lookup
			nameToArenaID[rating.Name] = arenaID
			cardNamesWithoutURL = append(cardNamesWithoutURL, rating.Name)
		}
	}

	log.Printf("[fetchArenaExclusiveSet] %d cards have Scryfall IDs from URLs, %d need name-based lookup",
		len(scryfallIDs), len(cardNamesWithoutURL))

	allCards := make([]*models.SetCard, 0, len(ratings))

	// Fetch cards by Scryfall ID (most reliable)
	if len(scryfallIDs) > 0 {
		log.Printf("[fetchArenaExclusiveSet] Fetching %d cards by Scryfall ID", len(scryfallIDs))
		scryfallCards, notFoundIDs, err := f.scryfallClient.GetCardsByScryfallIDs(ctx, scryfallIDs)
		if err != nil {
			log.Printf("[fetchArenaExclusiveSet] Batch fetch by Scryfall ID failed: %v", err)
			return 0, fmt.Errorf("batch fetch cards by Scryfall ID: %w", err)
		}

		log.Printf("[fetchArenaExclusiveSet] Scryfall ID fetch returned %d cards, %d not found",
			len(scryfallCards), len(notFoundIDs))

		for i := range scryfallCards {
			scryfallCard := &scryfallCards[i]
			card := convertScryfallCard(scryfallCard, mtgaSetCode, fetchedAt)

			// Assign Arena ID from 17Lands data using the Scryfall ID
			if arenaID, ok := scryfallIDToArenaID[scryfallCard.ID]; ok {
				card.ArenaID = arenaID
			}

			allCards = append(allCards, card)
		}

		if len(notFoundIDs) > 0 {
			log.Printf("[fetchArenaExclusiveSet] Scryfall IDs not found: %v", notFoundIDs)
		}
	}

	// Fetch remaining cards by name (fallback for cards without URLs)
	if len(cardNamesWithoutURL) > 0 {
		log.Printf("[fetchArenaExclusiveSet] Fetching %d cards by name (fallback)", len(cardNamesWithoutURL))
		scryfallCards, notFoundNames, err := f.scryfallClient.GetCardsByNames(ctx, cardNamesWithoutURL)
		if err != nil {
			log.Printf("[fetchArenaExclusiveSet] Batch fetch by name failed: %v", err)
			// Don't fail completely - we may have already fetched some cards by Scryfall ID
		} else {
			log.Printf("[fetchArenaExclusiveSet] Name fetch returned %d cards, %d not found",
				len(scryfallCards), len(notFoundNames))

			for i := range scryfallCards {
				scryfallCard := &scryfallCards[i]
				card := convertScryfallCard(scryfallCard, mtgaSetCode, fetchedAt)

				// Assign Arena ID from 17Lands data using the name
				if arenaID, ok := nameToArenaID[scryfallCard.Name]; ok {
					card.ArenaID = arenaID
				}

				allCards = append(allCards, card)
			}

			if len(notFoundNames) > 0 {
				log.Printf("[fetchArenaExclusiveSet] Cards not found by name: %v", notFoundNames)
			}
		}
	}

	log.Printf("[fetchArenaExclusiveSet] Total cards converted: %d", len(allCards))

	// Save all cards to database
	if len(allCards) > 0 {
		if err := f.setCardRepo.SaveCards(ctx, allCards); err != nil {
			log.Printf("[fetchArenaExclusiveSet] Failed to save cards: %v", err)
			return 0, fmt.Errorf("save cards to database: %w", err)
		}
		log.Printf("[fetchArenaExclusiveSet] Successfully saved %d cards for %s", len(allCards), mtgaSetCode)
	}

	return len(allCards), nil
}

// checkCacheCompleteness compares cached card count against expected card count.
// For Arena-exclusive sets (like TLA), uses 17Lands ratings count since Scryfall lacks Arena IDs.
// Returns true if cache needs refresh (cached count is significantly less than expected).
func (f *Fetcher) checkCacheCompleteness(ctx context.Context, mtgaSetCode, scryfallSetCode string) (bool, error) {
	// Get cached card count
	cachedCards, err := f.setCardRepo.GetCardsBySet(ctx, mtgaSetCode)
	if err != nil {
		return false, fmt.Errorf("get cached cards: %w", err)
	}
	cachedCount := len(cachedCards)

	// Query Scryfall for expected Arena card count (cards with arena_id)
	// Use game:arena filter to only count cards available in MTG Arena
	query := fmt.Sprintf("set:%s game:arena", scryfallSetCode)
	searchResult, err := f.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		// If Scryfall query fails, assume cache is valid
		return false, fmt.Errorf("scryfall search: %w", err)
	}

	expectedCount := searchResult.TotalCards
	log.Printf("[checkCacheCompleteness] Set %s: cached=%d, scryfall_expected=%d", mtgaSetCode, cachedCount, expectedCount)

	// If Scryfall has Arena cards, use that count
	if expectedCount > 0 {
		if cachedCount < (expectedCount * 9 / 10) {
			log.Printf("[checkCacheCompleteness] Cache is incomplete: %d < %d (90%% of %d)", cachedCount, expectedCount*9/10, expectedCount)
			return true, nil
		}
		return false, nil
	}

	// Arena-exclusive set: Scryfall has 0 game:arena cards
	// Use 17Lands ratings count as the expected count instead
	if f.ratingsRepo != nil {
		ratings, _, err := f.ratingsRepo.GetCardRatings(ctx, mtgaSetCode, "PremierDraft")
		if err != nil {
			log.Printf("[checkCacheCompleteness] Failed to get 17Lands ratings for %s: %v", mtgaSetCode, err)
			// Try QuickDraft as fallback
			ratings, _, err = f.ratingsRepo.GetCardRatings(ctx, mtgaSetCode, "QuickDraft")
			if err != nil {
				log.Printf("[checkCacheCompleteness] Failed to get QuickDraft ratings too: %v", err)
				return false, nil
			}
		}

		ratingsCount := len(ratings)
		log.Printf("[checkCacheCompleteness] Arena-exclusive set %s: cached=%d, 17lands_expected=%d", mtgaSetCode, cachedCount, ratingsCount)

		// If 17Lands has more cards than cache, refresh is needed
		if ratingsCount > 0 && cachedCount < (ratingsCount*9/10) {
			log.Printf("[checkCacheCompleteness] Arena-exclusive cache is incomplete: %d < %d (90%% of %d)", cachedCount, ratingsCount*9/10, ratingsCount)
			return true, nil
		}
	}

	return false, nil
}

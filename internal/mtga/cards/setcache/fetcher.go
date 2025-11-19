package setcache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

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

// Fetcher handles fetching and caching set cards from Scryfall.
type Fetcher struct {
	scryfallClient *scryfall.Client
	setCardRepo    repository.SetCardRepository
}

// NewFetcher creates a new set card fetcher.
func NewFetcher(scryfallClient *scryfall.Client, setCardRepo repository.SetCardRepository) *Fetcher {
	return &Fetcher{
		scryfallClient: scryfallClient,
		setCardRepo:    setCardRepo,
	}
}

// FetchAndCacheSet fetches all cards for a set from Scryfall and caches them.
// Returns the number of cards cached.
func (f *Fetcher) FetchAndCacheSet(ctx context.Context, mtgaSetCode string) (int, error) {
	// Map MTGA set code to Scryfall set code
	scryfallSetCode, ok := MTGASetToScryfall[mtgaSetCode]
	if !ok {
		scryfallSetCode = strings.ToLower(mtgaSetCode)
	}

	// Check if set is already cached
	isCached, err := f.setCardRepo.IsSetCached(ctx, mtgaSetCode)
	if err != nil {
		return 0, fmt.Errorf("check if set is cached: %w", err)
	}
	if isCached {
		return 0, nil // Already cached
	}

	// Search for all cards in the set (with pagination)
	query := fmt.Sprintf("set:%s", scryfallSetCode)
	allCards := []*models.SetCard{}
	fetchedAt := time.Now()
	pageNum := 1

	searchResult, err := f.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("search cards for set %s: %w", scryfallSetCode, err)
	}

	// Process first page
	for _, scryfallCard := range searchResult.Data {
		// Skip cards without Arena IDs (not in MTGA)
		if scryfallCard.ArenaID == nil {
			continue
		}

		card := convertScryfallCard(&scryfallCard, mtgaSetCode, fetchedAt)
		allCards = append(allCards, card)
	}

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
	if len(allCards) > 0 {
		if err := f.setCardRepo.SaveCards(ctx, allCards); err != nil {
			return 0, fmt.Errorf("save cards to database: %w", err)
		}
	}

	return len(allCards), nil
}

// convertScryfallCard converts a Scryfall card to a SetCard model.
func convertScryfallCard(scryfallCard *scryfall.Card, setCode string, fetchedAt time.Time) *models.SetCard {
	// Parse type line into types
	types := parseTypeLine(scryfallCard.TypeLine)

	// Get image URLs
	imageURL := ""
	imageURLSmall := ""
	imageURLArt := ""
	if scryfallCard.ImageURIs != nil {
		imageURL = scryfallCard.ImageURIs.Normal
		imageURLSmall = scryfallCard.ImageURIs.Small
		imageURLArt = scryfallCard.ImageURIs.ArtCrop
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
	// Delete existing cache
	if err := f.setCardRepo.DeleteSet(ctx, setCode); err != nil {
		return 0, fmt.Errorf("delete existing cache: %w", err)
	}

	// Fetch and cache again
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

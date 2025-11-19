package setcache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// MTGASetToScryfall maps MTGA set codes to Scryfall set codes.
var MTGASetToScryfall = map[string]string{
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

	// Search for all cards in the set
	query := fmt.Sprintf("set:%s", scryfallSetCode)
	searchResult, err := f.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("search cards for set %s: %w", scryfallSetCode, err)
	}

	// Convert Scryfall cards to SetCard models
	cards := make([]*models.SetCard, 0, len(searchResult.Data))
	fetchedAt := time.Now()

	for _, scryfallCard := range searchResult.Data {
		// Skip cards without Arena IDs (not in MTGA)
		if scryfallCard.ArenaID == nil {
			continue
		}

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

		card := &models.SetCard{
			SetCode:       mtgaSetCode,
			ArenaID:       fmt.Sprintf("%d", *scryfallCard.ArenaID),
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

		cards = append(cards, card)
	}

	// Save cards to database
	if len(cards) > 0 {
		if err := f.setCardRepo.SaveCards(ctx, cards); err != nil {
			return 0, fmt.Errorf("save cards to database: %w", err)
		}
	}

	// TODO: Handle pagination if there are more results
	// For now, we just cache the first page

	return len(cards), nil
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

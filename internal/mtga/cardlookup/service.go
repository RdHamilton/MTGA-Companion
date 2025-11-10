package cardlookup

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Service provides unified card lookup with caching.
// It integrates the storage layer and Scryfall API client.
type Service struct {
	storage        *storage.Service
	scryfallClient *scryfall.Client
	staleThreshold time.Duration
}

// ServiceOptions configures the card lookup service.
type ServiceOptions struct {
	// StaleThreshold is how old cached data can be before fetching from Scryfall.
	// Default: 7 days
	StaleThreshold time.Duration
}

// DefaultServiceOptions returns sensible defaults.
func DefaultServiceOptions() ServiceOptions {
	return ServiceOptions{
		StaleThreshold: 7 * 24 * time.Hour, // 7 days
	}
}

// NewService creates a new card lookup service.
func NewService(storage *storage.Service, scryfallClient *scryfall.Client, options ServiceOptions) *Service {
	if options.StaleThreshold == 0 {
		options.StaleThreshold = DefaultServiceOptions().StaleThreshold
	}

	return &Service{
		storage:        storage,
		scryfallClient: scryfallClient,
		staleThreshold: options.StaleThreshold,
	}
}

// GetCard retrieves a card by Arena ID.
// It checks the cache first, then falls back to Scryfall API if needed.
func (s *Service) GetCard(arenaID int) (*storage.Card, error) {
	ctx := context.Background()

	// Check cache first
	card, err := s.storage.GetCardByArenaID(ctx, arenaID)
	if err == nil && card != nil {
		// Check if card is stale
		if time.Since(card.LastUpdated) < s.staleThreshold {
			return card, nil
		}
	}

	// Cache miss or stale - fetch from Scryfall
	scryfallCard, err := s.scryfallClient.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		// If we have stale cache, return it as fallback
		if card != nil {
			return card, nil
		}
		return nil, fmt.Errorf("failed to fetch card from Scryfall: %w", err)
	}

	// Convert and save to cache
	storageCard := convertToStorageCard(scryfallCard)
	// Attempt to save to cache - ignore errors since we have the data from Scryfall
	_ = s.storage.SaveCard(ctx, storageCard)

	return storageCard, nil
}

// GetCards retrieves multiple cards by their Arena IDs.
// It batch-fetches from cache and only queries Scryfall for missing cards.
func (s *Service) GetCards(arenaIDs []int) ([]*storage.Card, error) {
	cards := make([]*storage.Card, 0, len(arenaIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Fetch each card concurrently
	for _, id := range arenaIDs {
		wg.Add(1)
		go func(arenaID int) {
			defer wg.Done()

			card, err := s.GetCard(arenaID)
			if err == nil && card != nil {
				mu.Lock()
				cards = append(cards, card)
				mu.Unlock()
			}
		}(id)
	}

	wg.Wait()
	return cards, nil
}

// SearchByName searches for cards by name.
func (s *Service) SearchByName(name string) ([]*storage.Card, error) {
	ctx := context.Background()
	return s.storage.SearchCards(ctx, name)
}

// GetCardsBySet retrieves all cards for a given set.
func (s *Service) GetCardsBySet(setCode string) ([]*storage.Card, error) {
	ctx := context.Background()
	return s.storage.GetCardsBySet(ctx, setCode)
}

// convertToStorageCard converts a Scryfall card to a storage card.
func convertToStorageCard(card *scryfall.Card) *storage.Card {
	// Convert Legalities struct to map
	legalities := map[string]string{
		"standard":        card.Legalities.Standard,
		"future":          card.Legalities.Future,
		"historic":        card.Legalities.Historic,
		"gladiator":       card.Legalities.Gladiator,
		"pioneer":         card.Legalities.Pioneer,
		"explorer":        card.Legalities.Explorer,
		"modern":          card.Legalities.Modern,
		"legacy":          card.Legalities.Legacy,
		"pauper":          card.Legalities.Pauper,
		"vintage":         card.Legalities.Vintage,
		"penny":           card.Legalities.Penny,
		"commander":       card.Legalities.Commander,
		"oathbreaker":     card.Legalities.Oathbreaker,
		"brawl":           card.Legalities.Brawl,
		"historicbrawl":   card.Legalities.HistoricBrawl,
		"alchemy":         card.Legalities.Alchemy,
		"paupercommander": card.Legalities.PauperCommander,
		"duel":            card.Legalities.Duel,
		"oldschool":       card.Legalities.OldSchool,
	}

	return &storage.Card{
		ID:              card.ID,
		ArenaID:         card.ArenaID,
		Name:            card.Name,
		ManaCost:        card.ManaCost,
		CMC:             card.CMC,
		TypeLine:        card.TypeLine,
		OracleText:      card.OracleText,
		Colors:          card.Colors,
		ColorIdentity:   card.ColorIdentity,
		Rarity:          card.Rarity,
		SetCode:         card.SetCode,
		CollectorNumber: card.CollectorNumber,
		Power:           card.Power,
		Toughness:       card.Toughness,
		Loyalty:         card.Loyalty,
		ImageURIs:       card.ImageURIs,
		Layout:          card.Layout,
		CardFaces:       card.CardFaces,
		Legalities:      legalities,
		ReleasedAt:      card.ReleasedAt,
		LastUpdated:     time.Now(),
	}
}

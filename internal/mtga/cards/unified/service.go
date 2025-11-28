package unified

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CardMetadataProvider provides card metadata (Scryfall data).
type CardMetadataProvider interface {
	GetCard(ctx context.Context, arenaID int) (*models.SetCard, error)
	GetCards(ctx context.Context, arenaIDs []int) ([]*models.SetCard, error)
	GetSetCards(ctx context.Context, setCode string) ([]*models.SetCard, error)
}

// DraftStatsProvider provides draft statistics (17Lands data).
type DraftStatsProvider interface {
	GetCardRating(ctx context.Context, arenaID int, expansion, format, colors string) (*storage.DraftCardRating, error)
	GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]*storage.DraftCardRating, error)
}

// Service composes card data from multiple sources.
type Service struct {
	metadata   CardMetadataProvider
	draftstats DraftStatsProvider
}

// NewService creates a new unified card service.
func NewService(metadata CardMetadataProvider, draftstats DraftStatsProvider) *Service {
	return &Service{
		metadata:   metadata,
		draftstats: draftstats,
	}
}

// GetCard retrieves a complete card with metadata and draft statistics.
// Returns card with metadata only if draft stats are unavailable.
func (s *Service) GetCard(ctx context.Context, arenaID int, setCode, format string) (*UnifiedCard, error) {
	// 1. Get card metadata (always required)
	card, err := s.metadata.GetCard(ctx, arenaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card metadata: %w", err)
	}

	// 2. Compose unified card
	unified := s.composeCard(card, nil)

	// 3. Try to add draft stats (best effort)
	if setCode == "" {
		setCode = card.SetCode
	}
	stats, err := s.draftstats.GetCardRating(ctx, arenaID, setCode, format, "")
	if err == nil && stats != nil {
		unified.DraftStats = convertDraftStats(stats)
		unified.StatsAge = time.Since(stats.LastUpdated)
		unified.StatsSource = SourceCache
	}

	return unified, nil
}

// GetCards retrieves multiple cards efficiently with batch operations.
// Returns cards with available data (missing stats is OK).
func (s *Service) GetCards(ctx context.Context, arenaIDs []int, format string) ([]*UnifiedCard, error) {
	if len(arenaIDs) == 0 {
		return []*UnifiedCard{}, nil
	}

	// 1. Batch fetch metadata
	cards, err := s.metadata.GetCards(ctx, arenaIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards metadata: %w", err)
	}

	// 2. Group cards by set for efficient stats fetching
	cardsBySet := make(map[string][]*models.SetCard)
	for _, card := range cards {
		cardsBySet[card.SetCode] = append(cardsBySet[card.SetCode], card)
	}

	// 3. Batch fetch stats per set
	statsMap := make(map[int]*storage.DraftCardRating)
	for setCode := range cardsBySet {
		ratings, err := s.draftstats.GetCardRatingsForSet(ctx, setCode, format, "")
		if err != nil {
			// Skip stats for this set if unavailable
			continue
		}
		for _, rating := range ratings {
			statsMap[rating.ArenaID] = rating
		}
	}

	// 4. Compose unified cards
	unified := make([]*UnifiedCard, len(cards))
	for i, card := range cards {
		var stats *storage.DraftCardRating
		arenaID := parseArenaID(card.ArenaID)
		if arenaID > 0 {
			stats = statsMap[arenaID]
		}
		unified[i] = s.composeCard(card, stats)
	}

	return unified, nil
}

// GetSetCards retrieves all cards for a set with draft statistics.
func (s *Service) GetSetCards(ctx context.Context, setCode, format string) ([]*UnifiedCard, error) {
	// 1. Get all cards in set
	cards, err := s.metadata.GetSetCards(ctx, setCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get set cards: %w", err)
	}

	// 2. Get draft stats for the set
	ratings, err := s.draftstats.GetCardRatingsForSet(ctx, setCode, format, "")
	statsMap := make(map[int]*storage.DraftCardRating)
	if err == nil {
		for _, rating := range ratings {
			statsMap[rating.ArenaID] = rating
		}
	}

	// 3. Compose unified cards
	unified := make([]*UnifiedCard, len(cards))
	for i, card := range cards {
		var stats *storage.DraftCardRating
		arenaID := parseArenaID(card.ArenaID)
		if arenaID > 0 {
			stats = statsMap[arenaID]
		}
		unified[i] = s.composeCard(card, stats)
	}

	return unified, nil
}

// parseArenaID converts string ArenaID to int
func parseArenaID(arenaID string) int {
	var id int
	_, _ = fmt.Sscanf(arenaID, "%d", &id)
	return id
}

// composeCard combines metadata and optional draft stats into a unified card.
func (s *Service) composeCard(card *models.SetCard, stats *storage.DraftCardRating) *UnifiedCard {
	unified := &UnifiedCard{
		ID:             fmt.Sprintf("%d", card.ID),
		ArenaID:        parseArenaID(card.ArenaID),
		Name:           card.Name,
		ManaCost:       card.ManaCost,
		CMC:            float64(card.CMC),
		TypeLine:       strings.Join(card.Types, " "),
		Colors:         card.Colors,
		ColorIdentity:  card.Colors, // SetCard doesn't have separate color identity
		Rarity:         card.Rarity,
		SetCode:        card.SetCode,
		Power:          card.Power,
		Toughness:      card.Toughness,
		MetadataSource: SourceCache,
	}

	// Add draft stats if available
	if stats != nil {
		unified.DraftStats = convertDraftStats(stats)
		unified.StatsAge = time.Since(stats.LastUpdated)
		unified.StatsSource = SourceCache
	}

	return unified
}

// convertDraftStats converts storage draft statistics to unified model.
func convertDraftStats(stats *storage.DraftCardRating) *DraftStatistics {
	return &DraftStatistics{
		GIHWR:       stats.GIHWR,
		OHWR:        stats.OHWR,
		GPWR:        stats.GPWR,
		GDWR:        stats.GDWR,
		IHDWR:       stats.IHDWR,
		GIHWRDelta:  stats.GIHWRDelta,
		OHWRDelta:   stats.OHWRDelta,
		GDWRDelta:   stats.GDWRDelta,
		IHDWRDelta:  stats.IHDWRDelta,
		ALSA:        stats.ALSA,
		ATA:         stats.ATA,
		GIH:         stats.GIH,
		OH:          stats.OH,
		GP:          stats.GP,
		GD:          stats.GD,
		IHD:         stats.IHD,
		GamesPlayed: stats.GamesPlayed,
		NumDecks:    stats.NumDecks,
		Format:      stats.Format,
		Colors:      stats.Colors,
		StartDate:   stats.StartDate,
		EndDate:     stats.EndDate,
		LastUpdated: stats.LastUpdated,
	}
}

// FilterByRarity filters unified cards by rarity.
func FilterByRarity(cards []*UnifiedCard, rarity string) []*UnifiedCard {
	var filtered []*UnifiedCard
	for _, card := range cards {
		if card.Rarity == rarity {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// FilterByColors filters unified cards by color identity.
func FilterByColors(cards []*UnifiedCard, colors []string) []*UnifiedCard {
	if len(colors) == 0 {
		return cards
	}

	var filtered []*UnifiedCard
	for _, card := range cards {
		if containsAllColors(card.ColorIdentity, colors) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// FilterByStats filters cards that have draft statistics.
func FilterByStats(cards []*UnifiedCard) []*UnifiedCard {
	var filtered []*UnifiedCard
	for _, card := range cards {
		if card.HasDraftStats() {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// SortByWinRate sorts cards by GIHWR (descending).
func SortByWinRate(cards []*UnifiedCard) {
	// Simple bubble sort (can optimize later)
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			iWR := 0.0
			jWR := 0.0

			if cards[i].HasDraftStats() {
				iWR = cards[i].DraftStats.GIHWR
			}
			if cards[j].HasDraftStats() {
				jWR = cards[j].DraftStats.GIHWR
			}

			if jWR > iWR {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}
}

// containsAllColors checks if a card's color identity contains all specified colors.
func containsAllColors(cardColors, requiredColors []string) bool {
	colorSet := make(map[string]bool)
	for _, c := range cardColors {
		colorSet[c] = true
	}

	for _, rc := range requiredColors {
		if !colorSet[rc] {
			return false
		}
	}

	return true
}

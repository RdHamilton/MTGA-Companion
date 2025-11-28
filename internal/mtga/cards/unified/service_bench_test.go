package unified

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// BenchmarkGetCard benchmarks single card retrieval.
func BenchmarkGetCard(b *testing.B) {
	arenaID := 12345
	metadata := &mockMetadataProvider{
		cards: map[int]*models.SetCard{
			arenaID: {
				ID:       1,
				ArenaID:  fmt.Sprintf("%d", arenaID),
				Name:     "Test Card",
				ManaCost: "{2}{R}",
				CMC:      3,
				Types:    []string{"Creature"},
				Rarity:   "rare",
				SetCode:  "BLB",
			},
		},
	}

	draftstats := &mockDraftStatsProvider{
		ratings: map[int]*storage.DraftCardRating{
			arenaID: {
				ArenaID:     arenaID,
				Expansion:   "BLB",
				Format:      "PremierDraft",
				GIHWR:       0.585,
				ALSA:        4.2,
				ATA:         2.1,
				GIH:         500,
				GamesPlayed: 450,
				LastUpdated: time.Now().Add(-1 * time.Hour),
			},
		},
	}

	service := NewService(metadata, draftstats)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetCard(ctx, arenaID, "BLB", "PremierDraft")
	}
}

// BenchmarkGetCards benchmarks batch card retrieval.
func BenchmarkGetCards(b *testing.B) {
	arenaID1 := 1111
	arenaID2 := 2222
	arenaID3 := 3333

	metadata := &mockMetadataProvider{
		cards: map[int]*models.SetCard{
			arenaID1: {
				ID:      1,
				ArenaID: fmt.Sprintf("%d", arenaID1),
				Name:    "Card 1",
				SetCode: "BLB",
			},
			arenaID2: {
				ID:      2,
				ArenaID: fmt.Sprintf("%d", arenaID2),
				Name:    "Card 2",
				SetCode: "BLB",
			},
			arenaID3: {
				ID:      3,
				ArenaID: fmt.Sprintf("%d", arenaID3),
				Name:    "Card 3",
				SetCode: "MKM",
			},
		},
	}

	draftstats := &mockDraftStatsProvider{
		setRatings: map[string][]*storage.DraftCardRating{
			"BLB": {
				{ArenaID: arenaID1, GIHWR: 0.55, GIH: 200, LastUpdated: time.Now()},
				{ArenaID: arenaID2, GIHWR: 0.60, GIH: 300, LastUpdated: time.Now()},
			},
			"MKM": {
				{ArenaID: arenaID3, GIHWR: 0.52, GIH: 150, LastUpdated: time.Now()},
			},
		},
	}

	service := NewService(metadata, draftstats)
	ctx := context.Background()
	arenaIDs := []int{arenaID1, arenaID2, arenaID3}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetCards(ctx, arenaIDs, "PremierDraft")
	}
}

// BenchmarkFilterByRarity benchmarks rarity filtering.
func BenchmarkFilterByRarity(b *testing.B) {
	cards := make([]*UnifiedCard, 100)
	for i := 0; i < 100; i++ {
		rarity := "common"
		if i%10 == 0 {
			rarity = "rare"
		} else if i%20 == 0 {
			rarity = "mythic"
		}
		cards[i] = &UnifiedCard{
			Name:   "Card",
			Rarity: rarity,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterByRarity(cards, "rare")
	}
}

// BenchmarkSortByWinRate benchmarks win rate sorting.
func BenchmarkSortByWinRate(b *testing.B) {
	cards := make([]*UnifiedCard, 50)
	for i := 0; i < 50; i++ {
		cards[i] = &UnifiedCard{
			Name: "Card",
			DraftStats: &DraftStatistics{
				GIHWR: float64(i) / 100.0,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Make a copy to avoid sorting the same pre-sorted array
		cardsCopy := make([]*UnifiedCard, len(cards))
		copy(cardsCopy, cards)
		SortByWinRate(cardsCopy)
	}
}

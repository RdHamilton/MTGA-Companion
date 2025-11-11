package query

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/unified"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// BenchmarkGet_CacheHit benchmarks cache hit scenario.
func BenchmarkGet_CacheHit(b *testing.B) {
	arenaID := 12345
	mockUnified := &mockUnifiedService{}
	mockStore := &mockStorage{
		cards: map[int]*storage.Card{
			arenaID: {
				ID:      "test-id",
				ArenaID: &arenaID,
				Name:    "Cached Card",
				SetCode: "BLB",
			},
		},
		ratings: map[int]*storage.DraftCardRating{
			arenaID: {
				ArenaID:     arenaID,
				GIHWR:       0.585,
				GIH:         500,
				LastUpdated: time.Now().Add(-1 * time.Hour),
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		b.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = query.Get(ctx, arenaID, opts)
	}
}

// BenchmarkGet_CacheMiss benchmarks cache miss scenario.
func BenchmarkGet_CacheMiss(b *testing.B) {
	arenaID := 12345
	mockUnified := &mockUnifiedService{
		cards: map[int]*unified.UnifiedCard{
			arenaID: {
				ArenaID: arenaID,
				Name:    "Fresh Card",
				DraftStats: &unified.DraftStatistics{
					GIHWR:       0.600,
					GIH:         200,
					LastUpdated: time.Now(),
				},
			},
		},
	}
	mockStore := &mockStorage{}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		b.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = query.Get(ctx, arenaID, opts)
	}
}

// BenchmarkGetMany_AllCached benchmarks batch query with all cached.
func BenchmarkGetMany_AllCached(b *testing.B) {
	arenaIDs := []int{1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999, 10000,
		11111, 12222, 13333, 14444, 15555, 16666, 17777, 18888, 19999, 20000}

	mockUnified := &mockUnifiedService{}
	mockStore := &mockStorage{
		cards:   make(map[int]*storage.Card),
		ratings: make(map[int]*storage.DraftCardRating),
	}

	// Populate cache with 20 cards
	for _, id := range arenaIDs {
		mockStore.cards[id] = &storage.Card{
			ID:      "test-id",
			ArenaID: &id,
			Name:    "Cached Card",
			SetCode: "BLB",
		}
		mockStore.ratings[id] = &storage.DraftCardRating{
			ArenaID:     id,
			GIHWR:       0.585,
			GIH:         500,
			LastUpdated: time.Now().Add(-1 * time.Hour),
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		b.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = query.GetMany(ctx, arenaIDs, opts)
	}
}

// BenchmarkGetMany_Mixed benchmarks batch query with mixed cached/fresh.
func BenchmarkGetMany_Mixed(b *testing.B) {
	cachedIDs := []int{1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999, 10000}
	freshIDs := []int{11111, 12222, 13333, 14444, 15555, 16666, 17777, 18888, 19999, 20000}
	allIDs := append(cachedIDs, freshIDs...)

	mockUnified := &mockUnifiedService{
		cards: make(map[int]*unified.UnifiedCard),
	}

	// Populate fresh cards in unified service
	for _, id := range freshIDs {
		mockUnified.cards[id] = &unified.UnifiedCard{
			ArenaID: id,
			Name:    "Fresh Card",
			DraftStats: &unified.DraftStatistics{
				GIHWR:       0.600,
				GIH:         200,
				LastUpdated: time.Now(),
			},
		}
	}

	mockStore := &mockStorage{
		cards:   make(map[int]*storage.Card),
		ratings: make(map[int]*storage.DraftCardRating),
	}

	// Populate cached cards in storage
	for _, id := range cachedIDs {
		mockStore.cards[id] = &storage.Card{
			ID:      "test-id",
			ArenaID: &id,
			Name:    "Cached Card",
			SetCode: "BLB",
		}
		mockStore.ratings[id] = &storage.DraftCardRating{
			ArenaID:     id,
			GIHWR:       0.585,
			GIH:         500,
			LastUpdated: time.Now().Add(-1 * time.Hour),
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		b.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = query.GetMany(ctx, allIDs, opts)
	}
}

// BenchmarkSearch benchmarks search functionality.
func BenchmarkSearch(b *testing.B) {
	arenaID := 12345
	mockUnified := &mockUnifiedService{}
	mockStore := &mockStorage{
		cards: map[int]*storage.Card{
			arenaID: {
				ID:      "test-id",
				ArenaID: &arenaID,
				Name:    "Lightning Bolt",
				SetCode: "BLB",
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		b.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := DefaultQueryOptions()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = query.Search(ctx, "Lightning Bolt", opts)
	}
}

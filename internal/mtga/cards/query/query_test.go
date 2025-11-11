package query

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/unified"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// mockUnifiedService implements unified.Service for testing.
type mockUnifiedService struct {
	cards    map[int]*unified.UnifiedCard
	setCards map[string][]*unified.UnifiedCard
	err      error
}

func (m *mockUnifiedService) GetCard(ctx context.Context, arenaID int, setCode, format string) (*unified.UnifiedCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	card, ok := m.cards[arenaID]
	if !ok {
		return nil, nil
	}
	return card, nil
}

func (m *mockUnifiedService) GetCards(ctx context.Context, arenaIDs []int, format string) ([]*unified.UnifiedCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	var cards []*unified.UnifiedCard
	for _, id := range arenaIDs {
		if card, ok := m.cards[id]; ok {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func (m *mockUnifiedService) GetSetCards(ctx context.Context, setCode, format string) ([]*unified.UnifiedCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.setCards[setCode], nil
}

// mockStorage implements minimal storage methods for testing.
type mockStorage struct {
	cards   map[int]*storage.Card
	ratings map[int]*storage.DraftCardRating
	err     error
}

func (m *mockStorage) GetCardByArenaID(ctx context.Context, arenaID int) (*storage.Card, error) {
	if m.err != nil {
		return nil, m.err
	}
	card, ok := m.cards[arenaID]
	if !ok {
		return nil, fmt.Errorf("card not found")
	}
	return card, nil
}

func (m *mockStorage) GetCardRating(ctx context.Context, arenaID int, expansion, format, colors string) (*storage.DraftCardRating, error) {
	if m.err != nil {
		return nil, m.err
	}
	rating, ok := m.ratings[arenaID]
	if !ok {
		return nil, fmt.Errorf("rating not found")
	}
	return rating, nil
}

func (m *mockStorage) SearchCards(ctx context.Context, name string) ([]*storage.Card, error) {
	if m.err != nil {
		return nil, m.err
	}
	var results []*storage.Card
	for _, card := range m.cards {
		if card.Name == name {
			results = append(results, card)
		}
	}
	return results, nil
}

// TestFallbackMode_String tests FallbackMode string representation.
func TestFallbackMode_String(t *testing.T) {
	tests := []struct {
		mode FallbackMode
		want string
	}{
		{RequireAll, "RequireAll"},
		{AllowPartial, "AllowPartial"},
		{CacheOnly, "CacheOnly"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("FallbackMode.String() = %s, want %s", got, tt.want)
		}
	}
}

// TestDefaultQueryOptions tests default query options.
func TestDefaultQueryOptions(t *testing.T) {
	opts := DefaultQueryOptions()

	if opts.Format != "PremierDraft" {
		t.Errorf("Expected format PremierDraft, got: %s", opts.Format)
	}
	if !opts.IncludeStats {
		t.Error("Expected IncludeStats to be true")
	}
	if opts.MaxStaleAge != 24*time.Hour {
		t.Errorf("Expected MaxStaleAge 24h, got: %v", opts.MaxStaleAge)
	}
	if opts.FallbackMode != AllowPartial {
		t.Errorf("Expected FallbackMode AllowPartial, got: %v", opts.FallbackMode)
	}
}

// TestNewCardQuery tests creating a new card query service.
func TestNewCardQuery(t *testing.T) {
	arenaID := 12345
	mockUnified := &mockUnifiedService{
		cards: map[int]*unified.UnifiedCard{
			arenaID: {
				ArenaID: arenaID,
				Name:    "Test Card",
			},
		},
	}
	mockStore := &mockStorage{}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Valid config
	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer query.Close()

	// Missing unified service
	_, err = NewCardQuery(QueryConfig{
		Storage: mockStore,
		Logger:  logger,
	})
	if err == nil {
		t.Error("Expected error for missing unified service")
	}

	// Missing storage
	_, err = NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Logger:         logger,
	})
	if err == nil {
		t.Error("Expected error for missing storage")
	}
}

// TestGet_CacheHit tests cache hit scenario.
func TestGet_CacheHit(t *testing.T) {
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
				LastUpdated: time.Now().Add(-1 * time.Hour), // Fresh
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
		t.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()
	card, err := query.Get(ctx, arenaID, opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if card.Name != "Cached Card" {
		t.Errorf("Expected name 'Cached Card', got: %s", card.Name)
	}
	if !card.HasDraftStats() {
		t.Error("Expected card to have draft stats")
	}
}

// TestGet_CacheMiss tests cache miss scenario.
func TestGet_CacheMiss(t *testing.T) {
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
		t.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()
	card, err := query.Get(ctx, arenaID, opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if card.Name != "Fresh Card" {
		t.Errorf("Expected name 'Fresh Card', got: %s", card.Name)
	}
	if card.DraftStats.GIHWR != 0.600 {
		t.Errorf("Expected GIHWR 0.600, got: %f", card.DraftStats.GIHWR)
	}
}

// TestGet_CacheOnly tests cache-only mode.
func TestGet_CacheOnly(t *testing.T) {
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
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	query, err := NewCardQuery(QueryConfig{
		UnifiedService: mockUnified,
		Storage:        mockStore,
		Logger:         logger,
		RefreshWorkers: 1,
	})
	if err != nil {
		t.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		MaxStaleAge:  1 * time.Hour,
		FallbackMode: CacheOnly,
	}

	ctx := context.Background()
	card, err := query.Get(ctx, arenaID, opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if card.Name != "Cached Card" {
		t.Errorf("Expected name 'Cached Card', got: %s", card.Name)
	}
}

// TestGetMany_Batch tests batch query optimization.
func TestGetMany_Batch(t *testing.T) {
	arenaID1 := 1111
	arenaID2 := 2222
	arenaID3 := 3333

	mockUnified := &mockUnifiedService{
		cards: map[int]*unified.UnifiedCard{
			arenaID3: {
				ArenaID: arenaID3,
				Name:    "Fresh Card 3",
			},
		},
	}

	mockStore := &mockStorage{
		cards: map[int]*storage.Card{
			arenaID1: {
				ID:      "id-1",
				ArenaID: &arenaID1,
				Name:    "Cached Card 1",
				SetCode: "BLB",
			},
			arenaID2: {
				ID:      "id-2",
				ArenaID: &arenaID2,
				Name:    "Cached Card 2",
				SetCode: "BLB",
			},
		},
		ratings: map[int]*storage.DraftCardRating{
			arenaID1: {
				ArenaID:     arenaID1,
				GIHWR:       0.55,
				GIH:         200,
				LastUpdated: time.Now().Add(-1 * time.Hour),
			},
			arenaID2: {
				ArenaID:     arenaID2,
				GIHWR:       0.60,
				GIH:         300,
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
		t.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour,
		FallbackMode: AllowPartial,
	}

	ctx := context.Background()
	cards, err := query.GetMany(ctx, []int{arenaID1, arenaID2, arenaID3}, opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(cards) != 3 {
		t.Fatalf("Expected 3 cards, got: %d", len(cards))
	}

	// Verify mix of cached and fresh cards
	names := make(map[string]bool)
	for _, card := range cards {
		names[card.Name] = true
	}

	if !names["Cached Card 1"] {
		t.Error("Missing Cached Card 1")
	}
	if !names["Cached Card 2"] {
		t.Error("Missing Cached Card 2")
	}
	if !names["Fresh Card 3"] {
		t.Error("Missing Fresh Card 3")
	}
}

// TestSearch tests search functionality.
func TestSearch(t *testing.T) {
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
		t.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := DefaultQueryOptions()
	ctx := context.Background()

	cards, err := query.Search(ctx, "Lightning Bolt", opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(cards) != 1 {
		t.Fatalf("Expected 1 card, got: %d", len(cards))
	}

	if cards[0].Name != "Lightning Bolt" {
		t.Errorf("Expected name 'Lightning Bolt', got: %s", cards[0].Name)
	}
}

// TestGetSet tests getting all cards in a set.
func TestGetSet(t *testing.T) {
	arenaID1 := 1111
	arenaID2 := 2222

	mockUnified := &mockUnifiedService{
		setCards: map[string][]*unified.UnifiedCard{
			"BLB": {
				{ArenaID: arenaID1, Name: "Card 1", SetCode: "BLB"},
				{ArenaID: arenaID2, Name: "Card 2", SetCode: "BLB"},
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
		t.Fatalf("Failed to create query: %v", err)
	}
	defer query.Close()

	opts := DefaultQueryOptions()
	ctx := context.Background()

	cards, err := query.GetSet(ctx, "BLB", opts)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("Expected 2 cards, got: %d", len(cards))
	}
}

// TestIsFresh tests freshness checking logic.
func TestIsFresh(t *testing.T) {
	q := &cardQuery{}

	// Fresh card
	fresh := &unified.UnifiedCard{
		MetadataAge: 1 * time.Hour,
		StatsAge:    30 * time.Minute,
		DraftStats:  &unified.DraftStatistics{},
	}

	if !q.isFresh(fresh, 24*time.Hour) {
		t.Error("Expected card to be fresh")
	}

	// Stale metadata
	staleMetadata := &unified.UnifiedCard{
		MetadataAge: 48 * time.Hour,
		StatsAge:    1 * time.Hour,
		DraftStats:  &unified.DraftStatistics{},
	}

	if q.isFresh(staleMetadata, 24*time.Hour) {
		t.Error("Expected card to be stale (metadata)")
	}

	// Stale stats
	staleStats := &unified.UnifiedCard{
		MetadataAge: 1 * time.Hour,
		StatsAge:    48 * time.Hour,
		DraftStats:  &unified.DraftStatistics{},
	}

	if q.isFresh(staleStats, 24*time.Hour) {
		t.Error("Expected card to be stale (stats)")
	}

	// Require fresh (maxStaleAge = 0)
	if q.isFresh(fresh, 0) {
		t.Error("Expected fresh requirement to fail")
	}
}

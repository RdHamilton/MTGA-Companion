package gui

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockCardFetcher implements CardFetcher interface for testing
type mockCardFetcher struct {
	fetchResults   map[int]*models.SetCard
	fetchErrors    map[int]error
	fetchCallCount int
	fetchCalls     []int
	mu             sync.Mutex
}

func newMockCardFetcher() *mockCardFetcher {
	return &mockCardFetcher{
		fetchResults: make(map[int]*models.SetCard),
		fetchErrors:  make(map[int]error),
		fetchCalls:   make([]int, 0),
	}
}

func (m *mockCardFetcher) FetchCardByArenaID(ctx context.Context, arenaID int) (*models.SetCard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchCallCount++
	m.fetchCalls = append(m.fetchCalls, arenaID)

	if err, exists := m.fetchErrors[arenaID]; exists {
		return nil, err
	}

	if card, exists := m.fetchResults[arenaID]; exists {
		return card, nil
	}

	return nil, nil // Card not found
}

func (m *mockCardFetcher) FetchCardByName(_ context.Context, _, _, _ string) (*models.SetCard, error) {
	return nil, nil // Not used in GetCollection tests
}

func (m *mockCardFetcher) FetchAndCacheSet(_ context.Context, _ string) (int, error) {
	return 0, nil // Not used in GetCollection tests
}

func (m *mockCardFetcher) RefreshSet(_ context.Context, _ string) (int, error) {
	return 0, nil // Not used in GetCollection tests
}

func (m *mockCardFetcher) GetCardByArenaID(_ context.Context, _ string) (*models.SetCard, error) {
	return nil, nil // Not used in GetCollection tests
}

func (m *mockCardFetcher) setCardResult(arenaID int, card *models.SetCard) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fetchResults[arenaID] = card
}

func (m *mockCardFetcher) setCardError(arenaID int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fetchErrors[arenaID] = err
}

// setupAutoFetchTestDB creates in-memory SQLite with required schema
func setupAutoFetchTestDB(t *testing.T, collectionCards map[int]int, knownCards map[int]*models.SetCard) (*storage.Service, func()) {
	t.Helper()

	// Use storage.Open with in-memory database and memory journal mode
	// MaxIdleConns must be > 0 to keep the in-memory db alive between queries
	cfg := &storage.Config{
		Path:         ":memory:",
		JournalMode:  "MEMORY",
		Synchronous:  "OFF",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	}

	db, err := storage.Open(cfg)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	sqlDB := db.Conn()

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			screen_name TEXT,
			client_id TEXT,
			is_default INTEGER NOT NULL DEFAULT 0,
			daily_wins INTEGER NOT NULL DEFAULT 0,
			weekly_wins INTEGER NOT NULL DEFAULT 0,
			mastery_level INTEGER NOT NULL DEFAULT 0,
			mastery_pass TEXT NOT NULL DEFAULT '',
			mastery_max INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS collection (
			card_id INTEGER PRIMARY KEY,
			quantity INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS set_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			set_code TEXT NOT NULL,
			arena_id TEXT NOT NULL,
			scryfall_id TEXT,
			name TEXT NOT NULL,
			mana_cost TEXT,
			cmc REAL DEFAULT 0,
			types TEXT,
			colors TEXT,
			rarity TEXT,
			text TEXT,
			power TEXT,
			toughness TEXT,
			image_url TEXT,
			image_url_small TEXT,
			image_url_art TEXT,
			fetched_at TIMESTAMP,
			UNIQUE(set_code, arena_id)
		);

		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, datetime('now'), datetime('now'));
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Insert collection cards
	for cardID, qty := range collectionCards {
		if _, err := sqlDB.Exec("INSERT INTO collection (card_id, quantity) VALUES (?, ?)", cardID, qty); err != nil {
			t.Fatalf("failed to insert collection card %d: %v", cardID, err)
		}
	}

	// Insert known cards (cards with metadata)
	// Must include all columns that GetCardByArenaID scans (no NULL values)
	for arenaID, card := range knownCards {
		if _, err := sqlDB.Exec(
			`INSERT INTO set_cards (set_code, arena_id, scryfall_id, name, mana_cost, cmc,
				types, colors, rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at)
			VALUES (?, ?, '', ?, '', 0, '[]', '[]', ?, '', '', '', '', '', '', datetime('now'))`,
			card.SetCode, fmt.Sprintf("%d", arenaID), card.Name, card.Rarity,
		); err != nil {
			t.Fatalf("failed to insert known card %d: %v", arenaID, err)
		}
	}

	storageSvc := storage.NewService(db)

	cleanup := func() {
		_ = db.Close()
	}

	return storageSvc, cleanup
}

// setupCollectionFacadeWithMocks creates a CollectionFacade with injectable mocks
func setupCollectionFacadeWithMocks(
	t *testing.T,
	mockFetcher *mockCardFetcher,
	collectionCards map[int]int,
	knownCards map[int]*models.SetCard,
) (*CollectionFacade, func()) {
	t.Helper()

	storageSvc, cleanup := setupAutoFetchTestDB(t, collectionCards, knownCards)

	services := &Services{
		Context: context.Background(),
		Storage: storageSvc,
	}

	// Only set SetFetcher if mockFetcher is not nil
	// (assigning a nil pointer to an interface creates a non-nil interface)
	if mockFetcher != nil {
		services.SetFetcher = mockFetcher
	}

	facade := NewCollectionFacade(services)

	return facade, cleanup
}

func TestAutoFetch_TriggersWhenUnknownCardsExist(t *testing.T) {
	mockFetcher := newMockCardFetcher()
	mockFetcher.setCardResult(102, &models.SetCard{
		ArenaID: "102", Name: "Fetched Card 102", SetCode: "TST", Rarity: "common",
	})
	mockFetcher.setCardResult(103, &models.SetCard{
		ArenaID: "103", Name: "Fetched Card 103", SetCode: "TST", Rarity: "rare",
	})

	knownCards := map[int]*models.SetCard{
		101: {ArenaID: "101", Name: "Known Card", SetCode: "TST", Rarity: "common"},
	}

	collectionCards := map[int]int{101: 4, 102: 2, 103: 1}

	facade, cleanup := setupCollectionFacadeWithMocks(t, mockFetcher, collectionCards, knownCards)
	defer cleanup()

	response, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockFetcher.fetchCallCount != 2 {
		t.Errorf("expected 2 fetch calls, got %d", mockFetcher.fetchCallCount)
	}

	if response.UnknownCardsFetched != 2 {
		t.Errorf("expected UnknownCardsFetched=2, got %d", response.UnknownCardsFetched)
	}

	if response.UnknownCardsRemaining != 0 {
		t.Errorf("expected UnknownCardsRemaining=0, got %d", response.UnknownCardsRemaining)
	}
}

func TestAutoFetch_RateLimiting_MaxAutoLookups(t *testing.T) {
	mockFetcher := newMockCardFetcher()

	// Create 15 unknown cards, all fetchable
	for i := 100; i < 115; i++ {
		mockFetcher.setCardResult(i, &models.SetCard{
			ArenaID: fmt.Sprintf("%d", i),
			Name:    fmt.Sprintf("Card %d", i),
			SetCode: "TST",
			Rarity:  "common",
		})
	}

	knownCards := map[int]*models.SetCard{} // No known cards

	collectionCards := make(map[int]int)
	for i := 100; i < 115; i++ {
		collectionCards[i] = 1
	}

	facade, cleanup := setupCollectionFacadeWithMocks(t, mockFetcher, collectionCards, knownCards)
	defer cleanup()

	response, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// maxAutoLookups is 10
	if mockFetcher.fetchCallCount != 10 {
		t.Errorf("expected exactly 10 fetch calls (maxAutoLookups), got %d", mockFetcher.fetchCallCount)
	}

	if response.UnknownCardsFetched != 10 {
		t.Errorf("expected UnknownCardsFetched=10, got %d", response.UnknownCardsFetched)
	}

	// 15 unknown - 10 fetched = 5 remaining
	if response.UnknownCardsRemaining != 5 {
		t.Errorf("expected UnknownCardsRemaining=5, got %d", response.UnknownCardsRemaining)
	}
}

func TestAutoFetch_FailedLookupTracking(t *testing.T) {
	mockFetcher := newMockCardFetcher()
	mockFetcher.setCardError(101, errors.New("scryfall API error"))
	// 102 returns nil implicitly (not found)
	mockFetcher.setCardResult(103, &models.SetCard{
		ArenaID: "103", Name: "Found Card", SetCode: "TST", Rarity: "common",
	})

	knownCards := map[int]*models.SetCard{}
	collectionCards := map[int]int{101: 1, 102: 1, 103: 1}

	facade, cleanup := setupCollectionFacadeWithMocks(t, mockFetcher, collectionCards, knownCards)
	defer cleanup()

	// First call
	response, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify only 1 was successfully fetched
	if response.UnknownCardsFetched != 1 {
		t.Errorf("expected UnknownCardsFetched=1, got %d", response.UnknownCardsFetched)
	}

	// Verify failed lookups are tracked
	facade.lookupMu.RLock()
	if _, exists := facade.failedLookups[101]; !exists {
		t.Error("expected card 101 to be in failedLookups after error")
	}
	if _, exists := facade.failedLookups[102]; !exists {
		t.Error("expected card 102 to be in failedLookups after nil result")
	}
	facade.lookupMu.RUnlock()

	// Second call - failed cards (101, 102) should NOT be retried within cooldown
	// However, card 103 (successfully fetched) is NOT cached in DB, so it will be fetched again
	initialCallCount := mockFetcher.fetchCallCount
	_, err = facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	// Only card 103 should be fetched again (not cached, not in failedLookups)
	// Cards 101 and 102 are in failedLookups and should NOT be refetched
	newFetchCalls := mockFetcher.fetchCallCount - initialCallCount
	if newFetchCalls != 1 {
		t.Errorf("expected 1 new fetch call (for card 103), got %d", newFetchCalls)
	}

	// Verify which cards were fetched on second call
	mockFetcher.mu.Lock()
	lastFetched := mockFetcher.fetchCalls[len(mockFetcher.fetchCalls)-1]
	mockFetcher.mu.Unlock()
	if lastFetched != 103 {
		t.Errorf("expected last fetch to be card 103, got %d", lastFetched)
	}
}

func TestAutoFetch_CooldownExpiry_AllowsRetry(t *testing.T) {
	mockFetcher := newMockCardFetcher()
	mockFetcher.setCardResult(101, &models.SetCard{
		ArenaID: "101", Name: "Now Available", SetCode: "TST", Rarity: "common",
	})

	knownCards := map[int]*models.SetCard{}
	collectionCards := map[int]int{101: 1}

	facade, cleanup := setupCollectionFacadeWithMocks(t, mockFetcher, collectionCards, knownCards)
	defer cleanup()

	// Pre-populate with expired failed lookup (2 hours ago)
	expiredTime := time.Now().Add(-2 * time.Hour)
	facade.lookupMu.Lock()
	facade.failedLookups[101] = expiredTime
	facade.lookupMu.Unlock()

	response, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockFetcher.fetchCallCount != 1 {
		t.Errorf("expected 1 fetch call after cooldown expiry, got %d", mockFetcher.fetchCallCount)
	}

	if response.UnknownCardsFetched != 1 {
		t.Errorf("expected UnknownCardsFetched=1 after retry, got %d", response.UnknownCardsFetched)
	}
}

func TestAutoFetch_NoSetFetcher_SkipsAutoFetch(t *testing.T) {
	knownCards := map[int]*models.SetCard{}
	collectionCards := map[int]int{101: 1, 102: 1}

	// Pass nil for SetFetcher
	facade, cleanup := setupCollectionFacadeWithMocks(t, nil, collectionCards, knownCards)
	defer cleanup()

	response, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.UnknownCardsFetched != 0 {
		t.Errorf("expected UnknownCardsFetched=0 with nil SetFetcher, got %d", response.UnknownCardsFetched)
	}

	// unknownCardsRemaining should still reflect the count of unknown cards
	if response.UnknownCardsRemaining != 2 {
		t.Errorf("expected UnknownCardsRemaining=2, got %d", response.UnknownCardsRemaining)
	}
}

func TestAutoFetch_ResponseFieldsCorrectlyComputed(t *testing.T) {
	mockFetcher := newMockCardFetcher()
	mockFetcher.setCardResult(101, &models.SetCard{ArenaID: "101", Name: "Card 1", SetCode: "TST", Rarity: "common"})
	mockFetcher.setCardResult(102, &models.SetCard{ArenaID: "102", Name: "Card 2", SetCode: "TST", Rarity: "rare"})
	mockFetcher.setCardError(103, errors.New("API error"))

	knownCards := map[int]*models.SetCard{
		100: {ArenaID: "100", Name: "Known Card", SetCode: "TST", Rarity: "uncommon"},
	}

	collectionCards := map[int]int{
		100: 4, // Known
		101: 3, // Unknown, will be fetched
		102: 2, // Unknown, will be fetched
		103: 1, // Unknown, will fail
	}

	facade, cleanup := setupCollectionFacadeWithMocks(t, mockFetcher, collectionCards, knownCards)
	defer cleanup()

	response, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 successfully fetched
	if response.UnknownCardsFetched != 2 {
		t.Errorf("expected UnknownCardsFetched=2, got %d", response.UnknownCardsFetched)
	}

	// 0 remaining after processing (103 failed but still processed)
	if response.UnknownCardsRemaining != 0 {
		t.Errorf("expected UnknownCardsRemaining=0, got %d", response.UnknownCardsRemaining)
	}

	// Total cards should be 4
	if len(response.Cards) != 4 {
		t.Errorf("expected 4 cards in response, got %d", len(response.Cards))
	}
}

func TestAutoFetch_ConcurrentSafety(t *testing.T) {
	mockFetcher := newMockCardFetcher()
	for i := 100; i < 105; i++ {
		mockFetcher.setCardResult(i, &models.SetCard{
			ArenaID: fmt.Sprintf("%d", i),
			Name:    fmt.Sprintf("Card %d", i),
			SetCode: "TST",
			Rarity:  "common",
		})
	}

	knownCards := map[int]*models.SetCard{}
	collectionCards := make(map[int]int)
	for i := 100; i < 105; i++ {
		collectionCards[i] = 1
	}

	facade, cleanup := setupCollectionFacadeWithMocks(t, mockFetcher, collectionCards, knownCards)
	defer cleanup()

	// Run concurrent calls
	var wg sync.WaitGroup
	errChan := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := facade.GetCollection(context.Background(), &CollectionFilter{OwnedOnly: true})
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent call failed: %v", err)
	}
}

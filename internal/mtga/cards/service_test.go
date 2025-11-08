package cards

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	config := DefaultServiceConfig()
	config.FallbackToAPI = false // Disable API calls in tests

	service, err := NewService(db, config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.cache == nil {
		t.Error("Cache should be initialized when EnableCache is true")
	}
}

func TestService_SaveAndGetCard(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	config := DefaultServiceConfig()
	config.FallbackToAPI = false

	service, err := NewService(db, config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	// Create a test card
	testCard := &Card{
		ArenaID:         12345,
		ScryfallID:      "test-scryfall-id",
		Name:            "Lightning Bolt",
		TypeLine:        "Instant",
		SetCode:         "LEA",
		SetName:         "Limited Edition Alpha",
		CMC:             1,
		Colors:          []string{"R"},
		ColorIdentity:   []string{"R"},
		Rarity:          "common",
		Layout:          "normal",
		CollectorNumber: "161",
		ReleasedAt:      time.Now(),
	}

	manaCost := "{R}"
	testCard.ManaCost = &manaCost

	// Save card to database
	err = service.saveCardToDB(testCard)
	if err != nil {
		t.Fatalf("saveCardToDB() error = %v", err)
	}

	// Retrieve card from database
	retrievedCard, err := service.getCardFromDB(12345)
	if err != nil {
		t.Fatalf("getCardFromDB() error = %v", err)
	}

	if retrievedCard == nil {
		t.Fatal("getCardFromDB() returned nil")
	}

	if retrievedCard.Name != "Lightning Bolt" {
		t.Errorf("Expected name 'Lightning Bolt', got %q", retrievedCard.Name)
	}

	if retrievedCard.ArenaID != 12345 {
		t.Errorf("Expected Arena ID 12345, got %d", retrievedCard.ArenaID)
	}
}

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(100, 1*time.Hour)

	testCard := &Card{
		ArenaID: 12345,
		Name:    "Test Card",
	}

	// Set card in cache
	cache.Set(12345, testCard)

	// Get card from cache
	retrieved := cache.Get(12345)
	if retrieved == nil {
		t.Fatal("Get() returned nil for existing card")
	}

	if retrieved.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got %q", retrieved.Name)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100, 100*time.Millisecond)

	testCard := &Card{
		ArenaID: 12345,
		Name:    "Test Card",
	}

	// Set card in cache
	cache.Set(12345, testCard)

	// Immediately get should succeed
	if retrieved := cache.Get(12345); retrieved == nil {
		t.Error("Card should be available immediately after setting")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Get should return nil after expiration
	if retrieved := cache.Get(12345); retrieved != nil {
		t.Error("Card should be expired after TTL")
	}
}

func TestCache_Eviction(t *testing.T) {
	cache := NewCache(2, 1*time.Hour) // Small cache for testing eviction

	card1 := &Card{ArenaID: 1, Name: "Card 1"}
	card2 := &Card{ArenaID: 2, Name: "Card 2"}
	card3 := &Card{ArenaID: 3, Name: "Card 3"}

	cache.Set(1, card1)
	cache.Set(2, card2)

	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}

	// Add third card, should evict first one
	cache.Set(3, card3)

	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2 after eviction, got %d", cache.Size())
	}

	// First card should be evicted
	if card := cache.Get(1); card != nil {
		t.Error("Card 1 should have been evicted")
	}

	// Other cards should still be present
	if card := cache.Get(2); card == nil {
		t.Error("Card 2 should still be in cache")
	}

	if card := cache.Get(3); card == nil {
		t.Error("Card 3 should be in cache")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(100, 1*time.Hour)

	cache.Set(1, &Card{ArenaID: 1, Name: "Card 1"})
	cache.Set(2, &Card{ArenaID: 2, Name: "Card 2"})

	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}

	if card := cache.Get(1); card != nil {
		t.Error("Cache should be empty after clear")
	}
}

func TestScryfallCard_ToCard(t *testing.T) {
	scryfallCard := &ScryfallCard{
		ID:              "test-id",
		OracleID:        "oracle-id",
		ArenaID:         12345,
		Name:            "Lightning Bolt",
		Lang:            "en",
		ReleasedAt:      "1993-08-05",
		Layout:          "normal",
		ManaCost:        "{R}",
		CMC:             1,
		TypeLine:        "Instant",
		OracleText:      "Lightning Bolt deals 3 damage to any target.",
		Power:           "",
		Toughness:       "",
		Colors:          []string{"R"},
		ColorIdentity:   []string{"R"},
		Set:             "LEA",
		SetName:         "Limited Edition Alpha",
		CollectorNumber: "161",
		Rarity:          "common",
	}

	card := scryfallCard.ToCard()

	if card.ArenaID != 12345 {
		t.Errorf("Expected Arena ID 12345, got %d", card.ArenaID)
	}

	if card.Name != "Lightning Bolt" {
		t.Errorf("Expected name 'Lightning Bolt', got %q", card.Name)
	}

	if card.CMC != 1 {
		t.Errorf("Expected CMC 1, got %f", card.CMC)
	}

	if card.ManaCost == nil || *card.ManaCost != "{R}" {
		t.Errorf("Expected mana cost '{R}', got %v", card.ManaCost)
	}

	if len(card.Colors) != 1 || card.Colors[0] != "R" {
		t.Errorf("Expected colors [R], got %v", card.Colors)
	}

	if card.Rarity != "common" {
		t.Errorf("Expected rarity 'common', got %q", card.Rarity)
	}
}

func TestService_GetCard_FromCache(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	config := DefaultServiceConfig()
	config.FallbackToAPI = false

	service, err := NewService(db, config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	// Add card to cache
	testCard := &Card{
		ArenaID: 12345,
		Name:    "Test Card",
	}
	service.cache.Set(12345, testCard)

	// Get card should retrieve from cache
	card, err := service.GetCard(12345)
	if err != nil {
		t.Fatalf("GetCard() error = %v", err)
	}

	if card == nil {
		t.Fatal("GetCard() returned nil")
	}

	if card.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got %q", card.Name)
	}
}

func TestService_GetCard_FromDB(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	config := DefaultServiceConfig()
	config.FallbackToAPI = false
	config.EnableCache = false // Disable cache to test DB

	service, err := NewService(db, config)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	// Save card to DB
	testCard := &Card{
		ArenaID:         12345,
		ScryfallID:      "test-id",
		Name:            "Test Card",
		TypeLine:        "Instant",
		SetCode:         "TST",
		SetName:         "Test Set",
		CMC:             1,
		Colors:          []string{"R"},
		ColorIdentity:   []string{"R"},
		Rarity:          "common",
		Layout:          "normal",
		CollectorNumber: "1",
		ReleasedAt:      time.Now(),
	}

	err = service.saveCardToDB(testCard)
	if err != nil {
		t.Fatalf("saveCardToDB() error = %v", err)
	}

	// Recreate service to ensure cache is empty
	service, _ = NewService(db, config)

	// Get card should retrieve from DB
	card, err := service.GetCard(12345)
	if err != nil {
		t.Fatalf("GetCard() error = %v", err)
	}

	if card == nil {
		t.Fatal("GetCard() returned nil")
	}

	if card.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got %q", card.Name)
	}
}

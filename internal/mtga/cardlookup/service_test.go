package cardlookup

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

func TestNewService(t *testing.T) {
	storageService := setupTestStorage(t)
	client := scryfall.NewClient()

	service := NewService(storageService, client, DefaultServiceOptions())
	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.storage == nil {
		t.Error("Service storage is nil")
	}

	if service.scryfallClient == nil {
		t.Error("Service scryfall client is nil")
	}

	if service.staleThreshold != 7*24*time.Hour {
		t.Errorf("Expected stale threshold 7 days, got %v", service.staleThreshold)
	}
}

func TestDefaultServiceOptions(t *testing.T) {
	opts := DefaultServiceOptions()

	if opts.StaleThreshold != 7*24*time.Hour {
		t.Errorf("Expected default stale threshold 7 days, got %v", opts.StaleThreshold)
	}
}

func TestGetCard_CacheHit(t *testing.T) {
	storageService := setupTestStorage(t)
	client := scryfall.NewClient()
	service := NewService(storageService, client, DefaultServiceOptions())

	// Pre-populate cache with a card
	arenaID := 12345
	cachedCard := &storage.Card{
		ID:          "test-card-id",
		ArenaID:     &arenaID,
		Name:        "Lightning Bolt",
		ManaCost:    "{R}",
		CMC:         1.0,
		TypeLine:    "Instant",
		OracleText:  "Deal 3 damage to any target.",
		Colors:      []string{"R"},
		Rarity:      "common",
		SetCode:     "blb",
		Layout:      "normal",
		ReleasedAt:  "2024-08-02",
		LastUpdated: time.Now(), // Fresh
	}

	ctx := context.Background()
	if err := storageService.SaveCard(ctx, cachedCard); err != nil {
		t.Fatalf("Failed to save test card: %v", err)
	}

	// Fetch card - should hit cache
	card, err := service.GetCard(arenaID)
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}

	if card == nil {
		t.Fatal("GetCard returned nil card")
	}

	if card.Name != "Lightning Bolt" {
		t.Errorf("Expected name 'Lightning Bolt', got %s", card.Name)
	}
}

func TestGetCard_StaleCacheFallback(t *testing.T) {
	storageService := setupTestStorage(t)
	client := scryfall.NewClient()

	// Use short stale threshold for testing
	opts := ServiceOptions{
		StaleThreshold: 1 * time.Millisecond,
	}
	service := NewService(storageService, client, opts)

	// Pre-populate cache with a stale card
	arenaID := 99999 // Non-existent card
	staleCard := &storage.Card{
		ID:          "stale-card-id",
		ArenaID:     &arenaID,
		Name:        "Stale Card",
		CMC:         0,
		TypeLine:    "Creature",
		Colors:      []string{},
		Rarity:      "common",
		SetCode:     "xxx",
		Layout:      "normal",
		ReleasedAt:  "2020-01-01",
		LastUpdated: time.Now().Add(-1 * time.Hour), // Very stale
	}

	ctx := context.Background()
	if err := storageService.SaveCard(ctx, staleCard); err != nil {
		t.Fatalf("Failed to save stale card: %v", err)
	}

	// Wait for card to become stale
	time.Sleep(2 * time.Millisecond)

	// Fetch card - should try Scryfall, fail, and fallback to stale cache
	card, err := service.GetCard(arenaID)
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}

	// Should return stale cache as fallback
	if card == nil {
		t.Fatal("GetCard returned nil card")
	}

	if card.Name != "Stale Card" {
		t.Errorf("Expected stale card as fallback, got %s", card.Name)
	}
}

func TestGetCards_Multiple(t *testing.T) {
	storageService := setupTestStorage(t)
	client := scryfall.NewClient()
	service := NewService(storageService, client, DefaultServiceOptions())

	// Pre-populate cache with multiple cards
	ctx := context.Background()
	arenaIDs := []int{10001, 10002, 10003}

	for i, id := range arenaIDs {
		card := &storage.Card{
			ID:          "test-card-" + string(rune('A'+i)),
			ArenaID:     &id,
			Name:        "Test Card " + string(rune('A'+i)),
			CMC:         float64(i + 1),
			TypeLine:    "Creature",
			Colors:      []string{"U"},
			Rarity:      "common",
			SetCode:     "tst",
			Layout:      "normal",
			ReleasedAt:  "2024-01-01",
			LastUpdated: time.Now(),
		}

		if err := storageService.SaveCard(ctx, card); err != nil {
			t.Fatalf("Failed to save test card: %v", err)
		}
	}

	// Fetch multiple cards
	cards, err := service.GetCards(arenaIDs)
	if err != nil {
		t.Fatalf("GetCards failed: %v", err)
	}

	if len(cards) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(cards))
	}

	// Verify all cards were fetched
	found := make(map[int]bool)
	for _, card := range cards {
		if card.ArenaID != nil {
			found[*card.ArenaID] = true
		}
	}

	for _, id := range arenaIDs {
		if !found[id] {
			t.Errorf("Card with Arena ID %d not found in results", id)
		}
	}
}

func TestSearchByName(t *testing.T) {
	storageService := setupTestStorage(t)
	client := scryfall.NewClient()
	service := NewService(storageService, client, DefaultServiceOptions())

	// Pre-populate cache with searchable cards
	ctx := context.Background()
	cards := []struct {
		arenaID int
		name    string
	}{
		{20001, "Lightning Bolt"},
		{20002, "Lightning Strike"},
		{20003, "Counterspell"},
	}

	for _, c := range cards {
		card := &storage.Card{
			ID:          "card-" + c.name,
			ArenaID:     &c.arenaID,
			Name:        c.name,
			CMC:         1.0,
			TypeLine:    "Instant",
			Colors:      []string{"R"},
			Rarity:      "common",
			SetCode:     "tst",
			Layout:      "normal",
			ReleasedAt:  "2024-01-01",
			LastUpdated: time.Now(),
		}

		if err := storageService.SaveCard(ctx, card); err != nil {
			t.Fatalf("Failed to save test card: %v", err)
		}
	}

	// Search for "Lightning"
	results, err := service.SearchByName("Lightning")
	if err != nil {
		t.Fatalf("SearchByName failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'Lightning', got %d", len(results))
	}
}

func TestGetCardsBySet(t *testing.T) {
	storageService := setupTestStorage(t)
	client := scryfall.NewClient()
	service := NewService(storageService, client, DefaultServiceOptions())

	// Pre-populate cache with cards from different sets
	ctx := context.Background()
	sets := []struct {
		arenaID int
		setCode string
		name    string
	}{
		{30001, "blb", "Card A"},
		{30002, "blb", "Card B"},
		{30003, "woe", "Card C"},
	}

	for _, s := range sets {
		card := &storage.Card{
			ID:              "card-" + s.name,
			ArenaID:         &s.arenaID,
			Name:            s.name,
			SetCode:         s.setCode,
			CollectorNumber: "001",
			CMC:             1.0,
			TypeLine:        "Creature",
			Colors:          []string{},
			Rarity:          "common",
			Layout:          "normal",
			ReleasedAt:      "2024-01-01",
			LastUpdated:     time.Now(),
		}

		if err := storageService.SaveCard(ctx, card); err != nil {
			t.Fatalf("Failed to save test card: %v", err)
		}
	}

	// Get cards from BLB set
	results, err := service.GetCardsBySet("blb")
	if err != nil {
		t.Fatalf("GetCardsBySet failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 cards from BLB set, got %d", len(results))
	}
}

func TestConvertToStorageCard(t *testing.T) {
	arenaID := 54321
	scryfallCard := &scryfall.Card{
		ID:              "scryfall-test-id",
		ArenaID:         &arenaID,
		Name:            "Test Card",
		ManaCost:        "{2}{U}",
		CMC:             3.0,
		TypeLine:        "Creature - Wizard",
		OracleText:      "Flying",
		Colors:          []string{"U"},
		ColorIdentity:   []string{"U"},
		Rarity:          "rare",
		SetCode:         "tst",
		CollectorNumber: "042",
		Power:           "2",
		Toughness:       "3",
		Layout:          "normal",
		ReleasedAt:      "2024-01-15",
		ImageURIs: &scryfall.ImageURIs{
			Small:  "https://example.com/small.jpg",
			Normal: "https://example.com/normal.jpg",
			Large:  "https://example.com/large.jpg",
		},
		Legalities: scryfall.Legalities{
			Standard: "legal",
			Modern:   "legal",
		},
	}

	storageCard := convertToStorageCard(scryfallCard)

	if storageCard.ID != "scryfall-test-id" {
		t.Errorf("Expected ID 'scryfall-test-id', got %s", storageCard.ID)
	}

	if storageCard.ArenaID == nil || *storageCard.ArenaID != 54321 {
		t.Error("Expected ArenaID 54321")
	}

	if storageCard.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got %s", storageCard.Name)
	}

	if storageCard.CMC != 3.0 {
		t.Errorf("Expected CMC 3.0, got %.1f", storageCard.CMC)
	}

	if len(storageCard.Colors) != 1 || storageCard.Colors[0] != "U" {
		t.Errorf("Expected colors [U], got %v", storageCard.Colors)
	}

	if storageCard.ImageURIs == nil {
		t.Error("Expected ImageURIs to be populated")
	}

	if storageCard.Legalities["standard"] != "legal" {
		t.Errorf("Expected standard legality 'legal', got %s", storageCard.Legalities["standard"])
	}

	if storageCard.LastUpdated.IsZero() {
		t.Error("Expected LastUpdated to be set")
	}
}

// setupTestStorage creates a test storage service with a temporary database.
func setupTestStorage(t *testing.T) *storage.Service {
	t.Helper()

	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	// Run migrations first
	migrationMgr, err := storage.NewMigrationManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	if err := migrationMgr.Up(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Close migration manager
	_ = migrationMgr.Close()

	// Open database
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create service
	service := storage.NewService(db)

	// Cleanup function
	t.Cleanup(func() {
		_ = db.Close()
	})

	return service
}

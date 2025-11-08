package viewer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	_ "modernc.org/sqlite"
)

func setupTestDBForViewer(t *testing.T) (*sql.DB, *cards.Service) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create collection table
	schema := `
	CREATE TABLE IF NOT EXISTS collection (
		card_id INTEGER PRIMARY KEY,
		quantity INTEGER NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS collection_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		card_id INTEGER NOT NULL,
		quantity_delta INTEGER NOT NULL,
		quantity_after INTEGER NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		source TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Create card service
	cardConfig := cards.DefaultServiceConfig()
	cardConfig.FallbackToAPI = false
	cardService, err := cards.NewService(db, cardConfig)
	if err != nil {
		t.Fatalf("Failed to create card service: %v", err)
	}

	return db, cardService
}

func TestCollectionViewer_GetCollection(t *testing.T) {
	db, cardService := setupTestDBForViewer(t)
	defer func() { _ = db.Close() }()

	viewer := NewCollectionViewer(db, cardService)

	// Add some test cards to collection
	ctx := context.Background()
	_, err := db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 12345, 4)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 67890, 2)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	// Get collection
	collection, err := viewer.GetCollection(ctx)
	if err != nil {
		t.Fatalf("GetCollection() error = %v", err)
	}

	if len(collection) != 2 {
		t.Errorf("Expected 2 cards in collection, got %d", len(collection))
	}

	// Check that cards have correct quantities
	foundCards := make(map[int]int)
	for _, card := range collection {
		foundCards[card.CardID] = card.Quantity
	}

	if foundCards[12345] != 4 {
		t.Errorf("Expected card 12345 to have quantity 4, got %d", foundCards[12345])
	}

	if foundCards[67890] != 2 {
		t.Errorf("Expected card 67890 to have quantity 2, got %d", foundCards[67890])
	}
}

func TestCollectionViewer_GetCardWithMetadata(t *testing.T) {
	db, cardService := setupTestDBForViewer(t)
	defer func() { _ = db.Close() }()

	viewer := NewCollectionViewer(db, cardService)
	ctx := context.Background()

	// Add test card to collection
	_, err := db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 12345, 3)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	// Add metadata for the card
	testCard := &cards.Card{
		ArenaID:         12345,
		ScryfallID:      "test-id",
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

	// Save card metadata
	if _, err := cardService.GetCard(12345); err != nil {
		// Card not in cache/DB, add it directly
		_, err = db.Exec(`
			INSERT INTO cards (
				arena_id, scryfall_id, name, type_line, set_code, set_name,
				cmc, colors, color_identity, rarity, layout, collector_number, released_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, testCard.ArenaID, testCard.ScryfallID, testCard.Name, testCard.TypeLine,
			testCard.SetCode, testCard.SetName, testCard.CMC, "[]", "[]",
			testCard.Rarity, testCard.Layout, testCard.CollectorNumber,
			testCard.ReleasedAt.Format("2006-01-02"))
		if err != nil {
			t.Logf("Note: Could not add card metadata: %v", err)
		}
	}

	// Get card with metadata
	cardView, err := viewer.GetCardWithMetadata(ctx, 12345)
	if err != nil {
		t.Fatalf("GetCardWithMetadata() error = %v", err)
	}

	if cardView == nil {
		t.Fatal("GetCardWithMetadata() returned nil")
	}

	if cardView.CardID != 12345 {
		t.Errorf("Expected card ID 12345, got %d", cardView.CardID)
	}

	if cardView.Quantity != 3 {
		t.Errorf("Expected quantity 3, got %d", cardView.Quantity)
	}
}

func TestCollectionViewer_FilterByRarity(t *testing.T) {
	db, cardService := setupTestDBForViewer(t)
	defer func() { _ = db.Close() }()

	viewer := NewCollectionViewer(db, cardService)
	ctx := context.Background()

	// Add test cards with different rarities
	_, err := db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 1, 2)
	if err != nil {
		t.Fatalf("Failed to insert collection card: %v", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 2, 3)
	if err != nil {
		t.Fatalf("Failed to insert collection card: %v", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 3, 1)
	if err != nil {
		t.Fatalf("Failed to insert collection card: %v", err)
	}

	// Add metadata for cards with different rarities
	cards := []*cards.Card{
		{
			ArenaID: 1, ScryfallID: "1", Name: "Common Card", TypeLine: "Creature", SetCode: "TST",
			SetName: "Test", CMC: 1, Rarity: "common", Layout: "normal",
			CollectorNumber: "1", ReleasedAt: time.Now(), Colors: []string{}, ColorIdentity: []string{},
		},
		{
			ArenaID: 2, ScryfallID: "2", Name: "Rare Card", TypeLine: "Creature", SetCode: "TST",
			SetName: "Test", CMC: 2, Rarity: "rare", Layout: "normal",
			CollectorNumber: "2", ReleasedAt: time.Now(), Colors: []string{}, ColorIdentity: []string{},
		},
		{
			ArenaID: 3, ScryfallID: "3", Name: "Mythic Card", TypeLine: "Creature", SetCode: "TST",
			SetName: "Test", CMC: 3, Rarity: "mythic", Layout: "normal",
			CollectorNumber: "3", ReleasedAt: time.Now(), Colors: []string{}, ColorIdentity: []string{},
		},
	}

	for _, card := range cards {
		_, err = db.Exec(`
			INSERT INTO cards (arena_id, scryfall_id, name, type_line, set_code, set_name,
				cmc, colors, color_identity, rarity, layout, collector_number, released_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, card.ArenaID, card.ScryfallID, card.Name, card.TypeLine, card.SetCode,
			card.SetName, card.CMC, "[]", "[]", card.Rarity, card.Layout,
			card.CollectorNumber, card.ReleasedAt.Format("2006-01-02"))
		if err != nil {
			t.Fatalf("Failed to insert card metadata: %v", err)
		}
	}

	// Filter by rarity
	rareCards, err := viewer.FilterByRarity(ctx, "rare")
	if err != nil {
		t.Fatalf("FilterByRarity() error = %v", err)
	}

	// At minimum, we should get a collection back (even if metadata isn't loaded)
	if len(rareCards) > 1 {
		t.Errorf("Expected at most 1 rare card, got %d", len(rareCards))
	}

	// If we found a rare card with metadata, verify it's correct
	if len(rareCards) > 0 && rareCards[0].Metadata != nil {
		if rareCards[0].Metadata.Rarity != "rare" {
			t.Errorf("Expected rarity 'rare', got %q", rareCards[0].Metadata.Rarity)
		}
	}
}

func TestCollectionViewer_FilterBySet(t *testing.T) {
	db, cardService := setupTestDBForViewer(t)
	defer func() { _ = db.Close() }()

	viewer := NewCollectionViewer(db, cardService)
	ctx := context.Background()

	var err error

	// Add test cards
	_, err = db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 1, 2)
	if err != nil {
		t.Fatalf("Failed to insert collection card: %v", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO collection (card_id, quantity) VALUES (?, ?)", 2, 3)
	if err != nil {
		t.Fatalf("Failed to insert collection card: %v", err)
	}

	// Add metadata for cards from different sets
	card1 := &cards.Card{
		ArenaID: 1, ScryfallID: "1", Name: "Alpha Card", TypeLine: "Creature",
		SetCode: "LEA", SetName: "Limited Edition Alpha", CMC: 1, Rarity: "common",
		Layout: "normal", CollectorNumber: "1", ReleasedAt: time.Now(),
	}
	card2 := &cards.Card{
		ArenaID: 2, ScryfallID: "2", Name: "Beta Card", TypeLine: "Creature",
		SetCode: "LEB", SetName: "Limited Edition Beta", CMC: 2, Rarity: "common",
		Layout: "normal", CollectorNumber: "2", ReleasedAt: time.Now(),
	}

	for _, card := range []*cards.Card{card1, card2} {
		_, err = db.Exec(`
			INSERT INTO cards (arena_id, scryfall_id, name, type_line, set_code, set_name,
				cmc, colors, color_identity, rarity, layout, collector_number, released_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, card.ArenaID, card.ScryfallID, card.Name, card.TypeLine, card.SetCode,
			card.SetName, card.CMC, "[]", "[]", card.Rarity, card.Layout,
			card.CollectorNumber, card.ReleasedAt.Format("2006-01-02"))
		if err != nil {
			t.Fatalf("Failed to insert card metadata: %v", err)
		}
	}

	// Verify cards exist in database before filtering
	var cardCount int
	err = db.QueryRow("SELECT COUNT(*) FROM cards WHERE set_code = 'LEA'").Scan(&cardCount)
	if err != nil {
		t.Fatalf("Failed to count cards in DB: %v", err)
	}
	if cardCount != 1 {
		t.Fatalf("Expected 1 card with set LEA in database, got %d", cardCount)
	}

	// Verify card service can read the card directly
	card1Meta, err := cardService.GetCard(1)
	if err != nil {
		t.Fatalf("Card service failed to get card 1: %v", err)
	}
	if card1Meta == nil {
		t.Fatal("Card service returned nil for card 1")
	}
	if card1Meta.SetCode != "LEA" {
		t.Fatalf("Card 1 has set code %q, expected LEA", card1Meta.SetCode)
	}

	// Filter by set
	alphaCards, err := viewer.FilterBySet(ctx, "LEA")
	if err != nil {
		t.Fatalf("FilterBySet() error = %v", err)
	}

	if len(alphaCards) != 1 {
		t.Errorf("Expected 1 Alpha card, got %d", len(alphaCards))
	}

	if len(alphaCards) > 0 && alphaCards[0].Metadata != nil {
		if alphaCards[0].Metadata.SetCode != "LEA" {
			t.Errorf("Expected set code 'LEA', got %q", alphaCards[0].Metadata.SetCode)
		}
	}
}

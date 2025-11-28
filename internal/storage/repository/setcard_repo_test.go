package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupSetCardTestDB creates an in-memory database with set_cards table.
func setupSetCardTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
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
		CREATE INDEX IF NOT EXISTS idx_set_cards_arena_id ON set_cards(arena_id);
		CREATE INDEX IF NOT EXISTS idx_set_cards_set_code ON set_cards(set_code);

		CREATE TABLE IF NOT EXISTS sets (
			code TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			released_at TEXT,
			card_count INTEGER,
			set_type TEXT,
			icon_svg_uri TEXT
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestSetCardRepository_GetMetadataStaleness(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	// Insert test cards with varying freshness
	now := time.Now()
	freshTime := now.Add(-1 * time.Hour)                        // 1 hour ago - fresh
	staleTime := now.Add(-10 * 24 * time.Hour)                  // 10 days ago - stale
	veryStaleTime := now.Add(-20 * 24 * time.Hour)              // 20 days ago - very stale
	staleAgeSeconds := int((7 * 24 * time.Hour).Seconds())      // 7 days
	veryStaleAgeSeconds := int((14 * 24 * time.Hour).Seconds()) // 14 days

	// Insert fresh card
	freshCard := &models.SetCard{
		SetCode:   "ONE",
		ArenaID:   "12345",
		Name:      "Fresh Card",
		Types:     []string{"Creature"},
		Colors:    []string{"W"},
		Rarity:    "common",
		FetchedAt: freshTime,
	}
	if err := repo.SaveCard(ctx, freshCard); err != nil {
		t.Fatalf("failed to save fresh card: %v", err)
	}

	// Insert stale card
	staleCard := &models.SetCard{
		SetCode:   "ONE",
		ArenaID:   "12346",
		Name:      "Stale Card",
		Types:     []string{"Creature"},
		Colors:    []string{"U"},
		Rarity:    "uncommon",
		FetchedAt: staleTime,
	}
	if err := repo.SaveCard(ctx, staleCard); err != nil {
		t.Fatalf("failed to save stale card: %v", err)
	}

	// Insert very stale card
	veryStaleCard := &models.SetCard{
		SetCode:   "ONE",
		ArenaID:   "12347",
		Name:      "Very Stale Card",
		Types:     []string{"Creature"},
		Colors:    []string{"B"},
		Rarity:    "rare",
		FetchedAt: veryStaleTime,
	}
	if err := repo.SaveCard(ctx, veryStaleCard); err != nil {
		t.Fatalf("failed to save very stale card: %v", err)
	}

	// Get staleness
	staleness, err := repo.GetMetadataStaleness(ctx, staleAgeSeconds, veryStaleAgeSeconds)
	if err != nil {
		t.Fatalf("failed to get metadata staleness: %v", err)
	}

	if staleness.Total != 3 {
		t.Errorf("expected total 3, got %d", staleness.Total)
	}

	if staleness.Fresh != 1 {
		t.Errorf("expected fresh 1, got %d", staleness.Fresh)
	}

	if staleness.Stale != 1 {
		t.Errorf("expected stale 1, got %d", staleness.Stale)
	}

	if staleness.VeryStale != 1 {
		t.Errorf("expected very stale 1, got %d", staleness.VeryStale)
	}
}

func TestSetCardRepository_GetStaleCards(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	now := time.Now()
	freshTime := now.Add(-1 * time.Hour)                   // 1 hour ago - fresh
	staleTime1 := now.Add(-10 * 24 * time.Hour)            // 10 days ago - stale (oldest)
	staleTime2 := now.Add(-8 * 24 * time.Hour)             // 8 days ago - stale
	staleAgeSeconds := int((7 * 24 * time.Hour).Seconds()) // 7 days

	// Insert fresh card
	freshCard := &models.SetCard{
		SetCode:   "ONE",
		ArenaID:   "12345",
		Name:      "Fresh Card",
		Types:     []string{"Creature"},
		Colors:    []string{"W"},
		Rarity:    "common",
		FetchedAt: freshTime,
	}
	if err := repo.SaveCard(ctx, freshCard); err != nil {
		t.Fatalf("failed to save fresh card: %v", err)
	}

	// Insert stale card 1 (oldest)
	staleCard1 := &models.SetCard{
		SetCode:   "ONE",
		ArenaID:   "12346",
		Name:      "Oldest Stale Card",
		Types:     []string{"Creature"},
		Colors:    []string{"U"},
		Rarity:    "uncommon",
		FetchedAt: staleTime1,
	}
	if err := repo.SaveCard(ctx, staleCard1); err != nil {
		t.Fatalf("failed to save stale card 1: %v", err)
	}

	// Insert stale card 2
	staleCard2 := &models.SetCard{
		SetCode:   "ONE",
		ArenaID:   "12347",
		Name:      "Newer Stale Card",
		Types:     []string{"Creature"},
		Colors:    []string{"B"},
		Rarity:    "rare",
		FetchedAt: staleTime2,
	}
	if err := repo.SaveCard(ctx, staleCard2); err != nil {
		t.Fatalf("failed to save stale card 2: %v", err)
	}

	// Get stale cards
	staleCards, err := repo.GetStaleCards(ctx, staleAgeSeconds, 10)
	if err != nil {
		t.Fatalf("failed to get stale cards: %v", err)
	}

	// Should have 2 stale cards (not the fresh one)
	if len(staleCards) != 2 {
		t.Errorf("expected 2 stale cards, got %d", len(staleCards))
	}

	// First card should be the oldest
	if len(staleCards) > 0 && staleCards[0].ArenaID != "12346" {
		t.Errorf("expected oldest card first (12346), got %s", staleCards[0].ArenaID)
	}

	// Test limit
	limitedCards, err := repo.GetStaleCards(ctx, staleAgeSeconds, 1)
	if err != nil {
		t.Fatalf("failed to get limited stale cards: %v", err)
	}

	if len(limitedCards) != 1 {
		t.Errorf("expected 1 limited stale card, got %d", len(limitedCards))
	}
}

func TestSetCardRepository_GetMetadataStaleness_EmptyDB(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	staleness, err := repo.GetMetadataStaleness(ctx, 604800, 1209600)
	if err != nil {
		t.Fatalf("failed to get metadata staleness from empty DB: %v", err)
	}

	if staleness.Total != 0 {
		t.Errorf("expected total 0, got %d", staleness.Total)
	}
}

func TestSetCardRepository_GetStaleCards_EmptyDB(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	staleCards, err := repo.GetStaleCards(ctx, 604800, 10)
	if err != nil {
		t.Fatalf("failed to get stale cards from empty DB: %v", err)
	}

	if len(staleCards) != 0 {
		t.Errorf("expected 0 stale cards, got %d", len(staleCards))
	}
}

func TestSetCardRepository_GetSetRarityCounts(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	// Insert set metadata
	_, err := db.ExecContext(ctx, `INSERT INTO sets (code, name) VALUES ('ONE', 'Phyrexia: All Will Be One')`)
	if err != nil {
		t.Fatalf("failed to insert set: %v", err)
	}

	// Insert cards with different rarities
	now := time.Now()
	cards := []*models.SetCard{
		{SetCode: "ONE", ArenaID: "12345", Name: "Common Card 1", Types: []string{"Creature"}, Colors: []string{"W"}, Rarity: "common", FetchedAt: now},
		{SetCode: "ONE", ArenaID: "12346", Name: "Common Card 2", Types: []string{"Creature"}, Colors: []string{"U"}, Rarity: "common", FetchedAt: now},
		{SetCode: "ONE", ArenaID: "12347", Name: "Uncommon Card", Types: []string{"Creature"}, Colors: []string{"B"}, Rarity: "uncommon", FetchedAt: now},
		{SetCode: "ONE", ArenaID: "12348", Name: "Rare Card", Types: []string{"Creature"}, Colors: []string{"R"}, Rarity: "rare", FetchedAt: now},
	}

	for _, card := range cards {
		if err := repo.SaveCard(ctx, card); err != nil {
			t.Fatalf("failed to save card: %v", err)
		}
	}

	// Get set rarity counts
	counts, err := repo.GetSetRarityCounts(ctx)
	if err != nil {
		t.Fatalf("failed to get set rarity counts: %v", err)
	}

	// Should have 3 entries (common, uncommon, rare for ONE)
	if len(counts) != 3 {
		t.Errorf("expected 3 rarity counts, got %d", len(counts))
	}

	// Check set name is populated from sets table
	for _, c := range counts {
		if c.SetCode == "ONE" && c.SetName != "Phyrexia: All Will Be One" {
			t.Errorf("expected set name 'Phyrexia: All Will Be One', got '%s'", c.SetName)
		}
	}

	// Find and verify common count
	for _, c := range counts {
		if c.SetCode == "ONE" && c.Rarity == "common" {
			if c.Total != 2 {
				t.Errorf("expected 2 common cards, got %d", c.Total)
			}
		}
	}
}

func TestSetCardRepository_GetSetRarityCounts_NoSetMetadata(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	// Insert card WITHOUT set metadata - should fall back to uppercase set code
	now := time.Now()
	card := &models.SetCard{
		SetCode:   "xyz",
		ArenaID:   "12345",
		Name:      "Test Card",
		Types:     []string{"Creature"},
		Colors:    []string{"W"},
		Rarity:    "common",
		FetchedAt: now,
	}
	if err := repo.SaveCard(ctx, card); err != nil {
		t.Fatalf("failed to save card: %v", err)
	}

	counts, err := repo.GetSetRarityCounts(ctx)
	if err != nil {
		t.Fatalf("failed to get set rarity counts: %v", err)
	}

	if len(counts) != 1 {
		t.Errorf("expected 1 rarity count, got %d", len(counts))
	}

	// Set name should fall back to uppercase set code
	if counts[0].SetName != "XYZ" {
		t.Errorf("expected set name 'XYZ' (fallback), got '%s'", counts[0].SetName)
	}
}

func TestSetCardRepository_GetAllCardSetInfo(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	// Insert cards
	now := time.Now()
	cards := []*models.SetCard{
		{SetCode: "ONE", ArenaID: "12345", Name: "Card 1", Types: []string{"Creature"}, Colors: []string{"W"}, Rarity: "common", FetchedAt: now},
		{SetCode: "ONE", ArenaID: "12346", Name: "Card 2", Types: []string{"Creature"}, Colors: []string{"U"}, Rarity: "uncommon", FetchedAt: now},
		{SetCode: "MOM", ArenaID: "12347", Name: "Card 3", Types: []string{"Creature"}, Colors: []string{"B"}, Rarity: "rare", FetchedAt: now},
	}

	for _, card := range cards {
		if err := repo.SaveCard(ctx, card); err != nil {
			t.Fatalf("failed to save card: %v", err)
		}
	}

	// Get all card set info
	infos, err := repo.GetAllCardSetInfo(ctx)
	if err != nil {
		t.Fatalf("failed to get card set info: %v", err)
	}

	if len(infos) != 3 {
		t.Errorf("expected 3 card infos, got %d", len(infos))
	}

	// Verify data is correctly mapped
	foundCard := false
	for _, info := range infos {
		if info.ArenaID == "12345" {
			foundCard = true
			if info.SetCode != "ONE" {
				t.Errorf("expected set code ONE, got %s", info.SetCode)
			}
			if info.Rarity != "common" {
				t.Errorf("expected rarity common, got %s", info.Rarity)
			}
		}
	}

	if !foundCard {
		t.Error("card with arena_id 12345 not found")
	}
}

func TestSetCardRepository_GetSetRarityCounts_EmptyDB(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	counts, err := repo.GetSetRarityCounts(ctx)
	if err != nil {
		t.Fatalf("failed to get set rarity counts from empty DB: %v", err)
	}

	if len(counts) != 0 {
		t.Errorf("expected 0 counts, got %d", len(counts))
	}
}

func TestSetCardRepository_GetAllCardSetInfo_EmptyDB(t *testing.T) {
	db := setupSetCardTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewSetCardRepository(db)
	ctx := context.Background()

	infos, err := repo.GetAllCardSetInfo(ctx)
	if err != nil {
		t.Fatalf("failed to get card set info from empty DB: %v", err)
	}

	if len(infos) != 0 {
		t.Errorf("expected 0 card infos, got %d", len(infos))
	}
}

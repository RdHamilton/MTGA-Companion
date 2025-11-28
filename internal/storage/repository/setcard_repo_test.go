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

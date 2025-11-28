package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// setupDraftRatingsTestDB creates an in-memory database with draft_card_ratings table.
func setupDraftRatingsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS draft_card_ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			set_code TEXT NOT NULL,
			draft_format TEXT NOT NULL,
			arena_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			color TEXT,
			rarity TEXT,
			gihwr REAL,
			ohwr REAL,
			alsa REAL,
			ata REAL,
			gih_count INTEGER,
			data_source TEXT,
			cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(set_code, draft_format, arena_id)
		);

		CREATE TABLE IF NOT EXISTS draft_color_ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			set_code TEXT NOT NULL,
			draft_format TEXT NOT NULL,
			color_combination TEXT NOT NULL,
			win_rate REAL,
			games_played INTEGER,
			cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(set_code, draft_format, color_combination)
		);

		CREATE INDEX IF NOT EXISTS idx_draft_card_ratings_set ON draft_card_ratings(set_code);
		CREATE INDEX IF NOT EXISTS idx_draft_card_ratings_format ON draft_card_ratings(draft_format);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDraftRatingsRepository_GetStatisticsStaleness(t *testing.T) {
	db := setupDraftRatingsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRatingsRepository(db)
	ctx := context.Background()

	// Insert test ratings with varying freshness
	now := time.Now()
	freshTime := now.Add(-1 * time.Hour)                   // 1 hour ago - fresh
	staleTime := now.Add(-3 * 24 * time.Hour)              // 3 days ago - stale
	staleAgeSeconds := int((1 * 24 * time.Hour).Seconds()) // 1 day

	// Insert fresh rating
	freshRatings := []seventeenlands.CardRating{
		{MTGAID: 12345, Name: "Fresh Card", Color: "W", Rarity: "common", GIHWR: 55.0, OHWR: 53.0, ALSA: 5.5, ATA: 4.0, GIH: 1000},
	}

	// Override cached_at by direct SQL insert
	_, err := db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count, cached_at)
		VALUES ('ONE', 'QuickDraft', 12345, 'Fresh Card', 'W', 'common', 55.0, 53.0, 5.5, 4.0, 1000, ?)
	`, freshTime)
	if err != nil {
		t.Fatalf("failed to insert fresh rating: %v", err)
	}

	// Insert stale rating
	_, err = db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count, cached_at)
		VALUES ('ONE', 'QuickDraft', 12346, 'Stale Card', 'U', 'uncommon', 52.0, 50.0, 6.5, 5.0, 500, ?)
	`, staleTime)
	if err != nil {
		t.Fatalf("failed to insert stale rating: %v", err)
	}

	// Get staleness
	staleness, err := repo.GetStatisticsStaleness(ctx, staleAgeSeconds)
	if err != nil {
		t.Fatalf("failed to get statistics staleness: %v", err)
	}

	if staleness.Total != 2 {
		t.Errorf("expected total 2, got %d", staleness.Total)
	}

	if staleness.Fresh != 1 {
		t.Errorf("expected fresh 1, got %d", staleness.Fresh)
	}

	if staleness.Stale != 1 {
		t.Errorf("expected stale 1, got %d", staleness.Stale)
	}

	// Discard the freshRatings variable warning
	_ = freshRatings
}

func TestDraftRatingsRepository_GetStaleSets(t *testing.T) {
	db := setupDraftRatingsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRatingsRepository(db)
	ctx := context.Background()

	now := time.Now()
	freshTime := now.Add(-1 * time.Hour)                   // 1 hour ago - fresh
	staleTime := now.Add(-3 * 24 * time.Hour)              // 3 days ago - stale
	staleAgeSeconds := int((1 * 24 * time.Hour).Seconds()) // 1 day

	// Insert fresh set (ONE)
	_, err := db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, cached_at)
		VALUES ('ONE', 'QuickDraft', 12345, 'Fresh Card', 'W', 'common', 55.0, ?)
	`, freshTime)
	if err != nil {
		t.Fatalf("failed to insert fresh rating: %v", err)
	}

	// Insert stale set (MOM)
	_, err = db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, cached_at)
		VALUES ('MOM', 'QuickDraft', 12346, 'Stale Card 1', 'U', 'uncommon', 52.0, ?)
	`, staleTime)
	if err != nil {
		t.Fatalf("failed to insert stale rating 1: %v", err)
	}

	// Insert another stale set (LCI)
	_, err = db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, cached_at)
		VALUES ('LCI', 'QuickDraft', 12347, 'Stale Card 2', 'B', 'rare', 48.0, ?)
	`, staleTime)
	if err != nil {
		t.Fatalf("failed to insert stale rating 2: %v", err)
	}

	// Get stale sets
	staleSets, err := repo.GetStaleSets(ctx, staleAgeSeconds)
	if err != nil {
		t.Fatalf("failed to get stale sets: %v", err)
	}

	if len(staleSets) != 2 {
		t.Errorf("expected 2 stale sets, got %d", len(staleSets))
	}

	// Should not include fresh set
	for _, set := range staleSets {
		if set == "ONE" {
			t.Errorf("fresh set ONE should not be in stale sets")
		}
	}
}

func TestDraftRatingsRepository_GetStaleStats(t *testing.T) {
	db := setupDraftRatingsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRatingsRepository(db)
	ctx := context.Background()

	now := time.Now()
	freshTime := now.Add(-1 * time.Hour)                   // 1 hour ago - fresh
	staleTime1 := now.Add(-5 * 24 * time.Hour)             // 5 days ago - stalest
	staleTime2 := now.Add(-3 * 24 * time.Hour)             // 3 days ago - stale
	staleAgeSeconds := int((1 * 24 * time.Hour).Seconds()) // 1 day

	// Insert fresh rating
	_, err := db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, cached_at)
		VALUES ('ONE', 'QuickDraft', 12345, 'Fresh Card', 'W', 'common', 55.0, ?)
	`, freshTime)
	if err != nil {
		t.Fatalf("failed to insert fresh rating: %v", err)
	}

	// Insert stalest rating (should come first)
	_, err = db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, cached_at)
		VALUES ('MOM', 'PremierDraft', 12346, 'Stalest Card', 'U', 'uncommon', 52.0, ?)
	`, staleTime1)
	if err != nil {
		t.Fatalf("failed to insert stalest rating: %v", err)
	}

	// Insert stale rating
	_, err = db.ExecContext(ctx, `
		INSERT INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, cached_at)
		VALUES ('LCI', 'QuickDraft', 12347, 'Stale Card', 'B', 'rare', 48.0, ?)
	`, staleTime2)
	if err != nil {
		t.Fatalf("failed to insert stale rating: %v", err)
	}

	// Get stale stats
	staleStats, err := repo.GetStaleStats(ctx, staleAgeSeconds)
	if err != nil {
		t.Fatalf("failed to get stale stats: %v", err)
	}

	if len(staleStats) != 2 {
		t.Errorf("expected 2 stale stats, got %d", len(staleStats))
	}

	// First should be the stalest (MOM PremierDraft)
	if len(staleStats) > 0 {
		if staleStats[0].SetCode != "MOM" {
			t.Errorf("expected oldest set MOM, got %s", staleStats[0].SetCode)
		}
		if staleStats[0].Format != "PremierDraft" {
			t.Errorf("expected format PremierDraft, got %s", staleStats[0].Format)
		}
	}
}

func TestDraftRatingsRepository_GetStatisticsStaleness_EmptyDB(t *testing.T) {
	db := setupDraftRatingsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRatingsRepository(db)
	ctx := context.Background()

	staleness, err := repo.GetStatisticsStaleness(ctx, 86400)
	if err != nil {
		t.Fatalf("failed to get statistics staleness from empty DB: %v", err)
	}

	if staleness.Total != 0 {
		t.Errorf("expected total 0, got %d", staleness.Total)
	}

	if len(staleness.StaleSets) != 0 {
		t.Errorf("expected 0 stale sets, got %d", len(staleness.StaleSets))
	}
}

func TestDraftRatingsRepository_GetStaleStats_EmptyDB(t *testing.T) {
	db := setupDraftRatingsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRatingsRepository(db)
	ctx := context.Background()

	staleStats, err := repo.GetStaleStats(ctx, 86400)
	if err != nil {
		t.Fatalf("failed to get stale stats from empty DB: %v", err)
	}

	if len(staleStats) != 0 {
		t.Errorf("expected 0 stale stats, got %d", len(staleStats))
	}
}

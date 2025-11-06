package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupStatsTestDB creates an in-memory database with player_stats table.
func setupStatsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE player_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATE NOT NULL,
			format TEXT NOT NULL,
			matches_played INTEGER DEFAULT 0,
			matches_won INTEGER DEFAULT 0,
			games_played INTEGER DEFAULT 0,
			games_won INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(date, format)
		);

		CREATE INDEX idx_player_stats_date ON player_stats(date);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestStatsRepository_Upsert(t *testing.T) {
	db := setupStatsTestDB(t)
	defer db.Close()

	repo := NewStatsRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(24 * time.Hour)

	stats := &models.PlayerStats{
		Date:          now,
		Format:        "Standard",
		MatchesPlayed: 10,
		MatchesWon:    6,
		GamesPlayed:   25,
		GamesWon:      15,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert
	err := repo.Upsert(ctx, stats)
	if err != nil {
		t.Fatalf("failed to upsert stats: %v", err)
	}

	if stats.ID == 0 {
		t.Error("expected ID to be set after insert")
	}

	// Verify insert
	retrieved, err := repo.GetByDate(ctx, now, "Standard")
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected stats to be found")
	}

	if retrieved.MatchesPlayed != 10 {
		t.Errorf("expected 10 matches played, got %d", retrieved.MatchesPlayed)
	}

	// Update
	stats.MatchesPlayed = 15
	stats.MatchesWon = 9
	stats.UpdatedAt = time.Now()

	err = repo.Upsert(ctx, stats)
	if err != nil {
		t.Fatalf("failed to upsert stats: %v", err)
	}

	// Verify update
	retrieved, err = repo.GetByDate(ctx, now, "Standard")
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if retrieved.MatchesPlayed != 15 {
		t.Errorf("expected 15 matches played after update, got %d", retrieved.MatchesPlayed)
	}

	if retrieved.MatchesWon != 9 {
		t.Errorf("expected 9 matches won after update, got %d", retrieved.MatchesWon)
	}
}

func TestStatsRepository_GetByDate(t *testing.T) {
	db := setupStatsTestDB(t)
	defer db.Close()

	repo := NewStatsRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(24 * time.Hour)

	// Test getting non-existent stats
	stats, err := repo.GetByDate(ctx, now, "Standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats != nil {
		t.Error("expected nil stats for non-existent record")
	}

	// Create and retrieve
	newStats := &models.PlayerStats{
		Date:          now,
		Format:        "Standard",
		MatchesPlayed: 5,
		MatchesWon:    3,
		GamesPlayed:   12,
		GamesWon:      7,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = repo.Upsert(ctx, newStats)
	if err != nil {
		t.Fatalf("failed to upsert stats: %v", err)
	}

	stats, err = repo.GetByDate(ctx, now, "Standard")
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats == nil {
		t.Fatal("expected stats to be found")
	}

	if stats.MatchesWon != 3 {
		t.Errorf("expected 3 matches won, got %d", stats.MatchesWon)
	}
}

func TestStatsRepository_GetByDateRange(t *testing.T) {
	db := setupStatsTestDB(t)
	defer db.Close()

	repo := NewStatsRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	// Create stats for multiple days
	statsData := []*models.PlayerStats{
		{
			Date:          twoDaysAgo,
			Format:        "Standard",
			MatchesPlayed: 5,
			MatchesWon:    2,
			GamesPlayed:   10,
			GamesWon:      5,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			Date:          yesterday,
			Format:        "Standard",
			MatchesPlayed: 8,
			MatchesWon:    5,
			GamesPlayed:   18,
			GamesWon:      11,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			Date:          now,
			Format:        "Standard",
			MatchesPlayed: 6,
			MatchesWon:    4,
			GamesPlayed:   14,
			GamesWon:      9,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, s := range statsData {
		if err := repo.Upsert(ctx, s); err != nil {
			t.Fatalf("failed to upsert stats: %v", err)
		}
	}

	// Query for all three days
	results, err := repo.GetByDateRange(ctx, twoDaysAgo, now, "Standard")
	if err != nil {
		t.Fatalf("failed to get stats by date range: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 stats records, got %d", len(results))
	}

	// Query for last two days
	results, err = repo.GetByDateRange(ctx, yesterday, now, "Standard")
	if err != nil {
		t.Fatalf("failed to get stats by date range: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 stats records, got %d", len(results))
	}

	// Verify descending order
	if !results[0].Date.After(results[1].Date) && !results[0].Date.Equal(results[1].Date) {
		t.Error("expected results in descending date order")
	}
}

func TestStatsRepository_GetAllFormats(t *testing.T) {
	db := setupStatsTestDB(t)
	defer db.Close()

	repo := NewStatsRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(24 * time.Hour)

	// Create stats for multiple formats
	statsData := []*models.PlayerStats{
		{
			Date:          now,
			Format:        "Standard",
			MatchesPlayed: 5,
			MatchesWon:    3,
			GamesPlayed:   12,
			GamesWon:      7,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			Date:          now,
			Format:        "Historic",
			MatchesPlayed: 8,
			MatchesWon:    4,
			GamesPlayed:   18,
			GamesWon:      10,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			Date:          now,
			Format:        "Limited",
			MatchesPlayed: 3,
			MatchesWon:    2,
			GamesPlayed:   7,
			GamesWon:      5,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, s := range statsData {
		if err := repo.Upsert(ctx, s); err != nil {
			t.Fatalf("failed to upsert stats: %v", err)
		}
	}

	// Get all formats for today
	results, err := repo.GetAllFormats(ctx, now)
	if err != nil {
		t.Fatalf("failed to get stats for all formats: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 stats records, got %d", len(results))
	}

	// Verify formats are present
	formats := make(map[string]bool)
	for _, s := range results {
		formats[s.Format] = true
	}

	if !formats["Standard"] || !formats["Historic"] || !formats["Limited"] {
		t.Error("expected all three formats to be present")
	}
}

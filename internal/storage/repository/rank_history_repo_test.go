package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupRankHistoryTestDB creates an in-memory database with rank_history table.
func setupRankHistoryTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS rank_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			format TEXT NOT NULL,
			season_ordinal INTEGER NOT NULL,
			rank_class TEXT,
			rank_level INTEGER,
			rank_step INTEGER,
			percentile REAL,
			created_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_rank_history_account_format ON rank_history(account_id, format);
		CREATE INDEX IF NOT EXISTS idx_rank_history_timestamp ON rank_history(timestamp);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestRankHistoryRepository_Create(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	rankClass := "Gold"
	rankLevel := 2
	rankStep := 3

	rank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     time.Now(),
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		RankLevel:     &rankLevel,
		RankStep:      &rankStep,
		CreatedAt:     time.Now(),
	}

	err := repo.Create(ctx, rank)
	if err != nil {
		t.Fatalf("failed to create rank history: %v", err)
	}

	if rank.ID == 0 {
		t.Error("expected ID to be set after create")
	}
}

func TestRankHistoryRepository_GetByFormat(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	// Create rank entries for different formats
	now := time.Now()
	rankClass := "Gold"
	rankLevel := 2

	constructedRank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		RankLevel:     &rankLevel,
		CreatedAt:     now,
	}

	limitedRank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now,
		Format:        "limited",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		RankLevel:     &rankLevel,
		CreatedAt:     now,
	}

	if err := repo.Create(ctx, constructedRank); err != nil {
		t.Fatalf("failed to create constructed rank: %v", err)
	}
	if err := repo.Create(ctx, limitedRank); err != nil {
		t.Fatalf("failed to create limited rank: %v", err)
	}

	// Get only constructed ranks
	results, err := repo.GetByFormat(ctx, 1, "constructed")
	if err != nil {
		t.Fatalf("failed to get ranks by format: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 constructed rank, got %d", len(results))
	}

	if results[0].Format != "constructed" {
		t.Errorf("expected format 'constructed', got '%s'", results[0].Format)
	}
}

func TestRankHistoryRepository_GetBySeason(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	now := time.Now()
	rankClass := "Silver"

	// Create ranks for different seasons
	season15 := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		CreatedAt:     now,
	}

	season16 := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now.Add(time.Hour),
		Format:        "constructed",
		SeasonOrdinal: 16,
		RankClass:     &rankClass,
		CreatedAt:     now.Add(time.Hour),
	}

	if err := repo.Create(ctx, season15); err != nil {
		t.Fatalf("failed to create season 15 rank: %v", err)
	}
	if err := repo.Create(ctx, season16); err != nil {
		t.Fatalf("failed to create season 16 rank: %v", err)
	}

	// Get only season 15 ranks
	results, err := repo.GetBySeason(ctx, 1, 15)
	if err != nil {
		t.Fatalf("failed to get ranks by season: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 rank for season 15, got %d", len(results))
	}

	if results[0].SeasonOrdinal != 15 {
		t.Errorf("expected season 15, got %d", results[0].SeasonOrdinal)
	}
}

func TestRankHistoryRepository_GetByDateRange(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)
	rankClass := "Platinum"

	// Create ranks at different times
	oldRank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     twoDaysAgo,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		CreatedAt:     twoDaysAgo,
	}

	recentRank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		CreatedAt:     now,
	}

	if err := repo.Create(ctx, oldRank); err != nil {
		t.Fatalf("failed to create old rank: %v", err)
	}
	if err := repo.Create(ctx, recentRank); err != nil {
		t.Fatalf("failed to create recent rank: %v", err)
	}

	// Get ranks from last 24 hours only
	results, err := repo.GetByDateRange(ctx, 1, yesterday, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("failed to get ranks by date range: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 rank in date range, got %d", len(results))
	}
}

func TestRankHistoryRepository_GetLatestByFormat(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	now := time.Now()
	earlier := now.Add(-time.Hour)
	silverClass := "Silver"
	goldClass := "Gold"

	// Create ranks at different times
	earlierRank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     earlier,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &silverClass,
		CreatedAt:     earlier,
	}

	laterRank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &goldClass,
		CreatedAt:     now,
	}

	if err := repo.Create(ctx, earlierRank); err != nil {
		t.Fatalf("failed to create earlier rank: %v", err)
	}
	if err := repo.Create(ctx, laterRank); err != nil {
		t.Fatalf("failed to create later rank: %v", err)
	}

	// Get latest should return the gold rank
	result, err := repo.GetLatestByFormat(ctx, 1, "constructed")
	if err != nil {
		t.Fatalf("failed to get latest rank: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if *result.RankClass != "Gold" {
		t.Errorf("expected Gold rank (latest), got %s", *result.RankClass)
	}
}

func TestRankHistoryRepository_GetLatestByFormat_NotFound(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	// Get latest from empty database
	result, err := repo.GetLatestByFormat(ctx, 1, "constructed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Error("expected nil result for empty database")
	}
}

func TestRankHistoryRepository_GetAll(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	now := time.Now()
	rankClass := "Diamond"

	// Create multiple ranks
	for i := 0; i < 3; i++ {
		rank := &models.RankHistory{
			AccountID:     1,
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			Format:        "constructed",
			SeasonOrdinal: 15,
			RankClass:     &rankClass,
			CreatedAt:     now.Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, rank); err != nil {
			t.Fatalf("failed to create rank %d: %v", i, err)
		}
	}

	// Also create one for a different account
	otherAccountRank := &models.RankHistory{
		AccountID:     2,
		Timestamp:     now,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     &rankClass,
		CreatedAt:     now,
	}
	if err := repo.Create(ctx, otherAccountRank); err != nil {
		t.Fatalf("failed to create other account rank: %v", err)
	}

	// Get all for account 1
	results, err := repo.GetAll(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get all ranks: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 ranks for account 1, got %d", len(results))
	}

	// Verify they're ordered by timestamp DESC
	for i := 1; i < len(results); i++ {
		if results[i-1].Timestamp.Before(results[i].Timestamp) {
			t.Error("results should be ordered by timestamp DESC")
		}
	}
}

func TestRankHistoryRepository_GetAll_EmptyDB(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	results, err := repo.GetAll(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRankHistoryRepository_NullableFields(t *testing.T) {
	db := setupRankHistoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewRankHistoryRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create rank with all nullable fields as nil
	rank := &models.RankHistory{
		AccountID:     1,
		Timestamp:     now,
		Format:        "constructed",
		SeasonOrdinal: 15,
		RankClass:     nil,
		RankLevel:     nil,
		RankStep:      nil,
		Percentile:    nil,
		CreatedAt:     now,
	}

	err := repo.Create(ctx, rank)
	if err != nil {
		t.Fatalf("failed to create rank with nil fields: %v", err)
	}

	// Retrieve and verify
	result, err := repo.GetLatestByFormat(ctx, 1, "constructed")
	if err != nil {
		t.Fatalf("failed to get rank: %v", err)
	}

	if result.RankClass != nil {
		t.Error("expected RankClass to be nil")
	}
	if result.RankLevel != nil {
		t.Error("expected RankLevel to be nil")
	}
	if result.RankStep != nil {
		t.Error("expected RankStep to be nil")
	}
	if result.Percentile != nil {
		t.Error("expected Percentile to be nil")
	}
}

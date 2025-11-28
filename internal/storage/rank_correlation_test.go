package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// setupRankCorrelationTestDB creates an in-memory database for rank correlation tests.
func setupRankCorrelationTestDB(t *testing.T) *DB {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE accounts (
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

		CREATE TABLE rank_history (
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
		CREATE INDEX idx_rank_history_account_format ON rank_history(account_id, format);

		-- Insert default account
		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, datetime('now'), datetime('now'));
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	db := &DB{conn: sqlDB}
	return db
}

func TestExtractRankSnapshots(t *testing.T) {
	tests := []struct {
		name           string
		entries        []*logreader.LogEntry
		expectedCount  int
		expectedFormat string
	}{
		{
			name: "constructed rank entry",
			entries: []*logreader.LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2024-01-15 10:30:00",
					JSON: map[string]interface{}{
						"constructedSeasonOrdinal": float64(15),
						"constructedClass":         "Gold",
						"constructedLevel":         float64(2),
						"constructedStep":          float64(3),
					},
				},
			},
			expectedCount:  1,
			expectedFormat: "constructed",
		},
		{
			name: "limited rank entry",
			entries: []*logreader.LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2024-01-15 10:30:00",
					JSON: map[string]interface{}{
						"limitedSeasonOrdinal": float64(15),
						"limitedClass":         "Silver",
						"limitedLevel":         float64(1),
						"limitedStep":          float64(5),
					},
				},
			},
			expectedCount:  1,
			expectedFormat: "limited",
		},
		{
			name: "both formats in one entry",
			entries: []*logreader.LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2024-01-15 10:30:00",
					JSON: map[string]interface{}{
						"constructedSeasonOrdinal": float64(15),
						"constructedClass":         "Gold",
						"limitedSeasonOrdinal":     float64(15),
						"limitedClass":             "Bronze",
					},
				},
			},
			expectedCount: 2, // Both constructed and limited
		},
		{
			name: "non-rank entry",
			entries: []*logreader.LogEntry{
				{
					IsJSON:    true,
					Timestamp: "2024-01-15 10:30:00",
					JSON: map[string]interface{}{
						"someOtherField": "value",
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "non-JSON entry",
			entries: []*logreader.LogEntry{
				{
					IsJSON:    false,
					Timestamp: "2024-01-15 10:30:00",
				},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshots := extractRankSnapshots(tt.entries)

			if len(snapshots) != tt.expectedCount {
				t.Errorf("expected %d snapshots, got %d", tt.expectedCount, len(snapshots))
			}

			if tt.expectedCount > 0 && tt.expectedFormat != "" {
				if snapshots[0].Format != tt.expectedFormat {
					t.Errorf("expected format %s, got %s", tt.expectedFormat, snapshots[0].Format)
				}
			}
		})
	}
}

func TestFormatRankString(t *testing.T) {
	tests := []struct {
		name     string
		snapshot RankSnapshot
		expected string
	}{
		{
			name: "full rank with class, level, and step",
			snapshot: RankSnapshot{
				RankClass: strPtr("Gold"),
				RankLevel: intPtr(2),
				RankStep:  intPtr(3),
			},
			expected: "Gold 2 Step 3",
		},
		{
			name: "rank with class and level only",
			snapshot: RankSnapshot{
				RankClass: strPtr("Silver"),
				RankLevel: intPtr(1),
			},
			expected: "Silver 1",
		},
		{
			name: "rank with class only",
			snapshot: RankSnapshot{
				RankClass: strPtr("Mythic"),
			},
			expected: "Mythic",
		},
		{
			name: "unranked (nil class)",
			snapshot: RankSnapshot{
				RankClass: nil,
			},
			expected: "Unranked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRankString(tt.snapshot)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractRankTier(t *testing.T) {
	tests := []struct {
		rankStr  string
		expected string
	}{
		{"Gold 2 Step 3", "Gold"},
		{"Silver 1", "Silver"},
		{"Mythic", "Mythic"},
		{"Bronze 4 Step 0", "Bronze"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.rankStr, func(t *testing.T) {
			result := extractRankTier(tt.rankStr)
			if result != tt.expected {
				t.Errorf("extractRankTier(%q) = %q, want %q", tt.rankStr, result, tt.expected)
			}
		})
	}
}

func TestStoreRankHistory(t *testing.T) {
	db := setupRankCorrelationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	svc := NewService(db)
	ctx := context.Background()

	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:00",
			JSON: map[string]interface{}{
				"constructedSeasonOrdinal": float64(15),
				"constructedClass":         "Gold",
				"constructedLevel":         float64(2),
				"constructedStep":          float64(3),
			},
		},
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 11:30:00",
			JSON: map[string]interface{}{
				"limitedSeasonOrdinal": float64(15),
				"limitedClass":         "Silver",
				"limitedLevel":         float64(1),
				"limitedStep":          float64(5),
			},
		},
	}

	err := svc.StoreRankHistory(ctx, entries)
	if err != nil {
		t.Fatalf("StoreRankHistory failed: %v", err)
	}

	// Verify constructed rank was stored
	constructedRanks, err := svc.GetRankHistoryByFormat(ctx, "constructed")
	if err != nil {
		t.Fatalf("failed to get constructed ranks: %v", err)
	}

	if len(constructedRanks) != 1 {
		t.Errorf("expected 1 constructed rank, got %d", len(constructedRanks))
	} else {
		if *constructedRanks[0].RankClass != "Gold" {
			t.Errorf("expected Gold rank, got %s", *constructedRanks[0].RankClass)
		}
	}

	// Verify limited rank was stored
	limitedRanks, err := svc.GetRankHistoryByFormat(ctx, "limited")
	if err != nil {
		t.Fatalf("failed to get limited ranks: %v", err)
	}

	if len(limitedRanks) != 1 {
		t.Errorf("expected 1 limited rank, got %d", len(limitedRanks))
	} else {
		if *limitedRanks[0].RankClass != "Silver" {
			t.Errorf("expected Silver rank, got %s", *limitedRanks[0].RankClass)
		}
	}
}

func TestStoreRankHistory_EmptyEntries(t *testing.T) {
	db := setupRankCorrelationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	svc := NewService(db)
	ctx := context.Background()

	// Should not error with empty entries
	err := svc.StoreRankHistory(ctx, []*logreader.LogEntry{})
	if err != nil {
		t.Fatalf("StoreRankHistory with empty entries failed: %v", err)
	}

	// Verify nothing was stored
	allRanks, err := svc.GetAllRankHistory(ctx)
	if err != nil {
		t.Fatalf("failed to get all ranks: %v", err)
	}

	if len(allRanks) != 0 {
		t.Errorf("expected 0 ranks, got %d", len(allRanks))
	}
}

func TestCorrelateRanksWithMatches(t *testing.T) {
	now := time.Now()

	matches := []matchData{
		{
			Match: &Match{
				ID:        "match1",
				Format:    "Standard",
				Timestamp: now,
			},
		},
		{
			Match: &Match{
				ID:        "match2",
				Format:    "Premier Draft",
				Timestamp: now.Add(time.Hour),
			},
		},
	}

	rankSnapshots := []RankSnapshot{
		{
			Timestamp: now.Add(-time.Minute),
			Format:    "constructed",
			RankClass: strPtr("Gold"),
			RankLevel: intPtr(2),
		},
		{
			Timestamp: now.Add(30 * time.Minute),
			Format:    "constructed",
			RankClass: strPtr("Gold"),
			RankLevel: intPtr(1),
		},
		{
			Timestamp: now.Add(-time.Minute),
			Format:    "limited",
			RankClass: strPtr("Silver"),
			RankLevel: intPtr(3),
		},
	}

	correlateRanksWithMatches(matches, rankSnapshots)

	// Standard match should have constructed rank
	if matches[0].Match.RankBefore == nil {
		t.Error("expected RankBefore to be set for Standard match")
	}

	// Premier Draft match should have limited rank
	if matches[1].Match.RankBefore == nil {
		t.Error("expected RankBefore to be set for Premier Draft match")
	}
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

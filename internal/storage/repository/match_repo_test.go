package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory database with the necessary schema for testing.
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create tables (simplified schema for testing)
	schema := `
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			screen_name TEXT,
			client_id TEXT,
			is_default INTEGER NOT NULL DEFAULT 0 CHECK(is_default IN (0, 1)),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		INSERT INTO accounts (name, is_default) VALUES ('Default Account', 1);

		CREATE TABLE matches (
			id TEXT PRIMARY KEY,
			account_id INTEGER NOT NULL,
			event_id TEXT NOT NULL,
			event_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			duration_seconds INTEGER,
			player_wins INTEGER NOT NULL,
			opponent_wins INTEGER NOT NULL,
			player_team_id INTEGER NOT NULL,
			deck_id TEXT,
			rank_before TEXT,
			rank_after TEXT,
			format TEXT NOT NULL,
			result TEXT NOT NULL,
			result_reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		);

		CREATE INDEX idx_matches_timestamp ON matches(timestamp);
		CREATE INDEX idx_matches_format ON matches(format);

		CREATE TABLE games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			match_id TEXT NOT NULL,
			game_number INTEGER NOT NULL,
			result TEXT NOT NULL,
			duration_seconds INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (match_id) REFERENCES matches(id),
			UNIQUE(match_id, game_number)
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestMatchRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	match := &models.Match{
		ID:           "match-1",
		AccountID:    1,
		EventID:      "event-1",
		EventName:    "Standard Ranked",
		Timestamp:    time.Now(),
		PlayerWins:   2,
		OpponentWins: 1,
		PlayerTeamID: 1,
		Format:       "Standard",
		Result:       "win",
		CreatedAt:    time.Now(),
	}

	err := repo.Create(ctx, match)
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, "match-1")
	if err != nil {
		t.Fatalf("failed to retrieve match: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected match to be found")
	}

	if retrieved.ID != match.ID {
		t.Errorf("expected ID %s, got %s", match.ID, retrieved.ID)
	}

	if retrieved.Result != match.Result {
		t.Errorf("expected result %s, got %s", match.Result, retrieved.Result)
	}
}

func TestMatchRepository_CreateGame(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// Create a match first
	match := &models.Match{
		ID:           "match-1",
		AccountID:    1,
		EventID:      "event-1",
		EventName:    "Standard Ranked",
		Timestamp:    time.Now(),
		PlayerWins:   2,
		OpponentWins: 1,
		PlayerTeamID: 1,
		Format:       "Standard",
		Result:       "win",
		CreatedAt:    time.Now(),
	}

	err := repo.Create(ctx, match)
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	// Create a game
	game := &models.Game{
		MatchID:    "match-1",
		GameNumber: 1,
		Result:     "win",
		CreatedAt:  time.Now(),
	}

	err = repo.CreateGame(ctx, game)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	if game.ID == 0 {
		t.Error("expected game ID to be set")
	}

	// Verify it was created
	games, err := repo.GetGamesForMatch(ctx, "match-1")
	if err != nil {
		t.Fatalf("failed to get games: %v", err)
	}

	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}

	if games[0].Result != "win" {
		t.Errorf("expected result 'win', got '%s'", games[0].Result)
	}
}

func TestMatchRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// Test getting non-existent match
	match, err := repo.GetByID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match != nil {
		t.Error("expected nil match for nonexistent ID")
	}

	// Create and retrieve
	newMatch := &models.Match{
		ID:           "match-1",
		EventID:      "event-1",
		EventName:    "Standard Ranked",
		Timestamp:    time.Now(),
		PlayerWins:   2,
		OpponentWins: 1,
		PlayerTeamID: 1,
		Format:       "Standard",
		Result:       "win",
		CreatedAt:    time.Now(),
	}

	err = repo.Create(ctx, newMatch)
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	match, err = repo.GetByID(ctx, "match-1")
	if err != nil {
		t.Fatalf("failed to get match: %v", err)
	}

	if match == nil {
		t.Fatal("expected match to be found")
	}

	if match.EventName != "Standard Ranked" {
		t.Errorf("expected EventName 'Standard Ranked', got '%s'", match.EventName)
	}
}

func TestMatchRepository_GetByDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Create matches at different times
	matches := []*models.Match{
		{
			ID:           "match-1",
			EventID:      "event-1",
			EventName:    "Match 1",
			Timestamp:    yesterday,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    yesterday,
		},
		{
			ID:           "match-2",
			EventID:      "event-2",
			EventName:    "Match 2",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Query for yesterday to tomorrow
	results, err := repo.GetByDateRange(ctx, yesterday.Add(-1*time.Hour), tomorrow, 0)
	if err != nil {
		t.Fatalf("failed to get matches by date range: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 matches, got %d", len(results))
	}

	// Query for only today
	results, err = repo.GetByDateRange(ctx, now.Add(-1*time.Hour), now.Add(1*time.Hour), 0)
	if err != nil {
		t.Fatalf("failed to get matches by date range: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 match, got %d", len(results))
	}
}

func TestMatchRepository_GetByFormat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create matches with different formats
	matches := []*models.Match{
		{
			ID:           "match-1",
			EventID:      "event-1",
			EventName:    "Standard Match",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-2",
			EventID:      "event-2",
			EventName:    "Historic Match",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "win",
			CreatedAt:    now,
		},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Query for Standard format
	results, err := repo.GetByFormat(ctx, "Standard", 0)
	if err != nil {
		t.Fatalf("failed to get matches by format: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 match, got %d", len(results))
	}

	if results[0].Format != "Standard" {
		t.Errorf("expected format 'Standard', got '%s'", results[0].Format)
	}
}

func TestMatchRepository_GetStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create some matches
	matches := []*models.Match{
		{
			ID:           "match-1",
			EventID:      "event-1",
			EventName:    "Match 1",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-2",
			EventID:      "event-2",
			EventName:    "Match 2",
			Timestamp:    now,
			PlayerWins:   1,
			OpponentWins: 2,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "loss",
			CreatedAt:    now,
		},
		{
			ID:           "match-3",
			EventID:      "event-3",
			EventName:    "Match 3",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Create games for matches
	games := []*models.Game{
		{MatchID: "match-1", GameNumber: 1, Result: "win", CreatedAt: now},
		{MatchID: "match-1", GameNumber: 2, Result: "win", CreatedAt: now},
		{MatchID: "match-2", GameNumber: 1, Result: "loss", CreatedAt: now},
		{MatchID: "match-2", GameNumber: 2, Result: "win", CreatedAt: now},
		{MatchID: "match-2", GameNumber: 3, Result: "loss", CreatedAt: now},
		{MatchID: "match-3", GameNumber: 1, Result: "win", CreatedAt: now},
		{MatchID: "match-3", GameNumber: 2, Result: "loss", CreatedAt: now},
		{MatchID: "match-3", GameNumber: 3, Result: "win", CreatedAt: now},
	}

	for _, g := range games {
		if err := repo.CreateGame(ctx, g); err != nil {
			t.Fatalf("failed to create game: %v", err)
		}
	}

	// Get stats without filter
	stats, err := repo.GetStats(ctx, models.StatsFilter{})
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalMatches != 3 {
		t.Errorf("expected 3 total matches, got %d", stats.TotalMatches)
	}

	if stats.MatchesWon != 2 {
		t.Errorf("expected 2 matches won, got %d", stats.MatchesWon)
	}

	if stats.MatchesLost != 1 {
		t.Errorf("expected 1 match lost, got %d", stats.MatchesLost)
	}

	expectedWinRate := 2.0 / 3.0
	if stats.WinRate < expectedWinRate-0.01 || stats.WinRate > expectedWinRate+0.01 {
		t.Errorf("expected win rate ~%.2f, got %.2f", expectedWinRate, stats.WinRate)
	}

	if stats.TotalGames != 8 {
		t.Errorf("expected 8 total games, got %d", stats.TotalGames)
	}

	if stats.GamesWon != 5 {
		t.Errorf("expected 5 games won, got %d", stats.GamesWon)
	}

	// Test with format filter
	format := "Standard"
	stats, err = repo.GetStats(ctx, models.StatsFilter{Format: &format})
	if err != nil {
		t.Fatalf("failed to get filtered stats: %v", err)
	}

	if stats.TotalMatches != 3 {
		t.Errorf("expected 3 matches for Standard, got %d", stats.TotalMatches)
	}
}

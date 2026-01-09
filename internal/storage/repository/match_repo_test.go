package repository

import (
	"context"
	"database/sql"
	"fmt"
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

		CREATE TABLE decks (
			id TEXT PRIMARY KEY,
			account_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			format TEXT NOT NULL,
			source TEXT,
			draft_event_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		);

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
			opponent_name TEXT,
			opponent_id TEXT,
			processed_for_ml BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id),
			FOREIGN KEY (deck_id) REFERENCES decks(id)
		);

		CREATE INDEX idx_matches_timestamp ON matches(timestamp);
		CREATE INDEX idx_matches_format ON matches(format);

		CREATE TABLE games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			match_id TEXT NOT NULL,
			game_number INTEGER NOT NULL,
			result TEXT NOT NULL,
			duration_seconds INTEGER,
			result_reason TEXT,
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

func TestMatchRepository_GetRecentMatches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	threeHoursAgo := now.Add(-3 * time.Hour)

	// Create matches with different timestamps
	matches := []*models.Match{
		{
			ID:           "match-1",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Oldest Match",
			Timestamp:    threeHoursAgo,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    threeHoursAgo,
		},
		{
			ID:           "match-2",
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "Middle Match",
			Timestamp:    twoHoursAgo,
			PlayerWins:   1,
			OpponentWins: 2,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "loss",
			CreatedAt:    twoHoursAgo,
		},
		{
			ID:           "match-3",
			AccountID:    1,
			EventID:      "event-3",
			EventName:    "Newer Match",
			Timestamp:    oneHourAgo,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    oneHourAgo,
		},
		{
			ID:           "match-4",
			AccountID:    1,
			EventID:      "event-4",
			EventName:    "Newest Match",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Limited",
			Result:       "win",
			CreatedAt:    now,
		},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Test getting recent 2 matches
	results, err := repo.GetRecentMatches(ctx, 2, 0)
	if err != nil {
		t.Fatalf("failed to get recent matches: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 matches, got %d", len(results))
	}

	// Should be ordered by timestamp DESC (newest first)
	if results[0].ID != "match-4" {
		t.Errorf("expected newest match first, got %s", results[0].ID)
	}

	if results[1].ID != "match-3" {
		t.Errorf("expected second newest match, got %s", results[1].ID)
	}

	// Test getting all matches
	results, err = repo.GetRecentMatches(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to get recent matches: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("expected 4 matches, got %d", len(results))
	}

	// Test getting 1 match
	results, err = repo.GetRecentMatches(ctx, 1, 0)
	if err != nil {
		t.Fatalf("failed to get recent match: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 match, got %d", len(results))
	}

	if results[0].ID != "match-4" {
		t.Errorf("expected newest match, got %s", results[0].ID)
	}
}

func TestMatchRepository_GetStatsByFormat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create matches with different formats
	matches := []*models.Match{
		// Standard matches (2 wins, 1 loss)
		{
			ID:           "match-1",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Standard Match 1",
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
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "Standard Match 2",
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
			AccountID:    1,
			EventID:      "event-3",
			EventName:    "Standard Match 3",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		// Historic matches (1 win, 1 loss)
		{
			ID:           "match-4",
			AccountID:    1,
			EventID:      "event-4",
			EventName:    "Historic Match 1",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-5",
			AccountID:    1,
			EventID:      "event-5",
			EventName:    "Historic Match 2",
			Timestamp:    now,
			PlayerWins:   0,
			OpponentWins: 2,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "loss",
			CreatedAt:    now,
		},
		// Limited matches (1 win)
		{
			ID:           "match-6",
			AccountID:    1,
			EventID:      "event-6",
			EventName:    "Limited Match 1",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Limited",
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
		// Standard match games
		{MatchID: "match-1", GameNumber: 1, Result: "win", CreatedAt: now},
		{MatchID: "match-1", GameNumber: 2, Result: "win", CreatedAt: now},
		{MatchID: "match-2", GameNumber: 1, Result: "loss", CreatedAt: now},
		{MatchID: "match-2", GameNumber: 2, Result: "win", CreatedAt: now},
		{MatchID: "match-2", GameNumber: 3, Result: "loss", CreatedAt: now},
		{MatchID: "match-3", GameNumber: 1, Result: "win", CreatedAt: now},
		{MatchID: "match-3", GameNumber: 2, Result: "loss", CreatedAt: now},
		{MatchID: "match-3", GameNumber: 3, Result: "win", CreatedAt: now},
		// Historic match games
		{MatchID: "match-4", GameNumber: 1, Result: "win", CreatedAt: now},
		{MatchID: "match-4", GameNumber: 2, Result: "win", CreatedAt: now},
		{MatchID: "match-5", GameNumber: 1, Result: "loss", CreatedAt: now},
		{MatchID: "match-5", GameNumber: 2, Result: "loss", CreatedAt: now},
		// Limited match games
		{MatchID: "match-6", GameNumber: 1, Result: "win", CreatedAt: now},
		{MatchID: "match-6", GameNumber: 2, Result: "loss", CreatedAt: now},
		{MatchID: "match-6", GameNumber: 3, Result: "win", CreatedAt: now},
	}

	for _, g := range games {
		if err := repo.CreateGame(ctx, g); err != nil {
			t.Fatalf("failed to create game: %v", err)
		}
	}

	// Get stats by format without filter
	statsByFormat, err := repo.GetStatsByFormat(ctx, models.StatsFilter{})
	if err != nil {
		t.Fatalf("failed to get stats by format: %v", err)
	}

	// Should have 3 formats
	if len(statsByFormat) != 3 {
		t.Errorf("expected 3 formats, got %d", len(statsByFormat))
	}

	// Check Standard stats
	standardStats, ok := statsByFormat["Standard"]
	if !ok {
		t.Fatal("Standard stats not found")
	}

	if standardStats.TotalMatches != 3 {
		t.Errorf("expected 3 Standard matches, got %d", standardStats.TotalMatches)
	}

	if standardStats.MatchesWon != 2 {
		t.Errorf("expected 2 Standard wins, got %d", standardStats.MatchesWon)
	}

	if standardStats.MatchesLost != 1 {
		t.Errorf("expected 1 Standard loss, got %d", standardStats.MatchesLost)
	}

	expectedWinRate := 2.0 / 3.0
	if standardStats.WinRate < expectedWinRate-0.01 || standardStats.WinRate > expectedWinRate+0.01 {
		t.Errorf("expected Standard win rate ~%.2f, got %.2f", expectedWinRate, standardStats.WinRate)
	}

	if standardStats.TotalGames != 8 {
		t.Errorf("expected 8 Standard games, got %d", standardStats.TotalGames)
	}

	if standardStats.GamesWon != 5 {
		t.Errorf("expected 5 Standard game wins, got %d", standardStats.GamesWon)
	}

	// Check Historic stats
	historicStats, ok := statsByFormat["Historic"]
	if !ok {
		t.Fatal("Historic stats not found")
	}

	if historicStats.TotalMatches != 2 {
		t.Errorf("expected 2 Historic matches, got %d", historicStats.TotalMatches)
	}

	if historicStats.MatchesWon != 1 {
		t.Errorf("expected 1 Historic win, got %d", historicStats.MatchesWon)
	}

	if historicStats.TotalGames != 4 {
		t.Errorf("expected 4 Historic games, got %d", historicStats.TotalGames)
	}

	// Check Limited stats
	limitedStats, ok := statsByFormat["Limited"]
	if !ok {
		t.Fatal("Limited stats not found")
	}

	if limitedStats.TotalMatches != 1 {
		t.Errorf("expected 1 Limited match, got %d", limitedStats.TotalMatches)
	}

	if limitedStats.MatchesWon != 1 {
		t.Errorf("expected 1 Limited win, got %d", limitedStats.MatchesWon)
	}

	if limitedStats.TotalGames != 3 {
		t.Errorf("expected 3 Limited games, got %d", limitedStats.TotalGames)
	}
}

func TestMatchRepository_GetMatches_WithDeckFormat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Create a deck with Standard format
	_, err := db.ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "deck-standard-1", 1, "Standard Deck", "Standard", "constructed", now, now)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Create a match linked to the Standard deck with queue type "Ladder"
	match := &models.Match{
		ID:           "match-with-deck",
		AccountID:    1,
		EventID:      "event-1",
		EventName:    "Ladder",
		Timestamp:    now,
		PlayerWins:   2,
		OpponentWins: 1,
		PlayerTeamID: 1,
		Format:       "Ladder", // Queue type from MTGA
		Result:       "win",
		CreatedAt:    now,
	}
	deckID := "deck-standard-1"
	match.DeckID = &deckID

	if err := repo.Create(ctx, match); err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	// Get matches - should include DeckFormat from JOIN
	matches, err := repo.GetMatches(ctx, models.StatsFilter{})
	if err != nil {
		t.Fatalf("failed to get matches: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	// Verify DeckFormat is populated from the deck's format
	if matches[0].DeckFormat == nil {
		t.Fatal("expected DeckFormat to be populated from deck JOIN")
	}

	if *matches[0].DeckFormat != "Standard" {
		t.Errorf("expected DeckFormat 'Standard', got '%s'", *matches[0].DeckFormat)
	}

	// Verify the queue type Format is still preserved
	if matches[0].Format != "Ladder" {
		t.Errorf("expected Format 'Ladder', got '%s'", matches[0].Format)
	}
}

func TestMatchRepository_GetMatches_WithoutDeck(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Create a match without a deck (draft match scenario)
	match := &models.Match{
		ID:           "match-no-deck",
		AccountID:    1,
		EventID:      "event-draft",
		EventName:    "QuickDraft_TLA_20251127",
		Timestamp:    now,
		PlayerWins:   3,
		OpponentWins: 2,
		PlayerTeamID: 1,
		Format:       "QuickDraft_TLA_20251127",
		Result:       "win",
		CreatedAt:    now,
	}

	if err := repo.Create(ctx, match); err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	// Get matches - DeckFormat should be nil for matches without deck
	matches, err := repo.GetMatches(ctx, models.StatsFilter{})
	if err != nil {
		t.Fatalf("failed to get matches: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	// Verify DeckFormat is nil when no deck is linked
	if matches[0].DeckFormat != nil {
		t.Errorf("expected DeckFormat to be nil for match without deck, got '%s'", *matches[0].DeckFormat)
	}

	// Format should still have the raw queue type
	if matches[0].Format != "QuickDraft_TLA_20251127" {
		t.Errorf("expected Format 'QuickDraft_TLA_20251127', got '%s'", matches[0].Format)
	}
}

func TestMatchRepository_GetMatches_FilterByDeckFormat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Create decks with different formats
	_, err := db.ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "deck-standard", 1, "Standard Deck", "Standard", "constructed", now, now)
	if err != nil {
		t.Fatalf("failed to create standard deck: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "deck-historic", 1, "Historic Deck", "Historic", "constructed", now, now)
	if err != nil {
		t.Fatalf("failed to create historic deck: %v", err)
	}

	// Create matches with different deck formats
	standardDeckID := "deck-standard"
	historicDeckID := "deck-historic"

	matches := []*models.Match{
		{
			ID:           "match-standard-1",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Ladder",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			DeckID:       &standardDeckID,
			Format:       "Ladder",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-standard-2",
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "Play",
			Timestamp:    now,
			PlayerWins:   1,
			OpponentWins: 2,
			PlayerTeamID: 1,
			DeckID:       &standardDeckID,
			Format:       "Play",
			Result:       "loss",
			CreatedAt:    now,
		},
		{
			ID:           "match-historic-1",
			AccountID:    1,
			EventID:      "event-3",
			EventName:    "Ladder",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			DeckID:       &historicDeckID,
			Format:       "Ladder",
			Result:       "win",
			CreatedAt:    now,
		},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Filter by DeckFormat = Standard
	deckFormat := "Standard"
	results, err := repo.GetMatches(ctx, models.StatsFilter{DeckFormat: &deckFormat})
	if err != nil {
		t.Fatalf("failed to get matches with deck format filter: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 Standard matches, got %d", len(results))
	}

	for _, m := range results {
		if m.DeckFormat == nil || *m.DeckFormat != "Standard" {
			t.Errorf("expected DeckFormat 'Standard', got '%v'", m.DeckFormat)
		}
	}

	// Filter by DeckFormat = Historic
	deckFormat = "Historic"
	results, err = repo.GetMatches(ctx, models.StatsFilter{DeckFormat: &deckFormat})
	if err != nil {
		t.Fatalf("failed to get matches with deck format filter: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 Historic match, got %d", len(results))
	}

	if results[0].DeckFormat == nil || *results[0].DeckFormat != "Historic" {
		t.Errorf("expected DeckFormat 'Historic', got '%v'", results[0].DeckFormat)
	}
}

func TestMatchRepository_GetMatches_MixedDeckAndNoDeck(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Create a deck
	_, err := db.ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "deck-alchemy", 1, "Alchemy Deck", "Alchemy", "constructed", now, now)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	alchemyDeckID := "deck-alchemy"

	// Create matches - some with deck, some without
	matches := []*models.Match{
		{
			ID:           "match-alchemy",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Alchemy_Ladder",
			Timestamp:    now,
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			DeckID:       &alchemyDeckID,
			Format:       "Alchemy_Ladder",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-draft",
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "PremierDraft_MKM_20241120",
			Timestamp:    now.Add(-1 * time.Hour),
			PlayerWins:   3,
			OpponentWins: 0,
			PlayerTeamID: 1,
			DeckID:       nil,
			Format:       "PremierDraft_MKM_20241120",
			Result:       "win",
			CreatedAt:    now.Add(-1 * time.Hour),
		},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Get all matches
	results, err := repo.GetMatches(ctx, models.StatsFilter{})
	if err != nil {
		t.Fatalf("failed to get matches: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}

	// First match (most recent) should have DeckFormat = Alchemy
	if results[0].DeckFormat == nil {
		t.Error("expected first match to have DeckFormat")
	} else if *results[0].DeckFormat != "Alchemy" {
		t.Errorf("expected DeckFormat 'Alchemy', got '%s'", *results[0].DeckFormat)
	}

	// Second match (draft) should have nil DeckFormat
	if results[1].DeckFormat != nil {
		t.Errorf("expected draft match to have nil DeckFormat, got '%s'", *results[1].DeckFormat)
	}
}

func TestMatchRepository_GetDailyWins(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// MTGA resets at 9 AM UTC, so we need to calculate the current "MTGA day"
	now := time.Now().UTC()
	// Calculate the start of the current MTGA day (9 AM UTC)
	mtgaDayStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	if now.Hour() < 9 {
		// If before 9 AM UTC, the MTGA day started yesterday at 9 AM
		mtgaDayStart = mtgaDayStart.Add(-24 * time.Hour)
	}
	previousMtgaDay := mtgaDayStart.Add(-24 * time.Hour)

	// Create matches:
	// - 2 wins in current MTGA day (after 9 AM UTC reset)
	// - 1 loss in current MTGA day
	// - 1 win in previous MTGA day (should not count)
	matches := []*models.Match{
		{
			ID:           "match-today-win-1",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(1 * time.Hour), // 10 AM UTC (1 hour after reset)
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-today-win-2",
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(3 * time.Hour), // 12 PM UTC (3 hours after reset)
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-today-loss",
			AccountID:    1,
			EventID:      "event-3",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(5 * time.Hour), // 2 PM UTC (5 hours after reset)
			PlayerWins:   1,
			OpponentWins: 2,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "loss",
			CreatedAt:    now,
		},
		{
			ID:           "match-yesterday-win",
			AccountID:    1,
			EventID:      "event-4",
			EventName:    "Standard Ranked",
			Timestamp:    previousMtgaDay.Add(6 * time.Hour), // Previous MTGA day, 3 PM UTC
			PlayerWins:   2,
			OpponentWins: 0,
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

	// Get daily wins for all accounts (accountID = 0)
	dailyWins, err := repo.GetDailyWins(ctx, 0)
	if err != nil {
		t.Fatalf("failed to get daily wins: %v", err)
	}

	if dailyWins != 2 {
		t.Errorf("expected 2 daily wins, got %d", dailyWins)
	}

	// Get daily wins for account 1
	dailyWins, err = repo.GetDailyWins(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get daily wins for account: %v", err)
	}

	if dailyWins != 2 {
		t.Errorf("expected 2 daily wins for account 1, got %d", dailyWins)
	}

	// Get daily wins for non-existent account
	dailyWins, err = repo.GetDailyWins(ctx, 999)
	if err != nil {
		t.Fatalf("failed to get daily wins for non-existent account: %v", err)
	}

	if dailyWins != 0 {
		t.Errorf("expected 0 daily wins for non-existent account, got %d", dailyWins)
	}
}

func TestMatchRepository_GetWeeklyWins(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// MTGA week starts on Sunday at 9 AM UTC
	now := time.Now().UTC()
	daysSinceSunday := int(now.Weekday())

	// Calculate the start of the current MTGA week (Sunday 9 AM UTC)
	mtgaWeekStart := time.Date(now.Year(), now.Month(), now.Day()-daysSinceSunday, 9, 0, 0, 0, time.UTC)
	// If it's Sunday and before 9 AM UTC, the MTGA week started last Sunday at 9 AM
	if now.Weekday() == time.Sunday && now.Hour() < 9 {
		mtgaWeekStart = mtgaWeekStart.Add(-7 * 24 * time.Hour)
	}
	lastMtgaWeek := mtgaWeekStart.Add(-7 * 24 * time.Hour) // Start of last MTGA week

	// Create matches:
	// - 3 wins this MTGA week (after Sunday 9 AM UTC)
	// - 1 loss this MTGA week
	// - 2 wins last MTGA week (should not count)
	matches := []*models.Match{
		{
			ID:           "match-thisweek-win-1",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Historic Ranked",
			Timestamp:    mtgaWeekStart.Add(1 * time.Hour), // Sunday 10 AM UTC (1 hour after reset)
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-thisweek-win-2",
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "Historic Ranked",
			Timestamp:    mtgaWeekStart.Add(25 * time.Hour), // Monday 10 AM UTC
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-thisweek-win-3",
			AccountID:    1,
			EventID:      "event-3",
			EventName:    "Historic Ranked",
			Timestamp:    mtgaWeekStart.Add(49 * time.Hour), // Tuesday 10 AM UTC
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-thisweek-loss",
			AccountID:    1,
			EventID:      "event-4",
			EventName:    "Historic Ranked",
			Timestamp:    mtgaWeekStart.Add(3 * time.Hour), // Sunday noon UTC
			PlayerWins:   0,
			OpponentWins: 2,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "loss",
			CreatedAt:    now,
		},
		{
			ID:           "match-lastweek-win-1",
			AccountID:    1,
			EventID:      "event-5",
			EventName:    "Historic Ranked",
			Timestamp:    lastMtgaWeek.Add(25 * time.Hour), // Last Monday 10 AM UTC
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Historic",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-lastweek-win-2",
			AccountID:    1,
			EventID:      "event-6",
			EventName:    "Historic Ranked",
			Timestamp:    lastMtgaWeek.Add(49 * time.Hour), // Last Tuesday 10 AM UTC
			PlayerWins:   2,
			OpponentWins: 0,
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

	// Get weekly wins for all accounts
	weeklyWins, err := repo.GetWeeklyWins(ctx, 0)
	if err != nil {
		t.Fatalf("failed to get weekly wins: %v", err)
	}

	if weeklyWins != 3 {
		t.Errorf("expected 3 weekly wins, got %d", weeklyWins)
	}

	// Get weekly wins for account 1
	weeklyWins, err = repo.GetWeeklyWins(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get weekly wins for account: %v", err)
	}

	if weeklyWins != 3 {
		t.Errorf("expected 3 weekly wins for account 1, got %d", weeklyWins)
	}

	// Get weekly wins for non-existent account
	weeklyWins, err = repo.GetWeeklyWins(ctx, 999)
	if err != nil {
		t.Fatalf("failed to get weekly wins for non-existent account: %v", err)
	}

	if weeklyWins != 0 {
		t.Errorf("expected 0 weekly wins for non-existent account, got %d", weeklyWins)
	}
}

func TestMatchRepository_GetDailyWins_MultipleAccounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create second account
	_, err := db.Exec(`INSERT INTO accounts (id, name, is_default) VALUES (2, 'Second Account', 0)`)
	if err != nil {
		t.Fatalf("failed to create second account: %v", err)
	}

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// MTGA resets at 9 AM UTC
	now := time.Now().UTC()
	mtgaDayStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	if now.Hour() < 9 {
		mtgaDayStart = mtgaDayStart.Add(-24 * time.Hour)
	}

	// Create wins for account 1 (3 wins) and account 2 (2 wins)
	matches := []*models.Match{
		{
			ID:           "match-acc1-win-1",
			AccountID:    1,
			EventID:      "event-1",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(2 * time.Hour), // 11 AM UTC
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-acc1-win-2",
			AccountID:    1,
			EventID:      "event-2",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(4 * time.Hour), // 1 PM UTC
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-acc1-win-3",
			AccountID:    1,
			EventID:      "event-3",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(6 * time.Hour), // 3 PM UTC
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-acc2-win-1",
			AccountID:    2,
			EventID:      "event-4",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(3 * time.Hour), // 12 PM UTC
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		},
		{
			ID:           "match-acc2-win-2",
			AccountID:    2,
			EventID:      "event-5",
			EventName:    "Standard Ranked",
			Timestamp:    mtgaDayStart.Add(5 * time.Hour), // 2 PM UTC
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

	// Get wins for account 1
	winsAcc1, err := repo.GetDailyWins(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get daily wins for account 1: %v", err)
	}
	if winsAcc1 != 3 {
		t.Errorf("expected 3 daily wins for account 1, got %d", winsAcc1)
	}

	// Get wins for account 2
	winsAcc2, err := repo.GetDailyWins(ctx, 2)
	if err != nil {
		t.Fatalf("failed to get daily wins for account 2: %v", err)
	}
	if winsAcc2 != 2 {
		t.Errorf("expected 2 daily wins for account 2, got %d", winsAcc2)
	}

	// Get total wins (all accounts)
	totalWins, err := repo.GetDailyWins(ctx, 0)
	if err != nil {
		t.Fatalf("failed to get total daily wins: %v", err)
	}
	if totalWins != 5 {
		t.Errorf("expected 5 total daily wins, got %d", totalWins)
	}
}

func TestMatchRepository_GetWeeklyWins_CappedAt15(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// MTGA week starts on Sunday at 9 AM UTC
	now := time.Now().UTC()
	daysSinceSunday := int(now.Weekday())
	mtgaWeekStart := time.Date(now.Year(), now.Month(), now.Day()-daysSinceSunday, 9, 0, 0, 0, time.UTC)
	if now.Weekday() == time.Sunday && now.Hour() < 9 {
		mtgaWeekStart = mtgaWeekStart.Add(-7 * 24 * time.Hour)
	}

	// Create 20 wins this week - should be capped at 15
	for i := 0; i < 20; i++ {
		m := &models.Match{
			ID:           fmt.Sprintf("match-win-%d", i),
			AccountID:    1,
			EventID:      fmt.Sprintf("event-%d", i),
			EventName:    "Standard Ranked",
			Timestamp:    mtgaWeekStart.Add(time.Duration(i+1) * time.Hour),
			PlayerWins:   2,
			OpponentWins: 0,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    now,
		}
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Get weekly wins - should be capped at 15 (MTGA max)
	weeklyWins, err := repo.GetWeeklyWins(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get weekly wins: %v", err)
	}

	if weeklyWins != 15 {
		t.Errorf("expected weekly wins to be capped at 15, got %d", weeklyWins)
	}
}

// ============================================================================
// ML Processing Tests (Issue #851)
// ============================================================================

func TestMatchRepository_GetMatchesForMLProcessing(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// Create a deck first (needed for the JOIN)
	_, err := db.Exec(`INSERT INTO decks (id, account_id, name, format) VALUES ('deck-1', 1, 'Test Deck', 'Standard')`)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	now := time.Now()

	// Create 3 matches - 2 unprocessed, 1 processed
	matches := []*models.Match{
		{ID: "match-1", AccountID: 1, EventID: "e1", EventName: "Standard Ranked", Timestamp: now, PlayerWins: 2, OpponentWins: 0, PlayerTeamID: 1, Format: "Standard", Result: "win", CreatedAt: now},
		{ID: "match-2", AccountID: 1, EventID: "e2", EventName: "Standard Ranked", Timestamp: now, PlayerWins: 0, OpponentWins: 2, PlayerTeamID: 1, Format: "Standard", Result: "loss", CreatedAt: now},
		{ID: "match-3", AccountID: 1, EventID: "e3", EventName: "Standard Ranked", Timestamp: now, PlayerWins: 2, OpponentWins: 1, PlayerTeamID: 1, Format: "Standard", Result: "win", CreatedAt: now},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Mark match-3 as processed
	_, err = db.Exec(`UPDATE matches SET processed_for_ml = TRUE WHERE id = 'match-3'`)
	if err != nil {
		t.Fatalf("failed to mark match as processed: %v", err)
	}

	// Get unprocessed matches
	filter := models.StatsFilter{}
	unprocessed, err := repo.GetMatchesForMLProcessing(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get matches for ML processing: %v", err)
	}

	if len(unprocessed) != 2 {
		t.Errorf("expected 2 unprocessed matches, got %d", len(unprocessed))
	}

	// Verify match-3 is not in the results
	for _, m := range unprocessed {
		if m.ID == "match-3" {
			t.Error("match-3 should not be in unprocessed results")
		}
	}
}

func TestMatchRepository_MarkMatchesAsProcessedForML(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create matches
	matches := []*models.Match{
		{ID: "match-1", AccountID: 1, EventID: "e1", EventName: "Test", Timestamp: now, PlayerWins: 2, OpponentWins: 0, PlayerTeamID: 1, Format: "Standard", Result: "win", CreatedAt: now},
		{ID: "match-2", AccountID: 1, EventID: "e2", EventName: "Test", Timestamp: now, PlayerWins: 2, OpponentWins: 1, PlayerTeamID: 1, Format: "Standard", Result: "win", CreatedAt: now},
		{ID: "match-3", AccountID: 1, EventID: "e3", EventName: "Test", Timestamp: now, PlayerWins: 0, OpponentWins: 2, PlayerTeamID: 1, Format: "Standard", Result: "loss", CreatedAt: now},
	}

	for _, m := range matches {
		if err := repo.Create(ctx, m); err != nil {
			t.Fatalf("failed to create match: %v", err)
		}
	}

	// Mark match-1 and match-2 as processed
	err := repo.MarkMatchesAsProcessedForML(ctx, []string{"match-1", "match-2"})
	if err != nil {
		t.Fatalf("failed to mark matches as processed: %v", err)
	}

	// Verify by getting unprocessed matches
	filter := models.StatsFilter{}
	unprocessed, err := repo.GetMatchesForMLProcessing(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get unprocessed matches: %v", err)
	}

	if len(unprocessed) != 1 {
		t.Errorf("expected 1 unprocessed match, got %d", len(unprocessed))
	}

	if len(unprocessed) > 0 && unprocessed[0].ID != "match-3" {
		t.Errorf("expected match-3 to be unprocessed, got %s", unprocessed[0].ID)
	}
}

func TestMatchRepository_MarkMatchesAsProcessedForML_EmptyList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	// Should not error with empty list
	err := repo.MarkMatchesAsProcessedForML(ctx, []string{})
	if err != nil {
		t.Errorf("expected no error with empty list, got: %v", err)
	}
}

func TestMatchRepository_GetMatchesForMLProcessing_IdempotencyCheck(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewMatchRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a match
	match := &models.Match{
		ID: "match-1", AccountID: 1, EventID: "e1", EventName: "Test",
		Timestamp: now, PlayerWins: 2, OpponentWins: 0, PlayerTeamID: 1,
		Format: "Standard", Result: "win", CreatedAt: now,
	}
	if err := repo.Create(ctx, match); err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	filter := models.StatsFilter{}

	// First call - should return 1 match
	unprocessed1, err := repo.GetMatchesForMLProcessing(ctx, filter)
	if err != nil {
		t.Fatalf("failed first call: %v", err)
	}
	if len(unprocessed1) != 1 {
		t.Errorf("expected 1 match on first call, got %d", len(unprocessed1))
	}

	// Mark as processed
	err = repo.MarkMatchesAsProcessedForML(ctx, []string{"match-1"})
	if err != nil {
		t.Fatalf("failed to mark as processed: %v", err)
	}

	// Second call - should return 0 matches (idempotent)
	unprocessed2, err := repo.GetMatchesForMLProcessing(ctx, filter)
	if err != nil {
		t.Fatalf("failed second call: %v", err)
	}
	if len(unprocessed2) != 0 {
		t.Errorf("expected 0 matches on second call (idempotent), got %d", len(unprocessed2))
	}
}

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

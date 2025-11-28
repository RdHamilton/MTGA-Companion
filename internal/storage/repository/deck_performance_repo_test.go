package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupDeckPerformanceTestDB creates an in-memory database with all performance tracking tables.
func setupDeckPerformanceTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE decks (
			id TEXT PRIMARY KEY,
			account_id INTEGER NOT NULL DEFAULT 1,
			name TEXT NOT NULL,
			format TEXT NOT NULL,
			description TEXT,
			color_identity TEXT,
			source TEXT NOT NULL DEFAULT 'constructed',
			draft_event_id TEXT,
			matches_played INTEGER NOT NULL DEFAULT 0,
			matches_won INTEGER NOT NULL DEFAULT 0,
			games_played INTEGER NOT NULL DEFAULT 0,
			games_won INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			modified_at DATETIME NOT NULL,
			last_played DATETIME
		);

		CREATE TABLE matches (
			id TEXT PRIMARY KEY,
			account_id INTEGER,
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
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE deck_performance_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			deck_id TEXT NOT NULL,
			match_id TEXT NOT NULL,
			archetype TEXT,
			secondary_archetype TEXT,
			archetype_confidence REAL,
			color_identity TEXT NOT NULL,
			card_count INTEGER NOT NULL,
			result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
			games_won INTEGER NOT NULL,
			games_lost INTEGER NOT NULL,
			duration_seconds INTEGER,
			format TEXT NOT NULL,
			event_type TEXT,
			opponent_archetype TEXT,
			rank_tier TEXT,
			match_timestamp DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
			FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE
		);

		CREATE TABLE deck_archetypes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			set_code TEXT,
			format TEXT NOT NULL,
			color_identity TEXT NOT NULL,
			signature_cards TEXT,
			synergy_patterns TEXT,
			total_matches INTEGER NOT NULL DEFAULT 0,
			total_wins INTEGER NOT NULL DEFAULT 0,
			avg_win_rate REAL,
			source TEXT NOT NULL DEFAULT 'system',
			external_id TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, set_code, format)
		);

		CREATE TABLE archetype_card_weights (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			archetype_id INTEGER NOT NULL,
			card_id INTEGER NOT NULL,
			weight REAL NOT NULL DEFAULT 1.0,
			is_signature INTEGER NOT NULL DEFAULT 0,
			source TEXT NOT NULL DEFAULT 'system',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (archetype_id) REFERENCES deck_archetypes(id) ON DELETE CASCADE,
			UNIQUE(archetype_id, card_id)
		);

		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

		INSERT INTO decks (id, account_id, name, format, color_identity, source, created_at, modified_at)
		VALUES ('deck-1', 1, 'Test Deck', 'Draft', 'WU', 'draft', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

		INSERT INTO matches (id, account_id, event_id, event_name, timestamp, player_wins, opponent_wins, player_team_id, format, result)
		VALUES ('match-1', 1, 'event-1', 'Quick Draft', CURRENT_TIMESTAMP, 2, 1, 1, 'Ladder', 'win');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDeckPerformanceRepository_CreateHistory(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	archetype := "UW Flyers"
	confidence := 0.85
	duration := 300

	history := &models.DeckPerformanceHistory{
		AccountID:           1,
		DeckID:              "deck-1",
		MatchID:             "match-1",
		Archetype:           &archetype,
		ArchetypeConfidence: &confidence,
		ColorIdentity:       "WU",
		CardCount:           40,
		Result:              "win",
		GamesWon:            2,
		GamesLost:           1,
		DurationSeconds:     &duration,
		Format:              "Draft",
		MatchTimestamp:      time.Now(),
	}

	err := repo.CreateHistory(ctx, history)
	if err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	if history.ID == 0 {
		t.Error("expected history ID to be set")
	}
}

func TestDeckPerformanceRepository_GetHistoryByDeck(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	// Create test histories
	for i := 0; i < 3; i++ {
		archetype := "UW Flyers"
		history := &models.DeckPerformanceHistory{
			AccountID:      1,
			DeckID:         "deck-1",
			MatchID:        "match-1",
			Archetype:      &archetype,
			ColorIdentity:  "WU",
			CardCount:      40,
			Result:         "win",
			GamesWon:       2,
			GamesLost:      1,
			Format:         "Draft",
			MatchTimestamp: time.Now(),
		}
		if err := repo.CreateHistory(ctx, history); err != nil {
			t.Fatalf("failed to create history: %v", err)
		}
	}

	histories, err := repo.GetHistoryByDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get histories: %v", err)
	}

	if len(histories) != 3 {
		t.Errorf("expected 3 histories, got %d", len(histories))
	}
}

func TestDeckPerformanceRepository_GetHistoryByArchetype(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	// Create histories with different archetypes
	archetypes := []string{"UW Flyers", "UW Flyers", "BR Sacrifice"}
	for i, arch := range archetypes {
		archetype := arch
		history := &models.DeckPerformanceHistory{
			AccountID:      1,
			DeckID:         "deck-1",
			MatchID:        "match-1",
			Archetype:      &archetype,
			ColorIdentity:  "WU",
			CardCount:      40,
			Result:         "win",
			GamesWon:       2,
			GamesLost:      i,
			Format:         "Draft",
			MatchTimestamp: time.Now(),
		}
		if err := repo.CreateHistory(ctx, history); err != nil {
			t.Fatalf("failed to create history: %v", err)
		}
	}

	histories, err := repo.GetHistoryByArchetype(ctx, "UW Flyers", "Draft")
	if err != nil {
		t.Fatalf("failed to get histories: %v", err)
	}

	if len(histories) != 2 {
		t.Errorf("expected 2 histories, got %d", len(histories))
	}
}

func TestDeckPerformanceRepository_GetArchetypePerformance(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	// Create test histories with mixed results
	archetype := "UW Flyers"
	results := []string{"win", "win", "loss", "win"}
	for _, result := range results {
		duration := 300
		history := &models.DeckPerformanceHistory{
			AccountID:       1,
			DeckID:          "deck-1",
			MatchID:         "match-1",
			Archetype:       &archetype,
			ColorIdentity:   "WU",
			CardCount:       40,
			Result:          result,
			GamesWon:        2,
			GamesLost:       1,
			DurationSeconds: &duration,
			Format:          "Draft",
			MatchTimestamp:  time.Now(),
		}
		if err := repo.CreateHistory(ctx, history); err != nil {
			t.Fatalf("failed to create history: %v", err)
		}
	}

	stats, err := repo.GetArchetypePerformance(ctx, "UW Flyers", "Draft")
	if err != nil {
		t.Fatalf("failed to get archetype performance: %v", err)
	}

	if stats.TotalMatches != 4 {
		t.Errorf("expected 4 total matches, got %d", stats.TotalMatches)
	}

	if stats.TotalWins != 3 {
		t.Errorf("expected 3 total wins, got %d", stats.TotalWins)
	}

	expectedWinRate := 0.75
	if stats.WinRate != expectedWinRate {
		t.Errorf("expected win rate %f, got %f", expectedWinRate, stats.WinRate)
	}
}

func TestDeckPerformanceRepository_CreateArchetype(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	err := repo.CreateArchetype(ctx, archetype)
	if err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	if archetype.ID == 0 {
		t.Error("expected archetype ID to be set")
	}
}

func TestDeckPerformanceRepository_GetArchetypeByID(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	if err := repo.CreateArchetype(ctx, archetype); err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	retrieved, err := repo.GetArchetypeByID(ctx, archetype.ID)
	if err != nil {
		t.Fatalf("failed to get archetype: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected archetype, got nil")
	}

	if retrieved.Name != "UW Flyers" {
		t.Errorf("expected name 'UW Flyers', got '%s'", retrieved.Name)
	}

	if *retrieved.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", *retrieved.SetCode)
	}
}

func TestDeckPerformanceRepository_GetArchetypeByName(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	if err := repo.CreateArchetype(ctx, archetype); err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	retrieved, err := repo.GetArchetypeByName(ctx, "UW Flyers", &setCode, "draft")
	if err != nil {
		t.Fatalf("failed to get archetype: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected archetype, got nil")
	}

	if retrieved.ID != archetype.ID {
		t.Errorf("expected ID %d, got %d", archetype.ID, retrieved.ID)
	}
}

func TestDeckPerformanceRepository_ListArchetypes(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	setCode := "FDN"
	archetypes := []struct {
		name   string
		format string
	}{
		{"UW Flyers", "draft"},
		{"BR Sacrifice", "draft"},
		{"Mono Red Aggro", "constructed"},
	}

	for _, a := range archetypes {
		arch := &models.DeckArchetype{
			Name:          a.name,
			SetCode:       &setCode,
			Format:        a.format,
			ColorIdentity: "WU",
			Source:        "system",
		}
		if err := repo.CreateArchetype(ctx, arch); err != nil {
			t.Fatalf("failed to create archetype: %v", err)
		}
	}

	// List all
	all, err := repo.ListArchetypes(ctx, nil, nil)
	if err != nil {
		t.Fatalf("failed to list archetypes: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 archetypes, got %d", len(all))
	}

	// Filter by format
	draftFormat := "draft"
	draftOnly, err := repo.ListArchetypes(ctx, nil, &draftFormat)
	if err != nil {
		t.Fatalf("failed to list archetypes by format: %v", err)
	}
	if len(draftOnly) != 2 {
		t.Errorf("expected 2 draft archetypes, got %d", len(draftOnly))
	}
}

func TestDeckPerformanceRepository_UpdateArchetypeStats(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	if err := repo.CreateArchetype(ctx, archetype); err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	err := repo.UpdateArchetypeStats(ctx, archetype.ID, 100, 60)
	if err != nil {
		t.Fatalf("failed to update stats: %v", err)
	}

	updated, err := repo.GetArchetypeByID(ctx, archetype.ID)
	if err != nil {
		t.Fatalf("failed to get archetype: %v", err)
	}

	if updated.TotalMatches != 100 {
		t.Errorf("expected 100 total matches, got %d", updated.TotalMatches)
	}

	if updated.TotalWins != 60 {
		t.Errorf("expected 60 total wins, got %d", updated.TotalWins)
	}

	if updated.AvgWinRate == nil || *updated.AvgWinRate != 0.6 {
		t.Errorf("expected avg win rate 0.6, got %v", updated.AvgWinRate)
	}
}

func TestDeckPerformanceRepository_CardWeights(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	// Create archetype first
	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	if err := repo.CreateArchetype(ctx, archetype); err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	// Create card weight
	weight := &models.ArchetypeCardWeight{
		ArchetypeID: archetype.ID,
		CardID:      12345,
		Weight:      8.5,
		IsSignature: true,
		Source:      "system",
	}

	err := repo.CreateCardWeight(ctx, weight)
	if err != nil {
		t.Fatalf("failed to create card weight: %v", err)
	}

	if weight.ID == 0 {
		t.Error("expected weight ID to be set")
	}

	// Get weights by archetype
	weights, err := repo.GetCardWeights(ctx, archetype.ID)
	if err != nil {
		t.Fatalf("failed to get card weights: %v", err)
	}

	if len(weights) != 1 {
		t.Errorf("expected 1 weight, got %d", len(weights))
	}

	if weights[0].Weight != 8.5 {
		t.Errorf("expected weight 8.5, got %f", weights[0].Weight)
	}

	if !weights[0].IsSignature {
		t.Error("expected IsSignature to be true")
	}

	// Get weights by card
	cardWeights, err := repo.GetCardWeightsByCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get weights by card: %v", err)
	}

	if len(cardWeights) != 1 {
		t.Errorf("expected 1 weight, got %d", len(cardWeights))
	}
}

func TestDeckPerformanceRepository_UpsertCardWeight(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	// Create archetype first
	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	if err := repo.CreateArchetype(ctx, archetype); err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	// Create initial weight
	weight := &models.ArchetypeCardWeight{
		ArchetypeID: archetype.ID,
		CardID:      12345,
		Weight:      5.0,
		IsSignature: false,
		Source:      "system",
	}

	err := repo.UpsertCardWeight(ctx, weight)
	if err != nil {
		t.Fatalf("failed to upsert card weight: %v", err)
	}

	// Update with new weight
	weight.Weight = 8.5
	weight.IsSignature = true

	err = repo.UpsertCardWeight(ctx, weight)
	if err != nil {
		t.Fatalf("failed to upsert updated card weight: %v", err)
	}

	// Verify only one weight exists with updated values
	weights, err := repo.GetCardWeights(ctx, archetype.ID)
	if err != nil {
		t.Fatalf("failed to get card weights: %v", err)
	}

	if len(weights) != 1 {
		t.Errorf("expected 1 weight after upsert, got %d", len(weights))
	}

	if weights[0].Weight != 8.5 {
		t.Errorf("expected weight 8.5, got %f", weights[0].Weight)
	}

	if !weights[0].IsSignature {
		t.Error("expected IsSignature to be true after upsert")
	}
}

func TestDeckPerformanceRepository_DeleteCardWeight(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	// Create archetype first
	setCode := "FDN"
	archetype := &models.DeckArchetype{
		Name:          "UW Flyers",
		SetCode:       &setCode,
		Format:        "draft",
		ColorIdentity: "WU",
		Source:        "system",
	}

	if err := repo.CreateArchetype(ctx, archetype); err != nil {
		t.Fatalf("failed to create archetype: %v", err)
	}

	// Create card weight
	weight := &models.ArchetypeCardWeight{
		ArchetypeID: archetype.ID,
		CardID:      12345,
		Weight:      8.5,
		IsSignature: true,
		Source:      "system",
	}

	if err := repo.CreateCardWeight(ctx, weight); err != nil {
		t.Fatalf("failed to create card weight: %v", err)
	}

	// Delete weight
	err := repo.DeleteCardWeight(ctx, archetype.ID, 12345)
	if err != nil {
		t.Fatalf("failed to delete card weight: %v", err)
	}

	// Verify deletion
	weights, err := repo.GetCardWeights(ctx, archetype.ID)
	if err != nil {
		t.Fatalf("failed to get card weights: %v", err)
	}

	if len(weights) != 0 {
		t.Errorf("expected 0 weights after deletion, got %d", len(weights))
	}
}

func TestDeckPerformanceRepository_GetPerformanceByDateRange(t *testing.T) {
	db := setupDeckPerformanceTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewDeckPerformanceRepository(db)
	ctx := context.Background()

	now := time.Now()
	archetype := "UW Flyers"

	// Create histories at different times
	times := []time.Time{
		now.Add(-48 * time.Hour), // 2 days ago
		now.Add(-24 * time.Hour), // 1 day ago
		now,                      // now
	}

	for _, ts := range times {
		history := &models.DeckPerformanceHistory{
			AccountID:      1,
			DeckID:         "deck-1",
			MatchID:        "match-1",
			Archetype:      &archetype,
			ColorIdentity:  "WU",
			CardCount:      40,
			Result:         "win",
			GamesWon:       2,
			GamesLost:      1,
			Format:         "Draft",
			MatchTimestamp: ts,
		}
		if err := repo.CreateHistory(ctx, history); err != nil {
			t.Fatalf("failed to create history: %v", err)
		}
	}

	// Query for last 36 hours
	start := now.Add(-36 * time.Hour)
	end := now.Add(time.Hour)

	histories, err := repo.GetPerformanceByDateRange(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("failed to get histories by date range: %v", err)
	}

	if len(histories) != 2 {
		t.Errorf("expected 2 histories in range, got %d", len(histories))
	}
}

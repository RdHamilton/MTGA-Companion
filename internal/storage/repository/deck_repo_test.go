package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupDeckTestDB creates an in-memory database with all deck-related tables.
func setupDeckTestDB(t *testing.T) *sql.DB {
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
			last_played DATETIME,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (draft_event_id) REFERENCES draft_events(id) ON DELETE SET NULL,
			CHECK(source IN ('draft', 'constructed', 'imported'))
		);

		CREATE TABLE deck_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id TEXT NOT NULL,
			card_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			board TEXT NOT NULL,
			from_draft_pick INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
			UNIQUE(deck_id, card_id, board),
			CHECK(from_draft_pick IN (0, 1))
		);

		CREATE TABLE deck_tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
			UNIQUE(deck_id, tag)
		);

		CREATE TABLE draft_events (
			id TEXT PRIMARY KEY,
			account_id INTEGER NOT NULL,
			event_name TEXT NOT NULL,
			set_code TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			wins INTEGER NOT NULL DEFAULT 0,
			losses INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		);

		CREATE TABLE draft_picks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			draft_event_id TEXT NOT NULL,
			pack_number INTEGER NOT NULL,
			pick_number INTEGER NOT NULL,
			selected_card INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (draft_event_id) REFERENCES draft_events(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_deck_cards_deck_id ON deck_cards(deck_id);
		CREATE INDEX idx_deck_tags_deck_id ON deck_tags(deck_id);
		CREATE INDEX idx_draft_picks_event_id ON draft_picks(draft_event_id);

		-- Insert a default test account
		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, '2024-01-01 00:00:00', '2024-01-01 00:00:00');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDeckRepository_Create(t *testing.T) {
	db := setupDeckTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()
	colorIdentity := "UR"

	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Izzet Phoenix",
		Format:        "Historic",
		ColorIdentity: &colorIdentity,
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to retrieve deck: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected deck to be found")
	}

	if retrieved.Name != "Izzet Phoenix" {
		t.Errorf("expected name 'Izzet Phoenix', got '%s'", retrieved.Name)
	}

	if retrieved.Format != "Historic" {
		t.Errorf("expected format 'Historic', got '%s'", retrieved.Format)
	}

	if retrieved.Source != "constructed" {
		t.Errorf("expected source 'constructed', got '%s'", retrieved.Source)
	}

	if retrieved.AccountID != 1 {
		t.Errorf("expected account_id 1, got %d", retrieved.AccountID)
	}
}

func TestDeckRepository_Update(t *testing.T) {
	db := setupDeckTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Original Name",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Update the deck
	deck.Name = "Updated Name"
	deck.ModifiedAt = time.Now()

	err = repo.Update(ctx, deck)
	if err != nil {
		t.Fatalf("failed to update deck: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to retrieve deck: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", retrieved.Name)
	}
}

func TestDeckRepository_GetByID(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	// Test getting non-existent deck
	deck, err := repo.GetByID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deck != nil {
		t.Error("expected nil deck for nonexistent ID")
	}
}

func TestDeckRepository_List(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create multiple decks for account 1
	decks := []*models.Deck{
		{
			ID:            "deck-1",
			AccountID:     1,
			Name:          "Deck 1",
			Format:        "Standard",
			Source:        "constructed",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now,
		},
		{
			ID:            "deck-2",
			AccountID:     1,
			Name:          "Deck 2",
			Format:        "Historic",
			Source:        "imported",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now.Add(1 * time.Hour),
		},
	}

	for _, d := range decks {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("failed to create deck: %v", err)
		}
	}

	// List all decks for account 1
	results, err := repo.List(ctx, 1)
	if err != nil {
		t.Fatalf("failed to list decks: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 decks, got %d", len(results))
	}

	// Should be ordered by modified_at DESC
	if len(results) == 2 && results[0].ID != "deck-2" {
		t.Error("expected decks to be ordered by modified_at DESC")
	}
}

func TestDeckRepository_GetByFormat(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create decks with different formats
	decks := []*models.Deck{
		{
			ID:            "deck-1",
			AccountID:     1,
			Name:          "Standard Deck",
			Format:        "Standard",
			Source:        "constructed",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now,
		},
		{
			ID:            "deck-2",
			AccountID:     1,
			Name:          "Historic Deck",
			Format:        "Historic",
			Source:        "constructed",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now,
		},
	}

	for _, d := range decks {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("failed to create deck: %v", err)
		}
	}

	// Get Standard decks for account 1
	results, err := repo.GetByFormat(ctx, 1, "Standard")
	if err != nil {
		t.Fatalf("failed to get decks by format: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 deck, got %d", len(results))
	}

	if results[0].Format != "Standard" {
		t.Errorf("expected format 'Standard', got '%s'", results[0].Format)
	}
}

func TestDeckRepository_Delete(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Deck to Delete",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Delete the deck
	err = repo.Delete(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to delete deck: %v", err)
	}

	// Verify it was deleted
	retrieved, err := repo.GetByID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get deck: %v", err)
	}

	if retrieved != nil {
		t.Error("expected deck to be deleted")
	}
}

func TestDeckRepository_AddCard(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck first
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Add a card
	card := &models.DeckCard{
		DeckID:        "deck-1",
		CardID:        12345,
		Quantity:      4,
		Board:         "main",
		FromDraftPick: false,
	}

	err = repo.AddCard(ctx, card)
	if err != nil {
		t.Fatalf("failed to add card: %v", err)
	}

	if card.ID == 0 {
		t.Error("expected card ID to be set")
	}

	// Verify it was added
	cards, err := repo.GetCards(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get cards: %v", err)
	}

	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}

	if cards[0].CardID != 12345 {
		t.Errorf("expected card ID 12345, got %d", cards[0].CardID)
	}

	if cards[0].Quantity != 4 {
		t.Errorf("expected quantity 4, got %d", cards[0].Quantity)
	}

	if cards[0].FromDraftPick != false {
		t.Errorf("expected FromDraftPick false, got %v", cards[0].FromDraftPick)
	}

	// Test upsert (update quantity)
	card.Quantity = 3
	err = repo.AddCard(ctx, card)
	if err != nil {
		t.Fatalf("failed to update card: %v", err)
	}

	cards, err = repo.GetCards(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get cards: %v", err)
	}

	if len(cards) != 1 {
		t.Errorf("expected still 1 card after upsert, got %d", len(cards))
	}

	if cards[0].Quantity != 3 {
		t.Errorf("expected quantity 3 after upsert, got %d", cards[0].Quantity)
	}
}

func TestDeckRepository_RemoveCard(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck and add a card
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	card := &models.DeckCard{
		DeckID:        "deck-1",
		CardID:        12345,
		Quantity:      4,
		Board:         "main",
		FromDraftPick: false,
	}

	err = repo.AddCard(ctx, card)
	if err != nil {
		t.Fatalf("failed to add card: %v", err)
	}

	// Remove the card
	err = repo.RemoveCard(ctx, "deck-1", 12345, "main")
	if err != nil {
		t.Fatalf("failed to remove card: %v", err)
	}

	// Verify it was removed
	cards, err := repo.GetCards(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get cards: %v", err)
	}

	if len(cards) != 0 {
		t.Errorf("expected 0 cards after removal, got %d", len(cards))
	}
}

func TestDeckRepository_ClearCards(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck and add multiple cards
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	cards := []*models.DeckCard{
		{DeckID: "deck-1", CardID: 12345, Quantity: 4, Board: "main", FromDraftPick: false},
		{DeckID: "deck-1", CardID: 67890, Quantity: 3, Board: "main", FromDraftPick: false},
		{DeckID: "deck-1", CardID: 11111, Quantity: 2, Board: "sideboard", FromDraftPick: false},
	}

	for _, c := range cards {
		if err := repo.AddCard(ctx, c); err != nil {
			t.Fatalf("failed to add card: %v", err)
		}
	}

	// Clear all cards
	err = repo.ClearCards(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to clear cards: %v", err)
	}

	// Verify all cards were removed
	retrieved, err := repo.GetCards(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get cards: %v", err)
	}

	if len(retrieved) != 0 {
		t.Errorf("expected 0 cards after clear, got %d", len(retrieved))
	}
}

// Tests for new v1.3 methods

func TestDeckRepository_GetBySource(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create decks with different sources
	decks := []*models.Deck{
		{
			ID:            "deck-1",
			AccountID:     1,
			Name:          "Draft Deck",
			Format:        "Limited",
			Source:        "draft",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now,
		},
		{
			ID:            "deck-2",
			AccountID:     1,
			Name:          "Constructed Deck",
			Format:        "Standard",
			Source:        "constructed",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now,
		},
		{
			ID:            "deck-3",
			AccountID:     1,
			Name:          "Imported Deck",
			Format:        "Historic",
			Source:        "imported",
			MatchesPlayed: 0,
			MatchesWon:    0,
			GamesPlayed:   0,
			GamesWon:      0,
			CreatedAt:     now,
			ModifiedAt:    now,
		},
	}

	for _, d := range decks {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("failed to create deck: %v", err)
		}
	}

	// Get draft decks
	draftDecks, err := repo.GetBySource(ctx, 1, "draft")
	if err != nil {
		t.Fatalf("failed to get draft decks: %v", err)
	}

	if len(draftDecks) != 1 {
		t.Errorf("expected 1 draft deck, got %d", len(draftDecks))
	}

	if len(draftDecks) > 0 && draftDecks[0].Source != "draft" {
		t.Errorf("expected source 'draft', got '%s'", draftDecks[0].Source)
	}

	// Get constructed decks
	constructedDecks, err := repo.GetBySource(ctx, 1, "constructed")
	if err != nil {
		t.Fatalf("failed to get constructed decks: %v", err)
	}

	if len(constructedDecks) != 1 {
		t.Errorf("expected 1 constructed deck, got %d", len(constructedDecks))
	}

	// Get imported decks
	importedDecks, err := repo.GetBySource(ctx, 1, "imported")
	if err != nil {
		t.Fatalf("failed to get imported decks: %v", err)
	}

	if len(importedDecks) != 1 {
		t.Errorf("expected 1 imported deck, got %d", len(importedDecks))
	}
}

func TestDeckRepository_GetByDraftEvent(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a draft event
	draftEventID := "draft-event-1"
	_, err := db.ExecContext(ctx, `
		INSERT INTO draft_events (id, account_id, event_name, set_code, start_time)
		VALUES (?, ?, ?, ?, ?)
	`, draftEventID, 1, "Quick Draft BRO", "BRO", now.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("failed to create draft event: %v", err)
	}

	// Create a draft deck
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Draft Deck",
		Format:        "Limited",
		Source:        "draft",
		DraftEventID:  &draftEventID,
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err = repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Get deck by draft event
	retrieved, err := repo.GetByDraftEvent(ctx, draftEventID)
	if err != nil {
		t.Fatalf("failed to get deck by draft event: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected deck to be found")
	}

	if retrieved.ID != "deck-1" {
		t.Errorf("expected deck ID 'deck-1', got '%s'", retrieved.ID)
	}

	if retrieved.DraftEventID == nil || *retrieved.DraftEventID != draftEventID {
		t.Error("expected draft event ID to match")
	}
}

func TestDeckRepository_UpdatePerformance(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Update performance - won match 2-1
	err = repo.UpdatePerformance(ctx, "deck-1", true, 2, 1)
	if err != nil {
		t.Fatalf("failed to update performance: %v", err)
	}

	// Verify performance was updated
	retrieved, err := repo.GetByID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get deck: %v", err)
	}

	if retrieved.MatchesPlayed != 1 {
		t.Errorf("expected 1 match played, got %d", retrieved.MatchesPlayed)
	}

	if retrieved.MatchesWon != 1 {
		t.Errorf("expected 1 match won, got %d", retrieved.MatchesWon)
	}

	if retrieved.GamesPlayed != 3 {
		t.Errorf("expected 3 games played, got %d", retrieved.GamesPlayed)
	}

	if retrieved.GamesWon != 2 {
		t.Errorf("expected 2 games won, got %d", retrieved.GamesWon)
	}

	// Update performance - lost match 1-2
	err = repo.UpdatePerformance(ctx, "deck-1", false, 1, 2)
	if err != nil {
		t.Fatalf("failed to update performance: %v", err)
	}

	// Verify cumulative performance
	retrieved, err = repo.GetByID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get deck: %v", err)
	}

	if retrieved.MatchesPlayed != 2 {
		t.Errorf("expected 2 matches played, got %d", retrieved.MatchesPlayed)
	}

	if retrieved.MatchesWon != 1 {
		t.Errorf("expected 1 match won, got %d", retrieved.MatchesWon)
	}

	if retrieved.GamesPlayed != 6 {
		t.Errorf("expected 6 games played, got %d", retrieved.GamesPlayed)
	}

	if retrieved.GamesWon != 3 {
		t.Errorf("expected 3 games won, got %d", retrieved.GamesWon)
	}
}

func TestDeckRepository_GetPerformance(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck with some performance data
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 10,
		MatchesWon:    7,
		GamesPlayed:   25,
		GamesWon:      17,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Get performance
	perf, err := repo.GetPerformance(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get performance: %v", err)
	}

	if perf.MatchesPlayed != 10 {
		t.Errorf("expected 10 matches played, got %d", perf.MatchesPlayed)
	}

	if perf.MatchesWon != 7 {
		t.Errorf("expected 7 matches won, got %d", perf.MatchesWon)
	}

	if perf.MatchesLost != 3 {
		t.Errorf("expected 3 matches lost, got %d", perf.MatchesLost)
	}

	// Win rate should be 7/10 = 0.70
	expectedWinRate := 0.70
	if perf.MatchWinRate < expectedWinRate-0.01 || perf.MatchWinRate > expectedWinRate+0.01 {
		t.Errorf("expected match win rate ~%.2f, got %.2f", expectedWinRate, perf.MatchWinRate)
	}

	// Game win rate should be 17/25 = 0.68
	expectedGameWinRate := 0.68
	if perf.GameWinRate < expectedGameWinRate-0.01 || perf.GameWinRate > expectedGameWinRate+0.01 {
		t.Errorf("expected game win rate ~%.2f, got %.2f", expectedGameWinRate, perf.GameWinRate)
	}
}

func TestDeckRepository_GetDraftCards(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a draft event
	draftEventID := "draft-event-1"
	_, err := db.ExecContext(ctx, `
		INSERT INTO draft_events (id, account_id, event_name, set_code, start_time)
		VALUES (?, ?, ?, ?, ?)
	`, draftEventID, 1, "Quick Draft BRO", "BRO", now.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("failed to create draft event: %v", err)
	}

	// Add some draft picks
	cardIDs := []int{12345, 67890, 11111, 22222, 33333}
	for i, cardID := range cardIDs {
		_, err := db.ExecContext(ctx, `
			INSERT INTO draft_picks (draft_event_id, pack_number, pick_number, selected_card, timestamp)
			VALUES (?, ?, ?, ?, ?)
		`, draftEventID, 1, i+1, cardID, now.Format("2006-01-02 15:04:05"))
		if err != nil {
			t.Fatalf("failed to add draft pick: %v", err)
		}
	}

	// Get draft cards
	draftCards, err := repo.GetDraftCards(ctx, draftEventID)
	if err != nil {
		t.Fatalf("failed to get draft cards: %v", err)
	}

	if len(draftCards) != len(cardIDs) {
		t.Errorf("expected %d draft cards, got %d", len(cardIDs), len(draftCards))
	}

	// Verify all card IDs are present
	cardSet := make(map[int]bool)
	for _, cardID := range draftCards {
		cardSet[cardID] = true
	}

	for _, expectedID := range cardIDs {
		if !cardSet[expectedID] {
			t.Errorf("expected card ID %d in draft cards", expectedID)
		}
	}
}

func TestDeckRepository_ValidateDraftDeck(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a draft event
	draftEventID := "draft-event-1"
	_, err := db.ExecContext(ctx, `
		INSERT INTO draft_events (id, account_id, event_name, set_code, start_time)
		VALUES (?, ?, ?, ?, ?)
	`, draftEventID, 1, "Quick Draft BRO", "BRO", now.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("failed to create draft event: %v", err)
	}

	// Add draft picks
	draftCardIDs := []int{12345, 67890, 11111}
	for i, cardID := range draftCardIDs {
		_, err := db.ExecContext(ctx, `
			INSERT INTO draft_picks (draft_event_id, pack_number, pick_number, selected_card, timestamp)
			VALUES (?, ?, ?, ?, ?)
		`, draftEventID, 1, i+1, cardID, now.Format("2006-01-02 15:04:05"))
		if err != nil {
			t.Fatalf("failed to add draft pick: %v", err)
		}
	}

	// Create a draft deck
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Draft Deck",
		Format:        "Limited",
		Source:        "draft",
		DraftEventID:  &draftEventID,
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err = repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Add cards from the draft
	for _, cardID := range draftCardIDs {
		card := &models.DeckCard{
			DeckID:        "deck-1",
			CardID:        cardID,
			Quantity:      1,
			Board:         "main",
			FromDraftPick: true,
		}
		if err := repo.AddCard(ctx, card); err != nil {
			t.Fatalf("failed to add card: %v", err)
		}
	}

	// Validate - should pass
	valid, err := repo.ValidateDraftDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to validate draft deck: %v", err)
	}

	if !valid {
		t.Error("expected draft deck to be valid")
	}

	// Add a card that wasn't drafted
	invalidCard := &models.DeckCard{
		DeckID:        "deck-1",
		CardID:        99999, // Not in draft
		Quantity:      1,
		Board:         "main",
		FromDraftPick: false,
	}
	if err := repo.AddCard(ctx, invalidCard); err != nil {
		t.Fatalf("failed to add invalid card: %v", err)
	}

	// Validate - should fail
	valid, err = repo.ValidateDraftDeck(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to validate draft deck: %v", err)
	}

	if valid {
		t.Error("expected draft deck to be invalid with non-drafted card")
	}

	// Test that constructed decks always validate
	constructedDeck := &models.Deck{
		ID:            "deck-2",
		AccountID:     1,
		Name:          "Constructed Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err = repo.Create(ctx, constructedDeck)
	if err != nil {
		t.Fatalf("failed to create constructed deck: %v", err)
	}

	// Add any card to constructed deck
	anyCard := &models.DeckCard{
		DeckID:        "deck-2",
		CardID:        99999,
		Quantity:      4,
		Board:         "main",
		FromDraftPick: false,
	}
	if err := repo.AddCard(ctx, anyCard); err != nil {
		t.Fatalf("failed to add card to constructed deck: %v", err)
	}

	// Validate - should always pass for non-draft decks
	valid, err = repo.ValidateDraftDeck(ctx, "deck-2")
	if err != nil {
		t.Fatalf("failed to validate constructed deck: %v", err)
	}

	if !valid {
		t.Error("expected constructed deck to always be valid")
	}
}

func TestDeckRepository_Tags(t *testing.T) {
	db := setupDeckTestDB(t)
	defer db.Close()

	repo := NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck
	deck := &models.Deck{
		ID:            "deck-1",
		AccountID:     1,
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "constructed",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Add tags
	tagNames := []string{"aggro", "meta", "tier1"}
	for _, tagName := range tagNames {
		tag := &models.DeckTag{
			DeckID:    "deck-1",
			Tag:       tagName,
			CreatedAt: now,
		}
		if err := repo.AddTag(ctx, tag); err != nil {
			t.Fatalf("failed to add tag '%s': %v", tagName, err)
		}
	}

	// Get tags
	retrievedTags, err := repo.GetTags(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get tags: %v", err)
	}

	if len(retrievedTags) != len(tagNames) {
		t.Errorf("expected %d tags, got %d", len(tagNames), len(retrievedTags))
	}

	// Verify all tags are present
	tagMap := make(map[string]bool)
	for _, tag := range retrievedTags {
		tagMap[tag.Tag] = true
	}

	for _, expectedTag := range tagNames {
		if !tagMap[expectedTag] {
			t.Errorf("expected tag '%s' in retrieved tags", expectedTag)
		}
	}

	// Test adding duplicate tag (should be idempotent)
	duplicateTag := &models.DeckTag{
		DeckID:    "deck-1",
		Tag:       "aggro",
		CreatedAt: now,
	}
	err = repo.AddTag(ctx, duplicateTag)
	if err != nil {
		t.Fatalf("failed to add duplicate tag: %v", err)
	}

	retrievedTags, err = repo.GetTags(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get tags after duplicate: %v", err)
	}

	if len(retrievedTags) != len(tagNames) {
		t.Errorf("expected %d tags after duplicate add, got %d", len(tagNames), len(retrievedTags))
	}

	// Remove a tag
	err = repo.RemoveTag(ctx, "deck-1", "meta")
	if err != nil {
		t.Fatalf("failed to remove tag: %v", err)
	}

	retrievedTags, err = repo.GetTags(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get tags after removal: %v", err)
	}

	if len(retrievedTags) != 2 {
		t.Errorf("expected 2 tags after removal, got %d", len(retrievedTags))
	}

	// Verify "meta" was removed
	for _, tag := range retrievedTags {
		if tag.Tag == "meta" {
			t.Error("expected 'meta' tag to be removed")
		}
	}
}

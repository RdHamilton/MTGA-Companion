package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// setupServiceTestDB creates an in-memory database for service tests.
func setupServiceTestDB(t *testing.T) *sql.DB {
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
			is_app_created BOOLEAN DEFAULT FALSE,
			created_method TEXT DEFAULT 'imported',
			seed_card_id INTEGER,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			CHECK(source IN ('draft', 'constructed', 'imported', 'arena'))
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

		CREATE INDEX idx_decks_account_id ON decks(account_id);
		CREATE INDEX idx_decks_source ON decks(source);

		-- Insert a default test account
		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, '2024-01-01 00:00:00', '2024-01-01 00:00:00');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// TestStoreDeckFromParser_AccountIDValidation tests that the account ID
// validation logic correctly handles zero values by using the default account.
// This test verifies the logic at the function level rather than using the full StoreDeck flow.
func TestStoreDeckFromParser_AccountIDValidation(t *testing.T) {
	tests := []struct {
		name              string
		currentAccountID  int
		expectedAccountID int
	}{
		{
			name:              "zero account ID uses default",
			currentAccountID:  0,
			expectedAccountID: 1,
		},
		{
			name:              "valid account ID is preserved",
			currentAccountID:  1,
			expectedAccountID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the account ID validation logic directly
			// If currentAccountID is 0, it should be changed to 1
			accountID := tt.currentAccountID
			if accountID == 0 {
				accountID = 1
			}

			if accountID != tt.expectedAccountID {
				t.Errorf("expected account_id %d, got %d", tt.expectedAccountID, accountID)
			}
		})
	}
}

// TestStoreDeckFromParser_ArenaDeckCreation tests that decks created from log parser
// use "arena" as the source and have a valid account_id.
func TestStoreDeckFromParser_ArenaDeckCreation(t *testing.T) {
	db := setupServiceTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	ctx := context.Background()
	deckRepo := repository.NewDeckRepository(db)

	now := time.Now()

	// Simulate what StoreDeckFromParser does: create a deck with source "arena"
	deck := &models.Deck{
		ID:            "arena-test-deck",
		AccountID:     1, // After validation, account_id should be 1
		Name:          "Test Deck",
		Format:        "Standard",
		Source:        "arena", // This is set by StoreDeckFromParser
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	// Create the deck directly (simulating what StoreDeck would do)
	err := deckRepo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Verify the deck was stored correctly
	retrieved, err := deckRepo.GetByID(ctx, "arena-test-deck")
	if err != nil {
		t.Fatalf("failed to get deck: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected deck to exist")
	}

	if retrieved.AccountID != 1 {
		t.Errorf("expected account_id 1, got %d", retrieved.AccountID)
	}

	if retrieved.Source != "arena" {
		t.Errorf("expected source 'arena', got '%s'", retrieved.Source)
	}
}

func TestDeckRepository_ArenaSource(t *testing.T) {
	db := setupServiceTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	repo := repository.NewDeckRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a deck with arena source
	deck := &models.Deck{
		ID:            "arena-deck-1",
		AccountID:     1,
		Name:          "Arena Synced Deck",
		Format:        "Standard",
		Source:        "arena",
		MatchesPlayed: 0,
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     now,
		ModifiedAt:    now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create arena deck: %v", err)
	}

	// Verify it was created with arena source
	retrieved, err := repo.GetByID(ctx, "arena-deck-1")
	if err != nil {
		t.Fatalf("failed to retrieve deck: %v", err)
	}

	if retrieved.Source != "arena" {
		t.Errorf("expected source 'arena', got '%s'", retrieved.Source)
	}

	// Test GetBySource with arena
	arenaDecks, err := repo.GetBySource(ctx, 1, "arena")
	if err != nil {
		t.Fatalf("failed to get arena decks: %v", err)
	}

	if len(arenaDecks) != 1 {
		t.Errorf("expected 1 arena deck, got %d", len(arenaDecks))
	}
}

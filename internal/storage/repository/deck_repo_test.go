package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupDeckTestDB creates an in-memory database with decks and deck_cards tables.
func setupDeckTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE decks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			format TEXT NOT NULL,
			description TEXT,
			color_identity TEXT,
			created_at DATETIME NOT NULL,
			modified_at DATETIME NOT NULL,
			last_played DATETIME
		);

		CREATE TABLE deck_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id TEXT NOT NULL,
			card_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			board TEXT NOT NULL,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
			UNIQUE(deck_id, card_id, board)
		);

		CREATE INDEX idx_deck_cards_deck_id ON deck_cards(deck_id);
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
		Name:          "Izzet Phoenix",
		Format:        "Historic",
		ColorIdentity: &colorIdentity,
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
		ID:         "deck-1",
		Name:       "Original Name",
		Format:     "Standard",
		CreatedAt:  now,
		ModifiedAt: now,
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

	// Create multiple decks
	decks := []*models.Deck{
		{
			ID:         "deck-1",
			Name:       "Deck 1",
			Format:     "Standard",
			CreatedAt:  now,
			ModifiedAt: now,
		},
		{
			ID:         "deck-2",
			Name:       "Deck 2",
			Format:     "Historic",
			CreatedAt:  now,
			ModifiedAt: now.Add(1 * time.Hour),
		},
	}

	for _, d := range decks {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("failed to create deck: %v", err)
		}
	}

	// List all decks
	results, err := repo.List(ctx)
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
			ID:         "deck-1",
			Name:       "Standard Deck",
			Format:     "Standard",
			CreatedAt:  now,
			ModifiedAt: now,
		},
		{
			ID:         "deck-2",
			Name:       "Historic Deck",
			Format:     "Historic",
			CreatedAt:  now,
			ModifiedAt: now,
		},
	}

	for _, d := range decks {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("failed to create deck: %v", err)
		}
	}

	// Get Standard decks
	results, err := repo.GetByFormat(ctx, "Standard")
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
		ID:         "deck-1",
		Name:       "Deck to Delete",
		Format:     "Standard",
		CreatedAt:  now,
		ModifiedAt: now,
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
		ID:         "deck-1",
		Name:       "Test Deck",
		Format:     "Standard",
		CreatedAt:  now,
		ModifiedAt: now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	// Add a card
	card := &models.DeckCard{
		DeckID:   "deck-1",
		CardID:   12345,
		Quantity: 4,
		Board:    "main",
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
		ID:         "deck-1",
		Name:       "Test Deck",
		Format:     "Standard",
		CreatedAt:  now,
		ModifiedAt: now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	card := &models.DeckCard{
		DeckID:   "deck-1",
		CardID:   12345,
		Quantity: 4,
		Board:    "main",
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
		ID:         "deck-1",
		Name:       "Test Deck",
		Format:     "Standard",
		CreatedAt:  now,
		ModifiedAt: now,
	}

	err := repo.Create(ctx, deck)
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	cards := []*models.DeckCard{
		{DeckID: "deck-1", CardID: 12345, Quantity: 4, Board: "main"},
		{DeckID: "deck-1", CardID: 67890, Quantity: 3, Board: "main"},
		{DeckID: "deck-1", CardID: 11111, Quantity: 2, Board: "sideboard"},
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

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupPermutationTestDB creates an in-memory database with permutation-related tables.
func setupPermutationTestDB(t *testing.T) *sql.DB {
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
			current_permutation_id INTEGER,
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
			UNIQUE(deck_id, card_id, board)
		);

		CREATE TABLE deck_permutations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			deck_id TEXT NOT NULL,
			parent_permutation_id INTEGER,
			cards TEXT NOT NULL,
			card_hash TEXT NOT NULL,
			version_number INTEGER NOT NULL DEFAULT 1,
			version_name TEXT,
			change_summary TEXT,
			matches_played INTEGER NOT NULL DEFAULT 0,
			matches_won INTEGER NOT NULL DEFAULT 0,
			games_played INTEGER NOT NULL DEFAULT 0,
			games_won INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_played_at DATETIME,
			FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_permutation_id) REFERENCES deck_permutations(id) ON DELETE SET NULL
		);

		CREATE INDEX idx_deck_permutations_deck_id ON deck_permutations(deck_id);
		CREATE INDEX idx_deck_permutations_parent ON deck_permutations(parent_permutation_id);
		CREATE INDEX idx_deck_permutations_created ON deck_permutations(deck_id, created_at DESC);
		CREATE INDEX idx_deck_permutations_hash ON deck_permutations(deck_id, card_hash);

		-- Insert a default test account
		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, '2024-01-01 00:00:00', '2024-01-01 00:00:00');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// testComputeCardHash generates a deterministic hash for test cards.
func testComputeCardHash(cards []models.DeckPermutationCard) string {
	sorted := make([]models.DeckPermutationCard, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CardID != sorted[j].CardID {
			return sorted[i].CardID < sorted[j].CardID
		}
		return sorted[i].Board < sorted[j].Board
	})

	var parts []string
	for _, card := range sorted {
		parts = append(parts, fmt.Sprintf("%d:%d:%s", card.CardID, card.Quantity, card.Board))
	}
	return strings.Join(parts, "|")
}

// createTestDeck creates a test deck with cards.
func createTestDeck(t *testing.T, db *sql.DB, deckID, name string) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	_, err := db.Exec(`
		INSERT INTO decks (id, account_id, name, format, source, created_at, modified_at)
		VALUES (?, 1, ?, 'Standard', 'constructed', ?, ?)
	`, deckID, name, now, now)
	if err != nil {
		t.Fatalf("failed to create test deck: %v", err)
	}

	// Add some test cards
	cards := []struct {
		cardID   int
		quantity int
		board    string
	}{
		{100, 4, "main"},
		{101, 4, "main"},
		{102, 3, "main"},
		{200, 2, "sideboard"},
	}

	for _, card := range cards {
		_, err := db.Exec(`
			INSERT INTO deck_cards (deck_id, card_id, quantity, board)
			VALUES (?, ?, ?, ?)
		`, deckID, card.cardID, card.quantity, card.board)
		if err != nil {
			t.Fatalf("failed to add test card: %v", err)
		}
	}
}

func TestDeckPermutationRepository_Create(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	cards := []models.DeckPermutationCard{
		{CardID: 100, Quantity: 4, Board: "main"},
		{CardID: 101, Quantity: 4, Board: "main"},
	}
	cardsJSON, _ := json.Marshal(cards)
	cardHash := testComputeCardHash(cards)

	versionName := "Initial Version"
	perm := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON),
		CardHash:      cardHash,
		VersionNumber: 1,
		VersionName:   &versionName,
		CreatedAt:     time.Now(),
	}

	err := repo.Create(ctx, perm)
	if err != nil {
		t.Fatalf("failed to create permutation: %v", err)
	}

	if perm.ID == 0 {
		t.Error("expected permutation ID to be set after create")
	}

	// Verify it was created
	retrieved, err := repo.GetByID(ctx, perm.ID)
	if err != nil {
		t.Fatalf("failed to retrieve permutation: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected permutation to be found")
	}

	if retrieved.DeckID != "deck-1" {
		t.Errorf("expected deck_id 'deck-1', got '%s'", retrieved.DeckID)
	}

	if retrieved.VersionNumber != 1 {
		t.Errorf("expected version_number 1, got %d", retrieved.VersionNumber)
	}

	if *retrieved.VersionName != "Initial Version" {
		t.Errorf("expected version_name 'Initial Version', got '%s'", *retrieved.VersionName)
	}

	if retrieved.CardHash != cardHash {
		t.Errorf("expected card_hash '%s', got '%s'", cardHash, retrieved.CardHash)
	}
}

func TestDeckPermutationRepository_GetByDeckID(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create multiple permutations
	for i := 1; i <= 3; i++ {
		cards := []models.DeckPermutationCard{{CardID: 100 + i, Quantity: 4, Board: "main"}}
		cardsJSON, _ := json.Marshal(cards)
		cardHash := testComputeCardHash(cards)

		perm := &models.DeckPermutation{
			DeckID:        "deck-1",
			Cards:         string(cardsJSON),
			CardHash:      cardHash,
			VersionNumber: i,
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, perm); err != nil {
			t.Fatalf("failed to create permutation %d: %v", i, err)
		}
	}

	// Retrieve all permutations
	perms, err := repo.GetByDeckID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get permutations: %v", err)
	}

	if len(perms) != 3 {
		t.Errorf("expected 3 permutations, got %d", len(perms))
	}

	// Verify they're ordered by version number ascending
	for i, perm := range perms {
		if perm.VersionNumber != i+1 {
			t.Errorf("expected version %d at index %d, got %d", i+1, i, perm.VersionNumber)
		}
	}
}

func TestDeckPermutationRepository_GetLatest(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create permutations with different timestamps
	for i := 1; i <= 3; i++ {
		cards := []models.DeckPermutationCard{{CardID: 100, Quantity: i, Board: "main"}}
		cardsJSON, _ := json.Marshal(cards)
		cardHash := testComputeCardHash(cards)

		perm := &models.DeckPermutation{
			DeckID:        "deck-1",
			Cards:         string(cardsJSON),
			CardHash:      cardHash,
			VersionNumber: i,
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, perm); err != nil {
			t.Fatalf("failed to create permutation: %v", err)
		}
	}

	latest, err := repo.GetLatest(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get latest permutation: %v", err)
	}

	if latest == nil {
		t.Fatal("expected to find latest permutation")
	}

	if latest.VersionNumber != 3 {
		t.Errorf("expected latest version to be 3, got %d", latest.VersionNumber)
	}
}

func TestDeckPermutationRepository_SetAndGetCurrent(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create two permutations
	var permIDs []int
	for i := 1; i <= 2; i++ {
		cards := []models.DeckPermutationCard{{CardID: 100, Quantity: i, Board: "main"}}
		cardsJSON, _ := json.Marshal(cards)
		cardHash := testComputeCardHash(cards)

		perm := &models.DeckPermutation{
			DeckID:        "deck-1",
			Cards:         string(cardsJSON),
			CardHash:      cardHash,
			VersionNumber: i,
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, perm); err != nil {
			t.Fatalf("failed to create permutation: %v", err)
		}
		permIDs = append(permIDs, perm.ID)
	}

	// Set the first permutation as current
	if err := repo.SetCurrentPermutation(ctx, "deck-1", permIDs[0]); err != nil {
		t.Fatalf("failed to set current permutation: %v", err)
	}

	// Get current should return the first permutation
	current, err := repo.GetCurrent(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get current permutation: %v", err)
	}

	if current.ID != permIDs[0] {
		t.Errorf("expected current permutation ID %d, got %d", permIDs[0], current.ID)
	}

	if current.VersionNumber != 1 {
		t.Errorf("expected current version 1, got %d", current.VersionNumber)
	}
}

func TestDeckPermutationRepository_SetCurrentPermutation_Validation(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	// Create two separate decks
	createTestDeck(t, db, "deck-1", "Test Deck 1")
	createTestDeck(t, db, "deck-2", "Test Deck 2")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create a permutation for deck-1
	cards := []models.DeckPermutationCard{{CardID: 100, Quantity: 4, Board: "main"}}
	cardsJSON, _ := json.Marshal(cards)
	cardHash := testComputeCardHash(cards)

	perm := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON),
		CardHash:      cardHash,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, perm); err != nil {
		t.Fatalf("failed to create permutation: %v", err)
	}

	// Try to set deck-1's permutation as current for deck-2 - should fail
	err := repo.SetCurrentPermutation(ctx, "deck-2", perm.ID)
	if err == nil {
		t.Error("expected error when setting permutation from different deck")
	}

	expectedErr := fmt.Sprintf("permutation %d does not belong to deck deck-2", perm.ID)
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestDeckPermutationRepository_GetByCardHash(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	cards := []models.DeckPermutationCard{
		{CardID: 100, Quantity: 4, Board: "main"},
		{CardID: 101, Quantity: 4, Board: "main"},
	}
	cardsJSON, _ := json.Marshal(cards)
	cardHash := testComputeCardHash(cards)

	perm := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON),
		CardHash:      cardHash,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, perm); err != nil {
		t.Fatalf("failed to create permutation: %v", err)
	}

	// Find by card hash
	found, err := repo.GetByCardHash(ctx, "deck-1", cardHash)
	if err != nil {
		t.Fatalf("failed to get by card hash: %v", err)
	}

	if found == nil {
		t.Fatal("expected to find permutation by card hash")
	}

	if found.ID != perm.ID {
		t.Errorf("expected permutation ID %d, got %d", perm.ID, found.ID)
	}

	// Search with different hash should return nil
	notFound, err := repo.GetByCardHash(ctx, "deck-1", "different-hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent hash")
	}
}

func TestDeckPermutationRepository_UpdatePerformance(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	cards := []models.DeckPermutationCard{{CardID: 100, Quantity: 4, Board: "main"}}
	cardsJSON, _ := json.Marshal(cards)
	cardHash := testComputeCardHash(cards)

	perm := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON),
		CardHash:      cardHash,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, perm); err != nil {
		t.Fatalf("failed to create permutation: %v", err)
	}

	// Update performance - win with 2-1 games
	if err := repo.UpdatePerformance(ctx, perm.ID, true, 2, 1); err != nil {
		t.Fatalf("failed to update performance: %v", err)
	}

	// Update performance - loss with 1-2 games
	if err := repo.UpdatePerformance(ctx, perm.ID, false, 1, 2); err != nil {
		t.Fatalf("failed to update performance: %v", err)
	}

	// Get performance
	perf, err := repo.GetPerformance(ctx, perm.ID)
	if err != nil {
		t.Fatalf("failed to get performance: %v", err)
	}

	if perf.MatchesPlayed != 2 {
		t.Errorf("expected 2 matches played, got %d", perf.MatchesPlayed)
	}

	if perf.MatchesWon != 1 {
		t.Errorf("expected 1 match won, got %d", perf.MatchesWon)
	}

	if perf.GamesPlayed != 6 {
		t.Errorf("expected 6 games played, got %d", perf.GamesPlayed)
	}

	if perf.GamesWon != 3 {
		t.Errorf("expected 3 games won, got %d", perf.GamesWon)
	}

	// Verify win rates
	expectedMatchWinRate := 0.5 // 1/2
	if perf.MatchWinRate != expectedMatchWinRate {
		t.Errorf("expected match win rate %.2f, got %.2f", expectedMatchWinRate, perf.MatchWinRate)
	}

	expectedGameWinRate := 0.5 // 3/6
	if perf.GameWinRate != expectedGameWinRate {
		t.Errorf("expected game win rate %.2f, got %.2f", expectedGameWinRate, perf.GameWinRate)
	}
}

func TestDeckPermutationRepository_GetDiff(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create first permutation
	cards1 := []models.DeckPermutationCard{
		{CardID: 100, Quantity: 4, Board: "main"},
		{CardID: 101, Quantity: 4, Board: "main"},
		{CardID: 102, Quantity: 3, Board: "main"},
	}
	cardsJSON1, _ := json.Marshal(cards1)
	cardHash1 := testComputeCardHash(cards1)

	perm1 := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON1),
		CardHash:      cardHash1,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, perm1); err != nil {
		t.Fatalf("failed to create first permutation: %v", err)
	}

	// Create second permutation with changes:
	// - Removed card 102
	// - Changed card 101 from 4 to 2
	// - Added card 103
	cards2 := []models.DeckPermutationCard{
		{CardID: 100, Quantity: 4, Board: "main"},
		{CardID: 101, Quantity: 2, Board: "main"},
		{CardID: 103, Quantity: 3, Board: "main"},
	}
	cardsJSON2, _ := json.Marshal(cards2)
	cardHash2 := testComputeCardHash(cards2)

	perm2 := &models.DeckPermutation{
		DeckID:              "deck-1",
		ParentPermutationID: &perm1.ID,
		Cards:               string(cardsJSON2),
		CardHash:            cardHash2,
		VersionNumber:       2,
		CreatedAt:           time.Now().Add(time.Hour),
	}
	if err := repo.Create(ctx, perm2); err != nil {
		t.Fatalf("failed to create second permutation: %v", err)
	}

	// Get diff
	diff, err := repo.GetDiff(ctx, perm1.ID, perm2.ID)
	if err != nil {
		t.Fatalf("failed to get diff: %v", err)
	}

	// Verify added cards
	if len(diff.AddedCards) != 1 {
		t.Errorf("expected 1 added card, got %d", len(diff.AddedCards))
	} else if diff.AddedCards[0].CardID != 103 {
		t.Errorf("expected added card 103, got %d", diff.AddedCards[0].CardID)
	}

	// Verify removed cards
	if len(diff.RemovedCards) != 1 {
		t.Errorf("expected 1 removed card, got %d", len(diff.RemovedCards))
	} else if diff.RemovedCards[0].CardID != 102 {
		t.Errorf("expected removed card 102, got %d", diff.RemovedCards[0].CardID)
	}

	// Verify changed cards
	if len(diff.ChangedCards) != 1 {
		t.Errorf("expected 1 changed card, got %d", len(diff.ChangedCards))
	} else {
		change := diff.ChangedCards[0]
		if change.CardID != 101 {
			t.Errorf("expected changed card 101, got %d", change.CardID)
		}
		if change.OldQuantity != 4 {
			t.Errorf("expected old quantity 4, got %d", change.OldQuantity)
		}
		if change.NewQuantity != 2 {
			t.Errorf("expected new quantity 2, got %d", change.NewQuantity)
		}
	}
}

func TestDeckPermutationRepository_CreateFromCurrentDeck(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	versionName := "v1 - Anti-Aggro"
	changeSummary := "Added more removal"

	perm, err := repo.CreateFromCurrentDeck(ctx, "deck-1", &versionName, &changeSummary)
	if err != nil {
		t.Fatalf("failed to create permutation from deck: %v", err)
	}

	if perm.ID == 0 {
		t.Error("expected permutation ID to be set")
	}

	if perm.VersionNumber != 1 {
		t.Errorf("expected version number 1, got %d", perm.VersionNumber)
	}

	if *perm.VersionName != "v1 - Anti-Aggro" {
		t.Errorf("expected version name 'v1 - Anti-Aggro', got '%s'", *perm.VersionName)
	}

	if perm.CardHash == "" {
		t.Error("expected card_hash to be set")
	}

	// Parse the cards JSON and verify
	var cards []models.DeckPermutationCard
	if err := json.Unmarshal([]byte(perm.Cards), &cards); err != nil {
		t.Fatalf("failed to parse cards JSON: %v", err)
	}

	// Our test deck has 4 cards (3 main, 1 sideboard)
	if len(cards) != 4 {
		t.Errorf("expected 4 cards in permutation, got %d", len(cards))
	}

	// Verify the permutation is set as current
	current, err := repo.GetCurrent(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get current permutation: %v", err)
	}

	if current.ID != perm.ID {
		t.Errorf("expected current permutation to be %d, got %d", perm.ID, current.ID)
	}
}

func TestDeckPermutationRepository_CreateFromCurrentDeck_ReturnsExisting(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create first permutation
	versionName1 := "v1"
	perm1, err := repo.CreateFromCurrentDeck(ctx, "deck-1", &versionName1, nil)
	if err != nil {
		t.Fatalf("failed to create first permutation: %v", err)
	}

	// Create second permutation with same cards - should return existing
	versionName2 := "v2"
	perm2, err := repo.CreateFromCurrentDeck(ctx, "deck-1", &versionName2, nil)
	if err != nil {
		t.Fatalf("failed to create second permutation: %v", err)
	}

	// Should return the same permutation (same card_hash)
	if perm2.ID != perm1.ID {
		t.Errorf("expected same permutation ID %d, got %d", perm1.ID, perm2.ID)
	}

	// Version name should still be from the original
	if *perm2.VersionName != "v1" {
		t.Errorf("expected version name 'v1', got '%s'", *perm2.VersionName)
	}

	// Check that only one permutation exists
	perms, err := repo.GetByDeckID(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get permutations: %v", err)
	}
	if len(perms) != 1 {
		t.Errorf("expected 1 permutation, got %d", len(perms))
	}
}

func TestDeckPermutationRepository_GetNextVersionNumber(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// First version should be 1
	nextVersion, err := repo.GetNextVersionNumber(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get next version number: %v", err)
	}

	if nextVersion != 1 {
		t.Errorf("expected next version 1 for new deck, got %d", nextVersion)
	}

	// Create a permutation
	cards := []models.DeckPermutationCard{{CardID: 100, Quantity: 4, Board: "main"}}
	cardsJSON, _ := json.Marshal(cards)
	cardHash := testComputeCardHash(cards)

	perm := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON),
		CardHash:      cardHash,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, perm); err != nil {
		t.Fatalf("failed to create permutation: %v", err)
	}

	// Next version should now be 2
	nextVersion, err = repo.GetNextVersionNumber(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get next version number: %v", err)
	}

	if nextVersion != 2 {
		t.Errorf("expected next version 2, got %d", nextVersion)
	}
}

func TestDeckPermutationRepository_Delete(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	cards := []models.DeckPermutationCard{{CardID: 100, Quantity: 4, Board: "main"}}
	cardsJSON, _ := json.Marshal(cards)
	cardHash := testComputeCardHash(cards)

	perm := &models.DeckPermutation{
		DeckID:        "deck-1",
		Cards:         string(cardsJSON),
		CardHash:      cardHash,
		VersionNumber: 1,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(ctx, perm); err != nil {
		t.Fatalf("failed to create permutation: %v", err)
	}

	// Delete the permutation
	if err := repo.Delete(ctx, perm.ID); err != nil {
		t.Fatalf("failed to delete permutation: %v", err)
	}

	// Verify it's gone
	deleted, err := repo.GetByID(ctx, perm.ID)
	if err != nil {
		t.Fatalf("error getting deleted permutation: %v", err)
	}

	if deleted != nil {
		t.Error("expected permutation to be deleted")
	}
}

func TestDeckPermutationRepository_GetAllPerformance(t *testing.T) {
	db := setupPermutationTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	createTestDeck(t, db, "deck-1", "Test Deck")

	repo := NewDeckPermutationRepository(db)
	ctx := context.Background()

	// Create multiple permutations with different performance
	for i := 1; i <= 3; i++ {
		cards := []models.DeckPermutationCard{{CardID: 100 + i, Quantity: i, Board: "main"}}
		cardsJSON, _ := json.Marshal(cards)
		cardHash := testComputeCardHash(cards)

		perm := &models.DeckPermutation{
			DeckID:        "deck-1",
			Cards:         string(cardsJSON),
			CardHash:      cardHash,
			VersionNumber: i,
			MatchesPlayed: i * 10,
			MatchesWon:    i * 5,
			GamesPlayed:   i * 20,
			GamesWon:      i * 10,
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour),
		}
		if err := repo.Create(ctx, perm); err != nil {
			t.Fatalf("failed to create permutation %d: %v", i, err)
		}
	}

	// Get all performance
	perfs, err := repo.GetAllPerformance(ctx, "deck-1")
	if err != nil {
		t.Fatalf("failed to get all performance: %v", err)
	}

	if len(perfs) != 3 {
		t.Errorf("expected 3 performance records, got %d", len(perfs))
	}

	// Verify order and data
	for i, perf := range perfs {
		expectedVersion := i + 1
		if perf.VersionNumber != expectedVersion {
			t.Errorf("expected version %d at index %d, got %d", expectedVersion, i, perf.VersionNumber)
		}

		expectedMatches := expectedVersion * 10
		if perf.MatchesPlayed != expectedMatches {
			t.Errorf("expected %d matches for version %d, got %d", expectedMatches, expectedVersion, perf.MatchesPlayed)
		}

		// Win rate should be 50%
		if perf.MatchWinRate != 0.5 {
			t.Errorf("expected 50%% match win rate for version %d, got %.2f", expectedVersion, perf.MatchWinRate)
		}
	}
}

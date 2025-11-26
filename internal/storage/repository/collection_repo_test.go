package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// setupCollectionTestDB creates an in-memory database with collection tables.
func setupCollectionTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Schema with composite primary key (account_id, card_id) as per migration 000002
	schema := `
		CREATE TABLE collection (
			account_id INTEGER NOT NULL DEFAULT 1,
			card_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (account_id, card_id)
		);

		CREATE TABLE collection_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_id INTEGER NOT NULL,
			quantity_delta INTEGER NOT NULL,
			quantity_after INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_collection_history_timestamp ON collection_history(timestamp);
		CREATE INDEX idx_collection_history_card_id ON collection_history(card_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestCollectionRepository_UpsertCard(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Insert new card
	err := repo.UpsertCard(ctx, 12345, 4)
	if err != nil {
		t.Fatalf("failed to upsert card: %v", err)
	}

	// Verify it was inserted
	quantity, err := repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	if quantity != 4 {
		t.Errorf("expected quantity 4, got %d", quantity)
	}

	// Update existing card
	err = repo.UpsertCard(ctx, 12345, 7)
	if err != nil {
		t.Fatalf("failed to upsert card: %v", err)
	}

	// Verify it was updated
	quantity, err = repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	if quantity != 7 {
		t.Errorf("expected quantity 7 after update, got %d", quantity)
	}
}

func TestCollectionRepository_GetCard(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Test getting non-existent card (should return 0)
	quantity, err := repo.GetCard(ctx, 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quantity != 0 {
		t.Errorf("expected quantity 0 for non-existent card, got %d", quantity)
	}

	// Add a card and retrieve it
	err = repo.UpsertCard(ctx, 12345, 3)
	if err != nil {
		t.Fatalf("failed to upsert card: %v", err)
	}

	quantity, err = repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	if quantity != 3 {
		t.Errorf("expected quantity 3, got %d", quantity)
	}
}

func TestCollectionRepository_GetAll(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Add multiple cards
	cards := map[int]int{
		12345: 4,
		67890: 2,
		11111: 1,
	}

	for cardID, quantity := range cards {
		err := repo.UpsertCard(ctx, cardID, quantity)
		if err != nil {
			t.Fatalf("failed to upsert card %d: %v", cardID, err)
		}
	}

	// Get all cards
	collection, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all cards: %v", err)
	}

	if len(collection) != 3 {
		t.Errorf("expected 3 cards, got %d", len(collection))
	}

	for cardID, expectedQty := range cards {
		if qty, ok := collection[cardID]; !ok {
			t.Errorf("card %d not found in collection", cardID)
		} else if qty != expectedQty {
			t.Errorf("card %d: expected quantity %d, got %d", cardID, expectedQty, qty)
		}
	}
}

func TestCollectionRepository_RecordChange(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	now := time.Now()
	source := "pack"

	// Record adding 4 cards
	err := repo.RecordChange(ctx, 12345, 4, now, &source)
	if err != nil {
		t.Fatalf("failed to record change: %v", err)
	}

	// Verify collection was updated
	quantity, err := repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	if quantity != 4 {
		t.Errorf("expected quantity 4, got %d", quantity)
	}

	// Record adding 2 more cards
	err = repo.RecordChange(ctx, 12345, 2, now.Add(1*time.Hour), &source)
	if err != nil {
		t.Fatalf("failed to record second change: %v", err)
	}

	// Verify collection was updated
	quantity, err = repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	if quantity != 6 {
		t.Errorf("expected quantity 6, got %d", quantity)
	}

	// Verify history
	history, err := repo.GetHistory(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}

	// History should be in reverse chronological order
	if history[0].QuantityDelta != 2 {
		t.Errorf("expected first entry delta 2, got %d", history[0].QuantityDelta)
	}

	if history[0].QuantityAfter != 6 {
		t.Errorf("expected first entry quantity after 6, got %d", history[0].QuantityAfter)
	}

	if history[1].QuantityDelta != 4 {
		t.Errorf("expected second entry delta 4, got %d", history[1].QuantityDelta)
	}

	if history[1].QuantityAfter != 4 {
		t.Errorf("expected second entry quantity after 4, got %d", history[1].QuantityAfter)
	}
}

func TestCollectionRepository_GetHistory(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	now := time.Now()
	source := "draft"

	// Record multiple changes
	cardID := 12345
	deltas := []int{2, 1, -1, 3}

	for i, delta := range deltas {
		timestamp := now.Add(time.Duration(i) * time.Hour)
		err := repo.RecordChange(ctx, cardID, delta, timestamp, &source)
		if err != nil {
			t.Fatalf("failed to record change %d: %v", i, err)
		}
	}

	// Get history
	history, err := repo.GetHistory(ctx, cardID)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 4 {
		t.Fatalf("expected 4 history entries, got %d", len(history))
	}

	// Verify descending order
	for i := 0; i < len(history)-1; i++ {
		if history[i].Timestamp.Before(history[i+1].Timestamp) {
			t.Error("expected history in descending timestamp order")
		}
	}
}

func TestCollectionRepository_GetRecentChanges(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	now := time.Now()
	source := "pack"

	// Record changes for multiple cards
	changes := []struct {
		cardID int
		delta  int
		offset time.Duration
	}{
		{12345, 2, 0},
		{67890, 1, 1 * time.Hour},
		{11111, 3, 2 * time.Hour},
		{22222, 1, 3 * time.Hour},
		{33333, 4, 4 * time.Hour},
	}

	for _, change := range changes {
		timestamp := now.Add(change.offset)
		err := repo.RecordChange(ctx, change.cardID, change.delta, timestamp, &source)
		if err != nil {
			t.Fatalf("failed to record change for card %d: %v", change.cardID, err)
		}
	}

	// Get recent changes (limit 3)
	recent, err := repo.GetRecentChanges(ctx, 3)
	if err != nil {
		t.Fatalf("failed to get recent changes: %v", err)
	}

	if len(recent) != 3 {
		t.Fatalf("expected 3 recent changes, got %d", len(recent))
	}

	// Should be in descending timestamp order
	expectedCardIDs := []int{33333, 22222, 11111}
	for i, h := range recent {
		if h.CardID != expectedCardIDs[i] {
			t.Errorf("position %d: expected card %d, got %d", i, expectedCardIDs[i], h.CardID)
		}
	}
}

func TestCollectionRepository_NegativeDelta(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	now := time.Now()
	source := "craft"

	// Add 5 cards
	err := repo.RecordChange(ctx, 12345, 5, now, &source)
	if err != nil {
		t.Fatalf("failed to record initial change: %v", err)
	}

	// Remove 2 cards
	err = repo.RecordChange(ctx, 12345, -2, now.Add(1*time.Hour), &source)
	if err != nil {
		t.Fatalf("failed to record negative change: %v", err)
	}

	// Verify final quantity
	quantity, err := repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}

	if quantity != 3 {
		t.Errorf("expected quantity 3 after negative delta, got %d", quantity)
	}

	// Verify history
	history, err := repo.GetHistory(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}

	if history[0].QuantityDelta != -2 {
		t.Errorf("expected delta -2, got %d", history[0].QuantityDelta)
	}
}

func TestCollectionRepository_UpsertMany(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Prepare test entries
	entries := []CollectionEntry{
		{CardID: 12345, Quantity: 4},
		{CardID: 67890, Quantity: 2},
		{CardID: 11111, Quantity: 1},
		{CardID: 22222, Quantity: 3},
	}

	// Upsert multiple cards
	err := repo.UpsertMany(ctx, entries)
	if err != nil {
		t.Fatalf("failed to upsert many: %v", err)
	}

	// Verify all cards were inserted
	collection, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all: %v", err)
	}

	if len(collection) != 4 {
		t.Errorf("expected 4 cards, got %d", len(collection))
	}

	for _, entry := range entries {
		if qty, ok := collection[entry.CardID]; !ok {
			t.Errorf("card %d not found", entry.CardID)
		} else if qty != entry.Quantity {
			t.Errorf("card %d: expected quantity %d, got %d", entry.CardID, entry.Quantity, qty)
		}
	}
}

func TestCollectionRepository_UpsertMany_Update(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Insert initial cards
	initialEntries := []CollectionEntry{
		{CardID: 12345, Quantity: 2},
		{CardID: 67890, Quantity: 1},
	}

	err := repo.UpsertMany(ctx, initialEntries)
	if err != nil {
		t.Fatalf("failed to upsert initial entries: %v", err)
	}

	// Update some, add new
	updatedEntries := []CollectionEntry{
		{CardID: 12345, Quantity: 4}, // Update
		{CardID: 67890, Quantity: 3}, // Update
		{CardID: 11111, Quantity: 1}, // New
	}

	err = repo.UpsertMany(ctx, updatedEntries)
	if err != nil {
		t.Fatalf("failed to upsert updated entries: %v", err)
	}

	// Verify updates
	collection, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all: %v", err)
	}

	expected := map[int]int{
		12345: 4,
		67890: 3,
		11111: 1,
	}

	for cardID, expectedQty := range expected {
		if qty := collection[cardID]; qty != expectedQty {
			t.Errorf("card %d: expected quantity %d, got %d", cardID, expectedQty, qty)
		}
	}
}

func TestCollectionRepository_UpsertMany_Empty(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Upsert empty slice should not error
	err := repo.UpsertMany(ctx, []CollectionEntry{})
	if err != nil {
		t.Errorf("UpsertMany with empty slice should not error: %v", err)
	}

	// Upsert nil should not error
	err = repo.UpsertMany(ctx, nil)
	if err != nil {
		t.Errorf("UpsertMany with nil should not error: %v", err)
	}
}

func TestCollectionRepository_UpsertMany_LargeCollection(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// Create a large collection (simulating full collection sync)
	const numCards = 5000
	entries := make([]CollectionEntry, numCards)
	for i := 0; i < numCards; i++ {
		entries[i] = CollectionEntry{
			CardID:   i + 1,
			Quantity: (i % 4) + 1, // 1-4 copies
		}
	}

	// Measure time
	start := time.Now()
	err := repo.UpsertMany(ctx, entries)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("failed to upsert large collection: %v", err)
	}

	// Verify count
	collection, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all: %v", err)
	}

	if len(collection) != numCards {
		t.Errorf("expected %d cards, got %d", numCards, len(collection))
	}

	// Should complete in reasonable time (< 1 second as per issue requirement)
	if elapsed > time.Second {
		t.Errorf("UpsertMany took too long: %v (should be < 1s)", elapsed)
	}

	t.Logf("UpsertMany of %d cards completed in %v", numCards, elapsed)
}

func TestCollectionRepository_GetChangesSince(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	now := time.Now()
	source := "pack"

	// Record changes at different times
	times := []time.Duration{
		-3 * time.Hour,    // Old change
		-2 * time.Hour,    // Old change
		-30 * time.Minute, // Recent change
		-10 * time.Minute, // Recent change
	}

	for i, offset := range times {
		timestamp := now.Add(offset)
		err := repo.RecordChange(ctx, 10000+i, 1, timestamp, &source)
		if err != nil {
			t.Fatalf("failed to record change %d: %v", i, err)
		}
	}

	// Get changes since 1 hour ago
	since := now.Add(-1 * time.Hour)
	changes, err := repo.GetChangesSince(ctx, since)
	if err != nil {
		t.Fatalf("failed to get changes since: %v", err)
	}

	// Should only return the 2 recent changes
	if len(changes) != 2 {
		t.Errorf("expected 2 recent changes, got %d", len(changes))
	}

	// Verify they are the correct cards (most recent first)
	if len(changes) >= 2 {
		if changes[0].CardID != 10003 {
			t.Errorf("expected first change to be card 10003, got %d", changes[0].CardID)
		}
		if changes[1].CardID != 10002 {
			t.Errorf("expected second change to be card 10002, got %d", changes[1].CardID)
		}
	}
}

func TestCollectionRepository_GetChangesSince_NoChanges(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	now := time.Now()
	source := "pack"

	// Record old changes only
	oldTime := now.Add(-2 * time.Hour)
	err := repo.RecordChange(ctx, 12345, 1, oldTime, &source)
	if err != nil {
		t.Fatalf("failed to record change: %v", err)
	}

	// Get changes since 1 hour ago (should be empty)
	since := now.Add(-1 * time.Hour)
	changes, err := repo.GetChangesSince(ctx, since)
	if err != nil {
		t.Fatalf("failed to get changes since: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestCollectionRepository_RecordHistoryEntry(t *testing.T) {
	db := setupCollectionTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewCollectionRepository(db)
	ctx := context.Background()

	// First, add a card to the collection
	err := repo.UpsertCard(ctx, 12345, 4)
	if err != nil {
		t.Fatalf("failed to upsert card: %v", err)
	}

	// Record a history entry without updating the collection
	now := time.Now()
	source := "sync"
	err = repo.RecordHistoryEntry(ctx, 12345, 2, 4, now, &source)
	if err != nil {
		t.Fatalf("failed to record history entry: %v", err)
	}

	// Verify collection was NOT updated (should still be 4)
	quantity, err := repo.GetCard(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get card: %v", err)
	}
	if quantity != 4 {
		t.Errorf("expected quantity to remain 4, got %d", quantity)
	}

	// Verify history was recorded
	history, err := repo.GetHistory(ctx, 12345)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(history))
	}

	if history[0].QuantityDelta != 2 {
		t.Errorf("expected delta 2, got %d", history[0].QuantityDelta)
	}
	if history[0].QuantityAfter != 4 {
		t.Errorf("expected quantity after 4, got %d", history[0].QuantityAfter)
	}
	if *history[0].Source != "sync" {
		t.Errorf("expected source 'sync', got '%s'", *history[0].Source)
	}
}

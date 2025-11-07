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

	schema := `
		CREATE TABLE collection (
			card_id INTEGER PRIMARY KEY,
			quantity INTEGER NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
	defer db.Close()

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

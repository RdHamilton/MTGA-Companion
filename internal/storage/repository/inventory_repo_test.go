package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// setupInventoryTestDB creates an in-memory database with inventory tables.
func setupInventoryTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE inventory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gold INTEGER NOT NULL DEFAULT 0,
			gems INTEGER NOT NULL DEFAULT 0,
			wc_common INTEGER NOT NULL DEFAULT 0,
			wc_uncommon INTEGER NOT NULL DEFAULT 0,
			wc_rare INTEGER NOT NULL DEFAULT 0,
			wc_mythic INTEGER NOT NULL DEFAULT 0,
			vault_progress REAL NOT NULL DEFAULT 0,
			draft_tokens INTEGER NOT NULL DEFAULT 0,
			sealed_tokens INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE inventory_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			field TEXT NOT NULL,
			previous_value INTEGER NOT NULL,
			new_value INTEGER NOT NULL,
			delta INTEGER NOT NULL,
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		INSERT INTO inventory (gold, gems, wc_common, wc_uncommon, wc_rare, wc_mythic, vault_progress, draft_tokens, sealed_tokens)
		VALUES (0, 0, 0, 0, 0, 0, 0, 0, 0);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestInventoryRepository_Get(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()

	// Get initial inventory (should be all zeros)
	inv, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get inventory: %v", err)
	}

	if inv.Gold != 0 {
		t.Errorf("expected gold 0, got %d", inv.Gold)
	}
	if inv.Gems != 0 {
		t.Errorf("expected gems 0, got %d", inv.Gems)
	}
	if inv.WCCommon != 0 {
		t.Errorf("expected common wildcards 0, got %d", inv.WCCommon)
	}
}

func TestInventoryRepository_Update(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// Update inventory
	newInv := &Inventory{
		Gold:          25000,
		Gems:          3500,
		WCCommon:      45,
		WCUncommon:    32,
		WCRare:        15,
		WCMythic:      8,
		VaultProgress: 75.5,
		DraftTokens:   2,
		SealedTokens:  1,
	}

	changes, err := repo.Update(ctx, newInv, &source)
	if err != nil {
		t.Fatalf("failed to update inventory: %v", err)
	}

	// Should have changes for all non-zero fields
	if len(changes) == 0 {
		t.Error("expected changes to be recorded")
	}

	// Verify inventory was updated
	inv, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("failed to get inventory: %v", err)
	}

	if inv.Gold != 25000 {
		t.Errorf("expected gold 25000, got %d", inv.Gold)
	}
	if inv.Gems != 3500 {
		t.Errorf("expected gems 3500, got %d", inv.Gems)
	}
	if inv.WCRare != 15 {
		t.Errorf("expected rare wildcards 15, got %d", inv.WCRare)
	}
	if inv.DraftTokens != 2 {
		t.Errorf("expected draft tokens 2, got %d", inv.DraftTokens)
	}
}

func TestInventoryRepository_Update_DetectsChanges(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// First update
	firstInv := &Inventory{
		Gold:     10000,
		Gems:     1000,
		WCRare:   5,
		WCMythic: 2,
	}
	_, _ = repo.Update(ctx, firstInv, &source)

	// Second update with changes
	secondInv := &Inventory{
		Gold:     15000, // +5000
		Gems:     1000,  // unchanged
		WCRare:   3,     // -2
		WCMythic: 4,     // +2
	}

	changes, err := repo.Update(ctx, secondInv, &source)
	if err != nil {
		t.Fatalf("failed to update inventory: %v", err)
	}

	// Should have exactly 3 changes (gold, wc_rare, wc_mythic)
	if len(changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(changes))
	}

	// Find gold change
	var goldChange *InventoryChange
	for i := range changes {
		if changes[i].Field == "gold" {
			goldChange = &changes[i]
			break
		}
	}

	if goldChange == nil {
		t.Fatal("expected gold change to be recorded")
	}

	if goldChange.PreviousValue != 10000 {
		t.Errorf("expected previous gold 10000, got %d", goldChange.PreviousValue)
	}
	if goldChange.NewValue != 15000 {
		t.Errorf("expected new gold 15000, got %d", goldChange.NewValue)
	}
	if goldChange.Delta != 5000 {
		t.Errorf("expected gold delta 5000, got %d", goldChange.Delta)
	}
}

func TestInventoryRepository_GetHistory(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// Make several updates to gold
	for gold := 1000; gold <= 5000; gold += 1000 {
		inv := &Inventory{Gold: gold}
		_, _ = repo.Update(ctx, inv, &source)
	}

	// Get gold history
	history, err := repo.GetHistory(ctx, "gold", 10)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 5 {
		t.Errorf("expected 5 history entries, got %d", len(history))
	}

	// Most recent should be last (5000)
	if history[0].NewValue != 5000 {
		t.Errorf("expected most recent to be 5000, got %d", history[0].NewValue)
	}
}

func TestInventoryRepository_GetRecentChanges(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// Update with multiple fields
	inv := &Inventory{
		Gold:     5000,
		Gems:     1000,
		WCRare:   10,
		WCMythic: 5,
	}
	_, _ = repo.Update(ctx, inv, &source)

	// Get recent changes
	changes, err := repo.GetRecentChanges(ctx, 10)
	if err != nil {
		t.Fatalf("failed to get recent changes: %v", err)
	}

	if len(changes) != 4 {
		t.Errorf("expected 4 changes, got %d", len(changes))
	}
}

func TestInventoryRepository_GetChangesSince(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// Update inventory
	inv := &Inventory{
		Gold: 5000,
		Gems: 1000,
	}
	_, _ = repo.Update(ctx, inv, &source)

	// Get changes since before the update
	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	changes, err := repo.GetChangesSince(ctx, oneMinuteAgo)
	if err != nil {
		t.Fatalf("failed to get changes since: %v", err)
	}

	if len(changes) != 2 {
		t.Errorf("expected 2 changes, got %d", len(changes))
	}

	// Get changes since the future (should be empty)
	future := time.Now().Add(1 * time.Hour)
	changes, err = repo.GetChangesSince(ctx, future)
	if err != nil {
		t.Fatalf("failed to get changes since: %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("expected 0 changes for future time, got %d", len(changes))
	}
}

func TestInventoryRepository_VaultProgressChange(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// First update with vault progress
	firstInv := &Inventory{
		VaultProgress: 50.0,
	}
	_, _ = repo.Update(ctx, firstInv, &source)

	// Second update with different vault progress
	secondInv := &Inventory{
		VaultProgress: 75.5,
	}
	changes, err := repo.Update(ctx, secondInv, &source)
	if err != nil {
		t.Fatalf("failed to update inventory: %v", err)
	}

	// Find vault progress change
	var vaultChange *InventoryChange
	for i := range changes {
		if changes[i].Field == "vault_progress" {
			vaultChange = &changes[i]
			break
		}
	}

	if vaultChange == nil {
		t.Fatal("expected vault progress change")
	}

	// Vault progress is stored as percentage * 100
	if vaultChange.PreviousValue != 5000 {
		t.Errorf("expected previous vault 5000, got %d", vaultChange.PreviousValue)
	}
	if vaultChange.NewValue != 7550 {
		t.Errorf("expected new vault 7550, got %d", vaultChange.NewValue)
	}
}

func TestInventoryRepository_NoChanges(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()
	source := "sync"

	// Update inventory
	inv := &Inventory{
		Gold: 5000,
		Gems: 1000,
	}
	_, _ = repo.Update(ctx, inv, &source)

	// Update with same values
	changes, err := repo.Update(ctx, inv, &source)
	if err != nil {
		t.Fatalf("failed to update inventory: %v", err)
	}

	// Should have no changes
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for same values, got %d", len(changes))
	}
}

func TestInventoryRepository_NilSource(t *testing.T) {
	db := setupInventoryTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewInventoryRepository(db)
	ctx := context.Background()

	// Update inventory without source
	inv := &Inventory{Gold: 5000}
	changes, err := repo.Update(ctx, inv, nil)
	if err != nil {
		t.Fatalf("failed to update inventory: %v", err)
	}

	// Should still record change
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Source != nil {
		t.Errorf("expected nil source, got %v", changes[0].Source)
	}
}

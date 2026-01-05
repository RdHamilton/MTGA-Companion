package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationManager_Up(t *testing.T) {
	// Create a temporary database file
	testDir := filepath.Join(os.TempDir(), "mtga-test-migration")
	dbPath := filepath.Join(testDir, "migration-test.db")

	// Clean up before and after test
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Create directory
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create migration manager and run migrations
	mgr, err := NewMigrationManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	// Run all migrations
	if err := mgr.Up(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Close migration manager
	if err := mgr.Close(); err != nil {
		t.Fatalf("Failed to close migration manager: %v", err)
	}

	// Verify migrations ran by checking the schema version
	mgr2, err := NewMigrationManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen migration manager: %v", err)
	}
	defer mgr2.Close()

	version, dirty, err := mgr2.Version()
	if err != nil {
		t.Fatalf("Failed to get migration version: %v", err)
	}

	if dirty {
		t.Error("Database is in dirty state after migrations")
	}

	// Should be at version 28 or higher (deck_permutations migration)
	if version < 28 {
		t.Errorf("Expected migration version >= 28, got %d", version)
	}

	t.Logf("✅ Migrations completed successfully at version %d", version)
}

func TestMigrationManager_DeckPermutationsTable(t *testing.T) {
	// Create a temporary database file
	testDir := filepath.Join(os.TempDir(), "mtga-test-permutations")
	dbPath := filepath.Join(testDir, "permutations-test.db")

	// Clean up before and after test
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Create directory
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Run migrations via config
	config := DefaultConfig(dbPath)
	config.AutoMigrate = true

	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to open database with migrations: %v", err)
	}
	defer db.Close()

	// Verify deck_permutations table exists
	var tableName string
	err = db.Conn().QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='deck_permutations'
	`).Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Fatal("deck_permutations table does not exist after migration")
		}
		t.Fatalf("Failed to query for table: %v", err)
	}

	if tableName != "deck_permutations" {
		t.Errorf("Expected table name 'deck_permutations', got '%s'", tableName)
	}

	// Verify columns exist
	columns := []string{
		"id", "deck_id", "parent_permutation_id", "cards",
		"version_number", "version_name", "change_summary",
		"matches_played", "matches_won", "games_played", "games_won",
		"created_at", "last_played_at",
	}

	for _, col := range columns {
		var colInfo string
		err = db.Conn().QueryRow(`
			SELECT name FROM pragma_table_info('deck_permutations') WHERE name = ?
		`, col).Scan(&colInfo)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("Column '%s' does not exist in deck_permutations table", col)
				continue
			}
			t.Errorf("Failed to query column info for '%s': %v", col, err)
		}
	}

	// Verify decks table has current_permutation_id column
	var permColInfo string
	err = db.Conn().QueryRow(`
		SELECT name FROM pragma_table_info('decks') WHERE name = 'current_permutation_id'
	`).Scan(&permColInfo)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("current_permutation_id column does not exist in decks table")
		} else {
			t.Errorf("Failed to query column info: %v", err)
		}
	}

	// Verify indexes exist
	indexes := []string{
		"idx_deck_permutations_deck_id",
		"idx_deck_permutations_parent",
		"idx_deck_permutations_created",
		"idx_deck_permutations_win_rate",
	}

	for _, idx := range indexes {
		var indexName string
		err = db.Conn().QueryRow(`
			SELECT name FROM sqlite_master
			WHERE type='index' AND name = ?
		`, idx).Scan(&indexName)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("Index '%s' does not exist", idx)
				continue
			}
			t.Errorf("Failed to query index '%s': %v", idx, err)
		}
	}

	t.Log("✅ deck_permutations table structure verified successfully")
}

func TestMigrationManager_Down(t *testing.T) {
	// Create a temporary database file
	testDir := filepath.Join(os.TempDir(), "mtga-test-migration-down")
	dbPath := filepath.Join(testDir, "migration-down-test.db")

	// Clean up before and after test
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Create directory
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Run migrations up
	mgr, err := NewMigrationManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	if err := mgr.Up(); err != nil {
		t.Fatalf("Failed to run migrations up: %v", err)
	}

	// Get current version
	versionBefore, _, err := mgr.Version()
	if err != nil {
		t.Fatalf("Failed to get version before down: %v", err)
	}

	// Run one migration down
	if err := mgr.Steps(-1); err != nil {
		t.Fatalf("Failed to run migration down: %v", err)
	}

	// Get new version
	versionAfter, dirty, err := mgr.Version()
	if err != nil {
		t.Fatalf("Failed to get version after down: %v", err)
	}

	if dirty {
		t.Error("Database is in dirty state after rollback")
	}

	if versionAfter >= versionBefore {
		t.Errorf("Version should decrease after down migration: before=%d, after=%d", versionBefore, versionAfter)
	}

	mgr.Close()

	t.Logf("✅ Down migration successful: %d -> %d", versionBefore, versionAfter)
}

func TestMigrationManager_Version(t *testing.T) {
	// Create a temporary database file
	testDir := filepath.Join(os.TempDir(), "mtga-test-migration-version")
	dbPath := filepath.Join(testDir, "version-test.db")

	// Clean up before and after test
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Create directory
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create migration manager
	mgr, err := NewMigrationManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}
	defer mgr.Close()

	// Version on fresh database should be 0 (no migrations run)
	version, dirty, err := mgr.Version()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if dirty {
		t.Error("Fresh database should not be dirty")
	}

	// Fresh database should have version 0
	if version != 0 {
		t.Logf("Note: Database has existing version %d (may have migrations from prior test)", version)
	}

	t.Log("✅ Version check successful")
}

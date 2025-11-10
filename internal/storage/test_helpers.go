package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestService creates a test service with a temporary database file.
func setupTestService(t *testing.T) *Service {
	t.Helper()

	// Create temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Run migrations first
	migrationMgr, err := NewMigrationManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	if err := migrationMgr.Up(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Close migration manager
	_ = migrationMgr.Close()

	// Open database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create service
	service := NewService(db)

	// Cleanup function
	t.Cleanup(func() {
		_ = db.Close()
		_ = os.Remove(dbPath)
	})

	return service
}

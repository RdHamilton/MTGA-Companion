package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_CreatesDirectory(t *testing.T) {
	// Test that Open creates parent directory if it doesn't exist
	testDir := filepath.Join(os.TempDir(), "mtga-test-db-creation")
	dbPath := filepath.Join(testDir, "test.db")
	
	// Clean up before test
	os.RemoveAll(testDir)
	
	// Verify directory doesn't exist
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Fatal("Test directory should not exist before test")
	}
	
	// Open database (should create directory)
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	defer os.RemoveAll(testDir)
	
	// Verify directory was created
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("Path is not a directory")
	}
	
	t.Log("✅ Directory creation successful!")
}

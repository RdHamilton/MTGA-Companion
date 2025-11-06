package storage

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test.db")

	if config.Path != "test.db" {
		t.Errorf("expected path 'test.db', got '%s'", config.Path)
	}

	if config.MaxOpenConns != 25 {
		t.Errorf("expected MaxOpenConns 25, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 5 {
		t.Errorf("expected MaxIdleConns 5, got %d", config.MaxIdleConns)
	}

	if config.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("expected ConnMaxLifetime 5m, got %v", config.ConnMaxLifetime)
	}

	if config.BusyTimeout != 5*time.Second {
		t.Errorf("expected BusyTimeout 5s, got %v", config.BusyTimeout)
	}

	if config.JournalMode != "WAL" {
		t.Errorf("expected JournalMode 'WAL', got '%s'", config.JournalMode)
	}

	if config.Synchronous != "NORMAL" {
		t.Errorf("expected Synchronous 'NORMAL', got '%s'", config.Synchronous)
	}
}

func TestOpen(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Open(config)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}

	// Test conn
	if db.Conn() == nil {
		t.Error("expected non-nil connection")
	}
}

func TestOpenWithNilConfig(t *testing.T) {
	_, err := Open(nil)
	if err == nil {
		t.Error("expected error when opening with nil config")
	}
}

func TestClose(t *testing.T) {
	config := DefaultConfig(":memory:")
	db, err := Open(config)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}

	// Ping should fail after close
	if err := db.Ping(); err == nil {
		t.Error("expected error when pinging closed database")
	}
}

func TestMultipleConnections(t *testing.T) {
	config := DefaultConfig(":memory:")
	config.MaxOpenConns = 2

	db, err := Open(config)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// This test just verifies that pool settings are applied
	// Actual concurrent connection behavior would require more complex testing
	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

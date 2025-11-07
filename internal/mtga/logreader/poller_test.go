package logreader

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPoller(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create empty log file
	if err := os.WriteFile(logPath, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	t.Run("ValidConfig", func(t *testing.T) {
		config := DefaultPollerConfig(logPath)
		poller, err := NewPoller(config)
		if err != nil {
			t.Fatalf("NewPoller() error = %v", err)
		}
		if poller == nil {
			t.Fatal("NewPoller() returned nil")
		}
		poller.Stop()
	})

	t.Run("NilConfig", func(t *testing.T) {
		_, err := NewPoller(nil)
		if err == nil {
			t.Error("NewPoller(nil) expected error, got nil")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		config := &PollerConfig{Path: ""}
		_, err := NewPoller(config)
		if err == nil {
			t.Error("NewPoller() with empty path expected error, got nil")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.log")
		config := DefaultPollerConfig(nonExistentPath)
		poller, err := NewPoller(config)
		if err != nil {
			t.Fatalf("NewPoller() with non-existent file error = %v", err)
		}
		if poller == nil {
			t.Fatal("NewPoller() returned nil")
		}
		poller.Stop()
	})
}

func TestPoller_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	testData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(testData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.Interval = 100 * time.Millisecond // Fast polling for tests
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	// Start poller
	updates := poller.Start()
	if !poller.IsRunning() {
		t.Error("Poller should be running after Start()")
	}

	// Wait a bit to ensure poller is running
	time.Sleep(150 * time.Millisecond)

	// Stop poller
	poller.Stop()

	// Wait for poller to stop
	time.Sleep(100 * time.Millisecond)

	if poller.IsRunning() {
		t.Error("Poller should not be running after Stop()")
	}

	// Verify updates channel is closed
	select {
	case _, ok := <-updates:
		if ok {
			t.Error("Updates channel should be closed after Stop()")
		}
	default:
		t.Error("Updates channel should be closed after Stop()")
	}
}

func TestPoller_ReadNewEntries(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.Interval = 100 * time.Millisecond
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Start poller
	updates := poller.Start()

	// Wait for initial position to be set
	time.Sleep(150 * time.Millisecond)

	// Append new entries
	newData := `[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2,"result":"win"}
[UnityCrossThreadLogger]{"type":"MatchResult","eventId":3}
`
	if err := os.WriteFile(logPath, append([]byte(initialData), []byte(newData)...), 0o644); err != nil {
		t.Fatalf("Failed to append to log file: %v", err)
	}

	// Wait for poller to detect changes
	time.Sleep(250 * time.Millisecond)

	// Collect new entries
	var receivedEntries []*LogEntry
	timeout := time.After(1 * time.Second)
	for {
		select {
		case entry, ok := <-updates:
			if !ok {
				// Channel closed
				goto done
			}
			receivedEntries = append(receivedEntries, entry)
		case <-timeout:
			goto done
		}
	}

done:
	// Should receive 2 new JSON entries
	if len(receivedEntries) != 2 {
		t.Errorf("Expected 2 new entries, got %d", len(receivedEntries))
	}

	// Verify entries
	if len(receivedEntries) > 0 {
		if receivedEntries[0].JSON["type"] != "GameEnd" {
			t.Errorf("First entry type = %v, want GameEnd", receivedEntries[0].JSON["type"])
		}
	}
	if len(receivedEntries) > 1 {
		if receivedEntries[1].JSON["type"] != "MatchResult" {
			t.Errorf("Second entry type = %v, want MatchResult", receivedEntries[1].JSON["type"])
		}
	}
}

func TestPoller_HandleLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file with some data
	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.Interval = 100 * time.Millisecond
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Start poller
	updates := poller.Start()

	// Wait for initial position to be set
	time.Sleep(150 * time.Millisecond)

	// Simulate log rotation by truncating and writing new data
	rotatedData := `[UnityCrossThreadLogger]{"type":"NewGame","eventId":10}
`
	if err := os.WriteFile(logPath, []byte(rotatedData), 0o644); err != nil {
		t.Fatalf("Failed to rotate log file: %v", err)
	}

	// Wait for poller to detect rotation
	time.Sleep(250 * time.Millisecond)

	// Collect new entries
	var receivedEntries []*LogEntry
	timeout := time.After(1 * time.Second)
	for {
		select {
		case entry, ok := <-updates:
			if !ok {
				goto done
			}
			receivedEntries = append(receivedEntries, entry)
		case <-timeout:
			goto done
		}
	}

done:
	// Should receive the new entry from rotated log
	if len(receivedEntries) != 1 {
		t.Errorf("Expected 1 new entry after rotation, got %d", len(receivedEntries))
	}

	if len(receivedEntries) > 0 {
		if receivedEntries[0].JSON["type"] != "NewGame" {
			t.Errorf("Entry type = %v, want NewGame", receivedEntries[0].JSON["type"])
		}
	}
}

func TestPoller_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	if err := os.WriteFile(logPath, []byte("test\n"), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.Interval = 100 * time.Millisecond
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Start poller
	_ = poller.Start()
	errChan := poller.Errors()

	// Wait for initial position
	time.Sleep(150 * time.Millisecond)

	// Remove file to trigger error
	if err := os.Remove(logPath); err != nil {
		t.Fatalf("Failed to remove log file: %v", err)
	}

	// Wait for poller to detect missing file
	time.Sleep(250 * time.Millisecond)

	// Check for errors (should handle gracefully)
	select {
	case err := <-errChan:
		if err != nil {
			t.Logf("Received expected error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		// No error is also acceptable (file not found is handled gracefully)
	}
}

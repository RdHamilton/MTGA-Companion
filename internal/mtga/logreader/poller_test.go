package logreader

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
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

// TestPoller_WithFileEvents tests the fsnotify-based event monitoring.
func TestPoller_WithFileEvents(t *testing.T) {
	// Skip if fsnotify is not available
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skipf("fsnotify not available: %v", err)
	}
	_ = watcher.Close()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.UseFileEvents = true // Enable file events
	config.Interval = 1 * time.Second
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Start poller
	updates := poller.Start()
	if !poller.IsRunning() {
		t.Error("Poller should be running after Start()")
	}

	// Wait for initial position to be set
	time.Sleep(200 * time.Millisecond)

	// Append new data - this should trigger a file event
	newData := `[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2,"result":"win"}
`
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	if _, err := file.WriteString(newData); err != nil {
		file.Close()
		t.Fatalf("Failed to append to log file: %v", err)
	}
	file.Close()

	// Wait for file event to be processed
	// Events should be faster than polling
	var receivedEntry *LogEntry
	select {
	case entry := <-updates:
		receivedEntry = entry
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for log entry via file events")
	}

	// Verify the entry
	if receivedEntry == nil {
		t.Fatal("Did not receive log entry")
	}
	if receivedEntry.JSON["type"] != "GameEnd" {
		t.Errorf("Entry type = %v, want GameEnd", receivedEntry.JSON["type"])
	}
}

// TestPoller_FileEventsDisabled tests that polling works when file events are disabled.
func TestPoller_FileEventsDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.UseFileEvents = false // Disable file events, use polling
	config.Interval = 200 * time.Millisecond
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Verify watcher is not initialized
	if poller.watcher != nil {
		t.Error("Watcher should be nil when file events are disabled")
	}

	// Start poller
	updates := poller.Start()

	// Wait for initial position
	time.Sleep(250 * time.Millisecond)

	// Append new data
	newData := `[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2}
`
	if err := os.WriteFile(logPath, append([]byte(initialData), []byte(newData)...), 0o644); err != nil {
		t.Fatalf("Failed to append to log file: %v", err)
	}

	// Wait for polling to detect changes
	var receivedEntry *LogEntry
	select {
	case entry := <-updates:
		receivedEntry = entry
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for log entry via polling")
	}

	// Verify the entry
	if receivedEntry == nil {
		t.Fatal("Did not receive log entry")
	}
	if receivedEntry.JSON["type"] != "GameEnd" {
		t.Errorf("Entry type = %v, want GameEnd", receivedEntry.JSON["type"])
	}
}

// TestPoller_FileRotationWithEvents tests file rotation detection with file events.
func TestPoller_FileRotationWithEvents(t *testing.T) {
	// Skip if fsnotify is not available
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skipf("fsnotify not available: %v", err)
	}
	_ = watcher.Close()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.UseFileEvents = true
	config.Interval = 500 * time.Millisecond
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Start poller
	updates := poller.Start()

	// Wait for initial position
	time.Sleep(200 * time.Millisecond)

	// Simulate log rotation: remove old file, create new one
	if err := os.Remove(logPath); err != nil {
		t.Fatalf("Failed to remove log file: %v", err)
	}

	// Wait for remove event to be processed
	time.Sleep(100 * time.Millisecond)

	// Create new log file (rotation complete)
	rotatedData := `[UnityCrossThreadLogger]{"type":"NewGame","eventId":10}
`
	if err := os.WriteFile(logPath, []byte(rotatedData), 0o644); err != nil {
		t.Fatalf("Failed to create rotated log file: %v", err)
	}

	// Wait for file event to be processed
	var receivedEntry *LogEntry
	select {
	case entry := <-updates:
		receivedEntry = entry
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for log entry after rotation")
	}

	// Verify the entry from rotated log
	if receivedEntry == nil {
		t.Fatal("Did not receive log entry after rotation")
	}
	if receivedEntry.JSON["type"] != "NewGame" {
		t.Errorf("Entry type = %v, want NewGame", receivedEntry.JSON["type"])
	}
}

// TestPoller_FallbackOnWatcherFailure tests that poller falls back to polling
// when watcher setup fails or file events are not available.
func TestPoller_FallbackOnWatcherFailure(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	config := DefaultPollerConfig(logPath)
	config.UseFileEvents = true
	config.Interval = 200 * time.Millisecond
	poller, err := NewPoller(config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	defer poller.Stop()

	// Start poller
	updates := poller.Start()
	errChan := poller.Errors()

	// Wait for initial setup
	time.Sleep(250 * time.Millisecond)

	// Check if we got a fallback error (may or may not happen depending on platform)
	select {
	case err := <-errChan:
		t.Logf("Received error (fallback scenario): %v", err)
	default:
		// No error is fine - watcher setup succeeded
	}

	// Append new data
	newData := `[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2}
`
	if err := os.WriteFile(logPath, append([]byte(initialData), []byte(newData)...), 0o644); err != nil {
		t.Fatalf("Failed to append to log file: %v", err)
	}

	// Should receive entry either via events or polling fallback
	var receivedEntry *LogEntry
	select {
	case entry := <-updates:
		receivedEntry = entry
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for log entry (events or polling)")
	}

	// Verify the entry
	if receivedEntry == nil {
		t.Fatal("Did not receive log entry")
	}
	if receivedEntry.JSON["type"] != "GameEnd" {
		t.Errorf("Entry type = %v, want GameEnd", receivedEntry.JSON["type"])
	}
}

// TestPoller_FileEventsPerformance compares event-based vs polling performance.
func TestPoller_FileEventsPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Skip if fsnotify is not available
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Skipf("fsnotify not available: %v", err)
	}
	_ = watcher.Close()

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create initial log file
	if err := os.WriteFile(logPath, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Test with file events
	t.Run("WithFileEvents", func(t *testing.T) {
		config := DefaultPollerConfig(logPath)
		config.UseFileEvents = true
		config.Interval = 100 * time.Millisecond
		poller, err := NewPoller(config)
		if err != nil {
			t.Fatalf("NewPoller() error = %v", err)
		}
		defer poller.Stop()

		updates := poller.Start()
		time.Sleep(100 * time.Millisecond)

		start := time.Now()
		// Append data
		newData := `[UnityCrossThreadLogger]{"type":"TestEvent","eventId":1}
`
		if err := os.WriteFile(logPath, []byte(newData), 0o644); err != nil {
			t.Fatalf("Failed to write to log file: %v", err)
		}

		// Wait for entry
		select {
		case <-updates:
			eventsLatency := time.Since(start)
			t.Logf("File events latency: %v", eventsLatency)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for entry with file events")
		}
	})

	// Reset file
	if err := os.WriteFile(logPath, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to reset log file: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Test with polling
	t.Run("WithPolling", func(t *testing.T) {
		config := DefaultPollerConfig(logPath)
		config.UseFileEvents = false
		config.Interval = 100 * time.Millisecond
		poller, err := NewPoller(config)
		if err != nil {
			t.Fatalf("NewPoller() error = %v", err)
		}
		defer poller.Stop()

		updates := poller.Start()
		time.Sleep(150 * time.Millisecond)

		start := time.Now()
		// Append data
		newData := `[UnityCrossThreadLogger]{"type":"TestEvent","eventId":1}
`
		if err := os.WriteFile(logPath, []byte(newData), 0o644); err != nil {
			t.Fatalf("Failed to write to log file: %v", err)
		}

		// Wait for entry
		select {
		case <-updates:
			pollingLatency := time.Since(start)
			t.Logf("Polling latency: %v", pollingLatency)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for entry with polling")
		}
	})
}

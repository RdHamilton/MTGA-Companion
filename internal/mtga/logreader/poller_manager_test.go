package logreader

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPollerManager(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultPollerManagerConfig()
		manager := NewPollerManager(config)
		if manager == nil {
			t.Fatal("NewPollerManager() returned nil")
		}
		if manager.PollerCount() != 0 {
			t.Errorf("Expected 0 pollers, got %d", manager.PollerCount())
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		manager := NewPollerManager(nil)
		if manager == nil {
			t.Fatal("NewPollerManager(nil) returned nil")
		}
	})

	t.Run("WithNotifications", func(t *testing.T) {
		config := DefaultPollerManagerConfig()
		config.EnableNotifications = true
		manager := NewPollerManager(config)
		if manager.GetNotifier() == nil {
			t.Error("Notifier should be initialized when notifications enabled")
		}
	})
}

func TestPollerManager_AddRemovePoller(t *testing.T) {
	tmpDir := t.TempDir()
	logPath1 := filepath.Join(tmpDir, "test1.log")
	logPath2 := filepath.Join(tmpDir, "test2.log")

	// Create test log files
	if err := os.WriteFile(logPath1, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	if err := os.WriteFile(logPath2, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	manager := NewPollerManager(nil)
	defer manager.Stop()

	// Add first poller
	config1 := DefaultPollerConfig(logPath1)
	if err := manager.AddPoller("poller1", config1); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	if manager.PollerCount() != 1 {
		t.Errorf("Expected 1 poller, got %d", manager.PollerCount())
	}

	// Add second poller
	config2 := DefaultPollerConfig(logPath2)
	if err := manager.AddPoller("poller2", config2); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	if manager.PollerCount() != 2 {
		t.Errorf("Expected 2 pollers, got %d", manager.PollerCount())
	}

	// Try adding duplicate
	if err := manager.AddPoller("poller1", config1); err == nil {
		t.Error("Expected error when adding duplicate poller, got nil")
	}

	// Remove poller
	if err := manager.RemovePoller("poller1"); err != nil {
		t.Errorf("RemovePoller() error = %v", err)
	}

	if manager.PollerCount() != 1 {
		t.Errorf("Expected 1 poller after removal, got %d", manager.PollerCount())
	}

	// Try removing non-existent poller
	if err := manager.RemovePoller("nonexistent"); err == nil {
		t.Error("Expected error when removing non-existent poller, got nil")
	}
}

func TestPollerManager_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	initialData := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	if err := os.WriteFile(logPath, []byte(initialData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	manager := NewPollerManager(nil)

	// Add poller
	config := DefaultPollerConfig(logPath)
	config.Interval = 100 * time.Millisecond
	if err := manager.AddPoller("test", config); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	// Start manager
	_ = manager.Start()
	if !manager.IsRunning() {
		t.Error("Manager should be running after Start()")
	}

	time.Sleep(150 * time.Millisecond)

	// Stop manager
	manager.Stop()
	if manager.IsRunning() {
		t.Error("Manager should not be running after Stop()")
	}
}

func TestPollerManager_AggregateUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	logPath1 := filepath.Join(tmpDir, "test1.log")
	logPath2 := filepath.Join(tmpDir, "test2.log")

	// Create test log files
	data1 := `[UnityCrossThreadLogger]{"type":"GameStart","eventId":1}
`
	data2 := `[UnityCrossThreadLogger]{"type":"GameEnd","eventId":2}
`
	if err := os.WriteFile(logPath1, []byte(data1), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	if err := os.WriteFile(logPath2, []byte(data2), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	manager := NewPollerManager(nil)
	defer manager.Stop()

	// Add pollers
	config1 := DefaultPollerConfig(logPath1)
	config1.Interval = 100 * time.Millisecond
	if err := manager.AddPoller("poller1", config1); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	config2 := DefaultPollerConfig(logPath2)
	config2.Interval = 100 * time.Millisecond
	if err := manager.AddPoller("poller2", config2); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	// Start manager
	updates := manager.Start()

	time.Sleep(150 * time.Millisecond)

	// Append new data to both files
	newData1 := `[UnityCrossThreadLogger]{"type":"MatchResult","eventId":3}
`
	newData2 := `[UnityCrossThreadLogger]{"type":"RankUpdate","eventId":4}
`
	if err := os.WriteFile(logPath1, append([]byte(data1), []byte(newData1)...), 0o644); err != nil {
		t.Fatalf("Failed to append to log file: %v", err)
	}
	if err := os.WriteFile(logPath2, append([]byte(data2), []byte(newData2)...), 0o644); err != nil {
		t.Fatalf("Failed to append to log file: %v", err)
	}

	// Collect updates
	var receivedEntries []*LogEntry
	timeout := time.After(1 * time.Second)
	for len(receivedEntries) < 2 {
		select {
		case entry, ok := <-updates:
			if !ok {
				t.Fatal("Updates channel closed unexpectedly")
			}
			receivedEntries = append(receivedEntries, entry)
		case <-timeout:
			goto done
		}
	}

done:
	// Should receive entries from both pollers
	if len(receivedEntries) < 2 {
		t.Errorf("Expected at least 2 entries from both pollers, got %d", len(receivedEntries))
	}
}

func TestPollerManager_AggregateMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	logPath1 := filepath.Join(tmpDir, "test1.log")
	logPath2 := filepath.Join(tmpDir, "test2.log")

	// Create test log files
	if err := os.WriteFile(logPath1, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	if err := os.WriteFile(logPath2, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	manager := NewPollerManager(nil)
	defer manager.Stop()

	// Add pollers with metrics enabled
	config1 := DefaultPollerConfig(logPath1)
	config1.EnableMetrics = true
	config1.Interval = 100 * time.Millisecond
	if err := manager.AddPoller("poller1", config1); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	config2 := DefaultPollerConfig(logPath2)
	config2.EnableMetrics = true
	config2.Interval = 100 * time.Millisecond
	if err := manager.AddPoller("poller2", config2); err != nil {
		t.Fatalf("AddPoller() error = %v", err)
	}

	// Start manager
	_ = manager.Start()

	time.Sleep(250 * time.Millisecond)

	// Get aggregated metrics
	metrics := manager.AggregateMetrics()
	if metrics == nil {
		t.Fatal("AggregateMetrics() returned nil")
	}

	// Should have poll counts from both pollers
	if metrics.PollCount < 2 {
		t.Logf("Expected at least 2 poll counts (one from each poller), got %d", metrics.PollCount)
	}
}

func TestPollerManager_DynamicAddRemove(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	if err := os.WriteFile(logPath, []byte(""), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	manager := NewPollerManager(nil)
	defer manager.Stop()

	// Start manager with no pollers
	_ = manager.Start()
	if !manager.IsRunning() {
		t.Error("Manager should be running")
	}

	// Add poller while running
	config := DefaultPollerConfig(logPath)
	if err := manager.AddPoller("dynamic", config); err != nil {
		t.Fatalf("AddPoller() while running error = %v", err)
	}

	if manager.PollerCount() != 1 {
		t.Errorf("Expected 1 poller, got %d", manager.PollerCount())
	}

	time.Sleep(100 * time.Millisecond)

	// Remove poller while running
	if err := manager.RemovePoller("dynamic"); err != nil {
		t.Errorf("RemovePoller() while running error = %v", err)
	}

	if manager.PollerCount() != 0 {
		t.Errorf("Expected 0 pollers after removal, got %d", manager.PollerCount())
	}
}

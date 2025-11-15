package daemon

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if config.Port != 9999 {
		t.Errorf("Expected default port 9999, got %d", config.Port)
	}

	if config.PollInterval != 2*time.Second {
		t.Errorf("Expected poll interval 2s, got %v", config.PollInterval)
	}

	if config.UseFSNotify {
		t.Error("Expected UseFSNotify to be false by default")
	}

	if config.EnableMetrics {
		t.Error("Expected EnableMetrics to be false by default")
	}
}

func TestConfig_CustomValues(t *testing.T) {
	config := &Config{
		Port:          8080,
		DBPath:        "/tmp/test.db",
		LogPath:       "/tmp/test.log",
		PollInterval:  5 * time.Second,
		UseFSNotify:   true,
		EnableMetrics: true,
	}

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	if config.DBPath != "/tmp/test.db" {
		t.Errorf("Expected DBPath /tmp/test.db, got %s", config.DBPath)
	}

	if config.LogPath != "/tmp/test.log" {
		t.Errorf("Expected LogPath /tmp/test.log, got %s", config.LogPath)
	}

	if config.PollInterval != 5*time.Second {
		t.Errorf("Expected poll interval 5s, got %v", config.PollInterval)
	}

	if !config.UseFSNotify {
		t.Error("Expected UseFSNotify to be true")
	}

	if !config.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}
}

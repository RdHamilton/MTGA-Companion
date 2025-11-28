package daemon

import (
	"os"
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

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	if !config.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be true by default")
	}

	if config.AllowedOrigins != nil {
		t.Error("Expected AllowedOrigins to be nil by default")
	}
}

func TestProductionCORSConfig(t *testing.T) {
	origins := []string{"https://example.com", "https://app.example.com"}
	config := ProductionCORSConfig(origins)

	if config.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be false for production config")
	}

	if len(config.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(config.AllowedOrigins))
	}

	if config.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("Expected first origin https://example.com, got %s", config.AllowedOrigins[0])
	}
}

func TestCORSConfigFromEnv_Default(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("MTGA_CORS_ALLOW_ALL")
	os.Unsetenv("MTGA_CORS_ORIGINS")

	config := CORSConfigFromEnv()

	if !config.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be true when no env vars set")
	}
}

func TestCORSConfigFromEnv_WithOrigins(t *testing.T) {
	os.Setenv("MTGA_CORS_ORIGINS", "https://example.com, https://app.example.com")
	defer os.Unsetenv("MTGA_CORS_ORIGINS")
	os.Unsetenv("MTGA_CORS_ALLOW_ALL")

	config := CORSConfigFromEnv()

	if config.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be false when origins are specified")
	}

	if len(config.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(config.AllowedOrigins))
	}

	if config.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("Expected first origin https://example.com, got %s", config.AllowedOrigins[0])
	}

	if config.AllowedOrigins[1] != "https://app.example.com" {
		t.Errorf("Expected second origin https://app.example.com, got %s", config.AllowedOrigins[1])
	}
}

func TestCORSConfigFromEnv_ExplicitAllowAll(t *testing.T) {
	os.Setenv("MTGA_CORS_ORIGINS", "https://example.com")
	os.Setenv("MTGA_CORS_ALLOW_ALL", "true")
	defer os.Unsetenv("MTGA_CORS_ORIGINS")
	defer os.Unsetenv("MTGA_CORS_ALLOW_ALL")

	config := CORSConfigFromEnv()

	// Explicit AllowAllOrigins=true should override having origins set
	if !config.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be true when explicitly set")
	}

	// Origins should still be parsed
	if len(config.AllowedOrigins) != 1 {
		t.Errorf("Expected 1 allowed origin, got %d", len(config.AllowedOrigins))
	}
}

func TestCORSConfigFromEnv_ExplicitDisableAllowAll(t *testing.T) {
	os.Unsetenv("MTGA_CORS_ORIGINS")
	os.Setenv("MTGA_CORS_ALLOW_ALL", "false")
	defer os.Unsetenv("MTGA_CORS_ALLOW_ALL")

	config := CORSConfigFromEnv()

	if config.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be false when explicitly set to false")
	}
}

func TestDefaultConfig_IncludesCORSConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.CORSConfig.AllowAllOrigins {
		t.Error("Expected default config to have AllowAllOrigins=true")
	}
}

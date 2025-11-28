package daemon

import (
	"os"
	"strings"
	"time"
)

// Config holds configuration for the daemon service.
type Config struct {
	// WebSocket server port
	Port int

	// Database path
	DBPath string

	// MTGA log file path (auto-detect if empty)
	LogPath string

	// Log poll interval
	PollInterval time.Duration

	// Enable file system events (fsnotify) for log watching
	UseFSNotify bool

	// Enable metrics
	EnableMetrics bool

	// Enable automatic log archival (Phase 4)
	EnableArchival bool

	// Directory path for archived logs (defaults to ~/.mtga-companion/archived_logs)
	ArchiveDir string

	// Interval for archiving active Player.log (defaults to 5 minutes)
	ArchiveInterval time.Duration

	// CORS configuration for WebSocket server
	CORSConfig CORSConfig
}

// CORSConfig holds CORS settings for the WebSocket server.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins for CORS.
	// If empty, all origins are allowed (development mode).
	// Use "*" to explicitly allow all origins in production.
	AllowedOrigins []string

	// AllowAllOrigins allows all origins regardless of AllowedOrigins list.
	// Defaults to true for development compatibility.
	AllowAllOrigins bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:            9999,
		DBPath:          "", // Will use default from storage
		LogPath:         "", // Will auto-detect
		PollInterval:    2 * time.Second,
		UseFSNotify:     false,
		EnableMetrics:   false,
		EnableArchival:  false,           // Disabled by default (opt-in)
		ArchiveDir:      "",              // Will use default ~/.mtga-companion/archived_logs
		ArchiveInterval: 5 * time.Minute, // Archive Player.log every 5 minutes
		CORSConfig: CORSConfig{
			AllowAllOrigins: true, // Allow all origins by default for development
			AllowedOrigins:  nil,  // Empty means use AllowAllOrigins setting
		},
	}
}

// DefaultCORSConfig returns a default CORS configuration for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowAllOrigins: true,
		AllowedOrigins:  nil,
	}
}

// ProductionCORSConfig returns a restrictive CORS configuration for production.
// Origins should be set via environment variable or config file.
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  allowedOrigins,
	}
}

// CORSConfigFromEnv creates a CORS configuration from environment variables.
// Environment variables:
//   - MTGA_CORS_ALLOW_ALL: Set to "true" to allow all origins (default: true)
//   - MTGA_CORS_ORIGINS: Comma-separated list of allowed origins
//
// If MTGA_CORS_ORIGINS is set, MTGA_CORS_ALLOW_ALL defaults to false.
func CORSConfigFromEnv() CORSConfig {
	config := DefaultCORSConfig()

	// Check for allowed origins first
	if origins := os.Getenv("MTGA_CORS_ORIGINS"); origins != "" {
		originList := strings.Split(origins, ",")
		for i, origin := range originList {
			originList[i] = strings.TrimSpace(origin)
		}
		config.AllowedOrigins = originList
		// If specific origins are set, default to not allowing all
		config.AllowAllOrigins = false
	}

	// Explicit allow-all setting takes precedence
	if allowAll := os.Getenv("MTGA_CORS_ALLOW_ALL"); allowAll != "" {
		config.AllowAllOrigins = strings.ToLower(allowAll) == "true"
	}

	return config
}

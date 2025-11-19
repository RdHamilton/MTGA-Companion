package daemon

import "time"

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
	}
}

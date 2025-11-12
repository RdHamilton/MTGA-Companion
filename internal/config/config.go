package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration.
type Config struct {
	// Log file configuration
	Log LogConfig `toml:"log"`

	// Cache configuration
	Cache CacheConfig `toml:"cache"`

	// Draft overlay configuration
	Overlay OverlayConfig `toml:"overlay"`

	// Application configuration
	App AppConfig `toml:"app"`
}

// LogConfig contains log file monitoring settings.
type LogConfig struct {
	FilePath      string `toml:"file_path"`      // Path to MTGA Player.log
	PollInterval  string `toml:"poll_interval"`  // Polling interval (e.g., "2s")
	UseFsnotify   bool   `toml:"use_fsnotify"`   // Use file system events
	EnableMetrics bool   `toml:"enable_metrics"` // Enable performance metrics
}

// CacheConfig contains caching settings.
type CacheConfig struct {
	Enabled bool   `toml:"enabled"`  // Enable caching
	TTL     string `toml:"ttl"`      // Cache TTL (e.g., "24h")
	MaxSize int    `toml:"max_size"` // Max cache entries (0 = unlimited)
}

// OverlayConfig contains draft overlay settings.
type OverlayConfig struct {
	SetFile     string `toml:"set_file"`     // Path to set file
	SetCode     string `toml:"set_code"`     // Set code (e.g., "BLB")
	Format      string `toml:"format"`       // Draft format
	Resume      bool   `toml:"resume"`       // Resume active drafts
	LookbackHrs int    `toml:"lookback_hrs"` // Hours to look back
}

// AppConfig contains general application settings.
type AppConfig struct {
	DebugMode bool `toml:"debug_mode"` // Enable debug logging
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Log: LogConfig{
			FilePath:      "",
			PollInterval:  "2s",
			UseFsnotify:   true,
			EnableMetrics: false,
		},
		Cache: CacheConfig{
			Enabled: true,
			TTL:     "24h",
			MaxSize: 0,
		},
		Overlay: OverlayConfig{
			SetFile:     "",
			SetCode:     "",
			Format:      "PremierDraft",
			Resume:      true,
			LookbackHrs: 24,
		},
		App: AppConfig{
			DebugMode: false,
		},
	}
}

// configPath returns the path to the configuration file.
func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".mtga-companion")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.toml"), nil
}

// Load loads the configuration from disk. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Parse TOML
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to disk.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Marshal to TOML
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration values.
func (c *Config) Validate() error {
	// Validate poll interval
	if _, err := time.ParseDuration(c.Log.PollInterval); err != nil {
		return fmt.Errorf("invalid poll interval %q: %w", c.Log.PollInterval, err)
	}

	// Validate cache TTL
	if _, err := time.ParseDuration(c.Cache.TTL); err != nil {
		return fmt.Errorf("invalid cache TTL %q: %w", c.Cache.TTL, err)
	}

	// Validate cache max size
	if c.Cache.MaxSize < 0 {
		return fmt.Errorf("cache max size cannot be negative: %d", c.Cache.MaxSize)
	}

	// Validate lookback hours
	if c.Overlay.LookbackHrs < 0 {
		return fmt.Errorf("lookback hours cannot be negative: %d", c.Overlay.LookbackHrs)
	}

	return nil
}

// GetLogPollInterval returns the log poll interval as a duration.
func (c *Config) GetLogPollInterval() (time.Duration, error) {
	return time.ParseDuration(c.Log.PollInterval)
}

// GetCacheTTL returns the cache TTL as a duration.
func (c *Config) GetCacheTTL() (time.Duration, error) {
	return time.ParseDuration(c.Cache.TTL)
}

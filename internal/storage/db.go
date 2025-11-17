// Package storage provides database access and persistence for MTGA data.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// DB wraps the database connection and provides access to repositories.
type DB struct {
	conn *sql.DB
}

// Config holds database configuration settings.
type Config struct {
	// Path is the file path to the SQLite database.
	// Use ":memory:" for an in-memory database (useful for testing).
	Path string

	// MaxOpenConns sets the maximum number of open connections to the database.
	// Default: 25
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections in the pool.
	// Default: 5
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum amount of time a connection may be reused.
	// Default: 5 minutes
	ConnMaxLifetime time.Duration

	// BusyTimeout sets how long to wait when the database is locked.
	// Default: 5 seconds
	BusyTimeout time.Duration

	// JournalMode sets the SQLite journal mode.
	// Options: DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF
	// Default: WAL (Write-Ahead Logging) for better concurrency
	JournalMode string

	// Synchronous sets the SQLite synchronous mode.
	// Options: OFF, NORMAL, FULL, EXTRA
	// Default: NORMAL for good balance of safety and performance
	Synchronous string

	// AutoMigrate automatically runs pending database migrations on Open.
	// Default: false (migrations must be run manually)
	AutoMigrate bool
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig(path string) *Config {
	return &Config{
		Path:            path,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		BusyTimeout:     5 * time.Second,
		JournalMode:     "WAL",
		Synchronous:     "NORMAL",
	}
}

// Open creates a new database connection with the given configuration.
// It configures connection pooling and SQLite-specific settings.
func Open(config *Config) (*DB, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create parent directory if it doesn't exist (unless using in-memory database)
	if config.Path != ":memory:" {
		dir := filepath.Dir(config.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Build DSN with pragma parameters
	dsn := fmt.Sprintf("%s?_busy_timeout=%d&_journal_mode=%s&_synchronous=%s&_foreign_keys=on",
		config.Path,
		config.BusyTimeout.Milliseconds(),
		config.JournalMode,
		config.Synchronous,
	)

	// Open database connection
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(config.MaxOpenConns)
	conn.SetMaxIdleConns(config.MaxIdleConns)
	conn.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Verify connection
	if err := conn.Ping(); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to close database after ping error: %w (original error: %v)", closeErr, err)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations if auto-migrate is enabled
	if config.AutoMigrate {
		// Close the connection temporarily for migration
		if err := conn.Close(); err != nil {
			return nil, fmt.Errorf("failed to close database for migration: %w", err)
		}

		// Run migrations
		mgr, err := NewMigrationManager(config.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to create migration manager: %w", err)
		}

		if err := mgr.Up(); err != nil {
			if closeErr := mgr.Close(); closeErr != nil {
				return nil, fmt.Errorf("failed to close migration manager after error: %w (original error: %v)", closeErr, err)
			}
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}

		if err := mgr.Close(); err != nil {
			return nil, fmt.Errorf("failed to close migration manager: %w", err)
		}

		// Reopen the connection
		conn, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to reopen database after migrations: %w", err)
		}

		// Reconfigure connection pool
		conn.SetMaxOpenConns(config.MaxOpenConns)
		conn.SetMaxIdleConns(config.MaxIdleConns)
		conn.SetConnMaxLifetime(config.ConnMaxLifetime)

		// Verify connection again
		if err := conn.Ping(); err != nil {
			if closeErr := conn.Close(); closeErr != nil {
				return nil, fmt.Errorf("failed to close database after ping error: %w (original error: %v)", closeErr, err)
			}
			return nil, fmt.Errorf("failed to ping database after migrations: %w", err)
		}
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

// Conn returns the underlying sql.DB connection.
// This is useful for raw SQL queries or custom operations.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Ping verifies the database connection is alive.
func (db *DB) Ping() error {
	return db.conn.Ping()
}

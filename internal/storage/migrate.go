package storage

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationManager handles database schema migrations.
type MigrationManager struct {
	migrate *migrate.Migrate
}

// NewMigrationManager creates a new migration manager.
// The dbPath should be a file path to the SQLite database.
func NewMigrationManager(dbPath string) (*MigrationManager, error) {
	// Create io/fs from embedded files
	migrationsDir, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to access migrations directory: %w", err)
	}

	// Create source driver from embedded filesystem
	sourceDriver, err := iofs.New(migrationsDir, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create database URL
	// Convert Windows backslashes to forward slashes and ensure absolute paths have leading slash
	normalizedPath := filepath.ToSlash(dbPath)
	if filepath.IsAbs(dbPath) && normalizedPath[0] != '/' {
		normalizedPath = "/" + normalizedPath
	}
	databaseURL := fmt.Sprintf("sqlite://%s", normalizedPath)

	// Create migrate instance
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return &MigrationManager{migrate: m}, nil
}

// Up applies all pending migrations.
func (mm *MigrationManager) Up() error {
	err := mm.migrate.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// Down rolls back the last migration.
func (mm *MigrationManager) Down() error {
	err := mm.migrate.Down()
	if err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}

// Steps applies n migrations. Positive n applies up migrations, negative applies down.
func (mm *MigrationManager) Steps(n int) error {
	err := mm.migrate.Steps(n)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate %d steps: %w", n, err)
	}
	return nil
}

// Version returns the current migration version and dirty state.
func (mm *MigrationManager) Version() (version uint, dirty bool, err error) {
	version, dirty, err = mm.migrate.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Goto migrates to a specific version.
func (mm *MigrationManager) Goto(version uint) error {
	err := mm.migrate.Migrate(version)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}
	return nil
}

// Force sets the migration version without running migrations.
// Use with caution - this is for recovering from failed migrations.
func (mm *MigrationManager) Force(version int) error {
	err := mm.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}
	return nil
}

// Close closes the migration manager and releases resources.
func (mm *MigrationManager) Close() error {
	srcErr, dbErr := mm.migrate.Close()
	if srcErr != nil {
		return fmt.Errorf("failed to close source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}

package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupManager handles database backup and restore operations.
type BackupManager struct {
	dbPath string
}

// NewBackupManager creates a new backup manager for the given database path.
func NewBackupManager(dbPath string) *BackupManager {
	return &BackupManager{
		dbPath: dbPath,
	}
}

// BackupConfig holds configuration for backup operations.
type BackupConfig struct {
	// BackupDir is the directory where backups will be stored.
	// If empty, defaults to a "backups" subdirectory in the database directory.
	BackupDir string

	// BackupName is the name of the backup file (without extension).
	// If empty, a timestamp-based name will be generated.
	BackupName string

	// VerifyBackup indicates whether to verify the backup after creation.
	VerifyBackup bool
}

// DefaultBackupConfig returns a BackupConfig with sensible defaults.
func DefaultBackupConfig() *BackupConfig {
	return &BackupConfig{
		BackupDir:    "",
		BackupName:   "",
		VerifyBackup: true,
	}
}

// Backup creates a backup of the database.
// For SQLite, this uses the VACUUM INTO command which is atomic and doesn't require exclusive locks.
func (bm *BackupManager) Backup(config *BackupConfig) (string, error) {
	if config == nil {
		config = DefaultBackupConfig()
	}

	// Determine backup directory
	backupDir := config.BackupDir
	if backupDir == "" {
		dbDir := filepath.Dir(bm.dbPath)
		backupDir = filepath.Join(dbDir, "backups")
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Determine backup filename
	backupName := config.BackupName
	if backupName == "" {
		timestamp := time.Now().Format("20060102_150405")
		backupName = fmt.Sprintf("backup_%s", timestamp)
	}
	backupPath := filepath.Join(backupDir, backupName+".db")

	// Open source database
	sourceDB, err := sql.Open("sqlite", bm.dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer func() {
		if closeErr := sourceDB.Close(); closeErr != nil {
			// Log error but don\'t fail backup

			_ = closeErr
		}
	}()

	// Use VACUUM INTO for atomic backup (SQLite 3.27+)
	// This creates a complete copy of the database without requiring exclusive locks
	vacuumSQL := fmt.Sprintf("VACUUM INTO %q", backupPath)
	if _, err := sourceDB.Exec(vacuumSQL); err != nil {
		// Fallback to file copy if VACUUM INTO is not supported
		return bm.backupByCopy(backupPath)
	}

	// Verify backup if requested
	if config.VerifyBackup {
		if err := bm.VerifyBackup(backupPath); err != nil {
			// Remove invalid backup
			_ = os.Remove(backupPath)
			return "", fmt.Errorf("backup verification failed: %w", err)
		}
	}

	return backupPath, nil
}

// backupByCopy creates a backup by copying the database file.
// This is a fallback method if VACUUM INTO is not available.
func (bm *BackupManager) backupByCopy(backupPath string) (string, error) {
	sourceFile, err := os.Open(bm.dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database file: %w", err)
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		_ = os.Remove(backupPath)
		return "", fmt.Errorf("failed to copy database file: %w", err)
	}

	return backupPath, nil
}

// Restore restores the database from a backup file.
// WARNING: This will overwrite the current database.
func (bm *BackupManager) Restore(backupPath string) error {
	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Verify backup integrity
	if err := bm.VerifyBackup(backupPath); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// Close any existing connections to the database
	// (This is handled by the caller closing the DB connection)

	// Create a temporary file for the restore
	tempPath := bm.dbPath + ".restore.tmp"

	// Copy backup to temporary location
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	destFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary restore file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to copy backup file: %w", err)
	}

	// Close files before rename
	if err := sourceFile.Close(); err != nil {
		_ = err
	}
	if err := destFile.Close(); err != nil {
		_ = err
	}

	// Verify the restored database
	if err := bm.VerifyBackup(tempPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("restored database verification failed: %w", err)
	}

	// Backup the current database (if it exists)
	if _, err := os.Stat(bm.dbPath); err == nil {
		oldBackupPath := bm.dbPath + ".old." + time.Now().Format("20060102_150405")
		if err := os.Rename(bm.dbPath, oldBackupPath); err != nil {
			_ = os.Remove(tempPath)
			return fmt.Errorf("failed to backup current database: %w", err)
		}
	}

	// Replace current database with restored backup
	if err := os.Rename(tempPath, bm.dbPath); err != nil {
		return fmt.Errorf("failed to replace database with restored backup: %w", err)
	}

	return nil
}

// VerifyBackup verifies that a backup file is a valid SQLite database.
func (bm *BackupManager) VerifyBackup(backupPath string) error {
	// Try to open the backup as a SQLite database
	db, err := sql.Open("sqlite", backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup as database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	// Verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping backup database: %w", err)
	}

	// Verify we can query the database
	var version string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
		return fmt.Errorf("failed to query backup database: %w", err)
	}

	return nil
}

// ListBackups returns a list of all backup files in the backup directory.
func (bm *BackupManager) ListBackups(backupDir string) ([]BackupInfo, error) {
	if backupDir == "" {
		dbDir := filepath.Dir(bm.dbPath)
		backupDir = filepath.Join(dbDir, "backups")
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only include .db files
		if filepath.Ext(entry.Name()) != ".db" {
			continue
		}

		backupPath := filepath.Join(backupDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Get file size
		size := info.Size()

		// Try to get modification time
		modTime := info.ModTime()

		// Calculate checksum
		checksum, err := calculateChecksum(backupPath)
		if err != nil {
			checksum = "unknown"
		}

		backups = append(backups, BackupInfo{
			Path:     backupPath,
			Name:     entry.Name(),
			Size:     size,
			ModTime:  modTime,
			Checksum: checksum,
		})
	}

	return backups, nil
}

// BackupInfo contains information about a backup file.
type BackupInfo struct {
	Path     string
	Name     string
	Size     int64
	ModTime  time.Time
	Checksum string
}

// calculateChecksum calculates the SHA-256 checksum of a file.
func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetBackupDir returns the default backup directory path.
func (bm *BackupManager) GetBackupDir() string {
	dbDir := filepath.Dir(bm.dbPath)
	return filepath.Join(dbDir, "backups")
}

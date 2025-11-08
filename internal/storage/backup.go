package storage

import (
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

	// Compress enables gzip compression for backups.
	// Default: false
	Compress bool

	// RetentionDays specifies how many days to keep backups.
	// Backups older than this will be automatically cleaned up.
	// 0 means no automatic cleanup.
	// Default: 0 (no cleanup)
	RetentionDays int

	// MaxBackups specifies the maximum number of backups to keep.
	// When this limit is exceeded, oldest backups are removed.
	// 0 means no limit.
	// Default: 0 (no limit)
	MaxBackups int

	// Encrypt enables encryption for backups using AES-256-GCM.
	// Requires EncryptionPassword to be set.
	// Default: false
	Encrypt bool

	// EncryptionPassword is the password/passphrase used for encryption.
	// Only used if Encrypt is true.
	EncryptionPassword string

	// EncryptionConfig provides advanced encryption settings (Argon2 parameters).
	// If nil and Encrypt is true, default settings will be used.
	EncryptionConfig *EncryptionConfig
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

	// Encrypt backup if requested
	if config.Encrypt {
		if config.EncryptionPassword == "" {
			_ = os.Remove(backupPath)
			return "", fmt.Errorf("encryption enabled but no password provided")
		}

		encryptedPath, err := bm.encryptBackup(backupPath, config)
		if err != nil {
			// Keep unencrypted backup if encryption fails
			return backupPath, fmt.Errorf("backup created but encryption failed: %w", err)
		}
		// Remove unencrypted backup
		_ = os.Remove(backupPath)
		backupPath = encryptedPath
	}

	// Compress backup if requested (after encryption)
	if config.Compress {
		compressedPath, err := bm.compressBackup(backupPath)
		if err != nil {
			// Keep uncompressed backup if compression fails
			return backupPath, fmt.Errorf("backup created but compression failed: %w", err)
		}
		// Remove uncompressed backup
		_ = os.Remove(backupPath)
		backupPath = compressedPath
	}

	// Run cleanup if retention policy is set
	if config.RetentionDays > 0 || config.MaxBackups > 0 {
		if err := bm.CleanupBackups(backupDir, config.RetentionDays, config.MaxBackups); err != nil {
			// Log error but don't fail backup
			_ = err
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
// For encrypted backups, pass the encryption password in encryptionPassword parameter.
func (bm *BackupManager) Restore(backupPath string, encryptionPassword ...string) error {
	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	actualBackupPath := backupPath
	var tempFiles []string

	// Check if backup is compressed and decompress if needed
	if strings.HasSuffix(backupPath, ".gz") {
		decompressedPath, err := bm.decompressBackup(backupPath)
		if err != nil {
			return fmt.Errorf("failed to decompress backup: %w", err)
		}
		actualBackupPath = decompressedPath
		tempFiles = append(tempFiles, decompressedPath)
	}

	// Check if backup is encrypted and decrypt if needed
	isEncrypted, err := IsEncrypted(actualBackupPath)
	if err != nil {
		for _, f := range tempFiles {
			_ = os.Remove(f)
		}
		return fmt.Errorf("failed to check if backup is encrypted: %w", err)
	}

	if isEncrypted {
		// Need password for decryption
		if len(encryptionPassword) == 0 || encryptionPassword[0] == "" {
			for _, f := range tempFiles {
				_ = os.Remove(f)
			}
			return fmt.Errorf("backup is encrypted but no password provided")
		}

		decryptedPath := actualBackupPath + ".decrypted"
		encConfig := DefaultEncryptionConfig(encryptionPassword[0])

		if err := DecryptFile(actualBackupPath, decryptedPath, encConfig); err != nil {
			for _, f := range tempFiles {
				_ = os.Remove(f)
			}
			return fmt.Errorf("failed to decrypt backup: %w", err)
		}

		tempFiles = append(tempFiles, decryptedPath)
		actualBackupPath = decryptedPath
	}

	// Clean up temporary files after restore
	defer func() {
		for _, f := range tempFiles {
			_ = os.Remove(f)
		}
	}()

	// Verify backup integrity
	if err := bm.VerifyBackup(actualBackupPath); err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}

	// Close any existing connections to the database
	// (This is handled by the caller closing the DB connection)

	// Create a temporary file for the restore
	tempPath := bm.dbPath + ".restore.tmp"

	// Copy backup to temporary location
	sourceFile, err := os.Open(actualBackupPath)
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

// encryptBackup encrypts a backup file using AES-256-GCM.
func (bm *BackupManager) encryptBackup(backupPath string, config *BackupConfig) (string, error) {
	// Prepare encryption config
	encConfig := config.EncryptionConfig
	if encConfig == nil {
		encConfig = DefaultEncryptionConfig(config.EncryptionPassword)
	} else {
		encConfig.Password = config.EncryptionPassword
	}

	// Create encrypted file path
	encryptedPath := backupPath + ".enc"

	// Encrypt the backup file
	if err := EncryptFile(backupPath, encryptedPath, encConfig); err != nil {
		return "", fmt.Errorf("failed to encrypt backup: %w", err)
	}

	return encryptedPath, nil
}

// compressBackup compresses a backup file using gzip.
func (bm *BackupManager) compressBackup(backupPath string) (string, error) {
	// Open source file
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to open backup file: %w", err)
	}
	defer func() { _ = sourceFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Create compressed file
	compressedPath := backupPath + ".gz"
	destFile, err := os.Create(compressedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer func() { _ = destFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Create gzip writer
	gzipWriter := gzip.NewWriter(destFile)
	defer func() { _ = gzipWriter.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Copy and compress
	if _, err := io.Copy(gzipWriter, sourceFile); err != nil {
		_ = os.Remove(compressedPath)
		return "", fmt.Errorf("failed to compress backup: %w", err)
	}

	return compressedPath, nil
}

// decompressBackup decompresses a gzipped backup file.
func (bm *BackupManager) decompressBackup(compressedPath string) (string, error) {
	// Open compressed file
	compressedFile, err := os.Open(compressedPath)
	if err != nil {
		return "", fmt.Errorf("failed to open compressed file: %w", err)
	}
	defer func() { _ = compressedFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Create gzip reader
	gzipReader, err := gzip.NewReader(compressedFile)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzipReader.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Create decompressed file (remove .gz extension)
	decompressedPath := strings.TrimSuffix(compressedPath, ".gz")
	destFile, err := os.Create(decompressedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create decompressed file: %w", err)
	}
	defer func() { _ = destFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Decompress
	if _, err := io.Copy(destFile, gzipReader); err != nil {
		_ = os.Remove(decompressedPath)
		return "", fmt.Errorf("failed to decompress backup: %w", err)
	}

	return decompressedPath, nil
}

// CleanupBackups removes old backups based on retention policy.
func (bm *BackupManager) CleanupBackups(backupDir string, retentionDays int, maxBackups int) error {
	if backupDir == "" {
		backupDir = bm.GetBackupDir()
	}

	// Get all backup files
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Collect backup files with their info
	type backupFile struct {
		path    string
		modTime time.Time
	}
	var backups []backupFile

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Include .db and .db.gz files
		ext := filepath.Ext(entry.Name())
		if ext != ".db" && ext != ".gz" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, backupFile{
			path:    filepath.Join(backupDir, entry.Name()),
			modTime: info.ModTime(),
		})
	}

	// Sort backups by modification time (oldest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.Before(backups[j].modTime)
	})

	var removed int

	// Remove backups older than retention days
	if retentionDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -retentionDays)
		for _, backup := range backups {
			if backup.modTime.Before(cutoff) {
				if err := os.Remove(backup.path); err == nil {
					removed++
				}
			}
		}
	}

	// Remove excess backups beyond maxBackups limit
	if maxBackups > 0 && len(backups)-removed > maxBackups {
		toRemove := len(backups) - removed - maxBackups
		for i := 0; i < toRemove && i < len(backups); i++ {
			if err := os.Remove(backups[i].path); err == nil {
				removed++
			}
		}
	}

	return nil
}

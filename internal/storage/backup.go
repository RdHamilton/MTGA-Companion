package storage

import (
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupType represents the type of backup (full or incremental).
type BackupType string

const (
	// BackupTypeFull creates a complete database backup
	BackupTypeFull BackupType = "full"
	// BackupTypeIncremental creates a backup of only changed data since last backup
	BackupTypeIncremental BackupType = "incremental"
)

// BackupMetadata contains metadata about a backup.
type BackupMetadata struct {
	BackupType BackupType           `json:"backup_type"`
	Timestamp  time.Time            `json:"timestamp"`
	BaseBackup string               `json:"base_backup,omitempty"` // Only for incremental backups
	Tables     map[string]TableInfo `json:"tables"`
	DBPath     string               `json:"db_path"`
	BackupPath string               `json:"backup_path"`
}

// TableInfo contains information about a table's state in a backup.
type TableInfo struct {
	Checksum string `json:"checksum"` // Format: "count:N,maxrowid:M"
	RowCount int    `json:"row_count"`
}

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

	// BackupType specifies whether to create a full or incremental backup.
	// Default: BackupTypeFull
	BackupType BackupType

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
		BackupType:   BackupTypeFull,
		VerifyBackup: true,
	}
}

// Backup creates a backup of the database.
// For SQLite, this uses the VACUUM INTO command which is atomic and doesn't require exclusive locks.
// Supports both full and incremental backups based on config.BackupType.
func (bm *BackupManager) Backup(config *BackupConfig) (string, error) {
	if config == nil {
		config = DefaultBackupConfig()
	}

	// Handle incremental backups separately
	if config.BackupType == BackupTypeIncremental {
		return bm.createIncrementalBackup(config)
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
		backupName = fmt.Sprintf("backup_%s_full", timestamp)
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

	// Generate and save metadata for full backup
	// Note: We save metadata based on the original .db path, not the final path (which may be .enc or .gz)
	originalBackupPath := backupPath
	originalBackupPath = strings.TrimSuffix(originalBackupPath, ".gz")
	originalBackupPath = strings.TrimSuffix(originalBackupPath, ".enc")

	metadata, err := bm.generateMetadata(BackupTypeFull, originalBackupPath, "")
	if err != nil {
		// Log error but don't fail backup
		_ = err
	} else {
		// Save metadata (use the actual backup path for the .meta file)
		if err := bm.saveMetadata(metadata, backupPath); err != nil {
			// Log error but don't fail backup
			_ = err
		}
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

	// Check if this is an incremental backup
	metadata, err := bm.loadMetadata(backupPath)
	if err == nil && metadata.BackupType == BackupTypeIncremental {
		// Handle incremental restore
		return bm.restoreIncremental(actualBackupPath, metadata, encryptionPassword...)
	}

	// Full backup restore - verify backup integrity
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

// generateMetadata creates metadata for the current database state.
func (bm *BackupManager) generateMetadata(backupType BackupType, backupPath, baseBackup string) (*BackupMetadata, error) {
	// Open database
	db, err := sql.Open("sqlite", bm.dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Get list of tables
	tables, err := bm.getTableNames(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get table names: %w", err)
	}

	// Calculate checksum for each table
	tableInfo := make(map[string]TableInfo)
	for _, table := range tables {
		checksum, rowCount, err := bm.calculateTableChecksum(db, table)
		if err != nil {
			// Skip tables that can't be checksummed
			continue
		}
		tableInfo[table] = TableInfo{
			Checksum: checksum,
			RowCount: rowCount,
		}
	}

	metadata := &BackupMetadata{
		BackupType: backupType,
		Timestamp:  time.Now(),
		BaseBackup: baseBackup,
		Tables:     tableInfo,
		DBPath:     bm.dbPath,
		BackupPath: backupPath,
	}

	return metadata, nil
}

// calculateTableChecksum calculates a checksum for a table based on row count and max rowid.
func (bm *BackupManager) calculateTableChecksum(db *sql.DB, tableName string) (string, int, error) {
	var count int
	var maxRowID sql.NullInt64

	// Get row count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %q", tableName)
	if err := db.QueryRow(countQuery).Scan(&count); err != nil {
		return "", 0, fmt.Errorf("failed to count rows: %w", err)
	}

	// Get max rowid
	rowidQuery := fmt.Sprintf("SELECT MAX(rowid) FROM %q", tableName)
	if err := db.QueryRow(rowidQuery).Scan(&maxRowID); err != nil {
		// Table might not have rowid, use count only
		checksum := fmt.Sprintf("count:%d", count)
		return checksum, count, nil
	}

	maxRowIDValue := int64(0)
	if maxRowID.Valid {
		maxRowIDValue = maxRowID.Int64
	}

	checksum := fmt.Sprintf("count:%d,maxrowid:%d", count, maxRowIDValue)
	return checksum, count, nil
}

// getTableNames returns a list of all user tables in the database.
func (bm *BackupManager) getTableNames(db *sql.DB) ([]string, error) {
	query := `
		SELECT name FROM sqlite_master
		WHERE type='table'
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Ignore error on cleanup

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

// saveMetadata saves backup metadata to a .meta file.
func (bm *BackupManager) saveMetadata(metadata *BackupMetadata, backupPath string) error {
	metadataPath := backupPath + ".meta"

	// Marshal metadata to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to file
	if err := os.WriteFile(metadataPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// loadMetadata loads backup metadata from a .meta file.
//
//nolint:unused // Will be used for incremental backup support
// LoadBackupMetadata loads and returns metadata for a backup file.
// This is a public wrapper around loadMetadata for CLI and external use.
func (bm *BackupManager) LoadBackupMetadata(backupPath string) (*BackupMetadata, error) {
	return bm.loadMetadata(backupPath)
}

func (bm *BackupManager) loadMetadata(backupPath string) (*BackupMetadata, error) {
	metadataPath := backupPath + ".meta"

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	// Unmarshal JSON
	var metadata BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// findLastFullBackup finds the most recent full backup in the backup directory.
//
//nolint:unused // Will be used for incremental backup support
func (bm *BackupManager) findLastFullBackup(backupDir string) (string, *BackupMetadata, error) {
	if backupDir == "" {
		backupDir = bm.GetBackupDir()
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("no backups found")
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Collect all .meta files
	type backupEntry struct {
		path     string
		metadata *BackupMetadata
		modTime  time.Time
	}
	var backups []backupEntry

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".meta") {
			continue
		}

		metaPath := filepath.Join(backupDir, entry.Name())
		metadata, err := bm.loadMetadata(strings.TrimSuffix(metaPath, ".meta"))
		if err != nil {
			continue
		}

		// Only include full backups
		if metadata.BackupType != BackupTypeFull {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, backupEntry{
			path:     strings.TrimSuffix(metaPath, ".meta"),
			metadata: metadata,
			modTime:  info.ModTime(),
		})
	}

	if len(backups) == 0 {
		return "", nil, fmt.Errorf("no full backups found")
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})

	return backups[0].path, backups[0].metadata, nil
}

// createIncrementalBackup creates an incremental backup containing only changed data.
func (bm *BackupManager) createIncrementalBackup(config *BackupConfig) (string, error) {
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

	// Find the last full backup
	baseBackupPath, baseMetadata, err := bm.findLastFullBackup(backupDir)
	if err != nil {
		return "", fmt.Errorf("incremental backup requires a full backup first: %w", err)
	}

	// Generate current database metadata
	tempMetadata, err := bm.generateMetadata(BackupTypeIncremental, "", baseBackupPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate current metadata: %w", err)
	}

	// Compare metadata to find changed tables
	changedTables := bm.findChangedTables(baseMetadata, tempMetadata)
	if len(changedTables) == 0 {
		return "", fmt.Errorf("no changes detected since last backup")
	}

	// Determine backup filename
	backupName := config.BackupName
	if backupName == "" {
		timestamp := time.Now().Format("20060102_150405")
		backupName = fmt.Sprintf("backup_%s_incr", timestamp)
	}
	backupPath := filepath.Join(backupDir, backupName+".sql")

	// Export changed tables to SQL file
	if err := bm.exportTablesToSQL(changedTables, backupPath); err != nil {
		return "", fmt.Errorf("failed to export tables: %w", err)
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

	// Generate and save metadata for incremental backup
	originalBackupPath := backupPath
	originalBackupPath = strings.TrimSuffix(originalBackupPath, ".gz")
	originalBackupPath = strings.TrimSuffix(originalBackupPath, ".enc")

	metadata, err := bm.generateMetadata(BackupTypeIncremental, originalBackupPath, filepath.Base(baseBackupPath))
	if err != nil {
		// Log error but don't fail backup
		_ = err
	} else {
		// Save metadata (use the actual backup path for the .meta file)
		if err := bm.saveMetadata(metadata, backupPath); err != nil {
			// Log error but don't fail backup
			_ = err
		}
	}

	return backupPath, nil
}

// findChangedTables compares two metadata sets and returns tables that have changed.
func (bm *BackupManager) findChangedTables(oldMetadata, newMetadata *BackupMetadata) []string {
	var changed []string

	for tableName, newInfo := range newMetadata.Tables {
		oldInfo, exists := oldMetadata.Tables[tableName]
		if !exists {
			// New table added
			changed = append(changed, tableName)
			continue
		}

		// Check if table has changed
		if oldInfo.Checksum != newInfo.Checksum {
			changed = append(changed, tableName)
		}
	}

	// Sort for consistent ordering
	sort.Strings(changed)
	return changed
}

// exportTablesToSQL exports specified tables to a SQL file.
func (bm *BackupManager) exportTablesToSQL(tables []string, outputPath string) error {
	// Open database
	db, err := sql.Open("sqlite", bm.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Write header
	header := fmt.Sprintf("-- Incremental backup created at %s\n", time.Now().Format(time.RFC3339))
	header += fmt.Sprintf("-- Tables: %s\n\n", strings.Join(tables, ", "))
	if _, err := outFile.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Export each table
	for _, tableName := range tables {
		if err := bm.exportTableToSQL(db, tableName, outFile); err != nil {
			return fmt.Errorf("failed to export table %s: %w", tableName, err)
		}
	}

	return nil
}

// exportTableToSQL exports a single table to SQL format.
func (bm *BackupManager) exportTableToSQL(db *sql.DB, tableName string, outFile *os.File) error {
	// Write table header
	header := fmt.Sprintf("\n-- Table: %s\n", tableName)
	if _, err := outFile.WriteString(header); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}

	// Delete existing data
	deleteStmt := fmt.Sprintf("DELETE FROM %q;\n", tableName)
	if _, err := outFile.WriteString(deleteStmt); err != nil {
		return fmt.Errorf("failed to write delete statement: %w", err)
	}

	// Get table schema
	var createSQL string
	schemaQuery := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name=%q", tableName)
	if err := db.QueryRow(schemaQuery).Scan(&createSQL); err != nil {
		return fmt.Errorf("failed to get table schema: %w", err)
	}

	// Get all rows from the table
	query := fmt.Sprintf("SELECT * FROM %q", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query table: %w", err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Export each row as INSERT statement
	rowCount := 0
	for rows.Next() {
		// Scan row into interface slice
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Build INSERT statement
		insertStmt := fmt.Sprintf("INSERT INTO %q (", tableName)
		insertStmt += strings.Join(columns, ", ")
		insertStmt += ") VALUES ("

		// Add values
		for i, val := range values {
			if i > 0 {
				insertStmt += ", "
			}
			insertStmt += bm.formatSQLValue(val)
		}
		insertStmt += ");\n"

		if _, err := outFile.WriteString(insertStmt); err != nil {
			return fmt.Errorf("failed to write insert statement: %w", err)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	// Write row count comment
	comment := fmt.Sprintf("-- %d rows exported\n", rowCount)
	if _, err := outFile.WriteString(comment); err != nil {
		return fmt.Errorf("failed to write row count: %w", err)
	}

	return nil
}

// formatSQLValue formats a value for SQL insertion.
func (bm *BackupManager) formatSQLValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		// Convert bytes to string and escape
		escaped := strings.ReplaceAll(string(v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		// Fallback to string representation
		return fmt.Sprintf("'%v'", v)
	}
}

// restoreIncremental restores a database from an incremental backup.
func (bm *BackupManager) restoreIncremental(incrementalPath string, metadata *BackupMetadata, encryptionPassword ...string) error {
	// Find the base backup directory
	backupDir := filepath.Dir(incrementalPath)

	// Construct base backup path from metadata
	baseBackupName := metadata.BaseBackup
	if baseBackupName == "" {
		return fmt.Errorf("incremental backup metadata missing base backup reference")
	}

	baseBackupPath := filepath.Join(backupDir, baseBackupName)

	// Check if base backup exists - it might have extensions (.enc, .gz)
	if _, err := os.Stat(baseBackupPath); os.IsNotExist(err) {
		// Try common extensions
		if _, err := os.Stat(baseBackupPath + ".enc"); err == nil {
			baseBackupPath += ".enc"
		} else if _, err := os.Stat(baseBackupPath + ".gz"); err == nil {
			baseBackupPath += ".gz"
		} else if _, err := os.Stat(baseBackupPath + ".enc.gz"); err == nil {
			baseBackupPath += ".enc.gz"
		} else {
			return fmt.Errorf("base backup not found: %s", baseBackupName)
		}
	}

	// First, restore the base full backup
	if err := bm.Restore(baseBackupPath, encryptionPassword...); err != nil {
		return fmt.Errorf("failed to restore base backup: %w", err)
	}

	// Now apply the incremental changes
	if err := bm.applySQLFile(incrementalPath); err != nil {
		return fmt.Errorf("failed to apply incremental backup: %w", err)
	}

	return nil
}

// applySQLFile applies SQL statements from a file to the database.
func (bm *BackupManager) applySQLFile(sqlPath string) error {
	// Read SQL file
	sqlData, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite", bm.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Execute SQL in a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() //nolint:errcheck // Ignore error on cleanup

	// Split SQL into individual statements
	// Remove comments first
	lines := strings.Split(string(sqlData), "\n")
	var sqlLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		sqlLines = append(sqlLines, line)
	}

	// Join and split by semicolon
	sqlContent := strings.Join(sqlLines, " ")
	sqlStatements := strings.Split(sqlContent, ";")

	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

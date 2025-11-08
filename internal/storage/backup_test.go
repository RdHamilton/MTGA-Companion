package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupManager_Backup(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create a simple table and insert data
	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO test (name) VALUES ('test')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Test backup with default config
	configBackup := DefaultBackupConfig()
	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file was not created: %s", backupPath)
	}

	// Verify backup
	if err := backupMgr.VerifyBackup(backupPath); err != nil {
		t.Fatalf("Backup verification failed: %v", err)
	}

	// Clean up
	_ = os.Remove(backupPath)
}

func TestBackupManager_BackupWithCustomName(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Test backup with custom name
	configBackup := DefaultBackupConfig()
	configBackup.BackupName = "custom-backup"
	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup file exists and has correct name
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Backup file was not created: %s", backupPath)
	}

	expectedName := "custom-backup.db"
	if filepath.Base(backupPath) != expectedName {
		t.Errorf("Expected backup name %s, got %s", expectedName, filepath.Base(backupPath))
	}

	// Clean up
	_ = os.Remove(backupPath)
}

func TestBackupManager_VerifyBackup(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create backup
	configBackup := DefaultBackupConfig()
	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	defer os.Remove(backupPath)

	// Test verification
	if err := backupMgr.VerifyBackup(backupPath); err != nil {
		t.Fatalf("Backup verification failed: %v", err)
	}

	// Test verification with non-existent file
	if err := backupMgr.VerifyBackup("/nonexistent/file.db"); err == nil {
		t.Error("Expected error when verifying non-existent file")
	}
}

func TestBackupManager_ListBackups(t *testing.T) {
	// Create a temporary directory for backups
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create a temporary database
	dbPath := filepath.Join(tmpDir, "test.db")
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create a few backups with unique names
	configBackup1 := DefaultBackupConfig()
	configBackup1.BackupDir = backupDir
	configBackup1.BackupName = "test-backup-1"
	_, err = backupMgr.Backup(configBackup1)
	if err != nil {
		t.Fatalf("Failed to create backup 1: %v", err)
	}

	configBackup2 := DefaultBackupConfig()
	configBackup2.BackupDir = backupDir
	configBackup2.BackupName = "test-backup-2"
	_, err = backupMgr.Backup(configBackup2)
	if err != nil {
		t.Fatalf("Failed to create backup 2: %v", err)
	}

	configBackup3 := DefaultBackupConfig()
	configBackup3.BackupDir = backupDir
	configBackup3.BackupName = "test-backup-3"
	_, err = backupMgr.Backup(configBackup3)
	if err != nil {
		t.Fatalf("Failed to create backup 3: %v", err)
	}

	// List backups
	backups, err := backupMgr.ListBackups(backupDir)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) < 3 {
		t.Errorf("Expected at least 3 backups, got %d", len(backups))
	}

	// Verify backup info
	for _, backup := range backups {
		if backup.Path == "" {
			t.Error("Backup path is empty")
		}
		if backup.Name == "" {
			t.Error("Backup name is empty")
		}
		if backup.Size <= 0 {
			t.Error("Backup size should be greater than 0")
		}
		if backup.Checksum == "" {
			t.Error("Backup checksum is empty")
		}
	}
}

func TestBackupManager_Restore(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database with data
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create a simple table and insert data
	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO test (name) VALUES ('original')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create backup
	configBackup := DefaultBackupConfig()
	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Modify the database
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}

	_, err = db.Conn().Exec("UPDATE test SET name = 'modified'")
	if err != nil {
		t.Fatalf("Failed to modify database: %v", err)
	}
	db.Close()

	// Restore from backup
	if err := backupMgr.Restore(backupPath); err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify restore
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database after restore: %v", err)
	}
	defer db.Close()

	var name string
	err = db.Conn().QueryRow("SELECT name FROM test WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query restored database: %v", err)
	}

	if name != "original" {
		t.Errorf("Expected 'original', got '%s'", name)
	}

	// Clean up
	_ = os.Remove(backupPath)
}

func TestBackupManager_GetBackupDir(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	backupMgr := NewBackupManager(dbPath)
	backupDir := backupMgr.GetBackupDir()

	expectedDir := filepath.Join(tmpDir, "backups")
	if backupDir != expectedDir {
		t.Errorf("Expected backup dir %s, got %s", expectedDir, backupDir)
	}
}

func TestBackupManager_EncryptedBackup(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database with data
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create a simple table and insert data
	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, secret TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO test (secret) VALUES ('confidential data')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create encrypted backup
	password := "secure-password-123"
	configBackup := DefaultBackupConfig()
	configBackup.Encrypt = true
	configBackup.EncryptionPassword = password

	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create encrypted backup: %v", err)
	}
	defer os.Remove(backupPath)

	// Verify backup file exists and has .enc extension
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatalf("Encrypted backup file was not created: %s", backupPath)
	}

	if filepath.Ext(backupPath) != ".enc" {
		t.Errorf("Expected .enc extension, got %s", filepath.Ext(backupPath))
	}

	// Verify backup is encrypted
	isEnc, err := IsEncrypted(backupPath)
	if err != nil {
		t.Fatalf("Failed to check if backup is encrypted: %v", err)
	}
	if !isEnc {
		t.Error("Backup should be encrypted")
	}

	// Restore encrypted backup
	if err := backupMgr.Restore(backupPath, password); err != nil {
		t.Fatalf("Failed to restore encrypted backup: %v", err)
	}

	// Verify restored data
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database after restore: %v", err)
	}
	defer db.Close()

	var secret string
	err = db.Conn().QueryRow("SELECT secret FROM test WHERE id = 1").Scan(&secret)
	if err != nil {
		t.Fatalf("Failed to query restored database: %v", err)
	}

	if secret != "confidential data" {
		t.Errorf("Expected 'confidential data', got '%s'", secret)
	}
}

func TestBackupManager_EncryptedBackupWrongPassword(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create encrypted backup
	correctPassword := "correct-password"
	configBackup := DefaultBackupConfig()
	configBackup.Encrypt = true
	configBackup.EncryptionPassword = correctPassword

	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create encrypted backup: %v", err)
	}
	defer os.Remove(backupPath)

	// Try to restore with wrong password
	wrongPassword := "wrong-password"
	err = backupMgr.Restore(backupPath, wrongPassword)
	if err == nil {
		t.Error("Restore with wrong password should fail")
	}
}

func TestBackupManager_EncryptedBackupNoPassword(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Try to create encrypted backup without password
	configBackup := DefaultBackupConfig()
	configBackup.Encrypt = true
	configBackup.EncryptionPassword = ""

	_, err = backupMgr.Backup(configBackup)
	if err == nil {
		t.Error("Creating encrypted backup without password should fail")
	}
}

func TestBackupManager_EncryptedAndCompressedBackup(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database with data
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create table with some data
	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, data TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO test (data) VALUES ('test data')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create encrypted and compressed backup
	password := "password-123"
	configBackup := DefaultBackupConfig()
	configBackup.Encrypt = true
	configBackup.EncryptionPassword = password
	configBackup.Compress = true

	backupPath, err := backupMgr.Backup(configBackup)
	if err != nil {
		t.Fatalf("Failed to create encrypted and compressed backup: %v", err)
	}
	defer os.Remove(backupPath)

	// Verify backup has both .enc and .gz extensions
	if filepath.Ext(backupPath) != ".gz" {
		t.Errorf("Expected .gz extension, got %s", filepath.Ext(backupPath))
	}

	// Verify base name has .enc
	baseName := filepath.Base(backupPath)
	if !contains(baseName, ".enc") {
		t.Error("Backup filename should contain .enc")
	}

	// Restore encrypted and compressed backup
	if err := backupMgr.Restore(backupPath, password); err != nil {
		t.Fatalf("Failed to restore encrypted and compressed backup: %v", err)
	}

	// Verify restored data
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database after restore: %v", err)
	}
	defer db.Close()

	var data string
	err = db.Conn().QueryRow("SELECT data FROM test WHERE id = 1").Scan(&data)
	if err != nil {
		t.Fatalf("Failed to query restored database: %v", err)
	}

	if data != "test data" {
		t.Errorf("Expected 'test data', got '%s'", data)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && substringMatch(s, substr)
}

func substringMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBackupManager_IncrementalBackup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database with initial data
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create tables
	_, err = db.Conn().Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Conn().Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)")
	if err != nil {
		t.Fatalf("Failed to create posts table: %v", err)
	}

	// Insert initial data
	_, err = db.Conn().Exec("INSERT INTO users (name) VALUES ('Alice')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO posts (title) VALUES ('First Post')")
	if err != nil {
		t.Fatalf("Failed to insert post: %v", err)
	}

	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create full backup
	fullConfig := DefaultBackupConfig()
	fullConfig.BackupType = BackupTypeFull
	fullBackupPath, err := backupMgr.Backup(fullConfig)
	if err != nil {
		t.Fatalf("Failed to create full backup: %v", err)
	}
	defer os.Remove(fullBackupPath)

	// Verify metadata was created
	metaPath := fullBackupPath + ".meta"
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Full backup metadata not created")
	}
	defer os.Remove(metaPath)

	// Modify database - change one table
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO users (name) VALUES ('Bob')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	db.Close()

	// Create incremental backup
	incrConfig := DefaultBackupConfig()
	incrConfig.BackupType = BackupTypeIncremental
	incrBackupPath, err := backupMgr.Backup(incrConfig)
	if err != nil {
		t.Fatalf("Failed to create incremental backup: %v", err)
	}
	defer os.Remove(incrBackupPath)

	// Verify incremental backup is SQL file
	if filepath.Ext(incrBackupPath) != ".sql" {
		t.Errorf("Expected .sql extension for incremental backup, got %s", filepath.Ext(incrBackupPath))
	}

	// Verify metadata
	incrMetaPath := incrBackupPath + ".meta"
	if _, err := os.Stat(incrMetaPath); os.IsNotExist(err) {
		t.Fatal("Incremental backup metadata not created")
	}
	defer os.Remove(incrMetaPath)

	// Verify SQL file contains only users table (the changed one)
	sqlData, err := os.ReadFile(incrBackupPath)
	if err != nil {
		t.Fatalf("Failed to read incremental backup: %v", err)
	}

	sqlContent := string(sqlData)
	if !contains(sqlContent, "users") {
		t.Error("Incremental backup should contain users table")
	}
	if contains(sqlContent, "posts") {
		t.Error("Incremental backup should not contain unchanged posts table")
	}
}

func TestBackupManager_IncrementalRestore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create table and insert data
	_, err = db.Conn().Exec("CREATE TABLE data (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO data (value) VALUES ('original')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	db.Close()

	// Create backup manager
	backupMgr := NewBackupManager(dbPath)

	// Create full backup
	fullConfig := DefaultBackupConfig()
	fullConfig.BackupType = BackupTypeFull
	fullBackupPath, err := backupMgr.Backup(fullConfig)
	if err != nil {
		t.Fatalf("Failed to create full backup: %v", err)
	}
	defer os.Remove(fullBackupPath)
	defer os.Remove(fullBackupPath + ".meta")

	// Modify database
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO data (value) VALUES ('incremental')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	db.Close()

	// Create incremental backup
	incrConfig := DefaultBackupConfig()
	incrConfig.BackupType = BackupTypeIncremental
	incrBackupPath, err := backupMgr.Backup(incrConfig)
	if err != nil {
		t.Fatalf("Failed to create incremental backup: %v", err)
	}
	defer os.Remove(incrBackupPath)
	defer os.Remove(incrBackupPath + ".meta")

	// Corrupt the database
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}

	_, err = db.Conn().Exec("DELETE FROM data")
	if err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	db.Close()

	// Restore from incremental backup (should restore full + incremental)
	if err := backupMgr.Restore(incrBackupPath); err != nil {
		t.Fatalf("Failed to restore incremental backup: %v", err)
	}

	// Verify restored data includes both original and incremental data
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database after restore: %v", err)
	}
	defer db.Close()

	var count int
	err = db.Conn().QueryRow("SELECT COUNT(*) FROM data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query restored database: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 rows after incremental restore, got %d", count)
	}
}

func TestBackupManager_IncrementalBackupNoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	db.Close()

	backupMgr := NewBackupManager(dbPath)

	// Create full backup
	fullConfig := DefaultBackupConfig()
	fullConfig.BackupType = BackupTypeFull
	fullBackupPath, err := backupMgr.Backup(fullConfig)
	if err != nil {
		t.Fatalf("Failed to create full backup: %v", err)
	}
	defer os.Remove(fullBackupPath)
	defer os.Remove(fullBackupPath + ".meta")

	// Try to create incremental backup without changes
	incrConfig := DefaultBackupConfig()
	incrConfig.BackupType = BackupTypeIncremental
	_, err = backupMgr.Backup(incrConfig)
	if err == nil {
		t.Error("Expected error when creating incremental backup with no changes")
	}
}

func TestBackupManager_IncrementalBackupNoFullBackup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	db.Close()

	backupMgr := NewBackupManager(dbPath)

	// Try to create incremental backup without a full backup first
	incrConfig := DefaultBackupConfig()
	incrConfig.BackupType = BackupTypeIncremental
	_, err = backupMgr.Backup(incrConfig)
	if err == nil {
		t.Error("Expected error when creating incremental backup without full backup")
	}
}

func TestBackupManager_IncrementalBackupEncrypted(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	_, err = db.Conn().Exec("CREATE TABLE data (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO data (value) VALUES ('test')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	db.Close()

	backupMgr := NewBackupManager(dbPath)
	password := "secure-password"

	// Create encrypted full backup
	fullConfig := DefaultBackupConfig()
	fullConfig.BackupType = BackupTypeFull
	fullConfig.Encrypt = true
	fullConfig.EncryptionPassword = password
	fullBackupPath, err := backupMgr.Backup(fullConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted full backup: %v", err)
	}
	defer os.Remove(fullBackupPath)
	defer os.Remove(fullBackupPath + ".meta")

	// Modify database
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO data (value) VALUES ('incremental')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	db.Close()

	// Create encrypted incremental backup
	incrConfig := DefaultBackupConfig()
	incrConfig.BackupType = BackupTypeIncremental
	incrConfig.Encrypt = true
	incrConfig.EncryptionPassword = password
	incrBackupPath, err := backupMgr.Backup(incrConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted incremental backup: %v", err)
	}
	defer os.Remove(incrBackupPath)
	defer os.Remove(incrBackupPath + ".meta")

	// Verify backup is encrypted
	if filepath.Ext(incrBackupPath) != ".enc" {
		t.Errorf("Expected .enc extension for encrypted incremental backup, got %s", filepath.Ext(incrBackupPath))
	}

	// Restore from encrypted incremental backup
	if err := backupMgr.Restore(incrBackupPath, password); err != nil {
		t.Fatalf("Failed to restore encrypted incremental backup: %v", err)
	}

	// Verify data
	db, err = Open(config)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db.Close()

	var count int
	err = db.Conn().QueryRow("SELECT COUNT(*) FROM data").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}

func TestBackupManager_MetadataGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	_, err = db.Conn().Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, data TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Conn().Exec("INSERT INTO test (data) VALUES ('test1'), ('test2')")
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	db.Close()

	backupMgr := NewBackupManager(dbPath)

	// Generate metadata
	metadata, err := backupMgr.generateMetadata(BackupTypeFull, dbPath, "")
	if err != nil {
		t.Fatalf("Failed to generate metadata: %v", err)
	}

	// Verify metadata
	if metadata.BackupType != BackupTypeFull {
		t.Errorf("Expected backup type %s, got %s", BackupTypeFull, metadata.BackupType)
	}

	if len(metadata.Tables) == 0 {
		t.Error("Metadata should contain table information")
	}

	testTable, exists := metadata.Tables["test"]
	if !exists {
		t.Error("Metadata should contain 'test' table")
	}

	if testTable.RowCount != 2 {
		t.Errorf("Expected row count 2, got %d", testTable.RowCount)
	}

	if testTable.Checksum == "" {
		t.Error("Table checksum should not be empty")
	}
}

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

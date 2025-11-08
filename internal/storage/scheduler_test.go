package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewBackupScheduler(t *testing.T) {
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

	// Test with default config
	scheduler := NewBackupScheduler(backupMgr, nil)
	if scheduler == nil {
		t.Fatal("NewBackupScheduler returned nil")
	}

	if scheduler.manager != backupMgr {
		t.Error("Scheduler manager not set correctly")
	}

	if scheduler.config == nil {
		t.Error("Scheduler config should be set to default")
	}

	if scheduler.config.Interval != 24*time.Hour {
		t.Errorf("Expected default interval 24h, got %v", scheduler.config.Interval)
	}

	// Test with custom config
	customConfig := &SchedulerConfig{
		Interval:     1 * time.Hour,
		BackupConfig: DefaultBackupConfig(),
	}

	scheduler = NewBackupScheduler(backupMgr, customConfig)
	if scheduler.config.Interval != 1*time.Hour {
		t.Errorf("Expected interval 1h, got %v", scheduler.config.Interval)
	}
}

func TestBackupScheduler_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	schedulerConfig := &SchedulerConfig{
		Interval:     1 * time.Second,
		BackupConfig: DefaultBackupConfig(),
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	// Test starting scheduler
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("Scheduler should be running")
	}

	// Test starting already running scheduler
	if err := scheduler.Start(); err == nil {
		t.Error("Starting already running scheduler should fail")
	}

	// Test stopping scheduler
	if err := scheduler.Stop(); err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("Scheduler should not be running")
	}

	// Test stopping already stopped scheduler
	if err := scheduler.Stop(); err == nil {
		t.Error("Stopping already stopped scheduler should fail")
	}
}

func TestBackupScheduler_StartImmediately(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	var mu sync.Mutex
	backupExecuted := false
	schedulerConfig := &SchedulerConfig{
		Interval:         10 * time.Second,
		BackupConfig:     DefaultBackupConfig(),
		StartImmediately: true,
		OnBackupComplete: func(backupPath string, err error) {
			mu.Lock()
			backupExecuted = true
			mu.Unlock()
		},
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// Wait for the immediate backup to execute with polling
	// Use longer timeout for slower CI environments
	maxWait := 3 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		mu.Lock()
		executed := backupExecuted
		mu.Unlock()
		if executed {
			break
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}

	mu.Lock()
	executed := backupExecuted
	mu.Unlock()
	if !executed {
		t.Error("Backup should have been executed immediately")
	}

	status := scheduler.Status()
	if status.BackupCount == 0 {
		t.Error("Backup count should be greater than 0")
	}
}

func TestBackupScheduler_ScheduledExecution(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	var mu sync.Mutex
	backupCount := 0
	schedulerConfig := &SchedulerConfig{
		Interval:     500 * time.Millisecond,
		BackupConfig: DefaultBackupConfig(),
		OnBackupComplete: func(backupPath string, err error) {
			mu.Lock()
			backupCount++
			mu.Unlock()
		},
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for at least 2 backups to execute with polling
	// Use longer timeout and check interval for slower CI environments
	maxWait := 5 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		mu.Lock()
		count := backupCount
		mu.Unlock()
		if count >= 2 {
			break
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}

	if err := scheduler.Stop(); err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}

	mu.Lock()
	count := backupCount
	mu.Unlock()
	if count < 2 {
		t.Errorf("Expected at least 2 backups, got %d", count)
	}

	status := scheduler.Status()
	if status.BackupCount < 2 {
		t.Errorf("Expected backup count >= 2, got %d", status.BackupCount)
	}
}

func TestBackupScheduler_TriggerBackup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	var mu sync.Mutex
	backupExecuted := false
	schedulerConfig := &SchedulerConfig{
		Interval:     1 * time.Hour, // Long interval
		BackupConfig: DefaultBackupConfig(),
		OnBackupComplete: func(backupPath string, err error) {
			mu.Lock()
			backupExecuted = true
			mu.Unlock()
		},
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	// Test triggering backup when not running
	if err := scheduler.TriggerBackup(); err == nil {
		t.Error("TriggerBackup should fail when scheduler is not running")
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// Trigger manual backup
	if err := scheduler.TriggerBackup(); err != nil {
		t.Fatalf("TriggerBackup failed: %v", err)
	}

	// Wait for backup to execute with polling
	maxWait := 3 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		mu.Lock()
		executed := backupExecuted
		mu.Unlock()
		if executed {
			break
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}

	mu.Lock()
	executed := backupExecuted
	mu.Unlock()
	if !executed {
		t.Error("Triggered backup should have executed")
	}

	status := scheduler.Status()
	if status.BackupCount == 0 {
		t.Error("Backup count should be greater than 0")
	}
}

func TestBackupScheduler_Status(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	schedulerConfig := &SchedulerConfig{
		Interval:     500 * time.Millisecond,
		BackupConfig: DefaultBackupConfig(),
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	// Test status when not running
	status := scheduler.Status()
	if status.Running {
		t.Error("Status should show not running")
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for at least one backup with polling
	maxWait := 3 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		status = scheduler.Status()
		if status.BackupCount > 0 {
			break
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}

	status = scheduler.Status()
	if !status.Running {
		t.Error("Status should show running")
	}

	if status.Interval != 500*time.Millisecond {
		t.Errorf("Expected interval 500ms, got %v", status.Interval)
	}

	if status.BackupCount == 0 {
		t.Error("Backup count should be greater than 0")
	}

	if status.LastBackup.IsZero() {
		t.Error("LastBackup should be set")
	}

	if status.NextBackup.IsZero() {
		t.Error("NextBackup should be set when running")
	}

	if status.FailureCount != 0 {
		t.Errorf("Expected 0 failures, got %d", status.FailureCount)
	}

	scheduler.Stop()

	// Test status string
	statusStr := status.String()
	if statusStr == "" {
		t.Error("Status string should not be empty")
	}
}

func TestBackupScheduler_UpdateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	schedulerConfig := &SchedulerConfig{
		Interval:     1 * time.Hour,
		BackupConfig: DefaultBackupConfig(),
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	// Test updating config when not running
	newConfig := &SchedulerConfig{
		Interval:     2 * time.Hour,
		BackupConfig: DefaultBackupConfig(),
	}

	if err := scheduler.UpdateConfig(newConfig); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	if scheduler.config.Interval != 2*time.Hour {
		t.Errorf("Expected interval 2h, got %v", scheduler.config.Interval)
	}

	// Test updating config while running
	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	if err := scheduler.UpdateConfig(newConfig); err == nil {
		t.Error("Updating config while running should fail")
	}

	// Test updating with nil config
	scheduler.Stop()
	if err := scheduler.UpdateConfig(nil); err == nil {
		t.Error("Updating with nil config should fail")
	}
}

func TestBackupScheduler_CallbackFunctionality(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	var mu sync.Mutex
	var callbackPath string
	var callbackErr error
	callbackCalled := false

	schedulerConfig := &SchedulerConfig{
		Interval:     500 * time.Millisecond,
		BackupConfig: DefaultBackupConfig(),
		OnBackupComplete: func(backupPath string, err error) {
			mu.Lock()
			callbackPath = backupPath
			callbackErr = err
			callbackCalled = true
			mu.Unlock()
		},
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for backup to execute with polling
	maxWait := 3 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		mu.Lock()
		called := callbackCalled
		mu.Unlock()
		if called {
			break
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}

	scheduler.Stop()

	mu.Lock()
	called := callbackCalled
	path := callbackPath
	cbErr := callbackErr
	mu.Unlock()

	if !called {
		t.Error("Callback should have been called")
	}

	if path == "" {
		t.Error("Callback should have received backup path")
	}

	if cbErr != nil {
		t.Errorf("Callback should not have received error: %v", cbErr)
	}

	// Verify backup file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Backup file from callback should exist: %s", path)
	}
}

func TestBackupScheduler_MultipleStartStopCycles(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	backupMgr := NewBackupManager(dbPath)

	schedulerConfig := &SchedulerConfig{
		Interval:     500 * time.Millisecond,
		BackupConfig: DefaultBackupConfig(),
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	// Run multiple start/stop cycles
	for i := 0; i < 3; i++ {
		if err := scheduler.Start(); err != nil {
			t.Fatalf("Cycle %d: Failed to start scheduler: %v", i, err)
		}

		if !scheduler.IsRunning() {
			t.Errorf("Cycle %d: Scheduler should be running", i)
		}

		// Wait a bit for scheduler to settle
		time.Sleep(300 * time.Millisecond)

		if err := scheduler.Stop(); err != nil {
			t.Fatalf("Cycle %d: Failed to stop scheduler: %v", i, err)
		}

		if scheduler.IsRunning() {
			t.Errorf("Cycle %d: Scheduler should not be running", i)
		}
	}

	status := scheduler.Status()
	if status.Running {
		t.Error("Scheduler should not be running after cycles")
	}
}

func TestBackupScheduler_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a simple database
	config := DefaultConfig(dbPath)
	db, err := Open(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	db.Close()

	backupMgr := NewBackupManager(dbPath)

	// Create a backup config that will cause errors
	// Use a file path as the backup directory (should fail on all platforms)
	invalidBackupDir := filepath.Join(tmpDir, "invalid.txt")
	if err := os.WriteFile(invalidBackupDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("Failed to create invalid backup dir file: %v", err)
	}

	backupConfig := DefaultBackupConfig()
	backupConfig.BackupDir = invalidBackupDir // This is a file, not a directory

	schedulerConfig := &SchedulerConfig{
		Interval:         500 * time.Millisecond,
		BackupConfig:     backupConfig,
		StartImmediately: true,
	}

	scheduler := NewBackupScheduler(backupMgr, schedulerConfig)

	if err := scheduler.Start(); err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	// Wait for backup attempt with polling
	// Use longer timeout for slower CI environments and to account for:
	// - StartImmediately goroutine scheduling delay
	// - Backup execution time
	// - Interval time (500ms)
	maxWait := 5 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		status := scheduler.Status()
		if status.FailureCount > 0 {
			break
		}
		time.Sleep(checkInterval)
		elapsed += checkInterval
	}

	scheduler.Stop()

	status := scheduler.Status()
	if status.FailureCount == 0 {
		t.Errorf("Failure count should be greater than 0 for failed backups (waited %v)", elapsed)
	}

	if status.LastError == nil {
		t.Error("LastError should be set after failed backup")
	}
}

func TestSchedulerStatus_String(t *testing.T) {
	// Test stopped status
	status := &SchedulerStatus{
		Running: false,
	}

	str := status.String()
	if str != "Scheduler: Stopped" {
		t.Errorf("Expected 'Scheduler: Stopped', got '%s'", str)
	}

	// Test running status
	now := time.Now()
	status = &SchedulerStatus{
		Running:      true,
		Interval:     24 * time.Hour,
		BackupCount:  5,
		FailureCount: 1,
		LastBackup:   now,
		NextBackup:   now.Add(24 * time.Hour),
	}

	str = status.String()
	if str == "" {
		t.Error("Running status string should not be empty")
	}

	// Should contain key information
	if !contains(str, "Running") {
		t.Error("Status string should contain 'Running'")
	}

	if !contains(str, "24h") {
		t.Error("Status string should contain interval")
	}
}

func TestDefaultSchedulerConfig(t *testing.T) {
	config := DefaultSchedulerConfig()

	if config == nil {
		t.Fatal("DefaultSchedulerConfig returned nil")
	}

	if config.Interval != 24*time.Hour {
		t.Errorf("Expected default interval 24h, got %v", config.Interval)
	}

	if config.BackupConfig == nil {
		t.Error("Default backup config should be set")
	}

	if config.StartImmediately {
		t.Error("StartImmediately should default to false")
	}
}

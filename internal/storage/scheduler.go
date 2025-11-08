package storage

import (
	"fmt"
	"sync"
	"time"
)

// BackupScheduler manages automatic scheduled backups.
type BackupScheduler struct {
	manager       *BackupManager
	config        *SchedulerConfig
	ticker        *time.Ticker
	stopChan      chan struct{}
	mu            sync.RWMutex
	running       bool
	lastBackup    time.Time
	lastError     error
	backupCount   int
	failureCount  int
	backupHandler func(backupPath string, err error)
}

// SchedulerConfig holds configuration for the backup scheduler.
type SchedulerConfig struct {
	// Interval is how often to run backups
	// Examples: 24*time.Hour (daily), 7*24*time.Hour (weekly)
	Interval time.Duration

	// BackupConfig is the configuration to use for each backup
	BackupConfig *BackupConfig

	// StartImmediately runs a backup immediately when scheduler starts
	// Default: false
	StartImmediately bool

	// OnBackupComplete is called after each backup attempt (success or failure)
	// Optional callback for logging or notifications
	OnBackupComplete func(backupPath string, err error)
}

// DefaultSchedulerConfig returns a scheduler config with daily backups.
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		Interval:         24 * time.Hour, // Daily
		BackupConfig:     DefaultBackupConfig(),
		StartImmediately: false,
	}
}

// NewBackupScheduler creates a new backup scheduler.
func NewBackupScheduler(manager *BackupManager, config *SchedulerConfig) *BackupScheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	return &BackupScheduler{
		manager:       manager,
		config:        config,
		stopChan:      make(chan struct{}),
		backupHandler: config.OnBackupComplete,
	}
}

// Start starts the backup scheduler.
// Returns an error if the scheduler is already running.
func (s *BackupScheduler) Start() error {
	s.mu.Lock()

	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}

	s.ticker = time.NewTicker(s.config.Interval)
	s.running = true
	ticker := s.ticker // Store reference before unlocking
	s.mu.Unlock()

	// Run backup immediately if configured
	if s.config.StartImmediately {
		go s.runBackup()
	}

	// Start scheduler goroutine with ticker reference
	go s.run(ticker)

	return nil
}

// Stop stops the backup scheduler.
// Blocks until the current backup (if any) completes.
func (s *BackupScheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is not running")
	}
	s.mu.Unlock()

	// Signal stop
	close(s.stopChan)

	// Wait for ticker to stop
	s.mu.Lock()
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
	s.running = false
	s.mu.Unlock()

	// Create new stop channel for potential restart
	s.stopChan = make(chan struct{})

	return nil
}

// run is the main scheduler loop.
func (s *BackupScheduler) run(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			s.runBackup()
		case <-s.stopChan:
			return
		}
	}
}

// runBackup executes a backup and updates statistics.
func (s *BackupScheduler) runBackup() {
	backupPath, err := s.manager.Backup(s.config.BackupConfig)

	s.mu.Lock()
	s.lastBackup = time.Now()
	s.lastError = err
	if err != nil {
		s.failureCount++
	} else {
		s.backupCount++
	}
	s.mu.Unlock()

	// Call handler if provided
	if s.backupHandler != nil {
		s.backupHandler(backupPath, err)
	}
}

// Status returns the current scheduler status.
func (s *BackupScheduler) Status() *SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nextBackup time.Time
	if s.running && !s.lastBackup.IsZero() {
		nextBackup = s.lastBackup.Add(s.config.Interval)
	}

	return &SchedulerStatus{
		Running:      s.running,
		Interval:     s.config.Interval,
		LastBackup:   s.lastBackup,
		NextBackup:   nextBackup,
		BackupCount:  s.backupCount,
		FailureCount: s.failureCount,
		LastError:    s.lastError,
		UptimeSince:  s.getUptimeSince(),
	}
}

// getUptimeSince returns how long the scheduler has been running.
func (s *BackupScheduler) getUptimeSince() time.Duration {
	if !s.running || s.lastBackup.IsZero() {
		return 0
	}
	return time.Since(s.lastBackup)
}

// IsRunning returns whether the scheduler is currently running.
func (s *BackupScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// TriggerBackup triggers an immediate backup without affecting the schedule.
// This is useful for manual backup requests while scheduler is running.
func (s *BackupScheduler) TriggerBackup() error {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		return fmt.Errorf("scheduler is not running")
	}

	go s.runBackup()
	return nil
}

// UpdateConfig updates the scheduler configuration.
// The scheduler must be stopped and restarted for changes to take effect.
func (s *BackupScheduler) UpdateConfig(config *SchedulerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("cannot update config while scheduler is running")
	}

	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	s.config = config
	s.backupHandler = config.OnBackupComplete

	return nil
}

// SchedulerStatus contains information about the scheduler state.
type SchedulerStatus struct {
	Running      bool
	Interval     time.Duration
	LastBackup   time.Time
	NextBackup   time.Time
	BackupCount  int
	FailureCount int
	LastError    error
	UptimeSince  time.Duration
}

// String returns a human-readable representation of the scheduler status.
func (s *SchedulerStatus) String() string {
	if !s.Running {
		return "Scheduler: Stopped"
	}

	status := "Scheduler: Running\n"
	status += fmt.Sprintf("  Interval: %s\n", s.Interval)
	status += fmt.Sprintf("  Total Backups: %d\n", s.BackupCount)
	status += fmt.Sprintf("  Failures: %d\n", s.FailureCount)

	if !s.LastBackup.IsZero() {
		status += fmt.Sprintf("  Last Backup: %s\n", s.LastBackup.Format(time.RFC3339))
	}

	if !s.NextBackup.IsZero() {
		status += fmt.Sprintf("  Next Backup: %s\n", s.NextBackup.Format(time.RFC3339))
		timeUntil := time.Until(s.NextBackup)
		if timeUntil > 0 {
			status += fmt.Sprintf("  Time Until Next: %s\n", timeUntil.Round(time.Second))
		}
	}

	if s.LastError != nil {
		status += fmt.Sprintf("  Last Error: %v\n", s.LastError)
	}

	return status
}

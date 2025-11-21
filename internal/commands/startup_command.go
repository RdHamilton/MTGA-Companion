package commands

import (
	"context"
	"fmt"
	"log"
)

// DaemonService defines the interface for daemon initialization and control.
type DaemonService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
}

// StorageService defines the interface for storage initialization.
type StorageService interface {
	Initialize(ctx context.Context, dbPath string) error
	IsInitialized() bool
}

// PollerService defines the interface for log poller operations.
type PollerService interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
}

// StartupRecoveryCommand encapsulates the operation of initializing the daemon
// and recovering from potential startup failures.
// Implements the Command pattern for daemon startup operations.
type StartupRecoveryCommand struct {
	BaseCommand
	storage       StorageService
	poller        PollerService
	daemon        DaemonService
	dbPath        string
	enablePoller  bool
	enableDaemon  bool
	retryAttempts int
}

// NewStartupRecoveryCommand creates a new command for daemon initialization and recovery.
// dbPath: path to the database file
// enablePoller: whether to start the log poller
// enableDaemon: whether to start the daemon service
// retryAttempts: number of retry attempts for failed operations
func NewStartupRecoveryCommand(
	storage StorageService,
	poller PollerService,
	daemon DaemonService,
	dbPath string,
	enablePoller bool,
	enableDaemon bool,
	retryAttempts int,
) *StartupRecoveryCommand {
	desc := fmt.Sprintf("Initialize application (DB: %s, Poller: %v, Daemon: %v, Retries: %d)",
		dbPath, enablePoller, enableDaemon, retryAttempts)

	return &StartupRecoveryCommand{
		BaseCommand: BaseCommand{
			name:        "StartupRecovery",
			description: desc,
		},
		storage:       storage,
		poller:        poller,
		daemon:        daemon,
		dbPath:        dbPath,
		enablePoller:  enablePoller,
		enableDaemon:  enableDaemon,
		retryAttempts: retryAttempts,
	}
}

// Execute performs the startup and recovery operations.
func (c *StartupRecoveryCommand) Execute(ctx context.Context) error {
	log.Printf("[StartupRecoveryCommand] Executing: %s", c.description)

	// Step 1: Initialize storage
	if err := c.initializeStorage(ctx); err != nil {
		return fmt.Errorf("storage initialization failed: %w", err)
	}

	// Step 2: Start poller if requested
	if c.enablePoller {
		if err := c.startPoller(ctx); err != nil {
			log.Printf("[StartupRecoveryCommand] Warning: Poller startup failed: %v", err)
			// Don't fail the entire command if poller fails - it's not critical
		}
	}

	// Step 3: Start daemon if requested
	if c.enableDaemon {
		if err := c.startDaemon(ctx); err != nil {
			log.Printf("[StartupRecoveryCommand] Warning: Daemon startup failed: %v", err)
			// Don't fail the entire command if daemon fails - fall back to standalone
		}
	}

	log.Printf("[StartupRecoveryCommand] Startup completed successfully")
	return nil
}

// initializeStorage initializes the storage service with retry logic.
func (c *StartupRecoveryCommand) initializeStorage(ctx context.Context) error {
	log.Printf("[StartupRecoveryCommand] Initializing storage: %s", c.dbPath)

	if c.storage.IsInitialized() {
		log.Printf("[StartupRecoveryCommand] Storage already initialized")
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			log.Printf("[StartupRecoveryCommand] Storage init retry attempt %d/%d", attempt, c.retryAttempts)
		}

		err := c.storage.Initialize(ctx, c.dbPath)
		if err == nil {
			log.Printf("[StartupRecoveryCommand] Storage initialized successfully")
			return nil
		}

		lastErr = err
		log.Printf("[StartupRecoveryCommand] Storage init attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("storage initialization failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// startPoller starts the log poller with retry logic.
func (c *StartupRecoveryCommand) startPoller(ctx context.Context) error {
	log.Printf("[StartupRecoveryCommand] Starting log poller")

	if c.poller.IsRunning() {
		log.Printf("[StartupRecoveryCommand] Poller already running")
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			log.Printf("[StartupRecoveryCommand] Poller start retry attempt %d/%d", attempt, c.retryAttempts)
		}

		err := c.poller.Start(ctx)
		if err == nil {
			log.Printf("[StartupRecoveryCommand] Poller started successfully")
			return nil
		}

		lastErr = err
		log.Printf("[StartupRecoveryCommand] Poller start attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("poller start failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// startDaemon starts the daemon service with retry logic.
func (c *StartupRecoveryCommand) startDaemon(ctx context.Context) error {
	log.Printf("[StartupRecoveryCommand] Starting daemon service")

	if c.daemon != nil && c.daemon.IsRunning() {
		log.Printf("[StartupRecoveryCommand] Daemon already running")
		return nil
	}

	if c.daemon == nil {
		log.Printf("[StartupRecoveryCommand] Daemon service not configured")
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			log.Printf("[StartupRecoveryCommand] Daemon start retry attempt %d/%d", attempt, c.retryAttempts)
		}

		err := c.daemon.Start(ctx)
		if err == nil {
			log.Printf("[StartupRecoveryCommand] Daemon started successfully")
			return nil
		}

		lastErr = err
		log.Printf("[StartupRecoveryCommand] Daemon start attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("daemon start failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// ShutdownCommand encapsulates graceful shutdown operations.
type ShutdownCommand struct {
	BaseCommand
	poller PollerService
	daemon DaemonService
}

// NewShutdownCommand creates a new command for graceful shutdown.
func NewShutdownCommand(poller PollerService, daemon DaemonService) *ShutdownCommand {
	return &ShutdownCommand{
		BaseCommand: BaseCommand{
			name:        "Shutdown",
			description: "Gracefully shutdown all services",
		},
		poller: poller,
		daemon: daemon,
	}
}

// Execute performs graceful shutdown.
func (c *ShutdownCommand) Execute(ctx context.Context) error {
	log.Printf("[ShutdownCommand] Executing graceful shutdown")

	// Stop poller first (less critical)
	if c.poller != nil && c.poller.IsRunning() {
		if err := c.poller.Stop(); err != nil {
			log.Printf("[ShutdownCommand] Warning: Failed to stop poller: %v", err)
		} else {
			log.Printf("[ShutdownCommand] Poller stopped successfully")
		}
	}

	// Stop daemon
	if c.daemon != nil && c.daemon.IsRunning() {
		if err := c.daemon.Stop(ctx); err != nil {
			log.Printf("[ShutdownCommand] Warning: Failed to stop daemon: %v", err)
		} else {
			log.Printf("[ShutdownCommand] Daemon stopped successfully")
		}
	}

	log.Printf("[ShutdownCommand] Shutdown completed")
	return nil
}

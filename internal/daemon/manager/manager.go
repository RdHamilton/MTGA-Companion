package manager

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Status represents the current status of the daemon process.
type Status string

const (
	// StatusStopped indicates the daemon is not running.
	StatusStopped Status = "stopped"
	// StatusStarting indicates the daemon is in the process of starting.
	StatusStarting Status = "starting"
	// StatusRunning indicates the daemon is running normally.
	StatusRunning Status = "running"
	// StatusStopping indicates the daemon is in the process of stopping.
	StatusStopping Status = "stopping"
	// StatusError indicates the daemon encountered an error.
	StatusError Status = "error"
)

// Config holds configuration for the DaemonManager.
type Config struct {
	// Port is the WebSocket server port the daemon listens on.
	Port int

	// BinaryPath is the path to the daemon executable.
	// If empty, the manager will attempt to locate the bundled daemon.
	BinaryPath string

	// StartupTimeout is how long to wait for the daemon to become healthy.
	StartupTimeout time.Duration

	// ShutdownTimeout is how long to wait for graceful shutdown before killing.
	ShutdownTimeout time.Duration

	// LogOutput is where to send daemon stdout/stderr. If nil, discards output.
	LogOutput io.Writer
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:            9999,
		BinaryPath:      "", // Will auto-detect
		StartupTimeout:  30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		LogOutput:       os.Stdout,
	}
}

// Manager manages the lifecycle of the mtga-tracker-daemon subprocess.
type Manager struct {
	config *Config

	mu        sync.RWMutex
	cmd       *exec.Cmd
	status    Status
	lastError error
	startTime time.Time
	pid       int

	// Context for managing the process lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Channels for coordination
	doneChan chan struct{}
}

// New creates a new DaemonManager with the given configuration.
func New(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		config: config,
		status: StatusStopped,
	}
}

// Start launches the daemon subprocess.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if m.status == StatusRunning || m.status == StatusStarting {
		return fmt.Errorf("daemon is already %s", m.status)
	}

	m.status = StatusStarting
	m.lastError = nil

	// Find daemon binary
	binaryPath, err := m.findDaemonBinary()
	if err != nil {
		m.status = StatusError
		m.lastError = err
		return fmt.Errorf("failed to find daemon binary: %w", err)
	}

	log.Printf("Starting daemon from: %s", binaryPath)

	// Create context for process management
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.doneChan = make(chan struct{})

	// Build command with arguments
	args := []string{
		"--port", fmt.Sprintf("%d", m.config.Port),
	}

	m.cmd = exec.CommandContext(m.ctx, binaryPath, args...)

	// Set up output redirection
	if m.config.LogOutput != nil {
		m.cmd.Stdout = m.config.LogOutput
		m.cmd.Stderr = m.config.LogOutput
	}

	// Start the process
	if err := m.cmd.Start(); err != nil {
		m.status = StatusError
		m.lastError = fmt.Errorf("failed to start daemon: %w", err)
		return m.lastError
	}

	m.pid = m.cmd.Process.Pid
	m.startTime = time.Now()

	log.Printf("Daemon started with PID %d on port %d", m.pid, m.config.Port)

	// Monitor process in background
	go m.monitorProcess()

	// Wait for daemon to become healthy
	go m.waitForHealthy()

	return nil
}

// Stop gracefully stops the daemon subprocess.
func (m *Manager) Stop() error {
	m.mu.Lock()

	if m.status == StatusStopped || m.status == StatusStopping {
		m.mu.Unlock()
		return nil
	}

	if m.cmd == nil || m.cmd.Process == nil {
		m.status = StatusStopped
		m.mu.Unlock()
		return nil
	}

	m.status = StatusStopping
	pid := m.pid
	m.mu.Unlock()

	log.Printf("Stopping daemon (PID %d)...", pid)

	// Cancel context to signal shutdown
	if m.cancel != nil {
		m.cancel()
	}

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- m.cmd.Wait()
	}()

	select {
	case err := <-done:
		m.mu.Lock()
		m.status = StatusStopped
		m.cmd = nil
		m.pid = 0
		m.mu.Unlock()

		if err != nil {
			log.Printf("Daemon exited with error: %v", err)
		} else {
			log.Println("Daemon stopped gracefully")
		}
		return nil

	case <-time.After(m.config.ShutdownTimeout):
		// Force kill
		log.Printf("Daemon did not stop gracefully, force killing...")
		if err := m.cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill daemon: %v", err)
		}

		m.mu.Lock()
		m.status = StatusStopped
		m.cmd = nil
		m.pid = 0
		m.mu.Unlock()

		return nil
	}
}

// Restart stops and starts the daemon.
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// Brief pause between stop and start
	time.Sleep(500 * time.Millisecond)

	return m.Start()
}

// Status returns the current daemon status.
func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// IsRunning returns true if the daemon is currently running.
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status == StatusRunning
}

// PID returns the process ID of the daemon, or 0 if not running.
func (m *Manager) PID() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pid
}

// Uptime returns how long the daemon has been running.
func (m *Manager) Uptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.status != StatusRunning || m.startTime.IsZero() {
		return 0
	}
	return time.Since(m.startTime)
}

// LastError returns the last error encountered, if any.
func (m *Manager) LastError() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastError
}

// Port returns the configured daemon port.
func (m *Manager) Port() int {
	return m.config.Port
}

// SetPort updates the daemon port. Requires restart to take effect.
func (m *Manager) SetPort(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Port = port
}

// monitorProcess monitors the daemon process and updates status when it exits.
func (m *Manager) monitorProcess() {
	if m.cmd == nil {
		return
	}

	// Wait for process to exit
	err := m.cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Don't update status if we're already stopping
	if m.status == StatusStopping {
		return
	}

	if err != nil {
		log.Printf("Daemon process exited unexpectedly: %v", err)
		m.status = StatusError
		m.lastError = fmt.Errorf("daemon exited unexpectedly: %w", err)
	} else {
		log.Println("Daemon process exited")
		m.status = StatusStopped
	}

	m.cmd = nil
	m.pid = 0

	// Signal completion
	if m.doneChan != nil {
		close(m.doneChan)
	}
}

// waitForHealthy waits for the daemon to become healthy after starting.
func (m *Manager) waitForHealthy() {
	// Give the process a moment to start
	time.Sleep(500 * time.Millisecond)

	m.mu.RLock()
	currentStatus := m.status
	m.mu.RUnlock()

	// If we're no longer starting, exit
	if currentStatus != StatusStarting {
		return
	}

	// Check if process is still alive
	m.mu.RLock()
	cmd := m.cmd
	m.mu.RUnlock()

	if cmd == nil || cmd.Process == nil {
		return
	}

	// TODO: Issue #596 will implement actual health checks via http://localhost:PORT/status
	// For now, assume healthy if process is still running after startup delay
	m.mu.Lock()
	if m.status == StatusStarting {
		m.status = StatusRunning
		log.Printf("Daemon is now running (PID %d)", m.pid)
	}
	m.mu.Unlock()
}

// findDaemonBinary locates the daemon executable.
func (m *Manager) findDaemonBinary() (string, error) {
	// If explicitly configured, use that path
	if m.config.BinaryPath != "" {
		if _, err := os.Stat(m.config.BinaryPath); err != nil {
			return "", fmt.Errorf("configured binary not found: %s", m.config.BinaryPath)
		}
		return m.config.BinaryPath, nil
	}

	// Get executable directory
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Platform-specific binary name
	binaryName := "mtga-tracker-daemon"
	if runtime.GOOS == "windows" {
		binaryName = "mtga-tracker-daemon.exe"
	}

	// Search paths in order of preference
	searchPaths := m.getDaemonSearchPaths(execDir, binaryName)

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("daemon binary not found in any search path: %v", searchPaths)
}

// getDaemonSearchPaths returns platform-specific paths to search for the daemon binary.
func (m *Manager) getDaemonSearchPaths(execDir, binaryName string) []string {
	var paths []string

	switch runtime.GOOS {
	case "darwin":
		// macOS: Check app bundle Resources directory
		// The app bundle structure is: MTGA-Companion.app/Contents/MacOS/mtga-companion
		// Daemon should be at: MTGA-Companion.app/Contents/Resources/daemon/mtga-tracker-daemon

		// Navigate from MacOS to Resources
		resourcesDir := filepath.Join(filepath.Dir(execDir), "Resources")

		// Check for architecture-specific daemon first (arm64 on Apple Silicon)
		if runtime.GOARCH == "arm64" {
			paths = append(paths, filepath.Join(resourcesDir, "daemon-arm64", binaryName))
		}

		// Then check the default daemon directory (x64)
		paths = append(paths, filepath.Join(resourcesDir, "daemon", binaryName))

	case "windows":
		// Windows: Check daemon subdirectory next to executable
		paths = append(paths, filepath.Join(execDir, "daemon", binaryName))

	case "linux":
		// Linux: Check daemon subdirectory next to executable
		paths = append(paths, filepath.Join(execDir, "daemon", binaryName))
	}

	// Development paths (for running from source)
	paths = append(paths,
		filepath.Join(execDir, "..", "resources", "daemon", binaryName),
		filepath.Join(execDir, "resources", "daemon", binaryName),
	)

	return paths
}

// Info returns information about the daemon manager state.
type Info struct {
	Status    Status        `json:"status"`
	PID       int           `json:"pid,omitempty"`
	Port      int           `json:"port"`
	Uptime    time.Duration `json:"uptime,omitempty"`
	LastError string        `json:"lastError,omitempty"`
}

// Info returns current daemon manager information.
func (m *Manager) Info() *Info {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := &Info{
		Status: m.status,
		PID:    m.pid,
		Port:   m.config.Port,
	}

	if m.status == StatusRunning && !m.startTime.IsZero() {
		info.Uptime = time.Since(m.startTime)
	}

	if m.lastError != nil {
		info.LastError = m.lastError.Error()
	}

	return info
}

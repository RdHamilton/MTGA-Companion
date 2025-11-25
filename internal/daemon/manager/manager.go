package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

	// HealthCheckInterval is how often to check daemon health.
	HealthCheckInterval time.Duration

	// EnableAutoRestart enables automatic restart on daemon crash.
	EnableAutoRestart bool

	// MaxRestartAttempts is the maximum number of restart attempts before giving up.
	// Set to 0 for unlimited attempts.
	MaxRestartAttempts int

	// OnHealthChange is called when daemon health status changes.
	// This allows the caller to handle health events (e.g., emit Wails events).
	OnHealthChange func(healthy bool, err error)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:                9999,
		BinaryPath:          "", // Will auto-detect
		StartupTimeout:      30 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		LogOutput:           os.Stdout,
		HealthCheckInterval: 10 * time.Second,
		EnableAutoRestart:   true,
		MaxRestartAttempts:  5,
		OnHealthChange:      nil,
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

	// Health check state
	healthCheckCancel context.CancelFunc
	lastHealthCheck   time.Time
	consecutiveFails  int
	restartAttempts   int
	healthy           bool
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
	m.healthy = false // Will be set true after successful health check

	log.Printf("Daemon started with PID %d on port %d", m.pid, m.config.Port)

	// Monitor process in background
	go m.monitorProcess()

	// Wait for daemon to become healthy
	go m.waitForHealthy()

	// Start health check loop
	go m.startHealthCheck()

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

	// Stop health check first
	m.stopHealthCheck()

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

// ============================================================================
// Health Check and Auto-Recovery
// ============================================================================

// HealthStatus represents the result of a health check.
type HealthStatus struct {
	Healthy          bool      `json:"healthy"`
	LastCheck        time.Time `json:"lastCheck"`
	ConsecutiveFails int       `json:"consecutiveFails"`
	RestartAttempts  int       `json:"restartAttempts"`
	Error            string    `json:"error,omitempty"`
}

// GetHealthStatus returns the current health status.
func (m *Manager) GetHealthStatus() *HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &HealthStatus{
		Healthy:          m.healthy,
		LastCheck:        m.lastHealthCheck,
		ConsecutiveFails: m.consecutiveFails,
		RestartAttempts:  m.restartAttempts,
	}

	if m.lastError != nil {
		status.Error = m.lastError.Error()
	}

	return status
}

// IsHealthy returns true if the daemon is running and healthy.
func (m *Manager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status == StatusRunning && m.healthy
}

// startHealthCheck starts the periodic health check goroutine.
func (m *Manager) startHealthCheck() {
	if m.config.HealthCheckInterval <= 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.healthCheckCancel = cancel

	go m.healthCheckLoop(ctx)
}

// stopHealthCheck stops the health check goroutine.
func (m *Manager) stopHealthCheck() {
	if m.healthCheckCancel != nil {
		m.healthCheckCancel()
		m.healthCheckCancel = nil
	}
}

// healthCheckLoop runs periodic health checks.
func (m *Manager) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	// Initial delay to let daemon fully start
	time.Sleep(2 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck checks daemon health and handles failures.
func (m *Manager) performHealthCheck() {
	m.mu.RLock()
	status := m.status
	port := m.config.Port
	m.mu.RUnlock()

	// Only check if we think we're running
	if status != StatusRunning {
		return
	}

	err := m.checkDaemonHealth(port)

	m.mu.Lock()
	m.lastHealthCheck = time.Now()

	wasHealthy := m.healthy

	if err != nil {
		m.consecutiveFails++
		m.healthy = false
		m.lastError = err

		log.Printf("Daemon health check failed (%d consecutive): %v", m.consecutiveFails, err)

		// Notify health change if callback is set
		if !wasHealthy || m.config.OnHealthChange != nil {
			m.mu.Unlock()
			if m.config.OnHealthChange != nil {
				m.config.OnHealthChange(false, err)
			}
			m.mu.Lock()
		}

		// Check if we should attempt auto-recovery
		if m.config.EnableAutoRestart {
			m.mu.Unlock()
			m.attemptRecovery()
			return
		}
	} else {
		// Health check passed
		if !wasHealthy {
			log.Println("Daemon health check passed, daemon is healthy")
		}
		m.healthy = true
		m.consecutiveFails = 0

		// Notify health change if callback is set and we were unhealthy
		if !wasHealthy && m.config.OnHealthChange != nil {
			m.mu.Unlock()
			m.config.OnHealthChange(true, nil)
			return
		}
	}
	m.mu.Unlock()
}

// checkDaemonHealth performs an HTTP health check against the daemon.
func (m *Manager) checkDaemonHealth(port int) error {
	url := fmt.Sprintf("http://localhost:%d/status", port)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	// Optionally parse the response to check detailed health
	var healthResp struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		// If we can't parse the response, just check status code
		return nil
	}

	if healthResp.Status != "healthy" && healthResp.Status != "ok" && healthResp.Status != "running" {
		return fmt.Errorf("daemon reported unhealthy status: %s", healthResp.Status)
	}

	return nil
}

// attemptRecovery attempts to recover a failed daemon.
func (m *Manager) attemptRecovery() {
	m.mu.Lock()
	m.restartAttempts++
	attempts := m.restartAttempts
	maxAttempts := m.config.MaxRestartAttempts
	m.mu.Unlock()

	// Check if we've exceeded max attempts
	if maxAttempts > 0 && attempts > maxAttempts {
		log.Printf("Maximum restart attempts (%d) exceeded, giving up", maxAttempts)
		m.mu.Lock()
		m.status = StatusError
		m.lastError = fmt.Errorf("maximum restart attempts (%d) exceeded", maxAttempts)
		m.mu.Unlock()
		return
	}

	// Calculate backoff delay using exponential backoff
	delay := m.calculateBackoff(attempts)
	log.Printf("Attempting daemon recovery (attempt %d) after %v delay...", attempts, delay)

	time.Sleep(delay)

	// Check if we're still supposed to be running
	m.mu.RLock()
	status := m.status
	m.mu.RUnlock()

	if status == StatusStopping || status == StatusStopped {
		log.Println("Recovery cancelled: daemon is stopping/stopped")
		return
	}

	// Attempt restart
	if err := m.Restart(); err != nil {
		log.Printf("Recovery restart failed: %v", err)
		// The health check loop will continue to try
	} else {
		log.Println("Daemon recovery successful")
		m.mu.Lock()
		m.restartAttempts = 0 // Reset on successful restart
		m.mu.Unlock()
	}
}

// calculateBackoff returns the backoff delay for the given attempt number.
// Uses exponential backoff: 0s, 5s, 30s, 2m, 2m, 2m...
func (m *Manager) calculateBackoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 0 // Immediate first retry
	case 2:
		return 5 * time.Second
	case 3:
		return 30 * time.Second
	default:
		return 2 * time.Minute // Cap at 2 minutes
	}
}

// ResetRestartAttempts resets the restart attempt counter.
// Call this after successfully connecting to the daemon.
func (m *Manager) ResetRestartAttempts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restartAttempts = 0
	m.consecutiveFails = 0
}

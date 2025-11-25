package manager

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Port != 9999 {
		t.Errorf("expected default port 9999, got %d", config.Port)
	}
	if config.StartupTimeout != 30*time.Second {
		t.Errorf("expected default startup timeout 30s, got %v", config.StartupTimeout)
	}
	if config.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected default shutdown timeout 10s, got %v", config.ShutdownTimeout)
	}
	if config.BinaryPath != "" {
		t.Errorf("expected empty binary path, got %s", config.BinaryPath)
	}
}

func TestNew(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		m := New(nil)
		if m == nil {
			t.Fatal("expected non-nil manager")
		}
		if m.config == nil {
			t.Fatal("expected non-nil config")
		}
		if m.config.Port != 9999 {
			t.Errorf("expected default port 9999, got %d", m.config.Port)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			Port:            8888,
			StartupTimeout:  5 * time.Second,
			ShutdownTimeout: 3 * time.Second,
		}
		m := New(config)
		if m.config.Port != 8888 {
			t.Errorf("expected port 8888, got %d", m.config.Port)
		}
	})
}

func TestManager_InitialState(t *testing.T) {
	m := New(nil)

	if m.Status() != StatusStopped {
		t.Errorf("expected initial status stopped, got %s", m.Status())
	}
	if m.IsRunning() {
		t.Error("expected IsRunning() to be false initially")
	}
	if m.PID() != 0 {
		t.Errorf("expected initial PID 0, got %d", m.PID())
	}
	if m.Uptime() != 0 {
		t.Errorf("expected initial uptime 0, got %v", m.Uptime())
	}
	if m.LastError() != nil {
		t.Errorf("expected no initial error, got %v", m.LastError())
	}
}

func TestManager_Port(t *testing.T) {
	m := New(nil)

	if m.Port() != 9999 {
		t.Errorf("expected default port 9999, got %d", m.Port())
	}

	m.SetPort(8080)
	if m.Port() != 8080 {
		t.Errorf("expected port 8080, got %d", m.Port())
	}
}

func TestManager_Info(t *testing.T) {
	m := New(nil)

	info := m.Info()
	if info == nil {
		t.Fatal("expected non-nil info")
	}
	if info.Status != StatusStopped {
		t.Errorf("expected status stopped, got %s", info.Status)
	}
	if info.Port != 9999 {
		t.Errorf("expected port 9999, got %d", info.Port)
	}
	if info.PID != 0 {
		t.Errorf("expected PID 0, got %d", info.PID)
	}
	if info.Uptime != 0 {
		t.Errorf("expected uptime 0, got %v", info.Uptime)
	}
	if info.LastError != "" {
		t.Errorf("expected empty last error, got %s", info.LastError)
	}
}

func TestManager_Start_BinaryNotFound(t *testing.T) {
	config := &Config{
		Port:           9999,
		BinaryPath:     "/nonexistent/path/to/daemon",
		StartupTimeout: 1 * time.Second,
	}
	m := New(config)

	err := m.Start()
	if err == nil {
		t.Fatal("expected error when binary not found")
	}

	if m.Status() != StatusError {
		t.Errorf("expected status error, got %s", m.Status())
	}
	if m.LastError() == nil {
		t.Error("expected last error to be set")
	}
}

func TestManager_StopWhenNotRunning(t *testing.T) {
	m := New(nil)

	// Should not error when stopping a stopped daemon
	err := m.Stop()
	if err != nil {
		t.Errorf("expected no error stopping stopped daemon, got %v", err)
	}
}

func TestManager_getDaemonSearchPaths(t *testing.T) {
	m := New(nil)
	execDir := "/app/bin"
	binaryName := "mtga-tracker-daemon"

	paths := m.getDaemonSearchPaths(execDir, binaryName)

	if len(paths) == 0 {
		t.Error("expected at least one search path")
	}

	// Should include development paths
	foundDevPath := false
	for _, p := range paths {
		if filepath.Base(filepath.Dir(p)) == "daemon" {
			foundDevPath = true
			break
		}
	}
	if !foundDevPath {
		t.Error("expected daemon search path to be included")
	}
}

func TestManager_getDaemonSearchPaths_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	m := New(nil)
	// Simulate app bundle structure: /Applications/MTGA-Companion.app/Contents/MacOS
	execDir := "/Applications/MTGA-Companion.app/Contents/MacOS"
	binaryName := "mtga-tracker-daemon"

	paths := m.getDaemonSearchPaths(execDir, binaryName)

	// Should include Resources/daemon path
	resourcesPath := "/Applications/MTGA-Companion.app/Contents/Resources/daemon/mtga-tracker-daemon"
	found := false
	for _, p := range paths {
		if p == resourcesPath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Resources/daemon path in search paths, got %v", paths)
	}
}

func TestManager_findDaemonBinary_Configured(t *testing.T) {
	// Create a temporary file to act as the binary
	tmpFile, err := os.CreateTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	m := New(&Config{
		BinaryPath: tmpFile.Name(),
	})

	path, err := m.findDaemonBinary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != tmpFile.Name() {
		t.Errorf("expected %s, got %s", tmpFile.Name(), path)
	}
}

func TestManager_findDaemonBinary_ConfiguredNotFound(t *testing.T) {
	m := New(&Config{
		BinaryPath: "/nonexistent/binary",
	})

	_, err := m.findDaemonBinary()
	if err == nil {
		t.Error("expected error when configured binary not found")
	}
}

// TestManager_StartWithMockBinary tests starting with a mock binary.
// This test creates a simple shell script that acts as the daemon.
func TestManager_StartWithMockBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping mock binary test on Windows")
	}

	// Create a temporary directory for the mock binary
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock daemon script that sleeps briefly then exits
	mockBinary := filepath.Join(tmpDir, "mtga-tracker-daemon")
	script := `#!/bin/sh
sleep 1
exit 0
`
	if err := os.WriteFile(mockBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write mock binary: %v", err)
	}

	// Verify the script is executable
	cmd := exec.Command("sh", "-c", "test -x "+mockBinary)
	if err := cmd.Run(); err != nil {
		t.Fatalf("mock binary is not executable: %v", err)
	}

	var logBuf bytes.Buffer
	config := &Config{
		Port:            19999, // Use non-standard port for testing
		BinaryPath:      mockBinary,
		StartupTimeout:  2 * time.Second,
		ShutdownTimeout: 2 * time.Second,
		LogOutput:       &logBuf,
	}

	m := New(config)

	// Start the daemon
	if err := m.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	// Wait for status to become running
	time.Sleep(800 * time.Millisecond)

	status := m.Status()
	if status != StatusRunning && status != StatusStarting {
		t.Errorf("expected status running or starting, got %s", status)
	}

	// Clean up
	if err := m.Stop(); err != nil {
		t.Errorf("failed to stop daemon: %v", err)
	}
}

func TestManager_StartAlreadyRunning(t *testing.T) {
	m := New(nil)

	// Simulate running state
	m.mu.Lock()
	m.status = StatusRunning
	m.mu.Unlock()

	err := m.Start()
	if err == nil {
		t.Error("expected error when starting already running daemon")
	}
}

func TestManager_StatusTransitions(t *testing.T) {
	m := New(nil)

	// Initial state
	if m.Status() != StatusStopped {
		t.Errorf("expected initial status stopped, got %s", m.Status())
	}

	// Simulate status transitions
	transitions := []struct {
		from Status
		to   Status
	}{
		{StatusStopped, StatusStarting},
		{StatusStarting, StatusRunning},
		{StatusRunning, StatusStopping},
		{StatusStopping, StatusStopped},
	}

	for _, tr := range transitions {
		m.mu.Lock()
		m.status = tr.from
		m.mu.Unlock()

		m.mu.Lock()
		m.status = tr.to
		m.mu.Unlock()

		if m.Status() != tr.to {
			t.Errorf("expected status %s, got %s", tr.to, m.Status())
		}
	}
}

func TestManager_Uptime_NotRunning(t *testing.T) {
	m := New(nil)

	// Should return 0 when not running
	if m.Uptime() != 0 {
		t.Errorf("expected uptime 0 when not running, got %v", m.Uptime())
	}
}

func TestManager_Uptime_Running(t *testing.T) {
	m := New(nil)

	// Simulate running state with start time
	m.mu.Lock()
	m.status = StatusRunning
	m.startTime = time.Now().Add(-5 * time.Second)
	m.mu.Unlock()

	uptime := m.Uptime()
	if uptime < 4*time.Second || uptime > 6*time.Second {
		t.Errorf("expected uptime around 5s, got %v", uptime)
	}
}

func TestManager_Restart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping mock binary test on Windows")
	}

	// Create a temporary directory for the mock binary
	tmpDir, err := os.MkdirTemp("", "daemon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock daemon script
	mockBinary := filepath.Join(tmpDir, "mtga-tracker-daemon")
	script := `#!/bin/sh
sleep 60
`
	if err := os.WriteFile(mockBinary, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write mock binary: %v", err)
	}

	config := &Config{
		Port:            19998,
		BinaryPath:      mockBinary,
		StartupTimeout:  2 * time.Second,
		ShutdownTimeout: 2 * time.Second,
	}

	m := New(config)

	// Start
	if err := m.Start(); err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}

	time.Sleep(800 * time.Millisecond)
	originalPID := m.PID()

	// Restart
	if err := m.Restart(); err != nil {
		t.Fatalf("failed to restart daemon: %v", err)
	}

	time.Sleep(800 * time.Millisecond)
	newPID := m.PID()

	// PIDs should be different after restart
	if newPID == originalPID && originalPID != 0 {
		t.Error("expected different PID after restart")
	}

	// Clean up
	if err := m.Stop(); err != nil {
		t.Errorf("failed to stop daemon: %v", err)
	}
}

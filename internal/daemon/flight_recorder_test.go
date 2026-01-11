package daemon

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFlightRecorder(t *testing.T) {
	fr := NewFlightRecorder(DefaultFlightRecorderConfig())
	if fr == nil {
		t.Fatal("Expected non-nil FlightRecorder")
	}
	if fr.recorder == nil {
		t.Fatal("Expected non-nil underlying recorder")
	}
}

func TestFlightRecorder_DefaultConfig(t *testing.T) {
	config := DefaultFlightRecorderConfig()

	if !config.Enabled {
		t.Error("Expected Enabled to be true by default")
	}
	if config.MinAge != 10*time.Second {
		t.Errorf("Expected MinAge 10s, got %v", config.MinAge)
	}
	if config.MaxBytes != 10*1024*1024 {
		t.Errorf("Expected MaxBytes 10MB, got %d", config.MaxBytes)
	}
	if config.MaxTraceFiles != 5 {
		t.Errorf("Expected MaxTraceFiles 5, got %d", config.MaxTraceFiles)
	}
}

func TestFlightRecorder_StartStop(t *testing.T) {
	fr := NewFlightRecorder(DefaultFlightRecorderConfig())

	// Start
	err := fr.Start()
	if err != nil {
		t.Fatalf("Failed to start flight recorder: %v", err)
	}

	if !fr.Enabled() {
		t.Error("Expected Enabled() to return true after Start()")
	}

	// Double start should be no-op
	err = fr.Start()
	if err != nil {
		t.Fatalf("Double start should not error: %v", err)
	}

	// Stop
	fr.Stop()

	if fr.Enabled() {
		t.Error("Expected Enabled() to return false after Stop()")
	}

	// Double stop should be no-op
	fr.Stop()
}

func TestFlightRecorder_DisabledConfig(t *testing.T) {
	config := DefaultFlightRecorderConfig()
	config.Enabled = false

	fr := NewFlightRecorder(config)

	// Start should succeed but not actually start
	err := fr.Start()
	if err != nil {
		t.Fatalf("Start with disabled config should not error: %v", err)
	}

	if fr.Enabled() {
		t.Error("Expected Enabled() to return false when config.Enabled is false")
	}
}

func TestFlightRecorder_WriteTo(t *testing.T) {
	fr := NewFlightRecorder(DefaultFlightRecorderConfig())

	// WriteTo before start should error
	var buf bytes.Buffer
	_, err := fr.WriteTo(&buf)
	if err == nil {
		t.Error("Expected error when WriteTo called before Start")
	}

	// Start and write
	if err := fr.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer fr.Stop()

	// Do some work to generate trace data
	for i := 0; i < 1000; i++ {
		_ = i * i
	}

	n, err := fr.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if n == 0 {
		t.Error("Expected some trace data to be written")
	}
}

func TestFlightRecorder_CaptureTrace(t *testing.T) {
	// Use temp directory for output
	tempDir := t.TempDir()

	config := DefaultFlightRecorderConfig()
	config.OutputDir = tempDir

	fr := NewFlightRecorder(config)

	// CaptureTrace before start should error
	_, err := fr.CaptureTrace("test")
	if err == nil {
		t.Error("Expected error when CaptureTrace called before Start")
	}

	// Start
	if err := fr.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer fr.Stop()

	// Do some work
	for i := 0; i < 1000; i++ {
		_ = i * i
	}

	// Capture trace
	path, err := fr.CaptureTrace("test-error")
	if err != nil {
		t.Fatalf("CaptureTrace failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Trace file not created: %s", path)
	}

	// Verify filename format
	filename := filepath.Base(path)
	if len(filename) < 20 {
		t.Errorf("Filename too short: %s", filename)
	}
	if filepath.Ext(filename) != ".out" {
		t.Errorf("Expected .out extension, got %s", filepath.Ext(filename))
	}
}

func TestFlightRecorder_CleanupOldTraces(t *testing.T) {
	tempDir := t.TempDir()

	config := DefaultFlightRecorderConfig()
	config.OutputDir = tempDir
	config.MaxTraceFiles = 2

	fr := NewFlightRecorder(config)

	if err := fr.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer fr.Stop()

	// Create 4 trace files with unique reasons to ensure unique filenames
	for i := 0; i < 4; i++ {
		reason := fmt.Sprintf("test-%d-%d", i, time.Now().UnixNano())
		_, err := fr.CaptureTrace(reason)
		if err != nil {
			t.Fatalf("CaptureTrace failed: %v", err)
		}
	}

	// Check that only MaxTraceFiles remain
	pattern := filepath.Join(tempDir, "mtga-companion-trace-*.out")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to glob trace files: %v", err)
	}

	if len(files) != config.MaxTraceFiles {
		t.Errorf("Expected %d trace files, got %d", config.MaxTraceFiles, len(files))
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with_spaces"},
		{"with-dashes", "with-dashes"},
		{"with_underscores", "with_underscores"},
		{"MixedCase123", "MixedCase123"},
		{"special!@#$chars", "specialchars"},
		{"", "unknown"},
		{"a", "a"},
		{"verylongstringthatexceedsfiftycharacterlimitandshouldbetruncated", "verylongstringthatexceedsfiftycharacterlimitandsho"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

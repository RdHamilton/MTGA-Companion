package daemon

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/trace"
	"sync"
	"time"
)

// FlightRecorderConfig configures the flight recorder behavior.
type FlightRecorderConfig struct {
	// Enabled controls whether flight recording is active.
	Enabled bool

	// MinAge is the minimum age of events to keep in the buffer.
	// Default: 10 seconds.
	MinAge time.Duration

	// MaxBytes is the maximum size of the trace buffer.
	// Default: 10MB.
	MaxBytes uint64

	// OutputDir is the directory where trace files are written.
	// Default: system temp directory.
	OutputDir string

	// MaxTraceFiles is the maximum number of trace files to keep.
	// Older files are deleted when this limit is exceeded.
	// Default: 5.
	MaxTraceFiles int
}

// DefaultFlightRecorderConfig returns sensible defaults for flight recording.
func DefaultFlightRecorderConfig() FlightRecorderConfig {
	return FlightRecorderConfig{
		Enabled:       true,
		MinAge:        10 * time.Second,
		MaxBytes:      10 * 1024 * 1024, // 10MB
		OutputDir:     "",               // Use temp directory
		MaxTraceFiles: 5,
	}
}

// FlightRecorder wraps runtime/trace.FlightRecorder with additional functionality.
type FlightRecorder struct {
	config   FlightRecorderConfig
	recorder *trace.FlightRecorder
	mu       sync.Mutex
	started  bool
}

// NewFlightRecorder creates a new flight recorder with the given configuration.
func NewFlightRecorder(config FlightRecorderConfig) *FlightRecorder {
	if config.MinAge == 0 {
		config.MinAge = 10 * time.Second
	}
	if config.MaxBytes == 0 {
		config.MaxBytes = 10 * 1024 * 1024
	}
	if config.MaxTraceFiles == 0 {
		config.MaxTraceFiles = 5
	}
	if config.OutputDir == "" {
		config.OutputDir = os.TempDir()
	}

	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   config.MinAge,
		MaxBytes: config.MaxBytes,
	})

	return &FlightRecorder{
		config:   config,
		recorder: fr,
	}
}

// Start begins flight recording.
func (fr *FlightRecorder) Start() error {
	if !fr.config.Enabled {
		return nil
	}

	fr.mu.Lock()
	defer fr.mu.Unlock()

	if fr.started {
		return nil
	}

	if err := fr.recorder.Start(); err != nil {
		return fmt.Errorf("failed to start flight recorder: %w", err)
	}

	fr.started = true
	log.Println("Flight recorder started")
	return nil
}

// Stop stops flight recording.
func (fr *FlightRecorder) Stop() {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	if !fr.started {
		return
	}

	fr.recorder.Stop()
	fr.started = false
	log.Println("Flight recorder stopped")
}

// Enabled returns whether the flight recorder is currently recording.
func (fr *FlightRecorder) Enabled() bool {
	fr.mu.Lock()
	defer fr.mu.Unlock()
	return fr.started && fr.recorder.Enabled()
}

// CaptureTrace writes the current trace buffer to a file.
// Returns the path to the trace file, or an error if capture failed.
func (fr *FlightRecorder) CaptureTrace(reason string) (string, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	if !fr.started {
		return "", fmt.Errorf("flight recorder not started")
	}

	// Generate filename with timestamp and reason
	timestamp := time.Now().Format("20060102-150405")
	safeReason := sanitizeFilename(reason)
	filename := fmt.Sprintf("trace-%s-%s.out", timestamp, safeReason)
	tracePath := filepath.Join(fr.config.OutputDir, filename)

	// Create trace file
	f, err := os.Create(tracePath)
	if err != nil {
		return "", fmt.Errorf("failed to create trace file: %w", err)
	}

	// Write trace data
	n, writeErr := fr.recorder.WriteTo(f)

	// Close the file and check for errors
	closeErr := f.Close()

	if writeErr != nil {
		return "", fmt.Errorf("failed to write trace: %w", writeErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("failed to close trace file: %w", closeErr)
	}

	log.Printf("Captured trace: %s (%d bytes, reason: %s)", tracePath, n, reason)

	// Clean up old trace files
	fr.cleanupOldTraces()

	return tracePath, nil
}

// WriteTo writes the current trace buffer to the given writer.
func (fr *FlightRecorder) WriteTo(w io.Writer) (int64, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	if !fr.started {
		return 0, fmt.Errorf("flight recorder not started")
	}

	return fr.recorder.WriteTo(w)
}

// cleanupOldTraces removes old trace files exceeding MaxTraceFiles limit.
func (fr *FlightRecorder) cleanupOldTraces() {
	pattern := filepath.Join(fr.config.OutputDir, "trace-*.out")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	if len(files) <= fr.config.MaxTraceFiles {
		return
	}

	// Sort files by name (which includes timestamp, so oldest first)
	// Note: files are already sorted alphabetically by Glob
	toDelete := len(files) - fr.config.MaxTraceFiles
	for i := 0; i < toDelete; i++ {
		if err := os.Remove(files[i]); err != nil {
			log.Printf("Failed to remove old trace file %s: %v", files[i], err)
		} else {
			log.Printf("Removed old trace file: %s", files[i])
		}
	}
}

// sanitizeFilename removes characters that are invalid in filenames.
func sanitizeFilename(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "unknown"
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return string(result)
}

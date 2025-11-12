package draft

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout during function execution.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestNewLogger(t *testing.T) {
	t.Run("DebugEnabled", func(t *testing.T) {
		logger := NewLogger(true)
		if logger == nil {
			t.Fatal("NewLogger() returned nil")
		}
		if !logger.IsDebugEnabled() {
			t.Error("Expected debug to be enabled")
		}
	})

	t.Run("DebugDisabled", func(t *testing.T) {
		logger := NewLogger(false)
		if logger == nil {
			t.Fatal("NewLogger() returned nil")
		}
		if logger.IsDebugEnabled() {
			t.Error("Expected debug to be disabled")
		}
	})
}

func TestLogger_Debug(t *testing.T) {
	t.Run("DebugEnabled", func(t *testing.T) {
		logger := NewLogger(true)

		output := captureOutput(func() {
			logger.Debug("test message")
		})

		if !strings.Contains(output, "[DEBUG]") {
			t.Error("Expected [DEBUG] prefix in output")
		}
		if !strings.Contains(output, "test message") {
			t.Error("Expected message in output")
		}
	})

	t.Run("DebugDisabled", func(t *testing.T) {
		logger := NewLogger(false)

		output := captureOutput(func() {
			logger.Debug("test message")
		})

		if output != "" {
			t.Errorf("Expected no output, got: %s", output)
		}
	})

	t.Run("WithFormatting", func(t *testing.T) {
		logger := NewLogger(true)

		output := captureOutput(func() {
			logger.Debug("Pack %d, Pick %d", 2, 8)
		})

		if !strings.Contains(output, "Pack 2, Pick 8") {
			t.Errorf("Expected formatted message in output, got: %s", output)
		}
	})
}

func TestLogger_Info(t *testing.T) {
	t.Run("AlwaysShown", func(t *testing.T) {
		logger := NewLogger(false) // Debug disabled

		output := captureOutput(func() {
			logger.Info("info message")
		})

		if !strings.Contains(output, "[INFO]") {
			t.Error("Expected [INFO] prefix in output")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Expected message in output")
		}
	})

	t.Run("WithFormatting", func(t *testing.T) {
		logger := NewLogger(false)

		output := captureOutput(func() {
			logger.Info("Draft complete! %d picks made", 45)
		})

		if !strings.Contains(output, "Draft complete! 45 picks made") {
			t.Errorf("Expected formatted message in output, got: %s", output)
		}
	})
}

func TestLogger_Error(t *testing.T) {
	t.Run("AlwaysShown", func(t *testing.T) {
		logger := NewLogger(false) // Debug disabled

		output := captureOutput(func() {
			logger.Error("error message")
		})

		if !strings.Contains(output, "[ERROR]") {
			t.Error("Expected [ERROR] prefix in output")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Expected message in output")
		}
	})

	t.Run("WithFormatting", func(t *testing.T) {
		logger := NewLogger(false)

		output := captureOutput(func() {
			logger.Error("Failed to load card %d: %s", 89001, "not found")
		})

		if !strings.Contains(output, "Failed to load card 89001: not found") {
			t.Errorf("Expected formatted message in output, got: %s", output)
		}
	})
}

func TestLogger_DebugVsInfo(t *testing.T) {
	t.Run("DebugOffShowsInfoOnly", func(t *testing.T) {
		logger := NewLogger(false)

		output := captureOutput(func() {
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Error("error message")
		})

		if strings.Contains(output, "debug message") {
			t.Error("Debug message should not appear when debug is off")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Info message should always appear")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should always appear")
		}
	})

	t.Run("DebugOnShowsAll", func(t *testing.T) {
		logger := NewLogger(true)

		output := captureOutput(func() {
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Error("error message")
		})

		if !strings.Contains(output, "debug message") {
			t.Error("Debug message should appear when debug is on")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Info message should always appear")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should always appear")
		}
	})
}

package draft

import (
	"fmt"
	"time"
)

// Logger provides leveled logging for draft overlay.
type Logger struct {
	debugEnabled bool
}

// NewLogger creates a new logger with the specified debug mode.
func NewLogger(debugEnabled bool) *Logger {
	return &Logger{
		debugEnabled: debugEnabled,
	}
}

// Debug logs a debug message (only if debug mode is enabled).
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.debugEnabled {
		return
	}
	timestamp := time.Now().Format("15:04:05.000")
	fmt.Printf("[DEBUG] %s - %s\n", timestamp, fmt.Sprintf(format, args...))
}

// Info logs an informational message (always shown).
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] %s\n", fmt.Sprintf(format, args...))
}

// Error logs an error message (always shown).
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] %s\n", fmt.Sprintf(format, args...))
}

// IsDebugEnabled returns whether debug mode is enabled.
func (l *Logger) IsDebugEnabled() bool {
	return l.debugEnabled
}

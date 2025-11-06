package logreader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultLogPath returns the default Player.log path for the current platform.
// It returns an error if the platform is unsupported or the log directory doesn't exist.
func DefaultLogPath() (string, error) {
	var logPath string

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home directory: %w", err)
		}
		logPath = filepath.Join(home, "Library", "Application Support", "com.wizards.mtga", "Logs", "Logs", "Player.log")

	case "windows":
		// Windows: C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home directory: %w", err)
		}
		logPath = filepath.Join(home, "AppData", "LocalLow", "Wizards Of The Coast", "MTGA", "Player.log")

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return logPath, nil
}

// LogExists checks if the log file exists at the given path.
func LogExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat log file: %w", err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("path is a directory, not a file")
	}
	return true, nil
}

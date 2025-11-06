package logreader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// DefaultLogPath returns the most recent MTGA log file path for the current platform.
// It checks multiple locations and prioritizes UTC_Log files over Player.log.
// It returns an error if the platform is unsupported or no log files are found.
func DefaultLogPath() (string, error) {
	logDirs := getLogDirectories()

	// Try to find the most recent UTC_Log file in all directories
	for _, logDir := range logDirs {
		utcLogPath, err := findMostRecentUTCLog(logDir)
		if err == nil {
			return utcLogPath, nil
		}
	}

	// Fall back to Player.log in any of the directories
	for _, logDir := range logDirs {
		playerLogPath := filepath.Join(logDir, "Player.log")
		exists, err := LogExists(playerLogPath)
		if err == nil && exists {
			return playerLogPath, nil
		}
	}

	return "", fmt.Errorf("no log files found in any MTGA log directories")
}

// getLogDirectories returns possible MTGA log directories for the current platform.
// Directories are returned in priority order.
func getLogDirectories() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS has multiple possible locations
		return []string{
			filepath.Join(home, "Library", "Application Support", "com.wizards.mtga", "Logs", "Logs"),
			filepath.Join(home, "Library", "Logs", "Wizards of the Coast", "MTGA"),
		}

	case "windows":
		// Windows: C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\
		return []string{
			filepath.Join(home, "AppData", "LocalLow", "Wizards Of The Coast", "MTGA"),
		}

	default:
		return nil
	}
}

// findMostRecentUTCLog finds the most recent UTC_Log file in the given directory.
func findMostRecentUTCLog(logDir string) (string, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "", fmt.Errorf("read log directory: %w", err)
	}

	var utcLogs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "UTC_Log") && strings.HasSuffix(name, ".log") {
			utcLogs = append(utcLogs, filepath.Join(logDir, name))
		}
	}

	if len(utcLogs) == 0 {
		return "", fmt.Errorf("no UTC_Log files found")
	}

	// Sort by modification time (most recent first)
	sort.Slice(utcLogs, func(i, j int) bool {
		infoI, errI := os.Stat(utcLogs[i])
		infoJ, errJ := os.Stat(utcLogs[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return utcLogs[0], nil
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

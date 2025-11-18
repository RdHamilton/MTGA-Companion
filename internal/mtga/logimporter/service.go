package logimporter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logprocessor"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// LogFileInfo contains information about a discovered log file.
type LogFileInfo struct {
	Path     string
	Name     string
	Size     int64
	ModTime  time.Time
	Selected bool
}

// ImportProgress tracks the progress of a historical import.
type ImportProgress struct {
	mu                sync.RWMutex
	TotalFiles        int
	ProcessedFiles    int
	CurrentFile       string
	TotalEntries      int
	ProcessedEntries  int
	MatchesImported   int
	GamesImported     int
	DecksImported     int
	RanksImported     int
	QuestsImported    int
	StartTime         time.Time
	EstimatedTimeLeft time.Duration
	Errors            []string
	Status            string // "idle", "running", "completed", "failed", "cancelled"
}

// Service handles historical log file import.
type Service struct {
	logProcessor *logprocessor.Service
	progress     *ImportProgress
	cancel       context.CancelFunc
	mu           sync.RWMutex
}

// NewService creates a new historical import service.
func NewService(logProcessor *logprocessor.Service) *Service {
	return &Service{
		logProcessor: logProcessor,
		progress: &ImportProgress{
			Status: "idle",
		},
	}
}

// DiscoverLogFiles finds all MTGA log files in the log directory.
func (s *Service) DiscoverLogFiles() ([]*LogFileInfo, error) {
	logDirs := getLogDirectories()

	var allFiles []*LogFileInfo

	for _, logDir := range logDirs {
		entries, err := os.ReadDir(logDir)
		if err != nil {
			// Skip directories that don't exist
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			// Match: Player.log, Player-prev.log, UTC_Log*.log
			if !isLogFile(name) {
				continue
			}

			path := filepath.Join(logDir, name)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			allFiles = append(allFiles, &LogFileInfo{
				Path:     path,
				Name:     name,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
				Selected: true, // Default to selected
			})
		}
	}

	// Sort by modification time (oldest first for chronological import)
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].ModTime.Before(allFiles[j].ModTime)
	})

	return allFiles, nil
}

// isLogFile returns true if the filename is a recognized MTGA log file.
func isLogFile(name string) bool {
	// Player.log, Player-prev.log, UTC_Log*.log
	if name == "Player.log" || name == "Player-prev.log" {
		return true
	}
	if strings.HasPrefix(name, "UTC_Log") && strings.HasSuffix(name, ".log") {
		return true
	}
	return false
}

// getLogDirectories returns possible MTGA log directories.
func getLogDirectories() []string {
	// Reuse the logic from logreader package
	// This is a simplified version - we could import from logreader if it's exported
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var dirs []string
	switch {
	case fileExists(filepath.Join(home, "Library", "Application Support", "com.wizards.mtga", "Logs", "Logs")):
		// macOS
		dirs = append(dirs,
			filepath.Join(home, "Library", "Application Support", "com.wizards.mtga", "Logs", "Logs"),
			filepath.Join(home, "Library", "Logs", "Wizards of the Coast", "MTGA"),
		)
	case fileExists(filepath.Join(home, "AppData", "LocalLow", "Wizards Of The Coast", "MTGA")):
		// Windows
		dirs = append(dirs, filepath.Join(home, "AppData", "LocalLow", "Wizards Of The Coast", "MTGA"))
	}

	return dirs
}

// fileExists checks if a path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetProgress returns the current import progress.
func (s *Service) GetProgress() *ImportProgress {
	s.progress.mu.RLock()
	defer s.progress.mu.RUnlock()

	// Create a copy to avoid race conditions
	return &ImportProgress{
		TotalFiles:        s.progress.TotalFiles,
		ProcessedFiles:    s.progress.ProcessedFiles,
		CurrentFile:       s.progress.CurrentFile,
		TotalEntries:      s.progress.TotalEntries,
		ProcessedEntries:  s.progress.ProcessedEntries,
		MatchesImported:   s.progress.MatchesImported,
		GamesImported:     s.progress.GamesImported,
		DecksImported:     s.progress.DecksImported,
		RanksImported:     s.progress.RanksImported,
		QuestsImported:    s.progress.QuestsImported,
		StartTime:         s.progress.StartTime,
		EstimatedTimeLeft: s.progress.EstimatedTimeLeft,
		Errors:            append([]string{}, s.progress.Errors...),
		Status:            s.progress.Status,
	}
}

// StartImport begins importing the selected log files.
func (s *Service) StartImport(ctx context.Context, files []*LogFileInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already running
	if s.progress.Status == "running" {
		return fmt.Errorf("import already in progress")
	}

	// Create cancellable context
	importCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Reset progress
	s.progress = &ImportProgress{
		TotalFiles: len(files),
		StartTime:  time.Now(),
		Status:     "running",
		Errors:     []string{},
	}

	// Start import in background
	go s.runImport(importCtx, files)

	return nil
}

// CancelImport cancels the current import operation.
func (s *Service) CancelImport() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.progress.Status != "running" {
		return fmt.Errorf("no import in progress")
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.progress.mu.Lock()
	s.progress.Status = "cancelled"
	s.progress.mu.Unlock()

	return nil
}

// runImport performs the actual import operation.
func (s *Service) runImport(ctx context.Context, files []*LogFileInfo) {
	defer func() {
		s.progress.mu.Lock()
		if s.progress.Status == "running" {
			s.progress.Status = "completed"
		}
		s.progress.mu.Unlock()
	}()

	for i, file := range files {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Update progress
		s.progress.mu.Lock()
		s.progress.CurrentFile = file.Name
		s.progress.ProcessedFiles = i
		s.progress.mu.Unlock()

		// Process this log file
		if err := s.processLogFile(ctx, file.Path); err != nil {
			s.progress.mu.Lock()
			s.progress.Errors = append(s.progress.Errors, fmt.Sprintf("%s: %v", file.Name, err))
			s.progress.mu.Unlock()
			continue
		}

		// Update file completion
		s.progress.mu.Lock()
		s.progress.ProcessedFiles = i + 1
		s.progress.mu.Unlock()

		// Update time estimate
		s.updateTimeEstimate()
	}
}

// processLogFile reads and processes a single log file.
func (s *Service) processLogFile(ctx context.Context, path string) error {
	// Create a poller that reads from start
	config := &logreader.PollerConfig{
		Path:          path,
		Interval:      100 * time.Millisecond, // Fast for batch reading
		BufferSize:    1000,                   // Large buffer for batch processing
		UseFileEvents: false,                  // Don't need file events for historical files
		ReadFromStart: true,                   // Read entire file
	}

	poller, err := logreader.NewPoller(config)
	if err != nil {
		return fmt.Errorf("create poller: %w", err)
	}

	// Start polling - returns the update channel
	updates := poller.Start()
	defer poller.Stop()

	// Collect entries in batches
	var batch []*logreader.LogEntry
	const batchSize = 1000

	timeout := time.After(30 * time.Second) // Timeout if no updates for 30s

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case entry, ok := <-updates:
			if !ok {
				// Channel closed, process final batch
				if len(batch) > 0 {
					if err := s.processBatch(ctx, batch); err != nil {
						return err
					}
				}
				return nil
			}

			batch = append(batch, entry)

			// Process batch when it reaches size limit
			if len(batch) >= batchSize {
				if err := s.processBatch(ctx, batch); err != nil {
					return err
				}
				batch = nil
				timeout = time.After(30 * time.Second) // Reset timeout
			}

		case <-timeout:
			// No updates for 30s, assume file is done
			if len(batch) > 0 {
				if err := s.processBatch(ctx, batch); err != nil {
					return err
				}
			}
			return nil
		}
	}
}

// processBatch processes a batch of log entries.
func (s *Service) processBatch(ctx context.Context, entries []*logreader.LogEntry) error {
	result, err := s.logProcessor.ProcessLogEntries(ctx, entries)
	if err != nil {
		return err
	}

	// Update progress with results
	s.progress.mu.Lock()
	s.progress.ProcessedEntries += len(entries)
	s.progress.MatchesImported += result.MatchesStored
	s.progress.GamesImported += result.GamesStored
	s.progress.DecksImported += result.DecksStored
	s.progress.RanksImported += result.RanksStored
	s.progress.QuestsImported += result.QuestsStored
	s.progress.mu.Unlock()

	return nil
}

// updateTimeEstimate calculates estimated time remaining.
func (s *Service) updateTimeEstimate() {
	s.progress.mu.Lock()
	defer s.progress.mu.Unlock()

	if s.progress.ProcessedFiles == 0 {
		return
	}

	elapsed := time.Since(s.progress.StartTime)
	filesRemaining := s.progress.TotalFiles - s.progress.ProcessedFiles
	if filesRemaining <= 0 {
		s.progress.EstimatedTimeLeft = 0
		return
	}

	avgTimePerFile := elapsed / time.Duration(s.progress.ProcessedFiles)
	s.progress.EstimatedTimeLeft = avgTimePerFile * time.Duration(filesRemaining)
}

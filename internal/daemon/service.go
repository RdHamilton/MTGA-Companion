package daemon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logprocessor"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Version is the daemon version
const Version = "1.0.0"

// Service represents the daemon service that runs continuously.
type Service struct {
	config       *Config
	storage      *storage.Service
	logProcessor *logprocessor.Service
	poller       *logreader.Poller
	wsServer     *WebSocketServer
	ctx          context.Context
	cancel       context.CancelFunc
	startTime    time.Time

	// Health tracking
	healthMu       sync.RWMutex
	lastLogRead    time.Time
	lastDBWrite    time.Time
	totalProcessed int64
	totalErrors    int64
}

// New creates a new daemon service.
func New(config *Config, storage *storage.Service) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		config:       config,
		storage:      storage,
		logProcessor: logprocessor.NewService(storage),
		wsServer:     NewWebSocketServer(config.Port),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the daemon service.
func (s *Service) Start() error {
	s.startTime = time.Now()
	log.Println("Starting MTGA Companion daemon...")

	// Determine log path
	logPath := s.config.LogPath
	if logPath == "" {
		detected, err := logreader.DefaultLogPath()
		if err != nil {
			return fmt.Errorf("failed to detect log path: %w", err)
		}
		logPath = detected
		log.Printf("Auto-detected log path: %s", logPath)
	}

	// Create and start log poller
	pollerConfig := logreader.DefaultPollerConfig(logPath)
	pollerConfig.Interval = s.config.PollInterval
	pollerConfig.UseFileEvents = s.config.UseFSNotify
	pollerConfig.EnableMetrics = s.config.EnableMetrics
	pollerConfig.ReadFromStart = true // Read entire log file on startup

	poller, err := logreader.NewPoller(pollerConfig)
	if err != nil {
		return fmt.Errorf("failed to create log poller: %w", err)
	}

	s.poller = poller

	// Set service reference for health checks
	s.wsServer.SetService(s)

	// Start WebSocket server in background
	go func() {
		if err := s.wsServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// Wait for WebSocket server to start
	time.Sleep(100 * time.Millisecond)

	// Start log poller
	updates := s.poller.Start()
	errChan := s.poller.Errors()

	log.Printf("Daemon started successfully")
	log.Printf("WebSocket server: ws://localhost:%d", s.config.Port)
	log.Printf("Status endpoint: http://localhost:%d/status", s.config.Port)

	// Send status event
	s.wsServer.Broadcast(Event{
		Type: "daemon:status",
		Data: map[string]interface{}{
			"status":  "running",
			"port":    s.config.Port,
			"logPath": logPath,
		},
	})

	// Process log updates
	go s.processUpdates(updates, errChan)

	// Send periodic status updates
	go s.sendPeriodicStatus()

	return nil
}

// Stop gracefully stops the daemon service.
func (s *Service) Stop() error {
	log.Println("Stopping daemon...")

	// Cancel context
	s.cancel()

	// Stop poller
	if s.poller != nil {
		s.poller.Stop()
	}

	// Stop WebSocket server
	if s.wsServer != nil {
		if err := s.wsServer.Stop(); err != nil {
			log.Printf("Error stopping WebSocket server: %v", err)
		}
	}

	log.Println("Daemon stopped")
	return nil
}

// processUpdates processes log updates and broadcasts events.
func (s *Service) processUpdates(updates <-chan *logreader.LogEntry, errChan <-chan error) {
	var entryBuffer []*logreader.LogEntry
	ticker := time.NewTicker(5 * time.Second) // Batch process every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case entry, ok := <-updates:
			if !ok {
				return
			}
			// Buffer entries for batch processing
			entryBuffer = append(entryBuffer, entry)

			// Update last log read time
			s.healthMu.Lock()
			s.lastLogRead = time.Now()
			s.healthMu.Unlock()
		case err, ok := <-errChan:
			if !ok {
				return
			}
			log.Printf("Poller error: %v", err)

			// Track error
			s.healthMu.Lock()
			s.totalErrors++
			s.healthMu.Unlock()

			// Broadcast error event
			s.wsServer.Broadcast(Event{
				Type: "daemon:error",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			})
		case <-ticker.C:
			// Process buffered entries
			if len(entryBuffer) > 0 {
				s.processEntries(entryBuffer)
				entryBuffer = nil // Clear buffer
			}
		}
	}
}

// processEntries processes a batch of log entries.
func (s *Service) processEntries(entries []*logreader.LogEntry) {
	log.Printf("Processing %d log entries...", len(entries))
	result, err := s.logProcessor.ProcessLogEntries(s.ctx, entries)
	if err != nil {
		log.Printf("Error processing log entries: %v", err)

		// Track error
		s.healthMu.Lock()
		s.totalErrors++
		s.healthMu.Unlock()
		return
	}

	// Track successful processing
	s.healthMu.Lock()
	s.totalProcessed += int64(len(entries))
	if result.MatchesStored > 0 || result.GamesStored > 0 || result.DecksStored > 0 || result.RanksStored > 0 || result.QuestsStored > 0 || result.DraftsStored > 0 {
		s.lastDBWrite = time.Now()
	}
	s.healthMu.Unlock()

	// Broadcast events for updates
	if result.MatchesStored > 0 || result.GamesStored > 0 {
		log.Printf("Stored %d matches, %d games", result.MatchesStored, result.GamesStored)
		s.wsServer.Broadcast(Event{
			Type: "stats:updated",
			Data: map[string]interface{}{
				"matches": result.MatchesStored,
				"games":   result.GamesStored,
			},
		})
	}

	if result.DecksStored > 0 {
		log.Printf("Stored %d deck(s)", result.DecksStored)
		s.wsServer.Broadcast(Event{
			Type: "deck:updated",
			Data: map[string]interface{}{
				"count": result.DecksStored,
			},
		})
	}

	if result.RanksStored > 0 {
		log.Printf("Stored %d rank update(s)", result.RanksStored)
		s.wsServer.Broadcast(Event{
			Type: "rank:updated",
			Data: map[string]interface{}{
				"count": result.RanksStored,
			},
		})
	}

	if result.QuestsStored > 0 {
		log.Printf("Stored %d quest(s)", result.QuestsStored)
		s.wsServer.Broadcast(Event{
			Type: "quest:updated",
			Data: map[string]interface{}{
				"count":     result.QuestsStored,
				"completed": result.QuestsCompleted,
			},
		})
	}

	if result.QuestsCompleted > 0 {
		log.Printf("Completed %d quest(s)", result.QuestsCompleted)
		s.wsServer.Broadcast(Event{
			Type: "quest:updated",
			Data: map[string]interface{}{
				"completed": result.QuestsCompleted,
			},
		})
	}

	if result.AchievementsStored > 0 {
		log.Printf("Stored %d achievement(s)", result.AchievementsStored)
		s.wsServer.Broadcast(Event{
			Type: "achievement:updated",
			Data: map[string]interface{}{
				"count": result.AchievementsStored,
			},
		})
	}

	if result.DraftsStored > 0 {
		log.Printf("Stored %d draft session(s) with %d picks", result.DraftsStored, result.DraftPicksStored)
		s.wsServer.Broadcast(Event{
			Type: "draft:updated",
			Data: map[string]interface{}{
				"count": result.DraftsStored,
				"picks": result.DraftPicksStored,
			},
		})
	}
}

// sendPeriodicStatus sends periodic status updates to clients.
func (s *Service) sendPeriodicStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			uptime := time.Since(s.startTime).Seconds()
			s.wsServer.Broadcast(Event{
				Type: "daemon:status",
				Data: map[string]interface{}{
					"status":  "running",
					"uptime":  uptime,
					"clients": s.wsServer.ClientCount(),
				},
			})
		}
	}
}

// GetUptime returns the daemon uptime in seconds.
func (s *Service) GetUptime() float64 {
	return time.Since(s.startTime).Seconds()
}

// GetClientCount returns the number of connected WebSocket clients.
func (s *Service) GetClientCount() int {
	return s.wsServer.ClientCount()
}

// HealthStatus represents the health status of the daemon.
type HealthStatus struct {
	Status     string           `json:"status"`
	Version    string           `json:"version"`
	Uptime     float64          `json:"uptime"`
	Database   DatabaseHealth   `json:"database"`
	LogMonitor LogMonitorHealth `json:"logMonitor"`
	WebSocket  WebSocketHealth  `json:"websocket"`
	Metrics    HealthMetrics    `json:"metrics"`
}

// DatabaseHealth represents database health status.
type DatabaseHealth struct {
	Status    string `json:"status"`
	LastWrite string `json:"lastWrite,omitempty"`
}

// LogMonitorHealth represents log monitor health status.
type LogMonitorHealth struct {
	Status   string `json:"status"`
	LastRead string `json:"lastRead,omitempty"`
}

// WebSocketHealth represents WebSocket server health status.
type WebSocketHealth struct {
	Status           string `json:"status"`
	ConnectedClients int    `json:"connectedClients"`
}

// HealthMetrics represents daemon performance metrics.
type HealthMetrics struct {
	TotalProcessed int64 `json:"totalProcessed"`
	TotalErrors    int64 `json:"totalErrors"`
}

// GetHealth returns the current health status of the daemon.
func (s *Service) GetHealth() *HealthStatus {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()

	status := &HealthStatus{
		Status:  "healthy",
		Version: Version,
		Uptime:  s.GetUptime(),
		Database: DatabaseHealth{
			Status: "ok",
		},
		LogMonitor: LogMonitorHealth{
			Status: "ok",
		},
		WebSocket: WebSocketHealth{
			Status:           "ok",
			ConnectedClients: s.GetClientCount(),
		},
		Metrics: HealthMetrics{
			TotalProcessed: s.totalProcessed,
			TotalErrors:    s.totalErrors,
		},
	}

	// Add last write time if available
	if !s.lastDBWrite.IsZero() {
		status.Database.LastWrite = s.lastDBWrite.Format(time.RFC3339)
	}

	// Add last read time if available
	if !s.lastLogRead.IsZero() {
		status.LogMonitor.LastRead = s.lastLogRead.Format(time.RFC3339)
	}

	// Determine overall health status based on component states
	now := time.Now()

	// If we haven't read logs in 5 minutes, log monitor might be unhealthy
	if !s.lastLogRead.IsZero() && now.Sub(s.lastLogRead) > 5*time.Minute {
		status.LogMonitor.Status = "warning"
		status.Status = "degraded"
	}

	// If error rate is high (>10% of processed entries), mark as degraded
	if s.totalProcessed > 0 && float64(s.totalErrors)/float64(s.totalProcessed) > 0.1 {
		status.Status = "degraded"
	}

	return status
}

// ReplayHistoricalLogs replays all historical log files through the processing pipeline.
// This is the CORRECT way to import historical data - it runs logs through the same
// business logic as real-time processing, ensuring GraphState updates, quest completion
// detection, rank progression, etc. all work correctly.
func (s *Service) ReplayHistoricalLogs(clearData bool) error {
	log.Println("Starting historical log replay...")

	// Broadcast replay start event
	s.wsServer.Broadcast(Event{
		Type: "replay:started",
		Data: map[string]interface{}{
			"clearData": clearData,
		},
	})

	// Step 1: Stop current poller
	log.Println("Stopping current log poller...")
	if s.poller != nil {
		s.poller.Stop()
	}

	// Step 2: Optionally clear all data
	if clearData {
		log.Println("Clearing all existing data...")
		if err := s.storage.ClearAllMatches(s.ctx); err != nil {
			s.wsServer.Broadcast(Event{
				Type: "replay:error",
				Data: map[string]interface{}{
					"error": fmt.Sprintf("Failed to clear data: %v", err),
				},
			})
			return fmt.Errorf("failed to clear data: %w", err)
		}
		log.Println("All data cleared successfully")
	}

	// Step 3: Discover all log files
	log.Println("Discovering log files...")
	logFiles, err := s.discoverLogFiles()
	if err != nil {
		s.wsServer.Broadcast(Event{
			Type: "replay:error",
			Data: map[string]interface{}{
				"error": fmt.Sprintf("Failed to discover log files: %v", err),
			},
		})
		return fmt.Errorf("failed to discover log files: %w", err)
	}

	if len(logFiles) == 0 {
		s.wsServer.Broadcast(Event{
			Type: "replay:completed",
			Data: map[string]interface{}{
				"message": "No log files found to replay",
			},
		})
		return nil
	}

	log.Printf("Found %d log file(s) to replay", len(logFiles))

	// Step 4: Collect all log entries from all files first
	// This ensures quest completion detection and other stateful parsing works correctly
	startTime := time.Now()
	var allEntries []*logreader.LogEntry

	for i, logFile := range logFiles {
		log.Printf("Reading file %d/%d: %s", i+1, len(logFiles), logFile.Name)

		// Broadcast progress
		s.wsServer.Broadcast(Event{
			Type: "replay:progress",
			Data: map[string]interface{}{
				"totalFiles":     len(logFiles),
				"processedFiles": i,
				"currentFile":    logFile.Name,
				"totalEntries":   len(allEntries),
			},
		})

		// Read all entries from this file
		entries, err := s.readLogFile(logFile.Path)
		if err != nil {
			log.Printf("Warning: Error reading file %s: %v", logFile.Name, err)
			// Continue with next file, don't fail entire replay
			continue
		}

		log.Printf("Read %d entries from %s", len(entries), logFile.Name)
		allEntries = append(allEntries, entries...)
	}

	log.Printf("Collected %d total entries from %d files", len(allEntries), len(logFiles))

	// Step 5: Enable bulk import mode for faster processing
	log.Println("Enabling bulk import mode for faster processing...")
	bulkSettings, err := s.storage.EnableBulkImportMode(s.ctx)
	if err != nil {
		log.Printf("Warning: Failed to enable bulk import mode: %v", err)
		// Continue anyway with normal mode
	}

	// Ensure we restore safe mode even if processing fails
	defer func() {
		if bulkSettings != nil {
			if err := s.storage.RestoreSafeMode(s.ctx, bulkSettings); err != nil {
				log.Printf("Error restoring safe mode: %v", err)
			}
		}
	}()

	// Step 6: Process entries in chunks to show incremental progress
	// This is critical for quest completion detection, which relies on seeing
	// quests disappear from subsequent QuestGetQuests responses
	log.Println("Processing all entries through business logic...")

	chunkSize := 5000 // Process 5000 entries at a time
	totalChunks := (len(allEntries) + chunkSize - 1) / chunkSize
	var totalResult *logprocessor.ProcessResult

	for chunkIdx := 0; chunkIdx < totalChunks; chunkIdx++ {
		start := chunkIdx * chunkSize
		end := start + chunkSize
		if end > len(allEntries) {
			end = len(allEntries)
		}

		chunk := allEntries[start:end]

		// Broadcast progress
		percentComplete := float64(end) / float64(len(allEntries)) * 100
		s.wsServer.Broadcast(Event{
			Type: "replay:progress",
			Data: map[string]interface{}{
				"totalFiles":       len(logFiles),
				"processedFiles":   len(logFiles),
				"currentFile":      fmt.Sprintf("Processing entries %d-%d of %d", start, end, len(allEntries)),
				"totalEntries":     len(allEntries),
				"processedEntries": end,
				"percentComplete":  percentComplete,
			},
		})

		log.Printf("Processing chunk %d/%d (%d entries)...", chunkIdx+1, totalChunks, len(chunk))

		result, err := s.logProcessor.ProcessLogEntries(s.ctx, chunk)
		if err != nil {
			s.wsServer.Broadcast(Event{
				Type: "replay:error",
				Data: map[string]interface{}{
					"error": fmt.Sprintf("Failed to process entries: %v", err),
				},
			})
			return fmt.Errorf("failed to process entries: %w", err)
		}

		// Accumulate results
		if totalResult == nil {
			totalResult = result
		} else {
			totalResult.MatchesStored += result.MatchesStored
			totalResult.GamesStored += result.GamesStored
			totalResult.DecksStored += result.DecksStored
			totalResult.RanksStored += result.RanksStored
			totalResult.QuestsStored += result.QuestsStored
			totalResult.QuestsCompleted += result.QuestsCompleted
			totalResult.AchievementsStored += result.AchievementsStored
			totalResult.DraftsStored += result.DraftsStored
			totalResult.DraftPicksStored += result.DraftPicksStored
			totalResult.Errors = append(totalResult.Errors, result.Errors...)
		}
	}

	result := totalResult
	elapsed := time.Since(startTime)
	log.Printf("Replay completed in %v: %d entries, %d matches, %d decks, %d quests, %d drafts",
		elapsed, len(allEntries), result.MatchesStored, result.DecksStored, result.QuestsStored, result.DraftsStored)

	// Broadcast completion
	s.wsServer.Broadcast(Event{
		Type: "replay:completed",
		Data: map[string]interface{}{
			"totalFiles":      len(logFiles),
			"totalEntries":    len(allEntries),
			"matchesImported": result.MatchesStored,
			"decksImported":   result.DecksStored,
			"questsImported":  result.QuestsStored,
			"draftsImported":  result.DraftsStored,
			"duration":        elapsed.Seconds(),
		},
	})

	// Step 6: Restart poller
	log.Println("Restarting log poller...")
	if err := s.restartPoller(); err != nil {
		return fmt.Errorf("failed to restart poller: %w", err)
	}

	return nil
}

// LogFileInfo contains information about a discovered log file.
type LogFileInfo struct {
	Path    string
	Name    string
	ModTime time.Time
}

// discoverLogFiles finds all MTGA log files and sorts them chronologically.
func (s *Service) discoverLogFiles() ([]LogFileInfo, error) {
	// Get log directories
	logDirs := getLogDirectories()

	var allFiles []LogFileInfo

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
			// Match: UTC_Log*.log, Player.log, Player-prev.log
			if !isLogFile(name) {
				continue
			}

			path := filepath.Join(logDir, name)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			allFiles = append(allFiles, LogFileInfo{
				Path:    path,
				Name:    name,
				ModTime: info.ModTime(),
			})
		}
	}

	// Sort by modification time (oldest first for chronological replay)
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].ModTime.Before(allFiles[j].ModTime)
	})

	return allFiles, nil
}

// isLogFile returns true if the filename is a recognized MTGA log file.
func isLogFile(name string) bool {
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
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var dirs []string
	// macOS
	macDir := filepath.Join(home, "Library", "Application Support", "com.wizards.mtga", "Logs", "Logs")
	if fileExists(macDir) {
		dirs = append(dirs, macDir)
	}
	macDir2 := filepath.Join(home, "Library", "Logs", "Wizards of the Coast", "MTGA")
	if fileExists(macDir2) {
		dirs = append(dirs, macDir2)
	}

	// Windows
	winDir := filepath.Join(home, "AppData", "LocalLow", "Wizards Of The Coast", "MTGA")
	if fileExists(winDir) {
		dirs = append(dirs, winDir)
	}

	return dirs
}

// fileExists checks if a path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readLogFile reads all log entries from a file without processing them.
// This is used during replay to collect entries from all files before processing.
func (s *Service) readLogFile(path string) ([]*logreader.LogEntry, error) {
	// Create a poller that reads from start
	config := &logreader.PollerConfig{
		Path:          path,
		Interval:      100 * time.Millisecond,
		BufferSize:    10000, // Larger buffer for reading
		UseFileEvents: false,
		ReadFromStart: true, // Read entire file
	}

	poller, err := logreader.NewPoller(config)
	if err != nil {
		return nil, fmt.Errorf("create poller: %w", err)
	}

	// Start polling - returns the update channel
	updates := poller.Start()
	defer poller.Stop()

	// Collect all entries
	var allEntries []*logreader.LogEntry
	timeout := time.After(30 * time.Second) // Timeout if no updates for 30s

	for {
		select {
		case entry, ok := <-updates:
			if !ok {
				// Channel closed, return all collected entries
				return allEntries, nil
			}

			allEntries = append(allEntries, entry)
			timeout = time.After(30 * time.Second) // Reset timeout on each entry

		case <-timeout:
			// No updates for 30s, assume file is done
			return allEntries, nil
		}
	}
}

// restartPoller restarts the log poller after replay.
func (s *Service) restartPoller() error {
	// Determine log path
	logPath := s.config.LogPath
	if logPath == "" {
		detected, err := logreader.DefaultLogPath()
		if err != nil {
			return fmt.Errorf("failed to detect log path: %w", err)
		}
		logPath = detected
	}

	// Create and start log poller (only monitor NEW entries, not from start)
	pollerConfig := logreader.DefaultPollerConfig(logPath)
	pollerConfig.Interval = s.config.PollInterval
	pollerConfig.UseFileEvents = s.config.UseFSNotify
	pollerConfig.EnableMetrics = s.config.EnableMetrics
	pollerConfig.ReadFromStart = false // Only monitor new entries after replay

	poller, err := logreader.NewPoller(pollerConfig)
	if err != nil {
		return fmt.Errorf("failed to create log poller: %w", err)
	}

	s.poller = poller

	// Start log poller
	updates := s.poller.Start()
	errChan := s.poller.Errors()

	// Process log updates
	go s.processUpdates(updates, errChan)

	log.Println("Log poller restarted successfully")
	return nil
}

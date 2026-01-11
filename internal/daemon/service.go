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

// EventForwarder is an interface for forwarding daemon events to external systems.
// This allows the API server to receive daemon events without connecting via WebSocket.
// The interface uses interface{} to avoid import cycles with packages that implement this.
type EventForwarder interface {
	ForwardEvent(event interface{})
}

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

	// Replay engine for testing
	replayEngine *ReplayEngine

	// Flight recorder for debugging (Go 1.25+)
	flightRecorder *FlightRecorder

	// Event forwarders for external systems (e.g., API server WebSocket)
	forwarders   []EventForwarder
	forwardersMu sync.RWMutex

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

	s := &Service{
		config:       config,
		storage:      storage,
		logProcessor: logprocessor.NewService(storage),
		wsServer:     NewWebSocketServerWithCORS(config.Port, config.CORSConfig),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize replay engine
	s.replayEngine = NewReplayEngine(s)

	// Initialize flight recorder for debugging (Go 1.25+)
	s.flightRecorder = NewFlightRecorder(DefaultFlightRecorderConfig())

	return s
}

// Start starts the daemon service.
func (s *Service) Start() error {
	s.startTime = time.Now()
	log.Println("Starting MTGA Companion daemon...")

	// Start flight recorder for debugging
	if err := s.flightRecorder.Start(); err != nil {
		log.Printf("Warning: Failed to start flight recorder: %v", err)
	}

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

	// Run startup recovery to process any missed log files
	// This captures data from UTC_Log files and Player-prev.log created while daemon was stopped
	if err := s.StartupRecovery(); err != nil {
		// Log error but don't fail startup - we can still monitor new entries
		log.Printf("Warning: Startup recovery failed: %v", err)
	}

	// Start UTC_Log monitoring in background
	// This detects and processes new UTC_Log files created during daemon runtime (Phase 2)
	go s.monitorUTCLogs()

	// Start Player.log archival monitoring in background (Phase 4)
	// This periodically archives the active Player.log file
	go s.monitorPlayerLogArchival()

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
	s.broadcastEvent(Event{
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
// The context can be used to enforce a shutdown deadline.
func (s *Service) Stop(ctx context.Context) error {
	log.Println("Stopping daemon...")

	// Create a channel to signal completion
	done := make(chan struct{})

	go func() {
		// Archive Player.log before shutdown (Phase 4)
		if err := s.archiveOnShutdown(); err != nil {
			log.Printf("Warning: Failed to archive on shutdown: %v", err)
		}

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

		// Stop flight recorder
		if s.flightRecorder != nil {
			s.flightRecorder.Stop()
		}

		close(done)
	}()

	// Wait for shutdown to complete or context to expire
	select {
	case <-done:
		log.Println("Daemon stopped")
		return nil
	case <-ctx.Done():
		log.Println("Daemon shutdown timed out, forcing stop")
		s.cancel() // Force cancel if not already done
		return ctx.Err()
	}
}

// RegisterEventForwarder registers an event forwarder to receive daemon events.
// This is used by the API server to forward events to its WebSocket clients.
func (s *Service) RegisterEventForwarder(forwarder EventForwarder) {
	s.forwardersMu.Lock()
	defer s.forwardersMu.Unlock()
	s.forwarders = append(s.forwarders, forwarder)
	log.Printf("Registered event forwarder (total: %d)", len(s.forwarders))
}

// broadcastEvent broadcasts an event to both the daemon's WebSocket clients
// and any registered event forwarders (e.g., the API server).
func (s *Service) broadcastEvent(event Event) {
	// Broadcast to daemon's WebSocket clients
	s.wsServer.Broadcast(event) // Keep direct WebSocket call here

	// Forward to registered forwarders
	s.forwardersMu.RLock()
	forwarders := s.forwarders
	s.forwardersMu.RUnlock()

	for _, forwarder := range forwarders {
		forwarder.ForwardEvent(event)
	}
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

			// Track error and capture trace if significant
			s.trackError("poller-error")

			// Broadcast error event
			s.broadcastEvent(Event{
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

		// Track error and capture trace if significant
		s.trackError("log-processing-error")
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
		s.broadcastEvent(Event{
			Type: "stats:updated",
			Data: map[string]interface{}{
				"matches": result.MatchesStored,
				"games":   result.GamesStored,
			},
		})
	}

	if result.DecksStored > 0 {
		log.Printf("Stored %d deck(s)", result.DecksStored)
		s.broadcastEvent(Event{
			Type: "deck:updated",
			Data: map[string]interface{}{
				"count": result.DecksStored,
			},
		})
	}

	if result.RanksStored > 0 {
		log.Printf("Stored %d rank update(s)", result.RanksStored)
		s.broadcastEvent(Event{
			Type: "rank:updated",
			Data: map[string]interface{}{
				"count": result.RanksStored,
			},
		})
	}

	if result.QuestsStored > 0 {
		log.Printf("Stored %d quest(s)", result.QuestsStored)
		s.broadcastEvent(Event{
			Type: "quest:updated",
			Data: map[string]interface{}{
				"count":     result.QuestsStored,
				"completed": result.QuestsCompleted,
			},
		})
	}

	if result.QuestsCompleted > 0 {
		log.Printf("Completed %d quest(s)", result.QuestsCompleted)
		s.broadcastEvent(Event{
			Type: "quest:updated",
			Data: map[string]interface{}{
				"completed": result.QuestsCompleted,
			},
		})
	}

	if result.DraftsStored > 0 {
		log.Printf("Stored %d draft session(s) with %d picks", result.DraftsStored, result.DraftPicksStored)
		s.broadcastEvent(Event{
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
			s.broadcastEvent(Event{
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

// trackError tracks an error and optionally captures a trace if conditions are met.
// It captures a trace when errors become significant (e.g., high error rate).
func (s *Service) trackError(reason string) {
	s.healthMu.Lock()
	s.totalErrors++
	errorCount := s.totalErrors
	processedCount := s.totalProcessed
	s.healthMu.Unlock()

	// Capture trace on significant errors:
	// - First 3 errors (for debugging startup issues)
	// - Every 10th error thereafter
	// - When error rate exceeds 10%
	shouldCapture := errorCount <= 3 ||
		errorCount%10 == 0 ||
		(processedCount > 0 && float64(errorCount)/float64(processedCount) > 0.1)

	if shouldCapture && s.flightRecorder != nil && s.flightRecorder.Enabled() {
		if path, err := s.flightRecorder.CaptureTrace(reason); err != nil {
			log.Printf("Failed to capture trace: %v", err)
		} else {
			log.Printf("Captured debug trace: %s", path)
		}
	}
}

// CaptureTrace manually captures a flight recorder trace.
// Useful for debugging or when investigating issues.
func (s *Service) CaptureTrace(reason string) (string, error) {
	if s.flightRecorder == nil {
		return "", fmt.Errorf("flight recorder not initialized")
	}
	return s.flightRecorder.CaptureTrace(reason)
}

// GetFlightRecorderEnabled returns whether flight recording is active.
func (s *Service) GetFlightRecorderEnabled() bool {
	if s.flightRecorder == nil {
		return false
	}
	return s.flightRecorder.Enabled()
}

// ReplayHistoricalLogs replays all historical log files through the processing pipeline.
// This is the CORRECT way to import historical data - it runs logs through the same
// business logic as real-time processing, ensuring GraphState updates, quest completion
// detection, rank progression, etc. all work correctly.
func (s *Service) ReplayHistoricalLogs(clearData bool) error {
	log.Println("Starting historical log replay...")

	// Broadcast replay start event
	s.broadcastEvent(Event{
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
			s.broadcastEvent(Event{
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
		s.broadcastEvent(Event{
			Type: "replay:error",
			Data: map[string]interface{}{
				"error": fmt.Sprintf("Failed to discover log files: %v", err),
			},
		})
		return fmt.Errorf("failed to discover log files: %w", err)
	}

	if len(logFiles) == 0 {
		s.broadcastEvent(Event{
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

		// Broadcast progress (field names match gui.LogReplayProgress)
		s.broadcastEvent(Event{
			Type: "replay:progress",
			Data: map[string]interface{}{
				"totalFiles":       len(logFiles),
				"processedFiles":   i,
				"currentFile":      logFile.Name,
				"totalEntries":     len(allEntries),
				"processedEntries": 0,
				"percentComplete":  0.0,
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

		// Broadcast progress (field names match gui.LogReplayProgress)
		percentComplete := float64(end) / float64(len(allEntries)) * 100
		s.broadcastEvent(Event{
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
			s.broadcastEvent(Event{
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
			totalResult.DraftsStored += result.DraftsStored
			totalResult.DraftPicksStored += result.DraftPicksStored
			totalResult.Errors = append(totalResult.Errors, result.Errors...)
		}
	}

	result := totalResult
	elapsed := time.Since(startTime)
	log.Printf("Replay completed in %v: %d entries, %d matches, %d decks, %d quests, %d drafts",
		elapsed, len(allEntries), result.MatchesStored, result.DecksStored, result.QuestsStored, result.DraftsStored)

	// Broadcast completion (field names match gui.LogReplayProgress)
	s.broadcastEvent(Event{
		Type: "replay:completed",
		Data: map[string]interface{}{
			"totalFiles":       len(logFiles),
			"processedFiles":   len(logFiles),
			"totalEntries":     len(allEntries),
			"processedEntries": len(allEntries),
			"percentComplete":  100.0,
			"matchesImported":  result.MatchesStored,
			"decksImported":    result.DecksStored,
			"questsImported":   result.QuestsStored,
			"draftsImported":   result.DraftsStored,
			"duration":         elapsed.Seconds(),
		},
	})

	// Step 6: Restart poller
	log.Println("Restarting log poller...")
	if err := s.restartPoller(); err != nil {
		return fmt.Errorf("failed to restart poller: %w", err)
	}

	return nil
}

// StartupRecovery processes any unprocessed log files on daemon startup.
// This captures data from UTC_Log files and Player-prev.log that may have been
// created while the daemon was stopped, preventing data loss from daemon downtime.
func (s *Service) StartupRecovery() error {
	log.Println("Starting startup recovery to process any missed log files...")

	// Discover all available log files
	logFiles, err := s.discoverLogFiles()
	if err != nil {
		return fmt.Errorf("failed to discover log files: %w", err)
	}

	if len(logFiles) == 0 {
		log.Println("No log files found for startup recovery")
		return nil
	}

	log.Printf("Found %d log file(s), checking for unprocessed files...", len(logFiles))

	processedCount := 0
	skippedCount := 0
	var totalEntries, totalMatches int

	for _, logFile := range logFiles {
		// Skip Player.log as it's being actively monitored by the poller
		if logFile.Name == "Player.log" {
			continue
		}

		// Check if we've already processed this file
		alreadyProcessed, err := s.storage.HasProcessedLogFile(s.ctx, logFile.Name)
		if err != nil {
			log.Printf("Warning: Failed to check if %s was processed: %v", logFile.Name, err)
			continue
		}

		if alreadyProcessed {
			skippedCount++
			continue
		}

		// Process this unprocessed file
		log.Printf("Processing unprocessed file: %s", logFile.Name)

		// Read all entries from the file
		entries, err := s.readLogFile(logFile.Path)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", logFile.Name, err)
			continue
		}

		if len(entries) == 0 {
			log.Printf("Skipping %s (no entries found)", logFile.Name)
			continue
		}

		// Process entries through business logic
		result, err := s.logProcessor.ProcessLogEntries(s.ctx, entries)
		if err != nil {
			log.Printf("Warning: Failed to process entries from %s: %v", logFile.Name, err)
			continue
		}

		// Get file size for tracking
		fileInfo, err := os.Stat(logFile.Path)
		var fileSizeBytes int64
		if err == nil {
			fileSizeBytes = fileInfo.Size()
		}

		// Mark file as processed
		err = s.storage.MarkLogFileProcessed(s.ctx, logFile.Name, len(entries), result.MatchesStored, fileSizeBytes)
		if err != nil {
			log.Printf("Warning: Failed to mark %s as processed: %v", logFile.Name, err)
			// Continue anyway - we've already processed the data
		}

		log.Printf("✓ Processed %s: %d entries, %d matches, %d decks, %d quests",
			logFile.Name, len(entries), result.MatchesStored, result.DecksStored, result.QuestsStored)

		processedCount++
		totalEntries += len(entries)
		totalMatches += result.MatchesStored
	}

	if processedCount > 0 {
		log.Printf("Startup recovery complete: processed %d file(s) (%d entries, %d matches), skipped %d already-processed file(s)",
			processedCount, totalEntries, totalMatches, skippedCount)
	} else {
		log.Printf("Startup recovery complete: all %d file(s) already processed, no new data to recover", skippedCount)
	}

	return nil
}

// monitorUTCLogs runs in the background and periodically checks for new UTC_Log files.
// When MTGA rotates logs (creates new session files), this detector captures and processes them.
// This prevents data loss during long daemon uptime when MTGA creates new UTC_Log files.
func (s *Service) monitorUTCLogs() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("UTC_Log monitoring started (checking every 5 minutes)")

	for {
		select {
		case <-ticker.C:
			if err := s.checkForNewLogFiles(); err != nil {
				log.Printf("Warning: UTC_Log check failed: %v", err)
			}
		case <-s.ctx.Done():
			log.Println("UTC_Log monitoring stopped")
			return
		}
	}
}

// checkForNewLogFiles discovers and processes any unprocessed log files.
// This is similar to StartupRecovery but runs periodically during daemon uptime.
// It only processes UTC_Log files (not Player.log or Player-prev.log).
func (s *Service) checkForNewLogFiles() error {
	logFiles, err := s.discoverLogFiles()
	if err != nil {
		return fmt.Errorf("failed to discover log files: %w", err)
	}

	processedCount := 0

	for _, logFile := range logFiles {
		// Skip Player.log (monitored by poller) and Player-prev.log (handled by startup recovery)
		if logFile.Name == "Player.log" || logFile.Name == "Player-prev.log" {
			continue
		}

		// Only process UTC_Log files
		if !strings.HasPrefix(logFile.Name, "UTC_Log") {
			continue
		}

		// Check if already processed
		alreadyProcessed, err := s.storage.HasProcessedLogFile(s.ctx, logFile.Name)
		if err != nil {
			log.Printf("Warning: Failed to check if %s was processed: %v", logFile.Name, err)
			continue
		}

		if alreadyProcessed {
			continue
		}

		// New UTC_Log file detected!
		log.Printf("New UTC_Log file detected: %s", logFile.Name)

		// Process it (same logic as StartupRecovery)
		entries, err := s.readLogFile(logFile.Path)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", logFile.Name, err)
			continue
		}

		if len(entries) == 0 {
			log.Printf("Skipping %s (no entries found)", logFile.Name)
			continue
		}

		result, err := s.logProcessor.ProcessLogEntries(s.ctx, entries)
		if err != nil {
			log.Printf("Warning: Failed to process %s: %v", logFile.Name, err)
			continue
		}

		fileInfo, err := os.Stat(logFile.Path)
		var fileSizeBytes int64
		if err == nil {
			fileSizeBytes = fileInfo.Size()
		}

		err = s.storage.MarkLogFileProcessed(s.ctx, logFile.Name, len(entries), result.MatchesStored, fileSizeBytes)
		if err != nil {
			log.Printf("Warning: Failed to mark %s as processed: %v", logFile.Name, err)
			// Continue anyway - we've already processed the data
		}

		log.Printf("✓ Processed new UTC_Log: %s (%d entries, %d matches, %d decks, %d quests)",
			logFile.Name, len(entries), result.MatchesStored, result.DecksStored, result.QuestsStored)

		// Archive the UTC_Log file if archival is enabled
		if s.config.EnableArchival {
			if _, err := s.archiveLogFile(logFile.Path); err != nil {
				log.Printf("Warning: Failed to archive %s: %v", logFile.Name, err)
			}
		}

		processedCount++
	}

	if processedCount > 0 {
		log.Printf("UTC_Log check: processed %d new file(s)", processedCount)
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
	// Use Reader to read entire file synchronously (much faster than Poller)
	reader, err := logreader.NewReader(path)
	if err != nil {
		return nil, fmt.Errorf("create reader: %w", err)
	}
	defer func() {
		_ = reader.Close() // Explicitly ignore error - file is read-only
	}()

	// Read all entries from the file in one pass
	entries, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read entries: %w", err)
	}

	return entries, nil
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

// getArchiveDir returns the archive directory path, creating it if needed.
func (s *Service) getArchiveDir() (string, error) {
	archiveDir := s.config.ArchiveDir

	// Use default directory if not specified
	if archiveDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		archiveDir = filepath.Join(home, ".mtga-companion", "archived_logs")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create archive directory: %w", err)
	}

	return archiveDir, nil
}

// archiveLogFile copies a log file to the archive directory.
// Returns the destination path and any error encountered.
func (s *Service) archiveLogFile(srcPath string) (string, error) {
	// Get archive directory
	archiveDir, err := s.getArchiveDir()
	if err != nil {
		return "", err
	}

	// Extract filename from source path
	filename := filepath.Base(srcPath)
	destPath := filepath.Join(archiveDir, filename)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		// File already archived, skip
		return destPath, nil
	}

	// Read source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write destination file: %w", err)
	}

	log.Printf("✓ Archived log file: %s (%d bytes) → %s", filename, len(data), archiveDir)
	return destPath, nil
}

// monitorPlayerLogArchival periodically archives the active Player.log file.
// This runs in a background goroutine and archives every config.ArchiveInterval.
func (s *Service) monitorPlayerLogArchival() {
	if !s.config.EnableArchival {
		return
	}

	ticker := time.NewTicker(s.config.ArchiveInterval)
	defer ticker.Stop()

	log.Printf("Player.log archival started (archiving every %v)", s.config.ArchiveInterval)

	for {
		select {
		case <-ticker.C:
			// Get current log path
			logPath := s.config.LogPath
			if logPath == "" {
				detected, err := logreader.DefaultLogPath()
				if err != nil {
					log.Printf("Warning: Failed to detect log path for archival: %v", err)
					continue
				}
				logPath = detected
			}

			// Archive the current Player.log
			if _, err := s.archiveLogFile(logPath); err != nil {
				log.Printf("Warning: Failed to archive Player.log: %v", err)
			}
		case <-s.ctx.Done():
			log.Println("Player.log archival stopped")
			return
		}
	}
}

// archiveOnShutdown archives the current Player.log before daemon shutdown.
// This ensures we capture the final state of the log file.
func (s *Service) archiveOnShutdown() error {
	if !s.config.EnableArchival {
		return nil
	}

	// Get current log path
	logPath := s.config.LogPath
	if logPath == "" {
		detected, err := logreader.DefaultLogPath()
		if err != nil {
			return fmt.Errorf("failed to detect log path: %w", err)
		}
		logPath = detected
	}

	// Archive the current Player.log
	if _, err := s.archiveLogFile(logPath); err != nil {
		return fmt.Errorf("failed to archive Player.log on shutdown: %w", err)
	}

	log.Println("✓ Archived Player.log before shutdown")
	return nil
}

// StartReplay starts replay of one or more log files with the specified speed and filter.
func (s *Service) StartReplay(logPaths []string, speed float64, filterType string, pauseOnDraft bool) error {
	return s.replayEngine.Start(logPaths, speed, filterType, pauseOnDraft)
}

// PauseReplay pauses the active replay.
func (s *Service) PauseReplay() error {
	return s.replayEngine.Pause()
}

// ResumeReplay resumes a paused replay.
func (s *Service) ResumeReplay() error {
	return s.replayEngine.Resume()
}

// StopReplay stops the active replay.
func (s *Service) StopReplay() error {
	return s.replayEngine.Stop()
}

// GetReplayStatus returns the current replay status.
func (s *Service) GetReplayStatus() map[string]interface{} {
	return s.replayEngine.GetStatus()
}

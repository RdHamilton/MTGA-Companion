package gui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ramonehamilton/MTGA-Companion/internal/events"
	"github.com/ramonehamilton/MTGA-Companion/internal/ipc"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/datasets"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/setcache"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckexport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckimport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/recommendations"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// SystemFacade handles system initialization, daemon communication, and replay operations
type SystemFacade struct {
	services        *Services
	pollerMu        sync.Mutex
	ipcClientMu     sync.Mutex
	pollerStop      context.CancelFunc
	eventDispatcher *events.EventDispatcher
}

// NewSystemFacade creates a new SystemFacade
func NewSystemFacade(services *Services) *SystemFacade {
	dispatcher := events.NewEventDispatcher()

	// Register Wails observer for frontend events
	dispatcher.Register(events.NewWailsObserver())

	return &SystemFacade{
		services:        services,
		eventDispatcher: dispatcher,
	}
}

// GetEventDispatcher returns the event dispatcher instance.
// Allows other facades to access the dispatcher for emitting events.
func (s *SystemFacade) GetEventDispatcher() *events.EventDispatcher {
	return s.eventDispatcher
}

// Initialize initializes the application with database path
func (s *SystemFacade) Initialize(ctx context.Context, dbPath string) error {
	// Use default path if empty
	if dbPath == "" {
		dbPath = getDefaultDBPath()
	}

	config := storage.DefaultConfig(dbPath)
	config.BusyTimeout = 10 * time.Second // Increase timeout to handle concurrent poller operations

	db, err := storage.Open(config)
	if err != nil {
		return err
	}
	s.services.Storage = storage.NewService(db)

	// Initialize card services
	scryfallClient := scryfall.NewClient()

	// Initialize dataset service for 17Lands ratings
	datasetService, err := datasets.NewService(datasets.DefaultServiceOptions())
	if err != nil {
		return fmt.Errorf("failed to initialize dataset service: %w", err)
	}
	s.services.DatasetService = datasetService

	// Initialize SetFetcher for card metadata
	s.services.SetFetcher = setcache.NewFetcher(
		scryfallClient,
		s.services.Storage.SetCardRepo(),
		s.services.Storage.DraftRatingsRepo(),
	)

	// Initialize RatingsFetcher for draft ratings
	s.services.RatingsFetcher = setcache.NewRatingsFetcherWithDatasets(
		datasetService,
		s.services.Storage.DraftRatingsRepo(),
	)

	// Initialize CardService for card metadata with caching
	cardService, err := cards.NewService(s.services.Storage.GetDB(), cards.DefaultServiceConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize card service: %w", err)
	}
	s.services.CardService = cardService

	// Initialize DeckImportParser (depends on CardService)
	s.services.DeckImportParser = deckimport.NewParser(cardService)

	// Initialize DeckExporter (uses CardService as CardProvider)
	s.services.DeckExporter = deckexport.NewExporter(cardService)

	// Initialize RecommendationEngine (depends on CardService)
	s.services.RecommendationEngine = recommendations.NewRuleBasedEngine(cardService)

	log.Println("Card services initialized successfully")
	return nil
}

// StartPoller starts the log file poller for real-time updates
func (s *SystemFacade) StartPoller(ctx context.Context) error {
	s.pollerMu.Lock()
	defer s.pollerMu.Unlock()

	if s.services.Storage == nil {
		return &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}

	// Stop existing poller if running
	if s.services.Poller != nil {
		return nil // Already running
	}

	// Get MTGA log path
	logPath, err := getMTGALogPath()
	if err != nil {
		log.Printf("Failed to find MTGA log file: %v", err)
		return err
	}

	log.Printf("Starting log file poller for: %s", logPath)

	// Create poller config
	config := logreader.DefaultPollerConfig(logPath)
	config.Interval = 5 * time.Second // Poll every 5 seconds

	// Create poller
	poller, err := logreader.NewPoller(config)
	if err != nil {
		log.Printf("Failed to create poller: %v", err)
		return err
	}

	s.services.Poller = poller

	// Start poller
	updates := poller.Start()
	errChan := poller.Errors()

	// Create cancellable context
	pollerCtx, cancel := context.WithCancel(ctx)
	s.pollerStop = cancel

	// Start background goroutine to process updates
	go s.processPollerUpdates(pollerCtx, updates, errChan)

	log.Println("Log file poller started successfully")
	return nil
}

// StopPoller stops the log file poller
func (s *SystemFacade) StopPoller() {
	s.pollerMu.Lock()
	defer s.pollerMu.Unlock()

	if s.pollerStop != nil {
		s.pollerStop()
		s.pollerStop = nil
	}

	if s.services.Poller != nil {
		s.services.Poller.Stop()
		s.services.Poller = nil
		log.Println("Log file poller stopped")
	}
}

// processPollerUpdates processes new log entries in the background
func (s *SystemFacade) processPollerUpdates(ctx context.Context, updates <-chan *logreader.LogEntry, errChan <-chan error) {
	var entryBuffer []*logreader.LogEntry
	ticker := time.NewTicker(5 * time.Second) // Batch process every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-updates:
			if !ok {
				return
			}
			// Buffer entries for batch processing
			entryBuffer = append(entryBuffer, entry)
		case err, ok := <-errChan:
			if !ok {
				return
			}
			log.Printf("Poller error: %v", err)
		case <-ticker.C:
			// Process buffered entries
			if len(entryBuffer) > 0 {
				// Note: processNewEntries would need to be implemented
				// This is a placeholder for the actual processing logic
				entryBuffer = nil // Clear buffer
			}
		}
	}
}

// getMTGALogPath returns the path to the MTGA Player.log file based on platform
func getMTGALogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var logPath string
	switch runtime.GOOS {
	case "darwin":
		// macOS
		logPath = filepath.Join(home, "Library", "Application Support", "com.wizards.mtga", "Logs", "Logs")
	case "windows":
		// Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		logPath = filepath.Join(appData, "Wizards of the Coast", "MTGA", "Logs")
	default:
		return "", &AppError{Message: "Unsupported platform for MTGA log detection"}
	}

	// Find the most recent Player.log file
	files, err := os.ReadDir(logPath)
	if err != nil {
		return "", err
	}

	var newestLog string
	var newestTime time.Time
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// Look for Player.log or UTC_Log files
		name := file.Name()
		if name == "Player.log" || filepath.Ext(name) == ".log" {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if newestLog == "" || info.ModTime().After(newestTime) {
				newestLog = filepath.Join(logPath, name)
				newestTime = info.ModTime()
			}
		}
	}

	if newestLog == "" {
		return "", &AppError{Message: "No MTGA log file found"}
	}

	return newestLog, nil
}

// getDefaultDBPath returns the default database path
func getDefaultDBPath() string {
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Error getting home directory: %v", err)
			return "mtga.db" // Fallback to current directory
		}
		dbPath = filepath.Join(home, ".mtga-companion", "mtga.db")
	}
	return dbPath
}

// connectToDaemon connects to the daemon service
func (s *SystemFacade) connectToDaemon(ctx context.Context) error {
	s.ipcClientMu.Lock()
	defer s.ipcClientMu.Unlock()

	// Create IPC client
	wsURL := fmt.Sprintf("ws://localhost:%d", s.services.DaemonPort)
	s.services.IPCClient = ipc.NewClient(wsURL)

	// Try to connect
	if err := s.services.IPCClient.Connect(); err != nil {
		s.services.IPCClient = nil
		return err
	}

	// Setup event handlers
	s.setupEventHandlers(ctx)

	// Start listening for events
	s.services.IPCClient.Start()

	return nil
}

// setupEventHandlers registers event handlers for daemon events.
// Uses the EventDispatcher to forward events to all registered observers.
func (s *SystemFacade) setupEventHandlers(ctx context.Context) {
	// Handle stats:updated events from daemon
	s.services.IPCClient.On("stats:updated", func(data map[string]interface{}) {
		log.Printf("Received stats:updated event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "stats:updated",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle rank:updated events from daemon
	s.services.IPCClient.On("rank:updated", func(data map[string]interface{}) {
		log.Printf("Received rank:updated event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "rank:updated",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle deck:updated events from daemon
	s.services.IPCClient.On("deck:updated", func(data map[string]interface{}) {
		log.Printf("Received deck:updated event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "deck:updated",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle quest:updated events from daemon
	s.services.IPCClient.On("quest:updated", func(data map[string]interface{}) {
		log.Printf("Received quest:updated event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "quest:updated",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle achievement:updated events from daemon
	s.services.IPCClient.On("achievement:updated", func(data map[string]interface{}) {
		log.Printf("Received achievement:updated event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "achievement:updated",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle daemon:status events
	s.services.IPCClient.On("daemon:status", func(data map[string]interface{}) {
		log.Printf("Daemon status: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "daemon:status",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle daemon:connected events
	s.services.IPCClient.On("daemon:connected", func(data map[string]interface{}) {
		log.Printf("Daemon connected: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "daemon:connected",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle daemon:error events
	s.services.IPCClient.On("daemon:error", func(data map[string]interface{}) {
		log.Printf("Daemon error: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "daemon:error",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle replay:started events from daemon
	s.services.IPCClient.On("replay:started", func(data map[string]interface{}) {
		log.Printf("Received replay:started event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "replay:started",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle replay:progress events from daemon
	s.services.IPCClient.On("replay:progress", func(data map[string]interface{}) {
		log.Printf("Received replay:progress event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "replay:progress",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle replay:paused events from daemon
	s.services.IPCClient.On("replay:paused", func(data map[string]interface{}) {
		log.Printf("Received replay:paused event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "replay:paused",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle replay:resumed events from daemon
	s.services.IPCClient.On("replay:resumed", func(data map[string]interface{}) {
		log.Printf("Received replay:resumed event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "replay:resumed",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle replay:completed events from daemon
	s.services.IPCClient.On("replay:completed", func(data map[string]interface{}) {
		log.Printf("Received replay:completed event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "replay:completed",
			Data:    data,
			Context: ctx,
		})
	})

	// Handle replay:error events from daemon
	s.services.IPCClient.On("replay:error", func(data map[string]interface{}) {
		log.Printf("Received replay:error event from daemon: %v", data)
		s.eventDispatcher.Dispatch(events.Event{
			Type:    "replay:error",
			Data:    data,
			Context: ctx,
		})
	})
}

// stopDaemonClient stops the daemon client connection
func (s *SystemFacade) stopDaemonClient(ctx context.Context) {
	s.ipcClientMu.Lock()
	defer s.ipcClientMu.Unlock()

	if s.services.IPCClient != nil {
		s.services.IPCClient.Stop()
		s.services.IPCClient = nil
		s.services.DaemonMode = false
		log.Println("Daemon client stopped")

		// Dispatch status change event
		if ctx != nil {
			s.eventDispatcher.Dispatch(events.Event{
				Type: "daemon:status",
				Data: map[string]interface{}{
					"status":    "standalone",
					"connected": false,
				},
				Context: ctx,
			})
		}
	}
}

// GetConnectionStatus returns current connection status for the frontend.
func (s *SystemFacade) GetConnectionStatus() map[string]interface{} {
	s.ipcClientMu.Lock()
	defer s.ipcClientMu.Unlock()

	status := "standalone"
	connected := false

	if s.services.IPCClient != nil && s.services.IPCClient.IsConnected() {
		status = "connected"
		connected = true
	} else if s.services.IPCClient != nil {
		status = "reconnecting"
	}

	return map[string]interface{}{
		"status":    status,
		"connected": connected,
		"mode":      s.getDaemonModeString(),
		"url":       s.getDaemonURL(),
		"port":      s.services.DaemonPort,
	}
}

// getDaemonModeString returns the current daemon mode as a string.
func (s *SystemFacade) getDaemonModeString() string {
	if s.services.DaemonMode {
		return "daemon"
	}
	return "standalone"
}

// getDaemonURL returns the WebSocket URL for the daemon.
func (s *SystemFacade) getDaemonURL() string {
	return fmt.Sprintf("ws://localhost:%d", s.services.DaemonPort)
}

// SetDaemonPort updates the daemon port and saves to config.
func (s *SystemFacade) SetDaemonPort(port int) error {
	if port < 1024 || port > 65535 {
		return &AppError{Message: fmt.Sprintf("Port must be between 1024 and 65535, got %d", port)}
	}

	s.services.DaemonPort = port
	log.Printf("Daemon port updated to %d", port)

	return nil
}

// ReconnectToDaemon attempts to reconnect to the daemon.
func (s *SystemFacade) ReconnectToDaemon(ctx context.Context) error {
	log.Println("Reconnecting to daemon...")

	// Stop existing client
	s.stopDaemonClient(ctx)

	// Try to connect
	if err := s.connectToDaemon(ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to reconnect to daemon: %v", err)}
	}

	log.Println("Successfully reconnected to daemon")
	return nil
}

// SwitchToStandaloneMode disconnects from daemon and starts embedded poller.
func (s *SystemFacade) SwitchToStandaloneMode(ctx context.Context) error {
	log.Println("Switching to standalone mode...")

	// Stop daemon client
	s.stopDaemonClient(ctx)

	// Start embedded poller
	if err := s.StartPoller(ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to start poller: %v", err)}
	}

	log.Println("Switched to standalone mode successfully")
	return nil
}

// SwitchToDaemonMode stops embedded poller and connects to daemon.
func (s *SystemFacade) SwitchToDaemonMode(ctx context.Context) error {
	log.Println("Switching to daemon mode...")

	// Stop poller if running
	s.StopPoller()

	// Connect to daemon
	if err := s.connectToDaemon(ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to connect to daemon: %v", err)}
	}

	log.Println("Switched to daemon mode successfully")
	return nil
}

// ReplayStatus represents the current state of the replay engine.
type ReplayStatus struct {
	IsActive        bool    `json:"isActive"`
	IsPaused        bool    `json:"isPaused"`
	CurrentEntry    int     `json:"currentEntry"`
	TotalEntries    int     `json:"totalEntries"`
	PercentComplete float64 `json:"percentComplete"`
	Elapsed         float64 `json:"elapsed"`
	Speed           float64 `json:"speed"`
	Filter          string  `json:"filter"`
}

// TriggerReplayLogs sends a command to the daemon to replay historical logs.
// This is only available when connected to the daemon (not standalone mode).
func (s *SystemFacade) TriggerReplayLogs(ctx context.Context, clearData bool) error {
	log.Printf("[TriggerReplayLogs] Called with clearData=%v", clearData)

	s.ipcClientMu.Lock()
	defer s.ipcClientMu.Unlock()

	log.Printf("[TriggerReplayLogs] IPC client nil? %v", s.services.IPCClient == nil)
	if s.services.IPCClient != nil {
		log.Printf("[TriggerReplayLogs] IPC client connected? %v", s.services.IPCClient.IsConnected())
	}

	if s.services.IPCClient == nil || !s.services.IPCClient.IsConnected() {
		log.Printf("[TriggerReplayLogs] ERROR: Not connected to daemon")
		return &AppError{Message: "Not connected to daemon. Replay logs requires daemon mode."}
	}

	// Send replay_logs command via IPC
	message := map[string]interface{}{
		"type":       "replay_logs",
		"clear_data": clearData,
	}

	log.Printf("[TriggerReplayLogs] Sending IPC message: %+v", message)
	if err := s.services.IPCClient.Send(message); err != nil {
		log.Printf("[TriggerReplayLogs] ERROR: Failed to send: %v", err)
		return &AppError{Message: fmt.Sprintf("Failed to send replay command to daemon: %v", err)}
	}

	log.Printf("[TriggerReplayLogs] Successfully sent replay_logs command to daemon (clear_data: %v)", clearData)
	return nil
}

// StartReplayWithFileDialog opens a file dialog and starts replay with the selected file.
// Only works in daemon mode.
func (s *SystemFacade) StartReplayWithFileDialog(ctx context.Context, speed float64, filterType string, pauseOnDraft bool) error {
	log.Printf("[StartReplayWithFileDialog] Called with speed=%.1fx, filter=%s, pauseOnDraft=%v", speed, filterType, pauseOnDraft)

	// Check if connected to daemon
	s.ipcClientMu.Lock()
	connectedToDaemon := s.services.IPCClient != nil && s.services.IPCClient.IsConnected()
	s.ipcClientMu.Unlock()

	if !connectedToDaemon {
		return &AppError{Message: "Replay feature requires daemon mode. Please start the daemon service."}
	}

	// Open file dialog to select multiple log files
	filePaths, err := wailsruntime.OpenMultipleFilesDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "Select MTGA Log File(s) for Replay",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "MTGA Log Files (*.log)", Pattern: "*.log"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open file dialog: %v", err)}
	}

	// User cancelled or selected no files
	if len(filePaths) == 0 {
		return nil
	}

	log.Printf("[StartReplayWithFileDialog] Selected %d file(s)", len(filePaths))

	// Send start_replay command via IPC
	message := map[string]interface{}{
		"type":           "start_replay",
		"file_paths":     filePaths,
		"speed":          speed,
		"filter":         filterType,
		"pause_on_draft": pauseOnDraft,
	}

	log.Printf("[StartReplayWithFileDialog] Sending IPC message: %+v", message)
	s.ipcClientMu.Lock()
	err = s.services.IPCClient.Send(message)
	s.ipcClientMu.Unlock()

	if err != nil {
		log.Printf("[StartReplayWithFileDialog] ERROR: Failed to send: %v", err)
		return &AppError{Message: fmt.Sprintf("Failed to send start replay command to daemon: %v", err)}
	}

	log.Printf("[StartReplayWithFileDialog] Successfully sent start_replay command to daemon")
	return nil
}

// PauseReplay pauses the active replay.
// Only works in daemon mode.
func (s *SystemFacade) PauseReplay(ctx context.Context) error {
	log.Println("[PauseReplay] Called")

	// Check if connected to daemon
	s.ipcClientMu.Lock()
	connectedToDaemon := s.services.IPCClient != nil && s.services.IPCClient.IsConnected()
	s.ipcClientMu.Unlock()

	if !connectedToDaemon {
		return &AppError{Message: "Replay feature requires daemon mode."}
	}

	message := map[string]interface{}{
		"type": "pause_replay",
	}

	s.ipcClientMu.Lock()
	err := s.services.IPCClient.Send(message)
	s.ipcClientMu.Unlock()

	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to send pause replay command: %v", err)}
	}

	return nil
}

// ResumeReplay resumes a paused replay.
// Only works in daemon mode.
func (s *SystemFacade) ResumeReplay(ctx context.Context) error {
	log.Println("[ResumeReplay] Called")

	// Check if connected to daemon
	s.ipcClientMu.Lock()
	connectedToDaemon := s.services.IPCClient != nil && s.services.IPCClient.IsConnected()
	s.ipcClientMu.Unlock()

	if !connectedToDaemon {
		return &AppError{Message: "Replay feature requires daemon mode."}
	}

	message := map[string]interface{}{
		"type": "resume_replay",
	}

	s.ipcClientMu.Lock()
	err := s.services.IPCClient.Send(message)
	s.ipcClientMu.Unlock()

	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to send resume replay command: %v", err)}
	}

	return nil
}

// StopReplay stops the active replay.
// Only works in daemon mode.
func (s *SystemFacade) StopReplay(ctx context.Context) error {
	log.Println("[StopReplay] Called")

	// Check if connected to daemon
	s.ipcClientMu.Lock()
	connectedToDaemon := s.services.IPCClient != nil && s.services.IPCClient.IsConnected()
	s.ipcClientMu.Unlock()

	if !connectedToDaemon {
		return &AppError{Message: "Replay feature requires daemon mode."}
	}

	message := map[string]interface{}{
		"type": "stop_replay",
	}

	s.ipcClientMu.Lock()
	err := s.services.IPCClient.Send(message)
	s.ipcClientMu.Unlock()

	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to send stop replay command: %v", err)}
	}

	return nil
}

// GetReplayStatus returns the current replay status.
// Only works in daemon mode. UI should use WebSocket events for real-time updates.
func (s *SystemFacade) GetReplayStatus(ctx context.Context) (*ReplayStatus, error) {
	// Note: This method is deprecated and only returns inactive status.
	// The UI should subscribe to 'replay:*' WebSocket events for real-time updates.
	// Session status management is handled by the daemon's log processor, not the frontend.
	return &ReplayStatus{IsActive: false}, nil
}

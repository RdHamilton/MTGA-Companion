package main

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

	"github.com/ramonehamilton/MTGA-Companion/internal/ipc"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// App struct
type App struct {
	ctx         context.Context
	service     *storage.Service
	poller      *logreader.Poller
	pollerStop  context.CancelFunc
	pollerMu    sync.Mutex
	ipcClient   *ipc.Client
	ipcClientMu sync.Mutex
	daemonMode  bool
	daemonPort  int // Configurable daemon port
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		daemonPort: 9999, // Default daemon port
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Auto-initialize database with default path
	dbPath := getDefaultDBPath()
	if err := a.Initialize(dbPath); err != nil {
		log.Printf("Warning: Failed to initialize database at %s: %v", dbPath, err)
		log.Printf("You may need to configure the database path in Settings")
		return
	}

	// Try to connect to daemon first
	if err := a.connectToDaemon(); err != nil {
		log.Printf("Daemon not available, falling back to standalone mode: %v", err)

		// Auto-start log file poller for real-time updates (fallback mode)
		if err := a.StartPoller(); err != nil {
			log.Printf("Warning: Failed to start log file poller: %v", err)
			log.Printf("Real-time updates will not be available")
		}
	} else {
		log.Println("Connected to daemon successfully")
		a.daemonMode = true
	}
}

// getDefaultDBPath returns the default database path
func getDefaultDBPath() string {
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Error getting home directory: %v", err)
			return "data.db" // Fallback to current directory
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}
	return dbPath
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	// Stop daemon client if running
	a.stopDaemonClient()

	// Stop poller if running
	a.StopPoller()

	// Close database
	if a.service != nil {
		if err := a.service.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

// Initialize initializes the application with database path
func (a *App) Initialize(dbPath string) error {
	config := storage.DefaultConfig(dbPath)
	config.BusyTimeout = 10 * time.Second // Increase timeout to handle concurrent poller operations

	db, err := storage.Open(config)
	if err != nil {
		return err
	}
	a.service = storage.NewService(db)
	return nil
}

// GetMatches returns matches based on filter
func (a *App) GetMatches(filter models.StatsFilter) ([]*models.Match, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetMatches(a.ctx, filter)
}

// GetStats returns statistics based on filter
func (a *App) GetStats(filter models.StatsFilter) (*models.Statistics, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetStats(a.ctx, filter)
}

// GetTrendAnalysis returns trend analysis
func (a *App) GetTrendAnalysis(startDate, endDate time.Time, periodType string, formats []string) (*storage.TrendAnalysis, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetTrendAnalysisWithFormats(a.ctx, startDate, endDate, periodType, formats)
}

// GetStatsByDeck returns statistics grouped by deck
func (a *App) GetStatsByDeck(filter models.StatsFilter) (map[string]*models.Statistics, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	log.Printf("GetStatsByDeck called with filter: %+v", filter)
	result, err := a.service.GetStatsByDeck(a.ctx, filter)
	if err != nil {
		log.Printf("GetStatsByDeck error: %v", err)
		return nil, err
	}
	log.Printf("GetStatsByDeck returned %d decks", len(result))
	for deckName, stats := range result {
		log.Printf("  Deck: %s - Matches: %d, Wins: %d", deckName, stats.TotalMatches, stats.MatchesWon)
	}
	return result, nil
}

// GetRankProgressionTimeline returns rank progression timeline
func (a *App) GetRankProgressionTimeline(format string, startDate, endDate *time.Time, periodType storage.TimelinePeriod) (*storage.RankTimeline, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetRankProgressionTimeline(a.ctx, format, startDate, endDate, periodType)
}

// GetRankProgression returns rank progression for a format
func (a *App) GetRankProgression(format string) (*models.RankProgression, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetRankProgression(a.ctx, format)
}

// GetStatsByFormat returns statistics grouped by format
func (a *App) GetStatsByFormat(filter models.StatsFilter) (map[string]*models.Statistics, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetStatsByFormat(a.ctx, filter)
}

// GetPerformanceMetrics returns performance metrics
func (a *App) GetPerformanceMetrics(filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return a.service.GetPerformanceMetrics(a.ctx, filter)
}

// AppError represents an application error
type AppError struct {
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
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

// StartPoller starts the log file poller for real-time updates
func (a *App) StartPoller() error {
	a.pollerMu.Lock()
	defer a.pollerMu.Unlock()

	if a.service == nil {
		return &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}

	// Stop existing poller if running
	if a.poller != nil {
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

	a.poller = poller

	// Start poller
	updates := poller.Start()
	errChan := poller.Errors()

	// Create cancellable context
	pollerCtx, cancel := context.WithCancel(a.ctx)
	a.pollerStop = cancel

	// Start background goroutine to process updates
	go a.processPollerUpdates(pollerCtx, updates, errChan)

	log.Println("Log file poller started successfully")
	return nil
}

// StopPoller stops the log file poller
func (a *App) StopPoller() {
	a.pollerMu.Lock()
	defer a.pollerMu.Unlock()

	if a.pollerStop != nil {
		a.pollerStop()
		a.pollerStop = nil
	}

	if a.poller != nil {
		a.poller.Stop()
		a.poller = nil
		log.Println("Log file poller stopped")
	}
}

// processPollerUpdates processes new log entries in the background
func (a *App) processPollerUpdates(ctx context.Context, updates <-chan *logreader.LogEntry, errChan <-chan error) {
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
				a.processNewEntries(ctx, entryBuffer)
				entryBuffer = nil // Clear buffer
			}
		}
	}
}

// processNewEntries processes new log entries and updates statistics
func (a *App) processNewEntries(ctx context.Context, entries []*logreader.LogEntry) {
	dataUpdated := false

	// Parse arena stats from new entries
	arenaStats, err := logreader.ParseArenaStats(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse arena stats from new entries: %v", err)
		return
	}

	// Store new stats if we have match data
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		if err := a.service.StoreArenaStats(ctx, arenaStats, entries); err != nil {
			log.Printf("Warning: Failed to store arena stats from poller: %v", err)
		} else {
			log.Printf("✓ Updated statistics: %d new matches, %d new games",
				arenaStats.TotalMatches, arenaStats.TotalGames)
			dataUpdated = true

			// Try to infer deck IDs for the new matches
			inferredCount, err := a.service.InferDeckIDsForMatches(ctx)
			if err != nil {
				log.Printf("Warning: Failed to infer deck IDs: %v", err)
			} else if inferredCount > 0 {
				log.Printf("✓ Linked %d match(es) to decks", inferredCount)
			}
		}
	}

	// Parse and store decks
	deckLibrary, err := logreader.ParseDecks(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse decks from new entries: %v", err)
	} else if deckLibrary != nil && len(deckLibrary.Decks) > 0 {
		log.Printf("Found %d deck(s) in new entries", len(deckLibrary.Decks))
		// Store decks and infer deck IDs for matches
		// (Same logic as in CLI main.go lines 340-408)
		storedCount := 0
		processedCount := 0
		for _, deck := range deckLibrary.Decks {
			// Small delay between deck operations to avoid database lock contention
			if processedCount > 0 {
				time.Sleep(50 * time.Millisecond)
			}
			processedCount++
			// Convert card slices
			mainDeck := make([]struct {
				CardID   int
				Quantity int
			}, len(deck.MainDeck))
			for i, card := range deck.MainDeck {
				mainDeck[i].CardID = card.CardID
				mainDeck[i].Quantity = card.Quantity
			}

			sideboard := make([]struct {
				CardID   int
				Quantity int
			}, len(deck.Sideboard))
			for i, card := range deck.Sideboard {
				sideboard[i].CardID = card.CardID
				sideboard[i].Quantity = card.Quantity
			}

			// Handle timestamps
			created := deck.Created
			if created.IsZero() && !deck.Modified.IsZero() {
				created = deck.Modified
			} else if created.IsZero() {
				created = time.Now()
			}

			modified := deck.Modified
			if modified.IsZero() {
				modified = time.Now()
			}

			err := a.service.StoreDeckFromParser(
				ctx,
				deck.DeckID,
				deck.Name,
				deck.Format,
				deck.Description,
				created,
				modified,
				deck.LastPlayed,
				mainDeck,
				sideboard,
			)
			if err != nil {
				log.Printf("Warning: Failed to store deck %s: %v", deck.Name, err)
			} else {
				storedCount++
			}
		}

		if storedCount > 0 {
			log.Printf("✓ Stored %d/%d deck(s)", storedCount, len(deckLibrary.Decks))
			dataUpdated = true

			// Infer deck IDs for matches
			inferredCount, err := a.service.InferDeckIDsForMatches(ctx)
			if err != nil {
				log.Printf("Warning: Failed to infer deck IDs: %v", err)
			} else if inferredCount > 0 {
				log.Printf("✓ Linked %d match(es) to decks", inferredCount)
			}
		}
	}

	// Parse and store rank updates
	rankUpdates, err := logreader.ParseRankUpdates(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse rank updates from new entries: %v", err)
	} else if len(rankUpdates) > 0 {
		log.Printf("Found %d rank update(s) in new entries", len(rankUpdates))
		storedCount := 0
		for _, update := range rankUpdates {
			// Small delay between operations to avoid database lock contention
			if storedCount > 0 {
				time.Sleep(25 * time.Millisecond)
			}

			if err := a.service.StoreRankUpdate(ctx, update); err != nil {
				log.Printf("Warning: Failed to store rank update: %v", err)
			} else {
				storedCount++
			}
		}

		if storedCount > 0 {
			log.Printf("✓ Stored %d/%d rank update(s)", storedCount, len(rankUpdates))
			dataUpdated = true
		}
	}

	// Emit event to frontend if any data was updated
	if dataUpdated {
		matches := 0
		games := 0
		if arenaStats != nil {
			matches = arenaStats.TotalMatches
			games = arenaStats.TotalGames
		}

		wailsruntime.EventsEmit(a.ctx, "stats:updated", map[string]interface{}{
			"matches": matches,
			"games":   games,
		})
		log.Println("📡 Emitted stats:updated event to frontend")
	}
}

// connectToDaemon attempts to connect to the daemon WebSocket server.
func (a *App) connectToDaemon() error {
	a.ipcClientMu.Lock()
	defer a.ipcClientMu.Unlock()

	// Create IPC client
	wsURL := fmt.Sprintf("ws://localhost:%d", a.daemonPort)
	a.ipcClient = ipc.NewClient(wsURL)

	// Try to connect
	if err := a.ipcClient.Connect(); err != nil {
		a.ipcClient = nil
		return err
	}

	// Setup event handlers
	a.setupEventHandlers()

	// Start listening for events
	a.ipcClient.Start()

	return nil
}

// setupEventHandlers registers event handlers for daemon events.
func (a *App) setupEventHandlers() {
	// Handle stats:updated events from daemon
	a.ipcClient.On("stats:updated", func(data map[string]interface{}) {
		log.Printf("Received stats:updated event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "stats:updated", data)
	})

	// Handle rank:updated events from daemon
	a.ipcClient.On("rank:updated", func(data map[string]interface{}) {
		log.Printf("Received rank:updated event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "rank:updated", data)
	})

	// Handle deck:updated events from daemon
	a.ipcClient.On("deck:updated", func(data map[string]interface{}) {
		log.Printf("Received deck:updated event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "deck:updated", data)
	})

	// Handle daemon:status events
	a.ipcClient.On("daemon:status", func(data map[string]interface{}) {
		log.Printf("Daemon status: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "daemon:status", data)
	})

	// Handle daemon:connected events
	a.ipcClient.On("daemon:connected", func(data map[string]interface{}) {
		log.Printf("Daemon connected: %v", data)
	})

	// Handle daemon:error events
	a.ipcClient.On("daemon:error", func(data map[string]interface{}) {
		log.Printf("Daemon error: %v", data)

		// Forward error event to frontend
		wailsruntime.EventsEmit(a.ctx, "daemon:error", data)
	})
}

// stopDaemonClient stops the daemon IPC client if running.
func (a *App) stopDaemonClient() {
	a.ipcClientMu.Lock()
	defer a.ipcClientMu.Unlock()

	if a.ipcClient != nil {
		a.ipcClient.Stop()
		a.ipcClient = nil
		a.daemonMode = false
		log.Println("Daemon client stopped")
	}
}

// GetConnectionStatus returns current connection status for the frontend.
func (a *App) GetConnectionStatus() map[string]interface{} {
	a.ipcClientMu.Lock()
	defer a.ipcClientMu.Unlock()

	status := "standalone"
	connected := false

	if a.ipcClient != nil && a.ipcClient.IsConnected() {
		status = "connected"
		connected = true
	} else if a.ipcClient != nil {
		status = "reconnecting"
	}

	return map[string]interface{}{
		"status":    status,
		"connected": connected,
		"mode":      a.getDaemonModeString(),
		"url":       a.getDaemonURL(),
		"port":      a.daemonPort,
	}
}

// getDaemonModeString returns the current daemon mode as a string.
func (a *App) getDaemonModeString() string {
	if a.daemonMode {
		return "daemon"
	}
	return "standalone"
}

// getDaemonURL returns the WebSocket URL for the daemon.
func (a *App) getDaemonURL() string {
	return fmt.Sprintf("ws://localhost:%d", a.daemonPort)
}

// SetDaemonPort updates the daemon port and saves to config.
func (a *App) SetDaemonPort(port int) error {
	if port < 1024 || port > 65535 {
		return &AppError{Message: fmt.Sprintf("Port must be between 1024 and 65535, got %d", port)}
	}

	a.daemonPort = port
	log.Printf("Daemon port updated to %d", port)

	return nil
}

// ReconnectToDaemon attempts to reconnect to the daemon.
func (a *App) ReconnectToDaemon() error {
	log.Println("Reconnecting to daemon...")

	// Stop existing client
	a.stopDaemonClient()

	// Try to connect
	if err := a.connectToDaemon(); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to reconnect to daemon: %v", err)}
	}

	log.Println("Successfully reconnected to daemon")
	return nil
}

// SwitchToStandaloneMode disconnects from daemon and starts embedded poller.
func (a *App) SwitchToStandaloneMode() error {
	log.Println("Switching to standalone mode...")

	// Stop daemon client
	a.stopDaemonClient()

	// Start embedded poller
	if err := a.StartPoller(); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to start poller: %v", err)}
	}

	log.Println("Switched to standalone mode successfully")
	return nil
}

// SwitchToDaemonMode stops embedded poller and connects to daemon.
func (a *App) SwitchToDaemonMode() error {
	log.Println("Switching to daemon mode...")

	// Stop poller if running
	a.StopPoller()

	// Connect to daemon
	if err := a.connectToDaemon(); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to connect to daemon: %v", err)}
	}

	log.Println("Switched to daemon mode successfully")
	return nil
}

// GetActiveQuests returns all active (incomplete) quests.
func (a *App) GetActiveQuests() ([]*models.Quest, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	quests, err := a.service.Quests().GetActiveQuests()
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get active quests: %v", err)}
	}

	return quests, nil
}

// GetQuestHistory returns quest history with optional date range and limit.
func (a *App) GetQuestHistory(startDate, endDate string, limit int) ([]*models.Quest, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var start, end *time.Time

	// Parse start date if provided
	if startDate != "" {
		parsedStart, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Invalid start date format: %v", err)}
		}
		start = &parsedStart
	}

	// Parse end date if provided
	if endDate != "" {
		parsedEnd, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Invalid end date format: %v", err)}
		}
		end = &parsedEnd
	}

	quests, err := a.service.Quests().GetQuestHistory(start, end, limit)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get quest history: %v", err)}
	}

	return quests, nil
}

// GetQuestStats returns quest statistics with optional date range.
func (a *App) GetQuestStats(startDate, endDate string) (*models.QuestStats, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var start, end *time.Time

	// Parse start date if provided
	if startDate != "" {
		parsedStart, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Invalid start date format: %v", err)}
		}
		start = &parsedStart
	}

	// Parse end date if provided
	if endDate != "" {
		parsedEnd, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Invalid end date format: %v", err)}
		}
		end = &parsedEnd
	}

	stats, err := a.service.Quests().GetQuestStats(start, end)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get quest stats: %v", err)}
	}

	return stats, nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ramonehamilton/MTGA-Companion/internal/export"
	"github.com/ramonehamilton/MTGA-Companion/internal/ipc"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/setcache"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// App struct
type App struct {
	ctx         context.Context
	service     *storage.Service
	setFetcher  *setcache.Fetcher
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
		log.Printf("ERROR: Failed to initialize database at %s: %v", dbPath, err)

		// Show error dialog to user
		_, diagErr := wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
			Type:    wailsruntime.ErrorDialog,
			Title:   "Database Initialization Failed",
			Message: fmt.Sprintf("Failed to initialize database at:\n%s\n\nError: %v\n\nPlease check:\n• Directory permissions\n• Disk space\n• You can configure a different path in Settings", dbPath, err),
		})
		if diagErr != nil {
			log.Printf("Failed to show error dialog: %v", diagErr)
		}
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
			return "mtga.db" // Fallback to current directory
		}
		dbPath = filepath.Join(home, ".mtga-companion", "mtga.db")
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

	// Initialize set card fetcher
	scryfallClient := scryfall.NewClient()
	a.setFetcher = setcache.NewFetcher(scryfallClient, a.service.SetCardRepo())

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

	// Handle quest:updated events from daemon
	a.ipcClient.On("quest:updated", func(data map[string]interface{}) {
		log.Printf("Received quest:updated event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "quest:updated", data)
	})

	// Handle achievement:updated events from daemon
	a.ipcClient.On("achievement:updated", func(data map[string]interface{}) {
		log.Printf("Received achievement:updated event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "achievement:updated", data)
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

	// Handle replay:started events from daemon
	a.ipcClient.On("replay:started", func(data map[string]interface{}) {
		log.Printf("Received replay:started event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "replay:started", data)
	})

	// Handle replay:progress events from daemon
	a.ipcClient.On("replay:progress", func(data map[string]interface{}) {
		log.Printf("Received replay:progress event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "replay:progress", data)
	})

	// Handle replay:completed events from daemon
	a.ipcClient.On("replay:completed", func(data map[string]interface{}) {
		log.Printf("Received replay:completed event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "replay:completed", data)
	})

	// Handle replay:error events from daemon
	a.ipcClient.On("replay:error", func(data map[string]interface{}) {
		log.Printf("Received replay:error event from daemon: %v", data)

		// Forward event to frontend
		wailsruntime.EventsEmit(a.ctx, "replay:error", data)
	})

	// Handle disconnect events
	a.ipcClient.OnDisconnect(func() {
		log.Println("Daemon connection lost - notifying frontend")

		// Emit status change event to frontend
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "daemon:status", map[string]interface{}{
				"status":    "standalone",
				"connected": false,
			})
		}
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

		// Emit status change event to frontend
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "daemon:status", map[string]interface{}{
				"status":    "standalone",
				"connected": false,
			})
		}
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

// GetCurrentAccount returns the current account with all fields including daily/weekly wins.
func (a *App) GetCurrentAccount() (*models.Account, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	account, err := a.service.GetCurrentAccount(a.ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get current account: %v", err)}
	}

	return account, nil
}

// GetActiveEvents returns all currently active draft events.
func (a *App) GetActiveEvents() ([]*models.DraftEvent, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	events, err := a.service.GetActiveEvents(a.ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get active events: %v", err)}
	}

	return events, nil
}

// GetEventWinDistribution returns the distribution of event win-loss records.
func (a *App) GetEventWinDistribution() ([]*storage.EventWinDistribution, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	distribution, err := a.service.GetEventWinDistribution(a.ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get event win distribution: %v", err)}
	}

	return distribution, nil
}

// ExportToJSON exports all match data to a JSON file.
func (a *App) ExportToJSON() error {
	if a.service == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select save location
	filePath, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("mtga-matches-%s.json", time.Now().Format("2006-01-02")),
		Title:           "Export Matches to JSON",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open save dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Get all matches
	matches, err := a.service.GetMatches(a.ctx, models.StatsFilter{})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get matches: %v", err)}
	}

	// Export to JSON
	exporter := export.NewExporter(export.Options{
		Format:     export.FormatJSON,
		FilePath:   filePath,
		PrettyJSON: true,
		Overwrite:  true,
	})

	if err := exporter.Export(matches); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to export to JSON: %v", err)}
	}

	log.Printf("Successfully exported %d matches to %s", len(matches), filePath)
	return nil
}

// ExportToCSV exports all match data to a CSV file.
func (a *App) ExportToCSV() error {
	if a.service == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select save location
	filePath, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("mtga-matches-%s.csv", time.Now().Format("2006-01-02")),
		Title:           "Export Matches to CSV",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "CSV Files (*.csv)", Pattern: "*.csv"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open save dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Get all matches
	matches, err := a.service.GetMatches(a.ctx, models.StatsFilter{})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get matches: %v", err)}
	}

	// Export to CSV
	exporter := export.NewExporter(export.Options{
		Format:    export.FormatCSV,
		FilePath:  filePath,
		Overwrite: true,
	})

	if err := exporter.Export(matches); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to export to CSV: %v", err)}
	}

	log.Printf("Successfully exported %d matches to %s", len(matches), filePath)
	return nil
}

// ImportFromFile imports match data from a JSON file.
func (a *App) ImportFromFile() error {
	if a.service == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Prompt user to select file
	filePath, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Import Matches from JSON",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to open file dialog: %v", err)}
	}
	if filePath == "" {
		// User cancelled
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to read file: %v", err)}
	}

	// Parse JSON
	var matches []*models.Match
	if err := json.Unmarshal(data, &matches); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to parse JSON: %v", err)}
	}

	// Import matches (this would need a service method to handle duplicate checking)
	imported := 0
	for _, match := range matches {
		// Save each match (skip duplicates)
		if err := a.service.SaveMatch(a.ctx, match); err != nil {
			log.Printf("Warning: Failed to import match %s: %v", match.ID, err)
			continue
		}
		imported++
	}

	log.Printf("Successfully imported %d/%d matches from %s", imported, len(matches), filePath)
	return nil
}

// ClearAllData clears all match history from the database.
func (a *App) ClearAllData() error {
	if a.service == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Show confirmation dialog
	selection, err := wailsruntime.MessageDialog(a.ctx, wailsruntime.MessageDialogOptions{
		Type:          wailsruntime.QuestionDialog,
		Title:         "Clear All Data",
		Message:       "⚠️ WARNING: This will permanently delete all match history and statistics.\n\nThis action cannot be undone.\n\nAre you sure you want to continue?",
		DefaultButton: "No",
		Buttons:       []string{"Yes, Delete All Data", "No"},
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to show confirmation dialog: %v", err)}
	}
	if selection != "Yes, Delete All Data" {
		// User cancelled or clicked No
		return nil
	}

	// Delete all matches
	if err := a.service.ClearAllMatches(a.ctx); err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to clear data: %v", err)}
	}

	log.Println("Successfully cleared all match history")
	return nil
}

// TriggerReplayLogs sends a command to the daemon to replay historical logs.
// This is only available when connected to the daemon (not standalone mode).
func (a *App) TriggerReplayLogs(clearData bool) error {
	log.Printf("[TriggerReplayLogs] Called with clearData=%v", clearData)

	a.ipcClientMu.Lock()
	defer a.ipcClientMu.Unlock()

	log.Printf("[TriggerReplayLogs] IPC client nil? %v", a.ipcClient == nil)
	if a.ipcClient != nil {
		log.Printf("[TriggerReplayLogs] IPC client connected? %v", a.ipcClient.IsConnected())
	}

	if a.ipcClient == nil || !a.ipcClient.IsConnected() {
		log.Printf("[TriggerReplayLogs] ERROR: Not connected to daemon")
		return &AppError{Message: "Not connected to daemon. Replay logs requires daemon mode."}
	}

	// Send replay_logs command via IPC
	message := map[string]interface{}{
		"type":       "replay_logs",
		"clear_data": clearData,
	}

	log.Printf("[TriggerReplayLogs] Sending IPC message: %+v", message)
	if err := a.ipcClient.Send(message); err != nil {
		log.Printf("[TriggerReplayLogs] ERROR: Failed to send: %v", err)
		return &AppError{Message: fmt.Sprintf("Failed to send replay command to daemon: %v", err)}
	}

	log.Printf("[TriggerReplayLogs] Successfully sent replay_logs command to daemon (clear_data: %v)", clearData)
	return nil
}

// ==================== Draft Methods ====================

// GetActiveDraftSessions returns all active draft sessions.
func (a *App) GetActiveDraftSessions() ([]*models.DraftSession, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	sessions, err := a.service.DraftRepo().GetActiveSessions(a.ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get active draft sessions: %v", err)}
	}

	return sessions, nil
}

// GetCompletedDraftSessions returns recently completed draft sessions.
func (a *App) GetCompletedDraftSessions(limit int) ([]*models.DraftSession, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if limit <= 0 {
		limit = 20 // Default limit
	}

	sessions, err := a.service.DraftRepo().GetCompletedSessions(a.ctx, limit)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get completed draft sessions: %v", err)}
	}

	return sessions, nil
}

// GetDraftSession returns a draft session by ID.
func (a *App) GetDraftSession(sessionID string) (*models.DraftSession, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	session, err := a.service.DraftRepo().GetSession(a.ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft session: %v", err)}
	}

	return session, nil
}

// GetDraftPicks returns all picks for a draft session.
func (a *App) GetDraftPicks(sessionID string) ([]*models.DraftPickSession, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	picks, err := a.service.DraftRepo().GetPicksBySession(a.ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft picks: %v", err)}
	}

	return picks, nil
}

// GetDraftPacks returns all packs for a draft session.
func (a *App) GetDraftPacks(sessionID string) ([]*models.DraftPackSession, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	packs, err := a.service.DraftRepo().GetPacksBySession(a.ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft packs: %v", err)}
	}

	return packs, nil
}

// GetSetCards returns all cards for a given set code.
// Automatically fetches and caches from Scryfall if not already cached.
func (a *App) GetSetCards(setCode string) ([]*models.SetCard, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Check if set is already cached
	isCached, err := a.service.SetCardRepo().IsSetCached(a.ctx, setCode)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to check set cache: %v", err)}
	}

	// If not cached, fetch from Scryfall
	if !isCached {
		log.Printf("Set %s not cached, fetching from Scryfall...", setCode)
		count, err := a.setFetcher.FetchAndCacheSet(a.ctx, setCode)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Failed to fetch set cards from Scryfall: %v", err)}
		}
		log.Printf("Fetched and cached %d cards for set %s", count, setCode)
	}

	cards, err := a.service.SetCardRepo().GetCardsBySet(a.ctx, setCode)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get set cards: %v", err)}
	}

	return cards, nil
}

// FetchSetCards manually fetches and caches set cards from Scryfall.
// Returns the number of cards fetched and cached.
func (a *App) FetchSetCards(setCode string) (int, error) {
	if a.service == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	log.Printf("Manually fetching set %s from Scryfall...", setCode)
	count, err := a.setFetcher.FetchAndCacheSet(a.ctx, setCode)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to fetch set cards: %v", err)}
	}

	log.Printf("Successfully fetched and cached %d cards for set %s", count, setCode)
	return count, nil
}

// RefreshSetCards deletes and re-fetches all cards for a set.
func (a *App) RefreshSetCards(setCode string) (int, error) {
	if a.service == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	log.Printf("Refreshing set %s from Scryfall...", setCode)
	count, err := a.setFetcher.RefreshSet(a.ctx, setCode)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to refresh set cards: %v", err)}
	}

	log.Printf("Successfully refreshed %d cards for set %s", count, setCode)
	return count, nil
}

// GetCardByArenaID returns a card by its Arena ID.
func (a *App) GetCardByArenaID(arenaID string) (*models.SetCard, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	card, err := a.service.SetCardRepo().GetCardByArenaID(a.ctx, arenaID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card: %v", err)}
	}

	return card, nil
}

// FixDraftSessionStatuses updates draft sessions that should be marked as completed
// based on their pick counts (42 for Quick Draft, 45 for Premier Draft).
func (a *App) FixDraftSessionStatuses() (int, error) {
	if a.service == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	// Get all active sessions
	activeSessions, err := a.service.DraftRepo().GetActiveSessions(a.ctx)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to get active sessions: %v", err)}
	}

	updated := 0
	for _, session := range activeSessions {
		// Get picks for this session
		picks, err := a.service.DraftRepo().GetPicksBySession(a.ctx, session.ID)
		if err != nil {
			log.Printf("Failed to get picks for session %s: %v", session.ID, err)
			continue
		}

		// Determine expected picks based on draft type
		expectedPicks := 42 // Quick Draft
		if session.DraftType == "PremierDraft" {
			expectedPicks = 45
		}

		// If session has all expected picks, mark as completed
		if len(picks) >= expectedPicks {
			// Use the timestamp of the last pick as end time
			var endTime *time.Time
			if len(picks) > 0 {
				lastPickTime := picks[len(picks)-1].Timestamp
				endTime = &lastPickTime
			}

			err := a.service.DraftRepo().UpdateSessionStatus(a.ctx, session.ID, "completed", endTime)
			if err != nil {
				log.Printf("Failed to update session %s status: %v", session.ID, err)
				continue
			}

			log.Printf("Updated session %s to completed (%d/%d picks)", session.ID, len(picks), expectedPicks)
			updated++
		}
	}

	return updated, nil
}

// CardRatingWithTier represents a card rating with calculated tier.
type CardRatingWithTier struct {
	seventeenlands.CardRating
	Tier string `json:"tier"` // S, A, B, C, D, or F
}

// GetCardRatings returns 17Lands card ratings for a set and draft format.
func (a *App) GetCardRatings(setCode string, draftFormat string) ([]CardRatingWithTier, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get ratings from repository
	ratings, _, err := a.service.DraftRatingsRepo().GetCardRatings(a.ctx, setCode, draftFormat)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card ratings: %v", err)}
	}

	// Add tier to each rating
	result := make([]CardRatingWithTier, len(ratings))
	for i, rating := range ratings {
		result[i] = CardRatingWithTier{
			CardRating: rating,
			Tier:       calculateTier(rating.GIHWR),
		}
	}

	return result, nil
}

// GetCardRatingByArenaID returns the 17Lands rating for a specific card.
func (a *App) GetCardRatingByArenaID(setCode string, draftFormat string, arenaID string) (*CardRatingWithTier, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	rating, err := a.service.DraftRatingsRepo().GetCardRatingByArenaID(a.ctx, setCode, draftFormat, arenaID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get card rating: %v", err)}
	}

	if rating == nil {
		return nil, nil
	}

	return &CardRatingWithTier{
		CardRating: *rating,
		Tier:       calculateTier(rating.GIHWR),
	}, nil
}

// GetColorRatings returns 17Lands color combination ratings for a set and draft format.
func (a *App) GetColorRatings(setCode string, draftFormat string) ([]seventeenlands.ColorRating, error) {
	if a.service == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	ratings, _, err := a.service.DraftRatingsRepo().GetColorRatings(a.ctx, setCode, draftFormat)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get color ratings: %v", err)}
	}

	return ratings, nil
}

// calculateTier determines the tier (S, A, B, C, D, F) based on GIHWR percentage.
// Tier thresholds:
// - S Tier (Bombs): GIHWR ≥ 60% - Format-defining cards
// - A Tier: 57-59% - Excellent cards, high picks
// - B Tier: 54-56% - Good playables
// - C Tier: 51-53% - Filler/role players
// - D Tier: 48-50% - Below average
// - F Tier: < 48% - Avoid/sideboard
func calculateTier(gihwr float64) string {
	if gihwr >= 60 {
		return "S"
	}
	if gihwr >= 57 {
		return "A"
	}
	if gihwr >= 54 {
		return "B"
	}
	if gihwr >= 51 {
		return "C"
	}
	if gihwr >= 48 {
		return "D"
	}
	return "F"
}

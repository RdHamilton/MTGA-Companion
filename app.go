package main

import (
	"context"
	"fmt"
	"log"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/ramonehamilton/MTGA-Companion/internal/daemon/manager"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/metrics"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/grading"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/insights"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/pickquality"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/prediction"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// App struct - Refactored to use Facade pattern (v1.2 Phase 1)
// All business logic has been moved to domain-specific facades in internal/gui/
// This struct now serves as a thin delegation layer between the Wails frontend and the backend.
type App struct {
	ctx context.Context

	// Facades - Domain-specific interfaces to backend services
	matchFacade      *gui.MatchFacade
	draftFacade      *gui.DraftFacade
	cardFacade       *gui.CardFacade
	deckFacade       *gui.DeckFacade
	exportFacade     *gui.ExportFacade
	systemFacade     *gui.SystemFacade
	collectionFacade *gui.CollectionFacade

	// Shared services used by facades
	services *gui.Services
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize daemon manager configuration
	daemonConfig := manager.DefaultConfig()

	// Initialize shared services
	services := &gui.Services{
		DaemonPort:      daemonConfig.Port, // Use daemon manager's default port
		DaemonManager:   manager.New(daemonConfig),
		DaemonAutoStart: false, // Disabled by default until daemon binary is bundled
		DraftMetrics:    metrics.NewDraftMetrics(),
	}

	// Create system facade first (it contains the event dispatcher)
	systemFacade := gui.NewSystemFacade(services)

	// Create app with facades, passing event dispatcher to those that need it
	return &App{
		services:         services,
		matchFacade:      gui.NewMatchFacade(services),
		draftFacade:      gui.NewDraftFacade(services),
		cardFacade:       gui.NewCardFacade(services),
		deckFacade:       gui.NewDeckFacade(services),
		exportFacade:     gui.NewExportFacade(services, systemFacade.GetEventDispatcher()),
		systemFacade:     systemFacade,
		collectionFacade: gui.NewCollectionFacade(services),
	}
}

// startup is called when the app starts. The context is saved so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.services.Context = ctx

	// Auto-initialize database with default path
	if err := a.systemFacade.Initialize(ctx, ""); err != nil {
		log.Printf("ERROR: Failed to initialize database: %v", err)

		// Show error dialog to user
		_, diagErr := wailsruntime.MessageDialog(ctx, wailsruntime.MessageDialogOptions{
			Type:    wailsruntime.ErrorDialog,
			Title:   "Database Initialization Failed",
			Message: fmt.Sprintf("Failed to initialize database\n\nError: %v\n\nPlease check:\n• Directory permissions\n• Disk space\n• You can configure a different path in Settings", err),
		})
		if diagErr != nil {
			log.Printf("Failed to show error dialog: %v", diagErr)
		}
		return
	}

	// Auto-start daemon subprocess if enabled
	if a.services.DaemonAutoStart {
		go func() {
			if err := a.systemFacade.AutoStartDaemon(ctx); err != nil {
				log.Printf("Failed to auto-start daemon: %v", err)
			} else {
				// Wait for daemon to start, then connect
				time.Sleep(2 * time.Second)
				if err := a.systemFacade.ReconnectToDaemon(ctx); err != nil {
					log.Printf("Failed to connect to auto-started daemon: %v", err)
				}
			}
		}()
	}

	// Try to reconnect to daemon first, fall back to standalone mode if not available
	if err := a.systemFacade.ReconnectToDaemon(ctx); err != nil {
		log.Printf("Daemon not available, falling back to standalone mode: %v", err)

		// Auto-start log file poller for real-time updates (fallback mode)
		if err := a.systemFacade.StartPoller(ctx); err != nil {
			log.Printf("Warning: Failed to start log file poller: %v", err)
			log.Printf("Real-time updates will not be available")
		}
	} else {
		log.Println("Connected to daemon successfully")
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	log.Println("App shutting down...")

	// Stop poller if running
	a.systemFacade.StopPoller()

	// Stop daemon subprocess if running (issue #597 - graceful shutdown)
	if err := a.systemFacade.StopDaemonProcess(ctx); err != nil {
		log.Printf("Warning: Failed to stop daemon subprocess: %v", err)
	}
}

// ========================================
// System Methods (SystemFacade)
// ========================================

// Initialize initializes the database connection
func (a *App) Initialize(dbPath string) error {
	return a.systemFacade.Initialize(a.ctx, dbPath)
}

// StartPoller starts the log file poller
func (a *App) StartPoller() error {
	return a.systemFacade.StartPoller(a.ctx)
}

// StopPoller stops the log file poller
func (a *App) StopPoller() {
	a.systemFacade.StopPoller()
}

// GetConnectionStatus returns the current connection status
func (a *App) GetConnectionStatus() *gui.ConnectionStatus {
	return a.systemFacade.GetConnectionStatus()
}

// SetDaemonPort sets the daemon port
func (a *App) SetDaemonPort(port int) error {
	return a.systemFacade.SetDaemonPort(port)
}

// ReconnectToDaemon attempts to reconnect to the daemon
func (a *App) ReconnectToDaemon() error {
	return a.systemFacade.ReconnectToDaemon(a.ctx)
}

// SwitchToStandaloneMode switches to standalone mode (embedded poller)
func (a *App) SwitchToStandaloneMode() error {
	return a.systemFacade.SwitchToStandaloneMode(a.ctx)
}

// SwitchToDaemonMode switches to daemon mode (IPC connection)
func (a *App) SwitchToDaemonMode() error {
	return a.systemFacade.SwitchToDaemonMode(a.ctx)
}

// ========================================
// Daemon Subprocess Management
// ========================================

// StartDaemonProcess starts the daemon subprocess
func (a *App) StartDaemonProcess() error {
	return a.systemFacade.StartDaemonProcess(a.ctx)
}

// StopDaemonProcess stops the daemon subprocess
func (a *App) StopDaemonProcess() error {
	return a.systemFacade.StopDaemonProcess(a.ctx)
}

// RestartDaemonProcess restarts the daemon subprocess
func (a *App) RestartDaemonProcess() error {
	return a.systemFacade.RestartDaemonProcess(a.ctx)
}

// GetDaemonProcessStatus returns the status of the daemon subprocess
func (a *App) GetDaemonProcessStatus() *gui.DaemonProcessStatus {
	return a.systemFacade.GetDaemonProcessStatus()
}

// SetDaemonAutoStart enables or disables daemon auto-start
func (a *App) SetDaemonAutoStart(enabled bool) {
	a.systemFacade.SetDaemonAutoStart(enabled)
}

// IsDaemonAutoStartEnabled returns whether daemon auto-start is enabled
func (a *App) IsDaemonAutoStartEnabled() bool {
	return a.systemFacade.IsDaemonAutoStartEnabled()
}

// TriggerReplayLogs triggers a replay of the log file
func (a *App) TriggerReplayLogs(clearData bool) error {
	return a.systemFacade.TriggerReplayLogs(a.ctx, clearData)
}

// StartReplayWithFileDialog starts replay with a file dialog
func (a *App) StartReplayWithFileDialog(speed float64, filterType string, pauseOnDraft bool) error {
	return a.systemFacade.StartReplayWithFileDialog(a.ctx, speed, filterType, pauseOnDraft)
}

// PauseReplay pauses the current replay
func (a *App) PauseReplay() error {
	return a.systemFacade.PauseReplay(a.ctx)
}

// ResumeReplay resumes the current replay
func (a *App) ResumeReplay() error {
	return a.systemFacade.ResumeReplay(a.ctx)
}

// StopReplay stops the current replay
func (a *App) StopReplay() error {
	return a.systemFacade.StopReplay(a.ctx)
}

// GetReplayStatus returns the current replay status
func (a *App) GetReplayStatus() (*gui.ReplayStatus, error) {
	return a.systemFacade.GetReplayStatus(a.ctx)
}

// GetLogReplayProgress returns an empty LogReplayProgress struct.
// This method exists to expose the type to Wails for TypeScript code generation.
// Actual progress is delivered via 'replay:progress' events.
func (a *App) GetLogReplayProgress() (*gui.LogReplayProgress, error) {
	return a.systemFacade.GetLogReplayProgress(a.ctx)
}

// ========================================
// Event Type Exposers
// These methods exist solely to expose event payload types to Wails for
// TypeScript code generation. They return empty structs and are not called
// at runtime. Actual event data is delivered via EventsEmit.
// ========================================

// GetStatsUpdatedEvent exposes StatsUpdatedEvent type to Wails.
func (a *App) GetStatsUpdatedEvent() (*gui.StatsUpdatedEvent, error) {
	return a.systemFacade.GetStatsUpdatedEvent(a.ctx)
}

// GetRankUpdatedEvent exposes RankUpdatedEvent type to Wails.
func (a *App) GetRankUpdatedEvent() (*gui.RankUpdatedEvent, error) {
	return a.systemFacade.GetRankUpdatedEvent(a.ctx)
}

// GetQuestUpdatedEvent exposes QuestUpdatedEvent type to Wails.
func (a *App) GetQuestUpdatedEvent() (*gui.QuestUpdatedEvent, error) {
	return a.systemFacade.GetQuestUpdatedEvent(a.ctx)
}

// GetDraftUpdatedEvent exposes DraftUpdatedEvent type to Wails.
func (a *App) GetDraftUpdatedEvent() (*gui.DraftUpdatedEvent, error) {
	return a.systemFacade.GetDraftUpdatedEvent(a.ctx)
}

// GetDeckUpdatedEvent exposes DeckUpdatedEvent type to Wails.
func (a *App) GetDeckUpdatedEvent() (*gui.DeckUpdatedEvent, error) {
	return a.systemFacade.GetDeckUpdatedEvent(a.ctx)
}

// GetAchievementUpdatedEvent exposes AchievementUpdatedEvent type to Wails.
func (a *App) GetAchievementUpdatedEvent() (*gui.AchievementUpdatedEvent, error) {
	return a.systemFacade.GetAchievementUpdatedEvent(a.ctx)
}

// GetDaemonErrorEvent exposes DaemonErrorEvent type to Wails.
func (a *App) GetDaemonErrorEvent() (*gui.DaemonErrorEvent, error) {
	return a.systemFacade.GetDaemonErrorEvent(a.ctx)
}

// GetReplayErrorEvent exposes ReplayErrorEvent type to Wails.
func (a *App) GetReplayErrorEvent() (*gui.ReplayErrorEvent, error) {
	return a.systemFacade.GetReplayErrorEvent(a.ctx)
}

// GetReplayDraftDetectedEvent exposes ReplayDraftDetectedEvent type to Wails.
func (a *App) GetReplayDraftDetectedEvent() (*gui.ReplayDraftDetectedEvent, error) {
	return a.systemFacade.GetReplayDraftDetectedEvent(a.ctx)
}

// ========================================
// Match & Statistics Methods (MatchFacade)
// ========================================

// GetMatches returns matches based on the provided filter
func (a *App) GetMatches(filter models.StatsFilter) ([]*models.Match, error) {
	return a.matchFacade.GetMatches(a.ctx, filter)
}

// GetMatchGames returns all games for a specific match
func (a *App) GetMatchGames(matchID string) ([]*models.Game, error) {
	return a.matchFacade.GetMatchGames(a.ctx, matchID)
}

// GetStats returns statistics based on the provided filter
func (a *App) GetStats(filter models.StatsFilter) (*models.Statistics, error) {
	return a.matchFacade.GetStats(a.ctx, filter)
}

// GetTrendAnalysis returns trend analysis for the specified time period
func (a *App) GetTrendAnalysis(startDate, endDate time.Time, periodType string, formats []string) (*storage.TrendAnalysis, error) {
	return a.matchFacade.GetTrendAnalysis(a.ctx, startDate, endDate, periodType, formats)
}

// GetStatsByDeck returns statistics grouped by deck
func (a *App) GetStatsByDeck(filter models.StatsFilter) (map[string]*models.Statistics, error) {
	return a.matchFacade.GetStatsByDeck(a.ctx, filter)
}

// GetRankProgressionTimeline returns rank progression timeline for a format
func (a *App) GetRankProgressionTimeline(format string, startDate, endDate *time.Time, periodType storage.TimelinePeriod) (*storage.RankTimeline, error) {
	return a.matchFacade.GetRankProgressionTimeline(a.ctx, format, startDate, endDate, periodType)
}

// GetRankProgression returns rank progression for a specific format
func (a *App) GetRankProgression(format string) (*models.RankProgression, error) {
	return a.matchFacade.GetRankProgression(a.ctx, format)
}

// GetStatsByFormat returns statistics grouped by format
func (a *App) GetStatsByFormat(filter models.StatsFilter) (map[string]*models.Statistics, error) {
	return a.matchFacade.GetStatsByFormat(a.ctx, filter)
}

// GetPerformanceMetrics returns performance metrics based on the filter
func (a *App) GetPerformanceMetrics(filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	return a.matchFacade.GetPerformanceMetrics(a.ctx, filter)
}

// GetActiveQuests returns all active quests
func (a *App) GetActiveQuests() ([]*models.Quest, error) {
	return a.matchFacade.GetActiveQuests(a.ctx)
}

// GetQuestHistory returns quest history for the specified date range
func (a *App) GetQuestHistory(startDate, endDate string, limit int) ([]*models.Quest, error) {
	return a.matchFacade.GetQuestHistory(a.ctx, startDate, endDate, limit)
}

// GetQuestStats returns quest statistics for the specified date range
func (a *App) GetQuestStats(startDate, endDate string) (*models.QuestStats, error) {
	return a.matchFacade.GetQuestStats(a.ctx, startDate, endDate)
}

// GetCurrentAccount returns the current account information
func (a *App) GetCurrentAccount() (*models.Account, error) {
	return a.matchFacade.GetCurrentAccount(a.ctx)
}

// ========================================
// Draft Methods (DraftFacade)
// ========================================

// GetActiveDraftSessions returns all active draft sessions
func (a *App) GetActiveDraftSessions() ([]*models.DraftSession, error) {
	return a.draftFacade.GetActiveDraftSessions(a.ctx)
}

// GetCompletedDraftSessions returns completed draft sessions
func (a *App) GetCompletedDraftSessions(limit int) ([]*models.DraftSession, error) {
	return a.draftFacade.GetCompletedDraftSessions(a.ctx, limit)
}

// GetDraftSession returns a specific draft session
func (a *App) GetDraftSession(sessionID string) (*models.DraftSession, error) {
	return a.draftFacade.GetDraftSession(a.ctx, sessionID)
}

// GetDraftPicks returns all picks for a draft session
func (a *App) GetDraftPicks(sessionID string) ([]*models.DraftPickSession, error) {
	return a.draftFacade.GetDraftPicks(a.ctx, sessionID)
}

// GetDraftDeckMetrics returns deck metrics for a draft session
func (a *App) GetDraftDeckMetrics(sessionID string) (*models.DeckMetrics, error) {
	return a.draftFacade.GetDraftDeckMetrics(a.ctx, sessionID)
}

// GetDraftPerformanceMetrics returns draft performance metrics
func (a *App) GetDraftPerformanceMetrics() *metrics.DraftStats {
	return a.draftFacade.GetDraftPerformanceMetrics(a.ctx)
}

// ResetDraftPerformanceMetrics resets draft performance metrics
func (a *App) ResetDraftPerformanceMetrics() {
	a.draftFacade.ResetDraftPerformanceMetrics(a.ctx)
}

// GetDraftPacks returns all packs for a draft session
func (a *App) GetDraftPacks(sessionID string) ([]*models.DraftPackSession, error) {
	return a.draftFacade.GetDraftPacks(a.ctx, sessionID)
}

// GetMissingCards returns missing cards analysis for a draft pack
func (a *App) GetMissingCards(sessionID string, packNum, pickNum int) (*models.MissingCardsAnalysis, error) {
	return a.draftFacade.GetMissingCards(a.ctx, sessionID, packNum, pickNum)
}

// AnalyzeSessionPickQuality analyzes pick quality for all picks in a session
func (a *App) AnalyzeSessionPickQuality(sessionID string) error {
	return a.draftFacade.AnalyzeSessionPickQuality(a.ctx, sessionID)
}

// GetPickAlternatives returns pick alternatives for a specific pick
func (a *App) GetPickAlternatives(sessionID string, packNum, pickNum int) (*pickquality.PickQuality, error) {
	return a.draftFacade.GetPickAlternatives(a.ctx, sessionID, packNum, pickNum)
}

// CalculateDraftGrade calculates and stores the draft grade for a session
func (a *App) CalculateDraftGrade(sessionID string) (*grading.DraftGrade, error) {
	return a.draftFacade.CalculateDraftGrade(a.ctx, sessionID)
}

// GetDraftGrade returns the stored draft grade for a session
func (a *App) GetDraftGrade(sessionID string) (*grading.DraftGrade, error) {
	return a.draftFacade.GetDraftGrade(a.ctx, sessionID)
}

// PredictDraftWinRate predicts the win rate for a draft deck
func (a *App) PredictDraftWinRate(sessionID string) (*prediction.DeckPrediction, error) {
	return a.draftFacade.PredictDraftWinRate(a.ctx, sessionID)
}

// GetDraftWinRatePrediction returns the stored win rate prediction for a draft
func (a *App) GetDraftWinRatePrediction(sessionID string) (*prediction.DeckPrediction, error) {
	return a.draftFacade.GetDraftWinRatePrediction(a.ctx, sessionID)
}

// RecalculateAllDraftGrades recalculates grades for all draft sessions
func (a *App) RecalculateAllDraftGrades() (int, error) {
	// Pass RefreshSetCards as the SetCardRefresher
	return a.draftFacade.RecalculateAllDraftGrades(a.ctx, a.cardFacade.RefreshSetCards)
}

// FixDraftSessionStatuses fixes draft session statuses
func (a *App) FixDraftSessionStatuses() (int, error) {
	// Create replay status checker function
	checkReplayActive := func() (bool, error) {
		status, err := a.systemFacade.GetReplayStatus(a.ctx)
		if err != nil {
			return false, err
		}
		return status.IsActive, nil
	}
	return a.draftFacade.FixDraftSessionStatuses(a.ctx, checkReplayActive)
}

// RepairDraftSession repairs a specific draft session
func (a *App) RepairDraftSession(sessionID string) error {
	return a.draftFacade.RepairDraftSession(a.ctx, sessionID)
}

// GetFormatInsights returns format insights (archetype performance, color pair win rates)
func (a *App) GetFormatInsights(setCode, draftFormat string) (*insights.FormatInsights, error) {
	return a.draftFacade.GetFormatInsights(a.ctx, setCode, draftFormat)
}

// GetArchetypeCards returns top cards for a specific archetype (color pair)
func (a *App) GetArchetypeCards(setCode, draftFormat, colors string) (*insights.ArchetypeCards, error) {
	return a.draftFacade.GetArchetypeCards(a.ctx, setCode, draftFormat, colors)
}

// ========================================
// Card Methods (CardFacade)
// ========================================

// GetSetCards returns all cards for a set, fetching from Scryfall if not cached
func (a *App) GetSetCards(setCode string) ([]*models.SetCard, error) {
	return a.cardFacade.GetSetCards(a.ctx, setCode)
}

// FetchSetCards manually fetches and caches set cards from Scryfall
func (a *App) FetchSetCards(setCode string) (int, error) {
	return a.cardFacade.FetchSetCards(a.ctx, setCode)
}

// RefreshSetCards deletes and re-fetches all cards for a set
func (a *App) RefreshSetCards(setCode string) (int, error) {
	return a.cardFacade.RefreshSetCards(a.ctx, setCode)
}

// FetchSetRatings fetches and caches 17Lands card ratings for a set and draft format
func (a *App) FetchSetRatings(setCode string, draftFormat string) error {
	return a.cardFacade.FetchSetRatings(a.ctx, setCode, draftFormat)
}

// RefreshSetRatings deletes and re-fetches 17Lands ratings for a set and draft format
func (a *App) RefreshSetRatings(setCode string, draftFormat string) error {
	return a.cardFacade.RefreshSetRatings(a.ctx, setCode, draftFormat)
}

// ClearDatasetCache clears all cached 17Lands datasets to free up disk space
func (a *App) ClearDatasetCache() error {
	return a.cardFacade.ClearDatasetCache(a.ctx)
}

// GetDatasetSource returns the data source for a given set and format ("s3" or "web_api")
func (a *App) GetDatasetSource(setCode string, draftFormat string) string {
	return a.cardFacade.GetDatasetSource(a.ctx, setCode, draftFormat)
}

// GetCardByArenaID returns a card by its Arena ID
func (a *App) GetCardByArenaID(arenaID string) (*models.SetCard, error) {
	return a.cardFacade.GetCardByArenaID(a.ctx, arenaID)
}

// GetCardRatings returns all card ratings for a set and draft format with tier information
func (a *App) GetCardRatings(setCode string, draftFormat string) ([]gui.CardRatingWithTier, error) {
	return a.cardFacade.GetCardRatings(a.ctx, setCode, draftFormat)
}

// GetCardRatingByArenaID returns the 17Lands rating for a specific card
func (a *App) GetCardRatingByArenaID(setCode string, draftFormat string, arenaID string) (*gui.CardRatingWithTier, error) {
	return a.cardFacade.GetCardRatingByArenaID(a.ctx, setCode, draftFormat, arenaID)
}

// GetColorRatings returns 17Lands color combination ratings for a set and draft format
func (a *App) GetColorRatings(setCode string, draftFormat string) ([]seventeenlands.ColorRating, error) {
	return a.cardFacade.GetColorRatings(a.ctx, setCode, draftFormat)
}

// GetSetInfo returns information about a specific set including its icon URL
func (a *App) GetSetInfo(setCode string) (*gui.SetInfo, error) {
	return a.cardFacade.GetSetInfo(a.ctx, setCode)
}

// GetAllSetInfo returns information about all known sets
func (a *App) GetAllSetInfo() ([]*gui.SetInfo, error) {
	return a.cardFacade.GetAllSetInfo(a.ctx)
}

// ========================================
// Export Methods (ExportFacade)
// ========================================

// ExportToJSON exports all data to JSON format
func (a *App) ExportToJSON() error {
	return a.exportFacade.ExportToJSON(a.ctx)
}

// ExportToCSV exports all data to CSV format
func (a *App) ExportToCSV() error {
	return a.exportFacade.ExportToCSV(a.ctx)
}

// ImportFromFile imports data from a file
func (a *App) ImportFromFile() error {
	return a.exportFacade.ImportFromFile(a.ctx)
}

// ClearAllData clears all data from the database
func (a *App) ClearAllData() error {
	return a.exportFacade.ClearAllData(a.ctx)
}

// ImportLogFile imports a log file (Player.log or Player-prev.log)
func (a *App) ImportLogFile() (*gui.ImportLogFileResult, error) {
	return a.exportFacade.ImportLogFile(a.ctx)
}

// ========================================
// Deck Builder Methods (DeckFacade)
// ========================================

// CreateDeck creates a new deck
func (a *App) CreateDeck(name, format, source string, draftEventID *string) (*models.Deck, error) {
	return a.deckFacade.CreateDeck(a.ctx, name, format, source, draftEventID)
}

// GetDeck retrieves a deck by ID with its cards and tags
func (a *App) GetDeck(deckID string) (*gui.DeckWithCards, error) {
	return a.deckFacade.GetDeck(a.ctx, deckID)
}

// ListDecks retrieves all decks for the current account
func (a *App) ListDecks() ([]*gui.DeckListItem, error) {
	return a.deckFacade.ListDecks(a.ctx)
}

// UpdateDeck updates an existing deck's metadata
func (a *App) UpdateDeck(deck *models.Deck) error {
	return a.deckFacade.UpdateDeck(a.ctx, deck)
}

// DeleteDeck deletes a deck and all its cards
func (a *App) DeleteDeck(deckID string) error {
	return a.deckFacade.DeleteDeck(a.ctx, deckID)
}

// CloneDeck creates a copy of an existing deck
func (a *App) CloneDeck(deckID, newName string) (*models.Deck, error) {
	return a.deckFacade.CloneDeck(a.ctx, deckID, newName)
}

// AddCard adds a card to a deck
func (a *App) AddCard(deckID string, cardID, quantity int, board string, fromDraft bool) error {
	return a.deckFacade.AddCard(a.ctx, deckID, cardID, quantity, board, fromDraft)
}

// RemoveCard removes a card from a deck
func (a *App) RemoveCard(deckID string, cardID int, board string) error {
	return a.deckFacade.RemoveCard(a.ctx, deckID, cardID, board)
}

// GetDeckLibrary retrieves all decks with advanced filtering and sorting
func (a *App) GetDeckLibrary(filter *gui.DeckLibraryFilter) ([]*gui.DeckListItem, error) {
	return a.deckFacade.GetDeckLibrary(a.ctx, filter)
}

// GetDecksBySource retrieves decks filtered by source (draft/constructed/imported)
func (a *App) GetDecksBySource(source string) ([]*gui.DeckListItem, error) {
	return a.deckFacade.GetDecksBySource(a.ctx, source)
}

// GetDecksByFormat retrieves decks filtered by format (Standard, Historic, etc.)
func (a *App) GetDecksByFormat(format string) ([]*gui.DeckListItem, error) {
	return a.deckFacade.GetDecksByFormat(a.ctx, format)
}

// GetDecksByTags retrieves decks that have ALL specified tags
func (a *App) GetDecksByTags(tags []string) ([]*gui.DeckListItem, error) {
	return a.deckFacade.GetDecksByTags(a.ctx, tags)
}

// AddTag adds a tag to a deck for categorization
func (a *App) AddTag(deckID, tag string) error {
	return a.deckFacade.AddTag(a.ctx, deckID, tag)
}

// RemoveTag removes a tag from a deck
func (a *App) RemoveTag(deckID, tag string) error {
	return a.deckFacade.RemoveTag(a.ctx, deckID, tag)
}

// GetDeckByDraftEvent retrieves the deck associated with a draft event
func (a *App) GetDeckByDraftEvent(draftEventID string) (*gui.DeckWithCards, error) {
	return a.deckFacade.GetDeckByDraftEvent(a.ctx, draftEventID)
}

// ImportDeck imports a deck from text (Arena format or plain text)
func (a *App) ImportDeck(req *gui.ImportDeckRequest) (*gui.ImportDeckResponse, error) {
	return a.deckFacade.ImportDeck(a.ctx, req)
}

// ExportDeck exports a deck to the requested format
func (a *App) ExportDeck(req *gui.ExportDeckRequest) (*gui.ExportDeckResponse, error) {
	return a.deckFacade.ExportDeck(a.ctx, req)
}

// GetRecommendations returns card recommendations for a deck
func (a *App) GetRecommendations(req *gui.GetRecommendationsRequest) (*gui.GetRecommendationsResponse, error) {
	return a.deckFacade.GetRecommendations(a.ctx, req)
}

// ExplainRecommendation explains why a card is recommended for a deck
func (a *App) ExplainRecommendation(req *gui.ExplainRecommendationRequest) (*gui.ExplainRecommendationResponse, error) {
	return a.deckFacade.ExplainRecommendation(a.ctx, req)
}

// GetDeckStatistics calculates comprehensive deck statistics
func (a *App) GetDeckStatistics(deckID string) (*gui.DeckStatistics, error) {
	return a.deckFacade.GetDeckStatistics(a.ctx, deckID)
}

// GetDeckPerformance retrieves performance metrics for a deck
func (a *App) GetDeckPerformance(deckID string) (*models.DeckPerformance, error) {
	return a.deckFacade.GetDeckPerformance(a.ctx, deckID)
}

// ValidateDraftDeck validates that all cards in a draft deck are from the associated draft
func (a *App) ValidateDraftDeck(deckID string) (bool, error) {
	return a.deckFacade.ValidateDraftDeck(a.ctx, deckID)
}

// ExportDeckToFile exports a deck and shows a native save dialog
func (a *App) ExportDeckToFile(deckID string) error {
	return a.deckFacade.ExportDeckToFile(a.ctx, deckID)
}

// ValidateDeckWithDialog validates a deck and shows result in a native dialog
func (a *App) ValidateDeckWithDialog(deckID string) error {
	return a.deckFacade.ValidateDeckWithDialog(a.ctx, deckID)
}

// ========================================
// Collection Methods (CollectionFacade)
// ========================================

// GetCollection returns the player's collection with optional filtering.
func (a *App) GetCollection(filter *gui.CollectionFilter) (*gui.CollectionResponse, error) {
	return a.collectionFacade.GetCollection(a.ctx, filter)
}

// GetCollectionStats returns summary statistics about the collection.
func (a *App) GetCollectionStats() (*gui.CollectionStats, error) {
	return a.collectionFacade.GetCollectionStats(a.ctx)
}

// GetSetCompletion returns set completion statistics.
func (a *App) GetSetCompletion() ([]*models.SetCompletion, error) {
	return a.collectionFacade.GetSetCompletion(a.ctx)
}

// GetRecentCollectionChanges returns recent collection changes.
func (a *App) GetRecentCollectionChanges(limit int) ([]*gui.CollectionChangeEntry, error) {
	return a.collectionFacade.GetRecentChanges(a.ctx, limit)
}

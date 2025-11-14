package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// App struct
type App struct {
	ctx     context.Context
	service *storage.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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
	if a.service != nil {
		a.service.Close()
	}
}

// Initialize initializes the application with database path
func (a *App) Initialize(dbPath string) error {
	db, err := storage.Open(&storage.Config{
		Path: dbPath,
	})
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

package main

import (
	"context"
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
	return a.service.GetMatches(a.ctx, filter)
}

// GetStats returns statistics based on filter
func (a *App) GetStats(filter models.StatsFilter) (*models.Statistics, error) {
	return a.service.GetStats(a.ctx, filter)
}

// GetTrendAnalysis returns trend analysis
func (a *App) GetTrendAnalysis(startDate, endDate time.Time, periodType string, formats []string) (*storage.TrendAnalysis, error) {
	return a.service.GetTrendAnalysisWithFormats(a.ctx, startDate, endDate, periodType, formats)
}

// GetStatsByDeck returns statistics grouped by deck
func (a *App) GetStatsByDeck(filter models.StatsFilter) (map[string]*storage.Statistics, error) {
	return a.service.GetStatsByDeck(a.ctx, filter)
}

// GetRankProgressionTimeline returns rank progression timeline
func (a *App) GetRankProgressionTimeline(format string, startDate, endDate *time.Time, periodType storage.TimelinePeriod) (*storage.RankTimeline, error) {
	return a.service.GetRankProgressionTimeline(a.ctx, format, startDate, endDate, periodType)
}

// GetRankProgression returns rank progression for a format
func (a *App) GetRankProgression(format string) (*models.RankProgression, error) {
	return a.service.GetRankProgression(a.ctx, format)
}

// GetStatsByFormat returns statistics grouped by format
func (a *App) GetStatsByFormat(filter models.StatsFilter) (map[string]*models.Statistics, error) {
	return a.service.GetStatsByFormat(a.ctx, filter)
}

// GetPerformanceMetrics returns performance metrics
func (a *App) GetPerformanceMetrics(filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	return a.service.GetPerformanceMetrics(a.ctx, filter)
}

package refresh

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/draftdata"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// RefreshScheduler manages automated data refresh operations.
type RefreshScheduler interface {
	// Start begins the scheduled refresh loop
	Start(ctx context.Context) error

	// GetNextRefresh returns the next scheduled refresh time
	GetNextRefresh() time.Time

	// GetStaleItems returns items due for refresh
	GetStaleItems(ctx context.Context) (*StalenessSummary, error)

	// ScheduleRefresh schedules specific items for refresh
	ScheduleRefresh(items []RefreshItem, priority RefreshPriority)

	// RunRefreshes executes queued refreshes
	RunRefreshes(ctx context.Context) error

	// Stop stops the scheduler
	Stop()
}

// SchedulerConfig configures the refresh scheduler.
type SchedulerConfig struct {
	Storage       *storage.Service
	DraftUpdater  *draftdata.Updater
	Logger        *slog.Logger
	RefreshHour   int           // Hour to run daily refresh (default: 2 AM)
	CheckInterval time.Duration // How often to check for scheduled tasks (default: 1 hour)
}

// scheduler implements RefreshScheduler.
type scheduler struct {
	config  SchedulerConfig
	tracker *StalenessTracker
	logger  *slog.Logger

	// Scheduled refreshes
	refreshQueue chan refreshTask
	ticker       *time.Ticker
	wg           sync.WaitGroup
	stopCh       chan struct{}
}

// refreshTask represents a queued refresh operation.
type refreshTask struct {
	Items    []RefreshItem
	Priority RefreshPriority
	QueuedAt time.Time
}

// NewScheduler creates a new refresh scheduler.
func NewScheduler(config SchedulerConfig) (RefreshScheduler, error) {
	if config.Storage == nil {
		return nil, fmt.Errorf("storage is required")
	}
	if config.DraftUpdater == nil {
		return nil, fmt.Errorf("draftUpdater is required")
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.RefreshHour == 0 {
		config.RefreshHour = 2 // 2 AM default
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 1 * time.Hour
	}

	s := &scheduler{
		config:       config,
		tracker:      NewStalenessTracker(config.Storage),
		logger:       config.Logger,
		refreshQueue: make(chan refreshTask, 100),
		stopCh:       make(chan struct{}),
	}

	return s, nil
}

// Start begins the scheduled refresh loop.
func (s *scheduler) Start(ctx context.Context) error {
	s.ticker = time.NewTicker(s.config.CheckInterval)

	// Start refresh worker
	s.wg.Add(1)
	go s.refreshWorker(ctx)

	// Start scheduler loop
	s.wg.Add(1)
	go s.schedulerLoop(ctx)

	s.logger.Info("Refresh scheduler started",
		"checkInterval", s.config.CheckInterval,
		"refreshHour", s.config.RefreshHour)

	return nil
}

// schedulerLoop checks for scheduled tasks.
func (s *scheduler) schedulerLoop(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-s.ticker.C:
			now := time.Now()

			// Daily refresh at configured hour
			if now.Hour() == s.config.RefreshHour {
				s.logger.Info("Running daily scheduled refresh")
				if err := s.runDailyRefresh(ctx); err != nil {
					s.logger.Error("Daily refresh failed", "error", err)
				}
			}

			// Weekly refresh on Sunday
			if now.Weekday() == time.Sunday && now.Hour() == s.config.RefreshHour {
				s.logger.Info("Running weekly scheduled refresh")
				if err := s.runWeeklyRefresh(ctx); err != nil {
					s.logger.Error("Weekly refresh failed", "error", err)
				}
			}
		}
	}
}

// refreshWorker processes refresh tasks from the queue.
func (s *scheduler) refreshWorker(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case task := <-s.refreshQueue:
			s.logger.Debug("Processing refresh task",
				"itemCount", len(task.Items),
				"priority", task.Priority)

			if err := s.processRefreshTask(ctx, task); err != nil {
				s.logger.Error("Refresh task failed", "error", err)
			}
		}
	}
}

// GetNextRefresh returns the next scheduled refresh time.
func (s *scheduler) GetNextRefresh() time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), s.config.RefreshHour, 0, 0, 0, now.Location())

	if now.Hour() >= s.config.RefreshHour {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// GetStaleItems returns items due for refresh.
func (s *scheduler) GetStaleItems(ctx context.Context) (*StalenessSummary, error) {
	return s.tracker.GetSummary(ctx)
}

// ScheduleRefresh queues specific items for refresh.
func (s *scheduler) ScheduleRefresh(items []RefreshItem, priority RefreshPriority) {
	if len(items) == 0 {
		return
	}

	// Sort by priority
	sorted := s.prioritizeRefreshes(items)

	task := refreshTask{
		Items:    sorted,
		Priority: priority,
		QueuedAt: time.Now(),
	}

	select {
	case s.refreshQueue <- task:
		s.logger.Debug("Refresh task queued",
			"itemCount", len(items),
			"priority", priority)
	default:
		s.logger.Warn("Refresh queue full, task dropped",
			"itemCount", len(items))
	}
}

// RunRefreshes executes queued refreshes immediately.
func (s *scheduler) RunRefreshes(ctx context.Context) error {
	s.logger.Info("Running on-demand refreshes")

	// Get all stale items
	summary, err := s.tracker.GetSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stale items: %w", err)
	}

	if summary.RefreshesNeeded == 0 {
		s.logger.Info("No stale items to refresh")
		return nil
	}

	// Get stale cards
	staleCards, err := s.tracker.GetStaleCards(ctx, 1000)
	if err != nil {
		return fmt.Errorf("failed to get stale cards: %w", err)
	}

	// Get stale stats
	staleStats, err := s.tracker.GetStaleStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stale stats: %w", err)
	}

	// Schedule refreshes
	if len(staleCards) > 0 {
		s.ScheduleRefresh(staleCards, PriorityMedium)
	}
	if len(staleStats) > 0 {
		s.ScheduleRefresh(staleStats, PriorityHigh)
	}

	s.logger.Info("Refreshes scheduled",
		"cards", len(staleCards),
		"stats", len(staleStats))

	return nil
}

// Stop stops the scheduler.
func (s *scheduler) Stop() {
	close(s.stopCh)
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.refreshQueue)
	s.wg.Wait()

	s.logger.Info("Refresh scheduler stopped")
}

// runDailyRefresh runs the daily scheduled refresh (active sets).
func (s *scheduler) runDailyRefresh(ctx context.Context) error {
	s.logger.Info("Starting daily refresh: active sets")

	// Get active sets with stale stats
	staleStats, err := s.tracker.GetStaleStats(ctx)
	if err != nil {
		return err
	}

	// Filter for active sets only (you'd need set metadata to determine this)
	// For now, refresh all stale stats
	if len(staleStats) > 0 {
		s.ScheduleRefresh(staleStats, PriorityHigh)
	}

	return nil
}

// runWeeklyRefresh runs the weekly scheduled refresh (rotated sets).
func (s *scheduler) runWeeklyRefresh(ctx context.Context) error {
	s.logger.Info("Starting weekly refresh: rotated sets")

	// Get all stale metadata
	staleCards, err := s.tracker.GetStaleCards(ctx, 10000)
	if err != nil {
		return err
	}

	if len(staleCards) > 0 {
		s.ScheduleRefresh(staleCards, PriorityMedium)
	}

	return nil
}

// processRefreshTask processes a single refresh task.
func (s *scheduler) processRefreshTask(ctx context.Context, task refreshTask) error {
	s.logger.Info("Processing refresh task",
		"itemCount", len(task.Items),
		"priority", task.Priority,
		"queueAge", time.Since(task.QueuedAt))

	for _, item := range task.Items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.refreshItem(ctx, item); err != nil {
			s.logger.Warn("Failed to refresh item",
				"type", item.Type,
				"arenaID", item.ArenaID,
				"setCode", item.SetCode,
				"error", err)
			continue
		}

		// Rate limit between refreshes
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// refreshItem refreshes a single item.
func (s *scheduler) refreshItem(ctx context.Context, item RefreshItem) error {
	switch item.Type {
	case DataTypeMetadata:
		// Refresh Scryfall metadata
		// TODO: Implement card metadata refresh
		s.logger.Debug("Refreshing metadata", "arenaID", item.ArenaID)
		return nil

	case DataTypeStatistics:
		// Refresh 17Lands statistics
		s.logger.Debug("Refreshing statistics", "set", item.SetCode, "format", item.Format)
		_, err := s.config.DraftUpdater.UpdateSet(ctx, item.SetCode)
		return err

	default:
		return fmt.Errorf("unsupported data type: %v", item.Type)
	}
}

// prioritizeRefreshes sorts refresh items by priority.
func (s *scheduler) prioritizeRefreshes(items []RefreshItem) []RefreshItem {
	sorted := make([]RefreshItem, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		// 1. Active sets first
		if sorted[i].IsActive != sorted[j].IsActive {
			return sorted[i].IsActive
		}

		// 2. More stale first
		if sorted[i].StaleDays != sorted[j].StaleDays {
			return sorted[i].StaleDays > sorted[j].StaleDays
		}

		// 3. Higher access count first
		return sorted[i].AccessCount > sorted[j].AccessCount
	})

	return sorted
}

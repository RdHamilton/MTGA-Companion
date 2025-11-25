package gui

import (
	"context"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MatchFacade handles all match and statistics-related operations.
type MatchFacade struct {
	services *Services
}

// NewMatchFacade creates a new MatchFacade with the given services.
func NewMatchFacade(services *Services) *MatchFacade {
	return &MatchFacade{
		services: services,
	}
}

// GetMatches returns matches based on the provided filter.
func (m *MatchFacade) GetMatches(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetMatches(ctx, filter)
}

// GetMatchGames returns all games for a specific match.
func (m *MatchFacade) GetMatchGames(ctx context.Context, matchID string) ([]*models.Game, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetGamesForMatch(ctx, matchID)
}

// GetStats returns statistics based on the provided filter.
func (m *MatchFacade) GetStats(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetStats(ctx, filter)
}

// GetTrendAnalysis returns trend analysis for the specified time period.
func (m *MatchFacade) GetTrendAnalysis(ctx context.Context, startDate, endDate time.Time, periodType string, formats []string) (*storage.TrendAnalysis, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetTrendAnalysisWithFormats(ctx, startDate, endDate, periodType, formats)
}

// GetStatsByDeck returns statistics grouped by deck.
func (m *MatchFacade) GetStatsByDeck(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	log.Printf("GetStatsByDeck called with filter: %+v", filter)
	result, err := m.services.Storage.GetStatsByDeck(ctx, filter)
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

// GetRankProgressionTimeline returns rank progression timeline for a format.
func (m *MatchFacade) GetRankProgressionTimeline(ctx context.Context, format string, startDate, endDate *time.Time, periodType storage.TimelinePeriod) (*storage.RankTimeline, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetRankProgressionTimeline(ctx, format, startDate, endDate, periodType)
}

// GetRankProgression returns rank progression for a specific format.
func (m *MatchFacade) GetRankProgression(ctx context.Context, format string) (*models.RankProgression, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetRankProgression(ctx, format)
}

// GetStatsByFormat returns statistics grouped by format.
func (m *MatchFacade) GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetStatsByFormat(ctx, filter)
}

// GetPerformanceMetrics returns performance metrics based on the filter.
func (m *MatchFacade) GetPerformanceMetrics(ctx context.Context, filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}
	return m.services.Storage.GetPerformanceMetrics(ctx, filter)
}

// GetActiveQuests returns all active quests.
func (m *MatchFacade) GetActiveQuests(ctx context.Context) ([]*models.Quest, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}

	quests, err := m.services.Storage.Quests().GetActiveQuests()
	if err != nil {
		log.Printf("Error fetching active quests: %v", err)
		return nil, err
	}

	log.Printf("Found %d active quests", len(quests))
	return quests, nil
}

// GetQuestHistory returns quest history for the specified date range.
func (m *MatchFacade) GetQuestHistory(ctx context.Context, startDate, endDate string, limit int) ([]*models.Quest, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}

	// Parse date strings to time.Time pointers
	var start, end *time.Time
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			start = &t
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			end = &t
		}
	}

	quests, err := m.services.Storage.Quests().GetQuestHistory(start, end, limit)
	if err != nil {
		log.Printf("Error fetching quest history: %v", err)
		return nil, err
	}

	log.Printf("Found %d quests in history (start=%s, end=%s, limit=%d)",
		len(quests), startDate, endDate, limit)
	return quests, nil
}

// GetQuestStats returns quest statistics for the specified date range.
func (m *MatchFacade) GetQuestStats(ctx context.Context, startDate, endDate string) (*models.QuestStats, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}

	// Parse date strings to time.Time pointers
	var start, end *time.Time
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			start = &t
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			end = &t
		}
	}

	stats, err := m.services.Storage.Quests().GetQuestStats(start, end)
	if err != nil {
		log.Printf("Error fetching quest stats: %v", err)
		return nil, err
	}

	log.Printf("Quest stats (start=%s, end=%s): total=%d, completed=%d, gold=%d",
		startDate, endDate, stats.TotalQuests, stats.CompletedQuests, stats.TotalGoldEarned)
	return stats, nil
}

// GetCurrentAccount returns the current account information.
func (m *MatchFacade) GetCurrentAccount(ctx context.Context) (*models.Account, error) {
	if m.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized. Please configure database path in Settings."}
	}

	account, err := m.services.Storage.GetCurrentAccount(ctx)
	if err != nil {
		log.Printf("Error fetching current account: %v", err)
		return nil, err
	}

	if account != nil {
		screenName := ""
		if account.ScreenName != nil {
			screenName = *account.ScreenName
		}
		log.Printf("Current account: %s (ID: %d)", screenName, account.ID)
	}
	return account, nil
}

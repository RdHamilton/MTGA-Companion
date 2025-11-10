package storage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// HourStats represents statistics for a specific hour of day.
type HourStats struct {
	Hour         int     `json:"hour"` // 0-23
	TotalMatches int     `json:"total_matches"`
	MatchesWon   int     `json:"matches_won"`
	MatchesLost  int     `json:"matches_lost"`
	TotalGames   int     `json:"total_games"`
	GamesWon     int     `json:"games_won"`
	GamesLost    int     `json:"games_lost"`
	WinRate      float64 `json:"win_rate"`      // Match win rate percentage
	GameWinRate  float64 `json:"game_win_rate"` // Game win rate percentage
}

// DayOfWeekStats represents statistics for a specific day of week.
type DayOfWeekStats struct {
	DayOfWeek    string  `json:"day_of_week"` // Monday, Tuesday, etc.
	DayNumber    int     `json:"day_number"`  // 0=Sunday, 1=Monday, ..., 6=Saturday
	TotalMatches int     `json:"total_matches"`
	MatchesWon   int     `json:"matches_won"`
	MatchesLost  int     `json:"matches_lost"`
	TotalGames   int     `json:"total_games"`
	GamesWon     int     `json:"games_won"`
	GamesLost    int     `json:"games_lost"`
	WinRate      float64 `json:"win_rate"`      // Match win rate percentage
	GameWinRate  float64 `json:"game_win_rate"` // Game win rate percentage
}

// TimePatternSummary provides insights about time-based performance patterns.
type TimePatternSummary struct {
	BestHour         int     `json:"best_hour"` // Hour with highest win rate
	BestHourWinRate  float64 `json:"best_hour_win_rate"`
	WorstHour        int     `json:"worst_hour"` // Hour with lowest win rate
	WorstHourWinRate float64 `json:"worst_hour_win_rate"`
	BestDay          string  `json:"best_day"` // Day with highest win rate
	BestDayWinRate   float64 `json:"best_day_win_rate"`
	WorstDay         string  `json:"worst_day"` // Day with lowest win rate
	WorstDayWinRate  float64 `json:"worst_day_win_rate"`
	MostActiveHour   int     `json:"most_active_hour"` // Hour with most matches
	MostActiveDay    string  `json:"most_active_day"`  // Day with most matches
}

// GetHourOfDayStats calculates statistics grouped by hour of day (0-23).
func (s *Service) GetHourOfDayStats(ctx context.Context, filter models.StatsFilter) ([]HourStats, error) {
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return []HourStats{}, nil
	}

	// Initialize stats for each hour
	hourData := make(map[int]*HourStats)
	for i := 0; i < 24; i++ {
		hourData[i] = &HourStats{Hour: i}
	}

	// Aggregate matches by hour
	for _, match := range matches {
		hour := match.Timestamp.Hour()
		stats := hourData[hour]

		stats.TotalMatches++
		stats.TotalGames += match.PlayerWins + match.OpponentWins
		stats.GamesWon += match.PlayerWins
		stats.GamesLost += match.OpponentWins

		if match.Result == "win" {
			stats.MatchesWon++
		} else {
			stats.MatchesLost++
		}
	}

	// Calculate win rates and convert to slice
	result := make([]HourStats, 0, 24)
	for i := 0; i < 24; i++ {
		stats := hourData[i]
		if stats.TotalMatches > 0 {
			stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches) * 100
		}
		if stats.TotalGames > 0 {
			stats.GameWinRate = float64(stats.GamesWon) / float64(stats.TotalGames) * 100
		}
		result = append(result, *stats)
	}

	return result, nil
}

// GetDayOfWeekStats calculates statistics grouped by day of week.
func (s *Service) GetDayOfWeekStats(ctx context.Context, filter models.StatsFilter) ([]DayOfWeekStats, error) {
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return []DayOfWeekStats{}, nil
	}

	// Initialize stats for each day
	dayData := make(map[time.Weekday]*DayOfWeekStats)
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	for i := 0; i < 7; i++ {
		dayData[time.Weekday(i)] = &DayOfWeekStats{
			DayOfWeek: dayNames[i],
			DayNumber: i,
		}
	}

	// Aggregate matches by day of week
	for _, match := range matches {
		day := match.Timestamp.Weekday()
		stats := dayData[day]

		stats.TotalMatches++
		stats.TotalGames += match.PlayerWins + match.OpponentWins
		stats.GamesWon += match.PlayerWins
		stats.GamesLost += match.OpponentWins

		if match.Result == "win" {
			stats.MatchesWon++
		} else {
			stats.MatchesLost++
		}
	}

	// Calculate win rates and convert to slice
	result := make([]DayOfWeekStats, 0, 7)
	for i := 0; i < 7; i++ {
		stats := dayData[time.Weekday(i)]
		if stats.TotalMatches > 0 {
			stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches) * 100
		}
		if stats.TotalGames > 0 {
			stats.GameWinRate = float64(stats.GamesWon) / float64(stats.TotalGames) * 100
		}
		result = append(result, *stats)
	}

	return result, nil
}

// GetTimePatternSummary analyzes time-based patterns and identifies peak performance times.
func (s *Service) GetTimePatternSummary(ctx context.Context, filter models.StatsFilter) (*TimePatternSummary, error) {
	hourStats, err := s.GetHourOfDayStats(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get hour stats: %w", err)
	}

	dayStats, err := s.GetDayOfWeekStats(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get day stats: %w", err)
	}

	summary := &TimePatternSummary{}

	// Find best/worst/most active hours (only consider hours with matches)
	var activeHours []HourStats
	for _, h := range hourStats {
		if h.TotalMatches > 0 {
			activeHours = append(activeHours, h)
		}
	}

	if len(activeHours) > 0 {
		// Sort by win rate for best/worst
		sort.Slice(activeHours, func(i, j int) bool {
			return activeHours[i].WinRate > activeHours[j].WinRate
		})
		summary.BestHour = activeHours[0].Hour
		summary.BestHourWinRate = activeHours[0].WinRate
		summary.WorstHour = activeHours[len(activeHours)-1].Hour
		summary.WorstHourWinRate = activeHours[len(activeHours)-1].WinRate

		// Sort by total matches for most active
		sort.Slice(activeHours, func(i, j int) bool {
			return activeHours[i].TotalMatches > activeHours[j].TotalMatches
		})
		summary.MostActiveHour = activeHours[0].Hour
	}

	// Find best/worst/most active days (only consider days with matches)
	var activeDays []DayOfWeekStats
	for _, d := range dayStats {
		if d.TotalMatches > 0 {
			activeDays = append(activeDays, d)
		}
	}

	if len(activeDays) > 0 {
		// Sort by win rate for best/worst
		sort.Slice(activeDays, func(i, j int) bool {
			return activeDays[i].WinRate > activeDays[j].WinRate
		})
		summary.BestDay = activeDays[0].DayOfWeek
		summary.BestDayWinRate = activeDays[0].WinRate
		summary.WorstDay = activeDays[len(activeDays)-1].DayOfWeek
		summary.WorstDayWinRate = activeDays[len(activeDays)-1].WinRate

		// Sort by total matches for most active
		sort.Slice(activeDays, func(i, j int) bool {
			return activeDays[i].TotalMatches > activeDays[j].TotalMatches
		})
		summary.MostActiveDay = activeDays[0].DayOfWeek
	}

	return summary, nil
}

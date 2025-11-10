package export

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// HourStatsExportRow represents hour-of-day statistics for CSV export.
type HourStatsExportRow struct {
	Hour         int     `csv:"hour" json:"hour"`
	TotalMatches int     `csv:"total_matches" json:"total_matches"`
	MatchesWon   int     `csv:"matches_won" json:"matches_won"`
	MatchesLost  int     `csv:"matches_lost" json:"matches_lost"`
	TotalGames   int     `csv:"total_games" json:"total_games"`
	GamesWon     int     `csv:"games_won" json:"games_won"`
	GamesLost    int     `csv:"games_lost" json:"games_lost"`
	WinRate      float64 `csv:"win_rate" json:"win_rate"`
	GameWinRate  float64 `csv:"game_win_rate" json:"game_win_rate"`
}

// DayOfWeekStatsExportRow represents day-of-week statistics for CSV export.
type DayOfWeekStatsExportRow struct {
	DayOfWeek    string  `csv:"day_of_week" json:"day_of_week"`
	DayNumber    int     `csv:"day_number" json:"day_number"`
	TotalMatches int     `csv:"total_matches" json:"total_matches"`
	MatchesWon   int     `csv:"matches_won" json:"matches_won"`
	MatchesLost  int     `csv:"matches_lost" json:"matches_lost"`
	TotalGames   int     `csv:"total_games" json:"total_games"`
	GamesWon     int     `csv:"games_won" json:"games_won"`
	GamesLost    int     `csv:"games_lost" json:"games_lost"`
	WinRate      float64 `csv:"win_rate" json:"win_rate"`
	GameWinRate  float64 `csv:"game_win_rate" json:"game_win_rate"`
}

// TimePatternSummaryExportRow represents time pattern summary for CSV export.
type TimePatternSummaryExportRow struct {
	BestHour         int     `csv:"best_hour" json:"best_hour"`
	BestHourWinRate  float64 `csv:"best_hour_win_rate" json:"best_hour_win_rate"`
	WorstHour        int     `csv:"worst_hour" json:"worst_hour"`
	WorstHourWinRate float64 `csv:"worst_hour_win_rate" json:"worst_hour_win_rate"`
	BestDay          string  `csv:"best_day" json:"best_day"`
	BestDayWinRate   float64 `csv:"best_day_win_rate" json:"best_day_win_rate"`
	WorstDay         string  `csv:"worst_day" json:"worst_day"`
	WorstDayWinRate  float64 `csv:"worst_day_win_rate" json:"worst_day_win_rate"`
	MostActiveHour   int     `csv:"most_active_hour" json:"most_active_hour"`
	MostActiveDay    string  `csv:"most_active_day" json:"most_active_day"`
}

// ExportHourOfDayStats exports hour-of-day statistics.
func ExportHourOfDayStats(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	stats, err := service.GetHourOfDayStats(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get hour-of-day stats: %w", err)
	}

	if len(stats) == 0 {
		return fmt.Errorf("no hour-of-day statistics found")
	}

	// Convert to export rows
	rows := make([]HourStatsExportRow, len(stats))
	for i, s := range stats {
		rows[i] = HourStatsExportRow{
			Hour:         s.Hour,
			TotalMatches: s.TotalMatches,
			MatchesWon:   s.MatchesWon,
			MatchesLost:  s.MatchesLost,
			TotalGames:   s.TotalGames,
			GamesWon:     s.GamesWon,
			GamesLost:    s.GamesLost,
			WinRate:      s.WinRate,
			GameWinRate:  s.GameWinRate,
		}
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// ExportDayOfWeekStats exports day-of-week statistics.
func ExportDayOfWeekStats(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	stats, err := service.GetDayOfWeekStats(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get day-of-week stats: %w", err)
	}

	if len(stats) == 0 {
		return fmt.Errorf("no day-of-week statistics found")
	}

	// Convert to export rows
	rows := make([]DayOfWeekStatsExportRow, len(stats))
	for i, s := range stats {
		rows[i] = DayOfWeekStatsExportRow{
			DayOfWeek:    s.DayOfWeek,
			DayNumber:    s.DayNumber,
			TotalMatches: s.TotalMatches,
			MatchesWon:   s.MatchesWon,
			MatchesLost:  s.MatchesLost,
			TotalGames:   s.TotalGames,
			GamesWon:     s.GamesWon,
			GamesLost:    s.GamesLost,
			WinRate:      s.WinRate,
			GameWinRate:  s.GameWinRate,
		}
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// ExportTimePatternSummary exports time pattern summary.
func ExportTimePatternSummary(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	summary, err := service.GetTimePatternSummary(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get time pattern summary: %w", err)
	}

	if summary == nil {
		return fmt.Errorf("no time pattern summary found")
	}

	// Convert to export row
	row := TimePatternSummaryExportRow{
		BestHour:         summary.BestHour,
		BestHourWinRate:  summary.BestHourWinRate,
		WorstHour:        summary.WorstHour,
		WorstHourWinRate: summary.WorstHourWinRate,
		BestDay:          summary.BestDay,
		BestDayWinRate:   summary.BestDayWinRate,
		WorstDay:         summary.WorstDay,
		WorstDayWinRate:  summary.WorstDayWinRate,
		MostActiveHour:   summary.MostActiveHour,
		MostActiveDay:    summary.MostActiveDay,
	}

	switch opts.Format {
	case FormatCSV:
		// Export as single-row CSV
		exporter := NewExporter(opts)
		return exporter.Export([]TimePatternSummaryExportRow{row})
	case FormatJSON:
		// Export summary directly in JSON
		exporter := NewExporter(opts)
		return exporter.Export(summary)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

package export

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MatchExportRow represents a match for CSV export.
type MatchExportRow struct {
	MatchID         string `csv:"match_id" json:"match_id"`
	EventID         string `csv:"event_id" json:"event_id"`
	EventName       string `csv:"event_name" json:"event_name"`
	Timestamp       string `csv:"timestamp" json:"timestamp"`
	Format          string `csv:"format" json:"format"`
	Result          string `csv:"result" json:"result"`
	ResultReason    string `csv:"result_reason" json:"result_reason"`
	PlayerWins      int    `csv:"player_wins" json:"player_wins"`
	OpponentWins    int    `csv:"opponent_wins" json:"opponent_wins"`
	DurationSeconds *int   `csv:"duration_seconds" json:"duration_seconds,omitempty"`
	DeckID          string `csv:"deck_id" json:"deck_id"`
	RankBefore      string `csv:"rank_before" json:"rank_before"`
	RankAfter       string `csv:"rank_after" json:"rank_after"`
	OpponentName    string `csv:"opponent_name" json:"opponent_name"`
	OpponentID      string `csv:"opponent_id" json:"opponent_id"`
}

// StatisticsExportRow represents aggregated statistics for CSV export.
type StatisticsExportRow struct {
	Format       string  `csv:"format" json:"format"`
	TotalMatches int     `csv:"total_matches" json:"total_matches"`
	MatchesWon   int     `csv:"matches_won" json:"matches_won"`
	MatchesLost  int     `csv:"matches_lost" json:"matches_lost"`
	TotalGames   int     `csv:"total_games" json:"total_games"`
	GamesWon     int     `csv:"games_won" json:"games_won"`
	GamesLost    int     `csv:"games_lost" json:"games_lost"`
	WinRate      float64 `csv:"win_rate" json:"win_rate"`
	GameWinRate  float64 `csv:"game_win_rate" json:"game_win_rate"`
}

// DailyStatsExportRow represents daily statistics for CSV export.
type DailyStatsExportRow struct {
	Date          string `csv:"date" json:"date"`
	Format        string `csv:"format" json:"format"`
	MatchesPlayed int    `csv:"matches_played" json:"matches_played"`
	MatchesWon    int    `csv:"matches_won" json:"matches_won"`
	GamesPlayed   int    `csv:"games_played" json:"games_played"`
	GamesWon      int    `csv:"games_won" json:"games_won"`
	WinRate       string `csv:"win_rate" json:"win_rate"`
}

// ExportMatchHistory exports match history to the specified format.
func ExportMatchHistory(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for the specified filter")
	}

	// Convert matches to export rows
	rows := make([]MatchExportRow, len(matches))
	for i, match := range matches {
		rows[i] = MatchExportRow{
			MatchID:         match.ID,
			EventID:         match.EventID,
			EventName:       match.EventName,
			Timestamp:       match.Timestamp.Format(time.RFC3339),
			Format:          match.Format,
			Result:          match.Result,
			ResultReason:    valueOrEmpty(match.ResultReason),
			PlayerWins:      match.PlayerWins,
			OpponentWins:    match.OpponentWins,
			DurationSeconds: match.DurationSeconds,
			DeckID:          valueOrEmpty(match.DeckID),
			RankBefore:      valueOrEmpty(match.RankBefore),
			RankAfter:       valueOrEmpty(match.RankAfter),
			OpponentName:    valueOrEmpty(match.OpponentName),
			OpponentID:      valueOrEmpty(match.OpponentID),
		}
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// ExportAggregatedStatistics exports aggregated statistics by format.
func ExportAggregatedStatistics(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	statsByFormat, err := service.GetStatsByFormat(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	if len(statsByFormat) == 0 {
		return fmt.Errorf("no statistics found for the specified filter")
	}

	// Convert statistics to export rows
	rows := make([]StatisticsExportRow, 0, len(statsByFormat))
	for format, stats := range statsByFormat {
		rows = append(rows, StatisticsExportRow{
			Format:       format,
			TotalMatches: stats.TotalMatches,
			MatchesWon:   stats.MatchesWon,
			MatchesLost:  stats.MatchesLost,
			TotalGames:   stats.TotalGames,
			GamesWon:     stats.GamesWon,
			GamesLost:    stats.GamesLost,
			WinRate:      stats.WinRate,
			GameWinRate:  stats.GameWinRate,
		})
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// ExportDailyStatistics exports daily statistics.
func ExportDailyStatistics(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	if filter.StartDate == nil || filter.EndDate == nil {
		return fmt.Errorf("start_date and end_date are required for daily statistics export")
	}

	// Get all matches in the date range
	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for the specified date range")
	}

	// Aggregate matches by date and format
	type dailyKey struct {
		date   string
		format string
	}

	dailyStats := make(map[dailyKey]struct {
		matchesPlayed int
		matchesWon    int
		gamesPlayed   int
		gamesWon      int
	})

	for _, match := range matches {
		date := match.Timestamp.Format("2006-01-02")
		key := dailyKey{date: date, format: match.Format}

		stats := dailyStats[key]
		stats.matchesPlayed++
		stats.gamesPlayed += match.PlayerWins + match.OpponentWins
		stats.gamesWon += match.PlayerWins

		if match.Result == "win" {
			stats.matchesWon++
		}

		dailyStats[key] = stats
	}

	// Convert to export rows
	rows := make([]DailyStatsExportRow, 0, len(dailyStats))
	for key, stats := range dailyStats {
		winRate := fmt.Sprintf("%.1f%%", float64(stats.matchesWon)/float64(stats.matchesPlayed)*100)

		rows = append(rows, DailyStatsExportRow{
			Date:          key.date,
			Format:        key.format,
			MatchesPlayed: stats.matchesPlayed,
			MatchesWon:    stats.matchesWon,
			GamesPlayed:   stats.gamesPlayed,
			GamesWon:      stats.gamesWon,
			WinRate:       winRate,
		})
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// ExportPerformanceMetrics exports performance metrics (duration-based stats).
func ExportPerformanceMetrics(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	metrics, err := service.GetPerformanceMetrics(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get performance metrics: %w", err)
	}

	if metrics == nil {
		return fmt.Errorf("no performance metrics found")
	}

	// Export as single-row data
	type PerformanceExportRow struct {
		AvgMatchDuration float64 `csv:"avg_match_duration" json:"avg_match_duration"`
		AvgGameDuration  float64 `csv:"avg_game_duration" json:"avg_game_duration"`
		FastestMatch     int     `csv:"fastest_match" json:"fastest_match"`
		SlowestMatch     int     `csv:"slowest_match" json:"slowest_match"`
		FastestGame      int     `csv:"fastest_game" json:"fastest_game"`
		SlowestGame      int     `csv:"slowest_game" json:"slowest_game"`
	}

	row := PerformanceExportRow{
		AvgMatchDuration: valueOrZero(metrics.AvgMatchDuration),
		AvgGameDuration:  valueOrZero(metrics.AvgGameDuration),
		FastestMatch:     valueOrZero(metrics.FastestMatch),
		SlowestMatch:     valueOrZero(metrics.SlowestMatch),
		FastestGame:      valueOrZero(metrics.FastestGame),
		SlowestGame:      valueOrZero(metrics.SlowestGame),
	}

	exporter := NewExporter(opts)
	return exporter.Export([]PerformanceExportRow{row})
}

// ExportOptions holds configuration for statistics export.
type ExportOptions struct {
	Type       string // "matches", "stats", "daily", "performance"
	Format     Format
	OutputPath string
	Filter     models.StatsFilter
	Overwrite  bool
	PrettyJSON bool
}

// ExportStatistics is a convenience function that exports statistics based on type.
func ExportStatistics(ctx context.Context, service *storage.Service, exportOpts ExportOptions) error {
	// Generate filename if not provided
	if exportOpts.OutputPath == "" {
		exportOpts.OutputPath = filepath.Join("exports", GenerateFilename(exportOpts.Type, exportOpts.Format))
	}

	opts := Options{
		Format:     exportOpts.Format,
		FilePath:   exportOpts.OutputPath,
		Overwrite:  exportOpts.Overwrite,
		PrettyJSON: exportOpts.PrettyJSON,
	}

	switch exportOpts.Type {
	case "matches":
		return ExportMatchHistory(ctx, service, exportOpts.Filter, opts)
	case "stats":
		return ExportAggregatedStatistics(ctx, service, exportOpts.Filter, opts)
	case "daily":
		return ExportDailyStatistics(ctx, service, exportOpts.Filter, opts)
	case "performance":
		return ExportPerformanceMetrics(ctx, service, exportOpts.Filter, opts)
	default:
		return fmt.Errorf("unknown export type: %s", exportOpts.Type)
	}
}

// Helper functions

func valueOrEmpty(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func valueOrZero[T int | float64](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

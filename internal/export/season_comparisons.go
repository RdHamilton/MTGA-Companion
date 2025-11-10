package export

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// SeasonStatsExportRow represents season statistics for CSV export.
type SeasonStatsExportRow struct {
	SeasonName   string  `csv:"season_name" json:"season_name"`
	StartDate    string  `csv:"start_date" json:"start_date"`
	EndDate      string  `csv:"end_date" json:"end_date"`
	TotalMatches int     `csv:"total_matches" json:"total_matches"`
	MatchesWon   int     `csv:"matches_won" json:"matches_won"`
	MatchesLost  int     `csv:"matches_lost" json:"matches_lost"`
	TotalGames   int     `csv:"total_games" json:"total_games"`
	GamesWon     int     `csv:"games_won" json:"games_won"`
	GamesLost    int     `csv:"games_lost" json:"games_lost"`
	WinRate      float64 `csv:"win_rate" json:"win_rate"`
	GameWinRate  float64 `csv:"game_win_rate" json:"game_win_rate"`
}

// SeasonComparisonExportRow represents two-season comparison for CSV export.
type SeasonComparisonExportRow struct {
	Season1Name          string  `csv:"season1_name" json:"season1_name"`
	Season1WinRate       float64 `csv:"season1_win_rate" json:"season1_win_rate"`
	Season1Matches       int     `csv:"season1_matches" json:"season1_matches"`
	Season2Name          string  `csv:"season2_name" json:"season2_name"`
	Season2WinRate       float64 `csv:"season2_win_rate" json:"season2_win_rate"`
	Season2Matches       int     `csv:"season2_matches" json:"season2_matches"`
	WinRateChange        float64 `csv:"win_rate_change" json:"win_rate_change"`
	GameWinRateChange    float64 `csv:"game_win_rate_change" json:"game_win_rate_change"`
	MatchCountChange     int     `csv:"match_count_change" json:"match_count_change"`
	MatchCountChangePerc float64 `csv:"match_count_change_perc" json:"match_count_change_perc"`
	Trend                string  `csv:"trend" json:"trend"`
}

// MultiSeasonComparisonExportRow represents multi-season summary for CSV export.
type MultiSeasonComparisonExportRow struct {
	BestSeason   string `csv:"best_season" json:"best_season"`
	WorstSeason  string `csv:"worst_season" json:"worst_season"`
	MostActive   string `csv:"most_active" json:"most_active"`
	OverallTrend string `csv:"overall_trend" json:"overall_trend"`
}

// ExportSeasonStats exports statistics for a single season.
func ExportSeasonStats(ctx context.Context, service *storage.Service, seasonName string, startDate, endDate time.Time, filter models.StatsFilter, opts Options) error {
	stats, err := service.GetSeasonStats(ctx, seasonName, startDate, endDate, filter.Format)
	if err != nil {
		return fmt.Errorf("failed to get season stats: %w", err)
	}

	if stats == nil {
		return fmt.Errorf("no season statistics found")
	}

	// Convert to export row
	row := SeasonStatsExportRow{
		SeasonName:   stats.SeasonName,
		StartDate:    stats.StartDate.Format("2006-01-02"),
		EndDate:      stats.EndDate.Format("2006-01-02"),
		TotalMatches: stats.TotalMatches,
		MatchesWon:   stats.MatchesWon,
		MatchesLost:  stats.MatchesLost,
		TotalGames:   stats.TotalGames,
		GamesWon:     stats.GamesWon,
		GamesLost:    stats.GamesLost,
		WinRate:      stats.WinRate,
		GameWinRate:  stats.GameWinRate,
	}

	switch opts.Format {
	case FormatCSV:
		exporter := NewExporter(opts)
		return exporter.Export([]SeasonStatsExportRow{row})
	case FormatJSON:
		exporter := NewExporter(opts)
		return exporter.Export(stats)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportSeasonComparison exports comparison between two seasons/time periods.
// Auto-generates names based on date ranges.
func ExportSeasonComparison(ctx context.Context, service *storage.Service,
	period1Start, period1End time.Time,
	period2Start, period2End time.Time,
	formatFilter *string, opts Options,
) error {
	// Auto-generate period names based on dates
	period1Name := fmt.Sprintf("%s to %s", period1Start.Format("2006-01-02"), period1End.Format("2006-01-02"))
	period2Name := fmt.Sprintf("%s to %s", period2Start.Format("2006-01-02"), period2End.Format("2006-01-02"))

	comparison, err := service.CompareSeasons(ctx, period1Name, period1Start, period1End, period2Name, period2Start, period2End, formatFilter)
	if err != nil {
		return fmt.Errorf("failed to compare periods: %w", err)
	}

	if comparison == nil {
		return fmt.Errorf("no period comparison found")
	}

	// Convert to export row
	row := SeasonComparisonExportRow{
		Season1Name:          comparison.Season1.SeasonName,
		Season1WinRate:       comparison.Season1.WinRate,
		Season1Matches:       comparison.Season1.TotalMatches,
		Season2Name:          comparison.Season2.SeasonName,
		Season2WinRate:       comparison.Season2.WinRate,
		Season2Matches:       comparison.Season2.TotalMatches,
		WinRateChange:        comparison.WinRateChange,
		GameWinRateChange:    comparison.GameWinRateChange,
		MatchCountChange:     comparison.MatchCountChange,
		MatchCountChangePerc: comparison.MatchCountChangePerc,
		Trend:                comparison.Trend,
	}

	switch opts.Format {
	case FormatCSV:
		exporter := NewExporter(opts)
		return exporter.Export([]SeasonComparisonExportRow{row})
	case FormatJSON:
		exporter := NewExporter(opts)
		return exporter.Export(comparison)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportMultiSeasonComparison exports comparison of multiple seasons.
func ExportMultiSeasonComparison(ctx context.Context, service *storage.Service,
	seasons []struct {
		Name      string
		StartDate time.Time
		EndDate   time.Time
	}, formatFilter *string, opts Options,
) error {
	comparison, err := service.CompareMultipleSeasons(ctx, seasons, formatFilter)
	if err != nil {
		return fmt.Errorf("failed to compare multiple seasons: %w", err)
	}

	if comparison == nil {
		return fmt.Errorf("no multi-season comparison found")
	}

	switch opts.Format {
	case FormatCSV:
		// Export summary row
		summaryRow := MultiSeasonComparisonExportRow{
			BestSeason:   comparison.BestSeason,
			WorstSeason:  comparison.WorstSeason,
			MostActive:   comparison.MostActive,
			OverallTrend: comparison.OverallTrend,
		}

		exporter := NewExporter(opts)
		return exporter.Export([]MultiSeasonComparisonExportRow{summaryRow})
	case FormatJSON:
		// Export full comparison with all season details
		exporter := NewExporter(opts)
		return exporter.Export(comparison)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

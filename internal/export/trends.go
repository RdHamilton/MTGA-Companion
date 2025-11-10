package export

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// TrendPeriodExportRow represents a single period in trend analysis for CSV export.
type TrendPeriodExportRow struct {
	PeriodLabel     string  `csv:"period_label" json:"period_label"`
	StartDate       string  `csv:"start_date" json:"start_date"`
	EndDate         string  `csv:"end_date" json:"end_date"`
	TotalMatches    int     `csv:"total_matches" json:"total_matches"`
	MatchesWon      int     `csv:"matches_won" json:"matches_won"`
	MatchesLost     int     `csv:"matches_lost" json:"matches_lost"`
	TotalGames      int     `csv:"total_games" json:"total_games"`
	GamesWon        int     `csv:"games_won" json:"games_won"`
	GamesLost       int     `csv:"games_lost" json:"games_lost"`
	WinRate         float64 `csv:"win_rate" json:"win_rate"`
	GameWinRate     float64 `csv:"game_win_rate" json:"game_win_rate"`
	OverallTrend    string  `csv:"overall_trend" json:"overall_trend"`
	OverallTrendPct float64 `csv:"overall_trend_pct" json:"overall_trend_pct"`
}

// TrendAnalysisExportJSON represents the complete trend analysis in JSON format.
type TrendAnalysisExportJSON struct {
	Periods     []TrendPeriodJSON `json:"periods"`
	Overall     *StatisticsJSON   `json:"overall"`
	Trend       string            `json:"trend"`
	TrendValue  float64           `json:"trend_value"`
	ExportedAt  string            `json:"exported_at"`
	StartDate   string            `json:"start_date"`
	EndDate     string            `json:"end_date"`
	PeriodType  string            `json:"period_type"`
	PeriodCount int               `json:"period_count"`
}

// TrendPeriodJSON represents a single period in JSON format.
type TrendPeriodJSON struct {
	Label       string          `json:"label"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	Stats       *StatisticsJSON `json:"stats"`
	WinRate     float64         `json:"win_rate"`
	GameWinRate float64         `json:"game_win_rate"`
}

// StatisticsJSON represents statistics in JSON format.
type StatisticsJSON struct {
	TotalMatches int     `json:"total_matches"`
	MatchesWon   int     `json:"matches_won"`
	MatchesLost  int     `json:"matches_lost"`
	TotalGames   int     `json:"total_games"`
	GamesWon     int     `json:"games_won"`
	GamesLost    int     `json:"games_lost"`
	WinRate      float64 `json:"win_rate"`
	GameWinRate  float64 `json:"game_win_rate"`
}

// ExportTrendAnalysis exports trend analysis data to the specified format.
func ExportTrendAnalysis(ctx context.Context, service *storage.Service, startDate, endDate time.Time, periodType string, opts Options) error {
	// Get trend analysis
	analysis, err := service.GetTrendAnalysis(ctx, startDate, endDate, periodType)
	if err != nil {
		return fmt.Errorf("failed to get trend analysis: %w", err)
	}

	if analysis == nil || len(analysis.Periods) == 0 {
		return fmt.Errorf("no trend data found for the specified period")
	}

	switch opts.Format {
	case FormatCSV:
		return exportTrendAnalysisCSV(analysis, startDate, endDate, periodType, opts)
	case FormatJSON:
		return exportTrendAnalysisJSON(analysis, startDate, endDate, periodType, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// exportTrendAnalysisCSV exports trend analysis to CSV format (one row per period).
func exportTrendAnalysisCSV(analysis *storage.TrendAnalysis, startDate, endDate time.Time, periodType string, opts Options) error {
	rows := make([]TrendPeriodExportRow, len(analysis.Periods))

	for i, period := range analysis.Periods {
		row := TrendPeriodExportRow{
			PeriodLabel:     period.Period.Label,
			StartDate:       period.Period.StartDate.Format("2006-01-02"),
			EndDate:         period.Period.EndDate.Format("2006-01-02"),
			WinRate:         period.WinRate,
			GameWinRate:     period.GameWinRate,
			OverallTrend:    analysis.Trend,
			OverallTrendPct: analysis.TrendValue,
		}

		if period.Stats != nil {
			row.TotalMatches = period.Stats.TotalMatches
			row.MatchesWon = period.Stats.MatchesWon
			row.MatchesLost = period.Stats.MatchesLost
			row.TotalGames = period.Stats.TotalGames
			row.GamesWon = period.Stats.GamesWon
			row.GamesLost = period.Stats.GamesLost
		}

		rows[i] = row
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// exportTrendAnalysisJSON exports trend analysis to JSON format.
func exportTrendAnalysisJSON(analysis *storage.TrendAnalysis, startDate, endDate time.Time, periodType string, opts Options) error {
	jsonData := TrendAnalysisExportJSON{
		Trend:       analysis.Trend,
		TrendValue:  analysis.TrendValue,
		ExportedAt:  time.Now().Format(time.RFC3339),
		StartDate:   startDate.Format("2006-01-02"),
		EndDate:     endDate.Format("2006-01-02"),
		PeriodType:  periodType,
		PeriodCount: len(analysis.Periods),
	}

	// Overall statistics
	if analysis.Overall != nil {
		jsonData.Overall = &StatisticsJSON{
			TotalMatches: analysis.Overall.TotalMatches,
			MatchesWon:   analysis.Overall.MatchesWon,
			MatchesLost:  analysis.Overall.MatchesLost,
			TotalGames:   analysis.Overall.TotalGames,
			GamesWon:     analysis.Overall.GamesWon,
			GamesLost:    analysis.Overall.GamesLost,
			WinRate:      analysis.Overall.WinRate,
			GameWinRate:  analysis.Overall.GameWinRate,
		}
	}

	// Period data
	jsonData.Periods = make([]TrendPeriodJSON, len(analysis.Periods))
	for i, period := range analysis.Periods {
		periodJSON := TrendPeriodJSON{
			Label:       period.Period.Label,
			StartDate:   period.Period.StartDate.Format("2006-01-02"),
			EndDate:     period.Period.EndDate.Format("2006-01-02"),
			WinRate:     period.WinRate,
			GameWinRate: period.GameWinRate,
		}

		if period.Stats != nil {
			periodJSON.Stats = &StatisticsJSON{
				TotalMatches: period.Stats.TotalMatches,
				MatchesWon:   period.Stats.MatchesWon,
				MatchesLost:  period.Stats.MatchesLost,
				TotalGames:   period.Stats.TotalGames,
				GamesWon:     period.Stats.GamesWon,
				GamesLost:    period.Stats.GamesLost,
				WinRate:      period.Stats.WinRate,
				GameWinRate:  period.Stats.GameWinRate,
			}
		}

		jsonData.Periods[i] = periodJSON
	}

	exporter := NewExporter(opts)
	return exporter.Export(jsonData)
}

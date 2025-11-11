package export

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// ComparisonExportRow represents a row in the comparison export.
type ComparisonExportRow struct {
	Group        string  `csv:"Group" json:"group"`
	TotalMatches int     `csv:"Total Matches" json:"total_matches"`
	MatchesWon   int     `csv:"Matches Won" json:"matches_won"`
	MatchesLost  int     `csv:"Matches Lost" json:"matches_lost"`
	WinRate      float64 `csv:"Win Rate" json:"win_rate"`
	TotalGames   int     `csv:"Total Games" json:"total_games"`
	GamesWon     int     `csv:"Games Won" json:"games_won"`
	GamesLost    int     `csv:"Games Lost" json:"games_lost"`
	GameWinRate  float64 `csv:"Game Win Rate" json:"game_win_rate"`
}

// ComparisonSummary provides overall comparison insights.
type ComparisonSummary struct {
	BestGroup      string  `json:"best_group"`
	BestWinRate    float64 `json:"best_win_rate"`
	WorstGroup     string  `json:"worst_group"`
	WorstWinRate   float64 `json:"worst_win_rate"`
	WinRateDiff    float64 `json:"win_rate_difference"`
	TotalMatches   int     `json:"total_matches"`
	GroupsCompared int     `json:"groups_compared"`
}

// ComparisonExport combines comparison data with summary.
type ComparisonExport struct {
	Summary ComparisonSummary      `json:"summary"`
	Groups  []*ComparisonExportRow `json:"groups"`
}

// ExportMatchComparison exports match comparison results.
func ExportMatchComparison(ctx context.Context, service *storage.Service, result *storage.ComparisonResult, opts Options) error {
	// Convert comparison result to export rows
	rows := make([]*ComparisonExportRow, 0, len(result.Groups))

	for _, group := range result.Groups {
		rows = append(rows, &ComparisonExportRow{
			Group:        group.Label,
			TotalMatches: group.Statistics.TotalMatches,
			MatchesWon:   group.Statistics.MatchesWon,
			MatchesLost:  group.Statistics.MatchesLost,
			WinRate:      group.Statistics.WinRate * 100, // Convert to percentage
			TotalGames:   group.Statistics.TotalGames,
			GamesWon:     group.Statistics.GamesWon,
			GamesLost:    group.Statistics.GamesLost,
			GameWinRate:  group.Statistics.GameWinRate * 100,
		})
	}

	// Create summary
	summary := ComparisonSummary{
		TotalMatches:   result.TotalMatches,
		GroupsCompared: len(result.Groups),
		WinRateDiff:    result.WinRateDiff * 100,
	}

	if result.BestGroup != nil {
		summary.BestGroup = result.BestGroup.Label
		summary.BestWinRate = result.BestGroup.Statistics.WinRate * 100
	}

	if result.WorstGroup != nil {
		summary.WorstGroup = result.WorstGroup.Label
		summary.WorstWinRate = result.WorstGroup.Statistics.WinRate * 100
	}

	// Export based on format
	exporter := NewExporter(opts)

	switch opts.Format {
	case FormatJSON:
		exportData := ComparisonExport{
			Summary: summary,
			Groups:  rows,
		}
		return exporter.Export(exportData)
	case FormatCSV:
		return exporter.Export(rows)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// ExportFormatComparison exports a comparison of different formats.
func ExportFormatComparison(ctx context.Context, service *storage.Service, formats []string, baseFilter storage.ComparisonGroup, opts Options) error {
	result, err := service.CompareFormats(ctx, formats, baseFilter.Filter)
	if err != nil {
		return fmt.Errorf("failed to compare formats: %w", err)
	}

	return ExportMatchComparison(ctx, service, result, opts)
}

// ExportTimePeriodComparison exports a comparison of different time periods.
func ExportTimePeriodComparison(ctx context.Context, service *storage.Service, periods []struct {
	Label string
	Start time.Time
	End   time.Time
}, baseFilter storage.ComparisonGroup, opts Options,
) error {
	result, err := service.CompareTimePeriods(ctx, periods, baseFilter.Filter)
	if err != nil {
		return fmt.Errorf("failed to compare time periods: %w", err)
	}

	return ExportMatchComparison(ctx, service, result, opts)
}

// ExportDeckComparison exports a comparison of different decks.
func ExportDeckComparison(ctx context.Context, service *storage.Service, deckIDs []string, baseFilter storage.ComparisonGroup, opts Options) error {
	result, err := service.CompareDecks(ctx, deckIDs, baseFilter.Filter)
	if err != nil {
		return fmt.Errorf("failed to compare decks: %w", err)
	}

	return ExportMatchComparison(ctx, service, result, opts)
}

package export

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// StreakDataExportRow represents streak data for CSV export.
type StreakDataExportRow struct {
	CurrentStreak      int    `csv:"current_streak" json:"current_streak"`
	CurrentStreakType  string `csv:"current_streak_type" json:"current_streak_type"`
	LongestWinStreak   int    `csv:"longest_win_streak" json:"longest_win_streak"`
	LongestLossStreak  int    `csv:"longest_loss_streak" json:"longest_loss_streak"`
	CurrentWinStreak   int    `csv:"current_win_streak" json:"current_win_streak"`
	CurrentLossStreak  int    `csv:"current_loss_streak" json:"current_loss_streak"`
	TotalStreaksOver5  int    `csv:"total_streaks_over_5" json:"total_streaks_over_5"`
	TotalStreaksOver10 int    `csv:"total_streaks_over_10" json:"total_streaks_over_10"`
	LastMatchResult    string `csv:"last_match_result" json:"last_match_result"`
	LastMatchTimestamp string `csv:"last_match_timestamp" json:"last_match_timestamp"`
}

// StreakHistoryExportRow represents a historical streak for CSV export.
type StreakHistoryExportRow struct {
	StreakType string `csv:"streak_type" json:"streak_type"`
	Length     int    `csv:"length" json:"length"`
	StartDate  string `csv:"start_date" json:"start_date"`
	EndDate    string `csv:"end_date" json:"end_date"`
	Format     string `csv:"format" json:"format"`
	EventName  string `csv:"event_name" json:"event_name"`
}

// ExportStreakData exports current streak statistics.
func ExportStreakData(ctx context.Context, service *storage.Service, filter models.StatsFilter, opts Options) error {
	data, err := service.GetStreakData(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get streak data: %w", err)
	}

	if data == nil {
		return fmt.Errorf("no streak data found")
	}

	// Convert to export row
	row := StreakDataExportRow{
		CurrentStreak:      data.CurrentStreak,
		CurrentStreakType:  data.CurrentStreakType,
		LongestWinStreak:   data.LongestWinStreak,
		LongestLossStreak:  data.LongestLossStreak,
		CurrentWinStreak:   data.CurrentWinStreak,
		CurrentLossStreak:  data.CurrentLossStreak,
		TotalStreaksOver5:  data.TotalStreaksOver5,
		TotalStreaksOver10: data.TotalStreaksOver10,
		LastMatchResult:    data.LastMatchResult,
		LastMatchTimestamp: data.LastMatchTimestamp,
	}

	switch opts.Format {
	case FormatCSV:
		// Export as single-row CSV
		exporter := NewExporter(opts)
		return exporter.Export([]StreakDataExportRow{row})
	case FormatJSON:
		// Export data directly in JSON
		exporter := NewExporter(opts)
		return exporter.Export(data)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportStreakHistory exports historical streaks (length >= minLength).
func ExportStreakHistory(ctx context.Context, service *storage.Service, filter models.StatsFilter, minLength int, opts Options) error {
	history, err := service.GetStreakHistory(ctx, filter, minLength)
	if err != nil {
		return fmt.Errorf("failed to get streak history: %w", err)
	}

	if len(history) == 0 {
		return fmt.Errorf("no streaks found with length >= %d", minLength)
	}

	// Convert to export rows
	rows := make([]StreakHistoryExportRow, len(history))
	for i, streak := range history {
		rows[i] = StreakHistoryExportRow{
			StreakType: streak.StreakType,
			Length:     streak.Length,
			StartDate:  streak.StartDate,
			EndDate:    streak.EndDate,
			Format:     streak.Format,
			EventName:  streak.EventName,
		}
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

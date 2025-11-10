package export

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// RankTimelineEntryExportRow represents a timeline entry for CSV export.
type RankTimelineEntryExportRow struct {
	Date          string  `csv:"date" json:"date"`
	Rank          string  `csv:"rank" json:"rank"`
	RankClass     string  `csv:"rank_class" json:"rank_class"`
	RankLevel     int     `csv:"rank_level" json:"rank_level"`
	RankStep      int     `csv:"rank_step" json:"rank_step"`
	Percentile    float64 `csv:"percentile" json:"percentile"`
	Format        string  `csv:"format" json:"format"`
	SeasonOrdinal int     `csv:"season_ordinal" json:"season_ordinal"`
	IsChange      bool    `csv:"is_change" json:"is_change"`
	IsMilestone   bool    `csv:"is_milestone" json:"is_milestone"`
}

// RankTimelineSummaryExportRow represents timeline summary for CSV export.
type RankTimelineSummaryExportRow struct {
	Format         string `csv:"format" json:"format"`
	StartDate      string `csv:"start_date" json:"start_date"`
	EndDate        string `csv:"end_date" json:"end_date"`
	StartRank      string `csv:"start_rank" json:"start_rank"`
	EndRank        string `csv:"end_rank" json:"end_rank"`
	HighestRank    string `csv:"highest_rank" json:"highest_rank"`
	LowestRank     string `csv:"lowest_rank" json:"lowest_rank"`
	TotalChanges   int    `csv:"total_changes" json:"total_changes"`
	Milestones     int    `csv:"milestones" json:"milestones"`
	TotalSnapshots int    `csv:"total_snapshots" json:"total_snapshots"`
}

// ExportRankTimeline exports the rank progression timeline.
func ExportRankTimeline(ctx context.Context, service *storage.Service, format string, filter models.StatsFilter, period storage.TimelinePeriod, opts Options) error {
	timeline, err := service.GetRankProgressionTimeline(ctx, format, filter.StartDate, filter.EndDate, period)
	if err != nil {
		return fmt.Errorf("failed to get rank timeline: %w", err)
	}

	if timeline == nil || len(timeline.Entries) == 0 {
		return fmt.Errorf("no rank timeline data available")
	}

	switch opts.Format {
	case FormatCSV:
		// Export timeline entries as rows
		rows := make([]RankTimelineEntryExportRow, len(timeline.Entries))
		for i, entry := range timeline.Entries {
			rankClass := ""
			if entry.RankClass != nil {
				rankClass = *entry.RankClass
			}
			rankLevel := 0
			if entry.RankLevel != nil {
				rankLevel = *entry.RankLevel
			}
			rankStep := 0
			if entry.RankStep != nil {
				rankStep = *entry.RankStep
			}
			percentile := 0.0
			if entry.Percentile != nil {
				percentile = *entry.Percentile
			}

			rows[i] = RankTimelineEntryExportRow{
				Date:          entry.Date,
				Rank:          entry.Rank,
				RankClass:     rankClass,
				RankLevel:     rankLevel,
				RankStep:      rankStep,
				Percentile:    percentile,
				Format:        entry.Format,
				SeasonOrdinal: entry.SeasonOrdinal,
				IsChange:      entry.IsChange,
				IsMilestone:   entry.IsMilestone,
			}
		}
		exporter := NewExporter(opts)
		return exporter.Export(rows)
	case FormatJSON:
		// Export full timeline structure in JSON
		exporter := NewExporter(opts)
		return exporter.Export(timeline)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportRankTimelineSummary exports a summary of the rank progression timeline.
func ExportRankTimelineSummary(ctx context.Context, service *storage.Service, format string, filter models.StatsFilter, period storage.TimelinePeriod, opts Options) error {
	timeline, err := service.GetRankProgressionTimeline(ctx, format, filter.StartDate, filter.EndDate, period)
	if err != nil {
		return fmt.Errorf("failed to get rank timeline: %w", err)
	}

	if timeline == nil {
		return fmt.Errorf("no rank timeline data available")
	}

	// Create summary row
	row := RankTimelineSummaryExportRow{
		Format:         timeline.Format,
		StartDate:      timeline.StartDate.Format("2006-01-02"),
		EndDate:        timeline.EndDate.Format("2006-01-02"),
		StartRank:      timeline.StartRank,
		EndRank:        timeline.EndRank,
		HighestRank:    timeline.HighestRank,
		LowestRank:     timeline.LowestRank,
		TotalChanges:   timeline.TotalChanges,
		Milestones:     timeline.Milestones,
		TotalSnapshots: len(timeline.Entries),
	}

	switch opts.Format {
	case FormatCSV:
		exporter := NewExporter(opts)
		return exporter.Export([]RankTimelineSummaryExportRow{row})
	case FormatJSON:
		exporter := NewExporter(opts)
		return exporter.Export(row)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

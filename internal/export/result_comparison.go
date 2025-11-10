package export

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ResultComparisonRow represents a comparison between two time periods for CSV export.
type ResultComparisonRow struct {
	ResultReason      string  `csv:"result_reason" json:"result_reason"`
	ResultType        string  `csv:"result_type" json:"result_type"`
	RecentCount       int     `csv:"recent_count" json:"recent_count"`
	RecentPercentage  float64 `csv:"recent_percentage" json:"recent_percentage"`
	AllTimeCount      int     `csv:"all_time_count" json:"all_time_count"`
	AllTimePercentage float64 `csv:"all_time_percentage" json:"all_time_percentage"`
	PercentageChange  float64 `csv:"percentage_change" json:"percentage_change"`
	Trend             string  `csv:"trend" json:"trend"`
}

// ResultComparisonJSON represents a comparison in JSON format.
type ResultComparisonJSON struct {
	RecentPeriod *ComparisonPeriodJSON    `json:"recent_period"`
	AllTime      *ComparisonPeriodJSON    `json:"all_time"`
	RecentDays   int                      `json:"recent_days"`
	RecentStart  string                   `json:"recent_start"`
	RecentEnd    string                   `json:"recent_end"`
	AllTimeStart string                   `json:"all_time_start"`
	AllTimeEnd   string                   `json:"all_time_end"`
	Comparison   []ResultComparisonDetail `json:"comparison"`
}

// ComparisonPeriodJSON represents a period breakdown for comparison.
type ComparisonPeriodJSON struct {
	Wins   *ResultTypeBreakdownJSON `json:"wins"`
	Losses *ResultTypeBreakdownJSON `json:"losses"`
}

// ResultComparisonDetail represents comparison details for a single result reason.
type ResultComparisonDetail struct {
	ResultReason      string  `json:"result_reason"`
	ResultType        string  `json:"result_type"`
	RecentCount       int     `json:"recent_count"`
	RecentPercentage  float64 `json:"recent_percentage"`
	AllTimeCount      int     `json:"all_time_count"`
	AllTimePercentage float64 `json:"all_time_percentage"`
	PercentageChange  float64 `json:"percentage_change"`
	Trend             string  `json:"trend"`
}

// ExportResultComparison exports a comparison of recent vs. all-time result breakdowns.
// recentDays specifies how many days to consider as "recent" (e.g., 7, 30).
func ExportResultComparison(ctx context.Context, service *storage.Service, recentDays int, formatFilter *string, opts Options) error {
	now := time.Now()
	recentStart := now.AddDate(0, 0, -recentDays)

	// Get recent matches
	recentFilter := models.StatsFilter{
		StartDate: &recentStart,
		EndDate:   &now,
		Format:    formatFilter,
	}
	recentMatches, err := service.GetMatches(ctx, recentFilter)
	if err != nil {
		return fmt.Errorf("failed to get recent matches: %w", err)
	}

	// Get all-time matches
	allTimeFilter := models.StatsFilter{
		Format: formatFilter,
	}
	allTimeMatches, err := service.GetMatches(ctx, allTimeFilter)
	if err != nil {
		return fmt.Errorf("failed to get all-time matches: %w", err)
	}

	if len(recentMatches) == 0 && len(allTimeMatches) == 0 {
		return fmt.Errorf("no matches found for comparison")
	}

	// Calculate breakdowns
	recentWins := calculateResultBreakdown(recentMatches, true)
	recentLosses := calculateResultBreakdown(recentMatches, false)
	allTimeWins := calculateResultBreakdown(allTimeMatches, true)
	allTimeLosses := calculateResultBreakdown(allTimeMatches, false)

	switch opts.Format {
	case FormatCSV:
		return exportResultComparisonCSV(recentWins, recentLosses, allTimeWins, allTimeLosses, recentDays, recentStart, now, opts)
	case FormatJSON:
		return exportResultComparisonJSON(recentWins, recentLosses, allTimeWins, allTimeLosses, recentDays, recentStart, now, allTimeMatches, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// exportResultComparisonCSV exports the comparison to CSV format.
func exportResultComparisonCSV(recentWins, recentLosses, allTimeWins, allTimeLosses resultBreakdown, recentDays int, recentStart, recentEnd time.Time, opts Options) error {
	var rows []ResultComparisonRow

	// Helper to add comparison rows
	addComparison := func(resultType, reason string, recentCount, allTimeCount int, recentTotal, allTimeTotal int) {
		recentPct := 0.0
		if recentTotal > 0 {
			recentPct = float64(recentCount) / float64(recentTotal) * 100
		}
		allTimePct := 0.0
		if allTimeTotal > 0 {
			allTimePct = float64(allTimeCount) / float64(allTimeTotal) * 100
		}

		change := recentPct - allTimePct
		trend := "stable"
		if change > 5.0 {
			trend = "increasing"
		} else if change < -5.0 {
			trend = "decreasing"
		}

		rows = append(rows, ResultComparisonRow{
			ResultReason:      reason,
			ResultType:        resultType,
			RecentCount:       recentCount,
			RecentPercentage:  recentPct,
			AllTimeCount:      allTimeCount,
			AllTimePercentage: allTimePct,
			PercentageChange:  change,
			Trend:             trend,
		})
	}

	// Add win comparisons
	addComparison("Win", "Normal", recentWins.Normal, allTimeWins.Normal, recentWins.Total, allTimeWins.Total)
	addComparison("Win", "Opponent Concede", recentWins.OpponentConcede, allTimeWins.OpponentConcede, recentWins.Total, allTimeWins.Total)
	addComparison("Win", "Opponent Timeout", recentWins.OpponentTimeout, allTimeWins.OpponentTimeout, recentWins.Total, allTimeWins.Total)
	addComparison("Win", "Opponent Disconnect", recentWins.OpponentDisconnect, allTimeWins.OpponentDisconnect, recentWins.Total, allTimeWins.Total)
	addComparison("Win", "Other", recentWins.Other, allTimeWins.Other, recentWins.Total, allTimeWins.Total)

	// Add loss comparisons
	addComparison("Loss", "Normal", recentLosses.Normal, allTimeLosses.Normal, recentLosses.Total, allTimeLosses.Total)
	addComparison("Loss", "Concede", recentLosses.Concede, allTimeLosses.Concede, recentLosses.Total, allTimeLosses.Total)
	addComparison("Loss", "Timeout", recentLosses.Timeout, allTimeLosses.Timeout, recentLosses.Total, allTimeLosses.Total)
	addComparison("Loss", "Disconnect", recentLosses.Disconnect, allTimeLosses.Disconnect, recentLosses.Total, allTimeLosses.Total)
	addComparison("Loss", "Other", recentLosses.Other, allTimeLosses.Other, recentLosses.Total, allTimeLosses.Total)

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// exportResultComparisonJSON exports the comparison to JSON format.
func exportResultComparisonJSON(recentWins, recentLosses, allTimeWins, allTimeLosses resultBreakdown, recentDays int, recentStart, recentEnd time.Time, allTimeMatches []*models.Match, opts Options) error {
	// Get all-time date range
	var allTimeStart, allTimeEnd time.Time
	if len(allTimeMatches) > 0 {
		allTimeStart = allTimeMatches[len(allTimeMatches)-1].Timestamp
		allTimeEnd = allTimeMatches[0].Timestamp
	}

	// Build comparison JSON
	comparison := ResultComparisonJSON{
		RecentPeriod: &ComparisonPeriodJSON{
			Wins:   breakdownToJSON(recentWins),
			Losses: breakdownToJSON(recentLosses),
		},
		AllTime: &ComparisonPeriodJSON{
			Wins:   breakdownToJSON(allTimeWins),
			Losses: breakdownToJSON(allTimeLosses),
		},
		RecentDays:   recentDays,
		RecentStart:  recentStart.Format("2006-01-02"),
		RecentEnd:    recentEnd.Format("2006-01-02"),
		AllTimeStart: allTimeStart.Format("2006-01-02"),
		AllTimeEnd:   allTimeEnd.Format("2006-01-02"),
		Comparison:   []ResultComparisonDetail{},
	}

	// Helper to add comparison details
	addDetail := func(resultType, reason string, recentCount, allTimeCount int, recentTotal, allTimeTotal int) {
		recentPct := 0.0
		if recentTotal > 0 {
			recentPct = float64(recentCount) / float64(recentTotal) * 100
		}
		allTimePct := 0.0
		if allTimeTotal > 0 {
			allTimePct = float64(allTimeCount) / float64(allTimeTotal) * 100
		}

		change := recentPct - allTimePct
		trend := "stable"
		if change > 5.0 {
			trend = "increasing"
		} else if change < -5.0 {
			trend = "decreasing"
		}

		comparison.Comparison = append(comparison.Comparison, ResultComparisonDetail{
			ResultReason:      reason,
			ResultType:        resultType,
			RecentCount:       recentCount,
			RecentPercentage:  recentPct,
			AllTimeCount:      allTimeCount,
			AllTimePercentage: allTimePct,
			PercentageChange:  change,
			Trend:             trend,
		})
	}

	// Add all comparisons
	addDetail("Win", "Normal", recentWins.Normal, allTimeWins.Normal, recentWins.Total, allTimeWins.Total)
	addDetail("Win", "Opponent Concede", recentWins.OpponentConcede, allTimeWins.OpponentConcede, recentWins.Total, allTimeWins.Total)
	addDetail("Win", "Opponent Timeout", recentWins.OpponentTimeout, allTimeWins.OpponentTimeout, recentWins.Total, allTimeWins.Total)
	addDetail("Win", "Opponent Disconnect", recentWins.OpponentDisconnect, allTimeWins.OpponentDisconnect, recentWins.Total, allTimeWins.Total)
	addDetail("Loss", "Normal", recentLosses.Normal, allTimeLosses.Normal, recentLosses.Total, allTimeLosses.Total)
	addDetail("Loss", "Concede", recentLosses.Concede, allTimeLosses.Concede, recentLosses.Total, allTimeLosses.Total)
	addDetail("Loss", "Timeout", recentLosses.Timeout, allTimeLosses.Timeout, recentLosses.Total, allTimeLosses.Total)
	addDetail("Loss", "Disconnect", recentLosses.Disconnect, allTimeLosses.Disconnect, recentLosses.Total, allTimeLosses.Total)

	exporter := NewExporter(opts)
	return exporter.Export(comparison)
}

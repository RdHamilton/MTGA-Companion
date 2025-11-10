package export

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ResultTrendPeriodRow represents result breakdown for a single period in CSV format.
type ResultTrendPeriodRow struct {
	PeriodLabel           string  `csv:"period_label" json:"period_label"`
	StartDate             string  `csv:"start_date" json:"start_date"`
	EndDate               string  `csv:"end_date" json:"end_date"`
	ResultType            string  `csv:"result_type" json:"result_type"`
	ResultReason          string  `csv:"result_reason" json:"result_reason"`
	Count                 int     `csv:"count" json:"count"`
	Percentage            float64 `csv:"percentage" json:"percentage"`
	TrendVsPreviousPeriod string  `csv:"trend_vs_previous" json:"trend_vs_previous"`
	ChangePct             float64 `csv:"change_pct" json:"change_pct"`
}

// ResultTrendAnalysisJSON represents the complete result trend analysis in JSON format.
type ResultTrendAnalysisJSON struct {
	Periods     []ResultTrendPeriodJSON `json:"periods"`
	StartDate   string                  `json:"start_date"`
	EndDate     string                  `json:"end_date"`
	PeriodType  string                  `json:"period_type"`
	PeriodCount int                     `json:"period_count"`
	Summary     *ResultTrendSummary     `json:"summary"`
}

// ResultTrendPeriodJSON represents a single period in JSON format.
type ResultTrendPeriodJSON struct {
	Label       string                   `json:"label"`
	StartDate   string                   `json:"start_date"`
	EndDate     string                   `json:"end_date"`
	Wins        *ResultTypeBreakdownJSON `json:"wins"`
	Losses      *ResultTypeBreakdownJSON `json:"losses"`
	TotalWins   int                      `json:"total_wins"`
	TotalLosses int                      `json:"total_losses"`
}

// ResultTrendSummary summarizes trends across all periods.
type ResultTrendSummary struct {
	MostIncreasingWinReason  string  `json:"most_increasing_win_reason"`
	MostDecreasingWinReason  string  `json:"most_decreasing_win_reason"`
	MostIncreasingLossReason string  `json:"most_increasing_loss_reason"`
	MostDecreasingLossReason string  `json:"most_decreasing_loss_reason"`
	OpponentConcedeChange    float64 `json:"opponent_concede_change"`
	NormalWinChange          float64 `json:"normal_win_change"`
	ConcedeChange            float64 `json:"concede_change"`
}

// ExportResultTrends exports result breakdown trends over time.
func ExportResultTrends(ctx context.Context, service *storage.Service, startDate, endDate time.Time, periodType string, formatFilter *string, opts Options) error {
	// Generate periods
	periods := generateResultPeriods(startDate, endDate, periodType)

	if len(periods) == 0 {
		return fmt.Errorf("no periods generated for the specified date range")
	}

	// Calculate result breakdowns for each period
	var periodData []periodResultData
	for _, period := range periods {
		filter := models.StatsFilter{
			StartDate: &period.StartDate,
			EndDate:   &period.EndDate,
			Format:    formatFilter,
		}

		matches, err := service.GetMatches(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to get matches for period %s: %w", period.Label, err)
		}

		wins := calculateResultBreakdown(matches, true)
		losses := calculateResultBreakdown(matches, false)

		periodData = append(periodData, periodResultData{
			Period: period,
			Wins:   wins,
			Losses: losses,
		})
	}

	switch opts.Format {
	case FormatCSV:
		return exportResultTrendsCSV(periodData, startDate, endDate, periodType, opts)
	case FormatJSON:
		return exportResultTrendsJSON(periodData, startDate, endDate, periodType, opts)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// periodResultData holds result breakdown data for a single period.
type periodResultData struct {
	Period resultPeriod
	Wins   resultBreakdown
	Losses resultBreakdown
}

// resultPeriod represents a time period for result trend analysis.
type resultPeriod struct {
	StartDate time.Time
	EndDate   time.Time
	Label     string
}

// generateResultPeriods generates time periods for result trend analysis.
func generateResultPeriods(startDate, endDate time.Time, periodType string) []resultPeriod {
	var periods []resultPeriod

	switch periodType {
	case "daily":
		current := startDate
		for current.Before(endDate) || current.Equal(endDate) {
			periodEnd := current.AddDate(0, 0, 1)
			if periodEnd.After(endDate) {
				periodEnd = endDate
			}
			periods = append(periods, resultPeriod{
				StartDate: current,
				EndDate:   periodEnd,
				Label:     current.Format("2006-01-02"),
			})
			current = periodEnd
		}

	case "weekly":
		current := startDate
		for current.Before(endDate) {
			weekday := int(current.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			weekStart := current.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			weekEnd := weekStart.AddDate(0, 0, 7)
			if weekEnd.After(endDate) {
				weekEnd = endDate
			}

			periods = append(periods, resultPeriod{
				StartDate: weekStart,
				EndDate:   weekEnd,
				Label:     fmt.Sprintf("Week of %s", weekStart.Format("2006-01-02")),
			})
			current = weekEnd
		}

	case "monthly":
		current := startDate
		for current.Before(endDate) {
			monthStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
			monthEnd := monthStart.AddDate(0, 1, 0)
			if monthEnd.After(endDate) {
				monthEnd = endDate
			}

			periods = append(periods, resultPeriod{
				StartDate: monthStart,
				EndDate:   monthEnd,
				Label:     monthStart.Format("2006-01"),
			})
			current = monthEnd
		}

	default:
		// Default to weekly
		return generateResultPeriods(startDate, endDate, "weekly")
	}

	return periods
}

// exportResultTrendsCSV exports result trends to CSV format.
func exportResultTrendsCSV(periodData []periodResultData, startDate, endDate time.Time, periodType string, opts Options) error {
	var rows []ResultTrendPeriodRow

	// Helper to add rows with trend calculation
	addRows := func(periodIdx int, period resultPeriod, breakdown resultBreakdown, resultType string) {
		var prevBreakdown resultBreakdown
		if periodIdx > 0 {
			if resultType == "Win" {
				prevBreakdown = periodData[periodIdx-1].Wins
			} else {
				prevBreakdown = periodData[periodIdx-1].Losses
			}
		}

		addRow := func(reason string, count int, prevCount int) {
			pct := 0.0
			if breakdown.Total > 0 {
				pct = float64(count) / float64(breakdown.Total) * 100
			}

			prevPct := 0.0
			if periodIdx > 0 && prevBreakdown.Total > 0 {
				prevPct = float64(prevCount) / float64(prevBreakdown.Total) * 100
			}

			change := pct - prevPct
			trend := "stable"
			if periodIdx > 0 {
				if change > 5.0 {
					trend = "increasing"
				} else if change < -5.0 {
					trend = "decreasing"
				}
			}

			rows = append(rows, ResultTrendPeriodRow{
				PeriodLabel:           period.Label,
				StartDate:             period.StartDate.Format("2006-01-02"),
				EndDate:               period.EndDate.Format("2006-01-02"),
				ResultType:            resultType,
				ResultReason:          reason,
				Count:                 count,
				Percentage:            pct,
				TrendVsPreviousPeriod: trend,
				ChangePct:             change,
			})
		}

		if resultType == "Win" {
			addRow("Normal", breakdown.Normal, prevBreakdown.Normal)
			addRow("Opponent Concede", breakdown.OpponentConcede, prevBreakdown.OpponentConcede)
			addRow("Opponent Timeout", breakdown.OpponentTimeout, prevBreakdown.OpponentTimeout)
			addRow("Opponent Disconnect", breakdown.OpponentDisconnect, prevBreakdown.OpponentDisconnect)
			addRow("Other", breakdown.Other, prevBreakdown.Other)
		} else {
			addRow("Normal", breakdown.Normal, prevBreakdown.Normal)
			addRow("Concede", breakdown.Concede, prevBreakdown.Concede)
			addRow("Timeout", breakdown.Timeout, prevBreakdown.Timeout)
			addRow("Disconnect", breakdown.Disconnect, prevBreakdown.Disconnect)
			addRow("Other", breakdown.Other, prevBreakdown.Other)
		}
	}

	// Add rows for each period
	for i, pd := range periodData {
		addRows(i, pd.Period, pd.Wins, "Win")
		addRows(i, pd.Period, pd.Losses, "Loss")
	}

	exporter := NewExporter(opts)
	return exporter.Export(rows)
}

// exportResultTrendsJSON exports result trends to JSON format.
func exportResultTrendsJSON(periodData []periodResultData, startDate, endDate time.Time, periodType string, opts Options) error {
	jsonData := ResultTrendAnalysisJSON{
		StartDate:   startDate.Format("2006-01-02"),
		EndDate:     endDate.Format("2006-01-02"),
		PeriodType:  periodType,
		PeriodCount: len(periodData),
		Periods:     make([]ResultTrendPeriodJSON, len(periodData)),
	}

	// Build periods
	for i, pd := range periodData {
		jsonData.Periods[i] = ResultTrendPeriodJSON{
			Label:       pd.Period.Label,
			StartDate:   pd.Period.StartDate.Format("2006-01-02"),
			EndDate:     pd.Period.EndDate.Format("2006-01-02"),
			Wins:        breakdownToJSON(pd.Wins),
			Losses:      breakdownToJSON(pd.Losses),
			TotalWins:   pd.Wins.Total,
			TotalLosses: pd.Losses.Total,
		}
	}

	// Calculate summary trends
	if len(periodData) >= 2 {
		first := periodData[0]
		last := periodData[len(periodData)-1]

		jsonData.Summary = &ResultTrendSummary{
			OpponentConcedeChange: calculatePercentageChange(first.Wins.OpponentConcede, first.Wins.Total, last.Wins.OpponentConcede, last.Wins.Total),
			NormalWinChange:       calculatePercentageChange(first.Wins.Normal, first.Wins.Total, last.Wins.Normal, last.Wins.Total),
			ConcedeChange:         calculatePercentageChange(first.Losses.Concede, first.Losses.Total, last.Losses.Concede, last.Losses.Total),
		}

		// Find most increasing/decreasing reasons
		jsonData.Summary.MostIncreasingWinReason = findMostChangingReason(first.Wins, last.Wins, true)
		jsonData.Summary.MostDecreasingWinReason = findMostChangingReason(first.Wins, last.Wins, false)
		jsonData.Summary.MostIncreasingLossReason = findMostChangingReason(first.Losses, last.Losses, true)
		jsonData.Summary.MostDecreasingLossReason = findMostChangingReason(first.Losses, last.Losses, false)
	}

	exporter := NewExporter(opts)
	return exporter.Export(jsonData)
}

// calculatePercentageChange calculates the percentage point change between two periods.
func calculatePercentageChange(firstCount, firstTotal, lastCount, lastTotal int) float64 {
	firstPct := 0.0
	if firstTotal > 0 {
		firstPct = float64(firstCount) / float64(firstTotal) * 100
	}
	lastPct := 0.0
	if lastTotal > 0 {
		lastPct = float64(lastCount) / float64(lastTotal) * 100
	}
	return lastPct - firstPct
}

// findMostChangingReason finds the result reason with the most change (increasing or decreasing).
func findMostChangingReason(first, last resultBreakdown, increasing bool) string {
	type reasonChange struct {
		name   string
		change float64
	}

	reasons := []reasonChange{
		{"Normal", calculatePercentageChange(first.Normal, first.Total, last.Normal, last.Total)},
		{"Concede", calculatePercentageChange(first.Concede, first.Total, last.Concede, last.Total)},
		{"Timeout", calculatePercentageChange(first.Timeout, first.Total, last.Timeout, last.Total)},
		{"Disconnect", calculatePercentageChange(first.Disconnect, first.Total, last.Disconnect, last.Total)},
		{"Opponent Concede", calculatePercentageChange(first.OpponentConcede, first.Total, last.OpponentConcede, last.Total)},
		{"Opponent Timeout", calculatePercentageChange(first.OpponentTimeout, first.Total, last.OpponentTimeout, last.Total)},
		{"Opponent Disconnect", calculatePercentageChange(first.OpponentDisconnect, first.Total, last.OpponentDisconnect, last.Total)},
	}

	if len(reasons) == 0 {
		return "None"
	}

	best := reasons[0]
	for _, r := range reasons[1:] {
		if increasing {
			if r.change > best.change {
				best = r
			}
		} else {
			if r.change < best.change {
				best = r
			}
		}
	}

	if (increasing && best.change <= 0) || (!increasing && best.change >= 0) {
		return "None"
	}

	return best.name
}

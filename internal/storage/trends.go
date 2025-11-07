package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// TrendPeriod represents a time period for trend analysis.
type TrendPeriod struct {
	StartDate time.Time
	EndDate   time.Time
	Label     string
}

// TrendData represents statistics for a specific time period.
type TrendData struct {
	Period      TrendPeriod
	Stats       *models.Statistics
	WinRate     float64
	GameWinRate float64
}

// TrendAnalysis represents historical trend analysis results.
type TrendAnalysis struct {
	Periods    []TrendData
	Overall    *models.Statistics
	Trend      string  // "improving", "declining", "stable"
	TrendValue float64 // Percentage change
}

// GetTrendAnalysis calculates historical trend analysis for the specified date range.
func (s *Service) GetTrendAnalysis(ctx context.Context, startDate, endDate time.Time, periodType string) (*TrendAnalysis, error) {
	analysis := &TrendAnalysis{
		Periods: []TrendData{},
	}

	// Generate periods based on type
	periods := generatePeriods(startDate, endDate, periodType)

	// Calculate stats for each period
	for _, period := range periods {
		filter := models.StatsFilter{
			StartDate: &period.StartDate,
			EndDate:   &period.EndDate,
		}

		stats, err := s.GetStats(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for period %s: %w", period.Label, err)
		}

		if stats != nil && stats.TotalMatches > 0 {
			trendData := TrendData{
				Period:      period,
				Stats:       stats,
				WinRate:     stats.WinRate,
				GameWinRate: stats.GameWinRate,
			}
			analysis.Periods = append(analysis.Periods, trendData)
		}
	}

	// Calculate overall stats
	overallFilter := models.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	overall, err := s.GetStats(ctx, overallFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}
	analysis.Overall = overall

	// Calculate trend
	if len(analysis.Periods) >= 2 {
		firstPeriod := analysis.Periods[0]
		lastPeriod := analysis.Periods[len(analysis.Periods)-1]

		if firstPeriod.Stats.TotalMatches > 0 && lastPeriod.Stats.TotalMatches > 0 {
			trendValue := lastPeriod.WinRate - firstPeriod.WinRate
			analysis.TrendValue = trendValue

			if trendValue > 0.05 { // 5% improvement
				analysis.Trend = "improving"
			} else if trendValue < -0.05 { // 5% decline
				analysis.Trend = "declining"
			} else {
				analysis.Trend = "stable"
			}
		}
	}

	return analysis, nil
}

// generatePeriods generates time periods for trend analysis.
func generatePeriods(startDate, endDate time.Time, periodType string) []TrendPeriod {
	var periods []TrendPeriod

	switch periodType {
	case "daily":
		current := startDate
		for current.Before(endDate) || current.Equal(endDate) {
			periodEnd := current.AddDate(0, 0, 1)
			if periodEnd.After(endDate) {
				periodEnd = endDate
			}
			periods = append(periods, TrendPeriod{
				StartDate: current,
				EndDate:   periodEnd,
				Label:     current.Format("2006-01-02"),
			})
			current = periodEnd
		}

	case "weekly":
		current := startDate
		for current.Before(endDate) {
			// Get start of week (Monday)
			weekday := int(current.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday is 7
			}
			weekStart := current.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			weekEnd := weekStart.AddDate(0, 0, 7)
			if weekEnd.After(endDate) {
				weekEnd = endDate
			}

			periods = append(periods, TrendPeriod{
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

			periods = append(periods, TrendPeriod{
				StartDate: monthStart,
				EndDate:   monthEnd,
				Label:     monthStart.Format("2006-01"),
			})
			current = monthEnd
		}

	default:
		// Default to weekly
		return generatePeriods(startDate, endDate, "weekly")
	}

	return periods
}

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// displayTrendAnalysis displays historical trend analysis.
func displayTrendAnalysis(analysis *storage.TrendAnalysis) {
	if analysis == nil || len(analysis.Periods) == 0 {
		fmt.Println("No trend data available.")
		return
	}

	fmt.Println("Win Rate Trends")
	fmt.Println("===============")
	fmt.Println()

	// Display trend summary
	if analysis.Overall != nil && analysis.Overall.TotalMatches > 0 {
		fmt.Printf("Overall: %d-%d (%.1f%% win rate)\n",
			analysis.Overall.MatchesWon, analysis.Overall.MatchesLost, analysis.Overall.WinRate*100)
		if analysis.Trend != "" {
			trendSymbol := ""
			switch analysis.Trend {
			case "improving":
				trendSymbol = "↑"
			case "declining":
				trendSymbol = "↓"
			case "stable":
				trendSymbol = "→"
			}
			fmt.Printf("Trend: %s %.1f%% from first to last period\n", trendSymbol, analysis.TrendValue*100)
		}
		fmt.Println()
	}

	// Display period-by-period trends
	for i, period := range analysis.Periods {
		if period.Stats.TotalMatches > 0 {
			fmt.Printf("%s: %d-%d (%.1f%% win rate)\n",
				period.Period.Label,
				period.Stats.MatchesWon,
				period.Stats.MatchesLost,
				period.WinRate*100)

			// Show trend arrow if not first period
			if i > 0 {
				prevPeriod := analysis.Periods[i-1]
				if prevPeriod.Stats.TotalMatches > 0 {
					change := period.WinRate - prevPeriod.WinRate
					if change > 0.01 {
						fmt.Printf("  ↑ +%.1f%%\n", change*100)
					} else if change < -0.01 {
						fmt.Printf("  ↓ %.1f%%\n", change*100)
					} else {
						fmt.Printf("  → %.1f%%\n", change*100)
					}
				}
			}
		}
	}

	fmt.Println()
}

// displayTrendAnalysisForPeriod displays trend analysis for a specific time period.
func displayTrendAnalysisForPeriod(service *storage.Service, ctx context.Context, days int, periodType string) {
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -days)

	analysis, err := service.GetTrendAnalysis(ctx, startDate, endDate, periodType, nil)
	if err != nil {
		log.Printf("Warning: Failed to retrieve trend analysis: %v", err)
		return
	}

	if analysis == nil || len(analysis.Periods) == 0 {
		fmt.Printf("No trend data available for the last %d days.\n", days)
		return
	}

	fmt.Printf("Win Rate Trends (Last %d Days)\n", days)
	fmt.Println("--------------------------------")
	displayTrendAnalysis(analysis)
}

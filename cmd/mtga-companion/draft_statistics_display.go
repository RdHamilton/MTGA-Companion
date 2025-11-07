package main

import (
	"fmt"
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// displayDraftStatistics displays aggregated draft performance statistics.
func displayDraftStatistics(stats *logreader.DraftStatistics) {
	if stats == nil {
		fmt.Println("No draft statistics available.")
		return
	}

	fmt.Println("Draft Statistics")
	fmt.Println("================")
	fmt.Println()

	// Overall statistics
	totalMatches := stats.TotalWins + stats.TotalLosses
	winRate := 0.0
	if totalMatches > 0 {
		winRate = float64(stats.TotalWins) / float64(totalMatches) * 100
	}

	fmt.Printf("Lifetime Record: %d-%d (%.1f%% win rate)\n", stats.TotalWins, stats.TotalLosses, winRate)
	fmt.Printf("Total Drafts: %d\n", stats.TotalDrafts)
	if stats.TotalDrafts > 0 {
		fmt.Printf("Average Wins: %.1f\n", stats.AverageWins)
	}
	if stats.BestRecord.Wins > 0 {
		fmt.Printf("Best Run: %d-%d (%s)\n", stats.BestRecord.Wins, stats.BestRecord.Losses, stats.BestRecord.EventName)
	}
	if stats.TrophyCount > 0 {
		fmt.Printf("Trophies (7+ wins): %d\n", stats.TrophyCount)
	}
	fmt.Println()

	// Format statistics
	if len(stats.DraftsByFormat) > 0 {
		fmt.Println("By Format:")
		formats := make([]string, 0, len(stats.DraftsByFormat))
		for format := range stats.DraftsByFormat {
			formats = append(formats, format)
		}
		sort.Strings(formats)

		for _, format := range formats {
			formatStats := stats.DraftsByFormat[format]
			if formatStats.DraftCount > 0 {
				fmt.Printf("  %s: %d drafts, %d-%d (%.1f%% win rate, %.1f avg wins)\n",
					format,
					formatStats.DraftCount,
					formatStats.Wins,
					formatStats.Losses,
					formatStats.WinRate*100,
					formatStats.AverageWins)
			}
		}
		fmt.Println()
	}

	// Set statistics
	if len(stats.DraftsBySet) > 0 {
		fmt.Println("By Set:")
		sets := make([]string, 0, len(stats.DraftsBySet))
		for set := range stats.DraftsBySet {
			sets = append(sets, set)
		}
		sort.Strings(sets)

		for _, set := range sets {
			setStats := stats.DraftsBySet[set]
			if setStats.DraftCount > 0 {
				fmt.Printf("  %s: %d drafts, %d-%d (%.1f%% win rate, %.1f avg wins)\n",
					set,
					setStats.DraftCount,
					setStats.Wins,
					setStats.Losses,
					setStats.WinRate*100,
					setStats.AverageWins)
			}
		}
		fmt.Println()
	}
}


package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// displaySeasonalProgression displays rank progression across seasons.
func displaySeasonalProgression(service *storage.Service, ctx context.Context, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	summaries, err := service.GetSeasonalRankSummary(ctx, format)
	if err != nil {
		log.Printf("Error retrieving seasonal rank summary: %v", err)
		return
	}

	if len(summaries) == 0 {
		fmt.Printf("No seasonal rank data available for %s format.\n", format)
		return
	}

	fmt.Printf("\nSeasonal Rank Progression - %s\n", capitalizeFirst(format))
	fmt.Println("============================")
	fmt.Println()

	for _, summary := range summaries {
		displaySeasonSummary(summary)
	}
}

// displaySeasonSummary displays a single season's rank summary.
func displaySeasonSummary(summary *models.SeasonalRankSummary) {
	fmt.Printf("Season %d\n", summary.SeasonOrdinal)
	fmt.Printf("  Period: %s to %s\n",
		summary.FirstSeen.Format("2006-01-02"),
		summary.LastSeen.Format("2006-01-02"))

	if summary.StartRank != nil {
		fmt.Printf("  Starting Rank: %s\n", *summary.StartRank)
	}

	if summary.EndRank != nil {
		fmt.Printf("  Ending Rank:   %s\n", *summary.EndRank)
	}

	if summary.HighestRank != nil {
		fmt.Printf("  Peak Rank:     %s\n", *summary.HighestRank)
	}

	if summary.LowestRank != nil {
		fmt.Printf("  Lowest Rank:   %s\n", *summary.LowestRank)
	}

	fmt.Printf("  Snapshots:     %d\n", summary.TotalSnapshots)

	// Calculate progression
	if summary.StartRank != nil && summary.EndRank != nil {
		if *summary.StartRank != *summary.EndRank {
			fmt.Printf("  Progress:      %s → %s\n", *summary.StartRank, *summary.EndRank)
		}
	}

	fmt.Println()
}

// displayRankAchievements displays all rank achievements for a format.
func displayRankAchievements(service *storage.Service, ctx context.Context, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	achievements, err := service.GetRankAchievements(ctx, format)
	if err != nil {
		log.Printf("Error retrieving rank achievements: %v", err)
		return
	}

	if len(achievements) == 0 {
		fmt.Printf("No rank achievements available for %s format.\n", format)
		return
	}

	fmt.Printf("\nRank Achievements - %s\n", capitalizeFirst(format))
	fmt.Println("======================")
	fmt.Println()

	// Display highest rank first
	for _, achievement := range achievements {
		if achievement.IsHighest {
			displayAchievement(achievement, true)
		}
	}

	// Display other achievements
	for _, achievement := range achievements {
		if !achievement.IsHighest {
			displayAchievement(achievement, false)
		}
	}
}

// displayAchievement displays a single rank achievement.
func displayAchievement(achievement *models.RankAchievement, highlight bool) {
	rankStr := achievement.RankClass
	if achievement.RankLevel != nil {
		rankStr = fmt.Sprintf("%s %d", rankStr, *achievement.RankLevel)
	}

	marker := " "
	if highlight {
		marker = "★"
	}

	fmt.Printf("%s %s\n", marker, rankStr)
	fmt.Printf("    First Achieved: %s (Season %d)\n",
		achievement.FirstAchieved.Format("2006-01-02 15:04"),
		achievement.SeasonOrdinal)

	if highlight {
		fmt.Println("    Status: Highest Rank Achieved")
	}

	fmt.Println()
}

// displaySeasonComparison displays side-by-side comparison of recent seasons.
func displaySeasonComparison(service *storage.Service, ctx context.Context, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	summaries, err := service.GetSeasonalRankSummary(ctx, format)
	if err != nil {
		log.Printf("Error retrieving seasonal rank summary: %v", err)
		return
	}

	if len(summaries) < 2 {
		fmt.Printf("Need at least 2 seasons for comparison. Current seasons: %d\n", len(summaries))
		return
	}

	fmt.Printf("\nSeason Comparison - %s\n", capitalizeFirst(format))
	fmt.Println("=======================")
	fmt.Println()

	// Compare most recent seasons (limit to 5)
	limit := 5
	if len(summaries) < limit {
		limit = len(summaries)
	}

	fmt.Printf("%-8s %-20s %-20s %-20s\n", "Season", "Starting Rank", "Ending Rank", "Peak Rank")
	fmt.Println(strings.Repeat("-", 70))

	for i := 0; i < limit; i++ {
		summary := summaries[i]
		startRank := "N/A"
		endRank := "N/A"
		peakRank := "N/A"

		if summary.StartRank != nil {
			startRank = *summary.StartRank
		}
		if summary.EndRank != nil {
			endRank = *summary.EndRank
		}
		if summary.HighestRank != nil {
			peakRank = *summary.HighestRank
		}

		fmt.Printf("%-8d %-20s %-20s %-20s\n",
			summary.SeasonOrdinal,
			startRank,
			endRank,
			peakRank)
	}

	fmt.Println()
}

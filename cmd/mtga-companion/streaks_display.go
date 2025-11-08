package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/stats"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// displayStreakStats displays win/loss streak information.
func displayStreakStats(service *storage.Service, ctx context.Context, filter models.StatsFilter) {
	streaks, err := service.GetStreakStats(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve streak statistics: %v", err)
		return
	}

	if streaks.LongestWinStreak == 0 && streaks.LongestLossStreak == 0 {
		// No streaks to display
		return
	}

	fmt.Println("Streaks")
	fmt.Println("-------")

	// Display current streak
	fmt.Printf("Current: %s\n", stats.FormatCurrentStreak(streaks.CurrentStreak))

	// Display longest streaks
	if streaks.LongestWinStreak > 0 {
		fmt.Printf("Longest win streak: %d\n", streaks.LongestWinStreak)
	}
	if streaks.LongestLossStreak > 0 {
		fmt.Printf("Longest loss streak: %d\n", streaks.LongestLossStreak)
	}

	fmt.Println()
}

// displayPerformanceMetrics displays duration-based performance metrics.
func displayPerformanceMetrics(service *storage.Service, ctx context.Context, filter models.StatsFilter) {
	metrics, err := service.GetPerformanceMetrics(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve performance metrics: %v", err)
		return
	}

	// Check if we have any metrics to display
	hasMetrics := metrics.AvgMatchDuration != nil ||
		metrics.AvgGameDuration != nil ||
		metrics.FastestMatch != nil ||
		metrics.SlowestMatch != nil ||
		metrics.FastestGame != nil ||
		metrics.SlowestGame != nil

	if !hasMetrics {
		return
	}

	fmt.Println("Performance Metrics")
	fmt.Println("-------------------")

	// Display match metrics
	if metrics.AvgMatchDuration != nil || metrics.FastestMatch != nil || metrics.SlowestMatch != nil {
		fmt.Println("Matches:")
		if metrics.AvgMatchDuration != nil {
			fmt.Printf("  Average duration: %s\n", formatDuration(time.Duration(*metrics.AvgMatchDuration)*time.Second))
		}
		if metrics.FastestMatch != nil {
			fmt.Printf("  Fastest: %s\n", formatDuration(time.Duration(*metrics.FastestMatch)*time.Second))
		}
		if metrics.SlowestMatch != nil {
			fmt.Printf("  Slowest: %s\n", formatDuration(time.Duration(*metrics.SlowestMatch)*time.Second))
		}
	}

	// Display game metrics
	if metrics.AvgGameDuration != nil || metrics.FastestGame != nil || metrics.SlowestGame != nil {
		fmt.Println("Games:")
		if metrics.AvgGameDuration != nil {
			fmt.Printf("  Average duration: %s\n", formatDuration(time.Duration(*metrics.AvgGameDuration)*time.Second))
		}
		if metrics.FastestGame != nil {
			fmt.Printf("  Fastest: %s\n", formatDuration(time.Duration(*metrics.FastestGame)*time.Second))
		}
		if metrics.SlowestGame != nil {
			fmt.Printf("  Slowest: %s\n", formatDuration(time.Duration(*metrics.SlowestGame)*time.Second))
		}
	}

	fmt.Println()
}

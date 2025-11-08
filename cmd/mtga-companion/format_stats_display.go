package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// displayStatsByFormat displays all-time statistics grouped by format.
func displayStatsByFormat(service *storage.Service, ctx context.Context) {
	fmt.Println("\nStatistics by Format (All-Time)")
	fmt.Println("===============================")

	filter := models.StatsFilter{}
	statsByFormat, err := service.GetStatsByFormat(ctx, filter)
	if err != nil {
		log.Printf("Error retrieving statistics by format: %v", err)
		return
	}

	if len(statsByFormat) == 0 {
		fmt.Println("No statistics found.")
		return
	}

	// Sort formats alphabetically for consistent display
	formats := make([]string, 0, len(statsByFormat))
	for format := range statsByFormat {
		formats = append(formats, format)
	}
	sort.Strings(formats)

	fmt.Println()

	// Display statistics for each format
	for _, format := range formats {
		stats := statsByFormat[format]

		fmt.Printf("%s:\n", format)
		fmt.Println("  Matches:")
		fmt.Printf("    Total:     %d\n", stats.TotalMatches)
		fmt.Printf("    Won:       %d\n", stats.MatchesWon)
		fmt.Printf("    Lost:      %d\n", stats.MatchesLost)
		if stats.TotalMatches > 0 {
			fmt.Printf("    Win Rate:  %.1f%%\n", stats.WinRate*100)
		}

		fmt.Println("  Games:")
		fmt.Printf("    Total:     %d\n", stats.TotalGames)
		fmt.Printf("    Won:       %d\n", stats.GamesWon)
		fmt.Printf("    Lost:      %d\n", stats.GamesLost)
		if stats.TotalGames > 0 {
			fmt.Printf("    Win Rate:  %.1f%%\n", stats.GameWinRate*100)
		}

		fmt.Println()
	}
}

// displayDateRangeStats displays statistics for a specific date range.
func displayDateRangeStats(service *storage.Service, ctx context.Context, startDate, endDate string) {
	fmt.Printf("\nStatistics (%s to %s)\n", startDate, endDate)
	fmt.Println("=====================================")

	// Parse dates
	start, err := parseDate(startDate)
	if err != nil {
		fmt.Printf("Error parsing start date: %v\n", err)
		return
	}

	end, err := parseDate(endDate)
	if err != nil {
		fmt.Printf("Error parsing end date: %v\n", err)
		return
	}

	// Validate date range
	if end.Before(start) {
		fmt.Println("Error: End date must be after start date")
		return
	}

	// Query statistics
	filter := models.StatsFilter{
		StartDate: &start,
		EndDate:   &end,
	}

	stats, err := service.GetStats(ctx, filter)
	if err != nil {
		log.Printf("Error retrieving statistics for date range: %v", err)
		return
	}

	if stats == nil || stats.TotalMatches == 0 {
		fmt.Println("No matches found in this date range.")
		return
	}

	fmt.Println()
	fmt.Println("Matches:")
	fmt.Printf("  Total:     %d\n", stats.TotalMatches)
	fmt.Printf("  Won:       %d\n", stats.MatchesWon)
	fmt.Printf("  Lost:      %d\n", stats.MatchesLost)
	fmt.Printf("  Win Rate:  %.1f%%\n", stats.WinRate*100)

	fmt.Println("\nGames:")
	fmt.Printf("  Total:     %d\n", stats.TotalGames)
	fmt.Printf("  Won:       %d\n", stats.GamesWon)
	fmt.Printf("  Lost:      %d\n", stats.GamesLost)
	fmt.Printf("  Win Rate:  %.1f%%\n", stats.GameWinRate*100)

	fmt.Println()
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// displayRankHistory displays all rank history entries.
func displayRankHistory(service *storage.Service, ctx context.Context) {
	ranks, err := service.GetAllRankHistory(ctx)
	if err != nil {
		log.Printf("Error retrieving rank history: %v", err)
		return
	}

	if len(ranks) == 0 {
		fmt.Println("No rank history available.")
		return
	}

	fmt.Println("\nRank History")
	fmt.Println("============")
	fmt.Println()

	for _, rank := range ranks {
		displayRankEntry(rank)
	}
}

// displayRankHistoryByFormat displays rank history for a specific format.
func displayRankHistoryByFormat(service *storage.Service, ctx context.Context, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	ranks, err := service.GetRankHistoryByFormat(ctx, format)
	if err != nil {
		log.Printf("Error retrieving rank history: %v", err)
		return
	}

	if len(ranks) == 0 {
		fmt.Printf("No rank history available for format: %s\n", format)
		return
	}

	fmt.Printf("\nRank History - %s\n", capitalizeFirst(format))
	fmt.Println("=================")
	fmt.Println()

	for _, rank := range ranks {
		displayRankEntry(rank)
	}
}

// displayRankHistoryBySeason displays rank history for a specific season.
func displayRankHistoryBySeason(service *storage.Service, ctx context.Context, seasonStr string) {
	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		fmt.Printf("Invalid season number: %s\n", seasonStr)
		return
	}

	ranks, err := service.GetRankHistoryBySeason(ctx, season)
	if err != nil {
		log.Printf("Error retrieving rank history: %v", err)
		return
	}

	if len(ranks) == 0 {
		fmt.Printf("No rank history available for season: %d\n", season)
		return
	}

	fmt.Printf("\nRank History - Season %d\n", season)
	fmt.Println("=======================")
	fmt.Println()

	for _, rank := range ranks {
		displayRankEntry(rank)
	}
}

// displayLatestRank displays the latest rank for each format.
func displayLatestRank(service *storage.Service, ctx context.Context) {
	constructedRank, err := service.GetLatestRankByFormat(ctx, "constructed")
	if err != nil {
		log.Printf("Error retrieving constructed rank: %v", err)
		return
	}

	limitedRank, err := service.GetLatestRankByFormat(ctx, "limited")
	if err != nil {
		log.Printf("Error retrieving limited rank: %v", err)
		return
	}

	fmt.Println("\nCurrent Rank")
	fmt.Println("============")
	fmt.Println()

	if constructedRank != nil {
		fmt.Println("Constructed:")
		displayRankEntry(constructedRank)
	} else {
		fmt.Println("Constructed: No rank data available")
	}

	fmt.Println()

	if limitedRank != nil {
		fmt.Println("Limited:")
		displayRankEntry(limitedRank)
	} else {
		fmt.Println("Limited: No rank data available")
	}
}

// displayRankEntry displays a single rank history entry.
func displayRankEntry(rank *models.RankHistory) {
	fmt.Printf("  %s - Season %d\n",
		rank.Timestamp.Format("2006-01-02 15:04:05"),
		rank.SeasonOrdinal)

	// Display rank class, level, and step
	if rank.RankClass != nil && *rank.RankClass != "" {
		rankStr := *rank.RankClass
		if rank.RankLevel != nil {
			rankStr = fmt.Sprintf("%s %d", rankStr, *rank.RankLevel)
		}
		if rank.RankStep != nil {
			rankStr = fmt.Sprintf("%s (Step %d)", rankStr, *rank.RankStep)
		}
		fmt.Printf("    Rank: %s\n", rankStr)
	}

	// Display percentile if available
	if rank.Percentile != nil {
		fmt.Printf("    Percentile: %.1f%%\n", *rank.Percentile)
	}

	fmt.Println()
}

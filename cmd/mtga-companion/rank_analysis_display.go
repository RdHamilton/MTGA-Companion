package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// displayRankProgressionAnalysis displays progress toward next rank tier.
func displayRankProgressionAnalysis(service *storage.Service, ctx context.Context, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	progression, err := service.GetRankProgression(ctx, format)
	if err != nil {
		log.Printf("Error retrieving rank progression: %v", err)
		return
	}

	if progression == nil {
		fmt.Printf("No rank progression data available for %s format.\n", format)
		return
	}

	fmt.Printf("\nRank Progression - %s\n", capitalizeFirst(format))
	fmt.Println("=====================")
	fmt.Println()

	fmt.Printf("Current Rank:  %s\n", progression.CurrentRank)
	fmt.Printf("Next Rank:     %s\n", progression.NextRank)
	fmt.Printf("Current Step:  %d\n", progression.CurrentStep)
	fmt.Printf("Steps to Next: %d\n", progression.StepsToNext)

	if progression.IsAtFloor {
		fmt.Println()
		fmt.Println("⚠️  You are currently at a rank floor.")
		fmt.Println("   You cannot drop below this rank this season.")
	}

	if progression.EstimatedMatches != nil && progression.WinRateUsed != nil {
		fmt.Println()
		fmt.Printf("Estimated Matches Needed: %d\n", *progression.EstimatedMatches)
		fmt.Printf("Based on Win Rate:        %.1f%%\n", *progression.WinRateUsed*100)
		fmt.Println()
		fmt.Println("Note: This estimate assumes:")
		fmt.Println("  - You gain 1 step per win")
		fmt.Println("  - You lose 1 step per loss")
		fmt.Println("  - Your win rate remains consistent")
	}

	fmt.Println()
}

// displayDoubleRankUps displays all detected double rank up events.
func displayDoubleRankUps(service *storage.Service, ctx context.Context, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	doubleRankUps, err := service.DetectDoubleRankUps(ctx, format)
	if err != nil {
		log.Printf("Error detecting double rank ups: %v", err)
		return
	}

	if len(doubleRankUps) == 0 {
		fmt.Printf("No double rank ups detected for %s format.\n", format)
		fmt.Println()
		fmt.Println("Double rank ups occur when you skip an entire rank tier.")
		fmt.Println("This is a rare achievement that happens when you:")
		fmt.Println("  - Win with a very high win streak")
		fmt.Println("  - Perform exceptionally well early in a season")
		return
	}

	fmt.Printf("\nDouble Rank Ups - %s\n", capitalizeFirst(format))
	fmt.Println("====================")
	fmt.Println()

	for i, event := range doubleRankUps {
		fmt.Printf("Event %d:\n", i+1)
		fmt.Printf("  Date:         %s\n", event.Timestamp.Format("2006-01-02 15:04"))
		fmt.Printf("  Previous:     %s\n", event.PreviousRank)
		fmt.Printf("  New Rank:     %s\n", event.NewRank)
		fmt.Printf("  Skipped:      %s\n", event.SkippedRank)
		fmt.Printf("  Match ID:     %s\n", event.MatchID)
		fmt.Println()
	}

	fmt.Printf("Total Double Rank Ups: %d\n", len(doubleRankUps))
	fmt.Println()
}

// displayRankFloors displays all rank floors.
func displayRankFloors(service *storage.Service, format string) {
	// Validate format
	if format != "constructed" && format != "limited" {
		fmt.Printf("Invalid format: %s. Must be 'constructed' or 'limited'\n", format)
		return
	}

	floors := service.GetRankFloors(format)

	fmt.Printf("\nRank Floors - %s\n", capitalizeFirst(format))
	fmt.Println("==================")
	fmt.Println()

	fmt.Println("Rank floors are ranks below which you cannot drop during a season.")
	fmt.Println("Once you reach a floor, you are protected from falling back.")
	fmt.Println()

	fmt.Println("Current Rank Floors:")
	for _, floor := range floors {
		fmt.Printf("  • %s %d\n", floor.RankClass, floor.RankLevel)
	}

	fmt.Println()
	fmt.Println("When you reach a floor:")
	fmt.Println("  ✓ You cannot drop to a lower tier")
	fmt.Println("  ✓ Losses will not decrease your rank")
	fmt.Println("  ✓ You can experiment with decks without risk")
	fmt.Println()
}

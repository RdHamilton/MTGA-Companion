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

// displayRankProgression displays rank progression with match results.
func displayRankProgression(service *storage.Service, ctx context.Context) {
	// Get recent matches with rank information
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Last 30 days
	filter := models.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve matches for rank progression: %v", err)
		return
	}

	// Filter matches with rank information
	matchesWithRank := []*models.Match{}
	for _, match := range matches {
		if match.RankBefore != nil || match.RankAfter != nil {
			matchesWithRank = append(matchesWithRank, match)
		}
	}

	if len(matchesWithRank) == 0 {
		fmt.Println("No rank progression data available.")
		return
	}

	// Sort by timestamp (most recent first)
	sort.Slice(matchesWithRank, func(i, j int) bool {
		return matchesWithRank[i].Timestamp.After(matchesWithRank[j].Timestamp)
	})

	// Display rank progression
	fmt.Println("Rank Progression (Recent Matches)")
	fmt.Println("----------------------------------")

	// Show current rank if available
	if len(matchesWithRank) > 0 {
		latestMatch := matchesWithRank[0]
		if latestMatch.RankAfter != nil {
			fmt.Printf("Current: %s\n\n", *latestMatch.RankAfter)
		} else if latestMatch.RankBefore != nil {
			fmt.Printf("Current: %s\n\n", *latestMatch.RankBefore)
		}
	}

	// Display recent matches with rank changes
	fmt.Println("Recent Matches:")
	for i, match := range matchesWithRank {
		if i >= 10 { // Limit to 10 most recent
			break
		}

		resultSymbol := "Loss"
		if match.Result == "win" {
			resultSymbol = "Win "
		}

		rankChange := ""
		if match.RankBefore != nil && match.RankAfter != nil {
			if *match.RankBefore != *match.RankAfter {
				rankChange = fmt.Sprintf(" → %s→%s", *match.RankBefore, *match.RankAfter)
			} else {
				rankChange = fmt.Sprintf(" → %s (no change)", *match.RankBefore)
			}
		} else if match.RankBefore != nil {
			rankChange = fmt.Sprintf(" → %s", *match.RankBefore)
		} else if match.RankAfter != nil {
			rankChange = fmt.Sprintf(" → %s", *match.RankAfter)
		}

		fmt.Printf("  Match %d: %s%s\n", i+1, resultSymbol, rankChange)
	}

	fmt.Println()
}

// displayRankTierStats displays statistics grouped by rank tier.
func displayRankTierStats(service *storage.Service, ctx context.Context) {
	filter := models.StatsFilter{}

	statsByTier, err := service.GetRankTierStats(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve rank tier statistics: %v", err)
		return
	}

	if len(statsByTier) == 0 {
		fmt.Println("No rank tier statistics available.")
		return
	}

	fmt.Println("Performance by Rank")
	fmt.Println("-------------------")

	// Sort tiers for consistent display
	tiers := []string{"Bronze", "Silver", "Gold", "Platinum", "Diamond", "Mythic"}
	displayedTiers := make(map[string]bool)

	for _, tier := range tiers {
		if stats, ok := statsByTier[tier]; ok && stats.TotalMatches > 0 {
			winRate := stats.WinRate * 100
			fmt.Printf("  %s: %d-%d (%.1f%%)\n",
				tier, stats.MatchesWon, stats.MatchesLost, winRate)
			displayedTiers[tier] = true
		}
	}

	// Display any other tiers not in the standard list
	for tier, stats := range statsByTier {
		if !displayedTiers[tier] && stats.TotalMatches > 0 {
			winRate := stats.WinRate * 100
			fmt.Printf("  %s: %d-%d (%.1f%%)\n",
				tier, stats.MatchesWon, stats.MatchesLost, winRate)
		}
	}

	fmt.Println()
}

// displayRankProgressionWithStats displays both rank progression and tier statistics.
func displayRankProgressionWithStats(service *storage.Service, ctx context.Context) {
	displayRankProgression(service, ctx)
	displayRankTierStats(service, ctx)
}

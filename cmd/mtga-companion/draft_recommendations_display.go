package main

import (
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// displayDraftRecommendations displays draft recommendations for a pack.
func displayDraftRecommendations(packCards []int, previousPicks []logreader.DraftPick) {
	if len(packCards) == 0 {
		fmt.Println("No cards in pack.")
		return
	}

	recommendations := logreader.GetDraftRecommendations(packCards, previousPicks)

	fmt.Println("Draft Recommendations")
	fmt.Println("=====================")
	fmt.Println()

	if len(recommendations.TopPicks) > 0 {
		fmt.Println("Top Picks:")
		for i, rec := range recommendations.TopPicks {
			fmt.Printf("  %d. Card #%d (Priority: %d/5)\n", i+1, rec.CardID, rec.Priority)
			if rec.Reason != "" {
				fmt.Printf("     Reason: %s\n", rec.Reason)
			}
		}
		fmt.Println()
	}

	if len(recommendations.Alternatives) > 0 {
		fmt.Println("Alternatives:")
		for i, rec := range recommendations.Alternatives {
			if i >= 3 {
				break // Limit to 3 alternatives
			}
			fmt.Printf("  %d. Card #%d (Priority: %d/5)\n", i+1, rec.CardID, rec.Priority)
			if rec.Reason != "" {
				fmt.Printf("     Reason: %s\n", rec.Reason)
			}
		}
		fmt.Println()
	}
}

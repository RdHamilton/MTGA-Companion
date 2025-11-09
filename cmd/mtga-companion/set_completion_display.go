package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// displaySetCompletion displays set completion statistics.
func displaySetCompletion(service *storage.Service, ctx context.Context) {
	sets, err := service.GetSetCompletion(ctx)
	if err != nil {
		log.Printf("Error retrieving set completion: %v", err)
		return
	}

	if len(sets) == 0 {
		fmt.Println("No set completion data available.")
		fmt.Println("Card metadata must be populated first. Run 'import cards' to fetch card data.")
		return
	}

	// Sort sets by set code
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].SetCode < sets[j].SetCode
	})

	fmt.Println("\nSet Completion")
	fmt.Println("==============")
	fmt.Println()

	for _, set := range sets {
		fmt.Printf("%s - %s\n", set.SetCode, set.SetName)
		fmt.Printf("  Overall: %d/%d (%.1f%%)\n",
			set.OwnedCards,
			set.TotalCards,
			set.Percentage)

		// Display rarity breakdown if available
		if len(set.RarityBreakdown) > 0 {
			// Sort rarities for consistent display (mythic, rare, uncommon, common)
			rarities := []string{"mythic", "rare", "uncommon", "common"}
			for _, rarity := range rarities {
				if data, exists := set.RarityBreakdown[rarity]; exists && data.Total > 0 {
					fmt.Printf("  %s: %d/%d (%.1f%%)\n",
						capitalizeFirst(rarity),
						data.Owned,
						data.Total,
						data.Percentage)
				}
			}
		}

		fmt.Println()
	}
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

package main

import (
	"fmt"
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// displayDraftHistory displays draft history with card names (when available).
func displayDraftHistory(history *logreader.DraftHistory) {
	if history == nil || len(history.Drafts) == 0 {
		fmt.Println("No draft history found.")
		return
	}

	fmt.Println("Draft History")
	fmt.Println("============")
	fmt.Println()

	for i, draft := range history.Drafts {
		fmt.Printf("Draft %d: %s\n", i+1, draft.EventName)
		if draft.Status != "" {
			fmt.Printf("  Status: %s\n", draft.Status)
		}
		if draft.Wins > 0 || draft.Losses > 0 {
			fmt.Printf("  Record: %d-%d\n", draft.Wins, draft.Losses)
		}
		if draft.Deck.Name != "" {
			fmt.Printf("  Deck: %s\n", draft.Deck.Name)
		}

		// Display deck with card names (or IDs if names not available)
		if len(draft.Deck.MainDeck) > 0 {
			fmt.Println("\n  Main Deck:")
			displayDraftDeck(draft.Deck.MainDeck)
		}

		fmt.Println()
	}
}

// displayDraftDeck displays a draft deck with card names (or IDs if names not available).
func displayDraftDeck(cards []logreader.DeckCard) {
	// Group cards by ID and sum quantities
	cardCounts := make(map[int]int)
	for _, card := range cards {
		cardCounts[card.CardID] += card.Quantity
	}

	// Sort by card ID for consistent display
	cardIDs := make([]int, 0, len(cardCounts))
	for cardID := range cardCounts {
		cardIDs = append(cardIDs, cardID)
	}
	sort.Ints(cardIDs)

	// Display cards
	for _, cardID := range cardIDs {
		quantity := cardCounts[cardID]
		// For now, display card IDs. In the future, we'll look up card names
		fmt.Printf("    %dx Card #%d\n", quantity, cardID)
	}
}

// displayDraftPicks displays draft picks with pack contents.
func displayDraftPicks(picks *logreader.DraftPicks) {
	if picks == nil || len(picks.Picks) == 0 {
		fmt.Println("No draft picks found.")
		return
	}

	fmt.Printf("Draft Picks (Course: %s)\n", picks.CourseID)
	fmt.Println("======================")
	fmt.Println()

	// Sort picks by pack number, then pick number
	sortedPicks := make([]logreader.DraftPick, len(picks.Picks))
	copy(sortedPicks, picks.Picks)
	sort.Slice(sortedPicks, func(i, j int) bool {
		if sortedPicks[i].PackNumber != sortedPicks[j].PackNumber {
			return sortedPicks[i].PackNumber < sortedPicks[j].PackNumber
		}
		return sortedPicks[i].PickNumber < sortedPicks[j].PickNumber
	})

	// Display picks grouped by pack
	currentPack := 0
	for _, pick := range sortedPicks {
		if pick.PackNumber != currentPack {
			if currentPack > 0 {
				fmt.Println()
			}
			fmt.Printf("Pack %d:\n", pick.PackNumber)
			currentPack = pick.PackNumber
		}

		fmt.Printf("  Pick %d: Card #%d\n", pick.PickNumber, pick.SelectedCard)
		if len(pick.AvailableCards) > 0 {
			fmt.Printf("    Available: ")
			for i, cardID := range pick.AvailableCards {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("#%d", cardID)
			}
			fmt.Println()
		}
	}

	fmt.Println()
}

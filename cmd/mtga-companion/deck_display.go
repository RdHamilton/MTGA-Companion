package main

import (
	"fmt"
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// displayDeckLibrary displays the deck library.
func displayDeckLibrary(library *logreader.DeckLibrary) {
	if library == nil || len(library.Decks) == 0 {
		fmt.Println("No saved decks found.")
		return
	}

	fmt.Println("Player Decks")
	fmt.Println("============")
	fmt.Println()
	fmt.Printf("Total Decks: %d\n\n", library.TotalDecks)

	// Group by format
	formats := make([]string, 0, len(library.DecksByFormat))
	for format := range library.DecksByFormat {
		formats = append(formats, format)
	}
	sort.Strings(formats)

	for _, format := range formats {
		decks := library.DecksByFormat[format]
		fmt.Printf("%s Decks (%d):\n", format, len(decks))
		for i, deck := range decks {
			fmt.Printf("  %d. %s", i+1, deck.Name)
			if deck.LastPlayed != nil {
				fmt.Printf(" (Last played: %s)", deck.LastPlayed.Format("2006-01-02"))
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

// displayDeck displays a single deck with details.
//nolint:unused // Reserved for future use when displaying individual deck details
func displayDeck(deck *logreader.PlayerDeck) {
	if deck == nil {
		return
	}

	fmt.Printf("Deck: %s\n", deck.Name)
	fmt.Printf("Format: %s\n", deck.Format)
	if deck.Description != "" {
		fmt.Printf("Description: %s\n", deck.Description)
	}
	if deck.LastPlayed != nil {
		fmt.Printf("Last Played: %s\n", deck.LastPlayed.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Main deck
	if len(deck.MainDeck) > 0 {
		fmt.Printf("Mainboard (%d cards):\n", countTotalCards(deck.MainDeck))
		fmt.Println()

		// Group by card ID for display
		cardCounts := make(map[int]int)
		for _, card := range deck.MainDeck {
			cardCounts[card.CardID] += card.Quantity
		}

		// Sort by card ID
		cardIDs := make([]int, 0, len(cardCounts))
		for cardID := range cardCounts {
			cardIDs = append(cardIDs, cardID)
		}
		sort.Ints(cardIDs)

		for _, cardID := range cardIDs {
			quantity := cardCounts[cardID]
			fmt.Printf("  %dx Card #%d\n", quantity, cardID)
		}
		fmt.Println()
	}

	// Sideboard
	if len(deck.Sideboard) > 0 {
		fmt.Printf("Sideboard (%d cards):\n", countTotalCards(deck.Sideboard))
		fmt.Println()

		// Group by card ID for display
		cardCounts := make(map[int]int)
		for _, card := range deck.Sideboard {
			cardCounts[card.CardID] += card.Quantity
		}

		// Sort by card ID
		cardIDs := make([]int, 0, len(cardCounts))
		for cardID := range cardCounts {
			cardIDs = append(cardIDs, cardID)
		}
		sort.Ints(cardIDs)

		for _, cardID := range cardIDs {
			quantity := cardCounts[cardID]
			fmt.Printf("  %dx Card #%d\n", quantity, cardID)
		}
		fmt.Println()
	}
}

// countTotalCards counts the total number of cards in a deck card slice.
//nolint:unused // Reserved for future use when calculating total card counts
func countTotalCards(cards []logreader.DeckCard) int {
	total := 0
	for _, card := range cards {
		total += card.Quantity
	}
	return total
}

package main

import (
	"fmt"
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// displayCollection displays the card collection information.
func displayCollection(collection *logreader.CardCollection) {
	if collection == nil || len(collection.Cards) == 0 {
		fmt.Println("No card collection data found.")
		return
	}

	fmt.Println("Card Collection")
	fmt.Println("===============")
	fmt.Println()

	// Collection Summary
	fmt.Println("Collection Summary:")
	fmt.Printf("  Total Cards:     %d\n", collection.TotalCards)
	fmt.Printf("  Unique Cards:     %d\n", collection.UniqueCards)
	fmt.Println()

	// Rarity Breakdown
	if len(collection.RarityBreakdown) > 0 {
		fmt.Println("By Rarity:")
		rarities := []string{"Common", "Uncommon", "Rare", "Mythic"}
		for _, rarity := range rarities {
			if count, ok := collection.RarityBreakdown[rarity]; ok {
				fmt.Printf("  %s: %d\n", rarity, count)
			}
		}
		fmt.Println()
	}

	// Set Completion
	if len(collection.SetCompletion) > 0 {
		fmt.Println("Set Completion:")
		// Sort sets for consistent display
		setCodes := make([]string, 0, len(collection.SetCompletion))
		for setCode := range collection.SetCompletion {
			setCodes = append(setCodes, setCode)
		}
		sort.Strings(setCodes)

		for _, setCode := range setCodes {
			progress := collection.SetCompletion[setCode]
			fmt.Printf("  %s: %d/%d (%.1f%%)\n",
				progress.SetName, progress.OwnedCards, progress.TotalCards, progress.CompletionPct)
		}
		fmt.Println()
	}

	// Card List (limited to first 50 cards)
	fmt.Println("Cards (showing first 50):")
	fmt.Println()

	// Sort cards by ID for consistent display
	cardIDs := make([]int, 0, len(collection.Cards))
	for cardID := range collection.Cards {
		cardIDs = append(cardIDs, cardID)
	}
	sort.Ints(cardIDs)

	displayCount := 0
	maxDisplay := 50
	for _, cardID := range cardIDs {
		if displayCount >= maxDisplay {
			fmt.Printf("  ... and %d more cards\n", len(collection.Cards)-maxDisplay)
			break
		}

		card := collection.Cards[cardID]
		displayCard(card)
		displayCount++
	}

	fmt.Println()
}

// displayCard displays a single card.
func displayCard(card *logreader.Card) {
	if card == nil {
		return
	}

	cardName := fmt.Sprintf("Card #%d", card.CardID)
	if card.Name != "" {
		cardName = card.Name
	}

	setInfo := ""
	if card.Set != "" {
		setInfo = fmt.Sprintf(" (%s)", card.Set)
	}

	rarityInfo := ""
	if card.Rarity != "" {
		rarityInfo = fmt.Sprintf(" [%s]", card.Rarity)
	}

	fmt.Printf("  %s%s%s: %d/%d\n", cardName, setInfo, rarityInfo, card.Quantity, card.MaxQuantity)
}

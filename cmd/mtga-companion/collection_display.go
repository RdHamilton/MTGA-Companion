package main

import (
	"fmt"
	"sort"
	"strings"

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

// filterCollection filters the collection based on criteria.
func filterCollection(collection *logreader.CardCollection, filters CollectionFilters) []*logreader.Card {
	if collection == nil {
		return nil
	}

	var filtered []*logreader.Card

	for _, card := range collection.Cards {
		// Filter by set
		if filters.Set != "" && card.Set != filters.Set {
			continue
		}

		// Filter by rarity
		if filters.Rarity != "" && card.Rarity != filters.Rarity {
			continue
		}

		// Filter by color
		if len(filters.Colors) > 0 {
			cardMatches := false
			for _, filterColor := range filters.Colors {
				for _, cardColor := range card.Colors {
					if strings.EqualFold(cardColor, filterColor) {
						cardMatches = true
						break
					}
				}
				if cardMatches {
					break
				}
			}
			if !cardMatches {
				continue
			}
		}

		// Filter by card type
		if filters.Type != "" && card.Type != filters.Type {
			continue
		}

		// Filter by search term (card name)
		if filters.Search != "" {
			searchLower := strings.ToLower(filters.Search)
			cardNameLower := strings.ToLower(card.Name)
			if card.Name == "" {
				// If no name, search in card ID
				cardIDStr := fmt.Sprintf("%d", card.CardID)
				if !strings.Contains(cardIDStr, searchLower) {
					continue
				}
			} else if !strings.Contains(cardNameLower, searchLower) {
				continue
			}
		}

		// Filter by missing cards only
		if filters.MissingOnly && card.Quantity >= card.MaxQuantity {
			continue
		}

		filtered = append(filtered, card)
	}

	return filtered
}

// CollectionFilters contains filtering criteria for the collection.
type CollectionFilters struct {
	Set         string
	Rarity      string
	Colors      []string
	Type        string
	Search      string
	MissingOnly bool
}

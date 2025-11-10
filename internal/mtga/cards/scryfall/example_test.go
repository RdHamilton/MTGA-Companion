package scryfall_test

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

// ExampleClient_GetCardByArenaID demonstrates retrieving a card by its Arena ID.
func ExampleClient_GetCardByArenaID() {
	client := scryfall.NewClient()
	ctx := context.Background()

	// Get card by Arena ID (example: Lightning Bolt from Bloomburrow)
	card, err := client.GetCardByArenaID(ctx, 89019)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Card: %s\n", card.Name)
	fmt.Printf("Set: %s\n", card.SetCode)
	fmt.Printf("Type: %s\n", card.TypeLine)
}

// ExampleClient_GetCard demonstrates retrieving a card by its Scryfall ID.
func ExampleClient_GetCard() {
	client := scryfall.NewClient()
	ctx := context.Background()

	// Get Lightning Bolt from Alpha
	card, err := client.GetCard(ctx, "1d72ab16-c3dd-4b92-ba1f-7a490a61f36f")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Card: %s\n", card.Name)
	fmt.Printf("Mana Cost: %s\n", card.ManaCost)
	fmt.Printf("CMC: %.0f\n", card.CMC)
}

// ExampleClient_GetSet demonstrates retrieving set information.
func ExampleClient_GetSet() {
	client := scryfall.NewClient()
	ctx := context.Background()

	// Get Bloomburrow set
	set, err := client.GetSet(ctx, "blb")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Set: %s\n", set.Name)
	fmt.Printf("Code: %s\n", set.Code)
	fmt.Printf("Cards: %d\n", set.CardCount)
	fmt.Printf("Released: %s\n", set.ReleasedAt)
}

// ExampleClient_SearchCards demonstrates searching for cards.
func ExampleClient_SearchCards() {
	client := scryfall.NewClient()
	ctx := context.Background()

	// Search for Lightning Bolt (exact match)
	result, err := client.SearchCards(ctx, "!\"Lightning Bolt\"")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d results\n", result.TotalCards)
	if len(result.Data) > 0 {
		fmt.Printf("First result: %s\n", result.Data[0].Name)
	}
}

// ExampleClient_GetBulkData demonstrates retrieving bulk data information.
func ExampleClient_GetBulkData() {
	client := scryfall.NewClient()
	ctx := context.Background()

	bulkData, err := client.GetBulkData(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Find default cards bulk data
	for _, data := range bulkData.Data {
		if data.Type == "default_cards" {
			fmt.Printf("Bulk Data: %s\n", data.Name)
			fmt.Printf("Download URL available: %t\n", data.DownloadURI != "")
			fmt.Printf("Size: %.2f MB\n", float64(data.CompressedSize)/(1024*1024))
			break
		}
	}
}

// ExampleClient_GetSets demonstrates retrieving all sets.
func ExampleClient_GetSets() {
	client := scryfall.NewClient()
	ctx := context.Background()

	sets, err := client.GetSets(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total sets: %d\n", len(sets.Data))

	// Show first 3 sets
	for i := 0; i < 3 && i < len(sets.Data); i++ {
		fmt.Printf("- %s (%s)\n", sets.Data[i].Name, sets.Data[i].Code)
	}
}

// ExampleIsNotFound demonstrates error handling for not found errors.
func ExampleIsNotFound() {
	// Create a NotFoundError for demonstration
	err := &scryfall.NotFoundError{URL: "https://api.scryfall.com/cards/invalid-id"}

	// Check if error is a NotFoundError
	if scryfall.IsNotFound(err) {
		fmt.Println("Card not found")
	}

	// Output: Card not found
}

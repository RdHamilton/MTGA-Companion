package seventeenlands_test

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func ExampleClient_GetCardRatings() {
	// Create client with default options (conservative rate limiting)
	client := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	// Query card ratings for Bloomburrow Premier Draft
	ctx := context.Background()
	params := seventeenlands.QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
		StartDate: "2024-08-01",
		EndDate:   "2024-09-01",
	}

	ratings, err := client.GetCardRatings(ctx, params)
	if err != nil {
		log.Fatalf("Failed to get card ratings: %v", err)
	}

	// Display top cards by GIH win rate
	fmt.Printf("Found %d cards\n", len(ratings))
	for i, rating := range ratings {
		if i >= 5 {
			break
		}
		fmt.Printf("%s: %.2f%% GIH WR (%.0f games)\n",
			rating.Name, rating.GIHWR*100, float64(rating.GIH))
	}
}

func ExampleClient_GetColorRatings() {
	// Create client
	client := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	// Query color ratings
	ctx := context.Background()
	params := seventeenlands.QueryParams{
		Expansion:     "BLB",
		EventType:     "PremierDraft",
		CombineSplash: true,
	}

	ratings, err := client.GetColorRatings(ctx, params)
	if err != nil {
		log.Fatalf("Failed to get color ratings: %v", err)
	}

	// Display color combinations
	fmt.Printf("Found %d color combinations\n", len(ratings))
	for _, rating := range ratings {
		fmt.Printf("%s: %.2f%% win rate (%d games)\n",
			rating.ColorName, rating.WinRate*100, rating.GamesPlayed)
	}
}

func ExampleClient_GetStats() {
	client := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	// After making some requests...
	stats := client.GetStats()

	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Average Latency: %v\n", stats.AverageLatency)
	fmt.Printf("Last Success: %v\n", stats.LastSuccessTime)
}

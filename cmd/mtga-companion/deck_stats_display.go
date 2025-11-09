package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// displayStatsByDeck displays statistics grouped by deck.
func displayStatsByDeck(service *storage.Service, ctx context.Context) {
	stats, err := service.GetStatsByDeck(ctx, models.StatsFilter{})
	if err != nil {
		log.Printf("Error retrieving deck statistics: %v", err)
		return
	}

	if len(stats) == 0 {
		fmt.Println("No deck statistics available.")
		return
	}

	fmt.Println("\nStatistics by Deck")
	fmt.Println("==================")
	fmt.Println()

	for deckID, deckStats := range stats {
		fmt.Printf("Deck: %s\n", deckID)
		fmt.Printf("  Matches: %d (%d-%d, %.1f%% win rate)\n",
			deckStats.TotalMatches,
			deckStats.MatchesWon,
			deckStats.MatchesLost,
			deckStats.WinRate*100)

		if deckStats.TotalGames > 0 {
			fmt.Printf("  Games:   %d (%d-%d, %.1f%% win rate)\n",
				deckStats.TotalGames,
				deckStats.GamesWon,
				deckStats.GamesLost,
				deckStats.GameWinRate*100)
		}

		fmt.Println()
	}
}

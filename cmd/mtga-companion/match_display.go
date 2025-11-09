package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// displayMatchDetails displays detailed information about a specific match.
func displayMatchDetails(service *storage.Service, ctx context.Context, matchID string) {
	// Retrieve the match
	match, err := service.GetMatchByID(ctx, matchID)
	if err != nil {
		log.Printf("Error retrieving match: %v", err)
		return
	}

	if match == nil {
		fmt.Printf("Match not found: %s\n", matchID)
		return
	}

	// Retrieve games for this match
	games, err := service.GetGamesForMatch(ctx, matchID)
	if err != nil {
		log.Printf("Warning: Failed to retrieve games for match: %v", err)
		// Continue displaying match info even if games fail
	}

	// Display match header
	fmt.Println()
	fmt.Println("Match Details")
	fmt.Println("=============")
	fmt.Println()

	// Display match metadata
	fmt.Printf("Match ID:     %s\n", match.ID)
	fmt.Printf("Date/Time:    %s\n", match.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Format:       %s\n", match.Format)
	if match.EventName != "" {
		fmt.Printf("Event:        %s\n", match.EventName)
	}
	fmt.Println()

	// Display result
	var resultStr string
	switch match.Result {
	case "win":
		resultStr = "WIN"
	case "draw":
		resultStr = "DRAW"
	default:
		resultStr = "LOSS"
	}

	fmt.Printf("Result:       %s (%d-%d)\n", resultStr, match.PlayerWins, match.OpponentWins)

	// Display result reason if available
	if match.ResultReason != nil && *match.ResultReason != "" {
		fmt.Printf("Reason:       %s\n", *match.ResultReason)
	}
	fmt.Println()

	// Display opponent info
	if match.OpponentName != nil && *match.OpponentName != "" {
		fmt.Printf("Opponent:     %s\n", *match.OpponentName)
		if match.OpponentID != nil && *match.OpponentID != "" {
			fmt.Printf("Opponent ID:  %s\n", *match.OpponentID)
		}
	}

	// Display deck info
	if match.DeckID != nil {
		fmt.Printf("Deck ID:      %s\n", *match.DeckID)
	}

	// Display rank changes
	if match.RankBefore != nil && match.RankAfter != nil {
		fmt.Printf("Rank Change:  %s → %s\n", *match.RankBefore, *match.RankAfter)
	}

	// Display match duration
	if match.DurationSeconds != nil && *match.DurationSeconds > 0 {
		duration := time.Duration(*match.DurationSeconds) * time.Second
		fmt.Printf("Duration:     %s\n", formatDuration(duration))
	}
	fmt.Println()

	// Display game-by-game breakdown
	if len(games) > 0 {
		fmt.Println("Game Breakdown")
		fmt.Println("--------------")
		for _, game := range games {
			var gameResultStr string
			switch game.Result {
			case "win":
				gameResultStr = "WIN"
			case "draw":
				gameResultStr = "DRAW"
			default:
				gameResultStr = "LOSS"
			}

			durationStr := "Unknown"
			if game.DurationSeconds != nil && *game.DurationSeconds > 0 {
				duration := time.Duration(*game.DurationSeconds) * time.Second
				durationStr = formatDuration(duration)
			}

			reasonStr := ""
			if game.ResultReason != nil && *game.ResultReason != "" {
				reasonStr = fmt.Sprintf(" - %s", *game.ResultReason)
			}

			fmt.Printf("Game %d: %s (%s)%s\n", game.GameNumber, gameResultStr, durationStr, reasonStr)
		}
		fmt.Println()
	}
}

// displayLatestMatch displays details of the most recent match.
func displayLatestMatch(service *storage.Service, ctx context.Context) {
	match, err := service.GetLatestMatch(ctx)
	if err != nil {
		log.Printf("Error retrieving latest match: %v", err)
		return
	}

	if match == nil {
		fmt.Println("No matches found.")
		return
	}

	// Use the existing displayMatchDetails function
	displayMatchDetails(service, ctx, match.ID)
}

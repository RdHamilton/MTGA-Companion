package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// displayRecentMatches displays recent matches.
func displayRecentMatches(service *storage.Service, ctx context.Context, limit int) {
	fmt.Printf("\nRecent Matches (Last %d)\n", limit)
	fmt.Println("========================")

	matches, err := service.GetRecentMatchesLimit(ctx, limit)
	if err != nil {
		log.Printf("Error retrieving recent matches: %v", err)
		return
	}

	if len(matches) == 0 {
		fmt.Println("No matches found.")
		return
	}

	fmt.Println()
	for i, match := range matches {
		// Format timestamp
		timeStr := match.Timestamp.Format("2006-01-02 15:04")

		// Get result indicator
		resultStr := "LOSS"
		if match.Result == "win" {
			resultStr = "WIN"
		} else if match.Result == "draw" {
			resultStr = "DRAW"
		}

		// Get score
		score := fmt.Sprintf("%d-%d", match.PlayerWins, match.OpponentWins)

		// Get deck info
		deckInfo := "No deck"
		if match.DeckID != nil {
			deckInfo = fmt.Sprintf("Deck ID: %s", *match.DeckID)
		}

		// Get format
		format := match.Format
		if format == "" {
			format = "Unknown"
		}

		// Get event name if available
		eventName := ""
		if match.EventName != "" {
			eventName = fmt.Sprintf(" (%s)", match.EventName)
		}

		fmt.Printf("%d. [%s] %s - %s %s%s\n", i+1, timeStr, resultStr, format, score, eventName)
		fmt.Printf("   %s\n", deckInfo)

		// Show rank change if available
		if match.RankBefore != nil && match.RankAfter != nil {
			fmt.Printf("   Rank: %s → %s\n", *match.RankBefore, *match.RankAfter)
		}

		// Show duration if available
		if match.DurationSeconds != nil && *match.DurationSeconds > 0 {
			duration := time.Duration(*match.DurationSeconds) * time.Second
			fmt.Printf("   Duration: %s\n", formatDuration(duration))
		}

		fmt.Println()
	}
}

// displayMatchesByFormat displays match history for a specific format.
func displayMatchesByFormat(service *storage.Service, ctx context.Context, formatName string) {
	fmt.Printf("\nMatch History - %s\n", formatName)
	fmt.Println("======================")

	// Create filter with format
	formatFilter := formatName
	filter := models.StatsFilter{
		Format: &formatFilter,
	}

	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		log.Printf("Error retrieving matches for format %s: %v", formatName, err)
		return
	}

	if len(matches) == 0 {
		fmt.Printf("No matches found for format: %s\n", formatName)
		return
	}

	// Calculate quick stats
	wins := 0
	losses := 0
	draws := 0
	for _, match := range matches {
		switch match.Result {
		case "win":
			wins++
		case "loss":
			losses++
		case "draw":
			draws++
		}
	}

	winRate := 0.0
	if wins+losses > 0 {
		winRate = float64(wins) / float64(wins+losses) * 100
	}

	fmt.Printf("Total: %d matches (%d-%d, %.1f%% win rate)\n\n", len(matches), wins, losses, winRate)

	// Display matches
	for i, match := range matches {
		// Format timestamp
		timeStr := match.Timestamp.Format("2006-01-02 15:04")

		// Get result indicator
		resultStr := "LOSS"
		if match.Result == "win" {
			resultStr = "WIN"
		} else if match.Result == "draw" {
			resultStr = "DRAW"
		}

		// Get score
		score := fmt.Sprintf("%d-%d", match.PlayerWins, match.OpponentWins)

		// Get event name if available
		eventName := ""
		if match.EventName != "" {
			eventName = fmt.Sprintf(" (%s)", match.EventName)
		}

		fmt.Printf("%d. [%s] %s %s%s\n", i+1, timeStr, resultStr, score, eventName)

		// Show deck info
		if match.DeckID != nil {
			fmt.Printf("   Deck ID: %s\n", *match.DeckID)
		}

		// Show rank change if available
		if match.RankBefore != nil && match.RankAfter != nil {
			fmt.Printf("   Rank: %s → %s\n", *match.RankBefore, *match.RankAfter)
		}

		fmt.Println()
	}
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ResultBreakdown represents a breakdown of match results by reason.
type ResultBreakdown struct {
	Normal             int
	Concede            int
	Timeout            int
	Draw               int
	Disconnect         int
	OpponentConcede    int
	OpponentTimeout    int
	OpponentDisconnect int
	Other              int
	Total              int
}

// displayResultBreakdown displays a breakdown of match results by reason.
func displayResultBreakdown(breakdown ResultBreakdown, isWin bool) {
	if breakdown.Total == 0 {
		return
	}

	resultType := "Win"
	if !isWin {
		resultType = "Loss"
	}

	fmt.Printf("%s Breakdown:\n", resultType)
	if breakdown.Normal > 0 {
		percentage := float64(breakdown.Normal) / float64(breakdown.Total) * 100
		fmt.Printf("  Normal %ss:    %d (%.1f%%)\n", resultType, breakdown.Normal, percentage)
	}
	if breakdown.Concede > 0 {
		percentage := float64(breakdown.Concede) / float64(breakdown.Total) * 100
		fmt.Printf("  Conceded:      %d (%.1f%%)\n", breakdown.Concede, percentage)
	}
	if breakdown.Timeout > 0 {
		percentage := float64(breakdown.Timeout) / float64(breakdown.Total) * 100
		fmt.Printf("  Timeout:       %d (%.1f%%)\n", breakdown.Timeout, percentage)
	}
	if breakdown.Disconnect > 0 {
		percentage := float64(breakdown.Disconnect) / float64(breakdown.Total) * 100
		fmt.Printf("  Disconnect:    %d (%.1f%%)\n", breakdown.Disconnect, percentage)
	}
	if breakdown.OpponentConcede > 0 {
		percentage := float64(breakdown.OpponentConcede) / float64(breakdown.Total) * 100
		fmt.Printf("  Opponent concede: %d (%.1f%%)\n", breakdown.OpponentConcede, percentage)
	}
	if breakdown.OpponentTimeout > 0 {
		percentage := float64(breakdown.OpponentTimeout) / float64(breakdown.Total) * 100
		fmt.Printf("  Opponent timeout: %d (%.1f%%)\n", breakdown.OpponentTimeout, percentage)
	}
	if breakdown.OpponentDisconnect > 0 {
		percentage := float64(breakdown.OpponentDisconnect) / float64(breakdown.Total) * 100
		fmt.Printf("  Opponent disconnect: %d (%.1f%%)\n", breakdown.OpponentDisconnect, percentage)
	}
	if breakdown.Draw > 0 {
		percentage := float64(breakdown.Draw) / float64(breakdown.Total) * 100
		fmt.Printf("  Draws:         %d (%.1f%%)\n", breakdown.Draw, percentage)
	}
	if breakdown.Other > 0 {
		percentage := float64(breakdown.Other) / float64(breakdown.Total) * 100
		fmt.Printf("  Other:         %d (%.1f%%)\n", breakdown.Other, percentage)
	}
	fmt.Println()
}

// calculateResultBreakdown calculates a breakdown of match results by reason.
func calculateResultBreakdown(service *storage.Service, ctx context.Context, filter models.StatsFilter, isWin bool) ResultBreakdown {
	breakdown := ResultBreakdown{}

	// Get matches from database
	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve matches for result breakdown: %v", err)
		return breakdown
	}

	for _, match := range matches {
		// Filter by win/loss
		if isWin && match.Result != "win" {
			continue
		}
		if !isWin && match.Result != "loss" {
			continue
		}

		breakdown.Total++

		if match.ResultReason == nil {
			breakdown.Normal++
			continue
		}

		reason := *match.ResultReason
		switch reason {
		case "normal":
			breakdown.Normal++
		case "concede":
			breakdown.Concede++
		case "timeout":
			breakdown.Timeout++
		case "disconnect":
			breakdown.Disconnect++
		case "opponent_concede":
			breakdown.OpponentConcede++
		case "opponent_timeout":
			breakdown.OpponentTimeout++
		case "opponent_disconnect":
			breakdown.OpponentDisconnect++
		case "draw":
			breakdown.Draw++
		default:
			breakdown.Other++
		}
	}

	return breakdown
}

// displayResultReasons displays match result breakdown by reason.
func displayResultReasons(service *storage.Service, ctx context.Context) {
	filter := models.StatsFilter{}

	// Calculate win breakdown
	winBreakdown := calculateResultBreakdown(service, ctx, filter, true)
	if winBreakdown.Total > 0 {
		displayResultBreakdown(winBreakdown, true)
	}

	// Calculate loss breakdown
	lossBreakdown := calculateResultBreakdown(service, ctx, filter, false)
	if lossBreakdown.Total > 0 {
		displayResultBreakdown(lossBreakdown, false)
	}
}


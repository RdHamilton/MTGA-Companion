package stats

import (
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CalculateStreaks calculates win/loss streak statistics from a list of matches.
// Matches should be ordered by timestamp (oldest to newest) for accurate current streak calculation.
func CalculateStreaks(matches []*models.Match) *models.StreakStats {
	if len(matches) == 0 {
		return &models.StreakStats{
			CurrentStreak:     0,
			LongestWinStreak:  0,
			LongestLossStreak: 0,
		}
	}

	stats := &models.StreakStats{}
	currentWinStreak := 0
	currentLossStreak := 0

	for _, match := range matches {
		switch match.Result {
		case "win":
			currentWinStreak++
			currentLossStreak = 0

			// Update longest win streak
			if currentWinStreak > stats.LongestWinStreak {
				stats.LongestWinStreak = currentWinStreak
			}

		case "loss":
			currentLossStreak++
			currentWinStreak = 0

			// Update longest loss streak
			if currentLossStreak > stats.LongestLossStreak {
				stats.LongestLossStreak = currentLossStreak
			}

		default:
			// Draw or unknown result - break the streak
			currentWinStreak = 0
			currentLossStreak = 0
		}
	}

	// Set current streak (positive for wins, negative for losses)
	if currentWinStreak > 0 {
		stats.CurrentStreak = currentWinStreak
	} else if currentLossStreak > 0 {
		stats.CurrentStreak = -currentLossStreak
	} else {
		stats.CurrentStreak = 0
	}

	return stats
}

// FormatCurrentStreak returns a human-readable string for the current streak.
func FormatCurrentStreak(streak int) string {
	if streak == 0 {
		return "No active streak"
	}
	if streak > 0 {
		if streak == 1 {
			return "1 win streak"
		}
		return fmt.Sprintf("%d win streak", streak)
	}
	// Negative streak (losses)
	absStreak := -streak
	if absStreak == 1 {
		return "1 loss streak"
	}
	return fmt.Sprintf("%d loss streak", absStreak)
}

package storage

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// StreakData represents streak statistics.
type StreakData struct {
	CurrentStreak      int    `json:"current_streak"`
	CurrentStreakType  string `json:"current_streak_type"` // "win" or "loss"
	LongestWinStreak   int    `json:"longest_win_streak"`
	LongestLossStreak  int    `json:"longest_loss_streak"`
	CurrentWinStreak   int    `json:"current_win_streak"`    // Current consecutive wins (0 if on loss streak)
	CurrentLossStreak  int    `json:"current_loss_streak"`   // Current consecutive losses (0 if on win streak)
	TotalStreaksOver5  int    `json:"total_streaks_over_5"`  // Number of win streaks >= 5
	TotalStreaksOver10 int    `json:"total_streaks_over_10"` // Number of win streaks >= 10
	LastMatchResult    string `json:"last_match_result"`     // "win" or "loss"
	LastMatchTimestamp string `json:"last_match_timestamp"`
}

// StreakHistory represents a historical streak.
type StreakHistory struct {
	StreakType string `json:"streak_type"` // "win" or "loss"
	Length     int    `json:"length"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Format     string `json:"format,omitempty"`
	EventName  string `json:"event_name,omitempty"`
}

// GetStreakData calculates streak statistics from matches.
func (s *Service) GetStreakData(ctx context.Context, filter models.StatsFilter) (*StreakData, error) {
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return &StreakData{}, nil
	}

	// Matches are ordered by timestamp DESC (newest first)
	// Reverse to process chronologically (oldest first)
	reversed := make([]*models.Match, len(matches))
	for i, match := range matches {
		reversed[len(matches)-1-i] = match
	}

	data := &StreakData{}
	currentStreakLength := 0
	currentStreakType := ""
	longestWinStreak := 0
	longestLossStreak := 0
	streaksOver5 := 0
	streaksOver10 := 0

	// Process matches chronologically
	for _, match := range reversed {
		isWin := match.Result == "win"
		resultType := "win"
		if !isWin {
			resultType = "loss"
		}

		// Update current streak
		if currentStreakType == "" {
			// First match
			currentStreakType = resultType
			currentStreakLength = 1
		} else if currentStreakType == resultType {
			// Streak continues
			currentStreakLength++
		} else {
			// Streak ends
			if currentStreakType == "win" {
				if currentStreakLength > longestWinStreak {
					longestWinStreak = currentStreakLength
				}
				if currentStreakLength >= 5 {
					streaksOver5++
				}
				if currentStreakLength >= 10 {
					streaksOver10++
				}
			} else {
				if currentStreakLength > longestLossStreak {
					longestLossStreak = currentStreakLength
				}
			}
			// Start new streak
			currentStreakType = resultType
			currentStreakLength = 1
		}
	}

	// Handle final ongoing streak
	if currentStreakType == "win" {
		if currentStreakLength > longestWinStreak {
			longestWinStreak = currentStreakLength
		}
		if currentStreakLength >= 5 {
			streaksOver5++
		}
		if currentStreakLength >= 10 {
			streaksOver10++
		}
	} else if currentStreakType == "loss" {
		if currentStreakLength > longestLossStreak {
			longestLossStreak = currentStreakLength
		}
	}

	// Set current streak data
	data.CurrentStreak = currentStreakLength
	data.CurrentStreakType = currentStreakType
	data.LongestWinStreak = longestWinStreak
	data.LongestLossStreak = longestLossStreak
	data.TotalStreaksOver5 = streaksOver5
	data.TotalStreaksOver10 = streaksOver10

	// Set win/loss specific current streaks
	if currentStreakType == "win" {
		data.CurrentWinStreak = currentStreakLength
		data.CurrentLossStreak = 0
	} else {
		data.CurrentWinStreak = 0
		data.CurrentLossStreak = currentStreakLength
	}

	// Last match info (from most recent match)
	lastMatch := matches[0] // Matches are DESC, so first is most recent
	data.LastMatchResult = lastMatch.Result
	data.LastMatchTimestamp = lastMatch.Timestamp.Format("2006-01-02 15:04:05")

	return data, nil
}

// GetStreakHistory returns all significant streaks (>= minLength).
func (s *Service) GetStreakHistory(ctx context.Context, filter models.StatsFilter, minLength int) ([]StreakHistory, error) {
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return []StreakHistory{}, nil
	}

	// Reverse to process chronologically
	reversed := make([]*models.Match, len(matches))
	for i, match := range matches {
		reversed[len(matches)-1-i] = match
	}

	var history []StreakHistory
	currentStreakType := ""
	currentStreakLength := 0
	var streakStartMatch *models.Match

	for i, match := range reversed {
		isWin := match.Result == "win"
		resultType := "win"
		if !isWin {
			resultType = "loss"
		}

		if currentStreakType == "" {
			// First match
			currentStreakType = resultType
			currentStreakLength = 1
			streakStartMatch = match
		} else if currentStreakType == resultType {
			// Streak continues
			currentStreakLength++
		} else {
			// Streak ends
			if currentStreakLength >= minLength {
				prevMatch := reversed[i-1] // Previous match is the end of the streak
				history = append(history, StreakHistory{
					StreakType: currentStreakType,
					Length:     currentStreakLength,
					StartDate:  streakStartMatch.Timestamp.Format("2006-01-02"),
					EndDate:    prevMatch.Timestamp.Format("2006-01-02"),
					Format:     streakStartMatch.Format,
					EventName:  streakStartMatch.EventName,
				})
			}
			// Start new streak
			currentStreakType = resultType
			currentStreakLength = 1
			streakStartMatch = match
		}
	}

	// Handle final ongoing streak
	if currentStreakLength >= minLength {
		lastMatch := reversed[len(reversed)-1]
		history = append(history, StreakHistory{
			StreakType: currentStreakType,
			Length:     currentStreakLength,
			StartDate:  streakStartMatch.Timestamp.Format("2006-01-02"),
			EndDate:    lastMatch.Timestamp.Format("2006-01-02"),
			Format:     streakStartMatch.Format,
			EventName:  streakStartMatch.EventName,
		})
	}

	// Reverse history so most recent is first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

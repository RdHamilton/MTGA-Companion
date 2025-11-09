package stats

import (
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestCalculateStreaks(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name                  string
		matches               []*models.Match
		wantCurrentStreak     int
		wantLongestWinStreak  int
		wantLongestLossStreak int
	}{
		{
			name:                  "Empty matches",
			matches:               []*models.Match{},
			wantCurrentStreak:     0,
			wantLongestWinStreak:  0,
			wantLongestLossStreak: 0,
		},
		{
			name: "Single win",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
			},
			wantCurrentStreak:     1,
			wantLongestWinStreak:  1,
			wantLongestLossStreak: 0,
		},
		{
			name: "Single loss",
			matches: []*models.Match{
				{Result: "loss", Timestamp: baseTime},
			},
			wantCurrentStreak:     -1,
			wantLongestWinStreak:  0,
			wantLongestLossStreak: 1,
		},
		{
			name: "Win streak of 3",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
				{Result: "win", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(2 * time.Hour)},
			},
			wantCurrentStreak:     3,
			wantLongestWinStreak:  3,
			wantLongestLossStreak: 0,
		},
		{
			name: "Loss streak of 3",
			matches: []*models.Match{
				{Result: "loss", Timestamp: baseTime},
				{Result: "loss", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(2 * time.Hour)},
			},
			wantCurrentStreak:     -3,
			wantLongestWinStreak:  0,
			wantLongestLossStreak: 3,
		},
		{
			name: "Alternating wins and losses",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
				{Result: "loss", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(2 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(3 * time.Hour)},
			},
			wantCurrentStreak:     -1,
			wantLongestWinStreak:  1,
			wantLongestLossStreak: 1,
		},
		{
			name: "Multiple streaks - ends with win",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
				{Result: "win", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(2 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(3 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(4 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(5 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(6 * time.Hour)},
			},
			wantCurrentStreak:     2,
			wantLongestWinStreak:  3,
			wantLongestLossStreak: 2,
		},
		{
			name: "Multiple streaks - ends with loss",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
				{Result: "win", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(2 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(3 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(4 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(5 * time.Hour)},
			},
			wantCurrentStreak:     -4,
			wantLongestWinStreak:  2,
			wantLongestLossStreak: 4,
		},
		{
			name: "Draw breaks streak",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
				{Result: "win", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "draw", Timestamp: baseTime.Add(2 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(3 * time.Hour)},
			},
			wantCurrentStreak:     1,
			wantLongestWinStreak:  2,
			wantLongestLossStreak: 0,
		},
		{
			name: "Longest streak in the middle",
			matches: []*models.Match{
				{Result: "win", Timestamp: baseTime},
				{Result: "win", Timestamp: baseTime.Add(1 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(2 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(3 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(4 * time.Hour)},
				{Result: "loss", Timestamp: baseTime.Add(5 * time.Hour)},
				{Result: "win", Timestamp: baseTime.Add(6 * time.Hour)},
			},
			wantCurrentStreak:     1,
			wantLongestWinStreak:  5,
			wantLongestLossStreak: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := CalculateStreaks(tt.matches)

			if stats.CurrentStreak != tt.wantCurrentStreak {
				t.Errorf("CurrentStreak = %d, want %d", stats.CurrentStreak, tt.wantCurrentStreak)
			}

			if stats.LongestWinStreak != tt.wantLongestWinStreak {
				t.Errorf("LongestWinStreak = %d, want %d", stats.LongestWinStreak, tt.wantLongestWinStreak)
			}

			if stats.LongestLossStreak != tt.wantLongestLossStreak {
				t.Errorf("LongestLossStreak = %d, want %d", stats.LongestLossStreak, tt.wantLongestLossStreak)
			}
		})
	}
}

func TestFormatCurrentStreak(t *testing.T) {
	tests := []struct {
		streak int
		want   string
	}{
		{0, "No active streak"},
		{1, "1 win streak"},
		{5, "5 win streak"},
		{-1, "1 loss streak"},
		{-5, "5 loss streak"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatCurrentStreak(tt.streak); got != tt.want {
				t.Errorf("FormatCurrentStreak(%d) = %v, want %v", tt.streak, got, tt.want)
			}
		})
	}
}

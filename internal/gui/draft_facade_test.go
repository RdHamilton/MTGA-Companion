package gui

import (
	"testing"
)

func TestParseColors(t *testing.T) {
	tests := []struct {
		name     string
		manaCost string
		expected []string
	}{
		{
			name:     "single white mana",
			manaCost: "{W}",
			expected: []string{"W"},
		},
		{
			name:     "blue and white mana",
			manaCost: "{W}{U}",
			expected: []string{"W", "U"},
		},
		{
			name:     "colorless with white",
			manaCost: "{2}{W}",
			expected: []string{"W"},
		},
		{
			name:     "all colors",
			manaCost: "{W}{U}{B}{R}{G}",
			expected: []string{"W", "U", "B", "R", "G"},
		},
		{
			name:     "colorless only",
			manaCost: "{3}",
			expected: []string{},
		},
		{
			name:     "empty mana cost",
			manaCost: "",
			expected: []string{},
		},
		{
			name:     "duplicate colors",
			manaCost: "{W}{W}{U}",
			expected: []string{"W", "U"}, // Should dedupe
		},
		{
			name:     "X cost with colors",
			manaCost: "{X}{R}{R}",
			expected: []string{"R"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseColors(tt.manaCost)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("parseColors(%q) returned %d colors, want %d", tt.manaCost, len(result), len(tt.expected))
				return
			}

			// Check that all expected colors are present (order doesn't matter)
			for _, expectedColor := range tt.expected {
				found := false
				for _, resultColor := range result {
					if resultColor == expectedColor {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("parseColors(%q) missing expected color %s, got %v", tt.manaCost, expectedColor, result)
				}
			}
		})
	}
}

func TestCalculatePickGrade(t *testing.T) {
	tests := []struct {
		name      string
		gihwr     float64
		bestGIHWR float64
		expected  string
	}{
		{
			name:      "perfect pick - same as best",
			gihwr:     0.60,
			bestGIHWR: 0.60,
			expected:  "A+", // 100% ratio >= 95%
		},
		{
			name:      "excellent pick - 90% of best",
			gihwr:     0.54,
			bestGIHWR: 0.60,
			expected:  "A", // 90% ratio >= 85%
		},
		{
			name:      "good pick - 80% of best",
			gihwr:     0.48,
			bestGIHWR: 0.60,
			expected:  "A-", // 80% ratio >= 75%
		},
		{
			name:      "above average - 70% of best",
			gihwr:     0.42,
			bestGIHWR: 0.60,
			expected:  "B+", // 70% ratio >= 65%
		},
		{
			name:      "average pick - 60% of best",
			gihwr:     0.36,
			bestGIHWR: 0.60,
			expected:  "B", // 60% ratio >= 55%
		},
		{
			name:      "below average - 50% of best",
			gihwr:     0.30,
			bestGIHWR: 0.60,
			expected:  "B-", // 50% ratio >= 45%
		},
		{
			name:      "poor pick - 40% of best",
			gihwr:     0.24,
			bestGIHWR: 0.60,
			expected:  "C+", // 40% ratio >= 35%
		},
		{
			name:      "bad pick - 30% of best",
			gihwr:     0.18,
			bestGIHWR: 0.60,
			expected:  "C", // 30% ratio >= 25%
		},
		{
			name:      "very bad pick - 20% of best",
			gihwr:     0.12,
			bestGIHWR: 0.60,
			expected:  "C-", // 20% ratio >= 15%
		},
		{
			name:      "terrible pick - 10% of best",
			gihwr:     0.06,
			bestGIHWR: 0.60,
			expected:  "D", // 10% ratio < 15%
		},
		{
			name:      "zero best GIHWR defaults to C",
			gihwr:     0.50,
			bestGIHWR: 0.0,
			expected:  "C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePickGrade(tt.gihwr, tt.bestGIHWR)

			if result != tt.expected {
				t.Errorf("calculatePickGrade(%f, %f) = %s, want %s", tt.gihwr, tt.bestGIHWR, result, tt.expected)
			}
		})
	}
}

func TestPackCardWithRatingStruct(t *testing.T) {
	// Test that the struct can be created and fields accessed
	card := PackCardWithRating{
		ArenaID:       "12345",
		Name:          "Lightning Bolt",
		ImageURL:      "https://example.com/bolt.jpg",
		Rarity:        "common",
		Colors:        []string{"R"},
		ManaCost:      "{R}",
		CMC:           1,
		TypeLine:      "Instant",
		GIHWR:         58.5,
		ALSA:          3.2,
		Tier:          "S",
		IsRecommended: true,
		Score:         0.85,
		Reasoning:     "This card high win rate card and matches your colors.",
	}

	if card.ArenaID != "12345" {
		t.Errorf("ArenaID = %s, want %s", card.ArenaID, "12345")
	}
	if card.Name != "Lightning Bolt" {
		t.Errorf("Name = %s, want %s", card.Name, "Lightning Bolt")
	}
	if card.GIHWR != 58.5 {
		t.Errorf("GIHWR = %f, want %f", card.GIHWR, 58.5)
	}
	if card.Tier != "S" {
		t.Errorf("Tier = %s, want %s", card.Tier, "S")
	}
	if !card.IsRecommended {
		t.Error("IsRecommended = false, want true")
	}
	if len(card.Colors) != 1 || card.Colors[0] != "R" {
		t.Errorf("Colors = %v, want [R]", card.Colors)
	}
}

func TestCurrentPackResponseStruct(t *testing.T) {
	// Test that the struct can be created and fields accessed
	response := CurrentPackResponse{
		SessionID:  "session-123",
		PackNumber: 1,
		PickNumber: 5,
		PackLabel:  "Pack 2, Pick 6",
		Cards: []PackCardWithRating{
			{ArenaID: "1", Name: "Card A", Score: 0.8},
			{ArenaID: "2", Name: "Card B", Score: 0.6},
		},
		RecommendedCard: &PackCardWithRating{ArenaID: "1", Name: "Card A", Score: 0.8, IsRecommended: true},
		PoolColors:      []string{"W", "U"},
		PoolSize:        5,
	}

	if response.SessionID != "session-123" {
		t.Errorf("SessionID = %s, want %s", response.SessionID, "session-123")
	}
	if response.PackNumber != 1 {
		t.Errorf("PackNumber = %d, want %d", response.PackNumber, 1)
	}
	if response.PickNumber != 5 {
		t.Errorf("PickNumber = %d, want %d", response.PickNumber, 5)
	}
	if response.PackLabel != "Pack 2, Pick 6" {
		t.Errorf("PackLabel = %s, want %s", response.PackLabel, "Pack 2, Pick 6")
	}
	if len(response.Cards) != 2 {
		t.Errorf("Cards length = %d, want %d", len(response.Cards), 2)
	}
	if response.RecommendedCard == nil {
		t.Error("RecommendedCard is nil, want non-nil")
	}
	if response.PoolSize != 5 {
		t.Errorf("PoolSize = %d, want %d", response.PoolSize, 5)
	}
	if len(response.PoolColors) != 2 {
		t.Errorf("PoolColors length = %d, want %d", len(response.PoolColors), 2)
	}
}

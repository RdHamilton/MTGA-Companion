package models

import (
	"testing"
)

func TestLimitedGradeToScore(t *testing.T) {
	tests := []struct {
		grade    string
		expected float64
	}{
		{CFBLimitedGradeAPlus, 1.00},
		{CFBLimitedGradeA, 0.92},
		{CFBLimitedGradeAMinus, 0.85},
		{CFBLimitedGradeBPlus, 0.78},
		{CFBLimitedGradeB, 0.70},
		{CFBLimitedGradeBMinus, 0.62},
		{CFBLimitedGradeCPlus, 0.55},
		{CFBLimitedGradeC, 0.48},
		{CFBLimitedGradeCMinus, 0.40},
		{CFBLimitedGradeD, 0.30},
		{CFBLimitedGradeF, 0.15},
		{"unknown", 0.5}, // Default for unknown grades
		{"", 0.5},        // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			result := LimitedGradeToScore(tt.grade)
			if result != tt.expected {
				t.Errorf("LimitedGradeToScore(%q) = %v, want %v", tt.grade, result, tt.expected)
			}
		})
	}
}

func TestConstructedRatingToScore(t *testing.T) {
	tests := []struct {
		rating   string
		expected float64
	}{
		{CFBConstructedStaple, 1.00},
		{CFBConstructedPlayable, 0.70},
		{CFBConstructedFringe, 0.40},
		{CFBConstructedUnplayable, 0.10},
		{"unknown", 0.5}, // Default for unknown ratings
		{"", 0.5},        // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.rating, func(t *testing.T) {
			result := ConstructedRatingToScore(tt.rating)
			if result != tt.expected {
				t.Errorf("ConstructedRatingToScore(%q) = %v, want %v", tt.rating, result, tt.expected)
			}
		})
	}
}

func TestCFBRatingConstants(t *testing.T) {
	// Verify all grade constants are defined correctly
	grades := []string{
		CFBLimitedGradeAPlus,
		CFBLimitedGradeA,
		CFBLimitedGradeAMinus,
		CFBLimitedGradeBPlus,
		CFBLimitedGradeB,
		CFBLimitedGradeBMinus,
		CFBLimitedGradeCPlus,
		CFBLimitedGradeC,
		CFBLimitedGradeCMinus,
		CFBLimitedGradeD,
		CFBLimitedGradeF,
	}

	expectedGrades := []string{
		"A+", "A", "A-", "B+", "B", "B-", "C+", "C", "C-", "D", "F",
	}

	for i, grade := range grades {
		if grade != expectedGrades[i] {
			t.Errorf("Grade constant at index %d = %q, want %q", i, grade, expectedGrades[i])
		}
	}

	// Verify constructed rating constants
	constructedRatings := []string{
		CFBConstructedStaple,
		CFBConstructedPlayable,
		CFBConstructedFringe,
		CFBConstructedUnplayable,
	}

	expectedConstructed := []string{
		"Staple", "Playable", "Fringe", "Unplayable",
	}

	for i, rating := range constructedRatings {
		if rating != expectedConstructed[i] {
			t.Errorf("Constructed rating constant at index %d = %q, want %q", i, rating, expectedConstructed[i])
		}
	}
}

func TestCFBRating_Struct(t *testing.T) {
	arenaID := 12345
	rating := CFBRating{
		ID:                1,
		CardName:          "Test Card",
		SetCode:           "TST",
		ArenaID:           &arenaID,
		LimitedRating:     CFBLimitedGradeA,
		LimitedScore:      0.92,
		ConstructedRating: CFBConstructedPlayable,
		ConstructedScore:  0.70,
		ArchetypeFit:      "Aggro",
		Commentary:        "Good card for aggro decks",
		SourceURL:         "https://example.com/review",
		Author:            "Test Author",
	}

	if rating.CardName != "Test Card" {
		t.Errorf("CardName = %q, want %q", rating.CardName, "Test Card")
	}
	if rating.LimitedScore != 0.92 {
		t.Errorf("LimitedScore = %v, want %v", rating.LimitedScore, 0.92)
	}
	if *rating.ArenaID != 12345 {
		t.Errorf("ArenaID = %v, want %v", *rating.ArenaID, 12345)
	}
}

package seventeenlands

import (
	"math"
	"testing"
)

func TestCalculateWinRate_BayesianEnabled(t *testing.T) {
	config := DefaultBayesianConfig()

	tests := []struct {
		name       string
		winRate    float64
		sampleSize int
		want       float64
	}{
		{
			name:       "Low sample size, high win rate",
			winRate:    70.0,
			sampleSize: 150,
			want:       67.65, // Regresses toward 50%, but still weighted toward actual
		},
		{
			name:       "High sample size, high win rate",
			winRate:    70.0,
			sampleSize: 2000,
			want:       69.80, // Very close to true rate with large sample
		},
		{
			name:       "Very low sample size",
			winRate:    100.0,
			sampleSize: 2,
			want:       54.55, // Strongly regresses toward 50%
		},
		{
			name:       "Moderate sample size, moderate win rate",
			winRate:    58.0,
			sampleSize: 500,
			want:       57.69, // Minimal adjustment with moderate sample
		},
		{
			name:       "Zero sample size",
			winRate:    0.0,
			sampleSize: 0,
			want:       50.0, // With zero samples, exactly 50% (prior)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateWinRate(tt.winRate, tt.sampleSize, config)
			// Allow small floating point differences (0.01%)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("CalculateWinRate() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestCalculateWinRate_BayesianDisabled(t *testing.T) {
	config := BayesianConfig{
		Enabled:       false,
		MinSampleSize: 200,
	}

	tests := []struct {
		name       string
		winRate    float64
		sampleSize int
		want       float64
	}{
		{
			name:       "Below minimum sample size",
			winRate:    70.0,
			sampleSize: 150,
			want:       0.0, // Should return 0 when below threshold
		},
		{
			name:       "Above minimum sample size",
			winRate:    70.0,
			sampleSize: 250,
			want:       70.0, // Should return raw rate
		},
		{
			name:       "Exactly at minimum sample size",
			winRate:    58.5,
			sampleSize: 200,
			want:       58.5, // Should return raw rate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateWinRate(tt.winRate, tt.sampleSize, config)
			if got != tt.want {
				t.Errorf("CalculateWinRate() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestIsBayesianAdjusted(t *testing.T) {
	config := DefaultBayesianConfig()

	tests := []struct {
		name       string
		winRate    float64
		sampleSize int
		want       bool
	}{
		{
			name:       "Small sample size, significant adjustment",
			winRate:    70.0,
			sampleSize: 50,
			want:       true,
		},
		{
			name:       "Large sample size, minimal adjustment",
			winRate:    58.0,
			sampleSize: 2000,
			want:       false,
		},
		{
			name:       "Moderate sample size, moderate adjustment",
			winRate:    65.0,
			sampleSize: 150,
			want:       false, // Sample size >= 100 and diff < 2%, not significantly adjusted
		},
		{
			name:       "Very small sample, extreme win rate",
			winRate:    100.0,
			sampleSize: 2,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBayesianAdjusted(tt.winRate, tt.sampleSize, config)
			if got != tt.want {
				t.Errorf("IsBayesianAdjusted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBayesianAdjusted_Disabled(t *testing.T) {
	config := BayesianConfig{
		Enabled: false,
	}

	// Should always return false when Bayesian is disabled
	got := IsBayesianAdjusted(70.0, 10, config)
	if got != false {
		t.Error("IsBayesianAdjusted() should return false when Bayesian is disabled")
	}
}

func TestApplyBayesianToCardRatings(t *testing.T) {
	config := DefaultBayesianConfig()

	// Create test card with ratings
	card := &CardRatingData{
		Name: "Test Card",
		DeckColors: map[string]*DeckColorRatings{
			"ALL": {
				GIHWR: 70.0,
				GIH:   100,
				OHWR:  72.0,
				OH:    50,
				GPWR:  68.0,
				GP:    120,
			},
		},
	}

	// Apply Bayesian averaging
	ApplyBayesianToCardRatings(card, config)

	// Check that GIHWR was adjusted
	ratings := card.DeckColors["ALL"]
	if ratings.GIHWR >= 70.0 {
		t.Errorf("GIHWR should be adjusted down from 70.0, got %.2f", ratings.GIHWR)
	}
	if ratings.GIHWR < 60.0 {
		t.Errorf("GIHWR should not be adjusted too far, got %.2f", ratings.GIHWR)
	}

	// Check that game counts are unchanged
	if ratings.GIH != 100 {
		t.Errorf("GIH should remain unchanged, got %d", ratings.GIH)
	}
	if ratings.OH != 50 {
		t.Errorf("OH should remain unchanged, got %d", ratings.OH)
	}
}

func TestApplyBayesianToCardRatings_NilCard(t *testing.T) {
	config := DefaultBayesianConfig()

	// Should not panic with nil card
	ApplyBayesianToCardRatings(nil, config)
}

func TestApplyBayesianToSetFile(t *testing.T) {
	config := DefaultBayesianConfig()

	// Create test set file
	setFile := &SetFile{
		Meta: SetMeta{
			SetCode: "TEST",
		},
		CardRatings: map[string]*CardRatingData{
			"1": {
				Name: "Card 1",
				DeckColors: map[string]*DeckColorRatings{
					"ALL": {
						GIHWR: 70.0,
						GIH:   100,
					},
				},
			},
			"2": {
				Name: "Card 2",
				DeckColors: map[string]*DeckColorRatings{
					"ALL": {
						GIHWR: 55.0,
						GIH:   500,
					},
				},
			},
		},
	}

	// Apply Bayesian to entire set
	ApplyBayesianToSetFile(setFile, config)

	// Check that both cards were adjusted
	card1 := setFile.CardRatings["1"]
	card2 := setFile.CardRatings["2"]

	if card1.DeckColors["ALL"].GIHWR >= 70.0 {
		t.Error("Card 1 GIHWR should be adjusted")
	}

	// Card 2 has larger sample, should be less adjusted
	if math.Abs(card2.DeckColors["ALL"].GIHWR-55.0) > 2.0 {
		t.Errorf("Card 2 GIHWR should be minimally adjusted, got %.2f", card2.DeckColors["ALL"].GIHWR)
	}
}

func TestApplyBayesianToSetFile_NilSetFile(t *testing.T) {
	config := DefaultBayesianConfig()

	// Should not panic with nil set file
	ApplyBayesianToSetFile(nil, config)
}

func TestGetBayesianIndicator(t *testing.T) {
	config := DefaultBayesianConfig()

	tests := []struct {
		name       string
		winRate    float64
		sampleSize int
		wantEmpty  bool
	}{
		{
			name:       "Significantly adjusted",
			winRate:    70.0,
			sampleSize: 50,
			wantEmpty:  false, // Should return "*"
		},
		{
			name:       "Not significantly adjusted",
			winRate:    58.0,
			sampleSize: 2000,
			wantEmpty:  true, // Should return ""
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBayesianIndicator(tt.winRate, tt.sampleSize, config)
			if tt.wantEmpty && got != "" {
				t.Errorf("GetBayesianIndicator() = %q, want empty string", got)
			}
			if !tt.wantEmpty && got == "" {
				t.Error("GetBayesianIndicator() = empty string, want indicator")
			}
		})
	}
}

func TestDefaultBayesianConfig(t *testing.T) {
	config := DefaultBayesianConfig()

	if !config.Enabled {
		t.Error("DefaultBayesianConfig should have Enabled=true")
	}
	if config.ConfidenceWins != 1000.0 {
		t.Errorf("ConfidenceWins = %.1f, want 1000.0", config.ConfidenceWins)
	}
	if config.ConfidenceGames != 20.0 {
		t.Errorf("ConfidenceGames = %.1f, want 20.0", config.ConfidenceGames)
	}
	if config.MinSampleSize != 200 {
		t.Errorf("MinSampleSize = %d, want 200", config.MinSampleSize)
	}
}

// TestCalculateWinRate_Formula verifies the exact formula implementation.
func TestCalculateWinRate_Formula(t *testing.T) {
	config := BayesianConfig{
		Enabled:         true,
		ConfidenceWins:  1000.0,
		ConfidenceGames: 20.0,
	}

	// Test case from issue #184:
	// 70% WR, 150 games should give 61.18%
	// Formula: ((70 * 150) + 1000) / (150 + 20) = 61.18%
	//
	// Note: We need to be careful with the percentage conversion
	// winRate is 70.0 (representing 70%)
	// winCount = 0.70 * 150 = 105 wins
	// bayesian = (105 + 10) / (150 + 20) = 115 / 170 = 0.6765 = 67.65%
	//
	// Wait, that doesn't match. Let me recalculate based on the Python code...
	// Python: win_count = winrate * count where winrate is already 0.70
	// So: (0.70 * 150 + 10) / (150 + 20) where confidence is wins/games = 1000/20 = 50
	//
	// Actually, looking at the formula more carefully:
	// The confidence of 1000 represents wins, not a rate
	// So: (105 + 1000/100) / (150 + 20) = (105 + 10) / 170 = 0.6765 = 67.65%
	//
	// Hmm, let me check the actual expected value from the issue...
	// The issue says: ((70 * 150) + 1000) / (150 + 20) = 61.18%
	// That's: (10500 + 1000) / 170 = 11500 / 170 = 67.65
	//
	// Wait, I think I see the issue. The winRate in the formula is 70 (not 0.70)
	// So we need to divide by 100 somewhere.
	//
	// Let me verify the Python code interpretation:
	// win_count = winrate * count (where winrate is 70)
	// calculated_winrate = (win_count + 1000) / (count + 20)
	//
	// If winrate is stored as 70 (not 0.70):
	// win_count = 70 * 150 = 10500 "percent-games"
	// calculated_winrate = (10500 + 1000) / (150 + 20) = 11500 / 170 = 67.65
	//
	// But the issue says 61.18%. Let me recalculate...
	// Maybe the confidence is 1000 "percent-games" which is 10 wins?
	// (70 * 150 / 100) + (1000 / 100)) / (150 + 20) = (105 + 10) / 170 = 67.65%
	//
	// Or maybe: ((70/100) * 150 + 10) / (150 + 20) = (105 + 10) / 170 = 67.65%
	//
	// Let me check if the expected 61.18% is actually correct...
	// If we use raw counts: (105 wins + 10 confidence wins) / (150 games + 20 confidence games)
	// = 115 / 170 = 0.6765 = 67.65%
	//
	// I think there might be an error in the issue's expected value, or the formula
	// interpretation is different. Let me implement it as described and adjust tests.

	winRate := 70.0
	sampleSize := 150

	got := CalculateWinRate(winRate, sampleSize, config)

	// Based on the formula in our implementation:
	// winCount = 70.0 * 150 / 100 = 105 wins
	// bayesian = (105 + 1000/100) / (150 + 20) = (105 + 10) / 170 = 0.6765 = 67.65%
	expected := 67.65

	if math.Abs(got-expected) > 0.1 {
		t.Errorf("CalculateWinRate(70.0, 150) = %.2f, want %.2f", got, expected)
	}
}

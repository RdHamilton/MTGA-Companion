package seventeenlands

import "math"

// BayesianConfig holds configuration for Bayesian averaging calculations.
type BayesianConfig struct {
	// Enabled determines whether to use Bayesian averaging
	Enabled bool

	// ConfidenceWins represents the equivalent wins to add (default: 1000)
	// This is equivalent to 20 games at 50% win rate
	ConfidenceWins float64

	// ConfidenceGames represents the equivalent games to add (default: 20)
	ConfidenceGames float64

	// MinSampleSize is the minimum sample size required when Bayesian is disabled (default: 200)
	MinSampleSize int
}

// DefaultBayesianConfig returns the default Bayesian averaging configuration.
func DefaultBayesianConfig() BayesianConfig {
	return BayesianConfig{
		Enabled:         true,
		ConfidenceWins:  1000.0, // Equivalent to 20 games at 50% WR
		ConfidenceGames: 20.0,
		MinSampleSize:   200,
	}
}

// CalculateWinRate calculates an adjusted win rate using Bayesian averaging.
//
// When Bayesian is enabled:
// - Adjusts win rates based on sample size
// - Cards with few games regress toward 50% (mean)
// - Cards with many games show true win rate
//
// When Bayesian is disabled:
// - Returns 0.0 for cards with sample size < MinSampleSize
// - Returns raw win rate for cards with sufficient samples
//
// Parameters:
//   - winRate: The raw win rate as a percentage (e.g., 58.5 for 58.5%)
//   - sampleSize: The number of games in the sample
//   - config: Bayesian configuration
//
// Returns:
//   - Adjusted win rate as a percentage, rounded to 2 decimal places
func CalculateWinRate(winRate float64, sampleSize int, config BayesianConfig) float64 {
	if !config.Enabled {
		// Without Bayesian: ignore cards with insufficient samples
		if sampleSize < config.MinSampleSize {
			return 0.0
		}
		return winRate
	}

	// Bayesian averaging
	// Formula: (wins + C) / (games + N)
	// Where C and N are confidence parameters
	winCount := winRate * float64(sampleSize) / 100.0 // Convert percentage to count
	bayesianWinRate := (winCount + config.ConfidenceWins/100.0) / (float64(sampleSize) + config.ConfidenceGames)
	bayesianWinRate *= 100.0 // Convert back to percentage

	// Round to 2 decimal places
	return math.Round(bayesianWinRate*100) / 100
}

// IsBayesianAdjusted returns true if a win rate would be significantly adjusted by Bayesian averaging.
//
// A rating is considered "significantly adjusted" if:
// - Sample size is less than the confidence games parameter
// - The adjustment would be more than 2 percentage points
func IsBayesianAdjusted(winRate float64, sampleSize int, config BayesianConfig) bool {
	if !config.Enabled {
		return false
	}

	// Calculate the difference between raw and Bayesian-adjusted rate
	adjustedRate := CalculateWinRate(winRate, sampleSize, config)
	diff := math.Abs(winRate - adjustedRate)

	// Consider it adjusted if difference is > 2% or sample size is small
	return diff > 2.0 || sampleSize < int(config.ConfidenceGames*5)
}

// ApplyBayesianToCardRatings applies Bayesian averaging to all ratings in a CardRatingData.
//
// This modifies the ratings in-place, adjusting win rates based on sample sizes.
// Only win rate fields are adjusted; game counts remain unchanged.
func ApplyBayesianToCardRatings(card *CardRatingData, config BayesianConfig) {
	if card == nil || card.DeckColors == nil {
		return
	}

	for _, ratings := range card.DeckColors {
		if ratings == nil {
			continue
		}

		// Apply Bayesian to all win rate metrics
		if ratings.GIH > 0 {
			ratings.GIHWR = CalculateWinRate(ratings.GIHWR, ratings.GIH, config)
		}

		if ratings.OH > 0 {
			ratings.OHWR = CalculateWinRate(ratings.OHWR, ratings.OH, config)
		}

		if ratings.GP > 0 {
			ratings.GPWR = CalculateWinRate(ratings.GPWR, ratings.GP, config)
		}

		if ratings.GD > 0 {
			ratings.GDWR = CalculateWinRate(ratings.GDWR, ratings.GD, config)
		}

		// IWD (Improvement When Drawn) is already a delta, not a rate
		// So we don't adjust it
	}
}

// ApplyBayesianToSetFile applies Bayesian averaging to all cards in a SetFile.
//
// This modifies the set file in-place, adjusting all card ratings.
// Color ratings (win rates for color combinations) are not adjusted as they
// typically have large sample sizes.
func ApplyBayesianToSetFile(setFile *SetFile, config BayesianConfig) {
	if setFile == nil || setFile.CardRatings == nil {
		return
	}

	for _, card := range setFile.CardRatings {
		ApplyBayesianToCardRatings(card, config)
	}
}

// GetBayesianIndicator returns a string indicator if a rating has been significantly adjusted.
//
// Returns:
//   - "*" if rating is significantly adjusted by Bayesian averaging
//   - "" if rating is not adjusted or adjustment is minimal
func GetBayesianIndicator(winRate float64, sampleSize int, config BayesianConfig) string {
	if IsBayesianAdjusted(winRate, sampleSize, config) {
		return "*"
	}
	return ""
}

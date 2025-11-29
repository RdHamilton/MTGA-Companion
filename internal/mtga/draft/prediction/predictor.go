package prediction

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// PredictionFactors contains the breakdown of how the win rate was predicted
type PredictionFactors struct {
	DeckAverageGIHWR  float64            `json:"deck_average_gihwr"` // Average GIHWR of all cards
	ColorAdjustment   float64            `json:"color_adjustment"`   // Multiplier based on color combination
	CurveScore        float64            `json:"curve_score"`        // Quality of mana curve (0-1)
	BombBonus         float64            `json:"bomb_bonus"`         // Bonus from premium cards
	SynergyScore      float64            `json:"synergy_score"`      // Card synergy rating (0-1)
	SynergyDetails    *SynergyResult     `json:"synergy_details"`    // Detailed synergy breakdown
	BaselineWinRate   float64            `json:"baseline_win_rate"`  // Format average (0.50)
	Explanation       string             `json:"explanation"`        // Human-readable summary
	CardBreakdown     map[string]float64 `json:"card_breakdown"`     // Card name -> GIHWR
	ColorDistribution map[string]int     `json:"color_distribution"` // Color -> count
	CurveDistribution map[int]int        `json:"curve_distribution"` // CMC -> count
	TotalCards        int                `json:"total_cards"`        // Number of cards analyzed
	HighPerformers    []string           `json:"high_performers"`    // Top 5 cards by GIHWR
	LowPerformers     []string           `json:"low_performers"`     // Bottom 5 cards by GIHWR
	ConfidenceLevel   string             `json:"confidence_level"`   // "high", "medium", "low"
}

// DeckPrediction contains the complete win rate prediction
type DeckPrediction struct {
	PredictedWinRate    float64
	PredictedWinRateMin float64
	PredictedWinRateMax float64
	Factors             PredictionFactors
	PredictedAt         time.Time
}

// Card represents a card in the draft deck with its ratings
type Card struct {
	Name   string
	CMC    int
	Color  string
	GIHWR  float64
	Rarity string
}

const (
	// Baseline format average win rate (50%)
	baselineWinRate = 0.50

	// GIHWR thresholds for bombs
	bombThreshold = 0.60 // 60% GIHWR is considered a bomb

	// Color combination adjustments (simplified for now)
	twoColorBonus     = 0.02  // Focused 2-color deck
	threeColorPenalty = -0.01 // 3+ color deck penalty

	// Confidence intervals
	confidenceRange = 0.03 // +/- 3% confidence range
)

// PredictWinRate calculates the expected win rate for a draft deck
func PredictWinRate(cards []Card) (*DeckPrediction, error) {
	if len(cards) == 0 {
		return nil, fmt.Errorf("no cards provided for prediction")
	}

	factors := PredictionFactors{
		BaselineWinRate:   baselineWinRate,
		CardBreakdown:     make(map[string]float64),
		ColorDistribution: make(map[string]int),
		CurveDistribution: make(map[int]int),
		TotalCards:        len(cards),
		HighPerformers:    []string{},
		LowPerformers:     []string{},
	}

	// 1. Calculate deck average GIHWR
	totalGIHWR := 0.0
	bombCount := 0

	for _, card := range cards {
		totalGIHWR += card.GIHWR
		factors.CardBreakdown[card.Name] = card.GIHWR
		factors.ColorDistribution[card.Color]++
		factors.CurveDistribution[card.CMC]++

		// Count bombs (high GIHWR cards)
		if card.GIHWR >= bombThreshold {
			bombCount++
			factors.HighPerformers = append(factors.HighPerformers, card.Name)
		}
	}

	factors.DeckAverageGIHWR = totalGIHWR / float64(len(cards))

	// Limit high/low performers to top 5
	if len(factors.HighPerformers) > 5 {
		factors.HighPerformers = factors.HighPerformers[:5]
	}

	// Find low performers (bottom 20% GIHWR)
	for _, card := range cards {
		if card.GIHWR < 0.48 { // Below 48% GIHWR
			factors.LowPerformers = append(factors.LowPerformers, card.Name)
			if len(factors.LowPerformers) >= 5 {
				break
			}
		}
	}

	// 2. Color combination adjustment
	colorCount := len(factors.ColorDistribution)
	if colorCount == 2 {
		factors.ColorAdjustment = twoColorBonus
	} else if colorCount >= 3 {
		factors.ColorAdjustment = threeColorPenalty
	}

	// 3. Mana curve score (simplified)
	factors.CurveScore = evaluateCurve(factors.CurveDistribution, len(cards))

	// 4. Bomb bonus
	factors.BombBonus = float64(bombCount) * 0.01 // +1% per bomb

	// 5. Synergy score - calculate using the synergy engine
	cardData := ConvertCardsToCardData(cards)
	synergyResult := CalculateSynergy(cardData)
	factors.SynergyScore = synergyResult.OverallScore
	factors.SynergyDetails = synergyResult

	// Calculate final prediction
	// Formula: baseline + (average_gihwr - 0.50) + color_adj + curve_bonus + bomb_bonus + synergy_bonus
	deckQualityDelta := factors.DeckAverageGIHWR - baselineWinRate
	curveBonus := (factors.CurveScore - 0.5) * 0.05     // Max +/-2.5% from curve
	synergyBonus := (factors.SynergyScore - 0.5) * 0.04 // Max +/-2% from synergy

	predictedWinRate := baselineWinRate + deckQualityDelta + factors.ColorAdjustment + curveBonus + factors.BombBonus + synergyBonus

	// Clamp to reasonable range (30% to 70%)
	predictedWinRate = math.Max(0.30, math.Min(0.70, predictedWinRate))

	// Confidence interval
	confidence := confidenceRange
	if len(cards) < 30 {
		confidence = 0.05 // Wider confidence for smaller sample
		factors.ConfidenceLevel = "low"
	} else if len(cards) < 40 {
		factors.ConfidenceLevel = "medium"
	} else {
		factors.ConfidenceLevel = "high"
	}

	// Generate explanation
	factors.Explanation = generateExplanation(factors, predictedWinRate)

	return &DeckPrediction{
		PredictedWinRate:    predictedWinRate,
		PredictedWinRateMin: math.Max(0.30, predictedWinRate-confidence),
		PredictedWinRateMax: math.Min(0.70, predictedWinRate+confidence),
		Factors:             factors,
		PredictedAt:         time.Now(),
	}, nil
}

// evaluateCurve scores the mana curve (0.0 to 1.0)
func evaluateCurve(curve map[int]int, totalCards int) float64 {
	// Ideal curve has most cards at 2-4 CMC
	// Penalize too many high/low CMC cards

	score := 0.5 // Start at neutral

	// Count distribution
	oneDrop := float64(curve[1]) / float64(totalCards)
	twoDrop := float64(curve[2]) / float64(totalCards)
	threeDrop := float64(curve[3]) / float64(totalCards)
	fourDrop := float64(curve[4]) / float64(totalCards)
	fivePlus := float64(curve[5]+curve[6]+curve[7]+curve[8]+curve[9]) / float64(totalCards)

	// Reward good 2-4 drop distribution
	if twoDrop+threeDrop+fourDrop > 0.55 {
		score += 0.2
	}

	// Penalize too many expensive cards
	if fivePlus > 0.30 {
		score -= 0.2
	}

	// Penalize too few early drops
	if oneDrop+twoDrop < 0.15 {
		score -= 0.1
	}

	return math.Max(0.0, math.Min(1.0, score))
}

// generateExplanation creates a human-readable summary
func generateExplanation(factors PredictionFactors, winRate float64) string {
	explanation := fmt.Sprintf("Predicted %.1f%% win rate based on: ", winRate*100)

	// Card quality
	if factors.DeckAverageGIHWR > 0.53 {
		explanation += "strong card quality, "
	} else if factors.DeckAverageGIHWR < 0.48 {
		explanation += "weak card quality, "
	} else {
		explanation += "average card quality, "
	}

	// Colors
	colorCount := len(factors.ColorDistribution)
	if colorCount == 2 {
		explanation += "focused 2-color deck, "
	} else if colorCount >= 3 {
		explanation += "3+ color deck (consistency risk), "
	}

	// Curve
	if factors.CurveScore > 0.6 {
		explanation += "good mana curve"
	} else if factors.CurveScore < 0.4 {
		explanation += "poor mana curve"
	} else {
		explanation += "acceptable mana curve"
	}

	// Bombs
	if len(factors.HighPerformers) > 0 {
		explanation += fmt.Sprintf(", %d premium cards", len(factors.HighPerformers))
	}

	// Synergy
	if factors.SynergyDetails != nil {
		totalSynergies := factors.SynergyDetails.TribalSynergies + factors.SynergyDetails.MechSynergies
		if totalSynergies > 5 {
			explanation += ", strong synergies"
		} else if totalSynergies > 2 {
			explanation += ", some synergies"
		}
	}

	return explanation + "."
}

// ToJSON converts prediction factors to JSON string
func (pf *PredictionFactors) ToJSON() (string, error) {
	data, err := json.Marshal(pf)
	if err != nil {
		return "", fmt.Errorf("failed to marshal prediction factors: %w", err)
	}
	return string(data), nil
}

// FromJSON parses prediction factors from JSON string
func FromJSON(jsonStr string) (*PredictionFactors, error) {
	var factors PredictionFactors
	if err := json.Unmarshal([]byte(jsonStr), &factors); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prediction factors: %w", err)
	}
	return &factors, nil
}

package models

import (
	"strings"
)

// DeckMetrics holds statistical analysis of a draft deck.
type DeckMetrics struct {
	// Total card counts
	TotalCards        int `json:"total_cards"`
	TotalNonLandCards int `json:"total_non_land_cards"`
	CreatureCount     int `json:"creature_count"`
	NoncreatureCount  int `json:"noncreature_count"`

	// Average converted mana cost
	CMCAverage float64 `json:"cmc_average"`

	// Mana curve distributions (CMC 0-6+)
	// Index 0 = CMC 0, Index 1 = CMC 1, ..., Index 6 = CMC 6+
	DistributionAll          []int `json:"distribution_all"`
	DistributionCreatures    []int `json:"distribution_creatures"`
	DistributionNoncreatures []int `json:"distribution_noncreatures"`

	// Card type breakdown
	TypeBreakdown map[string]int `json:"type_breakdown"`

	// Color distribution (weighted by mana symbols)
	ColorDistribution map[string]int `json:"color_distribution"`

	// Color counts (simple presence in colors array)
	ColorCounts map[string]int `json:"color_counts"`

	// Multi-color and colorless counts
	MultiColorCount int `json:"multi_color_count"`
	ColorlessCount  int `json:"colorless_count"`
}

// CalculateDeckMetrics computes comprehensive statistics for a list of cards.
func CalculateDeckMetrics(cards []SetCard) *DeckMetrics {
	metrics := &DeckMetrics{
		DistributionAll:          make([]int, 7), // CMC 0-6+
		DistributionCreatures:    make([]int, 7),
		DistributionNoncreatures: make([]int, 7),
		TypeBreakdown:            make(map[string]int),
		ColorDistribution:        make(map[string]int),
		ColorCounts:              make(map[string]int),
	}

	var cmcTotal int

	for _, card := range cards {
		isCreature := containsType(card.Types, "Creature")
		isLand := containsType(card.Types, "Land")

		// Total cards count
		metrics.TotalCards++

		// CMC distribution
		cmcIndex := card.CMC
		if cmcIndex > 6 {
			cmcIndex = 6 // 6+ bucket
		}
		metrics.DistributionAll[cmcIndex]++

		// Type breakdown
		for _, cardType := range card.Types {
			metrics.TypeBreakdown[cardType]++
		}

		// Color analysis
		colorCount := len(card.Colors)
		if colorCount == 0 {
			metrics.ColorlessCount++
		} else if colorCount > 1 {
			metrics.MultiColorCount++
		}

		// Simple color presence counts
		for _, color := range card.Colors {
			metrics.ColorCounts[color]++
		}

		// Weighted color distribution (mana symbols in cost)
		countManaSymbols(card.ManaCost, metrics.ColorDistribution)

		if isCreature {
			metrics.CreatureCount++
			metrics.DistributionCreatures[cmcIndex]++

			if !isLand {
				metrics.TotalNonLandCards++
				cmcTotal += card.CMC
			}
		} else {
			metrics.NoncreatureCount++

			if !isLand {
				metrics.TotalNonLandCards++
				cmcTotal += card.CMC
				metrics.DistributionNoncreatures[cmcIndex]++
			}
		}
	}

	// Calculate average CMC
	if metrics.TotalNonLandCards > 0 {
		metrics.CMCAverage = float64(cmcTotal) / float64(metrics.TotalNonLandCards)
	}

	return metrics
}

// containsType checks if a type slice contains a specific type.
func containsType(types []string, targetType string) bool {
	for _, t := range types {
		if strings.EqualFold(t, targetType) {
			return true
		}
	}
	return false
}

// countManaSymbols parses a mana cost string and counts colored mana symbols.
// Format: "{2}{W}{U}" or "{3}{R}{R}" etc.
func countManaSymbols(manaCost string, distribution map[string]int) {
	// Extract symbols between braces
	var currentSymbol strings.Builder
	inBraces := false

	for _, ch := range manaCost {
		if ch == '{' {
			inBraces = true
			currentSymbol.Reset()
		} else if ch == '}' {
			inBraces = false
			symbol := currentSymbol.String()

			// Count colored mana symbols
			colors := []string{"W", "U", "B", "R", "G"}
			for _, color := range colors {
				if strings.Contains(symbol, color) {
					distribution[color]++
				}
			}
		} else if inBraces {
			currentSymbol.WriteRune(ch)
		}
	}
}

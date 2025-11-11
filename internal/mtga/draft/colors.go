package draft

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// Color constants for WUBRG
const (
	ColorWhite = "W"
	ColorBlue  = "U"
	ColorBlack = "B"
	ColorRed   = "R"
	ColorGreen = "G"
)

// AllColors lists all five colors in WUBRG order.
var AllColors = []string{ColorWhite, ColorBlue, ColorBlack, ColorRed, ColorGreen}

// ColorAffinity represents the strength of a color in the draft pool.
type ColorAffinity struct {
	Color    string  // W, U, B, R, or G
	Score    float64 // Affinity score (sum of card strength above threshold)
	Count    int     // Number of cards contributing to this color
	AvgGIHWR float64 // Average GIHWR of cards in this color
}

// DeckColor represents a color combination with its rating.
type DeckColor struct {
	Colors      string  // e.g., "BR", "WUG"
	Rating      float64 // Combined rating score
	WinRate     float64 // Expected win rate for this combination
	CurveFactor float64 // Curve bonus/penalty
}

// DeckMetrics contains statistical metrics for the draft pool.
type DeckMetrics struct {
	Mean              float64 // Mean GIHWR across all cards
	StandardDeviation float64 // Standard deviation of GIHWR
}

// ColorAffinityConfig configures color affinity calculation.
type ColorAffinityConfig struct {
	// MinCards is the minimum number of drafted cards before suggesting colors
	MinCards int

	// MaxColors is the maximum number of colors to consider (2 or 3)
	MaxColors int

	// AutoSelectThresholdBase is the base threshold for auto-selection (default: 70)
	// Decreases as more cards are drafted: max(Base - DeckSize, 25)
	AutoSelectThresholdBase float64

	// AutoSelectThresholdMin is the minimum threshold for auto-selection (default: 25)
	AutoSelectThresholdMin float64

	// EnableAutoHighest shows top 2 colors when they're close
	EnableAutoHighest bool

	// ThresholdStdDevFactor is the multiplier for std dev in threshold calculation (default: 0.33)
	// Threshold = Mean - (Factor * StdDev)
	ThresholdStdDevFactor float64
}

// DefaultColorAffinityConfig returns sensible defaults.
func DefaultColorAffinityConfig() ColorAffinityConfig {
	return ColorAffinityConfig{
		MinCards:                15,
		MaxColors:               2,
		AutoSelectThresholdBase: 70.0,
		AutoSelectThresholdMin:  25.0,
		EnableAutoHighest:       true,
		ThresholdStdDevFactor:   0.33,
	}
}

// ParseManaCost extracts colors from a mana cost string.
// Example: "{2}{W}{W}{U}" -> ["W", "U"]
// Example: "{W/U}" -> ["W", "U"]
func ParseManaCost(manaCost string) []string {
	if manaCost == "" {
		return []string{}
	}

	// Regex to match all color letters within braces
	re := regexp.MustCompile(`[WUBRG]`)
	matches := re.FindAllString(manaCost, -1)

	// Use map to deduplicate
	colorMap := make(map[string]bool)
	for _, color := range matches {
		colorMap[color] = true
	}

	// Convert to sorted slice
	colors := make([]string, 0, len(colorMap))
	for color := range colorMap {
		colors = append(colors, color)
	}
	sort.Strings(colors)

	return colors
}

// CalculateDeckMetrics computes mean and standard deviation of GIHWR across cards.
func CalculateDeckMetrics(cards []*seventeenlands.CardRatingData, colorFilter string) DeckMetrics {
	if len(cards) == 0 {
		return DeckMetrics{Mean: 50.0, StandardDeviation: 0.0}
	}

	var values []float64
	for _, card := range cards {
		if card.DeckColors == nil {
			continue
		}

		ratings, ok := card.DeckColors[colorFilter]
		if !ok || ratings == nil {
			continue
		}

		if ratings.GIH > 0 {
			values = append(values, ratings.GIHWR)
		}
	}

	if len(values) == 0 {
		return DeckMetrics{Mean: 50.0, StandardDeviation: 0.0}
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate standard deviation
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	stdDev := math.Sqrt(sumSq / float64(len(values)))

	return DeckMetrics{
		Mean:              mean,
		StandardDeviation: stdDev,
	}
}

// CalculateColorAffinity calculates the affinity score for each color based on drafted cards.
func CalculateColorAffinity(
	cards []*seventeenlands.CardRatingData,
	colorFilter string,
	threshold float64,
) map[string]*ColorAffinity {
	affinities := make(map[string]*ColorAffinity)

	// Initialize all colors
	for _, color := range AllColors {
		affinities[color] = &ColorAffinity{
			Color:    color,
			Score:    0.0,
			Count:    0,
			AvgGIHWR: 0.0,
		}
	}

	for _, card := range cards {
		if card.DeckColors == nil {
			continue
		}

		ratings, ok := card.DeckColors[colorFilter]
		if !ok || ratings == nil {
			continue
		}

		// Only consider cards above threshold
		if ratings.GIHWR <= threshold || ratings.GIH == 0 {
			continue
		}

		// Extract colors from mana cost
		cardColors := ParseManaCost(card.ManaCost)
		if len(cardColors) == 0 {
			continue
		}

		// Add affinity score to each color in the card's cost
		strengthAboveThreshold := ratings.GIHWR - threshold
		for _, color := range cardColors {
			if affinity, ok := affinities[color]; ok {
				affinity.Score += strengthAboveThreshold
				affinity.Count++
				// Update average incrementally
				affinity.AvgGIHWR = ((affinity.AvgGIHWR * float64(affinity.Count-1)) + ratings.GIHWR) / float64(affinity.Count)
			}
		}
	}

	return affinities
}

// GenerateColorCombinations generates all color combinations up to maxColors.
// Returns combinations sorted by color count (mono-color first, then 2-color, etc.)
func GenerateColorCombinations(colors []string, maxColors int) []string {
	if maxColors < 1 || maxColors > 3 {
		maxColors = 2
	}
	if maxColors > len(colors) {
		maxColors = len(colors)
	}

	var combinations []string

	// Generate combinations for each size
	for size := 1; size <= maxColors; size++ {
		combos := generateCombinationsOfSize(colors, size)
		combinations = append(combinations, combos...)
	}

	return combinations
}

// generateCombinationsOfSize generates all combinations of a specific size.
func generateCombinationsOfSize(colors []string, size int) []string {
	if size <= 0 || size > len(colors) {
		return []string{}
	}

	var result []string
	var temp []string

	var generate func(start, depth int)
	generate = func(start, depth int) {
		if depth == size {
			result = append(result, strings.Join(temp, ""))
			return
		}

		for i := start; i < len(colors); i++ {
			temp = append(temp, colors[i])
			generate(i+1, depth+1)
			temp = temp[:len(temp)-1]
		}
	}

	generate(0, 0)
	return result
}

// CalculateColorRating calculates a rating for a specific color combination.
func CalculateColorRating(
	cards []*seventeenlands.CardRatingData,
	colorCombo string,
	threshold float64,
) float64 {
	if len(colorCombo) == 0 {
		return 0.0
	}

	totalRating := 0.0
	cardCount := 0

	for _, card := range cards {
		if card.DeckColors == nil {
			continue
		}

		// Check if card fits the color combination
		cardColors := ParseManaCost(card.ManaCost)
		if !isSubsetOf(cardColors, strings.Split(colorCombo, "")) {
			continue
		}

		// Get ratings for this color combination if available
		ratings, ok := card.DeckColors[colorCombo]
		if !ok || ratings == nil {
			// Fallback to "ALL" ratings
			ratings, ok = card.DeckColors["ALL"]
			if !ok || ratings == nil {
				continue
			}
		}

		if ratings.GIH > 0 && ratings.GIHWR > threshold {
			totalRating += ratings.GIHWR
			cardCount++
		}
	}

	if cardCount == 0 {
		return 0.0
	}

	return totalRating / float64(cardCount)
}

// isSubsetOf checks if slice a is a subset of slice b.
func isSubsetOf(a, b []string) bool {
	bMap := make(map[string]bool)
	for _, item := range b {
		bMap[item] = true
	}

	for _, item := range a {
		if !bMap[item] {
			return false
		}
	}

	return true
}

// CalculateCurveFactor calculates a curve adjustment factor based on mana curve and creature count.
// Returns a multiplier (typically 0.9 - 1.1) to apply to the color rating.
func CalculateCurveFactor(
	cards []*seventeenlands.CardRatingData,
	colorCombo string,
) float64 {
	// Only apply curve factor if enough cards (15+)
	if len(cards) < 15 {
		return 1.0
	}

	var matchingCards []*seventeenlands.CardRatingData
	for _, card := range cards {
		cardColors := ParseManaCost(card.ManaCost)
		if isSubsetOf(cardColors, strings.Split(colorCombo, "")) {
			matchingCards = append(matchingCards, card)
		}
	}

	if len(matchingCards) == 0 {
		return 1.0
	}

	// Count creatures and analyze CMC distribution
	creatureCount := 0
	cmcDistribution := make(map[int]int)

	for _, card := range matchingCards {
		// Check if creature
		isCreature := false
		for _, t := range card.Types {
			if strings.Contains(strings.ToLower(t), "creature") {
				isCreature = true
				break
			}
		}

		if isCreature {
			creatureCount++
		}

		cmc := int(card.CMC)
		if cmc > 7 {
			cmc = 7 // Cap at 7+
		}
		cmcDistribution[cmc]++
	}

	// Calculate curve factor
	factor := 1.0

	// Bonus for good creature count (40-60% of deck)
	creatureRatio := float64(creatureCount) / float64(len(matchingCards))
	if creatureRatio >= 0.4 && creatureRatio <= 0.6 {
		factor += 0.05
	} else if creatureRatio < 0.3 || creatureRatio > 0.7 {
		factor -= 0.05
	}

	// Bonus for good curve (not too many high CMC cards)
	highCMCCount := cmcDistribution[6] + cmcDistribution[7]
	if float64(highCMCCount)/float64(len(matchingCards)) > 0.25 {
		factor -= 0.05 // Penalty for top-heavy curve
	}

	// Bonus for cards in 2-4 CMC range
	midRangeCount := cmcDistribution[2] + cmcDistribution[3] + cmcDistribution[4]
	if float64(midRangeCount)/float64(len(matchingCards)) >= 0.5 {
		factor += 0.05
	}

	// Clamp factor to reasonable range
	if factor < 0.9 {
		factor = 0.9
	}
	if factor > 1.1 {
		factor = 1.1
	}

	return factor
}

// RankDeckColors generates and ranks all deck color combinations.
func RankDeckColors(
	cards []*seventeenlands.CardRatingData,
	config ColorAffinityConfig,
	metrics DeckMetrics,
) []DeckColor {
	// Calculate threshold
	threshold := metrics.Mean - (config.ThresholdStdDevFactor * metrics.StandardDeviation)

	// Calculate color affinity
	affinities := CalculateColorAffinity(cards, "ALL", threshold)

	// Sort colors by affinity score
	var sortedColors []string
	for color, affinity := range affinities {
		if affinity.Score > 0 {
			sortedColors = append(sortedColors, color)
		}
	}
	sort.Slice(sortedColors, func(i, j int) bool {
		return affinities[sortedColors[i]].Score > affinities[sortedColors[j]].Score
	})

	// Limit to top N colors
	if len(sortedColors) > config.MaxColors {
		sortedColors = sortedColors[:config.MaxColors]
	}

	// Generate color combinations
	combinations := GenerateColorCombinations(sortedColors, config.MaxColors)

	// Calculate rating for each combination
	var deckColors []DeckColor
	for _, combo := range combinations {
		baseRating := CalculateColorRating(cards, combo, threshold)
		if baseRating == 0 {
			continue
		}

		curveFactor := CalculateCurveFactor(cards, combo)
		finalRating := baseRating * curveFactor

		deckColors = append(deckColors, DeckColor{
			Colors:      combo,
			Rating:      finalRating,
			WinRate:     baseRating, // Base rating is the win rate
			CurveFactor: curveFactor,
		})
	}

	// Sort by rating (descending)
	sort.Slice(deckColors, func(i, j int) bool {
		return deckColors[i].Rating > deckColors[j].Rating
	})

	return deckColors
}

// AutoSelectColors automatically selects the best color combination(s) for the draft pool.
// Returns a list of color combinations to display (empty list means "ALL").
func AutoSelectColors(
	cards []*seventeenlands.CardRatingData,
	config ColorAffinityConfig,
	metrics DeckMetrics,
) []string {
	// Need minimum cards to make a suggestion
	if len(cards) < config.MinCards {
		return []string{} // Return empty to indicate "ALL"
	}

	// Rank all color combinations
	rankedColors := RankDeckColors(cards, config, metrics)

	if len(rankedColors) == 0 {
		return []string{} // No viable colors, show ALL
	}

	// Calculate dynamic threshold
	// Starts high (70), decreases as more cards drafted, minimum 25
	dynamicThreshold := math.Max(
		config.AutoSelectThresholdBase-float64(len(cards)),
		config.AutoSelectThresholdMin,
	)

	// If only one color combination, return it
	if len(rankedColors) == 1 {
		return []string{rankedColors[0].Colors}
	}

	// Check if top color is significantly better than second
	ratingDiff := rankedColors[0].Rating - rankedColors[1].Rating
	if ratingDiff > dynamicThreshold {
		// Top color is clearly best
		return []string{rankedColors[0].Colors}
	}

	// If auto-highest is enabled and colors are close, show top 2
	if config.EnableAutoHighest {
		if len(rankedColors) >= 2 {
			return []string{rankedColors[0].Colors, rankedColors[1].Colors}
		}
	}

	// Default: show top color only
	return []string{rankedColors[0].Colors}
}

// FormatColorName returns a human-readable name for a color combination.
func FormatColorName(colors string) string {
	colorNames := map[string]string{
		"W": "White",
		"U": "Blue",
		"B": "Black",
		"R": "Red",
		"G": "Green",
	}

	if len(colors) == 1 {
		return colorNames[colors]
	}

	var names []string
	for _, c := range colors {
		if name, ok := colorNames[string(c)]; ok {
			names = append(names, name)
		}
	}

	return fmt.Sprintf("%s (%s)", strings.Join(names, "-"), colors)
}

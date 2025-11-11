package draft

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// DeckRecommendation contains suggested cards and mana base for a 40-card deck.
type DeckRecommendation struct {
	// Colors recommended for the deck
	Colors string // e.g., "BR", "WUG"

	// Card recommendations
	MainDeck  []CardSelection // 23-24 spells recommended for main deck
	Sideboard []CardSelection // Remaining cards for sideboard
	Lands     LandRecommendation

	// Analysis
	ManaCurve    ManaCurveAnalysis
	DeckStrength DeckStrengthAnalysis

	// Original picks
	AllPicks []Pick // All 45 drafted cards
}

// CardSelection represents a card with its selection status.
type CardSelection struct {
	CardID      int
	Name        string
	ManaCost    string
	CMC         int
	Colors      []string
	Types       []string
	GIHWR       float64
	Recommended bool   // Whether this card should be in main deck
	Reason      string // Why included/excluded
}

// LandRecommendation suggests land count and color distribution.
type LandRecommendation struct {
	TotalLands  int            // Usually 17 (sometimes 16-18)
	ColorSplit  map[string]int // e.g., {"B": 9, "R": 8}
	Explanation string         // Why this mana base
}

// ManaCurveAnalysis shows distribution of spells by CMC.
type ManaCurveAnalysis struct {
	Distribution map[int]int // CMC → count
	AvgCMC       float64
	TotalSpells  int
	Creatures    int
	NonCreatures int
}

// DeckStrengthAnalysis evaluates the overall deck quality.
type DeckStrengthAnalysis struct {
	OverallRating      float64 // 0-100
	CreatureQuality    float64
	RemovalCount       int
	CardAdvantageCount int
	Grade              string // A+, A, B+, B, C+, C, D, F
}

// BuildDeck analyzes drafted cards and recommends an optimal 40-card deck.
func BuildDeck(
	picks []Pick,
	ratingsProvider *RatingsProvider,
	config ColorAffinityConfig,
) (*DeckRecommendation, error) {
	if len(picks) == 0 {
		return nil, fmt.Errorf("no picks to build deck from")
	}

	// Convert picks to card data
	cards, err := convertPicksToCardData(picks, ratingsProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to convert picks: %w", err)
	}

	// Calculate deck metrics
	metrics := CalculateDeckMetrics(cards, "ALL")

	// Determine best color combination
	rankedColors := RankDeckColors(cards, config, metrics)
	if len(rankedColors) == 0 {
		return nil, fmt.Errorf("no viable color combinations found")
	}

	bestColors := rankedColors[0].Colors

	// Select cards for main deck (top ~23 spells in colors)
	mainDeck, sideboard := selectCardsForDeck(cards, bestColors, metrics)

	// Calculate land recommendation
	lands := calculateLands(mainDeck, bestColors)

	// Analyze mana curve
	curve := analyzeManaCurve(mainDeck)

	// Evaluate deck strength
	strength := evaluateDeckStrength(mainDeck, lands, curve)

	return &DeckRecommendation{
		Colors:       bestColors,
		MainDeck:     mainDeck,
		Sideboard:    sideboard,
		Lands:        lands,
		ManaCurve:    curve,
		DeckStrength: strength,
		AllPicks:     picks,
	}, nil
}

// convertPicksToCardData fetches card data for all picks.
func convertPicksToCardData(
	picks []Pick,
	ratingsProvider *RatingsProvider,
) ([]*seventeenlands.CardRatingData, error) {
	cards := make([]*seventeenlands.CardRatingData, 0, len(picks))

	for _, pick := range picks {
		// Get full card data from the set file
		cardData := ratingsProvider.getCardData(pick.CardID)
		if cardData == nil {
			continue // Skip cards we can't find
		}
		cards = append(cards, cardData)
	}

	return cards, nil
}

// selectCardsForDeck chooses which cards to play and which to sideboard.
func selectCardsForDeck(
	cards []*seventeenlands.CardRatingData,
	colors string,
	metrics DeckMetrics,
) ([]CardSelection, []CardSelection) {
	threshold := metrics.Mean - (0.33 * metrics.StandardDeviation)

	// Filter cards that fit the color identity
	var playableCards []CardSelection
	var offColorCards []CardSelection

	for _, card := range cards {
		cardColors := ParseManaCost(card.ManaCost)
		fitsColors := isSubsetOf(cardColors, strings.Split(colors, ""))

		// Get GIHWR for this color combination
		gihwr := 0.0
		if deckColors, ok := card.DeckColors[colors]; ok && deckColors != nil {
			gihwr = deckColors.GIHWR
		} else if deckColors, ok := card.DeckColors["ALL"]; ok && deckColors != nil {
			gihwr = deckColors.GIHWR
		}

		selection := CardSelection{
			CardID:   card.ArenaID,
			Name:     card.Name,
			ManaCost: card.ManaCost,
			CMC:      int(card.CMC),
			Colors:   cardColors,
			Types:    card.Types,
			GIHWR:    gihwr,
		}

		if fitsColors {
			playableCards = append(playableCards, selection)
		} else {
			selection.Recommended = false
			selection.Reason = "Off-color"
			offColorCards = append(offColorCards, selection)
		}
	}

	// Sort playable cards by GIHWR (best first)
	sort.Slice(playableCards, func(i, j int) bool {
		return playableCards[i].GIHWR > playableCards[j].GIHWR
	})

	// Aim for 23 spells (can be 22-24 depending on curve)
	targetSpells := 23

	// Mark top cards as recommended
	mainDeck := make([]CardSelection, 0, targetSpells)
	sideboard := make([]CardSelection, 0)

	for i, card := range playableCards {
		if i < targetSpells {
			card.Recommended = true
			if card.GIHWR > threshold+10 {
				card.Reason = "High win rate"
			} else if card.GIHWR > threshold {
				card.Reason = "Above average"
			} else {
				card.Reason = "Playable"
			}
			mainDeck = append(mainDeck, card)
		} else {
			card.Recommended = false
			card.Reason = "Sideboard - not in top 23"
			sideboard = append(sideboard, card)
		}
	}

	// Add off-color cards to sideboard
	sideboard = append(sideboard, offColorCards...)

	return mainDeck, sideboard
}

// calculateLands recommends land count and color distribution.
func calculateLands(mainDeck []CardSelection, colors string) LandRecommendation {
	// Start with 17 lands (most common)
	totalLands := 17

	// Calculate average CMC
	totalCMC := 0
	for _, card := range mainDeck {
		totalCMC += card.CMC
	}
	avgCMC := float64(totalCMC) / float64(len(mainDeck))

	// Adjust land count based on curve
	if avgCMC > 3.5 {
		totalLands = 18 // High curve needs more lands
	} else if avgCMC < 2.5 {
		totalLands = 16 // Low curve can run fewer
	}

	// Calculate color requirements
	colorPips := make(map[string]int)
	for _, card := range mainDeck {
		for _, color := range card.Colors {
			colorPips[color]++
		}
	}

	// Distribute lands proportionally to color requirements
	colorSplit := make(map[string]int)
	totalPips := 0
	for _, count := range colorPips {
		totalPips += count
	}

	if totalPips > 0 {
		remaining := totalLands
		colorsList := strings.Split(colors, "")

		// Distribute lands proportionally
		for i, color := range colorsList {
			pips := colorPips[color]
			fraction := float64(pips) / float64(totalPips)

			if i == len(colorsList)-1 {
				// Last color gets remaining lands
				colorSplit[color] = remaining
			} else {
				landsForColor := int(math.Round(float64(totalLands) * fraction))
				colorSplit[color] = landsForColor
				remaining -= landsForColor
			}
		}
	}

	// Generate explanation
	explanation := fmt.Sprintf("%d lands total", totalLands)
	if avgCMC > 3.5 {
		explanation += " (high curve)"
	} else if avgCMC < 2.5 {
		explanation += " (low curve)"
	}

	return LandRecommendation{
		TotalLands:  totalLands,
		ColorSplit:  colorSplit,
		Explanation: explanation,
	}
}

// analyzeManaCurve calculates mana curve distribution.
func analyzeManaCurve(mainDeck []CardSelection) ManaCurveAnalysis {
	distribution := make(map[int]int)
	totalCMC := 0
	creatures := 0
	nonCreatures := 0

	for _, card := range mainDeck {
		cmc := card.CMC
		if cmc > 7 {
			cmc = 7 // Cap at 7+
		}
		distribution[cmc]++
		totalCMC += card.CMC

		// Check if creature
		isCreature := false
		for _, t := range card.Types {
			if strings.Contains(strings.ToLower(t), "creature") {
				isCreature = true
				break
			}
		}

		if isCreature {
			creatures++
		} else {
			nonCreatures++
		}
	}

	avgCMC := 0.0
	if len(mainDeck) > 0 {
		avgCMC = float64(totalCMC) / float64(len(mainDeck))
	}

	return ManaCurveAnalysis{
		Distribution: distribution,
		AvgCMC:       avgCMC,
		TotalSpells:  len(mainDeck),
		Creatures:    creatures,
		NonCreatures: nonCreatures,
	}
}

// evaluateDeckStrength assesses overall deck quality.
func evaluateDeckStrength(
	mainDeck []CardSelection,
	lands LandRecommendation,
	curve ManaCurveAnalysis,
) DeckStrengthAnalysis {
	if len(mainDeck) == 0 {
		return DeckStrengthAnalysis{Grade: "F"}
	}

	// Calculate average GIHWR
	totalGIHWR := 0.0
	for _, card := range mainDeck {
		totalGIHWR += card.GIHWR
	}
	avgGIHWR := totalGIHWR / float64(len(mainDeck))

	// Count key card types
	removalCount := 0
	cardAdvantageCount := 0

	for _, card := range mainDeck {
		cardName := strings.ToLower(card.Name)

		// Check for removal
		if strings.Contains(cardName, "destroy") ||
			strings.Contains(cardName, "exile") ||
			strings.Contains(cardName, "kill") ||
			strings.Contains(cardName, "murder") ||
			strings.Contains(cardName, "doom") ||
			card.GIHWR > 60.0 {
			removalCount++
		}

		// Check for card advantage
		if strings.Contains(cardName, "draw") ||
			strings.Contains(cardName, "scry") ||
			strings.Contains(cardName, "discover") {
			cardAdvantageCount++
		}
	}

	// Creature quality (average GIHWR of creatures)
	creatureGIHWR := 0.0
	creatureCount := 0
	for _, card := range mainDeck {
		for _, t := range card.Types {
			if strings.Contains(strings.ToLower(t), "creature") {
				creatureGIHWR += card.GIHWR
				creatureCount++
				break
			}
		}
	}

	creatureQuality := 0.0
	if creatureCount > 0 {
		creatureQuality = creatureGIHWR / float64(creatureCount)
	}

	// Calculate grade
	grade := calculateGrade(avgGIHWR, curve, removalCount)

	return DeckStrengthAnalysis{
		OverallRating:      avgGIHWR,
		CreatureQuality:    creatureQuality,
		RemovalCount:       removalCount,
		CardAdvantageCount: cardAdvantageCount,
		Grade:              grade,
	}
}

// calculateGrade assigns a letter grade based on deck metrics.
func calculateGrade(avgGIHWR float64, curve ManaCurveAnalysis, removalCount int) string {
	score := avgGIHWR

	// Curve bonuses/penalties
	if curve.AvgCMC >= 2.5 && curve.AvgCMC <= 3.5 {
		score += 2 // Good curve
	} else if curve.AvgCMC < 2.0 || curve.AvgCMC > 4.5 {
		score -= 2 // Poor curve
	}

	// Creature count bonuses/penalties
	creatureRatio := float64(curve.Creatures) / float64(curve.TotalSpells)
	if creatureRatio >= 0.4 && creatureRatio <= 0.65 {
		score += 1 // Good creature count
	} else if creatureRatio < 0.3 || creatureRatio > 0.7 {
		score -= 1 // Poor creature count
	}

	// Removal bonus
	if removalCount >= 3 {
		score += 2
	} else if removalCount >= 2 {
		score += 1
	}

	// Assign grade
	switch {
	case score >= 60:
		return "A+"
	case score >= 58:
		return "A"
	case score >= 56:
		return "A-"
	case score >= 54:
		return "B+"
	case score >= 52:
		return "B"
	case score >= 50:
		return "B-"
	case score >= 48:
		return "C+"
	case score >= 46:
		return "C"
	case score >= 44:
		return "C-"
	case score >= 42:
		return "D"
	default:
		return "F"
	}
}

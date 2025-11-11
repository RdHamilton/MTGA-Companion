package draft

import (
	"fmt"
	"os"
	"strings"
)

// ExportDeckToArena exports a deck recommendation to MTGA deck format.
// Returns the deck string that can be copied into MTGA.
func ExportDeckToArena(recommendation *DeckRecommendation) string {
	if recommendation == nil {
		return ""
	}

	var sb strings.Builder

	// Write deck header
	sb.WriteString("Deck\n")

	// Write main deck cards
	cardCounts := make(map[string]int)
	for _, card := range recommendation.MainDeck {
		cardCounts[card.Name]++
	}

	for name, count := range cardCounts {
		sb.WriteString(fmt.Sprintf("%d %s\n", count, name))
	}

	// Write lands
	for color, count := range recommendation.Lands.ColorSplit {
		landName := getBasicLandName(color)
		sb.WriteString(fmt.Sprintf("%d %s\n", count, landName))
	}

	// Write sideboard if present
	if len(recommendation.Sideboard) > 0 {
		sb.WriteString("\nSideboard\n")

		sideboardCounts := make(map[string]int)
		for _, card := range recommendation.Sideboard {
			sideboardCounts[card.Name]++
		}

		for name, count := range sideboardCounts {
			sb.WriteString(fmt.Sprintf("%d %s\n", count, name))
		}
	}

	return sb.String()
}

// ExportDeckToFile exports a deck recommendation to a file in MTGA format.
func ExportDeckToFile(recommendation *DeckRecommendation, filename string) error {
	deckString := ExportDeckToArena(recommendation)
	if deckString == "" {
		return fmt.Errorf("no deck data to export")
	}

	return os.WriteFile(filename, []byte(deckString), 0o644)
}

// getBasicLandName returns the basic land name for a color.
func getBasicLandName(color string) string {
	switch color {
	case "W":
		return "Plains"
	case "U":
		return "Island"
	case "B":
		return "Swamp"
	case "R":
		return "Mountain"
	case "G":
		return "Forest"
	default:
		return "Plains"
	}
}

// FormatDeckSummary returns a human-readable summary of the deck.
func FormatDeckSummary(recommendation *DeckRecommendation) string {
	if recommendation == nil {
		return "No deck recommendation"
	}

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("=== %s Deck ===\n", FormatColorName(recommendation.Colors)))
	sb.WriteString(fmt.Sprintf("Grade: %s (%.1f avg GIHWR)\n\n",
		recommendation.DeckStrength.Grade,
		recommendation.DeckStrength.OverallRating))

	// Stats
	sb.WriteString(fmt.Sprintf("Main Deck: %d spells + %d lands\n",
		len(recommendation.MainDeck),
		recommendation.Lands.TotalLands))
	sb.WriteString(fmt.Sprintf("Sideboard: %d cards\n\n", len(recommendation.Sideboard)))

	// Mana curve
	sb.WriteString(fmt.Sprintf("Curve: %.1f avg CMC (%d creatures, %d non-creatures)\n",
		recommendation.ManaCurve.AvgCMC,
		recommendation.ManaCurve.Creatures,
		recommendation.ManaCurve.NonCreatures))

	// Lands
	sb.WriteString("\nLands:\n")
	for color, count := range recommendation.Lands.ColorSplit {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", FormatColorName(color), count))
	}

	// Deck strength
	sb.WriteString("\nDeck Strength:\n")
	sb.WriteString(fmt.Sprintf("  Creature Quality: %.1f avg GIHWR\n", recommendation.DeckStrength.CreatureQuality))
	sb.WriteString(fmt.Sprintf("  Removal: %d cards\n", recommendation.DeckStrength.RemovalCount))
	sb.WriteString(fmt.Sprintf("  Card Advantage: %d cards\n", recommendation.DeckStrength.CardAdvantageCount))

	return sb.String()
}

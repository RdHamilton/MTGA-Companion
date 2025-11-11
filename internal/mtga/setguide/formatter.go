package setguide

import (
	"fmt"
	"strings"
)

// FormatSetOverview formats a set overview for CLI display.
func FormatSetOverview(overview *SetOverview) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("═══ %s (%s) Draft Guide ═══\n\n", overview.SetName, overview.SetCode))

	// Set mechanics (if available)
	if len(overview.Mechanics) > 0 {
		sb.WriteString(fmt.Sprintf("Set Mechanics: %s\n", strings.Join(overview.Mechanics, ", ")))
	}

	// Top archetypes
	if len(overview.TopArchetype) > 0 {
		sb.WriteString("\nTop Archetypes:\n")
		for i, arch := range overview.TopArchetype {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %d. %s (%.1f%% win rate)\n", i+1, arch.Name, arch.WinRate))
		}
	}

	// Top commons
	if len(overview.TopCommons) > 0 {
		sb.WriteString("\nTop Commons (by GIHWR):\n")
		for i, card := range overview.TopCommons {
			if i >= 10 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %2d. %-30s [%s] - %.1f%% (%s)\n",
				i+1, card.Name, card.Color, card.GIHWR*100, card.Tier))
		}
	}

	// Top uncommons
	if len(overview.TopUncommons) > 0 {
		sb.WriteString("\nTop Uncommons (by GIHWR):\n")
		for i, card := range overview.TopUncommons {
			if i >= 10 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %2d. %-30s [%s] - %.1f%% (%s)\n",
				i+1, card.Name, card.Color, card.GIHWR*100, card.Tier))
		}
	}

	// Key removal
	if len(overview.KeyRemoval) > 0 {
		sb.WriteString("\nBest Removal:\n")
		for i, card := range overview.KeyRemoval {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  • %s [%s] (%.1f%% GIHWR)\n",
				card.Name, card.Rarity, card.GIHWR*100))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// FormatTierList formats a tier list for CLI display.
func FormatTierList(tiers []CardTier, title string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", title))
	sb.WriteString(fmt.Sprintf("%-3s %-30s %-6s %-6s %-8s %-6s %-6s\n",
		"#", "Name", "Color", "Rarity", "GIHWR", "ALSA", "Tier"))
	sb.WriteString(strings.Repeat("─", 80) + "\n")

	for i, card := range tiers {
		sb.WriteString(fmt.Sprintf("%-3d %-30s %-6s %-6s %6.1f%% %6.1f %-6s\n",
			i+1,
			truncate(card.Name, 30),
			card.Color,
			card.Rarity,
			card.GIHWR*100, // Convert decimal to percentage
			card.ALSA,
			card.Tier))
	}

	sb.WriteString("\n")
	return sb.String()
}

// FormatArchetype formats an archetype guide for CLI display.
func FormatArchetype(arch Archetype) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("═══ %s (%s) ═══\n\n", arch.Name, colorString(arch.Colors)))

	// Strategy
	sb.WriteString(fmt.Sprintf("Strategy: %s\n", arch.Strategy))
	sb.WriteString(fmt.Sprintf("Curve: %s\n", arch.Curve))
	if arch.WinRate > 0 {
		sb.WriteString(fmt.Sprintf("Win Rate: %.1f%%\n", arch.WinRate))
	}
	sb.WriteString("\n")

	// Key cards
	if len(arch.KeyCards) > 0 {
		sb.WriteString("Key Cards:\n")
		for _, card := range arch.KeyCards {
			sb.WriteString(fmt.Sprintf("  • %s\n", card))
		}
		sb.WriteString("\n")
	}

	// P1P1 priorities
	if len(arch.Priorities) > 0 {
		sb.WriteString("P1P1 Priorities:\n")
		for i, card := range arch.Priorities {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, card))
		}
		sb.WriteString("\n")
	}

	// Cards to avoid
	if len(arch.Avoid) > 0 {
		sb.WriteString("Avoid:\n")
		for _, card := range arch.Avoid {
			sb.WriteString(fmt.Sprintf("  ⚠️  %s\n", card))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatColorPairs formats color pair win rates for CLI display.
func FormatColorPairs(colorRatings map[string]float64) string {
	var sb strings.Builder

	type colorPair struct {
		colors  string
		winRate float64
	}

	var pairs []colorPair
	for colors, winRate := range colorRatings {
		if len(colors) == 2 { // Two-color pairs only
			pairs = append(pairs, colorPair{colors, winRate})
		}
	}

	// Sort by win rate
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].winRate > pairs[i].winRate {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	sb.WriteString("═══ Color Pair Win Rates ═══\n\n")

	for i, pair := range pairs {
		tier := ""
		switch {
		case pair.winRate >= 55.0:
			tier = "A"
		case pair.winRate >= 53.0:
			tier = "B+"
		case pair.winRate >= 51.0:
			tier = "B"
		case pair.winRate >= 49.0:
			tier = "C+"
		default:
			tier = "C"
		}

		sb.WriteString(fmt.Sprintf("%2d. %s - %.1f%% (Tier %s) - %s\n",
			i+1,
			pair.colors,
			pair.winRate,
			tier,
			formatColorName(pair.colors)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatColorName(colors string) string {
	names := map[string]string{
		"W": "White",
		"U": "Blue",
		"B": "Black",
		"R": "Red",
		"G": "Green",
	}

	var parts []string
	for _, c := range colors {
		if name, ok := names[string(c)]; ok {
			parts = append(parts, name)
		}
	}

	if len(parts) == 0 {
		return "Unknown"
	}

	return strings.Join(parts, "-")
}

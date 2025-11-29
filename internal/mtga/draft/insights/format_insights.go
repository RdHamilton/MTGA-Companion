package insights

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// FormatInsights contains aggregated statistics and insights about a draft format.
type FormatInsights struct {
	SetCode     string `json:"set_code"`
	DraftFormat string `json:"draft_format"`

	// Color power rankings
	ColorRankings []ColorPowerRank `json:"color_rankings"`

	// Top cards by category
	TopBombs     []TopCard `json:"top_bombs"`     // Top rares/mythics by GIHWR
	TopRemoval   []TopCard `json:"top_removal"`   // Top removal spells
	TopCreatures []TopCard `json:"top_creatures"` // Top creatures by CMC
	TopCommons   []TopCard `json:"top_commons"`   // Best common cards

	// Format characteristics
	FormatSpeed   FormatSpeed    `json:"format_speed"`
	ColorAnalysis *ColorAnalysis `json:"color_analysis"`
}

// ColorPowerRank represents a color's power ranking.
type ColorPowerRank struct {
	Color       string  `json:"color"`    // e.g., "W", "UB"
	WinRate     float64 `json:"win_rate"` // Win rate percentage
	GamesPlayed int     `json:"games_played"`
	Popularity  float64 `json:"popularity"` // Percentage of decks
	Rating      string  `json:"rating"`     // "S", "A", "B", "C", "D"
}

// TopCard represents a top-performing card.
type TopCard struct {
	Name   string  `json:"name"`
	Color  string  `json:"color"`
	Rarity string  `json:"rarity"`
	GIHWR  float64 `json:"gihwr"` // Games In Hand Win Rate
	CMC    int     `json:"cmc,omitempty"`
}

// FormatSpeed characterizes the format's speed.
type FormatSpeed struct {
	Speed       string  `json:"speed"`                   // "Fast", "Medium", "Slow"
	AvgGameTurn float64 `json:"avg_game_turn,omitempty"` // Future: average game length
	Description string  `json:"description"`             // Human-readable description
}

// ColorAnalysis provides insights about color distribution and depth.
type ColorAnalysis struct {
	BestMonoColor     string             `json:"best_mono_color"`    // Best single color
	BestColorPair     string             `json:"best_color_pair"`    // Best 2-color combination
	DeepestColors     []string           `json:"deepest_colors"`     // Colors with most playables
	OverdraftedColors []OverdraftedColor `json:"overdrafted_colors"` // Colors picked too often for their win rate
}

// OverdraftedColor represents a color that's more popular than its win rate suggests.
type OverdraftedColor struct {
	Color      string  `json:"color"`
	WinRate    float64 `json:"win_rate"`
	Popularity float64 `json:"popularity"`
	Delta      float64 `json:"delta"` // Popularity - WinRate (high delta = overdrafted)
}

// Analyzer aggregates draft data into format-level insights.
// It uses the Strategy pattern to apply format-specific analysis logic.
type Analyzer struct {
	service  *storage.Service
	strategy InsightsStrategy
}

// NewAnalyzer creates a new format insights analyzer with a specific strategy.
func NewAnalyzer(service *storage.Service, strategy InsightsStrategy) *Analyzer {
	return &Analyzer{
		service:  service,
		strategy: strategy,
	}
}

// NewAnalyzerForFormat creates a new format insights analyzer with the appropriate
// strategy based on the draft format.
func NewAnalyzerForFormat(service *storage.Service, draftFormat string) *Analyzer {
	strategy := StrategyFactory(context.TODO(), draftFormat)
	return NewAnalyzer(service, strategy)
}

// NormalizeDraftFormat extracts the base format from a full event name.
// e.g., "QuickDraft_TLA_20251127" -> "QuickDraft"
func NormalizeDraftFormat(draftFormat string) string {
	if idx := strings.Index(draftFormat, "_"); idx != -1 {
		return draftFormat[:idx]
	}
	return draftFormat
}

// AnalyzeFormat generates insights for a specific format.
func (a *Analyzer) AnalyzeFormat(ctx context.Context, setCode, draftFormat string) (*FormatInsights, error) {
	// Normalize the draft format to handle both "QuickDraft" and "QuickDraft_TLA_20251127"
	normalizedFormat := NormalizeDraftFormat(draftFormat)

	// Get card and color ratings from database
	cardRatings, _, err := a.service.DraftRatingsRepo().GetCardRatings(ctx, setCode, normalizedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get card ratings: %w", err)
	}

	colorRatings, _, err := a.service.DraftRatingsRepo().GetColorRatings(ctx, setCode, normalizedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get color ratings: %w", err)
	}

	if len(cardRatings) == 0 {
		return nil, fmt.Errorf("no card ratings available for %s/%s", setCode, normalizedFormat)
	}

	insights := &FormatInsights{
		SetCode:     setCode,
		DraftFormat: normalizedFormat,
	}

	// Use strategy to analyze format-specific insights
	insights.ColorRankings = a.strategy.AnalyzeColorRankings(colorRatings)
	insights.TopBombs = a.strategy.FindTopBombs(cardRatings)
	insights.TopRemoval = a.strategy.FindTopRemoval(cardRatings)
	insights.TopCreatures = a.strategy.FindTopCreatures(cardRatings)
	insights.TopCommons = a.strategy.FindTopCommons(cardRatings)
	insights.FormatSpeed = a.strategy.AnalyzeFormatSpeed(cardRatings)
	insights.ColorAnalysis = a.strategy.AnalyzeColors(cardRatings, colorRatings)

	return insights, nil
}

// getRating assigns a letter grade based on win rate.
func getRating(winRate float64) string {
	if winRate >= 60.0 {
		return "S"
	} else if winRate >= 57.0 {
		return "A"
	} else if winRate >= 54.0 {
		return "B"
	} else if winRate >= 51.0 {
		return "C"
	}
	return "D"
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	// Simple case-insensitive contains
	return len(s) >= len(substr) && anyMatch(s, substr)
}

func anyMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if lower(s[i+j]) != lower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

// ArchetypeCards contains top cards for a specific color combination (archetype).
type ArchetypeCards struct {
	Colors       string    `json:"colors"`        // e.g., "UB", "WUR"
	TopCards     []TopCard `json:"top_cards"`     // Best cards overall for this archetype
	TopCreatures []TopCard `json:"top_creatures"` // Best creatures
	TopRemoval   []TopCard `json:"top_removal"`   // Best removal spells
	TopCommons   []TopCard `json:"top_commons"`   // Best commons
}

// GetArchetypeCards returns top cards for a specific color combination.
// colors parameter should be a color combination like "W", "UB", "WUR", etc.
func (a *Analyzer) GetArchetypeCards(ctx context.Context, setCode, draftFormat, colors string) (*ArchetypeCards, error) {
	// Normalize the draft format to handle both "QuickDraft" and "QuickDraft_TLA_20251127"
	normalizedFormat := NormalizeDraftFormat(draftFormat)

	// Get card ratings from database
	cardRatings, _, err := a.service.DraftRatingsRepo().GetCardRatings(ctx, setCode, normalizedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get card ratings: %w", err)
	}

	if len(cardRatings) == 0 {
		return nil, fmt.Errorf("no card ratings available for %s/%s", setCode, normalizedFormat)
	}

	// Use strategy to filter cards by color combination
	filteredCards := a.strategy.FilterCardsByColor(cardRatings, colors)

	archetype := &ArchetypeCards{
		Colors:       colors,
		TopCards:     findTopCardsForArchetype(filteredCards, 20),
		TopCreatures: findTopCardsForArchetype(filteredCards, 15),
		TopRemoval:   findTopRemovalForArchetype(filteredCards, 10),
		TopCommons:   findTopCommonsForArchetype(filteredCards, 15),
	}

	return archetype, nil
}

// isColorMatch checks if a card's color fits within the specified color combination.
// A card matches if all its colors are present in the requested colors.
// For example: card color "U" matches "UB", card color "UB" matches "UBR", but "UB" doesn't match "U".
func isColorMatch(cardColor, requestedColors string) bool {
	if cardColor == "" || cardColor == "C" {
		// Colorless cards match any color combination
		return true
	}

	// Check if all card colors are in the requested colors
	for i := 0; i < len(cardColor); i++ {
		found := false
		for j := 0; j < len(requestedColors); j++ {
			if cardColor[i] == requestedColors[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// findTopCardsForArchetype finds the best cards overall for an archetype.
func findTopCardsForArchetype(cardRatings []seventeenlands.CardRating, limit int) []TopCard {
	cards := []TopCard{}
	for _, card := range cardRatings {
		cards = append(cards, TopCard{
			Name:   card.Name,
			Color:  card.Color,
			Rarity: card.Rarity,
			GIHWR:  card.GIHWR * 100,
		})
	}

	// Sort by GIHWR
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].GIHWR > cards[j].GIHWR
	})

	if len(cards) > limit {
		cards = cards[:limit]
	}

	return cards
}

// findTopRemovalForArchetype finds the best removal spells for an archetype.
func findTopRemovalForArchetype(cardRatings []seventeenlands.CardRating, limit int) []TopCard {
	removalKeywords := []string{"Destroy", "Exile", "Murder", "Kill", "Deal", "Damage", "Remove", "Bounce"}

	removal := []TopCard{}
	for _, card := range cardRatings {
		isRemoval := false
		for _, keyword := range removalKeywords {
			if contains(card.Name, keyword) {
				isRemoval = true
				break
			}
		}

		if isRemoval {
			removal = append(removal, TopCard{
				Name:   card.Name,
				Color:  card.Color,
				Rarity: card.Rarity,
				GIHWR:  card.GIHWR * 100,
			})
		}
	}

	// Sort by GIHWR
	sort.Slice(removal, func(i, j int) bool {
		return removal[i].GIHWR > removal[j].GIHWR
	})

	if len(removal) > limit {
		removal = removal[:limit]
	}

	return removal
}

// findTopCommonsForArchetype finds the best common cards for an archetype.
func findTopCommonsForArchetype(cardRatings []seventeenlands.CardRating, limit int) []TopCard {
	commons := []TopCard{}
	for _, card := range cardRatings {
		if card.Rarity == "common" {
			commons = append(commons, TopCard{
				Name:   card.Name,
				Color:  card.Color,
				Rarity: card.Rarity,
				GIHWR:  card.GIHWR * 100,
			})
		}
	}

	// Sort by GIHWR
	sort.Slice(commons, func(i, j int) bool {
		return commons[i].GIHWR > commons[j].GIHWR
	})

	if len(commons) > limit {
		commons = commons[:limit]
	}

	return commons
}

package insights

import (
	"context"
	"fmt"
	"sort"

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
type Analyzer struct {
	service *storage.Service
}

// NewAnalyzer creates a new format insights analyzer.
func NewAnalyzer(service *storage.Service) *Analyzer {
	return &Analyzer{service: service}
}

// AnalyzeFormat generates insights for a specific format.
func (a *Analyzer) AnalyzeFormat(ctx context.Context, setCode, draftFormat string) (*FormatInsights, error) {
	// Get card and color ratings from database
	cardRatings, _, err := a.service.DraftRatingsRepo().GetCardRatings(ctx, setCode, draftFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get card ratings: %w", err)
	}

	colorRatings, _, err := a.service.DraftRatingsRepo().GetColorRatings(ctx, setCode, draftFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get color ratings: %w", err)
	}

	if len(cardRatings) == 0 {
		return nil, fmt.Errorf("no card ratings available for %s/%s", setCode, draftFormat)
	}

	insights := &FormatInsights{
		SetCode:     setCode,
		DraftFormat: draftFormat,
	}

	// Analyze color rankings
	insights.ColorRankings = a.analyzeColorRankings(colorRatings)

	// Find top cards by category
	insights.TopBombs = a.findTopBombs(cardRatings)
	insights.TopRemoval = a.findTopRemoval(cardRatings)
	insights.TopCreatures = a.findTopCreatures(cardRatings)
	insights.TopCommons = a.findTopCommons(cardRatings)

	// Analyze format speed
	insights.FormatSpeed = a.analyzeFormatSpeed(cardRatings)

	// Analyze color depth and overdrafted colors
	insights.ColorAnalysis = a.analyzeColors(cardRatings, colorRatings)

	return insights, nil
}

// analyzeColorRankings ranks colors by win rate.
func (a *Analyzer) analyzeColorRankings(colorRatings []seventeenlands.ColorRating) []ColorPowerRank {
	if len(colorRatings) == 0 {
		return []ColorPowerRank{}
	}

	// Calculate total games for popularity
	var totalGames int
	for _, rating := range colorRatings {
		totalGames += rating.GamesPlayed
	}

	rankings := make([]ColorPowerRank, 0, len(colorRatings))
	for _, rating := range colorRatings {
		// Skip splash combinations
		if rating.IsSplash {
			continue
		}

		popularity := 0.0
		if totalGames > 0 {
			popularity = (float64(rating.GamesPlayed) / float64(totalGames)) * 100
		}

		rank := ColorPowerRank{
			Color:       rating.ColorName,
			WinRate:     rating.WinRate * 100, // Convert to percentage
			GamesPlayed: rating.GamesPlayed,
			Popularity:  popularity,
			Rating:      getRating(rating.WinRate * 100),
		}
		rankings = append(rankings, rank)
	}

	// Sort by win rate descending
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].WinRate > rankings[j].WinRate
	})

	return rankings
}

// findTopBombs finds the best rare/mythic cards.
func (a *Analyzer) findTopBombs(cardRatings []seventeenlands.CardRating) []TopCard {
	bombs := []TopCard{}
	for _, card := range cardRatings {
		if card.Rarity == "rare" || card.Rarity == "mythic" {
			bombs = append(bombs, TopCard{
				Name:   card.Name,
				Color:  card.Color,
				Rarity: card.Rarity,
				GIHWR:  card.GIHWR * 100,
			})
		}
	}

	// Sort by GIHWR and take top 10
	sort.Slice(bombs, func(i, j int) bool {
		return bombs[i].GIHWR > bombs[j].GIHWR
	})

	if len(bombs) > 10 {
		bombs = bombs[:10]
	}

	return bombs
}

// findTopRemoval finds the best removal spells (heuristic: name contains common removal keywords).
func (a *Analyzer) findTopRemoval(cardRatings []seventeenlands.CardRating) []TopCard {
	// Simple heuristic: common removal keywords
	removalKeywords := []string{"Destroy", "Exile", "Murder", "Kill", "Deal", "Damage", "Remove", "Bounce"}

	removal := []TopCard{}
	for _, card := range cardRatings {
		// Check if card name suggests removal
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

	// Sort by GIHWR and take top 8
	sort.Slice(removal, func(i, j int) bool {
		return removal[i].GIHWR > removal[j].GIHWR
	})

	if len(removal) > 8 {
		removal = removal[:8]
	}

	return removal
}

// findTopCreatures finds the best creatures.
func (a *Analyzer) findTopCreatures(cardRatings []seventeenlands.CardRating) []TopCard {
	// Note: We don't have card type info in CardRating, so we can't filter by creature type
	// This is a limitation - would need to join with set_cards table or add type to ratings
	// For now, return top cards overall (which are likely creatures)
	creatures := []TopCard{}
	for _, card := range cardRatings {
		creatures = append(creatures, TopCard{
			Name:   card.Name,
			Color:  card.Color,
			Rarity: card.Rarity,
			GIHWR:  card.GIHWR * 100,
		})
	}

	// Sort by GIHWR and take top 15
	sort.Slice(creatures, func(i, j int) bool {
		return creatures[i].GIHWR > creatures[j].GIHWR
	})

	if len(creatures) > 15 {
		creatures = creatures[:15]
	}

	return creatures
}

// findTopCommons finds the best common cards.
func (a *Analyzer) findTopCommons(cardRatings []seventeenlands.CardRating) []TopCard {
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

	// Sort by GIHWR and take top 20
	sort.Slice(commons, func(i, j int) bool {
		return commons[i].GIHWR > commons[j].GIHWR
	})

	if len(commons) > 20 {
		commons = commons[:20]
	}

	return commons
}

// analyzeFormatSpeed determines if the format is fast/medium/slow.
func (a *Analyzer) analyzeFormatSpeed(cardRatings []seventeenlands.CardRating) FormatSpeed {
	// Heuristic: Look at average ALSA/ATA of top cards
	// Low ALSA = cards picked early = likely bombs/fast cards
	// High ALSA = cards last late = slow format

	if len(cardRatings) == 0 {
		return FormatSpeed{
			Speed:       "Unknown",
			Description: "Insufficient data to determine format speed",
		}
	}

	var totalALSA float64
	var count int
	for _, card := range cardRatings {
		if card.ALSA > 0 {
			totalALSA += card.ALSA
			count++
		}
	}

	avgALSA := totalALSA / float64(count)

	speed := FormatSpeed{}
	if avgALSA < 6.0 {
		speed.Speed = "Fast"
		speed.Description = "Aggressive format with early picks valued highly"
	} else if avgALSA < 8.0 {
		speed.Speed = "Medium"
		speed.Description = "Balanced format with mix of strategies"
	} else {
		speed.Speed = "Slow"
		speed.Description = "Slower format favoring late-game value"
	}

	return speed
}

// analyzeColors provides insights about color depth and overdrafted colors.
func (a *Analyzer) analyzeColors(cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating) *ColorAnalysis {
	if len(colorRatings) == 0 {
		return nil
	}

	analysis := &ColorAnalysis{
		DeepestColors:     []string{},
		OverdraftedColors: []OverdraftedColor{},
	}

	// Find best mono color and best pair
	var bestMonoWR, bestPairWR float64
	for _, rating := range colorRatings {
		if rating.IsSplash {
			continue
		}

		if len(rating.ColorName) == 1 { // Mono color
			if rating.WinRate > bestMonoWR {
				bestMonoWR = rating.WinRate
				analysis.BestMonoColor = rating.ColorName
			}
		} else if len(rating.ColorName) == 2 { // Two-color pair
			if rating.WinRate > bestPairWR {
				bestPairWR = rating.WinRate
				analysis.BestColorPair = rating.ColorName
			}
		}
	}

	// Calculate total games for overdrafted analysis
	var totalGames int
	for _, rating := range colorRatings {
		totalGames += rating.GamesPlayed
	}

	// Find overdrafted colors (high popularity, low win rate)
	for _, rating := range colorRatings {
		if rating.IsSplash || len(rating.ColorName) > 2 {
			continue
		}

		popularity := (float64(rating.GamesPlayed) / float64(totalGames)) * 100
		winRate := rating.WinRate * 100
		delta := popularity - winRate

		// If popularity exceeds win rate by >5%, it's overdrafted
		if delta > 5.0 {
			analysis.OverdraftedColors = append(analysis.OverdraftedColors, OverdraftedColor{
				Color:      rating.ColorName,
				WinRate:    winRate,
				Popularity: popularity,
				Delta:      delta,
			})
		}
	}

	// Sort overdrafted colors by delta
	sort.Slice(analysis.OverdraftedColors, func(i, j int) bool {
		return analysis.OverdraftedColors[i].Delta > analysis.OverdraftedColors[j].Delta
	})

	return analysis
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

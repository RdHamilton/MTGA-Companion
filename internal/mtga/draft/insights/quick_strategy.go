package insights

import (
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// QuickDraftStrategy implements insights analysis for Quick Draft format.
// Quick Draft features bot opponents, which leads to:
// - More predictable draft patterns
// - Bots may undervalue certain archetypes
// - Less adaptive strategies
// - Different card value assessments based on bot behavior
// - Opportunities to exploit bot tendencies
type QuickDraftStrategy struct{}

// NewQuickDraftStrategy creates a new Quick Draft insights strategy.
func NewQuickDraftStrategy() *QuickDraftStrategy {
	return &QuickDraftStrategy{}
}

// GetStrategyName returns the name of this strategy.
func (s *QuickDraftStrategy) GetStrategyName() string {
	return "Quick Draft"
}

// AnalyzeColorRankings ranks colors by win rate for Quick Draft.
// In Quick Draft, we also consider how bots value colors, which can
// create opportunities to draft undervalued archetypes.
func (s *QuickDraftStrategy) AnalyzeColorRankings(colorRatings []seventeenlands.ColorRating) []ColorPowerRank {
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

// FindTopBombs finds the best rare/mythic cards for Quick Draft.
// In Quick Draft, bots may pass bombs more frequently as they don't
// always recognize the most powerful cards. This can lead to different
// valuations compared to Premier Draft.
func (s *QuickDraftStrategy) FindTopBombs(cardRatings []seventeenlands.CardRating) []TopCard {
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

	// Sort by GIHWR and take top 12 (slightly more than Premier)
	// Bots may pass more bombs, so we show more options
	sort.Slice(bombs, func(i, j int) bool {
		return bombs[i].GIHWR > bombs[j].GIHWR
	})

	if len(bombs) > 12 {
		bombs = bombs[:12]
	}

	return bombs
}

// FindTopRemoval finds the best removal spells for Quick Draft.
// Removal is still valuable but bots may not protect their threats as well.
func (s *QuickDraftStrategy) FindTopRemoval(cardRatings []seventeenlands.CardRating) []TopCard {
	// Common removal keywords
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

	// Sort by GIHWR and take top 10 (slightly more than Premier)
	sort.Slice(removal, func(i, j int) bool {
		return removal[i].GIHWR > removal[j].GIHWR
	})

	if len(removal) > 10 {
		removal = removal[:10]
	}

	return removal
}

// FindTopCreatures finds the best creatures for Quick Draft.
// Bots may undervalue certain creature types, creating opportunities.
func (s *QuickDraftStrategy) FindTopCreatures(cardRatings []seventeenlands.CardRating) []TopCard {
	creatures := []TopCard{}
	for _, card := range cardRatings {
		creatures = append(creatures, TopCard{
			Name:   card.Name,
			Color:  card.Color,
			Rarity: card.Rarity,
			GIHWR:  card.GIHWR * 100,
		})
	}

	// Sort by GIHWR and take top 18 (more than Premier)
	// Show more options as bot behavior may create more opportunities
	sort.Slice(creatures, func(i, j int) bool {
		return creatures[i].GIHWR > creatures[j].GIHWR
	})

	if len(creatures) > 18 {
		creatures = creatures[:18]
	}

	return creatures
}

// FindTopCommons finds the best common cards for Quick Draft.
// Commons are especially important in Quick Draft as bots may
// pass high-value commons more consistently.
func (s *QuickDraftStrategy) FindTopCommons(cardRatings []seventeenlands.CardRating) []TopCard {
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

	// Sort by GIHWR and take top 25 (more than Premier)
	// Show more commons as they're more available in Quick Draft
	sort.Slice(commons, func(i, j int) bool {
		return commons[i].GIHWR > commons[j].GIHWR
	})

	if len(commons) > 25 {
		commons = commons[:25]
	}

	return commons
}

// AnalyzeFormatSpeed determines if Quick Draft format is fast/medium/slow.
// Bot behavior may lead to different format speeds than Premier Draft.
func (s *QuickDraftStrategy) AnalyzeFormatSpeed(cardRatings []seventeenlands.CardRating) FormatSpeed {
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

	if count == 0 {
		return FormatSpeed{
			Speed:       "Unknown",
			Description: "Insufficient data to determine format speed",
		}
	}

	avgALSA := totalALSA / float64(count)

	// Quick Draft may skew slightly faster due to bot behavior
	// Adjust thresholds slightly lower than Premier Draft
	speed := FormatSpeed{}
	if avgALSA < 5.5 {
		speed.Speed = "Fast"
		speed.Description = "Fast format - bots favor aggressive strategies"
	} else if avgALSA < 7.5 {
		speed.Speed = "Medium"
		speed.Description = "Medium-paced format with bot tendencies toward aggro"
	} else {
		speed.Speed = "Slow"
		speed.Description = "Slower format - good opportunity for value strategies against bots"
	}

	return speed
}

// AnalyzeColors provides insights about color depth and overdrafted colors.
// In Quick Draft, bots may systematically overdraft certain colors.
func (s *QuickDraftStrategy) AnalyzeColors(cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating) *ColorAnalysis {
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
	// In Quick Draft, use a lower threshold of 3% as bot behavior is more systematic
	for _, rating := range colorRatings {
		if rating.IsSplash || len(rating.ColorName) > 2 {
			continue
		}

		popularity := (float64(rating.GamesPlayed) / float64(totalGames)) * 100
		winRate := rating.WinRate * 100
		delta := popularity - winRate

		// If popularity exceeds win rate by >3%, it's overdrafted
		// Lower threshold than Premier as bot patterns are more consistent
		if delta > 3.0 {
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

// FilterCardsByColor filters cards based on color combination.
// Uses standard color matching logic for Quick Draft.
func (s *QuickDraftStrategy) FilterCardsByColor(cards []seventeenlands.CardRating, colors string) []seventeenlands.CardRating {
	if colors == "" {
		return cards
	}

	filtered := []seventeenlands.CardRating{}
	for _, card := range cards {
		// Check if card's color is a subset of the requested colors
		if isColorMatch(card.Color, colors) {
			filtered = append(filtered, card)
		}
	}

	return filtered
}

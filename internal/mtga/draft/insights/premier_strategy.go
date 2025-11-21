package insights

import (
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// PremierDraftStrategy implements insights analysis for Premier Draft format.
// Premier Draft features human opponents, which leads to:
// - More dynamic meta-game patterns
// - Adaptive drafting strategies
// - Different color popularity based on perceived power
// - Higher value on bombs and removal
type PremierDraftStrategy struct{}

// NewPremierDraftStrategy creates a new Premier Draft insights strategy.
func NewPremierDraftStrategy() *PremierDraftStrategy {
	return &PremierDraftStrategy{}
}

// GetStrategyName returns the name of this strategy.
func (s *PremierDraftStrategy) GetStrategyName() string {
	return "Premier Draft"
}

// AnalyzeColorRankings ranks colors by win rate for Premier Draft.
// In Premier Draft, we weight win rate more heavily as human opponents
// create a more competitive environment.
func (s *PremierDraftStrategy) AnalyzeColorRankings(colorRatings []seventeenlands.ColorRating) []ColorPowerRank {
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

// FindTopBombs finds the best rare/mythic cards for Premier Draft.
// In Premier Draft, bombs are highly valued as human opponents can
// better capitalize on powerful cards.
func (s *PremierDraftStrategy) FindTopBombs(cardRatings []seventeenlands.CardRating) []TopCard {
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

// FindTopRemoval finds the best removal spells for Premier Draft.
// Removal is premium in Premier Draft as humans play bombs more effectively.
func (s *PremierDraftStrategy) FindTopRemoval(cardRatings []seventeenlands.CardRating) []TopCard {
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

	// Sort by GIHWR and take top 8
	sort.Slice(removal, func(i, j int) bool {
		return removal[i].GIHWR > removal[j].GIHWR
	})

	if len(removal) > 8 {
		removal = removal[:8]
	}

	return removal
}

// FindTopCreatures finds the best creatures for Premier Draft.
// In Premier Draft, we return top cards overall as creatures are typically
// the highest-rated cards.
func (s *PremierDraftStrategy) FindTopCreatures(cardRatings []seventeenlands.CardRating) []TopCard {
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

// FindTopCommons finds the best common cards for Premier Draft.
// Commons form the backbone of Premier Draft decks.
func (s *PremierDraftStrategy) FindTopCommons(cardRatings []seventeenlands.CardRating) []TopCard {
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

// AnalyzeFormatSpeed determines if Premier Draft format is fast/medium/slow.
// Uses ALSA (Average Last Seen At) as a heuristic for format speed.
func (s *PremierDraftStrategy) AnalyzeFormatSpeed(cardRatings []seventeenlands.CardRating) FormatSpeed {
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

// AnalyzeColors provides insights about color depth and overdrafted colors.
// In Premier Draft, humans may overdraft popular colors based on perception.
func (s *PremierDraftStrategy) AnalyzeColors(cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating) *ColorAnalysis {
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
	// In Premier Draft, use a threshold of 5% to identify overdrafted colors
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

// FilterCardsByColor filters cards based on color combination.
// Uses standard color matching logic for Premier Draft.
func (s *PremierDraftStrategy) FilterCardsByColor(cards []seventeenlands.CardRating, colors string) []seventeenlands.CardRating {
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

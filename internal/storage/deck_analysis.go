package storage

import (
	"context"
	"fmt"
	"strings"
)

// DeckAnalysis contains comprehensive analysis of a deck.
type DeckAnalysis struct {
	DeckID        string
	DeckName      string
	TotalCards    int
	ManaCurve     ManaCurveAnalysis
	ColorDist     ColorDistribution
	TypeBreakdown TypeBreakdown
	AverageCMC    float64
	LandAnalysis  LandAnalysis
}

// ManaCurveAnalysis represents the distribution of cards by mana cost.
type ManaCurveAnalysis struct {
	Curve  map[int]int // CMC -> count
	MaxCMC int
}

// ColorDistribution represents the distribution of cards by color.
type ColorDistribution struct {
	White             int
	Blue              int
	Black             int
	Red               int
	Green             int
	Colorless         int
	Multicolor        int
	TotalColoredCards int // Cards that have at least one color
}

// TypeBreakdown represents the distribution of cards by type.
type TypeBreakdown struct {
	Creatures     int
	Instants      int
	Sorceries     int
	Enchantments  int
	Artifacts     int
	Planeswalkers int
	Lands         int
	Other         int
}

// LandAnalysis contains statistics about lands in the deck.
type LandAnalysis struct {
	TotalLands    int
	BasicLands    int
	NonBasicLands int
	LandRatio     float64 // Percentage of deck that is lands
	SpellsToLands float64 // Ratio of spells to lands
}

// AnalyzeDeck performs comprehensive analysis of a deck.
func (s *Service) AnalyzeDeck(ctx context.Context, deckID string) (*DeckAnalysis, error) {
	// Get deck information
	deck, err := s.decks.GetByID(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck: %w", err)
	}
	if deck == nil {
		return nil, fmt.Errorf("deck not found: %s", deckID)
	}

	// Get all cards in the deck (main deck only, excluding sideboard)
	deckCards, err := s.decks.GetCards(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck cards: %w", err)
	}

	// Filter to main deck only
	var mainDeckCards []*struct {
		CardID   int
		Quantity int
	}
	for _, dc := range deckCards {
		if dc.Board == "main" {
			mainDeckCards = append(mainDeckCards, &struct {
				CardID   int
				Quantity int
			}{CardID: dc.CardID, Quantity: dc.Quantity})
		}
	}

	// Initialize analysis
	analysis := &DeckAnalysis{
		DeckID:   deckID,
		DeckName: deck.Name,
		ManaCurve: ManaCurveAnalysis{
			Curve: make(map[int]int),
		},
	}

	totalCMC := 0.0
	totalCards := 0

	// Analyze each card
	for _, dc := range mainDeckCards {
		// Get card metadata
		card, err := s.GetCardByArenaID(ctx, dc.CardID)
		if err != nil {
			// Skip cards without metadata
			continue
		}
		if card == nil {
			continue
		}

		quantity := dc.Quantity
		totalCards += quantity

		// Mana curve analysis
		cmc := int(card.CMC)
		analysis.ManaCurve.Curve[cmc] += quantity
		if cmc > analysis.ManaCurve.MaxCMC {
			analysis.ManaCurve.MaxCMC = cmc
		}
		totalCMC += card.CMC * float64(quantity)

		// Color distribution analysis
		analyzeColors(card, quantity, &analysis.ColorDist)

		// Type breakdown analysis
		analyzeTypes(card, quantity, &analysis.TypeBreakdown)
	}

	analysis.TotalCards = totalCards

	// Calculate average CMC
	if totalCards > 0 {
		analysis.AverageCMC = totalCMC / float64(totalCards)
	}

	// Land analysis
	analysis.LandAnalysis.TotalLands = analysis.TypeBreakdown.Lands
	analysis.LandAnalysis.BasicLands = countBasicLands(mainDeckCards, s, ctx)
	analysis.LandAnalysis.NonBasicLands = analysis.LandAnalysis.TotalLands - analysis.LandAnalysis.BasicLands

	if totalCards > 0 {
		analysis.LandAnalysis.LandRatio = float64(analysis.LandAnalysis.TotalLands) / float64(totalCards) * 100
	}

	nonLands := totalCards - analysis.LandAnalysis.TotalLands
	if analysis.LandAnalysis.TotalLands > 0 {
		analysis.LandAnalysis.SpellsToLands = float64(nonLands) / float64(analysis.LandAnalysis.TotalLands)
	}

	return analysis, nil
}

// analyzeColors updates color distribution based on card colors.
func analyzeColors(card *Card, quantity int, dist *ColorDistribution) {
	if len(card.Colors) == 0 {
		dist.Colorless += quantity
		return
	}

	if len(card.Colors) > 1 {
		dist.Multicolor += quantity
		dist.TotalColoredCards += quantity
		return
	}

	// Single color
	dist.TotalColoredCards += quantity
	color := card.Colors[0]
	switch color {
	case "W":
		dist.White += quantity
	case "U":
		dist.Blue += quantity
	case "B":
		dist.Black += quantity
	case "R":
		dist.Red += quantity
	case "G":
		dist.Green += quantity
	}
}

// analyzeTypes updates type breakdown based on card type line.
func analyzeTypes(card *Card, quantity int, breakdown *TypeBreakdown) {
	typeLine := strings.ToLower(card.TypeLine)

	// Check for each type (order matters for multi-type cards)
	if strings.Contains(typeLine, "land") {
		breakdown.Lands += quantity
	} else if strings.Contains(typeLine, "creature") {
		breakdown.Creatures += quantity
	} else if strings.Contains(typeLine, "planeswalker") {
		breakdown.Planeswalkers += quantity
	} else if strings.Contains(typeLine, "instant") {
		breakdown.Instants += quantity
	} else if strings.Contains(typeLine, "sorcery") {
		breakdown.Sorceries += quantity
	} else if strings.Contains(typeLine, "enchantment") {
		breakdown.Enchantments += quantity
	} else if strings.Contains(typeLine, "artifact") {
		breakdown.Artifacts += quantity
	} else {
		breakdown.Other += quantity
	}
}

// countBasicLands counts the number of basic lands in a deck.
func countBasicLands(cards []*struct {
	CardID   int
	Quantity int
}, s *Service, ctx context.Context,
) int {
	basicLandNames := map[string]bool{
		"Plains":   true,
		"Island":   true,
		"Swamp":    true,
		"Mountain": true,
		"Forest":   true,
		"Wastes":   true,
	}

	count := 0
	for _, dc := range cards {
		card, err := s.GetCardByArenaID(ctx, dc.CardID)
		if err != nil || card == nil {
			continue
		}

		if basicLandNames[card.Name] {
			count += dc.Quantity
		}
	}

	return count
}

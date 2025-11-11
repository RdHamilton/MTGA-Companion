package draft

import (
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// RatingsProvider provides card ratings for draft picks.
type RatingsProvider struct {
	setFile *seventeenlands.SetFile
	config  BayesianConfig
}

// BayesianConfig configures Bayesian averaging for ratings.
type BayesianConfig struct {
	// MinGIH is the minimum games-in-hand threshold (default: 300)
	MinGIH int

	// PriorMean is the prior mean for Bayesian averaging (default: 50.0)
	PriorMean float64

	// PriorWeight is the weight of the prior (default: 100)
	PriorWeight int
}

// DefaultBayesianConfig returns sensible defaults for Bayesian averaging.
func DefaultBayesianConfig() BayesianConfig {
	return BayesianConfig{
		MinGIH:      300,
		PriorMean:   50.0,
		PriorWeight: 100,
	}
}

// NewRatingsProvider creates a ratings provider for a specific set.
func NewRatingsProvider(setFile *seventeenlands.SetFile, config BayesianConfig) *RatingsProvider {
	return &RatingsProvider{
		setFile: setFile,
		config:  config,
	}
}

// PackRatings contains ratings for all cards in a pack.
type PackRatings struct {
	Pack        *Pack
	CardRatings []*CardRating
	ColorFilter string // Color filter applied (e.g., "BR", "ALL")
}

// CardRating contains rating information for a single card.
type CardRating struct {
	CardID           int     // Arena card ID
	Name             string  // Card name
	ManaCost         string  // Mana cost
	CMC              float64 // Converted mana cost
	Colors           []string
	Rarity           string
	Types            []string
	GIHWR            float64 // Games In Hand Win Rate
	ALSA             float64 // Average Last Seen At
	ATA              float64 // Average Taken At
	IWD              float64 // Improvement When Drawn
	GIH              int     // Games In Hand (sample size)
	BayesianGIHWR    float64 // Bayesian-adjusted GIHWR
	IsBayesianAdjust bool    // Whether Bayesian adjustment was applied
}

// GetPackRatings retrieves ratings for all cards in a pack.
// colorFilter specifies which deck colors to use for ratings (e.g., "BR", "WU", "ALL").
func (rp *RatingsProvider) GetPackRatings(pack *Pack, colorFilter string) (*PackRatings, error) {
	if pack == nil {
		return nil, fmt.Errorf("pack is nil")
	}

	if colorFilter == "" {
		colorFilter = "ALL"
	}

	cardRatings := make([]*CardRating, 0, len(pack.CardIDs))

	for _, cardID := range pack.CardIDs {
		rating, err := rp.GetCardRating(cardID, colorFilter)
		if err != nil {
			// Card not found in set file - skip it
			continue
		}
		cardRatings = append(cardRatings, rating)
	}

	return &PackRatings{
		Pack:        pack,
		CardRatings: cardRatings,
		ColorFilter: colorFilter,
	}, nil
}

// GetCardRating retrieves rating information for a single card.
func (rp *RatingsProvider) GetCardRating(cardID int, colorFilter string) (*CardRating, error) {
	if rp.setFile == nil {
		return nil, fmt.Errorf("set file is nil")
	}

	// Find card in set file (CardRatings is a map keyed by Arena ID as string)
	cardKey := fmt.Sprintf("%d", cardID)
	cardData, ok := rp.setFile.CardRatings[cardKey]
	if !ok || cardData == nil {
		return nil, fmt.Errorf("card ID %d not found in set file", cardID)
	}

	// Get ratings for the specified color filter
	deckColorRatings, ok := cardData.DeckColors[colorFilter]
	if !ok {
		// Fallback to "ALL" if specific color not available
		deckColorRatings, ok = cardData.DeckColors["ALL"]
		if !ok {
			return nil, fmt.Errorf("no ratings available for card ID %d", cardID)
		}
	}

	// Extract basic ratings
	rating := &CardRating{
		CardID:   cardID,
		Name:     cardData.Name,
		ManaCost: cardData.ManaCost,
		CMC:      cardData.CMC,
		Colors:   ParseManaCost(cardData.ManaCost),
		Rarity:   cardData.Rarity,
		Types:    cardData.Types,
		GIHWR:    deckColorRatings.GIHWR,
		ALSA:     deckColorRatings.ALSA,
		ATA:      deckColorRatings.ATA,
		IWD:      deckColorRatings.IWD,
		GIH:      deckColorRatings.GIH,
	}

	// Apply Bayesian adjustment if sample size is below threshold
	if deckColorRatings.GIH < rp.config.MinGIH {
		rating.BayesianGIHWR = rp.calculateBayesianGIHWR(
			deckColorRatings.GIHWR,
			deckColorRatings.GIH,
		)
		rating.IsBayesianAdjust = true
	} else {
		rating.BayesianGIHWR = deckColorRatings.GIHWR
		rating.IsBayesianAdjust = false
	}

	return rating, nil
}

// calculateBayesianGIHWR applies Bayesian averaging to GIHWR.
// Formula: (PriorWeight * PriorMean + GIH * ObservedWR) / (PriorWeight + GIH)
func (rp *RatingsProvider) calculateBayesianGIHWR(observedGIHWR float64, gih int) float64 {
	numerator := float64(rp.config.PriorWeight)*rp.config.PriorMean + float64(gih)*observedGIHWR
	denominator := float64(rp.config.PriorWeight + gih)
	return numerator / denominator
}

// SortByRating sorts card ratings by Bayesian-adjusted GIHWR (descending).
func (pr *PackRatings) SortByRating() {
	cardRatings := pr.CardRatings
	for i := 0; i < len(cardRatings); i++ {
		for j := i + 1; j < len(cardRatings); j++ {
			if cardRatings[j].BayesianGIHWR > cardRatings[i].BayesianGIHWR {
				cardRatings[i], cardRatings[j] = cardRatings[j], cardRatings[i]
			}
		}
	}
}

// TopN returns the top N cards by Bayesian-adjusted GIHWR.
func (pr *PackRatings) TopN(n int) []*CardRating {
	pr.SortByRating()
	if n > len(pr.CardRatings) {
		n = len(pr.CardRatings)
	}
	return pr.CardRatings[:n]
}

// GetBestPick returns the highest-rated card in the pack.
func (pr *PackRatings) GetBestPick() *CardRating {
	if len(pr.CardRatings) == 0 {
		return nil
	}
	pr.SortByRating()
	return pr.CardRatings[0]
}

// FilterByColors returns only cards matching the specified colors.
func (pr *PackRatings) FilterByColors(colors []string) []*CardRating {
	if len(colors) == 0 {
		return pr.CardRatings
	}

	filtered := make([]*CardRating, 0)
	for _, rating := range pr.CardRatings {
		// Check if card's colors are a subset of the filter colors
		if isSubsetOf(rating.Colors, colors) {
			filtered = append(filtered, rating)
		}
	}
	return filtered
}

// GetCardByID finds a card rating by its Arena ID.
func (pr *PackRatings) GetCardByID(cardID int) *CardRating {
	for _, rating := range pr.CardRatings {
		if rating.CardID == cardID {
			return rating
		}
	}
	return nil
}

package pickquality

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Alternative represents an alternative card pick with its rating.
type Alternative struct {
	CardID   string  `json:"card_id"`
	CardName string  `json:"card_name"`
	GIHWR    float64 `json:"gihwr"`
	Rank     int     `json:"rank"`
}

// PickQuality represents the quality analysis of a draft pick.
type PickQuality struct {
	Grade           string        `json:"grade"`             // A+, A, B, C, D, F
	Rank            int           `json:"rank"`              // Position in pack (1 = best)
	PackBestGIHWR   float64       `json:"pack_best_gihwr"`   // Best GIHWR in pack
	PickedCardGIHWR float64       `json:"picked_card_gihwr"` // GIHWR of picked card
	Alternatives    []Alternative `json:"alternatives"`      // Top 5 alternatives
}

// Analyzer analyzes draft pick quality using 17Lands data.
type Analyzer struct {
	ratingsRepo repository.DraftRatingsRepository
	setCardRepo repository.SetCardRepository
}

// NewAnalyzer creates a new pick quality analyzer.
func NewAnalyzer(ratingsRepo repository.DraftRatingsRepository, setCardRepo repository.SetCardRepository) *Analyzer {
	return &Analyzer{
		ratingsRepo: ratingsRepo,
		setCardRepo: setCardRepo,
	}
}

// AnalyzePick analyzes the quality of a draft pick.
// Returns pick quality metrics or nil if ratings are unavailable.
func (a *Analyzer) AnalyzePick(ctx context.Context, setCode, draftFormat string, packCardIDs []string, pickedCardID string) (*PickQuality, error) {
	if len(packCardIDs) == 0 {
		return nil, fmt.Errorf("no cards in pack")
	}

	// Get 17Lands ratings for all cards in the pack
	cardRatings := make(map[string]float64) // arenaID -> GIHWR
	cardNames := make(map[string]string)    // arenaID -> name

	for _, arenaID := range packCardIDs {
		// Get rating from 17Lands
		rating, err := a.ratingsRepo.GetCardRatingByArenaID(ctx, setCode, draftFormat, arenaID)
		if err != nil || rating == nil {
			// Card not found in 17Lands data - assign 0 GIHWR
			cardRatings[arenaID] = 0.0
		} else {
			cardRatings[arenaID] = rating.GIHWR
		}

		// Get card name from set cards
		card, err := a.setCardRepo.GetCardByArenaID(ctx, arenaID)
		if err == nil && card != nil {
			cardNames[arenaID] = card.Name
		} else {
			cardNames[arenaID] = "Unknown Card"
		}
	}

	// Sort cards by GIHWR (descending)
	type cardWithRating struct {
		arenaID string
		gihwr   float64
		name    string
	}

	sortedCards := make([]cardWithRating, 0, len(cardRatings))
	for arenaID, gihwr := range cardRatings {
		sortedCards = append(sortedCards, cardWithRating{
			arenaID: arenaID,
			gihwr:   gihwr,
			name:    cardNames[arenaID],
		})
	}

	sort.Slice(sortedCards, func(i, j int) bool {
		return sortedCards[i].gihwr > sortedCards[j].gihwr
	})

	// Find rank of picked card
	pickedRank := 0
	pickedGIHWR := 0.0
	for i, card := range sortedCards {
		if card.arenaID == pickedCardID {
			pickedRank = i + 1 // 1-indexed
			pickedGIHWR = card.gihwr
			break
		}
	}

	if pickedRank == 0 {
		return nil, fmt.Errorf("picked card not found in pack")
	}

	// Determine grade based on rank
	grade := calculateGrade(pickedRank, len(sortedCards))

	// Get top alternatives (exclude picked card)
	alternatives := make([]Alternative, 0, 5)
	count := 0
	for i, card := range sortedCards {
		if card.arenaID == pickedCardID {
			continue
		}
		if count >= 5 {
			break
		}
		alternatives = append(alternatives, Alternative{
			CardID:   card.arenaID,
			CardName: card.name,
			GIHWR:    card.gihwr,
			Rank:     i + 1,
		})
		count++
	}

	return &PickQuality{
		Grade:           grade,
		Rank:            pickedRank,
		PackBestGIHWR:   sortedCards[0].gihwr,
		PickedCardGIHWR: pickedGIHWR,
		Alternatives:    alternatives,
	}, nil
}

// calculateGrade assigns a letter grade based on pick rank.
// Grading scale:
//   - A+: Rank 1 (best card in pack)
//   - A:  Rank 2-3 (top 3)
//   - B:  Rank 4-5 (top 5)
//   - C:  Rank 6-8 (top 8)
//   - D:  Rank 9-10
//   - F:  Rank 11+ (poor pick)
func calculateGrade(rank, packSize int) string {
	switch {
	case rank == 1:
		return "A+"
	case rank <= 3:
		return "A"
	case rank <= 5:
		return "B"
	case rank <= 8:
		return "C"
	case rank <= 10:
		return "D"
	default:
		return "F"
	}
}

// SerializeAlternatives converts alternatives to JSON string for database storage.
func SerializeAlternatives(alternatives []Alternative) (string, error) {
	data, err := json.Marshal(alternatives)
	if err != nil {
		return "", fmt.Errorf("marshal alternatives: %w", err)
	}
	return string(data), nil
}

// DeserializeAlternatives converts JSON string back to alternatives slice.
func DeserializeAlternatives(jsonStr string) ([]Alternative, error) {
	var alternatives []Alternative
	if err := json.Unmarshal([]byte(jsonStr), &alternatives); err != nil {
		return nil, fmt.Errorf("unmarshal alternatives: %w", err)
	}
	return alternatives, nil
}

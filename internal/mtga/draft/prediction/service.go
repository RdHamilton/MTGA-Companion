package prediction

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service handles win rate prediction for draft decks
type Service struct {
	draftRepo   repository.DraftRepository
	ratingsRepo repository.DraftRatingsRepository
	setCardRepo repository.SetCardRepository
	db          *sql.DB
}

// NewService creates a new prediction service
func NewService(db *sql.DB) *Service {
	return &Service{
		draftRepo:   repository.NewDraftRepository(db),
		ratingsRepo: repository.NewDraftRatingsRepository(db),
		setCardRepo: repository.NewSetCardRepository(db),
		db:          db,
	}
}

// PredictSessionWinRate calculates and stores the win rate prediction for a draft session
func (s *Service) PredictSessionWinRate(ctx context.Context, sessionID string) (*DeckPrediction, error) {
	// 1. Get all picks for the session
	picks, err := s.draftRepo.GetPicksBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get picks: %w", err)
	}

	if len(picks) == 0 {
		return nil, fmt.Errorf("no picks found for session %s", sessionID)
	}

	// 2. Get the session to determine set code and format
	session, err := s.draftRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// 3. Build card list with ratings
	cards := []Card{}
	cardsSeen := make(map[string]bool) // Track duplicates

	for _, pick := range picks {
		// Skip if we've already added this card (handles duplicate picks)
		cardKey := pick.CardID
		if cardsSeen[cardKey] {
			continue
		}
		cardsSeen[cardKey] = true

		// Try to get card rating from 17Lands data
		rating, err := s.ratingsRepo.GetCardRatingByArenaID(ctx, session.SetCode, "QuickDraft", pick.CardID)

		// If 17Lands data not available, try Scryfall data as fallback
		if err != nil || rating == nil {
			scryfallCard, scryfallErr := s.setCardRepo.GetCardByArenaID(ctx, pick.CardID)

			if scryfallErr != nil || scryfallCard == nil {
				// No data available from either source - use complete baseline
				log.Printf("Warning: No card data found for Arena ID %s, using complete baseline", pick.CardID)
				cards = append(cards, Card{
					Name:   pick.CardID,
					CMC:    0,
					Color:  "C",
					GIHWR:  0.50,
					Rarity: "common",
				})
				continue
			}

			// Use Scryfall data with baseline GIHWR
			color := parseCardColor(scryfallCard.Colors)
			log.Printf("Using Scryfall data for %s (no 17Lands rating)", scryfallCard.Name)

			cards = append(cards, Card{
				Name:   scryfallCard.Name,
				CMC:    scryfallCard.CMC,
				Color:  color,
				GIHWR:  0.50, // Baseline 50% when no 17Lands data
				Rarity: scryfallCard.Rarity,
			})
			continue
		}

		// We have 17Lands data - use it
		color := "C"
		if len(rating.Color) > 0 {
			color = string(rating.Color[0])
		}

		cmc := estimateCMC(rating)

		cards = append(cards, Card{
			Name:   rating.Name,
			CMC:    cmc,
			Color:  color,
			GIHWR:  rating.GIHWR / 100.0,
			Rarity: rating.Rarity,
		})
	}

	// 4. Calculate prediction
	prediction, err := PredictWinRate(cards)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate prediction: %w", err)
	}

	// 5. Store prediction in database
	err = s.storePrediction(sessionID, prediction)
	if err != nil {
		return nil, fmt.Errorf("failed to store prediction: %w", err)
	}

	return prediction, nil
}

// GetSessionPrediction retrieves the stored prediction for a draft session
func (s *Service) GetSessionPrediction(sessionID string) (*DeckPrediction, error) {
	session, err := s.draftRepo.GetSession(context.Background(), sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.PredictedWinRate == nil {
		return nil, fmt.Errorf("no prediction found for session %s", sessionID)
	}

	// Parse factors from JSON
	var factors *PredictionFactors
	if session.PredictionFactors != nil {
		factors, err = FromJSON(*session.PredictionFactors)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prediction factors: %w", err)
		}
	} else {
		factors = &PredictionFactors{}
	}

	return &DeckPrediction{
		PredictedWinRate:    *session.PredictedWinRate,
		PredictedWinRateMin: *session.PredictedWinRateMin,
		PredictedWinRateMax: *session.PredictedWinRateMax,
		Factors:             *factors,
		PredictedAt:         *session.PredictedAt,
	}, nil
}

// storePrediction saves the prediction to the database
func (s *Service) storePrediction(sessionID string, prediction *DeckPrediction) error {
	factorsJSON, err := prediction.Factors.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert factors to JSON: %w", err)
	}

	query := `
		UPDATE draft_sessions
		SET predicted_win_rate = ?,
		    predicted_win_rate_min = ?,
		    predicted_win_rate_max = ?,
		    prediction_factors = ?,
		    predicted_at = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err = s.db.Exec(query,
		prediction.PredictedWinRate,
		prediction.PredictedWinRateMin,
		prediction.PredictedWinRateMax,
		factorsJSON,
		prediction.PredictedAt,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session with prediction: %w", err)
	}

	return nil
}

// parseCardColor converts a color slice from Scryfall into a single-character color code
func parseCardColor(colors []string) string {
	if len(colors) == 0 {
		return "C" // Colorless
	}

	// Return first color (simplified - multicolor cards get first color)
	return colors[0]
}

// estimateCMC estimates the converted mana cost from card data
// This is a placeholder - ideally we'd have this data in the database
func estimateCMC(rating *seventeenlands.CardRating) int {
	// Use ALSA (Average Last Seen At) as a rough proxy
	// Early picks (ALSA 1-3) are often higher CMC bombs
	// Late picks (ALSA 10+) are often cheap cards
	// This is very approximate!

	if rating.ALSA < 2.0 {
		return 5 // Likely a bomb rare
	} else if rating.ALSA < 4.0 {
		return 4
	} else if rating.ALSA < 7.0 {
		return 3
	} else {
		return 2
	}
}

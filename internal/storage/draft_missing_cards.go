package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// GetMissingCardsAnalysis calculates which cards from the initial pack have been taken by other players.
func (s *Service) GetMissingCardsAnalysis(ctx context.Context, sessionID string, packNum, pickNum int) (*models.MissingCardsAnalysis, error) {
	log.Printf("[GetMissingCardsAnalysis] Called for session=%s, pack=%d, pick=%d", sessionID, packNum, pickNum)

	// Can't calculate missing cards for P1P1 (no initial pack to compare)
	if pickNum <= 1 {
		log.Printf("[GetMissingCardsAnalysis] Pick 1 - returning empty analysis")
		return &models.MissingCardsAnalysis{
			SessionID:    sessionID,
			PackNumber:   packNum,
			PickNumber:   pickNum,
			MissingCards: []models.MissingCard{},
			TotalMissing: 0,
			BombsMissing: 0,
		}, nil
	}

	// Get the initial pack (pick 1) for this pack number
	initialPack, err := s.DraftRepo().GetPack(ctx, sessionID, packNum, 1)
	if err != nil || initialPack == nil {
		// Pack data not available - return empty analysis
		log.Printf("[GetMissingCardsAnalysis] Initial pack not found for session %s, pack %d: %v", sessionID, packNum, err)
		return &models.MissingCardsAnalysis{
			SessionID:    sessionID,
			PackNumber:   packNum,
			PickNumber:   pickNum,
			MissingCards: []models.MissingCard{},
			TotalMissing: 0,
			BombsMissing: 0,
		}, nil
	}
	log.Printf("[GetMissingCardsAnalysis] Found initial pack with %d cards", len(initialPack.CardIDs))

	// Get the current pack
	currentPack, err := s.DraftRepo().GetPack(ctx, sessionID, packNum, pickNum)
	if err != nil || currentPack == nil {
		// Current pack not available - return empty analysis
		log.Printf("[GetMissingCardsAnalysis] Current pack not found for session %s, pack %d, pick %d: %v", sessionID, packNum, pickNum, err)
		return &models.MissingCardsAnalysis{
			SessionID:    sessionID,
			PackNumber:   packNum,
			PickNumber:   pickNum,
			MissingCards: []models.MissingCard{},
			TotalMissing: 0,
			BombsMissing: 0,
		}, nil
	}
	log.Printf("[GetMissingCardsAnalysis] Found current pack with %d cards", len(currentPack.CardIDs))

	// Get all picks made from this pack so far
	allPicks, err := s.DraftRepo().GetPicksBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get picks: %w", err)
	}

	// Filter picks for this specific pack
	pickedByMe := make([]string, 0)
	for _, pick := range allPicks {
		if pick.PackNumber == packNum && pick.PickNumber < pickNum {
			pickedByMe = append(pickedByMe, pick.CardID)
			log.Printf("[GetMissingCardsAnalysis] My pick P%dP%d: Card %s", pick.PackNumber+1, pick.PickNumber, pick.CardID)
		}
	}

	// Calculate missing cards: Initial - Current - PickedByMe
	missingCardIDs := calculateMissingCards(initialPack.CardIDs, currentPack.CardIDs, pickedByMe)
	log.Printf("[GetMissingCardsAnalysis] Calculation: Initial=%d cards, Current=%d cards, PickedByMe=%d cards → Missing=%d cards",
		len(initialPack.CardIDs), len(currentPack.CardIDs), len(pickedByMe), len(missingCardIDs))

	// Get session info for ratings
	session, err := s.DraftRepo().GetSession(ctx, sessionID)
	if err != nil || session == nil {
		log.Printf("[GetMissingCardsAnalysis] Session not found for %s: %v", sessionID, err)
		return &models.MissingCardsAnalysis{
			SessionID:    sessionID,
			PackNumber:   packNum,
			PickNumber:   pickNum,
			MissingCards: []models.MissingCard{},
			TotalMissing: 0,
			BombsMissing: 0,
		}, nil
	}

	// Fetch all ratings for this set once
	allRatings, _, err := s.DraftRatingsRepo().GetCardRatings(ctx, session.SetCode, session.DraftType)
	if err != nil {
		log.Printf("Warning: Failed to get ratings for %s/%s: %v", session.SetCode, session.DraftType, err)
		// Continue without ratings - we can still show missing cards without win rates
	}

	// Build a map for fast lookup
	ratingsMap := make(map[string]*struct {
		GIHWR float64
		ALSA  float64
		Tier  string
	})
	for i := range allRatings {
		rating := &allRatings[i]
		tier := calculateTier(rating.GIHWR)
		// Convert MTGAID to string for map key
		cardIDStr := fmt.Sprintf("%d", rating.MTGAID)
		ratingsMap[cardIDStr] = &struct {
			GIHWR float64
			ALSA  float64
			Tier  string
		}{
			GIHWR: rating.GIHWR,
			ALSA:  rating.ALSA,
			Tier:  tier,
		}
	}

	// Fetch ratings and details for missing cards
	missingCards := make([]models.MissingCard, 0, len(missingCardIDs))
	bombsCount := 0

	for _, cardID := range missingCardIDs {
		// Get card info
		card, err := s.SetCardRepo().GetCardByArenaID(ctx, cardID)
		if err != nil || card == nil {
			log.Printf("Warning: Failed to get card %s: %v", cardID, err)
			continue
		}

		// Get rating from map
		var gihwr, alsa float64
		var tier string
		if ratingData, ok := ratingsMap[cardID]; ok {
			gihwr = ratingData.GIHWR
			alsa = ratingData.ALSA
			tier = ratingData.Tier

			// Count bombs (A+ or S tier)
			if tier == "S" || tier == "A+" {
				bombsCount++
			}
		}

		// Calculate wheel probability using ALSA data
		wheelProb := calculateWheelProbability(pickNum, alsa)

		// Estimate when it was picked (cards likely picked in GIHWR order)
		pickedAt := estimatePickedAt(pickNum, gihwr)

		missingCards = append(missingCards, models.MissingCard{
			CardID:           cardID,
			CardName:         card.Name,
			GIHWR:            gihwr,
			Tier:             tier,
			PickedAt:         pickedAt,
			WheelProbability: wheelProb,
		})
	}

	analysis := &models.MissingCardsAnalysis{
		SessionID:    sessionID,
		PackNumber:   packNum,
		PickNumber:   pickNum,
		InitialCards: initialPack.CardIDs,
		CurrentCards: currentPack.CardIDs,
		PickedByMe:   pickedByMe,
		MissingCards: missingCards,
		TotalMissing: len(missingCards),
		BombsMissing: bombsCount,
	}
	log.Printf("[GetMissingCardsAnalysis] Returning analysis with %d missing cards, %d bombs", len(missingCards), bombsCount)
	return analysis, nil
}

// calculateMissingCards returns card IDs that are in initial but not in current or picked.
func calculateMissingCards(initial, current, picked []string) []string {
	// Create sets for faster lookup
	currentSet := make(map[string]bool)
	for _, id := range current {
		currentSet[id] = true
	}

	pickedSet := make(map[string]bool)
	for _, id := range picked {
		pickedSet[id] = true
	}

	// Find cards that are missing
	missing := make([]string, 0)
	for _, id := range initial {
		if !currentSet[id] && !pickedSet[id] {
			missing = append(missing, id)
		}
	}

	return missing
}

// calculateTier converts GIHWR to tier rating.
func calculateTier(gihwr float64) string {
	if gihwr >= 62 {
		return "S"
	} else if gihwr >= 60 {
		return "A+"
	} else if gihwr >= 58 {
		return "A"
	} else if gihwr >= 56 {
		return "B"
	} else if gihwr >= 54 {
		return "C"
	} else if gihwr >= 52 {
		return "D"
	}
	return "F"
}

// calculateWheelProbability estimates the probability a card will wheel back.
// Uses ALSA (Average Last Seen At) data if available.
func calculateWheelProbability(pickNum int, alsa float64) float64 {
	// If ALSA data is unavailable or below threshold, return 0
	if alsa < 2.0 {
		return 0.0
	}

	// Simple heuristic: if ALSA is greater than current pick, it might wheel
	// The closer ALSA is to pickNum + 8 (when it comes back), the higher the probability
	packSize := 15.0
	expectedReturn := float64(pickNum) + (packSize - float64(pickNum))

	if alsa > expectedReturn {
		// Card is typically seen late enough that it might wheel
		probability := (alsa - expectedReturn) * 10.0
		if probability > 100 {
			probability = 100
		}
		return probability
	}

	return 0.0
}

// estimatePickedAt estimates when a card was likely picked based on its win rate.
// Higher win rate cards are picked earlier.
func estimatePickedAt(currentPick int, gihwr float64) int {
	// Simple heuristic: cards with GIHWR > 60% likely picked in first 3 picks
	if gihwr > 60 {
		return min(currentPick-1, 1) // Likely pick 1-2
	} else if gihwr > 55 {
		return min(currentPick-1, 3) // Likely pick 2-4
	} else if gihwr > 50 {
		return min(currentPick-1, 5) // Likely pick 3-6
	}
	return min(currentPick-1, currentPick-2) // Later picks
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

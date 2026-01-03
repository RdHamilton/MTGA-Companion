package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/metrics"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/grading"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/insights"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/pickquality"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft/prediction"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftFacade handles all draft-related operations.
type DraftFacade struct {
	services *Services

	// Track in-flight ratings fetches
	ratingsEnsureMu sync.Mutex
	inFlightRatings map[string]bool

	// Track in-flight card fetches
	cardFetchMu     sync.Mutex
	inFlightFetches map[string]bool
}

// NewDraftFacade creates a new DraftFacade with the given services.
func NewDraftFacade(services *Services) *DraftFacade {
	return &DraftFacade{
		services:        services,
		inFlightRatings: make(map[string]bool),
		inFlightFetches: make(map[string]bool),
	}
}

// GetActiveDraftSessions returns all active draft sessions.
func (d *DraftFacade) GetActiveDraftSessions(ctx context.Context) ([]*models.DraftSession, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	sessions, err := d.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get active draft sessions: %v", err)}
	}

	log.Printf("🎯 [GetActiveDraftSessions] Returning %d active session(s)", len(sessions))
	for i, s := range sessions {
		log.Printf("🎯 [GetActiveDraftSessions] Session %d: ID=%s, Status=%s, SetCode=%s, TotalPicks=%d", i, s.ID, s.Status, s.SetCode, s.TotalPicks)

		// Ensure 17Lands ratings are fetched for this draft
		if s.SetCode != "" && s.DraftType != "" {
			go d.ensureRatingsForDraft(s.SetCode, s.DraftType)
		}
	}

	return sessions, nil
}

// ensureRatingsForDraft ensures 17Lands ratings are cached for a draft.
// Runs in background to avoid blocking the UI. Uses mutex to prevent duplicate fetches.
func (d *DraftFacade) ensureRatingsForDraft(setCode, eventName string) {
	ctx := context.Background()
	key := fmt.Sprintf("%s-%s", setCode, eventName)

	// Check if already being fetched
	d.ratingsEnsureMu.Lock()
	if d.inFlightRatings[key] {
		d.ratingsEnsureMu.Unlock()
		return // Already fetching
	}

	// Check if ratings are already cached and fresh
	_, lastFetch, err := d.services.Storage.DraftRatingsRepo().GetCardRatings(ctx, setCode, eventName)
	if err == nil && !lastFetch.IsZero() && time.Since(lastFetch) < 24*time.Hour {
		d.ratingsEnsureMu.Unlock()
		return // Already cached and fresh
	}

	// Mark as in-flight
	d.inFlightRatings[key] = true
	d.ratingsEnsureMu.Unlock()

	// Fetch ratings (outside mutex)
	defer func() {
		d.ratingsEnsureMu.Lock()
		delete(d.inFlightRatings, key)
		d.ratingsEnsureMu.Unlock()
	}()

	log.Printf("[ensureRatingsForDraft] Fetching 17Lands ratings for %s / %s", setCode, eventName)
	err = d.services.RatingsFetcher.FetchAndCacheRatings(ctx, setCode, eventName)
	if err != nil {
		log.Printf("[ensureRatingsForDraft] Failed to fetch ratings: %v", err)
	} else {
		log.Printf("[ensureRatingsForDraft] Successfully cached ratings for %s / %s", setCode, eventName)
	}
}

// GetCompletedDraftSessions returns recently completed draft sessions.
func (d *DraftFacade) GetCompletedDraftSessions(ctx context.Context, limit int) ([]*models.DraftSession, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	if limit <= 0 {
		limit = 20 // Default limit
	}

	sessions, err := d.services.Storage.DraftRepo().GetCompletedSessions(ctx, limit)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get completed draft sessions: %v", err)}
	}

	return sessions, nil
}

// GetDraftSession returns a draft session by ID.
func (d *DraftFacade) GetDraftSession(ctx context.Context, sessionID string) (*models.DraftSession, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	session, err := d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft session: %v", err)}
	}

	return session, nil
}

// GetDraftPicks returns all picks for a draft session.
func (d *DraftFacade) GetDraftPicks(ctx context.Context, sessionID string) ([]*models.DraftPickSession, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	log.Printf("[GetDraftPicks] Called for session %s", sessionID)

	var picks []*models.DraftPickSession
	err := storage.RetryOnBusy(func() error {
		var err error
		picks, err = d.services.Storage.DraftRepo().GetPicksBySession(ctx, sessionID)
		return err
	})
	if err != nil {
		log.Printf("[GetDraftPicks] Error getting picks: %v", err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft picks: %v", err)}
	}

	log.Printf("[GetDraftPicks] Found %d picks for session %s", len(picks), sessionID)

	// Fetch card images for picked cards synchronously (serialized to avoid database locks)
	if len(picks) > 0 {
		var session *models.DraftSession
		err := storage.RetryOnBusy(func() error {
			var err error
			session, err = d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
			return err
		})
		if err != nil {
			log.Printf("[GetDraftPicks] Error getting session: %v", err)
		} else if session == nil {
			log.Printf("[GetDraftPicks] Session is nil")
		} else {
			log.Printf("[GetDraftPicks] Fetching cards for %d picks (SetCode=%s, DraftType=%s)", len(picks), session.SetCode, session.DraftType)
			d.fetchCardsForPicksSync(ctx, session.SetCode, session.DraftType, picks)
			log.Printf("[GetDraftPicks] Finished fetching cards")
		}
	}

	return picks, nil
}

// GetDraftDeckMetrics calculates comprehensive statistics for drafted cards.
func (d *DraftFacade) GetDraftDeckMetrics(ctx context.Context, sessionID string) (*models.DeckMetrics, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	log.Printf("[GetDraftDeckMetrics] Called for session %s", sessionID)

	// Get session to retrieve SetCode
	var session *models.DraftSession
	err := storage.RetryOnBusy(func() error {
		var err error
		session, err = d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
		return err
	})
	if err != nil {
		log.Printf("[GetDraftDeckMetrics] Error getting session: %v", err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft session: %v", err)}
	}
	if session == nil {
		log.Printf("[GetDraftDeckMetrics] Session not found: %s", sessionID)
		return nil, &AppError{Message: "Draft session not found"}
	}

	// Get all picks for the session
	var picks []*models.DraftPickSession
	err = storage.RetryOnBusy(func() error {
		var err error
		picks, err = d.services.Storage.DraftRepo().GetPicksBySession(ctx, sessionID)
		return err
	})
	if err != nil {
		log.Printf("[GetDraftDeckMetrics] Error getting picks: %v", err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft picks: %v", err)}
	}

	if len(picks) == 0 {
		log.Printf("[GetDraftDeckMetrics] No picks found for session %s", sessionID)
		// Return empty metrics
		return &models.DeckMetrics{
			DistributionAll:          make([]int, 7),
			DistributionCreatures:    make([]int, 7),
			DistributionNoncreatures: make([]int, 7),
			TypeBreakdown:            make(map[string]int),
			ColorDistribution:        make(map[string]int),
			ColorCounts:              make(map[string]int),
		}, nil
	}

	// Get all picked cards
	pickedCards := make([]models.SetCard, 0, len(picks))
	for _, pick := range picks {
		var card *models.SetCard
		err = storage.RetryOnBusy(func() error {
			var err error
			card, err = d.services.SetFetcher.GetCardByArenaID(ctx, pick.CardID)
			return err
		})
		if err != nil {
			log.Printf("[GetDraftDeckMetrics] Error getting card %s: %v", pick.CardID, err)
			continue
		}
		if card != nil {
			pickedCards = append(pickedCards, *card)
		}
	}

	log.Printf("[GetDraftDeckMetrics] Calculating metrics for %d cards", len(pickedCards))

	// Calculate metrics
	metrics := models.CalculateDeckMetrics(pickedCards)

	log.Printf("[GetDraftDeckMetrics] Metrics calculated: Total=%d, Creatures=%d, AvgCMC=%.2f",
		metrics.TotalCards, metrics.CreatureCount, metrics.CMCAverage)

	return metrics, nil
}

// GetDraftPerformanceMetrics returns performance metrics for draft operations.
func (d *DraftFacade) GetDraftPerformanceMetrics(ctx context.Context) *metrics.DraftStats {
	if d.services.DraftMetrics == nil {
		// Return empty stats if metrics not initialized
		return &metrics.DraftStats{}
	}
	return d.services.DraftMetrics.GetStats()
}

// ResetDraftPerformanceMetrics resets all performance metrics.
func (d *DraftFacade) ResetDraftPerformanceMetrics(ctx context.Context) {
	if d.services.DraftMetrics != nil {
		d.services.DraftMetrics.Reset()
	}
}

// fetchCardsForPicksSync fetches card metadata for all picked cards using 17Lands ratings.
// This ensures card images are available even if Scryfall doesn't have Arena IDs yet.
// Fetches cards serially (one at a time) to avoid database lock contention.
// Uses global mutex to ensure only one card fetch happens at a time across all requests.
func (d *DraftFacade) fetchCardsForPicksSync(ctx context.Context, setCode, eventName string, picks []*models.DraftPickSession) {
	log.Printf("[fetchCardsForPicksSync] Starting to fetch %d cards for %s/%s", len(picks), setCode, eventName)

	for i, pick := range picks {
		log.Printf("[fetchCardsForPicksSync] Processing pick %d/%d: CardID=%s", i+1, len(picks), pick.CardID)

		// Check if card is already cached (fast check, with retry)
		var cachedCard *models.SetCard
		err := storage.RetryOnBusy(func() error {
			var err error
			cachedCard, err = d.services.SetFetcher.GetCardByArenaID(ctx, pick.CardID)
			return err
		})
		if err == nil && cachedCard != nil {
			log.Printf("[fetchCardsForPicksSync] Card %s already cached (Name=%s)", pick.CardID, cachedCard.Name)
			continue // Already cached
		}

		// Acquire global lock to ensure serial card fetching
		d.cardFetchMu.Lock()

		// Double-check cache after acquiring lock (another request might have fetched it)
		err = storage.RetryOnBusy(func() error {
			var err error
			cachedCard, err = d.services.SetFetcher.GetCardByArenaID(ctx, pick.CardID)
			return err
		})
		if err == nil && cachedCard != nil {
			log.Printf("[fetchCardsForPicksSync] Card %s cached after lock (Name=%s)", pick.CardID, cachedCard.Name)
			d.cardFetchMu.Unlock()
			continue // Already cached
		}

		// Get card name from 17Lands ratings (with retry)
		log.Printf("[fetchCardsForPicksSync] Looking up rating for card %s in %s/%s", pick.CardID, setCode, eventName)
		var rating *seventeenlands.CardRating
		err = storage.RetryOnBusy(func() error {
			var err error
			rating, err = d.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, setCode, eventName, pick.CardID)
			return err
		})
		if err != nil {
			log.Printf("[fetchCardsForPicksSync] Error getting rating for card %s: %v", pick.CardID, err)
			d.cardFetchMu.Unlock()
			continue
		}
		if rating == nil || rating.Name == "" {
			log.Printf("[fetchCardsForPicksSync] No rating or name found for card %s", pick.CardID)
			d.cardFetchMu.Unlock()
			continue // No rating or name available
		}

		// Fetch card from Scryfall by name (with lock held to serialize database writes)
		log.Printf("[fetchCardsForPicksSync] Fetching card from Scryfall: %s (ID: %s)", rating.Name, pick.CardID)
		var card *models.SetCard
		err = storage.RetryOnBusy(func() error {
			var err error
			card, err = d.services.SetFetcher.FetchCardByName(ctx, setCode, rating.Name, pick.CardID)
			return err
		})
		if err != nil {
			log.Printf("[fetchCardsForPicksSync] Failed to fetch card %s (ID: %s): %v", rating.Name, pick.CardID, err)
		} else {
			log.Printf("[fetchCardsForPicksSync] Successfully fetched and cached card: %s (ID: %s)", rating.Name, pick.CardID)
			if card != nil {
				log.Printf("[fetchCardsForPicksSync] Card details: Name=%s, ImageURL=%s", card.Name, card.ImageURL)
			}
		}

		// Release lock after card is fetched and saved
		d.cardFetchMu.Unlock()

		// Small delay to avoid hammering Scryfall API
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("[fetchCardsForPicksSync] Finished fetching cards for %s/%s", setCode, eventName)
}

// GetDraftPacks returns all packs for a draft session.
func (d *DraftFacade) GetDraftPacks(ctx context.Context, sessionID string) ([]*models.DraftPackSession, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var packs []*models.DraftPackSession
	err := storage.RetryOnBusy(func() error {
		var err error
		packs, err = d.services.Storage.DraftRepo().GetPacksBySession(ctx, sessionID)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get draft packs: %v", err)}
	}

	return packs, nil
}

// GetMissingCards analyzes which cards from the initial pack have been taken by other players.
func (d *DraftFacade) GetMissingCards(ctx context.Context, sessionID string, packNum, pickNum int) (*models.MissingCardsAnalysis, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	var analysis *models.MissingCardsAnalysis
	err := storage.RetryOnBusy(func() error {
		var err error
		analysis, err = d.services.Storage.GetMissingCardsAnalysis(ctx, sessionID, packNum, pickNum)
		return err
	})
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get missing cards: %v", err)}
	}

	return analysis, nil
}

// AnalyzeSessionPickQuality calculates pick quality for all picks in a draft session.
// This should be called after a draft session is completed to analyze all picks.
func (d *DraftFacade) AnalyzeSessionPickQuality(ctx context.Context, sessionID string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	// Get session to get set code and draft format
	session, err := d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get session: %v", err)}
	}
	if session == nil {
		return &AppError{Message: "Session not found"}
	}

	// Get all picks for this session
	picks, err := d.services.Storage.DraftRepo().GetPicksBySession(ctx, sessionID)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get picks: %v", err)}
	}

	// Create pick quality analyzer
	analyzer := pickquality.NewAnalyzer(
		d.services.Storage.DraftRatingsRepo(),
		d.services.Storage.SetCardRepo(),
	)

	// Analyze each pick
	for _, pick := range picks {
		// Get the pack for this pick
		pack, err := d.services.Storage.DraftRepo().GetPack(ctx, sessionID, pick.PackNumber, pick.PickNumber)
		if err != nil || pack == nil {
			log.Printf("Warning: Could not get pack for pick %d (P%dP%d): %v", pick.ID, pick.PackNumber+1, pick.PickNumber+1, err)
			continue
		}

		// Ensure we have card image data (FetchCardByName checks cache first)
		pickedRating, err := d.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, session.SetCode, session.EventName, pick.CardID)
		if err == nil && pickedRating != nil && pickedRating.Name != "" {
			card, err := d.services.SetFetcher.FetchCardByName(ctx, session.SetCode, pickedRating.Name, pick.CardID)
			if err != nil {
				log.Printf("Warning: Failed to fetch card %s (ID: %s) by name: %v", pickedRating.Name, pick.CardID, err)
			} else if card != nil {
				log.Printf("✓ Fetched/cached card: %s (ID: %s)", pickedRating.Name, pick.CardID)
			}
		}

		// Analyze pick quality
		quality, err := analyzer.AnalyzePick(ctx, session.SetCode, session.EventName, pack.CardIDs, pick.CardID)
		if err != nil {
			log.Printf("Warning: Could not analyze pick %d: %v", pick.ID, err)
			continue
		}

		// Serialize alternatives to JSON
		alternativesJSON, err := pickquality.SerializeAlternatives(quality.Alternatives)
		if err != nil {
			log.Printf("Warning: Could not serialize alternatives for pick %d: %v", pick.ID, err)
			continue
		}

		// Update pick quality in database
		err = d.services.Storage.DraftRepo().UpdatePickQuality(
			ctx,
			pick.ID,
			quality.Grade,
			quality.Rank,
			quality.PackBestGIHWR,
			quality.PickedCardGIHWR,
			alternativesJSON,
		)
		if err != nil {
			log.Printf("Warning: Could not update pick quality for pick %d: %v", pick.ID, err)
		}
	}

	// Automatically recalculate draft grade after pick quality analysis
	// This ensures the grade reflects the actual pick quality data
	log.Printf("Pick quality analysis complete for session %s, recalculating draft grade...", sessionID)
	_, err = d.CalculateDraftGrade(ctx, sessionID)
	if err != nil {
		log.Printf("Warning: Failed to automatically recalculate draft grade: %v", err)
		// Don't return error - pick quality analysis still succeeded
	}

	return nil
}

// GetPickAlternatives returns alternative picks for a specific pick.
// Used to show tooltips with "You could have picked..." information.
func (d *DraftFacade) GetPickAlternatives(ctx context.Context, sessionID string, packNum, pickNum int) (*pickquality.PickQuality, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get session
	session, err := d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get session: %v", err)}
	}
	if session == nil {
		return nil, &AppError{Message: "Session not found"}
	}

	// Get the pick
	pick, err := d.services.Storage.DraftRepo().GetPickByNumber(ctx, sessionID, packNum, pickNum)
	if err != nil || pick == nil {
		return nil, &AppError{Message: "Pick not found"}
	}

	// If pick quality is already calculated, deserialize and return
	if pick.PickQualityGrade != nil && pick.AlternativesJSON != nil {
		alternatives, err := pickquality.DeserializeAlternatives(*pick.AlternativesJSON)
		if err != nil {
			return nil, &AppError{Message: fmt.Sprintf("Failed to parse alternatives: %v", err)}
		}

		return &pickquality.PickQuality{
			Grade:           *pick.PickQualityGrade,
			Rank:            *pick.PickQualityRank,
			PackBestGIHWR:   *pick.PackBestGIHWR,
			PickedCardGIHWR: *pick.PickedCardGIHWR,
			Alternatives:    alternatives,
		}, nil
	}

	// Otherwise, calculate it on the fly
	pack, err := d.services.Storage.DraftRepo().GetPack(ctx, sessionID, packNum, pickNum)
	if err != nil || pack == nil {
		return nil, &AppError{Message: "Pack not found"}
	}

	analyzer := pickquality.NewAnalyzer(
		d.services.Storage.DraftRatingsRepo(),
		d.services.Storage.SetCardRepo(),
	)

	quality, err := analyzer.AnalyzePick(ctx, session.SetCode, session.EventName, pack.CardIDs, pick.CardID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to analyze pick: %v", err)}
	}

	return quality, nil
}

// CalculateDraftGrade calculates and stores the overall grade for a draft session.
// This should be called after pick quality analysis is complete.
func (d *DraftFacade) CalculateDraftGrade(ctx context.Context, sessionID string) (*grading.DraftGrade, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Create grade calculator
	calculator := grading.NewCalculator(
		d.services.Storage.DraftRepo(),
		d.services.Storage.DraftRatingsRepo(),
		d.services.Storage.SetCardRepo(),
	)

	// Calculate grade
	grade, err := calculator.CalculateGrade(ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to calculate grade: %v", err)}
	}

	// Store grade in database
	err = d.services.Storage.DraftRepo().UpdateSessionGrade(
		ctx,
		sessionID,
		grade.OverallGrade,
		grade.OverallScore,
		grade.PickQualityScore,
		grade.ColorDisciplineScore,
		grade.DeckCompositionScore,
		grade.StrategicScore,
	)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to store grade: %v", err)}
	}

	return grade, nil
}

// GetDraftGrade retrieves the stored grade for a draft session.
// If the grade hasn't been calculated yet, returns nil.
func (d *DraftFacade) GetDraftGrade(ctx context.Context, sessionID string) (*grading.DraftGrade, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get session to check if grade exists
	session, err := d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get session: %v", err)}
	}
	if session == nil {
		return nil, &AppError{Message: "Session not found"}
	}

	// If no grade calculated yet, return nil
	if session.OverallGrade == nil {
		return nil, nil
	}

	// Return stored grade
	return &grading.DraftGrade{
		OverallGrade:         *session.OverallGrade,
		OverallScore:         *session.OverallScore,
		PickQualityScore:     *session.PickQualityScore,
		ColorDisciplineScore: *session.ColorDisciplineScore,
		DeckCompositionScore: *session.DeckCompositionScore,
		StrategicScore:       *session.StrategicScore,
		// Best/worst picks and suggestions would need to be recalculated
		// or stored separately - for now just return the scores
	}, nil
}

// PredictDraftWinRate calculates and stores the win rate prediction for a draft session
func (d *DraftFacade) PredictDraftWinRate(ctx context.Context, sessionID string) (*prediction.DeckPrediction, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Create prediction service using storage repositories
	predictionService := prediction.NewService(
		d.services.Storage.DraftRepo(),
		d.services.Storage.DraftRatingsRepo(),
		d.services.Storage.SetCardRepo(),
	)

	// Calculate prediction
	pred, err := predictionService.PredictSessionWinRate(ctx, sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to predict win rate: %v", err)}
	}

	return pred, nil
}

// GetDraftWinRatePrediction retrieves the stored win rate prediction for a draft session
func (d *DraftFacade) GetDraftWinRatePrediction(ctx context.Context, sessionID string) (*prediction.DeckPrediction, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Create prediction service using storage repositories
	predictionService := prediction.NewService(
		d.services.Storage.DraftRepo(),
		d.services.Storage.DraftRatingsRepo(),
		d.services.Storage.SetCardRepo(),
	)

	// Get stored prediction
	pred, err := predictionService.GetSessionPrediction(sessionID)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get prediction: %v", err)}
	}

	return pred, nil
}

// SetCardRefresher is a function type that refreshes set cards from external sources.
type SetCardRefresher func(ctx context.Context, setCode string) (count int, err error)

// RecalculateAllDraftGrades recalculates grades and predictions for all draft sessions.
// This is useful after fetching new 17Lands data to update grades with actual card ratings.
// It also backfills pick quality data and fetches missing card metadata.
func (d *DraftFacade) RecalculateAllDraftGrades(ctx context.Context, refreshSetCards SetCardRefresher) (int, error) {
	if d.services.Storage == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	log.Printf("Recalculating all draft grades...")

	// Get all draft sessions (both active and completed)
	activeSessions, err := d.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to get active sessions: %v", err)}
	}

	completedSessions, err := d.services.Storage.DraftRepo().GetCompletedSessions(ctx, 1000)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to get completed sessions: %v", err)}
	}

	// Combine all sessions
	allSessions := append(activeSessions, completedSessions...)
	log.Printf("Found %d draft sessions to recalculate", len(allSessions))

	// Fetch missing card metadata for all unique sets
	uniqueSets := make(map[string]bool)
	for _, session := range allSessions {
		if session.SetCode != "" {
			uniqueSets[session.SetCode] = true
		}
	}
	for setCode := range uniqueSets {
		log.Printf("Fetching card metadata for set %s...", setCode)
		count, err := refreshSetCards(ctx, setCode)
		if err != nil {
			log.Printf("Warning: Failed to fetch card metadata for %s: %v", setCode, err)
		} else if count == 0 {
			log.Printf("Scryfall returned 0 cards for %s (likely no Arena IDs yet). Will fetch by name as needed.", setCode)
		} else {
			log.Printf("Cached %d cards for set %s", count, setCode)
		}
	}

	// Recalculate grade and prediction for each session
	successCount := 0
	for _, session := range allSessions {
		log.Printf("Recalculating session %s (%s - %s)", session.ID, session.SetCode, session.DraftType)

		// Backfill pick quality data with latest ratings
		if session.SetCode != "" {
			err := d.backfillPickQualityData(ctx, session.ID, session.SetCode, session.DraftType)
			if err != nil {
				log.Printf("Warning: Failed to backfill pick quality for session %s: %v", session.ID, err)
			} else {
				log.Printf("✓ Backfilled pick quality data for session %s", session.ID)
			}
		}

		// Recalculate grade
		_, err := d.CalculateDraftGrade(ctx, session.ID)
		if err != nil {
			log.Printf("Warning: Failed to recalculate grade for session %s: %v", session.ID, err)
			continue
		}

		// Recalculate prediction
		_, err = d.PredictDraftWinRate(ctx, session.ID)
		if err != nil {
			log.Printf("Warning: Failed to recalculate prediction for session %s: %v", session.ID, err)
			// Don't fail - continue even if prediction fails
		}

		successCount++
	}

	log.Printf("Successfully recalculated %d/%d draft sessions", successCount, len(allSessions))
	return successCount, nil
}

// RecalculateDraftGradesForSet recalculates grades for all drafts with a specific set code.
// This is called after refreshing ratings to update existing draft grades with new data.
//
// Note: This method processes sessions synchronously. In typical usage, the number of drafts
// per set is small (< 50), so this completes quickly. The frontend shows download progress
// while this runs. If large-scale batch processing is needed, consider using a background
// job queue instead.
func (d *DraftFacade) RecalculateDraftGradesForSet(ctx context.Context, setCode string) (int, error) {
	if d.services.Storage == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	if setCode == "" {
		return 0, &AppError{Message: "Set code is required"}
	}

	log.Printf("Recalculating draft grades for set %s...", setCode)

	// Get all draft sessions for this set
	activeSessions, err := d.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to get active sessions: %v", err)}
	}

	// Limit to most recent 5000 completed sessions. This should be more than sufficient
	// for typical usage. If a user has more than 5000 drafts for a single set, older
	// sessions won't be recalculated. Consider pagination if this becomes an issue.
	const maxCompletedSessions = 5000
	completedSessions, err := d.services.Storage.DraftRepo().GetCompletedSessions(ctx, maxCompletedSessions)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to get completed sessions: %v", err)}
	}

	// Filter to only sessions with matching set code
	var matchingSessions []*models.DraftSession
	for _, session := range append(activeSessions, completedSessions...) {
		if session.SetCode == setCode {
			matchingSessions = append(matchingSessions, session)
		}
	}

	log.Printf("Found %d draft sessions for set %s to recalculate", len(matchingSessions), setCode)

	if len(matchingSessions) == 0 {
		return 0, nil
	}

	// Recalculate grade and prediction for each session
	successCount := 0
	for _, session := range matchingSessions {
		log.Printf("Recalculating session %s (%s - %s)", session.ID, session.SetCode, session.DraftType)

		// Backfill pick quality data with latest ratings
		err := d.backfillPickQualityData(ctx, session.ID, session.SetCode, session.DraftType)
		if err != nil {
			log.Printf("Warning: Failed to backfill pick quality for session %s: %v", session.ID, err)
		} else {
			log.Printf("✓ Backfilled pick quality data for session %s", session.ID)
		}

		// Recalculate grade
		_, err = d.CalculateDraftGrade(ctx, session.ID)
		if err != nil {
			log.Printf("Warning: Failed to recalculate grade for session %s: %v", session.ID, err)
			continue
		}

		// Recalculate prediction
		_, err = d.PredictDraftWinRate(ctx, session.ID)
		if err != nil {
			log.Printf("Warning: Failed to recalculate prediction for session %s: %v", session.ID, err)
			// Don't fail - continue even if prediction fails
		}

		successCount++
	}

	log.Printf("Successfully recalculated %d/%d draft sessions for set %s", successCount, len(matchingSessions), setCode)
	return successCount, nil
}

// backfillPickQualityData updates pick quality data for all picks in a session
// using the latest 17Lands card ratings.
func (d *DraftFacade) backfillPickQualityData(ctx context.Context, sessionID, setCode, draftFormat string) error {
	// Get all picks for this session
	picks, err := d.services.Storage.DraftRepo().GetPicksBySession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get picks: %w", err)
	}

	// Get all packs for this session
	packs, err := d.services.Storage.DraftRepo().GetPacksBySession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get packs: %w", err)
	}

	// Create pack lookup map
	packMap := make(map[string]*models.DraftPackSession)
	for _, pack := range packs {
		key := fmt.Sprintf("%d_%d", pack.PackNumber, pack.PickNumber)
		packMap[key] = pack
	}

	// Update each pick
	updatedCount := 0
	for _, pick := range picks {
		// Get corresponding pack
		packKey := fmt.Sprintf("%d_%d", pick.PackNumber, pick.PickNumber)
		pack, hasPack := packMap[packKey]
		if !hasPack {
			// No pack data - mark as N/A since we can't calculate pick quality without alternatives
			if err := d.services.Storage.DraftRepo().UpdatePickQuality(ctx, pick.ID, "N/A", 0, 0.0, 0.0, "[]"); err != nil {
				log.Printf("Warning: Failed to update pick quality for pick %d: %v", pick.ID, err)
			} else {
				updatedCount++
			}
			continue
		}

		// Get rating for picked card
		pickedRating, err := d.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, setCode, draftFormat, pick.CardID)

		var grade string
		if err != nil || pickedRating == nil {
			// No rating available for this card - mark as N/A
			grade = "N/A"

			// Save pick with N/A grade
			if err := d.services.Storage.DraftRepo().UpdatePickQuality(ctx, pick.ID, grade, 0, 0.0, 0.0, "[]"); err != nil {
				log.Printf("Warning: Failed to update pick quality for pick %d: %v", pick.ID, err)
			} else {
				updatedCount++
			}
			continue
		}

		// Ensure we have Scryfall card data for images (FetchCardByName checks cache first)
		if pickedRating.Name != "" {
			card, err := d.services.SetFetcher.FetchCardByName(ctx, setCode, pickedRating.Name, pick.CardID)
			if err != nil {
				log.Printf("Warning: Failed to fetch card %s (ID: %s) by name: %v", pickedRating.Name, pick.CardID, err)
			} else if card != nil {
				log.Printf("Fetched/cached card: %s (ID: %s)", pickedRating.Name, pick.CardID)
			}
		}

		// Get ratings for all cards in pack to find alternatives
		packRatings := make(map[string]float64)
		bestGIHWR := 0.0
		for _, cardID := range pack.CardIDs {
			rating, err := d.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, setCode, draftFormat, cardID)
			if err == nil && rating != nil {
				packRatings[cardID] = rating.GIHWR
				if rating.GIHWR > bestGIHWR {
					bestGIHWR = rating.GIHWR
				}
			}
		}

		// Calculate pick quality grade
		gihwr := pickedRating.GIHWR
		grade = calculatePickGrade(gihwr, bestGIHWR)

		// Find rank (1 = best pick in pack)
		rank := 1
		for _, cardGIHWR := range packRatings {
			if cardGIHWR > gihwr {
				rank++
			}
		}

		// Build alternatives JSON (top 3 cards)
		type alternative struct {
			CardID string  `json:"card_id"`
			GIHWR  float64 `json:"gihwr"`
		}
		alternatives := []alternative{}
		for cardID, cardGIHWR := range packRatings {
			if cardID != pick.CardID && cardGIHWR >= gihwr {
				alternatives = append(alternatives, alternative{CardID: cardID, GIHWR: cardGIHWR})
			}
		}
		// Sort by GIHWR descending
		for i := 0; i < len(alternatives)-1; i++ {
			for j := i + 1; j < len(alternatives); j++ {
				if alternatives[j].GIHWR > alternatives[i].GIHWR {
					alternatives[i], alternatives[j] = alternatives[j], alternatives[i]
				}
			}
		}
		// Take top 3
		if len(alternatives) > 3 {
			alternatives = alternatives[:3]
		}

		alternativesJSON := ""
		if len(alternatives) > 0 {
			jsonBytes, _ := json.Marshal(alternatives)
			alternativesJSON = string(jsonBytes)
		}

		// Update pick in database
		err = d.services.Storage.DraftRepo().UpdatePickQuality(
			ctx,
			pick.ID,
			grade,
			rank,
			bestGIHWR,
			gihwr,
			alternativesJSON,
		)
		if err != nil {
			log.Printf("Warning: Failed to update pick quality for pick %d: %v", pick.ID, err)
			continue
		}

		updatedCount++
	}

	log.Printf("Updated pick quality for %d/%d picks in session %s", updatedCount, len(picks), sessionID)
	return nil
}

// calculatePickGrade converts GIHWR to a letter grade.
func calculatePickGrade(gihwr, bestGIHWR float64) string {
	// Calculate relative quality (how close to best pick)
	if bestGIHWR == 0 {
		return "C"
	}

	ratio := gihwr / bestGIHWR

	if ratio >= 0.95 {
		return "A+"
	} else if ratio >= 0.85 {
		return "A"
	} else if ratio >= 0.75 {
		return "A-"
	} else if ratio >= 0.65 {
		return "B+"
	} else if ratio >= 0.55 {
		return "B"
	} else if ratio >= 0.45 {
		return "B-"
	} else if ratio >= 0.35 {
		return "C+"
	} else if ratio >= 0.25 {
		return "C"
	} else if ratio >= 0.15 {
		return "C-"
	} else {
		return "D"
	}
}

// ReplayStatusChecker is a function type that checks if replay is active.
type ReplayStatusChecker func() (isActive bool, err error)

// FixDraftSessionStatuses updates draft sessions that should be marked as completed
// based on their pick counts (42 for Quick Draft, 45 for Premier Draft).
func (d *DraftFacade) FixDraftSessionStatuses(ctx context.Context, checkReplayActive ReplayStatusChecker) (int, error) {
	if d.services.Storage == nil {
		return 0, &AppError{Message: "Database not initialized"}
	}

	// Don't fix statuses while replay is active - replay mode keeps sessions as "in_progress" for testing
	isActive, err := checkReplayActive()
	if err == nil && isActive {
		log.Println("[FixDraftSessionStatuses] Skipping - replay is active")
		return 0, nil
	}

	// Get all active sessions
	activeSessions, err := d.services.Storage.DraftRepo().GetActiveSessions(ctx)
	if err != nil {
		return 0, &AppError{Message: fmt.Sprintf("Failed to get active sessions: %v", err)}
	}

	updated := 0
	for _, session := range activeSessions {
		// Get picks for this session
		picks, err := d.services.Storage.DraftRepo().GetPicksBySession(ctx, session.ID)
		if err != nil {
			log.Printf("Failed to get picks for session %s: %v", session.ID, err)
			continue
		}

		// Use TotalPicks from session if available (calculated from first pack)
		expectedPicks := session.TotalPicks
		if expectedPicks == 0 {
			// Fallback: try to calculate from first pack
			packs, err := d.services.Storage.DraftRepo().GetPacksBySession(ctx, session.ID)
			if err == nil && len(packs) > 0 {
				for _, pack := range packs {
					if pack.PackNumber == 0 && pack.PickNumber == 1 {
						expectedPicks = len(pack.CardIDs) * 3
						break
					}
				}
			}
		}
		if expectedPicks == 0 {
			// Last resort fallback
			expectedPicks = 42
		}

		// If session has all expected picks, mark as completed
		if len(picks) >= expectedPicks {
			// Use the timestamp of the last pick as end time
			var endTime *time.Time
			if len(picks) > 0 {
				lastPickTime := picks[len(picks)-1].Timestamp
				endTime = &lastPickTime
			}

			err := d.services.Storage.DraftRepo().UpdateSessionStatus(ctx, session.ID, "completed", endTime)
			if err != nil {
				log.Printf("Failed to update session %s status: %v", session.ID, err)
				continue
			}

			log.Printf("Updated session %s to completed (%d/%d picks)", session.ID, len(picks), expectedPicks)
			updated++
		}
	}

	return updated, nil
}

// RepairDraftSession repairs a draft session by reconstructing missing first pick packs
// and recalculating TotalPicks from actual pack size.
func (d *DraftFacade) RepairDraftSession(ctx context.Context, sessionID string) error {
	if d.services.Storage == nil {
		return &AppError{Message: "Database not initialized"}
	}

	log.Printf("[RepairDraftSession] Repairing session %s", sessionID)

	// Get session
	session, err := d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
	if err != nil || session == nil {
		return &AppError{Message: fmt.Sprintf("Session not found: %v", err)}
	}

	// Get all picks
	picks, err := d.services.Storage.DraftRepo().GetPicksBySession(ctx, sessionID)
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to get picks: %v", err)}
	}

	// Reconstruct missing first picks for each pack
	for packNum := 0; packNum < 3; packNum++ {
		// Check if P_N_P1 exists
		existingPack, err := d.services.Storage.DraftRepo().GetPack(ctx, sessionID, packNum, 1)
		if err == nil && existingPack != nil {
			log.Printf("[RepairDraftSession] P%dP1 already exists, skipping", packNum+1)
			continue
		}

		// Get P_N_P2 (pack after first pick)
		pack2, err := d.services.Storage.DraftRepo().GetPack(ctx, sessionID, packNum, 2)
		if err != nil || pack2 == nil {
			log.Printf("[RepairDraftSession] No P%dP2 found, cannot reconstruct P%dP1", packNum+1, packNum+1)
			continue
		}

		// Find cards picked in P_N_P1
		var pickedCards []string
		for _, pick := range picks {
			if pick.PackNumber == packNum && pick.PickNumber == 1 {
				pickedCards = append(pickedCards, pick.CardID)
			}
		}

		if len(pickedCards) == 0 {
			log.Printf("[RepairDraftSession] No pick found for P%dP1, cannot reconstruct", packNum+1)
			continue
		}

		// Reconstruct: P1 = P2 + picked_cards
		reconstructedPack := &models.DraftPackSession{
			SessionID:  sessionID,
			PackNumber: packNum,
			PickNumber: 1,
			CardIDs:    append(append([]string{}, pack2.CardIDs...), pickedCards...),
			Timestamp:  pack2.Timestamp,
		}

		log.Printf("[RepairDraftSession] Reconstructing P%dP1 with %d cards (P2 had %d, picked %d)",
			packNum+1, len(reconstructedPack.CardIDs), len(pack2.CardIDs), len(pickedCards))

		if err := d.services.Storage.DraftRepo().SavePack(ctx, reconstructedPack); err != nil {
			log.Printf("[RepairDraftSession] Warning: Failed to save reconstructed P%dP1: %v", packNum+1, err)
		}
	}

	// Recalculate TotalPicks from first pack size
	firstPack, err := d.services.Storage.DraftRepo().GetPack(ctx, sessionID, 0, 1)
	if err == nil && firstPack != nil {
		correctTotalPicks := len(firstPack.CardIDs) * 3
		if correctTotalPicks != session.TotalPicks {
			log.Printf("[RepairDraftSession] Updating TotalPicks from %d to %d", session.TotalPicks, correctTotalPicks)

			// Update session TotalPicks using repository
			if err := d.services.Storage.DraftRepo().UpdateSessionTotalPicks(ctx, sessionID, correctTotalPicks); err != nil {
				return &AppError{Message: fmt.Sprintf("Failed to update TotalPicks: %v", err)}
			}
		}
	}

	log.Printf("[RepairDraftSession] Session %s repaired successfully", sessionID)
	return nil
}

// GetFormatInsights generates aggregated insights for a draft format.
// Returns color rankings, top cards, format speed, and other meta analysis.
// Uses the Strategy pattern to apply format-specific analysis logic.
func (d *DraftFacade) GetFormatInsights(ctx context.Context, setCode, draftFormat string) (*insights.FormatInsights, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Create insights analyzer with format-specific strategy
	analyzer := insights.NewAnalyzerForFormat(d.services.Storage, draftFormat)

	// Generate insights using the appropriate strategy
	formatInsights, err := analyzer.AnalyzeFormat(ctx, setCode, draftFormat)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to analyze format: %v", err)}
	}

	return formatInsights, nil
}

// GetArchetypeCards returns top cards for a specific color combination (archetype).
// colors parameter should be like "W", "UB", "WUR", etc.
// Uses the Strategy pattern to apply format-specific filtering and analysis.
func (d *DraftFacade) GetArchetypeCards(ctx context.Context, setCode, draftFormat, colors string) (*insights.ArchetypeCards, error) {
	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Create insights analyzer with format-specific strategy
	analyzer := insights.NewAnalyzerForFormat(d.services.Storage, draftFormat)

	// Get archetype cards using the appropriate strategy
	archetypeCards, err := analyzer.GetArchetypeCards(ctx, setCode, draftFormat, colors)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get archetype cards: %v", err)}
	}

	return archetypeCards, nil
}

// PackCardWithRating represents a card in the current pack with its rating info.
type PackCardWithRating struct {
	ArenaID       string   `json:"arena_id"`
	Name          string   `json:"name"`
	ImageURL      string   `json:"image_url"`
	Rarity        string   `json:"rarity"`
	Colors        []string `json:"colors"`
	ManaCost      string   `json:"mana_cost"`
	CMC           int      `json:"cmc"`
	TypeLine      string   `json:"type_line"`
	GIHWR         float64  `json:"gihwr"`          // Games In Hand Win Rate
	ALSA          float64  `json:"alsa"`           // Average Last Seen At
	Tier          string   `json:"tier"`           // S, A, B, C, D, F
	IsRecommended bool     `json:"is_recommended"` // True if this is the recommended pick
	Score         float64  `json:"score"`          // Recommendation score (0-1)
	Reasoning     string   `json:"reasoning"`      // Why this card is recommended
}

// CurrentPackResponse contains the current pack with recommendations.
type CurrentPackResponse struct {
	SessionID       string               `json:"session_id"`
	PackNumber      int                  `json:"pack_number"` // 0-indexed
	PickNumber      int                  `json:"pick_number"` // 0-indexed
	PackLabel       string               `json:"pack_label"`  // Human readable, e.g., "Pack 1, Pick 3"
	Cards           []PackCardWithRating `json:"cards"`
	RecommendedCard *PackCardWithRating  `json:"recommended_card"` // The top recommendation
	PoolColors      []string             `json:"pool_colors"`      // Current color identity of pool
	PoolSize        int                  `json:"pool_size"`        // Number of cards picked so far
}

// GetCurrentPackWithRecommendation returns the current pack cards with ratings and pick recommendation.
func (d *DraftFacade) GetCurrentPackWithRecommendation(ctx context.Context, sessionID string) (*CurrentPackResponse, error) {
	log.Printf("[GetCurrentPackWithRecommendation] Called for session: %s", sessionID)

	if d.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	// Get session info
	session, err := d.services.Storage.DraftRepo().GetSession(ctx, sessionID)
	if err != nil {
		log.Printf("[GetCurrentPackWithRecommendation] Failed to get session: %v", err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get session: %v", err)}
	}
	if session == nil {
		log.Printf("[GetCurrentPackWithRecommendation] Session not found: %s", sessionID)
		return nil, &AppError{Message: "Session not found"}
	}
	log.Printf("[GetCurrentPackWithRecommendation] Found session: %s, SetCode: %s, Status: %s", session.ID, session.SetCode, session.Status)

	// Get all packs for this session
	packs, err := d.services.Storage.DraftRepo().GetPacksBySession(ctx, sessionID)
	if err != nil {
		log.Printf("[GetCurrentPackWithRecommendation] Failed to get packs: %v", err)
		return nil, &AppError{Message: fmt.Sprintf("Failed to get packs: %v", err)}
	}
	log.Printf("[GetCurrentPackWithRecommendation] Found %d packs for session %s", len(packs), sessionID)
	if len(packs) == 0 {
		return nil, &AppError{Message: "No pack data available"}
	}

	// Find the latest pack (highest pack number, then highest pick number)
	var currentPack *models.DraftPackSession
	for _, pack := range packs {
		if currentPack == nil {
			currentPack = pack
			continue
		}
		if pack.PackNumber > currentPack.PackNumber ||
			(pack.PackNumber == currentPack.PackNumber && pack.PickNumber > currentPack.PickNumber) {
			currentPack = pack
		}
	}

	if currentPack == nil || len(currentPack.CardIDs) == 0 {
		log.Printf("[GetCurrentPackWithRecommendation] Current pack is empty for session %s", sessionID)
		return nil, &AppError{Message: "Current pack is empty"}
	}
	log.Printf("[GetCurrentPackWithRecommendation] Current pack: P%d/P%d with %d cards", currentPack.PackNumber, currentPack.PickNumber, len(currentPack.CardIDs))

	// Get already picked cards for pool analysis
	picks, err := d.services.Storage.DraftRepo().GetPicksBySession(ctx, sessionID)
	if err != nil {
		log.Printf("Warning: Could not get picks for session %s: %v", sessionID, err)
		picks = []*models.DraftPickSession{}
	}

	// Analyze pool colors from picked cards
	// Use DraftType for rating lookup (e.g., "QuickDraft") instead of EventName (e.g., "QuickDraft_TLA_20251127")
	poolColors := d.analyzePoolColors(ctx, session.SetCode, session.DraftType, picks)

	// Build response with card ratings
	cards := make([]PackCardWithRating, 0, len(currentPack.CardIDs))
	var bestCard *PackCardWithRating
	bestScore := -1.0

	for _, cardID := range currentPack.CardIDs {
		cardWithRating := d.getCardWithRating(ctx, session.SetCode, session.DraftType, cardID, poolColors, len(picks))
		if cardWithRating != nil {
			cards = append(cards, *cardWithRating)

			// Track best recommendation
			if cardWithRating.Score > bestScore {
				bestScore = cardWithRating.Score
				bestCard = cardWithRating
			}
		} else {
			log.Printf("[GetCurrentPackWithRecommendation] Card %s not found in set %s", cardID, session.SetCode)
		}
	}
	log.Printf("[GetCurrentPackWithRecommendation] Built %d cards with ratings for session %s", len(cards), sessionID)

	// Mark the recommended card
	if bestCard != nil {
		bestCard.IsRecommended = true
		// Update in the cards slice too
		for i := range cards {
			if cards[i].ArenaID == bestCard.ArenaID {
				cards[i].IsRecommended = true
				break
			}
		}
	}

	// Sort cards by score (highest first)
	for i := 0; i < len(cards)-1; i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[j].Score > cards[i].Score {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}

	return &CurrentPackResponse{
		SessionID:       sessionID,
		PackNumber:      currentPack.PackNumber,
		PickNumber:      currentPack.PickNumber,
		PackLabel:       fmt.Sprintf("Pack %d, Pick %d", currentPack.PackNumber+1, currentPack.PickNumber+1),
		Cards:           cards,
		RecommendedCard: bestCard,
		PoolColors:      poolColors,
		PoolSize:        len(picks),
	}, nil
}

// analyzePoolColors determines the color identity of already picked cards.
func (d *DraftFacade) analyzePoolColors(ctx context.Context, setCode, eventName string, picks []*models.DraftPickSession) []string {
	colorCounts := make(map[string]int)

	for _, pick := range picks {
		// Get card info from ratings
		rating, err := d.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, setCode, eventName, pick.CardID)
		if err != nil || rating == nil {
			continue
		}

		// Parse colors from rating
		colors := parseColors(rating.Color)
		for _, c := range colors {
			colorCounts[c]++
		}
	}

	// Return colors with at least 2 cards (or any if pool is small)
	threshold := 2
	if len(picks) < 6 {
		threshold = 1
	}

	result := []string{}
	for color, count := range colorCounts {
		if count >= threshold {
			result = append(result, color)
		}
	}

	return result
}

// getCardWithRating builds a PackCardWithRating from card ID.
func (d *DraftFacade) getCardWithRating(ctx context.Context, setCode, eventName, cardID string, poolColors []string, poolSize int) *PackCardWithRating {
	// Get card rating from 17Lands data
	rating, err := d.services.Storage.DraftRatingsRepo().GetCardRatingByArenaID(ctx, setCode, eventName, cardID)
	if err != nil || rating == nil {
		log.Printf("Warning: No rating found for card %s", cardID)
		return nil
	}

	// Try to get card info from SetCard repo
	setCard, _ := d.services.Storage.SetCardRepo().GetCardByArenaID(ctx, cardID)

	// Calculate tier from GIHWR (multiply by 100 as calculateTier expects percentage)
	tier := calculateTier(rating.GIHWR * 100)

	// Parse colors
	colors := parseColors(rating.Color)

	// Calculate recommendation score
	score, reasoning := d.calculatePickScore(rating, colors, poolColors, poolSize)

	card := &PackCardWithRating{
		ArenaID:       cardID,
		Name:          rating.Name,
		Rarity:        rating.Rarity,
		Colors:        colors,
		ManaCost:      rating.Color,       // Color field contains mana cost info
		GIHWR:         rating.GIHWR * 100, // Convert to percentage
		ALSA:          rating.ALSA,
		Tier:          tier,
		IsRecommended: false,
		Score:         score,
		Reasoning:     reasoning,
	}

	// Add image URL and type info if we have SetCard data
	if setCard != nil {
		card.ImageURL = setCard.ImageURL
		// Construct TypeLine from Types array
		if len(setCard.Types) > 0 {
			card.TypeLine = setCard.Types[0]
			for i := 1; i < len(setCard.Types); i++ {
				card.TypeLine += " " + setCard.Types[i]
			}
		}
		card.CMC = setCard.CMC
		if len(setCard.Colors) > 0 {
			card.Colors = setCard.Colors
		}
		if setCard.ManaCost != "" {
			card.ManaCost = setCard.ManaCost
		}
	}

	return card
}

// calculatePickScore calculates recommendation score for a card (0-1).
func (d *DraftFacade) calculatePickScore(rating *seventeenlands.CardRating, cardColors, poolColors []string, poolSize int) (float64, string) {
	reasons := []string{}

	// Factor 1: Raw card quality from GIHWR (50% weight)
	// GIHWR typically ranges from 45% to 65%
	qualityScore := (rating.GIHWR - 0.45) / 0.20 // Maps 45-65% to 0-1
	if qualityScore < 0 {
		qualityScore = 0
	} else if qualityScore > 1 {
		qualityScore = 1
	}

	if qualityScore >= 0.7 {
		reasons = append(reasons, "high win rate card")
	}

	// Factor 2: Color fit (30% weight for picks 4+, 10% for early picks)
	colorScore := 1.0 // Default: colorless or no pool colors yet
	colorWeight := 0.10

	if poolSize >= 3 && len(poolColors) > 0 {
		colorWeight = 0.30 // After first few picks, color matters more

		if len(cardColors) == 0 {
			colorScore = 0.8 // Colorless is good but not optimal
		} else {
			matchingColors := 0
			for _, cc := range cardColors {
				for _, pc := range poolColors {
					if cc == pc {
						matchingColors++
						break
					}
				}
			}

			if matchingColors == len(cardColors) {
				colorScore = 1.0 // Perfect fit
				reasons = append(reasons, "matches your colors")
			} else if matchingColors > 0 {
				colorScore = 0.6 // Partial fit
				reasons = append(reasons, "partially on-color")
			} else {
				colorScore = 0.2 // Off-color
				reasons = append(reasons, "off-color")
			}
		}
	}

	// Factor 3: Pick availability (20% weight) - ALSA indicates how late cards wheel
	// Lower ALSA = card gets picked earlier = better
	alsaScore := 1.0 - ((rating.ALSA - 1.0) / 13.0) // Maps 1-14 to 1-0
	if alsaScore < 0 {
		alsaScore = 0
	} else if alsaScore > 1 {
		alsaScore = 1
	}

	if alsaScore >= 0.7 {
		reasons = append(reasons, "highly contested")
	}

	// Calculate weighted score
	qualityWeight := 0.50
	alsaWeight := 0.20
	// Adjust weights to sum to 1
	totalWeight := qualityWeight + colorWeight + alsaWeight
	score := (qualityScore*qualityWeight + colorScore*colorWeight + alsaScore*alsaWeight) / totalWeight

	// Build reasoning string
	reasoning := ""
	if len(reasons) > 0 {
		reasoning = reasons[0]
		for i := 1; i < len(reasons); i++ {
			if i == len(reasons)-1 {
				reasoning += " and " + reasons[i]
			} else {
				reasoning += ", " + reasons[i]
			}
		}
		reasoning = "This card " + reasoning + "."
	}

	return score, reasoning
}

// parseColors extracts colors from a mana cost string like "{2}{W}{U}".
func parseColors(manaCost string) []string {
	colorSet := make(map[string]bool)
	for _, c := range manaCost {
		switch c {
		case 'W':
			colorSet["W"] = true
		case 'U':
			colorSet["U"] = true
		case 'B':
			colorSet["B"] = true
		case 'R':
			colorSet["R"] = true
		case 'G':
			colorSet["G"] = true
		}
	}

	colors := make([]string, 0, len(colorSet))
	for c := range colorSet {
		colors = append(colors, c)
	}
	return colors
}

package draftdata

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// UpdaterConfig configures the draft data updater.
type UpdaterConfig struct {
	// ScryfallClient for fetching set information
	ScryfallClient *scryfall.Client

	// SeventeenLandsClient for fetching draft statistics
	SeventeenLandsClient *seventeenlands.Client

	// Storage for caching draft statistics
	Storage *storage.Service

	// NewSetThreshold is the age threshold for considering a set "new" (default: 30 days)
	NewSetThreshold time.Duration

	// StaleThreshold is the age threshold for considering data stale (default: 24 hours)
	StaleThreshold time.Duration

	// DateRangeWindow is how far back to fetch data (default: 7 days)
	DateRangeWindow time.Duration
}

// Updater handles periodic updates of draft statistics.
type Updater struct {
	scryfall        *scryfall.Client
	seventeenlands  *seventeenlands.Client
	storage         *storage.Service
	newSetThreshold time.Duration
	staleThreshold  time.Duration
	dateRangeWindow time.Duration
}

// NewUpdater creates a new draft data updater.
func NewUpdater(config UpdaterConfig) (*Updater, error) {
	if config.ScryfallClient == nil {
		return nil, fmt.Errorf("ScryfallClient is required")
	}
	if config.SeventeenLandsClient == nil {
		return nil, fmt.Errorf("SeventeenLandsClient is required")
	}
	if config.Storage == nil {
		return nil, fmt.Errorf("Storage is required")
	}

	// Set defaults
	if config.NewSetThreshold == 0 {
		config.NewSetThreshold = 30 * 24 * time.Hour // 30 days
	}
	if config.StaleThreshold == 0 {
		config.StaleThreshold = 24 * time.Hour // 24 hours
	}
	if config.DateRangeWindow == 0 {
		config.DateRangeWindow = 7 * 24 * time.Hour // 7 days
	}

	return &Updater{
		scryfall:        config.ScryfallClient,
		seventeenlands:  config.SeventeenLandsClient,
		storage:         config.Storage,
		newSetThreshold: config.NewSetThreshold,
		staleThreshold:  config.StaleThreshold,
		dateRangeWindow: config.DateRangeWindow,
	}, nil
}

// ActiveSet represents a set that should receive regular updates.
type ActiveSet struct {
	Code       string
	Name       string
	ReleasedAt time.Time
	IsNew      bool // Released within NewSetThreshold
}

// GetActiveSets returns sets that should receive regular updates.
// Active sets include:
// - Sets released within the NewSetThreshold (default: 30 days)
// - Sets that appear in Standard-legal card searches
func (u *Updater) GetActiveSets(ctx context.Context) ([]ActiveSet, error) {
	// Get all sets from Scryfall
	setList, err := u.scryfall.GetSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sets from Scryfall: %w", err)
	}

	now := time.Now()
	var activeSets []ActiveSet

	for _, set := range setList.Data {
		// Skip digital-only sets (Arena-only sets won't have paper draft data)
		if set.Digital {
			continue
		}

		// Skip sets without release date
		if set.ReleasedAt == "" {
			continue
		}

		// Parse release date
		releaseDate, err := time.Parse("2006-01-02", set.ReleasedAt)
		if err != nil {
			log.Printf("Warning: failed to parse release date for set %s: %v", set.Code, err)
			continue
		}

		// Only include sets that have been released
		if releaseDate.After(now) {
			continue
		}

		// Filter by set type - only include draftable sets
		// expansion: Regular expansions (e.g., BLB, MKM)
		// core: Core sets (e.g., M21)
		// masters: Masters sets (e.g., MH3)
		// draft_innovation: Special draft sets
		isDraftableType := set.SetType == "expansion" ||
			set.SetType == "core" ||
			set.SetType == "masters" ||
			set.SetType == "draft_innovation"

		if !isDraftableType {
			continue
		}

		// Calculate age
		age := now.Sub(releaseDate)

		// Include if:
		// 1. New set (released within threshold), OR
		// 2. Set is from the last 2 years (reasonable window for Standard-adjacent draft)
		isNew := age <= u.newSetThreshold
		isRecentEnough := age <= 2*365*24*time.Hour

		if isNew || isRecentEnough {
			activeSets = append(activeSets, ActiveSet{
				Code:       set.Code,
				Name:       set.Name,
				ReleasedAt: releaseDate,
				IsNew:      isNew,
			})
		}
	}

	return activeSets, nil
}

// UpdateResult contains the results of an update operation.
type UpdateResult struct {
	SetCode      string
	Success      bool
	CardRatings  int
	ColorRatings int
	Error        error
	Duration     time.Duration
}

// UpdateSet updates draft statistics for a specific set.
func (u *Updater) UpdateSet(ctx context.Context, setCode string) (*UpdateResult, error) {
	start := time.Now()
	result := &UpdateResult{
		SetCode: setCode,
	}

	// Calculate date range (last N days)
	endDate := time.Now()
	startDate := endDate.Add(-u.dateRangeWindow)

	// Format dates for 17Lands API (YYYY-MM-DD)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Fetch card ratings from 17Lands
	cardParams := seventeenlands.QueryParams{
		Expansion: setCode,
		Format:    "PremierDraft", // Primary format
		StartDate: startDateStr,
		EndDate:   endDateStr,
	}

	cardRatings, err := u.seventeenlands.GetCardRatings(ctx, cardParams)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch card ratings: %w", err)
		return result, result.Error
	}

	// Save card ratings to database
	if len(cardRatings) > 0 {
		err = u.storage.SaveCardRatings(ctx, cardRatings, setCode, "PremierDraft", "", startDateStr, endDateStr)
		if err != nil {
			result.Error = fmt.Errorf("failed to save card ratings: %w", err)
			return result, result.Error
		}
		result.CardRatings = len(cardRatings)
	}

	// Wait for rate limit (1 second between requests)
	time.Sleep(1 * time.Second)

	// Fetch color ratings from 17Lands
	colorRatings, err := u.seventeenlands.GetColorRatings(ctx, cardParams)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch color ratings: %w", err)
		return result, result.Error
	}

	// Save color ratings to database
	if len(colorRatings) > 0 {
		err = u.storage.SaveColorRatings(ctx, colorRatings, setCode, "PremierDraft", startDateStr, endDateStr)
		if err != nil {
			result.Error = fmt.Errorf("failed to save color ratings: %w", err)
			return result, result.Error
		}
		result.ColorRatings = len(colorRatings)
	}

	result.Success = true
	result.Duration = time.Since(start)
	return result, nil
}

// UpdateActiveSkipped sets updates active sets that are stale.
// Returns the number of sets updated, skipped, and any errors.
func (u *Updater) UpdateActiveSets(ctx context.Context) (updated int, skipped int, results []UpdateResult, err error) {
	// Get active sets
	activeSets, err := u.GetActiveSets(ctx)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to get active sets: %w", err)
	}

	// Check which sets are stale
	staleSets, err := u.storage.GetStaleCardRatings(ctx, u.staleThreshold)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to check stale ratings: %w", err)
	}

	// Create map of stale expansions for quick lookup
	staleMap := make(map[string]bool)
	for _, stale := range staleSets {
		staleMap[stale.Expansion] = true
	}

	// Update stale active sets
	for _, set := range activeSets {
		// Skip if not stale (unless it's a brand new set with no data)
		if !staleMap[set.Code] && !set.IsNew {
			skipped++
			continue
		}

		log.Printf("Updating draft statistics for set: %s (%s)", set.Code, set.Name)

		result, err := u.UpdateSet(ctx, set.Code)
		if err != nil {
			log.Printf("Failed to update set %s: %v", set.Code, err)
			results = append(results, *result)
			continue
		}

		log.Printf("Successfully updated %s: %d card ratings, %d color ratings (took %v)",
			set.Code, result.CardRatings, result.ColorRatings, result.Duration)
		results = append(results, *result)
		updated++

		// Rate limit: Wait 1 second between sets (already waited in UpdateSet, but add buffer)
		if set.Code != activeSets[len(activeSets)-1].Code {
			time.Sleep(1 * time.Second)
		}
	}

	return updated, skipped, results, nil
}

// CheckStaleness returns information about which sets have stale data.
func (u *Updater) CheckStaleness(ctx context.Context) ([]*storage.DraftCardRating, error) {
	return u.storage.GetStaleCardRatings(ctx, u.staleThreshold)
}

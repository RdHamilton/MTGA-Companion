package setcache

import (
	"context"
	"fmt"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// RatingsFetcher handles fetching and caching 17Lands ratings for draft sets.
type RatingsFetcher struct {
	seventeenLandsClient *seventeenlands.Client
	ratingsRepo          repository.DraftRatingsRepository
}

// NewRatingsFetcher creates a new ratings fetcher.
func NewRatingsFetcher(client *seventeenlands.Client, ratingsRepo repository.DraftRatingsRepository) *RatingsFetcher {
	return &RatingsFetcher{
		seventeenLandsClient: client,
		ratingsRepo:          ratingsRepo,
	}
}

// FetchAndCacheRatings fetches card and color ratings from 17Lands and caches them.
// draftFormat should be "PremierDraft" or "QuickDraft".
func (f *RatingsFetcher) FetchAndCacheRatings(ctx context.Context, setCode, draftFormat string) error {
	// Check if already cached
	isCached, err := f.ratingsRepo.IsSetRatingsCached(ctx, setCode, draftFormat)
	if err != nil {
		return fmt.Errorf("check if ratings cached: %w", err)
	}
	if isCached {
		return nil // Already cached
	}

	// Map MTGA set code to 17Lands expansion code
	expansion := mapSetCodeTo17Lands(setCode)

	// Fetch card ratings
	cardRatings, err := f.seventeenLandsClient.GetCardRatings(ctx, seventeenlands.QueryParams{
		Expansion: expansion,
		Format:    draftFormat,
	})
	if err != nil {
		return fmt.Errorf("fetch card ratings from 17Lands: %w", err)
	}

	// Fetch color ratings
	colorRatings, err := f.seventeenLandsClient.GetColorRatings(ctx, seventeenlands.QueryParams{
		Expansion: expansion,
		EventType: draftFormat,
	})
	if err != nil {
		return fmt.Errorf("fetch color ratings from 17Lands: %w", err)
	}

	// Save to database
	err = f.ratingsRepo.SaveSetRatings(ctx, setCode, draftFormat, cardRatings, colorRatings)
	if err != nil {
		return fmt.Errorf("save ratings to database: %w", err)
	}

	return nil
}

// GetCachedCardRatings retrieves cached card ratings for a set.
func (f *RatingsFetcher) GetCachedCardRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.CardRating, error) {
	ratings, _, err := f.ratingsRepo.GetCardRatings(ctx, setCode, draftFormat)
	return ratings, err
}

// GetCachedColorRatings retrieves cached color ratings for a set.
func (f *RatingsFetcher) GetCachedColorRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.ColorRating, error) {
	ratings, _, err := f.ratingsRepo.GetColorRatings(ctx, setCode, draftFormat)
	return ratings, err
}

// GetCardRating retrieves a specific card's rating by Arena ID.
func (f *RatingsFetcher) GetCardRating(ctx context.Context, setCode, draftFormat, arenaID string) (*seventeenlands.CardRating, error) {
	return f.ratingsRepo.GetCardRatingByArenaID(ctx, setCode, draftFormat, arenaID)
}

// RefreshRatings deletes and re-fetches ratings for a set.
func (f *RatingsFetcher) RefreshRatings(ctx context.Context, setCode, draftFormat string) error {
	// Delete existing cache
	if err := f.ratingsRepo.DeleteSetRatings(ctx, setCode, draftFormat); err != nil {
		return fmt.Errorf("delete existing ratings: %w", err)
	}

	// Fetch and cache again
	return f.FetchAndCacheRatings(ctx, setCode, draftFormat)
}

// mapSetCodeTo17Lands maps MTGA set codes to 17Lands expansion codes.
// Most are the same, but this allows for exceptions.
func mapSetCodeTo17Lands(setCode string) string {
	// Most codes are the same, just lowercase
	code := strings.ToUpper(setCode)

	// Handle special cases if needed
	specialCases := map[string]string{
		"TDM": "DSK", // Duskmourn may use different codes
	}

	if mapped, ok := specialCases[code]; ok {
		return mapped
	}

	return code
}

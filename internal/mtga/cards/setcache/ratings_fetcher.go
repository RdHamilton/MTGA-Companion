package setcache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/datasets"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// RatingsFetcher handles fetching and caching 17Lands ratings for draft sets.
type RatingsFetcher struct {
	seventeenLandsClient *seventeenlands.Client
	datasetService       *datasets.Service
	ratingsRepo          repository.DraftRatingsRepository
}

// NewRatingsFetcher creates a new ratings fetcher.
// Deprecated: Use NewRatingsFetcherWithDatasets instead.
func NewRatingsFetcher(client *seventeenlands.Client, ratingsRepo repository.DraftRatingsRepository) *RatingsFetcher {
	return &RatingsFetcher{
		seventeenLandsClient: client,
		ratingsRepo:          ratingsRepo,
	}
}

// NewRatingsFetcherWithDatasets creates a new ratings fetcher with dataset service support.
func NewRatingsFetcherWithDatasets(datasetService *datasets.Service, ratingsRepo repository.DraftRatingsRepository) *RatingsFetcher {
	return &RatingsFetcher{
		datasetService: datasetService,
		ratingsRepo:    ratingsRepo,
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
	log.Printf("[RatingsFetcher] Fetching card ratings for expansion=%s, format=%s", expansion, draftFormat)

	var cardRatings []seventeenlands.CardRating
	var dataSource string

	// Try new dataset service first (S3 or web API fallback)
	if f.datasetService != nil {
		log.Printf("[RatingsFetcher] Using dataset service (S3 + web API fallback)")
		cardRatings, err = f.datasetService.GetCardRatings(ctx, expansion, draftFormat)
		if err == nil && len(cardRatings) > 0 {
			dataSource = f.datasetService.GetDataSource(ctx, expansion, draftFormat)
			log.Printf("[RatingsFetcher] Successfully fetched %d ratings from dataset service (source: %s)", len(cardRatings), dataSource)
		} else {
			log.Printf("[RatingsFetcher] Dataset service failed or returned no data: %v", err)
		}
	}

	// Fallback to old client if dataset service failed or unavailable
	if f.datasetService == nil || len(cardRatings) == 0 {
		if f.seventeenLandsClient == nil {
			return fmt.Errorf("no data source available (dataset service and client both unavailable)")
		}

		log.Printf("[RatingsFetcher] Falling back to legacy API client")
		cardRatings, err = f.seventeenLandsClient.GetCardRatings(ctx, seventeenlands.QueryParams{
			Expansion: expansion,
			Format:    draftFormat,
		})
		if err != nil {
			log.Printf("[RatingsFetcher] Error fetching card ratings from legacy client: %v", err)
			return fmt.Errorf("fetch card ratings from 17Lands: %w", err)
		}
		dataSource = "api" // Legacy API
		log.Printf("[RatingsFetcher] Received %d card ratings from legacy API", len(cardRatings))
	}

	// Check if card ratings are empty (17Lands might not have data yet)
	if len(cardRatings) == 0 {
		return fmt.Errorf("17Lands returned no card ratings for set %s (%s) - data may not be available yet for this recently released set", setCode, draftFormat)
	}

	// Fetch color ratings (still using legacy client for now)
	// TODO: Implement color ratings in dataset service
	var colorRatings []seventeenlands.ColorRating
	if f.seventeenLandsClient != nil {
		endDate := getCurrentDateString()
		startDate := getDateDaysAgo(17)
		log.Printf("[RatingsFetcher] Fetching color ratings for expansion=%s, eventType=%s, dates=%s to %s", expansion, draftFormat, startDate, endDate)
		colorRatings, err = f.seventeenLandsClient.GetColorRatings(ctx, seventeenlands.QueryParams{
			Expansion: expansion,
			EventType: draftFormat,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			log.Printf("[RatingsFetcher] Error fetching color ratings: %v", err)
			// Don't fail if color ratings unavailable
			colorRatings = []seventeenlands.ColorRating{}
		} else {
			log.Printf("[RatingsFetcher] Received %d color ratings", len(colorRatings))
		}
	}

	// Save to database
	log.Printf("[RatingsFetcher] Saving %d card ratings and %d color ratings to database (source: %s)", len(cardRatings), len(colorRatings), dataSource)
	err = f.ratingsRepo.SaveSetRatings(ctx, setCode, draftFormat, cardRatings, colorRatings, dataSource)
	if err != nil {
		log.Printf("[RatingsFetcher] Error saving ratings to database: %v", err)
		return fmt.Errorf("save ratings to database: %w", err)
	}

	log.Printf("[RatingsFetcher] Successfully cached %d card ratings and %d color ratings for %s (%s) from %s", len(cardRatings), len(colorRatings), setCode, draftFormat, dataSource)
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

// getCurrentDateString returns today's date in YYYY-MM-DD format.
func getCurrentDateString() string {
	return time.Now().Format("2006-01-02")
}

// getDateDaysAgo returns the date N days ago in YYYY-MM-DD format.
func getDateDaysAgo(days int) string {
	return time.Now().AddDate(0, 0, -days).Format("2006-01-02")
}

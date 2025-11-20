package datasets

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// Service provides unified access to 17Lands card ratings from multiple sources.
type Service struct {
	downloader *Downloader
	parser     *CSVParser
	webScraper *WebScraper
}

// ServiceOptions configures the dataset service.
type ServiceOptions struct {
	DownloaderOptions DownloaderOptions
	WebScraperOptions WebScraperOptions
}

// DefaultServiceOptions returns default service options.
func DefaultServiceOptions() ServiceOptions {
	return ServiceOptions{
		DownloaderOptions: DefaultDownloaderOptions(),
		WebScraperOptions: DefaultWebScraperOptions(),
	}
}

// NewService creates a new dataset service.
func NewService(options ServiceOptions) (*Service, error) {
	// Create downloader
	downloader, err := NewDownloader(options.DownloaderOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create downloader: %w", err)
	}

	// Create web scraper
	webScraper := NewWebScraper(options.WebScraperOptions)

	// Create CSV parser
	parser := NewCSVParser()

	return &Service{
		downloader: downloader,
		parser:     parser,
		webScraper: webScraper,
	}, nil
}

// GetCardRatings fetches card ratings for a set and format.
// Strategy:
// 1. Try S3 public datasets first (recommended approach)
// 2. If S3 fails or dataset doesn't exist, fall back to web API
//
// This ensures we use the recommended datasets when available while
// still supporting newer sets like TLA that only have web API data.
func (s *Service) GetCardRatings(ctx context.Context, setCode, format string) ([]seventeenlands.CardRating, error) {
	log.Printf("[DatasetService] Fetching card ratings for %s / %s", setCode, format)

	// Strategy 1: Try S3 datasets first
	ratings, err := s.tryS3Dataset(ctx, setCode, format)
	if err == nil && len(ratings) > 0 {
		log.Printf("[DatasetService] Successfully fetched %d ratings from S3 datasets", len(ratings))
		return ratings, nil
	}

	// Log S3 failure but continue to fallback
	if err != nil {
		log.Printf("[DatasetService] S3 dataset unavailable: %v, falling back to web API", err)
	} else {
		log.Printf("[DatasetService] S3 dataset returned no data, falling back to web API")
	}

	// Strategy 2: Fall back to web API (for sets like TLA)
	ratings, err = s.webScraper.FetchCardRatings(ctx, setCode, format)
	if err != nil {
		return nil, fmt.Errorf("both S3 and web API failed: web API error: %w", err)
	}

	if len(ratings) == 0 {
		return nil, fmt.Errorf("no card ratings available for %s / %s", setCode, format)
	}

	log.Printf("[DatasetService] Successfully fetched %d ratings from web API (fallback)", len(ratings))
	return ratings, nil
}

// tryS3Dataset attempts to download and parse a dataset from S3.
func (s *Service) tryS3Dataset(ctx context.Context, setCode, format string) ([]seventeenlands.CardRating, error) {
	// Download dataset (or use cached)
	csvPath, err := s.downloader.DownloadDataset(ctx, setCode, format)
	if err != nil {
		return nil, fmt.Errorf("failed to download dataset: %w", err)
	}

	// Parse CSV to calculate card ratings
	// Note: Create new parser for each parse to avoid state conflicts
	parser := NewCSVParser()
	ratings, err := parser.ParseCSV(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return ratings, nil
}

// CheckDatasetAvailability checks if a dataset is available in S3 without downloading it.
func (s *Service) CheckDatasetAvailability(ctx context.Context, setCode, format string) (bool, error) {
	return s.webScraper.IsDatasetAvailable(ctx, setCode, format), nil
}

// ClearCache removes all cached datasets.
func (s *Service) ClearCache() error {
	return s.downloader.ClearCache()
}

// GetDataSource returns the data source used for a given set/format.
// Returns "s3" if S3 dataset is available, "web_api" otherwise.
func (s *Service) GetDataSource(ctx context.Context, setCode, format string) string {
	available := s.webScraper.IsDatasetAvailable(ctx, setCode, format)
	if available {
		return "s3"
	}
	return "web_api"
}

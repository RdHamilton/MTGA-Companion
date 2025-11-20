package datasets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

const (
	// Web API base URL
	WebAPIBaseURL = "https://www.17lands.com"

	// Default timeout
	WebAPITimeout = 30 * time.Second
)

// Default rate limit for web API (1 request per 2 seconds to be conservative)
var WebAPIRateLimit = rate.Every(2 * time.Second)

// WebScraper handles fallback data fetching from 17Lands web API for sets not in S3.
type WebScraper struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

// WebScraperOptions configures the web scraper.
type WebScraperOptions struct {
	// RateLimit controls request frequency (default: 1 req/2 seconds)
	RateLimit rate.Limit

	// Timeout for HTTP requests (default: 30 seconds)
	Timeout time.Duration

	// HTTPClient allows custom HTTP client
	HTTPClient *http.Client
}

// DefaultWebScraperOptions returns default web scraper options.
func DefaultWebScraperOptions() WebScraperOptions {
	return WebScraperOptions{
		RateLimit: WebAPIRateLimit,
		Timeout:   WebAPITimeout,
	}
}

// NewWebScraper creates a new web scraper for 17Lands API.
func NewWebScraper(options WebScraperOptions) *WebScraper {
	if options.RateLimit == 0 {
		options.RateLimit = WebAPIRateLimit
	}
	if options.Timeout == 0 {
		options.Timeout = WebAPITimeout
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: options.Timeout,
		}
	}

	return &WebScraper{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(options.RateLimit, 1),
	}
}

// FetchCardRatings fetches card ratings from the 17Lands web API.
// This is a fallback for sets not available in S3 datasets (like TLA).
func (w *WebScraper) FetchCardRatings(ctx context.Context, setCode, format string) ([]seventeenlands.CardRating, error) {
	// Build URL
	// Format: https://www.17lands.com/card_ratings/data?expansion=TLA&format=PremierDraft
	endpoint := fmt.Sprintf("%s/card_ratings/data", WebAPIBaseURL)
	params := url.Values{}
	params.Set("expansion", setCode)
	params.Set("format", format)

	fullURL := endpoint + "?" + params.Encode()
	log.Printf("[WebScraper] Fetching card ratings from web API: %s", fullURL)

	// Apply rate limiting
	if err := w.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	// Execute request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var ratings []seventeenlands.CardRating
	if err := json.Unmarshal(body, &ratings); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert decimal values (0-1) to percentages (0-100) for consistency with S3 parser
	// Web API returns win rates as decimals (0.5669), but our tier calculator expects percentages (56.69)
	for i := range ratings {
		ratings[i].GIHWR *= 100
		ratings[i].OHWR *= 100
		ratings[i].GPWR *= 100
		ratings[i].GDWR *= 100
		ratings[i].IHDWR *= 100
		ratings[i].GIHWRDelta *= 100
		ratings[i].OHWRDelta *= 100
		ratings[i].GDWRDelta *= 100
		ratings[i].IHDWRDelta *= 100
		ratings[i].PickRate *= 100
	}

	log.Printf("[WebScraper] Fetched %d card ratings from web API", len(ratings))
	return ratings, nil
}

// IsDatasetAvailable checks if a dataset exists in S3 without downloading it.
// Returns true if the dataset is available, false otherwise.
func (w *WebScraper) IsDatasetAvailable(ctx context.Context, setCode, format string) bool {
	// Build S3 URL
	filename := fmt.Sprintf("game_data_public.%s.%s.csv.gz", setCode, format)
	url := fmt.Sprintf("%s/%s", PublicDatasetsBaseURL, filename)

	// Create HEAD request to check if file exists
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		log.Printf("[WebScraper] Failed to create HEAD request: %v", err)
		return false
	}

	// Execute request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		log.Printf("[WebScraper] Failed to check dataset availability: %v", err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	// Dataset is available if status is 200 OK
	available := resp.StatusCode == http.StatusOK
	if available {
		log.Printf("[WebScraper] Dataset available in S3: %s", url)
	} else {
		log.Printf("[WebScraper] Dataset not available in S3 (status: %d): %s", resp.StatusCode, url)
	}

	return available
}

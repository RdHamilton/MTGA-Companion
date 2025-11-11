package seventeenlands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// Request timeout
	DefaultTimeout = 30 * time.Second

	// Backoff settings
	InitialBackoff = 2 * time.Second
	MaxBackoff     = 60 * time.Second
	BackoffFactor  = 2.0
)

// APIBase is the base URL for 17Lands API (variable for testing).
var APIBase = "https://www.17lands.com"

// Conservative rate limit: 1 request per second
var DefaultRateLimit = rate.Every(1 * time.Second)

// CacheStorage defines the interface for caching 17Lands data.
// Implementations should store and retrieve draft statistics for fallback when API is unavailable.
type CacheStorage interface {
	// SaveCardRatings caches card ratings
	SaveCardRatings(ctx context.Context, ratings []CardRating, expansion, format, colors, startDate, endDate string) error

	// GetCardRatingsForSet retrieves cached card ratings for a set
	GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]CardRating, time.Time, error)

	// SaveColorRatings caches color ratings
	SaveColorRatings(ctx context.Context, ratings []ColorRating, expansion, eventType, startDate, endDate string) error

	// GetColorRatings retrieves cached color ratings
	GetColorRatings(ctx context.Context, expansion, eventType string) ([]ColorRating, time.Time, error)
}

// Client provides access to 17Lands draft statistics.
type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	cache      CacheStorage // Optional cache for fallback
	stats      *ClientStats
	statsMu    sync.RWMutex

	// Backoff tracking
	backoff         time.Duration
	lastFailureTime time.Time
	backoffMu       sync.Mutex
}

// ClientOptions configures the 17Lands client.
type ClientOptions struct {
	// RateLimit controls request frequency (default: 1 req/second)
	RateLimit rate.Limit

	// Timeout for HTTP requests (default: 30 seconds)
	Timeout time.Duration

	// HTTPClient allows custom HTTP client
	HTTPClient *http.Client

	// Cache provides fallback storage when API is unavailable (optional)
	Cache CacheStorage
}

// DefaultClientOptions returns conservative default options.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   DefaultTimeout,
	}
}

// NewClient creates a new 17Lands API client with conservative rate limiting.
func NewClient(options ClientOptions) *Client {
	if options.RateLimit == 0 {
		options.RateLimit = DefaultRateLimit
	}
	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: options.Timeout,
		}
	}

	return &Client{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(options.RateLimit, 1),
		cache:      options.Cache,
		stats:      &ClientStats{},
		backoff:    InitialBackoff,
	}
}

// GetCardRatings fetches card performance statistics for a set.
// If the API is unavailable, falls back to cached data.
func (c *Client) GetCardRatings(ctx context.Context, params QueryParams) ([]CardRating, error) {
	// Validate required parameters
	if params.Expansion == "" {
		return nil, &APIError{
			Type:    ErrInvalidParams,
			Message: "expansion is required",
		}
	}
	if params.Format == "" {
		return nil, &APIError{
			Type:    ErrInvalidParams,
			Message: "format is required",
		}
	}

	// Build URL
	endpoint := fmt.Sprintf("%s/card_ratings/data", APIBase)
	queryParams := url.Values{}
	queryParams.Set("expansion", params.Expansion)
	queryParams.Set("format", params.Format)

	if params.StartDate != "" {
		queryParams.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		queryParams.Set("end_date", params.EndDate)
	}

	// Prepare colors for cache key
	colors := ""
	if len(params.Colors) > 0 {
		colors = strings.Join(params.Colors, "")
		queryParams.Set("colors", colors)
	}

	fullURL := endpoint + "?" + queryParams.Encode()

	// Try API request
	body, err := c.doRequest(ctx, fullURL)
	if err != nil {
		// API failed - try cache fallback if available
		if c.cache != nil {
			log.Printf("[17Lands] API unavailable, attempting cache fallback: %v", err)
			cached, cachedAt, cacheErr := c.cache.GetCardRatingsForSet(ctx, params.Expansion, params.Format, colors)
			if cacheErr == nil && len(cached) > 0 {
				age := time.Since(cachedAt)
				log.Printf("[17Lands] Using cached card ratings (age: %v, count: %d)", age, len(cached))

				// Update stats
				c.updateStats(func(s *ClientStats) {
					s.CachedResponses++
				})

				return cached, nil
			}
			log.Printf("[17Lands] Cache miss or empty: %v", cacheErr)
		}

		// No cache available or cache failed
		return nil, &APIError{
			Type:    ErrStatsUnavailable,
			Message: "17Lands API unavailable and no cached data available",
			Err:     err,
		}
	}

	// Parse response
	var ratings []CardRating
	if err := json.Unmarshal(body, &ratings); err != nil {
		return nil, &APIError{
			Type:    ErrParseError,
			Message: "failed to parse card ratings response",
			Err:     err,
		}
	}

	// Cache successful response if cache is available
	if c.cache != nil && len(ratings) > 0 {
		if err := c.cache.SaveCardRatings(ctx, ratings, params.Expansion, params.Format, colors, params.StartDate, params.EndDate); err != nil {
			log.Printf("[17Lands] Failed to cache card ratings: %v", err)
			// Don't fail the request if caching fails
		}
	}

	return ratings, nil
}

// GetColorRatings fetches color combination performance statistics.
// If the API is unavailable, falls back to cached data.
func (c *Client) GetColorRatings(ctx context.Context, params QueryParams) ([]ColorRating, error) {
	// Validate required parameters
	if params.Expansion == "" {
		return nil, &APIError{
			Type:    ErrInvalidParams,
			Message: "expansion is required",
		}
	}
	if params.EventType == "" {
		return nil, &APIError{
			Type:    ErrInvalidParams,
			Message: "event_type is required for color ratings",
		}
	}

	// Build URL
	endpoint := fmt.Sprintf("%s/color_ratings/data", APIBase)
	queryParams := url.Values{}
	queryParams.Set("expansion", params.Expansion)
	queryParams.Set("event_type", params.EventType)

	if params.StartDate != "" {
		queryParams.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		queryParams.Set("end_date", params.EndDate)
	}
	if params.CombineSplash {
		queryParams.Set("combine_splash", "true")
	}

	fullURL := endpoint + "?" + queryParams.Encode()

	// Try API request
	body, err := c.doRequest(ctx, fullURL)
	if err != nil {
		// API failed - try cache fallback if available
		if c.cache != nil {
			log.Printf("[17Lands] API unavailable, attempting cache fallback: %v", err)
			cached, cachedAt, cacheErr := c.cache.GetColorRatings(ctx, params.Expansion, params.EventType)
			if cacheErr == nil && len(cached) > 0 {
				age := time.Since(cachedAt)
				log.Printf("[17Lands] Using cached color ratings (age: %v, count: %d)", age, len(cached))

				// Update stats
				c.updateStats(func(s *ClientStats) {
					s.CachedResponses++
				})

				return cached, nil
			}
			log.Printf("[17Lands] Cache miss or empty: %v", cacheErr)
		}

		// No cache available or cache failed
		return nil, &APIError{
			Type:    ErrStatsUnavailable,
			Message: "17Lands API unavailable and no cached data available",
			Err:     err,
		}
	}

	// Parse response
	var ratings []ColorRating
	if err := json.Unmarshal(body, &ratings); err != nil {
		return nil, &APIError{
			Type:    ErrParseError,
			Message: "failed to parse color ratings response",
			Err:     err,
		}
	}

	// Cache successful response if cache is available
	if c.cache != nil && len(ratings) > 0 {
		if err := c.cache.SaveColorRatings(ctx, ratings, params.Expansion, params.EventType, params.StartDate, params.EndDate); err != nil {
			log.Printf("[17Lands] Failed to cache color ratings: %v", err)
			// Don't fail the request if caching fails
		}
	}

	return ratings, nil
}

// doRequest performs an HTTP request with rate limiting and backoff.
func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
	// Check backoff
	if err := c.checkBackoff(); err != nil {
		return nil, err
	}

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, &APIError{
			Type:    ErrRateLimited,
			Message: "rate limiter error",
			Err:     err,
		}
	}

	// Update stats
	c.updateStats(func(s *ClientStats) {
		s.TotalRequests++
		s.LastRequestTime = time.Now()
	})

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &APIError{
			Type:    ErrInvalidParams,
			Message: "failed to create request",
			Err:     err,
		}
	}

	// Set user agent
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		c.recordFailure()
		return nil, &APIError{
			Type:    ErrUnavailable,
			Message: "failed to execute request",
			Err:     err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		c.recordFailure()

		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			Type:       ErrUnavailable,
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code: %d, body: %s", resp.StatusCode, string(body)),
		}
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.recordFailure()
		return nil, &APIError{
			Type:    ErrUnavailable,
			Message: "failed to read response body",
			Err:     err,
		}
	}

	// Record success
	c.recordSuccess(latency)

	return body, nil
}

// checkBackoff checks if we're in a backoff period.
func (c *Client) checkBackoff() error {
	c.backoffMu.Lock()
	defer c.backoffMu.Unlock()

	if !c.lastFailureTime.IsZero() {
		elapsed := time.Since(c.lastFailureTime)
		if elapsed < c.backoff {
			return &APIError{
				Type:    ErrRateLimited,
				Message: fmt.Sprintf("in backoff period, %v remaining", c.backoff-elapsed),
			}
		}
	}

	return nil
}

// recordFailure records a failed request and increases backoff.
func (c *Client) recordFailure() {
	c.backoffMu.Lock()
	c.lastFailureTime = time.Now()
	c.backoff = time.Duration(float64(c.backoff) * BackoffFactor)
	if c.backoff > MaxBackoff {
		c.backoff = MaxBackoff
	}
	c.backoffMu.Unlock()

	c.updateStats(func(s *ClientStats) {
		s.FailedRequests++
		s.LastFailureTime = time.Now()
		s.ConsecutiveErrors++
	})
}

// recordSuccess records a successful request and resets backoff.
func (c *Client) recordSuccess(latency time.Duration) {
	c.backoffMu.Lock()
	c.backoff = InitialBackoff
	c.lastFailureTime = time.Time{} // Reset
	c.backoffMu.Unlock()

	c.updateStats(func(s *ClientStats) {
		s.LastSuccessTime = time.Now()
		s.ConsecutiveErrors = 0

		// Update average latency
		if s.AverageLatency == 0 {
			s.AverageLatency = latency
		} else {
			s.AverageLatency = (s.AverageLatency + latency) / 2
		}
	})
}

// updateStats safely updates client statistics.
func (c *Client) updateStats(fn func(*ClientStats)) {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	fn(c.stats)
}

// GetStats returns a copy of the current client statistics.
func (c *Client) GetStats() ClientStats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	return *c.stats
}

// ResetBackoff manually resets the backoff timer.
func (c *Client) ResetBackoff() {
	c.backoffMu.Lock()
	defer c.backoffMu.Unlock()
	c.backoff = InitialBackoff
	c.lastFailureTime = time.Time{}
}

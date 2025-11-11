package seventeenlands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// APIBase is the base URL for 17Lands API
	APIBase = "https://www.17lands.com"

	// Request timeout
	DefaultTimeout = 30 * time.Second

	// Backoff settings
	InitialBackoff = 2 * time.Second
	MaxBackoff     = 60 * time.Second
	BackoffFactor  = 2.0
)

// Conservative rate limit: 1 request per second
var DefaultRateLimit = rate.Every(1 * time.Second)

// Client provides access to 17Lands draft statistics.
type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
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
		stats:      &ClientStats{},
		backoff:    InitialBackoff,
	}
}

// GetCardRatings fetches card performance statistics for a set.
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
	if len(params.Colors) > 0 {
		queryParams.Set("colors", strings.Join(params.Colors, ""))
	}

	fullURL := endpoint + "?" + queryParams.Encode()

	// Make request
	body, err := c.doRequest(ctx, fullURL)
	if err != nil {
		return nil, err
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

	return ratings, nil
}

// GetColorRatings fetches color combination performance statistics.
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

	// Make request
	body, err := c.doRequest(ctx, fullURL)
	if err != nil {
		return nil, err
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

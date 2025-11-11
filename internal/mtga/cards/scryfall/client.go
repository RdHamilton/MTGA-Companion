package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL        = "https://api.scryfall.com"
	rateLimitDelay = 100 * time.Millisecond // 100ms between requests (10 req/sec)
	requestTimeout = 30 * time.Second
	maxRetries     = 3
	initialBackoff = 1 * time.Second
	maxBackoff     = 16 * time.Second
)

// Client represents a Scryfall API client with rate limiting.
type Client struct {
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	userAgent   string
}

// NewClient creates a new Scryfall API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		// Rate limiter: 1 request per 100ms = 10 req/sec
		rateLimiter: rate.NewLimiter(rate.Every(rateLimitDelay), 1),
		userAgent:   "MTGA-Companion/1.0",
	}
}

// GetCard retrieves a card by its Scryfall ID.
func (c *Client) GetCard(ctx context.Context, id string) (*Card, error) {
	url := fmt.Sprintf("%s/cards/%s", baseURL, id)

	var card Card
	if err := c.doRequest(ctx, url, &card); err != nil {
		return nil, fmt.Errorf("failed to get card %s: %w", id, err)
	}

	return &card, nil
}

// GetCardByArenaID retrieves a card by its MTGA Arena ID.
func (c *Client) GetCardByArenaID(ctx context.Context, arenaID int) (*Card, error) {
	url := fmt.Sprintf("%s/cards/arena/%d", baseURL, arenaID)

	var card Card
	if err := c.doRequest(ctx, url, &card); err != nil {
		return nil, fmt.Errorf("failed to get card by arena ID %d: %w", arenaID, err)
	}

	return &card, nil
}

// GetSet retrieves set information by set code.
func (c *Client) GetSet(ctx context.Context, code string) (*Set, error) {
	url := fmt.Sprintf("%s/sets/%s", baseURL, code)

	var set Set
	if err := c.doRequest(ctx, url, &set); err != nil {
		return nil, fmt.Errorf("failed to get set %s: %w", code, err)
	}

	return &set, nil
}

// SearchCards performs a full-text search for cards.
func (c *Client) SearchCards(ctx context.Context, query string) (*SearchResult, error) {
	url := fmt.Sprintf("%s/cards/search?q=%s", baseURL, query)

	var result SearchResult
	if err := c.doRequest(ctx, url, &result); err != nil {
		return nil, fmt.Errorf("failed to search cards with query '%s': %w", query, err)
	}

	return &result, nil
}

// GetBulkData retrieves bulk data download information.
func (c *Client) GetBulkData(ctx context.Context) (*BulkDataList, error) {
	url := fmt.Sprintf("%s/bulk-data", baseURL)

	var bulkData BulkDataList
	if err := c.doRequest(ctx, url, &bulkData); err != nil {
		return nil, fmt.Errorf("failed to get bulk data: %w", err)
	}

	return &bulkData, nil
}

// GetSets retrieves a list of all sets.
func (c *Client) GetSets(ctx context.Context) (*SetList, error) {
	url := fmt.Sprintf("%s/sets", baseURL)

	var sets SetList
	if err := c.doRequest(ctx, url, &sets); err != nil {
		return nil, fmt.Errorf("failed to get sets: %w", err)
	}

	return &sets, nil
}

// GetMigrations retrieves a list of card migrations.
// Migrations represent cards that have been merged or deleted by Scryfall.
func (c *Client) GetMigrations(ctx context.Context) (*MigrationList, error) {
	url := fmt.Sprintf("%s/migrations", baseURL)

	var migrations MigrationList
	if err := c.doRequest(ctx, url, &migrations); err != nil {
		return nil, fmt.Errorf("failed to get migrations: %w", err)
	}

	return &migrations, nil
}

// doRequest performs an HTTP request with rate limiting and retry logic.
func (c *Client) doRequest(ctx context.Context, url string, result interface{}) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)

			// Retry on network errors
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff = min(backoff*2, maxBackoff)
				continue
			}
			return lastErr
		}

		// Handle response
		defer func() { _ = resp.Body.Close() }()

		// Check status code
		switch resp.StatusCode {
		case http.StatusOK:
			// Success - parse response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("failed to parse JSON response: %w", err)
			}

			return nil

		case http.StatusTooManyRequests:
			// Rate limited - exponential backoff
			lastErr = fmt.Errorf("rate limited (HTTP 429)")

			if attempt < maxRetries {
				// Check for Retry-After header
				retryAfter := resp.Header.Get("Retry-After")
				if retryAfter != "" {
					// If Retry-After is provided, use it
					if duration, err := time.ParseDuration(retryAfter + "s"); err == nil {
						time.Sleep(duration)
					} else {
						time.Sleep(backoff)
					}
				} else {
					time.Sleep(backoff)
				}
				backoff = min(backoff*2, maxBackoff)
				continue
			}
			return lastErr

		case http.StatusNotFound:
			return &NotFoundError{URL: url}

		default:
			// Try to parse error response
			body, _ := io.ReadAll(resp.Body)

			var apiErr APIError
			if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Details != "" {
				return &apiErr
			}

			return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// min returns the minimum of two time.Duration values.
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

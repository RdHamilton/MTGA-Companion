package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClientConfig holds configuration for the daemon HTTP client.
type ClientConfig struct {
	// BaseURL is the base URL of the daemon API (e.g., "http://localhost:9999")
	BaseURL string

	// Timeout is the timeout for individual requests
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryBaseDelay is the base delay for exponential backoff
	RetryBaseDelay time.Duration
}

// DefaultClientConfig returns a ClientConfig with sensible defaults.
func DefaultClientConfig(port int) *ClientConfig {
	return &ClientConfig{
		BaseURL:        fmt.Sprintf("http://localhost:%d", port),
		Timeout:        10 * time.Second,
		MaxRetries:     3,
		RetryBaseDelay: 500 * time.Millisecond,
	}
}

// Client is an HTTP client for communicating with mtga-tracker-daemon.
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
}

// NewClient creates a new daemon HTTP client.
func NewClient(config *ClientConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Status represents the daemon's status response.
type Status struct {
	Status        string `json:"status"`        // "connected", "disconnected", "error"
	Connected     bool   `json:"connected"`     // true if connected to MTGA
	Version       string `json:"version"`       // daemon version
	MTGAConnected bool   `json:"mtgaConnected"` // true if MTGA is running and connected
	PlayerID      string `json:"playerId"`      // current player's Arena ID
	LastUpdate    string `json:"lastUpdate"`    // ISO timestamp of last data update
}

// CardCollection represents the player's full card collection.
type CardCollection struct {
	Cards map[int]int `json:"cards"` // arena_id -> count
}

// Inventory represents the player's currency and wildcard inventory.
type Inventory struct {
	Gold          int     `json:"gold"`
	Gems          int     `json:"gems"`
	CommonWC      int     `json:"wcCommon"`
	UncommonWC    int     `json:"wcUncommon"`
	RareWC        int     `json:"wcRare"`
	MythicWC      int     `json:"wcMythic"`
	VaultProgress float64 `json:"vaultProgress"`
	DraftTokens   int     `json:"draftTokens"`
	SealedTokens  int     `json:"sealedTokens"`
}

// PlayerInfo represents the current player's information.
type PlayerInfo struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// GetStatus retrieves the daemon's current status.
func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
	var status Status
	err := c.doRequest(ctx, "GET", "/status", &status)
	if err != nil {
		return nil, fmt.Errorf("failed to get daemon status: %w", err)
	}
	return &status, nil
}

// GetCards retrieves the player's full card collection.
func (c *Client) GetCards(ctx context.Context) (*CardCollection, error) {
	var collection CardCollection
	err := c.doRequest(ctx, "GET", "/cards", &collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get card collection: %w", err)
	}
	return &collection, nil
}

// GetInventory retrieves the player's currency and wildcard inventory.
func (c *Client) GetInventory(ctx context.Context) (*Inventory, error) {
	var inventory Inventory
	err := c.doRequest(ctx, "GET", "/inventory", &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return &inventory, nil
}

// GetPlayerID retrieves the current player's Arena ID.
func (c *Client) GetPlayerID(ctx context.Context) (string, error) {
	var playerInfo PlayerInfo
	err := c.doRequest(ctx, "GET", "/playerId", &playerInfo)
	if err != nil {
		return "", fmt.Errorf("failed to get player ID: %w", err)
	}
	return playerInfo.PlayerID, nil
}

// IsHealthy checks if the daemon is healthy and responding.
func (c *Client) IsHealthy(ctx context.Context) bool {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return false
	}
	return status.Status == "connected" || status.Status == "healthy"
}

// doRequest performs an HTTP request with retry logic.
func (c *Client) doRequest(ctx context.Context, method, path string, result interface{}) error {
	url := c.config.BaseURL + path

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := c.config.RetryBaseDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		defer func() {
			//nolint:errcheck // Ignore error on cleanup
			_ = resp.Body.Close()
		}()

		// Check for server errors (5xx) - these are retryable
		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
			continue
		}

		// Non-retryable errors
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
		}

		// Parse response
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	}

	return lastErr
}

// SetBaseURL updates the base URL for the client.
func (c *Client) SetBaseURL(url string) {
	c.config.BaseURL = url
}

// GetBaseURL returns the current base URL.
func (c *Client) GetBaseURL() string {
	return c.config.BaseURL
}

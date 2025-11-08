package cards

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// ScryfallAPIBase is the base URL for Scryfall API requests.
	ScryfallAPIBase = "https://api.scryfall.com"

	// ScryfallBulkDataURL is the URL for bulk data downloads.
	ScryfallBulkDataURL = ScryfallAPIBase + "/bulk-data"

	// ScryfallRateLimit is the recommended delay between API requests (50-100ms).
	ScryfallRateLimit = 100 * time.Millisecond
)

// ScryfallClient handles requests to the Scryfall API.
type ScryfallClient struct {
	httpClient  *http.Client
	lastRequest time.Time
	rateLimit   time.Duration
}

// NewScryfallClient creates a new Scryfall API client.
func NewScryfallClient() *ScryfallClient {
	return &ScryfallClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: ScryfallRateLimit,
	}
}

// GetCardByArenaID fetches a card by its Arena ID from Scryfall.
func (sc *ScryfallClient) GetCardByArenaID(arenaID int) (*Card, error) {
	// Rate limiting
	sc.waitForRateLimit()

	url := fmt.Sprintf("%s/cards/arena/%d", ScryfallAPIBase, arenaID)
	resp, err := sc.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch card from Scryfall: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Ignore error on cleanup

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("card with Arena ID %d not found", arenaID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Scryfall API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var scryfallCard ScryfallCard
	if err := json.Unmarshal(body, &scryfallCard); err != nil {
		return nil, fmt.Errorf("failed to parse Scryfall response: %w", err)
	}

	return scryfallCard.ToCard(), nil
}

// GetCardByName fetches a card by its exact name from Scryfall.
func (sc *ScryfallClient) GetCardByName(name string) (*Card, error) {
	// Rate limiting
	sc.waitForRateLimit()

	url := fmt.Sprintf("%s/cards/named?exact=%s", ScryfallAPIBase, name)
	resp, err := sc.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch card from Scryfall: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Ignore error on cleanup

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("card named %q not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Scryfall API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var scryfallCard ScryfallCard
	if err := json.Unmarshal(body, &scryfallCard); err != nil {
		return nil, fmt.Errorf("failed to parse Scryfall response: %w", err)
	}

	return scryfallCard.ToCard(), nil
}

// BulkDataInfo represents information about a Scryfall bulk data file.
type BulkDataInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DownloadURI string    `json:"download_uri"`
	UpdatedAt   time.Time `json:"updated_at"`
	Size        int64     `json:"size"`
}

// BulkDataList represents the list of available bulk data files.
type BulkDataList struct {
	Data []BulkDataInfo `json:"data"`
}

// GetBulkDataInfo fetches information about available bulk data downloads.
func (sc *ScryfallClient) GetBulkDataInfo() ([]BulkDataInfo, error) {
	// Rate limiting
	sc.waitForRateLimit()

	resp, err := sc.httpClient.Get(ScryfallBulkDataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bulk data info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Ignore error on cleanup

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Scryfall API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var bulkDataList BulkDataList
	if err := json.Unmarshal(body, &bulkDataList); err != nil {
		return nil, fmt.Errorf("failed to parse bulk data response: %w", err)
	}

	return bulkDataList.Data, nil
}

// DownloadBulkData downloads a bulk data file and returns the cards.
func (sc *ScryfallClient) DownloadBulkData(downloadURL string) ([]*ScryfallCard, error) {
	resp, err := sc.httpClient.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download bulk data: %w", err)
	}
	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Ignore error on cleanup

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bulk data download returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read bulk data: %w", err)
	}

	var cards []*ScryfallCard
	if err := json.Unmarshal(body, &cards); err != nil {
		return nil, fmt.Errorf("failed to parse bulk data: %w", err)
	}

	return cards, nil
}

// waitForRateLimit implements rate limiting for Scryfall API requests.
func (sc *ScryfallClient) waitForRateLimit() {
	if !sc.lastRequest.IsZero() {
		elapsed := time.Since(sc.lastRequest)
		if elapsed < sc.rateLimit {
			time.Sleep(sc.rateLimit - elapsed)
		}
	}
	sc.lastRequest = time.Now()
}

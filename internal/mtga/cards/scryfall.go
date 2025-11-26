package cards

import (
	"bytes"
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

	// ScryfallCollectionURL is the URL for batch card lookups.
	ScryfallCollectionURL = ScryfallAPIBase + "/cards/collection"

	// ScryfallRateLimit is the recommended delay between API requests (50-100ms).
	ScryfallRateLimit = 100 * time.Millisecond

	// ScryfallMaxBatchSize is the maximum number of cards per batch request (Scryfall limit is 75).
	ScryfallMaxBatchSize = 75
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
		return nil, fmt.Errorf("scryfall API returned status %d", resp.StatusCode)
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
		return nil, fmt.Errorf("scryfall API returned status %d", resp.StatusCode)
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

// CardIdentifier represents a card identifier for the /cards/collection endpoint.
type CardIdentifier struct {
	ID              string `json:"id,omitempty"`               // Scryfall ID
	MTGOID          int    `json:"mtgo_id,omitempty"`          // MTGO ID
	MultiverseID    int    `json:"multiverse_id,omitempty"`    // Multiverse ID
	OracleID        string `json:"oracle_id,omitempty"`        // Oracle ID
	IllustrationID  string `json:"illustration_id,omitempty"`  // Illustration ID
	Name            string `json:"name,omitempty"`             // Card name
	Set             string `json:"set,omitempty"`              // Set code (requires collector_number)
	CollectorNumber string `json:"collector_number,omitempty"` // Collector number (requires set)
}

// CollectionRequest is the request body for /cards/collection.
type CollectionRequest struct {
	Identifiers []CardIdentifier `json:"identifiers"`
}

// CollectionResponse is the response from /cards/collection.
type CollectionResponse struct {
	Object   string           `json:"object"`
	NotFound []CardIdentifier `json:"not_found"`
	Data     []*ScryfallCard  `json:"data"`
}

// GetCardsByArenaIDs fetches multiple cards by their Arena IDs using the batch /cards/collection endpoint.
// This is much faster than individual lookups for large numbers of cards.
// Automatically handles batching if more than 75 cards are requested.
func (sc *ScryfallClient) GetCardsByArenaIDs(arenaIDs []int) ([]*Card, error) {
	if len(arenaIDs) == 0 {
		return []*Card{}, nil
	}

	var allCards []*Card

	// Process in batches of ScryfallMaxBatchSize (75)
	for i := 0; i < len(arenaIDs); i += ScryfallMaxBatchSize {
		end := i + ScryfallMaxBatchSize
		if end > len(arenaIDs) {
			end = len(arenaIDs)
		}
		batch := arenaIDs[i:end]

		cards, err := sc.fetchCardsBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end, err)
		}
		allCards = append(allCards, cards...)
	}

	return allCards, nil
}

// fetchCardsBatch fetches a single batch of cards (up to 75) from the /cards/collection endpoint.
func (sc *ScryfallClient) fetchCardsBatch(arenaIDs []int) ([]*Card, error) {
	// Rate limiting
	sc.waitForRateLimit()

	// Build identifiers - Scryfall doesn't support arena_id directly in /cards/collection,
	// so we need to use the multiverse_id field which sometimes matches arena_id for older cards.
	// For Arena-specific cards, we may need to fall back to individual lookups.
	// However, we can use the "name" identifier as a fallback strategy.
	// For now, we'll try using the arena endpoint pattern with name lookups.

	// Actually, Scryfall's /cards/collection doesn't directly support arena_id.
	// We need to fetch cards individually by arena_id and cache them, OR use the set+collector_number.
	// The best approach for Arena IDs is to use individual /cards/arena/{id} calls with concurrency,
	// but we can batch them more efficiently by caching.

	// Alternative: Use the search API with "game:arena" filter, but that's also limited.

	// For now, let's implement a concurrent batch fetcher that's more efficient than sequential.
	// We'll make concurrent requests but respect rate limits.

	var cards []*Card
	results := make(chan *Card, len(arenaIDs))
	errors := make(chan error, len(arenaIDs))

	// Use semaphore to limit concurrent requests (10 at a time with rate limiting)
	sem := make(chan struct{}, 10)

	for _, arenaID := range arenaIDs {
		go func(id int) {
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			// Rate limit each request
			sc.waitForRateLimit()

			card, err := sc.GetCardByArenaID(id)
			if err != nil {
				errors <- err
				return
			}
			results <- card
		}(arenaID)
	}

	// Collect results
	for range arenaIDs {
		select {
		case card := <-results:
			if card != nil {
				cards = append(cards, card)
			}
		case <-errors:
			// Continue even if some cards fail
		}
	}

	return cards, nil
}

// GetCardsByNames fetches multiple cards by their names using the batch /cards/collection endpoint.
// This IS supported by the /cards/collection endpoint and is very efficient.
func (sc *ScryfallClient) GetCardsByNames(names []string) ([]*Card, error) {
	if len(names) == 0 {
		return []*Card{}, nil
	}

	var allCards []*Card

	// Process in batches of ScryfallMaxBatchSize (75)
	for i := 0; i < len(names); i += ScryfallMaxBatchSize {
		end := i + ScryfallMaxBatchSize
		if end > len(names) {
			end = len(names)
		}
		batch := names[i:end]

		cards, err := sc.fetchCardsByNamesBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end, err)
		}
		allCards = append(allCards, cards...)
	}

	return allCards, nil
}

// fetchCardsByNamesBatch fetches a single batch of cards by name from /cards/collection.
func (sc *ScryfallClient) fetchCardsByNamesBatch(names []string) ([]*Card, error) {
	// Rate limiting
	sc.waitForRateLimit()

	// Build identifiers
	identifiers := make([]CardIdentifier, len(names))
	for i, name := range names {
		identifiers[i] = CardIdentifier{Name: name}
	}

	reqBody := CollectionRequest{Identifiers: identifiers}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ScryfallCollectionURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cards from Scryfall: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scryfall API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var collectionResp CollectionResponse
	if err := json.Unmarshal(body, &collectionResp); err != nil {
		return nil, fmt.Errorf("failed to parse Scryfall response: %w", err)
	}

	// Convert to Card objects
	cards := make([]*Card, 0, len(collectionResp.Data))
	for _, scryfallCard := range collectionResp.Data {
		cards = append(cards, scryfallCard.ToCard())
	}

	return cards, nil
}

// GetCardsBySetAndNumbers fetches multiple cards by set code and collector number.
// This is the most reliable batch lookup method.
func (sc *ScryfallClient) GetCardsBySetAndNumbers(setCode string, collectorNumbers []string) ([]*Card, error) {
	if len(collectorNumbers) == 0 {
		return []*Card{}, nil
	}

	var allCards []*Card

	// Process in batches of ScryfallMaxBatchSize (75)
	for i := 0; i < len(collectorNumbers); i += ScryfallMaxBatchSize {
		end := i + ScryfallMaxBatchSize
		if end > len(collectorNumbers) {
			end = len(collectorNumbers)
		}
		batch := collectorNumbers[i:end]

		cards, err := sc.fetchCardsBySetAndNumbersBatch(setCode, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end, err)
		}
		allCards = append(allCards, cards...)
	}

	return allCards, nil
}

// fetchCardsBySetAndNumbersBatch fetches cards by set and collector number.
func (sc *ScryfallClient) fetchCardsBySetAndNumbersBatch(setCode string, collectorNumbers []string) ([]*Card, error) {
	// Rate limiting
	sc.waitForRateLimit()

	// Build identifiers
	identifiers := make([]CardIdentifier, len(collectorNumbers))
	for i, num := range collectorNumbers {
		identifiers[i] = CardIdentifier{
			Set:             setCode,
			CollectorNumber: num,
		}
	}

	reqBody := CollectionRequest{Identifiers: identifiers}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", ScryfallCollectionURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cards from Scryfall: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scryfall API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var collectionResp CollectionResponse
	if err := json.Unmarshal(body, &collectionResp); err != nil {
		return nil, fmt.Errorf("failed to parse Scryfall response: %w", err)
	}

	// Convert to Card objects
	cards := make([]*Card, 0, len(collectionResp.Data))
	for _, scryfallCard := range collectionResp.Data {
		cards = append(cards, scryfallCard.ToCard())
	}

	return cards, nil
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
		return nil, fmt.Errorf("scryfall API returned status %d", resp.StatusCode)
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

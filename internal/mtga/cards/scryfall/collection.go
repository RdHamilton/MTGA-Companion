package scryfall

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// CollectionURL is the URL for batch card lookups.
	collectionURL = baseURL + "/cards/collection"

	// MaxBatchSize is the maximum number of cards per batch request (Scryfall limit is 75).
	MaxBatchSize = 75
)

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
	Data     []Card           `json:"data"`
}

// GetCardsByNames fetches multiple cards by their names using the batch /cards/collection endpoint.
// This is much more efficient than individual lookups for large numbers of cards.
// Automatically handles batching if more than 75 cards are requested.
func (c *Client) GetCardsByNames(ctx context.Context, names []string) ([]Card, []string, error) {
	if len(names) == 0 {
		return []Card{}, nil, nil
	}

	var allCards []Card
	var allNotFound []string

	// Process in batches of MaxBatchSize (75)
	for i := 0; i < len(names); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(names) {
			end = len(names)
		}
		batch := names[i:end]

		cards, notFound, err := c.fetchCardsByNamesBatch(ctx, batch)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end, err)
		}
		allCards = append(allCards, cards...)
		allNotFound = append(allNotFound, notFound...)
	}

	return allCards, allNotFound, nil
}

// fetchCardsByNamesBatch fetches a single batch of cards by name from /cards/collection.
func (c *Client) fetchCardsByNamesBatch(ctx context.Context, names []string) ([]Card, []string, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build identifiers
	identifiers := make([]CardIdentifier, len(names))
	for i, name := range names {
		identifiers[i] = CardIdentifier{Name: name}
	}

	return c.doCollectionRequest(ctx, identifiers)
}

// GetCardsBySetAndNumbers fetches multiple cards by set code and collector number.
// This is the most reliable batch lookup method.
// Automatically handles batching if more than 75 cards are requested.
func (c *Client) GetCardsBySetAndNumbers(ctx context.Context, setCode string, collectorNumbers []string) ([]Card, []string, error) {
	if len(collectorNumbers) == 0 {
		return []Card{}, nil, nil
	}

	var allCards []Card
	var allNotFound []string

	// Process in batches of MaxBatchSize (75)
	for i := 0; i < len(collectorNumbers); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(collectorNumbers) {
			end = len(collectorNumbers)
		}
		batch := collectorNumbers[i:end]

		cards, notFound, err := c.fetchCardsBySetAndNumbersBatch(ctx, setCode, batch)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end, err)
		}
		allCards = append(allCards, cards...)
		allNotFound = append(allNotFound, notFound...)
	}

	return allCards, allNotFound, nil
}

// fetchCardsBySetAndNumbersBatch fetches cards by set and collector number.
func (c *Client) fetchCardsBySetAndNumbersBatch(ctx context.Context, setCode string, collectorNumbers []string) ([]Card, []string, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build identifiers
	identifiers := make([]CardIdentifier, len(collectorNumbers))
	for i, num := range collectorNumbers {
		identifiers[i] = CardIdentifier{
			Set:             setCode,
			CollectorNumber: num,
		}
	}

	return c.doCollectionRequest(ctx, identifiers)
}

// GetCardsByMixedIdentifiers fetches cards using a mixed set of identifiers.
// Each identifier can use different lookup methods (name, set+number, scryfall_id, etc).
// Automatically handles batching if more than 75 cards are requested.
func (c *Client) GetCardsByMixedIdentifiers(ctx context.Context, identifiers []CardIdentifier) ([]Card, []CardIdentifier, error) {
	if len(identifiers) == 0 {
		return []Card{}, nil, nil
	}

	var allCards []Card
	var allNotFound []CardIdentifier

	// Process in batches of MaxBatchSize (75)
	for i := 0; i < len(identifiers); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(identifiers) {
			end = len(identifiers)
		}
		batch := identifiers[i:end]

		cards, notFoundIDs, err := c.doCollectionRequestWithIdentifiers(ctx, batch)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end, err)
		}
		allCards = append(allCards, cards...)
		allNotFound = append(allNotFound, notFoundIDs...)
	}

	return allCards, allNotFound, nil
}

// doCollectionRequest performs a batch request to /cards/collection.
func (c *Client) doCollectionRequest(ctx context.Context, identifiers []CardIdentifier) ([]Card, []string, error) {
	cards, notFoundIDs, err := c.doCollectionRequestWithIdentifiers(ctx, identifiers)
	if err != nil {
		return nil, nil, err
	}

	// Convert CardIdentifiers to strings (names) for the notFound list
	notFound := make([]string, 0, len(notFoundIDs))
	for _, id := range notFoundIDs {
		if id.Name != "" {
			notFound = append(notFound, id.Name)
		} else if id.Set != "" && id.CollectorNumber != "" {
			notFound = append(notFound, fmt.Sprintf("%s#%s", id.Set, id.CollectorNumber))
		}
	}

	return cards, notFound, nil
}

// doCollectionRequestWithIdentifiers performs a batch request and returns full identifier info for not-found cards.
func (c *Client) doCollectionRequestWithIdentifiers(ctx context.Context, identifiers []CardIdentifier) ([]Card, []CardIdentifier, error) {
	reqBody := CollectionRequest{Identifiers: identifiers}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, collectionURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch cards from Scryfall: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("scryfall API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var collectionResp CollectionResponse
	if err := json.Unmarshal(body, &collectionResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse Scryfall response: %w", err)
	}

	return collectionResp.Data, collectionResp.NotFound, nil
}

// Package synergy provides card synergy analysis using external deck data sources.
package synergy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ArchidektClient fetches deck data from Archidekt for co-occurrence analysis.
type ArchidektClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewArchidektClient creates a new Archidekt API client.
func NewArchidektClient() *ArchidektClient {
	return &ArchidektClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://archidekt.com/api",
	}
}

// ArchidektDeck represents a deck from Archidekt.
type ArchidektDeck struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
	DeckFormat  int                  `json:"deckFormat"`
	Description string               `json:"description"`
	ViewCount   int                  `json:"viewCount"`
	Featured    string               `json:"featured"`
	Private     bool                 `json:"private"`
	Owner       *ArchidektOwner      `json:"owner"`
	Categories  []*ArchidektCategory `json:"categories"`
	Cards       []*ArchidektDeckCard `json:"cards"`
}

// ArchidektOwner represents the deck owner.
type ArchidektOwner struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// ArchidektCategory represents a deck category (e.g., "Commander", "Lands").
type ArchidektCategory struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	IsPremier      bool   `json:"isPremier"`
	IncludedInDeck bool   `json:"includedInDeck"`
}

// ArchidektDeckCard represents a card in an Archidekt deck.
type ArchidektDeckCard struct {
	ID         int            `json:"id"`
	Categories []string       `json:"categories"`
	Quantity   int            `json:"quantity"`
	Card       *ArchidektCard `json:"card"`
}

// ArchidektCard represents card metadata from Archidekt.
type ArchidektCard struct {
	ID              int                  `json:"id"`
	Artist          string               `json:"artist"`
	CollectorNumber string               `json:"collectorNumber"`
	Edition         *ArchidektEdition    `json:"edition"`
	OracleCard      *ArchidektOracleCard `json:"oracleCard"`
}

// ArchidektEdition represents card edition/set information.
type ArchidektEdition struct {
	EditionCode string `json:"editioncode"`
	EditionName string `json:"editionname"`
	EditionDate string `json:"editiondate"`
	EditionType string `json:"editiontype"`
}

// ArchidektOracleCard represents the oracle (canonical) card data.
type ArchidektOracleCard struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	CMC           float64  `json:"cmc"`
	ColorIdentity []string `json:"colorIdentity"`
	Colors        []string `json:"colors"`
	ManaCost      string   `json:"manaCost"`
	Text          string   `json:"text"`
	TypeLine      string   `json:"type"`
	EDHRecRank    int      `json:"edhrecRank"`
}

// ArchidektFormat maps format IDs to format names.
var ArchidektFormat = map[int]string{
	1:  "Standard",
	2:  "Modern",
	3:  "Commander",
	4:  "Legacy",
	5:  "Vintage",
	6:  "Pauper",
	7:  "Brawl",
	8:  "Pioneer",
	9:  "Historic",
	10: "Penny Dreadful",
	11: "Oathbreaker",
	12: "Explorer",
}

// FormatNameToID converts a format name to Archidekt's format ID.
func FormatNameToID(format string) int {
	for id, name := range ArchidektFormat {
		if name == format {
			return id
		}
	}
	return 0
}

// GetDeck fetches a single deck by ID.
func (c *ArchidektClient) GetDeck(ctx context.Context, deckID int) (*ArchidektDeck, error) {
	url := fmt.Sprintf("%s/decks/%d/", c.baseURL, deckID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deck: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("deck not found: %d", deckID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var deck ArchidektDeck
	if err := json.NewDecoder(resp.Body).Decode(&deck); err != nil {
		return nil, fmt.Errorf("failed to decode deck: %w", err)
	}

	return &deck, nil
}

// GetDecksBatch fetches multiple decks by ID with rate limiting.
// Returns a map of deck ID to deck data, skipping any decks that fail to fetch.
func (c *ArchidektClient) GetDecksBatch(ctx context.Context, deckIDs []int) (map[int]*ArchidektDeck, error) {
	decks := make(map[int]*ArchidektDeck)

	for _, id := range deckIDs {
		select {
		case <-ctx.Done():
			return decks, ctx.Err()
		default:
		}

		deck, err := c.GetDeck(ctx, id)
		if err != nil {
			// Log but continue - we don't want one bad deck to fail the whole batch
			continue
		}

		decks[id] = deck

		// Rate limit: 100ms between requests to be respectful of the API
		time.Sleep(100 * time.Millisecond)
	}

	return decks, nil
}

// ExtractCardIDs extracts unique card ArenaIDs from an Archidekt deck.
// Note: Archidekt uses internal IDs, not Arena IDs. This function extracts
// the oracle card names which can be matched against our local card database.
func (d *ArchidektDeck) ExtractCardNames() []string {
	names := make([]string, 0, len(d.Cards))
	seen := make(map[string]bool)

	for _, deckCard := range d.Cards {
		if deckCard.Card == nil || deckCard.Card.OracleCard == nil {
			continue
		}

		name := deckCard.Card.OracleCard.Name
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}

	return names
}

// ExtractCardPairs returns all unique pairs of cards in the deck for co-occurrence.
// Each pair is represented as (nameA, nameB) where nameA < nameB (alphabetically).
func (d *ArchidektDeck) ExtractCardPairs() [][2]string {
	names := d.ExtractCardNames()
	if len(names) < 2 {
		return nil
	}

	// Calculate number of pairs: n*(n-1)/2
	numPairs := len(names) * (len(names) - 1) / 2
	pairs := make([][2]string, 0, numPairs)

	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			// Ensure consistent ordering (alphabetically)
			if names[i] < names[j] {
				pairs = append(pairs, [2]string{names[i], names[j]})
			} else {
				pairs = append(pairs, [2]string{names[j], names[i]})
			}
		}
	}

	return pairs
}

// GetFormatName returns the format name for this deck.
func (d *ArchidektDeck) GetFormatName() string {
	if name, ok := ArchidektFormat[d.DeckFormat]; ok {
		return name
	}
	return "Unknown"
}

// DeckSource represents a source of deck data for co-occurrence analysis.
type DeckSource interface {
	// FetchDecks retrieves decks for analysis.
	FetchDecks(ctx context.Context, format string, limit int) ([]*SimpleDeck, error)
	// SourceName returns the name of this source (e.g., "archidekt", "local").
	SourceName() string
}

// SimpleDeck is a normalized deck representation for co-occurrence analysis.
type SimpleDeck struct {
	ID        string   // Unique identifier
	Name      string   // Deck name
	Format    string   // Format (Standard, Historic, etc.)
	CardNames []string // List of unique card names in the deck
}

// ArchidektSource implements DeckSource using the Archidekt API.
type ArchidektSource struct {
	client  *ArchidektClient
	deckIDs []int // Known deck IDs to fetch
}

// NewArchidektSource creates a new Archidekt deck source.
// deckIDs should be a curated list of known public deck IDs to analyze.
func NewArchidektSource(deckIDs []int) *ArchidektSource {
	return &ArchidektSource{
		client:  NewArchidektClient(),
		deckIDs: deckIDs,
	}
}

// FetchDecks fetches decks from Archidekt.
func (s *ArchidektSource) FetchDecks(ctx context.Context, format string, limit int) ([]*SimpleDeck, error) {
	formatID := FormatNameToID(format)

	// Filter deck IDs if we only have a subset
	idsToFetch := s.deckIDs
	if limit > 0 && limit < len(idsToFetch) {
		idsToFetch = idsToFetch[:limit]
	}

	decks, err := s.client.GetDecksBatch(ctx, idsToFetch)
	if err != nil {
		return nil, err
	}

	result := make([]*SimpleDeck, 0, len(decks))
	for id, deck := range decks {
		// Skip if format doesn't match (0 means any format)
		if formatID != 0 && deck.DeckFormat != formatID {
			continue
		}

		result = append(result, &SimpleDeck{
			ID:        fmt.Sprintf("archidekt-%d", id),
			Name:      deck.Name,
			Format:    deck.GetFormatName(),
			CardNames: deck.ExtractCardNames(),
		})
	}

	return result, nil
}

// SourceName returns the source name.
func (s *ArchidektSource) SourceName() string {
	return "archidekt"
}

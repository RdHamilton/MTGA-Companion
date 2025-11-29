package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GoldfishClient fetches meta data from MTGGoldfish.
type GoldfishClient struct {
	httpClient  *http.Client
	baseURL     string
	cache       *MetaCache
	cacheTTL    time.Duration
	rateLimiter *time.Ticker
	lastRequest time.Time
	mu          sync.Mutex
}

// GoldfishConfig configures the Goldfish client.
type GoldfishConfig struct {
	// BaseURL is the MTGGoldfish base URL.
	BaseURL string

	// CacheTTL is how long to cache meta data.
	CacheTTL time.Duration

	// RequestTimeout is the HTTP request timeout.
	RequestTimeout time.Duration

	// RateLimitMs is minimum milliseconds between requests.
	RateLimitMs int
}

// DefaultGoldfishConfig returns default configuration.
func DefaultGoldfishConfig() *GoldfishConfig {
	return &GoldfishConfig{
		BaseURL:        "https://www.mtggoldfish.com",
		CacheTTL:       4 * time.Hour,
		RequestTimeout: 30 * time.Second,
		RateLimitMs:    1000,
	}
}

// MetaDeck represents a deck in the meta.
type MetaDeck struct {
	Name           string     `json:"name"`
	ArchetypeName  string     `json:"archetype_name"`
	Format         string     `json:"format"`
	Tier           int        `json:"tier"` // 1, 2, 3, or 0 for untiered
	MetaShare      float64    `json:"meta_share"`
	WinRate        float64    `json:"win_rate,omitempty"`
	MatchCount     int        `json:"match_count,omitempty"`
	Colors         []string   `json:"colors"`
	MainboardCards []DeckCard `json:"mainboard,omitempty"`
	SideboardCards []DeckCard `json:"sideboard,omitempty"`
	URL            string     `json:"url,omitempty"`
	LastUpdated    time.Time  `json:"last_updated"`
}

// DeckCard represents a card in a deck list.
type DeckCard struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	ArenaID  int    `json:"arena_id,omitempty"`
}

// FormatMeta represents the meta for a specific format.
type FormatMeta struct {
	Format      string      `json:"format"`
	Decks       []*MetaDeck `json:"decks"`
	TotalDecks  int         `json:"total_decks"`
	LastUpdated time.Time   `json:"last_updated"`
	Source      string      `json:"source"`
}

// MetaCache caches meta data.
type MetaCache struct {
	data map[string]*CacheEntry
	mu   sync.RWMutex
}

// CacheEntry represents a cached meta entry.
type CacheEntry struct {
	Meta      *FormatMeta
	ExpiresAt time.Time
}

// NewGoldfishClient creates a new MTGGoldfish client.
func NewGoldfishClient(config *GoldfishConfig) *GoldfishClient {
	if config == nil {
		config = DefaultGoldfishConfig()
	}

	return &GoldfishClient{
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		baseURL:     config.BaseURL,
		cacheTTL:    config.CacheTTL,
		rateLimiter: time.NewTicker(time.Duration(config.RateLimitMs) * time.Millisecond),
		cache: &MetaCache{
			data: make(map[string]*CacheEntry),
		},
	}
}

// GetMeta retrieves meta data for a format.
func (c *GoldfishClient) GetMeta(ctx context.Context, format string) (*FormatMeta, error) {
	// Check cache first
	if cached := c.getFromCache(format); cached != nil {
		return cached, nil
	}

	// Rate limit
	c.waitForRateLimit()

	// Fetch from MTGGoldfish
	meta, err := c.fetchMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	// Cache result
	c.setCache(format, meta)

	return meta, nil
}

// GetTopDecks returns the top N decks for a format.
func (c *GoldfishClient) GetTopDecks(ctx context.Context, format string, limit int) ([]*MetaDeck, error) {
	meta, err := c.GetMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(meta.Decks) {
		return meta.Decks, nil
	}

	return meta.Decks[:limit], nil
}

// GetDeckByArchetype finds a deck by archetype name.
func (c *GoldfishClient) GetDeckByArchetype(ctx context.Context, format, archetype string) (*MetaDeck, error) {
	meta, err := c.GetMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	archetypeLower := strings.ToLower(archetype)
	for _, deck := range meta.Decks {
		if strings.ToLower(deck.ArchetypeName) == archetypeLower ||
			strings.ToLower(deck.Name) == archetypeLower {
			return deck, nil
		}
	}

	return nil, fmt.Errorf("archetype not found: %s", archetype)
}

// GetMetaShare returns the meta share percentage for an archetype.
func (c *GoldfishClient) GetMetaShare(ctx context.Context, format, archetype string) (float64, error) {
	deck, err := c.GetDeckByArchetype(ctx, format, archetype)
	if err != nil {
		return 0, err
	}
	return deck.MetaShare, nil
}

// fetchMeta fetches meta data from MTGGoldfish.
func (c *GoldfishClient) fetchMeta(ctx context.Context, format string) (*FormatMeta, error) {
	// Map format names to MTGGoldfish URLs
	formatURLs := map[string]string{
		"standard": "/metagame/standard/full",
		"historic": "/metagame/historic/full",
		"explorer": "/metagame/explorer/full",
		"pioneer":  "/metagame/pioneer/full",
		"modern":   "/metagame/modern/full",
		"legacy":   "/metagame/legacy/full",
		"vintage":  "/metagame/vintage/full",
		"pauper":   "/metagame/pauper/full",
		"alchemy":  "/metagame/alchemy/full",
		"timeless": "/metagame/timeless/full",
	}

	urlPath, ok := formatURLs[strings.ToLower(format)]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	url := c.baseURL + urlPath

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "MTGA-Companion/1.4.0")
	req.Header.Set("Accept", "text/html")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the HTML response
	meta := c.parseMetaPage(string(body), format)
	meta.Source = "mtggoldfish"
	meta.LastUpdated = time.Now()

	return meta, nil
}

// parseMetaPage parses the MTGGoldfish meta page HTML.
func (c *GoldfishClient) parseMetaPage(html, format string) *FormatMeta {
	meta := &FormatMeta{
		Format: format,
		Decks:  make([]*MetaDeck, 0),
	}

	// Parse archetype tiles from the page
	// MTGGoldfish structure (as of 2024):
	// <div class='archetype-tile' id='28086'>
	//   <div class='archetype-tile-title'>
	//     <a href="/archetype/...">Deck Name</a>
	//   </div>
	//   <div class='archetype-tile-statistic metagame-percentage'>
	//     <div class='archetype-tile-statistic-value'>
	//       21.3%
	//     </div>
	//   </div>
	// </div>

	// Primary pattern: matches archetype tiles with deck name in anchor and percentage in statistic-value
	// Uses ['\"] to handle both single and double quotes
	archetypePattern := regexp.MustCompile(`(?s)<div[^>]*class=['\"][^'\"]*archetype-tile[^'\"]*['\"][^>]*>.*?<div[^>]*class=['\"][^'\"]*archetype-tile-title[^'\"]*['\"][^>]*>.*?<a[^>]*>([^<]+)</a>.*?<div[^>]*class=['\"][^'\"]*archetype-tile-statistic-value[^'\"]*['\"][^>]*>\s*(\d+\.?\d*)%`)

	// Fallback: match from metagame-percentage section directly
	fallbackPattern := regexp.MustCompile(`(?s)<div[^>]*class=['\"][^'\"]*archetype-tile-title[^'\"]*['\"][^>]*>.*?<a[^>]*href=['\"][^'\"]*['\"][^>]*>([^<]+)</a>.*?<div[^>]*class=['\"][^'\"]*metagame-percentage[^'\"]*['\"][^>]*>.*?<div[^>]*class=['\"][^'\"]*archetype-tile-statistic-value[^'\"]*['\"][^>]*>\s*(\d+\.?\d*)%`)

	// Also try to match the metagame table format
	tablePattern := regexp.MustCompile(`(?s)<tr[^>]*>.*?<a[^>]*href="/archetype/([^"#]+)[^"]*"[^>]*>([^<]+)</a>.*?<td[^>]*>(\d+\.?\d*)%</td>`)

	// Try archetype tiles first
	matches := archetypePattern.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		// Try fallback pattern
		matches = fallbackPattern.FindAllStringSubmatch(html, -1)
	}
	if len(matches) == 0 {
		// Fall back to table format
		matches = tablePattern.FindAllStringSubmatch(html, -1)
	}

	tier := 1
	tierThresholds := []float64{5.0, 2.0, 0.5} // Tier 1 > 5%, Tier 2 > 2%, Tier 3 > 0.5%

	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		var name string
		var shareStr string

		if len(match) == 3 {
			// Archetype tile format
			name = strings.TrimSpace(match[1])
			shareStr = strings.TrimSpace(match[2])
		} else if len(match) >= 4 {
			// Table format
			name = strings.TrimSpace(match[2])
			shareStr = strings.TrimSpace(match[3])
		}

		// Parse meta share
		shareStr = strings.TrimSuffix(shareStr, "%")
		share, err := strconv.ParseFloat(shareStr, 64)
		if err != nil {
			continue
		}

		// Determine tier
		for j, threshold := range tierThresholds {
			if share >= threshold {
				tier = j + 1
				break
			}
		}
		if share < tierThresholds[len(tierThresholds)-1] {
			tier = len(tierThresholds) + 1
		}

		// Extract colors from deck name
		colors := c.extractColorsFromName(name)

		deck := &MetaDeck{
			Name:          name,
			ArchetypeName: c.normalizeArchetypeName(name),
			Format:        format,
			Tier:          tier,
			MetaShare:     share,
			Colors:        colors,
			LastUpdated:   time.Now(),
		}

		meta.Decks = append(meta.Decks, deck)

		// Limit to top 50 decks
		if i >= 49 {
			break
		}
	}

	meta.TotalDecks = len(meta.Decks)

	return meta
}

// extractColorsFromName attempts to extract color identity from a deck name.
func (c *GoldfishClient) extractColorsFromName(name string) []string {
	nameLower := strings.ToLower(name)
	colors := make([]string, 0)

	// Color word mappings
	colorMappings := map[string]string{
		"white": "W", "mono-white": "W", "mono white": "W",
		"blue": "U", "mono-blue": "U", "mono blue": "U",
		"black": "B", "mono-black": "B", "mono black": "B",
		"red": "R", "mono-red": "R", "mono red": "R",
		"green": "G", "mono-green": "G", "mono green": "G",
		// Guild names
		"azorius": "WU", "dimir": "UB", "rakdos": "BR",
		"gruul": "RG", "selesnya": "WG", "orzhov": "WB",
		"izzet": "UR", "golgari": "BG", "boros": "WR",
		"simic": "UG",
		// Shard/Wedge names
		"esper": "WUB", "grixis": "UBR", "jund": "BRG",
		"naya": "WRG", "bant": "WUG", "abzan": "WBG",
		"jeskai": "WUR", "sultai": "UBG", "mardu": "WBR",
		"temur": "URG",
		// 4-color
		"glint": "UBRG", "dune": "WBRG", "ink": "WURG",
		"witch": "WUBG", "yore": "WUBR",
		// 5-color
		"five-color": "WUBRG", "5-color": "WUBRG", "5c": "WUBRG",
	}

	for word, colorStr := range colorMappings {
		if strings.Contains(nameLower, word) {
			for _, c := range colorStr {
				color := string(c)
				found := false
				for _, existing := range colors {
					if existing == color {
						found = true
						break
					}
				}
				if !found {
					colors = append(colors, color)
				}
			}
			break
		}
	}

	return colors
}

// normalizeArchetypeName normalizes an archetype name for comparison.
func (c *GoldfishClient) normalizeArchetypeName(name string) string {
	// Remove common suffixes and prefixes
	normalized := strings.ToLower(name)
	normalized = strings.TrimSpace(normalized)

	// Remove format prefixes
	prefixes := []string{"standard ", "historic ", "explorer ", "pioneer ", "modern "}
	for _, prefix := range prefixes {
		normalized = strings.TrimPrefix(normalized, prefix)
	}

	return normalized
}

// waitForRateLimit waits for rate limiting.
func (c *GoldfishClient) waitForRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	<-c.rateLimiter.C
	c.lastRequest = time.Now()
}

// getFromCache retrieves meta from cache if not expired.
func (c *GoldfishClient) getFromCache(format string) *FormatMeta {
	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()

	entry, exists := c.cache.data[strings.ToLower(format)]
	if !exists {
		return nil
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry.Meta
}

// setCache stores meta in cache.
func (c *GoldfishClient) setCache(format string, meta *FormatMeta) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	c.cache.data[strings.ToLower(format)] = &CacheEntry{
		Meta:      meta,
		ExpiresAt: time.Now().Add(c.cacheTTL),
	}
}

// ClearCache clears the meta cache.
func (c *GoldfishClient) ClearCache() {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	c.cache.data = make(map[string]*CacheEntry)
}

// RefreshMeta forces a refresh of meta data for a format.
func (c *GoldfishClient) RefreshMeta(ctx context.Context, format string) (*FormatMeta, error) {
	// Clear cache for this format
	c.cache.mu.Lock()
	delete(c.cache.data, strings.ToLower(format))
	c.cache.mu.Unlock()

	return c.GetMeta(ctx, format)
}

// GetCacheStatus returns cache status for a format.
func (c *GoldfishClient) GetCacheStatus(format string) (cached bool, expiresAt time.Time) {
	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()

	entry, exists := c.cache.data[strings.ToLower(format)]
	if !exists {
		return false, time.Time{}
	}

	return true, entry.ExpiresAt
}

// Serialize serializes meta data to JSON.
func (m *FormatMeta) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeFormatMeta deserializes meta data from JSON.
func DeserializeFormatMeta(data []byte) (*FormatMeta, error) {
	var meta FormatMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

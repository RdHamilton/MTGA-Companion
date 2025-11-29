package meta

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Top8Client fetches tournament data from MTGTop8.
type Top8Client struct {
	httpClient  *http.Client
	baseURL     string
	cache       *TournamentCache
	cacheTTL    time.Duration
	rateLimiter *time.Ticker
	lastRequest time.Time
	mu          sync.Mutex
}

// Top8Config configures the MTGTop8 client.
type Top8Config struct {
	// BaseURL is the MTGTop8 base URL.
	BaseURL string

	// CacheTTL is how long to cache tournament data.
	CacheTTL time.Duration

	// RequestTimeout is the HTTP request timeout.
	RequestTimeout time.Duration

	// RateLimitMs is minimum milliseconds between requests.
	RateLimitMs int
}

// DefaultTop8Config returns default configuration.
func DefaultTop8Config() *Top8Config {
	return &Top8Config{
		BaseURL:        "https://www.mtgtop8.com",
		CacheTTL:       6 * time.Hour,
		RequestTimeout: 30 * time.Second,
		RateLimitMs:    1500,
	}
}

// Tournament represents a tournament from MTGTop8.
type Tournament struct {
	Name        string     `json:"name"`
	Format      string     `json:"format"`
	Date        time.Time  `json:"date"`
	Players     int        `json:"players"`
	URL         string     `json:"url,omitempty"`
	TopDecks    []*TopDeck `json:"top_decks"`
	LastUpdated time.Time  `json:"last_updated"`
}

// TopDeck represents a deck that placed in a tournament.
type TopDeck struct {
	Placement      int        `json:"placement"` // 1st, 2nd, 3-4, 5-8, etc.
	ArchetypeName  string     `json:"archetype_name"`
	PlayerName     string     `json:"player_name,omitempty"`
	Colors         []string   `json:"colors"`
	MainboardCards []DeckCard `json:"mainboard,omitempty"`
	SideboardCards []DeckCard `json:"sideboard,omitempty"`
	URL            string     `json:"url,omitempty"`
}

// TournamentMeta aggregates tournament data for a format.
type TournamentMeta struct {
	Format           string                     `json:"format"`
	Tournaments      []*Tournament              `json:"tournaments"`
	ArchetypeStats   map[string]*ArchetypeStats `json:"archetype_stats"`
	TotalTournaments int                        `json:"total_tournaments"`
	DateRange        DateRange                  `json:"date_range"`
	LastUpdated      time.Time                  `json:"last_updated"`
	Source           string                     `json:"source"`
}

// DateRange represents a date range for tournament data.
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ArchetypeStats aggregates statistics for an archetype.
type ArchetypeStats struct {
	ArchetypeName   string    `json:"archetype_name"`
	TotalPlacements int       `json:"total_placements"`
	Top8Count       int       `json:"top8_count"`
	WinCount        int       `json:"win_count"` // 1st place finishes
	AvgPlacement    float64   `json:"avg_placement"`
	Colors          []string  `json:"colors"`
	PopularCards    []string  `json:"popular_cards,omitempty"`
	TrendDirection  string    `json:"trend_direction"` // "up", "down", "stable"
	LastSeen        time.Time `json:"last_seen"`
}

// TournamentCache caches tournament data.
type TournamentCache struct {
	data map[string]*TournamentCacheEntry
	mu   sync.RWMutex
}

// TournamentCacheEntry represents a cached tournament entry.
type TournamentCacheEntry struct {
	Meta      *TournamentMeta
	ExpiresAt time.Time
}

// NewTop8Client creates a new MTGTop8 client.
func NewTop8Client(config *Top8Config) *Top8Client {
	if config == nil {
		config = DefaultTop8Config()
	}

	return &Top8Client{
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		baseURL:     config.BaseURL,
		cacheTTL:    config.CacheTTL,
		rateLimiter: time.NewTicker(time.Duration(config.RateLimitMs) * time.Millisecond),
		cache: &TournamentCache{
			data: make(map[string]*TournamentCacheEntry),
		},
	}
}

// GetTournamentMeta retrieves aggregated tournament data for a format.
func (c *Top8Client) GetTournamentMeta(ctx context.Context, format string) (*TournamentMeta, error) {
	// Check cache first
	if cached := c.getFromCache(format); cached != nil {
		return cached, nil
	}

	// Rate limit
	c.waitForRateLimit()

	// Fetch from MTGTop8
	meta, err := c.fetchTournamentMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	// Cache result
	c.setCache(format, meta)

	return meta, nil
}

// GetRecentTournaments returns recent tournaments for a format.
func (c *Top8Client) GetRecentTournaments(ctx context.Context, format string, limit int) ([]*Tournament, error) {
	meta, err := c.GetTournamentMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(meta.Tournaments) {
		return meta.Tournaments, nil
	}

	return meta.Tournaments[:limit], nil
}

// GetArchetypeStats returns stats for a specific archetype.
func (c *Top8Client) GetArchetypeStats(ctx context.Context, format, archetype string) (*ArchetypeStats, error) {
	meta, err := c.GetTournamentMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	archetypeLower := strings.ToLower(archetype)
	for name, stats := range meta.ArchetypeStats {
		if strings.ToLower(name) == archetypeLower {
			return stats, nil
		}
	}

	return nil, fmt.Errorf("archetype not found: %s", archetype)
}

// GetTopArchetypes returns the top N archetypes by placements.
func (c *Top8Client) GetTopArchetypes(ctx context.Context, format string, limit int) ([]*ArchetypeStats, error) {
	meta, err := c.GetTournamentMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	// Convert map to slice and sort by top8 count
	archetypes := make([]*ArchetypeStats, 0, len(meta.ArchetypeStats))
	for _, stats := range meta.ArchetypeStats {
		archetypes = append(archetypes, stats)
	}

	// Sort by Top8Count descending
	for i := 0; i < len(archetypes)-1; i++ {
		for j := 0; j < len(archetypes)-i-1; j++ {
			if archetypes[j].Top8Count < archetypes[j+1].Top8Count {
				archetypes[j], archetypes[j+1] = archetypes[j+1], archetypes[j]
			}
		}
	}

	if limit <= 0 || limit > len(archetypes) {
		return archetypes, nil
	}

	return archetypes[:limit], nil
}

// fetchTournamentMeta fetches tournament data from MTGTop8.
func (c *Top8Client) fetchTournamentMeta(ctx context.Context, format string) (*TournamentMeta, error) {
	// Map format names to MTGTop8 format codes
	formatCodes := map[string]string{
		"standard": "ST",
		"historic": "HI",
		"explorer": "EX",
		"pioneer":  "PI",
		"modern":   "MO",
		"legacy":   "LE",
		"vintage":  "VI",
		"pauper":   "PAU",
	}

	code, ok := formatCodes[strings.ToLower(format)]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	url := fmt.Sprintf("%s/format?f=%s", c.baseURL, code)

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
	meta := c.parseTournamentPage(string(body), format)
	meta.Source = "mtgtop8"
	meta.LastUpdated = time.Now()

	return meta, nil
}

// parseTournamentPage parses the MTGTop8 tournament page HTML.
func (c *Top8Client) parseTournamentPage(html, format string) *TournamentMeta {
	meta := &TournamentMeta{
		Format:         format,
		Tournaments:    make([]*Tournament, 0),
		ArchetypeStats: make(map[string]*ArchetypeStats),
	}

	// Parse recent tournaments
	// MTGTop8 format page lists recent events with their top decks
	// <div class="event_title">Event Name</div>
	// <div class="deck">Archetype Name</div>

	// Pattern for tournament entries
	tournamentPattern := regexp.MustCompile(`(?s)<div[^>]*class="[^"]*event_title[^"]*"[^>]*>([^<]+)</div>`)
	deckPattern := regexp.MustCompile(`(?s)<div[^>]*class="[^"]*S_steep[^"]*"[^>]*>.*?href="[^"]*">([^<]+)</a>.*?<div[^>]*class="[^"]*deck[^"]*"[^>]*>([^<]+)</div>`)

	// Also try the archetype summary table
	archetypeTablePattern := regexp.MustCompile(`(?s)<tr[^>]*>.*?<a[^>]*href="[^"]*archetype[^"]*"[^>]*>([^<]+)</a>.*?<td[^>]*>(\d+)</td>`)

	// Parse archetype summary first
	archetypeMatches := archetypeTablePattern.FindAllStringSubmatch(html, -1)
	for _, match := range archetypeMatches {
		if len(match) < 3 {
			continue
		}

		archName := strings.TrimSpace(match[1])
		count, err := strconv.Atoi(strings.TrimSpace(match[2]))
		if err != nil {
			continue
		}

		normalized := c.normalizeArchetypeName(archName)
		if _, exists := meta.ArchetypeStats[normalized]; !exists {
			meta.ArchetypeStats[normalized] = &ArchetypeStats{
				ArchetypeName: archName,
				Colors:        c.extractColorsFromName(archName),
				LastSeen:      time.Now(),
			}
		}
		meta.ArchetypeStats[normalized].Top8Count += count
		meta.ArchetypeStats[normalized].TotalPlacements += count
	}

	// Parse tournaments
	tournamentMatches := tournamentPattern.FindAllStringSubmatch(html, -1)
	for _, match := range tournamentMatches {
		if len(match) < 2 {
			continue
		}

		tournament := &Tournament{
			Name:        strings.TrimSpace(match[1]),
			Format:      format,
			Date:        time.Now(), // Would need more parsing for actual date
			TopDecks:    make([]*TopDeck, 0),
			LastUpdated: time.Now(),
		}

		meta.Tournaments = append(meta.Tournaments, tournament)

		if len(meta.Tournaments) >= 20 {
			break
		}
	}

	// Parse deck entries
	deckMatches := deckPattern.FindAllStringSubmatch(html, -1)
	placement := 1
	for _, match := range deckMatches {
		if len(match) < 3 {
			continue
		}

		playerName := strings.TrimSpace(match[1])
		archName := strings.TrimSpace(match[2])

		deck := &TopDeck{
			Placement:     placement,
			ArchetypeName: archName,
			PlayerName:    playerName,
			Colors:        c.extractColorsFromName(archName),
		}

		// Add to most recent tournament
		if len(meta.Tournaments) > 0 {
			meta.Tournaments[len(meta.Tournaments)-1].TopDecks = append(
				meta.Tournaments[len(meta.Tournaments)-1].TopDecks,
				deck,
			)
		}

		// Update archetype stats
		normalized := c.normalizeArchetypeName(archName)
		if _, exists := meta.ArchetypeStats[normalized]; !exists {
			meta.ArchetypeStats[normalized] = &ArchetypeStats{
				ArchetypeName: archName,
				Colors:        deck.Colors,
				LastSeen:      time.Now(),
			}
		}
		stats := meta.ArchetypeStats[normalized]
		stats.TotalPlacements++
		if placement <= 8 {
			stats.Top8Count++
		}
		if placement == 1 {
			stats.WinCount++
		}

		placement++
		if placement > 8 {
			placement = 1 // Reset for next tournament
		}
	}

	// Calculate average placements
	for _, stats := range meta.ArchetypeStats {
		if stats.TotalPlacements > 0 {
			// Approximate average (without actual placement data)
			stats.AvgPlacement = float64(stats.Top8Count) / float64(stats.TotalPlacements) * 4.0
		}
	}

	meta.TotalTournaments = len(meta.Tournaments)
	meta.DateRange = DateRange{
		Start: time.Now().AddDate(0, 0, -30),
		End:   time.Now(),
	}

	return meta
}

// extractColorsFromName attempts to extract color identity from a deck name.
func (c *Top8Client) extractColorsFromName(name string) []string {
	nameLower := strings.ToLower(name)
	colors := make([]string, 0)

	colorMappings := map[string]string{
		"white": "W", "mono-white": "W", "mono white": "W",
		"blue": "U", "mono-blue": "U", "mono blue": "U",
		"black": "B", "mono-black": "B", "mono black": "B",
		"red": "R", "mono-red": "R", "mono red": "R",
		"green": "G", "mono-green": "G", "mono green": "G",
		"azorius": "WU", "dimir": "UB", "rakdos": "BR",
		"gruul": "RG", "selesnya": "WG", "orzhov": "WB",
		"izzet": "UR", "golgari": "BG", "boros": "WR",
		"simic": "UG",
		"esper": "WUB", "grixis": "UBR", "jund": "BRG",
		"naya": "WRG", "bant": "WUG", "abzan": "WBG",
		"jeskai": "WUR", "sultai": "UBG", "mardu": "WBR",
		"temur": "URG",
	}

	for word, colorStr := range colorMappings {
		if strings.Contains(nameLower, word) {
			for _, ch := range colorStr {
				color := string(ch)
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

// normalizeArchetypeName normalizes an archetype name.
func (c *Top8Client) normalizeArchetypeName(name string) string {
	normalized := strings.ToLower(name)
	normalized = strings.TrimSpace(normalized)
	return normalized
}

// waitForRateLimit waits for rate limiting.
func (c *Top8Client) waitForRateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	<-c.rateLimiter.C
	c.lastRequest = time.Now()
}

// getFromCache retrieves meta from cache if not expired.
func (c *Top8Client) getFromCache(format string) *TournamentMeta {
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
func (c *Top8Client) setCache(format string, meta *TournamentMeta) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	c.cache.data[strings.ToLower(format)] = &TournamentCacheEntry{
		Meta:      meta,
		ExpiresAt: time.Now().Add(c.cacheTTL),
	}
}

// ClearCache clears the tournament cache.
func (c *Top8Client) ClearCache() {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	c.cache.data = make(map[string]*TournamentCacheEntry)
}

// RefreshMeta forces a refresh of tournament data for a format.
func (c *Top8Client) RefreshMeta(ctx context.Context, format string) (*TournamentMeta, error) {
	c.cache.mu.Lock()
	delete(c.cache.data, strings.ToLower(format))
	c.cache.mu.Unlock()

	return c.GetTournamentMeta(ctx, format)
}

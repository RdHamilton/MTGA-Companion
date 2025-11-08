package cards

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Service provides card metadata lookup with caching.
type Service struct {
	scryfall *ScryfallClient
	db       *sql.DB
	cache    *Cache
	config   *ServiceConfig
}

// ServiceConfig holds configuration for the card service.
type ServiceConfig struct {
	// EnableCache enables in-memory caching of card data.
	EnableCache bool

	// CacheSize is the maximum number of cards to cache in memory.
	// Default: 10000
	CacheSize int

	// CacheTTL is the time-to-live for cached entries.
	// Default: 24 hours
	CacheTTL time.Duration

	// EnableDB enables database storage of card metadata.
	EnableDB bool

	// FallbackToAPI enables falling back to Scryfall API if card not found in cache/DB.
	// Default: true
	FallbackToAPI bool
}

// DefaultServiceConfig returns a ServiceConfig with sensible defaults.
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		EnableCache:   true,
		CacheSize:     10000,
		CacheTTL:      24 * time.Hour,
		EnableDB:      true,
		FallbackToAPI: true,
	}
}

// NewService creates a new card metadata service.
func NewService(db *sql.DB, config *ServiceConfig) (*Service, error) {
	if config == nil {
		config = DefaultServiceConfig()
	}

	var cache *Cache
	if config.EnableCache {
		cache = NewCache(config.CacheSize, config.CacheTTL)
	}

	s := &Service{
		scryfall: NewScryfallClient(),
		db:       db,
		cache:    cache,
		config:   config,
	}

	// Initialize database schema if DB is enabled
	if config.EnableDB && db != nil {
		if err := s.initDB(); err != nil {
			return nil, fmt.Errorf("failed to initialize card database: %w", err)
		}
	}

	return s, nil
}

// GetCard retrieves card metadata by Arena ID.
// It checks cache first, then database, then falls back to Scryfall API.
func (s *Service) GetCard(arenaID int) (*Card, error) {
	// Check cache first
	if s.config.EnableCache && s.cache != nil {
		if card := s.cache.Get(arenaID); card != nil {
			return card, nil
		}
	}

	// Check database
	if s.config.EnableDB && s.db != nil {
		card, err := s.getCardFromDB(arenaID)
		if err == nil && card != nil {
			// Add to cache
			if s.config.EnableCache && s.cache != nil {
				s.cache.Set(arenaID, card)
			}
			return card, nil
		}
		// If error is not "not found", return it
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("database lookup failed: %w", err)
		}
	}

	// Fallback to Scryfall API
	if s.config.FallbackToAPI {
		card, err := s.scryfall.GetCardByArenaID(arenaID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch card from Scryfall: %w", err)
		}

		// Store in database for future lookups
		if s.config.EnableDB && s.db != nil {
			_ = s.saveCardToDB(card) // Ignore errors, we have the card data
		}

		// Add to cache
		if s.config.EnableCache && s.cache != nil {
			s.cache.Set(arenaID, card)
		}

		return card, nil
	}

	return nil, fmt.Errorf("card with Arena ID %d not found", arenaID)
}

// GetCards retrieves multiple cards by their Arena IDs.
func (s *Service) GetCards(arenaIDs []int) (map[int]*Card, error) {
	cards := make(map[int]*Card, len(arenaIDs))
	var mu sync.Mutex

	// Use a worker pool to fetch cards concurrently
	type result struct {
		arenaID int
		card    *Card
		err     error
	}

	results := make(chan result, len(arenaIDs))
	sem := make(chan struct{}, 10) // Limit concurrent requests

	var wg sync.WaitGroup
	for _, id := range arenaIDs {
		wg.Add(1)
		go func(arenaID int) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			card, err := s.GetCard(arenaID)
			results <- result{arenaID: arenaID, card: card, err: err}
		}(id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for res := range results {
		if res.err == nil && res.card != nil {
			mu.Lock()
			cards[res.arenaID] = res.card
			mu.Unlock()
		}
	}

	return cards, nil
}

// initDB initializes the database schema for card storage.
func (s *Service) initDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS cards (
		arena_id INTEGER PRIMARY KEY,
		scryfall_id TEXT NOT NULL,
		oracle_id TEXT,
		multiverse_id INTEGER,
		name TEXT NOT NULL,
		type_line TEXT NOT NULL,
		set_code TEXT NOT NULL,
		set_name TEXT NOT NULL,
		mana_cost TEXT,
		cmc REAL NOT NULL,
		colors TEXT,
		color_identity TEXT,
		rarity TEXT NOT NULL,
		power TEXT,
		toughness TEXT,
		loyalty TEXT,
		oracle_text TEXT,
		flavor_text TEXT,
		image_uri TEXT,
		layout TEXT NOT NULL,
		collector_number TEXT NOT NULL,
		released_at TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_cards_name ON cards(name);
	CREATE INDEX IF NOT EXISTS idx_cards_set ON cards(set_code);
	CREATE INDEX IF NOT EXISTS idx_cards_scryfall_id ON cards(scryfall_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// getCardFromDB retrieves a card from the database.
func (s *Service) getCardFromDB(arenaID int) (*Card, error) {
	query := `
	SELECT arena_id, scryfall_id, oracle_id, multiverse_id, name, type_line,
	       set_code, set_name, mana_cost, cmc, colors, color_identity, rarity,
	       power, toughness, loyalty, oracle_text, flavor_text, image_uri,
	       layout, collector_number, released_at
	FROM cards
	WHERE arena_id = ?
	`

	var card Card
	var colorsJSON, colorIdentityJSON sql.NullString
	var releasedAtStr string

	err := s.db.QueryRow(query, arenaID).Scan(
		&card.ArenaID,
		&card.ScryfallID,
		&card.OracleID,
		&card.MultiverseID,
		&card.Name,
		&card.TypeLine,
		&card.SetCode,
		&card.SetName,
		&card.ManaCost,
		&card.CMC,
		&colorsJSON,
		&colorIdentityJSON,
		&card.Rarity,
		&card.Power,
		&card.Toughness,
		&card.Loyalty,
		&card.OracleText,
		&card.FlavorText,
		&card.ImageURI,
		&card.Layout,
		&card.CollectorNumber,
		&releasedAtStr,
	)

	if err != nil {
		return nil, err
	}

	// Parse colors
	if colorsJSON.Valid && colorsJSON.String != "" {
		_ = json.Unmarshal([]byte(colorsJSON.String), &card.Colors)
	}

	// Parse color identity
	if colorIdentityJSON.Valid && colorIdentityJSON.String != "" {
		_ = json.Unmarshal([]byte(colorIdentityJSON.String), &card.ColorIdentity)
	}

	// Parse release date
	card.ReleasedAt, _ = time.Parse("2006-01-02", releasedAtStr)

	return &card, nil
}

// saveCardToDB saves a card to the database.
func (s *Service) saveCardToDB(card *Card) error {
	colorsJSON, _ := json.Marshal(card.Colors)
	colorIdentityJSON, _ := json.Marshal(card.ColorIdentity)

	query := `
	INSERT INTO cards (
		arena_id, scryfall_id, oracle_id, multiverse_id, name, type_line,
		set_code, set_name, mana_cost, cmc, colors, color_identity, rarity,
		power, toughness, loyalty, oracle_text, flavor_text, image_uri,
		layout, collector_number, released_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(arena_id) DO UPDATE SET
		scryfall_id = excluded.scryfall_id,
		oracle_id = excluded.oracle_id,
		multiverse_id = excluded.multiverse_id,
		name = excluded.name,
		type_line = excluded.type_line,
		set_code = excluded.set_code,
		set_name = excluded.set_name,
		mana_cost = excluded.mana_cost,
		cmc = excluded.cmc,
		colors = excluded.colors,
		color_identity = excluded.color_identity,
		rarity = excluded.rarity,
		power = excluded.power,
		toughness = excluded.toughness,
		loyalty = excluded.loyalty,
		oracle_text = excluded.oracle_text,
		flavor_text = excluded.flavor_text,
		image_uri = excluded.image_uri,
		layout = excluded.layout,
		collector_number = excluded.collector_number,
		released_at = excluded.released_at,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.Exec(query,
		card.ArenaID,
		card.ScryfallID,
		card.OracleID,
		card.MultiverseID,
		card.Name,
		card.TypeLine,
		card.SetCode,
		card.SetName,
		card.ManaCost,
		card.CMC,
		string(colorsJSON),
		string(colorIdentityJSON),
		card.Rarity,
		card.Power,
		card.Toughness,
		card.Loyalty,
		card.OracleText,
		card.FlavorText,
		card.ImageURI,
		card.Layout,
		card.CollectorNumber,
		card.ReleasedAt.Format("2006-01-02"),
	)

	return err
}

// ImportBulkData imports bulk card data from Scryfall.
// This is useful for initializing the database with all Arena cards.
func (s *Service) ImportBulkData() error {
	if !s.config.EnableDB || s.db == nil {
		return fmt.Errorf("database is not enabled")
	}

	// Get bulk data info
	bulkDataList, err := s.scryfall.GetBulkDataInfo()
	if err != nil {
		return fmt.Errorf("failed to get bulk data info: %w", err)
	}

	// Find the "default_cards" bulk data
	var downloadURL string
	for _, bulk := range bulkDataList {
		if bulk.Type == "default_cards" {
			downloadURL = bulk.DownloadURI
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("default_cards bulk data not found")
	}

	// Download bulk data
	scryfallCards, err := s.scryfall.DownloadBulkData(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download bulk data: %w", err)
	}

	// Filter to only Arena cards and import them
	imported := 0
	for _, scryfallCard := range scryfallCards {
		// Only import cards with Arena IDs
		if scryfallCard.ArenaID <= 0 {
			continue
		}

		card := scryfallCard.ToCard()
		if err := s.saveCardToDB(card); err != nil {
			// Log error but continue importing
			continue
		}
		imported++
	}

	return nil
}

// Cache provides in-memory caching for card metadata.
type Cache struct {
	mu       sync.RWMutex
	cards    map[int]*cacheEntry
	maxSize  int
	ttl      time.Duration
	eviction *evictionQueue
}

// cacheEntry represents a cached card with expiration.
type cacheEntry struct {
	card      *Card
	expiresAt time.Time
}

// evictionQueue tracks cache entries for LRU eviction.
type evictionQueue struct {
	entries []int
}

// NewCache creates a new card metadata cache.
func NewCache(maxSize int, ttl time.Duration) *Cache {
	return &Cache{
		cards:    make(map[int]*cacheEntry, maxSize),
		maxSize:  maxSize,
		ttl:      ttl,
		eviction: &evictionQueue{entries: make([]int, 0, maxSize)},
	}
}

// Get retrieves a card from the cache.
func (c *Cache) Get(arenaID int) *Card {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cards[arenaID]
	if !exists {
		return nil
	}

	// Check if entry has expired
	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.card
}

// Set adds a card to the cache.
func (c *Cache) Set(arenaID int, card *Card) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entry if cache is full
	if len(c.cards) >= c.maxSize {
		c.evictOldest()
	}

	c.cards[arenaID] = &cacheEntry{
		card:      card,
		expiresAt: time.Now().Add(c.ttl),
	}

	c.eviction.entries = append(c.eviction.entries, arenaID)
}

// evictOldest removes the oldest entry from the cache.
func (c *Cache) evictOldest() {
	if len(c.eviction.entries) == 0 {
		return
	}

	oldest := c.eviction.entries[0]
	delete(c.cards, oldest)
	c.eviction.entries = c.eviction.entries[1:]
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cards = make(map[int]*cacheEntry, c.maxSize)
	c.eviction.entries = make([]int, 0, c.maxSize)
}

// Size returns the current number of entries in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cards)
}

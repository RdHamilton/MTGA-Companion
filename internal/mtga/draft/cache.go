package draft

import (
	"fmt"
	"sync"
	"time"
)

// CardRatingsCache provides in-memory caching for card ratings.
// This reduces redundant lookups during an active draft session.
type CardRatingsCache struct {
	entries  map[string]*cacheEntry
	mu       sync.RWMutex
	ttl      time.Duration
	enabled  bool
	maxSize  int
	stats    CacheStats
	lastCleanup time.Time
}

// cacheEntry represents a single cached card rating with timestamp.
type cacheEntry struct {
	rating    *CardRating
	timestamp time.Time
}

// CacheStats tracks cache performance metrics.
type CacheStats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	Size       int
	TotalSize  int
}

// NewCardRatingsCache creates a new card ratings cache.
// ttl: time-to-live for cache entries (0 = no expiration)
// maxSize: maximum number of entries (0 = unlimited)
// enabled: whether caching is enabled
func NewCardRatingsCache(ttl time.Duration, maxSize int, enabled bool) *CardRatingsCache {
	return &CardRatingsCache{
		entries:     make(map[string]*cacheEntry),
		ttl:         ttl,
		maxSize:     maxSize,
		enabled:     enabled,
		lastCleanup: time.Now(),
	}
}

// Get retrieves a cached card rating.
// Returns nil if not found or expired.
func (c *CardRatingsCache) Get(cardID int, colorFilter string) *CardRating {
	if !c.enabled {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(cardID, colorFilter)
	entry, ok := c.entries[key]
	if !ok {
		c.stats.Misses++
		return nil
	}

	// Check if expired
	if c.ttl > 0 && time.Since(entry.timestamp) > c.ttl {
		c.stats.Misses++
		// Don't delete here (would need write lock) - cleanup handles this
		return nil
	}

	c.stats.Hits++
	return entry.rating
}

// Set stores a card rating in the cache.
func (c *CardRatingsCache) Set(cardID int, colorFilter string, rating *CardRating) {
	if !c.enabled || rating == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(cardID, colorFilter)

	// Check if we need to evict entries
	if c.maxSize > 0 && len(c.entries) >= c.maxSize {
		if _, exists := c.entries[key]; !exists {
			// Evict oldest entry (simple FIFO strategy)
			c.evictOldest()
		}
	}

	c.entries[key] = &cacheEntry{
		rating:    rating,
		timestamp: time.Now(),
	}
	c.stats.Size = len(c.entries)
	c.stats.TotalSize++

	// Periodic cleanup of expired entries
	if time.Since(c.lastCleanup) > 5*time.Minute {
		go c.cleanupExpired()
	}
}

// Clear removes all entries from the cache.
func (c *CardRatingsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
	c.stats.Size = 0
}

// GetStats returns current cache statistics.
func (c *CardRatingsCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.entries)
	return stats
}

// GetHitRate returns the cache hit rate as a percentage.
func (c *CardRatingsCache) GetHitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stats.Hits + c.stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(c.stats.Hits) / float64(total) * 100.0
}

// Enable enables caching.
func (c *CardRatingsCache) Enable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = true
}

// Disable disables caching and clears all entries.
func (c *CardRatingsCache) Disable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = false
	c.entries = make(map[string]*cacheEntry)
	c.stats.Size = 0
}

// IsEnabled returns whether caching is currently enabled.
func (c *CardRatingsCache) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// makeKey creates a cache key from cardID and colorFilter.
func (c *CardRatingsCache) makeKey(cardID int, colorFilter string) string {
	return fmt.Sprintf("%d_%s", cardID, colorFilter)
}

// evictOldest removes the oldest cache entry (FIFO eviction).
// Caller must hold write lock.
func (c *CardRatingsCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.stats.Evictions++
	}
}

// cleanupExpired removes all expired entries from the cache.
// This runs periodically to prevent unbounded growth.
func (c *CardRatingsCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttl == 0 {
		c.lastCleanup = time.Now()
		return
	}

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.entries, key)
		}
	}

	c.stats.Size = len(c.entries)
	c.lastCleanup = now
}

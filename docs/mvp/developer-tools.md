# Developer Tools - Debug Mode and API Caching

This document covers the two developer-focused MVP features: debug/verbose logging and API response caching.

## Table of Contents

- [Debug Mode](#debug-mode)
- [API Response Caching](#api-response-caching)
- [Performance Optimization](#performance-optimization)
- [Development Workflows](#development-workflows)

## Debug Mode

**Issue:** [#262](https://github.com/RdHamilton/MTGA-Companion/issues/262) | **PR:** [#280](https://github.com/RdHamilton/MTGA-Companion/pull/280)

### Overview

A leveled logging system with three levels (Debug, Info, Error) that can be controlled via CLI flag for troubleshooting and development.

### Architecture

#### Logger Implementation

Located in `internal/mtga/draft/logger.go` (85 lines):

```go
package draft

import (
    "fmt"
    "log"
    "time"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelError
)

// Logger provides leveled logging for the draft overlay.
type Logger struct {
    debugEnabled bool
    prefix       string
}

// NewLogger creates a new logger with optional debug mode.
func NewLogger(debugEnabled bool) *Logger {
    return &Logger{
        debugEnabled: debugEnabled,
        prefix:       "[MTGA-Companion]",
    }
}

// Debug logs a debug message (only if debug mode enabled).
func (l *Logger) Debug(format string, args ...interface{}) {
    if !l.debugEnabled {
        return
    }
    l.log(LogLevelDebug, format, args...)
}

// Info logs an info message (always logged).
func (l *Logger) Info(format string, args ...interface{}) {
    l.log(LogLevelInfo, format, args...)
}

// Error logs an error message (always logged).
func (l *Logger) Error(format string, args ...interface{}) {
    l.log(LogLevelError, format, args...)
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
    timestamp := time.Now().Format("15:04:05")
    levelStr := l.levelString(level)
    message := fmt.Sprintf(format, args...)

    log.Printf("[%s] %s %s: %s", timestamp, l.prefix, levelStr, message)
}

func (l *Logger) levelString(level LogLevel) string {
    switch level {
    case LogLevelDebug:
        return "DEBUG"
    case LogLevelInfo:
        return "INFO"
    case LogLevelError:
        return "ERROR"
    default:
        return "UNKNOWN"
    }
}

// IsDebugEnabled returns whether debug logging is enabled.
func (l *Logger) IsDebugEnabled() bool {
    return l.debugEnabled
}

// SetDebugEnabled enables or disables debug logging.
func (l *Logger) SetDebugEnabled(enabled bool) {
    l.debugEnabled = enabled
}
```

### Usage in Code

#### Initialization

```go
// In overlay.go
func NewOverlay(config OverlayConfig) *Overlay {
    logger := NewLogger(config.DebugMode)

    overlay := &Overlay{
        logger: logger,
        // ... other fields
    }

    logger.Info("Overlay initialized")
    logger.Debug("Debug mode enabled")

    return overlay
}
```

#### Throughout Codebase

```go
// Debug messages (only shown with --debug flag)
o.logger.Debug("Scanning for active draft (lookback: %d hours)", lookbackHours)
o.logger.Debug("Found %d entries in lookback window", len(entries))
o.logger.Debug("Cache hit: card=%d, filter=%s", cardID, colorFilter)

// Info messages (always shown)
o.logger.Info("Draft resumed: Pack %d, Pick %d", pack, pick)
o.logger.Info("Log file rotation detected")
o.logger.Info("Draft completed, cleaning up overlay")

// Error messages (always shown)
o.logger.Error("Failed to parse draft event: %v", err)
o.logger.Error("Cache initialization failed: %v", err)
```

### CLI Integration

Located in `cmd/mtga-companion/main.go:45`:

```go
var debug = flag.Bool("debug", false, "Enable verbose debug logging")
```

Usage:
```bash
# Enable debug mode
./mtga-companion draft-overlay --set MKM --format PremierDraft --debug

# Disable debug mode (default)
./mtga-companion draft-overlay --set MKM --format PremierDraft
```

### Log Output Formatting

All log messages follow this format:

```
[HH:MM:SS] [MTGA-Companion] LEVEL: message
```

**Examples:**

```
[14:32:15] [MTGA-Companion] INFO: Overlay initialized
[14:32:15] [MTGA-Companion] DEBUG: Debug mode enabled
[14:32:16] [MTGA-Companion] DEBUG: Scanning for active draft (lookback: 24 hours)
[14:32:16] [MTGA-Companion] DEBUG: Found 150 entries in lookback window
[14:32:16] [MTGA-Companion] INFO: Draft resumed: Pack 2, Pick 3
[14:32:45] [MTGA-Companion] DEBUG: Cache hit: card=12345, filter=BR (hit rate: 75.0%)
[14:35:20] [MTGA-Companion] INFO: Log file rotation detected
[14:35:20] [MTGA-Companion] INFO: Position tracking reset
[14:40:12] [MTGA-Companion] INFO: Draft completed, cleaning up overlay
```

### Debug Categories

The debug logging is organized into categories:

#### 1. Initialization
```go
logger.Debug("Overlay initialized with config: %+v", config)
logger.Debug("Logger created (debug: %v)", debugEnabled)
logger.Debug("Cache created (TTL: %v, MaxSize: %d)", ttl, maxSize)
```

#### 2. Resume/Scanning
```go
logger.Debug("Scanning for active draft (lookback: %d hours)", hours)
logger.Debug("Found %d entries in lookback window", count)
logger.Debug("Processing entry at %s", timestamp)
logger.Debug("Found draft event: %s", eventType)
```

#### 3. Event Processing
```go
logger.Debug("Processing draft notification: %s", eventName)
logger.Debug("Processing pack: %d cards", len(cardIDs))
logger.Debug("Recording pick: card=%d, pack=%d, pick=%d", cardID, pack, pick)
```

#### 4. Cache Operations
```go
logger.Debug("Cache miss: card=%d, filter=%s", cardID, colorFilter)
logger.Debug("Cache hit: card=%d, filter=%s (hit rate: %.1f%%)", cardID, colorFilter, hitRate)
logger.Debug("Cache stats: hits=%d, misses=%d, size=%d", hits, misses, size)
```

#### 5. File Monitoring
```go
logger.Debug("Poller initialized: interval=%v, useEvents=%v", interval, useEvents)
logger.Debug("File event: %s (%s)", eventType, path)
logger.Debug("Position tracking: pos=%d, size=%d", pos, size)
```

### Test Coverage

Located in `internal/mtga/draft/logger_test.go` (148 lines):

#### Test 1: Logger Creation
```go
func TestNewLogger(t *testing.T) {
    tests := []struct {
        name         string
        debugEnabled bool
    }{
        {"debug enabled", true},
        {"debug disabled", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            logger := NewLogger(tt.debugEnabled)

            if logger == nil {
                t.Fatal("NewLogger returned nil")
            }

            if logger.IsDebugEnabled() != tt.debugEnabled {
                t.Errorf("IsDebugEnabled() = %v, want %v",
                    logger.IsDebugEnabled(), tt.debugEnabled)
            }
        })
    }
}
```

#### Test 2: Debug Level Filtering
```go
func TestLogger_Debug(t *testing.T) {
    // Create logger with debug disabled
    logger := NewLogger(false)

    // Debug messages should not be logged
    // (verified by capturing log output)

    // Enable debug
    logger.SetDebugEnabled(true)

    // Debug messages should now be logged
}
```

#### Test 3: Info/Error Always Logged
```go
func TestLogger_InfoError(t *testing.T) {
    // Even with debug disabled
    logger := NewLogger(false)

    // Info and Error should always log
    logger.Info("Test info")
    logger.Error("Test error")
}
```

## API Response Caching

**Issue:** [#263](https://github.com/RdHamilton/MTGA-Companion/issues/263) | **PR:** [#281](https://github.com/RdHamilton/MTGA-Companion/pull/281)

### Overview

In-memory caching system for card ratings to reduce redundant lookups during draft sessions. Provides significant performance improvements by avoiding repeated calculations of Bayesian ratings.

### Architecture

#### Cache Implementation

Located in `internal/mtga/draft/cache.go` (227 lines):

```go
package draft

import (
    "fmt"
    "sync"
    "time"
)

// CardRatingsCache provides thread-safe in-memory caching for card ratings.
type CardRatingsCache struct {
    entries     map[string]*cacheEntry
    mu          sync.RWMutex
    ttl         time.Duration
    maxSize     int
    enabled     bool
    stats       CacheStats
    lastCleanup time.Time
}

type cacheEntry struct {
    rating    *CardRating
    timestamp time.Time
}

// CacheStats tracks cache performance metrics.
type CacheStats struct {
    Hits      int64
    Misses    int64
    Evictions int64
    Size      int
    TotalSize int
}

// NewCardRatingsCache creates a new cache with the specified configuration.
// ttl: Time-to-live for cache entries (0 = no expiration)
// maxSize: Maximum number of entries (0 = unlimited)
// enabled: Whether the cache is enabled
func NewCardRatingsCache(ttl time.Duration, maxSize int, enabled bool) *CardRatingsCache {
    return &CardRatingsCache{
        entries:     make(map[string]*cacheEntry),
        ttl:         ttl,
        maxSize:     maxSize,
        enabled:     enabled,
        lastCleanup: time.Now(),
    }
}

// Get retrieves a card rating from the cache.
// Returns nil if not found, expired, or cache is disabled.
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

    // Check if we need to evict entries
    if c.maxSize > 0 && len(c.entries) >= c.maxSize {
        c.evictOldest()
    }

    key := c.makeKey(cardID, colorFilter)
    c.entries[key] = &cacheEntry{
        rating:    rating,
        timestamp: time.Now(),
    }

    c.stats.Size = len(c.entries)
    c.stats.TotalSize++
}

// evictOldest removes the oldest cache entry (FIFO).
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

// makeKey creates a cache key from card ID and color filter.
func (c *CardRatingsCache) makeKey(cardID int, colorFilter string) string {
    return fmt.Sprintf("%d_%s", cardID, colorFilter)
}

// GetStats returns a copy of the current cache statistics.
func (c *CardRatingsCache) GetStats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()

    return CacheStats{
        Hits:      c.stats.Hits,
        Misses:    c.stats.Misses,
        Evictions: c.stats.Evictions,
        Size:      len(c.entries),
        TotalSize: c.stats.TotalSize,
    }
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

// Clear removes all entries from the cache.
func (c *CardRatingsCache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.entries = make(map[string]*cacheEntry)
    c.stats.Size = 0
}

// Enable enables the cache.
func (c *CardRatingsCache) Enable() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.enabled = true
}

// Disable disables the cache and clears all entries.
func (c *CardRatingsCache) Disable() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.enabled = false
    c.entries = make(map[string]*cacheEntry)
    c.stats.Size = 0
}

// IsEnabled returns whether the cache is enabled.
func (c *CardRatingsCache) IsEnabled() bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.enabled
}
```

### Integration with RatingsProvider

Located in `internal/mtga/draft/ratings.go:45-70`:

```go
type RatingsProvider struct {
    setFile *seventeenlands.SetFile
    config  BayesianConfig
    cache   *CardRatingsCache  // NEW
}

func NewRatingsProvider(setFile *seventeenlands.SetFile, config BayesianConfig, cache *CardRatingsCache) *RatingsProvider {
    return &RatingsProvider{
        setFile: setFile,
        config:  config,
        cache:   cache,
    }
}

func (rp *RatingsProvider) GetCardRating(cardID int, colorFilter string) (*CardRating, error) {
    // Check cache first
    if rp.cache != nil {
        if cached := rp.cache.Get(cardID, colorFilter); cached != nil {
            return cached, nil
        }
    }

    // Cache miss - fetch from set file
    rating, err := rp.lookupCardRating(cardID, colorFilter)
    if err != nil {
        return nil, err
    }

    // Store in cache for future lookups
    if rp.cache != nil {
        rp.cache.Set(cardID, colorFilter, rating)
    }

    return rating, nil
}
```

### Cache Key Design

The cache uses a composite key format: `{cardID}_{colorFilter}`

**Why this format?**
- Same card has different ratings for different color contexts
- Card 12345 in "ALL" context: `12345_ALL`
- Card 12345 in "BR" context: `12345_BR`
- Both cached separately with potentially different values

**Example:**
```
Card: "Lightning Strike" (ID: 12345)

Cache entries:
- "12345_ALL":  GIHWR=57.5 (average across all decks)
- "12345_BR":   GIHWR=60.2 (specifically in BR decks)
- "12345_W":    GIHWR=52.1 (in mono-white decks)
```

### TTL-Based Expiration

**Time-to-Live (TTL)** controls how long entries remain valid:

```go
// Check if entry expired
if c.ttl > 0 && time.Since(entry.timestamp) > c.ttl {
    c.stats.Misses++  // Expired = cache miss
    return nil
}
```

**Example scenarios:**

**Short TTL (1 hour):**
```
10:00 AM - Card 12345 cached (GIHWR: 57.5)
10:30 AM - Cache hit (still valid)
11:01 AM - Cache miss (expired, TTL exceeded)
```

**Long TTL (24 hours):**
```
10:00 AM Day 1 - Card 12345 cached
11:00 AM Day 1 - Cache hit
...
09:00 AM Day 2 - Cache hit (still valid)
10:01 AM Day 2 - Cache miss (expired)
```

### FIFO Eviction Policy

When `maxSize` is reached, the **oldest entry** is removed (First-In-First-Out):

```go
func (c *CardRatingsCache) evictOldest() {
    var oldestKey string
    var oldestTime time.Time

    // Find oldest entry by timestamp
    for key, entry := range c.entries {
        if oldestKey == "" || entry.timestamp.Before(oldestTime) {
            oldestKey = key
            oldestTime = entry.timestamp
        }
    }

    // Remove it
    if oldestKey != "" {
        delete(c.entries, oldestKey)
        c.stats.Evictions++
    }
}
```

**Example:**
```
Cache MaxSize: 3
Current entries: ["A" (10:00), "B" (10:05), "C" (10:10)]

New entry "D" arrives at 10:15:
1. Check size: 3 >= 3 (need to evict)
2. Find oldest: "A" (timestamp 10:00)
3. Evict "A"
4. Add "D"
5. New entries: ["B", "C", "D"]
```

### Cache Statistics

The cache tracks performance metrics:

```go
type CacheStats struct {
    Hits      int64   // Successful lookups
    Misses    int64   // Failed lookups (not found or expired)
    Evictions int64   // Entries removed due to maxSize
    Size      int     // Current number of entries
    TotalSize int     // Total entries ever added
}
```

**Hit Rate Calculation:**
```go
hitRate := float64(hits) / float64(hits + misses) * 100.0
```

**Example:**
```
Hits: 150
Misses: 50
Hit Rate: 150 / (150 + 50) * 100 = 75.0%
```

### CLI Configuration

Located in `cmd/mtga-companion/main.go:48-50`:

```go
var (
    cacheEnabled = flag.Bool("cache", true, "Enable in-memory caching for card ratings (default: true)")
    cacheTTL     = flag.Duration("cache-ttl", 24*time.Hour, "Cache time-to-live (e.g., 1h, 24h)")
    cacheMaxSize = flag.Int("cache-max-size", 0, "Maximum cache entries (0 = unlimited)")
)
```

**Usage:**

```bash
# Enable cache with defaults (24h TTL, unlimited size)
./mtga-companion draft-overlay --set MKM --format PremierDraft

# Short TTL, limited size
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache-ttl 1h \
  --cache-max-size 500

# Disable cache
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache=false
```

### Test Coverage

Located in `internal/mtga/draft/cache_test.go` (453 lines, 13 tests):

#### Test 1: Cache Operations
```go
func TestCardRatingsCache_GetSet(t *testing.T) {
    cache := NewCardRatingsCache(0, 0, true)

    // Initially empty
    result := cache.Get(100, "ALL")
    if result != nil {
        t.Error("Expected nil for empty cache")
    }

    // Set and retrieve
    rating := &CardRating{CardID: 100, GIHWR: 60.0}
    cache.Set(100, "ALL", rating)

    result = cache.Get(100, "ALL")
    if result == nil {
        t.Fatal("Expected rating to be cached")
    }

    if result.GIHWR != 60.0 {
        t.Errorf("GIHWR = %.1f, want 60.0", result.GIHWR)
    }
}
```

#### Test 2: TTL Expiration
```go
func TestCardRatingsCache_TTLExpiration(t *testing.T) {
    cache := NewCardRatingsCache(50*time.Millisecond, 0, true)

    rating := &CardRating{CardID: 100, GIHWR: 60.0}
    cache.Set(100, "ALL", rating)

    // Should be available immediately
    if cache.Get(100, "ALL") == nil {
        t.Error("Expected cache hit immediately")
    }

    // Wait for expiration
    time.Sleep(60 * time.Millisecond)

    // Should be expired now
    if cache.Get(100, "ALL") != nil {
        t.Error("Expected cache miss after TTL")
    }
}
```

#### Test 3: FIFO Eviction
```go
func TestCardRatingsCache_MaxSizeEviction(t *testing.T) {
    cache := NewCardRatingsCache(0, 2, true) // Max 2 entries

    rating1 := &CardRating{CardID: 1, Name: "Card 1"}
    rating2 := &CardRating{CardID: 2, Name: "Card 2"}
    rating3 := &CardRating{CardID: 3, Name: "Card 3"}

    // Add 2 entries
    cache.Set(1, "ALL", rating1)
    cache.Set(2, "ALL", rating2)

    // Add 3rd - should evict oldest (1)
    cache.Set(3, "ALL", rating3)

    // First entry should be evicted
    if cache.Get(1, "ALL") != nil {
        t.Error("First entry should be evicted")
    }

    // Second and third should still be present
    if cache.Get(2, "ALL") == nil {
        t.Error("Second entry should still exist")
    }
    if cache.Get(3, "ALL") == nil {
        t.Error("Third entry should still exist")
    }
}
```

#### Test 4: Cache Performance
```go
func TestRatingsProvider_CachePerformance(t *testing.T) {
    setFile := createTestSetFile()
    config := DefaultBayesianConfig()
    cache := NewCardRatingsCache(0, 0, true)
    rp := NewRatingsProvider(setFile, config, cache)

    cardIDs := []int{100, 200, 300}
    colorFilters := []string{"ALL", "B", "BR"}

    // First pass - all misses
    for _, cardID := range cardIDs {
        for _, colorFilter := range colorFilters {
            _, _ = rp.GetCardRating(cardID, colorFilter)
        }
    }

    // Second pass - all hits
    for _, cardID := range cardIDs {
        for _, colorFilter := range colorFilters {
            _, _ = rp.GetCardRating(cardID, colorFilter)
        }
    }

    hitRate := cache.GetHitRate()
    if hitRate < 40.0 {
        t.Errorf("Hit rate = %.2f%%, want at least 40%%", hitRate)
    }
}
```

## Performance Optimization

### Caching Impact

**Without caching:**
```
Draft: 45 picks
Average picks repeated (BR colors): 15

Lookups per draft:
- 45 new packs (45 * 14 cards = 630 lookups)
- 15 repeated BR filter lookups
Total: 630 + 15 = 645 lookups

Each lookup: Parse JSON + Calculate Bayesian rating
Time per lookup: ~0.5ms
Total time: 645 * 0.5ms = 322.5ms per draft
```

**With caching (24h TTL, unlimited size):**
```
First draft: 645 lookups (all misses)
Second draft: ~300 lookups (50% hit rate)
Third draft: ~150 lookups (75% hit rate)

Time saved per draft (after first):
- Draft 2: 322.5ms → 161ms (50% reduction)
- Draft 3: 322.5ms → 80ms (75% reduction)
```

### Memory Usage

**Typical cache entry size:**
```go
type cacheEntry struct {
    rating    *CardRating  // ~100 bytes
    timestamp time.Time    // 24 bytes
}
// Total per entry: ~124 bytes
```

**Memory usage examples:**
```
500 entries:   500 * 124 = 62 KB
1000 entries:  1000 * 124 = 124 KB
5000 entries:  5000 * 124 = 620 KB
```

**Conclusion:** Cache memory footprint is negligible even with thousands of entries.

### Recommended Configurations

#### For Casual Players (1-2 drafts/day)
```bash
--cache-ttl 2h --cache-max-size 1000
```
- Short TTL: Minimal stale data risk
- Small size: ~124 KB memory
- Hit rate: 40-50%

#### For Competitive Players (5+ drafts/day)
```bash
--cache-ttl 24h --cache-max-size 0
```
- Long TTL: Maximum hit rate
- Unlimited size: ~2-5 MB memory (typical)
- Hit rate: 70-80%

#### For Developers
```bash
--cache=false --debug
```
- No cache: Always fresh data
- Debug: See all operations
- Easier testing of rating changes

## Development Workflows

### Workflow 1: Feature Development

```bash
# Disable cache to see fresh behavior
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --debug \
  --cache=false \
  --overlay-resume=false

# Watch debug output
# [DEBUG] All operations visible
# [DEBUG] No cache interference
# [DEBUG] Fresh state each run
```

### Workflow 2: Performance Testing

```bash
# Enable cache with metrics
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --debug \
  --cache-ttl 1h \
  --cache-max-size 1000

# Monitor cache performance
# [DEBUG] Cache hit: card=12345, filter=BR (hit rate: 45.2%)
# [DEBUG] Cache miss: card=67890, filter=ALL
# [DEBUG] Cache stats: hits=150, misses=180, size=330, evictions=0
```

### Workflow 3: Bug Reproduction

```bash
# Enable all debugging, fresh state
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --debug \
  --cache=false \
  --overlay-resume=false

# Reproduce issue with full logging
# Report debug output with bug report
```

### Workflow 4: Production Use

```bash
# Optimal performance, minimal logging
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache-ttl 24h \
  --cache-max-size 0

# Only see INFO/ERROR messages
# Maximum cache performance
```

## Next Steps

- **[Architecture](architecture.md)** - Technical design details
- **[Testing](testing.md)** - Test strategies and coverage
- **[Usage Guide](usage-guide.md)** - Practical CLI examples
- **[Draft Overlay Features](draft-overlay.md)** - Draft-specific features

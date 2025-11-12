# MVP Architecture

This document describes the technical architecture and design decisions for the MVP features.

## Table of Contents

- [System Overview](#system-overview)
- [Component Architecture](#component-architecture)
- [Data Flow](#data-flow)
- [Concurrency Model](#concurrency-model)
- [Design Patterns](#design-patterns)
- [Error Handling](#error-handling)
- [Performance Considerations](#performance-considerations)

## System Overview

The MTGA-Companion Draft Overlay is a real-time monitoring and analysis system that:

1. **Monitors** MTGA's Player.log file for draft events
2. **Parses** JSON log entries into structured data
3. **Analyzes** draft packs with 17Lands ratings
4. **Displays** recommendations in an overlay window
5. **Manages** application lifecycle (start, resume, cleanup)

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    MTGA Application                     │
│               (Writes to Player.log)                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                   File System                           │
│              Player.log (JSON entries)                  │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼ (fsnotify / polling)
┌─────────────────────────────────────────────────────────┐
│              Log Reader (Poller)                        │
│  - File system monitoring                               │
│  - Rotation detection                                   │
│  - Entry parsing                                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼ (channel: LogEntry)
┌─────────────────────────────────────────────────────────┐
│              Draft Overlay                              │
│  - Event processing                                     │
│  - State management                                     │
│  - Resume functionality                                 │
│  - Draft end detection                                  │
└────────────────────┬────────────────────────────────────┘
                     │
                     ├───────────────────┐
                     ▼                   ▼
           ┌──────────────────┐  ┌──────────────────┐
           │ RatingsProvider  │  │     Logger       │
           │  - Card lookup   │  │  - Debug/Info    │
           │  - Bayesian calc │  │  - Error logging │
           │  - Cache integ.  │  └──────────────────┘
           └────────┬─────────┘
                    │
                    ▼
           ┌──────────────────┐
           │  Cache (Memory)  │
           │  - TTL tracking  │
           │  - FIFO eviction │
           │  - Statistics    │
           └──────────────────┘
```

## Component Architecture

### 1. Log Reader (`internal/mtga/logreader/`)

**Responsibilities:**
- Monitor MTGA Player.log file for new entries
- Handle log file rotation gracefully
- Parse JSON log entries
- Provide entries via channel

**Key Files:**
- `poller.go` (528 lines) - Main polling/monitoring logic
- `poller_test.go` - Test coverage
- `reader.go` - JSON parsing utilities
- `path.go` - Platform-specific path detection

**Design:**
```go
type Poller struct {
    path          string
    interval      time.Duration
    useFileEvents bool
    watcher       *fsnotify.Watcher
    lastPos       int64
    updates       chan *LogEntry
    errChan       chan error
    // ... other fields
}

type LogEntry struct {
    Raw       string
    Timestamp time.Time
    IsJSON    bool
    Data      map[string]interface{}
}
```

**File System Monitoring:**

Two modes are supported:

1. **Event-Based (fsnotify):**
   - Watches parent directory for file events
   - Reacts to WRITE, CREATE, REMOVE, RENAME events
   - More efficient (no constant polling)
   - Fallback to polling on failure

2. **Polling (timer-based):**
   - Checks file at regular intervals (default: 2s)
   - Size-based rotation detection
   - Works everywhere (no dependencies)

**Rotation Handling:**

```go
// Event-based detection
case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
    fmt.Printf("[INFO] Log rotation detected: %s\n", event.Op)
    p.resetPosition()

case event.Has(fsnotify.Create):
    if event.Name == p.path {
        fmt.Printf("[INFO] Log file recreated: %s\n", event.Name)
        p.checkForUpdates()
    }

// Size-based detection (fallback)
if stat.Size() < lastPos || stat.Size() < lastSize {
    fmt.Printf("[INFO] Rotation detected (size decreased)\n")
    p.resetPosition()
}
```

### 2. Draft Overlay (`internal/mtga/draft/`)

**Responsibilities:**
- Process draft events from log reader
- Maintain current draft state
- Provide card ratings via RatingsProvider
- Handle draft lifecycle (start, picks, end)
- Resume in-progress drafts

**Key Files:**
- `overlay.go` (800+ lines) - Core overlay logic
- `overlay_test.go` - Test coverage including resume tests
- `parser.go` - Event parsing
- `state.go` - Draft state management

**Design:**
```go
type Overlay struct {
    parser          *Parser
    ratingsProvider *RatingsProvider
    cache           *CardRatingsCache
    poller          *logreader.Poller
    logger          *Logger
    currentState    *DraftState
    updates         chan *OverlayUpdate
    mu              sync.RWMutex
}

type DraftState struct {
    Event       *DraftEvent
    Picks       []PickRecord
    CurrentPack *Pack
}

type OverlayUpdate struct {
    Type    UpdateType
    Pack    *Pack
    Ratings []*CardRating
}
```

**State Machine:**

```
┌─────────────┐
│    IDLE     │  No active draft
└──────┬──────┘
       │ Draft.Notify
       ▼
┌─────────────┐
│   ACTIVE    │  Draft in progress
│             │  - Process packs
│             │  - Record picks
│             │  - Send updates
└──────┬──────┘
       │ DraftStatus (InProgress: false)
       ▼
┌─────────────┐
│  COMPLETE   │  Draft ended
│             │  - Cleanup resources
│             │  - Close window
└─────────────┘
```

**Resume Functionality:**

The resume feature scans log history on startup:

```go
func (o *Overlay) scanForActiveDraft(lookbackHours int) error {
    cutoff := time.Now().Add(-time.Duration(lookbackHours) * time.Hour)

    // Read entries from log history
    entries, err := o.scanner.ScanSince(cutoff)
    if err != nil {
        return err
    }

    // Process relevant events
    for _, entry := range entries {
        // Find most recent draft state
        if draft, ok := entry.ParseDraftNotify(); ok {
            // Skip sealed events
            if strings.Contains(strings.ToLower(draft.EventName), "sealed") {
                continue
            }
            o.handleDraftNotify(draft)
        }

        if status, ok := entry.ParseDraftStatus(); ok {
            if status.InProgress {
                o.handleDraftStatus(status)
            }
        }
    }

    return nil
}
```

**Draft End Detection:**

```go
func (o *Overlay) handleDraftStatus(status DraftStatus) {
    o.mu.Lock()
    defer o.mu.Unlock()

    // Update state
    if o.currentState != nil && o.currentState.Event != nil {
        o.currentState.Event.InProgress = status.InProgress
    }

    // Draft ended - trigger cleanup
    if !status.InProgress {
        o.logger.Info("Draft completed, cleaning up overlay")
        o.cleanup()
    }
}

func (o *Overlay) cleanup() {
    // Close channels
    if o.updates != nil {
        close(o.updates)
        o.updates = nil
    }

    // Stop poller
    if o.poller != nil {
        o.poller.Stop()
        o.poller = nil
    }

    // Trigger window close
    if o.onClose != nil {
        o.onClose()
    }
}
```

### 3. Ratings Provider (`internal/mtga/draft/ratings.go`)

**Responsibilities:**
- Look up card ratings from 17Lands data
- Apply Bayesian adjustment to ratings
- Integrate with cache for performance
- Support color filtering

**Design:**
```go
type RatingsProvider struct {
    setFile *seventeenlands.SetFile
    config  BayesianConfig
    cache   *CardRatingsCache
}

type CardRating struct {
    CardID        int
    Name          string
    GIHWR         float64  // Games In Hand Win Rate
    BayesianGIHWR float64  // Bayesian-adjusted GIHWR
    ATA           float64  // Average Taken At
    ALSA          float64  // Average Last Seen At
    NumGames      int      // Sample size
}
```

**Bayesian Adjustment:**

Ratings with small sample sizes are adjusted toward the mean:

```go
func (rp *RatingsProvider) applyBayesianAdjustment(raw float64, numGames int) float64 {
    // Shrink toward mean based on sample size
    weight := float64(numGames) / (float64(numGames) + rp.config.PriorWeight)
    return weight*raw + (1-weight)*rp.config.GlobalMean
}
```

**Example:**
```
Card with 50 games: 70% raw GIHWR
Bayesian adjustment (prior weight = 100, mean = 50%):

weight = 50 / (50 + 100) = 0.333
adjusted = 0.333 * 70% + 0.667 * 50% = 56.6%

Card with 500 games: 70% raw GIHWR
weight = 500 / (500 + 100) = 0.833
adjusted = 0.833 * 70% + 0.167 * 50% = 66.7%

Higher sample size → closer to raw rating
Lower sample size → closer to mean (safer)
```

**Cache Integration:**

```go
func (rp *RatingsProvider) GetCardRating(cardID int, colorFilter string) (*CardRating, error) {
    // Try cache first
    if rp.cache != nil {
        if cached := rp.cache.Get(cardID, colorFilter); cached != nil {
            return cached, nil
        }
    }

    // Cache miss - lookup from set file
    rating, err := rp.lookupCardRating(cardID, colorFilter)
    if err != nil {
        return nil, err
    }

    // Store in cache
    if rp.cache != nil {
        rp.cache.Set(cardID, colorFilter, rating)
    }

    return rating, nil
}
```

### 4. Cache (`internal/mtga/draft/cache.go`)

**Responsibilities:**
- Store card ratings in memory
- Expire entries based on TTL
- Evict entries when size limit reached
- Track performance statistics
- Thread-safe operations

**Design:**
```go
type CardRatingsCache struct {
    entries     map[string]*cacheEntry  // Key: "cardID_colorFilter"
    mu          sync.RWMutex            // Protects entries and stats
    ttl         time.Duration           // Time-to-live for entries
    maxSize     int                     // Max entries (0 = unlimited)
    enabled     bool                    // Can be toggled on/off
    stats       CacheStats              // Performance tracking
    lastCleanup time.Time               // Last cleanup time
}

type cacheEntry struct {
    rating    *CardRating
    timestamp time.Time
}

type CacheStats struct {
    Hits      int64
    Misses    int64
    Evictions int64
    Size      int
    TotalSize int
}
```

**Thread Safety:**

Uses read-write mutex for concurrent access:

```go
// Read operation (multiple readers allowed)
func (c *CardRatingsCache) Get(cardID int, colorFilter string) *CardRating {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // ... lookup logic
}

// Write operation (exclusive access)
func (c *CardRatingsCache) Set(cardID int, colorFilter string, rating *CardRating) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // ... storage logic
}
```

### 5. Logger (`internal/mtga/draft/logger.go`)

**Responsibilities:**
- Provide leveled logging (Debug, Info, Error)
- Filter debug messages based on mode
- Format log output consistently
- Support runtime enable/disable

**Design:**
```go
type LogLevel int

const (
    LogLevelDebug LogLevel = iota
    LogLevelInfo
    LogLevelError
)

type Logger struct {
    debugEnabled bool
    prefix       string
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
    // Skip debug if not enabled
    if level == LogLevelDebug && !l.debugEnabled {
        return
    }

    timestamp := time.Now().Format("15:04:05")
    levelStr := l.levelString(level)
    message := fmt.Sprintf(format, args...)

    log.Printf("[%s] %s %s: %s", timestamp, l.prefix, levelStr, message)
}
```

## Data Flow

### Normal Draft Flow

```
1. MTGA writes draft event to Player.log
   {"Draft.Notify": {"EventName": "PremierDraft_MKM_20250112", ...}}

2. Poller detects new content (fsnotify WRITE event or polling)

3. Poller reads new lines, creates LogEntry
   LogEntry{Raw: "...", IsJSON: true, ...}

4. LogEntry sent through updates channel
   updates <- entry

5. Overlay receives LogEntry from channel
   for entry := range poller.Start() { ... }

6. Overlay parses event type
   if draft, ok := entry.ParseDraftNotify(); ok { ... }

7. Overlay updates state
   o.currentState.Event = draft

8. Overlay processes pack
   cardIDs := draft.PackCards
   for _, cardID := range cardIDs { ... }

9. RatingsProvider looks up ratings (cache-first)
   rating, err := rp.GetCardRating(cardID, colorFilter)

10. Cache returns rating (if hit) or nil (if miss)

11. RatingsProvider calculates Bayesian adjustment (if cache miss)

12. Rating stored in cache for future lookups

13. Overlay sends update to GUI
    o.updates <- &OverlayUpdate{Pack: pack, Ratings: ratings}

14. GUI displays recommendations
```

### Resume Flow

```
1. User starts overlay with --overlay-resume=true

2. Overlay calls scanForActiveDraft(24)

3. Scanner reads log entries from 24 hours ago

4. Scanner filters for draft events (skip sealed)

5. Scanner finds most recent draft state

6. Overlay restores state from found events
   - Draft event (name, pack, pick)
   - Previous picks (if any)
   - Current pack

7. Overlay continues normal monitoring from current position

8. User picks up where they left off
```

### Log Rotation Flow

```
1. MTGA decides to rotate log file (size/time threshold)

2. MTGA renames Player.log → Player-2025-01-12.log

3. Poller detects RENAME event (fsnotify)

4. Poller resets position tracking
   lastPos = 0
   lastSize = 0

5. Poller logs rotation event
   [INFO] Log rotation detected

6. MTGA creates new empty Player.log

7. Poller detects CREATE event

8. Poller logs recreation event
   [INFO] Log file recreated

9. Poller continues monitoring from position 0

10. No data loss, seamless transition
```

## Concurrency Model

The application uses goroutines and channels for concurrent operations:

### Goroutine Structure

```
┌────────────────┐
│   Main Thread  │
│  - Setup       │
│  - GUI loop    │
└────────┬───────┘
         │
         ├───────────────────────────────┐
         ▼                               ▼
┌────────────────┐            ┌──────────────────┐
│ Poller Goroutine│            │ Overlay Goroutine│
│  - File monitor│            │  - Event process │
│  - Read entries│            │  - State update  │
│  - Send to chan│            │  - Rating lookup │
└────────┬───────┘            └──────────┬───────┘
         │                               │
         │ (channel)                     │ (channel)
         │                               │
         ▼                               ▼
┌────────────────────────────────────────────────┐
│             Channel Buffers                    │
│  - updates: LogEntry                           │
│  - overlayUpdates: OverlayUpdate               │
└────────────────────────────────────────────────┘
```

### Channel Usage

**Poller → Overlay:**
```go
// Poller sends entries
type Poller struct {
    updates chan *LogEntry  // Buffer: 100
}

func (p *Poller) Start() <-chan *LogEntry {
    go p.poll()
    return p.updates
}

// Overlay receives entries
updates := poller.Start()
for entry := range updates {
    o.processEntry(entry)
}
```

**Overlay → GUI:**
```go
// Overlay sends updates
type Overlay struct {
    updates chan *OverlayUpdate  // Buffer: 10
}

func (o *Overlay) Start() <-chan *OverlayUpdate {
    go o.processLoop()
    return o.updates
}

// GUI receives updates
updates := overlay.Start()
for update := range updates {
    gui.Render(update)
}
```

### Synchronization

**RWMutex for State:**
```go
type Overlay struct {
    currentState *DraftState
    mu           sync.RWMutex
}

// Multiple readers
func (o *Overlay) GetCurrentPack() *Pack {
    o.mu.RLock()
    defer o.mu.RUnlock()
    return o.currentState.CurrentPack
}

// Single writer
func (o *Overlay) updateState(state *DraftState) {
    o.mu.Lock()
    defer o.mu.Unlock()
    o.currentState = state
}
```

**Cache Synchronization:**
```go
type CardRatingsCache struct {
    entries map[string]*cacheEntry
    mu      sync.RWMutex
    stats   CacheStats
}

// All operations protected by mutex
func (c *CardRatingsCache) Get(cardID int, colorFilter string) *CardRating {
    c.mu.RLock()
    defer c.mu.RUnlock()
    // ... safe concurrent reads
}
```

### Context-Based Cancellation

```go
type Poller struct {
    ctx    context.Context
    cancel context.CancelFunc
}

func NewPoller(config *PollerConfig) (*Poller, error) {
    ctx, cancel := context.WithCancel(context.Background())
    return &Poller{
        ctx:    ctx,
        cancel: cancel,
    }
}

func (p *Poller) Stop() {
    p.cancel()  // Signal goroutines to stop
    <-p.done    // Wait for completion
}

func (p *Poller) poll() {
    for {
        select {
        case <-p.ctx.Done():
            return  // Clean shutdown
        case <-ticker.C:
            p.checkForUpdates()
        }
    }
}
```

## Design Patterns

### 1. Producer-Consumer Pattern

**Poller (Producer) → Overlay (Consumer):**

```go
// Producer
func (p *Poller) poll() {
    for {
        entries := p.readNewEntries()
        for _, entry := range entries {
            select {
            case p.updates <- entry:  // Send to consumer
            case <-p.ctx.Done():
                return
            }
        }
    }
}

// Consumer
func (o *Overlay) processLoop() {
    for entry := range o.poller.Start() {
        o.processEntry(entry)
    }
}
```

### 2. Observer Pattern

**Overlay notifies GUI of updates:**

```go
type UpdateCallback func(*OverlayUpdate)

type OverlayConfig struct {
    UpdateCallback UpdateCallback
}

func (o *Overlay) sendUpdate(update *OverlayUpdate) {
    if o.updateCallback != nil {
        o.updateCallback(update)
    }
}
```

### 3. Strategy Pattern

**Bayesian configuration:**

```go
type BayesianConfig struct {
    GlobalMean  float64
    PriorWeight float64
}

func DefaultBayesianConfig() BayesianConfig {
    return BayesianConfig{
        GlobalMean:  50.0,
        PriorWeight: 100.0,
    }
}

func ConservativeBayesianConfig() BayesianConfig {
    return BayesianConfig{
        GlobalMean:  48.0,
        PriorWeight: 200.0,
    }
}
```

### 4. Singleton Pattern (Cache)

Only one cache instance per overlay:

```go
func NewOverlay(config OverlayConfig) *Overlay {
    var cache *CardRatingsCache
    if config.CacheEnabled {
        cache = NewCardRatingsCache(config.CacheTTL, config.CacheMaxSize, true)
    }

    ratingsProvider := NewRatingsProvider(setFile, bayesianConfig, cache)

    return &Overlay{
        ratingsProvider: ratingsProvider,
        cache:           cache,
    }
}
```

## Error Handling

### Error Propagation

```go
// Return errors up the stack
func (rp *RatingsProvider) GetCardRating(cardID int, colorFilter string) (*CardRating, error) {
    if rp.setFile == nil {
        return nil, fmt.Errorf("set file is nil")
    }

    rating, err := rp.lookupCardRating(cardID, colorFilter)
    if err != nil {
        return nil, fmt.Errorf("lookup card %d: %w", cardID, err)
    }

    return rating, nil
}
```

### Error Channels

```go
type Poller struct {
    errChan chan error
}

func (p *Poller) poll() {
    if err := p.checkForUpdates(); err != nil {
        select {
        case p.errChan <- err:
        default:
            // Channel full, skip
        }
    }
}

// Consumer can listen to errors
for {
    select {
    case entry := <-updates:
        process(entry)
    case err := <-poller.Errors():
        log.Printf("Error: %v", err)
    }
}
```

### Graceful Degradation

**Cache failure doesn't block overlay:**
```go
if cached := rp.cache.Get(cardID, colorFilter); cached != nil {
    return cached, nil
}
// Cache miss - continue without cache

rating, err := rp.lookupCardRating(cardID, colorFilter)
if err != nil {
    return nil, err
}

// Try to cache, but don't fail if cache errors
if rp.cache != nil {
    rp.cache.Set(cardID, colorFilter, rating)
}
```

**File events fallback to polling:**
```go
if p.useFileEvents {
    if err := p.setupWatcher(); err != nil {
        p.logger.Error("File watcher failed, falling back to polling: %v", err)
        p.pollWithTimer()
        return
    }
    p.pollWithEvents()
} else {
    p.pollWithTimer()
}
```

## Performance Considerations

### Memory Usage

**Cache:**
- ~124 bytes per entry (rating + timestamp)
- Configurable max size (default: unlimited)
- TTL-based expiration frees old entries
- Typical usage: 1-5 MB for active drafting session

**Buffers:**
- Poller updates channel: 100 entries (~10 KB)
- Overlay updates channel: 10 entries (~1 KB)
- Scanner buffers: 10 MB max for long JSON lines

**Total:**
- Baseline: ~20 MB (Go runtime + libraries)
- Active draft: ~25-30 MB
- With cache: ~30-35 MB

### CPU Usage

**Idle:**
- File event monitoring: near-zero CPU
- Polling (fallback): ~0.1% CPU (2s interval)

**Active Draft:**
- JSON parsing: ~5-10% CPU burst per event
- Bayesian calculation: ~1% CPU
- Cache lookup: negligible
- GUI rendering: ~5-10% CPU

**Total during draft:**
- Peak: ~20-25% CPU (one core)
- Average: ~5-10% CPU

### I/O Usage

**Read Operations:**
- Log file reading: only new bytes (incremental)
- Position tracking prevents re-reading old data
- Typical: 1-10 KB per draft event

**Write Operations:**
- None to log file (read-only)
- Cache writes: in-memory only

### Optimization Strategies

1. **Cache-first lookups** - Avoid redundant calculations
2. **Incremental log reading** - Only read new content
3. **Channel buffering** - Smooth out burst traffic
4. **RWMutex** - Allow concurrent reads
5. **Lazy cleanup** - Only evict when necessary

## Next Steps

- **[Testing](testing.md)** - Test strategies and coverage
- **[Usage Guide](usage-guide.md)** - Practical CLI examples
- **[Draft Overlay Features](draft-overlay.md)** - Draft-specific features
- **[Developer Tools](developer-tools.md)** - Debug mode and caching

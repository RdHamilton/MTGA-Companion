# 17Lands API Client

This package provides a Go client for the 17Lands draft statistics API with graceful fallback support.

## Features

- **Rate-limited API access** - Conservative 1 request/second rate limiting (configurable)
- **Automatic retries** - Exponential backoff on failures (2s, 4s, 8s, up to 60s)
- **Cache fallback** - Gracefully falls back to cached data when API is unavailable
- **Client statistics** - Track request counts, failures, and cache hits

## Usage

### Basic Usage (Without Cache)

```go
import "github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"

// Create client with default options
client := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

// Fetch card ratings
ratings, err := client.GetCardRatings(context.Background(), seventeenlands.QueryParams{
    Expansion: "BLB",
    Format:    "PremierDraft",
})
if err != nil {
    // Handle error - API unavailable
    return err
}

// Use ratings...
for _, rating := range ratings {
    fmt.Printf("%s: GIHWR %.2f%%\n", rating.Name, rating.GIHWR)
}
```

### With Cache Fallback (Recommended)

```go
import (
    "github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
    "github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Setup storage
db, err := storage.Open(ctx, dbPath)
if err != nil {
    return err
}
storageService := storage.NewService(db)

// Create storage adapter for caching
cache := seventeenlands.NewStorageAdapter(storageService)

// Create client with cache
options := seventeenlands.ClientOptions{
    RateLimit: seventeenlands.DefaultRateLimit,
    Timeout:   30 * time.Second,
    Cache:     cache, // Enable fallback
}
client := seventeenlands.NewClient(options)

// Fetch card ratings (will fall back to cache if API fails)
ratings, err := client.GetCardRatings(ctx, seventeenlands.QueryParams{
    Expansion: "BLB",
    Format:    "PremierDraft",
})
if err != nil {
    // Check error type
    if apiErr, ok := err.(*seventeenlands.APIError); ok {
        switch apiErr.Type {
        case seventeenlands.ErrStatsUnavailable:
            // No data available (API failed and no cache)
            log.Println("Draft statistics unavailable")
        case seventeenlands.ErrRateLimited:
            // Rate limited or in backoff period
            log.Println("Rate limited, try again later")
        case seventeenlands.ErrInvalidParams:
            // Invalid request parameters
            log.Println("Invalid parameters:", apiErr.Message)
        }
    }
    return err
}

// Use ratings (may be from API or cache)
```

## Fallback Behavior

The client implements a robust fallback strategy when the 17Lands API is unavailable:

```
Request Flow:
┌─────────────────┐
│  GetCardRatings │
└────────┬────────┘
         │
         ├─[Try API]──────────┐
         │                    │
         │              ┌─────▼──────┐
         │              │   Success  │
         │              └─────┬──────┘
         │                    │
         │              [Cache Response]
         │                    │
         │              [Return Ratings]
         │
         │              ┌─────▼──────┐
         ├─[API Failed]─┤  Failure   │
         │              └─────┬──────┘
         │                    │
         │              [Check Cache]
         │                    │
         │              ┌─────▼──────┐
         │         Yes  │Cache Found?│  No
         │          ┌───┤            ├───┐
         │          │   └────────────┘   │
         │   ┌──────▼────────┐    ┌─────▼────────┐
         │   │Return Cached  │    │     Error    │
         │   │ (even if stale)│    │ErrStatsUnav. │
         │   └───────────────┘    └──────────────┘
         │
         └──────────────┐
                        │
                  [Log & Stats]
```

### Cache Behavior

1. **API Success**: Data is cached automatically for future fallback
2. **API Failure**:
   - If cache available: Returns cached data (even if stale) + logs warning
   - If no cache: Returns `ErrStatsUnavailable` error
3. **Cache Save Failure**: Request still succeeds (caching is non-critical)

### Retry Strategy

The client uses exponential backoff on failures:

- **Initial backoff**: 2 seconds
- **Backoff multiplier**: 2x on each failure
- **Maximum backoff**: 60 seconds
- **Reset**: Backoff resets to 2s on successful request

During backoff period, requests return immediately with `ErrRateLimited` and fall back to cache if available.

## Error Types

The client returns `*APIError` with specific error types:

- `ErrRateLimited` - Rate limited or in backoff period
- `ErrUnavailable` - API returned error status (500, 503, etc.)
- `ErrInvalidParams` - Invalid request parameters
- `ErrParseError` - Failed to parse API response
- `ErrStatsUnavailable` - No data available (API failed, no cache)

## Client Statistics

Track client behavior using `GetStats()`:

```go
stats := client.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Failed requests: %d\n", stats.FailedRequests)
fmt.Printf("Cached responses: %d\n", stats.CachedResponses)
fmt.Printf("Average latency: %v\n", stats.AverageLatency)
fmt.Printf("Consecutive errors: %d\n", stats.ConsecutiveErrors)
```

## Logging

The client logs important events to standard logger:

- `[17Lands] API unavailable, attempting cache fallback: ...`
- `[17Lands] Using cached card ratings (age: 2h, count: 200)`
- `[17Lands] Cache miss or empty: ...`
- `[17Lands] Failed to cache card ratings: ...`

Configure log output as needed:

```go
import "log"

// Disable logging
log.SetOutput(io.Discard)

// Or redirect to custom logger
log.SetOutput(customWriter)
```

## Testing

The package includes comprehensive tests for fallback scenarios:

```bash
# Run all tests
go test ./internal/mtga/cards/seventeenlands/...

# Run specific fallback tests
go test ./internal/mtga/cards/seventeenlands/... -run TestGetCardRatings_APIFailure

# Run with verbose output
go test -v ./internal/mtga/cards/seventeenlands/...
```

Test scenarios covered:
- API success with caching
- API failure with cache hit
- API failure with cache miss
- Network timeout with fallback
- Invalid JSON responses
- Cache save failures
- Color ratings fallback

## Best Practices

### 1. Always Use Cache When Possible

```go
// ✅ GOOD - with cache fallback
cache := seventeenlands.NewStorageAdapter(storageService)
client := seventeenlands.NewClient(seventeenlands.ClientOptions{Cache: cache})

// ❌ AVOID - no fallback support
client := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())
```

### 2. Handle Unavailability Gracefully

```go
ratings, err := client.GetCardRatings(ctx, params)
if err != nil {
    if apiErr, ok := err.(*seventeenlands.APIError); ok {
        if apiErr.Type == seventeenlands.ErrStatsUnavailable {
            // Show UI without statistics
            return showBasicView(cards)
        }
    }
    return err
}

// Show UI with statistics
return showEnhancedView(cards, ratings)
```

### 3. Provide User Feedback

```go
// Check cache age from stats or implement CacheInfo method
if usingCachedData {
    ui.ShowWarning("Using cached statistics (last updated: 2 days ago)")
}
```

### 4. Monitor Client Statistics

```go
// Periodically check client health
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := client.GetStats()
        if stats.ConsecutiveErrors > 5 {
            alert.Send("17Lands API experiencing issues")
        }
    }
}()
```

## Implementation Details

### Cache Storage Interface

Implementations must provide:

```go
type CacheStorage interface {
    SaveCardRatings(ctx context.Context, ratings []CardRating,
        expansion, format, colors, startDate, endDate string) error

    GetCardRatingsForSet(ctx context.Context, expansion, format, colors string)
        ([]CardRating, time.Time, error)

    SaveColorRatings(ctx context.Context, ratings []ColorRating,
        expansion, eventType, startDate, endDate string) error

    GetColorRatings(ctx context.Context, expansion, eventType string)
        ([]ColorRating, time.Time, error)
}
```

### Storage Adapter

The `StorageAdapter` converts between:
- `storage.DraftCardRating` ↔ `seventeenlands.CardRating`
- `storage.DraftColorRating` ↔ `seventeenlands.ColorRating`

This allows the client to use the database layer without tight coupling.

## Related Documentation

- **Task #217**: [Graceful Fallback Implementation](https://github.com/RdHamilton/MTGA-Companion/issues/217)
- **Storage Layer**: `internal/storage/draft_statistics.go`
- **Database Schema**: Migration `000007_create_draft_statistics_tables`

## Contributing

When modifying the client:

1. Maintain backward compatibility
2. Add tests for new failure scenarios
3. Update this documentation
4. Follow project conventions (KISS, Effective Go)
5. Ensure rate limiting remains conservative

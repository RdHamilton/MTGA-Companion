# Scryfall API Client

A Go client library for the Scryfall API with built-in rate limiting and error handling.

## Features

- ✅ Rate limiting (100ms between requests, 10 req/sec)
- ✅ Automatic retry with exponential backoff
- ✅ HTTP 429 (rate limit) handling
- ✅ Comprehensive error handling
- ✅ Context support for cancellation/timeouts
- ✅ Full coverage of key Scryfall API endpoints

## Installation

```bash
go get golang.org/x/time/rate
```

## Usage

### Basic Example

```go
import (
    "context"
    "fmt"
    "github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

func main() {
    // Create client
    client := scryfall.NewClient()
    ctx := context.Background()

    // Get card by Arena ID
    card, err := client.GetCardByArenaID(ctx, 89019)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Card: %s\n", card.Name)
    fmt.Printf("Mana Cost: %s\n", card.ManaCost)
    fmt.Printf("Type: %s\n", card.TypeLine)
}
```

### Get Card by Scryfall ID

```go
card, err := client.GetCard(ctx, "1d72ab16-c3dd-4b92-ba1f-7a490a61f36f")
if err != nil {
    if scryfall.IsNotFound(err) {
        fmt.Println("Card not found")
    } else {
        log.Fatal(err)
    }
}
```

### Get Set Information

```go
set, err := client.GetSet(ctx, "blb")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Set: %s\n", set.Name)
fmt.Printf("Cards: %d\n", set.CardCount)
fmt.Printf("Released: %s\n", set.ReleasedAt)
```

### Search Cards

```go
// Search for Lightning Bolt
result, err := client.SearchCards(ctx, "!\"Lightning Bolt\"")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d results\n", result.TotalCards)
for _, card := range result.Data {
    fmt.Printf("- %s (%s)\n", card.Name, card.SetCode)
}
```

### Get Bulk Data

```go
bulkData, err := client.GetBulkData(ctx)
if err != nil {
    log.Fatal(err)
}

// Find default cards bulk data
for _, data := range bulkData.Data {
    if data.Type == "default_cards" {
        fmt.Printf("Download URL: %s\n", data.DownloadURI)
        fmt.Printf("Size: %.2f MB\n", float64(data.CompressedSize)/(1024*1024))
    }
}
```

### Get All Sets

```go
sets, err := client.GetSets(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Total sets: %d\n", len(sets.Data))
for _, set := range sets.Data {
    fmt.Printf("- %s (%s)\n", set.Name, set.Code)
}
```

## Error Handling

The client provides specific error types for common failures:

```go
card, err := client.GetCardByArenaID(ctx, 99999999)
if err != nil {
    if scryfall.IsNotFound(err) {
        // Handle not found error
        fmt.Println("Card does not exist")
    } else if apiErr, ok := err.(*scryfall.APIError); ok {
        // Handle API error
        fmt.Printf("API error: %s\n", apiErr.Details)
    } else {
        // Handle other errors
        log.Fatal(err)
    }
}
```

## Rate Limiting

The client automatically enforces rate limiting to comply with Scryfall's Terms of Service:

- **100ms delay** between requests (10 requests per second maximum)
- Automatic queuing via `golang.org/x/time/rate` limiter
- No manual delays required

## Retry Logic

Failed requests are automatically retried with exponential backoff:

- **Initial backoff**: 1 second
- **Maximum backoff**: 16 seconds
- **Maximum retries**: 3 attempts
- HTTP 429 responses trigger immediate backoff

## Context Support

All methods accept a `context.Context` parameter for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

card, err := client.GetCard(ctx, cardID)
```

## Testing

### Unit Tests

```bash
go test ./internal/mtga/cards/scryfall -v
```

### Integration Tests (real API calls)

```bash
go test ./internal/mtga/cards/scryfall -tags=integration -v
```

## Compliance

This client complies with Scryfall's API Terms of Service:

✅ Rate limiting (50-100ms between requests)
✅ User-Agent header set
✅ Caching support (implementation in storage layer)
✅ No paywalling of Scryfall data
✅ Respectful retry behavior

## API Coverage

- ✅ `GET /cards/:id` - Get card by Scryfall ID
- ✅ `GET /cards/arena/:id` - Get card by Arena ID
- ✅ `GET /sets/:code` - Get set information
- ✅ `GET /sets` - List all sets
- ✅ `GET /cards/search` - Search cards
- ✅ `GET /bulk-data` - Get bulk data downloads

## Data Models

All Scryfall response types are fully typed in `models.go`:

- `Card` - Full card object with all fields
- `Set` - Set information
- `SearchResult` - Paginated search results
- `BulkData` - Bulk data download information
- `CardFace` - Individual face of multi-faced cards
- `ImageURIs` - Card image URLs in various sizes
- `Legalities` - Legality by format
- `Prices` - Card prices

## References

- [Scryfall API Documentation](https://scryfall.com/docs/api)
- [Scryfall API Terms of Service](https://scryfall.com/docs/api#terms)
- Research: `docs/research/17lands-vs-scryfall-comparison.md`

## License

This project uses the Scryfall API in accordance with their Terms of Service.
All Scryfall data remains freely accessible and is not paywalled.

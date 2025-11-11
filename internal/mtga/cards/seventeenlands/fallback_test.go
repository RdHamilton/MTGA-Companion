package seventeenlands

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockCache implements CacheStorage for testing.
type mockCache struct {
	cardRatings  []CardRating
	colorRatings []ColorRating
	cachedAt     time.Time
	saveErr      error
	getErr       error
	saveCalled   bool
	getCalled    bool
}

func (m *mockCache) SaveCardRatings(ctx context.Context, ratings []CardRating, expansion, format, colors, startDate, endDate string) error {
	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}
	m.cardRatings = ratings
	m.cachedAt = time.Now()
	return nil
}

func (m *mockCache) GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]CardRating, time.Time, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, time.Time{}, m.getErr
	}
	return m.cardRatings, m.cachedAt, nil
}

func (m *mockCache) SaveColorRatings(ctx context.Context, ratings []ColorRating, expansion, eventType, startDate, endDate string) error {
	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}
	m.colorRatings = ratings
	m.cachedAt = time.Now()
	return nil
}

func (m *mockCache) GetColorRatings(ctx context.Context, expansion, eventType string) ([]ColorRating, time.Time, error) {
	m.getCalled = true
	if m.getErr != nil {
		return nil, time.Time{}, m.getErr
	}
	return m.colorRatings, m.cachedAt, nil
}

// TestGetCardRatings_APISuccess tests successful API response with caching.
func TestGetCardRatings_APISuccess(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"name": "Test Card",
				"mtga_id": 12345,
				"ever_drawn_win_rate": 58.5
			}
		]`))
	}))
	defer server.Close()

	// Create client with cache
	cache := &mockCache{}
	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
		Cache:     cache,
	}
	client := NewClient(options)

	// Override API base for testing
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(ratings) != 1 {
		t.Fatalf("Expected 1 rating, got %d", len(ratings))
	}
	if ratings[0].MTGAID != 12345 {
		t.Errorf("Expected MTGAID 12345, got %d", ratings[0].MTGAID)
	}

	// Verify cache was called to save
	if !cache.saveCalled {
		t.Error("Expected cache.SaveCardRatings to be called")
	}
	if cache.getCalled {
		t.Error("Expected cache.GetCardRatingsForSet NOT to be called on success")
	}
}

// TestGetCardRatings_APIFailure_CacheHit tests fallback to cache on API failure.
func TestGetCardRatings_APIFailure_CacheHit(t *testing.T) {
	// Setup mock server that returns 503
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Create client with cache that has data
	cachedRatings := []CardRating{
		{
			Name:   "Cached Card",
			MTGAID: 99999,
			GIHWR:  60.0,
		},
	}
	cache := &mockCache{
		cardRatings: cachedRatings,
		cachedAt:    time.Now().Add(-1 * time.Hour),
	}

	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
		Cache:     cache,
	}
	client := NewClient(options)

	// Override API base for testing
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Reset backoff so we don't have to wait
	client.ResetBackoff()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Verify
	if err != nil {
		t.Fatalf("Expected no error (cache fallback), got: %v", err)
	}
	if len(ratings) != 1 {
		t.Fatalf("Expected 1 cached rating, got %d", len(ratings))
	}
	if ratings[0].MTGAID != 99999 {
		t.Errorf("Expected cached MTGAID 99999, got %d", ratings[0].MTGAID)
	}

	// Verify cache was checked
	if !cache.getCalled {
		t.Error("Expected cache.GetCardRatingsForSet to be called on API failure")
	}

	// Verify stats
	stats := client.GetStats()
	if stats.CachedResponses != 1 {
		t.Errorf("Expected 1 cached response, got %d", stats.CachedResponses)
	}
}

// TestGetCardRatings_APIFailure_CacheMiss tests error when both API and cache fail.
func TestGetCardRatings_APIFailure_CacheMiss(t *testing.T) {
	// Setup mock server that returns 503
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Create client with empty cache
	cache := &mockCache{
		cardRatings: []CardRating{}, // Empty cache
	}

	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
		Cache:     cache,
	}
	client := NewClient(options)

	// Override API base for testing
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Reset backoff
	client.ResetBackoff()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Verify
	if err == nil {
		t.Fatal("Expected error when both API and cache fail")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.Type != ErrStatsUnavailable {
		t.Errorf("Expected error type %s, got %s", ErrStatsUnavailable, apiErr.Type)
	}

	if ratings != nil {
		t.Error("Expected nil ratings on complete failure")
	}

	// Verify cache was checked
	if !cache.getCalled {
		t.Error("Expected cache.GetCardRatingsForSet to be called")
	}
}

// TestGetCardRatings_NoCache tests behavior without cache.
func TestGetCardRatings_NoCache(t *testing.T) {
	// Setup mock server that returns 503
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Create client WITHOUT cache
	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
		Cache:     nil, // No cache
	}
	client := NewClient(options)

	// Override API base
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Reset backoff
	client.ResetBackoff()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Verify
	if err == nil {
		t.Fatal("Expected error when API fails and no cache")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.Type != ErrStatsUnavailable {
		t.Errorf("Expected error type %s, got %s", ErrStatsUnavailable, apiErr.Type)
	}

	if ratings != nil {
		t.Error("Expected nil ratings")
	}
}

// TestGetCardRatings_NetworkTimeout tests timeout handling.
func TestGetCardRatings_NetworkTimeout(t *testing.T) {
	// Setup slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Longer than timeout
	}))
	defer server.Close()

	// Create client with short timeout and cache
	cache := &mockCache{
		cardRatings: []CardRating{{MTGAID: 11111}},
		cachedAt:    time.Now(),
	}

	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   100 * time.Millisecond, // Short timeout
		Cache:     cache,
	}
	client := NewClient(options)

	// Override API base
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Should fall back to cache
	if err != nil {
		t.Fatalf("Expected cache fallback, got error: %v", err)
	}
	if len(ratings) != 1 {
		t.Fatalf("Expected 1 cached rating, got %d", len(ratings))
	}
}

// TestGetColorRatings_Fallback tests color ratings fallback.
func TestGetColorRatings_Fallback(t *testing.T) {
	// Setup mock server that fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with cache
	cache := &mockCache{
		colorRatings: []ColorRating{
			{ColorName: "WU", WinRate: 55.5},
		},
		cachedAt: time.Now().Add(-2 * time.Hour),
	}

	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
		Cache:     cache,
	}
	client := NewClient(options)

	// Override API base
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Reset backoff
	client.ResetBackoff()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetColorRatings(ctx, QueryParams{
		Expansion: "BLB",
		EventType: "PremierDraft",
	})

	// Verify fallback worked
	if err != nil {
		t.Fatalf("Expected cache fallback, got error: %v", err)
	}
	if len(ratings) != 1 {
		t.Fatalf("Expected 1 cached rating, got %d", len(ratings))
	}
	if ratings[0].ColorName != "WU" {
		t.Errorf("Expected cached color WU, got %s", ratings[0].ColorName)
	}

	// Verify stats
	stats := client.GetStats()
	if stats.CachedResponses != 1 {
		t.Errorf("Expected 1 cached response, got %d", stats.CachedResponses)
	}
}

// TestCacheSaveFailure tests that API success still works even if caching fails.
func TestCacheSaveFailure(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"name": "Test", "mtga_id": 123}]`))
	}))
	defer server.Close()

	// Create client with cache that fails to save
	cache := &mockCache{
		saveErr: errors.New("cache save error"),
	}

	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
		Cache:     cache,
	}
	client := NewClient(options)

	// Override API base
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Make request
	ctx := context.Background()
	ratings, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Should still succeed even though caching failed
	if err != nil {
		t.Fatalf("Expected success despite cache save failure, got: %v", err)
	}
	if len(ratings) != 1 {
		t.Fatalf("Expected 1 rating, got %d", len(ratings))
	}
}

// TestInvalidJSON tests handling of malformed API responses.
func TestInvalidJSON(t *testing.T) {
	// Setup mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	// Create client
	options := ClientOptions{
		RateLimit: DefaultRateLimit,
		Timeout:   5 * time.Second,
	}
	client := NewClient(options)

	// Override API base
	originalBase := APIBase
	APIBase = server.URL
	defer func() { APIBase = originalBase }()

	// Make request
	ctx := context.Background()
	_, err := client.GetCardRatings(ctx, QueryParams{
		Expansion: "BLB",
		Format:    "PremierDraft",
	})

	// Should get parse error
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}
	if apiErr.Type != ErrParseError {
		t.Errorf("Expected error type %s, got %s", ErrParseError, apiErr.Type)
	}
}

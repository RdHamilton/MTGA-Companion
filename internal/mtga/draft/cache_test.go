package draft

import (
	"testing"
	"time"
)

func TestNewCardRatingsCache(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		maxSize  int
		enabled  bool
		wantTTL  time.Duration
		wantSize int
	}{
		{
			name:     "cache with TTL and max size",
			ttl:      1 * time.Hour,
			maxSize:  100,
			enabled:  true,
			wantTTL:  1 * time.Hour,
			wantSize: 100,
		},
		{
			name:     "cache with no expiration",
			ttl:      0,
			maxSize:  0,
			enabled:  true,
			wantTTL:  0,
			wantSize: 0,
		},
		{
			name:     "disabled cache",
			ttl:      1 * time.Hour,
			maxSize:  100,
			enabled:  false,
			wantTTL:  1 * time.Hour,
			wantSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCardRatingsCache(tt.ttl, tt.maxSize, tt.enabled)

			if cache == nil {
				t.Fatal("NewCardRatingsCache() returned nil")
			}

			if cache.ttl != tt.wantTTL {
				t.Errorf("ttl = %v, want %v", cache.ttl, tt.wantTTL)
			}

			if cache.maxSize != tt.wantSize {
				t.Errorf("maxSize = %d, want %d", cache.maxSize, tt.wantSize)
			}

			if cache.enabled != tt.enabled {
				t.Errorf("enabled = %v, want %v", cache.enabled, tt.enabled)
			}

			if cache.entries == nil {
				t.Error("entries map not initialized")
			}
		})
	}
}

func TestCardRatingsCache_GetSet(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true) // No expiration, unlimited size

	rating := &CardRating{
		CardID:        100,
		Name:          "Test Card",
		GIHWR:         60.0,
		BayesianGIHWR: 60.0,
	}

	// Initially should return nil
	result := cache.Get(100, "ALL")
	if result != nil {
		t.Error("Get() on empty cache should return nil")
	}

	// Set and retrieve
	cache.Set(100, "ALL", rating)
	result = cache.Get(100, "ALL")
	if result == nil {
		t.Fatal("Get() after Set() returned nil")
	}

	if result.CardID != rating.CardID {
		t.Errorf("CardID = %d, want %d", result.CardID, rating.CardID)
	}

	if result.Name != rating.Name {
		t.Errorf("Name = %s, want %s", result.Name, rating.Name)
	}
}

func TestCardRatingsCache_ColorFilter(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true)

	ratingAll := &CardRating{
		CardID:        100,
		Name:          "Test Card",
		GIHWR:         60.0,
		BayesianGIHWR: 60.0,
	}

	ratingBR := &CardRating{
		CardID:        100,
		Name:          "Test Card",
		GIHWR:         65.0,
		BayesianGIHWR: 65.0,
	}

	// Set different ratings for different color filters
	cache.Set(100, "ALL", ratingAll)
	cache.Set(100, "BR", ratingBR)

	// Verify they're stored separately
	resultAll := cache.Get(100, "ALL")
	resultBR := cache.Get(100, "BR")

	if resultAll == nil || resultBR == nil {
		t.Fatal("Get() returned nil")
	}

	if resultAll.GIHWR != 60.0 {
		t.Errorf("ALL filter GIHWR = %.1f, want 60.0", resultAll.GIHWR)
	}

	if resultBR.GIHWR != 65.0 {
		t.Errorf("BR filter GIHWR = %.1f, want 65.0", resultBR.GIHWR)
	}
}

func TestCardRatingsCache_TTLExpiration(t *testing.T) {
	cache := NewCardRatingsCache(50*time.Millisecond, 0, true)

	rating := &CardRating{
		CardID:        100,
		Name:          "Test Card",
		GIHWR:         60.0,
		BayesianGIHWR: 60.0,
	}

	cache.Set(100, "ALL", rating)

	// Should be available immediately
	result := cache.Get(100, "ALL")
	if result == nil {
		t.Error("Get() immediately after Set() returned nil")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should be expired now
	result = cache.Get(100, "ALL")
	if result != nil {
		t.Error("Get() after TTL should return nil")
	}

	// Stats should show a miss
	stats := cache.GetStats()
	if stats.Misses == 0 {
		t.Error("Stats should show a miss for expired entry")
	}
}

func TestCardRatingsCache_MaxSizeEviction(t *testing.T) {
	cache := NewCardRatingsCache(0, 2, true) // Max 2 entries

	rating1 := &CardRating{CardID: 1, Name: "Card 1"}
	rating2 := &CardRating{CardID: 2, Name: "Card 2"}
	rating3 := &CardRating{CardID: 3, Name: "Card 3"}

	// Add 2 entries
	cache.Set(1, "ALL", rating1)
	cache.Set(2, "ALL", rating2)

	stats := cache.GetStats()
	if stats.Size != 2 {
		t.Errorf("Size = %d, want 2", stats.Size)
	}

	// Adding 3rd should evict oldest (FIFO)
	cache.Set(3, "ALL", rating3)

	stats = cache.GetStats()
	if stats.Size != 2 {
		t.Errorf("Size = %d, want 2 after eviction", stats.Size)
	}

	if stats.Evictions != 1 {
		t.Errorf("Evictions = %d, want 1", stats.Evictions)
	}

	// First entry should be evicted
	if cache.Get(1, "ALL") != nil {
		t.Error("First entry should be evicted")
	}

	// Second and third should still be present
	if cache.Get(2, "ALL") == nil {
		t.Error("Second entry should still be present")
	}
	if cache.Get(3, "ALL") == nil {
		t.Error("Third entry should still be present")
	}
}

func TestCardRatingsCache_Statistics(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true)

	rating := &CardRating{CardID: 100, Name: "Test Card"}
	cache.Set(100, "ALL", rating)

	// First get - miss (stats updated in Get)
	cache.Get(999, "ALL")

	// Second get - hit
	cache.Get(100, "ALL")

	// Third get - hit
	cache.Get(100, "ALL")

	stats := cache.GetStats()
	if stats.Hits != 2 {
		t.Errorf("Hits = %d, want 2", stats.Hits)
	}

	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}

	hitRate := cache.GetHitRate()
	expectedRate := 2.0 / 3.0 * 100.0 // 66.67%
	if hitRate < expectedRate-0.1 || hitRate > expectedRate+0.1 {
		t.Errorf("HitRate = %.2f%%, want %.2f%%", hitRate, expectedRate)
	}
}

func TestCardRatingsCache_Clear(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true)

	rating1 := &CardRating{CardID: 1, Name: "Card 1"}
	rating2 := &CardRating{CardID: 2, Name: "Card 2"}

	cache.Set(1, "ALL", rating1)
	cache.Set(2, "ALL", rating2)

	stats := cache.GetStats()
	if stats.Size != 2 {
		t.Errorf("Size = %d, want 2", stats.Size)
	}

	cache.Clear()

	stats = cache.GetStats()
	if stats.Size != 0 {
		t.Errorf("Size after Clear() = %d, want 0", stats.Size)
	}

	if cache.Get(1, "ALL") != nil {
		t.Error("Get() after Clear() should return nil")
	}
}

func TestCardRatingsCache_EnableDisable(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, false) // Start disabled

	rating := &CardRating{CardID: 100, Name: "Test Card"}

	// Set while disabled should do nothing
	cache.Set(100, "ALL", rating)
	if cache.Get(100, "ALL") != nil {
		t.Error("Disabled cache should not store entries")
	}

	// Enable and try again
	cache.Enable()
	cache.Set(100, "ALL", rating)
	if cache.Get(100, "ALL") == nil {
		t.Error("Enabled cache should store entries")
	}

	// Disable should clear
	cache.Disable()
	if cache.Get(100, "ALL") != nil {
		t.Error("Disable() should clear cache")
	}

	stats := cache.GetStats()
	if stats.Size != 0 {
		t.Errorf("Size after Disable() = %d, want 0", stats.Size)
	}
}

func TestCardRatingsCache_IsEnabled(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true)
	if !cache.IsEnabled() {
		t.Error("IsEnabled() = false, want true")
	}

	cache.Disable()
	if cache.IsEnabled() {
		t.Error("IsEnabled() after Disable() = true, want false")
	}

	cache.Enable()
	if !cache.IsEnabled() {
		t.Error("IsEnabled() after Enable() = false, want true")
	}
}

func TestCardRatingsCache_SetNilRating(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true)

	// Set with nil rating should do nothing
	cache.Set(100, "ALL", nil)

	stats := cache.GetStats()
	if stats.Size != 0 {
		t.Error("Set(nil) should not add entry to cache")
	}
}

func TestRatingsProvider_WithCache(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	cache := NewCardRatingsCache(0, 0, true)
	rp := NewRatingsProvider(setFile, config, cache)

	// First call - cache miss
	rating1, err := rp.GetCardRating(100, "ALL")
	if err != nil {
		t.Fatalf("GetCardRating() error = %v", err)
	}

	stats := cache.GetStats()
	if stats.Misses != 1 {
		t.Errorf("First call should be a cache miss, got %d misses", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("Cache size = %d, want 1", stats.Size)
	}

	// Second call - cache hit
	rating2, err := rp.GetCardRating(100, "ALL")
	if err != nil {
		t.Fatalf("GetCardRating() error = %v", err)
	}

	stats = cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Second call should be a cache hit, got %d hits", stats.Hits)
	}

	// Verify same data returned
	if rating1.CardID != rating2.CardID {
		t.Error("Cached rating doesn't match original")
	}
	if rating1.GIHWR != rating2.GIHWR {
		t.Error("Cached GIHWR doesn't match original")
	}
}

func TestRatingsProvider_WithoutCache(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil) // No cache

	// Should work normally without cache
	rating, err := rp.GetCardRating(100, "ALL")
	if err != nil {
		t.Fatalf("GetCardRating() error = %v", err)
	}

	if rating == nil {
		t.Fatal("GetCardRating() without cache returned nil")
	}

	if rating.CardID != 100 {
		t.Errorf("CardID = %d, want 100", rating.CardID)
	}
}

func TestRatingsProvider_CachePerformance(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	cache := NewCardRatingsCache(0, 0, true)
	rp := NewRatingsProvider(setFile, config, cache)

	// Simulate multiple lookups of the same cards
	cardIDs := []int{100, 200, 300}
	colorFilters := []string{"ALL", "B", "BR"}

	// First pass - all misses
	for _, cardID := range cardIDs {
		for _, colorFilter := range colorFilters {
			_, err := rp.GetCardRating(cardID, colorFilter)
			if err != nil && cardID != 300 { // 300 doesn't have BR
				t.Fatalf("GetCardRating(%d, %s) error = %v", cardID, colorFilter, err)
			}
		}
	}

	stats := cache.GetStats()
	// Card 300 doesn't have BR rating, but GetCardRating falls back to ALL
	// So we get: 100(ALL,B,BR) + 200(ALL,B,BR) + 300(ALL,W,BR-fallback) = 9 successful lookups
	expectedCached := 9
	if stats.Size != expectedCached {
		t.Errorf("Cache size after first pass = %d, want %d", stats.Size, expectedCached)
	}

	// Second pass - all hits
	for _, cardID := range cardIDs {
		for _, colorFilter := range colorFilters {
			_, err := rp.GetCardRating(cardID, colorFilter)
			if err != nil && cardID != 300 {
				t.Fatalf("GetCardRating(%d, %s) error = %v", cardID, colorFilter, err)
			}
		}
	}

	stats = cache.GetStats()
	if stats.Hits < int64(expectedCached) {
		t.Errorf("Cache hits = %d, want at least %d", stats.Hits, expectedCached)
	}

	hitRate := cache.GetHitRate()
	if hitRate < 40.0 { // Should be around 50% (8 misses, 8 hits)
		t.Errorf("Hit rate = %.2f%%, want at least 40%%", hitRate)
	}
}

func TestCardRatingsCache_MakeKey(t *testing.T) {
	cache := NewCardRatingsCache(0, 0, true)

	tests := []struct {
		cardID      int
		colorFilter string
		wantKey     string
	}{
		{100, "ALL", "100_ALL"},
		{200, "BR", "200_BR"},
		{999, "W", "999_W"},
	}

	for _, tt := range tests {
		t.Run(tt.wantKey, func(t *testing.T) {
			key := cache.makeKey(tt.cardID, tt.colorFilter)
			if key != tt.wantKey {
				t.Errorf("makeKey(%d, %s) = %s, want %s", tt.cardID, tt.colorFilter, key, tt.wantKey)
			}
		})
	}
}

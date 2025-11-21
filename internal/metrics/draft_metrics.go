package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// DraftMetrics tracks performance metrics for draft overlay operations.
type DraftMetrics struct {
	// Latency histograms (in milliseconds)
	ParseLatency    *Histogram
	RatingsLatency  *Histogram
	UIUpdateLatency *Histogram
	EndToEndLatency *Histogram

	// Counters (atomic operations for thread safety)
	EventsProcessed atomic.Uint64
	PacksRated      atomic.Uint64
	APIRequests     atomic.Uint64
	APIErrors       atomic.Uint64
	CacheHits       atomic.Uint64
	CacheMisses     atomic.Uint64

	// Start time for uptime calculation
	startTime time.Time
	mu        sync.RWMutex
}

// NewDraftMetrics creates a new metrics collector.
func NewDraftMetrics() *DraftMetrics {
	return &DraftMetrics{
		ParseLatency:    NewHistogram(10000),
		RatingsLatency:  NewHistogram(10000),
		UIUpdateLatency: NewHistogram(10000),
		EndToEndLatency: NewHistogram(10000),
		startTime:       time.Now(),
	}
}

// RecordParseDuration records the time taken to parse a log entry.
func (m *DraftMetrics) RecordParseDuration(d time.Duration) {
	m.ParseLatency.Record(d)
}

// RecordRatingsDuration records the time taken to fetch card ratings.
func (m *DraftMetrics) RecordRatingsDuration(d time.Duration) {
	m.RatingsLatency.Record(d)
}

// RecordUIUpdateDuration records the time taken to update the UI.
func (m *DraftMetrics) RecordUIUpdateDuration(d time.Duration) {
	m.UIUpdateLatency.Record(d)
}

// RecordEndToEndDuration records the total time from log event to UI display.
func (m *DraftMetrics) RecordEndToEndDuration(d time.Duration) {
	m.EndToEndLatency.Record(d)
}

// IncrementEventsProcessed increments the count of events processed.
func (m *DraftMetrics) IncrementEventsProcessed() {
	m.EventsProcessed.Add(1)
}

// IncrementPacksRated increments the count of packs rated.
func (m *DraftMetrics) IncrementPacksRated() {
	m.PacksRated.Add(1)
}

// IncrementAPIRequests increments the count of API requests.
func (m *DraftMetrics) IncrementAPIRequests() {
	m.APIRequests.Add(1)
}

// IncrementAPIErrors increments the count of API errors.
func (m *DraftMetrics) IncrementAPIErrors() {
	m.APIErrors.Add(1)
}

// IncrementCacheHits increments the count of cache hits.
func (m *DraftMetrics) IncrementCacheHits() {
	m.CacheHits.Add(1)
}

// IncrementCacheMisses increments the count of cache misses.
func (m *DraftMetrics) IncrementCacheMisses() {
	m.CacheMisses.Add(1)
}

// DraftStats contains the computed statistics from metrics.
type DraftStats struct {
	// Latency statistics (milliseconds)
	ParseLatency    LatencyStats `json:"parse_latency"`
	RatingsLatency  LatencyStats `json:"ratings_latency"`
	UIUpdateLatency LatencyStats `json:"ui_update_latency"`
	EndToEndLatency LatencyStats `json:"end_to_end_latency"`

	// Counters
	EventsProcessed uint64  `json:"events_processed"`
	PacksRated      uint64  `json:"packs_rated"`
	APIRequests     uint64  `json:"api_requests"`
	APIErrors       uint64  `json:"api_errors"`
	CacheHits       uint64  `json:"cache_hits"`
	CacheMisses     uint64  `json:"cache_misses"`
	CacheHitRate    float64 `json:"cache_hit_rate"`   // percentage
	APISuccessRate  float64 `json:"api_success_rate"` // percentage

	// System info
	Uptime string `json:"uptime"` // human-readable uptime
}

// LatencyStats contains statistics for a latency histogram.
type LatencyStats struct {
	Mean  float64 `json:"mean"`  // milliseconds
	P50   float64 `json:"p50"`   // median
	P95   float64 `json:"p95"`   // 95th percentile
	P99   float64 `json:"p99"`   // 99th percentile
	Min   float64 `json:"min"`   // minimum
	Max   float64 `json:"max"`   // maximum
	Count int     `json:"count"` // number of samples
}

// GetStats returns a snapshot of the current statistics.
func (m *DraftMetrics) GetStats() *DraftStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get counter values
	eventsProcessed := m.EventsProcessed.Load()
	packsRated := m.PacksRated.Load()
	apiRequests := m.APIRequests.Load()
	apiErrors := m.APIErrors.Load()
	cacheHits := m.CacheHits.Load()
	cacheMisses := m.CacheMisses.Load()

	// Calculate rates
	cacheHitRate := 0.0
	if cacheHits+cacheMisses > 0 {
		cacheHitRate = (float64(cacheHits) / float64(cacheHits+cacheMisses)) * 100
	}

	apiSuccessRate := 0.0
	if apiRequests > 0 {
		apiSuccessRate = (float64(apiRequests-apiErrors) / float64(apiRequests)) * 100
	}

	// Calculate uptime
	uptime := time.Since(m.startTime).Round(time.Second).String()

	return &DraftStats{
		ParseLatency:    m.getLatencyStats(m.ParseLatency),
		RatingsLatency:  m.getLatencyStats(m.RatingsLatency),
		UIUpdateLatency: m.getLatencyStats(m.UIUpdateLatency),
		EndToEndLatency: m.getLatencyStats(m.EndToEndLatency),
		EventsProcessed: eventsProcessed,
		PacksRated:      packsRated,
		APIRequests:     apiRequests,
		APIErrors:       apiErrors,
		CacheHits:       cacheHits,
		CacheMisses:     cacheMisses,
		CacheHitRate:    cacheHitRate,
		APISuccessRate:  apiSuccessRate,
		Uptime:          uptime,
	}
}

// getLatencyStats extracts latency statistics from a histogram.
func (m *DraftMetrics) getLatencyStats(h *Histogram) LatencyStats {
	return LatencyStats{
		Mean:  h.Mean(),
		P50:   h.Percentile(50),
		P95:   h.Percentile(95),
		P99:   h.Percentile(99),
		Min:   h.Min(),
		Max:   h.Max(),
		Count: h.Count(),
	}
}

// Reset clears all metrics.
func (m *DraftMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ParseLatency.Reset()
	m.RatingsLatency.Reset()
	m.UIUpdateLatency.Reset()
	m.EndToEndLatency.Reset()

	m.EventsProcessed.Store(0)
	m.PacksRated.Store(0)
	m.APIRequests.Store(0)
	m.APIErrors.Store(0)
	m.CacheHits.Store(0)
	m.CacheMisses.Store(0)

	m.startTime = time.Now()
}

package metrics

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Histogram tracks a distribution of duration values and calculates percentiles.
type Histogram struct {
	samples []float64 // duration in milliseconds
	mu      sync.RWMutex
	maxSize int // maximum number of samples to keep
}

// NewHistogram creates a new histogram with a maximum sample size.
// When maxSize is exceeded, oldest samples are removed.
func NewHistogram(maxSize int) *Histogram {
	if maxSize <= 0 {
		maxSize = 10000 // default to 10k samples
	}
	return &Histogram{
		samples: make([]float64, 0, maxSize),
		maxSize: maxSize,
	}
}

// Record adds a duration sample to the histogram.
func (h *Histogram) Record(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ms := float64(d.Microseconds()) / 1000.0 // convert to milliseconds

	// Add sample
	h.samples = append(h.samples, ms)

	// Trim if exceeded maxSize
	if len(h.samples) > h.maxSize {
		// Remove oldest 20% of samples to avoid constant trimming
		removeCount := h.maxSize / 5
		h.samples = h.samples[removeCount:]
	}
}

// Mean returns the average duration in milliseconds.
func (h *Histogram) Mean() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.samples) == 0 {
		return 0
	}

	var sum float64
	for _, v := range h.samples {
		sum += v
	}
	return sum / float64(len(h.samples))
}

// Percentile returns the value at the given percentile (0-100).
// For example, Percentile(95) returns the 95th percentile (p95).
func (h *Histogram) Percentile(p float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.samples) == 0 {
		return 0
	}

	// Make a copy and sort
	sorted := make([]float64, len(h.samples))
	copy(sorted, h.samples)
	sort.Float64s(sorted)

	// Calculate index
	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation between lower and upper
	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

// Min returns the minimum value.
func (h *Histogram) Min() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.samples) == 0 {
		return 0
	}

	min := h.samples[0]
	for _, v := range h.samples[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Max returns the maximum value.
func (h *Histogram) Max() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.samples) == 0 {
		return 0
	}

	max := h.samples[0]
	for _, v := range h.samples[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// Count returns the total number of samples.
func (h *Histogram) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.samples)
}

// Reset clears all samples.
func (h *Histogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.samples = h.samples[:0]
}

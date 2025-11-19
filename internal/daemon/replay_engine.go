package daemon

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// ReplayEngine simulates real-time log replay by streaming historical log entries
// with realistic timing delays. This enables cost-effective testing of draft/event features
// without requiring actual gameplay.
type ReplayEngine struct {
	service *Service
	speed   float64 // 1.0 = real-time, 2.0 = 2x speed, etc.

	// Replay state
	mu          sync.RWMutex
	isActive    bool
	isPaused    bool
	entries     []*logreader.LogEntry
	currentIdx  int
	startTime   time.Time
	pauseTime   time.Time
	totalPaused time.Duration
	filterType  string // "all", "draft", "match", "event"

	// Control channels
	ctx        context.Context
	cancel     context.CancelFunc
	pauseChan  chan bool
	resumeChan chan bool
	stopChan   chan bool
}

// NewReplayEngine creates a new replay engine.
func NewReplayEngine(service *Service) *ReplayEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &ReplayEngine{
		service:    service,
		speed:      1.0,
		filterType: "all",
		ctx:        ctx,
		cancel:     cancel,
		pauseChan:  make(chan bool, 1),
		resumeChan: make(chan bool, 1),
		stopChan:   make(chan bool, 1),
	}
}

// Start begins replay of a log file with the specified speed and filter.
// Returns error if replay is already active or if log file cannot be read.
func (r *ReplayEngine) Start(logPath string, speed float64, filterType string) error {
	r.mu.Lock()
	if r.isActive {
		r.mu.Unlock()
		return fmt.Errorf("replay already active")
	}
	r.isActive = true
	r.speed = speed
	r.filterType = filterType
	r.currentIdx = 0
	r.startTime = time.Now()
	r.totalPaused = 0
	r.mu.Unlock()

	// Read log file
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		r.mu.Lock()
		r.isActive = false
		r.mu.Unlock()
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	entries, err := reader.ReadAll()
	if err != nil {
		r.mu.Lock()
		r.isActive = false
		r.mu.Unlock()
		return fmt.Errorf("failed to read log file: %w", err)
	}

	if len(entries) == 0 {
		r.mu.Lock()
		r.isActive = false
		r.mu.Unlock()
		return fmt.Errorf("log file contains no entries")
	}

	// Filter entries if needed
	if filterType != "all" {
		entries = r.filterEntries(entries, filterType)
		if len(entries) == 0 {
			r.mu.Lock()
			r.isActive = false
			r.mu.Unlock()
			return fmt.Errorf("no entries match filter: %s", filterType)
		}
	}

	r.mu.Lock()
	r.entries = entries
	r.mu.Unlock()

	log.Printf("Starting replay: %d entries, %.1fx speed, filter: %s", len(entries), speed, filterType)

	// Enable dry run mode to prevent database pollution
	r.service.logProcessor.SetDryRun(true)
	log.Println("⚠️  REPLAY MODE: Data will be parsed and broadcast but NOT stored to database")

	// Broadcast replay started event
	r.service.wsServer.Broadcast(Event{
		Type: "replay:started",
		Data: map[string]interface{}{
			"totalEntries": len(entries),
			"speed":        speed,
			"filter":       filterType,
		},
	})

	// Stream entries in background goroutine
	go r.streamEntries()

	return nil
}

// streamEntries streams log entries with realistic timing delays.
func (r *ReplayEngine) streamEntries() {
	defer func() {
		r.mu.Lock()
		r.isActive = false
		r.isPaused = false
		r.mu.Unlock()

		// Disable dry run mode - return to normal operation
		r.service.logProcessor.SetDryRun(false)

		// Broadcast replay completed event
		r.service.wsServer.Broadcast(Event{
			Type: "replay:completed",
			Data: map[string]interface{}{
				"totalEntries": len(r.entries),
				"elapsed":      time.Since(r.startTime) - r.totalPaused,
			},
		})

		log.Println("Replay completed")
	}()

	var prevEntry *logreader.LogEntry
	batchSize := 10 // Process entries in batches for efficiency
	batch := make([]*logreader.LogEntry, 0, batchSize)

	for i := 0; i < len(r.entries); i++ {
		// Check for stop signal
		select {
		case <-r.stopChan:
			log.Println("Replay stopped by user")
			return
		case <-r.ctx.Done():
			log.Println("Replay cancelled")
			return
		default:
		}

		// Check for pause signal
		r.mu.RLock()
		isPaused := r.isPaused
		r.mu.RUnlock()

		if isPaused {
			r.mu.Lock()
			r.pauseTime = time.Now()
			r.mu.Unlock()

			log.Println("Replay paused")
			r.service.wsServer.Broadcast(Event{
				Type: "replay:paused",
				Data: map[string]interface{}{
					"currentEntry": i,
					"totalEntries": len(r.entries),
				},
			})

			// Wait for resume signal
			<-r.resumeChan

			r.mu.Lock()
			r.totalPaused += time.Since(r.pauseTime)
			r.mu.Unlock()

			log.Println("Replay resumed")
			r.service.wsServer.Broadcast(Event{
				Type: "replay:resumed",
				Data: map[string]interface{}{
					"currentEntry": i,
					"totalEntries": len(r.entries),
				},
			})
		}

		entry := r.entries[i]

		// Calculate delay based on timestamps
		if prevEntry != nil {
			delay := r.calculateDelay(prevEntry, entry)
			adjustedDelay := time.Duration(float64(delay) / r.speed)

			if adjustedDelay > 0 {
				select {
				case <-time.After(adjustedDelay):
				case <-r.stopChan:
					return
				case <-r.ctx.Done():
					return
				}
			}
		}

		// Add to batch
		batch = append(batch, entry)

		// Process batch when full or at end
		if len(batch) >= batchSize || i == len(r.entries)-1 {
			r.service.processEntries(batch)
			batch = batch[:0] // Clear batch
		}

		prevEntry = entry
		r.mu.Lock()
		r.currentIdx = i + 1
		r.mu.Unlock()

		// Broadcast progress every 50 entries or at end
		if (i+1)%50 == 0 || i == len(r.entries)-1 {
			percentComplete := float64(i+1) / float64(len(r.entries)) * 100
			elapsed := time.Since(r.startTime) - r.totalPaused

			r.service.wsServer.Broadcast(Event{
				Type: "replay:progress",
				Data: map[string]interface{}{
					"currentEntry":    i + 1,
					"totalEntries":    len(r.entries),
					"percentComplete": percentComplete,
					"elapsed":         elapsed.Seconds(),
					"isActive":        true,
				},
			})
		}
	}
}

// calculateDelay calculates the delay between two log entries based on their timestamps.
// Returns 0 if timestamps cannot be parsed or if delay would be too long.
func (r *ReplayEngine) calculateDelay(prev, current *logreader.LogEntry) time.Duration {
	if prev == nil {
		return 0
	}

	// Try to extract timestamps from entries
	prevTime := r.extractTimestamp(prev)
	currTime := r.extractTimestamp(current)

	if prevTime.IsZero() || currTime.IsZero() {
		return 0
	}

	delay := currTime.Sub(prevTime)

	// Cap maximum delay to 5 seconds (don't wait minutes between entries)
	if delay > 5*time.Second {
		delay = 5 * time.Second
	}

	// Minimum delay of 10ms to prevent overwhelming the system
	if delay < 10*time.Millisecond {
		delay = 10 * time.Millisecond
	}

	return delay
}

// extractTimestamp attempts to extract a timestamp from a log entry.
// Returns zero time if no timestamp can be extracted.
func (r *ReplayEngine) extractTimestamp(entry *logreader.LogEntry) time.Time {
	// Log entries have format like: [UnityCrossThreadLogger]2024-11-16 15:30:45.123
	// The Timestamp field is already extracted by the reader
	if entry.Timestamp == "" {
		return time.Time{}
	}

	// Parse various timestamp formats
	formats := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"15:04:05.000",
		"15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, entry.Timestamp); err == nil {
			return t
		}
	}

	return time.Time{}
}

// filterEntries filters log entries based on the specified filter type.
func (r *ReplayEngine) filterEntries(entries []*logreader.LogEntry, filterType string) []*logreader.LogEntry {
	if filterType == "all" {
		return entries
	}

	filtered := make([]*logreader.LogEntry, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Check JSON content for relevant event types
		switch filterType {
		case "draft":
			if r.isDraftEntry(entry) {
				filtered = append(filtered, entry)
			}
		case "match":
			if r.isMatchEntry(entry) {
				filtered = append(filtered, entry)
			}
		case "event":
			if r.isEventEntry(entry) {
				filtered = append(filtered, entry)
			}
		}
	}

	return filtered
}

// isDraftEntry checks if a log entry is related to draft events.
func (r *ReplayEngine) isDraftEntry(entry *logreader.LogEntry) bool {
	raw := entry.Raw
	return strings.Contains(raw, "Draft.") ||
		strings.Contains(raw, "DraftPick") ||
		strings.Contains(raw, "DraftMakePick") ||
		strings.Contains(raw, "DraftStatus")
}

// isMatchEntry checks if a log entry is related to match events.
func (r *ReplayEngine) isMatchEntry(entry *logreader.LogEntry) bool {
	raw := entry.Raw
	return strings.Contains(raw, "MatchGameRoomStateChanged") ||
		strings.Contains(raw, "GameRoomStateChanged") ||
		strings.Contains(raw, "EventMatchCreated")
}

// isEventEntry checks if a log entry is related to event (tournament) events.
func (r *ReplayEngine) isEventEntry(entry *logreader.LogEntry) bool {
	raw := entry.Raw
	return strings.Contains(raw, "Event_") ||
		strings.Contains(raw, "EventJoin") ||
		strings.Contains(raw, "EventGetCourses")
}

// Pause pauses the replay.
func (r *ReplayEngine) Pause() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isActive {
		return fmt.Errorf("replay not active")
	}

	if r.isPaused {
		return fmt.Errorf("replay already paused")
	}

	r.isPaused = true
	select {
	case r.pauseChan <- true:
	default:
	}

	return nil
}

// Resume resumes the replay.
func (r *ReplayEngine) Resume() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isActive {
		return fmt.Errorf("replay not active")
	}

	if !r.isPaused {
		return fmt.Errorf("replay not paused")
	}

	r.isPaused = false
	select {
	case r.resumeChan <- true:
	default:
	}

	return nil
}

// Stop stops the replay.
func (r *ReplayEngine) Stop() error {
	r.mu.RLock()
	isActive := r.isActive
	r.mu.RUnlock()

	if !isActive {
		return fmt.Errorf("replay not active")
	}

	select {
	case r.stopChan <- true:
	default:
	}

	return nil
}

// GetStatus returns the current replay status.
func (r *ReplayEngine) GetStatus() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.isActive {
		return map[string]interface{}{
			"isActive": false,
		}
	}

	percentComplete := 0.0
	if len(r.entries) > 0 {
		percentComplete = float64(r.currentIdx) / float64(len(r.entries)) * 100
	}

	elapsed := time.Since(r.startTime) - r.totalPaused

	return map[string]interface{}{
		"isActive":        true,
		"isPaused":        r.isPaused,
		"currentEntry":    r.currentIdx,
		"totalEntries":    len(r.entries),
		"percentComplete": percentComplete,
		"elapsed":         elapsed.Seconds(),
		"speed":           r.speed,
		"filter":          r.filterType,
	}
}

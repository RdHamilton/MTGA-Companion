package draft

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// Overlay manages the real-time draft overlay system.
type Overlay struct {
	parser          *Parser
	ratingsProvider *RatingsProvider
	cache           *CardRatingsCache
	colorConfig     ColorAffinityConfig
	currentState    *DraftState
	currentRatings  *PackRatings
	selectedColors  []string
	logPath         string
	updateCallback  func(*OverlayUpdate)
	stopChan        chan struct{}
	resumeEnabled   bool
	lookbackHours   int
	logger          *Logger
	mu              sync.RWMutex
}

// OverlayConfig configures the draft overlay.
type OverlayConfig struct {
	LogPath        string
	SetFile        *seventeenlands.SetFile
	BayesianConfig BayesianConfig
	ColorConfig    ColorAffinityConfig
	UpdateCallback func(*OverlayUpdate)
	PollInterval   time.Duration // How often to check log for updates
	ResumeEnabled  bool          // Whether to scan log history for active draft
	LookbackHours  int           // How many hours back to scan (default: 24)
	CacheEnabled   bool          // Whether to enable in-memory caching
	CacheTTL       time.Duration // Cache time-to-live (default: 24 hours)
	CacheMaxSize   int           // Maximum cache entries (0 = unlimited)
	DebugMode      bool          // Enable verbose debug logging
}

// OverlayUpdate represents an update to send to the UI.
type OverlayUpdate struct {
	Type               UpdateType
	DraftState         *DraftState
	PackRatings        *PackRatings
	BestPick           *CardRating
	TopPicks           []*CardRating
	ColorSuggestion    *ColorSuggestion
	DeckRecommendation *DeckRecommendation
	Timestamp          time.Time
}

// UpdateType represents the type of overlay update.
type UpdateType string

const (
	UpdateTypeDraftStart  UpdateType = "draft_start"
	UpdateTypeNewPack     UpdateType = "new_pack"
	UpdateTypePickMade    UpdateType = "pick_made"
	UpdateTypeDraftEnd    UpdateType = "draft_end"
	UpdateTypeColorRec    UpdateType = "color_recommendation"
	UpdateTypeDeckBuilder UpdateType = "deck_builder"
)

// ColorSuggestion represents suggested colors for the draft.
type ColorSuggestion struct {
	SuggestedColors []string // e.g., ["B", "R"] or ["BR"]
	Reason          string   // Why these colors were suggested
	Affinities      map[string]*ColorAffinity
	RankedColors    []DeckColor
}

// NewOverlay creates a new draft overlay.
func NewOverlay(config OverlayConfig) *Overlay {
	parser := NewParser()
	logger := NewLogger(config.DebugMode)

	// Create cache if enabled
	var cache *CardRatingsCache
	if config.CacheEnabled {
		cacheTTL := config.CacheTTL
		if cacheTTL == 0 {
			cacheTTL = 24 * time.Hour // Default: 24 hours
		}
		cache = NewCardRatingsCache(cacheTTL, config.CacheMaxSize, true)
		logger.Info("Card ratings cache enabled (TTL: %v, MaxSize: %d)", cacheTTL, config.CacheMaxSize)
	}

	ratingsProvider := NewRatingsProvider(config.SetFile, config.BayesianConfig, cache)

	if config.PollInterval == 0 {
		config.PollInterval = 20 * time.Millisecond // Very fast polling for minimal latency
	}

	if config.LookbackHours == 0 {
		config.LookbackHours = 24 // Default to last 24 hours
	}

	return &Overlay{
		parser:          parser,
		ratingsProvider: ratingsProvider,
		cache:           cache,
		colorConfig:     config.ColorConfig,
		logPath:         config.LogPath,
		updateCallback:  config.UpdateCallback,
		stopChan:        make(chan struct{}),
		resumeEnabled:   config.ResumeEnabled,
		lookbackHours:   config.LookbackHours,
		logger:          logger,
	}
}

// Start begins monitoring the log file for draft events.
func (o *Overlay) Start(ctx context.Context) error {
	file, err := os.Open(o.logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// If resume is enabled, scan log history for active draft
	if o.resumeEnabled {
		if err := o.scanForActiveDraft(file); err != nil {
			// No active draft found - continue monitoring for new events
			o.logger.Info("No active draft found in log history. Waiting for new draft...")
		} else {
			o.logger.Info("Successfully resumed active draft!")
		}
	}

	// Seek to end of file to only process new entries
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek to end of log: %w", err)
	}

	o.logger.Info("Monitoring log file for draft events (using file system notifications)...")
	fmt.Println()

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() {
		if closeErr := watcher.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Watch the log file for changes
	if err := watcher.Add(o.logPath); err != nil {
		return fmt.Errorf("failed to watch log file: %w", err)
	}

	reader := bufio.NewReader(file)

	// Also keep a ticker as backup (in case file events are delayed)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-o.stopChan:
			return nil
		case event := <-watcher.Events:
			// File was modified - read new content immediately
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := o.processNewLogLines(reader); err != nil {
					continue
				}
			}
		case err := <-watcher.Errors:
			fmt.Printf("[WARN] File watcher error: %v\n", err)
		case <-ticker.C:
			// Backup polling in case file events are missed
			if err := o.processNewLogLines(reader); err != nil {
				continue
			}
		}
	}
}

// Stop stops the overlay monitoring.
func (o *Overlay) Stop() {
	close(o.stopChan)
}

// scanForActiveDraft scans the log file history for an active draft.
// Returns nil if active draft found and state restored, error otherwise.
func (o *Overlay) scanForActiveDraft(file *os.File) error {
	o.logger.Info("Scanning log history for active draft...")

	// Seek to beginning of file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to start of log: %w", err)
	}

	// Use bufio.Reader instead of Scanner to handle long lines
	reader := bufio.NewReader(file)
	lineNumber := 0
	draftStartFound := false

	// Scan through entire log file
	botDraftLinesFound := 0
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading log file: %w", err)
		}

		lineNumber++

		// Track BotDraft lines for debugging
		if strings.Contains(line, `"CurrentModule":"BotDraft"`) {
			botDraftLinesFound++
			if botDraftLinesFound <= 3 {
				preview := line
				if len(preview) > 150 {
					preview = preview[:150] + "..."
				}
				o.logger.Debug("Found BotDraft line during scan: %s", preview)
			}
		}

		// Parse log entry
		event, parseErr := o.parser.ParseLogEntry(line, time.Now())
		if parseErr != nil {
			continue
		}

		if event != nil {
			o.logger.Debug("Resume scan parsed event: %s", event.Type)
		}

		if event == nil {
			continue
		}

		// Track if we've found a draft start (not sealed)
		if IsPackEvent(event.Type) && event.Type != LogEventGrantCardPool && event.Type != LogEventCoursesCardPool {
			draftStartFound = true
			o.logger.Debug("Draft start detected during resume scan! Event type: %s", event.Type)
		}

		// Update parser state
		o.mu.Lock()
		if err := o.parser.UpdateState(event); err != nil {
			o.mu.Unlock()
			continue
		}
		o.mu.Unlock()

		// Continue parsing all events to build up complete draft state
	}

	// If we found a draft and have a current state, check if it's still active
	o.mu.Lock()
	defer o.mu.Unlock()

	o.currentState = o.parser.GetCurrentState()
	if o.currentState == nil || !draftStartFound {
		return fmt.Errorf("no active draft found")
	}

	// If draft is marked as in progress and we have a pack, resume it
	if o.currentState.Event.InProgress && o.currentState.CurrentPack != nil {
		o.logger.Debug("Found active draft! Pack %d, Pick %d, %d cards in pack, %d picks made",
			o.currentState.Event.CurrentPack,
			o.currentState.Event.CurrentPick,
			len(o.currentState.CurrentPack.CardIDs),
			len(o.currentState.Picks))

		// Send draft start update
		if o.updateCallback != nil {
			o.updateCallback(&OverlayUpdate{
				Type:       UpdateTypeDraftStart,
				DraftState: o.currentState,
				Timestamp:  time.Now(),
			})
		}

		// Trigger pack event to show current pack ratings
		o.handlePackEvent()

		return nil
	}

	return fmt.Errorf("draft found but not in progress")
}

// processNewLogLines reads and processes any new lines from the log.
func (o *Overlay) processNewLogLines(reader *bufio.Reader) error {
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// No more lines to read
			return nil
		}
		if err != nil {
			return err
		}

		// Process the log line
		o.processLogLine(line)
	}
}

// processLogLine parses a log line and updates the overlay state.
func (o *Overlay) processLogLine(line string) {
	// Parse timestamp (simplified - actual MTGA logs have timestamps)
	timestamp := time.Now()

	// Parse log entry
	event, err := o.parser.ParseLogEntry(line, timestamp)
	if err != nil {
		o.logger.Debug("Parse error: %v", err)
		return
	}
	if event == nil {
		return // Silently skip non-draft lines
	}

	// Skip Sealed events (they interfere with active drafts)
	if event.Type == LogEventGrantCardPool || event.Type == LogEventCoursesCardPool {
		o.logger.Debug("Skipping Sealed event: %s", event.Type)
		return
	}

	// Debug: print detected events with timestamp
	o.logger.Debug("Detected event: %s", event.Type)

	o.mu.Lock()
	defer o.mu.Unlock()

	// Update parser state
	if err := o.parser.UpdateState(event); err != nil {
		o.logger.Debug("Error updating state: %v", err)
		return
	}

	// Get current state
	o.currentState = o.parser.GetCurrentState()
	if o.currentState == nil {
		return
	}

	// Handle different event types
	switch {
	case IsPackEvent(event.Type):
		o.logger.Debug("Handling pack event - Pack %d, Pick %d",
			o.currentState.Event.CurrentPack, o.currentState.Event.CurrentPick)
		o.handlePackEvent()
	case IsPickEvent(event.Type):
		o.logger.Debug("Handling pick event")
		o.handlePickEvent()
	}
}

// handlePackEvent processes a new pack event.
func (o *Overlay) handlePackEvent() {
	if o.currentState.CurrentPack == nil {
		o.logger.Debug("No current pack in state")
		return
	}

	o.logger.Debug("Current pack has %d cards", len(o.currentState.CurrentPack.CardIDs))

	// Update color suggestion based on picks so far
	o.updateColorSuggestion()

	// Determine which color filter to use
	colorFilter := "ALL"
	if len(o.selectedColors) > 0 {
		colorFilter = strings.Join(o.selectedColors, "")
	}

	o.logger.Debug("Getting ratings with color filter: %s", colorFilter)

	// Get ratings for the pack
	packRatings, err := o.ratingsProvider.GetPackRatings(o.currentState.CurrentPack, colorFilter)
	if err != nil {
		o.logger.Debug("Error getting pack ratings: %v", err)
		return
	}

	o.logger.Debug("Got ratings for %d cards, sending to UI", len(packRatings.CardRatings))

	o.currentRatings = packRatings

	// Send update to UI
	if o.updateCallback != nil {
		update := &OverlayUpdate{
			Type:        UpdateTypeNewPack,
			DraftState:  o.currentState,
			PackRatings: packRatings,
			BestPick:    packRatings.GetBestPick(),
			TopPicks:    packRatings.TopN(5),
			Timestamp:   time.Now(),
		}

		// Add color suggestion if we have enough picks
		if len(o.currentState.Picks) >= o.colorConfig.MinCards {
			update.ColorSuggestion = o.getColorSuggestion()
		}

		o.updateCallback(update)
	}
}

// handlePickEvent processes a pick event.
func (o *Overlay) handlePickEvent() {
	// After a pick, update color suggestions
	o.updateColorSuggestion()

	if o.updateCallback != nil {
		update := &OverlayUpdate{
			Type:       UpdateTypePickMade,
			DraftState: o.currentState,
			Timestamp:  time.Now(),
		}

		// Add color suggestion if we have enough picks
		if len(o.currentState.Picks) >= o.colorConfig.MinCards {
			update.ColorSuggestion = o.getColorSuggestion()
		}

		o.updateCallback(update)
	}

	// Check if draft is complete (45 picks = 3 packs × 15 picks)
	if o.isDraftComplete() {
		o.logger.Info("Draft complete! Building deck recommendations...")
		o.handleDraftComplete()
	}
}

// updateColorSuggestion recalculates the color suggestion based on current picks.
func (o *Overlay) updateColorSuggestion() {
	if len(o.currentState.Picks) < o.colorConfig.MinCards {
		return
	}

	// Get card data for all picks
	pickedCards := make([]*seventeenlands.CardRatingData, 0)
	for _, pick := range o.currentState.Picks {
		rating, err := o.ratingsProvider.GetCardRating(pick.CardID, "ALL")
		if err != nil {
			continue
		}

		// Convert CardRating back to CardRatingData for color analysis
		// This is a simplified conversion - in practice you'd need the full card data
		cardData := &seventeenlands.CardRatingData{
			ArenaID:  rating.CardID,
			Name:     rating.Name,
			ManaCost: rating.ManaCost,
			CMC:      rating.CMC,
			Types:    rating.Types,
			Colors:   rating.Colors,
			Rarity:   rating.Rarity,
		}
		pickedCards = append(pickedCards, cardData)
	}

	// Calculate metrics for auto-selection
	metrics := CalculateDeckMetrics(pickedCards, "ALL")

	// Auto-select colors
	selectedColors := AutoSelectColors(pickedCards, o.colorConfig, metrics)
	if len(selectedColors) > 0 {
		o.selectedColors = selectedColors
	}
}

// getColorSuggestion gets the current color suggestion.
func (o *Overlay) getColorSuggestion() *ColorSuggestion {
	if len(o.currentState.Picks) < o.colorConfig.MinCards {
		return nil
	}

	// Get card data for all picks (simplified)
	pickedCards := make([]*seventeenlands.CardRatingData, 0)
	for _, pick := range o.currentState.Picks {
		rating, err := o.ratingsProvider.GetCardRating(pick.CardID, "ALL")
		if err != nil {
			continue
		}

		cardData := &seventeenlands.CardRatingData{
			ArenaID:  rating.CardID,
			Name:     rating.Name,
			ManaCost: rating.ManaCost,
			CMC:      rating.CMC,
			Types:    rating.Types,
			Colors:   rating.Colors,
			Rarity:   rating.Rarity,
		}
		pickedCards = append(pickedCards, cardData)
	}

	// Calculate metrics
	metrics := CalculateDeckMetrics(pickedCards, "ALL")
	threshold := metrics.Mean - (o.colorConfig.ThresholdStdDevFactor * metrics.StandardDeviation)

	// Get color affinities
	affinities := CalculateColorAffinity(pickedCards, "ALL", threshold)

	// Rank deck colors
	rankedColors := RankDeckColors(pickedCards, o.colorConfig, metrics)

	// Auto-select
	selectedColors := AutoSelectColors(pickedCards, o.colorConfig, metrics)

	reason := fmt.Sprintf("Based on %d picks, suggested colors have highest affinity", len(o.currentState.Picks))

	return &ColorSuggestion{
		SuggestedColors: selectedColors,
		Reason:          reason,
		Affinities:      affinities,
		RankedColors:    rankedColors,
	}
}

// GetCurrentState returns the current draft state (thread-safe).
func (o *Overlay) GetCurrentState() *DraftState {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentState
}

// GetCurrentRatings returns the current pack ratings (thread-safe).
func (o *Overlay) GetCurrentRatings() *PackRatings {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.currentRatings
}

// GetSelectedColors returns the currently selected colors (thread-safe).
func (o *Overlay) GetSelectedColors() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.selectedColors
}

// isDraftComplete checks if the draft has reached 45 picks.
func (o *Overlay) isDraftComplete() bool {
	if o.currentState == nil {
		return false
	}

	// A draft has 45 picks total (3 packs × 15 picks)
	return len(o.currentState.Picks) >= 45
}

// Reset clears the current draft state and prepares for the next draft.
func (o *Overlay) Reset() {
	o.currentState = nil
	if o.parser != nil {
		o.parser.Reset()
	}
	fmt.Println("[INFO] Draft overlay reset, ready for next draft")
}

// GetCacheStats returns current cache statistics.
// Returns nil if caching is disabled.
func (o *Overlay) GetCacheStats() *CacheStats {
	if o.cache == nil {
		return nil
	}
	stats := o.cache.GetStats()
	return &stats
}

// GetCacheHitRate returns the cache hit rate as a percentage.
// Returns 0.0 if caching is disabled.
func (o *Overlay) GetCacheHitRate() float64 {
	if o.cache == nil {
		return 0.0
	}
	return o.cache.GetHitRate()
}

// handleDraftComplete processes draft completion and builds deck recommendations.
func (o *Overlay) handleDraftComplete() {
	if o.currentState == nil {
		return
	}

	// Mark draft as complete
	o.currentState.Event.InProgress = false
	endTime := time.Now()
	o.currentState.Event.EndTime = &endTime

	fmt.Printf("[INFO] Draft complete! %d picks made\n", len(o.currentState.Picks))

	// Send draft end notification first
	if o.updateCallback != nil {
		o.updateCallback(&OverlayUpdate{
			Type:       UpdateTypeDraftEnd,
			DraftState: o.currentState,
			Timestamp:  time.Now(),
		})
	}

	// Build deck recommendations
	recommendation, err := BuildDeck(
		o.currentState.Picks,
		o.ratingsProvider,
		o.colorConfig,
	)
	if err != nil {
		o.logger.Error("Failed to build deck recommendations: %v", err)
		return
	}

	o.logger.Info("Deck recommendations: %s deck, %d main / %d sideboard, grade: %s",
		FormatColorName(recommendation.Colors),
		len(recommendation.MainDeck),
		len(recommendation.Sideboard),
		recommendation.DeckStrength.Grade)

	// Send deck builder update
	if o.updateCallback != nil {
		o.updateCallback(&OverlayUpdate{
			Type:               UpdateTypeDeckBuilder,
			DraftState:         o.currentState,
			DeckRecommendation: recommendation,
			Timestamp:          time.Now(),
		})
	}
}

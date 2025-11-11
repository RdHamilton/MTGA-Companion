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

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// Overlay manages the real-time draft overlay system.
type Overlay struct {
	parser          *Parser
	ratingsProvider *RatingsProvider
	colorConfig     ColorAffinityConfig
	currentState    *DraftState
	currentRatings  *PackRatings
	selectedColors  []string
	logPath         string
	updateCallback  func(*OverlayUpdate)
	stopChan        chan struct{}
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
}

// OverlayUpdate represents an update to send to the UI.
type OverlayUpdate struct {
	Type            UpdateType
	DraftState      *DraftState
	PackRatings     *PackRatings
	BestPick        *CardRating
	TopPicks        []*CardRating
	ColorSuggestion *ColorSuggestion
	Timestamp       time.Time
}

// UpdateType represents the type of overlay update.
type UpdateType string

const (
	UpdateTypeDraftStart UpdateType = "draft_start"
	UpdateTypeNewPack    UpdateType = "new_pack"
	UpdateTypePickMade   UpdateType = "pick_made"
	UpdateTypeDraftEnd   UpdateType = "draft_end"
	UpdateTypeColorRec   UpdateType = "color_recommendation"
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
	ratingsProvider := NewRatingsProvider(config.SetFile, config.BayesianConfig)

	if config.PollInterval == 0 {
		config.PollInterval = 500 * time.Millisecond
	}

	return &Overlay{
		parser:          parser,
		ratingsProvider: ratingsProvider,
		colorConfig:     config.ColorConfig,
		logPath:         config.LogPath,
		updateCallback:  config.UpdateCallback,
		stopChan:        make(chan struct{}),
	}
}

// Start begins monitoring the log file for draft events.
func (o *Overlay) Start(ctx context.Context) error {
	file, err := os.Open(o.logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Seek to end of file to only process new entries
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek to end of log: %w", err)
	}

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-o.stopChan:
			return nil
		case <-ticker.C:
			// Read new lines from the log
			if err := o.processNewLogLines(reader); err != nil {
				// Log error but continue monitoring
				continue
			}
		}
	}
}

// Stop stops the overlay monitoring.
func (o *Overlay) Stop() {
	close(o.stopChan)
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
	if err != nil || event == nil {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	// Update parser state
	if err := o.parser.UpdateState(event); err != nil {
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
		o.handlePackEvent()
	case IsPickEvent(event.Type):
		o.handlePickEvent()
	}
}

// handlePackEvent processes a new pack event.
func (o *Overlay) handlePackEvent() {
	if o.currentState.CurrentPack == nil {
		return
	}

	// Update color suggestion based on picks so far
	o.updateColorSuggestion()

	// Determine which color filter to use
	colorFilter := "ALL"
	if len(o.selectedColors) > 0 {
		colorFilter = strings.Join(o.selectedColors, "")
	}

	// Get ratings for the pack
	packRatings, err := o.ratingsProvider.GetPackRatings(o.currentState.CurrentPack, colorFilter)
	if err != nil {
		return
	}

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

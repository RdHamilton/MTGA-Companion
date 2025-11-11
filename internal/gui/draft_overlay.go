package gui

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft"
)

// DraftOverlayWindow manages the draft overlay UI.
type DraftOverlayWindow struct {
	app           fyne.App
	window        fyne.Window
	overlay       *draft.Overlay
	ctx           context.Context
	cancel        context.CancelFunc
	packContainer *fyne.Container
	colorLabel    *widget.Label
	statusLabel   *widget.Label
	updateChan    chan *draft.OverlayUpdate
}

// NewDraftOverlayWindow creates a new draft overlay window.
func NewDraftOverlayWindow(overlayConfig draft.OverlayConfig) *DraftOverlayWindow {
	ctx, cancel := context.WithCancel(context.Background())

	dow := &DraftOverlayWindow{
		app:        app.New(),
		ctx:        ctx,
		cancel:     cancel,
		updateChan: make(chan *draft.OverlayUpdate, 10),
	}

	// Set update callback to send to channel for thread-safe UI updates
	overlayConfig.UpdateCallback = func(update *draft.OverlayUpdate) {
		select {
		case dow.updateChan <- update:
		case <-dow.ctx.Done():
		}
	}

	// Create overlay controller
	dow.overlay = draft.NewOverlay(overlayConfig)

	return dow
}

// Run starts the overlay window and begins monitoring.
func (dow *DraftOverlayWindow) Run() {
	dow.window = dow.app.NewWindow("MTGA Draft Overlay")

	// Configure window for overlay behavior
	dow.window.Resize(fyne.NewSize(400, 600))
	dow.window.SetFixedSize(true)

	// Create UI components
	dow.statusLabel = widget.NewLabel("Ready - waiting for draft...")
	dow.colorLabel = widget.NewLabel("")
	dow.packContainer = container.NewVBox()

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			dow.statusLabel,
			widget.NewSeparator(),
			dow.colorLabel,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewScroll(dow.packContainer),
	)

	dow.window.SetContent(content)

	// Start overlay monitoring in background
	go func() {
		if err := dow.overlay.Start(dow.ctx); err != nil && err != context.Canceled {
			dow.statusLabel.SetText(fmt.Sprintf("Error: %v", err))
		}
	}()

	// Process updates from channel in background, queue to UI thread
	go func() {
		for {
			select {
			case update := <-dow.updateChan:
				// Queue update to Fyne's UI thread
				fyne.DoAndWait(func() {
					dow.handleUpdate(update)
				})
			case <-dow.ctx.Done():
				return
			}
		}
	}()

	// Handle window close
	dow.window.SetOnClosed(func() {
		dow.cancel()
		close(dow.updateChan)
	})

	dow.window.ShowAndRun()
}

// handleUpdate processes overlay updates and refreshes the UI.
func (dow *DraftOverlayWindow) handleUpdate(update *draft.OverlayUpdate) {
	switch update.Type {
	case draft.UpdateTypeDraftStart:
		dow.handleDraftStart(update)
	case draft.UpdateTypeNewPack:
		dow.handleNewPack(update)
	case draft.UpdateTypePickMade:
		dow.handlePickMade(update)
	case draft.UpdateTypeDraftEnd:
		dow.handleDraftEnd(update)
	case draft.UpdateTypeDeckBuilder:
		dow.handleDeckBuilder(update)
	}
}

// handleDraftStart updates UI when draft starts.
func (dow *DraftOverlayWindow) handleDraftStart(update *draft.OverlayUpdate) {
	dow.statusLabel.SetText("Draft detected! Waiting for first pack...")
	dow.colorLabel.SetText("")
	dow.packContainer.RemoveAll()
}

// handleNewPack updates UI when a new pack is received.
func (dow *DraftOverlayWindow) handleNewPack(update *draft.OverlayUpdate) {
	if update.PackRatings == nil {
		return
	}

	// Update status
	packNum := update.PackRatings.Pack.PackNumber
	pickNum := update.PackRatings.Pack.PickNumber
	dow.statusLabel.SetText(fmt.Sprintf("Pack %d, Pick %d", packNum, pickNum))

	// Update color suggestion
	if update.ColorSuggestion != nil {
		dow.updateColorSuggestion(update.ColorSuggestion)
	}

	// Clear and rebuild pack display
	dow.packContainer.RemoveAll()

	// Show top picks
	topPicks := update.TopPicks
	if len(topPicks) == 0 {
		dow.packContainer.Add(widget.NewLabel("No ratings available"))
		dow.packContainer.Refresh()
		return
	}

	// Add header
	header := widget.NewLabelWithStyle(
		"Top Picks:",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	dow.packContainer.Add(header)

	// Add each pick
	for i, pick := range topPicks {
		pickWidget := dow.createPickWidget(i+1, pick, i == 0)
		dow.packContainer.Add(pickWidget)
	}

	dow.packContainer.Refresh()
}

// handlePickMade updates UI when a pick is made.
func (dow *DraftOverlayWindow) handlePickMade(update *draft.OverlayUpdate) {
	if update.ColorSuggestion != nil {
		dow.updateColorSuggestion(update.ColorSuggestion)
	}

	// Clear pack display after pick
	dow.packContainer.RemoveAll()
	dow.packContainer.Add(widget.NewLabel("Waiting for next pack..."))
	dow.packContainer.Refresh()
}

// handleDraftEnd updates UI when draft ends.
func (dow *DraftOverlayWindow) handleDraftEnd(update *draft.OverlayUpdate) {
	dow.statusLabel.SetText("Draft complete!")
	dow.colorLabel.SetText("")
	dow.packContainer.RemoveAll()
	dow.packContainer.Add(widget.NewLabel("Draft finished. Waiting for next draft..."))
	dow.packContainer.Refresh()
}

// handleDeckBuilder shows deck building recommendations.
func (dow *DraftOverlayWindow) handleDeckBuilder(update *draft.OverlayUpdate) {
	if update.DeckRecommendation == nil {
		return
	}

	rec := update.DeckRecommendation

	// Update status
	dow.statusLabel.SetText(fmt.Sprintf("Deck Builder - %s", draft.FormatColorName(rec.Colors)))

	// Show deck grade and stats
	gradeText := fmt.Sprintf("Grade: %s (%.1f avg GIHWR)",
		rec.DeckStrength.Grade,
		rec.DeckStrength.OverallRating)
	dow.colorLabel.SetText(gradeText)

	// Clear and rebuild deck display
	dow.packContainer.RemoveAll()

	// Add deck overview
	overview := widget.NewLabelWithStyle(
		fmt.Sprintf("Recommended: %d Main / %d Sideboard / %d Lands",
			len(rec.MainDeck), len(rec.Sideboard), rec.Lands.TotalLands),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	dow.packContainer.Add(overview)

	// Add land recommendation
	landText := fmt.Sprintf("Lands: %s", rec.Lands.Explanation)
	for color, count := range rec.Lands.ColorSplit {
		landText += fmt.Sprintf("\n  %s: %d", draft.FormatColorName(color), count)
	}
	dow.packContainer.Add(widget.NewLabel(landText))

	// Add mana curve
	curveText := fmt.Sprintf("Curve: Avg CMC %.1f (%d creatures, %d non-creatures)",
		rec.ManaCurve.AvgCMC, rec.ManaCurve.Creatures, rec.ManaCurve.NonCreatures)
	dow.packContainer.Add(widget.NewLabel(curveText))

	dow.packContainer.Add(widget.NewSeparator())

	// Add main deck section
	mainDeckHeader := widget.NewLabelWithStyle(
		"Main Deck (23 spells):",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	dow.packContainer.Add(mainDeckHeader)

	for i, card := range rec.MainDeck {
		if i >= 23 {
			break // Only show first 23
		}
		cardWidget := dow.createDeckCardWidget(card, true)
		dow.packContainer.Add(cardWidget)
	}

	dow.packContainer.Add(widget.NewSeparator())

	// Add sideboard section
	sideboardHeader := widget.NewLabelWithStyle(
		fmt.Sprintf("Sideboard (%d cards):", len(rec.Sideboard)),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	dow.packContainer.Add(sideboardHeader)

	// Show first 10 sideboard cards
	for i, card := range rec.Sideboard {
		if i >= 10 {
			break
		}
		cardWidget := dow.createDeckCardWidget(card, false)
		dow.packContainer.Add(cardWidget)
	}

	if len(rec.Sideboard) > 10 {
		dow.packContainer.Add(widget.NewLabel(fmt.Sprintf("... and %d more", len(rec.Sideboard)-10)))
	}

	dow.packContainer.Refresh()
}

// createDeckCardWidget creates a widget for a deck card selection.
func (dow *DraftOverlayWindow) createDeckCardWidget(card draft.CardSelection, isMainDeck bool) fyne.CanvasObject {
	// Format card name with colors
	nameText := card.Name
	if len(card.Colors) > 0 {
		nameText += fmt.Sprintf(" [%s]", strings.Join(card.Colors, ""))
	}

	// Format rating
	ratingText := fmt.Sprintf("%.1f%%", card.GIHWR)

	// Create container
	nameLabel := widget.NewLabel(fmt.Sprintf("%s - %s", nameText, ratingText))

	// Add styling for main deck cards
	if isMainDeck {
		nameLabel = widget.NewLabelWithStyle(
			fmt.Sprintf("%s - %s", nameText, ratingText),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)
	}

	return container.NewVBox(nameLabel)
}

// updateColorSuggestion updates the color suggestion display.
func (dow *DraftOverlayWindow) updateColorSuggestion(suggestion *draft.ColorSuggestion) {
	if suggestion == nil || len(suggestion.SuggestedColors) == 0 {
		dow.colorLabel.SetText("")
		return
	}

	// Format color names
	var colorNames []string
	for _, colors := range suggestion.SuggestedColors {
		colorNames = append(colorNames, draft.FormatColorName(colors))
	}

	text := fmt.Sprintf("Suggested Colors: %s", strings.Join(colorNames, " or "))
	dow.colorLabel.SetText(text)
}

// createPickWidget creates a widget displaying a single pick.
func (dow *DraftOverlayWindow) createPickWidget(rank int, pick *draft.CardRating, isBest bool) fyne.CanvasObject {
	// Format card name with rank
	nameText := fmt.Sprintf("%d. %s", rank, pick.Name)

	// Format rating
	ratingText := fmt.Sprintf("%.1f%%", pick.BayesianGIHWR)
	if pick.IsBayesianAdjust {
		ratingText += " (adj)"
	}

	// Format mana cost and rarity
	infoText := fmt.Sprintf("%s - %s", pick.ManaCost, pick.Rarity)

	// Create labels
	nameLabel := widget.NewLabel(nameText)
	if isBest {
		nameLabel = widget.NewLabelWithStyle(
			nameText,
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)
	}

	ratingLabel := widget.NewLabel(ratingText)
	infoLabel := widget.NewLabel(infoText)

	// Layout: name on left, rating on right
	topRow := container.NewBorder(nil, nil, nil, ratingLabel, nameLabel)

	return container.NewVBox(
		topRow,
		infoLabel,
		widget.NewSeparator(),
	)
}

// Stop stops the overlay monitoring.
func (dow *DraftOverlayWindow) Stop() {
	dow.cancel()
	dow.overlay.Stop()
}

package gui

import (
	"context"
	"fmt"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// deckPerformanceData holds deck statistics for display.
type deckPerformanceData struct {
	deckID     string
	deckName   string
	matches    int
	winRate    float64
	wins       int
	losses     int
	confidence string // "High", "Medium", "Low"
}

// DeckPerformanceDashboard manages the deck performance comparison view.
type DeckPerformanceDashboard struct {
	app         *App
	service     *storage.Service
	ctx         context.Context
	startDate   *time.Time
	endDate     *time.Time
	format      string // "all" or specific format
	sortBy      string // "winrate", "matches", "name"
	updateChart func() // Function to update the chart without recreating tabs
}

// NewDeckPerformanceDashboard creates a new deck performance dashboard.
func NewDeckPerformanceDashboard(app *App, service *storage.Service, ctx context.Context) *DeckPerformanceDashboard {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	return &DeckPerformanceDashboard{
		app:       app,
		service:   service,
		ctx:       ctx,
		startDate: &thirtyDaysAgo,
		endDate:   &now,
		format:    "all",
		sortBy:    "winrate",
	}
}

// CreateView creates the deck performance comparison view.
func (d *DeckPerformanceDashboard) CreateView() fyne.CanvasObject {
	// Create filter controls
	filterControls := d.createFilterControls()

	// Create a container for the chart view that we can update
	chartContainer := container.NewVBox()

	// Function to update the chart
	updateChart := func() {
		chartView := d.createChartView()
		chartContainer.Objects = []fyne.CanvasObject{chartView}
		chartContainer.Refresh()
	}

	// Store the update function so other methods can use it
	d.updateChart = updateChart

	// Initial chart render
	updateChart()

	// Layout
	return container.NewBorder(
		container.NewPadded(filterControls),
		nil,
		nil,
		nil,
		container.NewScroll(container.NewPadded(chartContainer)),
	)
}

// createFilterControls creates the filter control panel.
func (d *DeckPerformanceDashboard) createFilterControls() fyne.CanvasObject {
	// Date range selector
	dateRangeLabel := widget.NewLabelWithStyle("Date Range", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dateRangeSelect := widget.NewSelect(
		[]string{"Last 7 Days", "Last 30 Days", "Last 90 Days", "All Time"},
		func(selected string) {
			now := time.Now()
			switch selected {
			case "Last 7 Days":
				start := now.AddDate(0, 0, -7)
				d.startDate = &start
				d.endDate = &now
			case "Last 30 Days":
				start := now.AddDate(0, 0, -30)
				d.startDate = &start
				d.endDate = &now
			case "Last 90 Days":
				start := now.AddDate(0, 0, -90)
				d.startDate = &start
				d.endDate = &now
			case "All Time":
				d.startDate = nil
				d.endDate = nil
			}
			if d.updateChart != nil {
				d.updateChart()
			}
		},
	)
	dateRangeSelect.Selected = "Last 30 Days"

	// Format selector
	formatLabel := widget.NewLabelWithStyle("Format", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	formatSelect := widget.NewSelect(
		[]string{"All Formats", "Constructed", "Limited", "Standard", "Historic", "Alchemy", "Explorer", "Timeless"},
		func(selected string) {
			if selected == "All Formats" {
				d.format = "all"
			} else {
				d.format = selected
			}
			if d.updateChart != nil {
				d.updateChart()
			}
		},
	)
	formatSelect.Selected = "All Formats"

	// Sort by selector
	sortByLabel := widget.NewLabelWithStyle("Sort By", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	sortBySelect := widget.NewSelect(
		[]string{"Win Rate", "Match Count", "Deck Name"},
		func(selected string) {
			switch selected {
			case "Win Rate":
				d.sortBy = "winrate"
			case "Match Count":
				d.sortBy = "matches"
			case "Deck Name":
				d.sortBy = "name"
			}
			if d.updateChart != nil {
				d.updateChart()
			}
		},
	)
	sortBySelect.Selected = "Win Rate"

	// Refresh button
	refreshButton := widget.NewButton("Refresh", func() {
		if d.updateChart != nil {
			d.updateChart()
		}
	})

	// Layout controls
	return container.NewVBox(
		container.NewGridWithColumns(3,
			container.NewVBox(dateRangeLabel, dateRangeSelect),
			container.NewVBox(formatLabel, formatSelect),
			container.NewVBox(sortByLabel, sortBySelect),
		),
		container.NewHBox(refreshButton),
		widget.NewSeparator(),
	)
}

// createChartView creates the chart visualization.
func (d *DeckPerformanceDashboard) createChartView() fyne.CanvasObject {
	// Create filter for date range and format
	filter := storage.StatsFilter{
		StartDate: d.startDate,
		EndDate:   d.endDate,
	}

	// Add format filter if not "all"
	if d.format != "all" {
		filter.Format = &d.format
	}

	// Get statistics by deck
	statsByDeck, err := d.service.GetStatsByDeck(d.ctx, filter)
	if err != nil || len(statsByDeck) == 0 {
		return container.NewCenter(
			container.NewVBox(
				widget.NewLabel("No deck performance data available"),
				widget.NewLabel(fmt.Sprintf("Error: %v", err)),
			),
		)
	}

	// Convert to performance data
	var decks []deckPerformanceData

	for deckID, stats := range statsByDeck {
		if stats.TotalMatches > 0 {
			// Determine confidence level based on sample size
			confidence := "Low"
			if stats.TotalMatches >= 30 {
				confidence = "High"
			} else if stats.TotalMatches >= 10 {
				confidence = "Medium"
			}

			// Use deck ID as name for now
			// In the future, this could be enhanced to fetch actual deck names from the decks table
			deckName := deckID
			if len(deckName) > 30 {
				deckName = deckName[:27] + "..."
			}

			decks = append(decks, deckPerformanceData{
				deckID:     deckID,
				deckName:   deckName,
				matches:    stats.TotalMatches,
				winRate:    stats.WinRate * 100,
				wins:       stats.MatchesWon,
				losses:     stats.MatchesLost,
				confidence: confidence,
			})
		}
	}

	// Sort decks based on sort criteria
	d.sortDecks(decks)

	// Limit to top 10 decks for better visualization
	if len(decks) > 10 {
		decks = decks[:10]
	}

	// Create data points for chart
	dataPoints := make([]charts.DataPoint, len(decks))
	for i, deck := range decks {
		label := fmt.Sprintf("%s (%d)", deck.deckName, deck.matches)
		dataPoints[i] = charts.DataPoint{
			Label: label,
			Value: deck.winRate,
		}
	}

	// Create chart config
	config := charts.DefaultFyneChartConfig()
	config.Title = d.getChartTitle()
	config.Width = 900
	config.Height = 600

	// Create horizontal bar chart
	chart := charts.CreateFyneBarChart(dataPoints, config)

	// Create summary
	summary := d.createSummary(decks, statsByDeck)

	// Create clickable deck list
	deckList := d.createDeckList(decks)

	// Layout
	return container.NewVBox(
		chart,
		widget.NewSeparator(),
		summary,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Deck Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		deckList,
	)
}

// sortDecks sorts the decks based on the current sort criteria.
func (d *DeckPerformanceDashboard) sortDecks(decks []deckPerformanceData) {
	switch d.sortBy {
	case "winrate":
		sort.Slice(decks, func(i, j int) bool {
			return decks[i].winRate > decks[j].winRate
		})
	case "matches":
		sort.Slice(decks, func(i, j int) bool {
			return decks[i].matches > decks[j].matches
		})
	case "name":
		sort.Slice(decks, func(i, j int) bool {
			return decks[i].deckName < decks[j].deckName
		})
	}
}

// createSummary creates the summary information display.
func (d *DeckPerformanceDashboard) createSummary(decks []deckPerformanceData, allDecks map[string]*storage.Statistics) fyne.CanvasObject {
	// Calculate overall statistics
	totalMatches := 0
	totalWins := 0
	for _, stats := range allDecks {
		totalMatches += stats.TotalMatches
		totalWins += stats.MatchesWon
	}

	overallWinRate := 0.0
	if totalMatches > 0 {
		overallWinRate = float64(totalWins) / float64(totalMatches) * 100
	}

	// Build summary content
	summaryContent := fmt.Sprintf(`### Deck Performance Analysis

**Period**: %s
**Format**: %s
**Total Matches**: %d
**Overall Win Rate**: %.1f%%
**Unique Decks**: %d (showing top %d)
**Sort Order**: %s

`,
		d.getDateRangeDescription(),
		d.getFormatDisplayName(),
		totalMatches,
		overallWinRate,
		len(allDecks),
		len(decks),
		d.getSortDisplayName(),
	)

	// Add top performers
	if len(decks) > 0 {
		summaryContent += "**Top Performers**:\n\n"
		for i, deck := range decks {
			if i >= 3 {
				break
			}
			summaryContent += fmt.Sprintf("%d. **%s**: %.1f%% win rate (%d-%d, %d total) - %s confidence\n",
				i+1, deck.deckName, deck.winRate, deck.wins, deck.losses, deck.matches, deck.confidence)
		}
	}

	return container.NewVBox(
		widget.NewRichTextFromMarkdown(summaryContent),
	)
}

// createDeckList creates an interactive list of decks.
func (d *DeckPerformanceDashboard) createDeckList(decks []deckPerformanceData) fyne.CanvasObject {
	// Create list items
	var items []string
	for _, deck := range decks {
		items = append(items, fmt.Sprintf("%s - %.1f%% (%d matches) - %s confidence",
			deck.deckName, deck.winRate, deck.matches, deck.confidence))
	}

	// Create list widget
	list := widget.NewList(
		func() int {
			return len(items)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(items[id])
		},
	)

	// Set list size
	list.Resize(fyne.NewSize(800, 200))

	// Add tap handler to show deck details
	list.OnSelected = func(id widget.ListItemID) {
		deck := decks[id]
		d.showDeckDetails(deck)
	}

	return list
}

// showDeckDetails shows detailed information about a deck.
func (d *DeckPerformanceDashboard) showDeckDetails(deck deckPerformanceData) {
	message := fmt.Sprintf(`**Deck**: %s

**Performance**:
- Win Rate: %.1f%%
- Wins: %d
- Losses: %d
- Total Matches: %d
- Confidence: %s

**Sample Size Analysis**:
%s

**Note**: Click "View Match History" in the Match History tab to see detailed match data for this deck.`,
		deck.deckName,
		deck.winRate,
		deck.wins,
		deck.losses,
		deck.matches,
		deck.confidence,
		d.getSampleSizeAnalysis(deck),
	)

	dialog.ShowInformation("Deck Details", message, d.app.window)
}

// getSampleSizeAnalysis returns analysis of the sample size.
func (d *DeckPerformanceDashboard) getSampleSizeAnalysis(deck deckPerformanceData) string {
	if deck.matches < 10 {
		return "⚠️ Low confidence: Less than 10 matches. Win rate may not be representative."
	} else if deck.matches < 30 {
		return "⚠️ Medium confidence: 10-29 matches. Win rate is fairly reliable but could vary."
	}
	return "✅ High confidence: 30+ matches. Win rate is statistically reliable."
}

// getChartTitle returns the appropriate chart title.
func (d *DeckPerformanceDashboard) getChartTitle() string {
	title := "Deck Performance Comparison"

	if d.format != "all" {
		title += fmt.Sprintf(" - %s", d.format)
	}

	return title
}

// getDateRangeDescription returns a human-readable date range description.
func (d *DeckPerformanceDashboard) getDateRangeDescription() string {
	if d.startDate == nil || d.endDate == nil {
		return "All Time"
	}
	return fmt.Sprintf("%s to %s", d.startDate.Format("2006-01-02"), d.endDate.Format("2006-01-02"))
}

// getFormatDisplayName returns the display name for the current format filter.
func (d *DeckPerformanceDashboard) getFormatDisplayName() string {
	if d.format == "all" {
		return "All Formats"
	}
	return d.format
}

// getSortDisplayName returns the display name for the current sort order.
func (d *DeckPerformanceDashboard) getSortDisplayName() string {
	switch d.sortBy {
	case "winrate":
		return "Win Rate (Highest First)"
	case "matches":
		return "Match Count (Most Played First)"
	case "name":
		return "Deck Name (Alphabetical)"
	default:
		return "Unknown"
	}
}

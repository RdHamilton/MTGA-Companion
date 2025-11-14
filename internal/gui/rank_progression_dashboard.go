package gui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// RankProgressionDashboard manages the rank progression chart view.
type RankProgressionDashboard struct {
	app     *App
	service *storage.Service
	ctx     context.Context

	// Filter state
	format     string // Selected format (Constructed, Limited, Ladder, Play)
	startDate  *time.Time
	endDate    *time.Time
	periodType string

	updateChart func() // Function to update the chart without recreating tabs
}

// NewRankProgressionDashboard creates a new rank progression dashboard.
func NewRankProgressionDashboard(app *App, service *storage.Service, ctx context.Context) *RankProgressionDashboard {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	return &RankProgressionDashboard{
		app:        app,
		service:    service,
		ctx:        ctx,
		format:     "Ladder", // Default to ranked ladder
		startDate:  &thirtyDaysAgo,
		endDate:    &now,
		periodType: "weekly",
	}
}

// CreateView creates the rank progression dashboard view.
func (d *RankProgressionDashboard) CreateView() fyne.CanvasObject {
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
func (d *RankProgressionDashboard) createFilterControls() fyne.CanvasObject {
	// Format selector - using actual database format values
	formatLabel := widget.NewLabelWithStyle("Format", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	formatSelect := widget.NewSelect(
		[]string{"Ladder", "Play"},
		func(selected string) {
			d.format = selected
			if d.updateChart != nil {
				d.updateChart()
			}
		},
	)
	formatSelect.Selected = d.format

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

	// Layout controls in a grid
	return container.NewVBox(
		formatLabel,
		formatSelect,
		dateRangeLabel,
		dateRangeSelect,
	)
}

// createChartView creates the chart view with current filters.
func (d *RankProgressionDashboard) createChartView() fyne.CanvasObject {
	// Use default dates if nil
	startDate := d.startDate
	endDate := d.endDate
	if startDate == nil {
		defaultStart := time.Now().AddDate(-1, 0, 0) // 1 year ago
		startDate = &defaultStart
	}
	if endDate == nil {
		now := time.Now()
		endDate = &now
	}

	// Get rank progression timeline using actual database format value
	timeline, err := d.service.GetRankProgressionTimeline(d.ctx, d.format, startDate, endDate, storage.PeriodWeekly)
	if err != nil {
		return d.app.ErrorView("Error Loading Rank Data", err, nil)
	}

	if len(timeline.Entries) == 0 {
		return d.app.NoDataView("No Rank Data Available",
			fmt.Sprintf("No ranked matches found for %s in the selected time period.", d.format))
	}

	// Convert timeline entries to chart data points
	dataPoints := make([]charts.DataPoint, len(timeline.Entries))
	for i, entry := range timeline.Entries {
		dataPoints[i] = charts.DataPoint{
			Label: entry.Date,
			Value: d.app.rankToNumericValue(entry.RankClass, entry.RankLevel),
		}
	}

	// Create chart config
	config := charts.DefaultFyneChartConfig()
	config.Title = fmt.Sprintf("Rank Progression - %s", d.format)
	config.Width = 750
	config.Height = 450

	// Create chart
	chart := charts.CreateFyneLineChart(dataPoints, config)

	// Summary with markdown formatting
	summaryContent := fmt.Sprintf(`### Rank Progression Analysis

**Format**: %s
**Period**: %s to %s
**Start Rank**: %s
**End Rank**: %s
**Highest Rank**: %s
**Lowest Rank**: %s
**Total Changes**: %d
**Milestones**: %d`,
		d.format,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		timeline.StartRank,
		timeline.EndRank,
		timeline.HighestRank,
		timeline.LowestRank,
		timeline.TotalChanges,
		timeline.Milestones,
	)

	summary := widget.NewRichTextFromMarkdown(summaryContent)

	// Layout
	return container.NewVBox(
		chart,
		widget.NewSeparator(),
		summary,
	)
}

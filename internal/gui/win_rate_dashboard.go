package gui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// WinRateDashboard manages the win rate charts view with filtering and export.
type WinRateDashboard struct {
	app     *App
	service *storage.Service
	ctx     context.Context

	// Filter state
	dateRange  string     // "7days", "30days", "90days", "alltime", "custom"
	startDate  *time.Time // For custom range
	endDate    *time.Time // For custom range
	format     string     // "all", "Constructed", "Limited", etc.
	chartType  string     // "line", "bar"
	periodType string     // "daily", "weekly", "monthly"

	updateChart func() // Function to update the chart without recreating tabs
}

// NewWinRateDashboard creates a new win rate dashboard.
func NewWinRateDashboard(app *App, service *storage.Service, ctx context.Context) *WinRateDashboard {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	return &WinRateDashboard{
		app:        app,
		service:    service,
		ctx:        ctx,
		dateRange:  "30days",
		startDate:  &thirtyDaysAgo,
		endDate:    &now,
		format:     "all",
		chartType:  "line",
		periodType: "weekly",
	}
}

// CreateView creates the complete win rate dashboard view.
func (d *WinRateDashboard) CreateView() fyne.CanvasObject {
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
func (d *WinRateDashboard) createFilterControls() fyne.CanvasObject {
	// Date range selector
	dateRangeLabel := widget.NewLabelWithStyle("Date Range", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dateRangeSelect := widget.NewSelect(
		[]string{"Last 7 Days", "Last 30 Days", "Last 90 Days", "All Time", "Custom Range"},
		func(selected string) {
			now := time.Now()
			switch selected {
			case "Last 7 Days":
				d.dateRange = "7days"
				start := now.AddDate(0, 0, -7)
				d.startDate = &start
				d.endDate = &now
				d.periodType = "daily"
			case "Last 30 Days":
				d.dateRange = "30days"
				start := now.AddDate(0, 0, -30)
				d.startDate = &start
				d.endDate = &now
				d.periodType = "weekly"
			case "Last 90 Days":
				d.dateRange = "90days"
				start := now.AddDate(0, 0, -90)
				d.startDate = &start
				d.endDate = &now
				d.periodType = "weekly"
			case "All Time":
				d.dateRange = "alltime"
				d.startDate = nil
				d.endDate = nil
				d.periodType = "monthly"
			case "Custom Range":
				d.dateRange = "custom"
				d.showCustomDateDialog()
				return
			}
			if d.updateChart != nil {
				d.updateChart()
			}
		},
	)
	dateRangeSelect.Selected = "Last 30 Days"
	AddSelectTooltip(dateRangeSelect, TooltipDateRange)

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
	AddSelectTooltip(formatSelect, TooltipFormat)

	// Chart type selector
	chartTypeLabel := widget.NewLabelWithStyle("Chart Type", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	chartTypeSelect := widget.NewSelect(
		[]string{"Line Chart", "Bar Chart"},
		func(selected string) {
			if selected == "Line Chart" {
				d.chartType = "line"
			} else {
				d.chartType = "bar"
			}
			if d.updateChart != nil {
				d.updateChart()
			}
		},
	)
	chartTypeSelect.Selected = "Line Chart"
	AddSelectTooltip(chartTypeSelect, TooltipChartType)

	// Export button
	exportButton := widget.NewButton("Export as PNG", func() {
		d.exportChart()
	})
	AddButtonTooltip(exportButton, TooltipExport)

	// Refresh button
	refreshButton := widget.NewButton("Refresh", func() {
		if d.updateChart != nil {
			d.updateChart()
		}
	})
	AddButtonTooltip(refreshButton, TooltipRefresh)

	// Layout controls in a grid
	return container.NewVBox(
		container.NewGridWithColumns(2,
			container.NewVBox(dateRangeLabel, dateRangeSelect),
			container.NewVBox(formatLabel, formatSelect),
		),
		container.NewGridWithColumns(2,
			container.NewVBox(chartTypeLabel, chartTypeSelect),
			container.NewHBox(refreshButton, exportButton),
		),
		widget.NewSeparator(),
	)
}

// createChartView creates the chart visualization based on current filters.
func (d *WinRateDashboard) createChartView() fyne.CanvasObject {
	// Determine format filter
	var formatFilter *string
	if d.format != "all" {
		formatFilter = &d.format
	}

	// Get trend data
	analysis, err := d.service.GetTrendAnalysis(d.ctx, *d.startDate, *d.endDate, d.periodType, formatFilter)
	if err != nil {
		return d.app.ErrorView("Error Loading Trend Data", err, nil)
	}

	if len(analysis.Periods) == 0 {
		return d.app.NoDataView("No Trend Data Available",
			"No match data found for the selected time period.")
	}

	// Prepare data points
	dataPoints := make([]charts.DataPoint, len(analysis.Periods))
	for i, period := range analysis.Periods {
		dataPoints[i] = charts.DataPoint{
			Label: period.Period.Label,
			Value: period.WinRate,
		}
	}

	// Create chart config
	config := charts.DefaultFyneChartConfig()
	config.Title = d.getChartTitle()
	config.Width = 900
	config.Height = 500

	// Create chart based on type
	var chart fyne.CanvasObject
	if d.chartType == "line" {
		chart = charts.CreateFyneLineChart(dataPoints, config)
	} else {
		chart = charts.CreateFyneBarChart(dataPoints, config)
	}

	// Create summary
	summary := d.createSummary(analysis)

	// Layout
	return container.NewVBox(
		chart,
		widget.NewSeparator(),
		summary,
	)
}

// createSummary creates the summary information display.
func (d *WinRateDashboard) createSummary(analysis *storage.TrendAnalysis) fyne.CanvasObject {
	summaryContent := fmt.Sprintf(`### Win Rate Trend Analysis

**Period**: %s to %s
**Format**: %s
**Trend**: %s`,
		d.startDate.Format("2006-01-02"),
		d.endDate.Format("2006-01-02"),
		d.getFormatDisplayName(),
		analysis.Trend,
	)

	if analysis.TrendValue != 0 {
		summaryContent += fmt.Sprintf(" (%.1f%%)", analysis.TrendValue)
	}

	if analysis.Overall != nil {
		summaryContent += fmt.Sprintf(`
**Overall Win Rate**: %.1f%% (%d matches)`,
			analysis.Overall.WinRate,
			analysis.Overall.TotalMatches,
		)
	}

	return widget.NewRichTextFromMarkdown(summaryContent)
}

// getChartTitle returns the appropriate chart title based on current filters.
func (d *WinRateDashboard) getChartTitle() string {
	title := "Win Rate Trend"

	// Add format if filtered
	if d.format != "all" {
		title += fmt.Sprintf(" - %s", d.format)
	}

	// Add date range
	switch d.dateRange {
	case "7days":
		title += " (Last 7 Days)"
	case "30days":
		title += " (Last 30 Days)"
	case "90days":
		title += " (Last 90 Days)"
	case "alltime":
		title += " (All Time)"
	case "custom":
		title += " (Custom Range)"
	}

	return title
}

// getFormatDisplayName returns the display name for the current format filter.
func (d *WinRateDashboard) getFormatDisplayName() string {
	if d.format == "all" {
		return "All Formats"
	}
	return d.format
}

// showCustomDateDialog shows a dialog for selecting custom date range.
func (d *WinRateDashboard) showCustomDateDialog() {
	// Create date entry widgets
	startEntry := widget.NewEntry()
	startEntry.SetPlaceHolder("YYYY-MM-DD")
	if d.startDate != nil {
		startEntry.SetText(d.startDate.Format("2006-01-02"))
	}

	endEntry := widget.NewEntry()
	endEntry.SetPlaceHolder("YYYY-MM-DD")
	if d.endDate != nil {
		endEntry.SetText(d.endDate.Format("2006-01-02"))
	}

	// Create form items
	formItems := []*widget.FormItem{
		{Text: "Start Date", Widget: startEntry, HintText: "Format: YYYY-MM-DD (e.g., 2024-01-01)"},
		{Text: "End Date", Widget: endEntry, HintText: "Format: YYYY-MM-DD (e.g., 2024-12-31)"},
	}

	// Define submit handler
	onSubmit := func(submitted bool) {
		if !submitted {
			// User cancelled
			return
		}

		// Parse dates
		start, err := time.Parse("2006-01-02", startEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid start date: %w", err), d.app.window)
			return
		}

		end, err := time.Parse("2006-01-02", endEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid end date: %w", err), d.app.window)
			return
		}

		// Validate range
		if start.After(end) {
			dialog.ShowError(fmt.Errorf("start date must be before end date"), d.app.window)
			return
		}

		// Update state
		d.startDate = &start
		d.endDate = &end
		d.periodType = d.determinePeriodType(start, end)
		if d.updateChart != nil {
			d.updateChart()
		}
	}

	// Show dialog
	dialog.ShowForm("Custom Date Range", "Apply", "Cancel", formItems, onSubmit, d.app.window)
}

// determinePeriodType determines the best period type based on date range.
func (d *WinRateDashboard) determinePeriodType(start, end time.Time) string {
	days := int(end.Sub(start).Hours() / 24)

	if days <= 14 {
		return "daily"
	} else if days <= 180 {
		return "weekly"
	}
	return "monthly"
}

// exportChart exports the current chart as a PNG image.
// Note: Fyne doesn't have built-in image export, so for now we show a message.
// Future implementation could use headless rendering or external libraries.
func (d *WinRateDashboard) exportChart() {
	message := fmt.Sprintf(`Chart export functionality is coming soon!

For now, you can take a screenshot of the chart using:
• macOS: Cmd+Shift+4 (select area)
• Windows: Win+Shift+S
• Linux: PrtScn or Shift+PrtScn

Current view:
- Date Range: %s
- Format: %s
- Chart Type: %s`,
		d.getDateRangeDescription(),
		d.getFormatDisplayName(),
		d.chartType,
	)

	dialog.ShowInformation("Export Chart", message, d.app.window)
}

// getDateRangeDescription returns a human-readable date range description.
func (d *WinRateDashboard) getDateRangeDescription() string {
	if d.startDate == nil || d.endDate == nil {
		return "All Time"
	}
	return fmt.Sprintf("%s to %s", d.startDate.Format("2006-01-02"), d.endDate.Format("2006-01-02"))
}

package gui

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
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
	sevenDaysAgo := now.AddDate(0, 0, -7)

	return &WinRateDashboard{
		app:        app,
		service:    service,
		ctx:        ctx,
		dateRange:  "7days",
		startDate:  &sevenDaysAgo,
		endDate:    &now,
		format:     "all",
		chartType:  "line",
		periodType: "daily",
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

	// Show loading placeholder initially
	loadingLabel := widget.NewLabel("Loading chart data...")
	chartContainer.Objects = []fyne.CanvasObject{
		container.NewCenter(loadingLabel),
	}

	// Load chart asynchronously
	go func() {
		// Run the query in background
		chartView := d.createChartView()

		// Update UI on main thread
		chartContainer.Objects = []fyne.CanvasObject{chartView}
		chartContainer.Refresh()
	}()

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
	dateRangeSelect.Selected = "Last 7 Days"
	AddSelectTooltip(dateRangeSelect, TooltipDateRange)

	// Format selector (simplified to match actual database values)
	formatLabel := widget.NewLabelWithStyle("Format", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	formatSelect := widget.NewSelect(
		[]string{"All Formats", "Constructed", "Limited"},
		func(selected string) {
			d.format = selected
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

	// Layout controls in a grid (removed Refresh button, Export moved below chart)
	return container.NewVBox(
		container.NewGridWithColumns(3,
			container.NewVBox(dateRangeLabel, dateRangeSelect),
			container.NewVBox(formatLabel, formatSelect),
			container.NewVBox(chartTypeLabel, chartTypeSelect),
		),
		widget.NewSeparator(),
	)
}

// createChartView creates the chart visualization based on current filters.
func (d *WinRateDashboard) createChartView() fyne.CanvasObject {
	// Validate that we have date range
	if d.startDate == nil || d.endDate == nil {
		return d.app.NoDataView("No Date Range Selected",
			"Please select a date range to view trend data.")
	}

	// Map user-friendly format names to actual database values
	var formatFilter []string
	switch d.format {
	case "Constructed":
		// Constructed formats: Ladder (ranked) and Play (unranked)
		formatFilter = []string{"Ladder", "Play"}
	case "Limited":
		// Limited formats: Any event containing Draft or Sealed
		// Query database for all draft/sealed format values
		formatFilter = d.getLimitedFormats()
	case "All Formats":
		// No filter
	}

	// Get trend data using format array
	analysis, err := d.service.GetTrendAnalysisWithFormats(d.ctx, *d.startDate, *d.endDate, d.periodType, formatFilter)
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

	// Create chart config (no title - it's in the summary below)
	config := charts.DefaultFyneChartConfig()
	config.Title = "" // Remove title from chart, show in summary instead
	config.Width = 1400 // Wider to span more of the page
	config.Height = 500

	// Create chart based on type
	var chart fyne.CanvasObject
	if d.chartType == "line" {
		chart = charts.CreateFyneLineChart(dataPoints, config)
	} else {
		chart = charts.CreateFyneBarChart(dataPoints, config)
	}

	// Create a spacer that reserves vertical space for the chart
	// This ensures VBox layout positions subsequent items below the chart
	chartSpacer := canvas.NewRectangle(color.Transparent)
	chartSpacer.SetMinSize(fyne.NewSize(config.Width, config.Height))

	// Stack the chart on top of the spacer
	chartWithSpace := container.NewStack(chartSpacer, chart)

	// Create summary
	summary := d.createSummary(analysis)

	// Export button (below chart, right-aligned)
	exportButton := widget.NewButton("Export as PNG", func() {
		d.exportChart()
	})
	AddButtonTooltip(exportButton, TooltipExport)

	// Layout: chart, summary, then export button below
	return container.NewVBox(
		chartWithSpace,
		widget.NewSeparator(),
		widget.NewSeparator(), // Extra separator for spacing
		summary,
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			exportButton,
		),
	)
}

// createSummary creates the summary information display.
func (d *WinRateDashboard) createSummary(analysis *storage.TrendAnalysis) fyne.CanvasObject {
	// Build the title with chart type
	chartTypeStr := "Line Chart"
	if d.chartType == "bar" {
		chartTypeStr = "Bar Chart"
	}

	summaryContent := fmt.Sprintf(`## Win Rate Trend Analysis

**Chart Type**: %s
**Period**: %s to %s
**Format**: %s
**Trend**: %s`,
		chartTypeStr,
		d.startDate.Format("2006-01-02"),
		d.endDate.Format("2006-01-02"),
		d.getFormatDisplayName(),
		analysis.Trend,
	)

	if analysis.TrendValue != 0 {
		// TrendValue is stored as decimal (0-1), multiply by 100 for percentage
		summaryContent += fmt.Sprintf(" (%+.1f%%)", analysis.TrendValue*100)
	}

	if analysis.Overall != nil {
		// WinRate is stored as decimal (0-1), multiply by 100 for percentage
		summaryContent += fmt.Sprintf(`
**Overall Win Rate**: %.1f%% (%d matches)`,
			analysis.Overall.WinRate*100,
			analysis.Overall.TotalMatches,
		)
	}

	return widget.NewRichTextFromMarkdown(summaryContent)
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

// getLimitedFormats queries the database for all draft/sealed format values.
func (d *WinRateDashboard) getLimitedFormats() []string {
	// Query for all distinct format values that contain "Draft" or "Sealed"
	filter := models.StatsFilter{}
	matches, err := d.service.GetMatches(d.ctx, filter)
	if err != nil {
		return []string{}
	}

	formatMap := make(map[string]bool)
	for _, match := range matches {
		formatLower := strings.ToLower(match.Format)
		if strings.Contains(formatLower, "draft") || strings.Contains(formatLower, "sealed") {
			formatMap[match.Format] = true
		}
	}

	formats := make([]string, 0, len(formatMap))
	for format := range formatMap {
		formats = append(formats, format)
	}
	return formats
}

package gui

import (
	"context"
	"fmt"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// formatData holds format statistics for display.
type formatData struct {
	format  string
	matches int
	winRate float64
}

// FormatDistributionDashboard manages the format distribution view.
type FormatDistributionDashboard struct {
	app       *App
	service   *storage.Service
	ctx       context.Context
	chartType string // "pie" or "bar"
	startDate *time.Time
	endDate   *time.Time
}

// NewFormatDistributionDashboard creates a new format distribution dashboard.
func NewFormatDistributionDashboard(app *App, service *storage.Service, ctx context.Context) *FormatDistributionDashboard {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	return &FormatDistributionDashboard{
		app:       app,
		service:   service,
		ctx:       ctx,
		chartType: "pie",
		startDate: &thirtyDaysAgo,
		endDate:   &now,
	}
}

// CreateView creates the format distribution view.
func (d *FormatDistributionDashboard) CreateView() fyne.CanvasObject {
	// Create filter controls
	filterControls := d.createFilterControls()

	// Create chart view
	chartView := d.createChartView()

	// Layout
	return container.NewBorder(
		container.NewPadded(filterControls),
		nil,
		nil,
		nil,
		container.NewScroll(container.NewPadded(chartView)),
	)
}

// createFilterControls creates the filter control panel.
func (d *FormatDistributionDashboard) createFilterControls() fyne.CanvasObject {
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
			d.refresh()
		},
	)
	dateRangeSelect.Selected = "Last 30 Days"

	// Chart type selector
	chartTypeLabel := widget.NewLabelWithStyle("Chart Type", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	chartTypeSelect := widget.NewSelect(
		[]string{"Pie Chart", "Bar Chart"},
		func(selected string) {
			if selected == "Pie Chart" {
				d.chartType = "pie"
			} else {
				d.chartType = "bar"
			}
			d.refresh()
		},
	)
	chartTypeSelect.Selected = "Pie Chart"

	// Refresh button
	refreshButton := widget.NewButton("Refresh", func() {
		d.refresh()
	})

	// Layout controls
	return container.NewVBox(
		container.NewGridWithColumns(2,
			container.NewVBox(dateRangeLabel, dateRangeSelect),
			container.NewVBox(chartTypeLabel, chartTypeSelect),
		),
		container.NewHBox(refreshButton),
		widget.NewSeparator(),
	)
}

// createChartView creates the chart visualization.
func (d *FormatDistributionDashboard) createChartView() fyne.CanvasObject {
	// Create filter for date range
	filter := storage.StatsFilter{
		StartDate: d.startDate,
		EndDate:   d.endDate,
	}

	// Get statistics by format
	statsByFormat, err := d.service.GetStatsByFormat(d.ctx, filter)
	if err != nil || len(statsByFormat) == 0 {
		return container.NewCenter(
			container.NewVBox(
				widget.NewLabel("No format data available"),
				widget.NewLabel(fmt.Sprintf("Error: %v", err)),
			),
		)
	}

	// Convert to data points and calculate total
	var formats []formatData
	totalMatches := 0

	for format, stats := range statsByFormat {
		if stats.TotalMatches > 0 {
			formats = append(formats, formatData{
				format:  format,
				matches: stats.TotalMatches,
				winRate: stats.WinRate * 100,
			})
			totalMatches += stats.TotalMatches
		}
	}

	// Sort by match count (descending)
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].matches > formats[j].matches
	})

	// Create data points for charts
	dataPoints := make([]charts.DataPoint, len(formats))
	for i, f := range formats {
		percentage := float64(f.matches) / float64(totalMatches) * 100
		label := fmt.Sprintf("%s (%.1f%%)", f.format, percentage)
		dataPoints[i] = charts.DataPoint{
			Label: label,
			Value: float64(f.matches),
		}
	}

	// Create chart config
	config := charts.DefaultFyneChartConfig()
	config.Title = d.getChartTitle()
	config.Width = 900
	config.Height = 600

	// Create chart based on type
	var chart fyne.CanvasObject
	if d.chartType == "pie" {
		chart = charts.CreateFynePieChartBreakdown(dataPoints, config)
	} else {
		chart = charts.CreateFyneBarChart(dataPoints, config)
	}

	// Create summary
	summary := d.createSummary(formats, totalMatches)

	// Layout
	return container.NewVBox(
		chart,
		widget.NewSeparator(),
		summary,
	)
}

// createSummary creates the summary information display.
func (d *FormatDistributionDashboard) createSummary(formats []formatData, totalMatches int) fyne.CanvasObject {
	// Build summary content
	summaryContent := fmt.Sprintf(`### Format Distribution Analysis

**Period**: %s
**Total Matches**: %d
**Unique Formats**: %d

`,
		d.getDateRangeDescription(),
		totalMatches,
		len(formats),
	)

	// Add breakdown by format
	summaryContent += "**Breakdown by Format**:\n\n"
	for _, f := range formats {
		percentage := float64(f.matches) / float64(totalMatches) * 100
		summaryContent += fmt.Sprintf("- **%s**: %d matches (%.1f%%) - %.1f%% win rate\n",
			f.format, f.matches, percentage, f.winRate)
	}

	return container.NewVBox(
		widget.NewRichTextFromMarkdown(summaryContent),
	)
}

// getChartTitle returns the appropriate chart title.
func (d *FormatDistributionDashboard) getChartTitle() string {
	title := "Format Distribution"

	switch {
	case d.startDate == nil && d.endDate == nil:
		title += " (All Time)"
	case d.startDate != nil && d.endDate != nil:
		days := int(d.endDate.Sub(*d.startDate).Hours() / 24)
		if days <= 7 {
			title += " (Last 7 Days)"
		} else if days <= 30 {
			title += " (Last 30 Days)"
		} else if days <= 90 {
			title += " (Last 90 Days)"
		}
	}

	return title
}

// getDateRangeDescription returns a human-readable date range description.
func (d *FormatDistributionDashboard) getDateRangeDescription() string {
	if d.startDate == nil || d.endDate == nil {
		return "All Time"
	}
	return fmt.Sprintf("%s to %s", d.startDate.Format("2006-01-02"), d.endDate.Format("2006-01-02"))
}

// refresh refreshes the chart view.
func (d *FormatDistributionDashboard) refresh() {
	// Recreate the entire Charts tab
	d.app.window.SetContent(container.NewAppTabs(
		container.NewTabItem("Statistics", d.app.createStatsView()),
		container.NewTabItem("Match History", d.app.createMatchesView()),
		container.NewTabItem("Charts", d.app.createChartsView()),
		container.NewTabItem("Settings", d.app.createSettingsView()),
	))
}

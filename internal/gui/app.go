package gui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// App represents the GUI application.
type App struct {
	app     fyne.App
	window  fyne.Window
	service *storage.Service
	ctx     context.Context
}

// NewApp creates a new GUI application.
func NewApp(service *storage.Service) *App {
	return &App{
		app:     app.New(),
		service: service,
		ctx:     context.Background(),
	}
}

// Run starts the GUI application.
func (a *App) Run() {
	a.window = a.app.NewWindow("MTGA Companion")
	a.window.Resize(fyne.NewSize(800, 600))

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Statistics", a.createStatsView()),
		container.NewTabItem("Recent Matches", a.createMatchesView()),
		container.NewTabItem("Charts", a.createChartsView()),
	)

	a.window.SetContent(tabs)
	a.window.ShowAndRun()
}

// createStatsView creates the statistics view.
func (a *App) createStatsView() fyne.CanvasObject {
	stats, err := a.service.GetStats(a.ctx, storage.StatsFilter{})
	if err != nil {
		return widget.NewLabel(fmt.Sprintf("Error: %v", err))
	}

	content := fmt.Sprintf(`Overall Statistics
==================

Matches: %d (%d-%d)
Win Rate: %.1f%%

Games: %d (%d-%d)
Game Win Rate: %.1f%%
`,
		stats.TotalMatches, stats.MatchesWon, stats.MatchesLost,
		stats.WinRate*100,
		stats.TotalGames, stats.GamesWon, stats.GamesLost,
		stats.GameWinRate*100,
	)

	label := widget.NewLabel(content)
	label.Wrapping = fyne.TextWrapWord

	refreshBtn := widget.NewButton("Refresh", func() {
		a.window.SetContent(container.NewAppTabs(
			container.NewTabItem("Statistics", a.createStatsView()),
			container.NewTabItem("Recent Matches", a.createMatchesView()),
			container.NewTabItem("Charts", a.createChartsView()),
		))
	})

	return container.NewBorder(nil, refreshBtn, nil, nil, container.NewScroll(label))
}

// createMatchesView creates the recent matches view.
func (a *App) createMatchesView() fyne.CanvasObject {
	filter := storage.StatsFilter{}
	matches, err := a.service.GetMatches(a.ctx, filter)
	if err != nil {
		return widget.NewLabel(fmt.Sprintf("Error: %v", err))
	}

	if len(matches) == 0 {
		return widget.NewLabel("No matches found")
	}

	// Limit to 20 most recent
	limit := 20
	if len(matches) < limit {
		limit = len(matches)
	}

	var content string
	content += "Recent Matches\n"
	content += "==============\n\n"

	for i := 0; i < limit; i++ {
		match := matches[i]
		result := "W"
		if match.Result == "loss" {
			result = "L"
		}

		content += fmt.Sprintf("%s | %s | %s | %d-%d\n",
			result,
			match.Timestamp.Format("2006-01-02 15:04"),
			match.EventName,
			match.PlayerWins,
			match.OpponentWins,
		)
	}

	label := widget.NewLabel(content)
	label.Wrapping = fyne.TextWrapWord

	return container.NewScroll(label)
}

// createChartsView creates the charts view.
func (a *App) createChartsView() fyne.CanvasObject {
	// Create sub-tabs for different chart types
	chartTabs := container.NewAppTabs(
		container.NewTabItem("Win Rate Trend", a.createWinRateTrendView()),
		container.NewTabItem("Result Breakdown", a.createResultBreakdownView()),
	)

	return chartTabs
}

// createWinRateTrendView creates the win rate trend chart view.
func (a *App) createWinRateTrendView() fyne.CanvasObject {
	// Date range selector
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Get trend data for last 30 days
	analysis, err := a.service.GetTrendAnalysis(a.ctx, thirtyDaysAgo, now, "weekly", nil)
	if err != nil || len(analysis.Periods) == 0 {
		return widget.NewLabel(fmt.Sprintf("No chart data available: %v", err))
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
	config.Title = "Win Rate Trend (Last 30 Days)"
	config.Width = 750
	config.Height = 450

	// Create chart
	chart := charts.CreateFyneLineChart(dataPoints, config)

	// Add summary info
	summaryText := fmt.Sprintf(`
Period: %s to %s
Trend: %s`,
		thirtyDaysAgo.Format("2006-01-02"),
		now.Format("2006-01-02"),
		analysis.Trend,
	)

	if analysis.TrendValue != 0 {
		summaryText += fmt.Sprintf(" (%.1f%%)", analysis.TrendValue)
	}

	if analysis.Overall != nil {
		summaryText += fmt.Sprintf(`
Overall Win Rate: %.1f%% (%d matches)`,
			analysis.Overall.WinRate,
			analysis.Overall.TotalMatches,
		)
	}

	summary := widget.NewLabel(summaryText)
	summary.Wrapping = fyne.TextWrapWord

	// Create chart type selector
	chartTypeSelect := widget.NewSelect([]string{"Line Chart", "Bar Chart"}, func(selected string) {
		// Recreate the entire Charts tab with the new chart type
		a.window.SetContent(container.NewAppTabs(
			container.NewTabItem("Statistics", a.createStatsView()),
			container.NewTabItem("Recent Matches", a.createMatchesView()),
			container.NewTabItem("Charts", a.createChartsView()),
		))
	})
	chartTypeSelect.Selected = "Line Chart"

	// Layout: selector at top, chart in middle, summary at bottom
	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Chart Type:"),
			chartTypeSelect,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			summary,
		),
		nil, nil,
		container.NewScroll(chart),
	)
}

// createResultBreakdownView creates the result breakdown chart view.
func (a *App) createResultBreakdownView() fyne.CanvasObject {
	// Date range (last 30 days)
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Get matches
	filter := storage.StatsFilter{
		StartDate: &thirtyDaysAgo,
		EndDate:   &now,
	}

	matches, err := a.service.GetMatches(a.ctx, filter)
	if err != nil || len(matches) == 0 {
		return widget.NewLabel(fmt.Sprintf("No match data available: %v", err))
	}

	// Calculate breakdowns
	winBreakdown := a.calculateBreakdown(matches, true)
	lossBreakdown := a.calculateBreakdown(matches, false)

	// Prepare data for charts
	winData := a.breakdownToDataPoints(winBreakdown)
	lossData := a.breakdownToDataPoints(lossBreakdown)

	// Create chart configs
	winConfig := charts.DefaultFyneChartConfig()
	winConfig.Title = "Wins Breakdown"
	winConfig.Width = 750
	winConfig.Height = 350

	lossConfig := charts.DefaultFyneChartConfig()
	lossConfig.Title = "Losses Breakdown"
	lossConfig.Width = 750
	lossConfig.Height = 350

	// Create charts
	winChart := charts.CreateFynePieChartBreakdown(winData, winConfig)
	lossChart := charts.CreateFynePieChartBreakdown(lossData, lossConfig)

	// Summary
	summaryText := fmt.Sprintf(`
Period: %s to %s
Total Matches: %d
Wins: %d | Losses: %d`,
		thirtyDaysAgo.Format("2006-01-02"),
		now.Format("2006-01-02"),
		len(matches),
		winBreakdown["Total"],
		lossBreakdown["Total"],
	)

	summary := widget.NewLabel(summaryText)
	summary.Wrapping = fyne.TextWrapWord

	// Layout: summary at top, both charts below
	return container.NewBorder(
		container.NewVBox(
			summary,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewScroll(
			container.NewVBox(
				winChart,
				widget.NewSeparator(),
				lossChart,
			),
		),
	)
}

// calculateBreakdown calculates match result breakdown.
func (a *App) calculateBreakdown(matches []*storage.Match, isWin bool) map[string]int {
	breakdown := map[string]int{
		"Normal":             0,
		"Concede":            0,
		"Timeout":            0,
		"Draw":               0,
		"Disconnect":         0,
		"OpponentConcede":    0,
		"OpponentTimeout":    0,
		"OpponentDisconnect": 0,
		"Other":              0,
		"Total":              0,
	}

	for _, match := range matches {
		if isWin && match.Result != "win" {
			continue
		}
		if !isWin && match.Result != "loss" {
			continue
		}

		breakdown["Total"]++

		if match.ResultReason == nil {
			breakdown["Normal"]++
			continue
		}

		reason := *match.ResultReason
		switch reason {
		case "ResultReason_Game":
			breakdown["Normal"]++
		case "ResultReason_Concede":
			breakdown["Concede"]++
		case "ResultReason_Timeout":
			breakdown["Timeout"]++
		case "ResultReason_Draw":
			breakdown["Draw"]++
		case "ResultReason_Disconnect":
			breakdown["Disconnect"]++
		case "ResultReason_OpponentConcede":
			breakdown["OpponentConcede"]++
		case "ResultReason_OpponentTimeout":
			breakdown["OpponentTimeout"]++
		case "ResultReason_OpponentDisconnect":
			breakdown["OpponentDisconnect"]++
		default:
			breakdown["Other"]++
		}
	}

	return breakdown
}

// breakdownToDataPoints converts breakdown map to DataPoint slice (only non-zero values).
func (a *App) breakdownToDataPoints(breakdown map[string]int) []charts.DataPoint {
	dataPoints := []charts.DataPoint{}

	// Define order of categories
	categories := []string{
		"Normal", "Concede", "Timeout", "Draw", "Disconnect",
		"OpponentConcede", "OpponentTimeout", "OpponentDisconnect", "Other",
	}

	// Add labels for display
	labels := map[string]string{
		"Normal":             "Normal",
		"Concede":            "Concede",
		"Timeout":            "Timeout",
		"Draw":               "Draw",
		"Disconnect":         "Disconnect",
		"OpponentConcede":    "Opp. Concede",
		"OpponentTimeout":    "Opp. Timeout",
		"OpponentDisconnect": "Opp. Disconnect",
		"Other":              "Other",
	}

	for _, cat := range categories {
		if count := breakdown[cat]; count > 0 {
			dataPoints = append(dataPoints, charts.DataPoint{
				Label: labels[cat],
				Value: float64(count),
			})
		}
	}

	return dataPoints
}

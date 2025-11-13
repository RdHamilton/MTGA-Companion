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
	state   *AppState // Application state manager
}

// NewApp creates a new GUI application.
func NewApp(service *storage.Service) *App {
	// Load persisted state
	state, err := LoadState()
	if err != nil {
		// If state load fails, use new state
		state = NewAppState()
	}

	return &App{
		app:     app.New(),
		service: service,
		ctx:     context.Background(),
		state:   state,
	}
}

// Run starts the GUI application.
func (a *App) Run() {
	a.window = a.app.NewWindow("MTGA Companion")

	// Apply saved window size or use defaults
	windowSize := a.state.GetWindowSize()
	if windowSize.Width > 0 && windowSize.Height > 0 {
		a.window.Resize(fyne.NewSize(float32(windowSize.Width), float32(windowSize.Height)))
	} else {
		a.window.Resize(fyne.NewSize(800, 600))
	}

	// Save state on window close
	a.window.SetCloseIntercept(func() {
		// Save current window size
		size := a.window.Content().Size()
		a.state.UpdateWindowSize(int(size.Width), int(size.Height))

		// Save state to disk
		_ = a.state.Save()

		// Close the window
		a.window.Close()
	})

	// Show onboarding for first-time users
	a.showOnboarding()

	// Show onboarding for first-time users
	a.showOnboarding()

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Statistics", a.createStatsView()),
		container.NewTabItem("Match History", a.createMatchesView()),
		container.NewTabItem("Charts", a.createChartsView()),
		container.NewTabItem("Settings", a.createSettingsView()),
	)

	// Setup keyboard shortcuts
	a.setupKeyboardShortcuts(tabs)

	a.window.SetContent(tabs)
	a.window.ShowAndRun()
}

// createStatsView creates the statistics view with material design principles.
func (a *App) createStatsView() fyne.CanvasObject {
	return a.WithLoading("Loading statistics...", func() (fyne.CanvasObject, error) {
		stats, err := a.service.GetStats(a.ctx, storage.StatsFilter{})
		if err != nil {
			return nil, err
		}

		if stats.TotalMatches == 0 {
			return a.NoDataView("No Statistics Available",
				"No match data has been collected yet."), nil
		}

		return a.buildStatsView(stats), nil
	})
}

// buildStatsView builds the statistics display from stats data.
func (a *App) buildStatsView(stats *storage.Statistics) fyne.CanvasObject {

	// Create rich text with markdown for better formatting
	content := fmt.Sprintf(`## Overall Statistics

### Match Statistics
- **Total Matches**: %d
- **Wins**: %d
- **Losses**: %d
- **Win Rate**: %.1f%%

### Game Statistics
- **Total Games**: %d
- **Wins**: %d
- **Losses**: %d
- **Game Win Rate**: %.1f%%
`,
		stats.TotalMatches, stats.MatchesWon, stats.MatchesLost,
		stats.WinRate*100,
		stats.TotalGames, stats.GamesWon, stats.GamesLost,
		stats.GameWinRate*100,
	)

	richText := widget.NewRichTextFromMarkdown(content)

	refreshBtn := widget.NewButton("Refresh Statistics", func() {
		a.window.SetContent(container.NewAppTabs(
			container.NewTabItem("Statistics", a.createStatsView()),
			container.NewTabItem("Match History", a.createMatchesView()),
			container.NewTabItem("Charts", a.createChartsView()),
			container.NewTabItem("Settings", a.createSettingsView()),
		))
	})

	// Layout: stats in center with padding, refresh button at bottom
	return container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			refreshBtn,
		),
		nil, nil,
		container.NewScroll(
			container.NewPadded(richText),
		),
	)
}

// createMatchesView creates the enhanced match history view.
func (a *App) createMatchesView() fyne.CanvasObject {
	viewer := NewMatchHistoryViewer(a, a.service, a.ctx)
	return viewer.CreateView()
}

// createChartsView creates the charts view.
func (a *App) createChartsView() fyne.CanvasObject {
	// Create sub-tabs for different chart types
	chartTabs := container.NewAppTabs(
		container.NewTabItem("Win Rate Trend", a.createWinRateTrendView()),
		container.NewTabItem("Format Distribution", a.createFormatDistributionView()),
		container.NewTabItem("Deck Performance", a.createDeckPerformanceView()),
		container.NewTabItem("Result Breakdown", a.createResultBreakdownView()),
		container.NewTabItem("Rank Progression", a.createRankProgressionView()),
	)

	return chartTabs
}

// createWinRateTrendView creates the win rate trend chart view with enhanced filtering.
func (a *App) createWinRateTrendView() fyne.CanvasObject {
	dashboard := NewWinRateDashboard(a, a.service, a.ctx)
	return dashboard.CreateView()
}

// createFormatDistributionView creates the format distribution chart view.
func (a *App) createFormatDistributionView() fyne.CanvasObject {
	dashboard := NewFormatDistributionDashboard(a, a.service, a.ctx)
	return dashboard.CreateView()
}

// createDeckPerformanceView creates the deck performance comparison chart view.
func (a *App) createDeckPerformanceView() fyne.CanvasObject {
	dashboard := NewDeckPerformanceDashboard(a, a.service, a.ctx)
	return dashboard.CreateView()
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
	if err != nil {
		return a.ErrorView("Error Loading Match Data", err, a.createResultBreakdownView)
	}

	if len(matches) == 0 {
		return a.NoDataView("No Match Data Available",
			"No matches found for the selected time period.")
	}

	// Calculate breakdowns
	winBreakdown := a.calculateBreakdown(matches, true)
	lossBreakdown := a.calculateBreakdown(matches, false)

	// Prepare data for charts
	winData := a.breakdownToDataPoints(winBreakdown)
	lossData := a.breakdownToDataPoints(lossBreakdown)

	// Create chart configs with larger size
	winConfig := charts.DefaultFyneChartConfig()
	winConfig.Title = "Wins Breakdown"
	winConfig.Width = 900
	winConfig.Height = 450

	lossConfig := charts.DefaultFyneChartConfig()
	lossConfig.Title = "Losses Breakdown"
	lossConfig.Width = 900
	lossConfig.Height = 450

	// Create charts
	winChart := charts.CreateFynePieChartBreakdown(winData, winConfig)
	lossChart := charts.CreateFynePieChartBreakdown(lossData, lossConfig)

	// Summary with markdown formatting
	summaryContent := fmt.Sprintf(`### Result Breakdown Analysis

**Period**: %s to %s
**Total Matches**: %d
**Wins**: %d | **Losses**: %d`,
		thirtyDaysAgo.Format("2006-01-02"),
		now.Format("2006-01-02"),
		len(matches),
		winBreakdown["Total"],
		lossBreakdown["Total"],
	)

	summary := widget.NewRichTextFromMarkdown(summaryContent)

	// Layout: summary at top, both charts below with padding
	return container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				summary,
				widget.NewSeparator(),
			),
		),
		nil, nil, nil,
		container.NewScroll(
			container.NewPadded(
				container.NewVBox(
					winChart,
					widget.NewSeparator(),
					lossChart,
				),
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

// createRankProgressionView creates the rank progression chart view.
func (a *App) createRankProgressionView() fyne.CanvasObject {
	// Date range (last 30 days)
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Get rank progression timeline for constructed
	timeline, err := a.service.GetRankProgressionTimeline(a.ctx, "constructed", &thirtyDaysAgo, &now, storage.PeriodWeekly)
	if err != nil {
		return a.ErrorView("Error Loading Rank Data", err, a.createRankProgressionView)
	}

	if len(timeline.Entries) == 0 {
		return a.NoDataView("No Rank Data Available",
			"No ranked matches found for the selected time period.")
	}

	// Convert timeline entries to chart data points
	dataPoints := make([]charts.DataPoint, len(timeline.Entries))
	for i, entry := range timeline.Entries {
		dataPoints[i] = charts.DataPoint{
			Label: entry.Date,
			Value: a.rankToNumericValue(entry.RankClass, entry.RankLevel),
		}
	}

	// Create chart config
	config := charts.DefaultFyneChartConfig()
	config.Title = "Rank Progression - Constructed (Last 30 Days)"
	config.Width = 750
	config.Height = 450

	// Create chart
	chart := charts.CreateFyneLineChart(dataPoints, config)

	// Summary with markdown formatting
	summaryContent := fmt.Sprintf(`### Rank Progression Analysis

**Period**: %s to %s
**Start Rank**: %s
**End Rank**: %s
**Highest Rank**: %s
**Lowest Rank**: %s
**Total Changes**: %d
**Milestones**: %d`,
		thirtyDaysAgo.Format("2006-01-02"),
		now.Format("2006-01-02"),
		timeline.StartRank,
		timeline.EndRank,
		timeline.HighestRank,
		timeline.LowestRank,
		timeline.TotalChanges,
		timeline.Milestones,
	)

	summary := widget.NewRichTextFromMarkdown(summaryContent)

	// Format selector
	formatSelect := widget.NewSelect([]string{"Constructed", "Limited"}, func(selected string) {
		// Recreate the entire Charts tab with the new format
		a.window.SetContent(container.NewAppTabs(
			container.NewTabItem("Statistics", a.createStatsView()),
			container.NewTabItem("Match History", a.createMatchesView()),
			container.NewTabItem("Charts", a.createChartsView()),
			container.NewTabItem("Settings", a.createSettingsView()),
		))
	})
	formatSelect.Selected = "Constructed"

	// Layout: selector at top, chart in middle, summary at bottom with padding
	return container.NewBorder(
		container.NewPadded(
			container.NewVBox(
				widget.NewLabelWithStyle("Format", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				formatSelect,
				widget.NewSeparator(),
			),
		),
		container.NewPadded(
			container.NewVBox(
				widget.NewSeparator(),
				summary,
			),
		),
		nil, nil,
		container.NewScroll(container.NewPadded(chart)),
	)
}

// rankToNumericValue converts rank class and level to a numeric value for charting.
func (a *App) rankToNumericValue(rankClass *string, rankLevel *int) float64 {
	if rankClass == nil {
		return 0
	}

	// Map rank classes to base values
	rankClassValues := map[string]float64{
		"Bronze":   0,
		"Silver":   4,
		"Gold":     8,
		"Platinum": 12,
		"Diamond":  16,
		"Mythic":   20,
	}

	baseValue, ok := rankClassValues[*rankClass]
	if !ok {
		return 0
	}

	// Add level offset (higher level = higher value)
	// Rank levels go from 4 (lowest) to 1 (highest)
	if rankLevel != nil && *rankLevel >= 1 && *rankLevel <= 4 {
		baseValue += float64(5 - *rankLevel) // Convert so level 4=1, level 1=4
	} else if *rankClass == "Mythic" {
		// Mythic has no levels, just use base value
		baseValue += 4 // Treat Mythic as highest
	}

	return baseValue
}

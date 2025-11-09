package gui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

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

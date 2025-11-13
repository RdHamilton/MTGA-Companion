package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
)

// setupKeyboardShortcuts configures keyboard shortcuts for the application.
// Call this after the window is created in Run().
func (a *App) setupKeyboardShortcuts(tabs *container.AppTabs) {
	if a.window == nil {
		return
	}

	// Add keyboard shortcuts to the window canvas
	canvas := a.window.Canvas()

	// Refresh shortcut: Ctrl/Cmd + R
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: fyne.KeyModifierControl, // Works for both Ctrl (Windows/Linux) and Cmd (macOS)
	}, func(shortcut fyne.Shortcut) {
		a.refreshActiveTab(tabs)
	})

	// Help shortcut: Ctrl/Cmd + H
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyH,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		a.ShowHelp()
	})

	// Settings shortcut: Ctrl/Cmd + Comma
	canvas.AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyComma,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		tabs.SelectIndex(3) // Settings is the 4th tab (index 3)
	})
}

// refreshActiveTab refreshes the currently active tab.
func (a *App) refreshActiveTab(tabs *container.AppTabs) {
	currentIndex := tabs.SelectedIndex()

	// Recreate the current tab's content
	switch currentIndex {
	case 0: // Statistics
		tabs.Items[0].Content = a.createStatsView()
	case 1: // Match History
		tabs.Items[1].Content = a.createMatchesView()
	case 2: // Charts
		tabs.Items[2].Content = a.createChartsView()
	case 3: // Settings
		tabs.Items[3].Content = a.createSettingsView()
	}

	tabs.Refresh()
}

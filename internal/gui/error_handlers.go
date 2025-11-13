package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ErrorView creates a user-friendly error display with retry functionality.
func (a *App) ErrorView(title string, err error, retryFunc func() fyne.CanvasObject) fyne.CanvasObject {
	// Error title
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Error message with better formatting
	var errorMsg string
	if err != nil {
		errorMsg = fmt.Sprintf("Details: %v", err)
	} else {
		errorMsg = "No data available"
	}
	errorLabel := widget.NewLabel(errorMsg)
	errorLabel.Wrapping = fyne.TextWrapWord

	// Helpful guidance
	guidanceText := widget.NewRichTextFromMarkdown(`### Possible Solutions:

- Check that the MTGA log file path is correctly configured in Settings
- Ensure you have played at least one match
- Try clicking the Refresh button below
- If the problem persists, please report an issue`)

	// Retry button (if retry function provided)
	var buttons []fyne.CanvasObject
	if retryFunc != nil {
		retryButton := widget.NewButton("Refresh", func() {
			// Replace error view with retry result
			newContent := retryFunc()
			a.refreshCurrentTab(newContent)
		})
		buttons = append(buttons, retryButton)
	}

	// Help button
	helpButton := widget.NewButton("Get Help", func() {
		dialog.ShowInformation("Need Help?",
			fmt.Sprintf("For support, please visit:\n\n• Documentation: %s\n• Report Issue: %s",
				DocsURL, IssuesURL), a.window)
	})
	buttons = append(buttons, helpButton)

	// Layout
	return container.NewCenter(
		container.NewVBox(
			titleLabel,
			widget.NewSeparator(),
			errorLabel,
			widget.NewSeparator(),
			guidanceText,
			widget.NewSeparator(),
			container.NewHBox(buttons...),
		),
	)
}

// NoDataView creates a friendly message when no data is available.
func (a *App) NoDataView(title string, message string) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord

	guidanceText := widget.NewRichTextFromMarkdown(`### Getting Started:

1. **Configure Settings**: Ensure your MTGA log path is set correctly
2. **Play Matches**: Data is collected from MTGA match logs
3. **Wait for Updates**: The log file is monitored automatically

Once you've played some matches, data will appear here!`)

	return container.NewCenter(
		container.NewVBox(
			titleLabel,
			widget.NewSeparator(),
			messageLabel,
			widget.NewSeparator(),
			guidanceText,
		),
	)
}

// refreshCurrentTab is a helper to refresh the current tab content
// Note: This is a simplified version - full implementation would need tab tracking
func (a *App) refreshCurrentTab(newContent fyne.CanvasObject) {
	// Recreate the main tabs
	mainTabs := container.NewAppTabs(
		container.NewTabItem("Statistics", a.createStatsView()),
		container.NewTabItem("Match History", a.createMatchesView()),
		container.NewTabItem("Charts", a.createChartsView()),
		container.NewTabItem("Settings", a.createSettingsView()),
	)

	a.window.SetContent(mainTabs)
}

// ShowErrorDialog displays an error dialog with helpful information.
func (a *App) ShowErrorDialog(title string, err error) {
	message := fmt.Sprintf("%v\n\nIf this problem persists, please report it:\n%s", err, IssuesURL)
	dialog.ShowError(fmt.Errorf("%s", message), a.window)
}

// ShowSuccessDialog displays a success message.
func (a *App) ShowSuccessDialog(title string, message string) {
	dialog.ShowInformation(title, message, a.window)
}

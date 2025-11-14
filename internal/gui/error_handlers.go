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
	// Error title with icon
	titleLabel := widget.NewLabelWithStyle("⚠️  "+title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// User-friendly error message
	var errorMsg string
	var guidanceMarkdown string

	if err != nil {
		// Translate technical errors to user-friendly messages
		errorStr := err.Error()
		errorMsg = a.translateError(errorStr)
		guidanceMarkdown = a.getContextualGuidance(title, errorStr)
	} else {
		errorMsg = "An unexpected issue occurred. Please try refreshing."
		guidanceMarkdown = `### What You Can Do:

1. Click the **Refresh** button below to try again
2. Check your **Settings** to ensure MTGA is configured correctly
3. If the issue persists, click **Get Help** for support`
	}

	errorLabel := widget.NewLabel(errorMsg)
	errorLabel.Wrapping = fyne.TextWrapWord

	// Context-specific guidance
	guidanceText := widget.NewRichTextFromMarkdown(guidanceMarkdown)

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
	// Title with icon
	titleLabel := widget.NewLabelWithStyle("📊  "+title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Message
	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord
	messageLabel.Alignment = fyne.TextAlignCenter

	// Context-specific guidance based on title
	var guidanceMarkdown string
	if contains(title, "Statistics") || contains(title, "Match") {
		guidanceMarkdown = `### Let's Get Started! 🎮

**No match data yet.** Here's how to start tracking:

1. **Verify Settings** ⚙️
   - Open the Settings tab
   - Check that your MTGA log path is correct
   - Default paths work for most users

2. **Enable Detailed Logging** 📝
   - In MTGA: Options → View Account → Detailed Logs
   - This allows the companion app to read your matches

3. **Play Some Matches** 🃏
   - Any format works: Constructed, Limited, Draft
   - Data updates automatically after each match

4. **Return Here** 📈
   - Your statistics and match history will appear
   - Charts and analytics unlock as you play more

**Tip**: You need at least one match for data to appear!`
	} else if contains(title, "Chart") || contains(title, "Trend") || contains(title, "Progression") {
		guidanceMarkdown = `### Need More Data 📊

**This chart needs more match data to display.**

**Quick Fixes:**

1. **Broaden Your Filters** 🔍
   - Try selecting "All Time" for date range
   - Remove format or event filters
   - Check if you have matches in the selected period

2. **Play More Matches** 🎮
   - Some charts need minimum data to show trends
   - Win Rate Trend: Needs 2+ matches
   - Rank Progression: Needs ranked matches
   - Deck Performance: Needs matches with deck data

3. **Check Settings** ⚙️
   - Ensure MTGA log path is configured
   - Verify detailed logging is enabled in MTGA

**Tip**: The more you play, the richer your analytics become!`
	} else {
		guidanceMarkdown = `### Getting Started 🚀

**No data available yet.** Follow these steps:

1. **Configure Settings** ⚙️
   - Ensure your MTGA log path is set correctly
   - Enable any required features

2. **Play Matches** 🎮
   - Data is collected from MTGA match logs
   - All game modes are supported

3. **Monitor Progress** 📈
   - The log file is watched automatically
   - Data updates after each match

Once you've played some matches, data will appear here!`
	}

	guidanceText := widget.NewRichTextFromMarkdown(guidanceMarkdown)

	return container.NewCenter(
		container.NewPadded(
			container.NewVBox(
				titleLabel,
				widget.NewSeparator(),
				messageLabel,
				widget.NewSeparator(),
				guidanceText,
			),
		),
	)
}

// refreshCurrentTab is a helper to refresh the current tab content
// Note: This is a simplified version - full implementation would need tab tracking
func (a *App) refreshCurrentTab(newContent fyne.CanvasObject) {
	// Recreate the main tabs
	mainTabs := container.NewAppTabs(
		container.NewTabItem("Match History", a.createMatchesView()),
		container.NewTabItem("Charts", a.createChartsView()),
		container.NewTabItem("Settings", a.createSettingsView()),
	)

	// Recreate footer
	footer := a.createFooter()
	mainContent := container.NewBorder(
		nil,
		footer,
		nil,
		nil,
		mainTabs,
	)

	a.window.SetContent(mainContent)
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

// translateError converts technical error messages to user-friendly text.
func (a *App) translateError(errorStr string) string {
	// Common error patterns and their user-friendly translations
	if errorStr == "" {
		return "Something went wrong, but no specific error was reported."
	}

	// Database errors
	if contains(errorStr, "database") || contains(errorStr, "sql") {
		return "There was a problem accessing your match data. Your database may be locked or corrupted."
	}

	// File system errors
	if contains(errorStr, "no such file") || contains(errorStr, "cannot find") {
		return "The MTGA log file couldn't be found. Please check your Settings."
	}

	if contains(errorStr, "permission denied") {
		return "Permission was denied when trying to access a file. Please check file permissions."
	}

	// Network/API errors
	if contains(errorStr, "connection") || contains(errorStr, "timeout") {
		return "A connection problem occurred. Please check your internet connection and try again."
	}

	// Data parsing errors
	if contains(errorStr, "parse") || contains(errorStr, "unmarshal") || contains(errorStr, "json") {
		return "The data format was unexpected. Your MTGA log file may be corrupted or incomplete."
	}

	// Generic error - show simplified version
	if len(errorStr) > 100 {
		return "An error occurred: " + errorStr[:97] + "..."
	}

	return "An error occurred: " + errorStr
}

// getContextualGuidance provides context-specific guidance based on the error.
func (a *App) getContextualGuidance(title string, errorStr string) string {
	// Database-related errors
	if contains(errorStr, "database") || contains(errorStr, "sql") {
		return `### How to Fix This:

1. **Close MTGA** if it's running
2. **Restart MTGA Companion** to unlock the database
3. If the problem persists, your database may need repair:
   - Go to Settings
   - Use the backup/restore feature to create a backup
   - Consider restoring from a recent backup

**Need More Help?** Click "Get Help" below.`
	}

	// File not found errors
	if contains(errorStr, "no such file") || contains(errorStr, "cannot find") {
		return `### How to Fix This:

1. **Open Settings** and verify the MTGA log file path
2. **Default locations**:
   - macOS: ~/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log
   - Windows: C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log
3. **Enable Detailed Logging** in MTGA (Options → View Account → Detailed Logs)
4. **Play a match** in MTGA to generate log data

**Need More Help?** Click "Get Help" below.`
	}

	// Network/connection errors
	if contains(errorStr, "connection") || contains(errorStr, "timeout") {
		return `### How to Fix This:

1. **Check your internet connection**
2. **Try again** in a few moments
3. If offline features are affected, some data may be unavailable
4. Check if any firewall or antivirus is blocking the application

**Need More Help?** Click "Get Help" below.`
	}

	// Chart/data loading specific
	if contains(title, "Chart") || contains(title, "Data") || contains(title, "Trend") {
		return `### How to Fix This:

1. **Click Refresh** to try loading the data again
2. **Check your filters** - try selecting "All Time" or broader date ranges
3. **Play more matches** - some charts need minimum data to display
4. **Verify Settings** - ensure MTGA log path is correct

**Need More Help?** Click "Get Help" below.`
	}

	// Generic guidance
	return `### What You Can Do:

1. **Click Refresh** to try again
2. **Check Settings** to ensure everything is configured correctly
3. **Review the error details** above for specific information
4. **Play a match** in MTGA to generate fresh data
5. **Restart the application** if the problem persists

**Need More Help?** Click "Get Help" below.`
}

// contains is a helper function to check if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsMiddle(s, substr))))
}

// containsMiddle checks if substr appears in the middle of s.
func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

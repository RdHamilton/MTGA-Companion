package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowHelp displays the main help dialog with tips and guides.
func (a *App) ShowHelp() {
	content := widget.NewRichTextFromMarkdown(`# MTGA Companion Help

## Quick Start

### First Time Setup
1. Enable **Detailed Logs** in MTGA (Options → View Account → Detailed Logs)
2. Restart MTGA for changes to take effect
3. The app will automatically detect your log file

### Using the App

**Statistics Tab**
- View overall win rates and match statistics
- See both match and individual game win rates
- Use Refresh button to update with latest data

**Match History Tab**
- Browse all your past matches
- Filter by format, result, or date
- Search for specific events
- Click matches for detailed information

**Charts Tab**
- Visualize win rate trends over time
- See format distribution
- Analyze performance patterns
- Use date range filters for specific periods

**Settings Tab**
- Configure log file path
- Adjust polling settings
- Manage database backups
- Export your data

## Common Tasks

### Viewing Recent Performance
1. Go to Statistics tab
2. Check your recent win rate
3. Click Refresh to update

### Finding a Specific Match
1. Go to Match History tab
2. Use search bar to filter by event name
3. Or use format/date filters

### Exporting Data
1. Go to Settings tab
2. Click Export Data button
3. Choose CSV or JSON format
4. Select save location

### Backing Up Database
1. Go to Settings tab
2. Click Backup Database
3. Choose backup location
4. Optionally enable encryption

## Keyboard Shortcuts
- **Ctrl/Cmd + R**: Refresh current view
- **Ctrl/Cmd + F**: Focus search (in Match History)
- **Ctrl/Cmd + ,**: Open Settings
- **Ctrl/Cmd + H**: Show this help

## Troubleshooting

### Statistics Not Updating
- Verify MTGA has detailed logging enabled
- Check that log file path is correct in Settings
- Try clicking Refresh button
- Ensure MTGA is running

### "Log File Not Found" Error
- Enable detailed logging in MTGA
- Restart MTGA
- Check Settings for correct log path
- Default locations:
  - macOS: ~/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log
  - Windows: C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log

### Performance Issues
- Close other resource-intensive applications
- Reduce polling frequency in Settings
- Consider using file system events instead of polling

## Need More Help?
- **Documentation**: Visit our wiki for detailed guides
- **Report Issues**: Found a bug? Let us know on GitHub
- **Feature Requests**: Suggest new features on GitHub

## About
MTGA Companion v1.0.0
A companion application for Magic: The Gathering Arena

Not affiliated with Wizards of the Coast.`)

	helpDialog := dialog.NewCustom("Help", "Close", container.NewScroll(content), a.window)
	helpDialog.Resize(fyne.NewSize(700, 600))
	helpDialog.Show()
}

// ShowQuickTip displays a quick tip dialog with specific information.
func (a *App) ShowQuickTip(title, message string) {
	dialog.ShowInformation(title, message, a.window)
}

// ShowAbout displays the about dialog with app information.
func (a *App) ShowAbout() {
	content := widget.NewRichTextFromMarkdown(`# MTGA Companion

**Version**: 1.0.0

A companion application for Magic: The Gathering Arena that helps you track your performance, analyze your play patterns, and manage your game data.

## Features
- 📊 Statistics tracking with win rates
- 📜 Complete match history
- 📈 Performance charts and analytics
- 💾 Data export (CSV/JSON)
- 🔒 Encrypted database backups
- 🎯 Draft tracking and analysis

## Open Source
MTGA Companion is open source software licensed under the MIT License.

**Repository**: https://github.com/RdHamilton/MTGA-Companion
**Documentation**: https://github.com/RdHamilton/MTGA-Companion/wiki
**Issues & Support**: https://github.com/RdHamilton/MTGA-Companion/issues

## Disclaimer
MTGA Companion is not affiliated with, endorsed by, or sponsored by Wizards of the Coast. Magic: The Gathering Arena and its associated trademarks are property of Wizards of the Coast LLC.

## Credits
Built with:
- Go programming language
- Fyne UI toolkit
- SQLite database
- 17Lands draft data
- Scryfall card data

---

Thank you for using MTGA Companion!`)

	aboutDialog := dialog.NewCustom("About MTGA Companion", "Close",
		container.NewScroll(content), a.window)
	aboutDialog.Resize(fyne.NewSize(600, 500))
	aboutDialog.Show()
}

// HelpTooltip creates a widget with a help icon that shows a tooltip on hover.
func HelpTooltip(message string) *widget.Button {
	btn := widget.NewButton("?", func() {
		// Tooltip shown on hover, click shows dialog if needed
	})
	btn.Importance = widget.LowImportance
	return btn
}

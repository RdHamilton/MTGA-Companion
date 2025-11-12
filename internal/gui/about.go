package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// showAboutDialog displays the About dialog with app information, credits, and links.
func (a *App) showAboutDialog() {
	// App title and version
	titleText := widget.NewRichTextFromMarkdown(fmt.Sprintf("# %s\n## Version %s", AppName, Version))

	// Description
	descText := widget.NewLabel("A companion application for Magic: The Gathering Arena\nwith statistics tracking, draft overlay, and performance analytics.")

	// Copyright and license
	copyrightText := widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s**\n\n%s", Copyright, License))

	// Links section
	linksContent := fmt.Sprintf(`### Links

- **GitHub Repository**: <%s>
- **Documentation**: <%s>
- **Report Issues**: <%s>`,
		GitHubURL,
		DocsURL,
		IssuesURL,
	)
	linksText := widget.NewRichTextFromMarkdown(linksContent)

	// Acknowledgments section
	acknowledgementsContent := `### Acknowledgments

**MTGA Companion** is built with and relies on:

- **[Fyne](https://fyne.io/)** - Beautiful cross-platform GUI framework
- **[17Lands](https://www.17lands.com/)** - Draft ratings and card data
- **[Scryfall](https://scryfall.com/)** - Comprehensive MTG card database
- **[golang-migrate](https://github.com/golang-migrate/migrate)** - Database migrations
- **[SQLite](https://www.sqlite.org/)** - Embedded database engine

**Special Thanks** to the Magic: The Gathering Arena community for their support and feedback.`

	acknowledgementsText := widget.NewRichTextFromMarkdown(acknowledgementsContent)

	// Create scrollable content
	content := container.NewVBox(
		titleText,
		widget.NewSeparator(),
		descText,
		widget.NewSeparator(),
		copyrightText,
		widget.NewSeparator(),
		linksText,
		widget.NewSeparator(),
		acknowledgementsText,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(600, 500))

	// Create custom dialog
	customDialog := dialog.NewCustom("About MTGA Companion", "Close", scrollContent, a.window)
	customDialog.Resize(fyne.NewSize(650, 550))
	customDialog.Show()
}

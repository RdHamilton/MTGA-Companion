package gui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

const (
	matchesPerPage = 50
)

// MatchHistoryViewer represents the enhanced match history viewer.
type MatchHistoryViewer struct {
	app     *App
	service *storage.Service
	ctx     context.Context

	// UI components
	matchList      *widget.List
	searchEntry    *widget.Entry
	formatSelect   *widget.Select
	resultSelect   *widget.Select
	opponentSelect *widget.Select
	startDateEntry *widget.Entry
	endDateEntry   *widget.Entry
	statusLabel    *widget.Label
	pageLabel      *widget.Label

	// Data
	allMatches      []*storage.Match
	filteredMatches []*storage.Match
	currentPage     int
	totalPages      int
}

// NewMatchHistoryViewer creates a new enhanced match history viewer.
func NewMatchHistoryViewer(app *App, service *storage.Service, ctx context.Context) *MatchHistoryViewer {
	viewer := &MatchHistoryViewer{
		app:         app,
		service:     service,
		ctx:         ctx,
		currentPage: 0,
	}

	viewer.loadMatches()
	return viewer
}

// CreateView creates the match history view with filters and controls.
func (v *MatchHistoryViewer) CreateView() fyne.CanvasObject {
	// Search bar
	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("Search by event name or opponent...")
	v.searchEntry.OnChanged = func(text string) {
		v.filterMatches()
		v.currentPage = 0
		v.refreshList()
	}

	// Format filter
	formatOptions := []string{"All Formats", "Constructed", "Limited", "Draft", "Sealed"}
	v.formatSelect = widget.NewSelect(formatOptions, func(selected string) {
		v.filterMatches()
		v.currentPage = 0
		v.refreshList()
	})
	v.formatSelect.Selected = "All Formats"

	// Result filter
	resultOptions := []string{"All Results", "win", "loss"}
	v.resultSelect = widget.NewSelect(resultOptions, func(selected string) {
		v.filterMatches()
		v.currentPage = 0
		v.refreshList()
	})
	v.resultSelect.Selected = "All Results"

	// Opponent filter - will be populated after matches are loaded
	v.opponentSelect = widget.NewSelect([]string{"All Opponents"}, func(selected string) {
		v.filterMatches()
		v.currentPage = 0
		v.refreshList()
	})
	v.opponentSelect.Selected = "All Opponents"

	// Date range filters with calendar pickers
	v.startDateEntry = widget.NewEntry()
	v.startDateEntry.SetPlaceHolder("YYYY-MM-DD")
	v.endDateEntry = widget.NewEntry()
	v.endDateEntry.SetPlaceHolder("YYYY-MM-DD")

	// Start date picker button
	startPickerBtn := widget.NewButton("📅", func() {
		v.showDatePickerDialog("Select Start Date", v.startDateEntry)
	})

	// End date picker button
	endPickerBtn := widget.NewButton("📅", func() {
		v.showDatePickerDialog("Select End Date", v.endDateEntry)
	})

	dateApplyBtn := widget.NewButton("Apply Date Filter", func() {
		v.filterMatches()
		v.currentPage = 0
		v.refreshList()
	})

	dateClearBtn := widget.NewButton("Clear Dates", func() {
		v.startDateEntry.SetText("")
		v.endDateEntry.SetText("")
		v.filterMatches()
		v.currentPage = 0
		v.refreshList()
	})

	// Export button
	exportBtn := widget.NewButton("Export Matches", func() {
		v.exportMatches()
	})

	// Status label
	v.statusLabel = widget.NewLabel("")
	v.updateStatusLabel()

	// Create column headers with fixed widths for alignment
	resultHeader := widget.NewLabelWithStyle("Result", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dateHeader := widget.NewLabelWithStyle("Date/Time", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	eventHeader := widget.NewLabelWithStyle("Event", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	scoreHeader := widget.NewLabelWithStyle("Score", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	opponentHeader := widget.NewLabelWithStyle("Opponent", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	headers := container.New(
		layout.NewGridLayout(5),
		resultHeader,
		dateHeader,
		eventHeader,
		scoreHeader,
		opponentHeader,
	)

	// Create match list with matching grid layout
	v.matchList = widget.NewList(
		func() int {
			return v.getPageMatchCount()
		},
		func() fyne.CanvasObject {
			return container.New(
				layout.NewGridLayout(5),
				widget.NewLabel("W"),
				widget.NewLabel("2025-11-10 15:04"),
				widget.NewLabel("Event Name Here"),
				widget.NewLabel("2-1"),
				widget.NewLabel("Opponent"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			match := v.getPageMatch(i)
			if match == nil {
				return
			}

			box := o.(*fyne.Container)

			// Result
			result := "W"
			if match.Result == "loss" {
				result = "L"
			}
			box.Objects[0].(*widget.Label).SetText(result)

			// Timestamp
			box.Objects[1].(*widget.Label).SetText(match.Timestamp.Format("2006-01-02 15:04"))

			// Event name (truncated if needed)
			eventName := match.EventName
			if len(eventName) > 30 {
				eventName = eventName[:27] + "..."
			}
			box.Objects[2].(*widget.Label).SetText(eventName)

			// Score
			score := fmt.Sprintf("%d-%d", match.PlayerWins, match.OpponentWins)
			box.Objects[3].(*widget.Label).SetText(score)

			// Opponent
			opponent := "Unknown"
			if match.OpponentName != nil && *match.OpponentName != "" {
				opponent = *match.OpponentName
				if len(opponent) > 20 {
					opponent = opponent[:17] + "..."
				}
			}
			box.Objects[4].(*widget.Label).SetText(opponent)
		},
	)

	v.matchList.OnSelected = func(id widget.ListItemID) {
		match := v.getPageMatch(id)
		if match != nil {
			v.showMatchDetail(match)
		}
		v.matchList.UnselectAll()
	}

	// Pagination controls
	v.pageLabel = widget.NewLabel("Page 1 of 1")

	prevBtn := widget.NewButton("< Previous", func() {
		if v.currentPage > 0 {
			v.currentPage--
			v.refreshList()
		}
	})

	nextBtn := widget.NewButton("Next >", func() {
		if v.currentPage < v.totalPages-1 {
			v.currentPage++
			v.refreshList()
		}
	})

	pageControls := container.NewHBox(
		prevBtn,
		v.pageLabel,
		nextBtn,
	)

	// Layout: filters at top, list in middle, pagination at bottom
	// Combine date entries with calendar picker buttons
	startDateBox := container.NewBorder(nil, nil, nil, startPickerBtn, v.startDateEntry)
	endDateBox := container.NewBorder(nil, nil, nil, endPickerBtn, v.endDateEntry)

	// Use a more compact 4-column layout for better control
	filterGrid := container.New(
		layout.NewGridLayout(4),
		// Row 1
		widget.NewLabel("Format:"),
		v.formatSelect,
		widget.NewLabel("Result:"),
		v.resultSelect,
		// Row 2
		widget.NewLabel("Opponent:"),
		v.opponentSelect,
		widget.NewLabel("From:"),
		startDateBox,
		// Row 3 - need fillers for alignment
		widget.NewLabel(""),
		widget.NewLabel(""),
		widget.NewLabel("To:"),
		endDateBox,
	)

	filterButtons := container.NewHBox(
		exportBtn,
		layout.NewSpacer(),
		dateApplyBtn,
		dateClearBtn,
	)

	filtersSection := container.NewVBox(
		widget.NewLabel("Match History"),
		widget.NewSeparator(),
		v.searchEntry,
		filterGrid,
		filterButtons,
		v.statusLabel,
		widget.NewSeparator(),
	)

	// Combine headers and list
	listWithHeaders := container.NewBorder(
		container.NewVBox(headers, widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewScroll(v.matchList),
	)

	// Now that opponent select is created, populate it with actual opponents
	v.updateOpponentFilter()

	return container.NewBorder(
		container.NewPadded(filtersSection),
		container.NewPadded(
			container.NewVBox(
				widget.NewSeparator(),
				pageControls,
			),
		),
		nil,
		nil,
		container.NewPadded(listWithHeaders),
	)
}

// loadMatches loads all matches from the database.
func (v *MatchHistoryViewer) loadMatches() {
	filter := storage.StatsFilter{}
	matches, err := v.service.GetMatches(v.ctx, filter)
	if err != nil {
		v.allMatches = []*storage.Match{}
		v.filteredMatches = []*storage.Match{}
		return
	}

	v.allMatches = matches
	v.filteredMatches = matches

	// Update opponent dropdown with unique opponents
	v.updateOpponentFilter()

	v.calculatePagination()
	v.updateStatusLabel()
}

// filterMatches applies current filters to match list.
func (v *MatchHistoryViewer) filterMatches() {
	v.filteredMatches = []*storage.Match{}

	searchText := strings.ToLower(v.searchEntry.Text)
	formatFilter := v.formatSelect.Selected
	resultFilter := v.resultSelect.Selected
	opponentFilter := v.opponentSelect.Selected

	// Parse date filters
	var startDate, endDate *time.Time
	if v.startDateEntry.Text != "" {
		if t, err := time.Parse("2006-01-02", v.startDateEntry.Text); err == nil {
			// Start of day (00:00:00)
			startDate = &t
		}
	}
	if v.endDateEntry.Text != "" {
		if t, err := time.Parse("2006-01-02", v.endDateEntry.Text); err == nil {
			// End of day (23:59:59.999)
			endOfDay := t.Add(24*time.Hour - time.Nanosecond)
			endDate = &endOfDay
		}
	}

	for _, match := range v.allMatches {
		// Search filter
		if searchText != "" {
			eventName := strings.ToLower(match.EventName)
			opponentName := ""
			if match.OpponentName != nil {
				opponentName = strings.ToLower(*match.OpponentName)
			}

			if !strings.Contains(eventName, searchText) && !strings.Contains(opponentName, searchText) {
				continue
			}
		}

		// Format filter
		if formatFilter != "All Formats" {
			// Map user-facing format names to database values
			// Ladder and Play both map to "constructed"
			mappedFormat := v.mapFormat(formatFilter)
			if match.Format != mappedFormat {
				continue
			}
		}

		// Result filter
		if resultFilter != "All Results" && match.Result != resultFilter {
			continue
		}

		// Opponent filter
		if opponentFilter != "All Opponents" {
			matchOpponent := ""
			if match.OpponentName != nil && *match.OpponentName != "" {
				matchOpponent = *match.OpponentName
			}
			// Skip this match if it doesn't have a valid opponent name or doesn't match filter
			if matchOpponent == "" || matchOpponent != opponentFilter {
				continue
			}
		}

		// Date range filter
		if startDate != nil && match.Timestamp.Before(*startDate) {
			continue
		}
		if endDate != nil && match.Timestamp.After(*endDate) {
			continue
		}

		v.filteredMatches = append(v.filteredMatches, match)
	}

	v.calculatePagination()
	v.updateStatusLabel()
}

// mapFormat maps user-facing format names to database format values.
func (v *MatchHistoryViewer) mapFormat(format string) string {
	switch format {
	case "Constructed":
		return "constructed"
	case "Limited":
		return "limited"
	case "Draft":
		return "draft"
	case "Sealed":
		return "sealed"
	default:
		return strings.ToLower(format)
	}
}

// calculatePagination calculates pagination parameters.
func (v *MatchHistoryViewer) calculatePagination() {
	totalMatches := len(v.filteredMatches)
	v.totalPages = (totalMatches + matchesPerPage - 1) / matchesPerPage
	if v.totalPages == 0 {
		v.totalPages = 1
	}
}

// getPageMatchCount returns the number of matches on the current page.
func (v *MatchHistoryViewer) getPageMatchCount() int {
	start := v.currentPage * matchesPerPage
	end := start + matchesPerPage

	if start >= len(v.filteredMatches) {
		return 0
	}

	if end > len(v.filteredMatches) {
		end = len(v.filteredMatches)
	}

	return end - start
}

// getPageMatch returns a specific match from the current page.
func (v *MatchHistoryViewer) getPageMatch(index int) *storage.Match {
	actualIndex := v.currentPage*matchesPerPage + index
	if actualIndex >= len(v.filteredMatches) {
		return nil
	}
	return v.filteredMatches[actualIndex]
}

// refreshList refreshes the match list and pagination controls.
func (v *MatchHistoryViewer) refreshList() {
	if v.matchList != nil {
		v.matchList.Refresh()
	}
	if v.pageLabel != nil {
		v.pageLabel.SetText(fmt.Sprintf("Page %d of %d", v.currentPage+1, v.totalPages))
	}
	v.updateStatusLabel()
}

// updateStatusLabel updates the status label with match counts.
func (v *MatchHistoryViewer) updateStatusLabel() {
	if v.statusLabel != nil {
		v.statusLabel.SetText(fmt.Sprintf(
			"Showing %d of %d total matches",
			len(v.filteredMatches),
			len(v.allMatches),
		))
	}
}

// showMatchDetail shows a detail dialog for the selected match.
func (v *MatchHistoryViewer) showMatchDetail(match *storage.Match) {
	detailText := fmt.Sprintf(`Match Details
=============

Result: %s (%d-%d)
Event: %s
Format: %s
Date: %s

`,
		strings.ToUpper(match.Result),
		match.PlayerWins,
		match.OpponentWins,
		match.EventName,
		match.Format,
		match.Timestamp.Format("2006-01-02 15:04:05"),
	)

	if match.OpponentName != nil {
		detailText += fmt.Sprintf("Opponent: %s\n", *match.OpponentName)
	}

	if match.ResultReason != nil {
		detailText += fmt.Sprintf("Result Reason: %s\n", *match.ResultReason)
	}

	if match.DurationSeconds != nil {
		duration := time.Duration(*match.DurationSeconds) * time.Second
		detailText += fmt.Sprintf("Duration: %s\n", duration.String())
	}

	if match.RankBefore != nil && match.RankAfter != nil {
		detailText += fmt.Sprintf("\nRank Change: %s → %s\n", *match.RankBefore, *match.RankAfter)
	}

	detailText += fmt.Sprintf("\nMatch ID: %s\n", match.ID)

	detailLabel := widget.NewLabel(detailText)
	detailLabel.Wrapping = fyne.TextWrapWord

	dialog.ShowCustom(
		"Match Details",
		"Close",
		container.NewScroll(detailLabel),
		v.app.window,
	)
}

// exportMatches exports the filtered matches to a file.
func (v *MatchHistoryViewer) exportMatches() {
	if len(v.filteredMatches) == 0 {
		dialog.ShowInformation("Export", "No matches to export", v.app.window)
		return
	}

	// Create CSV content
	csv := "Result,Date,Time,Event,Format,Score,Opponent,Duration,Rank Before,Rank After,Match ID\n"

	for _, match := range v.filteredMatches {
		result := match.Result
		date := match.Timestamp.Format("2006-01-02")
		timeStr := match.Timestamp.Format("15:04:05")
		event := match.EventName
		format := match.Format
		score := fmt.Sprintf("%d-%d", match.PlayerWins, match.OpponentWins)

		opponent := ""
		if match.OpponentName != nil {
			opponent = *match.OpponentName
		}

		duration := ""
		if match.DurationSeconds != nil {
			duration = fmt.Sprintf("%d", *match.DurationSeconds)
		}

		rankBefore := ""
		if match.RankBefore != nil {
			rankBefore = *match.RankBefore
		}

		rankAfter := ""
		if match.RankAfter != nil {
			rankAfter = *match.RankAfter
		}

		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			result, date, timeStr, event, format, score, opponent, duration, rankBefore, rankAfter, match.ID)
	}

	// Show save dialog
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer func() {
			if closeErr := writer.Close(); closeErr != nil {
				dialog.ShowError(closeErr, v.app.window)
			}
		}()

		_, err = writer.Write([]byte(csv))
		if err != nil {
			dialog.ShowError(err, v.app.window)
			return
		}

		dialog.ShowInformation("Export Complete",
			fmt.Sprintf("Exported %d matches to %s", len(v.filteredMatches), writer.URI().Name()),
			v.app.window)
	}, v.app.window)

	saveDialog.SetFileName(fmt.Sprintf("mtga_matches_%s.csv", time.Now().Format("20060102_150405")))
	saveDialog.Show()
}

// updateOpponentFilter updates the opponent dropdown with unique opponents from matches.
func (v *MatchHistoryViewer) updateOpponentFilter() {
	// Skip if opponent select hasn't been created yet (during initialization)
	if v.opponentSelect == nil {
		return
	}

	// Collect unique opponents
	opponentSet := make(map[string]bool)
	for _, match := range v.allMatches {
		if match.OpponentName != nil && *match.OpponentName != "" {
			opponentSet[*match.OpponentName] = true
		}
		// Don't add "Unknown" to the list - only show actual opponent names
	}

	// Convert to sorted list
	opponents := []string{"All Opponents"}
	for opponent := range opponentSet {
		opponents = append(opponents, opponent)
	}

	// Sort alphabetically (except "All Opponents" stays first)
	sort.Strings(opponents[1:])

	// Update dropdown options
	currentSelection := v.opponentSelect.Selected
	v.opponentSelect.Options = opponents

	// Restore selection if it still exists
	found := false
	for _, opt := range opponents {
		if opt == currentSelection {
			v.opponentSelect.Selected = currentSelection
			found = true
			break
		}
	}
	if !found {
		v.opponentSelect.Selected = "All Opponents"
	}

	v.opponentSelect.Refresh()
}

// showDatePickerDialog shows a custom date picker dialog using a calendar widget.
func (v *MatchHistoryViewer) showDatePickerDialog(title string, targetEntry *widget.Entry) {
	// Parse current date if set, otherwise use today
	var selectedDate time.Time
	if targetEntry.Text != "" {
		if parsed, err := time.Parse("2006-01-02", targetEntry.Text); err == nil {
			selectedDate = parsed
		} else {
			selectedDate = time.Now()
		}
	} else {
		selectedDate = time.Now()
	}

	// Variable to hold the dialog reference
	var confirmDialog dialog.Dialog

	// Create calendar widget with auto-select on click
	calendar := widget.NewCalendar(selectedDate, func(t time.Time) {
		selectedDate = t
		// Auto-select on date click - apply and close the dialog
		targetEntry.SetText(selectedDate.Format("2006-01-02"))
		if confirmDialog != nil {
			confirmDialog.Hide()
		}
	})

	// Create dialog with calendar
	content := container.NewVBox(
		widget.NewLabel("Click a date to select"),
		widget.NewSeparator(),
		calendar,
	)

	confirmDialog = dialog.NewCustom(title, "Cancel", content, v.app.window)

	confirmDialog.Resize(fyne.NewSize(400, 500))
	confirmDialog.Show()
}

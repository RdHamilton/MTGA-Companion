package gui

import (
	"fyne.io/fyne/v2/widget"
)

// Tooltip constants for various UI elements throughout the application.
const (
	// Chart Dashboard Tooltips
	TooltipDateRange   = "Select the time period for displaying match data"
	TooltipFormat      = "Filter matches by game format (Constructed, Limited, etc.)"
	TooltipChartType   = "Toggle between line and bar chart visualizations"
	TooltipSortBy      = "Change the order in which decks are displayed"
	TooltipRefresh     = "Reload the chart with current filter settings"
	TooltipExport      = "Export the current chart as a PNG image (coming soon)"
	TooltipCustomRange = "Specify a custom date range for analysis"

	// Statistics Tooltips
	TooltipWinRate      = "Percentage of matches won out of total matches played"
	TooltipGameWinRate  = "Percentage of individual games won (best-of-3 matches count multiple games)"
	TooltipTotalMatches = "Total number of matches recorded in the database"

	// Settings Tooltips
	TooltipLogPath      = "Path to MTGA Player.log file (auto-detected if empty)"
	TooltipPollInterval = "How often to check the log file for updates (e.g., 2s, 5s)"
	TooltipFsnotify     = "Use file system events for real-time monitoring (recommended)"
	TooltipCacheEnabled = "Enable caching for faster card rating lookups"
	TooltipCacheTTL     = "How long to keep cached data before refreshing"
	TooltipDebugMode    = "Enable detailed debug logging for troubleshooting"
	TooltipSetFile      = "Direct path to 17Lands set data file for draft ratings"
	TooltipSetCode      = "Set code (e.g., BLB, DSK) for auto-loading draft ratings"
	TooltipDraftFormat  = "Default draft format (PremierDraft, QuickDraft, etc.)"
	TooltipResumeDraft  = "Automatically resume active drafts when launching"
	TooltipLookback     = "Hours to scan backward when looking for active drafts"

	// Deck Performance Tooltips
	TooltipConfidence  = "High (30+ matches), Medium (10-29), Low (<10) - indicates statistical reliability"
	TooltipDeckDetails = "Click to view detailed performance statistics for this deck"

	// Match History Tooltips
	TooltipMatchResult = "W = Win, L = Loss"
	TooltipMatchScore  = "Games won - Games lost in this match"
	TooltipEventName   = "The MTGA event or queue where this match was played"
)

// AddTooltip adds a tooltip to a widget.
// Note: Fyne v2 has limited tooltip support. This function provides a placeholder
// for future tooltip functionality. For now, tooltips are embedded in UI as helper text.
func AddTooltip(w interface{}, tooltip string) {
	// Placeholder for future tooltip support in Fyne
	// When Fyne adds tooltip API, this can be implemented
	_ = w
	_ = tooltip
}

// AddButtonTooltip adds a tooltip to a button widget.
func AddButtonTooltip(button *widget.Button, tooltip string) {
	AddTooltip(button, tooltip)
}

// AddSelectTooltip adds a tooltip to a select widget.
func AddSelectTooltip(sel *widget.Select, tooltip string) {
	AddTooltip(sel, tooltip)
}

// AddCheckTooltip adds a tooltip to a checkbox widget.
func AddCheckTooltip(check *widget.Check, tooltip string) {
	AddTooltip(check, tooltip)
}

// AddEntryTooltip adds a tooltip to an entry widget.
func AddEntryTooltip(entry *widget.Entry, tooltip string) {
	AddTooltip(entry, tooltip)
}

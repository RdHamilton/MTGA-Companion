package gui

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/config"
)

// createSettingsView creates the settings configuration view with material design principles.
func (a *App) createSettingsView() fyne.CanvasObject {
	// Load current configuration
	cfg, err := config.Load()
	if err != nil {
		return widget.NewLabel(fmt.Sprintf("Error loading config: %v", err))
	}

	// Create form fields for each setting
	// Log Configuration Section
	logFilePathEntry := widget.NewEntry()
	logFilePathEntry.SetPlaceHolder("Auto-detected if empty")
	logFilePathEntry.SetText(cfg.Log.FilePath)

	logPollIntervalEntry := widget.NewEntry()
	logPollIntervalEntry.SetPlaceHolder("e.g., 2s, 5s")
	logPollIntervalEntry.SetText(cfg.Log.PollInterval)

	logUseFsnotifyCheck := widget.NewCheck("Use file system events for real-time monitoring", nil)
	logUseFsnotifyCheck.Checked = cfg.Log.UseFsnotify

	logEnableMetricsCheck := widget.NewCheck("Enable performance metrics collection", nil)
	logEnableMetricsCheck.Checked = cfg.Log.EnableMetrics

	// Cache Configuration Section
	cacheEnabledCheck := widget.NewCheck("Enable in-memory caching for card ratings", nil)
	cacheEnabledCheck.Checked = cfg.Cache.Enabled

	cacheTTLEntry := widget.NewEntry()
	cacheTTLEntry.SetPlaceHolder("e.g., 24h, 48h")
	cacheTTLEntry.SetText(cfg.Cache.TTL)

	cacheMaxSizeEntry := widget.NewEntry()
	cacheMaxSizeEntry.SetPlaceHolder("0 for unlimited")
	cacheMaxSizeEntry.SetText(strconv.Itoa(cfg.Cache.MaxSize))

	// Overlay Configuration Section
	overlaySetFileEntry := widget.NewEntry()
	overlaySetFileEntry.SetPlaceHolder("Path to 17Lands set file")
	overlaySetFileEntry.SetText(cfg.Overlay.SetFile)

	overlaySetCodeEntry := widget.NewEntry()
	overlaySetCodeEntry.SetPlaceHolder("e.g., BLB, DSK, MKM")
	overlaySetCodeEntry.SetText(cfg.Overlay.SetCode)

	overlayFormatEntry := widget.NewEntry()
	overlayFormatEntry.SetPlaceHolder("e.g., PremierDraft, QuickDraft")
	overlayFormatEntry.SetText(cfg.Overlay.Format)

	overlayResumeCheck := widget.NewCheck("Automatically resume active drafts on launch", nil)
	overlayResumeCheck.Checked = cfg.Overlay.Resume

	overlayLookbackEntry := widget.NewEntry()
	overlayLookbackEntry.SetPlaceHolder("24")
	overlayLookbackEntry.SetText(strconv.Itoa(cfg.Overlay.LookbackHrs))

	// Application Configuration Section
	appDebugModeCheck := widget.NewCheck("Enable detailed debug logging", nil)
	appDebugModeCheck.Checked = cfg.App.DebugMode

	// Create section headers with better styling
	logHeader := widget.NewRichTextFromMarkdown("### Log File Configuration")
	cacheHeader := widget.NewRichTextFromMarkdown("### Cache Configuration")
	overlayHeader := widget.NewRichTextFromMarkdown("### Draft Overlay Configuration")
	appHeader := widget.NewRichTextFromMarkdown("### Application Settings")

	// Create form with organized sections
	form := &widget.Form{
		Items: []*widget.FormItem{
			// Log settings section
			{Text: "", Widget: logHeader},
			{Text: "Log File Path", Widget: logFilePathEntry, HintText: "Path to MTGA Player.log (auto-detected if empty)"},
			{Text: "Poll Interval", Widget: logPollIntervalEntry, HintText: "How often to check log file for updates"},
			{Text: "", Widget: logUseFsnotifyCheck},
			{Text: "", Widget: logEnableMetricsCheck},
			{Text: "", Widget: widget.NewSeparator()},

			// Cache settings section
			{Text: "", Widget: cacheHeader},
			{Text: "", Widget: cacheEnabledCheck},
			{Text: "Cache TTL", Widget: cacheTTLEntry, HintText: "How long to keep cached data"},
			{Text: "Max Cache Size", Widget: cacheMaxSizeEntry, HintText: "Maximum number of cached entries (0 = unlimited)"},
			{Text: "", Widget: widget.NewSeparator()},

			// Overlay settings section
			{Text: "", Widget: overlayHeader},
			{Text: "Set File Path", Widget: overlaySetFileEntry, HintText: "Direct path to 17Lands set data file"},
			{Text: "Set Code", Widget: overlaySetCodeEntry, HintText: "Set code for auto-loading set file"},
			{Text: "Draft Format", Widget: overlayFormatEntry, HintText: "Default draft format to use"},
			{Text: "", Widget: overlayResumeCheck},
			{Text: "Lookback Hours", Widget: overlayLookbackEntry, HintText: "How many hours back to scan for active drafts"},
			{Text: "", Widget: widget.NewSeparator()},

			// App settings section
			{Text: "", Widget: appHeader},
			{Text: "", Widget: appDebugModeCheck},
		},
		OnSubmit: func() {
			// Gather values from form
			newCfg := &config.Config{
				Log: config.LogConfig{
					FilePath:      logFilePathEntry.Text,
					PollInterval:  logPollIntervalEntry.Text,
					UseFsnotify:   logUseFsnotifyCheck.Checked,
					EnableMetrics: logEnableMetricsCheck.Checked,
				},
				Cache: config.CacheConfig{
					Enabled: cacheEnabledCheck.Checked,
					TTL:     cacheTTLEntry.Text,
					MaxSize: 0,
				},
				Overlay: config.OverlayConfig{
					SetFile:     overlaySetFileEntry.Text,
					SetCode:     overlaySetCodeEntry.Text,
					Format:      overlayFormatEntry.Text,
					Resume:      overlayResumeCheck.Checked,
					LookbackHrs: 0,
				},
				App: config.AppConfig{
					DebugMode: appDebugModeCheck.Checked,
				},
			}

			// Parse cache max size
			if maxSize, err := strconv.Atoi(cacheMaxSizeEntry.Text); err == nil {
				newCfg.Cache.MaxSize = maxSize
			} else {
				dialog.ShowError(fmt.Errorf("invalid cache max size: %w", err), a.window)
				return
			}

			// Parse lookback hours
			if lookbackHrs, err := strconv.Atoi(overlayLookbackEntry.Text); err == nil {
				newCfg.Overlay.LookbackHrs = lookbackHrs
			} else {
				dialog.ShowError(fmt.Errorf("invalid lookback hours: %w", err), a.window)
				return
			}

			// Validate configuration
			if err := newCfg.Validate(); err != nil {
				dialog.ShowError(err, a.window)
				return
			}

			// Save configuration
			if err := newCfg.Save(); err != nil {
				dialog.ShowError(fmt.Errorf("failed to save config: %w", err), a.window)
				return
			}

			// Show success dialog
			dialog.ShowInformation("Success", "Settings saved successfully.\n\nNote: Some settings may require an application restart to take effect.", a.window)
		},
		SubmitText: "Save Settings",
	}

	// Add Restore Defaults button
	restoreButton := widget.NewButton("Restore Defaults", func() {
		// Confirm with user
		dialog.ShowConfirm("Restore Defaults", "Are you sure you want to restore default settings?", func(confirmed bool) {
			if !confirmed {
				return
			}

			// Get default config
			defaultCfg := config.DefaultConfig()

			// Update form fields
			logFilePathEntry.SetText(defaultCfg.Log.FilePath)
			logPollIntervalEntry.SetText(defaultCfg.Log.PollInterval)
			logUseFsnotifyCheck.Checked = defaultCfg.Log.UseFsnotify
			logEnableMetricsCheck.Checked = defaultCfg.Log.EnableMetrics

			cacheEnabledCheck.Checked = defaultCfg.Cache.Enabled
			cacheTTLEntry.SetText(defaultCfg.Cache.TTL)
			cacheMaxSizeEntry.SetText(strconv.Itoa(defaultCfg.Cache.MaxSize))

			overlaySetFileEntry.SetText(defaultCfg.Overlay.SetFile)
			overlaySetCodeEntry.SetText(defaultCfg.Overlay.SetCode)
			overlayFormatEntry.SetText(defaultCfg.Overlay.Format)
			overlayResumeCheck.Checked = defaultCfg.Overlay.Resume
			overlayLookbackEntry.SetText(strconv.Itoa(defaultCfg.Overlay.LookbackHrs))

			appDebugModeCheck.Checked = defaultCfg.App.DebugMode

			// Refresh widgets
			logUseFsnotifyCheck.Refresh()
			logEnableMetricsCheck.Refresh()
			cacheEnabledCheck.Refresh()
			overlayResumeCheck.Refresh()
			appDebugModeCheck.Refresh()

			dialog.ShowInformation("Defaults Restored", "Default settings have been restored. Click 'Save Settings' to persist changes.", a.window)
		}, a.window)
	})

	// Add About button
	aboutButton := widget.NewButton("About MTGA Companion", func() {
		a.showAboutDialog()
	})

	// Create transparent spacers for left and right margins
	leftSpacer := canvas.NewRectangle(color.Transparent)
	leftSpacer.SetMinSize(fyne.NewSize(20, 0))
	rightSpacer := canvas.NewRectangle(color.Transparent)
	rightSpacer.SetMinSize(fyne.NewSize(20, 0))

	// Layout: form in center with margins, buttons at bottom
	return container.NewBorder(
		nil,
		container.NewPadded(
			container.NewVBox(
				widget.NewSeparator(),
				container.NewHBox(
					restoreButton,
					aboutButton,
				),
			),
		),
		leftSpacer,
		rightSpacer,
		container.NewScroll(
			container.NewPadded(form),
		),
	)
}

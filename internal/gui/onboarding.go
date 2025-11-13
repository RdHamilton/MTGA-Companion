package gui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/ramonehamilton/MTGA-Companion/internal/config"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// OnboardingWizard manages the multi-step onboarding experience.
type OnboardingWizard struct {
	app           *App
	dialog        dialog.Dialog
	contentBox    *fyne.Container
	currentStep   int
	totalSteps    int
	logPath       string
	logPathStatus string
}

// showOnboarding displays the first-run onboarding wizard.
func (a *App) showOnboarding() {
	// Check if onboarding has been completed
	cfg, _ := config.Load()
	if cfg != nil && cfg.App.OnboardingCompleted {
		return // Skip if already completed
	}

	wizard := &OnboardingWizard{
		app:         a,
		currentStep: 0,
		totalSteps:  4,
	}

	wizard.show()
}

// ShowOnboardingWizard allows manually triggering the onboarding wizard.
// Used by the Settings page "Show Onboarding" button.
func (a *App) ShowOnboardingWizard() {
	wizard := &OnboardingWizard{
		app:         a,
		currentStep: 0,
		totalSteps:  4,
	}

	wizard.show()
}

// show creates and displays the onboarding wizard dialog.
func (w *OnboardingWizard) show() {
	// Create content container that will be updated
	w.contentBox = container.NewVBox()

	// Create custom dialog (empty dismiss button label since we handle navigation with our own buttons)
	w.dialog = dialog.NewCustom("MTGA Companion Setup", "", w.contentBox, w.app.window)

	// Render first step
	w.renderStep()

	// Size the dialog
	w.dialog.Resize(fyne.NewSize(700, 600))
	w.dialog.Show()
}

// renderStep renders the current step's content.
func (w *OnboardingWizard) renderStep() {
	var content fyne.CanvasObject
	var buttons fyne.CanvasObject

	switch w.currentStep {
	case 0:
		content, buttons = w.welcomeStep()
	case 1:
		content, buttons = w.loggingStep()
	case 2:
		content, buttons = w.logFileStep()
	case 3:
		content, buttons = w.featuresStep()
	}

	// Update content
	w.contentBox.Objects = []fyne.CanvasObject{
		content,
		widget.NewSeparator(),
		buttons,
	}
	w.contentBox.Refresh()
}

// welcomeStep creates the welcome screen content.
func (w *OnboardingWizard) welcomeStep() (fyne.CanvasObject, fyne.CanvasObject) {
	content := widget.NewRichTextFromMarkdown(fmt.Sprintf(`# Welcome to %s!

Thank you for installing MTGA Companion. This quick setup will help you get started.

## What is MTGA Companion?

MTGA Companion reads your Magic: The Gathering Arena log files to provide:
- 📊 **Statistics tracking** - Win rates, performance analytics
- 🃏 **Draft history** - Record and analyze your drafts
- 📈 **Performance charts** - Visualize your progress over time
- 💾 **Data exports** - Export your data in CSV/JSON formats

## Quick Start Guide

We'll help you:
1. ✅ Enable detailed logging in MTGA (required)
2. 🔍 Locate your MTGA log file
3. ⚙️  Configure basic settings
4. 🎯 Take a quick tour of features

Click **Next** to continue or **Skip** if you want to explore on your own.`, AppName))

	nextButton := widget.NewButton("Next", func() {
		w.currentStep++
		w.renderStep()
	})
	nextButton.Importance = widget.HighImportance

	skipButton := widget.NewButton("Skip Tour", func() {
		w.completeOnboarding()
		w.dialog.Hide()
	})

	buttons := container.NewHBox(
		skipButton,
		widget.NewLabel(""), // Spacer
		nextButton,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(650, 450))

	return scrollContent, buttons
}

// loggingStep creates the detailed logging instructions.
func (w *OnboardingWizard) loggingStep() (fyne.CanvasObject, fyne.CanvasObject) {
	content := widget.NewRichTextFromMarkdown(`## Step 1: Enable Detailed Logging in MTGA

**IMPORTANT**: MTGA must have detailed logging enabled for this companion app to work.

### How to Enable Detailed Logging:

1. Launch **Magic: The Gathering Arena**
2. Click the **Adjust Options** gear icon ⚙️ (top right)
3. In the Options menu, click **View Account**
4. Find and check the **Detailed Logs** checkbox
   - May also be labeled "Enable Detailed Logs" or "Plugin Support"
5. **Restart MTGA** for changes to take effect

### Why is this needed?

Detailed logging allows MTGA to output game events in JSON format, enabling companion apps to:
- Track your statistics
- Monitor game state
- Analyze your collection
- Display deck information

**Note**: Detailed logging has no impact on game performance and is designed specifically for companion tools.

---

**Have you enabled detailed logging in MTGA?**`)

	nextButton := widget.NewButton("Yes, I've enabled it", func() {
		w.currentStep++
		w.renderStep()
	})
	nextButton.Importance = widget.HighImportance

	helpButton := widget.NewButton("I need help", func() {
		dialog.ShowInformation("Help with Detailed Logging",
			"If you can't find the Detailed Logs option:\n\n"+
				"1. Make sure MTGA is fully updated\n"+
				"2. Look under Options → View Account\n"+
				"3. The setting may be labeled differently in your client\n\n"+
				"For more help, visit:\n"+DocsURL, w.app.window)
	})

	backButton := widget.NewButton("Back", func() {
		w.currentStep--
		w.renderStep()
	})

	buttons := container.NewHBox(
		backButton,
		helpButton,
		widget.NewLabel(""), // Spacer
		nextButton,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(650, 450))

	return scrollContent, buttons
}

// logFileStep creates the log file detection screen.
func (w *OnboardingWizard) logFileStep() (fyne.CanvasObject, fyne.CanvasObject) {
	// Try to detect log file
	logPath, err := logreader.DefaultLogPath()
	var statusText string

	if err == nil && logPath != "" {
		if _, err := os.Stat(logPath); err == nil {
			statusText = fmt.Sprintf("✅ **Log file found!**\n\nPath: `%s`\n\nYou're all set! The app will automatically read from this location.", logPath)
		} else {
			statusText = fmt.Sprintf("⚠️ **Log file detected but not accessible**\n\nPath: `%s`\n\nThe file may not exist yet. Play a match in MTGA to create it.", logPath)
		}
	} else {
		statusText = "❌ **Log file not found**\n\nCouldn't auto-detect your MTGA log file. You can manually specify the path in Settings."
	}

	w.logPath = logPath
	w.logPathStatus = statusText

	content := widget.NewRichTextFromMarkdown(fmt.Sprintf(`## Step 2: Verify Log File Location

MTGA Companion needs to know where your MTGA log file is located.

%s

### Default Locations:

**macOS**:
`+"```"+`
~/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log
`+"```"+`

**Windows**:
`+"```"+`
C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log
`+"```"+`

If the auto-detected path is wrong, you can change it in Settings after onboarding.`, statusText))

	nextButton := widget.NewButton("Continue", func() {
		w.currentStep++
		w.renderStep()
	})
	nextButton.Importance = widget.HighImportance

	backButton := widget.NewButton("Back", func() {
		w.currentStep--
		w.renderStep()
	})

	buttons := container.NewHBox(
		backButton,
		widget.NewLabel(""), // Spacer
		nextButton,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(650, 450))

	return scrollContent, buttons
}

// featuresStep creates the features tour screen.
func (w *OnboardingWizard) featuresStep() (fyne.CanvasObject, fyne.CanvasObject) {
	content := widget.NewRichTextFromMarkdown(`## Step 3: Feature Tour

### Main Tabs

**📊 Statistics**
- View your overall win rates
- Track match and game statistics
- See performance trends

**📜 Match History**
- Browse your recent matches
- Filter by date, format, or result
- View detailed match information

**📈 Charts**
- Win rate trends over time
- Format distribution analysis
- Deck performance comparison
- Interactive filtering and visualization

**⚙️ Settings**
- Configure log file path
- Adjust polling settings
- Enable/disable features
- Backup and restore data

### Getting Help

- Click **About** in Settings for links and info
- Press **Ctrl/Cmd + H** to show help anytime
- Visit our documentation for detailed guides
- Report issues on GitHub

---

**You're all set!** Click **Finish** to start using MTGA Companion.`)

	finishButton := widget.NewButton("Finish", func() {
		w.completeOnboarding()
		w.dialog.Hide()
		dialog.ShowInformation("Setup Complete!",
			"Welcome to MTGA Companion!\n\n"+
				"Play some matches in MTGA to see your statistics appear.\n"+
				"The app monitors your log file automatically.", w.app.window)
	})
	finishButton.Importance = widget.HighImportance

	backButton := widget.NewButton("Back", func() {
		w.currentStep--
		w.renderStep()
	})

	buttons := container.NewHBox(
		backButton,
		widget.NewLabel(""), // Spacer
		finishButton,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(650, 450))

	return scrollContent, buttons
}

// completeOnboarding marks onboarding as completed and saves to config.
func (w *OnboardingWizard) completeOnboarding() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	cfg.App.OnboardingCompleted = true

	// Ensure config directory exists
	configDir := filepath.Dir(config.ConfigPath())
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		// Log error but don't block
		fmt.Printf("Warning: couldn't create config directory: %v\n", err)
		return
	}

	if err := cfg.Save(); err != nil {
		// Log error but don't block
		fmt.Printf("Warning: couldn't save onboarding state: %v\n", err)
	}
}

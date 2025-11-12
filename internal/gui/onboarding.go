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

// showOnboarding displays the first-run onboarding wizard.
func (a *App) showOnboarding() {
	// Check if onboarding has been completed
	cfg, _ := config.Load()
	if cfg != nil && cfg.App.OnboardingCompleted {
		return // Skip if already completed
	}

	// Create multi-step onboarding dialog
	a.showWelcomeStep()
}

// showWelcomeStep shows the welcome screen.
func (a *App) showWelcomeStep() {
	welcomeContent := widget.NewRichTextFromMarkdown(fmt.Sprintf(`# Welcome to %s!

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
		a.showLoggingStep()
	})

	skipButton := widget.NewButton("Skip Tour", func() {
		a.completeOnboarding()
	})

	content := container.NewVBox(
		welcomeContent,
		widget.NewSeparator(),
		container.NewHBox(
			skipButton,
			nextButton,
		),
	)

	customDialog := dialog.NewCustom("Welcome to MTGA Companion", "", content, a.window)
	customDialog.Resize(fyne.NewSize(600, 500))
	customDialog.Show()
}

// showLoggingStep shows instructions for enabling detailed logging.
func (a *App) showLoggingStep() {
	loggingContent := widget.NewRichTextFromMarkdown(`## Step 1: Enable Detailed Logging in MTGA

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

	yesButton := widget.NewButton("Yes, I've enabled it", func() {
		a.showLogFileStep()
	})

	helpButton := widget.NewButton("I need help", func() {
		dialog.ShowInformation("Help with Detailed Logging",
			"If you can't find the Detailed Logs option:\n\n"+
				"1. Make sure MTGA is fully updated\n"+
				"2. Look under Options → View Account\n"+
				"3. The setting may be labeled differently in your client\n\n"+
				"For more help, visit:\n"+DocsURL, a.window)
	})

	backButton := widget.NewButton("Back", func() {
		a.showWelcomeStep()
	})

	content := container.NewVBox(
		loggingContent,
		widget.NewSeparator(),
		container.NewHBox(
			backButton,
			helpButton,
			yesButton,
		),
	)

	customDialog := dialog.NewCustom("Enable MTGA Detailed Logging", "", content, a.window)
	customDialog.Resize(fyne.NewSize(600, 550))
	customDialog.Show()
}

// showLogFileStep shows log file detection.
func (a *App) showLogFileStep() {
	// Try to detect log file
	logPath, err := logreader.DefaultLogPath()
	var statusText string
	var statusColor string

	if err == nil && logPath != "" {
		if _, err := os.Stat(logPath); err == nil {
			statusText = fmt.Sprintf("✅ **Log file found!**\n\nPath: `%s`\n\nYou're all set! The app will automatically read from this location.", logPath)
			statusColor = "success"
		} else {
			statusText = fmt.Sprintf("⚠️ **Log file detected but not accessible**\n\nPath: `%s`\n\nThe file may not exist yet. Play a match in MTGA to create it.", logPath)
			statusColor = "warning"
		}
	} else {
		statusText = "❌ **Log file not found**\n\nCouldn't auto-detect your MTGA log file. You can manually specify the path in Settings."
		statusColor = "error"
	}

	logFileContent := widget.NewRichTextFromMarkdown(fmt.Sprintf(`## Step 2: Verify Log File Location

MTGA Companion needs to know where your MTGA log file is located.

%s

### Default Locations:

**macOS**:
` + "```" + `
~/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log
` + "```" + `

**Windows**:
` + "```" + `
C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log
` + "```" + `

If the auto-detected path is wrong, you can change it in Settings after onboarding.`, statusText))

	_ = statusColor // For future use with styling

	continueButton := widget.NewButton("Continue", func() {
		a.showFeaturesStep()
	})

	settingsButton := widget.NewButton("Open Settings", func() {
		// Close onboarding and navigate to Settings tab
		a.completeOnboarding()
		// Navigate to Settings would go here
	})

	backButton := widget.NewButton("Back", func() {
		a.showLoggingStep()
	})

	buttons := container.NewHBox(backButton, continueButton)
	if logPath == "" {
		buttons = container.NewHBox(backButton, settingsButton, continueButton)
	}

	content := container.NewVBox(
		logFileContent,
		widget.NewSeparator(),
		buttons,
	)

	customDialog := dialog.NewCustom("Log File Detection", "", content, a.window)
	customDialog.Resize(fyne.NewSize(600, 500))
	customDialog.Show()
}

// showFeaturesStep shows a quick tour of features.
func (a *App) showFeaturesStep() {
	featuresContent := widget.NewRichTextFromMarkdown(`## Step 3: Feature Tour

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

- Click **About MTGA Companion** in Settings for links and info
- Visit our documentation for detailed guides
- Report issues on GitHub

---

**You're all set!** Click **Finish** to start using MTGA Companion.`)

	finishButton := widget.NewButton("Finish", func() {
		a.completeOnboarding()
		dialog.ShowInformation("Setup Complete!",
			"Welcome to MTGA Companion!\n\n"+
				"Play some matches in MTGA to see your statistics appear.\n"+
				"The app monitors your log file automatically.", a.window)
	})

	backButton := widget.NewButton("Back", func() {
		a.showLogFileStep()
	})

	content := container.NewVBox(
		featuresContent,
		widget.NewSeparator(),
		container.NewHBox(
			backButton,
			finishButton,
		),
	)

	customDialog := dialog.NewCustom("Quick Tour", "", content, a.window)
	customDialog.Resize(fyne.NewSize(600, 550))
	customDialog.Show()
}

// completeOnboarding marks onboarding as completed and saves to config.
func (a *App) completeOnboarding() {
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

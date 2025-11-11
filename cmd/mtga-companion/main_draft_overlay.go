package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/draft"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

// runDraftOverlay launches the draft overlay window.
func runDraftOverlay() {
	fmt.Println("MTGA Draft Overlay")
	fmt.Println("==================")
	fmt.Println()

	// Determine set file path
	setFile := getSetFileForOverlay()
	if setFile == nil {
		log.Fatal("No set file specified or found. Use -set-file or -overlay-set flag.")
	}

	// Determine log path
	playerLogPath := getLogPathForOverlay()

	fmt.Printf("Set File: %s (%s)\n", setFile.Meta.SetCode, setFile.Meta.DraftFormat)
	fmt.Printf("Log Path: %s\n", playerLogPath)
	fmt.Printf("Resume Mode: %v\n\n", *overlayResume)

	// Create overlay configuration
	config := draft.OverlayConfig{
		LogPath:        playerLogPath,
		SetFile:        setFile,
		BayesianConfig: draft.DefaultBayesianConfig(),
		ColorConfig:    draft.DefaultColorAffinityConfig(),
		ResumeEnabled:  *overlayResume,
		LookbackHours:  *overlayLookback,
	}

	// Create and run overlay window
	overlay := gui.NewDraftOverlayWindow(config)
	overlay.Run()
}

// getSetFileForOverlay determines which set file to use for the overlay.
func getSetFileForOverlay() *seventeenlands.SetFile {
	// If explicit path provided, use it
	if *setFilePath != "" {
		fmt.Printf("Loading set file: %s\n", *setFilePath)
		setFile, err := loadSetFileFromPath(*setFilePath)
		if err != nil {
			log.Fatalf("Error loading set file: %v", err)
		}
		return setFile
	}

	// If set code provided, auto-load from standard location
	if *overlaySetCode != "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}

		setsDir := filepath.Join(homeDir, ".mtga-companion", "Sets")
		filename := fmt.Sprintf("%s_%s_data.json", *overlaySetCode, *overlayFormat)
		path := filepath.Join(setsDir, filename)

		fmt.Printf("Auto-loading set file: %s\n", path)
		setFile, err := loadSetFileFromPath(path)
		if err != nil {
			log.Fatalf("Error loading set file: %v\nTip: Download with: mtga-companion sets download --set %s --format %s",
				err, *overlaySetCode, *overlayFormat)
		}
		return setFile
	}

	// Try to find most recent set file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	setsDir := filepath.Join(homeDir, ".mtga-companion", "Sets")
	entries, err := os.ReadDir(setsDir)
	if err != nil || len(entries) == 0 {
		return nil
	}

	// Find first .json file
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(setsDir, entry.Name())
			fmt.Printf("Auto-detected set file: %s\n", path)
			setFile, err := loadSetFileFromPath(path)
			if err == nil {
				return setFile
			}
		}
	}

	return nil
}

// loadSetFileFromPath loads a set file from a JSON file path.
func loadSetFileFromPath(path string) (*seventeenlands.SetFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var setFile seventeenlands.SetFile
	if err := json.Unmarshal(data, &setFile); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &setFile, nil
}

// getLogPathForOverlay determines the MTGA Player.log path.
func getLogPathForOverlay() string {
	// If explicit path provided, use it
	if *logPath != "" {
		return *logPath
	}

	// Auto-detect platform default
	defaultPath, err := logreader.DefaultLogPath()
	if err != nil {
		log.Fatalf("Error detecting log path: %v\nPlease specify with -log-path flag", err)
	}

	// Verify it exists
	if _, err := os.Stat(defaultPath); err != nil {
		log.Fatalf("Log file not found at %s\nPlease specify with -log-path flag", defaultPath)
	}

	return defaultPath
}

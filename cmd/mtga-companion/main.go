package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

func main() {
	// Get the default log path for the current platform
	logPath, err := logreader.DefaultLogPath()
	if err != nil {
		log.Fatalf("Error getting default log path: %v", err)
	}

	fmt.Printf("MTGA Player.log path: %s\n", logPath)

	// Check if the log file exists
	exists, err := logreader.LogExists(logPath)
	if err != nil {
		log.Fatalf("Error checking if log exists: %v", err)
	}

	if !exists {
		fmt.Println("\nPlayer.log not found!")
		fmt.Println("Please ensure:")
		fmt.Println("  1. MTG Arena is installed")
		fmt.Println("  2. Detailed logging is enabled in MTG Arena settings")
		fmt.Println("  3. You have run MTG Arena at least once")
		fmt.Println("\nSee README.md for instructions on enabling detailed logging.")
		os.Exit(1)
	}

	fmt.Println("Player.log found!")

	// Create a reader
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		log.Fatalf("Error creating log reader: %v", err)
	}
	defer reader.Close()

	// Read all JSON entries
	fmt.Println("\nReading JSON entries from log...")
	entries, err := reader.ReadAllJSON()
	if err != nil {
		log.Fatalf("Error reading log entries: %v", err)
	}

	fmt.Printf("Found %d JSON entries in the log file.\n", len(entries))

	// Display first few JSON entries as examples
	displayCount := 5
	if len(entries) < displayCount {
		displayCount = len(entries)
	}

	if displayCount > 0 {
		fmt.Printf("\nFirst %d JSON entries:\n", displayCount)
		for i := 0; i < displayCount; i++ {
			entry := entries[i]
			fmt.Printf("\nEntry %d:\n", i+1)
			if entry.Timestamp != "" {
				fmt.Printf("  Timestamp: %s\n", entry.Timestamp)
			}
			fmt.Printf("  JSON Fields: %v\n", getJSONKeys(entry.JSON))
		}
	}
}

// getJSONKeys returns the keys from a JSON map for display purposes.
func getJSONKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

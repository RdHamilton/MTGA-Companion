// Package examples provides example code for using MTGA-Companion.
// This file demonstrates how to use the database layer with the log reader.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// This example demonstrates how to:
// 1. Initialize the database
// 2. Read and parse MTGA log files
// 3. Store parsed data in the database
// 4. Query the database for statistics
//
// NOTE: This requires database migrations to be run first (see issue #19).
func main() {
	// Get a path for the database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, ".mtga-companion", "data.db")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize database
	fmt.Println("Initializing database...")
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create service
	service := storage.NewService(db)
	ctx := context.Background()

	// Get log path
	logPath, err := logreader.DefaultLogPath()
	if err != nil {
		log.Fatalf("Failed to get log path: %v", err)
	}

	// Read log entries
	fmt.Printf("Reading log file: %s\n", logPath)
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()

	entries, err := reader.ReadAllJSON()
	if err != nil {
		log.Fatalf("Failed to read entries: %v", err)
	}

	fmt.Printf("Found %d JSON entries\n", len(entries))

	// Parse arena statistics
	arenaStats, err := logreader.ParseArenaStats(entries)
	if err != nil {
		log.Fatalf("Failed to parse arena stats: %v", err)
	}

	if arenaStats != nil {
		fmt.Println("\nStoring arena statistics...")

		// Store stats in database
		if err := service.StoreArenaStats(ctx, arenaStats); err != nil {
			log.Fatalf("Failed to store arena stats: %v", err)
		}

		fmt.Println("Successfully stored arena statistics")

		// Query stats back from database
		fmt.Println("\nRetrieving statistics from database...")
		stats, err := service.GetStats(ctx, storage.StatsFilter{})
		if err != nil {
			log.Fatalf("Failed to get stats: %v", err)
		}

		if stats != nil {
			fmt.Printf("\nOverall Statistics:\n")
			fmt.Printf("  Matches: %d-%d (%.1f%% win rate)\n",
				stats.MatchesWon, stats.MatchesLost, stats.WinRate*100)
			fmt.Printf("  Games:   %d-%d (%.1f%% win rate)\n",
				stats.GamesWon, stats.GamesLost, stats.GameWinRate*100)
		}
	}

	// Example: Store a match
	// In a real implementation, you would extract this data from log entries
	/*
		match := &storage.Match{
			ID:           "match-123",
			EventID:      "event-456",
			EventName:    "Standard Ranked",
			Timestamp:    time.Now(),
			PlayerWins:   2,
			OpponentWins: 1,
			PlayerTeamID: 1,
			Format:       "Standard",
			Result:       "win",
			CreatedAt:    time.Now(),
		}

		games := []*storage.Game{
			{MatchID: "match-123", GameNumber: 1, Result: "win", CreatedAt: time.Now()},
			{MatchID: "match-123", GameNumber: 2, Result: "loss", CreatedAt: time.Now()},
			{MatchID: "match-123", GameNumber: 3, Result: "win", CreatedAt: time.Now()},
		}

		if err := service.StoreMatch(ctx, match, games); err != nil {
			log.Fatalf("Failed to store match: %v", err)
		}

		fmt.Println("Successfully stored match")
	*/

	fmt.Println("\nDatabase example completed successfully!")
}

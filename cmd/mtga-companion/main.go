package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

func main() {
	// Check if this is a migration command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrationCommand()
		return
	}

	// Get the default log path for the current platform
	logPath, err := logreader.DefaultLogPath()
	if err != nil {
		log.Fatalf("Error getting default log path: %v", err)
	}

	fmt.Printf("MTGA Log Reader\n")
	fmt.Printf("===============\n\n")
	fmt.Printf("Log file: %s\n\n", logPath)

	// Check if the log file exists
	exists, err := logreader.LogExists(logPath)
	if err != nil {
		log.Fatalf("Error checking if log exists: %v", err)
	}

	if !exists {
		fmt.Println("Log file not found!")
		fmt.Println("\nPlease ensure:")
		fmt.Println("  1. MTG Arena is installed")
		fmt.Println("  2. Detailed logging is enabled in MTG Arena settings")
		fmt.Println("  3. You have run MTG Arena at least once")
		fmt.Println("\nSee README.md for instructions on enabling detailed logging.")
		os.Exit(1)
	}

	// Create a reader
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		log.Fatalf("Error creating log reader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader: %v", err)
		}
	}()

	// Read all JSON entries
	fmt.Println("Reading log file...")
	entries, err := reader.ReadAllJSON()
	if err != nil {
		log.Fatalf("Error reading log entries: %v", err)
	}

	fmt.Printf("Found %d JSON entries in the log file.\n\n", len(entries))

	// Parse all data
	profile, inventory, rank := logreader.ParseAll(entries)
	draftHistory, _ := logreader.ParseDraftHistory(entries)
	arenaStats, _ := logreader.ParseArenaStats(entries)

	// Display player profile
	if profile != nil && profile.ScreenName != "" {
		fmt.Println("Player Profile")
		fmt.Println("--------------")
		fmt.Printf("Screen Name: %s\n", profile.ScreenName)
		if profile.ClientID != "" {
			fmt.Printf("Client ID:   %s\n", profile.ClientID)
		}
		fmt.Println()
	}

	// Display inventory
	if inventory != nil {
		fmt.Println("Inventory")
		fmt.Println("---------")
		fmt.Printf("Gems:              %d\n", inventory.Gems)
		fmt.Printf("Gold:              %d\n", inventory.Gold)
		fmt.Printf("Vault Progress:    %d%%\n", inventory.TotalVaultProgress)
		fmt.Println()

		fmt.Println("Wildcards:")
		fmt.Printf("  Common:          %d\n", inventory.WildCardCommons)
		fmt.Printf("  Uncommon:        %d\n", inventory.WildCardUncommons)
		fmt.Printf("  Rare:            %d\n", inventory.WildCardRares)
		fmt.Printf("  Mythic:          %d\n", inventory.WildCardMythics)
		fmt.Println()

		if len(inventory.Boosters) > 0 {
			fmt.Println("Boosters:")
			for _, booster := range inventory.Boosters {
				fmt.Printf("  %s: %d\n", booster.SetCode, booster.Count)
			}
			fmt.Println()
		}

		if len(inventory.CustomTokens) > 0 {
			fmt.Println("Custom Tokens:")
			for token, count := range inventory.CustomTokens {
				fmt.Printf("  %s: %d\n", token, count)
			}
			fmt.Println()
		}
	}

	// Display rank
	if rank != nil {
		fmt.Println("Rank Information")
		fmt.Println("----------------")

		if rank.ConstructedClass != "" || rank.ConstructedLevel > 0 {
			fmt.Println("Constructed:")
			fmt.Printf("  Season:          %d\n", rank.ConstructedSeasonOrdinal)
			if rank.ConstructedClass != "" {
				fmt.Printf("  Rank:            %s %d\n", rank.ConstructedClass, rank.ConstructedLevel)
			} else {
				fmt.Printf("  Level:           %d\n", rank.ConstructedLevel)
			}
			if rank.ConstructedPercentile > 0 {
				fmt.Printf("  Percentile:      %.1f%%\n", rank.ConstructedPercentile)
			}
			if rank.ConstructedStep > 0 {
				fmt.Printf("  Step:            %d\n", rank.ConstructedStep)
			}
			fmt.Println()
		}

		if rank.LimitedClass != "" || rank.LimitedLevel > 0 {
			fmt.Println("Limited:")
			fmt.Printf("  Season:          %d\n", rank.LimitedSeasonOrdinal)
			if rank.LimitedClass != "" {
				fmt.Printf("  Rank:            %s %d\n", rank.LimitedClass, rank.LimitedLevel)
			} else {
				fmt.Printf("  Level:           %d\n", rank.LimitedLevel)
			}
			if rank.LimitedPercentile > 0 {
				fmt.Printf("  Percentile:      %.1f%%\n", rank.LimitedPercentile)
			}
			if rank.LimitedStep > 0 {
				fmt.Printf("  Step:            %d\n", rank.LimitedStep)
			}

			// Display win rate if we have match data
			totalMatches := rank.LimitedMatchesWon + rank.LimitedMatchesLost
			if totalMatches > 0 {
				winRate := float64(rank.LimitedMatchesWon) / float64(totalMatches) * 100
				fmt.Printf("  Matches:         %d-%d (%.1f%% win rate)\n",
					rank.LimitedMatchesWon, rank.LimitedMatchesLost, winRate)
			}
			fmt.Println()
		}
	}

	// Display draft history
	if draftHistory != nil && len(draftHistory.Drafts) > 0 {
		fmt.Println("Draft History")
		fmt.Println("-------------")
		fmt.Printf("Found %d draft/limited event(s)\n\n", len(draftHistory.Drafts))

		for i, draft := range draftHistory.Drafts {
			fmt.Printf("%d. %s\n", i+1, draft.EventName)
			fmt.Printf("   Status: %s\n", draft.Status)
			fmt.Printf("   Record: %d wins", draft.Wins)
			if draft.Losses > 0 {
				fmt.Printf(", %d losses", draft.Losses)
			}
			fmt.Println()

			if draft.Deck.Name != "" {
				fmt.Printf("   Deck: %s\n", draft.Deck.Name)
			}

			if len(draft.Deck.MainDeck) > 0 {
				totalCards := 0
				for _, card := range draft.Deck.MainDeck {
					totalCards += card.Quantity
				}
				fmt.Printf("   Main Deck: %d cards\n", totalCards)
			}

			fmt.Println()
		}
	}

	// Display arena statistics
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		fmt.Println("Arena Statistics (Current Log Session)")
		fmt.Println("---------------------------------------")

		// Overall match stats
		if arenaStats.TotalMatches > 0 {
			matchWinRate := 0.0
			if arenaStats.TotalMatches > 0 {
				matchWinRate = float64(arenaStats.MatchWins) / float64(arenaStats.TotalMatches) * 100
			}
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				arenaStats.MatchWins, arenaStats.MatchLosses, matchWinRate)
		}

		// Overall game stats
		if arenaStats.TotalGames > 0 {
			gameWinRate := 0.0
			if arenaStats.TotalGames > 0 {
				gameWinRate = float64(arenaStats.GameWins) / float64(arenaStats.TotalGames) * 100
			}
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				arenaStats.GameWins, arenaStats.GameLosses, gameWinRate)
		}

		// Format breakdown
		if len(arenaStats.FormatStats) > 0 {
			fmt.Println("\nBy Format:")
			for _, formatStat := range arenaStats.FormatStats {
				fmt.Printf("\n  %s:\n", formatStat.EventName)
				if formatStat.MatchesPlayed > 0 {
					matchWinRate := 0.0
					if formatStat.MatchesPlayed > 0 {
						matchWinRate = float64(formatStat.MatchWins) / float64(formatStat.MatchesPlayed) * 100
					}
					fmt.Printf("    Matches: %d-%d (%.1f%%)\n",
						formatStat.MatchWins, formatStat.MatchLosses, matchWinRate)
				}
				if formatStat.GamesPlayed > 0 {
					gameWinRate := 0.0
					if formatStat.GamesPlayed > 0 {
						gameWinRate = float64(formatStat.GameWins) / float64(formatStat.GamesPlayed) * 100
					}
					fmt.Printf("    Games:   %d-%d (%.1f%%)\n",
						formatStat.GameWins, formatStat.GameLosses, gameWinRate)
				}
			}
		}

		fmt.Println()
	}

	if profile == nil && inventory == nil && rank == nil && draftHistory == nil && arenaStats == nil {
		fmt.Println("No player data found in log file.")
		fmt.Println("Try playing a game or opening MTG Arena to generate log data.")
	}
}

func runMigrationCommand() {
	if len(os.Args) < 3 {
		printMigrationUsage()
		os.Exit(1)
	}

	// Get database path from environment or use default
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		log.Fatalf("Error creating database directory: %v", err)
	}

	// Create migration manager
	mgr, err := storage.NewMigrationManager(dbPath)
	if err != nil {
		log.Fatalf("Error creating migration manager: %v", err)
	}
	defer func() {
		if err := mgr.Close(); err != nil {
			log.Printf("Error closing migration manager: %v", err)
		}
	}()

	command := os.Args[2]

	switch command {
	case "up":
		fmt.Println("Applying all pending migrations...")
		if err := mgr.Up(); err != nil {
			log.Fatalf("Error applying migrations: %v", err)
		}
		version, dirty, err := mgr.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}
		fmt.Println("All migrations applied successfully!")

	case "down":
		fmt.Println("Rolling back last migration...")
		if err := mgr.Down(); err != nil {
			log.Fatalf("Error rolling back migration: %v", err)
		}
		version, dirty, err := mgr.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}
		fmt.Println("Migration rolled back successfully!")

	case "status", "version":
		version, dirty, err := mgr.Version()
		if err != nil {
			log.Fatalf("Error getting version: %v", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty - migration failed or interrupted)\n", version)
			fmt.Println("Use 'migrate force <version>' to recover")
		} else {
			fmt.Printf("Current version: %d\n", version)
		}

	case "force":
		if len(os.Args) < 4 {
			fmt.Println("Error: force command requires a version number")
			fmt.Println("Usage: mtga-companion migrate force <version>")
			os.Exit(1)
		}
		versionStr := os.Args[3]
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		fmt.Printf("Forcing migration version to %d...\n", version)
		fmt.Println("WARNING: This does not run migrations, only sets the version.")
		if err := mgr.Force(version); err != nil {
			log.Fatalf("Error forcing version: %v", err)
		}
		fmt.Println("Version forced successfully!")

	case "goto":
		if len(os.Args) < 4 {
			fmt.Println("Error: goto command requires a version number")
			fmt.Println("Usage: mtga-companion migrate goto <version>")
			os.Exit(1)
		}
		versionStr := os.Args[3]
		version, err := strconv.ParseUint(versionStr, 10, 32)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		fmt.Printf("Migrating to version %d...\n", version)
		if err := mgr.Goto(uint(version)); err != nil {
			log.Fatalf("Error migrating to version %d: %v", version, err)
		}
		fmt.Println("Migration successful!")

	default:
		fmt.Printf("Unknown migration command: %s\n\n", command)
		printMigrationUsage()
		os.Exit(1)
	}
}

func printMigrationUsage() {
	fmt.Println("MTGA Companion - Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mtga-companion migrate <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up                Apply all pending migrations")
	fmt.Println("  down              Rollback the last migration")
	fmt.Println("  status            Show current migration version")
	fmt.Println("  version           Show current migration version (alias for status)")
	fmt.Println("  goto <version>    Migrate to a specific version")
	fmt.Println("  force <version>   Force set migration version (use with caution)")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  MTGA_DB_PATH      Override default database path")
	fmt.Println("                    (default: ~/.mtga-companion/data.db)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mtga-companion migrate up")
	fmt.Println("  mtga-companion migrate status")
	fmt.Println("  mtga-companion migrate goto 1")
	fmt.Println("  MTGA_DB_PATH=/tmp/test.db mtga-companion migrate up")
}

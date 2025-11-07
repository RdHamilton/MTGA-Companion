package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

func main() {
	// Check if this is a migration command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrationCommand()
		return
	}

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, ".mtga-companion", "data.db")
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		log.Fatalf("Error creating database directory: %v", err)
	}

	// Open database with auto-migrate
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create storage service
	service := storage.NewService(db)
	defer func() {
		if err := service.Close(); err != nil {
			log.Printf("Error closing service: %v", err)
		}
	}()

	ctx := context.Background()

	// Get the default log path for the current platform
	logPath, err := logreader.DefaultLogPath()
	if err != nil {
		log.Fatalf("Error getting default log path: %v", err)
	}

	fmt.Printf("MTGA Companion\n")
	fmt.Printf("==============\n\n")
	fmt.Printf("Log file: %s\n", logPath)
	fmt.Printf("Database: %s\n\n", dbPath)

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
	collection, _ := logreader.ParseCollection(entries)
	deckLibrary, _ := logreader.ParseDecks(entries)

	// Store arena stats persistently (with deduplication)
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		fmt.Println("Storing statistics in database...")
		if err := service.StoreArenaStats(ctx, arenaStats, entries); err != nil {
			log.Printf("Warning: Failed to store arena stats: %v", err)
		} else {
			fmt.Println("Statistics stored successfully.")
		}
	}

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

	// Display arena statistics - both current session and all-time
	displayArenaStatistics(arenaStats, service, ctx)

	// Display card collection
	if collection != nil && len(collection.Cards) > 0 {
		fmt.Println()
		displayCollection(collection)
	}

	// Display deck library
	if deckLibrary != nil && len(deckLibrary.Decks) > 0 {
		fmt.Println()
		displayDeckLibrary(deckLibrary)
	}

	if profile == nil && inventory == nil && rank == nil && draftHistory == nil && arenaStats == nil && (collection == nil || len(collection.Cards) == 0) && (deckLibrary == nil || len(deckLibrary.Decks) == 0) {
		fmt.Println("No player data found in log file.")
		fmt.Println("Try playing a game or opening MTG Arena to generate log data.")
	}

	// Start log file poller for real-time updates
	fmt.Println("\nStarting log file poller for real-time updates...")
	pollerConfig := logreader.DefaultPollerConfig(logPath)
	pollerConfig.Interval = 2 * time.Second
	poller, err := logreader.NewPoller(pollerConfig)
	if err != nil {
		log.Printf("Warning: Failed to create poller: %v", err)
		log.Println("Continuing without real-time updates...")
		// Fall back to interactive console without poller
		fmt.Println("\nType 'exit' to quit, or press Enter to refresh statistics.")
		runInteractiveConsole(service, ctx, logPath, nil)
		return
	}
	defer poller.Stop()

	// Start poller
	updates := poller.Start()
	errChan := poller.Errors()

	// Start background goroutine to process updates
	pollerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go processPollerUpdates(pollerCtx, updates, errChan, service, logPath)

	// Interactive console loop
	fmt.Println("\nType 'exit' to quit, or press Enter to refresh statistics.")
	fmt.Println("Statistics will update automatically as new log entries are detected.")
	runInteractiveConsole(service, ctx, logPath, poller)
}

// displayArenaStatistics displays both current session and all-time statistics.
func displayArenaStatistics(arenaStats *logreader.ArenaStats, service *storage.Service, ctx context.Context) {
	// Display weekly and monthly statistics
	displayWeeklyStats(service, ctx)
	displayMonthlyStats(service, ctx)
	// Display current session statistics
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		fmt.Println("Arena Statistics (Current Log Session)")
		fmt.Println("---------------------------------------")

		// Overall match stats
		if arenaStats.TotalMatches > 0 {
			matchWinRate := float64(arenaStats.MatchWins) / float64(arenaStats.TotalMatches) * 100
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				arenaStats.MatchWins, arenaStats.MatchLosses, matchWinRate)
		}

		// Overall game stats
		if arenaStats.TotalGames > 0 {
			gameWinRate := float64(arenaStats.GameWins) / float64(arenaStats.TotalGames) * 100
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				arenaStats.GameWins, arenaStats.GameLosses, gameWinRate)
		}

		// Format breakdown
		if len(arenaStats.FormatStats) > 0 {
			fmt.Println("\nBy Format (Current Session):")
			for _, formatStat := range arenaStats.FormatStats {
				fmt.Printf("\n  %s:\n", formatStat.EventName)
				if formatStat.MatchesPlayed > 0 {
					matchWinRate := float64(formatStat.MatchWins) / float64(formatStat.MatchesPlayed) * 100
					fmt.Printf("    Matches: %d-%d (%.1f%%)\n",
						formatStat.MatchWins, formatStat.MatchLosses, matchWinRate)
				}
				if formatStat.GamesPlayed > 0 {
					gameWinRate := float64(formatStat.GameWins) / float64(formatStat.GamesPlayed) * 100
					fmt.Printf("    Games:   %d-%d (%.1f%%)\n",
						formatStat.GameWins, formatStat.GameLosses, gameWinRate)
				}
			}
		}

		fmt.Println()
	}

	// Display all-time statistics from database
	allTimeStats, err := service.GetStats(ctx, storage.StatsFilter{})
	if err != nil {
		log.Printf("Warning: Failed to retrieve all-time statistics: %v", err)
		return
	}

	if allTimeStats != nil && (allTimeStats.TotalMatches > 0 || allTimeStats.TotalGames > 0) {
		fmt.Println("Arena Statistics (All-Time)")
		fmt.Println("----------------------------")

		if allTimeStats.TotalMatches > 0 {
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				allTimeStats.MatchesWon, allTimeStats.MatchesLost, allTimeStats.WinRate)
		}

		if allTimeStats.TotalGames > 0 {
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				allTimeStats.GamesWon, allTimeStats.GamesLost, allTimeStats.GameWinRate)
		}

		fmt.Println()
	}
}

// displayWeeklyStats displays statistics for the current week.
func displayWeeklyStats(service *storage.Service, ctx context.Context) {
	now := time.Now()
	// Get start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7
	}
	weekStart := now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7)

	startDate := weekStart
	endDate := weekEnd
	filter := storage.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	stats, err := service.GetStats(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve weekly statistics: %v", err)
		return
	}

	if stats != nil && (stats.TotalMatches > 0 || stats.TotalGames > 0) {
		fmt.Println("Arena Statistics (This Week)")
		fmt.Println("----------------------------")
		fmt.Printf("Period: %s to %s\n", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))

		if stats.TotalMatches > 0 {
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				stats.MatchesWon, stats.MatchesLost, stats.WinRate*100)
		}

		if stats.TotalGames > 0 {
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				stats.GamesWon, stats.GamesLost, stats.GameWinRate*100)
		}

		fmt.Println()
	}
}

// displayMonthlyStats displays statistics for the current month.
func displayMonthlyStats(service *storage.Service, ctx context.Context) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	startDate := monthStart
	endDate := monthEnd
	filter := storage.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	stats, err := service.GetStats(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve monthly statistics: %v", err)
		return
	}

	if stats != nil && (stats.TotalMatches > 0 || stats.TotalGames > 0) {
		fmt.Println("Arena Statistics (This Month)")
		fmt.Println("------------------------------")
		fmt.Printf("Period: %s to %s\n", monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))

		if stats.TotalMatches > 0 {
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				stats.MatchesWon, stats.MatchesLost, stats.WinRate*100)
		}

		if stats.TotalGames > 0 {
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				stats.GamesWon, stats.GamesLost, stats.GameWinRate*100)
		}

		fmt.Println()
	}
}

// processPollerUpdates processes new log entries from the poller in the background.
func processPollerUpdates(ctx context.Context, updates <-chan *logreader.LogEntry, errChan <-chan error, service *storage.Service, logPath string) {
	var entryBuffer []*logreader.LogEntry
	ticker := time.NewTicker(5 * time.Second) // Batch process every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-updates:
			if !ok {
				return
			}
			// Buffer entries for batch processing
			entryBuffer = append(entryBuffer, entry)
		case err, ok := <-errChan:
			if !ok {
				return
			}
			log.Printf("Poller error: %v", err)
		case <-ticker.C:
			// Process buffered entries
			if len(entryBuffer) > 0 {
				processNewEntries(ctx, entryBuffer, service)
				entryBuffer = nil // Clear buffer
			}
		}
	}
}

// processNewEntries processes new log entries and updates statistics.
func processNewEntries(ctx context.Context, entries []*logreader.LogEntry, service *storage.Service) {
	// Parse arena stats from new entries
	arenaStats, err := logreader.ParseArenaStats(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse arena stats from new entries: %v", err)
		return
	}

	// Store new stats if we have match data
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		if err := service.StoreArenaStats(ctx, arenaStats, entries); err != nil {
			log.Printf("Warning: Failed to store arena stats from poller: %v", err)
		} else {
			log.Printf("Updated statistics: %d new matches, %d new games",
				arenaStats.TotalMatches, arenaStats.TotalGames)
		}
	}
}

// runInteractiveConsole runs an interactive console loop that waits for user input.
func runInteractiveConsole(service *storage.Service, ctx context.Context, logPath string, poller *logreader.Poller) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			// Refresh statistics on empty input
			refreshStatistics(service, ctx, logPath)
			continue
		}

		command := strings.ToLower(input)
		switch command {
		case "exit", "quit", "q":
			fmt.Println("Stopping poller...")
			if poller != nil {
				poller.Stop()
			}
			fmt.Println("Goodbye!")
			return
		case "refresh", "r":
			refreshStatistics(service, ctx, logPath)
		case "weekly", "week", "w":
			displayWeeklyStats(service, ctx)
		case "monthly", "month", "m":
			displayMonthlyStats(service, ctx)
		case "collection", "col", "c":
			// Refresh collection from log file
			refreshCollection(ctx, logPath)
		case "decks", "deck", "d":
			// Refresh decks from log file
			refreshDecks(ctx, logPath)
		case "help", "h":
			printHelp()
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", input)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

// refreshStatistics refreshes and displays statistics from the log file.
func refreshStatistics(service *storage.Service, ctx context.Context, logPath string) {
	fmt.Println("\nRefreshing statistics...")

	// Create a reader
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		fmt.Printf("Error creating log reader: %v\n", err)
		return
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader: %v", err)
		}
	}()

	// Read all JSON entries
	entries, err := reader.ReadAllJSON()
	if err != nil {
		fmt.Printf("Error reading log entries: %v\n", err)
		return
	}

	// Parse arena stats
	arenaStats, err := logreader.ParseArenaStats(entries)
	if err != nil {
		fmt.Printf("Error parsing arena stats: %v\n", err)
		return
	}

	// Store new stats
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		if err := service.StoreArenaStats(ctx, arenaStats, entries); err != nil {
			fmt.Printf("Warning: Failed to store arena stats: %v\n", err)
		}
	}

	// Display updated statistics
	displayArenaStatistics(arenaStats, service, ctx)
}

// refreshCollection refreshes and displays collection from the log file.
func refreshCollection(ctx context.Context, logPath string) {
	fmt.Println("\nRefreshing collection...")

	// Create a reader
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		fmt.Printf("Error creating log reader: %v\n", err)
		return
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader: %v", err)
		}
	}()

	// Read all JSON entries
	entries, err := reader.ReadAllJSON()
	if err != nil {
		fmt.Printf("Error reading log entries: %v\n", err)
		return
	}

	// Parse collection
	collection, err := logreader.ParseCollection(entries)
	if err != nil {
		fmt.Printf("Error parsing collection: %v\n", err)
		return
	}

	// Display collection
	if collection != nil && len(collection.Cards) > 0 {
		displayCollection(collection)
	} else {
		fmt.Println("No collection data found in log file.")
	}
}

// refreshDecks refreshes and displays decks from the log file.
func refreshDecks(ctx context.Context, logPath string) {
	fmt.Println("\nRefreshing decks...")

	// Create a reader
	reader, err := logreader.NewReader(logPath)
	if err != nil {
		fmt.Printf("Error creating log reader: %v\n", err)
		return
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader: %v", err)
		}
	}()

	// Read all JSON entries
	entries, err := reader.ReadAllJSON()
	if err != nil {
		fmt.Printf("Error reading log entries: %v\n", err)
		return
	}

	// Parse decks
	deckLibrary, err := logreader.ParseDecks(entries)
	if err != nil {
		fmt.Printf("Error parsing decks: %v\n", err)
		return
	}

	// Display decks
	if deckLibrary != nil && len(deckLibrary.Decks) > 0 {
		displayDeckLibrary(deckLibrary)
	} else {
		fmt.Println("No deck data found in log file.")
	}
}

// printHelp displays available commands.
func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  (empty)    - Refresh and display statistics")
	fmt.Println("  refresh, r - Refresh and display statistics")
	fmt.Println("  weekly, week, w - Display weekly statistics")
	fmt.Println("  monthly, month, m - Display monthly statistics")
	fmt.Println("  collection, col, c - Refresh and display card collection")
	fmt.Println("  decks, deck, d - Refresh and display saved decks")
	fmt.Println("  exit, quit, q - Exit the application")
	fmt.Println("  help, h    - Show this help message")
	fmt.Println()
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

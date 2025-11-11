package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/charts"
	"github.com/ramonehamilton/MTGA-Companion/internal/display"
	"github.com/ramonehamilton/MTGA-Companion/internal/export"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cardlookup"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/draftdata"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/imagecache"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/importer"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/migrations"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/query"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/unified"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/updater"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/viewer"
)

var (
	pollInterval  = flag.Duration("poll-interval", 2*time.Second, "Interval for polling log file (e.g., 1s, 2s, 5s)")
	enableMetrics = flag.Bool("enable-metrics", false, "Enable poller performance metrics collection")
	useFileEvents = flag.Bool("use-file-events", true, "Use file system events (fsnotify) for monitoring")
	useGUI        = flag.Bool("gui", false, "Launch GUI mode instead of CLI")
)

// getDBPath returns the database path from environment variable or default location.
func getDBPath() string {
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}
	return dbPath
}

func main() {
	// Parse flags before checking for subcommands
	flag.Parse()

	// Validate poll interval
	if *pollInterval < 1*time.Second {
		log.Fatalf("Poll interval must be at least 1 second, got %v", *pollInterval)
	}
	if *pollInterval > 1*time.Minute {
		log.Fatalf("Poll interval must be at most 1 minute, got %v", *pollInterval)
	}

	// Check if this is a migration command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrationCommand()
		return
	}

	// Check if this is a backup command
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		runBackupCommand()
		return
	}

	// Check if this is a draft command
	if len(os.Args) > 1 && os.Args[1] == "draft" {
		runDraftCommand()
		return
	}

	// Check if this is a cards command
	if len(os.Args) > 1 && os.Args[1] == "cards" {
		runCardsCommand()
		return
	}

	// Check if this is a draft-stats command
	if len(os.Args) > 1 && os.Args[1] == "draft-stats" {
		runDraftStatsCommand()
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

	// Check if GUI mode is requested
	if *useGUI {
		guiApp := gui.NewApp(service)
		guiApp.Run()
		return
	}

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
	_, _ = logreader.ParseDraftPicks(entries) // Parse draft picks (used in refreshDraftPicks)
	arenaStats, _ := logreader.ParseArenaStats(entries)
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

	// Display draft history with card names
	if draftHistory != nil && len(draftHistory.Drafts) > 0 {
		fmt.Println()
		displayDraftHistory(draftHistory)
	}

	// Display draft statistics
	if draftHistory != nil && len(draftHistory.Drafts) > 0 {
		draftStats := logreader.CalculateDraftStatistics(draftHistory)
		if draftStats != nil {
			fmt.Println()
			displayDraftStatistics(draftStats)
		}
	}

	// Display arena statistics - both current session and all-time
	displayArenaStatistics(arenaStats, service, ctx)

	// Display deck library
	if deckLibrary != nil && len(deckLibrary.Decks) > 0 {
		fmt.Println()
		displayDeckLibrary(deckLibrary)
	}

	if profile == nil && inventory == nil && rank == nil && draftHistory == nil && arenaStats == nil && (deckLibrary == nil || len(deckLibrary.Decks) == 0) {
		fmt.Println("No player data found in log file.")
		fmt.Println("Try playing a game or opening MTG Arena to generate log data.")
	}

	// Start log file poller for real-time updates
	fmt.Println("\nStarting log file poller for real-time updates...")
	pollerConfig := logreader.DefaultPollerConfig(logPath)
	pollerConfig.Interval = *pollInterval
	pollerConfig.UseFileEvents = *useFileEvents
	pollerConfig.EnableMetrics = *enableMetrics
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
	allTimeStats, err := service.GetStats(ctx, models.StatsFilter{})
	if err != nil {
		log.Printf("Warning: Failed to retrieve all-time statistics: %v", err)
		return
	}

	if allTimeStats != nil && (allTimeStats.TotalMatches > 0 || allTimeStats.TotalGames > 0) {
		fmt.Println("Arena Statistics (All-Time)")
		fmt.Println("----------------------------")

		if allTimeStats.TotalMatches > 0 {
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				allTimeStats.MatchesWon, allTimeStats.MatchesLost, allTimeStats.WinRate*100)
		}

		if allTimeStats.TotalGames > 0 {
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				allTimeStats.GamesWon, allTimeStats.GamesLost, allTimeStats.GameWinRate*100)
		}

		fmt.Println()

		// Display result breakdown
		displayResultReasons(service, ctx)

		// Display streaks (all-time)
		displayStreakStats(service, ctx, models.StatsFilter{})

		// Display performance metrics (all-time)
		displayPerformanceMetrics(service, ctx, models.StatsFilter{})
	}
}

// displayWeeklyStats displays statistics for the current week.
func displayWeeklyStats(service *storage.Service, ctx context.Context) {
	displayWeeklyStatsWithOffset(service, ctx, 0)
}

// displayWeeklyStatsWithOffset displays statistics for a week with an offset.
// offset = 0 means current week, -1 means last week, etc.
func displayWeeklyStatsWithOffset(service *storage.Service, ctx context.Context, offset int) {
	// Import stats package is needed - add at top of file
	// For now, calculate inline (will refactor imports later)

	// Get week range using same logic as stats.WeekRange
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	currentWeekStart := now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
	weekStart := currentWeekStart.AddDate(0, 0, offset*7)
	weekEnd := weekStart.AddDate(0, 0, 7)

	startDate := weekStart
	endDate := weekEnd
	filter := models.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	stats, err := service.GetStats(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve weekly statistics: %v", err)
		return
	}

	// Get label for the week
	var label string
	switch offset {
	case 0:
		label = "This Week"
	case -1:
		label = "Last Week"
	case -2:
		label = "Two Weeks Ago"
	default:
		if offset < 0 {
			label = fmt.Sprintf("%d Weeks Ago", -offset)
		} else {
			label = fmt.Sprintf("%d Weeks From Now", offset)
		}
	}

	if stats != nil && (stats.TotalMatches > 0 || stats.TotalGames > 0) {
		fmt.Printf("Arena Statistics (%s)\n", label)
		fmt.Println(strings.Repeat("-", len("Arena Statistics ("+label+")")))
		fmt.Printf("Period: %s to %s\n", weekStart.Format("2006-01-02"), weekEnd.AddDate(0, 0, -1).Format("2006-01-02"))

		if stats.TotalMatches > 0 {
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				stats.MatchesWon, stats.MatchesLost, stats.WinRate*100)
		}

		if stats.TotalGames > 0 {
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				stats.GamesWon, stats.GamesLost, stats.GameWinRate*100)
		}

		fmt.Println()
	} else {
		fmt.Printf("\nNo matches found for %s\n", label)
		fmt.Printf("Period: %s to %s\n\n", weekStart.Format("2006-01-02"), weekEnd.AddDate(0, 0, -1).Format("2006-01-02"))
	}
}

// displayMonthlyStats displays statistics for the current month.
func displayMonthlyStats(service *storage.Service, ctx context.Context) {
	displayMonthlyStatsWithOffset(service, ctx, 0)
}

// displayMonthlyStatsWithOffset displays statistics for a month with an offset.
// offset = 0 means current month, -1 means last month, etc.
func displayMonthlyStatsWithOffset(service *storage.Service, ctx context.Context, offset int) {
	now := time.Now()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthStart := currentMonthStart.AddDate(0, offset, 0)
	monthEnd := monthStart.AddDate(0, 1, 0)

	startDate := monthStart
	endDate := monthEnd
	filter := models.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	stats, err := service.GetStats(ctx, filter)
	if err != nil {
		log.Printf("Warning: Failed to retrieve monthly statistics: %v", err)
		return
	}

	// Get label for the month
	var label string
	switch offset {
	case 0:
		label = "This Month"
	case -1:
		label = "Last Month"
	case -2:
		label = "Two Months Ago"
	default:
		if offset < 0 {
			label = fmt.Sprintf("%d Months Ago", -offset)
		} else {
			label = fmt.Sprintf("%d Months From Now", offset)
		}
	}

	monthName := monthStart.Format("January 2006")

	if stats != nil && (stats.TotalMatches > 0 || stats.TotalGames > 0) {
		fmt.Printf("Arena Statistics (%s - %s)\n", label, monthName)
		fmt.Println(strings.Repeat("-", len("Arena Statistics ("+label+" - "+monthName+")")))
		fmt.Printf("Period: %s to %s\n", monthStart.Format("2006-01-02"), monthEnd.AddDate(0, 0, -1).Format("2006-01-02"))

		if stats.TotalMatches > 0 {
			fmt.Printf("Matches: %d-%d (%.1f%% win rate)\n",
				stats.MatchesWon, stats.MatchesLost, stats.WinRate*100)
		}

		if stats.TotalGames > 0 {
			fmt.Printf("Games:   %d-%d (%.1f%% win rate)\n",
				stats.GamesWon, stats.GamesLost, stats.GameWinRate*100)
		}

		fmt.Println()
	} else {
		fmt.Printf("\nNo matches found for %s (%s)\n", label, monthName)
		fmt.Printf("Period: %s to %s\n\n", monthStart.Format("2006-01-02"), monthEnd.AddDate(0, 0, -1).Format("2006-01-02"))
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

		// Parse command and arguments
		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

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
			// Support offset parameter: weekly -1 (last week), weekly -2 (two weeks ago)
			offset := 0
			if len(parts) > 1 {
				if parsedOffset, err := strconv.Atoi(parts[1]); err == nil {
					offset = parsedOffset
				}
			}
			displayWeeklyStatsWithOffset(service, ctx, offset)
		case "monthly", "month", "m":
			// Support offset parameter: monthly -1 (last month), monthly -2 (two months ago)
			offset := 0
			if len(parts) > 1 {
				if parsedOffset, err := strconv.Atoi(parts[1]); err == nil {
					offset = parsedOffset
				}
			}
			displayMonthlyStatsWithOffset(service, ctx, offset)
		case "setcomp", "sets", "completion":
			// Display set completion percentages
			displaySetCompletion(service, ctx)
		case "decks", "deck", "d":
			// Refresh decks from log file
			refreshDecks(ctx, logPath)
		case "trend", "trends", "t":
			// Display trend analysis
			displayTrendAnalysisForPeriod(service, ctx, 30, "weekly")
		case "results", "result", "res":
			// Display result breakdown
			displayResultReasons(service, ctx)
		case "rank", "ranks", "rankprog":
			// Display rank progression
			displayRankProgressionWithStats(service, ctx)
		case "rankhistory", "rh", "rankh":
			// Display rank history from rank_history table
			if len(parts) < 2 {
				displayLatestRank(service, ctx)
			} else {
				subCmd := parts[1]
				switch subCmd {
				case "all", "history":
					displayRankHistory(service, ctx)
				case "constructed", "c":
					displayRankHistoryByFormat(service, ctx, "constructed")
				case "limited", "l":
					displayRankHistoryByFormat(service, ctx, "limited")
				case "season", "s":
					if len(parts) >= 3 {
						displayRankHistoryBySeason(service, ctx, parts[2])
					} else {
						fmt.Println("Usage: rankhistory season <season_number>")
					}
				case "latest", "current":
					displayLatestRank(service, ctx)
				default:
					fmt.Println("Unknown rankhistory command. Usage:")
					fmt.Println("  rankhistory                     - Show current rank for all formats")
					fmt.Println("  rankhistory all                 - Show all rank history")
					fmt.Println("  rankhistory constructed         - Show constructed rank history")
					fmt.Println("  rankhistory limited             - Show limited rank history")
					fmt.Println("  rankhistory season <number>     - Show rank history for a specific season")
					fmt.Println("  rankhistory latest              - Show current rank (same as 'rankhistory')")
				}
			}
		case "draft", "drafts", "draftstats":
			// Display draft statistics
			refreshDraftStatistics(ctx, logPath)
		case "draftpicks", "picks":
			// Display draft picks
			refreshDraftPicks(ctx, logPath)
		case "seasons", "season", "seasonal":
			// Display seasonal rank progression
			format := "constructed" // default
			if len(parts) > 1 {
				if parts[1] == "limited" || parts[1] == "l" {
					format = "limited"
				} else if parts[1] == "constructed" || parts[1] == "c" {
					format = "constructed"
				} else if parts[1] == "compare" {
					// Season comparison view
					compareFormat := "constructed"
					if len(parts) > 2 && (parts[2] == "limited" || parts[2] == "l") {
						compareFormat = "limited"
					}
					displaySeasonComparison(service, ctx, compareFormat)
					continue
				}
			}
			displaySeasonalProgression(service, ctx, format)
		case "achievements", "achieve", "ach":
			// Display rank achievements
			format := "constructed" // default
			if len(parts) > 1 {
				if parts[1] == "limited" || parts[1] == "l" {
					format = "limited"
				}
			}
			displayRankAchievements(service, ctx, format)
		case "progress", "prog", "p":
			// Display rank progression and estimated matches
			format := "constructed" // default
			if len(parts) > 1 {
				if parts[1] == "limited" || parts[1] == "l" {
					format = "limited"
				}
			}
			displayRankProgressionAnalysis(service, ctx, format)
		case "floors", "floor", "f":
			// Display rank floors
			format := "constructed" // default
			if len(parts) > 1 {
				if parts[1] == "limited" || parts[1] == "l" {
					format = "limited"
				}
			}
			displayRankFloors(service, format)
		case "doublerankups", "dru", "doubleup":
			// Display double rank up events
			format := "constructed" // default
			if len(parts) > 1 {
				if parts[1] == "limited" || parts[1] == "l" {
					format = "limited"
				}
			}
			displayDoubleRankUps(service, ctx, format)
		case "backup", "b":
			runBackupCommandInteractive(service, ctx)
		case "account", "accounts", "acc", "a":
			handleAccountCommand(service, ctx, input)
		case "switch", "sw":
			handleAccountSwitch(service, ctx, input)
		case "recent", "recentmatches":
			// Display recent matches with optional limit
			limit := 10 // default
			if len(parts) > 1 {
				if parsedLimit, err := strconv.Atoi(parts[1]); err == nil && parsedLimit > 0 {
					limit = parsedLimit
				}
			}
			displayRecentMatches(service, ctx, limit)
		case "match", "matchview", "view":
			// Display detailed match information
			if len(parts) < 2 {
				fmt.Println("Usage: match view <match_id|latest>")
				fmt.Println("Examples:")
				fmt.Println("  match view latest              - View most recent match")
				fmt.Println("  match view 12345-abc-67890     - View specific match by ID")
			} else if parts[1] == "view" && len(parts) >= 3 {
				if parts[2] == "latest" {
					displayLatestMatch(service, ctx)
				} else {
					matchID := strings.Join(parts[2:], " ")
					displayMatchDetails(service, ctx, matchID)
				}
			} else if parts[1] == "latest" {
				displayLatestMatch(service, ctx)
			} else {
				matchID := strings.Join(parts[1:], " ")
				displayMatchDetails(service, ctx, matchID)
			}
		case "format":
			// Display match history for specific format
			if len(parts) < 2 {
				fmt.Println("Usage: format <format_name>")
				fmt.Println("Example: format Standard")
			} else {
				formatName := strings.Join(parts[1:], " ")
				displayMatchesByFormat(service, ctx, formatName)
			}
		case "formats", "byformat", "formatstats":
			// Display statistics grouped by format
			displayStatsByFormat(service, ctx)
		case "deckstats", "bydeck", "deckperf":
			// Display statistics grouped by deck
			displayStatsByDeck(service, ctx)
		case "daterange", "range":
			// Display statistics for date range
			if len(parts) < 3 {
				fmt.Println("Usage: daterange <start_date> <end_date>")
				fmt.Println("Example: daterange 2024-01-01 2024-01-31")
			} else {
				displayDateRangeStats(service, ctx, parts[1], parts[2])
			}
		case "export", "exp":
			// Export data to CSV/JSON
			handleExportCommand(service, ctx, parts[1:])
		case "chart", "charts":
			// Display trend charts
			handleTrendsCommand(service, ctx, parts[1:])
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

// refreshDraftStatistics refreshes and displays draft statistics.
func refreshDraftStatistics(ctx context.Context, logPath string) {
	fmt.Println("\nRefreshing draft statistics...")

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

	// Parse draft history
	draftHistory, err := logreader.ParseDraftHistory(entries)
	if err != nil {
		fmt.Printf("Error parsing draft history: %v\n", err)
		return
	}

	// Calculate and display statistics
	if draftHistory != nil && len(draftHistory.Drafts) > 0 {
		draftStats := logreader.CalculateDraftStatistics(draftHistory)
		if draftStats != nil {
			displayDraftStatistics(draftStats)
		} else {
			fmt.Println("No draft statistics available.")
		}
	} else {
		fmt.Println("No draft history found in log file.")
	}
}

// refreshDraftPicks refreshes and displays draft picks.
func refreshDraftPicks(ctx context.Context, logPath string) {
	fmt.Println("\nRefreshing draft picks...")

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

	// Parse draft picks
	draftPicks, err := logreader.ParseDraftPicks(entries)
	if err != nil {
		fmt.Printf("Error parsing draft picks: %v\n", err)
		return
	}

	// Display draft picks
	if len(draftPicks) > 0 {
		for _, picks := range draftPicks {
			displayDraftPicks(picks)
		}
	} else {
		fmt.Println("No draft picks found in log file.")
	}
}

// printHelp displays available commands.
// handleAccountCommand handles account management commands.
func handleAccountCommand(service *storage.Service, ctx context.Context, input string) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		// List all accounts
		accounts, err := service.GetAllAccounts(ctx)
		if err != nil {
			fmt.Printf("Error getting accounts: %v\n", err)
			return
		}

		currentAccount, err := service.GetCurrentAccount(ctx)
		if err != nil {
			fmt.Printf("Error getting current account: %v\n", err)
			return
		}

		fmt.Println("\n=== Accounts ===")
		for _, account := range accounts {
			marker := " "
			if account.ID == currentAccount.ID {
				marker = "*"
			}
			defaultMarker := ""
			if account.IsDefault {
				defaultMarker = " (default)"
			}
			screenName := ""
			if account.ScreenName != nil {
				screenName = fmt.Sprintf(" - %s", *account.ScreenName)
			}
			fmt.Printf("%s [%d] %s%s%s\n", marker, account.ID, account.Name, screenName, defaultMarker)
		}
		fmt.Println()
		return
	}

	command := strings.ToLower(parts[1])
	switch command {
	case "create", "new", "add":
		handleAccountCreate(service, ctx, parts[2:])
	case "info", "show":
		handleAccountInfo(service, ctx, parts[2:])
	case "list", "ls":
		handleAccountList(service, ctx)
	default:
		fmt.Printf("Unknown account command: %s\n", command)
		fmt.Println("Available commands: create, info, list")
	}
}

// handleAccountCreate creates a new account.
func handleAccountCreate(service *storage.Service, ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Print("Enter account name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return
		}
		name := strings.TrimSpace(scanner.Text())
		if name == "" {
			fmt.Println("Account name cannot be empty")
			return
		}

		account, err := service.CreateAccount(ctx, name, nil, nil)
		if err != nil {
			fmt.Printf("Error creating account: %v\n", err)
			return
		}

		fmt.Printf("Account created: [%d] %s\n", account.ID, account.Name)
		return
	}

	name := strings.Join(args, " ")
	account, err := service.CreateAccount(ctx, name, nil, nil)
	if err != nil {
		fmt.Printf("Error creating account: %v\n", err)
		return
	}

	fmt.Printf("Account created: [%d] %s\n", account.ID, account.Name)
}

// handleAccountInfo displays information about an account.
func handleAccountInfo(service *storage.Service, ctx context.Context, args []string) {
	var accountID int
	var err error

	if len(args) < 1 {
		// Show current account
		account, err := service.GetCurrentAccount(ctx)
		if err != nil {
			fmt.Printf("Error getting current account: %v\n", err)
			return
		}
		displayAccountInfo(account, service, ctx)
		return
	}

	accountID, err = strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("Invalid account ID: %s\n", args[0])
		return
	}

	account, err := service.GetAccount(ctx, accountID)
	if err != nil {
		fmt.Printf("Error getting account: %v\n", err)
		return
	}
	if account == nil {
		fmt.Printf("Account not found: %d\n", accountID)
		return
	}

	displayAccountInfo(account, service, ctx)
}

// displayAccountInfo displays detailed information about an account.
func displayAccountInfo(account *models.Account, service *storage.Service, ctx context.Context) {
	fmt.Printf("\n=== Account Information ===\n")
	fmt.Printf("ID: %d\n", account.ID)
	fmt.Printf("Name: %s\n", account.Name)
	if account.ScreenName != nil {
		fmt.Printf("Screen Name: %s\n", *account.ScreenName)
	}
	if account.ClientID != nil {
		fmt.Printf("Client ID: %s\n", *account.ClientID)
	}
	fmt.Printf("Default: %v\n", account.IsDefault)
	fmt.Printf("Created: %s\n", account.CreatedAt.Format("2006-01-02 15:04:05"))

	// Get statistics for this account
	accountID := account.ID
	filter := models.StatsFilter{
		AccountID: &accountID,
	}
	stats, err := service.GetStats(ctx, filter)
	if err != nil {
		fmt.Printf("Error getting statistics: %v\n", err)
		return
	}

	fmt.Println("\n=== Statistics ===")
	fmt.Printf("Total Matches: %d\n", stats.TotalMatches)
	fmt.Printf("Matches Won: %d\n", stats.MatchesWon)
	fmt.Printf("Matches Lost: %d\n", stats.MatchesLost)
	if stats.TotalMatches > 0 {
		fmt.Printf("Win Rate: %.2f%%\n", stats.WinRate*100)
	}
	fmt.Printf("Total Games: %d\n", stats.TotalGames)
	fmt.Printf("Games Won: %d\n", stats.GamesWon)
	fmt.Printf("Games Lost: %d\n", stats.GamesLost)
	if stats.TotalGames > 0 {
		fmt.Printf("Game Win Rate: %.2f%%\n", stats.GameWinRate*100)
	}
	fmt.Println()
}

// handleAccountList lists all accounts.
func handleAccountList(service *storage.Service, ctx context.Context) {
	accounts, err := service.GetAllAccounts(ctx)
	if err != nil {
		fmt.Printf("Error getting accounts: %v\n", err)
		return
	}

	currentAccount, err := service.GetCurrentAccount(ctx)
	if err != nil {
		fmt.Printf("Error getting current account: %v\n", err)
		return
	}

	fmt.Println("\n=== Accounts ===")
	for _, account := range accounts {
		marker := " "
		if account.ID == currentAccount.ID {
			marker = "*"
		}
		defaultMarker := ""
		if account.IsDefault {
			defaultMarker = " (default)"
		}
		screenName := ""
		if account.ScreenName != nil {
			screenName = fmt.Sprintf(" - %s", *account.ScreenName)
		}
		fmt.Printf("%s [%d] %s%s%s\n", marker, account.ID, account.Name, screenName, defaultMarker)
	}
	fmt.Println()
}

// handleAccountSwitch switches to a different account.
func handleAccountSwitch(service *storage.Service, ctx context.Context, input string) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		fmt.Print("Enter account ID to switch to: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return
		}
		accountIDStr := strings.TrimSpace(scanner.Text())
		accountID, err := strconv.Atoi(accountIDStr)
		if err != nil {
			fmt.Printf("Invalid account ID: %s\n", accountIDStr)
			return
		}

		if err := service.SetCurrentAccount(ctx, accountID); err != nil {
			fmt.Printf("Error switching account: %v\n", err)
			return
		}

		account, err := service.GetCurrentAccount(ctx)
		if err != nil {
			fmt.Printf("Error getting account: %v\n", err)
			return
		}

		fmt.Printf("Switched to account: [%d] %s\n", account.ID, account.Name)
		return
	}

	accountID, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("Invalid account ID: %s\n", parts[1])
		return
	}

	if err := service.SetCurrentAccount(ctx, accountID); err != nil {
		fmt.Printf("Error switching account: %v\n", err)
		return
	}

	account, err := service.GetCurrentAccount(ctx)
	if err != nil {
		fmt.Printf("Error getting account: %v\n", err)
		return
	}

	fmt.Printf("Switched to account: [%d] %s\n", account.ID, account.Name)
}

func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  (empty)    - Refresh and display statistics")
	fmt.Println("  refresh, r - Refresh and display statistics")
	fmt.Println("  weekly [offset] - Display weekly statistics (offset: 0=this week, -1=last week)")
	fmt.Println("  monthly [offset] - Display monthly statistics (offset: 0=this month, -1=last month)")
	fmt.Println("  daterange <start> <end> - Display statistics for date range (YYYY-MM-DD)")
	fmt.Println("  formats, byformat - Display statistics grouped by format")
	fmt.Println("  deckstats, bydeck, deckperf - Display statistics grouped by deck")
	fmt.Println("  recent [limit] - Display recent matches (default: 10)")
	fmt.Println("  format <name> - Display match history for specific format")
	fmt.Println("  match view <id|latest> - Display detailed match information")
	fmt.Println("  collection, col, c - Refresh and display card collection")
	fmt.Println("  setcomp, sets, completion - Display set completion percentages")
	fmt.Println("  decks, deck, d - Refresh and display saved decks")
	fmt.Println("  trend, trends, t - Display historical trend analysis")
	fmt.Println("  chart, charts - Display visual trend charts (type 'chart help' for details)")
	fmt.Println("  results, result, res - Display match result breakdown")
	fmt.Println("  rank, ranks, rankprog - Display rank progression and tier statistics")
	fmt.Println("  draft, drafts, draftstats - Display draft statistics")
	fmt.Println("  draftpicks, picks - Display draft picks")
	fmt.Println("  account, accounts, a - Manage accounts (list, create, info)")
	fmt.Println("  switch, sw <id> - Switch to a different account")
	fmt.Println("  backup, b  - Create or manage database backups")
	fmt.Println("  export, exp - Export data to CSV/JSON (type 'export' for details)")
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

// runBackupCommand handles backup and restore commands.
func runBackupCommand() {
	// Get database path from environment or use default
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Check if database exists (except for list command which doesn't need it)
	if len(os.Args) >= 3 && os.Args[2] != "list" && os.Args[2] != "ls" {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			log.Fatalf("Database file does not exist: %s", dbPath)
		}
	}

	// Create backup manager
	backupMgr := storage.NewBackupManager(dbPath)

	if len(os.Args) < 3 {
		printBackupUsage()
		os.Exit(1)
	}

	command := os.Args[2]

	switch command {
	case "create", "backup":
		// Define flags for create command
		createFlags := flag.NewFlagSet("create", flag.ExitOnError)
		backupType := createFlags.String("type", "full", "Backup type: 'full' or 'incremental'")
		backupDir := createFlags.String("dir", os.Getenv("MTGA_BACKUP_DIR"), "Backup directory")
		backupName := createFlags.String("name", "", "Backup name (default: auto-generated timestamp)")
		compress := createFlags.Bool("compress", false, "Compress backup with gzip")
		encrypt := createFlags.Bool("encrypt", false, "Encrypt backup")
		passwordEnv := createFlags.String("password-env", "", "Environment variable containing encryption password")
		verify := createFlags.Bool("verify", true, "Verify backup after creation")

		if err := createFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		// Build backup config
		config := storage.DefaultBackupConfig()
		config.BackupDir = *backupDir
		config.BackupName = *backupName
		config.VerifyBackup = *verify
		config.Compress = *compress
		config.Encrypt = *encrypt

		// Set backup type
		switch *backupType {
		case "full":
			config.BackupType = storage.BackupTypeFull
		case "incremental", "incr":
			config.BackupType = storage.BackupTypeIncremental
		default:
			log.Fatalf("Invalid backup type: %s (must be 'full' or 'incremental')", *backupType)
		}

		// Handle encryption password
		if *encrypt {
			if *passwordEnv == "" {
				log.Fatal("Error: --password-env is required when --encrypt is specified")
			}
			password := os.Getenv(*passwordEnv)
			if password == "" {
				log.Fatalf("Error: environment variable %s is not set or empty", *passwordEnv)
			}
			config.EncryptionPassword = password
		}

		// Print configuration
		fmt.Printf("Creating %s backup...\n", *backupType)
		if *compress {
			fmt.Println("  Compression: enabled")
		}
		if *encrypt {
			fmt.Println("  Encryption: enabled")
		}

		backupPath, err := backupMgr.Backup(config)
		if err != nil {
			log.Fatalf("Error creating backup: %v", err)
		}

		fmt.Printf("\n✓ Backup created successfully: %s\n", backupPath)

		// Display backup size
		info, err := os.Stat(backupPath)
		if err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("  Size: %.2f MB\n", sizeMB)
		}

	case "restore":
		// Define flags for restore command
		restoreFlags := flag.NewFlagSet("restore", flag.ExitOnError)
		passwordEnv := restoreFlags.String("password-env", "", "Environment variable containing decryption password")
		noConfirm := restoreFlags.Bool("yes", false, "Skip confirmation prompt")

		if err := restoreFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if restoreFlags.NArg() < 1 {
			fmt.Println("Error: restore command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup restore <backup-file> [flags]")
			fmt.Println("\nFlags:")
			restoreFlags.PrintDefaults()
			os.Exit(1)
		}
		backupPath := restoreFlags.Arg(0)

		// Check if backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			log.Fatalf("Backup file does not exist: %s", backupPath)
		}

		// Show warning and get confirmation
		if !*noConfirm {
			fmt.Println("WARNING: This will overwrite the current database!")
			fmt.Printf("Database: %s\n", dbPath)
			fmt.Printf("Backup:   %s\n", backupPath)
			fmt.Print("\nAre you sure you want to continue? (yes/no): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Error reading input: %v", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "yes" && response != "y" {
				fmt.Println("Restore cancelled.")
				return
			}
		}

		fmt.Println("\nRestoring database from backup...")

		// Handle decryption password if needed
		var password string
		if *passwordEnv != "" {
			password = os.Getenv(*passwordEnv)
			if password == "" {
				log.Fatalf("Error: environment variable %s is not set or empty", *passwordEnv)
			}
		}

		// Restore with optional password
		if password != "" {
			if err := backupMgr.Restore(backupPath, password); err != nil {
				log.Fatalf("Error restoring backup: %v", err)
			}
		} else {
			if err := backupMgr.Restore(backupPath); err != nil {
				log.Fatalf("Error restoring backup: %v", err)
			}
		}

		fmt.Println("✓ Database restored successfully!")

	case "list", "ls":
		// Define flags for list command
		listFlags := flag.NewFlagSet("list", flag.ExitOnError)
		format := listFlags.String("format", "table", "Output format: 'table' or 'json'")
		backupDir := listFlags.String("dir", os.Getenv("MTGA_BACKUP_DIR"), "Backup directory")

		if err := listFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if *backupDir == "" {
			*backupDir = backupMgr.GetBackupDir()
		}

		backups, err := backupMgr.ListBackups(*backupDir)
		if err != nil {
			log.Fatalf("Error listing backups: %v", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return
		}

		// Format output
		switch *format {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(backups); err != nil {
				log.Fatalf("Error encoding JSON: %v", err)
			}
		case "table":
			fmt.Printf("\nFound %d backup(s) in %s:\n\n", len(backups), *backupDir)
			for i, backup := range backups {
				sizeMB := float64(backup.Size) / (1024 * 1024)
				fmt.Printf("%d. %s\n", i+1, backup.Name)
				fmt.Printf("   Path:     %s\n", backup.Path)
				fmt.Printf("   Size:     %.2f MB\n", sizeMB)
				fmt.Printf("   Modified: %s\n", backup.ModTime.Format("2006-01-02 15:04:05"))
				fmt.Printf("   Checksum: %s\n", backup.Checksum)
				fmt.Println()
			}
		default:
			log.Fatalf("Invalid format: %s (must be 'table' or 'json')", *format)
		}

	case "verify":
		// Define flags for verify command
		verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
		if err := verifyFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if verifyFlags.NArg() < 1 {
			fmt.Println("Error: verify command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup verify <backup-file>")
			os.Exit(1)
		}
		backupPath := verifyFlags.Arg(0)

		fmt.Printf("Verifying backup: %s\n", backupPath)
		if err := backupMgr.VerifyBackup(backupPath); err != nil {
			log.Fatalf("Backup verification failed: %v", err)
		}

		fmt.Println("✓ Backup verification successful!")

	case "cleanup":
		// Define flags for cleanup command
		cleanupFlags := flag.NewFlagSet("cleanup", flag.ExitOnError)
		backupDir := cleanupFlags.String("dir", os.Getenv("MTGA_BACKUP_DIR"), "Backup directory")
		olderThan := cleanupFlags.Int("older-than", 0, "Delete backups older than N days (0 = disabled)")
		keepLast := cleanupFlags.Int("keep-last", 0, "Keep only the last N backups (0 = disabled)")
		dryRun := cleanupFlags.Bool("dry-run", false, "Show what would be deleted without actually deleting")

		if err := cleanupFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if *backupDir == "" {
			*backupDir = backupMgr.GetBackupDir()
		}

		if *olderThan == 0 && *keepLast == 0 {
			fmt.Println("Error: either --older-than or --keep-last must be specified")
			fmt.Println("Usage: mtga-companion backup cleanup [flags]")
			fmt.Println("\nFlags:")
			cleanupFlags.PrintDefaults()
			os.Exit(1)
		}

		// List backups first to show what would be deleted
		backups, err := backupMgr.ListBackups(*backupDir)
		if err != nil {
			log.Fatalf("Error listing backups: %v", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return
		}

		if *dryRun {
			fmt.Printf("DRY RUN: Would clean up backups in %s\n", *backupDir)
			fmt.Printf("Found %d backup(s)\n", len(backups))
			if *olderThan > 0 {
				fmt.Printf("  - Deleting backups older than %d days\n", *olderThan)
			}
			if *keepLast > 0 {
				fmt.Printf("  - Keeping only the last %d backups\n", *keepLast)
			}
			return
		}

		fmt.Printf("Cleaning up backups in %s...\n", *backupDir)
		if err := backupMgr.CleanupBackups(*backupDir, *olderThan, *keepLast); err != nil {
			log.Fatalf("Error cleaning up backups: %v", err)
		}

		// Show how many remain
		remainingBackups, err := backupMgr.ListBackups(*backupDir)
		if err == nil {
			fmt.Printf("✓ Cleanup complete. %d backup(s) remaining.\n", len(remainingBackups))
		}

	case "info":
		// Define flags for info command
		infoFlags := flag.NewFlagSet("info", flag.ExitOnError)
		format := infoFlags.String("format", "table", "Output format: 'table' or 'json'")

		if err := infoFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		if infoFlags.NArg() < 1 {
			fmt.Println("Error: info command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup info <backup-file> [flags]")
			fmt.Println("\nFlags:")
			infoFlags.PrintDefaults()
			os.Exit(1)
		}
		backupPath := infoFlags.Arg(0)

		// Check if backup exists
		info, err := os.Stat(backupPath)
		if os.IsNotExist(err) {
			log.Fatalf("Backup file does not exist: %s", backupPath)
		}
		if err != nil {
			log.Fatalf("Error accessing backup file: %v", err)
		}

		// Try to load metadata
		metadata, err := backupMgr.LoadBackupMetadata(backupPath)

		// Format output
		switch *format {
		case "json":
			type BackupDetails struct {
				Path     string                  `json:"path"`
				Size     int64                   `json:"size"`
				ModTime  time.Time               `json:"modified"`
				Metadata *storage.BackupMetadata `json:"metadata,omitempty"`
			}

			details := BackupDetails{
				Path:     backupPath,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
				Metadata: metadata,
			}

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(details); err != nil {
				log.Fatalf("Error encoding JSON: %v", err)
			}
		case "table":
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("\nBackup Information:\n")
			fmt.Printf("  Path:     %s\n", backupPath)
			fmt.Printf("  Size:     %.2f MB\n", sizeMB)
			fmt.Printf("  Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))

			if metadata != nil {
				fmt.Printf("\n  Type:     %s\n", metadata.BackupType)
				fmt.Printf("  Created:  %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05"))
				if metadata.BaseBackup != "" {
					fmt.Printf("  Base:     %s\n", metadata.BaseBackup)
				}
				if len(metadata.Tables) > 0 {
					fmt.Printf("\n  Tables:   %d\n", len(metadata.Tables))
				}
			} else if err != nil {
				fmt.Printf("\n  Metadata: Not available (%v)\n", err)
			}
		default:
			log.Fatalf("Invalid format: %s (must be 'table' or 'json')", *format)
		}

	default:
		fmt.Printf("Unknown backup command: %s\n\n", command)
		printBackupUsage()
		os.Exit(1)
	}
}

func printBackupUsage() {
	fmt.Println("MTGA Companion - Database Backup Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mtga-companion backup <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create     Create a new database backup")
	fmt.Println("  restore    Restore database from backup")
	fmt.Println("  list, ls   List all available backups")
	fmt.Println("  verify     Verify backup integrity")
	fmt.Println("  cleanup    Delete old backups based on retention policy")
	fmt.Println("  info       Show detailed backup information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create full backup")
	fmt.Println("  mtga-companion backup create")
	fmt.Println()
	fmt.Println("  # Create incremental backup with encryption")
	fmt.Println("  export BACKUP_PWD=mypassword")
	fmt.Println("  mtga-companion backup create --type incremental --encrypt --password-env BACKUP_PWD")
	fmt.Println()
	fmt.Println("  # Create compressed backup")
	fmt.Println("  mtga-companion backup create --compress")
	fmt.Println()
	fmt.Println("  # Restore from encrypted backup")
	fmt.Println("  mtga-companion backup restore backup.db --password-env BACKUP_PWD")
	fmt.Println()
	fmt.Println("  # List backups in JSON format")
	fmt.Println("  mtga-companion backup list --format json")
	fmt.Println()
	fmt.Println("  # Clean up old backups (keep last 10)")
	fmt.Println("  mtga-companion backup cleanup --keep-last 10")
	fmt.Println()
	fmt.Println("  # Clean up backups older than 30 days")
	fmt.Println("  mtga-companion backup cleanup --older-than 30")
	fmt.Println()
	fmt.Println("  # Show backup metadata")
	fmt.Println("  mtga-companion backup info backup.db")
	fmt.Println()
	fmt.Println("For command-specific help:")
	fmt.Println("  mtga-companion backup <command> --help")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  MTGA_DB_PATH     Path to database file (default: ~/.mtga-companion/data.db)")
	fmt.Println("  MTGA_BACKUP_DIR  Backup directory (default: ~/.mtga-companion/backups)")
	fmt.Println()
}

// runBackupCommandInteractive handles backup commands from the interactive console.
func runBackupCommandInteractive(service *storage.Service, ctx context.Context) {
	// Get database path from service
	// We need to get it from the environment or use default
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Create backup manager
	backupMgr := storage.NewBackupManager(dbPath)

	fmt.Println("\nBackup Management")
	fmt.Println("-----------------")
	fmt.Println("1. Create backup")
	fmt.Println("2. List backups")
	fmt.Println("3. Verify backup")
	fmt.Print("\nSelect option (1-3) or 'cancel' to go back: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	choice := strings.TrimSpace(strings.ToLower(input))

	switch choice {
	case "1", "create", "backup":
		backupDir := os.Getenv("MTGA_BACKUP_DIR")
		config := storage.DefaultBackupConfig()
		config.BackupDir = backupDir
		config.VerifyBackup = true

		fmt.Println("\nCreating database backup...")
		backupPath, err := backupMgr.Backup(config)
		if err != nil {
			fmt.Printf("Error creating backup: %v\n", err)
			return
		}

		fmt.Printf("Backup created successfully: %s\n", backupPath)

		// Calculate and display backup size
		info, err := os.Stat(backupPath)
		if err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("Backup size: %.2f MB\n", sizeMB)
		}

	case "2", "list", "ls":
		backupDir := os.Getenv("MTGA_BACKUP_DIR")
		if backupDir == "" {
			backupDir = backupMgr.GetBackupDir()
		}

		fmt.Println("\nListing backups...")
		backups, err := backupMgr.ListBackups(backupDir)
		if err != nil {
			fmt.Printf("Error listing backups: %v\n", err)
			return
		}

		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return
		}

		fmt.Printf("\nFound %d backup(s):\n\n", len(backups))
		for i, backup := range backups {
			sizeMB := float64(backup.Size) / (1024 * 1024)
			fmt.Printf("%d. %s\n", i+1, backup.Name)
			fmt.Printf("   Path:     %s\n", backup.Path)
			fmt.Printf("   Size:     %.2f MB\n", sizeMB)
			fmt.Printf("   Modified: %s\n", backup.ModTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("   Checksum: %s\n", backup.Checksum)
			fmt.Println()
		}

	case "3", "verify":
		fmt.Print("Enter backup file path: ")
		backupPath, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}
		backupPath = strings.TrimSpace(backupPath)

		fmt.Printf("\nVerifying backup: %s\n", backupPath)
		if err := backupMgr.VerifyBackup(backupPath); err != nil {
			fmt.Printf("Backup verification failed: %v\n", err)
			return
		}

		fmt.Println("Backup verification successful!")

	case "cancel", "c", "back":
		return

	default:
		fmt.Printf("Unknown option: %s\n", choice)
	}
}

// handleExportCommand handles export commands with various subcommands.
func handleExportCommand(service *storage.Service, ctx context.Context, args []string) {
	if len(args) < 1 {
		printExportHelp()
		return
	}

	exportType := strings.ToLower(args[0])

	// Handle deck exports separately
	if exportType == "deck" || exportType == "decks" {
		handleDeckExport(service, ctx, args[1:])
		return
	}

	// Parse common options for statistics exports
	format := export.FormatCSV
	outputPath := ""
	prettyJSON := true
	overwrite := true
	var startDate, endDate *time.Time
	var formatFilter *string
	var formats []string     // Multiple format filter
	var eventName *string    // Event name filter
	var eventNames []string  // Multiple event names
	var opponentName *string // Opponent name filter
	var opponentID *string   // Opponent ID filter
	var result *string       // Result filter (win/loss)
	var rankClass *string    // Rank class filter
	var resultReason *string // Result reason filter
	periodType := "daily"    // Default for trend exports
	recentDays := 30         // Default for result comparison exports
	recentMatches := 30      // Default for prediction analysis window
	projectionMatches := 10  // Default for prediction projection

	// Parse flags from args
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-json", "--json":
			format = export.FormatJSON
		case "-csv", "--csv":
			format = export.FormatCSV
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		case "-start", "--start-date":
			if i+1 < len(args) {
				if t, err := time.Parse("2006-01-02", args[i+1]); err == nil {
					startDate = &t
					i++
				}
			}
		case "-end", "--end-date":
			if i+1 < len(args) {
				if t, err := time.Parse("2006-01-02", args[i+1]); err == nil {
					endDate = &t
					i++
				}
			}
		case "-format", "--format":
			if i+1 < len(args) {
				formatFilter = &args[i+1]
				i++
			}
		case "-formats", "--formats":
			// Parse comma-separated list of formats
			if i+1 < len(args) {
				formats = strings.Split(args[i+1], ",")
				i++
			}
		case "-event", "--event":
			if i+1 < len(args) {
				eventName = &args[i+1]
				i++
			}
		case "-events", "--events":
			// Parse comma-separated list of events
			if i+1 < len(args) {
				eventNames = strings.Split(args[i+1], ",")
				i++
			}
		case "-opponent", "--opponent":
			if i+1 < len(args) {
				opponentName = &args[i+1]
				i++
			}
		case "-opponent-id", "--opponent-id":
			if i+1 < len(args) {
				opponentID = &args[i+1]
				i++
			}
		case "-result", "--result":
			if i+1 < len(args) {
				result = &args[i+1]
				i++
			}
		case "-rank", "--rank":
			if i+1 < len(args) {
				rankClass = &args[i+1]
				i++
			}
		case "-reason", "--reason":
			if i+1 < len(args) {
				resultReason = &args[i+1]
				i++
			}
		case "-period", "--period":
			if i+1 < len(args) {
				periodType = args[i+1]
				i++
			}
		case "-recent", "--recent-days":
			if i+1 < len(args) {
				if days, err := strconv.Atoi(args[i+1]); err == nil && days > 0 {
					recentDays = days
					i++
				}
			}
		case "-window", "--recent-matches":
			if i+1 < len(args) {
				if matches, err := strconv.Atoi(args[i+1]); err == nil && matches > 0 {
					recentMatches = matches
					i++
				}
			}
		case "-project", "--projection":
			if i+1 < len(args) {
				if matches, err := strconv.Atoi(args[i+1]); err == nil && matches > 0 {
					projectionMatches = matches
					i++
				}
			}
		}
	}

	// Handle trend exports separately
	if exportType == "trend" || exportType == "trends" {
		// Trend exports require date range
		if startDate == nil || endDate == nil {
			fmt.Println("Trend export requires -start and -end dates")
			fmt.Println("Example: export trend -start 2024-01-01 -end 2024-12-31 -period daily")
			return
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("trends", format))
		}

		if err := export.ExportTrendAnalysis(ctx, service, *startDate, *endDate, periodType, formatFilter, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Trend export successful: %s\n", opts.FilePath)
		return
	}

	// Handle result comparison exports separately
	if exportType == "result-comparison" || exportType == "results-comparison" {
		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("result_comparison", format))
		}

		if err := export.ExportResultComparison(ctx, service, recentDays, formatFilter, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Result comparison export successful: %s\n", opts.FilePath)
		return
	}

	// Handle result trend exports separately
	if exportType == "result-trends" || exportType == "results-trends" {
		// Result trend exports require date range
		if startDate == nil || endDate == nil {
			fmt.Println("Result trend export requires -start and -end dates")
			fmt.Println("Example: export result-trends -start 2024-01-01 -end 2024-12-31 -period daily")
			return
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("result_trends", format))
		}

		if err := export.ExportResultTrends(ctx, service, *startDate, *endDate, periodType, formatFilter, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Result trend export successful: %s\n", opts.FilePath)
		return
	}

	// Handle period comparison exports separately (requires 4 dates: period1 start/end, period2 start/end)
	if exportType == "compare-periods" || exportType == "period-comparison" {
		fmt.Println("Period comparison requires both periods' date ranges")
		fmt.Println("Use: export compare-periods -start 2024-01-01 -end 2024-01-31 -start2 2024-02-01 -end2 2024-02-29")
		fmt.Println("Note: -start2 and -end2 flags are not yet implemented. Use result-comparison for recent vs all-time.")
		return
	}

	// Handle prediction exports separately
	if exportType == "predict" || exportType == "prediction" || exportType == "predict-winrate" {
		filter := models.StatsFilter{
			StartDate: startDate,
			EndDate:   endDate,
			Format:    formatFilter,
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("prediction", format))
		}

		if err := export.ExportWinRatePrediction(ctx, service, filter, recentMatches, projectionMatches, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Prediction export successful: %s\n", opts.FilePath)
		return
	}

	// Handle format-specific prediction exports
	if exportType == "predict-formats" || exportType == "predictions-by-format" {
		filter := models.StatsFilter{
			StartDate: startDate,
			EndDate:   endDate,
			Format:    formatFilter,
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("predictions_by_format", format))
		}

		if err := export.ExportPredictionsByFormat(ctx, service, filter, recentMatches, projectionMatches, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Format predictions export successful: %s\n", opts.FilePath)
		return
	}

	// Handle prediction summary exports
	if exportType == "predict-summary" || exportType == "prediction-summary" {
		filter := models.StatsFilter{
			StartDate: startDate,
			EndDate:   endDate,
			Format:    formatFilter,
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("prediction_summary", format))
		}

		if err := export.ExportPredictionSummary(ctx, service, filter, recentMatches, projectionMatches, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Prediction summary export successful: %s\n", opts.FilePath)
		return
	}

	// Handle rank timeline exports separately (requires format parameter)
	if exportType == "rank-timeline" || exportType == "timeline" {
		// Rank format is required (constructed or limited)
		if formatFilter == nil || (*formatFilter != "constructed" && *formatFilter != "limited") {
			fmt.Println("Error: Rank timeline requires a format (-format constructed or -format limited)")
			return
		}

		filter := models.StatsFilter{
			StartDate: startDate,
			EndDate:   endDate,
		}

		// Parse period type
		var period storage.TimelinePeriod
		switch periodType {
		case "all":
			period = storage.PeriodAll
		case "daily":
			period = storage.PeriodDaily
		case "weekly":
			period = storage.PeriodWeekly
		case "monthly":
			period = storage.PeriodMonthly
		default:
			fmt.Printf("Invalid period type: %s (use all, daily, weekly, or monthly)\n", periodType)
			return
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("rank_timeline", format))
		}

		if err := export.ExportRankTimeline(ctx, service, *formatFilter, filter, period, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Rank timeline export successful: %s\n", opts.FilePath)
		return
	}

	// Handle rank timeline summary exports
	if exportType == "rank-timeline-summary" || exportType == "timeline-summary" {
		// Rank format is required (constructed or limited)
		if formatFilter == nil || (*formatFilter != "constructed" && *formatFilter != "limited") {
			fmt.Println("Error: Rank timeline summary requires a format (-format constructed or -format limited)")
			return
		}

		filter := models.StatsFilter{
			StartDate: startDate,
			EndDate:   endDate,
		}

		// Parse period type
		var period storage.TimelinePeriod
		switch periodType {
		case "all":
			period = storage.PeriodAll
		case "daily":
			period = storage.PeriodDaily
		case "weekly":
			period = storage.PeriodWeekly
		case "monthly":
			period = storage.PeriodMonthly
		default:
			fmt.Printf("Invalid period type: %s (use all, daily, weekly, or monthly)\n", periodType)
			return
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("rank_timeline_summary", format))
		}

		if err := export.ExportRankTimelineSummary(ctx, service, *formatFilter, filter, period, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Rank timeline summary export successful: %s\n", opts.FilePath)
		return
	}

	// Handle match history exports
	if exportType == "matches" || exportType == "match-history" {
		filter := models.StatsFilter{
			StartDate: startDate,
			EndDate:   endDate,
			Format:    formatFilter,
		}

		opts := export.Options{
			Format:     format,
			FilePath:   outputPath,
			Overwrite:  overwrite,
			PrettyJSON: prettyJSON,
		}

		if outputPath == "" {
			opts.FilePath = filepath.Join("exports", export.GenerateFilename("matches", format))
		}

		if err := export.ExportMatchHistory(ctx, service, filter, opts); err != nil {
			fmt.Printf("Export failed: %v\n", err)
			return
		}

		fmt.Printf("✓ Match history export successful: %s\n", opts.FilePath)
		return
	}

	// Handle card data exports (requires card integration services)
	if exportType == "cards" || exportType == "card-set" || exportType == "card-meta" {
		handleCardExport(service, ctx, exportType, args[1:])
		return
	}

	// Create filter for statistics exports
	filter := models.StatsFilter{
		StartDate:    startDate,
		EndDate:      endDate,
		Format:       formatFilter,
		Formats:      formats,
		EventName:    eventName,
		EventNames:   eventNames,
		OpponentName: opponentName,
		OpponentID:   opponentID,
		Result:       result,
		RankClass:    rankClass,
		ResultReason: resultReason,
	}

	// Execute export based on type
	opts := export.ExportOptions{
		Type:       exportType,
		Format:     format,
		OutputPath: outputPath,
		Filter:     filter,
		Overwrite:  overwrite,
		PrettyJSON: prettyJSON,
	}

	if err := export.ExportStatistics(ctx, service, opts); err != nil {
		fmt.Printf("Export failed: %v\n", err)
		return
	}

	// Determine actual output path if one was generated
	actualPath := outputPath
	if actualPath == "" {
		actualPath = filepath.Join("exports", export.GenerateFilename(exportType, format))
	}

	fmt.Printf("✓ Export successful: %s\n", actualPath)
}

// handleTrendsCommand handles trend visualization commands.
func handleTrendsCommand(service *storage.Service, ctx context.Context, args []string) {
	if len(args) < 1 {
		printTrendsHelp()
		return
	}

	trendType := args[0]

	// Default parameters
	var startDate, endDate *time.Time
	var formatFilter *string
	periodType := "weekly" // Default period
	chartType := "line"    // Default chart type

	// Parse flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-start", "--start-date":
			if i+1 < len(args) {
				if t, err := time.Parse("2006-01-02", args[i+1]); err == nil {
					startDate = &t
					i++
				}
			}
		case "-end", "--end-date":
			if i+1 < len(args) {
				if t, err := time.Parse("2006-01-02", args[i+1]); err == nil {
					endDate = &t
					i++
				}
			}
		case "-format", "--format":
			if i+1 < len(args) {
				formatFilter = &args[i+1]
				i++
			}
		case "-period", "--period":
			if i+1 < len(args) {
				periodType = args[i+1]
				i++
			}
		case "-type", "--chart-type":
			if i+1 < len(args) {
				chartType = args[i+1]
				i++
			}
		case "-line":
			chartType = "line"
		case "-bar":
			chartType = "bar"
		}
	}

	// Set default date range if not provided (last 30 days)
	if startDate == nil || endDate == nil {
		now := time.Now()
		if endDate == nil {
			endDate = &now
		}
		if startDate == nil {
			start := now.AddDate(0, 0, -30)
			startDate = &start
		}
	}

	switch trendType {
	case "winrate", "wr":
		displayWinRateTrendChart(service, ctx, *startDate, *endDate, periodType, formatFilter, chartType)
	case "results", "breakdown", "rb":
		displayResultBreakdownChart(service, ctx, *startDate, *endDate, formatFilter)
	case "rank", "progression", "rp":
		displayRankProgressionChart(service, ctx, *startDate, *endDate, periodType, formatFilter)
	case "help", "h":
		printTrendsHelp()
	default:
		// Default to win rate trend
		displayWinRateTrendChart(service, ctx, *startDate, *endDate, periodType, formatFilter, chartType)
	}
}

// displayWinRateTrendChart displays a win rate trend chart.
func displayWinRateTrendChart(service *storage.Service, ctx context.Context, startDate, endDate time.Time, periodType string, formatFilter *string, chartType string) {
	// Get trend analysis
	analysis, err := service.GetTrendAnalysis(ctx, startDate, endDate, periodType, formatFilter)
	if err != nil {
		fmt.Printf("Error getting trend analysis: %v\n", err)
		return
	}

	if len(analysis.Periods) == 0 {
		fmt.Println("No data available for the specified period")
		return
	}

	// Prepare data points
	dataPoints := make([]charts.DataPoint, len(analysis.Periods))
	for i, period := range analysis.Periods {
		dataPoints[i] = charts.DataPoint{
			Label: period.Period.Label,
			Value: period.WinRate,
		}
	}

	// Configure chart
	config := charts.DefaultChartConfig()
	config.Title = "Win Rate Trend"
	if formatFilter != nil {
		config.Title = fmt.Sprintf("Win Rate Trend (%s)", *formatFilter)
	}
	config.YAxisLabel = "Win Rate (%)"
	config.Width = "900px"
	config.Height = "500px"

	// Create output file
	outputPath := filepath.Join("charts", fmt.Sprintf("winrate_trend_%s.html", time.Now().Format("20060102_150405")))
	if err := os.MkdirAll("charts", 0o755); err != nil {
		fmt.Printf("Error creating charts directory: %v\n", err)
		return
	}

	// Render chart
	if chartType == "bar" {
		err = charts.RenderBarChart(dataPoints, config, outputPath)
	} else {
		err = charts.RenderLineChart(dataPoints, config, outputPath)
	}

	if err != nil {
		fmt.Printf("Error creating chart: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Chart created: %s\n", outputPath)

	// Open in browser
	if err := charts.OpenInBrowser(outputPath); err != nil {
		fmt.Printf("Note: Could not open browser automatically. Please open %s manually.\n", outputPath)
	} else {
		fmt.Println("Opening chart in browser...")
	}

	// Display summary
	fmt.Printf("\nPeriod: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	fmt.Printf("Trend: %s", analysis.Trend)
	if analysis.TrendValue != 0 {
		fmt.Printf(" (%.1f%%)\n", analysis.TrendValue)
	} else {
		fmt.Println()
	}
	if analysis.Overall != nil {
		fmt.Printf("Overall Win Rate: %.1f%% (%d matches)\n", analysis.Overall.WinRate, analysis.Overall.TotalMatches)
	}
	fmt.Println()
}

// displayResultBreakdownChart displays a result breakdown pie chart.
func displayResultBreakdownChart(service *storage.Service, ctx context.Context, startDate, endDate time.Time, formatFilter *string) {
	// Get matches
	filter := storage.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
		Format:    formatFilter,
	}

	matches, err := service.GetMatches(ctx, filter)
	if err != nil {
		fmt.Printf("Error getting matches: %v\n", err)
		return
	}

	if len(matches) == 0 {
		fmt.Println("No matches found for the specified period")
		return
	}

	// Calculate win and loss breakdowns
	winBreakdown := calculateMatchBreakdown(matches, true)
	lossBreakdown := calculateMatchBreakdown(matches, false)

	// Create win breakdown chart
	if winBreakdown.Total > 0 {
		createBreakdownPieChart(winBreakdown, "Wins", startDate, endDate, formatFilter)
	}

	// Create loss breakdown chart
	if lossBreakdown.Total > 0 {
		createBreakdownPieChart(lossBreakdown, "Losses", startDate, endDate, formatFilter)
	}

	// Display summary
	fmt.Printf("\nPeriod: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if formatFilter != nil {
		fmt.Printf("Format: %s\n", *formatFilter)
	}
	fmt.Printf("Total Matches: %d (Wins: %d, Losses: %d)\n\n", len(matches), winBreakdown.Total, lossBreakdown.Total)
}

// matchBreakdown represents a breakdown of match results by reason.
type matchBreakdown struct {
	Normal             int
	Concede            int
	Timeout            int
	Draw               int
	Disconnect         int
	OpponentConcede    int
	OpponentTimeout    int
	OpponentDisconnect int
	Other              int
	Total              int
}

// calculateMatchBreakdown calculates a breakdown of match results by reason.
func calculateMatchBreakdown(matches []*storage.Match, isWin bool) matchBreakdown {
	breakdown := matchBreakdown{}

	for _, match := range matches {
		// Filter by win/loss
		if isWin && match.Result != "win" {
			continue
		}
		if !isWin && match.Result != "loss" {
			continue
		}

		breakdown.Total++

		if match.ResultReason == nil {
			breakdown.Normal++
			continue
		}

		reason := *match.ResultReason
		switch reason {
		case "ResultReason_Game":
			breakdown.Normal++
		case "ResultReason_Concede":
			breakdown.Concede++
		case "ResultReason_Timeout":
			breakdown.Timeout++
		case "ResultReason_Draw":
			breakdown.Draw++
		case "ResultReason_Disconnect":
			breakdown.Disconnect++
		case "ResultReason_OpponentConcede":
			breakdown.OpponentConcede++
		case "ResultReason_OpponentTimeout":
			breakdown.OpponentTimeout++
		case "ResultReason_OpponentDisconnect":
			breakdown.OpponentDisconnect++
		default:
			breakdown.Other++
		}
	}

	return breakdown
}

// createBreakdownPieChart creates a pie chart for the given breakdown.
func createBreakdownPieChart(breakdown matchBreakdown, resultType string, startDate, endDate time.Time, formatFilter *string) {
	// Prepare data points (only include non-zero values)
	dataPoints := []charts.DataPoint{}

	if breakdown.Normal > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Normal", Value: float64(breakdown.Normal)})
	}
	if breakdown.Concede > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Concede", Value: float64(breakdown.Concede)})
	}
	if breakdown.Timeout > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Timeout", Value: float64(breakdown.Timeout)})
	}
	if breakdown.Draw > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Draw", Value: float64(breakdown.Draw)})
	}
	if breakdown.Disconnect > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Disconnect", Value: float64(breakdown.Disconnect)})
	}
	if breakdown.OpponentConcede > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Opponent Concede", Value: float64(breakdown.OpponentConcede)})
	}
	if breakdown.OpponentTimeout > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Opponent Timeout", Value: float64(breakdown.OpponentTimeout)})
	}
	if breakdown.OpponentDisconnect > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Opponent Disconnect", Value: float64(breakdown.OpponentDisconnect)})
	}
	if breakdown.Other > 0 {
		dataPoints = append(dataPoints, charts.DataPoint{Label: "Other", Value: float64(breakdown.Other)})
	}

	if len(dataPoints) == 0 {
		return
	}

	// Configure chart
	config := charts.DefaultChartConfig()
	config.Title = fmt.Sprintf("%s Breakdown", resultType)
	if formatFilter != nil {
		config.Title = fmt.Sprintf("%s Breakdown (%s)", resultType, *formatFilter)
	}
	config.Width = "900px"
	config.Height = "600px"

	// Create output file
	filename := fmt.Sprintf("%s_breakdown_%s.html", strings.ToLower(resultType), time.Now().Format("20060102_150405"))
	outputPath := filepath.Join("charts", filename)
	if err := os.MkdirAll("charts", 0o755); err != nil {
		fmt.Printf("Error creating charts directory: %v\n", err)
		return
	}

	// Render pie chart
	if err := charts.RenderPieChart(dataPoints, config, outputPath); err != nil {
		fmt.Printf("Error creating %s chart: %v\n", resultType, err)
		return
	}

	fmt.Printf("✓ %s chart created: %s\n", resultType, outputPath)

	// Open in browser
	if err := charts.OpenInBrowser(outputPath); err != nil {
		fmt.Printf("Note: Could not open browser automatically for %s chart.\n", resultType)
	} else {
		fmt.Printf("Opening %s chart in browser...\n", resultType)
	}
}

// displayRankProgressionChart displays a rank progression line chart.
func displayRankProgressionChart(service *storage.Service, ctx context.Context, startDate, endDate time.Time, periodType string, formatFilter *string) {
	// Default to constructed format if not specified
	format := "constructed"
	if formatFilter != nil {
		format = *formatFilter
	}

	// Convert period type to TimelinePeriod
	var period storage.TimelinePeriod
	switch periodType {
	case "daily":
		period = storage.PeriodDaily
	case "weekly":
		period = storage.PeriodWeekly
	case "monthly":
		period = storage.PeriodMonthly
	default:
		period = storage.PeriodWeekly
	}

	// Get rank progression timeline
	timeline, err := service.GetRankProgressionTimeline(ctx, format, &startDate, &endDate, period)
	if err != nil {
		fmt.Printf("Error getting rank progression: %v\n", err)
		return
	}

	if len(timeline.Entries) == 0 {
		fmt.Println("No rank progression data found for the specified period")
		return
	}

	// Convert timeline entries to chart data points
	dataPoints := make([]charts.DataPoint, len(timeline.Entries))
	for i, entry := range timeline.Entries {
		dataPoints[i] = charts.DataPoint{
			Label: entry.Date,
			Value: rankToNumericValue(entry.RankClass, entry.RankLevel),
		}
	}

	// Configure chart
	config := charts.DefaultChartConfig()
	// Capitalize format (constructed -> Constructed, limited -> Limited)
	formatTitle := format
	if len(format) > 0 {
		formatTitle = strings.ToUpper(format[:1]) + format[1:]
	}
	config.Title = fmt.Sprintf("Rank Progression (%s)", formatTitle)
	config.Subtitle = fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	config.YAxisLabel = "Rank"
	config.Width = "1000px"
	config.Height = "600px"

	// Create output file
	outputPath := filepath.Join("charts", fmt.Sprintf("rank_progression_%s_%s.html", format, time.Now().Format("20060102_150405")))
	if err := os.MkdirAll("charts", 0o755); err != nil {
		fmt.Printf("Error creating charts directory: %v\n", err)
		return
	}

	// Render line chart
	if err := charts.RenderLineChart(dataPoints, config, outputPath); err != nil {
		fmt.Printf("Error creating rank progression chart: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Rank progression chart created: %s\n", outputPath)

	// Open in browser
	if err := charts.OpenInBrowser(outputPath); err != nil {
		fmt.Printf("Note: Could not open browser automatically.\n")
	} else {
		fmt.Println("Opening rank progression chart in browser...")
	}

	// Display summary
	fmt.Printf("\nPeriod: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	fmt.Printf("Format: %s\n", format)
	fmt.Printf("Start Rank: %s\n", timeline.StartRank)
	fmt.Printf("End Rank: %s\n", timeline.EndRank)
	fmt.Printf("Highest Rank: %s\n", timeline.HighestRank)
	fmt.Printf("Lowest Rank: %s\n", timeline.LowestRank)
	fmt.Printf("Total Changes: %d\n", timeline.TotalChanges)
	fmt.Printf("Milestones: %d\n\n", timeline.Milestones)
}

// rankToNumericValue converts rank class and level to a numeric value for charting.
// Lower ranks = lower numbers, higher ranks = higher numbers.
func rankToNumericValue(rankClass *string, rankLevel *int) float64 {
	if rankClass == nil {
		return 0
	}

	// Map rank classes to base values
	rankClassValues := map[string]float64{
		"Bronze":   0,
		"Silver":   4,
		"Gold":     8,
		"Platinum": 12,
		"Diamond":  16,
		"Mythic":   20,
	}

	baseValue, ok := rankClassValues[*rankClass]
	if !ok {
		return 0
	}

	// Add level offset (higher level = higher value)
	// Rank levels go from 4 (lowest) to 1 (highest)
	if rankLevel != nil && *rankLevel >= 1 && *rankLevel <= 4 {
		baseValue += float64(5 - *rankLevel) // Convert so level 4=1, level 1=4
	} else if *rankClass == "Mythic" {
		// Mythic has no levels, just use base value
		baseValue += 4 // Treat Mythic as highest
	}

	return baseValue
}

// storageMetadataAdapter adapts storage.Service to implement unified.CardMetadataProvider.
type storageMetadataAdapter struct {
	storage *storage.Service
}

func (a *storageMetadataAdapter) GetCard(ctx context.Context, arenaID int) (*storage.Card, error) {
	return a.storage.GetCardByArenaID(ctx, arenaID)
}

func (a *storageMetadataAdapter) GetCards(ctx context.Context, arenaIDs []int) ([]*storage.Card, error) {
	cards := make([]*storage.Card, 0, len(arenaIDs))
	for _, id := range arenaIDs {
		card, err := a.storage.GetCardByArenaID(ctx, id)
		if err == nil && card != nil {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func (a *storageMetadataAdapter) GetSetCards(ctx context.Context, setCode string) ([]*storage.Card, error) {
	return a.storage.GetCardsBySet(ctx, setCode)
}

// storageDraftStatsAdapter adapts storage.Service to implement unified.DraftStatsProvider.
type storageDraftStatsAdapter struct {
	storage *storage.Service
}

func (a *storageDraftStatsAdapter) GetCardRating(ctx context.Context, arenaID int, expansion, format, colors string) (*storage.DraftCardRating, error) {
	return a.storage.GetCardRating(ctx, arenaID, expansion, format, colors)
}

func (a *storageDraftStatsAdapter) GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]*storage.DraftCardRating, error) {
	return a.storage.GetCardRatingsForSet(ctx, expansion, format, colors)
}

// handleCardExport handles card data export commands.
func handleCardExport(service *storage.Service, ctx context.Context, exportType string, args []string) {
	// Parse export options from CLI args
	var (
		setCode       string
		draftFormat   string
		outputFormat  string
		outputPath    string
		includeStats  bool
		topN          int
		sortBy        string
		filterRarity  string
		filterColors  string
		showAge       bool
		minSample     int
		onlyWithStats bool
		prettyJSON    bool
	)

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--set", "-s":
			if i+1 < len(args) {
				setCode = args[i+1]
				i++
			}
		case "--format", "-f":
			if i+1 < len(args) {
				draftFormat = args[i+1]
				i++
			}
		case "--csv":
			outputFormat = "csv"
		case "--json":
			outputFormat = "json"
		case "--markdown", "--md":
			outputFormat = "markdown"
		case "--arena":
			outputFormat = "arena"
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		case "--include-stats":
			includeStats = true
		case "--top":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					topN = n
				}
				i++
			}
		case "--sort":
			if i+1 < len(args) {
				sortBy = args[i+1]
				i++
			}
		case "--rarity":
			if i+1 < len(args) {
				filterRarity = args[i+1]
				i++
			}
		case "--colors":
			if i+1 < len(args) {
				filterColors = args[i+1]
				i++
			}
		case "--show-age":
			showAge = true
		case "--min-sample":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					minSample = n
				}
				i++
			}
		case "--only-with-stats":
			onlyWithStats = true
		case "--pretty":
			prettyJSON = true
		}
	}

	// Validate required parameters
	if setCode == "" {
		fmt.Println("Error: --set is required")
		fmt.Println("\nUsage:")
		fmt.Println("  export cards --set <set_code> [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  --set <code>           Set code (required, e.g., BLB, MKM)")
		fmt.Println("  --format <format>      Draft format (PremierDraft, QuickDraft, etc.)")
		fmt.Println("  --csv                  Export as CSV (default)")
		fmt.Println("  --json                 Export as JSON")
		fmt.Println("  --markdown, --md       Export as Markdown")
		fmt.Println("  --arena                Export as MTG Arena deck format")
		fmt.Println("  --include-stats        Include 17Lands draft statistics")
		fmt.Println("  --top <N>              Export only top N cards")
		fmt.Println("  --sort <field>         Sort by: gihwr, alsa, ata, cmc, name")
		fmt.Println("  --rarity <list>        Filter by rarity (comma-separated)")
		fmt.Println("  --colors <list>        Filter by colors (comma-separated)")
		fmt.Println("  --min-sample <N>       Minimum sample size")
		fmt.Println("  --only-with-stats      Only export cards with draft stats")
		fmt.Println("  --show-age             Include data freshness indicators")
		fmt.Println("  --pretty               Pretty-print JSON output")
		fmt.Println("  -o, --output <path>    Output file path (default: stdout)")
		fmt.Println("\nExamples:")
		fmt.Println("  export cards --set BLB --csv")
		fmt.Println("  export cards --set BLB --json --include-stats --pretty")
		fmt.Println("  export cards --set BLB --top 20 --sort gihwr --arena")
		fmt.Println("  export card-meta --set BLB --top 20 --sort gihwr --markdown")
		return
	}

	// Default values
	if outputFormat == "" {
		outputFormat = "csv"
	}
	if draftFormat == "" {
		draftFormat = "PremierDraft"
	}
	if sortBy == "" && topN > 0 {
		sortBy = "gihwr" // Default sort for top-N queries
	}

	// Initialize card services
	fmt.Println("Initializing card services...")

	// Create storage adapters for unified service
	metadataAdapter := &storageMetadataAdapter{storage: service}
	statsAdapter := &storageDraftStatsAdapter{storage: service}

	// Create unified service
	unifiedService := unified.NewService(metadataAdapter, statsAdapter)

	// Create query interface
	cardQuery, err := query.NewCardQuery(query.QueryConfig{
		UnifiedService: unifiedService,
		Storage:        service,
	})
	if err != nil {
		fmt.Printf("Error creating card query: %v\n", err)
		return
	}
	defer func() { _ = cardQuery.Close() }()

	// Fetch cards based on export type
	fmt.Printf("Fetching cards for set %s...\n", setCode)

	queryOpts := query.QueryOptions{
		Format:       draftFormat,
		IncludeStats: includeStats || onlyWithStats, // Always include stats if filtering by them
		MaxStaleAge:  24 * time.Hour,                // Accept data up to 1 day old
		FallbackMode: query.AllowPartial,            // Allow partial data
	}

	cards, err := cardQuery.GetSet(ctx, setCode, queryOpts)
	if err != nil {
		fmt.Printf("Error fetching cards: %v\n", err)
		return
	}

	if len(cards) == 0 {
		fmt.Printf("No cards found for set %s\n", setCode)
		return
	}

	fmt.Printf("Found %d cards\n", len(cards))

	// Build export options
	exportOpts := export.CardExportOptions{
		Format:        export.Format(outputFormat),
		IncludeStats:  includeStats,
		TopN:          topN,
		SortBy:        sortBy,
		ShowDataAge:   showAge,
		MinSampleSize: minSample,
		OnlyWithStats: onlyWithStats,
		PrettyJSON:    prettyJSON,
	}

	// Parse filter options
	if filterRarity != "" {
		exportOpts.FilterRarity = strings.Split(filterRarity, ",")
	}
	if filterColors != "" {
		exportOpts.FilterColors = strings.Split(filterColors, ",")
	}

	// Export to output
	var output io.Writer
	if outputPath != "" {
		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer func() { _ = file.Close() }()
		output = file
		fmt.Printf("Exporting to %s...\n", outputPath)
	} else {
		output = os.Stdout
		fmt.Println() // Blank line before output
	}

	// Perform export
	if err := export.ExportCards(output, cards, exportOpts); err != nil {
		fmt.Printf("Error exporting cards: %v\n", err)
		return
	}

	if outputPath != "" {
		fmt.Printf("✓ Export complete: %s\n", outputPath)
	}
}

// printTrendsHelp prints help for the trends/chart command.
func printTrendsHelp() {
	fmt.Println("\nChart Commands:")
	fmt.Println("  chart [type] [options]            - Create interactive HTML charts")
	fmt.Println("  chart winrate [options]           - Create win rate trend chart")
	fmt.Println("  chart results [options]           - Create result breakdown pie charts")
	fmt.Println("  chart rank [options]              - Create rank progression chart")
	fmt.Println("\nNote: Charts are saved as HTML files in the 'charts' directory and opened in your browser")
	fmt.Println("\nOptions:")
	fmt.Println("  -start, --start-date <date>       - Start date (YYYY-MM-DD)")
	fmt.Println("  -end, --end-date <date>           - End date (YYYY-MM-DD)")
	fmt.Println("  -format, --format <format>        - Filter by format (constructed/limited)")
	fmt.Println("  -period, --period <type>          - Period type: daily, weekly, monthly (default: weekly)")
	fmt.Println("  -type, --chart-type <type>        - Chart type: line, bar (default: line)")
	fmt.Println("  -line                             - Use line chart")
	fmt.Println("  -bar                              - Use bar chart")
	fmt.Println("\nExamples:")
	fmt.Println("  chart                             - Show win rate trend (last 30 days)")
	fmt.Println("  chart winrate                     - Show win rate trend chart")
	fmt.Println("  chart results                     - Show result breakdown pie charts (wins/losses)")
	fmt.Println("  chart rank                        - Show rank progression chart (constructed)")
	fmt.Println("  chart rank -format limited        - Show rank progression for limited")
	fmt.Println("  chart -start 2024-01-01 -end 2024-12-31 -period monthly")
	fmt.Println("  chart results -format constructed - Result breakdown for constructed format")
	fmt.Println("  chart -format limited -bar        - Limited win rate (bar chart)")
	fmt.Println()
}

// handleDeckExport handles deck export commands.
func handleDeckExport(service *storage.Service, ctx context.Context, args []string) {
	// Parse deck export options
	deckFormat := export.DeckFormatArena // default
	outputPath := ""
	var deckIDs []string
	var formatFilter string
	exportAll := false

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-arena":
			deckFormat = export.DeckFormatArena
		case "-text":
			deckFormat = export.DeckFormatText
		case "-json":
			deckFormat = export.DeckFormatJSON
		case "-csv":
			deckFormat = export.DeckFormatCSV
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		case "-id", "--deck-id":
			if i+1 < len(args) {
				deckIDs = append(deckIDs, args[i+1])
				i++
			}
		case "-format", "--format":
			if i+1 < len(args) {
				formatFilter = args[i+1]
				i++
			}
		case "-all", "--all":
			exportAll = true
		default:
			// Treat as deck ID if it doesn't start with -
			if !strings.HasPrefix(arg, "-") {
				deckIDs = append(deckIDs, arg)
			}
		}
	}

	// Create deck viewer
	deckViewer := viewer.NewDeckViewer(service.GetDB(), nil)

	var err error

	// Determine output path
	if outputPath == "" {
		if len(deckIDs) == 1 {
			// Get deck name for filename
			deck, err := deckViewer.GetDeck(ctx, deckIDs[0])
			if err == nil && deck != nil {
				outputPath = filepath.Join("exports", export.GenerateDeckFilename(deck.Deck.Name, deckFormat))
			}
		}
		if outputPath == "" {
			outputPath = filepath.Join("exports", fmt.Sprintf("decks_%s.%s", time.Now().Format("20060102_150405"), deckFormat))
		}
	}

	// Execute export based on options
	switch {
	case exportAll:
		err = export.ExportAllDecks(ctx, deckViewer, deckFormat, outputPath)
	case formatFilter != "":
		err = export.ExportDecksByFormat(ctx, deckViewer, formatFilter, deckFormat, outputPath)
	case len(deckIDs) > 0:
		err = export.ExportDecks(ctx, deckViewer, deckIDs, deckFormat, outputPath)
	default:
		fmt.Println("Error: No decks specified. Use -all, -format, or provide deck IDs.")
		printDeckExportHelp()
		return
	}

	if err != nil {
		fmt.Printf("Deck export failed: %v\n", err)
		return
	}

	fmt.Printf("✓ Deck export successful: %s\n", outputPath)
}

// printExportHelp prints help for the export command.
func printExportHelp() {
	fmt.Println("\nExport Commands:")
	fmt.Println("  export matches [options]          - Export match history")
	fmt.Println("  export stats [options]            - Export aggregated statistics")
	fmt.Println("  export daily [options]            - Export daily statistics")
	fmt.Println("  export performance [options]      - Export performance metrics")
	fmt.Println("  export trends [options]           - Export historical trend analysis")
	fmt.Println("  export results [options]          - Export result breakdown (win/loss reasons)")
	fmt.Println("  export result-comparison [opts]   - Compare recent vs. all-time results")
	fmt.Println("  export result-trends [options]    - Trend analysis for result reasons over time")
	fmt.Println("  export streaks [options]          - Export performance streaks")
	fmt.Println("  export hour-of-day [options]      - Export statistics by hour of day (0-23)")
	fmt.Println("  export day-of-week [options]      - Export statistics by day of week")
	fmt.Println("  export time-patterns [options]    - Export time-based performance summary")
	fmt.Println("  export predict [options]          - Predict future win rate based on trends")
	fmt.Println("  export predict-formats [options]  - Predict win rates for each format")
	fmt.Println("  export predict-summary [options]  - Prediction summary with insights")
	fmt.Println("  export rank-timeline [options]    - Export rank progression timeline")
	fmt.Println("  export timeline-summary [options] - Export rank progression summary")
	fmt.Println("  export cards [options]            - Export card data with statistics (requires card services)")
	fmt.Println("  export card-set [options]         - Export all cards in a set with stats")
	fmt.Println("  export card-meta [options]        - Export meta snapshot for a set")
	fmt.Println("  export deck/decks [options]       - Export deck lists")
	fmt.Println("\nOptions:")
	fmt.Println("  -json                             - Export as JSON (default: CSV)")
	fmt.Println("  -csv                              - Export as CSV")
	fmt.Println("  -o, --output <path>               - Specify output file path")
	fmt.Println("  -start, --start-date <date>       - Start date (YYYY-MM-DD)")
	fmt.Println("  -end, --end-date <date>           - End date (YYYY-MM-DD)")
	fmt.Println("  -format, --format <format>        - Filter by format (constructed/limited)")
	fmt.Println("  -formats, --formats <list>        - Filter by multiple formats (comma-separated)")
	fmt.Println("  -event, --event <name>            - Filter by event name")
	fmt.Println("  -events, --events <list>          - Filter by multiple events (comma-separated)")
	fmt.Println("  -opponent, --opponent <name>      - Filter by opponent name")
	fmt.Println("  -opponent-id, --opponent-id <id>  - Filter by opponent ID")
	fmt.Println("  -result, --result <win|loss>      - Filter by match result")
	fmt.Println("  -rank, --rank <class>             - Filter by rank class (e.g., Mythic, Diamond)")
	fmt.Println("  -reason, --reason <reason>        - Filter by result reason (e.g., concede, timeout)")
	fmt.Println("  -period, --period <type>          - Period type: daily, weekly, monthly (default: daily)")
	fmt.Println("  -recent, --recent-days <days>     - Recent period in days (default: 30)")
	fmt.Println("  -window, --recent-matches <num>   - Recent matches for prediction analysis (default: 30)")
	fmt.Println("  -project, --projection <num>      - Matches ahead to project (default: 10)")
	fmt.Println("\nExamples:")
	fmt.Println("  export matches -json              - Export all matches as JSON")
	fmt.Println("  export matches -formats Standard,Historic -json - Export matches from multiple formats")
	fmt.Println("  export matches -rank Mythic -result win -json - Export all Mythic wins")
	fmt.Println("  export matches -opponent \"PlayerName\" -csv  - Export matches vs specific opponent")
	fmt.Println("  export stats -csv -o stats.csv")
	fmt.Println("  export trends -start 2024-01-01 -end 2024-12-31 -period daily")
	fmt.Println("  export results -json              - Export result breakdown as JSON")
	fmt.Println("  export result-comparison -recent 7 -format constructed")
	fmt.Println("  export result-trends -start 2024-01-01 -end 2024-12-31 -period weekly")
	fmt.Println("  export streaks -json              - Export current streak data")
	fmt.Println("  export streaks -format constructed - Export streaks for constructed only")
	fmt.Println("  export hour-of-day -json          - Export win rates by hour (0-23)")
	fmt.Println("  export day-of-week -csv           - Export win rates by day of week")
	fmt.Println("  export time-patterns -json        - Export best/worst hours and days")
	fmt.Println("  export predict -json              - Predict future win rate with confidence")
	fmt.Println("  export predict -window 50 -project 20 - Analyze last 50, project 20 ahead")
	fmt.Println("  export predict-formats -json      - Predictions for each format")
	fmt.Println("  export predict-summary -csv       - Summary with strongest/weakest formats")
	fmt.Println("  export rank-timeline -format constructed -period daily")
	fmt.Println("  export rank-timeline -format limited -period all -json")
	fmt.Println("  export timeline-summary -format constructed -start 2024-01-01")
	fmt.Println("  export daily -start 2024-01-01 -end 2024-01-31")
	fmt.Println("  export matches -format constructed -json")
	fmt.Println("  export cards --set BLB --csv      - Export BLB cards (coming soon)")
	fmt.Println("  export cards --set BLB --json --include-stats - Export with 17Lands stats")
	fmt.Println("  export card-meta --set BLB --top 20 --markdown - Export top 20 cards")
	fmt.Println("  export decks -all -arena          - Export all decks in Arena format")
	fmt.Println("  export deck <deck_id> -json       - Export specific deck as JSON")
}

// printDeckExportHelp prints detailed help for deck exports.
func printDeckExportHelp() {
	fmt.Println("\nDeck Export Commands:")
	fmt.Println("  export deck/decks [options]   - Export deck lists")
	fmt.Println("\nDeck Format Options:")
	fmt.Println("  -arena                        - Arena format (default)")
	fmt.Println("  -text                         - Human-readable text format")
	fmt.Println("  -json                         - JSON format")
	fmt.Println("  -csv                          - CSV format (one row per card)")
	fmt.Println("\nDeck Selection Options:")
	fmt.Println("  -all                          - Export all decks")
	fmt.Println("  -format <format>              - Export decks of specific format (Standard, etc.)")
	fmt.Println("  -id, --deck-id <id>           - Export specific deck by ID")
	fmt.Println("  <deck_id>                     - Export specific deck (can be multiple)")
	fmt.Println("\nOther Options:")
	fmt.Println("  -o, --output <path>           - Specify output file path")
	fmt.Println("\nExamples:")
	fmt.Println("  export decks -all -arena      - Export all decks in Arena format")
	fmt.Println("  export deck <id> -json        - Export specific deck as JSON")
	fmt.Println("  export decks -format Standard -text")
	fmt.Println("  export deck <id1> <id2> -csv  - Export multiple decks as CSV")
}

// runDraftCommand handles draft-related commands.
func runDraftCommand() {
	// Get database path from environment or use default
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbPath)
	}

	// Open database
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Create storage service
	service := storage.NewService(db)
	defer func() {
		_ = service.Close()
	}()

	// Create card lookup service
	scryfallClient := scryfall.NewClient()
	cardService := cardlookup.NewService(service, scryfallClient, cardlookup.DefaultServiceOptions())

	ctx := context.Background()

	if len(os.Args) < 3 {
		printDraftUsage()
		os.Exit(1)
	}

	command := os.Args[2]

	switch command {
	case "picks", "show-picks", "list-picks":
		// Define flags for picks command
		picksFlags := flag.NewFlagSet("picks", flag.ExitOnError)
		draftEventID := picksFlags.String("event", "", "Draft event ID")
		format := picksFlags.String("format", "detailed", "Display format: 'detailed', 'compact', or 'summary'")
		listEvents := picksFlags.Bool("list-events", false, "List all draft events with picks")

		if err := picksFlags.Parse(os.Args[3:]); err != nil {
			log.Fatalf("Error parsing flags: %v", err)
		}

		// List events if requested
		if *listEvents {
			events, err := service.GetAllDraftEventsWithPicks(ctx)
			if err != nil {
				log.Fatalf("Error listing draft events: %v", err)
			}

			if len(events) == 0 {
				fmt.Println("No draft events with picks found.")
				return
			}

			fmt.Printf("\nDraft Events with Picks:\n")
			fmt.Printf("========================\n\n")
			for i, eventID := range events {
				count, err := service.GetDraftPicksCount(ctx, eventID)
				if err != nil {
					fmt.Printf("%d. %s (error getting count)\n", i+1, eventID)
					continue
				}
				fmt.Printf("%d. %s (%d picks)\n", i+1, eventID, count)
			}
			fmt.Println("\nUse: mtga-companion draft picks --event <event-id> to view picks")
			return
		}

		// Require event ID if not listing
		if *draftEventID == "" {
			fmt.Println("Error: --event flag is required")
			fmt.Println("Use --list-events to see available draft events")
			fmt.Println("\nUsage: mtga-companion draft picks --event <event-id> [--format detailed|compact|summary]")
			os.Exit(1)
		}

		// Get picks for the event
		picks, err := service.GetDraftPicks(ctx, *draftEventID)
		if err != nil {
			log.Fatalf("Error getting draft picks: %v", err)
		}

		if len(picks) == 0 {
			fmt.Printf("No picks found for draft event: %s\n", *draftEventID)
			return
		}

		// Create displayer
		displayer := display.NewDraftPicksDisplayer(cardService)

		// Display picks based on format
		switch *format {
		case "detailed", "full":
			if err := displayer.DisplayPicks(ctx, picks); err != nil {
				log.Fatalf("Error displaying picks: %v", err)
			}
		case "compact", "table":
			if err := displayer.DisplayPicksCompact(ctx, picks); err != nil {
				log.Fatalf("Error displaying picks: %v", err)
			}
		case "summary":
			if err := displayer.DisplayPicksSummary(ctx, picks); err != nil {
				log.Fatalf("Error displaying picks: %v", err)
			}
		default:
			log.Fatalf("Invalid format: %s (must be 'detailed', 'compact', or 'summary')", *format)
		}

	default:
		fmt.Printf("Unknown draft command: %s\n\n", command)
		printDraftUsage()
		os.Exit(1)
	}
}

func printDraftUsage() {
	fmt.Println("MTGA Companion - Draft Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mtga-companion draft <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  picks      Display draft picks for a draft event")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all draft events with picks")
	fmt.Println("  mtga-companion draft picks --list-events")
	fmt.Println()
	fmt.Println("  # Show detailed picks for a draft event")
	fmt.Println("  mtga-companion draft picks --event draft-abc123")
	fmt.Println()
	fmt.Println("  # Show compact table view")
	fmt.Println("  mtga-companion draft picks --event draft-abc123 --format compact")
	fmt.Println()
	fmt.Println("  # Show summary only")
	fmt.Println("  mtga-companion draft picks --event draft-abc123 --format summary")
	fmt.Println()
	fmt.Println("Display Formats:")
	fmt.Println("  detailed   - Full detailed view with pack contents (default)")
	fmt.Println("  compact    - Compact table showing pick summary")
	fmt.Println("  summary    - High-level summary only")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  MTGA_DB_PATH     Path to database file (default: ~/.mtga-companion/data.db)")
	fmt.Println()
}

func runCardsCommand() {
	if len(os.Args) < 3 {
		printCardsUsage()
		os.Exit(1)
	}

	command := os.Args[2]

	switch command {
	case "import":
		runCardsImport()
	case "check-updates":
		runCardsCheckUpdates()
	case "update":
		runCardsUpdate()
	case "auto-update":
		runCardsAutoUpdate()
	case "cache-clear":
		runCardsCacheClear()
	case "cache-stats":
		runCardsCacheStats()
	case "check-migrations":
		runCardsCheckMigrations()
	case "apply-migrations":
		runCardsApplyMigrations()
	case "migration-stats":
		runCardsMigrationStats()
	default:
		fmt.Printf("Unknown cards command: %s\n\n", command)
		printCardsUsage()
		os.Exit(1)
	}
}

func runCardsImport() {
	// Define flags for import command
	importFlags := flag.NewFlagSet("import", flag.ExitOnError)
	forceDownload := importFlags.Bool("force", false, "Force re-download of bulk data file")
	verbose := importFlags.Bool("verbose", false, "Enable verbose output")
	batchSize := importFlags.Int("batch-size", 500, "Number of cards to insert per batch")

	if err := importFlags.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
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

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		log.Fatalf("Error creating database directory: %v", err)
	}

	// Open database with auto-migrate to ensure schema is up to date
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create storage service
	service := storage.NewService(db)
	defer func() { _ = service.Close() }()

	// Create Scryfall client
	scryfallClient := scryfall.NewClient()

	// Create importer with options
	options := importer.DefaultBulkImportOptions()
	options.ForceDownload = *forceDownload
	options.Verbose = *verbose
	options.BatchSize = *batchSize

	bulkImporter := importer.NewBulkImporter(scryfallClient, service, options)

	// Run import
	ctx := context.Background()
	fmt.Println("Starting bulk card import from Scryfall...")
	fmt.Println()

	stats, err := bulkImporter.Import(ctx)
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	if !*verbose {
		// Print summary if not already printed in verbose mode
		fmt.Println("Import complete!")
		fmt.Printf("  Total cards processed: %d\n", stats.TotalCards)
		fmt.Printf("  Cards imported: %d\n", stats.ImportedCards)
		fmt.Printf("  Cards skipped (no Arena ID): %d\n", stats.SkippedCards)
		fmt.Printf("  Errors: %d\n", stats.ErrorCards)
		fmt.Printf("  Total time: %.2fs\n", stats.Duration.Seconds())
	}
}

func printCardsUsage() {
	fmt.Println("MTGA Companion - Card Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mtga-companion cards <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  import              Import card data from Scryfall bulk data")
	fmt.Println("  check-updates       Check for available card data updates")
	fmt.Println("  update              Apply card data updates (incremental)")
	fmt.Println("  auto-update         Automatically check and apply bulk data updates")
	fmt.Println("  cache-clear         Clear the image cache")
	fmt.Println("  cache-stats         Show image cache statistics")
	fmt.Println("  check-migrations    Check for new Scryfall card migrations")
	fmt.Println("  apply-migrations    Apply Scryfall card migrations")
	fmt.Println("  migration-stats     Show migration statistics")
	fmt.Println()
	fmt.Println("Import Flags:")
	fmt.Println("  --force        Force re-download of bulk data file")
	fmt.Println("  --verbose      Enable verbose output with progress")
	fmt.Println("  --batch-size   Number of cards to insert per batch (default: 500)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Initial import")
	fmt.Println("  mtga-companion cards import")
	fmt.Println()
	fmt.Println("  # Force re-import with verbose output")
	fmt.Println("  mtga-companion cards import --force --verbose")
	fmt.Println()
	fmt.Println("  # Import with custom batch size")
	fmt.Println("  mtga-companion cards import --batch-size 1000")
	fmt.Println()
	fmt.Println("Update Flags:")
	fmt.Println("  --force        Force update all sets regardless of staleness")
	fmt.Println("  --verbose      Enable verbose output")
	fmt.Println("  --set CODE     Update specific set only")
	fmt.Println()
	fmt.Println("Update Examples:")
	fmt.Println("  # Check for updates")
	fmt.Println("  mtga-companion cards check-updates")
	fmt.Println()
	fmt.Println("  # Apply updates")
	fmt.Println("  mtga-companion cards update")
	fmt.Println()
	fmt.Println("  # Force update with verbose output")
	fmt.Println("  mtga-companion cards update --force --verbose")
	fmt.Println()
	fmt.Println("  # Update specific set")
	fmt.Println("  mtga-companion cards update --set BLB")
	fmt.Println()
	fmt.Println("Auto-Update Flags:")
	fmt.Println("  --check-only   Check for updates without applying them")
	fmt.Println("  --force        Force update regardless of timestamps")
	fmt.Println("  --verbose      Enable verbose output")
	fmt.Println()
	fmt.Println("Auto-Update Examples:")
	fmt.Println("  # Check if bulk data updates are available")
	fmt.Println("  mtga-companion cards auto-update --check-only")
	fmt.Println()
	fmt.Println("  # Automatically apply bulk data updates if available")
	fmt.Println("  mtga-companion cards auto-update")
	fmt.Println()
	fmt.Println("  # Force bulk data update with verbose output")
	fmt.Println("  mtga-companion cards auto-update --force --verbose")
	fmt.Println()
	fmt.Println("Cache Management Examples:")
	fmt.Println("  # Show cache statistics")
	fmt.Println("  mtga-companion cards cache-stats")
	fmt.Println()
	fmt.Println("  # Clear all cached images")
	fmt.Println("  mtga-companion cards cache-clear")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  MTGA_DB_PATH     Path to database file (default: ~/.mtga-companion/data.db)")
	fmt.Println()
}

func runCardsCheckUpdates() {
	// Define flags for check-updates command
	checkFlags := flag.NewFlagSet("check-updates", flag.ExitOnError)
	if err := checkFlags.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Get database path
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Database not found. Run 'mtga-companion cards import' first.")
		os.Exit(1)
	}

	// Open database
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create storage service
	service := storage.NewService(db)
	defer func() { _ = service.Close() }()

	// Create Scryfall client
	scryfallClient := scryfall.NewClient()

	// Create updater
	options := updater.DefaultUpdateOptions()
	options.Progress = func(message string) {
		fmt.Println(message)
	}

	cardUpdater := updater.NewUpdater(scryfallClient, service, options)

	// Check for updates
	ctx := context.Background()
	fmt.Println("Checking for card data updates...")
	fmt.Println()

	info, err := cardUpdater.CheckUpdates(ctx)
	if err != nil {
		log.Fatalf("Error checking updates: %v", err)
	}

	// Display results
	if len(info.NewSets) == 0 && len(info.StaleSets) == 0 {
		fmt.Println("✓ Card data is up to date!")
		if !info.LastUpdate.IsZero() {
			fmt.Printf("  Last updated: %s\n", info.LastUpdate.Format("2006-01-02 15:04:05"))
		}
	} else {
		if len(info.NewSets) > 0 {
			fmt.Printf("New sets available (%d):\n", len(info.NewSets))
			for _, setCode := range info.NewSets {
				fmt.Printf("  - %s\n", setCode)
			}
			fmt.Printf("  Estimated new cards: %d\n", info.TotalNewCards)
			fmt.Println()
		}

		if len(info.StaleSets) > 0 {
			fmt.Printf("Stale sets (%d sets older than %d days):\n",
				info.TotalStaleSets,
				int(options.StaleSetThreshold.Hours()/24))
			for i, setCode := range info.StaleSets {
				if i < 10 {
					fmt.Printf("  - %s\n", setCode)
				}
			}
			if len(info.StaleSets) > 10 {
				fmt.Printf("  ... and %d more\n", len(info.StaleSets)-10)
			}
			fmt.Println()
		}

		fmt.Println("Run 'mtga-companion cards update' to apply updates.")
	}
}

func runCardsUpdate() {
	// Define flags for update command
	updateFlags := flag.NewFlagSet("update", flag.ExitOnError)
	forceUpdate := updateFlags.Bool("force", false, "Force update all sets")
	verbose := updateFlags.Bool("verbose", false, "Enable verbose output")
	specificSet := updateFlags.String("set", "", "Update specific set only")

	if err := updateFlags.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Get database path
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Database not found. Run 'mtga-companion cards import' first.")
		os.Exit(1)
	}

	// Open database
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create storage service
	service := storage.NewService(db)
	defer func() { _ = service.Close() }()

	// Create Scryfall client
	scryfallClient := scryfall.NewClient()

	// Create updater with options
	options := updater.DefaultUpdateOptions()
	options.Force = *forceUpdate
	options.Verbose = *verbose
	options.SpecificSet = *specificSet

	if *verbose {
		options.Progress = func(message string) {
			fmt.Println(message)
		}
	}

	cardUpdater := updater.NewUpdater(scryfallClient, service, options)

	// Run update
	ctx := context.Background()

	var stats *updater.UpdateStats
	var updateErr error

	if *specificSet != "" {
		fmt.Printf("Updating set: %s\n", *specificSet)
		stats, updateErr = cardUpdater.UpdateSpecificSet(ctx, *specificSet)
	} else {
		fmt.Println("Checking for updates...")
		stats, updateErr = cardUpdater.Update(ctx)
	}

	if updateErr != nil {
		log.Fatalf("Update failed: %v", updateErr)
	}

	// Print summary
	fmt.Println()
	fmt.Println("Update complete!")
	fmt.Printf("  Sets processed: %d\n", stats.SetsProcessed)
	fmt.Printf("  Sets updated: %d\n", stats.SetsUpdated)
	fmt.Printf("  Cards added: %d\n", stats.CardsAdded)
	fmt.Printf("  Cards updated: %d\n", stats.CardsUpdated)
	if stats.Errors > 0 {
		fmt.Printf("  Errors: %d\n", stats.Errors)
	}
	fmt.Printf("  Duration: %.2fs\n", stats.Duration.Seconds())
}

func runCardsAutoUpdate() {
	// Define flags for auto-update command
	autoFlags := flag.NewFlagSet("auto-update", flag.ExitOnError)
	forceUpdate := autoFlags.Bool("force", false, "Force update even if remote timestamp is not newer")
	verbose := autoFlags.Bool("verbose", false, "Enable verbose output")
	checkOnly := autoFlags.Bool("check-only", false, "Only check for updates without applying them")

	if err := autoFlags.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Get database path
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".mtga-companion", "data.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Database not found. Run 'mtga-companion cards import' first.")
		os.Exit(1)
	}

	// Open database
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create storage service
	service := storage.NewService(db)
	defer func() { _ = service.Close() }()

	// Create Scryfall client
	scryfallClient := scryfall.NewClient()

	// Create bulk import options
	importOptions := importer.DefaultBulkImportOptions()
	importOptions.Verbose = *verbose

	// Create bulk data updater
	bulkUpdater := updater.NewBulkDataUpdater(scryfallClient, service, importOptions)
	if *verbose {
		bulkUpdater.SetProgressCallback(func(message string) {
			fmt.Println(message)
		})
	}

	ctx := context.Background()

	// Check for updates
	fmt.Println("Checking for bulk data updates...")
	needsUpdate, remoteTime, err := bulkUpdater.CheckForUpdates(ctx)
	if err != nil {
		log.Fatalf("Error checking for updates: %v", err)
	}

	// Get last update time
	lastUpdate, err := service.GetLastBulkDataUpdate(ctx)
	if err != nil {
		log.Fatalf("Error getting last update time: %v", err)
	}

	if !lastUpdate.IsZero() {
		fmt.Printf("Last bulk data update: %s\n", lastUpdate.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("No previous bulk data update found")
	}

	if !remoteTime.IsZero() {
		fmt.Printf("Remote bulk data timestamp: %s\n", remoteTime.Format("2006-01-02 15:04:05"))
	}

	if *forceUpdate {
		fmt.Println("Force update enabled - will update regardless of timestamps")
		needsUpdate = true
	}

	if !needsUpdate {
		fmt.Println("✓ Bulk data is up to date!")
		return
	}

	fmt.Println("Updates available!")

	if *checkOnly {
		fmt.Println("Run 'mtga-companion cards auto-update' to apply updates.")
		return
	}

	// Perform update
	fmt.Println("Starting bulk data update...")
	updated, err := bulkUpdater.UpdateIfNeeded(ctx)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	if updated {
		fmt.Println("✓ Bulk data update complete!")
	} else {
		fmt.Println("No update was performed (data is up to date)")
	}
}

func runCardsCacheClear() {
	// No flags for cache-clear
	if len(os.Args) > 3 {
		fmt.Println("cache-clear command takes no arguments")
		os.Exit(1)
	}

	// Create cache with default options
	options := imagecache.DefaultCacheOptions()
	cache, err := imagecache.NewCache(options)
	if err != nil {
		log.Fatalf("Error creating cache: %v", err)
	}

	fmt.Printf("Clearing image cache at %s...\n", options.CacheDir)

	if err := cache.Clear(); err != nil {
		log.Fatalf("Error clearing cache: %v", err)
	}

	fmt.Println("✓ Image cache cleared successfully!")
}

func runCardsCacheStats() {
	// No flags for cache-stats
	if len(os.Args) > 3 {
		fmt.Println("cache-stats command takes no arguments")
		os.Exit(1)
	}

	// Create cache with default options
	options := imagecache.DefaultCacheOptions()
	cache, err := imagecache.NewCache(options)
	if err != nil {
		log.Fatalf("Error creating cache: %v", err)
	}

	stats := cache.GetCacheStats()

	fmt.Println("Image Cache Statistics")
	fmt.Println("=====================")
	fmt.Printf("Cache directory: %s\n", stats.CacheDir)
	fmt.Printf("Total files:     %d\n", stats.TotalFiles)
	fmt.Printf("Total size:      %.2f MB\n", float64(stats.TotalSize)/(1024*1024))
	if stats.MaxSize > 0 {
		fmt.Printf("Max size:        %.2f MB\n", float64(stats.MaxSize)/(1024*1024))
		percentUsed := float64(stats.TotalSize) / float64(stats.MaxSize) * 100
		fmt.Printf("Usage:           %.1f%%\n", percentUsed)
	} else {
		fmt.Println("Max size:        Unlimited")
	}
}

func runCardsCheckMigrations() {
	// Define flags for check-migrations command
	checkFlags := flag.NewFlagSet("check-migrations", flag.ExitOnError)
	verbose := checkFlags.Bool("verbose", false, "Enable verbose output")

	if err := checkFlags.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Get database path
	dbPath := getDBPath()

	// Open database with auto-migrate
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create storage service
	svc := storage.NewService(db)

	// Create Scryfall client
	scryfallClient := scryfall.NewClient()

	// Fetch migrations from Scryfall
	fmt.Println("Checking for Scryfall migrations...")
	migrationList, err := scryfallClient.GetMigrations(ctx)
	if err != nil {
		log.Fatalf("Error fetching migrations: %v", err)
	}

	// Get already processed migration IDs
	processedIDs, err := svc.GetProcessedMigrationIDs(ctx)
	if err != nil {
		log.Fatalf("Error getting processed migrations: %v", err)
	}

	processedMap := make(map[string]bool)
	for _, id := range processedIDs {
		processedMap[id] = true
	}

	// Count new migrations
	newMigrations := 0
	mergeCount := 0
	deleteCount := 0

	for _, migration := range migrationList.Data {
		if !processedMap[migration.ID] {
			newMigrations++
			switch migration.MigrationStrategy {
			case "merge":
				mergeCount++
				if *verbose {
					newID := ""
					if migration.NewScryfallID != nil {
						newID = *migration.NewScryfallID
					}
					fmt.Printf("  [MERGE] %s -> %s (%s)\n", migration.OldScryfallID, newID, migration.Note)
				}
			case "delete":
				deleteCount++
				if *verbose {
					fmt.Printf("  [DELETE] %s (%s)\n", migration.OldScryfallID, migration.Note)
				}
			}
		}
	}

	fmt.Println()
	fmt.Printf("Total migrations from Scryfall: %d\n", len(migrationList.Data))
	fmt.Printf("Already processed:              %d\n", len(processedIDs))
	fmt.Printf("New migrations:                 %d\n", newMigrations)
	fmt.Printf("  - Merges:                     %d\n", mergeCount)
	fmt.Printf("  - Deletes:                    %d\n", deleteCount)

	if newMigrations > 0 {
		fmt.Println()
		fmt.Println("Run 'cards apply-migrations' to apply these migrations.")
	}
}

func runCardsApplyMigrations() {
	// Define flags for apply-migrations command
	applyFlags := flag.NewFlagSet("apply-migrations", flag.ExitOnError)
	verbose := applyFlags.Bool("verbose", false, "Enable verbose output")

	if err := applyFlags.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Get database path
	dbPath := getDBPath()

	// Open database with auto-migrate
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create storage service
	svc := storage.NewService(db)

	// Create Scryfall client
	scryfallClient := scryfall.NewClient()

	// Create migration checker
	checkerOptions := migrations.DefaultCheckerOptions()
	if *verbose {
		checkerOptions.VerboseProgress = func(msg string) {
			fmt.Println(msg)
		}
	}

	checker := migrations.NewChecker(scryfallClient, svc, checkerOptions)

	// Check and apply migrations
	fmt.Println("Applying Scryfall migrations...")
	result, err := checker.CheckAndApply(ctx)
	if err != nil {
		log.Fatalf("Error applying migrations: %v", err)
	}

	fmt.Println()
	fmt.Println("Migration Results")
	fmt.Println("=================")
	fmt.Printf("Total migrations checked:  %d\n", result.TotalMigrations)
	fmt.Printf("New migrations applied:    %d\n", result.NewMigrations)
	fmt.Printf("  - Merges:                %d\n", result.MergeCount)
	fmt.Printf("  - Deletes:               %d\n", result.DeleteCount)
	fmt.Printf("Cards updated:             %d\n", result.CardsUpdated)
	fmt.Printf("Cards deleted:             %d\n", result.CardsDeleted)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered:        %d\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}

	fmt.Println()
	fmt.Println("✓ Migration application completed!")
}

func runCardsMigrationStats() {
	// No flags for migration-stats
	if len(os.Args) > 3 {
		fmt.Println("migration-stats command takes no arguments")
		os.Exit(1)
	}

	// Get database path
	dbPath := getDBPath()

	// Open database with auto-migrate
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create storage service
	svc := storage.NewService(db)

	// Get migration stats
	stats, err := svc.GetMigrationStats(ctx)
	if err != nil {
		log.Fatalf("Error getting migration stats: %v", err)
	}

	fmt.Println("Migration Statistics")
	fmt.Println("====================")
	fmt.Printf("Total migrations processed: %d\n", stats.TotalMigrations)
	fmt.Printf("  - Merges:                 %d\n", stats.MergeCount)
	fmt.Printf("  - Deletes:                %d\n", stats.DeleteCount)

	if stats.LastProcessedAt != nil {
		fmt.Printf("Last processed:             %s\n", stats.LastProcessedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Last processed:             Never")
	}
}

// Draft Stats Commands

func runDraftStatsCommand() {
	if len(os.Args) < 3 {
		printDraftStatsUsage()
		os.Exit(1)
	}

	command := os.Args[2]

	switch command {
	case "update":
		runDraftStatsUpdate()
	case "check":
		runDraftStatsCheck()
	case "history":
		runDraftStatsHistory()
	case "trends":
		runDraftStatsTrends()
	case "compare":
		runDraftStatsCompare()
	case "cleanup":
		runDraftStatsCleanup()
	default:
		fmt.Printf("Unknown draft-stats command: %s\n\n", command)
		printDraftStatsUsage()
		os.Exit(1)
	}
}

func printDraftStatsUsage() {
	fmt.Println("Usage: mtga-companion draft-stats <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  update        Update draft statistics from 17Lands")
	fmt.Println("  check         Check for stale draft statistics")
	fmt.Println("  history       View rating history for a card")
	fmt.Println("  trends        View win rate trends for a set")
	fmt.Println("  compare       Compare early vs late draft meta")
	fmt.Println("  cleanup       Clean up old snapshots")
	fmt.Println()
	fmt.Println("Update Options:")
	fmt.Println("  --set <code>  Update specific set (e.g., --set BLB)")
	fmt.Println("  --all         Update all sets (not just stale ones)")
	fmt.Println("  --force       Force update even if data is fresh")
	fmt.Println()
	fmt.Println("History Options:")
	fmt.Println("  --arena-id <id>  Card's Arena ID (required)")
	fmt.Println("  --set <code>     Expansion code (required)")
	fmt.Println("  --format <fmt>   Output format: csv, json (default: json)")
	fmt.Println("  --output <file>  Write to file instead of stdout")
	fmt.Println()
	fmt.Println("Trends Options:")
	fmt.Println("  --set <code>     Expansion code (required)")
	fmt.Println("  --arena-id <id>  Specific card (optional, shows all if omitted)")
	fmt.Println("  --days <n>       Number of days to include (default: 30)")
	fmt.Println("  --format <fmt>   Output format: csv, json (default: json)")
	fmt.Println("  --output <file>  Write to file instead of stdout")
	fmt.Println()
	fmt.Println("Compare Options:")
	fmt.Println("  --set <code>     Expansion code (required)")
	fmt.Println("  --early <days>   Early period length in days (default: 14)")
	fmt.Println("  --late <days>    Late period length in days (default: 7)")
	fmt.Println("  --format <fmt>   Output format: csv, json (default: json)")
	fmt.Println("  --output <file>  Write to file instead of stdout")
	fmt.Println()
	fmt.Println("Cleanup Options:")
	fmt.Println("  --dry-run        Show what would be deleted without deleting")
	fmt.Println("  --min-age <days> Minimum age before deletion (default: 90)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mtga-companion draft-stats update")
	fmt.Println("  mtga-companion draft-stats update --set BLB")
	fmt.Println("  mtga-companion draft-stats check")
	fmt.Println("  mtga-companion draft-stats history --arena-id 12345 --set BLB")
	fmt.Println("  mtga-companion draft-stats trends --set BLB --days 30")
	fmt.Println("  mtga-companion draft-stats compare --set BLB")
	fmt.Println("  mtga-companion draft-stats cleanup --dry-run")
}

func runDraftStatsUpdate() {
	// Parse flags
	fs := flag.NewFlagSet("draft-stats update", flag.ExitOnError)
	setCode := fs.String("set", "", "Specific set code to update")
	updateAll := fs.Bool("all", false, "Update all sets")
	force := fs.Bool("force", false, "Force update even if fresh")

	if err := fs.Parse(os.Args[3:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, ".mtga-companion", "data.db")
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

	service := storage.NewService(db)

	// Create clients
	scryfallClient := scryfall.NewClient()
	seventeenlandsClient := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	// Create updater
	updater, err := draftdata.NewUpdater(draftdata.UpdaterConfig{
		ScryfallClient:       scryfallClient,
		SeventeenLandsClient: seventeenlandsClient,
		Storage:              service,
	})
	if err != nil {
		log.Fatalf("Error creating updater: %v", err)
	}

	ctx := context.Background()

	// Update specific set
	if *setCode != "" {
		fmt.Printf("Updating draft statistics for set: %s\n", *setCode)
		result, err := updater.UpdateSet(ctx, *setCode)
		if err != nil {
			log.Fatalf("Error updating set %s: %v", *setCode, err)
		}

		fmt.Println()
		if result.Success {
			fmt.Println("✓ Update successful!")
			fmt.Printf("  Card ratings: %d\n", result.CardRatings)
			fmt.Printf("  Color ratings: %d\n", result.ColorRatings)
			fmt.Printf("  Duration: %v\n", result.Duration)
		} else {
			fmt.Println("✗ Update failed")
			if result.Error != nil {
				fmt.Printf("  Error: %v\n", result.Error)
			}
		}
		return
	}

	// Update all active sets
	fmt.Println("Updating draft statistics for active sets...")
	fmt.Println()

	updated, skipped, results, err := updater.UpdateActiveSets(ctx)
	if err != nil {
		log.Fatalf("Error updating active sets: %v", err)
	}

	fmt.Println()
	fmt.Println("Update complete!")
	fmt.Printf("  Sets updated: %d\n", updated)
	fmt.Printf("  Sets skipped (fresh): %d\n", skipped)
	fmt.Println()

	if len(results) > 0 {
		fmt.Println("Details:")
		for _, result := range results {
			if result.Success {
				fmt.Printf("  ✓ %s: %d card ratings, %d color ratings (%v)\n",
					result.SetCode, result.CardRatings, result.ColorRatings, result.Duration)
			} else {
				fmt.Printf("  ✗ %s: %v\n", result.SetCode, result.Error)
			}
		}
	}

	_ = updateAll
	_ = force
}

func runDraftStatsCheck() {
	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, ".mtga-companion", "data.db")
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

	service := storage.NewService(db)

	// Create clients
	scryfallClient := scryfall.NewClient()
	seventeenlandsClient := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	// Create updater
	updater, err := draftdata.NewUpdater(draftdata.UpdaterConfig{
		ScryfallClient:       scryfallClient,
		SeventeenLandsClient: seventeenlandsClient,
		Storage:              service,
	})
	if err != nil {
		log.Fatalf("Error creating updater: %v", err)
	}

	ctx := context.Background()

	// Check for stale data
	stale, err := updater.CheckStaleness(ctx)
	if err != nil {
		log.Fatalf("Error checking staleness: %v", err)
	}

	fmt.Println("Draft Statistics Staleness Check")
	fmt.Println("=================================")
	fmt.Println()

	if len(stale) == 0 {
		fmt.Println("✓ All draft statistics are up to date!")
		return
	}

	fmt.Printf("Found %d stale data entries:\n", len(stale))
	fmt.Println()

	// Group by expansion
	byExpansion := make(map[string][]*storage.DraftCardRating)
	for _, s := range stale {
		byExpansion[s.Expansion] = append(byExpansion[s.Expansion], s)
	}

	for expansion, entries := range byExpansion {
		fmt.Printf("  %s:\n", expansion)
		for _, entry := range entries {
			age := time.Since(entry.LastUpdated)
			fmt.Printf("    Format: %s, Colors: %s, Age: %v\n",
				entry.Format, entry.Colors, age.Round(time.Hour))
		}
	}

	fmt.Println()
	fmt.Println("Run 'mtga-companion draft-stats update' to refresh stale data.")
}

func runDraftStatsHistory() {
	// Parse flags
	fs := flag.NewFlagSet("draft-stats history", flag.ExitOnError)
	arenaID := fs.Int("arena-id", 0, "Card's Arena ID")
	setCode := fs.String("set", "", "Expansion code")
	format := fs.String("format", "json", "Output format: csv, json")
	output := fs.String("output", "", "Output file path (optional)")

	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Validate required flags
	if *arenaID == 0 {
		fmt.Println("Error: --arena-id is required")
		fmt.Println()
		printDraftStatsUsage()
		os.Exit(1)
	}
	if *setCode == "" {
		fmt.Println("Error: --set is required")
		fmt.Println()
		printDraftStatsUsage()
		os.Exit(1)
	}

	// Open database
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, ".mtga-companion", "data.db")
	}
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	service := storage.NewService(db)
	ctx := context.Background()

	// Get rating history
	history, err := service.GetRatingHistory(ctx, *arenaID, *setCode)
	if err != nil {
		log.Fatalf("Error getting rating history: %v", err)
	}

	if len(history) == 0 {
		fmt.Printf("No rating history found for Arena ID %d in set %s\n", *arenaID, *setCode)
		return
	}

	// Determine output writer
	var writer io.Writer = os.Stdout
	if *output != "" {
		file, err := os.Create(*output)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer func() { _ = file.Close() }()
		writer = file
	}

	// Export history
	if err := export.ExportRatingHistory(writer, history, *format); err != nil {
		log.Fatalf("Error exporting history: %v", err)
	}

	if *output != "" {
		fmt.Printf("✓ Rating history exported to %s (%d snapshots)\n", *output, len(history))
	}
}

func runDraftStatsTrends() {
	// Parse flags
	fs := flag.NewFlagSet("draft-stats trends", flag.ExitOnError)
	setCode := fs.String("set", "", "Expansion code")
	arenaID := fs.Int("arena-id", 0, "Card's Arena ID (optional)")
	days := fs.Int("days", 30, "Number of days to include")
	format := fs.String("format", "json", "Output format: csv, json")
	output := fs.String("output", "", "Output file path (optional)")

	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Validate required flags
	if *setCode == "" {
		fmt.Println("Error: --set is required")
		fmt.Println()
		printDraftStatsUsage()
		os.Exit(1)
	}

	// Open database
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, ".mtga-companion", "data.db")
	}
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	service := storage.NewService(db)
	ctx := context.Background()

	// Get trend data
	if *arenaID > 0 {
		// Single card trend
		trend, err := service.GetCardWinRateTrend(ctx, *arenaID, *setCode, *days)
		if err != nil {
			log.Fatalf("Error getting card trend: %v", err)
		}

		if len(trend.Points) == 0 {
			fmt.Printf("No trend data found for Arena ID %d in set %s\n", *arenaID, *setCode)
			return
		}

		// Determine output writer
		var writer io.Writer = os.Stdout
		if *output != "" {
			file, err := os.Create(*output)
			if err != nil {
				log.Fatalf("Error creating output file: %v", err)
			}
			defer func() { _ = file.Close() }()
			writer = file
		}

		// Export trend
		if err := export.ExportCardTrend(writer, trend, *format); err != nil {
			log.Fatalf("Error exporting trend: %v", err)
		}

		if *output != "" {
			fmt.Printf("✓ Card trend exported to %s (%d data points)\n", *output, len(trend.Points))
		}
	} else {
		// All cards in expansion
		trends, err := service.GetExpansionTrends(ctx, *setCode, *days)
		if err != nil {
			log.Fatalf("Error getting expansion trends: %v", err)
		}

		if len(trends) == 0 {
			fmt.Printf("No trend data found for set %s\n", *setCode)
			return
		}

		fmt.Printf("Found trends for %d cards in %s\n", len(trends), *setCode)
		fmt.Println("Use --arena-id <id> to view a specific card's trend")
		fmt.Println()

		// Show summary
		for arenaID, trend := range trends {
			fmt.Printf("  Arena ID %d: %d data points\n", arenaID, len(trend.Points))
		}
	}
}

func runDraftStatsCompare() {
	// Parse flags
	fs := flag.NewFlagSet("draft-stats compare", flag.ExitOnError)
	setCode := fs.String("set", "", "Expansion code")
	earlyDays := fs.Int("early", 14, "Early period length in days")
	lateDays := fs.Int("late", 7, "Late period length in days")
	format := fs.String("format", "json", "Output format: csv, json")
	output := fs.String("output", "", "Output file path (optional)")

	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Validate required flags
	if *setCode == "" {
		fmt.Println("Error: --set is required")
		fmt.Println()
		printDraftStatsUsage()
		os.Exit(1)
	}

	// Open database
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, ".mtga-companion", "data.db")
	}
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	service := storage.NewService(db)
	ctx := context.Background()

	// Compare meta periods
	comp, err := service.CompareMetaPeriods(ctx, *setCode, *earlyDays, *lateDays)
	if err != nil {
		log.Fatalf("Error comparing meta periods: %v", err)
	}

	// Determine output writer
	var writer io.Writer = os.Stdout
	if *output != "" {
		file, err := os.Create(*output)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer func() { _ = file.Close() }()
		writer = file
	}

	// Export comparison
	if err := export.ExportMetaComparison(writer, comp, *format); err != nil {
		log.Fatalf("Error exporting comparison: %v", err)
	}

	if *output != "" {
		fmt.Printf("✓ Meta comparison exported to %s\n", *output)
		fmt.Printf("  Early period: %d cards\n", comp.EarlyCards)
		fmt.Printf("  Late period: %d cards\n", comp.LateCards)
		fmt.Printf("  Top improvers: %d\n", len(comp.TopImprovers))
		fmt.Printf("  Top decliners: %d\n", len(comp.TopDecliners))
	}
}

func runDraftStatsCleanup() {
	// Parse flags
	fs := flag.NewFlagSet("draft-stats cleanup", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without deleting")
	minAge := fs.Int("min-age", 90, "Minimum age in days before deletion")

	if err := fs.Parse(os.Args[3:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Open database
	dbPath := os.Getenv("MTGA_DB_PATH")
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, ".mtga-companion", "data.db")
	}
	config := storage.DefaultConfig(dbPath)
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer func() { _ = db.Close() }()

	service := storage.NewService(db)
	ctx := context.Background()

	// Configure retention policy
	policy := storage.DefaultRetentionPolicy()
	policy.MinimumAge = time.Duration(*minAge) * 24 * time.Hour

	if *dryRun {
		fmt.Println("DRY RUN: Analyzing snapshots for cleanup...")
		fmt.Println()
	} else {
		fmt.Println("Cleaning up old draft statistics snapshots...")
		fmt.Println()
	}

	// Run cleanup
	result, err := service.CleanupOldSnapshots(ctx, policy, *dryRun)
	if err != nil {
		log.Fatalf("Error during cleanup: %v", err)
	}

	// Display results
	fmt.Printf("Cleanup Results:\n")
	fmt.Printf("================\n")
	fmt.Printf("Total snapshots:    %d\n", result.TotalSnapshots)
	fmt.Printf("Removed snapshots:  %d\n", result.RemovedSnapshots)
	fmt.Printf("Retained snapshots: %d\n", result.RetainedSnapshots)
	if !result.OldestSnapshot.IsZero() {
		fmt.Printf("Oldest snapshot:    %s\n", result.OldestSnapshot.Format("2006-01-02"))
	}
	if !result.NewestSnapshot.IsZero() {
		fmt.Printf("Newest snapshot:    %s\n", result.NewestSnapshot.Format("2006-01-02"))
	}
	fmt.Println()

	if len(result.RemovedBySet) > 0 {
		fmt.Println("Per-Set Breakdown:")
		for set, count := range result.RemovedBySet {
			retained := result.RetainedBySet[set]
			fmt.Printf("  %s: %d removed, %d retained\n", set, count, retained)
		}
		fmt.Println()
	}

	if *dryRun {
		fmt.Println("This was a dry run. No snapshots were actually deleted.")
		fmt.Println("Run without --dry-run to perform actual cleanup.")
	} else {
		fmt.Println("✓ Cleanup complete!")
	}
}

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/export"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
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

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbPath)
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
		// Get backup directory from environment or use default
		backupDir := os.Getenv("MTGA_BACKUP_DIR")
		backupName := ""
		if len(os.Args) >= 4 {
			backupName = os.Args[3]
		}

		config := storage.DefaultBackupConfig()
		config.BackupDir = backupDir
		config.BackupName = backupName
		config.VerifyBackup = true

		fmt.Println("Creating database backup...")
		backupPath, err := backupMgr.Backup(config)
		if err != nil {
			log.Fatalf("Error creating backup: %v", err)
		}

		fmt.Printf("Backup created successfully: %s\n", backupPath)

		// Calculate and display backup size
		info, err := os.Stat(backupPath)
		if err == nil {
			sizeMB := float64(info.Size()) / (1024 * 1024)
			fmt.Printf("Backup size: %.2f MB\n", sizeMB)
		}

	case "restore":
		if len(os.Args) < 4 {
			fmt.Println("Error: restore command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup restore <backup-file>")
			os.Exit(1)
		}
		backupPath := os.Args[3]

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

		fmt.Println("\nRestoring database from backup...")
		if err := backupMgr.Restore(backupPath); err != nil {
			log.Fatalf("Error restoring backup: %v", err)
		}

		fmt.Println("Database restored successfully!")

	case "list", "ls":
		backupDir := os.Getenv("MTGA_BACKUP_DIR")
		if backupDir == "" {
			backupDir = backupMgr.GetBackupDir()
		}

		fmt.Println("Listing backups...")
		backups, err := backupMgr.ListBackups(backupDir)
		if err != nil {
			log.Fatalf("Error listing backups: %v", err)
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

	case "verify":
		if len(os.Args) < 4 {
			fmt.Println("Error: verify command requires a backup file path")
			fmt.Println("Usage: mtga-companion backup verify <backup-file>")
			os.Exit(1)
		}
		backupPath := os.Args[3]

		fmt.Printf("Verifying backup: %s\n", backupPath)
		if err := backupMgr.VerifyBackup(backupPath); err != nil {
			log.Fatalf("Backup verification failed: %v", err)
		}

		fmt.Println("Backup verification successful!")

	default:
		fmt.Printf("Unknown backup command: %s\n\n", command)
		printBackupUsage()
		os.Exit(1)
	}
}

func printBackupUsage() {
	fmt.Println("Usage: mtga-companion backup <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create, backup [name]  - Create a new backup (optional name)")
	fmt.Println("  restore <backup-file>  - Restore database from backup")
	fmt.Println("  list, ls               - List all available backups")
	fmt.Println("  verify <backup-file>    - Verify backup integrity")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  MTGA_DB_PATH     - Path to the database file (default: ~/.mtga-companion/data.db)")
	fmt.Println("  MTGA_BACKUP_DIR  - Directory for backups (default: ~/.mtga-companion/backups)")
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
	periodType := "daily" // Default for trend exports
	recentDays := 30      // Default for result comparison exports

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

	// Create filter for statistics exports
	filter := models.StatsFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Format:    formatFilter,
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
	fmt.Println("  export deck/decks [options]       - Export deck lists")
	fmt.Println("\nOptions:")
	fmt.Println("  -json                             - Export as JSON (default: CSV)")
	fmt.Println("  -csv                              - Export as CSV")
	fmt.Println("  -o, --output <path>               - Specify output file path")
	fmt.Println("  -start, --start-date <date>       - Start date (YYYY-MM-DD)")
	fmt.Println("  -end, --end-date <date>           - End date (YYYY-MM-DD)")
	fmt.Println("  -format, --format <format>        - Filter by format (constructed/limited)")
	fmt.Println("  -period, --period <type>          - Period type: daily, weekly, monthly (default: daily)")
	fmt.Println("  -recent, --recent-days <days>     - Recent period in days (default: 30)")
	fmt.Println("\nExamples:")
	fmt.Println("  export matches -json              - Export all matches as JSON")
	fmt.Println("  export stats -csv -o stats.csv")
	fmt.Println("  export trends -start 2024-01-01 -end 2024-12-31 -period daily")
	fmt.Println("  export results -json              - Export result breakdown as JSON")
	fmt.Println("  export result-comparison -recent 7 -format constructed")
	fmt.Println("  export result-trends -start 2024-01-01 -end 2024-12-31 -period weekly")
	fmt.Println("  export daily -start 2024-01-01 -end 2024-01-31")
	fmt.Println("  export matches -format constructed -json")
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

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/stats"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service provides high-level operations for storing and retrieving MTGA data.
type Service struct {
	db                      *DB
	matches                 repository.MatchRepository
	stats                   repository.StatsRepository
	decks                   repository.DeckRepository
	collection              repository.CollectionRepository
	accounts                repository.AccountRepository
	rankHistory             repository.RankHistoryRepository
	quests                  *QuestRepository
	draft                   repository.DraftRepository
	setCard                 repository.SetCardRepository
	draftRatings            repository.DraftRatingsRepository
	inventory               repository.InventoryRepository
	settings                repository.SettingsRepository
	deckPerformance         repository.DeckPerformanceRepository
	recommendationFeedback  repository.RecommendationFeedbackRepository
	standard                repository.StandardRepository
	gamePlay                repository.GamePlayRepository
	cardPerformanceAnalysis repository.CardPerformanceRepository
	currentAccountID        int // Current active account ID
}

// ServiceConfig holds optional repository overrides for dependency injection.
// All fields are optional - if nil, the default implementation will be created.
type ServiceConfig struct {
	Matches                 repository.MatchRepository
	Stats                   repository.StatsRepository
	Decks                   repository.DeckRepository
	Collection              repository.CollectionRepository
	Accounts                repository.AccountRepository
	RankHistory             repository.RankHistoryRepository
	Quests                  *QuestRepository
	Draft                   repository.DraftRepository
	SetCard                 repository.SetCardRepository
	DraftRatings            repository.DraftRatingsRepository
	Inventory               repository.InventoryRepository
	Settings                repository.SettingsRepository
	DeckPerformance         repository.DeckPerformanceRepository
	RecommendationFeedback  repository.RecommendationFeedbackRepository
	Standard                repository.StandardRepository
	GamePlay                repository.GamePlayRepository
	CardPerformanceAnalysis repository.CardPerformanceRepository
}

// NewService creates a new storage service with default repository implementations.
func NewService(db *DB) *Service {
	return NewServiceWithConfig(db, nil)
}

// NewServiceWithConfig creates a new storage service with optional repository overrides.
// Any nil repository in the config will use the default implementation.
func NewServiceWithConfig(db *DB, cfg *ServiceConfig) *Service {
	if cfg == nil {
		cfg = &ServiceConfig{}
	}

	conn := db.Conn()

	svc := &Service{
		db:              db,
		matches:         orDefault(cfg.Matches, func() repository.MatchRepository { return repository.NewMatchRepository(conn) }),
		stats:           orDefault(cfg.Stats, func() repository.StatsRepository { return repository.NewStatsRepository(conn) }),
		decks:           orDefault(cfg.Decks, func() repository.DeckRepository { return repository.NewDeckRepository(conn) }),
		collection:      orDefault(cfg.Collection, func() repository.CollectionRepository { return repository.NewCollectionRepository(conn) }),
		accounts:        orDefault(cfg.Accounts, func() repository.AccountRepository { return repository.NewAccountRepository(conn) }),
		rankHistory:     orDefault(cfg.RankHistory, func() repository.RankHistoryRepository { return repository.NewRankHistoryRepository(conn) }),
		quests:          orDefaultQuest(cfg.Quests, func() *QuestRepository { return NewQuestRepository(conn) }),
		draft:           orDefault(cfg.Draft, func() repository.DraftRepository { return repository.NewDraftRepository(conn) }),
		setCard:         orDefault(cfg.SetCard, func() repository.SetCardRepository { return repository.NewSetCardRepository(conn) }),
		draftRatings:    orDefault(cfg.DraftRatings, func() repository.DraftRatingsRepository { return repository.NewDraftRatingsRepository(conn) }),
		inventory:       orDefault(cfg.Inventory, func() repository.InventoryRepository { return repository.NewInventoryRepository(conn) }),
		settings:        orDefault(cfg.Settings, func() repository.SettingsRepository { return repository.NewSettingsRepository(conn) }),
		deckPerformance: orDefault(cfg.DeckPerformance, func() repository.DeckPerformanceRepository { return repository.NewDeckPerformanceRepository(conn) }),
		recommendationFeedback: orDefault(cfg.RecommendationFeedback, func() repository.RecommendationFeedbackRepository {
			return repository.NewRecommendationFeedbackRepository(conn)
		}),
		standard: orDefault(cfg.Standard, func() repository.StandardRepository { return repository.NewStandardRepository(conn) }),
		gamePlay: orDefault(cfg.GamePlay, func() repository.GamePlayRepository { return repository.NewGamePlayRepository(conn) }),
		cardPerformanceAnalysis: orDefault(cfg.CardPerformanceAnalysis, func() repository.CardPerformanceRepository {
			return repository.NewCardPerformanceRepository(conn)
		}),
	}

	// Initialize default account if it doesn't exist
	ctx := context.Background()
	defaultAccount, err := svc.accounts.GetDefault(ctx)
	if err != nil {
		// Log error but continue - account will be created on first use
		_ = err
	}
	if defaultAccount == nil {
		// Create default account
		now := time.Now()
		defaultAccount = &models.Account{
			Name:      "Default Account",
			IsDefault: true,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := svc.accounts.Create(ctx, defaultAccount); err != nil {
			// Log error but continue
			_ = err
		}
	}
	if defaultAccount != nil {
		svc.currentAccountID = defaultAccount.ID
	}

	return svc
}

// orDefault returns the provided value if not nil, otherwise calls the factory function.
func orDefault[T any](val T, factory func() T) T {
	// Use reflection to check if val is nil (works for interface types)
	if any(val) == nil {
		return factory()
	}
	return val
}

// orDefaultQuest is a specialized version for *QuestRepository since it's a concrete type.
func orDefaultQuest(val *QuestRepository, factory func() *QuestRepository) *QuestRepository {
	if val == nil {
		return factory()
	}
	return val
}

// GetDB returns the underlying database connection.
func (s *Service) GetDB() *sql.DB {
	return s.db.Conn()
}

// CurrentAccountID returns the current active account ID.
func (s *Service) CurrentAccountID() int {
	return s.currentAccountID
}

// BulkImportSettings stores the original database settings before bulk import mode.
type BulkImportSettings struct {
	Synchronous string
	JournalMode string
	CacheSize   int
}

// EnableBulkImportMode configures the database for fast bulk imports.
// This trades durability for speed - only use for recoverable operations like historical replay.
// Always call RestoreSafeMode() when done (use defer).
func (s *Service) EnableBulkImportMode(ctx context.Context) (*BulkImportSettings, error) {
	db := s.db.Conn()

	// Save current settings
	settings := &BulkImportSettings{}

	if err := db.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&settings.Synchronous); err != nil {
		return nil, fmt.Errorf("failed to get synchronous setting: %w", err)
	}

	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&settings.JournalMode); err != nil {
		return nil, fmt.Errorf("failed to get journal_mode setting: %w", err)
	}

	if err := db.QueryRowContext(ctx, "PRAGMA cache_size").Scan(&settings.CacheSize); err != nil {
		return nil, fmt.Errorf("failed to get cache_size setting: %w", err)
	}

	log.Printf("Enabling bulk import mode (saved settings: sync=%s, journal=%s, cache=%d)",
		settings.Synchronous, settings.JournalMode, settings.CacheSize)

	// Enable bulk import optimizations
	if _, err := db.ExecContext(ctx, "PRAGMA synchronous = OFF"); err != nil {
		return nil, fmt.Errorf("failed to set synchronous: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = MEMORY"); err != nil {
		return nil, fmt.Errorf("failed to set journal_mode: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA cache_size = -64000"); err != nil {
		return nil, fmt.Errorf("failed to set cache_size: %w", err)
	}

	log.Println("Bulk import mode enabled: synchronous=OFF, journal_mode=MEMORY, cache_size=64MB")

	return settings, nil
}

// RestoreSafeMode restores the database to safe operation mode after bulk import.
func (s *Service) RestoreSafeMode(ctx context.Context, settings *BulkImportSettings) error {
	if settings == nil {
		return nil
	}

	db := s.db.Conn()

	log.Printf("Restoring safe mode (sync=%s, journal=%s, cache=%d)",
		settings.Synchronous, settings.JournalMode, settings.CacheSize)

	// Restore original settings
	if _, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA synchronous = %s", settings.Synchronous)); err != nil {
		return fmt.Errorf("failed to restore synchronous: %w", err)
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA journal_mode = %s", settings.JournalMode)); err != nil {
		return fmt.Errorf("failed to restore journal_mode: %w", err)
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA cache_size = %d", settings.CacheSize)); err != nil {
		return fmt.Errorf("failed to restore cache_size: %w", err)
	}

	// Force WAL checkpoint to ensure data is committed
	if _, err := db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		log.Printf("Warning: Failed to checkpoint WAL: %v", err)
	}

	log.Println("Safe mode restored and WAL checkpoint completed")

	return nil
}

// UpdateAccountFromEntries consolidates all account updates from log entries into a single operation.
// This replaces 3 separate fetch-update cycles with a single fetch-update, improving performance.
func (s *Service) UpdateAccountFromEntries(ctx context.Context, entries []*logreader.LogEntry) error {
	// Parse all data upfront
	profile, _ := logreader.ParseProfile(entries)
	periodicRewards, _ := logreader.ParsePeriodicRewards(entries)
	masteryPass, _ := logreader.ParseMasteryPass(entries)

	// If nothing to update, return early
	if profile == nil && periodicRewards == nil && masteryPass == nil {
		return nil
	}

	// Fetch account once
	account, err := s.accounts.GetByID(ctx, s.currentAccountID)
	if err != nil || account == nil {
		return err
	}

	// Apply all updates
	updated := false

	// Update profile (screen name, client ID)
	if profile != nil {
		if profile.ScreenName != "" && (account.ScreenName == nil || *account.ScreenName != profile.ScreenName) {
			account.ScreenName = &profile.ScreenName
			updated = true
		}
		if profile.ClientID != "" && (account.ClientID == nil || *account.ClientID != profile.ClientID) {
			account.ClientID = &profile.ClientID
			updated = true
		}
	}

	// Update daily/weekly wins
	if periodicRewards != nil {
		if periodicRewards.DailyWins != account.DailyWins {
			account.DailyWins = periodicRewards.DailyWins
			updated = true
		}
		if periodicRewards.WeeklyWins != account.WeeklyWins {
			account.WeeklyWins = periodicRewards.WeeklyWins
			updated = true
		}
	}

	// Update mastery pass
	if masteryPass != nil {
		if masteryPass.CurrentLevel != account.MasteryLevel {
			account.MasteryLevel = masteryPass.CurrentLevel
			updated = true
		}
		if masteryPass.PassType != "" && masteryPass.PassType != account.MasteryPass {
			account.MasteryPass = masteryPass.PassType
			updated = true
		}
		if masteryPass.MaxLevel != 0 && masteryPass.MaxLevel != account.MasteryMax {
			account.MasteryMax = masteryPass.MaxLevel
			updated = true
		}
	}

	// Single update at the end if anything changed
	if updated {
		account.UpdatedAt = time.Now()
		return s.accounts.Update(ctx, account)
	}

	return nil
}

// StoreArenaStats stores arena statistics parsed from the log.
// It creates match and game records with deduplication and updates daily stats.
func (s *Service) StoreArenaStats(ctx context.Context, arenaStats *logreader.ArenaStats, entries []*logreader.LogEntry) error {
	if arenaStats == nil {
		return nil
	}

	// Update account from all parsed data (consolidated to reduce DB queries)
	_ = s.UpdateAccountFromEntries(ctx, entries)

	// Extract match details from log entries for persistent storage
	matchesToStore, err := s.extractMatchesFromEntries(ctx, entries)
	if err != nil {
		return fmt.Errorf("failed to extract matches: %w", err)
	}

	// Correlate ranks with matches
	rankSnapshots := extractRankSnapshots(entries)
	correlateRanksWithMatches(matchesToStore, rankSnapshots)

	// Batch store all matches in single transaction (much faster than loop + individual queries)
	stored, err := s.BatchStoreMatches(ctx, matchesToStore)
	if err != nil {
		return fmt.Errorf("failed to batch store matches: %w", err)
	}

	if stored > 0 {
		log.Printf("Stored %d new matches (skipped %d duplicates)", stored, len(matchesToStore)-stored)
	}

	// Update daily stats for each format
	today := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	for eventName, formatStat := range arenaStats.FormatStats {
		// Get existing stats for today
		existing, err := s.stats.GetByDate(ctx, today, eventName)
		if err != nil {
			return fmt.Errorf("failed to get existing stats: %w", err)
		}

		// Calculate new totals (add current session to existing)
		matchesPlayed := formatStat.MatchesPlayed
		matchesWon := formatStat.MatchWins
		gamesPlayed := formatStat.GamesPlayed
		gamesWon := formatStat.GameWins

		if existing != nil {
			matchesPlayed += existing.MatchesPlayed
			matchesWon += existing.MatchesWon
			gamesPlayed += existing.GamesPlayed
			gamesWon += existing.GamesWon
		}

		stats := &PlayerStats{
			AccountID:     s.currentAccountID,
			Date:          today,
			Format:        eventName,
			MatchesPlayed: matchesPlayed,
			MatchesWon:    matchesWon,
			GamesPlayed:   gamesPlayed,
			GamesWon:      gamesWon,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.stats.Upsert(ctx, stats); err != nil {
			return fmt.Errorf("failed to store stats for %s: %w", eventName, err)
		}
	}

	return nil
}

// matchData holds a match and its associated games.
type matchData struct {
	Match *Match
	Games []*Game
}

// extractMatchesFromEntries extracts match and game details from log entries.
func (s *Service) extractMatchesFromEntries(ctx context.Context, entries []*logreader.LogEntry) ([]matchData, error) {
	var matches []matchData
	seenMatches := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Check for matchGameRoomStateChangedEvent
		if eventData, ok := entry.JSON["matchGameRoomStateChangedEvent"]; ok {
			eventMap, ok := eventData.(map[string]interface{})
			if !ok {
				continue
			}

			gameRoomInfo, ok := eventMap["gameRoomInfo"].(map[string]interface{})
			if !ok {
				continue
			}

			// Check if this is a match completion event
			finalMatchResult, hasFinalResult := gameRoomInfo["finalMatchResult"].(map[string]interface{})
			if !hasFinalResult {
				continue
			}

			// Get match ID
			matchID, _ := finalMatchResult["matchId"].(string)
			if matchID == "" || seenMatches[matchID] {
				continue
			}
			seenMatches[matchID] = true

			// Parse timestamp - try multiple sources in order of preference
			matchTime := time.Now()
			timestampFound := false

			// 1. Try JSON payload timestamp (Unix milliseconds)
			if tsVal, ok := entry.JSON["timestamp"]; ok {
				log.Printf("[ExtractMatches] Found timestamp field in JSON for match %s: %v (type: %T)", matchID, tsVal, tsVal)
				if tsStr, ok := tsVal.(string); ok {
					// Parse Unix milliseconds timestamp
					var tsMs int64
					if _, err := fmt.Sscanf(tsStr, "%d", &tsMs); err == nil {
						matchTime = time.Unix(tsMs/1000, (tsMs%1000)*1000000)
						timestampFound = true
						log.Printf("[ExtractMatches] ✓ Parsed JSON timestamp for match %s: %v (from %s ms)", matchID, matchTime, tsStr)
					} else {
						log.Printf("[ExtractMatches] Failed to parse JSON timestamp string '%s': %v", tsStr, err)
					}
				} else {
					log.Printf("[ExtractMatches] JSON timestamp field is not a string: %T", tsVal)
				}
			} else {
				log.Printf("[ExtractMatches] No 'timestamp' field found in JSON for match %s", matchID)
			}

			// 2. Try log entry prefix timestamp
			if !timestampFound && entry.Timestamp != "" {
				// Try to parse timestamp (format: [UnityCrossThreadLogger]11/16/2025 10:16:08 AM)
				if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
					matchTime = parsedTime
					timestampFound = true
				} else {
					log.Printf("[ExtractMatches] WARNING: Failed to parse prefix timestamp '%s' for match %s. Error: %v", entry.Timestamp, matchID, err)
				}
			}

			// 3. Fallback to current time
			if !timestampFound {
				log.Printf("[ExtractMatches] WARNING: No valid timestamp found for match %s, using current time", matchID)
			}

			// Get event information
			gameRoomConfig, ok := gameRoomInfo["gameRoomConfig"].(map[string]interface{})
			if !ok {
				continue
			}

			reservedPlayers, ok := gameRoomConfig["reservedPlayers"].([]interface{})
			if !ok || len(reservedPlayers) == 0 {
				continue
			}

			// Get current account's screen name for player identification
			var playerScreenName string
			account, err := s.accounts.GetByID(ctx, s.currentAccountID)
			if err == nil && account != nil && account.ScreenName != nil {
				playerScreenName = *account.ScreenName
			}

			// Find the actual player (not the opponent) by matching screen name
			var actualPlayer map[string]interface{}
			var opponentPlayer map[string]interface{}
			eventID := "Unknown"

			for _, playerData := range reservedPlayers {
				player, ok := playerData.(map[string]interface{})
				if !ok {
					continue
				}

				// Use the first player as fallback
				if actualPlayer == nil {
					actualPlayer = player
				}

				// Extract event ID from any player
				if evID, ok := player["eventId"].(string); ok && evID != "" {
					eventID = evID
				}

				// Match player by screen name if available
				if playerName, ok := player["playerName"].(string); ok && playerName != "" {
					if playerScreenName != "" && playerName == playerScreenName {
						// Found the actual player by screen name
						actualPlayer = player
					} else if playerScreenName != "" && playerName != playerScreenName {
						// This is the opponent (different screen name)
						opponentPlayer = player
					}
				}
			}

			if actualPlayer == nil {
				continue
			}

			eventName, _ := actualPlayer["eventId"].(string)
			if eventName == "" {
				eventName = eventID
			}

			playerTeamID, _ := actualPlayer["teamId"].(float64)

			// Extract deck ID if available
			var deckID *string
			if deckIDStr, ok := actualPlayer["deckId"].(string); ok && deckIDStr != "" {
				deckID = &deckIDStr
			} else if deckIDStr, ok := actualPlayer["DeckId"].(string); ok && deckIDStr != "" {
				deckID = &deckIDStr
			}

			// Parse result list to determine match result and games
			resultList, ok := finalMatchResult["resultList"].([]interface{})
			if !ok {
				continue
			}

			var matchResult string
			var resultReason *string
			var playerWins, opponentWins int
			var games []*Game
			gameNumber := 1

			for _, resultData := range resultList {
				resultMap, ok := resultData.(map[string]interface{})
				if !ok {
					continue
				}

				scope, _ := resultMap["scope"].(string)
				winningTeamID, _ := resultMap["winningTeamId"].(float64)
				playerWon := int(playerTeamID) == int(winningTeamID)

				// Extract result reason if available
				if reason, ok := resultMap["reason"].(string); ok && reason != "" {
					normalizedReason := normalizeResultReason(reason)
					resultReason = &normalizedReason
				} else if reason, ok := resultMap["Reason"].(string); ok && reason != "" {
					normalizedReason := normalizeResultReason(reason)
					resultReason = &normalizedReason
				}

				switch scope {
				case "MatchScope_Match":
					if playerWon {
						matchResult = "win"
					} else {
						matchResult = "loss"
					}
				case "MatchScope_Game":
					game := &Game{
						MatchID:    matchID,
						GameNumber: gameNumber,
						Result:     map[bool]string{true: "win", false: "loss"}[playerWon],
						CreatedAt:  matchTime,
					}
					games = append(games, game)
					gameNumber++
					// Track wins/losses from games
					if playerWon {
						playerWins++
					} else {
						opponentWins++
					}
				}
			}

			// If no result reason found in resultList, check finalMatchResult
			if resultReason == nil {
				if reason, ok := finalMatchResult["reason"].(string); ok && reason != "" {
					normalizedReason := normalizeResultReason(reason)
					resultReason = &normalizedReason
				} else if reason, ok := finalMatchResult["Reason"].(string); ok && reason != "" {
					normalizedReason := normalizeResultReason(reason)
					resultReason = &normalizedReason
				}
			}

			// If no games found, set match result based on match scope
			if len(games) == 0 && matchResult != "" {
				if matchResult == "win" {
					playerWins = 1
					opponentWins = 0
				} else {
					playerWins = 0
					opponentWins = 1
				}
			}

			// Extract opponent information if available
			var opponentName *string
			var opponentID *string
			if opponentPlayer != nil {
				if name, ok := opponentPlayer["playerName"].(string); ok && name != "" {
					opponentName = &name
				}
				if id, ok := opponentPlayer["userId"].(string); ok && id != "" {
					opponentID = &id
				}
			}

			// Create match record
			match := &Match{
				ID:           matchID,
				AccountID:    s.currentAccountID,
				DeckID:       deckID,
				EventID:      eventID,
				EventName:    eventName,
				Timestamp:    matchTime,
				PlayerWins:   playerWins,
				OpponentWins: opponentWins,
				PlayerTeamID: int(playerTeamID),
				Format:       eventID,
				Result:       matchResult,
				ResultReason: resultReason,
				OpponentName: opponentName,
				OpponentID:   opponentID,
				CreatedAt:    matchTime,
			}

			matches = append(matches, matchData{
				Match: match,
				Games: games,
			})
		}
	}

	return matches, nil
}

// normalizeResultReason normalizes MTGA result reason codes to readable descriptions.
func normalizeResultReason(reason string) string {
	// Map MTGA result codes to readable descriptions
	reasonMap := map[string]string{
		"Normal":             "normal",
		"Concede":            "concede",
		"Timeout":            "timeout",
		"Draw":               "draw",
		"Disconnect":         "disconnect",
		"ConnectionLost":     "disconnect",
		"OpponentConcede":    "opponent_concede",
		"OpponentTimeout":    "opponent_timeout",
		"OpponentDisconnect": "opponent_disconnect",
		"LifeReducedToZero":  "life_zero",
		"DeckEmpty":          "mill",
		"PoisonCounters":     "poison",
	}

	// Try exact match first
	if normalized, ok := reasonMap[reason]; ok {
		return normalized
	}

	// Try case-insensitive match
	reasonLower := strings.ToLower(reason)
	for key, value := range reasonMap {
		if strings.ToLower(key) == reasonLower {
			return value
		}
	}

	// Return lowercase version if no mapping found
	return strings.ToLower(reason)
}

// parseLogTimestamp attempts to parse a timestamp from the log entry format.
func parseLogTimestamp(timestampStr string) (time.Time, error) {
	// Format examples:
	// - [UnityCrossThreadLogger]11/16/2025 10:16:08 AM
	// - [UnityCrossThreadLogger]2024-01-15 10:30:45

	// Try to extract the date/time portion after the logger prefix
	parts := strings.Fields(timestampStr)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: not enough parts")
	}

	// Try common formats with different combinations of parts
	// For "11/16/2025 10:16:08 AM" we need 3 parts (date, time, AM/PM)
	// For "2024-01-15 10:30:45" we need 2 parts (date, time)

	formats := []struct {
		format   string
		numParts int
	}{
		// 12-hour format with AM/PM (MM/DD/YYYY)
		{"01/02/2006 03:04:05 PM", 3},
		{"1/2/2006 3:04:05 PM", 3},
		// 24-hour format (YYYY-MM-DD)
		{"2006-01-02 15:04:05", 2},
		{"2006-01-02T15:04:05", 2},
		{"2006-01-02 15:04:05.000", 2},
	}

	for _, fmt := range formats {
		if len(parts) < fmt.numParts {
			continue
		}

		// Build datetime string from last N parts
		var dateTimeStr string
		if fmt.numParts == 3 {
			dateTimeStr = parts[len(parts)-3] + " " + parts[len(parts)-2] + " " + parts[len(parts)-1]
		} else {
			dateTimeStr = parts[len(parts)-2] + " " + parts[len(parts)-1]
		}

		if t, err := time.Parse(fmt.format, dateTimeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp from: %s", timestampStr)
}

// StoreMatch stores a single match and its games.
// This is useful when processing match completion events from the log.
func (s *Service) StoreMatch(ctx context.Context, match *Match, games []*Game) error {
	// Use a transaction to ensure atomicity
	return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Create the match
		if err := s.matches.Create(ctx, match); err != nil {
			return fmt.Errorf("failed to create match: %w", err)
		}

		// Create each game
		for _, game := range games {
			if err := s.matches.CreateGame(ctx, game); err != nil {
				return fmt.Errorf("failed to create game: %w", err)
			}
		}

		return nil
	})
}

// BatchStoreMatches efficiently stores multiple matches in a single transaction.
// Uses INSERT OR IGNORE to skip duplicates without SELECT queries (10-20x faster than StoreMatch loop).
// Ideal for bulk imports like historical log replay.
func (s *Service) BatchStoreMatches(ctx context.Context, matchesData []matchData) (int, error) {
	if len(matchesData) == 0 {
		return 0, nil
	}

	db := s.db.Conn()
	stored := 0

	// Single transaction for all inserts
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error - will be nil if Commit() succeeds
	}()

	// Prepare match insert statement
	matchStmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO matches (
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id, rank_before, rank_after,
			format, result, result_reason, opponent_name, opponent_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare match statement: %w", err)
	}
	defer func() {
		_ = matchStmt.Close() // Explicitly ignore error - statement cleanup
	}()

	// Prepare game insert statement
	gameStmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO games (
			match_id, game_number, result, duration_seconds, result_reason, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare game statement: %w", err)
	}
	defer func() {
		_ = gameStmt.Close() // Explicitly ignore error - statement cleanup
	}()

	// Insert all matches and games
	for _, matchData := range matchesData {
		match := matchData.Match

		// Insert match
		result, err := matchStmt.ExecContext(ctx,
			match.ID,
			match.AccountID,
			match.EventID,
			match.EventName,
			match.Timestamp,
			match.DurationSeconds,
			match.PlayerWins,
			match.OpponentWins,
			match.PlayerTeamID,
			match.DeckID,
			match.RankBefore,
			match.RankAfter,
			match.Format,
			match.Result,
			match.ResultReason,
			match.OpponentName,
			match.OpponentID,
			match.CreatedAt,
		)
		if err != nil {
			return stored, fmt.Errorf("failed to insert match %s: %w", match.ID, err)
		}

		// Check if match was actually inserted (not ignored due to duplicate)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			stored++

			// Insert games for this match
			for _, game := range matchData.Games {
				_, err := gameStmt.ExecContext(ctx,
					game.MatchID,
					game.GameNumber,
					game.Result,
					game.DurationSeconds,
					game.ResultReason,
					game.CreatedAt,
				)
				if err != nil {
					return stored, fmt.Errorf("failed to insert game for match %s: %w", game.MatchID, err)
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return stored, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return stored, nil
}

// GetStats retrieves statistics with optional filtering.
func (s *Service) GetStats(ctx context.Context, filter StatsFilter) (*Statistics, error) {
	return s.matches.GetStats(ctx, filter)
}

// GetRecentMatches retrieves matches within a date range for the current account.
func (s *Service) GetRecentMatches(ctx context.Context, days int) ([]*Match, error) {
	end := time.Now()
	start := end.Add(-time.Duration(days) * 24 * time.Hour)
	return s.matches.GetByDateRange(ctx, start, end, s.currentAccountID)
}

// GetMatches retrieves matches based on the given filter with advanced filtering support.
// This method now supports multiple formats, rank filters, opponent filters, and more.
func (s *Service) GetMatches(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error) {
	// Use account filter if specified, otherwise use current account
	if filter.AccountID == nil {
		accountID := s.currentAccountID
		filter.AccountID = &accountID
	} else if *filter.AccountID == 0 {
		// 0 means all accounts - keep as nil
		filter.AccountID = nil
	}

	// Use the repository's GetMatches method which supports advanced filtering
	matches, err := s.matches.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	return matches, nil
}

// GetRecentMatchesLimit retrieves the N most recent matches.
// If accountID is 0, returns matches for all accounts.
func (s *Service) GetRecentMatchesLimit(ctx context.Context, limit int) ([]*models.Match, error) {
	// Use current account ID or 0 for all accounts
	accountID := s.currentAccountID
	if accountID == 0 {
		// Already 0, show all accounts
		accountID = 0
	}
	return s.matches.GetRecentMatches(ctx, limit, accountID)
}

// GetLatestMatch retrieves the most recent match.
func (s *Service) GetLatestMatch(ctx context.Context) (*models.Match, error) {
	return s.matches.GetLatestMatch(ctx, s.currentAccountID)
}

// GetMatchByID retrieves a match by its ID.
func (s *Service) GetMatchByID(ctx context.Context, matchID string) (*models.Match, error) {
	return s.matches.GetByID(ctx, matchID)
}

// GetGamesForMatch retrieves all games for a specific match.
func (s *Service) GetGamesForMatch(ctx context.Context, matchID string) ([]*models.Game, error) {
	return s.matches.GetGamesForMatch(ctx, matchID)
}

// GetStatsByFormat retrieves statistics grouped by format.
func (s *Service) GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	// Use account filter if specified, otherwise use current account
	if filter.AccountID == nil {
		accountID := s.currentAccountID
		filter.AccountID = &accountID
	}

	return s.matches.GetStatsByFormat(ctx, filter)
}

// GetStatsByDeck retrieves statistics grouped by deck.
func (s *Service) GetStatsByDeck(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	// Use account filter if specified, otherwise use current account
	if filter.AccountID == nil {
		accountID := s.currentAccountID
		filter.AccountID = &accountID
	}

	return s.matches.GetStatsByDeck(ctx, filter)
}

// GetPerformanceMetrics retrieves duration-based performance metrics.
func (s *Service) GetPerformanceMetrics(ctx context.Context, filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	// Use account filter if specified, otherwise use current account
	if filter.AccountID == nil {
		accountID := s.currentAccountID
		filter.AccountID = &accountID
	}

	return s.matches.GetPerformanceMetrics(ctx, filter)
}

// GetStreakStats calculates win/loss streak statistics.
func (s *Service) GetStreakStats(ctx context.Context, filter models.StatsFilter) (*models.StreakStats, error) {
	// Get matches ordered by timestamp (oldest to newest) for accurate streak calculation
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for streak calculation: %w", err)
	}

	return stats.CalculateStreaks(matches), nil
}

// StoreDeckFromParser stores a deck parsed from log reader.
func (s *Service) StoreDeckFromParser(ctx context.Context, deckID, name, format, description string, created, modified time.Time, lastPlayed *time.Time, mainDeck, sideboard []struct {
	CardID   int
	Quantity int
},
) error {
	// Ensure we have a valid account ID (fix for issue #618)
	// If currentAccountID is 0, use default account ID 1
	accountID := s.currentAccountID
	if accountID == 0 {
		accountID = 1
		log.Printf("[StoreDeck] WARNING: currentAccountID is 0, using default account ID 1 for deck '%s'", name)
	}

	// Convert to storage models
	// Source "arena" indicates this deck was synced from MTGA (parsed from logs)
	deck := &models.Deck{
		ID:         deckID,
		AccountID:  accountID,
		Name:       name,
		Format:     format,
		Source:     "arena",
		CreatedAt:  created,
		ModifiedAt: modified,
		LastPlayed: lastPlayed,
	}

	if lastPlayed != nil {
		log.Printf("[StoreDeck] Deck '%s' has LastPlayed: %v", name, *lastPlayed)
	} else {
		log.Printf("[StoreDeck] Deck '%s' has NO LastPlayed timestamp", name)
	}

	if description != "" {
		deck.Description = &description
	}

	// Convert cards
	var cards []*models.DeckCard
	for _, card := range mainDeck {
		cards = append(cards, &models.DeckCard{
			DeckID:   deckID,
			CardID:   card.CardID,
			Quantity: card.Quantity,
			Board:    "main",
		})
	}
	for _, card := range sideboard {
		cards = append(cards, &models.DeckCard{
			DeckID:   deckID,
			CardID:   card.CardID,
			Quantity: card.Quantity,
			Board:    "sideboard",
		})
	}

	return s.StoreDeck(ctx, deck, cards)
}

// StoreDeck stores a complete deck with its cards.
func (s *Service) StoreDeck(ctx context.Context, deck *models.Deck, cards []*models.DeckCard) error {
	return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Create or update the deck
		existing, err := s.decks.GetByID(ctx, deck.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing deck: %w", err)
		}

		if existing == nil {
			if err := s.decks.Create(ctx, deck); err != nil {
				return fmt.Errorf("failed to create deck: %w", err)
			}
		} else {
			if err := s.decks.Update(ctx, deck); err != nil {
				return fmt.Errorf("failed to update deck: %w", err)
			}
		}

		// Clear existing cards and add new ones
		if err := s.decks.ClearCards(ctx, deck.ID); err != nil {
			return fmt.Errorf("failed to clear deck cards: %w", err)
		}

		for _, card := range cards {
			if err := s.decks.AddCard(ctx, card); err != nil {
				return fmt.Errorf("failed to add card to deck: %w", err)
			}
		}

		return nil
	})
}

// ListDecks returns all decks for the current account.
func (s *Service) ListDecks(ctx context.Context) ([]*models.Deck, error) {
	return s.decks.List(ctx, s.currentAccountID)
}

// CleanupStaleArenaDecks removes arena-sourced decks that are no longer present in MTGA logs.
// It deletes all "arena" source decks for the current account EXCEPT those in the provided ID list.
// Returns the number of decks deleted.
func (s *Service) CleanupStaleArenaDecks(ctx context.Context, currentDeckIDs []string) (int, error) {
	accountID := s.currentAccountID
	if accountID == 0 {
		accountID = 1
	}
	return s.decks.DeleteBySourceExcluding(ctx, accountID, "arena", currentDeckIDs)
}

// InferDeckIDsForMatches attempts to link matches to decks based on timestamp proximity.
// This is a best-effort approach since MTGA logs don't include deck IDs in match events.
// It links each match without a deck_id to the deck with the closest lastPlayed timestamp.
func (s *Service) InferDeckIDsForMatches(ctx context.Context) (int, error) {
	// Get all matches without deck IDs
	matchesNeedingDecks, err := s.matches.GetMatchesWithoutDeckID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get matches without deck IDs: %w", err)
	}

	log.Printf("[InferDeckIDs] Found %d matches without deck IDs", len(matchesNeedingDecks))

	if len(matchesNeedingDecks) == 0 {
		return 0, nil
	}

	// Get all decks with LastPlayed timestamps for the current account
	allDecks, err := s.decks.List(ctx, s.currentAccountID)
	if err != nil {
		return 0, fmt.Errorf("failed to list decks: %w", err)
	}

	log.Printf("[InferDeckIDs] Found %d total decks in database", len(allDecks))

	// Filter to decks that have LastPlayed timestamp
	var decksWithTimestamp []*models.Deck
	for _, deck := range allDecks {
		if deck.LastPlayed != nil {
			decksWithTimestamp = append(decksWithTimestamp, deck)
		}
	}

	log.Printf("[InferDeckIDs] Found %d decks with LastPlayed timestamps", len(decksWithTimestamp))

	if len(decksWithTimestamp) == 0 {
		log.Printf("[InferDeckIDs] No decks have LastPlayed timestamps - cannot infer deck IDs")
		return 0, nil
	}

	updatedCount := 0
	const maxTimeDiff = 24 * time.Hour // Only link if within 24 hours (same play session day)

	// Sort decks by LastPlayed descending (most recent first)
	// This helps when match timestamps are missing and we fall back to time.Now()
	sort.Slice(decksWithTimestamp, func(i, j int) bool {
		return decksWithTimestamp[i].LastPlayed.After(*decksWithTimestamp[j].LastPlayed)
	})

	// Check if all match timestamps are suspiciously close together (within 1 minute)
	// This indicates they're using time.Now() fallback during batch replay
	suspiciousBatch := false
	if len(matchesNeedingDecks) > 1 {
		timeDiffBetweenMatches := matchesNeedingDecks[len(matchesNeedingDecks)-1].Timestamp.Sub(matchesNeedingDecks[0].Timestamp)
		// Take absolute value to handle both chronological and reverse-chronological ordering
		if timeDiffBetweenMatches < 0 {
			timeDiffBetweenMatches = -timeDiffBetweenMatches
		}
		if timeDiffBetweenMatches < 1*time.Minute {
			suspiciousBatch = true
			log.Printf("[InferDeckIDs] WARNING: All %d matches have timestamps within %v - likely using time.Now() fallback", len(matchesNeedingDecks), timeDiffBetweenMatches)
			log.Printf("[InferDeckIDs] Will link all matches to most recently played deck: '%s'", decksWithTimestamp[0].Name)
		}
	}

	log.Printf("[InferDeckIDs] Starting to match %d matches with %d decks (max time diff: %v)", len(matchesNeedingDecks), len(decksWithTimestamp), maxTimeDiff)
	log.Printf("[InferDeckIDs] Most recent deck: '%s' (LastPlayed: %v)", decksWithTimestamp[0].Name, *decksWithTimestamp[0].LastPlayed)

	for i, match := range matchesNeedingDecks {
		var bestDeck *models.Deck
		var minDiff time.Duration

		if i < 3 { // Log first 3 matches for debugging
			log.Printf("[InferDeckIDs] Match %d: timestamp=%v", i+1, match.Timestamp)
		}

		for j, deck := range decksWithTimestamp {
			// Calculate time difference
			diff := match.Timestamp.Sub(*deck.LastPlayed)
			if diff < 0 {
				diff = -diff
			}

			if i < 3 && j < 3 { // Log first few comparisons
				log.Printf("[InferDeckIDs]   Deck '%s': LastPlayed=%v, diff=%v", deck.Name, *deck.LastPlayed, diff)
			}

			// Check if this is the closest deck so far
			if bestDeck == nil || diff < minDiff {
				bestDeck = deck
				minDiff = diff
			}
		}

		// Determine if we should link this match to the best deck
		shouldLink := false
		var linkReason string

		if bestDeck != nil && minDiff <= maxTimeDiff {
			shouldLink = true
			linkReason = fmt.Sprintf("within time window (diff: %v)", minDiff)
		} else if suspiciousBatch && bestDeck != nil {
			// Fallback for batch replay: link to most recent deck
			shouldLink = true
			linkReason = "suspicious batch - using most recent deck"
		}

		if shouldLink && bestDeck != nil {
			if err := s.matches.UpdateDeckID(ctx, match.ID, bestDeck.ID); err != nil {
				return updatedCount, fmt.Errorf("failed to update match %s with deck ID: %w", match.ID, err)
			}
			updatedCount++
			if updatedCount <= 3 { // Log first few successful links
				log.Printf("[InferDeckIDs] ✓ Linked match %s to deck '%s' (%s)", match.ID, bestDeck.Name, linkReason)
			}
		} else {
			if i < 3 { // Log first few failures
				if bestDeck != nil {
					log.Printf("[InferDeckIDs] ✗ Match %s too far from best deck '%s' (diff: %v > max: %v)", match.ID, bestDeck.Name, minDiff, maxTimeDiff)
				} else {
					log.Printf("[InferDeckIDs] ✗ Match %s has no best deck", match.ID)
				}
			}
		}
	}

	log.Printf("[InferDeckIDs] Completed: linked %d/%d matches to decks", updatedCount, len(matchesNeedingDecks))

	return updatedCount, nil
}

// UpdateCollection updates the collection based on changes detected in the log.
// This would typically be called when processing inventory updates.
func (s *Service) UpdateCollection(ctx context.Context, cardID int, newQuantity int, source string) error {
	// Get current quantity
	currentQty, err := s.collection.GetCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get current quantity: %w", err)
	}

	// Calculate delta
	delta := newQuantity - currentQty

	if delta != 0 {
		// Record the change
		if err := s.collection.RecordChange(ctx, cardID, delta, time.Now(), &source); err != nil {
			return fmt.Errorf("failed to record collection change: %w", err)
		}
	}

	return nil
}

// GetCollection retrieves the entire collection.
func (s *Service) GetCollection(ctx context.Context) (map[int]int, error) {
	return s.collection.GetAll(ctx)
}

// GetRecentCollectionChanges retrieves recent changes to the collection.
func (s *Service) GetRecentCollectionChanges(ctx context.Context, limit int) ([]*CollectionHistory, error) {
	return s.collection.GetRecentChanges(ctx, limit)
}

// GetSetCompletion calculates set completion percentages.
// Uses SetCardRepository for data access and CollectionRepository for owned cards.
func (s *Service) GetSetCompletion(ctx context.Context) ([]*models.SetCompletion, error) {
	// Get set/rarity card counts from repository
	rarityCounts, err := s.setCard.GetSetRarityCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get set rarity counts: %w", err)
	}

	// Build set data structure
	setData := make(map[string]*models.SetCompletion)
	for _, rc := range rarityCounts {
		// Initialize set if not exists
		if _, exists := setData[rc.SetCode]; !exists {
			setData[rc.SetCode] = &models.SetCompletion{
				SetCode:         rc.SetCode,
				SetName:         rc.SetName,
				RarityBreakdown: make(map[string]*models.RarityCompletion),
			}
		}

		// Add rarity breakdown
		setData[rc.SetCode].RarityBreakdown[rc.Rarity] = &models.RarityCompletion{
			Rarity: rc.Rarity,
			Total:  rc.Total,
		}
		setData[rc.SetCode].TotalCards += rc.Total
	}

	// Get owned cards from collection
	collection, err := s.collection.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Get all card set info from repository
	cardInfos, err := s.setCard.GetAllCardSetInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get card set info: %w", err)
	}

	// Count owned cards per set and rarity
	for _, card := range cardInfos {
		// Convert arena_id from string to int for collection lookup
		var arenaID int
		if _, err := fmt.Sscanf(card.ArenaID, "%d", &arenaID); err != nil {
			continue // Skip cards with invalid arena IDs
		}

		// Check if player owns this card
		if quantity, owned := collection[arenaID]; owned && quantity > 0 {
			if set, exists := setData[card.SetCode]; exists {
				set.OwnedCards++
				if rarityData, rarityExists := set.RarityBreakdown[card.Rarity]; rarityExists {
					rarityData.Owned++
				}
			}
		}
	}

	// Calculate percentages
	result := make([]*models.SetCompletion, 0, len(setData))
	for _, set := range setData {
		if set.TotalCards > 0 {
			set.Percentage = float64(set.OwnedCards) / float64(set.TotalCards) * 100
		}

		// Calculate rarity percentages
		for _, rarity := range set.RarityBreakdown {
			if rarity.Total > 0 {
				rarity.Percentage = float64(rarity.Owned) / float64(rarity.Total) * 100
			}
		}

		result = append(result, set)
	}

	return result, nil
}

// Account Management Methods

// GetCurrentAccount returns the currently active account.
func (s *Service) GetCurrentAccount(ctx context.Context) (*models.Account, error) {
	return s.accounts.GetByID(ctx, s.currentAccountID)
}

// SetCurrentAccount sets the currently active account.
func (s *Service) SetCurrentAccount(ctx context.Context, accountID int) error {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account not found: %d", accountID)
	}
	s.currentAccountID = accountID
	return nil
}

// GetCurrentAccountID returns the ID of the currently active account.
func (s *Service) GetCurrentAccountID() int {
	return s.currentAccountID
}

// CreateAccount creates a new account.
func (s *Service) CreateAccount(ctx context.Context, name string, screenName, clientID *string) (*models.Account, error) {
	now := time.Now()
	account := &models.Account{
		Name:       name,
		ScreenName: screenName,
		ClientID:   clientID,
		IsDefault:  false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.accounts.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return account, nil
}

// GetAllAccounts retrieves all accounts.
func (s *Service) GetAllAccounts(ctx context.Context) ([]*models.Account, error) {
	return s.accounts.GetAll(ctx)
}

// GetAccount retrieves an account by ID.
func (s *Service) GetAccount(ctx context.Context, id int) (*models.Account, error) {
	return s.accounts.GetByID(ctx, id)
}

// UpdateAccount updates an account.
func (s *Service) UpdateAccount(ctx context.Context, account *models.Account) error {
	account.UpdatedAt = time.Now()
	return s.accounts.Update(ctx, account)
}

// SetDefaultAccount sets an account as the default account.
func (s *Service) SetDefaultAccount(ctx context.Context, accountID int) error {
	if err := s.accounts.SetDefault(ctx, accountID); err != nil {
		return fmt.Errorf("failed to set default account: %w", err)
	}
	// Also set as current account
	s.currentAccountID = accountID
	return nil
}

// DeleteAccount deletes an account.
func (s *Service) DeleteAccount(ctx context.Context, id int) error {
	// Don't allow deleting the current account
	if id == s.currentAccountID {
		return fmt.Errorf("cannot delete the currently active account")
	}
	return s.accounts.Delete(ctx, id)
}

// GetCombinedStatistics retrieves statistics across all accounts or a specific account.
func (s *Service) GetCombinedStatistics(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error) {
	// If AccountID is nil or 0, get combined stats for all accounts
	if filter.AccountID == nil || *filter.AccountID == 0 {
		// Set AccountID to 0 to get all accounts
		allAccounts := 0
		filter.AccountID = &allAccounts
	}
	return s.matches.GetStats(ctx, filter)
}

// Rank History Methods

// StoreRankSnapshot stores a rank snapshot in the database.
func (s *Service) StoreRankSnapshot(ctx context.Context, rank *models.RankHistory) error {
	rank.AccountID = s.currentAccountID
	rank.CreatedAt = time.Now()
	return s.rankHistory.Create(ctx, rank)
}

// StoreRankUpdate converts and stores a rank update from the log parser.
func (s *Service) StoreRankUpdate(ctx context.Context, update *logreader.RankUpdate) error {
	if update == nil {
		return nil
	}

	// Convert to models.RankHistory
	rank := &models.RankHistory{
		Timestamp:     update.Timestamp,
		Format:        update.FormatToDBFormat(), // "Constructed" -> "constructed", "Limited" -> "limited"
		SeasonOrdinal: update.SeasonOrdinal,
	}

	// Set rank class (nullable)
	if update.NewClass != "" {
		rank.RankClass = &update.NewClass
	}

	// Set rank level (nullable)
	if update.NewLevel > 0 {
		rank.RankLevel = &update.NewLevel
	}

	// Set rank step (nullable)
	if update.NewStep > 0 {
		rank.RankStep = &update.NewStep
	}

	// Percentile is not available in RankUpdated events, leave as nil

	return s.StoreRankSnapshot(ctx, rank)
}

// GetRankHistoryByFormat retrieves all rank history entries for a specific format.
func (s *Service) GetRankHistoryByFormat(ctx context.Context, format string) ([]*models.RankHistory, error) {
	return s.rankHistory.GetByFormat(ctx, s.currentAccountID, format)
}

// GetRankHistoryBySeason retrieves all rank history entries for a specific season.
func (s *Service) GetRankHistoryBySeason(ctx context.Context, seasonOrdinal int) ([]*models.RankHistory, error) {
	return s.rankHistory.GetBySeason(ctx, s.currentAccountID, seasonOrdinal)
}

// GetRankHistoryByDateRange retrieves rank history entries within a date range.
func (s *Service) GetRankHistoryByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.RankHistory, error) {
	return s.rankHistory.GetByDateRange(ctx, s.currentAccountID, startDate, endDate)
}

// GetLatestRankByFormat retrieves the most recent rank snapshot for a format.
func (s *Service) GetLatestRankByFormat(ctx context.Context, format string) (*models.RankHistory, error) {
	return s.rankHistory.GetLatestByFormat(ctx, s.currentAccountID, format)
}

// GetAllRankHistory retrieves all rank history entries.
func (s *Service) GetAllRankHistory(ctx context.Context) ([]*models.RankHistory, error) {
	return s.rankHistory.GetAll(ctx, s.currentAccountID)
}

// Seasonal Rank Progression Methods

// GetSeasonalRankSummary retrieves rank summary for each season for a specific format.
func (s *Service) GetSeasonalRankSummary(ctx context.Context, format string) ([]*models.SeasonalRankSummary, error) {
	// Get all rank history for the format
	history, err := s.GetRankHistoryByFormat(ctx, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get rank history: %w", err)
	}

	if len(history) == 0 {
		return nil, nil
	}

	// Group by season
	seasonMap := make(map[int][]*models.RankHistory)
	for _, rank := range history {
		seasonMap[rank.SeasonOrdinal] = append(seasonMap[rank.SeasonOrdinal], rank)
	}

	// Build summaries
	summaries := make([]*models.SeasonalRankSummary, 0, len(seasonMap))
	for seasonOrdinal, ranks := range seasonMap {
		// Sort ranks by timestamp
		sortByTimestamp := func(ranks []*models.RankHistory) {
			for i := 0; i < len(ranks); i++ {
				for j := i + 1; j < len(ranks); j++ {
					if ranks[i].Timestamp.After(ranks[j].Timestamp) {
						ranks[i], ranks[j] = ranks[j], ranks[i]
					}
				}
			}
		}
		sortByTimestamp(ranks)

		summary := &models.SeasonalRankSummary{
			SeasonOrdinal:  seasonOrdinal,
			Format:         format,
			TotalSnapshots: len(ranks),
			FirstSeen:      ranks[0].Timestamp,
			LastSeen:       ranks[len(ranks)-1].Timestamp,
		}

		// Set start and end ranks
		if ranks[0].RankClass != nil {
			startRank := formatRankHistoryString(ranks[0])
			summary.StartRank = &startRank
		}
		if ranks[len(ranks)-1].RankClass != nil {
			endRank := formatRankHistoryString(ranks[len(ranks)-1])
			summary.EndRank = &endRank
		}

		// Find highest and lowest ranks
		var highest, lowest *models.RankHistory
		for _, rank := range ranks {
			if rank.RankClass == nil {
				continue
			}
			if highest == nil || compareRanks(rank, highest) > 0 {
				highest = rank
			}
			if lowest == nil || compareRanks(rank, lowest) < 0 {
				lowest = rank
			}
		}

		if highest != nil {
			highestStr := formatRankHistoryString(highest)
			summary.HighestRank = &highestStr
		}
		if lowest != nil {
			lowestStr := formatRankHistoryString(lowest)
			summary.LowestRank = &lowestStr
		}

		summaries = append(summaries, summary)
	}

	// Sort summaries by season (most recent first)
	for i := 0; i < len(summaries); i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[i].SeasonOrdinal < summaries[j].SeasonOrdinal {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	return summaries, nil
}

// GetRankAchievements retrieves all rank achievements (first time reaching each rank).
func (s *Service) GetRankAchievements(ctx context.Context, format string) ([]*models.RankAchievement, error) {
	// Get all rank history for the format, sorted by timestamp
	history, err := s.GetRankHistoryByFormat(ctx, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get rank history: %w", err)
	}

	if len(history) == 0 {
		return nil, nil
	}

	// Sort by timestamp (oldest first)
	for i := 0; i < len(history); i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].Timestamp.After(history[j].Timestamp) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	// Track first occurrence of each rank class
	achievements := make(map[string]*models.RankAchievement)
	var highestRank *models.RankHistory

	for _, rank := range history {
		if rank.RankClass == nil || *rank.RankClass == "" {
			continue
		}

		rankKey := *rank.RankClass
		if rank.RankLevel != nil {
			rankKey = fmt.Sprintf("%s %d", *rank.RankClass, *rank.RankLevel)
		}

		// Track first achievement of this rank
		if _, exists := achievements[rankKey]; !exists {
			achievements[rankKey] = &models.RankAchievement{
				Format:        format,
				RankClass:     *rank.RankClass,
				RankLevel:     rank.RankLevel,
				FirstAchieved: rank.Timestamp,
				SeasonOrdinal: rank.SeasonOrdinal,
				IsHighest:     false,
			}
		}

		// Track highest rank
		if highestRank == nil || compareRanks(rank, highestRank) > 0 {
			highestRank = rank
		}
	}

	// Mark the highest rank achievement
	if highestRank != nil && highestRank.RankClass != nil {
		highestKey := *highestRank.RankClass
		if highestRank.RankLevel != nil {
			highestKey = fmt.Sprintf("%s %d", *highestRank.RankClass, *highestRank.RankLevel)
		}
		if achievement, exists := achievements[highestKey]; exists {
			achievement.IsHighest = true
		}
	}

	// Convert map to slice and sort by achievement date
	result := make([]*models.RankAchievement, 0, len(achievements))
	for _, achievement := range achievements {
		result = append(result, achievement)
	}

	// Sort by first achieved date (oldest first)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].FirstAchieved.After(result[j].FirstAchieved) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// Rank Progression Analysis Methods

// GetRankProgression calculates progress toward next rank tier.
func (s *Service) GetRankProgression(ctx context.Context, format string) (*models.RankProgression, error) {
	// Get latest rank for format
	latestRank, err := s.GetLatestRankByFormat(ctx, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest rank: %w", err)
	}

	if latestRank == nil || latestRank.RankClass == nil {
		return nil, nil
	}

	// Parse current rank
	currentRank := *latestRank.RankClass
	currentLevel := 1
	if latestRank.RankLevel != nil {
		currentLevel = *latestRank.RankLevel
	}
	currentStep := 0
	if latestRank.RankStep != nil {
		currentStep = *latestRank.RankStep
	}

	// Determine next rank
	nextRank := calculateNextRank(currentRank, currentLevel)

	// Calculate steps to next tier (assuming 6 steps per tier in most ranks)
	stepsPerTier := 6
	if currentRank == "Mythic" {
		stepsPerTier = 0 // Mythic has no tiers
	}
	stepsToNext := stepsPerTier - currentStep

	// Check if at floor
	isAtFloor := isRankFloor(currentRank, currentLevel)

	// Calculate estimated matches based on recent win rate
	filter := models.StatsFilter{
		Format: &format,
	}
	stats, err := s.GetStats(ctx, filter)
	if err == nil && stats != nil && stats.TotalMatches > 0 {
		winRate := stats.WinRate
		// Estimate matches needed: steps / (win rate - 0.5)
		// Assumes you gain 1 step per win and lose 1 per loss
		// Net gain per match = (winRate * 1) + ((1-winRate) * -1) = 2*winRate - 1
		netGainPerMatch := 2*winRate - 1
		if netGainPerMatch > 0 {
			estimated := int(float64(stepsToNext) / netGainPerMatch)
			return &models.RankProgression{
				CurrentRank:      formatRank(currentRank, currentLevel, currentStep),
				NextRank:         nextRank,
				CurrentStep:      currentStep,
				StepsToNext:      stepsToNext,
				IsAtFloor:        isAtFloor,
				EstimatedMatches: &estimated,
				WinRateUsed:      &winRate,
				Format:           format,
				LastUpdated:      time.Now(),
			}, nil
		}
	}

	return &models.RankProgression{
		CurrentRank: formatRank(currentRank, currentLevel, currentStep),
		NextRank:    nextRank,
		CurrentStep: currentStep,
		StepsToNext: stepsToNext,
		IsAtFloor:   isAtFloor,
		Format:      format,
		LastUpdated: time.Now(),
	}, nil
}

// DetectDoubleRankUps detects all double rank up events in match history.
func (s *Service) DetectDoubleRankUps(ctx context.Context, format string) ([]*models.DoubleRankUp, error) {
	// Get all matches with rank changes for the format
	filter := models.StatsFilter{
		Format: &format,
	}
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	var doubleRankUps []*models.DoubleRankUp

	for _, match := range matches {
		if match.RankBefore == nil || match.RankAfter == nil {
			continue
		}

		// Parse ranks
		beforeClass, beforeLevel := parseRankString(*match.RankBefore)
		afterClass, afterLevel := parseRankString(*match.RankAfter)

		if beforeClass == "" || afterClass == "" {
			continue
		}

		// Check if rank class jumped
		if beforeClass != afterClass {
			// Check if it's a double rank up (skipped a class)
			classOrder := []string{"Bronze", "Silver", "Gold", "Platinum", "Diamond", "Mythic"}
			beforeIdx := -1
			afterIdx := -1
			for i, class := range classOrder {
				if class == beforeClass {
					beforeIdx = i
				}
				if class == afterClass {
					afterIdx = i
				}
			}

			if afterIdx > beforeIdx+1 {
				// Skipped at least one class
				skippedClass := classOrder[beforeIdx+1]
				doubleRankUps = append(doubleRankUps, &models.DoubleRankUp{
					PreviousRank:  *match.RankBefore,
					NewRank:       *match.RankAfter,
					SkippedRank:   skippedClass + " (entire tier)",
					MatchID:       match.ID,
					Timestamp:     match.Timestamp,
					Format:        format,
					SeasonOrdinal: 0, // Would need to get from rank history
				})
			}
		} else if beforeLevel > 0 && afterLevel > 0 {
			// Same class, check if level jumped
			if beforeLevel-afterLevel > 1 {
				// Skipped a level
				skippedLevel := beforeLevel - 1
				doubleRankUps = append(doubleRankUps, &models.DoubleRankUp{
					PreviousRank:  *match.RankBefore,
					NewRank:       *match.RankAfter,
					SkippedRank:   fmt.Sprintf("%s %d", beforeClass, skippedLevel),
					MatchID:       match.ID,
					Timestamp:     match.Timestamp,
					Format:        format,
					SeasonOrdinal: 0,
				})
			}
		}
	}

	return doubleRankUps, nil
}

// GetRankFloors returns all rank floors for a format.
func (s *Service) GetRankFloors(format string) []*models.RankFloor {
	floors := []*models.RankFloor{
		{RankClass: "Bronze", RankLevel: 4, Format: format},
		{RankClass: "Silver", RankLevel: 4, Format: format},
		{RankClass: "Gold", RankLevel: 4, Format: format},
		{RankClass: "Platinum", RankLevel: 4, Format: format},
		{RankClass: "Diamond", RankLevel: 4, Format: format},
	}
	return floors
}

// Helper functions for seasonal and rank analysis

// formatRankHistoryString formats a rank history entry as a string.
func formatRankHistoryString(rank *models.RankHistory) string {
	if rank.RankClass == nil {
		return "Unranked"
	}
	result := *rank.RankClass
	if rank.RankLevel != nil {
		result = fmt.Sprintf("%s %d", result, *rank.RankLevel)
	}
	if rank.RankStep != nil {
		result = fmt.Sprintf("%s (Step %d)", result, *rank.RankStep)
	}
	return result
}

// compareRanks compares two ranks. Returns > 0 if a is higher, < 0 if b is higher, 0 if equal.
func compareRanks(a, b *models.RankHistory) int {
	if a.RankClass == nil || b.RankClass == nil {
		return 0
	}

	// Rank class order (higher is better)
	rankOrder := map[string]int{
		"Bronze":   1,
		"Silver":   2,
		"Gold":     3,
		"Platinum": 4,
		"Diamond":  5,
		"Mythic":   6,
	}

	aOrder := rankOrder[*a.RankClass]
	bOrder := rankOrder[*b.RankClass]

	if aOrder != bOrder {
		return aOrder - bOrder
	}

	// Same class, compare level (lower number is higher rank)
	if a.RankLevel != nil && b.RankLevel != nil {
		if *a.RankLevel != *b.RankLevel {
			return *b.RankLevel - *a.RankLevel
		}
	}

	// Same level, compare step (higher step is better)
	if a.RankStep != nil && b.RankStep != nil {
		return *a.RankStep - *b.RankStep
	}

	return 0
}

// Helper functions for rank analysis

func calculateNextRank(rankClass string, level int) string {
	if rankClass == "Mythic" {
		return "Mythic (Top Rank)"
	}

	if level > 1 {
		return fmt.Sprintf("%s %d", rankClass, level-1)
	}

	// Next class
	classOrder := map[string]string{
		"Bronze":   "Silver",
		"Silver":   "Gold",
		"Gold":     "Platinum",
		"Platinum": "Diamond",
		"Diamond":  "Mythic",
	}

	if nextClass, exists := classOrder[rankClass]; exists {
		if nextClass == "Mythic" {
			return "Mythic"
		}
		return fmt.Sprintf("%s 4", nextClass)
	}

	return "Unknown"
}

func isRankFloor(rankClass string, level int) bool {
	return level == 4
}

func formatRank(rankClass string, level, step int) string {
	if rankClass == "Mythic" {
		return "Mythic"
	}
	if level > 0 {
		return fmt.Sprintf("%s %d (Step %d)", rankClass, level, step)
	}
	return fmt.Sprintf("%s (Step %d)", rankClass, step)
}

func parseRankString(rankStr string) (class string, level int) {
	// Parse strings like "Gold 2" or "Mythic"
	parts := strings.Split(rankStr, " ")
	if len(parts) == 0 {
		return "", 0
	}

	class = parts[0]
	if len(parts) > 1 {
		_, _ = fmt.Sscanf(parts[1], "%d", &level)
	}
	return class, level
}

// Bulk Data Update Methods

// GetLastBulkDataUpdate retrieves the timestamp of the last bulk data update.
func (s *Service) GetLastBulkDataUpdate(ctx context.Context) (time.Time, error) {
	conn := s.db.Conn()

	query := `SELECT value FROM metadata WHERE key = 'bulk_data_last_update' LIMIT 1`
	var timestampStr string
	err := conn.QueryRowContext(ctx, query).Scan(&timestampStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// No update has occurred yet - return zero time
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to get last bulk data update: %w", err)
	}

	// Parse the timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse bulk data timestamp: %w", err)
	}

	return timestamp, nil
}

// SetLastBulkDataUpdate records the timestamp of the last bulk data update.
func (s *Service) SetLastBulkDataUpdate(ctx context.Context, timestamp time.Time) error {
	conn := s.db.Conn()

	timestampStr := timestamp.Format(time.RFC3339)
	query := `
		INSERT INTO metadata (key, value, updated_at)
		VALUES ('bulk_data_last_update', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := conn.ExecContext(ctx, query, timestampStr)
	if err != nil {
		return fmt.Errorf("failed to set last bulk data update: %w", err)
	}

	return nil
}

// Quests returns the quest repository for accessing quest data.
func (s *Service) Quests() *QuestRepository {
	return s.quests
}

// DraftRepo returns the draft repository.
func (s *Service) DraftRepo() repository.DraftRepository {
	return s.draft
}

// SetCardRepo returns the set card repository.
func (s *Service) SetCardRepo() repository.SetCardRepository {
	return s.setCard
}

// DraftRatingsRepo returns the draft ratings repository.
func (s *Service) DraftRatingsRepo() repository.DraftRatingsRepository {
	return s.draftRatings
}

// CollectionRepo returns the collection repository.
func (s *Service) CollectionRepo() repository.CollectionRepository {
	return s.collection
}

// DeckRepo returns the deck repository.
func (s *Service) DeckRepo() repository.DeckRepository {
	return s.decks
}

// InventoryRepo returns the inventory repository.
func (s *Service) InventoryRepo() repository.InventoryRepository {
	return s.inventory
}

// RankHistoryRepo returns the rank history repository.
func (s *Service) RankHistoryRepo() repository.RankHistoryRepository {
	return s.rankHistory
}

// SettingsRepo returns the settings repository.
func (s *Service) SettingsRepo() repository.SettingsRepository {
	return s.settings
}

// MatchRepo returns the match repository.
func (s *Service) MatchRepo() repository.MatchRepository {
	return s.matches
}

// StandardRepo returns the standard repository.
func (s *Service) StandardRepo() repository.StandardRepository {
	return s.standard
}

// GamePlayRepo returns the game play repository.
func (s *Service) GamePlayRepo() repository.GamePlayRepository {
	return s.gamePlay
}

// CardPerformanceAnalysisRepo returns the card performance analysis repository.
func (s *Service) CardPerformanceAnalysisRepo() repository.CardPerformanceRepository {
	return s.cardPerformanceAnalysis
}

// GetCardNames retrieves card names for multiple arena IDs.
func (s *Service) GetCardNames(ctx context.Context, arenaIDs []string) (map[string]string, error) {
	if len(arenaIDs) == 0 {
		return make(map[string]string), nil
	}

	// Build placeholders for IN clause
	placeholders := ""
	args := make([]interface{}, len(arenaIDs))
	for i, id := range arenaIDs {
		args[i] = id
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	query := fmt.Sprintf(`
		SELECT arena_id, name
		FROM set_cards
		WHERE arena_id IN (%s)
	`, placeholders)

	rows, err := s.db.Conn().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get card names: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]string)
	for rows.Next() {
		var arenaID, name string
		if err := rows.Scan(&arenaID, &name); err != nil {
			continue
		}
		result[arenaID] = name
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating card names: %w", err)
	}

	return result, nil
}

// ClearAllMatches deletes all matches and games for the current account.
func (s *Service) ClearAllMatches(ctx context.Context) error {
	return s.matches.DeleteAll(ctx, s.currentAccountID)
}

// SaveMatch saves a single match with duplicate checking.
// If a match with the same ID already exists, it will be skipped (not an error).
// This is useful for importing matches from backups.
func (s *Service) SaveMatch(ctx context.Context, match *models.Match) error {
	// Check if match already exists
	existing, err := s.matches.GetByID(ctx, match.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing match: %w", err)
	}

	// Skip if match already exists (not an error)
	if existing != nil {
		return nil
	}

	// Create the match (no games for imported matches - they're not stored separately in export)
	return s.matches.Create(ctx, match)
}

// HasProcessedLogFile checks if a log file has already been processed.
func (s *Service) HasProcessedLogFile(ctx context.Context, filename string) (bool, error) {
	db := s.db.Conn()

	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM processed_log_files WHERE filename = ?)
	`, filename).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check processed log file: %w", err)
	}

	return exists, nil
}

// MarkLogFileProcessed marks a log file as processed in the database.
func (s *Service) MarkLogFileProcessed(ctx context.Context, filename string, entryCount, matchesFound int, fileSizeBytes int64) error {
	db := s.db.Conn()

	_, err := db.ExecContext(ctx, `
		INSERT OR REPLACE INTO processed_log_files (
			filename, processed_at, entry_count, matches_found, file_size_bytes
		) VALUES (?, ?, ?, ?, ?)
	`, filename, time.Now(), entryCount, matchesFound, fileSizeBytes)
	if err != nil {
		return fmt.Errorf("failed to mark log file as processed: %w", err)
	}

	return nil
}

// GetProcessedLogFiles returns a list of all processed log files.
func (s *Service) GetProcessedLogFiles(ctx context.Context) ([]ProcessedLogFile, error) {
	db := s.db.Conn()

	rows, err := db.QueryContext(ctx, `
		SELECT filename, processed_at, entry_count, matches_found, file_size_bytes, created_at
		FROM processed_log_files
		ORDER BY processed_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query processed log files: %w", err)
	}
	defer func() {
		_ = rows.Close() // Explicitly ignore error - cleanup operation
	}()

	var files []ProcessedLogFile
	for rows.Next() {
		var file ProcessedLogFile
		err := rows.Scan(
			&file.Filename,
			&file.ProcessedAt,
			&file.EntryCount,
			&file.MatchesFound,
			&file.FileSizeBytes,
			&file.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan processed log file: %w", err)
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating processed log files: %w", err)
	}

	return files, nil
}

// ProcessedLogFile represents a log file that has been processed.
type ProcessedLogFile struct {
	Filename      string
	ProcessedAt   time.Time
	EntryCount    int
	MatchesFound  int
	FileSizeBytes int64
	CreatedAt     time.Time
}

// DeckPerformanceRepo returns the deck performance repository.
func (s *Service) DeckPerformanceRepo() repository.DeckPerformanceRepository {
	return s.deckPerformance
}

// RecommendationFeedbackRepo returns the recommendation feedback repository.
func (s *Service) RecommendationFeedbackRepo() repository.RecommendationFeedbackRepository {
	return s.recommendationFeedback
}

// NewNotesRepo creates a new notes repository using the service's database connection.
func (s *Service) NewNotesRepo() repository.NotesRepository {
	return repository.NewNotesRepository(s.db.Conn())
}

// NewSuggestionRepo creates a new suggestion repository using the service's database connection.
func (s *Service) NewSuggestionRepo() repository.SuggestionRepository {
	return repository.NewSuggestionRepository(s.db.Conn())
}

// NewGamePlayRepo creates a new game play repository using the service's database connection.
func (s *Service) NewGamePlayRepo() repository.GamePlayRepository {
	return repository.NewGamePlayRepository(s.db.Conn())
}

// NewMatchRepo creates a new match repository using the service's database connection.
func (s *Service) NewMatchRepo() repository.MatchRepository {
	return repository.NewMatchRepository(s.db.Conn())
}

// NewMLSuggestionRepo creates a new ML suggestion repository using the service's database connection.
func (s *Service) NewMLSuggestionRepo() *repository.MLSuggestionRepository {
	return repository.NewMLSuggestionRepository(s.db.Conn())
}

// NewDeckRepo creates a new deck repository using the service's database connection.
func (s *Service) NewDeckRepo() repository.DeckRepository {
	return repository.NewDeckRepository(s.db.Conn())
}

// NewSetCardRepo creates a new set card repository using the service's database connection.
func (s *Service) NewSetCardRepo() repository.SetCardRepository {
	return repository.NewSetCardRepository(s.db.Conn())
}

// NewOpponentRepo creates a new opponent repository using the service's database connection.
func (s *Service) NewOpponentRepo() repository.OpponentRepository {
	return repository.NewOpponentRepository(s.db.Conn())
}

// NewCFBRatingsRepo creates a new CFB ratings repository using the service's database connection.
func (s *Service) NewCFBRatingsRepo() repository.CFBRatingsRepository {
	return repository.NewCFBRatingsRepository(s.db.Conn())
}

// Close closes the database connection.
func (s *Service) Close() error {
	return s.db.Close()
}

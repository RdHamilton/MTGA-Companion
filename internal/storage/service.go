package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/stats"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service provides high-level operations for storing and retrieving MTGA data.
type Service struct {
	db               *DB
	matches          repository.MatchRepository
	stats            repository.StatsRepository
	decks            repository.DeckRepository
	collection       repository.CollectionRepository
	accounts         repository.AccountRepository
	rankHistory      repository.RankHistoryRepository
	currentAccountID int // Current active account ID
}

// NewService creates a new storage service.
func NewService(db *DB) *Service {
	svc := &Service{
		db:          db,
		matches:     repository.NewMatchRepository(db.Conn()),
		stats:       repository.NewStatsRepository(db.Conn()),
		decks:       repository.NewDeckRepository(db.Conn()),
		collection:  repository.NewCollectionRepository(db.Conn()),
		accounts:    repository.NewAccountRepository(db.Conn()),
		rankHistory: repository.NewRankHistoryRepository(db.Conn()),
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

// StoreArenaStats stores arena statistics parsed from the log.
// It creates match and game records with deduplication and updates daily stats.
func (s *Service) StoreArenaStats(ctx context.Context, arenaStats *logreader.ArenaStats, entries []*logreader.LogEntry) error {
	if arenaStats == nil {
		return nil
	}

	// Extract match details from log entries for persistent storage
	matchesToStore, err := s.extractMatchesFromEntries(ctx, entries)
	if err != nil {
		return fmt.Errorf("failed to extract matches: %w", err)
	}

	// Correlate ranks with matches
	rankSnapshots := extractRankSnapshots(entries)
	correlateRanksWithMatches(matchesToStore, rankSnapshots)

	// Store matches with deduplication
	for _, matchData := range matchesToStore {
		// Check if match already exists (deduplication)
		existing, err := s.matches.GetByID(ctx, matchData.Match.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing match: %w", err)
		}

		// Only store if match doesn't exist
		if existing == nil {
			if err := s.StoreMatch(ctx, matchData.Match, matchData.Games); err != nil {
				// Ignore duplicate key errors (race condition)
				if !strings.Contains(err.Error(), "UNIQUE constraint") {
					return fmt.Errorf("failed to store match %s: %w", matchData.Match.ID, err)
				}
			}
		}
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

			// Parse timestamp from entry
			matchTime := time.Now()
			if entry.Timestamp != "" {
				// Try to parse timestamp (format: [UnityCrossThreadLogger]2024-01-15 10:30:45)
				if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
					matchTime = parsedTime
				}
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

			firstPlayer, ok := reservedPlayers[0].(map[string]interface{})
			if !ok {
				continue
			}

			eventID, _ := firstPlayer["eventId"].(string)
			if eventID == "" {
				eventID = "Unknown"
			}

			eventName, _ := firstPlayer["eventId"].(string)
			if eventName == "" {
				eventName = eventID
			}

			playerTeamID, _ := firstPlayer["teamId"].(float64)

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

			// Create match record
			match := &Match{
				ID:           matchID,
				AccountID:    s.currentAccountID,
				EventID:      eventID,
				EventName:    eventName,
				Timestamp:    matchTime,
				PlayerWins:   playerWins,
				OpponentWins: opponentWins,
				PlayerTeamID: int(playerTeamID),
				Format:       eventID,
				Result:       matchResult,
				ResultReason: resultReason,
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
	// Format: [UnityCrossThreadLogger]2024-01-15 10:30:45
	// Try to extract the date/time portion
	parts := strings.Fields(timestampStr)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
	}

	dateTimeStr := parts[len(parts)-2] + " " + parts[len(parts)-1]
	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
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

// GetMatches retrieves matches based on the given filter.
func (s *Service) GetMatches(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error) {
	// Use account filter if specified, otherwise use current account
	accountID := s.currentAccountID
	if filter.AccountID != nil {
		if *filter.AccountID == 0 {
			// 0 means all accounts
			accountID = 0
		} else {
			accountID = *filter.AccountID
		}
	}

	var start, end time.Time
	if filter.StartDate != nil {
		start = *filter.StartDate
	} else {
		// Default to all time if no start date
		start = time.Time{}
	}
	if filter.EndDate != nil {
		end = *filter.EndDate
	} else {
		// Default to now if no end date
		end = time.Now()
	}

	matches, err := s.matches.GetByDateRange(ctx, start, end, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	// Filter by format if specified
	if filter.Format != nil {
		filtered := []*models.Match{}
		for _, match := range matches {
			if match.Format == *filter.Format {
				filtered = append(filtered, match)
			}
		}
		return filtered, nil
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

// StoreDeck stores a complete deck with its cards.
func (s *Service) StoreDeck(ctx context.Context, deck *Deck, cards []*DeckCard) error {
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
// Requires card metadata to be populated in the cards table.
func (s *Service) GetSetCompletion(ctx context.Context) ([]*models.SetCompletion, error) {
	// Query all sets and their card counts by rarity from the cards table
	query := `
		SELECT
			set_code,
			set_name,
			rarity,
			COUNT(*) as total
		FROM cards
		GROUP BY set_code, set_name, rarity
		ORDER BY set_code, rarity
	`

	conn := s.db.Conn()
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query card sets: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup
		_ = rows.Close()
	}()

	// Build set data structure
	setData := make(map[string]*models.SetCompletion)
	for rows.Next() {
		var setCode, setName, rarity string
		var total int

		if err := rows.Scan(&setCode, &setName, &rarity, &total); err != nil {
			return nil, fmt.Errorf("failed to scan set data: %w", err)
		}

		// Initialize set if not exists
		if _, exists := setData[setCode]; !exists {
			setData[setCode] = &models.SetCompletion{
				SetCode:         setCode,
				SetName:         setName,
				RarityBreakdown: make(map[string]*models.RarityCompletion),
			}
		}

		// Add rarity breakdown
		setData[setCode].RarityBreakdown[rarity] = &models.RarityCompletion{
			Rarity: rarity,
			Total:  total,
		}
		setData[setCode].TotalCards += total
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating set data: %w", err)
	}

	// Get owned cards from collection
	collection, err := s.collection.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Query card arena IDs with their sets and rarities
	cardQuery := `
		SELECT arena_id, set_code, rarity
		FROM cards
	`

	cardRows, err := conn.QueryContext(ctx, cardQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup
		_ = cardRows.Close()
	}()

	// Count owned cards per set and rarity
	for cardRows.Next() {
		var arenaID int
		var setCode, rarity string

		if err := cardRows.Scan(&arenaID, &setCode, &rarity); err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		// Check if player owns this card
		if quantity, owned := collection[arenaID]; owned && quantity > 0 {
			if set, exists := setData[setCode]; exists {
				set.OwnedCards++
				if rarityData, rarityExists := set.RarityBreakdown[rarity]; rarityExists {
					rarityData.Owned++
				}
			}
		}
	}

	if err = cardRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
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

// Close closes the database connection.
func (s *Service) Close() error {
	return s.db.Close()
}

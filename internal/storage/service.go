package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service provides high-level operations for storing and retrieving MTGA data.
type Service struct {
	db         *DB
	matches    repository.MatchRepository
	stats      repository.StatsRepository
	decks      repository.DeckRepository
	collection repository.CollectionRepository
}

// NewService creates a new storage service.
func NewService(db *DB) *Service {
	return &Service{
		db:         db,
		matches:    repository.NewMatchRepository(db.Conn()),
		stats:      repository.NewStatsRepository(db.Conn()),
		decks:      repository.NewDeckRepository(db.Conn()),
		collection: repository.NewCollectionRepository(db.Conn()),
	}
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

				if scope == "MatchScope_Match" {
					if playerWon {
						matchResult = "win"
					} else {
						matchResult = "loss"
					}
				} else if scope == "MatchScope_Game" {
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
				EventID:      eventID,
				EventName:    eventName,
				Timestamp:    matchTime,
				PlayerWins:   playerWins,
				OpponentWins: opponentWins,
				PlayerTeamID: int(playerTeamID),
				Format:       eventID,
				Result:       matchResult,
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

// GetRecentMatches retrieves matches within a date range.
func (s *Service) GetRecentMatches(ctx context.Context, days int) ([]*Match, error) {
	end := time.Now()
	start := end.Add(-time.Duration(days) * 24 * time.Hour)
	return s.matches.GetByDateRange(ctx, start, end)
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

// Close closes the database connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// Package repository provides data access layers for MTGA data.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MatchRepository handles database operations for matches and games.
type MatchRepository interface {
	// Create inserts a new match into the database.
	Create(ctx context.Context, match *models.Match) error

	// CreateGame inserts a new game for a match.
	CreateGame(ctx context.Context, game *models.Game) error

	// GetByID retrieves a match by its ID.
	GetByID(ctx context.Context, id string) (*models.Match, error)

	// GetByDateRange retrieves all matches within a date range.
	// If accountID is 0, returns matches for all accounts.
	GetByDateRange(ctx context.Context, start, end time.Time, accountID int) ([]*models.Match, error)

	// GetByFormat retrieves all matches for a specific format.
	// If accountID is 0, returns matches for all accounts.
	GetByFormat(ctx context.Context, format string, accountID int) ([]*models.Match, error)

	// GetRecentMatches retrieves the most recent matches.
	// If accountID is 0, returns matches for all accounts.
	GetRecentMatches(ctx context.Context, limit int, accountID int) ([]*models.Match, error)

	// GetLatestMatch retrieves the most recent match.
	// If accountID is 0, returns the latest match for all accounts.
	GetLatestMatch(ctx context.Context, accountID int) (*models.Match, error)

	// GetMatches retrieves matches based on the given filter with advanced filtering support.
	GetMatches(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error)

	// GetStats calculates statistics based on the given filter.
	GetStats(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error)

	// GetStatsByFormat calculates statistics grouped by format.
	GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error)

	// GetStatsByDeck calculates statistics grouped by deck.
	GetStatsByDeck(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error)

	// GetGamesForMatch retrieves all games for a specific match.
	GetGamesForMatch(ctx context.Context, matchID string) ([]*models.Game, error)

	// GetPerformanceMetrics calculates duration-based performance metrics.
	GetPerformanceMetrics(ctx context.Context, filter models.StatsFilter) (*models.PerformanceMetrics, error)

	// GetMatchesWithoutDeckID returns all matches that don't have a deck_id assigned.
	GetMatchesWithoutDeckID(ctx context.Context) ([]*models.Match, error)

	// UpdateDeckID updates the deck_id for a specific match.
	UpdateDeckID(ctx context.Context, matchID, deckID string) error

	// DeleteAll deletes all matches and games from the database.
	// If accountID is > 0, only deletes matches for that account.
	// If accountID is 0, deletes all matches for all accounts.
	DeleteAll(ctx context.Context, accountID int) error
}

// matchRepository is the concrete implementation of MatchRepository.
type matchRepository struct {
	db *sql.DB
}

// NewMatchRepository creates a new match repository.
func NewMatchRepository(db *sql.DB) MatchRepository {
	return &matchRepository{db: db}
}

// buildFilterWhereClause constructs a WHERE clause and args from a StatsFilter.
// This supports advanced filtering including multiple formats, rank ranges, opponent filters, etc.
// If tableAlias is provided (e.g., "m"), columns will be prefixed with it (e.g., "m.format").
func buildFilterWhereClause(filter models.StatsFilter, tableAlias string) (where string, args []interface{}) {
	where = "WHERE 1=1"
	args = make([]interface{}, 0)

	// Helper function to add table prefix if provided
	prefix := func(col string) string {
		if tableAlias != "" {
			return tableAlias + "." + col
		}
		return col
	}

	// Account filter
	if filter.AccountID != nil && *filter.AccountID > 0 {
		where += fmt.Sprintf(" AND %s = ?", prefix("account_id"))
		args = append(args, *filter.AccountID)
	}

	// Date range filters
	if filter.StartDate != nil {
		where += fmt.Sprintf(" AND %s >= ?", prefix("timestamp"))
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		where += fmt.Sprintf(" AND %s <= ?", prefix("timestamp"))
		args = append(args, *filter.EndDate)
	}

	// Format filters (support both single and multiple)
	if len(filter.Formats) > 0 {
		// Multiple formats with OR logic
		placeholders := ""
		for i, format := range filter.Formats {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, format)
		}
		where += fmt.Sprintf(" AND %s IN (%s)", prefix("format"), placeholders)
	} else if filter.Format != nil {
		// Single format (backward compatibility)
		where += fmt.Sprintf(" AND %s = ?", prefix("format"))
		args = append(args, *filter.Format)
	}

	// Deck filter
	if filter.DeckID != nil {
		where += fmt.Sprintf(" AND %s = ?", prefix("deck_id"))
		args = append(args, *filter.DeckID)
	}

	// Deck format filter (requires JOIN with decks table)
	if filter.DeckFormat != nil {
		where += " AND d.format = ?"
		args = append(args, *filter.DeckFormat)
	}

	// Event filters
	if len(filter.EventNames) > 0 {
		// Multiple event names with OR logic
		placeholders := ""
		for i, eventName := range filter.EventNames {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, eventName)
		}
		where += fmt.Sprintf(" AND %s IN (%s)", prefix("event_name"), placeholders)
	} else if filter.EventName != nil {
		// Single event name
		where += fmt.Sprintf(" AND %s = ?", prefix("event_name"))
		args = append(args, *filter.EventName)
	}

	// Opponent filters
	if filter.OpponentName != nil {
		where += fmt.Sprintf(" AND %s = ?", prefix("opponent_name"))
		args = append(args, *filter.OpponentName)
	}
	if filter.OpponentID != nil {
		where += fmt.Sprintf(" AND %s = ?", prefix("opponent_id"))
		args = append(args, *filter.OpponentID)
	}

	// Result filter
	if filter.Result != nil {
		where += fmt.Sprintf(" AND %s = ?", prefix("result"))
		args = append(args, *filter.Result)
	}

	// Result reason filter
	if filter.ResultReason != nil {
		where += fmt.Sprintf(" AND %s = ?", prefix("result_reason"))
		args = append(args, *filter.ResultReason)
	}

	// Rank filters (uses LIKE for rank_before or rank_after)
	if filter.RankClass != nil {
		where += fmt.Sprintf(" AND (%s LIKE ? OR %s LIKE ?)", prefix("rank_before"), prefix("rank_after"))
		rankPattern := "%" + *filter.RankClass + "%"
		args = append(args, rankPattern, rankPattern)
	}
	// Note: RankMinClass and RankMaxClass would require parsing rank strings
	// which is complex (Bronze < Silver < Gold < Platinum < Diamond < Mythic)
	// Deferring this for now as it requires rank comparison logic

	return where, args
}

// Create inserts a new match into the database.
func (r *matchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `
		INSERT INTO matches (
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason,
			opponent_name, opponent_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
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
		return fmt.Errorf("failed to create match: %w", err)
	}

	return nil
}

// CreateGame inserts a new game for a match.
func (r *matchRepository) CreateGame(ctx context.Context, game *models.Game) error {
	query := `
		INSERT INTO games (
			match_id, game_number, result, duration_seconds, result_reason, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		game.MatchID,
		game.GameNumber,
		game.Result,
		game.DurationSeconds,
		game.ResultReason,
		game.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	game.ID = int(id)
	return nil
}

// GetByID retrieves a match by its ID.
func (r *matchRepository) GetByID(ctx context.Context, id string) (*models.Match, error) {
	query := `
		SELECT
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason,
			opponent_name, opponent_id, created_at
		FROM matches
		WHERE id = ?
	`

	match := &models.Match{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.AccountID,
		&match.EventID,
		&match.EventName,
		&match.Timestamp,
		&match.DurationSeconds,
		&match.PlayerWins,
		&match.OpponentWins,
		&match.PlayerTeamID,
		&match.DeckID,
		&match.RankBefore,
		&match.RankAfter,
		&match.Format,
		&match.Result,
		&match.ResultReason,
		&match.OpponentName,
		&match.OpponentID,
		&match.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get match by id: %w", err)
	}

	return match, nil
}

// GetByDateRange retrieves all matches within a date range.
// If accountID is 0, returns matches for all accounts.
func (r *matchRepository) GetByDateRange(ctx context.Context, start, end time.Time, accountID int) ([]*models.Match, error) {
	query := `
		SELECT
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason,
			opponent_name, opponent_id, created_at
		FROM matches
		WHERE timestamp >= ? AND timestamp <= ?
	`
	args := []interface{}{start, end}

	if accountID > 0 {
		query += " AND account_id = ?"
		args = append(args, accountID)
	}

	query += " ORDER BY timestamp DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches by date range: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.AccountID,
			&match.EventID,
			&match.EventName,
			&match.Timestamp,
			&match.DurationSeconds,
			&match.PlayerWins,
			&match.OpponentWins,
			&match.PlayerTeamID,
			&match.DeckID,
			&match.RankBefore,
			&match.RankAfter,
			&match.Format,
			&match.Result,
			&match.ResultReason,
			&match.OpponentName,
			&match.OpponentID,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

// GetByFormat retrieves all matches for a specific format.
// If accountID is 0, returns matches for all accounts.
func (r *matchRepository) GetByFormat(ctx context.Context, format string, accountID int) ([]*models.Match, error) {
	query := `
		SELECT
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason,
			opponent_name, opponent_id, created_at
		FROM matches
		WHERE format = ?
	`
	args := []interface{}{format}

	if accountID > 0 {
		query += " AND account_id = ?"
		args = append(args, accountID)
	}

	query += " ORDER BY timestamp DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches by format: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.AccountID,
			&match.EventID,
			&match.EventName,
			&match.Timestamp,
			&match.DurationSeconds,
			&match.PlayerWins,
			&match.OpponentWins,
			&match.PlayerTeamID,
			&match.DeckID,
			&match.RankBefore,
			&match.RankAfter,
			&match.Format,
			&match.Result,
			&match.ResultReason,
			&match.OpponentName,
			&match.OpponentID,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

// GetMatches retrieves matches based on the given filter with advanced filtering support.
func (r *matchRepository) GetMatches(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error) {
	// Determine if we need to JOIN with decks table and use table aliases
	var fromClause string
	var tableAlias string
	if filter.DeckFormat != nil {
		fromClause = "FROM matches m LEFT JOIN decks d ON m.deck_id = d.id"
		tableAlias = "m" // Use table alias when joining
	} else {
		fromClause = "FROM matches m"
		tableAlias = "m" // Still use alias since we always have "m" in this query
	}

	// Build WHERE clause using the filter builder with table alias
	where, args := buildFilterWhereClause(filter, tableAlias)

	query := fmt.Sprintf(`
		SELECT
			m.id, m.account_id, m.event_id, m.event_name, m.timestamp, m.duration_seconds,
			m.player_wins, m.opponent_wins, m.player_team_id, m.deck_id,
			m.rank_before, m.rank_after, m.format, m.result, m.result_reason,
			m.opponent_name, m.opponent_id, m.created_at
		%s
		%s
		ORDER BY m.timestamp DESC
	`, fromClause, where)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered matches: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.AccountID,
			&match.EventID,
			&match.EventName,
			&match.Timestamp,
			&match.DurationSeconds,
			&match.PlayerWins,
			&match.OpponentWins,
			&match.PlayerTeamID,
			&match.DeckID,
			&match.RankBefore,
			&match.RankAfter,
			&match.Format,
			&match.Result,
			&match.ResultReason,
			&match.OpponentName,
			&match.OpponentID,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

// GetStats calculates statistics based on the given filter.
func (r *matchRepository) GetStats(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error) {
	// Determine table alias and FROM clause based on whether we need to JOIN
	var tableAlias string
	var matchFrom string
	var gameFrom string

	if filter.DeckFormat != nil {
		// Need to JOIN with decks table to filter by deck format
		tableAlias = "m"
		matchFrom = "FROM matches m LEFT JOIN decks d ON m.deck_id = d.id"
		gameFrom = "FROM games g INNER JOIN matches m ON g.match_id = m.id LEFT JOIN decks d ON m.deck_id = d.id"
	} else {
		// No JOIN needed
		tableAlias = ""
		matchFrom = "FROM matches"
		gameFrom = "FROM games g INNER JOIN matches m ON g.match_id = m.id"
	}

	// Build WHERE clause
	where, args := buildFilterWhereClause(filter, tableAlias)

	// Helper to prefix column with table alias if needed
	col := func(name string) string {
		if tableAlias != "" {
			return tableAlias + "." + name
		}
		return name
	}

	// Query for match statistics
	matchQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN %s = 'win' THEN 1 ELSE 0 END), 0) as wins,
			COALESCE(SUM(CASE WHEN %s = 'loss' THEN 1 ELSE 0 END), 0) as losses
		%s
		%s
	`, col("result"), col("result"), matchFrom, where)

	stats := &models.Statistics{}
	err := r.db.QueryRowContext(ctx, matchQuery, args...).Scan(
		&stats.TotalMatches,
		&stats.MatchesWon,
		&stats.MatchesLost,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get match stats: %w", err)
	}

	// Calculate match win rate
	if stats.TotalMatches > 0 {
		stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches)
	}

	// Query for game statistics
	gameQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN g.result = 'win' THEN 1 ELSE 0 END), 0) as wins,
			COALESCE(SUM(CASE WHEN g.result = 'loss' THEN 1 ELSE 0 END), 0) as losses
		%s
		%s
	`, gameFrom, where)

	err = r.db.QueryRowContext(ctx, gameQuery, args...).Scan(
		&stats.TotalGames,
		&stats.GamesWon,
		&stats.GamesLost,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game stats: %w", err)
	}

	// Calculate game win rate
	if stats.TotalGames > 0 {
		stats.GameWinRate = float64(stats.GamesWon) / float64(stats.TotalGames)
	}

	return stats, nil
}

// GetRecentMatches retrieves the most recent matches.
// If accountID is 0, returns matches for all accounts.
func (r *matchRepository) GetRecentMatches(ctx context.Context, limit int, accountID int) ([]*models.Match, error) {
	query := `
		SELECT
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason,
			opponent_name, opponent_id, created_at
		FROM matches
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if accountID > 0 {
		query += " AND account_id = ?"
		args = append(args, accountID)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent matches: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.AccountID,
			&match.EventID,
			&match.EventName,
			&match.Timestamp,
			&match.DurationSeconds,
			&match.PlayerWins,
			&match.OpponentWins,
			&match.PlayerTeamID,
			&match.DeckID,
			&match.RankBefore,
			&match.RankAfter,
			&match.Format,
			&match.Result,
			&match.ResultReason,
			&match.OpponentName,
			&match.OpponentID,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

// GetLatestMatch retrieves the most recent match.
func (r *matchRepository) GetLatestMatch(ctx context.Context, accountID int) (*models.Match, error) {
	query := `
		SELECT
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason,
			opponent_name, opponent_id, created_at
		FROM matches
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if accountID > 0 {
		query += " AND account_id = ?"
		args = append(args, accountID)
	}

	query += " ORDER BY timestamp DESC LIMIT 1"

	match := &models.Match{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&match.ID,
		&match.AccountID,
		&match.EventID,
		&match.EventName,
		&match.Timestamp,
		&match.DurationSeconds,
		&match.PlayerWins,
		&match.OpponentWins,
		&match.PlayerTeamID,
		&match.DeckID,
		&match.RankBefore,
		&match.RankAfter,
		&match.Format,
		&match.Result,
		&match.ResultReason,
		&match.OpponentName,
		&match.OpponentID,
		&match.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No matches found
		}
		return nil, fmt.Errorf("failed to get latest match: %w", err)
	}

	return match, nil
}

// GetStatsByFormat calculates statistics grouped by format.
func (r *matchRepository) GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	// Build WHERE clause based on filter (same as GetStats but without format filter)
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if filter.AccountID != nil && *filter.AccountID > 0 {
		where += " AND account_id = ?"
		args = append(args, *filter.AccountID)
	}
	if filter.StartDate != nil {
		where += " AND timestamp >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		where += " AND timestamp <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.DeckID != nil {
		where += " AND deck_id = ?"
		args = append(args, *filter.DeckID)
	}

	// Query for match statistics grouped by format
	matchQuery := fmt.Sprintf(`
		SELECT
			format,
			COUNT(*) as total,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM matches
		%s
		GROUP BY format
		ORDER BY format ASC
	`, where)

	rows, err := r.db.QueryContext(ctx, matchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by format: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	// Collect match stats by format
	formatStats := make(map[string]*models.Statistics)
	for rows.Next() {
		var format string
		stats := &models.Statistics{}
		err := rows.Scan(
			&format,
			&stats.TotalMatches,
			&stats.MatchesWon,
			&stats.MatchesLost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match stats: %w", err)
		}

		// Calculate match win rate
		if stats.TotalMatches > 0 {
			stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches)
		}

		formatStats[format] = stats
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating match stats: %w", err)
	}

	// Now get game statistics for each format
	for format, stats := range formatStats {
		gameQuery := fmt.Sprintf(`
			SELECT
				COUNT(*) as total,
				SUM(CASE WHEN g.result = 'win' THEN 1 ELSE 0 END) as wins,
				SUM(CASE WHEN g.result = 'loss' THEN 1 ELSE 0 END) as losses
			FROM games g
			INNER JOIN matches m ON g.match_id = m.id
			%s AND m.format = ?
		`, where)

		gameArgs := append(args, format)
		err = r.db.QueryRowContext(ctx, gameQuery, gameArgs...).Scan(
			&stats.TotalGames,
			&stats.GamesWon,
			&stats.GamesLost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get game stats for format %s: %w", format, err)
		}

		// Calculate game win rate
		if stats.TotalGames > 0 {
			stats.GameWinRate = float64(stats.GamesWon) / float64(stats.TotalGames)
		}
	}

	return formatStats, nil
}

// GetStatsByDeck calculates statistics grouped by deck.
func (r *matchRepository) GetStatsByDeck(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	// Build WHERE clause based on filter (same as GetStats but without deck filter)
	where := "WHERE m.deck_id IS NOT NULL" // Only include matches with a deck
	args := make([]interface{}, 0)

	if filter.AccountID != nil && *filter.AccountID > 0 {
		where += " AND m.account_id = ?"
		args = append(args, *filter.AccountID)
	}
	if filter.StartDate != nil {
		where += " AND m.timestamp >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		where += " AND m.timestamp <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.Format != nil {
		where += " AND m.format = ?"
		args = append(args, *filter.Format)
	}

	// Query for match statistics grouped by deck
	matchQuery := fmt.Sprintf(`
		SELECT
			COALESCE(d.name, m.deck_id, 'Unknown Deck') as deck_name,
			COUNT(*) as total,
			SUM(CASE WHEN m.result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN m.result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM matches m
		LEFT JOIN decks d ON m.deck_id = d.id
		%s
		GROUP BY deck_name
		ORDER BY total DESC
	`, where)

	rows, err := r.db.QueryContext(ctx, matchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by deck: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	// Collect match stats by deck
	deckStats := make(map[string]*models.Statistics)
	for rows.Next() {
		var deckName string
		stats := &models.Statistics{}
		err := rows.Scan(
			&deckName,
			&stats.TotalMatches,
			&stats.MatchesWon,
			&stats.MatchesLost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck match stats: %w", err)
		}

		// Calculate match win rate
		if stats.TotalMatches > 0 {
			stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches)
		}

		deckStats[deckName] = stats
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck match stats: %w", err)
	}

	// Query for game statistics grouped by deck (via match_id join)
	gameQuery := fmt.Sprintf(`
		SELECT
			COALESCE(d.name, m.deck_id, 'Unknown Deck') as deck_name,
			COUNT(*) as total,
			SUM(CASE WHEN g.result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN g.result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM games g
		JOIN matches m ON g.match_id = m.id
		LEFT JOIN decks d ON m.deck_id = d.id
		%s
		GROUP BY deck_name
	`, where)

	gameRows, err := r.db.QueryContext(ctx, gameQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get game stats by deck: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = gameRows.Close()
	}()

	for gameRows.Next() {
		var deckName string
		var totalGames, gamesWon, gamesLost int
		err := gameRows.Scan(
			&deckName,
			&totalGames,
			&gamesWon,
			&gamesLost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get game stats for deck %s: %w", deckName, err)
		}

		// Add to existing stats or create new if match stats don't exist
		if stats, exists := deckStats[deckName]; exists {
			stats.TotalGames = totalGames
			stats.GamesWon = gamesWon
			stats.GamesLost = gamesLost
		} else {
			deckStats[deckName] = &models.Statistics{
				TotalGames: totalGames,
				GamesWon:   gamesWon,
				GamesLost:  gamesLost,
			}
		}

		// Calculate game win rate
		if totalGames > 0 {
			deckStats[deckName].GameWinRate = float64(gamesWon) / float64(totalGames)
		}
	}

	if err = gameRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck game stats: %w", err)
	}

	return deckStats, nil
}

// GetGamesForMatch retrieves all games for a specific match.
func (r *matchRepository) GetGamesForMatch(ctx context.Context, matchID string) ([]*models.Game, error) {
	query := `
		SELECT id, match_id, game_number, result, duration_seconds, result_reason, created_at
		FROM games
		WHERE match_id = ?
		ORDER BY game_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get games for match: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var games []*models.Game
	for rows.Next() {
		game := &models.Game{}
		err := rows.Scan(
			&game.ID,
			&game.MatchID,
			&game.GameNumber,
			&game.Result,
			&game.DurationSeconds,
			&game.ResultReason,
			&game.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}

	return games, nil
}

// GetPerformanceMetrics calculates duration-based performance metrics.
func (r *matchRepository) GetPerformanceMetrics(ctx context.Context, filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	// Build WHERE clause based on filter
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if filter.AccountID != nil && *filter.AccountID > 0 {
		where += " AND account_id = ?"
		args = append(args, *filter.AccountID)
	}
	if filter.StartDate != nil {
		where += " AND timestamp >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		where += " AND timestamp <= ?"
		args = append(args, *filter.EndDate)
	}
	if filter.Format != nil {
		where += " AND format = ?"
		args = append(args, *filter.Format)
	}
	if filter.DeckID != nil {
		where += " AND deck_id = ?"
		args = append(args, *filter.DeckID)
	}

	// Get match duration metrics (only consider matches with duration data)
	matchQuery := fmt.Sprintf(`
		SELECT
			AVG(duration_seconds) as avg_duration,
			MIN(duration_seconds) as min_duration,
			MAX(duration_seconds) as max_duration
		FROM matches
		%s AND duration_seconds IS NOT NULL
	`, where)

	metrics := &models.PerformanceMetrics{}
	var avgMatch, minMatch, maxMatch sql.NullFloat64

	err := r.db.QueryRowContext(ctx, matchQuery, args...).Scan(
		&avgMatch,
		&minMatch,
		&maxMatch,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get match duration metrics: %w", err)
	}

	// Convert to int pointers for min/max
	if avgMatch.Valid {
		avg := avgMatch.Float64
		metrics.AvgMatchDuration = &avg
	}
	if minMatch.Valid {
		min := int(minMatch.Float64)
		metrics.FastestMatch = &min
	}
	if maxMatch.Valid {
		max := int(maxMatch.Float64)
		metrics.SlowestMatch = &max
	}

	// Get game duration metrics (only consider games with duration data)
	gameQuery := fmt.Sprintf(`
		SELECT
			AVG(g.duration_seconds) as avg_duration,
			MIN(g.duration_seconds) as min_duration,
			MAX(g.duration_seconds) as max_duration
		FROM games g
		INNER JOIN matches m ON g.match_id = m.id
		%s AND g.duration_seconds IS NOT NULL
	`, where)

	var avgGame, minGame, maxGame sql.NullFloat64

	err = r.db.QueryRowContext(ctx, gameQuery, args...).Scan(
		&avgGame,
		&minGame,
		&maxGame,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get game duration metrics: %w", err)
	}

	// Convert to int pointers for min/max
	if avgGame.Valid {
		avg := avgGame.Float64
		metrics.AvgGameDuration = &avg
	}
	if minGame.Valid {
		min := int(minGame.Float64)
		metrics.FastestGame = &min
	}
	if maxGame.Valid {
		max := int(maxGame.Float64)
		metrics.SlowestGame = &max
	}

	return metrics, nil
}

// GetMatchesWithoutDeckID returns all matches that don't have a deck_id assigned.
func (r *matchRepository) GetMatchesWithoutDeckID(ctx context.Context) ([]*models.Match, error) {
	query := `
		SELECT id, account_id, event_id, event_name, timestamp, player_wins, opponent_wins,
			   player_team_id, format, result, result_reason, created_at
		FROM matches
		WHERE deck_id IS NULL
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var matches []*models.Match
	for rows.Next() {
		var m models.Match
		var resultReason sql.NullString

		err := rows.Scan(
			&m.ID,
			&m.AccountID,
			&m.EventID,
			&m.EventName,
			&m.Timestamp,
			&m.PlayerWins,
			&m.OpponentWins,
			&m.PlayerTeamID,
			&m.Format,
			&m.Result,
			&resultReason,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}

		if resultReason.Valid {
			m.ResultReason = &resultReason.String
		}

		matches = append(matches, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}

// UpdateDeckID updates the deck_id for a specific match.
func (r *matchRepository) UpdateDeckID(ctx context.Context, matchID, deckID string) error {
	query := `UPDATE matches SET deck_id = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, deckID, matchID)
	if err != nil {
		return fmt.Errorf("failed to update match deck_id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("match not found: %s", matchID)
	}

	return nil
}

// DeleteAll deletes all matches and games from the database.
// If accountID is > 0, only deletes matches for that account.
// If accountID is 0, deletes all matches for all accounts.
func (r *matchRepository) DeleteAll(ctx context.Context, accountID int) error {
	// Start a transaction to ensure atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Delete games first (foreign key constraint)
	var gamesQuery string
	var gamesArgs []interface{}
	if accountID > 0 {
		gamesQuery = `DELETE FROM games WHERE match_id IN (SELECT id FROM matches WHERE account_id = ?)`
		gamesArgs = []interface{}{accountID}
	} else {
		gamesQuery = `DELETE FROM games`
		gamesArgs = []interface{}{}
	}

	if _, err := tx.ExecContext(ctx, gamesQuery, gamesArgs...); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete games: %w", err)
	}

	// Delete matches
	var matchesQuery string
	var matchesArgs []interface{}
	if accountID > 0 {
		matchesQuery = `DELETE FROM matches WHERE account_id = ?`
		matchesArgs = []interface{}{accountID}
	} else {
		matchesQuery = `DELETE FROM matches`
		matchesArgs = []interface{}{}
	}

	result, err := tx.ExecContext(ctx, matchesQuery, matchesArgs...)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete matches: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Deleted %d matches and associated games", rowsAffected)

	return nil
}

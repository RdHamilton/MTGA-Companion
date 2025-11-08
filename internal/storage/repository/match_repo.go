// Package repository provides data access layers for MTGA data.
package repository

import (
	"context"
	"database/sql"
	"fmt"
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

	// GetStats calculates statistics based on the given filter.
	GetStats(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error)

	// GetStatsByFormat calculates statistics grouped by format.
	GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error)

	// GetGamesForMatch retrieves all games for a specific match.
	GetGamesForMatch(ctx context.Context, matchID string) ([]*models.Game, error)
}

// matchRepository is the concrete implementation of MatchRepository.
type matchRepository struct {
	db *sql.DB
}

// NewMatchRepository creates a new match repository.
func NewMatchRepository(db *sql.DB) MatchRepository {
	return &matchRepository{db: db}
}

// Create inserts a new match into the database.
func (r *matchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `
		INSERT INTO matches (
			id, account_id, event_id, event_name, timestamp, duration_seconds,
			player_wins, opponent_wins, player_team_id, deck_id,
			rank_before, rank_after, format, result, result_reason, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			match_id, game_number, result, duration_seconds, created_at
		) VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		game.MatchID,
		game.GameNumber,
		game.Result,
		game.DurationSeconds,
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
			rank_before, rank_after, format, result, result_reason, created_at
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
			rank_before, rank_after, format, result, result_reason, created_at
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
			rank_before, rank_after, format, result, result_reason, created_at
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

	// Query for match statistics
	matchQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM matches
		%s
	`, where)

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
			SUM(CASE WHEN g.result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN g.result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM games g
		INNER JOIN matches m ON g.match_id = m.id
		%s
	`, where)

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
			rank_before, rank_after, format, result, result_reason, created_at
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

// GetGamesForMatch retrieves all games for a specific match.
func (r *matchRepository) GetGamesForMatch(ctx context.Context, matchID string) ([]*models.Game, error) {
	query := `
		SELECT id, match_id, game_number, result, duration_seconds, created_at
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

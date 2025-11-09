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

	// GetLatestMatch retrieves the most recent match.
	// If accountID is 0, returns the latest match for all accounts.
	GetLatestMatch(ctx context.Context, accountID int) (*models.Match, error)

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
	where := "WHERE deck_id IS NOT NULL" // Only include matches with a deck
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

	// Query for match statistics grouped by deck
	matchQuery := fmt.Sprintf(`
		SELECT
			deck_id,
			COUNT(*) as total,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM matches
		%s
		GROUP BY deck_id
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
		var deckID string
		stats := &models.Statistics{}
		err := rows.Scan(
			&deckID,
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

		deckStats[deckID] = stats
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck match stats: %w", err)
	}

	// Query for game statistics grouped by deck (via match_id join)
	gameQuery := fmt.Sprintf(`
		SELECT
			m.deck_id,
			COUNT(*) as total,
			SUM(CASE WHEN g.result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN g.result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM games g
		JOIN matches m ON g.match_id = m.id
		%s
		GROUP BY m.deck_id
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
		var deckID string
		var totalGames, gamesWon, gamesLost int
		err := gameRows.Scan(
			&deckID,
			&totalGames,
			&gamesWon,
			&gamesLost,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get game stats for deck %s: %w", deckID, err)
		}

		// Add to existing stats or create new if match stats don't exist
		if stats, exists := deckStats[deckID]; exists {
			stats.TotalGames = totalGames
			stats.GamesWon = gamesWon
			stats.GamesLost = gamesLost
		} else {
			deckStats[deckID] = &models.Statistics{
				TotalGames: totalGames,
				GamesWon:   gamesWon,
				GamesLost:  gamesLost,
			}
		}

		// Calculate game win rate
		if totalGames > 0 {
			deckStats[deckID].GameWinRate = float64(gamesWon) / float64(totalGames)
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

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// StatsRepository handles database operations for player statistics.
type StatsRepository interface {
	// Upsert inserts or updates player stats for a specific date and format.
	Upsert(ctx context.Context, stats *models.PlayerStats) error

	// GetByDate retrieves stats for a specific date and format.
	GetByDate(ctx context.Context, date time.Time, format string) (*models.PlayerStats, error)

	// GetByDateRange retrieves stats within a date range for a format.
	GetByDateRange(ctx context.Context, start, end time.Time, format string) ([]*models.PlayerStats, error)

	// GetAllFormats retrieves stats for all formats on a specific date.
	GetAllFormats(ctx context.Context, date time.Time) ([]*models.PlayerStats, error)
}

// statsRepository is the concrete implementation of StatsRepository.
type statsRepository struct {
	db *sql.DB
}

// NewStatsRepository creates a new stats repository.
func NewStatsRepository(db *sql.DB) StatsRepository {
	return &statsRepository{db: db}
}

// Upsert inserts or updates player stats for a specific date and format.
func (r *statsRepository) Upsert(ctx context.Context, stats *models.PlayerStats) error {
	query := `
		INSERT INTO player_stats (
			date, format, matches_played, matches_won,
			games_played, games_won, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(date, format) DO UPDATE SET
			matches_played = excluded.matches_played,
			matches_won = excluded.matches_won,
			games_played = excluded.games_played,
			games_won = excluded.games_won,
			updated_at = excluded.updated_at
	`

	result, err := r.db.ExecContext(ctx, query,
		stats.Date,
		stats.Format,
		stats.MatchesPlayed,
		stats.MatchesWon,
		stats.GamesPlayed,
		stats.GamesWon,
		stats.CreatedAt,
		stats.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert player stats: %w", err)
	}

	// If this is an insert, set the ID
	if stats.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			stats.ID = int(id)
		}
	}

	return nil
}

// GetByDate retrieves stats for a specific date and format.
func (r *statsRepository) GetByDate(ctx context.Context, date time.Time, format string) (*models.PlayerStats, error) {
	query := `
		SELECT
			id, date, format, matches_played, matches_won,
			games_played, games_won, created_at, updated_at
		FROM player_stats
		WHERE date = ? AND format = ?
	`

	stats := &models.PlayerStats{}
	err := r.db.QueryRowContext(ctx, query, date, format).Scan(
		&stats.ID,
		&stats.Date,
		&stats.Format,
		&stats.MatchesPlayed,
		&stats.MatchesWon,
		&stats.GamesPlayed,
		&stats.GamesWon,
		&stats.CreatedAt,
		&stats.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by date: %w", err)
	}

	return stats, nil
}

// GetByDateRange retrieves stats within a date range for a format.
func (r *statsRepository) GetByDateRange(ctx context.Context, start, end time.Time, format string) ([]*models.PlayerStats, error) {
	query := `
		SELECT
			id, date, format, matches_played, matches_won,
			games_played, games_won, created_at, updated_at
		FROM player_stats
		WHERE date >= ? AND date <= ? AND format = ?
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats by date range: %w", err)
	}
	defer rows.Close()

	var stats []*models.PlayerStats
	for rows.Next() {
		s := &models.PlayerStats{}
		err := rows.Scan(
			&s.ID,
			&s.Date,
			&s.Format,
			&s.MatchesPlayed,
			&s.MatchesWon,
			&s.GamesPlayed,
			&s.GamesWon,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stats: %w", err)
	}

	return stats, nil
}

// GetAllFormats retrieves stats for all formats on a specific date.
func (r *statsRepository) GetAllFormats(ctx context.Context, date time.Time) ([]*models.PlayerStats, error) {
	query := `
		SELECT
			id, date, format, matches_played, matches_won,
			games_played, games_won, created_at, updated_at
		FROM player_stats
		WHERE date = ?
		ORDER BY format ASC
	`

	rows, err := r.db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for all formats: %w", err)
	}
	defer rows.Close()

	var stats []*models.PlayerStats
	for rows.Next() {
		s := &models.PlayerStats{}
		err := rows.Scan(
			&s.ID,
			&s.Date,
			&s.Format,
			&s.MatchesPlayed,
			&s.MatchesWon,
			&s.GamesPlayed,
			&s.GamesWon,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stats: %w", err)
	}

	return stats, nil
}

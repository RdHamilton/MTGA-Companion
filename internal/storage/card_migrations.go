package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

// GetProcessedMigrationIDs returns a list of migration IDs that have already been processed.
func (s *Service) GetProcessedMigrationIDs(ctx context.Context) ([]string, error) {
	conn := s.db.Conn()

	query := `SELECT migration_id FROM migration_log ORDER BY processed_at ASC`

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query processed migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var migrationIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan migration ID: %w", err)
		}
		migrationIDs = append(migrationIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
	}

	return migrationIDs, nil
}

// LogMigration records a processed migration in the database.
func (s *Service) LogMigration(ctx context.Context, migration *scryfall.Migration) error {
	conn := s.db.Conn()

	var newScryfallID *string
	if migration.NewScryfallID != nil {
		newScryfallID = migration.NewScryfallID
	}

	query := `
		INSERT INTO migration_log (
			migration_id,
			old_scryfall_id,
			new_scryfall_id,
			strategy,
			note,
			performed_at,
			processed_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := conn.ExecContext(
		ctx,
		query,
		migration.ID,
		migration.OldScryfallID,
		newScryfallID,
		migration.MigrationStrategy,
		migration.Note,
		migration.PerformedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to log migration %s: %w", migration.ID, err)
	}

	return nil
}

// UpdateCardScryfallID updates a card's Scryfall ID (for merge migrations).
func (s *Service) UpdateCardScryfallID(ctx context.Context, oldID, newID string) error {
	conn := s.db.Conn()

	query := `UPDATE cards SET scryfall_id = ? WHERE scryfall_id = ?`

	result, err := conn.ExecContext(ctx, query, newID, oldID)
	if err != nil {
		return fmt.Errorf("failed to update card Scryfall ID: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// It's okay if no rows were affected - the card might not exist in our database
	_ = rowsAffected

	return nil
}

// DeleteCardByScryfallID deletes a card by its Scryfall ID (for delete migrations).
func (s *Service) DeleteCardByScryfallID(ctx context.Context, scryfallID string) error {
	conn := s.db.Conn()

	query := `DELETE FROM cards WHERE scryfall_id = ?`

	result, err := conn.ExecContext(ctx, query, scryfallID)
	if err != nil {
		return fmt.Errorf("failed to delete card by Scryfall ID: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// It's okay if no rows were affected - the card might not exist in our database
	_ = rowsAffected

	return nil
}

// GetMigrationStats returns statistics about processed migrations.
func (s *Service) GetMigrationStats(ctx context.Context) (*MigrationStats, error) {
	conn := s.db.Conn()

	stats := &MigrationStats{}

	// Get total count
	query := `SELECT COUNT(*) FROM migration_log`
	if err := conn.QueryRowContext(ctx, query).Scan(&stats.TotalMigrations); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get total migration count: %w", err)
		}
	}

	// Get merge count
	query = `SELECT COUNT(*) FROM migration_log WHERE strategy = 'merge'`
	if err := conn.QueryRowContext(ctx, query).Scan(&stats.MergeCount); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get merge count: %w", err)
		}
	}

	// Get delete count
	query = `SELECT COUNT(*) FROM migration_log WHERE strategy = 'delete'`
	if err := conn.QueryRowContext(ctx, query).Scan(&stats.DeleteCount); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get delete count: %w", err)
		}
	}

	// Get last processed time
	query = `SELECT MAX(processed_at) FROM migration_log`
	var lastProcessed sql.NullTime
	if err := conn.QueryRowContext(ctx, query).Scan(&lastProcessed); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get last processed time: %w", err)
		}
	}
	if lastProcessed.Valid {
		stats.LastProcessedAt = &lastProcessed.Time
	}

	return stats, nil
}

// MigrationStats contains statistics about processed migrations.
type MigrationStats struct {
	TotalMigrations int
	MergeCount      int
	DeleteCount     int
	LastProcessedAt *time.Time
}

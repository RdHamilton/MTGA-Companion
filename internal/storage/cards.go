package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Set represents a Magic set in the local cache.
type Set struct {
	Code            string
	Name            string
	ReleasedAt      *string // Date when set was released (may be NULL for unreleased sets)
	CardCount       int
	SetType         *string // Type of set (may be NULL)
	IconSVGURI      *string // URL to the set symbol SVG (may be NULL)
	CachedAt        time.Time
	LastUpdated     time.Time
	IsStandardLegal bool    // Whether the set is currently Standard-legal
	RotationDate    *string // Date when set rotates out of Standard (ISO 8601)
}

// SaveSet saves or updates a set in the database.
func (s *Service) SaveSet(ctx context.Context, set *Set) error {
	query := `
		INSERT INTO sets (
			code, name, released_at, card_count, set_type, icon_svg_uri, last_updated
		) VALUES (
			?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
		)
		ON CONFLICT(code) DO UPDATE SET
			name = excluded.name,
			released_at = excluded.released_at,
			card_count = excluded.card_count,
			set_type = excluded.set_type,
			icon_svg_uri = excluded.icon_svg_uri,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := s.db.Conn().ExecContext(ctx, query,
		set.Code, set.Name, set.ReleasedAt, set.CardCount, set.SetType, set.IconSVGURI,
	)
	if err != nil {
		return fmt.Errorf("failed to save set: %w", err)
	}

	return nil
}

// GetSet retrieves a set by its code.
func (s *Service) GetSet(ctx context.Context, code string) (*Set, error) {
	query := `
		SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated,
		       COALESCE(is_standard_legal, FALSE), rotation_date
		FROM sets
		WHERE code = ?
	`

	var set Set
	err := s.db.Conn().QueryRowContext(ctx, query, code).Scan(
		&set.Code, &set.Name, &set.ReleasedAt, &set.CardCount, &set.SetType, &set.IconSVGURI,
		&set.CachedAt, &set.LastUpdated, &set.IsStandardLegal, &set.RotationDate,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get set: %w", err)
	}

	return &set, nil
}

// GetAllSets retrieves all sets ordered by release date (newest first).
func (s *Service) GetAllSets(ctx context.Context) ([]*Set, error) {
	query := `
		SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated,
		       COALESCE(is_standard_legal, FALSE), rotation_date
		FROM sets
		ORDER BY released_at DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sets []*Set
	for rows.Next() {
		var set Set
		err := rows.Scan(
			&set.Code, &set.Name, &set.ReleasedAt, &set.CardCount, &set.SetType, &set.IconSVGURI,
			&set.CachedAt, &set.LastUpdated, &set.IsStandardLegal, &set.RotationDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan set: %w", err)
		}
		sets = append(sets, &set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sets: %w", err)
	}

	return sets, nil
}

// GetStaleSets retrieves sets that haven't been updated in the specified duration.
func (s *Service) GetStaleSets(ctx context.Context, olderThan time.Duration) ([]*Set, error) {
	// Calculate seconds for SQLite datetime modifier
	seconds := int64(olderThan.Seconds())

	query := `
		SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated,
		       COALESCE(is_standard_legal, FALSE), rotation_date
		FROM sets
		WHERE unixepoch(last_updated) <= unixepoch('now', '-' || ? || ' seconds')
		ORDER BY last_updated ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var sets []*Set
	for rows.Next() {
		var set Set
		err := rows.Scan(
			&set.Code, &set.Name, &set.ReleasedAt, &set.CardCount, &set.SetType, &set.IconSVGURI,
			&set.CachedAt, &set.LastUpdated, &set.IsStandardLegal, &set.RotationDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan set: %w", err)
		}
		sets = append(sets, &set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sets: %w", err)
	}

	return sets, nil
}

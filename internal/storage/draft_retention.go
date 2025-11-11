package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RetentionPolicy defines rules for keeping historical draft statistics.
type RetentionPolicy struct {
	// MinimumAge is the minimum age before any snapshot can be deleted
	MinimumAge time.Duration

	// ActiveSetRetention defines how long to keep snapshots for active (Standard) sets
	// 0 means keep forever
	ActiveSetRetention time.Duration

	// RotatedSetInterval defines the interval for keeping snapshots of rotated sets
	// e.g., 7 days means keep weekly snapshots
	RotatedSetInterval time.Duration

	// HistoricalSetInterval defines the interval for keeping snapshots of historical sets (>1 year old)
	// e.g., 30 days means keep monthly snapshots
	HistoricalSetInterval time.Duration
}

// DefaultRetentionPolicy returns the default retention policy.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MinimumAge:            90 * 24 * time.Hour, // 90 days minimum
		ActiveSetRetention:    0,                   // Keep all active set data
		RotatedSetInterval:    7 * 24 * time.Hour,  // Weekly for rotated
		HistoricalSetInterval: 30 * 24 * time.Hour, // Monthly for historical
	}
}

// SnapshotInfo contains information about a draft statistics snapshot.
type SnapshotInfo struct {
	ID          int
	ArenaID     int
	Expansion   string
	Format      string
	Colors      string
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time

	// Set metadata for retention decisions
	SetReleaseDate time.Time
	IsActive       bool
	IsRotated      bool
	IsHistorical   bool
}

// CleanupResult contains statistics about a cleanup operation.
type CleanupResult struct {
	TotalSnapshots    int
	RemovedSnapshots  int
	RetainedSnapshots int
	OldestSnapshot    time.Time
	NewestSnapshot    time.Time
	DryRun            bool
	RemovedBySet      map[string]int
	RetainedBySet     map[string]int
}

// GetAllSnapshots returns all draft card rating snapshots for analysis.
func (s *Service) GetAllSnapshots(ctx context.Context) ([]*SnapshotInfo, error) {
	query := `
		SELECT
			id, arena_id, expansion, format, colors,
			start_date, end_date, cached_at, last_updated
		FROM draft_card_ratings
		ORDER BY cached_at DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshots: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var snapshots []*SnapshotInfo
	for rows.Next() {
		snapshot := &SnapshotInfo{}
		var cachedAtStr, lastUpdatedStr string

		err := rows.Scan(
			&snapshot.ID,
			&snapshot.ArenaID,
			&snapshot.Expansion,
			&snapshot.Format,
			&snapshot.Colors,
			&snapshot.StartDate,
			&snapshot.EndDate,
			&cachedAtStr,
			&lastUpdatedStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		// Parse timestamps
		snapshot.CachedAt, _ = time.Parse("2006-01-02 15:04:05", cachedAtStr)
		snapshot.LastUpdated, _ = time.Parse("2006-01-02 15:04:05", lastUpdatedStr)

		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snapshots: %w", err)
	}

	return snapshots, nil
}

// CleanupOldSnapshots removes old snapshots according to the retention policy.
// If dryRun is true, returns what would be deleted without actually deleting.
func (s *Service) CleanupOldSnapshots(ctx context.Context, policy RetentionPolicy, dryRun bool) (*CleanupResult, error) {
	result := &CleanupResult{
		DryRun:        dryRun,
		RemovedBySet:  make(map[string]int),
		RetainedBySet: make(map[string]int),
	}

	// Get all snapshots
	snapshots, err := s.GetAllSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	result.TotalSnapshots = len(snapshots)
	if len(snapshots) == 0 {
		return result, nil
	}

	// Track oldest and newest
	result.NewestSnapshot = snapshots[0].CachedAt
	result.OldestSnapshot = snapshots[len(snapshots)-1].CachedAt

	// Group snapshots by (expansion, format, colors) for analysis
	type GroupKey struct {
		Expansion string
		Format    string
		Colors    string
	}

	groups := make(map[GroupKey][]*SnapshotInfo)
	for _, snapshot := range snapshots {
		key := GroupKey{
			Expansion: snapshot.Expansion,
			Format:    snapshot.Format,
			Colors:    snapshot.Colors,
		}
		groups[key] = append(groups[key], snapshot)
	}

	// Process each group
	var toDelete []int
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		// Determine retention strategy for this group
		expansion := group[0].Expansion

		// For now, keep all snapshots within minimum age
		now := time.Now()
		for _, snapshot := range group {
			age := now.Sub(snapshot.CachedAt)

			if age < policy.MinimumAge {
				// Always keep recent snapshots
				result.RetainedSnapshots++
				result.RetainedBySet[expansion]++
				continue
			}

			// For older snapshots, apply interval-based retention
			// Keep one snapshot per interval (weekly or monthly)
			shouldKeep := s.shouldKeepSnapshot(snapshot, group, policy, now)
			if shouldKeep {
				result.RetainedSnapshots++
				result.RetainedBySet[expansion]++
			} else {
				result.RemovedSnapshots++
				result.RemovedBySet[expansion]++
				toDelete = append(toDelete, snapshot.ID)
			}
		}
	}

	// Delete snapshots if not dry run
	if !dryRun && len(toDelete) > 0 {
		// Delete in batches
		batchSize := 100
		for i := 0; i < len(toDelete); i += batchSize {
			end := i + batchSize
			if end > len(toDelete) {
				end = len(toDelete)
			}
			batch := toDelete[i:end]

			if err := s.deleteSnapshotsBatch(ctx, batch); err != nil {
				return result, fmt.Errorf("failed to delete batch: %w", err)
			}
		}
	}

	return result, nil
}

// shouldKeepSnapshot determines if a snapshot should be kept based on retention policy.
func (s *Service) shouldKeepSnapshot(snapshot *SnapshotInfo, groupSnapshots []*SnapshotInfo, policy RetentionPolicy, now time.Time) bool {
	age := now.Sub(snapshot.CachedAt)

	// Determine interval based on set age
	// For now, use a simple heuristic:
	// - Recent sets (< 6 months): Keep weekly snapshots
	// - Older sets: Keep monthly snapshots
	var interval time.Duration
	if age < 180*24*time.Hour { // 6 months
		interval = policy.RotatedSetInterval // Weekly
	} else {
		interval = policy.HistoricalSetInterval // Monthly
	}

	// Check if this is the closest snapshot to an interval boundary
	// This is a simplified approach - keep one snapshot per interval
	intervalStart := snapshot.CachedAt.Truncate(interval)

	// Check if there's a more recent snapshot in this interval
	for _, other := range groupSnapshots {
		if other.ID == snapshot.ID {
			continue
		}

		otherIntervalStart := other.CachedAt.Truncate(interval)
		if otherIntervalStart.Equal(intervalStart) {
			// Another snapshot in same interval
			// Keep the more recent one
			if other.CachedAt.After(snapshot.CachedAt) {
				return false // Don't keep this one, keep the other
			}
		}
	}

	return true
}

// deleteSnapshotsBatch deletes a batch of snapshots by ID.
func (s *Service) deleteSnapshotsBatch(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	// Build IN clause
	query := `DELETE FROM draft_card_ratings WHERE id IN (`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	_, err := s.db.Conn().ExecContext(ctx, query, args...)
	return err
}

// GetSnapshotCount returns the number of snapshots for each expansion.
func (s *Service) GetSnapshotCount(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT expansion, COUNT(*) as count
		FROM draft_card_ratings
		GROUP BY expansion
		ORDER BY count DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	for rows.Next() {
		var expansion string
		var count int
		if err := rows.Scan(&expansion, &count); err != nil {
			return nil, fmt.Errorf("failed to scan count: %w", err)
		}
		counts[expansion] = count
	}

	return counts, rows.Err()
}

// GetOldestSnapshot returns the oldest snapshot date.
func (s *Service) GetOldestSnapshot(ctx context.Context) (time.Time, error) {
	query := `SELECT MIN(cached_at) FROM draft_card_ratings`

	var cachedAtStr sql.NullString
	err := s.db.Conn().QueryRowContext(ctx, query).Scan(&cachedAtStr)
	if err != nil {
		return time.Time{}, err
	}

	if !cachedAtStr.Valid {
		return time.Time{}, nil
	}

	return time.Parse("2006-01-02 15:04:05", cachedAtStr.String)
}

// GetNewestSnapshot returns the newest snapshot date.
func (s *Service) GetNewestSnapshot(ctx context.Context) (time.Time, error) {
	query := `SELECT MAX(cached_at) FROM draft_card_ratings`

	var cachedAtStr sql.NullString
	err := s.db.Conn().QueryRowContext(ctx, query).Scan(&cachedAtStr)
	if err != nil {
		return time.Time{}, err
	}

	if !cachedAtStr.Valid {
		return time.Time{}, nil
	}

	return time.Parse("2006-01-02 15:04:05", cachedAtStr.String)
}

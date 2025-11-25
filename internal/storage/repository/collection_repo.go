package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CollectionEntry represents a card and its quantity in the collection.
type CollectionEntry struct {
	CardID   int
	Quantity int
}

// CollectionRepository handles database operations for card collection.
type CollectionRepository interface {
	// UpsertCard inserts or updates a card in the collection.
	UpsertCard(ctx context.Context, cardID int, quantity int) error

	// UpsertMany inserts or updates multiple cards in a single transaction.
	// This is optimized for bulk sync operations (<1s for full collection).
	UpsertMany(ctx context.Context, entries []CollectionEntry) error

	// GetCard retrieves the quantity of a specific card.
	GetCard(ctx context.Context, cardID int) (int, error)

	// GetAll retrieves the entire collection as a map of cardID -> quantity.
	GetAll(ctx context.Context) (map[int]int, error)

	// RecordChange records a change to the collection in the history table
	// and updates the collection with the new quantity (current + delta).
	RecordChange(ctx context.Context, cardID int, delta int, timestamp time.Time, source *string) error

	// RecordHistoryEntry records a history entry without updating the collection.
	// Use this when the collection has already been updated (e.g., after UpsertMany).
	RecordHistoryEntry(ctx context.Context, cardID int, delta int, quantityAfter int, timestamp time.Time, source *string) error

	// GetHistory retrieves collection history for a specific card.
	GetHistory(ctx context.Context, cardID int) ([]*models.CollectionHistory, error)

	// GetRecentChanges retrieves recent collection changes.
	GetRecentChanges(ctx context.Context, limit int) ([]*models.CollectionHistory, error)

	// GetChangesSince retrieves collection changes since a specific time.
	GetChangesSince(ctx context.Context, since time.Time) ([]*models.CollectionHistory, error)
}

// collectionRepository is the concrete implementation of CollectionRepository.
type collectionRepository struct {
	db *sql.DB
}

// NewCollectionRepository creates a new collection repository.
func NewCollectionRepository(db *sql.DB) CollectionRepository {
	return &collectionRepository{db: db}
}

// UpsertCard inserts or updates a card in the collection.
func (r *collectionRepository) UpsertCard(ctx context.Context, cardID int, quantity int) error {
	query := `
		INSERT INTO collection (card_id, quantity, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(card_id) DO UPDATE SET
			quantity = excluded.quantity,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query, cardID, quantity, time.Now())
	if err != nil {
		return fmt.Errorf("failed to upsert card: %w", err)
	}

	return nil
}

// GetCard retrieves the quantity of a specific card.
func (r *collectionRepository) GetCard(ctx context.Context, cardID int) (int, error) {
	query := `SELECT quantity FROM collection WHERE card_id = ?`

	var quantity int
	err := r.db.QueryRowContext(ctx, query, cardID).Scan(&quantity)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get card quantity: %w", err)
	}

	return quantity, nil
}

// GetAll retrieves the entire collection as a map of cardID -> quantity.
func (r *collectionRepository) GetAll(ctx context.Context) (map[int]int, error) {
	query := `SELECT card_id, quantity FROM collection`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all cards: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	collection := make(map[int]int)
	for rows.Next() {
		var cardID, quantity int
		err := rows.Scan(&cardID, &quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		collection[cardID] = quantity
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating collection: %w", err)
	}

	return collection, nil
}

// RecordChange records a change to the collection in the history table.
func (r *collectionRepository) RecordChange(ctx context.Context, cardID int, delta int, timestamp time.Time, source *string) error {
	// Get current quantity
	currentQuantity, err := r.GetCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get current quantity: %w", err)
	}

	quantityAfter := currentQuantity + delta

	query := `
		INSERT INTO collection_history (
			card_id, quantity_delta, quantity_after, timestamp, source, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		cardID,
		delta,
		quantityAfter,
		timestamp,
		source,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record collection change: %w", err)
	}

	// Update the collection
	if err := r.UpsertCard(ctx, cardID, quantityAfter); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	return nil
}

// RecordHistoryEntry records a history entry without updating the collection.
// Use this when the collection has already been updated (e.g., after UpsertMany).
func (r *collectionRepository) RecordHistoryEntry(ctx context.Context, cardID int, delta int, quantityAfter int, timestamp time.Time, source *string) error {
	query := `
		INSERT INTO collection_history (
			card_id, quantity_delta, quantity_after, timestamp, source, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		cardID,
		delta,
		quantityAfter,
		timestamp,
		source,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record history entry: %w", err)
	}

	return nil
}

// GetHistory retrieves collection history for a specific card.
func (r *collectionRepository) GetHistory(ctx context.Context, cardID int) ([]*models.CollectionHistory, error) {
	query := `
		SELECT id, card_id, quantity_delta, quantity_after, timestamp, source, created_at
		FROM collection_history
		WHERE card_id = ?
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection history: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var history []*models.CollectionHistory
	for rows.Next() {
		h := &models.CollectionHistory{}
		err := rows.Scan(
			&h.ID,
			&h.CardID,
			&h.QuantityDelta,
			&h.QuantityAfter,
			&h.Timestamp,
			&h.Source,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}

// GetRecentChanges retrieves recent collection changes.
func (r *collectionRepository) GetRecentChanges(ctx context.Context, limit int) ([]*models.CollectionHistory, error) {
	query := `
		SELECT id, card_id, quantity_delta, quantity_after, timestamp, source, created_at
		FROM collection_history
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent changes: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var history []*models.CollectionHistory
	for rows.Next() {
		h := &models.CollectionHistory{}
		err := rows.Scan(
			&h.ID,
			&h.CardID,
			&h.QuantityDelta,
			&h.QuantityAfter,
			&h.Timestamp,
			&h.Source,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}

// UpsertMany inserts or updates multiple cards in a single transaction.
// Uses batch operations for performance (<1s for full collection).
func (r *collectionRepository) UpsertMany(ctx context.Context, entries []CollectionEntry) error {
	if len(entries) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on rollback after commit
		_ = tx.Rollback()
	}()

	// Prepare the upsert statement
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO collection (card_id, quantity, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(card_id) DO UPDATE SET
			quantity = excluded.quantity,
			updated_at = excluded.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup
		_ = stmt.Close()
	}()

	now := time.Now()
	for _, entry := range entries {
		_, err := stmt.ExecContext(ctx, entry.CardID, entry.Quantity, now)
		if err != nil {
			return fmt.Errorf("failed to upsert card %d: %w", entry.CardID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetChangesSince retrieves collection changes since a specific time.
func (r *collectionRepository) GetChangesSince(ctx context.Context, since time.Time) ([]*models.CollectionHistory, error) {
	query := `
		SELECT id, card_id, quantity_delta, quantity_after, timestamp, source, created_at
		FROM collection_history
		WHERE timestamp > ?
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes since %v: %w", since, err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var history []*models.CollectionHistory
	for rows.Next() {
		h := &models.CollectionHistory{}
		err := rows.Scan(
			&h.ID,
			&h.CardID,
			&h.QuantityDelta,
			&h.QuantityAfter,
			&h.Timestamp,
			&h.Source,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}

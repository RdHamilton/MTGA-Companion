package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CollectionRepository handles database operations for card collection.
type CollectionRepository interface {
	// UpsertCard inserts or updates a card in the collection.
	UpsertCard(ctx context.Context, cardID int, quantity int) error

	// GetCard retrieves the quantity of a specific card.
	GetCard(ctx context.Context, cardID int) (int, error)

	// GetAll retrieves the entire collection as a map of cardID -> quantity.
	GetAll(ctx context.Context) (map[int]int, error)

	// RecordChange records a change to the collection in the history table.
	RecordChange(ctx context.Context, cardID int, delta int, timestamp time.Time, source *string) error

	// GetHistory retrieves collection history for a specific card.
	GetHistory(ctx context.Context, cardID int) ([]*models.CollectionHistory, error)

	// GetRecentChanges retrieves recent collection changes.
	GetRecentChanges(ctx context.Context, limit int) ([]*models.CollectionHistory, error)
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
		if err := rows.Close(); err != nil {
			// Log error but don't fail - this is cleanup
		}
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
		if err := rows.Close(); err != nil {
			// Log error but don't fail - this is cleanup
		}
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
		if err := rows.Close(); err != nil {
			// Log error but don't fail - this is cleanup
		}
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

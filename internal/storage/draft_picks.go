package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// SaveDraftPick stores a single draft pick in the database.
func (s *Service) SaveDraftPick(ctx context.Context, pick *models.DraftPick) error {
	// Convert available cards slice to JSON
	availableCardsJSON, err := json.Marshal(pick.AvailableCards)
	if err != nil {
		return fmt.Errorf("failed to marshal available cards: %w", err)
	}

	query := `
		INSERT INTO draft_picks (draft_event_id, pack_number, pick_number, available_cards, selected_card, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(draft_event_id, pack_number, pick_number) DO UPDATE SET
			available_cards = excluded.available_cards,
			selected_card = excluded.selected_card,
			timestamp = excluded.timestamp
	`

	_, err = s.db.Conn().ExecContext(ctx, query,
		pick.DraftEventID,
		pick.PackNumber,
		pick.PickNumber,
		string(availableCardsJSON),
		pick.SelectedCard,
		pick.Timestamp,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save draft pick: %w", err)
	}

	return nil
}

// SaveDraftPicks stores multiple draft picks for a draft event.
func (s *Service) SaveDraftPicks(ctx context.Context, draftEventID string, picks []*models.DraftPick) error {
	if len(picks) == 0 {
		return nil
	}

	tx, err := s.db.Conn().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback returns error if already committed, which is fine
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO draft_picks (draft_event_id, pack_number, pick_number, available_cards, selected_card, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(draft_event_id, pack_number, pick_number) DO UPDATE SET
			available_cards = excluded.available_cards,
			selected_card = excluded.selected_card,
			timestamp = excluded.timestamp
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, pick := range picks {
		// Ensure the pick belongs to the correct draft event
		pick.DraftEventID = draftEventID

		// Convert available cards slice to JSON
		availableCardsJSON, err := json.Marshal(pick.AvailableCards)
		if err != nil {
			return fmt.Errorf("failed to marshal available cards: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			pick.DraftEventID,
			pick.PackNumber,
			pick.PickNumber,
			string(availableCardsJSON),
			pick.SelectedCard,
			pick.Timestamp,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to save draft pick: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDraftPicks retrieves all picks for a specific draft event.
func (s *Service) GetDraftPicks(ctx context.Context, draftEventID string) ([]*models.DraftPick, error) {
	query := `
		SELECT id, draft_event_id, pack_number, pick_number, available_cards, selected_card, timestamp, created_at
		FROM draft_picks
		WHERE draft_event_id = ?
		ORDER BY pack_number, pick_number
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, draftEventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query draft picks: %w", err)
	}
	defer rows.Close()

	var picks []*models.DraftPick
	for rows.Next() {
		var pick models.DraftPick
		var availableCardsJSON string

		err := rows.Scan(
			&pick.ID,
			&pick.DraftEventID,
			&pick.PackNumber,
			&pick.PickNumber,
			&availableCardsJSON,
			&pick.SelectedCard,
			&pick.Timestamp,
			&pick.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan draft pick: %w", err)
		}

		// Parse available cards JSON
		if err := json.Unmarshal([]byte(availableCardsJSON), &pick.AvailableCards); err != nil {
			return nil, fmt.Errorf("failed to unmarshal available cards: %w", err)
		}

		picks = append(picks, &pick)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating draft picks: %w", err)
	}

	return picks, nil
}

// GetDraftPickByNumber retrieves a specific pick from a draft event.
func (s *Service) GetDraftPickByNumber(ctx context.Context, draftEventID string, packNumber, pickNumber int) (*models.DraftPick, error) {
	query := `
		SELECT id, draft_event_id, pack_number, pick_number, available_cards, selected_card, timestamp, created_at
		FROM draft_picks
		WHERE draft_event_id = ? AND pack_number = ? AND pick_number = ?
	`

	var pick models.DraftPick
	var availableCardsJSON string

	err := s.db.Conn().QueryRowContext(ctx, query, draftEventID, packNumber, pickNumber).Scan(
		&pick.ID,
		&pick.DraftEventID,
		&pick.PackNumber,
		&pick.PickNumber,
		&availableCardsJSON,
		&pick.SelectedCard,
		&pick.Timestamp,
		&pick.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("draft pick not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query draft pick: %w", err)
	}

	// Parse available cards JSON
	if err := json.Unmarshal([]byte(availableCardsJSON), &pick.AvailableCards); err != nil {
		return nil, fmt.Errorf("failed to unmarshal available cards: %w", err)
	}

	return &pick, nil
}

// GetDraftPicksCount returns the number of picks stored for a draft event.
func (s *Service) GetDraftPicksCount(ctx context.Context, draftEventID string) (int, error) {
	query := `SELECT COUNT(*) FROM draft_picks WHERE draft_event_id = ?`

	var count int
	err := s.db.Conn().QueryRowContext(ctx, query, draftEventID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count draft picks: %w", err)
	}

	return count, nil
}

// DeleteDraftPicks removes all picks for a specific draft event.
func (s *Service) DeleteDraftPicks(ctx context.Context, draftEventID string) error {
	query := `DELETE FROM draft_picks WHERE draft_event_id = ?`

	_, err := s.db.Conn().ExecContext(ctx, query, draftEventID)
	if err != nil {
		return fmt.Errorf("failed to delete draft picks: %w", err)
	}

	return nil
}

// GetAllDraftEventsWithPicks returns all draft events that have stored picks.
func (s *Service) GetAllDraftEventsWithPicks(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT draft_event_id
		FROM draft_picks
		ORDER BY draft_event_id
	`

	rows, err := s.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query draft events with picks: %w", err)
	}
	defer rows.Close()

	var eventIDs []string
	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return nil, fmt.Errorf("failed to scan event ID: %w", err)
		}
		eventIDs = append(eventIDs, eventID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event IDs: %w", err)
	}

	return eventIDs, nil
}

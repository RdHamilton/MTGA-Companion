package storage

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// SaveDraftEvent stores or updates a draft event in the database.
func (s *Service) SaveDraftEvent(ctx context.Context, event *models.DraftEvent) error {
	// Ensure the event has an account ID
	if event.AccountID == 0 {
		event.AccountID = s.currentAccountID
	}

	query := `
		INSERT INTO draft_events (id, account_id, event_name, set_code, start_time, end_time, wins, losses, status, deck_id, entry_fee, rewards, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			end_time = excluded.end_time,
			wins = excluded.wins,
			losses = excluded.losses,
			status = excluded.status,
			deck_id = excluded.deck_id,
			rewards = excluded.rewards
	`

	_, err := s.db.Conn().ExecContext(ctx, query,
		event.ID,
		event.AccountID,
		event.EventName,
		event.SetCode,
		event.StartTime,
		event.EndTime,
		event.Wins,
		event.Losses,
		event.Status,
		event.DeckID,
		event.EntryFee,
		event.Rewards,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save draft event: %w", err)
	}

	return nil
}

// GetDraftEvent retrieves a draft event by ID.
func (s *Service) GetDraftEvent(ctx context.Context, id string) (*models.DraftEvent, error) {
	query := `
		SELECT id, account_id, event_name, set_code, start_time, end_time, wins, losses, status, deck_id, entry_fee, rewards, created_at
		FROM draft_events
		WHERE id = ?
	`

	var event models.DraftEvent
	err := s.db.Conn().QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.AccountID,
		&event.EventName,
		&event.SetCode,
		&event.StartTime,
		&event.EndTime,
		&event.Wins,
		&event.Losses,
		&event.Status,
		&event.DeckID,
		&event.EntryFee,
		&event.Rewards,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get draft event: %w", err)
	}

	return &event, nil
}

// GetAllDraftEvents retrieves all draft events for the current account.
func (s *Service) GetAllDraftEvents(ctx context.Context) ([]*models.DraftEvent, error) {
	query := `
		SELECT id, account_id, event_name, set_code, start_time, end_time, wins, losses, status, deck_id, entry_fee, rewards, created_at
		FROM draft_events
		WHERE account_id = ?
		ORDER BY start_time DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, s.currentAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query draft events: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var events []*models.DraftEvent
	for rows.Next() {
		var event models.DraftEvent
		err := rows.Scan(
			&event.ID,
			&event.AccountID,
			&event.EventName,
			&event.SetCode,
			&event.StartTime,
			&event.EndTime,
			&event.Wins,
			&event.Losses,
			&event.Status,
			&event.DeckID,
			&event.EntryFee,
			&event.Rewards,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan draft event: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating draft events: %w", err)
	}

	return events, nil
}

// GetActiveEvents retrieves all currently active draft events for the current account.
func (s *Service) GetActiveEvents(ctx context.Context) ([]*models.DraftEvent, error) {
	query := `
		SELECT id, account_id, event_name, set_code, start_time, end_time, wins, losses, status, deck_id, entry_fee, rewards, created_at
		FROM draft_events
		WHERE account_id = ? AND status = 'active'
		ORDER BY start_time DESC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, s.currentAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active draft events: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var events []*models.DraftEvent
	for rows.Next() {
		var event models.DraftEvent
		err := rows.Scan(
			&event.ID,
			&event.AccountID,
			&event.EventName,
			&event.SetCode,
			&event.StartTime,
			&event.EndTime,
			&event.Wins,
			&event.Losses,
			&event.Status,
			&event.DeckID,
			&event.EntryFee,
			&event.Rewards,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active draft event: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active draft events: %w", err)
	}

	return events, nil
}

// EventWinDistribution represents the distribution of event results.
type EventWinDistribution struct {
	Record string `json:"record"` // e.g., "7-0", "7-1", "6-3"
	Count  int    `json:"count"`  // Number of events with this record
}

// GetEventWinDistribution calculates the distribution of event win-loss records.
// Returns a map of record strings (e.g., "7-0", "7-1") to counts.
func (s *Service) GetEventWinDistribution(ctx context.Context) ([]*EventWinDistribution, error) {
	query := `
		SELECT
			printf('%d-%d', wins, losses) as record,
			COUNT(*) as count
		FROM draft_events
		WHERE account_id = ? AND status = 'completed'
		GROUP BY wins, losses
		ORDER BY wins DESC, losses ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, s.currentAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query event win distribution: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var distribution []*EventWinDistribution
	for rows.Next() {
		var record EventWinDistribution
		err := rows.Scan(&record.Record, &record.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event win distribution: %w", err)
		}
		distribution = append(distribution, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event win distribution: %w", err)
	}

	return distribution, nil
}

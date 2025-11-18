package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftRepository provides methods for managing draft sessions, picks, and packs.
type DraftRepository interface {
	// Sessions
	CreateSession(ctx context.Context, session *models.DraftSession) error
	GetSession(ctx context.Context, id string) (*models.DraftSession, error)
	GetActiveSessions(ctx context.Context) ([]*models.DraftSession, error)
	UpdateSessionStatus(ctx context.Context, id string, status string, endTime *time.Time) error
	IncrementSessionPicks(ctx context.Context, id string) error

	// Picks
	SavePick(ctx context.Context, pick *models.DraftPickSession) error
	GetPicksBySession(ctx context.Context, sessionID string) ([]*models.DraftPickSession, error)
	GetPickByNumber(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPickSession, error)

	// Packs
	SavePack(ctx context.Context, pack *models.DraftPackSession) error
	GetPacksBySession(ctx context.Context, sessionID string) ([]*models.DraftPackSession, error)
	GetPack(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPackSession, error)
}

type draftRepository struct {
	db *sql.DB
}

// NewDraftRepository creates a new draft repository.
func NewDraftRepository(db *sql.DB) DraftRepository {
	return &draftRepository{db: db}
}

// CreateSession creates a new draft session.
func (r *draftRepository) CreateSession(ctx context.Context, session *models.DraftSession) error {
	query := `
		INSERT INTO draft_sessions (id, event_name, set_code, draft_type, start_time, status, total_picks, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.EventName,
		session.SetCode,
		session.DraftType,
		session.StartTime,
		session.Status,
		session.TotalPicks,
		session.CreatedAt,
		session.UpdatedAt,
	)
	return err
}

// GetSession retrieves a draft session by ID.
func (r *draftRepository) GetSession(ctx context.Context, id string) (*models.DraftSession, error) {
	query := `
		SELECT id, event_name, set_code, draft_type, start_time, end_time, status, total_picks, created_at, updated_at
		FROM draft_sessions
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	session := &models.DraftSession{}
	var endTime sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.EventName,
		&session.SetCode,
		&session.DraftType,
		&session.StartTime,
		&endTime,
		&session.Status,
		&session.TotalPicks,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return session, nil
}

// GetActiveSessions retrieves all active draft sessions.
func (r *draftRepository) GetActiveSessions(ctx context.Context) ([]*models.DraftSession, error) {
	query := `
		SELECT id, event_name, set_code, draft_type, start_time, end_time, status, total_picks, created_at, updated_at
		FROM draft_sessions
		WHERE status = 'in_progress'
		ORDER BY start_time DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	sessions := []*models.DraftSession{}
	for rows.Next() {
		session := &models.DraftSession{}
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.EventName,
			&session.SetCode,
			&session.DraftType,
			&session.StartTime,
			&endTime,
			&session.Status,
			&session.TotalPicks,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// UpdateSessionStatus updates the status and optionally the end time of a draft session.
func (r *draftRepository) UpdateSessionStatus(ctx context.Context, id string, status string, endTime *time.Time) error {
	query := `
		UPDATE draft_sessions
		SET status = ?, end_time = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, endTime, time.Now(), id)
	return err
}

// IncrementSessionPicks increments the total_picks counter for a session.
func (r *draftRepository) IncrementSessionPicks(ctx context.Context, id string) error {
	query := `
		UPDATE draft_sessions
		SET total_picks = total_picks + 1, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// SavePick saves a draft pick.
func (r *draftRepository) SavePick(ctx context.Context, pick *models.DraftPickSession) error {
	query := `
		INSERT INTO draft_picks (session_id, pack_number, pick_number, card_id, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		pick.SessionID,
		pick.PackNumber,
		pick.PickNumber,
		pick.CardID,
		pick.Timestamp,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pick.ID = int(id)

	return nil
}

// GetPicksBySession retrieves all picks for a draft session.
func (r *draftRepository) GetPicksBySession(ctx context.Context, sessionID string) ([]*models.DraftPickSession, error) {
	query := `
		SELECT id, session_id, pack_number, pick_number, card_id, timestamp
		FROM draft_picks
		WHERE session_id = ?
		ORDER BY pack_number, pick_number
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	picks := []*models.DraftPickSession{}
	for rows.Next() {
		pick := &models.DraftPickSession{}

		err := rows.Scan(
			&pick.ID,
			&pick.SessionID,
			&pick.PackNumber,
			&pick.PickNumber,
			&pick.CardID,
			&pick.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		picks = append(picks, pick)
	}

	return picks, rows.Err()
}

// GetPickByNumber retrieves a specific pick by pack and pick number.
func (r *draftRepository) GetPickByNumber(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPickSession, error) {
	query := `
		SELECT id, session_id, pack_number, pick_number, card_id, timestamp
		FROM draft_picks
		WHERE session_id = ? AND pack_number = ? AND pick_number = ?
	`
	row := r.db.QueryRowContext(ctx, query, sessionID, packNum, pickNum)

	pick := &models.DraftPickSession{}
	err := row.Scan(
		&pick.ID,
		&pick.SessionID,
		&pick.PackNumber,
		&pick.PickNumber,
		&pick.CardID,
		&pick.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return pick, nil
}

// SavePack saves a draft pack.
func (r *draftRepository) SavePack(ctx context.Context, pack *models.DraftPackSession) error {
	// Convert []string to JSON for storage
	cardIDsJSON, err := json.Marshal(pack.CardIDs)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO draft_packs (session_id, pack_number, pick_number, card_ids, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		pack.SessionID,
		pack.PackNumber,
		pack.PickNumber,
		string(cardIDsJSON),
		pack.Timestamp,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pack.ID = int(id)

	return nil
}

// GetPacksBySession retrieves all packs for a draft session.
func (r *draftRepository) GetPacksBySession(ctx context.Context, sessionID string) ([]*models.DraftPackSession, error) {
	query := `
		SELECT id, session_id, pack_number, pick_number, card_ids, timestamp
		FROM draft_packs
		WHERE session_id = ?
		ORDER BY pack_number, pick_number
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	packs := []*models.DraftPackSession{}
	for rows.Next() {
		pack := &models.DraftPackSession{}
		var cardIDsJSON string

		err := rows.Scan(
			&pack.ID,
			&pack.SessionID,
			&pack.PackNumber,
			&pack.PickNumber,
			&cardIDsJSON,
			&pack.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON back to []string
		if err := json.Unmarshal([]byte(cardIDsJSON), &pack.CardIDs); err != nil {
			return nil, err
		}

		packs = append(packs, pack)
	}

	return packs, rows.Err()
}

// GetPack retrieves a specific pack by pack and pick number.
func (r *draftRepository) GetPack(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPackSession, error) {
	query := `
		SELECT id, session_id, pack_number, pick_number, card_ids, timestamp
		FROM draft_packs
		WHERE session_id = ? AND pack_number = ? AND pick_number = ?
	`
	row := r.db.QueryRowContext(ctx, query, sessionID, packNum, pickNum)

	pack := &models.DraftPackSession{}
	var cardIDsJSON string

	err := row.Scan(
		&pack.ID,
		&pack.SessionID,
		&pack.PackNumber,
		&pack.PickNumber,
		&cardIDsJSON,
		&pack.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse JSON back to []string
	if err := json.Unmarshal([]byte(cardIDsJSON), &pack.CardIDs); err != nil {
		return nil, err
	}

	return pack, nil
}

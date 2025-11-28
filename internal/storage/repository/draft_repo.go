package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftRepository provides methods for managing draft sessions, picks, and packs.
type DraftRepository interface {
	// Sessions
	CreateSession(ctx context.Context, session *models.DraftSession) error
	GetSession(ctx context.Context, id string) (*models.DraftSession, error)
	GetActiveSessions(ctx context.Context) ([]*models.DraftSession, error)
	GetCompletedSessions(ctx context.Context, limit int) ([]*models.DraftSession, error)
	UpdateSessionStatus(ctx context.Context, id string, status string, endTime *time.Time) error
	UpdateSessionTotalPicks(ctx context.Context, id string, totalPicks int) error
	IncrementSessionPicks(ctx context.Context, id string) error

	// Picks
	SavePick(ctx context.Context, pick *models.DraftPickSession) error
	GetPicksBySession(ctx context.Context, sessionID string) ([]*models.DraftPickSession, error)
	GetPickByNumber(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPickSession, error)
	UpdatePickQuality(ctx context.Context, pickID int, grade string, rank int, packBestGIHWR, pickedCardGIHWR float64, alternativesJSON string) error

	// Packs
	SavePack(ctx context.Context, pack *models.DraftPackSession) error
	GetPacksBySession(ctx context.Context, sessionID string) ([]*models.DraftPackSession, error)
	GetPack(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPackSession, error)

	// Grades
	UpdateSessionGrade(ctx context.Context, sessionID string, overallGrade string, overallScore int, pickQuality, colorDiscipline, deckComposition, strategic float64) error

	// Predictions
	UpdateSessionPrediction(ctx context.Context, sessionID string, winRate, winRateMin, winRateMax float64, factorsJSON string, predictedAt time.Time) error
}

type draftRepository struct {
	db *sql.DB
}

// NewDraftRepository creates a new draft repository.
func NewDraftRepository(db *sql.DB) DraftRepository {
	return &draftRepository{db: db}
}

// CreateSession creates a new draft session.
// Uses INSERT OR REPLACE to handle replays where the same draft session may be processed multiple times.
func (r *draftRepository) CreateSession(ctx context.Context, session *models.DraftSession) error {
	query := `
		INSERT OR REPLACE INTO draft_sessions (id, event_name, set_code, draft_type, start_time, status, total_picks, created_at, updated_at)
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
		SELECT id, event_name, set_code, draft_type, start_time, end_time, status, total_picks,
			overall_grade, overall_score, pick_quality_score, color_discipline_score,
			deck_composition_score, strategic_score,
			predicted_win_rate, predicted_win_rate_min, predicted_win_rate_max,
			prediction_factors, predicted_at,
			created_at, updated_at
		FROM draft_sessions
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	session := &models.DraftSession{}
	var endTime sql.NullTime
	var overallGrade sql.NullString
	var overallScore sql.NullInt64
	var pickQuality sql.NullFloat64
	var colorDiscipline sql.NullFloat64
	var deckComposition sql.NullFloat64
	var strategic sql.NullFloat64
	var predictedWinRate sql.NullFloat64
	var predictedWinRateMin sql.NullFloat64
	var predictedWinRateMax sql.NullFloat64
	var predictionFactors sql.NullString
	var predictedAt sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.EventName,
		&session.SetCode,
		&session.DraftType,
		&session.StartTime,
		&endTime,
		&session.Status,
		&session.TotalPicks,
		&overallGrade,
		&overallScore,
		&pickQuality,
		&colorDiscipline,
		&deckComposition,
		&strategic,
		&predictedWinRate,
		&predictedWinRateMin,
		&predictedWinRateMax,
		&predictionFactors,
		&predictedAt,
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
	if overallGrade.Valid {
		session.OverallGrade = &overallGrade.String
	}
	if overallScore.Valid {
		score := int(overallScore.Int64)
		session.OverallScore = &score
	}
	if pickQuality.Valid {
		session.PickQualityScore = &pickQuality.Float64
	}
	if colorDiscipline.Valid {
		session.ColorDisciplineScore = &colorDiscipline.Float64
	}
	if deckComposition.Valid {
		session.DeckCompositionScore = &deckComposition.Float64
	}
	if strategic.Valid {
		session.StrategicScore = &strategic.Float64
	}
	if predictedWinRate.Valid {
		session.PredictedWinRate = &predictedWinRate.Float64
	}
	if predictedWinRateMin.Valid {
		session.PredictedWinRateMin = &predictedWinRateMin.Float64
	}
	if predictedWinRateMax.Valid {
		session.PredictedWinRateMax = &predictedWinRateMax.Float64
	}
	if predictionFactors.Valid {
		session.PredictionFactors = &predictionFactors.String
	}
	if predictedAt.Valid {
		session.PredictedAt = &predictedAt.Time
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

// GetCompletedSessions retrieves completed draft sessions ordered by completion date.
func (r *draftRepository) GetCompletedSessions(ctx context.Context, limit int) ([]*models.DraftSession, error) {
	query := `
		SELECT id, event_name, set_code, draft_type, start_time, end_time, status, total_picks, created_at, updated_at
		FROM draft_sessions
		WHERE status = 'completed'
		ORDER BY start_time DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
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

// UpdateSessionTotalPicks updates the total_picks value for a session.
func (r *draftRepository) UpdateSessionTotalPicks(ctx context.Context, id string, totalPicks int) error {
	query := `
		UPDATE draft_sessions
		SET total_picks = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, totalPicks, time.Now(), id)
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
// Uses INSERT OR REPLACE to handle replays where the same pick may be processed multiple times.
func (r *draftRepository) SavePick(ctx context.Context, pick *models.DraftPickSession) error {
	query := `
		INSERT OR REPLACE INTO draft_picks (session_id, pack_number, pick_number, card_id, timestamp)
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
		SELECT id, session_id, pack_number, pick_number, card_id, timestamp,
			pick_quality_grade, pick_quality_rank, pack_best_gihwr, picked_card_gihwr, alternatives_json
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
		var grade sql.NullString
		var rank sql.NullInt64
		var packBestGIHWR sql.NullFloat64
		var pickedCardGIHWR sql.NullFloat64
		var alternativesJSON sql.NullString

		err := rows.Scan(
			&pick.ID,
			&pick.SessionID,
			&pick.PackNumber,
			&pick.PickNumber,
			&pick.CardID,
			&pick.Timestamp,
			&grade,
			&rank,
			&packBestGIHWR,
			&pickedCardGIHWR,
			&alternativesJSON,
		)
		if err != nil {
			return nil, err
		}

		if grade.Valid {
			pick.PickQualityGrade = &grade.String
		}
		if rank.Valid {
			rankInt := int(rank.Int64)
			pick.PickQualityRank = &rankInt
		}
		if packBestGIHWR.Valid {
			pick.PackBestGIHWR = &packBestGIHWR.Float64
		}
		if pickedCardGIHWR.Valid {
			pick.PickedCardGIHWR = &pickedCardGIHWR.Float64
		}
		if alternativesJSON.Valid {
			pick.AlternativesJSON = &alternativesJSON.String
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
// Uses INSERT OR REPLACE to handle replays where the same pack may be processed multiple times.
func (r *draftRepository) SavePack(ctx context.Context, pack *models.DraftPackSession) error {
	log.Printf("[SavePack] Saving pack: session=%s, pack=%d, pick=%d, cards=%d",
		pack.SessionID, pack.PackNumber, pack.PickNumber, len(pack.CardIDs))

	// Convert []string to JSON for storage
	cardIDsJSON, err := json.Marshal(pack.CardIDs)
	if err != nil {
		return err
	}

	query := `
		INSERT OR REPLACE INTO draft_packs (session_id, pack_number, pick_number, card_ids, timestamp)
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
		log.Printf("[SavePack] ERROR saving pack: %v", err)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pack.ID = int(id)

	log.Printf("[SavePack] Successfully saved pack with ID=%d", pack.ID)
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
	log.Printf("[GetPack] Looking for pack: session=%s, pack=%d, pick=%d", sessionID, packNum, pickNum)

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
			log.Printf("[GetPack] No pack found for session=%s, pack=%d, pick=%d", sessionID, packNum, pickNum)
			return nil, nil
		}
		log.Printf("[GetPack] ERROR querying pack: %v", err)
		return nil, err
	}

	// Parse JSON back to []string
	if err := json.Unmarshal([]byte(cardIDsJSON), &pack.CardIDs); err != nil {
		log.Printf("[GetPack] ERROR unmarshaling card IDs: %v", err)
		return nil, err
	}

	log.Printf("[GetPack] Found pack with %d cards", len(pack.CardIDs))
	return pack, nil
}

// UpdatePickQuality updates the pick quality analysis fields for a pick.
func (r *draftRepository) UpdatePickQuality(ctx context.Context, pickID int, grade string, rank int, packBestGIHWR, pickedCardGIHWR float64, alternativesJSON string) error {
	query := `
		UPDATE draft_picks
		SET pick_quality_grade = ?,
			pick_quality_rank = ?,
			pack_best_gihwr = ?,
			picked_card_gihwr = ?,
			alternatives_json = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, grade, rank, packBestGIHWR, pickedCardGIHWR, alternativesJSON, pickID)
	return err
}

// UpdateSessionGrade updates the grade fields for a draft session.
func (r *draftRepository) UpdateSessionGrade(ctx context.Context, sessionID string, overallGrade string, overallScore int, pickQuality, colorDiscipline, deckComposition, strategic float64) error {
	query := `
		UPDATE draft_sessions
		SET overall_grade = ?,
			overall_score = ?,
			pick_quality_score = ?,
			color_discipline_score = ?,
			deck_composition_score = ?,
			strategic_score = ?,
			updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, overallGrade, overallScore, pickQuality, colorDiscipline, deckComposition, strategic, time.Now(), sessionID)
	return err
}

// UpdateSessionPrediction updates the win rate prediction fields for a draft session.
func (r *draftRepository) UpdateSessionPrediction(ctx context.Context, sessionID string, winRate, winRateMin, winRateMax float64, factorsJSON string, predictedAt time.Time) error {
	query := `
		UPDATE draft_sessions
		SET predicted_win_rate = ?,
		    predicted_win_rate_min = ?,
		    predicted_win_rate_max = ?,
		    prediction_factors = ?,
		    predicted_at = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, winRate, winRateMin, winRateMax, factorsJSON, predictedAt, sessionID)
	return err
}

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// GamePlayRepository handles database operations for in-game play tracking.
type GamePlayRepository interface {
	// CreatePlay inserts a new game play record.
	CreatePlay(ctx context.Context, play *models.GamePlay) error

	// CreatePlays inserts multiple game plays in a batch.
	CreatePlays(ctx context.Context, plays []*models.GamePlay) error

	// CreateSnapshot inserts a new game state snapshot.
	CreateSnapshot(ctx context.Context, snapshot *models.GameStateSnapshot) error

	// CreateSnapshots inserts multiple game state snapshots in a batch.
	CreateSnapshots(ctx context.Context, snapshots []*models.GameStateSnapshot) error

	// RecordOpponentCard records or updates an opponent card observation.
	RecordOpponentCard(ctx context.Context, card *models.OpponentCardObserved) error

	// GetPlaysByMatch retrieves all plays for a match, ordered by sequence number.
	GetPlaysByMatch(ctx context.Context, matchID string) ([]*models.GamePlay, error)

	// GetPlaysByGame retrieves all plays for a specific game within a match.
	GetPlaysByGame(ctx context.Context, gameID int) ([]*models.GamePlay, error)

	// GetSnapshotsByMatch retrieves all snapshots for a match.
	GetSnapshotsByMatch(ctx context.Context, matchID string) ([]*models.GameStateSnapshot, error)

	// GetSnapshotsByGame retrieves all snapshots for a specific game.
	GetSnapshotsByGame(ctx context.Context, gameID int) ([]*models.GameStateSnapshot, error)

	// GetOpponentCardsByMatch retrieves all opponent cards observed in a match.
	GetOpponentCardsByMatch(ctx context.Context, matchID string) ([]*models.OpponentCardObserved, error)

	// GetOpponentCardsByGame retrieves opponent cards observed in a specific game.
	GetOpponentCardsByGame(ctx context.Context, gameID int) ([]*models.OpponentCardObserved, error)

	// GetPlayTimeline retrieves plays organized by turn/phase for timeline display.
	GetPlayTimeline(ctx context.Context, matchID string) ([]*models.PlayTimelineEntry, error)

	// GetPlaySummary retrieves aggregated play statistics for a match.
	GetPlaySummary(ctx context.Context, matchID string) (*models.GamePlaySummary, error)

	// DeletePlaysByMatch removes all plays for a match (for cleanup/replay).
	DeletePlaysByMatch(ctx context.Context, matchID string) error

	// DeletePlaysByGame removes all plays for a specific game.
	DeletePlaysByGame(ctx context.Context, gameID int) error
}

// gamePlayRepository is the concrete implementation.
type gamePlayRepository struct {
	db *sql.DB
}

// NewGamePlayRepository creates a new game play repository.
func NewGamePlayRepository(db *sql.DB) GamePlayRepository {
	return &gamePlayRepository{db: db}
}

// CreatePlay inserts a new game play record.
func (r *gamePlayRepository) CreatePlay(ctx context.Context, play *models.GamePlay) error {
	query := `
		INSERT INTO game_plays (
			game_id, match_id, turn_number, phase, step, player_type, action_type,
			card_id, card_name, zone_from, zone_to, timestamp, sequence_number
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	timestampStr := play.Timestamp.UTC().Format("2006-01-02 15:04:05.999999")

	result, err := r.db.ExecContext(ctx, query,
		play.GameID,
		play.MatchID,
		play.TurnNumber,
		play.Phase,
		play.Step,
		play.PlayerType,
		play.ActionType,
		play.CardID,
		play.CardName,
		play.ZoneFrom,
		play.ZoneTo,
		timestampStr,
		play.SequenceNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to create game play: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	play.ID = int(id)

	return nil
}

// CreatePlays inserts multiple game plays in a batch.
func (r *gamePlayRepository) CreatePlays(ctx context.Context, plays []*models.GamePlay) error {
	if len(plays) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO game_plays (
			game_id, match_id, turn_number, phase, step, player_type, action_type,
			card_id, card_name, zone_from, zone_to, timestamp, sequence_number
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, play := range plays {
		timestampStr := play.Timestamp.UTC().Format("2006-01-02 15:04:05.999999")
		result, err := stmt.ExecContext(ctx,
			play.GameID,
			play.MatchID,
			play.TurnNumber,
			play.Phase,
			play.Step,
			play.PlayerType,
			play.ActionType,
			play.CardID,
			play.CardName,
			play.ZoneFrom,
			play.ZoneTo,
			timestampStr,
			play.SequenceNumber,
		)
		if err != nil {
			return fmt.Errorf("failed to insert game play: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		play.ID = int(id)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateSnapshot inserts a new game state snapshot.
func (r *gamePlayRepository) CreateSnapshot(ctx context.Context, snapshot *models.GameStateSnapshot) error {
	query := `
		INSERT INTO game_state_snapshots (
			game_id, match_id, turn_number, active_player, player_life, opponent_life,
			player_cards_in_hand, opponent_cards_in_hand, player_lands_in_play,
			opponent_lands_in_play, board_state_json, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(game_id, turn_number) DO UPDATE SET
			active_player = excluded.active_player,
			player_life = excluded.player_life,
			opponent_life = excluded.opponent_life,
			player_cards_in_hand = excluded.player_cards_in_hand,
			opponent_cards_in_hand = excluded.opponent_cards_in_hand,
			player_lands_in_play = excluded.player_lands_in_play,
			opponent_lands_in_play = excluded.opponent_lands_in_play,
			board_state_json = excluded.board_state_json,
			timestamp = excluded.timestamp
	`

	timestampStr := snapshot.Timestamp.UTC().Format("2006-01-02 15:04:05.999999")

	_, err := r.db.ExecContext(ctx, query,
		snapshot.GameID,
		snapshot.MatchID,
		snapshot.TurnNumber,
		snapshot.ActivePlayer,
		snapshot.PlayerLife,
		snapshot.OpponentLife,
		snapshot.PlayerCardsInHand,
		snapshot.OpponentCardsInHand,
		snapshot.PlayerLandsInPlay,
		snapshot.OpponentLandsInPlay,
		snapshot.BoardStateJSON,
		timestampStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create game state snapshot: %w", err)
	}

	// Query for the actual ID since LastInsertId is unreliable for upserts
	var id int64
	err = r.db.QueryRowContext(ctx,
		`SELECT id FROM game_state_snapshots WHERE game_id = ? AND turn_number = ?`,
		snapshot.GameID, snapshot.TurnNumber,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to get snapshot id: %w", err)
	}
	snapshot.ID = int(id)

	return nil
}

// CreateSnapshots inserts multiple game state snapshots in a batch.
func (r *gamePlayRepository) CreateSnapshots(ctx context.Context, snapshots []*models.GameStateSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO game_state_snapshots (
			game_id, match_id, turn_number, active_player, player_life, opponent_life,
			player_cards_in_hand, opponent_cards_in_hand, player_lands_in_play,
			opponent_lands_in_play, board_state_json, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(game_id, turn_number) DO UPDATE SET
			active_player = excluded.active_player,
			player_life = excluded.player_life,
			opponent_life = excluded.opponent_life,
			player_cards_in_hand = excluded.player_cards_in_hand,
			opponent_cards_in_hand = excluded.opponent_cards_in_hand,
			player_lands_in_play = excluded.player_lands_in_play,
			opponent_lands_in_play = excluded.opponent_lands_in_play,
			board_state_json = excluded.board_state_json,
			timestamp = excluded.timestamp
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, snapshot := range snapshots {
		timestampStr := snapshot.Timestamp.UTC().Format("2006-01-02 15:04:05.999999")
		_, err := stmt.ExecContext(ctx,
			snapshot.GameID,
			snapshot.MatchID,
			snapshot.TurnNumber,
			snapshot.ActivePlayer,
			snapshot.PlayerLife,
			snapshot.OpponentLife,
			snapshot.PlayerCardsInHand,
			snapshot.OpponentCardsInHand,
			snapshot.PlayerLandsInPlay,
			snapshot.OpponentLandsInPlay,
			snapshot.BoardStateJSON,
			timestampStr,
		)
		if err != nil {
			return fmt.Errorf("failed to insert game state snapshot: %w", err)
		}

		// Query for the actual ID since LastInsertId is unreliable for upserts
		var id int64
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM game_state_snapshots WHERE game_id = ? AND turn_number = ?`,
			snapshot.GameID, snapshot.TurnNumber,
		).Scan(&id)
		if err != nil {
			return fmt.Errorf("failed to get snapshot id: %w", err)
		}
		snapshot.ID = int(id)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RecordOpponentCard records or updates an opponent card observation.
func (r *gamePlayRepository) RecordOpponentCard(ctx context.Context, card *models.OpponentCardObserved) error {
	query := `
		INSERT INTO opponent_cards_observed (
			game_id, match_id, card_id, card_name, zone_observed, turn_first_seen, times_seen
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(game_id, card_id) DO UPDATE SET
			times_seen = opponent_cards_observed.times_seen + 1,
			zone_observed = excluded.zone_observed
	`

	_, err := r.db.ExecContext(ctx, query,
		card.GameID,
		card.MatchID,
		card.CardID,
		card.CardName,
		card.ZoneObserved,
		card.TurnFirstSeen,
		card.TimesSeen,
	)
	if err != nil {
		return fmt.Errorf("failed to record opponent card: %w", err)
	}

	// Query for the actual ID since LastInsertId is unreliable for upserts
	var id int64
	err = r.db.QueryRowContext(ctx,
		`SELECT id FROM opponent_cards_observed WHERE game_id = ? AND card_id = ?`,
		card.GameID, card.CardID,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to get opponent card id: %w", err)
	}
	card.ID = int(id)

	return nil
}

// GetPlaysByMatch retrieves all plays for a match, ordered by sequence number.
func (r *gamePlayRepository) GetPlaysByMatch(ctx context.Context, matchID string) ([]*models.GamePlay, error) {
	query := `
		SELECT id, game_id, match_id, turn_number, phase, step, player_type, action_type,
		       card_id, card_name, zone_from, zone_to, timestamp, sequence_number, created_at
		FROM game_plays
		WHERE match_id = ?
		ORDER BY sequence_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plays by match: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanPlays(rows)
}

// GetPlaysByGame retrieves all plays for a specific game within a match.
func (r *gamePlayRepository) GetPlaysByGame(ctx context.Context, gameID int) ([]*models.GamePlay, error) {
	query := `
		SELECT id, game_id, match_id, turn_number, phase, step, player_type, action_type,
		       card_id, card_name, zone_from, zone_to, timestamp, sequence_number, created_at
		FROM game_plays
		WHERE game_id = ?
		ORDER BY sequence_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plays by game: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanPlays(rows)
}

// GetSnapshotsByMatch retrieves all snapshots for a match.
func (r *gamePlayRepository) GetSnapshotsByMatch(ctx context.Context, matchID string) ([]*models.GameStateSnapshot, error) {
	query := `
		SELECT id, game_id, match_id, turn_number, active_player, player_life, opponent_life,
		       player_cards_in_hand, opponent_cards_in_hand, player_lands_in_play,
		       opponent_lands_in_play, board_state_json, timestamp
		FROM game_state_snapshots
		WHERE match_id = ?
		ORDER BY game_id ASC, turn_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots by match: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanSnapshots(rows)
}

// GetSnapshotsByGame retrieves all snapshots for a specific game.
func (r *gamePlayRepository) GetSnapshotsByGame(ctx context.Context, gameID int) ([]*models.GameStateSnapshot, error) {
	query := `
		SELECT id, game_id, match_id, turn_number, active_player, player_life, opponent_life,
		       player_cards_in_hand, opponent_cards_in_hand, player_lands_in_play,
		       opponent_lands_in_play, board_state_json, timestamp
		FROM game_state_snapshots
		WHERE game_id = ?
		ORDER BY turn_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots by game: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanSnapshots(rows)
}

// GetOpponentCardsByMatch retrieves all opponent cards observed in a match.
func (r *gamePlayRepository) GetOpponentCardsByMatch(ctx context.Context, matchID string) ([]*models.OpponentCardObserved, error) {
	query := `
		SELECT id, game_id, match_id, card_id, card_name, zone_observed, turn_first_seen, times_seen
		FROM opponent_cards_observed
		WHERE match_id = ?
		ORDER BY turn_first_seen ASC
	`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get opponent cards by match: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanOpponentCards(rows)
}

// GetOpponentCardsByGame retrieves opponent cards observed in a specific game.
func (r *gamePlayRepository) GetOpponentCardsByGame(ctx context.Context, gameID int) ([]*models.OpponentCardObserved, error) {
	query := `
		SELECT id, game_id, match_id, card_id, card_name, zone_observed, turn_first_seen, times_seen
		FROM opponent_cards_observed
		WHERE game_id = ?
		ORDER BY turn_first_seen ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get opponent cards by game: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanOpponentCards(rows)
}

// GetPlayTimeline retrieves plays organized by turn/phase for timeline display.
func (r *gamePlayRepository) GetPlayTimeline(ctx context.Context, matchID string) ([]*models.PlayTimelineEntry, error) {
	// Get all plays for the match
	plays, err := r.GetPlaysByMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	// Get all snapshots for the match
	snapshots, err := r.GetSnapshotsByMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	// Build snapshot lookup by game_id and turn to avoid collisions across games
	type gameIDTurnKey struct {
		GameID int
		Turn   int
	}
	snapshotByGameTurn := make(map[gameIDTurnKey]*models.GameStateSnapshot)
	for _, snapshot := range snapshots {
		key := gameIDTurnKey{GameID: snapshot.GameID, Turn: snapshot.TurnNumber}
		snapshotByGameTurn[key] = snapshot
	}

	// Group plays by game_id, turn and phase
	type gameTurnPhaseKey struct {
		GameID int
		Turn   int
		Phase  string
	}
	playsByGameTurnPhase := make(map[gameTurnPhaseKey][]*models.GamePlay)
	var turnPhaseOrder []gameTurnPhaseKey

	for _, play := range plays {
		key := gameTurnPhaseKey{GameID: play.GameID, Turn: play.TurnNumber, Phase: play.Phase}
		if _, exists := playsByGameTurnPhase[key]; !exists {
			turnPhaseOrder = append(turnPhaseOrder, key)
		}
		playsByGameTurnPhase[key] = append(playsByGameTurnPhase[key], play)
	}

	// Build timeline entries
	var timeline []*models.PlayTimelineEntry
	for _, key := range turnPhaseOrder {
		snapshotKey := gameIDTurnKey{GameID: key.GameID, Turn: key.Turn}
		entry := &models.PlayTimelineEntry{
			Turn:     key.Turn,
			Phase:    key.Phase,
			Plays:    playsByGameTurnPhase[key],
			Snapshot: snapshotByGameTurn[snapshotKey],
		}
		timeline = append(timeline, entry)
	}

	return timeline, nil
}

// GetPlaySummary retrieves aggregated play statistics for a match.
func (r *gamePlayRepository) GetPlaySummary(ctx context.Context, matchID string) (*models.GamePlaySummary, error) {
	query := `
		SELECT
			match_id,
			COUNT(*) as total_plays,
			SUM(CASE WHEN player_type = 'player' THEN 1 ELSE 0 END) as player_plays,
			SUM(CASE WHEN player_type = 'opponent' THEN 1 ELSE 0 END) as opponent_plays,
			SUM(CASE WHEN action_type = 'play_card' THEN 1 ELSE 0 END) as card_plays,
			SUM(CASE WHEN action_type = 'attack' THEN 1 ELSE 0 END) as attacks,
			SUM(CASE WHEN action_type = 'block' THEN 1 ELSE 0 END) as blocks,
			SUM(CASE WHEN action_type = 'land_drop' THEN 1 ELSE 0 END) as land_drops,
			MAX(turn_number) as total_turns
		FROM game_plays
		WHERE match_id = ?
		GROUP BY match_id
	`

	summary := &models.GamePlaySummary{}
	err := r.db.QueryRowContext(ctx, query, matchID).Scan(
		&summary.MatchID,
		&summary.TotalPlays,
		&summary.PlayerPlays,
		&summary.OpponentPlays,
		&summary.CardPlays,
		&summary.Attacks,
		&summary.Blocks,
		&summary.LandDrops,
		&summary.TotalTurns,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get play summary: %w", err)
	}

	// Get opponent cards count separately
	opponentQuery := `
		SELECT COUNT(*) FROM opponent_cards_observed WHERE match_id = ?
	`
	err = r.db.QueryRowContext(ctx, opponentQuery, matchID).Scan(&summary.OpponentCardsSeen)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get opponent cards count: %w", err)
	}

	return summary, nil
}

// DeletePlaysByMatch removes all plays for a match.
func (r *gamePlayRepository) DeletePlaysByMatch(ctx context.Context, matchID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete plays
	_, err = tx.ExecContext(ctx, `DELETE FROM game_plays WHERE match_id = ?`, matchID)
	if err != nil {
		return fmt.Errorf("failed to delete plays: %w", err)
	}

	// Delete snapshots
	_, err = tx.ExecContext(ctx, `DELETE FROM game_state_snapshots WHERE match_id = ?`, matchID)
	if err != nil {
		return fmt.Errorf("failed to delete snapshots: %w", err)
	}

	// Delete opponent cards
	_, err = tx.ExecContext(ctx, `DELETE FROM opponent_cards_observed WHERE match_id = ?`, matchID)
	if err != nil {
		return fmt.Errorf("failed to delete opponent cards: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeletePlaysByGame removes all plays for a specific game.
func (r *gamePlayRepository) DeletePlaysByGame(ctx context.Context, gameID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete plays
	_, err = tx.ExecContext(ctx, `DELETE FROM game_plays WHERE game_id = ?`, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete plays: %w", err)
	}

	// Delete snapshots
	_, err = tx.ExecContext(ctx, `DELETE FROM game_state_snapshots WHERE game_id = ?`, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete snapshots: %w", err)
	}

	// Delete opponent cards
	_, err = tx.ExecContext(ctx, `DELETE FROM opponent_cards_observed WHERE game_id = ?`, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete opponent cards: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// scanPlays is a helper function to scan game plays from rows.
func (r *gamePlayRepository) scanPlays(rows *sql.Rows) ([]*models.GamePlay, error) {
	var plays []*models.GamePlay
	for rows.Next() {
		play := &models.GamePlay{}
		err := rows.Scan(
			&play.ID,
			&play.GameID,
			&play.MatchID,
			&play.TurnNumber,
			&play.Phase,
			&play.Step,
			&play.PlayerType,
			&play.ActionType,
			&play.CardID,
			&play.CardName,
			&play.ZoneFrom,
			&play.ZoneTo,
			&play.Timestamp,
			&play.SequenceNumber,
			&play.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game play: %w", err)
		}
		plays = append(plays, play)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game plays: %w", err)
	}

	return plays, nil
}

// scanSnapshots is a helper function to scan game state snapshots from rows.
func (r *gamePlayRepository) scanSnapshots(rows *sql.Rows) ([]*models.GameStateSnapshot, error) {
	var snapshots []*models.GameStateSnapshot
	for rows.Next() {
		snapshot := &models.GameStateSnapshot{}
		err := rows.Scan(
			&snapshot.ID,
			&snapshot.GameID,
			&snapshot.MatchID,
			&snapshot.TurnNumber,
			&snapshot.ActivePlayer,
			&snapshot.PlayerLife,
			&snapshot.OpponentLife,
			&snapshot.PlayerCardsInHand,
			&snapshot.OpponentCardsInHand,
			&snapshot.PlayerLandsInPlay,
			&snapshot.OpponentLandsInPlay,
			&snapshot.BoardStateJSON,
			&snapshot.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game state snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game state snapshots: %w", err)
	}

	return snapshots, nil
}

// scanOpponentCards is a helper function to scan opponent cards from rows.
func (r *gamePlayRepository) scanOpponentCards(rows *sql.Rows) ([]*models.OpponentCardObserved, error) {
	var cards []*models.OpponentCardObserved
	for rows.Next() {
		card := &models.OpponentCardObserved{}
		err := rows.Scan(
			&card.ID,
			&card.GameID,
			&card.MatchID,
			&card.CardID,
			&card.CardName,
			&card.ZoneObserved,
			&card.TurnFirstSeen,
			&card.TimesSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan opponent card: %w", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating opponent cards: %w", err)
	}

	return cards, nil
}

// BoardState represents the parsed board state from JSON.
type BoardState struct {
	PlayerPermanents   []PermanentState `json:"player_permanents"`
	OpponentPermanents []PermanentState `json:"opponent_permanents"`
}

// PermanentState represents a permanent on the battlefield.
type PermanentState struct {
	CardID    int    `json:"card_id"`
	CardName  string `json:"card_name"`
	IsTapped  bool   `json:"is_tapped"`
	Counters  int    `json:"counters,omitempty"`
	Attacking bool   `json:"attacking,omitempty"`
	Blocking  bool   `json:"blocking,omitempty"`
}

// ParseBoardState parses the board state JSON from a snapshot.
func ParseBoardState(jsonStr *string) (*BoardState, error) {
	if jsonStr == nil || *jsonStr == "" {
		return nil, nil
	}

	var state BoardState
	if err := json.Unmarshal([]byte(*jsonStr), &state); err != nil {
		return nil, fmt.Errorf("failed to parse board state: %w", err)
	}

	return &state, nil
}

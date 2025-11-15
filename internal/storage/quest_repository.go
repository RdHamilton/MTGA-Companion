package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// QuestRepository handles database operations for quests
type QuestRepository struct {
	db *sql.DB
}

// NewQuestRepository creates a new quest repository
func NewQuestRepository(db *sql.DB) *QuestRepository {
	return &QuestRepository{db: db}
}

// Save saves a quest to the database (insert or update)
func (r *QuestRepository) Save(quest *models.Quest) error {
	query := `
		INSERT INTO quests (
			quest_id, quest_type, goal, starting_progress, ending_progress,
			completed, can_swap, rewards, assigned_at, completed_at, rerolled
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(quest_id, assigned_at) DO UPDATE SET
			ending_progress = excluded.ending_progress,
			completed = excluded.completed,
			can_swap = excluded.can_swap,
			completed_at = excluded.completed_at,
			rerolled = excluded.rerolled
	`

	result, err := r.db.Exec(query,
		quest.QuestID, quest.QuestType, quest.Goal,
		quest.StartingProgress, quest.EndingProgress,
		quest.Completed, quest.CanSwap, quest.Rewards,
		quest.AssignedAt, quest.CompletedAt, quest.Rerolled,
	)
	if err != nil {
		return fmt.Errorf("failed to save quest: %w", err)
	}

	// Get the inserted ID if this was a new quest
	if quest.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			quest.ID = int(id)
		}
	}

	return nil
}

// GetActiveQuests returns all incomplete quests
func (r *QuestRepository) GetActiveQuests() ([]*models.Quest, error) {
	query := `
		SELECT id, quest_id, quest_type, goal, starting_progress, ending_progress,
		       completed, can_swap, rewards, assigned_at, completed_at, rerolled, created_at
		FROM quests
		WHERE completed = 0
		ORDER BY assigned_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active quests: %w", err)
	}
	defer rows.Close()

	return r.scanQuests(rows)
}

// GetQuestHistory returns quest history with optional filters
func (r *QuestRepository) GetQuestHistory(startDate, endDate *time.Time, limit int) ([]*models.Quest, error) {
	query := `
		SELECT id, quest_id, quest_type, goal, starting_progress, ending_progress,
		       completed, can_swap, rewards, assigned_at, completed_at, rerolled, created_at
		FROM quests
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != nil {
		query += " AND assigned_at >= ?"
		args = append(args, startDate)
	}

	if endDate != nil {
		query += " AND assigned_at <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY assigned_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get quest history: %w", err)
	}
	defer rows.Close()

	return r.scanQuests(rows)
}

// GetQuestStats returns analytics about quest completion
func (r *QuestRepository) GetQuestStats(startDate, endDate *time.Time) (*models.QuestStats, error) {
	stats := &models.QuestStats{}

	// Total and completed quests
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN completed = 0 THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN rerolled = 1 THEN 1 ELSE 0 END) as rerolled
		FROM quests
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != nil {
		query += " AND assigned_at >= ?"
		args = append(args, startDate)
	}

	if endDate != nil {
		query += " AND assigned_at <= ?"
		args = append(args, endDate)
	}

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalQuests, &stats.CompletedQuests, &stats.ActiveQuests, &stats.RerollCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get quest stats: %w", err)
	}

	// Calculate completion rate
	if stats.TotalQuests > 0 {
		stats.CompletionRate = float64(stats.CompletedQuests) / float64(stats.TotalQuests) * 100.0
	}

	// Average completion time
	query = `
		SELECT AVG(
			CAST((julianday(completed_at) - julianday(assigned_at)) * 86400000 AS INTEGER)
		)
		FROM quests
		WHERE completed = 1 AND completed_at IS NOT NULL
	`
	args = []interface{}{}

	if startDate != nil {
		query += " AND assigned_at >= ?"
		args = append(args, startDate)
	}

	if endDate != nil {
		query += " AND assigned_at <= ?"
		args = append(args, endDate)
	}

	var avgMS sql.NullInt64
	err = r.db.QueryRow(query, args...).Scan(&avgMS)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate average completion time: %w", err)
	}

	if avgMS.Valid {
		stats.AverageCompletionMS = avgMS.Int64
	}

	// TODO: Calculate total gold earned by parsing rewards JSON
	// For now, estimate based on completed quests (most quests give 500 or 750 gold)
	stats.TotalGoldEarned = stats.CompletedQuests * 500 // Conservative estimate

	return stats, nil
}

// GetQuestByID retrieves a quest by its database ID
func (r *QuestRepository) GetQuestByID(id int) (*models.Quest, error) {
	query := `
		SELECT id, quest_id, quest_type, goal, starting_progress, ending_progress,
		       completed, can_swap, rewards, assigned_at, completed_at, rerolled, created_at
		FROM quests
		WHERE id = ?
	`

	quest := &models.Quest{}
	var completedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&quest.ID, &quest.QuestID, &quest.QuestType, &quest.Goal,
		&quest.StartingProgress, &quest.EndingProgress, &quest.Completed,
		&quest.CanSwap, &quest.Rewards, &quest.AssignedAt,
		&completedAt, &quest.Rerolled, &quest.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quest not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}

	if completedAt.Valid {
		quest.CompletedAt = &completedAt.Time
	}

	return quest, nil
}

// MarkCompleted marks a quest as completed
func (r *QuestRepository) MarkCompleted(questID string, assignedAt time.Time, completedAt time.Time) error {
	query := `
		UPDATE quests
		SET completed = 1, completed_at = ?, ending_progress = goal
		WHERE quest_id = ? AND assigned_at = ?
	`

	_, err := r.db.Exec(query, completedAt, questID, assignedAt)
	if err != nil {
		return fmt.Errorf("failed to mark quest as completed: %w", err)
	}

	return nil
}

// MarkRerolled marks a quest as rerolled
func (r *QuestRepository) MarkRerolled(questID string, assignedAt time.Time) error {
	query := `
		UPDATE quests
		SET rerolled = 1
		WHERE quest_id = ? AND assigned_at = ?
	`

	_, err := r.db.Exec(query, questID, assignedAt)
	if err != nil {
		return fmt.Errorf("failed to mark quest as rerolled: %w", err)
	}

	return nil
}

// scanQuests is a helper to scan multiple quest rows
func (r *QuestRepository) scanQuests(rows *sql.Rows) ([]*models.Quest, error) {
	quests := []*models.Quest{}

	for rows.Next() {
		quest := &models.Quest{}
		var completedAt sql.NullTime

		err := rows.Scan(
			&quest.ID, &quest.QuestID, &quest.QuestType, &quest.Goal,
			&quest.StartingProgress, &quest.EndingProgress, &quest.Completed,
			&quest.CanSwap, &quest.Rewards, &quest.AssignedAt,
			&completedAt, &quest.Rerolled, &quest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quest: %w", err)
		}

		if completedAt.Valid {
			quest.CompletedAt = &completedAt.Time
		}

		quests = append(quests, quest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quest rows: %w", err)
	}

	return quests, nil
}

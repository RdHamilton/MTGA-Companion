package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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
	// First, check if a quest with this quest_id already exists
	existingQuery := `
		SELECT id, ending_progress, assigned_at FROM quests
		WHERE quest_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var existingID int
	var existingProgress int
	var existingAssignedAt time.Time
	err := r.db.QueryRow(existingQuery, quest.QuestID).Scan(&existingID, &existingProgress, &existingAssignedAt)

	if err == nil {
		// Quest exists - update it
		// Use the completion status from the parser (which detects completion via quest disappearance)
		// IMPORTANT: Preserve the original assigned_at timestamp for accurate duration calculation

		// Format timestamps for SQLite (ISO 8601 without timezone suffix)
		var completedAtStr *string
		if quest.CompletedAt != nil {
			formatted := quest.CompletedAt.UTC().Format("2006-01-02 15:04:05.999999")
			completedAtStr = &formatted
		}

		var lastSeenAtStr *string
		if quest.LastSeenAt != nil {
			formatted := quest.LastSeenAt.UTC().Format("2006-01-02 15:04:05.999999")
			lastSeenAtStr = &formatted
		}

		updateQuery := `
			UPDATE quests
			SET ending_progress = ?,
				completed = ?,
				completed_at = ?,
				last_seen_at = ?,
				can_swap = ?
			WHERE id = ?
		`

		_, err = r.db.Exec(updateQuery,
			quest.EndingProgress,
			quest.Completed,
			completedAtStr,
			lastSeenAtStr,
			quest.CanSwap,
			existingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update quest: %w", err)
		}

		quest.ID = existingID
		// Preserve the original assigned_at for accurate duration
		quest.AssignedAt = existingAssignedAt
		return nil
	}

	// Quest doesn't exist - insert it
	// Use the completion status from the parser (which detects completion via quest disappearance)

	query := `
		INSERT INTO quests (
			quest_id, quest_type, goal, starting_progress, ending_progress,
			completed, can_swap, rewards, assigned_at, completed_at, last_seen_at, rerolled
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Format timestamps for SQLite (ISO 8601 without timezone suffix)
	assignedAtStr := quest.AssignedAt.UTC().Format("2006-01-02 15:04:05.999999")
	var completedAtStr *string
	if quest.CompletedAt != nil {
		formatted := quest.CompletedAt.UTC().Format("2006-01-02 15:04:05.999999")
		completedAtStr = &formatted
	}

	var lastSeenAtStr *string
	if quest.LastSeenAt != nil {
		formatted := quest.LastSeenAt.UTC().Format("2006-01-02 15:04:05.999999")
		lastSeenAtStr = &formatted
	}

	result, err := r.db.Exec(query,
		quest.QuestID, quest.QuestType, quest.Goal,
		quest.StartingProgress, quest.EndingProgress,
		quest.Completed, quest.CanSwap, quest.Rewards,
		assignedAtStr, completedAtStr, lastSeenAtStr, quest.Rerolled,
	)
	if err != nil {
		return fmt.Errorf("failed to save quest: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err == nil {
		quest.ID = int(id)
	}

	return nil
}

// GetActiveQuests returns all incomplete, non-rerolled quests (one per unique quest_id).
// A quest is considered active if:
// - It has not been marked as completed (the parser detects completion via quest disappearance)
// - It has not been marked as rerolled (the parser detects rerolls via comparison with current MTGA state)
// - It has been seen in a QuestGetQuests response (last_seen_at is not null)
//
// Note: We no longer filter by last_seen_at timestamp because:
// - The daemon may not re-process old log entries on restart
// - Completion and reroll detection are the authoritative signals for quest state
// - If a quest is not completed and not rerolled, it should be considered active
func (r *QuestRepository) GetActiveQuests() ([]*models.Quest, error) {
	query := `
		SELECT q.id, q.quest_id, q.quest_type, q.goal, q.starting_progress, q.ending_progress,
		       q.completed, q.can_swap, q.rewards, q.assigned_at, q.completed_at, q.last_seen_at, q.rerolled, q.created_at
		FROM quests q
		INNER JOIN (
			SELECT quest_id, MAX(created_at) as max_created
			FROM quests
			WHERE completed = 0
			  AND rerolled = 0
			  AND last_seen_at IS NOT NULL
			GROUP BY quest_id
		) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		WHERE q.completed = 0
		  AND q.rerolled = 0
		  AND q.last_seen_at IS NOT NULL
		ORDER BY q.assigned_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active quests: %w", err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Ignore error on cleanup

	return r.scanQuests(rows)
}

// GetQuestHistory returns quest history with optional filters (one per unique quest_id)
func (r *QuestRepository) GetQuestHistory(startDate, endDate *time.Time, limit int) ([]*models.Quest, error) {
	query := `
		SELECT q.id, q.quest_id, q.quest_type, q.goal, q.starting_progress, q.ending_progress,
		       q.completed, q.can_swap, q.rewards, q.assigned_at, q.completed_at, q.last_seen_at, q.rerolled, q.created_at
		FROM quests q
		INNER JOIN (
			SELECT quest_id, MAX(created_at) as max_created
			FROM quests
			WHERE 1=1
	`
	args := []interface{}{}

	if startDate != nil {
		query += " AND DATE(created_at) >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}

	if endDate != nil {
		query += " AND DATE(created_at) <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}

	query += `
			GROUP BY quest_id
		) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		ORDER BY q.assigned_at DESC
	`

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get quest history: %w", err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Ignore error on cleanup

	return r.scanQuests(rows)
}

// GetQuestStats returns analytics about quest completion (deduplicates quests by quest_id)
func (r *QuestRepository) GetQuestStats(startDate, endDate *time.Time) (*models.QuestStats, error) {
	stats := &models.QuestStats{}

	// Get deduplicated quests (latest entry per quest_id)
	query := `
		WITH latest_quests AS (
			SELECT q.*
			FROM quests q
			INNER JOIN (
				SELECT quest_id, MAX(created_at) as max_created
				FROM quests
				WHERE 1=1
	`
	args := []interface{}{}

	if startDate != nil {
		query += " AND DATE(created_at) >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}

	if endDate != nil {
		query += " AND DATE(created_at) <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}

	query += `
				GROUP BY quest_id
			) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		)
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN completed = 0 THEN 1 ELSE 0 END), 0) as active,
			COALESCE(SUM(CASE WHEN rerolled = 1 THEN 1 ELSE 0 END), 0) as rerolled
		FROM latest_quests
	`

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

	// Average completion time (from deduplicated quests)
	query = `
		WITH latest_quests AS (
			SELECT q.*
			FROM quests q
			INNER JOIN (
				SELECT quest_id, MAX(created_at) as max_created
				FROM quests
				WHERE completed = 1 AND completed_at IS NOT NULL
	`
	args = []interface{}{}

	if startDate != nil {
		query += " AND DATE(created_at) >= DATE(?)"
		args = append(args, startDate)
	}

	if endDate != nil {
		query += " AND DATE(created_at) <= DATE(?)"
		args = append(args, endDate)
	}

	query += `
				GROUP BY quest_id
			) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		)
		SELECT AVG(
			CAST((julianday(completed_at) - julianday(assigned_at)) * 86400000 AS INTEGER)
		)
		FROM latest_quests
	`

	var avgMS sql.NullFloat64
	err = r.db.QueryRow(query, args...).Scan(&avgMS)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to calculate average completion time: %w", err)
	}

	if avgMS.Valid {
		stats.AverageCompletionMS = int64(avgMS.Float64)
	}

	// Calculate total gold earned by parsing rewards from completed quests
	stats.TotalGoldEarned = r.calculateTotalGoldEarned(startDate, endDate)

	return stats, nil
}

// calculateTotalGoldEarned sums the gold rewards from all completed quests.
// Falls back to estimate of 500 gold per quest if parsing fails.
func (r *QuestRepository) calculateTotalGoldEarned(startDate, endDate *time.Time) int {
	query := `
		WITH latest_quests AS (
			SELECT q.*
			FROM quests q
			INNER JOIN (
				SELECT quest_id, MAX(created_at) as max_created
				FROM quests
				WHERE completed = 1
	`
	args := []interface{}{}

	if startDate != nil {
		query += " AND DATE(created_at) >= ?"
		args = append(args, startDate.Format("2006-01-02"))
	}

	if endDate != nil {
		query += " AND DATE(created_at) <= ?"
		args = append(args, endDate.Format("2006-01-02"))
	}

	query += `
				GROUP BY quest_id
			) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		)
		SELECT COALESCE(rewards, '') as rewards
		FROM latest_quests
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		// Fall back to estimate on error
		return 0
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Ignore error on cleanup

	totalGold := 0

	for rows.Next() {
		var rewards string
		if err := rows.Scan(&rewards); err != nil {
			continue
		}

		gold := parseGoldFromRewards(rewards)
		totalGold += gold
	}

	return totalGold
}

// parseGoldFromRewards extracts the gold amount from a rewards string.
// The rewards field can be:
// - A numeric string like "500" or "750"
// - Empty string (defaults to 500)
// - Invalid data (defaults to 500)
func parseGoldFromRewards(rewards string) int {
	rewards = strings.TrimSpace(rewards)

	if rewards == "" {
		return 500 // Default estimate for missing data
	}

	// Try to parse as integer
	if gold, err := strconv.Atoi(rewards); err == nil && gold > 0 {
		return gold
	}

	// Fall back to conservative estimate
	return 500
}

// GetQuestByID retrieves a quest by its database ID
func (r *QuestRepository) GetQuestByID(id int) (*models.Quest, error) {
	query := `
		SELECT id, quest_id, quest_type, goal, starting_progress, ending_progress,
		       completed, can_swap, rewards, assigned_at, completed_at, last_seen_at, rerolled, created_at
		FROM quests
		WHERE id = ?
	`

	quest := &models.Quest{}
	var assignedAt string
	var completedAt sql.NullString
	var lastSeenAt sql.NullString
	var createdAt string

	err := r.db.QueryRow(query, id).Scan(
		&quest.ID, &quest.QuestID, &quest.QuestType, &quest.Goal,
		&quest.StartingProgress, &quest.EndingProgress, &quest.Completed,
		&quest.CanSwap, &quest.Rewards, &assignedAt,
		&completedAt, &lastSeenAt, &quest.Rerolled, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quest not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}

	// Parse assigned_at
	parsedAssignedAt, err := parseTimestamp(assignedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assigned_at: %w", err)
	}
	quest.AssignedAt = parsedAssignedAt

	// Parse completed_at if present
	if completedAt.Valid && completedAt.String != "" {
		parsedCompletedAt, err := parseTimestamp(completedAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse completed_at: %w", err)
		}
		quest.CompletedAt = &parsedCompletedAt
	}

	// Parse last_seen_at if present
	if lastSeenAt.Valid && lastSeenAt.String != "" {
		parsedLastSeenAt, err := parseTimestamp(lastSeenAt.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_seen_at: %w", err)
		}
		quest.LastSeenAt = &parsedLastSeenAt
	}

	// Parse created_at
	parsedCreatedAt, err := parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}
	quest.CreatedAt = parsedCreatedAt

	return quest, nil
}

// MarkCompleted marks a quest as completed
func (r *QuestRepository) MarkCompleted(questID string, assignedAt time.Time, completedAt time.Time) error {
	query := `
		UPDATE quests
		SET completed = 1, completed_at = ?, ending_progress = goal
		WHERE quest_id = ? AND assigned_at = ?
	`

	// Format timestamp for SQLite (ISO 8601 without timezone suffix)
	completedAtStr := completedAt.UTC().Format("2006-01-02 15:04:05.999999")
	assignedAtStr := assignedAt.UTC().Format("2006-01-02 15:04:05.999999")

	_, err := r.db.Exec(query, completedAtStr, questID, assignedAtStr)
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

	// Format timestamp for SQLite (ISO 8601 without timezone suffix)
	assignedAtStr := assignedAt.UTC().Format("2006-01-02 15:04:05.999999")

	_, err := r.db.Exec(query, questID, assignedAtStr)
	if err != nil {
		return fmt.Errorf("failed to mark quest as rerolled: %w", err)
	}

	return nil
}

// MarkQuestsCompletedByGraphState marks quests as completed based on GraphGetGraphState data.
// It maps Quest1-7 from the graph state to actual quests by their assigned order.
func (r *QuestRepository) MarkQuestsCompletedByGraphState(completedQuestNumbers map[int]bool, timestamp time.Time) error {
	// Get active quests ordered by assigned_at (oldest first)
	// This maps to Quest1-7 in the graph state
	query := `
		SELECT q.id, q.quest_id, q.quest_type, q.goal, q.starting_progress, q.ending_progress,
		       q.completed, q.can_swap, q.rewards, q.assigned_at, q.completed_at, q.last_seen_at, q.rerolled, q.created_at
		FROM quests q
		INNER JOIN (
			SELECT quest_id, MAX(created_at) as max_created
			FROM quests
			WHERE completed = 0
			GROUP BY quest_id
		) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		WHERE q.completed = 0
		ORDER BY q.assigned_at ASC
		LIMIT 7
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get active quests for completion: %w", err)
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // Ignore error on cleanup

	quests, err := r.scanQuests(rows)
	if err != nil {
		return fmt.Errorf("failed to scan quests for completion: %w", err)
	}

	// Mark quests as completed based on their position (Quest1 = oldest, Quest2 = 2nd oldest, etc.)
	for i, quest := range quests {
		questNumber := i + 1 // Quest1 = 1, Quest2 = 2, etc.

		// Check if this quest number is marked as completed in graph state
		if completedQuestNumbers[questNumber] {
			// Only mark as completed if it's not already completed
			if !quest.Completed {
				if err := r.MarkCompleted(quest.QuestID, quest.AssignedAt, timestamp); err != nil {
					return fmt.Errorf("failed to mark quest %d (%s) as completed: %w", questNumber, quest.QuestID, err)
				}
			}
		}
	}

	return nil
}

// MarkActiveQuestsCompleted marks all active quests that have reached their goal as completed.
// This is useful when we know quests should be completed but don't have specific quest IDs.
func (r *QuestRepository) MarkActiveQuestsCompleted(timestamp time.Time) error {
	query := `
		UPDATE quests
		SET completed = 1, completed_at = ?
		WHERE completed = 0
		  AND ending_progress >= goal
		  AND id IN (
			SELECT q.id
			FROM quests q
			INNER JOIN (
				SELECT quest_id, MAX(created_at) as max_created
				FROM quests
				WHERE completed = 0
				GROUP BY quest_id
			) latest ON q.quest_id = latest.quest_id AND q.created_at = latest.max_created
		  )
	`

	// Format timestamp for SQLite (ISO 8601 without timezone suffix)
	timestampStr := timestamp.UTC().Format("2006-01-02 15:04:05.999999")

	result, err := r.db.Exec(query, timestampStr)
	if err != nil {
		return fmt.Errorf("failed to mark active quests as completed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// Log how many quests were marked completed
		_ = rowsAffected // Prevent unused variable warning
	}

	return nil
}

// parseTimestamp attempts to parse a timestamp string in multiple formats
func parseTimestamp(s string) (time.Time, error) {
	// Trim any leading/trailing whitespace
	s = strings.TrimSpace(s)

	// Try RFC3339 with microseconds (e.g., "2025-11-23T06:01:35.548252Z")
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}

	// Try RFC3339 without microseconds (e.g., "2025-11-23T06:01:35Z")
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try SQLite format with fractional seconds (variable length 1-9 digits)
	// Go's time.Parse requires exact digit count, so try common lengths
	sqliteFormats := []string{
		"2006-01-02 15:04:05.999999999", // nanoseconds (9 digits)
		"2006-01-02 15:04:05.999999",    // microseconds (6 digits)
		"2006-01-02 15:04:05.99999",     // 5 digits
		"2006-01-02 15:04:05.9999",      // 4 digits
		"2006-01-02 15:04:05.999",       // milliseconds (3 digits)
		"2006-01-02 15:04:05.99",        // 2 digits
		"2006-01-02 15:04:05.9",         // 1 digit
	}

	for _, format := range sqliteFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Try SQLite format without fractional seconds (e.g., "2006-01-02 15:04:05")
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", s)
}

// scanQuests is a helper to scan multiple quest rows
func (r *QuestRepository) scanQuests(rows *sql.Rows) ([]*models.Quest, error) {
	quests := []*models.Quest{}

	for rows.Next() {
		quest := &models.Quest{}
		var assignedAt string
		var completedAt sql.NullString
		var lastSeenAt sql.NullString
		var createdAt string

		err := rows.Scan(
			&quest.ID, &quest.QuestID, &quest.QuestType, &quest.Goal,
			&quest.StartingProgress, &quest.EndingProgress, &quest.Completed,
			&quest.CanSwap, &quest.Rewards, &assignedAt,
			&completedAt, &lastSeenAt, &quest.Rerolled, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quest: %w", err)
		}

		// Parse assigned_at - try multiple formats
		parsedAssignedAt, err := parseTimestamp(assignedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse assigned_at: %w", err)
		}
		quest.AssignedAt = parsedAssignedAt

		// Parse completed_at if present
		if completedAt.Valid && completedAt.String != "" {
			parsedCompletedAt, err := parseTimestamp(completedAt.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse completed_at: %w", err)
			}
			quest.CompletedAt = &parsedCompletedAt
		}

		// Parse last_seen_at if present
		if lastSeenAt.Valid && lastSeenAt.String != "" {
			parsedLastSeenAt, err := parseTimestamp(lastSeenAt.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse last_seen_at: %w", err)
			}
			quest.LastSeenAt = &parsedLastSeenAt
		}

		// Parse created_at
		parsedCreatedAt, err := parseTimestamp(createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
		quest.CreatedAt = parsedCreatedAt

		quests = append(quests, quest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quest rows: %w", err)
	}

	return quests, nil
}

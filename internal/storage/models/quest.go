package models

import "time"

// Quest represents a daily quest in MTGA
type Quest struct {
	ID               int        `json:"id"`
	QuestID          string     `json:"quest_id"`
	QuestType        string     `json:"quest_type"`
	Goal             int        `json:"goal"`
	StartingProgress int        `json:"starting_progress"`
	EndingProgress   int        `json:"ending_progress"`
	Completed        bool       `json:"completed"`
	CanSwap          bool       `json:"can_swap"`
	Rewards          string     `json:"rewards"` // JSON string
	AssignedAt       time.Time  `json:"assigned_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty"` // Tracks when quest was last seen in QuestGetQuests response
	Rerolled         bool       `json:"rerolled"`
	CreatedAt        time.Time  `json:"created_at"`
}

// IsComplete returns whether the quest has been completed
func (q *Quest) IsComplete() bool {
	return q.EndingProgress >= q.Goal
}

// Progress returns the current progress as a percentage
func (q *Quest) Progress() float64 {
	if q.Goal == 0 {
		return 0.0
	}
	progress := float64(q.EndingProgress) / float64(q.Goal) * 100.0
	if progress > 100.0 {
		return 100.0
	}
	return progress
}

// QuestStats contains analytics about quest completion
type QuestStats struct {
	TotalQuests         int     `json:"total_quests"`
	CompletedQuests     int     `json:"completed_quests"`
	ActiveQuests        int     `json:"active_quests"`
	CompletionRate      float64 `json:"completion_rate"`
	TotalGoldEarned     int     `json:"total_gold_earned"`
	AverageCompletionMS int64   `json:"average_completion_ms"` // milliseconds
	RerollCount         int     `json:"reroll_count"`
}

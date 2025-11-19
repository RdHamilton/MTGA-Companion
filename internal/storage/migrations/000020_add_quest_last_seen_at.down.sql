-- Rollback: Remove last_seen_at column and index

DROP INDEX IF EXISTS idx_quests_last_seen_at;

-- SQLite doesn't support ALTER TABLE DROP COLUMN directly
-- We need to recreate the table without the column

-- Create temporary table with original schema
CREATE TABLE quests_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quest_id TEXT NOT NULL,
    quest_type TEXT NOT NULL,
    goal INTEGER NOT NULL,
    starting_progress INTEGER DEFAULT 0,
    ending_progress INTEGER DEFAULT 0,
    completed INTEGER DEFAULT 0,
    can_swap INTEGER DEFAULT 1,
    rewards TEXT,
    assigned_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    rerolled INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from current table
INSERT INTO quests_backup
SELECT id, quest_id, quest_type, goal, starting_progress, ending_progress,
       completed, can_swap, rewards, assigned_at, completed_at, rerolled, created_at
FROM quests;

-- Drop current table
DROP TABLE quests;

-- Rename backup to original name
ALTER TABLE quests_backup RENAME TO quests;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_quests_quest_id ON quests(quest_id);
CREATE INDEX IF NOT EXISTS idx_quests_completed ON quests(completed);
CREATE INDEX IF NOT EXISTS idx_quests_assigned_at ON quests(assigned_at);

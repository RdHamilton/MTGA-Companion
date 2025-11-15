-- Create quests table for tracking daily quest progress and completion
CREATE TABLE quests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quest_id TEXT NOT NULL,
    quest_type TEXT,
    goal INTEGER NOT NULL,
    starting_progress INTEGER NOT NULL DEFAULT 0,
    ending_progress INTEGER NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT 0,
    can_swap BOOLEAN NOT NULL DEFAULT 1,
    rewards TEXT, -- JSON string with reward details (e.g., chest description)
    assigned_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    rerolled BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quest_id, assigned_at)
);

-- Index for finding active (incomplete) quests
CREATE INDEX idx_quests_completed ON quests(completed);

-- Index for querying quests by date
CREATE INDEX idx_quests_assigned_at ON quests(assigned_at);

-- Index for completion statistics
CREATE INDEX idx_quests_completed_at ON quests(completed_at);

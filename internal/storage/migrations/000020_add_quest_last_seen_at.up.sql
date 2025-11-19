-- Add last_seen_at column to track when a quest was last seen in a QuestGetQuests response
-- This allows us to distinguish between truly active quests and stale quest data from historical logs

ALTER TABLE quests ADD COLUMN last_seen_at TIMESTAMP;

-- Initialize last_seen_at to created_at for existing quests
-- This assumes quests were seen when they were first created
UPDATE quests SET last_seen_at = created_at WHERE last_seen_at IS NULL;

-- Create index for efficient filtering by last_seen_at
CREATE INDEX IF NOT EXISTS idx_quests_last_seen_at ON quests(last_seen_at);

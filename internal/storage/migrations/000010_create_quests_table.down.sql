-- Drop quest indexes
DROP INDEX IF EXISTS idx_quests_completed_at;
DROP INDEX IF EXISTS idx_quests_assigned_at;
DROP INDEX IF EXISTS idx_quests_completed;

-- Drop quests table
DROP TABLE IF EXISTS quests;

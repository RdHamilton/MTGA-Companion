-- Drop improvement suggestions table
DROP INDEX IF EXISTS idx_suggestions_dismissed;
DROP INDEX IF EXISTS idx_suggestions_type;
DROP INDEX IF EXISTS idx_suggestions_deck_id;
DROP TABLE IF EXISTS improvement_suggestions;

-- Drop deck notes table
DROP INDEX IF EXISTS idx_deck_notes_category;
DROP INDEX IF EXISTS idx_deck_notes_deck_id;
DROP TABLE IF EXISTS deck_notes;

-- Remove notes and rating columns from matches table
-- SQLite doesn't support DROP COLUMN directly, need to recreate table
-- For simplicity in rollback, we'll leave the columns (they have defaults)
-- In production, would need to create new table without columns and migrate data

-- Remove deck creation tracking fields
DROP INDEX IF EXISTS idx_decks_app_created;
DROP INDEX IF EXISTS idx_decks_created_method;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
-- This is a simplified rollback - in practice you might want a full table rebuild
-- For now, we'll just leave the columns as they have defaults and won't affect functionality

-- Remove ML processing tracking from matches table

DROP INDEX IF EXISTS idx_matches_processed_for_ml;
DROP INDEX IF EXISTS idx_card_individual_stats_format;
DROP TABLE IF EXISTS card_individual_stats;

-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
-- This is handled by the migration framework

-- Remove pick quality analysis fields from draft_picks table
DROP INDEX IF EXISTS idx_draft_picks_quality_grade;

-- Note: SQLite doesn't support DROP COLUMN directly
-- We would need to recreate the table to remove columns
-- For now, just set columns to NULL in down migration

UPDATE draft_picks SET
    pick_quality_grade = NULL,
    pick_quality_rank = NULL,
    pack_best_gihwr = NULL,
    picked_card_gihwr = NULL,
    alternatives_json = NULL;

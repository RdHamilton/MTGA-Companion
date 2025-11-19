-- Remove draft grade indexes
DROP INDEX IF EXISTS idx_draft_sessions_score;
DROP INDEX IF EXISTS idx_draft_sessions_grade;

-- Note: SQLite doesn't support DROP COLUMN directly
-- Set columns to NULL in down migration

UPDATE draft_sessions SET
    overall_grade = NULL,
    overall_score = NULL,
    pick_quality_score = NULL,
    color_discipline_score = NULL,
    deck_composition_score = NULL,
    strategic_score = NULL;

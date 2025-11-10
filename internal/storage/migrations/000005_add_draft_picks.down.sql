-- Rollback draft_picks table
DROP INDEX IF EXISTS idx_draft_picks_selected_card;
DROP INDEX IF EXISTS idx_draft_picks_timestamp;
DROP INDEX IF EXISTS idx_draft_picks_event;
DROP TABLE IF EXISTS draft_picks;

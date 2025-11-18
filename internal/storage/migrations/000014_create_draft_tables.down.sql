-- Drop indexes first
DROP INDEX IF EXISTS idx_draft_sessions_status;
DROP INDEX IF EXISTS idx_draft_sessions_set_code;
DROP INDEX IF EXISTS idx_draft_sessions_start_time;
DROP INDEX IF EXISTS idx_draft_picks_session;
DROP INDEX IF EXISTS idx_draft_picks_timestamp;
DROP INDEX IF EXISTS idx_draft_packs_session;
DROP INDEX IF EXISTS idx_set_cards_arena_id;
DROP INDEX IF EXISTS idx_set_cards_set_code;
DROP INDEX IF EXISTS idx_set_cards_scryfall_id;

-- Drop tables
DROP TABLE IF EXISTS draft_picks;
DROP TABLE IF EXISTS draft_packs;
DROP TABLE IF EXISTS draft_sessions;
DROP TABLE IF EXISTS set_cards;

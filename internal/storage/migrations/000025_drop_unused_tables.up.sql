-- Drop unused tables that have been replaced or never populated
-- cards: Replaced by set_cards (migration 000014) which is populated by Scryfall fetcher
-- currency_history: Never implemented - was placeholder for future feature
-- draft_events: Replaced by draft_sessions (migration 000014)

-- Drop indices first
DROP INDEX IF EXISTS idx_cards_arena_id;
DROP INDEX IF EXISTS idx_cards_name;
DROP INDEX IF EXISTS idx_cards_set;
DROP INDEX IF EXISTS idx_cards_last_updated;
DROP INDEX IF EXISTS idx_draft_events_start_time;
DROP INDEX IF EXISTS idx_draft_events_status;

-- Drop the unused tables
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS currency_history;
DROP TABLE IF EXISTS draft_events;

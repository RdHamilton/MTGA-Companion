-- Rollback multi-account support
-- Removes account_id columns and accounts table

-- Remove account_id from draft_events
DROP INDEX IF EXISTS idx_draft_events_account_id;
ALTER TABLE draft_events DROP COLUMN account_id;

-- Remove account_id from rank_history
DROP INDEX IF EXISTS idx_rank_history_account_id;
ALTER TABLE rank_history DROP COLUMN account_id;

-- Remove account_id from collection_history
DROP INDEX IF EXISTS idx_collection_history_account_id;
ALTER TABLE collection_history DROP COLUMN account_id;

-- Recreate collection table without account_id
CREATE TABLE IF NOT EXISTS collection_old (
    card_id INTEGER PRIMARY KEY,
    quantity INTEGER NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO collection_old (card_id, quantity, updated_at)
SELECT card_id, quantity, updated_at FROM collection;
DROP TABLE collection;
ALTER TABLE collection_old RENAME TO collection;

-- Remove account_id from decks
DROP INDEX IF EXISTS idx_decks_account_id;
ALTER TABLE decks DROP COLUMN account_id;

-- Remove account_id from player_stats
DROP INDEX IF EXISTS idx_player_stats_account_id;
DROP INDEX IF EXISTS idx_player_stats_date_format_account;
CREATE UNIQUE INDEX IF NOT EXISTS idx_player_stats_date_format ON player_stats(date, format);
ALTER TABLE player_stats DROP COLUMN account_id;

-- Remove account_id from matches
DROP INDEX IF EXISTS idx_matches_account_id;
ALTER TABLE matches DROP COLUMN account_id;

-- Drop accounts table
DROP INDEX IF EXISTS idx_accounts_is_default;
DROP INDEX IF EXISTS idx_accounts_default;
DROP TABLE IF EXISTS accounts;


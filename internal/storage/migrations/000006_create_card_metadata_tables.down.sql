-- Rollback card metadata tables

-- Drop indices first
DROP INDEX IF EXISTS idx_sets_released_at;
DROP INDEX IF EXISTS idx_cards_last_updated;
DROP INDEX IF EXISTS idx_cards_set;
DROP INDEX IF EXISTS idx_cards_name;
DROP INDEX IF EXISTS idx_cards_arena_id;

-- Drop tables
DROP TABLE IF EXISTS sets;
DROP TABLE IF EXISTS cards;

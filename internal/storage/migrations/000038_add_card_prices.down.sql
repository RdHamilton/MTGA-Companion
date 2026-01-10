-- Remove price fields from set_cards table

DROP INDEX IF EXISTS idx_set_cards_prices;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
-- This is a destructive operation but acceptable for rollback

CREATE TABLE set_cards_backup AS SELECT
    id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors,
    rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
FROM set_cards;

DROP TABLE set_cards;

CREATE TABLE set_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    arena_id TEXT NOT NULL,
    scryfall_id TEXT,
    name TEXT NOT NULL,
    mana_cost TEXT,
    cmc INTEGER,
    types TEXT,
    colors TEXT,
    rarity TEXT,
    text TEXT,
    power TEXT,
    toughness TEXT,
    image_url TEXT,
    image_url_small TEXT,
    image_url_art TEXT,
    fetched_at TIMESTAMP,
    UNIQUE(set_code, arena_id)
);

INSERT INTO set_cards SELECT * FROM set_cards_backup;
DROP TABLE set_cards_backup;

-- Recreate indices
CREATE INDEX IF NOT EXISTS idx_set_cards_set_code ON set_cards(set_code);
CREATE INDEX IF NOT EXISTS idx_set_cards_arena_id ON set_cards(arena_id);
CREATE INDEX IF NOT EXISTS idx_set_cards_name ON set_cards(name);

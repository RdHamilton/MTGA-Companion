-- Rollback: Remove Standard legality tracking

-- Drop the standard_config table
DROP TABLE IF EXISTS standard_config;

-- Drop indexes
DROP INDEX IF EXISTS idx_sets_standard;
DROP INDEX IF EXISTS idx_set_cards_legalities;

-- SQLite doesn't support DROP COLUMN directly, so we recreate the tables

-- Recreate sets table without Standard columns
CREATE TABLE sets_temp (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    released_at TEXT,
    card_count INTEGER,
    set_type TEXT,
    icon_svg_uri TEXT,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO sets_temp (code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated)
SELECT code, name, released_at, card_count, set_type, icon_svg_uri, cached_at, last_updated
FROM sets;

DROP TABLE sets;
ALTER TABLE sets_temp RENAME TO sets;

-- Recreate index on sets
CREATE INDEX idx_sets_released_at ON sets(released_at);

-- Recreate set_cards table without legalities column
CREATE TABLE set_cards_temp (
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
    fetched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (set_code) REFERENCES sets(code) ON DELETE CASCADE,
    UNIQUE(set_code, arena_id)
);

INSERT INTO set_cards_temp (id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors, rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at)
SELECT id, set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors, rarity, text, power, toughness, image_url, image_url_small, image_url_art, fetched_at
FROM set_cards;

DROP TABLE set_cards;
ALTER TABLE set_cards_temp RENAME TO set_cards;

-- Recreate indexes on set_cards
CREATE INDEX idx_set_cards_arena_id ON set_cards(arena_id);
CREATE INDEX idx_set_cards_set_code ON set_cards(set_code);
CREATE INDEX idx_set_cards_scryfall_id ON set_cards(scryfall_id);

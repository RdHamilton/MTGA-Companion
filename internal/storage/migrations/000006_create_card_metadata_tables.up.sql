-- Card metadata cache tables for Scryfall data

-- Cards table: stores card metadata from Scryfall
CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,                    -- Scryfall ID
    arena_id INTEGER UNIQUE,                -- MTGA Arena ID (main lookup key)
    name TEXT NOT NULL,
    mana_cost TEXT,
    cmc REAL,
    type_line TEXT,
    oracle_text TEXT,
    colors TEXT,                            -- JSON array
    color_identity TEXT,                    -- JSON array
    rarity TEXT,
    set_code TEXT,
    collector_number TEXT,
    power TEXT,
    toughness TEXT,
    loyalty TEXT,
    image_uris TEXT,                        -- JSON object with all image sizes
    layout TEXT,                            -- normal, modal_dfc, transform, etc.
    card_faces TEXT,                        -- JSON for DFCs/MDFCs
    legalities TEXT,                        -- JSON object
    released_at TEXT,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indices for cards table
CREATE INDEX IF NOT EXISTS idx_cards_arena_id ON cards(arena_id);
CREATE INDEX IF NOT EXISTS idx_cards_name ON cards(name);
CREATE INDEX IF NOT EXISTS idx_cards_set ON cards(set_code);
CREATE INDEX IF NOT EXISTS idx_cards_last_updated ON cards(last_updated);

-- Sets table: stores set information from Scryfall
CREATE TABLE IF NOT EXISTS sets (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    released_at TEXT,
    card_count INTEGER,
    set_type TEXT,
    icon_svg_uri TEXT,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for sets by release date
CREATE INDEX IF NOT EXISTS idx_sets_released_at ON sets(released_at);

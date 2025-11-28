-- Recreate the dropped tables (for rollback)
-- Note: Data cannot be recovered, this just recreates the schema

-- Recreate cards table (from migration 000006)
CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    arena_id INTEGER UNIQUE,
    name TEXT NOT NULL,
    mana_cost TEXT,
    cmc REAL,
    type_line TEXT,
    oracle_text TEXT,
    colors TEXT,
    color_identity TEXT,
    rarity TEXT,
    set_code TEXT,
    collector_number TEXT,
    power TEXT,
    toughness TEXT,
    loyalty TEXT,
    image_uris TEXT,
    layout TEXT,
    card_faces TEXT,
    legalities TEXT,
    released_at TEXT,
    cached_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cards_arena_id ON cards(arena_id);
CREATE INDEX IF NOT EXISTS idx_cards_name ON cards(name);
CREATE INDEX IF NOT EXISTS idx_cards_set ON cards(set_code);
CREATE INDEX IF NOT EXISTS idx_cards_last_updated ON cards(last_updated);

-- Recreate currency_history table (from migration 000004)
CREATE TABLE IF NOT EXISTS currency_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    currency_type TEXT NOT NULL,
    amount INTEGER NOT NULL,
    delta INTEGER,
    source TEXT,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_currency_history_type ON currency_history(currency_type);
CREATE INDEX IF NOT EXISTS idx_currency_history_timestamp ON currency_history(timestamp);

-- Recreate draft_events table (from migration 000001)
CREATE TABLE IF NOT EXISTS draft_events (
    id TEXT PRIMARY KEY,
    event_name TEXT NOT NULL,
    set_code TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    wins INTEGER NOT NULL DEFAULT 0,
    losses INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL CHECK(status IN ('active', 'completed', 'abandoned')),
    deck_id TEXT,
    entry_fee TEXT,
    rewards TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (deck_id) REFERENCES decks(id)
);

CREATE INDEX IF NOT EXISTS idx_draft_events_start_time ON draft_events(start_time);
CREATE INDEX IF NOT EXISTS idx_draft_events_status ON draft_events(status);

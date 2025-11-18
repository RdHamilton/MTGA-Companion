-- Drop old draft tables if they exist (from migrations 000005 and earlier)
DROP TABLE IF EXISTS draft_picks;

-- Create draft_sessions table
CREATE TABLE IF NOT EXISTS draft_sessions (
    id TEXT PRIMARY KEY,
    event_name TEXT NOT NULL,
    set_code TEXT NOT NULL,
    draft_type TEXT DEFAULT 'quick_draft',
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    status TEXT DEFAULT 'in_progress',
    total_picks INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_draft_sessions_status ON draft_sessions(status);
CREATE INDEX idx_draft_sessions_set_code ON draft_sessions(set_code);
CREATE INDEX idx_draft_sessions_start_time ON draft_sessions(start_time DESC);

-- Create draft_picks table
CREATE TABLE IF NOT EXISTS draft_picks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    pack_number INTEGER NOT NULL,
    pick_number INTEGER NOT NULL,
    card_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (session_id) REFERENCES draft_sessions(id) ON DELETE CASCADE,
    UNIQUE(session_id, pack_number, pick_number)
);

CREATE INDEX idx_draft_picks_session ON draft_picks(session_id);
CREATE INDEX idx_draft_picks_timestamp ON draft_picks(timestamp);

-- Create draft_packs table
CREATE TABLE IF NOT EXISTS draft_packs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    pack_number INTEGER NOT NULL,
    pick_number INTEGER NOT NULL,
    card_ids TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (session_id) REFERENCES draft_sessions(id) ON DELETE CASCADE,
    UNIQUE(session_id, pack_number, pick_number)
);

CREATE INDEX idx_draft_packs_session ON draft_packs(session_id);

-- Create set_cards table for caching Scryfall data
CREATE TABLE IF NOT EXISTS set_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_code TEXT NOT NULL,
    arena_id TEXT NOT NULL,
    scryfall_id TEXT NOT NULL,
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
    UNIQUE(set_code, arena_id)
);

CREATE INDEX idx_set_cards_arena_id ON set_cards(arena_id);
CREATE INDEX idx_set_cards_set_code ON set_cards(set_code);
CREATE INDEX idx_set_cards_scryfall_id ON set_cards(scryfall_id);

-- Initial schema for MTGA Companion database
-- Creates core tables for matches, statistics, decks, and collection tracking

-- Matches table: stores match results and metadata
CREATE TABLE IF NOT EXISTS matches (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    event_name TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    duration_seconds INTEGER,
    player_wins INTEGER NOT NULL,
    opponent_wins INTEGER NOT NULL,
    player_team_id INTEGER NOT NULL,
    deck_id TEXT,
    rank_before TEXT,
    rank_after TEXT,
    format TEXT NOT NULL,
    result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
    result_reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_matches_timestamp ON matches(timestamp);
CREATE INDEX IF NOT EXISTS idx_matches_event_id ON matches(event_id);
CREATE INDEX IF NOT EXISTS idx_matches_format ON matches(format);
CREATE INDEX IF NOT EXISTS idx_matches_result ON matches(result);

-- Games table: stores individual game results within matches
CREATE TABLE IF NOT EXISTS games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    match_id TEXT NOT NULL,
    game_number INTEGER NOT NULL,
    result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
    duration_seconds INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
    UNIQUE(match_id, game_number)
);

CREATE INDEX IF NOT EXISTS idx_games_match_id ON games(match_id);

-- Player stats: aggregated statistics by date and format
CREATE TABLE IF NOT EXISTS player_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date DATE NOT NULL,
    format TEXT NOT NULL,
    matches_played INTEGER DEFAULT 0,
    matches_won INTEGER DEFAULT 0,
    games_played INTEGER DEFAULT 0,
    games_won INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, format)
);

CREATE INDEX IF NOT EXISTS idx_player_stats_date ON player_stats(date);
CREATE INDEX IF NOT EXISTS idx_player_stats_format ON player_stats(format);

-- Decks table: stores player deck lists
CREATE TABLE IF NOT EXISTS decks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    format TEXT NOT NULL,
    description TEXT,
    color_identity TEXT,
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    last_played DATETIME
);

CREATE INDEX IF NOT EXISTS idx_decks_format ON decks(format);
CREATE INDEX IF NOT EXISTS idx_decks_modified_at ON decks(modified_at);

-- Deck cards: stores cards within each deck
CREATE TABLE IF NOT EXISTS deck_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_id TEXT NOT NULL,
    card_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    board TEXT NOT NULL CHECK(board IN ('main', 'sideboard')),
    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
    UNIQUE(deck_id, card_id, board)
);

CREATE INDEX IF NOT EXISTS idx_deck_cards_deck_id ON deck_cards(deck_id);

-- Collection: stores player's card collection
CREATE TABLE IF NOT EXISTS collection (
    card_id INTEGER PRIMARY KEY,
    quantity INTEGER NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Collection history: tracks changes to collection over time
CREATE TABLE IF NOT EXISTS collection_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    card_id INTEGER NOT NULL,
    quantity_delta INTEGER NOT NULL,
    quantity_after INTEGER NOT NULL,
    timestamp DATETIME NOT NULL,
    source TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_collection_history_timestamp ON collection_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_collection_history_card_id ON collection_history(card_id);

-- Rank history: tracks rank progression over time
CREATE TABLE IF NOT EXISTS rank_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    format TEXT NOT NULL CHECK(format IN ('constructed', 'limited')),
    season_ordinal INTEGER NOT NULL,
    rank_class TEXT,
    rank_level INTEGER,
    rank_step INTEGER,
    percentile REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rank_history_timestamp ON rank_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_rank_history_format ON rank_history(format);
CREATE INDEX IF NOT EXISTS idx_rank_history_season ON rank_history(season_ordinal);

-- Draft events: stores draft/limited event records
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

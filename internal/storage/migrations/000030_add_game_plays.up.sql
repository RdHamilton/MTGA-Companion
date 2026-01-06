-- Migration: Add in-game play tracking for v1.4.1
-- Enables recording and analysis of every play made during a game

-- Game plays table: tracks individual card plays and actions
CREATE TABLE IF NOT EXISTS game_plays (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    match_id TEXT NOT NULL,
    turn_number INTEGER NOT NULL,
    phase TEXT,                       -- Main1, Combat, Main2, etc.
    step TEXT,                        -- BeginCombat, DeclareAttackers, etc.
    player_type TEXT NOT NULL,        -- 'player' or 'opponent'
    action_type TEXT NOT NULL,        -- 'play_card', 'attack', 'block', 'land_drop', 'mulligan'
    card_id INTEGER,                  -- Arena card ID
    card_name TEXT,                   -- Card name for display
    zone_from TEXT,                   -- Source zone (hand, library, graveyard, etc.)
    zone_to TEXT,                     -- Destination zone (battlefield, graveyard, etc.)
    timestamp TIMESTAMP NOT NULL,
    sequence_number INTEGER NOT NULL, -- Order within the game
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
);

-- Game state snapshots: captures board state at each turn
CREATE TABLE IF NOT EXISTS game_state_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    match_id TEXT NOT NULL,
    turn_number INTEGER NOT NULL,
    active_player TEXT NOT NULL,      -- 'player' or 'opponent'
    player_life INTEGER,
    opponent_life INTEGER,
    player_cards_in_hand INTEGER,
    opponent_cards_in_hand INTEGER,
    player_lands_in_play INTEGER,
    opponent_lands_in_play INTEGER,
    board_state_json TEXT,            -- JSON snapshot of all permanents
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
    UNIQUE(game_id, turn_number)
);

-- Opponent cards observed: tracks all cards revealed by opponent
CREATE TABLE IF NOT EXISTS opponent_cards_observed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL,
    match_id TEXT NOT NULL,
    card_id INTEGER NOT NULL,         -- Arena card ID
    card_name TEXT,
    zone_observed TEXT,               -- Where the card was seen (hand, battlefield, graveyard)
    turn_first_seen INTEGER,
    times_seen INTEGER DEFAULT 1,
    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
    UNIQUE(game_id, card_id)
);

-- Indexes for efficient queries
CREATE INDEX idx_game_plays_game_id ON game_plays(game_id);
CREATE INDEX idx_game_plays_match_id ON game_plays(match_id);
CREATE INDEX idx_game_plays_turn ON game_plays(game_id, turn_number);
CREATE INDEX idx_game_snapshots_game_id ON game_state_snapshots(game_id);
CREATE INDEX idx_game_snapshots_match_id ON game_state_snapshots(match_id);
CREATE INDEX idx_opponent_cards_game_id ON opponent_cards_observed(game_id);
CREATE INDEX idx_opponent_cards_match_id ON opponent_cards_observed(match_id);

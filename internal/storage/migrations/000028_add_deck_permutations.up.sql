-- Migration: Add deck permutation tracking for v1.4.1 Standard Play features
-- This enables tracking of deck modifications over time, allowing users to
-- see how their decks evolve and which versions perform best.

-- Table: deck_permutations
-- Records each unique version of a deck, tracking card changes and performance
CREATE TABLE deck_permutations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_id TEXT NOT NULL,
    parent_permutation_id INTEGER,          -- NULL for initial version, references parent for subsequent versions

    -- Card snapshot at this version (JSON array of {card_id, quantity, board})
    cards TEXT NOT NULL,
    -- Deterministic hash for detecting duplicate permutations (sorted by card_id, board)
    card_hash TEXT NOT NULL,

    -- Version metadata
    version_number INTEGER NOT NULL DEFAULT 1,
    version_name TEXT,                      -- Optional user-defined name like "Anti-Aggro Variant"
    change_summary TEXT,                    -- Auto-generated or user description of changes

    -- Performance tracking
    matches_played INTEGER NOT NULL DEFAULT 0,
    matches_won INTEGER NOT NULL DEFAULT 0,
    games_played INTEGER NOT NULL DEFAULT 0,
    games_won INTEGER NOT NULL DEFAULT 0,

    -- Timestamps
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_played_at DATETIME,

    FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_permutation_id) REFERENCES deck_permutations(id) ON DELETE SET NULL
);

-- Index for finding all permutations of a deck
CREATE INDEX idx_deck_permutations_deck_id ON deck_permutations(deck_id);

-- Index for navigating the version tree
CREATE INDEX idx_deck_permutations_parent ON deck_permutations(parent_permutation_id);

-- Index for finding most recent permutation
CREATE INDEX idx_deck_permutations_created ON deck_permutations(deck_id, created_at DESC);

-- Index for performance queries
CREATE INDEX idx_deck_permutations_win_rate ON deck_permutations(deck_id, matches_won, matches_played);

-- UNIQUE constraint to prevent duplicate permutations (enforces deduplication atomically)
CREATE UNIQUE INDEX idx_deck_permutations_hash ON deck_permutations(deck_id, card_hash);

-- Index for sorting by version_number in GetByDeckID and GetAllPerformance queries
CREATE INDEX idx_deck_permutations_version ON deck_permutations(deck_id, version_number);

-- Add current_permutation_id to decks table to track which version is currently active
ALTER TABLE decks ADD COLUMN current_permutation_id INTEGER REFERENCES deck_permutations(id);

-- Index for JOIN operations in GetCurrent query
CREATE INDEX idx_decks_current_permutation ON decks(current_permutation_id);

-- Create initial permutations for existing decks
-- This ensures existing decks have at least one permutation entry
INSERT INTO deck_permutations (deck_id, cards, card_hash, version_number, matches_played, matches_won, games_played, games_won, created_at, last_played_at)
SELECT
    d.id,
    COALESCE(
        (SELECT json_group_array(
            json_object(
                'card_id', dc.card_id,
                'quantity', dc.quantity,
                'board', dc.board
            )
        ) FROM deck_cards dc WHERE dc.deck_id = d.id ORDER BY dc.card_id, dc.board),
        '[]'
    ),
    -- Deterministic hash: sorted card_id:quantity:board pairs joined
    COALESCE(
        (SELECT group_concat(dc.card_id || ':' || dc.quantity || ':' || dc.board, '|')
         FROM deck_cards dc WHERE dc.deck_id = d.id ORDER BY dc.card_id, dc.board),
        ''
    ),
    1,
    d.matches_played,
    d.matches_won,
    d.games_played,
    d.games_won,
    d.created_at,
    d.last_played
FROM decks d;

-- Update decks to point to their initial permutation
UPDATE decks
SET current_permutation_id = (
    SELECT id FROM deck_permutations
    WHERE deck_permutations.deck_id = decks.id
    ORDER BY created_at ASC
    LIMIT 1
);

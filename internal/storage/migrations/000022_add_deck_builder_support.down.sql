-- Rollback v1.3 Deck Builder support

-- Drop deck_tags table
DROP INDEX IF EXISTS idx_deck_tags_tag;
DROP INDEX IF EXISTS idx_deck_tags_deck_id;
DROP TABLE IF EXISTS deck_tags;

-- Drop indexes (not dropping idx_decks_account_id as it was created by migration 000002)
DROP INDEX IF EXISTS idx_decks_draft_event_id;
DROP INDEX IF EXISTS idx_decks_source;

-- SQLite doesn't support DROP COLUMN in all versions, so we need to recreate the table
-- Create a temporary table with the schema from migration 000002 (includes account_id)
CREATE TABLE decks_old (
    id TEXT PRIMARY KEY,
    account_id INTEGER,
    name TEXT NOT NULL,
    format TEXT NOT NULL,
    description TEXT,
    color_identity TEXT,
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    last_played DATETIME
);

-- Copy data from the modified table to the old schema
INSERT INTO decks_old (id, account_id, name, format, description, color_identity, created_at, modified_at, last_played)
SELECT id, account_id, name, format, description, color_identity, created_at, modified_at, last_played
FROM decks;

-- Drop the modified table
DROP TABLE decks;

-- Rename the old table back
ALTER TABLE decks_old RENAME TO decks;

-- Recreate the original indexes
CREATE INDEX IF NOT EXISTS idx_decks_format ON decks(format);
CREATE INDEX IF NOT EXISTS idx_decks_modified_at ON decks(modified_at);
CREATE INDEX IF NOT EXISTS idx_decks_account_id ON decks(account_id);

-- Recreate deck_cards table without from_draft_pick
CREATE TABLE deck_cards_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_id TEXT NOT NULL,
    card_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    board TEXT NOT NULL CHECK(board IN ('main', 'sideboard')),
    UNIQUE(deck_id, card_id, board)
);

-- Copy data
INSERT INTO deck_cards_old (id, deck_id, card_id, quantity, board)
SELECT id, deck_id, card_id, quantity, board
FROM deck_cards;

-- Drop and rename
DROP TABLE deck_cards;
ALTER TABLE deck_cards_old RENAME TO deck_cards;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_deck_cards_deck_id ON deck_cards(deck_id);
CREATE INDEX IF NOT EXISTS idx_deck_cards_card_id ON deck_cards(card_id);

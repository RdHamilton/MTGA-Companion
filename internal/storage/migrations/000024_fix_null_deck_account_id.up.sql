-- Fix decks with NULL account_id by setting them to the default account (id=1)
-- This addresses issue #618 where decks parsed from logs before account setup
-- were stored with NULL account_id and became invisible in the UI
-- Also adds "arena" as a valid source for decks synced from MTGA logs

-- Update existing decks with NULL account_id to use default account
UPDATE decks SET account_id = 1 WHERE account_id IS NULL;

-- SQLite doesn't support modifying CHECK constraints directly
-- We need to recreate the table to add "arena" as a valid source

-- Step 1: Create new table with updated CHECK constraint
CREATE TABLE decks_new (
    id TEXT PRIMARY KEY,
    account_id INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL,
    format TEXT NOT NULL,
    description TEXT,
    color_identity TEXT,
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    last_played DATETIME,
    source TEXT NOT NULL DEFAULT 'constructed' CHECK(source IN ('draft', 'constructed', 'imported', 'arena')),
    draft_event_id TEXT,
    matches_played INTEGER NOT NULL DEFAULT 0,
    matches_won INTEGER NOT NULL DEFAULT 0,
    games_played INTEGER NOT NULL DEFAULT 0,
    games_won INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- Step 2: Copy data from old table
INSERT INTO decks_new (
    id, account_id, name, format, description, color_identity,
    created_at, modified_at, last_played, source, draft_event_id,
    matches_played, matches_won, games_played, games_won
)
SELECT
    id, COALESCE(account_id, 1), name, format, description, color_identity,
    created_at, modified_at, last_played, source, draft_event_id,
    matches_played, matches_won, games_played, games_won
FROM decks;

-- Step 3: Drop old table and rename new table
DROP TABLE decks;
ALTER TABLE decks_new RENAME TO decks;

-- Step 4: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_decks_account_id ON decks(account_id);
CREATE INDEX IF NOT EXISTS idx_decks_source ON decks(source);
CREATE INDEX IF NOT EXISTS idx_decks_draft_event_id ON decks(draft_event_id);

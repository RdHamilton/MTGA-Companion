-- Down migration for NULL account_id fix and arena source
-- Reverting to the original schema (without arena source)

-- Step 1: Create table with original CHECK constraint
CREATE TABLE decks_old (
    id TEXT PRIMARY KEY,
    account_id INTEGER,
    name TEXT NOT NULL,
    format TEXT NOT NULL,
    description TEXT,
    color_identity TEXT,
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    last_played DATETIME,
    source TEXT NOT NULL DEFAULT 'constructed' CHECK(source IN ('draft', 'constructed', 'imported')),
    draft_event_id TEXT,
    matches_played INTEGER NOT NULL DEFAULT 0,
    matches_won INTEGER NOT NULL DEFAULT 0,
    games_played INTEGER NOT NULL DEFAULT 0,
    games_won INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- Step 2: Copy data (convert arena to constructed)
INSERT INTO decks_old (
    id, account_id, name, format, description, color_identity,
    created_at, modified_at, last_played, source, draft_event_id,
    matches_played, matches_won, games_played, games_won
)
SELECT
    id, account_id, name, format, description, color_identity,
    created_at, modified_at, last_played,
    CASE WHEN source = 'arena' THEN 'constructed' ELSE source END,
    draft_event_id, matches_played, matches_won, games_played, games_won
FROM decks;

-- Step 3: Drop new table and rename old table
DROP TABLE decks;
ALTER TABLE decks_old RENAME TO decks;

-- Step 4: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_decks_account_id ON decks(account_id);
CREATE INDEX IF NOT EXISTS idx_decks_source ON decks(source);
CREATE INDEX IF NOT EXISTS idx_decks_draft_event_id ON decks(draft_event_id);

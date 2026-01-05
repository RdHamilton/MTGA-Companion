-- Rollback: Remove deck permutation tracking

-- Remove the current_permutation_id column from decks
-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table

-- Step 1: Create temporary table without current_permutation_id
CREATE TABLE decks_temp (
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

-- Step 2: Copy data
INSERT INTO decks_temp (
    id, account_id, name, format, description, color_identity,
    created_at, modified_at, last_played, source, draft_event_id,
    matches_played, matches_won, games_played, games_won
)
SELECT
    id, account_id, name, format, description, color_identity,
    created_at, modified_at, last_played, source, draft_event_id,
    matches_played, matches_won, games_played, games_won
FROM decks;

-- Step 3: Drop old table
DROP TABLE decks;

-- Step 4: Rename temp table
ALTER TABLE decks_temp RENAME TO decks;

-- Step 5: Recreate indexes on decks
CREATE INDEX idx_decks_format ON decks(format);
CREATE INDEX idx_decks_modified_at ON decks(modified_at);
CREATE INDEX idx_decks_account_id ON decks(account_id);
CREATE INDEX idx_decks_source ON decks(source);

-- Step 6: Drop permutations table
DROP TABLE IF EXISTS deck_permutations;

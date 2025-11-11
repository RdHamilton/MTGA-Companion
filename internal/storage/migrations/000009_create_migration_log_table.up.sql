-- Create migration_log table to track Scryfall card migrations
CREATE TABLE IF NOT EXISTS migration_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    migration_id TEXT NOT NULL UNIQUE,
    old_scryfall_id TEXT NOT NULL,
    new_scryfall_id TEXT,
    strategy TEXT NOT NULL CHECK(strategy IN ('merge', 'delete')),
    note TEXT,
    performed_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying by migration ID
CREATE INDEX IF NOT EXISTS idx_migration_log_migration_id ON migration_log(migration_id);

-- Index for querying by old Scryfall ID
CREATE INDEX IF NOT EXISTS idx_migration_log_old_scryfall_id ON migration_log(old_scryfall_id);

-- Index for querying by processed date
CREATE INDEX IF NOT EXISTS idx_migration_log_processed_at ON migration_log(processed_at);

-- Add inventory tracking tables for wildcards, currency, and vault progress

-- Current inventory snapshot
CREATE TABLE IF NOT EXISTS inventory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gold INTEGER NOT NULL DEFAULT 0,
    gems INTEGER NOT NULL DEFAULT 0,
    wc_common INTEGER NOT NULL DEFAULT 0,
    wc_uncommon INTEGER NOT NULL DEFAULT 0,
    wc_rare INTEGER NOT NULL DEFAULT 0,
    wc_mythic INTEGER NOT NULL DEFAULT 0,
    vault_progress REAL NOT NULL DEFAULT 0,
    draft_tokens INTEGER NOT NULL DEFAULT 0,
    sealed_tokens INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Inventory change history for tracking changes over time
CREATE TABLE IF NOT EXISTS inventory_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    field TEXT NOT NULL,
    previous_value INTEGER NOT NULL,
    new_value INTEGER NOT NULL,
    delta INTEGER NOT NULL,
    source TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient history queries
CREATE INDEX IF NOT EXISTS idx_inventory_history_field ON inventory_history(field);
CREATE INDEX IF NOT EXISTS idx_inventory_history_created_at ON inventory_history(created_at);

-- Insert default inventory record
INSERT INTO inventory (gold, gems, wc_common, wc_uncommon, wc_rare, wc_mythic, vault_progress, draft_tokens, sealed_tokens)
VALUES (0, 0, 0, 0, 0, 0, 0, 0, 0);

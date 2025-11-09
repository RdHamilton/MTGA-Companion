-- Add currency tracking history table
-- Tracks gems and gold changes over time with timestamps

CREATE TABLE IF NOT EXISTS currency_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    timestamp DATETIME NOT NULL,
    gems INTEGER NOT NULL,
    gold INTEGER NOT NULL,
    gems_delta INTEGER NOT NULL DEFAULT 0,
    gold_delta INTEGER NOT NULL DEFAULT 0,
    source TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_currency_history_account_id ON currency_history(account_id);
CREATE INDEX IF NOT EXISTS idx_currency_history_timestamp ON currency_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_currency_history_account_timestamp ON currency_history(account_id, timestamp);

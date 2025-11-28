-- Settings table for persisting user preferences
-- Uses key-value pairs with JSON values for flexibility
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,  -- JSON-encoded value
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT OR IGNORE INTO settings (key, value) VALUES
    ('autoRefresh', 'false'),
    ('refreshInterval', '30'),
    ('showNotifications', 'true'),
    ('theme', '"dark"'),
    ('daemonPort', '9999'),
    ('daemonMode', '"standalone"');

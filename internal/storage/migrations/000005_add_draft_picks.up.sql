-- Add draft picks table to store individual draft pick decisions
CREATE TABLE IF NOT EXISTS draft_picks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_event_id TEXT NOT NULL,
    pack_number INTEGER NOT NULL,
    pick_number INTEGER NOT NULL,
    available_cards TEXT NOT NULL,  -- JSON array of card IDs
    selected_card INTEGER NOT NULL,
    timestamp DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_event_id) REFERENCES draft_events(id) ON DELETE CASCADE,
    UNIQUE(draft_event_id, pack_number, pick_number)
);

CREATE INDEX IF NOT EXISTS idx_draft_picks_event ON draft_picks(draft_event_id);
CREATE INDEX IF NOT EXISTS idx_draft_picks_timestamp ON draft_picks(timestamp);
CREATE INDEX IF NOT EXISTS idx_draft_picks_selected_card ON draft_picks(selected_card);

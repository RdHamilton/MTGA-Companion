-- Migration: Add Standard format legality tracking for v1.4.1
-- This enables validation of Standard deck legality and rotation awareness.

-- Add Standard-specific fields to sets table
ALTER TABLE sets ADD COLUMN is_standard_legal BOOLEAN DEFAULT FALSE;
ALTER TABLE sets ADD COLUMN rotation_date TEXT;  -- Date when set rotates out of Standard (ISO 8601)

-- Create standard_config table for rotation configuration (singleton pattern)
CREATE TABLE IF NOT EXISTS standard_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Ensures only one row
    next_rotation_date TEXT NOT NULL,       -- e.g., "2027-01-23"
    rotation_enabled BOOLEAN DEFAULT TRUE,  -- Toggle rotation awareness (2026 has no rotation)
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initialize with known rotation info (January 2027 rotation)
INSERT INTO standard_config (id, next_rotation_date, rotation_enabled)
VALUES (1, '2027-01-23', TRUE);

-- Add legalities column to set_cards for individual card legality
-- JSON format: {"standard":"legal","historic":"legal","explorer":"legal",...}
ALTER TABLE set_cards ADD COLUMN legalities TEXT;

-- Index for efficient Standard queries
CREATE INDEX idx_sets_standard ON sets(is_standard_legal);

-- Index for card legality lookups
CREATE INDEX idx_set_cards_legalities ON set_cards(legalities) WHERE legalities IS NOT NULL;

-- Update current Standard-legal sets (as of January 2026)
-- These are the sets legal in Standard after the 2024 rotation
-- Sets rotate in cohorts based on release date

-- Foundations (FDN) - Legal until 2029
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2029-01-01' WHERE code = 'FDN';

-- Wilds of Eldraine (WOE) - Rotates January 2027
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2027-01-23' WHERE code = 'WOE';

-- Lost Caverns of Ixalan (LCI) - Rotates January 2027
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2027-01-23' WHERE code = 'LCI';

-- Murders at Karlov Manor (MKM) - Rotates January 2027
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2027-01-23' WHERE code = 'MKM';

-- Outlaws of Thunder Junction (OTJ) - Rotates January 2027
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2027-01-23' WHERE code = 'OTJ';

-- Bloomburrow (BLB) - Rotates January 2028
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2028-01-01' WHERE code = 'BLB';

-- Duskmourn: House of Horror (DSK) - Rotates January 2028
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2028-01-01' WHERE code = 'DSK';

-- Aetherdrift (AED) - Rotates January 2028 (releasing Q1 2025)
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2028-01-01' WHERE code = 'AED';

-- Tarkir: Dragonstorm (TLA) - Rotates January 2028 (releasing Q2 2025)
UPDATE sets SET is_standard_legal = TRUE, rotation_date = '2028-01-01' WHERE code = 'TLA';

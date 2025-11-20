-- Remove data_source columns from existing tables
-- Note: SQLite doesn't support DROP COLUMN directly, so we'll leave them

-- Drop dataset_metadata table
DROP TABLE IF EXISTS dataset_metadata;

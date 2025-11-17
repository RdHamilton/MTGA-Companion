-- Remove daily and weekly wins columns from accounts table
-- Note: SQLite doesn't support DROP COLUMN directly in older versions
-- We would need to recreate the table, but for development we'll document this limitation
-- ALTER TABLE accounts DROP COLUMN daily_wins;
-- ALTER TABLE accounts DROP COLUMN weekly_wins;

-- For SQLite < 3.35.0, use this approach:
-- 1. Create new table without these columns
-- 2. Copy data
-- 3. Drop old table
-- 4. Rename new table

-- For now, we'll just mark these columns as deprecated if rolling back
-- In practice, rolling back this migration requires manual intervention

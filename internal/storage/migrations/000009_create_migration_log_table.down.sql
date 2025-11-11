-- Drop migration_log table and indexes
DROP INDEX IF EXISTS idx_migration_log_processed_at;
DROP INDEX IF EXISTS idx_migration_log_old_scryfall_id;
DROP INDEX IF EXISTS idx_migration_log_migration_id;
DROP TABLE IF EXISTS migration_log;

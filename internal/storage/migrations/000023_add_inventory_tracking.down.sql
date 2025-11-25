-- Remove inventory tracking tables

DROP INDEX IF EXISTS idx_inventory_history_field;
DROP INDEX IF EXISTS idx_inventory_history_created_at;
DROP TABLE IF EXISTS inventory_history;
DROP TABLE IF EXISTS inventory;

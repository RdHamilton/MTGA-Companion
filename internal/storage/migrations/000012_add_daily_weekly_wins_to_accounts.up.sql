-- Add daily and weekly wins tracking to accounts table
ALTER TABLE accounts ADD COLUMN daily_wins INTEGER DEFAULT 0;
ALTER TABLE accounts ADD COLUMN weekly_wins INTEGER DEFAULT 0;

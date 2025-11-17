-- Add mastery pass tracking to accounts table
ALTER TABLE accounts ADD COLUMN mastery_level INTEGER DEFAULT 0;
ALTER TABLE accounts ADD COLUMN mastery_pass TEXT DEFAULT 'Basic';
ALTER TABLE accounts ADD COLUMN mastery_max INTEGER DEFAULT 80;

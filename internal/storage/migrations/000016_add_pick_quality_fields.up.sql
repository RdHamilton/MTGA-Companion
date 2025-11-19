-- Add pick quality analysis fields to draft_picks table
ALTER TABLE draft_picks ADD COLUMN pick_quality_grade TEXT;
ALTER TABLE draft_picks ADD COLUMN pick_quality_rank INTEGER;
ALTER TABLE draft_picks ADD COLUMN pack_best_gihwr REAL;
ALTER TABLE draft_picks ADD COLUMN picked_card_gihwr REAL;
ALTER TABLE draft_picks ADD COLUMN alternatives_json TEXT;

-- Create index for filtering by pick quality
CREATE INDEX idx_draft_picks_quality_grade ON draft_picks(pick_quality_grade);

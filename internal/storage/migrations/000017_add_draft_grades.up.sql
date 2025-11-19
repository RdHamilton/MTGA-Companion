-- Add draft grade fields to draft_sessions table
ALTER TABLE draft_sessions ADD COLUMN overall_grade TEXT;
ALTER TABLE draft_sessions ADD COLUMN overall_score INTEGER;
ALTER TABLE draft_sessions ADD COLUMN pick_quality_score REAL;
ALTER TABLE draft_sessions ADD COLUMN color_discipline_score REAL;
ALTER TABLE draft_sessions ADD COLUMN deck_composition_score REAL;
ALTER TABLE draft_sessions ADD COLUMN strategic_score REAL;

-- Create index for filtering by grade
CREATE INDEX idx_draft_sessions_grade ON draft_sessions(overall_grade);
CREATE INDEX idx_draft_sessions_score ON draft_sessions(overall_score DESC);

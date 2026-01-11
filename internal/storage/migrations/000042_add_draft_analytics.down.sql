-- Drop indexes first
DROP INDEX IF EXISTS idx_draft_match_results_session;
DROP INDEX IF EXISTS idx_draft_match_results_timestamp;
DROP INDEX IF EXISTS idx_draft_archetype_stats_set;
DROP INDEX IF EXISTS idx_draft_community_comparison_set;
DROP INDEX IF EXISTS idx_draft_temporal_trends_period;

-- Drop tables
DROP TABLE IF EXISTS draft_pattern_analysis;
DROP TABLE IF EXISTS draft_temporal_trends;
DROP TABLE IF EXISTS draft_community_comparison;
DROP TABLE IF EXISTS draft_archetype_stats;
DROP TABLE IF EXISTS draft_match_results;

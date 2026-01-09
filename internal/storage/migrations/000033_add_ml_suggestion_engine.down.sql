-- Rollback ML Suggestion Engine Schema

DROP TABLE IF EXISTS ml_model_metadata;
DROP TABLE IF EXISTS user_play_patterns;
DROP TABLE IF EXISTS card_affinity;
DROP TABLE IF EXISTS ml_suggestions;
DROP TABLE IF EXISTS card_combination_stats;

-- Rollback: Remove deck performance tracking tables

DROP INDEX IF EXISTS idx_archetype_card_weights_card;
DROP INDEX IF EXISTS idx_archetype_card_weights_archetype;
DROP TABLE IF EXISTS archetype_card_weights;

DROP INDEX IF EXISTS idx_rec_feedback_type_action;
DROP INDEX IF EXISTS idx_rec_feedback_card;
DROP INDEX IF EXISTS idx_rec_feedback_timestamp;
DROP INDEX IF EXISTS idx_rec_feedback_action;
DROP INDEX IF EXISTS idx_rec_feedback_type;
DROP INDEX IF EXISTS idx_rec_feedback_account;
DROP TABLE IF EXISTS recommendation_feedback;

DROP INDEX IF EXISTS idx_deck_archetypes_colors;
DROP INDEX IF EXISTS idx_deck_archetypes_format;
DROP INDEX IF EXISTS idx_deck_archetypes_set;
DROP TABLE IF EXISTS deck_archetypes;

DROP INDEX IF EXISTS idx_deck_perf_history_archetype_result;
DROP INDEX IF EXISTS idx_deck_perf_history_result;
DROP INDEX IF EXISTS idx_deck_perf_history_timestamp;
DROP INDEX IF EXISTS idx_deck_perf_history_format;
DROP INDEX IF EXISTS idx_deck_perf_history_archetype;
DROP INDEX IF EXISTS idx_deck_perf_history_deck;
DROP INDEX IF EXISTS idx_deck_perf_history_account;
DROP TABLE IF EXISTS deck_performance_history;

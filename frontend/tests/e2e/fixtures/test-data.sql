-- E2E Test Fixtures for MTGA Companion
-- This file contains test data for running E2E tests in CI environments
-- All data is fictional and for testing purposes only

-- ============================================================================
-- ACCOUNTS
-- ============================================================================
INSERT OR REPLACE INTO accounts (id, name, screen_name, client_id, is_default, daily_wins, weekly_wins, mastery_level, mastery_pass, mastery_max)
VALUES (1, 'TestPlayer', 'TestPlayer#12345', 'test-client-id-12345', 1, 3, 12, 45, 'Premium', 80);

-- ============================================================================
-- SETS (Recent sets for testing)
-- ============================================================================
INSERT OR REPLACE INTO sets (code, name, released_at, card_count, set_type, icon_svg_uri)
VALUES
    ('DSK', 'Duskmourn: House of Horror', '2024-09-27', 291, 'expansion', 'https://svgs.scryfall.io/sets/dsk.svg'),
    ('BLB', 'Bloomburrow', '2024-08-02', 276, 'expansion', 'https://svgs.scryfall.io/sets/blb.svg'),
    ('OTJ', 'Outlaws of Thunder Junction', '2024-04-19', 286, 'expansion', 'https://svgs.scryfall.io/sets/otj.svg');

-- ============================================================================
-- SET_CARDS (Sample cards from each set)
-- ============================================================================
-- DSK Cards
INSERT OR REPLACE INTO set_cards (set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors, rarity, text, power, toughness, image_url)
VALUES
    ('DSK', '90001', 'dsk-001', 'Fear of Missing Out', '{1}{R}', 2, 'Creature - Nightmare', 'R', 'uncommon', 'Haste. At the beginning of your end step, if you did not attack this turn, sacrifice Fear of Missing Out.', '3', '2', 'https://cards.scryfall.io/normal/front/dsk-001.jpg'),
    ('DSK', '90002', 'dsk-002', 'Reluctant Role Model', '{1}{W}', 2, 'Creature - Human Survivor', 'W', 'uncommon', 'Flying. When Reluctant Role Model enters the battlefield, create a 1/1 white Spirit token.', '2', '1', 'https://cards.scryfall.io/normal/front/dsk-002.jpg'),
    ('DSK', '90003', 'dsk-003', 'Doomsday Excruciator', '{5}{B}{B}', 7, 'Creature - Demon Horror', 'B', 'mythic', 'Flying, trample. When Doomsday Excruciator enters the battlefield, each opponent sacrifices half their creatures.', '6', '6', 'https://cards.scryfall.io/normal/front/dsk-003.jpg'),
    ('DSK', '90004', 'dsk-004', 'Enduring Curiosity', '{2}{U}{U}', 4, 'Creature - Cat Glimmer', 'U', 'rare', 'Flash. Whenever you draw a card except the first one you draw in each of your draw steps, put a +1/+1 counter on Enduring Curiosity.', '2', '2', 'https://cards.scryfall.io/normal/front/dsk-004.jpg'),
    ('DSK', '90005', 'dsk-005', 'Haunted Screen-Wall', '{2}', 2, 'Artifact', '', 'common', '{T}: Add {C}. {2}, {T}: Create a 1/1 colorless Spirit artifact creature token.', NULL, NULL, 'https://cards.scryfall.io/normal/front/dsk-005.jpg'),
    ('DSK', '90006', 'dsk-006', 'Vengeful Possession', '{2}{B}', 3, 'Enchantment - Aura', 'B', 'common', 'Enchant creature. Enchanted creature gets +2/+0 and has deathtouch.', NULL, NULL, 'https://cards.scryfall.io/normal/front/dsk-006.jpg'),
    ('DSK', '90007', 'dsk-007', 'Glimmer Seeker', '{1}{G}', 2, 'Creature - Human Scout', 'G', 'common', 'When Glimmer Seeker enters the battlefield, look at the top three cards of your library.', '2', '2', 'https://cards.scryfall.io/normal/front/dsk-007.jpg'),
    ('DSK', '90008', 'dsk-008', 'Clockwork Percussionist', '{R}', 1, 'Artifact Creature - Construct', 'R', 'common', 'Haste. Clockwork Percussionist gets +1/+0 for each other artifact you control.', '1', '1', 'https://cards.scryfall.io/normal/front/dsk-008.jpg'),
    ('DSK', '90009', 'dsk-009', 'Oblivion''s Hunger', '{1}{B}', 2, 'Instant', 'B', 'common', 'Target creature gets +2/+2 until end of turn. You lose 2 life.', NULL, NULL, 'https://cards.scryfall.io/normal/front/dsk-009.jpg'),
    ('DSK', '90010', 'dsk-010', 'Terror Tide', '{2}{U}{U}', 4, 'Sorcery', 'U', 'rare', 'Return all creatures to their owners'' hands.', NULL, NULL, 'https://cards.scryfall.io/normal/front/dsk-010.jpg');

-- BLB Cards
INSERT OR REPLACE INTO set_cards (set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors, rarity, text, power, toughness, image_url)
VALUES
    ('BLB', '91001', 'blb-001', 'Camellia the Seedmiser', '{2}{G}', 3, 'Legendary Creature - Mouse Warlock', 'G', 'rare', 'Whenever a creature you control deals combat damage, you gain 1 life and create a Food token.', '3', '3', 'https://cards.scryfall.io/normal/front/blb-001.jpg'),
    ('BLB', '91002', 'blb-002', 'Warren Warleader', '{1}{W}{W}', 3, 'Creature - Rabbit Soldier', 'W', 'uncommon', 'Whenever Warren Warleader attacks, create a 1/1 white Rabbit creature token.', '2', '3', 'https://cards.scryfall.io/normal/front/blb-002.jpg'),
    ('BLB', '91003', 'blb-003', 'Thornvault Forager', '{1}{G}', 2, 'Creature - Mouse Scout', 'G', 'common', 'When Thornvault Forager enters the battlefield, you may search your library for a basic land card.', '1', '1', 'https://cards.scryfall.io/normal/front/blb-003.jpg'),
    ('BLB', '91004', 'blb-004', 'Stormcatch Mentor', '{2}{U}', 3, 'Creature - Otter Wizard', 'U', 'uncommon', 'Prowess. Whenever you cast a noncreature spell, draw a card, then discard a card.', '2', '2', 'https://cards.scryfall.io/normal/front/blb-004.jpg'),
    ('BLB', '91005', 'blb-005', 'Blacksmith''s Talent', '{1}{R}', 2, 'Enchantment', 'R', 'uncommon', 'Equipped creatures you control get +1/+0.', NULL, NULL, 'https://cards.scryfall.io/normal/front/blb-005.jpg');

-- OTJ Cards
INSERT OR REPLACE INTO set_cards (set_code, arena_id, scryfall_id, name, mana_cost, cmc, types, colors, rarity, text, power, toughness, image_url)
VALUES
    ('OTJ', '92001', 'otj-001', 'Outlaw Stitcher', '{2}{B}', 3, 'Creature - Human Warlock', 'B', 'uncommon', 'When Outlaw Stitcher enters the battlefield, create a 2/2 black Zombie creature token.', '2', '3', 'https://cards.scryfall.io/normal/front/otj-001.jpg'),
    ('OTJ', '92002', 'otj-002', 'Slickshot Show-Off', '{1}{R}', 2, 'Creature - Human Rogue', 'R', 'rare', 'Flying, haste. Plot {1}{R}. Whenever you cast a spell, Slickshot Show-Off gets +2/+0 until end of turn.', '1', '2', 'https://cards.scryfall.io/normal/front/otj-002.jpg'),
    ('OTJ', '92003', 'otj-003', 'Dust Animus', '{4}{W}{W}', 6, 'Creature - Spirit', 'W', 'rare', 'Flying, vigilance. When Dust Animus enters the battlefield, exile target creature an opponent controls.', '4', '4', 'https://cards.scryfall.io/normal/front/otj-003.jpg'),
    ('OTJ', '92004', 'otj-004', 'Thunder Salvo', '{1}{R}', 2, 'Instant', 'R', 'common', 'Thunder Salvo deals 3 damage to target creature or planeswalker.', NULL, NULL, 'https://cards.scryfall.io/normal/front/otj-004.jpg'),
    ('OTJ', '92005', 'otj-005', 'Prosperity Tycoon', '{3}{G}', 4, 'Creature - Human Citizen', 'G', 'uncommon', 'When Prosperity Tycoon enters the battlefield, create two Treasure tokens.', '3', '4', 'https://cards.scryfall.io/normal/front/otj-005.jpg');

-- ============================================================================
-- COLLECTION (Cards owned by test account)
-- ============================================================================
INSERT OR REPLACE INTO collection (account_id, card_id, quantity)
VALUES
    -- DSK cards
    (1, 90001, 4), (1, 90002, 4), (1, 90003, 1), (1, 90004, 2), (1, 90005, 4),
    (1, 90006, 4), (1, 90007, 4), (1, 90008, 4), (1, 90009, 4), (1, 90010, 2),
    -- BLB cards
    (1, 91001, 2), (1, 91002, 4), (1, 91003, 4), (1, 91004, 3), (1, 91005, 4),
    -- OTJ cards
    (1, 92001, 4), (1, 92002, 3), (1, 92003, 1), (1, 92004, 4), (1, 92005, 4);

-- ============================================================================
-- DECKS
-- ============================================================================
INSERT OR REPLACE INTO decks (id, account_id, name, format, description, color_identity, created_at, modified_at, last_played, source, matches_played, matches_won, games_played, games_won)
VALUES
    ('deck-001', 1, 'Boros Aggro', 'Standard', 'Fast aggro deck with burn finish', 'WR', '2024-10-01 10:00:00', '2024-10-15 14:30:00', '2024-10-20 18:00:00', 'constructed', 25, 15, 55, 32),
    ('deck-002', 1, 'Dimir Control', 'Standard', 'Control deck with counterspells and removal', 'UB', '2024-09-15 09:00:00', '2024-10-10 11:00:00', '2024-10-18 20:00:00', 'constructed', 18, 11, 42, 26),
    ('deck-003', 1, 'Gruul Stompy', 'Historic', 'Big creatures and ramp', 'RG', '2024-08-20 15:00:00', '2024-09-05 16:00:00', '2024-10-12 19:00:00', 'constructed', 12, 8, 28, 18),
    ('deck-004', 1, 'DSK Draft Deck', 'Limited', 'Quick draft deck from Duskmourn', 'WB', '2024-10-20 14:00:00', '2024-10-20 15:30:00', '2024-10-20 17:00:00', 'draft', 3, 2, 7, 5),
    ('deck-005', 1, 'Mono Red Burn', 'Standard', 'All-in burn strategy', 'R', '2024-10-05 12:00:00', '2024-10-08 13:00:00', '2024-10-15 21:00:00', 'constructed', 8, 5, 18, 11);

-- ============================================================================
-- DECK_CARDS
-- ============================================================================
-- Boros Aggro (deck-001)
INSERT OR REPLACE INTO deck_cards (deck_id, card_id, quantity, board)
VALUES
    ('deck-001', 90001, 4, 'main'), ('deck-001', 90002, 4, 'main'), ('deck-001', 90008, 4, 'main'),
    ('deck-001', 92002, 3, 'main'), ('deck-001', 92004, 4, 'main'), ('deck-001', 91002, 4, 'main'),
    ('deck-001', 91005, 2, 'sideboard');

-- Dimir Control (deck-002)
INSERT OR REPLACE INTO deck_cards (deck_id, card_id, quantity, board)
VALUES
    ('deck-002', 90003, 2, 'main'), ('deck-002', 90004, 4, 'main'), ('deck-002', 90006, 4, 'main'),
    ('deck-002', 90009, 4, 'main'), ('deck-002', 90010, 3, 'main'), ('deck-002', 92001, 4, 'main');

-- DSK Draft Deck (deck-004)
INSERT OR REPLACE INTO deck_cards (deck_id, card_id, quantity, board, from_draft_pick)
VALUES
    ('deck-004', 90002, 2, 'main', 1), ('deck-004', 90003, 1, 'main', 1), ('deck-004', 90006, 3, 'main', 1),
    ('deck-004', 90005, 2, 'main', 1), ('deck-004', 90009, 2, 'main', 1);

-- ============================================================================
-- MATCHES
-- ============================================================================
INSERT OR REPLACE INTO matches (id, event_id, event_name, timestamp, duration_seconds, player_wins, opponent_wins, player_team_id, deck_id, rank_before, rank_after, format, result, result_reason, account_id, opponent_name, opponent_id)
VALUES
    -- Standard Ranked matches
    ('match-001', 'event-std-001', 'Ranked Standard', '2024-10-20 18:30:00', 1200, 2, 1, 1, 'deck-001', 'Gold-2', 'Gold-1', 'Standard', 'win', 'OpponentConceded', 1, 'Opponent_A', 'opp-001'),
    ('match-002', 'event-std-002', 'Ranked Standard', '2024-10-20 17:00:00', 900, 0, 2, 1, 'deck-001', 'Gold-3', 'Gold-2', 'Standard', 'loss', 'PlayerLost', 1, 'Opponent_B', 'opp-002'),
    ('match-003', 'event-std-003', 'Ranked Standard', '2024-10-19 20:00:00', 1500, 2, 0, 1, 'deck-002', 'Gold-1', 'Platinum-4', 'Standard', 'win', 'OpponentConceded', 1, 'Opponent_C', 'opp-003'),
    ('match-004', 'event-std-004', 'Ranked Standard', '2024-10-19 18:00:00', 800, 2, 1, 1, 'deck-002', 'Platinum-4', 'Platinum-3', 'Standard', 'win', 'OpponentConceded', 1, 'Opponent_D', 'opp-004'),
    ('match-005', 'event-std-005', 'Ranked Standard', '2024-10-18 21:00:00', 1100, 1, 2, 1, 'deck-005', 'Platinum-3', 'Platinum-4', 'Standard', 'loss', 'PlayerLost', 1, 'Opponent_E', 'opp-005'),

    -- Historic matches
    ('match-006', 'event-hist-001', 'Ranked Historic', '2024-10-17 19:00:00', 1400, 2, 0, 1, 'deck-003', 'Diamond-2', 'Diamond-1', 'Historic', 'win', 'OpponentConceded', 1, 'Opponent_F', 'opp-006'),
    ('match-007', 'event-hist-002', 'Ranked Historic', '2024-10-16 20:00:00', 950, 2, 1, 1, 'deck-003', 'Diamond-3', 'Diamond-2', 'Historic', 'win', 'OpponentConceded', 1, 'Opponent_G', 'opp-007'),

    -- Quick Draft matches
    ('match-008', 'event-draft-001', 'QuickDraft_DSK', '2024-10-20 16:00:00', 1300, 2, 0, 1, 'deck-004', NULL, NULL, 'QuickDraft_DSK', 'win', 'OpponentConceded', 1, 'Draft_Opp_A', 'opp-008'),
    ('match-009', 'event-draft-001', 'QuickDraft_DSK', '2024-10-20 15:00:00', 1100, 2, 1, 1, 'deck-004', NULL, NULL, 'QuickDraft_DSK', 'win', 'OpponentConceded', 1, 'Draft_Opp_B', 'opp-009'),
    ('match-010', 'event-draft-001', 'QuickDraft_DSK', '2024-10-20 14:30:00', 750, 0, 2, 1, 'deck-004', NULL, NULL, 'QuickDraft_DSK', 'loss', 'PlayerLost', 1, 'Draft_Opp_C', 'opp-010'),

    -- More Standard for variety
    ('match-011', 'event-std-006', 'Ranked Standard', '2024-10-15 19:00:00', 1050, 2, 0, 1, 'deck-001', 'Gold-4', 'Gold-3', 'Standard', 'win', 'OpponentConceded', 1, 'Opponent_H', 'opp-011'),
    ('match-012', 'event-std-007', 'Ranked Standard', '2024-10-14 18:30:00', 1250, 1, 2, 1, 'deck-002', 'Gold-3', 'Gold-4', 'Standard', 'loss', 'PlayerLost', 1, 'Opponent_I', 'opp-012');

-- ============================================================================
-- GAMES (Individual games within matches)
-- ============================================================================
INSERT OR REPLACE INTO games (match_id, game_number, result, duration_seconds, result_reason)
VALUES
    -- match-001 (2-1 win)
    ('match-001', 1, 'win', 400, NULL),
    ('match-001', 2, 'loss', 350, NULL),
    ('match-001', 3, 'win', 450, 'OpponentConceded'),
    -- match-002 (0-2 loss)
    ('match-002', 1, 'loss', 450, NULL),
    ('match-002', 2, 'loss', 450, NULL),
    -- match-003 (2-0 win)
    ('match-003', 1, 'win', 750, NULL),
    ('match-003', 2, 'win', 750, 'OpponentConceded'),
    -- match-004 (2-1 win)
    ('match-004', 1, 'win', 250, NULL),
    ('match-004', 2, 'loss', 300, NULL),
    ('match-004', 3, 'win', 250, 'OpponentConceded'),
    -- match-005 (1-2 loss)
    ('match-005', 1, 'loss', 400, NULL),
    ('match-005', 2, 'win', 350, NULL),
    ('match-005', 3, 'loss', 350, NULL);

-- ============================================================================
-- QUESTS
-- ============================================================================
INSERT OR REPLACE INTO quests (quest_id, quest_type, goal, starting_progress, ending_progress, completed, can_swap, rewards, assigned_at, completed_at, last_seen_at)
VALUES
    ('quest-daily-001', 'Cast 20 White or Blue spells', 20, 0, 8, 0, 1, '{"gold": 500, "xp": 500}', '2024-10-20 00:00:00', NULL, '2024-10-20 18:00:00'),
    ('quest-daily-002', 'Win 5 games', 5, 0, 3, 0, 1, '{"gold": 750, "xp": 500}', '2024-10-20 00:00:00', NULL, '2024-10-20 18:00:00'),
    ('quest-daily-003', 'Play 30 lands', 30, 0, 30, 1, 0, '{"gold": 500, "xp": 500}', '2024-10-19 00:00:00', '2024-10-19 22:00:00', '2024-10-20 18:00:00');

-- ============================================================================
-- DRAFT SESSIONS
-- ============================================================================
INSERT OR REPLACE INTO draft_sessions (id, event_name, set_code, draft_type, start_time, end_time, status, total_picks, overall_grade, overall_score, pick_quality_score, color_discipline_score)
VALUES
    ('draft-001', 'QuickDraft_DSK', 'DSK', 'quick_draft', '2024-10-20 13:00:00', '2024-10-20 13:45:00', 'completed', 45, 'B+', 82, 78.5, 85.0),
    ('draft-002', 'QuickDraft_BLB', 'BLB', 'quick_draft', '2024-10-15 14:00:00', '2024-10-15 14:30:00', 'completed', 45, 'A-', 88, 85.0, 90.0);

-- ============================================================================
-- DRAFT PICKS (Sample picks from draft-001)
-- ============================================================================
INSERT OR REPLACE INTO draft_picks (session_id, pack_number, pick_number, card_id, timestamp, pick_quality_grade, pick_quality_rank, pack_best_gihwr, picked_card_gihwr)
VALUES
    -- Pack 1
    ('draft-001', 1, 1, '90003', '2024-10-20 13:00:30', 'A', 1, 58.5, 58.5),
    ('draft-001', 1, 2, '90002', '2024-10-20 13:01:00', 'B+', 2, 56.0, 55.0),
    ('draft-001', 1, 3, '90006', '2024-10-20 13:01:30', 'B', 1, 54.0, 54.0),
    ('draft-001', 1, 4, '90009', '2024-10-20 13:02:00', 'B', 1, 53.0, 53.0),
    ('draft-001', 1, 5, '90005', '2024-10-20 13:02:30', 'C+', 3, 52.0, 50.0),
    -- Pack 2
    ('draft-001', 2, 1, '90004', '2024-10-20 13:15:00', 'B+', 1, 57.0, 57.0),
    ('draft-001', 2, 2, '90002', '2024-10-20 13:15:30', 'A-', 1, 56.5, 56.5),
    ('draft-001', 2, 3, '90006', '2024-10-20 13:16:00', 'B', 2, 54.5, 53.0),
    -- Pack 3
    ('draft-001', 3, 1, '90010', '2024-10-20 13:30:00', 'A', 1, 59.0, 59.0),
    ('draft-001', 3, 2, '90009', '2024-10-20 13:30:30', 'B', 1, 53.5, 53.5);

-- ============================================================================
-- DRAFT CARD RATINGS (17Lands-style ratings for draft picks)
-- ============================================================================
INSERT OR REPLACE INTO draft_card_ratings (set_code, draft_format, arena_id, name, color, rarity, gihwr, ohwr, alsa, ata, gih_count)
VALUES
    ('DSK', 'QuickDraft', 90001, 'Fear of Missing Out', 'R', 'uncommon', 54.5, 52.0, 4.5, 3.2, 15000),
    ('DSK', 'QuickDraft', 90002, 'Reluctant Role Model', 'W', 'uncommon', 56.0, 54.0, 3.8, 2.9, 18000),
    ('DSK', 'QuickDraft', 90003, 'Doomsday Excruciator', 'B', 'mythic', 58.5, 55.0, 1.5, 1.2, 5000),
    ('DSK', 'QuickDraft', 90004, 'Enduring Curiosity', 'U', 'rare', 57.0, 54.5, 2.0, 1.8, 8000),
    ('DSK', 'QuickDraft', 90005, 'Haunted Screen-Wall', '', 'common', 50.0, 48.0, 8.5, 7.0, 25000),
    ('DSK', 'QuickDraft', 90006, 'Vengeful Possession', 'B', 'common', 54.0, 52.0, 5.0, 4.2, 22000),
    ('DSK', 'QuickDraft', 90007, 'Glimmer Seeker', 'G', 'common', 52.5, 51.0, 6.0, 5.5, 20000),
    ('DSK', 'QuickDraft', 90008, 'Clockwork Percussionist', 'R', 'common', 51.0, 49.5, 7.0, 6.0, 18000),
    ('DSK', 'QuickDraft', 90009, 'Oblivion''s Hunger', 'B', 'common', 53.0, 51.5, 5.5, 4.8, 21000),
    ('DSK', 'QuickDraft', 90010, 'Terror Tide', 'U', 'rare', 59.0, 56.0, 1.8, 1.5, 7000);

-- ============================================================================
-- RANK HISTORY
-- ============================================================================
INSERT OR REPLACE INTO rank_history (timestamp, format, season_ordinal, rank_class, rank_level, rank_step, percentile, account_id)
VALUES
    ('2024-10-20 18:30:00', 'constructed', 24, 'Gold', 1, 2, 45.0, 1),
    ('2024-10-19 20:00:00', 'constructed', 24, 'Gold', 2, 1, 42.0, 1),
    ('2024-10-18 21:00:00', 'constructed', 24, 'Gold', 3, 3, 40.0, 1),
    ('2024-10-17 19:00:00', 'limited', 24, 'Diamond', 1, 2, 75.0, 1),
    ('2024-10-16 20:00:00', 'limited', 24, 'Diamond', 2, 1, 72.0, 1);

-- ============================================================================
-- INVENTORY
-- ============================================================================
INSERT OR REPLACE INTO inventory (gold, gems, wc_common, wc_uncommon, wc_rare, wc_mythic, vault_progress, draft_tokens, sealed_tokens)
VALUES (15000, 3500, 45, 32, 12, 4, 65.5, 2, 0);

-- ============================================================================
-- SETTINGS
-- ============================================================================
INSERT OR REPLACE INTO settings (key, value)
VALUES
    ('theme', '"dark"'),
    ('auto_track', 'true'),
    ('show_overlay', 'true'),
    ('log_path', '""');

-- ============================================================================
-- PLAYER STATS
-- ============================================================================
INSERT OR REPLACE INTO player_stats (date, format, matches_played, matches_won, games_played, games_won, account_id)
VALUES
    ('2024-10-20', 'Standard', 3, 2, 7, 4, 1),
    ('2024-10-20', 'QuickDraft_DSK', 3, 2, 7, 5, 1),
    ('2024-10-19', 'Standard', 2, 2, 4, 4, 1),
    ('2024-10-18', 'Standard', 1, 0, 3, 1, 1),
    ('2024-10-17', 'Historic', 1, 1, 2, 2, 1),
    ('2024-10-16', 'Historic', 1, 1, 3, 2, 1),
    ('2024-10-15', 'Standard', 1, 1, 2, 2, 1),
    ('2024-10-14', 'Standard', 1, 0, 3, 1, 1);

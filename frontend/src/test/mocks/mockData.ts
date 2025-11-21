export const mockDraftSession = {
  id: 'test-session-1',
  event_type: 'PremierDraft',
  set_code: 'BLB',
  start_time: '2025-11-20T10:00:00Z',
  status: 'active',
  picks_count: 0,
  wins: 0,
  losses: 0,
};

export const mockCompletedDraftSession = {
  ...mockDraftSession,
  id: 'test-session-2',
  status: 'completed',
  picks_count: 45,
  wins: 7,
  losses: 0,
};

export const mockSetCard = {
  arena_id: '12345',
  name: 'Test Card',
  set_code: 'BLB',
  rarity: 'rare',
  colors: ['W', 'U'],
  mana_cost: '{2}{W}{U}',
  cmc: 4,
  type_line: 'Creature - Human Wizard',
  card_type: 'Creature',
  image_uri: 'https://example.com/card.jpg',
};

export const mockCardRating = {
  arena_id: '12345',
  set_code: 'BLB',
  name: 'Test Card',
  gih_wr: 0.56,
  oh_wr: 0.54,
  gd_wr: 0.55,
  iwd: 0.02,
  tier: 'A',
};

export const mockDraftPick = {
  session_id: 'test-session-1',
  pack_number: 1,
  pick_number: 1,
  picked_card_id: '12345',
  available_cards: ['12345', '23456', '34567'],
  timestamp: '2025-11-20T10:05:00Z',
};

export const mockMatch = {
  id: 'match-1',
  event_id: 'event-1',
  opponent_name: 'TestOpponent',
  result: 'win',
  turns: 10,
  duration: 600,
  timestamp: '2025-11-20T11:00:00Z',
  player_deck_colors: ['W', 'U'],
  opponent_deck_colors: ['B', 'R'],
};

export const mockFormatStats = {
  set_code: 'BLB',
  format: 'PremierDraft',
  total_drafts: 100,
  total_matches: 350,
  win_rate: 0.55,
  avg_wins: 3.5,
  avg_losses: 2.1,
};

export const mockArchetype = {
  set_code: 'BLB',
  format: 'PremierDraft',
  color_identity: 'WU',
  archetype_name: 'Azorius Flyers',
  games_played: 5000,
  win_rate: 0.56,
  popularity: 0.12,
};

export const mockDraftGrade = {
  session_id: 'test-session-1',
  grade: 'B+',
  score: 75.5,
  gih_wr_score: 20,
  oh_wr_score: 18,
  gd_wr_score: 19,
  iwd_score: 18.5,
  total_score: 75.5,
  predicted_wins: 4.2,
};

export function createMockDraftSession(overrides = {}) {
  return { ...mockDraftSession, ...overrides };
}

export function createMockSetCard(overrides = {}) {
  return { ...mockSetCard, ...overrides };
}

export function createMockCardRating(overrides = {}) {
  return { ...mockCardRating, ...overrides };
}

export function createMockMatch(overrides = {}) {
  return { ...mockMatch, ...overrides };
}

export function createMockArchetype(overrides = {}) {
  return { ...mockArchetype, ...overrides };
}

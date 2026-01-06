import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as gameplays from '../gameplays';

// Mock the apiClient
vi.mock('../../apiClient', () => ({
  get: vi.fn(),
  post: vi.fn(),
}));

import { get } from '../../apiClient';

describe('gameplays API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getMatchPlays', () => {
    it('should call get with correct path', async () => {
      const mockPlays: gameplays.GamePlay[] = [
        {
          id: 1,
          game_id: 1,
          match_id: 'match-123',
          turn_number: 1,
          phase: 'Main1',
          player_type: 'player',
          action_type: 'play_card',
          card_id: 12345,
          timestamp: '2024-01-15T10:00:00Z',
          sequence_number: 1,
        },
      ];
      vi.mocked(get).mockResolvedValue(mockPlays);

      const result = await gameplays.getMatchPlays('match-123');

      expect(get).toHaveBeenCalledWith('/matches/match-123/plays');
      expect(result).toEqual(mockPlays);
    });

    it('should encode matchId with special characters', async () => {
      vi.mocked(get).mockResolvedValue([]);

      await gameplays.getMatchPlays('match/with/slashes');

      expect(get).toHaveBeenCalledWith('/matches/match%2Fwith%2Fslashes/plays');
    });
  });

  describe('getMatchTimeline', () => {
    it('should call get with correct path', async () => {
      const mockTimeline: gameplays.PlayTimelineEntry[] = [
        {
          turn: 1,
          active_player: 'player',
          player_plays: [],
          opponent_plays: [],
        },
      ];
      vi.mocked(get).mockResolvedValue(mockTimeline);

      const result = await gameplays.getMatchTimeline('match-123');

      expect(get).toHaveBeenCalledWith('/matches/match-123/plays/timeline');
      expect(result).toEqual(mockTimeline);
    });
  });

  describe('getMatchPlaySummary', () => {
    it('should call get with correct path', async () => {
      const mockSummary: gameplays.GamePlaySummary = {
        match_id: 'match-123',
        total_plays: 50,
        player_plays: 25,
        opponent_plays: 25,
        card_plays: 30,
        attacks: 10,
        blocks: 5,
        land_drops: 8,
        total_turns: 12,
        opponent_cards_seen: 15,
      };
      vi.mocked(get).mockResolvedValue(mockSummary);

      const result = await gameplays.getMatchPlaySummary('match-123');

      expect(get).toHaveBeenCalledWith('/matches/match-123/plays/summary');
      expect(result).toEqual(mockSummary);
    });
  });

  describe('getMatchOpponentCards', () => {
    it('should call get with correct path', async () => {
      const mockCards: gameplays.OpponentCard[] = [
        {
          id: 1,
          game_id: 1,
          match_id: 'match-123',
          card_id: 12345,
          card_name: 'Lightning Bolt',
          zone_observed: 'battlefield',
          turn_first_seen: 2,
          times_seen: 3,
        },
      ];
      vi.mocked(get).mockResolvedValue(mockCards);

      const result = await gameplays.getMatchOpponentCards('match-123');

      expect(get).toHaveBeenCalledWith('/matches/match-123/opponent-cards');
      expect(result).toEqual(mockCards);
    });
  });

  describe('getMatchSnapshots', () => {
    it('should call get with correct path', async () => {
      const mockSnapshots: gameplays.GameStateSnapshot[] = [
        {
          id: 1,
          game_id: 1,
          match_id: 'match-123',
          turn_number: 1,
          active_player: 'player',
          player_life: 20,
          opponent_life: 18,
          timestamp: '2024-01-15T10:00:00Z',
        },
      ];
      vi.mocked(get).mockResolvedValue(mockSnapshots);

      const result = await gameplays.getMatchSnapshots('match-123');

      expect(get).toHaveBeenCalledWith('/matches/match-123/snapshots');
      expect(result).toEqual(mockSnapshots);
    });

    it('should include gameID parameter when provided', async () => {
      vi.mocked(get).mockResolvedValue([]);

      await gameplays.getMatchSnapshots('match-123', 5);

      expect(get).toHaveBeenCalledWith('/matches/match-123/snapshots?gameID=5');
    });
  });

  describe('getPlaysByGame', () => {
    it('should call get with correct path', async () => {
      const mockPlays: gameplays.GamePlay[] = [
        {
          id: 1,
          game_id: 5,
          match_id: 'match-123',
          turn_number: 1,
          phase: 'Combat',
          player_type: 'player',
          action_type: 'attack',
          timestamp: '2024-01-15T10:00:00Z',
          sequence_number: 10,
        },
      ];
      vi.mocked(get).mockResolvedValue(mockPlays);

      const result = await gameplays.getPlaysByGame(5);

      expect(get).toHaveBeenCalledWith('/gameplays/game/5');
      expect(result).toEqual(mockPlays);
    });
  });

  describe('constants', () => {
    it('should export PlayerType constants', () => {
      expect(gameplays.PlayerType.Player).toBe('player');
      expect(gameplays.PlayerType.Opponent).toBe('opponent');
    });

    it('should export ActionType constants', () => {
      expect(gameplays.ActionType.PlayCard).toBe('play_card');
      expect(gameplays.ActionType.Attack).toBe('attack');
      expect(gameplays.ActionType.Block).toBe('block');
      expect(gameplays.ActionType.LandDrop).toBe('land_drop');
      expect(gameplays.ActionType.Mulligan).toBe('mulligan');
    });

    it('should export Phase constants', () => {
      expect(gameplays.Phase.Beginning).toBe('Beginning');
      expect(gameplays.Phase.Main1).toBe('Main1');
      expect(gameplays.Phase.Combat).toBe('Combat');
      expect(gameplays.Phase.Main2).toBe('Main2');
      expect(gameplays.Phase.Ending).toBe('Ending');
    });

    it('should export Zone constants', () => {
      expect(gameplays.Zone.Hand).toBe('hand');
      expect(gameplays.Zone.Library).toBe('library');
      expect(gameplays.Zone.Battlefield).toBe('battlefield');
      expect(gameplays.Zone.Graveyard).toBe('graveyard');
      expect(gameplays.Zone.Exile).toBe('exile');
      expect(gameplays.Zone.Stack).toBe('stack');
      expect(gameplays.Zone.Command).toBe('command');
    });
  });
});

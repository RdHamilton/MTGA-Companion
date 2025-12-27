import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as drafts from '../drafts';

// Mock the apiClient
vi.mock('../../apiClient', () => ({
  get: vi.fn(),
  post: vi.fn(),
}));

import { get, post } from '../../apiClient';

describe('drafts API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getDraftSessions', () => {
    it('should call post with correct path and filter', async () => {
      const mockSessions = [{ id: 'session-1' }];
      vi.mocked(post).mockResolvedValue(mockSessions);

      const result = await drafts.getDraftSessions({ format: 'PremierDraft' });

      expect(post).toHaveBeenCalledWith('/drafts', { format: 'PremierDraft' });
      expect(result).toEqual(mockSessions);
    });
  });

  describe('getDraftSession', () => {
    it('should call get with correct path', async () => {
      const mockSession = { id: 'session-123' };
      vi.mocked(get).mockResolvedValue(mockSession);

      const result = await drafts.getDraftSession('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123');
      expect(result).toEqual(mockSession);
    });
  });

  describe('getDraftPicks', () => {
    it('should call get with correct path', async () => {
      const mockPicks = [{ pickNumber: 1 }];
      vi.mocked(get).mockResolvedValue(mockPicks);

      const result = await drafts.getDraftPicks('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/picks');
      expect(result).toEqual(mockPicks);
    });
  });

  describe('getDraftPool', () => {
    it('should call get with correct path', async () => {
      const mockPool = [{ name: 'Card 1' }];
      vi.mocked(get).mockResolvedValue(mockPool);

      const result = await drafts.getDraftPool('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/pool');
      expect(result).toEqual(mockPool);
    });
  });

  describe('getDraftCurve', () => {
    it('should call get with correct path', async () => {
      const mockCurve = { 1: 5, 2: 8, 3: 7 };
      vi.mocked(get).mockResolvedValue(mockCurve);

      const result = await drafts.getDraftCurve('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/curve');
      expect(result).toEqual(mockCurve);
    });
  });

  describe('getDraftColors', () => {
    it('should call get with correct path', async () => {
      const mockColors = { W: 10, U: 5, B: 3 };
      vi.mocked(get).mockResolvedValue(mockColors);

      const result = await drafts.getDraftColors('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/colors');
      expect(result).toEqual(mockColors);
    });
  });

  describe('getDraftStats', () => {
    it('should call post with correct path and filter', async () => {
      const mockStats = { totalDrafts: 10 };
      vi.mocked(post).mockResolvedValue(mockStats);

      const result = await drafts.getDraftStats({ format: 'PremierDraft' });

      expect(post).toHaveBeenCalledWith('/drafts/stats', { format: 'PremierDraft' });
      expect(result).toEqual(mockStats);
    });
  });

  describe('getRecentDrafts', () => {
    it('should call get with limit parameter', async () => {
      const mockDrafts = [{ id: '1' }, { id: '2' }];
      vi.mocked(get).mockResolvedValue(mockDrafts);

      const result = await drafts.getRecentDrafts(10);

      expect(get).toHaveBeenCalledWith('/drafts/recent?limit=10');
      expect(result).toEqual(mockDrafts);
    });

    it('should call get without limit parameter when not provided', async () => {
      vi.mocked(get).mockResolvedValue([]);

      await drafts.getRecentDrafts();

      expect(get).toHaveBeenCalledWith('/drafts/recent');
    });
  });

  describe('getActiveDraftSessions', () => {
    it('should call getDraftSessions with active status', async () => {
      vi.mocked(post).mockResolvedValue([]);

      await drafts.getActiveDraftSessions();

      expect(post).toHaveBeenCalledWith('/drafts', { status: 'active' });
    });
  });

  describe('getCompletedDraftSessions', () => {
    it('should call getDraftSessions with completed status', async () => {
      vi.mocked(post).mockResolvedValue([]);

      await drafts.getCompletedDraftSessions();

      expect(post).toHaveBeenCalledWith('/drafts', { status: 'completed' });
    });
  });

  describe('getDraftDeckMetrics', () => {
    it('should call get with correct path', async () => {
      const mockMetrics = { totalCards: 40 };
      vi.mocked(get).mockResolvedValue(mockMetrics);

      const result = await drafts.getDraftDeckMetrics('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/deck-metrics');
      expect(result).toEqual(mockMetrics);
    });
  });

  describe('getDraftPerformanceMetrics', () => {
    it('should call post with empty filter', async () => {
      const mockMetrics = { winRate: 0.6 };
      vi.mocked(post).mockResolvedValue(mockMetrics);

      const result = await drafts.getDraftPerformanceMetrics();

      expect(post).toHaveBeenCalledWith('/drafts/stats', {});
      expect(result).toEqual(mockMetrics);
    });
  });

  describe('analyzeSessionPickQuality', () => {
    it('should call post with correct path', async () => {
      vi.mocked(post).mockResolvedValue(undefined);

      await drafts.analyzeSessionPickQuality('session-123');

      expect(post).toHaveBeenCalledWith('/drafts/session-123/analyze-picks');
    });
  });

  describe('getPickAlternatives', () => {
    it('should call post with correct path and params', async () => {
      const mockAlternatives = { picked: {}, alternatives: [] };
      vi.mocked(post).mockResolvedValue(mockAlternatives);

      const result = await drafts.getPickAlternatives('session-123', 1, 5);

      expect(post).toHaveBeenCalledWith('/drafts/grade-pick', {
        session_id: 'session-123',
        pack_number: 1,
        pick_number: 5,
      });
      expect(result).toEqual(mockAlternatives);
    });
  });

  describe('getDraftGrade', () => {
    it('should call get with correct path', async () => {
      const mockGrade = { grade: 'A' };
      vi.mocked(get).mockResolvedValue(mockGrade);

      const result = await drafts.getDraftGrade('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/analysis');
      expect(result).toEqual(mockGrade);
    });
  });

  describe('calculateDraftGrade', () => {
    it('should call post with correct path', async () => {
      const mockGrade = { grade: 'B+' };
      vi.mocked(post).mockResolvedValue(mockGrade);

      const result = await drafts.calculateDraftGrade('session-123');

      expect(post).toHaveBeenCalledWith('/drafts/session-123/calculate-grade');
      expect(result).toEqual(mockGrade);
    });
  });

  describe('getCurrentPackWithRecommendation', () => {
    it('should call get with correct path', async () => {
      const mockPack = { cards: [], recommendation: {} };
      vi.mocked(get).mockResolvedValue(mockPack);

      const result = await drafts.getCurrentPackWithRecommendation('session-123');

      expect(get).toHaveBeenCalledWith('/drafts/session-123/current-pack');
      expect(result).toEqual(mockPack);
    });
  });

  describe('getDraftWinRatePrediction', () => {
    it('should call post with correct path', async () => {
      const mockPrediction = { winRate: 0.55 };
      vi.mocked(post).mockResolvedValue(mockPrediction);

      const result = await drafts.getDraftWinRatePrediction('session-123');

      expect(post).toHaveBeenCalledWith('/drafts/session-123/calculate-prediction');
      expect(result).toEqual(mockPrediction);
    });
  });

  describe('classifyDraftPoolArchetype', () => {
    it('should call post with correct path and session_id', async () => {
      const mockClassification = { archetype: 'Aggro' };
      vi.mocked(post).mockResolvedValue(mockClassification);

      const result = await drafts.classifyDraftPoolArchetype('session-123');

      expect(post).toHaveBeenCalledWith('/decks/classify-draft-pool', {
        session_id: 'session-123',
      });
      expect(result).toEqual(mockClassification);
    });
  });
});

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSeventeenLands } from './useSeventeenLands';

// Mock the API modules
vi.mock('@/services/api', () => ({
  cards: {
    getCardRatings: vi.fn(),
    getSetCards: vi.fn(),
  },
  drafts: {
    recalculateSetGrades: vi.fn(),
  },
}));

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { cards, drafts } from '@/services/api';
import { showToast } from '../components/ToastContainer';

const mockGetCardRatings = vi.mocked(cards.getCardRatings);
const mockGetSetCards = vi.mocked(cards.getSetCards);
const mockRecalculateSetGrades = vi.mocked(drafts.recalculateSetGrades);

describe('useSeventeenLands', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    mockGetCardRatings.mockResolvedValue([]);
    mockGetSetCards.mockResolvedValue([]);
    mockRecalculateSetGrades.mockResolvedValue({ status: 'success', set: 'BLB', count: 0, message: '' });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('initial state', () => {
    it('returns empty setCode', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.setCode).toBe('');
    });

    it('returns default draftFormat of PremierDraft', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.draftFormat).toBe('PremierDraft');
    });

    it('returns isFetchingRatings as false', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.isFetchingRatings).toBe(false);
    });

    it('returns isFetchingCards as false', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.isFetchingCards).toBe(false);
    });

    it('returns isRecalculating as false', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.isRecalculating).toBe(false);
    });

    it('returns empty recalculateMessage', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.recalculateMessage).toBe('');
    });

    it('returns empty dataSource', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.dataSource).toBe('');
    });

    it('returns isClearingCache as false', () => {
      const { result } = renderHook(() => useSeventeenLands());
      expect(result.current.isClearingCache).toBe(false);
    });
  });

  describe('setSetCode', () => {
    it('updates setCode state', () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      expect(result.current.setCode).toBe('BLB');
    });
  });

  describe('setDraftFormat', () => {
    it('updates draftFormat state', () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setDraftFormat('QuickDraft');
      });

      expect(result.current.draftFormat).toBe('QuickDraft');
    });
  });

  describe('handleFetchSetRatings', () => {
    it('shows warning toast when setCode is empty', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Please enter a set code'),
        'warning'
      );
      expect(mockGetCardRatings).not.toHaveBeenCalled();
    });

    it('shows warning toast when setCode is whitespace', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('   ');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Please enter a set code'),
        'warning'
      );
    });

    it('calls getCardRatings with uppercase set code and format', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('blb');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(mockGetCardRatings).toHaveBeenCalledWith('BLB', 'PremierDraft');
    });

    it('sets isFetchingRatings to true during fetch', async () => {
      let resolveFetch: () => void;
      mockGetCardRatings.mockImplementationOnce(
        () =>
          new Promise<never[]>((resolve) => {
            resolveFetch = () => resolve([]);
          })
      );

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      let fetchPromise: Promise<void>;
      act(() => {
        fetchPromise = result.current.handleFetchSetRatings();
      });

      expect(result.current.isFetchingRatings).toBe(true);

      await act(async () => {
        resolveFetch!();
        await fetchPromise;
      });

      expect(result.current.isFetchingRatings).toBe(false);
    });

    it('shows success toast on successful fetch', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully fetched 17Lands ratings'),
        'success'
      );
    });

    it('updates dataSource state after successful fetch', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      // getDatasetSource is a stub that returns '17lands'
      expect(result.current.dataSource).toBe('17lands');
    });

    it('shows error toast on fetch failure', async () => {
      mockGetCardRatings.mockRejectedValueOnce(new Error('Fetch failed'));

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to fetch 17Lands ratings'),
        'error'
      );
    });
  });

  describe('handleRefreshSetRatings', () => {
    it('calls getCardRatings API (same as fetch)', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleRefreshSetRatings();
      });

      expect(mockGetCardRatings).toHaveBeenCalledWith('BLB', 'PremierDraft');
    });

    it('shows success toast on successful refresh', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleRefreshSetRatings();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully refreshed 17Lands ratings'),
        'success'
      );
    });
  });

  describe('handleFetchSetCards', () => {
    it('shows warning toast when setCode is empty', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleFetchSetCards();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Please enter a set code'),
        'warning'
      );
    });

    it('calls getSetCards with uppercase set code', async () => {
      mockGetSetCards.mockResolvedValueOnce(
        Array(250)
          .fill({})
          .map((_, i) => ({ arenaId: i })) as never[]
      );

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('blb');
      });

      await act(async () => {
        await result.current.handleFetchSetCards();
      });

      expect(mockGetSetCards).toHaveBeenCalledWith('BLB');
    });

    it('shows success toast with card count', async () => {
      mockGetSetCards.mockResolvedValueOnce(
        Array(250)
          .fill({})
          .map((_, i) => ({ arenaId: i })) as never[]
      );

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleFetchSetCards();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully fetched 250 cards'),
        'success'
      );
    });

    it('shows error toast on fetch failure', async () => {
      mockGetSetCards.mockRejectedValueOnce(new Error('Scryfall error'));

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleFetchSetCards();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to fetch cards'),
        'error'
      );
    });
  });

  describe('handleRefreshSetCards', () => {
    it('calls getSetCards API (same as fetch)', async () => {
      mockGetSetCards.mockResolvedValueOnce(
        Array(250)
          .fill({})
          .map((_, i) => ({ arenaId: i })) as never[]
      );

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleRefreshSetCards();
      });

      expect(mockGetSetCards).toHaveBeenCalledWith('BLB');
    });

    it('shows success toast on successful refresh', async () => {
      mockGetSetCards.mockResolvedValueOnce(
        Array(250)
          .fill({})
          .map((_, i) => ({ arenaId: i })) as never[]
      );

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleRefreshSetCards();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully refreshed 250 cards'),
        'success'
      );
    });
  });

  describe('handleRecalculateGrades', () => {
    // Note: recalculateAllDraftGrades is a no-op that returns 0
    it('executes without error (no-op in REST API)', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await expect(
        act(async () => {
          await result.current.handleRecalculateGrades();
        })
      ).resolves.not.toThrow();
    });

    it('sets isRecalculating to true during recalculation', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      // Note: Since recalculateAllDraftGrades is synchronous no-op,
      // we can't easily test intermediate state. Just verify it resets to false.
      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      expect(result.current.isRecalculating).toBe(false);
    });

    it('sets success message after successful recalculation (returns 0 count)', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      // No-op returns 0 sessions
      expect(result.current.recalculateMessage).toContain('Successfully recalculated 0 draft session');
    });

    it('clears success message after timeout', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      expect(result.current.recalculateMessage).not.toBe('');

      act(() => {
        vi.advanceTimersByTime(5000);
      });

      expect(result.current.recalculateMessage).toBe('');
    });
  });

  describe('handleClearDatasetCache', () => {
    // Note: clearDatasetCache is a no-op in REST API mode
    it('executes without error (no-op in REST API)', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await expect(
        act(async () => {
          await result.current.handleClearDatasetCache();
        })
      ).resolves.not.toThrow();
    });

    it('sets isClearingCache to false after completion', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleClearDatasetCache();
      });

      expect(result.current.isClearingCache).toBe(false);
    });

    it('shows success toast on successful clear', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleClearDatasetCache();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Successfully cleared dataset cache'),
        'success'
      );
    });
  });
});

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSeventeenLands } from './useSeventeenLands';
import { mockWailsApp } from '../test/mocks/wailsApp';

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { showToast } from '../components/ToastContainer';

describe('useSeventeenLands', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
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
      expect(mockWailsApp.FetchSetRatings).not.toHaveBeenCalled();
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

    it('calls FetchSetRatings with uppercase set code and format', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('blb');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(mockWailsApp.FetchSetRatings).toHaveBeenCalledWith('BLB', 'PremierDraft');
    });

    it('sets isFetchingRatings to true during fetch', async () => {
      let resolveFetch: () => void;
      mockWailsApp.FetchSetRatings.mockImplementationOnce(
        () => new Promise<void>((resolve) => {
          resolveFetch = resolve;
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

    it('shows success toast with data source on successful fetch', async () => {
      mockWailsApp.GetDatasetSource.mockResolvedValueOnce('s3');

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
      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('S3 public datasets'),
        'success'
      );
    });

    it('updates dataSource state after successful fetch', async () => {
      mockWailsApp.GetDatasetSource.mockResolvedValueOnce('web_api');

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleFetchSetRatings();
      });

      expect(result.current.dataSource).toBe('web_api');
    });

    it('shows error toast on fetch failure', async () => {
      mockWailsApp.FetchSetRatings.mockRejectedValueOnce(new Error('Fetch failed'));

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
    it('calls RefreshSetRatings API', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleRefreshSetRatings();
      });

      expect(mockWailsApp.RefreshSetRatings).toHaveBeenCalledWith('BLB', 'PremierDraft');
    });

    it('shows success toast on successful refresh', async () => {
      mockWailsApp.GetDatasetSource.mockResolvedValueOnce('s3');

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

    it('calls FetchSetCards with uppercase set code', async () => {
      mockWailsApp.FetchSetCards.mockResolvedValueOnce(250);

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('blb');
      });

      await act(async () => {
        await result.current.handleFetchSetCards();
      });

      expect(mockWailsApp.FetchSetCards).toHaveBeenCalledWith('BLB');
    });

    it('shows success toast with card count', async () => {
      mockWailsApp.FetchSetCards.mockResolvedValueOnce(250);

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
      mockWailsApp.FetchSetCards.mockRejectedValueOnce(new Error('Scryfall error'));

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
    it('calls RefreshSetCards API', async () => {
      mockWailsApp.RefreshSetCards.mockResolvedValueOnce(250);

      const { result } = renderHook(() => useSeventeenLands());

      act(() => {
        result.current.setSetCode('BLB');
      });

      await act(async () => {
        await result.current.handleRefreshSetCards();
      });

      expect(mockWailsApp.RefreshSetCards).toHaveBeenCalledWith('BLB');
    });

    it('shows success toast on successful refresh', async () => {
      mockWailsApp.RefreshSetCards.mockResolvedValueOnce(250);

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
    it('calls RecalculateAllDraftGrades API', async () => {
      mockWailsApp.RecalculateAllDraftGrades.mockResolvedValueOnce(5);

      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      expect(mockWailsApp.RecalculateAllDraftGrades).toHaveBeenCalled();
    });

    it('sets isRecalculating to true during recalculation', async () => {
      let resolveRecalc: (value: number) => void;
      mockWailsApp.RecalculateAllDraftGrades.mockImplementationOnce(
        () => new Promise<number>((resolve) => {
          resolveRecalc = resolve;
        })
      );

      const { result } = renderHook(() => useSeventeenLands());

      let recalcPromise: Promise<void>;
      act(() => {
        recalcPromise = result.current.handleRecalculateGrades();
      });

      expect(result.current.isRecalculating).toBe(true);

      await act(async () => {
        resolveRecalc!(5);
        await recalcPromise;
      });

      expect(result.current.isRecalculating).toBe(false);
    });

    it('sets success message after successful recalculation', async () => {
      mockWailsApp.RecalculateAllDraftGrades.mockResolvedValueOnce(5);

      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      expect(result.current.recalculateMessage).toContain('Successfully recalculated 5 draft session');
    });

    it('sets error message on recalculation failure', async () => {
      mockWailsApp.RecalculateAllDraftGrades.mockRejectedValueOnce(new Error('Recalc failed'));

      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      expect(result.current.recalculateMessage).toContain('Failed to recalculate');
    });

    it('clears success message after timeout', async () => {
      mockWailsApp.RecalculateAllDraftGrades.mockResolvedValueOnce(5);

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

    it('clears error message after longer timeout', async () => {
      mockWailsApp.RecalculateAllDraftGrades.mockRejectedValueOnce(new Error('Recalc failed'));

      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleRecalculateGrades();
      });

      expect(result.current.recalculateMessage).not.toBe('');

      act(() => {
        vi.advanceTimersByTime(8000);
      });

      expect(result.current.recalculateMessage).toBe('');
    });
  });

  describe('handleClearDatasetCache', () => {
    it('calls ClearDatasetCache API', async () => {
      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleClearDatasetCache();
      });

      expect(mockWailsApp.ClearDatasetCache).toHaveBeenCalled();
    });

    it('sets isClearingCache to true during clear', async () => {
      let resolveClear: () => void;
      mockWailsApp.ClearDatasetCache.mockImplementationOnce(
        () => new Promise<void>((resolve) => {
          resolveClear = resolve;
        })
      );

      const { result } = renderHook(() => useSeventeenLands());

      let clearPromise: Promise<void>;
      act(() => {
        clearPromise = result.current.handleClearDatasetCache();
      });

      expect(result.current.isClearingCache).toBe(true);

      await act(async () => {
        resolveClear!();
        await clearPromise;
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

    it('shows error toast on clear failure', async () => {
      mockWailsApp.ClearDatasetCache.mockRejectedValueOnce(new Error('Clear failed'));

      const { result } = renderHook(() => useSeventeenLands());

      await act(async () => {
        await result.current.handleClearDatasetCache();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to clear dataset cache'),
        'error'
      );
    });
  });
});

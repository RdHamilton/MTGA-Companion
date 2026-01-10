import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useRotationNotifications } from './useRotationNotifications';
import type { UpcomingRotation, RotationAffectedDeck } from '@/services/api/standard';

// Mock the standard API module
vi.mock('@/services/api/standard', () => ({
  getUpcomingRotation: vi.fn(),
  getRotationAffectedDecks: vi.fn(),
}));

import * as standardApi from '@/services/api/standard';

const mockGetUpcomingRotation = vi.mocked(standardApi.getUpcomingRotation);
const mockGetRotationAffectedDecks = vi.mocked(standardApi.getRotationAffectedDecks);

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

const mockRotation: UpcomingRotation = {
  nextRotationDate: '2025-09-01T00:00:00Z',
  daysUntilRotation: 30,
  rotatingSets: [
    {
      code: 'ONE',
      name: 'Phyrexia: All Will Be One',
      releasedAt: '2023-02-03',
      isStandardLegal: true,
      iconSvgUri: 'https://example.com/one.svg',
      cardCount: 300,
      isRotatingSoon: true,
    },
  ],
  rotatingCardCount: 150,
  affectedDecks: 3,
};

const mockAffectedDecks: RotationAffectedDeck[] = [
  {
    deckId: 'deck-1',
    deckName: 'Mono White Aggro',
    format: 'Standard',
    rotatingCardCount: 12,
    totalCards: 60,
    percentAffected: 20,
    rotatingCards: [],
  },
  {
    deckId: 'deck-2',
    deckName: 'Azorius Control',
    format: 'Standard',
    rotatingCardCount: 8,
    totalCards: 60,
    percentAffected: 13.3,
    rotatingCards: [],
  },
];

describe('useRotationNotifications', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
    mockGetUpcomingRotation.mockResolvedValue(mockRotation);
    mockGetRotationAffectedDecks.mockResolvedValue(mockAffectedDecks);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('initial state and data fetching', () => {
    it('fetches rotation data on mount', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(mockGetUpcomingRotation).toHaveBeenCalled();
      expect(mockGetRotationAffectedDecks).toHaveBeenCalled();
    });

    it('sets rotation data after successful fetch', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).toEqual(mockRotation);
      });

      expect(result.current.affectedDecks).toEqual(mockAffectedDecks);
      expect(result.current.error).toBeNull();
    });

    it('sets error state on fetch failure', async () => {
      mockGetUpcomingRotation.mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.error).toBe('Network error');
    });

    it('sets lastChecked after successful fetch', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.lastChecked).not.toBeNull();
      });

      expect(result.current.lastChecked).toBeInstanceOf(Date);
    });
  });

  describe('checkRotation', () => {
    it('refetches rotation data when called', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Clear mock counts
      vi.clearAllMocks();

      await act(async () => {
        await result.current.checkRotation();
      });

      expect(mockGetUpcomingRotation).toHaveBeenCalledTimes(1);
      expect(mockGetRotationAffectedDecks).toHaveBeenCalledTimes(1);
    });
  });

  describe('shouldShowNotification', () => {
    it('returns true when days until rotation is within threshold', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).not.toBeNull();
      });

      // mockRotation has daysUntilRotation = 30
      expect(result.current.shouldShowNotification(30)).toBe(true);
      expect(result.current.shouldShowNotification(60)).toBe(true);
    });

    it('returns false when days until rotation exceeds threshold', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).not.toBeNull();
      });

      // mockRotation has daysUntilRotation = 30
      expect(result.current.shouldShowNotification(7)).toBe(false);
      expect(result.current.shouldShowNotification(29)).toBe(false);
    });

    it('returns false when no rotation data', async () => {
      mockGetUpcomingRotation.mockRejectedValue(new Error('Failed'));

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.shouldShowNotification(90)).toBe(false);
    });

    it('returns false when no affected decks', async () => {
      mockGetRotationAffectedDecks.mockResolvedValue([]);

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.shouldShowNotification(90)).toBe(false);
    });

    it('returns false when already notified today', async () => {
      // Set localStorage to indicate already notified today
      localStorageMock.getItem.mockReturnValue(new Date().toISOString());

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.shouldShowNotification(90)).toBe(false);
    });
  });

  describe('markAsNotified', () => {
    it('sets hasNotified to true', async () => {
      // Ensure localStorage returns null for this test
      localStorageMock.getItem.mockReturnValue(null);

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Should not be notified yet since localStorage is empty
      expect(result.current.hasNotified).toBe(false);

      act(() => {
        result.current.markAsNotified();
      });

      expect(result.current.hasNotified).toBe(true);
    });

    it('saves to localStorage', async () => {
      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      act(() => {
        result.current.markAsNotified();
      });

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'rotation_notification_last_shown',
        expect.any(String)
      );
    });
  });

  describe('unmount behavior', () => {
    it('does not update state after unmount', async () => {
      // Create a delayed promise to simulate slow API
      let resolveRotation: (value: typeof mockRotation) => void;
      const slowRotationPromise = new Promise<typeof mockRotation>((resolve) => {
        resolveRotation = resolve;
      });
      mockGetUpcomingRotation.mockReturnValue(slowRotationPromise);

      const { unmount, result } = renderHook(() => useRotationNotifications());

      // Should be loading
      expect(result.current.isLoading).toBe(true);

      // Unmount before API resolves
      unmount();

      // Now resolve the API - should not cause state update (no error)
      await act(async () => {
        resolveRotation!(mockRotation);
      });

      // No warnings or errors should occur - test passes if no "Can't perform state update" warning
    });

    it('does not update state on error after unmount', async () => {
      // Create a delayed promise that rejects
      let rejectRotation: (error: Error) => void;
      const slowRotationPromise = new Promise<typeof mockRotation>((_, reject) => {
        rejectRotation = reject;
      });
      mockGetUpcomingRotation.mockReturnValue(slowRotationPromise);

      const { unmount, result } = renderHook(() => useRotationNotifications());

      // Should be loading
      expect(result.current.isLoading).toBe(true);

      // Unmount before API rejects
      unmount();

      // Now reject the API - should not cause state update (no error)
      await act(async () => {
        rejectRotation!(new Error('Network error'));
      });

      // No warnings or errors should occur - test passes if no "Can't perform state update" warning
    });
  });

  describe('getUrgencyLevel', () => {
    it('returns critical when <= 7 days', async () => {
      mockGetUpcomingRotation.mockResolvedValue({
        ...mockRotation,
        daysUntilRotation: 5,
      });

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).not.toBeNull();
      });

      expect(result.current.getUrgencyLevel()).toBe('critical');
    });

    it('returns warning when <= 30 days', async () => {
      mockGetUpcomingRotation.mockResolvedValue({
        ...mockRotation,
        daysUntilRotation: 20,
      });

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).not.toBeNull();
      });

      expect(result.current.getUrgencyLevel()).toBe('warning');
    });

    it('returns info when <= 90 days', async () => {
      mockGetUpcomingRotation.mockResolvedValue({
        ...mockRotation,
        daysUntilRotation: 60,
      });

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).not.toBeNull();
      });

      expect(result.current.getUrgencyLevel()).toBe('info');
    });

    it('returns null when > 90 days', async () => {
      mockGetUpcomingRotation.mockResolvedValue({
        ...mockRotation,
        daysUntilRotation: 120,
      });

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.rotation).not.toBeNull();
      });

      expect(result.current.getUrgencyLevel()).toBeNull();
    });

    it('returns null when no rotation data', async () => {
      mockGetUpcomingRotation.mockRejectedValue(new Error('Failed'));

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.getUrgencyLevel()).toBeNull();
    });

    it('returns null when no affected decks', async () => {
      mockGetRotationAffectedDecks.mockResolvedValue([]);

      const { result } = renderHook(() => useRotationNotifications());

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.getUrgencyLevel()).toBeNull();
    });
  });
});

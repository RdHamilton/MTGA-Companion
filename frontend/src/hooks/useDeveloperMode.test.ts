import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDeveloperMode } from './useDeveloperMode';

describe('useDeveloperMode', () => {
  const STORAGE_KEY = 'mtga-companion-developer-mode';

  beforeEach(() => {
    localStorage.clear();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('initial state', () => {
    it('returns false when localStorage is empty', () => {
      const { result } = renderHook(() => useDeveloperMode());
      expect(result.current.isDeveloperMode).toBe(false);
    });

    it('returns true when localStorage has developer mode enabled', () => {
      localStorage.setItem(STORAGE_KEY, 'true');
      const { result } = renderHook(() => useDeveloperMode());
      expect(result.current.isDeveloperMode).toBe(true);
    });

    it('returns false when localStorage has developer mode disabled', () => {
      localStorage.setItem(STORAGE_KEY, 'false');
      const { result } = renderHook(() => useDeveloperMode());
      expect(result.current.isDeveloperMode).toBe(false);
    });
  });

  describe('toggleDeveloperMode', () => {
    it('toggles from false to true', () => {
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.toggleDeveloperMode();
      });

      expect(result.current.isDeveloperMode).toBe(true);
    });

    it('toggles from true to false', () => {
      localStorage.setItem(STORAGE_KEY, 'true');
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.toggleDeveloperMode();
      });

      expect(result.current.isDeveloperMode).toBe(false);
    });

    it('persists changes to localStorage', () => {
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.toggleDeveloperMode();
      });

      expect(localStorage.getItem(STORAGE_KEY)).toBe('true');
    });
  });

  describe('handleVersionClick', () => {
    it('increments click count on first click', () => {
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.handleVersionClick();
      });

      expect(result.current.clickCount).toBe(1);
    });

    it('increments click count on subsequent clicks within timeout', () => {
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.handleVersionClick();
      });
      act(() => {
        vi.advanceTimersByTime(1000);
        result.current.handleVersionClick();
      });
      act(() => {
        vi.advanceTimersByTime(1000);
        result.current.handleVersionClick();
      });

      expect(result.current.clickCount).toBe(3);
    });

    it('resets click count when timeout exceeded', () => {
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.handleVersionClick();
      });
      act(() => {
        result.current.handleVersionClick();
      });

      // Advance past timeout (3000ms)
      act(() => {
        vi.advanceTimersByTime(3500);
        result.current.handleVersionClick();
      });

      expect(result.current.clickCount).toBe(1);
    });

    it('toggles developer mode after 5 clicks', () => {
      const { result } = renderHook(() => useDeveloperMode());

      expect(result.current.isDeveloperMode).toBe(false);

      // Click 5 times within timeout
      for (let i = 0; i < 5; i++) {
        act(() => {
          vi.advanceTimersByTime(500);
          result.current.handleVersionClick();
        });
      }

      expect(result.current.isDeveloperMode).toBe(true);
      expect(result.current.clickCount).toBe(0); // Reset after toggle
    });

    it('toggles off developer mode after another 5 clicks', () => {
      localStorage.setItem(STORAGE_KEY, 'true');
      const { result } = renderHook(() => useDeveloperMode());

      expect(result.current.isDeveloperMode).toBe(true);

      // Click 5 times within timeout
      for (let i = 0; i < 5; i++) {
        act(() => {
          vi.advanceTimersByTime(500);
          result.current.handleVersionClick();
        });
      }

      expect(result.current.isDeveloperMode).toBe(false);
    });
  });

  describe('persistence', () => {
    it('persists enabled state to localStorage', () => {
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.toggleDeveloperMode();
      });

      expect(localStorage.getItem(STORAGE_KEY)).toBe('true');
    });

    it('persists disabled state to localStorage', () => {
      localStorage.setItem(STORAGE_KEY, 'true');
      const { result } = renderHook(() => useDeveloperMode());

      act(() => {
        result.current.toggleDeveloperMode();
      });

      expect(localStorage.getItem(STORAGE_KEY)).toBe('false');
    });
  });
});

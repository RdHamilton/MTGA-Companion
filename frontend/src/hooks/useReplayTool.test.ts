import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useReplayTool } from './useReplayTool';

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

// Mock App module for getReplayState and subscribeToReplayState
const mockReplayState = {
  isActive: false,
  isPaused: false,
  progress: null as ReturnType<typeof import('@/types/models').gui.ReplayStatus.createFrom> | null,
};

const mockSubscribers: ((state: typeof mockReplayState) => void)[] = [];

vi.mock('../App', () => ({
  getReplayState: vi.fn(() => mockReplayState),
  subscribeToReplayState: vi.fn((callback: (state: typeof mockReplayState) => void) => {
    mockSubscribers.push(callback);
    return () => {
      const index = mockSubscribers.indexOf(callback);
      if (index > -1) {
        mockSubscribers.splice(index, 1);
      }
    };
  }),
}));

import { showToast } from '../components/ToastContainer';

// Helper to emit state changes
function emitReplayStateChange(newState: Partial<typeof mockReplayState>) {
  Object.assign(mockReplayState, newState);
  mockSubscribers.forEach((cb) => cb(mockReplayState));
}

describe('useReplayTool', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockReplayState.isActive = false;
    mockReplayState.isPaused = false;
    mockReplayState.progress = null;
    mockSubscribers.length = 0;
  });

  describe('initial state', () => {
    it('returns replayToolActive from global state initializer', () => {
      // Set state BEFORE rendering so useState picks it up
      mockReplayState.isActive = true;
      const { result } = renderHook(() => useReplayTool());
      expect(result.current.replayToolActive).toBe(true);
    });

    it('returns replayToolPaused from global state initializer', () => {
      mockReplayState.isPaused = true;
      const { result } = renderHook(() => useReplayTool());
      expect(result.current.replayToolPaused).toBe(true);
    });

    it('returns replayToolProgress from global state initializer', () => {
      mockReplayState.progress = { currentEntry: 5, totalEntries: 10 } as never;
      const { result } = renderHook(() => useReplayTool());
      expect(result.current.replayToolProgress).toEqual({ currentEntry: 5, totalEntries: 10 });
    });

    it('returns default replaySpeed of 1.0', () => {
      const { result } = renderHook(() => useReplayTool());
      expect(result.current.replaySpeed).toBe(1.0);
    });

    it('returns default replayFilter of all', () => {
      const { result } = renderHook(() => useReplayTool());
      expect(result.current.replayFilter).toBe('all');
    });

    it('returns default pauseOnDraft of false', () => {
      const { result } = renderHook(() => useReplayTool());
      expect(result.current.pauseOnDraft).toBe(false);
    });
  });

  describe('setReplaySpeed', () => {
    it('updates replaySpeed state', () => {
      const { result } = renderHook(() => useReplayTool());

      act(() => {
        result.current.setReplaySpeed(5.0);
      });

      expect(result.current.replaySpeed).toBe(5.0);
    });
  });

  describe('setReplayFilter', () => {
    it('updates replayFilter state', () => {
      const { result } = renderHook(() => useReplayTool());

      act(() => {
        result.current.setReplayFilter('draft');
      });

      expect(result.current.replayFilter).toBe('draft');
    });
  });

  describe('setPauseOnDraft', () => {
    it('updates pauseOnDraft state', () => {
      const { result } = renderHook(() => useReplayTool());

      act(() => {
        result.current.setPauseOnDraft(true);
      });

      expect(result.current.pauseOnDraft).toBe(true);
    });
  });

  describe('handleStartReplayTool', () => {
    it('shows warning toast when not connected', async () => {
      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleStartReplayTool(false);
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Replay tool requires daemon mode'),
        'warning'
      );
    });

    // Note: StartReplayWithFileDialog is a no-op in REST API mode
    it('does not throw when connected (no-op in REST API)', async () => {
      const { result } = renderHook(() => useReplayTool());

      await expect(
        act(async () => {
          await result.current.handleStartReplayTool(true);
        })
      ).resolves.not.toThrow();
    });
  });

  describe('handlePauseReplayTool', () => {
    // Note: PauseReplay is a no-op in REST API mode
    it('does not throw (no-op in REST API)', async () => {
      const { result } = renderHook(() => useReplayTool());

      await expect(
        act(async () => {
          await result.current.handlePauseReplayTool();
        })
      ).resolves.not.toThrow();
    });
  });

  describe('handleResumeReplayTool', () => {
    // Note: ResumeReplay is a no-op in REST API mode
    it('does not throw (no-op in REST API)', async () => {
      const { result } = renderHook(() => useReplayTool());

      await expect(
        act(async () => {
          await result.current.handleResumeReplayTool();
        })
      ).resolves.not.toThrow();
    });
  });

  describe('handleStopReplayTool', () => {
    // Note: StopReplay is a no-op in REST API mode
    it('does not throw (no-op in REST API)', async () => {
      const { result } = renderHook(() => useReplayTool());

      await expect(
        act(async () => {
          await result.current.handleStopReplayTool();
        })
      ).resolves.not.toThrow();
    });
  });

  describe('state subscription', () => {
    it('updates state when global replay state changes', async () => {
      const { result } = renderHook(() => useReplayTool());

      expect(result.current.replayToolActive).toBe(false);

      act(() => {
        emitReplayStateChange({
          isActive: true,
          isPaused: false,
          progress: { currentEntry: 10, totalEntries: 100 } as never,
        });
      });

      expect(result.current.replayToolActive).toBe(true);
      expect(result.current.replayToolProgress).toEqual({ currentEntry: 10, totalEntries: 100 });
    });

    it('updates paused state from subscription', async () => {
      const { result } = renderHook(() => useReplayTool());

      act(() => {
        emitReplayStateChange({
          isActive: true,
          isPaused: true,
        });
      });

      expect(result.current.replayToolPaused).toBe(true);
    });
  });

  describe('cleanup', () => {
    it('unsubscribes from global state on unmount', () => {
      const { unmount } = renderHook(() => useReplayTool());

      expect(mockSubscribers.length).toBe(1);

      unmount();

      expect(mockSubscribers.length).toBe(0);
    });
  });
});

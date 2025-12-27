import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useReplayTool } from './useReplayTool';
import { mockWailsApp } from '@/test/mocks/apiMock';

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
  progress: null as any,
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
      mockReplayState.progress = { currentEntry: 5, totalEntries: 10 };
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
      expect(mockWailsApp.StartReplayWithFileDialog).not.toHaveBeenCalled();
    });

    it('calls StartReplayWithFileDialog when connected', async () => {
      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleStartReplayTool(true);
      });

      expect(mockWailsApp.StartReplayWithFileDialog).toHaveBeenCalledWith(1.0, 'all', false);
    });

    it('passes current settings to StartReplayWithFileDialog', async () => {
      const { result } = renderHook(() => useReplayTool());

      act(() => {
        result.current.setReplaySpeed(10.0);
        result.current.setReplayFilter('draft');
        result.current.setPauseOnDraft(true);
      });

      await act(async () => {
        await result.current.handleStartReplayTool(true);
      });

      expect(mockWailsApp.StartReplayWithFileDialog).toHaveBeenCalledWith(10.0, 'draft', true);
    });

    it('shows error toast on API failure', async () => {
      mockWailsApp.StartReplayWithFileDialog.mockRejectedValueOnce(new Error('Start failed'));

      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleStartReplayTool(true);
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to start replay'),
        'error'
      );
    });
  });

  describe('handlePauseReplayTool', () => {
    it('calls PauseReplay API', async () => {
      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handlePauseReplayTool();
      });

      expect(mockWailsApp.PauseReplay).toHaveBeenCalled();
    });

    it('shows error toast on API failure', async () => {
      mockWailsApp.PauseReplay.mockRejectedValueOnce(new Error('Pause failed'));

      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handlePauseReplayTool();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to pause replay'),
        'error'
      );
    });
  });

  describe('handleResumeReplayTool', () => {
    it('calls ResumeReplay API', async () => {
      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleResumeReplayTool();
      });

      expect(mockWailsApp.ResumeReplay).toHaveBeenCalled();
    });

    it('shows error toast on API failure', async () => {
      mockWailsApp.ResumeReplay.mockRejectedValueOnce(new Error('Resume failed'));

      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleResumeReplayTool();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to resume replay'),
        'error'
      );
    });
  });

  describe('handleStopReplayTool', () => {
    it('calls StopReplay API', async () => {
      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleStopReplayTool();
      });

      expect(mockWailsApp.StopReplay).toHaveBeenCalled();
    });

    it('shows error toast on API failure', async () => {
      mockWailsApp.StopReplay.mockRejectedValueOnce(new Error('Stop failed'));

      const { result } = renderHook(() => useReplayTool());

      await act(async () => {
        await result.current.handleStopReplayTool();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to stop replay'),
        'error'
      );
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
          progress: { currentEntry: 10, totalEntries: 100 },
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

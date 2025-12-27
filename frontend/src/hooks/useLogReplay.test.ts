import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useLogReplay } from './useLogReplay';
import { mockEventEmitter } from '@/test/mocks/websocketMock';

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { showToast } from '../components/ToastContainer';

describe('useLogReplay', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
  });

  describe('initial state', () => {
    it('returns clearDataBeforeReplay as false', () => {
      const { result } = renderHook(() => useLogReplay());

      expect(result.current.clearDataBeforeReplay).toBe(false);
    });

    it('returns isReplaying as false', () => {
      const { result } = renderHook(() => useLogReplay());

      expect(result.current.isReplaying).toBe(false);
    });

    it('returns replayProgress as null', () => {
      const { result } = renderHook(() => useLogReplay());

      expect(result.current.replayProgress).toBe(null);
    });
  });

  describe('setClearDataBeforeReplay', () => {
    it('updates clearDataBeforeReplay state', () => {
      const { result } = renderHook(() => useLogReplay());

      act(() => {
        result.current.setClearDataBeforeReplay(true);
      });

      expect(result.current.clearDataBeforeReplay).toBe(true);
    });
  });

  describe('handleReplayLogs', () => {
    it('does not trigger replay when not connected', async () => {
      const { result } = renderHook(() => useLogReplay());

      await act(async () => {
        await result.current.handleReplayLogs(false);
      });

      // No error toast should be shown since we just return early
      expect(showToast.show).not.toHaveBeenCalled();
    });

    // Note: TriggerReplayLogs is a no-op in REST API mode
    // So this test just verifies no error is thrown
    it('does not throw when connected (no-op in REST API)', async () => {
      const { result } = renderHook(() => useLogReplay());

      await expect(
        act(async () => {
          await result.current.handleReplayLogs(true);
        })
      ).resolves.not.toThrow();
    });
  });

  describe('event subscriptions', () => {
    it('sets isReplaying to true on replay:started event', async () => {
      const { result } = renderHook(() => useLogReplay());

      act(() => {
        mockEventEmitter.emit('replay:started');
      });

      await waitFor(() => {
        expect(result.current.isReplaying).toBe(true);
      });
    });

    it('clears replayProgress on replay:started event', async () => {
      const { result } = renderHook(() => useLogReplay());

      // First set some progress
      act(() => {
        mockEventEmitter.emit('replay:progress', {
          processedFiles: 5,
          totalFiles: 10,
        });
      });

      await waitFor(() => {
        expect(result.current.replayProgress).not.toBe(null);
      });

      // Then start a new replay
      act(() => {
        mockEventEmitter.emit('replay:started');
      });

      await waitFor(() => {
        expect(result.current.replayProgress).toBe(null);
      });
    });

    it('updates replayProgress on replay:progress event', async () => {
      const { result } = renderHook(() => useLogReplay());

      const progressData = {
        processedFiles: 5,
        totalFiles: 10,
        totalEntries: 1000,
        matchesImported: 10,
        decksImported: 5,
        questsImported: 2,
        duration: 5.5,
        currentFile: 'Player-prev.log',
      };

      act(() => {
        mockEventEmitter.emit('replay:progress', progressData);
      });

      await waitFor(() => {
        expect(result.current.replayProgress).not.toBe(null);
        expect(result.current.replayProgress?.processedFiles).toBe(5);
        expect(result.current.replayProgress?.totalFiles).toBe(10);
      });
    });

    it('sets isReplaying to false on replay:completed event', async () => {
      const { result } = renderHook(() => useLogReplay());

      // First set isReplaying to true
      act(() => {
        mockEventEmitter.emit('replay:started');
      });

      await waitFor(() => {
        expect(result.current.isReplaying).toBe(true);
      });

      // Then complete the replay
      act(() => {
        mockEventEmitter.emit('replay:completed', {
          processedFiles: 10,
          totalFiles: 10,
        });
      });

      await waitFor(() => {
        expect(result.current.isReplaying).toBe(false);
      });
    });

    it('updates replayProgress on replay:completed event', async () => {
      const { result } = renderHook(() => useLogReplay());

      const completedData = {
        processedFiles: 10,
        totalFiles: 10,
        totalEntries: 5000,
        matchesImported: 50,
      };

      act(() => {
        mockEventEmitter.emit('replay:completed', completedData);
      });

      await waitFor(() => {
        expect(result.current.replayProgress?.processedFiles).toBe(10);
        expect(result.current.replayProgress?.totalEntries).toBe(5000);
      });
    });

    it('sets isReplaying to false on replay:error event', async () => {
      const { result } = renderHook(() => useLogReplay());

      // First set isReplaying to true
      act(() => {
        mockEventEmitter.emit('replay:started');
      });

      await waitFor(() => {
        expect(result.current.isReplaying).toBe(true);
      });

      // Then emit error
      act(() => {
        mockEventEmitter.emit('replay:error', { error: 'Something went wrong' });
      });

      await waitFor(() => {
        expect(result.current.isReplaying).toBe(false);
      });
    });
  });

  describe('cleanup', () => {
    it('unsubscribes from events on unmount', () => {
      const { unmount } = renderHook(() => useLogReplay());

      unmount();

      // After unmount, emitting events should not cause errors
      // (callback references should be cleaned up)
      expect(() => {
        mockEventEmitter.emit('replay:started');
        mockEventEmitter.emit('replay:progress', {});
        mockEventEmitter.emit('replay:completed', {});
        mockEventEmitter.emit('replay:error', {});
      }).not.toThrow();
    });
  });
});

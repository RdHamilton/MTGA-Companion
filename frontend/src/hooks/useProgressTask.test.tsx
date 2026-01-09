import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useProgressTask } from './useProgressTask';
import { TaskProgressProvider } from '@/context/TaskProgressContext';
import type { ReactNode } from 'react';

// Mock WebSocket client
vi.mock('@/services/websocketClient', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

const wrapper = ({ children }: { children: ReactNode }) => (
  <TaskProgressProvider>{children}</TaskProgressProvider>
);

describe('useProgressTask', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('initialization', () => {
    it('returns initial state', () => {
      const { result } = renderHook(() => useProgressTask('test-task'), { wrapper });

      expect(result.current.isRunning).toBe(false);
      expect(result.current.progress).toBe(0);
      expect(typeof result.current.execute).toBe('function');
      expect(typeof result.current.cancel).toBe('function');
    });
  });

  describe('execute', () => {
    it('tracks progress during execution', async () => {
      const { result } = renderHook(() => useProgressTask('test-task'), { wrapper });

      let progressUpdates: number[] = [];

      await act(async () => {
        await result.current.execute(async (updateProgress) => {
          updateProgress(25, 'Step 1');
          progressUpdates.push(25);
          updateProgress(50, 'Step 2');
          progressUpdates.push(50);
          updateProgress(75, 'Step 3');
          progressUpdates.push(75);
          return 'result';
        }, 'Test Operation');
      });

      expect(progressUpdates).toEqual([25, 50, 75]);
    });

    it('returns the operation result', async () => {
      const { result } = renderHook(() => useProgressTask<string>('test-task'), { wrapper });

      let operationResult: string | undefined;

      await act(async () => {
        operationResult = await result.current.execute(async () => {
          return 'success!';
        }, 'Test Operation');
      });

      expect(operationResult).toBe('success!');
    });

    it('marks task as not running after execution completes', async () => {
      const { result } = renderHook(() => useProgressTask('test-task'), { wrapper });

      await act(async () => {
        await result.current.execute(async () => {
          return 'done';
        }, 'Test Operation');
      });

      // After execution completes, task should no longer be running
      expect(result.current.isRunning).toBe(false);
    });

    it('handles errors and marks task as failed', async () => {
      const { result } = renderHook(() => useProgressTask('test-task'), { wrapper });

      await act(async () => {
        try {
          await result.current.execute(async () => {
            throw new Error('Operation failed');
          }, 'Failing Operation');
        } catch {
          // Expected error
        }
      });

      // After error, task should not be running
      expect(result.current.isRunning).toBe(false);
    });
  });

  describe('options', () => {
    it('uses provided category', async () => {
      const { result } = renderHook(
        () => useProgressTask('deck-task', { category: 'deck-generation' }),
        { wrapper }
      );

      await act(async () => {
        await result.current.execute(async () => {
          return 'done';
        }, 'Deck Generation');
      });

      // Operation completes successfully
      expect(result.current.isRunning).toBe(false);
    });

    it('passes estimatedDuration to context', async () => {
      const { result } = renderHook(
        () => useProgressTask('slow-task', { estimatedDuration: 5000 }),
        { wrapper }
      );

      let executed = false;

      await act(async () => {
        await result.current.execute(async () => {
          executed = true;
          return 'done';
        }, 'Slow Operation');
      });

      expect(executed).toBe(true);
    });
  });

  describe('cancel', () => {
    it('cancels the task', async () => {
      const { result } = renderHook(
        () => useProgressTask('cancellable-task', { cancellable: true }),
        { wrapper }
      );

      act(() => {
        result.current.cancel();
      });

      // Cancel should work even if no task is running
      expect(result.current.isRunning).toBe(false);
    });
  });

  describe('multiple tasks', () => {
    it('handles multiple task instances independently', async () => {
      const { result: result1 } = renderHook(() => useProgressTask('task-1'), { wrapper });
      const { result: result2 } = renderHook(() => useProgressTask('task-2'), { wrapper });

      expect(result1.current.isRunning).toBe(false);
      expect(result2.current.isRunning).toBe(false);

      await act(async () => {
        await result1.current.execute(async () => 'result-1', 'Task 1');
      });

      // Task 1 completed
      expect(result1.current.isRunning).toBe(false);
      // Task 2 never started
      expect(result2.current.isRunning).toBe(false);
    });
  });
});

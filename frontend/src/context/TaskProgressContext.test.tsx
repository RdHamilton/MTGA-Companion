import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act, renderHook } from '@testing-library/react';
import { TaskProgressProvider, useTaskProgress } from './TaskProgressContext';

// Mock WebSocket client
vi.mock('@/services/websocketClient', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

function TestConsumer() {
  const { state, startTask, updateTask, completeTask, failTask, cancelTask, isRunning, activeTask } =
    useTaskProgress();

  return (
    <div>
      <div data-testid="task-count">{state.tasks.size}</div>
      <div data-testid="is-running">{isRunning() ? 'running' : 'idle'}</div>
      <div data-testid="active-task">{activeTask?.title || 'none'}</div>
      <button onClick={() => startTask('test-1', 'Test Task', 'general')}>Start Task</button>
      <button onClick={() => updateTask('test-1', 50, 'Half done')}>Update Task</button>
      <button onClick={() => completeTask('test-1')}>Complete Task</button>
      <button onClick={() => failTask('test-1', 'Error occurred')}>Fail Task</button>
      <button onClick={() => cancelTask('test-1')}>Cancel Task</button>
    </div>
  );
}

describe('TaskProgressContext', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('provider', () => {
    it('provides default state', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      expect(result.current.state.tasks.size).toBe(0);
      expect(result.current.activeTask).toBeNull();
      expect(result.current.isRunning()).toBe(false);
    });

    it('throws error when used outside provider', () => {
      expect(() => renderHook(() => useTaskProgress())).toThrow(
        'useTaskProgress must be used within a TaskProgressProvider'
      );
    });
  });

  describe('startTask', () => {
    it('adds a new task to state', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      expect(result.current.state.tasks.size).toBe(1);
      expect(result.current.state.tasks.get('task-1')).toMatchObject({
        id: 'task-1',
        title: 'My Task',
        status: 'running',
        progress: 0,
      });
    });

    it('sets task as active when no other task is active', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'First Task');
      });

      expect(result.current.activeTask?.id).toBe('task-1');
    });

    it('preserves existing active task when starting another', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'First Task');
      });

      act(() => {
        result.current.startTask('task-2', 'Second Task');
      });

      expect(result.current.activeTask?.id).toBe('task-1');
      expect(result.current.state.tasks.size).toBe(2);
    });

    it('stores estimated duration and cancellable options', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task', 'deck-generation', {
          estimatedDuration: 5000,
          cancellable: true,
        });
      });

      const task = result.current.state.tasks.get('task-1');
      expect(task?.estimatedDuration).toBe(5000);
      expect(task?.cancellable).toBe(true);
    });
  });

  describe('updateTask', () => {
    it('updates task progress', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.updateTask('task-1', 50);
      });

      expect(result.current.state.tasks.get('task-1')?.progress).toBe(50);
    });

    it('updates task detail', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.updateTask('task-1', 50, 'Processing step 2...');
      });

      expect(result.current.state.tasks.get('task-1')?.detail).toBe('Processing step 2...');
    });

    it('clamps progress to valid range', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.updateTask('task-1', 150);
      });
      expect(result.current.state.tasks.get('task-1')?.progress).toBe(100);

      act(() => {
        result.current.updateTask('task-1', -50);
      });
      expect(result.current.state.tasks.get('task-1')?.progress).toBe(-1); // -1 is valid for indeterminate
    });

    it('does not update completed tasks', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.completeTask('task-1');
      });

      act(() => {
        result.current.updateTask('task-1', 50);
      });

      // Task should still be at 100% after completion
      expect(result.current.state.tasks.get('task-1')?.progress).toBe(100);
    });
  });

  describe('completeTask', () => {
    it('marks task as completed', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.completeTask('task-1');
      });

      expect(result.current.state.tasks.get('task-1')?.status).toBe('completed');
      expect(result.current.state.tasks.get('task-1')?.progress).toBe(100);
    });

    it('clears active task when completed', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.completeTask('task-1');
      });

      expect(result.current.activeTask).toBeNull();
    });

    it('removes task after cleanup timeout', async () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.completeTask('task-1');
      });

      expect(result.current.state.tasks.has('task-1')).toBe(true);

      act(() => {
        vi.advanceTimersByTime(3500);
      });

      expect(result.current.state.tasks.has('task-1')).toBe(false);
    });
  });

  describe('failTask', () => {
    it('marks task as error with message', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.failTask('task-1', 'Something went wrong');
      });

      expect(result.current.state.tasks.get('task-1')?.status).toBe('error');
      expect(result.current.state.tasks.get('task-1')?.error).toBe('Something went wrong');
    });

    it('removes task after error cleanup timeout', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.failTask('task-1', 'Error');
      });

      expect(result.current.state.tasks.has('task-1')).toBe(true);

      act(() => {
        vi.advanceTimersByTime(5500);
      });

      expect(result.current.state.tasks.has('task-1')).toBe(false);
    });
  });

  describe('cancelTask', () => {
    it('marks task as cancelled', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.cancelTask('task-1');
      });

      expect(result.current.state.tasks.get('task-1')?.status).toBe('cancelled');
    });

    it('removes task after short cleanup timeout', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      act(() => {
        result.current.cancelTask('task-1');
      });

      act(() => {
        vi.advanceTimersByTime(1500);
      });

      expect(result.current.state.tasks.has('task-1')).toBe(false);
    });
  });

  describe('isRunning', () => {
    it('returns true when any task is running', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      expect(result.current.isRunning()).toBe(false);

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      expect(result.current.isRunning()).toBe(true);
    });

    it('filters by category', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'Deck Task', 'deck-generation');
      });

      expect(result.current.isRunning('deck-generation')).toBe(true);
      expect(result.current.isRunning('ml-training')).toBe(false);
    });
  });

  describe('getTasksByCategory', () => {
    it('returns tasks filtered by category', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'Deck Task', 'deck-generation');
        result.current.startTask('task-2', 'ML Task', 'ml-training');
        result.current.startTask('task-3', 'Another Deck Task', 'deck-generation');
      });

      const deckTasks = result.current.getTasksByCategory('deck-generation');
      expect(deckTasks).toHaveLength(2);
      expect(deckTasks.map((t) => t.id)).toContain('task-1');
      expect(deckTasks.map((t) => t.id)).toContain('task-3');
    });
  });

  describe('getEstimatedTimeRemaining', () => {
    it('returns undefined when no estimated duration', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task');
      });

      expect(result.current.getEstimatedTimeRemaining('task-1')).toBeUndefined();
    });

    it('estimates based on elapsed time and progress', () => {
      const { result } = renderHook(() => useTaskProgress(), {
        wrapper: TaskProgressProvider,
      });

      act(() => {
        result.current.startTask('task-1', 'My Task', 'general', { estimatedDuration: 10000 });
      });

      // Advance time and update progress
      act(() => {
        vi.advanceTimersByTime(5000);
      });

      const remaining = result.current.getEstimatedTimeRemaining('task-1');
      expect(remaining).toBeDefined();
      // Should be approximately 5000ms remaining (10000 - 5000 elapsed)
      expect(remaining).toBeLessThanOrEqual(5000);
    });
  });

  describe('UI integration', () => {
    it('renders with consumer component', () => {
      render(
        <TaskProgressProvider>
          <TestConsumer />
        </TaskProgressProvider>
      );

      expect(screen.getByTestId('task-count')).toHaveTextContent('0');
      expect(screen.getByTestId('is-running')).toHaveTextContent('idle');
      expect(screen.getByTestId('active-task')).toHaveTextContent('none');
    });

    it('updates UI when task is started', () => {
      render(
        <TaskProgressProvider>
          <TestConsumer />
        </TaskProgressProvider>
      );

      act(() => {
        screen.getByText('Start Task').click();
      });

      expect(screen.getByTestId('task-count')).toHaveTextContent('1');
      expect(screen.getByTestId('is-running')).toHaveTextContent('running');
      expect(screen.getByTestId('active-task')).toHaveTextContent('Test Task');
    });
  });
});

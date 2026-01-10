import { createContext, useContext, useState, useCallback, useRef, useEffect, type ReactNode } from 'react';
import { EventsOn } from '@/services/websocketClient';

// Task status types
export type TaskStatus = 'pending' | 'running' | 'completed' | 'error' | 'cancelled';

// Task categories for grouping and display
export type TaskCategory = 'deck-generation' | 'ml-training' | 'analysis' | 'sync' | 'general';

export interface TaskProgress {
  id: string;
  category: TaskCategory;
  title: string;
  status: TaskStatus;
  progress: number; // 0-100, -1 for indeterminate
  detail?: string;
  error?: string;
  startedAt: number;
  estimatedDuration?: number; // milliseconds
  cancellable?: boolean;
}

interface TaskProgressState {
  tasks: Map<string, TaskProgress>;
  activeTaskId: string | null;
}

// Context type
interface TaskProgressContextType {
  state: TaskProgressState;
  /** Start a new task */
  startTask: (
    id: string,
    title: string,
    category?: TaskCategory,
    options?: { estimatedDuration?: number; cancellable?: boolean }
  ) => void;
  /** Update task progress */
  updateTask: (id: string, progress: number, detail?: string) => void;
  /** Mark task as completed */
  completeTask: (id: string) => void;
  /** Mark task as failed */
  failTask: (id: string, error: string) => void;
  /** Cancel a task */
  cancelTask: (id: string) => void;
  /** Get estimated time remaining for a task */
  getEstimatedTimeRemaining: (id: string) => number | undefined;
  /** Check if any task in a category is running */
  isRunning: (category?: TaskCategory) => boolean;
  /** Get all tasks in a category */
  getTasksByCategory: (category: TaskCategory) => TaskProgress[];
  /** Get the currently active/visible task */
  activeTask: TaskProgress | null;
}

// Create context
const TaskProgressContext = createContext<TaskProgressContextType | undefined>(undefined);

// Provider component
interface TaskProgressProviderProps {
  children: ReactNode;
}

export const TaskProgressProvider = ({ children }: TaskProgressProviderProps) => {
  const [state, setState] = useState<TaskProgressState>({
    tasks: new Map(),
    activeTaskId: null,
  });

  // Track cleanup timeouts
  const cleanupTimeoutsRef = useRef<Map<string, NodeJS.Timeout>>(new Map());

  // Cleanup on unmount
  useEffect(() => {
    const timeouts = cleanupTimeoutsRef.current;
    return () => {
      timeouts.forEach((timeout) => clearTimeout(timeout));
      timeouts.clear();
    };
  }, []);

  // Start a new task
  const startTask = useCallback(
    (
      id: string,
      title: string,
      category: TaskCategory = 'general',
      options?: { estimatedDuration?: number; cancellable?: boolean }
    ) => {
      // Clear any pending cleanup for this task
      const existingTimeout = cleanupTimeoutsRef.current.get(id);
      if (existingTimeout) {
        clearTimeout(existingTimeout);
        cleanupTimeoutsRef.current.delete(id);
      }

      setState((prev) => {
        const newTasks = new Map(prev.tasks);
        newTasks.set(id, {
          id,
          category,
          title,
          status: 'running',
          progress: 0,
          startedAt: Date.now(),
          estimatedDuration: options?.estimatedDuration,
          cancellable: options?.cancellable,
        });
        return {
          tasks: newTasks,
          activeTaskId: prev.activeTaskId || id,
        };
      });
    },
    []
  );

  // Update task progress
  const updateTask = useCallback((id: string, progress: number, detail?: string) => {
    setState((prev) => {
      const task = prev.tasks.get(id);
      if (!task || task.status !== 'running') return prev;

      const newTasks = new Map(prev.tasks);
      newTasks.set(id, {
        ...task,
        progress: Math.min(100, Math.max(-1, progress)), // -1 is valid for indeterminate
        detail,
      });

      return {
        ...prev,
        tasks: newTasks,
      };
    });
  }, []);

  // Complete a task
  const completeTask = useCallback((id: string) => {
    setState((prev) => {
      const task = prev.tasks.get(id);
      if (!task) return prev;

      const newTasks = new Map(prev.tasks);
      newTasks.set(id, {
        ...task,
        status: 'completed',
        progress: 100,
      });

      // Find next running task for activeTaskId
      const nextRunning = Array.from(newTasks.values()).find(
        (t) => t.id !== id && t.status === 'running'
      );

      return {
        tasks: newTasks,
        activeTaskId: nextRunning?.id || null,
      };
    });

    // Schedule cleanup after 3 seconds
    const timeout = setTimeout(() => {
      setState((prev) => {
        const newTasks = new Map(prev.tasks);
        newTasks.delete(id);
        return { ...prev, tasks: newTasks };
      });
      cleanupTimeoutsRef.current.delete(id);
    }, 3000);

    cleanupTimeoutsRef.current.set(id, timeout);
  }, []);

  // Fail a task
  const failTask = useCallback((id: string, error: string) => {
    setState((prev) => {
      const task = prev.tasks.get(id);
      if (!task) return prev;

      const newTasks = new Map(prev.tasks);
      newTasks.set(id, {
        ...task,
        status: 'error',
        error,
      });

      // Find next running task for activeTaskId
      const nextRunning = Array.from(newTasks.values()).find(
        (t) => t.id !== id && t.status === 'running'
      );

      return {
        tasks: newTasks,
        activeTaskId: nextRunning?.id || null,
      };
    });

    // Schedule cleanup after 5 seconds
    const timeout = setTimeout(() => {
      setState((prev) => {
        const newTasks = new Map(prev.tasks);
        newTasks.delete(id);
        return { ...prev, tasks: newTasks };
      });
      cleanupTimeoutsRef.current.delete(id);
    }, 5000);

    cleanupTimeoutsRef.current.set(id, timeout);
  }, []);

  // Cancel a task
  const cancelTask = useCallback((id: string) => {
    // Clear any pending cleanup
    const existingTimeout = cleanupTimeoutsRef.current.get(id);
    if (existingTimeout) {
      clearTimeout(existingTimeout);
      cleanupTimeoutsRef.current.delete(id);
    }

    setState((prev) => {
      const task = prev.tasks.get(id);
      if (!task || task.status !== 'running') return prev;

      const newTasks = new Map(prev.tasks);
      newTasks.set(id, {
        ...task,
        status: 'cancelled',
      });

      // Find next running task
      const nextRunning = Array.from(newTasks.values()).find(
        (t) => t.id !== id && t.status === 'running'
      );

      return {
        tasks: newTasks,
        activeTaskId: nextRunning?.id || null,
      };
    });

    // Schedule immediate cleanup
    const timeout = setTimeout(() => {
      setState((prev) => {
        const newTasks = new Map(prev.tasks);
        newTasks.delete(id);
        return { ...prev, tasks: newTasks };
      });
      cleanupTimeoutsRef.current.delete(id);
    }, 1000);

    cleanupTimeoutsRef.current.set(id, timeout);
  }, []);

  // Get estimated time remaining
  const getEstimatedTimeRemaining = useCallback(
    (id: string): number | undefined => {
      const task = state.tasks.get(id);
      if (!task || task.status !== 'running' || !task.estimatedDuration) {
        return undefined;
      }

      const elapsed = Date.now() - task.startedAt;
      if (task.progress <= 0) {
        return task.estimatedDuration - elapsed;
      }

      // Estimate based on progress
      const progressRate = task.progress / elapsed;
      if (progressRate <= 0) return undefined;

      const remainingProgress = 100 - task.progress;
      return remainingProgress / progressRate;
    },
    [state.tasks]
  );

  // Check if any task in category is running
  const isRunning = useCallback(
    (category?: TaskCategory): boolean => {
      return Array.from(state.tasks.values()).some(
        (task) =>
          task.status === 'running' && (category === undefined || task.category === category)
      );
    },
    [state.tasks]
  );

  // Get tasks by category
  const getTasksByCategory = useCallback(
    (category: TaskCategory): TaskProgress[] => {
      return Array.from(state.tasks.values()).filter((task) => task.category === category);
    },
    [state.tasks]
  );

  // Get active task
  const activeTask = state.activeTaskId ? state.tasks.get(state.activeTaskId) || null : null;

  // Listen for WebSocket progress events
  useEffect(() => {
    const unsubscribeProgress = EventsOn('task:progress', (rawData: unknown) => {
      const data = rawData as {
        id: string;
        title?: string;
        category?: TaskCategory;
        progress: number;
        detail?: string;
        estimatedDuration?: number;
      };

      setState((prev) => {
        const existing = prev.tasks.get(data.id);
        const newTasks = new Map(prev.tasks);

        if (!existing) {
          // Create new task from WebSocket event
          newTasks.set(data.id, {
            id: data.id,
            category: data.category || 'general',
            title: data.title || 'Processing...',
            status: 'running',
            progress: Math.min(100, Math.max(-1, data.progress)),
            detail: data.detail,
            startedAt: Date.now(),
            estimatedDuration: data.estimatedDuration,
          });
        } else {
          // Update existing task
          newTasks.set(data.id, {
            ...existing,
            progress: Math.min(100, Math.max(-1, data.progress)),
            detail: data.detail,
          });
        }

        return {
          tasks: newTasks,
          activeTaskId: prev.activeTaskId || data.id,
        };
      });
    });

    const unsubscribeComplete = EventsOn('task:complete', (rawData: unknown) => {
      const data = rawData as { id: string };
      completeTask(data.id);
    });

    const unsubscribeError = EventsOn('task:error', (rawData: unknown) => {
      const data = rawData as { id: string; error: string };
      failTask(data.id, data.error);
    });

    return () => {
      unsubscribeProgress?.();
      unsubscribeComplete?.();
      unsubscribeError?.();
    };
  }, [completeTask, failTask]);

  const value: TaskProgressContextType = {
    state,
    startTask,
    updateTask,
    completeTask,
    failTask,
    cancelTask,
    getEstimatedTimeRemaining,
    isRunning,
    getTasksByCategory,
    activeTask,
  };

  return <TaskProgressContext.Provider value={value}>{children}</TaskProgressContext.Provider>;
};

// Custom hook to use the task progress context
// eslint-disable-next-line react-refresh/only-export-components
export const useTaskProgress = (): TaskProgressContextType => {
  const context = useContext(TaskProgressContext);
  if (context === undefined) {
    throw new Error('useTaskProgress must be used within a TaskProgressProvider');
  }
  return context;
};

export default TaskProgressContext;

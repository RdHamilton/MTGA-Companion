import { useCallback, useRef } from 'react';
import { useTaskProgress, type TaskCategory } from '@/context/TaskProgressContext';

interface UseProgressTaskOptions {
  category?: TaskCategory;
  estimatedDuration?: number;
  cancellable?: boolean;
}

interface ProgressTaskReturn<T> {
  /** Execute an async operation with progress tracking */
  execute: (
    operation: (updateProgress: (progress: number, detail?: string) => void) => Promise<T>,
    title: string
  ) => Promise<T>;
  /** Whether this task is currently running */
  isRunning: boolean;
  /** Current progress (0-100, -1 for indeterminate) */
  progress: number;
  /** Cancel the current task if cancellable */
  cancel: () => void;
}

/**
 * Hook for running async operations with automatic progress tracking.
 *
 * @param taskId - Unique identifier for this task
 * @param options - Configuration options
 * @returns Object with execute function and status
 *
 * @example
 * ```tsx
 * const { execute, isRunning } = useProgressTask('deck-generation', {
 *   category: 'deck-generation',
 *   estimatedDuration: 5000,
 * });
 *
 * const handleGenerate = async () => {
 *   const result = await execute(async (updateProgress) => {
 *     updateProgress(10, 'Analyzing seed card...');
 *     await analyzeSeed();
 *     updateProgress(40, 'Finding synergies...');
 *     const synergies = await findSynergies();
 *     updateProgress(70, 'Building deck...');
 *     const deck = await buildDeck(synergies);
 *     updateProgress(90, 'Finalizing...');
 *     return deck;
 *   }, 'Generating Deck');
 * };
 * ```
 */
export function useProgressTask<T = void>(
  taskId: string,
  options: UseProgressTaskOptions = {}
): ProgressTaskReturn<T> {
  const { category = 'general', estimatedDuration, cancellable = false } = options;
  const { startTask, updateTask, completeTask, failTask, cancelTask, state } = useTaskProgress();

  const abortControllerRef = useRef<AbortController | null>(null);

  const task = state.tasks.get(taskId);
  const isRunning = task?.status === 'running';
  const progress = task?.progress ?? 0;

  const execute = useCallback(
    async (
      operation: (updateProgress: (progress: number, detail?: string) => void) => Promise<T>,
      title: string
    ): Promise<T> => {
      // Create abort controller for cancellation
      abortControllerRef.current = new AbortController();

      // Start tracking
      startTask(taskId, title, category, { estimatedDuration, cancellable });

      const updateProgress = (prog: number, detail?: string) => {
        updateTask(taskId, prog, detail);
      };

      try {
        const result = await operation(updateProgress);
        completeTask(taskId);
        return result;
      } catch (error) {
        if (error instanceof Error && error.name === 'AbortError') {
          cancelTask(taskId);
          throw error;
        }
        failTask(taskId, error instanceof Error ? error.message : 'Operation failed');
        throw error;
      } finally {
        abortControllerRef.current = null;
      }
    },
    [taskId, category, estimatedDuration, cancellable, startTask, updateTask, completeTask, failTask, cancelTask]
  );

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    cancelTask(taskId);
  }, [taskId, cancelTask]);

  return {
    execute,
    isRunning,
    progress,
    cancel,
  };
}

export default useProgressTask;

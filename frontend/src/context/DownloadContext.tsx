import { createContext, useContext, useState, useEffect, useCallback, useRef, type ReactNode } from 'react';
import { EventsOn } from '@/services/websocketClient';

// Download state types
export type DownloadStatus = 'idle' | 'downloading' | 'complete' | 'error';

export interface DownloadTask {
  id: string;
  description: string;
  progress: number; // 0-100
  status: DownloadStatus;
  error?: string;
}

interface DownloadState {
  tasks: DownloadTask[];
  activeTask: DownloadTask | null;
}

// Context type
interface DownloadContextType {
  state: DownloadState;
  startDownload: (id: string, description: string) => void;
  updateProgress: (id: string, progress: number) => void;
  completeDownload: (id: string) => void;
  failDownload: (id: string, error: string) => void;
  cancelDownload: (id: string) => void;
  isDownloading: boolean;
  overallProgress: number;
}

// Create context
const DownloadContext = createContext<DownloadContextType | undefined>(undefined);

// Provider component
interface DownloadProviderProps {
  children: ReactNode;
}

export const DownloadProvider = ({ children }: DownloadProviderProps) => {
  const [state, setState] = useState<DownloadState>({
    tasks: [],
    activeTask: null,
  });

  // Track error task removal timeouts for cleanup
  const errorTimeoutsRef = useRef<Map<string, NodeJS.Timeout>>(new Map());

  // Cleanup timeouts on unmount
  useEffect(() => {
    const timeoutsMap = errorTimeoutsRef.current;
    return () => {
      timeoutsMap.forEach((timeout) => clearTimeout(timeout));
      timeoutsMap.clear();
    };
  }, []);

  // Start a new download task
  const startDownload = useCallback((id: string, description: string) => {
    setState((prev) => {
      // Check if task already exists
      const existingIndex = prev.tasks.findIndex((t) => t.id === id);
      const newTask: DownloadTask = {
        id,
        description,
        progress: 0,
        status: 'downloading',
      };

      let newTasks: DownloadTask[];
      if (existingIndex >= 0) {
        // Update existing task
        newTasks = [...prev.tasks];
        newTasks[existingIndex] = newTask;
      } else {
        // Add new task
        newTasks = [...prev.tasks, newTask];
      }

      return {
        tasks: newTasks,
        // Only set activeTask if there isn't one already
        activeTask: prev.activeTask || newTask,
      };
    });
  }, []);

  // Update progress for a task
  const updateProgress = useCallback((id: string, progress: number) => {
    setState((prev) => {
      const taskIndex = prev.tasks.findIndex((t) => t.id === id);
      if (taskIndex < 0) return prev;

      const newTasks = [...prev.tasks];
      newTasks[taskIndex] = {
        ...newTasks[taskIndex],
        progress: Math.min(100, Math.max(0, progress)),
      };

      // Only update activeTask if it's null or matches the current task
      const shouldUpdateActive = !prev.activeTask || prev.activeTask.id === id;

      return {
        tasks: newTasks,
        activeTask: shouldUpdateActive ? newTasks[taskIndex] : prev.activeTask,
      };
    });
  }, []);

  // Complete a download task
  const completeDownload = useCallback((id: string) => {
    setState((prev) => {
      const newTasks = prev.tasks.filter((t) => t.id !== id);
      const nextActive = newTasks.find((t) => t.status === 'downloading') || null;

      return {
        tasks: newTasks,
        activeTask: nextActive,
      };
    });
  }, []);

  // Fail a download task
  const failDownload = useCallback((id: string, error: string) => {
    setState((prev) => {
      const taskIndex = prev.tasks.findIndex((t) => t.id === id);
      if (taskIndex < 0) return prev;

      const newTasks = [...prev.tasks];
      newTasks[taskIndex] = {
        ...newTasks[taskIndex],
        status: 'error',
        error,
      };

      // Find next downloading task for activeTask
      const nextActive = newTasks.find((t) => t.status === 'downloading') || null;

      return {
        tasks: newTasks,
        activeTask: nextActive,
      };
    });

    // Schedule removal of error task after 5 seconds (outside setState)
    // Clear any existing timeout for this id
    const existingTimeout = errorTimeoutsRef.current.get(id);
    if (existingTimeout) {
      clearTimeout(existingTimeout);
    }

    const timeoutId = setTimeout(() => {
      setState((current) => ({
        ...current,
        tasks: current.tasks.filter((t) => t.id !== id),
      }));
      errorTimeoutsRef.current.delete(id);
    }, 5000);

    errorTimeoutsRef.current.set(id, timeoutId);
  }, []);

  // Cancel a download task
  const cancelDownload = useCallback((id: string) => {
    // Clear any error timeout for this task
    const existingTimeout = errorTimeoutsRef.current.get(id);
    if (existingTimeout) {
      clearTimeout(existingTimeout);
      errorTimeoutsRef.current.delete(id);
    }

    setState((prev) => {
      const newTasks = prev.tasks.filter((t) => t.id !== id);
      const nextActive = newTasks.find((t) => t.status === 'downloading') || null;

      return {
        tasks: newTasks,
        activeTask: nextActive,
      };
    });
  }, []);

  // Listen for download progress WebSocket events
  useEffect(() => {
    const unsubscribeProgress = EventsOn('download:progress', (rawData: unknown) => {
      const data = rawData as { id: string; description: string; progress: number };
      // Handle progress update inline to avoid nested setState calls
      setState((prev) => {
        const existing = prev.tasks.find((t) => t.id === data.id);
        if (!existing) {
          // Create new task
          const newTask: DownloadTask = {
            id: data.id,
            description: data.description,
            progress: Math.min(100, Math.max(0, data.progress)),
            status: 'downloading',
          };
          return {
            tasks: [...prev.tasks, newTask],
            activeTask: prev.activeTask || newTask,
          };
        } else {
          // Update existing task progress
          const taskIndex = prev.tasks.findIndex((t) => t.id === data.id);
          const newTasks = [...prev.tasks];
          newTasks[taskIndex] = {
            ...newTasks[taskIndex],
            progress: Math.min(100, Math.max(0, data.progress)),
          };
          const shouldUpdateActive = !prev.activeTask || prev.activeTask.id === data.id;
          return {
            tasks: newTasks,
            activeTask: shouldUpdateActive ? newTasks[taskIndex] : prev.activeTask,
          };
        }
      });
    });

    const unsubscribeComplete = EventsOn('download:complete', (rawData: unknown) => {
      const data = rawData as { id: string };
      completeDownload(data.id);
    });

    const unsubscribeError = EventsOn('download:error', (rawData: unknown) => {
      const data = rawData as { id: string; error: string };
      failDownload(data.id, data.error);
    });

    return () => {
      unsubscribeProgress?.();
      unsubscribeComplete?.();
      unsubscribeError?.();
    };
  }, [completeDownload, failDownload]);

  // Computed values - only count downloading tasks for progress
  const isDownloading = state.tasks.some((t) => t.status === 'downloading');
  const downloadingTasks = state.tasks.filter((t) => t.status === 'downloading');
  const overallProgress = downloadingTasks.length > 0
    ? downloadingTasks.reduce((sum, t) => sum + t.progress, 0) / downloadingTasks.length
    : 0;

  const value: DownloadContextType = {
    state,
    startDownload,
    updateProgress,
    completeDownload,
    failDownload,
    cancelDownload,
    isDownloading,
    overallProgress,
  };

  return <DownloadContext.Provider value={value}>{children}</DownloadContext.Provider>;
};

// Custom hook to use the download context
// eslint-disable-next-line react-refresh/only-export-components
export const useDownload = (): DownloadContextType => {
  const context = useContext(DownloadContext);
  if (context === undefined) {
    throw new Error('useDownload must be used within a DownloadProvider');
  }
  return context;
};

export default DownloadContext;

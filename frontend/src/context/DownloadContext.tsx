import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
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
        activeTask: newTask,
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

      return {
        tasks: newTasks,
        activeTask: newTasks[taskIndex],
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

      // Remove error tasks after 5 seconds
      setTimeout(() => {
        setState((current) => ({
          ...current,
          tasks: current.tasks.filter((t) => t.id !== id),
        }));
      }, 5000);

      return {
        tasks: newTasks,
        activeTask: newTasks[taskIndex],
      };
    });
  }, []);

  // Cancel a download task
  const cancelDownload = useCallback((id: string) => {
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
    const unsubscribeProgress = EventsOn('download:progress', (data: { id: string; description: string; progress: number }) => {
      // Start download if not already started
      setState((prev) => {
        const existing = prev.tasks.find((t) => t.id === data.id);
        if (!existing) {
          startDownload(data.id, data.description);
        }
        return prev;
      });
      updateProgress(data.id, data.progress);
    });

    const unsubscribeComplete = EventsOn('download:complete', (data: { id: string }) => {
      completeDownload(data.id);
    });

    const unsubscribeError = EventsOn('download:error', (data: { id: string; error: string }) => {
      failDownload(data.id, data.error);
    });

    return () => {
      unsubscribeProgress?.();
      unsubscribeComplete?.();
      unsubscribeError?.();
    };
  }, [startDownload, updateProgress, completeDownload, failDownload]);

  // Computed values
  const isDownloading = state.tasks.some((t) => t.status === 'downloading');
  const overallProgress = state.tasks.length > 0
    ? state.tasks.reduce((sum, t) => sum + t.progress, 0) / state.tasks.length
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

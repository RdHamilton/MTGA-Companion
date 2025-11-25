import { useState, useEffect, useCallback } from 'react';
import { TriggerReplayLogs } from '../../wailsjs/go/main/App';
import { EventsOn, WindowReloadApp } from '../../wailsjs/runtime/runtime';
import { showToast } from '../components/ToastContainer';
import { gui } from '../../wailsjs/go/models';

export interface UseLogReplayReturn {
  /** Whether to clear data before replay */
  clearDataBeforeReplay: boolean;
  /** Set clear data option */
  setClearDataBeforeReplay: (value: boolean) => void;
  /** Whether replay is currently in progress */
  isReplaying: boolean;
  /** Current replay progress */
  replayProgress: gui.LogReplayProgress | null;
  /** Start replaying logs (requires daemon connection) */
  handleReplayLogs: (isConnected: boolean) => Promise<void>;
}

export function useLogReplay(): UseLogReplayReturn {
  const [clearDataBeforeReplay, setClearDataBeforeReplay] = useState(false);
  const [isReplaying, setIsReplaying] = useState(false);
  const [replayProgress, setReplayProgress] = useState<gui.LogReplayProgress | null>(null);

  useEffect(() => {
    const unsubscribeStarted = EventsOn('replay:started', () => {
      setIsReplaying(true);
      setReplayProgress(null);
    });

    const unsubscribeProgress = EventsOn('replay:progress', (data: unknown) => {
      setReplayProgress(gui.LogReplayProgress.createFrom(data));
    });

    const unsubscribeCompleted = EventsOn('replay:completed', (data: unknown) => {
      setIsReplaying(false);
      setReplayProgress(gui.LogReplayProgress.createFrom(data));
      // Keep progress visible for a moment, then reload using Wails native method
      setTimeout(() => {
        WindowReloadApp(); // Refresh to show updated data
      }, 2000);
    });

    const unsubscribeError = EventsOn('replay:error', () => {
      setIsReplaying(false);
    });

    return () => {
      unsubscribeStarted();
      unsubscribeProgress();
      unsubscribeCompleted();
      unsubscribeError();
    };
  }, []);

  const handleReplayLogs = useCallback(async (isConnected: boolean) => {
    if (!isConnected) {
      return;
    }

    try {
      await TriggerReplayLogs(clearDataBeforeReplay);
      // Progress UI will update automatically from events
    } catch (error) {
      showToast.show(`Failed to trigger replay: ${error}`, 'error');
    }
  }, [clearDataBeforeReplay]);

  return {
    clearDataBeforeReplay,
    setClearDataBeforeReplay,
    isReplaying,
    replayProgress,
    handleReplayLogs,
  };
}

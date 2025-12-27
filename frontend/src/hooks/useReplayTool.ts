import { useState, useEffect, useCallback } from 'react';
import { subscribeToReplayState, getReplayState } from '../App';
import { showToast } from '../components/ToastContainer';
import { gui } from '@/types/models';

// No-op stubs - replay functions not implemented in REST API
/* eslint-disable @typescript-eslint/no-unused-vars */
async function startReplayWithFileDialog(
  _speed: number,
  _filter: string,
  _pauseOnDraft: boolean
): Promise<void> {
/* eslint-enable @typescript-eslint/no-unused-vars */
  console.warn('StartReplayWithFileDialog: Not implemented in REST API');
}

async function pauseReplay(): Promise<void> {
  console.warn('PauseReplay: Not implemented in REST API');
}

async function resumeReplay(): Promise<void> {
  console.warn('ResumeReplay: Not implemented in REST API');
}

async function stopReplay(): Promise<void> {
  console.warn('StopReplay: Not implemented in REST API');
}

export interface UseReplayToolReturn {
  /** Whether replay tool is active */
  replayToolActive: boolean;
  /** Whether replay tool is paused */
  replayToolPaused: boolean;
  /** Current replay tool progress */
  replayToolProgress: gui.ReplayStatus | null;
  /** Replay speed multiplier */
  replaySpeed: number;
  /** Set replay speed */
  setReplaySpeed: (speed: number) => void;
  /** Event filter for replay */
  replayFilter: string;
  /** Set replay filter */
  setReplayFilter: (filter: string) => void;
  /** Whether to pause on draft events */
  pauseOnDraft: boolean;
  /** Set pause on draft */
  setPauseOnDraft: (pause: boolean) => void;
  /** Start replay tool */
  handleStartReplayTool: (isConnected: boolean) => Promise<void>;
  /** Pause replay tool */
  handlePauseReplayTool: () => Promise<void>;
  /** Resume replay tool */
  handleResumeReplayTool: () => Promise<void>;
  /** Stop replay tool */
  handleStopReplayTool: () => Promise<void>;
}

export function useReplayTool(): UseReplayToolReturn {
  // Use global state for active/paused to persist across navigation
  const [replayToolActive, setReplayToolActive] = useState(getReplayState().isActive);
  const [replayToolPaused, setReplayToolPaused] = useState(getReplayState().isPaused);
  const [replayToolProgress, setReplayToolProgress] = useState<gui.ReplayStatus | null>(
    getReplayState().progress
  );
  const [replaySpeed, setReplaySpeed] = useState(1.0);
  const [replayFilter, setReplayFilter] = useState('all');
  const [pauseOnDraft, setPauseOnDraft] = useState(false);

  useEffect(() => {
    // Subscribe to replay state changes from the global state manager
    // Initial state is already set via useState initializers above
    const unsubscribe = subscribeToReplayState((state) => {
      setReplayToolActive(state.isActive);
      setReplayToolPaused(state.isPaused);
      setReplayToolProgress(state.progress);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  const handleStartReplayTool = useCallback(async (isConnected: boolean) => {
    if (!isConnected) {
      showToast.show('Replay tool requires daemon mode. Please start the daemon service.', 'warning');
      return;
    }

    try {
      await startReplayWithFileDialog(replaySpeed, replayFilter, pauseOnDraft);
    } catch (error) {
      showToast.show(`Failed to start replay: ${error}`, 'error');
    }
  }, [replaySpeed, replayFilter, pauseOnDraft]);

  const handlePauseReplayTool = useCallback(async () => {
    try {
      await pauseReplay();
    } catch (error) {
      showToast.show(`Failed to pause replay: ${error}`, 'error');
    }
  }, []);

  const handleResumeReplayTool = useCallback(async () => {
    try {
      await resumeReplay();
    } catch (error) {
      showToast.show(`Failed to resume replay: ${error}`, 'error');
    }
  }, []);

  const handleStopReplayTool = useCallback(async () => {
    try {
      await stopReplay();
    } catch (error) {
      showToast.show(`Failed to stop replay: ${error}`, 'error');
    }
  }, []);

  return {
    replayToolActive,
    replayToolPaused,
    replayToolProgress,
    replaySpeed,
    setReplaySpeed,
    replayFilter,
    setReplayFilter,
    pauseOnDraft,
    setPauseOnDraft,
    handleStartReplayTool,
    handlePauseReplayTool,
    handleResumeReplayTool,
    handleStopReplayTool,
  };
}

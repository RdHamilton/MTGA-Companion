import { gui } from '@/types/models';

// Global replay state interface
export interface ReplayState {
  isActive: boolean;
  isPaused: boolean;
  progress: gui.ReplayStatus | null;
}

const initialReplayState: ReplayState = {
  isActive: false,
  isPaused: false,
  progress: null,
};

// Global replay state - accessible across all components
let globalReplayState: ReplayState = { ...initialReplayState };
const replayStateListeners: Array<(state: ReplayState) => void> = [];

export const getReplayState = (): ReplayState => globalReplayState;

export const subscribeToReplayState = (listener: (state: ReplayState) => void): (() => void) => {
  replayStateListeners.push(listener);
  return () => {
    const index = replayStateListeners.indexOf(listener);
    if (index > -1) {
      replayStateListeners.splice(index, 1);
    }
  };
};

export const updateReplayState = (updates: Partial<ReplayState>): void => {
  globalReplayState = { ...globalReplayState, ...updates };
  console.log('[Global Replay State] Updated:', globalReplayState, 'Listeners:', replayStateListeners.length);
  replayStateListeners.forEach(listener => listener(globalReplayState));
};

export const resetReplayState = (): void => {
  globalReplayState = { ...initialReplayState };
  replayStateListeners.forEach(listener => listener(globalReplayState));
};

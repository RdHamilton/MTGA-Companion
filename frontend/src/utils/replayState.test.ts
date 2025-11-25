import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  getReplayState,
  subscribeToReplayState,
  updateReplayState,
  resetReplayState,
} from './replayState';
import { gui } from '../../wailsjs/go/models';

describe('replayState', () => {
  beforeEach(() => {
    // Reset state before each test
    resetReplayState();
    vi.clearAllMocks();
  });

  describe('getReplayState', () => {
    it('should return initial state', () => {
      const state = getReplayState();

      expect(state).toEqual({
        isActive: false,
        isPaused: false,
        progress: null,
      });
    });

    it('should return updated state after update', () => {
      updateReplayState({ isActive: true });

      const state = getReplayState();
      expect(state.isActive).toBe(true);
      expect(state.isPaused).toBe(false);
      expect(state.progress).toBeNull();
    });
  });

  describe('updateReplayState', () => {
    it('should update isActive state', () => {
      updateReplayState({ isActive: true });

      expect(getReplayState().isActive).toBe(true);
    });

    it('should update isPaused state', () => {
      updateReplayState({ isPaused: true });

      expect(getReplayState().isPaused).toBe(true);
    });

    it('should update progress state', () => {
      const progress = new gui.ReplayStatus({
        isActive: true,
        isPaused: false,
        currentEntry: 10,
        totalEntries: 100,
        percentComplete: 10,
        elapsed: '00:01:00',
        speed: 1.0,
        filter: 'all',
      });

      updateReplayState({ progress });

      expect(getReplayState().progress).toBe(progress);
    });

    it('should merge partial updates', () => {
      updateReplayState({ isActive: true });
      updateReplayState({ isPaused: true });

      const state = getReplayState();
      expect(state.isActive).toBe(true);
      expect(state.isPaused).toBe(true);
    });

    it('should not overwrite unspecified fields', () => {
      updateReplayState({ isActive: true, isPaused: true });
      updateReplayState({ isPaused: false });

      const state = getReplayState();
      expect(state.isActive).toBe(true);
      expect(state.isPaused).toBe(false);
    });
  });

  describe('subscribeToReplayState', () => {
    it('should call listener when state updates', () => {
      const listener = vi.fn();
      subscribeToReplayState(listener);

      updateReplayState({ isActive: true });

      expect(listener).toHaveBeenCalledTimes(1);
      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({ isActive: true })
      );
    });

    it('should call multiple listeners on update', () => {
      const listener1 = vi.fn();
      const listener2 = vi.fn();

      subscribeToReplayState(listener1);
      subscribeToReplayState(listener2);

      updateReplayState({ isActive: true });

      expect(listener1).toHaveBeenCalledTimes(1);
      expect(listener2).toHaveBeenCalledTimes(1);
    });

    it('should return unsubscribe function', () => {
      const listener = vi.fn();
      const unsubscribe = subscribeToReplayState(listener);

      unsubscribe();
      updateReplayState({ isActive: true });

      expect(listener).not.toHaveBeenCalled();
    });

    it('should only remove specific listener on unsubscribe', () => {
      const listener1 = vi.fn();
      const listener2 = vi.fn();

      const unsubscribe1 = subscribeToReplayState(listener1);
      subscribeToReplayState(listener2);

      unsubscribe1();
      updateReplayState({ isActive: true });

      expect(listener1).not.toHaveBeenCalled();
      expect(listener2).toHaveBeenCalledTimes(1);
    });

    it('should handle unsubscribe called multiple times', () => {
      const listener = vi.fn();
      const unsubscribe = subscribeToReplayState(listener);

      unsubscribe();
      unsubscribe(); // Should not throw

      updateReplayState({ isActive: true });
      expect(listener).not.toHaveBeenCalled();
    });
  });

  describe('resetReplayState', () => {
    it('should reset state to initial values', () => {
      updateReplayState({ isActive: true, isPaused: true });

      resetReplayState();

      const state = getReplayState();
      expect(state).toEqual({
        isActive: false,
        isPaused: false,
        progress: null,
      });
    });

    it('should notify listeners when reset', () => {
      const listener = vi.fn();
      subscribeToReplayState(listener);

      updateReplayState({ isActive: true });
      listener.mockClear();

      resetReplayState();

      expect(listener).toHaveBeenCalledTimes(1);
      expect(listener).toHaveBeenCalledWith({
        isActive: false,
        isPaused: false,
        progress: null,
      });
    });

    it('should clear progress data on reset', () => {
      const progress = new gui.ReplayStatus({
        isActive: true,
        isPaused: false,
        currentEntry: 50,
        totalEntries: 100,
        percentComplete: 50,
        elapsed: '00:05:00',
        speed: 2.0,
        filter: 'draft',
      });

      updateReplayState({ isActive: true, progress });
      resetReplayState();

      expect(getReplayState().progress).toBeNull();
    });
  });

  describe('state immutability', () => {
    it('should return new object reference on update', () => {
      const state1 = getReplayState();
      updateReplayState({ isActive: true });
      const state2 = getReplayState();

      expect(state1).not.toBe(state2);
    });

    it('should maintain state integrity across multiple operations', () => {
      const listener = vi.fn();
      subscribeToReplayState(listener);

      // Simulate a replay session
      updateReplayState({ isActive: true });
      updateReplayState({ isPaused: true });
      updateReplayState({ isPaused: false });
      updateReplayState({ isActive: false });

      expect(listener).toHaveBeenCalledTimes(4);

      const finalState = getReplayState();
      expect(finalState.isActive).toBe(false);
      expect(finalState.isPaused).toBe(false);
    });
  });
});

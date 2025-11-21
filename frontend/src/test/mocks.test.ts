import { describe, it, expect, beforeEach } from 'vitest';
import {
  mockWailsRuntime,
  mockEventEmitter,
  mockWailsApp,
  resetMocks,
} from './index';

describe('Wails Mocks', () => {
  beforeEach(() => {
    mockEventEmitter.clear();
    resetMocks();
  });

  describe('Runtime Mocks', () => {
    it('should mock EventsOn and EventsEmit', () => {
      const callback = vi.fn();
      mockWailsRuntime.EventsOn('test-event', callback);
      mockWailsRuntime.EventsEmit('test-event', 'test-data');

      expect(callback).toHaveBeenCalledWith('test-data');
    });

    it('should mock EventsOnce', () => {
      const callback = vi.fn();
      mockWailsRuntime.EventsOnce('test-event', callback);

      mockWailsRuntime.EventsEmit('test-event', 'first');
      mockWailsRuntime.EventsEmit('test-event', 'second');

      expect(callback).toHaveBeenCalledTimes(1);
      expect(callback).toHaveBeenCalledWith('first');
    });

    it('should mock EventsOnMultiple', () => {
      const callback = vi.fn();
      mockWailsRuntime.EventsOnMultiple('test-event', callback, 2);

      mockWailsRuntime.EventsEmit('test-event', '1');
      mockWailsRuntime.EventsEmit('test-event', '2');
      mockWailsRuntime.EventsEmit('test-event', '3');

      expect(callback).toHaveBeenCalledTimes(2);
    });

    it('should mock EventsOff', () => {
      const callback = vi.fn();
      mockWailsRuntime.EventsOn('test-event', callback);
      mockWailsRuntime.EventsOff('test-event');
      mockWailsRuntime.EventsEmit('test-event', 'data');

      expect(callback).not.toHaveBeenCalled();
    });

    it('should mock window functions', async () => {
      expect(await mockWailsRuntime.WindowGetSize()).toEqual({ w: 1024, h: 768 });
      expect(await mockWailsRuntime.WindowIsFullscreen()).toBe(false);
      expect(await mockWailsRuntime.WindowIsMaximised()).toBe(false);
    });

    it('should mock environment', async () => {
      const env = await mockWailsRuntime.Environment();
      expect(env).toEqual({
        buildType: 'dev',
        platform: 'darwin',
        arch: 'amd64',
      });
    });
  });

  describe('App Mocks', () => {
    it('should mock GetActiveDraftSessions', async () => {
      const result = await mockWailsApp.GetActiveDraftSessions();
      expect(result).toEqual([]);
      expect(mockWailsApp.GetActiveDraftSessions).toHaveBeenCalled();
    });

    it('should mock GetMatches', async () => {
      const result = await mockWailsApp.GetMatches();
      expect(result).toEqual([]);
      expect(mockWailsApp.GetMatches).toHaveBeenCalled();
    });

    it('should reset mocks', () => {
      mockWailsApp.GetActiveDraftSessions();
      resetMocks();
      expect(mockWailsApp.GetActiveDraftSessions).not.toHaveBeenCalled();
    });
  });
});

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Unmock the websocketClient so we can test the actual implementation
vi.unmock('@/services/websocketClient');

import {
  configureWebSocket,
  getWebSocketConfig,
  EventsOn,
  EventsOnce,
  EventsOff,
  EventsEmit,
  getListenerCount,
  getRegisteredEventTypes,
} from '../websocketClient';

describe('websocketClient', () => {
  beforeEach(() => {
    // Reset configuration
    configureWebSocket({
      url: 'ws://localhost:8080/ws',
      reconnectInterval: 3000,
      maxReconnectAttempts: 10,
    });

    // Clear all event listeners
    const eventTypes = getRegisteredEventTypes();
    if (eventTypes.length > 0) {
      EventsOff(eventTypes[0], ...eventTypes.slice(1));
    }
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('configureWebSocket', () => {
    it('should update WebSocket configuration', () => {
      configureWebSocket({ url: 'ws://example.com/ws' });
      const config = getWebSocketConfig();
      expect(config.url).toBe('ws://example.com/ws');
    });

    it('should merge with existing config', () => {
      configureWebSocket({ reconnectInterval: 5000 });
      const config = getWebSocketConfig();
      expect(config.reconnectInterval).toBe(5000);
      expect(config.url).toBe('ws://localhost:8080/ws');
    });
  });

  describe('EventsOn', () => {
    it('should register an event listener', () => {
      const callback = vi.fn();
      EventsOn('test:event', callback);

      expect(getListenerCount('test:event')).toBe(1);
    });

    it('should allow multiple listeners for same event', () => {
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      EventsOn('test:event', callback1);
      EventsOn('test:event', callback2);

      expect(getListenerCount('test:event')).toBe(2);
    });

    it('should return unsubscribe function', () => {
      const callback = vi.fn();
      const unsubscribe = EventsOn('test:event', callback);

      expect(getListenerCount('test:event')).toBe(1);

      unsubscribe();

      expect(getListenerCount('test:event')).toBe(0);
    });

    it('should call callback when event is emitted', () => {
      const callback = vi.fn();
      EventsOn('test:event', callback);

      EventsEmit('test:event', { value: 42 });

      expect(callback).toHaveBeenCalledWith({ value: 42 });
    });

    it('should call all listeners when event is emitted', () => {
      const callback1 = vi.fn();
      const callback2 = vi.fn();

      EventsOn('test:event', callback1);
      EventsOn('test:event', callback2);

      EventsEmit('test:event', { data: 'test' });

      expect(callback1).toHaveBeenCalledWith({ data: 'test' });
      expect(callback2).toHaveBeenCalledWith({ data: 'test' });
    });
  });

  describe('EventsOnce', () => {
    it('should only call callback once', () => {
      const callback = vi.fn();
      EventsOnce('test:event', callback);

      EventsEmit('test:event', { first: true });
      EventsEmit('test:event', { second: true });

      expect(callback).toHaveBeenCalledTimes(1);
      expect(callback).toHaveBeenCalledWith({ first: true });
    });

    it('should automatically unsubscribe after first event', () => {
      const callback = vi.fn();
      EventsOnce('test:event', callback);

      expect(getListenerCount('test:event')).toBe(1);

      EventsEmit('test:event', {});

      expect(getListenerCount('test:event')).toBe(0);
    });

    it('should return unsubscribe function', () => {
      const callback = vi.fn();
      const unsubscribe = EventsOnce('test:event', callback);

      unsubscribe();

      EventsEmit('test:event', {});

      expect(callback).not.toHaveBeenCalled();
    });
  });

  describe('EventsOff', () => {
    it('should remove all listeners for an event', () => {
      EventsOn('test:event', vi.fn());
      EventsOn('test:event', vi.fn());

      expect(getListenerCount('test:event')).toBe(2);

      EventsOff('test:event');

      expect(getListenerCount('test:event')).toBe(0);
    });

    it('should remove listeners for multiple events', () => {
      EventsOn('event1', vi.fn());
      EventsOn('event2', vi.fn());
      EventsOn('event3', vi.fn());

      EventsOff('event1', 'event2');

      expect(getListenerCount('event1')).toBe(0);
      expect(getListenerCount('event2')).toBe(0);
      expect(getListenerCount('event3')).toBe(1);
    });
  });

  describe('EventsEmit', () => {
    it('should not throw when no listeners', () => {
      expect(() => {
        EventsEmit('nonexistent', { data: 'test' });
      }).not.toThrow();
    });

    it('should handle listener errors gracefully', () => {
      const errorCallback = vi.fn(() => {
        throw new Error('Listener error');
      });
      const normalCallback = vi.fn();

      EventsOn('test:event', errorCallback);
      EventsOn('test:event', normalCallback);

      // Should not throw
      expect(() => {
        EventsEmit('test:event', {});
      }).not.toThrow();

      // Both callbacks should be attempted
      expect(errorCallback).toHaveBeenCalled();
      expect(normalCallback).toHaveBeenCalled();
    });
  });

  describe('wildcard listener', () => {
    it('should receive all events', () => {
      const wildcardCallback = vi.fn();
      EventsOn('*', wildcardCallback);

      EventsEmit('event1', { a: 1 });
      EventsEmit('event2', { b: 2 });

      expect(wildcardCallback).toHaveBeenCalledTimes(2);
      expect(wildcardCallback).toHaveBeenCalledWith({ type: 'event1', data: { a: 1 } });
      expect(wildcardCallback).toHaveBeenCalledWith({ type: 'event2', data: { b: 2 } });
    });
  });

  describe('getRegisteredEventTypes', () => {
    it('should return all registered event types', () => {
      EventsOn('event1', vi.fn());
      EventsOn('event2', vi.fn());
      EventsOn('event3', vi.fn());

      const types = getRegisteredEventTypes();

      expect(types).toContain('event1');
      expect(types).toContain('event2');
      expect(types).toContain('event3');
      expect(types).toHaveLength(3);
    });

    it('should return empty array when no listeners', () => {
      const types = getRegisteredEventTypes();
      expect(types).toEqual([]);
    });
  });

  describe('real-world event scenarios', () => {
    it('should handle stats:updated event', () => {
      const callback = vi.fn();
      EventsOn('stats:updated', callback);

      EventsEmit('stats:updated', { matches: 10, games: 25 });

      expect(callback).toHaveBeenCalledWith({ matches: 10, games: 25 });
    });

    it('should handle draft:updated event', () => {
      const callback = vi.fn();
      EventsOn('draft:updated', callback);

      EventsEmit('draft:updated', { count: 5, picks: 42 });

      expect(callback).toHaveBeenCalledWith({ count: 5, picks: 42 });
    });

    it('should handle replay:progress event', () => {
      const callback = vi.fn();
      EventsOn('replay:progress', callback);

      EventsEmit('replay:progress', {
        current: 50,
        total: 100,
        percentage: 50.0,
        currentFile: 'log1.log',
      });

      expect(callback).toHaveBeenCalledWith({
        current: 50,
        total: 100,
        percentage: 50.0,
        currentFile: 'log1.log',
      });
    });
  });
});

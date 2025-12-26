/**
 * WebSocket client for real-time event handling.
 * Replaces Wails EventsOn/EventsOff/EventsEmit functionality.
 */

export interface WebSocketConfig {
  url: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export interface WebSocketEvent {
  type: string;
  data: unknown;
}

type EventCallback = (data: unknown) => void;

// Default configuration
let config: WebSocketConfig = {
  url: 'ws://localhost:8080/ws',
  reconnectInterval: 3000,
  maxReconnectAttempts: 10,
};

// Connection state
let socket: WebSocket | null = null;
let reconnectAttempts = 0;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
let isIntentionalClose = false;

// Event listeners
const eventListeners: Map<string, Set<EventCallback>> = new Map();

// Connection state listeners
type ConnectionStateCallback = (connected: boolean) => void;
const connectionListeners: Set<ConnectionStateCallback> = new Set();

/**
 * Configure the WebSocket client.
 */
export function configureWebSocket(newConfig: Partial<WebSocketConfig>): void {
  config = { ...config, ...newConfig };
}

/**
 * Get the current WebSocket configuration.
 */
export function getWebSocketConfig(): WebSocketConfig {
  return { ...config };
}

/**
 * Check if WebSocket is connected.
 */
export function isConnected(): boolean {
  return socket?.readyState === WebSocket.OPEN;
}

/**
 * Connect to the WebSocket server.
 */
export function connect(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (socket?.readyState === WebSocket.OPEN) {
      resolve();
      return;
    }

    isIntentionalClose = false;

    try {
      socket = new WebSocket(config.url);

      socket.onopen = () => {
        console.log('[WebSocket] Connected to', config.url);
        reconnectAttempts = 0;
        notifyConnectionState(true);
        resolve();
      };

      socket.onclose = (event) => {
        console.log('[WebSocket] Disconnected:', event.code, event.reason);
        notifyConnectionState(false);

        if (!isIntentionalClose) {
          scheduleReconnect();
        }
      };

      socket.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
        if (socket?.readyState !== WebSocket.OPEN) {
          reject(new Error('WebSocket connection failed'));
        }
      };

      socket.onmessage = (event) => {
        try {
          const message: WebSocketEvent = JSON.parse(event.data);
          dispatchEvent(message.type, message.data);
        } catch (error) {
          console.error('[WebSocket] Failed to parse message:', error);
        }
      };
    } catch (error) {
      reject(error);
    }
  });
}

/**
 * Disconnect from the WebSocket server.
 */
export function disconnect(): void {
  isIntentionalClose = true;

  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }

  if (socket) {
    socket.close(1000, 'Client disconnecting');
    socket = null;
  }

  notifyConnectionState(false);
}

/**
 * Schedule a reconnection attempt.
 */
function scheduleReconnect(): void {
  if (reconnectAttempts >= (config.maxReconnectAttempts ?? 10)) {
    console.log('[WebSocket] Max reconnect attempts reached');
    return;
  }

  reconnectAttempts++;
  const delay = config.reconnectInterval ?? 3000;

  console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${reconnectAttempts})`);

  reconnectTimeout = setTimeout(() => {
    connect().catch((error) => {
      console.error('[WebSocket] Reconnect failed:', error);
    });
  }, delay);
}

/**
 * Notify connection state listeners.
 */
function notifyConnectionState(connected: boolean): void {
  connectionListeners.forEach((callback) => {
    try {
      callback(connected);
    } catch (error) {
      console.error('[WebSocket] Connection listener error:', error);
    }
  });
}

/**
 * Dispatch an event to listeners.
 */
function dispatchEvent(eventType: string, data: unknown): void {
  const listeners = eventListeners.get(eventType);
  if (listeners) {
    listeners.forEach((callback) => {
      try {
        callback(data);
      } catch (error) {
        console.error(`[WebSocket] Event listener error for ${eventType}:`, error);
      }
    });
  }

  // Also dispatch to wildcard listeners
  const wildcardListeners = eventListeners.get('*');
  if (wildcardListeners) {
    wildcardListeners.forEach((callback) => {
      try {
        callback({ type: eventType, data });
      } catch (error) {
        console.error('[WebSocket] Wildcard listener error:', error);
      }
    });
  }
}

/**
 * Subscribe to an event type.
 * Returns an unsubscribe function.
 *
 * This is the replacement for Wails EventsOn.
 */
export function EventsOn(eventType: string, callback: EventCallback): () => void {
  if (!eventListeners.has(eventType)) {
    eventListeners.set(eventType, new Set());
  }

  eventListeners.get(eventType)!.add(callback);

  // Return unsubscribe function
  return () => {
    const listeners = eventListeners.get(eventType);
    if (listeners) {
      listeners.delete(callback);
      if (listeners.size === 0) {
        eventListeners.delete(eventType);
      }
    }
  };
}

/**
 * Subscribe to an event type for a single occurrence.
 *
 * This is the replacement for Wails EventsOnce.
 */
export function EventsOnce(eventType: string, callback: EventCallback): () => void {
  const wrappedCallback = (data: unknown) => {
    unsubscribe();
    callback(data);
  };

  const unsubscribe = EventsOn(eventType, wrappedCallback);
  return unsubscribe;
}

/**
 * Unsubscribe from one or more event types.
 *
 * This is the replacement for Wails EventsOff.
 */
export function EventsOff(eventType: string, ...additionalEventTypes: string[]): void {
  const allTypes = [eventType, ...additionalEventTypes];
  allTypes.forEach((type) => {
    eventListeners.delete(type);
  });
}

/**
 * Emit an event (client-side only, for testing/mocking).
 * Note: In production, events come from the server.
 *
 * This can be used for testing or local event dispatch.
 */
export function EventsEmit(eventType: string, data?: unknown): void {
  dispatchEvent(eventType, data);
}

/**
 * Subscribe to connection state changes.
 * Returns an unsubscribe function.
 */
export function onConnectionChange(callback: ConnectionStateCallback): () => void {
  connectionListeners.add(callback);

  // Immediately notify of current state
  callback(isConnected());

  return () => {
    connectionListeners.delete(callback);
  };
}

/**
 * Get the count of listeners for an event type.
 * Useful for debugging.
 */
export function getListenerCount(eventType: string): number {
  return eventListeners.get(eventType)?.size ?? 0;
}

/**
 * Get all registered event types.
 * Useful for debugging.
 */
export function getRegisteredEventTypes(): string[] {
  return Array.from(eventListeners.keys());
}

/**
 * Reload the application.
 * In browser mode, this reloads the page.
 * This replaces Wails WindowReloadApp.
 */
export function WindowReloadApp(): void {
  window.location.reload();
}

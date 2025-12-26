/**
 * Services Index
 *
 * This module provides a centralized export point for all service modules,
 * including the API client, WebSocket client, and API service modules.
 *
 * These services are designed to replace Wails bindings, enabling:
 * - REST API communication instead of direct Go function calls
 * - WebSocket-based real-time events instead of Wails EventsOn/EventsOff
 * - Better testability with standard HTTP mocking
 * - Potential for running as a standalone web app (without Wails)
 */

// API Client
export {
  configureApi,
  getApiConfig,
  get,
  post,
  put,
  patch,
  del,
  healthCheck,
  ApiRequestError,
  type ApiConfig,
  type ApiResponse,
  type ApiError,
} from './apiClient';

// WebSocket Client
export {
  configureWebSocket,
  getWebSocketConfig,
  connect,
  disconnect,
  isConnected,
  EventsOn,
  EventsOnce,
  EventsOff,
  EventsEmit,
  onConnectionChange,
  getListenerCount,
  getRegisteredEventTypes,
  type WebSocketConfig,
  type WebSocketEvent,
} from './websocketClient';

// API Service Modules
export * as api from './api';

// Re-export individual API modules for convenience
export { matches, drafts, decks, cards, collection, system } from './api';

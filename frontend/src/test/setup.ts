import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';
import { mockWailsRuntime, mockEventEmitter } from './mocks/websocketMock';
import { mockWailsApp, resetMocks } from './mocks/apiMock';

// Mock WebSocket client globally
vi.mock('@/services/websocketClient', () => mockWailsRuntime);

// Mock API legacy bindings globally
vi.mock('@/services/api/legacy', () => mockWailsApp);

// Cleanup after each test
afterEach(() => {
  cleanup();
  mockEventEmitter.clear();
  resetMocks();
});

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {}, // deprecated
    removeListener: () => {}, // deprecated
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => true,
  }),
});

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  takeRecords() {
    return [];
  }
  unobserve() {}
} as any;

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
} as any;

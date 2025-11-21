import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';
import { mockWailsRuntime, mockEventEmitter } from './mocks/wailsRuntime';
import { mockWailsApp, resetMocks } from './mocks/wailsApp';

// Mock Wails runtime globally
vi.mock('../../wailsjs/runtime/runtime', () => mockWailsRuntime);

// Mock Wails App bindings globally
vi.mock('../../wailsjs/go/main/App', () => mockWailsApp);

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

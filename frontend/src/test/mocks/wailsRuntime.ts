import { vi } from 'vitest';

type EventCallback = (...data: any[]) => void;

class MockEventEmitter {
  private listeners: Map<string, EventCallback[]> = new Map();

  on(eventName: string, callback: EventCallback): () => void {
    if (!this.listeners.has(eventName)) {
      this.listeners.set(eventName, []);
    }
    this.listeners.get(eventName)!.push(callback);

    return () => this.off(eventName, callback);
  }

  off(eventName: string, callback?: EventCallback): void {
    if (!callback) {
      this.listeners.delete(eventName);
      return;
    }

    const callbacks = this.listeners.get(eventName);
    if (callbacks) {
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    }
  }

  emit(eventName: string, ...data: any[]): void {
    const callbacks = this.listeners.get(eventName);
    if (callbacks) {
      callbacks.forEach((callback) => callback(...data));
    }
  }

  clear(): void {
    this.listeners.clear();
  }
}

export const mockEventEmitter = new MockEventEmitter();

export const mockWailsRuntime = {
  EventsOn: vi.fn((eventName: string, callback: EventCallback) => {
    return mockEventEmitter.on(eventName, callback);
  }),

  EventsOnce: vi.fn((eventName: string, callback: EventCallback) => {
    const unsubscribe = mockEventEmitter.on(eventName, (...data: any[]) => {
      callback(...data);
      unsubscribe();
    });
    return unsubscribe;
  }),

  EventsOnMultiple: vi.fn(
    (eventName: string, callback: EventCallback, maxCallbacks: number) => {
      let count = 0;
      const unsubscribe = mockEventEmitter.on(eventName, (...data: any[]) => {
        if (count < maxCallbacks) {
          callback(...data);
          count++;
          if (count >= maxCallbacks) {
            unsubscribe();
          }
        }
      });
      return unsubscribe;
    }
  ),

  EventsEmit: vi.fn((eventName: string, ...data: any[]) => {
    mockEventEmitter.emit(eventName, ...data);
  }),

  EventsOff: vi.fn((eventName: string, ...additionalEventNames: string[]) => {
    mockEventEmitter.off(eventName);
    additionalEventNames.forEach((name) => mockEventEmitter.off(name));
  }),

  EventsOffAll: vi.fn(() => {
    mockEventEmitter.clear();
  }),

  LogPrint: vi.fn(),
  LogTrace: vi.fn(),
  LogDebug: vi.fn(),
  LogError: vi.fn(),
  LogFatal: vi.fn(),
  LogInfo: vi.fn(),
  LogWarning: vi.fn(),

  WindowReload: vi.fn(),
  WindowReloadApp: vi.fn(),
  WindowSetAlwaysOnTop: vi.fn(),
  WindowSetSystemDefaultTheme: vi.fn(),
  WindowSetLightTheme: vi.fn(),
  WindowSetDarkTheme: vi.fn(),
  WindowCenter: vi.fn(),
  WindowSetTitle: vi.fn(),
  WindowFullscreen: vi.fn(),
  WindowUnfullscreen: vi.fn(),
  WindowIsFullscreen: vi.fn(() => Promise.resolve(false)),
  WindowSetSize: vi.fn(),
  WindowGetSize: vi.fn(() => Promise.resolve({ w: 1024, h: 768 })),
  WindowSetMaxSize: vi.fn(),
  WindowSetMinSize: vi.fn(),
  WindowSetPosition: vi.fn(),
  WindowGetPosition: vi.fn(() => Promise.resolve({ x: 0, y: 0 })),
  WindowHide: vi.fn(),
  WindowShow: vi.fn(),
  WindowMaximise: vi.fn(),
  WindowToggleMaximise: vi.fn(),
  WindowUnmaximise: vi.fn(),
  WindowIsMaximised: vi.fn(() => Promise.resolve(false)),
  WindowMinimise: vi.fn(),
  WindowUnminimise: vi.fn(),
  WindowIsMinimised: vi.fn(() => Promise.resolve(false)),
  WindowIsNormal: vi.fn(() => Promise.resolve(true)),
  WindowSetBackgroundColour: vi.fn(),

  ScreenGetAll: vi.fn(() =>
    Promise.resolve([
      {
        isCurrent: true,
        isPrimary: true,
        width: 1920,
        height: 1080,
      },
    ])
  ),

  BrowserOpenURL: vi.fn(),

  Environment: vi.fn(() =>
    Promise.resolve({
      buildType: 'dev',
      platform: 'darwin',
      arch: 'amd64',
    })
  ),

  Quit: vi.fn(),
  Hide: vi.fn(),
  Show: vi.fn(),

  ClipboardGetText: vi.fn(() => Promise.resolve('')),
  ClipboardSetText: vi.fn(() => Promise.resolve(true)),

  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),

  CanResolveFilePaths: vi.fn(() => true),
  ResolveFilePaths: vi.fn(),
};

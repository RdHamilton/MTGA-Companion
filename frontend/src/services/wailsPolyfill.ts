/**
 * Polyfill for Wails runtime when running outside of Wails.
 * This provides no-op implementations to prevent runtime errors.
 */

interface WailsRuntime {
  LogPrint: (message: string) => void;
  LogTrace: (message: string) => void;
  LogDebug: (message: string) => void;
  LogInfo: (message: string) => void;
  LogWarning: (message: string) => void;
  LogError: (message: string) => void;
  LogFatal: (message: string) => void;
  EventsOnMultiple: (
    eventName: string,
    callback: (...data: unknown[]) => void,
    maxCallbacks: number
  ) => () => void;
  EventsOff: (eventName: string, ...additionalEventNames: string[]) => void;
  EventsOffAll: () => void;
  EventsEmit: (eventName: string, ...args: unknown[]) => void;
  WindowReload: () => void;
  WindowReloadApp: () => void;
  WindowSetAlwaysOnTop: (b: boolean) => void;
  WindowSetSystemDefaultTheme: () => void;
  WindowSetLightTheme: () => void;
  WindowSetDarkTheme: () => void;
  WindowCenter: () => void;
  WindowSetTitle: (title: string) => void;
  WindowFullscreen: () => void;
  WindowUnfullscreen: () => void;
  WindowIsFullscreen: () => boolean;
  WindowGetSize: () => { w: number; h: number };
  WindowSetSize: (width: number, height: number) => void;
  WindowSetMaxSize: (width: number, height: number) => void;
  WindowSetMinSize: (width: number, height: number) => void;
  WindowSetPosition: (x: number, y: number) => void;
  WindowGetPosition: () => { x: number; y: number };
  WindowHide: () => void;
  WindowShow: () => void;
  WindowMaximise: () => void;
  WindowToggleMaximise: () => void;
  WindowUnmaximise: () => void;
  WindowIsMaximised: () => boolean;
  WindowMinimise: () => void;
  WindowUnminimise: () => void;
  WindowSetBackgroundColour: (R: number, G: number, B: number, A: number) => void;
  ScreenGetAll: () => unknown[];
  WindowIsMinimised: () => boolean;
  WindowIsNormal: () => boolean;
  BrowserOpenURL: (url: string) => void;
  Environment: () => { buildType: string; platform: string; arch: string };
  Quit: () => void;
  Hide: () => void;
  Show: () => void;
  ClipboardGetText: () => Promise<string>;
  ClipboardSetText: (text: string) => Promise<void>;
  OnFileDrop: (
    callback: (x: number, y: number, paths: string[]) => void,
    useDropTarget?: boolean
  ) => () => void;
  OnFileDropOff: () => void;
  CanResolveFilePaths: () => boolean;
  ResolveFilePaths: (files: FileList) => Promise<string[]>;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type WailsGoApp = Record<string, (...args: any[]) => Promise<any>>;

declare global {
  interface Window {
    runtime?: WailsRuntime;
    go?: {
      main?: {
        App?: WailsGoApp;
      };
    };
  }
}

// No-op event listener storage for polyfill
const eventListeners: Map<string, Set<(...data: unknown[]) => void>> = new Map();

/**
 * Install the Wails polyfill if window.runtime doesn't exist.
 * This should be called before any Wails code runs.
 */
export function installWailsPolyfill(): void {
  if (typeof window === 'undefined') {
    return; // SSR or non-browser environment
  }

  if (window.runtime) {
    console.log('[WailsPolyfill] Wails runtime detected, skipping polyfill');
    return;
  }

  console.log('[WailsPolyfill] Installing Wails runtime polyfill');

  window.runtime = {
    // Logging - output to console
    LogPrint: (message) => console.log('[Wails]', message),
    LogTrace: (message) => console.trace('[Wails]', message),
    LogDebug: (message) => console.debug('[Wails]', message),
    LogInfo: (message) => console.info('[Wails]', message),
    LogWarning: (message) => console.warn('[Wails]', message),
    LogError: (message) => console.error('[Wails]', message),
    LogFatal: (message) => console.error('[Wails Fatal]', message),

    // Events - store locally (the adapter will handle real events via WebSocket)
    EventsOnMultiple: (eventName, callback, maxCallbacks) => {
      if (!eventListeners.has(eventName)) {
        eventListeners.set(eventName, new Set());
      }
      const listeners = eventListeners.get(eventName)!;
      let callCount = 0;

      const wrappedCallback = (...data: unknown[]) => {
        if (maxCallbacks > 0 && callCount >= maxCallbacks) {
          listeners.delete(wrappedCallback);
          return;
        }
        callCount++;
        callback(...data);
      };

      listeners.add(wrappedCallback);

      // Return unsubscribe function
      return () => {
        listeners.delete(wrappedCallback);
      };
    },

    EventsOff: (eventName, ...additionalEventNames) => {
      [eventName, ...additionalEventNames].forEach((name) => {
        eventListeners.delete(name);
      });
    },

    EventsOffAll: () => {
      eventListeners.clear();
    },

    EventsEmit: (eventName, ...args) => {
      const listeners = eventListeners.get(eventName);
      if (listeners) {
        listeners.forEach((callback) => callback(...args));
      }
    },

    // Window management - no-ops
    WindowReload: () => window.location.reload(),
    WindowReloadApp: () => window.location.reload(),
    WindowSetAlwaysOnTop: () => {},
    WindowSetSystemDefaultTheme: () => {},
    WindowSetLightTheme: () => {},
    WindowSetDarkTheme: () => {},
    WindowCenter: () => {},
    WindowSetTitle: (title) => {
      document.title = title;
    },
    WindowFullscreen: () => {},
    WindowUnfullscreen: () => {},
    WindowIsFullscreen: () => false,
    WindowGetSize: () => ({ w: window.innerWidth, h: window.innerHeight }),
    WindowSetSize: () => {},
    WindowSetMaxSize: () => {},
    WindowSetMinSize: () => {},
    WindowSetPosition: () => {},
    WindowGetPosition: () => ({ x: window.screenX, y: window.screenY }),
    WindowHide: () => {},
    WindowShow: () => {},
    WindowMaximise: () => {},
    WindowToggleMaximise: () => {},
    WindowUnmaximise: () => {},
    WindowIsMaximised: () => false,
    WindowMinimise: () => {},
    WindowUnminimise: () => {},
    WindowSetBackgroundColour: () => {},
    ScreenGetAll: () => [],
    WindowIsMinimised: () => false,
    WindowIsNormal: () => true,
    BrowserOpenURL: (url) => window.open(url, '_blank'),
    Environment: () => ({
      buildType: 'development',
      platform: navigator.platform,
      arch: 'unknown',
    }),
    Quit: () => window.close(),
    Hide: () => {},
    Show: () => {},
    ClipboardGetText: () => navigator.clipboard.readText(),
    ClipboardSetText: (text) => navigator.clipboard.writeText(text),
    OnFileDrop: () => () => {},
    OnFileDropOff: () => {},
    CanResolveFilePaths: () => false,
    ResolveFilePaths: () => Promise.resolve([]),
  };
}

/**
 * Emit an event to local listeners (used by WebSocket adapter).
 */
export function emitEvent(eventName: string, ...data: unknown[]): void {
  if (window.runtime) {
    window.runtime.EventsEmit(eventName, ...data);
  }
}

// REST API client reference - will be set during initialization
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let restApiClient: Record<string, (...args: any[]) => Promise<any>> | null = null;

/**
 * Set the REST API client for the Go App polyfill.
 * This should be called after the adapter is initialized.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function setRestApiClient(client: Record<string, (...args: any[]) => Promise<any>>): void {
  restApiClient = client;
  console.log('[WailsPolyfill] REST API client set');
}

/**
 * Install the Go App polyfill (window.go.main.App).
 * This provides a Proxy that redirects method calls to the REST API.
 */
export function installGoAppPolyfill(): void {
  if (typeof window === 'undefined') {
    return;
  }

  if (window.go?.main?.App) {
    console.log('[WailsPolyfill] Wails Go App detected, skipping polyfill');
    return;
  }

  console.log('[WailsPolyfill] Installing Go App polyfill');

  // Create a Proxy that intercepts method calls
  const appProxy = new Proxy(
    {},
    {
      get(_target, prop: string) {
        // Return a function that calls the REST API
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return async (...args: any[]) => {
          if (!restApiClient) {
            console.error(`[WailsPolyfill] REST API client not initialized, cannot call ${prop}`);
            throw new Error(`REST API client not initialized`);
          }

          // Check if the method exists in the REST API client
          if (typeof restApiClient[prop] === 'function') {
            return restApiClient[prop](...args);
          }

          // Method not implemented in REST API
          console.warn(`[WailsPolyfill] Method ${prop} not implemented in REST API`);
          throw new Error(`Method ${prop} not available in REST API mode`);
        };
      },
    }
  );

  // Install the polyfill
  window.go = window.go || {};
  window.go.main = window.go.main || {};
  window.go.main.App = appProxy as WailsGoApp;
}

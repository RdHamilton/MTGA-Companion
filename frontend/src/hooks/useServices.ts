/**
 * useServices Hook
 *
 * Initializes and manages the API services layer. This hook handles:
 * - Service initialization (REST API or Wails bindings)
 * - WebSocket connection management in REST mode
 * - Cleanup on unmount
 *
 * Usage:
 *   function App() {
 *     const { isReady, isRestMode, error } = useServices();
 *
 *     if (!isReady) return <Loading />;
 *     if (error) return <Error message={error} />;
 *
 *     return <MainApp />;
 *   }
 */

import { useEffect, useState, useCallback } from 'react';
import {
  initializeServices,
  cleanupServices,
  isRestApiEnabled,
  setUseRestApi,
} from '../services/adapter';

export interface UseServicesOptions {
  /** Force REST API mode regardless of environment */
  forceRestApi?: boolean;
  /** Custom API base URL */
  apiBaseUrl?: string;
  /** Custom WebSocket URL */
  wsUrl?: string;
}

export interface UseServicesReturn {
  /** Whether services have been initialized */
  isReady: boolean;
  /** Whether REST API mode is active */
  isRestMode: boolean;
  /** Any initialization error */
  error: string | null;
  /** Manually switch to REST API mode */
  enableRestApi: () => void;
  /** Manually switch to Wails mode */
  disableRestApi: () => void;
}

/**
 * Hook for initializing and managing API services.
 */
export function useServices(options?: UseServicesOptions): UseServicesReturn {
  const [isReady, setIsReady] = useState(false);
  const [isRestMode, setIsRestMode] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    async function init() {
      try {
        await initializeServices({
          useRest: options?.forceRestApi,
          apiBaseUrl: options?.apiBaseUrl,
          wsUrl: options?.wsUrl,
        });

        if (mounted) {
          setIsRestMode(isRestApiEnabled());
          setIsReady(true);
        }
      } catch (err) {
        console.error('[useServices] Initialization failed:', err);
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Failed to initialize services');
          // Still mark as ready - we'll fall back to Wails bindings
          setIsReady(true);
        }
      }
    }

    init();

    return () => {
      mounted = false;
      cleanupServices();
    };
  }, [options?.forceRestApi, options?.apiBaseUrl, options?.wsUrl]);

  const enableRestApi = useCallback(() => {
    setUseRestApi(true);
    setIsRestMode(true);
  }, []);

  const disableRestApi = useCallback(() => {
    setUseRestApi(false);
    setIsRestMode(false);
  }, []);

  return {
    isReady,
    isRestMode,
    error,
    enableRestApi,
    disableRestApi,
  };
}

export default useServices;

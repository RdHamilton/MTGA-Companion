import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useServices } from './useServices';

// Mock the adapter module
vi.mock('../services/adapter', () => ({
  initializeServices: vi.fn().mockResolvedValue(undefined),
  cleanupServices: vi.fn(),
  isRestApiEnabled: vi.fn().mockReturnValue(false),
  setUseRestApi: vi.fn(),
}));

import {
  initializeServices,
  cleanupServices,
  isRestApiEnabled,
  setUseRestApi,
} from '../services/adapter';

describe('useServices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize services on mount', async () => {
    const { result } = renderHook(() => useServices());

    // Initially not ready
    expect(result.current.isReady).toBe(false);

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    expect(initializeServices).toHaveBeenCalledWith({
      useRest: undefined,
      apiBaseUrl: undefined,
      wsUrl: undefined,
    });
  });

  it('should pass options to initializeServices', async () => {
    const options = {
      forceRestApi: true,
      apiBaseUrl: 'http://localhost:9000/api',
      wsUrl: 'ws://localhost:9000/ws',
    };

    const { result } = renderHook(() => useServices(options));

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    expect(initializeServices).toHaveBeenCalledWith({
      useRest: true,
      apiBaseUrl: 'http://localhost:9000/api',
      wsUrl: 'ws://localhost:9000/ws',
    });
  });

  it('should cleanup services on unmount', async () => {
    const { result, unmount } = renderHook(() => useServices());

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    unmount();

    expect(cleanupServices).toHaveBeenCalled();
  });

  it('should report REST mode when enabled', async () => {
    vi.mocked(isRestApiEnabled).mockReturnValue(true);

    const { result } = renderHook(() => useServices());

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    expect(result.current.isRestMode).toBe(true);
  });

  it('should report Wails mode when REST is disabled', async () => {
    vi.mocked(isRestApiEnabled).mockReturnValue(false);

    const { result } = renderHook(() => useServices());

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    expect(result.current.isRestMode).toBe(false);
  });

  it('should handle initialization errors gracefully', async () => {
    vi.mocked(initializeServices).mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useServices());

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    expect(result.current.error).toBe('Network error');
  });

  it('should allow manually enabling REST API', async () => {
    const { result } = renderHook(() => useServices());

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    act(() => {
      result.current.enableRestApi();
    });

    expect(setUseRestApi).toHaveBeenCalledWith(true);
    expect(result.current.isRestMode).toBe(true);
  });

  it('should allow manually disabling REST API', async () => {
    vi.mocked(isRestApiEnabled).mockReturnValue(true);

    const { result } = renderHook(() => useServices());

    await waitFor(() => {
      expect(result.current.isReady).toBe(true);
    });

    act(() => {
      result.current.disableRestApi();
    });

    expect(setUseRestApi).toHaveBeenCalledWith(false);
    expect(result.current.isRestMode).toBe(false);
  });
});

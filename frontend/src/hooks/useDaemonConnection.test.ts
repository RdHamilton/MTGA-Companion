import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDaemonConnection } from './useDaemonConnection';

// Mock the API modules
vi.mock('@/services/api', () => ({
  system: {
    getStatus: vi.fn(),
  },
}));

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { system } from '@/services/api';
import { showToast } from '../components/ToastContainer';

const mockGetStatus = vi.mocked(system.getStatus);

describe('useDaemonConnection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Default mock for getStatus
    mockGetStatus.mockResolvedValue({
      status: 'standalone',
      connected: false,
      mode: 'standalone',
      url: 'ws://localhost:9999',
      port: 9999,
    });
  });

  describe('initial state', () => {
    it('returns default connection status', () => {
      const { result } = renderHook(() => useDaemonConnection());

      expect(result.current.connectionStatus.status).toBe('standalone');
      expect(result.current.connectionStatus.connected).toBe(false);
      expect(result.current.connectionStatus.mode).toBe('standalone');
      expect(result.current.connectionStatus.port).toBe(9999);
    });

    it('returns default daemon mode', () => {
      const { result } = renderHook(() => useDaemonConnection());

      expect(result.current.daemonMode).toBe('auto');
    });

    it('returns default daemon port', () => {
      const { result } = renderHook(() => useDaemonConnection());

      expect(result.current.daemonPort).toBe(9999);
    });

    it('returns isReconnecting as false', () => {
      const { result } = renderHook(() => useDaemonConnection());

      expect(result.current.isReconnecting).toBe(false);
    });
  });

  describe('loadConnectionStatus', () => {
    it('loads connection status on mount', async () => {
      mockGetStatus.mockResolvedValueOnce({
        status: 'connected',
        connected: true,
        mode: 'daemon',
        url: 'ws://localhost:8888',
        port: 8888,
      });

      const { result } = renderHook(() => useDaemonConnection());

      await waitFor(() => {
        expect(result.current.connectionStatus.status).toBe('connected');
      });

      expect(result.current.daemonPort).toBe(8888);
      expect(mockGetStatus).toHaveBeenCalled();
    });

    it('handles load error silently', async () => {
      mockGetStatus.mockRejectedValueOnce(new Error('Load failed'));

      const { result } = renderHook(() => useDaemonConnection());

      // Should still have default state after error
      await waitFor(() => {
        expect(mockGetStatus).toHaveBeenCalled();
      });

      expect(result.current.connectionStatus.status).toBe('standalone');
    });
  });

  describe('handleDaemonPortChange', () => {
    it('updates daemon port state', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      // Wait for initial load to complete
      await waitFor(() => {
        expect(mockGetStatus).toHaveBeenCalled();
      });

      await act(async () => {
        await result.current.handleDaemonPortChange(8080);
      });

      expect(result.current.daemonPort).toBe(8080);
    });

    it('rejects ports below 1024', async () => {
      const { result } = renderHook(() => useDaemonConnection());
      const initialPort = result.current.daemonPort;

      await act(async () => {
        await result.current.handleDaemonPortChange(1000);
      });

      expect(result.current.daemonPort).toBe(initialPort);
    });

    it('rejects ports above 65535', async () => {
      const { result } = renderHook(() => useDaemonConnection());
      const initialPort = result.current.daemonPort;

      await act(async () => {
        await result.current.handleDaemonPortChange(70000);
      });

      expect(result.current.daemonPort).toBe(initialPort);
    });

    // Note: SetDaemonPort is a no-op in REST API mode, so error toast test is removed
  });

  describe('handleReconnect', () => {
    it('sets isReconnecting to true during reconnection', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      let reconnectPromise: Promise<void>;
      act(() => {
        reconnectPromise = result.current.handleReconnect();
      });

      expect(result.current.isReconnecting).toBe(true);

      await act(async () => {
        await reconnectPromise;
      });

      expect(result.current.isReconnecting).toBe(false);
    });

    it('refreshes connection status after reconnect', async () => {
      mockGetStatus
        .mockResolvedValueOnce({
          status: 'standalone',
          connected: false,
          mode: 'standalone',
          url: 'ws://localhost:9999',
          port: 9999,
        })
        .mockResolvedValueOnce({
          status: 'connected',
          connected: true,
          mode: 'daemon',
          url: 'ws://localhost:9999',
          port: 9999,
        });

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleReconnect();
      });

      await waitFor(() => {
        expect(result.current.connectionStatus.status).toBe('connected');
      });
    });

    it('shows success toast on successful reconnect', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleReconnect();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'Successfully reconnected to daemon',
        'success'
      );
    });

    // Note: reconnectToDaemon is a no-op in REST API mode, so it won't fail
  });

  describe('handleModeChange', () => {
    it('updates daemon mode state', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('standalone');
      });

      expect(result.current.daemonMode).toBe('standalone');
    });

    it('does not fail for auto mode', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('auto');
      });

      expect(result.current.daemonMode).toBe('auto');
    });

    it('shows success toast for standalone mode switch', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('standalone');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'Switched to standalone mode',
        'success'
      );
    });

    it('shows success toast for daemon mode switch', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('daemon');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'Switched to daemon mode',
        'success'
      );
    });

    // Note: Mode switch functions are no-ops in REST API mode, so they won't fail
  });
});

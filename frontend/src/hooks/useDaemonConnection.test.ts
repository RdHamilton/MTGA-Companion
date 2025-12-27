import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDaemonConnection } from './useDaemonConnection';
import { mockWailsApp } from '@/test/mocks/apiMock';

// Mock showToast
vi.mock('../components/ToastContainer', () => ({
  showToast: {
    show: vi.fn(),
  },
}));

import { showToast } from '../components/ToastContainer';

describe('useDaemonConnection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
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
      mockWailsApp.GetConnectionStatus.mockResolvedValueOnce({
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
      expect(mockWailsApp.GetConnectionStatus).toHaveBeenCalled();
    });

    it('handles load error silently', async () => {
      mockWailsApp.GetConnectionStatus.mockRejectedValueOnce(new Error('Load failed'));

      const { result } = renderHook(() => useDaemonConnection());

      // Should still have default state after error
      await waitFor(() => {
        expect(mockWailsApp.GetConnectionStatus).toHaveBeenCalled();
      });

      expect(result.current.connectionStatus.status).toBe('standalone');
    });
  });

  describe('handleDaemonPortChange', () => {
    it('updates daemon port state', async () => {
      // Mock GetConnectionStatus to return initial state so port isn't overwritten
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
        status: 'standalone',
        connected: false,
        mode: 'standalone',
        url: 'ws://localhost:9999',
        port: 9999,
      });

      const { result } = renderHook(() => useDaemonConnection());

      // Wait for initial load to complete
      await waitFor(() => {
        expect(mockWailsApp.GetConnectionStatus).toHaveBeenCalled();
      });

      await act(async () => {
        await result.current.handleDaemonPortChange(8080);
      });

      expect(result.current.daemonPort).toBe(8080);
    });

    it('calls SetDaemonPort API', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleDaemonPortChange(8080);
      });

      expect(mockWailsApp.SetDaemonPort).toHaveBeenCalledWith(8080);
    });

    it('rejects ports below 1024', async () => {
      const { result } = renderHook(() => useDaemonConnection());
      const initialPort = result.current.daemonPort;

      await act(async () => {
        await result.current.handleDaemonPortChange(1000);
      });

      expect(result.current.daemonPort).toBe(initialPort);
      expect(mockWailsApp.SetDaemonPort).not.toHaveBeenCalled();
    });

    it('rejects ports above 65535', async () => {
      const { result } = renderHook(() => useDaemonConnection());
      const initialPort = result.current.daemonPort;

      await act(async () => {
        await result.current.handleDaemonPortChange(70000);
      });

      expect(result.current.daemonPort).toBe(initialPort);
      expect(mockWailsApp.SetDaemonPort).not.toHaveBeenCalled();
    });

    it('shows error toast on API failure', async () => {
      mockWailsApp.SetDaemonPort.mockRejectedValueOnce(new Error('Port error'));

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleDaemonPortChange(8080);
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to set daemon port'),
        'error'
      );
    });
  });

  describe('handleReconnect', () => {
    it('sets isReconnecting to true during reconnection', async () => {
      let resolveReconnect: () => void;
      mockWailsApp.ReconnectToDaemon.mockImplementationOnce(
        () => new Promise<void>((resolve) => {
          resolveReconnect = resolve;
        })
      );
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
        status: 'connected',
        connected: true,
        mode: 'daemon',
        url: 'ws://localhost:9999',
        port: 9999,
      });

      const { result } = renderHook(() => useDaemonConnection());

      let reconnectPromise: Promise<void>;
      act(() => {
        reconnectPromise = result.current.handleReconnect();
      });

      expect(result.current.isReconnecting).toBe(true);

      await act(async () => {
        resolveReconnect!();
        await reconnectPromise;
      });

      expect(result.current.isReconnecting).toBe(false);
    });

    it('calls ReconnectToDaemon API', async () => {
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
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

      expect(mockWailsApp.ReconnectToDaemon).toHaveBeenCalled();
    });

    it('refreshes connection status after reconnect', async () => {
      mockWailsApp.GetConnectionStatus
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
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
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

      expect(showToast.show).toHaveBeenCalledWith(
        'Successfully reconnected to daemon',
        'success'
      );
    });

    it('shows error toast on reconnect failure', async () => {
      mockWailsApp.ReconnectToDaemon.mockRejectedValueOnce(new Error('Connection refused'));

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleReconnect();
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to reconnect'),
        'error'
      );
    });
  });

  describe('handleModeChange', () => {
    it('updates daemon mode state', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('standalone');
      });

      expect(result.current.daemonMode).toBe('standalone');
    });

    it('calls SwitchToStandaloneMode for standalone mode', async () => {
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
        status: 'standalone',
        connected: false,
        mode: 'standalone',
        url: 'ws://localhost:9999',
        port: 9999,
      });

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('standalone');
      });

      expect(mockWailsApp.SwitchToStandaloneMode).toHaveBeenCalled();
    });

    it('calls SwitchToDaemonMode for daemon mode', async () => {
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
        status: 'connected',
        connected: true,
        mode: 'daemon',
        url: 'ws://localhost:9999',
        port: 9999,
      });

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('daemon');
      });

      expect(mockWailsApp.SwitchToDaemonMode).toHaveBeenCalled();
    });

    it('does not call API for auto mode', async () => {
      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('auto');
      });

      expect(mockWailsApp.SwitchToStandaloneMode).not.toHaveBeenCalled();
      expect(mockWailsApp.SwitchToDaemonMode).not.toHaveBeenCalled();
    });

    it('shows success toast for standalone mode switch', async () => {
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
        status: 'standalone',
        connected: false,
        mode: 'standalone',
        url: 'ws://localhost:9999',
        port: 9999,
      });

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
      mockWailsApp.GetConnectionStatus.mockResolvedValue({
        status: 'connected',
        connected: true,
        mode: 'daemon',
        url: 'ws://localhost:9999',
        port: 9999,
      });

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('daemon');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        'Switched to daemon mode',
        'success'
      );
    });

    it('shows error toast on mode switch failure', async () => {
      mockWailsApp.SwitchToStandaloneMode.mockRejectedValueOnce(new Error('Switch failed'));

      const { result } = renderHook(() => useDaemonConnection());

      await act(async () => {
        await result.current.handleModeChange('standalone');
      });

      expect(showToast.show).toHaveBeenCalledWith(
        expect.stringContaining('Failed to switch mode'),
        'error'
      );
    });
  });
});

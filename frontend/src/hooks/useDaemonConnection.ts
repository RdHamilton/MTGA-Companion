import { useState, useEffect, useCallback } from 'react';
import {
  GetConnectionStatus,
  SetDaemonPort,
  ReconnectToDaemon,
  SwitchToStandaloneMode,
  SwitchToDaemonMode,
} from '@/services/api/legacy';
import { showToast } from '../components/ToastContainer';
import { gui } from '@/types/models';

export interface UseDaemonConnectionReturn {
  /** Current connection status */
  connectionStatus: gui.ConnectionStatus;
  /** Current daemon mode setting */
  daemonMode: string;
  /** Current daemon port */
  daemonPort: number;
  /** Whether a reconnection is in progress */
  isReconnecting: boolean;
  /** Change the daemon port */
  handleDaemonPortChange: (port: number) => Promise<void>;
  /** Manually reconnect to daemon */
  handleReconnect: () => Promise<void>;
  /** Change connection mode */
  handleModeChange: (mode: string) => Promise<void>;
}

const defaultConnectionStatus = new gui.ConnectionStatus({
  status: 'standalone',
  connected: false,
  mode: 'standalone',
  url: 'ws://localhost:9999',
  port: 9999,
});

export function useDaemonConnection(): UseDaemonConnectionReturn {
  const [connectionStatus, setConnectionStatus] = useState<gui.ConnectionStatus>(defaultConnectionStatus);
  const [daemonMode, setDaemonMode] = useState('auto');
  const [daemonPort, setDaemonPortState] = useState(9999);
  const [isReconnecting, setIsReconnecting] = useState(false);

  const loadConnectionStatus = useCallback(async () => {
    try {
      const status = await GetConnectionStatus();
      setConnectionStatus(gui.ConnectionStatus.createFrom(status));
      setDaemonPortState(status.port || 9999);
    } catch {
      // Connection status load failed silently - UI will show default state
    }
  }, []);

  useEffect(() => {
    loadConnectionStatus();
  }, [loadConnectionStatus]);

  const handleDaemonPortChange = useCallback(async (port: number) => {
    if (port < 1024 || port > 65535) {
      return;
    }

    setDaemonPortState(port);

    try {
      await SetDaemonPort(port);
    } catch (error) {
      showToast.show(`Failed to set daemon port: ${error}`, 'error');
    }
  }, []);

  const handleReconnect = useCallback(async () => {
    setIsReconnecting(true);
    try {
      await ReconnectToDaemon();
      await loadConnectionStatus();
      showToast.show('Successfully reconnected to daemon', 'success');
    } catch (error) {
      showToast.show(`Failed to reconnect to daemon: ${error}`, 'error');
    } finally {
      setIsReconnecting(false);
    }
  }, [loadConnectionStatus]);

  const handleModeChange = useCallback(async (mode: string) => {
    setDaemonMode(mode);

    try {
      if (mode === 'standalone') {
        await SwitchToStandaloneMode();
        await loadConnectionStatus();
        showToast.show('Switched to standalone mode', 'success');
      } else if (mode === 'daemon') {
        await SwitchToDaemonMode();
        await loadConnectionStatus();
        showToast.show('Switched to daemon mode', 'success');
      }
      // 'auto' mode is handled automatically by the app
    } catch (error) {
      showToast.show(`Failed to switch mode: ${error}`, 'error');
    }
  }, [loadConnectionStatus]);

  return {
    connectionStatus,
    daemonMode,
    daemonPort,
    isReconnecting,
    handleDaemonPortChange,
    handleReconnect,
    handleModeChange,
  };
}

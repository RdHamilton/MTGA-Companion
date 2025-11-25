import LoadingButton from '../../LoadingButton';
import { SettingItem, SettingSelect } from '../';
import { gui } from '../../../../wailsjs/go/models';

export interface DaemonConnectionSectionProps {
  connectionStatus: gui.ConnectionStatus;
  daemonMode: string;
  daemonPort: number;
  isReconnecting: boolean;
  onDaemonPortChange: (port: number) => void;
  onReconnect: () => void;
  onModeChange: (mode: string) => void;
}

export function DaemonConnectionSection({
  connectionStatus,
  daemonMode,
  daemonPort,
  isReconnecting,
  onDaemonPortChange,
  onReconnect,
  onModeChange,
}: DaemonConnectionSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">Daemon Connection</h2>

      <div className="setting-item">
        <label className="setting-label">
          Connection Status
          <span className="setting-description">Current connection state to the daemon service</span>
        </label>
        <div className="setting-control">
          <div className={`connection-badge status-${connectionStatus.status}`}>
            <span className="status-dot"></span>
            {connectionStatus.status === 'connected' && 'Connected to Daemon'}
            {connectionStatus.status === 'standalone' && 'Standalone Mode'}
            {connectionStatus.status === 'reconnecting' && 'Reconnecting...'}
          </div>
        </div>
      </div>

      <SettingSelect
        label="Connection Mode"
        description="Choose how the app connects to the daemon"
        value={daemonMode}
        onChange={onModeChange}
        options={[
          { value: 'auto', label: 'Auto (try daemon, fallback to standalone)' },
          { value: 'daemon', label: 'Daemon Only' },
          { value: 'standalone', label: 'Standalone Only (embedded poller)' },
        ]}
      />

      <SettingItem
        label="Daemon Port"
        description="WebSocket port for daemon connection (1024-65535)"
        hint={`ws://localhost:${daemonPort}`}
      >
        <input
          type="number"
          value={daemonPort}
          onChange={(e) => onDaemonPortChange(parseInt(e.target.value))}
          min="1024"
          max="65535"
          className="number-input"
          disabled={daemonMode === 'standalone'}
        />
      </SettingItem>

      <SettingItem
        label="Reconnect"
        description="Manually reconnect to the daemon service"
      >
        <LoadingButton
          loading={isReconnecting}
          loadingText="Reconnecting..."
          onClick={onReconnect}
          disabled={daemonMode === 'standalone'}
        >
          Reconnect to Daemon
        </LoadingButton>
      </SettingItem>
    </div>
  );
}

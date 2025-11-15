import { useState, useEffect } from 'react';
import AboutDialog from '../components/AboutDialog';
import { GetConnectionStatus, SetDaemonPort, ReconnectToDaemon, SwitchToStandaloneMode, SwitchToDaemonMode } from '../../wailsjs/go/main/App';
import './Settings.css';

const Settings = () => {
  const [dbPath, setDbPath] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showNotifications, setShowNotifications] = useState(true);
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

  // Daemon settings
  const [connectionStatus, setConnectionStatus] = useState<any>({
    status: 'standalone',
    connected: false,
    mode: 'standalone',
    url: 'ws://localhost:9999',
    port: 9999
  });
  const [daemonMode, setDaemonMode] = useState('auto');
  const [daemonPort, setDaemonPortState] = useState(9999);
  const [isReconnecting, setIsReconnecting] = useState(false);

  // Load connection status on mount
  useEffect(() => {
    loadConnectionStatus();
  }, []);

  const loadConnectionStatus = async () => {
    try {
      const status = await GetConnectionStatus();
      setConnectionStatus(status);
      setDaemonPortState(status.port || 9999);
    } catch (error) {
      console.error('Failed to load connection status:', error);
    }
  };

  const handleDaemonPortChange = async (port: number) => {
    if (port < 1024 || port > 65535) {
      return;
    }

    setDaemonPortState(port);

    try {
      await SetDaemonPort(port);
      console.log('Daemon port updated to', port);
    } catch (error) {
      console.error('Failed to set daemon port:', error);
      alert(`Failed to set daemon port: ${error}`);
    }
  };

  const handleReconnect = async () => {
    setIsReconnecting(true);
    try {
      await ReconnectToDaemon();
      await loadConnectionStatus();
      alert('Successfully reconnected to daemon');
    } catch (error) {
      console.error('Failed to reconnect:', error);
      alert(`Failed to reconnect to daemon: ${error}`);
    } finally {
      setIsReconnecting(false);
    }
  };

  const handleModeChange = async (mode: string) => {
    setDaemonMode(mode);

    try {
      if (mode === 'standalone') {
        await SwitchToStandaloneMode();
        await loadConnectionStatus();
        alert('Switched to standalone mode');
      } else if (mode === 'daemon') {
        await SwitchToDaemonMode();
        await loadConnectionStatus();
        alert('Switched to daemon mode');
      }
      // 'auto' mode is handled automatically by the app
    } catch (error) {
      console.error('Failed to switch mode:', error);
      alert(`Failed to switch mode: ${error}`);
    }
  };

  const handleSave = () => {
    // TODO: Implement backend settings save
    // For now, just show success message
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const handleReset = () => {
    setDbPath('');
    setAutoRefresh(false);
    setRefreshInterval(30);
    setShowNotifications(true);
  };

  const handleExportData = () => {
    // TODO: Implement data export
    alert('Export functionality coming soon!');
  };

  const handleImportData = () => {
    // TODO: Implement data import
    alert('Import functionality coming soon!');
  };

  return (
    <div className="page-container">
      <div className="settings-header">
        <h1 className="page-title">Settings</h1>
        {saved && <div className="save-notification">Settings saved successfully!</div>}
      </div>

      <div className="settings-content">
        {/* Database Configuration */}
        <div className="settings-section">
          <h2 className="section-title">Database Configuration</h2>
          <div className="setting-item">
            <label className="setting-label">
              Database Path
              <span className="setting-description">Location of the MTGA Companion database file</span>
            </label>
            <div className="setting-control">
              <input
                type="text"
                value={dbPath}
                onChange={(e) => setDbPath(e.target.value)}
                placeholder="/Users/username/.mtga-companion/mtga.db"
                className="text-input"
              />
              <button className="browse-button">Browse...</button>
            </div>
          </div>
        </div>

        {/* Daemon Connection */}
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

          <div className="setting-item">
            <label className="setting-label">
              Connection Mode
              <span className="setting-description">Choose how the app connects to the daemon</span>
            </label>
            <div className="setting-control">
              <select
                className="select-input"
                value={daemonMode}
                onChange={(e) => handleModeChange(e.target.value)}
              >
                <option value="auto">Auto (try daemon, fallback to standalone)</option>
                <option value="daemon">Daemon Only</option>
                <option value="standalone">Standalone Only (embedded poller)</option>
              </select>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Daemon Port
              <span className="setting-description">WebSocket port for daemon connection (1024-65535)</span>
            </label>
            <div className="setting-control">
              <input
                type="number"
                value={daemonPort}
                onChange={(e) => handleDaemonPortChange(parseInt(e.target.value))}
                min="1024"
                max="65535"
                className="number-input"
                disabled={daemonMode === 'standalone'}
              />
              <span className="setting-hint">ws://localhost:{daemonPort}</span>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Reconnect
              <span className="setting-description">Manually reconnect to the daemon service</span>
            </label>
            <div className="setting-control">
              <button
                className="action-button"
                onClick={handleReconnect}
                disabled={isReconnecting || daemonMode === 'standalone'}
              >
                {isReconnecting ? 'Reconnecting...' : 'Reconnect to Daemon'}
              </button>
            </div>
          </div>
        </div>

        {/* Application Preferences */}
        <div className="settings-section">
          <h2 className="section-title">Application Preferences</h2>

          <div className="setting-item">
            <label className="setting-label">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="checkbox-input"
              />
              Auto-refresh data
              <span className="setting-description">Automatically refresh statistics at regular intervals</span>
            </label>
          </div>

          {autoRefresh && (
            <div className="setting-item indented">
              <label className="setting-label">
                Refresh Interval (seconds)
                <span className="setting-description">How often to refresh data automatically</span>
              </label>
              <div className="setting-control">
                <input
                  type="number"
                  value={refreshInterval}
                  onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                  min="10"
                  max="300"
                  step="10"
                  className="number-input"
                />
              </div>
            </div>
          )}

          <div className="setting-item">
            <label className="setting-label">
              <input
                type="checkbox"
                checked={showNotifications}
                onChange={(e) => setShowNotifications(e.target.checked)}
                className="checkbox-input"
              />
              Show notifications
              <span className="setting-description">Display notifications for match results and updates</span>
            </label>
          </div>
        </div>

        {/* Data Management */}
        <div className="settings-section">
          <h2 className="section-title">Data Management</h2>

          <div className="setting-item">
            <label className="setting-label">
              Export Data
              <span className="setting-description">Export your match history and statistics to a file</span>
            </label>
            <div className="setting-control">
              <button className="action-button" onClick={handleExportData}>
                Export to JSON
              </button>
              <button className="action-button" onClick={handleExportData}>
                Export to CSV
              </button>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Import Data
              <span className="setting-description">Import match history from a file</span>
            </label>
            <div className="setting-control">
              <button className="action-button" onClick={handleImportData}>
                Import from File
              </button>
            </div>
          </div>

          <div className="setting-item danger">
            <label className="setting-label">
              Clear All Data
              <span className="setting-description">Permanently delete all match history and statistics</span>
            </label>
            <div className="setting-control">
              <button className="danger-button" onClick={() => {
                if (window.confirm('Are you sure you want to delete all data? This cannot be undone.')) {
                  alert('Clear data functionality coming soon!');
                }
              }}>
                Clear All Data
              </button>
            </div>
          </div>
        </div>

        {/* Theme Preferences */}
        <div className="settings-section">
          <h2 className="section-title">Appearance</h2>

          <div className="setting-item">
            <label className="setting-label">
              Theme
              <span className="setting-description">Choose your preferred color scheme</span>
            </label>
            <div className="setting-control">
              <select className="select-input">
                <option value="dark">Dark (Default)</option>
                <option value="light" disabled>Light (Coming Soon)</option>
                <option value="auto" disabled>Auto (System Default)</option>
              </select>
            </div>
          </div>
        </div>

        {/* About */}
        <div className="settings-section">
          <h2 className="section-title">About</h2>

          <div className="about-content">
            <div className="about-item">
              <span className="about-label">Version:</span>
              <span className="about-value">1.0.0</span>
            </div>
            <div className="about-item">
              <span className="about-label">Build:</span>
              <span className="about-value">Development</span>
            </div>
            <div className="about-item">
              <span className="about-label">Platform:</span>
              <span className="about-value">Wails + React</span>
            </div>
            <div className="setting-control" style={{ marginTop: '16px' }}>
              <button className="action-button" onClick={() => setShowAbout(true)}>
                About MTGA Companion
              </button>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="settings-actions">
          <button className="primary-button" onClick={handleSave}>
            Save Settings
          </button>
          <button className="secondary-button" onClick={handleReset}>
            Reset to Defaults
          </button>
        </div>
      </div>

      {/* About Dialog */}
      <AboutDialog isOpen={showAbout} onClose={() => setShowAbout(false)} />
    </div>
  );
};

export default Settings;

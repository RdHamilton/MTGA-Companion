import { useState, useEffect } from 'react';
import AboutDialog from '../components/AboutDialog';
import { GetConnectionStatus, SetDaemonPort, ReconnectToDaemon, SwitchToStandaloneMode, SwitchToDaemonMode, ExportToJSON, ExportToCSV, ImportFromFile, ClearAllData, TriggerReplayLogs } from '../../wailsjs/go/main/App';
import { EventsOn, WindowReloadApp } from '../../wailsjs/runtime/runtime';
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

  // Replay logs settings
  const [clearDataBeforeReplay, setClearDataBeforeReplay] = useState(false);
  const [isReplaying, setIsReplaying] = useState(false);
  const [replayProgress, setReplayProgress] = useState<any>(null);

  // Load connection status on mount
  useEffect(() => {
    loadConnectionStatus();
  }, []);

  // Listen for replay events
  useEffect(() => {
    const unsubscribeStarted = EventsOn('replay:started', (data: any) => {
      console.log('Replay started:', data);
      setIsReplaying(true);
      setReplayProgress(null);
    });

    const unsubscribeProgress = EventsOn('replay:progress', (data: any) => {
      console.log('Replay progress:', data);
      setReplayProgress(data);
    });

    const unsubscribeCompleted = EventsOn('replay:completed', (data: any) => {
      console.log('Replay completed:', data);
      setIsReplaying(false);
      setReplayProgress(data);
      // Keep progress visible for a moment, then reload using Wails native method
      setTimeout(() => {
        WindowReloadApp(); // Refresh to show updated data
      }, 2000);
    });

    const unsubscribeError = EventsOn('replay:error', (data: any) => {
      console.error('Replay error:', data);
      setIsReplaying(false);
      // Error will be logged to console
    });

    return () => {
      unsubscribeStarted();
      unsubscribeProgress();
      unsubscribeCompleted();
      unsubscribeError();
    };
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

  const handleExportData = async (format: 'json' | 'csv') => {
    try {
      if (format === 'json') {
        await ExportToJSON();
      } else {
        await ExportToCSV();
      }
      alert(`Successfully exported data to ${format.toUpperCase()}!`);
    } catch (error) {
      console.error('Export failed:', error);
      alert(`Failed to export data: ${error}`);
    }
  };

  const handleImportData = async () => {
    try {
      await ImportFromFile();
      alert('Successfully imported data! Refresh the page to see updated statistics.');
    } catch (error) {
      console.error('Import failed:', error);
      alert(`Failed to import data: ${error}`);
    }
  };

  const handleReplayLogs = async () => {
    console.log('=== REPLAY LOGS CLICKED ===');
    console.log('handleReplayLogs called');
    console.log('Connection status:', connectionStatus);
    console.log('Clear data before replay:', clearDataBeforeReplay);

    // Check if connected to daemon
    if (connectionStatus.status !== 'connected') {
      console.error('Daemon not connected, status:', connectionStatus.status);
      return;
    }

    console.log('Calling TriggerReplayLogs...');
    try {
      await TriggerReplayLogs(clearDataBeforeReplay);
      console.log('TriggerReplayLogs succeeded - replay started on daemon');
      // Progress UI will update automatically from events
    } catch (error) {
      console.error('Failed to trigger replay:', error);
    }
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
              <button className="action-button" onClick={() => handleExportData('json')}>
                Export to JSON
              </button>
              <button className="action-button" onClick={() => handleExportData('csv')}>
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

          <div className="setting-item">
            <label className="setting-label">
              Replay Historical Logs
              <span className="setting-description">
                Process all historical MTGA log files through the daemon. This replays logs chronologically
                to ensure all game data, statistics, and quest progression are tracked correctly.
              </span>
            </label>
            <div className="setting-control">
              <div style={{ marginBottom: '8px' }}>
                <label style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <input
                    type="checkbox"
                    checked={clearDataBeforeReplay}
                    onChange={(e) => setClearDataBeforeReplay(e.target.checked)}
                    className="checkbox-input"
                    disabled={isReplaying}
                  />
                  <span>Clear all data before replay (recommended for first-time setup)</span>
                </label>
              </div>
              <button
                className="action-button primary"
                onClick={handleReplayLogs}
                disabled={isReplaying || connectionStatus.status !== 'connected'}
              >
                {isReplaying ? 'Replaying Logs...' : 'Replay Historical Logs'}
              </button>
              {connectionStatus.status !== 'connected' && (
                <div className="setting-hint" style={{ color: '#ff6b6b', marginTop: '8px' }}>
                  Daemon must be running to replay logs
                </div>
              )}
            </div>
          </div>

          {(isReplaying || replayProgress) && (
            <div className="setting-item">
              <div className="replay-progress" style={{
                background: '#2d2d2d',
                padding: '16px',
                borderRadius: '8px',
                marginTop: '8px'
              }}>
                <h3 style={{ marginTop: 0, color: isReplaying ? '#4a9eff' : '#00ff00' }}>
                  {isReplaying ? 'Replaying Historical Logs...' : '✓ Replay Complete'}
                </h3>
                {replayProgress && (
                  <>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '8px', marginBottom: '12px' }}>
                      <div>Files: {replayProgress.processedFiles || 0} / {replayProgress.totalFiles || 0}</div>
                      <div>Entries: {replayProgress.totalEntries || 0}</div>
                      <div>Matches: {replayProgress.matchesImported || 0}</div>
                      <div>Decks: {replayProgress.decksImported || 0}</div>
                      <div>Quests: {replayProgress.questsImported || 0}</div>
                      {replayProgress.duration && (
                        <div>Duration: {replayProgress.duration.toFixed(1)}s</div>
                      )}
                    </div>
                    {replayProgress.currentFile && isReplaying && (
                      <div style={{ fontSize: '0.9em', color: '#aaa' }}>
                        Current: {replayProgress.currentFile}
                      </div>
                    )}
                    {isReplaying && (
                      <div style={{
                        width: '100%',
                        height: '8px',
                        background: '#1e1e1e',
                        borderRadius: '4px',
                        overflow: 'hidden',
                        marginTop: '12px'
                      }}>
                        <div style={{
                          width: `${((replayProgress.processedFiles || 0) / (replayProgress.totalFiles || 1)) * 100}%`,
                          height: '100%',
                          background: '#4a9eff',
                          transition: 'width 0.3s ease'
                        }}></div>
                      </div>
                    )}
                    {!isReplaying && (
                      <div style={{ color: '#aaa', marginTop: '8px' }}>
                        Page will refresh in 2 seconds to show imported data...
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          )}

          <div className="setting-item danger">
            <label className="setting-label">
              Clear All Data
              <span className="setting-description">Permanently delete all match history and statistics</span>
            </label>
            <div className="setting-control">
              <button className="danger-button" onClick={async () => {
                try {
                  await ClearAllData();
                  alert('All data has been cleared successfully!');
                  window.location.reload(); // Refresh to show empty state
                } catch (error) {
                  console.error('Clear data failed:', error);
                  alert(`Failed to clear data: ${error}`);
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

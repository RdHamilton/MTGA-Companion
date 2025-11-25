import { useState, useEffect } from 'react';
import AboutDialog from '../components/AboutDialog';
import { GetConnectionStatus, SetDaemonPort, ReconnectToDaemon, SwitchToStandaloneMode, SwitchToDaemonMode, ExportToJSON, ExportToCSV, ImportFromFile, ImportLogFile, ClearAllData, TriggerReplayLogs, StartReplayWithFileDialog, PauseReplay, ResumeReplay, StopReplay, FetchSetRatings, RefreshSetRatings, FetchSetCards, RefreshSetCards, RecalculateAllDraftGrades, ClearDatasetCache, GetDatasetSource } from '../../wailsjs/go/main/App';
import { EventsOn, WindowReloadApp } from '../../wailsjs/runtime/runtime';
import { subscribeToReplayState, getReplayState } from '../App';
import { showToast } from '../components/ToastContainer';
import { gui } from '../../wailsjs/go/models';
import './Settings.css';

const Settings = () => {
  const [dbPath, setDbPath] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showNotifications, setShowNotifications] = useState(true);
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

  // Daemon settings
  const [connectionStatus, setConnectionStatus] = useState<gui.ConnectionStatus>(
    new gui.ConnectionStatus({
      status: 'standalone',
      connected: false,
      mode: 'standalone',
      url: 'ws://localhost:9999',
      port: 9999,
    })
  );
  const [daemonMode, setDaemonMode] = useState('auto');
  const [daemonPort, setDaemonPortState] = useState(9999);
  const [isReconnecting, setIsReconnecting] = useState(false);

  // Replay logs settings
  const [clearDataBeforeReplay, setClearDataBeforeReplay] = useState(false);
  const [isReplaying, setIsReplaying] = useState(false);
  const [replayProgress, setReplayProgress] = useState<gui.LogReplayProgress | null>(null);

  // Replay tool settings - use global state for active/paused to persist across navigation
  const [replayToolActive, setReplayToolActive] = useState(getReplayState().isActive);
  const [replayToolPaused, setReplayToolPaused] = useState(getReplayState().isPaused);
  const [replayToolProgress, setReplayToolProgress] = useState<gui.ReplayStatus | null>(getReplayState().progress);
  const [replaySpeed, setReplaySpeed] = useState(1.0);
  const [replayFilter, setReplayFilter] = useState('all');
  const [pauseOnDraft, setPauseOnDraft] = useState(false);

  // 17Lands data settings
  const [setCode, setSetCode] = useState('');
  const [draftFormat, setDraftFormat] = useState('PremierDraft');
  const [isFetchingRatings, setIsFetchingRatings] = useState(false);
  const [isFetchingCards, setIsFetchingCards] = useState(false);
  const [isRecalculating, setIsRecalculating] = useState(false);
  const [recalculateMessage, setRecalculateMessage] = useState('');
  const [dataSource, setDataSource] = useState<string>('');
  const [isClearingCache, setIsClearingCache] = useState(false);

  // Load connection status on mount
  useEffect(() => {
    loadConnectionStatus();
  }, []);

  // Subscribe to global replay state changes
  useEffect(() => {
    // Get initial state immediately
    const initialState = getReplayState();
    setReplayToolActive(initialState.isActive);
    setReplayToolPaused(initialState.isPaused);
    setReplayToolProgress(initialState.progress);

    // Subscribe to future changes
    const unsubscribe = subscribeToReplayState((state) => {
      setReplayToolActive(state.isActive);
      setReplayToolPaused(state.isPaused);
      setReplayToolProgress(state.progress);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  // Listen for replay events
  useEffect(() => {
    const unsubscribeStarted = EventsOn('replay:started', () => {
      setIsReplaying(true);
      setReplayProgress(null);
    });

    const unsubscribeProgress = EventsOn('replay:progress', (data: unknown) => {
      setReplayProgress(gui.LogReplayProgress.createFrom(data));
    });

    const unsubscribeCompleted = EventsOn('replay:completed', (data: unknown) => {
      setIsReplaying(false);
      setReplayProgress(gui.LogReplayProgress.createFrom(data));
      // Keep progress visible for a moment, then reload using Wails native method
      setTimeout(() => {
        WindowReloadApp(); // Refresh to show updated data
      }, 2000);
    });

    const unsubscribeError = EventsOn('replay:error', () => {
      setIsReplaying(false);
    });

    // Note: Replay tool events are now handled globally in App.tsx
    // This ensures state persists across navigation and enables automatic tab switching

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
      setConnectionStatus(gui.ConnectionStatus.createFrom(status));
      setDaemonPortState(status.port || 9999);
    } catch {
      // Connection status load failed silently - UI will show default state
    }
  };

  const handleDaemonPortChange = async (port: number) => {
    if (port < 1024 || port > 65535) {
      return;
    }

    setDaemonPortState(port);

    try {
      await SetDaemonPort(port);
    } catch (error) {
      showToast.show(`Failed to set daemon port: ${error}`, 'error');
    }
  };

  const handleReconnect = async () => {
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
  };

  const handleModeChange = async (mode: string) => {
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
      showToast.show(`Successfully exported data to ${format.toUpperCase()}!`, 'success');
    } catch (error) {
      showToast.show(`Failed to export data: ${error}`, 'error');
    }
  };

  const handleImportData = async () => {
    try {
      await ImportFromFile();
      showToast.show('Successfully imported data! Refresh the page to see updated statistics.', 'success');
    } catch (error) {
      showToast.show(`Failed to import data: ${error}`, 'error');
    }
  };

  const handleImportLogFile = async () => {
    try {
      const result = await ImportLogFile();

      // User cancelled
      if (!result) {
        return;
      }

      // Show success message with detailed results
      showToast.show(
        `Successfully imported ${result.fileName}! ` +
        `Entries: ${result.entriesRead}, ` +
        `Matches: ${result.matchesStored}, ` +
        `Games: ${result.gamesStored}, ` +
        `Decks: ${result.decksStored}, ` +
        `Ranks: ${result.ranksStored}, ` +
        `Quests: ${result.questsStored}, ` +
        `Drafts: ${result.draftsStored}. ` +
        `Refresh to see updated statistics.`,
        'success'
      );
    } catch (error) {
      showToast.show(`Failed to import log file: ${error}`, 'error');
    }
  };

  const handleReplayLogs = async () => {
    // Check if connected to daemon
    if (connectionStatus.status !== 'connected') {
      return;
    }

    try {
      await TriggerReplayLogs(clearDataBeforeReplay);
      // Progress UI will update automatically from events
    } catch (error) {
      showToast.show(`Failed to trigger replay: ${error}`, 'error');
    }
  };

  // Replay tool handlers
  const handleStartReplayTool = async () => {
    // Check if connected to daemon
    if (connectionStatus.status !== 'connected') {
      showToast.show('Replay tool requires daemon mode. Please start the daemon service.', 'warning');
      return;
    }

    try {
      await StartReplayWithFileDialog(replaySpeed, replayFilter, pauseOnDraft);
    } catch (error) {
      showToast.show(`Failed to start replay: ${error}`, 'error');
    }
  };

  const handlePauseReplayTool = async () => {
    try {
      await PauseReplay();
    } catch (error) {
      showToast.show(`Failed to pause replay: ${error}`, 'error');
    }
  };

  const handleResumeReplayTool = async () => {
    try {
      await ResumeReplay();
    } catch (error) {
      showToast.show(`Failed to resume replay: ${error}`, 'error');
    }
  };

  const handleStopReplayTool = async () => {
    try {
      await StopReplay();
    } catch (error) {
      showToast.show(`Failed to stop replay: ${error}`, 'error');
    }
  };

  // 17Lands handlers
  const handleFetchSetRatings = async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingRatings(true);
    try {
      await FetchSetRatings(setCode.trim().toUpperCase(), draftFormat);

      // Check data source after fetching
      try {
        const source = await GetDatasetSource(setCode.trim().toUpperCase(), draftFormat);
        setDataSource(source);

        const sourceLabel = source === 's3' ? 'S3 public datasets' :
                          source === 'web_api' ? 'web API' :
                          source === 'legacy_api' ? 'legacy API' : source;

        showToast.show(`Successfully fetched 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat}) from ${sourceLabel}! The data is now cached and ready for use in drafts.`, 'success');
      } catch {
        showToast.show(`Successfully fetched 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat})! The data is now cached and ready for use in drafts.`, 'success');
      }
    } catch (error) {
      showToast.show(`Failed to fetch 17Lands ratings: ${error}. Make sure: Set code is correct (e.g., TLA, BLB, DSK, FDN), you have internet connection, and 17Lands has data for this set.`, 'error');
    } finally {
      setIsFetchingRatings(false);
    }
  };

  const handleRefreshSetRatings = async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingRatings(true);
    try {
      await RefreshSetRatings(setCode.trim().toUpperCase(), draftFormat);

      // Check data source after refreshing
      try {
        const source = await GetDatasetSource(setCode.trim().toUpperCase(), draftFormat);
        setDataSource(source);

        const sourceLabel = source === 's3' ? 'S3 public datasets' :
                          source === 'web_api' ? 'web API' :
                          source === 'legacy_api' ? 'legacy API' : source;

        showToast.show(`Successfully refreshed 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat}) from ${sourceLabel}!`, 'success');
      } catch {
        showToast.show(`Successfully refreshed 17Lands ratings for ${setCode.toUpperCase()} (${draftFormat})!`, 'success');
      }
    } catch (error) {
      showToast.show(`Failed to refresh 17Lands ratings: ${error}`, 'error');
    } finally {
      setIsFetchingRatings(false);
    }
  };

  const handleFetchSetCards = async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingCards(true);
    try {
      const count = await FetchSetCards(setCode.trim().toUpperCase());
      showToast.show(`Successfully fetched ${count} cards for ${setCode.toUpperCase()} from Scryfall! Card data is now cached.`, 'success');
    } catch (error) {
      showToast.show(`Failed to fetch cards: ${error}. Make sure the set code is correct and you have internet connection.`, 'error');
    } finally {
      setIsFetchingCards(false);
    }
  };

  const handleRefreshSetCards = async () => {
    if (!setCode || setCode.trim() === '') {
      showToast.show('Please enter a set code (e.g., TLA, BLB, DSK, FDN)', 'warning');
      return;
    }

    setIsFetchingCards(true);
    try {
      const count = await RefreshSetCards(setCode.trim().toUpperCase());
      showToast.show(`Successfully refreshed ${count} cards for ${setCode.toUpperCase()} from Scryfall!`, 'success');
    } catch (error) {
      showToast.show(`Failed to refresh cards: ${error}`, 'error');
    } finally {
      setIsFetchingCards(false);
    }
  };

  const handleRecalculateGrades = async () => {
    setIsRecalculating(true);
    setRecalculateMessage('');

    try {
      const count = await RecalculateAllDraftGrades();

      setRecalculateMessage(`✓ Successfully recalculated ${count} draft session(s)! Draft grades and predictions have been updated.`);

      // Clear message after 5 seconds
      setTimeout(() => setRecalculateMessage(''), 5000);
    } catch (error) {
      setRecalculateMessage(`✗ Failed to recalculate draft grades: ${error}`);

      // Clear error message after 8 seconds
      setTimeout(() => setRecalculateMessage(''), 8000);
    } finally {
      setIsRecalculating(false);
    }
  };

  const handleClearDatasetCache = async () => {
    setIsClearingCache(true);
    try {
      await ClearDatasetCache();
      showToast.show('Successfully cleared dataset cache! Cached CSV files have been deleted to free up disk space. Ratings in the database are preserved.', 'success');
    } catch (error) {
      showToast.show(`Failed to clear dataset cache: ${error}`, 'error');
    } finally {
      setIsClearingCache(false);
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
              Import JSON Export
              <span className="setting-description">Import match data from a JSON file exported by this app (matches only, not full log data)</span>
            </label>
            <div className="setting-control">
              <button className="action-button" onClick={handleImportData}>
                Import from JSON
              </button>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Import Single Log File
              <span className="setting-description">
                Import one MTGA log file from anywhere (backup drive, shared file, etc.). Processes the selected file and imports all data (matches, decks, quests, drafts).
              </span>
            </label>
            <div className="setting-control">
              <button className="action-button" onClick={handleImportLogFile}>
                Select Log File...
              </button>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Replay All MTGA Logs (Daemon Only)
              <span className="setting-description">
                Auto-discover and process ALL log files from your MTGA installation directory in chronological order.
                Use this for complete recovery after fresh install or extended daemon downtime. Requires daemon connection.
              </span>
            </label>
            <div className="setting-control">
              <div className="checkbox-container">
                <label className="checkbox-label">
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
                <div className="setting-hint settings-daemon-hint">
                  Daemon must be running to replay logs
                </div>
              )}
            </div>
          </div>

          {(isReplaying || replayProgress) && (
            <div className="setting-item">
              <div className="replay-progress-container">
                <h3 className={`replay-progress-title ${isReplaying ? '' : 'complete'}`}>
                  {isReplaying ? 'Replaying Historical Logs...' : '✓ Replay Complete'}
                </h3>
                {replayProgress && (
                  <>
                    <div className="settings-grid-2col">
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
                      <div className="current-file-display">
                        Current: {replayProgress.currentFile}
                      </div>
                    )}
                    {isReplaying && (
                      <div className="settings-progress-bar">
                        <div
                          className="settings-progress-bar-fill"
                          style={{ width: `${((replayProgress.processedFiles || 0) / (replayProgress.totalFiles || 1)) * 100}%` }}
                        ></div>
                      </div>
                    )}
                    {!isReplaying && (
                      <div className="refresh-message">
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
                  showToast.show('All data has been cleared successfully!', 'success');
                  window.location.reload(); // Refresh to show empty state
                } catch (error) {
                  showToast.show(`Failed to clear data: ${error}`, 'error');
                }
              }}>
                Clear All Data
              </button>
            </div>
          </div>
        </div>

        {/* Replay Testing Tool */}
        <div className="settings-section">
          <h2 className="section-title">Replay Testing Tool (Daemon Only)</h2>
          <div className="setting-description settings-section-description">
            Test draft and event features by replaying historical log files at variable speeds.
            While replay is active, navigate to Draft or Events pages to watch them populate in real-time.
            <div className="settings-warning-box">
              ⚠️ <strong>Important:</strong> Clear draft/event data before starting replay to avoid database conflicts.
              Existing draft sessions with the same ID will cause storage failures.
            </div>
          </div>

          {connectionStatus.status !== 'connected' && (
            <div className="setting-hint settings-daemon-warning">
              ⚠️ Replay tool requires daemon mode. Please start the daemon service to use this feature.
            </div>
          )}

          {!replayToolActive && (
            <>
              <div className="setting-item">
                <label className="setting-label">
                  Replay Speed
                  <span className="setting-description">How fast to replay events (1x = real-time, 10x = 10x faster)</span>
                </label>
                <div className="setting-control">
                  <input
                    type="range"
                    min="1"
                    max="100"
                    step="1"
                    value={replaySpeed}
                    onChange={(e) => setReplaySpeed(parseFloat(e.target.value))}
                    className="slider-input input-width-200"
                  />
                  <span className="slider-value">
                    {replaySpeed}x
                  </span>
                </div>
              </div>

              <div className="setting-item">
                <label className="setting-label">
                  Event Filter
                  <span className="setting-description">Filter which types of events to replay</span>
                </label>
                <div className="setting-control">
                  <select
                    className="select-input select-width-200"
                    value={replayFilter}
                    onChange={(e) => setReplayFilter(e.target.value)}
                  >
                    <option value="all">All Events</option>
                    <option value="draft">Draft Only</option>
                    <option value="match">Matches Only</option>
                    <option value="event">Events Only</option>
                  </select>
                </div>
              </div>

              <div className="setting-item">
                <label className="setting-label">
                  <input
                    type="checkbox"
                    checked={pauseOnDraft}
                    onChange={(e) => setPauseOnDraft(e.target.checked)}
                    className="checkbox-input"
                  />
                  Auto-pause on draft events
                  <span className="setting-description">
                    Automatically pause when draft events are detected so you can navigate to the Draft tab to watch the active draft populate in real-time
                  </span>
                </label>
              </div>

              <div className="setting-item">
                <label className="setting-label">
                  Start Replay
                  <span className="setting-description">Select one or more log files and start replay with the settings above</span>
                </label>
                <div className="setting-control">
                  <button
                    className="action-button primary"
                    onClick={handleStartReplayTool}
                    disabled={connectionStatus.status !== 'connected'}
                  >
                    Select Log File(s) & Start
                  </button>
                </div>
              </div>
            </>
          )}

          {replayToolActive && (
            <div className="replay-tool-controls">
              <h3 className={`replay-tool-title ${replayToolPaused ? 'paused' : 'active'}`}>
                {replayToolPaused ? '⏸️ Replay Paused' : '▶️ Replay Active'}
                {replayToolProgress && (
                  <span className="replay-tool-subtitle">
                    ({replayToolProgress.speed || replaySpeed}x speed, {replayToolProgress.filter || replayFilter})
                  </span>
                )}
              </h3>

              {replayToolProgress && (
                <>
                  <div className="settings-grid-2col-wide">
                    <div>
                      <strong>Progress:</strong> {replayToolProgress.currentEntry || 0} / {replayToolProgress.totalEntries || 0} entries
                    </div>
                    <div>
                      <strong>Complete:</strong> {(replayToolProgress.percentComplete || 0).toFixed(1)}%
                    </div>
                    <div>
                      <strong>Elapsed:</strong> {(replayToolProgress.elapsed || 0).toFixed(1)}s
                    </div>
                  </div>

                  <div className="settings-progress-bar with-margin-bottom">
                    <div
                      className={`settings-progress-bar-fill ${replayToolPaused ? 'paused' : ''}`}
                      style={{ width: `${replayToolProgress.percentComplete || 0}%` }}
                    ></div>
                  </div>
                </>
              )}

              <div className="replay-tool-buttons">
                {!replayToolPaused && (
                  <button
                    className="action-button pause"
                    onClick={handlePauseReplayTool}
                  >
                    ⏸️ Pause
                  </button>
                )}
                {replayToolPaused && (
                  <button
                    className="action-button resume"
                    onClick={handleResumeReplayTool}
                  >
                    ▶️ Resume
                  </button>
                )}
                <button
                  className="danger-button"
                  onClick={handleStopReplayTool}
                >
                  ⏹️ Stop
                </button>
              </div>

              <div className="settings-info-box-dark">
                💡 <strong>Tip:</strong> Navigate to the Draft or Events page to watch them populate in real-time as the replay progresses!
              </div>
            </div>
          )}
        </div>

        {/* 17Lands Data Management */}
        <div className="settings-section">
          <h2 className="section-title">17Lands Card Ratings</h2>
          <div className="setting-description settings-section-description">
            Download card ratings and tier lists from 17Lands for draft assistance.
            Ratings are used in the Active Draft page to show pick quality and recommendations.
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Set Code
              <span className="setting-description">
                The MTG set code (e.g., TLA = Avatar: The Last Airbender, BLB = Bloomburrow, DSK = Duskmourn)
              </span>
            </label>
            <div className="setting-control">
              <input
                type="text"
                value={setCode}
                onChange={(e) => setSetCode(e.target.value.toUpperCase())}
                placeholder="e.g., TLA, BLB, DSK, FDN"
                className="text-input input-width-200"
                maxLength={5}
              />
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Draft Format
              <span className="setting-description">Choose between Premier Draft (BO1) or Quick Draft ratings</span>
            </label>
            <div className="setting-control">
              <select
                className="select-input select-width-200"
                value={draftFormat}
                onChange={(e) => setDraftFormat(e.target.value)}
              >
                <option value="PremierDraft">Premier Draft (BO1)</option>
                <option value="QuickDraft">Quick Draft</option>
                <option value="TradDraft">Traditional Draft (BO3)</option>
              </select>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Fetch Ratings
              <span className="setting-description">Download and cache 17Lands ratings for the selected set and format</span>
            </label>
            <div className="setting-control">
              <button
                className="action-button primary button-margin-right"
                onClick={handleFetchSetRatings}
                disabled={isFetchingRatings || !setCode}
              >
                {isFetchingRatings ? 'Fetching...' : 'Fetch Ratings'}
              </button>
              <button
                className="action-button"
                onClick={handleRefreshSetRatings}
                disabled={isFetchingRatings || !setCode}
              >
                Refresh (Re-download)
              </button>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Fetch Card Data (Scryfall)
              <span className="setting-description">Download and cache card details (names, images, text) from Scryfall for the selected set</span>
            </label>
            <div className="setting-control">
              <button
                className="action-button primary button-margin-right"
                onClick={handleFetchSetCards}
                disabled={isFetchingCards || !setCode}
              >
                {isFetchingCards ? 'Fetching...' : 'Fetch Card Data'}
              </button>
              <button
                className="action-button"
                onClick={handleRefreshSetCards}
                disabled={isFetchingCards || !setCode}
              >
                Refresh (Re-download)
              </button>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Recalculate Draft Grades
              <span className="setting-description">Update all draft grades and predictions with the latest 17Lands card ratings</span>
            </label>
            <div className="setting-control">
              <button
                className="action-button recalculate"
                onClick={handleRecalculateGrades}
                disabled={isRecalculating}
              >
                {isRecalculating ? 'Recalculating...' : 'Recalculate All Drafts'}
              </button>
              {recalculateMessage && (
                <div className={`recalculate-message ${recalculateMessage.startsWith('✓') ? 'success' : 'error'}`}>
                  {recalculateMessage}
                </div>
              )}
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Clear Dataset Cache
              <span className="setting-description">Remove cached 17Lands CSV files to free up disk space (ratings in database are preserved)</span>
            </label>
            <div className="setting-control">
              <button
                className="action-button clear-cache"
                onClick={handleClearDatasetCache}
                disabled={isClearingCache}
              >
                {isClearingCache ? 'Clearing...' : 'Clear Dataset Cache'}
              </button>
            </div>
          </div>

          {dataSource && (
            <div className="setting-hint settings-success-box">
              <strong>📊 Current Data Source:</strong> {
                dataSource === 's3' ? 'S3 Public Datasets (recommended, uses locally cached game data)' :
                dataSource === 'web_api' ? 'Web API (fallback for new sets like TLA)' :
                dataSource === 'legacy_api' ? 'Legacy API (older mode)' :
                dataSource
              }
            </div>
          )}

          <div className="setting-hint settings-info-box">
            <strong>💡 Common Set Codes:</strong>
            <div className="set-codes-grid">
              <div>• TLA - Avatar: The Last Airbender</div>
              <div>• FDN - Foundations</div>
              <div>• DSK - Duskmourn</div>
              <div>• BLB - Bloomburrow</div>
              <div>• OTJ - Outlaws of Thunder Junction</div>
              <div>• MKM - Murders at Karlov Manor</div>
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
            <div className="setting-control about-button-container">
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

import { useState } from 'react';
import AboutDialog from '../components/AboutDialog';
import LoadingButton from '../components/LoadingButton';
import { SettingItem, SettingToggle, SettingSelect } from '../components/settings';
import {
  useDaemonConnection,
  useLogReplay,
  useReplayTool,
  useSeventeenLands,
  useDataManagement,
} from '../hooks';
import './Settings.css';

const Settings = () => {
  // Local UI state (not extracted to hooks)
  const [dbPath, setDbPath] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showNotifications, setShowNotifications] = useState(true);
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

  // Custom hooks for state management
  const {
    connectionStatus,
    daemonMode,
    daemonPort,
    isReconnecting,
    handleDaemonPortChange,
    handleReconnect,
    handleModeChange,
  } = useDaemonConnection();

  const {
    clearDataBeforeReplay,
    setClearDataBeforeReplay,
    isReplaying,
    replayProgress,
    handleReplayLogs,
  } = useLogReplay();

  const {
    replayToolActive,
    replayToolPaused,
    replayToolProgress,
    replaySpeed,
    setReplaySpeed,
    replayFilter,
    setReplayFilter,
    pauseOnDraft,
    setPauseOnDraft,
    handleStartReplayTool,
    handlePauseReplayTool,
    handleResumeReplayTool,
    handleStopReplayTool,
  } = useReplayTool();

  const {
    setCode,
    setSetCode,
    draftFormat,
    setDraftFormat,
    isFetchingRatings,
    isFetchingCards,
    isRecalculating,
    recalculateMessage,
    dataSource,
    isClearingCache,
    handleFetchSetRatings,
    handleRefreshSetRatings,
    handleFetchSetCards,
    handleRefreshSetCards,
    handleRecalculateGrades,
    handleClearDatasetCache,
  } = useSeventeenLands();

  const {
    handleExportData,
    handleImportData,
    handleImportLogFile,
    handleClearAllData,
  } = useDataManagement();

  // Derived state
  const isConnected = connectionStatus.status === 'connected';

  // Local handlers
  const handleSave = () => {
    // TODO: Implement backend settings save
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const handleReset = () => {
    setDbPath('');
    setAutoRefresh(false);
    setRefreshInterval(30);
    setShowNotifications(true);
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

          <SettingSelect
            label="Connection Mode"
            description="Choose how the app connects to the daemon"
            value={daemonMode}
            onChange={handleModeChange}
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
              onChange={(e) => handleDaemonPortChange(parseInt(e.target.value))}
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
              onClick={handleReconnect}
              disabled={daemonMode === 'standalone'}
            >
              Reconnect to Daemon
            </LoadingButton>
          </SettingItem>
        </div>

        {/* Application Preferences */}
        <div className="settings-section">
          <h2 className="section-title">Application Preferences</h2>

          <SettingToggle
            label="Auto-refresh data"
            description="Automatically refresh statistics at regular intervals"
            checked={autoRefresh}
            onChange={setAutoRefresh}
          />

          {autoRefresh && (
            <SettingItem
              label="Refresh Interval (seconds)"
              description="How often to refresh data automatically"
              indented
            >
              <input
                type="number"
                value={refreshInterval}
                onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                min="10"
                max="300"
                step="10"
                className="number-input"
              />
            </SettingItem>
          )}

          <SettingToggle
            label="Show notifications"
            description="Display notifications for match results and updates"
            checked={showNotifications}
            onChange={setShowNotifications}
          />
        </div>

        {/* Data Management */}
        <div className="settings-section">
          <h2 className="section-title">Data Management</h2>

          <SettingItem
            label="Export Data"
            description="Export your match history and statistics to a file"
          >
            <button className="action-button" onClick={() => handleExportData('json')}>
              Export to JSON
            </button>
            <button className="action-button" onClick={() => handleExportData('csv')}>
              Export to CSV
            </button>
          </SettingItem>

          <SettingItem
            label="Import JSON Export"
            description="Import match data from a JSON file exported by this app (matches only, not full log data)"
          >
            <button className="action-button" onClick={handleImportData}>
              Import from JSON
            </button>
          </SettingItem>

          <SettingItem
            label="Import Single Log File"
            description="Import one MTGA log file from anywhere (backup drive, shared file, etc.). Processes the selected file and imports all data (matches, decks, quests, drafts)."
          >
            <button className="action-button" onClick={handleImportLogFile}>
              Select Log File...
            </button>
          </SettingItem>

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
              <LoadingButton
                loading={isReplaying}
                loadingText="Replaying Logs..."
                onClick={() => handleReplayLogs(isConnected)}
                disabled={!isConnected}
                variant="primary"
              >
                Replay Historical Logs
              </LoadingButton>
              {!isConnected && (
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
              <button className="danger-button" onClick={handleClearAllData}>
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

          {!isConnected && (
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
                    onClick={() => handleStartReplayTool(isConnected)}
                    disabled={!isConnected}
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

          <SettingSelect
            label="Draft Format"
            description="Choose between Premier Draft (BO1) or Quick Draft ratings"
            value={draftFormat}
            onChange={setDraftFormat}
            options={[
              { value: 'PremierDraft', label: 'Premier Draft (BO1)' },
              { value: 'QuickDraft', label: 'Quick Draft' },
              { value: 'TradDraft', label: 'Traditional Draft (BO3)' },
            ]}
          />

          <SettingItem
            label="Fetch Ratings"
            description="Download and cache 17Lands ratings for the selected set and format"
          >
            <LoadingButton
              loading={isFetchingRatings}
              loadingText="Fetching..."
              onClick={handleFetchSetRatings}
              disabled={!setCode}
              variant="primary"
              className="button-margin-right"
            >
              Fetch Ratings
            </LoadingButton>
            <button
              className="action-button"
              onClick={handleRefreshSetRatings}
              disabled={isFetchingRatings || !setCode}
            >
              Refresh (Re-download)
            </button>
          </SettingItem>

          <SettingItem
            label="Fetch Card Data (Scryfall)"
            description="Download and cache card details (names, images, text) from Scryfall for the selected set"
          >
            <LoadingButton
              loading={isFetchingCards}
              loadingText="Fetching..."
              onClick={handleFetchSetCards}
              disabled={!setCode}
              variant="primary"
              className="button-margin-right"
            >
              Fetch Card Data
            </LoadingButton>
            <button
              className="action-button"
              onClick={handleRefreshSetCards}
              disabled={isFetchingCards || !setCode}
            >
              Refresh (Re-download)
            </button>
          </SettingItem>

          <SettingItem
            label="Recalculate Draft Grades"
            description="Update all draft grades and predictions with the latest 17Lands card ratings"
          >
            <LoadingButton
              loading={isRecalculating}
              loadingText="Recalculating..."
              onClick={handleRecalculateGrades}
              variant="recalculate"
            >
              Recalculate All Drafts
            </LoadingButton>
            {recalculateMessage && (
              <div className={`recalculate-message ${recalculateMessage.startsWith('✓') ? 'success' : 'error'}`}>
                {recalculateMessage}
              </div>
            )}
          </SettingItem>

          <SettingItem
            label="Clear Dataset Cache"
            description="Remove cached 17Lands CSV files to free up disk space (ratings in database are preserved)"
          >
            <LoadingButton
              loading={isClearingCache}
              loadingText="Clearing..."
              onClick={handleClearDatasetCache}
              variant="clear-cache"
            >
              Clear Dataset Cache
            </LoadingButton>
          </SettingItem>

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

          <SettingSelect
            label="Theme"
            description="Choose your preferred color scheme"
            value="dark"
            onChange={() => {}}
            options={[
              { value: 'dark', label: 'Dark (Default)' },
              { value: 'light', label: 'Light (Coming Soon)' },
              { value: 'auto', label: 'Auto (System Default)' },
            ]}
          />
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

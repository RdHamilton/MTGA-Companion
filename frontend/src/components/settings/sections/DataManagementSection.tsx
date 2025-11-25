import LoadingButton from '../../LoadingButton';
import { SettingItem } from '../';
import { gui } from '../../../../wailsjs/go/models';

export interface DataManagementSectionProps {
  isConnected: boolean;
  clearDataBeforeReplay: boolean;
  onClearDataBeforeReplayChange: (value: boolean) => void;
  isReplaying: boolean;
  replayProgress: gui.LogReplayProgress | null;
  onExportData: (format: 'json' | 'csv') => void;
  onImportData: () => void;
  onImportLogFile: () => void;
  onReplayLogs: () => void;
  onClearAllData: () => void;
}

export function DataManagementSection({
  isConnected,
  clearDataBeforeReplay,
  onClearDataBeforeReplayChange,
  isReplaying,
  replayProgress,
  onExportData,
  onImportData,
  onImportLogFile,
  onReplayLogs,
  onClearAllData,
}: DataManagementSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">Data Management</h2>

      <SettingItem
        label="Export Data"
        description="Export your match history and statistics to a file"
      >
        <button className="action-button" onClick={() => onExportData('json')}>
          Export to JSON
        </button>
        <button className="action-button" onClick={() => onExportData('csv')}>
          Export to CSV
        </button>
      </SettingItem>

      <SettingItem
        label="Import JSON Export"
        description="Import match data from a JSON file exported by this app (matches only, not full log data)"
      >
        <button className="action-button" onClick={onImportData}>
          Import from JSON
        </button>
      </SettingItem>

      <SettingItem
        label="Import Single Log File"
        description="Import one MTGA log file from anywhere (backup drive, shared file, etc.). Processes the selected file and imports all data (matches, decks, quests, drafts)."
      >
        <button className="action-button" onClick={onImportLogFile}>
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
                onChange={(e) => onClearDataBeforeReplayChange(e.target.checked)}
                className="checkbox-input"
                disabled={isReplaying}
              />
              <span>Clear all data before replay (recommended for first-time setup)</span>
            </label>
          </div>
          <LoadingButton
            loading={isReplaying}
            loadingText="Replaying Logs..."
            onClick={onReplayLogs}
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
          <button className="danger-button" onClick={onClearAllData}>
            Clear All Data
          </button>
        </div>
      </div>
    </div>
  );
}

import { gui } from '../../../../wailsjs/go/models';

export interface ReplayToolSectionProps {
  isConnected: boolean;
  replayToolActive: boolean;
  replayToolPaused: boolean;
  replayToolProgress: gui.ReplayStatus | null;
  replaySpeed: number;
  onReplaySpeedChange: (speed: number) => void;
  replayFilter: string;
  onReplayFilterChange: (filter: string) => void;
  pauseOnDraft: boolean;
  onPauseOnDraftChange: (pause: boolean) => void;
  onStartReplayTool: () => void;
  onPauseReplayTool: () => void;
  onResumeReplayTool: () => void;
  onStopReplayTool: () => void;
}

export function ReplayToolSection({
  isConnected,
  replayToolActive,
  replayToolPaused,
  replayToolProgress,
  replaySpeed,
  onReplaySpeedChange,
  replayFilter,
  onReplayFilterChange,
  pauseOnDraft,
  onPauseOnDraftChange,
  onStartReplayTool,
  onPauseReplayTool,
  onResumeReplayTool,
  onStopReplayTool,
}: ReplayToolSectionProps) {
  return (
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
                onChange={(e) => onReplaySpeedChange(parseFloat(e.target.value))}
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
                onChange={(e) => onReplayFilterChange(e.target.value)}
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
                onChange={(e) => onPauseOnDraftChange(e.target.checked)}
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
                onClick={onStartReplayTool}
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
                onClick={onPauseReplayTool}
              >
                ⏸️ Pause
              </button>
            )}
            {replayToolPaused && (
              <button
                className="action-button resume"
                onClick={onResumeReplayTool}
              >
                ▶️ Resume
              </button>
            )}
            <button
              className="danger-button"
              onClick={onStopReplayTool}
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
  );
}

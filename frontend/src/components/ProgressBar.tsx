import './ProgressBar.css';

export interface ProgressBarProps {
  /** Progress percentage (0-100) */
  progress: number;
  /** Label to display alongside progress */
  label?: string;
  /** Optional detail text under the progress bar */
  detail?: string;
  /** Show cancel button */
  showCancel?: boolean;
  /** Cancel button callback */
  onCancel?: () => void;
  /** Size variant */
  size?: 'small' | 'medium' | 'large';
  /** Color variant */
  variant?: 'primary' | 'success' | 'warning' | 'error';
  /** Show percentage text */
  showPercentage?: boolean;
  /** Indeterminate mode (animating without specific progress) */
  indeterminate?: boolean;
  /** Estimated time remaining in milliseconds */
  estimatedTimeRemaining?: number;
  /** Additional CSS class */
  className?: string;
}

function formatTimeRemaining(ms: number): string {
  if (ms < 1000) return 'Less than a second';
  const seconds = Math.ceil(ms / 1000);
  if (seconds < 60) return `~${seconds}s remaining`;
  const minutes = Math.ceil(seconds / 60);
  return `~${minutes}m remaining`;
}

export default function ProgressBar({
  progress,
  label,
  detail,
  showCancel = false,
  onCancel,
  size = 'medium',
  variant = 'primary',
  showPercentage = true,
  indeterminate = false,
  estimatedTimeRemaining,
  className = '',
}: ProgressBarProps) {
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <div
      className={`progress-bar-container progress-bar-${size} ${className}`}
      role="progressbar"
      aria-valuenow={indeterminate ? undefined : clampedProgress}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={label || 'Progress'}
    >
      {/* Header row with label and percentage */}
      {(label || showPercentage) && (
        <div className="progress-bar-header">
          {label && <span className="progress-bar-label">{label}</span>}
          {showPercentage && !indeterminate && (
            <span className="progress-bar-percentage">{Math.round(clampedProgress)}%</span>
          )}
        </div>
      )}

      {/* Progress bar track */}
      <div className="progress-bar-track">
        <div
          className={`progress-bar-fill progress-bar-${variant} ${indeterminate ? 'indeterminate' : ''}`}
          style={indeterminate ? undefined : { width: `${clampedProgress}%` }}
        />
      </div>

      {/* Footer row with detail and time remaining */}
      {(detail || estimatedTimeRemaining || showCancel) && (
        <div className="progress-bar-footer">
          <div className="progress-bar-info">
            {detail && <span className="progress-bar-detail">{detail}</span>}
            {estimatedTimeRemaining !== undefined && estimatedTimeRemaining > 0 && (
              <span className="progress-bar-time">{formatTimeRemaining(estimatedTimeRemaining)}</span>
            )}
          </div>
          {showCancel && onCancel && (
            <button
              type="button"
              className="progress-bar-cancel"
              onClick={onCancel}
              aria-label="Cancel operation"
            >
              Cancel
            </button>
          )}
        </div>
      )}
    </div>
  );
}

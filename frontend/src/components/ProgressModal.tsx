import { useEffect, useCallback } from 'react';
import ProgressBar from './ProgressBar';
import './ProgressModal.css';

export interface ProgressModalProps {
  /** Whether the modal is visible */
  isOpen: boolean;
  /** Modal title */
  title: string;
  /** Progress percentage (0-100) */
  progress: number;
  /** Detailed status message */
  detail?: string;
  /** Estimated time remaining in milliseconds */
  estimatedTimeRemaining?: number;
  /** Whether the operation can be cancelled */
  cancellable?: boolean;
  /** Cancel callback */
  onCancel?: () => void;
  /** Whether to use indeterminate progress (pulsing animation) */
  indeterminate?: boolean;
  /** Optional icon to display (emoji or icon class) */
  icon?: string;
  /** Progress bar color variant */
  variant?: 'primary' | 'success' | 'warning' | 'error';
}

export default function ProgressModal({
  isOpen,
  title,
  progress,
  detail,
  estimatedTimeRemaining,
  cancellable = false,
  onCancel,
  indeterminate = false,
  icon,
  variant = 'primary',
}: ProgressModalProps) {
  // Handle Escape key to cancel if cancellable
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape' && cancellable && onCancel) {
        onCancel();
      }
    },
    [cancellable, onCancel]
  );

  useEffect(() => {
    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden';
    }
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = '';
    };
  }, [isOpen, handleKeyDown]);

  if (!isOpen) return null;

  return (
    <div className="progress-modal-overlay" role="dialog" aria-modal="true" aria-labelledby="progress-modal-title">
      <div className="progress-modal">
        {/* Header */}
        <div className="progress-modal-header">
          {icon && <span className="progress-modal-icon">{icon}</span>}
          <h2 id="progress-modal-title" className="progress-modal-title">
            {title}
          </h2>
        </div>

        {/* Progress content */}
        <div className="progress-modal-content">
          <ProgressBar
            progress={progress}
            detail={detail}
            showPercentage={!indeterminate}
            indeterminate={indeterminate}
            estimatedTimeRemaining={estimatedTimeRemaining}
            variant={variant}
            size="large"
          />
        </div>

        {/* Cancel button */}
        {cancellable && onCancel && (
          <div className="progress-modal-actions">
            <button type="button" className="progress-modal-cancel" onClick={onCancel}>
              Cancel
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

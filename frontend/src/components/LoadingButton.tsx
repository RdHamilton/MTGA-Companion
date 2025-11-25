import './LoadingButton.css';

export interface LoadingButtonProps {
  /** Whether the button is in a loading state */
  loading: boolean;
  /** Text to display when loading */
  loadingText: string;
  /** Click handler */
  onClick: () => void;
  /** Whether the button is disabled (independent of loading state) */
  disabled?: boolean;
  /** Button variant for styling */
  variant?: 'primary' | 'danger' | 'pause' | 'resume' | 'recalculate' | 'clear-cache' | 'default';
  /** Additional CSS class names */
  className?: string;
  /** Button content when not loading */
  children: React.ReactNode;
}

const LoadingButton = ({
  loading,
  loadingText,
  onClick,
  disabled = false,
  variant = 'default',
  className = '',
  children,
}: LoadingButtonProps) => {
  const variantClass = variant !== 'default' ? variant : '';
  const combinedClassName = `action-button ${variantClass} ${className}`.trim();

  return (
    <button
      className={`loading-button ${combinedClassName} ${loading ? 'loading' : ''}`}
      onClick={onClick}
      disabled={disabled || loading}
    >
      {loading && <span className="loading-button-spinner" />}
      <span className={loading ? 'loading-button-text' : ''}>
        {loading ? loadingText : children}
      </span>
    </button>
  );
};

export default LoadingButton;

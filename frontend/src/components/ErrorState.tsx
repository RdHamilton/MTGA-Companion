import './ErrorState.css';

interface ErrorStateProps {
  message: string;
  error?: Error | string;
  helpText?: string;
}

const ErrorState = ({ message, error, helpText }: ErrorStateProps) => {
  const errorDetails = error instanceof Error ? error.message : error;

  return (
    <div className="error-state">
      <div className="error-state-icon">⚠️</div>
      <h2 className="error-state-title">{message}</h2>
      {errorDetails && <p className="error-state-details">{errorDetails}</p>}
      {helpText && <p className="error-state-help">{helpText}</p>}
    </div>
  );
};

export default ErrorState;

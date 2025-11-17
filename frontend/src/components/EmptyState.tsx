import './EmptyState.css';

interface EmptyStateProps {
  icon?: string;
  title: string;
  message: string;
  helpText?: string;
}

const EmptyState = ({ icon, title, message, helpText }: EmptyStateProps) => {
  return (
    <div className="empty-state">
      {icon && <div className="empty-state-icon">{icon}</div>}
      <h2 className="empty-state-title">{title}</h2>
      <p className="empty-state-message">{message}</p>
      {helpText && <p className="empty-state-help">{helpText}</p>}
    </div>
  );
};

export default EmptyState;

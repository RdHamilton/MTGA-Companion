import { useEffect, useState } from 'react';
import './Toast.css';

interface ToastProps {
  message: string;
  type?: 'success' | 'info' | 'warning' | 'error';
  duration?: number;
  onClose?: () => void;
}

const Toast = ({ message, type = 'info', duration = 3000, onClose }: ToastProps) => {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false);
      if (onClose) {
        setTimeout(onClose, 300); // Wait for fade out animation
      }
    }, duration);

    return () => clearTimeout(timer);
  }, [duration, onClose]);

  if (!isVisible) return null;

  return (
    <div className={`toast toast-${type} ${!isVisible ? 'toast-hide' : ''}`}>
      <div className="toast-icon">
        {type === 'success' && '✓'}
        {type === 'info' && 'ℹ'}
        {type === 'warning' && '⚠'}
        {type === 'error' && '✗'}
      </div>
      <div className="toast-message">{message}</div>
    </div>
  );
};

export default Toast;

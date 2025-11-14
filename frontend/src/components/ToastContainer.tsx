import { useState, useCallback, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import Toast from './Toast';

interface ToastData {
  id: number;
  message: string;
  type: 'success' | 'info' | 'warning' | 'error';
}

let toastIdCounter = 0;

const ToastContainer = () => {
  const [toasts, setToasts] = useState<ToastData[]>([]);

  const addToast = useCallback((message: string, type: 'success' | 'info' | 'warning' | 'error' = 'info') => {
    const id = toastIdCounter++;
    setToasts(prev => [...prev, { id, message, type }]);
  }, []);

  const removeToast = useCallback((id: number) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  }, []);

  useEffect(() => {
    // Listen for stats:updated events from backend
    const unsubscribe = EventsOn('stats:updated', (data: any) => {
      const matches = data?.matches || 0;
      const games = data?.games || 0;

      if (matches > 0) {
        addToast(
          `New match detected! ${matches} match${matches > 1 ? 'es' : ''}, ${games} game${games > 1 ? 's' : ''} - Stats updated`,
          'success'
        );
      }
    });

    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, [addToast]);

  return (
    <div style={{ position: 'fixed', bottom: 0, right: 0, zIndex: 10000 }}>
      {toasts.map((toast, index) => (
        <div key={toast.id} style={{ marginBottom: index > 0 ? '10px' : '0' }}>
          <Toast
            message={toast.message}
            type={toast.type}
            onClose={() => removeToast(toast.id)}
          />
        </div>
      ))}
    </div>
  );
};

// Export the global toast function for use in other components
export const showToast = (() => {
  let addToastFn: ((message: string, type?: 'success' | 'info' | 'warning' | 'error') => void) | null = null;

  return {
    setAddFn: (fn: (message: string, type?: 'success' | 'info' | 'warning' | 'error') => void) => {
      addToastFn = fn;
    },
    show: (message: string, type: 'success' | 'info' | 'warning' | 'error' = 'info') => {
      if (addToastFn) {
        addToastFn(message, type);
      }
    }
  };
})();

export default ToastContainer;

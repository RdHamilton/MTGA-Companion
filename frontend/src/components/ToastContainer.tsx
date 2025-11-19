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

    // Auto-remove toast after 5 seconds
    setTimeout(() => {
      setToasts(prev => prev.filter(toast => toast.id !== id));
    }, 5000);
  }, []);

  const removeToast = useCallback((id: number) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  }, []);

  useEffect(() => {
    // Register global toast function
    showToast.setAddFn(addToast);
  }, [addToast]);

  useEffect(() => {
    // Listen for stats:updated events from backend
    const unsubscribeStats = EventsOn('stats:updated', (data: any) => {
      const matches = data?.matches || 0;
      const games = data?.games || 0;

      if (matches > 0) {
        addToast(
          `New match detected! ${matches} match${matches > 1 ? 'es' : ''}, ${games} game${games > 1 ? 's' : ''} - Stats updated`,
          'success'
        );
      }
    });

    // Listen for rank:updated events
    const unsubscribeRank = EventsOn('rank:updated', (data: any) => {
      const format = data?.format || 'Ranked';
      const tier = data?.tier || '';
      const step = data?.step || '';

      if (tier && step) {
        addToast(
          `Rank updated: ${format} ${tier} ${step}`,
          'info'
        );
      }
    });

    // Listen for quest update events
    const unsubscribeQuest = EventsOn('quest:updated', (data: any) => {
      const completed = data?.completed || 0;
      const count = data?.count || 0;

      if (completed > 0) {
        addToast(
          `Quest${completed > 1 ? 's' : ''} completed! (${completed})`,
          'success'
        );
      } else if (count > 0) {
        addToast(
          `Quest${count > 1 ? 's' : ''} updated (${count})`,
          'info'
        );
      }
    });

    // Listen for draft update events
    const unsubscribeDraft = EventsOn('draft:updated', (data: any) => {
      const count = data?.count || 0;
      const picks = data?.picks || 0;

      if (count > 0) {
        addToast(
          `Draft session${count > 1 ? 's' : ''} stored! (${count} session${count > 1 ? 's' : ''}, ${picks} pick${picks !== 1 ? 's' : ''})`,
          'success'
        );
      }
    });

    return () => {
      if (unsubscribeStats) unsubscribeStats();
      if (unsubscribeRank) unsubscribeRank();
      if (unsubscribeQuest) unsubscribeQuest();
      if (unsubscribeDraft) unsubscribeDraft();
    };
  }, [addToast]);

  return (
    <div style={{ position: 'fixed', bottom: '20px', right: '20px', zIndex: 10000, display: 'flex', flexDirection: 'column-reverse' }}>
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          message={toast.message}
          type={toast.type}
          onClose={() => removeToast(toast.id)}
        />
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

import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

// Detect platform (macOS uses metaKey/Cmd, others use ctrlKey)
const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;

interface KeyboardShortcutConfig {
  onRefresh?: () => void;
}

export const useKeyboardShortcuts = (config?: KeyboardShortcutConfig) => {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Use Cmd on Mac, Ctrl on Windows/Linux
      const modifier = isMac ? event.metaKey : event.ctrlKey;

      // Ignore if user is typing in an input/textarea
      const target = event.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
        return;
      }

      if (modifier) {
        switch (event.key) {
          case '1':
            event.preventDefault();
            navigate('/match-history');
            break;
          case '2':
            event.preventDefault();
            navigate('/quests');
            break;
          case '3':
            event.preventDefault();
            navigate('/events');
            break;
          case '4':
            event.preventDefault();
            navigate('/charts/win-rate-trend');
            break;
          case '5':
            event.preventDefault();
            navigate('/settings');
            break;
          case 'r':
          case 'R':
            event.preventDefault();
            if (config?.onRefresh) {
              config.onRefresh();
            } else {
              // Default refresh: reload current page data
              window.location.reload();
            }
            break;
          case ',':
            // Cmd+, or Ctrl+, -> Settings (macOS standard)
            event.preventDefault();
            navigate('/settings');
            break;
          default:
            break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [navigate, location, config]);
};

export default useKeyboardShortcuts;

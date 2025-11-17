import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { EventsEmit } from '../../wailsjs/runtime/runtime';

// Detect platform (macOS uses metaKey/Cmd, others use ctrlKey)
const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;

/**
 * KeyboardShortcutsHandler provides global keyboard shortcut support
 *
 * Shortcuts:
 * - Cmd/Ctrl + 1: Match History
 * - Cmd/Ctrl + 2: Quests
 * - Cmd/Ctrl + 3: Events
 * - Cmd/Ctrl + 4: Charts
 * - Cmd/Ctrl + 5: Settings
 * - Cmd/Ctrl + R: Refresh current page
 * - Cmd/Ctrl + ,: Settings (macOS standard)
 */
const KeyboardShortcutsHandler = () => {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Use Cmd on Mac, Ctrl on Windows/Linux
      const modifier = isMac ? event.metaKey : event.ctrlKey;

      // Ignore if user is typing in an input/textarea/select
      const target = event.target as HTMLElement;
      const isTyping =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable;

      if (isTyping) {
        return;
      }

      if (modifier) {
        let handled = false;

        switch (event.key) {
          case '1':
            navigate('/match-history');
            handled = true;
            break;
          case '2':
            navigate('/quests');
            handled = true;
            break;
          case '3':
            navigate('/events');
            handled = true;
            break;
          case '4':
            navigate('/charts/win-rate-trend');
            handled = true;
            break;
          case '5':
            navigate('/settings');
            handled = true;
            break;
          case 'r':
          case 'R':
            // Refresh current page by emitting a stats:updated event
            // This will trigger data reload in most pages
            EventsEmit('stats:updated');
            handled = true;
            break;
          case ',':
            // Cmd+, or Ctrl+, -> Settings (macOS standard)
            navigate('/settings');
            handled = true;
            break;
          default:
            break;
        }

        // Prevent default browser behavior for handled shortcuts
        if (handled) {
          event.preventDefault();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [navigate, location]);

  // This component doesn't render anything
  return null;
};

export default KeyboardShortcutsHandler;

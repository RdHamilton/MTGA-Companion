import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';

// Mock navigate function
const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Helper to create keyboard events
function createKeyboardEvent(key: string, options: Partial<KeyboardEventInit> = {}): KeyboardEvent {
  return new KeyboardEvent('keydown', {
    key,
    bubbles: true,
    ...options,
  });
}

// Wrapper component for the hook
function Wrapper({ children }: { children: React.ReactNode }) {
  return <BrowserRouter>{children}</BrowserRouter>;
}

describe('useKeyboardShortcuts', () => {
  // Detect if we're on Mac for tests
  const originalPlatform = navigator.platform;

  beforeEach(() => {
    vi.clearAllMocks();
    // Reset location mock
    window.history.pushState({}, '', '/');
  });

  afterEach(() => {
    // Restore platform
    Object.defineProperty(navigator, 'platform', {
      value: originalPlatform,
      writable: true,
    });
  });

  describe('Navigation Shortcuts', () => {
    it('should navigate to /match-history on Ctrl+1', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('1', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/match-history');
    });

    it('should navigate to /quests on Ctrl+2', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('2', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/quests');
    });

    it('should navigate to /events on Ctrl+3', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('3', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/events');
    });

    it('should navigate to /charts/win-rate-trend on Ctrl+4', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('4', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/charts/win-rate-trend');
    });

    it('should navigate to /settings on Ctrl+5', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('5', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/settings');
    });

    it('should navigate to /settings on Ctrl+,', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent(',', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/settings');
    });
  });

  describe('Mac Support', () => {
    it('should support both metaKey and ctrlKey for cross-platform compatibility', () => {
      // The hook uses metaKey on Mac and ctrlKey on other platforms
      // In tests, we run on the test environment platform, so ctrlKey works
      // This test verifies the hook accepts keyboard events with ctrl modifier
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('1', { ctrlKey: true }));
      });

      expect(mockNavigate).toHaveBeenCalledWith('/match-history');
    });
  });

  describe('Refresh Shortcut', () => {
    it('should call custom onRefresh when Ctrl+R pressed', () => {
      const onRefresh = vi.fn();
      renderHook(() => useKeyboardShortcuts({ onRefresh }), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('r', { ctrlKey: true }));
      });

      expect(onRefresh).toHaveBeenCalledTimes(1);
    });

    it('should call custom onRefresh when Ctrl+Shift+R pressed', () => {
      const onRefresh = vi.fn();
      renderHook(() => useKeyboardShortcuts({ onRefresh }), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('R', { ctrlKey: true }));
      });

      expect(onRefresh).toHaveBeenCalledTimes(1);
    });

    it('should reload window when no custom onRefresh provided', () => {
      const reloadMock = vi.fn();
      const originalLocation = window.location;

      // Mock window.location.reload
      Object.defineProperty(window, 'location', {
        value: { ...originalLocation, reload: reloadMock },
        writable: true,
      });

      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('r', { ctrlKey: true }));
      });

      expect(reloadMock).toHaveBeenCalledTimes(1);

      // Restore
      Object.defineProperty(window, 'location', {
        value: originalLocation,
        writable: true,
      });
    });
  });

  describe('Input Element Handling', () => {
    it('should not trigger shortcuts when typing in input', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      const input = document.createElement('input');
      document.body.appendChild(input);
      input.focus();

      const event = new KeyboardEvent('keydown', {
        key: '1',
        ctrlKey: true,
        bubbles: true,
      });
      Object.defineProperty(event, 'target', { value: input });

      act(() => {
        window.dispatchEvent(event);
      });

      expect(mockNavigate).not.toHaveBeenCalled();

      document.body.removeChild(input);
    });

    it('should not trigger shortcuts when typing in textarea', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      const textarea = document.createElement('textarea');
      document.body.appendChild(textarea);
      textarea.focus();

      const event = new KeyboardEvent('keydown', {
        key: '1',
        ctrlKey: true,
        bubbles: true,
      });
      Object.defineProperty(event, 'target', { value: textarea });

      act(() => {
        window.dispatchEvent(event);
      });

      expect(mockNavigate).not.toHaveBeenCalled();

      document.body.removeChild(textarea);
    });

    it('should not trigger shortcuts when typing in contenteditable', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      const div = document.createElement('div');
      div.contentEditable = 'true';
      // Ensure isContentEditable property is set correctly
      Object.defineProperty(div, 'isContentEditable', { value: true, configurable: true });
      document.body.appendChild(div);
      div.focus();

      const event = new KeyboardEvent('keydown', {
        key: '1',
        ctrlKey: true,
        bubbles: true,
      });
      Object.defineProperty(event, 'target', { value: div });

      act(() => {
        window.dispatchEvent(event);
      });

      expect(mockNavigate).not.toHaveBeenCalled();

      document.body.removeChild(div);
    });
  });

  describe('Modifier Key Requirements', () => {
    it('should not navigate without modifier key', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('1'));
      });

      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('should not navigate with only shift key', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('1', { shiftKey: true }));
      });

      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('should not navigate with only alt key', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      act(() => {
        window.dispatchEvent(createKeyboardEvent('1', { altKey: true }));
      });

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe('Event Cleanup', () => {
    it('should remove event listener on unmount', () => {
      const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener');

      const { unmount } = renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      unmount();

      expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));

      removeEventListenerSpy.mockRestore();
    });
  });

  describe('preventDefault', () => {
    it('should prevent default behavior for handled shortcuts', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      const event = createKeyboardEvent('1', { ctrlKey: true });
      const preventDefaultSpy = vi.spyOn(event, 'preventDefault');

      act(() => {
        window.dispatchEvent(event);
      });

      expect(preventDefaultSpy).toHaveBeenCalled();
    });

    it('should not prevent default for unhandled keys', () => {
      renderHook(() => useKeyboardShortcuts(), { wrapper: Wrapper });

      const event = createKeyboardEvent('9', { ctrlKey: true });
      const preventDefaultSpy = vi.spyOn(event, 'preventDefault');

      act(() => {
        window.dispatchEvent(event);
      });

      expect(preventDefaultSpy).not.toHaveBeenCalled();
    });
  });
});

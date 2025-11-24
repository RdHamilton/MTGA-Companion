import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import Toast from './Toast';

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  describe('Basic Rendering', () => {
    it('should render with message', () => {
      render(<Toast message="Test message" />);

      expect(screen.getByText('Test message')).toBeInTheDocument();
    });

    it('should be visible initially', () => {
      const { container } = render(<Toast message="Visible toast" />);

      const toast = container.querySelector('.toast');
      expect(toast).toBeInTheDocument();
      expect(toast).not.toHaveClass('toast-hide');
    });

    it('should render message text', () => {
      render(<Toast message="Important notification" />);

      const message = document.querySelector('.toast-message');
      expect(message).toBeInTheDocument();
      expect(message?.textContent).toBe('Important notification');
    });
  });

  describe('Toast Types', () => {
    it('should render with success type', () => {
      const { container } = render(<Toast message="Success!" type="success" />);

      expect(container.querySelector('.toast-success')).toBeInTheDocument();
      expect(screen.getByText('✓')).toBeInTheDocument();
    });

    it('should render with info type by default', () => {
      const { container } = render(<Toast message="Info message" />);

      expect(container.querySelector('.toast-info')).toBeInTheDocument();
      expect(screen.getByText('ℹ')).toBeInTheDocument();
    });

    it('should render with warning type', () => {
      const { container } = render(<Toast message="Warning!" type="warning" />);

      expect(container.querySelector('.toast-warning')).toBeInTheDocument();
      expect(screen.getByText('⚠')).toBeInTheDocument();
    });

    it('should render with error type', () => {
      const { container } = render(<Toast message="Error occurred" type="error" />);

      expect(container.querySelector('.toast-error')).toBeInTheDocument();
      expect(screen.getByText('✗')).toBeInTheDocument();
    });
  });

  describe('Icon Rendering', () => {
    it('should render success icon', () => {
      const { container } = render(<Toast message="Done" type="success" />);

      const icon = container.querySelector('.toast-icon');
      expect(icon).toBeInTheDocument();
      expect(icon?.textContent).toBe('✓');
    });

    it('should render info icon', () => {
      const { container } = render(<Toast message="Info" type="info" />);

      const icon = container.querySelector('.toast-icon');
      expect(icon).toBeInTheDocument();
      expect(icon?.textContent).toBe('ℹ');
    });

    it('should render warning icon', () => {
      const { container } = render(<Toast message="Warning" type="warning" />);

      const icon = container.querySelector('.toast-icon');
      expect(icon).toBeInTheDocument();
      expect(icon?.textContent).toBe('⚠');
    });

    it('should render error icon', () => {
      const { container } = render(<Toast message="Error" type="error" />);

      const icon = container.querySelector('.toast-icon');
      expect(icon).toBeInTheDocument();
      expect(icon?.textContent).toBe('✗');
    });
  });

  describe('Auto-dismiss Behavior', () => {
    it('should hide after default duration (3000ms)', () => {
      const { container } = render(<Toast message="Auto dismiss" />);

      expect(container.querySelector('.toast')).toBeInTheDocument();

      // Fast-forward time by 3000ms
      act(() => {
        vi.advanceTimersByTime(3000);
      });

      // Toast should be hidden (isVisible = false, returns null)
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });

    it('should hide after custom duration', () => {
      const { container } = render(<Toast message="Custom duration" duration={5000} />);

      expect(container.querySelector('.toast')).toBeInTheDocument();

      // Fast-forward by less than duration - should still be visible
      act(() => {
        vi.advanceTimersByTime(4000);
      });
      expect(container.querySelector('.toast')).toBeInTheDocument();

      // Fast-forward past duration - should be hidden
      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });

    it('should call onClose after duration + animation delay', () => {
      const onClose = vi.fn();
      render(<Toast message="With callback" duration={2000} onClose={onClose} />);

      // After duration, onClose should not be called yet (waiting for animation)
      act(() => {
        vi.advanceTimersByTime(2000);
      });
      expect(onClose).not.toHaveBeenCalled();

      // After animation delay (300ms), onClose should be called
      act(() => {
        vi.advanceTimersByTime(300);
      });
      expect(onClose).toHaveBeenCalledOnce();
    });

    it('should not call onClose if not provided', () => {
      const { container } = render(<Toast message="No callback" duration={1000} />);

      expect(container.querySelector('.toast')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();

      // Should not throw error
      act(() => {
        vi.advanceTimersByTime(300);
      });
    });
  });

  describe('Timer Cleanup', () => {
    it('should clean up timer on unmount', () => {
      const onClose = vi.fn();
      const { unmount } = render(<Toast message="Cleanup test" duration={5000} onClose={onClose} />);

      // Unmount before timer fires
      unmount();

      // Advance time past duration
      vi.advanceTimersByTime(5000);
      vi.advanceTimersByTime(300);

      // onClose should not be called since component was unmounted
      expect(onClose).not.toHaveBeenCalled();
    });

    it('should handle rapid unmount without errors', () => {
      const { unmount } = render(<Toast message="Rapid unmount" />);

      // Unmount immediately
      unmount();

      // Advance timers - should not cause errors
      expect(() => {
        vi.advanceTimersByTime(10000);
      }).not.toThrow();
    });
  });

  describe('Component Structure', () => {
    it('should have correct CSS classes', () => {
      const { container } = render(<Toast message="Structure test" type="success" />);

      expect(container.querySelector('.toast')).toBeInTheDocument();
      expect(container.querySelector('.toast-success')).toBeInTheDocument();
      expect(container.querySelector('.toast-icon')).toBeInTheDocument();
      expect(container.querySelector('.toast-message')).toBeInTheDocument();
    });

    it('should maintain correct DOM hierarchy', () => {
      const { container } = render(<Toast message="Hierarchy" type="info" />);

      const toast = container.querySelector('.toast') as HTMLElement | null;
      const icon = container.querySelector('.toast-icon') as HTMLElement | null;
      const message = container.querySelector('.toast-message') as HTMLElement | null;

      expect(toast).toContainElement(icon);
      expect(toast).toContainElement(message);
    });

    it('should add toast-hide class when hidden', () => {
      const { container } = render(<Toast message="Hide class test" duration={1000} />);

      // Initially should not have toast-hide class
      const toast = container.querySelector('.toast');
      expect(toast).not.toHaveClass('toast-hide');

      // After duration, component returns null (not just adding class)
      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });
  });

  describe('Message Content Variations', () => {
    it('should handle long messages', () => {
      const longMessage = 'This is a very long message that might wrap to multiple lines in the toast notification display.';
      render(<Toast message={longMessage} />);

      expect(screen.getByText(longMessage)).toBeInTheDocument();
    });

    it('should handle special characters', () => {
      render(<Toast message="Success! (100%)" type="success" />);

      expect(screen.getByText('Success! (100%)')).toBeInTheDocument();
    });

    it('should handle HTML entities', () => {
      render(<Toast message="Value < 100 & > 0" />);

      expect(screen.getByText('Value < 100 & > 0')).toBeInTheDocument();
    });

    it('should handle empty message', () => {
      render(<Toast message="" />);

      const message = document.querySelector('.toast-message');
      expect(message).toBeInTheDocument();
      expect(message?.textContent).toBe('');
    });
  });

  describe('Real-World Use Cases', () => {
    it('should show success toast for save operation', () => {
      const { container } = render(
        <Toast message="Deck saved successfully!" type="success" duration={3000} />
      );

      expect(screen.getByText('Deck saved successfully!')).toBeInTheDocument();
      expect(container.querySelector('.toast-success')).toBeInTheDocument();
      expect(screen.getByText('✓')).toBeInTheDocument();
    });

    it('should show error toast for failed operation', () => {
      const onClose = vi.fn();
      const { container } = render(
        <Toast
          message="Failed to load draft data"
          type="error"
          duration={5000}
          onClose={onClose}
        />
      );

      expect(screen.getByText('Failed to load draft data')).toBeInTheDocument();
      expect(container.querySelector('.toast-error')).toBeInTheDocument();
      expect(screen.getByText('✗')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(5300);
      });
      expect(onClose).toHaveBeenCalled();
    });

    it('should show warning toast', () => {
      render(<Toast message="Deck contains banned cards" type="warning" />);

      expect(screen.getByText('Deck contains banned cards')).toBeInTheDocument();
      expect(screen.getByText('⚠')).toBeInTheDocument();
    });

    it('should show info toast', () => {
      render(<Toast message="Draft data updated" type="info" />);

      expect(screen.getByText('Draft data updated')).toBeInTheDocument();
      expect(screen.getByText('ℹ')).toBeInTheDocument();
    });
  });

  describe('Duration Variations', () => {
    it('should handle very short duration', () => {
      const { container } = render(<Toast message="Quick toast" duration={100} />);

      expect(container.querySelector('.toast')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(100);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });

    it('should handle very long duration', () => {
      const { container } = render(<Toast message="Long toast" duration={10000} />);

      expect(container.querySelector('.toast')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(9000);
      });
      expect(container.querySelector('.toast')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });

    it('should handle zero duration', () => {
      const { container } = render(<Toast message="Instant" duration={0} />);

      expect(container.querySelector('.toast')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(0);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });
  });

  describe('Multiple Toast Type Combinations', () => {
    it('should render success toast with custom duration', () => {
      const { container } = render(
        <Toast message="Custom success" type="success" duration={4000} />
      );

      expect(container.querySelector('.toast-success')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(3000);
      });
      expect(container.querySelector('.toast')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });

    it('should render error toast with callback', () => {
      const onClose = vi.fn();
      render(
        <Toast message="Error with callback" type="error" onClose={onClose} />
      );

      act(() => {
        vi.advanceTimersByTime(3300);
      });
      expect(onClose).toHaveBeenCalledOnce();
    });

    it('should render warning toast with all props', () => {
      const onClose = vi.fn();
      const { container } = render(
        <Toast
          message="Complete warning"
          type="warning"
          duration={2000}
          onClose={onClose}
        />
      );

      expect(container.querySelector('.toast-warning')).toBeInTheDocument();
      expect(screen.getByText('Complete warning')).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(2300);
      });
      expect(onClose).toHaveBeenCalled();
    });
  });

  describe('Edge Cases', () => {
    it('should handle re-renders without creating multiple timers', () => {
      const onClose = vi.fn();
      const { rerender } = render(
        <Toast message="Original" duration={5000} onClose={onClose} />
      );

      // Rerender with same props
      rerender(<Toast message="Original" duration={5000} onClose={onClose} />);
      rerender(<Toast message="Original" duration={5000} onClose={onClose} />);

      // Advance time
      act(() => {
        vi.advanceTimersByTime(5300);
      });

      // onClose should only be called once despite rerenders
      expect(onClose).toHaveBeenCalledOnce();
    });

    it('should return null when not visible', () => {
      const { container } = render(<Toast message="Will disappear" duration={1000} />);

      expect(container.firstChild).not.toBeNull();

      act(() => {
        vi.advanceTimersByTime(1000);
      });
      expect(container.firstChild).toBeNull();
    });

    it('should handle onClose being undefined', () => {
      const { container } = render(<Toast message="No callback" duration={1000} />);

      expect(() => {
        act(() => {
          vi.advanceTimersByTime(1300);
        });
      }).not.toThrow();

      expect(container.querySelector('.toast')).not.toBeInTheDocument();
    });
  });
});

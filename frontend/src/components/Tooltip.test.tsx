import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import Tooltip from './Tooltip';

describe('Tooltip', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  describe('Basic Rendering', () => {
    it('should render children', () => {
      render(
        <Tooltip content="Tooltip text">
          <button>Hover me</button>
        </Tooltip>
      );

      expect(screen.getByText('Hover me')).toBeInTheDocument();
    });

    it('should not show tooltip initially', () => {
      const { container } = render(
        <Tooltip content="Hidden tooltip">
          <span>Target</span>
        </Tooltip>
      );

      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();
    });

    it('should render container element', () => {
      const { container } = render(
        <Tooltip content="Test">
          <div>Child</div>
        </Tooltip>
      );

      expect(container.querySelector('.tooltip-container')).toBeInTheDocument();
    });
  });

  describe('Mouse Interaction', () => {
    it('should show tooltip on mouse enter after delay', () => {
      const { container } = render(
        <Tooltip content="Mouse tooltip" delay={300}>
          <button>Hover</button>
        </Tooltip>
      );

      const button = screen.getByText('Hover');
      fireEvent.mouseEnter(button.parentElement!);

      // Tooltip should not be visible immediately
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();

      // Fast-forward past delay
      act(() => {
        vi.advanceTimersByTime(300);
      });

      // Tooltip should now be visible
      expect(screen.getByText('Mouse tooltip')).toBeInTheDocument();
    });

    it('should hide tooltip on mouse leave', () => {
      const { container } = render(
        <Tooltip content="Hide tooltip">
          <button>Hover</button>
        </Tooltip>
      );

      const button = screen.getByText('Hover');

      // Show tooltip
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });
      expect(screen.getByText('Hide tooltip')).toBeInTheDocument();

      // Hide tooltip
      fireEvent.mouseLeave(button.parentElement!);
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();
    });

    it('should cancel tooltip if mouse leaves before delay', () => {
      const { container } = render(
        <Tooltip content="Canceled tooltip" delay={500}>
          <button>Quick hover</button>
        </Tooltip>
      );

      const button = screen.getByText('Quick hover');

      // Start hovering
      fireEvent.mouseEnter(button.parentElement!);

      // Leave before delay completes
      act(() => {
        vi.advanceTimersByTime(200);
      });
      fireEvent.mouseLeave(button.parentElement!);

      // Complete the original delay
      act(() => {
        vi.advanceTimersByTime(300);
      });

      // Tooltip should not appear
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();
    });
  });

  describe('Keyboard Interaction', () => {
    it('should show tooltip on focus after delay', () => {
      render(
        <Tooltip content="Focus tooltip">
          <button>Focusable</button>
        </Tooltip>
      );

      const button = screen.getByText('Focusable');
      fireEvent.focus(button.parentElement!);

      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('Focus tooltip')).toBeInTheDocument();
    });

    it('should hide tooltip on blur', () => {
      const { container } = render(
        <Tooltip content="Blur tooltip">
          <button>Focusable</button>
        </Tooltip>
      );

      const button = screen.getByText('Focusable');

      // Show tooltip
      fireEvent.focus(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });
      expect(screen.getByText('Blur tooltip')).toBeInTheDocument();

      // Hide tooltip
      fireEvent.blur(button.parentElement!);
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();
    });

    it('should cancel tooltip if blur occurs before delay', () => {
      const { container } = render(
        <Tooltip content="Quick focus" delay={400}>
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');

      // Focus
      fireEvent.focus(button.parentElement!);

      // Blur before delay
      act(() => {
        vi.advanceTimersByTime(200);
      });
      fireEvent.blur(button.parentElement!);

      // Complete delay
      act(() => {
        vi.advanceTimersByTime(200);
      });

      // Tooltip should not appear
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();
    });
  });

  describe('Position Variants', () => {
    it('should render with top position by default', () => {
      render(
        <Tooltip content="Top tooltip">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      const tooltip = document.querySelector('.tooltip-top');
      expect(tooltip).toBeInTheDocument();
    });

    it('should render with bottom position', () => {
      render(
        <Tooltip content="Bottom tooltip" position="bottom">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      const tooltip = document.querySelector('.tooltip-bottom');
      expect(tooltip).toBeInTheDocument();
    });

    it('should render with left position', () => {
      render(
        <Tooltip content="Left tooltip" position="left">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      const tooltip = document.querySelector('.tooltip-left');
      expect(tooltip).toBeInTheDocument();
    });

    it('should render with right position', () => {
      render(
        <Tooltip content="Right tooltip" position="right">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      const tooltip = document.querySelector('.tooltip-right');
      expect(tooltip).toBeInTheDocument();
    });
  });

  describe('Delay Behavior', () => {
    it('should use default delay of 300ms', () => {
      const { container } = render(
        <Tooltip content="Default delay">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);

      // Should not be visible at 250ms
      act(() => {
        vi.advanceTimersByTime(250);
      });
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();

      // Should be visible at 300ms
      act(() => {
        vi.advanceTimersByTime(50);
      });
      expect(screen.getByText('Default delay')).toBeInTheDocument();
    });

    it('should use custom delay', () => {
      const { container } = render(
        <Tooltip content="Custom delay" delay={1000}>
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);

      // Should not be visible at 900ms
      act(() => {
        vi.advanceTimersByTime(900);
      });
      expect(container.querySelector('.tooltip-content')).not.toBeInTheDocument();

      // Should be visible at 1000ms
      act(() => {
        vi.advanceTimersByTime(100);
      });
      expect(screen.getByText('Custom delay')).toBeInTheDocument();
    });

    it('should handle zero delay', () => {
      render(
        <Tooltip content="Instant tooltip" delay={0}>
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);

      act(() => {
        vi.advanceTimersByTime(0);
      });

      expect(screen.getByText('Instant tooltip')).toBeInTheDocument();
    });
  });

  describe('Tooltip Content', () => {
    it('should render tooltip content text', () => {
      render(
        <Tooltip content="This is tooltip content">
          <button>Hover</button>
        </Tooltip>
      );

      const button = screen.getByText('Hover');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('This is tooltip content')).toBeInTheDocument();
    });

    it('should render arrow element', () => {
      const { container } = render(
        <Tooltip content="With arrow">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      const arrow = container.querySelector('.tooltip-arrow');
      expect(arrow).toBeInTheDocument();
    });

    it('should handle long content', () => {
      const longContent = 'This is a very long tooltip content that might wrap to multiple lines in the display';
      render(
        <Tooltip content={longContent}>
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText(longContent)).toBeInTheDocument();
    });

    it('should handle special characters', () => {
      render(
        <Tooltip content="Special: < > & 100%">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('Special: < > & 100%')).toBeInTheDocument();
    });
  });

  describe('Children Rendering', () => {
    it('should render button children', () => {
      render(
        <Tooltip content="Tooltip">
          <button type="button">Click me</button>
        </Tooltip>
      );

      const button = screen.getByRole('button', { name: 'Click me' });
      expect(button).toBeInTheDocument();
    });

    it('should render text children', () => {
      render(
        <Tooltip content="Info">
          <span>Hover text</span>
        </Tooltip>
      );

      expect(screen.getByText('Hover text')).toBeInTheDocument();
    });

    it('should render complex children', () => {
      render(
        <Tooltip content="Complex">
          <div>
            <span>Label:</span>
            <strong>Value</strong>
          </div>
        </Tooltip>
      );

      expect(screen.getByText('Label:')).toBeInTheDocument();
      expect(screen.getByText('Value')).toBeInTheDocument();
    });

    it('should render multiple children', () => {
      render(
        <Tooltip content="Multi">
          <>
            <span>First</span>
            <span>Second</span>
          </>
        </Tooltip>
      );

      expect(screen.getByText('First')).toBeInTheDocument();
      expect(screen.getByText('Second')).toBeInTheDocument();
    });
  });

  describe('Component Structure', () => {
    it('should have correct CSS classes', () => {
      const { container } = render(
        <Tooltip content="Test" position="bottom">
          <button>Button</button>
        </Tooltip>
      );

      expect(container.querySelector('.tooltip-container')).toBeInTheDocument();

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(container.querySelector('.tooltip-content')).toBeInTheDocument();
      expect(container.querySelector('.tooltip-bottom')).toBeInTheDocument();
      expect(container.querySelector('.tooltip-arrow')).toBeInTheDocument();
    });

    it('should maintain correct DOM hierarchy', () => {
      const { container } = render(
        <Tooltip content="Hierarchy test">
          <button>Button</button>
        </Tooltip>
      );

      const tooltipContainer = container.querySelector('.tooltip-container');
      const button = screen.getByText('Button');

      expect(tooltipContainer).toContainElement(button);

      fireEvent.mouseEnter(tooltipContainer!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      const tooltipContent = container.querySelector('.tooltip-content') as HTMLElement | null;
      expect(tooltipContainer).toContainElement(tooltipContent);
    });
  });

  describe('Real-World Use Cases', () => {
    it('should work with icon buttons', () => {
      render(
        <Tooltip content="Save changes">
          <button aria-label="Save">💾</button>
        </Tooltip>
      );

      const button = screen.getByLabelText('Save');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('Save changes')).toBeInTheDocument();
    });

    it('should work with disabled buttons', () => {
      render(
        <Tooltip content="Action unavailable">
          <button disabled>Disabled</button>
        </Tooltip>
      );

      const button = screen.getByText('Disabled');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('Action unavailable')).toBeInTheDocument();
    });

    it('should work with links', () => {
      render(
        <Tooltip content="Visit documentation">
          <a href="/docs">Docs</a>
        </Tooltip>
      );

      const link = screen.getByText('Docs');
      fireEvent.mouseEnter(link.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('Visit documentation')).toBeInTheDocument();
    });

    it('should work with form inputs', () => {
      render(
        <Tooltip content="Enter your email" position="right">
          <input type="email" placeholder="Email" />
        </Tooltip>
      );

      const input = screen.getByPlaceholderText('Email');
      fireEvent.focus(input.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      expect(screen.getByText('Enter your email')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid hover events', () => {
      render(
        <Tooltip content="Rapid hover">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');

      // Rapidly hover in and out
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(100);
      });
      fireEvent.mouseLeave(button.parentElement!);

      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(100);
      });
      fireEvent.mouseLeave(button.parentElement!);

      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      // Only the last hover should show the tooltip
      expect(screen.getByText('Rapid hover')).toBeInTheDocument();
    });

    it('should handle unmount during delay', () => {
      const { unmount } = render(
        <Tooltip content="Unmounting">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);

      // Unmount before delay completes
      unmount();

      // Should not cause errors
      expect(() => {
        act(() => {
          vi.advanceTimersByTime(300);
        });
      }).not.toThrow();
    });

    it('should handle empty content gracefully', () => {
      render(
        <Tooltip content="">
          <button>Button</button>
        </Tooltip>
      );

      const button = screen.getByText('Button');
      fireEvent.mouseEnter(button.parentElement!);
      act(() => {
        vi.advanceTimersByTime(300);
      });

      // Tooltip should render even with empty content
      const tooltipContent = document.querySelector('.tooltip-content');
      expect(tooltipContent).toBeInTheDocument();
      expect(tooltipContent?.textContent).toBe('');
    });
  });

  describe('Position and Delay Combinations', () => {
    it('should work with all position and delay combinations', () => {
      const positions: Array<'top' | 'bottom' | 'left' | 'right'> = ['top', 'bottom', 'left', 'right'];
      const delays = [0, 200, 500];

      positions.forEach(position => {
        delays.forEach(delay => {
          const { container, unmount } = render(
            <Tooltip content={`${position} ${delay}`} position={position} delay={delay}>
              <button>Button</button>
            </Tooltip>
          );

          const button = screen.getByText('Button');
          fireEvent.mouseEnter(button.parentElement!);
          act(() => {
            vi.advanceTimersByTime(delay);
          });

          expect(screen.getByText(`${position} ${delay}`)).toBeInTheDocument();
          expect(container.querySelector(`.tooltip-${position}`)).toBeInTheDocument();

          unmount();
        });
      });
    });
  });
});

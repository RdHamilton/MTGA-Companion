import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import LoadingButton from './LoadingButton';

describe('LoadingButton', () => {
  const defaultProps = {
    loading: false,
    loadingText: 'Loading...',
    onClick: vi.fn(),
    children: 'Click Me',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Default State', () => {
    it('should render children when not loading', () => {
      render(<LoadingButton {...defaultProps} />);

      expect(screen.getByText('Click Me')).toBeInTheDocument();
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('should not show spinner when not loading', () => {
      render(<LoadingButton {...defaultProps} />);

      expect(document.querySelector('.loading-button-spinner')).not.toBeInTheDocument();
    });

    it('should call onClick when clicked', () => {
      render(<LoadingButton {...defaultProps} />);

      fireEvent.click(screen.getByRole('button'));

      expect(defaultProps.onClick).toHaveBeenCalledTimes(1);
    });

    it('should not be disabled by default', () => {
      render(<LoadingButton {...defaultProps} />);

      expect(screen.getByRole('button')).not.toBeDisabled();
    });
  });

  describe('Loading State', () => {
    it('should show loading text when loading', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
      expect(screen.queryByText('Click Me')).not.toBeInTheDocument();
    });

    it('should show spinner when loading', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      expect(document.querySelector('.loading-button-spinner')).toBeInTheDocument();
    });

    it('should be disabled when loading', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      expect(screen.getByRole('button')).toBeDisabled();
    });

    it('should not call onClick when loading and clicked', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      fireEvent.click(screen.getByRole('button'));

      expect(defaultProps.onClick).not.toHaveBeenCalled();
    });

    it('should have loading class when loading', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      expect(screen.getByRole('button')).toHaveClass('loading');
    });
  });

  describe('Disabled State', () => {
    it('should be disabled when disabled prop is true', () => {
      render(<LoadingButton {...defaultProps} disabled={true} />);

      expect(screen.getByRole('button')).toBeDisabled();
    });

    it('should not call onClick when disabled and clicked', () => {
      render(<LoadingButton {...defaultProps} disabled={true} />);

      fireEvent.click(screen.getByRole('button'));

      expect(defaultProps.onClick).not.toHaveBeenCalled();
    });

    it('should be disabled when both loading and disabled are true', () => {
      render(<LoadingButton {...defaultProps} loading={true} disabled={true} />);

      expect(screen.getByRole('button')).toBeDisabled();
    });
  });

  describe('Variants', () => {
    it('should apply primary variant class', () => {
      render(<LoadingButton {...defaultProps} variant="primary" />);

      expect(screen.getByRole('button')).toHaveClass('primary');
    });

    it('should apply danger variant class', () => {
      render(<LoadingButton {...defaultProps} variant="danger" />);

      expect(screen.getByRole('button')).toHaveClass('danger');
    });

    it('should apply pause variant class', () => {
      render(<LoadingButton {...defaultProps} variant="pause" />);

      expect(screen.getByRole('button')).toHaveClass('pause');
    });

    it('should apply resume variant class', () => {
      render(<LoadingButton {...defaultProps} variant="resume" />);

      expect(screen.getByRole('button')).toHaveClass('resume');
    });

    it('should apply recalculate variant class', () => {
      render(<LoadingButton {...defaultProps} variant="recalculate" />);

      expect(screen.getByRole('button')).toHaveClass('recalculate');
    });

    it('should apply clear-cache variant class', () => {
      render(<LoadingButton {...defaultProps} variant="clear-cache" />);

      expect(screen.getByRole('button')).toHaveClass('clear-cache');
    });

    it('should not apply variant class for default variant', () => {
      render(<LoadingButton {...defaultProps} variant="default" />);

      const button = screen.getByRole('button');
      expect(button).toHaveClass('action-button');
      expect(button).not.toHaveClass('default');
    });
  });

  describe('Custom Class Name', () => {
    it('should apply custom className', () => {
      render(<LoadingButton {...defaultProps} className="custom-class" />);

      expect(screen.getByRole('button')).toHaveClass('custom-class');
    });

    it('should combine custom className with variant', () => {
      render(<LoadingButton {...defaultProps} variant="primary" className="custom-class" />);

      const button = screen.getByRole('button');
      expect(button).toHaveClass('primary');
      expect(button).toHaveClass('custom-class');
    });
  });

  describe('Base Classes', () => {
    it('should always have loading-button class', () => {
      render(<LoadingButton {...defaultProps} />);

      expect(screen.getByRole('button')).toHaveClass('loading-button');
    });

    it('should always have action-button class', () => {
      render(<LoadingButton {...defaultProps} />);

      expect(screen.getByRole('button')).toHaveClass('action-button');
    });
  });

  describe('Spinner Animation', () => {
    it('should have spinner with correct class', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      const spinner = document.querySelector('.loading-button-spinner');
      expect(spinner).toBeInTheDocument();
      expect(spinner).toHaveClass('loading-button-spinner');
    });
  });

  describe('Accessibility', () => {
    it('should be keyboard accessible when not disabled', () => {
      render(<LoadingButton {...defaultProps} />);

      const button = screen.getByRole('button');
      button.focus();

      expect(document.activeElement).toBe(button);
    });

    it('should announce loading state to screen readers via disabled attribute', () => {
      render(<LoadingButton {...defaultProps} loading={true} />);

      // When loading, button is disabled which indicates to assistive tech
      expect(screen.getByRole('button')).toBeDisabled();
    });
  });

  describe('Children Types', () => {
    it('should render string children', () => {
      render(<LoadingButton {...defaultProps}>Submit</LoadingButton>);

      expect(screen.getByText('Submit')).toBeInTheDocument();
    });

    it('should render JSX children', () => {
      render(
        <LoadingButton {...defaultProps}>
          <span data-testid="custom-child">Custom Content</span>
        </LoadingButton>
      );

      expect(screen.getByTestId('custom-child')).toBeInTheDocument();
    });
  });
});

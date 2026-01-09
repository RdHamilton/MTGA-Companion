import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ProgressBar from './ProgressBar';

describe('ProgressBar', () => {
  describe('rendering', () => {
    it('renders with default props', () => {
      render(<ProgressBar progress={50} />);
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays the label when provided', () => {
      render(<ProgressBar progress={50} label="Loading data..." />);
      expect(screen.getByText('Loading data...')).toBeInTheDocument();
    });

    it('displays percentage by default', () => {
      render(<ProgressBar progress={75} />);
      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    it('hides percentage when showPercentage is false', () => {
      render(<ProgressBar progress={75} showPercentage={false} />);
      expect(screen.queryByText('75%')).not.toBeInTheDocument();
    });

    it('displays detail text when provided', () => {
      render(<ProgressBar progress={50} detail="Processing 10 of 20 items..." />);
      expect(screen.getByText('Processing 10 of 20 items...')).toBeInTheDocument();
    });

    it('displays estimated time remaining when provided', () => {
      render(<ProgressBar progress={50} estimatedTimeRemaining={30000} />);
      expect(screen.getByText('~30s remaining')).toBeInTheDocument();
    });

    it('formats time remaining in minutes when >= 60 seconds', () => {
      render(<ProgressBar progress={50} estimatedTimeRemaining={120000} />);
      expect(screen.getByText('~2m remaining')).toBeInTheDocument();
    });

    it('formats time remaining as "Less than a second" when < 1000ms', () => {
      render(<ProgressBar progress={50} estimatedTimeRemaining={500} />);
      expect(screen.getByText('Less than a second')).toBeInTheDocument();
    });
  });

  describe('progress bar fill', () => {
    it('clamps progress to 0-100 range', () => {
      const { rerender } = render(<ProgressBar progress={-10} />);
      const progressbar = screen.getByRole('progressbar');
      expect(progressbar).toHaveAttribute('aria-valuenow', '0');

      rerender(<ProgressBar progress={150} />);
      expect(progressbar).toHaveAttribute('aria-valuenow', '100');
    });

    it('applies correct progress width style', () => {
      render(<ProgressBar progress={60} />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveStyle({ width: '60%' });
    });
  });

  describe('size variants', () => {
    it('applies small size class', () => {
      render(<ProgressBar progress={50} size="small" />);
      const container = document.querySelector('.progress-bar-container');
      expect(container).toHaveClass('progress-bar-small');
    });

    it('applies medium size class by default', () => {
      render(<ProgressBar progress={50} />);
      const container = document.querySelector('.progress-bar-container');
      expect(container).toHaveClass('progress-bar-medium');
    });

    it('applies large size class', () => {
      render(<ProgressBar progress={50} size="large" />);
      const container = document.querySelector('.progress-bar-container');
      expect(container).toHaveClass('progress-bar-large');
    });
  });

  describe('color variants', () => {
    it('applies primary variant by default', () => {
      render(<ProgressBar progress={50} />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('progress-bar-primary');
    });

    it('applies success variant', () => {
      render(<ProgressBar progress={50} variant="success" />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('progress-bar-success');
    });

    it('applies warning variant', () => {
      render(<ProgressBar progress={50} variant="warning" />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('progress-bar-warning');
    });

    it('applies error variant', () => {
      render(<ProgressBar progress={50} variant="error" />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('progress-bar-error');
    });
  });

  describe('indeterminate mode', () => {
    it('applies indeterminate class when enabled', () => {
      render(<ProgressBar progress={50} indeterminate />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('indeterminate');
    });

    it('hides percentage in indeterminate mode', () => {
      render(<ProgressBar progress={50} indeterminate showPercentage />);
      expect(screen.queryByText('50%')).not.toBeInTheDocument();
    });

    it('removes aria-valuenow in indeterminate mode', () => {
      render(<ProgressBar progress={50} indeterminate />);
      const progressbar = screen.getByRole('progressbar');
      expect(progressbar).not.toHaveAttribute('aria-valuenow');
    });
  });

  describe('cancel button', () => {
    it('does not show cancel button by default', () => {
      render(<ProgressBar progress={50} />);
      expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
    });

    it('shows cancel button when showCancel is true and onCancel is provided', () => {
      const onCancel = vi.fn();
      render(<ProgressBar progress={50} showCancel onCancel={onCancel} />);
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
    });

    it('calls onCancel when cancel button is clicked', () => {
      const onCancel = vi.fn();
      render(<ProgressBar progress={50} showCancel onCancel={onCancel} />);
      fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
      expect(onCancel).toHaveBeenCalledTimes(1);
    });

    it('does not show cancel button if showCancel is true but onCancel is not provided', () => {
      render(<ProgressBar progress={50} showCancel />);
      expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('has correct ARIA attributes', () => {
      render(<ProgressBar progress={50} label="Processing" />);
      const progressbar = screen.getByRole('progressbar');
      expect(progressbar).toHaveAttribute('aria-valuemin', '0');
      expect(progressbar).toHaveAttribute('aria-valuemax', '100');
      expect(progressbar).toHaveAttribute('aria-valuenow', '50');
      expect(progressbar).toHaveAttribute('aria-label', 'Processing');
    });

    it('uses default aria-label when no label provided', () => {
      render(<ProgressBar progress={50} />);
      const progressbar = screen.getByRole('progressbar');
      expect(progressbar).toHaveAttribute('aria-label', 'Progress');
    });
  });

  describe('custom className', () => {
    it('applies custom className', () => {
      render(<ProgressBar progress={50} className="custom-class" />);
      const container = document.querySelector('.progress-bar-container');
      expect(container).toHaveClass('custom-class');
    });
  });
});

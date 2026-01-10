import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ProgressModal from './ProgressModal';

describe('ProgressModal', () => {
  describe('visibility', () => {
    it('does not render when isOpen is false', () => {
      render(<ProgressModal isOpen={false} title="Test" progress={50} />);
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('renders when isOpen is true', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={50} />);
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
  });

  describe('content', () => {
    it('displays the title', () => {
      render(<ProgressModal isOpen={true} title="Generating Deck" progress={50} />);
      expect(screen.getByText('Generating Deck')).toBeInTheDocument();
    });

    it('displays the icon when provided', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={50} icon="🔥" />);
      expect(screen.getByText('🔥')).toBeInTheDocument();
    });

    it('displays detail text', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={50} detail="Processing files..." />);
      expect(screen.getByText('Processing files...')).toBeInTheDocument();
    });

    it('displays estimated time remaining', () => {
      render(
        <ProgressModal isOpen={true} title="Test" progress={50} estimatedTimeRemaining={45000} />
      );
      expect(screen.getByText('~45s remaining')).toBeInTheDocument();
    });
  });

  describe('progress', () => {
    it('passes progress to ProgressBar', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={75} />);
      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    it('hides percentage in indeterminate mode', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={75} indeterminate />);
      expect(screen.queryByText('75%')).not.toBeInTheDocument();
    });
  });

  describe('cancel functionality', () => {
    it('does not show cancel button by default', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={50} />);
      expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
    });

    it('shows cancel button when cancellable is true and onCancel is provided', () => {
      const onCancel = vi.fn();
      render(
        <ProgressModal isOpen={true} title="Test" progress={50} cancellable onCancel={onCancel} />
      );
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
    });

    it('calls onCancel when cancel button is clicked', () => {
      const onCancel = vi.fn();
      render(
        <ProgressModal isOpen={true} title="Test" progress={50} cancellable onCancel={onCancel} />
      );
      fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
      expect(onCancel).toHaveBeenCalledTimes(1);
    });
  });

  describe('keyboard handling', () => {
    it('calls onCancel when Escape is pressed and cancellable', () => {
      const onCancel = vi.fn();
      render(
        <ProgressModal isOpen={true} title="Test" progress={50} cancellable onCancel={onCancel} />
      );

      fireEvent.keyDown(document, { key: 'Escape' });
      expect(onCancel).toHaveBeenCalledTimes(1);
    });

    it('does not call onCancel when Escape is pressed and not cancellable', () => {
      const onCancel = vi.fn();
      render(
        <ProgressModal isOpen={true} title="Test" progress={50} cancellable={false} onCancel={onCancel} />
      );

      fireEvent.keyDown(document, { key: 'Escape' });
      expect(onCancel).not.toHaveBeenCalled();
    });
  });

  describe('body scroll lock', () => {
    it('locks body scroll when modal opens', () => {
      const { unmount } = render(<ProgressModal isOpen={true} title="Test" progress={50} />);
      expect(document.body.style.overflow).toBe('hidden');
      unmount();
    });

    it('restores body scroll when modal closes', () => {
      const { unmount } = render(<ProgressModal isOpen={true} title="Test" progress={50} />);
      expect(document.body.style.overflow).toBe('hidden');
      unmount();
      expect(document.body.style.overflow).toBe('');
    });
  });

  describe('accessibility', () => {
    it('has correct dialog ARIA attributes', () => {
      render(<ProgressModal isOpen={true} title="Test Title" progress={50} />);
      const dialog = screen.getByRole('dialog');
      expect(dialog).toHaveAttribute('aria-modal', 'true');
      expect(dialog).toHaveAttribute('aria-labelledby', 'progress-modal-title');
    });

    it('associates title with aria-labelledby', () => {
      render(<ProgressModal isOpen={true} title="Test Title" progress={50} />);
      const title = screen.getByText('Test Title');
      expect(title).toHaveAttribute('id', 'progress-modal-title');
    });
  });

  describe('variants', () => {
    it('passes variant to ProgressBar', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={50} variant="success" />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('progress-bar-success');
    });

    it('applies error variant', () => {
      render(<ProgressModal isOpen={true} title="Test" progress={50} variant="error" />);
      const fill = document.querySelector('.progress-bar-fill');
      expect(fill).toHaveClass('progress-bar-error');
    });
  });
});

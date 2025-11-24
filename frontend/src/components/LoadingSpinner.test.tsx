import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LoadingSpinner from './LoadingSpinner';

describe('LoadingSpinner', () => {
  describe('Rendering', () => {
    it('should render with default message', () => {
      render(<LoadingSpinner />);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
      expect(document.querySelector('.spinner')).toBeInTheDocument();
    });

    it('should render with custom message', () => {
      render(<LoadingSpinner message="Fetching draft data..." />);

      expect(screen.getByText('Fetching draft data...')).toBeInTheDocument();
    });

    it('should not render message when message is empty string', () => {
      render(<LoadingSpinner message="" />);

      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
      expect(document.querySelector('.loading-message')).not.toBeInTheDocument();
    });

    it('should render spinner element', () => {
      render(<LoadingSpinner />);

      const spinner = document.querySelector('.spinner');
      expect(spinner).toBeInTheDocument();

      const spinnerCircle = document.querySelector('.spinner-circle');
      expect(spinnerCircle).toBeInTheDocument();
    });
  });

  describe('Size Variants', () => {
    it('should render with small size', () => {
      render(<LoadingSpinner size="small" />);

      const spinner = document.querySelector('.spinner-small');
      expect(spinner).toBeInTheDocument();
    });

    it('should render with medium size by default', () => {
      render(<LoadingSpinner />);

      const spinner = document.querySelector('.spinner-medium');
      expect(spinner).toBeInTheDocument();
    });

    it('should render with large size', () => {
      render(<LoadingSpinner size="large" />);

      const spinner = document.querySelector('.spinner-large');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('Component Structure', () => {
    it('should render loading container', () => {
      render(<LoadingSpinner />);

      const container = document.querySelector('.loading-container');
      expect(container).toBeInTheDocument();
    });

    it('should render message in paragraph element', () => {
      render(<LoadingSpinner message="Please wait..." />);

      const message = screen.getByText('Please wait...');
      expect(message.tagName).toBe('P');
      expect(message).toHaveClass('loading-message');
    });

    it('should maintain correct DOM hierarchy', () => {
      render(<LoadingSpinner message="Loading data..." />);

      const container = document.querySelector('.loading-container') as HTMLElement | null;
      const spinner = document.querySelector('.spinner') as HTMLElement | null;
      const message = screen.getByText('Loading data...');

      expect(container).toContainElement(spinner);
      expect(container).toContainElement(message);
    });
  });

  describe('Props Combinations', () => {
    it('should handle small size with custom message', () => {
      render(<LoadingSpinner size="small" message="Quick load..." />);

      expect(screen.getByText('Quick load...')).toBeInTheDocument();
      expect(document.querySelector('.spinner-small')).toBeInTheDocument();
    });

    it('should handle large size with custom message', () => {
      render(<LoadingSpinner size="large" message="Processing large dataset..." />);

      expect(screen.getByText('Processing large dataset...')).toBeInTheDocument();
      expect(document.querySelector('.spinner-large')).toBeInTheDocument();
    });

    it('should handle empty message with different sizes', () => {
      const { rerender } = render(<LoadingSpinner size="small" message="" />);
      expect(document.querySelector('.spinner-small')).toBeInTheDocument();
      expect(document.querySelector('.loading-message')).not.toBeInTheDocument();

      rerender(<LoadingSpinner size="large" message="" />);
      expect(document.querySelector('.spinner-large')).toBeInTheDocument();
      expect(document.querySelector('.loading-message')).not.toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle very long message text', () => {
      const longMessage = 'Loading a very long message that might wrap across multiple lines in the UI and should still render correctly without breaking the layout';
      render(<LoadingSpinner message={longMessage} />);

      expect(screen.getByText(longMessage)).toBeInTheDocument();
    });

    it('should handle special characters in message', () => {
      render(<LoadingSpinner message="Loading... 100% complete! 🎉" />);

      expect(screen.getByText('Loading... 100% complete! 🎉')).toBeInTheDocument();
    });

    it('should handle message with HTML entities', () => {
      render(<LoadingSpinner message="Loading &amp; processing..." />);

      expect(screen.getByText('Loading & processing...')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible structure', () => {
      render(<LoadingSpinner message="Loading content..." />);

      // Message should be visible text
      const message = screen.getByText('Loading content...');
      expect(message).toBeVisible();
    });

    it('should maintain semantic HTML structure', () => {
      render(<LoadingSpinner message="Please wait..." />);

      // Container should be a div
      const container = document.querySelector('.loading-container');
      expect(container?.tagName).toBe('DIV');

      // Message should be a paragraph
      const message = screen.getByText('Please wait...');
      expect(message.tagName).toBe('P');
    });
  });
});

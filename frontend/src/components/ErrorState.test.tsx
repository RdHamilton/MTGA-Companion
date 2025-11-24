import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ErrorState from './ErrorState';

describe('ErrorState', () => {
  describe('Required Props Rendering', () => {
    it('should render with message', () => {
      render(<ErrorState message="An error occurred" />);

      expect(screen.getByText('An error occurred')).toBeInTheDocument();
    });

    it('should always render warning icon', () => {
      render(<ErrorState message="Error" />);

      expect(screen.getByText('⚠️')).toBeInTheDocument();
      expect(document.querySelector('.error-state-icon')).toBeInTheDocument();
    });

    it('should render message as h2 element', () => {
      render(<ErrorState message="Error Title" />);

      const title = screen.getByText('Error Title');
      expect(title.tagName).toBe('H2');
      expect(title).toHaveClass('error-state-title');
    });
  });

  describe('Error Object Handling', () => {
    it('should render error message from Error object', () => {
      const error = new Error('Failed to fetch data');
      render(<ErrorState message="Request Failed" error={error} />);

      expect(screen.getByText('Request Failed')).toBeInTheDocument();
      expect(screen.getByText('Failed to fetch data')).toBeInTheDocument();
    });

    it('should render error details as paragraph element', () => {
      const error = new Error('Network error');
      render(<ErrorState message="Error" error={error} />);

      const details = screen.getByText('Network error');
      expect(details.tagName).toBe('P');
      expect(details).toHaveClass('error-state-details');
    });

    it('should handle error as string', () => {
      render(<ErrorState message="Error Occurred" error="Connection timeout" />);

      expect(screen.getByText('Connection timeout')).toBeInTheDocument();
    });

    it('should not render error details when error is not provided', () => {
      render(<ErrorState message="Error" />);

      const details = document.querySelector('.error-state-details');
      expect(details).not.toBeInTheDocument();
    });

    it('should handle Error object with empty message', () => {
      const error = new Error('');
      render(<ErrorState message="Error" error={error} />);

      // Empty error message should not render details
      const details = document.querySelector('.error-state-details');
      expect(details).not.toBeInTheDocument();
    });

    it('should handle empty string error', () => {
      render(<ErrorState message="Error" error="" />);

      // Empty string should not render details
      const details = document.querySelector('.error-state-details');
      expect(details).not.toBeInTheDocument();
    });
  });

  describe('Help Text', () => {
    it('should render help text when provided', () => {
      render(
        <ErrorState
          message="Connection Failed"
          helpText="Please check your internet connection and try again."
        />
      );

      expect(screen.getByText('Please check your internet connection and try again.')).toBeInTheDocument();
    });

    it('should not render help text when not provided', () => {
      render(<ErrorState message="Error" />);

      const helpText = document.querySelector('.error-state-help');
      expect(helpText).not.toBeInTheDocument();
    });

    it('should render help text as paragraph element', () => {
      render(
        <ErrorState
          message="Error"
          helpText="Try refreshing the page."
        />
      );

      const helpText = screen.getByText('Try refreshing the page.');
      expect(helpText.tagName).toBe('P');
      expect(helpText).toHaveClass('error-state-help');
    });

    it('should not render help text when empty string', () => {
      render(<ErrorState message="Error" helpText="" />);

      const helpText = document.querySelector('.error-state-help');
      expect(helpText).not.toBeInTheDocument();
    });
  });

  describe('Component Structure', () => {
    it('should render with correct container class', () => {
      render(<ErrorState message="Error" />);

      const container = document.querySelector('.error-state');
      expect(container).toBeInTheDocument();
    });

    it('should maintain correct DOM hierarchy with all props', () => {
      const error = new Error('Database connection failed');
      render(
        <ErrorState
          message="Cannot Load Data"
          error={error}
          helpText="Please try again later."
        />
      );

      const container = document.querySelector('.error-state');
      const icon = document.querySelector('.error-state-icon');
      const title = screen.getByText('Cannot Load Data');
      const details = screen.getByText('Database connection failed');
      const helpText = screen.getByText('Please try again later.');

      expect(container).toContainElement(icon);
      expect(container).toContainElement(title);
      expect(container).toContainElement(details);
      expect(container).toContainElement(helpText);
    });

    it('should maintain correct DOM hierarchy without optional props', () => {
      render(<ErrorState message="Error occurred" />);

      const container = document.querySelector('.error-state');
      const icon = document.querySelector('.error-state-icon');
      const title = screen.getByText('Error occurred');

      expect(container).toContainElement(icon);
      expect(container).toContainElement(title);
      expect(document.querySelector('.error-state-details')).not.toBeInTheDocument();
      expect(document.querySelector('.error-state-help')).not.toBeInTheDocument();
    });
  });

  describe('Real-World Error Scenarios', () => {
    it('should handle network error', () => {
      const error = new Error('ERR_NETWORK: Failed to connect to server');
      render(
        <ErrorState
          message="Network Error"
          error={error}
          helpText="Check your internet connection and try again."
        />
      );

      expect(screen.getByText('Network Error')).toBeInTheDocument();
      expect(screen.getByText('ERR_NETWORK: Failed to connect to server')).toBeInTheDocument();
      expect(screen.getByText('Check your internet connection and try again.')).toBeInTheDocument();
    });

    it('should handle API error', () => {
      render(
        <ErrorState
          message="Failed to Load Decks"
          error="API returned 500: Internal Server Error"
          helpText="Our servers are experiencing issues. Please try again later."
        />
      );

      expect(screen.getByText('Failed to Load Decks')).toBeInTheDocument();
      expect(screen.getByText('API returned 500: Internal Server Error')).toBeInTheDocument();
    });

    it('should handle validation error', () => {
      render(
        <ErrorState
          message="Invalid Deck"
          error="Deck must contain at least 60 cards"
          helpText="Add more cards to meet the minimum deck size requirement."
        />
      );

      expect(screen.getByText('Invalid Deck')).toBeInTheDocument();
      expect(screen.getByText('Deck must contain at least 60 cards')).toBeInTheDocument();
    });

    it('should handle file loading error', () => {
      const error = new Error('File not found: deck_export.txt');
      render(
        <ErrorState
          message="Cannot Import Deck"
          error={error}
        />
      );

      expect(screen.getByText('Cannot Import Deck')).toBeInTheDocument();
      expect(screen.getByText('File not found: deck_export.txt')).toBeInTheDocument();
    });

    it('should handle generic error without details', () => {
      render(
        <ErrorState
          message="Something went wrong"
          helpText="Please refresh the page and try again."
        />
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText('Please refresh the page and try again.')).toBeInTheDocument();
      expect(document.querySelector('.error-state-details')).not.toBeInTheDocument();
    });
  });

  describe('Content Variations', () => {
    it('should handle long error message', () => {
      const longMessage = 'A critical error has occurred while processing your request and the application cannot continue';
      render(<ErrorState message={longMessage} />);

      expect(screen.getByText(longMessage)).toBeInTheDocument();
    });

    it('should handle long error details', () => {
      const longError = 'Failed to establish a secure connection to the database server. The connection was refused after multiple retry attempts with increasing backoff intervals.';
      render(
        <ErrorState
          message="Database Error"
          error={longError}
        />
      );

      expect(screen.getByText(longError)).toBeInTheDocument();
    });

    it('should handle long help text', () => {
      const longHelpText = 'This error typically occurs when the server is down for maintenance or experiencing high load. Please wait a few minutes and try again. If the problem persists, contact support.';
      render(
        <ErrorState
          message="Service Unavailable"
          helpText={longHelpText}
        />
      );

      expect(screen.getByText(longHelpText)).toBeInTheDocument();
    });

    it('should handle special characters in content', () => {
      const error = new Error('JSON parse error: Unexpected token < in JSON at position 0');
      render(
        <ErrorState
          message="Invalid Response (500)"
          error={error}
          helpText={'Expected JSON but got HTML. Check server configuration.'}
        />
      );

      expect(screen.getByText('Invalid Response (500)')).toBeInTheDocument();
      expect(screen.getByText('JSON parse error: Unexpected token < in JSON at position 0')).toBeInTheDocument();
    });
  });

  describe('Error Object Properties', () => {
    it('should extract message from Error object', () => {
      const error = new Error('Test error message');
      error.name = 'CustomError';

      render(<ErrorState message="Error" error={error} />);

      // Should only show error.message, not error.name
      expect(screen.getByText('Test error message')).toBeInTheDocument();
      expect(screen.queryByText('CustomError')).not.toBeInTheDocument();
    });

    it('should handle TypeError', () => {
      const error = new TypeError('Cannot read property of undefined');
      render(<ErrorState message="JavaScript Error" error={error} />);

      expect(screen.getByText('Cannot read property of undefined')).toBeInTheDocument();
    });

    it('should handle RangeError', () => {
      const error = new RangeError('Maximum call stack size exceeded');
      render(<ErrorState message="Stack Overflow" error={error} />);

      expect(screen.getByText('Maximum call stack size exceeded')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should use semantic heading for message', () => {
      render(<ErrorState message="Accessible Error" />);

      const heading = screen.getByRole('heading', { level: 2 });
      expect(heading).toHaveTextContent('Accessible Error');
    });

    it('should have readable text content', () => {
      const error = new Error('Something went wrong');
      render(
        <ErrorState
          message="Error Occurred"
          error={error}
          helpText="Please try again."
        />
      );

      expect(screen.getByText('Error Occurred')).toBeVisible();
      expect(screen.getByText('Something went wrong')).toBeVisible();
      expect(screen.getByText('Please try again.')).toBeVisible();
    });

    it('should maintain proper heading hierarchy', () => {
      render(<ErrorState message="Error" />);

      const heading = screen.getByRole('heading', { level: 2 });
      expect(heading).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle error with null message', () => {
      const error = new Error();
      error.message = null as unknown as string;

      render(<ErrorState message="Error" error={error} />);

      // Null message should not render details
      const details = document.querySelector('.error-state-details');
      expect(details).not.toBeInTheDocument();
    });

    it('should handle HTML entities in text', () => {
      const error = new Error('Value must be > 0 and < 100');
      render(
        <ErrorState
          message="Validation Error &amp; More"
          error={error}
          helpText="Enter a value between 1 &amp; 99"
        />
      );

      expect(screen.getByText('Validation Error & More')).toBeInTheDocument();
      expect(screen.getByText('Value must be > 0 and < 100')).toBeInTheDocument();
      expect(screen.getByText('Enter a value between 1 & 99')).toBeInTheDocument();
    });

    it('should handle multiline error messages', () => {
      const error = new Error('Line 1\nLine 2\nLine 3');
      render(<ErrorState message="Multiline Error" error={error} />);

      // Use a text matcher to handle newlines
      const details = screen.getByText((content, element) => {
        return element?.className === 'error-state-details' && content.includes('Line 1');
      });
      expect(details).toBeInTheDocument();
      expect(details.textContent).toBe('Line 1\nLine 2\nLine 3');
    });
  });

  describe('Props Combinations', () => {
    it('should render with message only', () => {
      render(<ErrorState message="Simple Error" />);

      expect(screen.getByText('Simple Error')).toBeInTheDocument();
      expect(screen.getByText('⚠️')).toBeInTheDocument();
      expect(document.querySelector('.error-state-details')).not.toBeInTheDocument();
      expect(document.querySelector('.error-state-help')).not.toBeInTheDocument();
    });

    it('should render with message and error string', () => {
      render(
        <ErrorState
          message="Error"
          error="Error details"
        />
      );

      expect(screen.getByText('Error')).toBeInTheDocument();
      expect(screen.getByText('Error details')).toBeInTheDocument();
      expect(document.querySelector('.error-state-help')).not.toBeInTheDocument();
    });

    it('should render with message and help text', () => {
      render(
        <ErrorState
          message="Error"
          helpText="Help text"
        />
      );

      expect(screen.getByText('Error')).toBeInTheDocument();
      expect(screen.getByText('Help text')).toBeInTheDocument();
      expect(document.querySelector('.error-state-details')).not.toBeInTheDocument();
    });

    it('should render with all props', () => {
      const error = new Error('Detailed error');
      render(
        <ErrorState
          message="Complete Error"
          error={error}
          helpText="How to fix"
        />
      );

      expect(screen.getByText('Complete Error')).toBeInTheDocument();
      expect(screen.getByText('Detailed error')).toBeInTheDocument();
      expect(screen.getByText('How to fix')).toBeInTheDocument();
    });
  });
});

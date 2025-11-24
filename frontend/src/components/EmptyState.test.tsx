import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import EmptyState from './EmptyState';

describe('EmptyState', () => {
  describe('Required Props Rendering', () => {
    it('should render with title and message', () => {
      render(
        <EmptyState
          title="No Decks Found"
          message="You haven't created any decks yet."
        />
      );

      expect(screen.getByText('No Decks Found')).toBeInTheDocument();
      expect(screen.getByText("You haven't created any decks yet.")).toBeInTheDocument();
    });

    it('should render title as h2 element', () => {
      render(
        <EmptyState
          title="Empty State"
          message="No data available"
        />
      );

      const title = screen.getByText('Empty State');
      expect(title.tagName).toBe('H2');
      expect(title).toHaveClass('empty-state-title');
    });

    it('should render message as paragraph element', () => {
      render(
        <EmptyState
          title="No Items"
          message="The list is empty."
        />
      );

      const message = screen.getByText('The list is empty.');
      expect(message.tagName).toBe('P');
      expect(message).toHaveClass('empty-state-message');
    });
  });

  describe('Optional Icon', () => {
    it('should render icon when provided', () => {
      render(
        <EmptyState
          icon="📦"
          title="No Items"
          message="Your inventory is empty."
        />
      );

      const icon = document.querySelector('.empty-state-icon');
      expect(icon).toBeInTheDocument();
      expect(icon?.textContent).toBe('📦');
    });

    it('should not render icon when not provided', () => {
      render(
        <EmptyState
          title="No Items"
          message="Your inventory is empty."
        />
      );

      const icon = document.querySelector('.empty-state-icon');
      expect(icon).not.toBeInTheDocument();
    });

    it('should handle emoji icons', () => {
      render(
        <EmptyState
          icon="🎴"
          title="No Cards"
          message="No cards to display."
        />
      );

      expect(screen.getByText('🎴')).toBeInTheDocument();
    });

    it('should handle text-based icons', () => {
      render(
        <EmptyState
          icon="⚠️"
          title="Warning"
          message="No data found."
        />
      );

      expect(screen.getByText('⚠️')).toBeInTheDocument();
    });
  });

  describe('Optional Help Text', () => {
    it('should render help text when provided', () => {
      render(
        <EmptyState
          title="No Drafts"
          message="You haven't participated in any drafts."
          helpText="Start a draft in MTGA to see your draft history here."
        />
      );

      const helpText = screen.getByText('Start a draft in MTGA to see your draft history here.');
      expect(helpText).toBeInTheDocument();
      expect(helpText).toHaveClass('empty-state-help');
    });

    it('should not render help text when not provided', () => {
      render(
        <EmptyState
          title="Empty"
          message="No data"
        />
      );

      const helpText = document.querySelector('.empty-state-help');
      expect(helpText).not.toBeInTheDocument();
    });

    it('should render help text as paragraph element', () => {
      render(
        <EmptyState
          title="No Matches"
          message="No matches recorded."
          helpText="Play some games to see match history."
        />
      );

      const helpText = screen.getByText('Play some games to see match history.');
      expect(helpText.tagName).toBe('P');
    });
  });

  describe('Component Structure', () => {
    it('should render with correct container class', () => {
      render(
        <EmptyState
          title="Test"
          message="Test message"
        />
      );

      const container = document.querySelector('.empty-state');
      expect(container).toBeInTheDocument();
    });

    it('should maintain correct DOM hierarchy with all props', () => {
      render(
        <EmptyState
          icon="🔍"
          title="Not Found"
          message="Search returned no results."
          helpText="Try different search terms."
        />
      );

      const container = document.querySelector('.empty-state');
      const icon = document.querySelector('.empty-state-icon');
      const title = screen.getByText('Not Found');
      const message = screen.getByText('Search returned no results.');
      const helpText = screen.getByText('Try different search terms.');

      expect(container).toContainElement(icon);
      expect(container).toContainElement(title);
      expect(container).toContainElement(message);
      expect(container).toContainElement(helpText);
    });

    it('should maintain correct DOM hierarchy without optional props', () => {
      render(
        <EmptyState
          title="Empty"
          message="No data available."
        />
      );

      const container = document.querySelector('.empty-state');
      const title = screen.getByText('Empty');
      const message = screen.getByText('No data available.');

      expect(container).toContainElement(title);
      expect(container).toContainElement(message);
      expect(document.querySelector('.empty-state-icon')).not.toBeInTheDocument();
      expect(document.querySelector('.empty-state-help')).not.toBeInTheDocument();
    });
  });

  describe('Content Variations', () => {
    it('should handle long title text', () => {
      const longTitle = 'This is a very long title that might span multiple lines in the UI';
      render(
        <EmptyState
          title={longTitle}
          message="Short message"
        />
      );

      expect(screen.getByText(longTitle)).toBeInTheDocument();
    });

    it('should handle long message text', () => {
      const longMessage = 'This is a very long message that provides detailed information about why the state is empty and what the user can do to populate it with data.';
      render(
        <EmptyState
          title="Empty State"
          message={longMessage}
        />
      );

      expect(screen.getByText(longMessage)).toBeInTheDocument();
    });

    it('should handle long help text', () => {
      const longHelpText = 'This is extensive help text that provides step-by-step instructions on how to get started and populate this empty state with meaningful data.';
      render(
        <EmptyState
          title="No Data"
          message="Nothing here yet."
          helpText={longHelpText}
        />
      );

      expect(screen.getByText(longHelpText)).toBeInTheDocument();
    });

    it('should handle special characters in content', () => {
      render(
        <EmptyState
          icon="⚡"
          title="No Results (0)"
          message="Search for 'Magic: The Gathering' returned nothing."
          helpText={'Try searching for "MTG" instead.'}
        />
      );

      expect(screen.getByText('No Results (0)')).toBeInTheDocument();
      expect(screen.getByText("Search for 'Magic: The Gathering' returned nothing.")).toBeInTheDocument();
      expect(screen.getByText('Try searching for "MTG" instead.')).toBeInTheDocument();
    });
  });

  describe('Real-World Use Cases', () => {
    it('should render empty deck state', () => {
      render(
        <EmptyState
          icon="🃏"
          title="No Decks"
          message="You haven't created any decks yet."
          helpText="Click the 'Create Deck' button to build your first deck."
        />
      );

      expect(screen.getByText('No Decks')).toBeInTheDocument();
      expect(screen.getByText("You haven't created any decks yet.")).toBeInTheDocument();
      expect(screen.getByText("Click the 'Create Deck' button to build your first deck.")).toBeInTheDocument();
    });

    it('should render empty draft state', () => {
      render(
        <EmptyState
          icon="🎯"
          title="No Draft History"
          message="You haven't participated in any drafts."
          helpText="Your draft picks and statistics will appear here after your first draft."
        />
      );

      expect(screen.getByText('No Draft History')).toBeInTheDocument();
      expect(screen.getByText("You haven't participated in any drafts.")).toBeInTheDocument();
    });

    it('should render empty search results', () => {
      render(
        <EmptyState
          icon="🔍"
          title="No Cards Found"
          message="Your search didn't match any cards."
          helpText="Try adjusting your filters or search terms."
        />
      );

      expect(screen.getByText('No Cards Found')).toBeInTheDocument();
      expect(screen.getByText("Your search didn't match any cards.")).toBeInTheDocument();
    });

    it('should render without icon or help text', () => {
      render(
        <EmptyState
          title="No Data"
          message="Nothing to display."
        />
      );

      expect(screen.getByText('No Data')).toBeInTheDocument();
      expect(screen.getByText('Nothing to display.')).toBeInTheDocument();
      expect(document.querySelector('.empty-state-icon')).not.toBeInTheDocument();
      expect(document.querySelector('.empty-state-help')).not.toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should use semantic heading for title', () => {
      render(
        <EmptyState
          title="Accessible Title"
          message="Accessible message"
        />
      );

      const title = screen.getByRole('heading', { level: 2 });
      expect(title).toHaveTextContent('Accessible Title');
    });

    it('should have readable text content', () => {
      render(
        <EmptyState
          icon="📋"
          title="No Items"
          message="The list is empty."
          helpText="Add items to see them here."
        />
      );

      expect(screen.getByText('No Items')).toBeVisible();
      expect(screen.getByText('The list is empty.')).toBeVisible();
      expect(screen.getByText('Add items to see them here.')).toBeVisible();
    });

    it('should maintain proper heading hierarchy', () => {
      render(
        <EmptyState
          title="Empty State"
          message="No content"
        />
      );

      // H2 heading should be accessible
      const heading = screen.getByRole('heading', { level: 2 });
      expect(heading).toBeInTheDocument();
      expect(heading).toHaveTextContent('Empty State');
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty string for icon', () => {
      render(
        <EmptyState
          icon=""
          title="No Icon"
          message="Icon is empty string"
        />
      );

      // Empty icon should still render the container but be empty
      const icon = document.querySelector('.empty-state-icon');
      expect(icon).not.toBeInTheDocument();
    });

    it('should handle empty string for helpText', () => {
      render(
        <EmptyState
          title="No Help"
          message="Help text is empty"
          helpText=""
        />
      );

      // Empty help text should not render the element
      const helpText = document.querySelector('.empty-state-help');
      expect(helpText).not.toBeInTheDocument();
    });

    it('should handle HTML entities in text', () => {
      render(
        <EmptyState
          title="No Results &amp; More"
          message="Less than 1 &lt; 2"
          helpText="Greater than 2 &gt; 1"
        />
      );

      expect(screen.getByText('No Results & More')).toBeInTheDocument();
      expect(screen.getByText('Less than 1 < 2')).toBeInTheDocument();
      expect(screen.getByText('Greater than 2 > 1')).toBeInTheDocument();
    });
  });
});

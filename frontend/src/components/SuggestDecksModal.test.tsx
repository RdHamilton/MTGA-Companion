import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SuggestDecksModal from './SuggestDecksModal';

// Mock the Wails bindings
vi.mock('../../wailsjs/go/main/App', () => ({
  SuggestDecks: vi.fn(),
  ApplySuggestedDeck: vi.fn(),
  ExportSuggestedDeck: vi.fn(),
}));

beforeEach(() => {
  vi.clearAllMocks();
});

describe('SuggestDecksModal', () => {
  it('should not render when isOpen is false', () => {
    render(
      <SuggestDecksModal
        isOpen={false}
        onClose={() => {}}
        draftEventID="test-draft-id"
        currentDeckID="test-deck-id"
        deckName="Test Deck"
        onDeckApplied={() => {}}
      />
    );

    expect(screen.queryByText('Suggested Decks')).not.toBeInTheDocument();
  });

  it('should render modal content when open', async () => {
    render(
      <SuggestDecksModal
        isOpen={true}
        onClose={() => {}}
        draftEventID="test-draft-id"
        currentDeckID="test-deck-id"
        deckName="Test Deck"
        onDeckApplied={() => {}}
      />
    );

    // Modal should be visible with content area
    expect(document.querySelector('.suggest-decks-content')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', async () => {
    const onClose = vi.fn();

    render(
      <SuggestDecksModal
        isOpen={true}
        onClose={onClose}
        draftEventID="test-draft-id"
        currentDeckID="test-deck-id"
        deckName="Test Deck"
        onDeckApplied={() => {}}
      />
    );

    const closeButton = screen.getByRole('button', { name: /×/ });
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalled();
  });

  it('should call onClose when clicking overlay', async () => {
    const onClose = vi.fn();

    render(
      <SuggestDecksModal
        isOpen={true}
        onClose={onClose}
        draftEventID="test-draft-id"
        currentDeckID="test-deck-id"
        deckName="Test Deck"
        onDeckApplied={() => {}}
      />
    );

    // Click the overlay (background)
    const overlay = document.querySelector('.suggest-decks-overlay');
    if (overlay) {
      fireEvent.click(overlay);
    }

    expect(onClose).toHaveBeenCalled();
  });

  it('should render modal header when open', () => {
    render(
      <SuggestDecksModal
        isOpen={true}
        onClose={() => {}}
        draftEventID="test-draft-id"
        currentDeckID="test-deck-id"
        deckName="Test Deck"
        onDeckApplied={() => {}}
      />
    );

    expect(screen.getByText('Suggested Decks')).toBeInTheDocument();
  });
});

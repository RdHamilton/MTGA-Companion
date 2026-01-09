import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { mockNotes } from '@/test/mocks/apiMock';
import type { ImprovementSuggestion } from '@/services/api/notes';
import ImprovementSuggestionsPanel from './ImprovementSuggestionsPanel';

// Mock the API module
vi.mock('@/services/api', () => ({
  notes: mockNotes,
}));

// Helper to create mock suggestions
function createMockSuggestion(overrides: Partial<ImprovementSuggestion> = {}): ImprovementSuggestion {
  return {
    id: 1,
    deckId: 'deck-1',
    suggestionType: 'curve',
    priority: 'medium',
    title: 'Test Suggestion',
    description: 'Test description for the suggestion',
    evidence: '',
    cardReferences: '',
    isDismissed: false,
    createdAt: '2024-01-15T10:00:00Z',
    ...overrides,
  };
}

describe('ImprovementSuggestionsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching suggestions', async () => {
      let resolvePromise: (value: ImprovementSuggestion[]) => void;
      const loadingPromise = new Promise<ImprovementSuggestion[]>((resolve) => {
        resolvePromise = resolve;
      });
      mockNotes.getDeckSuggestions.mockReturnValue(loadingPromise);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      expect(screen.getByText('Loading suggestions...')).toBeInTheDocument();

      resolvePromise!([createMockSuggestion()]);
      await waitFor(() => {
        expect(screen.queryByText('Loading suggestions...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no suggestions exist', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('No suggestions yet.')).toBeInTheDocument();
      });
    });

    it('should show hint about required games', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Requires at least 5 games played with this deck.')).toBeInTheDocument();
      });
    });
  });

  describe('Suggestions List', () => {
    it('should display suggestions when loaded', async () => {
      const suggestions = [
        createMockSuggestion({ id: 1, title: 'Curve Too High' }),
        createMockSuggestion({ id: 2, title: 'Add More Lands' }),
      ];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Curve Too High')).toBeInTheDocument();
        expect(screen.getByText('Add More Lands')).toBeInTheDocument();
      });
    });

    it('should sort suggestions by priority (high first)', async () => {
      const suggestions = [
        createMockSuggestion({ id: 1, title: 'Low Priority', priority: 'low' }),
        createMockSuggestion({ id: 2, title: 'High Priority', priority: 'high' }),
        createMockSuggestion({ id: 3, title: 'Medium Priority', priority: 'medium' }),
      ];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        const titles = screen.getAllByRole('heading', { level: 4 });
        expect(titles[0]).toHaveTextContent('High Priority');
        expect(titles[1]).toHaveTextContent('Medium Priority');
        expect(titles[2]).toHaveTextContent('Low Priority');
      });
    });

    it('should display suggestion type icon', async () => {
      const suggestions = [createMockSuggestion({ suggestionType: 'mana' })];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        // Mana type should show water drop emoji
        expect(screen.getByText('💧')).toBeInTheDocument();
      });
    });
  });

  describe('Expand/Collapse', () => {
    it('should expand suggestion to show details when clicked', async () => {
      const suggestions = [
        createMockSuggestion({ title: 'Test Title', description: 'Detailed description here' }),
      ];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Test Title')).toBeInTheDocument();
      });

      // Description should not be visible initially
      expect(screen.queryByText('Detailed description here')).not.toBeInTheDocument();

      // Click to expand
      fireEvent.click(screen.getByText('Test Title'));

      await waitFor(() => {
        expect(screen.getByText('Detailed description here')).toBeInTheDocument();
      });
    });

    it('should collapse suggestion when clicked again', async () => {
      const suggestions = [
        createMockSuggestion({ title: 'Test Title', description: 'Detailed description here' }),
      ];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Test Title')).toBeInTheDocument();
      });

      // Expand
      fireEvent.click(screen.getByText('Test Title'));
      await waitFor(() => {
        expect(screen.getByText('Detailed description here')).toBeInTheDocument();
      });

      // Collapse
      fireEvent.click(screen.getByText('Test Title'));
      await waitFor(() => {
        expect(screen.queryByText('Detailed description here')).not.toBeInTheDocument();
      });
    });
  });

  describe('Generate Suggestions', () => {
    it('should call generate when button clicked', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);
      mockNotes.generateSuggestions.mockResolvedValue([createMockSuggestion()]);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Generate New Suggestions')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Generate New Suggestions'));

      await waitFor(() => {
        expect(mockNotes.generateSuggestions).toHaveBeenCalledWith('deck-1');
      });
    });

    it('should show analyzing state while generating', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);
      let resolveGenerate: (value: ImprovementSuggestion[]) => void;
      const generatePromise = new Promise<ImprovementSuggestion[]>((resolve) => {
        resolveGenerate = resolve;
      });
      mockNotes.generateSuggestions.mockReturnValue(generatePromise);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Generate New Suggestions'));
      });

      expect(screen.getByText('Analyzing...')).toBeInTheDocument();

      resolveGenerate!([]);
      await waitFor(() => {
        expect(screen.queryByText('Analyzing...')).not.toBeInTheDocument();
      });
    });

    it('should show error when insufficient games', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);
      mockNotes.generateSuggestions.mockRejectedValue(new Error('insufficient games'));

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Generate New Suggestions'));
      });

      await waitFor(() => {
        expect(screen.getByText(/Not enough games played/)).toBeInTheDocument();
      });
    });
  });

  describe('Dismiss Suggestion', () => {
    it('should dismiss suggestion when dismiss button clicked', async () => {
      const suggestions = [createMockSuggestion({ id: 1, title: 'Test' })];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);
      mockNotes.dismissSuggestion.mockResolvedValue(undefined);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Test'));
      });

      // Find and click dismiss button
      fireEvent.click(screen.getByText('Dismiss'));

      await waitFor(() => {
        expect(mockNotes.dismissSuggestion).toHaveBeenCalledWith(1);
      });
    });
  });

  describe('Filter by Type', () => {
    it('should filter suggestions by type', async () => {
      const suggestions = [
        createMockSuggestion({ id: 1, title: 'Curve Issue', suggestionType: 'curve' }),
        createMockSuggestion({ id: 2, title: 'Mana Issue', suggestionType: 'mana' }),
      ];
      mockNotes.getDeckSuggestions.mockResolvedValue(suggestions);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Curve Issue')).toBeInTheDocument();
        expect(screen.getByText('Mana Issue')).toBeInTheDocument();
      });

      // Filter by curve
      fireEvent.change(screen.getByRole('combobox'), { target: { value: 'curve' } });

      await waitFor(() => {
        expect(screen.getByText('Curve Issue')).toBeInTheDocument();
        expect(screen.queryByText('Mana Issue')).not.toBeInTheDocument();
      });
    });
  });

  describe('Show Dismissed Toggle', () => {
    it('should toggle showing dismissed suggestions', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByLabelText('Show dismissed')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByLabelText('Show dismissed'));

      await waitFor(() => {
        // Should refetch with showDismissed = false (activeOnly = false)
        expect(mockNotes.getDeckSuggestions).toHaveBeenCalledWith('deck-1', false);
      });
    });
  });

  describe('Close Button', () => {
    it('should call onClose when close button clicked', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);
      const onClose = vi.fn();

      render(<ImprovementSuggestionsPanel deckId="deck-1" onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByTitle('Close')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByTitle('Close'));

      expect(onClose).toHaveBeenCalled();
    });

    it('should not show close button when onClose not provided', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.queryByTitle('Close')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should show error message when load fails', async () => {
      mockNotes.getDeckSuggestions.mockRejectedValue(new Error('Network error'));

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });
    });

    it('should allow dismissing error message', async () => {
      mockNotes.getDeckSuggestions.mockResolvedValue([]);
      mockNotes.generateSuggestions.mockRejectedValue(new Error('Generation failed'));

      render(<ImprovementSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Generate New Suggestions'));
      });

      await waitFor(() => {
        expect(screen.getByText('Generation failed')).toBeInTheDocument();
      });

      // Click dismiss on error banner
      const dismissButtons = screen.getAllByText('Dismiss');
      fireEvent.click(dismissButtons[0]);

      await waitFor(() => {
        expect(screen.queryByText('Generation failed')).not.toBeInTheDocument();
      });
    });
  });
});

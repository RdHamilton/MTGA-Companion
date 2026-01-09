import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { mockMLSuggestions } from '@/test/mocks/apiMock';
import type { MLSuggestion, MLSuggestionResult } from '@/services/api/mlSuggestions';
import MLSuggestionsPanel from './MLSuggestionsPanel';

// Mock the API module
vi.mock('@/services/api', () => ({
  mlSuggestions: mockMLSuggestions,
}));

// Helper to create mock ML suggestion
function createMockMLSuggestion(overrides: Partial<MLSuggestion> = {}): MLSuggestion {
  return {
    id: 1,
    deckId: 'deck-1',
    suggestionType: 'add',
    cardId: 12345,
    cardName: 'Lightning Bolt',
    swapForCardId: undefined,
    swapForCardName: undefined,
    confidence: 0.75,
    expectedWinRateChange: 2.5,
    title: 'Add Lightning Bolt',
    description: 'This card has strong synergy with your deck.',
    reasoning: JSON.stringify([
      { type: 'synergy', description: 'Works well with Goblin Guide', impact: 0.8, confidence: 0.7 },
    ]),
    evidence: undefined,
    isDismissed: false,
    wasApplied: false,
    outcomeWinRateChange: undefined,
    createdAt: '2024-01-15T10:00:00Z',
    appliedAt: undefined,
    outcomeRecordedAt: undefined,
    ...overrides,
  };
}

function createMockMLSuggestionResult(
  overrides: Partial<MLSuggestion> = {}
): MLSuggestionResult {
  return {
    suggestion: createMockMLSuggestion(overrides),
    synergyData: [],
    reasons: [
      { type: 'synergy', description: 'Works well with Goblin Guide', impact: 0.8, confidence: 0.7 },
    ],
  };
}

describe('MLSuggestionsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset mock implementations to defaults
    mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);
    mockMLSuggestions.generateMLSuggestions.mockResolvedValue([]);
    mockMLSuggestions.dismissMLSuggestion.mockResolvedValue(undefined);
    mockMLSuggestions.applyMLSuggestion.mockResolvedValue(undefined);
    mockMLSuggestions.parseReasons.mockImplementation((reasoning: string | undefined) => {
      if (!reasoning) return [];
      try {
        return JSON.parse(reasoning);
      } catch {
        return [];
      }
    });
    mockMLSuggestions.getMLSuggestionTypeLabel.mockImplementation((type: string) => {
      const labels: Record<string, string> = { add: 'Add Card', remove: 'Remove Card', swap: 'Swap Cards' };
      return labels[type] || type;
    });
    mockMLSuggestions.getMLSuggestionTypeIcon.mockImplementation((type: string) => {
      const icons: Record<string, string> = { add: '+', remove: '-', swap: '⇄' };
      return icons[type] || '?';
    });
    mockMLSuggestions.formatConfidence.mockImplementation((c: number) => `${Math.round(c * 100)}%`);
    mockMLSuggestions.formatWinRateChange.mockImplementation(
      (c: number) => `${c >= 0 ? '+' : ''}${c.toFixed(1)}%`
    );
    mockMLSuggestions.getConfidenceColor.mockImplementation((confidence: number) => {
      if (confidence >= 0.7) return 'text-green-400';
      if (confidence >= 0.4) return 'text-blue-400';
      return 'text-yellow-400';
    });
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching suggestions', async () => {
      let resolvePromise: (value: MLSuggestion[]) => void;
      const loadingPromise = new Promise<MLSuggestion[]>((resolve) => {
        resolvePromise = resolve;
      });
      mockMLSuggestions.getMLSuggestions.mockReturnValue(loadingPromise);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      expect(screen.getByText('Loading ML suggestions...')).toBeInTheDocument();

      resolvePromise!([createMockMLSuggestion()]);
      await waitFor(() => {
        expect(screen.queryByText('Loading ML suggestions...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no suggestions exist', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('No ML suggestions yet.')).toBeInTheDocument();
      });
    });

    it('should show hint about generating suggestions', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(
          screen.getByText(/Click.*Generate ML Suggestions.*to analyze card synergies!/)
        ).toBeInTheDocument();
      });
    });
  });

  describe('Suggestions List', () => {
    it('should display suggestions when loaded', async () => {
      const suggestions = [
        createMockMLSuggestion({ id: 1, title: 'Add Lightning Bolt' }),
        createMockMLSuggestion({ id: 2, title: 'Remove Shock' }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Add Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Remove Shock')).toBeInTheDocument();
      });
    });

    it('should sort suggestions by confidence (high first)', async () => {
      const suggestions = [
        createMockMLSuggestion({ id: 1, title: 'Low Confidence', confidence: 0.3 }),
        createMockMLSuggestion({ id: 2, title: 'High Confidence', confidence: 0.9 }),
        createMockMLSuggestion({ id: 3, title: 'Medium Confidence', confidence: 0.6 }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        const titles = screen.getAllByRole('heading', { level: 4 });
        expect(titles[0]).toHaveTextContent('High Confidence');
        expect(titles[1]).toHaveTextContent('Medium Confidence');
        expect(titles[2]).toHaveTextContent('Low Confidence');
      });
    });

    it('should display suggestion type icon', async () => {
      const suggestions = [createMockMLSuggestion({ suggestionType: 'add' })];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('+')).toBeInTheDocument();
      });
    });

    it('should display confidence badge', async () => {
      const suggestions = [createMockMLSuggestion({ confidence: 0.85 })];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('85%')).toBeInTheDocument();
      });
    });

    it('should display win rate change when non-zero', async () => {
      const suggestions = [createMockMLSuggestion({ expectedWinRateChange: 3.5 })];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('+3.5%')).toBeInTheDocument();
      });
    });
  });

  describe('Expand/Collapse', () => {
    it('should expand suggestion to show details when clicked', async () => {
      const suggestions = [
        createMockMLSuggestion({
          title: 'Test Title',
          description: 'Detailed ML description here',
        }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Test Title')).toBeInTheDocument();
      });

      // Description should not be visible initially
      expect(screen.queryByText('Detailed ML description here')).not.toBeInTheDocument();

      // Click to expand
      fireEvent.click(screen.getByText('Test Title'));

      await waitFor(() => {
        expect(screen.getByText('Detailed ML description here')).toBeInTheDocument();
      });
    });

    it('should collapse suggestion when clicked again', async () => {
      const suggestions = [
        createMockMLSuggestion({
          title: 'Test Title',
          description: 'Detailed ML description here',
        }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Test Title')).toBeInTheDocument();
      });

      // Expand
      fireEvent.click(screen.getByText('Test Title'));
      await waitFor(() => {
        expect(screen.getByText('Detailed ML description here')).toBeInTheDocument();
      });

      // Collapse
      fireEvent.click(screen.getByText('Test Title'));
      await waitFor(() => {
        expect(screen.queryByText('Detailed ML description here')).not.toBeInTheDocument();
      });
    });

    it('should show swap info for swap suggestions', async () => {
      const suggestions = [
        createMockMLSuggestion({
          suggestionType: 'swap',
          cardName: 'Shock',
          swapForCardName: 'Lightning Bolt',
        }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getAllByRole('heading', { level: 4 })[0]);
      });

      await waitFor(() => {
        expect(screen.getByText('Shock')).toBeInTheDocument();
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
    });
  });

  describe('Generate Suggestions', () => {
    it('should call generate when button clicked', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);
      mockMLSuggestions.generateMLSuggestions.mockResolvedValue([
        createMockMLSuggestionResult(),
      ]);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Generate ML Suggestions')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Generate ML Suggestions'));

      await waitFor(() => {
        expect(mockMLSuggestions.generateMLSuggestions).toHaveBeenCalledWith('deck-1');
      });
    });

    it('should show analyzing state while generating', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);
      let resolveGenerate: (value: MLSuggestionResult[]) => void;
      const generatePromise = new Promise<MLSuggestionResult[]>((resolve) => {
        resolveGenerate = resolve;
      });
      mockMLSuggestions.generateMLSuggestions.mockReturnValue(generatePromise);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Generate ML Suggestions'));
      });

      expect(screen.getByText('Analyzing synergies...')).toBeInTheDocument();

      resolveGenerate!([]);
      await waitFor(() => {
        expect(screen.queryByText('Analyzing synergies...')).not.toBeInTheDocument();
      });
    });

    it('should show error when no synergy data available', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);
      mockMLSuggestions.generateMLSuggestions.mockRejectedValue(new Error('no synergy data'));

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Generate ML Suggestions'));
      });

      await waitFor(() => {
        expect(
          screen.getByText(/No synergy data available/)
        ).toBeInTheDocument();
      });
    });
  });

  describe('Dismiss Suggestion', () => {
    it('should dismiss suggestion when dismiss button clicked', async () => {
      const suggestions = [createMockMLSuggestion({ id: 1, title: 'Test' })];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);
      mockMLSuggestions.dismissMLSuggestion.mockResolvedValue(undefined);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Test'));
      });

      // Find and click dismiss button
      fireEvent.click(screen.getByText('Dismiss'));

      await waitFor(() => {
        expect(mockMLSuggestions.dismissMLSuggestion).toHaveBeenCalledWith(1);
      });
    });
  });

  describe('Apply Suggestion', () => {
    it('should apply suggestion when apply button clicked', async () => {
      const suggestions = [createMockMLSuggestion({ id: 1, title: 'Test' })];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);
      mockMLSuggestions.applyMLSuggestion.mockResolvedValue(undefined);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Test'));
      });

      // Find and click apply button
      fireEvent.click(screen.getByText('Mark Applied'));

      await waitFor(() => {
        expect(mockMLSuggestions.applyMLSuggestion).toHaveBeenCalledWith(1);
      });
    });

    it('should not show action buttons for applied suggestions', async () => {
      const suggestions = [
        createMockMLSuggestion({
          id: 1,
          title: 'Applied Suggestion',
          wasApplied: true,
          appliedAt: '2024-01-16T10:00:00Z',
        }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Applied Suggestion'));
      });

      await waitFor(() => {
        expect(screen.queryByText('Mark Applied')).not.toBeInTheDocument();
        expect(screen.queryByText('Dismiss')).not.toBeInTheDocument();
      });
    });
  });

  describe('Filter by Type', () => {
    it('should filter suggestions by type', async () => {
      const suggestions = [
        createMockMLSuggestion({ id: 1, title: 'Add Lightning Bolt', suggestionType: 'add' }),
        createMockMLSuggestion({ id: 2, title: 'Remove Shock', suggestionType: 'remove' }),
      ];
      mockMLSuggestions.getMLSuggestions.mockResolvedValue(suggestions);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Add Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Remove Shock')).toBeInTheDocument();
      });

      // Filter by add
      fireEvent.change(screen.getByRole('combobox'), { target: { value: 'add' } });

      await waitFor(() => {
        expect(screen.getByText('Add Lightning Bolt')).toBeInTheDocument();
        expect(screen.queryByText('Remove Shock')).not.toBeInTheDocument();
      });
    });
  });

  describe('Show Dismissed Toggle', () => {
    it('should toggle showing dismissed suggestions', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByLabelText('Show dismissed')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByLabelText('Show dismissed'));

      await waitFor(() => {
        // Should refetch with showDismissed = false (activeOnly = false)
        expect(mockMLSuggestions.getMLSuggestions).toHaveBeenCalledWith('deck-1', false);
      });
    });
  });

  describe('Close Button', () => {
    it('should call onClose when close button clicked', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);
      const onClose = vi.fn();

      render(<MLSuggestionsPanel deckId="deck-1" onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByTitle('Close')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByTitle('Close'));

      expect(onClose).toHaveBeenCalled();
    });

    it('should not show close button when onClose not provided', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.queryByTitle('Close')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should show error message when load fails', async () => {
      mockMLSuggestions.getMLSuggestions.mockRejectedValue(new Error('Network error'));

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });
    });

    it('should allow dismissing error message', async () => {
      mockMLSuggestions.getMLSuggestions.mockResolvedValue([]);
      mockMLSuggestions.generateMLSuggestions.mockRejectedValue(new Error('Generation failed'));

      render(<MLSuggestionsPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Generate ML Suggestions'));
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

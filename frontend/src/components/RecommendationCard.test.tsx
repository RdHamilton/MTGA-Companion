import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import RecommendationCard from './RecommendationCard';

// Mock the Wails API
vi.mock('@/services/api/legacy', () => ({
  ExplainRecommendation: vi.fn(),
}));

import { ExplainRecommendation } from '@/services/api/legacy';

const mockRecommendation = {
  cardID: 12345,
  name: 'Lightning Bolt',
  typeLine: 'Instant',
  manaCost: '{R}',
  imageURI: 'https://example.com/lightning-bolt.jpg',
  score: 0.85,
  reasoning: 'Efficient removal spell that fits the aggro strategy',
  source: 'ml',
  confidence: 0.92,
  factors: {
    colorFit: 0.95,
    manaCurve: 0.8,
    synergy: 0.75,
    quality: 0.9,
    playable: 1.0,
  },
};

const defaultProps = {
  recommendation: mockRecommendation as any,
  deckID: 'test-deck-id',
  onAddCard: vi.fn() as unknown as (cardID: number, quantity: number, board: 'main' | 'sideboard') => void,
};

describe('RecommendationCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(ExplainRecommendation).mockResolvedValue({
      explanation: 'This card is recommended because it provides efficient removal.',
      error: '',
    });
  });

  describe('rendering', () => {
    it('renders the card name', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
    });

    it('renders the card type line', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('Instant')).toBeInTheDocument();
    });

    it('renders the mana cost', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('{R}')).toBeInTheDocument();
    });

    it('renders the card image when imageURI is provided', () => {
      render(<RecommendationCard {...defaultProps} />);
      const image = screen.getByAltText('Lightning Bolt');
      expect(image).toBeInTheDocument();
      expect(image).toHaveAttribute('src', 'https://example.com/lightning-bolt.jpg');
    });

    it('does not render image when imageURI is not provided', () => {
      const propsWithoutImage = {
        ...defaultProps,
        recommendation: { ...mockRecommendation, imageURI: '' } as any,
      };
      render(<RecommendationCard {...propsWithoutImage} />);
      expect(screen.queryByRole('img')).not.toBeInTheDocument();
    });

    it('renders the score as percentage', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('85%')).toBeInTheDocument();
    });

    it('renders the confidence badge', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('92% confidence')).toBeInTheDocument();
    });

    it('renders the reasoning text', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('Efficient removal spell that fits the aggro strategy')).toBeInTheDocument();
    });

    it('renders the Add button', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('+ Add')).toBeInTheDocument();
    });

    it('renders the Why button', () => {
      render(<RecommendationCard {...defaultProps} />);
      expect(screen.getByText('? Why')).toBeInTheDocument();
    });
  });

  describe('add card action', () => {
    it('calls onAddCard when Add button is clicked', () => {
      render(<RecommendationCard {...defaultProps} />);
      fireEvent.click(screen.getByText('+ Add'));
      expect(defaultProps.onAddCard).toHaveBeenCalledWith(12345, 1, 'main');
    });
  });

  describe('expand/collapse details', () => {
    it('expands details when Why button is clicked', async () => {
      render(<RecommendationCard {...defaultProps} />);

      expect(screen.queryByText('Score Breakdown')).not.toBeInTheDocument();

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('Score Breakdown')).toBeInTheDocument();
      });
    });

    it('shows Less button when expanded', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('- Less')).toBeInTheDocument();
      });
    });

    it('collapses details when Less button is clicked', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));
      await waitFor(() => {
        expect(screen.getByText('Score Breakdown')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('- Less'));

      await waitFor(() => {
        expect(screen.queryByText('Score Breakdown')).not.toBeInTheDocument();
      });
    });
  });

  describe('score factors', () => {
    it('displays score factors when expanded', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('Color Fit')).toBeInTheDocument();
        expect(screen.getByText('Mana Curve')).toBeInTheDocument();
        expect(screen.getByText('Synergy')).toBeInTheDocument();
        expect(screen.getByText('Card Quality')).toBeInTheDocument();
        expect(screen.getByText('Playability')).toBeInTheDocument();
      });
    });

    it('displays factor percentages when expanded', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('95%')).toBeInTheDocument(); // colorFit
        expect(screen.getByText('80%')).toBeInTheDocument(); // manaCurve
        expect(screen.getByText('75%')).toBeInTheDocument(); // synergy
        expect(screen.getByText('90%')).toBeInTheDocument(); // quality
        expect(screen.getByText('100%')).toBeInTheDocument(); // playable
      });
    });
  });

  describe('explanation loading', () => {
    it('calls ExplainRecommendation API when expanded', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(ExplainRecommendation).toHaveBeenCalledWith({
          deckID: 'test-deck-id',
          cardID: 12345,
        });
      });
    });

    it('shows loading state while fetching explanation', async () => {
      vi.mocked(ExplainRecommendation).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('Generating explanation...')).toBeInTheDocument();
      });
    });

    it('displays explanation when loaded', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('This card is recommended because it provides efficient removal.')).toBeInTheDocument();
      });
    });

    it('does not call API again when re-expanding', async () => {
      render(<RecommendationCard {...defaultProps} />);

      // Expand
      fireEvent.click(screen.getByText('? Why'));
      await waitFor(() => {
        expect(ExplainRecommendation).toHaveBeenCalledTimes(1);
      });

      // Collapse
      fireEvent.click(screen.getByText('- Less'));

      // Expand again
      fireEvent.click(screen.getByText('? Why'));

      // Should not call API again
      expect(ExplainRecommendation).toHaveBeenCalledTimes(1);
    });
  });

  describe('explanation error handling', () => {
    it('displays error when API returns error', async () => {
      vi.mocked(ExplainRecommendation).mockResolvedValueOnce({
        explanation: '',
        error: 'LLM not available',
      });

      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText(/Unable to generate explanation: LLM not available/)).toBeInTheDocument();
      });
    });

    it('displays error when API throws', async () => {
      vi.mocked(ExplainRecommendation).mockRejectedValueOnce(new Error('Network error'));

      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText(/Unable to generate explanation: Network error/)).toBeInTheDocument();
      });
    });

    it('falls back to reasoning when no explanation available', async () => {
      vi.mocked(ExplainRecommendation).mockResolvedValueOnce({
        explanation: '',
        error: '',
      });

      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        // Should show the reasoning as fallback (in the detailed explanation section)
        const reasoningElements = screen.getAllByText('Efficient removal spell that fits the aggro strategy');
        expect(reasoningElements.length).toBeGreaterThan(0);
      });
    });
  });

  describe('source display', () => {
    it('displays ML Model source', async () => {
      render(<RecommendationCard {...defaultProps} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('ML Model')).toBeInTheDocument();
      });
    });

    it('displays Metagame Data source', async () => {
      const metaRecommendation = { ...mockRecommendation, source: 'meta' };
      render(<RecommendationCard {...defaultProps} recommendation={metaRecommendation as any} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('Metagame Data')).toBeInTheDocument();
      });
    });

    it('displays Your Play History source', async () => {
      const personalRecommendation = { ...mockRecommendation, source: 'personal' };
      render(<RecommendationCard {...defaultProps} recommendation={personalRecommendation as any} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('Your Play History')).toBeInTheDocument();
      });
    });

    it('displays custom source value', async () => {
      const customRecommendation = { ...mockRecommendation, source: 'custom_source' };
      render(<RecommendationCard {...defaultProps} recommendation={customRecommendation as any} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('custom_source')).toBeInTheDocument();
      });
    });
  });

  describe('without factors', () => {
    it('does not show score breakdown when factors is null', async () => {
      const recWithoutFactors = { ...mockRecommendation, factors: null };
      render(<RecommendationCard {...defaultProps} recommendation={recWithoutFactors as any} />);

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(screen.getByText('Why This Card?')).toBeInTheDocument();
      });

      expect(screen.queryByText('Score Breakdown')).not.toBeInTheDocument();
    });
  });

  describe('without mana cost', () => {
    it('does not show mana cost when not provided', () => {
      const recWithoutMana = { ...mockRecommendation, manaCost: '' };
      render(<RecommendationCard {...defaultProps} recommendation={recWithoutMana as any} />);

      expect(screen.queryByText('{R}')).not.toBeInTheDocument();
    });
  });

  describe('CSS class states', () => {
    it('adds expanded class when expanded', async () => {
      const { container } = render(<RecommendationCard {...defaultProps} />);

      expect(container.querySelector('.recommendation-card.expanded')).not.toBeInTheDocument();

      fireEvent.click(screen.getByText('? Why'));

      await waitFor(() => {
        expect(container.querySelector('.recommendation-card.expanded')).toBeInTheDocument();
      });
    });

    it('adds active class to explain button when expanded', async () => {
      render(<RecommendationCard {...defaultProps} />);

      const button = screen.getByText('? Why');
      expect(button).not.toHaveClass('active');

      fireEvent.click(button);

      await waitFor(() => {
        const lessButton = screen.getByText('- Less');
        expect(lessButton).toHaveClass('active');
      });
    });
  });
});

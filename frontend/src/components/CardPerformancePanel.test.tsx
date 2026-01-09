import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import CardPerformancePanel from './CardPerformancePanel';
import * as decksApi from '@/services/api/decks';

// Mock the API module
vi.mock('@/services/api/decks', () => ({
  getCardPerformance: vi.fn(),
  getAllPerformanceRecommendations: vi.fn(),
}));

const mockAnalysis: decksApi.DeckPerformanceAnalysis = {
  deckId: 'deck-1',
  deckName: 'Test Deck',
  totalMatches: 10,
  totalGames: 15,
  overallWinRate: 0.6,
  cardPerformance: [
    {
      cardId: 1,
      cardName: 'Lightning Bolt',
      quantity: 4,
      gamesWithCard: 15,
      gamesDrawn: 10,
      gamesPlayed: 8,
      winRateWhenDrawn: 0.7,
      winRateWhenPlayed: 0.75,
      deckWinRate: 0.6,
      playRate: 0.8,
      winContribution: 0.1,
      impactScore: 0.5,
      confidenceLevel: 'high',
      sampleSize: 10,
      performanceGrade: 'excellent',
      avgTurnPlayed: 2.5,
      turnPlayedDist: {},
      mulliganedAway: 2,
      mulliganRate: 0.1,
    },
    {
      cardId: 2,
      cardName: 'Counterspell',
      quantity: 4,
      gamesWithCard: 15,
      gamesDrawn: 8,
      gamesPlayed: 5,
      winRateWhenDrawn: 0.5,
      winRateWhenPlayed: 0.6,
      deckWinRate: 0.6,
      playRate: 0.625,
      winContribution: -0.1,
      impactScore: -0.3,
      confidenceLevel: 'medium',
      sampleSize: 8,
      performanceGrade: 'poor',
      avgTurnPlayed: 3.5,
      turnPlayedDist: {},
      mulliganedAway: 1,
      mulliganRate: 0.05,
    },
  ],
  bestPerformers: ['Lightning Bolt'],
  worstPerformers: ['Counterspell'],
  analysisDate: '2024-01-01',
};

const mockRecommendations: decksApi.DeckRecommendationsResponse = {
  deckId: 'deck-1',
  deckName: 'Test Deck',
  currentWinRate: 0.6,
  projectedWinRate: 0.65,
  addRecommendations: [
    {
      type: 'add',
      cardId: 3,
      cardName: 'Path to Exile',
      reason: 'Similar decks have high win rates with this card',
      impactEstimate: 0.05,
      confidence: 'high',
      priority: 1,
      basedOnGames: 50,
    },
  ],
  removeRecommendations: [
    {
      type: 'remove',
      cardId: 2,
      cardName: 'Counterspell',
      reason: 'Underperforming compared to deck average',
      impactEstimate: -0.1,
      confidence: 'medium',
      priority: 1,
      basedOnGames: 8,
    },
  ],
  swapRecommendations: [],
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('CardPerformancePanel', () => {
  it('should display loading state initially', async () => {
    vi.mocked(decksApi.getCardPerformance).mockImplementation(
      () => new Promise(() => {}), // Never resolves
    );
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    expect(screen.getByText('Analyzing card performance...')).toBeInTheDocument();
  });

  it('should display error state when API fails', async () => {
    vi.mocked(decksApi.getCardPerformance).mockRejectedValue(new Error('API Error'));
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockRejectedValue(new Error('API Error'));

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('API Error')).toBeInTheDocument();
    });
  });

  it('should display card performance data when loaded', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Card Performance: Test Deck')).toBeInTheDocument();
    });

    // Check summary stats - use getAllByText since numbers may appear multiple places
    expect(screen.getAllByText('10').length).toBeGreaterThanOrEqual(1); // totalMatches
    expect(screen.getAllByText('15').length).toBeGreaterThanOrEqual(1); // totalGames
    // Win rate appears in summary
    const summaryStats = document.querySelectorAll('.summary-stat .stat-value');
    expect(summaryStats[2]).toHaveTextContent('60.0%'); // overallWinRate

    // Check card names in performance table
    expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
    expect(screen.getByText('Counterspell')).toBeInTheDocument();
  });

  it('should display performance grades for cards', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('excellent')).toBeInTheDocument();
      expect(screen.getByText('poor')).toBeInTheDocument();
    });
  });

  it('should display confidence levels for cards', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('high')).toBeInTheDocument();
      expect(screen.getByText('medium')).toBeInTheDocument();
    });
  });

  it('should switch to recommendations tab when clicked', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Card Performance: Test Deck')).toBeInTheDocument();
    });

    // Click on Recommendations tab
    const recsTab = screen.getByText('Recommendations');
    fireEvent.click(recsTab);

    // Should show recommendation content
    await waitFor(() => {
      expect(screen.getByText('Cards to Consider Removing')).toBeInTheDocument();
      expect(screen.getByText('Cards to Consider Adding')).toBeInTheDocument();
    });
  });

  it('should display recommendation details', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Card Performance: Test Deck')).toBeInTheDocument();
    });

    // Click on Recommendations tab
    const recsTab = screen.getByText('Recommendations');
    fireEvent.click(recsTab);

    await waitFor(() => {
      expect(screen.getByText('Path to Exile')).toBeInTheDocument();
      expect(screen.getByText('Similar decks have high win rates with this card')).toBeInTheDocument();
    });
  });

  it('should call onClose when close button is clicked', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);
    const onClose = vi.fn();

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={onClose} />);

    await waitFor(() => {
      expect(screen.getByText('Card Performance: Test Deck')).toBeInTheDocument();
    });

    const closeButton = screen.getByRole('button', { name: /×/ });
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalled();
  });

  it('should display empty state when no performance data', async () => {
    const emptyAnalysis: decksApi.DeckPerformanceAnalysis = {
      ...mockAnalysis,
      cardPerformance: [],
    };
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(emptyAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(null as unknown as decksApi.DeckRecommendationsResponse);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Not enough game data to analyze card performance.')).toBeInTheDocument();
    });
  });

  it('should display win contribution with correct formatting', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('+10.0%')).toBeInTheDocument(); // Lightning Bolt positive contribution
      expect(screen.getByText('-10.0%')).toBeInTheDocument(); // Counterspell negative contribution
    });
  });

  it('should sort cards by impact score by default', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      // First row is header, second should be Lightning Bolt (higher impact score)
      expect(rows[1]).toHaveTextContent('Lightning Bolt');
    });
  });

  it('should toggle sort direction when clicking same column', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Card Performance: Test Deck')).toBeInTheDocument();
    });

    // Click on Card column first time (sorts desc by name - Z to A)
    const cardHeader = screen.getByText('Card', { selector: 'th' });
    fireEvent.click(cardHeader);

    // Click again to toggle to ascending (A to Z)
    fireEvent.click(cardHeader);

    await waitFor(() => {
      const rows = screen.getAllByRole('row');
      // Counterspell comes before Lightning Bolt alphabetically (A-Z)
      expect(rows[1]).toHaveTextContent('Counterspell');
    });
  });

  it('should display projected win rate improvement in recommendations', async () => {
    vi.mocked(decksApi.getCardPerformance).mockResolvedValue(mockAnalysis);
    vi.mocked(decksApi.getAllPerformanceRecommendations).mockResolvedValue(mockRecommendations);

    render(<CardPerformancePanel deckId="deck-1" deckName="Test Deck" onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Card Performance: Test Deck')).toBeInTheDocument();
    });

    // Click on Recommendations tab
    const recsTab = screen.getByText('Recommendations');
    fireEvent.click(recsTab);

    await waitFor(() => {
      expect(screen.getByText('Current Win Rate')).toBeInTheDocument();
      expect(screen.getByText('Projected Win Rate')).toBeInTheDocument();
      expect(screen.getByText('65.0%')).toBeInTheDocument(); // projected win rate
    });
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import OpponentAnalysisPanel from './OpponentAnalysisPanel';
import { opponents } from '@/services/api';

// Mock the opponents API
vi.mock('@/services/api', () => ({
  opponents: {
    getOpponentAnalysis: vi.fn(),
    getDeckStyleDisplayName: vi.fn((style) => style || 'Unknown'),
    getConfidenceColorClass: vi.fn((conf) => (conf >= 0.7 ? 'text-green-400' : 'text-gray-400')),
    formatConfidence: vi.fn((conf) => `${Math.round(conf * 100)}%`),
    getPriorityColorClass: vi.fn((priority) => {
      const colors: Record<string, string> = {
        high: 'text-red-400',
        medium: 'text-yellow-400',
        low: 'text-blue-400',
      };
      return colors[priority] || 'text-gray-400';
    }),
    getCategoryDisplayName: vi.fn((cat) => cat || 'Unknown'),
  },
}));

const mockAnalysis = {
  profile: {
    id: 1,
    matchId: 'test-match-123',
    detectedArchetype: 'Mono Red Aggro',
    archetypeConfidence: 0.85,
    colorIdentity: 'R',
    deckStyle: 'aggro',
    cardsObserved: 12,
    estimatedDeckSize: 60,
    observedCardIds: '[12345, 12346]',
    inferredCardIds: null,
    signatureCards: '[12345]',
    format: 'Standard',
    metaArchetypeId: null,
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
  },
  observedCards: [
    {
      cardId: 12345,
      cardName: 'Lightning Bolt',
      zone: 'battlefield',
      turnFirstSeen: 2,
      timesSeen: 3,
      isSignature: true,
      category: 'removal',
    },
    {
      cardId: 12346,
      cardName: 'Monastery Swiftspear',
      zone: 'battlefield',
      turnFirstSeen: 1,
      timesSeen: 2,
      isSignature: false,
      category: 'threat',
    },
  ],
  expectedCards: [
    {
      cardId: 12347,
      cardName: 'Goblin Guide',
      inclusionRate: 0.95,
      avgCopies: 4.0,
      wasSeen: false,
      category: 'threat',
      playAround: 'Fast creature - expect early aggression',
    },
    {
      cardId: 12345,
      cardName: 'Lightning Bolt',
      inclusionRate: 0.99,
      avgCopies: 4.0,
      wasSeen: true,
      category: 'removal',
      playAround: '',
    },
  ],
  strategicInsights: [
    {
      type: 'archetype',
      description: 'Opponent is likely playing Mono Red Aggro',
      priority: 'high' as const,
      cards: [],
    },
    {
      type: 'strategy',
      description: 'Aggressive deck - prioritize early blockers',
      priority: 'high' as const,
      cards: [],
    },
  ],
  matchupStats: null,
  metaArchetype: null,
};

describe('OpponentAnalysisPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders collapsed state by default', () => {
    render(<OpponentAnalysisPanel matchId="test-match-123" />);

    expect(screen.getByText('Opponent Analysis')).toBeInTheDocument();
    expect(screen.queryByText('Archetype:')).not.toBeInTheDocument();
  });

  it('shows expand icon when collapsed', () => {
    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={false} />);

    // Should show right arrow when collapsed
    const expandIcon = screen.getByText('\u25B6');
    expect(expandIcon).toBeInTheDocument();
  });

  it('shows collapse icon when expanded', () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    // Should show down arrow when expanded
    const collapseIcon = screen.getByText('\u25BC');
    expect(collapseIcon).toBeInTheDocument();
  });

  it('calls onToggle when header is clicked', () => {
    const onToggle = vi.fn();
    render(<OpponentAnalysisPanel matchId="test-match-123" onToggle={onToggle} />);

    fireEvent.click(screen.getByText('Opponent Analysis'));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it('loads analysis when expanded', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(opponents.getOpponentAnalysis).toHaveBeenCalledWith('test-match-123');
    });
  });

  it('displays profile summary when analysis is loaded', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Archetype:')).toBeInTheDocument();
      expect(screen.getByText(/Mono Red Aggro/)).toBeInTheDocument();
    });
  });

  it('displays observed cards count in tab', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Observed (2)')).toBeInTheDocument();
    });
  });

  it('displays expected cards count in tab', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Expected (2)')).toBeInTheDocument();
    });
  });

  it('displays insights count in tab', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Insights (2)')).toBeInTheDocument();
    });
  });

  it('shows observed cards by default', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      expect(screen.getByText('Monastery Swiftspear')).toBeInTheDocument();
    });
  });

  it('shows signature badge for signature cards', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Signature')).toBeInTheDocument();
    });
  });

  it('switches to expected cards tab when clicked', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Expected (2)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Expected (2)'));

    await waitFor(() => {
      expect(screen.getByText('Goblin Guide')).toBeInTheDocument();
      expect(screen.getByText('95%')).toBeInTheDocument();
    });
  });

  it('switches to insights tab when clicked', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Insights (2)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Insights (2)'));

    await waitFor(() => {
      expect(screen.getByText('Opponent is likely playing Mono Red Aggro')).toBeInTheDocument();
    });
  });

  it('shows error message when API fails', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockRejectedValueOnce(new Error('API Error'));

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('API Error')).toBeInTheDocument();
    });
  });

  it('shows loading spinner while fetching', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve(mockAnalysis), 100))
    );

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    expect(screen.getByText('Analyzing opponent...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.queryByText('Analyzing opponent...')).not.toBeInTheDocument();
    });
  });

  it('displays cards to watch section for unseen expected cards', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Expected (2)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Expected (2)'));

    await waitFor(() => {
      expect(screen.getByText('Cards to Watch For (1)')).toBeInTheDocument();
    });
  });

  it('displays confirmed cards section for seen expected cards', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Expected (2)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Expected (2)'));

    await waitFor(() => {
      expect(screen.getByText('Confirmed Cards (1)')).toBeInTheDocument();
    });
  });

  it('renders color identity symbols', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Colors:')).toBeInTheDocument();
      // Should render R as a mana symbol
      const colorSymbol = screen.getByText('R');
      expect(colorSymbol).toBeInTheDocument();
    });
  });

  it('displays deck style from profile', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Style:')).toBeInTheDocument();
      expect(screen.getByText('aggro')).toBeInTheDocument();
    });
  });

  it('displays cards observed count', async () => {
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(mockAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('Cards Seen:')).toBeInTheDocument();
      expect(screen.getByText('12')).toBeInTheDocument();
    });
  });

  it('handles empty observed cards', async () => {
    const emptyAnalysis = {
      ...mockAnalysis,
      observedCards: [],
    };
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(emptyAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    await waitFor(() => {
      expect(screen.getByText('No cards observed during this match')).toBeInTheDocument();
    });
  });

  it('handles empty expected cards', async () => {
    const emptyAnalysis = {
      ...mockAnalysis,
      expectedCards: [],
    };
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(emptyAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    // Wait for loading to complete
    await waitFor(() => {
      expect(screen.getByText('Expected (0)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Expected (0)'));

    await waitFor(() => {
      expect(screen.getByText('No expected cards data available')).toBeInTheDocument();
    });
  });

  it('handles empty insights', async () => {
    const emptyAnalysis = {
      ...mockAnalysis,
      strategicInsights: [],
    };
    vi.mocked(opponents.getOpponentAnalysis).mockResolvedValueOnce(emptyAnalysis);

    render(<OpponentAnalysisPanel matchId="test-match-123" isExpanded={true} />);

    // Wait for loading to complete
    await waitFor(() => {
      expect(screen.getByText('Insights (0)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Insights (0)'));

    await waitFor(() => {
      expect(screen.getByText('No strategic insights available')).toBeInTheDocument();
    });
  });
});

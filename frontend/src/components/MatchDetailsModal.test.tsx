import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import MatchDetailsModal from './MatchDetailsModal';
import { models } from '@/types/models';

// Mock the API modules
vi.mock('@/services/api', () => ({
  matches: {
    getMatchGames: vi.fn(),
  },
}));

vi.mock('@/services/api/opponents', () => ({
  getOpponentAnalysis: vi.fn(),
  getDeckStyleDisplayName: vi.fn((style: string | null) => style || 'Unknown'),
  getConfidenceColorClass: vi.fn(() => ''),
  formatConfidence: vi.fn((confidence: number) => `${Math.round(confidence * 100)}%`),
  getPriorityColorClass: vi.fn(() => ''),
  getCategoryDisplayName: vi.fn((category: string) => category),
}));

vi.mock('@/services/api/gameplays', () => ({
  getMatchTimeline: vi.fn(),
}));

import { matches } from '@/services/api';
import * as gameplays from '@/services/api/gameplays';

const mockGetMatchGames = vi.mocked(matches.getMatchGames);
const mockGetMatchTimeline = vi.mocked(gameplays.getMatchTimeline);

const mockMatch: models.Match = {
  ID: 'match-123',
  EventName: 'Standard Event',
  Format: 'Standard',
  DeckID: 'deck-456',
  DeckName: 'Test Deck',
  Result: 'win',
  PlayerWins: 2,
  OpponentWins: 1,
  Timestamp: '2025-01-09T12:00:00Z',
  OpponentName: 'Opponent123',
  RankBefore: 'Gold 2',
  RankAfter: 'Gold 1',
};

const mockGames: models.Game[] = [
  {
    ID: 1,
    MatchID: 'match-123',
    GameNumber: 1,
    Result: 'win',
    DurationSeconds: 300,
    ResultReason: 'concede',
    Timestamp: '2025-01-09T12:00:00Z',
  },
  {
    ID: 2,
    MatchID: 'match-123',
    GameNumber: 2,
    Result: 'loss',
    DurationSeconds: 450,
    ResultReason: 'normal',
    Timestamp: '2025-01-09T12:10:00Z',
  },
];

describe('MatchDetailsModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetMatchGames.mockResolvedValue(mockGames);
    mockGetMatchTimeline.mockResolvedValue([]);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('rendering', () => {
    it('renders match summary', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByText('Match Details')).toBeInTheDocument();
        expect(screen.getByText('Standard')).toBeInTheDocument();
        expect(screen.getByText('Standard Event')).toBeInTheDocument();
        expect(screen.getByText('WIN 2-1')).toBeInTheDocument();
      });
    });

    it('renders opponent name', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByText('Opponent123')).toBeInTheDocument();
      });
    });

    it('renders rank change', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByText('Gold 2 → Gold 1')).toBeInTheDocument();
      });
    });

    it('renders games breakdown', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByText('Game Breakdown')).toBeInTheDocument();
        expect(screen.getByText('Game 1')).toBeInTheDocument();
        expect(screen.getByText('Game 2')).toBeInTheDocument();
      });
    });

    it('renders Opponent Analysis panel header', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Opponent Analysis/i })).toBeInTheDocument();
      });
    });

    it('renders Game Timeline panel header', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Game Timeline/i })).toBeInTheDocument();
      });
    });
  });

  describe('close behavior', () => {
    it('calls onClose when close button is clicked', async () => {
      const onClose = vi.fn();
      render(<MatchDetailsModal match={mockMatch} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Match Details')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('×'));

      expect(onClose).toHaveBeenCalled();
    });

    it('calls onClose when Close button is clicked', async () => {
      const onClose = vi.fn();
      render(<MatchDetailsModal match={mockMatch} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Close')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Close'));

      expect(onClose).toHaveBeenCalled();
    });

    it('calls onClose when Escape key is pressed', async () => {
      const onClose = vi.fn();
      render(<MatchDetailsModal match={mockMatch} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Match Details')).toBeInTheDocument();
      });

      fireEvent.keyDown(document, { key: 'Escape' });

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe('Game Timeline panel', () => {
    it('expands Game Timeline panel when clicked', async () => {
      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /Game Timeline/i })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: /Game Timeline/i }));

      await waitFor(() => {
        expect(mockGetMatchTimeline).toHaveBeenCalledWith('match-123');
      });
    });
  });

  describe('loading state', () => {
    it('shows loading spinner while fetching games', async () => {
      mockGetMatchGames.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockGames), 100))
      );

      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      expect(screen.getByText(/Loading games/i)).toBeInTheDocument();
    });
  });

  describe('error state', () => {
    it('shows error message when games fail to load', async () => {
      mockGetMatchGames.mockRejectedValue(new Error('Network error'));

      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByText(/Failed to load games: Network error/i)).toBeInTheDocument();
      });
    });
  });

  describe('empty state', () => {
    it('shows message when no games available', async () => {
      mockGetMatchGames.mockResolvedValue([]);

      render(<MatchDetailsModal match={mockMatch} onClose={() => {}} />);

      await waitFor(() => {
        expect(screen.getByText(/No game data available/i)).toBeInTheDocument();
      });
    });
  });
});

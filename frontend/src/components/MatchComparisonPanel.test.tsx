import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import MatchComparisonPanel from './MatchComparisonPanel';
import type { ComparisonResult } from '@/services/api/matches';
import * as matchesApi from '@/services/api/matches';

vi.mock('@/services/api/matches', () => ({
  compareFormats: vi.fn(),
  compareDecks: vi.fn(),
  compareTimePeriods: vi.fn(),
}));

const mockCompareFormats = vi.mocked(matchesApi.compareFormats);
const mockCompareDecks = vi.mocked(matchesApi.compareDecks);
const mockCompareTimePeriods = vi.mocked(matchesApi.compareTimePeriods);

// Mock filter with required fields
const mockFilter = { Formats: [], EventNames: [] };

const mockComparisonResult: ComparisonResult = {
  Groups: [
    {
      Label: 'Standard',
      Filter: mockFilter,
      Statistics: {
        TotalMatches: 50,
        MatchesWon: 30,
        MatchesLost: 20,
        TotalGames: 100,
        GamesWon: 60,
        GamesLost: 40,
        WinRate: 0.6,
        GameWinRate: 0.6,
      },
      MatchCount: 50,
    },
    {
      Label: 'Historic',
      Filter: mockFilter,
      Statistics: {
        TotalMatches: 30,
        MatchesWon: 15,
        MatchesLost: 15,
        TotalGames: 60,
        GamesWon: 30,
        GamesLost: 30,
        WinRate: 0.5,
        GameWinRate: 0.5,
      },
      MatchCount: 30,
    },
  ],
  BestGroup: {
    Label: 'Standard',
    Filter: mockFilter,
    Statistics: {
      TotalMatches: 50,
      MatchesWon: 30,
      MatchesLost: 20,
      TotalGames: 100,
      GamesWon: 60,
      GamesLost: 40,
      WinRate: 0.6,
      GameWinRate: 0.6,
    },
    MatchCount: 50,
  },
  WorstGroup: {
    Label: 'Historic',
    Filter: mockFilter,
    Statistics: {
      TotalMatches: 30,
      MatchesWon: 15,
      MatchesLost: 15,
      TotalGames: 60,
      GamesWon: 30,
      GamesLost: 30,
      WinRate: 0.5,
      GameWinRate: 0.5,
    },
    MatchCount: 30,
  },
  WinRateDiff: 0.1,
  TotalMatches: 80,
  ComparisonDate: '2025-01-09T12:00:00Z',
};

describe('MatchComparisonPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCompareFormats.mockResolvedValue(mockComparisonResult);
    mockCompareDecks.mockResolvedValue(mockComparisonResult);
    mockCompareTimePeriods.mockResolvedValue(mockComparisonResult);
  });

  describe('rendering', () => {
    it('renders the panel header', () => {
      render(<MatchComparisonPanel />);
      expect(screen.getByText('Match Comparison')).toBeInTheDocument();
    });

    it('renders comparison type buttons', () => {
      render(<MatchComparisonPanel />);
      expect(screen.getByText('Compare Formats')).toBeInTheDocument();
      expect(screen.getByText('Compare Decks')).toBeInTheDocument();
      expect(screen.getByText('Compare Time Periods')).toBeInTheDocument();
    });

    it('renders formats selector by default', () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);
      expect(screen.getByText('Select Formats to Compare')).toBeInTheDocument();
    });

    it('renders format options when formats provided', () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);
      expect(screen.getByText('Standard')).toBeInTheDocument();
      expect(screen.getByText('Historic')).toBeInTheDocument();
    });

    it('renders empty message when no formats', () => {
      render(<MatchComparisonPanel formats={[]} />);
      expect(screen.getByText('No formats available. Play some matches first.')).toBeInTheDocument();
    });

    it('renders close button when onClose provided', () => {
      const onClose = vi.fn();
      render(<MatchComparisonPanel onClose={onClose} />);
      expect(screen.getByText('×')).toBeInTheDocument();
    });
  });

  describe('comparison type selection', () => {
    it('switches to decks selector', () => {
      render(
        <MatchComparisonPanel
          deckIds={[
            { id: 'deck-1', name: 'Izzet Phoenix' },
            { id: 'deck-2', name: 'Mono White' },
          ]}
        />
      );
      fireEvent.click(screen.getByText('Compare Decks'));
      expect(screen.getByText('Select Decks to Compare')).toBeInTheDocument();
    });

    it('switches to time periods selector', () => {
      render(<MatchComparisonPanel />);
      fireEvent.click(screen.getByText('Compare Time Periods'));
      expect(screen.getByText('Select Time Periods to Compare')).toBeInTheDocument();
    });

    it('renders deck options when decks provided', () => {
      render(
        <MatchComparisonPanel
          deckIds={[
            { id: 'deck-1', name: 'Izzet Phoenix' },
            { id: 'deck-2', name: 'Mono White' },
          ]}
        />
      );
      fireEvent.click(screen.getByText('Compare Decks'));
      expect(screen.getByText('Izzet Phoenix')).toBeInTheDocument();
      expect(screen.getByText('Mono White')).toBeInTheDocument();
    });

    it('renders time period options', () => {
      render(<MatchComparisonPanel />);
      fireEvent.click(screen.getByText('Compare Time Periods'));
      expect(screen.getByText('Last 7 Days')).toBeInTheDocument();
      expect(screen.getByText('Last 30 Days')).toBeInTheDocument();
    });
  });

  describe('selection behavior', () => {
    it('toggles format selection', () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);
      const checkboxes = screen.getAllByRole('checkbox');
      const standardCheckbox = checkboxes[0];

      fireEvent.click(standardCheckbox);
      expect(standardCheckbox).toBeChecked();

      fireEvent.click(standardCheckbox);
      expect(standardCheckbox).not.toBeChecked();
    });
  });

  describe('comparison execution', () => {
    it('shows error when less than 2 formats selected', async () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);

      // Select only one format
      const checkboxes = screen.getAllByRole('checkbox');
      fireEvent.click(checkboxes[0]);

      // Click compare
      fireEvent.click(screen.getByText('Compare'));

      await waitFor(() => {
        expect(screen.getByText('Please select at least 2 formats to compare')).toBeInTheDocument();
      });
    });

    it('calls compareFormats API with selected formats', async () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic', 'Explorer']} />);

      // Select formats
      const checkboxes = screen.getAllByRole('checkbox');
      fireEvent.click(checkboxes[0]); // Standard
      fireEvent.click(checkboxes[1]); // Historic

      // Click compare
      fireEvent.click(screen.getByText('Compare'));

      await waitFor(() => {
        expect(mockCompareFormats).toHaveBeenCalledWith({
          formats: ['Standard', 'Historic'],
          base_filter: {},
        });
      });
    });

    it('displays comparison results', async () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);

      // Select formats
      const checkboxes = screen.getAllByRole('checkbox');
      fireEvent.click(checkboxes[0]); // Standard
      fireEvent.click(checkboxes[1]); // Historic

      // Click compare
      fireEvent.click(screen.getByText('Compare'));

      await waitFor(() => {
        expect(screen.getByText('Comparison Results')).toBeInTheDocument();
        expect(screen.getByText('Total Matches: 80')).toBeInTheDocument();
        // Win rates appear twice (WinRate and GameWinRate columns)
        expect(screen.getAllByText('60.0%').length).toBeGreaterThan(0);
        expect(screen.getAllByText('50.0%').length).toBeGreaterThan(0);
      });
    });

    it('shows best and worst badges', async () => {
      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);

      // Select formats
      const checkboxes = screen.getAllByRole('checkbox');
      fireEvent.click(checkboxes[0]); // Standard
      fireEvent.click(checkboxes[1]); // Historic

      // Click compare
      fireEvent.click(screen.getByText('Compare'));

      await waitFor(() => {
        expect(screen.getByText('Best')).toBeInTheDocument();
        expect(screen.getByText('Worst')).toBeInTheDocument();
      });
    });

    it('handles API errors gracefully', async () => {
      mockCompareFormats.mockRejectedValue(new Error('Network error'));

      render(<MatchComparisonPanel formats={['Standard', 'Historic']} />);

      // Select formats
      const checkboxes = screen.getAllByRole('checkbox');
      fireEvent.click(checkboxes[0]); // Standard
      fireEvent.click(checkboxes[1]); // Historic

      // Click compare
      fireEvent.click(screen.getByText('Compare'));

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });
    });
  });

  describe('close behavior', () => {
    it('calls onClose when close button clicked', () => {
      const onClose = vi.fn();
      render(<MatchComparisonPanel onClose={onClose} />);

      fireEvent.click(screen.getByText('×'));

      expect(onClose).toHaveBeenCalled();
    });
  });
});

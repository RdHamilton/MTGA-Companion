import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import ResultBreakdown from './ResultBreakdown';
import { mockWailsApp } from '@/test/mocks/apiMock';
import { AppProvider } from '../context/AppContext';
import { models } from '@/types/models';

// Helper function to create mock statistics
function createMockStatistics(overrides: Partial<models.Statistics> = {}): models.Statistics {
  return new models.Statistics({
    TotalMatches: 100,
    MatchesWon: 60,
    MatchesLost: 40,
    TotalGames: 220,
    GamesWon: 130,
    GamesLost: 90,
    WinRate: 0.6,
    GameWinRate: 0.591,
    ...overrides,
  });
}

// Wrapper component with AppProvider
function renderWithProvider(ui: React.ReactElement) {
  return render(<AppProvider>{ui}</AppProvider>);
}

// Helper to get select by finding the label then the next select sibling
function getSelectByLabel(labelText: string): HTMLSelectElement {
  const label = screen.getByText(labelText);
  const filterGroup = label.closest('.filter-group');
  return filterGroup?.querySelector('select') as HTMLSelectElement;
}

// Helper to get input by finding the label
function getInputByLabel(labelText: string): HTMLInputElement {
  const label = screen.getByText(labelText);
  const filterGroup = label.closest('.filter-group');
  return filterGroup?.querySelector('input') as HTMLInputElement;
}

describe('ResultBreakdown', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolvePromise: (value: models.Statistics) => void;
      const loadingPromise = new Promise<models.Statistics>((resolve) => {
        resolvePromise = resolve;
      });
      mockWailsApp.GetStats.mockReturnValue(loadingPromise);

      renderWithProvider(<ResultBreakdown />);

      expect(screen.getByText('Loading performance metrics...')).toBeInTheDocument();

      resolvePromise!(createMockStatistics());
      await waitFor(() => {
        expect(screen.queryByText('Loading performance metrics...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockWailsApp.GetStats.mockRejectedValue(new Error('Server error'));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load performance metrics' })).toBeInTheDocument();
      });
      expect(screen.getByText('Server error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockWailsApp.GetStats.mockRejectedValue('Unknown error');

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load performance metrics' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no metrics data', async () => {
      mockWailsApp.GetStats.mockResolvedValue(null);

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('No performance data')).toBeInTheDocument();
      });
      expect(
        screen.getByText('Play some matches to see your detailed performance breakdown.')
      ).toBeInTheDocument();
    });
  });

  describe('Data Display', () => {
    it('should display overall performance section', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Overall Performance')).toBeInTheDocument();
      });
    });

    it('should display game-level statistics section', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Game-Level Statistics')).toBeInTheDocument();
      });
    });

    it('should display performance analysis section', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Performance Analysis')).toBeInTheDocument();
      });
    });

    it('should display win/loss breakdown section', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Win/Loss Breakdown')).toBeInTheDocument();
      });
    });

    it('should display overall win rate correctly', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.6 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('60%')).toBeInTheDocument();
      });
    });

    it('should display total matches', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ TotalMatches: 100 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('100')).toBeInTheDocument();
      });
    });

    it('should display matches won and lost', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ MatchesWon: 60, MatchesLost: 40 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('60W - 40L')).toBeInTheDocument();
      });
    });

    it('should display game win rate', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ GameWinRate: 0.591 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('59.1%')).toBeInTheDocument();
      });
    });

    it('should display total games', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ TotalGames: 220 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('220')).toBeInTheDocument();
      });
    });

    it('should display games won and lost', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ GamesWon: 130, GamesLost: 90 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('130W - 90L')).toBeInTheDocument();
      });
    });
  });

  describe('Performance Analysis', () => {
    it('should calculate average games per match', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ TotalMatches: 100, TotalGames: 220 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('2.20')).toBeInTheDocument();
      });
    });

    it('should handle zero matches for average calculation', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ TotalMatches: 0, TotalGames: 0 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('0.00')).toBeInTheDocument();
      });
    });

    it('should calculate match to game win rate ratio', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ WinRate: 0.6, GameWinRate: 0.5 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('1.20')).toBeInTheDocument();
      });
    });

    it('should handle zero game win rate for ratio calculation', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ WinRate: 0.6, GameWinRate: 0 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        const ratioElements = screen.getAllByText('0.00');
        expect(ratioElements.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Performance Categories', () => {
    it('should display "Excellent" for win rate >= 55%', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.55 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Excellent')).toBeInTheDocument();
      });
    });

    it('should display "Good" for win rate >= 50% and < 55%', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.52 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Good')).toBeInTheDocument();
      });
    });

    it('should display "Average" for win rate >= 45% and < 50%', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.47 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Average')).toBeInTheDocument();
      });
    });

    it('should display "Below Average" for win rate < 45%', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.40 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Below Average')).toBeInTheDocument();
      });
    });
  });

  describe('Filters', () => {
    it('should render date range filter with default value', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range') as HTMLSelectElement;
        expect(dateRangeSelect.value).toBe('7days');
      });
    });

    it('should render format filter with default value', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        const formatSelect = getSelectByLabel('Format') as HTMLSelectElement;
        expect(formatSelect.value).toBe('all');
      });
    });

    it('should show custom date inputs when custom range selected', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Overall Performance')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'custom' } });

      await waitFor(() => {
        expect(getInputByLabel('Start Date')).toBeInTheDocument();
        expect(getInputByLabel('End Date')).toBeInTheDocument();
      });
    });

    it('should refetch data when date range changes', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(mockWailsApp.GetStats).toHaveBeenCalled();
      });

      const initialCallCount = mockWailsApp.GetStats.mock.calls.length;

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(mockWailsApp.GetStats.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should refetch data when format changes', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(mockWailsApp.GetStats).toHaveBeenCalled();
      });

      const initialCallCount = mockWailsApp.GetStats.mock.calls.length;

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        expect(mockWailsApp.GetStats.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });
  });

  describe('Page Header', () => {
    it('should display page title', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Result Breakdown');
      });
    });
  });

  describe('API Filter Parameters', () => {
    it('should pass constructed formats for constructed filter', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Overall Performance')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'constructed' } });

      await waitFor(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const lastCall = mockWailsApp.GetStats.mock.calls.slice(-1)[0] as any[];
        const filter = lastCall[0] as models.StatsFilter;
        expect(filter.Formats).toEqual(['Ladder', 'Play']);
      });
    });

    it('should pass single format for specific format filter', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Overall Performance')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const lastCall = mockWailsApp.GetStats.mock.calls.slice(-1)[0] as any[];
        const filter = lastCall[0] as models.StatsFilter;
        expect(filter.Format).toBe('Ladder');
      });
    });
  });

  describe('Win/Loss Bar Display', () => {
    it('should display win percentage in breakdown bar', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.6 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('60% Wins')).toBeInTheDocument();
      });
    });

    it('should display loss percentage in breakdown bar', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics({ WinRate: 0.6 }));

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('40% Losses')).toBeInTheDocument();
      });
    });

    it('should display matches won and lost in breakdown stats', async () => {
      mockWailsApp.GetStats.mockResolvedValue(
        createMockStatistics({ MatchesWon: 60, MatchesLost: 40 })
      );

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('60 Matches Won')).toBeInTheDocument();
        expect(screen.getByText('40 Matches Lost')).toBeInTheDocument();
      });
    });
  });

  describe('Metric Labels', () => {
    it('should display all metric labels', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Overall Win Rate')).toBeInTheDocument();
        expect(screen.getByText('Total Matches')).toBeInTheDocument();
        expect(screen.getByText('Matches Won')).toBeInTheDocument();
        expect(screen.getByText('Matches Lost')).toBeInTheDocument();
        expect(screen.getByText('Game Win Rate')).toBeInTheDocument();
        expect(screen.getByText('Total Games')).toBeInTheDocument();
        expect(screen.getByText('Games Won')).toBeInTheDocument();
        expect(screen.getByText('Games Lost')).toBeInTheDocument();
      });
    });

    it('should display analysis labels', async () => {
      mockWailsApp.GetStats.mockResolvedValue(createMockStatistics());

      renderWithProvider(<ResultBreakdown />);

      await waitFor(() => {
        expect(screen.getByText('Average Games per Match')).toBeInTheDocument();
        expect(screen.getByText('Match to Game Win Rate Ratio')).toBeInTheDocument();
        expect(screen.getByText('Performance Category')).toBeInTheDocument();
      });
    });
  });
});

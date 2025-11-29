import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import FormatDistribution from './FormatDistribution';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { AppProvider } from '../context/AppContext';
import { models } from '../../wailsjs/go/models';

// Mock Recharts to avoid rendering issues in tests
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  PieChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="pie-chart">{children}</div>
  ),
  Pie: ({ children }: { children: React.ReactNode }) => <div data-testid="pie">{children}</div>,
  Cell: () => <div data-testid="cell" />,
  BarChart: ({ children, data }: { children: React.ReactNode; data: unknown[] }) => (
    <div data-testid="bar-chart" data-chart-data={JSON.stringify(data)}>
      {children}
    </div>
  ),
  Bar: () => <div data-testid="bar" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
}));

// Helper function to create mock statistics
function createMockStatistics(overrides: Partial<models.Statistics> = {}): models.Statistics {
  return new models.Statistics({
    TotalMatches: 20,
    MatchesWon: 12,
    MatchesLost: 8,
    TotalGames: 45,
    GamesWon: 27,
    GamesLost: 18,
    WinRate: 0.6,
    GameWinRate: 0.6,
    ...overrides,
  });
}

// Helper function to create mock format stats response
function createMockFormatStatsResponse(): Record<string, models.Statistics> {
  return {
    Ladder: createMockStatistics({ WinRate: 0.65, TotalMatches: 50, MatchesWon: 33, MatchesLost: 17 }),
    Play: createMockStatistics({ WinRate: 0.55, TotalMatches: 30, MatchesWon: 17, MatchesLost: 13 }),
    Draft: createMockStatistics({ WinRate: 0.48, TotalMatches: 20, MatchesWon: 10, MatchesLost: 10 }),
  };
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

describe('FormatDistribution', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolvePromise: (value: Record<string, models.Statistics>) => void;
      const loadingPromise = new Promise<Record<string, models.Statistics>>((resolve) => {
        resolvePromise = resolve;
      });
      mockWailsApp.GetStatsByFormat.mockReturnValue(loadingPromise);

      renderWithProvider(<FormatDistribution />);

      expect(screen.getByText('Loading format statistics...')).toBeInTheDocument();

      resolvePromise!(createMockFormatStatsResponse());
      await waitFor(() => {
        expect(screen.queryByText('Loading format statistics...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockWailsApp.GetStatsByFormat.mockRejectedValue(new Error('Database unavailable'));

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load format statistics' })).toBeInTheDocument();
      });
      expect(screen.getByText('Database unavailable')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockWailsApp.GetStatsByFormat.mockRejectedValue('Unknown error');

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load format statistics' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no format data', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue({});

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('No format data')).toBeInTheDocument();
      });
      expect(
        screen.getByText('Play matches in different formats to see your format distribution.')
      ).toBeInTheDocument();
    });

    it('should show empty state when API returns null', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(null);

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('No format data')).toBeInTheDocument();
      });
    });
  });

  describe('Data Display', () => {
    it('should render bar chart by default', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByTestId('bar-chart')).toBeInTheDocument();
      });
      expect(screen.queryByTestId('pie-chart')).not.toBeInTheDocument();
    });

    it('should render format cards with statistics', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('Ladder')).toBeInTheDocument();
      });
      expect(screen.getByText('Play')).toBeInTheDocument();
      expect(screen.getByText('Draft')).toBeInTheDocument();
    });

    it('should display win rates correctly', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('65%')).toBeInTheDocument();
      });
      expect(screen.getByText('55%')).toBeInTheDocument();
      expect(screen.getByText('48%')).toBeInTheDocument();
    });

    it('should display format count and total matches', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('3 formats • 100 total matches')).toBeInTheDocument();
      });
    });

    it('should display singular format count for one format', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue({
        Ladder: createMockStatistics({ TotalMatches: 25 }),
      });

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('1 format • 25 total matches')).toBeInTheDocument();
      });
    });

    it('should display wins and losses', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue({
        Ladder: createMockStatistics({ MatchesWon: 20, MatchesLost: 5 }),
      });

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('20W - 5L')).toBeInTheDocument();
      });
    });

    it('should aggregate formats with underscore suffixes', async () => {
      // Two QuickDraft formats from different sets should be combined
      mockWailsApp.GetStatsByFormat.mockResolvedValue({
        'QuickDraft_TLA_20251127': createMockStatistics({ TotalMatches: 8, MatchesWon: 5, MatchesLost: 3 }),
        'QuickDraft_MKM_20241120': createMockStatistics({ TotalMatches: 12, MatchesWon: 7, MatchesLost: 5 }),
        Play: createMockStatistics({ TotalMatches: 10, MatchesWon: 6, MatchesLost: 4 }),
      });

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        // Should show normalized format name "QuickDraft" (not the full names with suffixes)
        expect(screen.getByText('QuickDraft')).toBeInTheDocument();
      });
      // Should NOT show the full format names with date suffixes
      expect(screen.queryByText('QuickDraft_TLA_20251127')).not.toBeInTheDocument();
      expect(screen.queryByText('QuickDraft_MKM_20241120')).not.toBeInTheDocument();
      // Should show Play as is (no underscore)
      expect(screen.getByText('Play')).toBeInTheDocument();
    });

    it('should combine stats when aggregating formats', async () => {
      // Two PremierDraft formats should have their wins/losses combined
      mockWailsApp.GetStatsByFormat.mockResolvedValue({
        'PremierDraft_TLA_20251127': createMockStatistics({ TotalMatches: 5, MatchesWon: 3, MatchesLost: 2 }),
        'PremierDraft_MKM_20241120': createMockStatistics({ TotalMatches: 5, MatchesWon: 2, MatchesLost: 3 }),
      });

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('PremierDraft')).toBeInTheDocument();
      });
      // Combined: 10 matches total, 5W - 5L
      expect(screen.getByText('5W - 5L')).toBeInTheDocument();
    });
  });

  describe('Filters', () => {
    it('should render date range filter with default value', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range') as HTMLSelectElement;
        expect(dateRangeSelect.value).toBe('7days');
      });
    });

    it('should render chart type filter', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(getSelectByLabel('Chart Type')).toBeInTheDocument();
      });
    });

    it('should render sort by filter', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        const sortBySelect = getSelectByLabel('Sort By') as HTMLSelectElement;
        expect(sortBySelect.value).toBe('matches');
      });
    });

    it('should render sort order filter', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        const sortOrderSelect = getSelectByLabel('Sort Order') as HTMLSelectElement;
        expect(sortOrderSelect.value).toBe('desc');
      });
    });

    it('should show custom date inputs when custom range selected', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('Ladder')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'custom' } });

      await waitFor(() => {
        expect(getInputByLabel('Start Date')).toBeInTheDocument();
        expect(getInputByLabel('End Date')).toBeInTheDocument();
      });
    });

    it('should switch to pie chart when chart type changes', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByTestId('bar-chart')).toBeInTheDocument();
      });

      const chartTypeSelect = getSelectByLabel('Chart Type');
      fireEvent.change(chartTypeSelect, { target: { value: 'pie' } });

      await waitFor(() => {
        expect(screen.getByTestId('pie-chart')).toBeInTheDocument();
      });
      expect(screen.queryByTestId('bar-chart')).not.toBeInTheDocument();
    });

    it('should refetch data when date range changes', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(mockWailsApp.GetStatsByFormat).toHaveBeenCalled();
      });

      const initialCallCount = mockWailsApp.GetStatsByFormat.mock.calls.length;

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(mockWailsApp.GetStatsByFormat.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });
  });

  describe('Sorting', () => {
    it('should sort by match count descending by default', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        const formatCards = screen.getAllByRole('heading', { level: 3 });
        expect(formatCards[0]).toHaveTextContent('Ladder');
        expect(formatCards[1]).toHaveTextContent('Play');
        expect(formatCards[2]).toHaveTextContent('Draft');
      });
    });

    it('should sort by win rate when selected', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('Ladder')).toBeInTheDocument();
      });

      const sortBySelect = getSelectByLabel('Sort By');
      fireEvent.change(sortBySelect, { target: { value: 'winRate' } });

      await waitFor(() => {
        const formatCards = screen.getAllByRole('heading', { level: 3 });
        expect(formatCards[0]).toHaveTextContent('Ladder'); // 65%
        expect(formatCards[1]).toHaveTextContent('Play'); // 55%
        expect(formatCards[2]).toHaveTextContent('Draft'); // 48%
      });
    });

    it('should sort by format name when selected', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('Ladder')).toBeInTheDocument();
      });

      const sortBySelect = getSelectByLabel('Sort By');
      const sortOrderSelect = getSelectByLabel('Sort Order');

      fireEvent.change(sortBySelect, { target: { value: 'name' } });
      fireEvent.change(sortOrderSelect, { target: { value: 'asc' } });

      await waitFor(() => {
        const formatCards = screen.getAllByRole('heading', { level: 3 });
        expect(formatCards[0]).toHaveTextContent('Draft');
        expect(formatCards[1]).toHaveTextContent('Ladder');
        expect(formatCards[2]).toHaveTextContent('Play');
      });
    });

    it('should sort ascending when order changed', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('Ladder')).toBeInTheDocument();
      });

      const sortOrderSelect = getSelectByLabel('Sort Order');
      fireEvent.change(sortOrderSelect, { target: { value: 'asc' } });

      await waitFor(() => {
        const formatCards = screen.getAllByRole('heading', { level: 3 });
        expect(formatCards[0]).toHaveTextContent('Draft'); // 20 matches
        expect(formatCards[2]).toHaveTextContent('Ladder'); // 50 matches
      });
    });
  });

  describe('Page Header', () => {
    it('should display page title', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Format Distribution');
      });
    });
  });

  describe('Unknown Format Handling', () => {
    it('should display "Unknown Format" for empty format name', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue({
        '': createMockStatistics(),
      });

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(screen.getByText('Unknown Format')).toBeInTheDocument();
      });
    });
  });

  describe('Chart Data Transformation', () => {
    it('should transform data correctly for bar chart', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        const barChart = screen.getByTestId('bar-chart');
        const chartData = JSON.parse(barChart.getAttribute('data-chart-data') || '[]');
        expect(chartData).toHaveLength(3);
        expect(chartData[0]).toHaveProperty('name');
        expect(chartData[0]).toHaveProperty('matches');
        expect(chartData[0]).toHaveProperty('winRate');
      });
    });
  });

  describe('API Calls', () => {
    it('should call GetStatsByFormat with filter', async () => {
      mockWailsApp.GetStatsByFormat.mockResolvedValue(createMockFormatStatsResponse());

      renderWithProvider(<FormatDistribution />);

      await waitFor(() => {
        expect(mockWailsApp.GetStatsByFormat).toHaveBeenCalled();
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const call = mockWailsApp.GetStatsByFormat.mock.calls[0] as any[];
      expect(call[0]).toBeInstanceOf(models.StatsFilter);
    });
  });
});

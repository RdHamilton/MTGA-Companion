import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import WinRateTrend from './WinRateTrend';
import { mockWailsApp } from '@/test/mocks/apiMock';
import { AppProvider } from '../context/AppContext';
// storage types imported but using any for mock flexibility

// Mock Recharts to avoid rendering issues in tests
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  LineChart: ({ children, data }: { children: React.ReactNode; data: unknown[] }) => (
    <div data-testid="line-chart" data-chart-data={JSON.stringify(data)}>
      {children}
    </div>
  ),
  BarChart: ({ children, data }: { children: React.ReactNode; data: unknown[] }) => (
    <div data-testid="bar-chart" data-chart-data={JSON.stringify(data)}>
      {children}
    </div>
  ),
  Line: () => <div data-testid="line" />,
  Bar: () => <div data-testid="bar" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
}));

// Helper function to create mock trend analysis data
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function createMockTrendAnalysis(overrides: Record<string, unknown> = {}): any {
  return {
    Periods: [
      {
        Period: {
          Label: 'Day 1',
          StartDate: '2024-01-01',
          EndDate: '2024-01-01',
        },
        Stats: {
          TotalMatches: 10,
          MatchesWon: 6,
          MatchesLost: 4,
          TotalGames: 20,
          GamesWon: 12,
          GamesLost: 8,
          WinRate: 0.6,
        },
        WinRate: 0.6,
      },
      {
        Period: {
          Label: 'Day 2',
          StartDate: '2024-01-02',
          EndDate: '2024-01-02',
        },
        Stats: {
          TotalMatches: 8,
          MatchesWon: 5,
          MatchesLost: 3,
          TotalGames: 16,
          GamesWon: 10,
          GamesLost: 6,
          WinRate: 0.625,
        },
        WinRate: 0.625,
      },
    ],
    Overall: {
      TotalMatches: 18,
      MatchesWon: 11,
      MatchesLost: 7,
      TotalGames: 36,
      GamesWon: 22,
      GamesLost: 14,
      WinRate: 0.611,
    },
    Trend: 'improving',
    TrendValue: 0.025,
    ...overrides,
  };
}

// Wrapper component with AppProvider
function renderWithProvider(ui: React.ReactElement) {
  return render(<AppProvider>{ui}</AppProvider>);
}

describe('WinRateTrend', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      // Create a promise that won't resolve immediately
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let resolvePromise: (value: any) => void;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const loadingPromise = new Promise<any>((resolve) => {
        resolvePromise = resolve;
      });
      mockWailsApp.GetTrendAnalysis.mockReturnValue(loadingPromise);

      renderWithProvider(<WinRateTrend />);

      expect(screen.getByText('Loading trend data...')).toBeInTheDocument();

      // Resolve to clean up
      resolvePromise!(createMockTrendAnalysis());
      await waitFor(() => {
        expect(screen.queryByText('Loading trend data...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockWailsApp.GetTrendAnalysis.mockRejectedValue(new Error('Network error'));

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('Failed to load trend data')).toBeInTheDocument();
      });
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockWailsApp.GetTrendAnalysis.mockRejectedValue('Unknown error');

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load trend data' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no analysis data', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(null);

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('Not enough data')).toBeInTheDocument();
      });
      expect(
        screen.getByText('Play at least 5 matches to see your win rate trends over time.')
      ).toBeInTheDocument();
    });

    it('should show empty state when periods array is empty', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue({
        Periods: [],
        Overall: null,
        Trend: '',
        TrendValue: 0,
      });

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('Not enough data')).toBeInTheDocument();
      });
    });
  });

  describe('Data Display', () => {
    it('should render line chart by default', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });
      expect(screen.queryByTestId('bar-chart')).not.toBeInTheDocument();
    });

    it('should display summary information', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('Win Rate Trend Analysis')).toBeInTheDocument();
      });
      expect(screen.getByText('Period:')).toBeInTheDocument();
      expect(screen.getByText('Format:')).toBeInTheDocument();
      expect(screen.getByText('Trend:')).toBeInTheDocument();
      expect(screen.getByText('Overall Win Rate:')).toBeInTheDocument();
    });

    it('should display trend with correct value', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText(/improving/)).toBeInTheDocument();
      });
    });

    it('should display overall win rate', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText(/61\.1%/)).toBeInTheDocument();
        expect(screen.getByText(/18 matches/)).toBeInTheDocument();
      });
    });

    it('should transform data correctly for chart', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        const lineChart = screen.getByTestId('line-chart');
        const chartData = JSON.parse(lineChart.getAttribute('data-chart-data') || '[]');
        expect(chartData).toHaveLength(2);
        expect(chartData[0].name).toBe('Day 1');
        expect(chartData[0].winRate).toBe(60);
        expect(chartData[0].matches).toBe(10);
      });
    });
  });

  describe('Filters', () => {
    // Helper to get select by finding the label then the next select sibling
    function getSelectByLabel(labelText: string): HTMLSelectElement {
      const label = screen.getByText(labelText);
      const filterGroup = label.closest('.filter-group');
      return filterGroup?.querySelector('select') as HTMLSelectElement;
    }

    it('should render date range filter with default value', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      expect(dateRangeSelect.value).toBe('7days');
    });

    it('should render format filter with default value', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      expect(formatSelect.value).toBe('all');
    });

    it('should render chart type filter with default value', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const chartTypeSelect = getSelectByLabel('Chart Type');
      expect(chartTypeSelect.value).toBe('line');
    });

    it('should update date range when filter changes', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(dateRangeSelect.value).toBe('30days');
      });
    });

    it('should update format when filter changes', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        expect(formatSelect.value).toBe('Ladder');
      });
    });

    it('should switch to bar chart when chart type changes', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const chartTypeSelect = getSelectByLabel('Chart Type');
      fireEvent.change(chartTypeSelect, { target: { value: 'bar' } });

      await waitFor(() => {
        expect(screen.getByTestId('bar-chart')).toBeInTheDocument();
      });
      expect(screen.queryByTestId('line-chart')).not.toBeInTheDocument();
    });
  });

  describe('API Calls', () => {
    // Helper to get select by finding the label then the next select sibling
    function getSelectByLabel(labelText: string): HTMLSelectElement {
      const label = screen.getByText(labelText);
      const filterGroup = label.closest('.filter-group');
      return filterGroup?.querySelector('select') as HTMLSelectElement;
    }

    it('should call GetTrendAnalysis with correct parameters for 7days', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(mockWailsApp.GetTrendAnalysis).toHaveBeenCalled();
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const call = mockWailsApp.GetTrendAnalysis.mock.calls[0] as any[];
      expect(call[2]).toBe('daily'); // periodType for 7days
      expect(call[3]).toEqual([]); // formats for 'all'
    });

    it('should refetch data when date range changes', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(mockWailsApp.GetTrendAnalysis).toHaveBeenCalledTimes(2);
      });
    });

    it('should refetch data when format changes', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        expect(mockWailsApp.GetTrendAnalysis).toHaveBeenCalledTimes(2);
      });
    });

    it('should pass constructed formats for constructed filter', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'constructed' } });

      await waitFor(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const lastCall = mockWailsApp.GetTrendAnalysis.mock.calls.slice(-1)[0] as any[];
        expect(lastCall[3]).toEqual(['Ladder', 'Play']);
      });
    });
  });

  describe('Export Button', () => {
    it('should render export button', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('Export as PNG')).toBeInTheDocument();
      });
    });

    it('should show alert when export button is clicked', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(createMockTrendAnalysis());
      const alertMock = vi.spyOn(window, 'alert').mockImplementation(() => {});

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('Export as PNG')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Export as PNG'));
      expect(alertMock).toHaveBeenCalledWith('Export functionality coming soon!');

      alertMock.mockRestore();
    });
  });

  describe('Trend Display', () => {
    it('should display declining trend correctly', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(
        createMockTrendAnalysis({
          Trend: 'declining',
          TrendValue: -0.05,
        })
      );

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText(/declining/)).toBeInTheDocument();
        expect(screen.getByText(/-5%/)).toBeInTheDocument();
      });
    });

    it('should display stable trend correctly', async () => {
      mockWailsApp.GetTrendAnalysis.mockResolvedValue(
        createMockTrendAnalysis({
          Trend: 'stable',
          TrendValue: 0,
        })
      );

      renderWithProvider(<WinRateTrend />);

      await waitFor(() => {
        expect(screen.getByText('stable')).toBeInTheDocument();
      });
    });
  });
});

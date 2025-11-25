import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import RankProgression from './RankProgression';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { AppProvider } from '../context/AppContext';
import { storage } from '../../wailsjs/go/models';

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
  Line: () => <div data-testid="line" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="cartesian-grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
}));

// Helper function to create mock timeline entry
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function createMockTimelineEntry(overrides: Record<string, any> = {}): storage.RankTimelineEntry {
  return new storage.RankTimelineEntry({
    timestamp: new Date('2024-01-15T10:00:00').toISOString(),
    rank: 'Gold 3',
    rank_class: 'Gold',
    rank_level: 3,
    rank_step: 2,
    is_change: false,
    ...overrides,
  });
}

// Helper function to create mock timeline response
function createMockTimelineResponse(): { entries: storage.RankTimelineEntry[] } {
  return {
    entries: [
      createMockTimelineEntry({
        timestamp: new Date('2024-01-10T10:00:00').toISOString(),
        rank: 'Silver 1',
        rank_class: 'Silver',
        rank_level: 1,
        rank_step: 3,
        is_change: true,
      }),
      createMockTimelineEntry({
        timestamp: new Date('2024-01-12T14:00:00').toISOString(),
        rank: 'Gold 4',
        rank_class: 'Gold',
        rank_level: 4,
        rank_step: 0,
        is_change: true,
      }),
      createMockTimelineEntry({
        timestamp: new Date('2024-01-15T10:00:00').toISOString(),
        rank: 'Gold 3',
        rank_class: 'Gold',
        rank_level: 3,
        rank_step: 2,
        is_change: true,
      }),
    ],
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

describe('RankProgression', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolvePromise: (value: { entries: storage.RankTimelineEntry[] }) => void;
      const loadingPromise = new Promise<{ entries: storage.RankTimelineEntry[] }>((resolve) => {
        resolvePromise = resolve;
      });
      mockWailsApp.GetRankProgressionTimeline.mockReturnValue(loadingPromise);

      renderWithProvider(<RankProgression />);

      expect(screen.getByText('Loading rank progression...')).toBeInTheDocument();

      resolvePromise!(createMockTimelineResponse());
      await waitFor(() => {
        expect(screen.queryByText('Loading rank progression...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockRejectedValue(new Error('Connection error'));

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load rank progression' })).toBeInTheDocument();
      });
      expect(screen.getByText('Connection error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockRejectedValue('Unknown error');

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load rank progression' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no timeline data', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({ entries: [] });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('No rank progression data')).toBeInTheDocument();
      });
      expect(
        screen.getByText('Play ranked constructed matches to track your rank progression over time.')
      ).toBeInTheDocument();
    });

    it('should show empty state with limited message when limited format selected', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({ entries: [] });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('No rank progression data')).toBeInTheDocument();
      });

      // Change to limited format
      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'limited' } });

      await waitFor(() => {
        expect(
          screen.getByText('Play limited (draft/sealed) matches to track your rank progression over time.')
        ).toBeInTheDocument();
      });
    });

    it('should show empty state when API returns null', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(null);

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('No rank progression data')).toBeInTheDocument();
      });
    });
  });

  describe('Data Display', () => {
    it('should render line chart when data is available', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });
    });

    it('should display progression summary', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('Progression Summary')).toBeInTheDocument();
      });
      expect(screen.getByText('Starting Rank')).toBeInTheDocument();
      expect(screen.getByText('Current Rank')).toBeInTheDocument();
      expect(screen.getByText('Direction')).toBeInTheDocument();
    });

    it('should display starting and current rank', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        // There may be multiple occurrences due to summary + timeline, check at least one exists
        const silver1Elements = screen.getAllByText('Silver 1');
        expect(silver1Elements.length).toBeGreaterThan(0);
      });
      const gold3Elements = screen.getAllByText('Gold 3');
      expect(gold3Elements.length).toBeGreaterThan(0);
    });

    it('should display climbing direction when rank increased', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('↑ Climbing')).toBeInTheDocument();
      });
    });

    it('should display falling direction when rank decreased', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({
        entries: [
          createMockTimelineEntry({
            timestamp: new Date('2024-01-10T10:00:00').toISOString(),
            rank: 'Gold 1',
            rank_class: 'Gold',
            rank_level: 1,
          }),
          createMockTimelineEntry({
            timestamp: new Date('2024-01-15T10:00:00').toISOString(),
            rank: 'Silver 2',
            rank_class: 'Silver',
            rank_level: 2,
          }),
        ],
      });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('↓ Falling')).toBeInTheDocument();
      });
    });

    it('should display stable direction when rank unchanged', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({
        entries: [
          createMockTimelineEntry({
            timestamp: new Date('2024-01-10T10:00:00').toISOString(),
            rank: 'Gold 3',
            rank_class: 'Gold',
            rank_level: 3,
          }),
          createMockTimelineEntry({
            timestamp: new Date('2024-01-15T10:00:00').toISOString(),
            rank: 'Gold 3',
            rank_class: 'Gold',
            rank_level: 3,
          }),
        ],
      });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('→ Stable')).toBeInTheDocument();
      });
    });

    it('should display total entries count', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('Total Entries')).toBeInTheDocument();
      });
      // '3' may appear multiple times (entries count, rank changes, step numbers)
      const threeElements = screen.getAllByText('3');
      expect(threeElements.length).toBeGreaterThan(0);
    });

    it('should display rank changes count', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('Rank Changes')).toBeInTheDocument();
      });
    });
  });

  describe('Detailed Timeline', () => {
    it('should display detailed timeline section', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('Detailed Timeline')).toBeInTheDocument();
      });
    });

    it('should mark changed entries in timeline', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        const changedElements = screen.getAllByText('(Changed)');
        expect(changedElements.length).toBe(3);
      });
    });

    it('should display step information when available', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('Step 2')).toBeInTheDocument();
        expect(screen.getByText('Step 3')).toBeInTheDocument();
        expect(screen.getByText('Step 0')).toBeInTheDocument();
      });
    });
  });

  describe('Filters', () => {
    it('should render format filter', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(getSelectByLabel('Format')).toBeInTheDocument();
      });
    });

    it('should render date range filter', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(getSelectByLabel('Date Range')).toBeInTheDocument();
      });
    });

    it('should update format when filter changes', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      expect(formatSelect.value).toBe('constructed');

      fireEvent.change(formatSelect, { target: { value: 'limited' } });

      await waitFor(() => {
        expect(formatSelect.value).toBe('limited');
      });
    });

    it('should refetch data when format changes', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(mockWailsApp.GetRankProgressionTimeline).toHaveBeenCalledTimes(1);
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'limited' } });

      await waitFor(() => {
        expect(mockWailsApp.GetRankProgressionTimeline).toHaveBeenCalledTimes(2);
      });
    });

    it('should update date range when filter changes', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect((dateRangeSelect as HTMLSelectElement).value).toBe('30days');
      });
    });

    it('should refetch data when date range changes', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(mockWailsApp.GetRankProgressionTimeline).toHaveBeenCalledTimes(1);
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(mockWailsApp.GetRankProgressionTimeline).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe('Page Header', () => {
    it('should display page title', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Rank Progression');
      });
    });

    it('should display format note for constructed', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(
          screen.getByText('Showing rank progression for Constructed (Draft/Sealed) ladder')
        ).toBeInTheDocument();
      });
    });

    it('should display format note for limited', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'limited' } });

      await waitFor(() => {
        expect(
          screen.getByText('Showing rank progression for Limited (Draft/Sealed) ladder')
        ).toBeInTheDocument();
      });
    });
  });

  describe('API Calls', () => {
    it('should call GetRankProgressionTimeline with constructed format by default', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(mockWailsApp.GetRankProgressionTimeline).toHaveBeenCalled();
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const call = mockWailsApp.GetRankProgressionTimeline.mock.calls[0] as any[];
      expect(call[0]).toBe('constructed');
      expect(call[3]).toBe('daily');
    });

    it('should call GetRankProgressionTimeline with limited format when selected', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'limited' } });

      await waitFor(() => {
        expect(mockWailsApp.GetRankProgressionTimeline).toHaveBeenCalledTimes(2);
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const calls = mockWailsApp.GetRankProgressionTimeline.mock.calls as any[][];
      const lastCall = calls[calls.length - 1];
      expect(lastCall[0]).toBe('limited');
      expect(lastCall[3]).toBe('daily');
    });
  });

  describe('Chart Data Transformation', () => {
    it('should transform timeline data correctly for chart', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue(createMockTimelineResponse());

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        const lineChart = screen.getByTestId('line-chart');
        const chartData = JSON.parse(lineChart.getAttribute('data-chart-data') || '[]');
        expect(chartData).toHaveLength(3);
        expect(chartData[0]).toHaveProperty('timestamp');
        expect(chartData[0]).toHaveProperty('rankValue');
        expect(chartData[0]).toHaveProperty('rankDisplay');
      });
    });
  });

  describe('Mythic Rank Handling', () => {
    it('should handle Mythic rank correctly', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({
        entries: [
          createMockTimelineEntry({
            timestamp: new Date('2024-01-10T10:00:00').toISOString(),
            rank: 'Diamond 1',
            rank_class: 'Diamond',
            rank_level: 1,
          }),
          createMockTimelineEntry({
            timestamp: new Date('2024-01-15T10:00:00').toISOString(),
            rank: 'Mythic',
            rank_class: 'Mythic',
            rank_level: 1, // Use a level value to get proper numeric conversion
          }),
        ],
      });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByText('↑ Climbing')).toBeInTheDocument();
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle entries with null rank_class', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({
        entries: [
          createMockTimelineEntry({
            rank: 'Unknown',
            rank_class: null,
            rank_level: null,
          }),
        ],
      });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });
    });

    it('should handle entries with undefined rank_level', async () => {
      mockWailsApp.GetRankProgressionTimeline.mockResolvedValue({
        entries: [
          createMockTimelineEntry({
            rank: 'Gold',
            rank_class: 'Gold',
            rank_level: undefined,
          }),
        ],
      });

      renderWithProvider(<RankProgression />);

      await waitFor(() => {
        expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      });
    });
  });
});

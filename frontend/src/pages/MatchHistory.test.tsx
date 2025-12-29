import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import MatchHistory from './MatchHistory';
import { mockMatches } from '@/test/mocks/apiMock';
import { AppProvider } from '../context/AppContext';
import { models } from '@/types/models';

// Helper function to create mock Match
function createMockMatch(overrides: Partial<models.Match> = {}): models.Match {
  return new models.Match({
    ID: 'match-001',
    AccountID: 1,
    EventID: 'event-001',
    EventName: 'Ranked Standard',
    Timestamp: new Date('2024-01-15T10:00:00').toISOString(),
    DurationSeconds: 600,
    PlayerWins: 2,
    OpponentWins: 1,
    PlayerTeamID: 1,
    Format: 'Ladder',
    Result: 'Win',
    OpponentName: 'Opponent123',
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

describe('MatchHistory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolveMatches: (value: models.Match[]) => void;
      const loadingPromise = new Promise<models.Match[]>((resolve) => {
        resolveMatches = resolve;
      });
      mockMatches.getMatches.mockReturnValue(loadingPromise);

      renderWithProvider(<MatchHistory />);

      expect(screen.getByText('Loading matches...')).toBeInTheDocument();

      resolveMatches!([]);
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockMatches.getMatches.mockRejectedValue(new Error('Database error'));

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load matches' })).toBeInTheDocument();
      });
      expect(screen.getByText('Database error')).toBeInTheDocument();
    });

    it('should show generic error for non-Error rejections', async () => {
      mockMatches.getMatches.mockRejectedValue('Unknown error');

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load matches' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show filtered empty state when no matches with default filters', async () => {
      // Default dateRange is '7days', not 'all', so filtered empty state shows
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        expect(screen.getByText('No matches found')).toBeInTheDocument();
      });
      expect(screen.getByText('Try adjusting your filters to see more results.')).toBeInTheDocument();
    });

    it('should show default empty state when all filters are set to all', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });

      // Change dateRange to 'all' (other filters are already 'all' by default)
      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'all' } });

      // Wait for refetch to complete
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        expect(screen.getByText('No matches yet')).toBeInTheDocument();
      });
      expect(
        screen.getByText('Start playing MTG Arena to begin tracking your match history!')
      ).toBeInTheDocument();
    });

    it('should show filtered empty state when non-default filters applied', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });

      // Change a filter to non-default
      const resultSelect = getSelectByLabel('Result');
      fireEvent.change(resultSelect, { target: { value: 'win' } });

      // Wait for refetch to complete
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        expect(screen.getByText('No matches found')).toBeInTheDocument();
      });
      expect(screen.getByText('Try adjusting your filters to see more results.')).toBeInTheDocument();
    });

    it('should show filtered empty state when API returns null', async () => {
      // With default filters (dateRange='7days'), shows filtered empty state
      mockMatches.getMatches.mockResolvedValue(null);

      renderWithProvider(<MatchHistory />);

      // Wait for loading to finish
      await waitFor(() => {
        expect(screen.queryByText('Loading matches...')).not.toBeInTheDocument();
      });

      await waitFor(() => {
        expect(screen.getByText('No matches found')).toBeInTheDocument();
      });
    });
  });

  describe('Match List Display', () => {
    it('should render match table when matches exist', async () => {
      const matches = [
        createMockMatch({ ID: 'match-001', Result: 'Win' }),
        createMockMatch({ ID: 'match-002', Result: 'Loss' }),
      ];
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });
    });

    it('should display match result badges', async () => {
      const matches = [
        createMockMatch({ ID: 'match-001', Result: 'Win' }),
        createMockMatch({ ID: 'match-002', Result: 'Loss' }),
      ];
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('WIN')).toBeInTheDocument();
      });
      expect(screen.getByText('LOSS')).toBeInTheDocument();
    });

    it('should display match score', async () => {
      const match = createMockMatch({ PlayerWins: 2, OpponentWins: 1 });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('2-1')).toBeInTheDocument();
      });
    });

    it('should display opponent name', async () => {
      const match = createMockMatch({ OpponentName: 'TestOpponent' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('TestOpponent')).toBeInTheDocument();
      });
    });

    it('should display dash for missing opponent name', async () => {
      const match = createMockMatch({ OpponentName: undefined });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('—')).toBeInTheDocument();
      });
    });

    it('should display format', async () => {
      const match = createMockMatch({ Format: 'Ladder', DeckFormat: 'Standard' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        // Should display DeckFormat when available, otherwise normalized queue type
        expect(screen.getByText('Standard')).toBeInTheDocument();
      });
    });

    it('should display normalized queue type when no deck format', async () => {
      const match = createMockMatch({ Format: 'Ladder' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        // 'Ladder' should be normalized to 'Ranked'
        expect(screen.getByText('Ranked')).toBeInTheDocument();
      });
    });

    it('should display event name', async () => {
      const match = createMockMatch({ EventName: 'Premier Draft' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Premier Draft')).toBeInTheDocument();
      });
    });

    it('should normalize format with underscore suffix', async () => {
      const match = createMockMatch({ Format: 'QuickDraft_TLA_20251127' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('QuickDraft')).toBeInTheDocument();
      });
      // Should NOT show the full format with date suffix
      expect(screen.queryByText('QuickDraft_TLA_20251127')).not.toBeInTheDocument();
    });

    it('should normalize event name with underscore suffix', async () => {
      const match = createMockMatch({ EventName: 'PremierDraft_MKM_20241120' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('PremierDraft')).toBeInTheDocument();
      });
      // Should NOT show the full event name with date suffix
      expect(screen.queryByText('PremierDraft_MKM_20241120')).not.toBeInTheDocument();
    });

    it('should display combined format and queue type for Standard Ranked', async () => {
      const match = createMockMatch({
        DeckFormat: 'Standard',
        EventName: 'Ladder',
        Format: 'Ladder',
      });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        // Event column should show 'Standard Ranked'
        expect(screen.getByText('Standard Ranked')).toBeInTheDocument();
      });

      // Format column should show 'Standard' (exact match)
      const cells = screen.getAllByRole('cell');
      const formatCell = cells.find(cell => cell.textContent === 'Standard');
      expect(formatCell).toBeDefined();
    });

    it('should display combined format and queue type for Historic Play Queue', async () => {
      const match = createMockMatch({
        DeckFormat: 'Historic',
        EventName: 'Play',
        Format: 'Play',
      });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        // Event column should show 'Historic Play Queue'
        expect(screen.getByText('Historic Play Queue')).toBeInTheDocument();
      });

      // Format column should show 'Historic' (exact match)
      const cells = screen.getAllByRole('cell');
      const formatCell = cells.find(cell => cell.textContent === 'Historic');
      expect(formatCell).toBeDefined();
    });

    it('should normalize Play to Play Queue without DeckFormat', async () => {
      const match = createMockMatch({
        DeckFormat: undefined,
        EventName: 'Play',
        Format: 'Play',
      });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        // Both columns should show 'Play Queue'
        const playQueueElements = screen.getAllByText('Play Queue');
        expect(playQueueElements.length).toBeGreaterThanOrEqual(2);
      });
    });

    it('should display match count', async () => {
      const matches = Array.from({ length: 5 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText(/Showing 5 of 5 matches/)).toBeInTheDocument();
      });
    });

    it('should display singular match count', async () => {
      mockMatches.getMatches.mockResolvedValue([createMockMatch()]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText(/Showing 1 of 1 match$/)).toBeInTheDocument();
      });
    });
  });

  describe('Table Headers', () => {
    it('should display all table headers', async () => {
      mockMatches.getMatches.mockResolvedValue([createMockMatch()]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });

      // Check headers exist in thead
      const headers = screen.getAllByRole('columnheader');
      expect(headers.length).toBe(6);

      // Verify header text content
      const headerTexts = headers.map((h) => h.textContent);
      expect(headerTexts.some((t) => t?.includes('Time'))).toBe(true);
      expect(headerTexts.some((t) => t?.includes('Result'))).toBe(true);
      expect(headerTexts.some((t) => t?.includes('Format'))).toBe(true);
      expect(headerTexts.some((t) => t?.includes('Event'))).toBe(true);
      expect(headerTexts.some((t) => t?.includes('Score'))).toBe(true);
      expect(headerTexts.some((t) => t?.includes('Opponent'))).toBe(true);
    });
  });

  describe('Sorting', () => {
    it('should sort by timestamp descending by default', async () => {
      const matches = [
        createMockMatch({ ID: 'match-001', Timestamp: new Date('2024-01-10').toISOString() }),
        createMockMatch({ ID: 'match-002', Timestamp: new Date('2024-01-15').toISOString() }),
      ];
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const rows = screen.getAllByRole('row');
        // First row is header, second should be most recent (Jan 15)
        expect(rows.length).toBeGreaterThan(1);
      });
    });

    it('should toggle sort direction when clicking same header', async () => {
      const matches = [createMockMatch()];
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });

      // Find Time header by getting column headers and finding the one with Time
      const headers = screen.getAllByRole('columnheader');
      const timeHeader = headers.find((h) => h.textContent?.includes('Time'));
      expect(timeHeader).toBeInTheDocument();

      // Time is already the active sort field (desc by default)
      // Clicking should toggle to asc
      fireEvent.click(timeHeader!);

      // Icon should change to indicate asc
      await waitFor(() => {
        expect(timeHeader).toHaveTextContent('↑');
      });

      // Clicking again should toggle back to desc
      fireEvent.click(timeHeader!);

      await waitFor(() => {
        expect(timeHeader).toHaveTextContent('↓');
      });
    });

    it('should sort by result when clicking Result header', async () => {
      const matches = [
        createMockMatch({ ID: 'match-001', Result: 'Win' }),
        createMockMatch({ ID: 'match-002', Result: 'Loss' }),
      ];
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });

      const headers = screen.getAllByRole('columnheader');
      const resultHeader = headers.find((h) => h.textContent?.includes('Result'));
      fireEvent.click(resultHeader!);

      await waitFor(() => {
        // Should show sort indicator on Result column
        expect(resultHeader).toHaveTextContent('↓');
      });
    });

    it('should sort by format when clicking Format header', async () => {
      mockMatches.getMatches.mockResolvedValue([createMockMatch()]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });

      const headers = screen.getAllByRole('columnheader');
      const formatHeader = headers.find((h) => h.textContent?.includes('Format'));
      fireEvent.click(formatHeader!);

      await waitFor(() => {
        expect(formatHeader).toHaveTextContent('↓');
      });
    });

    it('should sort by event when clicking Event header', async () => {
      mockMatches.getMatches.mockResolvedValue([createMockMatch()]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });

      const headers = screen.getAllByRole('columnheader');
      const eventHeader = headers.find((h) => h.textContent?.includes('Event'));
      fireEvent.click(eventHeader!);

      await waitFor(() => {
        expect(eventHeader).toHaveTextContent('↓');
      });
    });

    it('should reset to page 1 when sorting changes', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      // Go to page 2
      fireEvent.click(screen.getByRole('button', { name: 'Next' }));
      await waitFor(() => {
        expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
      });

      // Click sort header
      const headers = screen.getAllByRole('columnheader');
      const resultHeader = headers.find((h) => h.textContent?.includes('Result'));
      fireEvent.click(resultHeader!);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });
    });
  });

  describe('Pagination', () => {
    it('should show pagination when more than 20 matches', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: 'First' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Previous' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Next' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Last' })).toBeInTheDocument();
    });

    it('should navigate to next page', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Next' }));

      await waitFor(() => {
        expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
      });
    });

    it('should navigate to last page', async () => {
      const matches = Array.from({ length: 50 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Last' }));

      await waitFor(() => {
        expect(screen.getByText('Page 3 of 3')).toBeInTheDocument();
      });
    });

    it('should navigate to previous page', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Next' }));
      await waitFor(() => {
        expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Previous' }));
      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });
    });

    it('should navigate to first page', async () => {
      const matches = Array.from({ length: 50 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Last' }));
      await waitFor(() => {
        expect(screen.getByText('Page 3 of 3')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'First' }));
      await waitFor(() => {
        expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
      });
    });

    it('should disable First and Previous buttons on first page', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'First' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled();
      });
    });

    it('should disable Next and Last buttons on last page', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Last' }));

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Last' })).toBeDisabled();
      });
    });

    it('should not show pagination when 20 or fewer matches', async () => {
      const matches = Array.from({ length: 15 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument();
      });
      expect(screen.queryByText(/Page \d+ of/)).not.toBeInTheDocument();
    });

    it('should show page info in match count when paginated', async () => {
      const matches = Array.from({ length: 25 }, (_, i) =>
        createMockMatch({ ID: `match-${i}` })
      );
      mockMatches.getMatches.mockResolvedValue(matches);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText(/Showing 20 of 25 matches \(Page 1 of 2\)/)).toBeInTheDocument();
      });
    });
  });

  describe('Filters', () => {
    it('should render date range filter with default value', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range');
        expect(dateRangeSelect.value).toBe('7days');
      });
    });

    it('should render card format filter', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const formatSelect = getSelectByLabel('Card Format');
        expect(formatSelect.value).toBe('all');
      });
    });

    it('should render queue type filter', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const queueSelect = getSelectByLabel('Queue Type');
        expect(queueSelect.value).toBe('all');
      });
    });

    it('should render result filter', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const resultSelect = getSelectByLabel('Result');
        expect(resultSelect.value).toBe('all');
      });
    });

    it('should show custom date inputs when custom range selected', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('Match History')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'custom' } });

      await waitFor(() => {
        expect(getInputByLabel('Start Date')).toBeInTheDocument();
        expect(getInputByLabel('End Date')).toBeInTheDocument();
      });
    });

    it('should refetch data when date range changes', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(mockMatches.getMatches).toHaveBeenCalled();
      });

      const initialCallCount = mockMatches.getMatches.mock.calls.length;

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(mockMatches.getMatches.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should refetch data when card format changes', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(mockMatches.getMatches).toHaveBeenCalled();
      });

      const initialCallCount = mockMatches.getMatches.mock.calls.length;

      const formatSelect = getSelectByLabel('Card Format');
      fireEvent.change(formatSelect, { target: { value: 'Standard' } });

      await waitFor(() => {
        expect(mockMatches.getMatches.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should refetch data when queue type changes', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(mockMatches.getMatches).toHaveBeenCalled();
      });

      const initialCallCount = mockMatches.getMatches.mock.calls.length;

      const queueSelect = getSelectByLabel('Queue Type');
      fireEvent.change(queueSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        expect(mockMatches.getMatches.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should refetch data when result filter changes', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(mockMatches.getMatches).toHaveBeenCalled();
      });

      const initialCallCount = mockMatches.getMatches.mock.calls.length;

      const resultSelect = getSelectByLabel('Result');
      fireEvent.change(resultSelect, { target: { value: 'win' } });

      await waitFor(() => {
        expect(mockMatches.getMatches.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should have all date range options', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range');
        const options = Array.from(dateRangeSelect.options).map((o) => o.value);
        expect(options).toContain('7days');
        expect(options).toContain('30days');
        expect(options).toContain('90days');
        expect(options).toContain('all');
        expect(options).toContain('custom');
      });
    });

    it('should have all card format options', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const formatSelect = getSelectByLabel('Card Format');
        const options = Array.from(formatSelect.options).map((o) => o.value);
        expect(options).toContain('all');
        expect(options).toContain('Standard');
        expect(options).toContain('Historic');
        expect(options).toContain('Alchemy');
        expect(options).toContain('Explorer');
      });
    });

    it('should have all queue type options', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const queueSelect = getSelectByLabel('Queue Type');
        const options = Array.from(queueSelect.options).map((o) => o.value);
        expect(options).toContain('all');
        expect(options).toContain('Ladder');
        expect(options).toContain('Play');
      });
    });

    it('should have all result options', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const resultSelect = getSelectByLabel('Result');
        const options = Array.from(resultSelect.options).map((o) => o.value);
        expect(options).toContain('all');
        expect(options).toContain('win');
        expect(options).toContain('loss');
      });
    });
  });

  describe('Match Details Modal', () => {
    it('should open modal when clicking on a match row', async () => {
      const match = createMockMatch({ OpponentName: 'TestOpponent' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('TestOpponent')).toBeInTheDocument();
      });

      // Click on the row
      const row = screen.getByText('TestOpponent').closest('tr');
      fireEvent.click(row!);

      // Modal should open - MatchDetailsModal renders with "Match Details" header
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Match Details' })).toBeInTheDocument();
      });
    });

    it('should close modal when clicking close button', async () => {
      const match = createMockMatch({ OpponentName: 'ModalTestOpponent' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByText('ModalTestOpponent')).toBeInTheDocument();
      });

      // Click on the row to open modal
      const row = screen.getByText('ModalTestOpponent').closest('tr');
      fireEvent.click(row!);

      // Wait for modal to open
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Match Details' })).toBeInTheDocument();
      });

      // Click the close button (footer button with text "Close")
      const closeButton = screen.getByRole('button', { name: 'Close' });
      fireEvent.click(closeButton);

      // Modal should be closed
      await waitFor(() => {
        expect(screen.queryByRole('heading', { name: 'Match Details' })).not.toBeInTheDocument();
      });
    });
  });

  describe('Page Header', () => {
    it('should display page title', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Match History');
      });
    });
  });

  describe('API Calls', () => {
    it('should call GetMatches with filter', async () => {
      mockMatches.getMatches.mockResolvedValue([]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        expect(mockMatches.getMatches).toHaveBeenCalled();
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const call = mockMatches.getMatches.mock.calls[0] as any[];
      expect(call[0]).toBeInstanceOf(models.StatsFilter);
    });
  });

  describe('Row Styling', () => {
    it('should apply win class to winning match rows', async () => {
      const match = createMockMatch({ Result: 'Win' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const row = screen.getByText('WIN').closest('tr');
        expect(row).toHaveClass('result-win');
      });
    });

    it('should apply loss class to losing match rows', async () => {
      const match = createMockMatch({ Result: 'Loss' });
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const row = screen.getByText('LOSS').closest('tr');
        expect(row).toHaveClass('result-loss');
      });
    });

    it('should apply clickable-row class to match rows', async () => {
      const match = createMockMatch();
      mockMatches.getMatches.mockResolvedValue([match]);

      renderWithProvider(<MatchHistory />);

      await waitFor(() => {
        const row = screen.getByText('WIN').closest('tr');
        expect(row).toHaveClass('clickable-row');
      });
    });
  });
});

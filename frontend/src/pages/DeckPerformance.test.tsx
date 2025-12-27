import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import DeckPerformance from './DeckPerformance';
import { mockMatches } from '@/test/mocks/apiMock';
import { mockEventEmitter } from '@/test/mocks/websocketMock';
import { AppProvider } from '../context/AppContext';
import { models } from '@/types/models';

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

// Helper function to create mock deck stats response
function createMockDeckStatsResponse(): Record<string, models.Statistics> {
  return {
    'Mono Red Aggro': createMockStatistics({ WinRate: 0.65, TotalMatches: 30, MatchesWon: 20, MatchesLost: 10 }),
    'Azorius Control': createMockStatistics({ WinRate: 0.55, TotalMatches: 25, MatchesWon: 14, MatchesLost: 11 }),
    'Gruul Stompy': createMockStatistics({ WinRate: 0.45, TotalMatches: 15, MatchesWon: 7, MatchesLost: 8 }),
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

describe('DeckPerformance', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
    localStorage.clear();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolvePromise: (value: Record<string, models.Statistics>) => void;
      const loadingPromise = new Promise<Record<string, models.Statistics>>((resolve) => {
        resolvePromise = resolve;
      });
      mockMatches.getMatchupMatrix.mockReturnValue(loadingPromise);

      renderWithProvider(<DeckPerformance />);

      expect(screen.getByText('Loading deck statistics...')).toBeInTheDocument();

      resolvePromise!(createMockDeckStatsResponse());
      await waitFor(() => {
        expect(screen.queryByText('Loading deck statistics...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockMatches.getMatchupMatrix.mockRejectedValue(new Error('Database error'));

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load deck statistics' })).toBeInTheDocument();
      });
      expect(screen.getByText('Database error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockMatches.getMatchupMatrix.mockRejectedValue('Unknown error');

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Failed to load deck statistics' })).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no deck data', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue({});

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('No deck data')).toBeInTheDocument();
      });
      expect(
        screen.getByText('Play matches with different decks to see your deck performance statistics.')
      ).toBeInTheDocument();
    });

    it('should show empty state when API returns null', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(null);

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('No deck data')).toBeInTheDocument();
      });
    });
  });

  describe('Data Display', () => {
    it('should render deck cards with statistics', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });
      expect(screen.getByText('Azorius Control')).toBeInTheDocument();
      expect(screen.getByText('Gruul Stompy')).toBeInTheDocument();
    });

    it('should display win rate correctly', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('65%')).toBeInTheDocument();
      });
      expect(screen.getByText('55%')).toBeInTheDocument();
      expect(screen.getByText('45%')).toBeInTheDocument();
    });

    it('should display match counts', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('30')).toBeInTheDocument();
      });
      expect(screen.getByText('25')).toBeInTheDocument();
      expect(screen.getByText('15')).toBeInTheDocument();
    });

    it('should display deck count', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('3 decks found')).toBeInTheDocument();
      });
    });

    it('should display singular deck count for one deck', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue({
        'Mono Red Aggro': createMockStatistics(),
      });

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('1 deck found')).toBeInTheDocument();
      });
    });

    it('should display wins and losses', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue({
        'Test Deck': createMockStatistics({ MatchesWon: 15, MatchesLost: 5 }),
      });

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('15W - 5L')).toBeInTheDocument();
      });
    });
  });

  describe('Filters', () => {
    it('should render date range filter with default value', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        const dateRangeSelect = getSelectByLabel('Date Range') as HTMLSelectElement;
        expect(dateRangeSelect.value).toBe('7days');
      });
    });

    it('should render format filter', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(getSelectByLabel('Format')).toBeInTheDocument();
      });
    });

    it('should render sort by filter', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        const sortBySelect = getSelectByLabel('Sort By') as HTMLSelectElement;
        expect(sortBySelect.value).toBe('winRate');
      });
    });

    it('should render sort order filter', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        const sortOrderSelect = getSelectByLabel('Sort Order') as HTMLSelectElement;
        expect(sortOrderSelect.value).toBe('desc');
      });
    });

    it('should show custom date inputs when custom range selected', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: 'custom' } });

      await waitFor(() => {
        expect(getInputByLabel('Start Date')).toBeInTheDocument();
        expect(getInputByLabel('End Date')).toBeInTheDocument();
      });
    });

    it('should refetch data when date range changes', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(mockMatches.getMatchupMatrix).toHaveBeenCalled();
      });

      const initialCallCount = mockMatches.getMatchupMatrix.mock.calls.length;

      const dateRangeSelect = getSelectByLabel('Date Range');
      fireEvent.change(dateRangeSelect, { target: { value: '30days' } });

      await waitFor(() => {
        expect(mockMatches.getMatchupMatrix.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should refetch data when format changes', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(mockMatches.getMatchupMatrix).toHaveBeenCalled();
      });

      const initialCallCount = mockMatches.getMatchupMatrix.mock.calls.length;

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        expect(mockMatches.getMatchupMatrix.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });
  });

  describe('Sorting', () => {
    it('should sort by win rate descending by default', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        const deckCards = screen.getAllByRole('heading', { level: 3 });
        expect(deckCards[0]).toHaveTextContent('Mono Red Aggro');
        expect(deckCards[1]).toHaveTextContent('Azorius Control');
        expect(deckCards[2]).toHaveTextContent('Gruul Stompy');
      });
    });

    it('should sort by win rate ascending when changed', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const sortOrderSelect = getSelectByLabel('Sort Order');
      fireEvent.change(sortOrderSelect, { target: { value: 'asc' } });

      await waitFor(() => {
        const deckCards = screen.getAllByRole('heading', { level: 3 });
        expect(deckCards[0]).toHaveTextContent('Gruul Stompy');
        expect(deckCards[2]).toHaveTextContent('Mono Red Aggro');
      });
    });

    it('should sort by match count when selected', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const sortBySelect = getSelectByLabel('Sort By');
      fireEvent.change(sortBySelect, { target: { value: 'matches' } });

      await waitFor(() => {
        const deckCards = screen.getAllByRole('heading', { level: 3 });
        expect(deckCards[0]).toHaveTextContent('Mono Red Aggro'); // 30 matches
        expect(deckCards[1]).toHaveTextContent('Azorius Control'); // 25 matches
        expect(deckCards[2]).toHaveTextContent('Gruul Stompy'); // 15 matches
      });
    });

    it('should sort by deck name when selected', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const sortBySelect = getSelectByLabel('Sort By');
      const sortOrderSelect = getSelectByLabel('Sort Order');

      fireEvent.change(sortBySelect, { target: { value: 'name' } });
      fireEvent.change(sortOrderSelect, { target: { value: 'asc' } });

      await waitFor(() => {
        const deckCards = screen.getAllByRole('heading', { level: 3 });
        expect(deckCards[0]).toHaveTextContent('Azorius Control');
        expect(deckCards[1]).toHaveTextContent('Gruul Stompy');
        expect(deckCards[2]).toHaveTextContent('Mono Red Aggro');
      });
    });
  });

  describe('Real-time Updates', () => {
    it('should reload data when stats:updated event fires', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const initialCallCount = mockMatches.getMatchupMatrix.mock.calls.length;

      // Update mock to return different data
      mockMatches.getMatchupMatrix.mockResolvedValue({
        'Updated Deck': createMockStatistics(),
      });

      await act(async () => {
        mockEventEmitter.emit('stats:updated', {});
      });

      await waitFor(() => {
        expect(mockMatches.getMatchupMatrix.mock.calls.length).toBeGreaterThan(initialCallCount);
      });
    });

    it('should unsubscribe from events on unmount', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      const { unmount } = renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      unmount();

      const callCountAfterUnmount = mockMatches.getMatchupMatrix.mock.calls.length;

      await act(async () => {
        mockEventEmitter.emit('stats:updated', {});
      });

      // Should not have called GetStatsByDeck again after unmount
      expect(mockMatches.getMatchupMatrix.mock.calls.length).toBe(callCountAfterUnmount);
    });
  });

  describe('Page Title', () => {
    it('should display page title', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Deck Performance');
      });
    });
  });

  describe('Unknown Deck Handling', () => {
    it('should display "Unknown Deck" for empty deck name', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue({
        '': createMockStatistics(),
      });

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Unknown Deck')).toBeInTheDocument();
      });
    });
  });

  describe('API Filter Parameters', () => {
    it('should pass constructed formats for constructed filter', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'constructed' } });

      await waitFor(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const lastCall = mockMatches.getMatchupMatrix.mock.calls.slice(-1)[0] as any[];
        const filter = lastCall[0] as models.StatsFilter;
        expect(filter.Formats).toEqual(['Ladder', 'Play']);
      });
    });

    it('should pass single format for specific format filter', async () => {
      mockMatches.getMatchupMatrix.mockResolvedValue(createMockDeckStatsResponse());

      renderWithProvider(<DeckPerformance />);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const formatSelect = getSelectByLabel('Format');
      fireEvent.change(formatSelect, { target: { value: 'Ladder' } });

      await waitFor(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const lastCall = mockMatches.getMatchupMatrix.mock.calls.slice(-1)[0] as any[];
        const filter = lastCall[0] as models.StatsFilter;
        expect(filter.Format).toBe('Ladder');
      });
    });
  });
});

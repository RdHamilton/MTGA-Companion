import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { AppProvider, useAppContext } from './AppContext';

// Test component to access context
function TestConsumer({ onMount }: { onMount?: (context: ReturnType<typeof useAppContext>) => void }) {
  const context = useAppContext();
  if (onMount) {
    onMount(context);
  }
  return (
    <div>
      <span data-testid="date-range">{context.filters.matchHistory.dateRange}</span>
      <span data-testid="card-format">{context.filters.matchHistory.cardFormat}</span>
      <span data-testid="chart-type">{context.filters.winRateTrend.chartType}</span>
    </div>
  );
}

// Test component that updates filters
function TestUpdater() {
  const { filters, updateFilters, resetFilters } = useAppContext();
  return (
    <div>
      <span data-testid="date-range">{filters.matchHistory.dateRange}</span>
      <button
        data-testid="update-date-range"
        onClick={() => updateFilters('matchHistory', { dateRange: '30days' })}
      >
        Update Date Range
      </button>
      <button
        data-testid="update-format"
        onClick={() => updateFilters('matchHistory', { cardFormat: 'Standard' })}
      >
        Update Format
      </button>
      <button data-testid="reset-filters" onClick={resetFilters}>
        Reset
      </button>
    </div>
  );
}

describe('AppContext', () => {
  const STORAGE_KEY = 'mtga-companion-filters';

  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
    vi.clearAllMocks();
  });

  describe('AppProvider', () => {
    it('should render children', () => {
      render(
        <AppProvider>
          <div data-testid="child">Child Content</div>
        </AppProvider>
      );

      expect(screen.getByTestId('child')).toBeInTheDocument();
      expect(screen.getByText('Child Content')).toBeInTheDocument();
    });

    it('should provide default filter values', () => {
      render(
        <AppProvider>
          <TestConsumer />
        </AppProvider>
      );

      expect(screen.getByTestId('date-range')).toHaveTextContent('7days');
      expect(screen.getByTestId('card-format')).toHaveTextContent('all');
      expect(screen.getByTestId('chart-type')).toHaveTextContent('line');
    });

    it('should load filters from localStorage', () => {
      const savedFilters = {
        matchHistory: {
          dateRange: '30days',
          customStartDate: '',
          customEndDate: '',
          cardFormat: 'Standard',
          queueType: 'all',
          result: 'all',
        },
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(savedFilters));

      render(
        <AppProvider>
          <TestConsumer />
        </AppProvider>
      );

      expect(screen.getByTestId('date-range')).toHaveTextContent('30days');
      expect(screen.getByTestId('card-format')).toHaveTextContent('Standard');
    });

    it('should handle invalid JSON in localStorage gracefully', () => {
      localStorage.setItem(STORAGE_KEY, 'invalid-json');
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(
        <AppProvider>
          <TestConsumer />
        </AppProvider>
      );

      // Should fall back to defaults
      expect(screen.getByTestId('date-range')).toHaveTextContent('7days');
      expect(consoleSpy).toHaveBeenCalled();

      consoleSpy.mockRestore();
    });

    it('should merge saved filters with defaults', () => {
      // Save only partial filters
      const partialFilters = {
        matchHistory: {
          dateRange: '30days',
        },
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(partialFilters));

      render(
        <AppProvider>
          <TestConsumer />
        </AppProvider>
      );

      // Saved value should be used
      expect(screen.getByTestId('date-range')).toHaveTextContent('30days');
      // Other defaults should still be present
      expect(screen.getByTestId('chart-type')).toHaveTextContent('line');
    });
  });

  describe('useAppContext', () => {
    it('should throw error when used outside provider', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      expect(() => {
        render(<TestConsumer />);
      }).toThrow('useAppContext must be used within an AppProvider');

      consoleSpy.mockRestore();
    });

    it('should provide filters object', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      expect(capturedContext).not.toBeNull();
      expect(capturedContext!.filters).toBeDefined();
      expect(capturedContext!.filters.matchHistory).toBeDefined();
      expect(capturedContext!.filters.winRateTrend).toBeDefined();
      expect(capturedContext!.filters.deckPerformance).toBeDefined();
    });

    it('should provide updateFilters function', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      expect(typeof capturedContext!.updateFilters).toBe('function');
    });

    it('should provide resetFilters function', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      expect(typeof capturedContext!.resetFilters).toBe('function');
    });
  });

  describe('updateFilters', () => {
    it('should update matchHistory filters', async () => {
      render(
        <AppProvider>
          <TestUpdater />
        </AppProvider>
      );

      expect(screen.getByTestId('date-range')).toHaveTextContent('7days');

      await act(async () => {
        screen.getByTestId('update-date-range').click();
      });

      expect(screen.getByTestId('date-range')).toHaveTextContent('30days');
    });

    it('should merge updates without overwriting other fields', async () => {
      render(
        <AppProvider>
          <TestUpdater />
        </AppProvider>
      );

      await act(async () => {
        screen.getByTestId('update-date-range').click();
      });

      await act(async () => {
        screen.getByTestId('update-format').click();
      });

      // Both updates should be preserved
      expect(screen.getByTestId('date-range')).toHaveTextContent('30days');
    });

    it('should persist updates to localStorage', async () => {
      render(
        <AppProvider>
          <TestUpdater />
        </AppProvider>
      );

      await act(async () => {
        screen.getByTestId('update-date-range').click();
      });

      const savedFilters = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
      expect(savedFilters.matchHistory.dateRange).toBe('30days');
    });
  });

  describe('resetFilters', () => {
    it('should reset all filters to defaults', async () => {
      render(
        <AppProvider>
          <TestUpdater />
        </AppProvider>
      );

      // First update a filter
      await act(async () => {
        screen.getByTestId('update-date-range').click();
      });
      expect(screen.getByTestId('date-range')).toHaveTextContent('30days');

      // Then reset
      await act(async () => {
        screen.getByTestId('reset-filters').click();
      });

      expect(screen.getByTestId('date-range')).toHaveTextContent('7days');
    });

    it('should reset filters in localStorage to defaults', async () => {
      render(
        <AppProvider>
          <TestUpdater />
        </AppProvider>
      );

      // First update to create localStorage entry with custom value
      await act(async () => {
        screen.getByTestId('update-date-range').click();
      });
      const savedAfterUpdate = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
      expect(savedAfterUpdate.matchHistory.dateRange).toBe('30days');

      // Then reset
      await act(async () => {
        screen.getByTestId('reset-filters').click();
      });

      // After reset, the useEffect saves defaults to localStorage
      // The implementation calls localStorage.removeItem then the useEffect re-saves defaults
      const savedAfterReset = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
      expect(savedAfterReset.matchHistory.dateRange).toBe('7days');
    });
  });

  describe('default filter values', () => {
    it('should have correct matchHistory defaults', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      const matchHistory = capturedContext!.filters.matchHistory;
      expect(matchHistory.dateRange).toBe('7days');
      expect(matchHistory.customStartDate).toBe('');
      expect(matchHistory.customEndDate).toBe('');
      expect(matchHistory.cardFormat).toBe('all');
      expect(matchHistory.queueType).toBe('all');
      expect(matchHistory.result).toBe('all');
    });

    it('should have correct winRateTrend defaults', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      const winRateTrend = capturedContext!.filters.winRateTrend;
      expect(winRateTrend.dateRange).toBe('7days');
      expect(winRateTrend.format).toBe('all');
      expect(winRateTrend.chartType).toBe('line');
    });

    it('should have correct deckPerformance defaults', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      const deckPerformance = capturedContext!.filters.deckPerformance;
      expect(deckPerformance.dateRange).toBe('7days');
      expect(deckPerformance.format).toBe('all');
      expect(deckPerformance.chartType).toBe('bar');
      expect(deckPerformance.sortBy).toBe('winRate');
      expect(deckPerformance.sortDirection).toBe('desc');
    });

    it('should have correct rankProgression defaults', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      const rankProgression = capturedContext!.filters.rankProgression;
      expect(rankProgression.format).toBe('constructed');
      expect(rankProgression.dateRange).toBe('all');
    });

    it('should have correct formatDistribution defaults', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      const formatDistribution = capturedContext!.filters.formatDistribution;
      expect(formatDistribution.dateRange).toBe('7days');
      expect(formatDistribution.chartType).toBe('bar');
      expect(formatDistribution.sortBy).toBe('matches');
      expect(formatDistribution.sortDirection).toBe('desc');
    });

    it('should have correct resultBreakdown defaults', () => {
      let capturedContext: ReturnType<typeof useAppContext> | null = null;

      render(
        <AppProvider>
          <TestConsumer
            onMount={(ctx) => {
              capturedContext = ctx;
            }}
          />
        </AppProvider>
      );

      const resultBreakdown = capturedContext!.filters.resultBreakdown;
      expect(resultBreakdown.dateRange).toBe('7days');
      expect(resultBreakdown.format).toBe('all');
    });
  });
});

import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';

// Filter state types
interface FilterState {
  // Match History filters
  matchHistory: {
    dateRange: string;
    customStartDate: string;
    customEndDate: string;
    cardFormat: string; // Standard, Historic, Alchemy, etc.
    queueType: string; // Ladder, Play, etc.
    result: string;
  };
  // Chart filters
  winRateTrend: {
    dateRange: string;
    format: string;
    chartType: 'line' | 'bar';
  };
  deckPerformance: {
    dateRange: string;
    customStartDate: string;
    customEndDate: string;
    format: string;
    chartType: 'bar' | 'table';
    sortBy: 'winRate' | 'matches' | 'name';
    sortDirection: 'asc' | 'desc';
  };
  rankProgression: {
    format: string;
    dateRange: string;
  };
  formatDistribution: {
    dateRange: string;
    customStartDate: string;
    customEndDate: string;
    chartType: 'pie' | 'bar';
    sortBy: 'matches' | 'winRate' | 'name';
    sortDirection: 'asc' | 'desc';
  };
  resultBreakdown: {
    dateRange: string;
    customStartDate: string;
    customEndDate: string;
    format: string;
  };
}

// Default filter state
const defaultFilters: FilterState = {
  matchHistory: {
    dateRange: '7days',
    customStartDate: '',
    customEndDate: '',
    cardFormat: 'all',
    queueType: 'all',
    result: 'all',
  },
  winRateTrend: {
    dateRange: '7days',
    format: 'all',
    chartType: 'line',
  },
  deckPerformance: {
    dateRange: '7days',
    customStartDate: '',
    customEndDate: '',
    format: 'all',
    chartType: 'bar',
    sortBy: 'winRate',
    sortDirection: 'desc',
  },
  rankProgression: {
    format: 'constructed',
    dateRange: 'all',
  },
  formatDistribution: {
    dateRange: '7days',
    customStartDate: '',
    customEndDate: '',
    chartType: 'bar',
    sortBy: 'matches',
    sortDirection: 'desc',
  },
  resultBreakdown: {
    dateRange: '7days',
    customStartDate: '',
    customEndDate: '',
    format: 'all',
  },
};

// Context type
interface AppContextType {
  filters: FilterState;
  updateFilters: <K extends keyof FilterState>(
    page: K,
    updates: Partial<FilterState[K]>
  ) => void;
  resetFilters: () => void;
}

// Create context
const AppContext = createContext<AppContextType | undefined>(undefined);

// Local storage key
const STORAGE_KEY = 'mtga-companion-filters';

// Provider component
interface AppProviderProps {
  children: ReactNode;
}

export const AppProvider = ({ children }: AppProviderProps) => {
  // Load initial state from localStorage or use defaults
  const [filters, setFilters] = useState<FilterState>(() => {
    try {
      const savedFilters = localStorage.getItem(STORAGE_KEY);
      if (savedFilters) {
        return { ...defaultFilters, ...JSON.parse(savedFilters) };
      }
    } catch (error) {
      console.error('Failed to load filters from localStorage:', error);
    }
    return defaultFilters;
  });

  // Save filters to localStorage whenever they change
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(filters));
    } catch (error) {
      console.error('Failed to save filters to localStorage:', error);
    }
  }, [filters]);

  // Update filters for a specific page
  const updateFilters = <K extends keyof FilterState>(
    page: K,
    updates: Partial<FilterState[K]>
  ) => {
    setFilters((prev) => ({
      ...prev,
      [page]: {
        ...prev[page],
        ...updates,
      },
    }));
  };

  // Reset all filters to defaults
  const resetFilters = () => {
    setFilters(defaultFilters);
    localStorage.removeItem(STORAGE_KEY);
  };

  const value: AppContextType = {
    filters,
    updateFilters,
    resetFilters,
  };

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};

// Custom hook to use the app context
export const useAppContext = (): AppContextType => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useAppContext must be used within an AppProvider');
  }
  return context;
};

export default AppContext;

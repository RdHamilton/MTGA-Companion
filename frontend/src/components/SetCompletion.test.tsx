import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { models, gui } from '@/types/models';
import { mockCollection, mockCards } from '@/test/mocks/apiMock';

import SetCompletion from './SetCompletion';

// Helper function to create mock set completion
function createMockSetCompletion(overrides: Record<string, unknown> = {}): models.SetCompletion {
  return Object.assign(new models.SetCompletion({}), {
    SetCode: 'dsk',
    SetName: 'Duskmourn: House of Horror',
    TotalCards: 250,
    OwnedCards: 100,
    Percentage: 40.0,
    RarityBreakdown: {
      common: Object.assign(new models.RarityCompletion({}), {
        Rarity: 'common',
        Total: 100,
        Owned: 50,
        Percentage: 50.0,
      }),
      uncommon: Object.assign(new models.RarityCompletion({}), {
        Rarity: 'uncommon',
        Total: 80,
        Owned: 30,
        Percentage: 37.5,
      }),
      rare: Object.assign(new models.RarityCompletion({}), {
        Rarity: 'rare',
        Total: 50,
        Owned: 15,
        Percentage: 30.0,
      }),
      mythic: Object.assign(new models.RarityCompletion({}), {
        Rarity: 'mythic',
        Total: 20,
        Owned: 5,
        Percentage: 25.0,
      }),
    },
    ...overrides,
  });
}

// Helper to create mock set info
function createMockSetInfo(overrides: Record<string, unknown> = {}): gui.SetInfo {
  return Object.assign(new gui.SetInfo({}), {
    code: 'dsk',
    name: 'Duskmourn: House of Horror',
    iconSvgUri: 'https://example.com/dsk.svg',
    setType: 'expansion',
    releasedAt: '2024-09-27',
    cardCount: 250,
    ...overrides,
  });
}

describe('SetCompletion', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching data', async () => {
      let resolvePromise: (value: models.SetCompletion[]) => void;
      const loadingPromise = new Promise<models.SetCompletion[]>((resolve) => {
        resolvePromise = resolve;
      });
      mockCollection.getSetCompletion.mockReturnValue(loadingPromise);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      expect(screen.getByText('Loading set completion...')).toBeInTheDocument();

      resolvePromise!([createMockSetCompletion()]);
      await waitFor(() => {
        expect(screen.queryByText('Loading set completion...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error message when API fails', async () => {
      mockCollection.getSetCompletion.mockRejectedValue(new Error('Database error'));
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Database error')).toBeInTheDocument();
      });
    });
  });

  describe('Set Completion Display', () => {
    it('should render set completion data', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Duskmourn: House of Horror')).toBeInTheDocument();
      });
      expect(screen.getByText('DSK')).toBeInTheDocument();
      expect(screen.getByText('100/250 (40.0%)')).toBeInTheDocument();
    });

    it('should display page title', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Set Completion' })).toBeInTheDocument();
      });
    });

    it('should expand rarity breakdown when clicking on set', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Duskmourn: House of Horror')).toBeInTheDocument();
      });

      // Click to expand
      const setHeader = screen.getByText('Duskmourn: House of Horror').closest('.set-header');
      fireEvent.click(setHeader!);

      await waitFor(() => {
        expect(screen.getByText('Mythic')).toBeInTheDocument();
      });
      expect(screen.getByText('Rare')).toBeInTheDocument();
      expect(screen.getByText('Uncommon')).toBeInTheDocument();
      expect(screen.getByText('Common')).toBeInTheDocument();
    });

    it('should show rarity counts in breakdown', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Duskmourn: House of Horror')).toBeInTheDocument();
      });

      // Click to expand
      const setHeader = screen.getByText('Duskmourn: House of Horror').closest('.set-header');
      fireEvent.click(setHeader!);

      await waitFor(() => {
        expect(screen.getByText('5/20')).toBeInTheDocument(); // Mythic
      });
      expect(screen.getByText('15/50')).toBeInTheDocument(); // Rare
      expect(screen.getByText('30/80')).toBeInTheDocument(); // Uncommon
      expect(screen.getByText('50/100')).toBeInTheDocument(); // Common
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no data', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('No set completion data available.')).toBeInTheDocument();
      });
    });
  });

  describe('Sort Options', () => {
    it('should have sort dropdown', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Sort by:')).toBeInTheDocument();
      });
      expect(screen.getByText('Newest First')).toBeInTheDocument();
    });

    it('should sort by name when selected', async () => {
      const set1 = createMockSetCompletion({ SetCode: 'blb', SetName: 'Bloomburrow' });
      const set2 = createMockSetCompletion({ SetCode: 'dsk', SetName: 'Duskmourn: House of Horror' });
      mockCollection.getSetCompletion.mockResolvedValue([set1, set2]);
      mockCards.getAllSetInfo.mockResolvedValue([
        createMockSetInfo({ code: 'blb', name: 'Bloomburrow', releasedAt: '2024-08-02' }),
        createMockSetInfo({ code: 'dsk', name: 'Duskmourn: House of Horror', releasedAt: '2024-09-27' }),
      ]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Duskmourn: House of Horror')).toBeInTheDocument();
      });

      const sortSelect = screen.getByDisplayValue('Newest First');
      fireEvent.change(sortSelect, { target: { value: 'name-asc' } });

      // After sorting by name A-Z, Bloomburrow should come first
      const setNames = screen.getAllByText(/Bloomburrow|Duskmourn/);
      expect(setNames[0].textContent).toBe('Bloomburrow');
    });
  });

  describe('Close Button', () => {
    it('should call onClose when close button clicked', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);
      const onClose = vi.fn();

      render(<SetCompletion onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Duskmourn: House of Horror')).toBeInTheDocument();
      });

      const closeButton = screen.getByTitle('Close');
      fireEvent.click(closeButton);

      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it('should not show close button when onClose not provided', async () => {
      mockCollection.getSetCompletion.mockResolvedValue([createMockSetCompletion()]);
      mockCards.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      render(<SetCompletion />);

      await waitFor(() => {
        expect(screen.getByText('Duskmourn: House of Horror')).toBeInTheDocument();
      });

      expect(screen.queryByTitle('Close')).not.toBeInTheDocument();
    });
  });
});

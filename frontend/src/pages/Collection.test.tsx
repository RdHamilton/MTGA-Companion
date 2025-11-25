import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { gui } from '../../wailsjs/go/models';

// Create hoisted mock functions
const { mockGetCollection, mockGetCollectionStats, mockGetAllSetInfo } = vi.hoisted(() => ({
  mockGetCollection: vi.fn(),
  mockGetCollectionStats: vi.fn(),
  mockGetAllSetInfo: vi.fn(),
}));

// Mock the Wails App module
vi.mock('../../wailsjs/go/main/App', () => ({
  GetCollection: mockGetCollection,
  GetCollectionStats: mockGetCollectionStats,
  GetAllSetInfo: mockGetAllSetInfo,
}));

import Collection from './Collection';

// Helper function to create mock collection card
function createMockCollectionCard(overrides: Record<string, unknown> = {}): gui.CollectionCard {
  return new gui.CollectionCard({
    cardId: 12345,
    arenaId: 12345,
    quantity: 4,
    name: 'Lightning Bolt',
    setCode: 'sta',
    setName: 'Strixhaven Mystical Archive',
    rarity: 'rare',
    manaCost: '{R}',
    cmc: 1,
    typeLine: 'Instant',
    colors: ['R'],
    colorIdentity: ['R'],
    imageUri: 'https://example.com/card.jpg',
    power: '',
    toughness: '',
    ...overrides,
  });
}

// Helper to create mock collection response
function createMockCollectionResponse(cards: gui.CollectionCard[] = [], totalCount?: number, filterCount?: number): gui.CollectionResponse {
  return new gui.CollectionResponse({
    cards,
    totalCount: totalCount ?? cards.length,
    filterCount: filterCount ?? cards.length,
  });
}

// Helper to create mock collection stats
function createMockCollectionStats(overrides: Record<string, unknown> = {}): gui.CollectionStats {
  return new gui.CollectionStats({
    totalUniqueCards: 100,
    totalCards: 400,
    commonCount: 200,
    uncommonCount: 100,
    rareCount: 75,
    mythicCount: 25,
    ...overrides,
  });
}

// Helper to create mock set info
function createMockSetInfo(overrides: Record<string, unknown> = {}): gui.SetInfo {
  return new gui.SetInfo({
    code: 'sta',
    name: 'Strixhaven Mystical Archive',
    iconSvgUri: 'https://example.com/set.svg',
    setType: 'expansion',
    releasedAt: '2021-04-23',
    cardCount: 63,
    ...overrides,
  });
}

// Wrapper component with router
function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

// Setup window.go to simulate Wails runtime being ready
function setupWailsRuntime() {
  (window as unknown as Record<string, unknown>).go = {};
}

function clearWailsRuntime() {
  delete (window as unknown as Record<string, unknown>).go;
}

describe('Collection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setupWailsRuntime();
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    clearWailsRuntime();
    vi.useRealTimers();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching collection', async () => {
      let resolvePromise: (value: gui.CollectionResponse) => void;
      const loadingPromise = new Promise<gui.CollectionResponse>((resolve) => {
        resolvePromise = resolve;
      });
      mockGetCollection.mockReturnValue(loadingPromise);
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      expect(screen.getByText('Loading collection...')).toBeInTheDocument();

      resolvePromise!(createMockCollectionResponse());
      await waitFor(() => {
        expect(screen.queryByText('Loading collection...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockGetCollection.mockRejectedValue(new Error('Database error'));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Collection')).toBeInTheDocument();
      });
      expect(screen.getByText('Database error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockGetCollection.mockRejectedValue('Unknown error');
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Collection')).toBeInTheDocument();
      });
      expect(screen.getByText('Failed to load collection')).toBeInTheDocument();
    });

    it('should show error when Wails runtime not initialized after timeout', async () => {
      clearWailsRuntime();
      mockGetCollection.mockResolvedValue(createMockCollectionResponse());

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(5100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Collection')).toBeInTheDocument();
      });
      expect(screen.getByText('Wails runtime not initialized')).toBeInTheDocument();
    });

    it('should have retry button in error state', async () => {
      mockGetCollection.mockRejectedValue(new Error('Database error'));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Collection')).toBeInTheDocument();
      });

      // Verify retry button exists
      expect(screen.getByRole('button', { name: 'Retry' })).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no cards found', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse());
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats({ totalUniqueCards: 0, totalCards: 0 }));
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Cards Found')).toBeInTheDocument();
      });
    });
  });

  describe('Collection Display', () => {
    it('should render cards when collection exists', async () => {
      const mockCards = [
        createMockCollectionCard({ cardId: 1, name: 'Lightning Bolt' }),
        createMockCollectionCard({ cardId: 2, name: 'Counterspell', colors: ['U'], rarity: 'uncommon' }),
        createMockCollectionCard({ cardId: 3, name: 'Giant Growth', colors: ['G'], rarity: 'common' }),
      ];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
      expect(screen.getByText('Counterspell')).toBeInTheDocument();
      expect(screen.getByText('Giant Growth')).toBeInTheDocument();
    });

    it('should display page title', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });
    });

    it('should display collection stats', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats({
        totalUniqueCards: 150,
        totalCards: 600,
      }));
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('150')).toBeInTheDocument();
      });
      expect(screen.getByText('Unique Cards')).toBeInTheDocument();
      expect(screen.getByText('600')).toBeInTheDocument();
      expect(screen.getByText('Total Cards')).toBeInTheDocument();
    });

    it('should display card quantity badge', async () => {
      const mockCards = [createMockCollectionCard({ quantity: 4 })];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('x4')).toBeInTheDocument();
      });
    });

    it('should display set code on cards', async () => {
      const mockCards = [createMockCollectionCard({ setCode: 'dsk' })];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('DSK')).toBeInTheDocument();
      });
    });
  });

  describe('Filters', () => {
    it('should have search input', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search by name...')).toBeInTheDocument();
      });
    });

    it('should have set filter dropdown', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([
        createMockSetInfo({ code: 'sta', name: 'Strixhaven Mystical Archive' }),
        createMockSetInfo({ code: 'dsk', name: 'Duskmourn' }),
      ]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('All Sets')).toBeInTheDocument();
      });
    });

    it('should have rarity filter dropdown', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('All Rarities')).toBeInTheDocument();
      });
    });

    it('should have color filter buttons', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Colors:')).toBeInTheDocument();
      });
      // Check for color buttons by their title attributes
      expect(screen.getByTitle('White')).toBeInTheDocument();
      expect(screen.getByTitle('Blue')).toBeInTheDocument();
      expect(screen.getByTitle('Black')).toBeInTheDocument();
      expect(screen.getByTitle('Red')).toBeInTheDocument();
      expect(screen.getByTitle('Green')).toBeInTheDocument();
    });

    it('should have owned only checkbox', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Owned only')).toBeInTheDocument();
      });
    });

    it('should toggle color filter when clicking color button', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByTitle('Red')).toBeInTheDocument();
      });

      const redButton = screen.getByTitle('Red');
      expect(redButton).not.toHaveClass('active');

      fireEvent.click(redButton);

      await waitFor(() => {
        expect(redButton).toHaveClass('active');
      });
    });

    it('should display result count', async () => {
      const mockCards = [createMockCollectionCard()];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards, 100, 50));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Showing 50 of 100 cards')).toBeInTheDocument();
      });
    });

    it('should call GetCollection when search term changes', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search by name...')).toBeInTheDocument();
      });

      const searchInput = screen.getByPlaceholderText('Search by name...');
      fireEvent.change(searchInput, { target: { value: 'Bolt' } });

      // Wait for debounce
      await vi.advanceTimersByTimeAsync(350);

      await waitFor(() => {
        expect(mockGetCollection).toHaveBeenCalledTimes(2);
      });
    });
  });

  describe('Pagination', () => {
    it('should show pagination when multiple pages exist', async () => {
      const mockCards = [createMockCollectionCard()];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards, 100, 100));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText(/Page 1 of/)).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: 'First' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Previous' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Next' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Last' })).toBeInTheDocument();
    });

    it('should disable first/previous buttons on first page', async () => {
      const mockCards = [createMockCollectionCard()];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards, 100, 100));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'First' })).toBeDisabled();
      });
      expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled();
    });

    it('should navigate to next page when clicking next', async () => {
      const mockCards = [createMockCollectionCard()];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards, 100, 100));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Page 1 of 2')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Next' }));

      await waitFor(() => {
        expect(screen.getByText('Page 2 of 2')).toBeInTheDocument();
      });
    });

    it('should not show pagination when only one page', async () => {
      const mockCards = [createMockCollectionCard()];
      mockGetCollection.mockResolvedValue(createMockCollectionResponse(mockCards, 10, 10));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
      expect(screen.queryByText(/Page/)).not.toBeInTheDocument();
    });
  });

  describe('Sort Options', () => {
    it('should have sort dropdown', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Name (A-Z)')).toBeInTheDocument();
      });
    });

    it('should have all sort options', async () => {
      mockGetCollection.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockGetCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockGetAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Name (A-Z)')).toBeInTheDocument();
      });

      // Find the sort dropdown by its default value
      const sortSelect = screen.getByDisplayValue('Name (A-Z)') as HTMLSelectElement;
      const options = Array.from(sortSelect.options).map((opt) => opt.text);

      expect(options).toContain('Name (A-Z)');
      expect(options).toContain('Name (Z-A)');
      expect(options).toContain('Quantity (High)');
      expect(options).toContain('Quantity (Low)');
      expect(options).toContain('Rarity (High)');
      expect(options).toContain('Rarity (Low)');
      expect(options).toContain('CMC (Low)');
      expect(options).toContain('CMC (High)');
    });
  });
});

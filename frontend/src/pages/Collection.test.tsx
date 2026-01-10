import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
import { gui } from '@/types/models';
import { mockCollection, mockCards as mockCardsApi } from '@/test/mocks/apiMock';
import { renderWithRouter } from '@/test/utils/testUtils';
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

// Helper to create mock collection response
function createMockCollectionResponse(cards: gui.CollectionCard[]) {
  return {
    cards,
    totalCount: cards.length,
    filterCount: cards.length,
    unknownCardsRemaining: 0,
    unknownCardsFetched: 0,
  };
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
      let resolvePromise: (value: ReturnType<typeof createMockCollectionResponse>) => void;
      const loadingPromise = new Promise<ReturnType<typeof createMockCollectionResponse>>((resolve) => {
        resolvePromise = resolve;
      });
      mockCollection.getCollectionWithMetadata.mockReturnValue(loadingPromise);
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      expect(screen.getByText('Loading collection...')).toBeInTheDocument();

      resolvePromise!(createMockCollectionResponse([]));
      await waitFor(() => {
        expect(screen.queryByText('Loading collection...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockCollection.getCollectionWithMetadata.mockRejectedValue(new Error('Database error'));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Collection')).toBeInTheDocument();
      });
      expect(screen.getByText('Database error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockCollection.getCollectionWithMetadata.mockRejectedValue('Unknown error');
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Collection')).toBeInTheDocument();
      });
      expect(screen.getByText('Failed to load collection')).toBeInTheDocument();
    });

    it('should have retry button in error state', async () => {
      mockCollection.getCollectionWithMetadata.mockRejectedValue(new Error('Database error'));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats({ totalUniqueCards: 0, totalCards: 0 }));
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Cards are displayed as images with alt text containing the card name
      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'Lightning Bolt' })).toBeInTheDocument();
      });
      expect(screen.getByRole('img', { name: 'Counterspell' })).toBeInTheDocument();
      expect(screen.getByRole('img', { name: 'Giant Growth' })).toBeInTheDocument();
    });

    it('should display page title', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });
    });

    it('should display collection stats', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats({
        totalUniqueCards: 150,
        totalCards: 600,
      }));
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        // Cards in Set shows filterCount, Total Cards shows totalCount from response
        expect(screen.getByText('Cards in Set:')).toBeInTheDocument();
        expect(screen.getByText('Total Cards:')).toBeInTheDocument();
      });
    });

    it('should display card without quantity badge', async () => {
      const mockCards = [createMockCollectionCard({ quantity: 4, name: 'Test Card' })];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        // Card should render without quantity badge
        const cardImage = screen.getByRole('img', { name: 'Test Card' });
        expect(cardImage).toBeInTheDocument();
        // Quantity badge should not exist
        expect(screen.queryByText('x4')).not.toBeInTheDocument();
      });
    });

    it('should render card images with correct src', async () => {
      const mockCards = [createMockCollectionCard({
        name: 'Test Card',
        imageUri: 'https://cards.scryfall.io/normal/front/1/2/test.jpg'
      })];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        const img = screen.getByRole('img', { name: 'Test Card' });
        expect(img).toHaveAttribute('src', 'https://cards.scryfall.io/normal/front/1/2/test.jpg');
      });
    });
  });

  describe('Filters', () => {
    it('should have search input', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search by name...')).toBeInTheDocument();
      });
    });

    it('should have set filter dropdown', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([
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
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('All Rarities')).toBeInTheDocument();
      });
    });

    it('should have color filter buttons', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Owned only')).toBeInTheDocument();
      });
    });

    it('should toggle color filter when clicking color button', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        // With REST API, totalCount equals filterCount (both from array length)
        expect(screen.getByText('Showing 1 of 1 cards')).toBeInTheDocument();
      });
    });

    it('should filter client-side when search term changes (no API call)', async () => {
      // Create multiple cards to test client-side filtering
      const mockCards = [
        createMockCollectionCard({ cardId: 1, arenaId: 1, name: 'Lightning Bolt' }),
        createMockCollectionCard({ cardId: 2, arenaId: 2, name: 'Counterspell' }),
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search by name...')).toBeInTheDocument();
      });

      // Verify both cards are shown initially
      expect(screen.getByAltText('Lightning Bolt')).toBeInTheDocument();
      expect(screen.getByAltText('Counterspell')).toBeInTheDocument();

      const searchInput = screen.getByPlaceholderText('Search by name...');
      fireEvent.change(searchInput, { target: { value: 'Bolt' } });

      // Wait for debounce
      await vi.advanceTimersByTimeAsync(350);

      // API should only be called once (on mount) - search is client-side
      expect(mockCollection.getCollectionWithMetadata).toHaveBeenCalledTimes(1);
    });
  });

  describe('Pagination', () => {
    // Create >50 cards to trigger pagination (ITEMS_PER_PAGE = 50)
    function createManyCards(count: number): gui.CollectionCard[] {
      return Array.from({ length: count }, (_, i) =>
        createMockCollectionCard({ cardId: i + 1, arenaId: i + 1, name: `Card ${i + 1}` })
      );
    }

    it('should show pagination when multiple pages exist', async () => {
      const mockCards = createManyCards(75); // 2 pages
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      const mockCards = createManyCards(75); // 2 pages
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'First' })).toBeDisabled();
      });
      expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled();
    });

    it('should navigate to next page when clicking next', async () => {
      const mockCards = createManyCards(75); // 2 pages
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'Lightning Bolt' })).toBeInTheDocument();
      });
      expect(screen.queryByText(/Page/)).not.toBeInTheDocument();
    });
  });

  describe('Card Image Handling', () => {
    it('should use imageUri directly from card data', async () => {
      const mockCards = [
        createMockCollectionCard({
          cardId: 1,
          name: 'Test Card',
          imageUri: 'https://cards.scryfall.io/normal/front/1/2/test-card.jpg',
        }),
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        const img = screen.getByRole('img', { name: 'Test Card' });
        expect(img).toHaveAttribute('src', 'https://cards.scryfall.io/normal/front/1/2/test-card.jpg');
      });
    });

    it('should show card info fallback when imageUri is empty', async () => {
      const mockCards = [
        createMockCollectionCard({
          cardId: 1,
          name: 'Unknown Card',
          imageUri: '',
          setCode: 'TST',
          rarity: 'rare',
        }),
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        // Should show card info instead of placeholder image
        expect(screen.getByText('Unknown Card')).toBeInTheDocument();
        expect(screen.getByText('TST')).toBeInTheDocument();
        expect(screen.getByText('rare')).toBeInTheDocument();
        // No image should be present
        expect(screen.queryByRole('img', { name: 'Unknown Card' })).not.toBeInTheDocument();
      });
    });
  });

  describe('Set Completion Panel', () => {
    it('should not show Set Completion button when no set is selected (#756)', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Button should not be visible when no set is selected
      expect(screen.queryByRole('button', { name: 'Show Set Completion' })).not.toBeInTheDocument();
    });

    it('should show Set Completion button when a set is selected (#756)', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Select a set from the dropdown
      const setSelect = screen.getByDisplayValue('All Sets');
      fireEvent.change(setSelect, { target: { value: 'sta' } });

      await vi.advanceTimersByTimeAsync(100);

      // Button should now be visible
      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Show Set Completion' })).toBeInTheDocument();
      });
    });

    it('should toggle Set Completion panel visibility when set is selected', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);
      mockCollection.getSetCompletion.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Wait for collection to load
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Select a set
      const setSelect = screen.getByDisplayValue('All Sets');
      fireEvent.change(setSelect, { target: { value: 'sta' } });

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Show Set Completion' })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Show Set Completion' }));

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Hide Set Completion' })).toBeInTheDocument();
      });
    });

    it('should display Set Completion panel content when button is clicked (#756)', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);
      mockCollection.getSetCompletion.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Wait for collection to load
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Select a set first
      const setSelect = screen.getByDisplayValue('All Sets');
      fireEvent.change(setSelect, { target: { value: 'sta' } });

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Show Set Completion' })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: 'Show Set Completion' }));

      await waitFor(() => {
        // Verify the Set Completion panel heading is visible
        expect(screen.getByRole('heading', { name: 'Set Completion' })).toBeInTheDocument();
      });
    });

    it('should hide Set Completion panel when Hide button is clicked', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([createMockSetInfo()]);
      mockCollection.getSetCompletion.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Wait for collection to load
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Select a set first
      const setSelect = screen.getByDisplayValue('All Sets');
      fireEvent.change(setSelect, { target: { value: 'sta' } });

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Show Set Completion' })).toBeInTheDocument();
      });

      // Open the panel
      fireEvent.click(screen.getByRole('button', { name: 'Show Set Completion' }));

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Set Completion' })).toBeInTheDocument();
      });

      // Close the panel
      fireEvent.click(screen.getByRole('button', { name: 'Hide Set Completion' }));

      await waitFor(() => {
        expect(screen.queryByRole('heading', { name: 'Set Completion' })).not.toBeInTheDocument();
      });
    });
  });

  describe('Card Display Features', () => {
    it('should display card with not-owned class when quantity is 0', async () => {
      const mockCards = [createMockCollectionCard({ quantity: 0 })];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'Lightning Bolt' })).toBeInTheDocument();
      });

      const card = screen.getByRole('img', { name: 'Lightning Bolt' }).closest('.collection-card');
      expect(card).toHaveClass('not-owned');
    });

    it('should not show quantity badge for unowned cards', async () => {
      const mockCards = [createMockCollectionCard({ quantity: 0, name: 'Unowned Card' })];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        // Card should render with not-owned class but without quantity badge
        const card = screen.getByRole('img', { name: 'Unowned Card' }).closest('.collection-card');
        expect(card).toHaveClass('not-owned');
        expect(screen.queryByText('x0')).not.toBeInTheDocument();
      });
    });
  });

  describe('Empty State Variations', () => {
    it('should show filter adjustment suggestion when filters are active', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats({ totalUniqueCards: 100, totalCards: 100 }));
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Cards Found')).toBeInTheDocument();
      });

      // Toggle a color filter
      const redButton = screen.getByTitle('Red');
      fireEvent.click(redButton);

      await vi.advanceTimersByTimeAsync(350);

      await waitFor(() => {
        expect(screen.getByText('Try adjusting your filters')).toBeInTheDocument();
      });
    });

    it('should show "start playing" message when collection is truly empty', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats({ totalUniqueCards: 0, totalCards: 0 }));
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Your collection is empty. Start playing to add cards!')).toBeInTheDocument();
      });
    });
  });

  describe('Null/Undefined API Response Handling', () => {
    it('should handle null collection response gracefully', async () => {
      // Simulate API returning null (cast to bypass type check - this is what we're testing)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      mockCollection.getCollectionWithMetadata.mockResolvedValue(null as any);
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Cards Found')).toBeInTheDocument();
      });
      // Should not crash
      expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
    });

    it('should handle undefined collection response gracefully', async () => {
      // Simulate API returning undefined (cast to bypass type check - this is what we're testing)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      mockCollection.getCollectionWithMetadata.mockResolvedValue(undefined as any);
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Cards Found')).toBeInTheDocument();
      });
      // Should not crash
      expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
    });

    it('should handle non-array collection response gracefully', async () => {
      // API might return an object instead of array (cast to bypass type check - this is what we're testing)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      mockCollection.getCollectionWithMetadata.mockResolvedValue({ error: 'invalid' } as any);
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Cards Found')).toBeInTheDocument();
      });
      // Should not crash
      expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
    });

    it('should handle null sets response gracefully', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      // Simulate API returning null (cast to bypass type check - this is what we're testing)
      mockCardsApi.getAllSetInfo.mockResolvedValue(null as unknown as gui.SetInfo[]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'Lightning Bolt' })).toBeInTheDocument();
      });
      // Set dropdown should still render with just "All Sets"
      expect(screen.getByText('All Sets')).toBeInTheDocument();
    });

    it('should handle undefined sets response gracefully', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      // Simulate API returning undefined (cast to bypass type check - this is what we're testing)
      mockCardsApi.getAllSetInfo.mockResolvedValue(undefined as unknown as gui.SetInfo[]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'Lightning Bolt' })).toBeInTheDocument();
      });
      // Set dropdown should still render with just "All Sets"
      expect(screen.getByText('All Sets')).toBeInTheDocument();
    });
  });

  describe('Sort Options', () => {
    it('should have sort dropdown', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Name (A-Z)')).toBeInTheDocument();
      });
    });

    it('should have all sort options including price', async () => {
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse([createMockCollectionCard()]));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

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
      expect(options).toContain('Price (High)');
      expect(options).toContain('Price (Low)');
    });

    it('should sort cards by price when price sort is selected', async () => {
      const mockCards = [
        createMockCollectionCard({ cardId: 1, arenaId: 1, name: 'Cheap Card', priceUsd: 0.25 }),
        createMockCollectionCard({ cardId: 2, arenaId: 2, name: 'Expensive Card', priceUsd: 10.00 }),
        createMockCollectionCard({ cardId: 3, arenaId: 3, name: 'Medium Card', priceUsd: 2.50 }),
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByDisplayValue('Name (A-Z)')).toBeInTheDocument();
      });

      // Change sort to Price (High)
      const sortSelect = screen.getByDisplayValue('Name (A-Z)');
      fireEvent.change(sortSelect, { target: { value: 'price-desc' } });

      await vi.advanceTimersByTimeAsync(100);

      // With Price (High) sort, Expensive Card should be first in the card grid
      // Filter to only card images (excluding color icons)
      await waitFor(() => {
        const expensiveCard = screen.getByRole('img', { name: 'Expensive Card' });
        const cheapCard = screen.getByRole('img', { name: 'Cheap Card' });
        // Expensive should appear before Cheap in DOM order
        expect(expensiveCard.compareDocumentPosition(cheapCard) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
      });
    });
  });

  describe('Price Display Features', () => {
    it('should display price badge on cards with price', async () => {
      const mockCards = [
        createMockCollectionCard({ cardId: 1, name: 'Priced Card', priceUsd: 5.99 }),
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('$5.99')).toBeInTheDocument();
      });
    });

    it('should not display price badge when price is 0', async () => {
      const mockCards = [
        createMockCollectionCard({ cardId: 1, name: 'Free Card', priceUsd: 0 }),
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'Free Card' })).toBeInTheDocument();
      });

      expect(screen.queryByText('$0.00')).not.toBeInTheDocument();
    });

    it('should not display price badge when price is undefined', async () => {
      const mockCards = [
        createMockCollectionCard({ cardId: 1, name: 'No Price Card' }), // priceUsd not set
      ];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('img', { name: 'No Price Card' })).toBeInTheDocument();
      });

      // Should not have any price badge
      expect(screen.queryByText(/^\$/)).not.toBeInTheDocument();
    });

    it('should display collection value in header when value > 0', async () => {
      const mockCards = [createMockCollectionCard()];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);
      mockCollection.getCollectionValue.mockResolvedValue({
        totalValueUsd: 1234.56,
        totalValueEur: 1100.00,
        uniqueCardsWithPrice: 50,
        cardCount: 100,
        valueByRarity: {},
        topCards: [],
      });

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Est. Value:')).toBeInTheDocument();
      });
      expect(screen.getByText('$1,234.56')).toBeInTheDocument();
    });

    it('should not display collection value when value is 0', async () => {
      const mockCards = [createMockCollectionCard()];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);
      mockCollection.getCollectionValue.mockResolvedValue({
        totalValueUsd: 0,
        totalValueEur: 0,
        uniqueCardsWithPrice: 0,
        cardCount: 0,
        valueByRarity: {},
        topCards: [],
      });

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      expect(screen.queryByText('Est. Value:')).not.toBeInTheDocument();
    });

    it('should handle collection value API error gracefully', async () => {
      const mockCards = [createMockCollectionCard()];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);
      mockCollection.getCollectionValue.mockRejectedValue(new Error('API error'));

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Should render collection without crashing
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Value should not be shown
      expect(screen.queryByText('Est. Value:')).not.toBeInTheDocument();
    });
  });

  describe('Scryfall Auto-Fetch Behavior (#784)', () => {
    it('should not auto-refresh when no cards are fetched (all lookups failed)', async () => {
      // Simulate the case where all Scryfall lookups fail
      const mockCards = [createMockCollectionCard()];
      mockCollection.getCollectionWithMetadata.mockResolvedValue({
        ...createMockCollectionResponse(mockCards),
        unknownCardsFetched: 0, // No cards successfully fetched
        unknownCardsRemaining: 5, // Cards that failed lookup (now on cooldown)
      });
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Wait longer than auto-refresh timeout
      await vi.advanceTimersByTimeAsync(1000);

      // API should only be called once (no auto-refresh since unknownFetched is 0)
      expect(mockCollection.getCollectionWithMetadata).toHaveBeenCalledTimes(1);
    });

    it('should not trigger multiple simultaneous API calls', async () => {
      const mockCards = [createMockCollectionCard()];
      mockCollection.getCollectionWithMetadata.mockResolvedValue(createMockCollectionResponse(mockCards));
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      // Don't wait - check that only one call was made even with rapid mounting
      await vi.advanceTimersByTimeAsync(10);
      await vi.advanceTimersByTimeAsync(10);
      await vi.advanceTimersByTimeAsync(10);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Should be called exactly once on mount, not multiple times
      expect(mockCollection.getCollectionWithMetadata).toHaveBeenCalledTimes(1);
    });

    it('should auto-refresh when cards are successfully fetched and more remain', async () => {
      let callCount = 0;
      const mockCards = [createMockCollectionCard()];

      // First call: cards fetched, more remaining
      // Second call: no more cards to fetch
      mockCollection.getCollectionWithMetadata.mockImplementation(async () => {
        callCount++;
        if (callCount === 1) {
          return {
            ...createMockCollectionResponse(mockCards),
            unknownCardsFetched: 5, // Cards successfully fetched
            unknownCardsRemaining: 3, // More cards to fetch
          };
        }
        return {
          ...createMockCollectionResponse(mockCards),
          unknownCardsFetched: 3, // Final batch fetched
          unknownCardsRemaining: 0, // No more cards
        };
      });
      mockCollection.getCollectionStats.mockResolvedValue(createMockCollectionStats());
      mockCardsApi.getAllSetInfo.mockResolvedValue([]);

      renderWithRouter(<Collection />);

      await vi.advanceTimersByTimeAsync(100);

      // Wait for initial load
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'Collection' })).toBeInTheDocument();
      });

      // Wait for auto-refresh timeout (500ms + some buffer)
      await vi.advanceTimersByTimeAsync(600);

      await waitFor(() => {
        // Should be called twice (initial + auto-refresh)
        expect(mockCollection.getCollectionWithMetadata).toHaveBeenCalledTimes(2);
      });
    });
  });
});

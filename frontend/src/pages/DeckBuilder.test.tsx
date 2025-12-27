import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import DeckBuilder from './DeckBuilder';
import { mockDecks } from '@/test/mocks/apiMock';
import { models, gui } from '@/types/models';

// Mock download utility
vi.mock('@/utils/download', () => ({
  downloadTextFile: vi.fn(),
}));

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: vi.fn(() => ({ deckID: 'test-deck-id' })),
    useNavigate: vi.fn(() => mockNavigate),
  };
});

// Helper function to create mock deck
function createMockDeck(overrides: Partial<models.Deck> = {}): models.Deck {
  return new models.Deck({
    ID: 'test-deck-id',
    Name: 'Test Deck',
    Format: 'limited',
    Source: 'draft',
    DraftEventID: 'draft-event-123',
    Colors: ['W', 'U'],
    CreatedAt: new Date('2025-11-20T10:00:00Z'),
    UpdatedAt: new Date('2025-11-20T10:00:00Z'),
    ...overrides,
  });
}

// Helper function to create mock deck cards
function createMockDeckCard(overrides: Partial<models.DeckCard> = {}): models.DeckCard {
  return new models.DeckCard({
    ID: 1,
    DeckID: 'test-deck-id',
    CardID: 12345,
    Quantity: 1,
    Board: 'main',
    ...overrides,
  });
}

// Helper function to create mock deck statistics
function createMockDeckStatistics(overrides: Partial<gui.DeckStatistics> = {}): gui.DeckStatistics {
  return new gui.DeckStatistics({
    totalCards: 15,
    totalMainboard: 15,
    totalSideboard: 0,
    averageCMC: 2.8,
    manaCurve: { 0: 0, 1: 3, 2: 5, 3: 4, 4: 2, 5: 1 },
    maxCMC: 5,
    colors: {
      white: 5,
      blue: 5,
      black: 0,
      red: 0,
      green: 5,
      colorless: 0,
      multicolor: 0,
    },
    types: {
      creatures: 10,
      instants: 2,
      sorceries: 2,
      enchantments: 1,
      artifacts: 0,
      planeswalkers: 0,
      lands: 0,
      other: 0,
    },
    lands: {
      total: 0,
      basic: 0,
      nonBasic: 0,
      ratio: 0,
      recommended: 15,
      status: 'low',
      statusMessage: 'Add more lands',
    },
    creatures: {
      total: 10,
      averagePower: 2.5,
      averageToughness: 2.5,
      totalPower: 25,
      totalToughness: 25,
    },
    legality: {
      standard: true,
      historic: true,
      explorer: true,
      pioneer: true,
      modern: true,
      legacy: true,
      vintage: true,
      pauper: false,
      commander: true,
      brawl: true,
    },
    ...overrides,
  });
}

describe('DeckBuilder Component - Export and Validate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockClear();
  });

  describe('Export Deck Functionality', () => {
    it('should call exportDeck when Export button is clicked', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockDecks.getDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockDecks.getDeckStatistics.mockResolvedValue(mockStats);
      mockDecks.exportDeck.mockResolvedValue({ content: 'deck content', filename: 'test.txt' });

      render(<DeckBuilder />);

      // Wait for deck to load
      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      // Find and click Export button
      const exportButton = screen.getByRole('button', { name: /Export/i });
      await userEvent.click(exportButton);

      // Verify exportDeck was called with correct deck ID and format
      await waitFor(() => {
        expect(mockDecks.exportDeck).toHaveBeenCalledWith('test-deck-id', expect.any(Object));
      });
    });

    it('should handle export exception gracefully', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockDecks.getDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockDecks.getDeckStatistics.mockResolvedValue(mockStats);
      mockDecks.exportDeck.mockRejectedValue(new Error('Failed to export'));

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      const exportButton = screen.getByRole('button', { name: /Export/i });
      await userEvent.click(exportButton);

      // Error is logged to console, verify export was attempted
      await waitFor(() => {
        expect(mockDecks.exportDeck).toHaveBeenCalled();
      });
    });

    it('should not export when no deck is loaded', async () => {
      mockDecks.getDeck.mockResolvedValue(null);

      // Mock window.alert
      const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText(/Error Loading Deck/i)).toBeInTheDocument();
      });

      // Export button should not be visible when there's an error
      expect(screen.queryByRole('button', { name: /Export/i })).not.toBeInTheDocument();

      alertSpy.mockRestore();
    });
  });

  describe('Validate Deck Functionality', () => {
    it('should call validateDraftDeck when Validate button is clicked', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockDecks.getDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockDecks.getDeckStatistics.mockResolvedValue(mockStats);
      mockDecks.validateDraftDeck.mockResolvedValue(true);

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      // Find and click Validate button
      const validateButton = screen.getByRole('button', { name: /Validate/i });
      await userEvent.click(validateButton);

      // Verify validateDraftDeck was called with correct deck ID
      await waitFor(() => {
        expect(mockDecks.validateDraftDeck).toHaveBeenCalledWith('test-deck-id');
      });
    });

    it('should handle validation exception gracefully', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockDecks.getDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockDecks.getDeckStatistics.mockResolvedValue(mockStats);
      mockDecks.validateDraftDeck.mockRejectedValue(new Error('Validation service unavailable'));

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      const validateButton = screen.getByRole('button', { name: /Validate/i });
      await userEvent.click(validateButton);

      // Error is logged to console, verify validation was attempted
      await waitFor(() => {
        expect(mockDecks.validateDraftDeck).toHaveBeenCalled();
      });
    });
  });

  describe('Button Rendering', () => {
    it('should render both Export and Validate buttons in footer', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockDecks.getDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      // Verify both buttons are present
      expect(screen.getByRole('button', { name: /Export/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /Validate/i })).toBeInTheDocument();
    });

    it('should have correct button titles', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockDecks.getDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      const exportButton = screen.getByRole('button', { name: /Export/i });
      const validateButton = screen.getByRole('button', { name: /Validate/i });

      expect(exportButton).toHaveAttribute('title', 'Export deck');
      expect(validateButton).toHaveAttribute('title', 'Validate deck');
    });
  });
});

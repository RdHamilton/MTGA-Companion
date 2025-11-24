import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import DeckBuilder from './DeckBuilder';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { mockWailsRuntime } from '../test/mocks/wailsRuntime';
import { models, gui } from '../../wailsjs/go/models';

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
function createMockDeckStatistics(overrides: any = {}): any {
  return {
    totalMainboard: 15,
    totalSideboard: 0,
    averageCMC: 2.8,
    colors: {
      white: 5,
      blue: 5,
      black: 0,
      red: 0,
      green: 5,
    },
    lands: {
      total: 0,
      recommended: 15,
    },
    ...overrides,
  };
}

describe('DeckBuilder Component - Export and Validate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockClear();
  });

  describe('Export Deck Functionality', () => {
    it('should call ExportDeckToFile when Export button is clicked', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockWailsApp.GetDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockWailsApp.GetDeckStatistics.mockResolvedValue(mockStats);
      mockWailsApp.ExportDeckToFile.mockResolvedValue();

      render(<DeckBuilder />);

      // Wait for deck to load
      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      // Find and click Export button
      const exportButton = screen.getByRole('button', { name: /Export/i });
      await userEvent.click(exportButton);

      // Verify ExportDeckToFile was called with correct deck ID
      // Backend handles the native file dialog and saving
      await waitFor(() => {
        expect(mockWailsApp.ExportDeckToFile).toHaveBeenCalledWith('test-deck-id');
      });
    });

    it('should handle export exception gracefully', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockWailsApp.GetDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockWailsApp.GetDeckStatistics.mockResolvedValue(mockStats);
      mockWailsApp.ExportDeckToFile.mockRejectedValue(new Error('Failed to show file dialog'));

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      const exportButton = screen.getByRole('button', { name: /Export/i });
      await userEvent.click(exportButton);

      // Error is logged to console, no alert shown to user (backend handles dialogs)
      await waitFor(() => {
        expect(mockWailsApp.ExportDeckToFile).toHaveBeenCalled();
      });
    });

    it('should not export when no deck is loaded', async () => {
      mockWailsApp.GetDeck.mockResolvedValue(null);

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
    it('should call ValidateDeckWithDialog when Validate button is clicked', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockWailsApp.GetDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockWailsApp.GetDeckStatistics.mockResolvedValue(mockStats);
      mockWailsApp.ValidateDeckWithDialog.mockResolvedValue();

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      // Find and click Validate button
      const validateButton = screen.getByRole('button', { name: /Validate/i });
      await userEvent.click(validateButton);

      // Verify ValidateDeckWithDialog was called with correct deck ID
      // Backend handles validation and shows native dialog
      await waitFor(() => {
        expect(mockWailsApp.ValidateDeckWithDialog).toHaveBeenCalledWith('test-deck-id');
      });
    });

    it('should handle validation exception gracefully', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockWailsApp.GetDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockWailsApp.GetDeckStatistics.mockResolvedValue(mockStats);
      mockWailsApp.ValidateDeckWithDialog.mockRejectedValue(new Error('Validation service unavailable'));

      render(<DeckBuilder />);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });

      const validateButton = screen.getByRole('button', { name: /Validate/i });
      await userEvent.click(validateButton);

      // Error is logged to console, no alert shown to user (backend handles dialogs)
      await waitFor(() => {
        expect(mockWailsApp.ValidateDeckWithDialog).toHaveBeenCalled();
      });
    });
  });

  describe('Button Rendering', () => {
    it('should render both Export and Validate buttons in footer', async () => {
      const mockDeck = createMockDeck();
      const mockCards = [createMockDeckCard()];
      const mockStats = createMockDeckStatistics();

      mockWailsApp.GetDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockWailsApp.GetDeckStatistics.mockResolvedValue(mockStats);

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

      mockWailsApp.GetDeck.mockResolvedValue({
        deck: mockDeck,
        cards: mockCards,
        tags: [],
      });
      mockWailsApp.GetDeckStatistics.mockResolvedValue(mockStats);

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

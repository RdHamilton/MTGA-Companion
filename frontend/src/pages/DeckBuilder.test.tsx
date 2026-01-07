import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import DeckBuilder from './DeckBuilder';
import { mockDecks, mockDrafts, mockCards } from '@/test/mocks/apiMock';
import { ApiRequestError } from '@/services/apiClient';
import { models, gui } from '@/types/models';

// Mock download utility
vi.mock('@/utils/download', () => ({
  downloadTextFile: vi.fn(),
}));

// Mock react-router-dom with configurable params
const mockNavigate = vi.fn();
let mockParams: { deckID?: string; draftEventID?: string } = { deckID: 'test-deck-id' };

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: vi.fn(() => mockParams),
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

// Helper function to create mock draft session
function createMockDraftSession(overrides: Partial<models.DraftSession> = {}): models.DraftSession {
  return new models.DraftSession({
    ID: 'draft-event-123',
    EventName: 'QuickDraft_FDN',
    DraftType: 'QuickDraft',
    SetCode: 'FDN',
    Status: 'completed',
    StartTime: new Date('2025-11-20T10:00:00Z'),
    EndTime: new Date('2025-11-20T11:00:00Z'),
    ...overrides,
  });
}

// Helper function to create mock draft pick
function createMockDraftPick(overrides: Partial<models.DraftPickSession> = {}): models.DraftPickSession {
  return new models.DraftPickSession({
    ID: 1,
    SessionID: 'draft-event-123',
    CardID: '12345',
    PickNumber: 1,
    PackNumber: 1,
    ...overrides,
  });
}

// Helper function to create mock SetCard (card metadata)
function createMockSetCard(overrides: Partial<models.SetCard> = {}): models.SetCard {
  return new models.SetCard({
    ArenaID: 12345,
    Name: 'Test Card',
    SetCode: 'TST',
    CMC: 2,
    Types: ['Creature'],
    Colors: ['W'],
    Rarity: 'common',
    ...overrides,
  });
}

describe('DeckBuilder Component - Export and Validate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockClear();
    // Reset mockParams to default (load by deck ID)
    mockParams = { deckID: 'test-deck-id' };
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

  describe('Build Around Functionality', () => {
    it('should NOT show Build Around button for draft decks', async () => {
      const mockDeck = createMockDeck({ Source: 'draft', DraftEventID: 'draft-123' });
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

      // Build Around should NOT be visible for draft decks
      expect(screen.queryByRole('button', { name: /Build Around/i })).not.toBeInTheDocument();
      // But Suggest Decks should be visible for draft decks
      expect(screen.getByRole('button', { name: /Suggest Decks/i })).toBeInTheDocument();
    });

    it('should show Build Around button for non-draft decks', async () => {
      const mockDeck = createMockDeck({ Source: 'constructed', DraftEventID: '' });
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

      // Build Around should be visible for non-draft decks
      expect(screen.getByRole('button', { name: /Build Around/i })).toBeInTheDocument();
      // Suggest Decks should NOT be visible for non-draft decks
      expect(screen.queryByRole('button', { name: /Suggest Decks/i })).not.toBeInTheDocument();
    });

    it('should open Build Around modal when button is clicked', async () => {
      const mockDeck = createMockDeck({ Source: 'constructed', DraftEventID: '' });
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

      // Click Build Around button
      const buildAroundButton = screen.getByRole('button', { name: /Build Around/i });
      await userEvent.click(buildAroundButton);

      // Modal should be open (check for modal heading)
      await waitFor(() => {
        expect(screen.getByText('Build Around Card')).toBeInTheDocument();
      });
    });

    it('should close Build Around modal when close button is clicked', async () => {
      const mockDeck = createMockDeck({ Source: 'constructed', DraftEventID: '' });
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

      // Open modal
      const buildAroundButton = screen.getByRole('button', { name: /Build Around/i });
      await userEvent.click(buildAroundButton);

      await waitFor(() => {
        expect(screen.getByText('Build Around Card')).toBeInTheDocument();
      });

      // Close modal using close button (within the modal header)
      const modal = document.querySelector('.build-around-modal');
      const closeButton = modal!.querySelector('.close-button') as HTMLButtonElement;
      await userEvent.click(closeButton);

      // Modal should be closed
      await waitFor(() => {
        expect(screen.queryByText('Build Around Card')).not.toBeInTheDocument();
      });
    });

    it('should have correct title on Build Around button', async () => {
      const mockDeck = createMockDeck({ Source: 'constructed', DraftEventID: '' });
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

      const buildAroundButton = screen.getByRole('button', { name: /Build Around/i });
      expect(buildAroundButton).toHaveAttribute('title', 'Get suggestions based on current deck cards');
    });
  });
});

describe('DeckBuilder Component - Deck Creation from Draft', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockClear();
    // Configure for draft event ID route
    mockParams = { draftEventID: 'draft-event-123' };
  });

  it('should create deck from draft when no deck exists for draft event', async () => {
    const mockSession = createMockDraftSession();
    const mockPicks = [
      createMockDraftPick({ ID: 1, CardID: '11111', PickNumber: 1 }),
      createMockDraftPick({ ID: 2, CardID: '22222', PickNumber: 2 }),
      createMockDraftPick({ ID: 3, CardID: '33333', PickNumber: 3 }),
    ];
    const createdDeck = createMockDeck({
      ID: 'new-draft-deck-id',
      Name: 'QuickDraft_FDN Draft',
      Source: 'draft',
      DraftEventID: 'draft-event-123',
    });
    const mockDeckCards = [
      createMockDeckCard({ CardID: 11111 }),
      createMockDeckCard({ CardID: 22222, ID: 2 }),
      createMockDeckCard({ CardID: 33333, ID: 3 }),
    ];
    const mockStats = createMockDeckStatistics({ totalCards: 3 });

    // Mock getDeckByDraftEvent to throw 404 (no existing deck)
    mockDecks.getDeckByDraftEvent.mockRejectedValue(
      new ApiRequestError('Not found', 404)
    );

    // Mock draft session lookup
    mockDrafts.getActiveDraftSessions.mockResolvedValue([]);
    mockDrafts.getCompletedDraftSessions.mockResolvedValue([mockSession]);

    // Mock deck creation
    mockDecks.createDeck.mockResolvedValue(createdDeck);

    // Mock draft picks
    mockDrafts.getDraftPicks.mockResolvedValue(mockPicks);

    // Mock addCard
    mockDecks.addCard.mockResolvedValue(undefined);

    // Mock getDeck to return the deck with cards
    mockDecks.getDeck.mockResolvedValue({
      deck: createdDeck,
      cards: mockDeckCards,
      tags: [],
    });

    mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

    // Mock card metadata lookup
    mockCards.getCardByArenaId.mockImplementation((id: number) => {
      return Promise.resolve(createMockSetCard({ ArenaID: String(id), Name: `Card ${id}` }));
    });

    render(<DeckBuilder />);

    // Wait for deck to load and be displayed
    await waitFor(() => {
      expect(screen.getByText('QuickDraft_FDN Draft')).toBeInTheDocument();
    });

    // Verify getDeckByDraftEvent was called
    expect(mockDecks.getDeckByDraftEvent).toHaveBeenCalledWith('draft-event-123');

    // Verify createDeck was called with correct params
    expect(mockDecks.createDeck).toHaveBeenCalledWith({
      name: 'QuickDraft_FDN Draft',
      format: 'limited',
      source: 'draft',
      draft_event_id: 'draft-event-123',
    });

    // Verify draft picks were loaded
    expect(mockDrafts.getDraftPicks).toHaveBeenCalledWith('draft-event-123');

    // Verify cards were added to deck
    expect(mockDecks.addCard).toHaveBeenCalledTimes(3);
    expect(mockDecks.addCard).toHaveBeenCalledWith(
      expect.objectContaining({
        deck_id: 'new-draft-deck-id',
        arena_id: 11111,
        quantity: 1,
        zone: 'main',
      })
    );
  });

  it('should handle duplicate card picks correctly (count occurrences)', async () => {
    const mockSession = createMockDraftSession();
    // Same card picked twice
    const mockPicks = [
      createMockDraftPick({ ID: 1, CardID: '11111', PickNumber: 1 }),
      createMockDraftPick({ ID: 2, CardID: '11111', PickNumber: 2 }),
      createMockDraftPick({ ID: 3, CardID: '22222', PickNumber: 3 }),
    ];
    const createdDeck = createMockDeck({
      ID: 'new-draft-deck-id',
      Name: 'QuickDraft_FDN Draft',
      Source: 'draft',
      DraftEventID: 'draft-event-123',
    });
    const mockDeckCards = [
      createMockDeckCard({ CardID: 11111, Quantity: 2 }),
      createMockDeckCard({ CardID: 22222, ID: 2 }),
    ];
    const mockStats = createMockDeckStatistics({ totalCards: 3 });

    mockDecks.getDeckByDraftEvent.mockRejectedValue(
      new ApiRequestError('Not found', 404)
    );
    mockDrafts.getActiveDraftSessions.mockResolvedValue([]);
    mockDrafts.getCompletedDraftSessions.mockResolvedValue([mockSession]);
    mockDecks.createDeck.mockResolvedValue(createdDeck);
    mockDrafts.getDraftPicks.mockResolvedValue(mockPicks);
    mockDecks.addCard.mockResolvedValue(undefined);
    mockDecks.getDeck.mockResolvedValue({
      deck: createdDeck,
      cards: mockDeckCards,
      tags: [],
    });
    mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

    // Mock card metadata lookup
    mockCards.getCardByArenaId.mockImplementation((id: number) => {
      return Promise.resolve(createMockSetCard({ ArenaID: String(id), Name: `Card ${id}` }));
    });

    render(<DeckBuilder />);

    await waitFor(() => {
      expect(screen.getByText('QuickDraft_FDN Draft')).toBeInTheDocument();
    });

    // Verify addCard was called with quantity 2 for the duplicate card
    expect(mockDecks.addCard).toHaveBeenCalledTimes(2);
    expect(mockDecks.addCard).toHaveBeenCalledWith(
      expect.objectContaining({
        deck_id: 'new-draft-deck-id',
        arena_id: 11111,
        quantity: 2,
        zone: 'main',
      })
    );
    expect(mockDecks.addCard).toHaveBeenCalledWith(
      expect.objectContaining({
        deck_id: 'new-draft-deck-id',
        arena_id: 22222,
        quantity: 1,
        zone: 'main',
      })
    );
  });

  it('should load existing deck when deck already exists for draft event', async () => {
    const existingDeck = createMockDeck({
      ID: 'existing-draft-deck-id',
      Name: 'Existing Draft Deck',
      Source: 'draft',
      DraftEventID: 'draft-event-123',
    });
    const mockDeckCards = [createMockDeckCard()];
    const mockStats = createMockDeckStatistics();

    // Mock getDeckByDraftEvent to return existing deck
    mockDecks.getDeckByDraftEvent.mockResolvedValue({
      deck: existingDeck,
      cards: mockDeckCards,
      tags: [],
    });
    mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

    // Mock card metadata lookup
    mockCards.getCardByArenaId.mockImplementation((id: number) => {
      return Promise.resolve(createMockSetCard({ ArenaID: String(id), Name: `Card ${id}` }));
    });

    render(<DeckBuilder />);

    await waitFor(() => {
      expect(screen.getByText('Existing Draft Deck')).toBeInTheDocument();
    });

    // Verify we did NOT try to create a new deck
    expect(mockDecks.createDeck).not.toHaveBeenCalled();
    // Note: getDraftPicks might still be called for other purposes,
    // but addCard should NOT be called since we're loading an existing deck
    expect(mockDecks.addCard).not.toHaveBeenCalled();
  });

  it('should show error when draft session is not found', async () => {
    mockDecks.getDeckByDraftEvent.mockRejectedValue(
      new ApiRequestError('Not found', 404)
    );
    // Return empty arrays for both session queries
    mockDrafts.getActiveDraftSessions.mockResolvedValue([]);
    mockDrafts.getCompletedDraftSessions.mockResolvedValue([]);

    render(<DeckBuilder />);

    await waitFor(() => {
      expect(screen.getByText(/Draft session not found/i)).toBeInTheDocument();
    });

    // Verify no deck was created
    expect(mockDecks.createDeck).not.toHaveBeenCalled();
  });

  it('should handle empty draft picks gracefully', async () => {
    const mockSession = createMockDraftSession();
    const createdDeck = createMockDeck({
      ID: 'new-draft-deck-id',
      Name: 'QuickDraft_FDN Draft',
      Source: 'draft',
      DraftEventID: 'draft-event-123',
    });
    const mockStats = createMockDeckStatistics({ totalCards: 0 });

    mockDecks.getDeckByDraftEvent.mockRejectedValue(
      new ApiRequestError('Not found', 404)
    );
    mockDrafts.getActiveDraftSessions.mockResolvedValue([mockSession]);
    mockDrafts.getCompletedDraftSessions.mockResolvedValue([]);
    mockDecks.createDeck.mockResolvedValue(createdDeck);
    // Empty picks
    mockDrafts.getDraftPicks.mockResolvedValue([]);
    mockDecks.getDeck.mockResolvedValue({
      deck: createdDeck,
      cards: [],
      tags: [],
    });
    mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

    // Mock card metadata lookup (not needed for empty deck, but good practice)
    mockCards.getCardByArenaId.mockImplementation((id: number) => {
      return Promise.resolve(createMockSetCard({ ArenaID: String(id), Name: `Card ${id}` }));
    });

    render(<DeckBuilder />);

    await waitFor(() => {
      expect(screen.getByText('QuickDraft_FDN Draft')).toBeInTheDocument();
    });

    // Verify deck was created but no cards were added
    expect(mockDecks.createDeck).toHaveBeenCalled();
    expect(mockDrafts.getDraftPicks).toHaveBeenCalled();
    expect(mockDecks.addCard).not.toHaveBeenCalled();
  });

  it('should continue even if adding a card fails', async () => {
    const mockSession = createMockDraftSession();
    const mockPicks = [
      createMockDraftPick({ ID: 1, CardID: '11111', PickNumber: 1 }),
      createMockDraftPick({ ID: 2, CardID: '22222', PickNumber: 2 }),
    ];
    const createdDeck = createMockDeck({
      ID: 'new-draft-deck-id',
      Name: 'QuickDraft_FDN Draft',
      Source: 'draft',
      DraftEventID: 'draft-event-123',
    });
    const mockDeckCards = [createMockDeckCard({ CardID: 22222 })];
    const mockStats = createMockDeckStatistics({ totalCards: 1 });

    mockDecks.getDeckByDraftEvent.mockRejectedValue(
      new ApiRequestError('Not found', 404)
    );
    mockDrafts.getActiveDraftSessions.mockResolvedValue([]);
    mockDrafts.getCompletedDraftSessions.mockResolvedValue([mockSession]);
    mockDecks.createDeck.mockResolvedValue(createdDeck);
    mockDrafts.getDraftPicks.mockResolvedValue(mockPicks);

    // First addCard fails, second succeeds
    mockDecks.addCard
      .mockRejectedValueOnce(new Error('Failed to add card'))
      .mockResolvedValueOnce(undefined);

    mockDecks.getDeck.mockResolvedValue({
      deck: createdDeck,
      cards: mockDeckCards,
      tags: [],
    });
    mockDecks.getDeckStatistics.mockResolvedValue(mockStats);

    // Mock card metadata lookup
    mockCards.getCardByArenaId.mockImplementation((id: number) => {
      return Promise.resolve(createMockSetCard({ ArenaID: String(id), Name: `Card ${id}` }));
    });

    // Spy on console.error
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(<DeckBuilder />);

    await waitFor(() => {
      expect(screen.getByText('QuickDraft_FDN Draft')).toBeInTheDocument();
    });

    // Verify both addCard calls were made
    expect(mockDecks.addCard).toHaveBeenCalledTimes(2);

    // Verify error was logged
    expect(consoleErrorSpy).toHaveBeenCalledWith(
      expect.stringContaining('Failed to add card'),
      expect.any(Error)
    );

    consoleErrorSpy.mockRestore();
  });
});

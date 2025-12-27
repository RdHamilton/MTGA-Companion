import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import DeckList from './DeckList';
import { mockCards } from '@/test/mocks/apiMock';
import { models, gui } from '@/types/models';

// Helper function to create mock deck
function createMockDeck(overrides: Partial<models.Deck> = {}): models.Deck {
  return new models.Deck({
    ID: 'test-deck-id',
    Name: 'Test Deck',
    Format: 'limited',
    Source: 'manual',
    Colors: ['W', 'U'],
    CreatedAt: new Date('2025-11-20T10:00:00Z'),
    UpdatedAt: new Date('2025-11-20T10:00:00Z'),
    ...overrides,
  });
}

// Helper function to create mock deck card
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

// Helper function to create mock set card
function createMockSetCard(overrides: Partial<models.SetCard> = {}): models.SetCard {
  return new models.SetCard({
    ArenaID: '12345',
    Name: 'Test Card',
    SetCode: 'TST',
    CMC: 3,
    ManaCost: '{2}{U}',
    Types: ['Creature'],
    Colors: ['U'],
    Rarity: 'common',
    Power: '2',
    Toughness: '2',
    ...overrides,
  });
}

// Helper function to create mock deck statistics
function createMockStatistics(overrides: Partial<gui.DeckStatistics> = {}): gui.DeckStatistics {
  return new gui.DeckStatistics({
    totalCards: 40,
    totalMainboard: 40,
    totalSideboard: 0,
    averageCMC: 2.8,
    manaCurve: {
      0: 0,
      1: 5,
      2: 8,
      3: 10,
      4: 7,
      5: 5,
      6: 3,
      7: 2,
    },
    maxCMC: 7,
    colors: {
      white: 15,
      blue: 10,
      black: 0,
      red: 0,
      green: 5,
      colorless: 10,
      multicolor: 0,
    },
    types: {
      creatures: 20,
      instants: 5,
      sorceries: 5,
      enchantments: 2,
      artifacts: 1,
      planeswalkers: 0,
      lands: 17,
      other: 0,
    },
    lands: {
      total: 17,
      basic: 12,
      nonBasic: 5,
      ratio: 0.425,
      recommended: 17,
      status: 'good',
      statusMessage: 'Land count looks good',
    },
    creatures: {
      total: 20,
      averagePower: 2.5,
      averageToughness: 2.5,
      totalPower: 50,
      totalToughness: 50,
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

describe('DeckList Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock getAllSetInfo for SetSymbol component - return TST set
    mockCards.getAllSetInfo.mockResolvedValue([
      {
        code: 'TST',
        name: 'Test Set',
        iconSvgUri: 'https://example.com/tst.svg',
        setType: 'expansion',
        releasedAt: '2024-01-01',
        cardCount: 100,
      },
    ]);
  });

  describe('Loading State', () => {
    it('should display loading state while fetching card metadata', () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];

      // Mock GetCardByArenaID to never resolve
      mockCards.getCardByArenaId.mockImplementation(() => new Promise(() => {}));

      render(<DeckList deck={deck} cards={cards} />);

      expect(screen.getByText('Loading deck...')).toBeInTheDocument();
    });

    it('should not show loading state for empty deck', () => {
      const deck = createMockDeck();

      render(<DeckList deck={deck} cards={[]} />);

      expect(screen.queryByText('Loading deck...')).not.toBeInTheDocument();
    });
  });

  describe('Deck Header', () => {
    it('should display deck name and basic information', async () => {
      const deck = createMockDeck({ Name: 'My Awesome Deck', Format: 'standard', Source: 'manual' });
      const cards = [createMockDeckCard()];
      const mockCard = createMockSetCard();

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('My Awesome Deck')).toBeInTheDocument();
        expect(screen.getByText('standard')).toBeInTheDocument();
        expect(screen.getByText('manual')).toBeInTheDocument();
      });
    });

    it('should show draft indicator for draft decks', async () => {
      const deck = createMockDeck({ Source: 'draft', DraftEventID: 'draft-123' });
      const cards = [createMockDeckCard()];
      const mockCard = createMockSetCard();

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('Draft Deck')).toBeInTheDocument();
      });
    });

    it('should display deck tags if provided', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];
      const tags = [
        new models.DeckTag({ ID: 1, DeckID: 'test-deck-id', Tag: 'aggro' }),
        new models.DeckTag({ ID: 2, DeckID: 'test-deck-id', Tag: 'competitive' }),
      ];
      const mockCard = createMockSetCard();

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} tags={tags} />);

      await waitFor(() => {
        expect(screen.getByText('aggro')).toBeInTheDocument();
        expect(screen.getByText('competitive')).toBeInTheDocument();
      });
    });

    it('should display correct card counts', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 1, Quantity: 4, Board: 'main' }),
        createMockDeckCard({ CardID: 2, Quantity: 3, Board: 'main' }),
        createMockDeckCard({ CardID: 3, Quantity: 2, Board: 'sideboard' }),
      ];

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText(/Mainboard: 7/i)).toBeInTheDocument();
        expect(screen.getByText(/Sideboard: 2/i)).toBeInTheDocument();
      });
    });
  });

  describe('Card Display', () => {
    it('should display cards with correct quantities', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 12345, Quantity: 4, Board: 'main' }),
      ];
      const mockCard = createMockSetCard({ Name: 'Lightning Bolt', ArenaID: '12345' });

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('4x')).toBeInTheDocument();
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
    });

    it('should display mana cost when available', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];
      const mockCard = createMockSetCard({ ManaCost: '{2}{R}' });

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('{2}{R}')).toBeInTheDocument();
      });
    });

    it('should group cards by type correctly', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 1, Board: 'main' }),
        createMockDeckCard({ CardID: 2, Board: 'main' }),
        createMockDeckCard({ CardID: 3, Board: 'main' }),
      ];

      mockCards.getCardByArenaId.mockImplementation((arenaID) => {
        if (arenaID === 1) return Promise.resolve(createMockSetCard({ Name: 'Bear', Types: ['Creature'], ArenaID: '1' }));
        if (arenaID === 2) return Promise.resolve(createMockSetCard({ Name: 'Bolt', Types: ['Instant'], ArenaID: '2' }));
        if (arenaID === 3) return Promise.resolve(createMockSetCard({ Name: 'Island', Types: ['Land'], ArenaID: '3' }));
        return Promise.reject(new Error('Card not found'));
      });

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('Creatures')).toBeInTheDocument();
        expect(screen.getByText('Instants')).toBeInTheDocument();
        expect(screen.getByText('Lands')).toBeInTheDocument();
      });
    });

    it('should handle basic lands by ID even without metadata', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 81716, Quantity: 8, Board: 'main' }), // Plains
      ];

      mockCards.getCardByArenaId.mockRejectedValue(new Error('No metadata'));

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('Plains')).toBeInTheDocument();
        expect(screen.getByText('Lands')).toBeInTheDocument();
      });
    });

    it('should display "Unknown Card" for cards without metadata', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard({ CardID: 99999 })];

      mockCards.getCardByArenaId.mockRejectedValue(new Error('Card not found'));

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText('Unknown Card 99999')).toBeInTheDocument();
      });
    });
  });

  describe('Statistics Charts', () => {
    it('should display mana curve chart when statistics provided', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];
      const stats = createMockStatistics();

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} statistics={stats} />);

      await waitFor(() => {
        expect(screen.getByText('Mana Curve')).toBeInTheDocument();
        expect(screen.getByText('Average CMC: 2.80')).toBeInTheDocument();
      });
    });

    it('should display color distribution chart when statistics provided', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];
      const stats = createMockStatistics();

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} statistics={stats} />);

      await waitFor(() => {
        expect(screen.getByText('Color Distribution')).toBeInTheDocument();
      });
    });

    it('should display land recommendation when statistics provided', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];
      const stats = createMockStatistics({
        lands: {
          total: 17,
          basic: 12,
          nonBasic: 5,
          ratio: 0.425,
          recommended: 17,
          status: 'good',
          statusMessage: 'Land count looks good',
        },
      });

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} statistics={stats} />);

      await waitFor(() => {
        expect(screen.getByText('Land count looks good')).toBeInTheDocument();
      });
    });

    it('should not display charts when statistics not provided', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.queryByText('Mana Curve')).not.toBeInTheDocument();
        expect(screen.queryByText('Color Distribution')).not.toBeInTheDocument();
      });
    });
  });

  describe('Sideboard', () => {
    it('should display sideboard toggle when sideboard cards exist', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 1, Board: 'main' }),
        createMockDeckCard({ CardID: 2, Board: 'sideboard', Quantity: 3 }),
      ];

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.getByText(/Sideboard \(3\)/i)).toBeInTheDocument();
      });
    });

    it('should toggle sideboard visibility when clicked', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 1, Board: 'main' }),
        createMockDeckCard({ CardID: 2, Board: 'sideboard' }),
      ];

      mockCards.getCardByArenaId.mockImplementation((arenaID) => {
        if (arenaID === 2) {
          return Promise.resolve(createMockSetCard({ Name: 'Sideboard Card', ArenaID: '2' }));
        }
        return Promise.resolve(createMockSetCard({ ArenaID: String(arenaID) }));
      });

      render(<DeckList deck={deck} cards={cards} />);

      // Wait for cards to load
      await waitFor(() => {
        expect(screen.getByText(/Sideboard \(1\)/i)).toBeInTheDocument();
      });

      // Sideboard should be hidden initially (cards not visible)
      expect(screen.queryByText('Sideboard Card')).not.toBeInTheDocument();

      // Click to show sideboard
      const sideboardHeader = screen.getByText(/Sideboard \(1\)/i);
      await userEvent.click(sideboardHeader);

      // Sideboard should now be visible
      await waitFor(() => {
        expect(screen.getByText('Sideboard Card')).toBeInTheDocument();
      });

      // Click to hide sideboard again
      await userEvent.click(sideboardHeader);

      // Sideboard should be hidden again
      await waitFor(() => {
        expect(screen.queryByText('Sideboard Card')).not.toBeInTheDocument();
      });
    });

    it('should not display sideboard section when no sideboard cards', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard({ Board: 'main' })];

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.queryByText(/Sideboard/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Card Interactions', () => {
    it('should call onRemoveCard when remove button clicked', async () => {
      const onRemoveCard = vi.fn();
      const deck = createMockDeck();
      const cards = [createMockDeckCard({ CardID: 12345, Board: 'main' })];
      const mockCard = createMockSetCard({ ArenaID: '12345', Name: 'Test Card' });

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} onRemoveCard={onRemoveCard} />);

      // Wait for card to load first
      await waitFor(() => {
        expect(screen.getByText('Test Card')).toBeInTheDocument();
      });

      // Then find and click remove button (button content is "×")
      const removeButton = screen.getByRole('button', { name: '×' });
      await userEvent.click(removeButton);

      expect(onRemoveCard).toHaveBeenCalledWith(12345, 'main');
    });

    it('should not display remove buttons when onRemoveCard not provided', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.queryByRole('button', { name: '×' })).not.toBeInTheDocument();
      });
    });

    it('should call onCardHover when mouse enters card', async () => {
      const onCardHover = vi.fn();
      const deck = createMockDeck();
      const cards = [createMockDeckCard({ CardID: 12345 })];
      const mockCard = createMockSetCard({ Name: 'Test Card', ArenaID: '12345' });

      mockCards.getCardByArenaId.mockResolvedValue(mockCard);

      render(<DeckList deck={deck} cards={cards} onCardHover={onCardHover} />);

      await waitFor(() => {
        expect(screen.getByText('Test Card')).toBeInTheDocument();
      });

      const cardElement = screen.getByText('Test Card').closest('.deck-card');
      if (cardElement) {
        await userEvent.hover(cardElement);
      }

      expect(onCardHover).toHaveBeenCalledWith(mockCard);
    });

    it('should not call onCardHover for cards without metadata', async () => {
      const onCardHover = vi.fn();
      const deck = createMockDeck();
      const cards = [createMockDeckCard({ CardID: 99999 })];

      mockCards.getCardByArenaId.mockRejectedValue(new Error('Not found'));

      render(<DeckList deck={deck} cards={cards} onCardHover={onCardHover} />);

      await waitFor(() => {
        expect(screen.getByText('Unknown Card 99999')).toBeInTheDocument();
      });

      const cardElement = screen.getByText('Unknown Card 99999').closest('.deck-card');
      if (cardElement) {
        await userEvent.hover(cardElement);
      }

      expect(onCardHover).not.toHaveBeenCalled();
    });
  });

  describe('Empty State', () => {
    it('should display empty state message when no cards in deck', async () => {
      const deck = createMockDeck();

      render(<DeckList deck={deck} cards={[]} />);

      await waitFor(() => {
        expect(screen.getByText(/No cards in deck yet/i)).toBeInTheDocument();
      });
    });

    it('should not display empty state when deck has cards', async () => {
      const deck = createMockDeck();
      const cards = [createMockDeckCard()];

      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        expect(screen.queryByText(/No cards in deck yet/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('Card Sorting', () => {
    it('should sort cards by CMC then alphabetically within groups', async () => {
      const deck = createMockDeck();
      const cards = [
        createMockDeckCard({ CardID: 1, Board: 'main' }),
        createMockDeckCard({ CardID: 2, Board: 'main' }),
        createMockDeckCard({ CardID: 3, Board: 'main' }),
      ];

      mockCards.getCardByArenaId.mockImplementation((arenaID) => {
        if (arenaID === 1) return Promise.resolve(createMockSetCard({ Name: 'Zebra', CMC: 3, Types: ['Creature'], ArenaID: '1' }));
        if (arenaID === 2) return Promise.resolve(createMockSetCard({ Name: 'Bear', CMC: 2, Types: ['Creature'], ArenaID: '2' }));
        if (arenaID === 3) return Promise.resolve(createMockSetCard({ Name: 'Aardvark', CMC: 2, Types: ['Creature'], ArenaID: '3' }));
        return Promise.reject(new Error('Card not found'));
      });

      render(<DeckList deck={deck} cards={cards} />);

      await waitFor(() => {
        const creatureSection = screen.getByText('Creatures').closest('.card-group') as HTMLElement | null;
        expect(creatureSection).toBeInTheDocument();

        if (creatureSection) {
          const cardNames = within(creatureSection).getAllByText(/Aardvark|Bear|Zebra/);
          expect(cardNames).toHaveLength(3);
          // Should be sorted by CMC (2, 2, 3) then alphabetically (Aardvark, Bear, Zebra)
          expect(cardNames[0]).toHaveTextContent('Aardvark');
          expect(cardNames[1]).toHaveTextContent('Bear');
          expect(cardNames[2]).toHaveTextContent('Zebra');
        }
      });
    });
  });
});

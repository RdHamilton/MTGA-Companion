import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import CardSearch from './CardSearch';
import { mockCards } from '@/test/mocks/apiMock';
import { models } from '@/types/models';

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
    ImageURL: 'https://example.com/card.jpg',
    ...overrides,
  });
}

describe('CardSearch Component', () => {
  const mockOnAddCard = vi.fn();
  const mockOnRemoveCard = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnAddCard.mockResolvedValue(undefined);
    mockOnRemoveCard.mockResolvedValue(undefined);
  });

  describe('Draft Mode - Initial State', () => {
    it('should display draft mode indicator', async () => {
      const draftCardIDs = [12345, 67890];
      mockCards.getCardByArenaId.mockResolvedValue(createMockSetCard());

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={draftCardIDs}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Draft Mode')).toBeInTheDocument();
        expect(screen.getByText('2 cards in pool')).toBeInTheDocument();
      });
    });

    it('should load cards from draft pool', async () => {
      const card1 = createMockSetCard({ ArenaID: '111', Name: 'Card One' });
      const card2 = createMockSetCard({ ArenaID: '222', Name: 'Card Two' });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 111) return Promise.resolve(card1);
        if (arenaID === 222) return Promise.resolve(card2);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[111, 222]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Card One')).toBeInTheDocument();
        expect(screen.getByText('Card Two')).toBeInTheDocument();
      });
    });

    it('should show loading state while loading cards', () => {
      mockCards.getCardByArenaId.mockImplementation(() => new Promise(() => {}));

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[12345]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      expect(screen.getByText('Searching...')).toBeInTheDocument();
    });

    it('should display empty state when draft pool is empty', async () => {
      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      // Wait for loading to finish and empty state to appear
      await waitFor(() => {
        expect(screen.queryByText('Searching...')).not.toBeInTheDocument();
      });

      expect(screen.getByText('No cards available in draft pool')).toBeInTheDocument();
    });
  });

  describe('Constructed Mode', () => {
    it('should show instruction to type search term for constructed mode', async () => {
      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Type at least 2 characters to search')).toBeInTheDocument();
      });
    });

    it('should not display draft mode indicator', async () => {
      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.queryByText('Draft Mode')).not.toBeInTheDocument();
      });
    });

    it('should search cards when typing at least 2 characters', async () => {
      const card = { ...createMockSetCard({ ArenaID: '123', Name: 'Lightning Bolt' }), ownedQuantity: 2 };
      mockCards.searchCardsWithCollection.mockResolvedValue([card]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'lig');

      await waitFor(() => {
        expect(mockCards.searchCardsWithCollection).toHaveBeenCalledWith('lig', [], 100);
      });

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
    });

    it('should display set filter button for constructed mode', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([
        { code: 'dsk', name: 'Duskmourn', iconSvgUri: '', setType: 'expansion', releasedAt: '', cardCount: 0 },
      ]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Sets:/)).toBeInTheDocument();
      });
    });

    it('should display collection filter toggle for constructed mode', async () => {
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('All Cards')).toBeInTheDocument();
        expect(screen.getByText('My Collection')).toBeInTheDocument();
      });
    });

    it('should filter by collection when toggle is clicked', async () => {
      const ownedCard = { ...createMockSetCard({ ArenaID: '123', Name: 'Owned Card' }), ownedQuantity: 4 };
      mockCards.searchCardsWithCollection.mockResolvedValue([ownedCard]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      // Click on "My Collection" toggle
      const collectionToggle = screen.getByText('My Collection');
      await userEvent.click(collectionToggle);

      // Search for a card
      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'owned');

      await waitFor(() => {
        expect(mockCards.searchCardsWithCollection).toHaveBeenCalledWith('owned', [], 100);
      });
    });

    it('should display owned quantity for cards in constructed mode', async () => {
      const cardWithOwned = { ...createMockSetCard({ ArenaID: '123', Name: 'Test Card' }), ownedQuantity: 3 };
      mockCards.searchCardsWithCollection.mockResolvedValue([cardWithOwned]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'test');

      await waitFor(() => {
        expect(screen.getByText('3x owned')).toBeInTheDocument();
      });
    });

    it('should display "Not owned" for cards not in collection', async () => {
      const cardNotOwned = { ...createMockSetCard({ ArenaID: '123', Name: 'Test Card' }), ownedQuantity: 0 };
      mockCards.searchCardsWithCollection.mockResolvedValue([cardNotOwned]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'test');

      await waitFor(() => {
        expect(screen.getByText('Not owned')).toBeInTheDocument();
      });
    });

    it('should show collection empty message when collection filter active and no results', async () => {
      mockCards.searchCardsWithCollection.mockResolvedValue([]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      // Click on "My Collection" toggle
      const collectionToggle = screen.getByText('My Collection');
      await userEvent.click(collectionToggle);

      // Search for a card
      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'nonexistent');

      await waitFor(() => {
        expect(screen.getByText('No cards in your collection match this search')).toBeInTheDocument();
      });
    });
  });

  describe('Search Filtering', () => {
    it('should filter cards by search term', async () => {
      const card1 = createMockSetCard({ ArenaID: '1', Name: 'Lightning Bolt' });
      const card2 = createMockSetCard({ ArenaID: '2', Name: 'Counterspell' });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(card1);
        if (arenaID === 2) return Promise.resolve(card2);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      // Wait for cards to load
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Counterspell')).toBeInTheDocument();
      });

      // Search for "lightning"
      const searchInput = screen.getByPlaceholderText('Filter draft pool...');
      await userEvent.type(searchInput, 'lightning');

      // Only Lightning Bolt should be visible
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.queryByText('Counterspell')).not.toBeInTheDocument();
      });
    });

    it('should be case-insensitive', async () => {
      const card = createMockSetCard({ ArenaID: '1', Name: 'Lightning Bolt' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      const searchInput = screen.getByPlaceholderText('Filter draft pool...');
      await userEvent.type(searchInput, 'LIGHTNING');

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
    });

    it('should show "no results" message when no cards match', async () => {
      const card = createMockSetCard({ ArenaID: '1', Name: 'Lightning Bolt' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      const searchInput = screen.getByPlaceholderText('Filter draft pool...');
      await userEvent.type(searchInput, 'Nonexistent Card');

      await waitFor(() => {
        expect(screen.getByText('No cards match your search in draft pool')).toBeInTheDocument();
      });
    });
  });

  describe('CMC Filtering', () => {
    it('should filter by minimum CMC', async () => {
      const card1 = createMockSetCard({ ArenaID: '1', Name: 'Cheap Card', CMC: 1 });
      const card2 = createMockSetCard({ ArenaID: '2', Name: 'Expensive Card', CMC: 5 });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(card1);
        if (arenaID === 2) return Promise.resolve(card2);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Cheap Card')).toBeInTheDocument();
        expect(screen.getByText('Expensive Card')).toBeInTheDocument();
      });

      // Set minimum CMC to 3
      const minCmcInput = screen.getByPlaceholderText('Min');
      await userEvent.type(minCmcInput, '3');

      // Only expensive card should be visible
      await waitFor(() => {
        expect(screen.queryByText('Cheap Card')).not.toBeInTheDocument();
        expect(screen.getByText('Expensive Card')).toBeInTheDocument();
      });
    });

    it('should filter by maximum CMC', async () => {
      const card1 = createMockSetCard({ ArenaID: '1', Name: 'Cheap Card', CMC: 1 });
      const card2 = createMockSetCard({ ArenaID: '2', Name: 'Expensive Card', CMC: 5 });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(card1);
        if (arenaID === 2) return Promise.resolve(card2);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Cheap Card')).toBeInTheDocument();
        expect(screen.getByText('Expensive Card')).toBeInTheDocument();
      });

      // Set maximum CMC to 3
      const maxCmcInput = screen.getByPlaceholderText('Max');
      await userEvent.type(maxCmcInput, '3');

      // Only cheap card should be visible
      await waitFor(() => {
        expect(screen.getByText('Cheap Card')).toBeInTheDocument();
        expect(screen.queryByText('Expensive Card')).not.toBeInTheDocument();
      });
    });

    it('should filter by CMC range', async () => {
      const card1 = createMockSetCard({ ArenaID: '1', Name: 'CMC 1', CMC: 1 });
      const card2 = createMockSetCard({ ArenaID: '2', Name: 'CMC 3', CMC: 3 });
      const card3 = createMockSetCard({ ArenaID: '3', Name: 'CMC 7', CMC: 7 });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(card1);
        if (arenaID === 2) return Promise.resolve(card2);
        if (arenaID === 3) return Promise.resolve(card3);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2, 3]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('CMC 1')).toBeInTheDocument();
        expect(screen.getByText('CMC 3')).toBeInTheDocument();
        expect(screen.getByText('CMC 7')).toBeInTheDocument();
      });

      // Set CMC range 2-5
      const minCmcInput = screen.getByPlaceholderText('Min');
      const maxCmcInput = screen.getByPlaceholderText('Max');
      await userEvent.type(minCmcInput, '2');
      await userEvent.type(maxCmcInput, '5');

      // Only CMC 3 should be visible
      await waitFor(() => {
        expect(screen.queryByText('CMC 1')).not.toBeInTheDocument();
        expect(screen.getByText('CMC 3')).toBeInTheDocument();
        expect(screen.queryByText('CMC 7')).not.toBeInTheDocument();
      });
    });
  });

  describe('Color Filtering', () => {
    it('should filter by single color', async () => {
      const blueCard = createMockSetCard({ ArenaID: '1', Name: 'Blue Card', Colors: ['U'] });
      const redCard = createMockSetCard({ ArenaID: '2', Name: 'Red Card', Colors: ['R'] });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(blueCard);
        if (arenaID === 2) return Promise.resolve(redCard);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
        expect(screen.getByText('Red Card')).toBeInTheDocument();
      });

      // Click blue color filter (button has text "U" with title "Blue")
      const blueButton = screen.getByRole('button', { name: 'U' });
      await userEvent.click(blueButton);

      // Only blue card should be visible
      await waitFor(() => {
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
        expect(screen.queryByText('Red Card')).not.toBeInTheDocument();
      });
    });

    it('should filter colorless cards', async () => {
      const colorlessCard = createMockSetCard({ ArenaID: '1', Name: 'Colorless Artifact', Colors: [] });
      const coloredCard = createMockSetCard({ ArenaID: '2', Name: 'Blue Card', Colors: ['U'] });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(colorlessCard);
        if (arenaID === 2) return Promise.resolve(coloredCard);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Colorless Artifact')).toBeInTheDocument();
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
      });

      // Click colorless filter (button has text "C" with title "Colorless")
      const colorlessButton = screen.getByRole('button', { name: 'C' });
      await userEvent.click(colorlessButton);

      // Only colorless card should be visible
      await waitFor(() => {
        expect(screen.getByText('Colorless Artifact')).toBeInTheDocument();
        expect(screen.queryByText('Blue Card')).not.toBeInTheDocument();
      });
    });

    it('should filter multicolor cards', async () => {
      const multicolorCard = createMockSetCard({ ArenaID: '1', Name: 'Multicolor', Colors: ['W', 'U'] });
      const monoCard = createMockSetCard({ ArenaID: '2', Name: 'Mono Blue', Colors: ['U'] });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(multicolorCard);
        if (arenaID === 2) return Promise.resolve(monoCard);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Multicolor')).toBeInTheDocument();
        expect(screen.getByText('Mono Blue')).toBeInTheDocument();
      });

      // Click multicolor filter (button has text "M" with title "Multicolor")
      const multicolorButton = screen.getByRole('button', { name: 'M' });
      await userEvent.click(multicolorButton);

      // Only multicolor card should be visible
      await waitFor(() => {
        expect(screen.getByText('Multicolor')).toBeInTheDocument();
        expect(screen.queryByText('Mono Blue')).not.toBeInTheDocument();
      });
    });
  });

  describe('Type Filtering', () => {
    it('should filter by creature type', async () => {
      const creature = createMockSetCard({ ArenaID: '1', Name: 'Grizzly Bear', Types: ['Creature'] });
      const instant = createMockSetCard({ ArenaID: '2', Name: 'Lightning Bolt', Types: ['Instant'] });

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        if (arenaID === 1) return Promise.resolve(creature);
        if (arenaID === 2) return Promise.resolve(instant);
        return Promise.reject(new Error('Not found'));
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Grizzly Bear')).toBeInTheDocument();
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // Click creature filter - get the type filter button which shows "Creature"
      const creatureButtons = screen.getAllByRole('button', { name: /Creature/i });
      // The type filter button is the one in the type-filters section
      const creatureButton = creatureButtons.find(btn => btn.classList.contains('type-button'));
      await userEvent.click(creatureButton!);

      // Only creature should be visible
      await waitFor(() => {
        expect(screen.getByText('Grizzly Bear')).toBeInTheDocument();
        expect(screen.queryByText('Lightning Bolt')).not.toBeInTheDocument();
      });
    });
  });

  describe('Board Selection', () => {
    it('should default to mainboard', () => {
      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const mainboardButton = screen.getByRole('button', { name: /Maindeck/i });
      expect(mainboardButton).toHaveClass('active');
    });

    it('should switch to sideboard when clicked', async () => {
      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const sideboardButton = screen.getByRole('button', { name: /Sideboard/i });
      await userEvent.click(sideboardButton);

      expect(sideboardButton).toHaveClass('active');
    });
  });

  describe('Card Actions', () => {
    it('should add card to mainboard when add button clicked', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Test Card')).toBeInTheDocument();
      });

      // Button has text "+ Add" with title "Add to main"
      const addButton = screen.getByRole('button', { name: /\+ Add/i });
      await userEvent.click(addButton);

      expect(mockOnAddCard).toHaveBeenCalledWith(123, 1, 'main');
    });

    it('should add card to sideboard when sideboard selected', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('Test Card')).toBeInTheDocument();
      });

      // Switch to sideboard
      const sideboardButton = screen.getByRole('button', { name: /Sideboard/i });
      await userEvent.click(sideboardButton);

      // Click add button (text is "+ Add", title changes to "Add to sideboard")
      const addButton = screen.getByRole('button', { name: /\+ Add/i });
      await userEvent.click(addButton);

      expect(mockOnAddCard).toHaveBeenCalledWith(123, 1, 'sideboard');
    });

    it('should remove card when remove button clicked', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      const existingCards = new Map([[123, { quantity: 2, board: 'main' }]]);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123, 123]}
          existingCards={existingCards}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        // Multiple cards with same name may be rendered, use getAllByText
        expect(screen.getAllByText('Test Card').length).toBeGreaterThan(0);
        // Look for badge text showing quantity in deck (using inline styles now)
        expect(screen.getAllByText(/2x in main/).length).toBeGreaterThan(0);
      });

      // Button has text "- Remove" with title "Remove from deck" - may be multiple, click first
      const removeButtons = screen.getAllByRole('button', { name: /- Remove/i });
      await userEvent.click(removeButtons[0]);

      expect(mockOnRemoveCard).toHaveBeenCalledWith(123, 'main');
    });

    it('should show card is in deck with badge', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      const existingCards = new Map([[123, { quantity: 3, board: 'sideboard' }]]);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123, 123, 123]}
          existingCards={existingCards}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        // Multiple cards with same name may be rendered
        expect(screen.getAllByText('Test Card').length).toBeGreaterThan(0);
        // Look for badge text showing quantity in deck (using inline styles now)
        expect(screen.getAllByText(/3x in sideboard/).length).toBeGreaterThan(0);
      });
    });

    it('should not show add button when all copies are in deck (draft mode)', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      // 2 copies in pool, 2 in deck
      const existingCards = new Map([[123, { quantity: 2, board: 'main' }]]);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123, 123]}
          existingCards={existingCards}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        // Multiple cards with same name may be rendered
        expect(screen.getAllByText('Test Card').length).toBeGreaterThan(0);
      });

      // Add button should not be present (button text is "+ Add")
      expect(screen.queryByRole('button', { name: /\+ Add/i })).not.toBeInTheDocument();
      // But remove button should still be there (button text is "- Remove") - may be multiple
      expect(screen.getAllByRole('button', { name: /- Remove/i }).length).toBeGreaterThan(0);
    });

    it('should display available quantity in draft mode', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      // 3 copies in pool, 1 in deck
      const existingCards = new Map([[123, { quantity: 1, board: 'main' }]]);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123, 123, 123]}
          existingCards={existingCards}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        // Look for available quantity text (using inline styles now)
        expect(screen.getAllByText(/Available: 2 \/ 3/).length).toBeGreaterThan(0);
      });
    });
  });

  describe('Result Display', () => {
    it('should display result count', async () => {
      const cards = [
        createMockSetCard({ ArenaID: '1', Name: 'Card 1' }),
        createMockSetCard({ ArenaID: '2', Name: 'Card 2' }),
        createMockSetCard({ ArenaID: '3', Name: 'Card 3' }),
      ];

      mockCards.getCardByArenaId.mockImplementation((arenaID: number) => {
        const index = (arenaID as number) - 1;
        return Promise.resolve(cards[index]);
      });

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[1, 2, 3]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('3 cards found')).toBeInTheDocument();
      });
    });

    it('should display card image when available', async () => {
      const card = createMockSetCard({
        ArenaID: '123',
        Name: 'Test Card',
        ImageURL: 'https://example.com/card.jpg',
      });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        const image = screen.getByAltText('Test Card');
        expect(image).toBeInTheDocument();
        expect(image).toHaveAttribute('src', 'https://example.com/card.jpg');
      });
    });

    it('should display card type', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card', Types: ['Creature', 'Human'] });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        // Look for the type text (using inline styles now, types joined with em-dash)
        expect(screen.getByText('Creature — Human')).toBeInTheDocument();
      });
    });

    it('should display mana cost', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card', ManaCost: '{3}{U}{U}' });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        expect(screen.getByText('{3}{U}{U}')).toBeInTheDocument();
      });
    });

    it('should display CMC', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card', CMC: 5 });
      mockCards.getCardByArenaId.mockResolvedValue(card);

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[123]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      await waitFor(() => {
        // Look for CMC text (using inline styles now)
        expect(screen.getByText('CMC: 5')).toBeInTheDocument();
      });
    });
  });

  describe('Search Results Display with Inline Styles', () => {
    it('should display search results in a scrollable container', async () => {
      const card = { ...createMockSetCard({ ArenaID: '123', Name: 'Firebending Lesson' }), ownedQuantity: 2 };
      mockCards.searchCardsWithCollection.mockResolvedValue([card]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'firebending');

      await waitFor(() => {
        expect(screen.getByText('Firebending Lesson')).toBeInTheDocument();
      });
    });

    it('should display card image with correct src and alt attributes', async () => {
      const card = {
        ...createMockSetCard({
          ArenaID: '123',
          Name: 'Firebending Lesson',
          ImageURL: 'https://cards.scryfall.io/firebending.jpg',
        }),
        ownedQuantity: 2,
      };
      mockCards.searchCardsWithCollection.mockResolvedValue([card]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'firebending');

      await waitFor(() => {
        const image = screen.getByAltText('Firebending Lesson');
        expect(image).toBeInTheDocument();
        expect(image).toHaveAttribute('src', 'https://cards.scryfall.io/firebending.jpg');
      });
    });

    it('should display multiple search results with correct count', async () => {
      const cards = [
        { ...createMockSetCard({ ArenaID: '1', Name: 'Fire Bolt' }), ownedQuantity: 4 },
        { ...createMockSetCard({ ArenaID: '2', Name: 'Fire Elemental' }), ownedQuantity: 2 },
        { ...createMockSetCard({ ArenaID: '3', Name: 'Fireball' }), ownedQuantity: 0 },
      ];
      mockCards.searchCardsWithCollection.mockResolvedValue(cards);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'fire');

      await waitFor(() => {
        expect(screen.getByText('3 cards found')).toBeInTheDocument();
        expect(screen.getByText('Fire Bolt')).toBeInTheDocument();
        expect(screen.getByText('Fire Elemental')).toBeInTheDocument();
        expect(screen.getByText('Fireball')).toBeInTheDocument();
      });
    });

    it('should show owned quantity with correct styling', async () => {
      const ownedCard = { ...createMockSetCard({ ArenaID: '1', Name: 'Owned Card' }), ownedQuantity: 4 };
      const notOwnedCard = { ...createMockSetCard({ ArenaID: '2', Name: 'Unowned Card' }), ownedQuantity: 0 };
      mockCards.searchCardsWithCollection.mockResolvedValue([ownedCard, notOwnedCard]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'card');

      await waitFor(() => {
        expect(screen.getByText('4x owned')).toBeInTheDocument();
        expect(screen.getByText('Not owned')).toBeInTheDocument();
      });
    });

    it('should display add and remove buttons for cards in deck', async () => {
      const card = { ...createMockSetCard({ ArenaID: '123', Name: 'Test Card' }), ownedQuantity: 4 };
      mockCards.searchCardsWithCollection.mockResolvedValue([card]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      const existingCards = new Map([[123, { quantity: 2, board: 'main' }]]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={existingCards}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'test');

      await waitFor(() => {
        expect(screen.getByText('Test Card')).toBeInTheDocument();
        expect(screen.getByText(/2x in main/)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /\+ Add/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /- Remove/i })).toBeInTheDocument();
      });
    });

    it('should call onAddCard when add button is clicked in constructed mode', async () => {
      const card = { ...createMockSetCard({ ArenaID: '456', Name: 'Searchable Card' }), ownedQuantity: 4 };
      mockCards.searchCardsWithCollection.mockResolvedValue([card]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');
      await userEvent.type(searchInput, 'searchable');

      await waitFor(() => {
        expect(screen.getByText('Searchable Card')).toBeInTheDocument();
      });

      const addButton = screen.getByRole('button', { name: /\+ Add/i });
      await userEvent.click(addButton);

      expect(mockOnAddCard).toHaveBeenCalledWith(456, 1, 'main');
    });

    it('should debounce search input to prevent excessive API calls', async () => {
      mockCards.searchCardsWithCollection.mockResolvedValue([]);
      mockCards.getAllSetInfo.mockResolvedValue([]);

      render(
        <CardSearch
          isDraftDeck={false}
          draftCardIDs={[]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search cards (min 2 characters)...');

      // Type quickly - should debounce
      await userEvent.type(searchInput, 'fire');

      // Wait for debounce (300ms)
      await waitFor(() => {
        // Should only call once after debounce, not for each character
        expect(mockCards.searchCardsWithCollection).toHaveBeenCalledTimes(1);
        expect(mockCards.searchCardsWithCollection).toHaveBeenCalledWith('fire', [], 100);
      });
    });
  });
});

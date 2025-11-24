import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import CardSearch from './CardSearch';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { models } from '../../wailsjs/go/models';

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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(createMockSetCard());

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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '111') return Promise.resolve(card1);
        if (arenaID === '222') return Promise.resolve(card2);
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
      mockWailsApp.GetCardByArenaID.mockImplementation(() => new Promise(() => {}));

      render(
        <CardSearch
          isDraftDeck={true}
          draftCardIDs={[12345]}
          existingCards={new Map()}
          onAddCard={mockOnAddCard}
          onRemoveCard={mockOnRemoveCard}
        />
      );

      expect(screen.getByText('Loading cards...')).toBeInTheDocument();
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
        expect(screen.queryByText('Loading cards...')).not.toBeInTheDocument();
      });

      expect(screen.getByText('No cards available in draft pool')).toBeInTheDocument();
    });
  });

  describe('Constructed Mode', () => {
    it('should show error message for constructed mode', async () => {
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
        expect(screen.getByText('Please select a set to search for cards')).toBeInTheDocument();
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
  });

  describe('Search Filtering', () => {
    it('should filter cards by search term', async () => {
      const card1 = createMockSetCard({ ArenaID: '1', Name: 'Lightning Bolt' });
      const card2 = createMockSetCard({ ArenaID: '2', Name: 'Counterspell' });

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(card1);
        if (arenaID === '2') return Promise.resolve(card2);
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
      const searchInput = screen.getByPlaceholderText('Search card name...');
      await userEvent.type(searchInput, 'lightning');

      // Only Lightning Bolt should be visible
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.queryByText('Counterspell')).not.toBeInTheDocument();
      });
    });

    it('should be case-insensitive', async () => {
      const card = createMockSetCard({ ArenaID: '1', Name: 'Lightning Bolt' });
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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

      const searchInput = screen.getByPlaceholderText('Search card name...');
      await userEvent.type(searchInput, 'LIGHTNING');

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
    });

    it('should show "no results" message when no cards match', async () => {
      const card = createMockSetCard({ ArenaID: '1', Name: 'Lightning Bolt' });
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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

      const searchInput = screen.getByPlaceholderText('Search card name...');
      await userEvent.type(searchInput, 'Nonexistent Card');

      await waitFor(() => {
        expect(screen.getByText('No cards match your search criteria')).toBeInTheDocument();
      });
    });
  });

  describe('CMC Filtering', () => {
    it('should filter by minimum CMC', async () => {
      const card1 = createMockSetCard({ ArenaID: '1', Name: 'Cheap Card', CMC: 1 });
      const card2 = createMockSetCard({ ArenaID: '2', Name: 'Expensive Card', CMC: 5 });

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(card1);
        if (arenaID === '2') return Promise.resolve(card2);
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(card1);
        if (arenaID === '2') return Promise.resolve(card2);
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(card1);
        if (arenaID === '2') return Promise.resolve(card2);
        if (arenaID === '3') return Promise.resolve(card3);
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(blueCard);
        if (arenaID === '2') return Promise.resolve(redCard);
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(colorlessCard);
        if (arenaID === '2') return Promise.resolve(coloredCard);
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(multicolorCard);
        if (arenaID === '2') return Promise.resolve(monoCard);
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        if (arenaID === '1') return Promise.resolve(creature);
        if (arenaID === '2') return Promise.resolve(instant);
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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
        // Multiple badges may exist, check any of them
        const badges = document.querySelectorAll('.in-deck-badge');
        expect(badges.length).toBeGreaterThan(0);
        expect(badges[0]?.textContent).toBe('2x in main');
      });

      // Button has text "- Remove" with title "Remove from deck" - may be multiple, click first
      const removeButtons = screen.getAllByRole('button', { name: /- Remove/i });
      await userEvent.click(removeButtons[0]);

      expect(mockOnRemoveCard).toHaveBeenCalledWith(123, 'main');
    });

    it('should show card is in deck with badge', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
        // Text is broken into multiple nodes, so use a function matcher
        // There may be multiple badges, we just need at least one
        const badges = document.querySelectorAll('.in-deck-badge');
        expect(badges.length).toBeGreaterThan(0);
        expect(badges[0]?.textContent).toBe('3x in sideboard');
      });
    });

    it('should not show add button when all copies are in deck (draft mode)', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card' });
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
        // Multiple cards may be rendered, check any available-quantity element
        const quantities = document.querySelectorAll('.available-quantity');
        expect(quantities.length).toBeGreaterThan(0);
        expect(quantities[0]?.textContent).toBe('Available: 2 / 3');
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

      mockWailsApp.GetCardByArenaID.mockImplementation((arenaID) => {
        const index = parseInt(arenaID as string) - 1;
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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
        // The type display is in a .card-type div
        const typeElements = screen.getAllByText(/Creature/);
        const cardTypeElement = typeElements.find(el => el.classList.contains('card-type'));
        expect(cardTypeElement).toBeInTheDocument();
        expect(cardTypeElement?.textContent).toBe('Creature — Human');
      });
    });

    it('should display mana cost', async () => {
      const card = createMockSetCard({ ArenaID: '123', Name: 'Test Card', ManaCost: '{3}{U}{U}' });
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);

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
        // CMC text may be broken into nodes, so check the element content
        const cardStats = document.querySelector('.card-stats');
        expect(cardStats).toBeInTheDocument();
        expect(cardStats?.textContent).toContain('CMC: 5');
      });
    });
  });
});

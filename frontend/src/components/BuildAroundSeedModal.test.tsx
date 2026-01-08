import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import BuildAroundSeedModal from './BuildAroundSeedModal';

// Mock the API services
vi.mock('@/services/api', () => ({
  decks: {
    buildAroundSeed: vi.fn(),
    suggestNextCards: vi.fn(),
    generateCompleteDeck: vi.fn(),
    getArchetypeProfiles: vi.fn(),
  },
  cards: {
    searchCardsWithCollection: vi.fn(),
    getCardByArenaId: vi.fn(),
  },
}));

import { decks, cards } from '@/services/api';

const mockBuildAroundSeed = vi.mocked(decks.buildAroundSeed);
const mockSuggestNextCards = vi.mocked(decks.suggestNextCards);
const mockGenerateCompleteDeck = vi.mocked(decks.generateCompleteDeck);
const mockGetArchetypeProfiles = vi.mocked(decks.getArchetypeProfiles);
const mockSearchCardsWithCollection = vi.mocked(cards.searchCardsWithCollection);
const mockGetCardByArenaId = vi.mocked(cards.getCardByArenaId);

beforeEach(() => {
  vi.clearAllMocks();
});

describe('BuildAroundSeedModal', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    onApplyDeck: vi.fn(),
  };

  it('should not render when isOpen is false', () => {
    render(
      <BuildAroundSeedModal
        {...defaultProps}
        isOpen={false}
      />
    );

    expect(screen.queryByText('Build Around Card')).not.toBeInTheDocument();
  });

  it('should render modal content when open', () => {
    render(<BuildAroundSeedModal {...defaultProps} />);

    expect(screen.getByText('Build Around Card')).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/search for a card/i)).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn();

    render(
      <BuildAroundSeedModal
        {...defaultProps}
        onClose={onClose}
      />
    );

    const closeButton = screen.getByRole('button', { name: /close dialog/i });
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalled();
  });

  it('should call onClose when clicking overlay', () => {
    const onClose = vi.fn();

    render(
      <BuildAroundSeedModal
        {...defaultProps}
        onClose={onClose}
      />
    );

    const overlay = document.querySelector('.build-around-overlay');
    if (overlay) {
      fireEvent.click(overlay);
    }

    expect(onClose).toHaveBeenCalled();
  });

  it('should search for cards when typing in search input', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '12345',
        Name: 'Test Card',
        ManaCost: '{2}{W}',
        Types: ['Creature', 'Human'],
        Colors: ['W'],
        ImageURL: '',
      },
    ] as any);

    render(<BuildAroundSeedModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText(/search for a card/i);
    fireEvent.change(searchInput, { target: { value: 'Test' } });

    await waitFor(() => {
      expect(mockSearchCardsWithCollection).toHaveBeenCalledWith('Test', undefined, 50);
    });
  });

  it('should display search results', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '12345',
        Name: 'Sheoldred',
        ManaCost: '{2}{B}{B}',
        Types: ['Legendary', 'Creature', 'Phyrexian', 'Praetor'],
        Colors: ['B'],
        ImageURL: '',
      },
    ] as any);

    render(<BuildAroundSeedModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText(/search for a card/i);
    fireEvent.change(searchInput, { target: { value: 'Sheol' } });

    await waitFor(() => {
      expect(screen.getByText('Sheoldred')).toBeInTheDocument();
    });
  });

  it('should select a card and show build options', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '12345',
        Name: 'Test Build Card',
        ManaCost: '{2}{U}',
        Types: ['Creature', 'Wizard'],
        Colors: ['U'],
        ImageURL: '',
      },
    ] as any);

    render(<BuildAroundSeedModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText(/search for a card/i);
    fireEvent.change(searchInput, { target: { value: 'Test Build' } });

    await waitFor(() => {
      expect(screen.getByText('Test Build Card')).toBeInTheDocument();
    });

    // Click on search result
    fireEvent.click(screen.getByText('Test Build Card'));

    // Should show build button (Quick Build for one-shot mode)
    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });
  });

  it('should toggle budget mode checkbox', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '12345',
        Name: 'Test Card',
        ManaCost: '{1}',
        Types: ['Artifact'],
        Colors: [],
        ImageURL: '',
      },
    ] as any);

    render(<BuildAroundSeedModal {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText(/search for a card/i);
    fireEvent.change(searchInput, { target: { value: 'Test' } });

    await waitFor(() => {
      expect(screen.getByText('Test Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Test Card'));

    await waitFor(() => {
      expect(screen.getByText(/budget mode/i)).toBeInTheDocument();
    });

    const checkbox = screen.getByRole('checkbox');
    expect(checkbox).not.toBeChecked();
    fireEvent.click(checkbox);
    expect(checkbox).toBeChecked();
  });

  it('should build suggestions when Quick Build button is clicked', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '12345',
        Name: 'Build Around Me',
        ManaCost: '{2}{G}',
        Types: ['Creature', 'Elf'],
        Colors: ['G'],
        ImageURL: '',
      },
    ] as any);

    mockBuildAroundSeed.mockResolvedValue({
      seedCard: {
        cardID: 12345,
        name: 'Build Around Me',
        manaCost: '{2}{G}',
        cmc: 3,
        colors: ['G'],
        typeLine: 'Creature - Elf',
        score: 1.0,
        reasoning: 'This is your build-around card.',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      },
      suggestions: [
        {
          cardID: 11111,
          name: 'Suggested Creature',
          manaCost: '{1}{G}',
          cmc: 2,
          colors: ['G'],
          typeLine: 'Creature - Beast',
          score: 0.85,
          reasoning: 'Great synergy',
          inCollection: true,
          ownedCount: 2,
          neededCount: 2,
          currentCopies: 0,
          recommendedCopies: 4,
        },
      ],
      lands: [
        { cardID: 81720, name: 'Forest', quantity: 24, color: 'G' },
      ],
      analysis: {
        colorIdentity: ['G'],
        keywords: ['trample'],
        themes: ['ramp'],
        idealCurve: { 1: 4, 2: 8, 3: 8, 4: 6 },
        suggestedLandCount: 24,
        totalCards: 60,
        inCollectionCount: 30,
        missingCount: 6,
        missingWildcardCost: { rare: 4, uncommon: 2 },
      },
    });

    render(<BuildAroundSeedModal {...defaultProps} />);

    // Search and select card
    const searchInput = screen.getByPlaceholderText(/search for a card/i);
    fireEvent.change(searchInput, { target: { value: 'Build' } });

    await waitFor(() => {
      expect(screen.getByText('Build Around Me')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Build Around Me'));

    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });

    // Click build button
    fireEvent.click(screen.getByText('Quick Build (40 Cards)'));

    await waitFor(() => {
      expect(mockBuildAroundSeed).toHaveBeenCalledWith({
        seed_card_id: 12345,
        max_results: 40,
        budget_mode: false,
        set_restriction: 'all',
      });
    });

    // Should show analysis results
    await waitFor(() => {
      expect(screen.getByText('Deck Analysis')).toBeInTheDocument();
    });
  });

  it('should show suggestions after Quick Build', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '99999',
        Name: 'Seed Card',
        ManaCost: '{W}',
        Types: ['Creature'],
        Colors: ['W'],
        ImageURL: '',
      },
    ] as any);

    mockBuildAroundSeed.mockResolvedValue({
      seedCard: {
        cardID: 99999,
        name: 'Seed Card',
        manaCost: '{W}',
        cmc: 1,
        colors: ['W'],
        typeLine: 'Creature',
        score: 1.0,
        reasoning: 'Build around',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      },
      suggestions: [
        {
          cardID: 88888,
          name: 'White Creature',
          manaCost: '{1}{W}',
          cmc: 2,
          colors: ['W'],
          typeLine: 'Creature - Soldier',
          score: 0.9,
          reasoning: 'Good card',
          inCollection: false,
          ownedCount: 0,
          neededCount: 4,
          currentCopies: 0,
          recommendedCopies: 4,
        },
        {
          cardID: 77777,
          name: 'White Spell',
          manaCost: '{2}{W}',
          cmc: 3,
          colors: ['W'],
          typeLine: 'Instant',
          score: 0.8,
          reasoning: 'Also good',
          inCollection: true,
          ownedCount: 2,
          neededCount: 2,
          currentCopies: 0,
          recommendedCopies: 4,
        },
      ],
      lands: [
        { cardID: 81716, name: 'Plains', quantity: 24, color: 'W' },
      ],
      analysis: {
        colorIdentity: ['W'],
        keywords: ['lifelink'],
        themes: [],
        idealCurve: { 1: 4, 2: 8, 3: 8 },
        suggestedLandCount: 24,
        totalCards: 60,
        inCollectionCount: 20,
        missingCount: 16,
        missingWildcardCost: { rare: 8, uncommon: 4, common: 4 },
      },
    });

    render(<BuildAroundSeedModal {...defaultProps} />);

    // Search, select, and build
    const searchInput = screen.getByPlaceholderText(/search for a card/i);
    fireEvent.change(searchInput, { target: { value: 'Seed' } });

    await waitFor(() => {
      expect(screen.getByText('Seed Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Seed Card'));

    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Quick Build (40 Cards)'));

    // Wait for suggestions to load - check that card names appear
    await waitFor(() => {
      expect(screen.getByText('White Creature')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByText('White Spell')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByText('Plains')).toBeInTheDocument();
    });

    // Check category headers are shown (use queryAllByRole to check h4 headings exist)
    const headings = screen.getAllByRole('heading', { level: 4 });
    const headingTexts = headings.map(h => h.textContent);
    expect(headingTexts.some(text => text?.includes('Creatures'))).toBe(true);
    expect(headingTexts.some(text => text?.includes('Spells'))).toBe(true);
    expect(headingTexts.some(text => text?.includes('Lands'))).toBe(true);
  });

  it('should show ownership badges correctly', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '55555',
        Name: 'Ownership Test',
        ManaCost: '{B}',
        Types: ['Creature'],
        Colors: ['B'],
        ImageURL: '',
      },
    ] as any);

    mockBuildAroundSeed.mockResolvedValue({
      seedCard: {
        cardID: 55555,
        name: 'Ownership Test',
        manaCost: '{B}',
        cmc: 1,
        colors: ['B'],
        typeLine: 'Creature',
        score: 1.0,
        reasoning: 'Test',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      },
      suggestions: [
        {
          cardID: 44444,
          name: 'Owned Card',
          manaCost: '{1}{B}',
          cmc: 2,
          colors: ['B'],
          typeLine: 'Creature - Zombie',
          score: 0.9,
          reasoning: 'Owned',
          inCollection: true,
          ownedCount: 3,
          neededCount: 1,
          currentCopies: 0,
          recommendedCopies: 4,
        },
        {
          cardID: 33333,
          name: 'Missing Card',
          manaCost: '{2}{B}',
          cmc: 3,
          colors: ['B'],
          typeLine: 'Creature - Demon',
          score: 0.8,
          reasoning: 'Missing',
          inCollection: false,
          ownedCount: 0,
          neededCount: 4,
          currentCopies: 0,
          recommendedCopies: 4,
        },
      ],
      lands: [],
      analysis: {
        colorIdentity: ['B'],
        keywords: [],
        themes: [],
        idealCurve: {},
        suggestedLandCount: 0,
        totalCards: 3,
        inCollectionCount: 2,
        missingCount: 1,
        missingWildcardCost: {},
      },
    });

    render(<BuildAroundSeedModal {...defaultProps} />);

    // Search, select, build
    fireEvent.change(screen.getByPlaceholderText(/search for a card/i), { target: { value: 'Ownership' } });

    await waitFor(() => {
      expect(screen.getByText('Ownership Test')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Ownership Test'));

    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Quick Build (40 Cards)'));

    await waitFor(() => {
      expect(screen.getByText('Own 3')).toBeInTheDocument();
      expect(screen.getByText('Need 4')).toBeInTheDocument();
    });
  });

  it('should call onApplyDeck when apply button is clicked', async () => {
    const onApplyDeck = vi.fn();

    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '22222',
        Name: 'Apply Test',
        ManaCost: '{R}',
        Types: ['Creature'],
        Colors: ['R'],
        ImageURL: '',
      },
    ] as any);

    const mockSuggestions = [
      {
        cardID: 11111,
        name: 'Red Card',
        manaCost: '{1}{R}',
        cmc: 2,
        colors: ['R'],
        typeLine: 'Creature',
        score: 0.9,
        reasoning: 'Good',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      },
    ];

    const mockLands = [
      { cardID: 81719, name: 'Mountain', quantity: 24, color: 'R' },
    ];

    mockBuildAroundSeed.mockResolvedValue({
      seedCard: {
        cardID: 22222,
        name: 'Apply Test',
        manaCost: '{R}',
        cmc: 1,
        colors: ['R'],
        typeLine: 'Creature',
        score: 1.0,
        reasoning: 'Test',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      },
      suggestions: mockSuggestions,
      lands: mockLands,
      analysis: {
        colorIdentity: ['R'],
        keywords: [],
        themes: [],
        idealCurve: {},
        suggestedLandCount: 24,
        totalCards: 26,
        inCollectionCount: 26,
        missingCount: 0,
        missingWildcardCost: {},
      },
    });

    render(
      <BuildAroundSeedModal
        {...defaultProps}
        onApplyDeck={onApplyDeck}
      />
    );

    // Search, select, build
    fireEvent.change(screen.getByPlaceholderText(/search for a card/i), { target: { value: 'Apply' } });

    await waitFor(() => {
      expect(screen.getByText('Apply Test')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Apply Test'));

    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Quick Build (40 Cards)'));

    await waitFor(() => {
      expect(screen.getByText('Apply to Current Deck')).toBeInTheDocument();
    });

    // Click apply
    fireEvent.click(screen.getByText('Apply to Current Deck'));

    await waitFor(() => {
      expect(onApplyDeck).toHaveBeenCalledWith(mockSuggestions, mockLands);
    });
  });

  it('should show error message on API failure', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '66666',
        Name: 'Error Test',
        ManaCost: '{G}',
        Types: ['Creature'],
        Colors: ['G'],
        ImageURL: '',
      },
    ] as any);

    mockBuildAroundSeed.mockRejectedValue(new Error('API Error'));

    render(<BuildAroundSeedModal {...defaultProps} />);

    fireEvent.change(screen.getByPlaceholderText(/search for a card/i), { target: { value: 'Error' } });

    await waitFor(() => {
      expect(screen.getByText('Error Test')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Error Test'));

    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Quick Build (40 Cards)'));

    await waitFor(() => {
      expect(screen.getByText('API Error')).toBeInTheDocument();
    });
  });

  it('should clear selection when clear button is clicked', async () => {
    mockSearchCardsWithCollection.mockResolvedValue([
      {
        ArenaID: '77777',
        Name: 'Clear Test Card',
        ManaCost: '{U}',
        Types: ['Creature'],
        Colors: ['U'],
        ImageURL: '',
      },
    ] as any);

    render(<BuildAroundSeedModal {...defaultProps} />);

    fireEvent.change(screen.getByPlaceholderText(/search for a card/i), { target: { value: 'Clear' } });

    await waitFor(() => {
      expect(screen.getByText('Clear Test Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Clear Test Card'));

    await waitFor(() => {
      expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
    });

    // Click clear button
    fireEvent.click(screen.getByRole('button', { name: /clear/i }));

    // Build button should be gone
    expect(screen.queryByText('Quick Build (40 Cards)')).not.toBeInTheDocument();
  });

  describe('Search Filters', () => {
    it('should filter search results by color when color button is clicked', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        { ArenaID: '1', Name: 'White Card', Types: ['Creature'], Colors: ['W'], ImageURL: '' },
        { ArenaID: '2', Name: 'Blue Card', Types: ['Creature'], Colors: ['U'], ImageURL: '' },
        { ArenaID: '3', Name: 'Red Card', Types: ['Creature'], Colors: ['R'], ImageURL: '' },
      ] as any);

      render(<BuildAroundSeedModal {...defaultProps} />);

      // Search for cards
      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Card' } });

      // Wait for all results to appear
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
        expect(screen.getByText('Red Card')).toBeInTheDocument();
      });

      // Click the White color filter button
      const whiteButton = screen.getByTitle('White');
      fireEvent.click(whiteButton);

      // Only White Card should be visible now
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
        expect(screen.queryByText('Blue Card')).not.toBeInTheDocument();
        expect(screen.queryByText('Red Card')).not.toBeInTheDocument();
      });
    });

    it('should filter search results by type when type button is clicked', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        { ArenaID: '1', Name: 'Test Creature', Types: ['Creature'], Colors: ['W'], ImageURL: '' },
        { ArenaID: '2', Name: 'Test Instant', Types: ['Instant'], Colors: ['U'], ImageURL: '' },
        { ArenaID: '3', Name: 'Test Sorcery', Types: ['Sorcery'], Colors: ['R'], ImageURL: '' },
      ] as any);

      render(<BuildAroundSeedModal {...defaultProps} />);

      // Search for cards
      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Test' } });

      // Wait for all results to appear
      await waitFor(() => {
        expect(screen.getByText('Test Creature')).toBeInTheDocument();
        expect(screen.getByText('Test Instant')).toBeInTheDocument();
        expect(screen.getByText('Test Sorcery')).toBeInTheDocument();
      });

      // Click the Creature type filter button (shows as "Crea")
      const creatureButton = screen.getByText('Crea');
      fireEvent.click(creatureButton);

      // Only Creature should be visible now
      await waitFor(() => {
        expect(screen.getByText('Test Creature')).toBeInTheDocument();
        expect(screen.queryByText('Test Instant')).not.toBeInTheDocument();
        expect(screen.queryByText('Test Sorcery')).not.toBeInTheDocument();
      });
    });

    it('should show no results when filter excludes all cards', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        { ArenaID: '1', Name: 'White Card', Types: ['Creature'], Colors: ['W'], ImageURL: '' },
        { ArenaID: '2', Name: 'Blue Card', Types: ['Creature'], Colors: ['U'], ImageURL: '' },
      ] as any);

      render(<BuildAroundSeedModal {...defaultProps} />);

      // Search for cards
      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Card' } });

      // Wait for results
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
      });

      // Click Red color filter (no red cards in results)
      const redButton = screen.getByTitle('Red');
      fireEvent.click(redButton);

      // No cards should be visible
      await waitFor(() => {
        expect(screen.queryByText('White Card')).not.toBeInTheDocument();
        expect(screen.queryByText('Blue Card')).not.toBeInTheDocument();
      });
    });

    it('should allow multiple color filters (OR logic)', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        { ArenaID: '1', Name: 'White Card', Types: ['Creature'], Colors: ['W'], ImageURL: '' },
        { ArenaID: '2', Name: 'Blue Card', Types: ['Creature'], Colors: ['U'], ImageURL: '' },
        { ArenaID: '3', Name: 'Red Card', Types: ['Creature'], Colors: ['R'], ImageURL: '' },
      ] as any);

      render(<BuildAroundSeedModal {...defaultProps} />);

      // Search for cards
      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Card' } });

      // Wait for results
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
      });

      // Click White and Blue color filters
      fireEvent.click(screen.getByTitle('White'));
      fireEvent.click(screen.getByTitle('Blue'));

      // White and Blue cards should be visible, Red should not
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
        expect(screen.queryByText('Red Card')).not.toBeInTheDocument();
      });
    });

    it('should toggle filter off when clicked again', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        { ArenaID: '1', Name: 'White Card', Types: ['Creature'], Colors: ['W'], ImageURL: '' },
        { ArenaID: '2', Name: 'Blue Card', Types: ['Creature'], Colors: ['U'], ImageURL: '' },
      ] as any);

      render(<BuildAroundSeedModal {...defaultProps} />);

      // Search for cards
      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Card' } });

      // Wait for results
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
      });

      // Click White to filter
      fireEvent.click(screen.getByTitle('White'));

      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
        expect(screen.queryByText('Blue Card')).not.toBeInTheDocument();
      });

      // Click White again to remove filter
      fireEvent.click(screen.getByTitle('White'));

      // Both cards should be visible again
      await waitFor(() => {
        expect(screen.getByText('White Card')).toBeInTheDocument();
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
      });
    });
  });

  describe('Iterative Mode', () => {
    const iterativeProps = {
      isOpen: true,
      onClose: vi.fn(),
      onApplyDeck: vi.fn(),
      onCardAdded: vi.fn(),
      onFinishDeck: vi.fn(),
      currentDeckCards: [],
    };

    it('should show Start Building button when iterative callbacks provided', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Iterative Test',
          ManaCost: '{W}',
          Types: ['Creature'],
          Colors: ['W'],
          ImageURL: '',
        },
      ] as any);

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Iterative' } });

      await waitFor(() => {
        expect(screen.getByText('Iterative Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Iterative Test'));

      // Should show both buttons when iterative mode is available
      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
        expect(screen.getByText('Quick Build (40 Cards)')).toBeInTheDocument();
      });
    });

    it('should enter iterative mode when Start Building is clicked', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '54321',
          Name: 'Start Build Card',
          ManaCost: '{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [
          {
            cardID: 11111,
            name: 'Suggestion 1',
            manaCost: '{1}{U}',
            cmc: 2,
            colors: ['U'],
            typeLine: 'Creature',
            score: 0.9,
            reasoning: 'Good synergy',
            inCollection: true,
            ownedCount: 4,
            neededCount: 0,
            currentCopies: 0,
            recommendedCopies: 4,
          },
        ],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: ['Flying'],
          themes: ['Control'],
          currentCurve: { 2: 1 },
          recommendedLandCount: 24,
          totalCards: 1,
          inCollectionCount: 1,
        },
        slotsRemaining: 59,
        landSuggestions: [
          { cardID: 81717, name: 'Island', quantity: 24, color: 'U' },
        ],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Start' } });

      await waitFor(() => {
        expect(screen.getByText('Start Build Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Build Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      // Should show iterative mode UI
      await waitFor(() => {
        expect(screen.getByText(/Building:/)).toBeInTheDocument();
        expect(screen.getByText('59 slots remaining')).toBeInTheDocument();
      });

      // Should call the API
      expect(mockSuggestNextCards).toHaveBeenCalledWith({
        seed_card_id: 54321,
        deck_card_ids: [],
        max_results: 15,
        budget_mode: false,
      });
    });

    it('should open copy selection modal when clicking a suggestion in iterative mode', async () => {
      const onCardAdded = vi.fn();

      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '99999',
          Name: 'Pick Cards Test',
          ManaCost: '{G}',
          Types: ['Creature'],
          Colors: ['G'],
          ImageURL: '',
        },
      ] as any);

      const mockCard = {
        cardID: 22222,
        name: 'Pickable Suggestion',
        manaCost: '{1}{G}',
        cmc: 2,
        colors: ['G'],
        typeLine: 'Creature',
        score: 0.85,
        reasoning: 'Synergy',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      };

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [mockCard],
        deckAnalysis: {
          colorIdentity: ['G'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(
        <BuildAroundSeedModal
          {...iterativeProps}
          onCardAdded={onCardAdded}
        />
      );

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Pick' } });

      await waitFor(() => {
        expect(screen.getByText('Pick Cards Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Pick Cards Test'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      // Wait for iterative mode to load - look for the clickable grid container
      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Find the clickable suggestion card and click it to open copy modal
      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      expect(suggestionCards.length).toBeGreaterThan(0);
      fireEvent.click(suggestionCards[0]);

      // Copy selection modal should open
      await waitFor(() => {
        expect(screen.getByText(/Add Pickable Suggestion/)).toBeInTheDocument();
      });

      // Click the +1 button to add one copy
      const addOneButton = screen.getByRole('button', { name: /\+1/i });
      fireEvent.click(addOneButton);

      // onCardAdded should be called once
      expect(onCardAdded).toHaveBeenCalledTimes(1);
      expect(onCardAdded).toHaveBeenCalledWith(mockCard);
    });

    it('should add multiple copies when selecting higher count in copy modal', async () => {
      const onCardAdded = vi.fn();

      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '88888',
          Name: 'Multi Copy Test',
          ManaCost: '{R}',
          Types: ['Instant'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      const mockCard = {
        cardID: 33333,
        name: 'Multi Copy Card',
        manaCost: '{R}',
        cmc: 1,
        colors: ['R'],
        typeLine: 'Instant',
        score: 0.9,
        reasoning: 'High synergy',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      };

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [mockCard],
        deckAnalysis: {
          colorIdentity: ['R'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(
        <BuildAroundSeedModal
          {...iterativeProps}
          onCardAdded={onCardAdded}
        />
      );

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Multi' } });

      await waitFor(() => {
        expect(screen.getByText('Multi Copy Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Multi Copy Test'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      fireEvent.click(suggestionCards[0]);

      // Wait for copy modal
      await waitFor(() => {
        expect(screen.getByText(/Add Multi Copy Card/)).toBeInTheDocument();
      });

      // Click the +4 button
      const addFourButton = screen.getByRole('button', { name: /\+4/i });
      fireEvent.click(addFourButton);

      // onCardAdded should be called 4 times
      expect(onCardAdded).toHaveBeenCalledTimes(4);
    });

    it('should call onFinishDeck with land suggestions when Finish Deck is clicked', async () => {
      const onFinishDeck = vi.fn();

      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '11111',
          Name: 'Finish Deck Test',
          ManaCost: '{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      const mockLands = [
        { cardID: 81719, name: 'Mountain', quantity: 22, color: 'R' },
      ];

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [],
        deckAnalysis: {
          colorIdentity: ['R'],
          keywords: [],
          themes: [],
          currentCurve: { 1: 8, 2: 10, 3: 8, 4: 4, 5: 2 },
          recommendedLandCount: 22,
          totalCards: 32,
          inCollectionCount: 32,
        },
        slotsRemaining: 28, // Less than 30 so Finish Deck is enabled
        landSuggestions: mockLands,
      });

      render(
        <BuildAroundSeedModal
          {...iterativeProps}
          onFinishDeck={onFinishDeck}
        />
      );

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Finish' } });

      await waitFor(() => {
        expect(screen.getByText('Finish Deck Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Finish Deck Test'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      // Wait for iterative mode to load
      await waitFor(() => {
        expect(screen.getByText('28 slots remaining')).toBeInTheDocument();
      });

      // Click Finish Deck button
      const finishButton = screen.getByText('Finish Deck (Add Lands)');
      fireEvent.click(finishButton);

      expect(onFinishDeck).toHaveBeenCalledWith(mockLands);
    });

    it('should display deck analysis in iterative mode', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '88888',
          Name: 'Analysis Test',
          ManaCost: '{B}{G}',
          Types: ['Creature'],
          Colors: ['B', 'G'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [],
        deckAnalysis: {
          colorIdentity: ['B', 'G'],
          keywords: ['Deathtouch', 'Trample'],
          themes: ['Graveyard', 'Ramp'],
          currentCurve: { 1: 4, 2: 6, 3: 5, 4: 3 },
          recommendedLandCount: 24,
          totalCards: 18,
          inCollectionCount: 15,
        },
        slotsRemaining: 42,
        landSuggestions: [
          { cardID: 81718, name: 'Swamp', quantity: 12, color: 'B' },
          { cardID: 81720, name: 'Forest', quantity: 12, color: 'G' },
        ],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Analysis' } });

      await waitFor(() => {
        expect(screen.getByText('Analysis Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Analysis Test'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      // Wait for analysis to load
      await waitFor(() => {
        expect(screen.getByText('Deck Analysis')).toBeInTheDocument();
      });

      // Check analysis content
      expect(screen.getByText(/Total Cards: 18/)).toBeInTheDocument();
      expect(screen.getByText(/Recommended Lands: 24/)).toBeInTheDocument();

      // Check themes are displayed
      expect(screen.getByText('Graveyard')).toBeInTheDocument();
      expect(screen.getByText('Ramp')).toBeInTheDocument();
    });
  });

  describe('Copy Selection Modal', () => {
    const iterativeProps = {
      isOpen: true,
      onClose: vi.fn(),
      onApplyDeck: vi.fn(),
      onCardAdded: vi.fn(),
      onCardRemoved: vi.fn(),
      onFinishDeck: vi.fn(),
      currentDeckCards: [],
      deckCards: [],
    };

    // Helper to navigate to iterative mode and open copy modal
    // Must set up mocks BEFORE calling this
    const enterIterativeModeAndOpenCopyModal = async (cardName: string) => {
      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Test' } });

      await waitFor(() => {
        expect(screen.getByText('Test Seed Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Test Seed Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      fireEvent.click(suggestionCards[0]);

      await waitFor(() => {
        expect(screen.getByText(new RegExp(`Add ${cardName}`))).toBeInTheDocument();
      });
    };

    it('should close copy modal when cancel button is clicked', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Test Seed Card',
          ManaCost: '{2}{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 44444,
          name: 'Suggested Card',
          manaCost: '{1}{U}',
          cmc: 2,
          colors: ['U'],
          typeLine: 'Creature - Wizard',
          score: 0.75,
          reasoning: 'Good color match',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
        }],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      await enterIterativeModeAndOpenCopyModal('Suggested Card');

      // Click the cancel button inside the copy modal (not the main modal's cancel)
      const copyModal = document.querySelector('.copy-modal');
      expect(copyModal).not.toBeNull();
      const cancelButton = copyModal!.querySelector('.copy-modal-cancel') as HTMLButtonElement;
      expect(cancelButton).not.toBeNull();
      fireEvent.click(cancelButton);

      // Copy modal should close
      await waitFor(() => {
        expect(screen.queryByText(/Add Suggested Card/)).not.toBeInTheDocument();
      });
    });

    it('should close copy modal when clicking overlay', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Test Seed Card',
          ManaCost: '{2}{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 44444,
          name: 'Suggested Card',
          manaCost: '{1}{U}',
          cmc: 2,
          colors: ['U'],
          typeLine: 'Creature - Wizard',
          score: 0.75,
          reasoning: 'Good color match',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
        }],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      await enterIterativeModeAndOpenCopyModal('Suggested Card');

      // Click the overlay (background)
      const overlay = document.querySelector('.copy-modal-overlay');
      expect(overlay).not.toBeNull();
      fireEvent.click(overlay!);

      // Modal should close
      await waitFor(() => {
        expect(screen.queryByText(/Add Suggested Card/)).not.toBeInTheDocument();
      });
    });

    it('should display card info in copy modal', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Test Seed Card',
          ManaCost: '{2}{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 44444,
          name: 'Suggested Card',
          manaCost: '{1}{U}',
          cmc: 2,
          colors: ['U'],
          typeLine: 'Creature - Wizard',
          score: 0.75,
          reasoning: 'Good color match',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
        }],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      await enterIterativeModeAndOpenCopyModal('Suggested Card');

      // Check card info is displayed
      expect(screen.getByText('Creature - Wizard')).toBeInTheDocument();
      expect(screen.getByText('Good color match')).toBeInTheDocument();
      expect(screen.getByText(/In deck: 0/)).toBeInTheDocument();
      expect(screen.getByText(/Recommended: 4/)).toBeInTheDocument();
    });

    it('should disable copy buttons when they would exceed 4 copies', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Test Seed Card',
          ManaCost: '{2}{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 55555,
          name: 'Card With Copies',
          manaCost: '{U}',
          cmc: 1,
          colors: ['U'],
          typeLine: 'Instant',
          score: 0.8,
          reasoning: 'Great fit',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 2, // Already have 2 copies
          recommendedCopies: 4,
        }],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 2,
          inCollectionCount: 2,
        },
        slotsRemaining: 58,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      await enterIterativeModeAndOpenCopyModal('Card With Copies');

      // +1 and +2 should be enabled (2 + 1 = 3, 2 + 2 = 4)
      const addOneButton = screen.getByRole('button', { name: /\+1/i });
      const addTwoButton = screen.getByRole('button', { name: /\+2/i });
      expect(addOneButton).not.toBeDisabled();
      expect(addTwoButton).not.toBeDisabled();

      // +3 and +4 should be disabled (would exceed 4)
      const addThreeButton = screen.getByRole('button', { name: /\+3/i });
      const addFourButton = screen.getByRole('button', { name: /\+4/i });
      expect(addThreeButton).toBeDisabled();
      expect(addFourButton).toBeDisabled();
    });

    it('should show in-deck badge for cards already in deck', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Test Seed Card',
          ManaCost: '{2}{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 66666,
          name: 'Card Already In Deck',
          manaCost: '{U}{U}',
          cmc: 2,
          colors: ['U'],
          typeLine: 'Creature',
          score: 0.7,
          reasoning: 'Synergy',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 3,
          recommendedCopies: 4,
        }],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 3,
          inCollectionCount: 3,
        },
        slotsRemaining: 57,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Test' } });

      await waitFor(() => {
        expect(screen.getByText('Test Seed Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Test Seed Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Check for in-deck badge
      expect(screen.getByText('3 in deck')).toBeInTheDocument();

      // Card should have in-deck class
      const suggestionCard = document.querySelector('.clickable-suggestion-card.in-deck');
      expect(suggestionCard).not.toBeNull();
    });

    it('should call onCardRemoved when remove button is clicked', async () => {
      const onCardRemoved = vi.fn();

      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Test Seed Card',
          ManaCost: '{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [],
        deckAnalysis: {
          colorIdentity: ['R'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 2,
          inCollectionCount: 2,
        },
        slotsRemaining: 58,
        landSuggestions: [],
      });

      // Mock getCardByArenaId for card name lookup
      mockGetCardByArenaId.mockResolvedValue({
        ArenaID: '77777',
        Name: 'Removable Card',
      } as any);

      render(
        <BuildAroundSeedModal
          {...iterativeProps}
          onCardRemoved={onCardRemoved}
          deckCards={[{ CardID: 77777, Quantity: 2, Board: 'main' } as any]}
          currentDeckCards={[77777, 77777]}
        />
      );

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Test' } });

      await waitFor(() => {
        expect(screen.getByText('Test Seed Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Test Seed Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      // Wait for deck cards section to load
      await waitFor(() => {
        expect(screen.getByText(/Current Deck/)).toBeInTheDocument();
      });

      // Find and click remove button
      const removeButton = screen.getByRole('button', { name: /−/i });
      fireEvent.click(removeButton);

      expect(onCardRemoved).toHaveBeenCalledWith(77777);
    });
  });

  describe('Score Breakdown Display', () => {
    const iterativeProps = {
      isOpen: true,
      onClose: vi.fn(),
      onApplyDeck: vi.fn(),
      onCardAdded: vi.fn(),
      onFinishDeck: vi.fn(),
      currentDeckCards: [],
    };

    it('should display score breakdown bars in hover preview', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Score Test Card',
          ManaCost: '{2}{U}',
          Types: ['Creature'],
          Colors: ['U'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 44444,
          name: 'Card With Breakdown',
          manaCost: '{1}{U}',
          cmc: 2,
          colors: ['U'],
          typeLine: 'Creature - Wizard',
          score: 0.85,
          reasoning: 'High synergy card',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
          scoreBreakdown: {
            colorFit: 0.95,
            curveFit: 0.80,
            synergy: 0.90,
            quality: 0.75,
            overall: 0.85,
          },
          synergyDetails: [
            { type: 'keyword', name: 'flying', description: 'Matches 3 other flying creatures' },
            { type: 'theme', name: 'tokens', description: 'Supports tokens theme' },
          ],
        }],
        deckAnalysis: {
          colorIdentity: ['U'],
          keywords: ['Flying'],
          themes: ['Tokens'],
          currentCurve: { 2: 1 },
          recommendedLandCount: 24,
          totalCards: 1,
          inCollectionCount: 1,
        },
        slotsRemaining: 59,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Score' } });

      await waitFor(() => {
        expect(screen.getByText('Score Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Score Test Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Hover over the suggestion card to show preview
      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      expect(suggestionCards.length).toBeGreaterThan(0);
      fireEvent.mouseEnter(suggestionCards[0]);

      // Check that score breakdown section is visible
      await waitFor(() => {
        const scoreBreakdown = document.querySelector('.score-breakdown');
        expect(scoreBreakdown).toBeInTheDocument();
      });

      // Check that individual score bars exist
      const colorBar = document.querySelector('.score-bar.color-bar');
      const curveBar = document.querySelector('.score-bar.curve-bar');
      const synergyBar = document.querySelector('.score-bar.synergy-bar');
      const qualityBar = document.querySelector('.score-bar.quality-bar');

      expect(colorBar).toBeInTheDocument();
      expect(curveBar).toBeInTheDocument();
      expect(synergyBar).toBeInTheDocument();
      expect(qualityBar).toBeInTheDocument();
    });

    it('should display synergy details in hover preview', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Synergy Test Card',
          ManaCost: '{W}',
          Types: ['Creature'],
          Colors: ['W'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 55555,
          name: 'Card With Synergies',
          manaCost: '{1}{W}',
          cmc: 2,
          colors: ['W'],
          typeLine: 'Creature - Angel',
          score: 0.90,
          reasoning: 'Excellent synergy',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
          scoreBreakdown: {
            colorFit: 1.0,
            curveFit: 0.85,
            synergy: 0.95,
            quality: 0.80,
            overall: 0.90,
          },
          synergyDetails: [
            { type: 'keyword', name: 'flying', description: 'Matches 3 cards with flying' },
            { type: 'keyword', name: 'lifelink', description: 'Matches 2 cards with lifelink' },
            { type: 'creature_type', name: 'Angel', description: 'Angel tribal synergy' },
          ],
        }],
        deckAnalysis: {
          colorIdentity: ['W'],
          keywords: ['Flying', 'Lifelink'],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Synergy' } });

      await waitFor(() => {
        expect(screen.getByText('Synergy Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Synergy Test Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Hover over the suggestion card
      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      fireEvent.mouseEnter(suggestionCards[0]);

      // Check synergy details section exists
      await waitFor(() => {
        const synergyDetails = document.querySelector('.synergy-details');
        expect(synergyDetails).toBeInTheDocument();
      });

      // Check individual synergy items
      const synergyItems = document.querySelectorAll('.synergy-item');
      expect(synergyItems.length).toBe(3);
    });

    it('should toggle details expansion in copy modal', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Toggle Test Card',
          ManaCost: '{G}',
          Types: ['Creature'],
          Colors: ['G'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 66666,
          name: 'Expandable Card',
          manaCost: '{1}{G}',
          cmc: 2,
          colors: ['G'],
          typeLine: 'Creature - Elf',
          score: 0.80,
          reasoning: 'Good fit',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
          scoreBreakdown: {
            colorFit: 0.90,
            curveFit: 0.75,
            synergy: 0.80,
            quality: 0.70,
            overall: 0.80,
          },
          synergyDetails: [
            { type: 'creature_type', name: 'Elf', description: 'Elf tribal with 4 elves' },
          ],
        }],
        deckAnalysis: {
          colorIdentity: ['G'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Toggle' } });

      await waitFor(() => {
        expect(screen.getByText('Toggle Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Toggle Test Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Click suggestion to open copy modal
      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      fireEvent.click(suggestionCards[0]);

      await waitFor(() => {
        expect(screen.getByText(/Add Expandable Card/)).toBeInTheDocument();
      });

      // Should see "Show Details" button
      const detailsToggle = screen.getByText(/Show Details/);
      expect(detailsToggle).toBeInTheDocument();

      // Click to expand
      fireEvent.click(detailsToggle);

      // Should now show "Hide Details" and score breakdown
      await waitFor(() => {
        expect(screen.getByText(/Hide Details/)).toBeInTheDocument();
      });

      // Check that detailed score breakdown is visible in modal
      const modalScoreBreakdown = document.querySelector('.modal-score-breakdown');
      expect(modalScoreBreakdown).toBeInTheDocument();

      // Click to collapse
      fireEvent.click(screen.getByText(/Hide Details/));

      await waitFor(() => {
        expect(screen.getByText(/Show Details/)).toBeInTheDocument();
      });
    });

    it('should display score percentages in expanded copy modal', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Percentage Test Card',
          ManaCost: '{B}',
          Types: ['Creature'],
          Colors: ['B'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 77777,
          name: 'Percentage Display Card',
          manaCost: '{1}{B}',
          cmc: 2,
          colors: ['B'],
          typeLine: 'Creature - Zombie',
          score: 0.82,
          reasoning: 'Strong synergy with deck',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
          scoreBreakdown: {
            colorFit: 1.0,
            curveFit: 0.70,
            synergy: 0.85,
            quality: 0.60,
            overall: 0.82,
          },
          synergyDetails: [],
        }],
        deckAnalysis: {
          colorIdentity: ['B'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Percentage' } });

      await waitFor(() => {
        expect(screen.getByText('Percentage Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Percentage Test Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Click suggestion to open copy modal
      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      fireEvent.click(suggestionCards[0]);

      await waitFor(() => {
        expect(screen.getByText(/Add Percentage Display Card/)).toBeInTheDocument();
      });

      // Expand details
      fireEvent.click(screen.getByText(/Show Details/));

      // Check score labels and percentages are displayed (in separate spans)
      await waitFor(() => {
        expect(screen.getByText('Color Fit')).toBeInTheDocument();
        expect(screen.getByText('100%')).toBeInTheDocument();
        expect(screen.getByText('Curve Fit')).toBeInTheDocument();
        expect(screen.getByText('70%')).toBeInTheDocument();
        expect(screen.getByText('Synergy')).toBeInTheDocument();
        expect(screen.getByText('85%')).toBeInTheDocument();
        expect(screen.getByText('Quality')).toBeInTheDocument();
        expect(screen.getByText('60%')).toBeInTheDocument();
      });
    });

    it('should not show score breakdown when card has no breakdown data', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'No Breakdown Card',
          ManaCost: '{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockSuggestNextCards.mockResolvedValue({
        suggestions: [{
          cardID: 88888,
          name: 'Card Without Breakdown',
          manaCost: '{R}',
          cmc: 1,
          colors: ['R'],
          typeLine: 'Creature',
          score: 0.70,
          reasoning: 'Basic reasoning',
          inCollection: true,
          ownedCount: 4,
          neededCount: 0,
          currentCopies: 0,
          recommendedCopies: 4,
          // No scoreBreakdown or synergyDetails
        }],
        deckAnalysis: {
          colorIdentity: ['R'],
          keywords: [],
          themes: [],
          currentCurve: {},
          recommendedLandCount: 24,
          totalCards: 0,
          inCollectionCount: 0,
        },
        slotsRemaining: 60,
        landSuggestions: [],
      });

      render(<BuildAroundSeedModal {...iterativeProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'No Breakdown' } });

      await waitFor(() => {
        expect(screen.getByText('No Breakdown Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('No Breakdown Card'));

      await waitFor(() => {
        expect(screen.getByText('Start Building (Pick Cards)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Start Building (Pick Cards)'));

      await waitFor(() => {
        expect(screen.getByText('Click a card to add 1 copy to your deck')).toBeInTheDocument();
      });

      // Hover over the suggestion card
      const suggestionCards = document.querySelectorAll('.clickable-suggestion-card');
      fireEvent.mouseEnter(suggestionCards[0]);

      // Score breakdown should not be present
      await waitFor(() => {
        const scoreBreakdown = document.querySelector('.score-breakdown');
        expect(scoreBreakdown).not.toBeInTheDocument();
      });
    });
  });

  describe('Complete Deck Generation (Issue #774)', () => {
    const mockArchetypeProfiles = {
      aggro: {
        name: 'Aggro',
        landCount: 20,
        curveTargets: { 1: 8, 2: 14, 3: 10, 4: 4, 5: 4, 6: 0 },
        creatureRatio: 0.7,
        removalCount: 4,
        cardAdvantage: 2,
        description: 'Fast, aggressive deck that aims to win quickly with cheap threats.',
      },
      midrange: {
        name: 'Midrange',
        landCount: 24,
        curveTargets: { 1: 4, 2: 8, 3: 10, 4: 8, 5: 4, 6: 2 },
        creatureRatio: 0.55,
        removalCount: 6,
        cardAdvantage: 4,
        description: 'Balanced deck with efficient threats and answers.',
      },
      control: {
        name: 'Control',
        landCount: 26,
        curveTargets: { 1: 2, 2: 6, 3: 8, 4: 8, 5: 6, 6: 4 },
        creatureRatio: 0.25,
        removalCount: 10,
        cardAdvantage: 8,
        description: 'Slow, controlling deck that grinds out opponents.',
      },
    };

    const mockGeneratedDeck = {
      seedCard: {
        cardID: 12345,
        name: 'Seed Card',
        manaCost: '{2}{R}',
        cmc: 3,
        colors: ['R'],
        typeLine: 'Creature - Human',
        score: 1.0,
        reasoning: 'Seed card',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
        currentCopies: 0,
        recommendedCopies: 4,
      },
      spells: [
        {
          cardID: 11111,
          name: 'Aggressive Creature',
          manaCost: '{R}',
          cmc: 1,
          colors: ['R'],
          typeLine: 'Creature - Goblin',
          rarity: 'common',
          imageURI: '',
          quantity: 4,
          score: 0.9,
          reasoning: 'Fast creature',
          inCollection: true,
          ownedCount: 4,
        },
        {
          cardID: 22222,
          name: 'Burn Spell',
          manaCost: '{1}{R}',
          cmc: 2,
          colors: ['R'],
          typeLine: 'Instant',
          rarity: 'common',
          imageURI: '',
          quantity: 4,
          score: 0.85,
          reasoning: 'Removal spell',
          inCollection: true,
          ownedCount: 4,
        },
      ],
      lands: [
        {
          cardID: 81717,
          name: 'Mountain',
          quantity: 20,
          colors: ['R'],
          isBasic: true,
        },
      ],
      strategy: {
        summary: 'An aggressive mono-red deck focused on fast creatures and burn.',
        gamePlan: 'Deploy cheap threats and finish with burn spells.',
        keyCards: ['Aggressive Creature', 'Burn Spell'],
        mulligan: 'Keep hands with 2-3 lands and cheap creatures.',
        strengths: ['Fast', 'Consistent'],
        weaknesses: ['Weak to lifegain', 'Runs out of gas'],
      },
      analysis: {
        totalCards: 60,
        spellCount: 40,
        landCount: 20,
        creatureCount: 28,
        nonCreatureCount: 12,
        averageCMC: 2.1,
        manaCurve: { 1: 8, 2: 12, 3: 10, 4: 6, 5: 4 },
        colorDistribution: { R: 40 },
        inCollectionCount: 55,
        missingCount: 5,
        missingWildcardCost: { common: 3, uncommon: 2 },
        archetypeMatch: 0.85,
      },
    };

    it('should show Quick Generate button when card is selected', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Generate Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Generate' } });

      await waitFor(() => {
        expect(screen.getByText('Generate Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Generate Test Card'));

      await waitFor(() => {
        expect(screen.getByText('Quick Generate (60-Card Deck)')).toBeInTheDocument();
      });
    });

    it('should show archetype selector when Quick Generate is clicked', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Archetype Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: 'http://example.com/card.png',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Archetype' } });

      await waitFor(() => {
        expect(screen.getByText('Archetype Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Archetype Test Card'));

      await waitFor(() => {
        expect(screen.getByText('Quick Generate (60-Card Deck)')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Quick Generate (60-Card Deck)'));

      // Should show archetype selector
      await waitFor(() => {
        expect(screen.getByText('Choose Deck Archetype')).toBeInTheDocument();
        expect(screen.getByText('Building around: Archetype Test Card')).toBeInTheDocument();
      });

      // Should show all three archetype options
      expect(screen.getByText('Aggro')).toBeInTheDocument();
      expect(screen.getByText('Midrange')).toBeInTheDocument();
      expect(screen.getByText('Control')).toBeInTheDocument();
    });

    it('should display archetype descriptions and stats', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Desc Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Desc' } });

      await waitFor(() => {
        expect(screen.getByText('Desc Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Desc Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        // Check archetype stats are displayed
        expect(screen.getByText('20 lands')).toBeInTheDocument();
        expect(screen.getByText('70% creatures')).toBeInTheDocument();
        expect(screen.getByText('24 lands')).toBeInTheDocument();
        expect(screen.getByText('55% creatures')).toBeInTheDocument();
        expect(screen.getByText('26 lands')).toBeInTheDocument();
        expect(screen.getByText('25% creatures')).toBeInTheDocument();
      });
    });

    it('should generate deck when archetype is selected', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Gen Deck Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Gen' } });

      await waitFor(() => {
        expect(screen.getByText('Gen Deck Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Gen Deck Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      // Click Aggro archetype
      fireEvent.click(screen.getByText('Aggro'));

      // Should call the API with correct parameters
      await waitFor(() => {
        expect(mockGenerateCompleteDeck).toHaveBeenCalledWith({
          seed_card_id: 12345,
          archetype: 'aggro',
          budget_mode: false,
        });
      });
    });

    it('should display generated deck result with strategy', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Result Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Result' } });

      await waitFor(() => {
        expect(screen.getByText('Result Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Result Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      // Should show generated deck result
      await waitFor(() => {
        expect(screen.getByText(/Generated Aggro Deck/i)).toBeInTheDocument();
      });

      // Strategy panel
      expect(screen.getByText('Deck Strategy')).toBeInTheDocument();
      expect(screen.getByText(mockGeneratedDeck.strategy.summary)).toBeInTheDocument();

      // Game plan
      expect(screen.getByText('Game Plan')).toBeInTheDocument();
      expect(screen.getByText(mockGeneratedDeck.strategy.gamePlan)).toBeInTheDocument();

      // Mulligan guide
      expect(screen.getByText('Mulligan Guide')).toBeInTheDocument();
      expect(screen.getByText(mockGeneratedDeck.strategy.mulligan)).toBeInTheDocument();
    });

    it('should display deck statistics in generated result', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Stats Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Stats' } });

      await waitFor(() => {
        expect(screen.getByText('Stats Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Stats Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText(/Generated Aggro Deck/i)).toBeInTheDocument();
      });

      // Check stats display
      expect(screen.getByText('60')).toBeInTheDocument(); // Total Cards
      expect(screen.getByText('40')).toBeInTheDocument(); // Spells
      expect(screen.getByText('20')).toBeInTheDocument(); // Lands
      expect(screen.getByText('2.10')).toBeInTheDocument(); // Avg CMC
    });

    it('should display card lists grouped by type', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'List Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'List' } });

      await waitFor(() => {
        expect(screen.getByText('List Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('List Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText(/Generated Aggro Deck/i)).toBeInTheDocument();
      });

      // Check card lists are displayed (use getAllByText since cards may appear in multiple sections)
      expect(screen.getAllByText('Aggressive Creature').length).toBeGreaterThan(0);
      expect(screen.getAllByText('Burn Spell').length).toBeGreaterThan(0);
      expect(screen.getByText('Mountain')).toBeInTheDocument();
    });

    it('should have Apply Deck button that calls onApplyDeck', async () => {
      const onApplyDeck = vi.fn();
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Apply Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} onApplyDeck={onApplyDeck} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Apply' } });

      await waitFor(() => {
        expect(screen.getByText('Apply Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Apply Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText('Apply Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Apply Deck'));

      expect(onApplyDeck).toHaveBeenCalled();
    });

    it('should allow trying a different archetype', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Retry Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Retry' } });

      await waitFor(() => {
        expect(screen.getByText('Retry Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Retry Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText('Try Different Archetype')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Try Different Archetype'));

      // Should go back to archetype selector
      await waitFor(() => {
        expect(screen.getByText('Choose Deck Archetype')).toBeInTheDocument();
      });
    });

    it('should respect budget mode when generating deck', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Budget Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Budget' } });

      await waitFor(() => {
        expect(screen.getByText('Budget Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Budget Test Card'));

      // Enable budget mode
      const budgetCheckbox = screen.getByLabelText(/budget mode/i);
      fireEvent.click(budgetCheckbox);

      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Midrange')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Midrange'));

      await waitFor(() => {
        expect(mockGenerateCompleteDeck).toHaveBeenCalledWith({
          seed_card_id: 12345,
          archetype: 'midrange',
          budget_mode: true,
        });
      });
    });

    it('should show back button in archetype selector', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Back Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Back' } });

      await waitFor(() => {
        expect(screen.getByText('Back Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Back Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Back to Card Selection')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Back to Card Selection'));

      // Should go back to card selection
      await waitFor(() => {
        expect(screen.getByText('Build Around Card')).toBeInTheDocument();
        expect(screen.getByText('Quick Generate (60-Card Deck)')).toBeInTheDocument();
      });
    });

    it('should display strengths and weaknesses in strategy panel', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Strengths Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Strengths' } });

      await waitFor(() => {
        expect(screen.getByText('Strengths Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Strengths Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Control')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Control'));

      await waitFor(() => {
        expect(screen.getByText('Strengths')).toBeInTheDocument();
        expect(screen.getByText('Weaknesses')).toBeInTheDocument();
      });

      // Check actual strength/weakness content
      expect(screen.getByText('Fast')).toBeInTheDocument();
      expect(screen.getByText('Consistent')).toBeInTheDocument();
      expect(screen.getByText('Weak to lifegain')).toBeInTheDocument();
      expect(screen.getByText('Runs out of gas')).toBeInTheDocument();
    });

    it('should display key cards in strategy panel', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Key Cards Test',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Key' } });

      await waitFor(() => {
        expect(screen.getByText('Key Cards Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Key Cards Test'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText('Key Cards')).toBeInTheDocument();
      });

      // Key cards are displayed as tags
      const keyCardTags = document.querySelectorAll('.key-card-tag');
      expect(keyCardTags.length).toBe(2);
    });

    it('should show wildcard cost when missing cards', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Wildcard Test',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockResolvedValue(mockGeneratedDeck);

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Wildcard' } });

      await waitFor(() => {
        expect(screen.getByText('Wildcard Test')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Wildcard Test'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText('Wildcards needed:')).toBeInTheDocument();
      });

      // Should show wildcard badges
      expect(screen.getByText('3 common')).toBeInTheDocument();
      expect(screen.getByText('2 uncommon')).toBeInTheDocument();
    });

    it('should show error when deck generation fails', async () => {
      mockSearchCardsWithCollection.mockResolvedValue([
        {
          ArenaID: '12345',
          Name: 'Error Test Card',
          ManaCost: '{2}{R}',
          Types: ['Creature'],
          Colors: ['R'],
          ImageURL: '',
        },
      ] as any);

      mockGetArchetypeProfiles.mockResolvedValue(mockArchetypeProfiles);
      mockGenerateCompleteDeck.mockRejectedValue(new Error('Generation failed'));

      render(<BuildAroundSeedModal {...defaultProps} />);

      const searchInput = screen.getByPlaceholderText(/search for a card/i);
      fireEvent.change(searchInput, { target: { value: 'Error' } });

      await waitFor(() => {
        expect(screen.getByText('Error Test Card')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Error Test Card'));
      fireEvent.click(await screen.findByText('Quick Generate (60-Card Deck)'));

      await waitFor(() => {
        expect(screen.getByText('Aggro')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Aggro'));

      await waitFor(() => {
        expect(screen.getByText('Generation failed')).toBeInTheDocument();
      });
    });
  });
});

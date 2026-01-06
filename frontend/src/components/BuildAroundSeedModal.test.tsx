import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import BuildAroundSeedModal from './BuildAroundSeedModal';

// Mock the API services
vi.mock('@/services/api', () => ({
  decks: {
    buildAroundSeed: vi.fn(),
  },
  cards: {
    searchCards: vi.fn(),
  },
}));

import { decks, cards } from '@/services/api';

const mockBuildAroundSeed = vi.mocked(decks.buildAroundSeed);
const mockSearchCards = vi.mocked(cards.searchCards);

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

    const closeButton = screen.getByRole('button', { name: /×/ });
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
    mockSearchCards.mockResolvedValue([
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
      expect(mockSearchCards).toHaveBeenCalledWith({ query: 'Test', limit: 10 });
    });
  });

  it('should display search results', async () => {
    mockSearchCards.mockResolvedValue([
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
    mockSearchCards.mockResolvedValue([
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

    // Should show build button
    await waitFor(() => {
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });
  });

  it('should toggle budget mode checkbox', async () => {
    mockSearchCards.mockResolvedValue([
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

  it('should build suggestions when button is clicked', async () => {
    mockSearchCards.mockResolvedValue([
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
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });

    // Click build button
    fireEvent.click(screen.getByText('Build Around This Card'));

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

  it('should show suggestions after build', async () => {
    mockSearchCards.mockResolvedValue([
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
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Build Around This Card'));

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
    mockSearchCards.mockResolvedValue([
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
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Build Around This Card'));

    await waitFor(() => {
      expect(screen.getByText('Own 3')).toBeInTheDocument();
      expect(screen.getByText('Need 4')).toBeInTheDocument();
    });
  });

  it('should call onApplyDeck when apply button is clicked', async () => {
    const onApplyDeck = vi.fn();

    mockSearchCards.mockResolvedValue([
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
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Build Around This Card'));

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
    mockSearchCards.mockResolvedValue([
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
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Build Around This Card'));

    await waitFor(() => {
      expect(screen.getByText('API Error')).toBeInTheDocument();
    });
  });

  it('should clear selection when clear button is clicked', async () => {
    mockSearchCards.mockResolvedValue([
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
      expect(screen.getByText('Build Around This Card')).toBeInTheDocument();
    });

    // Click clear button
    fireEvent.click(screen.getByRole('button', { name: /clear/i }));

    // Build button should be gone
    expect(screen.queryByText('Build Around This Card')).not.toBeInTheDocument();
  });
});

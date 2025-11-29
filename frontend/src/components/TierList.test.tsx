import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import TierList from './TierList';
import { mockWailsApp } from '../test/mocks/wailsApp';
import { gui } from '../../wailsjs/go/models';

// Helper function to create mock card rating
function createMockCardRating(overrides: Partial<gui.CardRatingWithTier> = {}): gui.CardRatingWithTier {
  return new gui.CardRatingWithTier({
    name: 'Test Card',
    color: 'U',
    rarity: 'common',
    mtga_id: 12345,
    ever_drawn_win_rate: 55.5,
    opening_hand_win_rate: 54.0,
    ever_drawn_game_win_rate: 53.0,
    drawn_win_rate: 52.0,
    in_hand_win_rate: 51.0,
    ever_drawn_improvement_win_rate: 5.0,
    opening_hand_improvement_win_rate: 4.0,
    drawn_improvement_win_rate: 3.0,
    in_hand_improvement_win_rate: 2.0,
    avg_seen: 3.5,
    avg_pick: 2.5,
    pick_rate: 30.0,
    '# ever_drawn': 1000,
    '# opening_hand': 500,
    '# games': 2000,
    '# drawn': 800,
    '# in_hand_drawn': 400,
    '# games_played': 1500,
    '# decks': 100,
    tier: 'B',
    colors: ['U'],
    ...overrides,
  });
}

describe('TierList Component', () => {
  const defaultProps = {
    setCode: 'TST',
    draftFormat: 'QuickDraft',
    pickedCardIds: new Set<string>(),
    onCardClick: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockWailsApp.GetCardRatings.mockResolvedValue([]);
    mockWailsApp.GetSetCards.mockResolvedValue([]);
  });

  describe('Loading State', () => {
    it('should display loading state initially', () => {
      mockWailsApp.GetCardRatings.mockImplementation(() => new Promise(() => {}));
      mockWailsApp.GetSetCards.mockImplementation(() => new Promise(() => {}));

      render(<TierList {...defaultProps} />);

      expect(screen.getByText('Loading card ratings...')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should display error message when loading fails', async () => {
      mockWailsApp.GetCardRatings.mockRejectedValue(new Error('Failed to fetch'));
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText(/Failed to fetch/)).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should display empty state when no ratings available', async () => {
      mockWailsApp.GetCardRatings.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText(/No card ratings available/)).toBeInTheDocument();
      });
    });
  });

  describe('Search Functionality', () => {
    it('should render search input', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Counterspell', mtga_id: 2, tier: 'A' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByPlaceholderText('Search by card name...')).toBeInTheDocument();
      });
    });

    it('should filter cards by search term', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Counterspell', mtga_id: 2, tier: 'A' }),
        createMockCardRating({ name: 'Giant Growth', mtga_id: 3, tier: 'B' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      // Wait for cards to load
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // Type in search box
      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'lightning');

      // Only Lightning Bolt should be visible
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.queryByText('Counterspell')).not.toBeInTheDocument();
        expect(screen.queryByText('Giant Growth')).not.toBeInTheDocument();
      });
    });

    it('should be case-insensitive when searching', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'LIGHTNING');

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });
    });

    it('should show clear button when search term exists', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // Clear button should not be visible initially
      expect(screen.queryByTitle('Clear search')).not.toBeInTheDocument();

      // Type in search
      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'bolt');

      // Clear button should appear
      await waitFor(() => {
        expect(screen.getByTitle('Clear search')).toBeInTheDocument();
      });
    });

    it('should clear search when clear button is clicked', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Counterspell', mtga_id: 2, tier: 'A' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Counterspell')).toBeInTheDocument();
      });

      // Search for "lightning"
      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'lightning');

      // Only Lightning Bolt visible
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.queryByText('Counterspell')).not.toBeInTheDocument();
      });

      // Click clear button
      const clearButton = screen.getByTitle('Clear search');
      await userEvent.click(clearButton);

      // Both cards should be visible again
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Counterspell')).toBeInTheDocument();
      });
    });

    it('should show no results when search matches nothing', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'nonexistent');

      // The card should not be in any tier group
      await waitFor(() => {
        expect(screen.queryByText('Lightning Bolt')).not.toBeInTheDocument();
      });
    });

    it('should work with partial matches', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Chain Lightning', mtga_id: 2, tier: 'A' }),
        createMockCardRating({ name: 'Counterspell', mtga_id: 3, tier: 'B' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'light');

      // Both Lightning cards should be visible
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
        expect(screen.getByText('Chain Lightning')).toBeInTheDocument();
        expect(screen.queryByText('Counterspell')).not.toBeInTheDocument();
      });
    });

    it('should update filtered count when searching', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Counterspell', mtga_id: 2, tier: 'A' }),
        createMockCardRating({ name: 'Giant Growth', mtga_id: 3, tier: 'B' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      // Wait for initial load with 3 cards - check the header info span
      await waitFor(() => {
        const infoSpan = document.querySelector('.tier-list-info span');
        expect(infoSpan?.textContent).toContain('3 cards');
      });

      // Search for "lightning"
      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'lightning');

      // Should show 1 card now in the header
      await waitFor(() => {
        const infoSpan = document.querySelector('.tier-list-info span');
        expect(infoSpan?.textContent).toContain('1 cards');
      });
    });
  });

  describe('Tier Filtering', () => {
    it('should display cards grouped by tier', async () => {
      const cards = [
        createMockCardRating({ name: 'S-Tier Card', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'A-Tier Card', mtga_id: 2, tier: 'A' }),
        createMockCardRating({ name: 'B-Tier Card', mtga_id: 3, tier: 'B' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('S-Tier Card')).toBeInTheDocument();
        expect(screen.getByText('A-Tier Card')).toBeInTheDocument();
        expect(screen.getByText('B-Tier Card')).toBeInTheDocument();
      });
    });

    it('should filter by tier when tier button is clicked', async () => {
      const cards = [
        createMockCardRating({ name: 'S-Tier Card', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'A-Tier Card', mtga_id: 2, tier: 'A' }),
        createMockCardRating({ name: 'F-Tier Card', mtga_id: 3, tier: 'F' }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('S-Tier Card')).toBeInTheDocument();
      });

      // Click F tier button to deselect it
      const fTierButton = screen.getByRole('button', { name: 'F' });
      await userEvent.click(fTierButton);

      // F-Tier card should be hidden
      await waitFor(() => {
        expect(screen.getByText('S-Tier Card')).toBeInTheDocument();
        expect(screen.getByText('A-Tier Card')).toBeInTheDocument();
        expect(screen.queryByText('F-Tier Card')).not.toBeInTheDocument();
      });
    });
  });

  describe('Color Filtering', () => {
    it('should filter by color when color button is clicked', async () => {
      const cards = [
        createMockCardRating({ name: 'Blue Card', mtga_id: 1, tier: 'S', color: 'U', colors: ['U'] }),
        createMockCardRating({ name: 'Red Card', mtga_id: 2, tier: 'S', color: 'R', colors: ['R'] }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
        expect(screen.getByText('Red Card')).toBeInTheDocument();
      });

      // Click Blue color filter - the button text is emoji + " U"
      const colorButtons = screen.getAllByRole('button');
      const blueButton = colorButtons.find(btn => btn.textContent?.includes('U') && btn.classList.contains('color-btn'));
      expect(blueButton).toBeDefined();
      await userEvent.click(blueButton!);

      // Only Blue card should be visible
      await waitFor(() => {
        expect(screen.getByText('Blue Card')).toBeInTheDocument();
        expect(screen.queryByText('Red Card')).not.toBeInTheDocument();
      });
    });
  });

  describe('Sorting', () => {
    it('should sort by GIHWR by default (descending)', async () => {
      const cards = [
        createMockCardRating({ name: 'Low WR Card', mtga_id: 1, tier: 'S', ever_drawn_win_rate: 50.0 }),
        createMockCardRating({ name: 'High WR Card', mtga_id: 2, tier: 'S', ever_drawn_win_rate: 65.0 }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        const rows = screen.getAllByRole('row');
        // First data row should be High WR Card (sorted descending by GIHWR)
        expect(rows[1]).toHaveTextContent('High WR Card');
        expect(rows[2]).toHaveTextContent('Low WR Card');
      });
    });
  });

  describe('Combined Filters', () => {
    it('should combine search with other filters', async () => {
      const cards = [
        createMockCardRating({ name: 'Blue Lightning', mtga_id: 1, tier: 'S', color: 'U', colors: ['U'] }),
        createMockCardRating({ name: 'Red Lightning', mtga_id: 2, tier: 'S', color: 'R', colors: ['R'] }),
        createMockCardRating({ name: 'Blue Spell', mtga_id: 3, tier: 'S', color: 'U', colors: ['U'] }),
      ];
      mockWailsApp.GetCardRatings.mockResolvedValue(cards);
      mockWailsApp.GetSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Blue Lightning')).toBeInTheDocument();
      });

      // Filter by blue color - find the color button with class
      const colorButtons = screen.getAllByRole('button');
      const blueButton = colorButtons.find(btn => btn.textContent?.includes('U') && btn.classList.contains('color-btn'));
      expect(blueButton).toBeDefined();
      await userEvent.click(blueButton!);

      // Search for "lightning"
      const searchInput = screen.getByPlaceholderText('Search by card name...');
      await userEvent.type(searchInput, 'lightning');

      // Only Blue Lightning should be visible
      await waitFor(() => {
        expect(screen.getByText('Blue Lightning')).toBeInTheDocument();
        expect(screen.queryByText('Red Lightning')).not.toBeInTheDocument();
        expect(screen.queryByText('Blue Spell')).not.toBeInTheDocument();
      });
    });
  });
});

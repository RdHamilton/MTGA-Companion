import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import TierList from './TierList';
import { mockCards } from '@/test/mocks/apiMock';
import { gui } from '@/types/models';

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
    mockCards.getCardRatings.mockResolvedValue([]);
    mockCards.getSetCards.mockResolvedValue([]);
  });

  describe('Loading State', () => {
    it('should display loading state initially', () => {
      mockCards.getCardRatings.mockImplementation(() => new Promise(() => {}));
      mockCards.getSetCards.mockImplementation(() => new Promise(() => {}));

      render(<TierList {...defaultProps} />);

      expect(screen.getByText('Loading card ratings...')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should display error message when loading fails', async () => {
      mockCards.getCardRatings.mockRejectedValue(new Error('Failed to fetch'));
      mockCards.getSetCards.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText(/Failed to fetch/)).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should display empty state when no ratings available', async () => {
      mockCards.getCardRatings.mockResolvedValue([]);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);

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

  describe('CFB Ratings Integration', () => {
    beforeEach(() => {
      mockCards.getCFBRatings.mockResolvedValue([]);
    });

    it('should not show CFB column when no CFB ratings are available', async () => {
      const cards = [
        createMockCardRating({ name: 'Test Card', mtga_id: 1, tier: 'S' }),
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue([]);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Test Card')).toBeInTheDocument();
      });

      // CFB column should not be present
      expect(screen.queryByText('CFB')).not.toBeInTheDocument();
    });

    it('should show CFB column when CFB ratings are available', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      const cfbRatings = [
        {
          id: 1,
          cardName: 'Lightning Bolt',
          setCode: 'TST',
          limitedRating: 'A+',
          limitedScore: 10,
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue(cfbRatings);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // CFB column header should be present
      await waitFor(() => {
        const headers = screen.getAllByRole('columnheader');
        const cfbHeader = headers.find(h => h.textContent === 'CFB');
        expect(cfbHeader).toBeInTheDocument();
      });
    });

    it('should display CFB rating badge for cards with ratings', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      const cfbRatings = [
        {
          id: 1,
          cardName: 'Lightning Bolt',
          setCode: 'TST',
          limitedRating: 'A+',
          limitedScore: 10,
          commentary: 'Top tier removal',
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue(cfbRatings);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // CFB rating badge should be present with the grade
      await waitFor(() => {
        const badge = screen.getByTestId('cfb-rating-badge');
        expect(badge).toBeInTheDocument();
        expect(badge).toHaveTextContent('A+');
      });
    });

    it('should show dash for cards without CFB rating', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Counterspell', mtga_id: 2, tier: 'A' }),
      ];
      const cfbRatings = [
        {
          id: 1,
          cardName: 'Lightning Bolt',
          setCode: 'TST',
          limitedRating: 'A+',
          limitedScore: 10,
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
        // Note: No CFB rating for Counterspell
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue(cfbRatings);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Counterspell')).toBeInTheDocument();
      });

      // Should have a dash for the missing CFB rating
      await waitFor(() => {
        const noCfbElements = document.querySelectorAll('.no-cfb-rating');
        expect(noCfbElements.length).toBeGreaterThan(0);
        expect(noCfbElements[0]).toHaveTextContent('-');
      });
    });

    it('should match CFB ratings case-insensitively', async () => {
      const cards = [
        createMockCardRating({ name: 'LIGHTNING BOLT', mtga_id: 1, tier: 'S' }),
      ];
      const cfbRatings = [
        {
          id: 1,
          cardName: 'lightning bolt', // lowercase
          setCode: 'TST',
          limitedRating: 'B+',
          limitedScore: 7,
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue(cfbRatings);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('LIGHTNING BOLT')).toBeInTheDocument();
      });

      // CFB rating should still be found and displayed
      await waitFor(() => {
        const badge = screen.getByTestId('cfb-rating-badge');
        expect(badge).toBeInTheDocument();
        expect(badge).toHaveTextContent('B+');
      });
    });

    it('should handle CFB ratings API failure gracefully', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockRejectedValue(new Error('CFB ratings not available'));

      render(<TierList {...defaultProps} />);

      // Should still render the tier list without CFB column
      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // CFB column should not be present when API fails
      expect(screen.queryByText('CFB')).not.toBeInTheDocument();
    });

    it('should display CFB rating tooltip with commentary', async () => {
      const cards = [
        createMockCardRating({ name: 'Lightning Bolt', mtga_id: 1, tier: 'S' }),
      ];
      const cfbRatings = [
        {
          id: 1,
          cardName: 'Lightning Bolt',
          setCode: 'TST',
          limitedRating: 'A+',
          limitedScore: 10,
          commentary: 'Best removal in the format',
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue(cfbRatings);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Lightning Bolt')).toBeInTheDocument();
      });

      // Badge should have commentary as tooltip
      await waitFor(() => {
        const badge = screen.getByTestId('cfb-rating-badge');
        expect(badge).toHaveAttribute('title', 'Best removal in the format');
      });
    });

    it('should display multiple cards with different CFB grades', async () => {
      const cards = [
        createMockCardRating({ name: 'Bomb Rare', mtga_id: 1, tier: 'S' }),
        createMockCardRating({ name: 'Good Card', mtga_id: 2, tier: 'A' }),
        createMockCardRating({ name: 'Filler', mtga_id: 3, tier: 'C' }),
      ];
      const cfbRatings = [
        {
          id: 1,
          cardName: 'Bomb Rare',
          setCode: 'TST',
          limitedRating: 'A+',
          limitedScore: 10,
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
        {
          id: 2,
          cardName: 'Good Card',
          setCode: 'TST',
          limitedRating: 'B',
          limitedScore: 6,
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
        {
          id: 3,
          cardName: 'Filler',
          setCode: 'TST',
          limitedRating: 'C-',
          limitedScore: 2,
          importedAt: '2024-01-01',
          updatedAt: '2024-01-01',
        },
      ];
      mockCards.getCardRatings.mockResolvedValue(cards);
      mockCards.getSetCards.mockResolvedValue([]);
      mockCards.getCFBRatings.mockResolvedValue(cfbRatings);

      render(<TierList {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Bomb Rare')).toBeInTheDocument();
        expect(screen.getByText('Good Card')).toBeInTheDocument();
        expect(screen.getByText('Filler')).toBeInTheDocument();
      });

      // All CFB badges should be present
      await waitFor(() => {
        const badges = screen.getAllByTestId('cfb-rating-badge');
        expect(badges).toHaveLength(3);
        expect(badges[0]).toHaveTextContent('A+');
        expect(badges[1]).toHaveTextContent('B');
        expect(badges[2]).toHaveTextContent('C-');
      });
    });
  });
});

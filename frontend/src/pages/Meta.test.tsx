import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Meta from './Meta';
import { gui, time } from '../../wailsjs/go/models';

// Mock the Wails App functions
vi.mock('../../wailsjs/go/main/App', () => ({
  GetMetaDashboard: vi.fn(),
  RefreshMetaData: vi.fn(),
  GetSupportedFormats: vi.fn(),
}));

import {
  GetMetaDashboard,
  RefreshMetaData,
  GetSupportedFormats,
} from '../../wailsjs/go/main/App';

const mockGetMetaDashboard = vi.mocked(GetMetaDashboard);
const mockRefreshMetaData = vi.mocked(RefreshMetaData);
const mockGetSupportedFormats = vi.mocked(GetSupportedFormats);

const renderMeta = () => {
  return render(
    <BrowserRouter>
      <Meta />
    </BrowserRouter>
  );
};

// Create a mock time.Time that works with JSON
const createMockTime = (dateStr: string): time.Time => {
  const mockTime = new time.Time({});
  // Override toString to return the date string
  mockTime.toString = () => dateStr;
  return mockTime;
};

const createMockDashboardData = (overrides: Partial<{
  format: string;
  archetypes: gui.ArchetypeInfo[];
  tournaments: gui.TournamentInfo[];
  totalArchetypes: number;
  lastUpdated: time.Time;
  sources: string[];
  error: string;
}> = {}): gui.MetaDashboardResponse => {
  const response = new gui.MetaDashboardResponse({});

  response.format = overrides.format ?? 'standard';
  response.archetypes = overrides.archetypes ?? [
    Object.assign(new gui.ArchetypeInfo({}), {
      name: 'Mono Red Aggro',
      colors: ['R'],
      metaShare: 15.5,
      tournamentTop8s: 12,
      tournamentWins: 3,
      tier: 1,
      confidenceScore: 0.95,
      trendDirection: 'up',
    }),
    Object.assign(new gui.ArchetypeInfo({}), {
      name: 'Azorius Control',
      colors: ['W', 'U'],
      metaShare: 10.2,
      tournamentTop8s: 8,
      tournamentWins: 2,
      tier: 1,
      confidenceScore: 0.88,
      trendDirection: 'stable',
    }),
    Object.assign(new gui.ArchetypeInfo({}), {
      name: 'Golgari Midrange',
      colors: ['B', 'G'],
      metaShare: 5.5,
      tournamentTop8s: 4,
      tournamentWins: 0,
      tier: 2,
      confidenceScore: 0.72,
      trendDirection: 'down',
    }),
  ];
  response.tournaments = overrides.tournaments ?? [
    Object.assign(new gui.TournamentInfo({}), {
      name: 'Pro Tour Test',
      date: createMockTime('2024-01-15T00:00:00Z'),
      players: 256,
      format: 'standard',
      topDecks: ['Mono Red Aggro', 'Azorius Control', 'Golgari Midrange'],
      sourceUrl: 'https://example.com/tournament',
    }),
  ];
  response.totalArchetypes = overrides.totalArchetypes ?? 3;
  response.lastUpdated = overrides.lastUpdated ?? createMockTime('2024-01-20T12:00:00Z');
  response.sources = overrides.sources ?? ['MTGGoldfish', 'MTGTop8'];
  response.error = overrides.error ?? '';

  return response;
};

describe('Meta', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetSupportedFormats.mockResolvedValue(['standard', 'historic', 'explorer']);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('rendering', () => {
    it('renders the page title', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      expect(screen.getByText('Metagame Dashboard')).toBeInTheDocument();
    });

    it('renders the format selector', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByRole('combobox')).toBeInTheDocument();
      });
    });

    it('renders the refresh button', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      expect(screen.getByRole('button', { name: /refresh/i })).toBeInTheDocument();
    });

    it('shows loading state initially', async () => {
      mockGetMetaDashboard.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(createMockDashboardData()), 100))
      );

      renderMeta();

      expect(screen.getByText(/loading meta data/i)).toBeInTheDocument();
    });

    it('displays archetype data after loading', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      expect(screen.getByText('Azorius Control')).toBeInTheDocument();
      expect(screen.getByText('Golgari Midrange')).toBeInTheDocument();
    });

    it('displays tier badges', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getAllByText('Tier 1').length).toBeGreaterThan(0);
      });

      expect(screen.getAllByText('Tier 2').length).toBeGreaterThan(0);
    });

    it('displays meta share percentages', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('15.5% meta share')).toBeInTheDocument();
      });

      expect(screen.getByText('10.2% meta share')).toBeInTheDocument();
    });

    it('displays tournament top 8s', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('12 Top 8s')).toBeInTheDocument();
      });
    });

    it('displays tournament wins', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('3 Wins')).toBeInTheDocument();
      });
    });

    it('displays data sources', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('MTGGoldfish, MTGTop8')).toBeInTheDocument();
      });
    });

    it('displays total archetypes count', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('3')).toBeInTheDocument();
      });
    });
  });

  describe('tournaments section', () => {
    it('renders tournament information', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Pro Tour Test')).toBeInTheDocument();
      });
    });

    it('displays tournament player count', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('256 players')).toBeInTheDocument();
      });
    });

    it('displays tournament top decks', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText(/Top Decks:/)).toBeInTheDocument();
      });
    });

    it('renders tournament link when sourceUrl is present', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        const link = screen.getByText('View Details →');
        expect(link).toBeInTheDocument();
        expect(link).toHaveAttribute('href', 'https://example.com/tournament');
      });
    });
  });

  describe('error handling', () => {
    it('displays error message when API returns error', async () => {
      mockGetMetaDashboard.mockResolvedValue(
        createMockDashboardData({ error: 'Failed to fetch data' })
      );

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText(/failed to fetch data/i)).toBeInTheDocument();
      });
    });

    it('displays error when API call fails', async () => {
      mockGetMetaDashboard.mockRejectedValue(new Error('Network error'));

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });

    it('shows no data message when archetypes are empty', async () => {
      mockGetMetaDashboard.mockResolvedValue(
        createMockDashboardData({ archetypes: [], tournaments: [] })
      );

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('No Meta Data Available')).toBeInTheDocument();
      });
    });
  });

  describe('format selection', () => {
    it('loads supported formats on mount', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(mockGetSupportedFormats).toHaveBeenCalledTimes(1);
      });
    });

    it('changes format when selection changes', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      // Wait for initial load to complete
      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      // Clear the mock to track only the new call
      mockGetMetaDashboard.mockClear();

      const select = screen.getByRole('combobox');
      fireEvent.change(select, { target: { value: 'historic' } });

      await waitFor(() => {
        expect(mockGetMetaDashboard).toHaveBeenCalledWith('historic');
      });
    });

    it('renders all supported formats as options', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Standard')).toBeInTheDocument();
      });

      expect(screen.getByText('Historic')).toBeInTheDocument();
      expect(screen.getByText('Explorer')).toBeInTheDocument();
    });
  });

  describe('refresh functionality', () => {
    it('calls RefreshMetaData when refresh button is clicked', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());
      mockRefreshMetaData.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const refreshButton = screen.getByRole('button', { name: /refresh/i });
      fireEvent.click(refreshButton);

      await waitFor(() => {
        expect(mockRefreshMetaData).toHaveBeenCalledWith('standard');
      });
    });

    it('shows refreshing state when refreshing', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());
      mockRefreshMetaData.mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(createMockDashboardData()), 100))
      );

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const refreshButton = screen.getByRole('button', { name: /refresh/i });
      fireEvent.click(refreshButton);

      expect(screen.getByText(/refreshing/i)).toBeInTheDocument();
    });

    it('updates data after refresh', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      const updatedData = createMockDashboardData({
        archetypes: [
          Object.assign(new gui.ArchetypeInfo({}), {
            name: 'New Archetype',
            colors: ['W'],
            metaShare: 20.0,
            tournamentTop8s: 15,
            tournamentWins: 5,
            tier: 1,
            confidenceScore: 0.99,
            trendDirection: 'up',
          }),
        ],
      });
      mockRefreshMetaData.mockResolvedValue(updatedData);

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const refreshButton = screen.getByRole('button', { name: /refresh/i });
      fireEvent.click(refreshButton);

      await waitFor(() => {
        expect(screen.getByText('New Archetype')).toBeInTheDocument();
      });
    });

    it('displays error when refresh fails', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());
      mockRefreshMetaData.mockRejectedValue(new Error('Refresh failed'));

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const refreshButton = screen.getByRole('button', { name: /refresh/i });
      fireEvent.click(refreshButton);

      await waitFor(() => {
        expect(screen.getByText(/refresh failed/i)).toBeInTheDocument();
      });
    });
  });

  describe('trend indicators', () => {
    it('renders up trend icon for rising archetypes', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        const upTrends = screen.getAllByTitle('Trending up');
        expect(upTrends.length).toBeGreaterThan(0);
      });
    });

    it('renders down trend icon for falling archetypes', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        const downTrends = screen.getAllByTitle('Trending down');
        expect(downTrends.length).toBeGreaterThan(0);
      });
    });

    it('renders stable trend icon for stable archetypes', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        const stableTrends = screen.getAllByTitle('Stable');
        expect(stableTrends.length).toBeGreaterThan(0);
      });
    });
  });

  describe('color badges', () => {
    it('renders color pips for archetypes', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        // Should have R for Mono Red Aggro
        expect(screen.getByText('R')).toBeInTheDocument();
      });

      // Should have W and U for Azorius Control
      const whitePips = screen.getAllByText('W');
      expect(whitePips.length).toBeGreaterThan(0);

      const bluePips = screen.getAllByText('U');
      expect(bluePips.length).toBeGreaterThan(0);
    });
  });

  describe('accessibility', () => {
    it('has accessible format selector', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        const select = screen.getByRole('combobox');
        expect(select).toBeInTheDocument();
      });
    });

    it('has accessible refresh button', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      const refreshButton = screen.getByRole('button', { name: /refresh/i });
      expect(refreshButton).toBeInTheDocument();
    });

    it('opens external links in new tab', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        const link = screen.getByText('View Details →');
        expect(link).toHaveAttribute('target', '_blank');
        expect(link).toHaveAttribute('rel', 'noopener noreferrer');
      });
    });

    it('archetype cards are accessible with role button and tabIndex', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      // Check for role="button" on archetype cards
      const buttons = screen.getAllByRole('button');
      // Should have at least the archetype cards plus refresh button
      expect(buttons.length).toBeGreaterThan(1);
    });
  });

  describe('archetype detail view', () => {
    it('opens detail panel when clicking on an archetype card', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      // Click on the archetype card
      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      expect(archetypeCard).toBeInTheDocument();
      fireEvent.click(archetypeCard!);

      // Check that the detail panel opens
      await waitFor(() => {
        // The detail header should now show the archetype name in an h2
        expect(screen.getByRole('heading', { level: 2, name: 'Mono Red Aggro' })).toBeInTheDocument();
      });
    });

    it('shows meta share in detail panel', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText('Meta Share')).toBeInTheDocument();
        expect(screen.getByText('15.5%')).toBeInTheDocument();
      });
    });

    it('shows tournament top 8s in detail panel', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText('Tournament Top 8s')).toBeInTheDocument();
        expect(screen.getByText('12')).toBeInTheDocument();
      });
    });

    it('shows tournament wins in detail panel', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText('Tournament Wins')).toBeInTheDocument();
      });
    });

    it('shows data confidence in detail panel', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText('Data Confidence')).toBeInTheDocument();
        expect(screen.getByText('95%')).toBeInTheDocument();
      });
    });

    it('shows trend analysis section', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText('Trend Analysis')).toBeInTheDocument();
        expect(screen.getByText(/trending upward/i)).toBeInTheDocument();
      });
    });

    it('shows tier explanation section', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText('Tier Ranking')).toBeInTheDocument();
        expect(screen.getByText(/Tier 1 decks are the most competitive/i)).toBeInTheDocument();
      });
    });

    it('closes detail panel when clicking close button', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 2, name: 'Mono Red Aggro' })).toBeInTheDocument();
      });

      // Click the close button
      const closeButton = screen.getByText('×');
      fireEvent.click(closeButton);

      // Panel should close - the h2 heading in detail panel should be gone
      await waitFor(() => {
        expect(screen.queryByRole('heading', { level: 2, name: 'Mono Red Aggro' })).not.toBeInTheDocument();
      });
    });

    it('closes detail panel when clicking overlay', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 2, name: 'Mono Red Aggro' })).toBeInTheDocument();
      });

      // Click the overlay (background)
      const overlay = document.querySelector('.archetype-detail-overlay');
      fireEvent.click(overlay!);

      // Panel should close
      await waitFor(() => {
        expect(screen.queryByRole('heading', { level: 2, name: 'Mono Red Aggro' })).not.toBeInTheDocument();
      });
    });

    it('does not close panel when clicking inside panel', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 2, name: 'Mono Red Aggro' })).toBeInTheDocument();
      });

      // Click inside the panel
      const panel = document.querySelector('.archetype-detail-panel');
      fireEvent.click(panel!);

      // Panel should still be open
      expect(screen.getByRole('heading', { level: 2, name: 'Mono Red Aggro' })).toBeInTheDocument();
    });

    it('opens detail panel with keyboard Enter key', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const archetypeCard = screen.getByText('Mono Red Aggro').closest('.archetype-card');
      fireEvent.keyDown(archetypeCard!, { key: 'Enter' });

      await waitFor(() => {
        expect(screen.getByRole('heading', { level: 2, name: 'Mono Red Aggro' })).toBeInTheDocument();
      });
    });

    it('shows different trend message for down trend', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Golgari Midrange')).toBeInTheDocument();
      });

      // Click on Golgari Midrange which has down trend
      const archetypeCard = screen.getByText('Golgari Midrange').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText(/trending downward/i)).toBeInTheDocument();
      });
    });

    it('shows different trend message for stable trend', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Azorius Control')).toBeInTheDocument();
      });

      // Click on Azorius Control which has stable trend
      const archetypeCard = screen.getByText('Azorius Control').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText(/stable/i)).toBeInTheDocument();
      });
    });

    it('shows tier 2 explanation for tier 2 decks', async () => {
      mockGetMetaDashboard.mockResolvedValue(createMockDashboardData());

      renderMeta();

      await waitFor(() => {
        expect(screen.getByText('Golgari Midrange')).toBeInTheDocument();
      });

      // Click on Golgari Midrange which is tier 2
      const archetypeCard = screen.getByText('Golgari Midrange').closest('.archetype-card');
      fireEvent.click(archetypeCard!);

      await waitFor(() => {
        expect(screen.getByText(/Tier 2 decks are strong contenders/i)).toBeInTheDocument();
      });
    });
  });
});

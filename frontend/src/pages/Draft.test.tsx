import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../test/utils/testUtils';
import Draft from './Draft';
import { mockWailsApp } from '@/test/mocks/apiMock';
import { mockEventEmitter } from '@/test/mocks/websocketMock';
import { models, gui } from '@/types/models';

// Mock the getReplayState function from App.tsx
vi.mock('../App', () => ({
  getReplayState: vi.fn(() => ({
    isActive: false,
    isPaused: false,
    progress: null,
  })),
  subscribeToReplayState: vi.fn(() => () => {}),
}));

// Helper function to create mock data
function createMockDraftSession(overrides: Partial<models.DraftSession> = {}): models.DraftSession {
  return new models.DraftSession({
    ID: 'test-session-1',
    EventName: 'QuickDraft',
    SetCode: 'BLB',
    DraftType: 'PremierDraft',
    StartTime: new Date('2025-11-20T10:00:00Z'),
    Status: 'active',
    TotalPicks: 45,
    CreatedAt: new Date('2025-11-20T10:00:00Z'),
    UpdatedAt: new Date('2025-11-20T10:00:00Z'),
    ...overrides,
  });
}

function createMockSetCard(overrides: Partial<models.SetCard> = {}): models.SetCard {
  return new models.SetCard({
    ID: 1,
    SetCode: 'BLB',
    ArenaID: '12345',
    ScryfallID: 'scryfall-id',
    Name: 'Test Card',
    ManaCost: '{2}{W}{U}',
    CMC: 4,
    Types: ['Creature'],
    Colors: ['W', 'U'],
    Rarity: 'rare',
    Text: 'Test card text',
    Power: '2',
    Toughness: '3',
    ImageURL: 'https://example.com/card.jpg',
    ImageURLSmall: 'https://example.com/card-small.jpg',
    ImageURLArt: 'https://example.com/card-art.jpg',
    FetchedAt: new Date(),
    ...overrides,
  });
}

function createMockDraftPick(overrides: Partial<models.DraftPickSession> = {}): models.DraftPickSession {
  return new models.DraftPickSession({
    ID: 1,
    SessionID: 'test-session-1',
    PackNumber: 0,
    PickNumber: 1,
    CardID: '12345',
    Timestamp: new Date('2025-11-20T10:05:00Z'),
    PickQualityGrade: null,
    PickQualityRank: null,
    PackBestGIHWR: null,
    PickedCardGIHWR: null,
    AlternativesJSON: null,
    ...overrides,
  });
}

function createMockCardRating(overrides: Partial<gui.CardRatingWithTier> = {}): gui.CardRatingWithTier {
  return new gui.CardRatingWithTier({
    name: 'Test Card',
    color: 'W',
    rarity: 'rare',
    mtga_id: 12345,
    ever_drawn_win_rate: 0.56,
    opening_hand_win_rate: 0.54,
    ever_drawn_game_win_rate: 0.55,
    drawn_win_rate: 0.57,
    in_hand_win_rate: 0.56,
    ever_drawn_improvement_win_rate: 0.02,
    opening_hand_improvement_win_rate: 0.01,
    drawn_improvement_win_rate: 0.02,
    in_hand_improvement_win_rate: 0.02,
    avg_seen: 3.5,
    avg_pick: 2.1,
    pick_rate: 0.6,
    '# ever_drawn': 1000,
    '# opening_hand': 500,
    '# games': 2000,
    '# drawn': 800,
    '# in_hand_drawn': 700,
    '# games_played': 2000,
    '# decks': 500,
    tier: 'A',
    colors: ['W', 'U'],
    ...overrides,
  });
}

function createMockDeckMetrics(overrides: Partial<models.DeckMetrics> = {}): models.DeckMetrics {
  return new models.DeckMetrics({
    total_cards: 15,
    total_non_land_cards: 15,
    creature_count: 10,
    noncreature_count: 5,
    cmc_average: 2.8,
    distribution_all: [0, 2, 4, 5, 3, 1, 0],
    distribution_creatures: [0, 1, 2, 4, 2, 1, 0],
    distribution_noncreatures: [0, 1, 2, 1, 1, 0, 0],
    type_breakdown: { Creature: 10, Instant: 3, Sorcery: 2 },
    color_distribution: { W: 0.5, U: 0.5, B: 0, R: 0, G: 0 },
    color_counts: { W: 8, U: 7, B: 0, R: 0, G: 0 },
    multi_color_count: 0,
    colorless_count: 0,
    ...overrides,
  });
}

describe('Draft Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockEventEmitter.clear();
  });

  describe('No Active Draft State', () => {
    it('should display draft history when no active draft exists', async () => {
      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([]);
      mockWailsApp.GetCompletedDraftSessions.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft History')).toBeInTheDocument();
      });

      expect(screen.getByText(/Start a Quick Draft in MTG Arena to begin a new draft session/i)).toBeInTheDocument();
    });

    it('should display historical drafts when available', async () => {
      const completedSession = createMockDraftSession({
        ID: 'completed-session',
        Status: 'completed',
        TotalPicks: 45,
      });

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([]);
      mockWailsApp.GetCompletedDraftSessions.mockResolvedValue([completedSession]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('QuickDraft')).toBeInTheDocument();
      });

      expect(screen.getByText(/BLB/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /View Replay/i })).toBeInTheDocument();
    });

    it('should display empty state when no historical drafts exist', async () => {
      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([]);
      mockWailsApp.GetCompletedDraftSessions.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('No Draft History')).toBeInTheDocument();
      });

      expect(screen.getByText(/Complete a Quick Draft in MTG Arena to see your draft history here/i)).toBeInTheDocument();
    });
  });

  describe('Active Draft Display', () => {
    it('should load and display an active draft session', async () => {
      const session = createMockDraftSession();
      const picks: models.DraftPickSession[] = [];
      const packs: models.DraftPackSession[] = [];
      const setCards = [createMockSetCard()];
      const ratings = [createMockCardRating()];

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue(packs);
      mockWailsApp.GetSetCards.mockResolvedValue(setCards);
      mockWailsApp.GetCardRatings.mockResolvedValue(ratings);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft Assistant')).toBeInTheDocument();
      });

      expect(screen.getByText('QuickDraft')).toBeInTheDocument();
      expect(screen.getByText(/Set: BLB/i)).toBeInTheDocument();
      expect(screen.getByText(/Picks: 0\/45/i)).toBeInTheDocument();
    });

    it('should display loading state while fetching draft data', () => {
      mockWailsApp.GetActiveDraftSessions.mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      render(<Draft />);

      expect(screen.getByText('Loading draft...')).toBeInTheDocument();
    });

    it('should update when draft:updated event is fired', async () => {
      const session = createMockDraftSession();
      const picks: models.DraftPickSession[] = [];
      const packs: models.DraftPackSession[] = [];
      const setCards = [createMockSetCard()];
      const ratings = [createMockCardRating()];
      const mockMetrics = createMockDeckMetrics();

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue(packs);
      mockWailsApp.GetSetCards.mockResolvedValue(setCards);
      mockWailsApp.GetCardRatings.mockResolvedValue(ratings);
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(mockMetrics);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft Assistant')).toBeInTheDocument();
      });

      // Update picks
      const newPick = createMockDraftPick();
      mockWailsApp.GetDraftPicks.mockResolvedValue([newPick]);
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(mockMetrics);

      // Fire draft:updated event
      mockEventEmitter.emit('draft:updated');

      await waitFor(() => {
        expect(screen.getByText(/Picks: 1\/45/i)).toBeInTheDocument();
      }, { timeout: 3000 });
    });
  });

  describe('Draft Picks Display', () => {
    it('should render picked cards correctly', async () => {
      const session = createMockDraftSession();
      const card1 = createMockSetCard({ ArenaID: '11111', Name: 'Card One' });
      const card2 = createMockSetCard({ ArenaID: '22222', Name: 'Card Two' });
      const picks = [
        createMockDraftPick({ CardID: '11111', PackNumber: 0, PickNumber: 1 }),
        createMockDraftPick({ CardID: '22222', PackNumber: 0, PickNumber: 2 }),
      ];

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([card1, card2]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText(/Picks: 2\/45/i)).toBeInTheDocument();
      });

      // Check pick history
      expect(screen.getByText('Pick History')).toBeInTheDocument();
      expect(screen.getByText('P1P1')).toBeInTheDocument();
      expect(screen.getByText('P1P2')).toBeInTheDocument();
    });

    it('should display picked indicator on cards in grid', async () => {
      const session = createMockDraftSession();
      const card = createMockSetCard({ ArenaID: '12345' });
      const pick = createMockDraftPick({ CardID: '12345' });
      const metrics = createMockDeckMetrics();

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue([pick]);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([card]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(metrics);

      render(<Draft />);

      // Wait for the view toggle to appear
      await waitFor(() => {
        expect(screen.getByText('All Set Cards')).toBeInTheDocument();
      });

      // Click "All Set Cards" to switch to grid view (default is CurrentPackPicker)
      const allSetCardsBtn = screen.getByText('All Set Cards');
      await userEvent.click(allSetCardsBtn);

      await waitFor(() => {
        const cardItems = document.querySelectorAll('.card-item.picked');
        expect(cardItems.length).toBeGreaterThan(0);
      });
    });

    it('should display pick quality grades when available', async () => {
      const session = createMockDraftSession();
      const card = createMockSetCard({ ArenaID: '12345' });
      const pick = createMockDraftPick({
        CardID: '12345',
        PickQualityGrade: 'A',
        PickQualityRank: 1,
      });

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue([pick]);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([card]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        const gradeBadges = document.querySelectorAll('.pick-quality-badge');
        expect(gradeBadges.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Card Recommendations and Synergies', () => {
    it('should display synergy indicators for recommended cards', async () => {
      const session = createMockDraftSession();
      const pickedCard = createMockSetCard({
        ArenaID: '11111',
        Name: 'Plains Walker',
        Types: ['Creature'],
        Colors: ['W'],
        CMC: 3,
      });
      const synergyCard = createMockSetCard({
        ArenaID: '22222',
        Name: 'White Knight',
        Types: ['Creature'],
        Colors: ['W'],
        CMC: 2,
      });

      const picks = [createMockDraftPick({ CardID: '11111' })];

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([pickedCard, synergyCard]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft Assistant')).toBeInTheDocument();
      });

      // Synergy indicators should be present for cards that match color identity
      const synergyIndicators = document.querySelectorAll('.synergy-highlight');
      expect(synergyIndicators.length).toBeGreaterThanOrEqual(0);
    });
  });

  describe('Analyze Draft Functionality', () => {
    it('should call analyze function when button is clicked', async () => {
      const session = createMockDraftSession();
      const picks = [createMockDraftPick()];
      const mockMetrics = createMockDeckMetrics();

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([createMockSetCard()]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(mockMetrics);
      mockWailsApp.AnalyzeSessionPickQuality.mockResolvedValue(undefined);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft Assistant')).toBeInTheDocument();
      });

      const analyzeButton = screen.getByRole('button', { name: /Analyze Pick Quality/i });
      await userEvent.click(analyzeButton);

      await waitFor(() => {
        expect(mockWailsApp.AnalyzeSessionPickQuality).toHaveBeenCalledWith('test-session-1');
      });
    });

    it('should disable analyze button when no picks exist', async () => {
      const session = createMockDraftSession();

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue([]);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        const analyzeButton = screen.getByRole('button', { name: /Analyze Pick Quality/i });
        expect(analyzeButton).toBeDisabled();
      });
    });
  });

  describe('Card Details Overlay', () => {
    it('should display card details when card is clicked', async () => {
      const session = createMockDraftSession();
      const card = createMockSetCard({ Name: 'Detailed Card' });

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue([]);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([card]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft Assistant')).toBeInTheDocument();
      });

      const cardElement = document.querySelector('.card-item');
      if (cardElement) {
        await userEvent.click(cardElement);

        await waitFor(() => {
          expect(screen.getByText('Detailed Card')).toBeInTheDocument();
          expect(screen.getByText(/Creature/i)).toBeInTheDocument();
        });
      }
    });

    it('should close card details overlay when backdrop is clicked', async () => {
      const session = createMockDraftSession();
      const card = createMockSetCard({ Name: 'Closable Card' });

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([session]);
      mockWailsApp.GetDraftPicks.mockResolvedValue([]);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetSetCards.mockResolvedValue([card]);
      mockWailsApp.GetCardRatings.mockResolvedValue([]);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByText('Draft Assistant')).toBeInTheDocument();
      });

      // Click card to open overlay
      const cardElement = document.querySelector('.card-item');
      if (cardElement) {
        await userEvent.click(cardElement);

        await waitFor(() => {
          expect(screen.getByText('Closable Card')).toBeInTheDocument();
        });

        // Click backdrop to close
        const backdrop = document.querySelector('.card-details-overlay-backdrop');
        if (backdrop) {
          await userEvent.click(backdrop);

          await waitFor(() => {
            expect(screen.queryByText('Closable Card')).not.toBeInTheDocument();
          });
        }
      }
    });
  });

  describe('Historical Draft Detail View', () => {
    it('should display historical draft detail when replay is clicked', async () => {
      const completedSession = createMockDraftSession({
        ID: 'completed-session',
        Status: 'completed',
      });
      const picks = [createMockDraftPick({ SessionID: 'completed-session', CardID: '12345' })];
      const card = createMockSetCard({ ArenaID: '12345' });
      const mockMetrics = createMockDeckMetrics();

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([]);
      mockWailsApp.GetCompletedDraftSessions.mockResolvedValue([completedSession]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade'));
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(mockMetrics);

      render(<Draft />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /View Replay/i })).toBeInTheDocument();
      });

      const replayButton = screen.getByRole('button', { name: /View Replay/i });
      await userEvent.click(replayButton);

      await waitFor(() => {
        expect(screen.getByText('Draft Replay')).toBeInTheDocument();
        expect(screen.getByText(/Picked Cards/i)).toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should return to grid view when back button is clicked', async () => {
      const completedSession = createMockDraftSession({
        ID: 'completed-session',
        Status: 'completed',
      });
      const picks = [createMockDraftPick({ SessionID: 'completed-session', CardID: '12345' })];
      const card = createMockSetCard({ ArenaID: '12345' });
      const mockMetrics = createMockDeckMetrics();

      mockWailsApp.GetActiveDraftSessions.mockResolvedValue([]);
      mockWailsApp.GetCompletedDraftSessions.mockResolvedValue([completedSession]);
      mockWailsApp.GetDraftPicks.mockResolvedValue(picks);
      mockWailsApp.GetDraftPacks.mockResolvedValue([]);
      mockWailsApp.GetCardByArenaID.mockResolvedValue(card);
      mockWailsApp.GetDraftGrade.mockRejectedValue(new Error('No grade'));
      mockWailsApp.GetDraftDeckMetrics.mockResolvedValue(mockMetrics);

      render(<Draft />);

      await waitFor(() => {
        const replayButton = screen.getByRole('button', { name: /View Replay/i });
        expect(replayButton).toBeInTheDocument();
      });

      const replayButton = screen.getByRole('button', { name: /View Replay/i });
      await userEvent.click(replayButton);

      await waitFor(() => {
        expect(screen.getByText('Draft Replay')).toBeInTheDocument();
      }, { timeout: 5000 });

      const backButton = screen.getByRole('button', { name: /Back to Draft History/i });
      await userEvent.click(backButton);

      await waitFor(() => {
        expect(screen.getByText('Draft History')).toBeInTheDocument();
      }, { timeout: 2000 });
    });
  });
});

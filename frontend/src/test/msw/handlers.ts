/**
 * MSW handlers for integration testing.
 * These handlers return realistic API responses matching the actual backend.
 */
import { http, HttpResponse } from 'msw';

const API_BASE = 'http://localhost:8080/api/v1';

/**
 * Create a standard API success response wrapper.
 * The backend wraps all responses in { data: ... }
 */
function successResponse<T>(data: T) {
  return HttpResponse.json({ data });
}

/**
 * Create mock collection cards matching backend CollectionCard struct.
 */
export function createMockCollectionCard(overrides: Partial<{
  cardId: number;
  arenaId: number;
  quantity: number;
  name: string;
  setCode: string;
  setName: string;
  rarity: string;
  manaCost: string;
  cmc: number;
  typeLine: string;
  colors: string[];
  colorIdentity: string[];
  imageUri: string;
  power: string;
  toughness: string;
}> = {}) {
  return {
    cardId: 12345,
    arenaId: 12345,
    quantity: 4,
    name: 'Lightning Bolt',
    setCode: 'sta',
    setName: 'Strixhaven Mystical Archive',
    rarity: 'rare',
    manaCost: '{R}',
    cmc: 1,
    typeLine: 'Instant',
    colors: ['R'],
    colorIdentity: ['R'],
    imageUri: 'https://example.com/card.jpg',
    power: '',
    toughness: '',
    ...overrides,
  };
}

/**
 * Create mock Standard set matching backend StandardSet struct.
 */
export function createMockStandardSet(overrides: Partial<{
  code: string;
  name: string;
  releasedAt: string;
  rotationDate?: string;
  isStandardLegal: boolean;
  iconSvgUri: string;
  cardCount: number;
  daysUntilRotation?: number;
  isRotatingSoon: boolean;
}> = {}) {
  return {
    code: 'dsk',
    name: 'Duskmourn',
    releasedAt: '2024-09-27',
    isStandardLegal: true,
    iconSvgUri: 'https://example.com/set.svg',
    cardCount: 291,
    daysUntilRotation: 365,
    isRotatingSoon: false,
    ...overrides,
  };
}

/**
 * Create mock set info matching backend SetInfo struct.
 */
export function createMockSetInfo(overrides: Partial<{
  code: string;
  name: string;
  iconSvgUri: string;
  setType: string;
  releasedAt: string;
  cardCount: number;
}> = {}) {
  return {
    code: 'sta',
    name: 'Strixhaven Mystical Archive',
    iconSvgUri: 'https://example.com/set.svg',
    setType: 'expansion',
    releasedAt: '2021-04-23',
    cardCount: 63,
    ...overrides,
  };
}

/**
 * Default handlers for common API endpoints.
 * These return realistic response structures matching the actual backend.
 */
export const handlers = [
  // Collection endpoint - returns CollectionResponse with cards array
  http.post(`${API_BASE}/collection`, () => {
    return successResponse({
      cards: [
        createMockCollectionCard({ cardId: 1, name: 'Lightning Bolt' }),
        createMockCollectionCard({ cardId: 2, name: 'Counterspell', colors: ['U'] }),
        createMockCollectionCard({ cardId: 3, name: 'Giant Growth', colors: ['G'] }),
      ],
      totalCount: 3,
      filterCount: 3,
      unknownCardsRemaining: 0,
      unknownCardsFetched: 0,
    });
  }),

  // Collection stats endpoint
  http.get(`${API_BASE}/collection/stats`, () => {
    return successResponse({
      totalUniqueCards: 100,
      totalCards: 400,
      commonCount: 200,
      uncommonCount: 100,
      rareCount: 75,
      mythicCount: 25,
    });
  }),

  // Cards/sets endpoint
  http.get(`${API_BASE}/cards/sets`, () => {
    return successResponse([
      createMockSetInfo({ code: 'sta', name: 'Strixhaven Mystical Archive' }),
      createMockSetInfo({ code: 'dsk', name: 'Duskmourn' }),
    ]);
  }),

  // Set completion endpoint - uses PascalCase to match Go struct serialization
  http.get(`${API_BASE}/collection/sets`, () => {
    return successResponse([
      { SetCode: 'sta', SetName: 'Strixhaven', TotalCards: 63, OwnedCards: 50, Percentage: 79.4 },
      { SetCode: 'dsk', SetName: 'Duskmourn', TotalCards: 200, OwnedCards: 100, Percentage: 50.0 },
    ]);
  }),

  // Matches endpoint
  http.post(`${API_BASE}/matches`, () => {
    return successResponse([]);
  }),

  // Match stats endpoint
  http.post(`${API_BASE}/matches/stats`, () => {
    return successResponse({
      totalMatches: 0,
      wins: 0,
      losses: 0,
      winRate: 0,
    });
  }),

  // System status endpoint
  http.get(`${API_BASE}/system/status`, () => {
    return successResponse({
      status: 'standalone',
      connected: false,
      mode: 'standalone',
      url: 'ws://localhost:9999',
      port: 9999,
    });
  }),

  // Standard sets endpoint
  http.get(`${API_BASE}/standard/sets`, () => {
    return successResponse([
      createMockStandardSet({ code: 'dsk', name: 'Duskmourn' }),
      createMockStandardSet({ code: 'fdn', name: 'Foundations', daysUntilRotation: undefined }),
    ]);
  }),

  // Standard rotation endpoint
  http.get(`${API_BASE}/standard/rotation`, () => {
    return successResponse({
      nextRotationDate: '2027-01-01',
      daysUntilRotation: 365,
      rotatingSets: [
        createMockStandardSet({ code: 'mkm', name: 'Murders at Karlov Manor', isRotatingSoon: true }),
      ],
      rotatingCardCount: 286,
      affectedDecks: 3,
    });
  }),

  // Standard rotation affected decks endpoint
  http.get(`${API_BASE}/standard/rotation/affected-decks`, () => {
    return successResponse([
      {
        deckId: 'deck-1',
        deckName: 'Mono Red Aggro',
        format: 'Standard',
        rotatingCardCount: 12,
        totalCards: 60,
        percentAffected: 20,
        rotatingCards: [],
      },
    ]);
  }),

  // Standard config endpoint
  http.get(`${API_BASE}/standard/config`, () => {
    return successResponse({
      id: 1,
      nextRotationDate: '2027-01-01',
      rotationEnabled: true,
      updatedAt: '2024-01-01T00:00:00Z',
    });
  }),

  // Standard validate deck endpoint
  http.post(`${API_BASE}/standard/validate/:deckId`, () => {
    return successResponse({
      isLegal: true,
      errors: [],
      warnings: [],
      rotatingCards: [],
      setBreakdown: [],
    });
  }),

  // Standard card legality endpoint
  http.get(`${API_BASE}/standard/cards/:arenaId/legality`, () => {
    return successResponse({
      standard: 'legal',
      historic: 'legal',
      explorer: 'legal',
      pioneer: 'legal',
      modern: 'legal',
      alchemy: 'legal',
      brawl: 'legal',
      commander: 'legal',
    });
  }),

  // Build around seed endpoint
  http.post(`${API_BASE}/decks/build-around`, () => {
    return successResponse({
      seedCard: {
        cardID: 12345,
        name: 'Test Seed Card',
        manaCost: '{2}{W}',
        cmc: 3,
        colors: ['W'],
        typeLine: 'Creature - Human',
        rarity: 'rare',
        imageURI: 'https://example.com/card.jpg',
        score: 1.0,
        reasoning: 'This is your build-around card.',
        inCollection: true,
        ownedCount: 4,
        neededCount: 0,
      },
      suggestions: [
        {
          cardID: 11111,
          name: 'Suggested Card 1',
          manaCost: '{1}{W}',
          cmc: 2,
          colors: ['W'],
          typeLine: 'Creature - Soldier',
          rarity: 'uncommon',
          score: 0.85,
          reasoning: 'Synergizes with your strategy',
          inCollection: true,
          ownedCount: 3,
          neededCount: 1,
        },
        {
          cardID: 22222,
          name: 'Suggested Card 2',
          manaCost: '{2}{W}',
          cmc: 3,
          colors: ['W'],
          typeLine: 'Instant',
          rarity: 'rare',
          score: 0.78,
          reasoning: 'Good curve fit',
          inCollection: false,
          ownedCount: 0,
          neededCount: 4,
        },
      ],
      lands: [
        { cardID: 81716, name: 'Plains', quantity: 24, color: 'W' },
      ],
      analysis: {
        colorIdentity: ['W'],
        keywords: ['lifelink', 'vigilance'],
        themes: ['tokens'],
        idealCurve: { 1: 4, 2: 8, 3: 8, 4: 6, 5: 4, 6: 2 },
        suggestedLandCount: 24,
        totalCards: 60,
        inCollectionCount: 40,
        missingCount: 16,
        missingWildcardCost: { rare: 8, uncommon: 4, common: 4 },
      },
    });
  }),

  // Generate complete deck endpoint (Issue #774)
  http.post(`${API_BASE}/decks/generate`, () => {
    return successResponse({
      seedCard: {
        cardID: 12345,
        name: 'Test Seed Card',
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
          reasoning: 'Fast creature for aggro',
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
        gamePlan: 'Deploy cheap threats early and finish with burn spells.',
        keyCards: ['Aggressive Creature', 'Burn Spell'],
        mulligan: 'Keep hands with 2-3 lands and cheap creatures.',
        strengths: ['Fast', 'Consistent'],
        weaknesses: ['Weak to lifegain', 'Runs out of gas late'],
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
    });
  }),

  // Get archetype profiles endpoint (Issue #774)
  http.get(`${API_BASE}/decks/archetypes`, () => {
    return successResponse({
      aggro: {
        name: 'Aggro',
        landCount: 20,
        curveTargets: { 1: 8, 2: 14, 3: 10, 4: 4, 5: 4, 6: 0 },
        description: 'Fast, aggressive deck that aims to win quickly with cheap threats.',
      },
      midrange: {
        name: 'Midrange',
        landCount: 24,
        curveTargets: { 1: 4, 2: 8, 3: 10, 4: 8, 5: 4, 6: 2 },
        description: 'Balanced deck with efficient threats and answers.',
      },
      control: {
        name: 'Control',
        landCount: 26,
        curveTargets: { 1: 2, 2: 6, 3: 8, 4: 8, 5: 6, 6: 4 },
        description: 'Slow, controlling deck that grinds out opponents with removal.',
      },
    });
  }),

  // Card search endpoint
  http.get(`${API_BASE}/cards`, ({ request }) => {
    const url = new URL(request.url);
    const query = url.searchParams.get('q') || '';
    if (query.toLowerCase().includes('test')) {
      return successResponse([
        {
          ArenaID: '12345',
          Name: 'Test Card',
          ManaCost: '{2}{W}',
          CMC: 3,
          Types: ['Creature'],
          Colors: ['W'],
          ImageURL: 'https://example.com/test.jpg',
        },
      ]);
    }
    return successResponse([]);
  }),

  // Draft export to 17Lands endpoint
  http.get(`${API_BASE}/drafts/:sessionID/export/17lands`, ({ params }) => {
    const sessionID = params.sessionID as string;
    return successResponse({
      session_id: sessionID,
      file_name: `draft_TLA_2024-01-15_14-30-00.json`,
      export: {
        draft_id: sessionID,
        event_type: 'QuickDraft',
        set_code: 'TLA',
        draft_time: '2024-01-15T14:30:00Z',
        picks: [
          {
            pack_number: 1,
            pick_number: 1,
            pack: [12345, 12346, 12347],
            pick: 12345,
            pick_time: '2024-01-15T14:31:00Z',
          },
          {
            pack_number: 1,
            pick_number: 2,
            pack: [12346, 12348],
            pick: 12346,
            pick_time: '2024-01-15T14:32:00Z',
          },
        ],
        metadata: {
          exported_at: new Date().toISOString(),
          exported_from: 'MTGA-Companion',
          overall_grade: 'B+',
          overall_score: 78,
          predicted_win_rate: 0.55,
        },
      },
    });
  }),

  // Exportable drafts endpoint
  http.get(`${API_BASE}/drafts/exportable`, () => {
    return successResponse([
      {
        ID: 'test-session-123',
        SetCode: 'TLA',
        DraftType: 'QuickDraft',
        EventName: 'QuickDraft_TLA',
        Status: 'completed',
        TotalPicks: 45,
        StartTime: '2024-01-15T14:30:00Z',
      },
      {
        ID: 'test-session-456',
        SetCode: 'DSK',
        DraftType: 'PremierDraft',
        EventName: 'PremierDraft_DSK',
        Status: 'completed',
        TotalPicks: 45,
        StartTime: '2024-01-14T10:00:00Z',
      },
    ]);
  }),

  // Deck export endpoint
  http.post(`${API_BASE}/decks/:deckId/export`, async ({ request }) => {
    // Read format from request body (the API sends { format: 'arena' })
    let format = 'arena';
    try {
      const body = (await request.json()) as { format?: string };
      if (body.format) {
        format = body.format;
      }
    } catch {
      // Use default format if body parsing fails
    }

    const formatExtensions: Record<string, string> = {
      arena: '.txt',
      moxfield: '_moxfield.txt',
      archidekt: '_archidekt.txt',
      mtgo: '.dek',
      mtggoldfish: '.txt',
      plaintext: '.txt',
    };

    return successResponse({
      content: `Deck\n4 Lightning Bolt (STA) 1\n4 Mountain (M21) 269`,
      filename: `Test_Deck${formatExtensions[format] || '.txt'}`,
      error: '',
    });
  }),

  // Decks list endpoint
  http.get(`${API_BASE}/decks`, () => {
    return successResponse([
      {
        id: 'deck-1',
        name: 'Mono Red Aggro',
        format: 'Standard',
        source: 'manual',
        primaryArchetype: 'Aggro',
        modifiedAt: '2025-01-01T00:00:00Z',
        matchesPlayed: 10,
        matchWinRate: 0.6,
        currentStreak: 2,
        averageDuration: 600,
      },
      {
        id: 'deck-2',
        name: 'UW Control',
        format: 'Historic',
        source: 'import',
        primaryArchetype: 'Control',
        modifiedAt: '2025-01-02T00:00:00Z',
        matchesPlayed: 5,
        matchWinRate: 0.4,
        currentStreak: -1,
        averageDuration: 1200,
      },
    ]);
  }),
];

/**
 * Handler that returns null collection (for testing null handling).
 */
export const nullCollectionHandler = http.post(`${API_BASE}/collection`, () => {
  return successResponse(null);
});

/**
 * Handler that returns empty collection response.
 */
export const emptyCollectionHandler = http.post(`${API_BASE}/collection`, () => {
  return successResponse({
    cards: [],
    totalCount: 0,
    filterCount: 0,
  });
});

/**
 * Handler that returns collection API error.
 */
export const errorCollectionHandler = http.post(`${API_BASE}/collection`, () => {
  return HttpResponse.json(
    { error: 'Internal Server Error', message: 'Database error', code: 500 },
    { status: 500 }
  );
});

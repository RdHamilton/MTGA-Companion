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

  // Set completion endpoint
  http.get(`${API_BASE}/collection/sets/completion`, () => {
    return successResponse([
      { setCode: 'sta', setName: 'Strixhaven', owned: 50, total: 63, percentage: 79.4 },
      { setCode: 'dsk', setName: 'Duskmourn', owned: 100, total: 200, percentage: 50.0 },
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

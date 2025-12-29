/**
 * Integration tests for collection API service.
 *
 * These tests use MSW to mock HTTP requests at the network level,
 * testing the actual API transformation logic that unit tests miss
 * when mocking at the module level.
 */
import { describe, it, expect, beforeAll, afterAll, afterEach, vi } from 'vitest';
import { server } from '@/test/msw/server';
import {
  handlers,
  nullCollectionHandler,
  emptyCollectionHandler,
  errorCollectionHandler,
  createMockCollectionCard,
} from '@/test/msw/handlers';
import { http, HttpResponse } from 'msw';

// Unmock the API module so we test the real implementation
vi.unmock('@/services/api');

// Import the actual API functions after unmocking
import { collection } from '@/services/api';

const API_BASE = 'http://localhost:8080/api/v1';

describe('Collection API Integration Tests', () => {
  // Start MSW server before all tests
  beforeAll(() => {
    server.listen({ onUnhandledRequest: 'error' });
  });

  // Reset handlers after each test
  afterEach(() => {
    server.resetHandlers();
  });

  // Close server after all tests
  afterAll(() => {
    server.close();
  });

  describe('getCollection', () => {
    it('should extract cards array from backend CollectionResponse structure', async () => {
      // Use default handler which returns { cards: [...], totalCount, filterCount }
      const cards = await collection.getCollection({});

      expect(Array.isArray(cards)).toBe(true);
      expect(cards.length).toBe(3);
      expect(cards[0].name).toBe('Lightning Bolt');
      expect(cards[1].name).toBe('Counterspell');
      expect(cards[2].name).toBe('Giant Growth');
    });

    it('should handle null response gracefully', async () => {
      server.use(nullCollectionHandler);

      const cards = await collection.getCollection({});

      expect(Array.isArray(cards)).toBe(true);
      expect(cards.length).toBe(0);
    });

    it('should handle empty collection response', async () => {
      server.use(emptyCollectionHandler);

      const cards = await collection.getCollection({});

      expect(Array.isArray(cards)).toBe(true);
      expect(cards.length).toBe(0);
    });

    it('should handle response with missing cards field', async () => {
      // Simulate malformed response without cards field
      server.use(
        http.post(`${API_BASE}/collection`, () => {
          return HttpResponse.json({
            data: {
              totalCount: 5,
              filterCount: 5,
              // Missing 'cards' field
            },
          });
        })
      );

      const cards = await collection.getCollection({});

      expect(Array.isArray(cards)).toBe(true);
      expect(cards.length).toBe(0);
    });

    it('should throw error on API failure', async () => {
      server.use(errorCollectionHandler);

      await expect(collection.getCollection({})).rejects.toThrow();
    });

    it('should pass filter parameters to backend', async () => {
      let capturedBody: unknown = null;

      server.use(
        http.post(`${API_BASE}/collection`, async ({ request }) => {
          capturedBody = await request.json();
          return HttpResponse.json({
            data: {
              cards: [createMockCollectionCard()],
              totalCount: 1,
              filterCount: 1,
            },
          });
        })
      );

      await collection.getCollection({
        set_code: 'sta',
        rarity: 'rare',
        colors: ['R', 'U'],
        owned_only: true,
      });

      expect(capturedBody).toEqual({
        set_code: 'sta',
        rarity: 'rare',
        colors: ['R', 'U'],
        owned_only: true,
      });
    });

    it('should handle large collection response', async () => {
      // Generate 1000 cards
      const manyCards = Array.from({ length: 1000 }, (_, i) =>
        createMockCollectionCard({ cardId: i + 1, name: `Card ${i + 1}` })
      );

      server.use(
        http.post(`${API_BASE}/collection`, () => {
          return HttpResponse.json({
            data: {
              cards: manyCards,
              totalCount: 1000,
              filterCount: 1000,
            },
          });
        })
      );

      const cards = await collection.getCollection({});

      expect(cards.length).toBe(1000);
      expect(cards[0].name).toBe('Card 1');
      expect(cards[999].name).toBe('Card 1000');
    });
  });

  describe('getCollectionStats', () => {
    it('should return collection statistics', async () => {
      const stats = await collection.getCollectionStats();

      expect(stats.totalUniqueCards).toBe(100);
      expect(stats.totalCards).toBe(400);
      expect(stats.commonCount).toBe(200);
      expect(stats.uncommonCount).toBe(100);
      expect(stats.rareCount).toBe(75);
      expect(stats.mythicCount).toBe(25);
    });
  });

  describe('getSetCompletion', () => {
    it('should return set completion data', async () => {
      const completion = await collection.getSetCompletion();

      expect(Array.isArray(completion)).toBe(true);
      expect(completion.length).toBe(2);
      expect(completion[0].setCode).toBe('sta');
      expect(completion[1].setCode).toBe('dsk');
    });
  });
});

describe('API Response Structure Validation', () => {
  beforeAll(() => {
    server.listen({ onUnhandledRequest: 'error' });
  });

  afterEach(() => {
    server.resetHandlers();
  });

  afterAll(() => {
    server.close();
  });

  it('should correctly unwrap { data: ... } response envelope', async () => {
    // This test validates that the apiClient correctly unwraps the response
    server.use(
      http.post(`${API_BASE}/collection`, () => {
        // Backend wraps response in { data: ... }
        return HttpResponse.json({
          data: {
            cards: [createMockCollectionCard({ name: 'Test Card' })],
            totalCount: 1,
            filterCount: 1,
          },
        });
      })
    );

    const cards = await collection.getCollection({});

    // Should have unwrapped and extracted cards
    expect(cards[0].name).toBe('Test Card');
  });

  it('should handle nested response structure correctly', async () => {
    // Ensure we don't double-unwrap or miss the nested structure
    server.use(
      http.post(`${API_BASE}/collection`, () => {
        return HttpResponse.json({
          data: {
            cards: [
              {
                cardId: 1,
                arenaId: 1,
                quantity: 4,
                name: 'Nested Card',
                setCode: 'tst',
                setName: 'Test Set',
                rarity: 'common',
                manaCost: '{1}',
                cmc: 1,
                typeLine: 'Creature',
                colors: [],
                colorIdentity: [],
                imageUri: 'https://example.com/card.jpg',
              },
            ],
            totalCount: 1,
            filterCount: 1,
          },
        });
      })
    );

    const cards = await collection.getCollection({});

    expect(cards.length).toBe(1);
    expect(cards[0].cardId).toBe(1);
    expect(cards[0].name).toBe('Nested Card');
    expect(cards[0].quantity).toBe(4);
  });
});

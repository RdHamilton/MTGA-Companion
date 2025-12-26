import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as cards from '../cards';

// Mock the apiClient
vi.mock('../../apiClient', () => ({
  get: vi.fn(),
  post: vi.fn(),
}));

import { get, post } from '../../apiClient';

describe('cards API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('searchCards', () => {
    it('should call post with search request', async () => {
      const mockCards = [{ name: 'Lightning Bolt' }];
      vi.mocked(post).mockResolvedValue(mockCards);

      const result = await cards.searchCards({ query: 'lightning' });

      expect(post).toHaveBeenCalledWith('/cards/search', { query: 'lightning' });
      expect(result).toEqual(mockCards);
    });

    it('should include all search parameters', async () => {
      vi.mocked(post).mockResolvedValue([]);

      await cards.searchCards({
        query: 'bolt',
        set_code: 'MKM',
        colors: ['R'],
        types: ['Instant'],
        rarity: 'common',
        limit: 20,
      });

      expect(post).toHaveBeenCalledWith('/cards/search', {
        query: 'bolt',
        set_code: 'MKM',
        colors: ['R'],
        types: ['Instant'],
        rarity: 'common',
        limit: 20,
      });
    });
  });

  describe('getCardByArenaId', () => {
    it('should call get with correct path', async () => {
      const mockCard = { name: 'Opt', arena_id: 12345 };
      vi.mocked(get).mockResolvedValue(mockCard);

      const result = await cards.getCardByArenaId(12345);

      expect(get).toHaveBeenCalledWith('/cards/arena/12345');
      expect(result).toEqual(mockCard);
    });
  });

  describe('getAllSetInfo', () => {
    it('should call get with correct path', async () => {
      const mockSets = [{ code: 'MKM', name: 'Murders at Karlov Manor' }];
      vi.mocked(get).mockResolvedValue(mockSets);

      const result = await cards.getAllSetInfo();

      expect(get).toHaveBeenCalledWith('/cards/sets');
      expect(result).toEqual(mockSets);
    });
  });

  describe('getSetCards', () => {
    it('should call get with correct path', async () => {
      const mockCards = [{ name: 'Card 1' }, { name: 'Card 2' }];
      vi.mocked(get).mockResolvedValue(mockCards);

      const result = await cards.getSetCards('MKM');

      expect(get).toHaveBeenCalledWith('/cards/sets/MKM/cards');
      expect(result).toEqual(mockCards);
    });
  });

  describe('getCardRatings', () => {
    it('should call get with correct path', async () => {
      const mockRatings = [{ name: 'Card 1', ever_drawn_win_rate: 0.55 }];
      vi.mocked(get).mockResolvedValue(mockRatings);

      const result = await cards.getCardRatings('MKM', 'PremierDraft');

      expect(get).toHaveBeenCalledWith('/cards/ratings/MKM/PremierDraft');
      expect(result).toEqual(mockRatings);
    });
  });

  describe('getCollectionQuantities', () => {
    it('should call post with arena IDs', async () => {
      const mockQuantities = { 12345: 4, 67890: 2 };
      vi.mocked(post).mockResolvedValue(mockQuantities);

      const result = await cards.getCollectionQuantities([12345, 67890]);

      expect(post).toHaveBeenCalledWith('/cards/collection-quantities', {
        arena_ids: [12345, 67890],
      });
      expect(result).toEqual(mockQuantities);
    });
  });

  describe('getColorRatings', () => {
    it('should call get with correct path', async () => {
      const mockRatings = [{ color: 'W', win_rate: 0.52 }];
      vi.mocked(get).mockResolvedValue(mockRatings);

      const result = await cards.getColorRatings('MKM', 'PremierDraft');

      expect(get).toHaveBeenCalledWith('/cards/color-ratings/MKM/PremierDraft');
      expect(result).toEqual(mockRatings);
    });
  });

  describe('searchCardsWithCollection', () => {
    it('should call post with search parameters', async () => {
      const mockCards = [{ name: 'Card 1', quantity: 4 }];
      vi.mocked(post).mockResolvedValue(mockCards);

      const result = await cards.searchCardsWithCollection('lightning', ['MKM'], 50);

      expect(post).toHaveBeenCalledWith('/cards/search-with-collection', {
        query: 'lightning',
        sets: ['MKM'],
        limit: 50,
      });
      expect(result).toEqual(mockCards);
    });

    it('should handle missing optional parameters', async () => {
      vi.mocked(post).mockResolvedValue([]);

      await cards.searchCardsWithCollection('bolt');

      expect(post).toHaveBeenCalledWith('/cards/search-with-collection', {
        query: 'bolt',
        sets: undefined,
        limit: undefined,
      });
    });
  });
});

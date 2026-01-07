import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as decks from '../decks';

// Mock the apiClient
vi.mock('../../apiClient', () => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
}));

import { get, post, put, del } from '../../apiClient';

describe('decks API', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getDecks', () => {
    it('should call get with correct path', async () => {
      vi.mocked(get).mockResolvedValue([]);

      await decks.getDecks();

      expect(get).toHaveBeenCalledWith('/decks');
    });

    it('should call get with format query parameter', async () => {
      vi.mocked(get).mockResolvedValue([]);

      await decks.getDecks({ format: 'Standard' });

      expect(get).toHaveBeenCalledWith('/decks?format=Standard');
    });

    it('should call get with source query parameter', async () => {
      vi.mocked(get).mockResolvedValue([]);

      await decks.getDecks({ source: 'draft' });

      expect(get).toHaveBeenCalledWith('/decks?source=draft');
    });
  });

  describe('getDeck', () => {
    it('should call get with correct path', async () => {
      vi.mocked(get).mockResolvedValue({ deck: {}, cards: [] });

      await decks.getDeck('deck-123');

      expect(get).toHaveBeenCalledWith('/decks/deck-123');
    });
  });

  describe('createDeck', () => {
    it('should call post with correct path and body', async () => {
      vi.mocked(post).mockResolvedValue({ ID: 'new-deck' });

      await decks.createDeck({
        name: 'Test Deck',
        format: 'Standard',
        source: 'constructed',
      });

      expect(post).toHaveBeenCalledWith('/decks', {
        name: 'Test Deck',
        format: 'Standard',
        source: 'constructed',
      });
    });

    it('should include draft_event_id when provided', async () => {
      vi.mocked(post).mockResolvedValue({ ID: 'new-deck' });

      await decks.createDeck({
        name: 'Draft Deck',
        format: 'Limited',
        source: 'draft',
        draft_event_id: 'draft-event-123',
      });

      expect(post).toHaveBeenCalledWith('/decks', {
        name: 'Draft Deck',
        format: 'Limited',
        source: 'draft',
        draft_event_id: 'draft-event-123',
      });
    });
  });

  describe('updateDeck', () => {
    it('should call put with correct path and body', async () => {
      vi.mocked(put).mockResolvedValue({ deck: {}, cards: [] });

      await decks.updateDeck('deck-123', { name: 'Updated Name' });

      expect(put).toHaveBeenCalledWith('/decks/deck-123', { name: 'Updated Name' });
    });
  });

  describe('deleteDeck', () => {
    it('should call del with correct path', async () => {
      vi.mocked(del).mockResolvedValue(undefined);

      await decks.deleteDeck('deck-123');

      expect(del).toHaveBeenCalledWith('/decks/deck-123');
    });
  });

  describe('addCard', () => {
    it('should map frontend field names to backend field names', async () => {
      vi.mocked(post).mockResolvedValue({ status: 'success' });

      await decks.addCard({
        deck_id: 'deck-123',
        arena_id: 12345,
        quantity: 4,
        zone: 'main',
        is_sideboard: false,
      });

      // Verify the backend receives the correct field names
      expect(post).toHaveBeenCalledWith('/decks/deck-123/cards', {
        card_id: 12345, // NOT arena_id
        quantity: 4,
        board: 'main', // NOT zone
        from_draft: false,
      });
    });

    it('should map arena_id to card_id correctly', async () => {
      vi.mocked(post).mockResolvedValue({ status: 'success' });

      await decks.addCard({
        deck_id: 'deck-456',
        arena_id: 99999,
        quantity: 2,
        zone: 'sideboard',
        is_sideboard: true,
      });

      expect(post).toHaveBeenCalledWith('/decks/deck-456/cards', {
        card_id: 99999,
        quantity: 2,
        board: 'sideboard',
        from_draft: false,
      });
    });

    it('should pass from_draft flag when adding draft picks', async () => {
      vi.mocked(post).mockResolvedValue({ status: 'success' });

      await decks.addCard({
        deck_id: 'draft-deck-123',
        arena_id: 54321,
        quantity: 1,
        zone: 'main',
        is_sideboard: false,
        from_draft: true,
      });

      expect(post).toHaveBeenCalledWith('/decks/draft-deck-123/cards', {
        card_id: 54321,
        quantity: 1,
        board: 'main',
        from_draft: true,
      });
    });

    it('should default from_draft to false when not provided', async () => {
      vi.mocked(post).mockResolvedValue({ status: 'success' });

      await decks.addCard({
        deck_id: 'deck-789',
        arena_id: 11111,
        quantity: 3,
        zone: 'main',
        is_sideboard: false,
      });

      expect(post).toHaveBeenCalledWith('/decks/deck-789/cards', {
        card_id: 11111,
        quantity: 3,
        board: 'main',
        from_draft: false,
      });
    });
  });

  describe('removeCard', () => {
    it('should call del with correct path and query params', async () => {
      vi.mocked(del).mockResolvedValue(undefined);

      await decks.removeCard({
        deck_id: 'deck-123',
        arena_id: 12345,
        zone: 'main',
      });

      expect(del).toHaveBeenCalledWith('/decks/deck-123/cards/12345?zone=main');
    });
  });

  describe('exportDeck', () => {
    it('should call post with correct path and format', async () => {
      vi.mocked(post).mockResolvedValue({ content: 'deck list...' });

      await decks.exportDeck('deck-123', { format: 'mtga' });

      expect(post).toHaveBeenCalledWith('/decks/deck-123/export', { format: 'mtga' });
    });
  });

  describe('importDeck', () => {
    it('should call post with correct path and body', async () => {
      vi.mocked(post).mockResolvedValue({ deck: {}, cards: [] });

      await decks.importDeck({
        content: '4 Lightning Bolt',
        name: 'Burn',
        format: 'Modern',
      });

      expect(post).toHaveBeenCalledWith('/decks/import', {
        content: '4 Lightning Bolt',
        name: 'Burn',
        format: 'Modern',
      });
    });
  });

  describe('suggestDecks', () => {
    it('should call post with correct path and session_id', async () => {
      vi.mocked(post).mockResolvedValue([]);

      await decks.suggestDecks({ session_id: 'draft-session-123' });

      expect(post).toHaveBeenCalledWith('/decks/suggest', { session_id: 'draft-session-123' });
    });
  });

  describe('buildAroundSeed', () => {
    it('should call post with correct path and request', async () => {
      vi.mocked(post).mockResolvedValue({ seedCard: {}, suggestions: [], lands: [] });

      await decks.buildAroundSeed({
        seed_card_id: 12345,
        max_results: 40,
        budget_mode: true,
      });

      expect(post).toHaveBeenCalledWith('/decks/build-around', {
        seed_card_id: 12345,
        max_results: 40,
        budget_mode: true,
      });
    });
  });

  describe('suggestNextCards', () => {
    it('should call post with correct path and request', async () => {
      vi.mocked(post).mockResolvedValue({ suggestions: [], deckAnalysis: {}, slotsRemaining: 36 });

      await decks.suggestNextCards({
        seed_card_id: 12345,
        deck_card_ids: [111, 222, 333],
        max_results: 15,
      });

      expect(post).toHaveBeenCalledWith('/decks/build-around/suggest-next', {
        seed_card_id: 12345,
        deck_card_ids: [111, 222, 333],
        max_results: 15,
      });
    });
  });

  describe('getDeckStatistics', () => {
    it('should call get with correct path', async () => {
      vi.mocked(get).mockResolvedValue({ totalMainboard: 60 });

      await decks.getDeckStatistics('deck-123');

      expect(get).toHaveBeenCalledWith('/decks/deck-123/statistics');
    });
  });

  describe('validateDraftDeck', () => {
    it('should call post and return valid status', async () => {
      vi.mocked(post).mockResolvedValue({ valid: true });

      const result = await decks.validateDraftDeck('deck-123');

      expect(post).toHaveBeenCalledWith('/decks/deck-123/validate');
      expect(result).toBe(true);
    });

    it('should return false when deck is invalid', async () => {
      vi.mocked(post).mockResolvedValue({ valid: false });

      const result = await decks.validateDraftDeck('deck-123');

      expect(result).toBe(false);
    });
  });

  describe('cloneDeck', () => {
    it('should call post with correct path and name', async () => {
      vi.mocked(post).mockResolvedValue({ ID: 'cloned-deck' });

      await decks.cloneDeck('deck-123', 'Cloned Deck');

      expect(post).toHaveBeenCalledWith('/decks/deck-123/clone', { name: 'Cloned Deck' });
    });
  });

  describe('addTag', () => {
    it('should call post with correct path and tag', async () => {
      vi.mocked(post).mockResolvedValue(undefined);

      await decks.addTag('deck-123', 'aggro');

      expect(post).toHaveBeenCalledWith('/decks/deck-123/tags', { tag: 'aggro' });
    });
  });

  describe('removeTag', () => {
    it('should call del with correct path', async () => {
      vi.mocked(del).mockResolvedValue(undefined);

      await decks.removeTag('deck-123', 'aggro');

      expect(del).toHaveBeenCalledWith('/decks/deck-123/tags/aggro');
    });

    it('should encode special characters in tag', async () => {
      vi.mocked(del).mockResolvedValue(undefined);

      await decks.removeTag('deck-123', 'multi word tag');

      expect(del).toHaveBeenCalledWith('/decks/deck-123/tags/multi%20word%20tag');
    });
  });
});

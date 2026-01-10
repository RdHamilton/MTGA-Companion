import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDeckUndoRedo } from './useDeckUndoRedo';
import { decks } from '@/services/api';
import { type DeckWithCards } from '@/services/api/decks';
import type { models } from '@/types/models';

// Mock the API
vi.mock('@/services/api', () => ({
  decks: {
    addCard: vi.fn(),
    removeCard: vi.fn(),
    getDeck: vi.fn(),
  },
}));

const mockDecks = vi.mocked(decks);

describe('useDeckUndoRedo', () => {
  const mockDeckId = 'test-deck-123';

  const createMockCard = (id: number, quantity: number, board: string = 'main'): models.DeckCard => ({
    ID: id,
    DeckID: mockDeckId,
    CardID: id,
    Quantity: quantity,
    Board: board,
    FromDraftPick: false,
  });

  const createMockDeckWithCards = (cards: models.DeckCard[]): DeckWithCards => ({
    deck: {} as models.Deck,
    cards,
    tags: [],
  } as DeckWithCards);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('initialization', () => {
    it('returns initial state with no undo/redo available', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      expect(result.current.canUndo).toBe(false);
      expect(result.current.canRedo).toBe(false);
      expect(result.current.undoCount).toBe(0);
      expect(result.current.redoCount).toBe(0);
    });
  });

  describe('saveSnapshot and updateCurrentState', () => {
    it('enables undo after saving a snapshot', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cards = [createMockCard(1, 4)];

      act(() => {
        result.current.saveSnapshot(cards);
      });

      expect(result.current.canUndo).toBe(true);
      expect(result.current.undoCount).toBe(1);
    });

    it('provides updateCurrentState to track new state', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 4)];
      const cardsAfter = [createMockCard(1, 4), createMockCard(2, 2)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      // Description should show the change
      const desc = result.current.getUndoDescription();
      expect(desc).toContain('Add');
      expect(desc).toContain('Card 2');
    });
  });

  describe('undo', () => {
    it('makes API calls to restore previous state', async () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 4)];
      const cardsAfter = [createMockCard(1, 4), createMockCard(2, 2)];

      // Save state before adding card, then update to new state
      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      // Mock API for undo - should remove card 2
      mockDecks.removeCard.mockResolvedValue(undefined);
      mockDecks.getDeck.mockResolvedValue(createMockDeckWithCards(cardsBefore));

      let undoResult: models.DeckCard[] | null = null;
      await act(async () => {
        undoResult = await result.current.undo();
      });

      // Should have called removeCard to remove the added card
      expect(mockDecks.removeCard).toHaveBeenCalled();
      expect(undoResult).toEqual(cardsBefore);
    });

    it('returns null when no undo available', async () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      let undoResult: models.DeckCard[] | null = null;
      await act(async () => {
        undoResult = await result.current.undo();
      });

      expect(undoResult).toBeNull();
      expect(mockDecks.removeCard).not.toHaveBeenCalled();
      expect(mockDecks.addCard).not.toHaveBeenCalled();
    });

    it('calls onStateRestored callback after undo', async () => {
      const onStateRestored = vi.fn();
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId, onStateRestored })
      );

      const cardsBefore = [createMockCard(1, 4)];
      const cardsAfter = [createMockCard(1, 4), createMockCard(2, 2)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      mockDecks.removeCard.mockResolvedValue(undefined);
      mockDecks.getDeck.mockResolvedValue(createMockDeckWithCards(cardsBefore));

      await act(async () => {
        await result.current.undo();
      });

      expect(onStateRestored).toHaveBeenCalledWith(cardsBefore);
    });
  });

  describe('redo', () => {
    it('makes API calls to restore next state', async () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 4)];
      const cardsAfter = [createMockCard(1, 4), createMockCard(2, 2)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      // Undo first
      mockDecks.removeCard.mockResolvedValue(undefined);
      mockDecks.getDeck.mockResolvedValue(createMockDeckWithCards(cardsBefore));

      await act(async () => {
        await result.current.undo();
      });

      // Now redo - should add back card 2
      mockDecks.addCard.mockResolvedValue(undefined);
      mockDecks.getDeck.mockResolvedValue(createMockDeckWithCards(cardsAfter));

      let redoResult: models.DeckCard[] | null = null;
      await act(async () => {
        redoResult = await result.current.redo();
      });

      expect(mockDecks.addCard).toHaveBeenCalled();
      expect(redoResult).toEqual(cardsAfter);
    });

    it('returns null when no redo available', async () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      let redoResult: models.DeckCard[] | null = null;
      await act(async () => {
        redoResult = await result.current.redo();
      });

      expect(redoResult).toBeNull();
    });
  });

  describe('clear', () => {
    it('clears all undo and redo history', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      act(() => {
        result.current.saveSnapshot([createMockCard(1, 4)]);
        result.current.updateCurrentState([createMockCard(1, 4), createMockCard(2, 4)]);
      });

      expect(result.current.canUndo).toBe(true);

      act(() => {
        result.current.clear();
      });

      expect(result.current.canUndo).toBe(false);
      expect(result.current.canRedo).toBe(false);
    });
  });

  describe('descriptions', () => {
    it('generates undo description for added card', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 4)];
      const cardsAfter = [createMockCard(1, 4), createMockCard(2, 2)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      const desc = result.current.getUndoDescription();
      expect(desc).toContain('Undo');
      expect(desc).toContain('Add');
      expect(desc).toContain('Card 2');
    });

    it('generates undo description for removed card', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 4), createMockCard(2, 2)];
      const cardsAfter = [createMockCard(1, 4)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      const desc = result.current.getUndoDescription();
      expect(desc).toContain('Undo');
      expect(desc).toContain('Remove');
      expect(desc).toContain('Card 2');
    });

    it('generates undo description for quantity change', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 2)];
      const cardsAfter = [createMockCard(1, 4)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      const desc = result.current.getUndoDescription();
      expect(desc).toContain('Undo');
      expect(desc).toContain('+2');
      expect(desc).toContain('Card 1');
    });

    it('returns null description when no undo available', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      expect(result.current.getUndoDescription()).toBeNull();
    });
  });

  describe('maxStackSize', () => {
    it('respects custom maxStackSize option', () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId, maxStackSize: 3 })
      );

      act(() => {
        result.current.saveSnapshot([createMockCard(1, 1)]);
        result.current.saveSnapshot([createMockCard(1, 2)]);
        result.current.saveSnapshot([createMockCard(1, 3)]);
        result.current.saveSnapshot([createMockCard(1, 4)]);
        result.current.saveSnapshot([createMockCard(1, 5)]);
      });

      // Should only keep last 3 snapshots
      expect(result.current.undoCount).toBe(3);
    });
  });

  describe('diff calculation', () => {
    it('correctly calculates diff for adding multiple cards', async () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore: models.DeckCard[] = [];
      const cardsAfter = [createMockCard(1, 4), createMockCard(2, 3)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      mockDecks.removeCard.mockResolvedValue(undefined);
      mockDecks.getDeck.mockResolvedValue(createMockDeckWithCards(cardsBefore));

      await act(async () => {
        await result.current.undo();
      });

      // Should have called removeCard for each copy of each card
      // Card 1: 4 times, Card 2: 3 times = 7 total
      expect(mockDecks.removeCard).toHaveBeenCalledTimes(7);
    });

    it('correctly calculates diff for quantity increase', async () => {
      const { result } = renderHook(() =>
        useDeckUndoRedo({ deckId: mockDeckId })
      );

      const cardsBefore = [createMockCard(1, 2)];
      const cardsAfter = [createMockCard(1, 4)];

      act(() => {
        result.current.saveSnapshot(cardsBefore);
        result.current.updateCurrentState(cardsAfter);
      });

      mockDecks.removeCard.mockResolvedValue(undefined);
      mockDecks.getDeck.mockResolvedValue(createMockDeckWithCards(cardsBefore));

      await act(async () => {
        await result.current.undo();
      });

      // Should remove 2 copies (4 - 2 = 2)
      expect(mockDecks.removeCard).toHaveBeenCalledTimes(2);
    });
  });
});

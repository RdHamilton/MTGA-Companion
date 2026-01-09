import { useCallback, useRef } from 'react';
import { useUndoRedo, useUndoRedoKeyboard } from './useUndoRedo';
import { decks } from '@/services/api';
import type { models } from '@/types/models';

export interface DeckCardState {
  cardID: number;
  quantity: number;
  board: string;
  cardName?: string;
}

export interface DeckUndoRedoOptions {
  deckId: string;
  onStateRestored?: (cards: models.DeckCard[]) => void;
  maxStackSize?: number;
}

export interface DeckUndoRedoReturn {
  /** Whether undo is available */
  canUndo: boolean;
  /** Whether redo is available */
  canRedo: boolean;
  /** Number of undo steps available */
  undoCount: number;
  /** Number of redo steps available */
  redoCount: number;
  /** Save current deck state before making changes */
  saveSnapshot: (cards: models.DeckCard[]) => void;
  /** Update current state reference after changes complete */
  updateCurrentState: (cards: models.DeckCard[]) => void;
  /** Undo last change and sync with backend */
  undo: () => Promise<models.DeckCard[] | null>;
  /** Redo last undone change and sync with backend */
  redo: () => Promise<models.DeckCard[] | null>;
  /** Clear all undo/redo history */
  clear: () => void;
  /** Get description of what will be undone */
  getUndoDescription: () => string | null;
  /** Get description of what will be redone */
  getRedoDescription: () => string | null;
  /** Whether an undo/redo operation is in progress */
  isProcessing: boolean;
}

/**
 * Convert DeckCard[] to a normalized state for comparison
 */
function normalizeCards(cards: models.DeckCard[]): DeckCardState[] {
  return cards.map((card) => ({
    cardID: card.CardID,
    quantity: card.Quantity,
    board: card.Board,
    cardName: card.CardName,
  }));
}

/**
 * Generate a description of what changed between two states
 */
function generateChangeDescription(
  prevState: DeckCardState[],
  nextState: DeckCardState[]
): string {
  const prevMap = new Map(prevState.map((c) => [`${c.cardID}-${c.board}`, c]));
  const nextMap = new Map(nextState.map((c) => [`${c.cardID}-${c.board}`, c]));

  const added: { name: string; quantity: number }[] = [];
  const removed: { name: string; quantity: number }[] = [];
  const changed: { name: string; diff: number }[] = [];

  // Find added or changed cards
  for (const [key, nextCard] of nextMap) {
    const prevCard = prevMap.get(key);
    if (!prevCard) {
      added.push({ name: nextCard.cardName || `Card ${nextCard.cardID}`, quantity: nextCard.quantity });
    } else if (prevCard.quantity !== nextCard.quantity) {
      const diff = nextCard.quantity - prevCard.quantity;
      changed.push({ name: nextCard.cardName || `Card ${nextCard.cardID}`, diff });
    }
  }

  // Find removed cards
  for (const [key, prevCard] of prevMap) {
    if (!nextMap.has(key)) {
      removed.push({ name: prevCard.cardName || `Card ${prevCard.cardID}`, quantity: prevCard.quantity });
    }
  }

  const parts: string[] = [];

  if (added.length > 0) {
    parts.push(`Add ${added.map((c) => `${c.quantity}x ${c.name}`).join(', ')}`);
  }
  if (removed.length > 0) {
    parts.push(`Remove ${removed.map((c) => `${c.quantity}x ${c.name}`).join(', ')}`);
  }
  if (changed.length > 0) {
    parts.push(
      changed
        .map((c) => `${c.diff > 0 ? '+' : ''}${c.diff}x ${c.name}`)
        .join(', ')
    );
  }

  return parts.join('; ') || 'No changes';
}

/**
 * Calculate the operations needed to transform from one state to another
 */
function calculateDiff(
  fromState: DeckCardState[],
  toState: DeckCardState[]
): { adds: DeckCardState[]; removes: DeckCardState[] } {
  const fromMap = new Map(fromState.map((c) => [`${c.cardID}-${c.board}`, c]));
  const toMap = new Map(toState.map((c) => [`${c.cardID}-${c.board}`, c]));

  const adds: DeckCardState[] = [];
  const removes: DeckCardState[] = [];

  // Find cards to add or increase quantity
  for (const [key, toCard] of toMap) {
    const fromCard = fromMap.get(key);
    if (!fromCard) {
      // New card - add all copies
      adds.push(toCard);
    } else if (toCard.quantity > fromCard.quantity) {
      // Increased quantity - add the difference
      adds.push({ ...toCard, quantity: toCard.quantity - fromCard.quantity });
    } else if (toCard.quantity < fromCard.quantity) {
      // Decreased quantity - remove the difference
      removes.push({ ...toCard, quantity: fromCard.quantity - toCard.quantity });
    }
  }

  // Find cards to remove entirely
  for (const [key, fromCard] of fromMap) {
    if (!toMap.has(key)) {
      // Card no longer exists - remove all copies
      removes.push(fromCard);
    }
  }

  return { adds, removes };
}

/**
 * Hook for managing undo/redo in the deck builder
 */
export function useDeckUndoRedo(options: DeckUndoRedoOptions): DeckUndoRedoReturn {
  const { deckId, onStateRestored, maxStackSize = 20 } = options;

  const isProcessingRef = useRef(false);
  const currentCardsRef = useRef<DeckCardState[]>([]);

  const {
    canUndo,
    canRedo,
    undoCount,
    redoCount,
    saveState,
    setCurrentState,
    undo: undoBase,
    redo: redoBase,
    clear,
    getUndoDescription: getUndoDescBase,
    getRedoDescription: getRedoDescBase,
  } = useUndoRedo<DeckCardState[]>(generateChangeDescription, { maxStackSize });

  const saveSnapshot = useCallback((cards: models.DeckCard[]) => {
    const normalized = normalizeCards(cards);
    // Save to undo stack (this is the state BEFORE the change)
    saveState(normalized);
  }, [saveState]);

  // Update current state reference after cards change
  const updateCurrentState = useCallback((cards: models.DeckCard[]) => {
    const normalized = normalizeCards(cards);
    currentCardsRef.current = normalized;
    setCurrentState(normalized);
  }, [setCurrentState]);

  const applyState = useCallback(async (targetState: DeckCardState[]): Promise<models.DeckCard[]> => {
    const currentState = currentCardsRef.current;
    const { adds, removes } = calculateDiff(currentState, targetState);

    // Apply removes first
    for (const card of removes) {
      for (let i = 0; i < card.quantity; i++) {
        await decks.removeCard({
          deck_id: deckId,
          arena_id: card.cardID,
          zone: card.board,
        });
      }
    }

    // Then apply adds
    for (const card of adds) {
      await decks.addCard({
        deck_id: deckId,
        arena_id: card.cardID,
        quantity: card.quantity,
        zone: card.board,
        is_sideboard: card.board === 'sideboard',
      });
    }

    // Reload deck to get fresh state
    const deckData = await decks.getDeck(deckId);
    const newCards = deckData.cards || [];

    // Update current state reference
    currentCardsRef.current = normalizeCards(newCards);

    return newCards;
  }, [deckId]);

  const undo = useCallback(async (): Promise<models.DeckCard[] | null> => {
    if (isProcessingRef.current || !canUndo) return null;

    isProcessingRef.current = true;
    try {
      const previousState = undoBase();
      if (!previousState) return null;

      const newCards = await applyState(previousState);
      onStateRestored?.(newCards);
      return newCards;
    } finally {
      isProcessingRef.current = false;
    }
  }, [canUndo, undoBase, applyState, onStateRestored]);

  const redo = useCallback(async (): Promise<models.DeckCard[] | null> => {
    if (isProcessingRef.current || !canRedo) return null;

    isProcessingRef.current = true;
    try {
      const nextState = redoBase();
      if (!nextState) return null;

      const newCards = await applyState(nextState);
      onStateRestored?.(newCards);
      return newCards;
    } finally {
      isProcessingRef.current = false;
    }
  }, [canRedo, redoBase, applyState, onStateRestored]);

  const getUndoDescription = useCallback((): string | null => {
    const desc = getUndoDescBase();
    return desc ? `Undo: ${desc}` : null;
  }, [getUndoDescBase]);

  const getRedoDescription = useCallback((): string | null => {
    const desc = getRedoDescBase();
    return desc ? `Redo: ${desc}` : null;
  }, [getRedoDescBase]);

  return {
    canUndo,
    canRedo,
    undoCount,
    redoCount,
    saveSnapshot,
    updateCurrentState,
    undo,
    redo,
    clear,
    getUndoDescription,
    getRedoDescription,
    isProcessing: isProcessingRef.current,
  };
}

/**
 * Hook to add keyboard shortcuts for deck undo/redo
 */
export function useDeckUndoRedoKeyboard(
  onUndo: () => Promise<void>,
  onRedo: () => Promise<void>,
  enabled: boolean = true
): void {
  useUndoRedoKeyboard(
    () => { onUndo(); },
    () => { onRedo(); },
    enabled
  );
}

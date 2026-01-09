import { useState, useCallback, useEffect, useRef } from 'react';

export interface UndoRedoOptions {
  maxStackSize?: number;
}

export interface UndoRedoReturn<T> {
  /** Whether undo is available */
  canUndo: boolean;
  /** Whether redo is available */
  canRedo: boolean;
  /** Number of undo steps available */
  undoCount: number;
  /** Number of redo steps available */
  redoCount: number;
  /** Undo to previous state, returns the previous state or null if cannot undo */
  undo: () => T | null;
  /** Redo to next state, returns the next state or null if cannot redo */
  redo: () => T | null;
  /** Save state to undo stack (call before making changes) */
  saveState: (state: T) => void;
  /** Update the current state reference (call after making changes) */
  setCurrentState: (state: T) => void;
  /** Clear all undo/redo history */
  clear: () => void;
  /** Get description of what will be undone */
  getUndoDescription: () => string | null;
  /** Get description of what will be redone */
  getRedoDescription: () => string | null;
}

const DEFAULT_MAX_STACK_SIZE = 20;

/**
 * Hook for managing undo/redo state
 *
 * Usage:
 * 1. Call saveState(currentState) BEFORE making changes
 * 2. Make changes
 * 3. Call setCurrentState(newState) AFTER changes complete
 *
 * @param descriptionFn - Optional function to generate descriptions for state changes
 * @param options - Configuration options
 * @returns Undo/redo controls
 */
export function useUndoRedo<T>(
  descriptionFn?: (prevState: T, nextState: T) => string,
  options: UndoRedoOptions = {}
): UndoRedoReturn<T> {
  const { maxStackSize = DEFAULT_MAX_STACK_SIZE } = options;

  const [undoStack, setUndoStack] = useState<T[]>([]);
  const [redoStack, setRedoStack] = useState<T[]>([]);

  // Track current state for description generation and redo
  const currentStateRef = useRef<T | null>(null);

  const canUndo = undoStack.length > 0;
  const canRedo = redoStack.length > 0;

  // Save state to undo stack (call BEFORE making changes)
  const saveState = useCallback((state: T) => {
    setUndoStack((prev) => {
      const newStack = [...prev, state];
      // Limit stack size
      if (newStack.length > maxStackSize) {
        return newStack.slice(newStack.length - maxStackSize);
      }
      return newStack;
    });
    // Clear redo stack when new action is performed
    setRedoStack([]);
  }, [maxStackSize]);

  // Update current state reference (call AFTER changes complete)
  const setCurrentState = useCallback((state: T) => {
    currentStateRef.current = state;
  }, []);

  const undo = useCallback((): T | null => {
    if (undoStack.length === 0) return null;

    const previousState = undoStack[undoStack.length - 1];
    // Capture current state BEFORE modifying ref (callback runs asynchronously)
    const currentState = currentStateRef.current;

    setUndoStack((prev) => prev.slice(0, -1));

    // Save current state to redo stack
    if (currentState !== null) {
      setRedoStack((prev) => [...prev, currentState]);
    }

    currentStateRef.current = previousState;
    return previousState;
  }, [undoStack]);

  const redo = useCallback((): T | null => {
    if (redoStack.length === 0) return null;

    const nextState = redoStack[redoStack.length - 1];
    // Capture current state BEFORE modifying ref (callback runs asynchronously)
    const currentState = currentStateRef.current;

    setRedoStack((prev) => prev.slice(0, -1));

    // Save current state to undo stack
    if (currentState !== null) {
      setUndoStack((prev) => [...prev, currentState]);
    }

    currentStateRef.current = nextState;
    return nextState;
  }, [redoStack]);

  const clear = useCallback(() => {
    setUndoStack([]);
    setRedoStack([]);
    currentStateRef.current = null;
  }, []);

  const getUndoDescription = useCallback((): string | null => {
    if (!descriptionFn || undoStack.length === 0 || currentStateRef.current === null) {
      return null;
    }
    const previousState = undoStack[undoStack.length - 1];
    return descriptionFn(previousState, currentStateRef.current);
  }, [descriptionFn, undoStack]);

  const getRedoDescription = useCallback((): string | null => {
    if (!descriptionFn || redoStack.length === 0 || currentStateRef.current === null) {
      return null;
    }
    const nextState = redoStack[redoStack.length - 1];
    return descriptionFn(currentStateRef.current, nextState);
  }, [descriptionFn, redoStack]);

  return {
    canUndo,
    canRedo,
    undoCount: undoStack.length,
    redoCount: redoStack.length,
    undo,
    redo,
    saveState,
    setCurrentState,
    clear,
    getUndoDescription,
    getRedoDescription,
  };
}

/**
 * Hook for keyboard shortcuts (Ctrl+Z, Ctrl+Y, Cmd+Shift+Z)
 */
export function useUndoRedoKeyboard(
  onUndo: () => void,
  onRedo: () => void,
  enabled: boolean = true
): void {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
      const modKey = isMac ? e.metaKey : e.ctrlKey;

      if (modKey && e.key === 'z') {
        e.preventDefault();
        if (e.shiftKey) {
          onRedo();
        } else {
          onUndo();
        }
      }

      if (modKey && e.key === 'y') {
        e.preventDefault();
        onRedo();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onUndo, onRedo, enabled]);
}

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useUndoRedo, useUndoRedoKeyboard } from './useUndoRedo';

describe('useUndoRedo', () => {
  describe('initialization', () => {
    it('returns initial state with no undo/redo available', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      expect(result.current.canUndo).toBe(false);
      expect(result.current.canRedo).toBe(false);
      expect(result.current.undoCount).toBe(0);
      expect(result.current.redoCount).toBe(0);
    });
  });

  describe('saveState and setCurrentState', () => {
    it('enables undo after saving state', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
      });

      expect(result.current.canUndo).toBe(true);
      expect(result.current.undoCount).toBe(1);
    });

    it('clears redo stack when new state is saved', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      // Save initial states
      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      // Undo to create redo stack
      act(() => {
        result.current.undo();
      });

      expect(result.current.canRedo).toBe(true);

      // Save new state - should clear redo stack
      act(() => {
        result.current.saveState('state1');
      });

      expect(result.current.canRedo).toBe(false);
      expect(result.current.redoCount).toBe(0);
    });

    it('respects maxStackSize option', () => {
      const { result } = renderHook(() =>
        useUndoRedo<number>(undefined, { maxStackSize: 3 })
      );

      act(() => {
        result.current.saveState(1);
        result.current.saveState(2);
        result.current.saveState(3);
        result.current.saveState(4);
        result.current.saveState(5);
      });

      // Should only keep last 3 states
      expect(result.current.undoCount).toBe(3);
    });
  });

  describe('undo', () => {
    it('returns previous state when undoing', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      let undoneState: string | null = null;
      act(() => {
        undoneState = result.current.undo();
      });

      expect(undoneState).toBe('state1');
    });

    it('returns null when no undo available', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      let undoneState: string | null = null;
      act(() => {
        undoneState = result.current.undo();
      });

      expect(undoneState).toBeNull();
    });

    it('moves state to redo stack after undo', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      act(() => {
        result.current.undo();
      });

      expect(result.current.canRedo).toBe(true);
      expect(result.current.redoCount).toBe(1);
    });

    it('decrements undo count after undo', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.saveState('state2');
      });

      expect(result.current.undoCount).toBe(2);

      act(() => {
        result.current.undo();
      });

      expect(result.current.undoCount).toBe(1);
    });
  });

  describe('redo', () => {
    it('returns next state when redoing', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      act(() => {
        result.current.undo();
      });

      let redoneState: string | null = null;
      act(() => {
        redoneState = result.current.redo();
      });

      expect(redoneState).toBe('state2');
    });

    it('returns null when no redo available', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      let redoneState: string | null = null;
      act(() => {
        redoneState = result.current.redo();
      });

      expect(redoneState).toBeNull();
    });

    it('moves state back to undo stack after redo', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      act(() => {
        result.current.undo();
      });

      const undoCountBefore = result.current.undoCount;

      act(() => {
        result.current.redo();
      });

      expect(result.current.undoCount).toBe(undoCountBefore + 1);
      expect(result.current.canRedo).toBe(false);
    });
  });

  describe('clear', () => {
    it('clears all undo and redo history', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      act(() => {
        result.current.undo();
      });

      expect(result.current.canUndo).toBe(false); // After undo, undo stack is empty
      expect(result.current.canRedo).toBe(true);

      act(() => {
        result.current.clear();
      });

      expect(result.current.canUndo).toBe(false);
      expect(result.current.canRedo).toBe(false);
      expect(result.current.undoCount).toBe(0);
      expect(result.current.redoCount).toBe(0);
    });
  });

  describe('descriptions', () => {
    it('returns null description when no description function provided', () => {
      const { result } = renderHook(() => useUndoRedo<string>());

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      expect(result.current.getUndoDescription()).toBeNull();
      expect(result.current.getRedoDescription()).toBeNull();
    });

    it('generates description using provided function', () => {
      const descriptionFn = (prev: string, next: string) => `${prev} -> ${next}`;
      const { result } = renderHook(() => useUndoRedo<string>(descriptionFn));

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      // Undo stack has state1, current is state2
      // Description should be "state1 -> state2"
      expect(result.current.getUndoDescription()).toBe('state1 -> state2');
    });

    it('generates redo description after undo', () => {
      const descriptionFn = (prev: string, next: string) => `${prev} -> ${next}`;
      const { result } = renderHook(() => useUndoRedo<string>(descriptionFn));

      act(() => {
        result.current.saveState('state1');
        result.current.setCurrentState('state2');
      });

      act(() => {
        result.current.undo();
      });

      // After undo: current is state1, redo stack has state2
      // Redo description should be "state1 -> state2"
      expect(result.current.getRedoDescription()).toBe('state1 -> state2');
    });
  });

  describe('multiple undo/redo', () => {
    it('supports multiple sequential undos', () => {
      const { result } = renderHook(() => useUndoRedo<number>());

      act(() => {
        result.current.saveState(1);
        result.current.setCurrentState(2);
        result.current.saveState(2);
        result.current.setCurrentState(3);
        result.current.saveState(3);
        result.current.setCurrentState(4);
      });

      let state: number | null;

      act(() => {
        state = result.current.undo();
      });
      expect(state!).toBe(3);

      act(() => {
        state = result.current.undo();
      });
      expect(state!).toBe(2);

      act(() => {
        state = result.current.undo();
      });
      expect(state!).toBe(1);

      expect(result.current.canUndo).toBe(false);
    });

    it('supports alternating undo/redo', () => {
      const { result } = renderHook(() => useUndoRedo<number>());

      act(() => {
        result.current.saveState(1);
        result.current.setCurrentState(2);
        result.current.saveState(2);
        result.current.setCurrentState(3);
      });

      act(() => {
        result.current.undo(); // Go back to 2
      });

      act(() => {
        result.current.undo(); // Go back to 1
      });

      act(() => {
        result.current.redo(); // Go forward to 2
      });

      expect(result.current.undoCount).toBe(1);
      expect(result.current.redoCount).toBe(1);
    });
  });
});

describe('useUndoRedoKeyboard', () => {
  let addEventListenerSpy: ReturnType<typeof vi.spyOn>;
  let removeEventListenerSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    addEventListenerSpy = vi.spyOn(window, 'addEventListener');
    removeEventListenerSpy = vi.spyOn(window, 'removeEventListener');
  });

  afterEach(() => {
    addEventListenerSpy.mockRestore();
    removeEventListenerSpy.mockRestore();
  });

  it('adds keydown event listener when enabled', () => {
    const onUndo = vi.fn();
    const onRedo = vi.fn();

    renderHook(() => useUndoRedoKeyboard(onUndo, onRedo, true));

    expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
  });

  it('removes event listener on unmount', () => {
    const onUndo = vi.fn();
    const onRedo = vi.fn();

    const { unmount } = renderHook(() => useUndoRedoKeyboard(onUndo, onRedo, true));
    unmount();

    expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
  });

  it('does not add listener when disabled', () => {
    const onUndo = vi.fn();
    const onRedo = vi.fn();

    renderHook(() => useUndoRedoKeyboard(onUndo, onRedo, false));

    expect(addEventListenerSpy).not.toHaveBeenCalled();
  });

  it('calls onUndo for Ctrl+Z', () => {
    const onUndo = vi.fn();
    const onRedo = vi.fn();

    renderHook(() => useUndoRedoKeyboard(onUndo, onRedo, true));

    const event = new KeyboardEvent('keydown', {
      key: 'z',
      ctrlKey: true,
      bubbles: true,
    });
    window.dispatchEvent(event);

    expect(onUndo).toHaveBeenCalledTimes(1);
    expect(onRedo).not.toHaveBeenCalled();
  });

  it('calls onRedo for Ctrl+Shift+Z', () => {
    const onUndo = vi.fn();
    const onRedo = vi.fn();

    renderHook(() => useUndoRedoKeyboard(onUndo, onRedo, true));

    const event = new KeyboardEvent('keydown', {
      key: 'z',
      ctrlKey: true,
      shiftKey: true,
      bubbles: true,
    });
    window.dispatchEvent(event);

    expect(onRedo).toHaveBeenCalledTimes(1);
    expect(onUndo).not.toHaveBeenCalled();
  });

  it('calls onRedo for Ctrl+Y', () => {
    const onUndo = vi.fn();
    const onRedo = vi.fn();

    renderHook(() => useUndoRedoKeyboard(onUndo, onRedo, true));

    const event = new KeyboardEvent('keydown', {
      key: 'y',
      ctrlKey: true,
      bubbles: true,
    });
    window.dispatchEvent(event);

    expect(onRedo).toHaveBeenCalledTimes(1);
    expect(onUndo).not.toHaveBeenCalled();
  });
});

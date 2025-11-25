import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSettingsAccordion } from './useSettingsAccordion';

describe('useSettingsAccordion', () => {
  const STORAGE_KEY = 'mtga-companion-settings-expanded';

  const testSections = [
    { id: 'connection', label: 'Connection' },
    { id: 'preferences', label: 'Preferences' },
    { id: 'data', label: 'Data' },
  ];

  beforeEach(() => {
    localStorage.clear();
    window.location.hash = '';
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('initial state', () => {
    it('returns empty expanded sections when no defaults', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );
      expect(result.current.expandedSections.size).toBe(0);
    });

    it('respects defaultExpanded prop', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          defaultExpanded: ['connection'],
        })
      );
      expect(result.current.isExpanded('connection')).toBe(true);
      expect(result.current.isExpanded('preferences')).toBe(false);
    });

    it('reads from localStorage on mount', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(['preferences', 'data']));

      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      expect(result.current.isExpanded('preferences')).toBe(true);
      expect(result.current.isExpanded('data')).toBe(true);
      expect(result.current.isExpanded('connection')).toBe(false);
    });

    it('reads from URL hash on mount', () => {
      window.location.hash = '#preferences';

      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      expect(result.current.isExpanded('preferences')).toBe(true);
      expect(result.current.expandedSections.size).toBe(1);
    });

    it('URL hash takes precedence over localStorage', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(['connection', 'data']));
      window.location.hash = '#preferences';

      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      expect(result.current.isExpanded('preferences')).toBe(true);
      expect(result.current.expandedSections.size).toBe(1);
    });

    it('ignores invalid section IDs in localStorage', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(['invalid', 'connection']));

      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      expect(result.current.isExpanded('connection')).toBe(true);
      expect(result.current.expandedSections.size).toBe(1);
    });
  });

  describe('toggleSection', () => {
    it('expands a collapsed section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      act(() => {
        result.current.toggleSection('connection');
      });

      expect(result.current.isExpanded('connection')).toBe(true);
    });

    it('collapses an expanded section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          defaultExpanded: ['connection'],
        })
      );

      act(() => {
        result.current.toggleSection('connection');
      });

      expect(result.current.isExpanded('connection')).toBe(false);
    });

    it('allows multiple sections when allowMultiple is true', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          allowMultiple: true,
        })
      );

      act(() => {
        result.current.toggleSection('connection');
      });
      act(() => {
        result.current.toggleSection('preferences');
      });

      expect(result.current.isExpanded('connection')).toBe(true);
      expect(result.current.isExpanded('preferences')).toBe(true);
    });

    it('collapses other sections when allowMultiple is false', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          defaultExpanded: ['connection'],
          allowMultiple: false,
        })
      );

      act(() => {
        result.current.toggleSection('preferences');
      });

      expect(result.current.isExpanded('connection')).toBe(false);
      expect(result.current.isExpanded('preferences')).toBe(true);
    });
  });

  describe('expandSection / collapseSection', () => {
    it('expandSection expands a section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      act(() => {
        result.current.expandSection('connection');
      });

      expect(result.current.isExpanded('connection')).toBe(true);
    });

    it('expandSection does nothing if already expanded', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          defaultExpanded: ['connection'],
        })
      );

      act(() => {
        result.current.expandSection('connection');
      });

      expect(result.current.isExpanded('connection')).toBe(true);
    });

    it('collapseSection collapses a section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          defaultExpanded: ['connection'],
        })
      );

      act(() => {
        result.current.collapseSection('connection');
      });

      expect(result.current.isExpanded('connection')).toBe(false);
    });

    it('collapseSection does nothing if already collapsed', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      act(() => {
        result.current.collapseSection('connection');
      });

      expect(result.current.isExpanded('connection')).toBe(false);
    });
  });

  describe('expandAll / collapseAll', () => {
    it('expandAll expands all sections', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      act(() => {
        result.current.expandAll();
      });

      expect(result.current.isExpanded('connection')).toBe(true);
      expect(result.current.isExpanded('preferences')).toBe(true);
      expect(result.current.isExpanded('data')).toBe(true);
    });

    it('collapseAll collapses all sections', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({
          sections: testSections,
          defaultExpanded: ['connection', 'preferences', 'data'],
        })
      );

      act(() => {
        result.current.collapseAll();
      });

      expect(result.current.isExpanded('connection')).toBe(false);
      expect(result.current.isExpanded('preferences')).toBe(false);
      expect(result.current.isExpanded('data')).toBe(false);
    });
  });

  describe('persistence', () => {
    it('saves to localStorage when sections change', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      act(() => {
        result.current.toggleSection('connection');
      });

      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
      expect(stored).toContain('connection');
    });

    it('updates localStorage when multiple sections change', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections, allowMultiple: true })
      );

      act(() => {
        result.current.toggleSection('connection');
      });
      act(() => {
        result.current.toggleSection('preferences');
      });

      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
      expect(stored).toContain('connection');
      expect(stored).toContain('preferences');
    });
  });

  describe('keyboard navigation', () => {
    it('handles ArrowDown key', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const mockRef = document.createElement('button');
      mockRef.focus = vi.fn();

      act(() => {
        result.current.registerHeaderRef('preferences', mockRef);
      });

      const event = {
        key: 'ArrowDown',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'connection');
      });

      expect(event.preventDefault).toHaveBeenCalled();
      expect(mockRef.focus).toHaveBeenCalled();
    });

    it('handles ArrowUp key', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const mockRef = document.createElement('button');
      mockRef.focus = vi.fn();

      act(() => {
        result.current.registerHeaderRef('connection', mockRef);
      });

      const event = {
        key: 'ArrowUp',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'preferences');
      });

      expect(event.preventDefault).toHaveBeenCalled();
      expect(mockRef.focus).toHaveBeenCalled();
    });

    it('handles Enter key to toggle section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const event = {
        key: 'Enter',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'connection');
      });

      expect(event.preventDefault).toHaveBeenCalled();
      expect(result.current.isExpanded('connection')).toBe(true);
    });

    it('handles Space key to toggle section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const event = {
        key: ' ',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'connection');
      });

      expect(event.preventDefault).toHaveBeenCalled();
      expect(result.current.isExpanded('connection')).toBe(true);
    });

    it('handles Home key to focus first section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const mockRef = document.createElement('button');
      mockRef.focus = vi.fn();

      act(() => {
        result.current.registerHeaderRef('connection', mockRef);
      });

      const event = {
        key: 'Home',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'data');
      });

      expect(event.preventDefault).toHaveBeenCalled();
      expect(mockRef.focus).toHaveBeenCalled();
    });

    it('handles End key to focus last section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const mockRef = document.createElement('button');
      mockRef.focus = vi.fn();

      act(() => {
        result.current.registerHeaderRef('data', mockRef);
      });

      const event = {
        key: 'End',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'connection');
      });

      expect(event.preventDefault).toHaveBeenCalled();
      expect(mockRef.focus).toHaveBeenCalled();
    });

    it('wraps around on ArrowDown from last section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const mockRef = document.createElement('button');
      mockRef.focus = vi.fn();

      act(() => {
        result.current.registerHeaderRef('connection', mockRef);
      });

      const event = {
        key: 'ArrowDown',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'data');
      });

      expect(mockRef.focus).toHaveBeenCalled();
    });

    it('wraps around on ArrowUp from first section', () => {
      const { result } = renderHook(() =>
        useSettingsAccordion({ sections: testSections })
      );

      const mockRef = document.createElement('button');
      mockRef.focus = vi.fn();

      act(() => {
        result.current.registerHeaderRef('data', mockRef);
      });

      const event = {
        key: 'ArrowUp',
        preventDefault: vi.fn(),
      } as unknown as React.KeyboardEvent;

      act(() => {
        result.current.handleKeyDown(event, 'connection');
      });

      expect(mockRef.focus).toHaveBeenCalled();
    });
  });
});

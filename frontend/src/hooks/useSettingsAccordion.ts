import { useState, useEffect, useCallback, useRef } from 'react';

const STORAGE_KEY = 'mtga-companion-settings-expanded';

export interface AccordionSection {
  id: string;
  label: string;
  icon?: string;
}

export interface UseSettingsAccordionOptions {
  sections: AccordionSection[];
  defaultExpanded?: string[];
  allowMultiple?: boolean;
}

export function useSettingsAccordion({
  sections,
  defaultExpanded = [],
  allowMultiple = true,
}: UseSettingsAccordionOptions) {
  // Initialize expanded sections from URL hash, localStorage, or defaults
  const [expandedSections, setExpandedSections] = useState<Set<string>>(() => {
    // Check URL hash first
    const hash = window.location.hash.slice(1);
    if (hash && sections.some((s) => s.id === hash)) {
      return new Set([hash]);
    }

    // Check localStorage
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed) && parsed.length > 0) {
          return new Set(parsed.filter((id: string) => sections.some((s) => s.id === id)));
        }
      }
    } catch {
      // Ignore parse errors
    }

    // Use defaults
    return new Set(defaultExpanded);
  });

  const [focusedIndex, setFocusedIndex] = useState<number>(-1);
  const headerRefs = useRef<Map<string, HTMLButtonElement>>(new Map());

  // Persist to localStorage when expanded sections change
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify([...expandedSections]));
  }, [expandedSections]);

  // Update URL hash when a single section is focused
  useEffect(() => {
    if (expandedSections.size === 1) {
      const sectionId = [...expandedSections][0];
      window.history.replaceState(null, '', `#${sectionId}`);
    } else if (expandedSections.size === 0) {
      window.history.replaceState(null, '', window.location.pathname);
    }
  }, [expandedSections]);

  // Listen for URL hash changes
  useEffect(() => {
    const handleHashChange = () => {
      const hash = window.location.hash.slice(1);
      if (hash && sections.some((s) => s.id === hash)) {
        setExpandedSections(new Set([hash]));
        // Scroll to section
        const element = document.getElementById(`accordion-${hash}`);
        if (element) {
          element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
      }
    };

    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, [sections]);

  // Toggle a section
  const toggleSection = useCallback(
    (sectionId: string) => {
      setExpandedSections((prev) => {
        const next = new Set(prev);
        if (next.has(sectionId)) {
          next.delete(sectionId);
        } else {
          if (!allowMultiple) {
            next.clear();
          }
          next.add(sectionId);
        }
        return next;
      });
    },
    [allowMultiple]
  );

  // Check if a section is expanded
  const isExpanded = useCallback(
    (sectionId: string) => expandedSections.has(sectionId),
    [expandedSections]
  );

  // Expand a specific section
  const expandSection = useCallback(
    (sectionId: string) => {
      setExpandedSections((prev) => {
        if (prev.has(sectionId)) return prev;
        const next = new Set(allowMultiple ? prev : []);
        next.add(sectionId);
        return next;
      });
    },
    [allowMultiple]
  );

  // Collapse a specific section
  const collapseSection = useCallback((sectionId: string) => {
    setExpandedSections((prev) => {
      if (!prev.has(sectionId)) return prev;
      const next = new Set(prev);
      next.delete(sectionId);
      return next;
    });
  }, []);

  // Expand all sections
  const expandAll = useCallback(() => {
    setExpandedSections(new Set(sections.map((s) => s.id)));
  }, [sections]);

  // Collapse all sections
  const collapseAll = useCallback(() => {
    setExpandedSections(new Set());
  }, []);

  // Register header ref for keyboard navigation
  const registerHeaderRef = useCallback((sectionId: string, ref: HTMLButtonElement | null) => {
    if (ref) {
      headerRefs.current.set(sectionId, ref);
    } else {
      headerRefs.current.delete(sectionId);
    }
  }, []);

  // Keyboard navigation handler
  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent, sectionId: string) => {
      const currentIndex = sections.findIndex((s) => s.id === sectionId);
      let nextIndex = currentIndex;

      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault();
          nextIndex = (currentIndex + 1) % sections.length;
          break;
        case 'ArrowUp':
          event.preventDefault();
          nextIndex = (currentIndex - 1 + sections.length) % sections.length;
          break;
        case 'Home':
          event.preventDefault();
          nextIndex = 0;
          break;
        case 'End':
          event.preventDefault();
          nextIndex = sections.length - 1;
          break;
        case 'Enter':
        case ' ':
          event.preventDefault();
          toggleSection(sectionId);
          return;
        default:
          return;
      }

      if (nextIndex !== currentIndex) {
        const nextSection = sections[nextIndex];
        const nextRef = headerRefs.current.get(nextSection.id);
        if (nextRef) {
          nextRef.focus();
          setFocusedIndex(nextIndex);
        }
      }
    },
    [sections, toggleSection]
  );

  return {
    expandedSections,
    isExpanded,
    toggleSection,
    expandSection,
    collapseSection,
    expandAll,
    collapseAll,
    handleKeyDown,
    registerHeaderRef,
    focusedIndex,
  };
}

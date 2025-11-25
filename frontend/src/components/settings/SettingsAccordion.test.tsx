import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SettingsAccordion, SettingsAccordionItem } from './SettingsAccordion';

describe('SettingsAccordion', () => {
  const STORAGE_KEY = 'mtga-companion-settings-expanded';

  const testItems: SettingsAccordionItem[] = [
    { id: 'section1', label: 'Section 1', icon: '📌', content: <div>Content 1</div> },
    { id: 'section2', label: 'Section 2', icon: '⚙️', content: <div>Content 2</div> },
    { id: 'section3', label: 'Section 3', content: <div>Content 3</div> },
  ];

  beforeEach(() => {
    localStorage.clear();
    window.location.hash = '';
  });

  describe('rendering', () => {
    it('renders all accordion items', () => {
      render(<SettingsAccordion items={testItems} />);

      expect(screen.getByText('Section 1')).toBeInTheDocument();
      expect(screen.getByText('Section 2')).toBeInTheDocument();
      expect(screen.getByText('Section 3')).toBeInTheDocument();
    });

    it('renders icons when provided', () => {
      render(<SettingsAccordion items={testItems} />);

      expect(screen.getByText('📌')).toBeInTheDocument();
      expect(screen.getByText('⚙️')).toBeInTheDocument();
    });

    it('renders Expand All and Collapse All buttons', () => {
      render(<SettingsAccordion items={testItems} />);

      expect(screen.getByRole('button', { name: /expand all/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /collapse all/i })).toBeInTheDocument();
    });

    it('applies custom className', () => {
      const { container } = render(
        <SettingsAccordion items={testItems} className="custom-class" />
      );

      expect(container.querySelector('.settings-accordion')).toHaveClass('custom-class');
    });
  });

  describe('expand/collapse behavior', () => {
    it('expands section when header is clicked', () => {
      render(<SettingsAccordion items={testItems} />);

      const header = screen.getByRole('button', { name: /section 1/i });
      fireEvent.click(header);

      expect(screen.getByText('Content 1')).toBeVisible();
    });

    it('collapses section when header is clicked again', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1']} />);

      const header = screen.getByRole('button', { name: /section 1/i });
      fireEvent.click(header);

      expect(screen.getByText('Content 1')).not.toBeVisible();
    });

    it('expands defaultExpanded sections on mount', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1', 'section2']} />);

      expect(screen.getByText('Content 1')).toBeVisible();
      expect(screen.getByText('Content 2')).toBeVisible();
      expect(screen.getByText('Content 3')).not.toBeVisible();
    });

    it('allows multiple sections when allowMultiple is true', () => {
      render(<SettingsAccordion items={testItems} allowMultiple={true} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header2 = screen.getByRole('button', { name: /section 2/i });

      fireEvent.click(header1);
      fireEvent.click(header2);

      expect(screen.getByText('Content 1')).toBeVisible();
      expect(screen.getByText('Content 2')).toBeVisible();
    });

    it('collapses other sections when allowMultiple is false', () => {
      render(
        <SettingsAccordion
          items={testItems}
          allowMultiple={false}
          defaultExpanded={['section1']}
        />
      );

      const header2 = screen.getByRole('button', { name: /section 2/i });
      fireEvent.click(header2);

      expect(screen.getByText('Content 1')).not.toBeVisible();
      expect(screen.getByText('Content 2')).toBeVisible();
    });
  });

  describe('Expand All / Collapse All', () => {
    it('expands all sections when Expand All is clicked', () => {
      render(<SettingsAccordion items={testItems} />);

      const expandAllButton = screen.getByRole('button', { name: /expand all/i });
      fireEvent.click(expandAllButton);

      expect(screen.getByText('Content 1')).toBeVisible();
      expect(screen.getByText('Content 2')).toBeVisible();
      expect(screen.getByText('Content 3')).toBeVisible();
    });

    it('collapses all sections when Collapse All is clicked', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1', 'section2', 'section3']} />);

      const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });
      fireEvent.click(collapseAllButton);

      expect(screen.getByText('Content 1')).not.toBeVisible();
      expect(screen.getByText('Content 2')).not.toBeVisible();
      expect(screen.getByText('Content 3')).not.toBeVisible();
    });

    it('disables Expand All button when all sections are expanded', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1', 'section2', 'section3']} />);

      const expandAllButton = screen.getByRole('button', { name: /expand all/i });
      expect(expandAllButton).toBeDisabled();
    });

    it('disables Collapse All button when no sections are expanded', () => {
      render(<SettingsAccordion items={testItems} />);

      const collapseAllButton = screen.getByRole('button', { name: /collapse all/i });
      expect(collapseAllButton).toBeDisabled();
    });
  });

  describe('accessibility', () => {
    it('sets aria-expanded correctly', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1']} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header2 = screen.getByRole('button', { name: /section 2/i });

      expect(header1).toHaveAttribute('aria-expanded', 'true');
      expect(header2).toHaveAttribute('aria-expanded', 'false');
    });

    it('sets aria-controls correctly', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      expect(header1).toHaveAttribute('aria-controls', 'accordion-panel-section1');
    });

    it('has proper heading structure', () => {
      render(<SettingsAccordion items={testItems} />);

      const headings = screen.getAllByRole('heading', { level: 2 });
      expect(headings).toHaveLength(3);
    });

    it('panel has proper role and aria-labelledby', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1']} />);

      const panel = screen.getByRole('region', { name: /section 1/i });
      expect(panel).toHaveAttribute('aria-labelledby', 'accordion-header-section1');
    });
  });

  describe('keyboard navigation', () => {
    it('toggles section on Enter key', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      fireEvent.keyDown(header1, { key: 'Enter' });

      expect(screen.getByText('Content 1')).toBeVisible();
    });

    it('toggles section on Space key', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      fireEvent.keyDown(header1, { key: ' ' });

      expect(screen.getByText('Content 1')).toBeVisible();
    });

    it('moves focus to next section on ArrowDown', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header2 = screen.getByRole('button', { name: /section 2/i });

      header1.focus();
      fireEvent.keyDown(header1, { key: 'ArrowDown' });

      expect(document.activeElement).toBe(header2);
    });

    it('moves focus to previous section on ArrowUp', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header2 = screen.getByRole('button', { name: /section 2/i });

      header2.focus();
      fireEvent.keyDown(header2, { key: 'ArrowUp' });

      expect(document.activeElement).toBe(header1);
    });

    it('moves focus to first section on Home key', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header3 = screen.getByRole('button', { name: /section 3/i });

      header3.focus();
      fireEvent.keyDown(header3, { key: 'Home' });

      expect(document.activeElement).toBe(header1);
    });

    it('moves focus to last section on End key', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header3 = screen.getByRole('button', { name: /section 3/i });

      header1.focus();
      fireEvent.keyDown(header1, { key: 'End' });

      expect(document.activeElement).toBe(header3);
    });

    it('wraps focus from last to first on ArrowDown', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header3 = screen.getByRole('button', { name: /section 3/i });

      header3.focus();
      fireEvent.keyDown(header3, { key: 'ArrowDown' });

      expect(document.activeElement).toBe(header1);
    });

    it('wraps focus from first to last on ArrowUp', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const header3 = screen.getByRole('button', { name: /section 3/i });

      header1.focus();
      fireEvent.keyDown(header1, { key: 'ArrowUp' });

      expect(document.activeElement).toBe(header3);
    });
  });

  describe('persistence', () => {
    it('persists expanded sections to localStorage', () => {
      render(<SettingsAccordion items={testItems} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      fireEvent.click(header1);

      const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
      expect(stored).toContain('section1');
    });
  });

  describe('visual states', () => {
    it('applies expanded class to expanded items', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1']} />);

      const item1 = document.getElementById('accordion-section1');
      const item2 = document.getElementById('accordion-section2');

      expect(item1).toHaveClass('expanded');
      expect(item2).toHaveClass('collapsed');
    });

    it('applies expanded class to chevron when expanded', () => {
      render(<SettingsAccordion items={testItems} defaultExpanded={['section1']} />);

      const header1 = screen.getByRole('button', { name: /section 1/i });
      const chevron1 = header1.querySelector('.accordion-chevron');

      expect(chevron1).toHaveClass('expanded');
    });
  });
});

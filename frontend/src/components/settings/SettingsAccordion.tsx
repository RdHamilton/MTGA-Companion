import { useCallback } from 'react';
import type { ReactNode } from 'react';
import { useSettingsAccordion } from '../../hooks/useSettingsAccordion';
import type { AccordionSection } from '../../hooks/useSettingsAccordion';
import './SettingsAccordion.css';

export interface SettingsAccordionItem {
  id: string;
  label: string;
  icon?: string;
  content: ReactNode;
}

export interface SettingsAccordionProps {
  items: SettingsAccordionItem[];
  defaultExpanded?: string[];
  allowMultiple?: boolean;
  className?: string;
}

export function SettingsAccordion({
  items,
  defaultExpanded = [],
  allowMultiple = true,
  className = '',
}: SettingsAccordionProps) {
  const sections: AccordionSection[] = items.map((item) => ({
    id: item.id,
    label: item.label,
    icon: item.icon,
  }));

  const {
    isExpanded,
    toggleSection,
    handleKeyDown,
    registerHeaderRef,
    expandAll,
    collapseAll,
    expandedSections,
  } = useSettingsAccordion({
    sections,
    defaultExpanded,
    allowMultiple,
  });

  const handleHeaderClick = useCallback(
    (sectionId: string) => {
      toggleSection(sectionId);
    },
    [toggleSection]
  );

  const allExpanded = expandedSections.size === items.length;
  const noneExpanded = expandedSections.size === 0;

  return (
    <div className={`settings-accordion ${className}`}>
      {/* Expand/Collapse All Controls */}
      <div className="accordion-controls">
        <button
          className="accordion-control-button"
          onClick={expandAll}
          disabled={allExpanded}
          aria-label="Expand all sections"
        >
          Expand All
        </button>
        <button
          className="accordion-control-button"
          onClick={collapseAll}
          disabled={noneExpanded}
          aria-label="Collapse all sections"
        >
          Collapse All
        </button>
      </div>

      {/* Accordion Items */}
      <div className="accordion-items" role="region" aria-label="Settings sections">
        {items.map((item) => {
          const expanded = isExpanded(item.id);
          const panelId = `accordion-panel-${item.id}`;
          const headerId = `accordion-header-${item.id}`;

          return (
            <div
              key={item.id}
              id={`accordion-${item.id}`}
              className={`accordion-item ${expanded ? 'expanded' : 'collapsed'}`}
            >
              <h2 className="accordion-header-wrapper">
                <button
                  id={headerId}
                  ref={(ref) => registerHeaderRef(item.id, ref)}
                  className="accordion-header"
                  onClick={() => handleHeaderClick(item.id)}
                  onKeyDown={(e) => handleKeyDown(e, item.id)}
                  aria-expanded={expanded}
                  aria-controls={panelId}
                >
                  <span className="accordion-header-content">
                    {item.icon && <span className="accordion-icon">{item.icon}</span>}
                    <span className="accordion-label">{item.label}</span>
                  </span>
                  <span className={`accordion-chevron ${expanded ? 'expanded' : ''}`}>
                    ▼
                  </span>
                </button>
              </h2>
              <div
                id={panelId}
                role="region"
                aria-labelledby={headerId}
                className={`accordion-panel ${expanded ? 'expanded' : 'collapsed'}`}
                hidden={!expanded}
              >
                <div className="accordion-panel-content">{item.content}</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

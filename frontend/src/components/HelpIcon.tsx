import { useState, useRef, useEffect, type ReactNode } from 'react';
import './HelpIcon.css';

interface HelpIconProps {
  /** Title shown at top of popover */
  title: string;
  /** Main content of the help popover */
  children: ReactNode;
  /** Position of popover relative to icon */
  position?: 'top' | 'bottom' | 'left' | 'right';
  /** Size of the help icon */
  size?: 'small' | 'medium' | 'large';
}

/**
 * HelpIcon - A "?" icon that shows a help popover on click.
 * Use for longer explanations and contextual documentation.
 */
export default function HelpIcon({
  title,
  children,
  position = 'bottom',
  size = 'small',
}: HelpIconProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close on click outside
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  // Close on escape key
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen]);

  return (
    <div className="help-icon-container" ref={containerRef}>
      <button
        className={`help-icon-button help-icon-${size}`}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-label={`Help: ${title}`}
        type="button"
      >
        ?
      </button>
      {isOpen && (
        <div className={`help-popover help-popover-${position}`}>
          <div className="help-popover-header">
            <span className="help-popover-title">{title}</span>
            <button
              className="help-popover-close"
              onClick={() => setIsOpen(false)}
              aria-label="Close help"
              type="button"
            >
              &times;
            </button>
          </div>
          <div className="help-popover-content">{children}</div>
        </div>
      )}
    </div>
  );
}

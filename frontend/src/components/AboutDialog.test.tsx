import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import AboutDialog from './AboutDialog';

describe('AboutDialog', () => {
  describe('Visibility', () => {
    it('should render when isOpen is true', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('About MTGA Companion')).toBeInTheDocument();
    });

    it('should not render when isOpen is false', () => {
      render(<AboutDialog isOpen={false} onClose={vi.fn()} />);

      expect(screen.queryByText('About MTGA Companion')).not.toBeInTheDocument();
    });

    it('should render modal overlay when open', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const overlay = container.querySelector('.modal-overlay');
      expect(overlay).toBeInTheDocument();
    });

    it('should not render modal overlay when closed', () => {
      const { container } = render(<AboutDialog isOpen={false} onClose={vi.fn()} />);

      const overlay = container.querySelector('.modal-overlay');
      expect(overlay).not.toBeInTheDocument();
    });
  });

  describe('Header', () => {
    it('should render dialog title', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const title = screen.getByText('About MTGA Companion');
      expect(title.tagName).toBe('H2');
    });

    it('should render close button', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const closeButtons = screen.getAllByRole('button', { name: 'Close' });
      const headerCloseButton = closeButtons.find(btn => btn.textContent === '×');
      expect(headerCloseButton).toBeInTheDocument();
      expect(headerCloseButton?.textContent).toBe('×');
    });

    it('should call onClose when close button clicked', () => {
      const onClose = vi.fn();
      render(<AboutDialog isOpen={true} onClose={onClose} />);

      const closeButtons = screen.getAllByRole('button', { name: 'Close' });
      const headerCloseButton = closeButtons.find(btn => btn.textContent === '×');
      fireEvent.click(headerCloseButton!);

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe('App Information', () => {
    it('should render app icon', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('🃏')).toBeInTheDocument();
    });

    it('should render app name', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const appName = screen.getByText('MTGA Companion');
      expect(appName).toHaveClass('app-name');
      expect(appName.tagName).toBe('H3');
    });

    it('should render version number', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText(/Version 1\.3\.1/)).toBeInTheDocument();
    });
  });

  describe('About Section', () => {
    it('should render about heading', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('About')).toBeInTheDocument();
    });

    it('should render about description', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText(/MTGA Companion is a desktop application/)).toBeInTheDocument();
      expect(screen.getByText(/tracking and analyzing your Magic: The Gathering Arena matches/)).toBeInTheDocument();
    });
  });

  describe('Features Section', () => {
    it('should render features heading', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Features')).toBeInTheDocument();
    });

    it('should render all feature items', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Real-time match tracking from MTGA logs')).toBeInTheDocument();
      expect(screen.getByText('Comprehensive statistics and analytics')).toBeInTheDocument();
      expect(screen.getByText('Win rate trends and predictions')).toBeInTheDocument();
      expect(screen.getByText('Deck performance analysis')).toBeInTheDocument();
      expect(screen.getByText('Rank progression tracking')).toBeInTheDocument();
      expect(screen.getByText('Format-specific breakdowns')).toBeInTheDocument();
    });

    it('should render features as list', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const featureList = container.querySelector('.feature-list');
      expect(featureList?.tagName).toBe('UL');
      expect(featureList?.querySelectorAll('li').length).toBe(6);
    });
  });

  describe('Acknowledgments Section', () => {
    it('should render acknowledgments heading', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Acknowledgments')).toBeInTheDocument();
    });

    it('should render credit to Scryfall', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Scryfall')).toBeInTheDocument();
      expect(screen.getByText(/Card metadata and imagery/)).toBeInTheDocument();
    });

    it('should render credit to 17Lands', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('17Lands')).toBeInTheDocument();
      expect(screen.getByText(/Draft statistics and card ratings/)).toBeInTheDocument();
    });

    it('should render credit to Wizards of the Coast', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Wizards of the Coast')).toBeInTheDocument();

      // Check for "Magic: The Gathering Arena" in the credits list specifically
      const creditList = container.querySelector('.credit-list');
      expect(creditList?.textContent).toContain('Magic: The Gathering Arena');
    });

    it('should render credits as list', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const creditList = container.querySelector('.credit-list');
      expect(creditList?.tagName).toBe('UL');
      expect(creditList?.querySelectorAll('li').length).toBe(3);
    });
  });

  describe('License Section', () => {
    it('should render license heading', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('License')).toBeInTheDocument();
    });

    it('should render license information', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText(/MIT License/)).toBeInTheDocument();
    });
  });

  describe('Links Section', () => {
    it('should render links heading', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Links')).toBeInTheDocument();
    });

    it('should render GitHub repository link', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const link = screen.getByText('GitHub Repository').closest('a');
      expect(link).toHaveAttribute('href', 'https://github.com/RdHamilton/MTGA-Companion');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('should render issue tracker link', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const link = screen.getByText('Report an Issue').closest('a');
      expect(link).toHaveAttribute('href', 'https://github.com/RdHamilton/MTGA-Companion/issues');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('should render documentation link', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const link = screen.getByText('Documentation').closest('a');
      expect(link).toHaveAttribute('href', 'https://github.com/RdHamilton/MTGA-Companion/wiki');
      expect(link).toHaveAttribute('target', '_blank');
      expect(link).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('should have security attributes on all external links', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const externalLinks = container.querySelectorAll('a[target="_blank"]');
      expect(externalLinks.length).toBe(3);

      externalLinks.forEach(link => {
        expect(link).toHaveAttribute('rel', 'noopener noreferrer');
      });
    });
  });

  describe('Footer', () => {
    it('should render copyright notice', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const currentYear = new Date().getFullYear();
      const copyrightRegex = new RegExp(`2024-${currentYear} Ray Hamilton Engineering LLC`);
      expect(screen.getByText(copyrightRegex)).toBeInTheDocument();
    });

    it('should render disclaimer', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText(/not affiliated with or endorsed by Wizards of the Coast/)).toBeInTheDocument();
      expect(screen.getByText(/Magic: The Gathering Arena is a trademark/)).toBeInTheDocument();
    });

    it('should render close button in footer', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      // There are two close buttons: one in header (×) and one in footer (Close)
      const closeButtons = screen.getAllByRole('button', { name: /close/i });
      expect(closeButtons.length).toBe(2);

      const footerCloseButton = closeButtons.find(btn => btn.textContent === 'Close');
      expect(footerCloseButton).toBeInTheDocument();
    });

    it('should call onClose when footer close button clicked', () => {
      const onClose = vi.fn();
      render(<AboutDialog isOpen={true} onClose={onClose} />);

      const closeButtons = screen.getAllByRole('button', { name: 'Close' });
      const footerCloseButton = closeButtons.find(btn => btn.textContent === 'Close');
      fireEvent.click(footerCloseButton!);

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe('Modal Behavior', () => {
    it('should call onClose when overlay clicked', () => {
      const onClose = vi.fn();
      const { container } = render(<AboutDialog isOpen={true} onClose={onClose} />);

      const overlay = container.querySelector('.modal-overlay');
      fireEvent.click(overlay!);

      expect(onClose).toHaveBeenCalledOnce();
    });

    it('should not call onClose when modal content clicked', () => {
      const onClose = vi.fn();
      const { container } = render(<AboutDialog isOpen={true} onClose={onClose} />);

      const modalContent = container.querySelector('.modal-content');
      fireEvent.click(modalContent!);

      expect(onClose).not.toHaveBeenCalled();
    });

    it('should stop event propagation on content click', () => {
      const onClose = vi.fn();
      const { container } = render(<AboutDialog isOpen={true} onClose={onClose} />);

      // Click on modal content
      const modalContent = container.querySelector('.modal-content');
      fireEvent.click(modalContent!);

      // Overlay click handler should not fire
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe('Component Structure', () => {
    it('should have correct CSS classes', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      expect(container.querySelector('.modal-overlay')).toBeInTheDocument();
      expect(container.querySelector('.modal-content')).toBeInTheDocument();
      expect(container.querySelector('.about-dialog')).toBeInTheDocument();
      expect(container.querySelector('.modal-header')).toBeInTheDocument();
      expect(container.querySelector('.modal-body')).toBeInTheDocument();
      expect(container.querySelector('.modal-footer')).toBeInTheDocument();
    });

    it('should have multiple about sections', () => {
      const { container } = render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const sections = container.querySelectorAll('.about-section');
      expect(sections.length).toBeGreaterThan(5);
    });

    it('should have proper heading hierarchy', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      // Main title is h2
      expect(screen.getByText('About MTGA Companion').tagName).toBe('H2');

      // App name is h3
      expect(screen.getByText('MTGA Companion').tagName).toBe('H3');

      // Section titles are h4
      const sectionHeadings = ['About', 'Features', 'Acknowledgments', 'License', 'Links'];
      sectionHeadings.forEach(heading => {
        const element = screen.getByText(heading);
        expect(element.tagName).toBe('H4');
      });
    });
  });

  describe('Dynamic Content', () => {
    it('should display current year in copyright', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const currentYear = new Date().getFullYear();
      const copyrightText = screen.getByText(/Ray Hamilton Engineering LLC/);
      expect(copyrightText.textContent).toContain(currentYear.toString());
    });

    it('should display year range when not 2024', () => {
      // This test verifies the copyright shows 2024-CURRENT_YEAR format
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const currentYear = new Date().getFullYear();
      const copyrightText = screen.getByText(/2024-/);

      if (currentYear > 2024) {
        expect(copyrightText.textContent).toContain(`2024-${currentYear}`);
      } else {
        expect(copyrightText.textContent).toContain('2024');
      }
    });
  });

  describe('Accessibility', () => {
    it('should have close button with aria-label', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const headerCloseButton = screen.getByLabelText('Close');
      expect(headerCloseButton).toBeInTheDocument();
    });

    it('should have semantic heading structure', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const h2 = screen.getByRole('heading', { level: 2 });
      expect(h2).toHaveTextContent('About MTGA Companion');

      const h3 = screen.getByRole('heading', { level: 3 });
      expect(h3).toHaveTextContent('MTGA Companion');
    });

    it('should have clickable buttons', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should have clickable links', () => {
      render(<AboutDialog isOpen={true} onClose={vi.fn()} />);

      const links = screen.getAllByRole('link');
      expect(links.length).toBe(3); // GitHub, Issues, Documentation
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid open/close toggling', () => {
      const onClose = vi.fn();
      const { rerender } = render(<AboutDialog isOpen={false} onClose={onClose} />);

      rerender(<AboutDialog isOpen={true} onClose={onClose} />);
      expect(screen.getByText('About MTGA Companion')).toBeInTheDocument();

      rerender(<AboutDialog isOpen={false} onClose={onClose} />);
      expect(screen.queryByText('About MTGA Companion')).not.toBeInTheDocument();

      rerender(<AboutDialog isOpen={true} onClose={onClose} />);
      expect(screen.getByText('About MTGA Companion')).toBeInTheDocument();
    });

    it('should handle multiple close calls', () => {
      const onClose = vi.fn();
      render(<AboutDialog isOpen={true} onClose={onClose} />);

      const closeButtons = screen.getAllByRole('button', { name: 'Close' });
      const headerCloseButton = closeButtons.find(btn => btn.textContent === '×');

      fireEvent.click(headerCloseButton!);
      fireEvent.click(headerCloseButton!);
      fireEvent.click(headerCloseButton!);

      expect(onClose).toHaveBeenCalledTimes(3);
    });

    it('should handle undefined onClose gracefully', () => {
      // This should not throw an error
      const { container } = render(<AboutDialog isOpen={true} onClose={undefined as unknown as () => void} />);

      const overlay = container.querySelector('.modal-overlay');
      expect(() => fireEvent.click(overlay!)).not.toThrow();
    });
  });
});

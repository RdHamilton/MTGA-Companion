import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AboutSection } from './AboutSection';

describe('AboutSection', () => {
  const defaultProps = {
    onShowAboutDialog: vi.fn(),
    isDeveloperMode: false,
    onVersionClick: vi.fn(),
    onToggleDeveloperMode: vi.fn(),
  };

  it('renders section title', () => {
    render(<AboutSection {...defaultProps} />);
    expect(screen.getByText('About')).toBeInTheDocument();
  });

  it('displays version info', () => {
    render(<AboutSection {...defaultProps} />);
    expect(screen.getByText('Version:')).toBeInTheDocument();
    expect(screen.getByText('1.0.0')).toBeInTheDocument();
  });

  it('displays build info', () => {
    render(<AboutSection {...defaultProps} />);
    expect(screen.getByText('Build:')).toBeInTheDocument();
    expect(screen.getByText('Development')).toBeInTheDocument();
  });

  it('displays platform info', () => {
    render(<AboutSection {...defaultProps} />);
    expect(screen.getByText('Platform:')).toBeInTheDocument();
    expect(screen.getByText('Wails + React')).toBeInTheDocument();
  });

  it('renders about button', () => {
    render(<AboutSection {...defaultProps} />);
    expect(screen.getByRole('button', { name: 'About MTGA Companion' })).toBeInTheDocument();
  });

  it('calls onShowAboutDialog when button is clicked', () => {
    const onShowAboutDialog = vi.fn();
    render(<AboutSection {...defaultProps} onShowAboutDialog={onShowAboutDialog} />);

    fireEvent.click(screen.getByRole('button', { name: 'About MTGA Companion' }));

    expect(onShowAboutDialog).toHaveBeenCalled();
  });

  describe('version click (developer mode activation)', () => {
    it('calls onVersionClick when version is clicked', () => {
      const onVersionClick = vi.fn();
      render(<AboutSection {...defaultProps} onVersionClick={onVersionClick} />);

      fireEvent.click(screen.getByText('1.0.0'));

      expect(onVersionClick).toHaveBeenCalled();
    });

    it('version element has clickable class', () => {
      render(<AboutSection {...defaultProps} />);
      const versionElement = screen.getByText('1.0.0');
      expect(versionElement).toHaveClass('about-version-clickable');
    });
  });

  describe('developer mode indicator', () => {
    it('does not show developer mode indicator when disabled', () => {
      render(<AboutSection {...defaultProps} isDeveloperMode={false} />);
      expect(screen.queryByText('Developer Mode:')).not.toBeInTheDocument();
    });

    it('shows developer mode indicator when enabled', () => {
      render(<AboutSection {...defaultProps} isDeveloperMode={true} />);
      expect(screen.getByText('Developer Mode:')).toBeInTheDocument();
      expect(screen.getByText('Enabled')).toBeInTheDocument();
    });

    it('renders disable button when developer mode is enabled', () => {
      render(<AboutSection {...defaultProps} isDeveloperMode={true} />);
      expect(screen.getByRole('button', { name: 'Disable' })).toBeInTheDocument();
    });

    it('calls onToggleDeveloperMode when disable button is clicked', () => {
      const onToggleDeveloperMode = vi.fn();
      render(<AboutSection {...defaultProps} isDeveloperMode={true} onToggleDeveloperMode={onToggleDeveloperMode} />);

      fireEvent.click(screen.getByRole('button', { name: 'Disable' }));

      expect(onToggleDeveloperMode).toHaveBeenCalled();
    });
  });
});

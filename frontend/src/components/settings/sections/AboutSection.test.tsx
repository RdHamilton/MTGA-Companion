import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AboutSection } from './AboutSection';

describe('AboutSection', () => {
  const defaultProps = {
    onShowAboutDialog: vi.fn(),
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
    expect(screen.getByText('About MTGA Companion')).toBeInTheDocument();
  });

  it('calls onShowAboutDialog when button is clicked', () => {
    const onShowAboutDialog = vi.fn();
    render(<AboutSection onShowAboutDialog={onShowAboutDialog} />);

    fireEvent.click(screen.getByText('About MTGA Companion'));

    expect(onShowAboutDialog).toHaveBeenCalled();
  });
});

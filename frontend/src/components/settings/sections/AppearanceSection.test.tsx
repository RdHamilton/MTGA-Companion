import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AppearanceSection } from './AppearanceSection';

describe('AppearanceSection', () => {
  it('renders section title', () => {
    render(<AppearanceSection />);
    expect(screen.getByText('Appearance')).toBeInTheDocument();
  });

  it('renders theme selector', () => {
    render(<AppearanceSection />);
    expect(screen.getByText('Theme')).toBeInTheDocument();
  });

  it('renders theme options', () => {
    render(<AppearanceSection />);
    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
  });

  it('has dark theme selected by default', () => {
    render(<AppearanceSection />);
    const select = screen.getByRole('combobox');
    expect(select).toHaveValue('dark');
  });

  it('renders theme description', () => {
    render(<AppearanceSection />);
    expect(screen.getByText('Choose your preferred color scheme')).toBeInTheDocument();
  });

  it('includes all theme options', () => {
    render(<AppearanceSection />);
    expect(screen.getByText('Dark (Default)')).toBeInTheDocument();
    expect(screen.getByText('Light (Coming Soon)')).toBeInTheDocument();
    expect(screen.getByText('Auto (System Default)')).toBeInTheDocument();
  });
});

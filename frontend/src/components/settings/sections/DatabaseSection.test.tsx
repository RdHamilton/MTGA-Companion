import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DatabaseSection } from './DatabaseSection';

describe('DatabaseSection', () => {
  const defaultProps = {
    dbPath: '',
    onDbPathChange: vi.fn(),
  };

  it('renders section title', () => {
    render(<DatabaseSection {...defaultProps} />);
    expect(screen.getByText('Database Configuration')).toBeInTheDocument();
  });

  it('renders database path input', () => {
    render(<DatabaseSection {...defaultProps} />);
    expect(screen.getByPlaceholderText(/mtga\.db/)).toBeInTheDocument();
  });

  it('displays current database path', () => {
    render(<DatabaseSection {...defaultProps} dbPath="/path/to/db" />);
    expect(screen.getByDisplayValue('/path/to/db')).toBeInTheDocument();
  });

  it('calls onDbPathChange when input changes', () => {
    const onDbPathChange = vi.fn();
    render(<DatabaseSection {...defaultProps} onDbPathChange={onDbPathChange} />);

    const input = screen.getByPlaceholderText(/mtga\.db/);
    fireEvent.change(input, { target: { value: '/new/path' } });

    expect(onDbPathChange).toHaveBeenCalledWith('/new/path');
  });

  it('renders browse button', () => {
    render(<DatabaseSection {...defaultProps} />);
    expect(screen.getByText('Browse...')).toBeInTheDocument();
  });

  it('renders description text', () => {
    render(<DatabaseSection {...defaultProps} />);
    expect(screen.getByText(/Location of the MTGA Companion database file/)).toBeInTheDocument();
  });
});

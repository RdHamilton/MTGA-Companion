import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SettingSelect from './SettingSelect';

describe('SettingSelect', () => {
  const defaultProps = {
    label: 'Test Select',
    value: 'option1',
    onChange: vi.fn(),
    options: [
      { value: 'option1', label: 'Option 1' },
      { value: 'option2', label: 'Option 2' },
      { value: 'option3', label: 'Option 3' },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic Rendering', () => {
    it('should render label text', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(screen.getByText('Test Select')).toBeInTheDocument();
    });

    it('should render select element', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should render all options', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(screen.getByRole('option', { name: 'Option 1' })).toBeInTheDocument();
      expect(screen.getByRole('option', { name: 'Option 2' })).toBeInTheDocument();
      expect(screen.getByRole('option', { name: 'Option 3' })).toBeInTheDocument();
    });

    it('should render description when provided', () => {
      render(<SettingSelect {...defaultProps} description="Test Description" />);

      expect(screen.getByText('Test Description')).toBeInTheDocument();
    });

    it('should not render description when not provided', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(document.querySelector('.setting-description')).not.toBeInTheDocument();
    });

    it('should render hint when provided', () => {
      render(<SettingSelect {...defaultProps} hint="Test Hint" />);

      expect(screen.getByText('Test Hint')).toBeInTheDocument();
    });

    it('should not render hint when not provided', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(document.querySelector('.setting-hint')).not.toBeInTheDocument();
    });
  });

  describe('Select Value', () => {
    it('should show selected value', () => {
      render(<SettingSelect {...defaultProps} value="option2" />);

      const select = screen.getByRole('combobox') as HTMLSelectElement;
      expect(select.value).toBe('option2');
    });

    it('should call onChange with new value when selection changes', () => {
      render(<SettingSelect {...defaultProps} />);

      fireEvent.change(screen.getByRole('combobox'), { target: { value: 'option2' } });

      expect(defaultProps.onChange).toHaveBeenCalledWith('option2');
    });

    it('should call onChange with correct value for each option', () => {
      render(<SettingSelect {...defaultProps} />);
      const select = screen.getByRole('combobox');

      fireEvent.change(select, { target: { value: 'option3' } });
      expect(defaultProps.onChange).toHaveBeenCalledWith('option3');

      fireEvent.change(select, { target: { value: 'option1' } });
      expect(defaultProps.onChange).toHaveBeenCalledWith('option1');
    });
  });

  describe('Disabled State', () => {
    it('should not be disabled by default', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(screen.getByRole('combobox')).not.toBeDisabled();
    });

    it('should be disabled when disabled prop is true', () => {
      render(<SettingSelect {...defaultProps} disabled={true} />);

      expect(screen.getByRole('combobox')).toBeDisabled();
    });
  });

  describe('CSS Classes', () => {
    it('should have setting-item class', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(document.querySelector('.setting-item')).toBeInTheDocument();
    });

    it('should have select-input class on select', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(screen.getByRole('combobox')).toHaveClass('select-input');
    });

    it('should have setting-label class on label', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(document.querySelector('.setting-label')).toBeInTheDocument();
    });

    it('should have setting-control class on control wrapper', () => {
      render(<SettingSelect {...defaultProps} />);

      expect(document.querySelector('.setting-control')).toBeInTheDocument();
    });
  });

  describe('Structure', () => {
    it('should wrap description in setting-description span', () => {
      render(<SettingSelect {...defaultProps} description="Test Description" />);

      const description = screen.getByText('Test Description');
      expect(description.tagName).toBe('SPAN');
      expect(description).toHaveClass('setting-description');
    });

    it('should wrap hint in setting-hint span', () => {
      render(<SettingSelect {...defaultProps} hint="Test Hint" />);

      const hint = screen.getByText('Test Hint');
      expect(hint.tagName).toBe('SPAN');
      expect(hint).toHaveClass('setting-hint');
    });
  });

  describe('Options with Keys', () => {
    it('should use option value as key', () => {
      const { rerender } = render(<SettingSelect {...defaultProps} />);

      // Rerender with different options to ensure keys work correctly
      rerender(
        <SettingSelect
          {...defaultProps}
          options={[
            { value: 'new1', label: 'New Option 1' },
            { value: 'new2', label: 'New Option 2' },
          ]}
        />
      );

      expect(screen.getByRole('option', { name: 'New Option 1' })).toBeInTheDocument();
      expect(screen.getByRole('option', { name: 'New Option 2' })).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should be focusable', () => {
      render(<SettingSelect {...defaultProps} />);

      const select = screen.getByRole('combobox');
      select.focus();

      expect(document.activeElement).toBe(select);
    });

    it('should support keyboard navigation', () => {
      render(<SettingSelect {...defaultProps} />);

      const select = screen.getByRole('combobox');
      select.focus();

      // Simulate keyboard selection
      fireEvent.keyDown(select, { key: 'ArrowDown' });
      fireEvent.change(select, { target: { value: 'option2' } });

      expect(defaultProps.onChange).toHaveBeenCalledWith('option2');
    });
  });

  describe('Empty Options', () => {
    it('should render without options', () => {
      render(<SettingSelect {...defaultProps} options={[]} />);

      const select = screen.getByRole('combobox');
      expect(select.querySelectorAll('option')).toHaveLength(0);
    });
  });
});

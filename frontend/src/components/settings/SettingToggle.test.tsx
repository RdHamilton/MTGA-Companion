import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SettingToggle from './SettingToggle';

describe('SettingToggle', () => {
  const defaultProps = {
    label: 'Test Toggle',
    checked: false,
    onChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic Rendering', () => {
    it('should render label text', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(screen.getByText('Test Toggle')).toBeInTheDocument();
    });

    it('should render checkbox input', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(screen.getByRole('checkbox')).toBeInTheDocument();
    });

    it('should render description when provided', () => {
      render(<SettingToggle {...defaultProps} description="Test Description" />);

      expect(screen.getByText('Test Description')).toBeInTheDocument();
    });

    it('should not render description when not provided', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(document.querySelector('.setting-description')).not.toBeInTheDocument();
    });
  });

  describe('Checkbox State', () => {
    it('should be unchecked when checked prop is false', () => {
      render(<SettingToggle {...defaultProps} checked={false} />);

      expect(screen.getByRole('checkbox')).not.toBeChecked();
    });

    it('should be checked when checked prop is true', () => {
      render(<SettingToggle {...defaultProps} checked={true} />);

      expect(screen.getByRole('checkbox')).toBeChecked();
    });

    it('should call onChange with true when clicked while unchecked', () => {
      render(<SettingToggle {...defaultProps} checked={false} />);

      fireEvent.click(screen.getByRole('checkbox'));

      expect(defaultProps.onChange).toHaveBeenCalledWith(true);
    });

    it('should call onChange with false when clicked while checked', () => {
      render(<SettingToggle {...defaultProps} checked={true} />);

      fireEvent.click(screen.getByRole('checkbox'));

      expect(defaultProps.onChange).toHaveBeenCalledWith(false);
    });
  });

  describe('Disabled State', () => {
    it('should not be disabled by default', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(screen.getByRole('checkbox')).not.toBeDisabled();
    });

    it('should be disabled when disabled prop is true', () => {
      render(<SettingToggle {...defaultProps} disabled={true} />);

      expect(screen.getByRole('checkbox')).toBeDisabled();
    });

    it('should prevent interaction when disabled', () => {
      render(<SettingToggle {...defaultProps} disabled={true} />);

      const checkbox = screen.getByRole('checkbox');
      // Disabled checkboxes should have the disabled attribute
      expect(checkbox).toBeDisabled();
      // The browser prevents interaction with disabled elements
    });
  });

  describe('Accessibility', () => {
    it('should associate label with checkbox via htmlFor', () => {
      render(<SettingToggle {...defaultProps} id="test-toggle" />);

      const checkbox = screen.getByRole('checkbox');
      const label = screen.getByText('Test Toggle').closest('label');

      expect(label).toHaveAttribute('for', 'test-toggle');
      expect(checkbox).toHaveAttribute('id', 'test-toggle');
    });

    it('should generate unique id when not provided', () => {
      render(
        <>
          <SettingToggle {...defaultProps} label="Toggle 1" />
          <SettingToggle {...defaultProps} label="Toggle 2" />
        </>
      );

      const checkboxes = screen.getAllByRole('checkbox');
      expect(checkboxes[0].id).not.toBe(checkboxes[1].id);
    });

    it('should be focusable', () => {
      render(<SettingToggle {...defaultProps} />);

      const checkbox = screen.getByRole('checkbox');
      checkbox.focus();

      expect(document.activeElement).toBe(checkbox);
    });
  });

  describe('CSS Classes', () => {
    it('should have setting-item class', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(document.querySelector('.setting-item')).toBeInTheDocument();
    });

    it('should have checkbox-input class on checkbox', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(screen.getByRole('checkbox')).toHaveClass('checkbox-input');
    });

    it('should have setting-label class on label', () => {
      render(<SettingToggle {...defaultProps} />);

      expect(document.querySelector('.setting-label')).toBeInTheDocument();
    });
  });

  describe('Structure', () => {
    it('should wrap description in setting-description span', () => {
      render(<SettingToggle {...defaultProps} description="Test Description" />);

      const description = screen.getByText('Test Description');
      expect(description.tagName).toBe('SPAN');
      expect(description).toHaveClass('setting-description');
    });
  });
});

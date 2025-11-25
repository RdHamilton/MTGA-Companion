import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AppPreferencesSection } from './AppPreferencesSection';

describe('AppPreferencesSection', () => {
  const defaultProps = {
    autoRefresh: false,
    onAutoRefreshChange: vi.fn(),
    refreshInterval: 30,
    onRefreshIntervalChange: vi.fn(),
    showNotifications: true,
    onShowNotificationsChange: vi.fn(),
    theme: 'dark',
    onThemeChange: vi.fn(),
  };

  it('renders section title', () => {
    render(<AppPreferencesSection {...defaultProps} />);
    expect(screen.getByText('Preferences')).toBeInTheDocument();
  });

  describe('theme selector', () => {
    it('renders theme selector', () => {
      render(<AppPreferencesSection {...defaultProps} />);
      expect(screen.getByText('Theme')).toBeInTheDocument();
    });

    it('displays current theme value', () => {
      render(<AppPreferencesSection {...defaultProps} theme="dark" />);
      const select = screen.getByRole('combobox');
      expect(select).toHaveValue('dark');
    });

    it('calls onThemeChange when theme changes', () => {
      const onThemeChange = vi.fn();
      render(<AppPreferencesSection {...defaultProps} onThemeChange={onThemeChange} />);

      const select = screen.getByRole('combobox');
      fireEvent.change(select, { target: { value: 'light' } });

      expect(onThemeChange).toHaveBeenCalledWith('light');
    });

    it('has all theme options', () => {
      render(<AppPreferencesSection {...defaultProps} />);
      expect(screen.getByText('Dark (Default)')).toBeInTheDocument();
      expect(screen.getByText('Light (Coming Soon)')).toBeInTheDocument();
      expect(screen.getByText('Auto (System Default)')).toBeInTheDocument();
    });
  });

  describe('auto-refresh toggle', () => {
    it('renders auto-refresh toggle', () => {
      render(<AppPreferencesSection {...defaultProps} />);
      expect(screen.getByText('Auto-refresh data')).toBeInTheDocument();
    });

    it('shows toggle as unchecked when autoRefresh is false', () => {
      render(<AppPreferencesSection {...defaultProps} autoRefresh={false} />);
      const checkbox = screen.getAllByRole('checkbox')[0];
      expect(checkbox).not.toBeChecked();
    });

    it('shows toggle as checked when autoRefresh is true', () => {
      render(<AppPreferencesSection {...defaultProps} autoRefresh={true} />);
      const checkbox = screen.getAllByRole('checkbox')[0];
      expect(checkbox).toBeChecked();
    });

    it('calls onAutoRefreshChange when toggled', () => {
      const onAutoRefreshChange = vi.fn();
      render(<AppPreferencesSection {...defaultProps} onAutoRefreshChange={onAutoRefreshChange} />);

      const checkbox = screen.getAllByRole('checkbox')[0];
      fireEvent.click(checkbox);

      expect(onAutoRefreshChange).toHaveBeenCalledWith(true);
    });
  });

  describe('refresh interval', () => {
    it('does not show refresh interval when autoRefresh is false', () => {
      render(<AppPreferencesSection {...defaultProps} autoRefresh={false} />);
      expect(screen.queryByText('Refresh Interval (seconds)')).not.toBeInTheDocument();
    });

    it('shows refresh interval when autoRefresh is true', () => {
      render(<AppPreferencesSection {...defaultProps} autoRefresh={true} />);
      expect(screen.getByText('Refresh Interval (seconds)')).toBeInTheDocument();
    });

    it('displays current refresh interval value', () => {
      render(<AppPreferencesSection {...defaultProps} autoRefresh={true} refreshInterval={60} />);
      expect(screen.getByDisplayValue('60')).toBeInTheDocument();
    });

    it('calls onRefreshIntervalChange when changed', () => {
      const onRefreshIntervalChange = vi.fn();
      render(
        <AppPreferencesSection
          {...defaultProps}
          autoRefresh={true}
          onRefreshIntervalChange={onRefreshIntervalChange}
        />
      );

      const input = screen.getByDisplayValue('30');
      fireEvent.change(input, { target: { value: '60' } });

      expect(onRefreshIntervalChange).toHaveBeenCalledWith(60);
    });
  });

  describe('notifications toggle', () => {
    it('renders notifications toggle', () => {
      render(<AppPreferencesSection {...defaultProps} />);
      expect(screen.getByText('Show notifications')).toBeInTheDocument();
    });

    it('shows toggle as checked when showNotifications is true', () => {
      render(<AppPreferencesSection {...defaultProps} showNotifications={true} />);
      const checkbox = screen.getAllByRole('checkbox')[1];
      expect(checkbox).toBeChecked();
    });

    it('shows toggle as unchecked when showNotifications is false', () => {
      render(<AppPreferencesSection {...defaultProps} showNotifications={false} />);
      const checkbox = screen.getAllByRole('checkbox')[1];
      expect(checkbox).not.toBeChecked();
    });

    it('calls onShowNotificationsChange when toggled', () => {
      const onShowNotificationsChange = vi.fn();
      render(
        <AppPreferencesSection
          {...defaultProps}
          showNotifications={true}
          onShowNotificationsChange={onShowNotificationsChange}
        />
      );

      const checkbox = screen.getAllByRole('checkbox')[1];
      fireEvent.click(checkbox);

      expect(onShowNotificationsChange).toHaveBeenCalledWith(false);
    });
  });
});

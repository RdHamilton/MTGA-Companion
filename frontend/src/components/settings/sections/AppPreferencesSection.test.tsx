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
    rotationNotificationsEnabled: true,
    onRotationNotificationsEnabledChange: vi.fn(),
    rotationNotificationThreshold: 30,
    onRotationNotificationThresholdChange: vi.fn(),
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
      const selects = screen.getAllByRole('combobox');
      expect(selects[0]).toHaveValue('dark');
    });

    it('calls onThemeChange when theme changes', () => {
      const onThemeChange = vi.fn();
      render(<AppPreferencesSection {...defaultProps} onThemeChange={onThemeChange} />);

      const selects = screen.getAllByRole('combobox');
      fireEvent.change(selects[0], { target: { value: 'light' } });

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

  describe('rotation notifications', () => {
    it('renders rotation notifications subsection', () => {
      render(<AppPreferencesSection {...defaultProps} />);
      expect(screen.getByText('Standard Rotation')).toBeInTheDocument();
    });

    it('renders rotation notifications toggle', () => {
      render(<AppPreferencesSection {...defaultProps} />);
      expect(screen.getByText('Rotation notifications')).toBeInTheDocument();
    });

    it('shows toggle as checked when rotationNotificationsEnabled is true', () => {
      render(<AppPreferencesSection {...defaultProps} rotationNotificationsEnabled={true} />);
      const checkbox = screen.getAllByRole('checkbox')[2];
      expect(checkbox).toBeChecked();
    });

    it('shows toggle as unchecked when rotationNotificationsEnabled is false', () => {
      render(<AppPreferencesSection {...defaultProps} rotationNotificationsEnabled={false} />);
      const checkbox = screen.getAllByRole('checkbox')[2];
      expect(checkbox).not.toBeChecked();
    });

    it('calls onRotationNotificationsEnabledChange when toggled', () => {
      const onRotationNotificationsEnabledChange = vi.fn();
      render(
        <AppPreferencesSection
          {...defaultProps}
          rotationNotificationsEnabled={true}
          onRotationNotificationsEnabledChange={onRotationNotificationsEnabledChange}
        />
      );

      const checkbox = screen.getAllByRole('checkbox')[2];
      fireEvent.click(checkbox);

      expect(onRotationNotificationsEnabledChange).toHaveBeenCalledWith(false);
    });

    it('does not show notification timing when rotationNotificationsEnabled is false', () => {
      render(<AppPreferencesSection {...defaultProps} rotationNotificationsEnabled={false} />);
      expect(screen.queryByText('Notification timing')).not.toBeInTheDocument();
    });

    it('shows notification timing when rotationNotificationsEnabled is true', () => {
      render(<AppPreferencesSection {...defaultProps} rotationNotificationsEnabled={true} />);
      expect(screen.getByText('Notification timing')).toBeInTheDocument();
    });

    it('displays current threshold value', () => {
      render(
        <AppPreferencesSection
          {...defaultProps}
          rotationNotificationsEnabled={true}
          rotationNotificationThreshold={60}
        />
      );
      const selects = screen.getAllByRole('combobox');
      expect(selects[1]).toHaveValue('60');
    });

    it('calls onRotationNotificationThresholdChange when changed', () => {
      const onRotationNotificationThresholdChange = vi.fn();
      render(
        <AppPreferencesSection
          {...defaultProps}
          rotationNotificationsEnabled={true}
          onRotationNotificationThresholdChange={onRotationNotificationThresholdChange}
        />
      );

      const selects = screen.getAllByRole('combobox');
      fireEvent.change(selects[1], { target: { value: '90' } });

      expect(onRotationNotificationThresholdChange).toHaveBeenCalledWith(90);
    });

    it('has all threshold options', () => {
      render(<AppPreferencesSection {...defaultProps} rotationNotificationsEnabled={true} />);
      expect(screen.getByText('7 days before rotation')).toBeInTheDocument();
      expect(screen.getByText('30 days before rotation')).toBeInTheDocument();
      expect(screen.getByText('60 days before rotation')).toBeInTheDocument();
      expect(screen.getByText('90 days before rotation')).toBeInTheDocument();
    });
  });
});

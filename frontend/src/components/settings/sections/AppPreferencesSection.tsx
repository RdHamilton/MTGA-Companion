import { SettingItem, SettingToggle, SettingSelect } from '../';

export interface AppPreferencesSectionProps {
  autoRefresh: boolean;
  onAutoRefreshChange: (value: boolean) => void;
  refreshInterval: number;
  onRefreshIntervalChange: (value: number) => void;
  showNotifications: boolean;
  onShowNotificationsChange: (value: boolean) => void;
  theme: string;
  onThemeChange: (value: string) => void;
}

export function AppPreferencesSection({
  autoRefresh,
  onAutoRefreshChange,
  refreshInterval,
  onRefreshIntervalChange,
  showNotifications,
  onShowNotificationsChange,
  theme,
  onThemeChange,
}: AppPreferencesSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">Preferences</h2>

      <SettingSelect
        label="Theme"
        description="Choose your preferred color scheme"
        value={theme}
        onChange={onThemeChange}
        options={[
          { value: 'dark', label: 'Dark (Default)' },
          { value: 'light', label: 'Light (Coming Soon)' },
          { value: 'auto', label: 'Auto (System Default)' },
        ]}
      />

      <SettingToggle
        label="Auto-refresh data"
        description="Automatically refresh statistics at regular intervals"
        checked={autoRefresh}
        onChange={onAutoRefreshChange}
      />

      {autoRefresh && (
        <SettingItem
          label="Refresh Interval (seconds)"
          description="How often to refresh data automatically"
          indented
        >
          <input
            type="number"
            value={refreshInterval}
            onChange={(e) => onRefreshIntervalChange(parseInt(e.target.value))}
            min="10"
            max="300"
            step="10"
            className="number-input"
          />
        </SettingItem>
      )}

      <SettingToggle
        label="Show notifications"
        description="Display notifications for match results and updates"
        checked={showNotifications}
        onChange={onShowNotificationsChange}
      />
    </div>
  );
}

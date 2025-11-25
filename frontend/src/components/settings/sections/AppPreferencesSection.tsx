import { SettingItem, SettingToggle } from '../';

export interface AppPreferencesSectionProps {
  autoRefresh: boolean;
  onAutoRefreshChange: (value: boolean) => void;
  refreshInterval: number;
  onRefreshIntervalChange: (value: number) => void;
  showNotifications: boolean;
  onShowNotificationsChange: (value: boolean) => void;
}

export function AppPreferencesSection({
  autoRefresh,
  onAutoRefreshChange,
  refreshInterval,
  onRefreshIntervalChange,
  showNotifications,
  onShowNotificationsChange,
}: AppPreferencesSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">Application Preferences</h2>

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

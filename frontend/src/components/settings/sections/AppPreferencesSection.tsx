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
  // Rotation settings
  rotationNotificationsEnabled: boolean;
  onRotationNotificationsEnabledChange: (value: boolean) => void;
  rotationNotificationThreshold: number;
  onRotationNotificationThresholdChange: (value: number) => void;
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
  rotationNotificationsEnabled,
  onRotationNotificationsEnabledChange,
  rotationNotificationThreshold,
  onRotationNotificationThresholdChange,
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

      <h3 className="subsection-title">Standard Rotation</h3>

      <SettingToggle
        label="Rotation notifications"
        description="Get notified when your Standard decks have cards rotating out"
        checked={rotationNotificationsEnabled}
        onChange={onRotationNotificationsEnabledChange}
      />

      {rotationNotificationsEnabled && (
        <SettingSelect
          label="Notification timing"
          description="When to start showing rotation warnings"
          value={rotationNotificationThreshold.toString()}
          onChange={(value) => onRotationNotificationThresholdChange(parseInt(value))}
          options={[
            { value: '7', label: '7 days before rotation' },
            { value: '30', label: '30 days before rotation' },
            { value: '60', label: '60 days before rotation' },
            { value: '90', label: '90 days before rotation' },
          ]}
        />
      )}
    </div>
  );
}

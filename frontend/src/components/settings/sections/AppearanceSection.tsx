import { SettingSelect } from '../';

export function AppearanceSection() {
  return (
    <div className="settings-section">
      <h2 className="section-title">Appearance</h2>

      <SettingSelect
        label="Theme"
        description="Choose your preferred color scheme"
        value="dark"
        onChange={() => {}}
        options={[
          { value: 'dark', label: 'Dark (Default)' },
          { value: 'light', label: 'Light (Coming Soon)' },
          { value: 'auto', label: 'Auto (System Default)' },
        ]}
      />
    </div>
  );
}

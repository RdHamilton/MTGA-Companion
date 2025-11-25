import { SettingItem } from '../';

export interface ImportExportSectionProps {
  onExportData: (format: 'json' | 'csv') => void;
  onImportData: () => void;
}

export function ImportExportSection({
  onExportData,
  onImportData,
}: ImportExportSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">Import / Export</h2>
      <div className="setting-description settings-section-description">
        Export your data for backup or import previously exported data.
      </div>

      <SettingItem
        label="Export Data"
        description="Export your match history and statistics to a file for backup"
      >
        <button className="action-button" onClick={() => onExportData('json')}>
          Export to JSON
        </button>
        <button className="action-button" onClick={() => onExportData('csv')}>
          Export to CSV
        </button>
      </SettingItem>

      <SettingItem
        label="Import Data"
        description="Import match data from a JSON file exported by this app"
      >
        <button className="action-button" onClick={onImportData}>
          Import from JSON
        </button>
      </SettingItem>
    </div>
  );
}

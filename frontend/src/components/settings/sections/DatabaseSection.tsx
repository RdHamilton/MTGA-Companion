export interface DatabaseSectionProps {
  dbPath: string;
  onDbPathChange: (path: string) => void;
}

export function DatabaseSection({ dbPath, onDbPathChange }: DatabaseSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">Database Configuration</h2>
      <div className="setting-item">
        <label className="setting-label">
          Database Path
          <span className="setting-description">Location of the MTGA Companion database file</span>
        </label>
        <div className="setting-control">
          <input
            type="text"
            value={dbPath}
            onChange={(e) => onDbPathChange(e.target.value)}
            placeholder="/Users/username/.mtga-companion/mtga.db"
            className="text-input"
          />
          <button className="browse-button">Browse...</button>
        </div>
      </div>
    </div>
  );
}

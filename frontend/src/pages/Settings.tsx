import { useState } from 'react';
import AboutDialog from '../components/AboutDialog';
import './Settings.css';

const Settings = () => {
  const [dbPath, setDbPath] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showNotifications, setShowNotifications] = useState(true);
  const [saved, setSaved] = useState(false);
  const [showAbout, setShowAbout] = useState(false);

  const handleSave = () => {
    // TODO: Implement backend settings save
    // For now, just show success message
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const handleReset = () => {
    setDbPath('');
    setAutoRefresh(false);
    setRefreshInterval(30);
    setShowNotifications(true);
  };

  const handleExportData = () => {
    // TODO: Implement data export
    alert('Export functionality coming soon!');
  };

  const handleImportData = () => {
    // TODO: Implement data import
    alert('Import functionality coming soon!');
  };

  return (
    <div className="page-container">
      <div className="settings-header">
        <h1 className="page-title">Settings</h1>
        {saved && <div className="save-notification">Settings saved successfully!</div>}
      </div>

      <div className="settings-content">
        {/* Database Configuration */}
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
                onChange={(e) => setDbPath(e.target.value)}
                placeholder="/Users/username/.mtga-companion/mtga.db"
                className="text-input"
              />
              <button className="browse-button">Browse...</button>
            </div>
          </div>
        </div>

        {/* Application Preferences */}
        <div className="settings-section">
          <h2 className="section-title">Application Preferences</h2>

          <div className="setting-item">
            <label className="setting-label">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="checkbox-input"
              />
              Auto-refresh data
              <span className="setting-description">Automatically refresh statistics at regular intervals</span>
            </label>
          </div>

          {autoRefresh && (
            <div className="setting-item indented">
              <label className="setting-label">
                Refresh Interval (seconds)
                <span className="setting-description">How often to refresh data automatically</span>
              </label>
              <div className="setting-control">
                <input
                  type="number"
                  value={refreshInterval}
                  onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                  min="10"
                  max="300"
                  step="10"
                  className="number-input"
                />
              </div>
            </div>
          )}

          <div className="setting-item">
            <label className="setting-label">
              <input
                type="checkbox"
                checked={showNotifications}
                onChange={(e) => setShowNotifications(e.target.checked)}
                className="checkbox-input"
              />
              Show notifications
              <span className="setting-description">Display notifications for match results and updates</span>
            </label>
          </div>
        </div>

        {/* Data Management */}
        <div className="settings-section">
          <h2 className="section-title">Data Management</h2>

          <div className="setting-item">
            <label className="setting-label">
              Export Data
              <span className="setting-description">Export your match history and statistics to a file</span>
            </label>
            <div className="setting-control">
              <button className="action-button" onClick={handleExportData}>
                Export to JSON
              </button>
              <button className="action-button" onClick={handleExportData}>
                Export to CSV
              </button>
            </div>
          </div>

          <div className="setting-item">
            <label className="setting-label">
              Import Data
              <span className="setting-description">Import match history from a file</span>
            </label>
            <div className="setting-control">
              <button className="action-button" onClick={handleImportData}>
                Import from File
              </button>
            </div>
          </div>

          <div className="setting-item danger">
            <label className="setting-label">
              Clear All Data
              <span className="setting-description">Permanently delete all match history and statistics</span>
            </label>
            <div className="setting-control">
              <button className="danger-button" onClick={() => {
                if (window.confirm('Are you sure you want to delete all data? This cannot be undone.')) {
                  alert('Clear data functionality coming soon!');
                }
              }}>
                Clear All Data
              </button>
            </div>
          </div>
        </div>

        {/* Theme Preferences */}
        <div className="settings-section">
          <h2 className="section-title">Appearance</h2>

          <div className="setting-item">
            <label className="setting-label">
              Theme
              <span className="setting-description">Choose your preferred color scheme</span>
            </label>
            <div className="setting-control">
              <select className="select-input">
                <option value="dark">Dark (Default)</option>
                <option value="light" disabled>Light (Coming Soon)</option>
                <option value="auto" disabled>Auto (System Default)</option>
              </select>
            </div>
          </div>
        </div>

        {/* About */}
        <div className="settings-section">
          <h2 className="section-title">About</h2>

          <div className="about-content">
            <div className="about-item">
              <span className="about-label">Version:</span>
              <span className="about-value">1.0.0</span>
            </div>
            <div className="about-item">
              <span className="about-label">Build:</span>
              <span className="about-value">Development</span>
            </div>
            <div className="about-item">
              <span className="about-label">Platform:</span>
              <span className="about-value">Wails + React</span>
            </div>
            <div className="setting-control" style={{ marginTop: '16px' }}>
              <button className="action-button" onClick={() => setShowAbout(true)}>
                About MTGA Companion
              </button>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="settings-actions">
          <button className="primary-button" onClick={handleSave}>
            Save Settings
          </button>
          <button className="secondary-button" onClick={handleReset}>
            Reset to Defaults
          </button>
        </div>
      </div>

      {/* About Dialog */}
      <AboutDialog isOpen={showAbout} onClose={() => setShowAbout(false)} />
    </div>
  );
};

export default Settings;

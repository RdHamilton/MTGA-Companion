export interface AboutSectionProps {
  onShowAboutDialog: () => void;
  isDeveloperMode: boolean;
  onVersionClick: () => void;
  onToggleDeveloperMode: () => void;
}

export function AboutSection({
  onShowAboutDialog,
  isDeveloperMode,
  onVersionClick,
  onToggleDeveloperMode,
}: AboutSectionProps) {
  return (
    <div className="settings-section">
      <h2 className="section-title">About</h2>

      <div className="about-content">
        <div className="about-item">
          <span className="about-label">Version:</span>
          <span
            className="about-value about-version-clickable"
            onClick={onVersionClick}
            title="Click 5 times to toggle developer mode"
          >
            1.0.0
          </span>
        </div>
        <div className="about-item">
          <span className="about-label">Build:</span>
          <span className="about-value">Development</span>
        </div>
        <div className="about-item">
          <span className="about-label">Platform:</span>
          <span className="about-value">Wails + React</span>
        </div>
        {isDeveloperMode && (
          <div className="about-item developer-mode-indicator">
            <span className="about-label">Developer Mode:</span>
            <span className="about-value developer-mode-enabled">
              Enabled
              <button
                className="developer-mode-toggle"
                onClick={onToggleDeveloperMode}
                title="Disable developer mode"
              >
                Disable
              </button>
            </span>
          </div>
        )}
        <div className="setting-control about-button-container">
          <button className="action-button" onClick={onShowAboutDialog}>
            About MTGA Companion
          </button>
        </div>
      </div>
    </div>
  );
}

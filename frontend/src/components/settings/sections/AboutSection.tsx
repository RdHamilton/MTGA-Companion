export interface AboutSectionProps {
  onShowAboutDialog: () => void;
}

export function AboutSection({ onShowAboutDialog }: AboutSectionProps) {
  return (
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
        <div className="setting-control about-button-container">
          <button className="action-button" onClick={onShowAboutDialog}>
            About MTGA Companion
          </button>
        </div>
      </div>
    </div>
  );
}

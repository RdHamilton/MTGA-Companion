import { useState, useEffect } from 'react';
import { GetAppVersion } from '@/services/api/legacy';
import './AboutDialog.css';

interface AboutDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const AboutDialog = ({ isOpen, onClose }: AboutDialogProps) => {
  const [version, setVersion] = useState('Loading...');

  useEffect(() => {
    if (isOpen) {
      GetAppVersion()
        .then((v) => setVersion(v))
        .catch(() => setVersion('Unknown'));
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content about-dialog" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>About MTGA Companion</h2>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>

        <div className="modal-body">
          <div className="about-section">
            <div className="app-icon">
              {/* Placeholder for app icon */}
              <div className="icon-placeholder">
                <span style={{ fontSize: '48px' }}>🃏</span>
              </div>
            </div>
            <h3 className="app-name">MTGA Companion</h3>
            <p className="app-version">Version {version}</p>
          </div>

          <div className="about-section">
            <h4>About</h4>
            <p>
              MTGA Companion is a desktop application for tracking and analyzing your Magic: The Gathering Arena matches.
              Built to help you improve your game with detailed statistics, trends, and insights.
            </p>
          </div>

          <div className="about-section">
            <h4>Features</h4>
            <ul className="feature-list">
              <li>Real-time match tracking from MTGA logs</li>
              <li>Comprehensive statistics and analytics</li>
              <li>Win rate trends and predictions</li>
              <li>Deck performance analysis</li>
              <li>Rank progression tracking</li>
              <li>Format-specific breakdowns</li>
            </ul>
          </div>

          <div className="about-section">
            <h4>Acknowledgments</h4>
            <p>This project integrates data and services from:</p>
            <ul className="credit-list">
              <li><strong>Scryfall</strong> - Card metadata and imagery</li>
              <li><strong>17Lands</strong> - Draft statistics and card ratings</li>
              <li><strong>Wizards of the Coast</strong> - Magic: The Gathering Arena</li>
            </ul>
          </div>

          <div className="about-section">
            <h4>License</h4>
            <p>
              MTGA Companion is open source software licensed under the MIT License.
            </p>
          </div>

          <div className="about-section about-links">
            <h4>Links</h4>
            <div className="link-buttons">
              <a
                href="https://github.com/RdHamilton/MTGA-Companion"
                target="_blank"
                rel="noopener noreferrer"
                className="link-button"
              >
                GitHub Repository
              </a>
              <a
                href="https://github.com/RdHamilton/MTGA-Companion/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="link-button"
              >
                Report an Issue
              </a>
              <a
                href="https://github.com/RdHamilton/MTGA-Companion/wiki"
                target="_blank"
                rel="noopener noreferrer"
                className="link-button"
              >
                Documentation
              </a>
            </div>
          </div>

          <div className="about-section">
            <p className="copyright">
              © 2024-{new Date().getFullYear()} Ray Hamilton Engineering LLC
            </p>
            <p className="disclaimer">
              This application is not affiliated with or endorsed by Wizards of the Coast.
              Magic: The Gathering Arena is a trademark of Wizards of the Coast LLC.
            </p>
          </div>
        </div>

        <div className="modal-footer">
          <button className="btn btn-primary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default AboutDialog;

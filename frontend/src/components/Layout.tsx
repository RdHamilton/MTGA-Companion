import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import Footer from './Footer';
import './Layout.css';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  const location = useLocation();
  const [activeTab, setActiveTab] = useState<'match-history' | 'charts'>('match-history');

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="app-container">
      {/* Top Navigation Tabs */}
      <div className="tab-bar">
        <Link
          to="/match-history"
          className={`tab ${activeTab === 'match-history' ? 'active' : ''}`}
          onClick={() => setActiveTab('match-history')}
        >
          Match History
        </Link>
        <Link
          to="/charts/win-rate-trend"
          className={`tab ${activeTab === 'charts' ? 'active' : ''}`}
          onClick={() => setActiveTab('charts')}
        >
          Charts
        </Link>
        <Link
          to="/settings"
          className={`tab ${isActive('/settings') ? 'active' : ''}`}
          onClick={() => setActiveTab('match-history')}
        >
          Settings
        </Link>
      </div>

      {/* Sub-navigation for Charts */}
      {activeTab === 'charts' && (
        <div className="sub-tab-bar">
          <Link
            to="/charts/win-rate-trend"
            className={`sub-tab ${isActive('/charts/win-rate-trend') ? 'active' : ''}`}
          >
            Win Rate Trend
          </Link>
          <Link
            to="/charts/deck-performance"
            className={`sub-tab ${isActive('/charts/deck-performance') ? 'active' : ''}`}
          >
            Deck Performance
          </Link>
          <Link
            to="/charts/rank-progression"
            className={`sub-tab ${isActive('/charts/rank-progression') ? 'active' : ''}`}
          >
            Rank Progression
          </Link>
          <Link
            to="/charts/format-distribution"
            className={`sub-tab ${isActive('/charts/format-distribution') ? 'active' : ''}`}
          >
            Format Distribution
          </Link>
          <Link
            to="/charts/result-breakdown"
            className={`sub-tab ${isActive('/charts/result-breakdown') ? 'active' : ''}`}
          >
            Result Breakdown
          </Link>
        </div>
      )}

      {/* Main Content */}
      <div className="content">
        {children}
      </div>

      {/* Footer with Stats */}
      <Footer />
    </div>
  );
};

export default Layout;

import { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import Footer from './Footer';
import { GetConnectionStatus, ResumeReplay, StopReplay } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { getReplayState, subscribeToReplayState } from '../App';
import './Layout.css';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  const location = useLocation();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<'match-history' | 'quests' | 'events' | 'draft' | 'charts'>('match-history');
  const [connectionStatus, setConnectionStatus] = useState<any>({
    status: 'standalone',
    connected: false
  });
  const [replayActive, setReplayActive] = useState(getReplayState().isActive);
  const [replayPaused, setReplayPaused] = useState(getReplayState().isPaused);

  const isActive = (path: string) => location.pathname === path;

  // Sync activeTab with current route (for keyboard shortcuts and direct navigation)
  useEffect(() => {
    if (location.pathname === '/match-history' || location.pathname === '/') {
      setActiveTab('match-history');
    } else if (location.pathname === '/quests') {
      setActiveTab('quests');
    } else if (location.pathname === '/events') {
      setActiveTab('events');
    } else if (location.pathname === '/draft') {
      setActiveTab('draft');
    } else if (location.pathname.startsWith('/charts/')) {
      setActiveTab('charts');
    }
  }, [location.pathname]);

  // Subscribe to replay state changes
  useEffect(() => {
    console.log('[Layout] Subscribing to replay state changes');
    const unsubscribe = subscribeToReplayState((state) => {
      console.log('[Layout] Replay state updated:', state);
      setReplayActive(state.isActive);
      setReplayPaused(state.isPaused);
    });

    return unsubscribe;
  }, []);

  // Load connection status on mount
  useEffect(() => {
    loadConnectionStatus();

    // Listen for daemon events
    EventsOn('daemon:status', () => {
      loadConnectionStatus();
    });

    EventsOn('daemon:connected', () => {
      loadConnectionStatus();
    });

    return () => {
      EventsOff('daemon:status');
      EventsOff('daemon:connected');
    };
  }, []);

  const loadConnectionStatus = async () => {
    try {
      const status = await GetConnectionStatus();
      setConnectionStatus(status);
    } catch (error) {
      console.error('Failed to load connection status:', error);
    }
  };

  const handleResumeReplay = async () => {
    try {
      await ResumeReplay();
    } catch (error) {
      console.error('Failed to resume replay:', error);
    }
  };

  const handleStopReplay = async () => {
    try {
      await StopReplay();
      // Navigate to settings after stopping
      navigate('/settings');
    } catch (error) {
      console.error('Failed to stop replay:', error);
    }
  };

  return (
    <div className="app-container">
      {/* Top Navigation Tabs */}
      <div className="tab-bar">
        <div className="tab-links">
          <Link
            to="/match-history"
            className={`tab ${activeTab === 'match-history' ? 'active' : ''}`}
            onClick={() => setActiveTab('match-history')}
          >
            Match History
          </Link>
          <Link
            to="/quests"
            className={`tab ${activeTab === 'quests' ? 'active' : ''}`}
            onClick={() => setActiveTab('quests')}
          >
            Quests
          </Link>
          <Link
            to="/events"
            className={`tab ${activeTab === 'events' ? 'active' : ''}`}
            onClick={() => setActiveTab('events')}
          >
            Events
          </Link>
          <Link
            to="/draft"
            className={`tab ${activeTab === 'draft' ? 'active' : ''}`}
            onClick={() => setActiveTab('draft')}
          >
            Draft
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
        <div className="connection-status-indicator">
          <div className={`status-badge-compact status-${connectionStatus.status}`} title={connectionStatus.status}>
            <span className="status-dot-compact"></span>
          </div>
        </div>
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

      {/* Floating Replay Control Banner - Only shown when replay is paused and not on settings page */}
      {replayActive && replayPaused && location.pathname !== '/settings' && (
        <div style={{
          position: 'fixed',
          bottom: '60px',
          right: '20px',
          background: '#ff9800',
          color: 'white',
          padding: '16px 24px',
          borderRadius: '8px',
          boxShadow: '0 4px 12px rgba(0,0,0,0.3)',
          zIndex: 1000,
          display: 'flex',
          alignItems: 'center',
          gap: '16px',
          fontWeight: 'bold',
        }}>
          <span>⏸️ Replay Paused</span>
          <button
            onClick={handleResumeReplay}
            style={{
              background: '#00c853',
              color: 'white',
              border: 'none',
              padding: '8px 16px',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            ▶️ Resume
          </button>
          <button
            onClick={handleStopReplay}
            style={{
              background: '#f44336',
              color: 'white',
              border: 'none',
              padding: '8px 16px',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            ⏹️ Stop
          </button>
          <button
            onClick={() => navigate('/settings')}
            style={{
              background: 'transparent',
              color: 'white',
              border: '1px solid white',
              padding: '8px 16px',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Settings
          </button>
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

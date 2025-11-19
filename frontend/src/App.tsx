import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { useEffect, useState } from 'react';
import Layout from './components/Layout';
import ToastContainer from './components/ToastContainer';
import MatchHistory from './pages/MatchHistory';
import WinRateTrend from './pages/WinRateTrend';
import DeckPerformance from './pages/DeckPerformance';
import RankProgression from './pages/RankProgression';
import FormatDistribution from './pages/FormatDistribution';
import ResultBreakdown from './pages/ResultBreakdown';
import Quests from './pages/Quests';
import Events from './pages/Events';
import Draft from './pages/Draft';
import Settings from './pages/Settings';
import KeyboardShortcutsHandler from './components/KeyboardShortcutsHandler';
import { EventsOn } from '../wailsjs/runtime/runtime';
import './App.css';

// Global replay state context
export interface ReplayState {
  isActive: boolean;
  isPaused: boolean;
  progress: any;
}

const initialReplayState: ReplayState = {
  isActive: false,
  isPaused: false,
  progress: null,
};

// Global replay state - accessible across all components
let globalReplayState: ReplayState = { ...initialReplayState };
const replayStateListeners: Array<(state: ReplayState) => void> = [];

export const getReplayState = (): ReplayState => globalReplayState;

export const subscribeToReplayState = (listener: (state: ReplayState) => void): (() => void) => {
  replayStateListeners.push(listener);
  return () => {
    const index = replayStateListeners.indexOf(listener);
    if (index > -1) {
      replayStateListeners.splice(index, 1);
    }
  };
};

const updateReplayState = (updates: Partial<ReplayState>) => {
  globalReplayState = { ...globalReplayState, ...updates };
  console.log('[Global Replay State] Updated:', globalReplayState, 'Listeners:', replayStateListeners.length);
  replayStateListeners.forEach(listener => listener(globalReplayState));
};

// Component that handles global replay events
function ReplayEventHandler() {
  const navigate = useNavigate();
  const [hasShownDraftNotification, setHasShownDraftNotification] = useState(false);

  useEffect(() => {
    console.log('[ReplayEventHandler] Setting up global replay event listeners');

    // Listen for replay events and update global state
    const unsubscribeStarted = EventsOn('replay:started', (data: any) => {
      console.log('[ReplayEventHandler] Replay started:', data);
      updateReplayState({
        isActive: true,
        isPaused: false,
        progress: data,
      });
      setHasShownDraftNotification(false);
    });

    const unsubscribeProgress = EventsOn('replay:progress', (data: any) => {
      console.log('[ReplayEventHandler] Replay progress:', data);
      updateReplayState({
        progress: data,
      });
    });

    const unsubscribePaused = EventsOn('replay:paused', (data: any) => {
      console.log('[ReplayEventHandler] ✅✅✅ Replay paused EVENT RECEIVED:', data);
      console.log('[ReplayEventHandler] About to update state to isPaused=true');
      updateReplayState({
        isPaused: true,
      });
      console.log('[ReplayEventHandler] State update called');
    });

    const unsubscribeResumed = EventsOn('replay:resumed', (data: any) => {
      console.log('[ReplayEventHandler] Replay resumed:', data);
      updateReplayState({
        isPaused: false,
      });
    });

    const unsubscribeCompleted = EventsOn('replay:completed', (data: any) => {
      console.log('[ReplayEventHandler] Replay completed:', data);
      updateReplayState({
        isActive: false,
        isPaused: false,
        progress: data,
      });
      setHasShownDraftNotification(false);
    });

    const unsubscribeDraftDetected = EventsOn('replay:draft_detected', (data: any) => {
      console.log('[ReplayEventHandler] Draft detected during replay:', data);

      // Automatically navigate to Draft tab
      navigate('/draft');

      // Show notification only once per replay session
      if (!hasShownDraftNotification) {
        // We'll use a console log for now since alerts don't work in desktop mode
        // The toast system will handle the notification
        console.log('Draft event detected - navigated to Draft tab!');
        setHasShownDraftNotification(true);
      }
    });

    const unsubscribeError = EventsOn('replay:error', (data: any) => {
      console.error('[ReplayEventHandler] Replay error:', data);
      updateReplayState({
        isActive: false,
        isPaused: false,
      });
    });

    return () => {
      console.log('[ReplayEventHandler] Cleaning up global replay event listeners');
      unsubscribeStarted();
      unsubscribeProgress();
      unsubscribePaused();
      unsubscribeResumed();
      unsubscribeCompleted();
      unsubscribeDraftDetected();
      unsubscribeError();
    };
  }, [navigate, hasShownDraftNotification]);

  return null; // This component doesn't render anything
}

function App() {
  return (
    <Router>
      <ReplayEventHandler />
      <KeyboardShortcutsHandler />
      <Layout>
        <Routes>
          <Route path="/" element={<Navigate to="/match-history" replace />} />
          <Route path="/match-history" element={<MatchHistory />} />
          <Route path="/quests" element={<Quests />} />
          <Route path="/events" element={<Events />} />
          <Route path="/draft" element={<Draft />} />
          <Route path="/charts/win-rate-trend" element={<WinRateTrend />} />
          <Route path="/charts/deck-performance" element={<DeckPerformance />} />
          <Route path="/charts/rank-progression" element={<RankProgression />} />
          <Route path="/charts/format-distribution" element={<FormatDistribution />} />
          <Route path="/charts/result-breakdown" element={<ResultBreakdown />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </Layout>
      <ToastContainer />
    </Router>
  );
}

export default App;

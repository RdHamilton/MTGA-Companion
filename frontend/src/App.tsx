import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
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
import Settings from './pages/Settings';
import './App.css';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Navigate to="/match-history" replace />} />
          <Route path="/match-history" element={<MatchHistory />} />
          <Route path="/quests" element={<Quests />} />
          <Route path="/events" element={<Events />} />
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

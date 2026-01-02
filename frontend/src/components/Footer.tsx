import { useState, useEffect } from 'react';
import { EventsOn } from '@/services/websocketClient';
import { matches } from '@/services/api';
import { models } from '@/types/models';
import DownloadProgressBar from './DownloadProgressBar';
import './Footer.css';

const Footer = () => {
  const [stats, setStats] = useState<models.Statistics | null>(null);
  const [streak, setStreak] = useState<{ type: string; count: number }>({ type: '', count: 0 });
  const [lastMatch, setLastMatch] = useState<string>('');
  const [lastSynced, setLastSynced] = useState<string>('');
  const [loading, setLoading] = useState(true);

  const loadStats = async () => {
    try {
      // Get overall stats
      const filter = new models.StatsFilter();
      const statsData = await matches.getStats(matches.statsFilterToRequest(filter));
      setStats(statsData);

      // Get recent matches to calculate streak and last match time
      const matchData = await matches.getMatches(matches.statsFilterToRequest(filter));

      if (matchData && matchData.length > 0) {
        // Calculate current streak
        const lastResult = matchData[0].Result;
        let streakCount = 1;

        for (let i = 1; i < matchData.length; i++) {
          if (matchData[i].Result === lastResult) {
            streakCount++;
          } else {
            break;
          }
        }

        setStreak({
          type: lastResult === 'win' ? 'W' : 'L',
          count: streakCount
        });

        // Format last match time
        const lastMatchDate = new Date(matchData[0].Timestamp as string);
        setLastMatch(lastMatchDate.toLocaleString());
      }

      // Update last synced time
      setLastSynced(new Date().toLocaleTimeString());
    } catch (err) {
      console.error('Error loading footer stats:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStats();
  }, []);

  // Listen for real-time updates
  useEffect(() => {
    const unsubscribe = EventsOn('stats:updated', () => {
      console.log('Stats updated event received - reloading footer');
      loadStats();
    });

    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, []);

  if (loading) {
    return (
      <footer className="app-footer">
        <div className="footer-content">
          <span className="footer-loading">Loading stats...</span>
        </div>
      </footer>
    );
  }

  if (!stats || stats.TotalMatches === 0) {
    return (
      <footer className="app-footer">
        <div className="footer-content">
          <span className="footer-empty">No matches yet - play some games to see your stats!</span>
        </div>
      </footer>
    );
  }

  return (
    <footer className="app-footer">
      <div className="footer-content">
        <span className="footer-label">All Time</span>
        <span className="footer-separator">|</span>
        <span className="footer-stat">
          <strong>Matches:</strong> {stats.TotalMatches}
        </span>
        <span className="footer-separator">|</span>
        <span className="footer-stat">
          <strong>Win Rate:</strong> {Math.round(stats.WinRate * 1000) / 10}% ({stats.MatchesWon}-{stats.MatchesLost})
        </span>
        {streak.count > 0 && (
          <>
            <span className="footer-separator">|</span>
            <span className={`footer-stat streak-${streak.type.toLowerCase()}`}>
              <strong>Streak:</strong> {streak.type}{streak.count}
            </span>
          </>
        )}
        {lastMatch && (
          <>
            <span className="footer-separator">|</span>
            <span className="footer-stat footer-last-match">
              <strong>Last Played:</strong> {lastMatch}
            </span>
          </>
        )}
        {lastSynced && (
          <>
            <span className="footer-separator">|</span>
            <span className="footer-stat footer-last-synced">
              <strong>Synced:</strong> {lastSynced}
            </span>
          </>
        )}
        <DownloadProgressBar />
      </div>
    </footer>
  );
};

export default Footer;

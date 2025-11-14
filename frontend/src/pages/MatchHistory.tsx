import { useState, useEffect } from 'react';
import { GetMatches } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import './MatchHistory.css';

const MatchHistory = () => {
  const [matches, setMatches] = useState<models.Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [dateRange, setDateRange] = useState('7days');
  const [format, setFormat] = useState('all');
  const [result, setResult] = useState('all');
  const [opponent, setOpponent] = useState('');

  useEffect(() => {
    loadMatches();
  }, [dateRange, format, result, opponent]);

  const loadMatches = async () => {
    try {
      setLoading(true);
      setError(null);

      // Build filter
      const filter = new models.StatsFilter();

      // Date range
      if (dateRange !== 'all') {
        const now = new Date();
        const start = new Date();

        switch (dateRange) {
          case '7days':
            start.setDate(now.getDate() - 7);
            break;
          case '30days':
            start.setDate(now.getDate() - 30);
            break;
          case '90days':
            start.setDate(now.getDate() - 90);
            break;
        }

        filter.StartDate = start.toISOString().split('T')[0];
        filter.EndDate = now.toISOString().split('T')[0];
      }

      // Format filter
      if (format !== 'all') {
        if (format === 'constructed') {
          filter.Formats = ['Ladder', 'Play'];
        } else if (format === 'limited') {
          // Limited formats contain 'Draft' or 'Sealed'
          filter.Format = ''; // Backend will need to handle this specially
        } else {
          filter.Format = format;
        }
      }

      // Result filter
      if (result !== 'all') {
        filter.Result = result;
      }

      // Opponent filter
      if (opponent.trim()) {
        filter.OpponentName = opponent.trim();
      }

      const matchData = await GetMatches(filter);
      setMatches(matchData || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load matches');
      console.error('Error loading matches:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (timestamp: any) => {
    return new Date(timestamp).toLocaleString();
  };

  const formatScore = (wins: number, losses: number) => {
    return `${wins}-${losses}`;
  };

  const getRankDisplay = (rankBefore: string | undefined) => {
    if (!rankBefore) return '-';
    // Parse "Bronze_1" format
    const parts = rankBefore.split('_');
    if (parts.length === 2) {
      return `${parts[0]} ${parts[1]}`;
    }
    return rankBefore;
  };

  return (
    <div className="page-container">
      <h1 className="page-title">Match History</h1>

      {/* Filters */}
      <div className="filter-row">
        <div className="filter-group">
          <label className="filter-label">Date Range</label>
          <select value={dateRange} onChange={(e) => setDateRange(e.target.value)}>
            <option value="7days">Last 7 Days</option>
            <option value="30days">Last 30 Days</option>
            <option value="90days">Last 90 Days</option>
            <option value="all">All Time</option>
          </select>
        </div>

        <div className="filter-group">
          <label className="filter-label">Format</label>
          <select value={format} onChange={(e) => setFormat(e.target.value)}>
            <option value="all">All Formats</option>
            <option value="constructed">Constructed</option>
            <option value="limited">Limited</option>
            <option value="Ladder">Ranked (Ladder)</option>
            <option value="Play">Play Queue</option>
          </select>
        </div>

        <div className="filter-group">
          <label className="filter-label">Result</label>
          <select value={result} onChange={(e) => setResult(e.target.value)}>
            <option value="all">All Results</option>
            <option value="win">Wins Only</option>
            <option value="loss">Losses Only</option>
          </select>
        </div>

        <div className="filter-group">
          <label className="filter-label">Opponent</label>
          <input
            type="text"
            placeholder="Search opponent..."
            value={opponent}
            onChange={(e) => setOpponent(e.target.value)}
          />
        </div>
      </div>

      {/* Content */}
      {loading && <div className="no-data">Loading matches...</div>}

      {error && <div className="error">{error}</div>}

      {!loading && !error && matches.length === 0 && (
        <div className="no-data">No matches found for the selected filters</div>
      )}

      {!loading && !error && matches.length > 0 && (
        <>
          <div className="match-count">
            Showing {matches.length} match{matches.length !== 1 ? 'es' : ''}
          </div>

          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Result</th>
                <th>Format</th>
                <th>Event</th>
                <th>Opponent</th>
                <th>Score</th>
                <th>Rank</th>
              </tr>
            </thead>
            <tbody>
              {matches.map((match) => (
                <tr key={match.ID} className={`result-${match.Result.toLowerCase()}`}>
                  <td>{formatTimestamp(match.Timestamp)}</td>
                  <td>
                    <span className={`result-badge ${match.Result.toLowerCase()}`}>
                      {match.Result.toUpperCase()}
                    </span>
                  </td>
                  <td>{match.Format}</td>
                  <td>{match.EventName}</td>
                  <td>{match.OpponentName || 'Unknown'}</td>
                  <td>{formatScore(match.PlayerWins, match.OpponentWins)}</td>
                  <td>{getRankDisplay(match.RankBefore)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  );
};

export default MatchHistory;

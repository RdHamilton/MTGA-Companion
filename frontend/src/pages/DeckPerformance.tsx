import { useState, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetStatsByDeck } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import './DeckPerformance.css';

interface DeckStats {
  deckName: string;
  stats: models.Statistics;
}

const DeckPerformance = () => {
  const [deckStats, setDeckStats] = useState<DeckStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [dateRange, setDateRange] = useState('7days');
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');
  const [format, setFormat] = useState('all');
  const [sortBy, setSortBy] = useState<'winRate' | 'matches' | 'name'>('winRate');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

  useEffect(() => {
    loadDeckStats();
  }, [dateRange, customStartDate, customEndDate, format]);

  // Listen for real-time updates
  useEffect(() => {
    const unsubscribe = EventsOn('stats:updated', () => {
      console.log('Stats updated event received - reloading deck performance data');
      loadDeckStats();
    });

    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, [dateRange, customStartDate, customEndDate, format]);

  const loadDeckStats = async () => {
    try {
      setLoading(true);
      setError(null);

      const filter = new models.StatsFilter();

      // Date range
      if (dateRange === 'custom') {
        if (customStartDate) {
          const start = new Date(customStartDate + 'T00:00:00');
          filter.StartDate = start;
        }
        if (customEndDate) {
          const end = new Date(customEndDate + 'T00:00:00');
          end.setDate(end.getDate() + 1);
          filter.EndDate = end;
        }
      } else if (dateRange !== 'all') {
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

        start.setHours(0, 0, 0, 0);
        const end = new Date(now);
        end.setDate(end.getDate() + 1);
        end.setHours(0, 0, 0, 0);

        filter.StartDate = start;
        filter.EndDate = end;
      }

      // Format filter
      if (format !== 'all') {
        if (format === 'constructed') {
          filter.Formats = ['Ladder', 'Play'];
        } else {
          filter.Format = format;
        }
      }

      const data = await GetStatsByDeck(filter);

      // Convert map to array
      const statsArray: DeckStats[] = Object.entries(data || {}).map(([deckName, stats]) => ({
        deckName,
        stats
      }));

      setDeckStats(statsArray);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load deck statistics');
      console.error('Error loading deck stats:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatWinRate = (winRate: number) => {
    return `${Math.round(winRate * 100 * 10) / 10}%`;
  };

  const getTodayDateString = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  const getMinEndDate = () => {
    return customStartDate || undefined;
  };

  // Sort deck stats
  const sortedDecks = [...deckStats].sort((a, b) => {
    let aVal: number | string = 0;
    let bVal: number | string = 0;

    switch (sortBy) {
      case 'winRate':
        aVal = a.stats.WinRate;
        bVal = b.stats.WinRate;
        break;
      case 'matches':
        aVal = a.stats.TotalMatches;
        bVal = b.stats.TotalMatches;
        break;
      case 'name':
        aVal = a.deckName.toLowerCase();
        bVal = b.deckName.toLowerCase();
        break;
    }

    if (sortDirection === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
    }
  });

  return (
    <div className="page-container">
      <div className="deck-performance-header">
        <h1 className="page-title">Deck Performance</h1>

        {/* Filters */}
        <div className="filter-row">
          <div className="filter-group">
            <label className="filter-label">Date Range</label>
            <select value={dateRange} onChange={(e) => setDateRange(e.target.value)}>
              <option value="7days">Last 7 Days</option>
              <option value="30days">Last 30 Days</option>
              <option value="90days">Last 90 Days</option>
              <option value="all">All Time</option>
              <option value="custom">Custom Range</option>
            </select>
          </div>

          {dateRange === 'custom' && (
            <>
              <div className="filter-group">
                <label className="filter-label">Start Date</label>
                <input
                  type="date"
                  value={customStartDate}
                  max={getTodayDateString()}
                  onChange={(e) => setCustomStartDate(e.target.value)}
                />
              </div>

              <div className="filter-group">
                <label className="filter-label">End Date</label>
                <input
                  type="date"
                  value={customEndDate}
                  min={getMinEndDate()}
                  max={getTodayDateString()}
                  onChange={(e) => setCustomEndDate(e.target.value)}
                />
              </div>
            </>
          )}

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
            <label className="filter-label">Sort By</label>
            <select value={sortBy} onChange={(e) => setSortBy(e.target.value as any)}>
              <option value="winRate">Win Rate</option>
              <option value="matches">Match Count</option>
              <option value="name">Deck Name</option>
            </select>
          </div>

          <div className="filter-group">
            <label className="filter-label">Sort Order</label>
            <select value={sortDirection} onChange={(e) => setSortDirection(e.target.value as any)}>
              <option value="desc">Descending</option>
              <option value="asc">Ascending</option>
            </select>
          </div>
        </div>

        {!loading && !error && deckStats.length > 0 && (
          <div className="deck-count">
            {deckStats.length} deck{deckStats.length !== 1 ? 's' : ''} found
          </div>
        )}
      </div>

      {/* Content */}
      {loading && <div className="no-data">Loading deck statistics...</div>}

      {error && <div className="error">{error}</div>}

      {!loading && !error && deckStats.length === 0 && (
        <div className="no-data">No deck data found for the selected filters</div>
      )}

      {!loading && !error && sortedDecks.length > 0 && (
        <div className="deck-grid">
          {sortedDecks.map((deck) => (
            <div key={deck.deckName} className="deck-card">
              <h3 className="deck-name">{deck.deckName || 'Unknown Deck'}</h3>
              <div className="deck-stats">
                <div className="stat">
                  <span className="stat-label">Win Rate</span>
                  <span className="stat-value win-rate">{formatWinRate(deck.stats.WinRate)}</span>
                </div>
                <div className="stat">
                  <span className="stat-label">Matches</span>
                  <span className="stat-value">{deck.stats.TotalMatches}</span>
                </div>
                <div className="stat">
                  <span className="stat-label">Wins / Losses</span>
                  <span className="stat-value">{deck.stats.MatchesWon}W - {deck.stats.MatchesLost}L</span>
                </div>
                <div className="stat">
                  <span className="stat-label">Game Win Rate</span>
                  <span className="stat-value">{formatWinRate(deck.stats.GameWinRate)}</span>
                </div>
                <div className="stat">
                  <span className="stat-label">Games</span>
                  <span className="stat-value">{deck.stats.TotalGames} ({deck.stats.GamesWon}W - {deck.stats.GamesLost}L)</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default DeckPerformance;

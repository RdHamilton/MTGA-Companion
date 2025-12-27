import { useState, useEffect } from 'react';
import { matches } from '@/services/api';
import { models } from '@/types/models';
import LoadingSpinner from '../components/LoadingSpinner';
import EmptyState from '../components/EmptyState';
import ErrorState from '../components/ErrorState';
import { useAppContext } from '../context/AppContext';
import './ResultBreakdown.css';

const ResultBreakdown = () => {
  const { filters, updateFilters } = useAppContext();
  const { dateRange, customStartDate, customEndDate, format } = filters.resultBreakdown;

  const [metrics, setMetrics] = useState<models.Statistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadMetrics = async () => {
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

      const data = await matches.getStats(matches.statsFilterToRequest(filter));
      setMetrics(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load performance metrics');
      console.error('Error loading metrics:', err);
    } finally {
      setLoading(false);
    }
  };

    loadMetrics();
  }, [dateRange, customStartDate, customEndDate, format]);

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

  const getWinRateClass = (winRate: number) => {
    if (winRate >= 0.55) return 'excellent';
    if (winRate >= 0.50) return 'good';
    if (winRate >= 0.45) return 'average';
    return 'below-average';
  };

  return (
    <div className="page-container">
      <div className="result-breakdown-header">
        <h1 className="page-title">Result Breakdown</h1>

        {/* Filters */}
        <div className="filter-row">
          <div className="filter-group">
            <label className="filter-label">Date Range</label>
            <select value={dateRange} onChange={(e) => updateFilters('resultBreakdown', { dateRange: e.target.value })}>
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
                  onChange={(e) => updateFilters('resultBreakdown', { customStartDate: e.target.value })}
                />
              </div>

              <div className="filter-group">
                <label className="filter-label">End Date</label>
                <input
                  type="date"
                  value={customEndDate}
                  min={getMinEndDate()}
                  max={getTodayDateString()}
                  onChange={(e) => updateFilters('resultBreakdown', { customEndDate: e.target.value })}
                />
              </div>
            </>
          )}

          <div className="filter-group">
            <label className="filter-label">Format</label>
            <select value={format} onChange={(e) => updateFilters('resultBreakdown', { format: e.target.value })}>
              <option value="all">All Formats</option>
              <option value="constructed">Constructed</option>
              <option value="limited">Limited</option>
              <option value="Ladder">Ranked (Ladder)</option>
              <option value="Play">Play Queue</option>
            </select>
          </div>
        </div>
      </div>

      {/* Content */}
      {loading && <LoadingSpinner message="Loading performance metrics..." />}

      {error && (
        <ErrorState
          message="Failed to load performance metrics"
          error={error}
          helpText="Check that the database is accessible and try refreshing the page."
        />
      )}

      {!loading && !error && !metrics && (
        <EmptyState
          icon="📊"
          title="No performance data"
          message="Play some matches to see your detailed performance breakdown."
          helpText="Statistics include win rates, match outcomes, and performance trends."
        />
      )}

      {!loading && !error && metrics && (
        <div className="metrics-container">
          {/* Overall Performance */}
          <div className="metric-section">
            <h2 className="section-title">Overall Performance</h2>
            <div className="metric-grid">
              <div className="metric-card highlight">
                <div className="metric-label">Overall Win Rate</div>
                <div className={`metric-value large ${getWinRateClass(metrics.WinRate)}`}>
                  {formatWinRate(metrics.WinRate)}
                </div>
                <div className="metric-sublabel">
                  {metrics.MatchesWon}W - {metrics.MatchesLost}L
                </div>
              </div>

              <div className="metric-card">
                <div className="metric-label">Total Matches</div>
                <div className="metric-value">{metrics.TotalMatches}</div>
                <div className="metric-sublabel">Played</div>
              </div>

              <div className="metric-card">
                <div className="metric-label">Matches Won</div>
                <div className="metric-value win">{metrics.MatchesWon}</div>
                <div className="metric-sublabel">Victories</div>
              </div>

              <div className="metric-card">
                <div className="metric-label">Matches Lost</div>
                <div className="metric-value loss">{metrics.MatchesLost}</div>
                <div className="metric-sublabel">Defeats</div>
              </div>
            </div>
          </div>

          {/* Game-Level Performance */}
          <div className="metric-section">
            <h2 className="section-title">Game-Level Statistics</h2>
            <div className="metric-grid">
              <div className="metric-card highlight">
                <div className="metric-label">Game Win Rate</div>
                <div className={`metric-value large ${getWinRateClass(metrics.GameWinRate)}`}>
                  {formatWinRate(metrics.GameWinRate)}
                </div>
                <div className="metric-sublabel">
                  {metrics.GamesWon}W - {metrics.GamesLost}L
                </div>
              </div>

              <div className="metric-card">
                <div className="metric-label">Total Games</div>
                <div className="metric-value">{metrics.TotalGames}</div>
                <div className="metric-sublabel">Played</div>
              </div>

              <div className="metric-card">
                <div className="metric-label">Games Won</div>
                <div className="metric-value win">{metrics.GamesWon}</div>
                <div className="metric-sublabel">Victories</div>
              </div>

              <div className="metric-card">
                <div className="metric-label">Games Lost</div>
                <div className="metric-value loss">{metrics.GamesLost}</div>
                <div className="metric-sublabel">Defeats</div>
              </div>
            </div>
          </div>

          {/* Performance Analysis */}
          <div className="metric-section">
            <h2 className="section-title">Performance Analysis</h2>
            <div className="analysis-grid">
              <div className="analysis-card">
                <div className="analysis-label">Average Games per Match</div>
                <div className="analysis-value">
                  {metrics.TotalMatches > 0
                    ? (metrics.TotalGames / metrics.TotalMatches).toFixed(2)
                    : '0.00'
                  }
                </div>
              </div>

              <div className="analysis-card">
                <div className="analysis-label">Match to Game Win Rate Ratio</div>
                <div className="analysis-value">
                  {metrics.GameWinRate > 0
                    ? (metrics.WinRate / metrics.GameWinRate).toFixed(2)
                    : '0.00'
                  }
                </div>
              </div>

              <div className="analysis-card">
                <div className="analysis-label">Performance Category</div>
                <div className={`analysis-value ${getWinRateClass(metrics.WinRate)}`}>
                  {metrics.WinRate >= 0.55 ? 'Excellent' :
                   metrics.WinRate >= 0.50 ? 'Good' :
                   metrics.WinRate >= 0.45 ? 'Average' : 'Below Average'}
                </div>
              </div>
            </div>
          </div>

          {/* Win/Loss Breakdown */}
          <div className="metric-section">
            <h2 className="section-title">Win/Loss Breakdown</h2>
            <div className="breakdown-container">
              <div className="breakdown-bar">
                <div
                  className="breakdown-segment wins"
                  style={{ width: `${metrics.WinRate * 100}%` }}
                >
                  <span className="breakdown-label">
                    {formatWinRate(metrics.WinRate)} Wins
                  </span>
                </div>
                <div
                  className="breakdown-segment losses"
                  style={{ width: `${(1 - metrics.WinRate) * 100}%` }}
                >
                  <span className="breakdown-label">
                    {formatWinRate(1 - metrics.WinRate)} Losses
                  </span>
                </div>
              </div>
              <div className="breakdown-stats">
                <div className="breakdown-stat">
                  <span className="stat-dot wins"></span>
                  <span>{metrics.MatchesWon} Matches Won</span>
                </div>
                <div className="breakdown-stat">
                  <span className="stat-dot losses"></span>
                  <span>{metrics.MatchesLost} Matches Lost</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ResultBreakdown;

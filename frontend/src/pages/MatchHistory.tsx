import { useState, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetMatches } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import LoadingSpinner from '../components/LoadingSpinner';
import Tooltip from '../components/Tooltip';
import { useAppContext } from '../context/AppContext';
import './MatchHistory.css';

type SortField = 'Timestamp' | 'Result' | 'Format' | 'EventName';
type SortDirection = 'asc' | 'desc';

const MatchHistory = () => {
  const { filters, updateFilters } = useAppContext();
  const { dateRange, customStartDate, customEndDate, format, result } = filters.matchHistory;

  const [matches, setMatches] = useState<models.Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Pagination
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);

  // Sorting
  const [sortField, setSortField] = useState<SortField>('Timestamp');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  useEffect(() => {
    loadMatches();
  }, [dateRange, customStartDate, customEndDate, format, result]);

  // Listen for real-time updates
  useEffect(() => {
    const unsubscribe = EventsOn('stats:updated', () => {
      console.log('Stats updated event received - reloading match history');
      loadMatches();
    });

    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, [dateRange, customStartDate, customEndDate, format, result]);

  const loadMatches = async () => {
    try {
      setLoading(true);
      setError(null);

      // Build filter
      const filter = new models.StatsFilter();

      // Date range
      if (dateRange === 'custom') {
        // Use custom date range if provided
        if (customStartDate) {
          const start = new Date(customStartDate + 'T00:00:00');
          filter.StartDate = start;
        }
        if (customEndDate) {
          // Add 1 day to end date to make it inclusive
          // (e.g., end date "2024-11-14" becomes "2024-11-15T00:00:00")
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

        // Set start time to beginning of day
        start.setHours(0, 0, 0, 0);
        // Add 1 day to end date to make it inclusive (beginning of next day)
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

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      // Toggle direction
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // New field, default to descending
      setSortField(field);
      setSortDirection('desc');
    }
    setPage(1); // Reset to first page when sorting changes
  };

  // Sort and paginate matches
  const sortedMatches = [...matches].sort((a, b) => {
    let aVal: any = a[sortField];
    let bVal: any = b[sortField];

    // Handle timestamp
    if (sortField === 'Timestamp') {
      aVal = new Date(aVal).getTime();
      bVal = new Date(bVal).getTime();
    }

    // Handle nulls/undefined
    if (aVal == null) return 1;
    if (bVal == null) return -1;

    // String comparison
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      aVal = aVal.toLowerCase();
      bVal = bVal.toLowerCase();
    }

    if (sortDirection === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
    }
  });

  const totalPages = Math.ceil(sortedMatches.length / pageSize);
  const paginatedMatches = sortedMatches.slice((page - 1) * pageSize, page * pageSize);

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) return '⇅';
    return sortDirection === 'asc' ? '↑' : '↓';
  };

  // Get today's date in YYYY-MM-DD format for max date constraint
  const getTodayDateString = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  // Get min date for end date (must be >= start date)
  const getMinEndDate = () => {
    return customStartDate || undefined;
  };

  return (
    <div className="page-container">
      {/* Header Section - Fixed */}
      <div className="match-history-header">
        <h1 className="page-title">Match History</h1>

        {/* Filters */}
        <div className="filter-row">
          <div className="filter-group">
            <label className="filter-label">Date Range</label>
            <select value={dateRange} onChange={(e) => updateFilters('matchHistory', { dateRange: e.target.value })}>
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
                  onChange={(e) => updateFilters('matchHistory', { customStartDate: e.target.value })}
                />
              </div>

              <div className="filter-group">
                <label className="filter-label">End Date</label>
                <input
                  type="date"
                  value={customEndDate}
                  min={getMinEndDate()}
                  max={getTodayDateString()}
                  onChange={(e) => updateFilters('matchHistory', { customEndDate: e.target.value })}
                />
              </div>
            </>
          )}

          <div className="filter-group">
            <label className="filter-label">Format</label>
            <select value={format} onChange={(e) => updateFilters('matchHistory', { format: e.target.value })}>
              <option value="all">All Formats</option>
              <option value="constructed">Constructed</option>
              <option value="limited">Limited</option>
              <option value="Ladder">Ranked (Ladder)</option>
              <option value="Play">Play Queue</option>
            </select>
          </div>

          <div className="filter-group">
            <label className="filter-label">Result</label>
            <select value={result} onChange={(e) => updateFilters('matchHistory', { result: e.target.value })}>
              <option value="all">All Results</option>
              <option value="win">Wins Only</option>
              <option value="loss">Losses Only</option>
            </select>
          </div>
        </div>

        {!loading && !error && matches.length > 0 && (
          <div className="match-count">
            Showing {paginatedMatches.length} of {matches.length} match{matches.length !== 1 ? 'es' : ''}
            {totalPages > 1 && ` (Page ${page} of ${totalPages})`}
          </div>
        )}
      </div>

      {/* Content - Loading/Error/Empty States */}
      {loading && <LoadingSpinner message="Loading matches..." />}

      {error && <div className="error">{error}</div>}

      {!loading && !error && matches.length === 0 && (
        <div className="no-data">No matches found for the selected filters</div>
      )}

      {/* Table Container - Scrollable */}
      {!loading && !error && matches.length > 0 && (
        <>
          <div className="match-history-table-container">
            <table>
            <thead>
              <tr>
                <th onClick={() => handleSort('Timestamp')} style={{ cursor: 'pointer' }}>
                  <Tooltip content="Click to sort by match time">
                    <span>Time {getSortIcon('Timestamp')}</span>
                  </Tooltip>
                </th>
                <th onClick={() => handleSort('Result')} style={{ cursor: 'pointer' }}>
                  <Tooltip content="Click to sort by win/loss">
                    <span>Result {getSortIcon('Result')}</span>
                  </Tooltip>
                </th>
                <th onClick={() => handleSort('Format')} style={{ cursor: 'pointer' }}>
                  <Tooltip content="Click to sort by game format">
                    <span>Format {getSortIcon('Format')}</span>
                  </Tooltip>
                </th>
                <th onClick={() => handleSort('EventName')} style={{ cursor: 'pointer' }}>
                  <Tooltip content="Click to sort by event name">
                    <span>Event {getSortIcon('EventName')}</span>
                  </Tooltip>
                </th>
                <th>
                  <Tooltip content="Match score (Your wins - Opponent wins)">
                    <span>Score</span>
                  </Tooltip>
                </th>
                <th>
                  <Tooltip content="Opponent player name">
                    <span>Opponent</span>
                  </Tooltip>
                </th>
              </tr>
            </thead>
            <tbody>
              {paginatedMatches.map((match) => (
                <tr key={match.ID} className={`result-${match.Result.toLowerCase()}`}>
                  <td>{formatTimestamp(match.Timestamp)}</td>
                  <td>
                    <span className={`result-badge ${match.Result.toLowerCase()}`}>
                      {match.Result.toUpperCase()}
                    </span>
                  </td>
                  <td>{match.Format}</td>
                  <td>{match.EventName}</td>
                  <td>{formatScore(match.PlayerWins, match.OpponentWins)}</td>
                  <td>{match.OpponentName || '—'}</td>
                </tr>
              ))}
            </tbody>
            </table>
          </div>

          {/* Footer Section - Fixed Pagination */}
          {totalPages > 1 && (
            <div className="match-history-footer">
              <div className="pagination">
              <button
                onClick={() => setPage(1)}
                disabled={page === 1}
                className="pagination-btn"
              >
                First
              </button>
              <button
                onClick={() => setPage(page - 1)}
                disabled={page === 1}
                className="pagination-btn"
              >
                Previous
              </button>
              <span className="pagination-info">
                Page {page} of {totalPages}
              </span>
              <button
                onClick={() => setPage(page + 1)}
                disabled={page === totalPages}
                className="pagination-btn"
              >
                Next
              </button>
              <button
                onClick={() => setPage(totalPages)}
                disabled={page === totalPages}
                className="pagination-btn"
              >
                Last
              </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default MatchHistory;

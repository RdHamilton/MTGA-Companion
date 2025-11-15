import { useState, useEffect } from 'react';
import { PieChart, Pie, BarChart, Bar, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { GetStatsByFormat } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import LoadingSpinner from '../components/LoadingSpinner';
import { useAppContext } from '../context/AppContext';
import './FormatDistribution.css';

interface FormatStats {
  format: string;
  stats: models.Statistics;
}

const COLORS = ['#4a9eff', '#7dff7d', '#ff7d7d', '#ffaa00', '#aa00ff', '#00ffaa', '#ff00aa'];

const FormatDistribution = () => {
  const { filters, updateFilters } = useAppContext();
  const { dateRange, customStartDate, customEndDate, chartType, sortBy, sortDirection } = filters.formatDistribution;

  const [formatStats, setFormatStats] = useState<FormatStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadFormatStats();
  }, [dateRange, customStartDate, customEndDate]);

  const loadFormatStats = async () => {
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

      const data = await GetStatsByFormat(filter);

      // Convert map to array
      const statsArray: FormatStats[] = Object.entries(data || {}).map(([format, stats]) => ({
        format,
        stats
      }));

      setFormatStats(statsArray);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load format statistics');
      console.error('Error loading format stats:', err);
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

  // Sort format stats
  const sortedFormats = [...formatStats].sort((a, b) => {
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
        aVal = a.format.toLowerCase();
        bVal = b.format.toLowerCase();
        break;
    }

    if (sortDirection === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0;
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0;
    }
  });

  // Transform data for charts
  const chartData = sortedFormats.map(item => ({
    name: item.format,
    matches: item.stats.TotalMatches,
    winRate: Math.round(item.stats.WinRate * 100 * 10) / 10
  }));

  const totalMatches = formatStats.reduce((sum, item) => sum + item.stats.TotalMatches, 0);

  return (
    <div className="page-container">
      <div className="format-distribution-header">
        <h1 className="page-title">Format Distribution</h1>

        {/* Filters */}
        <div className="filter-row">
          <div className="filter-group">
            <label className="filter-label">Date Range</label>
            <select value={dateRange} onChange={(e) => updateFilters('formatDistribution', { dateRange: e.target.value })}>
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
                  onChange={(e) => updateFilters('formatDistribution', { customStartDate: e.target.value })}
                />
              </div>

              <div className="filter-group">
                <label className="filter-label">End Date</label>
                <input
                  type="date"
                  value={customEndDate}
                  min={getMinEndDate()}
                  max={getTodayDateString()}
                  onChange={(e) => updateFilters('formatDistribution', { customEndDate: e.target.value })}
                />
              </div>
            </>
          )}

          <div className="filter-group">
            <label className="filter-label">Chart Type</label>
            <select value={chartType} onChange={(e) => updateFilters('formatDistribution', { chartType: e.target.value as 'pie' | 'bar' })}>
              <option value="bar">Bar Chart</option>
              <option value="pie">Pie Chart</option>
            </select>
          </div>

          <div className="filter-group">
            <label className="filter-label">Sort By</label>
            <select value={sortBy} onChange={(e) => updateFilters('formatDistribution', { sortBy: e.target.value as 'matches' | 'winRate' | 'name' })}>
              <option value="matches">Match Count</option>
              <option value="winRate">Win Rate</option>
              <option value="name">Format Name</option>
            </select>
          </div>

          <div className="filter-group">
            <label className="filter-label">Sort Order</label>
            <select value={sortDirection} onChange={(e) => updateFilters('formatDistribution', { sortDirection: e.target.value as 'asc' | 'desc' })}>
              <option value="desc">Descending</option>
              <option value="asc">Ascending</option>
            </select>
          </div>
        </div>

        {!loading && !error && formatStats.length > 0 && (
          <div className="format-count">
            {formatStats.length} format{formatStats.length !== 1 ? 's' : ''} • {totalMatches} total matches
          </div>
        )}
      </div>

      {/* Content */}
      {loading && <LoadingSpinner message="Loading format statistics..." />}

      {error && <div className="error">{error}</div>}

      {!loading && !error && formatStats.length === 0 && (
        <div className="no-data">No format data found for the selected filters</div>
      )}

      {!loading && !error && sortedFormats.length > 0 && (
        <>
          {/* Chart */}
          <div className="chart-container">
            <ResponsiveContainer width="100%" height={400}>
              {chartType === 'pie' ? (
                <PieChart>
                  <Pie
                    data={chartData}
                    dataKey="matches"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    outerRadius={120}
                    label
                  >
                    {chartData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #3d3d3d' }}
                    labelStyle={{ color: '#ffffff' }}
                  />
                  <Legend />
                </PieChart>
              ) : (
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#3d3d3d" />
                  <XAxis dataKey="name" stroke="#ffffff" />
                  <YAxis stroke="#ffffff" />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #3d3d3d' }}
                    labelStyle={{ color: '#ffffff' }}
                  />
                  <Legend />
                  <Bar dataKey="matches" fill="#4a9eff" name="Matches" />
                </BarChart>
              )}
            </ResponsiveContainer>
          </div>

          {/* Format Cards */}
          <div className="format-grid">
            {sortedFormats.map((item, index) => (
              <div key={item.format} className="format-card">
                <div className="format-header">
                  <h3 className="format-name">{item.format || 'Unknown Format'}</h3>
                  <div className="format-color-badge" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                </div>
                <div className="format-stats">
                  <div className="stat">
                    <span className="stat-label">Win Rate</span>
                    <span className="stat-value win-rate">{formatWinRate(item.stats.WinRate)}</span>
                  </div>
                  <div className="stat">
                    <span className="stat-label">Matches</span>
                    <span className="stat-value">{item.stats.TotalMatches}</span>
                  </div>
                  <div className="stat">
                    <span className="stat-label">Wins / Losses</span>
                    <span className="stat-value">{item.stats.MatchesWon}W - {item.stats.MatchesLost}L</span>
                  </div>
                  <div className="stat">
                    <span className="stat-label">Game Win Rate</span>
                    <span className="stat-value">{formatWinRate(item.stats.GameWinRate)}</span>
                  </div>
                  <div className="stat">
                    <span className="stat-label">Games</span>
                    <span className="stat-value">{item.stats.TotalGames} ({item.stats.GamesWon}W - {item.stats.GamesLost}L)</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
};

export default FormatDistribution;

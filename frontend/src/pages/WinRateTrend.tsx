import { useState, useEffect } from 'react';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { GetTrendAnalysis } from '../../wailsjs/go/main/App';
import { storage } from '../../wailsjs/go/models';
import LoadingSpinner from '../components/LoadingSpinner';
import EmptyState from '../components/EmptyState';
import ErrorState from '../components/ErrorState';
import { useAppContext } from '../context/AppContext';
import './WinRateTrend.css';

const WinRateTrend = () => {
  const { filters, updateFilters } = useAppContext();
  const { dateRange, format, chartType } = filters.winRateTrend;

  const [analysis, setAnalysis] = useState<storage.TrendAnalysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadTrendData();
  }, [dateRange, format]);

  const loadTrendData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Calculate date range
      const now = new Date();
      const start = new Date();
      let periodType = 'daily';

      switch (dateRange) {
        case '7days':
          start.setDate(now.getDate() - 7);
          periodType = 'daily';
          break;
        case '30days':
          start.setDate(now.getDate() - 30);
          periodType = 'weekly';
          break;
        case '90days':
          start.setDate(now.getDate() - 90);
          periodType = 'weekly';
          break;
        case 'all':
          start.setFullYear(now.getFullYear() - 1);
          periodType = 'monthly';
          break;
      }

      // Build formats array
      let formats: string[] | null = null;
      if (format === 'constructed') {
        formats = ['Ladder', 'Play'];
      } else if (format !== 'all') {
        formats = [format];
      }

      const data = await GetTrendAnalysis(start, now, periodType, formats || []);
      setAnalysis(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load trend data');
      console.error('Error loading trend data:', err);
    } finally {
      setLoading(false);
    }
  };

  // Transform data for Recharts
  const chartData = analysis?.Periods?.map(period => ({
    name: period.Period.Label,
    winRate: Math.round(period.WinRate * 100 * 10) / 10, // Convert to percentage with 1 decimal
    matches: period.Stats?.TotalMatches || 0
  })) || [];

  return (
    <div className="page-container">
      {/* Filters */}
      <div className="filter-row">
        <div className="filter-group">
          <label className="filter-label">Date Range</label>
          <select value={dateRange} onChange={(e) => updateFilters('winRateTrend', { dateRange: e.target.value })}>
            <option value="7days">Last 7 Days</option>
            <option value="30days">Last 30 Days</option>
            <option value="90days">Last 90 Days</option>
            <option value="all">All Time</option>
          </select>
        </div>

        <div className="filter-group">
          <label className="filter-label">Format</label>
          <select value={format} onChange={(e) => updateFilters('winRateTrend', { format: e.target.value })}>
            <option value="all">All Formats</option>
            <option value="constructed">Constructed</option>
            <option value="Ladder">Ranked (Ladder)</option>
            <option value="Play">Play Queue</option>
          </select>
        </div>

        <div className="filter-group">
          <label className="filter-label">Chart Type</label>
          <select value={chartType} onChange={(e) => updateFilters('winRateTrend', { chartType: e.target.value as 'line' | 'bar' })}>
            <option value="line">Line Chart</option>
            <option value="bar">Bar Chart</option>
          </select>
        </div>
      </div>

      {/* Content */}
      {loading && <LoadingSpinner message="Loading trend data..." />}

      {error && (
        <ErrorState
          message="Failed to load trend data"
          error={error}
          helpText="Check that the database is accessible and try refreshing the page."
        />
      )}

      {!loading && !error && (!analysis || chartData.length === 0) && (
        <EmptyState
          icon="📈"
          title="Not enough data"
          message="Play at least 5 matches to see your win rate trends over time."
          helpText="Your win rate will be tracked daily, weekly, or monthly depending on your selected date range."
        />
      )}

      {!loading && !error && analysis && chartData.length > 0 && (
        <>
          {/* Chart */}
          <div className="chart-container">
            <ResponsiveContainer width="100%" height={500}>
              {chartType === 'line' ? (
                <LineChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#3d3d3d" />
                  <XAxis dataKey="name" stroke="#ffffff" />
                  <YAxis stroke="#ffffff" domain={[0, 100]} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #3d3d3d' }}
                    labelStyle={{ color: '#ffffff' }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="winRate"
                    stroke="#4a9eff"
                    name="Win Rate (%)"
                    strokeWidth={2}
                    dot={{ fill: '#4a9eff', r: 4 }}
                  />
                </LineChart>
              ) : (
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#3d3d3d" />
                  <XAxis dataKey="name" stroke="#ffffff" />
                  <YAxis stroke="#ffffff" domain={[0, 100]} />
                  <Tooltip
                    contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #3d3d3d' }}
                    labelStyle={{ color: '#ffffff' }}
                  />
                  <Legend />
                  <Bar dataKey="winRate" fill="#4a9eff" name="Win Rate (%)" />
                </BarChart>
              )}
            </ResponsiveContainer>
          </div>

          {/* Summary */}
          <div className="summary">
            <h3>Win Rate Trend Analysis</h3>
            <div className="summary-content">
              <div className="summary-grid">
                <div className="summary-item">
                  <span className="summary-label">Period:</span>
                  <span className="summary-value">
                    {analysis.Periods[0]?.Period.StartDate?.toString().split('T')[0]} to {analysis.Periods[analysis.Periods.length - 1]?.Period.EndDate?.toString().split('T')[0]}
                  </span>
                </div>
                <div className="summary-item">
                  <span className="summary-label">Format:</span>
                  <span className="summary-value">{format === 'all' ? 'All Formats' : format}</span>
                </div>
                <div className="summary-item">
                  <span className="summary-label">Trend:</span>
                  <span className={`summary-value trend-${analysis.Trend}`}>
                    {analysis.Trend} {analysis.TrendValue !== 0 && `(${analysis.TrendValue > 0 ? '+' : ''}${Math.round(analysis.TrendValue * 100 * 10) / 10}%)`}
                  </span>
                </div>
                {analysis.Overall && (
                  <div className="summary-item">
                    <span className="summary-label">Overall Win Rate:</span>
                    <span className="summary-value">
                      {Math.round(analysis.Overall.WinRate * 100 * 10) / 10}% ({analysis.Overall.TotalMatches} matches)
                    </span>
                  </div>
                )}
              </div>
              <button className="export-button" onClick={() => alert('Export functionality coming soon!')}>Export as PNG</button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default WinRateTrend;

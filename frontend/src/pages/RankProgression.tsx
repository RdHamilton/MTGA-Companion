import { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { GetRankProgressionTimeline } from '../../wailsjs/go/main/App';
import { storage } from '../../wailsjs/go/models';
import LoadingSpinner from '../components/LoadingSpinner';
import EmptyState from '../components/EmptyState';
import ErrorState from '../components/ErrorState';
import { useAppContext } from '../context/AppContext';
import './RankProgression.css';

const RANK_CLASSES = [
  'Bronze',
  'Silver',
  'Gold',
  'Platinum',
  'Diamond',
  'Mythic'
];

const RankProgression = () => {
  const { filters, updateFilters } = useAppContext();
  const { dateRange } = filters.rankProgression;

  const [timeline, setTimeline] = useState<storage.RankTimelineEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadTimeline = async () => {
    try {
      setLoading(true);
      setError(null);

      // Calculate date range
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
        case 'all':
          start.setFullYear(now.getFullYear() - 1);
          break;
      }

      // Rank progression is only tracked for constructed (ranked) play
      const data = await GetRankProgressionTimeline('constructed', start, now, 'daily');
      setTimeline(data?.entries || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load rank progression');
      console.error('Error loading rank progression:', err);
    } finally {
      setLoading(false);
    }
  };

    loadTimeline();
  }, [dateRange]);

  // Convert rank to numeric value for charting
  const rankToNumeric = (rankClass: string | null | undefined, rankLevel: number | null | undefined): number => {
    if (!rankClass || rankLevel == null) return 0;

    const classIndex = RANK_CLASSES.indexOf(rankClass);
    if (classIndex === -1) return 0;

    // Each class has 4 tiers (4, 3, 2, 1), except Mythic which is special
    if (rankClass === 'Mythic') {
      return classIndex * 4 + 4;
    }

    // Reverse the level so higher tiers appear higher on chart
    // Level 4 -> 0, Level 3 -> 1, Level 2 -> 2, Level 1 -> 3
    const tierOffset = 4 - rankLevel;
    return classIndex * 4 + tierOffset;
  };

  // Convert numeric value back to rank display
  const numericToRank = (value: number): string => {
    const classIndex = Math.floor(value / 4);
    const tierOffset = value % 4;

    if (classIndex >= RANK_CLASSES.length) {
      return 'Mythic';
    }

    const rankClass = RANK_CLASSES[classIndex];
    if (rankClass === 'Mythic') {
      return 'Mythic';
    }

    const level = 4 - tierOffset;
    return `${rankClass} ${level}`;
  };

  // Transform data for Recharts
  const chartData = timeline.map(point => ({
    timestamp: new Date(point.timestamp as unknown as string).toLocaleDateString(),
    rankValue: rankToNumeric(point.rank_class || null, point.rank_level || null),
    rankDisplay: point.rank,
    isChange: point.is_change
  }));

  // Calculate statistics
  const getProgressionStats = () => {
    if (timeline.length === 0) return null;

    const first = timeline[0];
    const last = timeline[timeline.length - 1];

    const firstRank = rankToNumeric(first.rank_class || null, first.rank_level || null);
    const lastRank = rankToNumeric(last.rank_class || null, last.rank_level || null);

    const change = lastRank - firstRank;
    const direction = change > 0 ? 'up' : change < 0 ? 'down' : 'stable';

    const totalChanges = timeline.filter(p => p.is_change).length;

    return {
      firstRank: first.rank,
      lastRank: last.rank,
      direction,
      change: Math.abs(change),
      totalChanges,
      totalEntries: timeline.length
    };
  };

  const stats = getProgressionStats();

  // Custom Y-axis tick formatter
  const formatYAxis = (value: number) => {
    return numericToRank(value);
  };

  return (
    <div className="page-container">
      <div className="rank-progression-header">
        <h1 className="page-title">Rank Progression</h1>

        {/* Filters */}
        <div className="filter-row">
          <div className="filter-group">
            <label className="filter-label">Date Range</label>
            <select value={dateRange} onChange={(e) => updateFilters('rankProgression', { dateRange: e.target.value })}>
              <option value="7days">Last 7 Days</option>
              <option value="30days">Last 30 Days</option>
              <option value="90days">Last 90 Days</option>
              <option value="all">All Time</option>
            </select>
          </div>
        </div>
        <div className="format-note">
          <span className="note-text">Showing rank progression for Constructed (Ranked) ladder only</span>
        </div>
      </div>

      {/* Content */}
      {loading && <LoadingSpinner message="Loading rank progression..." />}

      {error && (
        <ErrorState
          message="Failed to load rank progression"
          error={error}
          helpText="Check that the database is accessible and try refreshing the page."
        />
      )}

      {!loading && !error && timeline.length === 0 && (
        <EmptyState
          icon="🏆"
          title="No rank progression data"
          message="Play ranked matches to track your rank progression over time."
          helpText="Your rank changes will be displayed here as you climb the ladder in ranked play."
        />
      )}

      {!loading && !error && timeline.length > 0 && stats && (
        <>
          {/* Statistics Summary */}
          <div className="progression-summary">
            <h3 className="summary-title">Progression Summary</h3>
            <div className="summary-grid">
              <div className="summary-item">
                <span className="summary-label">Starting Rank</span>
                <span className="summary-value">{stats.firstRank}</span>
              </div>
              <div className="summary-item">
                <span className="summary-label">Current Rank</span>
                <span className="summary-value">{stats.lastRank}</span>
              </div>
              <div className="summary-item">
                <span className="summary-label">Direction</span>
                <span className={`summary-value trend-${stats.direction}`}>
                  {stats.direction === 'up' ? '↑ Climbing' :
                   stats.direction === 'down' ? '↓ Falling' :
                   '→ Stable'}
                </span>
              </div>
              <div className="summary-item">
                <span className="summary-label">Total Entries</span>
                <span className="summary-value">{stats.totalEntries}</span>
              </div>
              <div className="summary-item">
                <span className="summary-label">Rank Changes</span>
                <span className="summary-value">{stats.totalChanges}</span>
              </div>
            </div>
          </div>

          {/* Chart */}
          <div className="chart-container">
            <ResponsiveContainer width="100%" height={500}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#3d3d3d" />
                <XAxis
                  dataKey="timestamp"
                  stroke="#ffffff"
                  angle={-45}
                  textAnchor="end"
                  height={80}
                />
                <YAxis
                  stroke="#ffffff"
                  tickFormatter={formatYAxis}
                  domain={[0, 24]}
                  ticks={[0, 4, 8, 12, 16, 20, 24]}
                  width={90}
                />
                <Tooltip
                  contentStyle={{ backgroundColor: '#2d2d2d', border: '1px solid #3d3d3d' }}
                  labelStyle={{ color: '#ffffff' }}
                  formatter={(value: unknown, name: string) => {
                    if (name === 'Rank') {
                      return numericToRank(value as number);
                    }
                    return String(value);
                  }}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="rankValue"
                  name="Rank"
                  stroke="#4a9eff"
                  strokeWidth={2}
                  dot={{ fill: '#4a9eff', r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Detailed Timeline */}
          <div className="timeline-section">
            <h3 className="timeline-title">Detailed Timeline</h3>
            <div className="timeline-list">
              {timeline.slice().reverse().map((point, index) => (
                <div key={index} className={`timeline-item ${point.is_change ? 'changed' : ''}`}>
                  <div className="timeline-date">
                    {new Date(point.timestamp as unknown as string).toLocaleString()}
                  </div>
                  <div className="timeline-rank">
                    {point.rank}
                    {point.is_change && <span className="timeline-change"> (Changed)</span>}
                  </div>
                  {point.rank_step != null && (
                    <div className="timeline-steps">
                      Step {point.rank_step}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default RankProgression;

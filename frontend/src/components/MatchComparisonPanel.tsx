import { useState } from 'react';
import type {
  ComparisonResult,
  ComparisonGroup,
  StatsFilterRequest,
} from '@/services/api/matches';
import {
  compareFormats,
  compareDecks,
  compareTimePeriods,
} from '@/services/api/matches';
import './MatchComparisonPanel.css';

interface MatchComparisonPanelProps {
  formats?: string[];
  deckIds?: { id: string; name: string }[];
  onClose?: () => void;
}

type ComparisonType = 'formats' | 'decks' | 'time-periods';

interface TimePeriodOption {
  label: string;
  startDate: string;
  endDate: string;
}

function formatWinRate(rate: number | undefined): string {
  if (rate === undefined || rate === null) return '0%';
  return `${(rate * 100).toFixed(1)}%`;
}

function formatDiff(diff: number): string {
  const percentage = (diff * 100).toFixed(1);
  if (diff > 0) return `+${percentage}%`;
  if (diff < 0) return `${percentage}%`;
  return '0%';
}

function getDiffClass(diff: number): string {
  if (diff > 0.05) return 'diff-positive';
  if (diff < -0.05) return 'diff-negative';
  return 'diff-neutral';
}

function getDefaultTimePeriods(): TimePeriodOption[] {
  const now = new Date();
  const today = now.toISOString().split('T')[0];

  // Last 7 days
  const weekAgo = new Date(now);
  weekAgo.setDate(weekAgo.getDate() - 7);
  const weekAgoStr = weekAgo.toISOString().split('T')[0];

  // Last 30 days
  const monthAgo = new Date(now);
  monthAgo.setDate(monthAgo.getDate() - 30);
  const monthAgoStr = monthAgo.toISOString().split('T')[0];

  // Previous 30 days (30-60 days ago)
  const twoMonthsAgo = new Date(now);
  twoMonthsAgo.setDate(twoMonthsAgo.getDate() - 60);
  const twoMonthsAgoStr = twoMonthsAgo.toISOString().split('T')[0];

  return [
    { label: 'Last 7 Days', startDate: weekAgoStr, endDate: today },
    { label: 'Last 30 Days', startDate: monthAgoStr, endDate: today },
    { label: 'Previous 30 Days', startDate: twoMonthsAgoStr, endDate: monthAgoStr },
  ];
}

export default function MatchComparisonPanel({
  formats = [],
  deckIds = [],
  onClose,
}: MatchComparisonPanelProps) {
  const [comparisonType, setComparisonType] = useState<ComparisonType>('formats');
  const [selectedFormats, setSelectedFormats] = useState<string[]>([]);
  const [selectedDecks, setSelectedDecks] = useState<string[]>([]);
  const [selectedPeriods, setSelectedPeriods] = useState<TimePeriodOption[]>([]);
  const [result, setResult] = useState<ComparisonResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const timePeriodOptions = getDefaultTimePeriods();

  const handleCompare = async () => {
    setLoading(true);
    setError(null);
    try {
      let comparisonResult: ComparisonResult;
      const baseFilter: StatsFilterRequest = {};

      switch (comparisonType) {
        case 'formats':
          if (selectedFormats.length < 2) {
            setError('Please select at least 2 formats to compare');
            setLoading(false);
            return;
          }
          comparisonResult = await compareFormats({
            formats: selectedFormats,
            base_filter: baseFilter,
          });
          break;
        case 'decks':
          if (selectedDecks.length < 2) {
            setError('Please select at least 2 decks to compare');
            setLoading(false);
            return;
          }
          comparisonResult = await compareDecks({
            deck_ids: selectedDecks,
            base_filter: baseFilter,
          });
          break;
        case 'time-periods':
          if (selectedPeriods.length < 2) {
            setError('Please select at least 2 time periods to compare');
            setLoading(false);
            return;
          }
          comparisonResult = await compareTimePeriods({
            periods: selectedPeriods.map((p) => ({
              label: p.label,
              start_date: p.startDate,
              end_date: p.endDate,
            })),
            base_filter: baseFilter,
          });
          break;
        default:
          throw new Error('Invalid comparison type');
      }

      setResult(comparisonResult);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to compare matches');
    } finally {
      setLoading(false);
    }
  };

  const toggleFormat = (format: string) => {
    setSelectedFormats((prev) =>
      prev.includes(format) ? prev.filter((f) => f !== format) : [...prev, format]
    );
  };

  const toggleDeck = (deckId: string) => {
    setSelectedDecks((prev) =>
      prev.includes(deckId) ? prev.filter((d) => d !== deckId) : [...prev, deckId]
    );
  };

  const togglePeriod = (period: TimePeriodOption) => {
    setSelectedPeriods((prev) => {
      const exists = prev.some((p) => p.label === period.label);
      if (exists) {
        return prev.filter((p) => p.label !== period.label);
      }
      return [...prev, period];
    });
  };

  const renderSelector = () => {
    switch (comparisonType) {
      case 'formats':
        return (
          <div className="comparison-selector">
            <h4>Select Formats to Compare</h4>
            <div className="selector-options">
              {formats.length > 0 ? (
                formats.map((format) => (
                  <label key={format} className="selector-option">
                    <input
                      type="checkbox"
                      checked={selectedFormats.includes(format)}
                      onChange={() => toggleFormat(format)}
                    />
                    {format}
                  </label>
                ))
              ) : (
                <p className="no-options">No formats available. Play some matches first.</p>
              )}
            </div>
          </div>
        );
      case 'decks':
        return (
          <div className="comparison-selector">
            <h4>Select Decks to Compare</h4>
            <div className="selector-options">
              {deckIds.length > 0 ? (
                deckIds.map((deck) => (
                  <label key={deck.id} className="selector-option">
                    <input
                      type="checkbox"
                      checked={selectedDecks.includes(deck.id)}
                      onChange={() => toggleDeck(deck.id)}
                    />
                    {deck.name}
                  </label>
                ))
              ) : (
                <p className="no-options">No decks available. Create some decks first.</p>
              )}
            </div>
          </div>
        );
      case 'time-periods':
        return (
          <div className="comparison-selector">
            <h4>Select Time Periods to Compare</h4>
            <div className="selector-options">
              {timePeriodOptions.map((period) => (
                <label key={period.label} className="selector-option">
                  <input
                    type="checkbox"
                    checked={selectedPeriods.some((p) => p.label === period.label)}
                    onChange={() => togglePeriod(period)}
                  />
                  {period.label}
                </label>
              ))}
            </div>
          </div>
        );
      default:
        return null;
    }
  };

  const renderGroupRow = (group: ComparisonGroup, isBest: boolean, isWorst: boolean) => {
    const stats = group.Statistics;
    return (
      <tr
        key={group.Label}
        className={`${isBest ? 'best-group' : ''} ${isWorst ? 'worst-group' : ''}`}
      >
        <td className="group-label">
          {group.Label}
          {isBest && <span className="badge badge-best">Best</span>}
          {isWorst && <span className="badge badge-worst">Worst</span>}
        </td>
        <td className="stat-value">{group.MatchCount}</td>
        <td className="stat-value win-rate">{formatWinRate(stats?.WinRate)}</td>
        <td className="stat-value">{formatWinRate(stats?.GameWinRate)}</td>
        <td className="stat-value">{stats?.TotalGames || 0}</td>
      </tr>
    );
  };

  const renderResults = () => {
    if (!result) return null;

    return (
      <div className="comparison-results">
        <div className="results-header">
          <h4>Comparison Results</h4>
          <div className="results-summary">
            <span>Total Matches: {result.TotalMatches}</span>
            <span className={getDiffClass(result.WinRateDiff)}>
              Win Rate Spread: {formatDiff(result.WinRateDiff)}
            </span>
          </div>
        </div>

        <table className="comparison-table">
          <thead>
            <tr>
              <th>Group</th>
              <th>Matches</th>
              <th>Win Rate</th>
              <th>Game Win Rate</th>
              <th>Games</th>
            </tr>
          </thead>
          <tbody>
            {result.Groups.map((group) =>
              renderGroupRow(
                group,
                result.BestGroup?.Label === group.Label,
                result.WorstGroup?.Label === group.Label
              )
            )}
          </tbody>
        </table>

        {result.BestGroup && result.WorstGroup && result.BestGroup !== result.WorstGroup && (
          <div className="comparison-insight">
            <strong>Insight:</strong> Your best performance is in{' '}
            <span className="highlight">{result.BestGroup.Label}</span> with a{' '}
            {formatWinRate(result.BestGroup.Statistics?.WinRate)} win rate.
            {result.WinRateDiff > 0.1 && (
              <>
                {' '}
                Consider focusing on this format to maximize your win rate.
              </>
            )}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="match-comparison-panel">
      <div className="panel-header">
        <h3>Match Comparison</h3>
        {onClose && (
          <button className="close-button" onClick={onClose}>
            &times;
          </button>
        )}
      </div>

      <div className="comparison-type-selector">
        <button
          className={`type-button ${comparisonType === 'formats' ? 'active' : ''}`}
          onClick={() => setComparisonType('formats')}
        >
          Compare Formats
        </button>
        <button
          className={`type-button ${comparisonType === 'decks' ? 'active' : ''}`}
          onClick={() => setComparisonType('decks')}
        >
          Compare Decks
        </button>
        <button
          className={`type-button ${comparisonType === 'time-periods' ? 'active' : ''}`}
          onClick={() => setComparisonType('time-periods')}
        >
          Compare Time Periods
        </button>
      </div>

      {renderSelector()}

      {error && <div className="comparison-error">{error}</div>}

      <div className="comparison-actions">
        <button className="compare-button" onClick={handleCompare} disabled={loading}>
          {loading ? 'Comparing...' : 'Compare'}
        </button>
      </div>

      {renderResults()}
    </div>
  );
}

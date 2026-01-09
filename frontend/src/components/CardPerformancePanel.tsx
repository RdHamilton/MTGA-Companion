import { useState, useEffect } from 'react';
import {
  getCardPerformance,
  getAllPerformanceRecommendations,
} from '@/services/api/decks';
import type {
  DeckPerformanceAnalysis,
  DeckRecommendationsResponse,
  PerformanceCardRecommendation,
} from '@/services/api/decks';
import './CardPerformancePanel.css';

interface CardPerformancePanelProps {
  deckId: string;
  deckName: string;
  onClose: () => void;
}

type SortField = 'cardName' | 'winContribution' | 'impactScore' | 'gamesDrawn' | 'playRate';
type SortDirection = 'asc' | 'desc';

export const CardPerformancePanel = ({
  deckId,
  deckName,
  onClose,
}: CardPerformancePanelProps) => {
  const [analysis, setAnalysis] = useState<DeckPerformanceAnalysis | null>(null);
  const [recommendations, setRecommendations] = useState<DeckRecommendationsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'performance' | 'recommendations'>('performance');
  const [sortField, setSortField] = useState<SortField>('impactScore');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      setError(null);

      try {
        const [perfData, recsData] = await Promise.all([
          getCardPerformance(deckId),
          getAllPerformanceRecommendations(deckId).catch(() => null),
        ]);
        setAnalysis(perfData);
        setRecommendations(recsData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load performance data');
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [deckId]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedCards = analysis?.cardPerformance
    ? [...analysis.cardPerformance].sort((a, b) => {
        let comparison = 0;
        switch (sortField) {
          case 'cardName':
            comparison = a.cardName.localeCompare(b.cardName);
            break;
          case 'winContribution':
            comparison = a.winContribution - b.winContribution;
            break;
          case 'impactScore':
            comparison = a.impactScore - b.impactScore;
            break;
          case 'gamesDrawn':
            comparison = a.gamesDrawn - b.gamesDrawn;
            break;
          case 'playRate':
            comparison = a.playRate - b.playRate;
            break;
        }
        return sortDirection === 'asc' ? comparison : -comparison;
      })
    : [];

  const getGradeColor = (grade: string): string => {
    switch (grade) {
      case 'excellent':
        return '#4caf50';
      case 'good':
        return '#8bc34a';
      case 'average':
        return '#ffc107';
      case 'poor':
        return '#ff9800';
      case 'bad':
        return '#f44336';
      default:
        return '#888';
    }
  };

  const getImpactColor = (impact: number): string => {
    if (impact > 0.3) return '#4caf50';
    if (impact > 0.1) return '#8bc34a';
    if (impact > -0.1) return '#ffc107';
    if (impact > -0.3) return '#ff9800';
    return '#f44336';
  };

  const formatPercent = (value: number): string => {
    return `${(value * 100).toFixed(1)}%`;
  };

  const formatImpact = (value: number): string => {
    const sign = value >= 0 ? '+' : '';
    return `${sign}${(value * 100).toFixed(1)}%`;
  };

  const renderPerformanceTable = () => {
    if (!analysis || sortedCards.length === 0) {
      return (
        <div className="empty-state">
          <p>Not enough game data to analyze card performance.</p>
          <p className="hint">Play more games with this deck to see performance metrics.</p>
        </div>
      );
    }

    return (
      <div className="performance-table-container">
        <table className="performance-table">
          <thead>
            <tr>
              <th onClick={() => handleSort('cardName')} className="sortable">
                Card {sortField === 'cardName' && (sortDirection === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => handleSort('winContribution')} className="sortable">
                Win Impact {sortField === 'winContribution' && (sortDirection === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => handleSort('impactScore')} className="sortable">
                Score {sortField === 'impactScore' && (sortDirection === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => handleSort('gamesDrawn')} className="sortable">
                Games {sortField === 'gamesDrawn' && (sortDirection === 'asc' ? '↑' : '↓')}
              </th>
              <th onClick={() => handleSort('playRate')} className="sortable">
                Play Rate {sortField === 'playRate' && (sortDirection === 'asc' ? '↑' : '↓')}
              </th>
              <th>Grade</th>
              <th>Confidence</th>
            </tr>
          </thead>
          <tbody>
            {sortedCards.map((card) => (
              <tr key={card.cardId}>
                <td className="card-name">{card.cardName}</td>
                <td style={{ color: getImpactColor(card.winContribution) }}>
                  {formatImpact(card.winContribution)}
                </td>
                <td>
                  <div
                    className="impact-bar"
                    style={{
                      width: `${Math.abs(card.impactScore) * 50 + 50}%`,
                      backgroundColor: getImpactColor(card.impactScore),
                    }}
                  >
                    {card.impactScore.toFixed(2)}
                  </div>
                </td>
                <td>{card.gamesDrawn}</td>
                <td>{formatPercent(card.playRate)}</td>
                <td>
                  <span
                    className="grade-badge"
                    style={{ backgroundColor: getGradeColor(card.performanceGrade) }}
                  >
                    {card.performanceGrade}
                  </span>
                </td>
                <td>
                  <span className={`confidence-badge ${card.confidenceLevel}`}>
                    {card.confidenceLevel}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const renderRecommendationCard = (rec: PerformanceCardRecommendation) => {
    const getTypeIcon = (type: string) => {
      switch (type) {
        case 'add':
          return '+';
        case 'remove':
          return '-';
        case 'swap':
          return '↔';
        default:
          return '?';
      }
    };

    const getTypeColor = (type: string) => {
      switch (type) {
        case 'add':
          return '#4caf50';
        case 'remove':
          return '#f44336';
        case 'swap':
          return '#2196f3';
        default:
          return '#888';
      }
    };

    return (
      <div key={`${rec.type}-${rec.cardId}`} className="recommendation-card">
        <div className="rec-header">
          <span
            className="rec-type-icon"
            style={{ backgroundColor: getTypeColor(rec.type) }}
          >
            {getTypeIcon(rec.type)}
          </span>
          <span className="rec-card-name">{rec.cardName}</span>
          {rec.type === 'swap' && rec.swapForCardName && (
            <>
              <span className="swap-arrow">→</span>
              <span className="rec-card-name">{rec.swapForCardName}</span>
            </>
          )}
          <span className={`rec-confidence ${rec.confidence}`}>{rec.confidence}</span>
        </div>
        <div className="rec-reason">{rec.reason}</div>
        <div className="rec-footer">
          <span className="rec-impact" style={{ color: getImpactColor(rec.impactEstimate) }}>
            Est. Impact: {formatImpact(rec.impactEstimate)}
          </span>
          <span className="rec-games">Based on {rec.basedOnGames} games</span>
        </div>
      </div>
    );
  };

  const renderRecommendations = () => {
    if (!recommendations) {
      return (
        <div className="empty-state">
          <p>Not enough data to generate recommendations.</p>
          <p className="hint">Play more games to get personalized suggestions.</p>
        </div>
      );
    }

    const hasAnyRecs =
      recommendations.addRecommendations.length > 0 ||
      recommendations.removeRecommendations.length > 0 ||
      recommendations.swapRecommendations.length > 0;

    if (!hasAnyRecs) {
      return (
        <div className="empty-state">
          <p>No recommendations at this time.</p>
          <p className="hint">Your deck is performing well based on current data!</p>
        </div>
      );
    }

    return (
      <div className="recommendations-container">
        <div className="rec-summary">
          <div className="rec-stat">
            <span className="rec-stat-label">Current Win Rate</span>
            <span className="rec-stat-value">{formatPercent(recommendations.currentWinRate)}</span>
          </div>
          <div className="rec-stat">
            <span className="rec-stat-label">Projected Win Rate</span>
            <span
              className="rec-stat-value"
              style={{ color: getImpactColor(recommendations.projectedWinRate - recommendations.currentWinRate) }}
            >
              {formatPercent(recommendations.projectedWinRate)}
            </span>
          </div>
        </div>

        {recommendations.removeRecommendations.length > 0 && (
          <div className="rec-section">
            <h3>Cards to Consider Removing</h3>
            {recommendations.removeRecommendations.map(renderRecommendationCard)}
          </div>
        )}

        {recommendations.addRecommendations.length > 0 && (
          <div className="rec-section">
            <h3>Cards to Consider Adding</h3>
            {recommendations.addRecommendations.map(renderRecommendationCard)}
          </div>
        )}

        {recommendations.swapRecommendations.length > 0 && (
          <div className="rec-section">
            <h3>Suggested Swaps</h3>
            {recommendations.swapRecommendations.map(renderRecommendationCard)}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="card-performance-panel">
      <div className="panel-header">
        <h2>Card Performance: {deckName}</h2>
        <button className="close-button" onClick={onClose}>
          ×
        </button>
      </div>

      {loading ? (
        <div className="loading-state">
          <div className="loading-spinner" />
          <p>Analyzing card performance...</p>
        </div>
      ) : error ? (
        <div className="error-state">
          <p>{error}</p>
        </div>
      ) : (
        <>
          {analysis && (
            <div className="panel-summary">
              <div className="summary-stat">
                <span className="stat-value">{analysis.totalMatches}</span>
                <span className="stat-label">Matches</span>
              </div>
              <div className="summary-stat">
                <span className="stat-value">{analysis.totalGames}</span>
                <span className="stat-label">Games</span>
              </div>
              <div className="summary-stat">
                <span className="stat-value">{formatPercent(analysis.overallWinRate)}</span>
                <span className="stat-label">Win Rate</span>
              </div>
              <div className="summary-stat">
                <span className="stat-value">{analysis.cardPerformance.length}</span>
                <span className="stat-label">Cards Analyzed</span>
              </div>
            </div>
          )}

          <div className="panel-tabs">
            <button
              className={`tab-button ${activeTab === 'performance' ? 'active' : ''}`}
              onClick={() => setActiveTab('performance')}
            >
              Card Performance
            </button>
            <button
              className={`tab-button ${activeTab === 'recommendations' ? 'active' : ''}`}
              onClick={() => setActiveTab('recommendations')}
            >
              Recommendations
            </button>
          </div>

          <div className="panel-content">
            {activeTab === 'performance' ? renderPerformanceTable() : renderRecommendations()}
          </div>
        </>
      )}
    </div>
  );
};

export default CardPerformancePanel;

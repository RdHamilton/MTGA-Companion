import { useState, useEffect } from 'react';
import { GetMissingCards } from '@/services/api/legacy';
import { models } from '@/types/models';
import './MissingCards.css';

interface MissingCardsProps {
  sessionID: string;
  packNumber: number;
  pickNumber: number;
}

const MissingCards = ({ sessionID, packNumber, pickNumber }: MissingCardsProps) => {
  const [analysis, setAnalysis] = useState<models.MissingCardsAnalysis | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [expanded, setExpanded] = useState(false);

  useEffect(() => {
    const loadMissingCards = async () => {
      if (!sessionID || pickNumber <= 1) {
        // No missing cards for P1P1
        setAnalysis(null);
        return;
      }

      setLoading(true);
      setError(null);

      try {
        const result = await GetMissingCards(sessionID, packNumber, pickNumber);
        // Only set analysis if we have actual missing cards
        if (result && result.TotalMissing > 0) {
          setAnalysis(result);
        } else {
          setAnalysis(null);
        }
      } catch (err) {
        console.error('Failed to load missing cards:', err);
        // Don't show error to user - just log it and hide component
        // This is expected when pack data isn't available yet
        setAnalysis(null);
        setError(null);
      } finally {
        setLoading(false);
      }
    };

    loadMissingCards();
  }, [sessionID, packNumber, pickNumber]);

  // Don't show anything while loading or if there's no data
  if (loading || error || !analysis || analysis.TotalMissing === 0) {
    return null;
  }

  return (
    <div className="missing-cards-container">
      <div className="missing-cards-header" onClick={() => setExpanded(!expanded)}>
        <div className="missing-cards-summary">
          <span className="missing-cards-icon">📦</span>
          <span className="missing-cards-title">
            {analysis.TotalMissing} card{analysis.TotalMissing !== 1 ? 's' : ''} taken from this pack
          </span>
          {analysis.BombsMissing > 0 && (
            <span className="missing-cards-bombs">
              ({analysis.BombsMissing} bomb{analysis.BombsMissing !== 1 ? 's' : ''})
            </span>
          )}
        </div>
        <span className="missing-cards-toggle">{expanded ? '▼' : '▶'}</span>
      </div>

      {expanded && (
        <div className="missing-cards-list">
          <table className="missing-cards-table">
            <thead>
              <tr>
                <th>Card</th>
                <th>Tier</th>
                <th>Win Rate</th>
                <th>Likely Pick</th>
                <th>Wheel %</th>
              </tr>
            </thead>
            <tbody>
              {analysis.MissingCards
                .sort((a, b) => b.GIHWR - a.GIHWR) // Sort by rating descending
                .map((card, index) => (
                  <tr key={index} className={`missing-card-row tier-${card.Tier?.toLowerCase()}`}>
                    <td className="missing-card-name">{card.CardName}</td>
                    <td className="missing-card-tier">
                      <span className={`tier-badge tier-${card.Tier?.toLowerCase()}`}>
                        {card.Tier || 'N/A'}
                      </span>
                    </td>
                    <td className="missing-card-gihwr">
                      {card.GIHWR > 0 ? `${card.GIHWR.toFixed(1)}%` : 'N/A'}
                    </td>
                    <td className="missing-card-picked-at">
                      Pick {card.PickedAt > 0 ? card.PickedAt : '?'}
                    </td>
                    <td className="missing-card-wheel">
                      {card.WheelProbability > 0
                        ? `${card.WheelProbability.toFixed(0)}%`
                        : '-'}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default MissingCards;

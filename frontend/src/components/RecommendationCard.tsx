import { useState } from 'react';
import { drafts } from '@/services/api';
import { gui } from '@/types/models';
import './RecommendationCard.css';

interface RecommendationCardProps {
  recommendation: gui.CardRecommendation;
  deckID: string;
  onAddCard: (cardID: number, quantity: number, board: 'main' | 'sideboard') => void;
}

export default function RecommendationCard({
  recommendation,
  deckID,
  onAddCard,
}: RecommendationCardProps) {
  const [showDetails, setShowDetails] = useState(false);
  const [detailedExplanation, setDetailedExplanation] = useState<string | null>(null);
  const [loadingExplanation, setLoadingExplanation] = useState(false);
  const [explanationError, setExplanationError] = useState<string | null>(null);

  const handleShowDetails = async () => {
    if (showDetails) {
      setShowDetails(false);
      return;
    }

    setShowDetails(true);

    // Only fetch explanation if we don't have it yet
    if (!detailedExplanation && !loadingExplanation) {
      setLoadingExplanation(true);
      setExplanationError(null);

      try {
        const response = await drafts.explainRecommendation({
          deckID,
          cardID: recommendation.cardID,
        });

        if (response.error) {
          setExplanationError(response.error);
        } else {
          setDetailedExplanation(response.explanation);
        }
      } catch (err) {
        setExplanationError(
          err instanceof Error ? err.message : 'Failed to load explanation'
        );
      } finally {
        setLoadingExplanation(false);
      }
    }
  };

  const formatPercentage = (value: number): string => {
    return `${Math.round(value * 100)}%`;
  };

  const getScoreBarClass = (value: number): string => {
    if (value >= 0.8) return 'score-bar-excellent';
    if (value >= 0.6) return 'score-bar-good';
    if (value >= 0.4) return 'score-bar-fair';
    return 'score-bar-poor';
  };

  const rec = recommendation;

  return (
    <div className={`recommendation-card ${showDetails ? 'expanded' : ''}`}>
      <div className="rec-card-main">
        {rec.imageURI && (
          <img src={rec.imageURI} alt={rec.name} className="rec-card-image" />
        )}
        <div className="rec-card-info">
          <div className="rec-card-name">{rec.name}</div>
          <div className="rec-card-type">{rec.typeLine}</div>
          {rec.manaCost && <div className="rec-card-mana">{rec.manaCost}</div>}
          <div className="rec-score-summary">
            <span className="score-badge">
              {formatPercentage(rec.score)}
            </span>
            <span className="confidence-badge">
              {formatPercentage(rec.confidence)} confidence
            </span>
          </div>
          <div className="rec-reasoning">{rec.reasoning}</div>
        </div>
        <div className="rec-card-actions">
          <button
            className="add-rec-button"
            onClick={() => onAddCard(rec.cardID, 1, 'main')}
            title="Add to mainboard"
          >
            + Add
          </button>
          <button
            className={`explain-button ${showDetails ? 'active' : ''}`}
            onClick={handleShowDetails}
            title="Show detailed explanation"
          >
            {showDetails ? '- Less' : '? Why'}
          </button>
        </div>
      </div>

      {showDetails && (
        <div className="rec-card-details">
          {/* Score Factors Breakdown */}
          {rec.factors && (
            <div className="score-factors">
              <h4>Score Breakdown</h4>
              <div className="factor-list">
                <div className="factor-item">
                  <span className="factor-label">Color Fit</span>
                  <div className="factor-bar-container">
                    <div
                      className={`factor-bar ${getScoreBarClass(rec.factors.colorFit)}`}
                      style={{ width: `${rec.factors.colorFit * 100}%` }}
                    />
                  </div>
                  <span className="factor-value">{formatPercentage(rec.factors.colorFit)}</span>
                </div>
                <div className="factor-item">
                  <span className="factor-label">Mana Curve</span>
                  <div className="factor-bar-container">
                    <div
                      className={`factor-bar ${getScoreBarClass(rec.factors.manaCurve)}`}
                      style={{ width: `${rec.factors.manaCurve * 100}%` }}
                    />
                  </div>
                  <span className="factor-value">{formatPercentage(rec.factors.manaCurve)}</span>
                </div>
                <div className="factor-item">
                  <span className="factor-label">Synergy</span>
                  <div className="factor-bar-container">
                    <div
                      className={`factor-bar ${getScoreBarClass(rec.factors.synergy)}`}
                      style={{ width: `${rec.factors.synergy * 100}%` }}
                    />
                  </div>
                  <span className="factor-value">{formatPercentage(rec.factors.synergy)}</span>
                </div>
                <div className="factor-item">
                  <span className="factor-label">Card Quality</span>
                  <div className="factor-bar-container">
                    <div
                      className={`factor-bar ${getScoreBarClass(rec.factors.quality)}`}
                      style={{ width: `${rec.factors.quality * 100}%` }}
                    />
                  </div>
                  <span className="factor-value">{formatPercentage(rec.factors.quality)}</span>
                </div>
                <div className="factor-item">
                  <span className="factor-label">Playability</span>
                  <div className="factor-bar-container">
                    <div
                      className={`factor-bar ${getScoreBarClass(rec.factors.playable)}`}
                      style={{ width: `${rec.factors.playable * 100}%` }}
                    />
                  </div>
                  <span className="factor-value">{formatPercentage(rec.factors.playable)}</span>
                </div>
              </div>
            </div>
          )}

          {/* Detailed Explanation */}
          <div className="detailed-explanation">
            <h4>Why This Card?</h4>
            {loadingExplanation && (
              <div className="explanation-loading">
                <span className="loading-spinner" />
                Generating explanation...
              </div>
            )}
            {explanationError && (
              <div className="explanation-error">
                Unable to generate explanation: {explanationError}
              </div>
            )}
            {detailedExplanation && (
              <div className="explanation-text">{detailedExplanation}</div>
            )}
            {!loadingExplanation && !explanationError && !detailedExplanation && (
              <div className="explanation-text">{rec.reasoning}</div>
            )}
          </div>

          {/* Source Info */}
          <div className="rec-source">
            <span className="source-label">Source:</span>
            <span className={`source-value source-${rec.source}`}>
              {rec.source === 'ml' ? 'ML Model' :
               rec.source === 'meta' ? 'Metagame Data' :
               rec.source === 'personal' ? 'Your Play History' :
               rec.source}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

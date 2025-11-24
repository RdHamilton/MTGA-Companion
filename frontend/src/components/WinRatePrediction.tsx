import React, { useState, useEffect } from 'react';
import { PredictDraftWinRate, GetDraftWinRatePrediction } from '../../wailsjs/go/main/App';
import { prediction } from '../../wailsjs/go/models';
import './WinRatePrediction.css';

interface WinRatePredictionProps {
  sessionID: string;
  showPredictButton?: boolean;
  onPredictionCalculated?: (pred: prediction.DeckPrediction) => void;
  compact?: boolean;
}

export const WinRatePrediction: React.FC<WinRatePredictionProps> = ({
  sessionID,
  showPredictButton = false,
  onPredictionCalculated,
  compact = false,
}) => {
  const [pred, setPred] = useState<prediction.DeckPrediction | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showBreakdown, setShowBreakdown] = useState(false);

  useEffect(() => {
    const loadPrediction = async () => {
      try {
        setLoading(true);
        setError(null);
        const p = await GetDraftWinRatePrediction(sessionID);
        setPred(p);
      } finally {
        setLoading(false);
      }
    };

    loadPrediction();
  }, [sessionID]);

  const calculatePrediction = async () => {
    try {
      setLoading(true);
      setError(null);
      const p = await PredictDraftWinRate(sessionID);
      setPred(p);
      if (onPredictionCalculated) {
        onPredictionCalculated(p);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to calculate prediction');
    } finally {
      setLoading(false);
    }
  };

  const getWinRateColor = (winRate: number): string => {
    if (winRate >= 0.60) return '#44ff88'; // 60%+ green
    if (winRate >= 0.55) return '#4a9eff'; // 55-60% blue
    if (winRate >= 0.50) return '#ffaa44'; // 50-55% orange
    return '#ff4444'; // < 50% red
  };

  if (loading) {
    return (
      <div className={`win-rate-prediction ${compact ? 'compact' : ''}`}>
        <div className="loading">Loading prediction...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`win-rate-prediction ${compact ? 'compact' : ''}`}>
        <div className="error">{error}</div>
      </div>
    );
  }

  if (!pred) {
    if (!showPredictButton) {
      return null;
    }
    return (
      <div className={`win-rate-prediction ${compact ? 'compact' : ''}`}>
        <button onClick={calculatePrediction} className="predict-button">
          📊 Predict Win Rate
        </button>
      </div>
    );
  }

  const winRatePercent = Math.round(pred.PredictedWinRate * 100);
  const minPercent = Math.round(pred.PredictedWinRateMin * 100);
  const maxPercent = Math.round(pred.PredictedWinRateMax * 100);
  const winRateColor = getWinRateColor(pred.PredictedWinRate);

  if (compact) {
    return (
      <div
        className="win-rate-badge"
        style={{ borderColor: winRateColor }}
        onClick={() => setShowBreakdown(true)}
        title={`Expected ${winRatePercent}% win rate (${minPercent}-${maxPercent}%)`}
      >
        <div className="win-rate-value" style={{ color: winRateColor }}>
          {winRatePercent}%
        </div>
        <div className="win-rate-label">Win Rate</div>
      </div>
    );
  }

  return (
    <>
      <div className="win-rate-prediction">
        <div className="prediction-card" onClick={() => setShowBreakdown(true)}>
          <h3>Predicted Win Rate</h3>
          <div className="win-rate-display">
            <div className="win-rate-main" style={{ color: winRateColor }}>
              {winRatePercent}%
            </div>
            <div className="win-rate-range">
              ({minPercent}-{maxPercent}%)
            </div>
          </div>
          <div className="win-rate-explanation">
            {pred.Factors.explanation}
          </div>
          <div className="prediction-hint">Click for details</div>
        </div>
      </div>

      {showBreakdown && (
        <PredictionBreakdownModal
          prediction={pred}
          onClose={() => setShowBreakdown(false)}
        />
      )}
    </>
  );
};

interface PredictionBreakdownModalProps {
  prediction: prediction.DeckPrediction;
  onClose: () => void;
}

const PredictionBreakdownModal: React.FC<PredictionBreakdownModalProps> = ({ prediction, onClose }) => {
  const winRatePercent = Math.round(prediction.PredictedWinRate * 100);
  const minPercent = Math.round(prediction.PredictedWinRateMin * 100);
  const maxPercent = Math.round(prediction.PredictedWinRateMax * 100);

  const factors = [
    {
      label: 'Card Quality (GIHWR)',
      value: prediction.Factors.deck_average_gihwr,
      format: (v: number) => `${(v * 100).toFixed(1)}%`,
      tooltip: 'Average win rate when cards are in hand'
    },
    {
      label: 'Curve Score',
      value: prediction.Factors.curve_score,
      format: (v: number) => `${(v * 100).toFixed(0)}%`,
      tooltip: 'Quality of mana curve distribution'
    },
    {
      label: 'Color Discipline',
      value: prediction.Factors.color_adjustment,
      format: (v: number) => `${v >= 0 ? '+' : ''}${(v * 100).toFixed(1)}%`,
      tooltip: '2-color bonus or 3+ color penalty'
    },
    {
      label: 'Bomb Bonus',
      value: prediction.Factors.bomb_bonus,
      format: (v: number) => `+${(v * 100).toFixed(1)}%`,
      tooltip: 'Bonus from premium cards (60%+ GIHWR)'
    },
  ];

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content prediction-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Win Rate Prediction Breakdown</h2>
          <button className="close-button" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="prediction-summary">
            <div className="predicted-win-rate-large">
              {winRatePercent}%
            </div>
            <div className="confidence-interval">
              Confidence: {minPercent}% - {maxPercent}%
            </div>
            <div className="confidence-badge">
              {prediction.Factors.confidence_level} confidence
            </div>
          </div>

          <div className="explanation-section">
            <p className="explanation-text">{prediction.Factors.explanation}</p>
          </div>

          <div className="factors-section">
            <h3>Prediction Factors</h3>
            {factors.map((factor) => (
              <div key={factor.label} className="factor-row">
                <div className="factor-label" title={factor.tooltip}>
                  {factor.label}
                  <span className="info-icon">ⓘ</span>
                </div>
                <div className="factor-value">{factor.format(factor.value)}</div>
              </div>
            ))}
          </div>

          {prediction.Factors.high_performers && prediction.Factors.high_performers.length > 0 && (
            <div className="performers-section">
              <h3>🌟 Premium Cards</h3>
              <ul className="performers-list">
                {prediction.Factors.high_performers.map((card: string, i: number) => (
                  <li key={i} className="performer-item premium">{card}</li>
                ))}
              </ul>
            </div>
          )}

          {prediction.Factors.low_performers && prediction.Factors.low_performers.length > 0 && (
            <div className="performers-section">
              <h3>⚠️ Weak Cards</h3>
              <ul className="performers-list">
                {prediction.Factors.low_performers.map((card: string, i: number) => (
                  <li key={i} className="performer-item weak">{card}</li>
                ))}
              </ul>
            </div>
          )}

          <div className="distribution-section">
            <h3>Color Distribution</h3>
            <div className="color-bars">
              {Object.entries(prediction.Factors.color_distribution || {}).map(([color, count]) => (
                <div key={color} className="color-bar">
                  <span className="color-label">{color}:</span>
                  <div className="bar-container">
                    <div
                      className="bar-fill"
                      style={{
                        width: `${(Number(count) / prediction.Factors.total_cards) * 100}%`,
                        backgroundColor: getColorCode(color)
                      }}
                    />
                  </div>
                  <span className="color-count">{count as number}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="distribution-section">
            <h3>Mana Curve</h3>
            <div className="curve-chart">
              {(() => {
                const curveEntries = Object.entries(prediction.Factors.curve_distribution || {});
                const maxCount = Math.max(...curveEntries.map(([, count]) => Number(count)), 1);
                const maxHeight = 160; // Max bar height in pixels

                return curveEntries.map(([cmc, count]) => (
                  <div key={cmc} className="curve-bar">
                    <div
                      className="curve-bar-fill"
                      style={{ height: `${(Number(count) / maxCount) * maxHeight}px` }}
                    />
                    <div className="curve-label">{cmc}</div>
                    <div className="curve-count">{count as number}</div>
                  </div>
                ));
              })()}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

function getColorCode(color: string): string {
  const colors: Record<string, string> = {
    'W': '#f0f0c0',
    'U': '#0e68ab',
    'B': '#150b00',
    'R': '#d3202a',
    'G': '#00733e',
    'C': '#ccc',
  };
  return colors[color] || '#888';
}

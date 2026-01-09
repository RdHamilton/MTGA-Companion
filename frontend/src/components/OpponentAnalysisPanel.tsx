import { useState, useEffect } from 'react';
import { opponents } from '@/services/api';
import type { OpponentAnalysis, ObservedCard, ExpectedCard, StrategicInsight } from '@/services/api';
import LoadingSpinner from './LoadingSpinner';
import './OpponentAnalysisPanel.css';

interface OpponentAnalysisPanelProps {
  matchId: string;
  isExpanded?: boolean;
  onToggle?: () => void;
}

const OpponentAnalysisPanel = ({ matchId, isExpanded = false, onToggle }: OpponentAnalysisPanelProps) => {
  const [analysis, setAnalysis] = useState<OpponentAnalysis | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'observed' | 'expected' | 'insights'>('observed');

  useEffect(() => {
    if (isExpanded && !analysis && !loading) {
      loadAnalysis();
    }
  }, [isExpanded, matchId]);

  const loadAnalysis = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await opponents.getOpponentAnalysis(matchId);
      setAnalysis(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load opponent analysis');
      console.error('Error loading opponent analysis:', err);
    } finally {
      setLoading(false);
    }
  };

  const renderColorIdentity = (colors: string) => {
    const colorMap: Record<string, string> = {
      W: 'white',
      U: 'blue',
      B: 'black',
      R: 'red',
      G: 'green',
    };

    return (
      <span className="color-identity">
        {colors.split('').map((color, index) => (
          <span key={index} className={`mana-symbol mana-${colorMap[color] || 'colorless'}`}>
            {color}
          </span>
        ))}
      </span>
    );
  };

  const renderDeckStyle = (style: string | null) => {
    return opponents.getDeckStyleDisplayName(style);
  };

  const renderConfidence = (confidence: number) => {
    const colorClass = opponents.getConfidenceColorClass(confidence);
    return (
      <span className={`confidence-value ${colorClass}`}>
        {opponents.formatConfidence(confidence)}
      </span>
    );
  };

  const renderPriority = (priority: 'high' | 'medium' | 'low') => {
    const colorClass = opponents.getPriorityColorClass(priority);
    return <span className={`priority-badge ${colorClass}`}>{priority}</span>;
  };

  const renderObservedCards = (cards: ObservedCard[]) => {
    if (cards.length === 0) {
      return <div className="empty-state">No cards observed during this match</div>;
    }

    return (
      <div className="cards-list">
        {cards.map((card, index) => (
          <div key={index} className={`card-item ${card.isSignature ? 'signature' : ''}`}>
            <div className="card-main">
              <span className="card-name">{card.cardName}</span>
              {card.isSignature && <span className="signature-badge">Signature</span>}
            </div>
            <div className="card-details">
              <span className="card-zone">{card.zone}</span>
              <span className="card-turn">Turn {card.turnFirstSeen}</span>
              {card.category && (
                <span className="card-category">{opponents.getCategoryDisplayName(card.category)}</span>
              )}
            </div>
          </div>
        ))}
      </div>
    );
  };

  const renderExpectedCards = (cards: ExpectedCard[]) => {
    if (cards.length === 0) {
      return <div className="empty-state">No expected cards data available</div>;
    }

    const unseenCards = cards.filter((c) => !c.wasSeen);
    const seenCards = cards.filter((c) => c.wasSeen);

    return (
      <div className="expected-cards">
        {unseenCards.length > 0 && (
          <div className="expected-section">
            <h4>Cards to Watch For ({unseenCards.length})</h4>
            <div className="cards-list">
              {unseenCards.slice(0, 10).map((card, index) => (
                <div key={index} className="card-item unseen">
                  <div className="card-main">
                    <span className="card-name">{card.cardName}</span>
                    <span className="inclusion-rate">{Math.round(card.inclusionRate * 100)}%</span>
                  </div>
                  <div className="card-details">
                    <span className="card-category">{opponents.getCategoryDisplayName(card.category)}</span>
                    {card.playAround && <span className="play-around">{card.playAround}</span>}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {seenCards.length > 0 && (
          <div className="expected-section">
            <h4>Confirmed Cards ({seenCards.length})</h4>
            <div className="cards-list">
              {seenCards.map((card, index) => (
                <div key={index} className="card-item seen">
                  <div className="card-main">
                    <span className="card-name">{card.cardName}</span>
                    <span className="inclusion-rate">{Math.round(card.inclusionRate * 100)}%</span>
                  </div>
                  <div className="card-details">
                    <span className="card-category">{opponents.getCategoryDisplayName(card.category)}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

  const renderInsights = (insights: StrategicInsight[]) => {
    if (insights.length === 0) {
      return <div className="empty-state">No strategic insights available</div>;
    }

    return (
      <div className="insights-list">
        {insights.map((insight, index) => (
          <div key={index} className={`insight-item priority-${insight.priority}`}>
            <div className="insight-header">
              <span className="insight-type">{insight.type}</span>
              {renderPriority(insight.priority)}
            </div>
            <p className="insight-description">{insight.description}</p>
          </div>
        ))}
      </div>
    );
  };

  return (
    <div className="opponent-analysis-panel">
      <div className="panel-header" onClick={onToggle}>
        <h3>Opponent Analysis</h3>
        <span className={`expand-icon ${isExpanded ? 'expanded' : ''}`}>
          {isExpanded ? '\u25BC' : '\u25B6'}
        </span>
      </div>

      {isExpanded && (
        <div className="panel-content">
          {loading && <LoadingSpinner message="Analyzing opponent..." />}

          {error && <div className="error-message">{error}</div>}

          {!loading && !error && analysis && (
            <>
              {/* Profile Summary */}
              {analysis.profile && (
                <div className="profile-summary">
                  <div className="profile-row">
                    <span className="profile-label">Archetype:</span>
                    <span className="profile-value">
                      {analysis.profile.detectedArchetype || 'Unknown'}
                      {analysis.profile.archetypeConfidence > 0 && (
                        <> ({renderConfidence(analysis.profile.archetypeConfidence)})</>
                      )}
                    </span>
                  </div>
                  <div className="profile-row">
                    <span className="profile-label">Colors:</span>
                    <span className="profile-value">
                      {renderColorIdentity(analysis.profile.colorIdentity)}
                    </span>
                  </div>
                  <div className="profile-row">
                    <span className="profile-label">Style:</span>
                    <span className="profile-value">{renderDeckStyle(analysis.profile.deckStyle)}</span>
                  </div>
                  <div className="profile-row">
                    <span className="profile-label">Cards Seen:</span>
                    <span className="profile-value">{analysis.profile.cardsObserved}</span>
                  </div>
                </div>
              )}

              {/* Tabs */}
              <div className="analysis-tabs">
                <button
                  className={`tab-btn ${activeTab === 'observed' ? 'active' : ''}`}
                  onClick={() => setActiveTab('observed')}
                >
                  Observed ({analysis.observedCards.length})
                </button>
                <button
                  className={`tab-btn ${activeTab === 'expected' ? 'active' : ''}`}
                  onClick={() => setActiveTab('expected')}
                >
                  Expected ({analysis.expectedCards.length})
                </button>
                <button
                  className={`tab-btn ${activeTab === 'insights' ? 'active' : ''}`}
                  onClick={() => setActiveTab('insights')}
                >
                  Insights ({analysis.strategicInsights.length})
                </button>
              </div>

              {/* Tab Content */}
              <div className="tab-content">
                {activeTab === 'observed' && renderObservedCards(analysis.observedCards)}
                {activeTab === 'expected' && renderExpectedCards(analysis.expectedCards)}
                {activeTab === 'insights' && renderInsights(analysis.strategicInsights)}
              </div>
            </>
          )}

          {!loading && !error && !analysis && (
            <div className="empty-state">Click to load opponent analysis</div>
          )}
        </div>
      )}
    </div>
  );
};

export default OpponentAnalysisPanel;

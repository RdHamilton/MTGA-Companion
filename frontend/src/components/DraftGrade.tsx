import React, { useState, useEffect } from 'react';
import { CalculateDraftGrade, GetDraftGrade } from '../../wailsjs/go/main/App';
import { grading } from '../../wailsjs/go/models';
import './DraftGrade.css';

interface DraftGradeProps {
  sessionID: string;
  onGradeCalculated?: (grade: grading.DraftGrade) => void;
  showCalculateButton?: boolean;
  compact?: boolean;
}

export const DraftGrade: React.FC<DraftGradeProps> = ({
  sessionID,
  onGradeCalculated,
  showCalculateButton = false,
  compact = false,
}) => {
  const [grade, setGrade] = useState<grading.DraftGrade | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showBreakdown, setShowBreakdown] = useState(false);

  useEffect(() => {
    loadGrade();
  }, [sessionID]);

  const loadGrade = async () => {
    try {
      setLoading(true);
      setError(null);
      const g = await GetDraftGrade(sessionID);
      setGrade(g);
    } catch (err) {
      // Grade might not exist yet - not necessarily an error
      setGrade(null);
    } finally {
      setLoading(false);
    }
  };

  const calculateGrade = async () => {
    try {
      setLoading(true);
      setError(null);
      const g = await CalculateDraftGrade(sessionID);
      setGrade(g);
      if (onGradeCalculated) {
        onGradeCalculated(g);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to calculate grade');
    } finally {
      setLoading(false);
    }
  };

  const getGradeColor = (letterGrade: string): string => {
    const letter = letterGrade.charAt(0).toUpperCase();
    switch (letter) {
      case 'A': return '#44ff88'; // green
      case 'B': return '#4a9eff'; // blue
      case 'C': return '#ffaa44'; // orange/yellow
      case 'D':
      case 'F': return '#ff4444'; // red
      default: return '#aaaaaa';
    }
  };

  if (loading) {
    return (
      <div className={`draft-grade-container ${compact ? 'compact' : ''}`}>
        <div className="loading">Loading grade...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`draft-grade-container ${compact ? 'compact' : ''}`}>
        <div className="error">{error}</div>
      </div>
    );
  }

  if (!grade) {
    if (!showCalculateButton) {
      return null;
    }
    return (
      <div className={`draft-grade-container ${compact ? 'compact' : ''}`}>
        <button onClick={calculateGrade} className="calculate-button">
          Calculate Draft Grade
        </button>
      </div>
    );
  }

  const gradeColor = getGradeColor(grade.overall_grade);

  if (compact) {
    return (
      <div
        className="draft-grade-badge"
        style={{ backgroundColor: gradeColor }}
        onClick={() => setShowBreakdown(true)}
        title={`Click to view breakdown (${grade.overall_score}/100)`}
      >
        {grade.overall_grade}
      </div>
    );
  }

  return (
    <>
      <div className="draft-grade-container">
        <div className="grade-card" onClick={() => setShowBreakdown(true)}>
          <div className="grade-letter" style={{ color: gradeColor }}>
            {grade.overall_grade}
          </div>
          <div className="grade-score">
            {Math.round(grade.overall_score)}/100
          </div>
          <div className="grade-hint">Click for breakdown</div>
        </div>
      </div>

      {showBreakdown && (
        <GradeBreakdownModal
          grade={grade}
          onClose={() => setShowBreakdown(false)}
        />
      )}
    </>
  );
};

interface GradeBreakdownModalProps {
  grade: grading.DraftGrade;
  onClose: () => void;
}

const GradeBreakdownModal: React.FC<GradeBreakdownModalProps> = ({ grade, onClose }) => {
  const getGradeColor = (letterGrade: string): string => {
    const letter = letterGrade.charAt(0).toUpperCase();
    switch (letter) {
      case 'A': return '#44ff88';
      case 'B': return '#4a9eff';
      case 'C': return '#ffaa44';
      case 'D':
      case 'F': return '#ff4444';
      default: return '#aaaaaa';
    }
  };

  const components = [
    {
      label: 'Pick Quality',
      score: grade.pick_quality_score,
      tooltip: 'How well your picks align with 17Lands data. Measures if you picked the highest win-rate cards available in each pack.'
    },
    {
      label: 'Color Discipline',
      score: grade.color_discipline_score,
      tooltip: 'How focused your color commitment is. Staying in 2-3 colors early and committing to a pair improves this score.'
    },
    {
      label: 'Deck Composition',
      score: grade.deck_composition_score,
      tooltip: 'How well-balanced your deck is. Good mana curve, creature/spell ratio, and removal spells improve this score.'
    },
    {
      label: 'Strategic Picks',
      score: grade.strategic_score,
      tooltip: 'Advanced drafting decisions like reading signals, hate-drafting bombs, and taking fixing when needed.'
    },
  ];

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Draft Grade Breakdown</h2>
          <button className="close-button" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="overall-grade-section">
            <div
              className="overall-grade-large"
              style={{ color: getGradeColor(grade.overall_grade) }}
            >
              {grade.overall_grade}
            </div>
            <div className="overall-score-large">
              {Math.round(grade.overall_score)}/100
            </div>
          </div>

          <div className="component-scores">
            <h3>Component Scores</h3>
            {components.map((component) => (
              <div key={component.label} className="component-score-row">
                <div className="component-label">
                  {component.label}
                  <span className="info-icon" title={component.tooltip}>ⓘ</span>
                </div>
                <div className="component-progress">
                  <div
                    className="component-progress-bar"
                    style={{
                      width: `${component.score}%`,
                      backgroundColor: getScoreColor(component.score)
                    }}
                  />
                </div>
                <div className="component-value">{Math.round(component.score)}</div>
              </div>
            ))}
          </div>

          {grade.best_picks && grade.best_picks.length > 0 && (
            <div className="picks-section">
              <h3 className="picks-title best">✅ Best Picks</h3>
              <ul className="picks-list">
                {grade.best_picks.map((pick, i) => (
                  <li key={i} className="pick-item best">{pick}</li>
                ))}
              </ul>
            </div>
          )}

          {grade.worst_picks && grade.worst_picks.length > 0 && (
            <div className="picks-section">
              <h3 className="picks-title worst">⚠️ Worst Picks</h3>
              <ul className="picks-list">
                {grade.worst_picks.map((pick, i) => (
                  <li key={i} className="pick-item worst">{pick}</li>
                ))}
              </ul>
            </div>
          )}

          {grade.suggestions && grade.suggestions.length > 0 && (
            <div className="suggestions-section">
              <h3>💡 Improvement Suggestions</h3>
              <ul className="suggestions-list">
                {grade.suggestions.map((suggestion, i) => (
                  <li key={i} className="suggestion-item">{suggestion}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

function getScoreColor(score: number): string {
  if (score >= 90) return '#44ff88'; // A
  if (score >= 80) return '#4a9eff'; // B
  if (score >= 70) return '#ffaa44'; // C
  return '#ff4444'; // D/F
}

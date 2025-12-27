import { gui } from '@/types/models';
import './SuggestDecksModal.css';

interface DeckSuggestionCardProps {
  suggestion: gui.SuggestedDeckResponse;
  isExpanded: boolean;
  onToggleExpand: () => void;
  onUseDeck: () => void;
  onExport: () => void;
  isApplying: boolean;
  isExporting: boolean;
  rank: number;
}

// Color name to mana symbol mapping
const colorSymbols: Record<string, string> = {
  W: 'W',
  U: 'U',
  B: 'B',
  R: 'R',
  G: 'G',
};

// Viability badge styling
const viabilityStyles: Record<string, { bg: string; text: string }> = {
  strong: { bg: '#22c55e', text: '#fff' },
  viable: { bg: '#eab308', text: '#000' },
  weak: { bg: '#ef4444', text: '#fff' },
};

export default function DeckSuggestionCard({
  suggestion,
  isExpanded,
  onToggleExpand,
  onUseDeck,
  onExport,
  isApplying,
  isExporting,
  rank,
}: DeckSuggestionCardProps) {
  const viabilityStyle = viabilityStyles[suggestion.viability] || viabilityStyles.viable;

  // Group spells by type
  const creatures = suggestion.spells.filter((s) =>
    s.typeLine.toLowerCase().includes('creature')
  );
  const nonCreatures = suggestion.spells.filter(
    (s) => !s.typeLine.toLowerCase().includes('creature')
  );

  // Calculate mana curve data for visualization
  const curveData = suggestion.analysis?.manaCurve || {};
  const curveValues = Object.values(curveData) as number[];
  const maxCurveCount = Math.max(...curveValues, 1);

  return (
    <div className={`deck-suggestion-card ${isExpanded ? 'expanded' : ''}`}>
      <div className="suggestion-header" onClick={onToggleExpand}>
        <div className="suggestion-rank">#{rank}</div>
        <div className="suggestion-colors">
          {suggestion.colorCombo.colors.map((color) => (
            <span key={color} className={`mana-pip mana-${color.toLowerCase()}`}>
              {colorSymbols[color]}
            </span>
          ))}
        </div>
        <div className="suggestion-name">{suggestion.colorCombo.name}</div>
        <div className="suggestion-score">
          {Math.round(suggestion.score * 100)}%
        </div>
        <div
          className="suggestion-viability"
          style={{ backgroundColor: viabilityStyle.bg, color: viabilityStyle.text }}
        >
          {suggestion.viability}
        </div>
        <div className="suggestion-expand-icon">{isExpanded ? '▼' : '▶'}</div>
      </div>

      {isExpanded && (
        <div className="suggestion-details">
          {/* Stats row */}
          <div className="suggestion-stats">
            <div className="stat">
              <span className="stat-value">{suggestion.totalCards}</span>
              <span className="stat-label">Cards</span>
            </div>
            <div className="stat">
              <span className="stat-value">{suggestion.analysis?.creatureCount || 0}</span>
              <span className="stat-label">Creatures</span>
            </div>
            <div className="stat">
              <span className="stat-value">{suggestion.analysis?.spellCount || 0}</span>
              <span className="stat-label">Spells</span>
            </div>
            <div className="stat">
              <span className="stat-value">
                {suggestion.analysis?.averageCMC?.toFixed(2) || 'N/A'}
              </span>
              <span className="stat-label">Avg CMC</span>
            </div>
          </div>

          {/* Mana curve visualization */}
          <div className="suggestion-curve">
            <div className="curve-title">Mana Curve</div>
            <div className="curve-bars">
              {[1, 2, 3, 4, 5, 6, 7].map((cmc) => {
                const count = curveData[cmc] || 0;
                const height = count > 0 ? (count / maxCurveCount) * 100 : 0;
                return (
                  <div key={cmc} className="curve-bar-container">
                    <div
                      className="curve-bar"
                      style={{ height: `${height}%` }}
                      title={`${count} cards at ${cmc} CMC`}
                    >
                      {count > 0 && <span className="curve-count">{count}</span>}
                    </div>
                    <div className="curve-label">{cmc}+</div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Synergies */}
          {suggestion.analysis?.synergies && suggestion.analysis.synergies.length > 0 && (
            <div className="suggestion-synergies">
              <div className="synergies-title">Synergies</div>
              <div className="synergies-list">
                {suggestion.analysis.synergies.map((synergy, i) => (
                  <span key={i} className="synergy-tag">
                    {synergy}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Card lists */}
          <div className="suggestion-cards">
            {/* Creatures */}
            {creatures.length > 0 && (
              <div className="card-section">
                <div className="section-title">Creatures ({creatures.length})</div>
                <div className="card-list">
                  {creatures.map((card) => (
                    <div key={card.cardID} className="card-item">
                      <span className="card-name">{card.name}</span>
                      <span className="card-mana">{card.manaCost}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Non-creatures */}
            {nonCreatures.length > 0 && (
              <div className="card-section">
                <div className="section-title">Spells ({nonCreatures.length})</div>
                <div className="card-list">
                  {nonCreatures.map((card) => (
                    <div key={card.cardID} className="card-item">
                      <span className="card-name">{card.name}</span>
                      <span className="card-mana">{card.manaCost}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Lands */}
            <div className="card-section">
              <div className="section-title">
                Lands ({suggestion.lands.reduce((sum, l) => sum + l.quantity, 0)})
              </div>
              <div className="card-list">
                {suggestion.lands.map((land) => (
                  <div key={land.cardID} className="card-item">
                    <span className="card-name">
                      {land.quantity}x {land.name}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="suggestion-actions">
            <button
              className="action-btn use-deck-btn"
              onClick={(e) => {
                e.stopPropagation();
                onUseDeck();
              }}
              disabled={isApplying || isExporting}
            >
              {isApplying ? 'Applying...' : 'Use This Deck'}
            </button>
            <button
              className="action-btn export-btn"
              onClick={(e) => {
                e.stopPropagation();
                onExport();
              }}
              disabled={isApplying || isExporting}
            >
              {isExporting ? 'Exporting...' : 'Export'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

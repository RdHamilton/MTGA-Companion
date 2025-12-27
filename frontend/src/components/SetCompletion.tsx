import { useState, useEffect } from 'react';
import { collection, cards } from '@/services/api';
import { models, gui } from '@/types/models';
import './SetCompletion.css';

// Rarity colors for progress bars
const rarityColors: Record<string, string> = {
  common: '#1a1a1a',
  uncommon: '#6b7c8d',
  rare: '#d4af37',
  mythic: '#e67e22',
};

// Rarity display order
const rarityOrder = ['mythic', 'rare', 'uncommon', 'common'];

interface SetCompletionProps {
  onClose?: () => void;
}

export default function SetCompletion({ onClose }: SetCompletionProps) {
  const [completionData, setCompletionData] = useState<models.SetCompletion[]>([]);
  const [setInfo, setSetInfo] = useState<Map<string, gui.SetInfo>>(new Map());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedSets, setExpandedSets] = useState<Set<string>>(new Set());
  const [sortBy, setSortBy] = useState<'name' | 'completion' | 'release'>('release');
  const [sortDesc, setSortDesc] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      setError(null);
      try {
        const [completion, sets] = await Promise.all([
          collection.getSetCompletion(),
          cards.getAllSetInfo(),
        ]);
        setCompletionData(completion || []);

        const setMap = new Map<string, gui.SetInfo>();
        (sets || []).forEach((set) => {
          setMap.set(set.code, set);
        });
        setSetInfo(setMap);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load set completion');
        console.error('Failed to load set completion:', err);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, []);

  const toggleExpanded = (setCode: string) => {
    setExpandedSets((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(setCode)) {
        newSet.delete(setCode);
      } else {
        newSet.add(setCode);
      }
      return newSet;
    });
  };

  const sortedData = [...completionData].sort((a, b) => {
    let comparison = 0;
    switch (sortBy) {
      case 'name':
        comparison = a.SetName.localeCompare(b.SetName);
        break;
      case 'completion':
        comparison = a.Percentage - b.Percentage;
        break;
      case 'release': {
        const aRelease = setInfo.get(a.SetCode)?.releasedAt || '';
        const bRelease = setInfo.get(b.SetCode)?.releasedAt || '';
        comparison = aRelease.localeCompare(bRelease);
        break;
      }
    }
    return sortDesc ? -comparison : comparison;
  });

  if (loading) {
    return (
      <div className="set-completion-panel loading">
        <div className="loading-spinner"></div>
        <p>Loading set completion...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="set-completion-panel error">
        <p>{error}</p>
      </div>
    );
  }

  return (
    <div className="set-completion-panel">
      <div className="set-completion-header">
        <h2>Set Completion</h2>
        <div className="header-controls">
          <div className="sort-controls">
            <label>Sort by:</label>
            <select
              value={`${sortBy}-${sortDesc ? 'desc' : 'asc'}`}
              onChange={(e) => {
                const [field, direction] = e.target.value.split('-');
                setSortBy(field as 'name' | 'completion' | 'release');
                setSortDesc(direction === 'desc');
              }}
            >
              <option value="release-desc">Newest First</option>
              <option value="release-asc">Oldest First</option>
              <option value="completion-desc">Most Complete</option>
              <option value="completion-asc">Least Complete</option>
              <option value="name-asc">Name (A-Z)</option>
              <option value="name-desc">Name (Z-A)</option>
            </select>
          </div>
          {onClose && (
            <button className="close-button" onClick={onClose} title="Close">
              x
            </button>
          )}
        </div>
      </div>

      <div className="set-completion-list">
        {sortedData.length === 0 ? (
          <div className="empty-state">
            <p>No set completion data available.</p>
            <p>Start collecting cards to see your progress!</p>
          </div>
        ) : (
          sortedData.map((set) => {
            const isExpanded = expandedSets.has(set.SetCode);
            const info = setInfo.get(set.SetCode);

            return (
              <div key={set.SetCode} className="set-completion-item">
                <div
                  className="set-header"
                  onClick={() => toggleExpanded(set.SetCode)}
                >
                  <div className="set-info">
                    {info?.iconSvgUri && (
                      <img
                        src={info.iconSvgUri}
                        alt={set.SetCode}
                        className="set-icon"
                      />
                    )}
                    <div className="set-details">
                      <span className="set-name">{set.SetName}</span>
                      <span className="set-code">{set.SetCode.toUpperCase()}</span>
                    </div>
                  </div>
                  <div className="set-progress">
                    <div className="progress-bar-container">
                      <div
                        className="progress-bar-fill"
                        style={{ width: `${set.Percentage}%` }}
                      />
                    </div>
                    <span className="progress-text">
                      {set.OwnedCards}/{set.TotalCards} ({set.Percentage.toFixed(1)}%)
                    </span>
                  </div>
                  <span className={`expand-icon ${isExpanded ? 'expanded' : ''}`}>
                    &gt;
                  </span>
                </div>

                {isExpanded && set.RarityBreakdown && (
                  <div className="rarity-breakdown">
                    {rarityOrder.map((rarity) => {
                      const breakdown = set.RarityBreakdown[rarity];
                      if (!breakdown || breakdown.Total === 0) return null;

                      return (
                        <div key={rarity} className="rarity-row">
                          <span
                            className="rarity-label"
                            style={{ color: rarityColors[rarity] || '#fff' }}
                          >
                            {rarity.charAt(0).toUpperCase() + rarity.slice(1)}
                          </span>
                          <div className="progress-bar-container small">
                            <div
                              className="progress-bar-fill"
                              style={{
                                width: `${breakdown.Percentage}%`,
                                backgroundColor: rarityColors[rarity] || '#4a9eff',
                              }}
                            />
                          </div>
                          <span className="rarity-count">
                            {breakdown.Owned}/{breakdown.Total}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

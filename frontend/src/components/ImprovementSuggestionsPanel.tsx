import { useState, useEffect, useCallback } from 'react';
import { notes as notesApi } from '@/services/api';
import type { ImprovementSuggestion, SuggestionType, SuggestionPriority } from '@/services/api/notes';
import { getSuggestionTypeLabel, getPriorityLabel, getPriorityColor } from '@/services/api/notes';
import './ImprovementSuggestionsPanel.css';

interface ImprovementSuggestionsPanelProps {
  deckId: string;
  onClose?: () => void;
}

const SUGGESTION_TYPE_OPTIONS: { value: SuggestionType | 'all'; label: string }[] = [
  { value: 'all', label: 'All Types' },
  { value: 'curve', label: 'Mana Curve' },
  { value: 'mana', label: 'Mana Base' },
  { value: 'removal', label: 'Removal' },
  { value: 'sequencing', label: 'Sequencing' },
  { value: 'sideboard', label: 'Sideboard' },
];

export default function ImprovementSuggestionsPanel({
  deckId,
  onClose,
}: ImprovementSuggestionsPanelProps) {
  const [suggestions, setSuggestions] = useState<ImprovementSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filterType, setFilterType] = useState<SuggestionType | 'all'>('all');
  const [showDismissed, setShowDismissed] = useState(false);
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const loadSuggestions = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await notesApi.getDeckSuggestions(deckId, !showDismissed);
      setSuggestions(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load suggestions');
      console.error('Failed to load suggestions:', err);
    } finally {
      setLoading(false);
    }
  }, [deckId, showDismissed]);

  useEffect(() => {
    loadSuggestions();
  }, [loadSuggestions]);

  const handleGenerate = async () => {
    setGenerating(true);
    setError(null);
    try {
      const newSuggestions = await notesApi.generateSuggestions(deckId);
      setSuggestions(newSuggestions || []);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to generate suggestions';
      // Check if it's the "insufficient games" error
      if (message.includes('insufficient games')) {
        setError('Not enough games played with this deck. Play at least 5 games to generate suggestions.');
      } else {
        setError(message);
      }
    } finally {
      setGenerating(false);
    }
  };

  const handleDismiss = async (suggestionId: number) => {
    try {
      await notesApi.dismissSuggestion(suggestionId);
      setSuggestions((prev) =>
        prev.map((s) => (s.id === suggestionId ? { ...s, isDismissed: true } : s))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to dismiss suggestion');
    }
  };

  const getPriorityIcon = (priority: SuggestionPriority) => {
    switch (priority) {
      case 'high':
        return '!!';
      case 'medium':
        return '!';
      case 'low':
        return '-';
      default:
        return '';
    }
  };

  const getTypeIcon = (type: SuggestionType) => {
    switch (type) {
      case 'curve':
        return '📊';
      case 'mana':
        return '💧';
      case 'removal':
        return '⚔️';
      case 'sequencing':
        return '🔄';
      case 'sideboard':
        return '📋';
      default:
        return '💡';
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  // Filter suggestions
  const filteredSuggestions = suggestions.filter((s) => {
    if (filterType !== 'all' && s.suggestionType !== filterType) return false;
    if (!showDismissed && s.isDismissed) return false;
    return true;
  });

  // Sort by priority (high first)
  const sortedSuggestions = [...filteredSuggestions].sort((a, b) => {
    const priorityOrder: Record<SuggestionPriority, number> = {
      high: 0,
      medium: 1,
      low: 2,
    };
    return priorityOrder[a.priority] - priorityOrder[b.priority];
  });

  if (loading) {
    return (
      <div className="suggestions-panel loading">
        <div className="loading-spinner"></div>
        <p>Loading suggestions...</p>
      </div>
    );
  }

  return (
    <div className="suggestions-panel">
      <div className="suggestions-header">
        <h3>Improvement Suggestions</h3>
        <div className="header-controls">
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value as SuggestionType | 'all')}
            className="type-filter"
          >
            {SUGGESTION_TYPE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          {onClose && (
            <button className="close-button" onClick={onClose} title="Close">
              x
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="error-banner">
          <span>{error}</span>
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      <div className="suggestions-actions">
        <button
          className="generate-btn"
          onClick={handleGenerate}
          disabled={generating}
        >
          {generating ? 'Analyzing...' : 'Generate New Suggestions'}
        </button>
        <label className="show-dismissed">
          <input
            type="checkbox"
            checked={showDismissed}
            onChange={(e) => setShowDismissed(e.target.checked)}
          />
          Show dismissed
        </label>
      </div>

      <div className="suggestions-list">
        {sortedSuggestions.length === 0 ? (
          <div className="empty-state">
            <p>No suggestions yet.</p>
            <p>Click &quot;Generate New Suggestions&quot; to analyze your play patterns!</p>
            <p className="hint">Requires at least 5 games played with this deck.</p>
          </div>
        ) : (
          sortedSuggestions.map((suggestion) => (
            <div
              key={suggestion.id}
              className={`suggestion-item ${suggestion.isDismissed ? 'dismissed' : ''} ${
                expandedId === suggestion.id ? 'expanded' : ''
              }`}
            >
              <div
                className="suggestion-main"
                onClick={() => setExpandedId(expandedId === suggestion.id ? null : suggestion.id)}
              >
                <div className="suggestion-icon">{getTypeIcon(suggestion.suggestionType)}</div>
                <div className="suggestion-content">
                  <div className="suggestion-title-row">
                    <span className={`priority-badge ${suggestion.priority}`}>
                      {getPriorityIcon(suggestion.priority)}
                    </span>
                    <h4 className="suggestion-title">{suggestion.title}</h4>
                  </div>
                  <span className="suggestion-type">
                    {getSuggestionTypeLabel(suggestion.suggestionType)}
                  </span>
                </div>
                <span className={`expand-icon ${expandedId === suggestion.id ? 'expanded' : ''}`}>
                  ▶
                </span>
              </div>

              {expandedId === suggestion.id && (
                <div className="suggestion-details">
                  <p className="suggestion-description">{suggestion.description}</p>
                  <div className="suggestion-meta">
                    <span className={`priority ${getPriorityColor(suggestion.priority)}`}>
                      {getPriorityLabel(suggestion.priority)}
                    </span>
                    <span className="date">Generated: {formatDate(suggestion.createdAt)}</span>
                  </div>
                  {!suggestion.isDismissed && (
                    <div className="suggestion-actions">
                      <button
                        className="dismiss-btn"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDismiss(suggestion.id);
                        }}
                      >
                        Dismiss
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}

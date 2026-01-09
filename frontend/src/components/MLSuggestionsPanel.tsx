import { useState, useEffect, useCallback } from 'react';
import { mlSuggestions as mlApi } from '@/services/api';
import type {
  MLSuggestion,
  MLSuggestionType,
  MLSuggestionResult,
} from '@/services/api/mlSuggestions';
import {
  getMLSuggestionTypeLabel,
  getMLSuggestionTypeIcon,
  formatConfidence,
  formatWinRateChange,
  getConfidenceColor,
  parseReasons,
} from '@/services/api/mlSuggestions';
import './MLSuggestionsPanel.css';

interface MLSuggestionsPanelProps {
  deckId: string;
  onClose?: () => void;
}

const SUGGESTION_TYPE_OPTIONS: { value: MLSuggestionType | 'all'; label: string }[] = [
  { value: 'all', label: 'All Types' },
  { value: 'add', label: 'Add Card' },
  { value: 'remove', label: 'Remove Card' },
  { value: 'swap', label: 'Swap Cards' },
];

export default function MLSuggestionsPanel({
  deckId,
  onClose,
}: MLSuggestionsPanelProps) {
  const [suggestions, setSuggestions] = useState<MLSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filterType, setFilterType] = useState<MLSuggestionType | 'all'>('all');
  const [showDismissed, setShowDismissed] = useState(false);
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const loadSuggestions = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await mlApi.getMLSuggestions(deckId, !showDismissed);
      setSuggestions(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load ML suggestions');
      console.error('Failed to load ML suggestions:', err);
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
      const results: MLSuggestionResult[] = await mlApi.generateMLSuggestions(deckId);
      setSuggestions(results?.map((r) => r.suggestion) || []);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to generate ML suggestions';
      if (message.includes('no synergy data')) {
        setError('No synergy data available. Process match history first to enable ML suggestions.');
      } else {
        setError(message);
      }
    } finally {
      setGenerating(false);
    }
  };

  const handleDismiss = async (suggestionId: number) => {
    try {
      await mlApi.dismissMLSuggestion(suggestionId);
      setSuggestions((prev) =>
        prev.map((s) => (s.id === suggestionId ? { ...s, isDismissed: true } : s))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to dismiss suggestion');
    }
  };

  const handleApply = async (suggestionId: number) => {
    try {
      await mlApi.applyMLSuggestion(suggestionId);
      setSuggestions((prev) =>
        prev.map((s) => (s.id === suggestionId ? { ...s, wasApplied: true } : s))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to apply suggestion');
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

  // Sort by confidence (high first)
  const sortedSuggestions = [...filteredSuggestions].sort(
    (a, b) => b.confidence - a.confidence
  );

  if (loading) {
    return (
      <div className="ml-suggestions-panel loading">
        <div className="loading-spinner"></div>
        <p>Loading ML suggestions...</p>
      </div>
    );
  }

  return (
    <div className="ml-suggestions-panel">
      <div className="suggestions-header">
        <h3>ML-Powered Suggestions</h3>
        <div className="header-controls">
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value as MLSuggestionType | 'all')}
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
          {generating ? 'Analyzing synergies...' : 'Generate ML Suggestions'}
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
            <p>No ML suggestions yet.</p>
            <p>Click &quot;Generate ML Suggestions&quot; to analyze card synergies!</p>
            <p className="hint">Based on win rates of card combinations in your matches.</p>
          </div>
        ) : (
          sortedSuggestions.map((suggestion) => (
            <div
              key={suggestion.id}
              className={`ml-suggestion-item ${suggestion.isDismissed ? 'dismissed' : ''} ${
                suggestion.wasApplied ? 'applied' : ''
              } ${expandedId === suggestion.id ? 'expanded' : ''}`}
            >
              <div
                className="suggestion-main"
                onClick={() => setExpandedId(expandedId === suggestion.id ? null : suggestion.id)}
              >
                <div className="suggestion-type-icon">
                  {getMLSuggestionTypeIcon(suggestion.suggestionType)}
                </div>
                <div className="suggestion-content">
                  <div className="suggestion-title-row">
                    <h4 className="suggestion-title">{suggestion.title}</h4>
                    <span className={`confidence-badge ${getConfidenceColor(suggestion.confidence)}`}>
                      {formatConfidence(suggestion.confidence)}
                    </span>
                  </div>
                  <div className="suggestion-meta-row">
                    <span className="suggestion-type">
                      {getMLSuggestionTypeLabel(suggestion.suggestionType)}
                    </span>
                    {suggestion.expectedWinRateChange !== 0 && (
                      <span
                        className={`win-rate-change ${
                          suggestion.expectedWinRateChange >= 0 ? 'positive' : 'negative'
                        }`}
                      >
                        {formatWinRateChange(suggestion.expectedWinRateChange)}
                      </span>
                    )}
                  </div>
                </div>
                <span className={`expand-icon ${expandedId === suggestion.id ? 'expanded' : ''}`}>
                  {'\u25B6'}
                </span>
              </div>

              {expandedId === suggestion.id && (
                <div className="suggestion-details">
                  {suggestion.description && (
                    <p className="suggestion-description">{suggestion.description}</p>
                  )}

                  {suggestion.suggestionType === 'swap' && suggestion.swapForCardName && (
                    <div className="swap-info">
                      <span className="swap-from">{suggestion.cardName}</span>
                      <span className="swap-arrow">{'\u2192'}</span>
                      <span className="swap-to">{suggestion.swapForCardName}</span>
                    </div>
                  )}

                  {suggestion.reasoning && (
                    <div className="reasoning-list">
                      <h5>Reasons:</h5>
                      <ul>
                        {parseReasons(suggestion.reasoning).map((reason, idx) => (
                          <li key={idx} className="reason-item">
                            <span className="reason-type">{reason.type}:</span>
                            <span className="reason-desc">{reason.description}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}

                  <div className="suggestion-footer">
                    <span className="date">Generated: {formatDate(suggestion.createdAt)}</span>
                    {suggestion.wasApplied && suggestion.appliedAt && (
                      <span className="applied-date">
                        Applied: {formatDate(suggestion.appliedAt)}
                      </span>
                    )}
                  </div>

                  {!suggestion.isDismissed && !suggestion.wasApplied && (
                    <div className="suggestion-actions">
                      <button
                        className="apply-btn"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleApply(suggestion.id);
                        }}
                      >
                        Mark Applied
                      </button>
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

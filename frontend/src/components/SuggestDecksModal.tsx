import { useState, useEffect } from 'react';
import {
  SuggestDecks,
  ApplySuggestedDeck,
  ExportSuggestedDeck,
} from '../../wailsjs/go/main/App';
import { gui } from '../../wailsjs/go/models';
import DeckSuggestionCard from './DeckSuggestionCard';
import './SuggestDecksModal.css';

interface SuggestDecksModalProps {
  isOpen: boolean;
  onClose: () => void;
  draftEventID: string;
  currentDeckID: string;
  deckName: string;
  onDeckApplied: () => void;
}

export default function SuggestDecksModal({
  isOpen,
  onClose,
  draftEventID,
  currentDeckID,
  deckName,
  onDeckApplied,
}: SuggestDecksModalProps) {
  const [suggestions, setSuggestions] = useState<gui.SuggestDecksResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);
  const [applying, setApplying] = useState(false);
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    if (isOpen && draftEventID) {
      loadSuggestions();
    }
  }, [isOpen, draftEventID]);

  const loadSuggestions = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await SuggestDecks(draftEventID);
      if (response.error) {
        setError(response.error);
      } else {
        setSuggestions(response);
        // Auto-expand the best deck
        if (response.suggestions && response.suggestions.length > 0) {
          setExpandedIndex(0);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load suggestions');
    } finally {
      setLoading(false);
    }
  };

  const handleApplyDeck = async (suggestion: gui.SuggestedDeckResponse) => {
    setApplying(true);
    try {
      await ApplySuggestedDeck(currentDeckID, suggestion);
      onDeckApplied();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to apply deck');
    } finally {
      setApplying(false);
    }
  };

  const handleExportDeck = async (suggestion: gui.SuggestedDeckResponse) => {
    setExporting(true);
    try {
      const exportName = `${deckName} - ${suggestion.colorCombo.name}`;
      await ExportSuggestedDeck(suggestion, exportName);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export deck');
    } finally {
      setExporting(false);
    }
  };

  const handleToggleExpand = (index: number) => {
    setExpandedIndex(expandedIndex === index ? null : index);
  };

  if (!isOpen) return null;

  return (
    <div className="suggest-decks-overlay" onClick={onClose}>
      <div className="suggest-decks-modal" onClick={(e) => e.stopPropagation()}>
        <div className="suggest-decks-header">
          <h2>Suggested Decks</h2>
          <button className="close-button" onClick={onClose}>
            &times;
          </button>
        </div>

        <div className="suggest-decks-content">
          {loading && (
            <div className="suggest-decks-loading">
              <div className="loading-spinner"></div>
              <p>Analyzing your draft pool...</p>
            </div>
          )}

          {error && (
            <div className="suggest-decks-error">
              <p>{error}</p>
              <button onClick={loadSuggestions}>Try Again</button>
            </div>
          )}

          {!loading && !error && suggestions && (
            <>
              <div className="suggest-decks-summary">
                <p>
                  Found <strong>{suggestions.viableCombos}</strong> viable color combinations
                  out of {suggestions.totalCombos} possible.
                </p>
                {suggestions.bestCombo && (
                  <p className="best-combo">
                    Best option: <strong>{suggestions.bestCombo.name}</strong>
                  </p>
                )}
              </div>

              {suggestions.suggestions && suggestions.suggestions.length > 0 ? (
                <div className="suggest-decks-list">
                  {suggestions.suggestions.map((suggestion, index) => (
                    <DeckSuggestionCard
                      key={`${suggestion.colorCombo.name}-${index}`}
                      suggestion={suggestion}
                      isExpanded={expandedIndex === index}
                      onToggleExpand={() => handleToggleExpand(index)}
                      onUseDeck={() => handleApplyDeck(suggestion)}
                      onExport={() => handleExportDeck(suggestion)}
                      isApplying={applying}
                      isExporting={exporting}
                      rank={index + 1}
                    />
                  ))}
                </div>
              ) : (
                <div className="suggest-decks-empty">
                  <p>No viable deck combinations found.</p>
                  <p>Try adding more cards to your draft pool.</p>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

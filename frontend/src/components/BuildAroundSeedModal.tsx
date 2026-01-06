import { useState, useCallback, useRef, useEffect } from 'react';
import { decks, cards as cardsApi } from '@/services/api';
import type {
  BuildAroundSeedResponse,
  CardWithOwnership,
  SuggestedLandResponse,
  IterativeBuildAroundResponse,
  LiveDeckAnalysis,
} from '@/services/api/decks';
import { models } from '@/types/models';
import './BuildAroundSeedModal.css';

interface BuildAroundSeedModalProps {
  isOpen: boolean;
  onClose: () => void;
  onApplyDeck: (suggestions: CardWithOwnership[], lands: SuggestedLandResponse[]) => void;
  onCardAdded?: (card: CardWithOwnership) => void;
  onCardRemoved?: (cardId: number) => void;
  onFinishDeck?: (lands: SuggestedLandResponse[]) => void;
  currentDeckCards?: number[];
  deckCards?: models.DeckCard[];
}

interface SearchResult {
  arenaID: number;
  name: string;
  manaCost?: string;
  types?: string[];
  imageURI?: string;
  colors?: string[];
}

export default function BuildAroundSeedModal({
  isOpen,
  onClose,
  onApplyDeck,
  onCardAdded,
  onCardRemoved,
  onFinishDeck,
  currentDeckCards = [],
  deckCards = [],
}: BuildAroundSeedModalProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [selectedCard, setSelectedCard] = useState<SearchResult | null>(null);
  const [suggestions, setSuggestions] = useState<BuildAroundSeedResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [budgetMode, setBudgetMode] = useState(false);
  const [applying, setApplying] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Iterative mode state
  const [iterativeMode, setIterativeMode] = useState(false);
  const [seedCardId, setSeedCardId] = useState<number | null>(null);
  const [iterativeSuggestions, setIterativeSuggestions] = useState<CardWithOwnership[]>([]);
  const [deckAnalysis, setDeckAnalysis] = useState<LiveDeckAnalysis | null>(null);
  const [slotsRemaining, setSlotsRemaining] = useState(60);
  const [landSuggestions, setLandSuggestions] = useState<SuggestedLandResponse[]>([]);

  // Cleanup debounce timer on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  // Fetch suggestions when in iterative mode and deck changes
  const fetchIterativeSuggestions = useCallback(async () => {
    if (!seedCardId || !iterativeMode) return;

    setLoading(true);
    try {
      const response: IterativeBuildAroundResponse = await decks.suggestNextCards({
        seed_card_id: seedCardId,
        deck_card_ids: currentDeckCards,
        max_results: 15,
        budget_mode: budgetMode,
      });

      setIterativeSuggestions(response.suggestions);
      setDeckAnalysis(response.deckAnalysis);
      setSlotsRemaining(response.slotsRemaining);
      setLandSuggestions(response.landSuggestions);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get suggestions');
    } finally {
      setLoading(false);
    }
  }, [seedCardId, currentDeckCards, budgetMode, iterativeMode]);

  // Debounced fetch when deck changes in iterative mode
  useEffect(() => {
    if (!iterativeMode || !seedCardId) return;

    const timer = setTimeout(fetchIterativeSuggestions, 300);
    return () => clearTimeout(timer);
  }, [currentDeckCards, iterativeMode, seedCardId, fetchIterativeSuggestions]);

  const handleSearch = useCallback((query: string) => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    if (query.length < 2) {
      setSearchResults([]);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      setSearching(true);
      try {
        const results = await cardsApi.searchCards({ query, limit: 10 });
        setSearchResults(
          results
            .filter(card => card.ArenaID && !isNaN(parseInt(card.ArenaID, 10)))
            .map(card => ({
              arenaID: parseInt(card.ArenaID, 10),
              name: card.Name,
              manaCost: card.ManaCost,
              types: card.Types,
              imageURI: card.ImageURL,
              colors: card.Colors,
            }))
        );
      } catch {
        setSearchResults([]);
      } finally {
        setSearching(false);
      }
    }, 300);
  }, []);

  const handleSelectCard = (card: SearchResult) => {
    setSelectedCard(card);
    setSearchQuery(card.name);
    setSearchResults([]);
    setSuggestions(null);
  };

  // Start iterative building mode
  const handleStartBuilding = async () => {
    if (!selectedCard) return;

    setSeedCardId(selectedCard.arenaID);
    setIterativeMode(true);
    setLoading(true);
    setError(null);

    try {
      const response = await decks.suggestNextCards({
        seed_card_id: selectedCard.arenaID,
        deck_card_ids: currentDeckCards,
        max_results: 15,
        budget_mode: budgetMode,
      });

      setIterativeSuggestions(response.suggestions);
      setDeckAnalysis(response.deckAnalysis);
      setSlotsRemaining(response.slotsRemaining);
      setLandSuggestions(response.landSuggestions);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start building');
      setIterativeMode(false);
    } finally {
      setLoading(false);
    }
  };

  // Quick build (original one-shot mode)
  const handleBuildAround = async () => {
    if (!selectedCard) return;

    setLoading(true);
    setError(null);
    setSuggestions(null);

    try {
      const response = await decks.buildAroundSeed({
        seed_card_id: selectedCard.arenaID,
        max_results: 40,
        budget_mode: budgetMode,
        set_restriction: 'all',
      });
      setSuggestions(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate suggestions');
    } finally {
      setLoading(false);
    }
  };

  // Handle picking a card in iterative mode
  const handlePickCard = (card: CardWithOwnership) => {
    if (onCardAdded) {
      onCardAdded(card);
      // Suggestions will auto-refresh via useEffect when currentDeckCards changes
    }
  };

  // Handle finishing the deck
  const handleFinishDeck = () => {
    if (onFinishDeck) {
      onFinishDeck(landSuggestions);
    }
    handleClose();
  };

  const handleApply = async () => {
    if (!suggestions) return;

    setApplying(true);
    try {
      onApplyDeck(suggestions.suggestions, suggestions.lands);
      onClose();
    } finally {
      setApplying(false);
    }
  };

  const handleClose = () => {
    // Reset state
    setIterativeMode(false);
    setSeedCardId(null);
    setIterativeSuggestions([]);
    setDeckAnalysis(null);
    setLandSuggestions([]);
    onClose();
  };

  const handleClear = () => {
    setSearchQuery('');
    setSearchResults([]);
    setSelectedCard(null);
    setSuggestions(null);
    setError(null);
    setIterativeMode(false);
    setSeedCardId(null);
    setIterativeSuggestions([]);
    setDeckAnalysis(null);
  };

  if (!isOpen) return null;

  const renderColorPips = (colors: string[] | undefined) => {
    if (!colors || colors.length === 0) return null;
    return (
      <div className="color-pips">
        {colors.map((color, i) => (
          <span key={i} className={`mana-pip mana-${color.toLowerCase()}`}>
            {color}
          </span>
        ))}
      </div>
    );
  };

  const renderOwnershipBadge = (card: CardWithOwnership) => {
    if (card.inCollection) {
      return <span className="ownership-badge owned">Own {card.ownedCount}</span>;
    }
    return <span className="ownership-badge needed">Need {card.neededCount}</span>;
  };

  // Iterative mode UI
  if (iterativeMode && selectedCard) {
    return (
      <div className="build-around-overlay" onClick={handleClose}>
        <div className="build-around-modal iterative-mode" onClick={(e) => e.stopPropagation()}>
          <div className="build-around-header">
            <h2>Building: {selectedCard.name}</h2>
            <button className="close-button" onClick={handleClose}>
              &times;
            </button>
          </div>

          <div className="build-around-content">
            {/* Status Bar */}
            <div className="iterative-status-bar">
              <span className="slots-remaining">{slotsRemaining} slots remaining</span>
              {deckAnalysis && renderColorPips(deckAnalysis.colorIdentity)}
              <label className="option-checkbox">
                <input
                  type="checkbox"
                  checked={budgetMode}
                  onChange={(e) => setBudgetMode(e.target.checked)}
                />
                <span>Budget Mode</span>
              </label>
            </div>

            {/* Error State */}
            {error && (
              <div className="build-around-error">
                <p>{error}</p>
                <button onClick={fetchIterativeSuggestions}>Try Again</button>
              </div>
            )}

            {/* Loading */}
            {loading && <div className="loading-indicator">Loading suggestions...</div>}

            {/* Suggestions Grid */}
            {!loading && iterativeSuggestions.length > 0 && (
              <div className="iterative-suggestions">
                <h3>Click a card to add 1 copy to your deck</h3>
                <div className="suggestions-clickable-grid">
                  {iterativeSuggestions.map(card => (
                    <div
                      key={card.cardID}
                      className="clickable-suggestion-card"
                      onClick={() => handlePickCard(card)}
                    >
                      {card.imageURI ? (
                        <img src={card.imageURI} alt={card.name} className="suggestion-image" />
                      ) : (
                        <div className="suggestion-placeholder">
                          <span className="card-name">{card.name}</span>
                          <span className="card-mana">{card.manaCost}</span>
                        </div>
                      )}
                      <div className="suggestion-overlay">
                        <span className="card-name">{card.name}</span>
                        {renderOwnershipBadge(card)}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Current Deck Cards */}
            {deckCards.length > 0 && (
              <div className="current-deck-cards">
                <h4>Current Deck ({deckCards.reduce((sum, c) => sum + c.Quantity, 0)} cards)</h4>
                <div className="deck-cards-list">
                  {deckCards
                    .filter(card => card.Board === 'main')
                    .map(card => (
                      <div key={`${card.CardID}-${card.Board}`} className="deck-card-item">
                        <span className="card-quantity">{card.Quantity}x</span>
                        <span className="card-name">{card.Name}</span>
                        {onCardRemoved && (
                          <button
                            className="remove-card-btn"
                            onClick={() => onCardRemoved(card.CardID)}
                            title="Remove 1 copy"
                          >
                            −
                          </button>
                        )}
                      </div>
                    ))}
                </div>
              </div>
            )}

            {/* Deck Analysis */}
            {deckAnalysis && (
              <div className="live-deck-analysis">
                <h4>Deck Analysis</h4>
                <div className="analysis-row">
                  <span>Total Cards: {deckAnalysis.totalCards}</span>
                  <span>Recommended Lands: {deckAnalysis.recommendedLandCount}</span>
                </div>
                {deckAnalysis.themes.length > 0 && (
                  <div className="themes-section">
                    {deckAnalysis.themes.map((theme, i) => (
                      <span key={i} className="theme-tag">{theme}</span>
                    ))}
                  </div>
                )}
                {/* Mana Curve */}
                <div className="mana-curve-mini">
                  <span>Curve: </span>
                  {Object.entries(deckAnalysis.currentCurve)
                    .sort(([a], [b]) => parseInt(a) - parseInt(b))
                    .map(([cmc, count]) => (
                      <span key={cmc} className="curve-pip">
                        {cmc}:{count}
                      </span>
                    ))}
                </div>
              </div>
            )}

            {/* Land Suggestions */}
            {landSuggestions.length > 0 && (
              <div className="land-suggestions-preview">
                <h4>Lands ({landSuggestions.reduce((sum, l) => sum + l.quantity, 0)})</h4>
                <div className="land-list">
                  {landSuggestions.map(land => (
                    <span key={land.cardID} className="land-item">
                      {land.name} &times;{land.quantity}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="suggestions-actions">
              <button
                className="action-btn apply-btn"
                onClick={handleFinishDeck}
                disabled={slotsRemaining > 30}
              >
                Finish Deck (Add Lands)
              </button>
              <button
                className="action-btn cancel-btn"
                onClick={handleClose}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Original seed selection UI
  return (
    <div className="build-around-overlay" onClick={handleClose}>
      <div className="build-around-modal" onClick={(e) => e.stopPropagation()}>
        <div className="build-around-header">
          <h2>Build Around Card</h2>
          <button className="close-button" onClick={handleClose}>
            &times;
          </button>
        </div>

        <div className="build-around-content">
          {/* Card Search Section */}
          <div className="search-section">
            <div className="search-input-container">
              <input
                type="text"
                placeholder="Search for a card to build around..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  handleSearch(e.target.value);
                }}
                className="search-input"
              />
              {selectedCard && (
                <button className="clear-button" onClick={handleClear}>
                  Clear
                </button>
              )}
            </div>

            {/* Search Results Dropdown */}
            {searchResults.length > 0 && (
              <div className="search-results">
                {searchResults.map((card) => (
                  <div
                    key={card.arenaID}
                    className="search-result-item"
                    onClick={() => handleSelectCard(card)}
                  >
                    <span className="result-name">{card.name}</span>
                    {renderColorPips(card.colors)}
                    <span className="result-type">{card.types?.join(' ')}</span>
                  </div>
                ))}
              </div>
            )}

            {searching && <div className="searching-indicator">Searching...</div>}
          </div>

          {/* Selected Card Preview */}
          {selectedCard && (
            <div className="selected-card-section">
              <div className="selected-card">
                {selectedCard.imageURI ? (
                  <img
                    src={selectedCard.imageURI}
                    alt={selectedCard.name}
                    className="card-image"
                  />
                ) : (
                  <div className="card-placeholder">
                    <span>{selectedCard.name}</span>
                  </div>
                )}
                <div className="selected-card-info">
                  <h3>{selectedCard.name}</h3>
                  <p className="selected-type">{selectedCard.types?.join(' ')}</p>
                  {renderColorPips(selectedCard.colors)}
                </div>
              </div>

              {/* Options */}
              <div className="build-options">
                <label className="option-checkbox">
                  <input
                    type="checkbox"
                    checked={budgetMode}
                    onChange={(e) => setBudgetMode(e.target.checked)}
                  />
                  <span>Budget Mode (only cards in collection)</span>
                </label>
              </div>

              {/* Build Mode Buttons */}
              <div className="build-mode-buttons">
                {onCardAdded && onFinishDeck && (
                  <button
                    className="build-button primary"
                    onClick={handleStartBuilding}
                    disabled={loading}
                  >
                    {loading ? 'Starting...' : 'Start Building (Pick Cards)'}
                  </button>
                )}
                <button
                  className="build-button secondary"
                  onClick={handleBuildAround}
                  disabled={loading}
                >
                  {loading ? 'Generating...' : 'Quick Build (Auto-fill Deck)'}
                </button>
              </div>
            </div>
          )}

          {/* Error State */}
          {error && (
            <div className="build-around-error">
              <p>{error}</p>
              <button onClick={handleBuildAround}>Try Again</button>
            </div>
          )}

          {/* Suggestions Results (Quick Build Mode) */}
          {suggestions && (
            <div className="suggestions-section">
              {/* Analysis Summary */}
              <div className="analysis-summary">
                <div className="summary-header">
                  <h3>Deck Analysis</h3>
                  {renderColorPips(suggestions.analysis.colorIdentity)}
                </div>
                <div className="summary-stats">
                  <div className="stat">
                    <span className="stat-value">{suggestions.analysis.totalCards}</span>
                    <span className="stat-label">Total Cards</span>
                  </div>
                  <div className="stat">
                    <span className="stat-value">{suggestions.analysis.inCollectionCount}</span>
                    <span className="stat-label">In Collection</span>
                  </div>
                  <div className="stat">
                    <span className="stat-value">{suggestions.analysis.missingCount}</span>
                    <span className="stat-label">Missing</span>
                  </div>
                  <div className="stat">
                    <span className="stat-value">{suggestions.analysis.suggestedLandCount}</span>
                    <span className="stat-label">Lands</span>
                  </div>
                </div>

                {/* Themes & Keywords */}
                {(suggestions.analysis.themes?.length > 0 || suggestions.analysis.keywords?.length > 0) && (
                  <div className="themes-section">
                    {suggestions.analysis.themes?.map((theme, i) => (
                      <span key={`theme-${i}`} className="theme-tag">{theme}</span>
                    ))}
                    {suggestions.analysis.keywords?.slice(0, 5).map((keyword, i) => (
                      <span key={`keyword-${i}`} className="keyword-tag">{keyword}</span>
                    ))}
                  </div>
                )}

                {/* Wildcard Cost */}
                {suggestions.analysis.missingCount > 0 && (
                  <div className="wildcard-cost">
                    <span className="cost-label">Wildcards needed:</span>
                    {Object.entries(suggestions.analysis.missingWildcardCost || {}).map(([rarity, count]) => (
                      count > 0 && (
                        <span key={rarity} className={`wildcard-badge ${rarity}`}>
                          {count} {rarity}
                        </span>
                      )
                    ))}
                  </div>
                )}
              </div>

              {/* Suggested Cards */}
              <div className="suggestions-grid">
                {/* Creatures */}
                <div className="card-category">
                  <h4>Creatures</h4>
                  <div className="card-list">
                    {suggestions.suggestions
                      .filter(c => c.typeLine?.toLowerCase().includes('creature'))
                      .slice(0, 15)
                      .map(card => (
                        <div key={card.cardID} className="suggestion-card">
                          <span className="card-name">{card.name}</span>
                          <span className="card-mana">{card.manaCost}</span>
                          {renderOwnershipBadge(card)}
                        </div>
                      ))}
                  </div>
                </div>

                {/* Spells */}
                <div className="card-category">
                  <h4>Spells</h4>
                  <div className="card-list">
                    {suggestions.suggestions
                      .filter(c => !c.typeLine?.toLowerCase().includes('creature') && !c.typeLine?.toLowerCase().includes('land'))
                      .slice(0, 15)
                      .map(card => (
                        <div key={card.cardID} className="suggestion-card">
                          <span className="card-name">{card.name}</span>
                          <span className="card-mana">{card.manaCost}</span>
                          {renderOwnershipBadge(card)}
                        </div>
                      ))}
                  </div>
                </div>

                {/* Lands */}
                <div className="card-category">
                  <h4>Lands ({suggestions.analysis.suggestedLandCount})</h4>
                  <div className="card-list">
                    {suggestions.lands.map(land => (
                      <div key={land.cardID} className="suggestion-card land-card">
                        <span className="card-name">{land.name}</span>
                        <span className="land-quantity">&times;{land.quantity}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="suggestions-actions">
                <button
                  className="action-btn apply-btn"
                  onClick={handleApply}
                  disabled={applying}
                >
                  {applying ? 'Applying...' : 'Apply to Current Deck'}
                </button>
                <button
                  className="action-btn cancel-btn"
                  onClick={handleClose}
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

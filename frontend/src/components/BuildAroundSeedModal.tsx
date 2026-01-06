import { useState, useCallback } from 'react';
import { decks, cards as cardsApi } from '@/services/api';
import type { BuildAroundSeedResponse, CardWithOwnership, SuggestedLandResponse } from '@/services/api/decks';
import './BuildAroundSeedModal.css';

interface BuildAroundSeedModalProps {
  isOpen: boolean;
  onClose: () => void;
  onApplyDeck: (suggestions: CardWithOwnership[], lands: SuggestedLandResponse[]) => void;
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

  const handleSearch = useCallback(async (query: string) => {
    if (query.length < 2) {
      setSearchResults([]);
      return;
    }

    setSearching(true);
    try {
      const results = await cardsApi.searchCards({ query, limit: 10 });
      setSearchResults(results.map(card => ({
        arenaID: parseInt(card.ArenaID, 10) || 0,
        name: card.Name,
        manaCost: card.ManaCost,
        types: card.Types,
        imageURI: card.ImageURL,
        colors: card.Colors,
      })));
    } catch {
      setSearchResults([]);
    } finally {
      setSearching(false);
    }
  }, []);

  const handleSelectCard = (card: SearchResult) => {
    setSelectedCard(card);
    setSearchQuery(card.name);
    setSearchResults([]);
    setSuggestions(null);
  };

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

  const handleClear = () => {
    setSearchQuery('');
    setSearchResults([]);
    setSelectedCard(null);
    setSuggestions(null);
    setError(null);
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

  return (
    <div className="build-around-overlay" onClick={onClose}>
      <div className="build-around-modal" onClick={(e) => e.stopPropagation()}>
        <div className="build-around-header">
          <h2>Build Around Card</h2>
          <button className="close-button" onClick={onClose}>
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

              <button
                className="build-button"
                onClick={handleBuildAround}
                disabled={loading}
              >
                {loading ? 'Generating Suggestions...' : 'Build Around This Card'}
              </button>
            </div>
          )}

          {/* Error State */}
          {error && (
            <div className="build-around-error">
              <p>{error}</p>
              <button onClick={handleBuildAround}>Try Again</button>
            </div>
          )}

          {/* Suggestions Results */}
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
                  onClick={onClose}
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

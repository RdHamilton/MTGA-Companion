import { useState, useCallback, useRef, useEffect, useMemo } from 'react';
import { decks, cards as cardsApi } from '@/services/api';
import type {
  BuildAroundSeedResponse,
  CardWithOwnership,
  SuggestedLandResponse,
  IterativeBuildAroundResponse,
  LiveDeckAnalysis,
  GenerateCompleteDeckResponse,
  ArchetypeProfile,
} from '@/services/api/decks';
import { models } from '@/types/models';
import './BuildAroundSeedModal.css';

type ArchetypeKey = 'aggro' | 'midrange' | 'control';

interface HoverPreview {
  card: CardWithOwnership;
  position: { x: number; y: number };
}

interface BuildAroundSeedModalProps {
  isOpen: boolean;
  onClose: () => void;
  onApplyDeck: (suggestions: CardWithOwnership[], lands: SuggestedLandResponse[]) => void;
  onCardAdded?: (card: CardWithOwnership) => void;
  onCardRemoved?: (cardId: number) => void;
  onFinishDeck?: (lands: SuggestedLandResponse[]) => void;
  currentDeckCards?: number[];
  deckCards?: models.DeckCard[];
  /** When true, skips seed selection and uses current deck cards directly */
  useDeckCardsAsSeed?: boolean;
}

interface SearchResult {
  arenaID: number;
  name: string;
  manaCost?: string;
  types?: string[];
  imageURI?: string;
  colors?: string[];
  cmc?: number;
}

interface ColorFilter {
  W: boolean;
  U: boolean;
  B: boolean;
  R: boolean;
  G: boolean;
}

interface TypeFilter {
  creature: boolean;
  instant: boolean;
  sorcery: boolean;
  enchantment: boolean;
  artifact: boolean;
  planeswalker: boolean;
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
  useDeckCardsAsSeed = false,
}: BuildAroundSeedModalProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [rawSearchResults, setRawSearchResults] = useState<SearchResult[]>([]); // Unfiltered results from API
  const [selectedCard, setSelectedCard] = useState<SearchResult | null>(null);
  const [suggestions, setSuggestions] = useState<BuildAroundSeedResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [budgetMode, setBudgetMode] = useState(false);
  const [applying, setApplying] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Filter state
  const [colorFilter, setColorFilter] = useState<ColorFilter>({
    W: false, U: false, B: false, R: false, G: false,
  });
  const [typeFilter, setTypeFilter] = useState<TypeFilter>({
    creature: false, instant: false, sorcery: false,
    enchantment: false, artifact: false, planeswalker: false,
  });

  // Iterative mode state
  const [iterativeMode, setIterativeMode] = useState(false);
  const [seedCardId, setSeedCardId] = useState<number | null>(null);
  const [iterativeSuggestions, setIterativeSuggestions] = useState<CardWithOwnership[]>([]);
  const [deckAnalysis, setDeckAnalysis] = useState<LiveDeckAnalysis | null>(null);
  const [slotsRemaining, setSlotsRemaining] = useState(60);
  const [landSuggestions, setLandSuggestions] = useState<SuggestedLandResponse[]>([]);
  // Map of cardID to card name for display purposes
  const [cardNameMap, setCardNameMap] = useState<Map<number, string>>(new Map());
  // Hover preview for card magnification
  const [hoverPreview, setHoverPreview] = useState<HoverPreview | null>(null);
  // Copy selection modal state
  const [copyModalCard, setCopyModalCard] = useState<CardWithOwnership | null>(null);
  // Track whether details are expanded in copy modal
  const [detailsExpanded, setDetailsExpanded] = useState(false);

  // Ref to track if we've already attempted initial fetch (prevent infinite loops)
  const hasAttemptedFetch = useRef(false);

  // Track which card IDs we've already fetched names for
  const fetchedCardIdsRef = useRef<Set<number>>(new Set());

  // Complete deck generation state (Issue #774)
  const [showArchetypeSelector, setShowArchetypeSelector] = useState(false);
  const [selectedArchetype, setSelectedArchetype] = useState<ArchetypeKey | null>(null);
  const [archetypeProfiles, setArchetypeProfiles] = useState<Record<string, ArchetypeProfile> | null>(null);
  const [generatedDeck, setGeneratedDeck] = useState<GenerateCompleteDeckResponse | null>(null);
  const [generating, setGenerating] = useState(false);

  // Fetch card names for deck cards when modal opens
  useEffect(() => {
    if (!isOpen || deckCards.length === 0) return;

    const fetchCardNames = async () => {
      // Get card IDs that we haven't fetched yet
      const missingCardIds = deckCards
        .map(c => c.CardID)
        .filter(id => !fetchedCardIdsRef.current.has(id));

      if (missingCardIds.length === 0) return;

      // Mark these as being fetched
      missingCardIds.forEach(id => fetchedCardIdsRef.current.add(id));

      try {
        // Fetch card details for each missing card in parallel
        const cardPromises = missingCardIds.map(id => cardsApi.getCardByArenaId(id));
        const cards = await Promise.all(cardPromises);

        // Update the card name map
        setCardNameMap(prev => {
          const newMap = new Map(prev);
          cards.forEach(card => {
            if (card && card.ArenaID) {
              newMap.set(parseInt(card.ArenaID, 10), card.Name);
            }
          });
          return newMap;
        });
      } catch (err) {
        console.error('Failed to fetch card names:', err);
      }
    };

    fetchCardNames();
  }, [isOpen, deckCards]);

  // Define handleClose before useEffects that depend on it
  const handleClose = useCallback(() => {
    // Reset state
    setIterativeMode(false);
    setSeedCardId(null);
    setIterativeSuggestions([]);
    setDeckAnalysis(null);
    setLandSuggestions([]);
    setCardNameMap(new Map());
    setError(null);
    hasAttemptedFetch.current = false; // Reset for next open
    fetchedCardIdsRef.current = new Set(); // Reset fetched card IDs
    // Reset complete deck generation state
    setShowArchetypeSelector(false);
    setSelectedArchetype(null);
    setGeneratedDeck(null);
    setGenerating(false);
    onClose();
  }, [onClose]);

  // Cleanup debounce timer on unmount
  useEffect(() => {
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  // Handle Escape key to close modal
  useEffect(() => {
    if (!isOpen) return;

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') handleClose();
    };

    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, handleClose]);

  // Fetch suggestions when in iterative mode and deck changes
  const fetchIterativeSuggestions = useCallback(async () => {
    if (!iterativeMode) return;
    // Either need a seed card OR using deck cards mode
    if (!seedCardId && !useDeckCardsAsSeed) return;
    if (currentDeckCards.length === 0) return;

    setLoading(true);
    try {
      const response: IterativeBuildAroundResponse = await decks.suggestNextCards({
        seed_card_id: seedCardId || undefined, // Optional - API analyzes all deck cards
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
  }, [seedCardId, currentDeckCards, budgetMode, iterativeMode, useDeckCardsAsSeed]);

  // Debounced fetch when deck changes in iterative mode
  useEffect(() => {
    if (!iterativeMode) return;
    // Need either seed card or deck cards mode
    if (!seedCardId && !useDeckCardsAsSeed) return;

    const timer = setTimeout(fetchIterativeSuggestions, 300);
    return () => clearTimeout(timer);
  }, [currentDeckCards, iterativeMode, seedCardId, useDeckCardsAsSeed, fetchIterativeSuggestions]);

  // Auto-start iterative mode when useDeckCardsAsSeed is true and modal opens
  useEffect(() => {
    if (!isOpen || !useDeckCardsAsSeed || currentDeckCards.length === 0) return;
    if (iterativeMode) return; // Already in iterative mode
    if (hasAttemptedFetch.current) return; // Already attempted, don't retry on error

    // Mark that we've attempted the fetch
    hasAttemptedFetch.current = true;

    // No seed card needed - the API analyzes ALL deck cards collectively
    setIterativeMode(true);
    setLoading(true);
    setError(null);

    // Fetch suggestions based on collective analysis of all deck cards
    const fetchSuggestions = async () => {
      try {
        const response = await decks.suggestNextCards({
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
        setIterativeMode(false);
      } finally {
        setLoading(false);
      }
    };

    fetchSuggestions();
  }, [isOpen, useDeckCardsAsSeed, currentDeckCards, budgetMode, iterativeMode]);

  // Filtered search results - derived from raw results and current filter state
  const searchResults = useMemo(() => {
    const activeColors = Object.entries(colorFilter).filter(([, v]) => v).map(([k]) => k);
    const activeTypes = Object.entries(typeFilter).filter(([, v]) => v).map(([k]) => k);

    return rawSearchResults.filter(card => {
      // Color filter: if any color selected, card must have at least one of them
      if (activeColors.length > 0) {
        const cardColors = card.colors || [];
        const hasMatchingColor = activeColors.some(c => cardColors.includes(c));
        if (!hasMatchingColor) return false;
      }

      // Type filter: if any type selected, card must have at least one of them
      if (activeTypes.length > 0) {
        const typeLine = (card.types || []).join(' ').toLowerCase();
        const hasMatchingType = activeTypes.some(t => typeLine.includes(t));
        if (!hasMatchingType) return false;
      }

      return true;
    });
  }, [rawSearchResults, colorFilter, typeFilter]);

  const handleSearch = useCallback((query: string) => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    if (query.length < 2) {
      setRawSearchResults([]);
      return;
    }

    debounceRef.current = setTimeout(async () => {
      setSearching(true);
      try {
        // Use searchCardsWithCollection for better results (includes all cards, not just exact matches)
        const results = await cardsApi.searchCardsWithCollection(query, undefined, 50);
        const mapped = results
          .filter(card => card.ArenaID && !isNaN(parseInt(card.ArenaID, 10)))
          .map(card => ({
            arenaID: parseInt(card.ArenaID, 10),
            name: card.Name,
            manaCost: card.ManaCost,
            types: card.Types,
            imageURI: card.ImageURL,
            colors: card.Colors,
            cmc: card.CMC,
          }));
        setRawSearchResults(mapped);
      } catch {
        setRawSearchResults([]);
      } finally {
        setSearching(false);
      }
    }, 300);
  }, []);

  const handleSelectCard = (card: SearchResult) => {
    setSelectedCard(card);
    setSearchQuery(card.name);
    setRawSearchResults([]);
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

  // Handle clicking a card - open copy selection modal
  const handlePickCard = (card: CardWithOwnership) => {
    setCopyModalCard(card);
    setHoverPreview(null); // Hide hover preview when modal opens
  };

  // Handle adding copies from the modal
  const handleAddCopies = (card: CardWithOwnership, count: number) => {
    if (onCardAdded) {
      // Store the card name for display
      setCardNameMap(prev => {
        const newMap = new Map(prev);
        newMap.set(card.cardID, card.name);
        return newMap;
      });
      // Add the specified number of copies
      for (let i = 0; i < count; i++) {
        onCardAdded(card);
      }
    }
    setCopyModalCard(null); // Close modal
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

  const handleClear = () => {
    setSearchQuery('');
    setRawSearchResults([]);
    setSelectedCard(null);
    setSuggestions(null);
    setError(null);
    setIterativeMode(false);
    setSeedCardId(null);
    setIterativeSuggestions([]);
    setDeckAnalysis(null);
    setCardNameMap(new Map());
    // Reset complete deck generation state
    setShowArchetypeSelector(false);
    setSelectedArchetype(null);
    setGeneratedDeck(null);
  };

  // Show archetype selector for Quick Generate (Issue #774)
  const handleShowArchetypeSelector = async () => {
    setShowArchetypeSelector(true);
    setError(null);
    // Fetch archetype profiles if not already loaded
    if (!archetypeProfiles) {
      try {
        const profiles = await decks.getArchetypeProfiles();
        setArchetypeProfiles(profiles);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load archetypes');
      }
    }
  };

  // Generate complete deck with selected archetype
  const handleGenerateCompleteDeck = async (archetype: ArchetypeKey) => {
    if (!selectedCard) return;

    setSelectedArchetype(archetype);
    setGenerating(true);
    setError(null);

    try {
      const response = await decks.generateCompleteDeck({
        seed_card_id: selectedCard.arenaID,
        archetype,
        budget_mode: budgetMode,
      });
      setGeneratedDeck(response);
      setShowArchetypeSelector(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate deck');
    } finally {
      setGenerating(false);
    }
  };

  // Apply generated deck to the current deck
  const handleApplyGeneratedDeck = () => {
    if (!generatedDeck) return;

    // Convert spells and lands to the format expected by onApplyDeck
    const spellSuggestions: CardWithOwnership[] = generatedDeck.spells.map(spell => ({
      cardID: spell.cardID,
      name: spell.name,
      manaCost: spell.manaCost,
      cmc: spell.cmc,
      colors: spell.colors,
      typeLine: spell.typeLine,
      rarity: spell.rarity,
      imageURI: spell.imageURI,
      score: spell.score,
      reasoning: spell.reasoning,
      inCollection: spell.inCollection,
      ownedCount: spell.ownedCount,
      neededCount: spell.neededCount,
      currentCopies: 0,
      recommendedCopies: spell.quantity,
    }));

    const landSuggestions: SuggestedLandResponse[] = generatedDeck.lands.map(land => ({
      cardID: land.cardID,
      name: land.name,
      quantity: land.quantity,
      color: land.colors.join(''),
    }));

    onApplyDeck(spellSuggestions, landSuggestions);
    handleClose();
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

  // Render recommended copies badge
  const renderCopyRecommendation = (card: CardWithOwnership) => {
    // Default to 4 copies if backend doesn't provide recommendation
    const recommended = card.recommendedCopies > 0 ? card.recommendedCopies : 4;
    const current = card.currentCopies || 0;
    const remaining = recommended - current;
    if (remaining <= 0) return null;
    if (current > 0) {
      return <span className="copy-badge has-copies">+{remaining} more (have {current})</span>;
    }
    return <span className="copy-badge">{recommended}x recommended</span>;
  };

  // Handle card hover for preview - position intelligently to stay on screen
  const handleCardHover = (card: CardWithOwnership, e: React.MouseEvent) => {
    const previewHeight = 500; // Approximate height of preview card
    const previewWidth = 280;

    // Calculate position - show above cursor if near bottom of screen
    let y = e.clientY - 100;
    let x = e.clientX + 20;

    // If preview would go off bottom, show above the cursor instead
    if (y + previewHeight > window.innerHeight) {
      y = e.clientY - previewHeight - 20;
    }
    // If still off top, just position at top
    if (y < 10) {
      y = 10;
    }

    // Keep within horizontal bounds
    if (x + previewWidth > window.innerWidth) {
      x = e.clientX - previewWidth - 20;
    }
    if (x < 10) {
      x = 10;
    }

    setHoverPreview({
      card,
      position: { x, y },
    });
  };

  const handleCardHoverEnd = () => {
    setHoverPreview(null);
  };

  // Iterative mode UI - show when in iterative mode (with selected card OR using deck cards as seed)
  if (iterativeMode && (selectedCard || useDeckCardsAsSeed)) {
    const modalTitle = selectedCard
      ? `Building: ${selectedCard.name}`
      : `Build Around Your Deck (${currentDeckCards.length} cards)`;

    return (
      <div className="build-around-overlay" onClick={handleClose}>
        <div
          className="build-around-modal iterative-mode"
          role="dialog"
          aria-modal="true"
          aria-labelledby="iterative-modal-title"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="build-around-header">
            <h2 id="iterative-modal-title">{modalTitle}</h2>
            <button className="close-button" onClick={handleClose} aria-label="Close dialog">
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
                      className={`clickable-suggestion-card ${card.currentCopies > 0 ? 'in-deck' : ''}`}
                      onClick={() => handlePickCard(card)}
                      onMouseEnter={(e) => handleCardHover(card, e)}
                      onMouseLeave={handleCardHoverEnd}
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
                        {renderCopyRecommendation(card)}
                        {renderOwnershipBadge(card)}
                      </div>
                      {card.currentCopies > 0 && (
                        <div className="in-deck-badge">{card.currentCopies} in deck</div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Hover Preview */}
            {hoverPreview && (
              <div
                className="card-hover-preview"
                style={{
                  position: 'fixed',
                  top: hoverPreview.position.y,
                  left: hoverPreview.position.x,
                  zIndex: 10000,
                }}
              >
                <div className="preview-card">
                  {hoverPreview.card.imageURI && (
                    <img
                      src={hoverPreview.card.imageURI}
                      alt={hoverPreview.card.name}
                      className="preview-image"
                    />
                  )}
                  <div className="preview-details">
                    <h3 className="preview-name">{hoverPreview.card.name}</h3>
                    <p className="preview-type">{hoverPreview.card.typeLine}</p>
                    <div className="preview-stats">
                      {hoverPreview.card.manaCost && (
                        <span className="preview-mana">Mana: {hoverPreview.card.manaCost}</span>
                      )}
                      <span className="preview-score">Score: {(hoverPreview.card.score * 100).toFixed(0)}%</span>
                    </div>
                    {/* Score Breakdown */}
                    {hoverPreview.card.scoreBreakdown && (
                      <div className="score-breakdown">
                        <div className="score-bar-row">
                          <span className="score-label">Color</span>
                          <div className="score-bar-container">
                            <div
                              className="score-bar color-bar"
                              style={{ width: `${hoverPreview.card.scoreBreakdown.colorFit * 100}%` }}
                            />
                          </div>
                          <span className="score-value">{(hoverPreview.card.scoreBreakdown.colorFit * 100).toFixed(0)}%</span>
                        </div>
                        <div className="score-bar-row">
                          <span className="score-label">Curve</span>
                          <div className="score-bar-container">
                            <div
                              className="score-bar curve-bar"
                              style={{ width: `${hoverPreview.card.scoreBreakdown.curveFit * 100}%` }}
                            />
                          </div>
                          <span className="score-value">{(hoverPreview.card.scoreBreakdown.curveFit * 100).toFixed(0)}%</span>
                        </div>
                        <div className="score-bar-row">
                          <span className="score-label">Synergy</span>
                          <div className="score-bar-container">
                            <div
                              className="score-bar synergy-bar"
                              style={{ width: `${hoverPreview.card.scoreBreakdown.synergy * 100}%` }}
                            />
                          </div>
                          <span className="score-value">{(hoverPreview.card.scoreBreakdown.synergy * 100).toFixed(0)}%</span>
                        </div>
                        <div className="score-bar-row">
                          <span className="score-label">Quality</span>
                          <div className="score-bar-container">
                            <div
                              className="score-bar quality-bar"
                              style={{ width: `${hoverPreview.card.scoreBreakdown.quality * 100}%` }}
                            />
                          </div>
                          <span className="score-value">{(hoverPreview.card.scoreBreakdown.quality * 100).toFixed(0)}%</span>
                        </div>
                      </div>
                    )}
                    {/* Synergy Details */}
                    {hoverPreview.card.synergyDetails && hoverPreview.card.synergyDetails.length > 0 && (
                      <div className="synergy-details">
                        <span className="synergy-label">Synergies:</span>
                        <ul className="synergy-list">
                          {hoverPreview.card.synergyDetails.slice(0, 3).map((detail, i) => (
                            <li key={i} className={`synergy-item synergy-${detail.type}`}>
                              {detail.name}
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}
                    <p className="preview-reasoning">{hoverPreview.card.reasoning}</p>
                    <div className="preview-recommendation">
                      <strong>Recommended: {hoverPreview.card.recommendedCopies > 0 ? hoverPreview.card.recommendedCopies : 4} copies</strong>
                      {(hoverPreview.card.currentCopies || 0) > 0 && (
                        <span> (have {hoverPreview.card.currentCopies})</span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Copy Selection Modal */}
            {copyModalCard && (
              <div className="copy-modal-overlay" onClick={() => { setCopyModalCard(null); setDetailsExpanded(false); }}>
                <div className="copy-modal" onClick={(e) => e.stopPropagation()}>
                  <div className="copy-modal-header">
                    <h3>Add {copyModalCard.name}</h3>
                    <button className="close-button" onClick={() => { setCopyModalCard(null); setDetailsExpanded(false); }}>&times;</button>
                  </div>
                  <div className="copy-modal-content">
                    {copyModalCard.imageURI && (
                      <img src={copyModalCard.imageURI} alt={copyModalCard.name} className="copy-modal-image" />
                    )}
                    <div className="copy-modal-info">
                      <p className="copy-modal-type">{copyModalCard.typeLine}</p>
                      <p className="copy-modal-reasoning">{copyModalCard.reasoning}</p>

                      {/* Expandable Details Toggle */}
                      {(copyModalCard.scoreBreakdown || (copyModalCard.synergyDetails && copyModalCard.synergyDetails.length > 0)) && (
                        <button
                          className="details-toggle"
                          onClick={() => setDetailsExpanded(!detailsExpanded)}
                        >
                          {detailsExpanded ? '▼ Hide Details' : '▶ Show Details'}
                        </button>
                      )}

                      {/* Expandable Details Section */}
                      {detailsExpanded && (
                        <div className="copy-modal-details">
                          {/* Score Breakdown */}
                          {copyModalCard.scoreBreakdown && (
                            <div className="modal-score-breakdown">
                              <h4>Score Breakdown</h4>
                              <div className="modal-score-grid">
                                <div className="modal-score-item">
                                  <span className="modal-score-name">Color Fit</span>
                                  <span className="modal-score-percent">{(copyModalCard.scoreBreakdown.colorFit * 100).toFixed(0)}%</span>
                                </div>
                                <div className="modal-score-item">
                                  <span className="modal-score-name">Curve Fit</span>
                                  <span className="modal-score-percent">{(copyModalCard.scoreBreakdown.curveFit * 100).toFixed(0)}%</span>
                                </div>
                                <div className="modal-score-item">
                                  <span className="modal-score-name">Synergy</span>
                                  <span className="modal-score-percent">{(copyModalCard.scoreBreakdown.synergy * 100).toFixed(0)}%</span>
                                </div>
                                <div className="modal-score-item">
                                  <span className="modal-score-name">Quality</span>
                                  <span className="modal-score-percent">{(copyModalCard.scoreBreakdown.quality * 100).toFixed(0)}%</span>
                                </div>
                              </div>
                            </div>
                          )}

                          {/* Synergy Details */}
                          {copyModalCard.synergyDetails && copyModalCard.synergyDetails.length > 0 && (
                            <div className="modal-synergy-details">
                              <h4>Synergies Found</h4>
                              <ul className="modal-synergy-list">
                                {copyModalCard.synergyDetails.map((detail, i) => (
                                  <li key={i} className={`modal-synergy-item synergy-type-${detail.type}`}>
                                    <span className="synergy-name">{detail.name}</span>
                                    <span className="synergy-desc">{detail.description}</span>
                                  </li>
                                ))}
                              </ul>
                            </div>
                          )}
                        </div>
                      )}

                      <div className="copy-modal-stats">
                        <span>In deck: {copyModalCard.currentCopies || 0}</span>
                        <span>Recommended: {copyModalCard.recommendedCopies > 0 ? copyModalCard.recommendedCopies : 4}</span>
                      </div>
                    </div>
                  </div>
                  <div className="copy-modal-actions">
                    <span className="copy-modal-label">Add copies:</span>
                    {[1, 2, 3, 4].map(count => {
                      const current = copyModalCard.currentCopies || 0;
                      const maxCanAdd = 4 - current;
                      const disabled = count > maxCanAdd;
                      return (
                        <button
                          key={count}
                          className={`copy-count-btn ${count === 1 ? 'primary' : ''}`}
                          onClick={() => handleAddCopies(copyModalCard, count)}
                          disabled={disabled}
                          title={disabled ? `Already have ${current} copies` : `Add ${count} ${count === 1 ? 'copy' : 'copies'}`}
                        >
                          +{count}
                        </button>
                      );
                    })}
                  </div>
                  <button className="copy-modal-cancel" onClick={() => { setCopyModalCard(null); setDetailsExpanded(false); }}>
                    Cancel
                  </button>
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
                        <span className="card-name">{cardNameMap.get(card.CardID) || `Card #${card.CardID}`}</span>
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
                {/* Mana Curve - shows count of cards at each mana cost */}
                {Object.keys(deckAnalysis.currentCurve).length > 0 && (
                  <div className="mana-curve-simple">
                    <span className="curve-label">Mana Curve:</span>
                    <div className="curve-items">
                      {Object.entries(deckAnalysis.currentCurve)
                        .sort(([a], [b]) => parseInt(a) - parseInt(b))
                        .map(([cmc, count]) => (
                          <span key={cmc} className="curve-item">
                            {cmc}-drop: {count}
                          </span>
                        ))}
                    </div>
                  </div>
                )}
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
                title={slotsRemaining > 30 ? `Add at least ${slotsRemaining - 30} more cards before finishing` : 'Complete deck with lands'}
              >
                Finish Deck (Add Lands)
              </button>
              {slotsRemaining > 30 && (
                <p className="helper-text">Add at least {slotsRemaining - 30} more cards to finish</p>
              )}
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

  // Generated deck result view (Issue #774)
  if (generatedDeck && selectedCard) {
    return (
      <div className="build-around-overlay" onClick={handleClose}>
        <div
          className="build-around-modal generated-deck-mode"
          role="dialog"
          aria-modal="true"
          aria-labelledby="generated-deck-title"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="build-around-header">
            <h2 id="generated-deck-title">
              Generated {selectedArchetype?.charAt(0).toUpperCase()}{selectedArchetype?.slice(1)} Deck
            </h2>
            <button className="close-button" onClick={handleClose} aria-label="Close dialog">
              &times;
            </button>
          </div>

          <div className="build-around-content">
            {/* Strategy Panel */}
            <div className="strategy-panel">
              <div className="strategy-header">
                <h3>Deck Strategy</h3>
                {renderColorPips(generatedDeck.analysis.colorDistribution
                  ? Object.keys(generatedDeck.analysis.colorDistribution)
                  : [])}
              </div>
              <p className="strategy-summary">{generatedDeck.strategy.summary}</p>
              <div className="strategy-sections">
                <div className="strategy-section">
                  <h4>Game Plan</h4>
                  <p>{generatedDeck.strategy.gamePlan}</p>
                </div>
                <div className="strategy-section">
                  <h4>Mulligan Guide</h4>
                  <p>{generatedDeck.strategy.mulligan}</p>
                </div>
                {generatedDeck.strategy.keyCards.length > 0 && (
                  <div className="strategy-section">
                    <h4>Key Cards</h4>
                    <div className="key-cards-list">
                      {generatedDeck.strategy.keyCards.map((card, i) => (
                        <span key={i} className="key-card-tag">{card}</span>
                      ))}
                    </div>
                  </div>
                )}
                {generatedDeck.strategy.strengths.length > 0 && (
                  <div className="strategy-section">
                    <h4>Strengths</h4>
                    <ul className="pros-cons-list">
                      {generatedDeck.strategy.strengths.map((s, i) => (
                        <li key={i} className="strength-item">{s}</li>
                      ))}
                    </ul>
                  </div>
                )}
                {generatedDeck.strategy.weaknesses.length > 0 && (
                  <div className="strategy-section">
                    <h4>Weaknesses</h4>
                    <ul className="pros-cons-list">
                      {generatedDeck.strategy.weaknesses.map((w, i) => (
                        <li key={i} className="weakness-item">{w}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </div>

            {/* Deck Analysis Stats */}
            <div className="generated-deck-stats">
              <div className="stat-row">
                <div className="stat">
                  <span className="stat-value">{generatedDeck.analysis.totalCards}</span>
                  <span className="stat-label">Total Cards</span>
                </div>
                <div className="stat">
                  <span className="stat-value">{generatedDeck.analysis.spellCount}</span>
                  <span className="stat-label">Spells</span>
                </div>
                <div className="stat">
                  <span className="stat-value">{generatedDeck.analysis.landCount}</span>
                  <span className="stat-label">Lands</span>
                </div>
                <div className="stat">
                  <span className="stat-value">{generatedDeck.analysis.averageCMC.toFixed(2)}</span>
                  <span className="stat-label">Avg CMC</span>
                </div>
              </div>

              {/* Mana Curve Visualization */}
              <div className="curve-visualization">
                <h4>Mana Curve</h4>
                <div className="curve-bars">
                  {Object.entries(generatedDeck.analysis.manaCurve)
                    .sort(([a], [b]) => parseInt(a) - parseInt(b))
                    .map(([cmc, count]) => (
                      <div key={cmc} className="curve-bar-wrapper">
                        <div
                          className="curve-bar"
                          style={{ height: `${Math.min(count * 8, 100)}px` }}
                        >
                          <span className="curve-count">{count}</span>
                        </div>
                        <span className="curve-cmc">{cmc}</span>
                      </div>
                    ))}
                </div>
              </div>

              {/* Wildcard Cost */}
              {generatedDeck.analysis.missingCount > 0 && (
                <div className="wildcard-cost">
                  <span className="cost-label">Wildcards needed:</span>
                  {Object.entries(generatedDeck.analysis.missingWildcardCost || {}).map(([rarity, count]) => (
                    count > 0 && (
                      <span key={rarity} className={`wildcard-badge ${rarity}`}>
                        {count} {rarity}
                      </span>
                    )
                  ))}
                </div>
              )}
            </div>

            {/* Deck Lists */}
            <div className="generated-deck-lists">
              {/* Creatures */}
              <div className="deck-list-section">
                <h4>Creatures ({generatedDeck.spells.filter(s => s.typeLine.toLowerCase().includes('creature')).reduce((sum, s) => sum + s.quantity, 0)})</h4>
                <div className="deck-card-list">
                  {generatedDeck.spells
                    .filter(s => s.typeLine.toLowerCase().includes('creature'))
                    .sort((a, b) => a.cmc - b.cmc)
                    .map(spell => (
                      <div key={spell.cardID} className="deck-list-card">
                        <span className="card-quantity">{spell.quantity}x</span>
                        <span className="card-name">{spell.name}</span>
                        <span className="card-mana">{spell.manaCost}</span>
                      </div>
                    ))}
                </div>
              </div>

              {/* Non-Creature Spells */}
              <div className="deck-list-section">
                <h4>Spells ({generatedDeck.spells.filter(s => !s.typeLine.toLowerCase().includes('creature')).reduce((sum, s) => sum + s.quantity, 0)})</h4>
                <div className="deck-card-list">
                  {generatedDeck.spells
                    .filter(s => !s.typeLine.toLowerCase().includes('creature'))
                    .sort((a, b) => a.cmc - b.cmc)
                    .map(spell => (
                      <div key={spell.cardID} className="deck-list-card">
                        <span className="card-quantity">{spell.quantity}x</span>
                        <span className="card-name">{spell.name}</span>
                        <span className="card-mana">{spell.manaCost}</span>
                      </div>
                    ))}
                </div>
              </div>

              {/* Lands */}
              <div className="deck-list-section">
                <h4>Lands ({generatedDeck.lands.reduce((sum, l) => sum + l.quantity, 0)})</h4>
                <div className="deck-card-list">
                  {generatedDeck.lands.map(land => (
                    <div key={land.cardID} className="deck-list-card land-card">
                      <span className="card-quantity">{land.quantity}x</span>
                      <span className="card-name">{land.name}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="suggestions-actions">
              <button
                className="action-btn apply-btn"
                onClick={handleApplyGeneratedDeck}
              >
                Apply Deck
              </button>
              <button
                className="action-btn secondary-btn"
                onClick={() => {
                  setGeneratedDeck(null);
                  setShowArchetypeSelector(true);
                }}
              >
                Try Different Archetype
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

  // Archetype selector view (Issue #774)
  if (showArchetypeSelector && selectedCard) {
    return (
      <div className="build-around-overlay" onClick={handleClose}>
        <div
          className="build-around-modal archetype-selector-mode"
          role="dialog"
          aria-modal="true"
          aria-labelledby="archetype-selector-title"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="build-around-header">
            <h2 id="archetype-selector-title">Choose Deck Archetype</h2>
            <button className="close-button" onClick={handleClose} aria-label="Close dialog">
              &times;
            </button>
          </div>

          <div className="build-around-content">
            {/* Selected Seed Card Preview */}
            <div className="seed-card-preview">
              <h3>Building around: {selectedCard.name}</h3>
              {selectedCard.imageURI && (
                <img src={selectedCard.imageURI} alt={selectedCard.name} className="seed-preview-image" />
              )}
            </div>

            {/* Error State */}
            {error && (
              <div className="build-around-error">
                <p>{error}</p>
              </div>
            )}

            {/* Loading */}
            {generating && (
              <div className="loading-indicator">
                Generating your {selectedArchetype} deck...
              </div>
            )}

            {/* Archetype Options */}
            {!generating && (
              <div className="archetype-options">
                <p className="archetype-instructions">Select a deck style to generate a complete 60-card deck:</p>

                <div className="archetype-buttons">
                  {/* Aggro */}
                  <button
                    className="archetype-btn aggro"
                    onClick={() => handleGenerateCompleteDeck('aggro')}
                  >
                    <div className="archetype-icon">⚡</div>
                    <div className="archetype-info">
                      <h4>Aggro</h4>
                      <p>{archetypeProfiles?.aggro?.description || 'Fast, aggressive deck that aims to win quickly with cheap threats.'}</p>
                      <div className="archetype-stats">
                        <span>20 lands</span>
                        <span>70% creatures</span>
                      </div>
                    </div>
                  </button>

                  {/* Midrange */}
                  <button
                    className="archetype-btn midrange"
                    onClick={() => handleGenerateCompleteDeck('midrange')}
                  >
                    <div className="archetype-icon">⚖️</div>
                    <div className="archetype-info">
                      <h4>Midrange</h4>
                      <p>{archetypeProfiles?.midrange?.description || 'Balanced deck that can play offense or defense as needed.'}</p>
                      <div className="archetype-stats">
                        <span>24 lands</span>
                        <span>55% creatures</span>
                      </div>
                    </div>
                  </button>

                  {/* Control */}
                  <button
                    className="archetype-btn control"
                    onClick={() => handleGenerateCompleteDeck('control')}
                  >
                    <div className="archetype-icon">🛡️</div>
                    <div className="archetype-info">
                      <h4>Control</h4>
                      <p>{archetypeProfiles?.control?.description || 'Reactive deck that answers threats and wins with powerful finishers.'}</p>
                      <div className="archetype-stats">
                        <span>26 lands</span>
                        <span>25% creatures</span>
                      </div>
                    </div>
                  </button>
                </div>

                {/* Budget Mode Toggle */}
                <div className="archetype-options-footer">
                  <label className="option-checkbox">
                    <input
                      type="checkbox"
                      checked={budgetMode}
                      onChange={(e) => setBudgetMode(e.target.checked)}
                    />
                    <span>Budget Mode (only cards in collection)</span>
                  </label>
                </div>
              </div>
            )}

            {/* Back Button */}
            <div className="suggestions-actions">
              <button
                className="action-btn cancel-btn"
                onClick={() => {
                  setShowArchetypeSelector(false);
                  setSelectedArchetype(null);
                }}
                disabled={generating}
              >
                Back to Card Selection
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
      <div
        className="build-around-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="seed-modal-title"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="build-around-header">
          <h2 id="seed-modal-title">Build Around Card</h2>
          <button className="close-button" onClick={handleClose} aria-label="Close dialog">
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

            {/* Filters - clicking toggles filter, useMemo automatically re-filters results */}
            <div className="seed-search-filters">
              <div className="filter-row">
                <span className="filter-label">Colors:</span>
                <div className="color-filter-buttons">
                  {(['W', 'U', 'B', 'R', 'G'] as const).map((color) => (
                    <button
                      key={color}
                      className={`color-filter-btn mana-${color.toLowerCase()} ${colorFilter[color] ? 'active' : ''}`}
                      onClick={() => setColorFilter(prev => ({ ...prev, [color]: !prev[color] }))}
                      title={color === 'W' ? 'White' : color === 'U' ? 'Blue' : color === 'B' ? 'Black' : color === 'R' ? 'Red' : 'Green'}
                    >
                      {color}
                    </button>
                  ))}
                </div>
              </div>
              <div className="filter-row">
                <span className="filter-label">Types:</span>
                <div className="type-filter-buttons">
                  {(['creature', 'instant', 'sorcery', 'enchantment', 'artifact', 'planeswalker'] as const).map((type) => (
                    <button
                      key={type}
                      className={`type-filter-btn ${typeFilter[type] ? 'active' : ''}`}
                      onClick={() => setTypeFilter(prev => ({ ...prev, [type]: !prev[type] }))}
                    >
                      {type.charAt(0).toUpperCase() + type.slice(1, 4)}
                    </button>
                  ))}
                </div>
              </div>
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
                {/* Quick Generate - Complete 60-card deck (Issue #774) */}
                <button
                  className="build-button primary"
                  onClick={handleShowArchetypeSelector}
                  disabled={loading}
                >
                  Quick Generate (60-Card Deck)
                </button>
                {onCardAdded && onFinishDeck && (
                  <button
                    className="build-button secondary"
                    onClick={handleStartBuilding}
                    disabled={loading}
                  >
                    {loading ? 'Starting...' : 'Start Building (Pick Cards)'}
                  </button>
                )}
                <button
                  className="build-button tertiary"
                  onClick={handleBuildAround}
                  disabled={loading}
                >
                  {loading ? 'Generating...' : 'Quick Build (40 Cards)'}
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

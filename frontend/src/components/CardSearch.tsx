import { useState, useEffect, useMemo, useCallback } from 'react';
import { GetCardByArenaID, GetAllSetInfo, SearchCardsWithCollection } from '../../wailsjs/go/main/App';
import { models, gui } from '../../wailsjs/go/models';
import SetSymbol from './SetSymbol';
import './CardSearch.css';

interface CardSearchProps {
  isDraftDeck: boolean;
  draftCardIDs?: number[]; // Available cards from draft pool
  existingCards: Map<number, { quantity: number; board: string }>; // Cards already in deck
  onAddCard: (cardID: number, quantity: number, board: 'main' | 'sideboard') => Promise<void>;
  onRemoveCard: (cardID: number, board: 'main' | 'sideboard') => Promise<void>;
}

interface ColorFilter {
  white: boolean;
  blue: boolean;
  black: boolean;
  red: boolean;
  green: boolean;
  colorless: boolean;
  multicolor: boolean;
}

interface TypeFilter {
  creature: boolean;
  instant: boolean;
  sorcery: boolean;
  enchantment: boolean;
  artifact: boolean;
  planeswalker: boolean;
  land: boolean;
}

// Card with ownership information from the API
interface CardWithOwned extends models.SetCard {
  ownedQuantity?: number;
}

export default function CardSearch({
  isDraftDeck,
  draftCardIDs = [],
  existingCards,
  onAddCard,
  onRemoveCard,
}: CardSearchProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');
  const [allCards, setAllCards] = useState<CardWithOwned[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedBoard, setSelectedBoard] = useState<'main' | 'sideboard'>('main');
  const [cmcMin, setCmcMin] = useState<number | ''>('');
  const [cmcMax, setCmcMax] = useState<number | ''>('');
  const [sets, setSets] = useState<gui.SetInfo[]>([]);
  const [selectedSets, setSelectedSets] = useState<string[]>([]);
  const [showSetFilter, setShowSetFilter] = useState(false);
  const [collectionOnly, setCollectionOnly] = useState(false);
  const [colorFilter, setColorFilter] = useState<ColorFilter>({
    white: false,
    blue: false,
    black: false,
    red: false,
    green: false,
    colorless: false,
    multicolor: false,
  });
  const [typeFilter, setTypeFilter] = useState<TypeFilter>({
    creature: false,
    instant: false,
    sorcery: false,
    enchantment: false,
    artifact: false,
    planeswalker: false,
    land: false,
  });

  // Debounce search term
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchTerm(searchTerm);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchTerm]);

  // Load available sets for filtering (only for constructed)
  useEffect(() => {
    if (!isDraftDeck) {
      GetAllSetInfo()
        .then((setInfo) => setSets(setInfo || []))
        .catch((err) => console.error('Failed to load sets:', err));
    }
  }, [isDraftDeck]);

  // Load draft pool cards for draft decks
  useEffect(() => {
    if (isDraftDeck) {
      const loadDraftCards = async () => {
        setLoading(true);
        setError(null);
        try {
          if (draftCardIDs.length > 0) {
            const cards: CardWithOwned[] = [];
            for (const cardID of draftCardIDs) {
              try {
                const card = await GetCardByArenaID(String(cardID));
                if (card) {
                  cards.push(card as CardWithOwned);
                }
              } catch (err) {
                console.error(`Failed to load card ${cardID}:`, err);
              }
            }
            setAllCards(cards);
          } else {
            setAllCards([]);
          }
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to load cards');
        } finally {
          setLoading(false);
        }
      };
      loadDraftCards();
    }
  }, [isDraftDeck, draftCardIDs]);

  // Search cards for constructed decks
  const searchConstructedCards = useCallback(async () => {
    if (isDraftDeck || debouncedSearchTerm.length < 2) {
      if (!isDraftDeck) {
        setAllCards([]);
      }
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const results = await SearchCardsWithCollection(debouncedSearchTerm, selectedSets, 100, collectionOnly);
      // Map results to CardWithOwned interface
      const cardsWithOwned: CardWithOwned[] = (results || []).map((r: gui.CardWithOwned) => ({
        ...r,
        ownedQuantity: r.ownedQuantity,
      }));
      setAllCards(cardsWithOwned);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to search cards');
      setAllCards([]);
    } finally {
      setLoading(false);
    }
  }, [isDraftDeck, debouncedSearchTerm, selectedSets, collectionOnly]);

  // Trigger search when debounced term, set filter, or collection filter changes
  useEffect(() => {
    searchConstructedCards();
  }, [searchConstructedCards]);

  // Filter cards based on local filters (for draft, filter the draft pool; for constructed, filter search results)
  const filteredCards = useMemo(() => {
    return allCards.filter((card) => {
      // For draft decks, also filter by search term locally
      if (isDraftDeck && searchTerm && !card.Name.toLowerCase().includes(searchTerm.toLowerCase())) {
        return false;
      }

      // CMC filter
      if (cmcMin !== '' && card.CMC < cmcMin) {
        return false;
      }
      if (cmcMax !== '' && card.CMC > cmcMax) {
        return false;
      }

      // Color filter
      const anyColorSelected = Object.values(colorFilter).some((v) => v);
      if (anyColorSelected) {
        const colors = card.Colors || [];
        const isColorless = colors.length === 0;
        const isMulticolor = colors.length > 1;

        if (colorFilter.colorless && !isColorless) return false;
        if (colorFilter.multicolor && !isMulticolor) return false;

        // Check individual colors
        if (!colorFilter.colorless && !colorFilter.multicolor) {
          const colorMatch =
            (colorFilter.white && colors.includes('W')) ||
            (colorFilter.blue && colors.includes('U')) ||
            (colorFilter.black && colors.includes('B')) ||
            (colorFilter.red && colors.includes('R')) ||
            (colorFilter.green && colors.includes('G'));
          if (!colorMatch && !isColorless && !isMulticolor) return false;
        }
      }

      // Type filter
      const anyTypeSelected = Object.values(typeFilter).some((v) => v);
      if (anyTypeSelected) {
        const typeLine = (card.Types || []).join(' ').toLowerCase();
        const typeMatch =
          (typeFilter.creature && typeLine.includes('creature')) ||
          (typeFilter.instant && typeLine.includes('instant')) ||
          (typeFilter.sorcery && typeLine.includes('sorcery')) ||
          (typeFilter.enchantment && typeLine.includes('enchantment')) ||
          (typeFilter.artifact && typeLine.includes('artifact')) ||
          (typeFilter.planeswalker && typeLine.includes('planeswalker')) ||
          (typeFilter.land && typeLine.includes('land'));
        if (!typeMatch) return false;
      }

      return true;
    });
  }, [allCards, searchTerm, isDraftDeck, cmcMin, cmcMax, colorFilter, typeFilter]);

  const handleAddCard = async (card: CardWithOwned, quantity: number = 1) => {
    try {
      const arenaID = parseInt(card.ArenaID);
      await onAddCard(arenaID, quantity, selectedBoard);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to add card');
    }
  };

  const handleRemoveCard = async (card: CardWithOwned) => {
    try {
      const arenaID = parseInt(card.ArenaID);
      const existing = existingCards.get(arenaID);
      if (existing) {
        await onRemoveCard(arenaID, existing.board as 'main' | 'sideboard');
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to remove card');
    }
  };

  const getCardInDeck = (cardID: number) => {
    return existingCards.get(cardID);
  };

  const getAvailableQuantity = (cardID: number): number => {
    if (!isDraftDeck) return 99; // No limit for constructed

    // For draft: count how many of this card are in the draft pool
    return draftCardIDs.filter((id) => id === cardID).length;
  };

  const toggleSetFilter = (setCode: string) => {
    setSelectedSets((prev) =>
      prev.includes(setCode) ? prev.filter((s) => s !== setCode) : [...prev, setCode]
    );
  };

  const clearSetFilter = () => {
    setSelectedSets([]);
  };

  return (
    <div className="card-search">
      <div className="card-search-header">
        <h3>Card Search</h3>
        {isDraftDeck && (
          <div className="draft-mode-indicator">
            <span className="draft-badge">Draft Mode</span>
            <span className="draft-pool-count">{draftCardIDs.length} cards in pool</span>
          </div>
        )}
      </div>

      {/* Search Input */}
      <div className="search-input-container">
        <input
          type="text"
          className="search-input"
          placeholder={isDraftDeck ? 'Filter draft pool...' : 'Search cards (min 2 characters)...'}
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          autoComplete="off"
        />
      </div>

      {/* Collection Filter Toggle for Constructed */}
      {!isDraftDeck && (
        <div className="collection-filter-group">
          <span className="filter-label">Show:</span>
          <div className="collection-toggle">
            <button
              className={`toggle-option ${!collectionOnly ? 'active' : ''}`}
              onClick={() => setCollectionOnly(false)}
            >
              All Cards
            </button>
            <button
              className={`toggle-option ${collectionOnly ? 'active' : ''}`}
              onClick={() => setCollectionOnly(true)}
            >
              My Collection
            </button>
          </div>
        </div>
      )}

      {/* Set Filter for Constructed */}
      {!isDraftDeck && (
        <div className="filter-group set-filter-group">
          <button
            className={`set-filter-toggle ${selectedSets.length > 0 ? 'has-filters' : ''}`}
            onClick={() => setShowSetFilter(!showSetFilter)}
          >
            Sets: {selectedSets.length > 0 ? `${selectedSets.length} selected` : 'All'}
            <span className="toggle-icon">{showSetFilter ? '▲' : '▼'}</span>
          </button>
          {selectedSets.length > 0 && (
            <button className="clear-sets-button" onClick={clearSetFilter}>
              Clear
            </button>
          )}
          {showSetFilter && (
            <div className="set-filter-dropdown">
              {sets.map((set) => (
                <label key={set.code} className="set-filter-option">
                  <input
                    type="checkbox"
                    checked={selectedSets.includes(set.code)}
                    onChange={() => toggleSetFilter(set.code)}
                  />
                  <span className="set-name">
                    {set.name} ({set.code.toUpperCase()})
                  </span>
                </label>
              ))}
              {sets.length === 0 && <div className="no-sets">No sets cached yet</div>}
            </div>
          )}
        </div>
      )}

      {/* Filters */}
      <div className="search-filters">
        {/* CMC Filter */}
        <div className="filter-group">
          <label>CMC Range:</label>
          <input
            type="number"
            min="0"
            max="20"
            placeholder="Min"
            value={cmcMin}
            onChange={(e) => setCmcMin(e.target.value ? Number(e.target.value) : '')}
            className="cmc-input"
          />
          <span>to</span>
          <input
            type="number"
            min="0"
            max="20"
            placeholder="Max"
            value={cmcMax}
            onChange={(e) => setCmcMax(e.target.value ? Number(e.target.value) : '')}
            className="cmc-input"
          />
        </div>

        {/* Color Filter */}
        <div className="filter-group">
          <label>Colors:</label>
          <div className="color-filters">
            {(['white', 'blue', 'black', 'red', 'green', 'colorless', 'multicolor'] as const).map((color) => (
              <button
                key={color}
                className={`color-button ${color} ${colorFilter[color] ? 'active' : ''}`}
                onClick={() => setColorFilter({ ...colorFilter, [color]: !colorFilter[color] })}
                title={color.charAt(0).toUpperCase() + color.slice(1)}
              >
                {color === 'white' && 'W'}
                {color === 'blue' && 'U'}
                {color === 'black' && 'B'}
                {color === 'red' && 'R'}
                {color === 'green' && 'G'}
                {color === 'colorless' && 'C'}
                {color === 'multicolor' && 'M'}
              </button>
            ))}
          </div>
        </div>

        {/* Type Filter */}
        <div className="filter-group">
          <label>Types:</label>
          <div className="type-filters">
            {(['creature', 'instant', 'sorcery', 'enchantment', 'artifact', 'planeswalker', 'land'] as const).map(
              (type) => (
                <button
                  key={type}
                  className={`type-button ${typeFilter[type] ? 'active' : ''}`}
                  onClick={() => setTypeFilter({ ...typeFilter, [type]: !typeFilter[type] })}
                >
                  {type.charAt(0).toUpperCase() + type.slice(1)}
                </button>
              )
            )}
          </div>
        </div>

        {/* Board Selection */}
        <div className="filter-group">
          <label>Add to:</label>
          <div className="board-selection">
            <button
              className={`board-button ${selectedBoard === 'main' ? 'active' : ''}`}
              onClick={() => setSelectedBoard('main')}
            >
              Maindeck
            </button>
            <button
              className={`board-button ${selectedBoard === 'sideboard' ? 'active' : ''}`}
              onClick={() => setSelectedBoard('sideboard')}
            >
              Sideboard
            </button>
          </div>
        </div>
      </div>

      {/* Results */}
      <div className="search-results">
        {loading && <div className="loading">Searching...</div>}
        {error && <div className="error">{error}</div>}

        {!loading && !error && filteredCards.length === 0 && (
          <div className="no-results">
            {isDraftDeck
              ? searchTerm
                ? 'No cards match your search in draft pool'
                : 'No cards available in draft pool'
              : searchTerm.length < 2
                ? 'Type at least 2 characters to search'
                : collectionOnly
                  ? 'No cards in your collection match this search'
                  : 'No cards found. Try a different search term.'}
          </div>
        )}

        {!loading && !error && filteredCards.length > 0 && (
          <div className="card-list">
            <div className="result-count">{filteredCards.length} cards found</div>
            {filteredCards.map((card) => {
              const arenaID = parseInt(card.ArenaID);
              const inDeck = getCardInDeck(arenaID);
              const available = getAvailableQuantity(arenaID);
              const inDeckQuantity = inDeck?.quantity || 0;
              const ownedQuantity = card.ownedQuantity || 0;

              return (
                <div key={`${card.ArenaID}-${card.SetCode}`} className={`card-result ${inDeck ? 'in-deck' : ''}`}>
                  {card.ImageURL && <img src={card.ImageURL} alt={card.Name} className="card-image" />}
                  <div className="card-info">
                    <div className="card-name">{card.Name}</div>
                    <div className="card-type">{(card.Types || []).join(' — ')}</div>
                    {card.ManaCost && <div className="card-mana-cost">{card.ManaCost}</div>}
                    <div className="card-stats">
                      <span>CMC: {card.CMC}</span>
                      {card.SetCode && (
                        <span className="card-set">
                          <SetSymbol
                            setCode={card.SetCode}
                            size="small"
                            rarity={card.Rarity?.toLowerCase() as 'common' | 'uncommon' | 'rare' | 'mythic' | undefined}
                          />
                        </span>
                      )}
                      {isDraftDeck && (
                        <span className="available-quantity">
                          Available: {available - inDeckQuantity} / {available}
                        </span>
                      )}
                      {!isDraftDeck && ownedQuantity > 0 && (
                        <span className="owned-quantity">{ownedQuantity}x owned</span>
                      )}
                      {!isDraftDeck && ownedQuantity === 0 && (
                        <span className="not-owned">Not owned</span>
                      )}
                    </div>
                  </div>
                  <div className="card-actions">
                    {inDeck && (
                      <div className="in-deck-info">
                        <span className="in-deck-badge">
                          {inDeck.quantity}x in {inDeck.board}
                        </span>
                      </div>
                    )}
                    {(!isDraftDeck || inDeckQuantity < available) && (
                      <button
                        className="add-button"
                        onClick={() => handleAddCard(card, 1)}
                        title={`Add to ${selectedBoard}`}
                      >
                        + Add
                      </button>
                    )}
                    {inDeck && (
                      <button
                        className="remove-button"
                        onClick={() => handleRemoveCard(card)}
                        title="Remove from deck"
                      >
                        - Remove
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

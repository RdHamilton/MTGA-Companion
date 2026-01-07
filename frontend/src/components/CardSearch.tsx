import { useState, useEffect, useMemo, useCallback } from 'react';
import { cards as cardsApi } from '@/services/api';
import { gui } from '@/types/models';
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

// Card data with ownership information
interface CardWithOwned {
  ID?: number;
  SetCode: string;
  ArenaID: string;
  ScryfallID?: string;
  Name: string;
  ManaCost?: string;
  CMC: number;
  Types?: string[];
  Colors?: string[];
  Rarity?: string;
  Text?: string;
  Power?: string;
  Toughness?: string;
  ImageURL?: string;
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
      cardsApi.getAllSetInfo()
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
                const card = await cardsApi.getCardByArenaId(cardID);
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
      const results = await cardsApi.searchCardsWithCollection(debouncedSearchTerm, selectedSets, 100);
      // Results already match CardWithOwned interface
      setAllCards((results || []) as CardWithOwned[]);
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

      {/* Results - using inline styles only (no CSS classes) */}
      <div style={{ background: '#252525', borderRadius: '6px', padding: '1rem', marginTop: '1rem' }}>
        {loading && <div style={{ textAlign: 'center', padding: '2rem', color: '#aaa' }}>Searching...</div>}
        {error && <div style={{ textAlign: 'center', padding: '2rem', color: '#ff6b6b' }}>{error}</div>}

        {!loading && !error && filteredCards.length === 0 && (
          <div style={{ textAlign: 'center', padding: '2rem', color: '#aaa' }}>
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
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div style={{ color: '#aaa', fontSize: '0.875rem', marginBottom: '0.5rem', paddingBottom: '0.5rem', borderBottom: '1px solid #333' }}>
              {filteredCards.length} cards found
            </div>
            {filteredCards.map((card) => {
              const arenaID = parseInt(card.ArenaID);
              const inDeck = getCardInDeck(arenaID);
              const available = getAvailableQuantity(arenaID);
              const inDeckQuantity = inDeck?.quantity || 0;
              const ownedQuantity = card.ownedQuantity || 0;

              return (
                <div
                  key={`${card.ArenaID}-${card.SetCode}`}
                  style={{
                    display: 'flex',
                    gap: '1rem',
                    padding: '1rem',
                    background: inDeck ? 'rgba(74, 158, 255, 0.1)' : '#2a2a2a',
                    borderRadius: '6px',
                    border: inDeck ? '2px solid #4a9eff' : '2px solid transparent',
                    position: 'relative'
                  }}
                >
                  {card.ImageURL && (
                    <img
                      src={card.ImageURL}
                      alt={card.Name}
                      style={{ width: '120px', height: 'auto', borderRadius: '6px', objectFit: 'cover', flexShrink: 0 }}
                    />
                  )}
                  <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <div style={{ fontSize: '1.125rem', fontWeight: 600, color: '#fff' }}>{card.Name}</div>
                    <div style={{ fontSize: '0.875rem', color: '#aaa' }}>{(card.Types || []).join(' — ')}</div>
                    {card.ManaCost && <div style={{ fontSize: '0.875rem', color: '#4a9eff', fontFamily: 'Courier New, monospace' }}>{card.ManaCost}</div>}
                    <div style={{ display: 'flex', gap: '1rem', fontSize: '0.875rem', color: '#aaa', marginTop: 'auto' }}>
                      <span>CMC: {card.CMC}</span>
                      {card.SetCode && (
                        <span>
                          <SetSymbol
                            setCode={card.SetCode}
                            size="small"
                            rarity={card.Rarity?.toLowerCase() as 'common' | 'uncommon' | 'rare' | 'mythic' | undefined}
                          />
                        </span>
                      )}
                      {isDraftDeck && (
                        <span style={{ color: '#4a9eff', fontWeight: 600 }}>
                          Available: {available - inDeckQuantity} / {available}
                        </span>
                      )}
                      {!isDraftDeck && ownedQuantity > 0 && (
                        <span style={{ color: '#4caf50', fontWeight: 600 }}>{ownedQuantity}x owned</span>
                      )}
                      {!isDraftDeck && ownedQuantity === 0 && (
                        <span style={{ color: '#888', fontStyle: 'italic' }}>Not owned</span>
                      )}
                    </div>
                  </div>
                  <div style={{ position: 'absolute', bottom: '1rem', right: '1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem', alignItems: 'flex-end' }}>
                    {inDeck && (
                      <span style={{ background: 'rgba(74, 158, 255, 0.2)', color: '#4a9eff', padding: '0.25rem 0.75rem', borderRadius: '12px', fontSize: '0.75rem', fontWeight: 600, border: '1px solid #4a9eff' }}>
                        {inDeck.quantity}x in {inDeck.board}
                      </span>
                    )}
                    {(!isDraftDeck || inDeckQuantity < available) && (
                      <button
                        onClick={() => handleAddCard(card, 1)}
                        title={`Add to ${selectedBoard}`}
                        style={{ padding: '0.5rem 1rem', border: 'none', borderRadius: '4px', fontSize: '0.875rem', fontWeight: 600, cursor: 'pointer', minWidth: '100px', background: '#4caf50', color: '#fff' }}
                      >
                        + Add
                      </button>
                    )}
                    {inDeck && (
                      <button
                        onClick={() => handleRemoveCard(card)}
                        title="Remove from deck"
                        style={{ padding: '0.5rem 1rem', border: 'none', borderRadius: '4px', fontSize: '0.875rem', fontWeight: 600, cursor: 'pointer', minWidth: '100px', background: '#f44336', color: '#fff' }}
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

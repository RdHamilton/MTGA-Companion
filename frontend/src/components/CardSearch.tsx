import { useState, useEffect, useMemo } from 'react';
import { GetCardByArenaID } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
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

export default function CardSearch({
  isDraftDeck,
  draftCardIDs = [],
  existingCards,
  onAddCard,
  onRemoveCard,
}: CardSearchProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [allCards, setAllCards] = useState<models.SetCard[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedBoard, setSelectedBoard] = useState<'main' | 'sideboard'>('main');
  const [cmcMin, setCmcMin] = useState<number | ''>('');
  const [cmcMax, setCmcMax] = useState<number | ''>('');
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

  // Load all available cards (or draft pool for draft decks)
  useEffect(() => {
    const loadCards = async () => {
      setLoading(true);
      setError(null);
      try {
        if (isDraftDeck) {
          // For draft decks: load only draft pool cards
          if (draftCardIDs.length > 0) {
            const cards: models.SetCard[] = [];
            for (const cardID of draftCardIDs) {
              try {
                const card = await GetCardByArenaID(String(cardID));
                if (card) {
                  cards.push(card);
                }
              } catch (err) {
                console.error(`Failed to load card ${cardID}:`, err);
              }
            }
            setAllCards(cards);
          } else {
            // Empty draft pool - let the empty state message show
            setAllCards([]);
          }
        } else {
          // For constructed decks: would need to load all cards
          // For now, we'll show a message that constructed search needs set selection
          setError('Please select a set to search for cards');
          setAllCards([]);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load cards');
      } finally {
        setLoading(false);
      }
    };

    loadCards();
  }, [isDraftDeck, draftCardIDs]);

  // Filter cards based on search term and filters
  const filteredCards = useMemo(() => {
    return allCards.filter((card) => {
      // Search term filter (case-insensitive)
      if (searchTerm && !card.Name.toLowerCase().includes(searchTerm.toLowerCase())) {
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

        // Check individual colors (for mono-color cards)
        if (!isColorless && !isMulticolor) {
          const colorMatch =
            (colorFilter.white && colors.includes('W')) ||
            (colorFilter.blue && colors.includes('U')) ||
            (colorFilter.black && colors.includes('B')) ||
            (colorFilter.red && colors.includes('R')) ||
            (colorFilter.green && colors.includes('G'));
          if (!colorMatch) return false;
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
  }, [allCards, searchTerm, cmcMin, cmcMax, colorFilter, typeFilter]);

  const handleAddCard = async (card: models.SetCard, quantity: number = 1) => {
    try {
      const arenaID = parseInt(card.ArenaID);
      await onAddCard(arenaID, quantity, selectedBoard);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to add card');
    }
  };

  const handleRemoveCard = async (card: models.SetCard) => {
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
          placeholder="Search card name..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          autoComplete="off"
        />
      </div>

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
        {loading && <div className="loading">Loading cards...</div>}
        {error && <div className="error">{error}</div>}

        {!loading && !error && filteredCards.length === 0 && (
          <div className="no-results">
            {searchTerm || Object.values(colorFilter).some((v) => v) || Object.values(typeFilter).some((v) => v)
              ? 'No cards match your search criteria'
              : isDraftDeck
              ? 'No cards available in draft pool'
              : 'Start typing to search for cards'}
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

              return (
                <div key={card.ArenaID} className={`card-result ${inDeck ? 'in-deck' : ''}`}>
                  {card.ImageURL && (
                    <img src={card.ImageURL} alt={card.Name} className="card-image" />
                  )}
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
                    </div>
                  </div>
                  <div className="card-actions">
                    {inDeck && (
                      <div className="in-deck-info">
                        <span className="in-deck-badge">{inDeck.quantity}x in {inDeck.board}</span>
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

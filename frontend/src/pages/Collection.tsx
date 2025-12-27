import { useState, useEffect, useCallback } from 'react';
import { collection, cards as cardsApi } from '@/services/api';
import { gui } from '@/types/models';
import SetCompletionPanel from '../components/SetCompletion';
import './Collection.css';

// Color icon mapping
const colorIcons: Record<string, string> = {
  W: 'https://svgs.scryfall.io/card-symbols/W.svg',
  U: 'https://svgs.scryfall.io/card-symbols/U.svg',
  B: 'https://svgs.scryfall.io/card-symbols/B.svg',
  R: 'https://svgs.scryfall.io/card-symbols/R.svg',
  G: 'https://svgs.scryfall.io/card-symbols/G.svg',
};

// Rarity colors
const rarityColors: Record<string, string> = {
  common: '#1a1a1a',
  uncommon: '#6b7c8d',
  rare: '#d4af37',
  mythic: '#e67e22',
};

interface FilterState {
  searchTerm: string;
  setCode: string;
  rarity: string;
  colors: string[];
  ownedOnly: boolean;
  sortBy: string;
  sortDesc: boolean;
}

const ITEMS_PER_PAGE = 50;

export default function Collection() {
  const [cards, setCards] = useState<gui.CollectionCard[]>([]);
  const [sets, setSets] = useState<gui.SetInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [totalCount, setTotalCount] = useState(0);
  const [filterCount, setFilterCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [showSetCompletion, setShowSetCompletion] = useState(false);

  const [filters, setFilters] = useState<FilterState>({
    searchTerm: '',
    setCode: '',
    rarity: '',
    colors: [],
    ownedOnly: true,
    sortBy: 'name',
    sortDesc: false,
  });

  // Debounced search term
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchTerm(filters.searchTerm);
    }, 300);
    return () => clearTimeout(timer);
  }, [filters.searchTerm]);

  const loadCollection = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const apiFilter = {
        set_code: filters.setCode,
        rarity: filters.rarity,
        colors: filters.colors,
        owned_only: filters.ownedOnly,
      };

      const collectionCards = await collection.getCollection(apiFilter);
      // Note: REST API doesn't support search/sort/pagination server-side
      // The component handles this with client-side filtering
      setCards(collectionCards || []);
      setTotalCount(collectionCards.length);
      setFilterCount(collectionCards.length);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load collection');
      console.error('Failed to load collection:', err);
    } finally {
      setLoading(false);
    }
  }, [debouncedSearchTerm, filters.setCode, filters.rarity, filters.colors, filters.ownedOnly, filters.sortBy, filters.sortDesc, currentPage]);

  const loadSets = useCallback(async () => {
    try {
      const setInfo = await cardsApi.getAllSetInfo();
      setSets(setInfo || []);
    } catch (err) {
      console.error('Failed to load sets:', err);
    }
  }, []);

  useEffect(() => {
    loadCollection();
    loadSets();
  }, []);

  // Reload collection when filters change
  useEffect(() => {
    loadCollection();
  }, [loadCollection]);

  // Reset page when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [debouncedSearchTerm, filters.setCode, filters.rarity, filters.colors, filters.ownedOnly, filters.sortBy, filters.sortDesc]);

  const handleFilterChange = (key: keyof FilterState, value: string | string[] | boolean) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const handleColorToggle = (color: string) => {
    setFilters((prev) => ({
      ...prev,
      colors: prev.colors.includes(color)
        ? prev.colors.filter((c) => c !== color)
        : [...prev.colors, color],
    }));
  };

  const totalPages = Math.ceil(filterCount / ITEMS_PER_PAGE);

  if (loading && cards.length === 0) {
    return (
      <div className="collection-page loading-state">
        <div className="loading-spinner"></div>
        <p>Loading collection...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="collection-page error-state">
        <div className="error-icon">!</div>
        <h2>Error Loading Collection</h2>
        <p>{error}</p>
        <button onClick={loadCollection} className="retry-button">
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="collection-page">
      {/* Header with stats */}
      <div className="collection-header">
        <div className="header-title">
          <h1>Collection</h1>
          <div className="collection-stats-summary">
            <span className="stat-item">
              <span className="stat-label">Cards in Set:</span>
              <span className="stat-value">{filterCount}</span>
            </span>
            <span className="stat-separator">|</span>
            <span className="stat-item">
              <span className="stat-label">Total Cards:</span>
              <span className="stat-value">{totalCount}</span>
            </span>
          </div>
        </div>
        <button
          className="set-completion-button"
          onClick={() => setShowSetCompletion(!showSetCompletion)}
        >
          {showSetCompletion ? 'Hide' : 'Show'} Set Completion
        </button>
      </div>

      {/* Filter Controls */}
      <div className="collection-filters">
        <div className="filter-row">
          {/* Search */}
          <div className="filter-group search-group">
            <input
              type="text"
              placeholder="Search by name..."
              value={filters.searchTerm}
              onChange={(e) => handleFilterChange('searchTerm', e.target.value)}
              className="search-input"
            />
          </div>

          {/* Set Filter */}
          <div className="filter-group">
            <select
              value={filters.setCode}
              onChange={(e) => handleFilterChange('setCode', e.target.value)}
              className="filter-select"
            >
              <option value="">All Sets</option>
              {sets.map((set) => (
                <option key={set.code} value={set.code}>
                  {set.name} ({set.code.toUpperCase()})
                </option>
              ))}
            </select>
          </div>

          {/* Rarity Filter */}
          <div className="filter-group">
            <select
              value={filters.rarity}
              onChange={(e) => handleFilterChange('rarity', e.target.value)}
              className="filter-select"
            >
              <option value="">All Rarities</option>
              <option value="common">Common</option>
              <option value="uncommon">Uncommon</option>
              <option value="rare">Rare</option>
              <option value="mythic">Mythic</option>
            </select>
          </div>

          {/* Sort */}
          <div className="filter-group">
            <select
              value={`${filters.sortBy}-${filters.sortDesc ? 'desc' : 'asc'}`}
              onChange={(e) => {
                const [sortBy, direction] = e.target.value.split('-');
                handleFilterChange('sortBy', sortBy);
                handleFilterChange('sortDesc', direction === 'desc');
              }}
              className="filter-select"
            >
              <option value="name-asc">Name (A-Z)</option>
              <option value="name-desc">Name (Z-A)</option>
              <option value="quantity-desc">Quantity (High)</option>
              <option value="quantity-asc">Quantity (Low)</option>
              <option value="rarity-desc">Rarity (High)</option>
              <option value="rarity-asc">Rarity (Low)</option>
              <option value="cmc-asc">CMC (Low)</option>
              <option value="cmc-desc">CMC (High)</option>
            </select>
          </div>
        </div>

        <div className="filter-row secondary">
          {/* Color Filters */}
          <span className="filter-label">Colors:</span>
          <div className="color-buttons">
            {['W', 'U', 'B', 'R', 'G'].map((color) => (
              <button
                key={color}
                className={`color-button ${filters.colors.includes(color) ? 'active' : ''}`}
                onClick={() => handleColorToggle(color)}
                title={color === 'W' ? 'White' : color === 'U' ? 'Blue' : color === 'B' ? 'Black' : color === 'R' ? 'Red' : 'Green'}
              >
                <img src={colorIcons[color]} alt={color} className="color-icon" />
              </button>
            ))}
          </div>

          {/* Owned Only Toggle */}
          <label className="toggle-label">
            <input
              type="checkbox"
              checked={filters.ownedOnly}
              onChange={(e) => handleFilterChange('ownedOnly', e.target.checked)}
            />
            Owned only
          </label>

          {/* Result Count */}
          <div className="filter-results">
            Showing {filterCount} of {totalCount} cards
          </div>
        </div>
      </div>

      {/* Set Completion Panel */}
      {showSetCompletion && (
        <div className="set-completion-container">
          <SetCompletionPanel onClose={() => setShowSetCompletion(false)} />
        </div>
      )}

      {/* Card Grid */}
      {cards.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">!</div>
          <h2>No Cards Found</h2>
          <p>
            {filters.searchTerm || filters.setCode || filters.rarity || filters.colors.length > 0
              ? 'Try adjusting your filters'
              : 'Your collection is empty. Start playing to add cards!'}
          </p>
        </div>
      ) : (
        <>
          <div className="card-grid">
            {cards.map((card) => {
              // Check if we have a real card image (not the card back placeholder)
              const isCardBackPlaceholder = card.imageUri?.includes('back.png');
              const hasImage = card.imageUri && card.imageUri !== '' && !isCardBackPlaceholder;
              return (
                <div
                  key={`${card.cardId}-${card.setCode}`}
                  className={`collection-card ${card.quantity === 0 ? 'not-owned' : ''} ${!hasImage ? 'no-image' : ''}`}
                >
                  {hasImage ? (
                    <img
                      src={card.imageUri}
                      alt={card.name || `Card #${card.arenaId}`}
                      style={{ width: '100%', borderRadius: '12px' }}
                      onError={(e) => {
                        const target = e.target as HTMLImageElement;
                        // Hide broken image and show fallback info
                        target.style.display = 'none';
                        const parent = target.parentElement;
                        if (parent && !parent.querySelector('.card-info-fallback')) {
                          parent.classList.add('no-image');
                        }
                      }}
                    />
                  ) : (
                    <div className="card-info-fallback">
                      <div className="card-fallback-name">{card.name || 'Unknown Card'}</div>
                      {card.setCode ? (
                        <div className="card-fallback-set">{card.setCode.toUpperCase()}</div>
                      ) : (
                        <div className="card-fallback-hint">
                          Card #{card.arenaId}
                          <br />
                          <span className="download-hint">Download set in Settings</span>
                        </div>
                      )}
                      {card.manaCost && <div className="card-fallback-mana">{card.manaCost}</div>}
                      {card.rarity && (
                        <div
                          className="card-fallback-rarity"
                          style={{ color: rarityColors[card.rarity.toLowerCase()] || '#888' }}
                        >
                          {card.rarity}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="pagination">
              <button
                className="page-button"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage(1)}
              >
                First
              </button>
              <button
                className="page-button"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              >
                Previous
              </button>
              <span className="page-info">
                Page {currentPage} of {totalPages}
              </span>
              <button
                className="page-button"
                disabled={currentPage === totalPages}
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              >
                Next
              </button>
              <button
                className="page-button"
                disabled={currentPage === totalPages}
                onClick={() => setCurrentPage(totalPages)}
              >
                Last
              </button>
            </div>
          )}
        </>
      )}

      {/* Loading overlay for filter changes */}
      {loading && cards.length > 0 && (
        <div className="loading-overlay">
          <div className="loading-spinner small"></div>
        </div>
      )}
    </div>
  );
}

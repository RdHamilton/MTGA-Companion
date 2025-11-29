import React, { useState, useEffect } from 'react';
import { GetCardRatings, RefreshSetRatings, GetSetCards } from '../../wailsjs/go/main/App';
import { gui, models } from '../../wailsjs/go/models';
import './TierList.css';

type CardRating = gui.CardRatingWithTier;

interface TierListProps {
    setCode: string;
    draftFormat: string;
    pickedCardIds: Set<string>;
    onCardClick?: (arenaId: number) => void;
}

type SortColumn = 'name' | 'gihwr' | 'alsa' | 'rarity';
type SortDirection = 'asc' | 'desc';

const TierList: React.FC<TierListProps> = ({ setCode, draftFormat, pickedCardIds, onCardClick }) => {
    const [ratings, setRatings] = useState<CardRating[]>([]);
    const [setCards, setSetCards] = useState<models.SetCard[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [refreshing, setRefreshing] = useState(false);

    // Filters
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedColors, setSelectedColors] = useState<Set<string>>(new Set());
    const [selectedRarities, setSelectedRarities] = useState<Set<string>>(new Set());
    const [selectedTiers, setSelectedTiers] = useState<Set<string>>(new Set(['S', 'A', 'B', 'C', 'D', 'F']));
    const [selectedTypes, setSelectedTypes] = useState<Set<string>>(new Set());
    const [showPickedOnly, setShowPickedOnly] = useState(false);

    // Sorting
    const [sortColumn, setSortColumn] = useState<SortColumn>('gihwr');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

    useEffect(() => {
        const loadRatings = async () => {
            try {
                setLoading(true);
                setError(null);
                const [ratingsData, cardsData] = await Promise.all([
                    GetCardRatings(setCode, draftFormat),
                    GetSetCards(setCode)
                ]);
                setRatings(ratingsData || []);
                setSetCards(cardsData || []);
            } catch (err) {
                console.error('Failed to load card ratings:', err);
                setError(err instanceof Error ? err.message : 'Failed to load card ratings');
            } finally {
                setLoading(false);
            }
        };

        loadRatings();
    }, [setCode, draftFormat]);

    const handleRefresh = async () => {
        const loadRatings = async () => {
            try {
                setLoading(true);
                setError(null);
                const [ratingsData, cardsData] = await Promise.all([
                    GetCardRatings(setCode, draftFormat),
                    GetSetCards(setCode)
                ]);
                setRatings(ratingsData || []);
                setSetCards(cardsData || []);
            } catch (err) {
                console.error('Failed to load card ratings:', err);
                setError(err instanceof Error ? err.message : 'Failed to load card ratings');
            } finally {
                setLoading(false);
            }
        };
        try {
            setRefreshing(true);
            setError(null);
            console.log(`Refreshing 17Lands data for ${setCode} / ${draftFormat}...`);
            await RefreshSetRatings(setCode, draftFormat);
            console.log('Refresh complete, reloading ratings...');
            await loadRatings();
        } catch (err) {
            console.error('Failed to refresh card ratings:', err);
            setError(err instanceof Error ? err.message : 'Failed to refresh card ratings');
        } finally {
            setRefreshing(false);
        }
    };

    const toggleFilter = (filterSet: Set<string>, setFilterSet: (set: Set<string>) => void, value: string) => {
        const newSet = new Set(filterSet);
        if (newSet.has(value)) {
            newSet.delete(value);
        } else {
            newSet.add(value);
        }
        setFilterSet(newSet);
    };

    const handleSort = (column: SortColumn) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('desc');
        }
    };

    const getTierColor = (tier: string): string => {
        switch (tier) {
            case 'S': return '#ffd700'; // Gold
            case 'A': return '#c0c0c0'; // Silver
            case 'B': return '#cd7f32'; // Bronze
            case 'C': return '#4a9eff'; // Blue
            case 'D': return '#888888'; // Gray
            case 'F': return '#ff4444'; // Red
            default: return '#aaaaaa';
        }
    };

    const getColorSymbol = (color: string): string => {
        switch (color) {
            case 'W': return '⚪'; // White
            case 'U': return '🔵'; // Blue
            case 'B': return '⚫'; // Black
            case 'R': return '🔴'; // Red
            case 'G': return '🟢'; // Green
            default: return '⚪'; // Colorless
        }
    };


    // Filter and sort ratings
    const filteredRatings = ratings
        .filter(rating => {
            // Search filter - filter by card name
            if (searchTerm && !rating.name.toLowerCase().includes(searchTerm.toLowerCase())) {
                return false;
            }

            // Picked cards filter
            if (showPickedOnly) {
                const isPicked = rating.mtga_id ? pickedCardIds.has(String(rating.mtga_id)) : false;
                if (!isPicked) return false;
            }

            // Tier filter
            if (!selectedTiers.has(rating.tier)) return false;

            // Color filter - show card if ANY of its colors match ANY selected color
            if (selectedColors.size > 0) {
                const cardColors = rating.colors && rating.colors.length > 0
                    ? rating.colors
                    : [rating.color];
                const hasMatchingColor = cardColors.some((color: string) => selectedColors.has(color));
                if (!hasMatchingColor) return false;
            }

            // Rarity filter
            if (selectedRarities.size > 0 && !selectedRarities.has(rating.rarity)) return false;

            // Type filter - look up card in setCards to get type information
            if (selectedTypes.size > 0) {
                const card = setCards.find(c => c.Name === rating.name);
                if (card && card.Types && card.Types.length > 0) {
                    const hasMatchingType = card.Types.some(type =>
                        selectedTypes.has(type) ||
                        Array.from(selectedTypes).some(selectedType =>
                            type.toLowerCase().includes(selectedType.toLowerCase())
                        )
                    );
                    if (!hasMatchingType) return false;
                } else {
                    // If we don't have type data for this card, exclude it when filtering by type
                    return false;
                }
            }

            return true;
        })
        .sort((a, b) => {
            let comparison = 0;
            switch (sortColumn) {
                case 'name':
                    comparison = a.name.localeCompare(b.name);
                    break;
                case 'gihwr':
                    comparison = a.ever_drawn_win_rate - b.ever_drawn_win_rate;
                    break;
                case 'alsa':
                    comparison = a.avg_seen - b.avg_seen;
                    break;
                case 'rarity': {
                    const rarityOrder: Record<string, number> = { 'mythic': 4, 'rare': 3, 'uncommon': 2, 'common': 1 };
                    comparison = (rarityOrder[a.rarity.toLowerCase()] || 0) - (rarityOrder[b.rarity.toLowerCase()] || 0);
                    break;
                }
            }
            return sortDirection === 'asc' ? comparison : -comparison;
        });

    // Group by tier
    const groupedByTier = filteredRatings.reduce((acc, rating) => {
        if (!acc[rating.tier]) {
            acc[rating.tier] = [];
        }
        acc[rating.tier].push(rating);
        return acc;
    }, {} as Record<string, CardRating[]>);

    if (loading) {
        return (
            <div className="tier-list-loading">
                <div className="loading-spinner"></div>
                <p>Loading card ratings...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="tier-list-error">
                <p>⚠️ {error}</p>
                <p className="error-help">Make sure 17Lands data is available for {setCode}</p>
            </div>
        );
    }

    if (ratings.length === 0) {
        return (
            <div className="tier-list-empty">
                <p>No card ratings available for {setCode} ({draftFormat})</p>
                <p className="empty-help">Card ratings will appear once 17Lands data is fetched</p>
            </div>
        );
    }

    return (
        <div className="tier-list-container">
            <div className="tier-list-header">
                <h2>Card Tier List</h2>
                <div className="tier-list-info">
                    <span>{filteredRatings.length} cards</span>
                    <span>•</span>
                    <span>17Lands data</span>
                    <button
                        className="refresh-button"
                        onClick={handleRefresh}
                        disabled={refreshing || loading}
                        title="Refresh 17Lands data from API"
                    >
                        {refreshing ? '⟳ Refreshing...' : '↻ Refresh'}
                    </button>
                </div>
            </div>

            {/* Filters */}
            <div className="tier-list-filters">
                {/* Search Input */}
                <div className="filter-group search-group">
                    <label>Search:</label>
                    <input
                        type="text"
                        className="search-input"
                        placeholder="Search by card name..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                    {searchTerm && (
                        <button
                            className="clear-search-btn"
                            onClick={() => setSearchTerm('')}
                            title="Clear search"
                        >
                            X
                        </button>
                    )}
                </div>

                {/* Picked Cards Filter */}
                <div className="filter-group">
                    <label className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={showPickedOnly}
                            onChange={(e) => setShowPickedOnly(e.target.checked)}
                            className="filter-checkbox"
                        />
                        <span>Show Picked Cards Only</span>
                        {showPickedOnly && pickedCardIds.size > 0 && (
                            <span className="picked-count"> ({pickedCardIds.size} picked)</span>
                        )}
                    </label>
                </div>

                {/* Tier Filter */}
                <div className="filter-group">
                    <label>Tiers:</label>
                    <div className="filter-buttons">
                        {['S', 'A', 'B', 'C', 'D', 'F'].map(tier => (
                            <button
                                key={tier}
                                className={`filter-btn tier-btn ${selectedTiers.has(tier) ? 'active' : ''}`}
                                style={{
                                    borderColor: getTierColor(tier),
                                    color: selectedTiers.has(tier) ? getTierColor(tier) : '#888888'
                                }}
                                onClick={() => toggleFilter(selectedTiers, setSelectedTiers, tier)}
                            >
                                {tier}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Color Filter */}
                <div className="filter-group">
                    <label>Colors:</label>
                    <div className="filter-buttons">
                        {['W', 'U', 'B', 'R', 'G'].map(color => (
                            <button
                                key={color}
                                className={`filter-btn color-btn ${selectedColors.has(color) ? 'active' : ''}`}
                                onClick={() => toggleFilter(selectedColors, setSelectedColors, color)}
                            >
                                {getColorSymbol(color)} {color}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Rarity Filter */}
                <div className="filter-group">
                    <label>Rarities:</label>
                    <div className="filter-buttons">
                        {['common', 'uncommon', 'rare', 'mythic'].map(rarity => (
                            <button
                                key={rarity}
                                className={`filter-btn rarity-btn ${selectedRarities.has(rarity) ? 'active' : ''}`}
                                onClick={() => toggleFilter(selectedRarities, setSelectedRarities, rarity)}
                            >
                                {rarity.charAt(0).toUpperCase() + rarity.slice(1)}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Type Filter */}
                <div className="filter-group">
                    <label>Types:</label>
                    <div className="filter-buttons">
                        {['Creature', 'Instant', 'Sorcery', 'Enchantment', 'Artifact'].map(type => (
                            <button
                                key={type}
                                className={`filter-btn type-btn ${selectedTypes.has(type) ? 'active' : ''}`}
                                onClick={() => toggleFilter(selectedTypes, setSelectedTypes, type)}
                            >
                                {type}
                            </button>
                        ))}
                    </div>
                </div>
            </div>

            {/* Tier Groups */}
            <div className="tier-groups">
                {['S', 'A', 'B', 'C', 'D', 'F'].map(tier => {
                    const tierCards = groupedByTier[tier] || [];
                    if (tierCards.length === 0 || !selectedTiers.has(tier)) return null;

                    return (
                        <div key={tier} className="tier-group">
                            <div className="tier-group-header" style={{ borderLeftColor: getTierColor(tier) }}>
                                <span className="tier-badge" style={{ backgroundColor: getTierColor(tier) }}>
                                    {tier}
                                </span>
                                <span className="tier-count">{tierCards.length} cards</span>
                            </div>

                            <div className="tier-table">
                                <table>
                                    <thead>
                                        <tr>
                                            <th onClick={() => handleSort('name')} className="sortable">
                                                Card Name {sortColumn === 'name' && (sortDirection === 'asc' ? '▲' : '▼')}
                                            </th>
                                            <th>Color</th>
                                            <th onClick={() => handleSort('rarity')} className="sortable">
                                                Rarity {sortColumn === 'rarity' && (sortDirection === 'asc' ? '▲' : '▼')}
                                            </th>
                                            <th
                                                onClick={() => handleSort('gihwr')}
                                                className="sortable"
                                                title="Games In Hand Win Rate - Win rate when this card is in your hand during the game"
                                            >
                                                GIHWR {sortColumn === 'gihwr' && (sortDirection === 'asc' ? '▲' : '▼')}
                                            </th>
                                            <th
                                                onClick={() => handleSort('alsa')}
                                                className="sortable"
                                                title="Average Last Seen At - Average pick number when this card is last seen in packs (lower = picked earlier = better)"
                                            >
                                                ALSA {sortColumn === 'alsa' && (sortDirection === 'asc' ? '▲' : '▼')}
                                            </th>
                                            <th title="Card tier based on GIHWR (S = Bomb, A = Excellent, B = Good, C = Playable, D = Below Average, F = Avoid)">TIER</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {tierCards.map((rating: CardRating) => {
                                            const isPicked = rating.mtga_id ? pickedCardIds.has(String(rating.mtga_id)) : false;

                                            return (
                                                <tr
                                                    key={rating.mtga_id || rating.name}
                                                    className={isPicked ? 'picked-card' : ''}
                                                    onClick={() => onCardClick && rating.mtga_id && onCardClick(rating.mtga_id)}
                                                >
                                                    <td className="card-name">
                                                        {isPicked && <span className="picked-marker">✓</span>}
                                                        {rating.name}
                                                    </td>
                                                    <td className="card-color">
                                                        {rating.colors && rating.colors.length > 0
                                                            ? rating.colors.map((c: string) => getColorSymbol(c)).join('')
                                                            : getColorSymbol(rating.color)}
                                                    </td>
                                                    <td className="card-rarity">{rating.rarity}</td>
                                                    <td className="card-gihwr">{rating.ever_drawn_win_rate.toFixed(1)}%</td>
                                                    <td className="card-alsa">{rating.avg_seen.toFixed(1)}</td>
                                                    <td className="card-tier">
                                                        <span
                                                            className="tier-badge-inline"
                                                            style={{ color: getTierColor(rating.tier), fontWeight: 700 }}
                                                        >
                                                            {rating.tier}
                                                        </span>
                                                    </td>
                                                </tr>
                                            );
                                        })}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default TierList;

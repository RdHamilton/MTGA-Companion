import React, { useState, useEffect, useMemo } from 'react';
import { models } from '../../wailsjs/go/models';
import { main } from '../../wailsjs/go/models';
import { GetSetCards, GetCardRatings, GetCompletedDraftSessions } from '../../wailsjs/go/main/App';
import './TierList.css';

interface TierListState {
    setCards: models.SetCard[];
    ratings: main.CardRatingWithTier[];
    availableSets: string[];
    loading: boolean;
    error: string | null;
}

interface Filters {
    setCode: string;
    draftFormat: 'PremierDraft' | 'QuickDraft';
    colors: string[];
    rarities: string[];
    types: string[];
    sortBy: 'gihwr' | 'alsa' | 'name';
    viewMode: 'compact' | 'detailed';
}

interface TierGroup {
    tier: string;
    tierName: string;
    tierDescription: string;
    color: string;
    cards: CardWithRating[];
}

interface CardWithRating {
    card: models.SetCard;
    rating: main.CardRatingWithTier;
}

const TierList: React.FC = () => {
    const [state, setState] = useState<TierListState>({
        setCards: [],
        ratings: [],
        availableSets: [],
        loading: true,
        error: null,
    });

    const [filters, setFilters] = useState<Filters>({
        setCode: 'TLA', // Default to Thunderous Liberty set
        draftFormat: 'PremierDraft',
        colors: [],
        rarities: [],
        types: [],
        sortBy: 'gihwr',
        viewMode: 'detailed',
    });

    // Load available sets on mount (from draft sessions)
    useEffect(() => {
        const loadAvailableSets = async () => {
            try {
                const sessions = await GetCompletedDraftSessions(100);
                const uniqueSets = Array.from(new Set(sessions.map(s => s.SetCode))).sort();

                setState(prev => ({ ...prev, availableSets: uniqueSets, loading: false }));

                // If the default set isn't available, use the first available one
                if (uniqueSets.length > 0 && !uniqueSets.includes('TLA')) {
                    setFilters(prev => ({ ...prev, setCode: uniqueSets[0] }));
                }
            } catch (error) {
                console.error('Failed to load available sets:', error);
                // Continue with default set even if we can't load session history
                setState(prev => ({ ...prev, availableSets: ['TLA'], loading: false }));
            }
        };
        loadAvailableSets();
    }, []);

    // Load cards and ratings when set or format changes
    useEffect(() => {
        if (!filters.setCode) return;

        const loadData = async () => {
            setState(prev => ({ ...prev, loading: true, error: null }));
            try {
                const [setCards, ratings] = await Promise.all([
                    GetSetCards(filters.setCode),
                    GetCardRatings(filters.setCode, filters.draftFormat),
                ]);

                setState(prev => ({
                    ...prev,
                    setCards: setCards || [],
                    ratings: ratings || [],
                    loading: false,
                }));
            } catch (error) {
                console.error('Failed to load tier list data:', error);
                setState(prev => ({
                    ...prev,
                    error: 'Failed to load card data',
                    loading: false,
                }));
            }
        };
        loadData();
    }, [filters.setCode, filters.draftFormat]);

    // Calculate tier groups with filtering and sorting
    const tierGroups = useMemo((): TierGroup[] => {
        if (state.setCards.length === 0 || state.ratings.length === 0) {
            return [];
        }

        // Create rating lookup map
        const ratingMap = new Map<number, main.CardRatingWithTier>();
        state.ratings.forEach(r => {
            if (r.mtga_id) {
                ratingMap.set(r.mtga_id, r);
            }
        });

        // Combine cards with ratings
        const cardsWithRatings: CardWithRating[] = state.setCards
            .map(card => {
                const arenaId = parseInt(card.ArenaID);
                const rating = ratingMap.get(arenaId);
                if (!rating) return null;
                return { card, rating };
            })
            .filter((item): item is CardWithRating => item !== null);

        // Apply filters
        let filteredCards = cardsWithRatings;

        // Color filter
        if (filters.colors.length > 0) {
            filteredCards = filteredCards.filter(({ card }) => {
                if (filters.colors.includes('C')) {
                    // Colorless cards have empty Colors array
                    if (card.Colors.length === 0) return true;
                }
                if (filters.colors.includes('M')) {
                    // Multicolor cards have 2+ colors
                    if (card.Colors.length >= 2) return true;
                }
                // Single color cards
                return card.Colors.some(c => filters.colors.includes(c));
            });
        }

        // Rarity filter
        if (filters.rarities.length > 0) {
            filteredCards = filteredCards.filter(({ card }) =>
                filters.rarities.includes(card.Rarity)
            );
        }

        // Type filter
        if (filters.types.length > 0) {
            filteredCards = filteredCards.filter(({ card }) =>
                filters.types.some(type =>
                    card.Types.some(t => t.toLowerCase().includes(type.toLowerCase()))
                )
            );
        }

        // Group by tier
        const tierDefinitions = [
            { tier: 'S', name: 'S Tier (Bombs)', description: 'Format-defining cards', color: '#ffd700', minWR: 60 },
            { tier: 'A', name: 'A Tier', description: 'Excellent cards, high picks', color: '#c0c0c0', minWR: 57 },
            { tier: 'B', name: 'B Tier', description: 'Good playables', color: '#cd7f32', minWR: 54 },
            { tier: 'C', name: 'C Tier', description: 'Filler/role players', color: '#4a9eff', minWR: 51 },
            { tier: 'D', name: 'D Tier', description: 'Below average', color: '#888888', minWR: 48 },
            { tier: 'F', name: 'F Tier', description: 'Avoid/sideboard', color: '#ff4444', minWR: 0 },
        ];

        const groups: TierGroup[] = tierDefinitions.map(def => {
            const tierCards = filteredCards.filter(({ rating }) => {
                const gihwr = rating.ever_drawn_win_rate;
                if (def.tier === 'F') return gihwr < 48;
                if (def.tier === 'D') return gihwr >= 48 && gihwr < 51;
                if (def.tier === 'C') return gihwr >= 51 && gihwr < 54;
                if (def.tier === 'B') return gihwr >= 54 && gihwr < 57;
                if (def.tier === 'A') return gihwr >= 57 && gihwr < 60;
                if (def.tier === 'S') return gihwr >= 60;
                return false;
            });

            // Sort within tier
            tierCards.sort((a, b) => {
                if (filters.sortBy === 'gihwr') {
                    return b.rating.ever_drawn_win_rate - a.rating.ever_drawn_win_rate;
                } else if (filters.sortBy === 'alsa') {
                    return (a.rating.avg_seen || 0) - (b.rating.avg_seen || 0);
                } else {
                    return a.card.Name.localeCompare(b.card.Name);
                }
            });

            return {
                tier: def.tier,
                tierName: def.name,
                tierDescription: def.description,
                color: def.color,
                cards: tierCards,
            };
        });

        // Only return tiers with cards
        return groups.filter(g => g.cards.length > 0);
    }, [state.setCards, state.ratings, filters]);

    const getConfidenceIndicator = (gihCount: number | undefined): { icon: string; label: string; color: string } => {
        const count = gihCount || 0;
        if (count >= 1000) {
            return { icon: '✓', label: 'High confidence', color: '#4caf50' };
        } else if (count >= 100) {
            return { icon: '⚠', label: 'Moderate confidence', color: '#ff9800' };
        } else {
            return { icon: '❌', label: 'Low confidence', color: '#f44336' };
        }
    };

    const toggleFilter = (filterType: keyof Pick<Filters, 'colors' | 'rarities' | 'types'>, value: string) => {
        setFilters(prev => {
            const currentValues = prev[filterType];
            const newValues = currentValues.includes(value)
                ? currentValues.filter(v => v !== value)
                : [...currentValues, value];
            return { ...prev, [filterType]: newValues };
        });
    };

    if (state.loading) {
        return (
            <div className="tier-list-container">
                <div className="tier-list-loading">Loading tier list...</div>
            </div>
        );
    }

    if (state.error) {
        return (
            <div className="tier-list-container">
                <div className="tier-list-error">{state.error}</div>
            </div>
        );
    }

    if (state.availableSets.length === 0 && !state.loading) {
        return (
            <div className="tier-list-container">
                <div className="tier-list-empty">
                    No draft history available. Complete some drafts or manually enter a set code.
                </div>
            </div>
        );
    }

    return (
        <div className="tier-list-container">
            <div className="tier-list-header">
                <h1>Card Tier List</h1>
                <p className="tier-list-subtitle">
                    Cards ranked by 17Lands Games In Hand Win Rate (GIHWR)
                </p>
            </div>

            <div className="tier-list-filters">
                {/* Set Selection */}
                <div className="filter-group">
                    <label>Set:</label>
                    <select
                        value={filters.setCode}
                        onChange={(e) => setFilters(prev => ({ ...prev, setCode: e.target.value }))}
                    >
                        {state.availableSets.length > 0 ? (
                            state.availableSets.map(setCode => (
                                <option key={setCode} value={setCode}>
                                    {setCode}
                                </option>
                            ))
                        ) : (
                            <option value="TLA">TLA (Default)</option>
                        )}
                    </select>
                </div>

                {/* Draft Format */}
                <div className="filter-group">
                    <label>Format:</label>
                    <select
                        value={filters.draftFormat}
                        onChange={(e) => setFilters(prev => ({
                            ...prev,
                            draftFormat: e.target.value as 'PremierDraft' | 'QuickDraft'
                        }))}
                    >
                        <option value="PremierDraft">Premier Draft</option>
                        <option value="QuickDraft">Quick Draft</option>
                    </select>
                </div>

                {/* Color Filters */}
                <div className="filter-group">
                    <label>Colors:</label>
                    <div className="filter-chips">
                        {['W', 'U', 'B', 'R', 'G', 'C', 'M'].map(color => (
                            <button
                                key={color}
                                className={`filter-chip ${filters.colors.includes(color) ? 'active' : ''}`}
                                onClick={() => toggleFilter('colors', color)}
                                title={
                                    color === 'C' ? 'Colorless' :
                                    color === 'M' ? 'Multicolor' :
                                    color
                                }
                            >
                                {color}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Rarity Filters */}
                <div className="filter-group">
                    <label>Rarity:</label>
                    <div className="filter-chips">
                        {['common', 'uncommon', 'rare', 'mythic'].map(rarity => (
                            <button
                                key={rarity}
                                className={`filter-chip ${filters.rarities.includes(rarity) ? 'active' : ''}`}
                                onClick={() => toggleFilter('rarities', rarity)}
                            >
                                {rarity[0].toUpperCase()}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Type Filters */}
                <div className="filter-group">
                    <label>Type:</label>
                    <div className="filter-chips">
                        {['Creature', 'Instant', 'Sorcery', 'Enchantment', 'Artifact'].map(type => (
                            <button
                                key={type}
                                className={`filter-chip ${filters.types.includes(type) ? 'active' : ''}`}
                                onClick={() => toggleFilter('types', type)}
                            >
                                {type}
                            </button>
                        ))}
                    </div>
                </div>

                {/* Sort Options */}
                <div className="filter-group">
                    <label>Sort By:</label>
                    <select
                        value={filters.sortBy}
                        onChange={(e) => setFilters(prev => ({ ...prev, sortBy: e.target.value as any }))}
                    >
                        <option value="gihwr">Win Rate (Highest)</option>
                        <option value="alsa">Pick Order (Earliest)</option>
                        <option value="name">Name (A-Z)</option>
                    </select>
                </div>

                {/* View Mode */}
                <div className="filter-group">
                    <label>View:</label>
                    <select
                        value={filters.viewMode}
                        onChange={(e) => setFilters(prev => ({ ...prev, viewMode: e.target.value as any }))}
                    >
                        <option value="detailed">Detailed</option>
                        <option value="compact">Compact</option>
                    </select>
                </div>
            </div>

            <div className="tier-list-content">
                {tierGroups.length === 0 ? (
                    <div className="tier-list-no-results">
                        No cards match the selected filters. Try adjusting your filters.
                    </div>
                ) : (
                    tierGroups.map(group => (
                        <div key={group.tier} className="tier-group">
                            <div className="tier-group-header">
                                <span
                                    className="tier-badge"
                                    style={{ backgroundColor: group.color }}
                                >
                                    {group.tier}
                                </span>
                                <div className="tier-group-info">
                                    <h2>{group.tierName}</h2>
                                    <p>{group.tierDescription}</p>
                                </div>
                                <span className="tier-card-count">{group.cards.length} cards</span>
                            </div>

                            <div className={`tier-cards ${filters.viewMode}`}>
                                {group.cards.map(({ card, rating }) => {
                                    const gihCount = rating["# ever_drawn"];
                                    const confidence = getConfidenceIndicator(gihCount);
                                    return (
                                        <div
                                            key={card.ArenaID}
                                            className="tier-card"
                                            title={card.Name}
                                        >
                                            <div className="tier-card-image">
                                                {card.ImageURLSmall ? (
                                                    <img src={card.ImageURLSmall} alt={card.Name} />
                                                ) : (
                                                    <div className="tier-card-placeholder">
                                                        {card.Name.substring(0, 2)}
                                                    </div>
                                                )}
                                            </div>

                                            <div className="tier-card-info">
                                                <div className="tier-card-header">
                                                    <span className="tier-card-name">{card.Name}</span>
                                                    <span className="tier-card-rarity">{card.Rarity[0].toUpperCase()}</span>
                                                </div>

                                                <div className="tier-card-stats">
                                                    <div className="stat">
                                                        <span className="stat-label">GIHWR:</span>
                                                        <span className="stat-value gihwr">
                                                            {rating.ever_drawn_win_rate.toFixed(1)}%
                                                        </span>
                                                    </div>
                                                    <div className="stat">
                                                        <span className="stat-label">ALSA:</span>
                                                        <span className="stat-value">
                                                            {rating.avg_seen?.toFixed(1) || 'N/A'}
                                                        </span>
                                                    </div>
                                                    <div className="stat">
                                                        <span className="stat-label">Sample:</span>
                                                        <span
                                                            className="stat-value confidence"
                                                            style={{ color: confidence.color }}
                                                            title={confidence.label}
                                                        >
                                                            {confidence.icon} {gihCount || 0}
                                                        </span>
                                                    </div>
                                                </div>

                                                {filters.viewMode === 'detailed' && (
                                                    <div className="tier-card-meta">
                                                        <span className="meta-item">{card.ManaCost || `{${card.CMC}}`}</span>
                                                        <span className="meta-item">{card.Types.join(' ')}</span>
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default TierList;

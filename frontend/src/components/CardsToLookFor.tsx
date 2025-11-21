import React, { useMemo } from 'react';
import { models, gui } from '../../wailsjs/go/models';
import { analyzeSynergies, calculateCardSynergyScore, getSynergyReason } from '../utils/synergy';
import './CardsToLookFor.css';

interface CardsToLookForProps {
    pickedCards: models.SetCard[];
    availableCards: models.SetCard[];
    ratings: gui.CardRatingWithTier[];
    onCardClick?: (card: models.SetCard) => void;
}

interface SuggestedCard {
    card: models.SetCard;
    synergyScore: number;
    synergyReason: string;
    gihwr: number;
    tier: string;
}

const CardsToLookFor: React.FC<CardsToLookForProps> = ({
    pickedCards,
    availableCards,
    ratings,
    onCardClick
}) => {
    const suggestions = useMemo(() => {
        if (pickedCards.length === 0) {
            return {
                highPriority: [],
                colorSynergy: [],
                curveFiller: [],
                typeSynergy: []
            };
        }

        // Analyze synergies from picked cards
        const synergy = analyzeSynergies(pickedCards);

        // Filter out already picked cards
        const pickedArenaIds = new Set(pickedCards.map(c => c.ArenaID));
        const unpickedCards = availableCards.filter(c => !pickedArenaIds.has(c.ArenaID));

        // Create rating lookup map
        const ratingMap = new Map<number, gui.CardRatingWithTier>();
        ratings.forEach(r => {
            if (r.mtga_id) {
                ratingMap.set(r.mtga_id, r);
            }
        });

        // Score each card
        const scoredCards: SuggestedCard[] = unpickedCards
            .map(card => {
                const synergyScore = calculateCardSynergyScore(card, synergy);
                const synergyReason = getSynergyReason(card, synergy);

                // Get 17Lands rating
                const arenaId = parseInt(card.ArenaID);
                const rating = ratingMap.get(arenaId);
                const gihwr = rating?.ever_drawn_win_rate || 0;
                const tier = rating?.tier || 'F';

                return {
                    card,
                    synergyScore,
                    synergyReason,
                    gihwr,
                    tier
                };
            })
            .filter(s => s.synergyScore >= 20); // Only show cards with meaningful synergy

        // Sort by combined score (synergy + GIHWR)
        scoredCards.sort((a, b) => {
            const scoreA = a.synergyScore + (a.gihwr / 2); // Synergy weighted more than raw power
            const scoreB = b.synergyScore + (b.gihwr / 2);
            return scoreB - scoreA;
        });

        // Categorize suggestions
        const highPriority = scoredCards.filter(s => s.synergyScore >= 60).slice(0, 5);
        const colorSynergy = scoredCards.filter(s =>
            s.synergyReason.includes('colors') && s.synergyScore >= 30
        ).slice(0, 5);
        const curveFiller = scoredCards.filter(s =>
            s.synergyReason.includes('gap') || s.synergyReason.includes('drop')
        ).slice(0, 5);
        const typeSynergy = scoredCards.filter(s =>
            s.synergyReason.includes('synergy') && !s.synergyReason.includes('colors')
        ).slice(0, 5);

        return {
            highPriority,
            colorSynergy,
            curveFiller,
            typeSynergy
        };
    }, [pickedCards, availableCards, ratings]);

    const getTierColor = (tier: string): string => {
        switch (tier) {
            case 'S': return '#ffd700';
            case 'A': return '#c0c0c0';
            case 'B': return '#cd7f32';
            case 'C': return '#4a9eff';
            case 'D': return '#888888';
            case 'F': return '#ff4444';
            default: return '#aaaaaa';
        }
    };

    const renderSuggestionCard = (suggestion: SuggestedCard) => (
        <div
            key={suggestion.card.ArenaID}
            className="suggestion-card"
            onClick={() => onCardClick && onCardClick(suggestion.card)}
            title={suggestion.card.Name}
        >
            <div className="suggestion-card-image">
                {suggestion.card.ImageURLSmall ? (
                    <img src={suggestion.card.ImageURLSmall} alt={suggestion.card.Name} />
                ) : (
                    <div className="suggestion-card-placeholder">
                        {suggestion.card.Name.substring(0, 2)}
                    </div>
                )}
            </div>
            <div className="suggestion-card-info">
                <div className="suggestion-card-header">
                    <span className="suggestion-card-name">{suggestion.card.Name}</span>
                    <span
                        className="suggestion-card-tier"
                        style={{ color: getTierColor(suggestion.tier) }}
                    >
                        {suggestion.tier}
                    </span>
                </div>
                <div className="suggestion-card-stats">
                    <span className="suggestion-card-cmc">{suggestion.card.ManaCost || `{${suggestion.card.CMC}}`}</span>
                    <span className="suggestion-card-gihwr">{suggestion.gihwr.toFixed(1)}%</span>
                </div>
                <div className="suggestion-card-reason">
                    {suggestion.synergyReason}
                </div>
            </div>
        </div>
    );

    const renderSection = (title: string, icon: string, cards: SuggestedCard[]) => {
        if (cards.length === 0) return null;

        return (
            <div className="suggestion-section">
                <div className="suggestion-section-header">
                    <span className="suggestion-section-icon">{icon}</span>
                    <span className="suggestion-section-title">{title}</span>
                    <span className="suggestion-section-count">({cards.length})</span>
                </div>
                <div className="suggestion-section-cards">
                    {cards.map(renderSuggestionCard)}
                </div>
            </div>
        );
    };

    if (pickedCards.length === 0) {
        return (
            <div className="cards-to-look-for">
                <div className="cards-to-look-for-header">
                    <h3>Cards to Look For</h3>
                </div>
                <div className="cards-to-look-for-empty">
                    <p>Pick some cards to get suggestions!</p>
                    <p className="cards-to-look-for-hint">
                        Suggestions will appear based on:<br />
                        • Card types and tribal synergies<br />
                        • Your color choices<br />
                        • Mana curve gaps<br />
                        • 17Lands power level
                    </p>
                </div>
            </div>
        );
    }

    const totalSuggestions =
        suggestions.highPriority.length +
        suggestions.colorSynergy.length +
        suggestions.curveFiller.length +
        suggestions.typeSynergy.length;

    if (totalSuggestions === 0) {
        return (
            <div className="cards-to-look-for">
                <div className="cards-to-look-for-header">
                    <h3>Cards to Look For</h3>
                </div>
                <div className="cards-to-look-for-empty">
                    <p>No strong synergies detected yet.</p>
                    <p className="cards-to-look-for-hint">
                        Keep picking cards with matching types or colors!
                    </p>
                </div>
            </div>
        );
    }

    return (
        <div className="cards-to-look-for">
            <div className="cards-to-look-for-header">
                <h3>Cards to Look For</h3>
                <span className="cards-to-look-for-count">{totalSuggestions} suggestions</span>
            </div>

            <div className="cards-to-look-for-content">
                {renderSection('⭐ High Priority', '⭐', suggestions.highPriority)}
                {renderSection('🎨 Color Synergy', '🎨', suggestions.colorSynergy)}
                {renderSection('⚡ Type Synergy', '⚡', suggestions.typeSynergy)}
                {renderSection('📊 Curve Filler', '📊', suggestions.curveFiller)}
            </div>
        </div>
    );
};

export default CardsToLookFor;

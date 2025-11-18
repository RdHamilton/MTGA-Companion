import React, { useState, useEffect } from 'react';
import { GetActiveDraftSessions, GetDraftPicks, GetDraftPacks, GetSetCards } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import './Draft.css';

interface DraftState {
    session: models.DraftSession | null;
    picks: models.DraftPickSession[];
    packs: models.DraftPackSession[];
    setCards: models.SetCard[];
    loading: boolean;
    error: string | null;
}

const Draft: React.FC = () => {
    const [state, setState] = useState<DraftState>({
        session: null,
        picks: [],
        packs: [],
        setCards: [],
        loading: true,
        error: null,
    });

    const [selectedCard, setSelectedCard] = useState<models.SetCard | null>(null);

    useEffect(() => {
        loadActiveDraft();
    }, []);

    const loadActiveDraft = async () => {
        try {
            setState(prev => ({ ...prev, loading: true, error: null }));

            // Get active draft sessions
            const sessions = await GetActiveDraftSessions();

            if (!sessions || sessions.length === 0) {
                setState(prev => ({
                    ...prev,
                    loading: false,
                    error: null,
                }));
                return;
            }

            const session = sessions[0]; // Use first active session

            // Load draft data
            const [picks, packs, setCards] = await Promise.all([
                GetDraftPicks(session.id),
                GetDraftPacks(session.id),
                GetSetCards(session.set_code),
            ]);

            setState({
                session,
                picks: picks || [],
                packs: packs || [],
                setCards: setCards || [],
                loading: false,
                error: null,
            });
        } catch (error) {
            console.error('Failed to load draft:', error);
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load draft',
            }));
        }
    };

    const handleCardHover = (card: models.SetCard | null) => {
        setSelectedCard(card);
    };

    const getPickedCardIds = (): Set<string> => {
        return new Set(state.picks.map(pick => pick.card_id));
    };

    if (state.loading) {
        return (
            <div className="draft-container">
                <div className="draft-loading">
                    <div className="loading-spinner"></div>
                    <p>Loading draft...</p>
                </div>
            </div>
        );
    }

    if (!state.session) {
        return (
            <div className="draft-container">
                <div className="draft-empty">
                    <h2>No Active Draft</h2>
                    <p>Start a Quick Draft in MTG Arena to begin tracking.</p>
                    <div className="empty-help">
                        <h3>How it works:</h3>
                        <ul>
                            <li>Start a Quick Draft in MTG Arena</li>
                            <li>The draft assistant will automatically detect and display</li>
                            <li>See all cards from the set with pick highlighting</li>
                            <li>View your pick history and synergies</li>
                        </ul>
                    </div>
                </div>
            </div>
        );
    }

    const pickedCardIds = getPickedCardIds();

    return (
        <div className="draft-container">
            <div className="draft-header">
                <h1>Draft Assistant</h1>
                <div className="draft-info">
                    <span className="draft-event">{state.session.event_name}</span>
                    <span className="draft-set">Set: {state.session.set_code}</span>
                    <span className="draft-picks">Picks: {state.picks.length}/{state.session.total_picks || 45}</span>
                </div>
            </div>

            <div className="draft-content">
                {/* Left: Card Grid (~25% width) */}
                <div className="card-grid-section">
                    <h2>Set Cards ({state.setCards.length})</h2>
                    <div className="card-grid">
                        {state.setCards.map(card => {
                            const isPicked = pickedCardIds.has(card.arena_id);
                            return (
                                <div
                                    key={card.id}
                                    className={`card-item ${isPicked ? 'picked' : ''}`}
                                    onMouseEnter={() => handleCardHover(card)}
                                    onMouseLeave={() => handleCardHover(null)}
                                >
                                    {card.image_url_small ? (
                                        <img src={card.image_url_small} alt={card.name} />
                                    ) : (
                                        <div className="card-placeholder">{card.name}</div>
                                    )}
                                    {isPicked && <div className="picked-indicator">✓</div>}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Right: Pick History & Details */}
                <div className="draft-details-section">
                    {/* Pick History */}
                    <div className="pick-history">
                        <h2>Pick History</h2>
                        <div className="pick-history-grid">
                            {state.picks.map((pick, index) => {
                                const card = state.setCards.find(c => c.arena_id === pick.card_id);
                                return (
                                    <div key={pick.id} className="pick-history-item">
                                        <div className="pick-number">P{pick.pack_number + 1}P{pick.pick_number + 1}</div>
                                        {card && card.image_url_small && (
                                            <img src={card.image_url_small} alt={card.name} title={card.name} />
                                        )}
                                        {card && !card.image_url_small && (
                                            <div className="card-name-small">{card.name}</div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    {/* Card Tooltip/Details */}
                    {selectedCard && (
                        <div className="card-details">
                            <h3>{selectedCard.name}</h3>
                            {selectedCard.image_url && (
                                <img src={selectedCard.image_url} alt={selectedCard.name} className="card-detail-image" />
                            )}
                            <div className="card-stats">
                                <div className="stat">
                                    <span className="stat-label">Mana Cost:</span>
                                    <span className="stat-value">{selectedCard.mana_cost || 'N/A'}</span>
                                </div>
                                <div className="stat">
                                    <span className="stat-label">Type:</span>
                                    <span className="stat-value">{selectedCard.types || 'N/A'}</span>
                                </div>
                                <div className="stat">
                                    <span className="stat-label">Rarity:</span>
                                    <span className="stat-value">{selectedCard.rarity}</span>
                                </div>
                            </div>
                            <div className="card-text">
                                <p>{selectedCard.text}</p>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default Draft;

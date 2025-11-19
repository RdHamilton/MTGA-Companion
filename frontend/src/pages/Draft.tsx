import React, { useState, useEffect } from 'react';
import { GetActiveDraftSessions, GetCompletedDraftSessions, GetDraftPicks, GetDraftPacks, GetSetCards, GetCardByArenaID, FixDraftSessionStatuses } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import TierList from '../components/TierList';
import './Draft.css';

interface DraftState {
    session: models.DraftSession | null;
    picks: models.DraftPickSession[];
    packs: models.DraftPackSession[];
    setCards: models.SetCard[];
    loading: boolean;
    error: string | null;
}

interface HistoricalDraftsState {
    sessions: models.DraftSession[];
    loading: boolean;
    error: string | null;
}

interface HistoricalDraftDetailState {
    session: models.DraftSession | null;
    picks: models.DraftPickSession[];
    packs: models.DraftPackSession[];
    pickedCards: models.SetCard[];
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

    const [historicalState, setHistoricalState] = useState<HistoricalDraftsState>({
        sessions: [],
        loading: false,
        error: null,
    });

    const [historicalDetailState, setHistoricalDetailState] = useState<HistoricalDraftDetailState>({
        session: null,
        picks: [],
        packs: [],
        pickedCards: [],
        loading: false,
        error: null,
    });

    const [selectedCard, setSelectedCard] = useState<models.SetCard | null>(null);

    useEffect(() => {
        // Fix any draft sessions that should be marked as completed
        FixDraftSessionStatuses().then((count) => {
            if (count > 0) {
                console.log(`Fixed ${count} draft session(s) status`);
            }
            // Load active draft after fixing statuses
            loadActiveDraft();
        }).catch((error) => {
            console.error('Failed to fix draft session statuses:', error);
            // Still load active draft even if fix fails
            loadActiveDraft();
        });

        // Listen for draft updates from backend
        const unsubscribe = EventsOn('draft:updated', () => {
            // Refresh both active draft and historical drafts when draft data changes
            loadActiveDraft();
        });

        return () => {
            if (unsubscribe) unsubscribe();
        };
    }, []);

    const loadHistoricalDrafts = async () => {
        try {
            setHistoricalState(prev => ({ ...prev, loading: true, error: null }));
            const sessions = await GetCompletedDraftSessions(20); // Get last 20 completed drafts
            setHistoricalState({
                sessions: sessions || [],
                loading: false,
                error: null,
            });
        } catch (error) {
            console.error('Failed to load historical drafts:', error);
            setHistoricalState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load historical drafts',
            }));
        }
    };

    const loadHistoricalDraftDetail = async (session: models.DraftSession) => {
        try {
            setHistoricalDetailState(prev => ({ ...prev, loading: true, error: null }));

            // Load picks and packs
            const [picks, packs] = await Promise.all([
                GetDraftPicks(session.ID),
                GetDraftPacks(session.ID),
            ]);

            // Get unique card IDs from picks
            const uniqueCardIds = new Set((picks || []).map(p => p.CardID));

            // Fetch each picked card
            const pickedCardsPromises = Array.from(uniqueCardIds).map(cardId =>
                GetCardByArenaID(cardId).catch(() => null)
            );
            const pickedCardsResults = await Promise.all(pickedCardsPromises);
            const pickedCards = pickedCardsResults.filter(c => c !== null) as models.SetCard[];

            setHistoricalDetailState({
                session,
                picks: picks || [],
                packs: packs || [],
                pickedCards,
                loading: false,
                error: null,
            });
        } catch (error) {
            console.error('Failed to load historical draft detail:', error);
            setHistoricalDetailState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load draft details',
            }));
        }
    };

    const handleBackToGrid = () => {
        setHistoricalDetailState({
            session: null,
            picks: [],
            packs: [],
            pickedCards: [],
            loading: false,
            error: null,
        });
    };

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
                // Load historical drafts when no active draft
                loadHistoricalDrafts();
                return;
            }

            const session = sessions[0]; // Use first active session

            // Load draft data
            const [picks, packs, setCards] = await Promise.all([
                GetDraftPicks(session.ID),
                GetDraftPacks(session.ID),
                GetSetCards(session.SetCode),
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
        return new Set(state.picks.map(pick => pick.CardID));
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

    // Historical draft detail view
    if (!state.session && historicalDetailState.session) {
        return (
            <div className="draft-container">
                <div className="draft-header">
                    <button className="btn-back" onClick={handleBackToGrid}>
                        ← Back to Draft History
                    </button>
                    <h1>Draft Replay</h1>
                    <div className="draft-info">
                        <span className="draft-event">{historicalDetailState.session.EventName}</span>
                        <span className="draft-set">Set: {historicalDetailState.session.SetCode}</span>
                        <span className="draft-picks">Picks: {historicalDetailState.picks.length}/{historicalDetailState.session.TotalPicks || 45}</span>
                    </div>
                </div>

                <div className="draft-content">
                    {/* Left: Picked Cards Only */}
                    <div className="card-grid-section">
                        <h2>Picked Cards ({historicalDetailState.pickedCards.length})</h2>
                        <div className="card-grid">
                            {historicalDetailState.pickedCards.map(card => {
                                return (
                                    <div
                                        key={card.ID}
                                        className="card-item picked"
                                        onClick={() => handleCardHover(card)}
                                    >
                                        {card.ImageURLSmall ? (
                                            <img src={card.ImageURLSmall} alt={card.Name} />
                                        ) : (
                                            <div className="card-placeholder">{card.Name}</div>
                                        )}
                                        <div className="picked-indicator">✓</div>
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    {/* Right: Pick History */}
                    <div className="draft-details-section">
                        {/* Pick History */}
                        <div className="pick-history">
                            <h2>Pick History</h2>
                            <div className="pick-history-grid">
                                {historicalDetailState.picks.map((pick) => {
                                    const card = historicalDetailState.pickedCards.find(c => c.ArenaID === pick.CardID);
                                    return (
                                        <div key={pick.ID} className="pick-history-item">
                                            <div className="pick-number">P{pick.PackNumber + 1}P{pick.PickNumber + 1}</div>
                                            {card && card.ImageURLSmall && (
                                                <img
                                                    src={card.ImageURLSmall}
                                                    alt={card.Name}
                                                    title={card.Name}
                                                    onClick={() => handleCardHover(card)}
                                                    style={{ cursor: 'pointer' }}
                                                />
                                            )}
                                            {card && !card.ImageURLSmall && (
                                                <div className="card-name-small">{card.Name}</div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    </div>
                </div>

                {/* Card Details Overlay */}
                {selectedCard && (
                    <>
                        <div className="card-details-overlay-backdrop" onClick={() => handleCardHover(null)} />
                        <div className="card-details-overlay">
                            <h3>{selectedCard.Name}</h3>
                            <p className="card-detail-type">{selectedCard.Types || 'Unknown Type'}</p>
                            <p className="card-detail-set">
                                <span>{selectedCard.SetCode}</span>
                                <span>•</span>
                                <span>{selectedCard.Rarity}</span>
                            </p>
                            {selectedCard.ImageURL && (
                                <img src={selectedCard.ImageURL} alt={selectedCard.Name} className="card-detail-image" />
                            )}
                            <div className="card-stats-section">
                                <h4>Card Stats</h4>
                                <div className="card-stats">
                                    <div className="stat">
                                        <span className="stat-label">Mana Cost</span>
                                        <span className="stat-value">{selectedCard.ManaCost || 'N/A'}</span>
                                    </div>
                                    <div className="stat">
                                        <span className="stat-label">CMC</span>
                                        <span className="stat-value">{selectedCard.CMC || 0}</span>
                                    </div>
                                    {selectedCard.Power && (
                                        <div className="stat">
                                            <span className="stat-label">Power</span>
                                            <span className="stat-value">{selectedCard.Power}</span>
                                        </div>
                                    )}
                                    {selectedCard.Toughness && (
                                        <div className="stat">
                                            <span className="stat-label">Toughness</span>
                                            <span className="stat-value">{selectedCard.Toughness}</span>
                                        </div>
                                    )}
                                </div>
                            </div>
                            {selectedCard.Text && (
                                <div className="card-text">
                                    <p>{selectedCard.Text}</p>
                                </div>
                            )}
                        </div>
                    </>
                )}
            </div>
        );
    }

    // Historical drafts grid view
    if (!state.session) {
        return (
            <div className="draft-container">
                <div className="draft-header">
                    <h1>Draft History</h1>
                    <p>Start a Quick Draft in MTG Arena to begin a new draft session</p>
                </div>

                {historicalState.loading ? (
                    <div className="draft-loading">
                        <div className="loading-spinner"></div>
                        <p>Loading draft history...</p>
                    </div>
                ) : historicalState.sessions.length === 0 ? (
                    <div className="draft-empty">
                        <h2>No Draft History</h2>
                        <p>Complete a Quick Draft in MTG Arena to see your draft history here.</p>
                        <div className="empty-help">
                            <h3>How it works:</h3>
                            <ul>
                                <li>Start a Quick Draft in MTG Arena</li>
                                <li>The draft assistant will automatically detect and display</li>
                                <li>See all cards from the set with pick highlighting</li>
                                <li>View your pick history and synergies</li>
                                <li>Completed drafts will appear here with stats</li>
                            </ul>
                        </div>
                    </div>
                ) : (
                    <div className="historical-drafts">
                        <div className="drafts-grid">
                            {historicalState.sessions.map((session) => {
                                const startDate = new Date(session.StartTime as any);
                                const formattedDate = startDate.toLocaleDateString('en-US', {
                                    month: 'short',
                                    day: 'numeric',
                                    year: 'numeric'
                                });
                                const formattedTime = startDate.toLocaleTimeString('en-US', {
                                    hour: 'numeric',
                                    minute: '2-digit'
                                });

                                return (
                                    <div key={session.ID} className="draft-card">
                                        <div className="draft-card-header">
                                            <h3>{session.EventName}</h3>
                                            <span className="draft-set-badge">{session.SetCode}</span>
                                        </div>
                                        <div className="draft-card-info">
                                            <div className="draft-stat">
                                                <span className="stat-label">Date:</span>
                                                <span className="stat-value">{formattedDate}</span>
                                            </div>
                                            <div className="draft-stat">
                                                <span className="stat-label">Time:</span>
                                                <span className="stat-value">{formattedTime}</span>
                                            </div>
                                            <div className="draft-stat">
                                                <span className="stat-label">Picks:</span>
                                                <span className="stat-value">{session.TotalPicks || 0}</span>
                                            </div>
                                        </div>
                                        <div className="draft-card-actions">
                                            <button
                                                className="btn-view-replay"
                                                onClick={() => loadHistoricalDraftDetail(session)}
                                            >
                                                View Replay
                                            </button>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                )}
            </div>
        );
    }

    // Active draft view
    const pickedCardIds = getPickedCardIds();

    return (
        <div className="draft-container">
            <div className="draft-header">
                <h1>Draft Assistant</h1>
                <div className="draft-info">
                    <span className="draft-event">{state.session.EventName}</span>
                    <span className="draft-set">Set: {state.session.SetCode}</span>
                    <span className="draft-picks">Picks: {state.picks.length}/{state.session.TotalPicks || 45}</span>
                </div>
            </div>

            <div className="draft-content">
                {/* Left: Card Grid (~25% width) - ALL SET CARDS */}
                <div className="card-grid-section">
                    <h2>Set Cards ({state.setCards.length})</h2>
                    <div className="card-grid">
                        {state.setCards.map(card => {
                            const isPicked = pickedCardIds.has(card.ArenaID);
                            return (
                                <div
                                    key={card.ID}
                                    className={`card-item ${isPicked ? 'picked' : ''}`}
                                    onClick={() => handleCardHover(card)}
                                >
                                    {card.ImageURLSmall ? (
                                        <img src={card.ImageURLSmall} alt={card.Name} />
                                    ) : (
                                        <div className="card-placeholder">{card.Name}</div>
                                    )}
                                    {isPicked && <div className="picked-indicator">✓</div>}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Right: Pick History and Tier List */}
                <div className="draft-details-section">
                    {/* Pick History */}
                    <div className="pick-history">
                        <h2>Pick History</h2>
                        <div className="pick-history-grid">
                            {state.picks.map((pick) => {
                                const card = state.setCards.find(c => c.ArenaID === pick.CardID);
                                return (
                                    <div key={pick.ID} className="pick-history-item">
                                        <div className="pick-number">P{pick.PackNumber + 1}P{pick.PickNumber + 1}</div>
                                        {card && card.ImageURLSmall && (
                                            <img
                                                src={card.ImageURLSmall}
                                                alt={card.Name}
                                                title={card.Name}
                                                onClick={() => handleCardHover(card)}
                                                style={{ cursor: 'pointer' }}
                                            />
                                        )}
                                        {card && !card.ImageURLSmall && (
                                            <div className="card-name-small">{card.Name}</div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    {/* Tier List */}
                    <TierList
                        setCode={state.session.SetCode}
                        draftFormat={state.session.EventName}
                        pickedCardIds={pickedCardIds}
                        onCardClick={(arenaId) => {
                            const card = state.setCards.find(c => c.ArenaID === String(arenaId));
                            if (card) {
                                handleCardHover(card);
                            }
                        }}
                    />
                </div>

                {/* Card Details Overlay */}
                {selectedCard && (
                    <>
                        <div className="card-details-overlay-backdrop" onClick={() => handleCardHover(null)} />
                        <div className="card-details-overlay">
                            <h3>{selectedCard.Name}</h3>
                            <p className="card-detail-type">{selectedCard.Types || 'Unknown Type'}</p>
                            <p className="card-detail-set">
                                <span>{selectedCard.SetCode}</span>
                                <span>•</span>
                                <span>{selectedCard.Rarity}</span>
                            </p>
                            {selectedCard.ImageURL && (
                                <img src={selectedCard.ImageURL} alt={selectedCard.Name} className="card-detail-image" />
                            )}
                            <div className="card-stats-section">
                                <h4>Card Stats</h4>
                                <div className="card-stats">
                                    <div className="stat">
                                        <span className="stat-label">Mana Cost</span>
                                        <span className="stat-value">{selectedCard.ManaCost || 'N/A'}</span>
                                    </div>
                                    <div className="stat">
                                        <span className="stat-label">CMC</span>
                                        <span className="stat-value">{selectedCard.CMC || 0}</span>
                                    </div>
                                    {selectedCard.Power && (
                                        <div className="stat">
                                            <span className="stat-label">Power</span>
                                            <span className="stat-value">{selectedCard.Power}</span>
                                        </div>
                                    )}
                                    {selectedCard.Toughness && (
                                        <div className="stat">
                                            <span className="stat-label">Toughness</span>
                                            <span className="stat-value">{selectedCard.Toughness}</span>
                                        </div>
                                    )}
                                </div>
                            </div>
                            {selectedCard.Text && (
                                <div className="card-text">
                                    <p>{selectedCard.Text}</p>
                                </div>
                            )}
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

export default Draft;

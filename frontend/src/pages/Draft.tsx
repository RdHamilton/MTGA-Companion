import React, { useState, useEffect, useMemo, useRef } from 'react';
import { GetActiveDraftSessions, GetCompletedDraftSessions, GetDraftPicks, GetDraftPacks, GetSetCards, GetCardByArenaID, AnalyzeSessionPickQuality, GetPickAlternatives, GetDraftGrade, GetCardRatings, PauseReplay, ResumeReplay, StopReplay } from '../../wailsjs/go/main/App';
import { models, pickquality, grading, main } from '../../wailsjs/go/models';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { getReplayState } from '../App';
import TierList from '../components/TierList';
import { DraftGrade } from '../components/DraftGrade';
import { WinRatePrediction } from '../components/WinRatePrediction';
import CardsToLookFor from '../components/CardsToLookFor';
import MissingCards from '../components/MissingCards';
import { analyzeSynergies, shouldHighlightCard } from '../utils/synergy';
import './Draft.css';

interface DraftState {
    session: models.DraftSession | null;
    picks: models.DraftPickSession[];
    packs: models.DraftPackSession[];
    setCards: models.SetCard[];
    ratings: main.CardRatingWithTier[];
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
    grade: grading.DraftGrade | null;
    loading: boolean;
    error: string | null;
}

const Draft: React.FC = () => {
    const [state, setState] = useState<DraftState>({
        session: null,
        picks: [],
        packs: [],
        setCards: [],
        ratings: [],
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
        grade: null,
        loading: false,
        error: null,
    });

    const [selectedCard, setSelectedCard] = useState<models.SetCard | null>(null);
    const [isAnalyzing, setIsAnalyzing] = useState(false);
    const [pickAlternatives, setPickAlternatives] = useState<Map<string, pickquality.PickQuality>>(new Map());

    // Refs for deduplication and debouncing
    const loadingRef = useRef<boolean>(false);
    const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        // Load active draft immediately
        // Note: We don't call FixDraftSessionStatuses() here because:
        // 1. It interferes with replay mode (marks replayed sessions as completed)
        // 2. Session status management should be handled by the daemon, not the frontend
        loadActiveDraft();

        // Listen for draft updates from backend
        const unsubscribe = EventsOn('draft:updated', () => {
            // Debounce draft:updated events to prevent rapid-fire database queries
            console.log('[Draft.tsx] Received draft:updated event, debouncing...');
            debouncedLoadActiveDraft();
        });

        return () => {
            if (unsubscribe) unsubscribe();
            // Clear debounce timer on cleanup
            if (debounceTimerRef.current) {
                clearTimeout(debounceTimerRef.current);
            }
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

            // Try to load grade if it exists
            let grade: grading.DraftGrade | null = null;
            try {
                grade = await GetDraftGrade(session.ID);
            } catch {
                // Grade doesn't exist yet, that's okay
            }

            setHistoricalDetailState({
                session,
                picks: picks || [],
                packs: packs || [],
                pickedCards,
                grade,
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
            grade: null,
            loading: false,
            error: null,
        });
    };

    const loadActiveDraft = async () => {
        // Deduplication: Skip if already loading
        if (loadingRef.current) {
            console.log('[loadActiveDraft] Already loading, skipping duplicate call');
            return;
        }

        loadingRef.current = true;

        try {
            console.log('[loadActiveDraft] Starting...');
            setState(prev => ({ ...prev, loading: true, error: null }));

            // Get active draft sessions
            const sessions = await GetActiveDraftSessions();
            console.log('[loadActiveDraft] Got sessions:', sessions?.length || 0);

            if (!sessions || sessions.length === 0) {
                console.log('[loadActiveDraft] No active sessions, loading historical drafts');
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
            console.log('[loadActiveDraft] Session:', session.ID, 'SetCode:', session.SetCode);

            // Load draft data
            console.log('[loadActiveDraft] Loading data for session:', session.ID, 'SetCode:', session.SetCode);
            const [picks, packs, setCards, ratings] = await Promise.all([
                GetDraftPicks(session.ID),
                GetDraftPacks(session.ID),
                GetSetCards(session.SetCode),
                GetCardRatings(session.SetCode, session.DraftType),
            ]);
            console.log('[loadActiveDraft] Data loaded successfully:');
            console.log('  - Picks:', picks?.length || 0);
            console.log('  - Packs:', packs?.length || 0);
            console.log('  - SetCards:', setCards?.length || 0);
            console.log('  - Ratings:', ratings?.length || 0);

            // Always store ALL set cards for synergy analysis
            // In replay mode, we'll filter the display later in the render
            console.log('[loadActiveDraft] Setting state with all set cards');
            setState({
                session,
                picks: picks || [],
                packs: packs || [],
                setCards: setCards || [],
                ratings: ratings || [],
                loading: false,
                error: null,
            });

            // Auto-analyze picks after draft loads (if picks exist)
            if (session.ID && picks && picks.length > 0) {
                console.log('[loadActiveDraft] Auto-analyzing', picks.length, 'picks for session', session.ID);
                try {
                    await AnalyzeSessionPickQuality(session.ID);
                    console.log('[loadActiveDraft] Auto-analysis complete');
                    // Reload picks to get updated quality data
                    const updatedPicks = await GetDraftPicks(session.ID);
                    setState(prev => ({
                        ...prev,
                        picks: updatedPicks || [],
                    }));
                } catch (error) {
                    console.error('[loadActiveDraft] Failed to auto-analyze picks:', error);
                    // Don't set error state - this is a non-critical failure
                }
            }
        } catch (error) {
            console.error('Failed to load draft:', error);
            setState(prev => ({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : 'Failed to load draft',
            }));
        } finally {
            loadingRef.current = false;
        }
    };

    const debouncedLoadActiveDraft = () => {
        // Clear any existing timer
        if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current);
        }

        // Set new timer to delay execution by 500ms
        debounceTimerRef.current = setTimeout(() => {
            loadActiveDraft();
        }, 500);
    };

    const handleCardHover = (card: models.SetCard | null) => {
        setSelectedCard(card);
    };

    const getPickedCardIds = (): Set<string> => {
        return new Set(state.picks.map(pick => pick.CardID));
    };

    const getPickedCards = (): models.SetCard[] => {
        const pickedCardIds = getPickedCardIds();
        return state.setCards.filter(card => pickedCardIds.has(card.ArenaID));
    };

    const handleAnalyzeDraft = async () => {
        if (!state.session) return;

        try {
            setIsAnalyzing(true);
            await AnalyzeSessionPickQuality(state.session.ID);
            // Reload picks to get updated quality data
            await loadActiveDraft();
        } catch (error) {
            console.error('Failed to analyze draft:', error);
        } finally {
            setIsAnalyzing(false);
        }
    };

    const getPickQualityClass = (grade: string | undefined): string => {
        if (!grade) return '';
        switch (grade) {
            case 'A+':
                return 'quality-a-plus';
            case 'A':
                return 'quality-a';
            case 'B':
                return 'quality-b';
            case 'C':
                return 'quality-c';
            case 'D':
                return 'quality-d';
            case 'F':
                return 'quality-f';
            case 'N/A':
                return 'quality-n-a';
            default:
                return '';
        }
    };

    const loadPickAlternatives = async (sessionID: string, packNum: number, pickNum: number) => {
        const key = `${sessionID}-${packNum}-${pickNum}`;
        if (pickAlternatives.has(key)) {
            return pickAlternatives.get(key);
        }

        try {
            const alternatives = await GetPickAlternatives(sessionID, packNum, pickNum);
            if (alternatives) {
                setPickAlternatives(prev => new Map(prev).set(key, alternatives));
                return alternatives;
            }
        } catch (error) {
            console.error('Failed to load pick alternatives:', error);
        }
        return null;
    };

    // ========================================
    // HOOKS MUST BE CALLED UNCONDITIONALLY
    // Call all hooks at the top before any conditional returns
    // ========================================

    // Get replay state
    const replayState = getReplayState();
    const isReplayMode = replayState.isActive;
    const isReplayPaused = replayState.isPaused;

    // Compute display cards for active draft (used in render)
    const pickedCardIds = state.session ? getPickedCardIds() : new Set<string>();
    const displayCards = useMemo(() => {
        if (!state.session) return [];
        console.log('[displayCards] Computing display cards. isReplayMode:', isReplayMode, 'picks:', state.picks.length, 'setCards:', state.setCards.length);
        if (isReplayMode && state.picks.length > 0) {
            // Recalculate picked card IDs inside useMemo to ensure correct dependencies
            const pickedIds = new Set(state.picks.map(pick => pick.CardID));
            const filtered = state.setCards.filter(card => pickedIds.has(card.ArenaID));
            console.log('[displayCards] Replay mode: filtered to', filtered.length, 'picked cards');
            return filtered;
        }
        console.log('[displayCards] Normal mode: showing all', state.setCards.length, 'cards');
        return state.setCards;
    }, [state.session, isReplayMode, state.setCards, state.picks]);

    // Calculate synergy analysis for highlighting
    const synergyAnalysis = useMemo(() => {
        if (!state.session) return { types: [], colors: { colors: [], count: 0, percentage: 0 }, curve: { avgCMC: 0, archetype: 'midrange' as const, gaps: [] }, pickedCardsCount: 0 };
        const pickedCards = getPickedCards();
        return analyzeSynergies(pickedCards);
    }, [state.session, state.picks, state.setCards]);

    // ========================================
    // NOW WE CAN DO CONDITIONAL RENDERING
    // ========================================

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
                    {historicalDetailState.picks.length > 0 && historicalDetailState.session && (
                        <>
                            <DraftGrade
                                sessionID={historicalDetailState.session.ID}
                                showCalculateButton={true}
                                onGradeCalculated={async (grade) => {
                                    // Reload the grade to refresh best/worst pick highlighting
                                    setHistoricalDetailState(prev => ({ ...prev, grade }));
                                }}
                            />
                            <WinRatePrediction
                                sessionID={historicalDetailState.session.ID}
                                showPredictButton={true}
                                onPredictionCalculated={(pred) => {
                                    console.log('Prediction calculated:', pred);
                                }}
                            />
                        </>
                    )}
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
                                    const hasQuality = pick.PickQualityGrade !== null && pick.PickQualityGrade !== undefined;
                                    const altKey = `${pick.SessionID}-${pick.PackNumber}-${pick.PickNumber}`;
                                    const alternatives = pickAlternatives.get(altKey);

                                    // Check if this pick is in best/worst picks
                                    const isBestPick = historicalDetailState.grade?.best_picks?.some(bp =>
                                        card && bp.includes(card.Name)
                                    );
                                    const isWorstPick = historicalDetailState.grade?.worst_picks?.some(wp =>
                                        card && wp.includes(card.Name)
                                    );

                                    let highlightClass = '';
                                    if (isBestPick) highlightClass = 'best-pick-highlight';
                                    if (isWorstPick) highlightClass = 'worst-pick-highlight';

                                    return (
                                        <div key={pick.ID} className={`pick-history-item ${highlightClass}`}>
                                            <div className="pick-number">P{pick.PackNumber + 1}P{pick.PickNumber}</div>
                                            <div className="card-image-container">
                                                {card && card.ImageURLSmall && (
                                                    <img
                                                        src={card.ImageURLSmall}
                                                        alt={card.Name}
                                                        title={card.Name}
                                                        onClick={() => handleCardHover(card)}
                                                        style={{ cursor: 'pointer' }}
                                                        onMouseEnter={() => {
                                                            if (hasQuality && !alternatives) {
                                                                loadPickAlternatives(pick.SessionID, pick.PackNumber, pick.PickNumber);
                                                            }
                                                        }}
                                                    />
                                                )}
                                                {card && !card.ImageURLSmall && (
                                                    <div className="card-name-small">{card.Name}</div>
                                                )}
                                                {hasQuality && (
                                                    <div className={`pick-quality-badge ${getPickQualityClass(pick.PickQualityGrade)}`}>
                                                        {pick.PickQualityGrade}
                                                    </div>
                                                )}
                                            </div>
                                            {hasQuality && alternatives && (
                                                <div className="pick-quality-tooltip">
                                                    <h4>Pick Quality Analysis</h4>
                                                    <div className="picked-stats">
                                                        <div>
                                                            <span className="label">Grade:</span>
                                                            <span className="value">{alternatives.grade}</span>
                                                        </div>
                                                        <div>
                                                            <span className="label">Rank in Pack:</span>
                                                            <span className="value">#{alternatives.rank}</span>
                                                        </div>
                                                        <div>
                                                            <span className="label">GIHWR:</span>
                                                            <span className="value">{alternatives.picked_card_gihwr.toFixed(1)}%</span>
                                                        </div>
                                                    </div>
                                                    {alternatives.alternatives && alternatives.alternatives.length > 0 && (
                                                        <div className="alternatives">
                                                            <h5>Better Options in Pack:</h5>
                                                            {alternatives.alternatives.slice(0, 3).map((alt: pickquality.Alternative, idx: number) => (
                                                                <div key={idx} className="alternative-card">
                                                                    <span className="card-name">{alt.card_name}</span>
                                                                    <span className="gihwr">{alt.gihwr.toFixed(1)}%</span>
                                                                </div>
                                                            ))}
                                                        </div>
                                                    )}
                                                </div>
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
                                            <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                                                <span className="draft-set-badge">{session.SetCode}</span>
                                                <DraftGrade sessionID={session.ID} compact={true} />
                                                <WinRatePrediction sessionID={session.ID} compact={true} />
                                            </div>
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

    // Active draft view - all hooks have been called at the top already
    console.log('[Draft] Rendering active draft. DisplayCards:', displayCards.length, 'Picks:', state.picks.length);

    return (
        <div className="draft-container">
            {/* Replay Controls - Bottom Left */}
            {isReplayMode && (
                <div className="replay-controls">
                    {isReplayPaused ? (
                        <button
                            className="replay-btn replay-resume"
                            onClick={async () => {
                                try {
                                    await ResumeReplay();
                                } catch (error) {
                                    console.error('Failed to resume replay:', error);
                                }
                            }}
                            title="Resume replay"
                        >
                            ▶️ Resume
                        </button>
                    ) : (
                        <button
                            className="replay-btn replay-pause"
                            onClick={async () => {
                                try {
                                    await PauseReplay();
                                } catch (error) {
                                    console.error('Failed to pause replay:', error);
                                }
                            }}
                            title="Pause replay"
                        >
                            ⏸️ Pause
                        </button>
                    )}
                    <button
                        className="replay-btn replay-stop"
                        onClick={async () => {
                            try {
                                await StopReplay();
                            } catch (error) {
                                console.error('Failed to stop replay:', error);
                            }
                        }}
                        title="Stop replay and return to normal mode"
                    >
                        ⏹️ Stop
                    </button>
                </div>
            )}

            <div className="draft-header">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
                    <div>
                        <h1>Draft Assistant</h1>
                        <div className="draft-info">
                            <span className="draft-event">{state.session.EventName}</span>
                            <span className="draft-set">Set: {state.session.SetCode}</span>
                            <span className="draft-picks">Picks: {state.picks.length}/{state.session.TotalPicks || 45}</span>
                        </div>
                    </div>
                    <button
                        className="btn-analyze-draft"
                        onClick={handleAnalyzeDraft}
                        disabled={isAnalyzing || state.picks.length === 0}
                    >
                        {isAnalyzing ? (
                            <>
                                <div className="spinner"></div>
                                Analyzing...
                            </>
                        ) : (
                            '🎯 Analyze Pick Quality'
                        )}
                    </button>
                </div>
            </div>

            <div className="draft-content">
                {/* Left: Card Grid (~25% width) */}
                <div className="card-grid-section">
                    <h2>{isReplayMode ? 'Picked Cards' : 'Set Cards'} ({displayCards.length})</h2>
                    <div className="card-grid">
                        {displayCards.map(card => {
                            const isPicked = pickedCardIds.has(card.ArenaID);
                            const pick = isPicked ? state.picks.find(p => p.CardID === card.ArenaID) : null;
                            const hasGrade = pick && pick.PickQualityGrade;
                            const hasSynergy = !isPicked && shouldHighlightCard(card, synergyAnalysis);
                            return (
                                <div
                                    key={card.ID}
                                    className={`card-item ${isPicked ? 'picked' : ''} ${hasSynergy ? 'synergy-highlight' : ''}`}
                                    onClick={() => handleCardHover(card)}
                                >
                                    {card.ImageURLSmall ? (
                                        <img src={card.ImageURLSmall} alt={card.Name} />
                                    ) : (
                                        <div className="card-placeholder">{card.Name}</div>
                                    )}
                                    {isPicked && <div className="picked-indicator">✓</div>}
                                    {hasSynergy && !isPicked && <div className="synergy-indicator">★</div>}
                                    {hasGrade && (
                                        <div className={`pick-quality-badge ${getPickQualityClass(pick!.PickQualityGrade)}`}>
                                            {pick!.PickQualityGrade}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Middle: Cards to Look For Panel */}
                <div className="cards-to-look-for-section">
                    {/* Missing Cards Analysis */}
                    {state.session && state.picks.length > 0 && (
                        <MissingCards
                            sessionID={state.session.ID}
                            packNumber={state.picks[state.picks.length - 1]?.PackNumber ?? 0}
                            pickNumber={state.picks[state.picks.length - 1]?.PickNumber ?? 1}
                        />
                    )}

                    <CardsToLookFor
                        pickedCards={getPickedCards()}
                        availableCards={state.setCards}
                        ratings={state.ratings}
                        onCardClick={(card) => handleCardHover(card)}
                    />
                </div>

                {/* Right: Pick History and Tier List */}
                <div className="draft-details-section">
                    {/* Pick History */}
                    <div className="pick-history">
                        <h2>Pick History</h2>
                        <div className="pick-history-grid">
                            {state.picks.map((pick) => {
                                const card = state.setCards.find(c => c.ArenaID === pick.CardID);
                                const hasQuality = pick.PickQualityGrade !== null && pick.PickQualityGrade !== undefined;
                                const altKey = `${pick.SessionID}-${pick.PackNumber}-${pick.PickNumber}`;
                                const alternatives = pickAlternatives.get(altKey);

                                return (
                                    <div key={pick.ID} className="pick-history-item">
                                        <div className="pick-number">P{pick.PackNumber + 1}P{pick.PickNumber}</div>
                                        <div className="card-image-container">
                                            {card && card.ImageURLSmall && (
                                                <img
                                                    src={card.ImageURLSmall}
                                                    alt={card.Name}
                                                    title={card.Name}
                                                    onClick={() => handleCardHover(card)}
                                                    style={{ cursor: 'pointer' }}
                                                    onMouseEnter={() => {
                                                        if (hasQuality && !alternatives) {
                                                            loadPickAlternatives(pick.SessionID, pick.PackNumber, pick.PickNumber);
                                                        }
                                                    }}
                                                />
                                            )}
                                            {card && !card.ImageURLSmall && (
                                                <div className="card-name-small">{card.Name}</div>
                                            )}
                                            {hasQuality && (
                                                <div className={`pick-quality-badge ${getPickQualityClass(pick.PickQualityGrade)}`}>
                                                    {pick.PickQualityGrade}
                                                </div>
                                            )}
                                        </div>
                                        {hasQuality && alternatives && (
                                            <div className="pick-quality-tooltip">
                                                <h4>Pick Quality Analysis</h4>
                                                <div className="picked-stats">
                                                    <div>
                                                        <span className="label">Grade:</span>
                                                        <span className="value">{alternatives.grade}</span>
                                                    </div>
                                                    <div>
                                                        <span className="label">Rank in Pack:</span>
                                                        <span className="value">#{alternatives.rank}</span>
                                                    </div>
                                                    <div>
                                                        <span className="label">GIHWR:</span>
                                                        <span className="value">{alternatives.picked_card_gihwr.toFixed(1)}%</span>
                                                    </div>
                                                </div>
                                                {alternatives.alternatives && alternatives.alternatives.length > 0 && (
                                                    <div className="alternatives">
                                                        <h5>Better Options in Pack:</h5>
                                                        {alternatives.alternatives.slice(0, 3).map((alt: pickquality.Alternative, idx: number) => (
                                                            <div key={idx} className="alternative-card">
                                                                <span className="card-name">{alt.card_name}</span>
                                                                <span className="gihwr">{alt.gihwr.toFixed(1)}%</span>
                                                            </div>
                                                        ))}
                                                    </div>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    {/* Tier List */}
                    <TierList
                        setCode={state.session.SetCode}
                        draftFormat={state.session.DraftType}
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

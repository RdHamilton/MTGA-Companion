import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  GetDeck,
  AddCard,
  RemoveCard,
  GetDeckStatistics,
  GetDeckByDraftEvent,
  CreateDeck,
  GetActiveDraftSessions,
  GetCompletedDraftSessions,
  GetDraftPicks,
  GetRecommendations,
} from '../../wailsjs/go/main/App';
import { models, gui } from '../../wailsjs/go/models';
import DeckList from '../components/DeckList';
import CardSearch from '../components/CardSearch';
import './DeckBuilder.css';

export default function DeckBuilder() {
  const { deckID } = useParams<{ deckID?: string }>();
  const { draftEventID } = useParams<{ draftEventID?: string }>();
  const navigate = useNavigate();
  const creatingDeckRef = useRef(false);

  const [deck, setDeck] = useState<models.Deck | null>(null);
  const [cards, setCards] = useState<models.DeckCard[]>([]);
  const [tags, setTags] = useState<models.DeckTag[]>([]);
  const [statistics, setStatistics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCardSearch, setShowCardSearch] = useState(false);
  const [draftCardIDs, setDraftCardIDs] = useState<number[]>([]);
  const [recommendations, setRecommendations] = useState<gui.CardRecommendation[]>([]);
  const [showRecommendations, setShowRecommendations] = useState(false);
  const [loadingRecommendations, setLoadingRecommendations] = useState(false);
  const [addingLands, setAddingLands] = useState(false);

  // Load deck data
  useEffect(() => {
    const loadDeck = async () => {
      setLoading(true);
      setError(null);

      try {
        let deckData;

        if (deckID) {
          // Load by deck ID
          deckData = await GetDeck(deckID);
        } else if (draftEventID) {
          // Load by draft event ID, create if doesn't exist
          deckData = await GetDeckByDraftEvent(draftEventID);

          if (!deckData || !deckData.deck) {
            // No deck exists yet - create one from draft picks
            // Guard against duplicate creation (React.StrictMode can cause double-invocation)
            if (creatingDeckRef.current) {
              setLoading(false);
              return;
            }

            try {
              creatingDeckRef.current = true;

              // Get draft session to get the event name for the deck
              const [activeSessions, completedSessions] = await Promise.all([
                GetActiveDraftSessions(),
                GetCompletedDraftSessions(100), // Get last 100 completed drafts
              ]);
              const allSessions = [...activeSessions, ...completedSessions];
              const session = allSessions.find((s: any) => s.ID === draftEventID);

              if (!session) {
                setError('Draft session not found');
                setLoading(false);
                creatingDeckRef.current = false;
                return;
              }

              const deckName = `${session.EventName} Draft`;

              // Create deck linked to this draft event
              const newDeck = await CreateDeck(deckName, 'limited', 'draft', draftEventID);

              // Load the newly created deck
              deckData = await GetDeck(newDeck.ID);
            } catch (createErr) {
              setError(createErr instanceof Error ? createErr.message : 'Failed to create deck from draft');
              setLoading(false);
              creatingDeckRef.current = false;
              return;
            } finally {
              creatingDeckRef.current = false;
            }
          }
        } else {
          setError('No deck ID or draft event ID provided');
          setLoading(false);
          return;
        }

        if (!deckData.deck) {
          setError('Invalid deck data');
          setLoading(false);
          return;
        }

        setDeck(deckData.deck);
        setCards(deckData.cards || []);
        setTags(deckData.tags || []);

        // Load statistics
        const stats = await GetDeckStatistics(deckData.deck.ID);
        setStatistics(stats);

        // If this is a draft deck, get the draft card IDs
        if (deckData.deck.Source === 'draft' && deckData.deck.DraftEventID) {
          try {
            const picks = await GetDraftPicks(deckData.deck.DraftEventID);
            // Extract unique card IDs from draft picks
            const uniqueCardIDs = Array.from(
              new Set(picks.map((pick) => parseInt(pick.CardID, 10)))
            ).filter((id) => !isNaN(id));
            setDraftCardIDs(uniqueCardIDs);
          } catch (pickErr) {
            console.error('Failed to load draft picks:', pickErr);
            setDraftCardIDs([]);
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load deck');
        console.error('Failed to load deck:', err);
      } finally {
        setLoading(false);
      }
    };

    if (deckID || draftEventID) {
      loadDeck();
    } else {
      setLoading(false);
    }
  }, [deckID, draftEventID]);

  const handleAddCard = async (cardID: number, quantity: number, board: 'main' | 'sideboard') => {
    if (!deck) return;

    try {
      await AddCard(deck.ID, cardID, quantity, board, deck.Source === 'draft');

      // Reload deck data
      const deckData = await GetDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics
      const stats = await GetDeckStatistics(deck.ID);
      setStatistics(stats);

      // Reload recommendations after adding a card
      if (deckData.cards && deckData.cards.length >= 3) {
        loadRecommendations();
      }
    } catch (err) {
      throw err; // Re-throw to let CardSearch handle the error
    }
  };

  const handleRemoveCard = async (cardID: number, board: string) => {
    if (!deck) return;

    try {
      await RemoveCard(deck.ID, cardID, board);

      // Reload deck data
      const deckData = await GetDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics
      const stats = await GetDeckStatistics(deck.ID);
      setStatistics(stats);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to remove card');
    }
  };

  const loadRecommendations = async () => {
    if (!deck) return;

    setLoadingRecommendations(true);
    try {
      const request: gui.GetRecommendationsRequest = {
        deckID: deck.ID,
        maxResults: 10,
        minScore: 0.3,
        includeLands: true,
        onlyDraftPool: deck.Source === 'draft',
      };

      const response = await GetRecommendations(request);
      if (response.error) {
        console.error('Recommendations error:', response.error);
        setRecommendations([]);
      } else {
        setRecommendations(response.recommendations || []);
        if (response.recommendations && response.recommendations.length > 0) {
          setShowRecommendations(true);
        }
      }
    } catch (err) {
      console.error('Failed to load recommendations:', err);
      setRecommendations([]);
    } finally {
      setLoadingRecommendations(false);
    }
  };

  const handleAddSuggestedLands = async () => {
    console.log('handleAddSuggestedLands called');
    if (!deck || !statistics) {
      console.error('Missing deck or statistics:', { deck, statistics });
      return;
    }

    setAddingLands(true);
    try {
      // Use statistics colors if available (backend returns colors, not colorDistribution)
      const colors = (statistics as any).colors || {};
      console.log('Full statistics object:', statistics);
      console.log('Color distribution from backend:', colors);

      // Calculate color distribution from mainboard cards
      // Only count mono-colored cards for land distribution
      const colorCounts = {
        W: colors.white || 0,
        U: colors.blue || 0,
        B: colors.black || 0,
        R: colors.red || 0,
        G: colors.green || 0,
      };

      console.log('Color counts (mono-colored only):', colorCounts);
      console.log('Color counts after assignment - W:', colorCounts.W, 'U:', colorCounts.U, 'B:', colorCounts.B, 'R:', colorCounts.R, 'G:', colorCounts.G);

      // Calculate target: 40 cards for limited, 60 for constructed
      const targetDeckSize = deck.Format === 'limited' ? 40 : 60;
      const currentMainboard = statistics.totalMainboard || 0;
      const currentLands = ((statistics as any).lands?.total) || 0;
      console.log('Deck stats:', { targetDeckSize, currentMainboard, currentLands });

      // Calculate how many more lands we need
      const nonLandCards = currentMainboard - currentLands;
      const landsNeeded = Math.max(0, targetDeckSize - nonLandCards - currentLands);
      console.log('Lands needed:', landsNeeded);

      if (landsNeeded === 0) {
        console.log('Deck already has enough lands');
        window.alert('Your deck already has enough lands!');
        return;
      }

      // Calculate total color weight
      const totalColors = Object.values(colorCounts).reduce((sum, count) => sum + count, 0);
      console.log('Total colors:', totalColors);

      if (totalColors === 0) {
        console.log('No colors detected');
        window.alert('Could not determine deck colors. Please add more colored cards first.');
        return;
      }

      // Basic land arena IDs (these are standard across all sets)
      const basicLands: Record<string, { name: string; arenaID: number }> = {
        W: { name: 'Plains', arenaID: 81716 },
        U: { name: 'Island', arenaID: 81717 },
        B: { name: 'Swamp', arenaID: 81718 },
        R: { name: 'Mountain', arenaID: 81719 },
        G: { name: 'Forest', arenaID: 81720 },
      };

      // Distribute lands proportionally
      const landDistribution: Record<string, number> = {};
      let landsAllocated = 0;

      // First pass: allocate proportionally
      Object.keys(colorCounts).forEach((color) => {
        const proportion = colorCounts[color as keyof typeof colorCounts] / totalColors;
        const count = Math.floor(landsNeeded * proportion);
        landDistribution[color] = count;
        landsAllocated += count;
      });

      // Second pass: distribute remaining lands to most prominent colors
      let remaining = landsNeeded - landsAllocated;
      const sortedColors = Object.keys(colorCounts).sort(
        (a, b) => colorCounts[b as keyof typeof colorCounts] - colorCounts[a as keyof typeof colorCounts]
      );

      for (let i = 0; i < remaining; i++) {
        const color = sortedColors[i % sortedColors.length];
        landDistribution[color] = (landDistribution[color] || 0) + 1;
      }

      // Add lands to deck
      console.log('Land distribution:', landDistribution);
      for (const [color, count] of Object.entries(landDistribution)) {
        if (count > 0 && color in basicLands) {
          const land = basicLands[color as keyof typeof basicLands];
          console.log(`Adding ${count}x ${land.name} (arena_id=${land.arenaID})`);
          await AddCard(deck.ID, land.arenaID, count, 'main', false);
        }
      }

      // Reload deck data
      console.log('Reloading deck data...');
      const deckData = await GetDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics
      const stats = await GetDeckStatistics(deck.ID);
      setStatistics(stats);

      console.log(`Successfully added ${landsNeeded} lands!`);
      window.alert(`Added ${landsNeeded} suggested lands to your deck!`);
    } catch (err) {
      console.error('Error adding lands:', err);
      window.alert(err instanceof Error ? err.message : 'Failed to add lands');
    } finally {
      setAddingLands(false);
    }
  };

  // Create a map of existing cards for CardSearch
  const existingCardsMap = new Map(
    cards.map((card) => [
      card.CardID,
      { quantity: card.Quantity, board: card.Board },
    ])
  );

  if (loading) {
    return (
      <div className="deck-builder loading-state">
        <div className="loading-spinner"></div>
        <p>Loading deck...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="deck-builder error-state">
        <div className="error-icon">⚠️</div>
        <h2>Error Loading Deck</h2>
        <p>{error}</p>
        <button onClick={() => navigate('/decks')} className="back-button">
          Back to Decks
        </button>
      </div>
    );
  }

  if (!deck) {
    return (
      <div className="deck-builder error-state">
        <div className="error-icon">📦</div>
        <h2>No Deck Found</h2>
        <p>The requested deck could not be found.</p>
        <button onClick={() => navigate('/decks')} className="back-button">
          Back to Decks
        </button>
      </div>
    );
  }

  return (
    <div className="deck-builder">
      {/* Header */}
      <div className="deck-builder-header">
        <button onClick={() => navigate('/decks')} className="back-button">
          ← Back to Decks
        </button>
        <h1>Deck Builder</h1>
        <div className="header-actions">
          <button
            className={`toggle-search-button ${showCardSearch ? 'active' : ''}`}
            onClick={() => setShowCardSearch(!showCardSearch)}
          >
            {showCardSearch ? '✕ Hide Search' : '🔍 Add Cards'}
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="deck-builder-content">
        {/* Deck List (always visible) */}
        <div className="deck-list-panel">
          <DeckList
            deck={deck}
            cards={cards}
            tags={tags}
            statistics={statistics}
            onRemoveCard={handleRemoveCard}
          />
        </div>

        {/* Card Search (toggleable) */}
        {showCardSearch && (
          <div className="card-search-panel">
            <CardSearch
              isDraftDeck={deck.Source === 'draft'}
              draftCardIDs={draftCardIDs}
              existingCards={existingCardsMap}
              onAddCard={handleAddCard}
              onRemoveCard={handleRemoveCard}
            />
          </div>
        )}

        {/* Recommendations Panel (toggleable) */}
        {showRecommendations && (
          <div className="recommendations-panel">
            <div className="recommendations-header">
              <h3>Card Recommendations</h3>
              <button className="close-recommendations" onClick={() => setShowRecommendations(false)}>
                ✕
              </button>
            </div>

            {loadingRecommendations ? (
              <div className="recommendations-loading">Loading recommendations...</div>
            ) : recommendations.length === 0 ? (
              <div className="recommendations-empty">
                No recommendations available. Add more cards to get suggestions!
              </div>
            ) : (
              <div className="recommendations-list">
                {recommendations.map((rec) => (
                  <div key={rec.cardID} className="recommendation-card">
                    {rec.imageURI && (
                      <img src={rec.imageURI} alt={rec.name} className="rec-card-image" />
                    )}
                    <div className="rec-card-info">
                      <div className="rec-card-name">{rec.name}</div>
                      <div className="rec-card-type">{rec.typeLine}</div>
                      {rec.manaCost && <div className="rec-card-mana">{rec.manaCost}</div>}
                      <div className="rec-score">
                        Score: {(rec.score * 100).toFixed(0)}% | Confidence: {(rec.confidence * 100).toFixed(0)}%
                      </div>
                      <div className="rec-reasoning">{rec.reasoning}</div>
                    </div>
                    <div className="rec-card-actions">
                      <button
                        className="add-rec-button"
                        onClick={() => handleAddCard(rec.cardID, 1, 'main')}
                      >
                        + Add
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Quick Actions Footer */}
      <div className="deck-builder-footer">
        <div className="quick-stats">
          <span>Mainboard: {statistics?.totalMainboard || 0}</span>
          <span>Sideboard: {statistics?.totalSideboard || 0}</span>
          <span>Avg CMC: {statistics?.averageCMC?.toFixed(2) || 'N/A'}</span>
        </div>
        <div className="quick-actions">
          <button className="action-button" title="Export deck">
            ⤓ Export
          </button>
          <button
            className={`action-button ${showRecommendations ? 'active' : ''}`}
            title="Get recommendations"
            onClick={() => {
              if (!showRecommendations && recommendations.length === 0) {
                loadRecommendations();
              }
              setShowRecommendations(!showRecommendations);
            }}
          >
            ✨ Suggestions
          </button>
          <button
            className="action-button"
            title="Add suggested lands based on deck colors"
            disabled={addingLands || (statistics?.totalMainboard || 0) < 2}
            onClick={handleAddSuggestedLands}
          >
            {addingLands ? '⏳ Adding...' : '🏔️ Add Lands'}
          </button>
          <button className="action-button" title="Validate deck">
            ✓ Validate
          </button>
        </div>
      </div>
    </div>
  );
}

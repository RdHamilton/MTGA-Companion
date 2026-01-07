import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { decks, drafts } from '@/services/api';
import type { CardWithOwnership, SuggestedLandResponse } from '@/services/api/decks';
import { ApiRequestError } from '@/services/apiClient';
import { downloadTextFile } from '@/utils/download';
import { models, gui } from '@/types/models';
import DeckList from '../components/DeckList';
import CardSearch from '../components/CardSearch';
import RecommendationCard from '../components/RecommendationCard';
import SuggestDecksModal from '../components/SuggestDecksModal';
import BuildAroundSeedModal from '../components/BuildAroundSeedModal';
import './DeckBuilder.css';

// Export deck to file using native file save dialog
async function exportDeckToFile(deckId: string, format: string = 'txt'): Promise<void> {
  const response = await decks.exportDeck(deckId, { format });
  downloadTextFile(response.content, response.filename || `deck.${format}`);
}

export default function DeckBuilder() {
  const { deckID } = useParams<{ deckID?: string }>();
  const { draftEventID } = useParams<{ draftEventID?: string }>();
  const navigate = useNavigate();
  const creatingDeckRef = useRef(false);

  const [deck, setDeck] = useState<models.Deck | null>(null);
  const [cards, setCards] = useState<models.DeckCard[]>([]);
  const [tags, setTags] = useState<models.DeckTag[]>([]);
  const [statistics, setStatistics] = useState<gui.DeckStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCardSearch, setShowCardSearch] = useState(false);
  const [draftCardIDs, setDraftCardIDs] = useState<number[]>([]);
  const [recommendations, setRecommendations] = useState<gui.CardRecommendation[]>([]);
  const [showRecommendations, setShowRecommendations] = useState(false);
  const [loadingRecommendations, setLoadingRecommendations] = useState(false);
  const [addingLands, setAddingLands] = useState(false);
  const [showSuggestDecks, setShowSuggestDecks] = useState(false);
  const [showBuildAround, setShowBuildAround] = useState(false);

  // Load deck data
  useEffect(() => {
    const loadDeck = async () => {
      setLoading(true);
      setError(null);

      try {
        let deckData;

        if (deckID) {
          // Load by deck ID
          deckData = await decks.getDeck(deckID);
        } else if (draftEventID) {
          // Load by draft event ID, create if doesn't exist
          try {
            deckData = await decks.getDeckByDraftEvent(draftEventID);
          } catch (fetchErr) {
            // 404 means no deck exists yet - we'll create one below
            // Any other error should be re-thrown
            if (fetchErr instanceof ApiRequestError && fetchErr.status === 404) {
              deckData = null;
            } else {
              throw fetchErr;
            }
          }

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
                drafts.getActiveDraftSessions(),
                drafts.getCompletedDraftSessions(),
              ]);
              const allSessions = [...activeSessions, ...completedSessions];
              const session = allSessions.find((s) => s.ID === draftEventID);

              if (!session) {
                setError('Draft session not found');
                setLoading(false);
                creatingDeckRef.current = false;
                return;
              }

              const deckName = `${session.EventName} Draft`;

              // Create deck linked to this draft event
              const newDeck = await decks.createDeck({
                name: deckName,
                format: 'limited',
                source: 'draft',
                draft_event_id: draftEventID,
              });

              // Load draft picks and add them to the new deck
              const picks = await drafts.getDraftPicks(draftEventID);
              if (picks && picks.length > 0) {
                // Count occurrences of each card (draft can have multiple copies)
                const cardCounts = new Map<number, number>();
                for (const pick of picks) {
                  const cardID = parseInt(pick.CardID, 10);
                  if (!isNaN(cardID)) {
                    cardCounts.set(cardID, (cardCounts.get(cardID) || 0) + 1);
                  }
                }

                // Add each card to the deck
                for (const [cardID, quantity] of cardCounts) {
                  try {
                    await decks.addCard({
                      deck_id: newDeck.ID,
                      arena_id: cardID,
                      quantity,
                      zone: 'main',
                      is_sideboard: false,
                      from_draft: true,
                    });
                  } catch (addErr) {
                    console.error(`Failed to add card ${cardID} to deck:`, addErr);
                  }
                }
              }

              // Load the newly created deck (now with cards)
              deckData = await decks.getDeck(newDeck.ID);
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

        // Load statistics (may not exist for new decks)
        try {
          const stats = await decks.getDeckStatistics(deckData.deck.ID);
          setStatistics(stats);
        } catch (statsErr) {
          // Statistics may not exist for new decks - this is OK
          console.log('No statistics for deck (may be new):', statsErr);
          setStatistics(null);
        }

        // If this is a draft deck, get the draft card IDs
        if (deckData.deck.Source === 'draft' && deckData.deck.DraftEventID) {
          try {
            const picks = await drafts.getDraftPicks(deckData.deck.DraftEventID);
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
      await decks.addCard({
        deck_id: deck.ID,
        arena_id: cardID,
        quantity,
        zone: board,
        is_sideboard: board === 'sideboard',
      });

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        // Statistics may not exist for new decks - this is OK
        setStatistics(null);
      }

      // Reload recommendations after adding a card
      if (deckData.cards && deckData.cards.length >= 3) {
        loadRecommendations();
      }
    } catch (err) {
      // Silently ignore 400 errors (validation errors like 4-copy limit)
      // but log other errors for debugging
      if (!(err instanceof ApiRequestError && err.status === 400)) {
        console.error('Failed to add card:', err);
      }
    }
  };

  const handleRemoveCard = async (cardID: number, board: string) => {
    if (!deck) return;

    try {
      await decks.removeCard({
        deck_id: deck.ID,
        arena_id: cardID,
        zone: board,
      });

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to remove card');
    }
  };

  const handleRemoveAllCopies = async (cardID: number, board: string) => {
    if (!deck) return;

    try {
      await decks.removeAllCopies({
        deck_id: deck.ID,
        arena_id: cardID,
        zone: board,
      });

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to remove all copies');
    }
  };

  const handleAddOneCard = async (cardID: number, board: string) => {
    if (!deck) return;

    try {
      await decks.addCard({
        deck_id: deck.ID,
        arena_id: cardID,
        quantity: 1,
        zone: board,
        is_sideboard: board === 'sideboard',
      });

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }
    } catch {
      // Silently ignore - card is likely at 4-copy limit
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

      const response = await drafts.getRecommendations(request);
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
      const colors = statistics.colors;
      console.log('Full statistics object:', statistics);
      console.log('Color distribution from backend:', colors);

      // Calculate color distribution from mainboard cards
      // Only count mono-colored cards for land distribution
      const colorCounts = {
        W: colors?.white || 0,
        U: colors?.blue || 0,
        B: colors?.black || 0,
        R: colors?.red || 0,
        G: colors?.green || 0,
      };

      console.log('Color counts (mono-colored only):', colorCounts);
      console.log('Color counts after assignment - W:', colorCounts.W, 'U:', colorCounts.U, 'B:', colorCounts.B, 'R:', colorCounts.R, 'G:', colorCounts.G);

      // Get backend's land recommendation
      const currentLands = statistics.lands?.total || 0;
      const recommendedLands = statistics.lands?.recommended || 0;
      console.log('Deck stats:', { currentLands, recommendedLands });

      if (recommendedLands === 0) {
        console.log('No land recommendation available');
        window.alert('Could not determine land recommendation. Please add more cards to your deck first.');
        return;
      }

      // Calculate how many more lands we need based on backend recommendation
      const landsNeeded = Math.max(0, recommendedLands - currentLands);
      console.log('Lands needed:', landsNeeded);

      if (landsNeeded === 0) {
        console.log('Deck already has enough lands');
        window.alert('Your deck already has the recommended number of lands!');
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
      const remaining = landsNeeded - landsAllocated;
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
          await decks.addCard({
            deck_id: deck.ID,
            arena_id: land.arenaID,
            quantity: count,
            zone: 'main',
            is_sideboard: false,
          });
        }
      }

      // Reload deck data
      console.log('Reloading deck data...');
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }

      console.log(`Successfully added ${landsNeeded} lands!`);
      window.alert(`Added ${landsNeeded} suggested lands to your deck!`);
    } catch (err) {
      console.error('Error adding lands:', err);
      window.alert(err instanceof Error ? err.message : 'Failed to add lands');
    } finally {
      setAddingLands(false);
    }
  };

  const handleExportDeck = async () => {
    if (!deck) {
      return;
    }

    try {
      // Call backend which will show native SaveFileDialog
      await exportDeckToFile(deck.ID);
    } catch (err) {
      console.error('Error exporting deck:', err);
    }
  };

  const handleValidateDeck = async () => {
    if (!deck) {
      return;
    }

    try {
      // Call backend which will show native MessageDialog with result
      await decks.validateDraftDeck(deck.ID);
    } catch (err) {
      console.error('Error validating deck:', err);
    }
  };

  const handleDeckApplied = async () => {
    if (!deck) return;

    // Reload deck data after a suggested deck is applied
    const deckData = await decks.getDeck(deck.ID);
    if (deckData.deck) {
      setDeck(deckData.deck);
    }
    setCards(deckData.cards || []);
    setTags(deckData.tags || []);

    // Reload statistics (may not exist for new/small decks)
    try {
      const stats = await decks.getDeckStatistics(deck.ID);
      setStatistics(stats);
    } catch {
      setStatistics(null);
    }
  };

  const handleApplyBuildAround = async (suggestions: CardWithOwnership[], lands: SuggestedLandResponse[]) => {
    if (!deck) return;

    try {
      // Clear existing cards first
      for (const card of cards) {
        await decks.removeCard({
          deck_id: deck.ID,
          arena_id: card.CardID,
          zone: card.Board,
        });
      }

      // Add suggested cards (4 copies each, up to 36 total cards)
      // 60 card deck - 24 lands = 36 spell slots, so max 9 unique cards × 4 copies = 36
      let cardsAdded = 0;
      for (const card of suggestions) {
        if (cardsAdded >= 36) break; // Leave room for lands
        const quantity = Math.min(4, 36 - cardsAdded);
        await decks.addCard({
          deck_id: deck.ID,
          arena_id: card.cardID,
          quantity,
          zone: 'main',
          is_sideboard: false,
        });
        cardsAdded += quantity;
      }

      // Add lands
      for (const land of lands) {
        await decks.addCard({
          deck_id: deck.ID,
          arena_id: land.cardID,
          quantity: land.quantity,
          zone: 'main',
          is_sideboard: false,
        });
      }

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }
    } catch (err) {
      console.error('Failed to apply build-around suggestions:', err);
      alert(err instanceof Error ? err.message : 'Failed to apply suggestions');
    }
  };

  // Handle adding a single card in iterative build-around mode
  const handleBuildAroundCardAdded = async (card: CardWithOwnership) => {
    if (!deck) return;

    try {
      // Check if deck is already at 60 cards (excluding lands for now, lands added at finish)
      const currentMainboard = statistics?.totalMainboard || 0;
      if (currentMainboard >= 60) {
        alert('Deck is already at 60 cards!');
        return;
      }

      // Add 1 copy at a time for granular control
      await decks.addCard({
        deck_id: deck.ID,
        arena_id: card.cardID,
        quantity: 1,
        zone: 'main',
        is_sideboard: false,
      });

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }
    } catch (err) {
      console.error('Failed to add card:', err);
      alert(err instanceof Error ? err.message : 'Failed to add card');
    }
  };

  // Handle removing a card in iterative build-around mode
  const handleBuildAroundCardRemoved = async (cardId: number) => {
    if (!deck) return;

    try {
      // Remove 1 copy at a time
      await decks.removeCard({
        deck_id: deck.ID,
        arena_id: cardId,
        zone: 'main',
      });

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }
    } catch (err) {
      console.error('Failed to remove card:', err);
      alert(err instanceof Error ? err.message : 'Failed to remove card');
    }
  };

  // Handle finishing deck in iterative build-around mode
  const handleBuildAroundFinishDeck = async (lands: SuggestedLandResponse[]) => {
    if (!deck) return;

    try {
      // Add suggested lands
      for (const land of lands) {
        await decks.addCard({
          deck_id: deck.ID,
          arena_id: land.cardID,
          quantity: land.quantity,
          zone: 'main',
          is_sideboard: false,
        });
      }

      // Reload deck data
      const deckData = await decks.getDeck(deck.ID);
      setCards(deckData.cards || []);

      // Reload statistics (may not exist for new/small decks)
      try {
        const stats = await decks.getDeckStatistics(deck.ID);
        setStatistics(stats);
      } catch {
        setStatistics(null);
      }

      setShowBuildAround(false);
    } catch (err) {
      console.error('Failed to finish deck:', err);
      alert(err instanceof Error ? err.message : 'Failed to add lands');
    }
  };

  // Get current deck card IDs for iterative build-around mode
  // Expand based on quantity - backend expects duplicate entries for multiple copies
  const currentDeckCardIDs = cards.flatMap((card) =>
    Array(card.Quantity).fill(card.CardID)
  );

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
            statistics={statistics ?? undefined}
            onAddCard={handleAddOneCard}
            onRemoveCard={handleRemoveCard}
            onRemoveAllCopies={handleRemoveAllCopies}
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
                  <RecommendationCard
                    key={rec.cardID}
                    recommendation={rec}
                    deckID={deck.ID}
                    onAddCard={handleAddCard}
                  />
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
          <button className="action-button" title="Export deck" onClick={handleExportDeck}>
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
          <button className="action-button" title="Validate deck" onClick={handleValidateDeck}>
            ✓ Validate
          </button>
          {deck.Source === 'draft' && deck.DraftEventID && (
            <button
              className="action-button suggest-decks-btn"
              title="Generate complete deck suggestions from your draft pool"
              onClick={() => setShowSuggestDecks(true)}
            >
              Suggest Decks
            </button>
          )}
          {deck.Source !== 'draft' && (
            <button
              className="action-button build-around-btn"
              title={cards.length === 0 ? 'Add cards to your deck first' : 'Get suggestions based on current deck cards'}
              disabled={cards.length === 0}
              onClick={() => setShowBuildAround(true)}
            >
              Build Around
            </button>
          )}
        </div>
      </div>

      {/* Suggest Decks Modal */}
      {deck.Source === 'draft' && deck.DraftEventID && (
        <SuggestDecksModal
          isOpen={showSuggestDecks}
          onClose={() => setShowSuggestDecks(false)}
          draftEventID={deck.DraftEventID}
          currentDeckID={deck.ID}
          deckName={deck.Name}
          onDeckApplied={handleDeckApplied}
        />
      )}

      {/* Build Around Seed Modal */}
      <BuildAroundSeedModal
        isOpen={showBuildAround}
        onClose={() => setShowBuildAround(false)}
        onApplyDeck={handleApplyBuildAround}
        onCardAdded={handleBuildAroundCardAdded}
        onCardRemoved={handleBuildAroundCardRemoved}
        onFinishDeck={handleBuildAroundFinishDeck}
        currentDeckCards={currentDeckCardIDs}
        deckCards={cards}
        useDeckCardsAsSeed={cards.length > 0}
      />
    </div>
  );
}

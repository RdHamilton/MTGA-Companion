import { useState, useEffect } from 'react';
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
} from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import DeckList from '../components/DeckList';
import CardSearch from '../components/CardSearch';
import './DeckBuilder.css';

export default function DeckBuilder() {
  const { deckID } = useParams<{ deckID?: string }>();
  const { draftEventID } = useParams<{ draftEventID?: string }>();
  const navigate = useNavigate();

  const [deck, setDeck] = useState<models.Deck | null>(null);
  const [cards, setCards] = useState<models.DeckCard[]>([]);
  const [tags, setTags] = useState<models.DeckTag[]>([]);
  const [statistics, setStatistics] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCardSearch, setShowCardSearch] = useState(false);
  const [draftCardIDs, setDraftCardIDs] = useState<number[]>([]);

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
            try {
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
              return;
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
          <button className="action-button" title="Get recommendations">
            ✨ Suggestions
          </button>
          <button className="action-button" title="Validate deck">
            ✓ Validate
          </button>
        </div>
      </div>
    </div>
  );
}

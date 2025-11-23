import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ListDecks, CreateDeck } from '../../wailsjs/go/main/App';
import './Decks.css';

interface DeckListItem {
  ID: string;
  Name: string;
  Format: string;
  Source: string;
  CreatedAt: string;
  LastModified?: string;
  CardCount: number;
  DraftEventID?: string;
}

export default function Decks() {
  const navigate = useNavigate();
  const [decks, setDecks] = useState<DeckListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newDeckName, setNewDeckName] = useState('');
  const [newDeckFormat, setNewDeckFormat] = useState('standard');

  useEffect(() => {
    loadDecks();
  }, []);

  const loadDecks = async () => {
    setLoading(true);
    setError(null);
    try {
      const deckList = await ListDecks();
      setDecks(deckList || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load decks');
      console.error('Failed to load decks:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateDeck = async () => {
    if (!newDeckName.trim()) {
      alert('Please enter a deck name');
      return;
    }

    try {
      const deck = await CreateDeck(newDeckName.trim(), newDeckFormat, 'manual', null);
      setShowCreateDialog(false);
      setNewDeckName('');
      navigate(`/deck-builder/${deck.ID}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create deck');
    }
  };

  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'N/A';
    return new Date(dateStr).toLocaleDateString();
  };

  if (loading) {
    return (
      <div className="decks-page loading-state">
        <div className="loading-spinner"></div>
        <p>Loading decks...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="decks-page error-state">
        <div className="error-icon">⚠️</div>
        <h2>Error Loading Decks</h2>
        <p>{error}</p>
        <button onClick={loadDecks} className="retry-button">
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="decks-page">
      {/* Header */}
      <div className="decks-header">
        <h1>My Decks</h1>
        <button className="create-deck-button" onClick={() => setShowCreateDialog(true)}>
          + Create New Deck
        </button>
      </div>

      {/* Decks Grid */}
      {decks.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📦</div>
          <h2>No Decks Yet</h2>
          <p>Create your first deck to get started!</p>
          <button className="create-deck-button-large" onClick={() => setShowCreateDialog(true)}>
            + Create New Deck
          </button>
        </div>
      ) : (
        <div className="decks-grid">
          {decks.map((deck) => (
            <div
              key={deck.ID}
              className="deck-card"
              onClick={() => navigate(`/deck-builder/${deck.ID}`)}
            >
              <div className="deck-card-header">
                <h3>{deck.Name}</h3>
                {deck.Source === 'draft' && (
                  <span className="source-badge draft">Draft</span>
                )}
                {deck.Source === 'import' && (
                  <span className="source-badge import">Import</span>
                )}
              </div>
              <div className="deck-card-body">
                <div className="deck-info">
                  <span className="deck-format">{deck.Format}</span>
                  <span className="deck-date">Created: {formatDate(deck.CreatedAt)}</span>
                  {deck.LastModified && (
                    <span className="deck-date">Updated: {formatDate(deck.LastModified)}</span>
                  )}
                </div>
              </div>
              <div className="deck-card-footer">
                <button
                  className="edit-button"
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(`/deck-builder/${deck.ID}`);
                  }}
                >
                  ✏️ Edit
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Deck Dialog */}
      {showCreateDialog && (
        <div className="modal-overlay" onClick={() => setShowCreateDialog(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Create New Deck</h2>
              <button className="close-button" onClick={() => setShowCreateDialog(false)}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label htmlFor="deck-name">Deck Name</label>
                <input
                  id="deck-name"
                  type="text"
                  value={newDeckName}
                  onChange={(e) => setNewDeckName(e.target.value)}
                  placeholder="My Awesome Deck"
                  autoFocus
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      handleCreateDeck();
                    }
                  }}
                />
              </div>
              <div className="form-group">
                <label htmlFor="deck-format">Format</label>
                <select
                  id="deck-format"
                  value={newDeckFormat}
                  onChange={(e) => setNewDeckFormat(e.target.value)}
                >
                  <option value="standard">Standard</option>
                  <option value="alchemy">Alchemy</option>
                  <option value="explorer">Explorer</option>
                  <option value="historic">Historic</option>
                  <option value="timeless">Timeless</option>
                  <option value="brawl">Brawl</option>
                  <option value="limited">Limited</option>
                </select>
              </div>
            </div>
            <div className="modal-footer">
              <button className="cancel-button" onClick={() => setShowCreateDialog(false)}>
                Cancel
              </button>
              <button className="create-button" onClick={handleCreateDeck}>
                Create Deck
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

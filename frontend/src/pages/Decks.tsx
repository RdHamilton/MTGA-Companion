import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ListDecks, CreateDeck, DeleteDeck } from '../../wailsjs/go/main/App';
import { gui } from '../../wailsjs/go/models';
import './Decks.css';

export default function Decks() {
  const navigate = useNavigate();
  const [decks, setDecks] = useState<gui.DeckListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newDeckName, setNewDeckName] = useState('');
  const [newDeckFormat, setNewDeckFormat] = useState('standard');
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [deckToDelete, setDeckToDelete] = useState<gui.DeckListItem | null>(null);

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

  useEffect(() => {
    // Wait for Wails runtime to be ready before loading decks
    const checkWailsReady = setInterval(() => {
      if (typeof window !== 'undefined' && (window as any).go) {
        clearInterval(checkWailsReady);
        loadDecks();
      }
    }, 100);

    // Fallback timeout after 5 seconds
    const timeout = setTimeout(() => {
      clearInterval(checkWailsReady);
      if (!(window as any).go) {
        setError('Wails runtime not initialized');
        setLoading(false);
      }
    }, 5000);

    return () => {
      clearInterval(checkWailsReady);
      clearTimeout(timeout);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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

  const handleDeleteClick = (deck: gui.DeckListItem, e: React.MouseEvent) => {
    e.stopPropagation();
    setDeckToDelete(deck);
    setShowDeleteDialog(true);
  };

  const handleDeleteConfirm = async () => {
    if (!deckToDelete) return;

    try {
      await DeleteDeck(deckToDelete.id);
      setShowDeleteDialog(false);
      setDeckToDelete(null);
      await loadDecks();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete deck');
    }
  };

  const handleDeleteCancel = () => {
    setShowDeleteDialog(false);
    setDeckToDelete(null);
  };

  const formatDate = (date: any) => {
    if (!date) return 'N/A';
    return new Date(date).toLocaleDateString();
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
      {/* Header - Only show button when there are decks */}
      <div className="decks-header">
        <h1>My Decks</h1>
        {decks.length > 0 && (
          <button className="create-deck-button" onClick={() => setShowCreateDialog(true)}>
            + Create New Deck
          </button>
        )}
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
              key={deck.id}
              className="deck-card"
              onClick={() => navigate(`/deck-builder/${deck.id}`)}
            >
              <div className="deck-card-header">
                <h3>{deck.name}</h3>
                {deck.source === 'draft' && (
                  <span className="source-badge draft">Draft</span>
                )}
                {deck.source === 'import' && (
                  <span className="source-badge import">Import</span>
                )}
              </div>
              <div className="deck-card-body">
                <div className="deck-info">
                  <span className="deck-format">{deck.format}</span>
                  {deck.modifiedAt && (
                    <span className="deck-date">Modified: {formatDate(deck.modifiedAt)}</span>
                  )}
                </div>
              </div>
              <div className="deck-card-footer">
                <button
                  className="edit-button"
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(`/deck-builder/${deck.id}`);
                  }}
                >
                  Edit
                </button>
                <button
                  className="delete-button"
                  onClick={(e) => handleDeleteClick(deck, e)}
                >
                  Delete
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

      {/* Delete Confirmation Dialog */}
      {showDeleteDialog && deckToDelete && (
        <div className="modal-overlay" onClick={handleDeleteCancel}>
          <div className="modal-content delete-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Delete Deck</h2>
              <button className="close-button" onClick={handleDeleteCancel}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <p>Are you sure you want to delete <strong>{deckToDelete.name}</strong>?</p>
              <p className="warning-text">This action cannot be undone.</p>
            </div>
            <div className="modal-footer">
              <button className="cancel-button" onClick={handleDeleteCancel}>
                Cancel
              </button>
              <button className="delete-button-confirm" onClick={handleDeleteConfirm}>
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

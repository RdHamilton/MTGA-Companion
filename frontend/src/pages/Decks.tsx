import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { decks as decksApi } from '@/services/api';
import { gui } from '@/types/models';
import { useRotationNotifications } from '@/hooks/useRotationNotifications';
import { useSettings } from '@/hooks/useSettings';
import { RotationBanner } from '@/components/RotationBanner';
import './Decks.css';

type ExportFormat = 'arena' | 'moxfield' | 'archidekt' | 'mtgo' | 'mtggoldfish' | 'plaintext';

export default function Decks() {
  const navigate = useNavigate();
  const [deckList, setDeckList] = useState<gui.DeckListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newDeckName, setNewDeckName] = useState('');
  const [newDeckFormat, setNewDeckFormat] = useState('standard');
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [deckToDelete, setDeckToDelete] = useState<gui.DeckListItem | null>(null);
  const [showExportDialog, setShowExportDialog] = useState(false);
  const [deckToExport, setDeckToExport] = useState<gui.DeckListItem | null>(null);
  const [exportFormat, setExportFormat] = useState<ExportFormat>('arena');
  const [isExporting, setIsExporting] = useState(false);

  // Rotation notifications
  const {
    rotation,
    affectedDecks,
    shouldShowNotification,
    markAsNotified,
  } = useRotationNotifications();
  const { rotationNotificationsEnabled, rotationNotificationThreshold } = useSettings();

  // Determine if we should show the rotation banner
  const showRotationBanner =
    rotationNotificationsEnabled &&
    rotation &&
    affectedDecks.length > 0 &&
    shouldShowNotification(rotationNotificationThreshold);

  const loadDecks = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await decksApi.getDecks();
      setDeckList(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load decks');
      console.error('Failed to load decks:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDecks();
  }, []);

  const handleCreateDeck = async () => {
    if (!newDeckName.trim()) {
      alert('Please enter a deck name');
      return;
    }

    try {
      const deck = await decksApi.createDeck({
        name: newDeckName.trim(),
        format: newDeckFormat,
        source: 'manual',
      });
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
      await decksApi.deleteDeck(deckToDelete.id);
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

  const handleExportClick = (deck: gui.DeckListItem, e: React.MouseEvent) => {
    e.stopPropagation();
    setDeckToExport(deck);
    setShowExportDialog(true);
  };

  const handleExportConfirm = useCallback(async () => {
    if (!deckToExport) return;

    try {
      setIsExporting(true);
      const response = await decksApi.exportDeck(deckToExport.id, { format: exportFormat });

      if (response.Error) {
        alert(`Export failed: ${response.Error}`);
        return;
      }

      // Create blob and trigger download
      const blob = new Blob([response.Content], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = response.Filename || `${deckToExport.name}.txt`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);

      setShowExportDialog(false);
      setDeckToExport(null);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to export deck');
    } finally {
      setIsExporting(false);
    }
  }, [deckToExport, exportFormat]);

  const handleExportCancel = () => {
    setShowExportDialog(false);
    setDeckToExport(null);
  };

  const handleCopyToClipboard = useCallback(async () => {
    if (!deckToExport) return;

    try {
      setIsExporting(true);
      const response = await decksApi.exportDeck(deckToExport.id, { format: exportFormat });

      if (response.Error) {
        alert(`Export failed: ${response.Error}`);
        return;
      }

      await navigator.clipboard.writeText(response.Content);
      alert('Deck copied to clipboard!');
      setShowExportDialog(false);
      setDeckToExport(null);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to copy to clipboard');
    } finally {
      setIsExporting(false);
    }
  }, [deckToExport, exportFormat]);

  const formatDate = (date: unknown) => {
    if (!date) return 'N/A';
    return new Date(String(date)).toLocaleDateString();
  };

  const formatStreak = (streak: number) => {
    if (streak === 0) return null;
    if (streak > 0) {
      return { text: `${streak}W`, className: 'win-streak', icon: '🔥' };
    }
    return { text: `${Math.abs(streak)}L`, className: 'loss-streak', icon: '❄️' };
  };

  const formatDuration = (seconds: number | undefined) => {
    if (!seconds) return null;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
      return `~${minutes}m avg`;
    }
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `~${hours}h ${mins}m avg`;
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
        {deckList.length > 0 && (
          <button className="create-deck-button" onClick={() => setShowCreateDialog(true)}>
            + Create New Deck
          </button>
        )}
      </div>

      {/* Rotation Banner */}
      {showRotationBanner && rotation && (
        <RotationBanner
          rotation={rotation}
          affectedDecks={affectedDecks}
          onDismiss={markAsNotified}
        />
      )}

      {/* Decks Grid */}
      {deckList.length === 0 ? (
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
          {deckList.map((deck) => (
            <div
              key={deck.id}
              className="deck-card"
              onClick={() => navigate(`/deck-builder/${deck.id}`)}
            >
              <div className="deck-card-header">
                <h3>{deck.name}</h3>
                <div className="deck-badges">
                  {deck.primaryArchetype && (
                    <span className="archetype-badge">{deck.primaryArchetype}</span>
                  )}
                  {deck.source === 'draft' && (
                    <span className="source-badge draft">Draft</span>
                  )}
                  {deck.source === 'import' && (
                    <span className="source-badge import">Import</span>
                  )}
                </div>
              </div>
              <div className="deck-card-body">
                <div className="deck-info">
                  <span className="deck-format">{deck.format}</span>
                  {deck.modifiedAt && (
                    <span className="deck-date">Modified: {formatDate(deck.modifiedAt)}</span>
                  )}
                </div>
                {deck.matchesPlayed > 0 && (
                  <div className="deck-stats-row">
                    <span className="deck-win-rate">
                      {Math.round(deck.matchWinRate * 100)}% WR ({deck.matchesPlayed} matches)
                    </span>
                    {formatStreak(deck.currentStreak) && (
                      <span className={`deck-streak ${formatStreak(deck.currentStreak)?.className}`}>
                        {formatStreak(deck.currentStreak)?.icon} {formatStreak(deck.currentStreak)?.text}
                      </span>
                    )}
                    {formatDuration(deck.averageDuration) && (
                      <span className="deck-duration">{formatDuration(deck.averageDuration)}</span>
                    )}
                  </div>
                )}
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
                  className="export-button"
                  onClick={(e) => handleExportClick(deck, e)}
                >
                  Export
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

      {/* Export Deck Dialog */}
      {showExportDialog && deckToExport && (
        <div className="modal-overlay" onClick={handleExportCancel}>
          <div className="modal-content export-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Export Deck</h2>
              <button className="close-button" onClick={handleExportCancel}>
                ×
              </button>
            </div>
            <div className="modal-body">
              <p>Export <strong>{deckToExport.name}</strong></p>
              <div className="form-group">
                <label htmlFor="export-format">Export Format</label>
                <select
                  id="export-format"
                  value={exportFormat}
                  onChange={(e) => setExportFormat(e.target.value as ExportFormat)}
                  disabled={isExporting}
                >
                  <option value="arena">MTG Arena</option>
                  <option value="moxfield">Moxfield</option>
                  <option value="archidekt">Archidekt</option>
                  <option value="mtgo">MTGO</option>
                  <option value="mtggoldfish">MTGGoldfish</option>
                  <option value="plaintext">Plain Text</option>
                </select>
              </div>
              <p className="export-hint">
                {exportFormat === 'arena' && 'Standard MTGA import format with set codes'}
                {exportFormat === 'moxfield' && 'Import directly into Moxfield'}
                {exportFormat === 'archidekt' && 'Import directly into Archidekt'}
                {exportFormat === 'mtgo' && 'MTGO deck format (.dek)'}
                {exportFormat === 'mtggoldfish' && 'MTGGoldfish import format'}
                {exportFormat === 'plaintext' && 'Simple text list (4x Card Name)'}
              </p>
            </div>
            <div className="modal-footer export-footer">
              <button className="cancel-button" onClick={handleExportCancel} disabled={isExporting}>
                Cancel
              </button>
              <button
                className="copy-button"
                onClick={handleCopyToClipboard}
                disabled={isExporting}
              >
                {isExporting ? 'Copying...' : 'Copy to Clipboard'}
              </button>
              <button
                className="export-button-confirm"
                onClick={handleExportConfirm}
                disabled={isExporting}
              >
                {isExporting ? 'Exporting...' : 'Download File'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

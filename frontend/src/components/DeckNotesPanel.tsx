import { useState, useEffect, useCallback } from 'react';
import { notes as notesApi } from '@/services/api';
import type { DeckNote, NoteCategory } from '@/services/api/notes';
import './DeckNotesPanel.css';

interface DeckNotesPanelProps {
  deckId: string;
  onClose?: () => void;
}

const CATEGORY_OPTIONS: { value: NoteCategory; label: string }[] = [
  { value: 'general', label: 'General' },
  { value: 'matchup', label: 'Matchup' },
  { value: 'sideboard', label: 'Sideboard' },
  { value: 'mulligan', label: 'Mulligan' },
];

export default function DeckNotesPanel({ deckId, onClose }: DeckNotesPanelProps) {
  const [notesList, setNotesList] = useState<DeckNote[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterCategory, setFilterCategory] = useState<NoteCategory | 'all'>('all');

  // New note form state
  const [newNoteContent, setNewNoteContent] = useState('');
  const [newNoteCategory, setNewNoteCategory] = useState<NoteCategory>('general');
  const [isAddingNote, setIsAddingNote] = useState(false);

  // Edit state
  const [editingNoteId, setEditingNoteId] = useState<number | null>(null);
  const [editContent, setEditContent] = useState('');
  const [editCategory, setEditCategory] = useState<NoteCategory>('general');

  const loadNotes = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const category = filterCategory === 'all' ? undefined : filterCategory;
      const data = await notesApi.getDeckNotes(deckId, category);
      setNotesList(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load notes');
      console.error('Failed to load notes:', err);
    } finally {
      setLoading(false);
    }
  }, [deckId, filterCategory]);

  useEffect(() => {
    loadNotes();
  }, [loadNotes]);

  const handleAddNote = async () => {
    if (!newNoteContent.trim()) return;

    try {
      const newNote = await notesApi.createDeckNote(deckId, {
        content: newNoteContent.trim(),
        category: newNoteCategory,
      });
      setNotesList((prev) => [newNote, ...prev]);
      setNewNoteContent('');
      setIsAddingNote(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add note');
    }
  };

  const handleUpdateNote = async (noteId: number) => {
    if (!editContent.trim()) return;

    try {
      const updatedNote = await notesApi.updateDeckNote(deckId, noteId, {
        content: editContent.trim(),
        category: editCategory,
      });
      setNotesList((prev) =>
        prev.map((note) => (note.id === noteId ? updatedNote : note))
      );
      setEditingNoteId(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update note');
    }
  };

  const handleDeleteNote = async (noteId: number) => {
    try {
      await notesApi.deleteDeckNote(deckId, noteId);
      setNotesList((prev) => prev.filter((note) => note.id !== noteId));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete note');
    }
  };

  const startEditing = (note: DeckNote) => {
    setEditingNoteId(note.id);
    setEditContent(note.content);
    setEditCategory(note.category);
  };

  const cancelEditing = () => {
    setEditingNoteId(null);
    setEditContent('');
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getCategoryColor = (category: NoteCategory): string => {
    const colors: Record<NoteCategory, string> = {
      general: '#6b7280',
      matchup: '#3b82f6',
      sideboard: '#10b981',
      mulligan: '#f59e0b',
    };
    return colors[category] || '#6b7280';
  };

  if (loading) {
    return (
      <div className="deck-notes-panel loading">
        <div className="loading-spinner"></div>
        <p>Loading notes...</p>
      </div>
    );
  }

  return (
    <div className="deck-notes-panel">
      <div className="deck-notes-header">
        <h3>Deck Notes</h3>
        <div className="header-controls">
          <select
            value={filterCategory}
            onChange={(e) => setFilterCategory(e.target.value as NoteCategory | 'all')}
            className="category-filter"
          >
            <option value="all">All Categories</option>
            {CATEGORY_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          {onClose && (
            <button className="close-button" onClick={onClose} title="Close">
              x
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="error-banner">
          <span>{error}</span>
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      {/* Add Note Form */}
      {isAddingNote ? (
        <div className="add-note-form">
          <textarea
            value={newNoteContent}
            onChange={(e) => setNewNoteContent(e.target.value)}
            placeholder="Write your note here..."
            rows={3}
            autoFocus
          />
          <div className="form-controls">
            <select
              value={newNoteCategory}
              onChange={(e) => setNewNoteCategory(e.target.value as NoteCategory)}
            >
              {CATEGORY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <div className="button-group">
              <button className="cancel-btn" onClick={() => setIsAddingNote(false)}>
                Cancel
              </button>
              <button
                className="save-btn"
                onClick={handleAddNote}
                disabled={!newNoteContent.trim()}
              >
                Save Note
              </button>
            </div>
          </div>
        </div>
      ) : (
        <button className="add-note-btn" onClick={() => setIsAddingNote(true)}>
          + Add Note
        </button>
      )}

      {/* Notes List */}
      <div className="notes-list">
        {notesList.length === 0 ? (
          <div className="empty-state">
            <p>No notes yet.</p>
            <p>Add notes to track matchup strategies, sideboard plans, and more!</p>
          </div>
        ) : (
          notesList.map((note) => (
            <div key={note.id} className="note-item">
              {editingNoteId === note.id ? (
                <div className="edit-note-form">
                  <textarea
                    value={editContent}
                    onChange={(e) => setEditContent(e.target.value)}
                    rows={3}
                    autoFocus
                  />
                  <div className="form-controls">
                    <select
                      value={editCategory}
                      onChange={(e) => setEditCategory(e.target.value as NoteCategory)}
                    >
                      {CATEGORY_OPTIONS.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                    <div className="button-group">
                      <button className="cancel-btn" onClick={cancelEditing}>
                        Cancel
                      </button>
                      <button
                        className="save-btn"
                        onClick={() => handleUpdateNote(note.id)}
                        disabled={!editContent.trim()}
                      >
                        Save
                      </button>
                    </div>
                  </div>
                </div>
              ) : (
                <>
                  <div className="note-header">
                    <span
                      className="note-category"
                      style={{ backgroundColor: getCategoryColor(note.category) }}
                    >
                      {note.category}
                    </span>
                    <span className="note-date">{formatDate(note.createdAt)}</span>
                  </div>
                  <p className="note-content">{note.content}</p>
                  <div className="note-actions">
                    <button
                      className="edit-btn"
                      onClick={() => startEditing(note)}
                      title="Edit note"
                    >
                      Edit
                    </button>
                    <button
                      className="delete-btn"
                      onClick={() => handleDeleteNote(note.id)}
                      title="Delete note"
                    >
                      Delete
                    </button>
                  </div>
                </>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}

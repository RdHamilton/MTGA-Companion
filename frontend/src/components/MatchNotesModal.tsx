import { useState, useEffect } from 'react';
import { notes as notesApi } from '@/services/api';
import type { MatchNotes } from '@/services/api/notes';
import './MatchNotesModal.css';

interface MatchNotesModalProps {
  matchId: string;
  isOpen: boolean;
  onClose: () => void;
  onSave?: (notes: MatchNotes) => void;
}

const QUICK_TAGS = [
  { label: 'Misplay', value: 'Made a misplay - ' },
  { label: 'Mana Issues', value: 'Had mana issues - ' },
  { label: 'Great Game', value: 'Great game! ' },
  { label: 'Close Match', value: 'Close match - ' },
  { label: 'Bad Draws', value: 'Bad draws - ' },
  { label: 'Opponent Error', value: 'Opponent misplayed - ' },
];

export default function MatchNotesModal({
  matchId,
  isOpen,
  onClose,
  onSave,
}: MatchNotesModalProps) {
  const [notesContent, setNotesContent] = useState('');
  const [rating, setRating] = useState(0);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    const loadNotes = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await notesApi.getMatchNotes(matchId);
        setNotesContent(data.notes || '');
        setRating(data.rating || 0);
      } catch {
        // If no notes exist yet, that's fine - start with empty
        setNotesContent('');
        setRating(0);
      } finally {
        setLoading(false);
      }
    };

    loadNotes();
  }, [matchId, isOpen]);

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      const updated = await notesApi.updateMatchNotes(matchId, {
        notes: notesContent,
        rating,
      });
      onSave?.(updated);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save notes');
    } finally {
      setSaving(false);
    }
  };

  const handleQuickTag = (tag: string) => {
    setNotesContent((prev) => {
      if (prev) {
        return prev + '\n' + tag;
      }
      return tag;
    });
  };

  const handleRatingClick = (value: number) => {
    // Toggle off if clicking the same rating
    setRating((prev) => (prev === value ? 0 : value));
  };

  if (!isOpen) return null;

  return (
    <div className="match-notes-modal-overlay" onClick={onClose}>
      <div className="match-notes-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Match Notes</h3>
          <button className="close-button" onClick={onClose}>
            x
          </button>
        </div>

        {loading ? (
          <div className="modal-loading">
            <div className="loading-spinner"></div>
            <p>Loading...</p>
          </div>
        ) : (
          <div className="modal-content">
            {error && (
              <div className="error-message">
                <span>{error}</span>
              </div>
            )}

            {/* Rating */}
            <div className="rating-section">
              <label>Self-Rating:</label>
              <div className="star-rating">
                {[1, 2, 3, 4, 5].map((star) => (
                  <button
                    key={star}
                    className={`star ${star <= rating ? 'active' : ''}`}
                    onClick={() => handleRatingClick(star)}
                    title={`${star} star${star > 1 ? 's' : ''}`}
                  >
                    ★
                  </button>
                ))}
              </div>
              <span className="rating-hint">
                {rating === 0 && 'Rate your performance'}
                {rating === 1 && 'Poor - Many mistakes'}
                {rating === 2 && 'Below Average'}
                {rating === 3 && 'Average'}
                {rating === 4 && 'Good'}
                {rating === 5 && 'Excellent - Played perfectly'}
              </span>
            </div>

            {/* Quick Tags */}
            <div className="quick-tags-section">
              <label>Quick Tags:</label>
              <div className="quick-tags">
                {QUICK_TAGS.map((tag) => (
                  <button
                    key={tag.label}
                    className="quick-tag"
                    onClick={() => handleQuickTag(tag.value)}
                  >
                    {tag.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Notes Textarea */}
            <div className="notes-section">
              <label>Notes:</label>
              <textarea
                value={notesContent}
                onChange={(e) => setNotesContent(e.target.value)}
                placeholder="Add notes about this match... What went well? What could you improve?"
                rows={6}
              />
            </div>

            {/* Actions */}
            <div className="modal-actions">
              <button className="cancel-btn" onClick={onClose} disabled={saving}>
                Cancel
              </button>
              <button className="save-btn" onClick={handleSave} disabled={saving}>
                {saving ? 'Saving...' : 'Save Notes'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

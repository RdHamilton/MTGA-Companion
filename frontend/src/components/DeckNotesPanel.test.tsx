import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { mockNotes } from '@/test/mocks/apiMock';
import type { DeckNote } from '@/services/api/notes';
import DeckNotesPanel from './DeckNotesPanel';

// Mock the API module
vi.mock('@/services/api', () => ({
  notes: mockNotes,
}));

// Helper to create mock notes
function createMockNote(overrides: Partial<DeckNote> = {}): DeckNote {
  return {
    id: 1,
    deckId: 'deck-1',
    content: 'Test note content',
    category: 'general',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
    ...overrides,
  };
}

describe('DeckNotesPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching notes', async () => {
      let resolvePromise: (value: DeckNote[]) => void;
      const loadingPromise = new Promise<DeckNote[]>((resolve) => {
        resolvePromise = resolve;
      });
      mockNotes.getDeckNotes.mockReturnValue(loadingPromise);

      render(<DeckNotesPanel deckId="deck-1" />);

      expect(screen.getByText('Loading notes...')).toBeInTheDocument();

      resolvePromise!([createMockNote()]);
      await waitFor(() => {
        expect(screen.queryByText('Loading notes...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no notes exist', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('No notes yet.')).toBeInTheDocument();
      });
    });
  });

  describe('Notes List', () => {
    it('should display notes when loaded', async () => {
      const notes = [
        createMockNote({ id: 1, content: 'First note' }),
        createMockNote({ id: 2, content: 'Second note' }),
      ];
      mockNotes.getDeckNotes.mockResolvedValue(notes);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('First note')).toBeInTheDocument();
        expect(screen.getByText('Second note')).toBeInTheDocument();
      });
    });

    it('should display note category badge', async () => {
      const notes = [createMockNote({ category: 'matchup' })];
      mockNotes.getDeckNotes.mockResolvedValue(notes);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('matchup')).toBeInTheDocument();
      });
    });
  });

  describe('Add Note', () => {
    it('should show add note form when button clicked', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('+ Add Note')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Add Note'));

      expect(screen.getByPlaceholderText('Write your note here...')).toBeInTheDocument();
      expect(screen.getByText('Save Note')).toBeInTheDocument();
    });

    it('should create note when form submitted', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);
      mockNotes.createDeckNote.mockResolvedValue(createMockNote({ content: 'New note' }));

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('+ Add Note'));
      });

      const textarea = screen.getByPlaceholderText('Write your note here...');
      fireEvent.change(textarea, { target: { value: 'New note content' } });

      fireEvent.click(screen.getByText('Save Note'));

      await waitFor(() => {
        expect(mockNotes.createDeckNote).toHaveBeenCalledWith('deck-1', {
          content: 'New note content',
          category: 'general',
        });
      });
    });

    it('should cancel add note form', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('+ Add Note'));
      });

      expect(screen.getByText('Cancel')).toBeInTheDocument();
      fireEvent.click(screen.getByText('Cancel'));

      expect(screen.queryByPlaceholderText('Write your note here...')).not.toBeInTheDocument();
    });
  });

  describe('Edit Note', () => {
    it('should show edit form when edit button clicked', async () => {
      const notes = [createMockNote({ content: 'Original content' })];
      mockNotes.getDeckNotes.mockResolvedValue(notes);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('Original content')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByTitle('Edit note'));

      const textarea = screen.getByDisplayValue('Original content');
      expect(textarea).toBeInTheDocument();
    });

    it('should update note when edit saved', async () => {
      const notes = [createMockNote({ id: 1, content: 'Original' })];
      mockNotes.getDeckNotes.mockResolvedValue(notes);
      mockNotes.updateDeckNote.mockResolvedValue(createMockNote({ content: 'Updated' }));

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        fireEvent.click(screen.getByTitle('Edit note'));
      });

      const textarea = screen.getByDisplayValue('Original');
      fireEvent.change(textarea, { target: { value: 'Updated content' } });

      fireEvent.click(screen.getByText('Save'));

      await waitFor(() => {
        expect(mockNotes.updateDeckNote).toHaveBeenCalledWith('deck-1', 1, {
          content: 'Updated content',
          category: 'general',
        });
      });
    });
  });

  describe('Delete Note', () => {
    it('should delete note when delete button clicked', async () => {
      const notes = [createMockNote({ id: 1, content: 'To delete' })];
      mockNotes.getDeckNotes.mockResolvedValue(notes);
      mockNotes.deleteDeckNote.mockResolvedValue(undefined);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByText('To delete')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByTitle('Delete note'));

      await waitFor(() => {
        expect(mockNotes.deleteDeckNote).toHaveBeenCalledWith('deck-1', 1);
      });
    });
  });

  describe('Category Filter', () => {
    it('should filter notes by category', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.getByRole('combobox')).toBeInTheDocument();
      });

      fireEvent.change(screen.getByRole('combobox'), { target: { value: 'matchup' } });

      await waitFor(() => {
        expect(mockNotes.getDeckNotes).toHaveBeenCalledWith('deck-1', 'matchup');
      });
    });
  });

  describe('Close Button', () => {
    it('should call onClose when close button clicked', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);
      const onClose = vi.fn();

      render(<DeckNotesPanel deckId="deck-1" onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByTitle('Close')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByTitle('Close'));

      expect(onClose).toHaveBeenCalled();
    });

    it('should not show close button when onClose not provided', async () => {
      mockNotes.getDeckNotes.mockResolvedValue([]);

      render(<DeckNotesPanel deckId="deck-1" />);

      await waitFor(() => {
        expect(screen.queryByTitle('Close')).not.toBeInTheDocument();
      });
    });
  });
});

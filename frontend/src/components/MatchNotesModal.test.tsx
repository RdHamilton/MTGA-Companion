import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { mockNotes } from '@/test/mocks/apiMock';
import type { MatchNotes } from '@/services/api/notes';
import MatchNotesModal from './MatchNotesModal';

// Mock the API module
vi.mock('@/services/api', () => ({
  notes: mockNotes,
}));

describe('MatchNotesModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Closed State', () => {
    it('should not render when isOpen is false', () => {
      render(<MatchNotesModal matchId="match-1" isOpen={false} onClose={vi.fn()} />);

      expect(screen.queryByText('Match Notes')).not.toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching notes', async () => {
      let resolvePromise: (value: MatchNotes) => void;
      const loadingPromise = new Promise<MatchNotes>((resolve) => {
        resolvePromise = resolve;
      });
      mockNotes.getMatchNotes.mockReturnValue(loadingPromise);

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      expect(screen.getByText('Loading...')).toBeInTheDocument();

      resolvePromise!({ matchId: 'match-1', notes: '', rating: 0 });
      await waitFor(() => {
        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Display Notes', () => {
    it('should display existing notes when loaded', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: 'Existing notes content',
        rating: 3,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        const textarea = screen.getByPlaceholderText(/Add notes about this match/);
        expect(textarea).toHaveValue('Existing notes content');
      });
    });

    it('should display existing rating when loaded', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 4,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        // Stars have title attributes like "1 star", "2 stars", etc.
        const star1 = screen.getByTitle('1 star');
        const star2 = screen.getByTitle('2 stars');
        const star3 = screen.getByTitle('3 stars');
        const star4 = screen.getByTitle('4 stars');
        const star5 = screen.getByTitle('5 stars');
        // 4 stars should be active
        expect(star1).toHaveClass('active');
        expect(star2).toHaveClass('active');
        expect(star3).toHaveClass('active');
        expect(star4).toHaveClass('active');
        expect(star5).not.toHaveClass('active');
      });
    });

    it('should show rating hint based on selected rating', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 5,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Excellent - Played perfectly')).toBeInTheDocument();
      });
    });
  });

  describe('Rating Interaction', () => {
    it('should update rating when star clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Rate your performance')).toBeInTheDocument();
      });

      // Click 3rd star using title
      fireEvent.click(screen.getByTitle('3 stars'));

      await waitFor(() => {
        expect(screen.getByText('Average')).toBeInTheDocument();
      });
    });

    it('should toggle rating off when same star clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 3,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Average')).toBeInTheDocument();
      });

      // Click the same 3rd star again using title
      fireEvent.click(screen.getByTitle('3 stars'));

      await waitFor(() => {
        expect(screen.getByText('Rate your performance')).toBeInTheDocument();
      });
    });
  });

  describe('Quick Tags', () => {
    it('should display all quick tags', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Misplay')).toBeInTheDocument();
        expect(screen.getByText('Mana Issues')).toBeInTheDocument();
        expect(screen.getByText('Great Game')).toBeInTheDocument();
        expect(screen.getByText('Close Match')).toBeInTheDocument();
        expect(screen.getByText('Bad Draws')).toBeInTheDocument();
        expect(screen.getByText('Opponent Error')).toBeInTheDocument();
      });
    });

    it('should append quick tag to notes when clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Misplay')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Misplay'));

      const textarea = screen.getByPlaceholderText(/Add notes about this match/);
      expect(textarea).toHaveValue('Made a misplay - ');
    });

    it('should append quick tag on new line when notes already exist', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: 'Existing notes',
        rating: 0,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        expect(screen.getByText('Great Game')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Great Game'));

      const textarea = screen.getByPlaceholderText(/Add notes about this match/);
      expect(textarea).toHaveValue('Existing notes\nGreat game! ');
    });
  });

  describe('Save Notes', () => {
    it('should call save when save button clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      mockNotes.updateMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: 'New notes',
        rating: 4,
      });
      const onClose = vi.fn();

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Save Notes')).toBeInTheDocument();
      });

      // Add notes
      const textarea = screen.getByPlaceholderText(/Add notes about this match/);
      fireEvent.change(textarea, { target: { value: 'New notes' } });

      // Set rating using title
      fireEvent.click(screen.getByTitle('4 stars'));

      // Save
      fireEvent.click(screen.getByText('Save Notes'));

      await waitFor(() => {
        expect(mockNotes.updateMatchNotes).toHaveBeenCalledWith('match-1', {
          notes: 'New notes',
          rating: 4,
        });
      });
    });

    it('should call onSave callback with updated notes', async () => {
      const updatedNotes = { matchId: 'match-1', notes: 'Updated', rating: 5 };
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      mockNotes.updateMatchNotes.mockResolvedValue(updatedNotes);
      const onSave = vi.fn();
      const onClose = vi.fn();

      render(
        <MatchNotesModal
          matchId="match-1"
          isOpen={true}
          onClose={onClose}
          onSave={onSave}
        />
      );

      await waitFor(() => {
        fireEvent.click(screen.getByText('Save Notes'));
      });

      await waitFor(() => {
        expect(onSave).toHaveBeenCalledWith(updatedNotes);
      });
    });

    it('should close modal after successful save', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      mockNotes.updateMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      const onClose = vi.fn();

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={onClose} />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Save Notes'));
      });

      await waitFor(() => {
        expect(onClose).toHaveBeenCalled();
      });
    });

    it('should show saving state while saving', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      let resolveSave: (value: MatchNotes) => void;
      const savePromise = new Promise<MatchNotes>((resolve) => {
        resolveSave = resolve;
      });
      mockNotes.updateMatchNotes.mockReturnValue(savePromise);

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Save Notes'));
      });

      expect(screen.getByText('Saving...')).toBeInTheDocument();

      resolveSave!({ matchId: 'match-1', notes: '', rating: 0 });
    });
  });

  describe('Cancel Button', () => {
    it('should call onClose when cancel clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      const onClose = vi.fn();

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Cancel')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Cancel'));

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe('Close Button', () => {
    it('should call onClose when X button clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      const onClose = vi.fn();

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Match Notes')).toBeInTheDocument();
      });

      // Find the close button in the header
      const closeButtons = screen.getAllByRole('button');
      const closeButton = closeButtons.find((btn) => btn.textContent === 'x');
      expect(closeButton).toBeInTheDocument();
      fireEvent.click(closeButton!);

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe('Overlay Click', () => {
    it('should call onClose when overlay clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      const onClose = vi.fn();

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Match Notes')).toBeInTheDocument();
      });

      // Click the overlay (not the modal content)
      const overlay = document.querySelector('.match-notes-modal-overlay');
      fireEvent.click(overlay!);

      expect(onClose).toHaveBeenCalled();
    });

    it('should not close when modal content clicked', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      const onClose = vi.fn();

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={onClose} />);

      await waitFor(() => {
        expect(screen.getByText('Match Notes')).toBeInTheDocument();
      });

      // Click the modal content
      const modal = document.querySelector('.match-notes-modal');
      fireEvent.click(modal!);

      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe('Error Handling', () => {
    it('should start with empty form if loading fails', async () => {
      mockNotes.getMatchNotes.mockRejectedValue(new Error('Not found'));

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        const textarea = screen.getByPlaceholderText(/Add notes about this match/);
        expect(textarea).toHaveValue('');
        expect(screen.getByText('Rate your performance')).toBeInTheDocument();
      });
    });

    it('should show error message when save fails', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });
      mockNotes.updateMatchNotes.mockRejectedValue(new Error('Save failed'));

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        fireEvent.click(screen.getByText('Save Notes'));
      });

      await waitFor(() => {
        expect(screen.getByText('Save failed')).toBeInTheDocument();
      });
    });
  });

  describe('Notes Input', () => {
    it('should update notes content when typing', async () => {
      mockNotes.getMatchNotes.mockResolvedValue({
        matchId: 'match-1',
        notes: '',
        rating: 0,
      });

      render(<MatchNotesModal matchId="match-1" isOpen={true} onClose={vi.fn()} />);

      await waitFor(() => {
        const textarea = screen.getByPlaceholderText(/Add notes about this match/);
        fireEvent.change(textarea, { target: { value: 'My custom notes' } });
        expect(textarea).toHaveValue('My custom notes');
      });
    });
  });
});

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Decks from './Decks';
import { mockDecks } from '@/test/mocks/apiMock';
import { gui } from '@/types/models';

// Mock useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Helper function to create mock deck list item
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function createMockDeckListItem(overrides: Record<string, any> = {}): gui.DeckListItem {
  return new gui.DeckListItem({
    id: 'deck-1',
    name: 'Test Deck',
    format: 'standard',
    source: 'manual',
    colorIdentity: 'WU',
    cardCount: 60,
    matchesPlayed: 10,
    matchWinRate: 0.6,
    modifiedAt: new Date('2024-01-15T10:00:00').toISOString(),
    lastPlayed: new Date('2024-01-14T10:00:00').toISOString(),
    tags: [],
    ...overrides,
  });
}

// Helper to create multiple mock decks
function createMockDeckList(): gui.DeckListItem[] {
  return [
    createMockDeckListItem({
      id: 'deck-1',
      name: 'Mono Red Aggro',
      format: 'standard',
      source: 'manual',
    }),
    createMockDeckListItem({
      id: 'deck-2',
      name: 'Azorius Control',
      format: 'historic',
      source: 'draft',
    }),
    createMockDeckListItem({
      id: 'deck-3',
      name: 'Imported Deck',
      format: 'explorer',
      source: 'import',
    }),
  ];
}

// Wrapper component with router
function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

// Setup window.go to simulate Wails runtime being ready
function setupWailsRuntime() {
  (window as unknown as Record<string, unknown>).go = {};
}

function clearWailsRuntime() {
  delete (window as unknown as Record<string, unknown>).go;
}

describe('Decks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockClear();
    setupWailsRuntime();
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    clearWailsRuntime();
    vi.useRealTimers();
  });

  describe('Loading State', () => {
    it('should show loading spinner while fetching decks', async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let resolvePromise: (value: any) => void;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const loadingPromise = new Promise<any>((resolve) => {
        resolvePromise = resolve;
      });
      mockDecks.getDecks.mockReturnValue(loadingPromise);

      renderWithRouter(<Decks />);

      // Advance timers to trigger the interval check
      await vi.advanceTimersByTimeAsync(100);

      expect(screen.getByText('Loading decks...')).toBeInTheDocument();

      resolvePromise!(createMockDeckList());
      await waitFor(() => {
        expect(screen.queryByText('Loading decks...')).not.toBeInTheDocument();
      });
    });
  });

  describe('Error State', () => {
    it('should show error state when API fails', async () => {
      mockDecks.getDecks.mockRejectedValue(new Error('Database error'));

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Decks')).toBeInTheDocument();
      });
      expect(screen.getByText('Database error')).toBeInTheDocument();
    });

    it('should show generic error message for non-Error rejections', async () => {
      mockDecks.getDecks.mockRejectedValue('Unknown error');

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Decks')).toBeInTheDocument();
      });
      expect(screen.getByText('Failed to load decks')).toBeInTheDocument();
    });

    it('should have retry button that reloads decks', async () => {
      mockDecks.getDecks.mockRejectedValueOnce(new Error('Temporary error'));
      mockDecks.getDecks.mockResolvedValueOnce(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Error Loading Decks')).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: 'Retry' });
      fireEvent.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show empty state when no decks exist', async () => {
      mockDecks.getDecks.mockResolvedValue([]);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Decks Yet')).toBeInTheDocument();
      });
      expect(screen.getByText('Create your first deck to get started!')).toBeInTheDocument();
    });

    it('should show empty state when API returns null', async () => {
      mockDecks.getDecks.mockResolvedValue(null);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('No Decks Yet')).toBeInTheDocument();
      });
    });

    it('should show create button in empty state', async () => {
      mockDecks.getDecks.mockResolvedValue([]);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });
    });
  });

  describe('Deck List Display', () => {
    it('should render deck cards when decks exist', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });
      expect(screen.getByText('Azorius Control')).toBeInTheDocument();
      expect(screen.getByText('Imported Deck')).toBeInTheDocument();
    });

    it('should display page title', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'My Decks' })).toBeInTheDocument();
      });
    });

    it('should display format for each deck', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('standard')).toBeInTheDocument();
      });
      expect(screen.getByText('historic')).toBeInTheDocument();
      expect(screen.getByText('explorer')).toBeInTheDocument();
    });

    it('should display draft badge for draft decks', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Draft')).toBeInTheDocument();
      });
    });

    it('should display import badge for imported decks', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Import')).toBeInTheDocument();
      });
    });

    it('should display archetype badge when primaryArchetype is present', async () => {
      mockDecks.getDecks.mockResolvedValue([
        createMockDeckListItem({
          id: 'deck-1',
          name: 'Boros Aggro',
          primaryArchetype: 'Boros Aggro',
        }),
      ]);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Boros Aggro', { selector: '.archetype-badge' })).toBeInTheDocument();
      });
    });

    it('should not display archetype badge when primaryArchetype is not present', async () => {
      mockDecks.getDecks.mockResolvedValue([
        createMockDeckListItem({
          id: 'deck-1',
          name: 'Test Deck',
          primaryArchetype: undefined,
        }),
      ]);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Test Deck')).toBeInTheDocument();
      });
      expect(document.querySelector('.archetype-badge')).not.toBeInTheDocument();
    });

    it('should display multiple badges (archetype and source) together', async () => {
      mockDecks.getDecks.mockResolvedValue([
        createMockDeckListItem({
          id: 'deck-1',
          name: 'Draft Midrange',
          source: 'draft',
          primaryArchetype: 'Dimir Midrange',
        }),
      ]);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Dimir Midrange', { selector: '.archetype-badge' })).toBeInTheDocument();
      });
      expect(screen.getByText('Draft', { selector: '.source-badge' })).toBeInTheDocument();
    });

    it('should display modified date when available', async () => {
      mockDecks.getDecks.mockResolvedValue([
        createMockDeckListItem({
          modifiedAt: new Date('2024-01-15T10:00:00').toISOString(),
        }),
      ]);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText(/Modified:/)).toBeInTheDocument();
      });
    });

    it('should show create button in header when decks exist', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });
    });
  });

  describe('Navigation', () => {
    it('should navigate to deck builder when clicking deck card', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deckCard = screen.getByText('Mono Red Aggro').closest('.deck-card');
      fireEvent.click(deckCard!);

      expect(mockNavigate).toHaveBeenCalledWith('/deck-builder/deck-1');
    });

    it('should navigate to deck builder when clicking edit button', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const editButtons = screen.getAllByRole('button', { name: 'Edit' });
      fireEvent.click(editButtons[0]);

      expect(mockNavigate).toHaveBeenCalledWith('/deck-builder/deck-1');
    });
  });

  describe('Create Deck Dialog', () => {
    it('should open create dialog when clicking create button', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));

      expect(screen.getByText('Create New Deck')).toBeInTheDocument();
      expect(screen.getByLabelText('Deck Name')).toBeInTheDocument();
      expect(screen.getByLabelText('Format')).toBeInTheDocument();
    });

    it('should close create dialog when clicking cancel', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));
      expect(screen.getByText('Create New Deck')).toBeInTheDocument();

      fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

      await waitFor(() => {
        expect(screen.queryByText('Create New Deck')).not.toBeInTheDocument();
      });
    });

    it('should close create dialog when clicking close button', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));
      expect(screen.getByText('Create New Deck')).toBeInTheDocument();

      fireEvent.click(screen.getByRole('button', { name: '×' }));

      await waitFor(() => {
        expect(screen.queryByText('Create New Deck')).not.toBeInTheDocument();
      });
    });

    it('should close create dialog when clicking overlay', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));
      expect(screen.getByText('Create New Deck')).toBeInTheDocument();

      const overlay = document.querySelector('.modal-overlay');
      fireEvent.click(overlay!);

      await waitFor(() => {
        expect(screen.queryByText('Create New Deck')).not.toBeInTheDocument();
      });
    });

    it('should create deck and navigate to deck builder', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      mockDecks.createDeck.mockResolvedValue({ ID: 'new-deck-id' });

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));

      const nameInput = screen.getByLabelText('Deck Name');
      fireEvent.change(nameInput, { target: { value: 'My New Deck' } });

      fireEvent.click(screen.getByRole('button', { name: 'Create Deck' }));

      await waitFor(() => {
        expect(mockDecks.createDeck).toHaveBeenCalledWith({
          name: 'My New Deck',
          format: 'standard',
          source: 'manual',
        });
      });
      expect(mockNavigate).toHaveBeenCalledWith('/deck-builder/new-deck-id');
    });

    it('should show alert when creating deck with empty name', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      const alertMock = vi.spyOn(window, 'alert').mockImplementation(() => {});

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));
      fireEvent.click(screen.getByRole('button', { name: 'Create Deck' }));

      expect(alertMock).toHaveBeenCalledWith('Please enter a deck name');
      alertMock.mockRestore();
    });

    it('should allow selecting different formats', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      mockDecks.createDeck.mockResolvedValue({ ID: 'new-deck-id' });

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));

      const formatSelect = screen.getByLabelText('Format');
      fireEvent.change(formatSelect, { target: { value: 'historic' } });

      const nameInput = screen.getByLabelText('Deck Name');
      fireEvent.change(nameInput, { target: { value: 'Historic Deck' } });

      fireEvent.click(screen.getByRole('button', { name: 'Create Deck' }));

      await waitFor(() => {
        expect(mockDecks.createDeck).toHaveBeenCalledWith({
          name: 'Historic Deck',
          format: 'historic',
          source: 'manual',
        });
      });
    });

    it('should create deck when pressing Enter in name input', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      mockDecks.createDeck.mockResolvedValue({ ID: 'new-deck-id' });

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));

      const nameInput = screen.getByLabelText('Deck Name');
      fireEvent.change(nameInput, { target: { value: 'Enter Test Deck' } });
      fireEvent.keyDown(nameInput, { key: 'Enter' });

      await waitFor(() => {
        expect(mockDecks.createDeck).toHaveBeenCalledWith({
          name: 'Enter Test Deck',
          format: 'standard',
          source: 'manual',
        });
      });
    });

    it('should show error alert when deck creation fails', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      mockDecks.createDeck.mockRejectedValue(new Error('Creation failed'));
      const alertMock = vi.spyOn(window, 'alert').mockImplementation(() => {});

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));

      const nameInput = screen.getByLabelText('Deck Name');
      fireEvent.change(nameInput, { target: { value: 'Failing Deck' } });
      fireEvent.click(screen.getByRole('button', { name: 'Create Deck' }));

      await waitFor(() => {
        expect(alertMock).toHaveBeenCalledWith('Creation failed');
      });
      alertMock.mockRestore();
    });
  });

  describe('Delete Deck Dialog', () => {
    it('should open delete dialog when clicking delete button', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      expect(screen.getByText('Delete Deck')).toBeInTheDocument();
      expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();
      expect(screen.getByText('Mono Red Aggro', { selector: 'strong' })).toBeInTheDocument();
    });

    it('should close delete dialog when clicking cancel', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      expect(screen.getByText('Delete Deck')).toBeInTheDocument();

      fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

      await waitFor(() => {
        expect(screen.queryByText('Delete Deck')).not.toBeInTheDocument();
      });
    });

    it('should close delete dialog when clicking overlay', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      const overlay = document.querySelector('.modal-overlay');
      fireEvent.click(overlay!);

      await waitFor(() => {
        expect(screen.queryByText('Delete Deck')).not.toBeInTheDocument();
      });
    });

    it('should delete deck when confirming deletion', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      mockDecks.deleteDeck.mockResolvedValue(undefined);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      // Click the confirm delete button in the modal (has class delete-button-confirm)
      const confirmDeleteButton = document.querySelector('.delete-button-confirm') as HTMLButtonElement;
      fireEvent.click(confirmDeleteButton);

      await waitFor(() => {
        expect(mockDecks.deleteDeck).toHaveBeenCalledWith('deck-1');
      });
    });

    it('should reload decks after successful deletion', async () => {
      mockDecks.getDecks.mockResolvedValueOnce(createMockDeckList());
      mockDecks.getDecks.mockResolvedValueOnce([
        createMockDeckListItem({ id: 'deck-2', name: 'Azorius Control' }),
      ]);
      mockDecks.deleteDeck.mockResolvedValue(undefined);

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      // Click the confirm delete button in the modal
      const confirmDeleteButton = document.querySelector('.delete-button-confirm') as HTMLButtonElement;
      fireEvent.click(confirmDeleteButton);

      await waitFor(() => {
        expect(mockDecks.getDecks).toHaveBeenCalledTimes(2);
      });
    });

    it('should show error alert when deletion fails', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());
      mockDecks.deleteDeck.mockRejectedValue(new Error('Deletion failed'));
      const alertMock = vi.spyOn(window, 'alert').mockImplementation(() => {});

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      // Click the confirm delete button in the modal
      const confirmDeleteButton = document.querySelector('.delete-button-confirm') as HTMLButtonElement;
      fireEvent.click(confirmDeleteButton);

      await waitFor(() => {
        expect(alertMock).toHaveBeenCalledWith('Deletion failed');
      });
      alertMock.mockRestore();
    });

    it('should display warning text in delete dialog', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('Mono Red Aggro')).toBeInTheDocument();
      });

      const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
      fireEvent.click(deleteButtons[0]);

      expect(screen.getByText('This action cannot be undone.')).toBeInTheDocument();
    });
  });

  describe('Format Options', () => {
    it('should have all format options in create dialog', async () => {
      mockDecks.getDecks.mockResolvedValue(createMockDeckList());

      renderWithRouter(<Decks />);

      await vi.advanceTimersByTimeAsync(100);

      await waitFor(() => {
        expect(screen.getByText('+ Create New Deck')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('+ Create New Deck'));

      const formatSelect = screen.getByLabelText('Format') as HTMLSelectElement;
      const options = Array.from(formatSelect.options).map((opt) => opt.value);

      expect(options).toContain('standard');
      expect(options).toContain('alchemy');
      expect(options).toContain('explorer');
      expect(options).toContain('historic');
      expect(options).toContain('timeless');
      expect(options).toContain('brawl');
      expect(options).toContain('limited');
    });
  });
});

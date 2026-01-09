package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// NotesRepository handles database operations for deck notes and match notes.
type NotesRepository interface {
	// CreateDeckNote creates a new note for a deck.
	CreateDeckNote(ctx context.Context, note *models.DeckNote) error

	// GetDeckNotes retrieves all notes for a deck.
	GetDeckNotes(ctx context.Context, deckID string) ([]*models.DeckNote, error)

	// GetDeckNotesByCategory retrieves notes for a deck filtered by category.
	GetDeckNotesByCategory(ctx context.Context, deckID, category string) ([]*models.DeckNote, error)

	// GetDeckNoteByID retrieves a single note by ID.
	GetDeckNoteByID(ctx context.Context, id int64) (*models.DeckNote, error)

	// UpdateDeckNote updates an existing note.
	UpdateDeckNote(ctx context.Context, note *models.DeckNote) error

	// DeleteDeckNote deletes a note by ID.
	DeleteDeckNote(ctx context.Context, id int64) error

	// DeleteDeckNotesByDeck deletes all notes for a deck.
	DeleteDeckNotesByDeck(ctx context.Context, deckID string) error

	// UpdateMatchNotes updates the notes and rating for a match.
	UpdateMatchNotes(ctx context.Context, matchID string, notes string, rating int) error

	// GetMatchNotes retrieves the notes and rating for a match.
	GetMatchNotes(ctx context.Context, matchID string) (*models.MatchNotes, error)
}

// notesRepository is the concrete implementation of NotesRepository.
type notesRepository struct {
	db *sql.DB
}

// NewNotesRepository creates a new notes repository.
func NewNotesRepository(db *sql.DB) NotesRepository {
	return &notesRepository{db: db}
}

// CreateDeckNote creates a new note for a deck.
func (r *notesRepository) CreateDeckNote(ctx context.Context, note *models.DeckNote) error {
	query := `
		INSERT INTO deck_notes (deck_id, content, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now().UTC()
	note.CreatedAt = now
	note.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		note.DeckID,
		note.Content,
		note.Category,
		now.Format("2006-01-02 15:04:05.999999"),
		now.Format("2006-01-02 15:04:05.999999"),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	note.ID = id

	return nil
}

// GetDeckNotes retrieves all notes for a deck.
func (r *notesRepository) GetDeckNotes(ctx context.Context, deckID string) ([]*models.DeckNote, error) {
	query := `
		SELECT id, deck_id, content, category, created_at, updated_at
		FROM deck_notes
		WHERE deck_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanDeckNotes(rows)
}

// GetDeckNotesByCategory retrieves notes for a deck filtered by category.
func (r *notesRepository) GetDeckNotesByCategory(ctx context.Context, deckID, category string) ([]*models.DeckNote, error) {
	query := `
		SELECT id, deck_id, content, category, created_at, updated_at
		FROM deck_notes
		WHERE deck_id = ? AND category = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID, category)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanDeckNotes(rows)
}

// GetDeckNoteByID retrieves a single note by ID.
func (r *notesRepository) GetDeckNoteByID(ctx context.Context, id int64) (*models.DeckNote, error) {
	query := `
		SELECT id, deck_id, content, category, created_at, updated_at
		FROM deck_notes
		WHERE id = ?
	`

	note := &models.DeckNote{}
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&note.ID,
		&note.DeckID,
		&note.Content,
		&note.Category,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	note.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
	note.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", updatedAt)

	return note, nil
}

// UpdateDeckNote updates an existing note.
func (r *notesRepository) UpdateDeckNote(ctx context.Context, note *models.DeckNote) error {
	query := `
		UPDATE deck_notes
		SET content = ?, category = ?, updated_at = ?
		WHERE id = ?
	`

	note.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query,
		note.Content,
		note.Category,
		note.UpdatedAt.Format("2006-01-02 15:04:05.999999"),
		note.ID,
	)
	return err
}

// DeleteDeckNote deletes a note by ID.
func (r *notesRepository) DeleteDeckNote(ctx context.Context, id int64) error {
	query := `DELETE FROM deck_notes WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteDeckNotesByDeck deletes all notes for a deck.
func (r *notesRepository) DeleteDeckNotesByDeck(ctx context.Context, deckID string) error {
	query := `DELETE FROM deck_notes WHERE deck_id = ?`
	_, err := r.db.ExecContext(ctx, query, deckID)
	return err
}

// UpdateMatchNotes updates the notes and rating for a match.
func (r *notesRepository) UpdateMatchNotes(ctx context.Context, matchID string, notes string, rating int) error {
	query := `
		UPDATE matches
		SET notes = ?, rating = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, notes, rating, matchID)
	return err
}

// GetMatchNotes retrieves the notes and rating for a match.
func (r *notesRepository) GetMatchNotes(ctx context.Context, matchID string) (*models.MatchNotes, error) {
	query := `
		SELECT id, COALESCE(notes, '') as notes, COALESCE(rating, 0) as rating
		FROM matches
		WHERE id = ?
	`

	mn := &models.MatchNotes{}
	err := r.db.QueryRowContext(ctx, query, matchID).Scan(&mn.MatchID, &mn.Notes, &mn.Rating)
	if err != nil {
		return nil, err
	}

	return mn, nil
}

// scanDeckNotes scans multiple deck note rows.
func (r *notesRepository) scanDeckNotes(rows *sql.Rows) ([]*models.DeckNote, error) {
	var notes []*models.DeckNote

	for rows.Next() {
		note := &models.DeckNote{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&note.ID,
			&note.DeckID,
			&note.Content,
			&note.Category,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		note.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
		note.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", updatedAt)

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DeckRepository handles database operations for decks.
type DeckRepository interface {
	// Create inserts a new deck into the database.
	Create(ctx context.Context, deck *models.Deck) error

	// Update updates an existing deck.
	Update(ctx context.Context, deck *models.Deck) error

	// GetByID retrieves a deck by its ID.
	GetByID(ctx context.Context, id string) (*models.Deck, error)

	// List retrieves all decks.
	List(ctx context.Context) ([]*models.Deck, error)

	// GetByFormat retrieves all decks for a specific format.
	GetByFormat(ctx context.Context, format string) ([]*models.Deck, error)

	// Delete deletes a deck by its ID.
	Delete(ctx context.Context, id string) error

	// AddCard adds a card to a deck.
	AddCard(ctx context.Context, card *models.DeckCard) error

	// GetCards retrieves all cards in a deck.
	GetCards(ctx context.Context, deckID string) ([]*models.DeckCard, error)

	// RemoveCard removes a card from a deck.
	RemoveCard(ctx context.Context, deckID string, cardID int, board string) error

	// ClearCards removes all cards from a deck.
	ClearCards(ctx context.Context, deckID string) error
}

// deckRepository is the concrete implementation of DeckRepository.
type deckRepository struct {
	db *sql.DB
}

// NewDeckRepository creates a new deck repository.
func NewDeckRepository(db *sql.DB) DeckRepository {
	return &deckRepository{db: db}
}

// Create inserts a new deck into the database.
func (r *deckRepository) Create(ctx context.Context, deck *models.Deck) error {
	query := `
		INSERT INTO decks (
			id, name, format, description, color_identity,
			created_at, modified_at, last_played
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		deck.ID,
		deck.Name,
		deck.Format,
		deck.Description,
		deck.ColorIdentity,
		deck.CreatedAt,
		deck.ModifiedAt,
		deck.LastPlayed,
	)
	if err != nil {
		return fmt.Errorf("failed to create deck: %w", err)
	}

	return nil
}

// Update updates an existing deck.
func (r *deckRepository) Update(ctx context.Context, deck *models.Deck) error {
	query := `
		UPDATE decks
		SET name = ?, format = ?, description = ?, color_identity = ?,
		    modified_at = ?, last_played = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		deck.Name,
		deck.Format,
		deck.Description,
		deck.ColorIdentity,
		deck.ModifiedAt,
		deck.LastPlayed,
		deck.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deck: %w", err)
	}

	return nil
}

// GetByID retrieves a deck by its ID.
func (r *deckRepository) GetByID(ctx context.Context, id string) (*models.Deck, error) {
	query := `
		SELECT id, name, format, description, color_identity,
		       created_at, modified_at, last_played
		FROM decks
		WHERE id = ?
	`

	deck := &models.Deck{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&deck.ID,
		&deck.Name,
		&deck.Format,
		&deck.Description,
		&deck.ColorIdentity,
		&deck.CreatedAt,
		&deck.ModifiedAt,
		&deck.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deck by id: %w", err)
	}

	return deck, nil
}

// List retrieves all decks.
func (r *deckRepository) List(ctx context.Context) ([]*models.Deck, error) {
	query := `
		SELECT id, name, format, description, color_identity,
		       created_at, modified_at, last_played
		FROM decks
		ORDER BY modified_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list decks: %w", err)
	}
	defer rows.Close()

	var decks []*models.Deck
	for rows.Next() {
		deck := &models.Deck{}
		err := rows.Scan(
			&deck.ID,
			&deck.Name,
			&deck.Format,
			&deck.Description,
			&deck.ColorIdentity,
			&deck.CreatedAt,
			&deck.ModifiedAt,
			&deck.LastPlayed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}
		decks = append(decks, deck)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// GetByFormat retrieves all decks for a specific format.
func (r *deckRepository) GetByFormat(ctx context.Context, format string) ([]*models.Deck, error) {
	query := `
		SELECT id, name, format, description, color_identity,
		       created_at, modified_at, last_played
		FROM decks
		WHERE format = ?
		ORDER BY modified_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get decks by format: %w", err)
	}
	defer rows.Close()

	var decks []*models.Deck
	for rows.Next() {
		deck := &models.Deck{}
		err := rows.Scan(
			&deck.ID,
			&deck.Name,
			&deck.Format,
			&deck.Description,
			&deck.ColorIdentity,
			&deck.CreatedAt,
			&deck.ModifiedAt,
			&deck.LastPlayed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}
		decks = append(decks, deck)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// Delete deletes a deck by its ID.
func (r *deckRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM decks WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
	}

	return nil
}

// AddCard adds a card to a deck.
func (r *deckRepository) AddCard(ctx context.Context, card *models.DeckCard) error {
	query := `
		INSERT INTO deck_cards (deck_id, card_id, quantity, board)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(deck_id, card_id, board) DO UPDATE SET
			quantity = excluded.quantity
	`

	result, err := r.db.ExecContext(ctx, query,
		card.DeckID,
		card.CardID,
		card.Quantity,
		card.Board,
	)
	if err != nil {
		return fmt.Errorf("failed to add card to deck: %w", err)
	}

	// If this is an insert, set the ID
	if card.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			card.ID = int(id)
		}
	}

	return nil
}

// GetCards retrieves all cards in a deck.
func (r *deckRepository) GetCards(ctx context.Context, deckID string) ([]*models.DeckCard, error) {
	query := `
		SELECT id, deck_id, card_id, quantity, board
		FROM deck_cards
		WHERE deck_id = ?
		ORDER BY board, card_id
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.DeckCard
	for rows.Next() {
		card := &models.DeckCard{}
		err := rows.Scan(
			&card.ID,
			&card.DeckID,
			&card.CardID,
			&card.Quantity,
			&card.Board,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck card: %w", err)
		}
		cards = append(cards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck cards: %w", err)
	}

	return cards, nil
}

// RemoveCard removes a card from a deck.
func (r *deckRepository) RemoveCard(ctx context.Context, deckID string, cardID int, board string) error {
	query := `DELETE FROM deck_cards WHERE deck_id = ? AND card_id = ? AND board = ?`

	_, err := r.db.ExecContext(ctx, query, deckID, cardID, board)
	if err != nil {
		return fmt.Errorf("failed to remove card from deck: %w", err)
	}

	return nil
}

// ClearCards removes all cards from a deck.
func (r *deckRepository) ClearCards(ctx context.Context, deckID string) error {
	query := `DELETE FROM deck_cards WHERE deck_id = ?`

	_, err := r.db.ExecContext(ctx, query, deckID)
	if err != nil {
		return fmt.Errorf("failed to clear deck cards: %w", err)
	}

	return nil
}

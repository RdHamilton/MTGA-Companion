package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// AccountRepository provides methods for managing accounts.
type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) error
	GetByID(ctx context.Context, id int) (*models.Account, error)
	GetDefault(ctx context.Context) (*models.Account, error)
	GetAll(ctx context.Context) ([]*models.Account, error)
	Update(ctx context.Context, account *models.Account) error
	SetDefault(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}

type accountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new account repository.
func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

// Create creates a new account.
func (r *accountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (name, screen_name, client_id, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		account.Name,
		account.ScreenName,
		account.ClientID,
		account.IsDefault,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	account.ID = int(id)

	// If this is the default account, unset other defaults
	if account.IsDefault {
		if err := r.setDefaultOnly(ctx, int(id)); err != nil {
			return err
		}
	}

	return nil
}

// GetByID retrieves an account by ID.
func (r *accountRepository) GetByID(ctx context.Context, id int) (*models.Account, error) {
	query := `
		SELECT id, name, screen_name, client_id, is_default, created_at, updated_at
		FROM accounts
		WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	account := &models.Account{}
	var screenName, clientID sql.NullString
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&account.ID,
		&account.Name,
		&screenName,
		&clientID,
		&account.IsDefault,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if screenName.Valid {
		account.ScreenName = &screenName.String
	}
	if clientID.Valid {
		account.ClientID = &clientID.String
	}
	account.CreatedAt = createdAt
	account.UpdatedAt = updatedAt

	return account, nil
}

// GetDefault retrieves the default account.
func (r *accountRepository) GetDefault(ctx context.Context) (*models.Account, error) {
	query := `
		SELECT id, name, screen_name, client_id, is_default, created_at, updated_at
		FROM accounts
		WHERE is_default = 1
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query)

	account := &models.Account{}
	var screenName, clientID sql.NullString
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&account.ID,
		&account.Name,
		&screenName,
		&clientID,
		&account.IsDefault,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if screenName.Valid {
		account.ScreenName = &screenName.String
	}
	if clientID.Valid {
		account.ClientID = &clientID.String
	}
	account.CreatedAt = createdAt
	account.UpdatedAt = updatedAt

	return account, nil
}

// GetAll retrieves all accounts.
func (r *accountRepository) GetAll(ctx context.Context) ([]*models.Account, error) {
	query := `
		SELECT id, name, screen_name, client_id, is_default, created_at, updated_at
		FROM accounts
		ORDER BY is_default DESC, name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	accounts := []*models.Account{}
	for rows.Next() {
		account := &models.Account{}
		var screenName, clientID sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&account.ID,
			&account.Name,
			&screenName,
			&clientID,
			&account.IsDefault,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		if screenName.Valid {
			account.ScreenName = &screenName.String
		}
		if clientID.Valid {
			account.ClientID = &clientID.String
		}
		account.CreatedAt = createdAt
		account.UpdatedAt = updatedAt

		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

// Update updates an account.
func (r *accountRepository) Update(ctx context.Context, account *models.Account) error {
	query := `
		UPDATE accounts
		SET name = ?, screen_name = ?, client_id = ?, is_default = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		account.Name,
		account.ScreenName,
		account.ClientID,
		account.IsDefault,
		time.Now(),
		account.ID,
	)
	if err != nil {
		return err
	}

	// If this is the default account, unset other defaults
	if account.IsDefault {
		if err := r.setDefaultOnly(ctx, account.ID); err != nil {
			return err
		}
	}

	return nil
}

// SetDefault sets an account as the default account.
func (r *accountRepository) SetDefault(ctx context.Context, id int) error {
	return r.setDefaultOnly(ctx, id)
}

// setDefaultOnly sets the specified account as default and unsets all others.
func (r *accountRepository) setDefaultOnly(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unset all defaults
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET is_default = 0")
	if err != nil {
		return err
	}

	// Set the specified account as default
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET is_default = 1 WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete deletes an account.
func (r *accountRepository) Delete(ctx context.Context, id int) error {
	// Check if this is the default account
	account, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if account == nil {
		return nil // Account doesn't exist
	}

	// Don't allow deleting the default account if it's the only account
	allAccounts, err := r.GetAll(ctx)
	if err != nil {
		return err
	}
	if len(allAccounts) == 1 {
		return sql.ErrNoRows // Can't delete the last account
	}

	// If deleting the default account, set another account as default
	if account.IsDefault {
		for _, acc := range allAccounts {
			if acc.ID != id {
				if err := r.SetDefault(ctx, acc.ID); err != nil {
					return err
				}
				break
			}
		}
	}

	query := "DELETE FROM accounts WHERE id = ?"
	_, err = r.db.ExecContext(ctx, query, id)
	return err
}


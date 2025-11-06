package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// TxFunc is a function that runs within a transaction.
type TxFunc func(*sql.Tx) error

// WithTransaction executes the given function within a database transaction.
// It automatically commits on success or rolls back on error.
// If the function panics, the transaction is rolled back and the panic is re-raised.
func (db *DB) WithTransaction(ctx context.Context, fn TxFunc) (err error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is closed
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback()
			panic(p) // Re-raise panic
		} else if err != nil {
			// Rollback on error
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
			}
		} else {
			// Commit on success
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("failed to commit transaction: %w", err)
			}
		}
	}()

	err = fn(tx)
	return err
}

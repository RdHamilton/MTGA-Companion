package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// SettingsRepository provides access to user settings.
type SettingsRepository interface {
	// Get retrieves a setting value by key.
	// Returns the JSON-encoded value or error if not found.
	Get(ctx context.Context, key string) (string, error)

	// GetTyped retrieves a setting and unmarshals it to the target type.
	GetTyped(ctx context.Context, key string, target interface{}) error

	// Set stores a setting value.
	// The value is JSON-encoded before storage.
	Set(ctx context.Context, key string, value interface{}) error

	// GetAll retrieves all settings as a map.
	GetAll(ctx context.Context) (map[string]interface{}, error)

	// SetMany stores multiple settings at once.
	SetMany(ctx context.Context, settings map[string]interface{}) error

	// Delete removes a setting.
	Delete(ctx context.Context, key string) error
}

// settingsRepository implements SettingsRepository using SQLite.
type settingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository creates a new settings repository.
func NewSettingsRepository(db *sql.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

// Get retrieves a setting value by key.
func (r *settingsRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("setting not found: %s", key)
		}
		return "", fmt.Errorf("failed to get setting %s: %w", key, err)
	}
	return value, nil
}

// GetTyped retrieves a setting and unmarshals it to the target type.
func (r *settingsRepository) GetTyped(ctx context.Context, key string, target interface{}) error {
	value, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(value), target); err != nil {
		return fmt.Errorf("failed to unmarshal setting %s: %w", key, err)
	}
	return nil
}

// Set stores a setting value.
func (r *settingsRepository) Set(ctx context.Context, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal setting %s: %w", key, err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, string(jsonValue), time.Now())
	if err != nil {
		return fmt.Errorf("failed to set setting %s: %w", key, err)
	}
	return nil
}

// GetAll retrieves all settings as a map.
func (r *settingsRepository) GetAll(ctx context.Context) (map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer func() {
		_ = rows.Close() // Explicitly ignore error - cleanup operation
	}()

	settings := make(map[string]interface{})
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		// Try to unmarshal JSON value
		var parsed interface{}
		if err := json.Unmarshal([]byte(value), &parsed); err != nil {
			// If unmarshal fails, use raw string
			settings[key] = value
		} else {
			settings[key] = parsed
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}

// SetMany stores multiple settings at once.
func (r *settingsRepository) SetMany(ctx context.Context, settings map[string]interface{}) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Explicitly ignore error - will be nil if Commit() succeeds
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close() // Explicitly ignore error - cleanup operation
	}()

	now := time.Now()
	for key, value := range settings {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal setting %s: %w", key, err)
		}
		if _, err := stmt.ExecContext(ctx, key, string(jsonValue), now); err != nil {
			return fmt.Errorf("failed to set setting %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete removes a setting.
func (r *settingsRepository) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM settings WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("failed to delete setting %s: %w", key, err)
	}
	return nil
}

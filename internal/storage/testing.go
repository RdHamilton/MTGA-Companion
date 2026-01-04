package storage

import (
	"database/sql"
)

// NewTestDB creates a new in-memory database wrapped in a DB struct for testing.
// This helper is exported for use in other package tests.
func NewTestDB(sqlDB *sql.DB) *DB {
	return &DB{conn: sqlDB}
}

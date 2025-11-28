package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupSettingsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create settings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create settings table: %v", err)
	}

	return db
}

func TestSettingsRepository_SetAndGet(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Test setting a string value
	err := repo.Set(ctx, "theme", "dark")
	if err != nil {
		t.Fatalf("Failed to set string value: %v", err)
	}

	var theme string
	err = repo.GetTyped(ctx, "theme", &theme)
	if err != nil {
		t.Fatalf("Failed to get string value: %v", err)
	}
	if theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", theme)
	}
}

func TestSettingsRepository_SetAndGetInt(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Test setting an int value
	err := repo.Set(ctx, "refreshInterval", 30)
	if err != nil {
		t.Fatalf("Failed to set int value: %v", err)
	}

	var interval int
	err = repo.GetTyped(ctx, "refreshInterval", &interval)
	if err != nil {
		t.Fatalf("Failed to get int value: %v", err)
	}
	if interval != 30 {
		t.Errorf("Expected interval 30, got %d", interval)
	}
}

func TestSettingsRepository_SetAndGetBool(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Test setting a bool value
	err := repo.Set(ctx, "autoRefresh", true)
	if err != nil {
		t.Fatalf("Failed to set bool value: %v", err)
	}

	var autoRefresh bool
	err = repo.GetTyped(ctx, "autoRefresh", &autoRefresh)
	if err != nil {
		t.Fatalf("Failed to get bool value: %v", err)
	}
	if !autoRefresh {
		t.Error("Expected autoRefresh true, got false")
	}
}

func TestSettingsRepository_GetAll(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Set multiple values
	_ = repo.Set(ctx, "theme", "dark")
	_ = repo.Set(ctx, "refreshInterval", 30)
	_ = repo.Set(ctx, "autoRefresh", false)

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get all settings: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 settings, got %d", len(all))
	}

	if all["theme"] != "dark" {
		t.Errorf("Expected theme 'dark', got '%v'", all["theme"])
	}
}

func TestSettingsRepository_SetMany(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	settings := map[string]interface{}{
		"theme":           "dark",
		"refreshInterval": 30,
		"autoRefresh":     true,
	}

	err := repo.SetMany(ctx, settings)
	if err != nil {
		t.Fatalf("Failed to set many settings: %v", err)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get all settings: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 settings, got %d", len(all))
	}
}

func TestSettingsRepository_UpdateExisting(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Set initial value
	err := repo.Set(ctx, "theme", "light")
	if err != nil {
		t.Fatalf("Failed to set initial value: %v", err)
	}

	// Update value
	err = repo.Set(ctx, "theme", "dark")
	if err != nil {
		t.Fatalf("Failed to update value: %v", err)
	}

	var theme string
	err = repo.GetTyped(ctx, "theme", &theme)
	if err != nil {
		t.Fatalf("Failed to get updated value: %v", err)
	}
	if theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", theme)
	}
}

func TestSettingsRepository_Delete(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	// Set and then delete
	_ = repo.Set(ctx, "theme", "dark")
	err := repo.Delete(ctx, "theme")
	if err != nil {
		t.Fatalf("Failed to delete setting: %v", err)
	}

	_, err = repo.Get(ctx, "theme")
	if err == nil {
		t.Error("Expected error when getting deleted setting, got nil")
	}
}

func TestSettingsRepository_GetNotFound(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	repo := NewSettingsRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent setting, got nil")
	}
}

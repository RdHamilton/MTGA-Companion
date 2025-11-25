package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Inventory represents the player's current inventory.
type Inventory struct {
	ID            int       `json:"id"`
	Gold          int       `json:"gold"`
	Gems          int       `json:"gems"`
	WCCommon      int       `json:"wcCommon"`
	WCUncommon    int       `json:"wcUncommon"`
	WCRare        int       `json:"wcRare"`
	WCMythic      int       `json:"wcMythic"`
	VaultProgress float64   `json:"vaultProgress"`
	DraftTokens   int       `json:"draftTokens"`
	SealedTokens  int       `json:"sealedTokens"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// InventoryChange represents a change to an inventory field.
type InventoryChange struct {
	ID            int       `json:"id"`
	Field         string    `json:"field"`
	PreviousValue int       `json:"previousValue"`
	NewValue      int       `json:"newValue"`
	Delta         int       `json:"delta"`
	Source        *string   `json:"source,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// InventoryRepository handles database operations for inventory tracking.
type InventoryRepository interface {
	// Get retrieves the current inventory.
	Get(ctx context.Context) (*Inventory, error)

	// Update updates the inventory with new values and records changes.
	Update(ctx context.Context, inv *Inventory, source *string) ([]InventoryChange, error)

	// GetHistory retrieves inventory change history.
	GetHistory(ctx context.Context, field string, limit int) ([]*InventoryChange, error)

	// GetRecentChanges retrieves all recent inventory changes.
	GetRecentChanges(ctx context.Context, limit int) ([]*InventoryChange, error)

	// GetChangesSince retrieves inventory changes since a specific time.
	GetChangesSince(ctx context.Context, since time.Time) ([]*InventoryChange, error)
}

// inventoryRepository is the concrete implementation of InventoryRepository.
type inventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository creates a new inventory repository.
func NewInventoryRepository(db *sql.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

// Get retrieves the current inventory.
func (r *inventoryRepository) Get(ctx context.Context) (*Inventory, error) {
	query := `
		SELECT id, gold, gems, wc_common, wc_uncommon, wc_rare, wc_mythic,
		       vault_progress, draft_tokens, sealed_tokens, updated_at
		FROM inventory
		ORDER BY id DESC
		LIMIT 1
	`

	inv := &Inventory{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&inv.ID,
		&inv.Gold,
		&inv.Gems,
		&inv.WCCommon,
		&inv.WCUncommon,
		&inv.WCRare,
		&inv.WCMythic,
		&inv.VaultProgress,
		&inv.DraftTokens,
		&inv.SealedTokens,
		&inv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Return empty inventory if none exists
		return &Inventory{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inv, nil
}

// Update updates the inventory with new values and records changes.
func (r *inventoryRepository) Update(ctx context.Context, newInv *Inventory, source *string) ([]InventoryChange, error) {
	// Get current inventory for change detection
	current, err := r.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current inventory: %w", err)
	}

	// Detect changes
	changes := r.detectChanges(current, newInv, source)

	// Update inventory
	query := `
		UPDATE inventory SET
			gold = ?,
			gems = ?,
			wc_common = ?,
			wc_uncommon = ?,
			wc_rare = ?,
			wc_mythic = ?,
			vault_progress = ?,
			draft_tokens = ?,
			sealed_tokens = ?,
			updated_at = ?
		WHERE id = 1
	`

	_, err = r.db.ExecContext(ctx, query,
		newInv.Gold,
		newInv.Gems,
		newInv.WCCommon,
		newInv.WCUncommon,
		newInv.WCRare,
		newInv.WCMythic,
		newInv.VaultProgress,
		newInv.DraftTokens,
		newInv.SealedTokens,
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Record changes in history
	for i := range changes {
		if err := r.recordChange(ctx, &changes[i]); err != nil {
			// Log but don't fail the update
			continue
		}
	}

	return changes, nil
}

// detectChanges detects changes between current and new inventory.
func (r *inventoryRepository) detectChanges(current, newInv *Inventory, source *string) []InventoryChange {
	var changes []InventoryChange
	now := time.Now()

	checkField := func(field string, oldVal, newVal int) {
		if oldVal != newVal {
			changes = append(changes, InventoryChange{
				Field:         field,
				PreviousValue: oldVal,
				NewValue:      newVal,
				Delta:         newVal - oldVal,
				Source:        source,
				CreatedAt:     now,
			})
		}
	}

	checkField("gold", current.Gold, newInv.Gold)
	checkField("gems", current.Gems, newInv.Gems)
	checkField("wc_common", current.WCCommon, newInv.WCCommon)
	checkField("wc_uncommon", current.WCUncommon, newInv.WCUncommon)
	checkField("wc_rare", current.WCRare, newInv.WCRare)
	checkField("wc_mythic", current.WCMythic, newInv.WCMythic)
	checkField("draft_tokens", current.DraftTokens, newInv.DraftTokens)
	checkField("sealed_tokens", current.SealedTokens, newInv.SealedTokens)

	// Check vault progress (convert to int for comparison, as small changes may be noise)
	oldVault := int(current.VaultProgress * 100)
	newVault := int(newInv.VaultProgress * 100)
	if oldVault != newVault {
		changes = append(changes, InventoryChange{
			Field:         "vault_progress",
			PreviousValue: oldVault,
			NewValue:      newVault,
			Delta:         newVault - oldVault,
			Source:        source,
			CreatedAt:     now,
		})
	}

	return changes
}

// recordChange records a single inventory change.
func (r *inventoryRepository) recordChange(ctx context.Context, change *InventoryChange) error {
	query := `
		INSERT INTO inventory_history (field, previous_value, new_value, delta, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		change.Field,
		change.PreviousValue,
		change.NewValue,
		change.Delta,
		change.Source,
		change.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record inventory change: %w", err)
	}

	return nil
}

// GetHistory retrieves inventory change history for a specific field.
func (r *inventoryRepository) GetHistory(ctx context.Context, field string, limit int) ([]*InventoryChange, error) {
	query := `
		SELECT id, field, previous_value, new_value, delta, source, created_at
		FROM inventory_history
		WHERE field = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, field, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory history: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup
		_ = rows.Close()
	}()

	return r.scanHistory(rows)
}

// GetRecentChanges retrieves all recent inventory changes.
func (r *inventoryRepository) GetRecentChanges(ctx context.Context, limit int) ([]*InventoryChange, error) {
	query := `
		SELECT id, field, previous_value, new_value, delta, source, created_at
		FROM inventory_history
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent changes: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup
		_ = rows.Close()
	}()

	return r.scanHistory(rows)
}

// GetChangesSince retrieves inventory changes since a specific time.
func (r *inventoryRepository) GetChangesSince(ctx context.Context, since time.Time) ([]*InventoryChange, error) {
	query := `
		SELECT id, field, previous_value, new_value, delta, source, created_at
		FROM inventory_history
		WHERE created_at > ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes since: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup
		_ = rows.Close()
	}()

	return r.scanHistory(rows)
}

// scanHistory scans rows into InventoryChange slice.
func (r *inventoryRepository) scanHistory(rows *sql.Rows) ([]*InventoryChange, error) {
	var history []*InventoryChange
	for rows.Next() {
		h := &InventoryChange{}
		err := rows.Scan(
			&h.ID,
			&h.Field,
			&h.PreviousValue,
			&h.NewValue,
			&h.Delta,
			&h.Source,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}

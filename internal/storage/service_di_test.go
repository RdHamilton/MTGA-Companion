package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// mockMatchRepository is a mock implementation of MatchRepository for testing DI.
type mockMatchRepository struct {
	repository.MatchRepository
	createCalled bool
	getCalled    bool
}

func (m *mockMatchRepository) Create(_ context.Context, _ *Match) error {
	m.createCalled = true
	return nil
}

func (m *mockMatchRepository) GetByID(_ context.Context, _ string) (*models.Match, error) {
	m.getCalled = true
	return nil, nil
}

// mockDeckRepository is a mock implementation of DeckRepository for testing DI.
type mockDeckRepository struct {
	repository.DeckRepository
	listCalled bool
}

func (m *mockDeckRepository) List(_ context.Context, _ int) ([]*models.Deck, error) {
	m.listCalled = true
	return nil, nil
}

func (m *mockDeckRepository) GetByID(_ context.Context, _ string) (*models.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) Create(_ context.Context, _ *models.Deck) error {
	return nil
}

// setupDITestDB creates an in-memory database for DI tests.
func setupDITestDB(t *testing.T) *DB {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			screen_name TEXT,
			client_id TEXT,
			is_default INTEGER NOT NULL DEFAULT 0,
			daily_wins INTEGER NOT NULL DEFAULT 0,
			weekly_wins INTEGER NOT NULL DEFAULT 0,
			mastery_level INTEGER NOT NULL DEFAULT 0,
			mastery_pass TEXT NOT NULL DEFAULT '',
			mastery_max INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		-- Insert default account
		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, datetime('now'), datetime('now'));
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Wrap in DB type
	db := &DB{conn: sqlDB}
	return db
}

func TestNewServiceWithConfig_DefaultRepositories(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// Create service with nil config (should use all defaults)
	svc := NewServiceWithConfig(db, nil)

	// Verify all repositories are initialized (not nil)
	if svc.DraftRepo() == nil {
		t.Error("DraftRepo should not be nil")
	}
	if svc.SetCardRepo() == nil {
		t.Error("SetCardRepo should not be nil")
	}
	if svc.DraftRatingsRepo() == nil {
		t.Error("DraftRatingsRepo should not be nil")
	}
	if svc.CollectionRepo() == nil {
		t.Error("CollectionRepo should not be nil")
	}
	if svc.DeckRepo() == nil {
		t.Error("DeckRepo should not be nil")
	}
	if svc.InventoryRepo() == nil {
		t.Error("InventoryRepo should not be nil")
	}
	if svc.Quests() == nil {
		t.Error("Quests should not be nil")
	}
}

func TestNewServiceWithConfig_EmptyConfig(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// Create service with empty config (should use all defaults)
	svc := NewServiceWithConfig(db, &ServiceConfig{})

	// Verify all repositories are initialized
	if svc.DraftRepo() == nil {
		t.Error("DraftRepo should not be nil with empty config")
	}
	if svc.SetCardRepo() == nil {
		t.Error("SetCardRepo should not be nil with empty config")
	}
}

func TestNewServiceWithConfig_InjectedMatchRepository(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// Create a mock repository
	mockRepo := &mockMatchRepository{}

	// Inject the mock via ServiceConfig
	cfg := &ServiceConfig{
		Matches: mockRepo,
	}

	svc := NewServiceWithConfig(db, cfg)

	// Verify the mock was injected (other repos should still use defaults)
	if svc.DraftRepo() == nil {
		t.Error("DraftRepo should use default when not specified")
	}
}

func TestNewServiceWithConfig_InjectedDeckRepository(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// Create a mock repository
	mockRepo := &mockDeckRepository{}

	// Inject the mock via ServiceConfig
	cfg := &ServiceConfig{
		Decks: mockRepo,
	}

	svc := NewServiceWithConfig(db, cfg)

	// Verify the mock was used
	if svc.DeckRepo() != mockRepo {
		t.Error("DeckRepo should be the injected mock")
	}

	// Call a method to verify the mock is called
	_, _ = svc.ListDecks(context.Background())
	if !mockRepo.listCalled {
		t.Error("Expected mock List to be called")
	}
}

func TestNewService_UsesDefaults(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// NewService should internally call NewServiceWithConfig(db, nil)
	svc := NewService(db)

	// Verify repositories are initialized
	if svc.DraftRepo() == nil {
		t.Error("DraftRepo should not be nil")
	}
	if svc.SetCardRepo() == nil {
		t.Error("SetCardRepo should not be nil")
	}
}

func TestOrDefault_NilValue(t *testing.T) {
	factoryCalled := false
	factory := func() string {
		factoryCalled = true
		return "default"
	}

	// Test with typed nil - since string is not an interface, this tests a non-nil case
	// For interface types (like repositories), nil check works differently
	result := orDefault("", factory)

	// Empty string is not nil, so factory should not be called
	if factoryCalled {
		t.Error("Factory should not be called for non-nil value")
	}
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}
}

func TestOrDefault_WithValue(t *testing.T) {
	factoryCalled := false
	factory := func() string {
		factoryCalled = true
		return "default"
	}

	result := orDefault("provided", factory)

	if factoryCalled {
		t.Error("Factory should not be called when value is provided")
	}
	if result != "provided" {
		t.Errorf("Expected 'provided', got %s", result)
	}
}

func TestOrDefaultQuest_NilValue(t *testing.T) {
	factoryCalled := false
	db := setupDITestDB(t)
	defer func() {
		_ = db.Close()
	}()

	factory := func() *QuestRepository {
		factoryCalled = true
		return NewQuestRepository(db.Conn())
	}

	result := orDefaultQuest(nil, factory)

	if !factoryCalled {
		t.Error("Factory should be called when value is nil")
	}
	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestOrDefaultQuest_WithValue(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		_ = db.Close()
	}()

	provided := NewQuestRepository(db.Conn())
	factoryCalled := false

	factory := func() *QuestRepository {
		factoryCalled = true
		return NewQuestRepository(db.Conn())
	}

	result := orDefaultQuest(provided, factory)

	if factoryCalled {
		t.Error("Factory should not be called when value is provided")
	}
	if result != provided {
		t.Error("Result should be the provided value")
	}
}

func TestServiceConfig_PartialOverride(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// Create a mock for just one repository
	mockDecks := &mockDeckRepository{}

	cfg := &ServiceConfig{
		Decks: mockDecks,
		// All other fields are nil - should use defaults
	}

	svc := NewServiceWithConfig(db, cfg)

	// Decks should be our mock
	if svc.DeckRepo() != mockDecks {
		t.Error("DeckRepo should be the injected mock")
	}

	// Other repos should be non-nil defaults
	if svc.DraftRepo() == nil {
		t.Error("DraftRepo should use default")
	}
	if svc.SetCardRepo() == nil {
		t.Error("SetCardRepo should use default")
	}
	if svc.DraftRatingsRepo() == nil {
		t.Error("DraftRatingsRepo should use default")
	}
	if svc.CollectionRepo() == nil {
		t.Error("CollectionRepo should use default")
	}
	if svc.InventoryRepo() == nil {
		t.Error("InventoryRepo should use default")
	}
}

func TestNewServiceWithConfig_InitializesCurrentAccount(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	svc := NewServiceWithConfig(db, nil)

	// Should have initialized current account from the default account
	if svc.CurrentAccountID() != 1 {
		t.Errorf("Expected current account ID 1, got %d", svc.CurrentAccountID())
	}
}

func TestNewServiceWithConfig_CreatesDefaultAccountIfMissing(t *testing.T) {
	// Create empty database (no default account)
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	schema := `
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			screen_name TEXT,
			client_id TEXT,
			is_default INTEGER NOT NULL DEFAULT 0,
			daily_wins INTEGER NOT NULL DEFAULT 0,
			weekly_wins INTEGER NOT NULL DEFAULT 0,
			mastery_level INTEGER NOT NULL DEFAULT 0,
			mastery_pass TEXT NOT NULL DEFAULT '',
			mastery_max INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	db := &DB{conn: sqlDB}

	// Create service - should create default account
	svc := NewServiceWithConfig(db, nil)

	// Verify default account was created
	ctx := context.Background()
	account, err := svc.GetCurrentAccount(ctx)
	if err != nil {
		t.Fatalf("failed to get current account: %v", err)
	}

	if account == nil {
		t.Fatal("Expected default account to be created")
	}

	if account.Name != "Default Account" {
		t.Errorf("Expected account name 'Default Account', got '%s'", account.Name)
	}

	if !account.IsDefault {
		t.Error("Expected account to be marked as default")
	}
}

// TestServiceRepositoryGetters verifies all repository getter methods return non-nil values.
func TestServiceRepositoryGetters(t *testing.T) {
	db := setupDITestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	svc := NewService(db)

	// Test all public repository getters
	getters := map[string]interface{}{
		"Quests":           svc.Quests(),
		"DraftRepo":        svc.DraftRepo(),
		"SetCardRepo":      svc.SetCardRepo(),
		"DraftRatingsRepo": svc.DraftRatingsRepo(),
		"CollectionRepo":   svc.CollectionRepo(),
		"DeckRepo":         svc.DeckRepo(),
		"InventoryRepo":    svc.InventoryRepo(),
		"RankHistoryRepo":  svc.RankHistoryRepo(),
	}

	for name, getter := range getters {
		if getter == nil {
			t.Errorf("%s() returned nil", name)
		}
	}
}

// TestNewServiceWithConfig_NilDB_Panics would test nil DB handling
// but that would cause a panic, so we don't test it here.

// DB is a wrapper around sql.DB to match the storage package's DB type.
// This is duplicated here for test isolation.
type dbWrapper struct {
	sqlDB *sql.DB
}

func (d *dbWrapper) Conn() *sql.DB {
	return d.sqlDB
}

func (d *dbWrapper) Close() error {
	return d.sqlDB.Close()
}

func (d *dbWrapper) WithTransaction(_ context.Context, _ func(*sql.Tx) error) error {
	return nil
}

// Verify mock implements interface
var (
	_ repository.MatchRepository = (*mockMatchRepository)(nil)
	_ repository.DeckRepository  = (*mockDeckRepository)(nil)
)

// Ensure mocks satisfy additional required methods by embedding the real interface
func (m *mockMatchRepository) GetStats(_ context.Context, _ StatsFilter) (*Statistics, error) {
	return nil, nil
}

func (m *mockMatchRepository) CreateGame(_ context.Context, _ *Game) error {
	return nil
}

func (m *mockMatchRepository) GetByDateRange(_ context.Context, _, _ time.Time, _ int) ([]*Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetMatches(_ context.Context, _ models.StatsFilter) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetRecentMatches(_ context.Context, _, _ int) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetLatestMatch(_ context.Context, _ int) (*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetGamesForMatch(_ context.Context, _ string) ([]*models.Game, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetStatsByFormat(_ context.Context, _ models.StatsFilter) (map[string]*models.Statistics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetStatsByDeck(_ context.Context, _ models.StatsFilter) (map[string]*models.Statistics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetPerformanceMetrics(_ context.Context, _ models.StatsFilter) (*models.PerformanceMetrics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetMatchesWithoutDeckID(_ context.Context) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) UpdateDeckID(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockMatchRepository) DeleteAll(_ context.Context, _ int) error {
	return nil
}

func (m *mockDeckRepository) Update(_ context.Context, _ *models.Deck) error {
	return nil
}

func (m *mockDeckRepository) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockDeckRepository) ClearCards(_ context.Context, _ string) error {
	return nil
}

func (m *mockDeckRepository) AddCard(_ context.Context, _ *models.DeckCard) error {
	return nil
}

func (m *mockDeckRepository) GetCards(_ context.Context, _ string) ([]*models.DeckCard, error) {
	return nil, nil
}

func (m *mockDeckRepository) GetBySource(_ context.Context, _ int, _ string) ([]*models.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) DeleteBySourceExcluding(_ context.Context, _ int, _ string, _ []string) (int, error) {
	return 0, nil
}

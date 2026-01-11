package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	_ "modernc.org/sqlite"
)

func setupMTGZoneTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create required tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS mtgzone_archetypes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			format TEXT NOT NULL,
			tier TEXT,
			description TEXT,
			play_style TEXT,
			source_url TEXT,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, format)
		);

		CREATE INDEX IF NOT EXISTS idx_mtgzone_archetypes_format ON mtgzone_archetypes(format);
		CREATE INDEX IF NOT EXISTS idx_mtgzone_archetypes_tier ON mtgzone_archetypes(tier);

		CREATE TABLE IF NOT EXISTS mtgzone_archetype_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			archetype_id INTEGER NOT NULL,
			card_name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'flex',
			copies INTEGER DEFAULT 0,
			importance TEXT DEFAULT 'optional',
			notes TEXT,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (archetype_id) REFERENCES mtgzone_archetypes(id) ON DELETE CASCADE,
			UNIQUE(archetype_id, card_name)
		);

		CREATE INDEX IF NOT EXISTS idx_mtgzone_archetype_cards_archetype ON mtgzone_archetype_cards(archetype_id);
		CREATE INDEX IF NOT EXISTS idx_mtgzone_archetype_cards_card ON mtgzone_archetype_cards(card_name);
		CREATE INDEX IF NOT EXISTS idx_mtgzone_archetype_cards_role ON mtgzone_archetype_cards(role);

		CREATE TABLE IF NOT EXISTS mtgzone_synergies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_a TEXT NOT NULL,
			card_b TEXT NOT NULL,
			reason TEXT NOT NULL,
			source_url TEXT,
			archetype_context TEXT,
			confidence REAL DEFAULT 0.5,
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(card_a, card_b, archetype_context)
		);

		CREATE INDEX IF NOT EXISTS idx_mtgzone_synergies_card_a ON mtgzone_synergies(card_a);
		CREATE INDEX IF NOT EXISTS idx_mtgzone_synergies_card_b ON mtgzone_synergies(card_b);

		CREATE TABLE IF NOT EXISTS mtgzone_articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			article_type TEXT,
			format TEXT,
			archetype TEXT,
			published_at TIMESTAMP,
			processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			cards_mentioned TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_mtgzone_articles_type ON mtgzone_articles(article_type);
		CREATE INDEX IF NOT EXISTS idx_mtgzone_articles_format ON mtgzone_articles(format);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	return db
}

func TestMTGZoneRepository_UpsertArchetype(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	archetype := &models.MTGZoneArchetype{
		Name:        "Mono-Red Aggro",
		Format:      "Standard",
		Tier:        "S",
		Description: "Fast aggressive deck",
		PlayStyle:   "aggro",
		SourceURL:   "https://mtgazone.com/mono-red",
	}

	id, err := repo.UpsertArchetype(ctx, archetype)
	if err != nil {
		t.Fatalf("Failed to upsert archetype: %v", err)
	}

	if id == 0 {
		t.Error("Expected archetype ID to be returned after creation")
	}
}

func TestMTGZoneRepository_GetArchetypeByID(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create an archetype
	archetype := &models.MTGZoneArchetype{
		Name:      "Domain Ramp",
		Format:    "Standard",
		Tier:      "S",
		PlayStyle: "midrange",
	}
	id, _ := repo.UpsertArchetype(ctx, archetype)

	// Retrieve it
	retrieved, err := repo.GetArchetypeByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get archetype by ID: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected archetype, got nil")
	}

	if retrieved.Name != "Domain Ramp" {
		t.Errorf("Name = %q, want %q", retrieved.Name, "Domain Ramp")
	}
	if retrieved.Tier != "S" {
		t.Errorf("Tier = %q, want %q", retrieved.Tier, "S")
	}
}

func TestMTGZoneRepository_GetArchetype(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetypes
	archetype1 := &models.MTGZoneArchetype{
		Name:   "Azorius Control",
		Format: "Standard",
		Tier:   "A",
	}
	archetype2 := &models.MTGZoneArchetype{
		Name:   "Azorius Control",
		Format: "Historic",
		Tier:   "B",
	}
	_, _ = repo.UpsertArchetype(ctx, archetype1)
	_, _ = repo.UpsertArchetype(ctx, archetype2)

	// Get by name and format
	retrieved, err := repo.GetArchetype(ctx, "Azorius Control", "Standard")
	if err != nil {
		t.Fatalf("Failed to get archetype: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected archetype, got nil")
	}

	if retrieved.Format != "Standard" {
		t.Errorf("Format = %q, want %q", retrieved.Format, "Standard")
	}
	if retrieved.Tier != "A" {
		t.Errorf("Tier = %q, want %q", retrieved.Tier, "A")
	}
}

func TestMTGZoneRepository_GetArchetype_CaseInsensitive(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	archetype := &models.MTGZoneArchetype{
		Name:   "Mono-Red Aggro",
		Format: "Standard",
		Tier:   "S",
	}
	_, _ = repo.UpsertArchetype(ctx, archetype)

	// Search with different case
	retrieved, err := repo.GetArchetype(ctx, "mono-red aggro", "standard")
	if err != nil {
		t.Fatalf("Failed to get archetype: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected case-insensitive match, got nil")
	}
}

func TestMTGZoneRepository_GetArchetypesByFormat(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetypes in different formats
	archetypes := []*models.MTGZoneArchetype{
		{Name: "Deck A", Format: "Standard", Tier: "S"},
		{Name: "Deck B", Format: "Standard", Tier: "A"},
		{Name: "Deck C", Format: "Historic", Tier: "S"},
	}
	for _, a := range archetypes {
		_, _ = repo.UpsertArchetype(ctx, a)
	}

	// Get Standard archetypes
	standardArchetypes, err := repo.GetArchetypesByFormat(ctx, "Standard")
	if err != nil {
		t.Fatalf("Failed to get archetypes by format: %v", err)
	}

	if len(standardArchetypes) != 2 {
		t.Errorf("Expected 2 Standard archetypes, got %d", len(standardArchetypes))
	}
}

func TestMTGZoneRepository_GetTopTierArchetypes(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetypes with different tiers
	archetypes := []*models.MTGZoneArchetype{
		{Name: "Deck A", Format: "Standard", Tier: "S"},
		{Name: "Deck B", Format: "Standard", Tier: "A"},
		{Name: "Deck C", Format: "Standard", Tier: "B"},
		{Name: "Deck D", Format: "Standard", Tier: "C"},
	}
	for _, a := range archetypes {
		_, _ = repo.UpsertArchetype(ctx, a)
	}

	// Get top tier archetypes (S, A, A+, A-)
	topArchetypes, err := repo.GetTopTierArchetypes(ctx, "Standard", 10)
	if err != nil {
		t.Fatalf("Failed to get top tier archetypes: %v", err)
	}

	if len(topArchetypes) != 2 {
		t.Errorf("Expected 2 top tier archetypes (S and A), got %d", len(topArchetypes))
	}
}

func TestMTGZoneRepository_UpsertArchetype_Updates(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetype
	archetype := &models.MTGZoneArchetype{
		Name:   "Test Deck",
		Format: "Standard",
		Tier:   "B",
	}
	firstID, _ := repo.UpsertArchetype(ctx, archetype)

	// Upsert with updated tier
	archetype2 := &models.MTGZoneArchetype{
		Name:        "Test Deck",
		Format:      "Standard",
		Tier:        "A",
		Description: "Updated description",
	}
	secondID, err := repo.UpsertArchetype(ctx, archetype2)
	if err != nil {
		t.Fatalf("Failed to upsert archetype: %v", err)
	}

	// Should return same ID (update, not insert)
	if secondID != firstID {
		t.Errorf("Expected same ID after upsert, got %d and %d", firstID, secondID)
	}

	// Verify update
	retrieved, _ := repo.GetArchetypeByID(ctx, firstID)
	if retrieved.Tier != "A" {
		t.Errorf("Tier = %q, want %q", retrieved.Tier, "A")
	}
	if retrieved.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", retrieved.Description, "Updated description")
	}
}

func TestMTGZoneRepository_DeleteArchetype(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetype
	archetype := &models.MTGZoneArchetype{
		Name:   "To Delete",
		Format: "Standard",
		Tier:   "C",
	}
	id, _ := repo.UpsertArchetype(ctx, archetype)

	// Delete it
	err := repo.DeleteArchetype(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete archetype: %v", err)
	}

	// Verify deletion
	retrieved, err := repo.GetArchetypeByID(ctx, id)
	if err != nil {
		t.Fatalf("Error getting archetype: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected archetype to be deleted")
	}
}

func TestMTGZoneRepository_UpsertArchetypeCard(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetype
	archetype := &models.MTGZoneArchetype{
		Name:   "Test Deck",
		Format: "Standard",
	}
	id, _ := repo.UpsertArchetype(ctx, archetype)

	// Add card
	card := &models.MTGZoneArchetypeCard{
		ArchetypeID: id,
		CardName:    "Lightning Bolt",
		Role:        models.CardRoleCore,
		Copies:      4,
		Importance:  models.CardImportanceEssential,
	}
	err := repo.UpsertArchetypeCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to upsert archetype card: %v", err)
	}
}

func TestMTGZoneRepository_GetArchetypeCards(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetype and cards
	archetype := &models.MTGZoneArchetype{
		Name:   "Test Deck",
		Format: "Standard",
	}
	id, _ := repo.UpsertArchetype(ctx, archetype)

	cards := []*models.MTGZoneArchetypeCard{
		{ArchetypeID: id, CardName: "Card A", Role: models.CardRoleCore, Copies: 4},
		{ArchetypeID: id, CardName: "Card B", Role: models.CardRoleFlex, Copies: 2},
		{ArchetypeID: id, CardName: "Card C", Role: models.CardRoleSideboard, Copies: 3},
	}
	for _, c := range cards {
		_ = repo.UpsertArchetypeCard(ctx, c)
	}

	// Get all cards
	retrievedCards, err := repo.GetArchetypeCards(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get archetype cards: %v", err)
	}

	if len(retrievedCards) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(retrievedCards))
	}
}

func TestMTGZoneRepository_GetCoreCards(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetype and cards
	archetype := &models.MTGZoneArchetype{
		Name:   "Test Deck",
		Format: "Standard",
	}
	id, _ := repo.UpsertArchetype(ctx, archetype)

	cards := []*models.MTGZoneArchetypeCard{
		{ArchetypeID: id, CardName: "Core A", Role: models.CardRoleCore, Copies: 4},
		{ArchetypeID: id, CardName: "Core B", Role: models.CardRoleCore, Copies: 4},
		{ArchetypeID: id, CardName: "Flex C", Role: models.CardRoleFlex, Copies: 2},
	}
	for _, c := range cards {
		_ = repo.UpsertArchetypeCard(ctx, c)
	}

	// Get only core cards
	coreCards, err := repo.GetCoreCards(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get core cards: %v", err)
	}

	if len(coreCards) != 2 {
		t.Errorf("Expected 2 core cards, got %d", len(coreCards))
	}
}

func TestMTGZoneRepository_GetArchetypesForCard(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetypes
	id1, _ := repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck A", Format: "Standard", Tier: "S"})
	id2, _ := repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck B", Format: "Standard", Tier: "A"})
	id3, _ := repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck C", Format: "Standard", Tier: "B"})

	// Add shared card to multiple archetypes
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id1, CardName: "Lightning Bolt", Role: models.CardRoleCore})
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id2, CardName: "Lightning Bolt", Role: models.CardRoleFlex})
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id3, CardName: "Different Card", Role: models.CardRoleCore})

	// Get archetypes for Lightning Bolt
	archetypes, err := repo.GetArchetypesForCard(ctx, "Lightning Bolt")
	if err != nil {
		t.Fatalf("Failed to get archetypes for card: %v", err)
	}

	if len(archetypes) != 2 {
		t.Errorf("Expected 2 archetypes with Lightning Bolt, got %d", len(archetypes))
	}

	// Verify both S and A tier archetypes are returned
	tiers := make(map[string]bool)
	for _, a := range archetypes {
		tiers[a.Tier] = true
	}
	if !tiers["S"] || !tiers["A"] {
		t.Error("Expected both S and A tier archetypes")
	}
}

func TestMTGZoneRepository_GetArchetypesForCard_CaseInsensitive(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	id, _ := repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck A", Format: "Standard"})
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id, CardName: "Lightning Bolt", Role: models.CardRoleCore})

	// Search with different case
	archetypes, err := repo.GetArchetypesForCard(ctx, "lightning bolt")
	if err != nil {
		t.Fatalf("Failed to get archetypes: %v", err)
	}

	if len(archetypes) != 1 {
		t.Errorf("Expected case-insensitive match, got %d archetypes", len(archetypes))
	}
}

func TestMTGZoneRepository_UpsertSynergy(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	synergy := &models.MTGZoneSynergy{
		CardA:            "Fatal Push",
		CardB:            "Thoughtseize",
		Reason:           "Both provide efficient interaction in Rakdos",
		ArchetypeContext: "Rakdos Midrange",
		Confidence:       0.85,
	}

	err := repo.UpsertSynergy(ctx, synergy)
	if err != nil {
		t.Fatalf("Failed to upsert synergy: %v", err)
	}
}

func TestMTGZoneRepository_GetSynergiesForCard(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create synergies
	synergies := []*models.MTGZoneSynergy{
		{CardA: "Fatal Push", CardB: "Thoughtseize", Reason: "Synergy 1", Confidence: 0.8},
		{CardA: "Lightning Bolt", CardB: "Fatal Push", Reason: "Synergy 2", Confidence: 0.7},
		{CardA: "Other Card", CardB: "Another Card", Reason: "Synergy 3", Confidence: 0.6},
	}
	for _, s := range synergies {
		_ = repo.UpsertSynergy(ctx, s)
	}

	// Get synergies for Fatal Push (should appear in both card_a and card_b)
	pushSynergies, err := repo.GetSynergiesForCard(ctx, "Fatal Push", 10)
	if err != nil {
		t.Fatalf("Failed to get synergies: %v", err)
	}

	if len(pushSynergies) != 2 {
		t.Errorf("Expected 2 synergies for Fatal Push, got %d", len(pushSynergies))
	}
}

func TestMTGZoneRepository_GetSynergyBetween(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	synergy := &models.MTGZoneSynergy{
		CardA:      "Fatal Push",
		CardB:      "Thoughtseize",
		Reason:     "Both are efficient black spells",
		Confidence: 0.9,
	}
	_ = repo.UpsertSynergy(ctx, synergy)

	// Get synergy between the two cards
	retrieved, err := repo.GetSynergyBetween(ctx, "Fatal Push", "Thoughtseize")
	if err != nil {
		t.Fatalf("Failed to get synergy between cards: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected synergy, got nil")
	}

	if retrieved.Reason != "Both are efficient black spells" {
		t.Errorf("Reason = %q, want %q", retrieved.Reason, "Both are efficient black spells")
	}
}

func TestMTGZoneRepository_GetSynergyBetween_Bidirectional(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	synergy := &models.MTGZoneSynergy{
		CardA:      "Card A",
		CardB:      "Card B",
		Reason:     "Test synergy",
		Confidence: 0.8,
	}
	_ = repo.UpsertSynergy(ctx, synergy)

	// Should find synergy regardless of order
	retrieved1, _ := repo.GetSynergyBetween(ctx, "Card A", "Card B")
	retrieved2, _ := repo.GetSynergyBetween(ctx, "Card B", "Card A")

	if retrieved1 == nil || retrieved2 == nil {
		t.Error("Expected bidirectional synergy lookup to work")
	}
}

func TestMTGZoneRepository_GetSynergiesInArchetype(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	synergies := []*models.MTGZoneSynergy{
		{CardA: "Card A", CardB: "Card B", Reason: "Synergy 1", ArchetypeContext: "Rakdos Midrange", Confidence: 0.8},
		{CardA: "Card C", CardB: "Card D", Reason: "Synergy 2", ArchetypeContext: "Rakdos Midrange", Confidence: 0.7},
		{CardA: "Card E", CardB: "Card F", Reason: "Synergy 3", ArchetypeContext: "Mono-Red Aggro", Confidence: 0.9},
	}
	for _, s := range synergies {
		_ = repo.UpsertSynergy(ctx, s)
	}

	rakdosSynergies, err := repo.GetSynergiesInArchetype(ctx, "Rakdos Midrange")
	if err != nil {
		t.Fatalf("Failed to get synergies in archetype: %v", err)
	}

	if len(rakdosSynergies) != 2 {
		t.Errorf("Expected 2 synergies for Rakdos Midrange, got %d", len(rakdosSynergies))
	}
}

func TestMTGZoneRepository_UpsertArticle(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	article := &models.MTGZoneArticle{
		URL:            "https://mtgazone.com/article/1",
		Title:          "Top Standard Decks",
		ArticleType:    models.ArticleTypeTierList,
		Format:         "Standard",
		CardsMentioned: `["Card A", "Card B"]`,
	}

	err := repo.UpsertArticle(ctx, article)
	if err != nil {
		t.Fatalf("Failed to upsert article: %v", err)
	}
}

func TestMTGZoneRepository_GetArticle(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	article := &models.MTGZoneArticle{
		URL:         "https://mtgazone.com/unique-article",
		Title:       "Unique Article",
		ArticleType: models.ArticleTypeDeckGuide,
	}
	_ = repo.UpsertArticle(ctx, article)

	retrieved, err := repo.GetArticle(ctx, "https://mtgazone.com/unique-article")
	if err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected article, got nil")
	}

	if retrieved.Title != "Unique Article" {
		t.Errorf("Title = %q, want %q", retrieved.Title, "Unique Article")
	}
}

func TestMTGZoneRepository_GetArticlesByFormat(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	articles := []*models.MTGZoneArticle{
		{URL: "https://mtgazone.com/1", Title: "Article 1", Format: "Standard", ArticleType: models.ArticleTypeDeckGuide},
		{URL: "https://mtgazone.com/2", Title: "Article 2", Format: "Standard", ArticleType: models.ArticleTypeTierList},
		{URL: "https://mtgazone.com/3", Title: "Article 3", Format: "Historic", ArticleType: models.ArticleTypeDeckGuide},
	}
	for _, a := range articles {
		_ = repo.UpsertArticle(ctx, a)
	}

	standardArticles, err := repo.GetArticlesByFormat(ctx, "Standard", 10)
	if err != nil {
		t.Fatalf("Failed to get articles by format: %v", err)
	}

	if len(standardArticles) != 2 {
		t.Errorf("Expected 2 Standard articles, got %d", len(standardArticles))
	}
}

func TestMTGZoneRepository_IsArticleProcessed(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Not processed yet
	processed, err := repo.IsArticleProcessed(ctx, "https://mtgazone.com/new")
	if err != nil {
		t.Fatalf("Failed to check article: %v", err)
	}
	if processed {
		t.Error("Expected article to not be processed")
	}

	// Add article
	_ = repo.UpsertArticle(ctx, &models.MTGZoneArticle{URL: "https://mtgazone.com/new", Title: "New"})

	// Now processed
	processed, err = repo.IsArticleProcessed(ctx, "https://mtgazone.com/new")
	if err != nil {
		t.Fatalf("Failed to check article: %v", err)
	}
	if !processed {
		t.Error("Expected article to be processed")
	}
}

func TestMTGZoneRepository_GetArchetypeCount(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Initial count
	count, err := repo.GetArchetypeCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 archetypes, got %d", count)
	}

	// Add archetypes
	_, _ = repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck A", Format: "Standard"})
	_, _ = repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck B", Format: "Standard"})

	count, _ = repo.GetArchetypeCount(ctx)
	if count != 2 {
		t.Errorf("Expected 2 archetypes, got %d", count)
	}
}

func TestMTGZoneRepository_GetSynergyCount(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Initial count
	count, err := repo.GetSynergyCount(ctx)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 synergies, got %d", count)
	}

	// Add synergies
	_ = repo.UpsertSynergy(ctx, &models.MTGZoneSynergy{CardA: "A", CardB: "B", Reason: "Test"})
	_ = repo.UpsertSynergy(ctx, &models.MTGZoneSynergy{CardA: "C", CardB: "D", Reason: "Test"})

	count, _ = repo.GetSynergyCount(ctx)
	if count != 2 {
		t.Errorf("Expected 2 synergies, got %d", count)
	}
}

func TestMTGZoneRepository_ClearAll(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Add data
	id, _ := repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Deck A", Format: "Standard"})
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id, CardName: "Card A", Role: models.CardRoleCore})
	_ = repo.UpsertSynergy(ctx, &models.MTGZoneSynergy{CardA: "A", CardB: "B", Reason: "Test"})
	_ = repo.UpsertArticle(ctx, &models.MTGZoneArticle{URL: "https://test.com", Title: "Test"})

	// Clear all
	err := repo.ClearAll(ctx)
	if err != nil {
		t.Fatalf("Failed to clear all: %v", err)
	}

	// Verify empty
	archetypeCount, _ := repo.GetArchetypeCount(ctx)
	synergyCount, _ := repo.GetSynergyCount(ctx)

	if archetypeCount != 0 {
		t.Errorf("Expected 0 archetypes after clear, got %d", archetypeCount)
	}
	if synergyCount != 0 {
		t.Errorf("Expected 0 synergies after clear, got %d", synergyCount)
	}
}

func TestMTGZoneRepository_DeleteArchetypeCards(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create archetype with cards
	id, _ := repo.UpsertArchetype(ctx, &models.MTGZoneArchetype{Name: "Test Deck", Format: "Standard"})
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id, CardName: "Card A", Role: models.CardRoleCore})
	_ = repo.UpsertArchetypeCard(ctx, &models.MTGZoneArchetypeCard{ArchetypeID: id, CardName: "Card B", Role: models.CardRoleFlex})

	// Delete cards
	err := repo.DeleteArchetypeCards(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete archetype cards: %v", err)
	}

	// Verify deletion
	cards, _ := repo.GetArchetypeCards(ctx, id)
	if len(cards) != 0 {
		t.Errorf("Expected 0 cards after deletion, got %d", len(cards))
	}
}

func TestMTGZoneRepository_DeleteSynergiesForArchetype(t *testing.T) {
	db := setupMTGZoneTestDB(t)
	defer db.Close()

	repo := NewMTGZoneRepository(db)
	ctx := context.Background()

	// Create synergies
	_ = repo.UpsertSynergy(ctx, &models.MTGZoneSynergy{CardA: "A", CardB: "B", Reason: "Test", ArchetypeContext: "Rakdos"})
	_ = repo.UpsertSynergy(ctx, &models.MTGZoneSynergy{CardA: "C", CardB: "D", Reason: "Test", ArchetypeContext: "Rakdos"})
	_ = repo.UpsertSynergy(ctx, &models.MTGZoneSynergy{CardA: "E", CardB: "F", Reason: "Test", ArchetypeContext: "Mono-Red"})

	// Delete Rakdos synergies
	err := repo.DeleteSynergiesForArchetype(ctx, "Rakdos")
	if err != nil {
		t.Fatalf("Failed to delete synergies: %v", err)
	}

	// Verify
	rakdos, _ := repo.GetSynergiesInArchetype(ctx, "Rakdos")
	monoRed, _ := repo.GetSynergiesInArchetype(ctx, "Mono-Red")

	if len(rakdos) != 0 {
		t.Errorf("Expected 0 Rakdos synergies after deletion, got %d", len(rakdos))
	}
	if len(monoRed) != 1 {
		t.Errorf("Expected 1 Mono-Red synergy (not deleted), got %d", len(monoRed))
	}
}

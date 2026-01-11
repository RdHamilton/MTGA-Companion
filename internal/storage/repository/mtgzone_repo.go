package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MTGZoneRepository handles MTGZone archetype and synergy data operations.
type MTGZoneRepository interface {
	// Archetype operations
	UpsertArchetype(ctx context.Context, archetype *models.MTGZoneArchetype) (int64, error)
	GetArchetype(ctx context.Context, name, format string) (*models.MTGZoneArchetype, error)
	GetArchetypeByID(ctx context.Context, id int64) (*models.MTGZoneArchetype, error)
	GetArchetypesByFormat(ctx context.Context, format string) ([]*models.MTGZoneArchetype, error)
	GetTopTierArchetypes(ctx context.Context, format string, limit int) ([]*models.MTGZoneArchetype, error)
	DeleteArchetype(ctx context.Context, id int64) error

	// Archetype card operations
	UpsertArchetypeCard(ctx context.Context, card *models.MTGZoneArchetypeCard) error
	GetArchetypeCards(ctx context.Context, archetypeID int64) ([]*models.MTGZoneArchetypeCard, error)
	GetCoreCards(ctx context.Context, archetypeID int64) ([]*models.MTGZoneArchetypeCard, error)
	GetArchetypesForCard(ctx context.Context, cardName string) ([]*models.MTGZoneArchetype, error)
	DeleteArchetypeCards(ctx context.Context, archetypeID int64) error

	// Synergy operations
	UpsertSynergy(ctx context.Context, synergy *models.MTGZoneSynergy) error
	GetSynergiesForCard(ctx context.Context, cardName string, limit int) ([]*models.MTGZoneSynergy, error)
	GetSynergyBetween(ctx context.Context, cardA, cardB string) (*models.MTGZoneSynergy, error)
	GetSynergiesInArchetype(ctx context.Context, archetype string) ([]*models.MTGZoneSynergy, error)
	DeleteSynergiesForArchetype(ctx context.Context, archetype string) error

	// Article operations
	UpsertArticle(ctx context.Context, article *models.MTGZoneArticle) error
	GetArticle(ctx context.Context, url string) (*models.MTGZoneArticle, error)
	GetArticlesByFormat(ctx context.Context, format string, limit int) ([]*models.MTGZoneArticle, error)
	IsArticleProcessed(ctx context.Context, url string) (bool, error)

	// Utility
	GetArchetypeCount(ctx context.Context) (int, error)
	GetSynergyCount(ctx context.Context) (int, error)
	ClearAll(ctx context.Context) error
}

// mtgzoneRepo implements MTGZoneRepository.
type mtgzoneRepo struct {
	db *sql.DB
}

// NewMTGZoneRepository creates a new MTGZone repository.
func NewMTGZoneRepository(db *sql.DB) MTGZoneRepository {
	return &mtgzoneRepo{db: db}
}

// UpsertArchetype inserts or updates an archetype.
func (r *mtgzoneRepo) UpsertArchetype(ctx context.Context, archetype *models.MTGZoneArchetype) (int64, error) {
	query := `
		INSERT INTO mtgzone_archetypes (name, format, tier, description, play_style, source_url, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(name, format) DO UPDATE SET
			tier = excluded.tier,
			description = excluded.description,
			play_style = excluded.play_style,
			source_url = excluded.source_url,
			last_updated = CURRENT_TIMESTAMP
	`

	result, err := r.db.ExecContext(ctx, query,
		archetype.Name, archetype.Format, archetype.Tier,
		archetype.Description, archetype.PlayStyle, archetype.SourceURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert archetype: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		// On conflict, get the existing ID
		existing, err := r.GetArchetype(ctx, archetype.Name, archetype.Format)
		if err != nil {
			return 0, err
		}
		return existing.ID, nil
	}

	return id, nil
}

// GetArchetype gets an archetype by name and format.
func (r *mtgzoneRepo) GetArchetype(ctx context.Context, name, format string) (*models.MTGZoneArchetype, error) {
	query := `
		SELECT id, name, format, tier, description, play_style, source_url, last_updated
		FROM mtgzone_archetypes
		WHERE LOWER(name) = LOWER(?) AND LOWER(format) = LOWER(?)
	`

	var a models.MTGZoneArchetype
	var tier, description, playStyle, sourceURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, name, format).Scan(
		&a.ID, &a.Name, &a.Format, &tier, &description, &playStyle, &sourceURL, &a.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype: %w", err)
	}

	a.Tier = tier.String
	a.Description = description.String
	a.PlayStyle = playStyle.String
	a.SourceURL = sourceURL.String

	return &a, nil
}

// GetArchetypeByID gets an archetype by ID.
func (r *mtgzoneRepo) GetArchetypeByID(ctx context.Context, id int64) (*models.MTGZoneArchetype, error) {
	query := `
		SELECT id, name, format, tier, description, play_style, source_url, last_updated
		FROM mtgzone_archetypes
		WHERE id = ?
	`

	var a models.MTGZoneArchetype
	var tier, description, playStyle, sourceURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Name, &a.Format, &tier, &description, &playStyle, &sourceURL, &a.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype: %w", err)
	}

	a.Tier = tier.String
	a.Description = description.String
	a.PlayStyle = playStyle.String
	a.SourceURL = sourceURL.String

	return &a, nil
}

// GetArchetypesByFormat gets all archetypes for a format.
func (r *mtgzoneRepo) GetArchetypesByFormat(ctx context.Context, format string) ([]*models.MTGZoneArchetype, error) {
	query := `
		SELECT id, name, format, tier, description, play_style, source_url, last_updated
		FROM mtgzone_archetypes
		WHERE LOWER(format) = LOWER(?)
		ORDER BY tier ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get archetypes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var archetypes []*models.MTGZoneArchetype
	for rows.Next() {
		var a models.MTGZoneArchetype
		var tier, description, playStyle, sourceURL sql.NullString

		if err := rows.Scan(
			&a.ID, &a.Name, &a.Format, &tier, &description, &playStyle, &sourceURL, &a.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan archetype: %w", err)
		}

		a.Tier = tier.String
		a.Description = description.String
		a.PlayStyle = playStyle.String
		a.SourceURL = sourceURL.String

		archetypes = append(archetypes, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating archetypes: %w", err)
	}

	return archetypes, nil
}

// GetTopTierArchetypes gets top tier archetypes for a format.
func (r *mtgzoneRepo) GetTopTierArchetypes(ctx context.Context, format string, limit int) ([]*models.MTGZoneArchetype, error) {
	query := `
		SELECT id, name, format, tier, description, play_style, source_url, last_updated
		FROM mtgzone_archetypes
		WHERE LOWER(format) = LOWER(?) AND tier IN ('S', 'A', 'A+', 'A-')
		ORDER BY tier ASC, name ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, format, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top tier archetypes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var archetypes []*models.MTGZoneArchetype
	for rows.Next() {
		var a models.MTGZoneArchetype
		var tier, description, playStyle, sourceURL sql.NullString

		if err := rows.Scan(
			&a.ID, &a.Name, &a.Format, &tier, &description, &playStyle, &sourceURL, &a.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan archetype: %w", err)
		}

		a.Tier = tier.String
		a.Description = description.String
		a.PlayStyle = playStyle.String
		a.SourceURL = sourceURL.String

		archetypes = append(archetypes, &a)
	}

	return archetypes, nil
}

// DeleteArchetype deletes an archetype by ID.
func (r *mtgzoneRepo) DeleteArchetype(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM mtgzone_archetypes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete archetype: %w", err)
	}
	return nil
}

// UpsertArchetypeCard inserts or updates an archetype card.
func (r *mtgzoneRepo) UpsertArchetypeCard(ctx context.Context, card *models.MTGZoneArchetypeCard) error {
	query := `
		INSERT INTO mtgzone_archetype_cards (archetype_id, card_name, role, copies, importance, notes, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(archetype_id, card_name) DO UPDATE SET
			role = excluded.role,
			copies = excluded.copies,
			importance = excluded.importance,
			notes = excluded.notes,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		card.ArchetypeID, card.CardName, card.Role,
		card.Copies, card.Importance, card.Notes,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert archetype card: %w", err)
	}

	return nil
}

// GetArchetypeCards gets all cards for an archetype.
func (r *mtgzoneRepo) GetArchetypeCards(ctx context.Context, archetypeID int64) ([]*models.MTGZoneArchetypeCard, error) {
	query := `
		SELECT id, archetype_id, card_name, role, copies, importance, notes, last_updated
		FROM mtgzone_archetype_cards
		WHERE archetype_id = ?
		ORDER BY role ASC, copies DESC
	`

	rows, err := r.db.QueryContext(ctx, query, archetypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*models.MTGZoneArchetypeCard
	for rows.Next() {
		var c models.MTGZoneArchetypeCard
		var importance, notes sql.NullString

		if err := rows.Scan(
			&c.ID, &c.ArchetypeID, &c.CardName, &c.Role,
			&c.Copies, &importance, &notes, &c.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan archetype card: %w", err)
		}

		c.Importance = models.CardImportance(importance.String)
		c.Notes = notes.String

		cards = append(cards, &c)
	}

	return cards, nil
}

// GetCoreCards gets core cards for an archetype.
func (r *mtgzoneRepo) GetCoreCards(ctx context.Context, archetypeID int64) ([]*models.MTGZoneArchetypeCard, error) {
	query := `
		SELECT id, archetype_id, card_name, role, copies, importance, notes, last_updated
		FROM mtgzone_archetype_cards
		WHERE archetype_id = ? AND role = 'core'
		ORDER BY copies DESC
	`

	rows, err := r.db.QueryContext(ctx, query, archetypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get core cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*models.MTGZoneArchetypeCard
	for rows.Next() {
		var c models.MTGZoneArchetypeCard
		var importance, notes sql.NullString

		if err := rows.Scan(
			&c.ID, &c.ArchetypeID, &c.CardName, &c.Role,
			&c.Copies, &importance, &notes, &c.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan archetype card: %w", err)
		}

		c.Importance = models.CardImportance(importance.String)
		c.Notes = notes.String

		cards = append(cards, &c)
	}

	return cards, nil
}

// GetArchetypesForCard gets all archetypes that include a specific card.
func (r *mtgzoneRepo) GetArchetypesForCard(ctx context.Context, cardName string) ([]*models.MTGZoneArchetype, error) {
	query := `
		SELECT DISTINCT a.id, a.name, a.format, a.tier, a.description, a.play_style, a.source_url, a.last_updated
		FROM mtgzone_archetypes a
		JOIN mtgzone_archetype_cards ac ON a.id = ac.archetype_id
		WHERE LOWER(ac.card_name) = LOWER(?)
		ORDER BY a.tier ASC, a.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cardName)
	if err != nil {
		return nil, fmt.Errorf("failed to get archetypes for card: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var archetypes []*models.MTGZoneArchetype
	for rows.Next() {
		var a models.MTGZoneArchetype
		var tier, description, playStyle, sourceURL sql.NullString

		if err := rows.Scan(
			&a.ID, &a.Name, &a.Format, &tier, &description, &playStyle, &sourceURL, &a.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan archetype: %w", err)
		}

		a.Tier = tier.String
		a.Description = description.String
		a.PlayStyle = playStyle.String
		a.SourceURL = sourceURL.String

		archetypes = append(archetypes, &a)
	}

	return archetypes, nil
}

// DeleteArchetypeCards deletes all cards for an archetype.
func (r *mtgzoneRepo) DeleteArchetypeCards(ctx context.Context, archetypeID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM mtgzone_archetype_cards WHERE archetype_id = ?", archetypeID)
	if err != nil {
		return fmt.Errorf("failed to delete archetype cards: %w", err)
	}
	return nil
}

// UpsertSynergy inserts or updates a synergy.
func (r *mtgzoneRepo) UpsertSynergy(ctx context.Context, synergy *models.MTGZoneSynergy) error {
	query := `
		INSERT INTO mtgzone_synergies (card_a, card_b, reason, source_url, archetype_context, confidence, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(card_a, card_b, archetype_context) DO UPDATE SET
			reason = excluded.reason,
			source_url = excluded.source_url,
			confidence = excluded.confidence,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query,
		synergy.CardA, synergy.CardB, synergy.Reason,
		synergy.SourceURL, synergy.ArchetypeContext, synergy.Confidence,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert synergy: %w", err)
	}

	return nil
}

// GetSynergiesForCard gets all synergies involving a card.
func (r *mtgzoneRepo) GetSynergiesForCard(ctx context.Context, cardName string, limit int) ([]*models.MTGZoneSynergy, error) {
	query := `
		SELECT id, card_a, card_b, reason, source_url, archetype_context, confidence, last_updated
		FROM mtgzone_synergies
		WHERE LOWER(card_a) = LOWER(?) OR LOWER(card_b) = LOWER(?)
		ORDER BY confidence DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, cardName, cardName)
	if err != nil {
		return nil, fmt.Errorf("failed to get synergies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var synergies []*models.MTGZoneSynergy
	for rows.Next() {
		var s models.MTGZoneSynergy
		var sourceURL, archetypeContext sql.NullString

		if err := rows.Scan(
			&s.ID, &s.CardA, &s.CardB, &s.Reason,
			&sourceURL, &archetypeContext, &s.Confidence, &s.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan synergy: %w", err)
		}

		s.SourceURL = sourceURL.String
		s.ArchetypeContext = archetypeContext.String

		synergies = append(synergies, &s)
	}

	return synergies, nil
}

// GetSynergyBetween gets the synergy between two specific cards.
func (r *mtgzoneRepo) GetSynergyBetween(ctx context.Context, cardA, cardB string) (*models.MTGZoneSynergy, error) {
	query := `
		SELECT id, card_a, card_b, reason, source_url, archetype_context, confidence, last_updated
		FROM mtgzone_synergies
		WHERE (LOWER(card_a) = LOWER(?) AND LOWER(card_b) = LOWER(?))
		   OR (LOWER(card_a) = LOWER(?) AND LOWER(card_b) = LOWER(?))
		ORDER BY confidence DESC
		LIMIT 1
	`

	var s models.MTGZoneSynergy
	var sourceURL, archetypeContext sql.NullString

	err := r.db.QueryRowContext(ctx, query, cardA, cardB, cardB, cardA).Scan(
		&s.ID, &s.CardA, &s.CardB, &s.Reason,
		&sourceURL, &archetypeContext, &s.Confidence, &s.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get synergy: %w", err)
	}

	s.SourceURL = sourceURL.String
	s.ArchetypeContext = archetypeContext.String

	return &s, nil
}

// GetSynergiesInArchetype gets all synergies for an archetype context.
func (r *mtgzoneRepo) GetSynergiesInArchetype(ctx context.Context, archetype string) ([]*models.MTGZoneSynergy, error) {
	query := `
		SELECT id, card_a, card_b, reason, source_url, archetype_context, confidence, last_updated
		FROM mtgzone_synergies
		WHERE LOWER(archetype_context) = LOWER(?)
		ORDER BY confidence DESC
	`

	rows, err := r.db.QueryContext(ctx, query, archetype)
	if err != nil {
		return nil, fmt.Errorf("failed to get synergies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var synergies []*models.MTGZoneSynergy
	for rows.Next() {
		var s models.MTGZoneSynergy
		var sourceURL, archetypeContext sql.NullString

		if err := rows.Scan(
			&s.ID, &s.CardA, &s.CardB, &s.Reason,
			&sourceURL, &archetypeContext, &s.Confidence, &s.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan synergy: %w", err)
		}

		s.SourceURL = sourceURL.String
		s.ArchetypeContext = archetypeContext.String

		synergies = append(synergies, &s)
	}

	return synergies, nil
}

// DeleteSynergiesForArchetype deletes all synergies for an archetype.
func (r *mtgzoneRepo) DeleteSynergiesForArchetype(ctx context.Context, archetype string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM mtgzone_synergies WHERE LOWER(archetype_context) = LOWER(?)", archetype)
	if err != nil {
		return fmt.Errorf("failed to delete synergies: %w", err)
	}
	return nil
}

// UpsertArticle inserts or updates an article.
func (r *mtgzoneRepo) UpsertArticle(ctx context.Context, article *models.MTGZoneArticle) error {
	query := `
		INSERT INTO mtgzone_articles (url, title, article_type, format, archetype, published_at, processed_at, cards_mentioned)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(url) DO UPDATE SET
			title = excluded.title,
			article_type = excluded.article_type,
			format = excluded.format,
			archetype = excluded.archetype,
			published_at = excluded.published_at,
			processed_at = CURRENT_TIMESTAMP,
			cards_mentioned = excluded.cards_mentioned
	`

	_, err := r.db.ExecContext(ctx, query,
		article.URL, article.Title, article.ArticleType,
		article.Format, article.Archetype, article.PublishedAt, article.CardsMentioned,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert article: %w", err)
	}

	return nil
}

// GetArticle gets an article by URL.
func (r *mtgzoneRepo) GetArticle(ctx context.Context, url string) (*models.MTGZoneArticle, error) {
	query := `
		SELECT id, url, title, article_type, format, archetype, published_at, processed_at, cards_mentioned
		FROM mtgzone_articles
		WHERE url = ?
	`

	var a models.MTGZoneArticle
	var articleType, format, archetype, cardsMentioned sql.NullString
	var publishedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, url).Scan(
		&a.ID, &a.URL, &a.Title, &articleType, &format, &archetype,
		&publishedAt, &a.ProcessedAt, &cardsMentioned,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	a.ArticleType = models.ArticleType(articleType.String)
	a.Format = format.String
	a.Archetype = archetype.String
	a.CardsMentioned = cardsMentioned.String
	if publishedAt.Valid {
		a.PublishedAt = &publishedAt.Time
	}

	return &a, nil
}

// GetArticlesByFormat gets articles for a format.
func (r *mtgzoneRepo) GetArticlesByFormat(ctx context.Context, format string, limit int) ([]*models.MTGZoneArticle, error) {
	query := `
		SELECT id, url, title, article_type, format, archetype, published_at, processed_at, cards_mentioned
		FROM mtgzone_articles
		WHERE LOWER(format) = LOWER(?)
		ORDER BY processed_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var articles []*models.MTGZoneArticle
	for rows.Next() {
		var a models.MTGZoneArticle
		var articleType, format, archetype, cardsMentioned sql.NullString
		var publishedAt sql.NullTime

		if err := rows.Scan(
			&a.ID, &a.URL, &a.Title, &articleType, &format, &archetype,
			&publishedAt, &a.ProcessedAt, &cardsMentioned,
		); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		a.ArticleType = models.ArticleType(articleType.String)
		a.Format = format.String
		a.Archetype = archetype.String
		a.CardsMentioned = cardsMentioned.String
		if publishedAt.Valid {
			a.PublishedAt = &publishedAt.Time
		}

		articles = append(articles, &a)
	}

	return articles, nil
}

// IsArticleProcessed checks if an article has been processed.
func (r *mtgzoneRepo) IsArticleProcessed(ctx context.Context, url string) (bool, error) {
	article, err := r.GetArticle(ctx, url)
	if err != nil {
		return false, err
	}
	return article != nil, nil
}

// GetArchetypeCount gets the count of archetypes.
func (r *mtgzoneRepo) GetArchetypeCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mtgzone_archetypes").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get archetype count: %w", err)
	}
	return count, nil
}

// GetSynergyCount gets the count of synergies.
func (r *mtgzoneRepo) GetSynergyCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mtgzone_synergies").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get synergy count: %w", err)
	}
	return count, nil
}

// ClearAll removes all MTGZone data.
func (r *mtgzoneRepo) ClearAll(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tables := []string{"mtgzone_articles", "mtgzone_synergies", "mtgzone_archetype_cards", "mtgzone_archetypes"}
	for _, table := range tables {
		_, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

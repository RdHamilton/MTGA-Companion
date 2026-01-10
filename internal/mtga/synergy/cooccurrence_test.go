package synergy

import (
	"context"
	"fmt"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockCooccurrenceRepo implements CooccurrenceRepository for testing.
type mockCooccurrenceRepo struct {
	cooccurrences  map[string]*models.CardCooccurrence
	frequencies    map[string]*models.CardFrequency
	sources        map[string]*models.CooccurrenceSource
	incrementCalls int
	upsertCalls    int
	pmiUpdateCalls int
}

func newMockCooccurrenceRepo() *mockCooccurrenceRepo {
	return &mockCooccurrenceRepo{
		cooccurrences: make(map[string]*models.CardCooccurrence),
		frequencies:   make(map[string]*models.CardFrequency),
		sources:       make(map[string]*models.CooccurrenceSource),
	}
}

func (m *mockCooccurrenceRepo) makeKey(cardA, cardB int, format string) string {
	if cardA > cardB {
		cardA, cardB = cardB, cardA
	}
	return fmt.Sprintf("%d-%d-%s", cardA, cardB, format)
}

func (m *mockCooccurrenceRepo) UpsertCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string, count int) error {
	m.upsertCalls++
	key := m.makeKey(cardAArenaID, cardBArenaID, format)
	m.cooccurrences[key] = &models.CardCooccurrence{
		CardAArenaID: cardAArenaID,
		CardBArenaID: cardBArenaID,
		Format:       format,
		Count:        count,
	}
	return nil
}

func (m *mockCooccurrenceRepo) IncrementCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string) error {
	m.incrementCalls++
	key := m.makeKey(cardAArenaID, cardBArenaID, format)
	if cooc, ok := m.cooccurrences[key]; ok {
		cooc.Count++
	} else {
		m.cooccurrences[key] = &models.CardCooccurrence{
			CardAArenaID: cardAArenaID,
			CardBArenaID: cardBArenaID,
			Format:       format,
			Count:        1,
		}
	}
	return nil
}

func (m *mockCooccurrenceRepo) GetCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (*models.CardCooccurrence, error) {
	key := m.makeKey(cardAArenaID, cardBArenaID, format)
	return m.cooccurrences[key], nil
}

func (m *mockCooccurrenceRepo) GetTopCooccurrences(ctx context.Context, cardArenaID int, format string, limit int) ([]*models.CardCooccurrence, error) {
	var result []*models.CardCooccurrence
	for _, cooc := range m.cooccurrences {
		if cooc.Format == format && (cooc.CardAArenaID == cardArenaID || cooc.CardBArenaID == cardArenaID) {
			result = append(result, cooc)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockCooccurrenceRepo) UpdatePMIScores(ctx context.Context, format string) error {
	m.pmiUpdateCalls++
	// Simplified PMI calculation for testing
	for key, cooc := range m.cooccurrences {
		if cooc.Format == format {
			// Simple score based on count
			cooc.PMIScore = float64(cooc.Count) * 0.5
			m.cooccurrences[key] = cooc
		}
	}
	return nil
}

func (m *mockCooccurrenceRepo) GetCooccurrenceScore(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (float64, error) {
	cooc, err := m.GetCooccurrence(ctx, cardAArenaID, cardBArenaID, format)
	if err != nil || cooc == nil {
		return 0, err
	}
	return cooc.PMIScore, nil
}

func (m *mockCooccurrenceRepo) UpsertCardFrequency(ctx context.Context, cardArenaID int, format string, deckCount, totalDecks int) error {
	key := fmt.Sprintf("%d-%s", cardArenaID, format)
	m.frequencies[key] = &models.CardFrequency{
		CardArenaID: cardArenaID,
		Format:      format,
		DeckCount:   deckCount,
		TotalDecks:  totalDecks,
		Frequency:   float64(deckCount) / float64(totalDecks),
	}
	return nil
}

func (m *mockCooccurrenceRepo) GetCardFrequency(ctx context.Context, cardArenaID int, format string) (*models.CardFrequency, error) {
	key := fmt.Sprintf("%d-%s", cardArenaID, format)
	return m.frequencies[key], nil
}

func (m *mockCooccurrenceRepo) UpsertSource(ctx context.Context, sourceType, sourceID, format string, deckCount, cardCount int) error {
	key := sourceType + "-" + sourceID + "-" + format
	m.sources[key] = &models.CooccurrenceSource{
		SourceType: sourceType,
		SourceID:   sourceID,
		Format:     format,
		DeckCount:  deckCount,
		CardCount:  cardCount,
	}
	return nil
}

func (m *mockCooccurrenceRepo) GetSource(ctx context.Context, sourceType, sourceID, format string) (*models.CooccurrenceSource, error) {
	key := sourceType + "-" + sourceID + "-" + format
	return m.sources[key], nil
}

func (m *mockCooccurrenceRepo) ClearFormat(ctx context.Context, format string) error {
	for key, cooc := range m.cooccurrences {
		if cooc.Format == format {
			delete(m.cooccurrences, key)
		}
	}
	for key, freq := range m.frequencies {
		if freq.Format == format {
			delete(m.frequencies, key)
		}
	}
	for key, src := range m.sources {
		if src.Format == format {
			delete(m.sources, key)
		}
	}
	return nil
}

// mockCardNameMapper implements CardNameMapper for testing.
type mockCardNameMapper struct {
	nameToID map[string]int
}

func newMockCardNameMapper() *mockCardNameMapper {
	return &mockCardNameMapper{
		nameToID: map[string]int{
			"Card A": 1,
			"Card B": 2,
			"Card C": 3,
			"Card D": 4,
		},
	}
}

func (m *mockCardNameMapper) GetArenaIDByName(ctx context.Context, name string) (int, error) {
	if id, ok := m.nameToID[name]; ok {
		return id, nil
	}
	return 0, nil
}

func TestCooccurrenceAnalyzer_AnalyzeDecks(t *testing.T) {
	repo := newMockCooccurrenceRepo()
	mapper := newMockCardNameMapper()
	analyzer := NewCooccurrenceAnalyzer(repo, mapper)

	decks := []*SimpleDeck{
		{ID: "1", Name: "Deck 1", Format: "Standard", CardNames: []string{"Card A", "Card B", "Card C"}},
		{ID: "2", Name: "Deck 2", Format: "Standard", CardNames: []string{"Card A", "Card B", "Card D"}},
	}

	source := NewLocalDeckSource(decks)

	result, err := analyzer.AnalyzeDecks(context.Background(), source, "Standard", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.DecksAnalyzed != 2 {
		t.Errorf("expected 2 decks analyzed, got %d", result.DecksAnalyzed)
	}

	if result.SourceName != "local" {
		t.Errorf("expected source name 'local', got '%s'", result.SourceName)
	}

	// Deck 1 has 3 cards = 3 pairs: (A,B), (A,C), (B,C)
	// Deck 2 has 3 cards = 3 pairs: (A,B), (A,D), (B,D)
	// Total = 6 pairs
	if result.PairsCreated != 6 {
		t.Errorf("expected 6 pairs created, got %d", result.PairsCreated)
	}

	// Cards A and B appear in both decks, so count should be 2
	if repo.incrementCalls != 6 {
		t.Errorf("expected 6 increment calls, got %d", repo.incrementCalls)
	}

	// PMI should have been calculated
	if repo.pmiUpdateCalls != 1 {
		t.Errorf("expected 1 PMI update call, got %d", repo.pmiUpdateCalls)
	}
}

func TestCooccurrenceAnalyzer_GetSynergyScore(t *testing.T) {
	repo := newMockCooccurrenceRepo()
	mapper := newMockCardNameMapper()
	analyzer := NewCooccurrenceAnalyzer(repo, mapper)

	// Add some test co-occurrence data
	repo.cooccurrences["1-2-Standard"] = &models.CardCooccurrence{
		CardAArenaID: 1,
		CardBArenaID: 2,
		Format:       "Standard",
		Count:        10,
		PMIScore:     2.5, // Positive PMI
	}

	score, err := analyzer.GetSynergyScore(context.Background(), 1, 2, "Standard")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// PMI of 2.5 should normalize to 0.5 (2.5/5.0)
	expectedScore := 0.5
	if score != expectedScore {
		t.Errorf("expected score %f, got %f", expectedScore, score)
	}
}

func TestCooccurrenceAnalyzer_GetSynergyScore_NoPMI(t *testing.T) {
	repo := newMockCooccurrenceRepo()
	mapper := newMockCardNameMapper()
	analyzer := NewCooccurrenceAnalyzer(repo, mapper)

	score, err := analyzer.GetSynergyScore(context.Background(), 1, 2, "Standard")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if score != 0.0 {
		t.Errorf("expected score 0.0 for no data, got %f", score)
	}
}

func TestCooccurrenceAnalyzer_GetTopSynergies(t *testing.T) {
	repo := newMockCooccurrenceRepo()
	mapper := newMockCardNameMapper()
	analyzer := NewCooccurrenceAnalyzer(repo, mapper)

	// Add test data
	repo.cooccurrences["1-2-Standard"] = &models.CardCooccurrence{
		CardAArenaID: 1,
		CardBArenaID: 2,
		Format:       "Standard",
		Count:        10,
		PMIScore:     2.5,
	}
	repo.cooccurrences["1-3-Standard"] = &models.CardCooccurrence{
		CardAArenaID: 1,
		CardBArenaID: 3,
		Format:       "Standard",
		Count:        5,
		PMIScore:     1.0,
	}

	synergies, err := analyzer.GetTopSynergies(context.Background(), 1, "Standard", 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(synergies) != 2 {
		t.Errorf("expected 2 synergies, got %d", len(synergies))
	}
}

func TestNormalizePMI(t *testing.T) {
	tests := []struct {
		pmi      float64
		expected float64
	}{
		{-5.0, 0.0}, // Negative PMI = 0
		{0.0, 0.0},  // Zero PMI = 0
		{1.0, 0.2},  // 1/5 = 0.2
		{2.5, 0.5},  // 2.5/5 = 0.5
		{5.0, 1.0},  // 5/5 = 1.0
		{10.0, 1.0}, // Capped at 1.0
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := normalizePMI(tt.pmi)
			if result != tt.expected {
				t.Errorf("normalizePMI(%f) = %f, want %f", tt.pmi, result, tt.expected)
			}
		})
	}
}

func TestCardSynergy_Fields(t *testing.T) {
	synergy := CardSynergy{
		CardArenaID: 12345,
		Score:       0.75,
		RawPMI:      3.5,
		Count:       50,
	}

	if synergy.CardArenaID != 12345 {
		t.Errorf("expected CardArenaID 12345, got %d", synergy.CardArenaID)
	}
	if synergy.Score != 0.75 {
		t.Errorf("expected Score 0.75, got %f", synergy.Score)
	}
	if synergy.RawPMI != 3.5 {
		t.Errorf("expected RawPMI 3.5, got %f", synergy.RawPMI)
	}
	if synergy.Count != 50 {
		t.Errorf("expected Count 50, got %d", synergy.Count)
	}
}

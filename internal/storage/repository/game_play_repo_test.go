package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupGamePlayTestDB creates an in-memory database with game play related tables.
func setupGamePlayTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE matches (
			id TEXT PRIMARY KEY,
			account_id INTEGER NOT NULL DEFAULT 1,
			event_id TEXT NOT NULL,
			event_name TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			duration_seconds INTEGER,
			player_wins INTEGER NOT NULL DEFAULT 0,
			opponent_wins INTEGER NOT NULL DEFAULT 0,
			player_team_id INTEGER NOT NULL DEFAULT 0,
			deck_id TEXT,
			rank_before TEXT,
			rank_after TEXT,
			format TEXT NOT NULL,
			result TEXT NOT NULL,
			result_reason TEXT,
			opponent_name TEXT,
			opponent_id TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE games (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			match_id TEXT NOT NULL,
			game_number INTEGER NOT NULL,
			result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
			duration_seconds INTEGER,
			result_reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (match_id) REFERENCES matches(id) ON DELETE CASCADE,
			UNIQUE(match_id, game_number)
		);

		CREATE TABLE game_plays (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			match_id TEXT NOT NULL,
			turn_number INTEGER NOT NULL,
			phase TEXT,
			step TEXT,
			player_type TEXT NOT NULL,
			action_type TEXT NOT NULL,
			card_id INTEGER,
			card_name TEXT,
			zone_from TEXT,
			zone_to TEXT,
			timestamp TIMESTAMP NOT NULL,
			sequence_number INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
		);

		CREATE TABLE game_state_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			match_id TEXT NOT NULL,
			turn_number INTEGER NOT NULL,
			active_player TEXT NOT NULL,
			player_life INTEGER,
			opponent_life INTEGER,
			player_cards_in_hand INTEGER,
			opponent_cards_in_hand INTEGER,
			player_lands_in_play INTEGER,
			opponent_lands_in_play INTEGER,
			board_state_json TEXT,
			timestamp TIMESTAMP NOT NULL,
			FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
			UNIQUE(game_id, turn_number)
		);

		CREATE TABLE opponent_cards_observed (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id INTEGER NOT NULL,
			match_id TEXT NOT NULL,
			card_id INTEGER NOT NULL,
			card_name TEXT,
			zone_observed TEXT,
			turn_first_seen INTEGER,
			times_seen INTEGER DEFAULT 1,
			FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
			UNIQUE(game_id, card_id)
		);

		CREATE INDEX idx_game_plays_game_id ON game_plays(game_id);
		CREATE INDEX idx_game_plays_match_id ON game_plays(match_id);
		CREATE INDEX idx_game_plays_turn ON game_plays(game_id, turn_number);
		CREATE INDEX idx_game_snapshots_game_id ON game_state_snapshots(game_id);
		CREATE INDEX idx_game_snapshots_match_id ON game_state_snapshots(match_id);
		CREATE INDEX idx_opponent_cards_game_id ON opponent_cards_observed(game_id);
		CREATE INDEX idx_opponent_cards_match_id ON opponent_cards_observed(match_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// createTestMatch creates a test match and game for testing.
func createTestMatch(t *testing.T, db *sql.DB, matchID string) int {
	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	_, err := db.Exec(`
		INSERT INTO matches (id, account_id, event_id, event_name, timestamp, format, result, created_at)
		VALUES (?, 1, 'Ladder', 'Ranked', ?, 'Standard', 'win', ?)
	`, matchID, now, now)
	if err != nil {
		t.Fatalf("failed to create test match: %v", err)
	}

	result, err := db.Exec(`
		INSERT INTO games (match_id, game_number, result, created_at)
		VALUES (?, 1, 'win', ?)
	`, matchID, now)
	if err != nil {
		t.Fatalf("failed to create test game: %v", err)
	}

	gameID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}
	return int(gameID)
}

func TestGamePlayRepository_CreatePlay(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-001")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	cardID := 12345
	cardName := "Lightning Bolt"
	zoneFrom := "hand"
	zoneTo := "battlefield"

	play := &models.GamePlay{
		GameID:         gameID,
		MatchID:        "match-001",
		TurnNumber:     3,
		Phase:          "Main1",
		Step:           "",
		PlayerType:     "player",
		ActionType:     "play_card",
		CardID:         &cardID,
		CardName:       &cardName,
		ZoneFrom:       &zoneFrom,
		ZoneTo:         &zoneTo,
		Timestamp:      time.Now(),
		SequenceNumber: 1,
	}

	err := repo.CreatePlay(ctx, play)
	if err != nil {
		t.Fatalf("CreatePlay failed: %v", err)
	}

	if play.ID == 0 {
		t.Error("Expected play ID to be set after creation")
	}
}

func TestGamePlayRepository_CreatePlays_Batch(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-002")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	plays := []*models.GamePlay{
		{
			GameID:         gameID,
			MatchID:        "match-002",
			TurnNumber:     1,
			Phase:          "Main1",
			PlayerType:     "player",
			ActionType:     "land_drop",
			Timestamp:      time.Now(),
			SequenceNumber: 1,
		},
		{
			GameID:         gameID,
			MatchID:        "match-002",
			TurnNumber:     1,
			Phase:          "Main1",
			PlayerType:     "player",
			ActionType:     "play_card",
			Timestamp:      time.Now(),
			SequenceNumber: 2,
		},
		{
			GameID:         gameID,
			MatchID:        "match-002",
			TurnNumber:     2,
			Phase:          "Combat",
			Step:           "DeclareAttackers",
			PlayerType:     "opponent",
			ActionType:     "attack",
			Timestamp:      time.Now(),
			SequenceNumber: 3,
		},
	}

	err := repo.CreatePlays(ctx, plays)
	if err != nil {
		t.Fatalf("CreatePlays failed: %v", err)
	}

	for i, play := range plays {
		if play.ID == 0 {
			t.Errorf("Expected play %d ID to be set after creation", i)
		}
	}

	// Verify all plays were created
	retrieved, err := repo.GetPlaysByMatch(ctx, "match-002")
	if err != nil {
		t.Fatalf("GetPlaysByMatch failed: %v", err)
	}
	if len(retrieved) != 3 {
		t.Errorf("Expected 3 plays, got %d", len(retrieved))
	}
}

func TestGamePlayRepository_CreateSnapshot(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-003")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	playerLife := 20
	opponentLife := 18
	playerCards := 7
	opponentCards := 6
	playerLands := 2
	opponentLands := 1
	boardState := `{"player_permanents":[],"opponent_permanents":[]}`

	snapshot := &models.GameStateSnapshot{
		GameID:              gameID,
		MatchID:             "match-003",
		TurnNumber:          1,
		ActivePlayer:        "player",
		PlayerLife:          &playerLife,
		OpponentLife:        &opponentLife,
		PlayerCardsInHand:   &playerCards,
		OpponentCardsInHand: &opponentCards,
		PlayerLandsInPlay:   &playerLands,
		OpponentLandsInPlay: &opponentLands,
		BoardStateJSON:      &boardState,
		Timestamp:           time.Now(),
	}

	err := repo.CreateSnapshot(ctx, snapshot)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	if snapshot.ID == 0 {
		t.Error("Expected snapshot ID to be set after creation")
	}
}

func TestGamePlayRepository_CreateSnapshot_Upsert(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-004")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	playerLife := 20
	playerLife2 := 18

	// Create first snapshot
	snapshot1 := &models.GameStateSnapshot{
		GameID:       gameID,
		MatchID:      "match-004",
		TurnNumber:   1,
		ActivePlayer: "player",
		PlayerLife:   &playerLife,
		Timestamp:    time.Now(),
	}
	err := repo.CreateSnapshot(ctx, snapshot1)
	if err != nil {
		t.Fatalf("CreateSnapshot 1 failed: %v", err)
	}

	// Create second snapshot for same turn (should update)
	snapshot2 := &models.GameStateSnapshot{
		GameID:       gameID,
		MatchID:      "match-004",
		TurnNumber:   1,
		ActivePlayer: "player",
		PlayerLife:   &playerLife2,
		Timestamp:    time.Now(),
	}
	err = repo.CreateSnapshot(ctx, snapshot2)
	if err != nil {
		t.Fatalf("CreateSnapshot 2 failed: %v", err)
	}

	// Verify only one snapshot exists for turn 1
	snapshots, err := repo.GetSnapshotsByGame(ctx, gameID)
	if err != nil {
		t.Fatalf("GetSnapshotsByGame failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(snapshots))
	}
	if snapshots[0].PlayerLife != nil && *snapshots[0].PlayerLife != 18 {
		t.Errorf("Expected player life 18 (updated), got %d", *snapshots[0].PlayerLife)
	}
}

func TestGamePlayRepository_RecordOpponentCard(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-005")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	cardName := "Counterspell"
	card := &models.OpponentCardObserved{
		GameID:        gameID,
		MatchID:       "match-005",
		CardID:        54321,
		CardName:      &cardName,
		ZoneObserved:  "hand",
		TurnFirstSeen: 2,
		TimesSeen:     1,
	}

	err := repo.RecordOpponentCard(ctx, card)
	if err != nil {
		t.Fatalf("RecordOpponentCard failed: %v", err)
	}

	// Record same card again (should increment times_seen)
	err = repo.RecordOpponentCard(ctx, card)
	if err != nil {
		t.Fatalf("RecordOpponentCard second time failed: %v", err)
	}

	// Verify card was recorded with incremented count
	cards, err := repo.GetOpponentCardsByGame(ctx, gameID)
	if err != nil {
		t.Fatalf("GetOpponentCardsByGame failed: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("Expected 1 card, got %d", len(cards))
	}
	if cards[0].TimesSeen != 2 {
		t.Errorf("Expected times_seen 2, got %d", cards[0].TimesSeen)
	}
}

func TestGamePlayRepository_GetPlaysByMatch(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-006")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	// Create some plays
	for i := 1; i <= 5; i++ {
		play := &models.GamePlay{
			GameID:         gameID,
			MatchID:        "match-006",
			TurnNumber:     i,
			Phase:          "Main1",
			PlayerType:     "player",
			ActionType:     "play_card",
			Timestamp:      time.Now(),
			SequenceNumber: i,
		}
		if err := repo.CreatePlay(ctx, play); err != nil {
			t.Fatalf("CreatePlay %d failed: %v", i, err)
		}
	}

	plays, err := repo.GetPlaysByMatch(ctx, "match-006")
	if err != nil {
		t.Fatalf("GetPlaysByMatch failed: %v", err)
	}

	if len(plays) != 5 {
		t.Errorf("Expected 5 plays, got %d", len(plays))
	}

	// Verify order by sequence number
	for i, play := range plays {
		if play.SequenceNumber != i+1 {
			t.Errorf("Expected sequence number %d, got %d", i+1, play.SequenceNumber)
		}
	}
}

func TestGamePlayRepository_GetPlaysByGame(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID1 := createTestMatch(t, db, "match-007")

	// Create a second game for the same match
	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	result, err := db.Exec(`INSERT INTO games (match_id, game_number, result, created_at) VALUES (?, 2, 'loss', ?)`, "match-007", now)
	if err != nil {
		t.Fatalf("failed to create second test game: %v", err)
	}
	gameID2, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id for second game: %v", err)
	}

	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	// Create plays for game 1
	for i := 1; i <= 3; i++ {
		play := &models.GamePlay{
			GameID:         gameID1,
			MatchID:        "match-007",
			TurnNumber:     i,
			Phase:          "Main1",
			PlayerType:     "player",
			ActionType:     "play_card",
			Timestamp:      time.Now(),
			SequenceNumber: i,
		}
		if err := repo.CreatePlay(ctx, play); err != nil {
			t.Fatalf("CreatePlay game1 %d failed: %v", i, err)
		}
	}

	// Create plays for game 2
	for i := 1; i <= 2; i++ {
		play := &models.GamePlay{
			GameID:         int(gameID2),
			MatchID:        "match-007",
			TurnNumber:     i,
			Phase:          "Main1",
			PlayerType:     "player",
			ActionType:     "play_card",
			Timestamp:      time.Now(),
			SequenceNumber: i + 10,
		}
		if err := repo.CreatePlay(ctx, play); err != nil {
			t.Fatalf("CreatePlay game2 %d failed: %v", i, err)
		}
	}

	// Get plays for game 1 only
	plays, err := repo.GetPlaysByGame(ctx, gameID1)
	if err != nil {
		t.Fatalf("GetPlaysByGame failed: %v", err)
	}

	if len(plays) != 3 {
		t.Errorf("Expected 3 plays for game 1, got %d", len(plays))
	}
}

func TestGamePlayRepository_GetPlayTimeline(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-008")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	playerLife := 20
	opponentLife := 20

	// Create snapshots for turns 1 and 2
	for turn := 1; turn <= 2; turn++ {
		snapshot := &models.GameStateSnapshot{
			GameID:       gameID,
			MatchID:      "match-008",
			TurnNumber:   turn,
			ActivePlayer: "player",
			PlayerLife:   &playerLife,
			OpponentLife: &opponentLife,
			Timestamp:    time.Now(),
		}
		if err := repo.CreateSnapshot(ctx, snapshot); err != nil {
			t.Fatalf("CreateSnapshot turn %d failed: %v", turn, err)
		}
	}

	// Create plays for different turns and phases
	plays := []*models.GamePlay{
		{GameID: gameID, MatchID: "match-008", TurnNumber: 1, Phase: "Main1", PlayerType: "player", ActionType: "land_drop", Timestamp: time.Now(), SequenceNumber: 1},
		{GameID: gameID, MatchID: "match-008", TurnNumber: 1, Phase: "Main1", PlayerType: "player", ActionType: "play_card", Timestamp: time.Now(), SequenceNumber: 2},
		{GameID: gameID, MatchID: "match-008", TurnNumber: 1, Phase: "Combat", PlayerType: "player", ActionType: "attack", Timestamp: time.Now(), SequenceNumber: 3},
		{GameID: gameID, MatchID: "match-008", TurnNumber: 2, Phase: "Main1", PlayerType: "opponent", ActionType: "land_drop", Timestamp: time.Now(), SequenceNumber: 4},
	}

	if err := repo.CreatePlays(ctx, plays); err != nil {
		t.Fatalf("CreatePlays failed: %v", err)
	}

	timeline, err := repo.GetPlayTimeline(ctx, "match-008")
	if err != nil {
		t.Fatalf("GetPlayTimeline failed: %v", err)
	}

	// Should have 3 entries: Turn 1 Main1, Turn 1 Combat, Turn 2 Main1
	if len(timeline) != 3 {
		t.Errorf("Expected 3 timeline entries, got %d", len(timeline))
	}

	// First entry should be Turn 1 Main1 with 2 plays
	if timeline[0].Turn != 1 || timeline[0].Phase != "Main1" {
		t.Errorf("Expected Turn 1 Main1, got Turn %d %s", timeline[0].Turn, timeline[0].Phase)
	}
	if len(timeline[0].Plays) != 2 {
		t.Errorf("Expected 2 plays in Turn 1 Main1, got %d", len(timeline[0].Plays))
	}

	// First entry should have snapshot
	if timeline[0].Snapshot == nil {
		t.Error("Expected snapshot for Turn 1")
	}
}

func TestGamePlayRepository_GetPlaySummary(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-009")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	// Create varied plays
	plays := []*models.GamePlay{
		{GameID: gameID, MatchID: "match-009", TurnNumber: 1, Phase: "Main1", PlayerType: "player", ActionType: "land_drop", Timestamp: time.Now(), SequenceNumber: 1},
		{GameID: gameID, MatchID: "match-009", TurnNumber: 1, Phase: "Main1", PlayerType: "player", ActionType: "play_card", Timestamp: time.Now(), SequenceNumber: 2},
		{GameID: gameID, MatchID: "match-009", TurnNumber: 2, Phase: "Combat", PlayerType: "player", ActionType: "attack", Timestamp: time.Now(), SequenceNumber: 3},
		{GameID: gameID, MatchID: "match-009", TurnNumber: 2, Phase: "Combat", PlayerType: "opponent", ActionType: "block", Timestamp: time.Now(), SequenceNumber: 4},
		{GameID: gameID, MatchID: "match-009", TurnNumber: 3, Phase: "Main1", PlayerType: "opponent", ActionType: "land_drop", Timestamp: time.Now(), SequenceNumber: 5},
	}

	if err := repo.CreatePlays(ctx, plays); err != nil {
		t.Fatalf("CreatePlays failed: %v", err)
	}

	// Record an opponent card
	cardName := "Test Card"
	if err := repo.RecordOpponentCard(ctx, &models.OpponentCardObserved{
		GameID:        gameID,
		MatchID:       "match-009",
		CardID:        100,
		CardName:      &cardName,
		ZoneObserved:  "battlefield",
		TurnFirstSeen: 1,
		TimesSeen:     1,
	}); err != nil {
		t.Fatalf("RecordOpponentCard failed: %v", err)
	}

	summary, err := repo.GetPlaySummary(ctx, "match-009")
	if err != nil {
		t.Fatalf("GetPlaySummary failed: %v", err)
	}

	if summary == nil {
		t.Fatal("Expected summary to be non-nil")
	}

	if summary.TotalPlays != 5 {
		t.Errorf("Expected 5 total plays, got %d", summary.TotalPlays)
	}
	if summary.PlayerPlays != 3 {
		t.Errorf("Expected 3 player plays, got %d", summary.PlayerPlays)
	}
	if summary.OpponentPlays != 2 {
		t.Errorf("Expected 2 opponent plays, got %d", summary.OpponentPlays)
	}
	if summary.LandDrops != 2 {
		t.Errorf("Expected 2 land drops, got %d", summary.LandDrops)
	}
	if summary.Attacks != 1 {
		t.Errorf("Expected 1 attack, got %d", summary.Attacks)
	}
	if summary.Blocks != 1 {
		t.Errorf("Expected 1 block, got %d", summary.Blocks)
	}
	if summary.TotalTurns != 3 {
		t.Errorf("Expected 3 total turns, got %d", summary.TotalTurns)
	}
	if summary.OpponentCardsSeen != 1 {
		t.Errorf("Expected 1 opponent card seen, got %d", summary.OpponentCardsSeen)
	}
}

func TestGamePlayRepository_DeletePlaysByMatch(t *testing.T) {
	db := setupGamePlayTestDB(t)
	defer func() { _ = db.Close() }()

	gameID := createTestMatch(t, db, "match-010")
	repo := NewGamePlayRepository(db)
	ctx := context.Background()

	// Create plays, snapshot, and opponent card
	play := &models.GamePlay{
		GameID:         gameID,
		MatchID:        "match-010",
		TurnNumber:     1,
		Phase:          "Main1",
		PlayerType:     "player",
		ActionType:     "play_card",
		Timestamp:      time.Now(),
		SequenceNumber: 1,
	}
	if err := repo.CreatePlay(ctx, play); err != nil {
		t.Fatalf("CreatePlay failed: %v", err)
	}

	playerLife := 20
	snapshot := &models.GameStateSnapshot{
		GameID:       gameID,
		MatchID:      "match-010",
		TurnNumber:   1,
		ActivePlayer: "player",
		PlayerLife:   &playerLife,
		Timestamp:    time.Now(),
	}
	if err := repo.CreateSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	if err := repo.RecordOpponentCard(ctx, &models.OpponentCardObserved{
		GameID:        gameID,
		MatchID:       "match-010",
		CardID:        100,
		ZoneObserved:  "hand",
		TurnFirstSeen: 1,
		TimesSeen:     1,
	}); err != nil {
		t.Fatalf("RecordOpponentCard failed: %v", err)
	}

	// Delete all data for match
	if err := repo.DeletePlaysByMatch(ctx, "match-010"); err != nil {
		t.Fatalf("DeletePlaysByMatch failed: %v", err)
	}

	// Verify everything is deleted
	plays, _ := repo.GetPlaysByMatch(ctx, "match-010")
	if len(plays) != 0 {
		t.Errorf("Expected 0 plays after deletion, got %d", len(plays))
	}

	snapshots, _ := repo.GetSnapshotsByMatch(ctx, "match-010")
	if len(snapshots) != 0 {
		t.Errorf("Expected 0 snapshots after deletion, got %d", len(snapshots))
	}

	cards, _ := repo.GetOpponentCardsByMatch(ctx, "match-010")
	if len(cards) != 0 {
		t.Errorf("Expected 0 opponent cards after deletion, got %d", len(cards))
	}
}

func TestParseBoardState(t *testing.T) {
	tests := []struct {
		name    string
		json    *string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "nil input",
			json:    nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "empty string",
			json:    strPtr(""),
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "valid empty board",
			json:    strPtr(`{"player_permanents":[],"opponent_permanents":[]}`),
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "valid board with permanents",
			json:    strPtr(`{"player_permanents":[{"card_id":100,"card_name":"Plains","is_tapped":false}],"opponent_permanents":[{"card_id":200,"card_name":"Island","is_tapped":true}]}`),
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    strPtr(`{invalid json}`),
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBoardState(tt.json)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.wantNil && result != nil {
				t.Error("Expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}

// strPtr is a helper to get a pointer to a string.
func strPtr(s string) *string {
	return &s
}

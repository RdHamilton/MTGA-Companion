package storage

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestSaveDraftPick(t *testing.T) {
	ctx := context.Background()
	service := setupTestService(t)

	// Create a draft event first
	draftEvent := &models.DraftEvent{
		ID:        "draft-test-1",
		AccountID: 1,
		EventName: "Premier Draft BLB",
		SetCode:   "BLB",
		StartTime: time.Now(),
		Wins:      0,
		Losses:    0,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	err := service.SaveDraftEvent(ctx, draftEvent)
	if err != nil {
		t.Fatalf("Failed to save draft event: %v", err)
	}

	// Create a draft pick
	pick := &models.DraftPick{
		DraftEventID:   "draft-test-1",
		PackNumber:     1,
		PickNumber:     1,
		AvailableCards: []int{100, 200, 300, 400, 500},
		SelectedCard:   200,
		Timestamp:      time.Now(),
	}

	// Save the pick
	err = service.SaveDraftPick(ctx, pick)
	if err != nil {
		t.Fatalf("Failed to save draft pick: %v", err)
	}

	// Retrieve the pick
	retrieved, err := service.GetDraftPickByNumber(ctx, "draft-test-1", 1, 1)
	if err != nil {
		t.Fatalf("Failed to get draft pick: %v", err)
	}

	// Verify the pick
	if retrieved.DraftEventID != pick.DraftEventID {
		t.Errorf("Expected DraftEventID %s, got %s", pick.DraftEventID, retrieved.DraftEventID)
	}
	if retrieved.PackNumber != pick.PackNumber {
		t.Errorf("Expected PackNumber %d, got %d", pick.PackNumber, retrieved.PackNumber)
	}
	if retrieved.PickNumber != pick.PickNumber {
		t.Errorf("Expected PickNumber %d, got %d", pick.PickNumber, retrieved.PickNumber)
	}
	if retrieved.SelectedCard != pick.SelectedCard {
		t.Errorf("Expected SelectedCard %d, got %d", pick.SelectedCard, retrieved.SelectedCard)
	}
	if len(retrieved.AvailableCards) != len(pick.AvailableCards) {
		t.Errorf("Expected %d available cards, got %d", len(pick.AvailableCards), len(retrieved.AvailableCards))
	}
}

func TestSaveDraftPicks(t *testing.T) {
	ctx := context.Background()
	service := setupTestService(t)

	// Create a draft event
	draftEvent := &models.DraftEvent{
		ID:        "draft-test-2",
		AccountID: 1,
		EventName: "Premier Draft WOE",
		SetCode:   "WOE",
		StartTime: time.Now(),
		Wins:      0,
		Losses:    0,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	err := service.SaveDraftEvent(ctx, draftEvent)
	if err != nil {
		t.Fatalf("Failed to save draft event: %v", err)
	}

	// Create multiple picks
	picks := []*models.DraftPick{
		{
			DraftEventID:   "draft-test-2",
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200, 300},
			SelectedCard:   200,
			Timestamp:      time.Now(),
		},
		{
			DraftEventID:   "draft-test-2",
			PackNumber:     1,
			PickNumber:     2,
			AvailableCards: []int{100, 300, 400},
			SelectedCard:   300,
			Timestamp:      time.Now(),
		},
		{
			DraftEventID:   "draft-test-2",
			PackNumber:     1,
			PickNumber:     3,
			AvailableCards: []int{100, 400, 500},
			SelectedCard:   500,
			Timestamp:      time.Now(),
		},
	}

	// Save all picks
	err = service.SaveDraftPicks(ctx, "draft-test-2", picks)
	if err != nil {
		t.Fatalf("Failed to save draft picks: %v", err)
	}

	// Retrieve all picks
	retrieved, err := service.GetDraftPicks(ctx, "draft-test-2")
	if err != nil {
		t.Fatalf("Failed to get draft picks: %v", err)
	}

	// Verify count
	if len(retrieved) != len(picks) {
		t.Errorf("Expected %d picks, got %d", len(picks), len(retrieved))
	}

	// Verify order (should be by pack then pick number)
	for i, pick := range retrieved {
		if pick.PackNumber != picks[i].PackNumber {
			t.Errorf("Pick %d: Expected PackNumber %d, got %d", i, picks[i].PackNumber, pick.PackNumber)
		}
		if pick.PickNumber != picks[i].PickNumber {
			t.Errorf("Pick %d: Expected PickNumber %d, got %d", i, picks[i].PickNumber, pick.PickNumber)
		}
	}
}

func TestGetDraftPicksCount(t *testing.T) {
	ctx := context.Background()
	service := setupTestService(t)

	// Create a draft event
	draftEvent := &models.DraftEvent{
		ID:        "draft-test-3",
		AccountID: 1,
		EventName: "Quick Draft LCI",
		SetCode:   "LCI",
		StartTime: time.Now(),
		Wins:      0,
		Losses:    0,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	err := service.SaveDraftEvent(ctx, draftEvent)
	if err != nil {
		t.Fatalf("Failed to save draft event: %v", err)
	}

	// Initially should have 0 picks
	count, err := service.GetDraftPicksCount(ctx, "draft-test-3")
	if err != nil {
		t.Fatalf("Failed to get picks count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 picks, got %d", count)
	}

	// Add some picks
	picks := []*models.DraftPick{
		{
			DraftEventID:   "draft-test-3",
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200},
			SelectedCard:   100,
			Timestamp:      time.Now(),
		},
		{
			DraftEventID:   "draft-test-3",
			PackNumber:     1,
			PickNumber:     2,
			AvailableCards: []int{200, 300},
			SelectedCard:   200,
			Timestamp:      time.Now(),
		},
	}

	err = service.SaveDraftPicks(ctx, "draft-test-3", picks)
	if err != nil {
		t.Fatalf("Failed to save draft picks: %v", err)
	}

	// Should now have 2 picks
	count, err = service.GetDraftPicksCount(ctx, "draft-test-3")
	if err != nil {
		t.Fatalf("Failed to get picks count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 picks, got %d", count)
	}
}

func TestDeleteDraftPicks(t *testing.T) {
	ctx := context.Background()
	service := setupTestService(t)

	// Create a draft event
	draftEvent := &models.DraftEvent{
		ID:        "draft-test-4",
		AccountID: 1,
		EventName: "Premier Draft MKM",
		SetCode:   "MKM",
		StartTime: time.Now(),
		Wins:      0,
		Losses:    0,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	err := service.SaveDraftEvent(ctx, draftEvent)
	if err != nil {
		t.Fatalf("Failed to save draft event: %v", err)
	}

	// Add picks
	picks := []*models.DraftPick{
		{
			DraftEventID:   "draft-test-4",
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200},
			SelectedCard:   100,
			Timestamp:      time.Now(),
		},
	}

	err = service.SaveDraftPicks(ctx, "draft-test-4", picks)
	if err != nil {
		t.Fatalf("Failed to save draft picks: %v", err)
	}

	// Verify picks exist
	count, err := service.GetDraftPicksCount(ctx, "draft-test-4")
	if err != nil {
		t.Fatalf("Failed to get picks count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 pick before delete, got %d", count)
	}

	// Delete picks
	err = service.DeleteDraftPicks(ctx, "draft-test-4")
	if err != nil {
		t.Fatalf("Failed to delete draft picks: %v", err)
	}

	// Verify picks are gone
	count, err = service.GetDraftPicksCount(ctx, "draft-test-4")
	if err != nil {
		t.Fatalf("Failed to get picks count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 picks after delete, got %d", count)
	}
}

func TestGetAllDraftEventsWithPicks(t *testing.T) {
	ctx := context.Background()
	service := setupTestService(t)

	// Create multiple draft events
	events := []*models.DraftEvent{
		{
			ID:        "draft-test-5",
			AccountID: 1,
			EventName: "Premier Draft A",
			SetCode:   "AAA",
			StartTime: time.Now(),
			Wins:      0,
			Losses:    0,
			Status:    "active",
			CreatedAt: time.Now(),
		},
		{
			ID:        "draft-test-6",
			AccountID: 1,
			EventName: "Premier Draft B",
			SetCode:   "BBB",
			StartTime: time.Now(),
			Wins:      0,
			Losses:    0,
			Status:    "active",
			CreatedAt: time.Now(),
		},
	}

	for _, event := range events {
		err := service.SaveDraftEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to save draft event: %v", err)
		}
	}

	// Add picks only to first event
	picks := []*models.DraftPick{
		{
			DraftEventID:   "draft-test-5",
			PackNumber:     1,
			PickNumber:     1,
			AvailableCards: []int{100, 200},
			SelectedCard:   100,
			Timestamp:      time.Now(),
		},
	}

	err := service.SaveDraftPicks(ctx, "draft-test-5", picks)
	if err != nil {
		t.Fatalf("Failed to save draft picks: %v", err)
	}

	// Get all events with picks
	eventIDs, err := service.GetAllDraftEventsWithPicks(ctx)
	if err != nil {
		t.Fatalf("Failed to get draft events with picks: %v", err)
	}

	// Should only have one event with picks
	if len(eventIDs) != 1 {
		t.Errorf("Expected 1 event with picks, got %d", len(eventIDs))
	}
	if len(eventIDs) > 0 && eventIDs[0] != "draft-test-5" {
		t.Errorf("Expected event ID 'draft-test-5', got '%s'", eventIDs[0])
	}
}

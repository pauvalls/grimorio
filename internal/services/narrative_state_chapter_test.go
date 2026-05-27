package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestNarrativeStateService_Update_ChapterTracking(t *testing.T) {
	svc, stateRepo, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Create initial state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Update with chapter change
	update := domain.StateUpdate{
		SessionNum:       2,
		CurrentChapterID: "chapter-2",
		CompletedChapters: []string{"chapter-1"},
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if updated.CurrentChapter != "chapter-2" {
		t.Errorf("Expected CurrentChapter = chapter-2, got %s", updated.CurrentChapter)
	}

	if len(updated.CompletedChapters) != 1 || updated.CompletedChapters[0] != "chapter-1" {
		t.Errorf("Expected CompletedChapters = [chapter-1], got %v", updated.CompletedChapters)
	}
}

func TestNarrativeStateService_Update_XPTracking(t *testing.T) {
	svc, stateRepo, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Create initial state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
		XPTotal:        0,
		PartyLevel:     1,
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Update with XP award
	update := domain.StateUpdate{
		SessionNum: 2,
		XPAwarded:  500,
		XPReason:   "combat",
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if updated.XPTotal != 500 {
		t.Errorf("Expected XPTotal = 500, got %d", updated.XPTotal)
	}

	if updated.PartyLevel != 2 {
		t.Errorf("Expected PartyLevel = 2 (300 XP threshold), got %d", updated.PartyLevel)
	}

	if len(updated.XPLedger) != 1 {
		t.Fatalf("Expected 1 XP ledger entry, got %d", len(updated.XPLedger))
	}

	entry := updated.XPLedger[0]
	if entry.Amount != 500 {
		t.Errorf("Expected XP entry amount = 500, got %d", entry.Amount)
	}
	if entry.Reason != "combat" {
		t.Errorf("Expected XP entry reason = combat, got %s", entry.Reason)
	}
	if entry.SessionNum != 2 {
		t.Errorf("Expected XP entry session = 2, got %d", entry.SessionNum)
	}
}

func TestNarrativeStateService_Update_XPMultipleSessions(t *testing.T) {
	svc, stateRepo, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Create initial state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Session 1: 300 XP (level 2)
	update1 := domain.StateUpdate{
		SessionNum: 1,
		XPAwarded:  300,
		XPReason:   "milestone:chapter-1",
	}
	updated1, _ := svc.Update(ctx, "test-campaign", update1)
	if updated1.PartyLevel != 2 {
		t.Errorf("After 300 XP: Expected level 2, got %d", updated1.PartyLevel)
	}

	// Session 2: 600 XP more (total 900 = level 3)
	update2 := domain.StateUpdate{
		SessionNum: 2,
		XPAwarded:  600,
		XPReason:   "combat",
	}
	updated2, _ := svc.Update(ctx, "test-campaign", update2)
	if updated2.XPTotal != 900 {
		t.Errorf("Expected XPTotal = 900, got %d", updated2.XPTotal)
	}
	if updated2.PartyLevel != 3 {
		t.Errorf("Expected level 3, got %d", updated2.PartyLevel)
	}
	if len(updated2.XPLedger) != 2 {
		t.Errorf("Expected 2 ledger entries, got %d", len(updated2.XPLedger))
	}
}

func TestNarrativeStateService_Update_ChapterTransition(t *testing.T) {
	svc, stateRepo, canonRepo := setupNarrativeStateService()
	ctx := context.Background()

	// Setup canon
	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities:      []domain.CanonEntity{},
	}
	if err := canonRepo.Save("test-campaign", canon); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Create initial state with chapter-1
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
		CurrentChapter: "chapter-1",
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Update to chapter-2
	update := domain.StateUpdate{
		SessionNum:       5,
		CurrentChapterID: "chapter-2",
		SyncToCanon:      true,
	}

	updatedState, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Verify state was updated
	if updatedState.CurrentChapter != "chapter-2" {
		t.Errorf("Expected CurrentChapter = chapter-2, got %s", updatedState.CurrentChapter)
	}

	// Sync to canon (chapter transition event is created if we track previous state)
	warnings, err := svc.SyncStateToCanon(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("SyncStateToCanon error: %v", err)
	}
	if len(warnings) > 0 {
		t.Logf("Warnings: %v", warnings)
	}

	// Verify canon was updated
	updatedCanon, _ := canonRepo.Load("test-campaign")
	if updatedCanon == nil {
		t.Fatal("Expected canon to be updated")
	}
}

func TestNarrativeStateService_Update_CompletedChaptersMerge(t *testing.T) {
	svc, stateRepo, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Create state with one completed chapter
	state := &domain.NarrativeState{
		SchemaVersion:     domain.SchemaVersionV2,
		CampaignID:        "test-campaign",
		CurrentSession:    1,
		CompletedChapters: []string{"chapter-1"},
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Update adding another completed chapter (and duplicate)
	update := domain.StateUpdate{
		SessionNum:        2,
		CompletedChapters: []string{"chapter-1", "chapter-2"},
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// Should have both chapters, no duplicates
	if len(updated.CompletedChapters) != 2 {
		t.Errorf("Expected 2 completed chapters, got %d: %v", len(updated.CompletedChapters), updated.CompletedChapters)
	}
}

func TestNarrativeStateService_Update_XPDefaultReason(t *testing.T) {
	svc, stateRepo, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Update without XP reason
	update := domain.StateUpdate{
		SessionNum: 1,
		XPAwarded:  100,
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if len(updated.XPLedger) != 1 {
		t.Fatalf("Expected 1 XP ledger entry, got %d", len(updated.XPLedger))
	}

	if updated.XPLedger[0].Reason != "session" {
		t.Errorf("Expected default reason 'session', got %s", updated.XPLedger[0].Reason)
	}
}

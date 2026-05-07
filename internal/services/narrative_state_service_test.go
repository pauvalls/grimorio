package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func setupNarrativeStateService() (*NarrativeStateService, *repository.MemoryNarrativeStateRepository, *repository.MemoryCanonRepository) {
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	svc := NewNarrativeStateService(stateRepo, canonRepo)
	return svc, stateRepo, canonRepo
}

func TestNarrativeStateService_LoadSave(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
	}

	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	loaded, err := svc.Load(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.CampaignID != "test-campaign" {
		t.Fatalf("expected campaign ID test-campaign, got %s", loaded.CampaignID)
	}
	if loaded.CurrentSession != 1 {
		t.Fatalf("expected session 1, got %d", loaded.CurrentSession)
	}
}

func TestNarrativeStateService_Update(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Initialize state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
		ActiveQuests: []domain.QuestState{
			{ID: "q-001", Name: "Find the Sword", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver"},
		},
	}
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum: 2,
		RevealedClues: []domain.RevealedClue{
			{ID: "clue-001", Description: "The password is 'swordfish'", SourceAct: "act-1", SessionRevealed: 2},
		},
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-villain", Name: "Lord Dark", Session: 2, Cause: "combat"},
		},
		CompletedQuests: []string{"q-001"},
		NewQuests: []domain.QuestState{
			{ID: "q-002", Name: "Escape the dungeon", Status: "active", SourceAct: "act-2", GiverNPC: "npc-ally"},
		},
		KeyItems: []domain.KeyItem{
			{ID: "item-001", Name: "Magic Key", Holder: "party", SessionFound: 2},
		},
		KeyDecisions: []domain.Decision{
			{ID: "dec-001", Description: "Spare the villain", ChoiceMade: "spare", ImpactScope: "story"},
		},
		XPAwarded:        500,
		LootAcquired:     []string{"gold-100", "potion-2"},
		SessionSummary:   "The party defeated Lord Dark.",
		DMNotes:          "Player hesitation noted.",
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	// Verify session incremented
	if updated.CurrentSession != 2 {
		t.Fatalf("expected current session 2, got %d", updated.CurrentSession)
	}

	// Verify revealed clue appended
	if len(updated.RevealedClues) != 1 {
		t.Fatalf("expected 1 revealed clue, got %d", len(updated.RevealedClues))
	}
	if updated.RevealedClues[0].ID != "clue-001" {
		t.Fatalf("expected clue clue-001, got %s", updated.RevealedClues[0].ID)
	}

	// Verify quest moved to completed
	if len(updated.ActiveQuests) != 1 {
		t.Fatalf("expected 1 active quest, got %d", len(updated.ActiveQuests))
	}
	if updated.ActiveQuests[0].ID != "q-002" {
		t.Fatalf("expected active quest q-002, got %s", updated.ActiveQuests[0].ID)
	}
	if len(updated.CompletedQuests) != 1 {
		t.Fatalf("expected 1 completed quest, got %d", len(updated.CompletedQuests))
	}
	if updated.CompletedQuests[0].ID != "q-001" {
		t.Fatalf("expected completed quest q-001, got %s", updated.CompletedQuests[0].ID)
	}
	if updated.CompletedQuests[0].Status != "completed" {
		t.Fatalf("expected completed status, got %s", updated.CompletedQuests[0].Status)
	}

	// Verify dead NPCs
	if len(updated.DeadNPCs) != 1 {
		t.Fatalf("expected 1 dead NPC, got %d", len(updated.DeadNPCs))
	}
	if updated.DeadNPCs[0].NPCID != "npc-villain" {
		t.Fatalf("expected dead npc-villain, got %s", updated.DeadNPCs[0].NPCID)
	}

	// Verify key items
	if len(updated.KeyItems) != 1 {
		t.Fatalf("expected 1 key item, got %d", len(updated.KeyItems))
	}

	// Verify session log
	if len(updated.SessionLog) != 1 {
		t.Fatalf("expected 1 session log entry, got %d", len(updated.SessionLog))
	}
	log := updated.SessionLog[0]
	if log.SessionNum != 2 {
		t.Fatalf("expected log session 2, got %d", log.SessionNum)
	}
	if log.Summary != "The party defeated Lord Dark." {
		t.Fatalf("expected summary 'The party defeated Lord Dark.', got %s", log.Summary)
	}
	if log.XPAwarded != 500 {
		t.Fatalf("expected XP 500, got %d", log.XPAwarded)
	}
	if len(log.KeyDecisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(log.KeyDecisions))
	}
	if len(log.LootAcquired) != 2 {
		t.Fatalf("expected 2 loot items, got %d", len(log.LootAcquired))
	}
}

func TestNarrativeStateService_Update_NoState(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	update := domain.StateUpdate{
		SessionNum:     1,
		SessionSummary: "First session.",
	}

	// Should create initial state automatically instead of failing
	updated, err := svc.Update(ctx, "missing-campaign", update)
	if err != nil {
		t.Fatalf("expected no error for missing campaign state (should create initial), got: %v", err)
	}
	if updated.CurrentSession != 1 {
		t.Fatalf("expected current session 1, got %d", updated.CurrentSession)
	}
	if len(updated.SessionLog) != 1 {
		t.Fatalf("expected 1 session log entry, got %d", len(updated.SessionLog))
	}
}

func TestNarrativeStateService_GetSessionPrepContext(t *testing.T) {
	svc, _, canonRepo := setupNarrativeStateService()
	ctx := context.Background()

	// Set up canon with entities
	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "npc-giver", Name: "Old Sage", Type: domain.EntityTypeNPC, Role: "ally", CanonState: domain.EntityStateAlive},
			{ID: "npc-villain", Name: "Lord Dark", Type: domain.EntityTypeNPC, Role: "villain", CanonState: domain.EntityStateAlive},
		},
	}
	if err := canonRepo.Save("test-campaign", canon); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
		ActiveQuests: []domain.QuestState{
			{ID: "q-001", Name: "Find the Sword", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver"},
		},
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-villain", Name: "Lord Dark", Session: 2},
		},
		SessionLog: []domain.SessionRecord{
			{SessionNum: 1, Summary: "The party arrived in town."},
			{SessionNum: 2, Summary: "The party defeated Lord Dark."},
		},
	}
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	prep, err := svc.GetSessionPrepContext(ctx, "test-campaign", 3)
	if err != nil {
		t.Fatalf("failed to get session prep context: %v", err)
	}

	// PreviouslyOn should be from last session
	if prep.PreviouslyOn != "The party defeated Lord Dark." {
		t.Fatalf("expected previously on 'The party defeated Lord Dark.', got %s", prep.PreviouslyOn)
	}

	// ActiveQuests
	if len(prep.ActiveQuests) != 1 {
		t.Fatalf("expected 1 active quest, got %d", len(prep.ActiveQuests))
	}

	// RelevantNPCs should include quest giver
	if len(prep.RelevantNPCs) != 1 {
		t.Fatalf("expected 1 relevant NPC, got %d", len(prep.RelevantNPCs))
	}
	if prep.RelevantNPCs[0].ID != "npc-giver" {
		t.Fatalf("expected relevant npc-giver, got %s", prep.RelevantNPCs[0].ID)
	}

	// DM Warnings should mention dead NPC
	if len(prep.DMWarnings) != 1 {
		t.Fatalf("expected 1 DM warning, got %d", len(prep.DMWarnings))
	}
	if prep.DMWarnings[0] != "Lord Dark está muerto" {
		t.Fatalf("expected DM warning 'Lord Dark está muerto', got %s", prep.DMWarnings[0])
	}
}

func TestNarrativeStateService_GetSessionPrepContext_NoLog(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 0,
	}
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	prep, err := svc.GetSessionPrepContext(ctx, "test-campaign", 1)
	if err != nil {
		t.Fatalf("failed to get session prep context: %v", err)
	}

	if prep.PreviouslyOn != "" {
		t.Fatalf("expected empty previously on, got %s", prep.PreviouslyOn)
	}
}

func TestNarrativeStateService_Update_KeyItemReplacement(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
		KeyItems: []domain.KeyItem{
			{ID: "item-001", Name: "Magic Key", Holder: "party", SessionFound: 1},
		},
	}
	svc.Save(ctx, state)

	update := domain.StateUpdate{
		SessionNum: 2,
		KeyItems: []domain.KeyItem{
			{ID: "item-001", Name: "Magic Key", Holder: "npc-merchant", SessionFound: 1},
		},
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	if len(updated.KeyItems) != 1 {
		t.Fatalf("expected 1 key item, got %d", len(updated.KeyItems))
	}
	if updated.KeyItems[0].Holder != "npc-merchant" {
		t.Fatalf("expected holder npc-merchant, got %s", updated.KeyItems[0].Holder)
	}
}

func TestNarrativeStateService_Update_NonMatchingCompletedQuest(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
		ActiveQuests: []domain.QuestState{
			{ID: "q-001", Name: "Find the Sword", Status: "active"},
		},
	}
	svc.Save(ctx, state)

	update := domain.StateUpdate{
		SessionNum:      2,
		CompletedQuests: []string{"q-nonexistent"},
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	if len(updated.ActiveQuests) != 1 {
		t.Fatalf("expected 1 active quest, got %d", len(updated.ActiveQuests))
	}
	if len(updated.CompletedQuests) != 0 {
		t.Fatalf("expected 0 completed quests, got %d", len(updated.CompletedQuests))
	}
}

func TestNarrativeStateService_GetSessionPrepContext_MissingCanon(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "no-canon",
		CurrentSession: 1,
		ActiveQuests: []domain.QuestState{
			{ID: "q-001", Name: "Find the Sword", Status: "active", GiverNPC: "npc-giver"},
		},
		SessionLog: []domain.SessionRecord{
			{SessionNum: 1, Summary: "The party started."},
		},
	}
	svc.Save(ctx, state)

	prep, err := svc.GetSessionPrepContext(ctx, "no-canon", 2)
	if err != nil {
		t.Fatalf("failed to get session prep context: %v", err)
	}

	if prep.PreviouslyOn != "The party started." {
		t.Fatalf("expected previously on, got %s", prep.PreviouslyOn)
	}
	if len(prep.RelevantNPCs) != 0 {
		t.Fatalf("expected 0 relevant NPCs when canon missing, got %d", len(prep.RelevantNPCs))
	}
}

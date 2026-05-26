package services

import (
	"context"
	"strings"
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

func TestNarrativeStateService_Update_PendingEffects(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 1,
		PendingEffects: []domain.DelayedEffect{
			{ID: "rule-001-0", Description: "Existing effect"},
		},
	}
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum: 2,
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	// PendingEffects should be preserved and normalized to non-nil
	if updated.PendingEffects == nil {
		t.Fatalf("expected pending effects to be non-nil after update")
	}
	if len(updated.PendingEffects) != 1 {
		t.Fatalf("expected 1 pending effect preserved, got %d", len(updated.PendingEffects))
	}
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

func TestNarrativeStateService_Update_SessionNumZero(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Initialize state with current session 2
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
	}
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum:     0,
		SessionSummary: "Third session summary.",
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	// SessionNum=0 should auto-increment to CurrentSession+1 = 3
	if updated.CurrentSession != 3 {
		t.Fatalf("expected current session 3 (auto-increment), got %d", updated.CurrentSession)
	}
	if len(updated.SessionLog) != 1 {
		t.Fatalf("expected 1 session log entry, got %d", len(updated.SessionLog))
	}
	if updated.SessionLog[0].SessionNum != 3 {
		t.Fatalf("expected log session_num 3, got %d", updated.SessionLog[0].SessionNum)
	}
}

func TestNarrativeStateService_Update_SessionNumZeroWithExistingState(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	// No pre-existing state — Update should create initial state and auto-increment from 0 to 1
	update := domain.StateUpdate{
		SessionNum:     0,
		SessionSummary: "First session.",
	}

	updated, err := svc.Update(ctx, "new-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	if updated.CurrentSession != 1 {
		t.Fatalf("expected current session 1 (0+1), got %d", updated.CurrentSession)
	}
	if len(updated.SessionLog) != 1 {
		t.Fatalf("expected 1 session log entry, got %d", len(updated.SessionLog))
	}
	if updated.SessionLog[0].SessionNum != 1 {
		t.Fatalf("expected log session_num 1, got %d", updated.SessionLog[0].SessionNum)
	}
}

func TestNarrativeStateService_Update_NegativeSessionNum(t *testing.T) {
	svc, _, _ := setupNarrativeStateService()
	ctx := context.Background()

	// Initialize state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
	}
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum:     -1,
		SessionSummary: "Negative session should fail.",
	}

	_, err := svc.Update(ctx, "test-campaign", update)
	if err == nil {
		t.Fatal("expected error for negative session_num, got nil")
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
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum: 2,
		KeyItems: []domain.KeyItem{
			{ID: "item-001", Name: "Magic Key", Holder: "npc-001", SessionFound: 2},
		},
	}

	updated, err := svc.Update(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	if len(updated.KeyItems) != 1 {
		t.Fatalf("expected 1 key item, got %d", len(updated.KeyItems))
	}
	if updated.KeyItems[0].Holder != "npc-001" {
		t.Fatalf("expected holder npc-001, got %s", updated.KeyItems[0].Holder)
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
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

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
	if err := svc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

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

func TestDedupClues(t *testing.T) {
	tests := []struct {
		name     string
		existing []domain.RevealedClue
		incoming []domain.RevealedClue
		want     []domain.RevealedClue
		wantLen  int
	}{
		{
			name:     "no duplicates",
			existing: []domain.RevealedClue{{ID: "c1", Description: "existing"}},
			incoming: []domain.RevealedClue{{ID: "c2", Description: "new"}},
			wantLen:  2,
		},
		{
			name:     "duplicate incoming skipped",
			existing: []domain.RevealedClue{{ID: "c1", Description: "existing"}},
			incoming: []domain.RevealedClue{{ID: "c1", Description: "dup"}, {ID: "c2", Description: "new"}},
			want:     []domain.RevealedClue{{ID: "c1", Description: "existing"}, {ID: "c2", Description: "new"}},
			wantLen:  2,
		},
		{
			name:     "duplicate within incoming keeps first",
			existing: []domain.RevealedClue{},
			incoming: []domain.RevealedClue{{ID: "c1", Description: "first"}, {ID: "c1", Description: "second"}},
			want:     []domain.RevealedClue{{ID: "c1", Description: "first"}},
			wantLen:  1,
		},
		{
			name:     "empty incoming",
			existing: []domain.RevealedClue{{ID: "c1", Description: "existing"}},
			incoming: []domain.RevealedClue{},
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupClues(tt.existing, tt.incoming)
			if len(got) != tt.wantLen {
				t.Fatalf("expected %d clues, got %d", tt.wantLen, len(got))
			}
			if tt.want != nil {
				for i := range tt.want {
					if got[i].ID != tt.want[i].ID || got[i].Description != tt.want[i].Description {
						t.Fatalf("expected clue %v at index %d, got %v", tt.want[i], i, got[i])
					}
				}
			}
		})
	}
}

func TestDedupDeadNPCs(t *testing.T) {
	tests := []struct {
		name     string
		existing []domain.NPCDeathRecord
		incoming []domain.NPCDeathRecord
		wantLen  int
		wantID   string
	}{
		{
			name:     "no duplicates",
			existing: []domain.NPCDeathRecord{{NPCID: "npc1", Name: "Alice"}},
			incoming: []domain.NPCDeathRecord{{NPCID: "npc2", Name: "Bob"}},
			wantLen:  2,
		},
		{
			name:     "duplicate skipped preserves existing",
			existing: []domain.NPCDeathRecord{{NPCID: "npc1", Name: "Alice"}},
			incoming: []domain.NPCDeathRecord{{NPCID: "npc1", Name: "Alice2"}},
			wantLen:  1,
			wantID:   "npc1",
		},
		{
			name:     "duplicate within incoming keeps first",
			existing: []domain.NPCDeathRecord{},
			incoming: []domain.NPCDeathRecord{{NPCID: "npc1", Name: "First"}, {NPCID: "npc1", Name: "Second"}},
			wantLen:  1,
			wantID:   "npc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedupDeadNPCs(tt.existing, tt.incoming)
			if len(got) != tt.wantLen {
				t.Fatalf("expected %d dead NPCs, got %d", tt.wantLen, len(got))
			}
			if tt.wantID != "" && len(got) > 0 {
				if got[0].NPCID != tt.wantID {
					t.Fatalf("expected NPCID %s, got %s", tt.wantID, got[0].NPCID)
				}
			}
		})
	}
}

func TestMergeQuests(t *testing.T) {
	tests := []struct {
		name     string
		existing []domain.QuestState
		incoming []domain.QuestState
		wantLen  int
		wantVals map[string]string // id -> status
	}{
		{
			name:     "append new quests",
			existing: []domain.QuestState{{ID: "q1", Name: "Quest 1", Status: "active"}},
			incoming: []domain.QuestState{{ID: "q2", Name: "Quest 2", Status: "active"}},
			wantLen:  2,
			wantVals: map[string]string{"q1": "active", "q2": "active"},
		},
		{
			name:     "update existing quest in place",
			existing: []domain.QuestState{{ID: "q1", Name: "Quest 1", Status: "active"}},
			incoming: []domain.QuestState{{ID: "q1", Name: "Quest 1 Updated", Status: "completed"}},
			wantLen:  1,
			wantVals: map[string]string{"q1": "completed"},
		},
		{
			name:     "update and append mixed",
			existing: []domain.QuestState{{ID: "q1", Name: "Quest 1", Status: "active"}},
			incoming: []domain.QuestState{{ID: "q1", Name: "Quest 1", Status: "completed"}, {ID: "q2", Name: "Quest 2", Status: "active"}},
			wantLen:  2,
			wantVals: map[string]string{"q1": "completed", "q2": "active"},
		},
		{
			name:     "order preserved for existing",
			existing: []domain.QuestState{{ID: "q1", Name: "A", Status: "active"}, {ID: "q2", Name: "B", Status: "active"}},
			incoming: []domain.QuestState{{ID: "q3", Name: "C", Status: "active"}},
			wantLen:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeQuests(tt.existing, tt.incoming)
			if len(got) != tt.wantLen {
				t.Fatalf("expected %d quests, got %d", tt.wantLen, len(got))
			}
			for _, q := range got {
				if wantStatus, ok := tt.wantVals[q.ID]; ok && q.Status != wantStatus {
					t.Fatalf("expected quest %s status %s, got %s", q.ID, wantStatus, q.Status)
				}
			}
		})
	}
}

func TestMergeKeyItems(t *testing.T) {
	tests := []struct {
		name     string
		existing []domain.KeyItem
		incoming []domain.KeyItem
		wantLen  int
		wantVals map[string]string // id -> holder
	}{
		{
			name:     "append new items",
			existing: []domain.KeyItem{{ID: "i1", Name: "Item 1", Holder: "party"}},
			incoming: []domain.KeyItem{{ID: "i2", Name: "Item 2", Holder: "npc"}},
			wantLen:  2,
			wantVals: map[string]string{"i1": "party", "i2": "npc"},
		},
		{
			name:     "update existing item in place",
			existing: []domain.KeyItem{{ID: "i1", Name: "Item 1", Holder: "party"}},
			incoming: []domain.KeyItem{{ID: "i1", Name: "Item 1", Holder: "npc"}},
			wantLen:  1,
			wantVals: map[string]string{"i1": "npc"},
		},
		{
			name:     "update and append mixed",
			existing: []domain.KeyItem{{ID: "i1", Name: "Item 1", Holder: "party"}},
			incoming: []domain.KeyItem{{ID: "i1", Name: "Item 1", Holder: "npc"}, {ID: "i2", Name: "Item 2", Holder: "party"}},
			wantLen:  2,
			wantVals: map[string]string{"i1": "npc", "i2": "party"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeKeyItems(tt.existing, tt.incoming)
			if len(got) != tt.wantLen {
				t.Fatalf("expected %d items, got %d", tt.wantLen, len(got))
			}
			for _, item := range got {
				if wantHolder, ok := tt.wantVals[item.ID]; ok && item.Holder != wantHolder {
					t.Fatalf("expected item %s holder %s, got %s", item.ID, wantHolder, item.Holder)
				}
			}
		})
	}
}

func TestSyncStateToCanon_DeadNPC(t *testing.T) {
	svc, stateRepo, canonRepo := setupNarrativeStateService()
	ctx := context.Background()

	// Set up canon with alive NPC
	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "npc-villain", Name: "Lord Dark", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
			{ID: "npc-ally", Name: "Good Guy", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
		},
	}
	if err := canonRepo.Save("test-campaign", canon); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Set up narrative state with completed quest and dead NPC
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
		CompletedQuests: []domain.QuestState{
			{ID: "q-001", Name: "Find the Sword", Status: "completed"},
		},
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-villain", Name: "Lord Dark", Session: 2},
		},
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum:      2,
		CompletedQuests: []string{"q-001"},
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-villain", Name: "Lord Dark", Session: 2},
		},
	}

	warnings, err := svc.SyncStateToCanon(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("SyncStateToCanon error: %v", err)
	}

	// Should have no warnings
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %v", warnings)
	}

	// Verify canon entity is now dead
	updatedCanon, err := canonRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	found := false
	for _, e := range updatedCanon.Entities {
		if e.ID == "npc-villain" {
			found = true
			if e.CanonState != domain.EntityStateDead {
				t.Fatalf("expected npc-villain to be dead, got %s", e.CanonState)
			}
		}
		if e.ID == "npc-ally" {
			if e.CanonState != domain.EntityStateAlive {
				t.Fatalf("expected npc-ally to remain alive, got %s", e.CanonState)
			}
		}
	}
	if !found {
		t.Fatal("npc-villain not found in canon")
	}

	// Verify timeline event appended
	if len(updatedCanon.Timeline) != 1 {
		t.Fatalf("expected 1 timeline event, got %d", len(updatedCanon.Timeline))
	}
	if !strings.Contains(updatedCanon.Timeline[0].Description, "Find the Sword") {
		t.Fatalf("expected timeline event to mention 'Find the Sword', got %s", updatedCanon.Timeline[0].Description)
	}
}

func TestSyncStateToCanon_AlreadyDead(t *testing.T) {
	svc, stateRepo, canonRepo := setupNarrativeStateService()
	ctx := context.Background()

	// Set up canon with already-dead NPC
	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "npc-villain", Name: "Lord Dark", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateDead},
		},
	}
	if err := canonRepo.Save("test-campaign", canon); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-villain", Name: "Lord Dark", Session: 2},
		},
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum: 2,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-villain", Name: "Lord Dark", Session: 2},
		},
	}

	warnings, err := svc.SyncStateToCanon(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("SyncStateToCanon error: %v", err)
	}

	// Should have 0 warnings (already dead is a no-op, not a warning)
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings for already-dead NPC, got %v", warnings)
	}

	// Verify still dead
	updatedCanon, err := canonRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	for _, e := range updatedCanon.Entities {
		if e.ID == "npc-villain" && e.CanonState != domain.EntityStateDead {
			t.Fatalf("expected npc-villain to remain dead, got %s", e.CanonState)
		}
	}
}

func TestSyncStateToCanon_MissingEntity(t *testing.T) {
	svc, stateRepo, canonRepo := setupNarrativeStateService()
	ctx := context.Background()

	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities:      []domain.CanonEntity{},
	}
	if err := canonRepo.Save("test-campaign", canon); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-missing", Name: "Missing NPC", Session: 2},
		},
	}
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	update := domain.StateUpdate{
		SessionNum: 2,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-missing", Name: "Missing NPC", Session: 2},
		},
	}

	warnings, err := svc.SyncStateToCanon(ctx, "test-campaign", update)
	if err != nil {
		t.Fatalf("SyncStateToCanon error: %v", err)
	}

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for missing entity, got %d", len(warnings))
	}
	if !strings.Contains(warnings[0], "npc-missing") {
		t.Fatalf("expected warning to mention npc-missing, got %s", warnings[0])
	}
}

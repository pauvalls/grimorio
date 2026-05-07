package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestSessionPrepService_GetPrep(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()

	svc := NewSessionPrepService(canonRepo, stateRepo)

	seedFullState := func() {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "test-campaign",
			Entities: []domain.CanonEntity{
				{ID: "npc-giver", Name: "Eldrin the Wise", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
				{ID: "npc-ally", Name: "Thorn the Brave", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
				{ID: "npc-dead", Name: "Villain X", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateDead},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "npc-giver", To: "quest-main", Type: domain.RelationshipTypeAlly, Strength: 5},
			},
		}
		_ = canonRepo.Save("test-campaign", doc)

		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     "test-campaign",
			CurrentSession: 3,
			SessionLog: []domain.SessionRecord{
				{
					SessionNum: 1,
					Date:       time.Now().Add(-48 * time.Hour),
					Summary:    "The party arrived in town and met Eldrin.",
				},
				{
					SessionNum: 2,
					Date:       time.Now().Add(-24 * time.Hour),
					Summary:    "The party explored the dungeon and found a key artifact.",
				},
				{
					SessionNum: 3,
					Date:       time.Now(),
					Summary:    "The party defeated Villain X and rescued Thorn.",
					KeyDecisions: []domain.Decision{
						{ID: "dec-1", Description: "Spare the villain's minions", ChoiceMade: "spared"},
					},
				},
			},
			ActiveQuests: []domain.QuestState{
				{ID: "quest-main", Name: "Find the Artifact", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver"},
				{ID: "quest-side", Name: "Help Thorn", Status: "active", SourceAct: "act-2", GiverNPC: "npc-ally"},
			},
			DeadNPCs: []domain.NPCDeathRecord{
				{NPCID: "npc-dead", Name: "Villain X", Session: 3, Cause: "combat"},
			},
			DMOverrides: []domain.DMOverride{
				{ID: "ovr-1", TargetID: "npc-giver", Field: "location", NewValue: "castle", Reason: "moved for safety"},
			},
		}
		_ = stateRepo.Save("test-campaign", state)
	}

	t.Run("full state returns complete prep", func(t *testing.T) {
		seedFullState()

		prep, warnings, err := svc.GetPrep(ctx, "test-campaign", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prep == nil {
			t.Fatalf("expected prep, got nil")
		}

		// SessionNum should default to CurrentSession+1 = 4
		if prep.SessionNum != 4 {
			t.Fatalf("session_num = %d, want 4", prep.SessionNum)
		}

		// PreviouslyOn should be from last session (session 3)
		if !strings.Contains(prep.PreviouslyOn, "defeated Villain X") {
			t.Fatalf("previously_on = %q, want to contain 'defeated Villain X'", prep.PreviouslyOn)
		}

		// Should have active quests
		if len(prep.ActiveQuests) == 0 {
			t.Fatalf("expected active quests, got none")
		}

		// Should have relevant NPCs (alive ones connected to quests)
		foundEldrin := false
		for _, npc := range prep.RelevantNPCs {
			if strings.Contains(npc, "Eldrin") {
				foundEldrin = true
			}
		}
		if !foundEldrin {
			t.Fatalf("expected Eldrin in relevant NPCs, got %v", prep.RelevantNPCs)
		}

		// Should have reminders (dead NPC with canon state dead, DM override)
		if len(prep.Reminders) == 0 {
			t.Fatalf("expected reminders, got none")
		}

		// Should have prep date set
		if prep.PrepDate.IsZero() {
			t.Fatalf("expected prep_date to be set")
		}

		// Warnings should be empty for full state
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
	})

	t.Run("empty state returns warnings but valid sheet", func(t *testing.T) {
		emptyCanonRepo := repository.NewMemoryCanonRepository()
		emptyStateRepo := repository.NewMemoryNarrativeStateRepository()
		emptySvc := NewSessionPrepService(emptyCanonRepo, emptyStateRepo)

		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "empty-campaign",
		}
		_ = emptyCanonRepo.Save("empty-campaign", doc)

		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     "empty-campaign",
			CurrentSession: 0,
		}
		_ = emptyStateRepo.Save("empty-campaign", state)

		prep, warnings, err := emptySvc.GetPrep(ctx, "empty-campaign", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prep == nil {
			t.Fatalf("expected prep, got nil")
		}

		// Empty state should have placeholder previously_on
		if prep.PreviouslyOn == "" {
			t.Fatalf("expected placeholder previously_on, got empty")
		}

		// Should have warnings
		if len(warnings) == 0 {
			t.Fatalf("expected warnings for empty state")
		}
	})

	t.Run("missing canon returns warning", func(t *testing.T) {
		canonOnlyRepo := repository.NewMemoryCanonRepository()
		stateOnlyRepo := repository.NewMemoryNarrativeStateRepository()
		partialSvc := NewSessionPrepService(canonOnlyRepo, stateOnlyRepo)

		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     "no-canon",
			CurrentSession: 1,
		}
		_ = stateOnlyRepo.Save("no-canon", state)

		prep, warnings, err := partialSvc.GetPrep(ctx, "no-canon", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prep == nil {
			t.Fatalf("expected prep, got nil")
		}

		foundCanonWarning := false
		for _, w := range warnings {
			if strings.Contains(w, "canon") {
				foundCanonWarning = true
			}
		}
		if !foundCanonWarning {
			t.Fatalf("expected canon warning, got %v", warnings)
		}
	})

	t.Run("no sessions returns placeholder", func(t *testing.T) {
		noSessionCanonRepo := repository.NewMemoryCanonRepository()
		noSessionStateRepo := repository.NewMemoryNarrativeStateRepository()
		noSessionSvc := NewSessionPrepService(noSessionCanonRepo, noSessionStateRepo)

		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "no-sessions",
			Entities: []domain.CanonEntity{
				{ID: "npc-1", Name: "Test NPC", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
			},
		}
		_ = noSessionCanonRepo.Save("no-sessions", doc)

		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     "no-sessions",
			CurrentSession: 0,
			SessionLog:     []domain.SessionRecord{},
		}
		_ = noSessionStateRepo.Save("no-sessions", state)

		prep, _, err := noSessionSvc.GetPrep(ctx, "no-sessions", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(prep.PreviouslyOn, "No previous sessions") {
			t.Fatalf("expected placeholder for no sessions, got %q", prep.PreviouslyOn)
		}
	})

	t.Run("specific session number overrides default", func(t *testing.T) {
		seedFullState()

		prep, _, err := svc.GetPrep(ctx, "test-campaign", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prep.SessionNum != 10 {
			t.Fatalf("session_num = %d, want 10", prep.SessionNum)
		}
	})

	t.Run("dead npcs excluded from relevant", func(t *testing.T) {
		seedFullState()

		prep, _, err := svc.GetPrep(ctx, "test-campaign", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, npc := range prep.RelevantNPCs {
			if strings.Contains(npc, "Villain X") {
				t.Fatalf("dead NPC Villain X should not be in relevant NPCs")
			}
		}
	})
}

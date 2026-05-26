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

	svc := NewSessionPrepService(canonRepo, stateRepo, nil)

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
		emptySvc := NewSessionPrepService(emptyCanonRepo, emptyStateRepo, nil)

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
		partialSvc := NewSessionPrepService(canonOnlyRepo, stateOnlyRepo, nil)

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
		noSessionSvc := NewSessionPrepService(noSessionCanonRepo, noSessionStateRepo, nil)

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

	t.Run("missing narrative state creates initial state", func(t *testing.T) {
		missingStateCanonRepo := repository.NewMemoryCanonRepository()
		missingStateRepo := repository.NewMemoryNarrativeStateRepository()
		missingSvc := NewSessionPrepService(missingStateCanonRepo, missingStateRepo, nil)

		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "missing-state-campaign",
			Entities: []domain.CanonEntity{
				{ID: "npc-1", Name: "Test NPC", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
			},
		}
		_ = missingStateCanonRepo.Save("missing-state-campaign", doc)

		// NOTE: deliberately do NOT save narrative state — service should create initial state

		prep, warnings, err := missingSvc.GetPrep(ctx, "missing-state-campaign", 0)
		if err != nil {
			t.Fatalf("expected no error for missing narrative state (should create initial), got: %v", err)
		}
		if prep == nil {
			t.Fatalf("expected prep, got nil")
		}
		if prep.SessionNum != 1 {
			t.Fatalf("expected session_num 1 (0+1 from initial state), got %d", prep.SessionNum)
		}
		if !strings.Contains(prep.PreviouslyOn, "No previous sessions") {
			t.Fatalf("expected placeholder previously_on, got %q", prep.PreviouslyOn)
		}
		// Should have warnings for empty state
		if len(warnings) == 0 {
			t.Fatalf("expected warnings for empty initial state")
		}
	})

	t.Run("nil state from repo creates initial state", func(t *testing.T) {
		// Use a repo that returns nil state without error
		nilStateRepo := &nilReturningStateRepo{}
		nilCanonRepo := repository.NewMemoryCanonRepository()
		nilSvc := NewSessionPrepService(nilCanonRepo, nilStateRepo, nil)

		prep, warnings, err := nilSvc.GetPrep(ctx, "nil-state-campaign", 0)
		if err != nil {
			t.Fatalf("expected no error for nil state (should create initial), got: %v", err)
		}
		if prep == nil {
			t.Fatalf("expected prep, got nil")
		}
		if prep.SessionNum != 1 {
			t.Fatalf("expected session_num 1, got %d", prep.SessionNum)
		}
		if len(warnings) == 0 {
			t.Fatalf("expected warnings for empty initial state")
		}
	})
}

// nilReturningStateRepo is a mock repo that returns nil state without error
type nilReturningStateRepo struct{}

func (r *nilReturningStateRepo) Load(campaignID string) (*domain.NarrativeState, error) {
	return nil, nil
}

func (r *nilReturningStateRepo) Save(campaignID string, state *domain.NarrativeState) error {
	return nil
}

func (r *nilReturningStateRepo) Exists(campaignID string) bool {
	return false
}

func TestSessionPrepService_generatePreviouslyOn(t *testing.T) {
	svc := NewSessionPrepService(nil, nil, nil)

	t.Run("no sessions returns placeholder", func(t *testing.T) {
		state := &domain.NarrativeState{SessionLog: []domain.SessionRecord{}}
		got := svc.generatePreviouslyOn(state)
		if !strings.Contains(got, "No previous sessions") {
			t.Fatalf("expected placeholder, got %q", got)
		}
	})

	t.Run("one session", func(t *testing.T) {
		state := &domain.NarrativeState{
			SessionLog: []domain.SessionRecord{
				{SessionNum: 1, Summary: "First session."},
			},
		}
		got := svc.generatePreviouslyOn(state)
		if !strings.Contains(got, "Arc context") {
			t.Fatalf("expected arc context line, got %q", got)
		}
		if !strings.Contains(got, "First session.") {
			t.Fatalf("expected first session summary, got %q", got)
		}
	})

	t.Run("three sessions shows all", func(t *testing.T) {
		state := &domain.NarrativeState{
			SessionLog: []domain.SessionRecord{
				{SessionNum: 1, Summary: "Session 1."},
				{SessionNum: 2, Summary: "Session 2."},
				{SessionNum: 3, Summary: "Session 3."},
			},
		}
		got := svc.generatePreviouslyOn(state)
		if !strings.Contains(got, "Session 3.") {
			t.Fatalf("expected session 3, got %q", got)
		}
		if !strings.Contains(got, "Session 2.") {
			t.Fatalf("expected session 2, got %q", got)
		}
		if !strings.Contains(got, "Session 1.") {
			t.Fatalf("expected session 1, got %q", got)
		}
	})

	t.Run("four sessions shows last 3", func(t *testing.T) {
		state := &domain.NarrativeState{
			SessionLog: []domain.SessionRecord{
				{SessionNum: 1, Summary: "Session 1."},
				{SessionNum: 2, Summary: "Session 2."},
				{SessionNum: 3, Summary: "Session 3."},
				{SessionNum: 4, Summary: "Session 4."},
			},
		}
		got := svc.generatePreviouslyOn(state)
		if !strings.Contains(got, "Session 4.") {
			t.Fatalf("expected session 4, got %q", got)
		}
		if !strings.Contains(got, "Session 3.") {
			t.Fatalf("expected session 3, got %q", got)
		}
		if !strings.Contains(got, "Session 2.") {
			t.Fatalf("expected session 2, got %q", got)
		}
		if strings.Contains(got, "Session 1.") {
			t.Fatalf("expected session 1 to be excluded, got %q", got)
		}
	})
}

func TestSessionPrepService_generateLikelyScenarios(t *testing.T) {
	svc := NewSessionPrepService(nil, nil, nil)

	t.Run("empty quests returns empty", func(t *testing.T) {
		state := &domain.NarrativeState{ActiveQuests: []domain.QuestState{}}
		got := svc.generateLikelyScenarios(state, nil, 4)
		if len(got) != 0 {
			t.Fatalf("expected 0 scenarios, got %d", len(got))
		}
	})

	t.Run("caps at 7", func(t *testing.T) {
		state := &domain.NarrativeState{
			ActiveQuests: []domain.QuestState{
				{ID: "q1", Name: "Quest 1", Status: "active", SourceAct: "act-1"},
				{ID: "q2", Name: "Quest 2", Status: "active", SourceAct: "act-1"},
				{ID: "q3", Name: "Quest 3", Status: "active", SourceAct: "act-1"},
				{ID: "q4", Name: "Quest 4", Status: "active", SourceAct: "act-1"},
				{ID: "q5", Name: "Quest 5", Status: "active", SourceAct: "act-1"},
				{ID: "q6", Name: "Quest 6", Status: "active", SourceAct: "act-1"},
				{ID: "q7", Name: "Quest 7", Status: "active", SourceAct: "act-1"},
				{ID: "q8", Name: "Quest 8", Status: "active", SourceAct: "act-1"},
			},
		}
		got := svc.generateLikelyScenarios(state, nil, 4)
		if len(got) != 7 {
			t.Fatalf("expected 7 scenarios (capped), got %d", len(got))
		}
	})

	t.Run("pending effects first priority", func(t *testing.T) {
		state := &domain.NarrativeState{
			PendingEffects: []domain.DelayedEffect{
				{ID: "e1", Description: "Effect 1", ApplySession: 4},
			},
			ActiveQuests: []domain.QuestState{
				{ID: "q1", Name: "Quest 1", Status: "active", SourceAct: "act-1"},
			},
		}
		got := svc.generateLikelyScenarios(state, nil, 4)
		if len(got) < 2 {
			t.Fatalf("expected at least 2 scenarios, got %d", len(got))
		}
		if !strings.Contains(got[0], "Effect 1") {
			t.Fatalf("expected pending effect first, got %q", got[0])
		}
	})
}

func TestSessionPrepService_generateReminders(t *testing.T) {
	svc := NewSessionPrepService(nil, nil, nil)

	t.Run("due pending effects in reminders", func(t *testing.T) {
		state := &domain.NarrativeState{
			PendingEffects: []domain.DelayedEffect{
				{ID: "e1", Description: "Village burns down", Target: "Village", ApplySession: 4},
			},
		}
		doc := &domain.CanonDocument{}
		got := svc.generateReminders(state, doc, 4)
		found := false
		for _, r := range got {
			if strings.Contains(r, "Village burns down") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected reminder for due pending effect, got %v", got)
		}
	})

	t.Run("non-due effects excluded", func(t *testing.T) {
		state := &domain.NarrativeState{
			PendingEffects: []domain.DelayedEffect{
				{ID: "e1", Description: "Future effect", Target: "Town", ApplySession: 6},
			},
		}
		doc := &domain.CanonDocument{}
		got := svc.generateReminders(state, doc, 4)
		for _, r := range got {
			if strings.Contains(r, "Future effect") {
				t.Fatalf("expected future effect excluded, got %q", r)
			}
		}
	})
}

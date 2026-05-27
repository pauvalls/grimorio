package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestConsequenceEngine_Evaluate(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Rules: []domain.CanonRule{
			{
				ID:   "rule-informant-death",
				// Stored as JSON payload in Statement for simplicity in this phase
				Statement: `{"trigger":{"type":"npc_death","entity_id":"informant"},"effects":[{"type":"spawn","target":"npc","description":"spawn replacement"}],"priority":5}`,
				Domain: "consequence",
			},
			{
				ID:   "rule-priority-high",
				Statement: `{"trigger":{"type":"npc_death","entity_id":"informant"},"effects":[{"type":"alert","target":"faction","description":"alert faction"}],"priority":10}`,
				Domain: "consequence",
			},
			{
				ID:   "rule-dm-override",
				Statement: `{"trigger":{"type":"any"},"effects":[{"type":"override","target":"all"}],"priority":1,"dm_override":true}`,
				Domain: "consequence",
			},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	engine := NewConsequenceEngine(canonRepo)

	t.Run("rule fires on npc death", func(t *testing.T) {
		state := &domain.NarrativeState{
			CampaignID:     "test-campaign",
			CurrentSession: 2,
			DeadNPCs: []domain.NPCDeathRecord{
				{NPCID: "informant", Name: "The Informant", Session: 2},
			},
		}

		eval, err := engine.Evaluate(ctx, "test-campaign", state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eval.TriggeredRules) == 0 {
			t.Fatalf("expected triggered rules, got none")
		}
		found := false
		for _, r := range eval.TriggeredRules {
			if r.ID == "rule-informant-death" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected informant-death rule to trigger")
		}
	})

	t.Run("priority order", func(t *testing.T) {
		state := &domain.NarrativeState{
			CampaignID:     "test-campaign",
			CurrentSession: 2,
			DeadNPCs: []domain.NPCDeathRecord{
				{NPCID: "informant", Name: "The Informant", Session: 2},
			},
		}

		eval, err := engine.Evaluate(ctx, "test-campaign", state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eval.ImmediateEffects) < 2 {
			t.Fatalf("expected at least 2 effects, got %d", len(eval.ImmediateEffects))
		}
		// High priority (10) should come before low priority (5)
		if eval.ImmediateEffects[0].Description != "alert faction" {
			t.Fatalf("first effect = %q, want alert faction", eval.ImmediateEffects[0].Description)
		}
	})

	t.Run("dm override always applies", func(t *testing.T) {
		state := &domain.NarrativeState{
			CampaignID:     "test-campaign",
			CurrentSession: 1,
		}

		eval, err := engine.Evaluate(ctx, "test-campaign", state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		found := false
		for _, r := range eval.TriggeredRules {
			if r.ID == "rule-dm-override" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected DM override rule to trigger")
		}
	})

	t.Run("empty ruleset no panic", func(t *testing.T) {
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "empty-campaign",
			Rules:         []domain.CanonRule{},
		}
		_ = freshCanonRepo.Save("empty-campaign", freshDoc)
		freshEngine := NewConsequenceEngine(freshCanonRepo)

		state := &domain.NarrativeState{CampaignID: "empty-campaign"}
		eval, err := freshEngine.Evaluate(ctx, "empty-campaign", state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(eval.TriggeredRules) != 0 {
			t.Fatalf("expected 0 triggered rules, got %d", len(eval.TriggeredRules))
		}
	})
}

func TestConsequenceEngine_RegisterReactor(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	engine := NewConsequenceEngine(canonRepo)

	called := false
	engine.RegisterReactor("custom_trigger", func(trigger domain.Trigger, state *domain.NarrativeState) ([]domain.Effect, error) {
		called = true
		return []domain.Effect{{Type: "custom", Description: "custom effect"}}, nil
	})

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "reactor-campaign",
		Rules: []domain.CanonRule{
			{
				ID:        "rule-custom",
				Statement: `{"trigger":{"type":"custom_trigger","entity_id":"test"},"effects":[{"type":"custom","target":"test"}],"priority":1}`,
				Domain:    "consequence",
			},
		},
	}
	_ = canonRepo.Save("reactor-campaign", doc)

	state := &domain.NarrativeState{CampaignID: "reactor-campaign"}
	_, err := engine.Evaluate(context.Background(), "reactor-campaign", state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected reactor to be called")
	}
}

func TestConsequenceEngine_DelayedEffects(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "delayed-campaign",
		Rules: []domain.CanonRule{
			{
				ID:        "rule-delayed",
				Statement: `{"trigger":{"type":"npc_death","entity_id":"boss"},"effects":[{"type":"alert","target":"faction","delay":"2 sessions","description":"delayed alert"}],"priority":5}`,
				Domain:    "consequence",
			},
			{
				ID:        "rule-immediate",
				Statement: `{"trigger":{"type":"npc_death","entity_id":"boss"},"effects":[{"type":"spawn","target":"npc","description":"immediate spawn"}],"priority":5}`,
				Domain:    "consequence",
			},
		},
	}
	_ = canonRepo.Save("delayed-campaign", doc)

	engine := NewConsequenceEngine(canonRepo)

	state := &domain.NarrativeState{
		CampaignID:     "delayed-campaign",
		CurrentSession: 3,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "boss", Name: "Big Boss", Session: 3},
		},
	}

	eval, err := engine.Evaluate(ctx, "delayed-campaign", state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(eval.ImmediateEffects) != 1 {
		t.Fatalf("expected 1 immediate effect, got %d", len(eval.ImmediateEffects))
	}
	if eval.ImmediateEffects[0].Description != "immediate spawn" {
		t.Fatalf("expected immediate spawn, got %s", eval.ImmediateEffects[0].Description)
	}

	if len(eval.DelayedEffects) != 1 {
		t.Fatalf("expected 1 delayed effect, got %d", len(eval.DelayedEffects))
	}
	de := eval.DelayedEffects[0]
	if de.Effect.Description != "delayed alert" {
		t.Fatalf("expected delayed alert, got %s", de.Effect.Description)
	}
	if de.TriggerSession != 3 {
		t.Fatalf("expected trigger session 3, got %d", de.TriggerSession)
	}
	if de.ApplySession != 5 {
		t.Fatalf("expected apply session 5, got %d", de.ApplySession)
	}
}

func TestParseDelaySessions(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"2 sessions", 2},
		{"1", 1},
		{"5", 5},
		{"invalid", 1},
		{"", 1},
		{"0", 1},
		{"-1", 1},
	}

	for _, tc := range tests {
		got := parseDelaySessions(tc.input)
		if got != tc.expected {
			t.Errorf("parseDelaySessions(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

func TestConsequenceEngine_Evaluate_EnrichedDelayedEffects(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "enriched-campaign",
		Rules: []domain.CanonRule{
			{
				ID:        "rule-delayed-enriched",
				Statement: `{"trigger":{"type":"npc_death","entity_id":"boss"},"effects":[{"type":"spawn","target":"npc","delay":"2 sessions","description":"replacement appears"}],"priority":5}`,
				Domain:    "consequence",
			},
		},
	}
	_ = canonRepo.Save("enriched-campaign", doc)

	engine := NewConsequenceEngine(canonRepo)
	state := &domain.NarrativeState{
		CampaignID:     "enriched-campaign",
		CurrentSession: 3,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "boss", Name: "Big Boss", Session: 3},
		},
	}

	eval, err := engine.Evaluate(ctx, "enriched-campaign", state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eval.DelayedEffects) != 1 {
		t.Fatalf("expected 1 delayed effect, got %d", len(eval.DelayedEffects))
	}
	de := eval.DelayedEffects[0]
	if de.ID != "rule-delayed-enriched-0" {
		t.Fatalf("id = %q, want rule-delayed-enriched-0", de.ID)
	}
	if de.Description != "replacement appears" {
		t.Fatalf("description = %q, want 'replacement appears'", de.Description)
	}
	if de.EffectType != "spawn" {
		t.Fatalf("effect_type = %q, want spawn", de.EffectType)
	}
	if de.Target != "npc" {
		t.Fatalf("target = %q, want npc", de.Target)
	}
	if de.TriggerSession != 3 {
		t.Fatalf("trigger_session = %d, want 3", de.TriggerSession)
	}
	if de.ApplySession != 5 {
		t.Fatalf("apply_session = %d, want 5", de.ApplySession)
	}
	if de.Applied != false {
		t.Fatalf("applied = %v, want false", de.Applied)
	}
}

func TestConsequenceEngine_ApplyEvaluation(t *testing.T) {
	engine := NewConsequenceEngine(repository.NewMemoryCanonRepository())

	t.Run("append new delayed effects", func(t *testing.T) {
		state := &domain.NarrativeState{CampaignID: "test", PendingEffects: []domain.DelayedEffect{}}
		eval := &domain.ConsequenceEvaluation{
			SessionNum: 3,
			DelayedEffects: []domain.DelayedEffect{
				{ID: "rule-001-0", Description: "Effect 1", ApplySession: 5},
				{ID: "rule-001-1", Description: "Effect 2", ApplySession: 6},
			},
		}

		err := engine.ApplyEvaluation(eval, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(state.PendingEffects) != 2 {
			t.Fatalf("expected 2 pending effects, got %d", len(state.PendingEffects))
		}
		if state.PendingEffects[0].Applied != false {
			t.Fatalf("expected Applied=false, got %v", state.PendingEffects[0].Applied)
		}
		if state.PendingEffects[0].TriggerSession != 3 {
			t.Fatalf("expected TriggerSession=3, got %d", state.PendingEffects[0].TriggerSession)
		}
	})

	t.Run("deduplicate by ID", func(t *testing.T) {
		state := &domain.NarrativeState{
			CampaignID:     "test",
			PendingEffects: []domain.DelayedEffect{{ID: "rule-001-0", Description: "Existing"}},
		}
		eval := &domain.ConsequenceEvaluation{
			SessionNum: 3,
			DelayedEffects: []domain.DelayedEffect{
				{ID: "rule-001-0", Description: "Duplicate"},
				{ID: "rule-002-0", Description: "New"},
			},
		}

		err := engine.ApplyEvaluation(eval, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(state.PendingEffects) != 2 {
			t.Fatalf("expected 2 pending effects, got %d", len(state.PendingEffects))
		}
		if state.PendingEffects[0].Description != "Existing" {
			t.Fatalf("expected existing preserved, got %s", state.PendingEffects[0].Description)
		}
	})

	t.Run("non-repeatable rule blocked", func(t *testing.T) {
		state := &domain.NarrativeState{
			CampaignID:     "test",
			PendingEffects: []domain.DelayedEffect{{ID: "rule-once-0", Description: "First"}},
		}
		eval := &domain.ConsequenceEvaluation{
			SessionNum: 3,
			TriggeredRules: []domain.ConsequenceRule{
				{ID: "rule-once", IsRepeatable: false},
			},
			DelayedEffects: []domain.DelayedEffect{
				{ID: "rule-once-1", Description: "Second from same rule"},
			},
		}

		err := engine.ApplyEvaluation(eval, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(state.PendingEffects) != 1 {
			t.Fatalf("expected 1 pending effect (non-repeatable blocked), got %d", len(state.PendingEffects))
		}
	})

	t.Run("repeatable rule allowed", func(t *testing.T) {
		state := &domain.NarrativeState{
			CampaignID:     "test",
			PendingEffects: []domain.DelayedEffect{{ID: "rule-repeat-0", Description: "First"}},
		}
		eval := &domain.ConsequenceEvaluation{
			SessionNum: 3,
			TriggeredRules: []domain.ConsequenceRule{
				{ID: "rule-repeat", IsRepeatable: true},
			},
			DelayedEffects: []domain.DelayedEffect{
				{ID: "rule-repeat-1", Description: "Second from repeatable rule"},
			},
		}

		err := engine.ApplyEvaluation(eval, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(state.PendingEffects) != 2 {
			t.Fatalf("expected 2 pending effects (repeatable allowed), got %d", len(state.PendingEffects))
		}
	})
}

func TestConsequenceEngine_GetPendingEffects(t *testing.T) {
	engine := NewConsequenceEngine(repository.NewMemoryCanonRepository())

	state := &domain.NarrativeState{
		CampaignID: "test",
		PendingEffects: []domain.DelayedEffect{
			{ID: "e1", Description: "Due session 2", ApplySession: 2, Applied: false},
			{ID: "e2", Description: "Due session 3", ApplySession: 3, Applied: false},
			{ID: "e3", Description: "Due session 5", ApplySession: 5, Applied: false},
			{ID: "e4", Description: "Already applied", ApplySession: 2, Applied: true},
		},
	}

	pending := engine.GetPendingEffects(state, 3)
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending effects for session 3, got %d", len(pending))
	}
	if pending[0].ID != "e1" {
		t.Fatalf("expected e1 first (sorted), got %s", pending[0].ID)
	}
	if pending[1].ID != "e2" {
		t.Fatalf("expected e2 second, got %s", pending[1].ID)
	}
}

func TestConsequenceEngine_Conditions(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "conditions-campaign",
		Rules: []domain.CanonRule{
			{
				ID:        "rule-quest-active",
				Statement: `{"trigger":{"type":"any"},"conditions":[{"type":"quest_active","target":"main-quest"}],"effects":[{"type":"reward","target":"party","description":"bonus reward"}],"priority":5}`,
				Domain:    "consequence",
			},
			{
				ID:        "rule-session-min",
				Statement: `{"trigger":{"type":"any"},"conditions":[{"type":"session_min","value":5}],"effects":[{"type":"event","target":"world","description":"late game event"}],"priority":5}`,
				Domain:    "consequence",
			},
		},
	}
	_ = canonRepo.Save("conditions-campaign", doc)

	engine := NewConsequenceEngine(canonRepo)

	// Quest active condition passes
	state := &domain.NarrativeState{
		CampaignID:     "conditions-campaign",
		CurrentSession: 3,
		ActiveQuests: []domain.QuestState{
			{ID: "main-quest", Name: "Main Quest", Status: "active"},
		},
	}

	eval, err := engine.Evaluate(ctx, "conditions-campaign", state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundQuestRule := false
	for _, r := range eval.TriggeredRules {
		if r.ID == "rule-quest-active" {
			foundQuestRule = true
			break
		}
	}
	if !foundQuestRule {
		t.Fatalf("expected quest_active rule to trigger when quest is active")
	}

	// Session min condition fails (session 3 < 5)
	foundSessionRule := false
	for _, r := range eval.TriggeredRules {
		if r.ID == "rule-session-min" {
			foundSessionRule = true
			break
		}
	}
	if foundSessionRule {
		t.Fatalf("expected session_min rule NOT to trigger when session < 5")
	}

	// Now session >= 5, should trigger
	state.CurrentSession = 5
	eval, err = engine.Evaluate(ctx, "conditions-campaign", state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundSessionRule = false
	for _, r := range eval.TriggeredRules {
		if r.ID == "rule-session-min" {
			foundSessionRule = true
			break
		}
	}
	if !foundSessionRule {
		t.Fatalf("expected session_min rule to trigger when session >= 5")
	}
}

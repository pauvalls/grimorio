package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestE2E_FactionUpdatePropagationMatrix(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()
	factionRepo := repository.NewMemoryFactionRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "e2e-campaign",
		Entities: []domain.CanonEntity{
			{ID: "dragon-cult", Name: "Dragon Cult", Type: domain.EntityTypeFaction},
			{ID: "harpers", Name: "Harpers", Type: domain.EntityTypeFaction},
		},
		Relationships: []domain.CanonRelationship{
			{ID: "rel-1", From: "dragon-cult", To: "harpers", Type: domain.RelationshipTypeEnemy, Strength: 5},
		},
	}
	_ = canonRepo.Save("e2e-campaign", doc)

	factionSvc := NewFactionService(canonRepo, factionRepo)
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := NewCanonService(canonRepo, stateRepo)
	stateSvc := NewNarrativeStateService(stateRepo, canonRepo)
	validationEngine := NewValidationEngine(canonSvc, stateSvc, factionRepo)
	_ = canonRepo.Save("e2e-campaign", doc)
	_ = stateRepo.Save("e2e-campaign", &domain.NarrativeState{SchemaVersion: domain.SchemaVersionV2, CampaignID: "e2e-campaign"})

	// Step 1: Update reputation with Dragon Cult negatively
	result, err := factionSvc.UpdateReputation(ctx, "e2e-campaign", "dragon-cult", "party-1", -80, "defeated cultists", "combat")
	if err != nil {
		t.Fatalf("update reputation error: %v", err)
	}
	if result.DirectChange.Score != -80 {
		t.Fatalf("direct score = %d, want -80", result.DirectChange.Score)
	}

	// Step 2: Verify propagation to enemy (Harpers should get +40)
	var harperFound bool
	for _, p := range result.PropagatedChanges {
		if p.FactionID == "harpers" {
			harperFound = true
			if p.Score != 40 {
				t.Fatalf("harper propagated score = %d, want 40", p.Score)
			}
		}
	}
	if !harperFound {
		t.Fatalf("expected harpers in propagated changes")
	}

	// Step 3: Get reputation matrix
	matrix, err := factionSvc.GetReputationMatrix(ctx, "e2e-campaign")
	if err != nil {
		t.Fatalf("get matrix error: %v", err)
	}
	if len(matrix.Entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(matrix.Entries))
	}

	// Step 4: Validate that a proposal with hostile faction being helpful triggers gate error
	proposal := domain.ContentProposal{
		ID:      "act-001",
		Type:    "act",
		Content: "The faction Dragon Cult is helpful to the party.",
	}
	report, err := validationEngine.ValidateAct(ctx, "e2e-campaign", "act-001", proposal.Content, nil)
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}

	var gateFailed bool
	for _, check := range report.Checks {
		if check.Rule == "faction_reputation_gate" && !check.Passed {
			gateFailed = true
		}
	}
	if !gateFailed {
		t.Fatalf("expected faction_reputation_gate to fail for hostile faction being helpful")
	}
}

func TestE2E_ConsequenceTriggerEffect(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "e2e-campaign",
		Rules: []domain.CanonRule{
			{
				ID:   "rule-informant",
				Statement: `{"trigger":{"type":"npc_death","entity_id":"informant"},"effects":[{"type":"spawn","target":"npc","description":"spawn replacement"}],"priority":5}`,
				Domain: "consequence",
			},
		},
	}
	_ = canonRepo.Save("e2e-campaign", doc)

	engine := NewConsequenceEngine(canonRepo)
	state := &domain.NarrativeState{
		CampaignID:     "e2e-campaign",
		CurrentSession: 2,
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "informant", Name: "The Informant", Session: 2},
		},
	}

	eval, err := engine.Evaluate(ctx, "e2e-campaign", state)
	if err != nil {
		t.Fatalf("evaluate error: %v", err)
	}
	if len(eval.TriggeredRules) != 1 {
		t.Fatalf("expected 1 triggered rule, got %d", len(eval.TriggeredRules))
	}
	if len(eval.ImmediateEffects) != 1 {
		t.Fatalf("expected 1 immediate effect, got %d", len(eval.ImmediateEffects))
	}
	if eval.ImmediateEffects[0].Description != "spawn replacement" {
		t.Fatalf("effect description = %q, want spawn replacement", eval.ImmediateEffects[0].Description)
	}
}

func TestE2E_RandomTableSeededFromCanon(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "e2e-campaign",
		Facts: []domain.CanonFact{
			{ID: "fact-1", Category: "creature", Statement: "Goblins CR 1/4 lurk in the forest", Source: "lore"},
		},
	}
	_ = canonRepo.Save("e2e-campaign", doc)

	svc := NewRandomTableService(canonRepo)
	tbl, err := svc.GenerateTable(ctx, "e2e-campaign", domain.TableTypeEncounter, domain.TableContext{})
	if err != nil {
		t.Fatalf("generate table error: %v", err)
	}
	if len(tbl.Entries) == 0 {
		t.Fatalf("expected at least one entry")
	}
	if tbl.Entries[0].SourceFact != "fact-1" {
		t.Fatalf("source fact = %q, want fact-1", tbl.Entries[0].SourceFact)
	}
}

func TestE2E_HandoutDualVersion(t *testing.T) {
	ctx := context.Background()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()

	_ = questRepo.Save(&domain.Quest{
		ID:          "quest-001",
		CampaignID:  "e2e-campaign",
		Title:       "The Secret",
		Description: "A noble quest. [DM] The noble is actually a vampire.",
	})

	svc := NewHandoutService(questRepo, canonRepo)
	handout, err := svc.GenerateHandout(ctx, "e2e-campaign", domain.HandoutTypeQuest, []string{"quest-001"}, domain.HandoutVersionBoth)
	if err != nil {
		t.Fatalf("generate handout error: %v", err)
	}
	if handout.PlayerVersion == "" {
		t.Fatalf("expected player version")
	}
	if handout.DMVersion == "" {
		t.Fatalf("expected DM version")
	}
	// Player should not know the secret
	if len(handout.PlayerVersion) >= len(handout.DMVersion) {
		t.Fatalf("player version should be shorter than DM version")
	}
}

package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func setupValidationEngine() (*ValidationEngine, *CanonService, *NarrativeStateService) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := NewCanonService(canonRepo, stateRepo)
	stateSvc := NewNarrativeStateService(stateRepo, canonRepo)
	validator := NewValidationEngine(canonSvc, stateSvc)
	return validator, canonSvc, stateSvc
}

func TestValidationEngine_ValidateAct_DeadNPC(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	// Initialize campaign
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add NPC to canon
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-informador",
		Name:       "El Informador",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	// Mark NPC as dead in narrative state
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-informador",
		Name:    "El Informador",
		Session: 2,
		Cause:   "combat",
	})
	stateSvc.Save(ctx, state)

	report, err := validator.ValidateAct(ctx, "test-campaign", "act-3", "El Informador enters the tavern.", []domain.EntityReference{
		{EntityID: "npc-informador", Location: "act_3"},
	})
	if err != nil {
		t.Fatalf("ValidateAct error: %v", err)
	}

	if report.OverallStatus != "rejected" {
		t.Fatalf("expected rejected, got %s", report.OverallStatus)
	}

	var foundNPCCheck bool
	for _, check := range report.Checks {
		if check.Rule == "npc_alive_check" && !check.Passed {
			foundNPCCheck = true
			break
		}
	}
	if !foundNPCCheck {
		t.Fatal("expected npc_alive_check to fail")
	}
}

func TestValidationEngine_ValidateQuest_MissingEntity(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	report, err := validator.ValidateQuest(ctx, "test-campaign", "quest-1", "Find the lost artifact.", []domain.EntityReference{
		{EntityID: "npc-zarth", Location: "quest_1"},
	})
	if err != nil {
		t.Fatalf("ValidateQuest error: %v", err)
	}

	if report.OverallStatus != "rejected" {
		t.Fatalf("expected rejected, got %s", report.OverallStatus)
	}

	var foundEntityCheck bool
	for _, check := range report.Checks {
		if check.Rule == "entity_not_found" && !check.Passed {
			foundEntityCheck = true
			break
		}
	}
	if !foundEntityCheck {
		t.Fatal("expected entity_not_found to fail")
	}
}

func TestValidationEngine_ValidateAct_LoreViolation(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add lore rule
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Rules = append(doc.Rules, domain.CanonRule{
		ID:        "rule-001",
		Domain:    "magic",
		Statement: "Arcane magic is banned in the city",
	})
	canonSvc.SaveCanon(ctx, doc)

	report, err := validator.ValidateAct(ctx, "test-campaign", "act-3", "The wizards hold a public arcane fair in the city square.", nil)
	if err != nil {
		t.Fatalf("ValidateAct error: %v", err)
	}

	if report.OverallStatus != "rejected" {
		t.Fatalf("expected rejected, got %s", report.OverallStatus)
	}

	var foundLoreCheck bool
	for _, check := range report.Checks {
		if check.Rule == "lore_rule_compliance" && !check.Passed {
			foundLoreCheck = true
			break
		}
	}
	if !foundLoreCheck {
		t.Fatal("expected lore_rule_compliance to fail")
	}
}

func TestValidationEngine_ValidateAct_Approved(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-guard",
		Name:       "City Guard",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	report, err := validator.ValidateAct(ctx, "test-campaign", "act-1", "The party talks to the City Guard.", []domain.EntityReference{
		{EntityID: "npc-guard", Location: "act_1"},
	})
	if err != nil {
		t.Fatalf("ValidateAct error: %v", err)
	}

	if report.OverallStatus != "approved" {
		t.Fatalf("expected approved, got %s", report.OverallStatus)
	}
}

func TestValidationEngine_CheckConsistency_McguffinMissing(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// State says party has the mcguffin but canon mcguffin is not marked as found in key items
	// Actually, mcguffin_continuity checks that if there's a mcguffin in canon,
	// and state has it as a key item, they match.
	// Let's create a scenario where state says they have the mcguffin but it's not in canon entities.
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.KeyItems = append(state.KeyItems, domain.KeyItem{
		ID:           "mcguffin-missing",
		Name:         "The Missing Orb",
		Holder:       "party",
		SessionFound: 1,
		IsMcGuffin:   true,
	})
	stateSvc.Save(ctx, state)

	report, err := validator.CheckConsistency(ctx, "test-campaign", domain.ConsistencyScopeFull)
	if err != nil {
		t.Fatalf("CheckConsistency error: %v", err)
	}

	// There should be at least a warning about mcguffin continuity
	var foundMcguffinCheck bool
	for _, issue := range report.Issues {
		if issue.Rule == "mcguffin_continuity" && !issue.Passed {
			foundMcguffinCheck = true
			break
		}
	}
	if !foundMcguffinCheck {
		t.Fatalf("expected mcguffin_continuity issue, got issues: %+v", report.Issues)
	}
}

func TestValidationEngine_CheckConsistency_Healthy(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add matching mcguffin to state
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.KeyItems = append(state.KeyItems, domain.KeyItem{
		ID:           "mcguffin-test-campaign",
		Name:         "The Artifact McGuffin",
		Holder:       "party",
		SessionFound: 1,
		IsMcGuffin:   true,
	})
	stateSvc.Save(ctx, state)

	report, err := validator.CheckConsistency(ctx, "test-campaign", domain.ConsistencyScopeFull)
	if err != nil {
		t.Fatalf("CheckConsistency error: %v", err)
	}

	// Should be healthy — no criticals or errors
	if report.Criticals > 0 {
		t.Fatalf("expected 0 criticals, got %d", report.Criticals)
	}
	if report.Errors > 0 {
		t.Fatalf("expected 0 errors, got %d", report.Errors)
	}
}

func TestValidationEngine_CheckConsistency_DeadNPCInCanon(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add NPC to canon as alive
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-dead",
		Name:       "Dead Guy",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	// But state says he's dead
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-dead",
		Name:    "Dead Guy",
		Session: 1,
	})
	stateSvc.Save(ctx, state)

	report, err := validator.CheckConsistency(ctx, "test-campaign", domain.ConsistencyScopeFull)
	if err != nil {
		t.Fatalf("CheckConsistency error: %v", err)
	}

	// Should have an npc_alive_check warning/error because canon says alive but state says dead
	var foundNPCCheck bool
	for _, issue := range report.Issues {
		if issue.Rule == "npc_alive_check" && !issue.Passed {
			foundNPCCheck = true
			break
		}
	}
	if !foundNPCCheck {
		t.Fatalf("expected npc_alive_check issue, got issues: %+v", report.Issues)
	}
}

func TestValidationEngine_CheckConsistency_LoreOnlyScope(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add NPC to canon as alive but state says dead
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-dead",
		Name:       "Dead Guy",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-dead",
		Name:    "Dead Guy",
		Session: 1,
	})
	stateSvc.Save(ctx, state)

	// With lore_only scope, the npc_alive_check might still run (our impl runs it regardless)
	// but we test that lore_only scope doesn't crash
	report, err := validator.CheckConsistency(ctx, "test-campaign", domain.ConsistencyScopeLoreOnly)
	if err != nil {
		t.Fatalf("CheckConsistency error: %v", err)
	}

	if report.CampaignID != "test-campaign" {
		t.Fatalf("expected campaign ID test-campaign, got %s", report.CampaignID)
	}
}

func TestValidationEngine_ValidateAct_McguffinViolation(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// State says party has the mcguffin
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.KeyItems = append(state.KeyItems, domain.KeyItem{
		ID:           "mcguffin-test-campaign",
		Name:         "The Artifact McGuffin",
		Holder:       "party",
		SessionFound: 1,
		IsMcGuffin:   true,
	})
	stateSvc.Save(ctx, state)

	report, err := validator.ValidateAct(ctx, "test-campaign", "act-3", "The Artifact McGuffin is found in the villain's lair.", nil)
	if err != nil {
		t.Fatalf("ValidateAct error: %v", err)
	}

	var foundMcguffinCheck bool
	for _, check := range report.Checks {
		if check.Rule == "mcguffin_continuity" && !check.Passed {
			foundMcguffinCheck = true
			break
		}
	}
	if !foundMcguffinCheck {
		t.Fatalf("expected mcguffin_continuity check to fail, got checks: %+v", report.Checks)
	}
}

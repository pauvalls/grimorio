package services

import (
	"context"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func setupValidationEngine() (*ValidationEngine, *CanonService, *NarrativeStateService) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := NewCanonService(canonRepo, stateRepo)
	stateSvc := NewNarrativeStateService(stateRepo, canonRepo)
	validator := NewValidationEngine(canonSvc, stateSvc, nil)
	return validator, canonSvc, stateSvc
}

func TestValidationEngine_ValidateAct_DeadNPC(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	// Initialize campaign
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add NPC to canon
	doc, err := canonSvc.LoadCanon(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-informador",
		Name:       "El Informador",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonSvc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Mark NPC as dead in narrative state
	state, err := stateSvc.Load(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-informador",
		Name:    "El Informador",
		Session: 2,
		Cause:   "combat",
	})
	if err := stateSvc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add lore rule
	doc, err := canonSvc.LoadCanon(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Rules = append(doc.Rules, domain.CanonRule{
		ID:        "rule-001",
		Domain:    "magic",
		Statement: "Arcane magic is banned in the city",
	})
	if err := canonSvc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	doc, err := canonSvc.LoadCanon(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-guard",
		Name:       "City Guard",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonSvc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// State says party has the mcguffin but canon mcguffin is not marked as found in key items
	// Actually, mcguffin_continuity checks that if there's a mcguffin in canon,
	// and state has it as a key item, they match.
	// Let's create a scenario where state says they have the mcguffin but it's not in canon entities.
	state, err := stateSvc.Load(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	state.KeyItems = append(state.KeyItems, domain.KeyItem{
		ID:           "mcguffin-missing",
		Name:         "The Missing Orb",
		Holder:       "party",
		SessionFound: 1,
		IsMcGuffin:   true,
	})
	if err := stateSvc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add matching mcguffin to state
	state, err := stateSvc.Load(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	state.KeyItems = append(state.KeyItems, domain.KeyItem{
		ID:           "mcguffin-test-campaign",
		Name:         "The Artifact McGuffin",
		Holder:       "party",
		SessionFound: 1,
		IsMcGuffin:   true,
	})
	if err := stateSvc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

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

// === New Rules Phase 2 ===

func TestValidationEngine_QuestRewardExistence_Valid(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:   "item-reward",
		Name: "Ring of Invisibility",
		Type: domain.EntityTypeItem,
	})
	canonSvc.SaveCanon(ctx, doc)

	report, err := validator.ValidateQuest(ctx, "test-campaign", "quest-1", "Find the lost cat. Reward: Ring of Invisibility", nil)
	if err != nil {
		t.Fatalf("ValidateQuest error: %v", err)
	}

	// Verify no quest_reward_existence failures
	for _, check := range report.Checks {
		if check.Rule == "quest_reward_existence" && !check.Passed {
			t.Fatalf("expected quest_reward_existence to pass, got: %s", check.Message)
		}
	}
	// Overall should be approved (no critical errors)
	if report.OverallStatus == "rejected" {
		t.Fatalf("expected approved or warning, got rejected. Checks: %+v", report.Checks)
	}
}

func TestValidationEngine_QuestRewardExistence_Missing(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	report, err := validator.ValidateQuest(ctx, "test-campaign", "quest-1", "Find the lost cat. Reward: Sword of Doom", nil)
	if err != nil {
		t.Fatalf("ValidateQuest error: %v", err)
	}

	var foundCheck bool
	for _, check := range report.Checks {
		if check.Rule == "quest_reward_existence" && !check.Passed {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Fatalf("expected quest_reward_existence to fail, got checks: %+v", report.Checks)
	}
}

func TestValidationEngine_LevelEncounterBalance_Valid(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", LevelRange: "1-3", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "enc-001",
		Type:    "encounter",
		Content: "3 goblins (CR 1/4 each)",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	for _, check := range report.Checks {
		if check.Rule == "level_encounter_balance" && !check.Passed {
			t.Fatalf("expected level_encounter_balance to pass, got: %s", check.Message)
		}
	}
	if report.OverallStatus == "rejected" {
		t.Fatalf("expected approved or warning, got rejected. Checks: %+v", report.Checks)
	}
}

func TestValidationEngine_LevelEncounterBalance_TooHard(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", LevelRange: "1-3", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "enc-001",
		Type:    "encounter",
		Content: "Adult Dragon (CR 13)",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	var foundCheck bool
	for _, check := range report.Checks {
		if check.Rule == "level_encounter_balance" && !check.Passed {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Fatalf("expected level_encounter_balance to fail, got checks: %+v", report.Checks)
	}
}

func TestValidationEngine_LocationExistence_Valid(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:   "loc-tavern",
		Name: "The Rusty Nail",
		Type: domain.EntityTypeLocation,
	})
	canonSvc.SaveCanon(ctx, doc)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "act-1",
		Type:    "act",
		Content: "The party arrives at The Rusty Nail.",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	for _, check := range report.Checks {
		if check.Rule == "location_existence" && !check.Passed {
			t.Fatalf("expected location_existence to pass, got: %s", check.Message)
		}
	}
	if report.OverallStatus == "rejected" {
		t.Fatalf("expected approved or warning, got rejected. Checks: %+v", report.Checks)
	}
}

func TestValidationEngine_LocationExistence_Missing(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "act-1",
		Type:    "act",
		Content: "The party arrives at The Golden Dragon Inn.",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	var foundCheck bool
	for _, check := range report.Checks {
		if check.Rule == "location_existence" && !check.Passed {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Fatalf("expected location_existence to fail, got checks: %+v", report.Checks)
	}
}

func TestValidationEngine_TimelineConsistency_Valid(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Timeline = append(doc.Timeline, domain.CanonTimelineEvent{
		ID:          "evt-001",
		Timestamp:   "Year 1",
		Description: "The king is crowned",
	})
	doc.Timeline = append(doc.Timeline, domain.CanonTimelineEvent{
		ID:          "evt-002",
		Timestamp:   "Year 5",
		Description: "The rebellion starts",
	})
	canonSvc.SaveCanon(ctx, doc)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "act-2",
		Type:    "act",
		Content: "Event evt-002 happens after evt-001. Timeline: evt-001, evt-002",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	// Should not fail - this is a basic test; real timeline logic may vary
	// At minimum, no timeline_consistency error should appear
	for _, check := range report.Checks {
		if check.Rule == "timeline_consistency" && !check.Passed {
			t.Fatalf("expected timeline_consistency to pass for valid timeline, got: %s", check.Message)
		}
	}
}

func TestValidationEngine_PrerequisiteClueCheck_MissingClue(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Clue not revealed in state
	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "act-3",
		Type:    "act",
		Content: "Requires secret-password to enter the vault.",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	var foundCheck bool
	for _, check := range report.Checks {
		if check.Rule == "prerequisite_clue_check" && !check.Passed {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Fatalf("expected prerequisite_clue_check to fail, got checks: %+v", report.Checks)
	}
}

func TestValidationEngine_PrerequisiteClueCheck_ClueRevealed(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Reveal the clue in state
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.RevealedClues = append(state.RevealedClues, domain.RevealedClue{
		ID:              "secret-password",
		Description:     "The vault password",
		SourceAct:       "act-2",
		SessionRevealed: 1,
		IsCritical:      true,
	})
	stateSvc.Save(ctx, state)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "act-3",
		Type:    "act",
		Content: "Requires secret-password to enter the vault.",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	for _, check := range report.Checks {
		if check.Rule == "prerequisite_clue_check" && !check.Passed {
			t.Fatalf("expected prerequisite_clue_check to pass when clue is revealed, got: %s", check.Message)
		}
	}
}

func TestValidationEngine_FactionReputationGate_Placeholder(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)
	_ = stateSvc

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:      "act-1",
		Type:    "act",
		Content: "The faction Redbrand reacts hostilely to the party.",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	// faction_reputation_gate is a placeholder - should not produce errors
	for _, check := range report.Checks {
		if check.Rule == "faction_reputation_gate" && !check.Passed {
			// It's OK if it produces warnings but not errors in placeholder mode
			if check.Severity != "warning" {
				t.Fatalf("expected faction_reputation_gate to be no-op or warning in placeholder, got: %+v", check)
			}
		}
	}
}

func TestValidationEngine_CheckConsistency_FullScope(t *testing.T) {
	validator, canonSvc, stateSvc := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", LevelRange: "1-3", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Setup canon with various entities
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:   "item-reward",
		Name: "Gold Coin",
		Type: domain.EntityTypeItem,
	})
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:   "loc-tavern",
		Name: "The Rusty Nail",
		Type: domain.EntityTypeLocation,
	})
	doc.Rules = append(doc.Rules, domain.CanonRule{
		ID:        "rule-001",
		Domain:    "magic",
		Statement: "Necromancy is banned",
	})
	canonSvc.SaveCanon(ctx, doc)

	// Setup state with matching mcguffin
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

	// With a healthy campaign, should be excellent or good
	if report.Criticals > 0 {
		t.Fatalf("expected 0 criticals in healthy campaign, got %d", report.Criticals)
	}
	if report.Errors > 0 {
		t.Fatalf("expected 0 errors in healthy campaign, got %d", report.Errors)
	}
}

func TestValidationEngine_FactionContext(t *testing.T) {
	validator, canonSvc, _ := setupValidationEngine()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	report, err := validator.validate(ctx, "test-campaign", domain.ContentProposal{
		ID:             "act-1",
		Type:           "act",
		Content:        "The party meets with faction leaders.",
		FactionContext: "diplomatic-summit",
	})
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	var foundContextCheck bool
	for _, check := range report.Checks {
		if check.Rule == "faction_context" && check.Passed {
			foundContextCheck = true
			if !strings.Contains(check.Message, "diplomatic-summit") {
				t.Fatalf("expected faction_context to mention 'diplomatic-summit', got: %s", check.Message)
			}
			break
		}
	}
	if !foundContextCheck {
		t.Fatalf("expected faction_context check to pass, got checks: %+v", report.Checks)
	}
}

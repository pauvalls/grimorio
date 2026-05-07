package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func setupGateService(t *testing.T) (*ConsistencyGateService, *repository.MemoryCanonRepository, *repository.MemoryNarrativeStateRepository) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := NewCanonService(canonRepo, stateRepo)
	stateSvc := NewNarrativeStateService(stateRepo, canonRepo)
	validator := NewValidationEngine(canonSvc, stateSvc)

	gateSvc := NewConsistencyGateService(canonSvc, stateSvc, validator)
	return gateSvc, canonRepo, stateRepo
}

func setupCampaign(t *testing.T, canonSvc *CanonService, campaignID string) {
	ctx := context.Background()
	brief := domain.CampaignBrief{
		Name:         campaignID,
		LevelRange:   "1-3",
		Tone:         "grim",
		SettingType:  "urban",
		VillainType:  "lich",
		McGuffinType: "amulet",
	}
	_, err := canonSvc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}
}

func TestConsistencyGateService_ProcessBatch_Approve(t *testing.T) {
	ctx := context.Background()
	gateSvc, canonRepo, _ := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	// Add an NPC to canon
	doc, err := canonRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonRepo.Save("test-campaign", doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	proposal := domain.BatchProposal{
		BatchID:    "batch-001",
		CampaignID: "test-campaign",
		Artifacts: []domain.ContentProposal{
			{
				ID:      "npc-001-ref",
				Type:    "npc",
				Content: "The NPC Test NPC is present in the tavern.",
				EntityReferences: []domain.EntityReference{
					{EntityID: "npc-001", RequiredState: domain.EntityStateAlive},
				},
			},
		},
		Attempt: 1,
	}

	result, err := gateSvc.ProcessBatch(ctx, proposal, false)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
	if result.Status != domain.GateStatusApproved {
		t.Fatalf("expected status approved, got %s", result.Status)
	}
	if !result.CanonUpdated {
		t.Fatal("expected canon_updated to be true")
	}
	if !result.StateUpdated {
		t.Fatal("expected state_updated to be true")
	}
}

func TestConsistencyGateService_ProcessBatch_Reject(t *testing.T) {
	ctx := context.Background()
	gateSvc, canonRepo, stateRepo := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	// Add an NPC to canon
	doc, err := canonRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonRepo.Save("test-campaign", doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Kill the NPC in narrative state
	state, err := stateRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-001",
		Name:    "Test NPC",
		Session: 1,
	})
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	proposal := domain.BatchProposal{
		BatchID:    "batch-001",
		CampaignID: "test-campaign",
		Artifacts: []domain.ContentProposal{
			{
				ID:      "npc-001-ref",
				Type:    "npc",
				Content: "The NPC Test NPC gives a quest.",
				EntityReferences: []domain.EntityReference{
					{EntityID: "npc-001", RequiredState: domain.EntityStateAlive},
				},
			},
		},
		Attempt: 1,
	}

	result, err := gateSvc.ProcessBatch(ctx, proposal, false)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
	if result.Status != domain.GateStatusRejected {
		t.Fatalf("expected status rejected, got %s", result.Status)
	}
	if len(result.Suggestions) == 0 {
		t.Fatal("expected suggestions for rejected batch")
	}
	if result.RetryPrompt == "" {
		t.Fatal("expected retry prompt for rejected batch")
	}
	if result.CanonUpdated {
		t.Fatal("expected canon_updated to be false for rejected batch")
	}
}

func TestConsistencyGateService_ProcessBatch_LockPreventsConcurrent(t *testing.T) {
	ctx := context.Background()
	gateSvc, canonRepo, _ := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	// Add an NPC
	doc, err := canonRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonRepo.Save("test-campaign", doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Manually acquire lock
	gateSvc.acquireLock("test-campaign", "batch-000")

	proposal := domain.BatchProposal{
		BatchID:    "batch-001",
		CampaignID: "test-campaign",
		Artifacts: []domain.ContentProposal{
			{ID: "npc-001-ref", Type: "npc", Content: "Test", EntityReferences: []domain.EntityReference{{EntityID: "npc-001"}}},
		},
		Attempt: 1,
	}

	_, err = gateSvc.ProcessBatch(ctx, proposal, false)
	if err == nil {
		t.Fatal("expected error due to lock, got nil")
	}

	// Release lock and retry
	gateSvc.releaseLock("test-campaign", "batch-000")
	result, err := gateSvc.ProcessBatch(ctx, proposal, false)
	if err != nil {
		t.Fatalf("ProcessBatch failed after releasing lock: %v", err)
	}
	if result.Status != domain.GateStatusApproved {
		t.Fatalf("expected approved after releasing lock, got %s", result.Status)
	}
}

func TestConsistencyGateService_GetGateStatus(t *testing.T) {
	ctx := context.Background()
	gateSvc, canonRepo, _ := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	doc, _ := canonRepo.Load("test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonRepo.Save("test-campaign", doc)

	proposal := domain.BatchProposal{
		BatchID:    "batch-001",
		CampaignID: "test-campaign",
		Artifacts: []domain.ContentProposal{
			{ID: "npc-001-ref", Type: "npc", Content: "Test", EntityReferences: []domain.EntityReference{{EntityID: "npc-001"}}},
		},
		Attempt: 1,
	}

	_, err := gateSvc.ProcessBatch(ctx, proposal, false)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	status, err := gateSvc.GetGateStatus(ctx, "test-campaign", "batch-001")
	if err != nil {
		t.Fatalf("GetGateStatus failed: %v", err)
	}
	if status.Status != domain.GateStatusApproved {
		t.Fatalf("expected approved status, got %s", status.Status)
	}
}

func TestConsistencyGateService_ResetGate(t *testing.T) {
	ctx := context.Background()
	gateSvc, _, _ := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	// Acquire a lock manually
	gateSvc.acquireLock("test-campaign", "batch-001")

	err := gateSvc.ResetGate(ctx, "test-campaign", "batch-001")
	if err != nil {
		t.Fatalf("ResetGate failed: %v", err)
	}

	// Verify lock is released
	if gateSvc.isLocked("test-campaign") {
		t.Fatal("expected lock to be released after reset")
	}
}

func TestConsistencyGateService_MaxRetries(t *testing.T) {
	ctx := context.Background()
	gateSvc, canonRepo, stateRepo := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	// Setup: dead NPC referenced as alive
	doc, err := canonRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonRepo.Save("test-campaign", doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	state, err := stateRepo.Load("test-campaign")
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-001",
		Name:    "Test NPC",
		Session: 1,
	})
	if err := stateRepo.Save("test-campaign", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	proposal := domain.BatchProposal{
		BatchID:    "batch-001",
		CampaignID: "test-campaign",
		Artifacts: []domain.ContentProposal{
			{
				ID:      "npc-001-ref",
				Type:    "npc",
				Content: "The NPC Test NPC gives a quest.",
				EntityReferences: []domain.EntityReference{
					{EntityID: "npc-001", RequiredState: domain.EntityStateAlive},
				},
			},
		},
		Attempt: 3, // Max retries exceeded
	}

	result, err := gateSvc.ProcessBatch(ctx, proposal, false)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
	if result.Status != domain.GateStatusRejected {
		t.Fatalf("expected rejected, got %s", result.Status)
	}
	if !containsStr(result.RetryPrompt, "Human review required") {
		t.Fatalf("expected retry prompt to mention human review, got: %s", result.RetryPrompt)
	}
}

func TestConsistencyGateService_AutoSave(t *testing.T) {
	ctx := context.Background()
	gateSvc, canonRepo, stateRepo := setupGateService(t)
	setupCampaign(t, gateSvc.canonSvc, "test-campaign")

	// Record initial timestamps
	initialDoc, _ := canonRepo.Load("test-campaign")
	initialState, _ := stateRepo.Load("test-campaign")
	initialDocTime := initialDoc.UpdatedAt
	initialStateTime := initialState.LastUpdated

	// Add an NPC
	doc, _ := canonRepo.Load("test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Test NPC",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonRepo.Save("test-campaign", doc)

	proposal := domain.BatchProposal{
		BatchID:    "batch-001",
		CampaignID: "test-campaign",
		Artifacts: []domain.ContentProposal{
			{ID: "npc-001-ref", Type: "npc", Content: "Test", EntityReferences: []domain.EntityReference{{EntityID: "npc-001"}}},
		},
		Attempt: 1,
	}

	_, err := gateSvc.ProcessBatch(ctx, proposal, false)
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}

	// Verify timestamps updated
	updatedDoc, _ := canonRepo.Load("test-campaign")
	updatedState, _ := stateRepo.Load("test-campaign")
	if !updatedDoc.UpdatedAt.After(initialDocTime) {
		t.Fatal("expected canon UpdatedAt to be updated after approval")
	}
	if !updatedState.LastUpdated.After(initialStateTime) {
		t.Fatal("expected state LastUpdated to be updated after approval")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func setupCanonService() (*CanonService, *repository.MemoryCanonRepository, *repository.MemoryNarrativeStateRepository) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	svc := NewCanonService(canonRepo, stateRepo)
	return svc, canonRepo, stateRepo
}

func TestCanonService_InitializeCanon(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{
		Name:         "shadows-of-thornvale",
		LevelRange:   "1-5",
		Tone:         "dark",
		SettingType:  "gothic",
		Themes:       []string{"corruption", "redemption"},
		VillainType:  "lich",
		McGuffinType: "artifact",
	}

	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	if doc.SchemaVersion != domain.SchemaVersionV2 {
		t.Fatalf("expected schema version %s, got %s", domain.SchemaVersionV2, doc.SchemaVersion)
	}

	if doc.CampaignID != brief.Name {
		t.Fatalf("expected campaign ID %s, got %s", brief.Name, doc.CampaignID)
	}

	if len(doc.Facts) == 0 {
		t.Fatal("expected at least one lore fact")
	}

	if len(doc.Entities) == 0 {
		t.Fatal("expected at least one entity (mcguffin)")
	}

	if doc.Entities[0].Role != "mcguffin" {
		t.Fatalf("expected first entity to be mcguffin, got %s", doc.Entities[0].Role)
	}
}

func TestCanonService_InitializeCanon_InvalidName(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "Invalid Name"}
	_, err := svc.InitializeCanon(ctx, brief)
	if err == nil {
		t.Fatal("expected error for invalid campaign name")
	}
}

func TestCanonService_LoadSaveCanon(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	// Initialize first
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Modify and save
	doc.Facts = append(doc.Facts, domain.CanonFact{
		ID:        "fact-002",
		Category:  "history",
		Statement: "The ancient kingdom fell 500 years ago",
		Immutable: true,
	})

	if err := svc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Load and verify
	loaded, err := svc.LoadCanon(ctx, brief.Name)
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}

	if len(loaded.Facts) != 3 {
		t.Fatalf("expected 3 facts (v3.0 adds default fact), got %d", len(loaded.Facts))
	}
}

func TestCanonService_RegisterFact(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := svc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	fact := domain.CanonFact{
		ID:        "fact-002",
		Category:  "politics",
		Statement: "The king is secretly a vampire",
		Immutable: true,
	}

	if err := svc.RegisterFact(ctx, brief.Name, fact); err != nil {
		t.Fatalf("failed to register fact: %v", err)
	}

	doc, err := svc.LoadCanon(ctx, brief.Name)
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	if len(doc.Facts) != 3 {
		t.Fatalf("expected 3 facts (v3.0 adds default fact), got %d", len(doc.Facts))
	}
}

func TestCanonService_QueryEntity(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add more entities
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Lord Vex",
		Type:       domain.EntityTypeNPC,
		Role:       "ally",
		CanonState: domain.EntityStateAlive,
	}, domain.CanonEntity{
		ID:         "npc-002",
		Name:       "Dark Lady",
		Type:       domain.EntityTypeNPC,
		Role:       "villain",
		CanonState: domain.EntityStateAlive,
	})
	if err := svc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Query by type
	npcs, err := svc.QueryEntity(ctx, brief.Name, domain.EntityFilter{Type: domain.EntityTypeNPC})
	if err != nil {
		t.Fatalf("failed to query entities: %v", err)
	}
	if len(npcs) != 2 {
		t.Fatalf("expected 2 NPCs, got %d", len(npcs))
	}

	// Query by role
	villains, err := svc.QueryEntity(ctx, brief.Name, domain.EntityFilter{Role: "villain"})
	if err != nil {
		t.Fatalf("failed to query entities: %v", err)
	}
	if len(villains) != 1 {
		t.Fatalf("expected 1 villain, got %d", len(villains))
	}

	// Query by name
	results, err := svc.QueryEntity(ctx, brief.Name, domain.EntityFilter{NameQuery: "vex"})
	if err != nil {
		t.Fatalf("failed to query entities: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'vex', got %d", len(results))
	}
}

func TestCanonService_UpdateEntityState(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-001",
		Name:       "Lord Vex",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := svc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	if err := svc.UpdateEntityState(ctx, brief.Name, "npc-001", domain.EntityStateDead); err != nil {
		t.Fatalf("failed to update entity state: %v", err)
	}

	loaded, err := svc.LoadCanon(ctx, brief.Name)
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	for _, e := range loaded.Entities {
		if e.ID == "npc-001" {
			if e.CanonState != domain.EntityStateDead {
				t.Fatalf("expected state dead, got %s", e.CanonState)
			}
			return
		}
	}
	t.Fatal("entity not found after update")
}

func TestCanonService_ValidateProposal(t *testing.T) {
	svc, _, stateRepo := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := svc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add an NPC and a rule to the canon
	doc, err := svc.LoadCanon(ctx, brief.Name)
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-informador",
		Name:       "El Informador",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	doc.Rules = append(doc.Rules, domain.CanonRule{
		ID:        "rule-001",
		Domain:    "magic",
		Statement: "Arcane magic is banned in the city",
	})
	if err := svc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Set NPC as dead in narrative state
	state, err := stateRepo.Load(brief.Name)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	state.DeadNPCs = append(state.DeadNPCs, domain.NPCDeathRecord{
		NPCID:   "npc-informador",
		Name:    "El Informador",
		Session: 2,
		Cause:   "combat",
	})
	if err := stateRepo.Save(brief.Name, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Test 1: Valid proposal referencing existing entity
	proposal1 := domain.ContentProposal{
		ID:      "act-3-draft",
		Type:    "act",
		Content: "The party enters the city.",
		EntityReferences: []domain.EntityReference{
			{EntityID: "npc-informador", Location: "act_3, area_12"},
		},
	}
	report1, err := svc.ValidateProposal(ctx, brief.Name, proposal1)
	if err != nil {
		t.Fatalf("failed to validate proposal: %v", err)
	}
	if report1.OverallStatus != "rejected" {
		t.Fatalf("expected rejected due to dead NPC, got %s", report1.OverallStatus)
	}

	// Check that npc_alive_check failed
	var foundAliveCheck bool
	for _, check := range report1.Checks {
		if check.Rule == "npc_alive_check" && !check.Passed {
			foundAliveCheck = true
			break
		}
	}
	if !foundAliveCheck {
		t.Fatal("expected npc_alive_check to fail")
	}

	// Test 2: Missing entity
	proposal2 := domain.ContentProposal{
		ID:      "quest-1",
		Type:    "quest",
		Content: "Find the lost artifact.",
		EntityReferences: []domain.EntityReference{
			{EntityID: "npc-zarth", Location: "quest_1"},
		},
	}
	report2, _ := svc.ValidateProposal(ctx, brief.Name, proposal2)
	if report2.OverallStatus != "rejected" {
		t.Fatalf("expected rejected due to missing entity, got %s", report2.OverallStatus)
	}

	// Test 3: Lore rule violation
	proposal3 := domain.ContentProposal{
		ID:      "act-3",
		Type:    "act",
		Content: "The wizards hold a public arcane fair in the city square.",
	}
	report3, _ := svc.ValidateProposal(ctx, brief.Name, proposal3)
	if report3.OverallStatus != "rejected" {
		t.Fatalf("expected rejected due to lore violation, got %s", report3.OverallStatus)
	}
}

func TestCanonService_GetRelationshipGraph(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := svc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	graph, err := svc.GetRelationshipGraph(ctx, brief.Name)
	if err != nil {
		t.Fatalf("failed to get relationship graph: %v", err)
	}

	if graph.CampaignID != brief.Name {
		t.Fatalf("expected campaign ID %s, got %s", brief.Name, graph.CampaignID)
	}
}

func TestCanonService_CacheHit(t *testing.T) {
	svc, canonRepo, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "cache-test", McGuffinType: "artifact"}
	if _, err := svc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// First load — cache miss
	_, err := svc.LoadCanon(ctx, brief.Name)
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}

	// Corrupt repo directly to prove second load comes from cache
	canonRepo.Delete(brief.Name)

	// Second load — should be cache hit
	loaded, err := svc.LoadCanon(ctx, brief.Name)
	if err != nil {
		t.Fatalf("expected cache hit, got error: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected cached document, got nil")
	}
	if loaded.CampaignID != brief.Name {
		t.Fatalf("expected campaign ID %s, got %s", brief.Name, loaded.CampaignID)
	}
}

func TestCanonService_CacheInvalidateOnSave(t *testing.T) {
	svc, canonRepo, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "invalidate-test", McGuffinType: "artifact"}
	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Load to warm cache
	if _, err := svc.LoadCanon(ctx, brief.Name); err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}

	// Save modified doc
	doc.Facts = append(doc.Facts, domain.CanonFact{ID: "fact-002", Category: "test", Statement: "test"})
	if err := svc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Corrupt repo to prove cache was invalidated
	canonRepo.Delete(brief.Name)

	// Should fail because cache was invalidated and repo is corrupted
	_, err = svc.LoadCanon(ctx, brief.Name)
	if err == nil {
		t.Fatal("expected error after cache invalidation, got nil")
	}
}

func TestCanonService_CacheInvalidateOnRegisterFact(t *testing.T) {
	svc, canonRepo, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "fact-invalidate", McGuffinType: "artifact"}
	if _, err := svc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Warm cache
	if _, err := svc.LoadCanon(ctx, brief.Name); err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}

	// Register fact
	fact := domain.CanonFact{ID: "fact-002", Category: "test", Statement: "test"}
	if err := svc.RegisterFact(ctx, brief.Name, fact); err != nil {
		t.Fatalf("failed to register fact: %v", err)
	}

	// Corrupt repo
	canonRepo.Delete(brief.Name)

	// Should fail — cache invalidated
	_, err := svc.LoadCanon(ctx, brief.Name)
	if err == nil {
		t.Fatal("expected error after RegisterFact invalidation")
	}
}

func TestCanonService_CacheInvalidateOnUpdateEntityState(t *testing.T) {
	svc, canonRepo, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "entity-invalidate", McGuffinType: "artifact"}
	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}
	doc.Entities = append(doc.Entities, domain.CanonEntity{ID: "npc-001", Name: "Test", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive})
	if err := svc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Warm cache
	if _, err := svc.LoadCanon(ctx, brief.Name); err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}

	// Update entity state
	if err := svc.UpdateEntityState(ctx, brief.Name, "npc-001", domain.EntityStateDead); err != nil {
		t.Fatalf("failed to update entity state: %v", err)
	}

	// Corrupt repo
	canonRepo.Delete(brief.Name)

	// Should fail — cache invalidated
	_, err = svc.LoadCanon(ctx, brief.Name)
	if err == nil {
		t.Fatal("expected error after UpdateEntityState invalidation")
	}
}

func TestCanonService_DegradedMode_LoadCanon(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	// No canon initialized — set degraded
	svc.SetDegraded(true)

	doc, err := svc.LoadCanon(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("expected no error in degraded mode, got %v", err)
	}
	if doc == nil {
		t.Fatal("expected empty document in degraded mode, got nil")
	}
	if doc.CampaignID != "nonexistent" {
		t.Fatalf("expected campaign ID 'nonexistent', got %s", doc.CampaignID)
	}
}

func TestCanonService_DegradedMode_ValidateProposal(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	svc.SetDegraded(true)

	proposal := domain.ContentProposal{
		ID:      "test-proposal",
		Type:    "act",
		Content: "Test content",
	}

	report, err := svc.ValidateProposal(ctx, "any-campaign", proposal)
	if err != nil {
		t.Fatalf("expected no error in degraded mode, got %v", err)
	}
	if report == nil {
		t.Fatal("expected report in degraded mode, got nil")
	}
	if report.OverallStatus != "approved" {
		t.Fatalf("expected auto-approved in degraded mode, got %s", report.OverallStatus)
	}
}

func TestCanonService_DegradedMode_QueryEntity(t *testing.T) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	svc.SetDegraded(true)

	results, err := svc.QueryEntity(ctx, "any-campaign", domain.EntityFilter{Type: domain.EntityTypeNPC})
	if err != nil {
		t.Fatalf("expected no error in degraded mode, got %v", err)
	}
	if results == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results in degraded mode, got %d", len(results))
	}
}

func TestCanonService_IsDegraded(t *testing.T) {
	svc, _, _ := setupCanonService()

	if svc.IsDegraded() {
		t.Fatal("expected degraded to be false by default")
	}

	svc.SetDegraded(true)
	if !svc.IsDegraded() {
		t.Fatal("expected degraded to be true after SetDegraded(true)")
	}

	svc.SetDegraded(false)
	if svc.IsDegraded() {
		t.Fatal("expected degraded to be false after SetDegraded(false)")
	}
}

func BenchmarkCanonService_LoadCanon(b *testing.B) {
	svc, _, _ := setupCanonService()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "bench-campaign", McGuffinType: "artifact"}
	doc, err := svc.InitializeCanon(ctx, brief)
	if err != nil {
		b.Fatalf("failed to initialize canon: %v", err)
	}

	// Add 50 entities
	for i := 0; i < 50; i++ {
		doc.Entities = append(doc.Entities, domain.CanonEntity{
			ID:         fmt.Sprintf("npc-%03d", i),
			Name:       fmt.Sprintf("NPC %d", i),
			Type:       domain.EntityTypeNPC,
			Role:       "ally",
			CanonState: domain.EntityStateAlive,
		})
	}
	if err := svc.SaveCanon(ctx, doc); err != nil {
		b.Fatalf("failed to save canon: %v", err)
	}

	// Warm cache
	if _, err := svc.LoadCanon(ctx, brief.Name); err != nil {
		b.Fatalf("failed to load canon: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.LoadCanon(ctx, brief.Name)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

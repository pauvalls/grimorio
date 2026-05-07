package handlers

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupTestCanonHandlers() (*CanonHandlers, *services.CanonService, *services.NarrativeStateService, *services.ValidationEngine) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonService := services.NewCanonService(canonRepo, stateRepo)
	stateService := services.NewNarrativeStateService(stateRepo, canonRepo)
	validator := services.NewValidationEngine(canonService, stateService, nil)
	gateService := services.NewConsistencyGateService(canonService, stateService, validator)
	return NewCanonHandlers(canonService, stateService, validator, gateService), canonService, stateService, validator
}

func TestHandleGenerateAdventureBible(t *testing.T) {
	handlers, _, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	handler := handlers.HandleGenerateAdventureBible()
	args := map[string]any{
		"campaign_id":   "test-campaign",
		"name":          "test-campaign",
		"level_range":   "1-5",
		"tone":          "dark",
		"setting_type":  "gothic",
		"themes":        []interface{}{"corruption"},
		"villain_type":  "lich",
		"mcguffin_type": "artifact",
	}

	result, err := handler(ctx, newToolRequest("generate_adventure_bible", args))
	if err != nil {
		t.Fatalf("HandleGenerateAdventureBible error: %v", err)
	}
	if result.IsError {
		t.Fatalf("HandleGenerateAdventureBible returned error: %v", result.Content)
	}
}

func TestHandleGenerateAdventureBible_MissingArgs(t *testing.T) {
	handlers, _, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	handler := handlers.HandleGenerateAdventureBible()
	args := map[string]any{
		"campaign_id": "test-campaign",
	}

	result, err := handler(ctx, newToolRequest("generate_adventure_bible", args))
	if err != nil {
		t.Fatalf("HandleGenerateAdventureBible error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing required args")
	}
}

func TestHandleValidateCanon(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	// Initialize canon first
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Add an entity
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-guard",
		Name:       "City Guard",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	canonSvc.SaveCanon(ctx, doc)

	handler := handlers.HandleValidateCanon()
	args := map[string]any{
		"campaign_id": "test-campaign",
		"proposal": map[string]any{
			"id":      "act-1",
			"type":    "act",
			"content": "The party meets the City Guard.",
			"entity_references": []interface{}{
				map[string]any{"entity_id": "npc-guard", "location": "act_1"},
			},
		},
	}

	result, err := handler(ctx, newToolRequest("validate_canon", args))
	if err != nil {
		t.Fatalf("HandleValidateCanon error: %v", err)
	}
	if result.IsError {
		t.Fatalf("HandleValidateCanon returned error: %v", result.Content)
	}
}

func TestHandleUpdateNarrativeState(t *testing.T) {
	handlers, canonSvc, stateSvc, _ := setupTestCanonHandlers()
	ctx := context.Background()

	// Initialize campaign
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	// Set up initial state with an active quest
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.ActiveQuests = append(state.ActiveQuests, domain.QuestState{
		ID: "q-001", Name: "Find the Sword", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver",
	})
	stateSvc.Save(ctx, state)

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id":       "test-campaign",
		"session_num":       float64(2),
		"revealed_clues":    []interface{}{},
		"dead_npcs":         []interface{}{},
		"completed_quests":  []interface{}{"q-001"},
		"new_quests":        []interface{}{},
		"key_decisions":     []interface{}{},
		"xp_awarded":        float64(500),
		"loot_acquired":     []interface{}{"gold-100"},
		"session_summary":   "The party found the sword.",
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if result.IsError {
		t.Fatalf("HandleUpdateNarrativeState returned error: %v", result.Content)
	}
}

func TestHandleCheckConsistency(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	canonSvc.InitializeCanon(ctx, brief)

	handler := handlers.HandleCheckConsistency()
	args := map[string]any{
		"campaign_id":       "test-campaign",
		"scope":             "full",
		"severity_threshold": "info",
	}

	result, err := handler(ctx, newToolRequest("check_consistency", args))
	if err != nil {
		t.Fatalf("HandleCheckConsistency error: %v", err)
	}
	if result.IsError {
		t.Fatalf("HandleCheckConsistency returned error: %v", result.Content)
	}
}

func TestHandleCheckConsistency_MissingCampaign(t *testing.T) {
	handlers, _, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	handler := handlers.HandleCheckConsistency()
	args := map[string]any{
		"scope": "full",
	}

	result, err := handler(ctx, newToolRequest("check_consistency", args))
	if err != nil {
		t.Fatalf("HandleCheckConsistency error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing campaign")
	}
}

func TestHandleValidateCanon_MissingCampaign(t *testing.T) {
	handlers, _, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	handler := handlers.HandleValidateCanon()
	args := map[string]any{
		"proposal_id":   "act-1",
		"proposal_type": "act",
		"content":       "test",
	}

	result, err := handler(ctx, newToolRequest("validate_canon", args))
	if err != nil {
		t.Fatalf("HandleValidateCanon error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing campaign_id")
	}
}

func TestHandleUpdateNarrativeState_MissingCampaign(t *testing.T) {
	handlers, _, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"session_num": float64(1),
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing campaign_id")
	}
}

func TestHandleUpdateNarrativeState_InvalidSession(t *testing.T) {
	handlers, _, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id": "test-campaign",
		"session_num": float64(0),
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for invalid session_num")
	}
}

func TestGetBoolArg(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		key      string
		expected bool
	}{
		{"true value", map[string]any{"flag": true}, "flag", true},
		{"false value", map[string]any{"flag": false}, "flag", false},
		{"missing key", map[string]any{}, "flag", false},
		{"wrong type", map[string]any{"flag": "true"}, "flag", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolArg(tt.args, tt.key)
			if result != tt.expected {
				t.Errorf("getBoolArg() = %v, want %v", result, tt.expected)
			}
		})
	}
}

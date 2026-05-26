package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupTestCanonHandlers() (*CanonHandlers, *services.CanonService, *services.NarrativeStateService, *services.ValidationEngine) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonService := services.NewCanonService(canonRepo, stateRepo)
	stateService := services.NewNarrativeStateService(stateRepo, canonRepo)
	validator := services.NewValidationEngine(canonService, stateService, nil, "")
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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add an entity
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-guard",
		Name:       "City Guard",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonSvc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Set up initial state with an active quest
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.ActiveQuests = append(state.ActiveQuests, domain.QuestState{
		ID: "q-001", Name: "Find the Sword", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver",
	})
	if err := stateSvc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

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
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

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
	handlers, canonSvc, stateSvc, _ := setupTestCanonHandlers()
	ctx := context.Background()

	// Initialize campaign
	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Set up initial state with current session 2
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.CurrentSession = 2
	state.ActiveQuests = append(state.ActiveQuests, domain.QuestState{
		ID: "q-001", Name: "Find the Sword", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver",
	})
	if err := stateSvc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id":     "test-campaign",
		"session_num":     float64(0),
		"session_summary": "Auto-incremented session.",
		"completed_quests": []interface{}{"q-001"},
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success for session_num=0 (auto-increment), got error: %v", result.Content)
	}
}

func TestHandleUpdateNarrativeState_StringArrays(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id":      "test-campaign",
		"session_num":      float64(1),
		"revealed_clues":   []interface{}{"The well water is poisoned", "The mayor is a werewolf"},
		"dead_npcs":        []interface{}{"Village Guard", "Wolf"},
		"key_decisions":    []interface{}{"Players spared the mayor", "Party took the left path"},
		"active_quests":    []interface{}{"Find the Cure", "Save the Town"},
		"key_items":        []interface{}{"Silver Sword", "Healing Potion"},
		"session_summary":  "The party explored the village and fought wolves.",
		"xp_awarded":       float64(500),
		"loot_acquired":    []interface{}{"50 gold", "Health Potion"},
		"dm_notes":         "Players were cautious but effective.",
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify the result mentions the counts
	if len(result.Content) == 0 {
		t.Fatal("expected result content")
	}
	resultText := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(resultText, "2") {
		t.Error("expected result to mention 2 revealed clues")
	}
}

func TestHandleUpdateNarrativeState_StringSlice(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	// Use []string instead of []interface{} to simulate real MCP behavior
	args := map[string]any{
		"campaign_id":     "test-campaign",
		"session_num":     float64(1),
		"revealed_clues":  []string{"The well water is poisoned", "The mayor is a werewolf"},
		"dead_npcs":       []string{"Village Guard"},
		"key_decisions":   []string{"Players spared the mayor"},
		"active_quests":   []string{"Find the Cure"},
		"key_items":       []string{"Silver Sword"},
		"loot_acquired":   []string{"50 gold"},
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify the result mentions the counts
	if len(result.Content) == 0 {
		t.Fatal("expected result content")
	}
	resultText := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(resultText, "2") {
		t.Error("expected result to mention 2 revealed clues from []string")
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

func TestStableID(t *testing.T) {
	// Same seed returns equal
	id1 := stableID("clue", "campaign-test|The well is poisoned")
	id2 := stableID("clue", "campaign-test|The well is poisoned")
	if id1 != id2 {
		t.Fatalf("expected same ID for same seed, got %s vs %s", id1, id2)
	}

	// Different seeds return different
	id3 := stableID("clue", "campaign-test|The mayor is a werewolf")
	if id1 == id3 {
		t.Fatalf("expected different IDs for different seeds, got %s", id1)
	}

	// Different prefixes return different
	id4 := stableID("npc", "campaign-test|The well is poisoned")
	if id1 == id4 {
		t.Fatalf("expected different IDs for different prefixes, got %s", id1)
	}

	// Format check: prefix-8hex
	if !strings.HasPrefix(id1, "clue-") {
		t.Fatalf("expected prefix 'clue-', got %s", id1)
	}
	hexPart := strings.TrimPrefix(id1, "clue-")
	if len(hexPart) != 8 {
		t.Fatalf("expected 8 hex chars, got %d in %s", len(hexPart), hexPart)
	}
}

func TestHandleUpdateNarrativeState_EmptyClueID(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id":    "test-campaign",
		"session_num":    float64(1),
		"revealed_clues": []any{map[string]any{"id": "", "description": "No ID clue"}},
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for empty clue ID")
	}
}

func TestHandleUpdateNarrativeState_EmptyNPCID(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id": "test-campaign",
		"session_num": float64(1),
		"dead_npcs":   []any{map[string]any{"npc_id": "", "name": "No ID NPC"}},
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for empty NPC ID")
	}
}

func TestHandleUpdateNarrativeState_EmptyQuestID(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id":   "test-campaign",
		"session_num":   float64(1),
		"active_quests": []any{map[string]any{"id": "", "name": "No ID Quest"}},
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for empty quest ID")
	}
}

func TestHandleUpdateNarrativeState_EmptyItemID(t *testing.T) {
	handlers, canonSvc, _, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id": "test-campaign",
		"session_num": float64(1),
		"key_items":   []any{map[string]any{"id": "", "name": "No ID Item"}},
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for empty item ID")
	}
}

func TestHandleUpdateNarrativeState_SyncToCanon(t *testing.T) {
	handlers, canonSvc, stateSvc, _ := setupTestCanonHandlers()
	ctx := context.Background()

	brief := domain.CampaignBrief{Name: "test-campaign", McGuffinType: "artifact"}
	if _, err := canonSvc.InitializeCanon(ctx, brief); err != nil {
		t.Fatalf("failed to initialize canon: %v", err)
	}

	// Add an NPC entity to canon
	doc, _ := canonSvc.LoadCanon(ctx, "test-campaign")
	doc.Entities = append(doc.Entities, domain.CanonEntity{
		ID:         "npc-villain",
		Name:       "Lord Dark",
		Type:       domain.EntityTypeNPC,
		CanonState: domain.EntityStateAlive,
	})
	if err := canonSvc.SaveCanon(ctx, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	// Set up state with a quest
	state, _ := stateSvc.Load(ctx, "test-campaign")
	state.ActiveQuests = append(state.ActiveQuests, domain.QuestState{
		ID: "q-001", Name: "Find the Sword", Status: "active", SourceAct: "act-1", GiverNPC: "npc-giver",
	})
	if err := stateSvc.Save(ctx, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	handler := handlers.HandleUpdateNarrativeState()
	args := map[string]any{
		"campaign_id":      "test-campaign",
		"session_num":      float64(2),
		"completed_quests": []any{"q-001"},
		"dead_npcs":        []any{map[string]any{"npc_id": "npc-villain", "name": "Lord Dark"}},
		"sync_to_canon":    true,
		"session_summary":  "The party defeated Lord Dark.",
	}

	result, err := handler(ctx, newToolRequest("update_narrative_state", args))
	if err != nil {
		t.Fatalf("HandleUpdateNarrativeState error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify canon updated: NPC is dead
	updatedDoc, err := canonSvc.LoadCanon(ctx, "test-campaign")
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}
	foundDead := false
	for _, e := range updatedDoc.Entities {
		if e.ID == "npc-villain" {
			if e.CanonState != domain.EntityStateDead {
				t.Fatalf("expected npc-villain to be dead, got %s", e.CanonState)
			}
			foundDead = true
		}
	}
	if !foundDead {
		t.Fatal("npc-villain not found in updated canon")
	}

	// Verify timeline event for quest completion
	if len(updatedDoc.Timeline) != 1 {
		t.Fatalf("expected 1 timeline event, got %d", len(updatedDoc.Timeline))
	}
}

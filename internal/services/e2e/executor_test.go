package e2e

import (
	"context"
	"testing"
)

func TestExecuteStep_CreateCampaign(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	step := TestStep{
		Action: "create_campaign",
		Params: map[string]any{
			"name":  "test-executor-campaign",
			"title": "Executor Test",
		},
	}

	if err := executeStep(ctx, harness, step); err != nil {
		t.Fatalf("executeStep(create_campaign) error = %v", err)
	}
}

func TestExecuteStep_UnsupportedAction(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	step := TestStep{
		Action: "nonexistent_action",
		Params: map[string]any{},
	}

	err := executeStep(ctx, harness, step)
	if err == nil {
		t.Fatal("executeStep(nonexistent_action) should return error")
	}
	if err.Error() != "unsupported action: nonexistent_action" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecuteStep_SaveLore(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	// Create campaign first
	if err := executeStep(ctx, harness, TestStep{
		Action: "create_campaign",
		Params: map[string]any{"name": "test-lore-campaign"},
	}); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	step := TestStep{
		Action: "save_lore",
		Params: map[string]any{
			"campaign": "test-lore-campaign",
			"content":  "# Lore\n\nAncient history...",
		},
	}

	if err := executeStep(ctx, harness, step); err != nil {
		t.Fatalf("executeStep(save_lore) error = %v", err)
	}
}

func TestExecuteStep_SaveNPCs(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	if err := executeStep(ctx, harness, TestStep{
		Action: "create_campaign",
		Params: map[string]any{"name": "test-npc-campaign"},
	}); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	step := TestStep{
		Action: "save_npcs",
		Params: map[string]any{
			"campaign": "test-npc-campaign",
			"content":  "# NPCs\n\n## Gandalf\nA wizard...",
		},
	}

	if err := executeStep(ctx, harness, step); err != nil {
		t.Fatalf("executeStep(save_npcs) error = %v", err)
	}
}

func TestExecuteStep_SaveEncounters(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	if err := executeStep(ctx, harness, TestStep{
		Action: "create_campaign",
		Params: map[string]any{"name": "test-enc-campaign"},
	}); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	step := TestStep{
		Action: "save_encounters",
		Params: map[string]any{
			"campaign": "test-enc-campaign",
			"content":  "# Encounters\n\n## Ambush\nBandits...",
		},
	}

	if err := executeStep(ctx, harness, step); err != nil {
		t.Fatalf("executeStep(save_encounters) error = %v", err)
	}
}

func TestExecuteStep_GenerateCharacter(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	if err := executeStep(ctx, harness, TestStep{
		Action: "create_campaign",
		Params: map[string]any{"name": "test-char-campaign"},
	}); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	step := TestStep{
		Action: "generate_character",
		Params: map[string]any{
			"campaign":   "test-char-campaign",
			"name":       "Aragorn",
			"race":       "humano",
			"class":      "guerrero",
			"level":      float64(5),
			"background": "soldado",
			"alignment":  "LG",
		},
	}

	if err := executeStep(ctx, harness, step); err != nil {
		t.Fatalf("executeStep(generate_character) error = %v", err)
	}
}

func TestExecuteStep_CreateQuest(t *testing.T) {
	harness := NewTestHarness(t)
	ctx := context.Background()

	if err := executeStep(ctx, harness, TestStep{
		Action: "create_campaign",
		Params: map[string]any{"name": "test-quest-campaign"},
	}); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	step := TestStep{
		Action: "create_quest",
		Params: map[string]any{
			"campaign":    "test-quest-campaign",
			"quest_title": "Find the Sword",
			"quest_type":  "main",
			"hook":        "A stranger approaches...",
			"stakes":      "The kingdom's fate",
			"reward":      "1000 gold",
		},
	}

	if err := executeStep(ctx, harness, step); err != nil {
		t.Fatalf("executeStep(create_quest) error = %v", err)
	}
}

func TestGetStringParam(t *testing.T) {
	params := map[string]any{
		"name":  "test-value",
		"empty": "",
	}

	if got := getStringParam(params, "name"); got != "test-value" {
		t.Errorf("getStringParam(name) = %q, want %q", got, "test-value")
	}
	if got := getStringParam(params, "missing"); got != "" {
		t.Errorf("getStringParam(missing) = %q, want empty", got)
	}
	if got := getStringParam(params, "empty"); got != "" {
		t.Errorf("getStringParam(empty) = %q, want empty", got)
	}
}

func TestGetIntParam(t *testing.T) {
	params := map[string]any{
		"int_val":    float64(42),
		"actual_int": 7,
		"string":     "not-a-number",
	}

	if got := getIntParam(params, "int_val"); got != 42 {
		t.Errorf("getIntParam(int_val) = %d, want 42", got)
	}
	if got := getIntParam(params, "actual_int"); got != 7 {
		t.Errorf("getIntParam(actual_int) = %d, want 7", got)
	}
	if got := getIntParam(params, "missing"); got != 0 {
		t.Errorf("getIntParam(missing) = %d, want 0", got)
	}
	if got := getIntParam(params, "string"); got != 0 {
		t.Errorf("getIntParam(string) = %d, want 0", got)
	}
}

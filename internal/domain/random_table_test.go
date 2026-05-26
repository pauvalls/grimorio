package domain

import (
	"encoding/json"
	"testing"
)

func TestIsValidTableType(t *testing.T) {
	tests := []struct {
		name string
		t    string
		want bool
	}{
		{"encounter", "encounter", true},
		{"rumor", "rumor", true},
		{"weather", "weather", true},
		{"treasure", "treasure", true},
		{"invalid", "combat", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTableType(tt.t)
			if got != tt.want {
				t.Fatalf("IsValidTableType(%q) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestTableContext_JSONSerialization(t *testing.T) {
	tests := []struct {
		name           string
		ctx            TableContext
		wantJSONFields []string // Fields that should appear in JSON
	}{
		{
			name: "with all fields",
			ctx: TableContext{
				LevelRange:   "1-5",
				SettingType:  "wilderness",
				PartySize:    4,
				LocationHint: "forest",
			},
			wantJSONFields: []string{"level_range", "setting_type", "party_size", "location_hint"},
		},
		{
			name: "empty context",
			ctx:  TableContext{},
			wantJSONFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.ctx)
			if err != nil {
				t.Fatalf("json.Marshal failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("json.Unmarshal failed: %v", err)
			}

			// Check expected fields are present
			for _, field := range tt.wantJSONFields {
				if _, ok := result[field]; !ok {
					t.Errorf("expected field %q in JSON, got: %v", field, result)
				}
			}

			// Check that faction_context and narrative_state are NOT in JSON (json:"-" tag)
			if _, ok := result["faction_context"]; ok {
				t.Errorf("faction_context should be excluded from JSON (json:\"-\" tag)")
			}
			if _, ok := result["narrative_state"]; ok {
				t.Errorf("narrative_state should be excluded from JSON (json:\"-\" tag)")
			}
		})
	}
}

func TestTableContext_WithFactionAndNarrative(t *testing.T) {
	// Test that TableContext can hold faction and narrative state references
	factionCtx := &FactionReputationMatrix{
		CampaignID: "test-campaign",
		Entries:    []ReputationEntry{},
	}
	narrativeState := &NarrativeState{
		SchemaVersion: "v2",
		CampaignID:    "test-campaign",
		DeadNPCs:      []NPCDeathRecord{},
	}

	ctx := TableContext{
		LocationHint:   "forest",
		FactionContext: factionCtx,
		NarrativeState: narrativeState,
	}

	if ctx.FactionContext == nil {
		t.Fatal("FactionContext should not be nil")
	}
	if ctx.FactionContext.CampaignID != "test-campaign" {
		t.Errorf("expected CampaignID 'test-campaign', got %q", ctx.FactionContext.CampaignID)
	}

	if ctx.NarrativeState == nil {
		t.Fatal("NarrativeState should not be nil")
	}
	if ctx.NarrativeState.CampaignID != "test-campaign" {
		t.Errorf("expected CampaignID 'test-campaign', got %q", ctx.NarrativeState.CampaignID)
	}
}

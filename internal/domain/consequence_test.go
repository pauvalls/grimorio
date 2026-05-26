package domain

import (
	"encoding/json"
	"testing"
)

func TestDelayedEffectJSON(t *testing.T) {
	de := DelayedEffect{
		ID:             "rule-001-0",
		Description:    "Village burns down",
		EffectType:     "spawn",
		Target:         "Village",
		Effect:         Effect{Type: "spawn", Target: "Village", Description: "Village burns down"},
		TriggerSession: 2,
		ApplySession:   4,
		Applied:        false,
	}

	data, err := json.Marshal(de)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded DelayedEffect
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ID != "rule-001-0" {
		t.Fatalf("id = %q, want rule-001-0", decoded.ID)
	}
	if decoded.Description != "Village burns down" {
		t.Fatalf("description = %q, want 'Village burns down'", decoded.Description)
	}
	if decoded.EffectType != "spawn" {
		t.Fatalf("effect_type = %q, want spawn", decoded.EffectType)
	}
	if decoded.Target != "Village" {
		t.Fatalf("target = %q, want Village", decoded.Target)
	}
	if decoded.TriggerSession != 2 {
		t.Fatalf("trigger_session = %d, want 2", decoded.TriggerSession)
	}
	if decoded.ApplySession != 4 {
		t.Fatalf("apply_session = %d, want 4", decoded.ApplySession)
	}
	if decoded.Applied != false {
		t.Fatalf("applied = %v, want false", decoded.Applied)
	}
}

func TestDelayedEffect_BackwardCompat(t *testing.T) {
	// Old JSON without new fields should unmarshal with zero values
	oldJSON := `{"effect":{"type":"spawn","target":"npc"},"trigger_session":2,"apply_session":4}`
	var de DelayedEffect
	if err := json.Unmarshal([]byte(oldJSON), &de); err != nil {
		t.Fatalf("unmarshal old json failed: %v", err)
	}
	if de.ID != "" {
		t.Fatalf("id should be empty for old json, got %q", de.ID)
	}
	if de.ApplySession != 4 {
		t.Fatalf("apply_session = %d, want 4", de.ApplySession)
	}
}

func TestConsequenceRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    ConsequenceRule
		wantErr bool
		errField string
	}{
		{
			name: "valid rule",
			rule: ConsequenceRule{
				ID:       "rule-001",
				Name:     "NPC Death Replacement",
				Priority: 5,
				Effects:  []Effect{{Type: "spawn", Target: "npc"}},
			},
			wantErr: false,
		},
		{
			name: "missing id",
			rule: ConsequenceRule{
				Name:     "NPC Death Replacement",
				Priority: 5,
				Effects:  []Effect{{Type: "spawn", Target: "npc"}},
			},
			wantErr:  true,
			errField: "id",
		},
		{
			name: "missing name",
			rule: ConsequenceRule{
				ID:       "rule-001",
				Priority: 5,
				Effects:  []Effect{{Type: "spawn", Target: "npc"}},
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "invalid priority zero",
			rule: ConsequenceRule{
				ID:       "rule-001",
				Name:     "NPC Death Replacement",
				Priority: 0,
				Effects:  []Effect{{Type: "spawn", Target: "npc"}},
			},
			wantErr:  true,
			errField: "priority",
		},
		{
			name: "no effects",
			rule: ConsequenceRule{
				ID:       "rule-001",
				Name:     "NPC Death Replacement",
				Priority: 5,
			},
			wantErr:  true,
			errField: "effects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if ve, ok := err.(*ValidationError); ok && ve.Field != tt.errField {
					t.Fatalf("expected error on field %q, got %q", tt.errField, ve.Field)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

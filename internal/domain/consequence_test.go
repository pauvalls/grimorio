package domain

import "testing"

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

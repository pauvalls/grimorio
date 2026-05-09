package domain

import (
	"strings"
	"testing"
)

func TestTactics_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tactics *Tactics
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tactics",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				EncounterID:      "encounter_1",
				IntelligenceTier: TierTactical,
				OpeningMove:      "Cast fireball on clustered enemies",
				TargetPriority: []TargetPriority{
					{Priority: 1, TargetType: "healer", Reasoning: "Prevent healing"},
					{Priority: 2, TargetType: "squishy", Reasoning: "Eliminate damage dealers"},
				},
				RetreatConditions: []RetreatCondition{
					{Trigger: "HP < 25%", Method: "Disengage and flee"},
				},
				AbilityUsage: []AbilityTactic{
					{AbilityName: "Fireball", UsageCondition: "When 3+ enemies clustered"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing monster_id",
			tactics: &Tactics{
				EncounterID:      "encounter_1",
				IntelligenceTier: TierTactical,
				OpeningMove:      "Attack",
				TargetPriority:   []TargetPriority{{Priority: 1, TargetType: "nearest"}},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "monster_id is required",
		},
		{
			name: "missing encounter_id",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				IntelligenceTier: TierTactical,
				OpeningMove:      "Attack",
				TargetPriority:   []TargetPriority{{Priority: 1, TargetType: "nearest"}},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "encounter_id is required",
		},
		{
			name: "missing intelligence_tier",
			tactics: &Tactics{
				MonsterID:   "monster_1",
				EncounterID: "encounter_1",
				OpeningMove: "Attack",
				TargetPriority:   []TargetPriority{{Priority: 1, TargetType: "nearest"}},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "intelligence_tier is required",
		},
		{
			name: "invalid intelligence_tier",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				EncounterID:      "encounter_1",
				IntelligenceTier: "super_smart",
				OpeningMove:      "Attack",
				TargetPriority:   []TargetPriority{{Priority: 1, TargetType: "nearest"}},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "invalid intelligence tier",
		},
		{
			name: "less than 2 target priorities",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				EncounterID:      "encounter_1",
				IntelligenceTier: TierTactical,
				OpeningMove:      "Attack",
				TargetPriority:   []TargetPriority{{Priority: 1, TargetType: "nearest"}},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "at least 2 target priorities",
		},
		{
			name: "no retreat conditions",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				EncounterID:      "encounter_1",
				IntelligenceTier: TierTactical,
				OpeningMove:      "Attack",
				TargetPriority: []TargetPriority{
					{Priority: 1, TargetType: "nearest"},
					{Priority: 2, TargetType: "healer"},
				},
			},
			wantErr: true,
			errMsg:  "at least 1 retreat condition",
		},
		{
			name: "missing opening move",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				EncounterID:      "encounter_1",
				IntelligenceTier: TierTactical,
				TargetPriority: []TargetPriority{
					{Priority: 1, TargetType: "nearest"},
					{Priority: 2, TargetType: "healer"},
				},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "opening_move is required",
		},
		{
			name: "non-sequential target priorities",
			tactics: &Tactics{
				MonsterID:        "monster_1",
				EncounterID:      "encounter_1",
				IntelligenceTier: TierTactical,
				OpeningMove:      "Attack",
				TargetPriority: []TargetPriority{
					{Priority: 1, TargetType: "nearest"},
					{Priority: 3, TargetType: "healer"}, // Skips 2
				},
				RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
			},
			wantErr: true,
			errMsg:  "expected priority 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tactics.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGetIntelligenceTierFromScore(t *testing.T) {
	tests := []struct {
		name     string
		intScore int
		want     IntelligenceTier
	}{
		{"INT 1", 1, TierInstinctive},
		{"INT 3", 3, TierInstinctive},
		{"INT 4", 4, TierInstinctive},
		{"INT 5", 5, TierSimple},
		{"INT 7", 7, TierSimple},
		{"INT 9", 9, TierSimple},
		{"INT 10", 10, TierTactical},
		{"INT 12", 12, TierTactical},
		{"INT 14", 14, TierTactical},
		{"INT 15", 15, TierStrategic},
		{"INT 18", 18, TierStrategic},
		{"INT 20", 20, TierStrategic},
		{"INT 0", 0, TierInstinctive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIntelligenceTierFromScore(tt.intScore)
			if got != tt.want {
				t.Errorf("GetIntelligenceTierFromScore(%d) = %s, want %s", tt.intScore, got, tt.want)
			}
		})
	}
}

func TestGetTacticalComplexity(t *testing.T) {
	tests := []struct {
		tier IntelligenceTier
		want string
	}{
		{TierInstinctive, "instinct"},
		{TierSimple, "Simple"},
		{TierTactical, "Coordinated"},
		{TierStrategic, "Advanced"},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			got := GetTacticalComplexity(tt.tier)
			if !strings.Contains(got, tt.want) {
				t.Errorf("GetTacticalComplexity(%s) = %q, expected to contain %q", tt.tier, got, tt.want)
			}
		})
	}
}

func TestTactics_HasPackTactics(t *testing.T) {
	tacticsWithPack := &Tactics{
		MonsterID:        "wolf",
		EncounterID:      "enc_1",
		IntelligenceTier: TierSimple,
		OpeningMove:      "Surround prey",
		TargetPriority: []TargetPriority{
			{Priority: 1, TargetType: "nearest"},
			{Priority: 2, TargetType: "isolated"},
		},
		RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Flee"}},
		PackBehavior: &PackTactic{
			Type:        "pack_tactics",
			Description: "Advantage when adjacent to ally",
		},
	}

	tacticsWithoutPack := &Tactics{
		MonsterID:        "dragon",
		EncounterID:      "enc_2",
		IntelligenceTier: TierStrategic,
		OpeningMove:      "Fly and breathe fire",
		TargetPriority: []TargetPriority{
			{Priority: 1, TargetType: "healer"},
			{Priority: 2, TargetType: "squishy"},
		},
		RetreatConditions: []RetreatCondition{{Trigger: "HP < 25%", Method: "Fly away"}},
	}

	if !tacticsWithPack.HasPackTactics() {
		t.Error("HasPackTactics() = false, want true")
	}
	if tacticsWithoutPack.HasPackTactics() {
		t.Error("HasPackTactics() = true, want false")
	}
}

func TestTactics_GetPrimaryTarget(t *testing.T) {
	tactics := &Tactics{
		TargetPriority: []TargetPriority{
			{Priority: 1, TargetType: "healer", Reasoning: "Priority target"},
			{Priority: 2, TargetType: "squishy", Reasoning: "Secondary"},
		},
	}

	primary := tactics.GetPrimaryTarget()
	if primary != "healer" {
		t.Errorf("GetPrimaryTarget() = %s, want healer", primary)
	}
}

func TestTactics_ShouldRetreat(t *testing.T) {
	tactics := &Tactics{
		RetreatConditions: []RetreatCondition{
			{Trigger: "HP < 25%", Method: "Disengage and flee"},
			{Trigger: "HP < 50%", Method: "Fight defensively"},
		},
	}

	tests := []struct {
		name        string
		hpPercent   int
		wantRetreat bool
		wantMethod  string
	}{
		{"HP 10%", 10, true, "Disengage and flee"},
		{"HP 25%", 25, true, "Disengage and flee"},
		{"HP 40%", 40, true, "Fight defensively"},
		{"HP 50%", 50, true, "Fight defensively"},
		{"HP 75%", 75, false, ""},
		{"HP 100%", 100, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retreat, method := tactics.ShouldRetreat(tt.hpPercent)
			if retreat != tt.wantRetreat {
				t.Errorf("ShouldRetreat(%d) retreat = %v, want %v", tt.hpPercent, retreat, tt.wantRetreat)
			}
			if method != tt.wantMethod {
				t.Errorf("ShouldRetreat(%d) method = %q, want %q", tt.hpPercent, method, tt.wantMethod)
			}
		})
	}
}

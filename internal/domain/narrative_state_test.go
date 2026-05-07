package domain

import (
	"testing"
	"time"
)

func TestNarrativeState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		state   NarrativeState
		wantErr bool
	}{
		{
			name: "valid narrative state",
			state: NarrativeState{
				SchemaVersion:  SchemaVersionV2,
				CampaignID:     "test-campaign",
				CurrentSession: 1,
				LastUpdated:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing schema version",
			state: NarrativeState{
				CampaignID:     "test-campaign",
				CurrentSession: 1,
			},
			wantErr: true,
		},
		{
			name: "unsupported schema version",
			state: NarrativeState{
				SchemaVersion:  "1.0",
				CampaignID:     "test-campaign",
				CurrentSession: 1,
			},
			wantErr: true,
		},
		{
			name: "missing campaign ID",
			state: NarrativeState{
				SchemaVersion:  SchemaVersionV2,
				CurrentSession: 1,
			},
			wantErr: true,
		},
		{
			name: "negative session",
			state: NarrativeState{
				SchemaVersion:  SchemaVersionV2,
				CampaignID:     "test-campaign",
				CurrentSession: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRevealedClue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		clue    RevealedClue
		wantErr bool
	}{
		{
			name: "valid clue",
			clue: RevealedClue{
				ID:              "clue-001",
				Description:     "The diary mentions the password",
				SourceAct:       "act_1",
				SessionRevealed: 1,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			clue: RevealedClue{
				Description:     "The diary mentions the password",
				SourceAct:       "act_1",
				SessionRevealed: 1,
			},
			wantErr: true,
		},
		{
			name: "missing description",
			clue: RevealedClue{
				ID:              "clue-001",
				SourceAct:       "act_1",
				SessionRevealed: 1,
			},
			wantErr: true,
		},
		{
			name: "missing source act",
			clue: RevealedClue{
				ID:              "clue-001",
				Description:     "The diary mentions the password",
				SessionRevealed: 1,
			},
			wantErr: true,
		},
		{
			name: "negative session",
			clue: RevealedClue{
				ID:              "clue-001",
				Description:     "The diary mentions the password",
				SourceAct:       "act_1",
				SessionRevealed: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.clue.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestQuestState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		quest   QuestState
		wantErr bool
	}{
		{
			name: "valid active quest",
			quest: QuestState{
				ID:     "quest-001",
				Name:   "Find the stone",
				Status: "active",
			},
			wantErr: false,
		},
		{
			name: "valid completed quest",
			quest: QuestState{
				ID:     "quest-001",
				Name:   "Find the stone",
				Status: "completed",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			quest: QuestState{
				Name:   "Find the stone",
				Status: "active",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			quest: QuestState{
				ID:     "quest-001",
				Status: "active",
			},
			wantErr: true,
		},
		{
			name: "missing status",
			quest: QuestState{
				ID:   "quest-001",
				Name: "Find the stone",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			quest: QuestState{
				ID:     "quest-001",
				Name:   "Find the stone",
				Status: "pending",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.quest.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNPCDeathRecord_Validate(t *testing.T) {
	tests := []struct {
		name    string
		death   NPCDeathRecord
		wantErr bool
	}{
		{
			name: "valid death record",
			death: NPCDeathRecord{
				NPCID:   "npc-001",
				Name:    "El Informador",
				Session: 2,
				Cause:   "combat",
				KilledBy: "villain",
			},
			wantErr: false,
		},
		{
			name: "missing NPC ID",
			death: NPCDeathRecord{
				Name:    "El Informador",
				Session: 2,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			death: NPCDeathRecord{
				NPCID:   "npc-001",
				Session: 2,
			},
			wantErr: true,
		},
		{
			name: "negative session",
			death: NPCDeathRecord{
				NPCID:   "npc-001",
				Name:    "El Informador",
				Session: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.death.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestKeyItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    KeyItem
		wantErr bool
	}{
		{
			name: "valid key item",
			item: KeyItem{
				ID:           "item-001",
				Name:         "Stone of Golorr",
				Holder:       "party",
				SessionFound: 2,
				IsMcGuffin:   true,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			item: KeyItem{
				Name:         "Stone of Golorr",
				Holder:       "party",
				SessionFound: 2,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			item: KeyItem{
				ID:           "item-001",
				Holder:       "party",
				SessionFound: 2,
			},
			wantErr: true,
		},
		{
			name: "missing holder",
			item: KeyItem{
				ID:           "item-001",
				Name:         "Stone of Golorr",
				SessionFound: 2,
			},
			wantErr: true,
		},
		{
			name: "negative session found",
			item: KeyItem{
				ID:           "item-001",
				Name:         "Stone of Golorr",
				Holder:       "party",
				SessionFound: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSessionRecord_Validate(t *testing.T) {
	tests := []struct {
		name    string
		record  SessionRecord
		wantErr bool
	}{
		{
			name: "valid session record",
			record: SessionRecord{
				SessionNum: 1,
				Date:       time.Now(),
				Summary:    "The party entered the dungeon",
			},
			wantErr: false,
		},
		{
			name: "negative session number",
			record: SessionRecord{
				SessionNum: -1,
				Summary:    "The party entered the dungeon",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDecision_Validate(t *testing.T) {
	tests := []struct {
		name     string
		decision Decision
		wantErr  bool
	}{
		{
			name: "valid decision",
			decision: Decision{
				ID:          "dec-001",
				Description: "Should we trust the informant?",
				ChoiceMade:  "Yes",
				ImpactScope: "local",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			decision: Decision{
				Description: "Should we trust the informant?",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			decision: Decision{
				ID: "dec-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.decision.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDMOverride_Validate(t *testing.T) {
	tests := []struct {
		name     string
		override DMOverride
		wantErr  bool
	}{
		{
			name: "valid override",
			override: DMOverride{
				ID:         "ovr-001",
				TargetType: "entity",
				TargetID:   "npc-001",
				Field:      "canon_state",
				NewValue:   "dead",
				Reason:     "Player killed him",
				SessionNum: 2,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			override: DMOverride{
				TargetID: "npc-001",
				Field:    "canon_state",
			},
			wantErr: true,
		},
		{
			name: "missing target ID",
			override: DMOverride{
				ID:      "ovr-001",
				Field:   "canon_state",
				NewValue: "dead",
			},
			wantErr: true,
		},
		{
			name: "missing field",
			override: DMOverride{
				ID:       "ovr-001",
				TargetID: "npc-001",
				NewValue: "dead",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.override.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

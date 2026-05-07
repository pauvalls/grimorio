package domain

import (
	"testing"
	"time"
)

func TestCanonDocument_Validate(t *testing.T) {
	tests := []struct {
		name    string
		doc     CanonDocument
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid canon document",
			doc: CanonDocument{
				SchemaVersion: SchemaVersionV2,
				CampaignID:    "test-campaign",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing schema version",
			doc: CanonDocument{
				CampaignID: "test-campaign",
			},
			wantErr: true,
			errMsg:  "schema version is required",
		},
		{
			name: "unsupported schema version",
			doc: CanonDocument{
				SchemaVersion: "1.0",
				CampaignID:    "test-campaign",
			},
			wantErr: true,
			errMsg:  "unsupported schema version",
		},
		{
			name: "missing campaign ID",
			doc: CanonDocument{
				SchemaVersion: SchemaVersionV2,
			},
			wantErr: true,
			errMsg:  "campaign ID is required",
		},
		{
			name: "invalid campaign ID format",
			doc: CanonDocument{
				SchemaVersion: SchemaVersionV2,
				CampaignID:    "Test Campaign",
			},
			wantErr: true,
			errMsg:  "kebab-case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg && !containsStr(err.Error(), tt.errMsg) {
					t.Fatalf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCanonFact_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fact    CanonFact
		wantErr bool
	}{
		{
			name: "valid fact",
			fact: CanonFact{
				ID:        "fact-001",
				Category:  "lore",
				Statement: "The ancient curse comes from the god Morbus",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			fact: CanonFact{
				Category:  "lore",
				Statement: "The ancient curse comes from the god Morbus",
			},
			wantErr: true,
		},
		{
			name: "missing statement",
			fact: CanonFact{
				ID:       "fact-001",
				Category: "lore",
			},
			wantErr: true,
		},
		{
			name: "short statement",
			fact: CanonFact{
				ID:        "fact-001",
				Category:  "lore",
				Statement: "Short",
			},
			wantErr: true,
		},
		{
			name: "missing category",
			fact: CanonFact{
				ID:        "fact-001",
				Statement: "The ancient curse comes from the god Morbus",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fact.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCanonEntity_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entity  CanonEntity
		wantErr bool
	}{
		{
			name: "valid NPC entity",
			entity: CanonEntity{
				ID:         "npc-001",
				Name:       "Lord Vex",
				Type:       EntityTypeNPC,
				CanonState: EntityStateAlive,
			},
			wantErr: false,
		},
		{
			name: "valid location entity",
			entity: CanonEntity{
				ID:         "loc-001",
				Name:       "Thornvale Keep",
				Type:       EntityTypeLocation,
				CanonState: EntityStateAlive,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			entity: CanonEntity{
				Name:       "Lord Vex",
				Type:       EntityTypeNPC,
				CanonState: EntityStateAlive,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			entity: CanonEntity{
				ID:         "npc-001",
				Type:       EntityTypeNPC,
				CanonState: EntityStateAlive,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			entity: CanonEntity{
				ID:         "npc-001",
				Name:       "Lord Vex",
				Type:       "invalid",
				CanonState: EntityStateAlive,
			},
			wantErr: true,
		},
		{
			name: "invalid state",
			entity: CanonEntity{
				ID:         "npc-001",
				Name:       "Lord Vex",
				Type:       EntityTypeNPC,
				CanonState: "ghost",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCanonRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rule    CanonRule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: CanonRule{
				ID:        "rule-001",
				Domain:    "magic",
				Statement: "Arcane magic is banned in the city",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			rule: CanonRule{
				Domain:    "magic",
				Statement: "Arcane magic is banned",
			},
			wantErr: true,
		},
		{
			name: "missing statement",
			rule: CanonRule{
				ID:     "rule-001",
				Domain: "magic",
			},
			wantErr: true,
		},
		{
			name: "missing domain",
			rule: CanonRule{
				ID:        "rule-001",
				Statement: "Arcane magic is banned",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCanonRelationship_Validate(t *testing.T) {
	tests := []struct {
		name         string
		relationship CanonRelationship
		wantErr      bool
	}{
		{
			name: "valid relationship",
			relationship: CanonRelationship{
				ID:       "rel-001",
				From:     "npc-001",
				To:       "npc-002",
				Type:     RelationshipTypeAlly,
				Strength: 5,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			relationship: CanonRelationship{
				From:     "npc-001",
				To:       "npc-002",
				Type:     RelationshipTypeAlly,
				Strength: 5,
			},
			wantErr: true,
		},
		{
			name: "missing from",
			relationship: CanonRelationship{
				ID:       "rel-001",
				To:       "npc-002",
				Type:     RelationshipTypeAlly,
				Strength: 5,
			},
			wantErr: true,
		},
		{
			name: "missing to",
			relationship: CanonRelationship{
				ID:       "rel-001",
				From:     "npc-001",
				Type:     RelationshipTypeAlly,
				Strength: 5,
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			relationship: CanonRelationship{
				ID:       "rel-001",
				From:     "npc-001",
				To:       "npc-002",
				Type:     "friend",
				Strength: 5,
			},
			wantErr: true,
		},
		{
			name: "strength too high",
			relationship: CanonRelationship{
				ID:       "rel-001",
				From:     "npc-001",
				To:       "npc-002",
				Type:     RelationshipTypeAlly,
				Strength: 15,
			},
			wantErr: true,
		},
		{
			name: "strength too low",
			relationship: CanonRelationship{
				ID:       "rel-001",
				From:     "npc-001",
				To:       "npc-002",
				Type:     RelationshipTypeEnemy,
				Strength: -15,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.relationship.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCanonTimelineEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   CanonTimelineEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: CanonTimelineEvent{
				ID:          "evt-001",
				Timestamp:   "6 months ago",
				Description: "The curse fell upon Thornvale",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			event: CanonTimelineEvent{
				Description: "The curse fell upon Thornvale",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			event: CanonTimelineEvent{
				ID: "evt-001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

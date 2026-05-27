package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestBuildDeadNPCMap(t *testing.T) {
	tests := []struct {
		name         string
		narrativeState *domain.NarrativeState
		wantMapSize  int
		wantContains []string // Keys that should be in the map
	}{
		{
			name:         "nil narrative state",
			narrativeState: nil,
			wantMapSize:  0,
			wantContains: nil,
		},
		{
			name: "single dead NPC",
			narrativeState: &domain.NarrativeState{
				DeadNPCs: []domain.NPCDeathRecord{
					{NPCID: "npc_gareth", Name: "Gareth"},
				},
			},
			wantMapSize:  2, // Both ID and lowercase name
			wantContains: []string{"npc_gareth", "gareth"},
		},
		{
			name: "multiple dead NPCs",
			narrativeState: &domain.NarrativeState{
				DeadNPCs: []domain.NPCDeathRecord{
					{NPCID: "npc_gareth", Name: "Gareth"},
					{NPCID: "npc_maria", Name: "Maria"},
				},
			},
			wantMapSize:  4,
			wantContains: []string{"npc_gareth", "gareth", "npc_maria", "maria"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDeadNPCMap(tt.narrativeState)
			if tt.narrativeState == nil {
				if got != nil {
					t.Errorf("buildDeadNPCMap(nil) should return nil, got %v", got)
				}
				return
			}
			if len(got) != tt.wantMapSize {
				t.Errorf("buildDeadNPCMap() map size = %d, want %d", len(got), tt.wantMapSize)
			}
			for _, key := range tt.wantContains {
				if !got[key] {
					t.Errorf("buildDeadNPCMap() should contain key %q", key)
				}
			}
		})
	}
}

func TestBuildRevealedClueMap(t *testing.T) {
	tests := []struct {
		name         string
		narrativeState *domain.NarrativeState
		wantMapSize  int
	}{
		{
			name:         "nil narrative state",
			narrativeState: nil,
			wantMapSize:  0,
		},
		{
			name: "single revealed clue",
			narrativeState: &domain.NarrativeState{
				RevealedClues: []domain.RevealedClue{
					{ID: "clue_1", Description: "ancient ruin treasure"},
				},
			},
			wantMapSize: 2, // ID and lowercase description
		},
		{
			name: "multiple revealed clues",
			narrativeState: &domain.NarrativeState{
				RevealedClues: []domain.RevealedClue{
					{ID: "clue_1", Description: "ancient ruin treasure"},
					{ID: "clue_2", Description: "secret tunnel"},
				},
			},
			wantMapSize: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRevealedClueMap(tt.narrativeState)
			if tt.narrativeState == nil {
				if got != nil {
					t.Errorf("buildRevealedClueMap(nil) should return nil, got %v", got)
				}
				return
			}
			if len(got) != tt.wantMapSize {
				t.Errorf("buildRevealedClueMap() map size = %d, want %d", len(got), tt.wantMapSize)
			}
		})
	}
}

func TestShouldExcludeFact(t *testing.T) {
	tests := []struct {
		name      string
		fact      domain.CanonFact
		deadNPCMap map[string]bool
		tableType domain.TableType
		wantExclude bool
	}{
		{
			name: "dead NPC in encounter table excluded",
			fact: domain.CanonFact{
				Statement: "Gareth the Guard patrols the gate",
			},
			deadNPCMap: map[string]bool{"gareth": true, "npc_gareth": true},
			tableType:  domain.TableTypeEncounter,
			wantExclude: true,
		},
		{
			name: "dead NPC in rumor table excluded",
			fact: domain.CanonFact{
				Statement: "Rumors about Gareth spread",
			},
			deadNPCMap: map[string]bool{"gareth": true},
			tableType:  domain.TableTypeRumor,
			wantExclude: true,
		},
		{
			name: "dead NPC in treasure table not excluded if indirect",
			fact: domain.CanonFact{
				Statement: "Guard's equipment found",
			},
			deadNPCMap: map[string]bool{"gareth": true},
			tableType:  domain.TableTypeTreasure,
			wantExclude: false,
		},
		{
			name: "no dead NPC reference not excluded",
			fact: domain.CanonFact{
				Statement: "Wolf attacks",
			},
			deadNPCMap: map[string]bool{"gareth": true},
			tableType:  domain.TableTypeEncounter,
			wantExclude: false,
		},
		{
			name: "nil dead NPC map never excludes",
			fact: domain.CanonFact{
				Statement: "Gareth patrols",
			},
			deadNPCMap: nil,
			tableType:  domain.TableTypeEncounter,
			wantExclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldExcludeFact(tt.fact, tt.deadNPCMap, tt.tableType)
			if got != tt.wantExclude {
				t.Errorf("shouldExcludeFact() = %v, want %v", got, tt.wantExclude)
			}
		})
	}
}

func TestFactReferencesDeadNPC(t *testing.T) {
	tests := []struct {
		name       string
		fact       domain.CanonFact
		deadNPCMap map[string]bool
		wantRef    bool
	}{
		{
			name: "NPC ID reference",
			fact: domain.CanonFact{
				Statement: "npc_gareth was seen",
			},
			deadNPCMap: map[string]bool{"npc_gareth": true},
			wantRef:    true,
		},
		{
			name: "NPC name reference",
			fact: domain.CanonFact{
				Statement: "Gareth patrols the gate",
			},
			deadNPCMap: map[string]bool{"gareth": true},
			wantRef:    true,
		},
		{
			name: "no reference",
			fact: domain.CanonFact{
				Statement: "Wolf attacks",
			},
			deadNPCMap: map[string]bool{"gareth": true},
			wantRef:    false,
		},
		{
			name: "case insensitive",
			fact: domain.CanonFact{
				Statement: "GARETH the Guard",
			},
			deadNPCMap: map[string]bool{"gareth": true},
			wantRef:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := factReferencesDeadNPC(tt.fact, tt.deadNPCMap)
			if got != tt.wantRef {
				t.Errorf("factReferencesDeadNPC() = %v, want %v", got, tt.wantRef)
			}
		})
	}
}

func TestIsDirectNPCReference(t *testing.T) {
	tests := []struct {
		name      string
		fact      domain.CanonFact
		wantDirect bool
	}{
		{
			name: "direct with 'the'",
			fact: domain.CanonFact{
				Statement: "Gareth the Guard patrols",
			},
			wantDirect: true,
		},
		{
			name: "direct with title",
			fact: domain.CanonFact{
				Statement: "the captain Maria leads",
			},
			wantDirect: true,
		},
		{
			name: "indirect reference",
			fact: domain.CanonFact{
				Statement: "guard's equipment found",
			},
			wantDirect: false,
		},
		{
			name: "no reference",
			fact: domain.CanonFact{
				Statement: "Wolf attacks",
			},
			wantDirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDirectNPCReference(tt.fact)
			if got != tt.wantDirect {
				t.Errorf("isDirectNPCReference() = %v, want %v", got, tt.wantDirect)
			}
		})
	}
}

func TestApplyNarrativeWeightModifier(t *testing.T) {
	tests := []struct {
		name           string
		baseWeight     int
		fact           domain.CanonFact
		revealedClueMap map[string]bool
		wantWeight     int
	}{
		{
			name:       "revealed clue boosts weight",
			baseWeight: 5,
			fact: domain.CanonFact{
				Statement: "Treasure hidden in ancient ruin",
			},
			revealedClueMap: map[string]bool{"ancient ruin": true},
			wantWeight:     8, // 5 + 3
		},
		{
			name:       "no matching clue no boost",
			baseWeight: 5,
			fact: domain.CanonFact{
				Statement: "Wolf attacks",
			},
			revealedClueMap: map[string]bool{"ancient ruin": true},
			wantWeight:     5,
		},
		{
			name:       "nil clue map no boost",
			baseWeight: 5,
			fact: domain.CanonFact{
				Statement: "Any statement",
			},
			revealedClueMap: nil,
			wantWeight:     5,
		},
		{
			name:       "case insensitive match",
			baseWeight: 5,
			fact: domain.CanonFact{
				Statement: "TREASURE in ANCIENT RUIN",
			},
			revealedClueMap: map[string]bool{"ancient ruin": true},
			wantWeight:     8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyNarrativeWeightModifier(tt.baseWeight, tt.fact, tt.revealedClueMap)
			if got != tt.wantWeight {
				t.Errorf("applyNarrativeWeightModifier(%d, fact, map) = %d, want %d",
					tt.baseWeight, got, tt.wantWeight)
			}
		})
	}
}

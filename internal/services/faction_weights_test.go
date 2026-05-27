package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestGetReputationStatus(t *testing.T) {
	tests := []struct {
		name   string
		score  int8
		wantStatus string
	}{
		{"hostile -100", -100, "hostile"},
		{"hostile -50", -50, "hostile"},
		{"hostile -30", -30, "hostile"},
		{"unfriendly -29", -29, "unfriendly"},
		{"unfriendly -10", -10, "unfriendly"},
		{"neutral -9", -9, "neutral"},
		{"neutral 0", 0, "neutral"},
		{"neutral 29", 29, "neutral"},
		{"friendly 30", 30, "friendly"},
		{"friendly 79", 79, "friendly"},
		{"allied 80", 80, "allied"},
		{"allied 100", 100, "allied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getReputationStatus(tt.score)
			if got != tt.wantStatus {
				t.Errorf("getReputationStatus(%d) = %q, want %q", tt.score, got, tt.wantStatus)
			}
		})
	}
}

func TestIsHostileFact(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		wantHostile bool
	}{
		{"ambush keyword", "Ambush by goblins", true},
		{"attack keyword", "Dragon attack on village", true},
		{"enemy keyword", "Enemy patrol spotted", true},
		{"hostile keyword", "Hostile creatures ahead", true},
		{"threat keyword", "Threat from the shadows", true},
		{"danger keyword", "Danger in the dungeon", true},
		{"assault keyword", "Assault on the castle", true},
		{"raid keyword", "Orc raid party", true},
		{"neutral statement", "Merchant offers goods", false},
		{"helpful statement", "Ally offers assistance", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fact := domain.CanonFact{Statement: tt.statement}
			got := isHostileFact(fact)
			if got != tt.wantHostile {
				t.Errorf("isHostileFact(%q) = %v, want %v", tt.statement, got, tt.wantHostile)
			}
		})
	}
}

func TestIsHelpfulFact(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		wantHelpful bool
	}{
		{"help keyword", "NPC offers help", true},
		{"ally keyword", "Ally provides support", true},
		{"friend keyword", "Friend shares information", true},
		{"offer keyword", "Merchant offers discount", true},
		{"assist keyword", "Guard assists the party", true},
		{"support keyword", "Support from the guild", true},
		{"aid keyword", "Aid arrives just in time", true},
		{"discount keyword", "Shopkeeper gives discount", true},
		{"information keyword", "Informant shares information", true},
		{"hostile statement", "Enemy attacks", false},
		{"neutral statement", "The weather is clear", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fact := domain.CanonFact{Statement: tt.statement}
			got := isHelpfulFact(fact)
			if got != tt.wantHelpful {
				t.Errorf("isHelpfulFact(%q) = %v, want %v", tt.statement, got, tt.wantHelpful)
			}
		})
	}
}

func TestApplyFactionWeightModifier(t *testing.T) {
	tests := []struct {
		name         string
		baseWeight   int
		fact         domain.CanonFact
		factionScore int8
		wantWeight   int
	}{
		{
			name:       "hostile faction boosts hostile encounter",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Thieves Guild ambush attack"},
			factionScore: -50, // Hostile
			wantWeight: 7,     // 5 * 1.5 = 7.5 -> 7
		},
		{
			name:       "hostile faction reduces helpful rumor",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Thieves Guild member offers help with discount"},
			factionScore: -50, // Hostile
			wantWeight: 1,     // 5 * 0.2 = 1 (clamped)
		},
		{
			name:       "friendly faction reduces hostile",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Enemy ambush attack near merchant guild"},
			factionScore: 50, // Friendly
			wantWeight: 2,    // 5 * 0.5 = 2.5 -> 2
		},
		{
			name:       "friendly faction boosts helpful",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Merchant Guild ally offers assistance and aid"},
			factionScore: 50, // Friendly
			wantWeight: 7,    // 5 * 1.4 = 7
		},
		{
			name:       "allied faction heavily reduces hostile",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Thieves Guild hostile attack"},
			factionScore: 90, // Allied
			wantWeight: 1,    // 5 * 0.2 = 1 (clamped)
		},
		{
			name:       "allied faction heavily boosts helpful",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Merchant Guild ally provides aid and help"},
			factionScore: 90, // Allied
			wantWeight: 8,    // 5 * 1.6 = 8
		},
		{
			name:       "neutral faction no modifier",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Thieves Guild patrol"},
			factionScore: 0, // Neutral
			wantWeight: 5,   // No change
		},
		{
			name:       "no faction association returns base",
			baseWeight: 5,
			fact:       domain.CanonFact{Statement: "Random wolf encounter"},
			factionScore: 0,
			wantWeight: 5,
		},
		{
			name:       "minimum weight clamped to 1",
			baseWeight: 1,
			fact:       domain.CanonFact{Statement: "Thieves Guild offers help and aid"},
			factionScore: -50, // Hostile to thieves guild, helpful fact
			wantWeight: 1,     // Should be clamped to minimum 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factionCtx := &domain.FactionReputationMatrix{
				CampaignID: "test",
				Entries: []domain.ReputationEntry{
					{
						FactionID: "faction_thieves_guild",
						PartyID:   "party_1",
						Score:     tt.factionScore,
						Status:    getReputationStatus(tt.factionScore),
					},
					{
						FactionID: "faction_merchant_guild",
						PartyID:   "party_1",
						Score:     tt.factionScore,
						Status:    getReputationStatus(tt.factionScore),
					},
				},
			}
			got := applyFactionWeightModifier(tt.baseWeight, tt.fact, factionCtx)
			if got != tt.wantWeight {
				t.Errorf("applyFactionWeightModifier(%d, fact, score=%d) = %d, want %d",
					tt.baseWeight, tt.factionScore, got, tt.wantWeight)
			}
		})
	}
}

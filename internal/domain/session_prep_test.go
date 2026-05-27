package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSessionPrep_JSON(t *testing.T) {
	prep := SessionPrep{
		CampaignID:      "test-campaign",
		SessionNum:      4,
		PreviouslyOn:    "Last session summary",
		LikelyScenarios: []string{"scenario 1"},
		PrepDate:        time.Now(),
		PendingEffects: []DelayedEffect{
			{ID: "rule-001-0", Description: "Village burns", EffectType: "spawn", Target: "Village", ApplySession: 4},
		},
		FactionSnapshot: []ReputationEntry{
			{FactionID: "thieves-guild", Score: 15, Status: "friendly"},
		},
	}

	data, err := json.Marshal(prep)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded SessionPrep
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(decoded.PendingEffects) != 1 {
		t.Fatalf("pending_effects len = %d, want 1", len(decoded.PendingEffects))
	}
	if decoded.PendingEffects[0].ID != "rule-001-0" {
		t.Fatalf("pending_effects[0].id = %q, want rule-001-0", decoded.PendingEffects[0].ID)
	}
	if len(decoded.FactionSnapshot) != 1 {
		t.Fatalf("faction_snapshot len = %d, want 1", len(decoded.FactionSnapshot))
	}
	if decoded.FactionSnapshot[0].FactionID != "thieves-guild" {
		t.Fatalf("faction_snapshot[0].faction_id = %q, want thieves-guild", decoded.FactionSnapshot[0].FactionID)
	}
}

func TestSessionPrep_BackwardCompat(t *testing.T) {
	oldJSON := `{"campaign_id":"test","session_num":4,"previously_on":"summary","likely_scenarios":[],"prep_date":"2024-01-01T00:00:00Z"}`
	var prep SessionPrep
	if err := json.Unmarshal([]byte(oldJSON), &prep); err != nil {
		t.Fatalf("unmarshal old json failed: %v", err)
	}
	if prep.PendingEffects != nil {
		t.Fatalf("pending_effects should be nil for old json, got %v", prep.PendingEffects)
	}
	if prep.FactionSnapshot != nil {
		t.Fatalf("faction_snapshot should be nil for old json, got %v", prep.FactionSnapshot)
	}
}

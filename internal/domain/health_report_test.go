package domain

import (
	"encoding/json"
	"testing"
)

func TestCampaignHealthReport_StructFields(t *testing.T) {
	report := CampaignHealthReport{
		OverallHealth:      85,
		CanonCompleteness:  90,
		NarrativeCoherence: 80,
		FactionBalance:     75,
		WotCCompliance:     95,
		HookCoverage:       70,
		Status:             "healthy",
	}

	if report.OverallHealth != 85 {
		t.Errorf("expected OverallHealth 85, got %d", report.OverallHealth)
	}
	if report.CanonCompleteness != 90 {
		t.Errorf("expected CanonCompleteness 90, got %d", report.CanonCompleteness)
	}
	if report.NarrativeCoherence != 80 {
		t.Errorf("expected NarrativeCoherence 80, got %d", report.NarrativeCoherence)
	}
	if report.FactionBalance != 75 {
		t.Errorf("expected FactionBalance 75, got %d", report.FactionBalance)
	}
	if report.WotCCompliance != 95 {
		t.Errorf("expected WotCCompliance 95, got %d", report.WotCCompliance)
	}
	if report.HookCoverage != 70 {
		t.Errorf("expected HookCoverage 70, got %d", report.HookCoverage)
	}
	if report.Status != "healthy" {
		t.Errorf("expected Status 'healthy', got %q", report.Status)
	}
}

func TestCampaignHealthReport_JSONRoundTrip(t *testing.T) {
	original := CampaignHealthReport{
		OverallHealth:      60,
		CanonCompleteness:  55,
		NarrativeCoherence: 65,
		FactionBalance:     70,
		WotCCompliance:     50,
		HookCoverage:       60,
		Status:             "fair",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CampaignHealthReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.OverallHealth != original.OverallHealth {
		t.Errorf("expected OverallHealth %d, got %d", original.OverallHealth, decoded.OverallHealth)
	}
	if decoded.Status != original.Status {
		t.Errorf("expected Status %q, got %q", original.Status, decoded.Status)
	}
}

package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestGetTemplate_Existing(t *testing.T) {
	tests := []struct {
		name        string
		wantName    string
		wantTone    string
		wantSetting string
		wantMode    string
	}{
		{"Urban Fantasy", "Urban Fantasy", "grim", "urban", "sandbox_urbano"},
		{"Gothic Horror", "Gothic Horror", "horror", "wilderness", "investigacion"},
		{"Maritime Adventure", "Maritime Adventure", "heroic", "maritime", "viaje"},
		{"Dungeon Crawl", "Dungeon Crawl", "heroic", "dungeon", "dungeon_lineal"},
		{"Political Intrigue", "Political Intrigue", "political", "urban", "intriga"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tmpl.Name != tt.wantName {
				t.Errorf("expected Name %q, got %q", tt.wantName, tmpl.Name)
			}
			if tmpl.Tone != tt.wantTone {
				t.Errorf("expected Tone %q, got %q", tt.wantTone, tmpl.Tone)
			}
			if tmpl.SettingType != tt.wantSetting {
				t.Errorf("expected SettingType %q, got %q", tt.wantSetting, tmpl.SettingType)
			}
			if tmpl.GameMode != tt.wantMode {
				t.Errorf("expected GameMode %q, got %q", tt.wantMode, tmpl.GameMode)
			}
		})
	}
}

func TestGetTemplate_Missing(t *testing.T) {
	_, err := GetTemplate("Nonexistent")
	if err == nil {
		t.Error("expected error for missing template, got nil")
	}
}

func TestGetTemplate_CaseInsensitive(t *testing.T) {
	tmpl, err := GetTemplate("urban fantasy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "Urban Fantasy" {
		t.Errorf("expected Name 'Urban Fantasy', got %q", tmpl.Name)
	}
}

func TestApplyTemplate(t *testing.T) {
	campaign := &domain.Campaign{
		Name:    "test-campaign",
		Title:   "Test Campaign",
		Setting: "A city",
	}

	tmpl := domain.CampaignTemplate{
		Tone:          "grim",
		SettingType:   "urban",
		GameMode:      "sandbox_urbano",
		DefaultThemes: []string{"corruption", "magic"},
	}

	ApplyTemplate(campaign, tmpl)

	if campaign.Status != "active" {
		t.Errorf("expected Status 'active', got %q", campaign.Status)
	}
	// ApplyTemplate should not overwrite existing Title/Setting
	if campaign.Title != "Test Campaign" {
		t.Errorf("Title was overwritten unexpectedly: %q", campaign.Title)
	}
}

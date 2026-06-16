package domain

import (
	"encoding/json"
	"testing"
)

func TestCampaignTemplate_StructFields(t *testing.T) {
	template := CampaignTemplate{
		Name:        "Urban Fantasy",
		Tone:        "grim",
		SettingType: "urban",
		GameMode:    "sandbox_urbano",
		DefaultThemes: []string{"corruption", "magic", "mystery"},
	}

	if template.Name != "Urban Fantasy" {
		t.Errorf("expected Name 'Urban Fantasy', got %q", template.Name)
	}
	if template.Tone != "grim" {
		t.Errorf("expected Tone 'grim', got %q", template.Tone)
	}
	if template.SettingType != "urban" {
		t.Errorf("expected SettingType 'urban', got %q", template.SettingType)
	}
	if template.GameMode != "sandbox_urbano" {
		t.Errorf("expected GameMode 'sandbox_urbano', got %q", template.GameMode)
	}
	if len(template.DefaultThemes) != 3 {
		t.Errorf("expected 3 DefaultThemes, got %d", len(template.DefaultThemes))
	}
}

func TestCampaignTemplate_JSONRoundTrip(t *testing.T) {
	original := CampaignTemplate{
		Name:        "Dungeon Crawl",
		Tone:        "heroic",
		SettingType: "dungeon",
		GameMode:    "dungeon_lineal",
		DefaultThemes: []string{"exploration", "combat", "treasure"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CampaignTemplate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("expected Name %q, got %q", original.Name, decoded.Name)
	}
	if decoded.Tone != original.Tone {
		t.Errorf("expected Tone %q, got %q", original.Tone, decoded.Tone)
	}
	if decoded.SettingType != original.SettingType {
		t.Errorf("expected SettingType %q, got %q", original.SettingType, decoded.SettingType)
	}
	if decoded.GameMode != original.GameMode {
		t.Errorf("expected GameMode %q, got %q", original.GameMode, decoded.GameMode)
	}
	if len(decoded.DefaultThemes) != len(original.DefaultThemes) {
		t.Errorf("expected %d themes, got %d", len(original.DefaultThemes), len(decoded.DefaultThemes))
	}
}

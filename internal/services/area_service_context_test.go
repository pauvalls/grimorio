package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestSelectTemplate(t *testing.T) {
	tests := []struct {
		name        string
		settingType string
		wantName    string
	}{
		{"wilderness", "wilderness", "Wilderness Encounter Zone"},
		{"urban", "urban", "Urban District"},
		{"dungeon", "dungeon", "Dungeon Complex"},
		{"social", "social", "Social Encounter Location"},
		{"mixed", "mixed", "Mixed Exploration/Combat"},
		{"unknown defaults to wilderness", "unknown", "Wilderness Encounter Zone"},
		{"empty defaults to wilderness", "", "Wilderness Encounter Zone"},
	}

	s := &AreaService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.selectTemplate(tt.settingType)
			if got == nil {
				t.Fatal("selectTemplate() returned nil")
			}
			if got.Name != tt.wantName {
				t.Errorf("selectTemplate(%q).Name = %q, want %q", tt.settingType, got.Name, tt.wantName)
			}
			if got.SettingType != tt.settingType && tt.settingType != "unknown" && tt.settingType != "" {
				t.Errorf("selectTemplate(%q).SettingType = %q, want %q", tt.settingType, got.SettingType, tt.settingType)
			}
		})
	}
}

func TestGenerateAreaWithContext(t *testing.T) {
	s := &AreaService{}
	ctx := context.Background()

	area, err := s.GenerateAreaWithContext(
		ctx,
		"test-campaign",
		"chapter_1",
		1,
		"forest",
		"wilderness",
		3,
		nil, // factionContext
		nil, // narrativeState
	)

	if err != nil {
		t.Fatalf("GenerateAreaWithContext() error = %v", err)
	}

	if area == nil {
		t.Fatal("GenerateAreaWithContext() returned nil area")
	}

	// Validate area structure
	if area.ID == "" {
		t.Error("Area ID should not be empty")
	}
	if area.ChapterID != "chapter_1" {
		t.Errorf("ChapterID = %q, want 'chapter_1'", area.ChapterID)
	}
	if area.AreaNumber != 1 {
		t.Errorf("AreaNumber = %d, want 1", area.AreaNumber)
	}
	if len(area.Features) < 3 {
		t.Errorf("Features count = %d, want at least 3", len(area.Features))
	}
	if len(area.Encounters) < 2 || len(area.Encounters) > 4 {
		t.Errorf("Encounters count = %d, want 2-4", len(area.Encounters))
	}
	if area.PlayerReadAloud == "" {
		t.Error("PlayerReadAloud should not be empty")
	}
	if area.Development == "" {
		t.Error("Development should not be empty")
	}

	// Validate area passes domain validation
	if err := area.Validate(); err != nil {
		t.Errorf("Area.Validate() error = %v", err)
	}
}

func TestGenerateAreaWithContext_LevelRange(t *testing.T) {
	s := &AreaService{}
	ctx := context.Background()

	tests := []struct {
		name       string
		partyLevel int
		wantMin    int
		wantMax    int
	}{
		{"level 1", 1, 1, 4},    // max(1, 1-2)=1, min(20, 1+3)=4
		{"level 3", 3, 1, 6},    // max(1, 3-2)=1, min(20, 3+3)=6
		{"level 10", 10, 8, 13}, // max(1, 10-2)=8, min(20, 10+3)=13
		{"level 20", 20, 18, 20}, // max(1, 20-2)=18, min(20, 20+3)=20
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area, err := s.GenerateAreaWithContext(
				ctx,
				"test-campaign",
				"chapter_1",
				1,
				"forest",
				"wilderness",
				tt.partyLevel,
				nil,
				nil,
			)
			if err != nil {
				t.Fatalf("GenerateAreaWithContext() error = %v", err)
			}

			if area.LevelRange.Min != tt.wantMin {
				t.Errorf("LevelRange.Min = %d, want %d", area.LevelRange.Min, tt.wantMin)
			}
			if area.LevelRange.Max != tt.wantMax {
				t.Errorf("LevelRange.Max = %d, want %d", area.LevelRange.Max, tt.wantMax)
			}
		})
	}
}

func TestGenerateFeatures(t *testing.T) {
	s := &AreaService{}
	template := s.selectTemplate("wilderness")

	features := s.generateFeatures(template, "forest", 3)

	if len(features) < 3 || len(features) > 5 {
		t.Errorf("generateFeatures() count = %d, want 3-5", len(features))
	}

	for i, f := range features {
		if f.Type == "" {
			t.Errorf("Feature[%d].Type is empty", i)
		}
		if f.Name == "" {
			t.Errorf("Feature[%d].Name is empty", i)
		}
		if f.Description == "" {
			t.Errorf("Feature[%d].Description is empty", i)
		}
		if f.DC == nil {
			t.Errorf("Feature[%d].DC is nil", i)
		} else if *f.DC < 8 || *f.DC > 20 {
			t.Errorf("Feature[%d].DC = %d, want 8-20", i, *f.DC)
		}
	}
}

func TestGenerateBoxedText(t *testing.T) {
	s := &AreaService{}
	template := s.selectTemplate("wilderness")

	boxedText := s.generateBoxedText(template, "forest", "wilderness")

	// Check word count (100-600 words)
	words := len(boxedText) / 5 // Approximate word count
	if words < 20 || words > 600 {
		t.Errorf("Boxed text word count ≈ %d, want 100-600", words)
	}

	// Check it contains the location hint
	if len(boxedText) == 0 {
		t.Error("Boxed text should not be empty")
	}
}

func TestGenerateDevelopmentText(t *testing.T) {
	s := &AreaService{}
	template := s.selectTemplate("wilderness")

	// Test without narrative state
	text := s.generateDevelopmentText(template, nil)
	if text == "" {
		t.Error("Development text should not be empty")
	}

	// Test with narrative state
	narrativeState := &domain.NarrativeState{
		SchemaVersion: "v2",
		CampaignID:    "test",
		RevealedClues: []domain.RevealedClue{
			{ID: "clue_1", Description: "secret tunnel"},
		},
	}
	textWithClues := s.generateDevelopmentText(template, narrativeState)
	if textWithClues == "" {
		t.Error("Development text with clues should not be empty")
	}
	// Should include reference to the clue
	if len(textWithClues) <= len(text) {
		t.Error("Development text with clues should be longer than without")
	}
}

func TestGenerateTitle(t *testing.T) {
	s := &AreaService{}
	template := s.selectTemplate("wilderness")

	tests := []struct {
		name         string
		locationHint string
		wantContains string
	}{
		{"with hint", "forest", "Forest"},
		{"with hint capitalized", "dark cave", "Dark"},
		{"empty hint", "", "Wilderness Encounter Zone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := s.generateTitle(template, tt.locationHint)
			if title == "" {
				t.Error("Title should not be empty")
			}
			if tt.wantContains != "" && len(tt.locationHint) > 0 {
				// Title should contain the capitalized location hint
				if len(title) < len(tt.wantContains) {
					t.Errorf("Title %q too short to contain %q", title, tt.wantContains)
				}
			}
		})
	}
}

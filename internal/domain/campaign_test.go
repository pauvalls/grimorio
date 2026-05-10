package domain

import (
	"strings"
	"testing"
)

func TestActValidate_ChapterNarrativeFields(t *testing.T) {
	tests := []struct {
		name    string
		act     Act
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid act with all chapter fields",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700), // 700 words
				AssetHandoff:      "Objeto: mapa del tesoro",
			},
			wantErr: false,
		},
		{
			name: "invalid game mode",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "invalid_mode",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "invalid mode",
		},
		{
			name: "missing game mode",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "game mode is required",
		},
		{
			name: "secondary mode equals primary mode",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				GameModeSecondary: "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "secondary mode must differ",
		},
		{
			name: "too few objectives",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Solo uno"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "must have 2-3 objectives",
		},
		{
			name: "too many objectives",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Obj 1", "Obj 2", "Obj 3", "Obj 4"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "must have 2-3 objectives",
		},
		{
			name: "invalid duration format",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "varias sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "must match pattern",
		},
		{
			name: "missing duration",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "estimated duration is required",
		},
		{
			name: "valid single session duration",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "1 sesión",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: false,
		},
		{
			name: "invalid tone",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "invalid_tone",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "invalid tone",
		},
		{
			name: "missing tone",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "tone is required",
		},
		{
			name: "running guidance too short",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   "Demasiado corto",
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: true,
			errMsg:  "must be at least 700 words",
		},
		{
			name: "running guidance too long",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 1000), // 1000 words - OK, no maximum
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: false,
		},
		{
			name: "missing asset handoff",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "",
			},
			wantErr: true,
			errMsg:  "asset handoff is required",
		},
		{
			name: "all valid game modes",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "dungeon_lineal",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: false,
		},
		{
			name: "all valid tones",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "grim",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: false,
		},
		{
			name: "valid with secondary mode",
			act: Act{
				CampaignID:        "test-campaign",
				Number:            1,
				Title:             "Test Act",
				GameMode:          "investigacion",
				GameModeSecondary: "dungeon_lineal",
				ChapterObjectives: []string{"Objetivo 1", "Objetivo 2"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "heroic",
				RunningGuidance:   strings.Repeat("palabra ", 700),
				AssetHandoff:      "Objeto: mapa",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.act.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Act.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Act.Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestIsValidGameMode(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{"investigacion", true},
		{"sandbox_urbano", true},
		{"dungeon_lineal", true},
		{"escape", true},
		{"viaje", true},
		{"intriga", true},
		{"confrontacion", true},
		{"downtime", true},
		{"invalid_mode", false},
		{"", false},
		{"Investigacion", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if got := isValidGameMode(tt.mode); got != tt.want {
				t.Errorf("isValidGameMode(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

func TestIsValidTone(t *testing.T) {
	tests := []struct {
		tone string
		want bool
	}{
		{"grim", true},
		{"whimsical", true},
		{"heroic", true},
		{"horror", true},
		{"political", true},
		{"mystery", true},
		{"invalid_tone", false},
		{"", false},
		{"Heroic", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.tone, func(t *testing.T) {
			if got := isValidTone(tt.tone); got != tt.want {
				t.Errorf("isValidTone(%q) = %v, want %v", tt.tone, got, tt.want)
			}
		})
	}
}

func TestIsValidDurationFormat(t *testing.T) {
	tests := []struct {
		duration string
		want     bool
	}{
		{"1 sesión", true},
		{"2-3 sesiones", true},
		{"4-6 sesiones", true},
		{"10-15 sesiones", true},
		{"varias sesiones", false},
		{"2 horas", false},
		{"largo", false},
		{"", false},
		{"2-3", false},         // missing "sesiones"
		{"sesiones", false},     // missing numbers
		{"2 -3 sesiones", false}, // extra space
		{"10 sesiones", false},   // must be range or singular "1 sesión"
	}

	for _, tt := range tests {
		t.Run(tt.duration, func(t *testing.T) {
			if got := isValidDurationFormat(tt.duration); got != tt.want {
				t.Errorf("isValidDurationFormat(%q) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"", 0},
		{"uno", 1},
		{"uno dos tres", 3},
		{strings.Repeat("palabra ", 200), 200},
		{"  espacios   extra  ", 2},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := countWords(tt.text); got != tt.want {
				t.Errorf("countWords(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

// Helper function for tests
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

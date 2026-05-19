package domain

import (
	"testing"
	"time"
)

func TestPrologue_Validate(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name     string
		prologue Prologue
		want     bool
	}{
		{
			name: "valid 4 parts sequential",
			prologue: Prologue{
				CampaignID: "test-campaign",
				Tone:       "heroic",
				Parts: []ProloguePart{
					{Order: 1, Title: "Hook", Content: "Hook text", IsReadAloud: true},
					{Order: 2, Title: "Context", Content: "Context text", IsReadAloud: false},
					{Order: 3, Title: "Connections", Content: "Connections text", IsReadAloud: false},
					{Order: 4, Title: "Road Ahead", Content: "Road ahead text", IsReadAloud: true},
				},
				GeneratedAt: baseTime,
			},
			want: true,
		},
		{
			name: "invalid - only 2 parts",
			prologue: Prologue{
				CampaignID: "test-campaign",
				Tone:       "heroic",
				Parts: []ProloguePart{
					{Order: 1, Title: "Hook", Content: "Hook text", IsReadAloud: true},
					{Order: 2, Title: "Context", Content: "Context text", IsReadAloud: false},
				},
				GeneratedAt: baseTime,
			},
			want: false,
		},
		{
			name: "invalid - 0 parts",
			prologue: Prologue{
				CampaignID: "test-campaign",
				Tone:       "heroic",
				Parts:      []ProloguePart{},
				GeneratedAt: baseTime,
			},
			want: false,
		},
		{
			name: "invalid - nil parts",
			prologue: Prologue{
				CampaignID:  "test-campaign",
				Tone:        "heroic",
				Parts:       nil,
				GeneratedAt: baseTime,
			},
			want: false,
		},
		{
			name: "invalid - wrong order (1,2,4,3)",
			prologue: Prologue{
				CampaignID: "test-campaign",
				Tone:       "heroic",
				Parts: []ProloguePart{
					{Order: 1, Title: "Hook", Content: "Hook text", IsReadAloud: true},
					{Order: 2, Title: "Context", Content: "Context text", IsReadAloud: false},
					{Order: 4, Title: "Road Ahead", Content: "Road ahead text", IsReadAloud: true},
					{Order: 3, Title: "Connections", Content: "Connections text", IsReadAloud: false},
				},
				GeneratedAt: baseTime,
			},
			want: false,
		},
		{
			name: "invalid - non-sequential order (1,3,5,7)",
			prologue: Prologue{
				CampaignID: "test-campaign",
				Tone:       "heroic",
				Parts: []ProloguePart{
					{Order: 1, Title: "Hook", Content: "Hook text", IsReadAloud: true},
					{Order: 3, Title: "Context", Content: "Context text", IsReadAloud: false},
					{Order: 5, Title: "Connections", Content: "Connections text", IsReadAloud: false},
					{Order: 7, Title: "Road Ahead", Content: "Road ahead text", IsReadAloud: true},
				},
				GeneratedAt: baseTime,
			},
			want: false,
		},
		{
			name: "invalid - 5 parts (too many)",
			prologue: Prologue{
				CampaignID: "test-campaign",
				Tone:       "heroic",
				Parts: []ProloguePart{
					{Order: 1, Title: "Hook", Content: "Hook text", IsReadAloud: true},
					{Order: 2, Title: "Context", Content: "Context text", IsReadAloud: false},
					{Order: 3, Title: "Connections", Content: "Connections text", IsReadAloud: false},
					{Order: 4, Title: "Road Ahead", Content: "Road ahead text", IsReadAloud: true},
					{Order: 5, Title: "Extra", Content: "Extra text", IsReadAloud: false},
				},
				GeneratedAt: baseTime,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prologue.Validate()
			if got != tt.want {
				t.Errorf("Prologue.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

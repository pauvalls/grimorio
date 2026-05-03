package domain

import (
	"fmt"
	"testing"
)

func TestCampaignValidation(t *testing.T) {
	tests := []struct {
		name    string
		campaign Campaign
		wantErr bool
		errField string
	}{
		{
			name: "valid campaign",
			campaign: Campaign{
				Name:    "my-campaign",
				Title:   "My Campaign",
				Setting: "Forgotten Realms",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			campaign: Campaign{
				Title: "My Campaign",
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "invalid kebab-case - uppercase",
			campaign: Campaign{
				Name:  "MyCampaign",
				Title: "My Campaign",
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "invalid kebab-case - spaces",
			campaign: Campaign{
				Name:  "my campaign",
				Title: "My Campaign",
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "invalid kebab-case - starts with hyphen",
			campaign: Campaign{
				Name:  "-my-campaign",
				Title: "My Campaign",
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "title defaults to name",
			campaign: Campaign{
				Name:    "my-campaign",
				Setting: "Test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.campaign.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if ve, ok := err.(ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("Validate() error field = %v, want %v", ve.Field, tt.errField)
					}
				} else {
					t.Errorf("Validate() expected ValidationError but got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
				if tt.campaign.Title == "" {
					t.Errorf("Validate() title should default to name but is empty")
				}
			}
		})
	}
}

func TestIsValidKebabCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple", "my-campaign", true},
		{"with numbers", "campaign-123", true},
		{"single word", "campaign", true},
		{"multiple hyphens", "my-awesome-campaign", true},
		{"empty", "", false},
		{"uppercase", "My-Campaign", false},
		{"spaces", "my campaign", false},
		{"starts with hyphen", "-campaign", false},
		{"ends with hyphen", "campaign-", false},
		{"double hyphen", "my--campaign", false},
		{"special chars", "my@campaign", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidKebabCase(tt.input)
			if got != tt.want {
				t.Errorf("IsValidKebabCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCharacterValidation(t *testing.T) {
	tests := []struct {
		name      string
		character Character
		wantErr   bool
		errField  string
	}{
		{
			name: "valid character",
			character: Character{
				CampaignID: "my-campaign",
				Name:       "Gandalf",
				Race:       "humano",
				Class:      "mago",
				Level:      5,
				Background: "sabio",
				Alignment:  "LG",
			},
			wantErr: false,
		},
		{
			name: "missing campaign",
			character: Character{
				Name:  "Gandalf",
				Race:  "humano",
				Class: "mago",
				Level: 1,
			},
			wantErr:  true,
			errField: "campaign_id",
		},
		{
			name: "missing name",
			character: Character{
				CampaignID: "my-campaign",
				Race:       "humano",
				Class:      "mago",
				Level:      1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "invalid level - too low",
			character: Character{
				CampaignID: "my-campaign",
				Name:       "Gandalf",
				Level:      0,
			},
			wantErr:  true,
			errField: "level",
		},
		{
			name: "invalid level - too high",
			character: Character{
				CampaignID: "my-campaign",
				Name:       "Gandalf",
				Level:      21,
			},
			wantErr:  true,
			errField: "level",
		},
		{
			name: "invalid race",
			character: Character{
				CampaignID: "my-campaign",
				Name:       "Gandalf",
				Race:       "dragon",
				Level:      1,
			},
			wantErr:  true,
			errField: "race",
		},
		{
			name: "invalid class",
			character: Character{
				CampaignID: "my-campaign",
				Name:       "Gandalf",
				Class:      "ninja",
				Level:      1,
			},
			wantErr:  true,
			errField: "class",
		},
		{
			name: "invalid alignment",
			character: Character{
				CampaignID: "my-campaign",
				Name:       "Gandalf",
				Alignment:  "XYZ",
				Level:      1,
			},
			wantErr:  true,
			errField: "alignment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.character.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if ve, ok := err.(ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("Validate() error field = %v, want %v", ve.Field, tt.errField)
					}
				} else {
					t.Errorf("Validate() expected ValidationError but got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestQuestValidation(t *testing.T) {
	tests := []struct {
		name    string
		quest   Quest
		wantErr bool
		errField string
	}{
		{
			name: "valid quest",
			quest: Quest{
				CampaignID: "my-campaign",
				Title:      "Find the Sword",
				Type:       QuestTypeMain,
				Status:     QuestStatusActive,
			},
			wantErr: false,
		},
		{
			name: "missing campaign",
			quest: Quest{
				Title:  "Find the Sword",
				Type:   QuestTypeMain,
				Status: QuestStatusActive,
			},
			wantErr:  true,
			errField: "campaign_id",
		},
		{
			name: "missing title",
			quest: Quest{
				CampaignID: "my-campaign",
				Type:       QuestTypeMain,
				Status:     QuestStatusActive,
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "invalid type",
			quest: Quest{
				CampaignID: "my-campaign",
				Title:      "Find the Sword",
				Type:       "invalid",
				Status:     QuestStatusActive,
			},
			wantErr:  true,
			errField: "type",
		},
		{
			name: "invalid status",
			quest: Quest{
				CampaignID: "my-campaign",
				Title:      "Find the Sword",
				Type:       QuestTypeMain,
				Status:     "invalid",
			},
			wantErr:  true,
			errField: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.quest.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if ve, ok := err.(ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("Validate() error field = %v, want %v", ve.Field, tt.errField)
					}
				} else {
					t.Errorf("Validate() expected ValidationError but got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCalculateModifier(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		{1, -4},
		{10, 0},
		{12, 1},
		{15, 2},
		{18, 4},
		{20, 5},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score-%d", tt.score), func(t *testing.T) {
			got := CalculateModifier(tt.score)
			if got != tt.want {
				t.Errorf("CalculateModifier(%d) = %d, want %d", tt.score, got, tt.want)
			}
		})
	}
}

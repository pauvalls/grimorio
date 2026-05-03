package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

func TestCampaignService_CreateCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "wkhtmltopdf")

	tests := []struct {
		name    string
		campaignName string
		title   string
		setting string
		wantErr bool
	}{
		{
			name:         "create valid campaign",
			campaignName: "test-campaign",
			title:        "Test Campaign",
			setting:      "Forgotten Realms",
			wantErr:      false,
		},
		{
			name:         "create campaign without title",
			campaignName: "no-title",
			setting:      "Test",
			wantErr:      false,
		},
		{
			name:         "create duplicate campaign",
			campaignName: "test-campaign",
			title:        "Duplicate",
			wantErr:      true,
		},
		{
			name:         "invalid campaign name",
			campaignName: "Invalid Name",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign, err := service.CreateCampaign(tt.campaignName, tt.title, tt.setting)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCampaign() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateCampaign() unexpected error: %v", err)
				return
			}
			if campaign.Name != tt.campaignName {
				t.Errorf("CreateCampaign() name = %v, want %v", campaign.Name, tt.campaignName)
			}
			if tt.title != "" && campaign.Title != tt.title {
				t.Errorf("CreateCampaign() title = %v, want %v", campaign.Title, tt.title)
			}
		})
	}
}

func TestCampaignService_GetCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "wkhtmltopdf")

	// Create a campaign first
	_, err := service.CreateCampaign("get-test", "Get Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	tests := []struct {
		name    string
		campaignName string
		wantErr bool
	}{
		{
			name:         "get existing campaign",
			campaignName: "get-test",
			wantErr:      false,
		},
		{
			name:         "get non-existent campaign",
			campaignName: "does-not-exist",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign, err := service.GetCampaign(tt.campaignName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCampaign() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetCampaign() unexpected error: %v", err)
				return
			}
			if campaign.Name != tt.campaignName {
				t.Errorf("GetCampaign() name = %v, want %v", campaign.Name, tt.campaignName)
			}
		})
	}
}

func TestCampaignService_SaveAct(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "wkhtmltopdf")

	// Create a campaign first
	_, err := service.CreateCampaign("act-test", "Act Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	tests := []struct {
		name       string
		campaignID string
		actNumber  int
		title      string
		content    string
		wantErr    bool
	}{
		{
			name:       "save valid act",
			campaignID: "act-test",
			actNumber:  1,
			title:      "The Beginning",
			content:    "Once upon a time...",
			wantErr:    false,
		},
		{
			name:       "save act to non-existent campaign",
			campaignID: "does-not-exist",
			actNumber:  1,
			title:      "Act",
			content:    "Content",
			wantErr:    true,
		},
		{
			name:       "save act with invalid number",
			campaignID: "act-test",
			actNumber:  0,
			title:      "Invalid",
			content:    "Content",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveAct(tt.campaignID, tt.actNumber, tt.title, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveAct() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveAct() unexpected error: %v", err)
				return
			}

			// Verify act was saved
			act, err := actRepo.Read(tt.campaignID, tt.actNumber)
			if err != nil {
				t.Errorf("SaveAct() act not saved: %v", err)
				return
			}
			if act.Number != tt.actNumber {
				t.Errorf("SaveAct() act number = %v, want %v", act.Number, tt.actNumber)
			}
		})
	}
}

func TestCampaignService_ListCampaigns(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "wkhtmltopdf")

	// Create some campaigns
	_, _ = service.CreateCampaign("list-1", "List 1", "Setting 1")
	_, _ = service.CreateCampaign("list-2", "List 2", "Setting 2")

	campaigns, err := service.ListCampaigns()
	if err != nil {
		t.Errorf("ListCampaigns() unexpected error: %v", err)
		return
	}

	if len(campaigns) != 2 {
		t.Errorf("ListCampaigns() got %d campaigns, want 2", len(campaigns))
	}
}
